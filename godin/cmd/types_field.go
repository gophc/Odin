package cmd

import "sync"

var (
	entityAnyDataOnce sync.Once
	entityAnyData     *Entity
	entityAnyID       *Entity

	entityQ64Once sync.Once
	entityQ64W    *Entity
	entityQ64X    *Entity
	entityQ64Y    *Entity
	entityQ64Z    *Entity

	entityQ128Once sync.Once
	entityQ128W    *Entity
	entityQ128X    *Entity
	entityQ128Y    *Entity
	entityQ128Z    *Entity

	entityQ256Once sync.Once
	entityQ256W    *Entity
	entityQ256X    *Entity
	entityQ256Y    *Entity
	entityQ256Z    *Entity

	entityUntypedQOnce sync.Once
	entityUntypedQW    *Entity
	entityUntypedQX    *Entity
	entityUntypedQY    *Entity
	entityUntypedQZ    *Entity
)

func lookupField(type_ *Type, fieldName InternedString, isType bool, allowBlankIdent ...bool) Selection {
	allowBlank := false
	if len(allowBlankIdent) > 0 {
		allowBlank = allowBlankIdent[0]
	}
	return lookupFieldWithSelection(type_, fieldName, isType, emptySelection, allowBlank)
}

func lookupFieldFromIndex(type_ *Type, index i64) Selection {
	if !isTypeStruct(type_) && !isTypeUnion(type_) && !isTypeTuple(type_) {
		gbAssertHandler("Assertion Failure", "is_type_struct(type) || is_type_union(type) || is_type_tuple(type)", "types_field.go", 0)
	}
	type_ = baseType(type_)

	var maxCount isize
	switch type_.Kind {
	case TypeStruct:
		type_.Struct.FieldsWaitSignal.Wait()
		maxCount = isize(len(type_.Struct.Fields))
	case TypeTuple:
		maxCount = isize(len(type_.Tuple.Variables))
	}

	if index >= i64(maxCount) {
		return emptySelection
	}

	switch type_.Kind {
	case TypeStruct:
		type_.Struct.FieldsWaitSignal.Wait()
		for i, f := range type_.Struct.Fields {
			if f.Kind == EntityVariable && f.Variable.FieldIndex == index {
				sel := makeSelection(f, []int32{int32(i)}, false)
				return sel
			}
		}
	case TypeTuple:
		for i, f := range type_.Tuple.Variables {
			if i64(i) == index {
				sel := makeSelection(f, []int32{int32(i)}, false)
				return sel
			}
		}
	}

	gbAssertHandler("Assertion Failure", "Illegal index", "types_field.go", 0)
	return emptySelection
}

