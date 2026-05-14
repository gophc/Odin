package cmd

import (
	"sync/atomic"
	"unsafe"
)

type TypeAndValueMutexStripe struct {
	Mutex BlockingMutex
}

const typeAndValueMutexStripesCount = 128

var tavMutexStripes [typeAndValueMutexStripesCount]TypeAndValueMutexStripe

var globalProcedureBodyInWorkerQueue atomic.Bool
var globalAfterCheckingProcedureBodies atomic.Bool

func type_and_value_of_expr(expr *Ast) TypeAndValue {
	if expr != nil {
		return expr.TAV
	}
	return TypeAndValue{}
}

func type_of_expr(expr *Ast) *Type {
	tav := expr.TAV
	if tav.Mode != AddressingInvalid {
		return tav.Type
	}
	if entity := entity_of_node(expr); entity != nil {
		return entity.Type
	}
	return nil
}

func implicit_entity_of_node(clause *Ast) *Entity {
	if clause != nil && clause.Kind == AstCaseClause {
		return clause.CaseClause.ImplicitEntity.Load()
	}
	return nil
}

func entity_of_node(expr *Ast) *Entity {
retry:
	expr = unparen_expr(expr)
	switch expr.Kind {
	case AstIdent:
		return expr.Ident.Entity.Load()
	case AstSelectorExpr:
		s := unselector_expr(expr.SelectorExpr.Selector)
		return entity_of_node(s)
	case AstCaseClause:
		return expr.CaseClause.ImplicitEntity.Load()
	case AstCallExpr:
		return expr.CallExpr.EntityProcedureOf.Load()
	case AstTernaryWhenExpr:
		we := &expr.TernaryWhenExpr
		if we.Cond == nil {
			break
		}
		if we.Cond.TAV.Value.Kind != ExactValueBool {
			break
		}
		if we.Cond.TAV.Value.ValueBool {
			expr = we.X
		} else {
			expr = we.Y
		}
		goto retry
	}
	return nil
}

func decl_info_of_entity(e *Entity) *DeclInfo {
	if e != nil {
		return e.DeclInfo
	}
	return nil
}

func check_get_expr_info(c *CheckerContext, expr *Ast) *ExprInfo {
	if c.Untyped != nil {
		found := PtrMapGet(*c.Untyped, expr)
		if found != nil {
			return *found
		}
		return nil
	} else {
		rw_mutex_shared_lock(&c.Info.GlobalUntypedMutex)
		found := PtrMapGet(c.Info.GlobalUntyped, expr)
		rw_mutex_shared_unlock(&c.Info.GlobalUntypedMutex)
		if found != nil {
			return *found
		}
		return nil
	}
}

func check_set_expr_info(c *CheckerContext, expr *Ast, mode AddressingMode, type_ *Type, value ExactValue) {
	if c.Untyped != nil {
		PtrMapSet(*c.Untyped, expr, MakeExprInfo(mode, type_, value, false))
	} else {
		rw_mutex_lock(&c.Info.GlobalUntypedMutex)
		PtrMapSet(c.Info.GlobalUntyped, expr, MakeExprInfo(mode, type_, value, false))
		rw_mutex_unlock(&c.Info.GlobalUntypedMutex)
	}
}

func check_remove_expr_info(c *CheckerContext, e *Ast) {
	if c.Untyped != nil {
		PtrMapRemove(*c.Untyped, e)
		gb_assert_handler("Assertion Failure", "PtrMapGet(*c.Untyped, e) == nil", "checker_context.go", 0)
	} else {
		untyped := c.Info.GlobalUntyped
		rw_mutex_lock(&c.Info.GlobalUntypedMutex)
		PtrMapRemove(untyped, e)
		gb_assert_handler("Assertion Failure", "PtrMapGet(untyped, e) == nil", "checker_context.go", 0)
		rw_mutex_unlock(&c.Info.GlobalUntypedMutex)
	}
}

type typeInfoPair = TypeInfoPair

