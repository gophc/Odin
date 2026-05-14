package cmd

import (
	"strings"
	"unsafe"
)

// =============================================================================
// lb_clone_struct_type - clones a named struct type by creating a new anonymous struct
// =============================================================================
func lb_clone_struct_type(m *lbModule, t *Type) LLVMTypeRef {
	bt := base_type(t)
	if bt == nil || bt.Kind != TypeStruct {
		return 0
	}
	var elements []LLVMTypeRef
	for _, field := range bt.Struct.Fields {
		if field.Kind != EntityVariable {
			continue
		}
		fieldType := lb_type(m, field.Type)
		elements = append(elements, fieldType)
	}
	packed := bt.Struct.IsPacked
	st := LLVMStructCreateNamed(m.Ctx, "")
	LLVMStructSetBody(st, elements, uint(len(elements)), boolToLLVM(packed))
	return st
}

// =============================================================================
// lb_type_internal_for_procedures_raw - gets/creates LLVM function type for procedures
// =============================================================================
func lb_type_internal_for_procedures_raw(m *lbModule, t *Type) LLVMTypeRef {
	if t == nil {
		return 0
	}
	bt := base_type(t)
	if bt == nil || bt.Kind != TypeProc {
		return lb_type_internal(m, t)
	}

	key := uint64(uintptr(unsafe.Pointer(bt)))
	m.FuncRawTypesMutex.Lock()
	if existing, ok := m.FuncRawTypes[key]; ok {
		m.FuncRawTypesMutex.Unlock()
		return existing
	}
	m.FuncRawTypesMutex.Unlock()

	pt := bt.Proc
	var paramTypes []LLVMTypeRef
	var paramCount uint
	if pt.Params != nil {
		paramsBT := base_type(pt.Params)
		if paramsBT != nil && paramsBT.Kind == TypeTuple {
			paramCount = uint(len(paramsBT.Tuple.Variables))
			paramTypes = make([]LLVMTypeRef, 0, paramCount)
			for _, v := range paramsBT.Tuple.Variables {
				ptype := lb_type(m, v.Type)
				paramTypes = append(paramTypes, ptype)
			}
		}
	}

	var ret LLVMTypeRef
	returnIsDefined := false
	returnIsTuple := false

	if pt.Results != nil {
		resultBT := base_type(pt.Results)
		if resultBT != nil && resultBT.Kind == TypeTuple {
			vars := resultBT.Tuple.Variables
			if len(vars) == 0 {
				ret = LLVMVoidTypeInContext(m.Ctx)
			} else if len(vars) == 1 {
				ret = lb_type(m, vars[0].Type)
				returnIsDefined = true
			} else {
				returnIsTuple = true
			}
		} else if resultBT != nil {
			ret = lb_type(m, resultBT)
			returnIsDefined = true
		}
	}
	if ret == 0 {
		ret = LLVMVoidTypeInContext(m.Ctx)
	}

	if m.InternalTypeLevel == 0 {
		return 0
	}

	ft := lb_get_abi_info(m, paramTypes, paramCount, ret, returnIsDefined, returnIsTuple, pt.CallingConvention, bt)
	if ft == nil {
		return 0
	}
	llvmType := lb_function_type_to_llvm_raw(ft, pt.CVararg)

	m.FuncRawTypesMutex.Lock()
	m.FuncRawTypes[key] = llvmType
	m.FuncRawTypesMutex.Unlock()
	return llvmType
}

// =============================================================================
// lb_type_internal_union_block_type - determines the LLVM type for a union variant block
// =============================================================================
func lb_type_internal_union_block_type(m *lbModule, t *Type) LLVMTypeRef {
	if t == nil {
		return 0
	}
	bt := base_type(t)
	if bt == nil || bt.Kind != TypeUnion {
		return 0
	}
	ut := bt.Union
	variantCount := len(ut.Variants)
	if variantCount == 0 {
		return 0
	}
	if variantCount == 1 {
		v := ut.Variants[0]
		if is_type_internally_pointer_like(v) {
			return LLVMInt8TypeInContext(m.Ctx)
		}
		elemType := lb_type(m, v)
		if elemType == 0 {
			return 0
		}
		st := LLVMStructCreateNamed(m.Ctx, "")
		LLVMStructSetBody(st, []LLVMTypeRef{elemType}, 1, 0)
		return st
	}

	largestSize := int64(0)
	for _, v := range ut.Variants {
		sz := type_size_of(v)
		if sz > largestSize {
			largestSize = sz
		}
	}
	return LLVMArrayType2(LLVMInt8TypeInContext(m.Ctx), uint64(largestSize))
}

