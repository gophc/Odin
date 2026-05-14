package cmd

import "sync/atomic"

func typeSetOffsetsOf(fields []*Entity, isPacked, isRawUnion bool, minFieldAlign, maxFieldAlign int64) []int64 {
	if minFieldAlign == 0 {
		minFieldAlign = 1
	}
	offsets := make([]int64, len(fields))
	var path TypePath
	typePathInit(&path)
	defer typePathFree(&path)

	currOffset := int64(0)

	if isRawUnion {
		for i := range fields {
			offsets[i] = 0
		}
	} else if isPacked {
		for i := range fields {
			if fields[i].Kind != EntityVariable {
				offsets[i] = -1
			} else {
				size := typeSizeOfInternal(fields[i].Type, &path)
				offsets[i] = currOffset
				currOffset += size
			}
		}
	} else {
		for i := range fields {
			if fields[i].Kind != EntityVariable {
				offsets[i] = -1
			} else {
				t := fields[i].Type
				align := typeAlignOfInternal(t, &path)
				if align < minFieldAlign {
					align = minFieldAlign
				}
				if maxFieldAlign > minFieldAlign && align > maxFieldAlign {
					align = maxFieldAlign
				}
				size := typeSizeOfInternal(t, &path)
				if size < 0 {
					size = 0
				}
				currOffset = alignFormula(currOffset, align)
				offsets[i] = currOffset
				currOffset += size
			}
		}
	}
	return offsets
}

func typeSetOffsets(t *Type) bool {
	t = baseType(t)
	if t.Kind == TypeStruct {
		mutexLock(&t.Struct.OffsetMutex)
		if !t.Struct.AreOffsetsSet {
			t.Struct.AreOffsetsBeingProcessed = true
			t.Struct.Offsets = typeSetOffsetsOf(t.Struct.Fields, t.Struct.IsPacked, t.Struct.IsRawUnion, t.Struct.CustomMinFieldAlign, t.Struct.CustomMaxFieldAlign)
			t.Struct.AreOffsetsBeingProcessed = false
			t.Struct.AreOffsetsSet = true
		}
		mutexUnlock(&t.Struct.OffsetMutex)
		return true
	} else if isTypeTuple(t) {
		mutexLock(&t.Tuple.Mutex)
		if !t.Tuple.AreOffsetsSet {
			t.Tuple.AreOffsetsBeingProcessed = true
			t.Tuple.Offsets = typeSetOffsetsOf(t.Tuple.Variables, t.Tuple.IsPacked, false, 1, 0)
			t.Tuple.AreOffsetsBeingProcessed = false
			t.Tuple.AreOffsetsSet = true
		}
		mutexUnlock(&t.Tuple.Mutex)
		return true
	}
	return false
}

func typeSizeOf(t *Type) int64 {
	if t == nil {
		return 0
	}
	if t.Kind == TypeBasic {
		var size int64 = -1
		switch t.Basic.Kind {
		case BasicString:
			size = 2 * buildContext.IntSize
		case BasicCstring:
			size = buildContext.PtrSize
		case BasicString16:
			size = 2 * buildContext.IntSize
		case BasicCstring16:
			size = buildContext.PtrSize
		case BasicAny:
			size = 16
		case BasicTypeid:
			size = 8
		case BasicInt, BasicUint:
			size = buildContext.IntSize
		case BasicUintptr, BasicRawptr:
			size = buildContext.PtrSize
		default:
			size = t.Basic.Size
		}
		atomic.StoreInt64(&t.CachedSize, size)
		return size
	} else if t.Kind != TypeNamed {
		cached := atomic.LoadInt64(&t.CachedSize)
		if cached >= 0 {
			return cached
		}
	}
	var path TypePath
	typePathInit(&path)
	size := typeSizeOfInternal(t, &path)
	atomic.StoreInt64(&t.CachedSize, size)
	typePathFree(&path)
	return size
}

