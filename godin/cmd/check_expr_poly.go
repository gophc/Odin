package cmd

import "sort"

func valid_index_and_score_cmp(a, b ValidIndexAndScore) int {
	if a.Score < b.Score {
		return 1
	} else if a.Score > b.Score {
		return -1
	}
	return 0
}

func sort_valid_indices(valids []ValidIndexAndScore) {
	sort.Slice(valids, func(i, j int) bool {
		return valids[i].Score > valids[j].Score
	})
}

func find_or_generate_polymorphic_procedure(old_c *CheckerContext, base_entity *Entity, type_ *Type, param_operands []Operand, poly_def_node *Ast, poly_proc_data *PolyProcData) bool {
	info := old_c.Info

	if base_entity == nil {
		return false
	}

	if !is_type_proc(base_entity.Type) {
		return false
	}

	if base_entity.Flags&EntityFlag_Disabled != 0 {
		return false
	}

	name := base_entity.Token.String

	src := base_type(base_entity.Type)
	var dst *Type
	if type_ != nil {
		dst = base_type(type_)
	}

	if param_operands == nil {
		if dst == nil {
			gb_assert_handler("Assertion Failure", "dst != nil", "cmd_check_expr_poly.go", 0)
		}
	}
	if param_operands != nil {
		if dst != nil {
			gb_assert_handler("Assertion Failure", "dst == nil", "cmd_check_expr_poly.go", 0)
		}
	}

	if !src.Proc.IsPolymorphic || src.Proc.IsPolySpecialized {
		return false
	}

	if dst != nil {
		if dst.Proc.IsPolymorphic {
			return false
		}

		if dst.Proc.ParamCount != src.Proc.ParamCount || dst.Proc.ResultCount != src.Proc.ResultCount {
			return false
		}
	}

	old_decl := decl_info_of_entity(base_entity)
	if old_decl == nil {
		return false
	}

	operands := param_operands
	if param_operands == nil {
		operands = make([]Operand, 0, dst.Proc.ParamCount)
		for i := isize(0); i < isize(dst.Proc.ParamCount); i++ {
			param := dst.Proc.Params.Tuple.Variables[i]
			o := Operand{Mode: Addressing_Value}
			o.Type = param.Type
			operands = append(operands, o)
		}
	}

	nctx := *old_c

	scope := create_scope(info, base_entity.Scope)
	scope.flags |= ScopeFlag_Proc
	nctx.Scope = scope
	nctx.AllowPolymorphicTypes = true
	if nctx.PolymorphicScope == nil {
		nctx.PolymorphicScope = scope
	}

	pt := &src.Proc

	final_proc_type := alloc_type_proc(scope, nil, 0, nil, 0, false, pt.CallingConvention)
	success := check_procedure_type(&nctx, final_proc_type, pt.Node, operands)

	if !success {
		return false
	}

	var gen_procs *GenProcsData

	gb_assert_handler("Assertion Failure", base_entity.Identifier.Load().Kind == Ast_Ident, "cmd_check_expr_poly.go", 0)
	gb_assert_handler("Assertion Failure", base_entity.Kind == Entity_Procedure, "cmd_check_expr_poly.go", 0)

	mutex_lock(&base_entity.Procedure.GenProcsMutex)
	gen_procs = base_entity.Procedure.GenProcs
	if gen_procs != nil {
		rw_mutex_shared_lock(&gen_procs.Mutex)

		mutex_unlock(&base_entity.Procedure.GenProcsMutex)

		for _, other := range gen_procs.Procs {
			pt2 := base_type(other.Type)
			if are_types_identical(pt2, final_proc_type) {
				rw_mutex_shared_unlock(&gen_procs.Mutex)

				if poly_proc_data != nil {
					poly_proc_data.GenEntity = *other
				}
				return true
			}
		}

		rw_mutex_shared_unlock(&gen_procs.Mutex)
	} else {
		gen_procs = permanent_alloc_item[GenProcsData]()
		gen_procs.Procs = make([]*Entity, 0)
		base_entity.Procedure.GenProcs = gen_procs
		mutex_unlock(&base_entity.Procedure.GenProcsMutex)
	}

	{
		prev_no_polymorphic_errors := nctx.NoPolymorphicErrors
		defer func() { nctx.NoPolymorphicErrors = prev_no_polymorphic_errors }()
		nctx.NoPolymorphicErrors = false

		scope.head_child = nil
		scope_map_clear(&scope.elements)
		ptr_set_clear(&scope.imported)

		cloned_proc_type_node := clone_ast(pt.Node, nil)
		success = check_procedure_type(&nctx, final_proc_type, cloned_proc_type_node, operands)
		if !success {
			return false
		}

		rw_mutex_shared_lock(&gen_procs.Mutex)
		for _, other := range gen_procs.Procs {
			pt2 := base_type(other.Type)
			if are_types_identical(pt2, final_proc_type) {
				rw_mutex_shared_unlock(&gen_procs.Mutex)

				if poly_proc_data != nil {
					poly_proc_data.GenEntity = *other
				}

				decl := other.DeclInfo
				if decl.ProcCheckedState != ProcCheckedState_Checked {
					proc_info := permanent_alloc_item[ProcInfo]()
					proc_info.File = other.File
					proc_info.Token = other.Token
					proc_info.Decl = decl
					proc_info.Type = other.Type
					proc_info.Body = decl.ProcLit.ProcLit.Body
					proc_info.Tags = other.Procedure.Tags
					proc_info.GeneratedFromPolymorphic = true
					proc_info.PolyDefNode = poly_def_node

					check_procedure_later(nctx.Checker, proc_info)
				}

				return true
			}
		}
		rw_mutex_shared_unlock(&gen_procs.Mutex)
	}

	proc_lit := clone_ast(old_decl.ProcLit, nil)
	pl := &proc_lit.ProcLit

	add_scope(&nctx, pl.Type, final_proc_type.Proc.Scope)
	final_proc_type.Proc.IsPolySpecialized = true
	final_proc_type.Proc.IsPolymorphic = true

	final_proc_type.Proc.Variadic = src.Proc.Variadic
	final_proc_type.Proc.RequireResults = src.Proc.RequireResults
	final_proc_type.Proc.CVararg = src.Proc.CVararg
	final_proc_type.Proc.HasNamedResults = src.Proc.HasNamedResults
	final_proc_type.Proc.Diverging = src.Proc.Diverging
	final_proc_type.Proc.ReturnByPointer = src.Proc.ReturnByPointer
	final_proc_type.Proc.OptionalOK = src.Proc.OptionalOK
	final_proc_type.Proc.EnableTargetFeature = src.Proc.EnableTargetFeature
	final_proc_type.Proc.RequireTargetFeature = src.Proc.RequireTargetFeature

	for i := range operands {
		o := operands[i]
		if final_proc_type == o.Type || base_entity.Type == o.Type {
			final_proc_type.Proc.IsPolySpecialized = false
			break
		}
	}

	tags := base_entity.Procedure.Tags
	ident := clone_ast(base_entity.Identifier, nil)
	token := ident.Ident.Token
	d := make_decl_info(scope, old_decl.Parent)
	d.GenProcType = final_proc_type
	d.TypeExpr = pl.Type
	d.ProcLit = proc_lit
	d.ProcCheckedState = ProcCheckedState_Unchecked
	d.DeferUseChecked = false
	d.ParaPolyOriginal = old_decl.Entity.Load()

	entity := alloc_entity_procedure(nil, token, final_proc_type, tags)
	entity.State = EntityState_Resolved
	entity.Identifier = ident

	add_entity_and_decl_info(&nctx, ident, entity, d, false)
	entity.Scope = scope.Parent
	entity.File = base_entity.File
	entity.Pkg = base_entity.Pkg
	entity.Flags = 0

	entity.Procedure.OptimizationMode = base_entity.Procedure.OptimizationMode

	if base_entity.Flags&EntityFlag_Cold != 0 {
		entity.Flags |= EntityFlag_Cold
	}
	if base_entity.Flags&EntityFlag_Disabled != 0 {
		entity.Flags |= EntityFlag_Disabled
	}

	d.Entity.Store(entity)

	var file *AstFile
	{
		s := entity.Scope
		for s != nil && s.File == nil {
			file = s.File
			s = s.Parent
		}
	}

	rw_mutex_lock(&gen_procs.Mutex)
	gen_procs.Procs = append(gen_procs.Procs, entity)
	rw_mutex_unlock(&gen_procs.Mutex)

	proc_info := permanent_alloc_item[ProcInfo]()
	proc_info.File = file
	proc_info.Token = token
	proc_info.Decl = d
	proc_info.Type = final_proc_type
	proc_info.Body = pl.Body
	proc_info.Tags = tags
	proc_info.GeneratedFromPolymorphic = true
	proc_info.PolyDefNode = poly_def_node

	if poly_proc_data != nil {
		poly_proc_data.GenEntity = *entity
		poly_proc_data.ProcInfo = *proc_info
		entity.Procedure.GeneratedFromPolymorphic = proc_info.GeneratedFromPolymorphic
	}

	check_procedure_later(nctx.Checker, proc_info)

	return true
}