func type_info_index_pair(info *CheckerInfo, pair TypeInfoPair, error_on_failure bool) isize {
	rw_mutex_shared_lock(&info.MinimumDependencyTypeInfoMutex)
	entryIndex := isize(-1)
	foundEntryIndex := PtrMapGet(info.MinDepTypeInfoIndexMap, pair.Hash)
	if foundEntryIndex != nil {
		entryIndex = *foundEntryIndex
	}
	rw_mutex_shared_unlock(&info.MinimumDependencyTypeInfoMutex)
	if error_on_failure && entryIndex < 0 {
		compiler_error("Type_Info for '%s' could not be found", type_to_string(pair.Type))
	}
	return entryIndex
}

func type_info_index(info *CheckerInfo, type_ *Type, error_on_failure bool) isize {
	type_ = default_type(type_)
	if type_ == t_llvm_bool {
		type_ = t_bool
	}
	hash := TypeHashCanonicalType(type_)
	return type_info_index_pair(info, TypeInfoPair{Type: type_, Hash: hash}, error_on_failure)
}

func add_untyped(c *CheckerContext, expr *Ast, mode AddressingMode, type_ *Type, value ExactValue) {
	if expr == nil {
		return
	}
	if mode == AddressingInvalid {
		return
	}
	if mode == AddressingConstant && type_ == t_invalid {
		compiler_error("add_untyped - invalid type: %s", type_to_string(type_))
	}
	if !is_type_untyped(type_) {
		return
	}
	check_set_expr_info(c, expr, mode, type_, value)
}

func tav_mutex_for_node(node *Ast) *BlockingMutex {
	gb_assert_handler("Assertion Failure", "node != nil", "checker_context.go", 0)
	h := uintptr(unsafe.Pointer(node))
	h ^= h >> 6
	return &tavMutexStripes[h%typeAndValueMutexStripesCount].Mutex
}

func add_type_and_value(ctx *CheckerContext, expr *Ast, mode AddressingMode, type_ *Type, value ExactValue) {
	if expr == nil {
		return
	}
	if mode == AddressingInvalid {
		return
	}
	if mode == AddressingConstant && type_ == t_invalid {
		return
	}
	mutex := tav_mutex_for_node(expr)
	mutex_lock(mutex)

	for {
		expr.TAV.Mode = mode
		if type_ != nil && expr.TAV.Type != nil &&
			is_type_any(type_) && is_type_untyped(expr.TAV.Type) {
		} else {
			expr.TAV.Type = type_
		}
		if mode == AddressingConstant || mode == AddressingInvalid {
			expr.TAV.Value = value
		} else if mode == AddressingValue && type_ != nil && is_type_typeid(type_) {
			expr.TAV.Value = value
		} else if mode == AddressingValue && type_ != nil && is_type_proc(type_) {
			expr.TAV.Value = value
		}
		next := unparen_expr(expr)
		if next == nil || next == expr {
			break
		}
		expr = next
	}

	mutex_unlock(mutex)
}

func add_entity_definition(i *CheckerInfo, identifier *Ast, entity *Entity) {
	gb_assert_handler("Assertion Failure", "identifier != nil", "checker_context.go", 0)
	if identifier.Kind != AstIdent {
		return
	}
	if identifier.Ident.Entity.Load() != nil {
		return
	}
	gb_assert_handler("Assertion Failure", "entity != nil", "checker_context.go", 0)
	identifier.Ident.Entity.Store(entity)
	entity.Identifier.Store(identifier)
	mpsc_enqueue(i.DefinitionQueue, entity)
}

