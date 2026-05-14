package cmd

import "unsafe"

// =============================================================================
// Type Info Data Setup
// Translates llvm_backend_type.i.cpp lines 143-908
// =============================================================================

// lb_setup_modified_types_for_type_info creates modified Type_Info struct types
// with per-variant Union field types. Each variant type gets a modified version
// that includes: variant_data + padding + tag
func lb_setup_modified_types_for_type_info(m *lbModule) {
	bt := base_type(t_type_info)
	if bt == nil || bt.Kind != Type_Struct {
		return
	}

	variantField := bt.Struct.Fields[4]
	variantUnionType := base_type(variantField.Type)
	if variantUnionType == nil || variantUnionType.Kind != Type_Union {
		return
	}

	variants := variantUnionType.Union.Variants
	if len(variants) == 0 {
		return
	}

	unionBlockSize := variantUnionType.Union.VariantBlockSize
	if unionBlockSize <= 0 {
		return
	}

	for _, variantType := range variants {
		if variantType == nil {
			continue
		}
		variantSize := typeSizeOf(variantType)
		if variantSize < 0 {
			continue
		}
		paddingSize := unionBlockSize - variantSize
		if paddingSize < 0 {
			paddingSize = 0
		}

		variantLLVM := lb_type(m, variantType)
		tagLLVM := lb_type(m, t_type_info_enum_value)

		var modifiedElements []LLVMTypeRef
		modifiedElements = append(modifiedElements, variantLLVM)
		if paddingSize > 0 {
			paddingLLVM := LLVMArrayType(LLVMInt8TypeInContext(m.Ctx), uint64(paddingSize))
			modifiedElements = append(modifiedElements, paddingLLVM)
		}
		modifiedElements = append(modifiedElements, tagLLVM)

		modifiedName := "ti." + type_to_string(variantType)
		modifiedType := LLVMStructCreateNamed(m.Ctx, modifiedName)
		LLVMStructSetBody(modifiedType, modifiedElements, uint(len(modifiedElements)), 0)
	}
}

// lb_setup_type_info_data_giant_array is the primary function that creates all
// type info entries in the global type info array.
func lb_setup_type_info_data_giant_array(m *lbModule) {
	info := m.Info
	if info == nil {
		return
	}

	typeInfoHash := info.TypeInfoTypesHashMap
	if len(typeInfoHash) == 0 {
		return
	}

	typeInfoLLVMType := lb_type(m, t_type_info)

	typeInfoData := make([]LLVMValueRef, 0)
	memberTypesData := make([]LLVMValueRef, 0)
	memberNamesData := make([]LLVMValueRef, 0)
	memberOffsetsData := make([]LLVMValueRef, 0)
	memberUsingsData := make([]LLVMValueRef, 0)
	memberTagsData := make([]LLVMValueRef, 0)

	entryGlobals := make(map[int64]LLVMValueRef)

	for _, pair := range typeInfoHash {
		type_ := pair.Type
		if type_ == nil {
			continue
		}
		bt := base_type(type_)
		if bt == nil {
			continue
		}

		name := type_to_string(bt)
		entryGlobal := LLVMAddGlobal(m.Mod, typeInfoLLVMType, "typeinfo$"+name)
		lb_make_global_private_const(entryGlobal)
		lb_set_odin_rtti_section(entryGlobal)

		sizeVal := lb_const_int(m, t_int, u64(typeSizeOf(bt))).Value
		alignVal := lb_const_int(m, t_int, u64(typeAlignOf(bt))).Value
		flags := typeInfoFlagsOfType(bt)
		flagsVal := lb_const_int(m, t_u32, u64(flags)).Value
		typeidKindVal := lb_const_int(m, t_type_info_enum_value, lb_typeid_kind(m, type_)).Value

		variantType := variant_type_for_kind(m, bt)
		variantValue := build_variant_data(m, bt, variantType,
			&memberTypesData, &memberNamesData, &memberOffsetsData, &memberUsingsData, &memberTagsData,
			entryGlobals)

		if variantValue == 0 {
			variantValue = LLVMConstNull(lb_type(m, t_type_info_named))
		}

		typeInfoElements := []LLVMValueRef{
			sizeVal,
			alignVal,
			flagsVal,
			typeidKindVal,
			variantValue,
		}

		typeInfoInit := llvm_const_named_struct(m, t_type_info, typeInfoElements, isize(len(typeInfoElements)))
		LLVMSetInitializer(entryGlobal, typeInfoInit)

		index := lb_type_info_index(info, type_, false)
		if index >= 0 {
			for isize(len(typeInfoData)) <= index {
				typeInfoData = append(typeInfoData, 0)
			}
			typeInfoData[index] = entryGlobal
			entryGlobals[int64(index)] = entryGlobal
		}
	}

	init_giant_arrays(m, typeInfoData, memberTypesData, memberNamesData,
		memberOffsetsData, memberUsingsData, memberTagsData)
}