func check_polymorphic_procedure_assignment(c *CheckerContext, operand *Operand, type_ *Type, poly_def_node *Ast, poly_proc_data *PolyProcData) bool {
	if operand.Expr == nil {
		return false
	}
	base_entity := entity_from_expr(operand.Expr)
	if base_entity == nil {
		return false
	}
	return find_or_generate_polymorphic_procedure(c, base_entity, type_, nil, poly_def_node, poly_proc_data)
}

func find_or_generate_polymorphic_procedure_from_parameters(c *CheckerContext, base_entity *Entity, operands []Operand, poly_def_node *Ast, poly_proc_data *PolyProcData) bool {
	return find_or_generate_polymorphic_procedure(c, base_entity, nil, operands, poly_def_node, poly_proc_data)
}

func polymorphic_assign_index(gt **Type, dst_count *int64, source_count int64) bool {
	gt_ := *gt

	if gt_.Kind != Type_Generic {
		gb_assert_handler("Assertion Failure", "gt->kind == Type_Generic", "cmd_check_expr_poly.go", 0)
	}
	e := scope_lookup(gt_.Generic.Scope, gt_.Generic.InternedName, 0)
	if e == nil {
		gb_assert_handler("Assertion Failure", "e != nil", "cmd_check_expr_poly.go", 0)
	}
	if e.Kind == Entity_TypeName {
		*gt = nil
		*dst_count = source_count

		e.Kind = Entity_Constant
		e.Constant.Value = exact_value_i64(source_count)
		e.Type = t_untyped_integer
		return true
	} else if e.Kind == Entity_Constant {
		*gt = nil
		if e.Constant.Value.Kind != ExactValue_Integer {
			return false
		}
		count := big_int_to_i64(&e.Constant.Value.ValueInteger)
		if count != source_count {
			return false
		}
		*dst_count = source_count
		return true
	}
	return false
}