func redeclaration_error(name String, prev *Entity, found *Entity) bool {
	pos := found.Token.Pos
	up := found.UsingParent
	if up != nil {
		if pos == up.Token.Pos {
			return false
		}
		if found.Flags&EntityFlag_Result != 0 {
			error_pos(prev.Token.Pos,
				"Direct shadowing of the named return value '%s' in this scope through 'using'\n"+
					"\tat %s",
				goStr(name), goStr(token_pos_to_string(up.Token.Pos)))
		} else {
			error_pos(prev.Token.Pos,
				"Redeclaration of '%s' in this scope through 'using'\n"+
					"\tat %s",
				goStr(name), goStr(token_pos_to_string(up.Token.Pos)))
		}
	} else {
		if pos == prev.Token.Pos {
			return false
		}
		if found.Flags&EntityFlag_Result != 0 {
			error_pos(prev.Token.Pos,
				"Direct shadowing of the named return value '%s' in this scope\n"+
					"\tat %s",
				goStr(name), goStr(token_pos_to_string(pos)))
		} else {
			error_pos(prev.Token.Pos,
				"Redeclaration of '%s' in this scope\n"+
					"\tat %s",
				goStr(name), goStr(token_pos_to_string(pos)))
		}
	}
	return false
}

func add_entity_flags_from_file(c *CheckerContext, e *Entity, scope *Scope) {
	if c.File != nil && (c.File.Flags&AstFileIsLazy) != 0 && scope.Flags&ScopeFlag_File != 0 {
		pkg := c.File.Pkg
		if pkg.Kind == PackageInit && e.Kind == EntityProcedure && e.Token.String == "main" {
		} else if e.Flags&(EntityFlag_Test|EntityFlag_Init|EntityFlag_Fini) != 0 {
		} else {
			e.Flags |= EntityFlag_Lazy
		}
	}
}

func add_entity_with_name(c *CheckerContext, scope *Scope, identifier *Ast, entity *Entity, name String) bool {
	if scope == nil {
		return false
	}
	if !is_blank_ident_str(name) {
		ie := scope_insert(scope, entity)
		if ie != nil {
			return redeclaration_error(name, entity, ie)
		}
	}
	if identifier != nil {
		if entity.File == nil {
			entity.File = c.File
		}
		add_entity_definition(c.Info, identifier, entity)
	}
	return true
}

func add_entity_with_name_info(info *CheckerInfo, scope *Scope, identifier *Ast, entity *Entity, name String) bool {
	if scope == nil {
		return false
	}
	if !is_blank_ident_str(name) {
		ie := scope_insert(scope, entity)
		if ie != nil {
			return redeclaration_error(name, entity, ie)
		}
	}
	if identifier != nil {
		gb_assert_handler("Assertion Failure", "entity.File != nil", "checker_context.go", 0)
		add_entity_definition(info, identifier, entity)
	}
	return true
}

func add_entity(c *CheckerContext, scope *Scope, identifier *Ast, entity *Entity) bool {
	return add_entity_with_name(c, scope, identifier, entity, entity.Token.String)
}

func add_entity_use(c *CheckerContext, identifier *Ast, entity *Entity) {
	if entity == nil {
		return
	}
	add_declaration_dependency(c, entity)
	entity.Flags |= EntityFlag_Used
	if entity_has_deferred_procedure(entity) {
		deferred := entity.Procedure.DeferredProcedure.Entity
		if deferred != entity {
			add_entity_use(c, nil, deferred)
		}
	}
	if identifier == nil || identifier.Kind != AstIdent {
		return
	}
	entity.Identifier.Store(identifier)
	identifier.Ident.Entity.Store(entity)
	dmsg := entity.DeprecatedMessage
	if dmsg.Len > 0 {
		warning(identifier, "'%s' is deprecated: %s", goStr(entity.Token.String), goStr(dmsg))
	}
	wmsg := entity.WarningMessage
	if wmsg.Len > 0 {
		warning(identifier, "%s: %s", goStr(entity.Token.String), goStr(wmsg))
	}
}

