package cmd

import (
	"fmt"
	"unsafe"
)

// =============================================================================
// lb_call_sret_eligible - checks if a call expression is eligible for sret optimization
// =============================================================================
func lb_call_sret_eligible(p *lbProcedure, call_expr *Ast, dst_type *Type) *lbFunctionType {
	if call_expr.Kind != Ast_CallExpr {
		return nil
	}
	proc_expr := unparen_expr(call_expr.CallExpr.Proc)
	proc_tv := type_and_value_of_expr(proc_expr)
	if proc_tv.Mode == Addressing_Type || proc_tv.Mode == Addressing_Builtin {
		return nil
	}
	pt := base_type(proc_tv.Type)
	if pt == nil || pt.Kind != Type_Proc || pt.Proc.Results == nil {
		return nil
	}
	callee_ft := lb_get_function_type(p.Module, pt)
	if callee_ft.Ret.Kind != lbArg_Indirect {
		return nil
	}
	callee_ret := reduce_tuple_to_single_type(pt.Proc.Results)
	if callee_ret == nil || !are_types_identical(dst_type, callee_ret) {
		return nil
	}
	return callee_ft
}

// =============================================================================
// lb_scan_for_sret_rvo - scans a procedure body for RVO opportunities
// =============================================================================
func lb_scan_for_sret_rvo(p *lbProcedure) {
	if p.Body == nil || p.Body.Kind != Ast_BlockStmt {
		return
	}
	proc_type := p.Type
	if proc_type.Proc.ResultCount != 1 || proc_type.Proc.Results == nil {
		return
	}
	ft := lb_get_function_type(p.Module, proc_type)
	if ft.Ret.Kind != lbArg_Indirect {
		return
	}
	stmts := p.Body.BlockStmt.Stmts
	if len(stmts) < 2 {
		return
	}
	last := stmts[len(stmts)-1]
	if last.Kind != Ast_ReturnStmt {
		return
	}
	results := last.ReturnStmt.Results
	if len(results) != 1 {
		return
	}
	ret_expr := unparen_expr(results[0])
	if ret_expr.Kind != Ast_Ident {
		return
	}
	ret_entity := entity_of_node(ret_expr)
	if ret_entity == nil || ret_entity.Kind != Entity_Variable {
		return
	}
	ret_type := reduce_tuple_to_single_type(proc_type.Proc.Results)
	if ret_type == nil || !are_types_identical(ret_entity.Type, ret_type) {
		return
	}
	decl_index := int64(-1)
	for i := int64(len(stmts) - 2); i >= 0; i-- {
		stmt := stmts[i]
		switch stmt.Kind {
		case Ast_ValueDecl:
			vd := &stmt.ValueDecl
			if !vd.IsMutable {
				continue
			}
			if len(vd.Names) == 1 && len(vd.Values) == 1 {
				e := entity_of_node(vd.Names[0])
				if e == ret_entity {
					rhs := unparen_expr(vd.Values[0])
					if rhs.Kind == Ast_CallExpr && lb_call_sret_eligible(p, rhs, e.Type) {
						decl_index = i
					}
					goto done_scanning
				}
			}
			for _, name := range vd.Names {
				if entity_of_node(name) == ret_entity {
					goto done_scanning
				}
			}
			continue
		case Ast_ExprStmt:
			continue
		case Ast_AssignStmt:
			for _, lhs := range stmt.AssignStmt.LHS {
				l := unparen_expr(lhs)
				if l.Kind == Ast_Ident && entity_of_node(l) == ret_entity {
					goto done_scanning
				}
			}
			continue
		default:
			goto done_scanning
		}
	}
done_scanning:
	if decl_index >= 0 {
		p.SretRvoEntity = ret_entity
	}
}

