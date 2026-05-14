package cmd

func lb_const_low_bits_mask(type_ LLVMTypeRef, bit_count u64) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "bit_count <= 64", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 5, 0)
	if bit_count == 0 {
		return LLVMConstInt(type_, 0, 0)
	}
	mask := ^u64(0)
	if bit_count != 64 {
		mask = (u64(1) << bit_count) - 1
	}
	return LLVMConstInt(type_, mask, 0)
}

func lb_emit_logical_binary_expr(p *lbProcedure, op TokenKind, left *Ast, right *Ast, final_type *Type) lbValue {
	m := p.Module
	rhs := lb_create_block(p, "logical.cmp.rhs")
	done := lb_create_block(p, "logical.cmp.done")
	short_circuit := lbValue{}
	if op == Token_CmpAnd {
		lb_build_cond(p, left, rhs, done)
		short_circuit = lb_const_bool(m, t_llvm_bool, false)
	} else if op == Token_CmpOr {
		lb_build_cond(p, left, done, rhs)
		short_circuit = lb_const_bool(m, t_llvm_bool, true)
	}
	if len(rhs.Preds) == 0 {
		lb_start_block(p, done)
		return short_circuit
	}
	if len(done.Preds) == 0 {
		lb_start_block(p, rhs)
		if lb_is_expr_untyped_const(right) {
			return lb_expr_untyped_const_to_typed(m, right, default_type(final_type))
		}
		return lb_build_expr(p, right)
	}
	incoming_values := make([]LLVMValueRef, len(done.Preds)+1)
	incoming_blocks := make([]LLVMBasicBlockRef, len(done.Preds)+1)
	for i := 0; i < len(done.Preds); i++ {
		incoming_values[i] = short_circuit.Value
		incoming_blocks[i] = done.Preds[i].Block
	}
	lb_start_block(p, rhs)
	edge := lbValue{}
	if lb_is_expr_untyped_const(right) {
		edge = lb_expr_untyped_const_to_typed(m, right, t_llvm_bool)
	} else {
		edge = lb_emit_conv(p, lb_build_expr(p, right), t_llvm_bool)
	}
	gb_assert_handler("Assertion Failure", "edge.type == t_llvm_bool", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 58, 0)
	incoming_values[len(done.Preds)] = edge.Value
	incoming_blocks[len(done.Preds)] = p.CurrBlock.Block
	lb_emit_jump(p, done)
	lb_start_block(p, done)
	dst_type := lb_type(m, t_llvm_bool)
	var phi LLVMValueRef
	gb_assert_handler("Assertion Failure", "incoming_values.count == incoming_blocks.count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 69, 0)
	gb_assert_handler("Assertion Failure", "incoming_values.count > 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 70, 0)
	var phi_type LLVMTypeRef
	for _, incoming_value := range incoming_values {
		if LLVMIsConstant(incoming_value) == 0 {
			phi_type = LLVMTypeOf(incoming_value)
			break
		}
	}
	res := lbValue{}
	if phi_type == 0 {
		phi = LLVMBuildPhi(p.Builder, dst_type, "")
		LLVMAddIncoming(phi, incoming_values, incoming_blocks, uint(len(incoming_values)))
		res.Value = phi
		res.Type = t_llvm_bool
	} else {
		for i := 0; i < len(incoming_values); i++ {
			incoming_value := incoming_values[i]
			incoming_type := LLVMTypeOf(incoming_value)
			if phi_type != incoming_type {
				gb_assert_handler("Assertion Failure", "LLVMIsConstant(incoming_value)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 93, LLVMPrintTypeToString(phi_type), LLVMPrintTypeToString(incoming_type))
				ok := LLVMConstIntGetZExtValue(incoming_value) != 0
				var okVal u64
				if ok {
					okVal = 1
				}
				incoming_values[i] = LLVMConstInt(phi_type, okVal, 0)
			}
		}
		phi = LLVMBuildPhi(p.Builder, phi_type, "")
		LLVMAddIncoming(phi, incoming_values, incoming_blocks, uint(len(incoming_values)))
		res.Value = phi
		res.Type = t_llvm_bool
	}
	return lb_emit_conv(p, res, default_type(final_type))
}

func lb_emit_unary_arith(p *lbProcedure, op TokenKind, x lbValue, type_ *Type) lbValue {
	switch op {
	case Token_Add:
		return x
	case Token_Not:
	case Token_Xor:
	case Token_Sub:
		break
	case Token_Pointer:
		gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 121, "This should be handled elsewhere")
		break
	}
	if is_type_array_like(x.Type) {
		tl := base_type(x.Type)
		val := lb_address_from_load_or_generate_local(p, x)
		gb_assert_handler("Assertion Failure", "is_type_array_like(type_)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 129, 0)
		elem_type := base_array_type(type_)
		res_addr := lb_add_local(p, type_, nil, false, true)
		res := lb_addr_get_ptr(p, res_addr)
		inline_array_arith := lb_can_try_to_inline_array_arith(type_)
		count := i32(get_array_type_count(tl))
		var vector_type LLVMTypeRef
		if op != Token_Not && lb_try_vector_cast(p.Module, val, &vector_type) {
			vp := LLVMBuildPointerCast(p.Builder, val.Value, LLVMPointerType(vector_type, 0), "")
			v := OdinLLVMBuildLoad(p, vector_type, vp)
			var opv LLVMValueRef
			switch op {
			case Token_Xor:
				opv = LLVMBuildNot(p.Builder, v, "")
				if is_type_bit_set(elem_type) {
					ev_mask := exact_bit_set_all_set_mask(elem_type)
					mask := lb_const_value(p.Module, elem_type, ev_mask)
					opv = LLVMBuildAnd(p.Builder, opv, mask.Value, "")
				}
			case Token_Sub:
				if is_type_float(elem_type) {
					opv = LLVMBuildFNeg(p.Builder, v, "")
				} else {
					opv = LLVMBuildNeg(p.Builder, v, "")
				}
			}
			if opv != 0 {
				LLVMSetAlignment(res.Value, uint(lb_alignof(vector_type)))
				res_ptr := LLVMBuildPointerCast(p.Builder, res.Value, LLVMPointerType(vector_type, 0), "")
				LLVMBuildStore(p.Builder, opv, res_ptr)
				return lb_emit_conv(p, lb_emit_load(p, res), type_)
			}
		}
		if inline_array_arith {
			for i := i32(0); i < count; i++ {
				e := lb_emit_load(p, lb_emit_array_epi(p, val, isize(i)))
				z := lb_emit_unary_arith(p, op, e, elem_type)
				lb_emit_store(p, lb_emit_array_epi(p, res, isize(i)), z)
			}
		} else {
			loop_data := lb_loop_start(p, isize(count), t_i32)
			e := lb_emit_load(p, lb_emit_array_ep(p, val, loop_data.Idx))
			z := lb_emit_unary_arith(p, op, e, elem_type)
			lb_emit_store(p, lb_emit_array_ep(p, res, loop_data.Idx), z)
			lb_loop_end(p, loop_data)
		}
		return lb_emit_load(p, res)
	}
	if op == Token_Xor {
		cmp := lbValue{}
		cmp.Type = x.Type
		if is_type_bit_set(x.Type) {
			ev_mask := exact_bit_set_all_set_mask(x.Type)
			mask := lb_const_value(p.Module, x.Type, ev_mask)
			cmp.Value = LLVMBuildXor(p.Builder, x.Value, mask.Value, "")
		} else {
			cmp.Value = LLVMBuildNot(p.Builder, x.Value, "")
		}
		return lb_emit_conv(p, cmp, type_)
	}
	if op == Token_Not {
		cmp := lbValue{}
		zero := LLVMConstInt(lb_type(p.Module, x.Type), 0, 0)
		cmp.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, x.Value, zero, "")
		cmp.Type = t_llvm_bool
		return lb_emit_conv(p, cmp, type_)
	}
	if op == Token_Sub && is_type_integer(type_) && is_type_different_to_arch_endianness(type_) {
		platform_type := integer_endian_type_to_platform_type(type_)
		v := lb_emit_byte_swap(p, x, platform_type)
		res := lbValue{}
		res.Value = LLVMBuildNeg(p.Builder, v.Value, "")
		res.Type = platform_type
		return lb_emit_byte_swap(p, res, type_)
	}
	if op == Token_Sub && is_type_float(type_) && is_type_different_to_arch_endianness(type_) {
		platform_type := integer_endian_type_to_platform_type(type_)
		v := lb_emit_byte_swap(p, x, platform_type)
		res := lbValue{}
		res.Value = LLVMBuildFNeg(p.Builder, v.Value, "")
		res.Type = platform_type
		return lb_emit_byte_swap(p, res, type_)
	}
	bt := base_type(type_)
	res := lbValue{}
	switch op {
	case Token_Not:
	case Token_Xor:
		res.Value = LLVMBuildNot(p.Builder, x.Value, "")
		res.Type = x.Type
		return res
	case Token_Sub:
		if is_type_integer(x.Type) {
			res.Value = LLVMBuildNeg(p.Builder, x.Value, "")
		} else if bt.Kind == Type_Enum && is_type_integer(bt.Enum.BaseType) {
			res.Value = LLVMBuildNeg(p.Builder, x.Value, "")
		} else if is_type_float(x.Type) {
			res.Value = LLVMBuildFNeg(p.Builder, x.Value, "")
		} else if is_type_complex(x.Type) {
			v0 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 0, ""), "")
			v1 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 1, ""), "")
			et := base_complex_elem_type(x.Type)
			fields := []lbValue{{v0, et}, {v1, et}}
			return lb_build_struct_value(p, x.Type, fields, isize(len(fields)))
		} else if is_type_quaternion(x.Type) {
			v0 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 0, ""), "")
			v1 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 1, ""), "")
			v2 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 2, ""), "")
			v3 := LLVMBuildFNeg(p.Builder, LLVMBuildExtractValue(p.Builder, x.Value, 3, ""), "")
			et := base_complex_elem_type(x.Type)
			fields := []lbValue{{v0, et}, {v1, et}, {v2, et}, {v3, et}}
			return lb_build_struct_value(p, x.Type, fields, isize(len(fields)))
		} else if is_type_simd_vector(x.Type) {
			elem := base_array_type(x.Type)
			if is_type_float(elem) {
				res.Value = LLVMBuildFNeg(p.Builder, x.Value, "")
			} else {
				res.Value = LLVMBuildNeg(p.Builder, x.Value, "")
			}
		} else if is_type_matrix(x.Type) {
			zero := lbValue{}
			zero.Value = LLVMConstNull(lb_type(p.Module, type_))
			zero.Type = type_
			return lb_emit_arith_matrix(p, Token_Sub, zero, x, type_, true)
		} else {
			gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 280, "Unhandled type %s", type_to_string(x.Type))
		}
		res.Type = x.Type
		return res
	}
	return res
}