func lookupFieldWithSelection(type_ *Type, fieldName InternedString, isType bool, sel Selection, allowBlankIdent bool) Selection {
	if type_ == nil {
		gbAssertHandler("Assertion Failure", "type_ != nullptr", "types_field.go", 0)
	}

	if !allowBlankIdent && fieldName.IsBlank() {
		return emptySelection
	}

	typ := typeDeref(type_)
	isPtr := typ != type_
	sel.Indirect = sel.Indirect || isPtr

	originalType := typ

	typ = baseType(typ)

	if isType {
		if hasTypeGotObjcClassAttribute(originalType) && originalType.Kind == TypeNamed {
			e := originalType.Named.TypeName
			if e.Kind != EntityTypeName {
				gbAssertHandler("Assertion Failure", "e->kind == Entity_TypeName", "types_field.go", 0)
			}
			if e.TypeName.ObjcMetadata != nil {
				md := e.TypeName.ObjcMetadata
				mutexLock(md.Mutex)
				func() {
					defer mutexUnlock(md.Mutex)
					for _, entry := range md.TypeEntries {
						if entry.Entity.Kind != EntityProcedure && entry.Entity.Kind != EntityProcGroup {
							gbAssertHandler("Assertion Failure", "entry.entity->kind == Entity_Procedure || entry.entity->kind == Entity_ProcGroup", "types_field.go", 0)
						}
						if entry.Interned == fieldName {
							sel.Entity = entry.Entity
							sel.PseudoField = true
							return
						}
					}
				}()
				if sel.Entity != nil {
					return sel
				}
			}
			if typ.Kind == TypeStruct {
				for _, f := range typ.Struct.Fields {
					if f.Flags&EntityFlagUsing != 0 {
						sel = lookupFieldWithSelection(f.Type, fieldName, isType, sel, allowBlankIdent)
						if sel.Entity != nil {
							return sel
						}
					}
				}
			}
		}

		if isTypeEnum(typ) {
			for _, f := range typ.Enum.Fields {
				if f.Kind != EntityConstant {
					gbAssertHandler("Assertion Failure", "f->kind == Entity_Constant", "types_field.go", 0)
				}
				str := entityInternedName(f)
				if fieldName == str {
					sel.Entity = f
					return sel
				}
			}
		}

		if typ.Kind == TypeStruct {
			s := typ.Struct.Scope
			if s != nil {
				found := scopeLookupCurrent(s, fieldName)
				if found != nil && found.Kind != EntityVariable {
					sel.Entity = found
					return sel
				}
			}
		} else if typ.Kind == TypeUnion {
			s := typ.Union.Scope
			if s != nil {
				found := scopeLookupCurrent(s, fieldName)
				if found != nil && found.Kind != EntityVariable {
					sel.Entity = found
					return sel
				}
			}
		} else if typ.Kind == TypeBitSet {
			return lookupFieldWithSelection(typ.BitSet.Elem, fieldName, true, sel, allowBlankIdent)
		}

		if typ.Kind == TypeGeneric && typ.Generic.Specialized != nil {
			specialized := typ.Generic.Specialized
			return lookupFieldWithSelection(specialized, fieldName, isType, sel, allowBlankIdent)
		}

	} else if typ.Kind == TypeUnion {
		// no-op in C++ code
	} else if typ.Kind == TypeStruct {
		if hasTypeGotObjcClassAttribute(originalType) && originalType.Kind == TypeNamed {
			e := originalType.Named.TypeName
			if e.Kind != EntityTypeName {
				gbAssertHandler("Assertion Failure", "e->kind == Entity_TypeName", "types_field.go", 0)
			}
			if e.TypeName.ObjcMetadata != nil {
				md := e.TypeName.ObjcMetadata
				mutexLock(md.Mutex)
				func() {
					defer mutexUnlock(md.Mutex)
					for _, entry := range md.ValueEntries {
						if entry.Entity.Kind != EntityProcedure && entry.Entity.Kind != EntityProcGroup {
							gbAssertHandler("Assertion Failure", "entry.entity->kind == Entity_Procedure || entry.entity->kind == Entity_ProcGroup", "types_field.go", 0)
						}
						if entry.Interned == fieldName {
							sel.Entity = entry.Entity
							sel.PseudoField = true
							return
						}
					}
				}()
				if sel.Entity != nil {
					return sel
				}
			}

			objcIvarType := e.TypeName.ObjcIvar
			if objcIvarType != nil {
				sel = lookupFieldWithSelection(objcIvarType, fieldName, false, sel, allowBlankIdent)
				if sel.Entity != nil {
					sel.PseudoField = true
					return sel
				}
			}
		}

		if isTypePolymorphic(typ) {
			return sel
		}

		typ.Struct.FieldsWaitSignal.Wait()
		if len(typ.Struct.Fields) > 0 {
			for i, f := range typ.Struct.Fields {
				if f.Kind != EntityVariable || (f.Flags&EntityFlagField) == 0 {
					continue
				}
				str := entityInternedName(f)
				if fieldName == str {
					selectionAddIndex(&sel, isize(i))
					sel.Entity = f
					return sel
				}

				if f.Flags&EntityFlagUsing != 0 {
					prevCount := len(sel.Index)
					prevIndirect := sel.Indirect
					selectionAddIndex(&sel, isize(i))

					sel = lookupFieldWithSelection(f.Type, fieldName, isType, sel, allowBlankIdent)

					if sel.Entity != nil {
						if isTypePointer(f.Type) {
							sel.Indirect = true
						}
						return sel
					}
					sel.Index = sel.Index[:prevCount]
					sel.Indirect = prevIndirect
				}
			}
		}

		isSOA := typ.Struct.SoaKind != StructSoaNone
		isSOAOfArray := isSOA && isTypeArray(typ.Struct.SoaElem)

		if isSOAOfArray {
			var mappedFieldName InternedString
			n := fieldName.String()
			switch n {
			case "r":
				mappedFieldName = stringInternerInsert("x")
			case "g":
				mappedFieldName = stringInternerInsert("y")
			case "b":
				mappedFieldName = stringInternerInsert("z")
			case "a":
				mappedFieldName = stringInternerInsert("w")
			}
			if mappedFieldName.String() != "" {
				return lookupFieldWithSelection(typ, mappedFieldName, isType, sel, allowBlankIdent)
			}
		}
	} else if typ.Kind == TypeBitField {
		for i, f := range typ.BitField.Fields {
			if f.Kind != EntityVariable || (f.Flags&EntityFlagField) == 0 {
				continue
			}
			str := entityInternedName(f)
			if fieldName == str {
				selectionAddIndex(&sel, isize(i))
				sel.Entity = f
				sel.IsBitField = true
				return sel
			}
		}
	} else if typ.Kind == TypeBasic {
		switch typ.Basic.Kind {
		case BasicAny:
			entityAnyDataOnce.Do(func() {
				entityAnyData = allocEntityField(nil, makeTokenIdent("data"), tRawptr, false, 0)
				entityAnyID = allocEntityField(nil, makeTokenIdent("id"), tTypeid, false, 1)
			})
			n := fieldName.String()
			if n == "data" {
				selectionAddIndex(&sel, 0)
				sel.Entity = entityAnyData
				return sel
			} else if n == "id" {
				selectionAddIndex(&sel, 1)
				sel.Entity = entityAnyID
				return sel
			}

		case BasicQuaternion64:
			entityQ64Once.Do(func() {
				entityQ64W = allocEntityField(nil, makeTokenIdent("w"), tF16, false, 3)
				entityQ64X = allocEntityField(nil, makeTokenIdent("x"), tF16, false, 0)
				entityQ64Y = allocEntityField(nil, makeTokenIdent("y"), tF16, false, 1)
				entityQ64Z = allocEntityField(nil, makeTokenIdent("z"), tF16, false, 2)
			})
			n := fieldName.String()
			if n == "w" {
				selectionAddIndex(&sel, 3)
				sel.Entity = entityQ64W
				return sel
			} else if n == "x" {
				selectionAddIndex(&sel, 0)
				sel.Entity = entityQ64X
				return sel
			} else if n == "y" {
				selectionAddIndex(&sel, 1)
				sel.Entity = entityQ64Y
				return sel
			} else if n == "z" {
				selectionAddIndex(&sel, 2)
				sel.Entity = entityQ64Z
				return sel
			}

		case BasicQuaternion128:
			entityQ128Once.Do(func() {
				entityQ128W = allocEntityField(nil, makeTokenIdent("w"), tF32, false, 3)
				entityQ128X = allocEntityField(nil, makeTokenIdent("x"), tF32, false, 0)
				entityQ128Y = allocEntityField(nil, makeTokenIdent("y"), tF32, false, 1)
				entityQ128Z = allocEntityField(nil, makeTokenIdent("z"), tF32, false, 2)
			})
			n := fieldName.String()
			if n == "w" {
				selectionAddIndex(&sel, 3)
				sel.Entity = entityQ128W
				return sel
			} else if n == "x" {
				selectionAddIndex(&sel, 0)
				sel.Entity = entityQ128X
				return sel
			} else if n == "y" {
				selectionAddIndex(&sel, 1)
				sel.Entity = entityQ128Y
				return sel
			} else if n == "z" {
				selectionAddIndex(&sel, 2)
				sel.Entity = entityQ128Z
				return sel
			}

		case BasicQuaternion256:
			entityQ256Once.Do(func() {
				entityQ256W = allocEntityField(nil, makeTokenIdent("w"), tF64, false, 3)
				entityQ256X = allocEntityField(nil, makeTokenIdent("x"), tF64, false, 0)
				entityQ256Y = allocEntityField(nil, makeTokenIdent("y"), tF64, false, 1)
				entityQ256Z = allocEntityField(nil, makeTokenIdent("z"), tF64, false, 2)
			})
			n := fieldName.String()
			if n == "w" {
				selectionAddIndex(&sel, 3)
				sel.Entity = entityQ256W
				return sel
			} else if n == "x" {
				selectionAddIndex(&sel, 0)
				sel.Entity = entityQ256X
				return sel
			} else if n == "y" {
				selectionAddIndex(&sel, 1)
				sel.Entity = entityQ256Y
				return sel
			} else if n == "z" {
				selectionAddIndex(&sel, 2)
				sel.Entity = entityQ256Z
				return sel
			}

		case BasicUntypedQuaternion:
			entityUntypedQOnce.Do(func() {
				entityUntypedQW = allocEntityField(nil, makeTokenIdent("w"), tUntypedFloat, false, 3)
				entityUntypedQX = allocEntityField(nil, makeTokenIdent("x"), tUntypedFloat, false, 0)
				entityUntypedQY = allocEntityField(nil, makeTokenIdent("y"), tUntypedFloat, false, 1)
				entityUntypedQZ = allocEntityField(nil, makeTokenIdent("z"), tUntypedFloat, false, 2)
			})
			n := fieldName.String()
			if n == "w" {
				selectionAddIndex(&sel, 3)
				sel.Entity = entityUntypedQW
				return sel
			} else if n == "x" {
				selectionAddIndex(&sel, 0)
				sel.Entity = entityUntypedQX
				return sel
			} else if n == "y" {
				selectionAddIndex(&sel, 1)
				sel.Entity = entityUntypedQY
				return sel
			} else if n == "z" {
				selectionAddIndex(&sel, 2)
				sel.Entity = entityUntypedQZ
				return sel
			}
		}

		return sel

	} else if typ.Kind == TypeDynamicArray {
		if tAllocator == nil {
			gbAssertHandler("Assertion Failure", "t_allocator != nullptr", "types_field.go", 0)
		}
		n := fieldName.String()
		if n == "allocator" {
			selectionAddIndex(&sel, 3)
			sel.Entity = allocEntityField(nil, makeTokenIdent("allocator"), tAllocator, false, 3)
			return sel
		}
	} else if typ.Kind == TypeMap {
		if tAllocator == nil {
			gbAssertHandler("Assertion Failure", "t_allocator != nullptr", "types_field.go", 0)
		}
		n := fieldName.String()
		if n == "allocator" {
			selectionAddIndex(&sel, 2)
			sel.Entity = allocEntityField(nil, makeTokenIdent("allocator"), tAllocator, false, 2)
			return sel
		}
	} else if typ.Kind == TypeArray {
		n := fieldName.String()
		elem := typ.Array.Elem

		if typ.Array.Count <= 4 {
			switch typ.Array.Count {
			case 4:
				if n == "w" || n == "a" {
					selectionAddIndex(&sel, 3)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 3)
					return sel
				}
				fallthrough
			case 3:
				if n == "z" || n == "b" {
					selectionAddIndex(&sel, 2)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 2)
					return sel
				}
				fallthrough
			case 2:
				if n == "y" || n == "g" {
					selectionAddIndex(&sel, 1)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 1)
					return sel
				}
				fallthrough
			case 1:
				if n == "x" || n == "r" {
					selectionAddIndex(&sel, 0)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 0)
					return sel
				}
			}
		}
	} else if typ.Kind == TypeSimdVector {
		n := fieldName.String()
		elem := typ.SimdVector.Elem

		if typ.SimdVector.Count <= 4 {
			switch typ.SimdVector.Count {
			case 4:
				if n == "w" || n == "a" {
					selectionAddIndex(&sel, 3)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 3)
					return sel
				}
				fallthrough
			case 3:
				if n == "z" || n == "b" {
					selectionAddIndex(&sel, 2)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 2)
					return sel
				}
				fallthrough
			case 2:
				if n == "y" || n == "g" {
					selectionAddIndex(&sel, 1)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 1)
					return sel
				}
				fallthrough
			case 1:
				if n == "x" || n == "r" {
					selectionAddIndex(&sel, 0)
					sel.Entity = allocEntityArrayElem(nil, makeTokenIdent(n), elem, 0)
					return sel
				}
			}
		}
	}

	return sel
}

