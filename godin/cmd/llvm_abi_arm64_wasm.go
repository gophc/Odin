package cmd

// Translated from src/cipp/llvm_abi.i.cpp namespaces:
//   lbAbiArm64 (lines 958-1141)
//   lbAbiWasm  (lines 1142-1322)
//   lbAbiArm32 (lines 1323-1397)
//   lbAbiRiscv64 (lines 1398-1585)

// ---- lbAbiArm64 ----

func lbAbiArm64_abi_info(m *lbModule, arg_types []LLVMTypeRef, arg_count uint32, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	c := m.Ctx
	ft := new(lbFunctionType)
	ft.Ctx = c
	ft.Args = lbAbiArm64_compute_arg_types(c, arg_types, arg_count)
	ft.Ret = lbAbiArm64_compute_return_type(ft, c, return_type, return_is_defined, return_is_tuple)
	ft.CallingConvention = calling_convention
	return ft
}

func lbAbiArm64_is_register(typ LLVMTypeRef) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMIntegerTypeKind, LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind, LLVMPointerTypeKind:
		return true
	}
	return false
}

func lbAbiArm64_non_struct(c LLVMContextRef, typ LLVMTypeRef) lbArgType {
	var attr LLVMAttributeRef
	i1 := LLVMInt1TypeInContext(c)
	if typ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(typ, 0, 0, attr)
}

func lbAbiArm64_is_homogenous_array(c LLVMContextRef, typ LLVMTypeRef, base_type_ *LLVMTypeRef, member_count_ *uint32) bool {
	if !lb_is_type_kind(typ, LLVMArrayTypeKind) {
		return false
	}
	length := LLVMGetArrayLength(typ)
	if length == 0 {
		return false
	}
	elem := OdinLLVMGetArrayElementType(typ)
	var base_type LLVMTypeRef
	var member_count uint32
	if lbAbiArm64_is_homogenous_aggregate(c, elem, &base_type, &member_count) {
		if base_type_ != nil {
			*base_type_ = base_type
		}
		if member_count_ != nil {
			*member_count_ = member_count * uint32(length)
		}
		return true
	}
	return false
}

func lbAbiArm64_is_homogenous_struct(c LLVMContextRef, typ LLVMTypeRef, base_type_ *LLVMTypeRef, member_count_ *uint32) bool {
	if !lb_is_type_kind(typ, LLVMStructTypeKind) {
		return false
	}
	elem_count := LLVMCountStructElementTypes(typ)
	if elem_count == 0 {
		return false
	}
	var base_type LLVMTypeRef
	var member_count uint32
	for i := uint(0); i < elem_count; i++ {
		var field_type LLVMTypeRef
		var field_member_count uint32
		elem := LLVMStructGetTypeAtIndex(typ, i)
		if !lbAbiArm64_is_homogenous_aggregate(c, elem, &field_type, &field_member_count) {
			return false
		}
		if base_type == 0 {
			base_type = field_type
			member_count = field_member_count
		} else {
			if base_type != field_type {
				return false
			}
			member_count += field_member_count
		}
	}
	if base_type == 0 {
		return false
	}
	if lb_sizeof(typ) == lb_sizeof(base_type)*int64(member_count) {
		if base_type_ != nil {
			*base_type_ = base_type
		}
		if member_count_ != nil {
			*member_count_ = member_count
		}
		return true
	}
	return false
}

func lbAbiArm64_is_homogenous_aggregate(c LLVMContextRef, typ LLVMTypeRef, base_type_ *LLVMTypeRef, member_count_ *uint32) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMFloatTypeKind, LLVMDoubleTypeKind:
		if base_type_ != nil {
			*base_type_ = typ
		}
		if member_count_ != nil {
			*member_count_ = 1
		}
		return true
	case LLVMArrayTypeKind:
		return lbAbiArm64_is_homogenous_array(c, typ, base_type_, member_count_)
	case LLVMStructTypeKind:
		return lbAbiArm64_is_homogenous_struct(c, typ, base_type_, member_count_)
	}
	return false
}