func lb_get_file_feature_flags(p *lbProcedure, expr ...*Ast) u64 {
	var file *AstFile
	if len(expr) > 0 && expr[0] != nil {
		file = expr[0].file()
	}
	if file == nil && p.Body != nil && p.Body.file() != nil {
		file = p.Body.file()
	}
	if file == nil && p.TypeExpr != nil && p.TypeExpr.file() != nil {
		file = p.TypeExpr.file()
	}
	if file == nil && p.Entity != nil && p.Entity.File != nil {
		file = p.Entity.File
	}
	if file != nil && file.FeatureFlagsSet {
		return file.FeatureFlags
	}
	return 0
}

func lb_check_for_integer_division_by_zero_behaviour(p *lbProcedure) IntegerDivisionByZeroKind {
	flags := OptInFeatureFlags(lb_get_file_feature_flags(p))
	if flags&OptInFeatureFlagIntegerDivisionByZeroTrap != 0 {
		return IntegerDivisionByZeroTrap
	}
	if flags&OptInFeatureFlagIntegerDivisionByZeroZero != 0 {
		return IntegerDivisionByZeroZero
	}
	if flags&OptInFeatureFlagIntegerDivisionByZeroSelf != 0 {
		return IntegerDivisionByZeroSelf
	}
	if flags&OptInFeatureFlagIntegerDivisionByZeroAllBits != 0 {
		return IntegerDivisionByZeroAllBits
	}
	return buildContext.IntegerDivisionByZeroBehaviour
}