func checkIsAssignableToUsingSubtype(src *Type, dst *Type, level isize, srcIsPtr bool, allowPolymorphic bool) isize {
	prevSrc := src
	src = typeDeref(src)
	if !srcIsPtr {
		srcIsPtr = src != prevSrc
	}
	src = baseType(src)

	if !isTypeStruct(src) {
		return 0
	}

	dstIsPolymorphic := isTypePolymorphic(dst)

	for _, f := range src.Struct.Fields {
		if f.Kind != EntityVariable || (f.Flags&EntityFlagsIsSubtype) == 0 {
			continue
		}
		if allowPolymorphic && dstIsPolymorphic {
			fb := baseType(typeDeref(f.Type))
			if fb.Kind == TypeStruct {
				if fb.Struct.PolymorphicParent == dst {
					return 1
				}
			}
		}

		if areTypesIdentical(f.Type, dst) {
			return level + 1
		}
		if srcIsPtr && isTypePointer(dst) {
			if areTypesIdentical(f.Type, typeDeref(dst)) {
				return level + 1
			}
		}
		nestedLevel := checkIsAssignableToUsingSubtype(f.Type, dst, level+1, srcIsPtr, allowPolymorphic)
		if nestedLevel > 0 {
			return nestedLevel
		}
	}

	return 0
}

