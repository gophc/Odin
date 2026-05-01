package exact_value

import (
	"github.com/gophc/Odin/godin/big_int"
)

// tokenKind represents token types for operators (simplified)
type tokenKind int

const (
	tokenAdd tokenKind = iota
	tokenSub
	tokenMul
	tokenQuo
	tokenQuoEq // integer division
	tokenMod
	tokenModMod // Euclidean mod
	tokenAnd
	tokenOr
	tokenXor
	tokenAndNot
	tokenShl
	tokenShr
	tokenCmpAnd
	tokenCmpOr
	tokenNot
	tokenCmpEq
	tokenNotEq
	tokenLt
	tokenLtEq
	tokenGt
	tokenGtEq
)

func order(v ExactValue) int {
	switch v.Kind {
	case ExactValue_Invalid, ExactValue_Compound:
		return 0
	case ExactValue_Bool, ExactValue_String, ExactValue_String16:
		return 1
	case ExactValue_Integer:
		return 2
	case ExactValue_Float:
		return 3
	case ExactValue_Complex:
		return 4
	case ExactValue_Quaternion:
		return 5
	case ExactValue_Pointer:
		return 6
	case ExactValue_Procedure:
		return 7
	case ExactValue_Typeid:
		return 8
	default:
		return -1
	}
}

func matchExactValues(x, y *ExactValue) {
	if order(*y) < order(*x) {
		matchExactValues(y, x)
		return
	}

	switch x.Kind {
	case ExactValue_Invalid:
		*y = *x
		return
	case ExactValue_Bool, ExactValue_String, ExactValue_String16,
		ExactValue_Quaternion, ExactValue_Pointer, ExactValue_Compound,
		ExactValue_Procedure, ExactValue_Typeid:
		return
	case ExactValue_Integer:
		switch y.Kind {
		case ExactValue_Integer:
			return
		case ExactValue_Float:
			*x = NewFloat(big_int.ToFloat64(&x.valueInteger))
			return
		case ExactValue_Complex:
			*x = NewComplex(big_int.ToFloat64(&x.valueInteger), 0)
			return
		case ExactValue_Quaternion:
			*x = NewQuaternion(big_int.ToFloat64(&x.valueInteger), 0, 0, 0)
			return
		}
	case ExactValue_Float:
		switch y.Kind {
		case ExactValue_Float:
			return
		case ExactValue_Complex:
			*x = toComplex(*x)
			return
		case ExactValue_Quaternion:
			*x = toQuaternion(*x)
			return
		}
	case ExactValue_Complex:
		switch y.Kind {
		case ExactValue_Complex:
			return
		case ExactValue_Quaternion:
			*x = toQuaternion(*x)
			return
		}
	}
}

func UnaryOperator(op tokenKind, v ExactValue, precision int, isUnsigned bool) ExactValue {
	switch op {
	case tokenAdd:
		switch v.Kind {
		case ExactValue_Invalid, ExactValue_Integer, ExactValue_Float,
			ExactValue_Complex, ExactValue_Quaternion:
			return v
		}
	case tokenSub:
		switch v.Kind {
		case ExactValue_Invalid:
			return v
		case ExactValue_Integer:
			neg := big_int.Neg(&v.valueInteger)
			return NewInteger(neg)
		case ExactValue_Float:
			return NewFloat(-v.valueFloat)
		case ExactValue_Complex:
			if v.valueComplex != nil {
				return NewComplex(-v.valueComplex.Real, -v.valueComplex.Imag)
			}
		case ExactValue_Quaternion:
			if v.valueQuaternion != nil {
				return NewQuaternion(
					-v.valueQuaternion.Real,
					-v.valueQuaternion.Imag,
					-v.valueQuaternion.Jmag,
					-v.valueQuaternion.Kmag,
				)
			}
		}
	case tokenXor:
		switch v.Kind {
		case ExactValue_Invalid:
			return v
		case ExactValue_Integer:
			if precision == 0 {
				return NewInvalid()
			}
			result := big_int.Not(&v.valueInteger, precision, !isUnsigned)
			return NewInteger(result)
		}
	case tokenNot:
		switch v.Kind {
		case ExactValue_Invalid:
			return v
		case ExactValue_Bool:
			return NewBool(!v.valueBool)
		}
	}
	return NewInvalid()
}

func toInteger(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Bool:
		if v.valueBool {
			return NewIntegerFromI64(1)
		}
		return NewIntegerFromI64(0)
	case ExactValue_Integer:
		return v
	case ExactValue_Float:
		i := int64(v.valueFloat)
		if float64(i) == v.valueFloat {
			return NewIntegerFromI64(i)
		}
	case ExactValue_Pointer:
		return NewIntegerFromI64(v.valuePointer)
	}
	return NewInvalid()
}

func toFloat(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer:
		return NewFloat(big_int.ToFloat64(&v.valueInteger))
	case ExactValue_Float:
		return v
	}
	return NewInvalid()
}

func toComplex(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer:
		return NewComplex(big_int.ToFloat64(&v.valueInteger), 0)
	case ExactValue_Float:
		return NewComplex(v.valueFloat, 0)
	case ExactValue_Complex:
		return v
	}
	return NewInvalid()
}

func toQuaternion(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer:
		return NewQuaternion(big_int.ToFloat64(&v.valueInteger), 0, 0, 0)
	case ExactValue_Float:
		return NewQuaternion(v.valueFloat, 0, 0, 0)
	case ExactValue_Complex:
		if v.valueComplex != nil {
			return NewQuaternion(v.valueComplex.Real, v.valueComplex.Imag, 0, 0)
		}
	case ExactValue_Quaternion:
		return v
	}
	return NewInvalid()
}

func Real(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer, ExactValue_Float:
		return v
	case ExactValue_Complex:
		if v.valueComplex != nil {
			return NewFloat(v.valueComplex.Real)
		}
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			return NewFloat(v.valueQuaternion.Real)
		}
	}
	return NewInvalid()
}

func Imag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer, ExactValue_Float:
		return NewIntegerFromI64(0)
	case ExactValue_Complex:
		if v.valueComplex != nil {
			return NewFloat(v.valueComplex.Imag)
		}
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			return NewFloat(v.valueQuaternion.Imag)
		}
	}
	return NewInvalid()
}

func Jmag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer, ExactValue_Float, ExactValue_Complex:
		return NewIntegerFromI64(0)
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			return NewFloat(v.valueQuaternion.Jmag)
		}
	}
	return NewInvalid()
}

func Kmag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValue_Integer, ExactValue_Float, ExactValue_Complex:
		return NewIntegerFromI64(0)
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			return NewFloat(v.valueQuaternion.Kmag)
		}
	}
	return NewInvalid()
}

func ToI64(v ExactValue) int64 {
	v = toInteger(v)
	if v.Kind == ExactValue_Integer {
		return big_int.ToI64(&v.valueInteger)
	}
	return 0
}

func ToU64(v ExactValue) uint64 {
	v = toInteger(v)
	if v.Kind == ExactValue_Integer {
		return big_int.ToU64(&v.valueInteger)
	}
	return 0
}

func ToF64(v ExactValue) float64 {
	v = toFloat(v)
	if v.Kind == ExactValue_Float {
		return v.valueFloat
	}
	return 0.0
}