func is_simd_able_type(t *Type) bool {
	if t.Kind != Type_Basic {
		return false
	}
	if t.Basic.Flags&(BasicFlagBoolean|BasicFlagInteger|BasicFlagFloat|BasicFlagRune) != 0 {
		kind := type_endian_kind_of(t)
		switch kind {
		case TypeEndianPlatform:
			return true
		case TypeEndianLittle:
			return buildContext.EndianKind == TargetEndianLittle
		case TypeEndianBig:
			return buildContext.EndianKind == TargetEndianBig
		}
	}
	return false
}

func lb_try_direct_vector_arith(p *lbProcedure, op TokenKind, lhs lbValue, rhs lbValue, type_ *Type, res_ *lbValue) bool {
	gb_assert_handler("Assertion Failure", "is_type_array_like(type_)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 354, 0)
	elem_type := base_array_type(type_)
	if op == Token_Shl || op == Token_Shr {
		return false
	}
	if LLVMIsALoadInst(lhs.Value) == 0 || LLVMIsALoadInst(rhs.Value) == 0 {
		return false
	}
	lhs_ptr := lbValue{}
	rhs_ptr := lbValue{}
	lhs_ptr.Value = LLVMGetOperand(lhs.Value, 0)
	lhs_ptr.Type = alloc_type_pointer(lhs.Type)
	rhs_ptr.Value = LLVMGetOperand(rhs.Value, 0)
	rhs_ptr.Type = alloc_type_pointer(rhs.Type)
	var vector_type0 LLVMTypeRef
	var vector_type1 LLVMTypeRef
	if lb_try_vector_cast(p.Module, lhs_ptr, &vector_type0) &&
		lb_try_vector_cast(p.Module, rhs_ptr, &vector_type1) {
		gb_assert_handler("Assertion Failure", "vector_type0 == vector_type1", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 377, 0)
		vector_type := vector_type0
		integral_type := base_type(elem_type)
		if is_type_simd_vector(integral_type) {
			integral_type = core_array_type(integral_type)
		}
		if is_type_bit_set(integral_type) {
			switch op {
			case Token_Add:
				op = Token_Or
			case Token_Sub:
				op = Token_AndNot
			}
			u := bit_set_to_int(type_)
			if is_type_array(u) {
				return false
			}
		}
		lhs_vp := LLVMBuildPointerCast(p.Builder, lhs_ptr.Value, LLVMPointerType(vector_type, 0), "")
		rhs_vp := LLVMBuildPointerCast(p.Builder, rhs_ptr.Value, LLVMPointerType(vector_type, 0), "")
		x := OdinLLVMBuildLoad(p, vector_type, lhs_vp)
		y := OdinLLVMBuildLoad(p, vector_type, rhs_vp)
		var z LLVMValueRef
		if is_type_float(integral_type) {
			switch op {
			case Token_Add:
				z = LLVMBuildFAdd(p.Builder, x, y, "")
			case Token_Sub:
				z = LLVMBuildFSub(p.Builder, x, y, "")
			case Token_Mul:
				z = LLVMBuildFMul(p.Builder, x, y, "")
			case Token_Quo:
				z = LLVMBuildFDiv(p.Builder, x, y, "")
			case Token_Mod:
				z = LLVMBuildFRem(p.Builder, x, y, "")
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 419, "Unsupported vector operation %s", goStr(token_strings[op]))
			}
		} else {
			switch op {
			case Token_Add:
				z = LLVMBuildAdd(p.Builder, x, y, "")
			case Token_Sub:
				z = LLVMBuildSub(p.Builder, x, y, "")
			case Token_Mul:
				z = LLVMBuildMul(p.Builder, x, y, "")
			case Token_Quo:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildUDiv(p.Builder, x, y, "")
				} else {
					z = LLVMBuildSDiv(p.Builder, x, y, "")
				}
			case Token_Mod:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildURem(p.Builder, x, y, "")
				} else {
					z = LLVMBuildSRem(p.Builder, x, y, "")
				}
			case Token_ModMod:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildURem(p.Builder, x, y, "")
				} else {
					a := LLVMBuildSRem(p.Builder, x, y, "")
					b := LLVMBuildAdd(p.Builder, a, y, "")
					z = LLVMBuildSRem(p.Builder, b, y, "")
				}
			case Token_And:
				z = LLVMBuildAnd(p.Builder, x, y, "")
			case Token_AndNot:
				z = LLVMBuildAnd(p.Builder, x, LLVMBuildNot(p.Builder, y, ""), "")
			case Token_Or:
				z = LLVMBuildOr(p.Builder, x, y, "")
			case Token_Xor:
				z = LLVMBuildXor(p.Builder, x, y, "")
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 468, "Unsupported vector operation")
			}
		}
		if z != 0 {
			res := lb_add_local_generated_temp(p, type_, lb_alignof(vector_type))
			vp := LLVMBuildPointerCast(p.Builder, res.Addr.Value, LLVMPointerType(vector_type, 0), "")
			LLVMBuildStore(p.Builder, z, vp)
			v := lb_addr_load(p, res)
			if res_ != nil {
				*res_ = v
			}
			return true
		}
	}
	count := get_array_type_count(type_)
	if false &&
		count <= 64 &&
		is_simd_able_type(elem_type) && is_simd_able_type(elem_type) {
		integral_type := elem_type
		vector_elem_type := lb_type(p.Module, integral_type)
		vector_type := LLVMVectorType(vector_elem_type, uint(count))
		if is_type_bit_set(integral_type) {
			switch op {
			case Token_Add:
				op = Token_Or
			case Token_Sub:
				op = Token_AndNot
			}
			u := bit_set_to_int(type_)
			if is_type_array(u) {
				return false
			}
		}
		lhs_vp := LLVMBuildPointerCast(p.Builder, lhs_ptr.Value, LLVMPointerType(vector_type, 0), "")
		rhs_vp := LLVMBuildPointerCast(p.Builder, rhs_ptr.Value, LLVMPointerType(vector_type, 0), "")
		x := OdinLLVMBuildLoad(p, vector_type, lhs_vp)
		y := OdinLLVMBuildLoad(p, vector_type, rhs_vp)
		LLVMSetAlignment(x, uint(type_align_of(integral_type)))
		LLVMSetAlignment(y, uint(type_align_of(integral_type)))
		var z LLVMValueRef
		if is_type_float(integral_type) {
			switch op {
			case Token_Add:
				z = LLVMBuildFAdd(p.Builder, x, y, "")
			case Token_Sub:
				z = LLVMBuildFSub(p.Builder, x, y, "")
			case Token_Mul:
				z = LLVMBuildFMul(p.Builder, x, y, "")
			case Token_Quo:
				z = LLVMBuildFDiv(p.Builder, x, y, "")
			case Token_Mod:
				z = LLVMBuildFRem(p.Builder, x, y, "")
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 533, "Unsupported vector operation %s", goStr(token_strings[op]))
			}
		} else {
			switch op {
			case Token_Add:
				z = LLVMBuildAdd(p.Builder, x, y, "")
			case Token_Sub:
				z = LLVMBuildSub(p.Builder, x, y, "")
			case Token_Mul:
				z = LLVMBuildMul(p.Builder, x, y, "")
			case Token_Quo:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildUDiv(p.Builder, x, y, "")
				} else {
					z = LLVMBuildSDiv(p.Builder, x, y, "")
				}
			case Token_Mod:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildURem(p.Builder, x, y, "")
				} else {
					z = LLVMBuildSRem(p.Builder, x, y, "")
				}
			case Token_ModMod:
				if is_type_unsigned(integral_type) {
					z = LLVMBuildURem(p.Builder, x, y, "")
				} else {
					a := LLVMBuildSRem(p.Builder, x, y, "")
					b := LLVMBuildAdd(p.Builder, a, y, "")
					z = LLVMBuildSRem(p.Builder, b, y, "")
				}
			case Token_And:
				z = LLVMBuildAnd(p.Builder, x, y, "")
			case Token_AndNot:
				z = LLVMBuildAnd(p.Builder, x, LLVMBuildNot(p.Builder, y, ""), "")
			case Token_Or:
				z = LLVMBuildOr(p.Builder, x, y, "")
			case Token_Xor:
				z = LLVMBuildXor(p.Builder, x, y, "")
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 582, "Unsupported vector operation")
			}
		}
		if z != 0 {
			res := lb_add_local_generated_temp(p, type_, lb_alignof(vector_type))
			vp := LLVMBuildPointerCast(p.Builder, res.Addr.Value, LLVMPointerType(vector_type, 0), "")
			store := LLVMBuildStore(p.Builder, z, vp)
			LLVMSetAlignment(store, uint(type_align_of(type_)))
			v := lb_addr_load(p, res)
			if res_ != nil {
				*res_ = v
			}
			return true
		}
	}
	return false
}