// =============================================================================
// lb_type_internal - maps Odin types to LLVM types (the big function)
// =============================================================================
func lb_type_internal(m *lbModule, t *Type) LLVMTypeRef {
	if t == nil || t.Kind == TypeInvalid {
		return 0
	}
	if t.Kind == TypeBasic {
		return lb_type_internal_basic(m, t)
	}

	m.InternalTypeLevel++
	defer func() { m.InternalTypeLevel-- }()

	bt := base_type(t)
	if bt == nil {
		return 0
	}

	switch bt.Kind {
	case TypeNamed:
		if bt.Named.Base != nil {
			return lb_type(m, bt.Named.Base)
		}
		return 0

	case TypePointer:
		elem := bt.Pointer.Elem
		if elem == nil {
			return LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
		}
		if elem.Kind == TypeGeneric {
			return LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
		}
		elemType := lb_type(m, elem)
		if elemType == 0 {
			elemType = LLVMInt8TypeInContext(m.Ctx)
		}
		return LLVMPointerType(elemType, 0)

	case TypeMultiPointer:
		return LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)

	case TypeArray:
		elem := bt.Array.Elem
		count := bt.Array.Count
		elemType := lb_type(m, elem)
		if elemType == 0 {
			elemType = LLVMInt8TypeInContext(m.Ctx)
		}
		if count < 0 {
			count = 0
		}
		return LLVMArrayType2(elemType, uint64(count))

	case TypeEnumeratedArray:
		elem := bt.EnumeratedArray.Elem
		count := bt.EnumeratedArray.Count
		elemType := lb_type(m, elem)
		if elemType == 0 {
			return 0
		}
		if count < 0 {
			count = 0
		}
		return LLVMArrayType2(elemType, uint64(count))

	case TypeSlice:
		elem := bt.Slice.Elem
		ptrType := LLVMPointerType(lb_type(m, elem), 0)
		lenType := intptr_type_in_context(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{ptrType, lenType}, 2, 0)

	case TypeDynamicArray:
		elem := bt.DynamicArray.Elem
		ptrType := LLVMPointerType(lb_type(m, elem), 0)
		intType := intptr_type_in_context(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{ptrType, intType, intType, ptrType}, 4, 0)

	case TypeFixedCapacityDynamicArray:
		elem := bt.FixedCapacityDynamicArray.Elem
		elemType := lb_type(m, elem)
		if elemType == 0 {
			return 0
		}
		count := bt.FixedCapacityDynamicArray.Capacity
		if count < 0 {
			count = 0
		}
		arrType := LLVMArrayType2(elemType, uint64(count))
		lenType := intptr_type_in_context(m.Ctx)

		padding := bt.FixedCapacityDynamicArray.PaddingNeeded
		if padding > 0 {
			padType := lb_type_padding_filler(m, padding, lb_alignof(elemType))
			if padType != 0 {
				st := LLVMStructCreateNamed(m.Ctx, "")
				LLVMStructSetBody(st, []LLVMTypeRef{arrType, padType, lenType}, 3, 0)
				return st
			}
		}
		st := LLVMStructCreateNamed(m.Ctx, "")
		LLVMStructSetBody(st, []LLVMTypeRef{arrType, lenType}, 2, 0)
		return st

	case TypeMap:
		init_map_internal_debug_types(bt)
		mapType := LLVMStructCreateNamed(m.Ctx, "")
		ptrType := LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
		uintptrType := intptr_type_in_context(m.Ctx)
		elements := []LLVMTypeRef{ptrType, ptrType, ptrType, uintptrType, uintptrType, uintptrType, ptrType}
		LLVMStructSetBody(mapType, elements, uint(len(elements)), 0)
		return mapType

	case TypeStruct:
		return lb_type_internal_struct(m, bt)

	case TypeUnion:
		return lb_type_internal_union(m, bt)

	case TypeEnum:
		base := bt.Enum.BaseType
		if base != nil {
			return lb_type(m, base)
		}
		return LLVMInt32TypeInContext(m.Ctx)

	case TypeTuple:
		vars := bt.Tuple.Variables
		if len(vars) == 0 {
			return LLVMVoidTypeInContext(m.Ctx)
		}
		if len(vars) == 1 {
			v := vars[0]
			if v != nil && v.Type != nil {
				return lb_type(m, v.Type)
			}
			return 0
		}
		elements := make([]LLVMTypeRef, len(vars))
		for i, v := range vars {
			elements[i] = lb_type(m, v.Type)
		}
		packed := bt.Tuple.IsPacked
		return LLVMStructTypeInContext(m.Ctx, elements, uint(len(elements)), boolToLLVM(packed))

	case TypeProc:
		return lb_type_internal_for_procedures_raw(m, bt)

	case TypeBitSet:
		intType := bit_set_to_int(bt)
		if intType != nil {
			return lb_type(m, intType)
		}
		return LLVMInt64TypeInContext(m.Ctx)

	case TypeSimdVector:
		elem := bt.SimdVector.Elem
		count := bt.SimdVector.Count
		elemType := lb_type(m, elem)
		if elemType == 0 {
			return 0
		}
		if count <= 0 {
			count = 1
		}
		return LLVMVectorType(elemType, uint(count))

	case TypeMatrix:
		elem := bt.Matrix.Elem
		elemType := lb_type(m, elem)
		if elemType == 0 {
			return 0
		}
		total := bt.Matrix.RowCount * bt.Matrix.ColumnCount
		if total <= 0 {
			total = 1
		}
		return LLVMVectorType(elemType, uint(total))

	case TypeSoaPointer:
		elem := bt.SoaPointer.Elem
		soaBT := base_type(elem)
		if soaBT != nil && soaBT.Kind == TypeStruct && soaBT.Struct.SoaKind != StructSoaNone {
			soaType := LLVMStructCreateNamed(m.Ctx, "")
			var soaElements []LLVMTypeRef
			for _, field := range soaBT.Struct.Fields {
				if field.Kind != EntityVariable {
					continue
				}
				ft := lb_type(m, field.Type)
				soaElements = append(soaElements, ft)
			}
			LLVMStructSetBody(soaType, soaElements, uint(len(soaElements)), 0)
			return LLVMPointerType(soaType, 0)
		}
		elemType := lb_type(m, elem)
		if elemType == 0 {
			return 0
		}
		return LLVMPointerType(elemType, 0)

	case TypeBitField:
		backing := bt.BitField.BackingType
		if backing != nil {
			return lb_type(m, backing)
		}
		return LLVMInt64TypeInContext(m.Ctx)

	case TypeGeneric:
		return 0
	}
	return 0
}