func checkIsAssignableToUsingOffsetZeroSubtype(src *Type, dst *Type) bool {
	srcStruct := baseType(src)
	if !isTypeStruct(srcStruct) {
		return false
	}

	for i, f := range srcStruct.Struct.Fields {
		if f.Kind != EntityVariable || (f.Flags&EntityFlagsIsSubtype) == 0 {
			continue
		}

		var fieldType *Type
		offset := typeOffsetOf(srcStruct, isize(i), &fieldType)

		if offset != 0 {
			return false
		}

		if areTypesIdentical(fieldType, dst) {
			return true
		}

		if checkIsAssignableToUsingOffsetZeroSubtype(fieldType, dst) {
			return true
		}
	}

	return false
}

func isTypeSubtypeOf(src *Type, dst *Type) bool {
	if areTypesIdentical(src, dst) {
		return true
	}
	return 0 < checkIsAssignableToUsingSubtype(src, dst, 0, isTypePointer(src), false)
}

func isTypeSubtypeOfAndAllowPolymorphic(src *Type, dst *Type) bool {
	if areTypesIdentical(src, dst) {
		return true
	}
	return 0 < checkIsAssignableToUsingSubtype(src, dst, 0, isTypePointer(src), true)
}

func hasTypeGotObjcClassAttribute(t *Type) bool {
	return t.Kind == TypeNamed && t.Named.TypeName != nil && t.Named.TypeName.TypeName.ObjcClassName != ""
}

