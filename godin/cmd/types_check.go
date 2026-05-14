package cmd

import "strings"

func is_type_named(t *Type) bool {
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return true
	}
	return t.Kind == TypeNamed
}

func is_type_boolean(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagBoolean) != 0
	}
	return false
}

func is_type_integer(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagInteger) != 0
	}
	return false
}

func is_type_integer_like(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & (BasicFlagInteger | BasicFlagBoolean)) != 0
	}
	if t.Kind == TypeBitSet {
		if t.BitSet.Underlying != nil {
			return is_type_integer_like(t.BitSet.Underlying)
		}
		return true
	}
	return false
}

func is_type_unsigned(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagUnsigned) != 0
	}
	if t.Kind == TypeEnum {
		return (t.Enum.BaseType.Basic.Flags & BasicFlagUnsigned) != 0
	}
	return false
}

func is_type_integer_128bit(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagInteger) != 0 && t.Basic.Size == 16
	}
	return false
}

func is_type_rune(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagRune) != 0
	}
	return false
}

func is_type_integer_or_float(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & (BasicFlagInteger | BasicFlagFloat)) != 0
	}
	return false
}

func is_type_numeric(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagNumeric) != 0
	} else if t.Kind == TypeEnum {
		return is_type_numeric(t.Enum.BaseType)
	}
	if t.Kind == TypeArray {
		return is_type_numeric(t.Array.Elem)
	}
	return false
}

func is_type_string(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagString) != 0
	}
	return false
}

func is_type_string16(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicString16
	}
	return false
}

func is_type_cstring(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicCstring
	}
	return false
}

func is_type_cstring16(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicCstring16
	}
	return false
}

func is_type_typed(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagUntyped) == 0
	}
	return true
}

func is_type_untyped(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagUntyped) != 0
	}
	return false
}

func is_type_ordered(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	switch t.Kind {
	case TypeBasic:
		return (t.Basic.Flags & BasicFlagOrdered) != 0
	case TypePointer:
		return true
	case TypeMultiPointer:
		return true
	}
	return false
}

func is_type_ordered_numeric(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	switch t.Kind {
	case TypeBasic:
		return (t.Basic.Flags & BasicFlagOrderedNumeric) != 0
	}
	return false
}

func is_type_constant_type(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	switch t.Kind {
	case TypeBasic:
		if t.Basic.Kind == BasicTypeid {
			return true
		}
		return (t.Basic.Flags & BasicFlagConstantType) != 0
	case TypeBitSet:
		return true
	case TypeProc:
		return true
	case TypeArray:
		return is_type_constant_type(t.Array.Elem)
	case TypeEnumeratedArray:
		return is_type_constant_type(t.EnumeratedArray.Elem)
	}
	return false
}

func is_type_float(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagFloat) != 0
	}
	return false
}

func is_type_complex(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagComplex) != 0
	}
	return false
}

func is_type_quaternion(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagQuaternion) != 0
	}
	return false
}

func is_type_complex_or_quaternion(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & (BasicFlagComplex | BasicFlagQuaternion)) != 0
	}
	return false
}

func is_type_pointer(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & BasicFlagPointer) != 0
	}
	return t.Kind == TypePointer
}

func is_type_soa_pointer(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeSoaPointer
}

func is_type_multi_pointer(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeMultiPointer
}

func is_type_internally_pointer_like(t *Type) bool {
	return is_type_pointer(t) || is_type_multi_pointer(t) || is_type_cstring(t) || is_type_proc(t)
}

func is_type_tuple(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeTuple
}

func is_type_uintptr(t *Type) bool {
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicUintptr
	}
	return false
}

func is_type_rawptr(t *Type) bool {
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicRawptr
	}
	return false
}

func is_type_u8(t *Type) bool {
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicU8
	}
	return false
}

func is_type_u16(t *Type) bool {
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return t.Basic.Kind == BasicU16
	}
	return false
}

func is_type_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeArray
}

func is_type_enumerated_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeEnumeratedArray
}

func is_type_matrix(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeMatrix
}

func is_type_dynamic_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeDynamicArray
}

func is_type_fixed_capacity_dynamic_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeFixedCapacityDynamicArray
}

func is_type_slice(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeSlice
}

func is_type_proc(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeProc
}

func is_type_asm_proc(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeProc && t.Proc.CallingConvention == ProcCC_InlineAsm
}

func is_type_simd_vector(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeSimdVector
}

func is_type_generic(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeGeneric
}

