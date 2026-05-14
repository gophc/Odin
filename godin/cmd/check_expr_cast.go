package cmd

func check_update_float_precision(value *ExactValue, type_ *Type) bool {
	type_ = core_type(type_)
	if type_.Kind != Type_Basic {
		return false
	}
	switch type_.Basic.kind {
	case Basic_f16, Basic_f16le, Basic_f16be, Basic_f32, Basic_f32le, Basic_f32be:
		x_64 := exact_value_to_f64(*value)
		x_32 := float32(x_64)
		if float64(x_32) != x_64 {
			*value = exact_value_float(float64(x_32))
			return true
		}
	}
	return false
}

func check_representable_as_constant(c *CheckerContext, in_value ExactValue, type_ *Type, out_value *ExactValue) bool {
	if in_value.Kind == ExactValue_Invalid {
		return true
	}
	type_ = core_type(type_)
	if type_ == tInvalid {
		return false
	} else if is_type_boolean(type_) {
		return in_value.Kind == ExactValue_Bool
	} else if is_type_string(type_) {
		if in_value.Kind == ExactValue_String16 {
			return is_type_string16(type_) || is_type_cstring16(type_)
		}
		return in_value.Kind == ExactValue_String
	} else if is_type_integer(type_) || is_type_rune(type_) {
		v := exact_value_to_integer(in_value)
		if v.Kind != ExactValue_Integer {
			return false
		}
		if out_value != nil {
			*out_value = v
		}
		if is_type_untyped(type_) {
			return true
		}
		i := v.ValueInteger
		byte_size := type_size_of(type_)
		var umax BigInt
		var imin BigInt
		var imax BigInt
		var umax_64 uint64
		var imin_64 int64
		var imax_64 int64

		if c.bit_field_bit_size > 0 {
			bit_size := int64(8 * byte_size)
			if bit_size > c.bit_field_bit_size {
				bit_size = c.bit_field_bit_size
			}
			big_int_from_u64(&umax, 1)
			big_int_from_i64(&imin, 1)
			big_int_from_i64(&imax, 1)
			var bu BigInt
			var bi BigInt
			big_int_from_i64(&bu, bit_size)
			big_int_from_i64(&bi, bit_size-1)
			big_int_shl(&umax, &umax, &bu)
			tmp_one := big_int_make_u64(1)
			big_int_sub(&umax, &umax, &tmp_one)
			big_int_shl(&imin, &imin, &bi)
			big_int_neg(&imin, &imin)
			big_int_shl(&imax, &imax, &bi)
			big_int_sub(&imax, &imax, &tmp_one)
		} else {
			if byte_size <= 8 {
				umax_64 = unsigned_integer_maxs[byte_size]
				imin_64 = signed_integer_mins[byte_size]
				imax_64 = signed_integer_maxs[byte_size]
			} else {
				big_int_from_u64(&umax, 1)
				big_int_from_i64(&imin, 1)
				big_int_from_i64(&imax, 1)
				var bi128 BigInt
				var bi127 BigInt
				big_int_from_i64(&bi128, 128)
				big_int_from_i64(&bi127, 127)
				big_int_shl(&umax, &umax, &bi128)
				one := big_int_make_u64(1)
				big_int_sub(&umax, &umax, &one)
				big_int_shl(&imin, &imin, &bi127)
				big_int_neg(&imin, &imin)
				big_int_shl(&imax, &imax, &bi127)
				big_int_sub(&imax, &imax, &one)
			}
		}
		switch type_.Basic.kind {
		case Basic_rune,
			Basic_i8, Basic_i16, Basic_i32, Basic_i64, Basic_int,
			Basic_i16le, Basic_i32le, Basic_i64le,
			Basic_i16be, Basic_i32be, Basic_i64be:
			if c.bit_field_bit_size == 0 {
				if !basic_int_can_be_represented_in_64_bits(&i) {
					return false
				}
				val64 := big_int_to_i64(&i)
				return imin_64 <= val64 && val64 <= imax_64
			}
			fallthrough
		case Basic_i128, Basic_i128le, Basic_i128be:
			{
				a := big_int_cmp(&imin, &i)
				b := big_int_cmp(&i, &imax)
				return (a <= 0) && (b <= 0)
			}
		case Basic_u8, Basic_u16, Basic_u32, Basic_u64, Basic_uint, Basic_uintptr,
			Basic_u16le, Basic_u32le, Basic_u64le,
			Basic_u16be, Basic_u32be, Basic_u64be:
			if c.bit_field_bit_size == 0 {
				if big_int_is_neg(&i) {
					return false
				}
				if !basic_int_can_be_represented_in_64_bits(&i) {
					return false
				}
				val64 := big_int_to_u64(&i)
				return val64 <= umax_64
			}
			fallthrough
		case Basic_u128, Basic_u128le, Basic_u128be:
			{
				b := big_int_cmp(&i, &umax)
				return !big_int_is_neg(&i) && (b <= 0)
			}
		case Basic_UntypedInteger:
			return true
		}
	} else if is_type_float(type_) {
		v := exact_value_to_float(in_value)
		if v.Kind != ExactValue_Float {
			return false
		}
		check_update_float_precision(&v, type_)
		if out_value != nil {
			*out_value = v
		}
		switch type_.Basic.kind {
		case Basic_f16, Basic_f16le, Basic_f16be,
			Basic_f32, Basic_f32le, Basic_f32be,
			Basic_f64, Basic_f64le, Basic_f64be,
			Basic_UntypedFloat:
			return true
		}
	} else if is_type_complex(type_) {
		v := exact_value_to_complex(in_value)
		if v.Kind != ExactValue_Complex {
			return false
		}
		switch type_.Basic.kind {
		case Basic_complex32, Basic_complex64, Basic_complex128:
			real := exact_value_real(v)
			imag := exact_value_imag(v)
			if real.Kind != ExactValue_Invalid && imag.Kind != ExactValue_Invalid {
				if out_value != nil {
					*out_value = exact_value_complex(exact_value_to_f64(real), exact_value_to_f64(imag))
				}
				return true
			}
		case Basic_UntypedComplex:
			return true
		}
		return false
	} else if is_type_quaternion(type_) {
		v := exact_value_to_quaternion(in_value)
		if v.Kind != ExactValue_Quaternion {
			return false
		}
		switch type_.Basic.kind {
		case Basic_quaternion64, Basic_quaternion128, Basic_quaternion256:
			real := exact_value_real(v)
			imag := exact_value_imag(v)
			jmag := exact_value_jmag(v)
			kmag := exact_value_kmag(v)
			if real.Kind != ExactValue_Invalid && imag.Kind != ExactValue_Invalid {
				if out_value != nil {
					*out_value = exact_value_quaternion(exact_value_to_f64(real), exact_value_to_f64(imag), exact_value_to_f64(jmag), exact_value_to_f64(kmag))
				}
				return true
			}
		case Basic_UntypedComplex:
			if out_value != nil {
				*out_value = exact_value_to_quaternion(*out_value)
			}
			return true
		case Basic_UntypedQuaternion:
			return true
		}
		return false
	} else if is_type_pointer(type_) {
		if in_value.Kind == ExactValue_Pointer {
			return true
		}
		if in_value.Kind == ExactValue_Integer {
			return false
		}
		if in_value.Kind == ExactValue_String {
			return false
		}
		if in_value.Kind == ExactValue_String16 {
			return false
		}
		if out_value != nil {
			*out_value = in_value
		}
	} else if is_type_bit_set(type_) {
		if in_value.Kind == ExactValue_Integer {
			return true
		}
	} else if is_type_typeid(type_) {
		if in_value.Kind == ExactValue_Compound {
			cl := &in_value.ValueCompound.CompoundLit
			if len(cl.Elems) == 0 {
				in_value = exact_value_typeid(nil)
			} else {
				return false
			}
		}
		if in_value.Kind == ExactValue_Typeid {
			if out_value != nil {
				*out_value = in_value
			}
			return true
		}
	}
	return false
}

