package cmd

func lb_emit_unary_arith_not(p *lbProcedure, x lbValue, typ *Type) lbValue {
	res := lbValue{}
	zero := LLVMConstInt(lb_type(p.Module, x.Type), 0, false)
	res.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, x.Value, zero, "")
	res.Type = t_llvm_bool
	return lb_emit_conv(p, res, typ)
}

func lb_compare_records(p *lbProcedure, op_kind TokenKind, left lbValue, right lbValue, typ *Type) lbValue {
	if !((is_type_struct(typ) || is_type_soa_pointer(typ) || is_type_union(typ)) && is_type_comparable(typ)) {
		panic("type must be comparable record")
	}

	left_ptr := lb_address_from_load_or_generate_local(p, left)
	right_ptr := lb_address_from_load_or_generate_local(p, right)
	size := type_size_of(typ)

	res := lbValue{}
	if size == 0 {
		switch op_kind {
		case Token_CmpEq:
			return lb_const_bool(p.Module, t_bool, true)
		case Token_NotEq:
			return lb_const_bool(p.Module, t_bool, false)
		}
		panic("invalid operator")
	}

	if is_type_simple_compare(typ) {
		args := []lbValue{
			lb_emit_conv(p, left_ptr, t_rawptr),
			lb_emit_conv(p, right_ptr, t_rawptr),
			lb_const_int(p.Module, t_int, uint64(type_size_of(typ))),
		}
		res = lb_emit_runtime_call(p, "memory_equal", args)
	} else {
		value := lb_equal_proc_for_type(p.Module, typ)
		args := []lbValue{
			lb_emit_conv(p, left_ptr, t_rawptr),
			lb_emit_conv(p, right_ptr, t_rawptr),
		}
		res = lb_emit_call(p, value, args)
	}

	if op_kind == Token_NotEq {
		res = lb_emit_unary_arith_not(p, res, res.Type)
	}
	return res
}

