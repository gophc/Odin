package cmd

func lb_get_soa_variable_addr(p *lbProcedure, e *Entity) lbAddr {
	return p.Module.SoaValues[e]
}

func lb_get_using_variable(p *lbProcedure, e *Entity) lbValue {
	interned := entity_interned_name(e)
	parent := e.UsingParent
	sel := lookup_field(parent.Type, interned, false)
	pv := p.Module.Values[parent]

	v := lbValue{}

	is_soa := false
	if pv.Value == 0 && parent.Flags&EntityFlag_SoaPtrField != 0 {
		is_soa = true
		parent_addr := lb_get_soa_variable_addr(p, parent)
		v = lb_addr_get_ptr(p, parent_addr)
	} else if pv.Value != 0 {
		v = pv
	} else {
		if e.UsingExpr == nil {
			gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4676,
				"%.*s", len(e.Token.String), e.Token.String)
		}
		v = lb_build_addr_ptr(p, e.UsingExpr)
	}

	ptr := lb_emit_deep_field_gep(p, v, sel)
	if parent.Scope != nil {
		if parent.Scope.Flags&(ScopeFlag_File|ScopeFlag_Pkg) == 0 {
			lb_add_debug_local_variable(p, ptr.Value, e.Type, e.Token)
		}
	} else {
		lb_add_debug_local_variable(p, ptr.Value, e.Type, e.Token)
	}
	return ptr
}

func lb_build_addr_from_entity(p *lbProcedure, e *Entity, expr *Ast) lbAddr {
	if e.Kind == Entity_Constant {
		t := default_type(type_of_expr(expr))
		v := lb_const_value(p.Module, t, e.Constant.Value, LB_CONST_CONTEXT_DEFAULT_NO_LOCAL, e.Type)
		if LLVMIsConstant(v.Value) {
			g := lb_add_global_generated_from_procedure(p, t, v)
			return g
		}
		ptr := lbValue{}
		ptr.Value = LLVMGetOperand(v.Value, 0)
		ptr.Type = alloc_type_pointer(t)
		return lb_addr(ptr)
	}

	v := lbValue{}
	found, ok := p.Module.Values[e]
	if ok {
		v = found
	} else if e.Kind == Entity_Variable && e.Flags&EntityFlag_Using != 0 {
		v = lb_get_using_variable(p, e)
	} else if e.Flags&EntityFlag_SoaPtrField != 0 {
		return lb_get_soa_variable_addr(p, e)
	}

	if v.Value == 0 {
		return lb_addr(lb_find_value_from_entity(p.Module, e))
	}

	return lb_addr(v)
}

