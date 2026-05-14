package cmd

func check_builtin_simd_operation(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type) bool {
	ce := call.CallExpr
	builtin_name := builtin_procs[id].Name

	switch id {
	case BuiltinProc_simd_add,
		BuiltinProc_simd_sub,
		BuiltinProc_simd_mul,
		BuiltinProc_simd_div,
		BuiltinProc_simd_min,
		BuiltinProc_simd_max,
		BuiltinProc_simd_pairwise_add,
		BuiltinProc_simd_pairwise_sub:

		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer(elem) && !is_type_float(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer or floating point element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		if id == BuiltinProc_simd_div && is_type_integer(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' is not supported for integer elements, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_saturating_add,
		BuiltinProc_simd_saturating_sub,
		BuiltinProc_simd_bit_and,
		BuiltinProc_simd_bit_or,
		BuiltinProc_simd_bit_xor,
		BuiltinProc_simd_bit_and_not:

		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		elem := base_array_type(x.Type)
		switch id {
		case BuiltinProc_simd_saturating_add, BuiltinProc_simd_saturating_sub:
			if !is_type_integer(elem) {
				xs := type_to_string(x.Type)
				error(x.Expr, "'%.*s' expected a #simd type with an integer element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
				gb_string_free(xs)
				return false
			}
		default:
			if !is_type_integer(elem) && !is_type_boolean(elem) {
				xs := type_to_string(x.Type)
				error(x.Expr, "'%.*s' expected a #simd type with an integer or boolean element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
				gb_string_free(xs)
				return false
			}
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_shl,
		BuiltinProc_simd_shr,
		BuiltinProc_simd_shl_masked,
		BuiltinProc_simd_shr_masked:

		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr(c, &y, ce.Args[1])
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			if is_type_untyped(y.Type) || is_type_unsigned(y.Type) {
				rhs_type := type_unsigned_equivalent(x.Type)
				convert_to_typed(c, &y, rhs_type)
				if y.Mode == Addressing_Invalid {
					return false
				}
			} else {
				convert_to_typed(c, &y, x.Type)
				if y.Mode == Addressing_Invalid {
					return false
				}
			}
		}
		if !is_type_simd_vector(y.Type) {
			s := type_to_string(y.Type)
			error(y.Expr, "'%.*s' expected a simd vector type or unsigned integer, got %s", builtin_name.Len, builtin_name.Data, s)
			gb_string_free(s)
			return false
		}
		gb_assert_handler("Assertion Failure", "x.Type.Kind == Type_SimdVector", "cmd_check_builtin_simd.go", 0)
		gb_assert_handler("Assertion Failure", "y.Type.Kind == Type_SimdVector", "cmd_check_builtin_simd.go", 0)
		xt := x.Type
		yt := y.Type
		if xt.SimdVector.Count != yt.SimdVector.Count {
			error(x.Expr, "'%.*s' mismatched simd vector lengths, got '%d' vs '%d'",
				builtin_name.Len, builtin_name.Data,
				xt.SimdVector.Count,
				yt.SimdVector.Count)
			return false
		}
		if !is_type_integer(base_array_type(x.Type)) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		if !is_type_unsigned(base_array_type(y.Type)) {
			ys := type_to_string(y.Type)
			error(y.Expr, "'%.*s' expected a #simd type with an unsigned integer element as the shifting operand, got '%s'", builtin_name.Len, builtin_name.Data, ys)
			gb_string_free(ys)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_neg,
		BuiltinProc_simd_abs:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer(elem) && !is_type_float(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer or floating point element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_lanes_eq,
		BuiltinProc_simd_lanes_ne,
		BuiltinProc_simd_lanes_lt,
		BuiltinProc_simd_lanes_le,
		BuiltinProc_simd_lanes_gt,
		BuiltinProc_simd_lanes_ge:

		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		switch id {
		case BuiltinProc_simd_lanes_eq, BuiltinProc_simd_lanes_ne:
			if !is_type_integer(elem) && !is_type_float(elem) && !is_type_boolean(elem) {
				xs := type_to_string(x.Type)
				error(x.Expr, "'%.*s' expected a #simd type with an integer, floating point, or boolean element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
				gb_string_free(xs)
				return false
			}
		default:
			if !is_type_integer(elem) && !is_type_float(elem) {
				xs := type_to_string(x.Type)
				error(x.Expr, "'%.*s' expected a #simd type with an integer or floating point element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
				gb_string_free(xs)
				return false
			}
		}
		if !are_types_identical(x.Type, y.Type) {
			tx := type_to_string(x.Type)
			ty := type_to_string(y.Type)
			error(call, "Mismatched types to '%.*s', '%s' vs '%s'", builtin_name.Len, builtin_name.Data, tx, ty)
			gb_string_free(ty)
			gb_string_free(tx)
		}
		vt := base_type(x.Type)
		gb_assert_handler("Assertion Failure", "vt.Kind == Type_SimdVector", "cmd_check_builtin_simd.go", 0)
		count := vt.SimdVector.Count
		sz := type_size_of(elem)
		var new_elem *Type
		switch sz {
		case 1:
			new_elem = t_u8
		case 2:
			new_elem = t_u16
		case 4:
			new_elem = t_u32
		case 8:
			new_elem = t_u64
		case 16:
			error(x.Expr, "'%.*s' not supported 128-bit integer backed simd vector types", builtin_name.Len, builtin_name.Data)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = alloc_type_simd_vector(count, new_elem, nil)
		return true

	case BuiltinProc_simd_gather,
		BuiltinProc_simd_scatter,
		BuiltinProc_simd_masked_load,
		BuiltinProc_simd_masked_store,
		BuiltinProc_simd_masked_expand_load,
		BuiltinProc_simd_masked_compress_store:

		var ptr, values, mask Operand
		check_expr(c, &ptr, ce.Args[0])
		if ptr.Mode == Addressing_Invalid {
			return false
		}
		check_expr(c, &values, ce.Args[1])
		if values.Mode == Addressing_Invalid {
			return false
		}
		check_expr(c, &mask, ce.Args[2])
		if mask.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(values.Type) {
			error(values.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(mask.Type) {
			error(mask.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if id == BuiltinProc_simd_gather || id == BuiltinProc_simd_scatter {
			if !is_type_simd_vector(ptr.Type) {
				error(ptr.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
				return false
			}
			ptr_elem := base_array_type(ptr.Type)
			if !is_type_rawptr(ptr_elem) {
				s := type_to_string(ptr.Type)
				error(ptr.Expr, "Expected a simd vector of 'rawptr' for the addresses, got %s", s)
				gb_string_free(s)
				return false
			}
		} else {
			if !is_type_pointer(ptr.Type) {
				s := type_to_string(ptr.Type)
				error(ptr.Expr, "Expected a pointer type for the address, got %s", s)
				gb_string_free(s)
				return false
			}
		}
		mask_elem := base_array_type(mask.Type)
		if !is_type_integer(mask_elem) && !is_type_boolean(mask_elem) {
			s := type_to_string(mask.Type)
			error(mask.Expr, "Expected a simd vector of integers or booleans for the mask, got %s", s)
			gb_string_free(s)
			return false
		}
		if id == BuiltinProc_simd_gather || id == BuiltinProc_simd_scatter {
			ptr_count := get_array_type_count(ptr.Type)
			values_count := get_array_type_count(values.Type)
			mask_count := get_array_type_count(mask.Type)
			if ptr_count != values_count || values_count != mask_count || mask_count != ptr_count {
				s := type_to_string(mask.Type)
				error(mask.Expr, "All simd vectors must be of the same length, got %d vs %d vs %d", ptr_count, values_count, mask_count)
				gb_string_free(s)
				return false
			}
		} else {
			values_count := get_array_type_count(values.Type)
			mask_count := get_array_type_count(mask.Type)
			if values_count != mask_count {
				s := type_to_string(mask.Type)
				error(mask.Expr, "All simd vectors must be of the same length, got %d vs %d", values_count, mask_count)
				gb_string_free(s)
				return false
			}
		}
		if id == BuiltinProc_simd_gather ||
			id == BuiltinProc_simd_masked_load ||
			id == BuiltinProc_simd_masked_expand_load {
			operand.Mode = Addressing_Value
			operand.Type = values.Type
		} else {
			operand.Mode = Addressing_NoValue
			operand.Type = nil
		}
		return true

	case BuiltinProc_simd_indices:
		var x Operand
		check_expr_or_type(c, &x, ce.Args[0], nil)
		if x.Mode == Addressing_Invalid {
			return false
		}
		if x.Mode != Addressing_Type {
			s := expr_to_string(x.Expr)
			error(x.Expr, "'%.*s' expected a simd vector type, got '%s'", builtin_name.Len, builtin_name.Data, s)
			gb_string_free(s)
			return false
		}
		if !is_type_simd_vector(x.Type) {
			s := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a simd vector type, got '%s'", builtin_name.Len, builtin_name.Data, s)
			gb_string_free(s)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_numeric(elem) {
			s := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a simd vector type with a numeric element type, got '%s'", builtin_name.Len, builtin_name.Data, s)
			gb_string_free(s)
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_extract:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		max_count := x.Type.SimdVector.Count
		var value int64 = -1
		if !check_index_value(c, x.Type, false, ce.Args[1], max_count, &value) {
			return false
		}
		if max_count < 0 {
			error(ce.Args[1], "'%.*s' expected a constant integer index, got '%d'", builtin_name.Len, builtin_name.Data, value)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = elem
		return true

	case BuiltinProc_simd_replace:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		max_count := x.Type.SimdVector.Count
		var value int64 = -1
		if !check_index_value(c, x.Type, false, ce.Args[1], max_count, &value) {
			return false
		}
		if max_count < 0 {
			error(ce.Args[1], "'%.*s' expected a constant integer index, got '%d'", builtin_name.Len, builtin_name.Data, value)
			return false
		}
		var y Operand
		check_expr_with_type_hint(c, &y, ce.Args[2], elem)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, elem)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(y.Type, elem) {
			et := type_to_string(elem)
			yt := type_to_string(y.Type)
			error(y.Expr, "'%.*s' expected a type of '%s' to insert, got '%s'", builtin_name.Len, builtin_name.Data, et, yt)
			gb_string_free(yt)
			gb_string_free(et)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_reduce_add_bisect,
		BuiltinProc_simd_reduce_mul_bisect,
		BuiltinProc_simd_reduce_add_ordered,
		BuiltinProc_simd_reduce_mul_ordered,
		BuiltinProc_simd_reduce_add_pairs,
		BuiltinProc_simd_reduce_mul_pairs,
		BuiltinProc_simd_reduce_min,
		BuiltinProc_simd_reduce_max:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer(elem) && !is_type_float(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer or floating point element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = base_array_type(x.Type)
		return true

	case BuiltinProc_simd_reduce_and,
		BuiltinProc_simd_reduce_or,
		BuiltinProc_simd_reduce_xor:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer(elem) && !is_type_boolean(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer or boolean element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = base_array_type(x.Type)
		return true

	case BuiltinProc_simd_reduce_any,
		BuiltinProc_simd_reduce_all:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_boolean(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with a boolean element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = t_untyped_bool
		return true

	case BuiltinProc_simd_extract_lsbs,
		BuiltinProc_simd_extract_msbs:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a simd vector type, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer_like(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with integer or boolean elements, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		num_elems := get_array_type_count(x.Type)
		result_type := alloc_type_bit_set()
		result_type.BitSet.Elem = t_int
		result_type.BitSet.Lower = 0
		result_type.BitSet.Upper = num_elems - 1
		operand.Mode = Addressing_Value
		operand.Type = result_type
		return true

	case BuiltinProc_simd_shuffle:
		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		elem := base_array_type(x.Type)
		max_count := x.Type.SimdVector.Count + y.Type.SimdVector.Count
		var arg_count int64
		for i := range ce.Args {
			if i < 2 {
				continue
			}
			arg := ce.Args[i]
			var op Operand
			check_expr(c, &op, arg)
			if op.Mode == Addressing_Invalid {
				return false
			}
			arg_type := base_type(op.Type)
			if !is_type_integer(arg_type) || op.Mode != Addressing_Constant {
				error(op.Expr, "Indices to '%.*s' must be constant integers", builtin_name.Len, builtin_name.Data)
				return false
			}
			if big_int_is_neg(&op.Value.ValueInteger) {
				error(op.Expr, "Negative '%.*s' index", builtin_name.Len, builtin_name.Data)
				return false
			}
			var mc BigInt
			big_int_from_i64(&mc, max_count)
			if big_int_cmp(&mc, &op.Value.ValueInteger) <= 0 {
				error(op.Expr, "'%.*s' index exceeds length", builtin_name.Len, builtin_name.Data)
				return false
			}
			arg_count++
		}
		if arg_count > max_count {
			error(call, "Too many '%.*s' indices, %d > %d", builtin_name.Len, builtin_name.Data, arg_count, max_count)
			return false
		}
		if !is_power_of_two(arg_count) {
			error(call, "'%.*s' must have a power of two index arguments, got %d", builtin_name.Len, builtin_name.Data, arg_count)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = alloc_type_simd_vector(arg_count, elem, nil)
		return true

	case BuiltinProc_simd_odd_even:
		var x, y Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_select:
		var cond Operand
		check_expr(c, &cond, ce.Args[0])
		if cond.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(cond.Type) {
			error(cond.Expr, "'%.*s' expected a simd vector boolean type", builtin_name.Len, builtin_name.Data)
			return false
		}
		cond_elem := base_array_type(cond.Type)
		if !is_type_boolean(cond_elem) && !is_type_integer(cond_elem) {
			cond_str := type_to_string(cond.Type)
			error(cond.Expr, "'%.*s' expected a simd vector boolean or integer type, got '%s'", builtin_name.Len, builtin_name.Data, cond_str)
			gb_string_free(cond_str)
			return false
		}
		var x, y Operand
		check_expr(c, &x, ce.Args[1])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[2], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 results of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		if cond.Type.SimdVector.Count != x.Type.SimdVector.Count {
			error(x.Expr, "'%.*s' expected condition vector to match the length of the result lengths, got '%d' vs '%d'",
				builtin_name.Len, builtin_name.Data,
				cond.Type.SimdVector.Count,
				x.Type.SimdVector.Count)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_runtime_swizzle:
		if len(ce.Args) != 2 {
			error(call, "'%.*s' expected 2 arguments, got %d", builtin_name.Len, builtin_name.Data, len(ce.Args))
			return false
		}
		var src, indices Operand
		check_expr(c, &src, ce.Args[0])
		if src.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &indices, ce.Args[1], src.Type)
		if indices.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(src.Type) {
			error(src.Expr, "'%.*s' expected first argument to be a simd vector", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(indices.Type) {
			error(indices.Expr, "'%.*s' expected second argument (indices) to be a simd vector", builtin_name.Len, builtin_name.Data)
			return false
		}
		src_elem := base_array_type(src.Type)
		indices_elem := base_array_type(indices.Type)
		if !is_type_integer(src_elem) {
			src_str := type_to_string(src.Type)
			error(src.Expr, "'%.*s' expected first argument to be a simd vector of integers, got '%s'", builtin_name.Len, builtin_name.Data, src_str)
			gb_string_free(src_str)
			return false
		}
		if !is_type_integer(indices_elem) {
			indices_str := type_to_string(indices.Type)
			error(indices.Expr, "'%.*s' expected indices to be a simd vector of integers, got '%s'", builtin_name.Len, builtin_name.Data, indices_str)
			gb_string_free(indices_str)
			return false
		}
		if !are_types_identical(src.Type, indices.Type) {
			src_str := type_to_string(src.Type)
			indices_str := type_to_string(indices.Type)
			error(indices.Expr, "'%.*s' expected both arguments to have the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, src_str, indices_str)
			gb_string_free(indices_str)
			gb_string_free(src_str)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = src.Type
		return true

	case BuiltinProc_simd_sums_of_n:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		bt := base_type(x.Type)
		max_count := uint64(bt.SimdVector.Count)
		elem := bt.SimdVector.Elem
		var y Operand
		check_expr(c, &y, ce.Args[1])
		if y.Mode == Addressing_Invalid {
			return false
		}
		arg_type := base_type(y.Type)
		if !is_type_integer(arg_type) || y.Mode != Addressing_Constant {
			error(y.Expr, "Indices to '%.*s' must be constant integers", builtin_name.Len, builtin_name.Data)
			return false
		}
		if big_int_is_neg(&y.Value.ValueInteger) {
			error(y.Expr, "Negative '%.*s' index", builtin_name.Len, builtin_name.Data)
			return false
		}
		n := exact_value_to_u64(y.Value)

		if !(is_power_of_two_u64(n) && n >= 2) {
			error(y.Expr, "'%.*s' requires a power of two 'n' parameter >= 2, got %d", builtin_name.Len, builtin_name.Data, n)
			return false
		}
		if n > max_count {
			error(y.Expr, "'%.*s' requires that the 'n' parameter is <= than the #simd length, got %d vs %d", builtin_name.Len, builtin_name.Data, n, max_count)
			return false
		}
		if max_count%n != 0 {
			error(y.Expr, "'%.*s' requires the #simd length to be a multiple of the 'n' parameter, got #simd length=%d, n=%d", builtin_name.Len, builtin_name.Data, max_count, n)
			return false
		}
		operand.Mode = Addressing_Value
		result_count := max_count / n
		if result_count == 1 {
			operand.Type = elem
		} else {
			operand.Type = alloc_type_simd_vector(int64(result_count), elem, nil)
		}
		return true

	case BuiltinProc_simd_ceil,
		BuiltinProc_simd_floor,
		BuiltinProc_simd_trunc,
		BuiltinProc_simd_nearest,
		BuiltinProc_simd_approx_recip,
		BuiltinProc_simd_approx_recip_sqrt:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_float(elem) {
			x_str := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a simd vector floating point type, got '%s'", builtin_name.Len, builtin_name.Data, x_str)
			gb_string_free(x_str)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_lanes_reverse:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		operand.Type = x.Type
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_simd_lanes_rotate_left,
		BuiltinProc_simd_lanes_rotate_right:

		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		var offset Operand
		check_expr(c, &offset, ce.Args[1])
		if offset.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &offset, t_i64)
		if !is_type_integer(offset.Type) || offset.Mode != Addressing_Constant {
			error(offset.Expr, "'%.*s' expected a constant integer offset", builtin_name.Len, builtin_name.Data)
			return false
		}
		check_assignment(c, &offset, t_i64, builtin_name)
		operand.Type = x.Type
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_simd_clamp:
		var x, y, z Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &y, ce.Args[1], x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		check_expr_with_type_hint(c, &z, ce.Args[2], x.Type)
		if z.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &y, x.Type)
		if y.Mode == Addressing_Invalid {
			return false
		}
		convert_to_typed(c, &z, x.Type)
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(y.Type) {
			error(y.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !is_type_simd_vector(z.Type) {
			error(z.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		if !are_types_identical(x.Type, y.Type) {
			xs := type_to_string(x.Type)
			ys := type_to_string(y.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, ys)
			gb_string_free(ys)
			gb_string_free(xs)
			return false
		}
		if !are_types_identical(x.Type, z.Type) {
			xs := type_to_string(x.Type)
			zs := type_to_string(z.Type)
			error(x.Expr, "'%.*s' expected 2 arguments of the same type, got '%s' vs '%s'", builtin_name.Len, builtin_name.Data, xs, zs)
			gb_string_free(zs)
			gb_string_free(xs)
			return false
		}
		elem := base_array_type(x.Type)
		if !is_type_integer(elem) && !is_type_float(elem) {
			xs := type_to_string(x.Type)
			error(x.Expr, "'%.*s' expected a #simd type with an integer or floating point element, got '%s'", builtin_name.Len, builtin_name.Data, xs)
			gb_string_free(xs)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = x.Type
		return true

	case BuiltinProc_simd_to_bits:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		count := get_array_type_count(x.Type)
		bit_elem := type_unsigned_equivalent(elem)
		operand.Type = alloc_type_simd_vector(count, bit_elem, nil)
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_simd_to_bits_signed:
		var x Operand
		check_expr(c, &x, ce.Args[0])
		if x.Mode == Addressing_Invalid {
			return false
		}
		if !is_type_simd_vector(x.Type) {
			error(x.Expr, "'%.*s' expected a simd vector type", builtin_name.Len, builtin_name.Data)
			return false
		}
		elem := base_array_type(x.Type)
		count := get_array_type_count(x.Type)
		bit_elem := type_signed_equivalent(elem)
		operand.Type = alloc_type_simd_vector(count, bit_elem, nil)
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_simd_x86__MM_SHUFFLE:
		var x [4]Operand
		for i := 0; i < 4; i++ {
			check_expr(c, &x[i], ce.Args[i])
			if x[i].Mode == Addressing_Invalid {
				return false
			}
		}
		offsets := [4]uint32{6, 4, 2, 0}
		var result uint32
		for i := 0; i < 4; i++ {
			if !is_type_integer(x[i].Type) || x[i].Mode != Addressing_Constant {
				error(x[i].Expr, "'%.*s' expected a constant integer", builtin_name.Len, builtin_name.Data)
				return false
			}
			val := exact_value_to_i64(x[i].Value)
			if val < 0 || val > 3 {
				error(x[i].Expr, "'%.*s' expected a constant integer in the range 0..<4, got %d", builtin_name.Len, builtin_name.Data, val)
				return false
			}
			result |= uint32(val) << offsets[i]
		}
		operand.Type = t_untyped_integer
		operand.Mode = Addressing_Constant
		operand.Value = exact_value_i64(int64(result))
		return true

	default:
		gb_assert_handler("Panic", "Unhandled simd intrinsic", "cmd_check_builtin_simd.go", 0)
	}
	return false
}