func lb_emit_comp(p *lbProcedure, op_kind TokenKind, left lbValue, right lbValue) lbValue {
	a := core_type(left.Type)
	b := core_type(right.Type)

	if !((op_kind > Token__ComparisonBegin || true) && (op_kind < Token__ComparisonEnd || true)) {
		// comparison check - handled by caller
	}

	nil_check := lbValue{}

	if is_type_array_like(left.Type) || is_type_array_like(right.Type) {
		// don't do nil check if array-like
	} else if is_type_untyped_nil(left.Type) {
		nil_check = lb_emit_comp_against_nil(p, op_kind, right)
	} else if is_type_untyped_nil(right.Type) {
		nil_check = lb_emit_comp_against_nil(p, op_kind, left)
	}
	if nil_check.Value != 0 {
		return nil_check
	}

	if are_types_identical(a, b) {
		// No conversion needed
	} else if (lb_is_const(left) && !is_type_array(left.Type)) || lb_is_const_nil(left) {
		if lb_is_const_nil(left) {
			if internal_check_is_assignable_to(right.Type, left.Type) {
				right = lb_emit_conv(p, right, left.Type)
			}
			if LLVMTypeOf(left.Value) == LLVMTypeOf(right.Value) {
				return lb_emit_comp_against_nil(p, op_kind, right)
			}
		}
		left = lb_emit_conv(p, left, right.Type)
	} else if (lb_is_const(right) && !is_type_array(right.Type)) || lb_is_const_nil(right) {
		if lb_is_const_nil(right) {
			if internal_check_is_assignable_to(left.Type, right.Type) {
				left = lb_emit_conv(p, left, right.Type)
			}
			if LLVMTypeOf(left.Value) == LLVMTypeOf(right.Value) {
				return lb_emit_comp_against_nil(p, op_kind, left)
			}
		}
		right = lb_emit_conv(p, right, left.Type)
	} else {
		lt := left.Type
		rt := right.Type

		ls := type_size_of(lt)
		rs := type_size_of(rt)

		if ls < rs {
			left = lb_emit_conv(p, left, rt)
		} else if ls > rs {
			right = lb_emit_conv(p, right, lt)
		} else {
			if is_type_union(rt) {
				left = lb_emit_conv(p, left, rt)
			} else {
				right = lb_emit_conv(p, right, lt)
			}
		}
	}

	a = core_type(left.Type)
	b = core_type(right.Type)

	if is_type_matrix(a) && (op_kind == Token_CmpEq || op_kind == Token_NotEq) {
		tl := base_type(a)
		lhs := lb_address_from_load_or_generate_local(p, left)
		rhs := lb_address_from_load_or_generate_local(p, right)

		args := []lbValue{
			lb_emit_conv(p, lhs, t_rawptr),
			lb_emit_conv(p, rhs, t_rawptr),
			lb_const_int(p.Module, t_int, uint64(type_size_of(tl))),
		}
		val := lb_emit_runtime_call(p, "memory_compare", args)
		res := lb_emit_comp(p, op_kind, val, lb_const_nil(p.Module, val.Type))
		return lb_emit_conv(p, res, t_bool)
	}
	if is_type_array_like(a) {
		tl := base_type(a)
		lhs := lb_address_from_load_or_generate_local(p, left)
		rhs := lb_address_from_load_or_generate_local(p, right)

		cmp_op := Token_And
		res := lb_const_bool(p.Module, t_llvm_bool, true)
		if op_kind == Token_NotEq {
			res = lb_const_bool(p.Module, t_llvm_bool, false)
			cmp_op = Token_Or
		} else if op_kind == Token_CmpEq {
			res = lb_const_bool(p.Module, t_llvm_bool, true)
			cmp_op = Token_And
		}

		inline_array_arith := lb_can_try_to_inline_array_arith(tl)
		count := int32(0)
		switch tl.Kind {
		case Type_Array:
			count = int32(tl.Array.Count)
		case Type_EnumeratedArray:
			count = int32(tl.EnumeratedArray.Count)
		}

		if inline_array_arith {
			val := lb_add_local_generated(p, t_bool, false)
			lb_addr_store(p, val, res)
			for i := int32(0); i < count; i++ {
				x := lb_emit_load(p, lb_emit_array_epi(p.Module, lhs, isize(i)))
				y := lb_emit_load(p, lb_emit_array_epi(p.Module, rhs, isize(i)))
				cmp := lb_emit_comp(p, op_kind, x, y)
				new_res := lb_emit_arith(p, cmp_op, lb_addr_load(p, val), cmp, t_bool)
				lb_addr_store(p, val, lb_emit_conv(p, new_res, t_bool))
			}

			return lb_addr_load(p, val)
		} else {
			if is_type_simple_compare(tl) && (op_kind == Token_CmpEq || op_kind == Token_NotEq) {
				args := []lbValue{
					lb_emit_conv(p, lhs, t_rawptr),
					lb_emit_conv(p, rhs, t_rawptr),
					lb_const_int(p.Module, t_int, uint64(type_size_of(tl))),
				}
				val := lb_emit_runtime_call(p, "memory_compare", args)
				res := lb_emit_comp(p, op_kind, val, lb_const_nil(p.Module, val.Type))
				return lb_emit_conv(p, res, t_bool)
			} else {
				val := lb_add_local_generated(p, t_bool, false)
				lb_addr_store(p, val, res)
				loop_data := lb_loop_start(p, isize(count), t_i32)
				{
					i := loop_data.Idx
					x := lb_emit_load(p, lb_emit_array_ep(p, lhs, i))
					y := lb_emit_load(p, lb_emit_array_ep(p, rhs, i))
					cmp := lb_emit_comp(p, op_kind, x, y)
					new_res := lb_emit_arith(p, cmp_op, lb_addr_load(p, val), cmp, t_bool)
					lb_addr_store(p, val, lb_emit_conv(p, new_res, t_bool))
				}
				lb_loop_end(p, loop_data)

				return lb_addr_load(p, val)
			}
		}
	}

	if (is_type_struct(a) || is_type_union(a)) && is_type_comparable(a) {
		return lb_compare_records(p, op_kind, left, right, a)
	}

	if (is_type_struct(b) || is_type_union(b)) && is_type_comparable(b) {
		return lb_compare_records(p, op_kind, left, right, b)
	}

	if is_type_string16(a) || is_type_cstring16(a) {
		if is_type_cstring16(a) && is_type_cstring16(b) {
			left = lb_emit_conv(p, left, t_cstring16)
			right = lb_emit_conv(p, right, t_cstring16)
			runtime_procedure := ""
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "cstring16_eq"
			case Token_NotEq:
				runtime_procedure = "cstring16_ne"
			case Token_Lt:
				runtime_procedure = "cstring16_lt"
			case Token_Gt:
				runtime_procedure = "cstring16_gt"
			case Token_LtEq:
				runtime_procedure = "cstring16_le"
			case Token_GtEq:
				runtime_procedure = "cstring16_ge"
			}
			if runtime_procedure == "" {
				panic("unhandled cstring16 comparison")
			}

			args := []lbValue{left, right}
			return lb_emit_runtime_call(p, runtime_procedure, args)
		}

		if is_type_cstring16(a) != is_type_cstring16(b) {
			left = lb_emit_conv(p, left, t_string16)
			right = lb_emit_conv(p, right, t_string16)
		}

		runtime_procedure := ""
		switch op_kind {
		case Token_CmpEq:
			runtime_procedure = "string16_eq"
		case Token_NotEq:
			runtime_procedure = "string16_ne"
		case Token_Lt:
			runtime_procedure = "string16_lt"
		case Token_Gt:
			runtime_procedure = "string16_gt"
		case Token_LtEq:
			runtime_procedure = "string16_le"
		case Token_GtEq:
			runtime_procedure = "string16_ge"
		}
		if runtime_procedure == "" {
			panic("unhandled string16 comparison")
		}

		args := []lbValue{left, right}
		return lb_emit_runtime_call(p, runtime_procedure, args)
	}

	if is_type_string(a) {
		if is_type_cstring(a) && is_type_cstring(b) {
			left = lb_emit_conv(p, left, t_cstring)
			right = lb_emit_conv(p, right, t_cstring)
			runtime_procedure := ""
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "cstring_eq"
			case Token_NotEq:
				runtime_procedure = "cstring_ne"
			case Token_Lt:
				runtime_procedure = "cstring_lt"
			case Token_Gt:
				runtime_procedure = "cstring_gt"
			case Token_LtEq:
				runtime_procedure = "cstring_le"
			case Token_GtEq:
				runtime_procedure = "cstring_ge"
			}
			if runtime_procedure == "" {
				panic("unhandled cstring comparison")
			}

			args := []lbValue{left, right}
			return lb_emit_runtime_call(p, runtime_procedure, args)
		}

		if is_type_cstring(a) != is_type_cstring(b) {
			left = lb_emit_conv(p, left, t_string)
			right = lb_emit_conv(p, right, t_string)
		}

		runtime_procedure := ""
		switch op_kind {
		case Token_CmpEq:
			runtime_procedure = "string_eq"
		case Token_NotEq:
			runtime_procedure = "string_ne"
		case Token_Lt:
			runtime_procedure = "string_lt"
		case Token_Gt:
			runtime_procedure = "string_gt"
		case Token_LtEq:
			runtime_procedure = "string_le"
		case Token_GtEq:
			runtime_procedure = "string_ge"
		}
		if runtime_procedure == "" {
			panic("unhandled string comparison")
		}

		args := []lbValue{left, right}
		return lb_emit_runtime_call(p, runtime_procedure, args)
	}

	if is_type_complex(a) {
		runtime_procedure := ""
		sz := 8 * type_size_of(a)
		switch sz {
		case 32:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "complex32_eq"
			case Token_NotEq:
				runtime_procedure = "complex32_ne"
			}
		case 64:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "complex64_eq"
			case Token_NotEq:
				runtime_procedure = "complex64_ne"
			}
		case 128:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "complex128_eq"
			case Token_NotEq:
				runtime_procedure = "complex128_ne"
			}
		}
		if runtime_procedure == "" {
			panic("unhandled complex comparison")
		}

		args := []lbValue{left, right}
		return lb_emit_runtime_call(p, runtime_procedure, args)
	}

	if is_type_quaternion(a) {
		runtime_procedure := ""
		sz := 8 * type_size_of(a)
		switch sz {
		case 64:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "quaternion64_eq"
			case Token_NotEq:
				runtime_procedure = "quaternion64_ne"
			}
		case 128:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "quaternion128_eq"
			case Token_NotEq:
				runtime_procedure = "quaternion128_ne"
			}
		case 256:
			switch op_kind {
			case Token_CmpEq:
				runtime_procedure = "quaternion256_eq"
			case Token_NotEq:
				runtime_procedure = "quaternion256_ne"
			}
		}
		if runtime_procedure == "" {
			panic("unhandled quaternion comparison")
		}

		args := []lbValue{left, right}
		return lb_emit_runtime_call(p, runtime_procedure, args)
	}

	if is_type_bit_set(a) {
		switch op_kind {
		case Token_Lt, Token_LtEq, Token_Gt, Token_GtEq:
			it := bit_set_to_int(a)
			lhs := lb_emit_transmute(p, left, it)
			rhs := lb_emit_transmute(p, right, it)
			if is_type_different_to_arch_endianness(it) {
				it = integer_endian_type_to_platform_type(it)
				lhs = lb_emit_byte_swap(p, lhs, it)
				rhs = lb_emit_byte_swap(p, rhs, it)
			}

			res := lb_emit_arith(p, Token_And, lhs, rhs, it)

			if op_kind == Token_Lt || op_kind == Token_LtEq {
				res.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, res.Value, lhs.Value, "")
				res.Type = t_llvm_bool
			} else if op_kind == Token_Gt || op_kind == Token_GtEq {
				res.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, res.Value, rhs.Value, "")
				res.Type = t_llvm_bool
			}

			if op_kind == Token_Lt || op_kind == Token_Gt {
				eq := lbValue{}
				eq.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, lhs.Value, rhs.Value, "")
				eq.Type = t_llvm_bool
				res = lb_emit_arith(p, Token_AndNot, res, eq, t_llvm_bool)
			}

			return res

		case Token_CmpEq, Token_NotEq:
			pred := LLVMIntPredicate(0)
			switch op_kind {
			case Token_CmpEq:
				pred = LLVMIntEQ
			case Token_NotEq:
				pred = LLVMIntNE
			}
			res := lbValue{}
			res.Type = t_llvm_bool
			res.Value = LLVMBuildICmp(p.Builder, pred, left.Value, right.Value, "")
			return res
		}
	}

	if op_kind != Token_CmpEq && op_kind != Token_NotEq {
		t := left.Type
		if is_type_integer(t) && is_type_different_to_arch_endianness(t) {
			platform_type := integer_endian_type_to_platform_type(t)
			x := lb_emit_byte_swap(p, left, platform_type)
			y := lb_emit_byte_swap(p, right, platform_type)
			left = x
			right = y
		} else if is_type_float(t) && is_type_different_to_arch_endianness(t) {
			platform_type := integer_endian_type_to_platform_type(t)
			x := lb_emit_conv(p, left, platform_type)
			y := lb_emit_conv(p, right, platform_type)
			left = x
			right = y
		}
	}

	a = core_type(left.Type)
	b = core_type(right.Type)

	res := lbValue{}
	res.Type = t_llvm_bool
	if is_type_integer(a) ||
		is_type_boolean(a) ||
		is_type_pointer(a) ||
		is_type_multi_pointer(a) ||
		is_type_proc(a) ||
		is_type_enum(a) {
		pred := LLVMIntPredicate(0)

		if is_type_unsigned(left.Type) {
			switch op_kind {
			case Token_Gt:
				pred = LLVMIntUGT
			case Token_GtEq:
				pred = LLVMIntUGE
			case Token_Lt:
				pred = LLVMIntULT
			case Token_LtEq:
				pred = LLVMIntULE
			}
		} else {
			switch op_kind {
			case Token_Gt:
				pred = LLVMIntSGT
			case Token_GtEq:
				pred = LLVMIntSGE
			case Token_Lt:
				pred = LLVMIntSLT
			case Token_LtEq:
				pred = LLVMIntSLE
			}
		}
		switch op_kind {
		case Token_CmpEq:
			pred = LLVMIntEQ
		case Token_NotEq:
			pred = LLVMIntNE
		}
		lhs := left.Value
		rhs := right.Value
		if LLVMTypeOf(lhs) != LLVMTypeOf(rhs) {
			if lb_is_type_kind(LLVMTypeOf(lhs), LLVMPointerTypeKind) {
				rhs = LLVMBuildPointerCast(p.Builder, rhs, LLVMTypeOf(lhs), "")
			}
		}

		if is_type_different_to_arch_endianness(left.Type) {
			pt := integer_endian_type_to_platform_type(left.Type)
			lhs = lb_emit_byte_swap(p, lbValue{Value: lhs, Type: pt}, pt).Value
			rhs = lb_emit_byte_swap(p, lbValue{Value: rhs, Type: pt}, pt).Value
		}

		res.Value = LLVMBuildICmp(p.Builder, pred, lhs, rhs, "")
	} else if is_type_float(a) {
		pred := LLVMRealPredicate(0)
		switch op_kind {
		case Token_CmpEq:
			pred = LLVMRealOEQ
		case Token_Gt:
			pred = LLVMRealOGT
		case Token_GtEq:
			pred = LLVMRealOGE
		case Token_Lt:
			pred = LLVMRealOLT
		case Token_LtEq:
			pred = LLVMRealOLE
		case Token_NotEq:
			pred = LLVMRealUNE
		}

		if is_type_different_to_arch_endianness(left.Type) {
			pt := integer_endian_type_to_platform_type(left.Type)
			left = lb_emit_byte_swap(p, left, pt)
			right = lb_emit_byte_swap(p, right, pt)
		}

		res.Value = LLVMBuildFCmp(p.Builder, pred, left.Value, right.Value, "")
	} else if is_type_typeid(a) {
		pred := LLVMIntPredicate(0)
		switch op_kind {
		case Token_Gt:
			pred = LLVMIntUGT
		case Token_GtEq:
			pred = LLVMIntUGE
		case Token_Lt:
			pred = LLVMIntULT
		case Token_LtEq:
			pred = LLVMIntULE
		case Token_CmpEq:
			pred = LLVMIntEQ
		case Token_NotEq:
			pred = LLVMIntNE
		}
		res.Value = LLVMBuildICmp(p.Builder, pred, left.Value, right.Value, "")
	} else if is_type_simd_vector(a) {
		mask := LLVMValueRef(0)
		elem := base_array_type(a)
		if is_type_float(elem) {
			pred := LLVMRealPredicate(0)
			switch op_kind {
			case Token_CmpEq:
				pred = LLVMRealOEQ
			case Token_NotEq:
				pred = LLVMRealUNE
			}
			mask = LLVMBuildFCmp(p.Builder, pred, left.Value, right.Value, "")
		} else {
			pred := LLVMIntPredicate(0)
			switch op_kind {
			case Token_CmpEq:
				pred = LLVMIntEQ
			case Token_NotEq:
				pred = LLVMIntNE
			}
			mask = LLVMBuildICmp(p.Builder, pred, left.Value, right.Value, "")
		}
		if mask == 0 {
			panic("Unhandled comparison kind for SIMD vector")
		}

		count := uint(get_array_type_count(a))
		elem_sz := uint(type_size_of(elem) * 8)
		mask_type := LLVMVectorType(LLVMIntTypeInContext(p.Module.Ctx, elem_sz), count)
		mask = LLVMBuildSExtOrBitCast(p.Builder, mask, mask_type, "")

		mask_int_type := LLVMIntTypeInContext(p.Module.Ctx, uint(8*type_size_of(a)))
		mask_int := LLVMBuildBitCast(p.Builder, mask, mask_int_type, "")

		switch op_kind {
		case Token_CmpEq:
			res.Value = LLVMBuildICmp(p.Builder, LLVMIntEQ, mask_int, LLVMConstInt(mask_int_type, ^uint64(0), true), "")
		case Token_NotEq:
			res.Value = LLVMBuildICmp(p.Builder, LLVMIntNE, mask_int, LLVMConstNull(mask_int_type), "")
		}

		return res
	} else if is_type_soa_pointer(a) {
		return lb_compare_records(p, op_kind, left, right, a)
	} else {
		panic("Unhandled comparison kind")
	}

	return res
}