func check_integer_exceed_suggestion(c *CheckerContext, o *Operand, type_ *Type, max_bit_size ...int64) bool {
	if is_type_integer(type_) && o.Value.Kind == ExactValue_Integer {
		b := type_to_string(type_)
		defer gb_string_free(b)
		if is_type_enum(o.Type) {
			if check_is_castable_to(c, o, type_) {
				ot := type_to_string(o.Type)
				error_line("\tSuggestion: Try casting the '%s' expression to '%s'", ot, b)
				gb_string_free(ot)
			}
			return true
		}
		sz := type_size_of(type_)
		bit_size := 8 * sz
		size_changed := false
		if len(max_bit_size) > 0 && max_bit_size[0] > 0 {
			size_changed = (bit_size != max_bit_size[0])
			if bit_size > max_bit_size[0] {
				bit_size = max_bit_size[0]
			}
		}
		bi := &o.Value.ValueInteger
		if is_type_unsigned(type_) {
			one := big_int_make_u64(1)
			max_size := big_int_make_u64(1)
			bits := big_int_make_i64(bit_size)
			big_int_shl(&max_size, &max_size, &bits)
			big_int_sub(&max_size, &max_size, &one)
			if big_int_is_neg(bi) {
				error_line("\tA negative value cannot be represented by the unsigned integer type '%s'\n", b)
				var dst BigInt
				big_int_neg(&dst, bi)
				if big_int_cmp(&dst, &max_size) < 0 {
					big_int_sub(&dst, &dst, &one)
					dst_str := big_int_to_string(temporary_allocator(), &dst)
					t := type_to_string(type_)
					error_line("\tSuggestion: ~%s(%.*s)\n", t, len(dst_str), dst_str.Text)
					gb_string_free(t)
				}
			} else {
				max_size_str := big_int_to_string(temporary_allocator(), &max_size)
				if size_changed {
					error_line("\tThe maximum value that can be represented with that bit_field's field of '%s | %d' is '%.*s'\n", b, bit_size, len(max_size_str), max_size_str.Text)
				} else {
					error_line("\tThe maximum value that can be represented by '%s' is '%.*s'\n", b, len(max_size_str), max_size_str.Text)
				}
			}
		} else {
			one := big_int_make_u64(1)
			max_size := big_int_make_u64(1)
			bits := big_int_make_i64(bit_size - 1)
			big_int_shl(&max_size, &max_size, &bits)
			var max_size_str String
			if big_int_is_neg(bi) {
				big_int_neg(&max_size, &max_size)
				max_size_str = big_int_to_string(temporary_allocator(), &max_size)
			} else {
				big_int_sub(&max_size, &max_size, &one)
				max_size_str = big_int_to_string(temporary_allocator(), &max_size)
			}
			if size_changed {
				error_line("\tThe maximum value that can be represented with that bit_field's field of '%s | %d' is '%.*s'\n", b, bit_size, len(max_size_str), max_size_str.Text)
			} else {
				error_line("\tThe maximum value that can be represented by '%s' is '%.*s'\n", b, len(max_size_str), max_size_str.Text)
			}
		}
		return true
	}
	return false
}