// =============================================================================
// lb_type_internal_basic - handles Type_Basic cases
// =============================================================================
func lb_type_internal_basic(m *lbModule, t *Type) LLVMTypeRef {
	if t == nil || t.Kind != TypeBasic {
		return 0
	}
	bk := t.Basic.Kind
	switch bk {
	case BasicInvalid:
		return 0
	case BasicLLVMBool:
		return LLVMInt1TypeInContext(m.Ctx)
	case BasicBool:
		return LLVMInt8TypeInContext(m.Ctx)
	case BasicB8:
		return LLVMInt8TypeInContext(m.Ctx)
	case BasicB16:
		return LLVMInt16TypeInContext(m.Ctx)
	case BasicB32:
		return LLVMInt32TypeInContext(m.Ctx)
	case BasicB64:
		return LLVMInt64TypeInContext(m.Ctx)
	case BasicI8, BasicU8:
		return LLVMInt8TypeInContext(m.Ctx)
	case BasicI16, BasicU16:
		return LLVMInt16TypeInContext(m.Ctx)
	case BasicI32, BasicU32, BasicRune:
		return LLVMInt32TypeInContext(m.Ctx)
	case BasicI64, BasicU64:
		return LLVMInt64TypeInContext(m.Ctx)
	case BasicI128, BasicU128:
		return LLVMInt128TypeInContext(m.Ctx)
	case BasicInt, BasicUint:
		return intptr_type_in_context(m.Ctx)
	case BasicUintptr, BasicRawptr:
		return intptr_type_in_context(m.Ctx)
	case BasicF16:
		return LLVMHalfTypeInContext(m.Ctx)
	case BasicF32:
		return LLVMFloatTypeInContext(m.Ctx)
	case BasicF64:
		return LLVMDoubleTypeInContext(m.Ctx)
	case BasicComplex32:
		elem := LLVMFloatTypeInContext(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{elem, elem}, 2, 0)
	case BasicComplex64:
		elem := LLVMDoubleTypeInContext(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{elem, elem}, 2, 0)
	case BasicComplex128:
		elem := LLVMInt128TypeInContext(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{elem, elem}, 2, 0)
	case BasicQuaternion64:
		elem := LLVMFloatTypeInContext(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{elem, elem, elem, elem}, 4, 0)
	case BasicQuaternion128:
		elem := LLVMDoubleTypeInContext(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{elem, elem, elem, elem}, 4, 0)
	case BasicString:
		ptr := LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
		lenType := intptr_type_in_context(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{ptr, lenType}, 2, 0)
	case BasicCstring:
		return LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
	case BasicString16:
		ptr := LLVMPointerType(LLVMInt16TypeInContext(m.Ctx), 0)
		lenType := intptr_type_in_context(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{ptr, lenType}, 2, 0)
	case BasicCstring16:
		return LLVMPointerType(LLVMInt16TypeInContext(m.Ctx), 0)
	case BasicAny:
		ptr := LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
		typeidType := intptr_type_in_context(m.Ctx)
		return LLVMStructTypeInContext(m.Ctx, []LLVMTypeRef{ptr, typeidType}, 2, 0)
	case BasicTypeid:
		return intptr_type_in_context(m.Ctx)
	}
	return 0
}