// =============================================================================
// lb_build_constant_value_decl - builds constant value declarations
// =============================================================================
func lb_build_constant_value_decl(p *lbProcedure, vd *AstValueDecl) {
	if vd == nil || vd.IsMutable {
		return
	}
	for _, ident := range vd.Names {
		e := entity_of_node(ident)
		if e.Kind != Entity_TypeName {
			continue
		}
		polymorphic_struct := false
		if e.Type != nil && e.Kind == Entity_TypeName {
			bt := base_type(e.Type)
			if bt.Kind == Type_Struct {
				polymorphic_struct = bt.Struct.IsPolymorphic
			}
		}
		if !polymorphic_struct && e.MinDepCount.Load() == 0 {
			continue
		}
		if e.TypeName.IrMangledName.Len != 0 {
			continue
		}
		_ = lb_get_entity_name(p.Module, e)
	}
	for i := 0; i < len(vd.Names); i++ {
		ident := vd.Names[i]
		e := entity_of_node(ident)
		if e.Kind != Entity_Procedure {
			continue
		}
		value := unparen_expr(vd.Values[i])
		if value.Kind != Ast_ProcLit {
			continue
		}
		decl := decl_info_of_entity(e)
		pl := &decl.ProcLit.ProcLit
		if pl.Body != nil {
			gpd := e.Procedure.GenProcs
			if gpd != nil {
				gpd.Mutex.RLock()
				for _, procEntity := range gpd.Procs {
					if procEntity.MinDepCount.Load() == 0 {
						continue
					}
					d := decl_info_of_entity(procEntity)
					lb_build_nested_proc(p, &d.ProcLit.ProcLit, procEntity)
				}
				gpd.Mutex.RUnlock()
			} else {
				lb_build_nested_proc(p, pl, e)
			}
		} else {
			original_name := e.Token.String
			name := original_name
			if e.Procedure.IsForeign {
				lb_add_foreign_library_path(p.Module, e.Procedure.ForeignLibrary)
			}
			if e.Procedure.LinkName.Len > 0 {
				name = e.Procedure.LinkName
			}
			prev_value := string_map_get(&p.Module.Members, name)
			if prev_value != nil {
				return
			}
			e.Procedure.LinkName = name
			nested_proc := lb_create_procedure(p.Module, e)
			value_val := lbValue{}
			value_val.Value = nested_proc.Value
			value_val.Type = nested_proc.Type
			mpsc_enqueue(&p.Module.ProceduresToGenerate, nested_proc)
			p.Children = append(p.Children, nested_proc)
			string_map_set(&p.Module.Members, name, value_val)
		}
	}
}

// =============================================================================
// lb_build_static_variables - builds static variable declarations
// =============================================================================
func lb_build_static_variables(p *lbProcedure, vd *AstValueDecl) {
	for i := 0; i < len(vd.Names); i++ {
		var value lbValue
		ident := vd.Names[i]
		e := entity_of_node(ident)
		name_str_raw := e.Token.String

		if len(vd.Values) > 0 {
			ast_value := vd.Values[i]
			cc := LB_CONST_CONTEXT_DEFAULT_NO_LOCAL
			if e.Variable.IsRodata {
				cc.IsRodata = true
			}
			value = lb_const_value(p.Module, ast_value.TAV.Type, ast_value.TAV.Value, cc)
		}
		mangled_name := fmt.Sprintf("%s-.%s-%d", goStr(p.Name), goStr(name_str_raw), e.ID)
		mangled_name_str := S(mangled_name)
		c_name := alloc_cstring(permanent_allocator(), mangled_name_str)
		global := LLVMAddGlobal(p.Module.Mod, lb_type(p.Module, e.Type), c_name)
		LLVMSetAlignment(global, uint32(type_align_of(e.Type)))
		LLVMSetInitializer(global, LLVMConstNull(lb_type(p.Module, e.Type)))
		if e.Variable.IsRodata {
			LLVMSetGlobalConstant(global, true)
		}
		if !lb_apply_thread_local_model(global, e.Variable.ThreadLocalModel) {
			LLVMSetLinkage(global, LLVMInternalLinkage)
		}
		if value.Value != 0 {
			if is_type_any(e.Type) {
				var_type := default_type(value.Type)
				var_name := "__$static_any::" + mangled_name
				var_global := lb_add_global_generated_with_name(p.Module, var_type, value, var_name, nil)
				var_global_ref := var_global.Addr.Value
				if e.Variable.IsRodata {
					LLVMSetGlobalConstant(var_global_ref, true)
				}
				if !lb_apply_thread_local_model(var_global_ref, e.Variable.ThreadLocalModel) {
					LLVMSetLinkage(var_global_ref, LLVMInternalLinkage)
				}
				vals := make([]LLVMValueRef, 0, 3)
				vals = append(vals, lb_emit_conv(p, var_global.Addr, t_rawptr).Value)
				if buildContext.Metrics.PtrSize == 4 {
					vals = append(vals, LLVMConstNull(lb_type_padding_filler(p.Module, 4, 4)))
				}
				vals = append(vals, lb_typeid(p.Module, var_type).Value)
				init := llvm_const_named_struct(p.Module, e.Type, vals, isize(len(vals)))
				LLVMSetInitializer(global, init)
			} else {
				LLVMSetInitializer(global, value.Value)
			}
		}
		global_val := lbValue{Value: global, Type: alloc_type_pointer(e.Type)}
		lb_add_entity(p.Module, e, global_val)
		lb_add_member(p.Module, mangled_name_str, global_val)
	}
}