// lb_setup_type_info_data is the entry point for type info data setup.
func lb_setup_type_info_data(m *lbModule) {
	if buildContext.NoRtti {
		return
	}

	typeTable := lb_find_runtime_value(m, "type_table")
	if typeTable.Value == 0 {
		return
	}

	lb_setup_type_info_data_giant_array(m)
}

// =============================================================================
// Helpers
// =============================================================================

func variant_type_for_kind(m *lbModule, bt *Type) *Type {
	if bt == nil {
		return nil
	}
	switch bt.Kind {
	case Type_Named:
		return t_type_info_named
	case Type_Basic:
		return basic_type_info_variant(bt.Basic.Kind)
	case Type_Pointer:
		return t_type_info_pointer
	case Type_MultiPointer:
		return t_type_info_multi_pointer
	case Type_SoaPointer:
		return t_type_info_soa_pointer
	case Type_Array:
		return t_type_info_array
	case Type_EnumeratedArray:
		return t_type_info_enumerated_array
	case Type_DynamicArray:
		return t_type_info_dynamic_array
	case Type_FixedCapacityDynamicArray:
		return t_type_info_fixed_capacity_dynamic_array
	case Type_Slice:
		return t_type_info_slice
	case Type_Proc:
		return t_type_info_procedure
	case Type_Tuple:
		return t_type_info_parameters
	case Type_Enum:
		return t_type_info_enum
	case Type_Union:
		return t_type_info_union
	case Type_Struct:
		return t_type_info_struct
	case Type_Map:
		return t_type_info_map
	case Type_BitSet:
		return t_type_info_bit_set
	case Type_SimdVector:
		return t_type_info_simd_vector
	case Type_Matrix:
		return t_type_info_matrix
	case Type_BitField:
		return t_type_info_bit_field
	}
	return t_type_info_named
}

func basic_type_info_variant(kind BasicKind) *Type {
	switch kind {
	case BasicBool, BasicB8, BasicB16, BasicB32, BasicB64:
		return t_type_info_boolean
	case BasicI8, BasicI16, BasicI32, BasicI64, BasicI128,
		BasicU8, BasicU16, BasicU32, BasicU64, BasicU128,
		BasicInt, BasicUint, BasicUintptr,
		BasicI16le, BasicI32le, BasicI64le, BasicI128le,
		BasicU16le, BasicU32le, BasicU64le, BasicU128le,
		BasicI16be, BasicI32be, BasicI64be, BasicI128be,
		BasicU16be, BasicU32be, BasicU64be, BasicU128be:
		return t_type_info_integer
	case BasicRune:
		return t_type_info_rune
	case BasicF16, BasicF32, BasicF64,
		BasicF16le, BasicF32le, BasicF64le,
		BasicF16be, BasicF32be, BasicF64be:
		return t_type_info_float
	case BasicComplex32, BasicComplex64, BasicComplex128:
		return t_type_info_complex
	case BasicQuaternion64, BasicQuaternion128, BasicQuaternion256:
		return t_type_info_quaternion
	case BasicString, BasicCstring:
		return t_type_info_string
	case BasicAny:
		return t_type_info_any
	case BasicTypeid:
		return t_type_info_typeid
	case BasicRawptr:
		return t_type_info_pointer
	}
	return t_type_info_boolean
}