func lbAbiArm64_is_homogenous_aggregate_small_enough(base_type LLVMTypeRef, member_count uint32) bool {
	return member_count <= 4
}

func lbAbiArm64_compute_return_type(ft *lbFunctionType, c LLVMContextRef, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool) lbArgType {
	var homo_base_type LLVMTypeRef
	var homo_member_count uint32
	if !return_is_defined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c))
	} else if lbAbiArm64_is_register(return_type) {
		return lbAbiArm64_non_struct(c, return_type)
	} else if lbAbiArm64_is_homogenous_aggregate(c, return_type, &homo_base_type, &homo_member_count) {
		if lbAbiArm64_is_homogenous_aggregate_small_enough(homo_base_type, homo_member_count) {
			return lb_arg_type_direct(return_type, llvm_array_type(homo_base_type, uint64(homo_member_count)), 0, 0)
		} else {
			if return_is_tuple {
				if newRet := lbAbiArm64_modifyReturnIsTuple(ft, c, return_type); newRet.Type != 0 {
					return newRet
				}
			}
			attr := lb_create_enum_attribute_with_type(c, "sret", return_type)
			return lb_arg_type_indirect(return_type, attr)
		}
	} else {
		size := lb_sizeof(return_type)
		if size > 16 {
			if return_is_tuple {
				if newRet := lbAbiArm64_modifyReturnIsTuple(ft, c, return_type); newRet.Type != 0 {
					return newRet
				}
			}
			attr := lb_create_enum_attribute_with_type(c, "sret", return_type)
			return lb_arg_type_indirect(return_type, attr)
		}
		var cast_type LLVMTypeRef
		if size == 0 {
			cast_type = LLVMStructTypeInContext(c, nil, 0, false)
		} else if size <= 8 {
			cast_type = LLVMIntTypeInContext(c, uint(size*8))
		} else {
			llvm_i64 := LLVMIntTypeInContext(c, 64)
			cast_type = llvm_array_type(llvm_i64, 2)
		}
		return lb_arg_type_direct(return_type, cast_type, 0, 0)
	}
}

func lbAbiArm64_compute_arg_types(c LLVMContextRef, arg_types []LLVMTypeRef, arg_count uint32) []lbArgType {
	args := make([]lbArgType, arg_count)
	for i := uint32(0); i < arg_count; i++ {
		typ := arg_types[i]
		var homo_base_type LLVMTypeRef
		var homo_member_count uint32
		if lbAbiArm64_is_register(typ) {
			args[i] = lbAbiArm64_non_struct(c, typ)
		} else if lbAbiArm64_is_homogenous_aggregate(c, typ, &homo_base_type, &homo_member_count) {
			if lbAbiArm64_is_homogenous_aggregate_small_enough(homo_base_type, homo_member_count) {
				args[i] = lb_arg_type_direct(typ, llvm_array_type(homo_base_type, uint64(homo_member_count)), 0, 0)
			} else {
				args[i] = lb_arg_type_indirect(typ, 0)
			}
		} else {
			size := lb_sizeof(typ)
			if size <= 16 {
				var cast_type LLVMTypeRef
				if size == 0 {
					cast_type = LLVMStructTypeInContext(c, nil, 0, false)
				} else if size <= 8 {
					cast_type = LLVMIntTypeInContext(c, uint(size*8))
				} else {
					count := uint((size + 7) / 8)
					llvm_i64 := LLVMIntTypeInContext(c, 64)
					types := make([]LLVMTypeRef, count)
					size_copy := size
					for j := uint(0); j < count; j++ {
						if size_copy >= 8 {
							types[j] = llvm_i64
						} else {
							types[j] = LLVMIntTypeInContext(c, uint(8*size_copy))
						}
						size_copy -= 8
					}
					cast_type = LLVMStructTypeInContext(c, types, uint(len(types)), true)
				}
				args[i] = lb_arg_type_direct(typ, cast_type, 0, 0)
			} else {
				args[i] = lb_arg_type_indirect(typ, 0)
			}
		}
	}
	return args
}