func lb_build_array_swizzle_addr(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbAddr {
	index_count := len(ce.Args) - 1
	addr := lb_build_addr(p, ce.Args[0])
	if index_count == 0 {
		return addr
	}
	type_ := base_type(lb_addr_type(addr))
	count := type_.Array.Count
	if count <= 4 && index_count <= 4 {
		var indices [4]u8
		var icount u8 = 0
		for i := 1; i < len(ce.Args); i++ {
			tv := type_and_value_of_expr(ce.Args[i])
			src_index := big_int_to_i64(&tv.Value.ValueInteger)
			indices[icount] = u8(src_index)
			icount++
		}
		return lb_addr_swizzle(lb_addr_get_ptr(p, addr), tv.Type, int(icount), indices[:])
	}
	indices := make([]i32, len(ce.Args)-1)
	index_index := 0
	for i := 1; i < len(ce.Args); i++ {
		tv := type_and_value_of_expr(ce.Args[i])
		src_index := big_int_to_i64(&tv.Value.ValueInteger)
		indices[index_index] = i32(src_index)
		index_index++
	}
	return lb_addr_swizzle_large(lb_addr_get_ptr(p, addr), tv.Type, indices)
}

func lb_build_addr(p *lbProcedure, expr *Ast) lbAddr {
	expr = unparen_expr(expr)

	if expr.StateFlags&uint16(StateFlag_SelectorCallExpr) != 0 {
		pp, ok := p.SelectorAddr[expr]
		if ok {
			res := pp
			delete(p.SelectorAddr, expr)
			return res
		}
	}
	addr := lb_build_addr_internal(p, expr)
	if expr.StateFlags&uint16(StateFlag_SelectorCallExpr) != 0 {
		p.SelectorAddr[expr] = addr
	}
	return addr
}

func lb_build_addr_index_expr(p *lbProcedure, expr *Ast) lbAddr {
	ie := &expr.IndexExpr

	t := base_type(type_of_expr(ie.Expr))

	deref := is_type_pointer(t)
	t = base_type(type_deref(t))
	if is_type_soa_struct(t) {
		val := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			val = lb_emit_load(p, val)
		}
		index := lb_build_expr(p, ie.Index)
		return lb_addr_soa_variable(val, index, ie.Index)
	}

	if ie.Expr.Tav.Mode == Addressing_SoaVariable {
		field := lb_build_expr(p, ie.Expr)
		index := lb_build_expr(p, ie.Index)

		if !build_context.NoBoundsCheck {
			se_expr := unparen_expr(ie.Expr)
			if se_expr.Kind == Ast_SelectorExpr {
				se := &se_expr.SelectorExpr
				var len_ lbValue

				type_ := base_type(type_deref(type_of_expr(se.Expr)))
				if type_.Struct.SoaKind == StructSoa_Fixed {
					len_ = lb_const_int(p.Module, t_int, type_.Struct.SoaCount)
				} else {
					found, ok := p.SelectorAddr[se_expr]
					if ok {
						addr := found
						parent := lb_addr_get_ptr(p, addr)
						if is_type_pointer(type_deref(parent.Type)) {
							parent = lb_emit_load(p, parent)
						}
						len_ = lb_soa_struct_len(p, parent)
					}
				}

				if len_.Value != 0 {
					lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
				}
			}
		}
		val := lb_emit_ptr_offset(p, field, index)
		val.Type = alloc_type_multi_pointer_to_pointer(val.Type)
		return lb_addr(val)
	}

	if !is_type_indexable(t) {
		gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4999,
			"%s %s", type_to_string(t), expr_to_string(expr))
	}

	if is_type_map(t) {
		map_addr := lb_build_addr(p, ie.Expr)
		key := lb_build_expr(p, ie.Index)
		key = lb_emit_conv(p, key, t.Map.Key)

		result_type := type_of_expr(expr)
		map_ptr := lb_addr_get_ptr(p, map_addr)
		if is_type_pointer(type_deref(map_ptr.Type)) {
			map_ptr = lb_emit_load(p, map_ptr)
		}
		return lb_addr_map(map_ptr, key, t, result_type)
	}

	switch t.Kind {
	case TypeArray:
		array := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			array = lb_emit_load(p, array)
		}
		index := lb_build_expr(p, ie.Index)
		index = lb_emit_conv(p, index, t_int)
		elem := lb_emit_array_ep(p, array, index)

		index_tv := type_and_value_of_expr(ie.Index)
		if index_tv.Mode != Addressing_Constant {
			len_ := lb_const_int(p.Module, t_int, t.Array.Count)
			lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		}
		return lb_addr(elem)

	case Type_FixedCapacityDynamicArray:
		array := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			array = lb_emit_load(p, array)
		}
		index := lb_build_expr(p, ie.Index)
		index = lb_emit_conv(p, index, t_int)

		array_ptr := lb_emit_struct_ep(p, array, 0)
		elem := lb_emit_array_ep(p, array_ptr, index)

		index_tv := type_and_value_of_expr(ie.Index)
		if index_tv.Mode != Addressing_Constant {
			len_ := lb_fixed_capacity_dynamic_array_len(p, array)
			lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		}
		return lb_addr(elem)

	case Type_EnumeratedArray:
		array := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			array = lb_emit_load(p, array)
		}

		index_type := t.EnumeratedArray.Index

		index_tv := type_and_value_of_expr(ie.Index)

		var index lbValue
		if compare_exact_values(Token_NotEq, *t.EnumeratedArray.MinValue, exact_value_i64(0)) {
			if index_tv.Mode == Addressing_Constant {
				idx := exact_value_sub(index_tv.Value, *t.EnumeratedArray.MinValue)
				index = lb_const_value(p.Module, index_type, idx)
			} else {
				index = lb_emit_arith(p, Token_Sub,
					lb_build_expr(p, ie.Index),
					lb_const_value(p.Module, index_type, *t.EnumeratedArray.MinValue),
					index_type)
				index = lb_emit_conv(p, index, t_int)
			}
		} else {
			index = lb_emit_conv(p, lb_build_expr(p, ie.Index), t_int)
		}

		elem := lb_emit_array_ep(p, array, index)

		if index_tv.Mode != Addressing_Constant {
			len_ := lb_const_int(p.Module, t_int, t.EnumeratedArray.Count)
			lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		}
		return lb_addr(elem)

	case Type_Slice:
		slice := lb_build_expr(p, ie.Expr)
		if deref {
			slice = lb_emit_load(p, slice)
		}
		elem := lb_slice_elem(p, slice)
		index := lb_emit_conv(p, lb_build_expr(p, ie.Index), t_int)
		len_ := lb_slice_len(p, slice)
		lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		v := lb_emit_ptr_offset(p, elem, index)
		return lb_addr(v)

	case Type_MultiPointer:
		multi_ptr := lb_build_expr(p, ie.Expr)
		if deref {
			multi_ptr = lb_emit_load(p, multi_ptr)
		}
		index := lb_build_expr(p, ie.Index)
		index = lb_emit_conv(p, index, t_int)
		v := lbValue{}
		indices := [1]LLVMValueRef{index.Value}
		v.Value = LLVMBuildGEP2(p.Builder, lb_type(p.Module, t.MultiPointer.Elem), multi_ptr.Value, indices[:], 1, "")
		v.Type = alloc_type_pointer(t.MultiPointer.Elem)
		return lb_addr(v)

	case Type_DynamicArray:
		dynamic_array := lb_build_expr(p, ie.Expr)
		if deref {
			dynamic_array = lb_emit_load(p, dynamic_array)
		}
		elem := lb_dynamic_array_elem(p, dynamic_array)
		len_ := lb_dynamic_array_len(p, dynamic_array)
		index := lb_emit_conv(p, lb_build_expr(p, ie.Index), t_int)
		lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		v := lb_emit_ptr_offset(p, elem, index)
		return lb_addr(v)

	case Type_Matrix:
		matrix := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			matrix = lb_emit_load(p, matrix)
		}
		index := lb_build_expr(p, ie.Index)
		index = lb_emit_conv(p, index, t_int)

		var bounds_len isize
		var elem lbValue
		if t.Matrix.IsRowMajor {
			bounds_len = t.Matrix.RowCount
			elem = lb_emit_matrix_ep(p, matrix, index, lb_const_int(p.Module, t_int, 0))
		} else {
			bounds_len = t.Matrix.ColumnCount
			elem = lb_emit_matrix_ep(p, matrix, lb_const_int(p.Module, t_int, 0), index)
		}
		elem = lb_emit_conv(p, elem, alloc_type_pointer(type_of_expr(expr)))

		index_tv := type_and_value_of_expr(ie.Index)
		if index_tv.Mode != Addressing_Constant {
			len_ := lb_const_int(p.Module, t_int, bounds_len)
			lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		}
		return lb_addr(elem)

	case Type_Basic:
		str := lb_build_expr(p, ie.Expr)
		if deref {
			str = lb_emit_load(p, str)
		}
		elem := lb_string_elem(p, str)
		len_ := lb_string_len(p, str)
		index := lb_emit_conv(p, lb_build_expr(p, ie.Index), t_int)
		lb_emit_bounds_check(p, ast_token(ie.Index), index, len_)
		return lb_addr(lb_emit_ptr_offset(p, elem, index))
	}
	return lbAddr{}
}