// =============================================================================
// lb_append_tuple_values - appends tuple values to a destination array
// =============================================================================
func lb_append_tuple_values(p *lbProcedure, dst_values *[]lbValue, src_value lbValue) isize {
	init_count := len(*dst_values)
	t := src_value.Type
	if t.Kind == Type_Tuple {
		tf, ok := p.TupleFixMap[src_value.Value]
		if ok {
			for _, v := range tf.Values {
				*dst_values = append(*dst_values, v)
			}
		} else {
			for i := 0; i < len(t.Tuple.Variables); i++ {
				v := lb_emit_tuple_ev(p, src_value, int32(i))
				*dst_values = append(*dst_values, v)
			}
		}
	} else {
		*dst_values = append(*dst_values, src_value)
	}
	return isize(len(*dst_values) - init_count)
}

// =============================================================================
// lb_build_assignment - builds assignment statement evaluating RHS and storing to LHS
// =============================================================================
func lb_build_assignment(p *lbProcedure, lvals []lbAddr, values []*Ast) {
	if len(values) == 0 {
		return
	}
	inits := make([]lbValue, 0, len(lvals))
	for _, rhs := range values {
		init := lb_build_expr(p, rhs)
		lb_append_tuple_values(p, &inits, init)
	}
	prev_in_assignment := p.InMultiAssignment
	lval_count := 0
	for _, lval := range lvals {
		if lval.Addr.Value != 0 {
			lval_count++
		}
	}
	p.InMultiAssignment = lval_count > 1
	if len(lvals) != len(inits) {
		panic("lvals.count != inits.count")
	}
	for i, init := range inits {
		lval := lvals[i]
		lb_addr_store(p, lval, init)
	}
	p.InMultiAssignment = prev_in_assignment
}