func lbAbiArm64_modifyReturnIsTuple(ft *lbFunctionType, c LLVMContextRef, return_type LLVMTypeRef) lbArgType {
	var return_arg lbArgType
	if lb_is_type_kind(return_type, LLVMStructTypeKind) {
		field_count := LLVMCountStructElementTypes(return_type)
		if field_count > 1 {
			ft.OriginalArgCount = int64(len(ft.Args))
			ft.MultipleReturnOriginalType = return_type
			for i := uint(0); i < field_count-1; i++ {
				field_type := LLVMStructGetTypeAtIndex(return_type, i)
				field_pointer_type := LLVMPointerType(field_type, 0)
				ret_partial := lb_arg_type_direct(field_pointer_type)
				ft.Args = append(ft.Args, ret_partial)
			}
			new_return_type := LLVMStructGetTypeAtIndex(return_type, field_count-1)
			return_arg = lbAbiArm64_compute_return_type(ft, c, new_return_type, true, false)
		}
	}
	return return_arg
}

// ---- lbAbiWasm ----

const lbAbiWasm_MAX_DIRECT_STRUCT_SIZE = 32

func lbAbiWasm_abi_info(m *lbModule, arg_types []LLVMTypeRef, arg_count uint32, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	c := m.Ctx
	ft := new(lbFunctionType)
	ft.Ctx = c
	ft.CallingConvention = calling_convention
	ft.Args = lbAbiWasm_compute_arg_types(c, arg_types, arg_count, calling_convention, original_type)
	ft.Ret = lbAbiWasm_compute_return_type(ft, c, return_type, return_is_defined, return_is_tuple, original_type.Proc.Results)
	return ft
}

func lbAbiWasm_non_struct(c LLVMContextRef, typ LLVMTypeRef, is_return bool) lbArgType {
	if typ == LLVMIntTypeInContext(c, 128) {
		return lb_arg_type_direct(typ, 0, 0, 0)
	}
	if !is_return && lb_sizeof(typ) > 8 {
		return lb_arg_type_indirect(typ, 0)
	}
	var attr LLVMAttributeRef
	i1 := LLVMInt1TypeInContext(c)
	if typ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(typ, 0, 0, attr)
}

func lbAbiWasm_is_basic_register_type(typ LLVMTypeRef) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind, LLVMPointerTypeKind:
		return true
	case LLVMIntegerTypeKind:
		return lb_sizeof(typ) <= 16
	}
	return false
}