func lb_build_slice_expr_value(p *lbProcedure, expr *Ast) lbValue {
	se := &expr.SliceExpr

	{
		inner_ast := unparen_expr(se.Expr)
		if inner_ast.Kind == Ast_SliceExpr {
			const MAX_CHAIN = 8

			var chain [MAX_CHAIN]*Ast
			var chain_count isize = 0
			chain[chain_count] = expr
			chain_count++

			root := inner_ast
			for root.Kind == Ast_SliceExpr && chain_count < MAX_CHAIN {
				chain[chain_count] = root
				chain_count++
				root = unparen_expr(root.SliceExpr.Expr)
			}

			root_type := base_type(type_of_expr(root))
			if is_type_pointer(root_type) {
				root_type = base_type(type_deref(root_type))
			}
			if is_type_slice(root_type) ||
				is_type_dynamic_array(root_type) ||
				is_type_string(root_type) ||
				is_type_string16(root_type) {
				m := p.Module

				src := lb_build_expr(p, root)
				src_type := base_type(src.Type)
				if is_type_pointer(src_type) {
					src_type = base_type(type_deref(src_type))
					src = lb_emit_load(p, src)
				}

				var cur_ptr lbValue
				var cur_len lbValue
				if is_type_slice(src_type) {
					cur_ptr = lb_slice_elem(p, src)
					cur_len = lb_slice_len(p, src)
				} else if is_type_dynamic_array(src_type) {
					cur_ptr = lb_dynamic_array_elem(p, src)
					cur_len = lb_dynamic_array_len(p, src)
				} else {
					cur_ptr = lb_string_elem(p, src)
					cur_len = lb_string_len(p, src)
				}

				for i := chain_count - 1; i >= 0; i-- {
					s := &chain[i].SliceExpr

					lo := lb_const_int(m, t_int, 0)
					hi := lbValue{}
					if s.Low != nil {
						lo = lb_correct_endianness(p, lb_build_expr(p, s.Low))
					}
					if s.High != nil {
						hi = lb_correct_endianness(p, lb_build_expr(p, s.High))
					}
					if hi.Value == 0 {
						hi = cur_len
					}

					no_indices := s.Low == nil && s.High == nil
					if !no_indices {
						lb_emit_slice_bounds_check(p, s.Open, lo, hi, cur_len, s.Low != nil)
					}

					if s.Low == nil {
						cur_ptr = cur_ptr
					} else {
						cur_ptr = lb_emit_ptr_offset(p, cur_ptr, lo)
					}
					if s.Low == nil {
						cur_len = lb_emit_conv(p, hi, t_int)
					} else {
						cur_len = lb_emit_arith(p, Token_Sub, hi, lo, t_int)
					}
				}

				result_type := type_of_expr(expr)
				if is_type_string(result_type) {
					return lb_make_string_value(p, result_type, cur_ptr, cur_len)
				}
				return lb_make_slice_value(p, result_type, cur_ptr, cur_len)
			}
		}
	}

	base := lb_build_expr(p, se.Expr)
	type_ := base_type(base.Type)

	if is_type_pointer(type_) {
		type_ = base_type(type_deref(type_))
		base = lb_emit_load(p, base)
	}

	low := lb_const_int(p.Module, t_int, 0)
	high := lbValue{}

	if se.Low != nil {
		low = lb_correct_endianness(p, lb_build_expr(p, se.Low))
	}
	if se.High != nil {
		high = lb_correct_endianness(p, lb_build_expr(p, se.High))
	}

	no_indices := se.Low == nil && se.High == nil

	switch type_.Kind {
	case Type_Slice:
		slice_type := type_
		len_ := lb_slice_len(p, base)
		if high.Value == 0 {
			high = len_
		}
		if !no_indices {
			lb_emit_slice_bounds_check(p, se.Open, low, high, len_, se.Low != nil)
		}
		elem := lb_emit_ptr_offset(p, lb_slice_elem(p, base), low)
		var new_len lbValue
		if se.Low == nil {
			new_len = lb_emit_conv(p, high, t_int)
		} else {
			new_len = lb_emit_arith(p, Token_Sub, high, low, t_int)
		}
		return lb_make_slice_value(p, slice_type, elem, new_len)

	case Type_DynamicArray:
		elem_type := type_.DynamicArray.Elem
		slice_type := alloc_type_slice(elem_type)

		len_ := lb_dynamic_array_len(p, base)
		if high.Value == 0 {
			high = len_
		}
		if !no_indices {
			lb_emit_slice_bounds_check(p, se.Open, low, high, len_, se.Low != nil)
		}
		elem := lb_emit_ptr_offset(p, lb_dynamic_array_elem(p, base), low)
		var new_len lbValue
		if se.Low == nil {
			new_len = lb_emit_conv(p, high, t_int)
		} else {
			new_len = lb_emit_arith(p, Token_Sub, high, low, t_int)
		}
		return lb_make_slice_value(p, slice_type, elem, new_len)

	case Type_Basic:
		if is_type_string16(type_) {
			len_ := lb_string_len(p, base)
			if high.Value == 0 {
				high = len_
			}
			if !no_indices {
				lb_emit_slice_bounds_check(p, se.Open, low, high, len_, se.Low != nil)
			}
			elem := lb_emit_ptr_offset(p, lb_string_elem(p, base), low)
			var new_len lbValue
			if se.Low == nil {
				new_len = lb_emit_conv(p, high, t_int)
			} else {
				new_len = lb_emit_arith(p, Token_Sub, high, low, t_int)
			}
			return lb_make_string_value(p, t_string16, elem, new_len)
		}

		len_ := lb_string_len(p, base)
		if high.Value == 0 {
			high = len_
		}
		if !no_indices {
			lb_emit_slice_bounds_check(p, se.Open, low, high, len_, se.Low != nil)
		}
		elem := lb_emit_ptr_offset(p, lb_string_elem(p, base), low)
		var new_len lbValue
		if se.Low == nil {
			new_len = lb_emit_conv(p, high, t_int)
		} else {
			new_len = lb_emit_arith(p, Token_Sub, high, low, t_int)
		}
		return lb_make_string_value(p, t_string, elem, new_len)

	case Type_MultiPointer:
		if se.High == nil {
			gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 5368, "multi-pointer slice requires high bound")
		}
		elem_type := type_.MultiPointer.Elem
		slice_type := alloc_type_slice(elem_type)

		low = lb_emit_conv(p, low, t_int)
		high = lb_emit_conv(p, high, t_int)

		lb_emit_multi_pointer_slice_bounds_check(p, se.Open, low, high)

		indices := [1]LLVMValueRef{low.Value}
		ptr := lbValue{}
		ptr.Value = LLVMBuildGEP2(p.Builder, lb_type(p.Module, elem_type), base.Value, indices[:], 1, "")
		ptr.Type = alloc_type_pointer(elem_type)

		var new_len lbValue
		if se.Low == nil {
			new_len = lb_emit_conv(p, high, t_int)
		} else {
			new_len = lb_emit_arith(p, Token_Sub, high, low, t_int)
		}
		return lb_make_slice_value(p, slice_type, ptr, new_len)

	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 5390,
			"lb_build_slice_expr_value: unexpected type %s", type_to_string(base.Type))
		return lbValue{}
	}
}