func could_entity_be_lazy(e *Entity, d *DeclInfo) bool {
	if e.Flags&EntityFlag_Lazy == 0 {
		return false
	}
	if e.Flags&(EntityFlag_Test|EntityFlag_Init|EntityFlag_Fini) != 0 {
		return false
	} else if e.Kind == EntityVariable && e.Variable.IsExport {
		return false
	} else if e.Kind == EntityProcedure && e.Procedure.IsExport {
		return false
	}
	for _, attr := range d.Attributes {
		if attr.Kind != AstAttribute {
			continue
		}
		for _, elem := range attr.Attribute.Elems {
			var name String
			switch elem.Kind {
			case AstIdent:
				name = elem.Ident.Token.String
			case AstImplicit:
				name = elem.Implicit.String
			case AstFieldValue:
				if elem.FieldValue.Field.Kind == AstIdent {
					name = elem.FieldValue.Field.Ident.Token.String
				}
			}
			if name.Len != 0 {
				if name == "test" || name == "export" || name == "init" || name == "linkage" {
					return false
				}
			}
		}
	}
	return true
}

func add_entity_and_decl_info(c *CheckerContext, identifier *Ast, e *Entity, d *DeclInfo, is_exported bool) {
	if identifier == nil {
		error_pos(e.Token.Pos, "Invalid variable declaration")
		return
	}
	if identifier.Kind != AstIdent {
		s := expr_to_string(identifier)
		error(identifier, "A variable declaration must be an identifier, got %s", goStr(s))
		gb_string_free(s)
		return
	}
	gb_assert_handler("Assertion Failure", "e != nil && d != nil", "checker_context.go", 0)
	gb_assert_handler("Assertion Failure", "identifier.Ident.Token.String == e.Token.String", "checker_context.go", 0)
	if !could_entity_be_lazy(e, d) {
		e.Flags &= ^EntityFlag_Lazy
	}
	if e.Scope != nil {
		scope := e.Scope
		if scope.Flags&ScopeFlag_File != 0 && is_entity_kind_exported(e.Kind) && is_exported {
			pkg := scope.File.Pkg
			gb_assert_handler("Assertion Failure", "pkg.Scope == scope.Parent", "checker_context.go", 0)
			gb_assert_handler("Assertion Failure", "c.Pkg == pkg", "checker_context.go", 0)
			ee := AstPackageExportedEntity{Identifier: identifier, Entity: e}
			mpmc_enqueue(&pkg.ExportedEntityQueue, ee)
		} else {
			add_entity(c, scope, identifier, e)
		}
	}
	info := c.Info
	add_entity_definition(info, identifier, e)
	gb_assert_handler("Assertion Failure", "e.DeclInfo == nil", "checker_context.go", 0)
	e.DeclInfo = d
	e.Pkg = c.Pkg
	d.Entity = e
	queueCount := isize(-1)
	isLazy := false
	isLazy = (e.Flags & EntityFlag_Lazy) == EntityFlag_Lazy
	if !isLazy {
		queueCount = mpsc_enqueue(info.EntityQueue, e)
	}
	if e.Token.Pos.FileID != 0 {
		e.OrderInSrc = uint64(e.Token.Pos.FileID)<<32 | uint64(e.Token.Pos.Offset)
	} else {
		gb_assert_handler("Assertion Failure", "!isLazy", "checker_context.go", 0)
		e.OrderInSrc = uint64(1 + queueCount)
	}
}

func add_implicit_entity(c *CheckerContext, clause *Ast, e *Entity) {
	gb_assert_handler("Assertion Failure", "clause != nil", "checker_context.go", 0)
	gb_assert_handler("Assertion Failure", "e != nil", "checker_context.go", 0)
	gb_assert_handler("Assertion Failure", "clause.Kind == AstCaseClause", "checker_context.go", 0)
	clause.CaseClause.ImplicitEntity.Store(e)
}

func add_type_info_type(c *CheckerContext, t *Type) {
	if buildContext.NoRTTI {
		return
	}
	if t == nil {
		return
	}
	t = default_type(t)
	if is_type_untyped(t) {
		return
	}
	if is_type_polymorphic(t) {
		return
	}
	add_type_info_type_internal(c, t)
}

