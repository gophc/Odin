package exact_value

import (
	"github.com/gophc/Odin/godin/big_int"
)

// String represents a byte string
type String struct {
	Text []byte
}

// String16 represents a UTF-16 string
type String16 struct {
	Text []uint16
}

// Complex128 represents a complex number
type Complex128 struct {
	Real float64
	Imag float64
}

// Quaternion256 represents a quaternion
type Quaternion256 struct {
	Real float64
	Imag float64
	Jmag float64
	Kmag float64
}

// ExactValue is the core type for compile-time constant values
type ExactValue struct {
	Kind ExactValueKind

	// Union fields
	valueBool       bool
	valueString     String
	valueString16   String16
	valueInteger    big_int.BigInt
	valueFloat      float64
	valueComplex    *Complex128
	valueQuaternion *Quaternion256
	valuePointer    int64
	valueCompound   uintptr // In Go, we can use uintptr for opaque pointers
	valueProcedure  uintptr
	valueTypeid     uintptr
}

// Helper methods to access union fields safely
func (v ExactValue) ValueBool() bool {
	if v.Kind != ExactValue_Bool {
		return false
	}
	return v.valueBool
}

func (v ExactValue) ValueString() String {
	if v.Kind != ExactValue_String {
		return String{}
	}
	return v.valueString
}

func (v ExactValue) ValueString16() String16 {
	if v.Kind != ExactValue_String16 {
		return String16{}
	}
	return v.valueString16
}

func (v ExactValue) ValueInteger() big_int.BigInt {
	if v.Kind != ExactValue_Integer {
		return big_int.BigInt{}
	}
	return v.valueInteger
}

func (v ExactValue) ValueFloat() float64 {
	if v.Kind != ExactValue_Float {
		return 0
	}
	return v.valueFloat
}

func (v ExactValue) ValueComplex() *Complex128 {
	if v.Kind != ExactValue_Complex {
		return nil
	}
	return v.valueComplex
}

func (v ExactValue) ValueQuaternion() *Quaternion256 {
	if v.Kind != ExactValue_Quaternion {
		return nil
	}
	return v.valueQuaternion
}

func (v ExactValue) ValuePointer() int64 {
	if v.Kind != ExactValue_Pointer {
		return 0
	}
	return v.valuePointer
}

func (v ExactValue) ValueCompound() uintptr {
	if v.Kind != ExactValue_Compound {
		return 0
	}
	return v.valueCompound
}

func (v ExactValue) ValueProcedure() uintptr {
	if v.Kind != ExactValue_Procedure {
		return 0
	}
	return v.valueProcedure
}

func (v ExactValue) ValueTypeid() uintptr {
	if v.Kind != ExactValue_Typeid {
		return 0
	}
	return v.valueTypeid
}

// Invalid returns an invalid exact value
func Invalid() ExactValue {
	return ExactValue{Kind: ExactValue_Invalid}
}

// IsValid checks if the value is valid
func (v ExactValue) IsValid() bool {
	return v.Kind != ExactValue_Invalid
}