func lb_build_addr_internal(p *lbProcedure, expr *Ast) lbAddr {
	switch expr.Kind {
	case Ast_Implicit:
		i := &expr.Implicit
		v := lbAddr{}
		switch i.Kind {
		case Token_context:
			v = lb_find_or_generate_context_ptr(p)
		}
		return v

	case Ast_Ident:
		if is_blank_ident(expr) {
			return lbAddr{}
		}
		e := entity_of_node(expr)
		return lb_build_addr_from_entity(p, e, expr)

	case Ast_SelectorExpr:
		se := &expr.SelectorExpr
		sel_node := unparen_expr(se.Selector)
		if sel_node.Kind == Ast_Ident {
			selector := sel_node.Ident.Interned
			tav := type_and_value_of_expr(se.Expr)

			if tav.Mode == Addressing_Invalid {
				imp := entity_of_node(se.Expr)
				if imp != nil {
					_ = imp
				}
				return lb_build_addr(p, unparen_expr(se.Selector))
			}

			if tav.Mode == Addressing_Type {
				sel := lookup_field(tav.Type, selector, true)
				if sel.PseudoField {
					e := entity_of_node(sel_node)
					return lb_addr(lb_find_value_from_entity(p.Module, e))
				}
				gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6350, "Unreachable %s", selector.String())
			}

			if se.SwizzleCount > 0 {
				array_type := base_type(type_deref(tav.Type))
				swizzle_count := se.SwizzleCount
				swizzle_indices_raw := se.SwizzleIndices
				var swizzle_indices [4]u8
				for i := u8(0); i < swizzle_count; i++ {
					index := (swizzle_indices_raw >> (i * 2)) & 3
					swizzle_indices[i] = index
				}
				a := lbValue{}
				if is_type_pointer(tav.Type) {
					a = lb_build_expr(p, se.Expr)
				} else {
					addr := lb_build_addr(p, se.Expr)
					a = lb_addr_get_ptr(p, addr)
				}

				type_ := type_deref(expr.Tav.Type)
				return lb_addr_swizzle(a, type_, int(swizzle_count), swizzle_indices[:])
			}

			sel := lookup_field(tav.Type, selector, false)
			if sel.PseudoField && (sel.Entity.Kind == Entity_Procedure || sel.Entity.Kind == Entity_ProcGroup) {
				e := entity_of_node(sel_node)
				return lb_addr(lb_find_value_from_entity(p.Module, e))
			}

			addr := lb_build_addr(p, se.Expr)

			deref_type := type_deref(tav.Type)
			if tav.Type.Kind == Type_Pointer && deref_type.Kind == Type_Named && deref_type.Named.TypeName.TypeName.ObjcIvar {
				addr = lb_addr(lb_emit_load(p, addr.Addr))
				ivar_ptr := lb_handle_objc_ivar_for_objc_object_pointer(p, addr.Addr)
				addr = lb_addr(ivar_ptr)
			}

			if sel.IsBitField {
				sub_sel := sel
				sub_sel.Index.Count -= 1

				ptr := lb_addr_get_ptr(p, addr)
				if sub_sel.Index.Count > 0 {
					ptr = lb_emit_deep_field_gep(p, ptr, sub_sel)
				}
				if is_type_pointer(type_deref(ptr.Type)) {
					ptr = lb_emit_load(p, ptr)
				}

				bf_type := type_deref(ptr.Type)
				bf_type = base_type(bf_type)

				index := sel.Index.Data[sel.Index.Count-1]

				f := bf_type.BitField.Fields[index]
				bit_size := bf_type.BitField.BitSizes[index]
				bit_offset := bf_type.BitField.BitOffsets[index]

				return lb_addr_bit_field(ptr, f.Type, bit_offset, i64(bit_size))
			}

			{
				if addr.Kind == lbAddr_Map {
					v := lb_addr_load(p, addr)
					a := lb_address_from_load_or_generate_local(p, v)
					a = lb_emit_deep_field_gep(p, a, sel)
					return lb_addr(a)
				} else if addr.Kind == lbAddr_Context {
					if addr.CtxSel.Index.Count >= 0 {
						sel = selection_combine(addr.CtxSel, sel)
					}
					addr.CtxSel = sel
					addr.Kind = lbAddr_Context
					return addr
				} else if addr.Kind == lbAddr_SoaVariable {
					index := addr.SoaIndex
					first_index := sel.Index.Data[0]
					sub_sel := sel
					sub_sel.Index.Data = sub_sel.Index.Data[1:]
					sub_sel.Index.Count -= 1

					arr := lb_emit_struct_ep(p, addr.Addr, first_index)

					t := base_type(type_deref(addr.Addr.Type))
					if addr.SoaIndexExpr != nil && (!lb_is_const(addr.SoaIndex) || t.Struct.SoaKind != StructSoa_Fixed) {
						len_ := lb_soa_struct_len(p, addr.Addr)
						lb_emit_bounds_check(p, ast_token(addr.SoaIndexExpr), addr.SoaIndex, len_)
					}

					var item lbValue
					if t.Struct.SoaKind == StructSoa_Fixed {
						item = lb_emit_array_ep(p, arr, index)
					} else {
						item = lb_emit_ptr_offset(p, lb_emit_load(p, arr), index)
					}
					if sub_sel.Index.Count > 0 {
						item = lb_emit_deep_field_gep(p, item, sub_sel)
					}
					item.Type = alloc_type_multi_pointer_to_pointer(item.Type)

					return lb_addr(item)
				} else if addr.Kind == lbAddr_Swizzle {
					sel.Index.Data[0] = i32(addr.SwizzleIndices[sel.Index.Data[0]])
				} else if addr.Kind == lbAddr_SwizzleLarge {
					sel.Index.Data[0] = addr.SwizzleLargeIndices[sel.Index.Data[0]]
				}

				atype := type_deref(lb_addr_type(addr))
				if is_type_soa_struct(atype) {
					p.SelectorAddr[expr] = addr
				}

				a := lb_addr_get_ptr(p, addr)
				a = lb_emit_deep_field_gep(p, a, sel)
				return lb_addr(a)
			}
		} else {
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6488, "Unsupported selector expression")
		}

	case Ast_SelectorCallExpr:
		e := lb_build_expr(p, expr)
		return lb_addr(lb_address_from_load_or_generate_local(p, e))

	case Ast_TypeAssertion:
		ta := &expr.TypeAssertion
		pos := ast_token(expr).Pos
		e := lb_build_expr(p, ta.Expr)
		t := type_deref(e.Type)
		if is_type_union(t) {
			type_ := type_of_expr(expr)
			v := lb_add_local_generated(p, type_, false)
			lb_addr_store(p, v, lb_emit_union_cast(p, e, type_, pos))
			return v
		} else if is_type_any(t) {
			type_ := type_of_expr(expr)
			return lb_emit_any_cast_addr(p, e, type_, pos)
		} else {
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6510,
				"TODO(bill): type assertion %s", type_to_string(e.Type))
		}

	case Ast_UnaryExpr:
		ue := &expr.UnaryExpr
		switch ue.Op.Kind {
		case Token_And:
			ptr := lb_build_expr(p, expr)
			return lb_addr(lb_address_from_load_or_generate_local(p, ptr))
		default:
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6521, "Invalid unary expression for lb_build_addr")
		}

	case Ast_BinaryExpr:
		v := lb_build_expr(p, expr)
		t := v.Type
		if is_type_pointer(t) {
			return lb_addr(v)
		}
		return lb_addr(lb_address_from_load_or_generate_local(p, v))

	case Ast_IndexExpr:
		return lb_build_addr_index_expr(p, expr)

	case Ast_MatrixIndexExpr:
		ie := &expr.MatrixIndexExpr
		t := base_type(type_of_expr(ie.Expr))

		deref := is_type_pointer(t)
		t = base_type(type_deref(t))

		m := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			m = lb_emit_load(p, m)
		}
		row_index := lb_build_expr(p, ie.RowIndex)
		column_index := lb_build_expr(p, ie.ColumnIndex)
		row_index = lb_emit_conv(p, row_index, t_int)
		column_index = lb_emit_conv(p, column_index, t_int)
		elem := lb_emit_matrix_ep(p, m, row_index, column_index)

		row_index_tv := type_and_value_of_expr(ie.RowIndex)
		column_index_tv := type_and_value_of_expr(ie.ColumnIndex)
		if row_index_tv.Mode != Addressing_Constant || column_index_tv.Mode != Addressing_Constant {
			row_count := lb_const_int(p.Module, t_int, t.Matrix.RowCount)
			column_count := lb_const_int(p.Module, t_int, t.Matrix.ColumnCount)
			lb_emit_matrix_bounds_check(p, ast_token(ie.RowIndex), row_index, column_index, row_count, column_count)
		}
		return lb_addr(elem)

	case Ast_SliceExpr:
		return lb_build_addr_slice_expr(p, expr)

	case Ast_DerefExpr:
		de := &expr.DerefExpr
		t := type_of_expr(de.Expr)
		if is_type_soa_pointer(t) {
			value := lb_build_expr(p, de.Expr)
			ptr := lb_emit_struct_ev(p, value, 0)
			idx := lb_emit_struct_ev(p, value, 1)
			return lb_addr_soa_variable(ptr, idx, nil)
		}
		addr := lb_build_expr(p, de.Expr)
		return lb_addr(addr)

	case Ast_CallExpr:
		ce := &expr.CallExpr
		builtin_id := BuiltinProc_Invalid
		if ce.Proc.Tav.Mode == Addressing_Builtin {
			e := entity_of_node(ce.Proc)
			if e != nil {
				builtin_id = BuiltinProcId(e.Builtin.Id)
			} else {
				builtin_id = BuiltinProc_DIRECTIVE
			}
		}
		tv := expr.Tav
		if builtin_id == BuiltinProc_swizzle && is_type_array(tv.Type) {
			return lb_build_array_swizzle_addr(p, ce, tv)
		}

		e := lb_build_expr(p, expr)
		return lb_addr(lb_address_from_load_or_generate_local(p, e))

	case Ast_CompoundLit:
		return lb_build_addr_compound_lit(p, expr)

	case Ast_TypeCast:
		tc := &expr.TypeCast
		type_ := type_of_expr(expr)
		x := lb_build_expr(p, tc.Expr)
		var e lbValue
		switch tc.Token.Kind {
		case Token_cast:
			e = lb_emit_conv(p, x, type_)
		case Token_transmute:
			e = lb_emit_transmute(p, x, type_)
		default:
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6627, "Invalid AST TypeCast")
		}
		v := lb_add_local_generated(p, type_, false)
		lb_addr_store(p, v, e)
		return v

	case Ast_AutoCast:
		ac := &expr.AutoCast
		return lb_build_addr(p, ac.Expr)

	case Ast_TernaryIfExpr:
		te := &expr.TernaryIfExpr
		var incoming_values [2]LLVMValueRef
		var incoming_blocks [2]LLVMBasicBlockRef

		then_ := lb_create_block(p, "if.then")
		done := lb_create_block(p, "if.done")
		else_ := lb_create_block(p, "if.else")

		lb_build_cond(p, te.Cond, then_, else_)
		lb_start_block(p, then_)

		ptr_type := alloc_type_pointer(default_type(type_of_expr(expr)))

		incoming_values[0] = lb_emit_conv(p, lb_build_addr_ptr(p, te.X), ptr_type).Value

		lb_emit_jump(p, done)
		lb_start_block(p, else_)

		incoming_values[1] = lb_emit_conv(p, lb_build_addr_ptr(p, te.Y), ptr_type).Value

		lb_emit_jump(p, done)
		lb_start_block(p, done)

		res := lbValue{}
		res.Value = LLVMBuildPhi(p.Builder, lb_type(p.Module, ptr_type), "")
		res.Type = ptr_type

		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block

		LLVMAddIncoming(res.Value, incoming_values[:], incoming_blocks[:], 2)

		return lb_addr(res)

	case Ast_OrElseExpr:
		ptr := lb_address_from_load_or_generate_local(p, lb_build_expr(p, expr))
		return lb_addr(ptr)

	case Ast_OrReturnExpr:
		ptr := lb_address_from_load_or_generate_local(p, lb_build_expr(p, expr))
		return lb_addr(ptr)

	case Ast_OrBranchExpr:
		be := &expr.OrBranchExpr
		var block *lbBlock

		if be.Label != nil {
			bb := lb_lookup_branch_blocks(p, be.Label)
			switch be.Token.Kind {
			case Token_or_break:
				block = bb.Break_
			case Token_or_continue:
				block = bb.Continue_
			}
		} else {
			for t := p.TargetList; t != nil && block == nil; t = t.Prev {
				if t.IsBlock {
					continue
				}
				switch be.Token.Kind {
				case Token_or_break:
					block = t.Break_
				case Token_or_continue:
					block = t.Continue_
				}
			}
		}

		if block == nil {
			gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 6708, "or_branch block not found")
		}

		tv := expr.Tav

		var lhs lbValue
		var rhs lbValue
		lb_emit_try_lhs_rhs(p, be.Expr, tv, &lhs, &rhs)
		type_ := default_type(tv.Type)
		if lhs.Value != 0 {
			lhs = lb_emit_conv(p, lhs, type_)
		} else if type_ != nil && type_ != t_invalid {
			lhs = lb_const_nil(p.Module, type_)
		}

		then_ := lb_create_block(p, "or_branch.then")
		else_ := lb_create_block(p, "or_branch.else")

		lb_emit_if(p, lb_emit_try_has_value(p, rhs), then_, else_)
		lb_start_block(p, else_)
		lb_emit_defer_stmts(p, lbDeferExit_Branch, block, expr)
		lb_emit_jump(p, block)
		lb_start_block(p, then_)

		return lb_addr(lb_address_from_load_or_generate_local(p, lhs))
	}

	token_pos := ast_token(expr).Pos
	gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 6735,
		"Unexpected address expression\n\tAst: %.*s @ %s\n",
		len(ast_strings[expr.Kind]), ast_strings[expr.Kind],
		token_pos_to_string(token_pos))

	return lbAddr{}
}