func lb_emit_arith_array(p *lbProcedure, op TokenKind, lhs lbValue, rhs lbValue, type_ *Type) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_array_like(lhs.type) || is_type_array_like(rhs.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 605, 0)
	lhs = lb_emit_conv(p, lhs, type_)
	rhs = lb_emit_conv(p, rhs, type_)
	gb_assert_handler("Assertion Failure", "is_type_array_like(type_)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 610, 0)
	elem_type := base_array_type(type_)
	count := get_array_type_count(type_)
	n := uint(count)
	direct_vector_res := lbValue{}
	if lb_try_direct_vector_arith(p, op, lhs, rhs, type_, &direct_vector_res) {
		return direct_vector_res
	}
	inline_array_arith := lb_can_try_to_inline_array_arith(type_)
	if inline_array_arith {
		dst_ptrs := make([]lbValue, n)
		a_loads := make([]lbValue, n)
		b_loads := make([]lbValue, n)
		c_ops := make([]lbValue, n)
		for i := uint(0); i < n; i++ {
			a_loads[i].Value = LLVMBuildExtractValue(p.Builder, lhs.Value, i, "")
			a_loads[i].Type = elem_type
		}
		for i := uint(0); i < n; i++ {
			b_loads[i].Value = LLVMBuildExtractValue(p.Builder, rhs.Value, i, "")
			b_loads[i].Type = elem_type
		}
		for i := uint(0); i < n; i++ {
			c_ops[i] = lb_emit_arith(p, op, a_loads[i], b_loads[i], elem_type)
		}
		res := lb_add_local_generated(p, type_, false)
		for i := uint(0); i < n; i++ {
			dst_ptrs[i] = lb_emit_array_epi(p, res.Addr, isize(i))
		}
		for i := uint(0); i < n; i++ {
			lb_emit_store(p, dst_ptrs[i], c_ops[i])
		}
		return lb_addr_load(p, res)
	} else {
		x := lb_address_from_load_or_generate_local(p, lhs)
		y := lb_address_from_load_or_generate_local(p, rhs)
		res := lb_add_local_generated(p, type_, false)
		loop_data := lb_loop_start(p, isize(count), t_i32)
		a_ptr := lb_emit_array_ep(p, x, loop_data.Idx)
		b_ptr := lb_emit_array_ep(p, y, loop_data.Idx)
		dst_ptr := lb_emit_array_ep(p, res.Addr, loop_data.Idx)
		a := lb_emit_load(p, a_ptr)
		b := lb_emit_load(p, b_ptr)
		c := lb_emit_arith(p, op, a, b, elem_type)
		lb_emit_store(p, dst_ptr, c)
		lb_loop_end(p, loop_data)
		return lb_addr_load(p, res)
	}
}

