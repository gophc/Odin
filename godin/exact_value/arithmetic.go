package exact_value

import (
	"math"

	"github.com/gophc/Odin/godin/big_int"
)

// BinaryOperator applies a binary operator to two exact values
func BinaryOperator(op tokenKind, x, y ExactValue) ExactValue {
	matchExactValues(&x, &y)

	switch x.Kind {
	case ExactValue_Invalid:
		return x

	case ExactValue_Bool:
		switch op {
		case tokenCmpAnd:
			return NewBool(x.valueBool && y.valueBool)
		case tokenCmpOr:
			return NewBool(x.valueBool || y.valueBool)
		case tokenAnd:
			return NewBool(x.valueBool && y.valueBool)
		case tokenOr:
			return NewBool(x.valueBool || y.valueBool)
		case tokenAndNot:
			return NewBool(x.valueBool && !y.valueBool)
		case tokenXor:
			return NewBool((x.valueBool && !y.valueBool) || (!x.valueBool && y.valueBool))
		}

	case ExactValue_Integer:
		a := &x.valueInteger
		b := &y.valueInteger
		var c big_int.BigInt

		switch op {
		case tokenAdd:
			c = *big_int.Add(a, b)
		case tokenSub:
			c = *big_int.Sub(a, b)
		case tokenMul:
			c = *big_int.Mul(a, b)
		case tokenQuo:
			// Float division for integers
			return NewFloat(math.Mod(big_int.ToFloat64(a), big_int.ToFloat64(b)))
		case tokenQuoEq:
			c = *big_int.Quo(a, b)
		case tokenMod:
			c = *big_int.Rem(a, b)
		case tokenModMod:
			c = *big_int.EuclidMod(a, b)
		case tokenAnd:
			c = *big_int.And(a, b)
		case tokenOr:
			c = *big_int.Or(a, b)
		case tokenXor:
			c = *big_int.Xor(a, b)
		case tokenAndNot:
			c = *big_int.AndNot(a, b)
		case tokenShl:
			// y is a BigInt, but shift amount should be uint
			shift := big_int.ToU64(b)
			c = *big_int.Shl(a, uint32(shift))
		case tokenShr:
			shift := big_int.ToU64(b)
			c = *big_int.Shr(a, uint32(shift))
		default:
			return NewInvalid()
		}
		return NewInteger(&c)

	case ExactValue_Float:
		a := x.valueFloat
		b := y.valueFloat
		switch op {
		case tokenAdd:
			return NewFloat(a + b)
		case tokenSub:
			return NewFloat(a - b)
		case tokenMul:
			return NewFloat(a * b)
		case tokenQuo:
			return NewFloat(a / b)
		}

	case ExactValue_Complex:
		y = toComplex(y)
		if x.valueComplex == nil || y.valueComplex == nil {
			return NewInvalid()
		}
		a, b := x.valueComplex.Real, x.valueComplex.Imag
		c, d := y.valueComplex.Real, y.valueComplex.Imag
		var real, imag float64

		switch op {
		case tokenAdd:
			real, imag = a+c, b+d
		case tokenSub:
			real, imag = a-c, b-d
		case tokenMul:
			real = a*c - b*d
			imag = b*c + a*d
		case tokenQuo:
			s := c*c + d*d
			if s == 0 {
				return NewInvalid()
			}
			real = (a*c + b*d) / s
			imag = (b*c - a*d) / s
		default:
			return NewInvalid()
		}
		return NewComplex(real, imag)

	case ExactValue_Quaternion:
		y = toQuaternion(y)
		if x.valueQuaternion == nil || y.valueQuaternion == nil {
			return NewInvalid()
		}
		xr, xi, xj, xk := x.valueQuaternion.Real, x.valueQuaternion.Imag, x.valueQuaternion.Jmag, x.valueQuaternion.Kmag
		yr, yi, yj, yk := y.valueQuaternion.Real, y.valueQuaternion.Imag, y.valueQuaternion.Jmag, y.valueQuaternion.Kmag
		var real, imag, jmag, kmag float64

		switch op {
		case tokenAdd:
			real, imag, jmag, kmag = xr+yr, xi+yi, xj+yj, xk+yk
		case tokenSub:
			real, imag, jmag, kmag = xr-yr, xi-yi, xj-yj, xk-yk
		case tokenMul:
			imag = xr*yi + xi*yr + xj*yk - xk*yj
			jmag = xr*yj - xi*yk + xj*yr + xk*yi
			kmag = xr*yk + xi*yj - xj*yi + xk*yr
			real = xr*yr - xi*yi - xj*yj - xk*yk
		case tokenQuo:
			invMag2 := 1.0 / (yr*yr + yi*yi + yj*yj + yk*yk)
			imag = (xr*-yi + xi*yr + xj*-yk - xk*-yj) * invMag2
			jmag = (xr*-yj - xi*-yk + xj*yr + xk*-yi) * invMag2
			kmag = (xr*-yk + xi*-yj - xj*-yi + xk*yr) * invMag2
			real = (xr*yr - xi*-yi - xj*-yj - xk*-yk) * invMag2
		default:
			return NewInvalid()
		}
		return NewQuaternion(real, imag, jmag, kmag)

	case ExactValue_String:
		if op != tokenAdd {
			return NewInvalid()
		}
		return NewString(string(x.valueString.Text) + string(y.valueString.Text))

	case ExactValue_String16:
		if op != tokenAdd {
			return NewInvalid()
		}
		combined := make([]uint16, len(x.valueString16.Text)+len(y.valueString16.Text))
		copy(combined, x.valueString16.Text)
		copy(combined[len(x.valueString16.Text):], y.valueString16.Text)
		return NewString16(combined)
	}

	return NewInvalid()
}

// Add returns x + y
func Add(x, y ExactValue) ExactValue {
	return BinaryOperator(tokenAdd, x, y)
}

// Sub returns x - y
func Sub(x, y ExactValue) ExactValue {
	return BinaryOperator(tokenSub, x, y)
}

// Mul returns x * y
func Mul(x, y ExactValue) ExactValue {
	return BinaryOperator(tokenMul, x, y)
}

// Quo returns x / y
func Quo(x, y ExactValue) ExactValue {
	return BinaryOperator(tokenQuo, x, y)
}

// Shift returns x << y or x >> y
func Shift(op tokenKind, x, y ExactValue) ExactValue {
	return BinaryOperator(op, x, y)
}
