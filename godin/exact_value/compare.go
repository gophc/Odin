package exact_value

import (
	"math"

	"github.com/gophc/Odin/godin/big_int"
)

// Compare compares two exact values with the given comparison operator.
func Compare(op tokenKind, x, y ExactValue) bool {
	matchExactValues(&x, &y)

	switch x.Kind {
	case ExactValue_Invalid:
		return false
	case ExactValue_Bool:
		switch op {
		case tokenCmpEq:
			return x.valueBool == y.valueBool
		case tokenNotEq:
			return x.valueBool != y.valueBool
		}
	case ExactValue_Integer:
		cmp := big_int.Cmp(&x.valueInteger, &y.valueInteger)
		switch op {
		case tokenCmpEq:
			return cmp == 0
		case tokenNotEq:
			return cmp != 0
		case tokenLt:
			return cmp < 0
		case tokenLtEq:
			return cmp <= 0
		case tokenGt:
			return cmp > 0
		case tokenGtEq:
			return cmp >= 0
		}
	case ExactValue_Float:
		a := x.valueFloat
		b := y.valueFloat
		if math.IsNaN(a) || math.IsNaN(b) {
			return op == tokenNotEq
		}
		switch op {
		case tokenCmpEq:
			return cmpF64(a, b) == 0
		case tokenNotEq:
			return cmpF64(a, b) != 0
		case tokenLt:
			return cmpF64(a, b) < 0
		case tokenLtEq:
			return cmpF64(a, b) <= 0
		case tokenGt:
			return cmpF64(a, b) > 0
		case tokenGtEq:
			return cmpF64(a, b) >= 0
		}
	case ExactValue_Complex:
		if x.valueComplex == nil || y.valueComplex == nil {
			return false
		}
		a, b := x.valueComplex.Real, x.valueComplex.Imag
		c, d := y.valueComplex.Real, y.valueComplex.Imag
		switch op {
		case tokenCmpEq:
			return cmpF64(a, c) == 0 && cmpF64(b, d) == 0
		case tokenNotEq:
			return cmpF64(a, c) != 0 || cmpF64(b, d) != 0
		}
	case ExactValue_Quaternion:
		if x.valueQuaternion == nil || y.valueQuaternion == nil {
			return false
		}
		switch op {
		case tokenCmpEq:
			return x.valueQuaternion.Real == y.valueQuaternion.Real &&
				x.valueQuaternion.Imag == y.valueQuaternion.Imag &&
				x.valueQuaternion.Jmag == y.valueQuaternion.Jmag &&
				x.valueQuaternion.Kmag == y.valueQuaternion.Kmag
		case tokenNotEq:
			return !(x.valueQuaternion.Real == y.valueQuaternion.Real &&
				x.valueQuaternion.Imag == y.valueQuaternion.Imag &&
				x.valueQuaternion.Jmag == y.valueQuaternion.Jmag &&
				x.valueQuaternion.Kmag == y.valueQuaternion.Kmag)
		}
	case ExactValue_String:
		a := string(x.valueString.Text)
		b := string(y.valueString.Text)
		switch op {
		case tokenCmpEq:
			return a == b
		case tokenNotEq:
			return a != b
		case tokenLt:
			return a < b
		case tokenLtEq:
			return a <= b
		case tokenGt:
			return a > b
		case tokenGtEq:
			return a >= b
		}
	case ExactValue_String16:
		a := x.valueString16.Text
		b := y.valueString16.Text
		cmp := compareUint16Slice(a, b)
		switch op {
		case tokenCmpEq:
			return cmp == 0
		case tokenNotEq:
			return cmp != 0
		case tokenLt:
			return cmp < 0
		case tokenLtEq:
			return cmp <= 0
		case tokenGt:
			return cmp > 0
		case tokenGtEq:
			return cmp >= 0
		}
	case ExactValue_Pointer:
		switch op {
		case tokenCmpEq:
			return x.valuePointer == y.valuePointer
		case tokenNotEq:
			return x.valuePointer != y.valuePointer
		case tokenLt:
			return x.valuePointer < y.valuePointer
		case tokenLtEq:
			return x.valuePointer <= y.valuePointer
		case tokenGt:
			return x.valuePointer > y.valuePointer
		case tokenGtEq:
			return x.valuePointer >= y.valuePointer
		}
	case ExactValue_Typeid:
		switch op {
		case tokenCmpEq:
			return x.valueTypeid == y.valueTypeid
		case tokenNotEq:
			return x.valueTypeid != y.valueTypeid
		}
	case ExactValue_Procedure:
		switch op {
		case tokenCmpEq:
			return x.valueProcedure == y.valueProcedure
		case tokenNotEq:
			return x.valueProcedure != y.valueProcedure
		}
	case ExactValue_Compound:
		if op != tokenCmpEq && op != tokenNotEq {
			return false
		}
		if x.Kind != y.Kind {
			return false
		}
		// For compound literals, use pointer equality as placeholder
		if op == tokenCmpEq {
			return x.valueCompound == y.valueCompound
		}
		return x.valueCompound != y.valueCompound
	}
	return false
}

func cmpF64(a, b float64) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

func compareUint16Slice(a, b []uint16) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}
