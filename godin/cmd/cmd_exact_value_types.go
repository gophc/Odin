// Depends on: common.odin (String, String16, BigInt, Ast, Type, isize, i64, u64, f64, u8, u16, u32, uintptr)
package cmd

type ExactValueKind int

const (
	ExactValueInvalid    ExactValueKind = 0
	ExactValueBool       ExactValueKind = 1
	ExactValueString     ExactValueKind = 2
	ExactValueInteger    ExactValueKind = 3
	ExactValueFloat      ExactValueKind = 4
	ExactValueComplex    ExactValueKind = 5
	ExactValueQuaternion ExactValueKind = 6
	ExactValuePointer    ExactValueKind = 7
	ExactValueCompound   ExactValueKind = 8
	ExactValueProcedure  ExactValueKind = 9
	ExactValueTypeid     ExactValueKind = 10
	ExactValueString16   ExactValueKind = 11
	ExactValueCount
)

type Complex128 struct {
	Real float64
	Imag float64
}

type Quaternion256 struct {
	Imag float64
	Jmag float64
	Kmag float64
	Real float64
}

type ExactValue struct {
	Kind          ExactValueKind
	ValueBool     bool
	ValueString   String
	ValueInteger  BigInt
	ValueFloat    float64
	ValuePointer  int64
	ValueComplex  *Complex128
	ValueQuaternion *Quaternion256
	ValueCompound   *Ast
	ValueProcedure  *Ast
	ValueTypeid     *Type
	ValueString16   String16
}

var EmptyExactValue = ExactValue{}

func hashExactValue(v ExactValue) uintptr {
	var res uintptr
	switch v.Kind {
	case ExactValueInvalid:
		return 0
	case ExactValueBool:
		res = uintptr(gb_fnv32a(&v.ValueBool, isize(1)))
	case ExactValueString:
		res = uintptr(gb_fnv32a(v.ValueString.Text, v.ValueString.Len))
	case ExactValueString16:
		res = uintptr(gb_fnv32a(v.ValueString16.Text, v.ValueString16.Len*isize(2)))
	case ExactValueInteger:
		key := gb_fnv32a(v.ValueInteger.Dp, isize(len(v.ValueInteger.Dp)))
		last := uint8(v.ValueInteger.Sign)
		res = uintptr((key ^ uint32(last)) * 0x01000193)
	case ExactValueFloat:
		res = uintptr(gb_fnv32a(&v.ValueFloat, isize(8)))
	case ExactValuePointer:
		res = ptr_map_hash_key(v.ValuePointer)
	case ExactValueComplex:
		res = uintptr(gb_fnv32a(v.ValueComplex, isize(16)))
	case ExactValueQuaternion:
		res = uintptr(gb_fnv32a(v.ValueQuaternion, isize(32)))
	case ExactValueCompound:
		res = ptr_map_hash_key(v.ValueCompound)
	case ExactValueProcedure:
		res = ptr_map_hash_key(v.ValueProcedure)
	case ExactValueTypeid:
		res = ptr_map_hash_key(v.ValueTypeid)
	default:
		res = uintptr(gb_fnv32a(&v, isize(len(v))))
	}
	return res & 0x7fffffff
}

func exact_value_compound(node *Ast) ExactValue {
	return ExactValue{Kind: ExactValueCompound, ValueCompound: node}
}

func exact_value_bool(b bool) ExactValue {
	return ExactValue{Kind: ExactValueBool, ValueBool: b}
}

func exact_value_string(str String) ExactValue {
	return ExactValue{Kind: ExactValueString, ValueString: str}
}

func exact_value_string16(str String16) ExactValue {
	return ExactValue{Kind: ExactValueString16, ValueString16: str}
}

func exact_value_i64(i int64) ExactValue {
	result := ExactValue{Kind: ExactValueInteger}
	big_int_from_i64(&result.ValueInteger, i)
	return result
}

func exact_value_u64(i uint64) ExactValue {
	result := ExactValue{Kind: ExactValueInteger}
	big_int_from_u64(&result.ValueInteger, i)
	return result
}

func exact_value_float(f float64) ExactValue {
	return ExactValue{Kind: ExactValueFloat, ValueFloat: f}
}

func exact_value_complex(real, imag float64) ExactValue {
	result := ExactValue{Kind: ExactValueComplex}
	result.ValueComplex = &Complex128{Real: real, Imag: imag}
	return result
}

func exact_value_quaternion(real, imag, jmag, kmag float64) ExactValue {
	result := ExactValue{Kind: ExactValueQuaternion}
	result.ValueQuaternion = &Quaternion256{Real: real, Imag: imag, Jmag: jmag, Kmag: kmag}
	return result
}

func exact_value_pointer(ptr int64) ExactValue {
	return ExactValue{Kind: ExactValuePointer, ValuePointer: ptr}
}

func exact_value_procedure(node *Ast) ExactValue {
	return ExactValue{Kind: ExactValueProcedure, ValueProcedure: node}
}

func exact_value_typeid(t *Type) ExactValue {
	return ExactValue{Kind: ExactValueTypeid, ValueTypeid: t}
}
