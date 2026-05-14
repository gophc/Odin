package cmd

func lb_is_type_aggregate(t *Type) bool {
	t = base_type(t)
	switch t.Kind {
	case Type_Basic:
		switch t.Basic.Kind {
		case Basic_string, Basic_string16, Basic_any:
			return true
		case Basic_complex32, Basic_complex64, Basic_complex128:
			return true
		case Basic_quaternion64, Basic_quaternion128, Basic_quaternion256:
			return true
		}
	case Type_Pointer:
		return false
	case Type_Array, Type_Slice, Type_Struct, Type_Union:
		return true
	case Type_Tuple, Type_DynamicArray, Type_Map, Type_SimdVector:
		return true
	case Type_Named:
		return lb_is_type_aggregate(t.Named.Base)
	}
	return false
}

func lb_emit_unreachable(p *lbProcedure) {
	instr := LLVMGetLastInstruction(p.CurrBlock.Block)
	if instr == 0 || !lb_is_instr_terminating(instr) {
		lb_call_intrinsic(p, "llvm.trap", nil, 0, nil, 0)
		LLVMBuildUnreachable(p.Builder)
	}
}

func lb_correct_endianness(p *lbProcedure, value lbValue) lbValue {
	src := core_type(value.Type)
	if !(is_type_integer(src) || is_type_float(src)) {
		gb_assert_handler("Assertion Failure", "is_type_integer(src) || is_type_float(src)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 53, 0)
	}
	if is_type_different_to_arch_endianness(src) {
		platform_src_type := integer_endian_type_to_platform_type(src)
		value = lb_emit_byte_swap(p, value, platform_src_type)
	}
	return value
}

func lb_set_metadata_custom_u64(m *lbModule, v_ref LLVMValueRef, name string, value u64) {
	md_id := LLVMGetMDKindIDInContext(m.Ctx, name, uint(len(name)))
	md := LLVMValueAsMetadata(LLVMConstInt(lb_type(m, t_u64), value, false))
	node := LLVMMetadataAsValue(m.Ctx, LLVMMDNodeInContext2(m.Ctx, &md, 1))
	LLVMSetMetadata(v_ref, md_id, node)
}

func lb_get_metadata_custom_u64(m *lbModule, v_ref LLVMValueRef, name string) u64 {
	md_id := LLVMGetMDKindIDInContext(m.Ctx, name, uint(len(name)))
	v_md := LLVMGetMetadata(v_ref, md_id)
	if v_md == 0 {
		return 0
	}
	node_count := LLVMGetMDNodeNumOperands(v_md)
	if node_count == 0 {
		return 0
	}
	if node_count != 1 {
		gb_assert_handler("Assertion Failure", "node_count == 1", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 78, 0)
	}
	var value LLVMValueRef
	LLVMGetMDNodeOperands(v_md, &value)
	return LLVMConstIntGetZExtValue(value)
}

func lb_mem_zero_ptr_internal(p *lbProcedure, ptr LLVMValueRef, len_ any, alignment uint, is_volatile bool) LLVMValueRef {
	switch v := len_.(type) {
	case uint, u64, i64, isize:
		var val u64
		switch vv := v.(type) {
		case uint:
			val = u64(vv)
		case u64:
			val = vv
		case i64:
			val = u64(vv)
		case isize:
			val = u64(vv)
		}
		return lb_mem_zero_ptr_internal(p, ptr, LLVMConstInt(lb_type(p.Module, t_u64), val, false), alignment, is_volatile)
	}
	len_val := len_.(LLVMValueRef)
	is_inlinable := false
	const_len := i64(0)
	if !p.IsStartup && LLVMIsConstant(len_val) != 0 {
		const_len = LLVMConstIntGetSExtValue(len_val)
		if const_len <= lb_max_zero_init_size() {
			is_inlinable = true
		}
	}
	name := "llvm.memset"
	if is_inlinable {
		name = "llvm.memset.inline"
	}
	types := []LLVMTypeRef{
		lb_type(p.Module, t_rawptr),
		lb_type(p.Module, t_int),
	}
	args := []LLVMValueRef{
		LLVMBuildPointerCast(p.Builder, ptr, types[0], ""),
		LLVMConstInt(LLVMInt8TypeInContext(p.Module.Ctx), 0, false),
		LLVMBuildIntCast2(p.Builder, len_val, types[1], false, ""),
		LLVMConstInt(LLVMInt1TypeInContext(p.Module.Ctx), boolToLLVM(is_volatile), false),
	}
	return lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
}

func lb_mem_zero_ptr(p *lbProcedure, ptr LLVMValueRef, typ *Type, alignment uint) {
	llvm_type := lb_type(p.Module, typ)
	kind := LLVMGetTypeKind(llvm_type)
	sz := type_size_of(typ)
	switch kind {
	case LLVMStructTypeKind, LLVMArrayTypeKind:
		if is_type_tuple(typ) {
			if typ.Kind != Type_Tuple {
				gb_assert_handler("Assertion Failure", "type->kind == Type_Tuple", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 130, 0)
			}
			n := len(typ.Tuple.Variables) - 1
			end_offset := typ.Tuple.Offsets[n] + type_size_of(typ.Tuple.Variables[n].Type)
			lb_mem_zero_ptr_internal(p, ptr, lb_const_int(p.Module, t_int, u64(end_offset)).Value, alignment, false)
		} else {
			lb_mem_zero_ptr_internal(p, ptr, lb_const_int(p.Module, t_int, u64(sz)).Value, alignment, false)
		}
	default:
		LLVMBuildStore(p.Builder, LLVMConstNull(lb_type(p.Module, typ)), ptr)
	}
}

func lb_emit_select(p *lbProcedure, cond lbValue, x lbValue, y lbValue) lbValue {
	cond = lb_emit_conv(p, cond, t_llvm_bool)
	res := lbValue{}
	res.Value = LLVMBuildSelect(p.Builder, cond.Value, x.Value, y.Value, "")
	res.Type = x.Type
	return res
}

func lb_emit_min(p *lbProcedure, t *Type, x lbValue, y lbValue) lbValue {
	x = lb_emit_conv(p, x, t)
	y = lb_emit_conv(p, y, t)
	use_llvm_intrinsic := !is_arch_wasm() && (is_type_float(t) || (is_type_simd_vector(t) && is_type_float(base_array_type(t))))
	if use_llvm_intrinsic {
		args := []LLVMValueRef{x.Value, y.Value}
		types := []LLVMTypeRef{lb_type(p.Module, t)}
		v := lb_call_intrinsic(p, "llvm.minnum", args, uint(len(args)), types, uint(len(types)))
		return lbValue{Value: v, Type: t}
	}
	return lb_emit_select(p, lb_emit_comp(p, Token_Lt, x, y), x, y)
}

func lb_emit_max(p *lbProcedure, t *Type, x lbValue, y lbValue) lbValue {
	x = lb_emit_conv(p, x, t)
	y = lb_emit_conv(p, y, t)
	use_llvm_intrinsic := !is_arch_wasm() && (is_type_float(t) || (is_type_simd_vector(t) && is_type_float(base_array_type(t))))
	if use_llvm_intrinsic {
		args := []LLVMValueRef{x.Value, y.Value}
		types := []LLVMTypeRef{lb_type(p.Module, t)}
		v := lb_call_intrinsic(p, "llvm.maxnum", args, uint(len(args)), types, uint(len(types)))
		return lbValue{Value: v, Type: t}
	}
	return lb_emit_select(p, lb_emit_comp(p, Token_Gt, x, y), x, y)
}

func lb_emit_clamp(p *lbProcedure, t *Type, x lbValue, min lbValue, max lbValue) lbValue {
	z := lb_emit_max(p, t, x, min)
	z = lb_emit_min(p, t, z, max)
	return z
}

func lb_emit_string16(p *lbProcedure, str_elem lbValue, str_len lbValue) lbValue {
	if false && lb_is_const(str_elem) && lb_is_const(str_len) {
		values := []LLVMValueRef{str_elem.Value, str_len.Value}
		res := lbValue{}
		res.Type = t_string16
		res.Value = llvm_const_named_struct(p.Module, t_string16, values, isize(len(values)))
		return res
	} else {
		res := lb_add_local_generated(p, t_string16, false)
		lb_emit_store(p, lb_emit_struct_ep(p, res.Addr, 0), str_elem)
		lb_emit_store(p, lb_emit_struct_ep(p, res.Addr, 1), str_len)
		return lb_addr_load(p, res)
	}
}

func lb_emit_string(p *lbProcedure, str_elem lbValue, str_len lbValue) lbValue {
	if false && lb_is_const(str_elem) && lb_is_const(str_len) {
		values := []LLVMValueRef{str_elem.Value, str_len.Value}
		res := lbValue{}
		res.Type = t_string
		res.Value = llvm_const_named_struct(p.Module, t_string, values, isize(len(values)))
		return res
	} else {
		res := lb_add_local_generated(p, t_string, false)
		lb_emit_store(p, lb_emit_struct_ep(p, res.Addr, 0), str_elem)
		lb_emit_store(p, lb_emit_struct_ep(p, res.Addr, 1), str_len)
		return lb_addr_load(p, res)
	}
}

func lb_emit_transmute(p *lbProcedure, value lbValue, t *Type) lbValue {
	src_type := value.Type
	if are_types_identical(t, src_type) {
		return value
	}
	res := lbValue{}
	res.Type = t
	src := base_type(src_type)
	dst := base_type(t)
	m := p.Module
	sz := type_size_of(src)
	dz := type_size_of(dst)
	if sz != dz {
		s := lb_type(m, src)
		d := lb_type(m, dst)
		llvm_sz := lb_sizeof(s)
		llvm_dz := lb_sizeof(d)
		if llvm_sz != llvm_dz {
			gb_assert_handler("Assertion Failure", "llvm_sz == llvm_dz", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 255, LLVMPrintTypeToString(s), LLVMPrintTypeToString(d))
		}
	}
	if sz != dz {
		gb_assert_handler("Assertion Failure", "sz == dz", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 258, type_to_string(src_type), type_to_string(t))
	}
	if is_type_internally_pointer_like(src) {
		if is_type_integer(dst) {
			res.Value = LLVMBuildPtrToInt(p.Builder, value.Value, lb_type(m, t), "")
			return res
		} else if is_type_internally_pointer_like(dst) {
			res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(p.Module, t), "")
			return res
		} else if is_type_float(dst) {
			the_int := LLVMBuildPtrToInt(p.Builder, value.Value, lb_type(m, t_uintptr), "")
			res.Value = LLVMBuildBitCast(p.Builder, the_int, lb_type(m, t), "")
			return res
		}
	}
	if is_type_internally_pointer_like(dst) {
		if is_type_uintptr(src) && is_type_internally_pointer_like(dst) {
			res.Value = LLVMBuildIntToPtr(p.Builder, value.Value, lb_type(m, t), "")
			return res
		} else if is_type_integer(src) && is_type_internally_pointer_like(dst) {
			res.Value = LLVMBuildIntToPtr(p.Builder, value.Value, lb_type(m, t), "")
			return res
		} else if is_type_float(src) {
			the_int := LLVMBuildBitCast(p.Builder, value.Value, lb_type(m, t_uintptr), "")
			res.Value = LLVMBuildIntToPtr(p.Builder, the_int, lb_type(m, t), "")
			return res
		}
	}
	is_simd_vector_bitcastable := false
	if is_type_simd_vector(src) && is_type_simd_vector(dst) {
		if !is_type_internally_pointer_like(src.SimdVector.Elem) && !is_type_internally_pointer_like(dst.SimdVector.Elem) {
			is_simd_vector_bitcastable = true
		}
	}
	if is_simd_vector_bitcastable {
		res.Value = LLVMBuildBitCast(p.Builder, value.Value, lb_type(p.Module, t), "")
		return res
	} else if is_type_array_like(src) && (is_type_simd_vector(dst) || is_type_integer_128bit(dst)) {
		align := type_align_of(src)
		if d_align := type_align_of(dst); d_align > align {
			align = d_align
		}
		ptr := lb_address_from_load_or_generate_local(p, value)
		if lb_try_update_alignment(ptr, u64(align)) {
			result_type := lb_type(p.Module, t)
			res.Value = LLVMBuildPointerCast(p.Builder, ptr.Value, LLVMPointerType(result_type, 0), "")
			res.Value = OdinLLVMBuildLoad(p, result_type, res.Value)
			return res
		}
		addr := lb_add_local_generated(p, t, false)
		ap := lb_addr_get_ptr(p, addr)
		ap = lb_emit_conv(p, ap, alloc_type_pointer(value.Type))
		lb_emit_store(p, ap, value)
		return lb_addr_load(p, addr)
	} else if is_type_map(src) && are_types_identical(t_raw_map, t) {
		res.Value = value.Value
		res.Type = t
		return res
	} else if lb_is_type_aggregate(src) || lb_is_type_aggregate(dst) {
		s := lb_address_from_load_or_generate_local(p, value)
		d := lb_emit_transmute(p, s, alloc_type_pointer(t))
		return lb_emit_load(p, d)
	}
	res.Value = OdinLLVMBuildTransmute(p, value.Value, lb_type(m, res.Type))
	return res
}