func add_type_info_type_internal(c *CheckerContext, t *Type) {
	if t == nil || c == nil {
		return
	}
	add_type_info_dependency(c.Info, c.Decl, t)
}

func check_procedure_later(c *Checker, info *ProcInfo) {
	gb_assert_handler("Assertion Failure", "info != nil", "checker_context.go", 0)
	gb_assert_handler("Assertion Failure", "info.Decl != nil", "checker_context.go", 0)
	if globalAfterCheckingProcedureBodies.Load() {
		e := info.Decl.Entity
		gb_printf("CHECK PROCEDURE LATER! %.*s :: %s {...}\n", e.Token.String.Len, e.Token.String.Data, type_to_string(e.Type))
	}
	if globalProcedureBodyInWorkerQueue.Load() {
		thread_pool_add_task(checkProcInfoWorkerProc, unsafe.Pointer(info))
	} else {
		c.ProcsToCheck = append(c.ProcsToCheck, info)
	}
	{
		gb_assert_handler("Assertion Failure", "info != nil", "checker_context.go", 0)
		gb_assert_handler("Assertion Failure", "info.Decl != nil", "checker_context.go", 0)
		mpsc_enqueue(c.Info.AllProceduresQueue, info)
	}
}

func check_procedure_later_full(c *Checker, file *AstFile, token Token, decl *DeclInfo, type_ *Type, body *Ast, tags uint64) {
	info := permanentAllocItem[ProcInfo]()
	info.File = file
	info.Token = token
	info.Decl = decl
	info.Type = type_
	info.Body = body
	info.Tags = tags
	check_procedure_later(c, info)
}

func check_type_path_push(c *CheckerContext, e *Entity) {
	gb_assert_handler("Assertion Failure", "c.TypePath != nil", "checker_context.go", 0)
	gb_assert_handler("Assertion Failure", "e != nil", "checker_context.go", 0)
	*c.TypePath = append(*c.TypePath, e)
}

func check_type_path_pop(c *CheckerContext) *Entity {
	gb_assert_handler("Assertion Failure", "c.TypePath != nil", "checker_context.go", 0)
	tp := *c.TypePath
	n := len(tp)
	last := tp[n-1]
	*c.TypePath = tp[:n-1]
	return last
}

func proc_group_entities(c *CheckerContext, o Operand) []*Entity {
	if o.Mode == AddressingProcGroup {
		gb_assert_handler("Assertion Failure", "o.ProcGroup != nil", "checker_context.go", 0)
		if o.ProcGroup.Kind == EntityProcGroup {
			check_entity_decl(c, o.ProcGroup, nil, nil)
			return o.ProcGroup.ProcGroup.Entities
		}
	}
	return nil
}

func proc_group_entities_cloned(c *CheckerContext, o Operand) []*Entity {
	entities := proc_group_entities(c, o)
	if len(entities) == 0 {
		return nil
	}
	result := make([]*Entity, len(entities))
	copy(result, entities)
	return result
}

