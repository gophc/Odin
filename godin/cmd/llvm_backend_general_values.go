package cmd

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

func lb_find_or_add_entity_string_ptr(m *lbModule, str string, customLinkSection bool) LLVMValueRef {
	if !customLinkSection {
		if v, ok := m.ConstStrings[str]; ok {
			return v
		}
	}

	indices := [2]LLVMValueRef{llvm_zero(m), llvm_zero(m)}
	data := LLVMConstStringInContext(m.Ctx, str, uint(len(str)), 0)

	id := m.GlobalArrayIndex.Add(1)
	name := fmt.Sprintf("csbs$%s$%x", m.ModuleName, id)

	typ := LLVMTypeOf(data)
	globalData := LLVMAddGlobal(m.Mod, typ, name)
	LLVMSetInitializer(globalData, data)
	lb_make_global_private_const(globalData)
	LLVMSetAlignment(globalData, 1)

	ptr := LLVMConstInBoundsGEP2(typ, globalData, indices[:], 2)
	if !customLinkSection {
		m.ConstStrings[str] = ptr
	}
	return ptr
}

func lb_find_or_add_entity_string16_ptr(m *lbModule, str string, customLinkSection bool) LLVMValueRef {
	if !customLinkSection {
		if v, ok := m.ConstString16s[hash_string16(str)]; ok {
			return v
		}
	}

	indices := [2]LLVMValueRef{llvm_zero(m), llvm_zero(m)}
	llvmU16 := LLVMInt16TypeInContext(m.Ctx)

	runes := []rune(str)
	values := make([]LLVMValueRef, len(runes)+1)
	for i, r := range runes {
		values[i] = LLVMConstInt(llvmU16, uint64(r), 0)
	}
	values[len(runes)] = LLVMConstInt(llvmU16, 0, 0)

	data := LLVMConstArray(llvmU16, values, uint(len(values)))

	id := m.GlobalArrayIndex.Add(1)
	name := fmt.Sprintf("csbs$%s$%x", m.ModuleName, id)

	typ := LLVMTypeOf(data)
	globalData := LLVMAddGlobal(m.Mod, typ, name)
	LLVMSetInitializer(globalData, data)
	lb_make_global_private_const(globalData)
	LLVMSetAlignment(globalData, 2)

	ptr := LLVMConstInBoundsGEP2(typ, globalData, indices[:], 2)
	if !customLinkSection {
		m.ConstString16s[hash_string16(str)] = ptr
	}
	return ptr
}

func lb_find_or_add_entity_string(m *lbModule, str string, customLinkSection bool) lbValue {
	var ptr LLVMValueRef
	if len(str) != 0 {
		ptr = lb_find_or_add_entity_string_ptr(m, str, customLinkSection)
	} else {
		ptr = LLVMConstNull(lb_type(m, t_u8_ptr))
	}
	strLen := LLVMConstInt(lb_type(m, t_int), uint64(len(str)), 1)

	res := lbValue{}
	res.Value = llvm_const_string_internal(m, t_string, ptr, strLen)
	res.Type = t_string
	return res
}

func lb_find_or_add_entity_string_byte_slice_with_type(m *lbModule, str string, sliceType *Type) lbValue {
	if !is_type_slice(sliceType) {
		panic("expected slice type")
	}

	indices := [2]LLVMValueRef{llvm_zero(m), llvm_zero(m)}
	data := LLVMConstStringInContext(m.Ctx, str, uint(len(str)), 0)

	id := m.GlobalArrayIndex.Add(1)
	name := fmt.Sprintf("csba$%s$%x", m.ModuleName, id)

	typ := LLVMTypeOf(data)
	globalData := LLVMAddGlobal(m.Mod, typ, name)
	LLVMSetInitializer(globalData, data)
	lb_make_global_private_const(globalData)
	LLVMSetAlignment(globalData, 1)

	dataLen := int64(len(str))
	var ptr LLVMValueRef
	if dataLen != 0 {
		ptr = LLVMConstInBoundsGEP2(typ, globalData, indices[:], 2)
	} else {
		ptr = LLVMConstNull(lb_type(m, t_u8_ptr))
	}
	align := MINIMUM_SLICE_ALIGNMENT
	if !is_type_u8_slice(sliceType) {
		bt := base_type(sliceType)
		elem := bt.Slice.Elem
		sz := type_size_of(elem)
		align = max64(type_align_of(elem), align)
		if sz <= 0 || align <= 0 {
			panic("invalid slice element")
		}
		LLVMSetAlignment(globalData, uint32(align))
		ptr = LLVMConstPointerCast(ptr, lb_type(m, alloc_type_pointer(elem)))
		dataLen /= sz
	}

	lenVal := LLVMConstInt(lb_type(m, t_int), uint64(dataLen), 1)
	values := [2]LLVMValueRef{ptr, lenVal}

	res := lbValue{}
	res.Value = llvm_const_named_struct(m, sliceType, values[:], 2)
	res.Type = sliceType
	return res
}