func typeAlignOf(t *Type) int64 {
	if t == nil {
		return 1
	}
	if t.Kind != TypeNamed {
		cached := atomic.LoadInt64(&t.CachedAlign)
		if cached > 0 {
			return cached
		}
	}
	var path TypePath
	typePathInit(&path)
	align := typeAlignOfInternal(t, &path)
	atomic.StoreInt64(&t.CachedAlign, align)
	typePathFree(&path)
	return align
}

func typeAlignOfInternal(t *Type, path *TypePath) int64 {
	if path == nil {
		gbAssertHandler("type_align_of_internal: path is nil")
	}
	if t.Failure {
		return FailureAlignment
	}
	t = baseType(t)

	switch t.Kind {
	case TypeBasic:
		switch t.Basic.Kind {
		case BasicString:
			return buildContext.IntSize
		case BasicCstring:
			return buildContext.PtrSize
		case BasicString16:
			return buildContext.IntSize
		case BasicCstring16:
			return buildContext.PtrSize
		case BasicAny:
			return 8
		case BasicTypeid:
			return 8
		case BasicInt, BasicUint:
			return buildContext.IntSize
		case BasicUintptr, BasicRawptr:
			return buildContext.PtrSize
		case BasicComplex32, BasicComplex64, BasicComplex128:
			return typeSizeOfInternal(t, path) / 2
		case BasicQuaternion64, BasicQuaternion128, BasicQuaternion256:
			return typeSizeOfInternal(t, path) / 4
		}

	case TypeArray:
		elem := t.Array.Elem
		pop := typePathPush(path, elem)
		if path.Failure {
			return FailureAlignment
		}
		align := typeAlignOfInternal(elem, path)
		if pop {
			typePathPop(path)
		}
		return align

	case TypeEnumeratedArray:
		elem := t.EnumeratedArray.Elem
		pop := typePathPush(path, elem)
		if path.Failure {
			return FailureAlignment
		}
		align := typeAlignOfInternal(elem, path)
		if pop {
			typePathPop(path)
		}
		return align

	case TypeDynamicArray:
		return buildContext.IntSize

	case TypeFixedCapacityDynamicArray:
		elem := t.FixedCapacityDynamicArray.Elem
		pop := typePathPush(path, elem)
		if path.Failure {
			return FailureAlignment
		}
		align := typeAlignOfInternal(elem, path)
		if pop {
			typePathPop(path)
		}
		if align < buildContext.IntSize {
			return buildContext.IntSize
		}
		return align

	case TypeSlice:
		return buildContext.IntSize

	case TypeBitField:
		return typeAlignOfInternal(t.BitField.BackingType, path)

	case TypeTuple:
		max := int64(1)
		for i := range t.Tuple.Variables {
			align := typeAlignOfInternal(t.Tuple.Variables[i].Type, path)
			if max < align {
				max = align
			}
		}
		return max

	case TypeMap:
		return buildContext.PtrSize

	case TypeEnum:
		return typeAlignOfInternal(t.Enum.BaseType, path)

	case TypeUnion:
		if len(t.Union.Variants) == 0 {
			return 1
		}
		if t.Union.CustomAlign > 0 {
			if t.Union.CustomAlign < 1 {
				return 1
			}
			return t.Union.CustomAlign
		}
		max := int64(1)
		for i := range t.Union.Variants {
			variant := t.Union.Variants[i]
			pop := typePathPush(path, variant)
			if path.Failure {
				return FailureAlignment
			}
			align := typeAlignOfInternal(variant, path)
			if pop {
				typePathPop(path)
			}
			if max < align {
				max = align
			}
		}
		return max

	case TypeStruct:
		if t.Struct.CustomAlign > 0 {
			if t.Struct.CustomAlign < 1 {
				return 1
			}
			return t.Struct.CustomAlign
		}
		if t.Struct.IsPacked {
			return 1
		}
		max := int64(1)
		for i := range t.Struct.Fields {
			fieldType := t.Struct.Fields[i].Type
			pop := typePathPush(path, fieldType)
			if path.Failure {
				return FailureAlignment
			}
			align := typeAlignOfInternal(fieldType, path)
			if pop {
				typePathPop(path)
			}
			if max < align {
				max = align
			}
		}
		if t.Struct.CustomMinFieldAlign > 0 && max < t.Struct.CustomMinFieldAlign {
			max = t.Struct.CustomMinFieldAlign
		}
		if t.Struct.CustomMaxFieldAlign != 0 &&
			t.Struct.CustomMaxFieldAlign > t.Struct.CustomMinFieldAlign &&
			max > t.Struct.CustomMaxFieldAlign {
			max = t.Struct.CustomMaxFieldAlign
		}
		return max

	case TypeBitSet:
		if t.BitSet.Underlying != nil {
			return typeAlignOf(t.BitSet.Underlying)
		}
		bits := t.BitSet.Upper - t.BitSet.Lower + 1
		switch {
		case bits <= 8:
			return 1
		case bits <= 16:
			return 2
		case bits <= 32:
			return 4
		case bits <= 64:
			return 8
		case bits <= 128:
			return 16
		default:
			return 8
		}

	case TypeSimdVector:
		size := typeSizeOfInternal(t, path)
		np2 := nextPow2(size)
		align := np2
		if align < 1 {
			align = 1
		}
		if align > buildContext.MaxSimdAlign*2 {
			align = buildContext.MaxSimdAlign * 2
		}
		return align

	case TypeMatrix:
		return matrixAlignOf(t, path)

	case TypeSoaPointer:
		return buildContext.IntSize
	}

	size := typeSizeOfInternal(t, path)
	np2 := nextPow2(size)
	align := np2
	if align < 1 {
		align = 1
	}
	if align > buildContext.MaxAlign {
		align = buildContext.MaxAlign
	}
	return align
}