func check_assignment_error_suggestion(c *CheckerContext, o *Operand, type_ *Type, max_bit_size ...int64) {
	a := expr_to_string(o.Expr)
	b := type_to_string(type_)
	defer func() {
		gb_string_free(b)
		gb_string_free(a)
	}()
	src := base_type(o.Type)
	dst := base_type(type_)
	if is_type_array(src) && is_type_slice(dst) {
		s := src.Array.elem
		d := dst.Slice.elem
		if are_types_identical(s, d) {
			error_line("\tSuggestion: The array expression may be sliced with %s[:]\n", a)
		}
	} else if is_type_dynamic_array(src) && is_type_slice(dst) {
		s := src.DynamicArray.elem
		d := dst.Slice.elem
		if are_types_identical(s, d) {
			error_line("\tSuggestion: The dynamic array expression may be sliced with %s[:]\n", a)
		}
	} else if are_types_identical(src, dst) && !are_types_identical(o.Type, type_) {
		error_line("\tSuggestion: The expression may be directly casted to type %s\n", b)
	} else if are_types_identical(src, tString) && is_type_u8_slice(dst) {
		error_line("\tSuggestion: A string may be transmuted to %s\n", b)
		error_line("\t            This is an UNSAFE operation as string data is assumed to be immutable,\n")
		error_line("\t            whereas slices in general are assumed to be mutable.\n")
	} else if is_type_u8_slice(src) && are_types_identical(dst, tString) && o.Mode != Addressing_Constant {
		error_line("\tSuggestion: The expression may be casted to %s\n", b)
	} else if check_integer_exceed_suggestion(c, o, type_, max_bit_size...) {
		return
	} else if is_expr_inferred_fixed_array(c.type_hint_expr) && is_type_array_like(type_) && is_type_array_like(o.Type) {
		s := expr_to_string(c.type_hint_expr)
		error_line("\tSuggestion: Make sure that `%s` is attached to the compound literal directly\n", s)
		gb_string_free(s)
	} else if is_type_pointer(type_) &&
		o.Mode == Addressing_Variable &&
		are_types_identical(type_deref(type_), o.Type) {
		s := expr_to_string(o.Expr)
		error_line("\tSuggestion: Did you mean `&%s`\n", s)
		gb_string_free(s)
	} else if is_type_pointer(o.Type) &&
		are_types_identical(type_deref(o.Type), type_) {
		s := expr_to_string(o.Expr)
		if s != "" && s[0] == '&' {
			error_line("\tSuggestion: Did you mean `%s`\n", s[1:])
		} else {
			error_line("\tSuggestion: Did you mean `%s^`\n", s)
		}
		gb_string_free(s)
	}
}