func lb_find_or_add_entity_string16_slice_with_type(m *lbModule, str string, sliceType *Type) lbValue {
	if !is_type_slice(sliceType) {
		panic("expected slice type")
	}

	indices := [2]LLVMValueRef{llvm_zero(m), llvm_zero(m)}
	llvmU16 := LLVMInt16TypeInContext(m.Ctx)

	runes := []rune(str)
	values := make([]LLVMValueRef, len(runes)+1)
	for i, r := range runes {
		values[i] = LLVMConstInt(llvmU16, uint64(r), 0)
	}
	values[len(runes)] = LLVMConstInt(llvmU16, 0, 0)

	data := LLVMConstArray(llvmU16, values, uint(len(values)))

	id := m.GlobalArrayIndex.Add(1)
	name := fmt.Sprintf("csba$%s$%x", m.ModuleName, id)

	typ := LLVMTypeOf(data)
	globalData := LLVMAddGlobal(m.Mod, typ, name)
	LLVMSetInitializer(globalData, data)
	lb_make_global_private_const(globalData)
	LLVMSetAlignment(globalData, 2)

	dataLen := int64(len(runes))
	var ptr LLVMValueRef
	if dataLen != 0 {
		ptr = LLVMConstInBoundsGEP2(typ, globalData, indices[:], 2)
	} else {
		ptr = LLVMConstNull(lb_type(m, t_u8_ptr))
	}
	align := MINIMUM_SLICE_ALIGNMENT
	if !is_type_u16_slice(sliceType) {
		bt := base_type(sliceType)
		elem := bt.Slice.Elem
		sz := type_size_of(elem)
		align = max64(type_align_of(elem), align)
		if sz <= 0 || align <= 0 {
			panic("invalid slice element")
		}
		LLVMSetAlignment(globalData, uint32(align))
		ptr = LLVMConstPointerCast(ptr, lb_type(m, alloc_type_pointer(elem)))
		dataLen /= sz
	}

	lenVal := LLVMConstInt(lb_type(m, t_int), uint64(dataLen), 1)
	val := [2]LLVMValueRef{ptr, lenVal}

	res := lbValue{}
	res.Value = llvm_const_named_struct(m, sliceType, val[:], 2)
	res.Type = sliceType
	return res
}

func lb_find_ident(p *lbProcedure, m *lbModule, e *Entity, expr *Ast) lbValue {
	if e.Flags&EntityFlag_Param != 0 {
		if found, ok := p.DirectParameters[unsafe.Pointer(e)]; ok {
			return found
		}
	}

	m.ValuesMutex.RLock()
	found, ok := m.Values[unsafe.Pointer(e)]
	m.ValuesMutex.RUnlock()

	if ok {
		v := found
		if is_type_proc(v.Type) {
			return v
		}
		return lb_emit_load(p, v)
	} else if e != nil && e.Kind == EntityVariable {
		return lb_addr_load(p, lb_build_addr(p, expr))
	}

	if e.Kind == EntityProcedure {
		return lb_find_procedure_value_from_entity(m, e)
	}
	if USE_SEPARATE_MODULES {
		otherModule := lb_module_of_entity(m.Gen, e, m)
		if otherModule != m {
			name := lb_get_entity_name(otherModule, e)
			lb_set_entity_from_other_modules_linkage_correctly(otherModule, e, name)

			g := lbValue{}
			g.Value = LLVMAddGlobal(m.Mod, lb_type(m, e.Type), name)
			g.Type = alloc_type_pointer(e.Type)
			LLVMSetLinkage(g.Value, LLVMExternalLinkage)

			m.ValuesMutex.Lock()
			m.Values[unsafe.Pointer(e)] = g
			m.ValuesMutex.Unlock()
			m.Members[name] = g
			return lb_emit_load(p, g)
		}
	}

	var pkg string
	if e.Pkg != nil {
		pkg = e.Pkg.Name
	}
	_ = pkg
	panic(fmt.Sprintf("nullptr value for expression from identifier: %s.%s", pkg, e.Token.String))
}

