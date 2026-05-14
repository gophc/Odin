package cmd

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

func exact_value_integer_from_string(s String) ExactValue {
	result := ExactValue{Kind: ExactValueInteger}
	var success bool
	big_int_from_string(&result.ValueInteger, s, &success)
	if !success {
		return ExactValue{}
	}
	return result
}

func float_from_string(s String) (float64, bool) {
	str := goStr(s)
	str = strings.ReplaceAll(str, "_", "")
	str = strings.ReplaceAll(str, "E", "e")
	f, err := strconv.ParseFloat(str, 64)
	return f, err == nil
}

func exact_value_float_from_string(s String) ExactValue {
	str := goStr(s)
	if len(str) > 2 && str[:2] == "0h" {
		digit_count := 0
		for _, c := range str[2:] {
			if c != '_' {
				digit_count++
			}
		}
		u := u64_from_string(s)
		switch digit_count {
		case 4:
			x := u16(u)
			f := f16_to_f32(x)
			return exact_value_float(float64(f))
		case 8:
			x := u32(u)
			f := math.Float32frombits(x)
			return exact_value_float(float64(f))
		case 16:
			f := math.Float64frombits(u)
			return exact_value_float(f)
		default:
			panic("Invalid hexadecimal float, expected 4, 8, or 16 digits")
		}
	}
	if !string_contains_char(s, '.') && !string_contains_char(s, '-') {
		return exact_value_integer_from_string(s)
	}
	f, ok := float_from_string(s)
	if !ok {
		return ExactValue{}
	}
	return exact_value_float(f)
}

func exact_value_from_basic_literal(kind TokenKind, s String) ExactValue {
	switch kind {
	case Token_String:
		return exact_value_string(s)
	case Token_Integer:
		return exact_value_integer_from_string(s)
	case Token_Float:
		return exact_value_float_from_string(s)
	case Token_Imag:
		str := goStr(s)
		if len(str) == 0 {
			return ExactValue{}
		}
		lastRune := rune(str[len(str)-1])
		trimmed := str[:len(str)-1]
		imag, _ := float_from_string(S(trimmed))
		switch lastRune {
		case 'i':
			return exact_value_complex(0, imag)
		case 'j':
			return exact_value_quaternion(0, 0, imag, 0)
		case 'k':
			return exact_value_quaternion(0, 0, 0, imag)
		default:
			panic("Invalid imaginary basic literal")
		}
	case Token_Rune:
		str := goStr(s)
		r := rune(0xfffd)
		if len(str) == 1 {
			r = rune(str[0])
		} else {
			r, _ = utf8.DecodeRuneInString(str)
		}
		return exact_value_i64(int64(r))
	}
	return ExactValue{}
}

func exact_value_to_integer(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueBool:
		i := int64(0)
		if v.ValueBool {
			i = 1
		}
		return exact_value_i64(i)
	case ExactValueInteger:
		return v
	case ExactValueFloat:
		i := int64(v.ValueFloat)
		if float64(i) == v.ValueFloat {
			return exact_value_i64(i)
		}
	case ExactValuePointer:
		return exact_value_i64(int64(v.ValuePointer))
	}
	return ExactValue{}
}

func exact_value_to_float(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger:
		return exact_value_float(big_int_to_f64(&v.ValueInteger))
	case ExactValueFloat:
		return v
	}
	return ExactValue{}
}

func exact_value_to_complex(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger:
		return exact_value_complex(big_int_to_f64(&v.ValueInteger), 0)
	case ExactValueFloat:
		return exact_value_complex(v.ValueFloat, 0)
	case ExactValueComplex:
		return v
	}
	return ExactValue{}
}

func exact_value_to_quaternion(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger:
		return exact_value_quaternion(big_int_to_f64(&v.ValueInteger), 0, 0, 0)
	case ExactValueFloat:
		return exact_value_quaternion(v.ValueFloat, 0, 0, 0)
	case ExactValueComplex:
		if v.ValueComplex != nil {
			return exact_value_quaternion(v.ValueComplex.Real, v.ValueComplex.Imag, 0, 0)
		}
	case ExactValueQuaternion:
		return v
	}
	return ExactValue{}
}

func exact_value_real(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger, ExactValueFloat:
		return v
	case ExactValueComplex:
		if v.ValueComplex != nil {
			return exact_value_float(v.ValueComplex.Real)
		}
	case ExactValueQuaternion:
		if v.ValueQuaternion != nil {
			return exact_value_float(v.ValueQuaternion.Real)
		}
	}
	return ExactValue{}
}

func exact_value_imag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger, ExactValueFloat:
		return exact_value_i64(0)
	case ExactValueComplex:
		if v.ValueComplex != nil {
			return exact_value_float(v.ValueComplex.Imag)
		}
	case ExactValueQuaternion:
		if v.ValueQuaternion != nil {
			return exact_value_float(v.ValueQuaternion.Imag)
		}
	}
	return ExactValue{}
}

func exact_value_jmag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger, ExactValueFloat, ExactValueComplex:
		return exact_value_i64(0)
	case ExactValueQuaternion:
		if v.ValueQuaternion != nil {
			return exact_value_float(v.ValueQuaternion.Jmag)
		}
	}
	return ExactValue{}
}

func exact_value_kmag(v ExactValue) ExactValue {
	switch v.Kind {
	case ExactValueInteger, ExactValueFloat, ExactValueComplex:
		return exact_value_i64(0)
	case ExactValueQuaternion:
		if v.ValueQuaternion != nil {
			return exact_value_float(v.ValueQuaternion.Kmag)
		}
	}
	return ExactValue{}
}

func exact_value_to_i64(v ExactValue) int64 {
	iv := exact_value_to_integer(v)
	if iv.Kind == ExactValueInteger {
		return big_int_to_i64(&iv.ValueInteger)
	}
	return 0
}

func exact_value_to_u64(v ExactValue) uint64 {
	iv := exact_value_to_integer(v)
	if iv.Kind == ExactValueInteger {
		return big_int_to_u64(&iv.ValueInteger)
	}
	return 0
}

func exact_value_to_f64(v ExactValue) float64 {
	fv := exact_value_to_float(v)
	if fv.Kind == ExactValueFloat {
		return fv.ValueFloat
	}
	return 0.0
}