func typeSizeOfInternal(t *Type, path *TypePath) int64 {
	if t.Failure {
		return FailureSize
	}

	switch t.Kind {
	case TypeNamed:
		pop := typePathPush(path, t)
		if path.Failure {
			return FailureSize
		}
		size := typeSizeOfInternal(t.Named.Base, path)
		if pop {
			typePathPop(path)
		}
		return size

	case TypeBasic:
		kind := t.Basic.Kind
		size := t.Basic.Size
		if size > 0 {
			return size
		}
		switch kind {
		case BasicString:
			return 2 * buildContext.IntSize
		case BasicCstring:
			return buildContext.PtrSize
		case BasicString16:
			return 2 * buildContext.IntSize
		case BasicCstring16:
			return buildContext.PtrSize
		case BasicAny:
			return 16
		case BasicTypeid:
			return 8
		case BasicInt, BasicUint:
			return buildContext.IntSize
		case BasicUintptr, BasicRawptr:
			return buildContext.PtrSize
		}

	case TypePointer:
		return buildContext.PtrSize

	case TypeMultiPointer:
		return buildContext.PtrSize

	case TypeSoaPointer:
		return 2 * buildContext.IntSize

	case TypeArray:
		count := t.Array.Count
		if count == 0 {
			return 0
		}
		align := typeAlignOfInternal(t.Array.Elem, path)
		if path.Failure {
			return FailureSize
		}
		size := typeSizeOfInternal(t.Array.Elem, path)
		alignment := alignFormula(size, align)
		return alignment*(count-1) + size

	case TypeEnumeratedArray:
		count := t.EnumeratedArray.Count
		if count == 0 {
			return 0
		}
		align := typeAlignOfInternal(t.EnumeratedArray.Elem, path)
		if path.Failure {
			return FailureSize
		}
		size := typeSizeOfInternal(t.EnumeratedArray.Elem, path)
		alignment := alignFormula(size, align)
		return alignment*(count-1) + size

	case TypeSlice:
		return 2 * buildContext.IntSize

	case TypeDynamicArray:
		return 3*buildContext.IntSize + 2*buildContext.PtrSize

	case TypeFixedCapacityDynamicArray:
		capacity := t.FixedCapacityDynamicArray.Capacity
		elem := t.FixedCapacityDynamicArray.Elem
		align := typeAlignOfInternal(elem, path)
		if path.Failure {
			return FailureSize
		}
		if align < buildContext.IntSize {
			align = buildContext.IntSize
		}
		elemSize := typeSizeOf(elem)
		size := elemSize * capacity
		size = alignFormula(size, buildContext.IntSize)

		padding := size - elemSize*capacity
		t.FixedCapacityDynamicArray.PaddingNeeded = padding

		size += 1 * buildContext.IntSize
		size = alignFormula(size, align)
		return size

	case TypeMap:
		return (1 + 1 + 2) * buildContext.PtrSize

	case TypeTuple:
		count := len(t.Tuple.Variables)
		if count == 0 {
			return 0
		}
		align := typeAlignOfInternal(t, path)
		typeSetOffsets(t)
		lastIdx := count - 1
		size := t.Tuple.Offsets[lastIdx] + typeSizeOfInternal(t.Tuple.Variables[lastIdx].Type, path)
		return alignFormula(size, align)

	case TypeEnum:
		return typeSizeOfInternal(t.Enum.BaseType, path)

	case TypeUnion:
		if len(t.Union.Variants) == 0 {
			return 0
		}
		align := typeAlignOfInternal(t, path)
		if path.Failure {
			return FailureSize
		}
		max := int64(0)
		for i := range t.Union.Variants {
			variantType := t.Union.Variants[i]
			size := typeSizeOfInternal(variantType, path)
			if max < size {
				max = size
			}
		}
		var size int64
		if isTypeUnionMaybePointer(t) {
			size = max
			t.Union.TagSize = 0
			t.Union.VariantBlockSize = size
		} else {
			tagSize := unionTagSize(t)
			size = alignFormula(max, tagSize)
			t.Union.TagSize = int16(tagSize)
			t.Union.VariantBlockSize = size
			size += tagSize
		}
		return alignFormula(size, align)

	case TypeStruct:
		if t.Struct.IsRawUnion {
			count := len(t.Struct.Fields)
			align := typeAlignOfInternal(t, path)
			if path.Failure {
				return FailureSize
			}
			max := int64(0)
			for i := 0; i < count; i++ {
				size := typeSizeOfInternal(t.Struct.Fields[i].Type, path)
				if max < size {
					max = size
				}
			}
			return alignFormula(max, align)
		} else {
			count := len(t.Struct.Fields)
			if count == 0 {
				return 0
			}
			align := typeAlignOfInternal(t, path)
			if path.Failure {
				return FailureSize
			}
			{
				mutexLock(&t.Struct.OffsetMutex)
				if t.Struct.AreOffsetsBeingProcessed && t.Struct.Offsets == nil {
					typePathPrintIllegalCycle(path, len(path.Path)-1)
					mutexUnlock(&t.Struct.OffsetMutex)
					return FailureSize
				}
				mutexUnlock(&t.Struct.OffsetMutex)
			}
			typeSetOffsets(t)
			lastIdx := count - 1
			size := t.Struct.Offsets[lastIdx] + typeSizeOfInternal(t.Struct.Fields[lastIdx].Type, path)
			return alignFormula(size, align)
		}

	case TypeBitSet:
		if t.BitSet.Underlying != nil {
			return typeSizeOf(t.BitSet.Underlying)
		}
		bits := t.BitSet.Upper - t.BitSet.Lower + 1
		switch {
		case bits <= 8:
			return 1
		case bits <= 16:
			return 2
		case bits <= 32:
			return 4
		case bits <= 64:
			return 8
		case bits <= 128:
			return 16
		default:
			return 8
		}

	case TypeSimdVector:
		count := t.SimdVector.Count
		elem := t.SimdVector.Elem
		return count * typeSizeOfInternal(elem, path)

	case TypeMatrix:
		strideInBytes := matrixTypeStrideInBytes(t, path)
		if t.Matrix.IsRowMajor {
			return strideInBytes * t.Matrix.RowCount
		} else {
			return strideInBytes * t.Matrix.ColumnCount
		}

	case TypeBitField:
		return typeSizeOfInternal(t.BitField.BackingType, path)
	}

	return buildContext.PtrSize
}