// =============================================================================
// lb_type_internal_struct - handles Type_Struct
// =============================================================================
func lb_type_internal_struct(m *lbModule, bt *Type) LLVMTypeRef {
	if bt == nil || bt.Kind != TypeStruct {
		return 0
	}
	st := bt.Struct

	typeSetOffsets(bt)

	var elements []LLVMTypeRef
	if st.SoaKind != StructSoaNone && st.SoaElem != nil {
		soaElemType := lb_type(m, st.SoaElem)
		if soaElemType == 0 {
			return 0
		}
		if st.SoaKind == StructSoaFixed {
			for _, field := range st.Fields {
				if field.Kind != EntityVariable {
					continue
				}
				elements = append(elements, soaElemType)
			}
		} else if st.SoaKind == StructSoaSlice {
			ptrType := LLVMPointerType(soaElemType, 0)
			lenType := intptr_type_in_context(m.Ctx)
			elements = []LLVMTypeRef{ptrType, lenType}
		} else if st.SoaKind == StructSoaDynamic {
			ptrType := LLVMPointerType(soaElemType, 0)
			intType := intptr_type_in_context(m.Ctx)
			elements = []LLVMTypeRef{ptrType, intType, intType}
		}
	} else {
		for _, field := range st.Fields {
			if field.Kind != EntityVariable {
				continue
			}
			ft := lb_type(m, field.Type)
			if ft == 0 {
				return 0
			}
			elements = append(elements, ft)
		}

		if !st.IsPacked {
			remapping := lb_get_struct_remapping(m, bt)
			if len(remapping) > 0 {
				reordered := make([]LLVMTypeRef, len(remapping))
				for i, idx := range remapping {
					reordered[i] = elements[idx]
				}
				elements = reordered
			}
		}
	}

	packed := st.IsPacked || st.IsRawUnion

	structName := ""
	if bt.Named.TypeName != nil {
		structName = lb_mangle_name(bt.Named.TypeName)
	}

	structType := LLVMStructCreateNamed(m.Ctx, structName)
	LLVMStructSetBody(structType, elements, uint(len(elements)), boolToLLVM(packed))
	return structType
}

