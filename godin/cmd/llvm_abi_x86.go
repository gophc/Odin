package cmd

func lbAbi386_abi_info(m *lbModule, argTypes []LLVMTypeRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool, callingConvention ProcCallingConvention, originalType *Type) *lbFunctionType {
	c := m.Ctx
	ft := permanent_alloc_item[lbFunctionType]()
	ft.Ctx = c
	ft.Args = lbAbi386_compute_arg_types(c, argTypes)
	ft.Ret = lbAbi386_compute_return_type(ft, c, returnType, returnIsDefined, returnIsTuple)
	ft.CallingConvention = callingConvention
	return ft
}

func lbAbi386_non_struct(c LLVMContextRef, type_ LLVMTypeRef, isReturn bool) lbArgType {
	if !isReturn && lb_sizeof(type_) > 8 {
		return lb_arg_type_indirect(type_, 0)
	}
	if buildContext.Metrics.Os == TargetOsWindows &&
		buildContext.PtrSize == 8 &&
		lb_is_type_kind(type_, LLVMIntegerTypeKind) &&
		type_ == LLVMIntTypeInContext(c, 128) {
		castType := LLVMVectorType(LLVMInt64TypeInContext(c), 2)
		return lb_arg_type_direct(type_, castType, nil, 0)
	}
	var attr LLVMAttributeRef
	i1 := LLVMInt1TypeInContext(c)
	if type_ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(type_, 0, nil, attr)
}

func lbAbi386_compute_arg_types(c LLVMContextRef, argTypes []LLVMTypeRef) []lbArgType {
	args := make([]lbArgType, len(argTypes))
	for i, t := range argTypes {
		kind := LLVMGetTypeKind(t)
		sz := lb_sizeof(t)
		if kind == LLVMStructTypeKind || kind == LLVMArrayTypeKind {
			if sz == 0 {
				args[i] = lb_arg_type_ignore(t)
			} else {
				args[i] = lb_arg_type_indirect(t, 0)
			}
		} else {
			args[i] = lbAbi386_non_struct(c, t, false)
		}
	}
	return args
}

func lbAbi386_compute_return_type(ft *lbFunctionType, c LLVMContextRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool) lbArgType {
	if !returnIsDefined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c), 0, nil, 0)
	} else if lb_is_type_kind(returnType, LLVMStructTypeKind) || lb_is_type_kind(returnType, LLVMArrayTypeKind) {
		sz := lb_sizeof(returnType)
		switch sz {
		case 1:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 8), nil, 0)
		case 2:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 16), nil, 0)
		case 4:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 32), nil, 0)
		case 8:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 64), nil, 0)
		}
		if returnIsTuple {
			newReturnType := lb_abi_modify_return_is_tuple(ft, c, returnType, lbAbi386_compute_return_type)
			if newReturnType.Type != 0 {
				return newReturnType
			}
		}
		attr := lb_create_enum_attribute_with_type(c, "sret", returnType)
		return lb_arg_type_indirect(returnType, attr)
	}
	return lbAbi386_non_struct(c, returnType, true)
}

func lbAbiAmd64Win64_abi_info(m *lbModule, argTypes []LLVMTypeRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool, callingConvention ProcCallingConvention, originalType *Type) *lbFunctionType {
	c := m.Ctx
	ft := permanent_alloc_item[lbFunctionType]()
	ft.Ctx = c
	ft.Args = lbAbiAmd64Win64_compute_arg_types(c, argTypes)
	ft.Ret = lbAbiAmd64Win64_compute_return_type(ft, c, returnType, returnIsDefined, returnIsTuple)
	ft.CallingConvention = callingConvention
	return ft
}

func lbAbiAmd64Win64_compute_arg_types(c LLVMContextRef, argTypes []LLVMTypeRef) []lbArgType {
	args := make([]lbArgType, len(argTypes))
	for i, t := range argTypes {
		kind := LLVMGetTypeKind(t)
		if kind == LLVMStructTypeKind || kind == LLVMArrayTypeKind {
			sz := lb_sizeof(t)
			switch sz {
			case 1, 2, 4, 8:
				args[i] = lb_arg_type_direct(t, LLVMIntTypeInContext(c, 8*uint(sz)), nil, 0)
			default:
				args[i] = lb_arg_type_indirect(t, 0)
			}
		} else {
			args[i] = lbAbi386_non_struct(c, t, false)
		}
	}
	return args
}

func lbAbiAmd64Win64_compute_return_type(ft *lbFunctionType, c LLVMContextRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool) lbArgType {
	if !returnIsDefined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c), 0, nil, 0)
	} else if lb_is_type_kind(returnType, LLVMStructTypeKind) || lb_is_type_kind(returnType, LLVMArrayTypeKind) {
		sz := lb_sizeof(returnType)
		switch sz {
		case 1:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 8), nil, 0)
		case 2:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 16), nil, 0)
		case 4:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 32), nil, 0)
		case 8:
			return lb_arg_type_direct(returnType, LLVMIntTypeInContext(c, 64), nil, 0)
		}
		if returnIsTuple {
			newReturnType := lb_abi_modify_return_is_tuple(ft, c, returnType, lbAbiAmd64Win64_compute_return_type)
			if newReturnType.Type != 0 {
				return newReturnType
			}
		}
		attr := lb_create_enum_attribute_with_type(c, "sret", returnType)
		return lb_arg_type_indirect(returnType, attr)
	}
	return lbAbi386_non_struct(c, returnType, true)
}

func is_llvm_type_slice_like(type_ LLVMTypeRef) bool {
	if !lb_is_type_kind(type_, LLVMStructTypeKind) {
		return false
	}
	if LLVMCountStructElementTypes(type_) != 2 {
		return false
	}
	var fields [2]LLVMTypeRef
	LLVMGetStructElementTypes(type_, fields[:])
	if !lb_is_type_kind(fields[0], LLVMPointerTypeKind) {
		return false
	}
	return lb_is_type_kind(fields[1], LLVMIntegerTypeKind) && lb_sizeof(fields[1]) == 8
}