func init_core_type_info(c *Checker) {
	if t_type_info != nil {
		return
	}
	typeInfoEntity := find_core_entity(c, S("Type_Info"))
	gb_assert_handler("Assertion Failure", "typeInfoEntity != nil", "checker_context.go", 0)
	if typeInfoEntity.Type == nil {
		check_single_global_entity(c, typeInfoEntity, typeInfoEntity.DeclInfo)
	}
	gb_assert_handler("Assertion Failure", "typeInfoEntity.Type != nil", "checker_context.go", 0)
	t_type_info = typeInfoEntity.Type
	t_type_info_ptr = alloc_type_pointer(t_type_info)
	gb_assert_handler("Assertion Failure", "is_type_struct(typeInfoEntity.Type)", "checker_context.go", 0)
	tis := &base_type(typeInfoEntity.Type).Struct
	typeInfoEnumValue := find_core_entity(c, S("Type_Info_Enum_Value"))
	t_type_info_enum_value = typeInfoEnumValue.Type
	t_type_info_enum_value_ptr = alloc_type_pointer(t_type_info_enum_value)
	gb_assert_handler("Assertion Failure", "len(tis.Fields) == 5", "checker_context.go", 0)
	typeInfoStringEncodingKind := find_core_entity(c, S("Type_Info_String_Encoding_Kind"))
	t_type_info_string_encoding_kind = typeInfoStringEncodingKind.Type
	typeInfoVariant := tis.Fields[4]
	tivType := typeInfoVariant.Type
	gb_assert_handler("Assertion Failure", "is_type_union(tivType)", "checker_context.go", 0)
	t_type_info_named = find_core_type(c, S("Type_Info_Named"))
	t_type_info_integer = find_core_type(c, S("Type_Info_Integer"))
	t_type_info_rune = find_core_type(c, S("Type_Info_Rune"))
	t_type_info_float = find_core_type(c, S("Type_Info_Float"))
	t_type_info_quaternion = find_core_type(c, S("Type_Info_Quaternion"))
	t_type_info_complex = find_core_type(c, S("Type_Info_Complex"))
	t_type_info_string = find_core_type(c, S("Type_Info_String"))
	t_type_info_boolean = find_core_type(c, S("Type_Info_Boolean"))
	t_type_info_any = find_core_type(c, S("Type_Info_Any"))
	t_type_info_typeid = find_core_type(c, S("Type_Info_Type_Id"))
	t_type_info_pointer = find_core_type(c, S("Type_Info_Pointer"))
	t_type_info_multi_pointer = find_core_type(c, S("Type_Info_Multi_Pointer"))
	t_type_info_procedure = find_core_type(c, S("Type_Info_Procedure"))
	t_type_info_array = find_core_type(c, S("Type_Info_Array"))
	t_type_info_enumerated_array = find_core_type(c, S("Type_Info_Enumerated_Array"))
	t_type_info_dynamic_array = find_core_type(c, S("Type_Info_Dynamic_Array"))
	t_type_info_slice = find_core_type(c, S("Type_Info_Slice"))
	t_type_info_parameters = find_core_type(c, S("Type_Info_Parameters"))
	t_type_info_struct = find_core_type(c, S("Type_Info_Struct"))
	t_type_info_union = find_core_type(c, S("Type_Info_Union"))
	t_type_info_enum = find_core_type(c, S("Type_Info_Enum"))
	t_type_info_map = find_core_type(c, S("Type_Info_Map"))
	t_type_info_bit_set = find_core_type(c, S("Type_Info_Bit_Set"))
	t_type_info_simd_vector = find_core_type(c, S("Type_Info_Simd_Vector"))
	t_type_info_matrix = find_core_type(c, S("Type_Info_Matrix"))
	t_type_info_soa_pointer = find_core_type(c, S("Type_Info_Soa_Pointer"))
	t_type_info_bit_field = find_core_type(c, S("Type_Info_Bit_Field"))
	t_type_info_fixed_capacity_dynamic_array = find_core_type(c, S("Type_Info_Fixed_Capacity_Dynamic_Array"))
	t_type_info_named_ptr = alloc_type_pointer(t_type_info_named)
	t_type_info_integer_ptr = alloc_type_pointer(t_type_info_integer)
	t_type_info_rune_ptr = alloc_type_pointer(t_type_info_rune)
	t_type_info_float_ptr = alloc_type_pointer(t_type_info_float)
	t_type_info_quaternion_ptr = alloc_type_pointer(t_type_info_quaternion)
	t_type_info_complex_ptr = alloc_type_pointer(t_type_info_complex)
	t_type_info_string_ptr = alloc_type_pointer(t_type_info_string)
	t_type_info_boolean_ptr = alloc_type_pointer(t_type_info_boolean)
	t_type_info_any_ptr = alloc_type_pointer(t_type_info_any)
	t_type_info_typeid_ptr = alloc_type_pointer(t_type_info_typeid)
	t_type_info_pointer_ptr = alloc_type_pointer(t_type_info_pointer)
	t_type_info_multi_pointer_ptr = alloc_type_pointer(t_type_info_multi_pointer)
	t_type_info_procedure_ptr = alloc_type_pointer(t_type_info_procedure)
	t_type_info_array_ptr = alloc_type_pointer(t_type_info_array)
	t_type_info_enumerated_array_ptr = alloc_type_pointer(t_type_info_enumerated_array)
	t_type_info_dynamic_array_ptr = alloc_type_pointer(t_type_info_dynamic_array)
	t_type_info_slice_ptr = alloc_type_pointer(t_type_info_slice)
	t_type_info_parameters_ptr = alloc_type_pointer(t_type_info_parameters)
	t_type_info_struct_ptr = alloc_type_pointer(t_type_info_struct)
	t_type_info_union_ptr = alloc_type_pointer(t_type_info_union)
	t_type_info_enum_ptr = alloc_type_pointer(t_type_info_enum)
	t_type_info_map_ptr = alloc_type_pointer(t_type_info_map)
	t_type_info_bit_set_ptr = alloc_type_pointer(t_type_info_bit_set)
	t_type_info_simd_vector_ptr = alloc_type_pointer(t_type_info_simd_vector)
	t_type_info_matrix_ptr = alloc_type_pointer(t_type_info_matrix)
	t_type_info_soa_pointer_ptr = alloc_type_pointer(t_type_info_soa_pointer)
	t_type_info_bit_field_ptr = alloc_type_pointer(t_type_info_bit_field)
	t_type_info_fixed_capacity_dynamic_array_ptr = alloc_type_pointer(t_type_info_fixed_capacity_dynamic_array)
}