func lb_emit_comp_against_nil(p *lbProcedure, op_kind TokenKind, x lbValue) lbValue {
	res := lbValue{}
	res.Type = t_llvm_bool
	t := x.Type
	bt := base_type(t)
	type_kind := bt.Kind

	switch type_kind {
	case Type_Basic:
		switch bt.Basic.Kind {
		case Basic_rawptr, Basic_cstring:
			if op_kind == Token_CmpEq {
				res.Value = LLVMBuildIsNull(p.Builder, x.Value, "")
			} else if op_kind == Token_NotEq {
				res.Value = LLVMBuildIsNotNull(p.Builder, x.Value, "")
			}
			return res
		case Basic_cstring16:
			if op_kind == Token_CmpEq {
				res.Value = LLVMBuildIsNull(p.Builder, x.Value, "")
			} else if op_kind == Token_NotEq {
				res.Value = LLVMBuildIsNotNull(p.Builder, x.Value, "")
			}
			return res
		case Basic_any:
			data := lb_emit_struct_ev(p, x, 0)
			ti := lb_emit_struct_ev(p, x, 1)
			if op_kind == Token_CmpEq {
				a := LLVMBuildIsNull(p.Builder, data.Value, "")
				b := LLVMBuildIsNull(p.Builder, ti.Value, "")
				res.Value = LLVMBuildOr(p.Builder, a, b, "")
				return res
			} else if op_kind == Token_NotEq {
				a := LLVMBuildIsNotNull(p.Builder, data.Value, "")
				b := LLVMBuildIsNotNull(p.Builder, ti.Value, "")
				res.Value = LLVMBuildAnd(p.Builder, a, b, "")
				return res
			}
		case Basic_typeid:
			invalid_typeid := lb_const_value(p.Module, t_typeid, exact_value_i64(0))
			return lb_emit_comp(p, op_kind, x, invalid_typeid)
		}

	case Type_Enum, Type_Pointer, Type_MultiPointer, Type_Proc:
		if op_kind == Token_CmpEq {
			res.Value = LLVMBuildIsNull(p.Builder, x.Value, "")
		} else if op_kind == Token_NotEq {
			res.Value = LLVMBuildIsNotNull(p.Builder, x.Value, "")
		}
		return res

	case Type_BitSet:
		u := bit_set_to_int(bt)
		if is_type_array(u) {
			args := []lbValue{
				lb_emit_conv(p, lb_address_from_load_or_generate_local(p, x), t_rawptr),
				lb_const_int(p.Module, t_int, uint64(type_size_of(t))),
			}
			val := lb_emit_runtime_call(p, "memory_compare_zero", args)
			res := lb_emit_comp(p, op_kind, val, lb_const_int(p.Module, t_int, 0))
			return res
		} else {
			if op_kind == Token_CmpEq {
				res.Value = LLVMBuildIsNull(p.Builder, x.Value, "")
			} else if op_kind == Token_NotEq {
				res.Value = LLVMBuildIsNotNull(p.Builder, x.Value, "")
			}
			return res
		}

	case Type_Slice:
		data := lb_emit_struct_ev(p, x, 0)
		if op_kind == Token_CmpEq {
			res.Value = LLVMBuildIsNull(p.Builder, data.Value, "")
			return res
		} else if op_kind == Token_NotEq {
			res.Value = LLVMBuildIsNotNull(p.Builder, data.Value, "")
			return res
		}

	case Type_DynamicArray:
		data := lb_emit_struct_ev(p, x, 0)
		if op_kind == Token_CmpEq {
			res.Value = LLVMBuildIsNull(p.Builder, data.Value, "")
			return res
		} else if op_kind == Token_NotEq {
			res.Value = LLVMBuildIsNotNull(p.Builder, data.Value, "")
			return res
		}

	case Type_Map:
		data_ptr := lb_emit_struct_ev(p, x, 0)
		if op_kind == Token_CmpEq {
			res.Value = LLVMBuildIsNull(p.Builder, data_ptr.Value, "")
			return res
		} else {
			res.Value = LLVMBuildIsNotNull(p.Builder, data_ptr.Value, "")
			return res
		}

	case Type_SoaPointer:
		ptr := lb_emit_struct_ev(p, x, 0)
		if op_kind == Token_CmpEq {
			res.Value = LLVMBuildIsNull(p.Builder, ptr.Value, "")
		} else if op_kind == Token_NotEq {
			res.Value = LLVMBuildIsNotNull(p.Builder, ptr.Value, "")
		}
		return res

	case Type_Union:
		if type_size_of(t) == 0 {
			if op_kind == Token_CmpEq {
				return lb_const_bool(p.Module, t_llvm_bool, true)
			} else if op_kind == Token_NotEq {
				return lb_const_bool(p.Module, t_llvm_bool, false)
			}
		} else if is_type_union_maybe_pointer(t) {
			tag := lb_emit_transmute(p, x, t_rawptr)
			return lb_emit_comp_against_nil(p, op_kind, tag)
		} else {
			tag := lb_emit_union_tag_value(p, x)
			return lb_emit_comp(p, op_kind, tag, lb_zero(p.Module, tag.Type))
		}

	case Type_Struct:
		if is_type_soa_struct(t) {
			bt := base_type(t)
			if bt.Struct.SoaKind == StructSoa_Slice {
				the_value := LLVMValueRef(0)
				if len(bt.Struct.Fields) == 0 {
					len_ := lb_soa_struct_len(p, x)
					the_value = len_.Value
				} else {
					first_field := lb_emit_struct_ev(p, x, 0)
					the_value = first_field.Value
				}
				if op_kind == Token_CmpEq {
					res.Value = LLVMBuildIsNull(p.Builder, the_value, "")
					return res
				} else if op_kind == Token_NotEq {
					res.Value = LLVMBuildIsNotNull(p.Builder, the_value, "")
					return res
				}
			} else if bt.Struct.SoaKind == StructSoa_Dynamic {
				the_value := LLVMValueRef(0)
				if len(bt.Struct.Fields) == 0 {
					cap_ := lb_soa_struct_cap(p, x)
					the_value = cap_.Value
				} else {
					first_field := lb_emit_struct_ev(p, x, 0)
					the_value = first_field.Value
				}
				if op_kind == Token_CmpEq {
					res.Value = LLVMBuildIsNull(p.Builder, the_value, "")
					return res
				} else if op_kind == Token_NotEq {
					res.Value = LLVMBuildIsNotNull(p.Builder, the_value, "")
					return res
				}
			}
		} else if is_type_struct(t) && type_has_nil(t) {
			args := []lbValue{
				lb_emit_conv(p, lb_address_from_load_or_generate_local(p, x), t_rawptr),
				lb_const_int(p.Module, t_int, uint64(type_size_of(t))),
			}
			val := lb_emit_runtime_call(p, "memory_compare_zero", args)
			res := lb_emit_comp(p, op_kind, val, lb_const_int(p.Module, t_int, 0))
			return res
		}
	}

	panic("Unknown handled type for nil comparison")
}

