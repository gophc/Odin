package exact_value

import (
	"github.com/gophc/Odin/godin/big_int"
)

// NewInvalid creates an invalid exact value
func NewInvalid() ExactValue {
	return ExactValue{Kind: ExactValue_Invalid}
}

// NewBool creates a boolean exact value
func NewBool(b bool) ExactValue {
	return ExactValue{
		Kind:      ExactValue_Bool,
		valueBool: b,
	}
}

// NewString creates a string exact value
func NewString(s string) ExactValue {
	return ExactValue{
		Kind: ExactValue_String,
		valueString: String{
			Text: []byte(s),
		},
	}
}

// NewString16 creates a UTF-16 string exact value
func NewString16(s []uint16) ExactValue {
	return ExactValue{
		Kind: ExactValue_String16,
		valueString16: String16{
			Text: s,
		},
	}
}

// NewInteger creates an integer exact value from big.Int
func NewInteger(i *big_int.BigInt) ExactValue {
	if i == nil {
		return NewInvalid()
	}
	return ExactValue{
		Kind:         ExactValue_Integer,
		valueInteger: *i,
	}
}

// NewIntegerFromI64 creates an integer exact value from int64
func NewIntegerFromI64(i int64) ExactValue {
	return ExactValue{
		Kind:         ExactValue_Integer,
		valueInteger: *big_int.NewFromI64(i),
	}
}

// NewIntegerFromU64 creates an integer exact value from uint64
func NewIntegerFromU64(i uint64) ExactValue {
	return ExactValue{
		Kind:         ExactValue_Integer,
		valueInteger: *big_int.NewFromU64(i),
	}
}

// NewFloat creates a float exact value
func NewFloat(f float64) ExactValue {
	return ExactValue{
		Kind:       ExactValue_Float,
		valueFloat: f,
	}
}

// NewComplex creates a complex exact value
func NewComplex(real, imag float64) ExactValue {
	return ExactValue{
		Kind: ExactValue_Complex,
		valueComplex: &Complex128{
			Real: real,
			Imag: imag,
		},
	}
}

// NewQuaternion creates a quaternion exact value
func NewQuaternion(real, imag, jmag, kmag float64) ExactValue {
	return ExactValue{
		Kind: ExactValue_Quaternion,
		valueQuaternion: &Quaternion256{
			Real: real,
			Imag: imag,
			Jmag: jmag,
			Kmag: kmag,
		},
	}
}

// NewPointer creates a pointer exact value
func NewPointer(ptr int64) ExactValue {
	return ExactValue{
		Kind:         ExactValue_Pointer,
		valuePointer: ptr,
	}
}

// NewCompound creates a compound exact value
func NewCompound(ptr uintptr) ExactValue {
	return ExactValue{
		Kind:          ExactValue_Compound,
		valueCompound: ptr,
	}
}

// NewProcedure creates a procedure exact value
func NewProcedure(ptr uintptr) ExactValue {
	return ExactValue{
		Kind:           ExactValue_Procedure,
		valueProcedure: ptr,
	}
}

// NewTypeid creates a typeid exact value
func NewTypeid(ptr uintptr) ExactValue {
	return ExactValue{
		Kind:        ExactValue_Typeid,
		valueTypeid: ptr,
	}
}