func lbAbiWasm_type_can_be_direct(typ LLVMTypeRef, original_type *Type, calling_convention ProcCallingConvention) bool {
	kind := LLVMGetTypeKind(typ)
	sz := lb_sizeof(typ)
	if sz == 0 {
		return false
	}
	if calling_convention == ProcCC_CDecl {
		if kind == LLVMArrayTypeKind {
			return false
		} else if kind == LLVMStructTypeKind {
			count := LLVMCountStructElementTypes(typ)
			bt := base_type(original_type)
			if bt.Kind == Type_Struct && bt.Struct.IsRawUnion {
				count = uint(len(bt.Struct.Fields))
			}
			if count == 1 {
				return lbAbiWasm_type_can_be_direct(
					LLVMStructGetTypeAtIndex(typ, 0),
					type_internal_index(original_type, 0),
					calling_convention,
				)
			}
		} else if lbAbiWasm_is_basic_register_type(typ) {
			return true
		}
	} else if sz <= lbAbiWasm_MAX_DIRECT_STRUCT_SIZE {
		if kind == LLVMArrayTypeKind {
			if lbAbiWasm_is_basic_register_type(OdinLLVMGetArrayElementType(typ)) {
				return true
			}
		} else if kind == LLVMStructTypeKind {
			count := LLVMCountStructElementTypes(typ)
			for i := uint(0); i < count; i++ {
				elem := LLVMStructGetTypeAtIndex(typ, i)
				if !lbAbiWasm_is_basic_register_type(elem) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func lbAbiWasm_is_struct(c LLVMContextRef, typ LLVMTypeRef, original_type *Type, calling_convention ProcCallingConvention) lbArgType {
	kind := LLVMGetTypeKind(typ)
	sz := lb_sizeof(typ)
	if sz == 0 {
		return lb_arg_type_ignore(typ)
	}
	if lbAbiWasm_type_can_be_direct(typ, original_type, calling_convention) {
		return lb_arg_type_direct(typ)
	}
	return lb_arg_type_indirect(typ, 0)
}

func lbAbiWasm_pseudo_slice(c LLVMContextRef, typ LLVMTypeRef, original_type *Type, calling_convention ProcCallingConvention) lbArgType {
	if build_context.metrics.PtrSize < build_context.metrics.IntSize &&
		lbAbiWasm_type_can_be_direct(typ, original_type, calling_convention) {
		types := []LLVMTypeRef{
			LLVMStructGetTypeAtIndex(typ, 0),
			LLVMStructGetTypeAtIndex(typ, 2),
		}
		new_type := LLVMStructTypeInContext(c, types, uint(len(types)), false)
		return lb_arg_type_direct(typ, new_type, 0, 0)
	} else {
		return lbAbiWasm_is_struct(c, typ, original_type, calling_convention)
	}
}

func lbAbiWasm_compute_arg_types(c LLVMContextRef, arg_types []LLVMTypeRef, arg_count uint32, calling_convention ProcCallingConvention, original_type *Type) []lbArgType {
	args := make([]lbArgType, arg_count)
	params := original_type.Proc.Params.Tuple.Variables
	for i, j := uint32(0), 0; i < arg_count; i, j = i+1, j+1 {
		for params[j].Kind != Entity_Variable {
			j++
		}
		ptype := params[j].Type
		t := arg_types[i]
		kind := LLVMGetTypeKind(t)
		if kind == LLVMStructTypeKind || kind == LLVMArrayTypeKind {
			if is_type_slice(ptype) || is_type_string(ptype) {
				args[i] = lbAbiWasm_pseudo_slice(c, t, ptype, calling_convention)
			} else {
				args[i] = lbAbiWasm_is_struct(c, t, ptype, calling_convention)
			}
		} else {
			args[i] = lbAbiWasm_non_struct(c, t, false)
		}
	}
	return args
}

func lbAbiWasm_compute_return_type(ft *lbFunctionType, c LLVMContextRef, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, original_type *Type) lbArgType {
	if !return_is_defined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c))
	} else if lb_is_type_kind(return_type, LLVMStructTypeKind) || lb_is_type_kind(return_type, LLVMArrayTypeKind) {
		if lbAbiWasm_type_can_be_direct(return_type, original_type, ft.CallingConvention) {
			return lb_arg_type_direct(return_type)
		} else if ft.CallingConvention != ProcCC_CDecl {
			sz := lb_sizeof(return_type)
			switch sz {
			case 1:
				return lb_arg_type_direct(return_type, LLVMIntTypeInContext(c, 8), 0, 0)
			case 2:
				return lb_arg_type_direct(return_type, LLVMIntTypeInContext(c, 16), 0, 0)
			case 4:
				return lb_arg_type_direct(return_type, LLVMIntTypeInContext(c, 32), 0, 0)
			case 8:
				return lb_arg_type_direct(return_type, LLVMIntTypeInContext(c, 64), 0, 0)
			}
		}
		if return_is_tuple {
			var return_arg lbArgType
			if lb_is_type_kind(return_type, LLVMStructTypeKind) {
				field_count := LLVMCountStructElementTypes(return_type)
				if field_count > 1 {
					ft.OriginalArgCount = int64(len(ft.Args))
					ft.MultipleReturnOriginalType = return_type
					for i := uint(0); i < field_count-1; i++ {
						field_type := LLVMStructGetTypeAtIndex(return_type, i)
						field_pointer_type := LLVMPointerType(field_type, 0)
						ret_partial := lb_arg_type_direct(field_pointer_type)
						ft.Args = append(ft.Args, ret_partial)
					}
					return_arg = lbAbiWasm_compute_return_type(
						ft, c,
						LLVMStructGetTypeAtIndex(return_type, field_count-1),
						true, false,
						type_internal_index(original_type, int64(field_count-1)),
					)
				}
			}
			if return_arg.Type != 0 {
				return return_arg
			}
		}
		attr := lb_create_enum_attribute_with_type(c, "sret", return_type)
		return lb_arg_type_indirect(return_type, attr)
	}
	return lbAbiWasm_non_struct(c, return_type, true)
}

// ---- lbAbiArm32 ----

func lbAbiArm32_abi_info(m *lbModule, arg_types []LLVMTypeRef, arg_count uint32, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	c := m.Ctx
	ft := new(lbFunctionType)
	ft.Ctx = c
	ft.Args = lbAbiArm32_compute_arg_types(c, arg_types, arg_count, calling_convention)
	ft.Ret = lbAbiArm32_compute_return_type(c, return_type, return_is_defined)
	ft.CallingConvention = calling_convention
	return ft
}

func lbAbiArm32_is_register(typ LLVMTypeRef, is_return bool) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind:
		return true
	case LLVMIntegerTypeKind:
		return lb_sizeof(typ) <= 8
	case LLVMFunctionTypeKind, LLVMPointerTypeKind, LLVMVectorTypeKind:
		return true
	}
	return false
}