func is_polymorphic_type_assignable(c *CheckerContext, poly *Type, source *Type, compound bool, modify_type bool) bool {
	o := Operand{Mode: Addressing_Value}
	o.Type = source
	switch poly.Kind {
	case Type_Basic:
		if compound {
			return are_types_identical(poly, source)
		}
		return check_is_assignable_to(c, &o, poly)

	case Type_Named:
		if check_type_specialization_to(c, poly, source, compound, modify_type) {
			return true
		}
		if compound || !is_type_generic(poly) {
			return are_types_identical(poly, source)
		}
		return check_is_assignable_to(c, &o, poly)

	case Type_Generic:
		if poly.Generic.Specialized != nil {
			s := poly.Generic.Specialized
			if !check_type_specialization_to(c, s, source, compound, modify_type) {
				return false
			}
		}
		if modify_type {
			ds := default_type(source)
			*poly = *ds
		}
		return true

	case Type_Pointer:
		if source.Kind == Type_Pointer {
			level := check_is_assignable_to_using_subtype(source.Pointer.Elem, poly.Pointer.Elem, 0, false, true)
			if level > 0 {
				return true
			}
			return is_polymorphic_type_assignable(c, poly.Pointer.Elem, source.Pointer.Elem, true, modify_type)
		} else if source.Kind == Type_MultiPointer {
			level := check_is_assignable_to_using_subtype(source.MultiPointer.Elem, poly.Pointer.Elem)
			if level > 0 {
				return true
			}
			return is_polymorphic_type_assignable(c, poly.Pointer.Elem, source.MultiPointer.Elem, true, modify_type)
		}
		return false

	case Type_MultiPointer:
		if source.Kind == Type_MultiPointer {
			level := check_is_assignable_to_using_subtype(source.MultiPointer.Elem, poly.MultiPointer.Elem)
			if level > 0 {
				return true
			}
			return is_polymorphic_type_assignable(c, poly.MultiPointer.Elem, source.MultiPointer.Elem, true, modify_type)
		} else if source.Kind == Type_Pointer {
			level := check_is_assignable_to_using_subtype(source.Pointer.Elem, poly.MultiPointer.Elem)
			if level > 0 {
				return true
			}
			return is_polymorphic_type_assignable(c, poly.MultiPointer.Elem, source.Pointer.Elem, true, modify_type)
		}
		return false

	case Type_SoaPointer:
		if source.Kind == Type_SoaPointer {
			level := check_is_assignable_to_using_subtype(source.SoaPointer.Elem, poly.SoaPointer.Elem, 0, false, true)
			if level > 0 {
				return true
			}
			return is_polymorphic_type_assignable(c, poly.SoaPointer.Elem, source.SoaPointer.Elem, true, modify_type)
		}
		return false

	case Type_Array:
		if source.Kind == Type_Array {
			if poly.Array.GenericCount != nil {
				if !polymorphic_assign_index(&poly.Array.GenericCount, &poly.Array.Count, source.Array.Count) {
					return false
				}
			}
			if poly.Array.Count == source.Array.Count {
				return is_polymorphic_type_assignable(c, poly.Array.Elem, source.Array.Elem, true, modify_type)
			}
		} else if source.Kind == Type_EnumeratedArray {
			if poly.Array.GenericCount != nil {
				gt := poly.Array.GenericCount
				if gt.Kind != Type_Generic {
					gb_assert_handler("Assertion Failure", "gt->kind == Type_Generic", "cmd_check_expr_poly.go", 0)
				}
				e := scope_lookup(gt.Generic.Scope, gt.Generic.InternedName, 0)
				if e == nil {
					gb_assert_handler("Assertion Failure", "e != nil", "cmd_check_expr_poly.go", 0)
				}
				if e.Kind == Entity_TypeName {
					index := source.EnumeratedArray.Index
					it := base_type(index)
					if it.Kind != Type_Enum {
						return false
					}

					poly.Kind = Type_EnumeratedArray
					poly.CachedSize = -1
					poly.CachedAlign = -1
					poly.Flags = source.Flags
					poly.Failure = false
					poly.EnumeratedArray.Elem = source.EnumeratedArray.Elem
					poly.EnumeratedArray.Index = source.EnumeratedArray.Index
					poly.EnumeratedArray.MinValue = source.EnumeratedArray.MinValue
					poly.EnumeratedArray.MaxValue = source.EnumeratedArray.MaxValue
					poly.EnumeratedArray.Count = source.EnumeratedArray.Count
					poly.EnumeratedArray.Op = source.EnumeratedArray.Op

					e.Kind = Entity_TypeName
					e.TypeName.IsTypeAlias = true
					e.Type = index

					if poly.EnumeratedArray.Count == source.EnumeratedArray.Count {
						return is_polymorphic_type_assignable(c, poly.EnumeratedArray.Elem, source.EnumeratedArray.Elem, true, modify_type)
					}
				}
			}
		}
		return false

	case Type_EnumeratedArray:
		if source.Kind == Type_EnumeratedArray {
			if poly.EnumeratedArray.Op != source.EnumeratedArray.Op {
				return false
			}
			if poly.EnumeratedArray.Op {
				if poly.EnumeratedArray.Count != source.EnumeratedArray.Count {
					return false
				}
				if compare_exact_values(Token_NotEq, *poly.EnumeratedArray.MinValue, *source.EnumeratedArray.MinValue) {
					return false
				}
				if compare_exact_values(Token_NotEq, *poly.EnumeratedArray.MaxValue, *source.EnumeratedArray.MaxValue) {
					return false
				}
				return is_polymorphic_type_assignable(c, poly.EnumeratedArray.Index, source.EnumeratedArray.Index, true, modify_type)
			}
			index := is_polymorphic_type_assignable(c, poly.EnumeratedArray.Index, source.EnumeratedArray.Index, true, modify_type)
			elem := is_polymorphic_type_assignable(c, poly.EnumeratedArray.Elem, source.EnumeratedArray.Elem, true, modify_type)
			return index || elem
		}
		return false

	case Type_DynamicArray:
		if source.Kind == Type_DynamicArray {
			return is_polymorphic_type_assignable(c, poly.DynamicArray.Elem, source.DynamicArray.Elem, true, modify_type)
		}
		return false

	case Type_FixedCapacityDynamicArray:
		if source.Kind == Type_FixedCapacityDynamicArray {
			if poly.FixedCapacityDynamicArray.GenericCapacity != nil {
				if !polymorphic_assign_index(&poly.FixedCapacityDynamicArray.GenericCapacity, &poly.FixedCapacityDynamicArray.Capacity, source.FixedCapacityDynamicArray.Capacity) {
					return false
				}
			}
			if poly.FixedCapacityDynamicArray.Capacity == source.FixedCapacityDynamicArray.Capacity {
				return is_polymorphic_type_assignable(c, poly.FixedCapacityDynamicArray.Elem, source.FixedCapacityDynamicArray.Elem, true, modify_type)
			}
		}
		return false

	case Type_Slice:
		if source.Kind == Type_Slice {
			return is_polymorphic_type_assignable(c, poly.Slice.Elem, source.Slice.Elem, true, modify_type)
		}
		return false

	case Type_Enum:
		return false

	case Type_BitSet:
		if source.Kind == Type_BitSet {
			if !is_type_polymorphic(poly.BitSet.Elem) {
				if poly.BitSet.Upper != source.BitSet.Upper || poly.BitSet.Lower != source.BitSet.Lower {
					return false
				}
			}
			if !is_polymorphic_type_assignable(c, poly.BitSet.Elem, source.BitSet.Elem, true, modify_type) {
				return false
			}

			if poly.BitSet.Upper == 0 && modify_type {
				poly.BitSet.Upper = source.BitSet.Upper
			}
			if poly.BitSet.Lower == 0 && modify_type {
				poly.BitSet.Lower = source.BitSet.Lower
			}

			if poly.BitSet.Underlying == nil {
				if modify_type {
					poly.BitSet.Underlying = source.BitSet.Underlying
				}
			} else if !is_polymorphic_type_assignable(c, poly.BitSet.Underlying, source.BitSet.Underlying, true, modify_type) {
				return false
			}
			return true
		}
		return false

	case Type_Union:
		if source.Kind == Type_Union {
			x := &poly.Union
			y := &source.Union
			if len(x.Variants) != len(y.Variants) {
				return false
			}
			for i := range x.Variants {
				a := x.Variants[i]
				b := y.Variants[i]
				ok := is_polymorphic_type_assignable(c, a, b, false, modify_type)
				if !ok {
					return false
				}
			}
			return true
		}
		return false

	case Type_Struct:
		if source.Kind == Type_Struct {
			if poly.Struct.SoaKind == source.Struct.SoaKind && poly.Struct.SoaKind != StructSoa_None {
				ok := is_polymorphic_type_assignable(c, poly.Struct.SoaElem, source.Struct.SoaElem, true, modify_type)
				if ok {
					switch source.Struct.SoaKind {
					case StructSoa_None:
					default:
						gb_assert_handler("Panic", 0, "cmd_check_expr_poly.go", 0, "Unhandled SOA Kind")
					case StructSoa_Fixed:
						if modify_type {
							type_ := make_soa_struct_fixed(c, nil, poly.Struct.Node, poly.Struct.SoaElem, poly.Struct.SoaCount, nil)
							*poly = *type_
						}
					case StructSoa_Slice:
						if modify_type {
							type_ := make_soa_struct_slice(c, nil, poly.Struct.Node, poly.Struct.SoaElem)
							*poly = *type_
						}
					case StructSoa_Dynamic:
						if modify_type {
							type_ := make_soa_struct_dynamic_array(c, nil, poly.Struct.Node, poly.Struct.SoaElem)
							*poly = *type_
						}
					}
					return ok
				}
			}
		}
		return false

	case Type_BitField:
		if source.Kind == Type_BitField {
			return is_polymorphic_type_assignable(c, poly.BitField.BackingType, source.BitField.BackingType, true, modify_type)
		}
		return false

	case Type_Tuple:
		gb_assert_handler("Panic", 0, "cmd_check_expr_poly.go", 0, "This should never happen")
		return false

	case Type_Proc:
		if source.Kind == Type_Proc {
			x := &poly.Proc
			y := &source.Proc
			if x.CallingConvention != y.CallingConvention {
				return false
			}
			if x.CVararg != y.CVararg {
				return false
			}
			if x.Variadic != y.Variadic {
				return false
			}
			if x.ParamCount != y.ParamCount {
				return false
			}
			if x.ResultCount != y.ResultCount {
				return false
			}

			for i := isize(0); i < isize(x.ParamCount); i++ {
				a := x.Params.Tuple.Variables[i]
				b := y.Params.Tuple.Variables[i]
				ok := is_polymorphic_type_assignable(c, a.Type, b.Type, false, modify_type)
				if !ok {
					return false
				}
			}
			for i := isize(0); i < isize(x.ResultCount); i++ {
				a := x.Results.Tuple.Variables[i]
				b := y.Results.Tuple.Variables[i]
				ok := is_polymorphic_type_assignable(c, a.Type, b.Type, false, modify_type)
				if !ok {
					return false
				}
			}

			return true
		}
		return false

	case Type_Map:
		if source.Kind == Type_Map {
			key := is_polymorphic_type_assignable(c, poly.Map.Key, source.Map.Key, true, modify_type)
			value := is_polymorphic_type_assignable(c, poly.Map.Value, source.Map.Value, true, modify_type)
			if key || value {
				poly.Map.LookupResultType = nil
				init_map_internal_types(poly)
				return true
			}
		}
		return false

	case Type_Matrix:
		if source.Kind == Type_Matrix {
			if poly.Matrix.GenericRowCount != nil {
				poly.Matrix.StrideInBytes = 0
				if !polymorphic_assign_index(&poly.Matrix.GenericRowCount, &poly.Matrix.RowCount, source.Matrix.RowCount) {
					return false
				}
			}
			if poly.Matrix.GenericColumnCount != nil {
				poly.Matrix.StrideInBytes = 0
				if !polymorphic_assign_index(&poly.Matrix.GenericColumnCount, &poly.Matrix.ColumnCount, source.Matrix.ColumnCount) {
					return false
				}
			}
			if poly.Matrix.RowCount == source.Matrix.RowCount && poly.Matrix.ColumnCount == source.Matrix.ColumnCount {
				return is_polymorphic_type_assignable(c, poly.Matrix.Elem, source.Matrix.Elem, true, modify_type)
			}
		}
		return false

	case Type_SimdVector:
		if source.Kind == Type_SimdVector {
			if poly.SimdVector.GenericCount != nil {
				if !polymorphic_assign_index(&poly.SimdVector.GenericCount, &poly.SimdVector.Count, source.SimdVector.Count) {
					return false
				}
			}
			if poly.SimdVector.Count == source.SimdVector.Count {
				return is_polymorphic_type_assignable(c, poly.SimdVector.Elem, source.SimdVector.Elem, true, modify_type)
			}
		}
		return false
	}
	return false
}