func lb_find_procedure_value_from_entity(m *lbModule, e *Entity) lbValue {
	gen := m.Gen

	if e == nil {
		panic("entity is nil")
	}
	if !is_type_proc(e.Type) {
		panic("entity type is not a procedure")
	}
	e = strip_entity_wrapping(e)
	if e == nil {
		panic("entity is nil after stripping wrapping")
	}
	if e.Kind != EntityProcedure {
		panic("entity is not a procedure")
	}

	m.ValuesMutex.RLock()
	found, ok := m.Values[unsafe.Pointer(e)]
	m.ValuesMutex.RUnlock()
	if ok {
		return found
	}

	ignoreBody := false
	otherModule := m
	if USE_SEPARATE_MODULES {
		otherModule = lb_module_of_entity(gen, e, m)
	}
	if otherModule == m {
		_ = fmt.Sprintf("Missing Procedure (lb_find_procedure_value_from_entity): %s module %p", e.Token.String, m)
	}
	ignoreBody = otherModule != m

	missingProc := lb_create_procedure(m, e, ignoreBody)
	if missingProc == nil {
		return lbValue{}
	}

	if ignoreBody {
		if otherModule != nil {
			otherModule.ValuesMutex.RLock()
			_, foundInOther := otherModule.Values[unsafe.Pointer(e)]
			otherModule.ValuesMutex.RUnlock()
			if !foundInOther {
				missingProcInOtherModule := lb_create_procedure(otherModule, e, false)
				_ = missingProcInOtherModule
				// mpsc_enqueue(&otherModule.missing_procedures_to_check, missingProcInOtherModule)
				otherModule.MissingProceduresToCheck = append(otherModule.MissingProceduresToCheck, missingProcInOtherModule)
			}
		}
	} else {
		// mpsc_enqueue(&m->missing_procedures_to_check, missingProc)
		m.MissingProceduresToCheck = append(m.MissingProceduresToCheck, missingProc)
	}

	m.ValuesMutex.RLock()
	found, ok = m.Values[unsafe.Pointer(e)]
	m.ValuesMutex.RUnlock()
	if ok {
		return found
	}

	panic(fmt.Sprintf("missing procedure %s", e.Token.String))
}

var lb_name_id atomic.Int32

func lb_generate_anonymous_proc_lit(m *lbModule, prefixName string, expr *Ast, parent ...*lbProcedure) lbValue {
	gen := m.Gen
	_ = gen

	pl := expr.ProcLit

	if pl.Decl.Entity.Load() != nil {
		return lb_find_procedure_value_from_entity(m, pl.Decl.Entity.Load())
	}

	pos := ast_token(expr).Pos
	_ = pos

	nameID := lb_name_id.Add(1)
	name := fmt.Sprintf("%s$anon-%d", prefixName, 1+nameID)

	typ := type_of_expr(expr)

	if pl.Decl.Entity != nil {
		panic("entity should be nil")
	}
	token := Token{}
	token.Pos = ast_token(expr).Pos
	token.Kind = Token_Ident
	token.String = name
	e := alloc_entity_procedure(nil, token, typ, pl.Tags)
	e.File = expr.File()
	e.Scope = e.File.Scope

	targetModule := m
	if targetModule == nil {
		panic("target module is nil")
	}

	pl.Decl.CodeGenModule = targetModule
	e.DeclInfo = pl.Decl
	e.ParentProcDecl = pl.Decl.Parent
	e.Procedure.IsAnonymous = true
	e.Flags |= EntityFlag_ProcBodyChecked

	pl.Decl.Entity.Store(e)

	if targetModule != m {
		targetModule.ValuesMutex.RLock()
		_, found := targetModule.Values[unsafe.Pointer(e)]
		targetModule.ValuesMutex.RUnlock()
		if !found {
			missingProcInTargetModule := lb_create_procedure(targetModule, e, false)
			_ = missingProcInTargetModule
			// mpsc_enqueue(&targetModule.missing_procedures_to_check, missingProcInTargetModule)
			targetModule.MissingProceduresToCheck = append(targetModule.MissingProceduresToCheck, missingProcInTargetModule)
		}

		p := lb_create_procedure(m, e, true)
		value := lbValue{}
		value.Value = p.Value
		value.Type = p.Type
		return value
	} else {
		p := lb_create_procedure(m, e)

		value := lbValue{}
		value.Value = p.Value
		value.Type = p.Type

		// mpsc_enqueue(&m->procedures_to_generate, p)
		m.ProceduresToGenerate = append(m.ProceduresToGenerate, p)
		if len(parent) > 0 && parent[0] != nil {
			parent[0].Children = append(parent[0].Children, p)
		} else {
			m.Members[name] = value
		}
		return value
	}
}