func is_type_u8_slice(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeSlice {
		return is_type_u8(t.Slice.Elem)
	}
	return false
}

func is_type_u8_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeArray {
		return is_type_u8(t.Array.Elem)
	}
	return false
}

func is_type_u8_ptr(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypePointer {
		return is_type_u8(t.Pointer.Elem)
	}
	return false
}

func is_type_u8_multi_ptr(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeMultiPointer {
		return is_type_u8(t.MultiPointer.Elem)
	}
	return false
}

func is_type_rune_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeArray {
		return is_type_rune(t.Array.Elem)
	}
	return false
}

func is_type_u16_slice(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeSlice {
		return is_type_u16(t.Slice.Elem)
	}
	return false
}

func is_type_u16_array(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeArray {
		return is_type_u16(t.Array.Elem)
	}
	return false
}

func is_type_u16_ptr(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypePointer {
		return is_type_u16(t.Pointer.Elem)
	}
	return false
}

func is_type_u16_multi_ptr(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeMultiPointer {
		return is_type_u16(t.MultiPointer.Elem)
	}
	return false
}

func is_type_array_like(t *Type) bool {
	return is_type_array(t) || is_type_enumerated_array(t)
}

func is_type_struct(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeStruct
}

func is_type_union(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeUnion
}

func is_type_soa_struct(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeStruct && t.Struct.SoaKind != StructSoaNone
}

func is_type_raw_union(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeStruct && t.Struct.IsRawUnion
}

func is_type_enum(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeEnum
}

func is_type_bit_set(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBitSet
}

func is_type_bit_field(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBitField
}

func is_type_map(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeMap
}

func is_type_union_maybe_pointer(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeUnion && len(t.Union.Variants) == 1 {
		v := t.Union.Variants[0]
		return is_type_internally_pointer_like(v)
	}
	return false
}

func is_type_union_maybe_pointer_original_alignment(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeUnion && len(t.Union.Variants) == 1 {
		v := t.Union.Variants[0]
		if is_type_internally_pointer_like(v) {
			return type_align_of(v) == type_align_of(t)
		}
	}
	return false
}

func is_type_any(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBasic && t.Basic.Kind == BasicAny
}

func is_type_typeid(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBasic && t.Basic.Kind == BasicTypeid
}

func is_type_untyped_nil(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBasic && (t.Basic.Kind == BasicUntypedNil || t.Basic.Kind == BasicUntypedUninit)
}

func is_type_untyped_uninit(t *Type) bool {
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeBasic && t.Basic.Kind == BasicUntypedUninit
}

func is_type_empty_union(t *Type) bool {
	if t == nil {
		return false
	}
	t = base_type(t)
	if t == nil {
		return false
	}
	return t.Kind == TypeUnion && len(t.Union.Variants) == 0
}

func is_type_valid_for_keys(t *Type) bool {
	t = core_type(t)
	if t.Kind == TypeGeneric {
		return true
	}
	if is_type_untyped(t) {
		return false
	}
	return type_size_of(t) > 0 && is_type_comparable(t)
}

func is_type_valid_bit_set_elem(t *Type) bool {
	if is_type_enum(t) {
		return true
	}
	t = core_type(t)
	if t.Kind == TypeGeneric {
		return true
	}
	return false
}

func is_valid_bit_field_backing_type(type_ *Type) bool {
	if type_ == nil {
		return false
	}
	type_ = base_type(type_)
	if is_type_untyped(type_) {
		return false
	}
	if is_type_integer(type_) {
		return true
	}
	if type_.Kind == TypeArray {
		return is_type_integer(type_.Array.Elem)
	}
	return false
}

func is_type_valid_vector_elem(t *Type) bool {
	t = base_type(t)
	if t.Kind == TypeBasic {
		if (t.Basic.Flags & BasicFlagEndianLittle) != 0 {
			return false
		}
		if (t.Basic.Flags & BasicFlagEndianBig) != 0 {
			return false
		}
		if is_type_integer(t) {
			return !is_type_integer_128bit(t)
		}
		if is_type_float(t) {
			return true
		}
		if is_type_boolean(t) {
			return true
		}
		if t.Basic.Kind == BasicRawptr {
			return true
		}
	}
	return false
}