func lookup_polymorphic_record_parameter(t *Type, parameter_name String) isize {
	if !is_type_polymorphic_record(t) {
		return -1
	}

	params := get_record_polymorphic_params(t)
	if params == nil {
		return -1
	}
	for i, e := range params.Variables {
		name := e.Token.String
		if is_blank_ident(name) {
			continue
		}
		if name == parameter_name {
			return isize(i)
		}
	}
	return -1
}

func populate_proc_parameter_list(c *CheckerContext, proc_type *Type, lhs_count_ *isize) []*Entity {
	var lhs []*Entity
	lhs_count := isize(-1)

	if proc_type == nil || proc_type == t_invalid {
		return nil
	}

	if !is_type_proc(proc_type) {
		gb_assert_handler("Assertion Failure", "is_type_proc(proc_type)", "cmd_check_expr_poly.go", 0)
	}
	pt := &base_type(proc_type).Proc

	if !pt.IsPolymorphic || pt.IsPolySpecialized {
		if pt.Params != nil {
			lhs = pt.Params.Tuple.Variables
			lhs_count = isize(len(pt.Params.Tuple.Variables))
		}
	} else {
		if pt.Params == nil {
			lhs_count = 0
		} else {
			lhs_count = isize(len(pt.Params.Tuple.Variables))
		}
		lhs = make([]*Entity, lhs_count)
		for i := isize(0); i < lhs_count; i++ {
			e := pt.Params.Tuple.Variables[i]
			if !is_type_polymorphic(e.Type) {
				lhs[i] = e
			}
		}
	}

	if lhs_count_ != nil {
		*lhs_count_ = lhs_count
	}

	return lhs
}

