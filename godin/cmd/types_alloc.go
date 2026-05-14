package cmd

func alloc_type(kind TypeKind) *Type {
	return &Type{
		Kind:        kind,
		CachedSize:  -1,
		CachedAlign: -1,
	}
}

func alloc_type_generic(scope *Scope, id int64, internedName InternedString, specialized *Type) *Type {
	t := alloc_type(TypeGeneric)
	t.Generic.ID = id
	t.Generic.Name = internedName.String()
	t.Generic.InternedName = internedName
	t.Generic.Specialized = specialized
	t.Generic.Scope = scope
	return t
}

func alloc_type_pointer(elem *Type) *Type {
	t := alloc_type(TypePointer)
	t.Pointer.Elem = elem
	return t
}

func alloc_type_multi_pointer(elem *Type) *Type {
	t := alloc_type(TypeMultiPointer)
	t.MultiPointer.Elem = elem
	return t
}

func alloc_type_soa_pointer(elem *Type) *Type {
	t := alloc_type(TypeSoaPointer)
	t.SoaPointer.Elem = elem
	return t
}

func alloc_type_pointer_to_multi_pointer(ptr *Type) *Type {
	original := ptr
	ptr = base_type(ptr)
	switch {
	case ptr != nil && ptr.Kind == TypePointer:
		return alloc_type_multi_pointer(ptr.Pointer.Elem)
	case ptr == nil || ptr.Kind != TypeMultiPointer:
		gb_assert_handler("Panic", 0, "", 0, "Invalid type: %s", type_to_string(original))
	}
	return original
}

func alloc_type_multi_pointer_to_pointer(ptr *Type) *Type {
	original := ptr
	ptr = base_type(ptr)
	switch {
	case ptr != nil && ptr.Kind == TypeMultiPointer:
		return alloc_type_pointer(ptr.MultiPointer.Elem)
	case ptr == nil || ptr.Kind != TypePointer:
		gb_assert_handler("Panic", 0, "", 0, "Invalid type: %s", type_to_string(original))
	}
	return original
}

func alloc_type_array(elem *Type, count int64, genericCount *Type) *Type {
	t := alloc_type(TypeArray)
	t.Array.Elem = elem
	t.Array.Count = count
	t.Array.GenericCount = genericCount
	return t
}

func alloc_type_matrix(elem *Type, rowCount, columnCount int64, genericRowCount, genericColumnCount *Type, isRowMajor bool) *Type {
	t := alloc_type(TypeMatrix)
	t.Matrix.Elem = elem
	t.Matrix.RowCount = rowCount
	t.Matrix.ColumnCount = columnCount
	t.Matrix.GenericRowCount = genericRowCount
	t.Matrix.GenericColumnCount = genericColumnCount
	t.Matrix.IsRowMajor = isRowMajor
	return t
}

func alloc_type_enumerated_array(elem *Type, index *Type, minValue, maxValue *ExactValue, count isize, op TokenKind) *Type {
	t := alloc_type(TypeEnumeratedArray)
	t.EnumeratedArray.Elem = elem
	t.EnumeratedArray.Index = index
	t.EnumeratedArray.MinValue = new(ExactValue)
	t.EnumeratedArray.MaxValue = new(ExactValue)
	*t.EnumeratedArray.MinValue = *minValue
	*t.EnumeratedArray.MaxValue = *maxValue
	t.EnumeratedArray.Op = op
	if count == 0 {
		t.EnumeratedArray.Count = 0
	} else {
		t.EnumeratedArray.Count = 1 + exact_value_to_i64(exact_value_sub(*maxValue, *minValue))
	}
	return t
}

func alloc_type_slice(elem *Type) *Type {
	t := alloc_type(TypeSlice)
	t.Slice.Elem = elem
	return t
}

func alloc_type_dynamic_array(elem *Type) *Type {
	t := alloc_type(TypeDynamicArray)
	t.DynamicArray.Elem = elem
	return t
}

func alloc_type_fixed_capacity_dynamic_array(elem *Type, capacity int64, genericCapacity *Type) *Type {
	t := alloc_type(TypeFixedCapacityDynamicArray)
	t.FixedCapacityDynamicArray.Elem = elem
	t.FixedCapacityDynamicArray.Capacity = capacity
	t.FixedCapacityDynamicArray.GenericCapacity = genericCapacity
	t.FixedCapacityDynamicArray.PaddingNeeded = -1
	return t
}

func alloc_type_struct() *Type {
	return alloc_type(TypeStruct)
}

func alloc_type_struct_complete() *Type {
	t := alloc_type(TypeStruct)
	wait_signal_set(&t.Struct.FieldsWaitSignal)
	wait_signal_set(&t.Struct.PolymorphicWaitSignal)
	return t
}

func alloc_type_union() *Type {
	return alloc_type(TypeUnion)
}

func alloc_type_enum() *Type {
	t := alloc_type(TypeEnum)
	t.Enum.MinValue = new(ExactValue)
	t.Enum.MaxValue = new(ExactValue)
	return t
}

func alloc_type_bit_field() *Type {
	return alloc_type(TypeBitField)
}

func alloc_type_named(name string, base *Type, typeName *Entity) *Type {
	t := alloc_type(TypeNamed)
	t.Named.Name = name
	t.Named.Base = base
	if base != t {
		t.Named.Base = base_type(base)
	}
	t.Named.TypeName = typeName
	return t
}

func is_calling_convention_none(cc ProcCallingConvention) bool {
	switch cc {
	case ProcCC_None, ProcCC_InlineAsm:
		return true
	}
	return false
}

func is_calling_convention_odin(cc ProcCallingConvention) bool {
	switch cc {
	case ProcCC_Odin, ProcCC_Contextless:
		return true
	}
	return false
}