func is_type_indexable(t *Type) bool {
	bt := base_type(t)
	switch bt.Kind {
	case TypeBasic:
		return bt.Basic.Kind == BasicString || bt.Basic.Kind == BasicString16
	case TypeArray:
	case TypeSlice:
	case TypeDynamicArray:
	case TypeFixedCapacityDynamicArray:
	case TypeMap:
		return true
	case TypeMultiPointer:
		return true
	case TypeEnumeratedArray:
		return true
	case TypeMatrix:
		return true
	}
	return false
}

func is_type_sliceable(t *Type) bool {
	bt := base_type(t)
	switch bt.Kind {
	case TypeBasic:
		return bt.Basic.Kind == BasicString || bt.Basic.Kind == BasicString16
	case TypeArray:
	case TypeSlice:
	case TypeDynamicArray:
	case TypeFixedCapacityDynamicArray:
		return true
	case TypeEnumeratedArray:
		return false
	case TypeMatrix:
		return false
	}
	return false
}

func is_type_polymorphic(t *Type, or_specialized ...bool) bool {
	orSpec := false
	if len(or_specialized) > 0 {
		orSpec = or_specialized[0]
	}

	if t == nil {
		return false
	}
	if t.Flags&TypeFlagInProcessOfCheckingPolymorphic != 0 {
		return false
	}

	switch t.Kind {
	case TypeGeneric:
		return true

	case TypeNamed:
		flags := t.Flags
		t.Flags |= TypeFlagInProcessOfCheckingPolymorphic
		ok := is_type_polymorphic(t.Named.Base, orSpec)
		t.Flags = flags
		return ok

	case TypePointer:
		return is_type_polymorphic(t.Pointer.Elem, orSpec)

	case TypeMultiPointer:
		return is_type_polymorphic(t.MultiPointer.Elem, orSpec)

	case TypeSoaPointer:
		return is_type_polymorphic(t.SoaPointer.Elem, orSpec)

	case TypeEnumeratedArray:
		if is_type_polymorphic(t.EnumeratedArray.Index, orSpec) {
			return true
		}
		return is_type_polymorphic(t.EnumeratedArray.Elem, orSpec)

	case TypeArray:
		if t.Array.GenericCount != nil {
			return true
		}
		return is_type_polymorphic(t.Array.Elem, orSpec)

	case TypeSimdVector:
		if t.SimdVector.GenericCount != nil {
			return true
		}
		return is_type_polymorphic(t.SimdVector.Elem, orSpec)

	case TypeDynamicArray:
		return is_type_polymorphic(t.DynamicArray.Elem, orSpec)

	case TypeFixedCapacityDynamicArray:
		if t.FixedCapacityDynamicArray.GenericCapacity != nil {
			return true
		}
		return is_type_polymorphic(t.FixedCapacityDynamicArray.Elem, orSpec)

	case TypeSlice:
		return is_type_polymorphic(t.Slice.Elem, orSpec)

	case TypeMatrix:
		if t.Matrix.GenericRowCount != nil {
			return true
		}
		if t.Matrix.GenericColumnCount != nil {
			return true
		}
		return is_type_polymorphic(t.Matrix.Elem, orSpec)

	case TypeTuple:
		for _, e := range t.Tuple.Variables {
			if e.Kind == EntityConstant {
				if e.Constant.Value.Kind != ExactValue_Invalid {
					return orSpec
				}
			} else if is_type_polymorphic(e.Type, orSpec) {
				return true
			}
		}

	case TypeProc:
		if t.Proc.IsPolymorphic {
			return true
		}
		if t.Proc.ParamCount > 0 && is_type_polymorphic(t.Proc.Params, orSpec) {
			return true
		}
		if t.Proc.ResultCount > 0 && is_type_polymorphic(t.Proc.Results, orSpec) {
			return true
		}

	case TypeEnum:
		if t.Kind == TypeEnum {
			if t.Enum.BaseType != nil {
				return is_type_polymorphic(t.Enum.BaseType, orSpec)
			}
			return false
		}

	case TypeUnion:
		if t.Union.IsPolymorphic {
			return true
		}
		if orSpec && t.Union.IsPolySpecialized {
			return true
		}

	case TypeStruct:
		if t.Struct.IsPolymorphic {
			return true
		}
		if orSpec && t.Struct.IsPolySpecialized {
			return true
		}

	case TypeMap:
		if t.Map.Key == nil || t.Map.Value == nil {
			return false
		}
		if is_type_polymorphic(t.Map.Key, orSpec) {
			return true
		}
		if is_type_polymorphic(t.Map.Value, orSpec) {
			return true
		}

	case TypeBitSet:
		if is_type_polymorphic(t.BitSet.Elem, orSpec) {
			return true
		}
		if t.BitSet.Underlying != nil && is_type_polymorphic(t.BitSet.Underlying, orSpec) {
			return true
		}
	}
	return false
}