func lb_integer_division(p *lbProcedure, lhs LLVMValueRef, rhs LLVMValueRef, is_signed bool) LLVMValueRef {
	type_ := LLVMTypeOf(rhs)
	gb_assert_handler("Assertion Failure", "LLVMTypeOf(lhs) == type_", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1309, 0)
	zero := LLVMConstNull(type_)
	all_bits := LLVMConstNot(zero)
	behaviour := lb_check_for_integer_division_by_zero_behaviour(p)
	var call func(B LLVMBuilderRef, V LLVMValueRef, W LLVMValueRef, Name string) LLVMValueRef
	if is_signed {
		call = LLVMBuildSDiv
	} else {
		call = LLVMBuildUDiv
	}
	if LLVMIsConstant(rhs) != 0 {
		if LLVMIsNull(rhs) != 0 {
			switch behaviour {
			case IntegerDivisionByZeroSelf:
				return lhs
			case IntegerDivisionByZeroZero:
				return zero
			case IntegerDivisionByZeroAllBits:
				return all_bits
			}
		} else {
			if !is_signed && lb_sizeof(type_) <= 8 {
				v := LLVMConstIntGetZExtValue(rhs)
				if v == 1 {
					return lhs
				} else if is_power_of_two_u64(v) {
					n := floor_log2(v)
					bits := LLVMConstInt(type_, n, 0)
					return LLVMBuildLShr(p.Builder, lhs, bits, "")
				}
			}
			return call(p.Builder, lhs, rhs, "")
		}
	}
	var incoming_values [2]LLVMValueRef
	var incoming_blocks [2]LLVMBasicBlockRef
	safe_block := lb_create_block(p, "div.safe")
	edge_case_block := lb_create_block(p, "div.edge")
	done_block := lb_create_block(p, "div.done")
	dem_check := LLVMBuildICmp(p.Builder, LLVMIntNE, rhs, zero, "")
	cond := lbValue{dem_check, t_untyped_bool}
	lb_emit_if(p, cond, safe_block, edge_case_block)
	lb_start_block(p, safe_block)
	incoming_values[0] = call(p.Builder, lhs, rhs, "")
	lb_emit_jump(p, done_block)
	lb_start_block(p, edge_case_block)
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		lb_call_intrinsic(p, "llvm.trap", nil, 0, nil, 0)
		LLVMBuildUnreachable(p.Builder)
	case IntegerDivisionByZeroZero:
		incoming_values[1] = zero
	case IntegerDivisionByZeroSelf:
		incoming_values[1] = lhs
	case IntegerDivisionByZeroAllBits:
		incoming_values[1] = all_bits
	}
	var res LLVMValueRef
	lb_emit_jump(p, done_block)
	lb_start_block(p, done_block)
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		res = incoming_values[0]
	case IntegerDivisionByZeroSelf:
	case IntegerDivisionByZeroZero:
	case IntegerDivisionByZeroAllBits:
		res = LLVMBuildPhi(p.Builder, type_, "")
		gb_assert_handler("Assertion Failure", "p->curr_block->preds.count >= 2", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1391, 0)
		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block
		gb_assert_handler("Assertion Failure", "incoming_blocks[0] == safe_block->block", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1395, 0)
		gb_assert_handler("Assertion Failure", "incoming_blocks[1] == edge_case_block->block", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1396, 0)
		LLVMAddIncoming(res, incoming_values[:], incoming_blocks[:], 2)
	}
	return res
}