func lb_add_global_generated_with_name(m *lbModule, typ *Type, value lbValue, name string, entityOut ...**Entity) lbAddr {
	if len(name) == 0 {
		panic("name is empty")
	}
	if typ == nil {
		panic("type is nil")
	}
	typ = default_type(typ)

	actualType := lb_type(m, typ)
	if value.Value != 0 {
		valueType := LLVMTypeOf(value.Value)
		if lb_sizeof(actualType) != lb_sizeof(valueType) {
			panic(fmt.Sprintf("type size mismatch: %d vs %d", lb_sizeof(actualType), lb_sizeof(valueType)))
		}
		actualType = valueType
	}

	e := alloc_entity_variable(nil, make_token_ident(name), typ)
	g := lbValue{}
	g.Type = alloc_type_pointer(typ)
	g.Value = LLVMAddGlobal(m.Mod, actualType, name)
	if value.Value != 0 {
		if !LLVMIsConstant(value.Value) {
			panic("value is not constant")
		}
		LLVMSetInitializer(g.Value, value.Value)
	} else {
		LLVMSetInitializer(g.Value, LLVMConstNull(lb_type(m, typ)))
	}

	g.Value = LLVMConstPointerCast(g.Value, lb_type(m, g.Type))

	m.ValuesMutex.Lock()
	m.Values[unsafe.Pointer(e)] = g
	m.ValuesMutex.Unlock()
	m.Members[name] = g

	if len(entityOut) > 0 && entityOut[0] != nil {
		*entityOut[0] = e
	}

	return lb_addr(g)
}

func lb_add_global_generated_from_procedure(p *lbProcedure, typ *Type, value lbValue) lbAddr {
	if typ == nil {
		panic("type is nil")
	}
	typ = default_type(typ)

	index := atomic.AddUint32(&p.GlobalGeneratedIndex, 1)

	name := fmt.Sprintf("ggv$%s$%d", p.Name, index)
	return lb_add_global_generated_with_name(p.Module, typ, value, name)
}

func lb_find_runtime_value(m *lbModule, name string) lbValue {
	pkg := m.Info.RuntimePackage
	e := scope_lookup_current(pkg.Scope, string_interner_insert(name))
	return lb_find_value_from_entity(m, e)
}

func lb_find_package_value(m *lbModule, pkg string, name string) lbValue {
	e := find_entity_in_pkg(m.Info, pkg, name)
	return lb_find_value_from_entity(m, e)
}

func lb_generate_local_array(p *lbProcedure, elemType *Type, count int64, zeroInit ...bool) lbValue {
	zero := false
	if len(zeroInit) > 0 {
		zero = zeroInit[0]
	}
	addr := lb_add_local_generated(p, alloc_type_array(elemType, count), zero)
	return lb_addr_get_ptr(p, addr)
}