func is_type_polymorphic_record(t *Type) bool {
	t = base_type(t)
	if t.Kind == TypeStruct {
		return t.Struct.IsPolymorphic
	} else if t.Kind == TypeUnion {
		return t.Union.IsPolymorphic
	}
	return false
}

func polymorphic_record_parent_scope(t *Type) *Scope {
	t = base_type(t)
	if is_type_polymorphic_record(t) {
		if t.Kind == TypeStruct {
			return t.Struct.Scope.Parent
		} else if t.Kind == TypeUnion {
			return t.Union.Scope.Parent
		}
	}
	return nil
}

func is_type_polymorphic_record_specialized(t *Type) bool {
	t = base_type(t)
	if t.Kind == TypeStruct {
		return t.Struct.IsPolySpecialized
	} else if t.Kind == TypeUnion {
		return t.Union.IsPolySpecialized
	}
	return false
}

func is_type_polymorphic_record_unspecialized(t *Type) bool {
	t = base_type(t)
	if t.Kind == TypeStruct {
		return t.Struct.IsPolymorphic && !t.Struct.IsPolySpecialized
	} else if t.Kind == TypeUnion {
		return t.Union.IsPolymorphic && !t.Union.IsPolySpecialized
	}
	return false
}

func get_record_polymorphic_params(t *Type) *TypeTuple {
	t = base_type(t)
	switch t.Kind {
	case TypeStruct:
		if t.Struct.PolymorphicParams != nil {
			return &t.Struct.PolymorphicParams.Tuple
		}
	case TypeUnion:
		if t.Union.PolymorphicParams != nil {
			return &t.Union.PolymorphicParams.Tuple
		}
	}
	return nil
}

func type_get_polymorphic_parent(t *Type, params_ **Type) *Entity {
	t = base_type(t)
	if t == nil {
		return nil
	}
	var parent *Type
	if t.Kind == TypeStruct {
		parent = t.Struct.PolymorphicParent
		if params_ != nil {
			*params_ = t.Struct.PolymorphicParams
		}
	} else if t.Kind == TypeUnion {
		parent = t.Union.PolymorphicParent
		if params_ != nil {
			*params_ = t.Union.PolymorphicParams
		}
	}
	if parent != nil {
		return parent.Named.TypeName
	}
	return nil
}

func type_has_nil(t *Type) bool {
	t = base_type(t)
	switch t.Kind {
	case TypeBasic:
		switch t.Basic.Kind {
		case BasicRawptr:
			return true
		case BasicAny:
			return true
		case BasicCstring:
			return true
		case BasicCstring16:
			return true
		case BasicTypeid:
			return true
		}
		return false
	case TypeEnum:
		return true
	case TypeBitSet:
		return true
	case TypeSlice:
		return true
	case TypeProc:
		return true
	case TypePointer:
		return true
	case TypeSoaPointer:
		return true
	case TypeMultiPointer:
		return true
	case TypeDynamicArray:
		return true
	case TypeMap:
		return true
	case TypeFixedCapacityDynamicArray:
		return false
	case TypeUnion:
		return t.Union.Kind != UnionTypeNoNil
	case TypeStruct:
		if is_type_soa_struct(t) {
			switch t.Struct.SoaKind {
			case StructSoaFixed:
				return false
			case StructSoaSlice:
				return true
			case StructSoaDynamic:
				return true
			}
		}
		return false
	}
	return false
}

func is_type_union_constantable(type_ *Type) bool {
	bt := base_type(type_)
	if bt.Kind != TypeUnion {
		gbAssertHandler("is_type_union_constantable: expected union type")
	}
	if len(bt.Union.Variants) == 0 {
		return true
	} else if len(bt.Union.Variants) == 1 {
		return is_type_constant_type(bt.Union.Variants[0])
	}
	for _, v := range bt.Union.Variants {
		if !is_type_constant_type(v) {
			return false
		}
	}
	return true
}

func is_type_raw_union_constantable(type_ *Type) bool {
	bt := base_type(type_)
	if bt.Kind != TypeStruct {
		gbAssertHandler("is_type_raw_union_constantable: expected struct type")
	}
	if !bt.Struct.IsRawUnion {
		gbAssertHandler("is_type_raw_union_constantable: expected raw union type")
	}
	for _, f := range bt.Struct.Fields {
		if !is_type_constant_type(f.Type) {
			return false
		}
	}
	return false
}