func typeOffsetOf(t *Type, index int64, fieldType_ **Type) int64 {
	t = baseType(t)
	switch t.Kind {
	case TypeStruct:
		typeSetOffsets(t)
		if 0 <= index && index < int64(len(t.Struct.Fields)) {
			if fieldType_ != nil {
				*fieldType_ = t.Struct.Fields[index].Type
			}
			return t.Struct.Offsets[index]
		}

	case TypeTuple:
		typeSetOffsets(t)
		if 0 <= index && index < int64(len(t.Tuple.Variables)) {
			if fieldType_ != nil {
				*fieldType_ = t.Tuple.Variables[index].Type
			}
			offset := t.Tuple.Offsets[index]
			return offset
		}

	case TypeArray:
		if 0 <= index && index < t.Array.Count {
			return index * typeSizeOf(t.Array.Elem)
		}

	case TypeBasic:
		if t.Basic.Kind == BasicString {
			switch index {
			case 0:
				if fieldType_ != nil {
					*fieldType_ = tU8Ptr
				}
				return 0
			case 1:
				if fieldType_ != nil {
					*fieldType_ = tInt
				}
				return buildContext.IntSize
			}
		} else if t.Basic.Kind == BasicString16 {
			switch index {
			case 0:
				if fieldType_ != nil {
					*fieldType_ = tU16Ptr
				}
				return 0
			case 1:
				if fieldType_ != nil {
					*fieldType_ = tInt
				}
				return buildContext.IntSize
			}
		} else if t.Basic.Kind == BasicAny {
			switch index {
			case 0:
				if fieldType_ != nil {
					*fieldType_ = tRawptr
				}
				return 0
			case 1:
				if fieldType_ != nil {
					*fieldType_ = tTypeid
				}
				return 8
			default:
				gbAssertHandler("index > 1 for any type")
			}
		}

	case TypeSlice:
		switch index {
		case 0:
			if fieldType_ != nil {
				*fieldType_ = allocTypeMultiPointer(t.Slice.Elem)
			}
			return 0
		case 1:
			if fieldType_ != nil {
				*fieldType_ = tInt
			}
			return 1 * buildContext.IntSize
		}

	case TypeDynamicArray:
		switch index {
		case 0:
			if fieldType_ != nil {
				*fieldType_ = allocTypeMultiPointer(t.DynamicArray.Elem)
			}
			return 0
		case 1:
			if fieldType_ != nil {
				*fieldType_ = tInt
			}
			return 1 * buildContext.IntSize
		case 2:
			if fieldType_ != nil {
				*fieldType_ = tInt
			}
			return 2 * buildContext.IntSize
		case 3:
			if fieldType_ != nil {
				*fieldType_ = tAllocator
			}
			return 3 * buildContext.IntSize
		}

	case TypeFixedCapacityDynamicArray:
		switch index {
		case 0:
			if fieldType_ != nil {
				*fieldType_ = allocTypeArray(t.FixedCapacityDynamicArray.Elem, t.FixedCapacityDynamicArray.Capacity)
			}
			return 0
		case 1:
			if fieldType_ != nil {
				*fieldType_ = tInt
			}
			elemSize := typeSizeOf(t.FixedCapacityDynamicArray.Elem)
			offset := elemSize * t.FixedCapacityDynamicArray.Capacity
			offset = alignFormula(offset, buildContext.IntSize)
			return offset
		}

	case TypeUnion:
		if !isTypeUnionMaybePointer(t) {
			typeSizeOf(t)
			switch index {
			case -1:
				if fieldType_ != nil {
					*fieldType_ = unionTagType(t)
				}
				unionTagSize(t)
				return t.Union.VariantBlockSize
			}
		}
	}

	if index == 0 {
		return 0
	}
	gbAssertHandler("type_offset_of: unhandled type or invalid index")
	return 0
}