func alloc_type_tuple() *Type {
	return alloc_type(TypeTuple)
}

func alloc_type_proc(scope *Scope, params *Type, paramCount isize, results *Type, resultCount isize, variadic bool, callingConvention ProcCallingConvention) *Type {
	t := alloc_type(TypeProc)
	if variadic {
		if params == nil || params.Kind != TypeTuple {
			gb_assert_handler("Panic", 0, "", 0, "variadic proc params must be a tuple")
		}
		if paramCount < 1 {
			gb_assert_handler("Panic", 0, "", 0, "variadic proc must have at least one parameter")
		}
		e := params.Tuple.Variables[paramCount-1]
		if base_type(e.Type).Kind != TypeSlice {
			gb_assert_handler("Panic", 0, "", 0, "Invalid type: %s", type_to_string(e.Type))
		}
	}
	t.Proc.Scope = scope
	t.Proc.Params = params
	t.Proc.ParamCount = int32(paramCount)
	t.Proc.Results = results
	t.Proc.ResultCount = int32(resultCount)
	t.Proc.Variadic = variadic
	t.Proc.CallingConvention = callingConvention
	return t
}

func alloc_type_bit_set() *Type {
	return alloc_type(TypeBitSet)
}

func alloc_type_simd_vector(count int64, elem *Type, genericCount *Type) *Type {
	t := alloc_type(TypeSimdVector)
	t.SimdVector.Count = count
	t.SimdVector.Elem = elem
	t.SimdVector.GenericCount = genericCount
	return t
}

func type_deref(t *Type, allowMultiPointer bool) *Type {
	if t != nil {
		bt := base_type(t)
		if bt == nil {
			return nil
		}
		switch bt.Kind {
		case TypePointer:
			return bt.Pointer.Elem
		case TypeSoaPointer:
			elem := base_type(bt.SoaPointer.Elem)
			if elem != nil && elem.Kind == TypeStruct && elem.Struct.SoaKind != StructSoaNone {
				return elem.Struct.SoaElem
			}
			return t
		case TypeMultiPointer:
			if allowMultiPointer {
				return bt.MultiPointer.Elem
			}
		}
	}
	return t
}

func base_type(t *Type) *Type {
	for {
		if t == nil {
			break
		}
		if t.Kind != TypeNamed {
			break
		}
		if t == t.Named.Base {
			return tInvalid
		}
		t = t.Named.Base
	}
	return t
}

func base_named_type(t *Type) *Type {
	if t.Kind != TypeNamed {
		return tInvalid
	}
	prevNamed := t
	t = t.Named.Base
	for {
		if t == nil {
			break
		}
		if t.Kind != TypeNamed {
			break
		}
		if t == t.Named.Base {
			return tInvalid
		}
		prevNamed = t
		t = t.Named.Base
	}
	return prevNamed
}

func base_enum_type(t *Type) *Type {
	bt := base_type(t)
	if bt != nil && bt.Kind == TypeEnum {
		return bt.Enum.BaseType
	}
	return t
}

func core_type(t *Type) *Type {
	for {
		if t == nil {
			break
		}
		switch t.Kind {
		case TypeNamed:
			if t == t.Named.Base {
				return tInvalid
			}
			t = t.Named.Base
			continue
		case TypeEnum:
			t = t.Enum.BaseType
			continue
		case TypeBitField:
			t = t.BitField.BackingType
			continue
		}
		break
	}
	return t
}

func set_base_type(t *Type, base *Type) {
	if t != nil && t.Kind == TypeNamed {
		t.Named.Base = base
	}
}

func type_path_init(tp *TypePath) {}

func type_path_free(tp *TypePath) {}

func type_path_print_illegal_cycle(tp *TypePath, startIndex isize) {
	if tp == nil || startIndex >= isize(len(tp.Path)) {
		return
	}
	e := tp.Path[startIndex]
	if e == nil {
		return
	}
	error(e.Token, "Illegal type declaration cycle of `%s`", gostr(e.Token.String))
	for j := startIndex; j < isize(len(tp.Path)); j++ {
		e2 := tp.Path[j]
		error(e2.Token, "\t%s refers to", gostr(e2.Token.String))
	}
	error(e.Token, "\t%s", gostr(e.Token.String))
	tp.Failure = true
	if e.Type != nil {
		e.Type.Failure = true
		if bt := base_type(e.Type); bt != nil {
			bt.Failure = true
		}
	}
}

func type_path_push(tp *TypePath, t *Type) bool {
	if tp == nil {
		return false
	}
	if t.Kind != TypeNamed {
		return false
	}
	e := t.Named.TypeName
	if e == nil {
		return false
	}
	mutex_lock(&tp.Mutex)
	defer mutex_unlock(&tp.Mutex)
	for i, p := range tp.Path {
		if p == e {
			type_path_print_illegal_cycle(tp, isize(i))
			return true
		}
	}
	tp.Path = append(tp.Path, e)
	return true
}

func type_path_pop(tp *TypePath) {
	if tp == nil {
		return
	}
	mutex_lock(&tp.Mutex)
	defer mutex_unlock(&tp.Mutex)
	if len(tp.Path) > 0 {
		tp.Path = tp.Path[:len(tp.Path)-1]
	}
}

func typeInfoFlagsOfType(type_ *Type) uint32 {
	if type_ == nil {
		return 0
	}
	flags := uint32(0)
	if isTypeComparable(type_) {
		flags |= TypeInfoFlagComparable
	}
	if isTypeSimpleCompare(type_) {
		flags |= TypeInfoFlagComparable | TypeInfoFlagSimpleCompare
	}
	return flags
}