func init_mem_allocator(c *Checker) {
	if t_allocator != nil {
		return
	}
	t_allocator = find_core_type(c, S("Allocator"))
	t_allocator_ptr = alloc_type_pointer(t_allocator)
	t_allocator_error = find_core_type(c, S("Allocator_Error"))
}

func init_core_context(c *Checker) {
	if t_context != nil {
		return
	}
	t_context = find_core_type(c, S("Context"))
	t_context_ptr = alloc_type_pointer(t_context)
}

func init_core_source_code_location(c *Checker) {
	if t_source_code_location != nil {
		return
	}
	t_source_code_location = find_core_type(c, S("Source_Code_Location"))
	t_source_code_location_ptr = alloc_type_pointer(t_source_code_location)
}

func init_core_load_directory_file(c *Checker) {
	if t_load_directory_file != nil {
		return
	}
	t_load_directory_file = find_core_type(c, S("Load_Directory_File"))
	t_load_directory_file_ptr = alloc_type_pointer(t_load_directory_file)
	t_load_directory_file_slice = alloc_type_slice(t_load_directory_file)
}

func init_core_map_type(c *Checker) {
	if t_map_info != nil {
		return
	}
	init_mem_allocator(c)
	t_map_info = find_core_type(c, S("Map_Info"))
	t_map_cell_info = find_core_type(c, S("Map_Cell_Info"))
	t_raw_map = find_core_type(c, S("Raw_Map"))
	t_map_info_ptr = alloc_type_pointer(t_map_info)
	t_map_cell_info_ptr = alloc_type_pointer(t_map_cell_info)
	t_raw_map_ptr = alloc_type_pointer(t_raw_map)
}

func init_core_objc_c(c *Checker) {
	if buildContext.Metrics.Os == TargetOsDarwin {
		t_objc_super = find_core_type(c, S("objc_super"))
		t_objc_super_ptr = alloc_type_pointer(t_objc_super)
	}
}

func init_preload(c *Checker) {
	init_core_type_info(c)
	init_mem_allocator(c)
	init_core_context(c)
	init_core_source_code_location(c)
	init_core_map_type(c)
	init_core_objc_c(c)
}

// checkProcInfoWorkerProc is referenced by check_procedure_later
func checkProcInfoWorkerProc(data unsafe.Pointer) isize {
	return 0
}
