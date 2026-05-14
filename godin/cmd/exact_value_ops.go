// Rewrite of src/cipp/done/exact_value.i.cpp (lines 411-911).
// Arithmetic operations, comparison, and string output for ExactValue.
package cmd

import (
	"math"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Unary operator evaluation
// ---------------------------------------------------------------------------

func exact_unary_operator_value(op TokenKind, v ExactValue, precision int32, is_unsigned bool) ExactValue {
	switch op {
	case Token_Add:
		switch v.Kind {
		case ExactValueInvalid, ExactValueInteger, ExactValueFloat, ExactValueComplex, ExactValueQuaternion:
			return v
		}
	case Token_Sub:
		switch v.Kind {
		case ExactValueInvalid:
			return v
		case ExactValueInteger:
			i := ExactValue{Kind: ExactValueInteger}
			big_int_neg(&i.ValueInteger, &v.ValueInteger)
			return i
		case ExactValueFloat:
			i := v
			i.ValueFloat = -i.ValueFloat
			return i
		case ExactValueComplex:
			if v.ValueComplex != nil {
				return exact_value_complex(-v.ValueComplex.Real, -v.ValueComplex.Imag)
			}
		case ExactValueQuaternion:
			if v.ValueQuaternion != nil {
				q := v.ValueQuaternion
				return exact_value_quaternion(-q.Real, -q.Imag, -q.Jmag, -q.Kmag)
			}
		}
	case Token_Xor:
		switch v.Kind {
		case ExactValueInvalid:
			return v
		case ExactValueInteger:
			if precision == 0 {
				gb_assert_handler("Assertion Failure", "precision != 0", "", 0)
			}
			i := ExactValue{Kind: ExactValueInteger}
			big_int_not(&i.ValueInteger, &v.ValueInteger, precision, !is_unsigned)
			return i
		default:
			return EmptyExactValue
		}
	case Token_Not:
		switch v.Kind {
		case ExactValueInvalid:
			return v
		case ExactValueBool:
			return exact_value_bool(!v.ValueBool)
		}
	}
	return EmptyExactValue
}

// ---------------------------------------------------------------------------
// Value ordering (for match_exact_values promotion lattice)
// ---------------------------------------------------------------------------

func exact_value_order(v ExactValue) int32 {
	switch v.Kind {
	case ExactValueInvalid, ExactValueCompound:
		return 0
	case ExactValueBool, ExactValueString, ExactValueString16:
		return 1
	case ExactValueInteger:
		return 2
	case ExactValueFloat:
		return 3
	case ExactValueComplex:
		return 4
	case ExactValueQuaternion:
		return 5
	case ExactValuePointer:
		return 6
	case ExactValueProcedure:
		return 7
	default:
		compiler_error("How'd you get here? Invalid Value.kind %d", v.Kind)
		return -1
	}
}

// ---------------------------------------------------------------------------
// match_exact_values — promote x to y's type level for binary ops
// ---------------------------------------------------------------------------

func match_exact_values(x, y *ExactValue) {
	if exact_value_order(*y) < exact_value_order(*x) {
		match_exact_values(y, x)
		return
	}
	switch x.Kind {
	case ExactValueInvalid:
		*y = *x
		return
	case ExactValueBool, ExactValueString, ExactValueString16, ExactValueQuaternion,
		ExactValuePointer, ExactValueCompound, ExactValueProcedure, ExactValueTypeid:
		return
	case ExactValueInteger:
		switch y.Kind {
		case ExactValueInteger:
			return
		case ExactValueFloat:
			*x = exact_value_float(big_int_to_f64(&x.ValueInteger))
			return
		case ExactValueComplex:
			*x = exact_value_complex(big_int_to_f64(&x.ValueInteger), 0)
			return
		case ExactValueQuaternion:
			*x = exact_value_quaternion(big_int_to_f64(&x.ValueInteger), 0, 0, 0)
			return
		}
	case ExactValueFloat:
		switch y.Kind {
		case ExactValueFloat:
			return
		case ExactValueComplex:
			*x = exact_value_to_complex(*x)
			return
		case ExactValueQuaternion:
			*x = exact_value_to_quaternion(*x)
			return
		}
	case ExactValueComplex:
		switch y.Kind {
		case ExactValueComplex:
			return
		case ExactValueQuaternion:
			*x = exact_value_to_quaternion(*x)
			return
		}
	}
	compiler_error("match_exact_values: How'd you get here? Invalid ExactValueKind %d", x.Kind)
}

// ---------------------------------------------------------------------------
// Binary operator evaluation
// ---------------------------------------------------------------------------

func exact_binary_operator_value(op TokenKind, x_in, y_in ExactValue) ExactValue {
	x, y := x_in, y_in
	match_exact_values(&x, &y)

	switch x.Kind {
	case ExactValueInvalid:
		return x

	case ExactValueBool:
		switch op {
		case Token_CmpAnd:
			return exact_value_bool(x.ValueBool && y.ValueBool)
		case Token_CmpOr:
			return exact_value_bool(x.ValueBool || y.ValueBool)
		case Token_And:
			return exact_value_bool(x.ValueBool && y.ValueBool)
		case Token_Or:
			return exact_value_bool(x.ValueBool || y.ValueBool)
		case Token_AndNot:
			return exact_value_bool(x.ValueBool && !y.ValueBool)
		case Token_Xor:
			return exact_value_bool((x.ValueBool && !y.ValueBool) || (!x.ValueBool && y.ValueBool))
		default:
			return EmptyExactValue
		}

	case ExactValueInteger:
		a := &x.ValueInteger
		b := &y.ValueInteger
		var c BigInt
		switch op {
		case Token_Add:
			big_int_add(&c, a, b)
		case Token_Sub:
			big_int_sub(&c, a, b)
		case Token_Mul:
			big_int_mul(&c, a, b)
		case Token_Quo:
			return exact_value_float(math.Mod(big_int_to_f64(a), big_int_to_f64(b)))
		case Token_QuoEq:
			big_int_quo(&c, a, b)
		case Token_Mod:
			big_int_rem(&c, a, b)
		case Token_ModMod:
			big_int_euclidean_mod(&c, a, b)
		case Token_And:
			big_int_and(&c, a, b)
		case Token_Or:
			big_int_or(&c, a, b)
		case Token_Xor:
			big_int_xor(&c, a, b)
		case Token_AndNot:
			big_int_and_not(&c, a, b)
		case Token_Shl:
			big_int_shl(&c, a, b)
		case Token_Shr:
			big_int_shr(&c, a, b)
		default:
			return EmptyExactValue
		}
		return ExactValue{Kind: ExactValueInteger, ValueInteger: c}

	case ExactValueFloat:
		a := x.ValueFloat
		b := y.ValueFloat
		switch op {
		case Token_Add:
			return exact_value_float(a + b)
		case Token_Sub:
			return exact_value_float(a - b)
		case Token_Mul:
			return exact_value_float(a * b)
		case Token_Quo:
			return exact_value_float(a / b)
		default:
			return EmptyExactValue
		}

	case ExactValueComplex:
		y = exact_value_to_complex(y)
		a := x.ValueComplex.Real
		b := x.ValueComplex.Imag
		c := y.ValueComplex.Real
		d := y.ValueComplex.Imag
		switch op {
		case Token_Add:
			return exact_value_complex(a+c, b+d)
		case Token_Sub:
			return exact_value_complex(a-c, b-d)
		case Token_Mul:
			return exact_value_complex(a*c-b*d, b*c+a*d)
		case Token_Quo:
			s := c*c + d*d
			return exact_value_complex((a*c+b*d)/s, (b*c-a*d)/s)
		default:
			return EmptyExactValue
		}

	case ExactValueQuaternion:
		y = exact_value_to_quaternion(y)
		xr, xi, xj, xk := x.ValueQuaternion.Real, x.ValueQuaternion.Imag, x.ValueQuaternion.Jmag, x.ValueQuaternion.Kmag
		yr, yi, yj, yk := y.ValueQuaternion.Real, y.ValueQuaternion.Imag, y.ValueQuaternion.Jmag, y.ValueQuaternion.Kmag
		switch op {
		case Token_Add:
			return exact_value_quaternion(xr+yr, xi+yi, xj+yj, xk+yk)
		case Token_Sub:
			return exact_value_quaternion(xr-yr, xi-yi, xj-yj, xk-yk)
		case Token_Mul:
			return exact_value_quaternion(
				xr*yr - xi*yi - xj*yj - xk*yk,
				xr*yi + xi*yr + xj*yk - xk*yj,
				xr*yj - xi*yk + xj*yr + xk*yi,
				xr*yk + xi*yj - xj*yi + xk*yr,
			)
		case Token_Quo:
			invmag2 := 1.0 / (yr*yr + yi*yi + yj*yj + yk*yk)
			return exact_value_quaternion(
				(xr*+yr - xi*-yi - xj*-yj - xk*-yk) * invmag2,
				(xr*-yi + xi*+yr + xj*-yk - xk*-yj) * invmag2,
				(xr*-yj - xi*-yk + xj*+yr + xk*-yi) * invmag2,
				(xr*-yk + xi*-yj - xj*-yi + xk*+yr) * invmag2,
			)
		default:
			return EmptyExactValue
		}

	case ExactValueString:
		if op != Token_Add {
			return EmptyExactValue
		}
		sx := x.ValueString
		sy := y.ValueString
		length := sx.Len + sy.Len
		data := (*byte)(gb_alloc(permanent_allocator(), uintptr(length)))
		gb_memmove(unsafe.Pointer(data), unsafe.Pointer(sx.Data), uintptr(sx.Len))
		gb_memmove(unsafe.Pointer(uintptr(unsafe.Pointer(data))+uintptr(sx.Len)), unsafe.Pointer(sy.Data), uintptr(sy.Len))
		return exact_value_string(make_string(data, length))

	case ExactValueString16:
		if op != Token_Add {
			return EmptyExactValue
		}
		sx := x.ValueString16
		sy := y.ValueString16
		length := sx.Len + sy.Len
		data := (*uint16)(gb_alloc(permanent_allocator(), uintptr(length)*2))
		gb_memmove(unsafe.Pointer(data), unsafe.Pointer(sx.Data), uintptr(sx.Len)*2)
		gb_memmove(unsafe.Pointer(uintptr(unsafe.Pointer(data))+uintptr(sx.Len)*2), unsafe.Pointer(sy.Data), uintptr(sy.Len)*2)
		return exact_value_string16(make_string16(data, length))
	}

	return EmptyExactValue
}

// ---------------------------------------------------------------------------
// Convenience wrappers
// ---------------------------------------------------------------------------

func exact_value_add(x, y ExactValue) ExactValue {
	return exact_binary_operator_value(Token_Add, x, y)
}

func exact_value_sub(x, y ExactValue) ExactValue {
	return exact_binary_operator_value(Token_Sub, x, y)
}

func exact_value_mul(x, y ExactValue) ExactValue {
	return exact_binary_operator_value(Token_Mul, x, y)
}

func exact_value_quo(x, y ExactValue) ExactValue {
	return exact_binary_operator_value(Token_Quo, x, y)
}

func exact_value_shift(op TokenKind, x, y ExactValue) ExactValue {
	return exact_binary_operator_value(op, x, y)
}

func exact_value_increment_one(x ExactValue) ExactValue {
	return exact_binary_operator_value(Token_Add, x, exact_value_i64(1))
}

// ---------------------------------------------------------------------------
// f64 comparison helper
// ---------------------------------------------------------------------------

func cmp_f64(a, b float64) int32 {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

// ---------------------------------------------------------------------------
// String / String16 content comparison helpers
// ---------------------------------------------------------------------------

func string_eq(a, b String) bool {
	if a.Len != b.Len {
		return false
	}
	if a.Data == b.Data {
		return true
	}
	for i := isize(0); i < a.Len; i++ {
		ca := *(*byte)(unsafe.Add(unsafe.Pointer(a.Data), i))
		cb := *(*byte)(unsafe.Add(unsafe.Pointer(b.Data), i))
		if ca != cb {
			return false
		}
	}
	return true
}

func string_compare(a, b String) int {
	na, nb := int(a.Len), int(b.Len)
	minLen := na
	if nb < minLen {
		minLen = nb
	}
	for i := 0; i < minLen; i++ {
		ca := *(*byte)(unsafe.Add(unsafe.Pointer(a.Data), i))
		cb := *(*byte)(unsafe.Add(unsafe.Pointer(b.Data), i))
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
	}
	if na < nb {
		return -1
	}
	if na > nb {
		return 1
	}
	return 0
}

func string16_eq(a, b String16) bool {
	if a.Len != b.Len {
		return false
	}
	if a.Data == b.Data {
		return true
	}
	for i := isize(0); i < a.Len; i++ {
		ca := *(*uint16)(unsafe.Add(unsafe.Pointer(a.Data), i*2))
		cb := *(*uint16)(unsafe.Add(unsafe.Pointer(b.Data), i*2))
		if ca != cb {
			return false
		}
	}
	return true
}

func string16_compare(a, b String16) int {
	na, nb := int(a.Len), int(b.Len)
	minLen := na
	if nb < minLen {
		minLen = nb
	}
	for i := 0; i < minLen; i++ {
		ca := *(*uint16)(unsafe.Add(unsafe.Pointer(a.Data), i*2))
		cb := *(*uint16)(unsafe.Add(unsafe.Pointer(b.Data), i*2))
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
	}
	if na < nb {
		return -1
	}
	if na > nb {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// compare_exact_values_compound_lit — forward declaration (defined elsewhere)
// ---------------------------------------------------------------------------

func compare_exact_values_compound_lit(op TokenKind, x, y ExactValue) bool

// ---------------------------------------------------------------------------
// Value comparison
// ---------------------------------------------------------------------------

func compare_exact_values(op TokenKind, x_in, y_in ExactValue) bool {
	x, y := x_in, y_in
	match_exact_values(&x, &y)

	switch x.Kind {
	case ExactValueInvalid:
		return false

	case ExactValueBool:
		switch op {
		case Token_CmpEq:
			return x.ValueBool == y.ValueBool
		case Token_NotEq:
			return x.ValueBool != y.ValueBool
		}

	case ExactValueInteger:
		cmp := big_int_cmp(&x.ValueInteger, &y.ValueInteger)
		switch op {
		case Token_CmpEq:
			return cmp == 0
		case Token_NotEq:
			return cmp != 0
		case Token_Lt:
			return cmp < 0
		case Token_LtEq:
			return cmp <= 0
		case Token_Gt:
			return cmp > 0
		case Token_GtEq:
			return cmp >= 0
		}

	case ExactValueFloat:
		a, b := x.ValueFloat, y.ValueFloat
		if math.IsNaN(a) || math.IsNaN(b) {
			return op == Token_NotEq
		}
		c := cmp_f64(a, b)
		switch op {
		case Token_CmpEq:
			return c == 0
		case Token_NotEq:
			return c != 0
		case Token_Lt:
			return c < 0
		case Token_LtEq:
			return c <= 0
		case Token_Gt:
			return c > 0
		case Token_GtEq:
			return c >= 0
		}

	case ExactValueComplex:
		if x.ValueComplex == nil || y.ValueComplex == nil {
			return false
		}
		a, b := x.ValueComplex.Real, x.ValueComplex.Imag
		c, d := y.ValueComplex.Real, y.ValueComplex.Imag
		switch op {
		case Token_CmpEq:
			return cmp_f64(a, c) == 0 && cmp_f64(b, d) == 0
		case Token_NotEq:
			return cmp_f64(a, c) != 0 || cmp_f64(b, d) != 0
		}

	case ExactValueString:
		a, b := x.ValueString, y.ValueString
		switch op {
		case Token_CmpEq:
			return string_eq(a, b)
		case Token_NotEq:
			return !string_eq(a, b)
		case Token_Lt:
			return string_compare(a, b) < 0
		case Token_LtEq:
			return string_compare(a, b) <= 0
		case Token_Gt:
			return string_compare(a, b) > 0
		case Token_GtEq:
			return string_compare(a, b) >= 0
		}

	case ExactValueString16:
		a, b := x.ValueString16, y.ValueString16
		switch op {
		case Token_CmpEq:
			return string16_eq(a, b)
		case Token_NotEq:
			return !string16_eq(a, b)
		case Token_Lt:
			return string16_compare(a, b) < 0
		case Token_LtEq:
			return string16_compare(a, b) <= 0
		case Token_Gt:
			return string16_compare(a, b) > 0
		case Token_GtEq:
			return string16_compare(a, b) >= 0
		}

	case ExactValuePointer:
		switch op {
		case Token_CmpEq:
			return x.ValuePointer == y.ValuePointer
		case Token_NotEq:
			return x.ValuePointer != y.ValuePointer
		case Token_Lt:
			return x.ValuePointer < y.ValuePointer
		case Token_LtEq:
			return x.ValuePointer <= y.ValuePointer
		case Token_Gt:
			return x.ValuePointer > y.ValuePointer
		case Token_GtEq:
			return x.ValuePointer >= y.ValuePointer
		}

	case ExactValueTypeid:
		switch op {
		case Token_CmpEq:
			return x.ValueTypeid == y.ValueTypeid
		case Token_NotEq:
			return x.ValueTypeid != y.ValueTypeid
		}

	case ExactValueProcedure:
		switch op {
		case Token_CmpEq:
			return x.ValueProcedure == y.ValueProcedure
		case Token_NotEq:
			return x.ValueProcedure != y.ValueProcedure
		}

	case ExactValueCompound:
		if op != Token_CmpEq && op != Token_NotEq {
			return false
		}
		if x.Kind != y.Kind {
			return false
		}
		return compare_exact_values_compound_lit(op, x, y)
	}

	compiler_error("Invalid comparison: %d", x.Kind)
	return false
}

// ---------------------------------------------------------------------------
// String conversion
// ---------------------------------------------------------------------------

func max_isize(a, b isize) isize {
	if a > b {
		return a
	}
	return b
}

func write_exact_value_to_string(str gbString, v ExactValue, string_limit ...isize) gbString {
	limit := isize(36)
	if len(string_limit) > 0 {
		limit = string_limit[0]
	}
	limit = max_isize(limit, 36)

	switch v.Kind {
	case ExactValueInvalid:
		return str

	case ExactValueBool:
		if v.ValueBool {
			return gb_string_appendc(str, "true")
		}
		return gb_string_appendc(str, "false")

	case ExactValueString:
		s := quote_to_ascii(heap_allocator(), v.ValueString)
		if s.Len <= limit {
			str = gb_string_append_length(str, s.Data, s.Len)
		} else {
			n := limit / 5
			if n < 1 {
				n = 1
			}
			str = gb_string_append_length(str, s.Data, n)
			str = gb_string_append_fmt(str, "\"..%d chars..\"", s.Len-2*n)
			str = gb_string_append_length(str, (*byte)(unsafe.Add(unsafe.Pointer(s.Data), s.Len-n)), n)
		}
		gb_free(heap_allocator(), unsafe.Pointer(s.Data))
		return str

	case ExactValueString16:
		s := quote_to_ascii(heap_allocator(), v.ValueString16)
		if s.Len <= limit {
			str = gb_string_append_length(str, s.Data, s.Len)
		} else {
			n := limit / 5
			if n < 1 {
				n = 1
			}
			str = gb_string_append_length(str, s.Data, n)
			str = gb_string_append_fmt(str, "\"..%d chars..\"", s.Len-2*n)
			str = gb_string_append_length(str, (*byte)(unsafe.Add(unsafe.Pointer(s.Data), s.Len-n)), n)
		}
		gb_free(heap_allocator(), unsafe.Pointer(s.Data))
		return str

	case ExactValueInteger:
		s := big_int_to_string(heap_allocator(), &v.ValueInteger)
		str = gb_string_append_length(str, s.Data, s.Len)
		gb_free(heap_allocator(), unsafe.Pointer(s.Data))
		return str

	case ExactValueFloat:
		return gb_string_append_fmt(str, "%f", v.ValueFloat)

	case ExactValueComplex:
		if v.ValueComplex != nil {
			return gb_string_append_fmt(str, "%f+%fi", v.ValueComplex.Real, v.ValueComplex.Imag)
		}
		return str

	case ExactValueQuaternion:
		if v.ValueQuaternion != nil {
			q := v.ValueQuaternion
			return gb_string_append_fmt(str, "%f+%fi+%fj+%fk", q.Real, q.Imag, q.Jmag, q.Kmag)
		}
		return str

	case ExactValuePointer:
		return str

	case ExactValueCompound:
		return write_expr_to_string(str, v.ValueCompound, false)

	case ExactValueProcedure:
		return write_expr_to_string(str, v.ValueProcedure, false)
	}

	return str
}

func exact_value_to_string(v ExactValue, string_limit ...isize) gbString {
	limit := isize(36)
	if len(string_limit) > 0 {
		limit = string_limit[0]
	}
	return write_exact_value_to_string(gb_string_make(heap_allocator(), ""), v, limit)
}