// =============================================================================
// lb_build_return_stmt_internal - internal return statement builder
// =============================================================================
func lb_build_return_stmt_internal(p *lbProcedure, res lbValue, pos TokenPos) {
	ft := lb_get_function_type(p.Module, p.Type)
	return_by_pointer := ft.Ret.Kind == lbArg_Indirect
	split_returns := ft.MultipleReturnOriginalType != nil

	if return_by_pointer {
		if res.Value != 0 {
			res_val := res.Value
			sz := type_size_of(res.Type)
			if LLVMIsALoadInst(res_val) != 0 && sz > buildContext.IntSize {
				ptr := lb_address_from_load_or_generate_local(p, res)
				lb_mem_copy_non_overlapping(p, p.ReturnPtr.Addr, ptr, lb_const_int(p.Module, t_int, sz))
			} else {
				LLVMBuildStore(p.Builder, res_val, p.ReturnPtr.Addr.Value)
			}
		} else {
			LLVMBuildStore(p.Builder, LLVMConstNull(p.AbiFunctionType.Ret.Type), p.ReturnPtr.Addr.Value)
		}
		lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
		instr := LLVMGetLastInstruction(p.CurrBlock.Block)
		if !lb_is_instr_terminating(instr) {
			LLVMBuildRetVoid(p.Builder)
		}
	} else {
		ret_val := res.Value
		ret_type := p.AbiFunctionType.Ret.Type
		if cast_type := p.AbiFunctionType.Ret.CastType; cast_type != 0 {
			ret_type = cast_type
		}
		if LLVMGetTypeKind(ret_type) == LLVMStructTypeKind {
			src_type := LLVMTypeOf(ret_val)
			if p.TempCalleeReturnStructMemory == 0 {
				max_align := max(lb_alignof(ret_type), lb_alignof(src_type))
				p.TempCalleeReturnStructMemory = llvm_alloca(p, ret_type, max_align)
			}
			ptr := p.TempCalleeReturnStructMemory
			nptr := LLVMBuildPointerCast(p.Builder, ptr, LLVMPointerType(src_type, 0), "")
			LLVMBuildStore(p.Builder, ret_val, nptr)
			ret_val = OdinLLVMBuildLoad(p, ret_type, ptr)
		} else {
			ret_val = OdinLLVMBuildTransmute(p, ret_val, ret_type)
		}
		lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
		instr := LLVMGetLastInstruction(p.CurrBlock.Block)
		if !lb_is_instr_terminating(instr) {
			LLVMBuildRet(p.Builder, ret_val)
		}
	}
}