// =============================================================================
// lb_type_internal_union - handles Type_Union
// =============================================================================
func lb_type_internal_union(m *lbModule, bt *Type) LLVMTypeRef {
	if bt == nil || bt.Kind != TypeUnion {
		return 0
	}
	ut := bt.Union
	variantCount := len(ut.Variants)

	if variantCount == 0 {
		st := LLVMStructCreateNamed(m.Ctx, "")
		LLVMStructSetBody(st, nil, 0, 0)
		return st
	}

	if variantCount == 1 {
		v := ut.Variants[0]
		if is_type_internally_pointer_like(v) {
			return lb_type(m, v)
		}
	}

	blockType := lb_type_internal_union_block_type(m, bt)
	if blockType == 0 {
		return 0
	}

	tagSize := ut.TagSize
	var tagType LLVMTypeRef
	if tagSize <= 0 {
		tagType = LLVMInt32TypeInContext(m.Ctx)
	} else {
		tagType = LLVMIntTypeInContext(m.Ctx, uint(tagSize*8))
	}

	align := type_align_of(bt)
	tagAlign := lb_alignof(tagType)
	needPadding := align > tagAlign

	var elements []LLVMTypeRef
	if needPadding {
		pad := align - tagAlign
		padType := lb_type_padding_filler(m, pad, tagAlign)
		if padType != 0 {
			elements = []LLVMTypeRef{tagType, padType, blockType}
		} else {
			elements = []LLVMTypeRef{tagType, blockType}
		}
	} else {
		elements = []LLVMTypeRef{tagType, blockType}
	}

	unionName := ""
	if bt.Named.TypeName != nil {
		unionName = lb_mangle_name(bt.Named.TypeName)
	}

	unionType := LLVMStructCreateNamed(m.Ctx, unionName)
	LLVMStructSetBody(unionType, elements, uint(len(elements)), 0)
	return unionType
}

// =============================================================================
// lb_type - thread-safe cached type lookup
// =============================================================================
func lb_type(m *lbModule, t *Type) LLVMTypeRef {
	if t == nil {
		return 0
	}
	key := uint64(uintptr(unsafe.Pointer(t)))
	m.TypesMutex.Lock()
	existing, ok := m.Types[key]
	m.TypesMutex.Unlock()
	if ok {
		return existing
	}

	llvmType := lb_type_internal(m, t)
	if llvmType != 0 {
		m.TypesMutex.Lock()
		m.Types[key] = llvmType
		m.TypesMutex.Unlock()
	}
	return llvmType
}

// =============================================================================
// lb_get_function_type
// =============================================================================
func lb_get_function_type(m *lbModule, t *Type) *lbFunctionType {
	if t == nil {
		return nil
	}
	bt := base_type(t)
	if bt == nil || bt.Kind != TypeProc {
		return nil
	}
	pt := bt.Proc

	var paramTypes []LLVMTypeRef
	paramCount := uint(0)
	if pt.Params != nil {
		paramsBT := base_type(pt.Params)
		if paramsBT != nil && paramsBT.Kind == TypeTuple {
			paramCount = uint(len(paramsBT.Tuple.Variables))
			paramTypes = make([]LLVMTypeRef, 0, paramCount)
			for _, v := range paramsBT.Tuple.Variables {
				paramTypes = append(paramTypes, lb_type(m, v.Type))
			}
		}
	}

	var ret LLVMTypeRef
	returnIsDefined := false
	returnIsTuple := false

	if pt.Results != nil {
		resultBT := base_type(pt.Results)
		if resultBT != nil && resultBT.Kind == TypeTuple {
			vars := resultBT.Tuple.Variables
			if len(vars) == 0 {
				ret = LLVMVoidTypeInContext(m.Ctx)
			} else if len(vars) == 1 {
				ret = lb_type(m, vars[0].Type)
				returnIsDefined = true
			} else {
				returnIsTuple = true
			}
		} else if resultBT != nil {
			ret = lb_type(m, resultBT)
			returnIsDefined = true
		}
	}
	if ret == 0 {
		ret = LLVMVoidTypeInContext(m.Ctx)
	}

	ft := lb_get_abi_info(m, paramTypes, paramCount, ret, returnIsDefined, returnIsTuple, pt.CallingConvention, bt)
	return ft
}

// =============================================================================
// lb_ensure_abi_function_type
// =============================================================================
func lb_ensure_abi_function_type(m *lbModule, p *lbProcedure) {
	if p.AbiFunctionType != nil {
		return
	}
	p.AbiFunctionType = lb_get_function_type(m, p.Type)
}

