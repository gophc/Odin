package cmd

func llvm_splat_int(count i64, typ LLVMTypeRef, value i64, is_signed ...bool) LLVMValueRef {
	signed := false
	if len(is_signed) > 0 {
		signed = is_signed[0]
	}
	v := LLVMConstInt(typ, uint64(value), signed)
	values := make([]LLVMValueRef, count)
	for i := i64(0); i < count; i++ {
		values[i] = v
	}
	return LLVMConstVector(values, uint(count))
}

func lb_llvm_simd_bulk_op_unary(p *lbProcedure, arg lbValue, result_ *LLVMValueRef, intrinsic_name string, default_width uint, extra_args []LLVMValueRef, extra_args_count uint) bool {
	if intrinsic_name == "" || default_width == 0 {
		return false
	}
	do_call_intrinsic := func(p *lbProcedure, intrinsic_name string, val LLVMValueRef, extra_args []LLVMValueRef, extra_args_count uint) LLVMValueRef {
		if extra_args != nil && extra_args_count > 0 {
			args := make([]LLVMValueRef, extra_args_count+1)
			args[0] = val
			copy(args[1:], extra_args[:extra_args_count])
			return lb_call_intrinsic(p, intrinsic_name, args, extra_args_count+1, nil, 0)
		} else {
			args := [1]LLVMValueRef{val}
			return lb_call_intrinsic(p, intrinsic_name, args[:], uint(len(args)), nil, 0)
		}
	}
	vt := base_type(arg.Type)
	if !is_type_simd_vector(vt) {
		gb_assert_handler("Assertion Failure", "is_type_simd_vector(vt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1354), 0)
	}
	val := arg.Value
	count := uint(vt.SimdVector.Count)
	elem := base_type(vt.SimdVector.Elem)
	if count < default_width {
		grow_indices := make([]LLVMValueRef, default_width)
		shrink_indices := make([]LLVMValueRef, count)
		for i := uint(0); i < count; i++ {
			var idx ExactValue
			if is_type_float(elem) {
				idx = exact_value_float(float64(i))
			} else {
				idx = exact_value_u64(uint64(i))
			}
			shrink_indices[i] = lb_const_value(p.Module, elem, idx).Value
			grow_indices[i] = shrink_indices[i]
		}
		for i := count; i < default_width; i++ {
			grow_indices[i] = shrink_indices[0]
		}
		grow_mask := LLVMConstVector(grow_indices, default_width)
		shrink_mask := LLVMConstVector(shrink_indices, count)
		val = LLVMBuildShuffleVector(p.Builder, val, val, grow_mask, "")
		val = do_call_intrinsic(p, intrinsic_name, val, extra_args, extra_args_count)
		val = LLVMBuildShuffleVector(p.Builder, val, val, shrink_mask, "")
		*result_ = val
		return true
	} else if count == default_width {
		val = do_call_intrinsic(p, intrinsic_name, val, extra_args, extra_args_count)
		*result_ = val
		return true
	} else {
		if !(count > default_width) {
			gb_assert_handler("Assertion Failure", "count > default_width", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1389), 0)
		}
		parts_count := count / default_width
		parts := make([]LLVMValueRef, parts_count)
		for i := uint(0); i < parts_count; i++ {
			indices_part := make([]LLVMValueRef, default_width)
			for j := uint(0); j < default_width; j++ {
				indices_part[j] = lb_const_value(p.Module, t_u32, exact_value_u64(uint64(4*uint(i)+j))).Value
			}
			parts[i] = LLVMBuildShuffleVector(p.Builder, val, val, LLVMConstVector(indices_part, default_width), "")
		}
		for i := uint(0); i < parts_count; i++ {
			parts[i] = do_call_intrinsic(p, intrinsic_name, parts[i], extra_args, extra_args_count)
		}
		indices := make([]LLVMValueRef, count)
		for j := uint(0); j < count; j++ {
			indices[j] = lb_const_value(p.Module, t_u32, exact_value_i64(int64(j))).Value
		}
		sub_parts_remaining := uint64(parts_count)
		for sub_parts_remaining > 1 {
			for i := uint64(0); i < sub_parts_remaining; i += 2 {
				indices_count := 2 * uint((uint64(count) / sub_parts_remaining))
				parts[i] = LLVMBuildShuffleVector(p.Builder, parts[i], parts[i+1], LLVMConstVector(indices, indices_count), "")
			}
			for i := uint64(2); i < sub_parts_remaining; i += 2 {
				parts[i/2] = parts[i]
			}
			sub_parts_remaining >>= 1
		}
		*result_ = parts[0]
		return true
	}
}

