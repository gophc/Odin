package cmd

// =============================================================================
// lb_set_odin_rtti_section - sets the section to ".odinti" (not on darwin)
// =============================================================================
func lb_set_odin_rtti_section(value LLVMValueRef) {
	if buildContext.Metrics.OS != TargetOsDarwin {
		LLVMSetSection(value, ".odinti")
	}
}

// =============================================================================
// lb_type_info_index_pair - TypeInfoPair version
// =============================================================================
func lb_type_info_index_pair(info *CheckerInfo, pair TypeInfoPair, err_on_not_found ...bool) isize {
	errorOnNotFound := true
	if len(err_on_not_found) > 0 {
		errorOnNotFound = err_on_not_found[0]
	}

	index := type_info_index_pair(info, pair, errorOnNotFound)
	if index >= 0 {
		return index
	}
	if errorOnNotFound {
		gb_printf_err("NOT FOUND lb_type_info_index_pair:\n\t%s\n\t@ index %td\n\tmax count: %d\nFound:\n",
			type_to_string(pair.Type), index, PtrMapCount(info.MinDepTypeInfoIndexMap))
		for entry := range PtrMapIterate(info.MinDepTypeInfoIndexMap) {
			type_info_index := entry.Value
			gb_printf_err("\t%s\n", type_to_string(info.TypeInfoTypesHashMap[type_info_index].Type))
		}
		gb_assert_handler("Panic", 0, "...", "NOT FOUND")
	}
	return -1
}

// =============================================================================
// lb_type_info_index - Type* version
// =============================================================================
func lb_type_info_index(info *CheckerInfo, type_ *Type, err_on_not_found ...bool) isize {
	hash := TypeHashCanonicalType(type_)
	return lb_type_info_index_pair(info, TypeInfoPair{Type: type_, Hash: hash}, err_on_not_found...)
}

// =============================================================================
// lb_typeid_kind - returns the TypeidKind enum value for a given type
// =============================================================================
func lb_typeid_kind(m *lbModule, type_ *Type, id ...u64) u64 {
	gb_assert_handler("Assertion Failure", "!buildContext.NoRTTI", "llvm_backend_type_helpers.go", 0)
	type_ = default_type(type_)

	var idVal u64
	if len(id) > 0 {
		idVal = id[0]
	} else {
		idVal = 0
	}
	if idVal == 0 {
		idVal = u64(lb_type_info_index(m.Info, type_))
	}

	kind := u64(TypeidInvalid)
	bt := base_type(type_)
	tk := bt.Kind
	switch tk {
	case TypeBasic:
		flags := bt.Basic.Flags
		if flags&BasicFlagBoolean != 0 {
			kind = u64(TypeidBoolean)
		}
		if flags&BasicFlagInteger != 0 {
			kind = u64(TypeidInteger)
		}
		if flags&BasicFlagUnsigned != 0 {
			kind = u64(TypeidInteger)
		}
		if flags&BasicFlagFloat != 0 {
			kind = u64(TypeidFloat)
		}
		if flags&BasicFlagComplex != 0 {
			kind = u64(TypeidComplex)
		}
		if flags&BasicFlagPointer != 0 {
			kind = u64(TypeidPointer)
		}
		if flags&BasicFlagString != 0 {
			kind = u64(TypeidString)
		}
		if flags&BasicFlagRune != 0 {
			kind = u64(TypeidRune)
		}
		if bt.Basic.Kind == BasicTypeid {
			kind = u64(TypeidTypeId)
		}
	case TypePointer:
		kind = u64(TypeidPointer)
	case TypeMultiPointer:
		kind = u64(TypeidMultiPointer)
	case TypeArray:
		kind = u64(TypeidArray)
	case TypeMatrix:
		kind = u64(TypeidMatrix)
	case TypeEnumeratedArray:
		kind = u64(TypeidEnumeratedArray)
	case TypeSlice:
		kind = u64(TypeidSlice)
	case TypeDynamicArray:
		kind = u64(TypeidDynamicArray)
	case TypeMap:
		kind = u64(TypeidMap)
	case TypeStruct:
		kind = u64(TypeidStruct)
	case TypeEnum:
		kind = u64(TypeidEnum)
	case TypeUnion:
		kind = u64(TypeidUnion)
	case TypeTuple:
		kind = u64(TypeidTuple)
	case TypeProc:
		kind = u64(TypeidProcedure)
	case TypeBitSet:
		kind = u64(TypeidBitSet)
	case TypeSimdVector:
		kind = u64(TypeidSimdVector)
	case TypeSoaPointer:
		kind = u64(TypeidSoaPointer)
	case TypeBitField:
		kind = u64(TypeidBitField)
	case TypeFixedCapacityDynamicArray:
		kind = u64(TypeidFixedCapacityDynamicArray)
	}
	return kind
}

// =============================================================================
// lb_typeid - returns the typeid hash as an LLVM constant
// =============================================================================
func lb_typeid(m *lbModule, type_ *Type) lbValue {
	gb_assert_handler("Assertion Failure", "!buildContext.NoRTTI", "llvm_backend_type_helpers.go", 0)
	type_ = default_type(type_)

	data := TypeHashCanonicalType(type_)
	gb_assert_handler("Assertion Failure", "data != 0", "llvm_backend_type_helpers.go", 0)

	res := lbValue{}
	res.Value = LLVMConstInt(lb_type(m, t_typeid), uint64(data), 0)
	res.Type = t_typeid
	return res
}

