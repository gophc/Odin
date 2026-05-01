package exact_value

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gophc/Odin/godin/big_int"
)

// BasicLiteralKind represents the type of basic literal
type BasicLiteralKind int

const (
	LitString BasicLiteralKind = iota
	LitInteger
	LitFloat
	LitImag
	LitRune
)

// floatFromString parses a float from a string, ignoring underscores and converting 'E' to 'e'
func floatFromString(s string, success *bool) float64 {
	var buf []byte
	if len(s) > 128 {
		buf = make([]byte, 0, len(s))
	} else {
		buf = make([]byte, 0, len(s))
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			continue
		}
		if c == 'E' {
			c = 'e'
		}
		buf = append(buf, c)
	}
	str := string(buf)
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		if success != nil {
			*success = false
		}
		return 0
	}
	if success != nil {
		*success = true
	}
	return f
}

// ExactValueFromString creates an exact value from a string literal (integer, float, or hex float)
func ExactValueFromString(s string) ExactValue {
	if len(s) > 2 && s[0] == '0' && (s[1] == 'h' || s[1] == 'H') {
		// Hexadecimal float: 0h...
		digitStr := s[2:]
		digitCount := 0
		for i := 0; i < len(digitStr); i++ {
			if digitStr[i] != '_' {
				digitCount++
			}
		}
		u := hexToUint64(digitStr)
		switch digitCount {
		case 4:
			x := uint16(u)
			f := float16ToFloat32(x)
			return NewFloat(float64(f))
		case 8:
			x := uint32(u)
			f := float32FromBits(x)
			return NewFloat(float64(f))
		case 16:
			f := float64FromBits(u)
			return NewFloat(f)
		default:
			return NewInvalid()
		}
	}
	// Check if it's an integer (no '.' and no exponent)
	if !strings.ContainsAny(s, ".") && !strings.ContainsAny(s, "eE") {
		return exactValueIntegerFromString(s)
	}
	var success bool
	f := floatFromString(s, &success)
	if !success {
		return NewInvalid()
	}
	return NewFloat(f)
}

func hexToUint64(s string) uint64 {
	s = strings.ReplaceAll(s, "_", "")
	if len(s) == 0 {
		return 0
	}
	u, _ := strconv.ParseUint(s, 16, 64)
	return u
}

// float16ToFloat32 converts a 16-bit float (IEEE 754 binary16) to float32
func float16ToFloat32(x uint16) float32 {
	sign := (x >> 15) & 1
	exp := (x >> 10) & 0x1F
	mant := x & 0x3FF

	var f float32
	if exp == 0 {
		if mant == 0 {
			f = 0
		} else {
			f = float32(mant) * float32FromBits(0x33800000) // 2^-24
		}
	} else if exp == 0x1F {
		if mant == 0 {
			f = float32(float64Inf(int(sign)))
		} else {
			f = float32(float64NaN())
		}
	} else {
		f32bits := (uint32(sign) << 31) | (uint32(exp+112) << 23) | (uint32(mant) << 13)
		f = float32FromBits(f32bits)
	}
	if sign == 1 {
		f = -f
	}
	return f
}

func float32FromBits(bits uint32) float32 {
	return float32(math.Float32frombits(bits))
}

func float64FromBits(bits uint64) float64 {
	return math.Float64frombits(bits)
}

func float64Inf(sign int) float64 {
	return float64FromBits(0x7FF0000000000000 | (uint64(sign&1) << 63))
}

func float64NaN() float64 {
	return float64FromBits(0x7FF8000000000000)
}

func exactValueIntegerFromString(s string) ExactValue {
	z := big_int.New()
	ok := big_int.FromString(z, s, 0)
	if !ok {
		return NewInvalid()
	}
	return NewInteger(z)
}

// ExactValueFromBasicLiteral creates an exact value from a basic literal
func ExactValueFromBasicLiteral(kind BasicLiteralKind, s string) ExactValue {
	switch kind {
	case LitString:
		return NewString(s)
	case LitInteger:
		return exactValueIntegerFromString(s)
	case LitFloat:
		return ExactValueFromString(s)
	case LitImag:
		if len(s) == 0 {
			return NewInvalid()
		}
		lastRune := rune(s[len(s)-1])
		str := s[:len(s)-1]
		imag := floatFromString(str, nil)
		switch lastRune {
		case 'i':
			return NewComplex(0, imag)
		case 'j':
			return NewQuaternion(0, 0, imag, 0)
		case 'k':
			return NewQuaternion(0, 0, 0, imag)
		default:
			return NewInvalid()
		}
	case LitRune:
		var r rune
		if len(s) == 1 {
			r = rune(s[0])
		} else {
			r, _ = utf8.DecodeRuneInString(s)
			if r == utf8.RuneError {
				r = 0xFFFD
			}
		}
		return NewIntegerFromI64(int64(r))
	}
	return NewInvalid()
}

// NewIntegerFromString is a convenience function to create an integer exact value from a string
func NewIntegerFromString(s string) ExactValue {
	return exactValueIntegerFromString(s)
}