func elem_type_can_be_constant(t *Type) bool {
	t = base_type(t)
	if t == tInvalid {
		return false
	}
	if is_type_any(t) {
		return false
	}
	if is_type_raw_union(t) {
		return is_type_raw_union_constantable(t)
	}
	if is_type_union(t) {
		return is_type_union_constantable(t)
	}
	return true
}

func elem_cannot_be_constant(t *Type) bool {
	if is_type_any(t) {
		return true
	}
	if is_type_union(t) {
		return !is_type_union_constantable(t)
	}
	if is_type_raw_union(t) {
		return !is_type_raw_union_constantable(t)
	}
	return false
}

func is_type_lock_free(t *Type) bool {
	t = core_type(t)
	if t == tInvalid {
		return false
	}
	sz := type_size_of(t)
	return sz <= buildContext.MaxAlign
}

func is_type_comparable(t *Type) bool {
	t = base_type(t)
	switch t.Kind {
	case TypeBasic:
		switch t.Basic.Kind {
		case BasicUntypedNil:
			return false
		case BasicAny:
			return false
		case BasicRune:
			return true
		case BasicString:
			return true
		case BasicCstring:
			return true
		case BasicString16:
			return true
		case BasicCstring16:
			return true
		case BasicTypeid:
			return true
		}
		return true
	case TypePointer:
		return true
	case TypeSoaPointer:
		return true
	case TypeMultiPointer:
		return true
	case TypeEnum:
		return is_type_comparable(core_type(t))
	case TypeEnumeratedArray:
		return is_type_comparable(t.EnumeratedArray.Elem)
	case TypeArray:
		return is_type_comparable(t.Array.Elem)
	case TypeProc:
		return true
	case TypeMatrix:
		return is_type_comparable(t.Matrix.Elem)
	case TypeFixedCapacityDynamicArray:
		return false
	case TypeBitSet:
		return true
	case TypeStruct:
		if t.Struct.SoaKind != StructSoaNone {
			return false
		}
		if t.Struct.IsRawUnion {
			return is_type_simple_compare(t)
		}
		for _, f := range t.Struct.Fields {
			if !is_type_comparable(f.Type) {
				return false
			}
		}
		return true
	case TypeUnion:
		for _, v := range t.Union.Variants {
			if !is_type_comparable(v) {
				return false
			}
		}
		return true
	case TypeSimdVector:
		return true
	case TypeBitField:
		return is_type_comparable(t.BitField.BackingType)
	}
	return false
}

func is_type_simple_compare(t *Type) bool {
	t = core_type(t)
	switch t.Kind {
	case TypeArray:
		return is_type_simple_compare(t.Array.Elem)
	case TypeEnumeratedArray:
		return is_type_simple_compare(t.EnumeratedArray.Elem)
	case TypeFixedCapacityDynamicArray:
		return false
	case TypeBasic:
		if (t.Basic.Flags & BasicFlagSimpleCompare) != 0 {
			return true
		}
		if t.Basic.Kind == BasicTypeid {
			return true
		}
		return false
	case TypePointer:
		return true
	case TypeMultiPointer:
		return true
	case TypeSoaPointer:
		return true
	case TypeProc:
		return true
	case TypeBitSet:
		return true
	case TypeBitField:
		return true
	case TypeMatrix:
		return is_type_simple_compare(t.Matrix.Elem)
	case TypeStruct:
		if t.Struct.IsSimple {
			return true
		}
		for _, f := range t.Struct.Fields {
			if !is_type_simple_compare(f.Type) {
				return false
			}
		}
		return true
	case TypeUnion:
		for _, v := range t.Union.Variants {
			if !is_type_simple_compare(v) {
				return false
			}
		}
		return len(t.Union.Variants) == 1
	case TypeSimdVector:
		return is_type_simple_compare(t.SimdVector.Elem)
	case TypeTuple:
		if len(t.Tuple.Variables) == 1 {
			return is_type_simple_compare(t.Tuple.Variables[0].Type)
		}
	case TypeSlice:
		return false
	case TypeDynamicArray:
		return false
	case TypeMap:
		return false
	}
	return false
}