func areStructFieldsReordered(type_ *Type) bool {
	type_ = baseType(type_)
	if type_.Kind != TypeStruct {
		gbAssertHandler("are_struct_fields_reordered: expected struct type")
	}
	typeSetOffsets(type_)
	if len(type_.Struct.Fields) == 0 {
		return false
	}
	prevOffset := int64(0)
	for i := range type_.Struct.Fields {
		offset := type_.Struct.Offsets[i]
		if prevOffset > offset {
			return true
		}
		prevOffset = offset
	}
	return false
}

func structFieldsIndexByIncreasingOffset(allocator interface{}, type_ *Type) []int32 {
	type_ = baseType(type_)
	if type_.Kind != TypeStruct {
		gbAssertHandler("struct_fields_index_by_increasing_offset: expected struct type")
	}
	typeSetOffsets(type_)
	if len(type_.Struct.Fields) == 0 {
		return nil
	}
	indices := make([]int32, len(type_.Struct.Fields))

	prevOffset := int64(0)
	isOrdered := true
	for i := range indices {
		indices[i] = int32(i)
		offset := type_.Struct.Offsets[i]
		if isOrdered && prevOffset > offset {
			isOrdered = false
		}
		prevOffset = offset
	}
	if !isOrdered {
		n := len(indices)
		for i := 1; i < n; i++ {
			j := i
			for j > 0 && type_.Struct.Offsets[indices[j-1]] > type_.Struct.Offsets[indices[j]] {
				indices[j-1], indices[j] = indices[j], indices[j-1]
				j--
			}
		}
	}
	return indices
}