// =============================================================================
// lb_type_info - loads type info pointer from global array
// =============================================================================
func lb_type_info(p *lbProcedure, type_ *Type) lbValue {
	gb_assert_handler("Assertion Failure", "!buildContext.NoRTTI", "llvm_backend_type_helpers.go", 0)
	type_ = default_type(type_)
	m := p.Module

	index := lb_type_info_index(m.Info, type_)
	gb_assert_handler("Assertion Failure", "index >= 0", "llvm_backend_type_helpers.go", 0)

	global := lb_global_type_info_data_ptr(m)
	ptr := lb_emit_array_epi(p, global, index)
	return lb_emit_load(p, ptr)
}

// =============================================================================
// lb_global_type_info_data_ptr - returns the global type info data pointer
// =============================================================================
func lb_global_type_info_data_ptr(m *lbModule) lbValue {
	return lb_find_value_from_entity(m, lb_global_type_info_data_entity)
}

// =============================================================================
// lb_get_procedure_raw_type
// =============================================================================
func lb_get_procedure_raw_type(m *lbModule, type_ *Type) LLVMTypeRef {
	return lb_type_internal_for_procedures_raw(m, type_)
}

// =============================================================================
// lb_const_array_epi - GEP into constant array at index, returns pointer
// =============================================================================
func lb_const_array_epi(m *lbModule, value lbValue, index isize) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_pointer(value.Type)", "llvm_backend_type_helpers.go", 0)
	type_ := type_deref(value.Type)

	indices := [2]LLVMValueRef{
		LLVMConstInt(lb_type(m, t_int), 0, 0),
		LLVMConstInt(lb_type(m, t_int), uint64(index), 0),
	}
	llvm_type := lb_type(m, type_)
	res := lbValue{}
	ptr := base_array_type(type_)
	res.Type = alloc_type_pointer(ptr)
	gb_assert_handler("Assertion Failure", "LLVMIsConstant(value.Value)", "llvm_backend_type_helpers.go", 0)
	res.Value = LLVMConstGEP2(llvm_type, value.Value, indices[:], 2)
	return res
}

// =============================================================================
// lb_type_info_member_types_offset
// =============================================================================
func lb_type_info_member_types_offset(m *lbModule, count isize, offset_ ...*i64) lbValue {
	gb_assert_handler("Assertion Failure", "m == &m.Gen.DefaultModule", "llvm_backend_type_helpers.go", 0)
	if len(offset_) > 0 && offset_[0] != nil {
		*offset_[0] = lb_global_type_info_member_types_index
	}
	offset := lb_const_array_epi(m, lb_global_type_info_member_types.Addr, lb_global_type_info_member_types_index)
	lb_global_type_info_member_types_index += i32(count)
	return offset
}

// =============================================================================
// lb_type_info_member_names_offset
// =============================================================================
func lb_type_info_member_names_offset(m *lbModule, count isize, offset_ ...*i64) lbValue {
	gb_assert_handler("Assertion Failure", "m == &m.Gen.DefaultModule", "llvm_backend_type_helpers.go", 0)
	if len(offset_) > 0 && offset_[0] != nil {
		*offset_[0] = lb_global_type_info_member_names_index
	}
	offset := lb_const_array_epi(m, lb_global_type_info_member_names.Addr, lb_global_type_info_member_names_index)
	lb_global_type_info_member_names_index += i32(count)
	return offset
}

// =============================================================================
// lb_type_info_member_offsets_offset
// =============================================================================
func lb_type_info_member_offsets_offset(m *lbModule, count isize, offset_ ...*i64) lbValue {
	gb_assert_handler("Assertion Failure", "m == &m.Gen.DefaultModule", "llvm_backend_type_helpers.go", 0)
	if len(offset_) > 0 && offset_[0] != nil {
		*offset_[0] = lb_global_type_info_member_offsets_index
	}
	offset := lb_const_array_epi(m, lb_global_type_info_member_offsets.Addr, lb_global_type_info_member_offsets_index)
	lb_global_type_info_member_offsets_index += i32(count)
	return offset
}

// =============================================================================
// lb_type_info_member_usings_offset
// =============================================================================
func lb_type_info_member_usings_offset(m *lbModule, count isize, offset_ ...*i64) lbValue {
	gb_assert_handler("Assertion Failure", "m == &m.Gen.DefaultModule", "llvm_backend_type_helpers.go", 0)
	if len(offset_) > 0 && offset_[0] != nil {
		*offset_[0] = lb_global_type_info_member_usings_index
	}
	offset := lb_const_array_epi(m, lb_global_type_info_member_usings.Addr, lb_global_type_info_member_usings_index)
	lb_global_type_info_member_usings_index += i32(count)
	return offset
}

// =============================================================================
// lb_type_info_member_tags_offset
// =============================================================================
func lb_type_info_member_tags_offset(m *lbModule, count isize, offset_ ...*i64) lbValue {
	gb_assert_handler("Assertion Failure", "m == &m.Gen.DefaultModule", "llvm_backend_type_helpers.go", 0)
	if len(offset_) > 0 && offset_[0] != nil {
		*offset_[0] = lb_global_type_info_member_tags_index
	}
	offset := lb_const_array_epi(m, lb_global_type_info_member_tags.Addr, lb_global_type_info_member_tags_index)
	lb_global_type_info_member_tags_index += i32(count)
	return offset
}