func check_cast_error_suggestion(c *CheckerContext, o *Operand, type_ *Type) {
	a := expr_to_string(o.Expr)
	b := type_to_string(type_)
	defer func() {
		gb_string_free(b)
		gb_string_free(a)
	}()
	src := base_type(o.Type)
	dst := base_type(type_)
	if is_type_array(src) && is_type_slice(dst) {
		s := src.Array.elem
		d := dst.Slice.elem
		if are_types_identical(s, d) {
			error_line("\tSuggestion: the array expression may be sliced with %s[:]\n", a)
		}
	} else if is_type_pointer(o.Type) && is_type_integer(type_) {
		if is_type_uintptr(type_) {
			error_line("\tSuggestion: a pointer may be directly casted to %s\n", b)
		} else {
			error_line("\tSuggestion: for a pointer to be casted to an integer, it must be converted to 'uintptr' first\n")
			x := type_size_of(o.Type)
			y := type_size_of(type_)
			if x != y {
				error_line("\tNote: the type of expression and the type of the cast have a different size in bytes, %d vs %d\n", x, y)
			}
		}
	} else if is_type_integer(o.Type) && is_type_pointer(type_) {
		if is_type_uintptr(o.Type) {
			error_line("\tSuggestion: %s may be directly casted to %s\n", a, b)
		} else {
			error_line("\tSuggestion: for an integer to be casted to a pointer, it must be converted to 'uintptr' first\n")
		}
	} else if are_types_identical(src, tString) && is_type_u8_slice(dst) {
		error_line("\tSuggestion: a string may be transmuted to %s\n", b)
	} else if check_integer_exceed_suggestion(c, o, type_) {
		return
	}
}

func check_is_expressible(ctx *CheckerContext, o *Operand, type_ *Type) bool {
	out_value := o.Value
	if is_type_constant_type(type_) && check_representable_as_constant(ctx, o.Value, type_, &out_value) {
		o.Value = out_value
		return true
	} else {
		o.Value = out_value
		a := expr_to_string(o.Expr)
		b := type_to_string(type_)
		c := type_to_string(o.Type)
		s := exact_value_to_string(o.Value)
		defer func() {
			gb_string_free(s)
			gb_string_free(c)
			gb_string_free(b)
			gb_string_free(a)
			o.Mode = Addressing_Invalid
		}()
		begin_error_block()
		defer end_error_block()
		if is_type_numeric(o.Type) && is_type_numeric(type_) {
			if !is_type_integer(o.Type) && is_type_integer(type_) {
				error(o.Expr, "'%s' truncated to '%s', got %s", a, b, s)
			} else {
				var max_bit_size int64
	if ctx.bit_field_bit_size > 0 {
				max_bit_size = ctx.bit_field_bit_size
				}
				if are_types_identical(o.Type, type_) {
					error(o.Expr, "Numeric value '%s' from '%s' cannot be represented by '%s'", s, a, b)
				} else {
					error(o.Expr, "Cannot convert numeric value '%s' from '%s' to '%s' from '%s'", s, a, b, c)
				}
				check_assignment_error_suggestion(ctx, o, type_, max_bit_size)
			}
		} else {
			error(o.Expr, "Cannot convert '%s' to '%s' from '%s', got %s", a, b, c, s)
			check_assignment_error_suggestion(ctx, o, type_)
		}
		return false
	}
}

func check_is_not_addressable(c *CheckerContext, o *Operand) bool {
	if o.Expr != nil && o.Expr.Kind == Ast_SelectorExpr {
		if o.Expr.SelectorExpr.IsBitField {
			return true
		}
	}
	if o.Mode == Addressing_OptionalOk {
		expr := unselector_expr(o.Expr)
		if expr.Kind != Ast_TypeAssertion {
			return true
		}
		ta := &expr.TypeAssertion
		tv := ta.Expr.TAV
		if is_type_pointer(tv.Type) {
			return false
		}
		if is_type_union(tv.Type) && tv.Mode == Addressing_Variable {
			return false
		}
		if is_type_any(tv.Type) {
			return false
		}
		return true
	}
	if o.Mode == Addressing_MapIndex {
		return false
	}
	expr := unparen_expr(o.Expr)
	if expr.Kind == Ast_CompoundLit {
		return false
	}
	return o.Mode != Addressing_Variable && o.Mode != Addressing_SoaVariable
}