func lbAbiArm32_non_struct(c LLVMContextRef, typ LLVMTypeRef, is_return bool) lbArgType {
	var attr LLVMAttributeRef
	i1 := LLVMInt1TypeInContext(c)
	if typ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(typ, 0, 0, attr)
}

func lbAbiArm32_compute_arg_types(c LLVMContextRef, arg_types []LLVMTypeRef, arg_count uint32, calling_convention ProcCallingConvention) []lbArgType {
	args := make([]lbArgType, arg_count)
	for i := uint32(0); i < arg_count; i++ {
		t := arg_types[i]
		if lbAbiArm32_is_register(t, false) {
			args[i] = lbAbiArm32_non_struct(c, t, false)
		} else {
			sz := lb_sizeof(t)
			a := lb_alignof(t)
			if is_calling_convention_odin(calling_convention) && sz > 8 {
				args[i] = lb_arg_type_indirect(t, 0)
			} else if a <= 4 {
				n := uint64((sz + 3) / 4)
				args[i] = lb_arg_type_direct(llvm_array_type(LLVMIntTypeInContext(c, 32), n))
			} else {
				n := uint64((sz + 7) / 8)
				args[i] = lb_arg_type_direct(llvm_array_type(LLVMIntTypeInContext(c, 64), n))
			}
		}
	}
	return args
}

func lbAbiArm32_compute_return_type(c LLVMContextRef, return_type LLVMTypeRef, return_is_defined bool) lbArgType {
	if !return_is_defined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c))
	} else if !lbAbiArm32_is_register(return_type, true) {
		switch lb_sizeof(return_type) {
		case 1:
			return lb_arg_type_direct(LLVMIntTypeInContext(c, 8), return_type, 0, 0)
		case 2:
			return lb_arg_type_direct(LLVMIntTypeInContext(c, 16), return_type, 0, 0)
		case 3, 4:
			return lb_arg_type_direct(LLVMIntTypeInContext(c, 32), return_type, 0, 0)
		}
		attr := lb_create_enum_attribute_with_type(c, "sret", return_type)
		return lb_arg_type_indirect(return_type, attr)
	}
	return lbAbiArm32_non_struct(c, return_type, true)
}

// ---- lbAbiRiscv64 ----