func lb_integer_division_fixed_point_intrinsics(p *lbProcedure, lhs LLVMValueRef, rhs LLVMValueRef, scale LLVMValueRef, platform_type *Type, name string) LLVMValueRef {
	type_ := LLVMTypeOf(rhs)
	gb_assert_handler("Assertion Failure", "LLVMTypeOf(lhs) == type_", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1407, 0)
	zero := LLVMConstNull(type_)
	all_bits := LLVMConstNot(zero)
	behaviour := lb_check_for_integer_division_by_zero_behaviour(p)
	do_op := func() LLVMValueRef {
		types := []LLVMTypeRef{lb_type(p.Module, platform_type)}
		args := []LLVMValueRef{lhs, rhs, scale}
		return lb_call_intrinsic(p, name, args, uint(len(args)), types, uint(len(types)))
	}
	if LLVMIsConstant(rhs) != 0 {
		if LLVMIsNull(rhs) != 0 {
			switch behaviour {
			case IntegerDivisionByZeroSelf:
				return lhs
			case IntegerDivisionByZeroZero:
				return zero
			}
		} else {
			return do_op()
		}
	}
	var incoming_values [2]LLVMValueRef
	var incoming_blocks [2]LLVMBasicBlockRef
	safe_block := lb_create_block(p, "div.safe")
	edge_case_block := lb_create_block(p, "div.edge")
	done_block := lb_create_block(p, "div.done")
	dem_check := LLVMBuildICmp(p.Builder, LLVMIntNE, rhs, zero, "")
	cond := lbValue{dem_check, t_untyped_bool}
	lb_emit_if(p, cond, safe_block, edge_case_block)
	lb_start_block(p, safe_block)
	incoming_values[0] = do_op()
	lb_emit_jump(p, done_block)
	lb_start_block(p, edge_case_block)
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		lb_call_intrinsic(p, "llvm.trap", nil, 0, nil, 0)
		LLVMBuildUnreachable(p.Builder)
	case IntegerDivisionByZeroZero:
		incoming_values[1] = zero
	case IntegerDivisionByZeroSelf:
		incoming_values[1] = lhs
	case IntegerDivisionByZeroAllBits:
		incoming_values[1] = all_bits
	}
	lb_emit_jump(p, done_block)
	lb_start_block(p, done_block)
	var res LLVMValueRef
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		res = incoming_values[0]
	case IntegerDivisionByZeroSelf:
	case IntegerDivisionByZeroZero:
	case IntegerDivisionByZeroAllBits:
		res = LLVMBuildPhi(p.Builder, type_, "")
		gb_assert_handler("Assertion Failure", "p->curr_block->preds.count >= 2", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1486, 0)
		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block
		LLVMAddIncoming(res, incoming_values[:], incoming_blocks[:], 2)
	}
	return res
}

func lb_integer_modulo(p *lbProcedure, lhs LLVMValueRef, rhs LLVMValueRef, is_unsigned bool, is_floored bool) LLVMValueRef {
	type_ := LLVMTypeOf(rhs)
	gb_assert_handler("Assertion Failure", "LLVMTypeOf(lhs) == type_", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1500, 0)
	zero := LLVMConstNull(type_)
	behaviour := lb_check_for_integer_division_by_zero_behaviour(p)
	do_op := func() LLVMValueRef {
		if is_floored {
			if is_unsigned {
				return LLVMBuildURem(p.Builder, lhs, rhs, "")
			} else {
				a := LLVMBuildSRem(p.Builder, lhs, rhs, "")
				b := LLVMBuildAdd(p.Builder, a, rhs, "")
				c := LLVMBuildSRem(p.Builder, b, rhs, "")
				return c
			}
		} else {
			if is_unsigned {
				return LLVMBuildURem(p.Builder, lhs, rhs, "")
			} else {
				return LLVMBuildSRem(p.Builder, lhs, rhs, "")
			}
		}
	}
	if LLVMIsConstant(rhs) != 0 {
		if LLVMIsNull(rhs) != 0 {
			switch behaviour {
			case IntegerDivisionByZeroTrap:
				break
			case IntegerDivisionByZeroSelf:
				return zero
			case IntegerDivisionByZeroZero:
			case IntegerDivisionByZeroAllBits:
				return lhs
			}
		} else {
			return do_op()
		}
	}
	var incoming_values [2]LLVMValueRef
	var incoming_blocks [2]LLVMBasicBlockRef
	safe_block := lb_create_block(p, "mod.safe")
	edge_case_block := lb_create_block(p, "mod.edge")
	done_block := lb_create_block(p, "mod.done")
	dem_check := LLVMBuildICmp(p.Builder, LLVMIntNE, rhs, zero, "")
	cond := lbValue{dem_check, t_untyped_bool}
	lb_emit_if(p, cond, safe_block, edge_case_block)
	lb_start_block(p, safe_block)
	incoming_values[0] = do_op()
	lb_emit_jump(p, done_block)
	lb_start_block(p, edge_case_block)
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		lb_call_intrinsic(p, "llvm.trap", nil, 0, nil, 0)
		LLVMBuildUnreachable(p.Builder)
	case IntegerDivisionByZeroZero:
	case IntegerDivisionByZeroAllBits:
		incoming_values[1] = lhs
	case IntegerDivisionByZeroSelf:
		incoming_values[1] = zero
	}
	lb_emit_jump(p, done_block)
	lb_start_block(p, done_block)
	res := incoming_values[0]
	switch behaviour {
	case IntegerDivisionByZeroTrap:
		res = incoming_values[0]
	case IntegerDivisionByZeroSelf:
	case IntegerDivisionByZeroZero:
	case IntegerDivisionByZeroAllBits:
		res = LLVMBuildPhi(p.Builder, type_, "")
		gb_assert_handler("Assertion Failure", "p->curr_block->preds.count >= 2", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1597, 0)
		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block
		LLVMAddIncoming(res, incoming_values[:], incoming_blocks[:], 2)
	}
	return res
}