func build_basic_variant_vals(m *lbModule, bt *Type) ([]LLVMValueRef, *Type) {
	vt := basic_type_info_variant(bt.Basic.Kind)
	if vt == nil {
		return nil, nil
	}

	switch bt.Basic.Kind {
	case BasicBool, BasicB8, BasicB16, BasicB32, BasicB64:
		return nil, vt

	case BasicI8, BasicI16, BasicI32, BasicI64, BasicI128,
		BasicU8, BasicU16, BasicU32, BasicU64, BasicU128,
		BasicInt, BasicUint, BasicUintptr,
		BasicI16le, BasicI32le, BasicI64le, BasicI128le,
		BasicU16le, BasicU32le, BasicU64le, BasicU128le,
		BasicI16be, BasicI32be, BasicI64be, BasicI128be,
		BasicU16be, BasicU32be, BasicU64be, BasicU128be:
		bitSize := lb_const_int(m, t_int, u64(8*typeSizeOf(bt))).Value
		isUnsigned := (bt.Basic.Flags & BasicFlagUnsigned) != 0
		unsignedVal := lb_const_bool(m, t_bool, isUnsigned).Value
		return []LLVMValueRef{bitSize, unsignedVal}, vt

	case BasicRune:
		return nil, vt

	case BasicF16, BasicF32, BasicF64,
		BasicF16le, BasicF32le, BasicF64le,
		BasicF16be, BasicF32be, BasicF64be:
		bitSize := lb_const_int(m, t_int, u64(8*typeSizeOf(bt))).Value
		return []LLVMValueRef{bitSize}, vt

	case BasicComplex32, BasicComplex64, BasicComplex128:
		bitSize := lb_const_int(m, t_int, u64(8*typeSizeOf(bt))).Value
		return []LLVMValueRef{bitSize}, vt

	case BasicQuaternion64, BasicQuaternion128, BasicQuaternion256:
		bitSize := lb_const_int(m, t_int, u64(8*typeSizeOf(bt))).Value
		return []LLVMValueRef{bitSize}, vt

	case BasicString:
		isCString := lb_const_bool(m, t_bool, false).Value
		encoding := lb_const_int(m, t_type_info_string_encoding_kind, 0).Value
		return []LLVMValueRef{isCString, encoding}, vt

	case BasicCstring:
		isCString := lb_const_bool(m, t_bool, true).Value
		encoding := lb_const_int(m, t_type_info_string_encoding_kind, 0).Value
		return []LLVMValueRef{isCString, encoding}, vt

	case BasicAny, BasicTypeid, BasicRawptr:
		return nil, vt
	}
	return nil, vt
}

func make_string_val(m *lbModule, s string) LLVMValueRef {
	str := String{
		Data: unsafe.StringData(s),
		Len:  isize(len(s)),
	}
	return lb_const_string(m, str).Value
}