func lb_build_unary_and(p *lbProcedure, expr *Ast) lbValue {
	ue := expr.UnaryExpr
	tv := type_and_value_of_expr(expr)

	ue_expr := unparen_expr(ue.Expr)
	if ue_expr.Kind == Ast_IndexExpr && tv.Mode == Addressing_OptionalOkPtr && is_type_tuple(tv.Type) {
		tuple := tv.Type

		map_type := type_of_expr(ue_expr.IndexExpr.Expr)
		ot := base_type(map_type)
		t := base_type(type_deref(ot))
		deref := t != ot
		if t.Kind != Type_Map {
			panic("expected map type")
		}
		ie := ue_expr.IndexExpr

		map_val := lb_build_addr_ptr(p, ie.Expr)
		if deref {
			map_val = lb_emit_load(p, map_val)
		}

		key := lb_build_expr(p, ie.Index)
		key = lb_emit_conv(p, key, t.Map.Key)

		addr := lb_addr_map(map_val, key, t, alloc_type_pointer(t.Map.Value))
		ptr := lb_addr_get_ptr(p, addr)

		ok := lb_emit_comp_against_nil(p, Token_NotEq, ptr)
		ok = lb_emit_conv(p, ok, tuple.Tuple.Variables[1].Type)

		res := lb_add_local_generated(p, tuple, false)
		gep0 := lb_emit_struct_ep(p, res.Addr, 0)
		gep1 := lb_emit_struct_ep(p, res.Addr, 1)
		lb_emit_store(p, gep0, ptr)
		lb_emit_store(p, gep1, ok)
		return lb_addr_load(p, res)

	} else if is_type_soa_pointer(tv.Type) {
		ie := ue_expr.IndexExpr
		addr := lb_build_addr_ptr(p, ie.Expr)

		if is_type_pointer(type_deref(addr.Type)) {
			addr = lb_emit_load(p, addr)
		}
		if !is_type_pointer(addr.Type) {
			panic("expected pointer type")
		}

		index := lb_build_expr(p, ie.Index)

		if !build_context.NoBoundsCheck {
			// TODO(bill): soa bounds checking
		}

		return lb_make_soa_pointer(p, tv.Type, addr, index)

	} else if ue_expr.Kind == Ast_CompoundLit {
		v := lb_build_expr(p, ue.Expr)

		typ := v.Type
		addr := lbAddr{}
		if p.IsStartup {
			addr = lb_add_global_generated_from_procedure(p, typ, v)
		} else {
			addr = lb_add_local_generated(p, typ, false)
		}
		lb_addr_store(p, addr, v)
		return addr.Addr

	} else if ue_expr.Kind == Ast_TypeAssertion {
		if is_type_tuple(tv.Type) {
			tuple := tv.Type
			ptr_type := tuple.Tuple.Variables[0].Type
			ok_type := tuple.Tuple.Variables[1].Type

			ta := ue_expr.TypeAssertion
			pos := ast_token(expr).Pos
			typ := type_of_expr(ue_expr)
			if is_type_tuple(typ) {
				panic("unexpected tuple type")
			}

			e := lb_build_expr(p, ta.Expr)
			t := type_deref(e.Type)
			if is_type_union(t) {
				v := e
				if !is_type_pointer(v.Type) {
					v = lb_address_from_load_or_generate_local(p, v)
				}
				src_type := type_deref(v.Type)
				dst_type := typ

				src_tag := lbValue{}
				dst_tag := lbValue{}
				if is_type_union_maybe_pointer(src_type) {
					src_tag = lb_emit_comp_against_nil(p, Token_NotEq, v)
					dst_tag = lb_const_bool(p.Module, t_bool, true)
				} else {
					src_tag = lb_emit_load(p, lb_emit_union_tag_ptr(p, v))
					dst_tag = lb_const_union_tag(p.Module, src_type, dst_type)
				}

				ok := lb_emit_comp(p, Token_CmpEq, src_tag, dst_tag)

				data_ptr := lb_emit_conv(p, v, ptr_type)
				res := lb_add_local_generated(p, tuple, true)
				gep0 := lb_emit_struct_ep(p, res.Addr, 0)
				gep1 := lb_emit_struct_ep(p, res.Addr, 1)
				lb_emit_store(p, gep0, lb_emit_select(p, ok, data_ptr, lb_const_nil(p.Module, ptr_type)))
				lb_emit_store(p, gep1, lb_emit_conv(p, ok, ok_type))
				return lb_addr_load(p, res)
			} else if is_type_any(t) {
				v := e
				if is_type_pointer(v.Type) {
					v = lb_emit_load(p, v)
				}

				data_ptr := lb_emit_conv(p, lb_emit_struct_ev(p, v, 0), ptr_type)
				any_id := lb_emit_struct_ev(p, v, 1)
				id := lb_typeid(p.Module, typ)

				ok := lb_emit_comp(p, Token_CmpEq, any_id, id)

				res := lb_add_local_generated(p, tuple, false)
				gep0 := lb_emit_struct_ep(p, res.Addr, 0)
				gep1 := lb_emit_struct_ep(p, res.Addr, 1)
				lb_emit_store(p, gep0, lb_emit_select(p, ok, data_ptr, lb_const_nil(p.Module, ptr_type)))
				lb_emit_store(p, gep1, lb_emit_conv(p, ok, ok_type))
				return lb_addr_load(p, res)
			} else {
				panic("TODO(bill): type assertion")
			}

		} else {
			if !is_type_pointer(tv.Type) {
				panic("expected pointer type")
			}

			ta := ue_expr.TypeAssertion
			pos := ast_token(expr).Pos
			typ := type_of_expr(ue_expr)
			if is_type_tuple(typ) {
				panic("unexpected tuple type")
			}

			do_type_check := true
			if build_context.NoTypeAssert {
				feature_flags := lb_get_file_feature_flags(p, ue_expr)
				if (feature_flags & OptInFeatureFlag_ForceTypeAssert) == 0 {
					do_type_check = false
				}
			} else if (p.StateFlags & StateFlag_no_type_assert) != 0 {
				do_type_check = false
			}

			e := lb_build_expr(p, ta.Expr)
			t := type_deref(e.Type)
			if is_type_union(t) {
				v := e
				if !is_type_pointer(v.Type) {
					v = lb_address_from_load_or_generate_local(p, v)
				}
				src_type := type_deref(v.Type)
				dst_type := typ

				if do_type_check {
					src_tag := lbValue{}
					dst_tag := lbValue{}
					if is_type_union_maybe_pointer(src_type) {
						src_tag = lb_emit_comp_against_nil(p, Token_NotEq, v)
						dst_tag = lb_const_bool(p.Module, t_bool, true)
					} else {
						src_tag = lb_emit_load(p, lb_emit_union_tag_ptr(p, v))
						dst_tag = lb_const_union_tag(p.Module, src_type, dst_type)
					}

					arg_count := isize(6)
					if build_context.NoRtti {
						arg_count = 4
					}

					ok := lb_emit_comp(p, Token_CmpEq, src_tag, dst_tag)
					args := make([]lbValue, arg_count)
					args[0] = ok

					lb_set_file_line_col(p, args[1:], pos)

					if !build_context.NoRtti {
						args[4] = lb_typeid(p.Module, src_type)
						args[5] = lb_typeid(p.Module, dst_type)
					}

					name := "type_assertion_check_contextless"
					if len(p.ContextStack) > 0 {
						name = "type_assertion_check_with_context"
					}
					lb_emit_runtime_call(p, name, args)
				}

				data_ptr := v
				return lb_emit_conv(p, data_ptr, tv.Type)
			} else if is_type_any(t) {
				v := e
				if is_type_pointer(v.Type) {
					v = lb_emit_load(p, v)
				}
				data_ptr := lb_emit_struct_ev(p, v, 0)
				if do_type_check {
					if build_context.NoRtti {
						panic("RTTI required for type assertion")
					}

					any_id := lb_emit_struct_ev(p, v, 1)

					arg_count := isize(6)
					if build_context.NoRtti {
						arg_count = 4
					}

					id := lb_typeid(p.Module, typ)
					ok := lb_emit_comp(p, Token_CmpEq, any_id, id)
					args := make([]lbValue, arg_count)
					args[0] = ok

					lb_set_file_line_col(p, args[1:], pos)

					if !build_context.NoRtti {
						args[4] = any_id
						args[5] = id
					}

					name := "type_assertion_check_contextless"
					if len(p.ContextStack) > 0 {
						name = "type_assertion_check_with_context"
					}
					lb_emit_runtime_call(p, name, args)
				}

				return lb_emit_conv(p, data_ptr, tv.Type)
			} else {
				panic("TODO(bill): type assertion")
			}
		}
	}

	return lb_build_addr_ptr(p, ue.Expr)
}