func typeSizeOfStructPretendIsPacked(ot *Type) int64 {
	if ot == nil {
		return 0
	}
	t := coreType(ot)
	if t.Kind != TypeStruct {
		return typeSizeOf(ot)
	}
	if t.Struct.IsPacked {
		return typeSizeOf(ot)
	}
	fields := t.Struct.Fields
	if len(fields) == 0 {
		return 0
	}
	size := int64(0)
	for i := range fields {
		size += typeSizeOf(fields[i].Type)
	}
	return alignFormula(size, 1)
}

func typeOffsetOfFromSelection(type_ *Type, sel Selection) int64 {
	if sel.Indirect {
		gbAssertHandler("type_offset_of_from_selection: indirect selection not supported")
	}
	t := type_
	offset := int64(0)
	for i := range sel.Index {
		index := sel.Index[i]
		t = baseType(t)
		offset += typeOffsetOf(t, int64(index), nil)
		if t.Kind == TypeStruct {
			t = t.Struct.Fields[index].Type
		} else if t.Kind == TypeArray {
			t = t.Array.Elem
		} else {
			switch t.Kind {
			case TypeBasic:
				if t.Basic.Kind == BasicString {
					switch index {
					case 0:
						t = tRawptr
					case 1:
						t = tInt
					}
				} else if t.Basic.Kind == BasicString16 {
					switch index {
					case 0:
						t = tRawptr
					case 1:
						t = tInt
					}
				} else if t.Basic.Kind == BasicAny {
					switch index {
					case 0:
						t = tRawptr
					case 1:
						t = tTypeid
					default:
						gbAssertHandler("index > 1")
					}
				}
			case TypeSlice:
				switch index {
				case 0:
					t = tRawptr
				case 1:
					t = tInt
				case 2:
					t = tInt
				}
			case TypeDynamicArray:
				switch index {
				case 0:
					t = tRawptr
				case 1:
					t = tInt
				case 2:
					t = tInt
				case 3:
					t = tAllocator
				}
			case TypeFixedCapacityDynamicArray:
				switch index {
				case 0:
					t = allocTypeArray(t.FixedCapacityDynamicArray.Elem, t.FixedCapacityDynamicArray.Capacity)
				case 1:
					t = tInt
				}
			}
		}
	}
	return offset
}