func lb_find_value_from_entity(m *lbModule, e *Entity) lbValue {
	e = strip_entity_wrapping(e)
	if e == nil {
		panic("entity is nil after stripping wrapping")
	}

	if e.Token.String == "_" {
		panic("entity name is _")
	}

	if e.Kind == EntityProcedure {
		return lb_find_procedure_value_from_entity(m, e)
	}

	m.ValuesMutex.RLock()
	found, ok := m.Values[unsafe.Pointer(e)]
	m.ValuesMutex.RUnlock()
	if ok {
		return found
	}

	if USE_SEPARATE_MODULES {
		otherModule := lb_module_of_entity(m.Gen, e, m)

		isExternal := otherModule != m
		if !isExternal {
			if e.CodeGenModule != nil && e.CodeGenModule.Load() != nil {
				otherModule = e.CodeGenModule.Load()
			} else {
				otherModule = &m.Gen.DefaultModule
			}
			isExternal = otherModule != m
		}

		if isExternal {
			name := lb_get_entity_name(otherModule, e)

			g := lbValue{}
			g.Value = LLVMAddGlobal(m.Mod, lb_type(m, e.Type), name)
			g.Type = alloc_type_pointer(e.Type)

			LLVMSetLinkage(g.Value, LLVMExternalLinkage)

			m.ValuesMutex.Lock()
			m.Values[unsafe.Pointer(e)] = g
			m.ValuesMutex.Unlock()
			m.Members[name] = g

			lb_set_entity_from_other_modules_linkage_correctly(otherModule, e, name)

			lb_apply_thread_local_model(g.Value, e.Variable.ThreadLocalModel)

			return g
		}
	}

	panic(fmt.Sprintf("missing value '%s' in module %s", e.Token.String, m.ModuleName))
}

func lb_generate_global_array(m *lbModule, elemType *Type, count int64, prefix string, id int64) lbValue {
	name := fmt.Sprintf("%s-%d", prefix, id)

	t := alloc_type_array(elemType, count)
	g := lbValue{}
	g.Value = LLVMAddGlobal(m.Mod, lb_type(m, t), name)
	g.Type = alloc_type_pointer(t)
	LLVMSetInitializer(g.Value, LLVMConstNull(lb_type(m, t)))
	LLVMSetLinkage(g.Value, LLVMPrivateLinkage)
	m.Members[name] = g
	return g
}

func lb_build_cond(p *lbProcedure, cond *Ast, trueBlock *lbBlock, falseBlock *lbBlock) lbValue {
	if cond == nil || trueBlock == nil || falseBlock == nil {
		panic("nil argument to lb_build_cond")
	}

	switch cond.Kind {
	case Ast_ParenExpr:
		pe := cond.ParenExpr
		return lb_build_cond(p, pe.Expr, trueBlock, falseBlock)

	case Ast_UnaryExpr:
		ue := cond.UnaryExpr
		if ue.Op.Kind == Token_Not {
			condVal := lb_build_cond(p, ue.Expr, falseBlock, trueBlock)
			if condVal.Value != 0 && LLVMIsConstant(condVal.Value) {
				return lb_const_bool(p.Module, condVal.Type, LLVMConstIntGetZExtValue(condVal.Value) == 0)
			}
			return lbValue{}
		}

	case Ast_BinaryExpr:
		be := cond.BinaryExpr
		if be.Op.Kind == Token_CmpAnd {
			block := lb_create_block(p, "cmp.and")
			lb_build_cond(p, be.Left, block, falseBlock)
			lb_start_block(p, block)
			lb_build_cond(p, be.Right, trueBlock, falseBlock)
			return lbValue{}
		} else if be.Op.Kind == Token_CmpOr {
			block := lb_create_block(p, "cmp.or")
			lb_build_cond(p, be.Left, trueBlock, block)
			lb_start_block(p, block)
			lb_build_cond(p, be.Right, trueBlock, falseBlock)
			return lbValue{}
		}
	}

	var v lbValue
	if lb_is_expr_untyped_const(cond) {
		v = lb_expr_untyped_const_to_typed(p.Module, cond, t_llvm_bool)
	} else {
		v = lb_build_expr(p, cond)
	}

	v = lb_emit_conv(p, v, t_llvm_bool)

	lb_emit_if(p, v, trueBlock, falseBlock)

	return v
}