// =============================================================================
// lb_add_entity
// =============================================================================
func lb_add_entity(m *lbModule, e *Entity) lbValue {
	if e == nil {
		return lbValue{}
	}
	name := lb_get_entity_name(m, e)
	if name == "" {
		return lbValue{}
	}
	llvmType := lb_type(m, e.Type)
	if llvmType == 0 {
		return lbValue{}
	}
	global := LLVMAddGlobal(m.Mod, llvmType, name)
	val := lbValue{Value: global, Type: e.Type}
	m.ValuesMutex.Lock()
	m.Values[unsafe.Pointer(e)] = val
	m.ValuesMutex.Unlock()
	return val
}

// =============================================================================
// lb_add_member
// =============================================================================
func lb_add_member(m *lbModule, name string, typ *Type) lbValue {
	llvmType := lb_type(m, typ)
	if llvmType == 0 {
		return lbValue{}
	}
	global := LLVMAddGlobal(m.Mod, llvmType, name)
	val := lbValue{Value: global, Type: typ}
	m.Members[name] = val
	return val
}

// =============================================================================
// lb_add_procedure_value
// =============================================================================
func lb_add_procedure_value(m *lbModule, name string, typ *Type) lbValue {
	return lb_add_member(m, name, typ)
}

// =============================================================================
// lb_create_enum_attribute_with_type
// =============================================================================
func lb_create_enum_attribute_with_type(ctx LLVMContextRef, name string, typ LLVMTypeRef) LLVMAttributeRef {
	kindID := LLVMGetEnumAttributeKindForName(name, uint(len(name)))
	if kindID == 0 {
		return 0
	}
	return LLVMCreateTypeAttribute(ctx, kindID, typ)
}

// =============================================================================
// lb_create_enum_attribute
// =============================================================================
func lb_create_enum_attribute(ctx LLVMContextRef, name string, value ...u64) LLVMAttributeRef {
	kindID := LLVMGetEnumAttributeKindForName(name, uint(len(name)))
	if kindID == 0 {
		return 0
	}
	val := u64(0)
	if len(value) > 0 {
		val = value[0]
	}
	return LLVMCreateEnumAttribute(ctx, kindID, uint64(val))
}

// =============================================================================
// lb_create_string_attribute
// =============================================================================
func lb_create_string_attribute(ctx LLVMContextRef, key string, value string) LLVMAttributeRef {
	return LLVMCreateStringAttribute(ctx, key, uint(len(key)), value, uint(len(value)))
}

// =============================================================================
// lb_add_proc_attribute_at_index - by name and optional value
// =============================================================================
func lb_add_proc_attribute_at_index(p *lbProcedure, index isize, name string, value ...u64) {
	if p == nil || p.Value == 0 {
		return
	}
	attr := lb_create_enum_attribute(p.Module.Ctx, name, value...)
	if attr != 0 {
		LLVMAddAttributeAtIndex(p.Value, LLVMAttributeIndex(index), attr)
	}
}

// =============================================================================
// lb_add_proc_attribute_at_index_with_string
// =============================================================================
func lb_add_proc_attribute_at_index_with_string(p *lbProcedure, index isize, key string, value string) {
	if p == nil || p.Value == 0 {
		return
	}
	attr := lb_create_string_attribute(p.Module.Ctx, key, value)
	if attr != 0 {
		LLVMAddAttributeAtIndex(p.Value, LLVMAttributeIndex(index), attr)
	}
}

// =============================================================================
// lb_add_nocapture_proc_attribute_at_index
// =============================================================================
func lb_add_nocapture_proc_attribute_at_index(p *lbProcedure, index isize) {
	if p == nil || p.Value == 0 {
		return
	}
	attr := lb_create_enum_attribute(p.Module.Ctx, "nocapture")
	if attr != 0 {
		LLVMAddAttributeAtIndex(p.Value, LLVMAttributeIndex(index), attr)
	}
}

// =============================================================================
// lb_add_attribute_to_proc - adds attribute at FunctionIndex
// =============================================================================
func lb_add_attribute_to_proc(p *lbProcedure, name string, value ...u64) {
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FunctionIndex), name, value...)
}