func isTypeObjcObject(t *Type) bool {
	return internalCheckIsAssignableTo(t, tObjcObject)
}

func isTypeObjcPtrToObject(t *Type) bool {
	elem := typeDeref(t)
	return elem != t && elem.Kind == TypeNamed && isTypeObjcObject(elem)
}

func getStructFieldType(t *Type, index isize) *Type {
	t = baseType(typeDeref(t))
	if t.Kind != TypeStruct {
		gbAssertHandler("Assertion Failure", "t->kind == Type_Struct", "types_field.go", 0)
	}
	return t.Struct.Fields[index].Type
}

func reduceTupleToSingleType(originalType *Type) *Type {
	if originalType != nil {
		t := coreType(originalType)
		if t.Kind == TypeTuple && len(t.Tuple.Variables) == 1 {
			return t.Tuple.Variables[0].Type
		}
	}
	return originalType
}

func allocTypeTupleFromFieldTypes(fieldTypes []*Type, isPacked bool, mustBeTuple bool) *Type {
	if len(fieldTypes) == 0 {
		return nil
	}
	if !mustBeTuple && len(fieldTypes) == 1 {
		return fieldTypes[0]
	}

	t := allocTypeTuple()
	t.Tuple.Variables = make([]*Entity, len(fieldTypes))

	var scope *Scope
	for i, ft := range fieldTypes {
		t.Tuple.Variables[i] = allocEntityParam(scope, blankToken, ft, false, false)
	}
	t.Tuple.IsPacked = isPacked

	return t
}