func is_type_nearly_simple_compare(t *Type) bool {
	t = core_type(t)
	switch t.Kind {
	case TypeArray:
		return is_type_nearly_simple_compare(t.Array.Elem)
	case TypeEnumeratedArray:
		return is_type_nearly_simple_compare(t.EnumeratedArray.Elem)
	case TypeFixedCapacityDynamicArray:
		return false
	case TypeBasic:
		if (t.Basic.Flags & (BasicFlagSimpleCompare | BasicFlagNumeric)) != 0 {
			return true
		}
		if t.Basic.Kind == BasicTypeid {
			return true
		}
		return false
	case TypePointer:
		return true
	case TypeMultiPointer:
		return true
	case TypeSoaPointer:
		return true
	case TypeProc:
		return true
	case TypeBitSet:
		return true
	case TypeBitField:
		return true
	case TypeMatrix:
		return is_type_nearly_simple_compare(t.Matrix.Elem)
	case TypeStruct:
		if t.Struct.IsSimple {
			return true
		}
		for _, f := range t.Struct.Fields {
			if !is_type_nearly_simple_compare(f.Type) {
				return false
			}
		}
		return true
	case TypeUnion:
		for _, v := range t.Union.Variants {
			if !is_type_nearly_simple_compare(v) {
				return false
			}
		}
		return len(t.Union.Variants) == 1
	case TypeSimdVector:
		return is_type_nearly_simple_compare(t.SimdVector.Elem)
	case TypeTuple:
		if len(t.Tuple.Variables) == 1 {
			return is_type_nearly_simple_compare(t.Tuple.Variables[0].Type)
		}
	case TypeSlice:
		return false
	case TypeDynamicArray:
		return false
	case TypeMap:
		return false
	}
	return false
}

func is_type_load_safe(type_ *Type) bool {
	if type_ == nil {
		gbAssertHandler("is_type_load_safe: nil type")
	}
	type_ = core_type(core_array_type(type_))
	switch type_.Kind {
	case TypeBasic:
		return (type_.Basic.Flags & (BasicFlagBoolean | BasicFlagNumeric | BasicFlagRune)) != 0
	case TypeBitSet:
		if type_.BitSet.Underlying != nil {
			return is_type_load_safe(type_.BitSet.Underlying)
		}
		return true
	case TypePointer:
		return false
	case TypeMultiPointer:
		return false
	case TypeSlice:
		return false
	case TypeDynamicArray:
		return false
	case TypeProc:
		return false
	case TypeSoaPointer:
		return false
	case TypeFixedCapacityDynamicArray:
		return false
	case TypeEnum:
		compilerError("should never be hit")
		return false
	case TypeEnumeratedArray:
		compilerError("should never be hit")
		return false
	case TypeArray:
		compilerError("should never be hit")
		return false
	case TypeSimdVector:
		compilerError("should never be hit")
		return false
	case TypeMatrix:
		compilerError("should never be hit")
		return false
	case TypeStruct:
		for _, f := range type_.Struct.Fields {
			if !is_type_load_safe(f.Type) {
				return false
			}
		}
		return type_size_of(type_) > 0
	case TypeUnion:
		for _, v := range type_.Union.Variants {
			if !is_type_load_safe(v) {
				return false
			}
		}
		return type_size_of(type_) > 0
	}
	return false
}

func get_array_type_count(t *Type) int64 {
	bt := base_type(t)
	if bt.Kind == TypeArray {
		return bt.Array.Count
	} else if bt.Kind == TypeEnumeratedArray {
		return bt.EnumeratedArray.Count
	} else if bt.Kind == TypeSimdVector {
		return bt.SimdVector.Count
	}
	if !is_type_array_like(t) {
		gbAssertHandler("get_array_type_count: expected array-like type")
	}
	return -1
}

func base_array_type(t *Type) *Type {
	bt := base_type(t)
	if is_type_array(bt) {
		return bt.Array.Elem
	} else if is_type_enumerated_array(bt) {
		return bt.EnumeratedArray.Elem
	} else if is_type_simd_vector(bt) {
		return bt.SimdVector.Elem
	} else if is_type_fixed_capacity_dynamic_array(bt) {
		return bt.FixedCapacityDynamicArray.Elem
	} else if is_type_matrix(bt) {
		return bt.Matrix.Elem
	}
	return t
}