func get_procedure_param_count_excluding_defaults(pt *Type, param_count_ *isize) isize {
	if pt == nil {
		gb_assert_handler("Assertion Failure", "pt != nil", "cmd_check_expr_poly.go", 0)
	}
	if pt.Kind != Type_Proc {
		gb_assert_handler("Assertion Failure", "pt->kind == Type_Proc", "cmd_check_expr_poly.go", 0)
	}
	param_count := isize(0)
	param_count_excluding_defaults := isize(0)
	variadic := pt.Proc.Variadic
	var param_tuple *TypeTuple

	if pt.Proc.Params != nil {
		param_tuple = &pt.Proc.Params.Tuple

		param_count = isize(len(param_tuple.Variables))
		if variadic {
			for i := param_count - 1; i >= 0; i-- {
				e := param_tuple.Variables[i]
				if e.Kind == Entity_TypeName {
					break
				}

				if e.Kind == Entity_Variable {
					if e.Variable.ParamValue.Kind != ParameterValue_Invalid {
						param_count--
						continue
					}
				}
				break
			}
			param_count--
		}
	}

	param_count_excluding_defaults = param_count
	if param_tuple != nil {
		for i := param_count - 1; i >= 0; i-- {
			e := param_tuple.Variables[i]
			if e.Kind == Entity_TypeName {
				break
			}

			if e.Kind == Entity_Variable {
				if e.Variable.ParamValue.Kind != ParameterValue_Invalid {
					param_count_excluding_defaults--
					continue
				}
			}
		}
	}

	if param_count_ != nil {
		*param_count_ = param_count
	}
	return param_count_excluding_defaults
}

func lookup_procedure_parameter_in_pt(pt *TypeProc, parameter_name String) isize {
	param_count := isize(pt.ParamCount)
	for i := isize(0); i < param_count; i++ {
		e := pt.Params.Tuple.Variables[i]
		name := e.Token.String
		if is_blank_ident(name) {
			continue
		}
		if name == parameter_name {
			return i
		}
	}
	return -1
}

func lookup_procedure_parameter(type_ *Type, parameter_name String) isize {
	type_ = base_type(type_)
	if type_.Kind != Type_Proc {
		gb_assert_handler("Assertion Failure", "type->kind == Type_Proc", "cmd_check_expr_poly.go", 0)
	}
	return lookup_procedure_parameter_in_pt(&type_.Proc, parameter_name)
}