func allocTypeProcFromTypes(paramTypes []*Type, results *Type, isCVararg bool, callingConvention ProcCallingConvention) *Type {
	params := allocTypeTupleFromFieldTypes(paramTypes, false, true)
	var resultsCount isize
	if results != nil {
		if results.Kind != TypeTuple {
			results = allocTypeTupleFromFieldTypes([]*Type{results}, false, true)
		}
		resultsCount = isize(len(results.Tuple.Variables))
	}

	var scope *Scope
	t := allocTypeProc(scope, params, int32(len(paramTypes)), results, int32(resultsCount), false, callingConvention)
	t.Proc.CVararg = isCVararg
	return t
}

func typeInternalIndex(t *Type, index isize) *Type {
	bt := baseType(t)
	if bt == nil {
		return nil
	}

	switch bt.Kind {
	case TypeBasic:
		switch bt.Basic.Kind {
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
		case BasicString:
			if index == 0 {
				return tU8Ptr
			} else if index == 1 {
				return tInt
			}
			gbAssertHandler("Assertion Failure", "index == 0 || index == 1", "types_field.go", 0)
		case BasicString16:
			if index == 0 {
				return tU16Ptr
			} else if index == 1 {
				return tInt
			}
			gbAssertHandler("Assertion Failure", "index == 0 || index == 1", "types_field.go", 0)
		case BasicAny:
			if index == 0 {
				return tRawptr
			} else if index == 1 {
				return tTypeid
			}
			gbAssertHandler("Assertion Failure", "index == 0 || index == 1", "types_field.go", 0)
		}

	case TypeArray:
		return bt.Array.Elem
	case TypeEnumeratedArray:
		return bt.EnumeratedArray.Elem
	case TypeSimdVector:
		return bt.SimdVector.Elem
	case TypeSlice:
		if index == 0 {
			return tRawptr
		} else if index == 1 {
			return tInt
		}
		gbAssertHandler("Assertion Failure", "index == 0 || index == 1", "types_field.go", 0)
	case TypeDynamicArray:
		switch index {
		case 0:
			return tRawptr
		case 1:
			return tInt
		case 2:
			return tInt
		case 3:
			return tAllocator
		default:
			gbAssertHandler("Assertion Failure", "invalid raw dynamic array index", "types_field.go", 0)
		}
	case TypeFixedCapacityDynamicArray:
		switch index {
		case 0:
			return allocTypeArray(bt.FixedCapacityDynamicArray.Elem, bt.FixedCapacityDynamicArray.Capacity)
		case 1:
			return tInt
		default:
			gbAssertHandler("Assertion Failure", "invalid raw fixed capacity dynamic array index", "types_field.go", 0)
		}
	case TypeStruct:
		return getStructFieldType(bt, index)
	case TypeUnion:
		if index < isize(len(bt.Union.Variants)) {
			return bt.Union.Variants[index]
		}
		return unionTagType(bt)
	case TypeTuple:
		return bt.Tuple.Variables[index].Type
	case TypeMatrix:
		return bt.Matrix.Elem
	case TypeSoaPointer:
		if index == 0 {
			return tRawptr
		} else if index == 1 {
			return tInt
		}
		gbAssertHandler("Assertion Failure", "index == 0 || index == 1", "types_field.go", 0)
	case TypeMap:
		return typeInternalIndex(bt.Map.DebugMetadataType, index)
	case TypeBitField:
		return typeInternalIndex(bt.BitField.BackingType, index)
	case TypeGeneric:
		return typeInternalIndex(bt.Generic.Specialized, index)
	}

	gbAssertHandler("Assertion Failure", "Unhandled type", "types_field.go", 0)
	return nil
}