func lb_emit_arith(p *lbProcedure, op TokenKind, lhs lbValue, rhs lbValue, type_ *Type) lbValue {
	if is_type_array_like(lhs.Type) || is_type_array_like(rhs.Type) {
		return lb_emit_arith_array(p, op, lhs, rhs, type_)
	} else if is_type_matrix(lhs.Type) || is_type_matrix(rhs.Type) {
		return lb_emit_arith_matrix(p, op, lhs, rhs, type_, false)
	} else if is_type_complex(type_) {
		lhs = lb_emit_conv(p, lhs, type_)
		rhs = lb_emit_conv(p, rhs, type_)
		ft := base_complex_elem_type(type_)
		if op == Token_Quo {
			args := make([]lbValue, 2)
			args[0] = lhs
			args[1] = rhs
			switch type_size_of(ft) {
			case 2:
				return lb_emit_runtime_call(p, "quo_complex32", args)
			case 4:
				return lb_emit_runtime_call(p, "quo_complex64", args)
			case 8:
				return lb_emit_runtime_call(p, "quo_complex128", args)
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1631, "Unknown float type")
			}
		}
		a := lb_emit_struct_ev(p, lhs, 0)
		b := lb_emit_struct_ev(p, lhs, 1)
		c := lb_emit_struct_ev(p, rhs, 0)
		d := lb_emit_struct_ev(p, rhs, 1)
		var real lbValue
		var imag lbValue
		switch op {
		case Token_Add:
		case Token_Sub:
			if type_size_of(ft) == 2 {
				a = lb_emit_conv(p, a, t_f32)
				b = lb_emit_conv(p, b, t_f32)
				c = lb_emit_conv(p, c, t_f32)
				d = lb_emit_conv(p, d, t_f32)
				real = lb_emit_arith(p, op, a, c, t_f32)
				imag = lb_emit_arith(p, op, b, d, t_f32)
				real = lb_emit_conv(p, real, ft)
				imag = lb_emit_conv(p, imag, ft)
			} else {
				real = lb_emit_arith(p, op, a, c, ft)
				imag = lb_emit_arith(p, op, b, d, ft)
			}
		case Token_Mul:
			x := lb_emit_arith(p, Token_Mul, a, c, ft)
			y := lb_emit_arith(p, Token_Mul, b, d, ft)
			real = lb_emit_arith(p, Token_Sub, x, y, ft)
			z := lb_emit_arith(p, Token_Mul, b, c, ft)
			w := lb_emit_arith(p, Token_Mul, a, d, ft)
			imag = lb_emit_arith(p, Token_Add, z, w, ft)
		}
		fields := []lbValue{real, imag}
		return lb_build_struct_value(p, type_, fields, isize(len(fields)))
	} else if is_type_quaternion(type_) {
		lhs = lb_emit_conv(p, lhs, type_)
		rhs = lb_emit_conv(p, rhs, type_)
		ft := base_complex_elem_type(type_)
		if op == Token_Add || op == Token_Sub {
			immediate_type := ft
			if type_size_of(ft) == 2 {
				immediate_type = t_f32
			}
			x0 := lb_emit_struct_ev(p, lhs, 0)
			x1 := lb_emit_struct_ev(p, lhs, 1)
			x2 := lb_emit_struct_ev(p, lhs, 2)
			x3 := lb_emit_struct_ev(p, lhs, 3)
			y0 := lb_emit_struct_ev(p, rhs, 0)
			y1 := lb_emit_struct_ev(p, rhs, 1)
			y2 := lb_emit_struct_ev(p, rhs, 2)
			y3 := lb_emit_struct_ev(p, rhs, 3)
			if immediate_type != ft {
				x0 = lb_emit_conv(p, x0, immediate_type)
				x1 = lb_emit_conv(p, x1, immediate_type)
				x2 = lb_emit_conv(p, x2, immediate_type)
				x3 = lb_emit_conv(p, x3, immediate_type)
				y0 = lb_emit_conv(p, y0, immediate_type)
				y1 = lb_emit_conv(p, y1, immediate_type)
				y2 = lb_emit_conv(p, y2, immediate_type)
				y3 = lb_emit_conv(p, y3, immediate_type)
			}
			z0 := lb_emit_arith(p, op, x0, y0, immediate_type)
			z1 := lb_emit_arith(p, op, x1, y1, immediate_type)
			z2 := lb_emit_arith(p, op, x2, y2, immediate_type)
			z3 := lb_emit_arith(p, op, x3, y3, immediate_type)
			if immediate_type != ft {
				z0 = lb_emit_conv(p, z0, ft)
				z1 = lb_emit_conv(p, z1, ft)
				z2 = lb_emit_conv(p, z2, ft)
				z3 = lb_emit_conv(p, z3, ft)
			}
			fields := []lbValue{z0, z1, z2, z3}
			return lb_build_struct_value(p, type_, fields, isize(len(fields)))
		} else if op == Token_Mul {
			args := make([]lbValue, 2)
			args[0] = lhs
			args[1] = rhs
			switch 8 * type_size_of(ft) {
			case 16:
				return lb_emit_runtime_call(p, "mul_quaternion64", args)
			case 32:
				return lb_emit_runtime_call(p, "mul_quaternion128", args)
			case 64:
				return lb_emit_runtime_call(p, "mul_quaternion256", args)
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1735, "Unknown float type")
			}
		} else if op == Token_Quo {
			args := make([]lbValue, 2)
			args[0] = lhs
			args[1] = rhs
			switch 8 * type_size_of(ft) {
			case 16:
				return lb_emit_runtime_call(p, "quo_quaternion64", args)
			case 32:
				return lb_emit_runtime_call(p, "quo_quaternion128", args)
			case 64:
				return lb_emit_runtime_call(p, "quo_quaternion256", args)
			default:
				gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1748, "Unknown float type")
			}
		}
	}
	lhs = lb_emit_conv(p, lhs, type_)
	rhs = lb_emit_conv(p, rhs, type_)
	if is_type_integer(type_) && is_type_different_to_arch_endianness(type_) {
		switch op {
		case Token_AndNot:
		case Token_And:
		case Token_Or:
		case Token_Xor:
			goto handle_op
		}
		platform_type := integer_endian_type_to_platform_type(type_)
		x := lb_emit_byte_swap(p, lhs, integer_endian_type_to_platform_type(lhs.Type))
		y := lb_emit_byte_swap(p, rhs, integer_endian_type_to_platform_type(rhs.Type))
		res := lb_emit_arith(p, op, x, y, platform_type)
		return lb_emit_byte_swap(p, res, type_)
	}
	if is_type_float(type_) && is_type_different_to_arch_endianness(type_) {
		platform_type := integer_endian_type_to_platform_type(type_)
		x := lb_emit_conv(p, lhs, integer_endian_type_to_platform_type(lhs.Type))
		y := lb_emit_conv(p, rhs, integer_endian_type_to_platform_type(rhs.Type))
		res := lb_emit_arith(p, op, x, y, platform_type)
		return lb_emit_byte_swap(p, res, type_)
	}