func check_is_castable_to(c *CheckerContext, operand *Operand, y *Type) bool {
	if are_types_identical(operand.Type, y) {
		return true
	}
	if check_is_assignable_to(c, operand, y) {
		return true
	}
	is_constant := operand.Mode == Addressing_Constant
	x := operand.Type
	src := core_type(x)
	dst := core_type(y)
	if are_types_identical(src, dst) {
		return true
	}
	if is_constant && is_type_untyped(src) && is_type_string(src) {
		if is_type_u8_array(dst) {
			s := operand.Value.ValueString
			return s.Len == dst.Array.count
		}
		if is_type_rune_array(dst) {
			s := operand.Value.ValueString
			return gb_utf8_strnlen(s.Text, s.Len) == dst.Array.count
		}
	}
	if dst.Kind == Type_Array && src.Kind == Type_Array {
		if dst.Array.count == src.Array.count {
			if are_types_identical(dst.Array.elem, src.Array.elem) {
				return true
			}
			op := *operand
			op.Type = src.Array.elem
			return check_is_castable_to(c, &op, dst.Array.elem)
		}
	}
	if dst.Kind == Type_Slice && src.Kind == Type_Slice {
		return are_types_identical(dst.Slice.elem, src.Slice.elem)
	}
	if is_type_boolean(src) || is_type_integer(src) {
		if is_type_boolean(dst) || is_type_integer(dst) {
			return true
		}
	}
	if is_type_integer(src) || is_type_float(src) {
		if is_type_integer(dst) || is_type_float(dst) {
			return true
		}
	}
	if is_type_bit_field(src) {
		return are_types_identical(core_type(src.BitField.BackingType), dst)
	}
	if is_type_bit_field(dst) {
		return are_types_identical(src, core_type(dst.BitField.BackingType))
	}
	if is_type_integer(src) && is_type_rune(dst) {
		return true
	}
	if is_type_rune(src) && is_type_integer(dst) {
		return true
	}
	if is_type_complex(src) && is_type_complex(dst) {
		return true
	}
	if is_type_float(src) && is_type_complex(dst) {
		return true
	}
	if is_type_float(src) && is_type_quaternion(dst) {
		return true
	}
	if is_type_complex(src) && is_type_quaternion(dst) {
		return true
	}
	if is_type_quaternion(src) && is_type_quaternion(dst) {
		return true
	}
	if is_type_matrix(src) && is_type_matrix(dst) {
		op := *operand
		op.Type = src.Matrix.elem
		if !check_is_castable_to(c, &op, dst.Matrix.elem) {
			return false
		}
		if src.Matrix.row_count != src.Matrix.column_count {
			src_count := src.Matrix.row_count * src.Matrix.column_count
			dst_count := dst.Matrix.row_count * dst.Matrix.column_count
			return src_count == dst_count
		}
		return is_matrix_square(dst) && is_matrix_square(src)
	}
	if is_type_pointer(src) && is_type_pointer(dst) {
		return true
	}
	if is_type_multi_pointer(src) && is_type_multi_pointer(dst) {
		return true
	}
	if is_type_multi_pointer(src) && is_type_pointer(dst) {
		return true
	}
	if is_type_pointer(src) && is_type_multi_pointer(dst) {
		return true
	}
	if is_type_uintptr(src) && is_type_pointer(dst) {
		return true
	}
	if is_type_pointer(src) && is_type_uintptr(dst) {
		return true
	}
	if is_type_uintptr(src) && is_type_multi_pointer(dst) {
		return true
	}
	if is_type_multi_pointer(src) && is_type_uintptr(dst) {
		return true
	}
	if is_type_u8_slice(src) && (is_type_string(dst) && !is_type_cstring(dst)) {
		return true
	}
	if is_type_u16_slice(src) && (is_type_string16(dst) && !is_type_cstring16(dst)) {
		return true
	}
	if are_types_identical(src, tCstring) && are_types_identical(dst, tString) {
		if operand.Mode != Addressing_Constant {
			add_package_dependency(c, "runtime", "cstring_to_string", true)
		}
		return true
	}
	if are_types_identical(src, tCstring16) && are_types_identical(dst, tString16) {
		if operand.Mode != Addressing_Constant {
			add_package_dependency(c, "runtime", "cstring16_to_string16", true)
		}
		return true
	}
	if are_types_identical(src, tCstring) && is_type_u8_ptr(dst) {
		return !is_constant
	}
	if are_types_identical(src, tCstring) && is_type_u8_multi_ptr(dst) {
		return !is_constant
	}
	if are_types_identical(src, tCstring) && is_type_rawptr(dst) {
		return !is_constant
	}
	if is_type_u8_ptr(src) && are_types_identical(dst, tCstring) {
		return !is_constant
	}
	if is_type_u8_multi_ptr(src) && are_types_identical(dst, tCstring) {
		return !is_constant
	}
	if is_type_rawptr(src) && are_types_identical(dst, tCstring) {
		return !is_constant
	}
	if are_types_identical(src, tCstring16) && is_type_u16_ptr(dst) {
		return !is_constant
	}
	if are_types_identical(src, tCstring16) && is_type_u16_multi_ptr(dst) {
		return !is_constant
	}
	if are_types_identical(src, tCstring16) && is_type_rawptr(dst) {
		return !is_constant
	}
	if is_type_u16_ptr(src) && are_types_identical(dst, tCstring16) {
		return !is_constant
	}
	if is_type_u16_multi_ptr(src) && are_types_identical(dst, tCstring16) {
		return !is_constant
	}
	if is_type_rawptr(src) && are_types_identical(dst, tCstring16) {
		return !is_constant
	}
	if is_type_proc(src) && is_type_proc(dst) {
		if is_type_polymorphic(dst) {
			if is_type_polymorphic(src) &&
				operand.Mode == Addressing_Variable {
				return true
			}
			return false
		}
		return true
	}
	if is_type_proc(src) && is_type_rawptr(dst) {
		return true
	}
	if is_type_rawptr(src) && is_type_proc(dst) {
		return true
	}
	if is_type_array(dst) {
		elem := base_array_type(dst)
		if check_is_castable_to(c, operand, elem) {
			return true
		}
	}
	if is_type_simd_vector(src) && is_type_simd_vector(dst) {
		if src.SimdVector.count != dst.SimdVector.count {
			return false
		}
		elem_src := base_array_type(src)
		elem_dst := base_array_type(dst)
		var x Operand
		x.Type = elem_src
		x.Mode = Addressing_Value
		return check_is_castable_to(c, &x, elem_dst)
	}
	if is_type_simd_vector(dst) {
		elem := base_array_type(dst)
		if check_is_castable_to(c, operand, elem) {
			return true
		}
	}
	return false
}