func base_any_array_type(t *Type) *Type {
	bt := base_type(t)
	if is_type_array(bt) {
		return bt.Array.Elem
	} else if is_type_slice(bt) {
		return bt.Slice.Elem
	} else if is_type_dynamic_array(bt) {
		return bt.DynamicArray.Elem
	} else if is_type_fixed_capacity_dynamic_array(bt) {
		return bt.FixedCapacityDynamicArray.Elem
	} else if is_type_enumerated_array(bt) {
		return bt.EnumeratedArray.Elem
	} else if is_type_simd_vector(bt) {
		return bt.SimdVector.Elem
	} else if is_type_matrix(bt) {
		return bt.Matrix.Elem
	}
	return t
}

func core_array_type(t *Type) *Type {
	for {
		t = base_array_type(t)
		switch t.Kind {
		case TypeArray:
		case TypeEnumeratedArray:
		case TypeSimdVector:
		case TypeMatrix:
		default:
			return t
		}
	}
}

func base_complex_elem_type(t *Type) *Type {
	t = core_type(t)
	if t.Kind == TypeBasic {
		switch t.Basic.Kind {
		case BasicComplex32:
			return tF16
		case BasicComplex64:
			return tF32
		case BasicComplex128:
			return tF64
		case BasicQuaternion64:
			return tF16
		case BasicQuaternion128:
			return tF32
		case BasicQuaternion256:
			return tF64
		case BasicUntypedComplex:
			return tUntypedFloat
		case BasicUntypedQuaternion:
			return tUntypedFloat
		}
	}
	compilerError("Invalid complex type")
	return tInvalid
}

func type_unsigned_equivalent(t *Type) *Type {
	originalType := t
	t = base_type(t)
	if is_type_simd_vector(t) {
		if is_type_unsigned(t.SimdVector.Elem) {
			return originalType
		}
		return alloc_type_simd_vector(t.SimdVector.Count, type_unsigned_equivalent(t.SimdVector.Elem), nil)
	}
	sz := type_size_of(t)
	switch sz {
	case 1:
		return tU8
	case 2:
		return tU16
	case 4:
		return tU32
	case 8:
		return tU64
	case 16:
		return tU128
	}
	compilerError("No known equivalent unsigned integer sized for %s", type_to_string(t))
	return tInvalid
}

func type_signed_equivalent(t *Type) *Type {
	originalType := t
	t = base_type(t)
	if is_type_simd_vector(t) {
		if !is_type_unsigned(t.SimdVector.Elem) {
			return originalType
		}
		return alloc_type_simd_vector(t.SimdVector.Count, type_signed_equivalent(t.SimdVector.Elem), nil)
	}
	sz := type_size_of(t)
	switch sz {
	case 1:
		return tI8
	case 2:
		return tI16
	case 4:
		return tI32
	case 8:
		return tI64
	case 16:
		return tI128
	}
	compilerError("No known equivalent signed integer sized for %s", type_to_string(t))
	return tInvalid
}

func bit_set_to_int(t *Type) *Type {
	if !is_type_bit_set(t) {
		gbAssertHandler("bit_set_to_int: expected bit_set type")
	}
	bt := base_type(t)
	underlying := bt.BitSet.Underlying
	if underlying != nil && is_type_integer(underlying) {
		return underlying
	}
	if underlying != nil && is_valid_bit_field_backing_type(underlying) {
		return underlying
	}
	sz := type_size_of(t)
	switch sz {
	case 0:
		return tU8
	case 1:
		return tU8
	case 2:
		return tU16
	case 4:
		return tU32
	case 8:
		return tU64
	case 16:
		return tU128
	}
	compilerError("Unknown bit_set size")
	return nil
}

func type_math_rank(t *Type) int32 {
	rank := int32(0)
	for {
		t = base_type(t)
		switch t.Kind {
		case TypeArray:
			rank += 1
			t = t.Array.Elem
		case TypeMatrix:
			rank += 2
			t = t.Matrix.Elem
		default:
			return rank
		}
	}
}

func type_endian_kind_of(t *Type) TypeEndianKind {
	t = core_type(t)
	if t.Kind == TypeBasic {
		if (t.Basic.Flags & BasicFlagEndianLittle) != 0 {
			return TypeEndianLittle
		}
		if (t.Basic.Flags & BasicFlagEndianBig) != 0 {
			return TypeEndianBig
		}
	} else if t.Kind == TypeBitSet {
		return type_endian_kind_of(bit_set_to_int(t))
	}
	return TypeEndianPlatform
}