func lb_copy_value_to_ptr(p *lbProcedure, val lbValue, new_type *Type, alignment i64) lbValue {
	type_alignment := type_align_of(new_type)
	if alignment < type_alignment {
		alignment = type_alignment
	}
	if !are_types_identical(new_type, val.Type) {
		gb_assert_handler("Assertion Failure", "are_types_identical(new_type, val.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 332, type_to_string(new_type), type_to_string(val.Type))
	}
	ptr := lb_add_local_generated(p, new_type, false)
	LLVMSetAlignment(ptr.Addr.Value, uint(alignment))
	lb_addr_store(p, ptr, val)
	return ptr.Addr
}

func lb_emit_try_lhs_rhs(p *lbProcedure, arg *Ast, tv TypeAndValue, lhs_ *lbValue, rhs_ *lbValue) {
	lhs := lbValue{}
	rhs := lbValue{}
	value := lb_build_expr(p, arg)
	if is_type_tuple(value.Type) {
		n := len(value.Type.Tuple.Variables) - 1
		if len(value.Type.Tuple.Variables) == 2 {
			lhs = lb_emit_tuple_ev(p, value, 0)
		} else {
			lhs_addr := lb_add_local_generated(p, tv.Type, false)
			lhs_ptr := lb_addr_get_ptr(p, lhs_addr)
			for i := i32(0); i < i32(n); i++ {
				lb_emit_store(p, lb_emit_struct_ep(p, lhs_ptr, i), lb_emit_tuple_ev(p, value, i))
			}
			lhs = lb_addr_load(p, lhs_addr)
		}
		rhs = lb_emit_tuple_ev(p, value, i32(n))
	} else {
		rhs = value
	}
	if rhs.Value == 0 {
		gb_assert_handler("Assertion Failure", "rhs.value != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 422, 0)
	}
	if lhs_ != nil {
		*lhs_ = lhs
	}
	if rhs_ != nil {
		*rhs_ = rhs
	}
}

func lb_emit_try_has_value(p *lbProcedure, rhs lbValue) lbValue {
	has_value := lbValue{}
	if is_type_boolean(rhs.Type) {
		has_value = rhs
	} else {
		if !type_has_nil(rhs.Type) {
			gb_assert_handler("Assertion Failure", "type_has_nil(rhs.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 434, type_to_string(rhs.Type))
		}
		has_value = lb_emit_comp_against_nil(p, Token_CmpEq, rhs)
	}
	if has_value.Value == 0 {
		gb_assert_handler("Assertion Failure", "has_value.value != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 437, 0)
	}
	return has_value
}

func lb_emit_or_else(p *lbProcedure, arg *Ast, else_expr *Ast, tv TypeAndValue) lbValue {
	if arg.StateFlags&uint8(StateFlag_DirectiveWasFalse) != 0 {
		return lb_build_expr(p, else_expr)
	}
	lhs := lbValue{}
	rhs := lbValue{}
	lb_emit_try_lhs_rhs(p, arg, tv, &lhs, &rhs)
	if else_expr == nil {
		gb_assert_handler("Assertion Failure", "else_expr != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 495, 0)
	}
	typ := default_type(tv.Type)
	if is_diverging_expr(else_expr) {
		then_ := lb_create_block(p, "or_else.then")
		else_ := lb_create_block(p, "or_else.else")
		lb_emit_if(p, lb_emit_try_has_value(p, rhs), then_, else_)
		lb_start_block(p, else_)
		lb_build_expr(p, else_expr)
		lb_emit_unreachable(p)
		lb_start_block(p, then_)
		return lb_emit_conv(p, lhs, typ)
	} else {
		if lb_is_type_trivial(typ) && lb_is_expr_trivial(else_expr) {
			has_value := lb_emit_try_has_value(p, rhs)
			then_val := lb_emit_conv(p, lhs, typ)
			else_val := lb_emit_conv(p, lb_build_expr(p, else_expr), typ)
			return lb_emit_select(p, has_value, then_val, else_val)
		}
		incoming_values := make([]LLVMValueRef, 2)
		incoming_blocks := make([]LLVMBasicBlockRef, 2)
		then_ := lb_create_block(p, "or_else.then")
		done := lb_create_block(p, "or_else.done")
		else_ := lb_create_block(p, "or_else.else")
		lb_emit_if(p, lb_emit_try_has_value(p, rhs), then_, else_)
		lb_start_block(p, then_)
		incoming_values[0] = lb_emit_conv(p, lhs, typ).Value
		lb_emit_jump(p, done)
		lb_start_block(p, else_)
		incoming_values[1] = lb_emit_conv(p, lb_build_expr(p, else_expr), typ).Value
		lb_emit_jump(p, done)
		lb_start_block(p, done)
		res := lbValue{}
		res.Value = LLVMBuildPhi(p.Builder, lb_type(p.Module, typ), "")
		res.Type = typ
		if len(p.CurrBlock.Preds) < 2 {
			gb_assert_handler("Assertion Failure", "p->curr_block->preds.count >= 2", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 545, 0)
		}
		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block
		LLVMAddIncoming(res.Value, incoming_values, incoming_blocks, 2)
		return res
	}
}

func lb_build_return_stmt(p *lbProcedure, return_results []*Ast, pos TokenPos) {}
func lb_build_return_stmt_internal(p *lbProcedure, res lbValue, pos TokenPos)  {}

func lb_emit_or_return(p *lbProcedure, arg *Ast, tv TypeAndValue) lbValue {
	lhs := lbValue{}
	rhs := lbValue{}
	lb_emit_try_lhs_rhs(p, arg, tv, &lhs, &rhs)
	return_block := lb_create_block(p, "or_return.return")
	continue_block := lb_create_block(p, "or_return.continue")
	lb_emit_if(p, lb_emit_try_has_value(p, rhs), continue_block, return_block)
	lb_start_block(p, return_block)
	{
		proc_type := base_type(p.Type)
		results := proc_type.Proc.Results
		if results == nil || results.Kind != Type_Tuple {
			gb_assert_handler("Assertion Failure", "results != nullptr && results->kind == Type_Tuple", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 571, 0)
		}
		tuple := &results.Tuple
		if len(tuple.Variables) == 0 {
			gb_assert_handler("Assertion Failure", "tuple->variables.count != 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 574, 0)
		}
		end_entity := tuple.Variables[len(tuple.Variables)-1]
		rhs = lb_emit_conv(p, rhs, end_entity.Type)
		if p.Type.Proc.HasNamedResults {
			if end_entity.Token.String.Len == 0 {
				gb_assert_handler("Assertion Failure", "end_entity->token.string.len != 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 579, 0)
			}
			found := map_must_get(&p.Module.Values, end_entity)
			lb_emit_store(p, found, rhs)
			lb_build_return_stmt(p, nil, ast_token(arg).Pos)
		} else {
			if len(tuple.Variables) != 1 {
				gb_assert_handler("Assertion Failure", "tuple->variables.count == 1", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 587, 0)
			}
			lb_build_return_stmt_internal(p, rhs, ast_token(arg).Pos)
		}
	}
	lb_start_block(p, continue_block)
	if tv.Type != nil {
		return lb_emit_conv(p, lhs, tv.Type)
	}
	return lbValue{}
}

func lb_emit_increment(p *lbProcedure, addr lbValue) {
	if !is_type_pointer(addr.Type) {
		gb_assert_handler("Assertion Failure", "is_type_pointer(addr.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 600, 0)
	}
	typ := type_deref(addr.Type, false)
	v_one := lb_const_value(p.Module, typ, exact_value_i64(1))
	lb_emit_store(p, addr, lb_emit_arith(p, Token_Add, lb_emit_load(p, addr), v_one, typ))
}

func lb_emit_byte_swap(p *lbProcedure, value lbValue, end_type *Type) lbValue {
	if type_size_of(value.Type) != type_size_of(end_type) {
		gb_assert_handler("Assertion Failure", "type_size_of(value.type) == type_size_of(end_type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 608, 0)
	}
	if type_size_of(value.Type) < 2 {
		return value
	}
	original_type := value.Type
	if is_type_float(original_type) {
		sz := type_size_of(original_type)
		var integer_type *Type
		switch sz {
		case 2:
			integer_type = t_u16
		case 4:
			integer_type = t_u32
		case 8:
			integer_type = t_u64
		}
		if integer_type == nil {
			gb_assert_handler("Assertion Failure", "integer_type != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 623, 0)
		}
		value = lb_emit_transmute(p, value, integer_type)
	}
	name := "llvm.bswap"
	types := []LLVMTypeRef{lb_type(p.Module, value.Type)}
	args := []LLVMValueRef{value.Value}
	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	res.Type = value.Type
	if is_type_float(original_type) {
		res = lb_emit_transmute(p, res, original_type)
	}
	res.Type = end_type
	return res
}

func lb_emit_count_ones(p *lbProcedure, x lbValue, typ *Type) lbValue {
	x = lb_emit_conv(p, x, typ)
	name := "llvm.ctpop"
	types := []LLVMTypeRef{lb_type(p.Module, typ)}
	args := []LLVMValueRef{x.Value}
	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	res.Type = typ
	return res
}

func lb_emit_count_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue {
	elem := base_array_type(typ)
	sz := 8 * type_size_of(elem)
	size := lb_const_int(p.Module, elem, u64(sz))
	size = lb_emit_conv(p, size, typ)
	count := lb_emit_count_ones(p, x, typ)
	return lb_emit_arith(p, Token_Sub, size, count, typ)
}

func lb_emit_count_trailing_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue {
	x = lb_emit_conv(p, x, typ)
	name := "llvm.cttz"
	types := []LLVMTypeRef{lb_type(p.Module, typ)}
	args := []LLVMValueRef{x.Value, LLVMConstNull(LLVMInt1TypeInContext(p.Module.Ctx))}
	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	res.Type = typ
	return res
}

func lb_emit_count_leading_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue {
	x = lb_emit_conv(p, x, typ)
	name := "llvm.ctlz"
	types := []LLVMTypeRef{lb_type(p.Module, typ)}
	args := []LLVMValueRef{x.Value, LLVMConstNull(LLVMInt1TypeInContext(p.Module.Ctx))}
	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	res.Type = typ
	return res
}

func lb_emit_count_trailing_ones(p *lbProcedure, x lbValue, typ *Type) lbValue {
	z := lb_emit_unary_arith(p, Token_Xor, x, typ)
	return lb_emit_count_trailing_zeros(p, z, typ)
}

func lb_emit_count_leading_ones(p *lbProcedure, x lbValue, typ *Type) lbValue {
	z := lb_emit_unary_arith(p, Token_Xor, x, typ)
	return lb_emit_count_leading_zeros(p, z, typ)
}

func lb_emit_reverse_bits(p *lbProcedure, x lbValue, typ *Type) lbValue {
	x = lb_emit_conv(p, x, typ)
	name := "llvm.bitreverse"
	types := []LLVMTypeRef{lb_type(p.Module, typ)}
	args := []LLVMValueRef{x.Value}
	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	res.Type = typ
	return res
}

func lb_emit_union_cast_only_ok_check(p *lbProcedure, value lbValue, typ *Type, pos TokenPos) lbValue {
	if !is_type_tuple(typ) {
		gb_assert_handler("Assertion Failure", "is_type_tuple(type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 732, 0)
	}
	m := p.Module
	src_type := value.Type
	is_ptr := is_type_pointer(src_type)
	ok_type := typ.Tuple.Variables[1].Type
	gen_tuple_types := []*Type{ok_type, ok_type}
	gen_tuple := alloc_type_tuple_from_field_types(gen_tuple_types, len(gen_tuple_types), false, true)
	v := lb_add_local_generated(p, gen_tuple, false)
	if is_ptr {
		value = lb_emit_load(p, value)
	}
	src := base_type(type_deref(src_type, false))
	if !is_type_union(src) {
		gb_assert_handler("Assertion Failure", "is_type_union(src)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 758, type_to_string(src_type))
	}
	dst := typ.Tuple.Variables[0].Type
	var cond lbValue
	if is_type_union_maybe_pointer(src) {
		data := lb_emit_transmute(p, value, dst)
		cond = lb_emit_comp_against_nil(p, Token_NotEq, data)
	} else {
		tag := lb_emit_union_tag_value(p, value)
		dst_tag := lb_const_union_tag(m, src, dst)
		cond = lb_emit_comp(p, Token_CmpEq, tag, dst_tag)
	}
	gep1 := lb_emit_struct_ep(p, v.Addr, 1)
	lb_emit_store(p, gep1, cond)
	return lb_addr_load(p, v)
}

func lb_emit_union_cast(p *lbProcedure, value lbValue, typ *Type, pos TokenPos) lbValue {
	m := p.Module
	src_type := value.Type
	is_ptr := is_type_pointer(src_type)
	is_tuple := true
	tuple := typ
	if typ.Kind != Type_Tuple {
		is_tuple = false
		tuple = make_optional_ok_type(typ)
	}
	v := lb_add_local_generated(p, tuple, true)
	if is_ptr {
		value = lb_emit_load(p, value)
	}
	src := base_type(type_deref(src_type, false))
	if !is_type_union(src) {
		gb_assert_handler("Assertion Failure", "is_type_union(src)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 797, type_to_string(src_type))
	}
	dst := tuple.Tuple.Variables[0].Type
	value_ := lb_address_from_load_or_generate_local(p, value)
	if (p.StateFlags&uint16(StateFlag_no_type_assert)) != 0 && !is_tuple {
		ptr := lb_emit_conv(p, value_, alloc_type_pointer(typ))
		return lb_emit_load(p, ptr)
	}
	var tag, dst_tag, cond, data lbValue
	var gep0, gep1 lbValue
	gep0 = lb_emit_struct_ep(p, v.Addr, 0)
	gep1 = lb_emit_struct_ep(p, v.Addr, 1)
	if is_type_union_maybe_pointer(src) {
		data = lb_emit_load(p, lb_emit_conv(p, value_, gep0.Type))
	} else {
		tag = lb_emit_load(p, lb_emit_union_tag_ptr(p, value_))
		dst_tag = lb_const_union_tag(m, src, dst)
	}
	ok_block := lb_create_block(p, "union_cast.ok")
	end_block := lb_create_block(p, "union_cast.end")
	if data.Value != 0 {
		if !is_type_union_maybe_pointer(src) {
			gb_assert_handler("Assertion Failure", "is_type_union_maybe_pointer(src)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 827, 0)
		}
		cond = lb_emit_comp_against_nil(p, Token_NotEq, data)
	} else {
		cond = lb_emit_comp(p, Token_CmpEq, tag, dst_tag)
	}
	lb_emit_if(p, cond, ok_block, end_block)
	lb_start_block(p, ok_block)
	if data.Value == 0 {
		data = lb_emit_load(p, lb_emit_conv(p, value_, gep0.Type))
	}
	lb_emit_store(p, gep0, data)
	lb_emit_store(p, gep1, lb_const_bool(m, t_bool, true))
	lb_emit_jump(p, end_block)
	lb_start_block(p, end_block)
	if !is_tuple {
		if !buildContext.NoTypeAssert {
			if (p.StateFlags & uint16(StateFlag_no_type_assert)) != 0 {
				gb_assert_handler("Assertion Failure", "(p->state_flags & StateFlag_no_type_assert) == 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 849, 0)
			}
			dst_type := tuple.Tuple.Variables[0].Type
			arg_count := isize(7)
			if buildContext.NoRTTI {
				arg_count = 4
			}
			ok := lb_emit_load(p, lb_emit_struct_ep(p, v.Addr, 1))
			args := make([]lbValue, arg_count)
			args[0] = ok
			lb_set_file_line_col(p, args[1:], pos)
			if !buildContext.NoRTTI {
				args[4] = lb_typeid(m, src_type)
				args[5] = lb_typeid(m, dst_type)
				args[6] = lb_emit_conv(p, value_, t_rawptr)
			}
			name := "type_assertion_check2_contextless"
			if len(p.ContextStack) > 0 {
				name = "type_assertion_check2_with_context"
			}
			lb_emit_runtime_call(p, name, args)
		}
		return lb_emit_load(p, lb_emit_struct_ep(p, v.Addr, 0))
	}
	return lb_addr_load(p, v)
}

func lb_emit_any_cast_addr(p *lbProcedure, value lbValue, typ *Type, pos TokenPos) lbAddr {
	m := p.Module
	src_type := value.Type
	if is_type_pointer(src_type) {
		value = lb_emit_load(p, value)
	}
	is_tuple := true
	tuple := typ
	if typ.Kind != Type_Tuple {
		is_tuple = false
		tuple = make_optional_ok_type(typ)
	}
	dst_type := tuple.Tuple.Variables[0].Type
	if (p.StateFlags&uint16(StateFlag_no_type_assert)) != 0 && !is_tuple {
		ptr := lb_emit_struct_ev(p, value, 0)
		ptr = lb_emit_conv(p, ptr, alloc_type_pointer(typ))
		return lb_addr(ptr)
	}
	v := lb_add_local_generated(p, tuple, true)
	dst_typeid := lb_typeid(m, dst_type)
	any_typeid := lb_emit_struct_ev(p, value, 1)
	ok_block := lb_create_block(p, "any_cast.ok")
	end_block := lb_create_block(p, "any_cast.end")
	cond := lb_emit_comp(p, Token_CmpEq, any_typeid, dst_typeid)
	lb_emit_if(p, cond, ok_block, end_block)
	lb_start_block(p, ok_block)
	gep0 := lb_emit_struct_ep(p, v.Addr, 0)
	gep1 := lb_emit_struct_ep(p, v.Addr, 1)
	any_data := lb_emit_struct_ev(p, value, 0)
	ptr := lb_emit_conv(p, any_data, alloc_type_pointer(dst_type))
	lb_emit_store(p, gep0, lb_emit_load(p, ptr))
	lb_emit_store(p, gep1, lb_const_bool(m, t_bool, true))
	lb_emit_jump(p, end_block)
	lb_start_block(p, end_block)
	if !is_tuple {
		if !buildContext.NoTypeAssert {
			ok := lb_emit_load(p, lb_emit_struct_ep(p, v.Addr, 1))
			arg_count := isize(7)
			if buildContext.NoRTTI {
				arg_count = 4
			}
			args := make([]lbValue, arg_count)
			args[0] = ok
			lb_set_file_line_col(p, args[1:], pos)
			if !buildContext.NoRTTI {
				args[4] = any_typeid
				args[5] = dst_typeid
				args[6] = lb_emit_struct_ev(p, value, 0)
			}
			name := "type_assertion_check2_contextless"
			if len(p.ContextStack) > 0 {
				name = "type_assertion_check2_with_context"
			}
			lb_emit_runtime_call(p, name, args)
		}
		return lb_addr(lb_emit_struct_ep(p, v.Addr, 0))
	}
	return v
}

func lb_emit_any_cast(p *lbProcedure, value lbValue, typ *Type, pos TokenPos) lbValue {
	return lb_addr_load(p, lb_emit_any_cast_addr(p, value, typ, pos))
}

func lb_find_or_generate_context_ptr(p *lbProcedure) lbAddr {
	if len(p.ContextStack) > 0 {
		return p.ContextStack[len(p.ContextStack)-1].Ctx
	}
	pt := base_type(p.Type)
	if pt.Kind != Type_Proc {
		gb_assert_handler("Assertion Failure", "pt->kind == Type_Proc", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 971, 0)
	}
	if pt.Proc.CallingConvention == ProcCC_Odin {
		gb_assert_handler("Assertion Failure", "pt->Proc.calling_convention != ProcCC_Odin", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 972, 0)
	}
	c := lb_add_local_generated(p, t_context, true)
	c.Kind = lbAddr_Context
	lb_emit_init_context(p, c)
	lb_push_context_onto_stack(p, c)
	lb_add_debug_context_variable(p, c)
	return c
}

func lb_address_from_load_or_generate_local(p *lbProcedure, value lbValue) lbValue {
	if LLVMIsALoadInst(value.Value) != 0 {
		res := lbValue{}
		res.Value = LLVMGetOperand(value.Value, 0)
		res.Type = alloc_type_pointer(value.Type)
		return res
	}
	if !is_type_typed(value.Type) {
		gb_assert_handler("Assertion Failure", "is_type_typed(value.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 991, 0)
	}
	res := lb_add_local_generated(p, value.Type, false)
	lb_addr_store(p, res, value)
	return res.Addr
}

func lb_address_from_load(p *lbProcedure, value lbValue) lbValue {
	if LLVMIsALoadInst(value.Value) != 0 {
		res := lbValue{}
		res.Value = LLVMGetOperand(value.Value, 0)
		res.Type = alloc_type_pointer(value.Type)
		return res
	}
	gb_assert_handler("Panic", "0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 1005, "lb_address_from_load")
	return lbValue{}
}

func lb_address_from_load_if_readonly_parameter(p *lbProcedure, x lbValue) lbValue {
	if LLVMIsALoadInst(x.Value) == 0 {
		return lbValue{}
	}
	optr := LLVMGetOperand(x.Value, 0)
	for optr != 0 && LLVMIsABitCastInst(optr) != 0 {
		optr = LLVMGetOperand(optr, 0)
	}
	param_index := LLVMAttributeIndex(1)
	if p.ReturnPtr.Addr.Value != 0 {
		param_index++
	}
	is_parameter := false
	for _, param := range p.RawInputParameters {
		if param == optr {
			is_parameter = true
			break
		}
		param_index++
	}
	if is_parameter {
		readonly_attr_kind := LLVMGetEnumAttributeKindForName("readonly", 8)
		n := LLVMGetAttributeCountAtIndex(p.Value, param_index)
		if n != 0 {
			attrs := make([]LLVMAttributeRef, n)
			LLVMGetAttributesAtIndex(p.Value, param_index, &attrs[0])
			for i := uint(0); i < n; i++ {
				if LLVMGetEnumAttributeKind(attrs[i]) == readonly_attr_kind {
					return lb_address_from_load_or_generate_local(p, x)
				}
			}
		}
	}
	return lbValue{}
}

func llvm_atomic_ordering_from_odin(value ExactValue) LLVMAtomicOrdering {
	if value.Kind != ExactValue_Integer {
		gb_assert_handler("Assertion Failure", "value.kind == ExactValue_Integer", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 3091, 0)
	}
	v := exact_value_to_i64(value)
	switch v {
	case OdinAtomicMemoryOrderRelaxed:
		return LLVMAtomicOrderingMonotonic
	case OdinAtomicMemoryOrderConsume:
		return LLVMAtomicOrderingAcquire
	case OdinAtomicMemoryOrderAcquire:
		return LLVMAtomicOrderingAcquire
	case OdinAtomicMemoryOrderRelease:
		return LLVMAtomicOrderingRelease
	case OdinAtomicMemoryOrderAcqRel:
		return LLVMAtomicOrderingAcquireRelease
	case OdinAtomicMemoryOrderSeqCst:
		return LLVMAtomicOrderingSequentiallyConsistent
	}
	gb_assert_handler("Panic", "0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_utility.cpp", 3101, "Unknown atomic ordering")
	return LLVMAtomicOrderingSequentiallyConsistent
}

func llvm_atomic_ordering_from_odin_expr(expr *Ast) LLVMAtomicOrdering {
	value := expr.TAV.Value
	return llvm_atomic_ordering_from_odin(value)
}