// =============================================================================
// lb_proc_has_attribute
// =============================================================================
func lb_proc_has_attribute(p *lbProcedure, name string) bool {
	if p == nil || p.Value == 0 {
		return false
	}
	kindID := LLVMGetEnumAttributeKindForName(name, uint(len(name)))
	if kindID == 0 {
		return false
	}
	attr := LLVMGetAttributeAtIndex(p.Value, LLVMAttributeIndex_FunctionIndex)
	return attr != 0
}

// =============================================================================
// lb_apply_thread_local_model
// =============================================================================
func lb_apply_thread_local_model(val LLVMValueRef, tlsModel string) {
	var mode LLVMThreadLocalMode
	switch strings.ToLower(tlsModel) {
	case "general":
		mode = LLVMGeneralDynamicTLSModel
	case "local":
		mode = LLVMLocalDynamicTLSModel
	case "initial":
		mode = LLVMInitialExecTLSModel
	case "local_exec":
		mode = LLVMLocalExecTLSModel
	default:
		mode = LLVMGeneralDynamicTLSModel
	}
	LLVMSetThreadLocalMode(val, mode)
}

// =============================================================================
// lb_add_edge
// =============================================================================
func lb_add_edge(pred, succ *lbBlock) {
	if pred == nil || succ == nil {
		return
	}
	pred.Succs = append(pred.Succs, succ)
	succ.Preds = append(succ.Preds, pred)
}

// =============================================================================
// lb_create_block
// =============================================================================
func lb_create_block(p *lbProcedure, name string, append_ ...bool) *lbBlock {
	block := &lbBlock{
		Block:      LLVMCreateBasicBlockInContext(p.Module.Ctx, name),
		Scope:      p.CurrScope,
		ScopeIndex: isize(p.ScopeIndex),
		Preds:      make([]*lbBlock, 0),
		Succs:      make([]*lbBlock, 0),
	}

	doAppend := true
	if len(append_) > 0 {
		doAppend = append_[0]
	}
	if doAppend {
		LLVMAppendExistingBasicBlock(p.Value, block.Block)
		block.Appended = true
	}

	p.Blocks = append(p.Blocks, block)
	return block
}

// =============================================================================
// lb_emit_jump
// =============================================================================
func lb_emit_jump(p *lbProcedure, target_block *lbBlock) {
	if p.CurrBlock == nil || target_block == nil {
		return
	}
	lastInst := LLVMGetLastInstruction(p.CurrBlock.Block)
	if lastInst != 0 {
		oc := LLVMGetInstructionOpcode(lastInst)
		if oc == LLVMBr || oc == LLVMIndirectBr || oc == LLVMSwitch {
			return
		}
	}
	LLVMBuildBr(p.Builder, target_block.Block)
	lb_add_edge(p.CurrBlock, target_block)
}

// =============================================================================
// lb_emit_if
// =============================================================================
func lb_emit_if(p *lbProcedure, cond lbValue, true_block *lbBlock, false_block *lbBlock) {
	if p.CurrBlock == nil {
		return
	}
	lastInst := LLVMGetLastInstruction(p.CurrBlock.Block)
	if lastInst != 0 {
		oc := LLVMGetInstructionOpcode(lastInst)
		if oc == LLVMBr || oc == LLVMIndirectBr || oc == LLVMSwitch {
			return
		}
	}
	condVal := cond.Value
	if condVal == 0 {
		return
	}
	condType := LLVMTypeOf(condVal)
	if LLVMGetTypeKind(condType) != LLVMIntegerTypeKind || LLVMGetIntTypeWidth(condType) != 1 {
		condVal = LLVMBuildTruncOrBitCast(p.Builder, condVal, LLVMInt1TypeInContext(p.Module.Ctx), "")
	}
	LLVMBuildCondBr(p.Builder, condVal, true_block.Block, false_block.Block)
	lb_add_edge(p.CurrBlock, true_block)
	lb_add_edge(p.CurrBlock, false_block)
}

// =============================================================================
// OdinLLVMGetInternalElementType
// =============================================================================
func OdinLLVMGetInternalElementType(typ LLVMTypeRef) LLVMTypeRef {
	kind := LLVMGetTypeKind(typ)
	switch kind {
	case LLVMArrayTypeKind:
		return OdinLLVMGetArrayElementType(typ)
	case LLVMVectorTypeKind:
		return OdinLLVMGetVectorElementType(typ)
	case LLVMPointerTypeKind:
		return LLVMGetElementType(typ)
	}
	return 0
}