// =============================================================================
// lb_build_return_stmt - full return statement builder
// =============================================================================
func lb_build_return_stmt(p *lbProcedure, return_results []*Ast, pos TokenPos) {
	lb_ensure_abi_function_type(p.Module, p)
	return_count := int(p.Type.Proc.ResultCount)
	if return_count == 0 {
		lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
		instr := LLVMGetLastInstruction(p.CurrBlock.Block)
		if !lb_is_instr_terminating(instr) {
			LLVMBuildRetVoid(p.Builder)
		}
		return
	}
	var res lbValue
	tuple := &p.Type.Proc.Results.Tuple
	res_count := len(return_results)
	ft := lb_get_function_type(p.Module, p.Type)
	return_by_pointer := ft.Ret.Kind == lbArg_Indirect

	if return_count == 1 {
		e := tuple.Variables[0]
		if res_count == 1 && return_by_pointer {
			ret_expr := unparen_expr(return_results[0])
			if len(p.DeferStmts) == 0 {
				if ret_expr.Kind == Ast_CallExpr && lb_call_sret_eligible(p, ret_expr, e.Type) {
					sret_ptr := p.ReturnPtr.Addr
					lb_build_call_expr(p, ret_expr, &sret_ptr)
					if p.Type.Proc.HasNamedResults && e.Token.String.Len != 0 {
						res_val := lb_emit_load(p, p.ReturnPtr.Addr)
						p.Module.ValuesMutex.RLock()
						found := p.Module.Values[unsafe.Pointer(e)]
						p.Module.ValuesMutex.RUnlock()
						lb_emit_store(p, found, lb_emit_conv(p, res_val, e.Type))
					}
					LLVMBuildRetVoid(p.Builder)
					return
				}
			}
			if p.SretRvoEntity != nil {
				if ret_expr.Kind == Ast_Ident {
					ret_e := entity_of_node(ret_expr)
					if ret_e == p.SretRvoEntity {
						lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
						instr := LLVMGetLastInstruction(p.CurrBlock.Block)
						if !lb_is_instr_terminating(instr) {
							LLVMBuildRetVoid(p.Builder)
						}
						return
					}
				}
			}
		}
		if res_count == 0 {
			p.Module.ValuesMutex.RLock()
			found := p.Module.Values[unsafe.Pointer(e)]
			p.Module.ValuesMutex.RUnlock()
			res = lb_emit_load(p, found)
		} else {
			res = lb_build_expr(p, return_results[0])
			res = lb_emit_conv(p, res, e.Type)
		}
		if p.Type.Proc.HasNamedResults {
			if e.Token.String.Len != 0 {
				p.Module.ValuesMutex.RLock()
				found := p.Module.Values[unsafe.Pointer(e)]
				p.Module.ValuesMutex.RUnlock()
				lb_emit_store(p, found, lb_emit_conv(p, res, e.Type))
			}
		}
	} else {
		results := make([]lbValue, 0, return_count)
		if res_count != 0 {
			for res_index := 0; res_index < res_count; res_index++ {
				r := lb_build_expr(p, return_results[res_index])
				lb_append_tuple_values(p, &results, r)
			}
		} else {
			for res_index := 0; res_index < return_count; res_index++ {
				e := tuple.Variables[res_index]
				p.Module.ValuesMutex.RLock()
				found := p.Module.Values[unsafe.Pointer(e)]
				p.Module.ValuesMutex.RUnlock()
				r := lb_emit_load(p, found)
				results = append(results, r)
			}
		}
		if len(results) != return_count {
			panic("results.count != return_count")
		}
		if p.Type.Proc.HasNamedResults {
			named_results := make([]lbValue, len(results))
			values := make([]lbValue, len(results))
			for i := 0; i < len(p.Type.Proc.Results.Tuple.Variables); i++ {
				e := p.Type.Proc.Results.Tuple.Variables[i]
				if e.Kind != Entity_Variable {
					continue
				}
				if e.Token.String.Len == 0 {
					continue
				}
				p.Module.ValuesMutex.RLock()
				named_results[i] = p.Module.Values[unsafe.Pointer(e)]
				p.Module.ValuesMutex.RUnlock()
				values[i] = lb_emit_conv(p, results[i], e.Type)
			}
			for i := 0; i < len(named_results); i++ {
				lb_emit_store(p, named_results[i], values[i])
			}
		}
		split_returns := ft.MultipleReturnOriginalType != nil
		if split_returns {
			result_values := make([]lbValue, len(results))
			result_eps := make([]lbValue, len(results)-1)
			for i := 0; i < len(results); i++ {
				result_values[i] = lb_emit_conv(p, results[i], tuple.Variables[i].Type)
			}
			param_offset := 0
			if return_by_pointer {
				param_offset = 1
			}
			param_offset += int(ft.OriginalArgCount)
			for i := 0; i < len(result_eps); i++ {
				var result_ep lbValue
				result_ep.Value = LLVMGetParam(p.Value, uint32(param_offset+i))
				result_ep.Type = alloc_type_pointer(tuple.Variables[i].Type)
				result_eps[i] = result_ep
			}
			for i := 0; i < len(result_eps); i++ {
				lb_emit_store(p, result_eps[i], result_values[i])
			}
			if return_by_pointer {
				lb_addr_store(p, p.ReturnPtr, result_values[len(result_values)-1])
				lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
				LLVMBuildRetVoid(p.Builder)
				return
			} else {
				lb_build_return_stmt_internal(p, result_values[len(result_values)-1], pos)
				return
			}
		} else {
			ret_type := p.Type.Proc.Results
			result_values := make([]lbValue, len(results))
			for i := 0; i < len(results); i++ {
				result_values[i] = lb_emit_conv(p, results[i], tuple.Variables[i].Type)
			}
			if return_by_pointer {
				res_val := p.ReturnPtr.Addr
				result_eps := make([]lbValue, len(results))
				for i := 0; i < len(results); i++ {
					result_eps[i] = lb_emit_struct_ep(p, res_val, int32(i))
				}
				for i := 0; i < len(result_eps); i++ {
					lb_emit_store(p, result_eps[i], result_values[i])
				}
				lb_emit_defer_stmts(p, lbDeferExit_Return, nil, pos)
				LLVMBuildRetVoid(p.Builder)
				return
			}
			res = lb_build_struct_value(p, ret_type, result_values, len(result_values))
		}
	}
	lb_build_return_stmt_internal(p, res, pos)
}

