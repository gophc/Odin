package common

import "math"

// F32ToF16 converts a 32-bit float to a 16-bit half-float (binary16).
func F32ToF16(value f32) u16 {
	bits := math.Float32bits(float32(value))
	s := u16((bits >> 16) & 0x8000)
	exp := int32((bits>>23)&0xff) - 127 + 15
	mant := bits & 0x7fffff

	if exp <= 0 {
		if exp < -10 {
			return s
		}
		mant = (mant | 0x800000) >> uint32(1-exp)
		if mant&0x1000 != 0 {
			mant += 0x2000
		}
		return s | u16(mant>>13)
	}
	if exp >= 31 {
		return s | 0x7c00 | u16(mant>>13)
	}
	if mant&0x1000 != 0 {
		mant += 0x2000
		if mant&0x800000 != 0 {
			mant = 0
			exp++
		}
	}
	return s | u16(exp&0x1f)<<10 | u16(mant>>13)
}

// F16ToF32 converts a 16-bit half-float to a 32-bit float.
func F16ToF32(value u16) f32 {
	s := uint32(value&0x8000) << 16
	exp := int32((value >> 10) & 0x1f)
	mant := uint32(value & 0x3ff)

	if exp == 0 {
		if mant == 0 {
			return f32(math.Float32frombits(s))
		}
		for (mant & 0x400) == 0 {
			mant <<= 1
			exp--
		}
		exp++
		mant &= 0x3ff
	}
	if exp == 31 {
		return f32(math.Float32frombits(s | 0xff<<23 | mant<<13))
	}
	exp = exp - 15 + 127
	return f32(math.Float32frombits(s | uint32(exp)<<23 | mant<<13))
}

// GbSqrt computes the square root of x.
func GbSqrt(x float64) float64 {
	return math.Sqrt(x)
}