func check_cast_internal(c *CheckerContext, x *Operand, type_ *Type) bool {
	is_const_expr := x.Mode == Addressing_Constant
	bt := base_type(type_)
	if is_const_expr && is_type_constant_type(bt) {
		elem := core_array_type(bt)
		if core_type(bt).Kind == Type_Basic {
			return check_representable_as_constant(c, x.Value, type_, &x.Value) ||
				(is_type_pointer(type_) && check_is_castable_to(c, x, type_))
		} else if !are_types_identical(elem, bt) && elem.Kind == Type_Basic && x.Type.Kind == Type_Basic {
			return check_representable_as_constant(c, x.Value, elem, &x.Value) ||
				(is_type_pointer(elem) && check_is_castable_to(c, x, elem))
		} else if check_is_castable_to(c, x, type_) {
			x.Value = EmptyExactValue
			x.Mode = Addressing_Value
			return true
		}
	} else if check_is_castable_to(c, x, type_) {
		if x.Mode != Addressing_Constant {
			x.Mode = Addressing_Value
		} else if is_type_slice(type_) && is_type_string(x.Type) {
			x.Mode = Addressing_Value
		} else if is_type_union(type_) {
			if is_type_union_constantable(type_) {
				return true
			}
			x.Mode = Addressing_Value
		}
		if x.Mode == Addressing_Value {
			x.Value = EmptyExactValue
		}
		return true
	}
	return false
}

func check_cast(c *CheckerContext, x *Operand, type_ *Type, forbid_identical ...bool) {
	forbid := len(forbid_identical) > 0 && forbid_identical[0]

	if !is_operand_value(*x) {
		error(x.Expr, "Only values can be casted")
		x.Mode = Addressing_Invalid
		return
	}
	is_const_expr := x.Mode == Addressing_Constant
	can_convert := check_cast_internal(c, x, type_)
	if !can_convert {
		expr_str := expr_to_string(x.Expr)
		to_type := type_to_string(type_)
		from_type := type_to_string(x.Type)
		x.Mode = Addressing_Invalid
		begin_error_block()
		error(x.Expr, "Cannot cast '%s' as '%s' from '%s'", expr_str, to_type, from_type)
		if is_const_expr {
			val_str := exact_value_to_string(x.Value)
			if is_type_float(x.Type) && is_type_integer(type_) {
				error_line("\t%s cannot be represented without truncation/rounding as the type '%s'\n", val_str, to_type)
				x.Mode = Addressing_Constant
				x.Type = type_
			} else {
				error_line("\t'%s' cannot be represented as the type '%s'\n", val_str, to_type)
				if is_type_numeric(type_) {
					x.Mode = Addressing_Constant
					x.Type = type_
				}
			}
			gb_string_free(val_str)
		}
		end_error_block()
		check_cast_error_suggestion(c, x, type_)
		gb_string_free(expr_str)
		gb_string_free(to_type)
		gb_string_free(from_type)
		return
	}
	if is_type_untyped(x.Type) {
		final_type := type_
		if is_const_expr && !is_type_constant_type(type_) {
			if is_type_union(type_) {
				convert_to_typed(c, x, type_)
			}
			final_type = default_type(x.Type)
		}
		update_untyped_expr_type(c, x.Expr, final_type, true)
	} else {
		src := core_type(x.Type)
		dst := core_type(type_)
		if src != dst {
			const REQUIRE = true
			if is_type_integer_128bit(src) && is_type_float(dst) {
				add_package_dependency(c, "runtime", "floattidf_unsigned", REQUIRE)
				add_package_dependency(c, "runtime", "floattidf", REQUIRE)
			} else if is_type_integer_128bit(dst) && is_type_float(src) {
				add_package_dependency(c, "runtime", "fixunsdfti", REQUIRE)
				add_package_dependency(c, "runtime", "fixunsdfdi", REQUIRE)
			} else if src == tF16 && is_type_float(dst) {
				add_package_dependency(c, "runtime", "gnu_h2f_ieee", REQUIRE)
				add_package_dependency(c, "runtime", "extendhfsf2", REQUIRE)
			} else if is_type_float(dst) && dst == tF16 {
				add_package_dependency(c, "runtime", "truncsfhf2", REQUIRE)
				add_package_dependency(c, "runtime", "truncdfhf2", REQUIRE)
				add_package_dependency(c, "runtime", "gnu_f2h_ieee", REQUIRE)
			}
		}
		if forbid && (check_vet_flags(c)&VetFlag_Cast) != 0 &&
			(c.CurrProcSig == nil || !is_type_polymorphic(c.CurrProcSig)) {
			src_exact := x.Type
			dst_exact := type_
			if src_exact != nil &&
				dst_exact != nil &&
				are_types_identical(src_exact, dst_exact) {
				oper_str := expr_to_string(x.Expr)
				to_type := type_to_string(dst_exact)
				error(x.Expr, "Unneeded cast of '%s' to identical type '%s'", oper_str, to_type)
				gb_string_free(oper_str)
				gb_string_free(to_type)
			}
		}
	}
	if is_const_expr {
		src := core_type(x.Type)
		dst := core_type(type_)
		if is_type_string(src) && is_type_string(dst) {
			src_utf16 := is_type_string16(src) || is_type_cstring16(src)
			dst_utf16 := is_type_string16(dst) || is_type_cstring16(dst)
			if !src_utf16 && dst_utf16 {
				x.Value = exact_value_string16(string_to_string16(permanent_allocator(), x.Value.ValueString))
			}
			if src_utf16 && !dst_utf16 {
				x.Value = exact_value_string(string16_to_string(permanent_allocator(), x.Value.ValueString16))
			}
		}
	}
	x.Type = type_
}