func lb_add_local(p *lbProcedure, typ *Type, e *Entity, zeroInit bool, forceNoInit bool) lbAddr {
	if p.DeclBlock == p.CurrBlock {
		panic("lb_add_local called in decl block")
	}
	LLVMPositionBuilderAtEnd(p.Builder, p.DeclBlock.Block)

	name := ""
	if e != nil && len(e.Token.String) > 0 && e.Token.String != "_" {
		name = e.Token.String
	}

	llvmType := lb_type(p.Module, typ)

	alignment := uint(max64(type_align_of(typ), lb_alignof(llvmType)))
	if is_type_matrix(typ) {
		alignment *= 2
	}

	ptr := llvm_alloca(p, llvmType, int64(alignment), name)

	if !zeroInit && !forceNoInit {
		kind := LLVMGetTypeKind(llvmType)
		if kind == LLVMArrayTypeKind {
			kind = LLVMGetTypeKind(lb_type(p.Module, core_array_type(typ)))
		}

		if kind == LLVMStructTypeKind {
			sz := type_size_of(typ)
			if type_size_of_struct_pretend_is_packed(typ) != sz {
				zeroInit = true
			}
		}
	}

	val := lbValue{}
	val.Value = ptr
	val.Type = alloc_type_pointer(typ)

	if e != nil {
		p.Module.ValuesMutex.Lock()
		p.Module.Values[unsafe.Pointer(e)] = val
		p.Module.ValuesMutex.Unlock()
		lb_add_debug_local_variable(p, ptr, typ, e.Token)

		if (buildContext.SanitizerFlags&SanitizerFlag_Address) != 0 && !p.Entity.Procedure.NoSanitizeAddress {
			p.AsanStackLocals = append(p.AsanStackLocals, val)
		}
	}

	if zeroInit {
		lb_mem_zero_ptr(p, ptr, typ, alignment)
	}

	return lb_addr(val)
}

func lb_add_local_generated(p *lbProcedure, typ *Type, zeroInit bool) lbAddr {
	llvmType := lb_type(p.Module, typ)

	alignment := uint(max64(type_align_of(typ), lb_alignof(llvmType)))
	if is_type_matrix(typ) {
		alignment *= 2
	}

	ptr := llvm_alloca(p, llvmType, int64(alignment), "")

	if zeroInit {
		lb_mem_zero_ptr(p, ptr, typ, alignment)
	}

	val := lbValue{}
	val.Value = ptr
	val.Type = alloc_type_pointer(typ)
	return lb_addr(val)
}

func lb_add_local_generated_temp(p *lbProcedure, typ *Type, minAlignment int64) lbAddr {
	res := lb_add_local(p, typ, nil, false, true)
	lb_try_update_alignment(res.Addr, uint(minAlignment))
	return res
}

func lb_set_linkage_from_entity_flags(m *lbModule, value LLVMValueRef, flags uint64) {
	if flags&EntityFlag_CustomLinkage_Internal != 0 {
		LLVMSetLinkage(value, LLVMInternalLinkage)
	} else if flags&EntityFlag_CustomLinkage_Strong != 0 {
		LLVMSetLinkage(value, LLVMExternalLinkage)
	} else if flags&EntityFlag_CustomLinkage_Weak != 0 {
		LLVMSetLinkage(value, LLVMExternalWeakLinkage)
	} else if flags&EntityFlag_CustomLinkage_LinkOnce != 0 {
		LLVMSetLinkage(value, LLVMLinkOnceAnyLinkage)
	}
}

func llvm_const_string_internal(m *lbModule, t *Type, data LLVMValueRef, lenVal LLVMValueRef) LLVMValueRef {
	if buildContext.PtrSize < buildContext.IntSize {
		values := [3]LLVMValueRef{
			data,
			LLVMConstNull(lb_type(m, t_i32)),
			lenVal,
		}
		return llvm_const_named_struct_internal(m, lb_type(m, t), values[:], 3)
	} else {
		values := [2]LLVMValueRef{
			data,
			lenVal,
		}
		return llvm_const_named_struct_internal(m, lb_type(m, t), values[:], 2)
	}
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func hash_string16(s string) uint64 {
	h := uint64(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint64(s[i])
	}
	return h
}