handle_op:
	res := lbValue{}
	res.Type = type_
	if is_type_bit_set(type_) {
		switch op {
		case Token_Add:
			op = Token_Or
		case Token_Sub:
			op = Token_AndNot
		}
		u := bit_set_to_int(type_)
		if is_type_array(u) {
			lhs.Type = u
			rhs.Type = u
			res = lb_emit_arith(p, op, lhs, rhs, u)
			res.Type = type_
			return res
		}
	}
	integral_type := type_
	if is_type_simd_vector(integral_type) {
		integral_type = core_array_type(integral_type)
	}
	switch op {
	case Token_Add:
		if is_type_float(integral_type) {
			res.Value = LLVMBuildFAdd(p.Builder, lhs.Value, rhs.Value, "")
			return res
		}
		res.Value = LLVMBuildAdd(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Sub:
		if is_type_float(integral_type) {
			res.Value = LLVMBuildFSub(p.Builder, lhs.Value, rhs.Value, "")
			return res
		}
		res.Value = LLVMBuildSub(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Mul:
		if is_type_float(integral_type) {
			res.Value = LLVMBuildFMul(p.Builder, lhs.Value, rhs.Value, "")
			return res
		}
		res.Value = LLVMBuildMul(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Quo:
		if is_type_float(integral_type) {
			res.Value = LLVMBuildFDiv(p.Builder, lhs.Value, rhs.Value, "")
			return res
		} else {
			res.Value = lb_integer_division(p, lhs.Value, rhs.Value, !is_type_unsigned(integral_type))
			return res
		}
	case Token_Mod:
		if is_type_float(integral_type) {
			res.Value = LLVMBuildFRem(p.Builder, lhs.Value, rhs.Value, "")
			return res
		}
		res.Value = lb_integer_modulo(p, lhs.Value, rhs.Value, is_type_unsigned(integral_type), false)
		return res
	case Token_ModMod:
		res.Value = lb_integer_modulo(p, lhs.Value, rhs.Value, is_type_unsigned(integral_type), true)
		return res
	case Token_And:
		res.Value = LLVMBuildAnd(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Or:
		res.Value = LLVMBuildOr(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Xor:
		res.Value = LLVMBuildXor(p.Builder, lhs.Value, rhs.Value, "")
		return res
	case Token_Shl:
		rhs = lb_emit_conv(p, rhs, lhs.Type)
		lhsval := lhs.Value
		bits := rhs.Value
		bit_size := LLVMConstInt(lb_type(p.Module, rhs.Type), 8*type_size_of(lhs.Type), 0)
		width_test := LLVMBuildICmp(p.Builder, LLVMIntULT, bits, bit_size, "")
		res.Value = LLVMBuildShl(p.Builder, lhsval, bits, "")
		zero := LLVMConstNull(lb_type(p.Module, lhs.Type))
		res.Value = LLVMBuildSelect(p.Builder, width_test, res.Value, zero, "")
		return res
	case Token_Shr:
		rhs = lb_emit_conv(p, rhs, lhs.Type)
		lhsval := lhs.Value
		bits := rhs.Value
		is_unsigned := is_type_unsigned(integral_type)
		bit_size := LLVMConstInt(lb_type(p.Module, rhs.Type), 8*type_size_of(lhs.Type), 0)
		width_test := LLVMBuildICmp(p.Builder, LLVMIntULT, bits, bit_size, "")
		if is_unsigned {
			res.Value = LLVMBuildLShr(p.Builder, lhsval, bits, "")
		} else {
			res.Value = LLVMBuildAShr(p.Builder, lhsval, bits, "")
		}
		zero := LLVMConstNull(lb_type(p.Module, lhs.Type))
		res.Value = LLVMBuildSelect(p.Builder, width_test, res.Value, zero, "")
		return res
	case Token_AndNot:
		new_rhs := LLVMBuildNot(p.Builder, rhs.Value, "")
		res.Value = LLVMBuildAnd(p.Builder, lhs.Value, new_rhs, "")
		return res
	}
	gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1904, "unhandled operator of lb_emit_arith")
	return lbValue{}
}

func lb_emit_c_vararg(p *lbProcedure, arg lbValue, type_ *Type) lbValue {
	core := core_type(type_)
	if core.Kind == Type_BitSet {
		core = core_type(bit_set_to_int(core))
		arg = lb_emit_transmute(p, arg, core)
	}
	promoted := c_vararg_promote_type(core)
	return lb_emit_conv(p, arg, promoted)
}