func check_transmute(c *CheckerContext, node *Ast, o *Operand, t *Type, forbid_identical ...bool) bool {
	forbid := len(forbid_identical) > 0 && forbid_identical[0]

	if !is_operand_value(*o) {
		error(o.Expr, "'transmute' can only be applied to values")
		o.Mode = Addressing_Invalid
		return false
	}
	src := *o
	src_t := o.Type
	dst_t := t
	src_bt := base_type(src_t)
	dst_bt := base_type(dst_t)
	if is_type_untyped(src_t) {
		expr_str := expr_to_string(o.Expr)
		error(o.Expr, "Cannot transmute untyped expression: '%s'", expr_str)
		gb_string_free(expr_str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return false
	}
	if dst_bt == nil || dst_bt == tInvalid {
		o.Mode = Addressing_Invalid
		o.Expr = node
		return false
	}
	if src_bt == nil || src_bt == tInvalid {
		o.Mode = Addressing_Value
		o.Expr = node
		o.Type = dst_t
		return true
	}
	srcz := type_size_of(src_t)
	dstz := type_size_of(dst_t)
	if srcz != dstz {
		expr_str := expr_to_string(o.Expr)
		type_str := type_to_string(dst_t)
		error(o.Expr, "Cannot transmute '%s' to '%s', %d vs %d bytes", expr_str, type_str, srcz, dstz)
		gb_string_free(type_str)
		gb_string_free(expr_str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return false
	}
	o.Expr = node
	o.Type = dst_t
	if o.Mode == Addressing_Constant {
		if are_types_identical(src_bt, dst_bt) {
			return true
		}
		if (is_type_integer(src_t) && is_type_integer(dst_t)) ||
			(is_type_integer(src_t) && is_type_bit_set(dst_t)) {
			if types_have_same_internal_endian(src_t, dst_t) {
				src_v := exact_value_to_integer(o.Value)
				v := src_v.ValueInteger
				var smax BigInt
				var umax BigInt
				big_int_from_u64(&smax, 0)
				big_int_not(&smax, &smax, int32(srcz*8-1), false)
				big_int_from_u64(&umax, 1)
				sz_in_bits := big_int_make_i64(srcz * 8)
				big_int_shl(&umax, &umax, &sz_in_bits)
				if is_type_unsigned(src_t) && !is_type_unsigned(dst_t) {
					if big_int_cmp(&v, &smax) >= 0 {
						big_int_sub(&v, &v, &umax)
					}
				} else if !is_type_unsigned(src_t) && is_type_unsigned(dst_t) {
					if big_int_is_neg(&v) {
						big_int_add(&v, &v, &umax)
					}
				}
				o.Value.Kind = ExactValue_Integer
				o.Value.ValueInteger = v
				return true
			}
		}
	} else {
		if forbid && (check_vet_flags(c)&VetFlag_Cast) != 0 &&
			(c.CurrProcSig == nil || !is_type_polymorphic(c.CurrProcSig)) &&
			check_is_castable_to(c, &src, dst_t) {
			if are_types_identical(src_t, dst_t) {
				oper_str := expr_to_string(o.Expr)
				to_type := type_to_string(dst_t)
				error(o.Expr, "Unneeded transmute of '%s' to identical type '%s'", oper_str, to_type)
				gb_string_free(oper_str)
				gb_string_free(to_type)
			} else if is_type_internally_pointer_like(src_t) &&
				is_type_internally_pointer_like(dst_t) {
				error(o.Expr, "Use of 'transmute' where 'cast' would be preferred since the types are pointer-like")
			} else if are_types_identical(src_bt, dst_bt) {
				oper_str := expr_to_string(o.Expr)
				to_type := type_to_string(dst_t)
				error(o.Expr, "Unneeded transmute of '%s' to identical type '%s'", oper_str, to_type)
				gb_string_free(oper_str)
				gb_string_free(to_type)
			} else if is_type_integer(src_t) && is_type_integer(dst_t) &&
				types_have_same_internal_endian(src_t, dst_t) &&
				type_endian_kind_of(src_t) == type_endian_kind_of(dst_t) {
				oper_type := type_to_string(src_t)
				to_type := type_to_string(dst_t)
				error(o.Expr, "Use of 'transmute' where 'cast' would be preferred since both are integers of the same endianness, from '%s' to '%s'", oper_type, to_type)
				gb_string_free(to_type)
				gb_string_free(oper_type)
			}
		}
	}
	o.Mode = Addressing_Value
	o.Value = EmptyExactValue
	return true
}

func check_for_integer_division_by_zero(c *CheckerContext, node *Ast) IntegerDivisionByZeroKind {
	flags := check_feature_flags(c, node)
	if (flags & OptInFeatureFlagIntegerDivisionByZeroTrap) != 0 {
		return IntegerDivisionByZero_Trap
	}
	if (flags & OptInFeatureFlagIntegerDivisionByZeroZero) != 0 {
		return IntegerDivisionByZero_Zero
	}
	if (flags & OptInFeatureFlagIntegerDivisionByZeroSelf) != 0 {
		return IntegerDivisionByZero_Self
	}
	if (flags & OptInFeatureFlagIntegerDivisionByZeroAllBits) != 0 {
		return IntegerDivisionByZero_AllBits
	}
	return buildContext.IntegerDivisionByZeroBehaviour
}

func is_exact_value_zero(v ExactValue) bool {
	switch v.Kind {
	case ExactValue_Invalid:
		return true
	case ExactValue_Bool:
		return !v.ValueBool
	case ExactValue_String:
		return v.ValueString.Len == 0
	case ExactValue_String16:
		return v.ValueString16.Len == 0
	case ExactValue_Integer:
		return big_int_is_zero(&v.ValueInteger)
	case ExactValue_Float:
		return v.ValueFloat == 0.0
	case ExactValue_Complex:
		if v.ValueComplex != nil {
			return v.ValueComplex.Real == 0.0 && v.ValueComplex.Imag == 0.0
		}
		return true
	case ExactValue_Quaternion:
		if v.ValueQuaternion != nil {
			return v.ValueQuaternion.Real == 0.0 &&
				v.ValueQuaternion.Imag == 0.0 &&
				v.ValueQuaternion.Jmag == 0.0 &&
				v.ValueQuaternion.Kmag == 0.0
		}
		return true
	case ExactValue_Pointer:
		return v.ValuePointer == 0
	case ExactValue_Compound:
		if v.ValueCompound == nil {
			return true
		} else {
			cl := &v.ValueCompound.CompoundLit
			if len(cl.Elems) == 0 {
				return true
			} else {
				for _, elem := range cl.Elems {
					if elem.TAV.Mode != Addressing_Constant {
						return false
					}
					if !is_exact_value_zero(elem.TAV.Value) {
						return false
					}
				}
				return true
			}
		}
	case ExactValue_Procedure:
		return v.ValueProcedure == nil
	case ExactValue_Typeid:
		return v.ValueTypeid == nil
	}
	return true
}

func compare_exact_values_compound_lit(op TokenKind, x ExactValue, y ExactValue) bool {
	x_cl := &x.ValueCompound.CompoundLit
	y_cl := &y.ValueCompound.CompoundLit
	if len(x_cl.Elems) != len(y_cl.Elems) {
		return false
	}
	test := op == Token_CmpEq
	for i := 0; i < len(x_cl.Elems); i++ {
		lhs := x_cl.Elems[i]
		rhs := y_cl.Elems[i]
		if compare_exact_values(op, lhs.TAV.Value, rhs.TAV.Value) != test {
			return !test
		}
	}
	return test
}