// =============================================================================
// lb_build_assign_stmt_array - builds compound assignment for array types
// =============================================================================
func lb_build_assign_stmt_array(p *lbProcedure, op TokenKind, lhs lbAddr, value lbValue) {
	lhs_type := lb_addr_type(lhs)
	array_type := base_type(lhs_type)
	count := get_array_type_count(array_type)
	elem_type := base_array_type(array_type)
	rhs := lb_emit_conv(p, value, lhs_type)
	inline_array_arith := lb_can_try_to_inline_array_arith(array_type)

	if lhs.Kind == lbAddr_Swizzle {
		type ValueAndIndex struct {
			value lbValue
			index uint8
		}
		var indices_handled [4]bool
		var indices [4]int32
		index_count := 0
		for i := uint8(0); i < lhs.SwizzleCount; i++ {
			idx := lhs.SwizzleIndices[i]
			if indices_handled[idx] {
				continue
			}
			indices[index_count] = int32(idx)
			indices_handled[idx] = true
			index_count++
		}
		var lhs_ptrs [4]lbValue
		var x_loads [4]lbValue
		var y_loads [4]lbValue
		var ops [4]lbValue
		for i := 0; i < index_count; i++ {
			lhs_ptrs[i] = lb_emit_array_epi(p, lhs.Addr, indices[i])
		}
		for i := 0; i < index_count; i++ {
			x_loads[i] = lb_emit_load(p, lhs_ptrs[i])
		}
		for i := 0; i < index_count; i++ {
			y_loads[i].Value = LLVMBuildExtractValue(p.Builder, rhs.Value, uint32(i), "")
			y_loads[i].Type = elem_type
		}
		for i := 0; i < index_count; i++ {
			ops[i] = lb_emit_arith(p, op, x_loads[i], y_loads[i], elem_type)
		}
		for i := 0; i < index_count; i++ {
			lb_emit_store(p, lhs_ptrs[i], ops[i])
		}
		return
	} else if lhs.Kind == lbAddr_SwizzleLarge {
		type ValueAndIndex struct {
			value lbValue
			index uint32
		}
		bt := base_type(lhs_type)
		indices_handled := make([]bool, bt.Array.Count)
		indices := make([]int32, bt.Array.Count)
		index_count := 0
		for _, idx := range lhs.SwizzleLargeIndices {
			if indices_handled[idx] {
				continue
			}
			indices[index_count] = idx
			indices_handled[idx] = true
			index_count++
		}
		var lhs_ptrs [4]lbValue
		var x_loads [4]lbValue
		var y_loads [4]lbValue
		var ops [4]lbValue
		for i := 0; i < index_count; i++ {
			lhs_ptrs[i] = lb_emit_array_epi(p, lhs.Addr, indices[i])
		}
		for i := 0; i < index_count; i++ {
			x_loads[i] = lb_emit_load(p, lhs_ptrs[i])
		}
		for i := 0; i < index_count; i++ {
			y_loads[i].Value = LLVMBuildExtractValue(p.Builder, rhs.Value, uint32(i), "")
			y_loads[i].Type = elem_type
		}
		for i := 0; i < index_count; i++ {
			ops[i] = lb_emit_arith(p, op, x_loads[i], y_loads[i], elem_type)
		}
		for i := 0; i < index_count; i++ {
			lb_emit_store(p, lhs_ptrs[i], ops[i])
		}
		return
	}

	x := lb_addr_get_ptr(p, lhs)
	if inline_array_arith {
		n := uint32(count)
		lhs_ptrs := make([]lbValue, n)
		x_loads := make([]lbValue, n)
		y_loads := make([]lbValue, n)
		ops := make([]lbValue, n)
		for i := uint32(0); i < n; i++ {
			lhs_ptrs[i] = lb_emit_array_epi(p, x, int32(i))
		}
		for i := uint32(0); i < n; i++ {
			x_loads[i] = lb_emit_load(p, lhs_ptrs[i])
		}
		for i := uint32(0); i < n; i++ {
			y_loads[i].Value = LLVMBuildExtractValue(p.Builder, rhs.Value, i, "")
			y_loads[i].Type = elem_type
		}
		for i := uint32(0); i < n; i++ {
			ops[i] = lb_emit_arith(p, op, x_loads[i], y_loads[i], elem_type)
		}
		for i := uint32(0); i < n; i++ {
			lb_emit_store(p, lhs_ptrs[i], ops[i])
		}
	} else {
		y := lb_address_from_load_or_generate_local(p, rhs)
		loop_data := lb_loop_start(p, isize(count), t_i32)
		a_ptr := lb_emit_array_ep(p, x, loop_data.Idx)
		b_ptr := lb_emit_array_ep(p, y, loop_data.Idx)
		a := lb_emit_load(p, a_ptr)
		b := lb_emit_load(p, b_ptr)
		c := lb_emit_arith(p, op, a, b, elem_type)
		lb_emit_store(p, a_ptr, c)
		lb_loop_end(p, loop_data)
	}
}