func is_type_endian_big(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		if (t.Basic.Flags & BasicFlagEndianBig) != 0 {
			return true
		} else if (t.Basic.Flags & BasicFlagEndianLittle) != 0 {
			return false
		}
		return buildContext.EndianKind == TargetEndian_Big
	} else if t.Kind == TypeBitSet {
		return is_type_endian_big(bit_set_to_int(t))
	} else if t.Kind == TypePointer {
		return is_type_endian_big(tUintptr)
	}
	return buildContext.EndianKind == TargetEndian_Big
}

func is_type_endian_little(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		if (t.Basic.Flags & BasicFlagEndianLittle) != 0 {
			return true
		} else if (t.Basic.Flags & BasicFlagEndianBig) != 0 {
			return false
		}
		return buildContext.EndianKind == TargetEndian_Little
	} else if t.Kind == TypeBitSet {
		return is_type_endian_little(bit_set_to_int(t))
	} else if t.Kind == TypePointer {
		return is_type_endian_little(tUintptr)
	}
	return buildContext.EndianKind == TargetEndian_Little
}

func is_type_endian_platform(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBasic {
		return (t.Basic.Flags & (BasicFlagEndianLittle | BasicFlagEndianBig)) == 0
	} else if t.Kind == TypeBitSet {
		return is_type_endian_platform(bit_set_to_int(t))
	} else if t.Kind == TypePointer {
		return is_type_endian_platform(tUintptr)
	}
	return false
}

func types_have_same_internal_endian(a, b *Type) bool {
	return is_type_endian_little(a) == is_type_endian_little(b)
}

func is_type_endian_specific(t *Type) bool {
	t = core_type(t)
	if t == nil {
		return false
	}
	if t.Kind == TypeBitSet {
		t = bit_set_to_int(t)
	}
	if t.Kind == TypeBasic {
		switch t.Basic.Kind {
		case BasicI16le:
			return true
		case BasicU16le:
			return true
		case BasicI32le:
			return true
		case BasicU32le:
			return true
		case BasicI64le:
			return true
		case BasicU64le:
			return true
		case BasicI128le:
			return true
		case BasicU128le:
			return true

		case BasicI16be:
			return true
		case BasicU16be:
			return true
		case BasicI32be:
			return true
		case BasicU32be:
			return true
		case BasicI64be:
			return true
		case BasicU64be:
			return true
		case BasicI128be:
			return true
		case BasicU128be:
			return true

		case BasicF16le:
			return true
		case BasicF16be:
			return true
		case BasicF32le:
			return true
		case BasicF32be:
			return true
		case BasicF64le:
			return true
		case BasicF64be:
			return true
		}
	}
	return false
}

func is_type_dereferenceable(t *Type) bool {
	if is_type_rawptr(t) {
		return false
	}
	return is_type_pointer(t) || is_type_soa_pointer(t)
}

func is_type_different_to_arch_endianness(t *Type) bool {
	switch buildContext.EndianKind {
	case TargetEndian_Little:
		return !is_type_endian_little(t)
	case TargetEndian_Big:
		return !is_type_endian_big(t)
	}
	return false
}

func integer_endian_type_to_platform_type(t *Type) *Type {
	t = core_type(t)
	if t.Kind == TypeBitSet {
		t = bit_set_to_int(t)
	}
	if t.Kind != TypeBasic {
		gbAssertHandler("integer_endian_type_to_platform_type: expected basic type")
	}
	switch t.Basic.Kind {
	case BasicI16le:
		return tI16
	case BasicU16le:
		return tU16
	case BasicI32le:
		return tI32
	case BasicU32le:
		return tU32
	case BasicI64le:
		return tI64
	case BasicU64le:
		return tU64
	case BasicI128le:
		return tI128
	case BasicU128le:
		return tU128

	case BasicI16be:
		return tI16
	case BasicU16be:
		return tU16
	case BasicI32be:
		return tI32
	case BasicU32be:
		return tU32
	case BasicI64be:
		return tI64
	case BasicU64be:
		return tU64
	case BasicI128be:
		return tI128
	case BasicU128be:
		return tU128

	case BasicF16le:
		return tF16
	case BasicF16be:
		return tF16
	case BasicF32le:
		return tF32
	case BasicF32be:
		return tF32
	case BasicF64le:
		return tF64
	case BasicF64be:
		return tF64
	}
	return t
}

func matched_target_features(t TypeProc) int {
	if len(t.RequireTargetFeature) == 0 {
		return 0
	}
	matches := 0
	for _, part := range strings.Split(t.RequireTargetFeature, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if checkTargetFeatureIsValidForTargetArch(part, "") {
			matches++
		}
	}
	return matches
}