// =============================================================================
// OdinLLVMGetArrayElementType
// =============================================================================
func OdinLLVMGetArrayElementType(typ LLVMTypeRef) LLVMTypeRef {
	if LLVMGetTypeKind(typ) != LLVMArrayTypeKind {
		return 0
	}
	return LLVMGetElementType(typ)
}

// =============================================================================
// OdinLLVMGetVectorElementType
// =============================================================================
func OdinLLVMGetVectorElementType(typ LLVMTypeRef) LLVMTypeRef {
	if LLVMGetTypeKind(typ) != LLVMVectorTypeKind {
		return 0
	}
	return LLVMGetElementType(typ)
}

// =============================================================================
// OdinLLVMBuildTransmute - bitcast/memcpy-based transmute between types
// =============================================================================
func OdinLLVMBuildTransmute(p *lbProcedure, value lbValue, t *Type) lbValue {
	if p == nil || value.Value == 0 {
		return lbValue{}
	}
	m := p.Module
	srcType := value.Type
	dstType := t
	if srcType == nil || dstType == nil {
		return lbValue{Value: value.Value, Type: dstType}
	}

	srcSize := type_size_of(srcType)
	dstSize := type_size_of(dstType)
	if srcSize < 0 || dstSize < 0 {
		return lbValue{}
	}

	dstLLVMType := lb_type(m, dstType)
	srcLLVMType := lb_type(m, srcType)
	if dstLLVMType == 0 || srcLLVMType == 0 {
		return lbValue{}
	}

	if srcSize != dstSize {
		alloca := LLVMBuildAlloca(p.Builder, srcLLVMType, "")
		LLVMBuildStore(p.Builder, value.Value, alloca)
		bc := LLVMBuildBitCast(p.Builder, alloca, LLVMPointerType(dstLLVMType, 0), "")
		return lbValue{Value: LLVMBuildLoad2(p.Builder, dstLLVMType, bc, ""), Type: dstType}
	}

	if srcLLVMType == dstLLVMType {
		return lbValue{Value: value.Value, Type: dstType}
	}

	srcKind := LLVMGetTypeKind(srcLLVMType)
	dstKind := LLVMGetTypeKind(dstLLVMType)

	if srcKind == dstKind {
		if srcKind == LLVMIntegerTypeKind || srcKind == LLVMPointerTypeKind ||
			srcKind == LLVMFloatTypeKind || srcKind == LLVMHalfTypeKind || srcKind == LLVMDoubleTypeKind {
			return lbValue{Value: LLVMBuildBitCast(p.Builder, value.Value, dstLLVMType, ""), Type: dstType}
		}
		if srcKind == LLVMStructTypeKind || srcKind == LLVMArrayTypeKind || srcKind == LLVMVectorTypeKind {
			return lbValue{Value: LLVMBuildBitCast(p.Builder, value.Value, dstLLVMType, ""), Type: dstType}
		}
	}

	if srcKind == LLVMPointerTypeKind && dstKind == LLVMIntegerTypeKind {
		return lbValue{Value: LLVMBuildPtrToInt(p.Builder, value.Value, dstLLVMType, ""), Type: dstType}
	}
	if srcKind == LLVMIntegerTypeKind && dstKind == LLVMPointerTypeKind {
		return lbValue{Value: LLVMBuildIntToPtr(p.Builder, value.Value, dstLLVMType, ""), Type: dstType}
	}

	alloca := LLVMBuildAlloca(p.Builder, srcLLVMType, "")
	LLVMBuildStore(p.Builder, value.Value, alloca)
	bc := LLVMBuildBitCast(p.Builder, alloca, LLVMPointerType(dstLLVMType, 0), "")
	return lbValue{Value: LLVMBuildLoad2(p.Builder, dstLLVMType, bc, ""), Type: dstType}
}

// =============================================================================
// Helpers
// =============================================================================

func boolToLLVM(b bool) LLVMBool {
	if b {
		return 1
	}
	return 0
}

func intptr_type_in_context(ctx LLVMContextRef) LLVMTypeRef {
	if buildContext.PtrSize == 8 {
		return LLVMInt64TypeInContext(ctx)
	}
	return LLVMInt32TypeInContext(ctx)
}