// =============================================================================
// lb_build_assign_stmt - main assignment statement builder
// =============================================================================
func lb_build_assign_stmt(p *lbProcedure, as *AstAssignStmt) {
	if as.Op.Kind == Token_Eq {
		if buildContext.EnableRVO {
			if len(as.LHS) == 1 && len(as.RHS) == 1 && !is_blank_ident(as.LHS[0]) {
				rhs_expr := unparen_expr(as.RHS[0])
				if rhs_expr.Kind == Ast_CallExpr {
					lval := lb_build_addr(p, as.LHS[0])
					if LLVMIsAAllocaInst(lval.Addr.Value) != 0 && lval.Kind == lbAddr_Default {
						if lb_call_sret_eligible(p, rhs_expr, lb_addr_type(lval)) {
							dst := lval.Addr
							lb_build_call_expr(p, rhs_expr, &dst)
							return
						}
					}
				}
			}
		}
		lvals := make([]lbAddr, 0, len(as.LHS))
		for _, lhs := range as.LHS {
			var lval lbAddr
			if !is_blank_ident(lhs) {
				lval = lb_build_addr(p, lhs)
			}
			lvals = append(lvals, lval)
		}
		lb_build_assignment(p, lvals, as.RHS)
		return
	}
	if len(as.LHS) != 1 || len(as.RHS) != 1 {
		panic("as->lhs.count == 1 && as->rhs.count == 1")
	}
	op_ := int32(as.Op.Kind)
	op_ += Token_Add - Token_AddEq
	op := TokenKind(op_)
	if op == Token_CmpAnd || op == Token_CmpOr {
		typ := as.LHS[0].TAV.Type
		new_value := lb_emit_logical_binary_expr(p, op, as.LHS[0], as.RHS[0], typ)
		lhs := lb_build_addr(p, as.LHS[0])
		lb_addr_store(p, lhs, new_value)
	} else {
		lhs := lb_build_addr(p, as.LHS[0])
		val := lb_build_expr(p, as.RHS[0])
		lhs_type := lb_addr_type(lhs)
		if op == Token_Mul && is_type_matrix(val.Type) && is_type_array(lhs_type) {
			old_value := lb_addr_load(p, lhs)
			typ := old_value.Type
			new_value := lb_emit_vector_mul_matrix(p, old_value, val, typ)
			lb_addr_store(p, lhs, new_value)
			return
		}
		if is_type_array(lhs_type) {
			lb_build_assign_stmt_array(p, op, lhs, val)
			return
		} else {
			old_value := lb_addr_load(p, lhs)
			typ := old_value.Type
			change := lb_emit_conv(p, val, typ)
			new_value := lb_emit_arith(p, op, old_value, change, typ)
			lb_addr_store(p, lhs, new_value)
		}
	}
}