func lb_build_builtin_simd_proc(p *lbProcedure, expr *Ast, tv TypeAndValue, builtin_id BuiltinProcId) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == GB_JOIN2(Ast_, CallExpr)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1434), "expected '%.*s' got '%.*s'", int(len(ast_strings[Ast_CallExpr])), ast_strings[Ast_CallExpr].Text, int(len(ast_strings[expr.Kind])), ast_strings[expr.Kind].Text)
	}
	m := p.Module
	res := lbValue{}
	res.Type = tv.Type
	switch builtin_id {
	case BuiltinProc_simd_indices:
		typ := base_type(res.Type)
		if typ.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "type->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1444), 0)
		}
		elem := typ.SimdVector.Elem
		count := typ.SimdVector.Count
		scalars := make([]LLVMValueRef, count)
		for i := i64(0); i < count; i++ {
			scalars[i] = lb_const_value(m, elem, exact_value_i64(i)).Value
		}
		res.Value = LLVMConstVector(scalars, uint(count))
		return res
	}
	arg0 := lbValue{}
	if ce.Args.Count > 0 {
		arg0 = lb_build_expr(p, ce.Args[0])
	}
	arg1 := lbValue{}
	if ce.Args.Count > 1 {
		arg1 = lb_build_expr(p, ce.Args[1])
	}
	arg2 := lbValue{}
	if ce.Args.Count > 2 {
		arg2 = lb_build_expr(p, ce.Args[2])
	}
	elem := base_array_type(arg0.Type)
	is_float := is_type_float(elem)
	is_signed := !is_type_unsigned(elem)
	op_code := LLVMOpcode(0)
	switch builtin_id {
	case BuiltinProc_simd_add, BuiltinProc_simd_sub, BuiltinProc_simd_mul, BuiltinProc_simd_div, BuiltinProc_simd_rem:
		if is_float {
			switch builtin_id {
			case BuiltinProc_simd_add:
				op_code = LLVMFAdd
			case BuiltinProc_simd_sub:
				op_code = LLVMFSub
			case BuiltinProc_simd_mul:
				op_code = LLVMFMul
			case BuiltinProc_simd_div:
				op_code = LLVMFDiv
			}
		} else {
			switch builtin_id {
			case BuiltinProc_simd_add:
				op_code = LLVMAdd
			case BuiltinProc_simd_sub:
				op_code = LLVMSub
			case BuiltinProc_simd_mul:
				op_code = LLVMMul
			case BuiltinProc_simd_div:
				if is_signed {
					op_code = LLVMSDiv
				} else {
					op_code = LLVMUDiv
				}
			case BuiltinProc_simd_rem:
				if is_signed {
					op_code = LLVMSRem
				} else {
					op_code = LLVMURem
				}
			}
		}
		if op_code != 0 {
			res.Value = LLVMBuildBinOp(p.Builder, op_code, arg0.Value, arg1.Value, "")
			return res
		}
	case BuiltinProc_simd_shl, BuiltinProc_simd_shr, BuiltinProc_simd_shl_masked, BuiltinProc_simd_shr_masked:
		sz := type_size_of(elem)
		if arg0.Type.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "arg0.type->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1514), 0)
		}
		count := arg0.Type.SimdVector.Count
		elem1 := base_array_type(arg1.Type)
		is_masked := false
		switch builtin_id {
		case BuiltinProc_simd_shl:
			op_code = LLVMShl
			is_masked = false
		case BuiltinProc_simd_shr:
			if is_signed {
				op_code = LLVMAShr
			} else {
				op_code = LLVMLShr
			}
			is_masked = false
		case BuiltinProc_simd_shl_masked:
			op_code = LLVMShl
			is_masked = true
		case BuiltinProc_simd_shr_masked:
			if is_signed {
				op_code = LLVMAShr
			} else {
				op_code = LLVMLShr
			}
			is_masked = true
		}
		if op_code != 0 {
			bits := llvm_splat_int(count, lb_type(m, elem1), sz*8-1)
			if is_masked {
				shift := LLVMBuildAnd(p.Builder, arg1.Value, bits, "")
				res.Value = LLVMBuildBinOp(p.Builder, op_code, arg0.Value, shift, "")
			} else {
				zero := lb_const_nil(m, arg1.Type).Value
				mask := LLVMBuildICmp(p.Builder, LLVMIntULE, arg1.Value, bits, "")
				shift := LLVMBuildBinOp(p.Builder, op_code, arg0.Value, arg1.Value, "")
				res.Value = LLVMBuildSelect(p.Builder, mask, shift, zero, "")
			}
			return res
		}
	case BuiltinProc_simd_bit_and, BuiltinProc_simd_bit_or, BuiltinProc_simd_bit_xor, BuiltinProc_simd_bit_and_not:
		switch builtin_id {
		case BuiltinProc_simd_bit_and:
			op_code = LLVMAnd
		case BuiltinProc_simd_bit_or:
			op_code = LLVMOr
		case BuiltinProc_simd_bit_xor:
			op_code = LLVMXor
		case BuiltinProc_simd_bit_and_not:
			op_code = LLVMAnd
			arg1.Value = LLVMBuildNot(p.Builder, arg1.Value, "")
		}
		if op_code != 0 {
			res.Value = LLVMBuildBinOp(p.Builder, op_code, arg0.Value, arg1.Value, "")
			return res
		}
	case BuiltinProc_simd_neg:
		if is_float {
			res.Value = LLVMBuildFNeg(p.Builder, arg0.Value, "")
		} else {
			res.Value = LLVMBuildNeg(p.Builder, arg0.Value, "")
		}
		return res
	case BuiltinProc_simd_abs:
		if is_float {
			pos := arg0.Value
			neg := LLVMBuildFNeg(p.Builder, pos, "")
			cond := LLVMBuildFCmp(p.Builder, LLVMRealOGT, pos, neg, "")
			res.Value = LLVMBuildSelect(p.Builder, cond, pos, neg, "")
		} else {
			pos := arg0.Value
			neg := LLVMBuildNeg(p.Builder, pos, "")
			var cond LLVMValueRef
			if is_signed {
				cond = LLVMBuildICmp(p.Builder, LLVMIntSGT, pos, neg, "")
			} else {
				cond = LLVMBuildICmp(p.Builder, LLVMIntUGT, pos, neg, "")
			}
			res.Value = LLVMBuildSelect(p.Builder, cond, pos, neg, "")
		}
		return res
	case BuiltinProc_simd_min:
		if is_float {
			return lb_emit_min(p, res.Type, arg0, arg1)
		} else {
			var cond LLVMValueRef
			if is_signed {
				cond = LLVMBuildICmp(p.Builder, LLVMIntSLT, arg0.Value, arg1.Value, "")
			} else {
				cond = LLVMBuildICmp(p.Builder, LLVMIntULT, arg0.Value, arg1.Value, "")
			}
			res.Value = LLVMBuildSelect(p.Builder, cond, arg0.Value, arg1.Value, "")
		}
		return res
	case BuiltinProc_simd_max:
		if is_float {
			return lb_emit_max(p, res.Type, arg0, arg1)
		} else {
			var cond LLVMValueRef
			if is_signed {
				cond = LLVMBuildICmp(p.Builder, LLVMIntSGT, arg0.Value, arg1.Value, "")
			} else {
				cond = LLVMBuildICmp(p.Builder, LLVMIntUGT, arg0.Value, arg1.Value, "")
			}
			res.Value = LLVMBuildSelect(p.Builder, cond, arg0.Value, arg1.Value, "")
		}
		return res
	case BuiltinProc_simd_lanes_eq, BuiltinProc_simd_lanes_ne, BuiltinProc_simd_lanes_lt, BuiltinProc_simd_lanes_le, BuiltinProc_simd_lanes_gt, BuiltinProc_simd_lanes_ge:
		if is_float {
			pred := LLVMRealPredicate(0)
			switch builtin_id {
			case BuiltinProc_simd_lanes_eq:
				pred = LLVMRealOEQ
			case BuiltinProc_simd_lanes_ne:
				pred = LLVMRealUNE
			case BuiltinProc_simd_lanes_lt:
				pred = LLVMRealOLT
			case BuiltinProc_simd_lanes_le:
				pred = LLVMRealOLE
			case BuiltinProc_simd_lanes_gt:
				pred = LLVMRealOGT
			case BuiltinProc_simd_lanes_ge:
				pred = LLVMRealOGE
			}
			if pred != 0 {
				res.Value = LLVMBuildFCmp(p.Builder, pred, arg0.Value, arg1.Value, "")
				res.Value = LLVMBuildSExtOrBitCast(p.Builder, res.Value, lb_type(m, tv.Type), "")
				return res
			}
		} else {
			pred := LLVMIntPredicate(0)
			switch builtin_id {
			case BuiltinProc_simd_lanes_eq:
				pred = LLVMIntEQ
			case BuiltinProc_simd_lanes_ne:
				pred = LLVMIntNE
			case BuiltinProc_simd_lanes_lt:
				if is_signed {
					pred = LLVMIntSLT
				} else {
					pred = LLVMIntULT
				}
			case BuiltinProc_simd_lanes_le:
				if is_signed {
					pred = LLVMIntSLE
				} else {
					pred = LLVMIntULE
				}
			case BuiltinProc_simd_lanes_gt:
				if is_signed {
					pred = LLVMIntSGT
				} else {
					pred = LLVMIntUGT
				}
			case BuiltinProc_simd_lanes_ge:
				if is_signed {
					pred = LLVMIntSGE
				} else {
					pred = LLVMIntUGE
				}
			}
			if pred != 0 {
				res.Value = LLVMBuildICmp(p.Builder, pred, arg0.Value, arg1.Value, "")
				res.Value = LLVMBuildSExtOrBitCast(p.Builder, res.Value, lb_type(m, tv.Type), "")
				return res
			}
		}
	case BuiltinProc_simd_extract:
		res.Value = LLVMBuildExtractElement(p.Builder, arg0.Value, arg1.Value, "")
		return res
	case BuiltinProc_simd_replace:
		res.Value = LLVMBuildInsertElement(p.Builder, arg0.Value, arg2.Value, arg1.Value, "")
		return res
	case BuiltinProc_simd_reduce_add_bisect, BuiltinProc_simd_reduce_mul_bisect:
		if arg0.Type.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "arg0.type->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1646), 0)
		}
		num_elems := arg0.Type.SimdVector.Count
		indices := make([]LLVMValueRef, num_elems)
		for i := i64(0); i < num_elems; i++ {
			indices[i] = lb_const_int(m, t_uint, uint64(i)).Value
		}
		switch builtin_id {
		case BuiltinProc_simd_reduce_add_bisect:
			if is_float {
				op_code = LLVMFAdd
			} else {
				op_code = LLVMAdd
			}
		case BuiltinProc_simd_reduce_mul_bisect:
			if is_float {
				op_code = LLVMFMul
			} else {
				op_code = LLVMMul
			}
		}
		remaining := arg0.Value
		num_remaining := num_elems
		for num_remaining > 1 {
			num_remaining /= 2
			left_indices := LLVMConstVector(indices[0:num_remaining], uint(num_remaining))
			left_value := LLVMBuildShuffleVector(p.Builder, remaining, remaining, left_indices, "")
			right_indices := LLVMConstVector(indices[num_remaining:2*num_remaining], uint(num_remaining))
			right_value := LLVMBuildShuffleVector(p.Builder, remaining, remaining, right_indices, "")
			remaining = LLVMBuildBinOp(p.Builder, op_code, left_value, right_value, "")
		}
		res.Value = LLVMBuildExtractElement(p.Builder, remaining, indices[0], "")
		return res
	case BuiltinProc_simd_reduce_add_ordered, BuiltinProc_simd_reduce_mul_ordered:
		llvm_elem := lb_type(m, elem)
		var args [2]LLVMValueRef
		args_count := 0
		var name string
		switch builtin_id {
		case BuiltinProc_simd_reduce_add_ordered:
			if is_float {
				name = "llvm.vector.reduce.fadd"
				args[args_count] = LLVMConstReal(llvm_elem, 0.0)
				args_count++
			} else {
				name = "llvm.vector.reduce.add"
			}
		case BuiltinProc_simd_reduce_mul_ordered:
			if is_float {
				name = "llvm.vector.reduce.fmul"
				args[args_count] = LLVMConstReal(llvm_elem, 1.0)
				args_count++
			} else {
				name = "llvm.vector.reduce.mul"
			}
		}
		args[args_count] = arg0.Value
		args_count++
		types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(args_count), types[:], uint(len(types)))
		return res
	case BuiltinProc_simd_reduce_add_pairs, BuiltinProc_simd_reduce_mul_pairs:
		if arg0.Type.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "arg0.type->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1711), 0)
		}
		num_elems := arg0.Type.SimdVector.Count
		indices := make([]LLVMValueRef, num_elems)
		for i := i64(0); i < num_elems/2; i++ {
			indices[i] = lb_const_int(m, t_uint, uint64(2*i)).Value
			indices[i+num_elems/2] = lb_const_int(m, t_uint, uint64(2*i+1)).Value
		}
		switch builtin_id {
		case BuiltinProc_simd_reduce_add_pairs:
			if is_float {
				op_code = LLVMFAdd
			} else {
				op_code = LLVMAdd
			}
		case BuiltinProc_simd_reduce_mul_pairs:
			if is_float {
				op_code = LLVMFMul
			} else {
				op_code = LLVMMul
			}
		}
		remaining := arg0.Value
		num_remaining := num_elems
		for num_remaining > 1 {
			num_remaining /= 2
			left_indices := LLVMConstVector(indices[0:num_remaining], uint(num_remaining))
			left_value := LLVMBuildShuffleVector(p.Builder, remaining, remaining, left_indices, "")
			right_indices := LLVMConstVector(indices[num_elems/2:num_elems/2+num_remaining], uint(num_remaining))
			right_value := LLVMBuildShuffleVector(p.Builder, remaining, remaining, right_indices, "")
			remaining = LLVMBuildBinOp(p.Builder, op_code, left_value, right_value, "")
		}
		res.Value = LLVMBuildExtractElement(p.Builder, remaining, indices[0], "")
		return res
	case BuiltinProc_simd_reduce_min, BuiltinProc_simd_reduce_max, BuiltinProc_simd_reduce_and, BuiltinProc_simd_reduce_or, BuiltinProc_simd_reduce_xor:
		var name string
		switch builtin_id {
		case BuiltinProc_simd_reduce_min:
			if is_float {
				name = "llvm.vector.reduce.fmin"
			} else if is_signed {
				name = "llvm.vector.reduce.smin"
			} else {
				name = "llvm.vector.reduce.umin"
			}
		case BuiltinProc_simd_reduce_max:
			if is_float {
				name = "llvm.vector.reduce.fmax"
			} else if is_signed {
				name = "llvm.vector.reduce.smax"
			} else {
				name = "llvm.vector.reduce.umax"
			}
		case BuiltinProc_simd_reduce_and:
			name = "llvm.vector.reduce.and"
		case BuiltinProc_simd_reduce_or:
			name = "llvm.vector.reduce.or"
		case BuiltinProc_simd_reduce_xor:
			name = "llvm.vector.reduce.xor"
		}
		types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
		args := [1]LLVMValueRef{arg0.Value}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
		return res
	case BuiltinProc_simd_reduce_any, BuiltinProc_simd_reduce_all:
		var name string
		switch builtin_id {
		case BuiltinProc_simd_reduce_any:
			name = "llvm.vector.reduce.or"
		case BuiltinProc_simd_reduce_all:
			name = "llvm.vector.reduce.and"
		}
		types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
		args := [1]LLVMValueRef{arg0.Value}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
		return res
	case BuiltinProc_simd_extract_lsbs, BuiltinProc_simd_extract_msbs:
		vt := base_type(arg0.Type)
		if vt.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "vt->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1799), 0)
		}
		elem_bits := 8 * type_size_of(elem)
		num_elems := get_array_type_count(vt)
		broadcast_value := arg0.Value
		if builtin_id == BuiltinProc_simd_extract_msbs {
			word_type := lb_type(m, elem)
			shift_value := llvm_splat_int(num_elems, word_type, elem_bits-1)
			broadcast_value = LLVMBuildAShr(p.Builder, broadcast_value, shift_value, "")
		}
		bitvec_type := LLVMVectorType(LLVMInt1TypeInContext(m.Ctx), uint(num_elems))
		bitvec_value := LLVMBuildTrunc(p.Builder, broadcast_value, bitvec_type, "")
		mask_type := LLVMIntTypeInContext(m.Ctx, uint(num_elems))
		mask_value := LLVMBuildBitCast(p.Builder, bitvec_value, mask_type, "")
		result_type := lb_type(m, res.Type)
		res.Value = LLVMBuildZExtOrBitCast(p.Builder, mask_value, result_type, "")
		return res
	case BuiltinProc_simd_shuffle:
		vt := base_type(arg0.Type)
		if vt.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "vt->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1827), 0)
		}
		indices_count := int64(ce.Args.Count - 2)
		max_count := vt.SimdVector.Count * 2
		if !(indices_count <= max_count) {
			gb_assert_handler("Assertion Failure", "indices_count <= max_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1831), 0)
		}
		values := make([]LLVMValueRef, indices_count)
		for i := isize(0); i < isize(indices_count); i++ {
			idx := lb_build_expr(p, ce.Args[i+2])
			if LLVMIsConstant(idx.Value) == 0 {
				gb_assert_handler("Assertion Failure", "LLVMIsConstant(idx.value)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1836), 0)
			}
			values[i] = idx.Value
		}
		indices := LLVMConstVector(values, uint(indices_count))
		res.Value = LLVMBuildShuffleVector(p.Builder, arg0.Value, arg1.Value, indices, "")
		return res
	case BuiltinProc_simd_odd_even:
		vt := base_type(arg0.Type)
		if vt.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "vt->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1848), 0)
		}
		indices_count := uint64(vt.SimdVector.Count)
		vals := make([]LLVMValueRef, indices_count)
		for i := uint64(0); i < indices_count/2; i++ {
			val := 2*i + 1
			vals[i] = LLVMConstInt(lb_type(p.Module, t_u32), val, false)
		}
		for i := uint64(0); i < indices_count/2; i++ {
			val := 2*i + indices_count
			vals[i+indices_count/2] = LLVMConstInt(lb_type(p.Module, t_u32), val, false)
		}
		indices := LLVMConstVector(vals, uint(indices_count))
		res.Value = LLVMBuildShuffleVector(p.Builder, arg0.Value, arg1.Value, indices, "")
		return res
	case BuiltinProc_simd_select:
		cond := arg0.Value
		x := lb_build_expr(p, ce.Args[1]).Value
		y := lb_build_expr(p, ce.Args[2]).Value
		cond = LLVMBuildICmp(p.Builder, LLVMIntNE, cond, LLVMConstNull(LLVMTypeOf(cond)), "")
		res.Value = LLVMBuildSelect(p.Builder, cond, x, y, "")
		return res
	case BuiltinProc_simd_runtime_swizzle:
		src := arg0.Value
		indices := lb_build_expr(p, ce.Args[1]).Value
		vt := base_type(arg0.Type)
		if vt.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "vt->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(1886), 0)
		}
		count := vt.SimdVector.Count
		elem_type := vt.SimdVector.Elem
		elem_size := type_size_of(elem_type)
		var intrinsic_name string
		use_hardware_runtime_swizzle := false
		if elem_size == 1 {
			use_hardware_runtime_swizzle = true
			if buildContext.Metrics.Arch == TargetArchAmd64 || buildContext.Metrics.Arch == TargetArchI386 {
				switch count {
				case 16:
					intrinsic_name = "llvm.x86.ssse3.pshuf.b.128"
				case 32:
					intrinsic_name = "llvm.x86.avx2.pshuf.b"
				case 64:
					intrinsic_name = "llvm.x86.avx512.pshuf.b.512"
				default:
					use_hardware_runtime_swizzle = false
				}
			} else if buildContext.Metrics.Arch == TargetArchArm64 {
				switch count {
				case 16:
					intrinsic_name = "llvm.aarch64.neon.tbl1"
				case 32:
					intrinsic_name = "llvm.aarch64.neon.tbl2"
				case 48:
					intrinsic_name = "llvm.aarch64.neon.tbl3"
				case 64:
					intrinsic_name = "llvm.aarch64.neon.tbl4"
				default:
					use_hardware_runtime_swizzle = false
				}
			} else if buildContext.Metrics.Arch == TargetArchArm32 {
				switch count {
				case 8:
					intrinsic_name = "llvm.arm.neon.vtbl1"
				case 16:
					intrinsic_name = "llvm.arm.neon.vtbl2"
				case 24:
					intrinsic_name = "llvm.arm.neon.vtbl3"
				case 32:
					intrinsic_name = "llvm.arm.neon.vtbl4"
				default:
					use_hardware_runtime_swizzle = false
				}
			} else if buildContext.Metrics.Arch == TargetArchWasm32 || buildContext.Metrics.Arch == TargetArchWasm64p32 {
				if count == 16 {
					intrinsic_name = "llvm.wasm.swizzle"
				} else {
					use_hardware_runtime_swizzle = false
				}
			} else {
				use_hardware_runtime_swizzle = false
			}
		}
		if use_hardware_runtime_swizzle && intrinsic_name != "" {
			features_enabled := true
			if buildContext.Metrics.Arch == TargetArchAmd64 || buildContext.Metrics.Arch == TargetArchI386 {
				if count == 16 {
					if !check_target_feature_is_enabled(make_string_c("ssse3"), nil) {
						features_enabled = false
					}
				} else if count == 32 {
					if !check_target_feature_is_enabled(make_string_c("ssse3"), nil) ||
						!check_target_feature_is_enabled(make_string_c("avx2"), nil) {
						features_enabled = false
					}
				} else if count == 64 {
					if !check_target_feature_is_enabled(make_string_c("ssse3"), nil) ||
						!check_target_feature_is_enabled(make_string_c("avx2"), nil) ||
						!check_target_feature_is_enabled(make_string_c("avx512f"), nil) ||
						!check_target_feature_is_enabled(make_string_c("avx512bw"), nil) {
						features_enabled = false
					}
				}
			} else if buildContext.Metrics.Arch == TargetArchArm64 || buildContext.Metrics.Arch == TargetArchArm32 {
				if !check_target_feature_is_enabled(make_string_c("neon"), nil) {
					features_enabled = false
				}
			}
			if features_enabled {
				if buildContext.Metrics.Arch == TargetArchAmd64 || buildContext.Metrics.Arch == TargetArchI386 {
					if count == 16 {
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), make_string_c("+ssse3"))
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), make_string_c("128"))
					} else if count == 32 {
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), make_string_c("+avx,+avx2,+ssse3"))
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), make_string_c("256"))
					} else if count == 64 {
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), make_string_c("+avx,+avx2,+avx512f,+avx512bw,+ssse3"))
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), make_string_c("512"))
					}
				} else if buildContext.Metrics.Arch == TargetArchArm64 {
					lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), make_string_c("+neon"))
					if count >= 32 {
						lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), make_string_c("256"))
					}
				} else if buildContext.Metrics.Arch == TargetArchArm32 {
					lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), make_string_c("+neon"))
				}
				if buildContext.Metrics.Arch == TargetArchArm64 && count > 16 {
					num_tables := int(count / 16)
					if count%16 != 0 {
						gb_assert_handler("Assertion Failure", "count % 16 == 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2030), "ARM64 src size must be multiple of 16 bytes, got %lld bytes", count)
					}
					if !(num_tables <= 4) {
						gb_assert_handler("Assertion Failure", "num_tables <= 4", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2031), "ARM64 NEON supports maximum 4 tables (tbl4), got %d tables for %lld-byte vector", num_tables, count)
					}
					var src_parts [4]LLVMValueRef
					for i := 0; i < num_tables; i++ {
						var indices_for_extract [16]LLVMValueRef
						for j := 0; j < 16; j++ {
							indices_for_extract[j] = LLVMConstInt(LLVMInt32TypeInContext(p.Module.Ctx), uint64(i*16+j), false)
						}
						extract_mask := LLVMConstVector(indices_for_extract[:], 16)
						src_parts[i] = LLVMBuildShuffleVector(p.Builder, src, LLVMGetUndef(LLVMTypeOf(src)), extract_mask, "")
					}
					if count == 32 {
						args := [3]LLVMValueRef{src_parts[0], src_parts[1], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 3, nil, 0)
					} else if count == 48 {
						args := [4]LLVMValueRef{src_parts[0], src_parts[1], src_parts[2], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 4, nil, 0)
					} else if count == 64 {
						args := [5]LLVMValueRef{src_parts[0], src_parts[1], src_parts[2], src_parts[3], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 5, nil, 0)
					}
				} else if buildContext.Metrics.Arch == TargetArchArm32 && count > 8 {
					num_tables := int(count / 8)
					if count%8 != 0 {
						gb_assert_handler("Assertion Failure", "count % 8 == 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2058), "ARM32 src size must be multiple of 8 bytes, got %lld bytes", count)
					}
					if !(num_tables <= 4) {
						gb_assert_handler("Assertion Failure", "num_tables <= 4", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2059), "ARM32 NEON supports maximum 4 tables (vtbl4), got %d tables for %lld-byte vector", num_tables, count)
					}
					var src_parts [4]LLVMValueRef
					for i := 0; i < num_tables; i++ {
						var indices_for_extract [8]LLVMValueRef
						for j := 0; j < 8; j++ {
							indices_for_extract[j] = LLVMConstInt(LLVMInt32TypeInContext(p.Module.Ctx), uint64(i*8+j), false)
						}
						extract_mask := LLVMConstVector(indices_for_extract[:], 8)
						src_parts[i] = LLVMBuildShuffleVector(p.Builder, src, LLVMGetUndef(LLVMTypeOf(src)), extract_mask, "")
					}
					if count == 16 {
						args := [3]LLVMValueRef{src_parts[0], src_parts[1], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 3, nil, 0)
					} else if count == 24 {
						args := [4]LLVMValueRef{src_parts[0], src_parts[1], src_parts[2], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 4, nil, 0)
					} else if count == 32 {
						args := [5]LLVMValueRef{src_parts[0], src_parts[1], src_parts[2], src_parts[3], indices}
						res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], 5, nil, 0)
					}
				} else {
					args := [2]LLVMValueRef{src, indices}
					res.Value = lb_call_intrinsic(p, intrinsic_name, args[:], uint(len(args)), nil, 0)
				}
				return res
			} else {
				use_hardware_runtime_swizzle = false
			}
		}
		if !(count > 0 && count <= 64) {
			gb_assert_handler("Assertion Failure", "count > 0 && count <= 64", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2096), 0)
		}
		values := make([]LLVMValueRef, count)
		i32_type := LLVMInt32TypeInContext(p.Module.Ctx)
		elem_llvm_type := lb_type(p.Module, elem_type)
		max_index := count - 1
		index_mask := LLVMConstInt(elem_llvm_type, uint64(max_index), false)
		for i := i64(0); i < count; i++ {
			idx_i := LLVMConstInt(i32_type, uint64(i), false)
			index_elem := LLVMBuildExtractElement(p.Builder, indices, idx_i, "")
			masked_index := LLVMBuildAnd(p.Builder, index_elem, index_mask, "")
			var index_i32 LLVMValueRef
			if LLVMGetIntTypeWidth(LLVMTypeOf(masked_index)) < 32 {
				index_i32 = LLVMBuildZExt(p.Builder, masked_index, i32_type, "")
			} else if LLVMGetIntTypeWidth(LLVMTypeOf(masked_index)) > 32 {
				index_i32 = LLVMBuildTrunc(p.Builder, masked_index, i32_type, "")
			} else {
				index_i32 = masked_index
			}
			values[i] = LLVMBuildExtractElement(p.Builder, src, index_i32, "")
		}
		res.Value = LLVMGetUndef(LLVMTypeOf(src))
		for i := i64(0); i < count; i++ {
			idx_i := LLVMConstInt(i32_type, uint64(i), false)
			res.Value = LLVMBuildInsertElement(p.Builder, res.Value, values[i], idx_i, "")
		}
		return res
	case BuiltinProc_simd_sums_of_n:
		vt := base_type(arg0.Type)
		if vt.Kind != Type_SimdVector {
			gb_assert_handler("Assertion Failure", "vt->kind == Type_SimdVector", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2155), 0)
		}
		is_float := is_type_float(vt.SimdVector.Elem)
		llvm_elem := lb_type(m, elem)
		val := arg0.Value
		llvm_u32 := lb_type(m, t_u32)
		max_count := uint64(vt.SimdVector.Count)
		if ce.Args[1].TAV.Mode != Addressing_Constant {
			gb_assert_handler("Assertion Failure", "ce->args[1]->tav.mode == Addressing_Constant", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2164), 0)
		}
		n := exact_value_to_u64(ce.Args[1].TAV.Value)
		if !(max_count >= n) {
			gb_assert_handler("Assertion Failure", "max_count >= n", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2166), 0)
		}
		if !(max_count%n == 0) {
			gb_assert_handler("Assertion Failure", "max_count %% n == 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2167), 0)
		}
		new_size := max_count / n
		if max_count == n {
			var args [2]LLVMValueRef
			args_count := 0
			var name string
			if is_float {
				name = "llvm.vector.reduce.fadd"
				args[args_count] = LLVMConstReal(llvm_elem, 0.0)
				args_count++
			} else {
				name = "llvm.vector.reduce.add"
			}
			args[args_count] = arg0.Value
			args_count++
			types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
			res.Value = lb_call_intrinsic(p, name, args[:], uint(args_count), types[:], uint(len(types)))
			return res
		} else if n == 2 {
			left_vals := make([]LLVMValueRef, new_size)
			right_vals := make([]LLVMValueRef, new_size)
			for i := uint64(0); i < new_size; i++ {
				left_vals[i] = LLVMConstInt(llvm_u32, 2*i, false)
				right_vals[i] = LLVMConstInt(llvm_u32, 2*i+1, false)
			}
			left_indices := LLVMConstVector(left_vals, uint(new_size))
			right_indices := LLVMConstVector(right_vals, uint(new_size))
			left := LLVMBuildShuffleVector(p.Builder, val, val, left_indices, "")
			right := LLVMBuildShuffleVector(p.Builder, val, val, right_indices, "")
			if is_float {
				res.Value = LLVMBuildFAdd(p.Builder, left, right, "")
			} else {
				res.Value = LLVMBuildAdd(p.Builder, left, right, "")
			}
		} else {
			shuffled := make([]LLVMValueRef, new_size)
			reductions := make([]LLVMValueRef, new_size)
			for i := uint64(0); i < new_size; i++ {
				offset := i * n
				index_vals := make([]LLVMValueRef, n)
				for j := uint64(0); j < n; j++ {
					index_vals[j] = LLVMConstInt(llvm_u32, offset+j, false)
				}
				indices := LLVMConstVector(index_vals, uint(n))
				shuffled[i] = LLVMBuildShuffleVector(p.Builder, val, val, indices, "")
			}
			for i := uint64(0); i < new_size; i++ {
				var args [2]LLVMValueRef
				args_count := 0
				var name string
				if is_float {
					name = "llvm.vector.reduce.fadd"
					args[args_count] = LLVMConstReal(llvm_elem, 0.0)
					args_count++
				} else {
					name = "llvm.vector.reduce.add"
				}
				args[args_count] = shuffled[i]
				args_count++
				this_simd_type := LLVMVectorType(llvm_elem, uint(n))
				types := [1]LLVMTypeRef{this_simd_type}
				reductions[i] = lb_call_intrinsic(p, name, args[:], uint(args_count), types[:], uint(len(types)))
			}
			res.Value = LLVMConstNull(LLVMVectorType(llvm_elem, uint(new_size)))
			for i := uint64(0); i < new_size; i++ {
				idx := LLVMConstInt(llvm_u32, i, false)
				res.Value = LLVMBuildInsertElement(p.Builder, res.Value, reductions[i], idx, "")
			}
		}
		return res
	case BuiltinProc_simd_pairwise_add, BuiltinProc_simd_pairwise_sub:
		if is_float {
			switch builtin_id {
			case BuiltinProc_simd_pairwise_add:
				op_code = LLVMFAdd
			case BuiltinProc_simd_pairwise_sub:
				op_code = LLVMFSub
			}
		} else {
			switch builtin_id {
			case BuiltinProc_simd_pairwise_add:
				op_code = LLVMAdd
			case BuiltinProc_simd_pairwise_sub:
				op_code = LLVMSub
			}
		}
		if op_code != 0 {
			a := arg0.Value
			b := arg1.Value
			count := LLVMGetVectorSize(LLVMTypeOf(a))
			evens := make([]LLVMValueRef, count)
			odds := make([]LLVMValueRef, count)
			llvm_u32 := lb_type(m, t_u32)
			for i := uint(0); i < count; i++ {
				evens[i] = LLVMConstInt(llvm_u32, uint64(2*i), false)
				odds[i] = LLVMConstInt(llvm_u32, uint64(2*i+1), false)
			}
			x := LLVMBuildShuffleVector(p.Builder, a, b, LLVMConstVector(evens, count), "")
			y := LLVMBuildShuffleVector(p.Builder, a, b, LLVMConstVector(odds, count), "")
			res.Value = LLVMBuildBinOp(p.Builder, op_code, x, y, "")
			return res
		}
	case BuiltinProc_simd_ceil, BuiltinProc_simd_floor, BuiltinProc_simd_trunc, BuiltinProc_simd_nearest:
		var name string
		switch builtin_id {
		case BuiltinProc_simd_ceil:
			name = "llvm.ceil"
		case BuiltinProc_simd_floor:
			name = "llvm.floor"
		case BuiltinProc_simd_trunc:
			name = "llvm.trunc"
		case BuiltinProc_simd_nearest:
			name = "llvm.nearbyint"
		}
		types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
		args := [1]LLVMValueRef{arg0.Value}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
		return res
	case BuiltinProc_simd_approx_recip:
		vt := base_type(arg0.Type)
		if !is_type_simd_vector(vt) {
			gb_assert_handler("Assertion Failure", "is_type_simd_vector(vt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2309), 0)
		}
		val := arg0.Value
		count := vt.SimdVector.Count
		elem := base_type(vt.SimdVector.Elem)
		var intrinsic_name string
		var default_width uint
		var extra_args []LLVMValueRef
		extra_args_count := uint(0)
		var features_to_enable String
		var min_legal_vector_width String
		if buildContext.Metrics.Arch == TargetArchI386 || buildContext.Metrics.Arch == TargetArchAmd64 {
			if are_types_identical(elem, t_f32) {
				intrinsic_name = "llvm.x86.sse.rcp.ps"
				default_width = 4
				if count >= 8 {
					if check_target_feature_is_enabled(make_string_c("avx"), nil) {
						intrinsic_name = "llvm.x86.avx.rcp.ps.256"
						default_width = 8
					}
				}
				if count >= 16 {
					if check_target_feature_is_enabled(make_string_c("avx512vl"), nil) {
						features_to_enable = make_string_c("+avx512f,+evex512")
						min_legal_vector_width = make_string_c("512")
						intrinsic_name = "llvm.x86.avx512.rcp14.ps.512"
						default_width = 16
						extra_args_count = 2
						extra_args = make([]LLVMValueRef, 2)
						extra_args[0] = LLVMGetUndef(LLVMVectorType(lb_type(p.Module, t_f32), 16))
						extra_args[1] = LLVMConstInt(lb_type(p.Module, t_i16), uint64((1<<16)-1), true)
					}
				}
			}
		}
		if lb_llvm_simd_bulk_op_unary(p, arg0, &res.Value, intrinsic_name, default_width, extra_args, extra_args_count) {
			if features_to_enable.Len > 0 {
				lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), features_to_enable)
			}
			if min_legal_vector_width.Len > 0 {
				lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), min_legal_vector_width)
			}
			return res
		}
		one := lb_const_value(p.Module, vt.SimdVector.Elem, exact_value_float(1)).Value
		ones := make([]LLVMValueRef, count)
		for i := i64(0); i < count; i++ {
			ones[i] = one
		}
		one_vector := LLVMConstVector(ones, uint(count))
		res.Value = LLVMBuildFDiv(p.Builder, one_vector, val, "")
		return res
	case BuiltinProc_simd_approx_recip_sqrt:
		const_broadcast_vector_float := func(p *lbProcedure, elem_type *Type, count i64, elem float64) LLVMValueRef {
			val := lb_const_value(p.Module, elem_type, exact_value_float(elem)).Value
			return llvm_vector_broadcast(p, val, uint(count))
		}
		const_broadcast_vector_unsigned := func(p *lbProcedure, elem_type *Type, count i64, elem uint64) LLVMValueRef {
			val := lb_const_value(p.Module, elem_type, exact_value_u64(elem)).Value
			return llvm_vector_broadcast(p, val, uint(count))
		}
		vt := base_type(arg0.Type)
		if !is_type_simd_vector(vt) {
			gb_assert_handler("Assertion Failure", "is_type_simd_vector(vt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2386), 0)
		}
		elem := base_type(vt.SimdVector.Elem)
		val := arg0.Value
		count := vt.SimdVector.Count
		var intrinsic_name string
		var default_width uint
		var extra_args []LLVMValueRef
		extra_args_count := uint(0)
		var features_to_enable String
		var min_legal_vector_width String
		if buildContext.Metrics.Arch == TargetArchI386 || buildContext.Metrics.Arch == TargetArchAmd64 {
			if are_types_identical(elem, t_f32) {
				intrinsic_name = "llvm.x86.sse.rsqrt.ps"
				default_width = 4
				if count >= 8 {
					if check_target_feature_is_enabled(make_string_c("avx"), nil) {
						intrinsic_name = "llvm.x86.avx.rsqrt.ps.256"
						default_width = 8
					}
				}
				if count >= 16 {
					if check_target_feature_is_enabled(make_string_c("avx512vl"), nil) {
						features_to_enable = make_string_c("+avx512f,+evex512")
						min_legal_vector_width = make_string_c("512")
						intrinsic_name = "llvm.x86.avx512.rsqrt14.ps.512"
						default_width = 16
						extra_args_count = 2
						extra_args = make([]LLVMValueRef, 2)
						extra_args[0] = LLVMGetUndef(LLVMVectorType(lb_type(p.Module, t_f32), 16))
						extra_args[1] = LLVMConstInt(lb_type(p.Module, t_i16), uint64((1<<16)-1), true)
					}
				}
			}
		}
		if lb_llvm_simd_bulk_op_unary(p, arg0, &res.Value, intrinsic_name, default_width, extra_args, extra_args_count) {
			if features_to_enable.Len > 0 {
				lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("target-features"), features_to_enable)
			}
			if min_legal_vector_width.Len > 0 {
				lb_add_attribute_to_proc_with_string(p.Module, p.Value, make_string_c("min-legal-vector-width"), min_legal_vector_width)
			}
			return res
		}
		if are_types_identical(elem, t_f64) {
			half := const_broadcast_vector_float(p, elem, count, 0.5)
			three_halfs := const_broadcast_vector_float(p, elem, count, 1.5)
			unsigned_one_vector := const_broadcast_vector_unsigned(p, t_u64, count, 1)
			half_val := LLVMBuildFMul(p.Builder, val, half, "")
			magic := const_broadcast_vector_unsigned(p, t_u64, count, 0x5FE6EB50C7B537A9)
			i := LLVMBuildBitCast(p.Builder, val, LLVMTypeOf(magic), "")
			i = LLVMBuildLShr(p.Builder, i, unsigned_one_vector, "")
			guess := LLVMBuildSub(p.Builder, magic, i, "")
			guess = LLVMBuildBitCast(p.Builder, guess, LLVMTypeOf(val), "")
			for iter := isize(0); iter < 1; iter++ {
				half_guess := LLVMBuildFMul(p.Builder, half_val, guess, "")
				half_guess = LLVMBuildFNeg(p.Builder, half_guess, "")
				var nma LLVMValueRef
				{
					name := "llvm.fma"
					types := [1]LLVMTypeRef{lb_type(p.Module, vt)}
					args := [3]LLVMValueRef{half_guess, guess, three_halfs}
					nma = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
				}
				guess = LLVMBuildFMul(p.Builder, guess, nma, "")
			}
			res.Value = guess
			return res
		} else {
			name := "llvm.sqrt"
			types := [1]LLVMTypeRef{lb_type(p.Module, vt)}
			args := [1]LLVMValueRef{val}
			res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
			one_vector := const_broadcast_vector_float(p, elem, count, 1.0)
			res.Value = LLVMBuildFDiv(p.Builder, one_vector, res.Value, "")
			return res
		}
	case BuiltinProc_simd_lanes_reverse:
		count := get_array_type_count(arg0.Type)
		values := make([]LLVMValueRef, count)
		llvm_u32 := lb_type(m, t_u32)
		for i := i64(0); i < count; i++ {
			values[i] = LLVMConstInt(llvm_u32, uint64(count-1-i), false)
		}
		mask := LLVMConstVector(values, uint(count))
		v := arg0.Value
		res.Value = LLVMBuildShuffleVector(p.Builder, v, v, mask, "")
		return res
	case BuiltinProc_simd_lanes_rotate_left, BuiltinProc_simd_lanes_rotate_right:
		count := get_array_type_count(arg0.Type)
		if !is_power_of_two(count) {
			gb_assert_handler("Assertion Failure", "is_power_of_two(count)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2512), 0)
		}
		var bi_count BigInt
		big_int_from_i64(&bi_count, count)
		tv := ce.Args[1].TAV
		val := exact_value_to_integer(tv.Value)
		if val.Kind != ExactValue_Integer {
			gb_assert_handler("Assertion Failure", "val.kind == ExactValue_Integer", "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2518), 0)
		}
		bi := &val.ValueInteger
		if builtin_id == BuiltinProc_simd_lanes_rotate_right {
			big_int_neg(bi, bi)
		}
		big_int_rem(bi, bi, &bi_count)
		big_int_dealloc(&bi_count)
		left := big_int_to_i64(bi)
		values := make([]LLVMValueRef, count)
		llvm_u32 := lb_type(m, t_u32)
		for i := i64(0); i < count; i++ {
			idx := uint64(i+left) & uint64(count-1)
			values[i] = LLVMConstInt(llvm_u32, idx, false)
		}
		mask := LLVMConstVector(values, uint(count))
		v := arg0.Value
		res.Value = LLVMBuildShuffleVector(p.Builder, v, v, mask, "")
		return res
	case BuiltinProc_simd_saturating_add, BuiltinProc_simd_saturating_sub:
		var name string
		switch builtin_id {
		case BuiltinProc_simd_saturating_add:
			if is_signed {
				name = "llvm.sadd.sat"
			} else {
				name = "llvm.uadd.sat"
			}
		case BuiltinProc_simd_saturating_sub:
			if is_signed {
				name = "llvm.ssub.sat"
			} else {
				name = "llvm.usub.sat"
			}
		}
		types := [1]LLVMTypeRef{lb_type(p.Module, arg0.Type)}
		args := [2]LLVMValueRef{arg0.Value, arg1.Value}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
		return res
	case BuiltinProc_simd_clamp:
		v := arg0.Value
		min := arg1.Value
		max := arg2.Value
		if is_float {
			v = LLVMBuildSelect(p.Builder, LLVMBuildFCmp(p.Builder, LLVMRealOLT, v, min, ""), min, v, "")
			res.Value = LLVMBuildSelect(p.Builder, LLVMBuildFCmp(p.Builder, LLVMRealOGT, v, max, ""), max, v, "")
		} else if is_signed {
			v = LLVMBuildSelect(p.Builder, LLVMBuildICmp(p.Builder, LLVMIntSLT, v, min, ""), min, v, "")
			res.Value = LLVMBuildSelect(p.Builder, LLVMBuildICmp(p.Builder, LLVMIntSGT, v, max, ""), max, v, "")
		} else {
			v = LLVMBuildSelect(p.Builder, LLVMBuildICmp(p.Builder, LLVMIntULT, v, min, ""), min, v, "")
			res.Value = LLVMBuildSelect(p.Builder, LLVMBuildICmp(p.Builder, LLVMIntUGT, v, max, ""), max, v, "")
		}
		return res
	case BuiltinProc_simd_to_bits, BuiltinProc_simd_to_bits_signed:
		res.Value = LLVMBuildBitCast(p.Builder, arg0.Value, lb_type(m, tv.Type), "")
		return res
	case BuiltinProc_simd_gather, BuiltinProc_simd_scatter, BuiltinProc_simd_masked_load, BuiltinProc_simd_masked_store, BuiltinProc_simd_masked_expand_load, BuiltinProc_simd_masked_compress_store:
		ptr := arg0.Value
		val := arg1.Value
		mask := arg2.Value
		count := uint(get_array_type_count(arg1.Type))
		mask_type := LLVMVectorType(LLVMInt1TypeInContext(p.Module.Ctx), count)
		mask = LLVMBuildTrunc(p.Builder, mask, mask_type, "")
		var name string
		switch builtin_id {
		case BuiltinProc_simd_gather:
			name = "llvm.masked.gather"
		case BuiltinProc_simd_scatter:
			name = "llvm.masked.scatter"
		case BuiltinProc_simd_masked_load:
			name = "llvm.masked.load"
		case BuiltinProc_simd_masked_store:
			name = "llvm.masked.store"
		case BuiltinProc_simd_masked_expand_load:
			name = "llvm.masked.expandload"
		case BuiltinProc_simd_masked_compress_store:
			name = "llvm.masked.compressstore"
		}
		type_count := uint(2)
		types := [2]LLVMTypeRef{
			lb_type(p.Module, arg1.Type),
			lb_type(p.Module, arg0.Type),
		}
		alignment := uint64(type_align_of(base_array_type(arg1.Type)))
		align := LLVMConstInt(LLVMInt32TypeInContext(p.Module.Ctx), alignment, false)
		align_idx := int32(-1)
		arg_count := uint(4)
		var args [4]LLVMValueRef
		switch builtin_id {
		case BuiltinProc_simd_masked_load:
			types[1] = lb_type(p.Module, t_rawptr)
			args[0] = ptr
			args[1] = align
			args[2] = mask
			args[3] = val
		case BuiltinProc_simd_gather:
			args[0] = ptr
			args[1] = align
			args[2] = mask
			args[3] = val
		case BuiltinProc_simd_masked_store:
			types[1] = lb_type(p.Module, t_rawptr)
			args[0] = val
			args[1] = ptr
			args[2] = align
			args[3] = mask
		case BuiltinProc_simd_scatter:
			args[0] = val
			args[1] = ptr
			args[2] = align
			args[3] = mask
		case BuiltinProc_simd_masked_expand_load:
			arg_count = 3
			type_count = 1
			args[0] = ptr
			args[1] = mask
			args[2] = val
		case BuiltinProc_simd_masked_compress_store:
			arg_count = 3
			type_count = 1
			args[0] = val
			args[1] = ptr
			args[2] = mask
		}
		res.Value = lb_call_intrinsic(p, name, args[:], arg_count, types[:], type_count)
		if align_idx >= 0 {
			align_attr := lb_create_enum_attribute(p.Module.Ctx, "align", alignment)
			LLVMAddAttributeAtIndex(res.Value, uint(align_idx), align_attr)
		}
		return res
	}
	gb_assert_handler("Panic", nil, "G:\\b0pass-win\\Odin\\src\\llvm_backend_proc.cpp", i64(2685), "Unhandled simd intrinsic: '%.*s'", int(len(builtin_procs[builtin_id].Name)), builtin_procs[builtin_id].Name.Text)
	return lbValue{}
}