func lbAbiRiscv64_abi_info(m *lbModule, arg_types []LLVMTypeRef, arg_count uint32, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	ft := new(lbFunctionType)
	ft.Ctx = m.Ctx
	ft.CallingConvention = calling_convention
	gprs := 8
	fprs := 8
	ft.Args = lbAbiRiscv64_compute_arg_types(m, arg_types, arg_count, calling_convention, original_type, &gprs, &fprs)
	ft.Ret = lbAbiRiscv64_compute_return_type(ft, m, return_type, return_is_defined, return_is_tuple, original_type, &gprs)
	return ft
}

func lbAbiRiscv64_is_register(typ LLVMTypeRef) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMIntegerTypeKind, LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind, LLVMPointerTypeKind:
		return true
	}
	return false
}

func lbAbiRiscv64_is_float(typ LLVMTypeRef) bool {
	switch LLVMGetTypeKind(typ) {
	case LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind:
		return true
	default:
		return false
	}
}

func lbAbiRiscv64_non_struct(c LLVMContextRef, typ LLVMTypeRef) lbArgType {
	var attr LLVMAttributeRef
	i1 := LLVMInt1TypeInContext(c)
	if typ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(typ, 0, 0, attr)
}

func lbAbiRiscv64_flatten(m *lbModule, fields *[]LLVMTypeRef, typ LLVMTypeRef, with_padding bool) {
	kind := LLVMGetTypeKind(typ)
	switch kind {
	case LLVMStructTypeKind:
		if LLVMIsPackedStruct(typ) != 0 {
			*fields = append(*fields, typ)
			break
		}
		if !with_padding {
			if fieldRemapping := map_get(&m.StructFieldRemapping, uintptr(typ)); fieldRemapping != nil {
				remap := *fieldRemapping
				for i := 0; i < len(remap); i++ {
					lbAbiRiscv64_flatten(m, fields, LLVMStructGetTypeAtIndex(typ, uint(remap[i])), with_padding)
				}
				break
			} else {
				debugf("no field mapping for type: %s\n", LLVMPrintTypeToString(typ))
			}
		}
		elem_count := LLVMCountStructElementTypes(typ)
		for i := uint(0); i < elem_count; i++ {
			lbAbiRiscv64_flatten(m, fields, LLVMStructGetTypeAtIndex(typ, i), with_padding)
		}
	case LLVMArrayTypeKind:
		length := LLVMGetArrayLength(typ)
		elem := OdinLLVMGetArrayElementType(typ)
		for i := uint64(0); i < length; i++ {
			lbAbiRiscv64_flatten(m, fields, elem, with_padding)
		}
	default:
		*fields = append(*fields, typ)
	}
}