func build_variant_data(m *lbModule, bt *Type, variantType *Type,
	memberTypesData *[]LLVMValueRef,
	memberNamesData *[]LLVMValueRef,
	memberOffsetsData *[]LLVMValueRef,
	memberUsingsData *[]LLVMValueRef,
	memberTagsData *[]LLVMValueRef,
	entryGlobals map[int64]LLVMValueRef) LLVMValueRef {

	if variantType == nil {
		return 0
	}

	switch bt.Kind {
	case Type_Named:
		nameStr := type_to_string(bt)
		nameVal := make_string_val(m, nameStr)
		baseTypePtr := make_type_info_ptr(m, bt.Named.Base, entryGlobals)
		vals := []LLVMValueRef{nameVal, baseTypePtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Basic:
		vals, vt := build_basic_variant_vals(m, bt)
		if vt != nil && len(vals) > 0 {
			return llvm_const_named_struct(m, vt, vals, isize(len(vals)))
		}
		if vt != nil {
			return LLVMConstNull(lb_type(m, vt))
		}
		return 0

	case Type_Pointer:
		elemPtr := make_type_info_ptr(m, bt.Pointer.Elem, entryGlobals)
		vals := []LLVMValueRef{elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_MultiPointer:
		elemPtr := make_type_info_ptr(m, bt.MultiPointer.Elem, entryGlobals)
		vals := []LLVMValueRef{elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_SoaPointer:
		elemPtr := make_type_info_ptr(m, bt.SoaPointer.Elem, entryGlobals)
		vals := []LLVMValueRef{elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Array:
		elemPtr := make_type_info_ptr(m, bt.Array.Elem, entryGlobals)
		countVal := lb_const_int(m, t_int, u64(bt.Array.Count)).Value
		vals := []LLVMValueRef{elemPtr, countVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_EnumeratedArray:
		ea := bt.EnumeratedArray
		elemPtr := make_type_info_ptr(m, ea.Elem, entryGlobals)
		indexPtr := make_type_info_ptr(m, ea.Index, entryGlobals)
		countVal := lb_const_int(m, t_int, u64(ea.Count)).Value
		minVal := lb_const_value(m, t_type_info_enum_value, *ea.MinValue).Value
		maxVal := lb_const_value(m, t_type_info_enum_value, *ea.MaxValue).Value
		vals := []LLVMValueRef{elemPtr, indexPtr, countVal, minVal, maxVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_DynamicArray:
		elemPtr := make_type_info_ptr(m, bt.DynamicArray.Elem, entryGlobals)
		vals := []LLVMValueRef{elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_FixedCapacityDynamicArray:
		elemPtr := make_type_info_ptr(m, bt.FixedCapacityDynamicArray.Elem, entryGlobals)
		capacityVal := lb_const_int(m, t_int, u64(bt.FixedCapacityDynamicArray.Capacity)).Value
		vals := []LLVMValueRef{elemPtr, capacityVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Slice:
		elemPtr := make_type_info_ptr(m, bt.Slice.Elem, entryGlobals)
		vals := []LLVMValueRef{elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Proc:
		pt := bt.Proc
		var paramTypesArr, resultTypesArr LLVMValueRef
		paramCount := int64(0)
		resultCount := int64(0)

		if pt.Params != nil {
			paramBT := base_type(pt.Params)
			if paramBT != nil && paramBT.Kind == Type_Tuple {
				paramCount = int64(len(paramBT.Tuple.Variables))
				paramList := make([]LLVMValueRef, paramCount)
				for pi, pv := range paramBT.Tuple.Variables {
					paramList[pi] = make_type_info_ptr(m, pv.Type, entryGlobals)
				}
				paramTypesArr = LLVMConstArray(lb_type(m, t_type_info_ptr), paramList, uint(paramCount))
			}
		}
		if pt.Results != nil {
			resultBT := base_type(pt.Results)
			if resultBT != nil && resultBT.Kind == Type_Tuple {
				resultCount = int64(len(resultBT.Tuple.Variables))
				resultList := make([]LLVMValueRef, resultCount)
				for ri, rv := range resultBT.Tuple.Variables {
					resultList[ri] = make_type_info_ptr(m, rv.Type, entryGlobals)
				}
				resultTypesArr = LLVMConstArray(lb_type(m, t_type_info_ptr), resultList, uint(resultCount))
			}
		}

		paramTypesSlice := llvm_const_slice_internal(m, paramTypesArr, lb_const_int(m, t_int, u64(paramCount)).Value)
		resultTypesSlice := llvm_const_slice_internal(m, resultTypesArr, lb_const_int(m, t_int, u64(resultCount)).Value)
		variadicVal := lb_const_bool(m, t_bool, pt.Variadic).Value
		cVarargVal := lb_const_bool(m, t_bool, pt.CVararg).Value

		vals := []LLVMValueRef{paramTypesSlice, resultTypesSlice, variadicVal, cVarargVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Tuple:
		tup := bt.Tuple
		fieldCount := int64(len(tup.Variables))
		return build_parameters_variant(m, bt, tup, fieldCount,
			memberTypesData, memberNamesData, memberOffsetsData, memberUsingsData, memberTagsData,
			entryGlobals)

	case Type_Enum:
		enum := bt.Enum
		fieldCount := int64(len(enum.Fields))
		tPtr := make_type_info_ptr(m, enum.BaseType, entryGlobals)
		minVal := lb_const_value(m, t_type_info_enum_value, *enum.MinValue).Value
		maxVal := lb_const_value(m, t_type_info_enum_value, *enum.MaxValue).Value

		var memberTypes, memberNames LLVMValueRef
		if fieldCount > 0 {
			mTypes := make([]LLVMValueRef, fieldCount)
			mNames := make([]LLVMValueRef, fieldCount)
			for fi, f := range enum.Fields {
				mTypes[fi] = make_type_info_ptr(m, f.Type, entryGlobals)
				mNames[fi] = make_string_val(m, goStr(f.Token.String))
			}
			memberTypes = LLVMConstArray(lb_type(m, t_type_info_ptr), mTypes, uint(fieldCount))
			memberNames = LLVMConstArray(lb_type(m, t_string), mNames, uint(fieldCount))
			*memberTypesData = append(*memberTypesData, memberTypes)
			*memberNamesData = append(*memberNamesData, memberNames)
		}

		typesSlice := llvm_const_slice_internal(m, memberTypes, lb_const_int(m, t_int, u64(fieldCount)).Value)
		namesSlice := llvm_const_slice_internal(m, memberNames, lb_const_int(m, t_int, u64(fieldCount)).Value)
		vals := []LLVMValueRef{typesSlice, namesSlice, tPtr, minVal, maxVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Union:
		un := bt.Union
		variantCount := int64(len(un.Variants))

		var memberTypes, memberNames LLVMValueRef
		if variantCount > 0 {
			mTypes := make([]LLVMValueRef, variantCount)
			for vi, v := range un.Variants {
				mTypes[vi] = make_type_info_ptr(m, v, entryGlobals)
			}
			memberTypes = LLVMConstArray(lb_type(m, t_type_info_ptr), mTypes, uint(variantCount))
			memberNames = LLVMConstNull(LLVMArrayType(lb_type(m, t_string), uint64(variantCount)))
			*memberTypesData = append(*memberTypesData, memberTypes)
		}

		typesSlice := llvm_const_slice_internal(m, memberTypes, lb_const_int(m, t_int, u64(variantCount)).Value)
		namesSlice := llvm_const_slice_internal(m, memberNames, lb_const_int(m, t_int, u64(variantCount)).Value)
		tagSizeVal := lb_const_int(m, t_int, u64(int64(un.TagSize))).Value
		isMaybe := boolToLLVM(un.Kind == UnionTypeNoNil || un.Kind == UnionTypeSharedNil)
		customAlign := lb_const_int(m, t_int, u64(un.CustomAlign)).Value
		vals := []LLVMValueRef{typesSlice, namesSlice, tagSizeVal, isMaybe, customAlign}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Struct:
		st := bt.Struct

		var rawVarFields []int
		var varEntityCount int64
		for ri, f := range st.Fields {
			if f.Kind == Entity_Variable {
				rawVarFields = append(rawVarFields, ri)
				varEntityCount++
			}
		}

		var memberTypes, memberNames, memberOffsets, memberUsings, memberTags LLVMValueRef
		if varEntityCount > 0 {
			mTypes := make([]LLVMValueRef, varEntityCount)
			mNames := make([]LLVMValueRef, varEntityCount)
			mOffsets := make([]LLVMValueRef, varEntityCount)
			mUsings := make([]LLVMValueRef, varEntityCount)
			mTags := make([]LLVMValueRef, varEntityCount)
			for fi, rawIdx := range rawVarFields {
				f := st.Fields[rawIdx]
				mTypes[fi] = make_type_info_ptr(m, f.Type, entryGlobals)
				mNames[fi] = make_string_val(m, goStr(f.Token.String))
				mOffsets[fi] = lb_const_int(m, t_int, u64(typeOffsetOf(bt, int64(rawIdx), nil))).Value
				mUsings[fi] = lb_const_bool(m, t_bool, (f.Flags&EntityFlag_Using) != 0).Value
				tagStr := ""
				if rawIdx < len(st.Tags) {
					tagStr = st.Tags[rawIdx]
				}
				mTags[fi] = make_string_val(m, tagStr)
			}
			tp := lb_type(m, t_type_info_ptr)
			ts := lb_type(m, t_string)
			ti := lb_type(m, t_int)
			tb := lb_type(m, t_bool)
			memberTypes = LLVMConstArray(tp, mTypes, uint(varEntityCount))
			memberNames = LLVMConstArray(ts, mNames, uint(varEntityCount))
			memberOffsets = LLVMConstArray(ti, mOffsets, uint(varEntityCount))
			memberUsings = LLVMConstArray(tb, mUsings, uint(varEntityCount))
			memberTags = LLVMConstArray(ts, mTags, uint(varEntityCount))
			*memberTypesData = append(*memberTypesData, memberTypes)
			*memberNamesData = append(*memberNamesData, memberNames)
			*memberOffsetsData = append(*memberOffsetsData, memberOffsets)
			*memberUsingsData = append(*memberUsingsData, memberUsings)
			*memberTagsData = append(*memberTagsData, memberTags)
		}

		typesSlice := llvm_const_slice_internal(m, memberTypes, lb_const_int(m, t_int, u64(varEntityCount)).Value)
		namesSlice := llvm_const_slice_internal(m, memberNames, lb_const_int(m, t_int, u64(varEntityCount)).Value)
		offsetsSlice := llvm_const_slice_internal(m, memberOffsets, lb_const_int(m, t_int, u64(varEntityCount)).Value)
		usingsSlice := llvm_const_slice_internal(m, memberUsings, lb_const_int(m, t_int, u64(varEntityCount)).Value)
		tagsSlice := llvm_const_slice_internal(m, memberTags, lb_const_int(m, t_int, u64(varEntityCount)).Value)
		isRawUnion := boolToLLVM(st.IsRawUnion)
		customAlign := lb_const_int(m, t_int, u64(st.CustomAlign)).Value
		vals := []LLVMValueRef{typesSlice, namesSlice, offsetsSlice, usingsSlice, tagsSlice, isRawUnion, customAlign}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Map:
		mapType := bt
		init_map_internal_debug_types(mapType)
		mapInfoPtr := lb_gen_map_info_ptr(m, mapType)
		equalProc := lb_equal_proc_for_type(m, mapType.Map.Key)
		hasherProc := lb_hasher_proc_for_type(m, mapType.Map.Key)
		vals := []LLVMValueRef{mapInfoPtr.Value, equalProc.Value, hasherProc.Value}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_BitSet:
		bs := bt.BitSet
		bitsCount := bs.Upper - bs.Lower + 1
		bitsCountVal := lb_const_int(m, t_int, u64(bitsCount)).Value
		lowerVal := lb_const_int(m, t_int, u64(bs.Lower)).Value
		upperVal := lb_const_int(m, t_int, u64(bs.Upper)).Value
		underlyingPtr := make_type_info_ptr(m, bs.Underlying, entryGlobals)
		elemPtr := make_type_info_ptr(m, bs.Elem, entryGlobals)
		vals := []LLVMValueRef{bitsCountVal, lowerVal, upperVal, underlyingPtr, elemPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_SimdVector:
		sv := bt.SimdVector
		elemPtr := make_type_info_ptr(m, sv.Elem, entryGlobals)
		countVal := lb_const_int(m, t_int, u64(sv.Count)).Value
		vals := []LLVMValueRef{elemPtr, countVal}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_Matrix:
		mt := bt.Matrix
		elemPtr := make_type_info_ptr(m, mt.Elem, entryGlobals)
		rowCountVal := lb_const_int(m, t_int, u64(mt.RowCount)).Value
		columnCountVal := lb_const_int(m, t_int, u64(mt.ColumnCount)).Value
		stride := matrixTypeStrideInElems(bt)
		strideVal := lb_const_int(m, t_int, u64(stride)).Value
		isRowMajor := boolToLLVM(mt.IsRowMajor)
		vals := []LLVMValueRef{elemPtr, rowCountVal, columnCountVal, strideVal, isRowMajor}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))

	case Type_BitField:
		bf := bt.BitField
		fieldCount := int64(len(bf.Fields))
		var memberTypes, memberNames LLVMValueRef
		if fieldCount > 0 {
			mTypes := make([]LLVMValueRef, fieldCount)
			mNames := make([]LLVMValueRef, fieldCount)
			for fi, f := range bf.Fields {
				mTypes[fi] = make_type_info_ptr(m, f.Type, entryGlobals)
				mNames[fi] = make_string_val(m, goStr(f.Token.String))
			}
			memberTypes = LLVMConstArray(lb_type(m, t_type_info_ptr), mTypes, uint(fieldCount))
			memberNames = LLVMConstArray(lb_type(m, t_string), mNames, uint(fieldCount))
			*memberTypesData = append(*memberTypesData, memberTypes)
			*memberNamesData = append(*memberNamesData, memberNames)
		}
		typesSlice := llvm_const_slice_internal(m, memberTypes, lb_const_int(m, t_int, u64(fieldCount)).Value)
		namesSlice := llvm_const_slice_internal(m, memberNames, lb_const_int(m, t_int, u64(fieldCount)).Value)
		backingPtr := make_type_info_ptr(m, bf.BackingType, entryGlobals)
		vals := []LLVMValueRef{typesSlice, namesSlice, backingPtr}
		return llvm_const_named_struct(m, variantType, vals, isize(len(vals)))
	}

	return 0
}

func build_parameters_variant(m *lbModule, tupleType *Type, tup TypeTuple, fieldCount int64,
	memberTypesData *[]LLVMValueRef,
	memberNamesData *[]LLVMValueRef,
	memberOffsetsData *[]LLVMValueRef,
	memberUsingsData *[]LLVMValueRef,
	memberTagsData *[]LLVMValueRef,
	entryGlobals map[int64]LLVMValueRef) LLVMValueRef {

	var memberTypes, memberNames LLVMValueRef

	if fieldCount > 0 {
		mTypes := make([]LLVMValueRef, fieldCount)
		mNames := make([]LLVMValueRef, fieldCount)
		mOffsets := make([]LLVMValueRef, fieldCount)
		mUsings := make([]LLVMValueRef, fieldCount)
		mTags := make([]LLVMValueRef, fieldCount)

		for vi, v := range tup.Variables {
			mTypes[vi] = make_type_info_ptr(m, v.Type, entryGlobals)
			mNames[vi] = make_string_val(m, goStr(v.Token.String))
			mOffsets[vi] = lb_const_int(m, t_int, u64(typeOffsetOf(tupleType, int64(vi), nil))).Value
			mUsings[vi] = lb_const_bool(m, t_bool, (v.Flags&EntityFlag_Using) != 0).Value
			mTags[vi] = make_string_val(m, "")
		}

		tp := lb_type(m, t_type_info_ptr)
		ts := lb_type(m, t_string)
		ti := lb_type(m, t_int)
		tb := lb_type(m, t_bool)

		memberTypes = LLVMConstArray(tp, mTypes, uint(fieldCount))
		memberNames = LLVMConstArray(ts, mNames, uint(fieldCount))
		memberOffsetsArr := LLVMConstArray(ti, mOffsets, uint(fieldCount))
		memberUsingsArr := LLVMConstArray(tb, mUsings, uint(fieldCount))
		memberTagsArr := LLVMConstArray(ts, mTags, uint(fieldCount))

		*memberTypesData = append(*memberTypesData, memberTypes)
		*memberNamesData = append(*memberNamesData, memberNames)
		*memberOffsetsData = append(*memberOffsetsData, memberOffsetsArr)
		*memberUsingsData = append(*memberUsingsData, memberUsingsArr)
		*memberTagsData = append(*memberTagsData, memberTagsArr)
	}

	typesSlice := llvm_const_slice_internal(m, memberTypes, lb_const_int(m, t_int, u64(fieldCount)).Value)
	namesSlice := llvm_const_slice_internal(m, memberNames, lb_const_int(m, t_int, u64(fieldCount)).Value)
	vals := []LLVMValueRef{typesSlice, namesSlice}
	return llvm_const_named_struct(m, t_type_info_parameters, vals, isize(len(vals)))
}

func make_type_info_ptr(m *lbModule, t *Type, entryGlobals map[int64]LLVMValueRef) LLVMValueRef {
	if t == nil {
		return LLVMConstNull(lb_type(m, t_type_info_ptr))
	}
	index := lb_type_info_index(m.Info, t, false)
	if index < 0 {
		return LLVMConstNull(lb_type(m, t_type_info_ptr))
	}
	if global, ok := entryGlobals[int64(index)]; ok {
		return LLVMConstPointerCast(global, lb_type(m, t_type_info_ptr))
	}
	return LLVMConstNull(lb_type(m, t_type_info_ptr))
}

func init_giant_arrays(m *lbModule,
	typeInfoData []LLVMValueRef,
	memberTypesData, memberNamesData, memberOffsetsData, memberUsingsData, memberTagsData []LLVMValueRef) {

	count := isize(len(typeInfoData))
	if count == 0 {
		return
	}

	typeInfoArray := LLVMConstArray(lb_type(m, t_type_info), typeInfoData, uint(count))
	giantArray := lb_generate_global_array(m, t_type_info, int64(count), "type_info_data", 0)
	LLVMSetInitializer(giantArray.Value, typeInfoArray)
	lb_make_global_private_const(giantArray.Value)
	lb_set_odin_rtti_section(giantArray.Value)
	lb_global_type_info_data_entity = &Entity{Kind: Entity_Variable, Type: giantArray.Type}
	lb_global_type_info_data_ptr = lb_addr(giantArray)

	if len(memberTypesData) > 0 {
		arr := LLVMConstArray(lb_type(m, t_type_info_ptr), memberTypesData, uint(len(memberTypesData)))
		g := lb_generate_global_array(m, t_type_info_ptr, int64(len(memberTypesData)), "type_info_member_types", 0)
		LLVMSetInitializer(g.Value, arr)
		lb_make_global_private_const(g.Value)
		lb_set_odin_rtti_section(g.Value)
		lb_global_type_info_member_types = lb_addr(g)
	}
	if len(memberNamesData) > 0 {
		arr := LLVMConstArray(lb_type(m, t_string), memberNamesData, uint(len(memberNamesData)))
		g := lb_generate_global_array(m, t_string, int64(len(memberNamesData)), "type_info_member_names", 0)
		LLVMSetInitializer(g.Value, arr)
		lb_make_global_private_const(g.Value)
		lb_set_odin_rtti_section(g.Value)
		lb_global_type_info_member_names = lb_addr(g)
	}
	if len(memberOffsetsData) > 0 {
		arr := LLVMConstArray(lb_type(m, t_int), memberOffsetsData, uint(len(memberOffsetsData)))
		g := lb_generate_global_array(m, t_int, int64(len(memberOffsetsData)), "type_info_member_offsets", 0)
		LLVMSetInitializer(g.Value, arr)
		lb_make_global_private_const(g.Value)
		lb_set_odin_rtti_section(g.Value)
		lb_global_type_info_member_offsets = lb_addr(g)
	}
	if len(memberUsingsData) > 0 {
		arr := LLVMConstArray(lb_type(m, t_bool), memberUsingsData, uint(len(memberUsingsData)))
		g := lb_generate_global_array(m, t_bool, int64(len(memberUsingsData)), "type_info_member_usings", 0)
		LLVMSetInitializer(g.Value, arr)
		lb_make_global_private_const(g.Value)
		lb_set_odin_rtti_section(g.Value)
		lb_global_type_info_member_usings = lb_addr(g)
	}
	if len(memberTagsData) > 0 {
		arr := LLVMConstArray(lb_type(m, t_string), memberTagsData, uint(len(memberTagsData)))
		g := lb_generate_global_array(m, t_string, int64(len(memberTagsData)), "type_info_member_tags", 0)
		LLVMSetInitializer(g.Value, arr)
		lb_make_global_private_const(g.Value)
		lb_set_odin_rtti_section(g.Value)
		lb_global_type_info_member_tags = lb_addr(g)
	}
}