func lbAbiRiscv64_compute_arg_type(m *lbModule, typ LLVMTypeRef, gprs_left *int, fprs_left *int, odin_type *Type) lbArgType {
	c := m.Ctx
	xlen := 8
	flen := 8
	kind := LLVMGetTypeKind(typ)
	size := lb_sizeof(typ)
	if size == 0 {
		return lb_arg_type_direct(typ, LLVMStructTypeInContext(c, nil, 0, false), 0, 0)
	}
	orig_type := typ
	if kind == LLVMStructTypeKind && size <= int64(max(2*xlen, 2*flen)) {
		fields := make([]LLVMTypeRef, 0, LLVMCountStructElementTypes(typ))
		lbAbiRiscv64_flatten(m, &fields, typ, false)
		if len(fields) == 1 {
			typ = fields[0]
		} else {
			typ = LLVMStructTypeInContext(c, fields, uint(len(fields)), false)
		}
		kind = LLVMGetTypeKind(typ)
		size = lb_sizeof(typ)
	}
	if lbAbiRiscv64_is_float(typ) && size <= int64(flen) && *fprs_left >= 1 {
		*fprs_left -= 1
		return lbAbiRiscv64_non_struct(c, orig_type)
	}
	if kind == LLVMStructTypeKind && size <= int64(2*flen) {
		elem_count := LLVMCountStructElementTypes(typ)
		if elem_count == 2 {
			ty1 := LLVMStructGetTypeAtIndex(typ, 0)
			ty1s := lb_sizeof(ty1)
			ty2 := LLVMStructGetTypeAtIndex(typ, 1)
			ty2s := lb_sizeof(ty2)
			if lbAbiRiscv64_is_float(ty1) && lbAbiRiscv64_is_float(ty2) && ty1s <= int64(flen) && ty2s <= int64(flen) && *fprs_left >= 2 {
				*fprs_left -= 2
				return lb_arg_type_direct(orig_type, typ, 0, 0)
			}
			if lbAbiRiscv64_is_float(ty1) && lbAbiRiscv64_is_register(ty2) && ty1s <= int64(flen) && ty2s <= int64(xlen) && *fprs_left >= 1 && *gprs_left >= 1 {
				*fprs_left -= 1
				*gprs_left -= 1
				return lb_arg_type_direct(orig_type, typ, 0, 0)
			}
			if lbAbiRiscv64_is_register(ty1) && lbAbiRiscv64_is_float(ty2) && ty1s <= int64(xlen) && ty2s <= int64(flen) && *gprs_left >= 1 && *fprs_left >= 1 {
				*fprs_left -= 1
				*gprs_left -= 1
				return lb_arg_type_direct(orig_type, typ, 0, 0)
			}
		}
	}
	if size <= int64(xlen) {
		*gprs_left -= 1
		if lbAbiRiscv64_is_register(typ) {
			return lbAbiRiscv64_non_struct(c, orig_type)
		} else {
			return lb_arg_type_direct(orig_type, LLVMIntTypeInContext(c, uint(size*8)), 0, 0)
		}
	} else if size <= int64(2*xlen) {
		fields := []LLVMTypeRef{
			LLVMIntTypeInContext(c, uint(xlen*8)),
			LLVMIntTypeInContext(c, uint((size-int64(xlen))*8)),
		}
		*gprs_left -= 2
		return lb_arg_type_direct(orig_type, LLVMStructTypeInContext(c, fields, 2, false), 0, 0)
	} else {
		return lb_arg_type_indirect(orig_type, 0)
	}
}

func lbAbiRiscv64_compute_arg_types(m *lbModule, arg_types []LLVMTypeRef, arg_count uint32, calling_convention ProcCallingConvention, odin_type *Type, gprs *int, fprs *int) []lbArgType {
	args := make([]lbArgType, arg_count)
	for i := uint32(0); i < arg_count; i++ {
		typ := arg_types[i]
		args[i] = lbAbiRiscv64_compute_arg_type(m, typ, gprs, fprs, odin_type)
	}
	return args
}

func lbAbiRiscv64_compute_return_type(ft *lbFunctionType, m *lbModule, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, odin_type *Type, agprs *int) lbArgType {
	c := m.Ctx
	if !return_is_defined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c))
	}
	gprs := 2
	fprs := 2
	ret := lbAbiRiscv64_compute_arg_type(m, return_type, &gprs, &fprs, odin_type)
	if ret.Kind == lbArg_Indirect {
		if return_is_tuple {
			if lb_is_type_kind(return_type, LLVMStructTypeKind) {
				field_count := int(LLVMCountStructElementTypes(return_type))
				if field_count > 1 && field_count <= *agprs {
					ft.OriginalArgCount = int64(len(ft.Args))
					ft.MultipleReturnOriginalType = return_type
					for i := 0; i < field_count-1; i++ {
						field_type := LLVMStructGetTypeAtIndex(return_type, uint(i))
						field_pointer_type := LLVMPointerType(field_type, 0)
						ret_partial := lb_arg_type_direct(field_pointer_type)
						ft.Args = append(ft.Args, ret_partial)
						*agprs -= 1
					}
					new_return_type := LLVMStructGetTypeAtIndex(return_type, uint(field_count-1))
					return lbAbiRiscv64_compute_return_type(ft, m, new_return_type, true, false, odin_type, agprs)
				}
			}
		}
		attr := lb_create_enum_attribute_with_type(c, "sret", ret.Type)
		return lb_arg_type_indirect(ret.Type, attr)
	}
	return ret
}
