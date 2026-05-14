package cmd

func lb_abi_modify_return_is_tuple(ft *lbFunctionType, c LLVMContextRef, return_type LLVMTypeRef, compute_return_type lbAbiComputeReturnType) lbArgType {
	if return_type == 0 {
		gb_assert_handler("Assertion Failure", "return_type != nullptr", "llvm_abi_dispatch.go", 0)
	}
	if compute_return_type == nil {
		gb_assert_handler("Assertion Failure", "compute_return_type != nullptr", "llvm_abi_dispatch.go", 0)
	}
	var return_arg lbArgType
	if lb_is_type_kind(return_type, LLVMStructTypeKind) {
		field_count := LLVMCountStructElementTypes(return_type)
		if field_count > 1 {
			ft.OriginalArgCount = isize(len(ft.Args))
			ft.MultipleReturnOriginalType = return_type
			for i := uint(0); i < field_count-1; i++ {
				field_type := LLVMStructGetTypeAtIndex(return_type, i)
				field_pointer_type := LLVMPointerType(field_type, 0)
				ret_partial := lb_arg_type_direct(field_pointer_type)
				ft.Args = append(ft.Args, ret_partial)
			}
			new_return_type := LLVMStructGetTypeAtIndex(return_type, field_count-1)
			return_arg = compute_return_type(ft, c, new_return_type, true, false)
		}
	}
	return return_arg
}

func lb_get_abi_info_internal(m *lbModule, arg_types []LLVMTypeRef, arg_count uint, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	c := m.Ctx
	switch calling_convention {
	case ProcCC_None, ProcCC_InlineAsm:
		ft := &lbFunctionType{}
		ft.Ctx = c
		ft.Args = make([]lbArgType, arg_count)
		for i := uint(0); i < arg_count; i++ {
			ft.Args[i] = lb_arg_type_direct(arg_types[i])
		}
		if return_is_defined {
			ft.Ret = lb_arg_type_direct(return_type)
		} else {
			ft.Ret = lb_arg_type_direct(LLVMVoidTypeInContext(c))
		}
		ft.CallingConvention = calling_convention
		return ft

	case ProcCC_Win64:
		if buildContext.Metrics.Arch != TargetArchAmd64 {
			gb_assert_handler("Assertion Failure", "build_context.metrics.arch == TargetArch_amd64", "llvm_abi_dispatch.go", 0)
		}
		return lbAbiAmd64Win64_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)

	case ProcCC_SysV:
		if buildContext.Metrics.Arch != TargetArchAmd64 {
			gb_assert_handler("Assertion Failure", "build_context.metrics.arch == TargetArch_amd64", "llvm_abi_dispatch.go", 0)
		}
		return lbAbiAmd64SysV_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	}

	switch buildContext.Metrics.Arch {
	case TargetArchAmd64:
		if buildContext.Metrics.Os == TargetOsWindows {
			return lbAbiAmd64Win64_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
		} else if buildContext.Metrics.ABI == TargetABIWin64 {
			return lbAbiAmd64Win64_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
		} else if buildContext.Metrics.ABI == TargetABISysV {
			return lbAbiAmd64SysV_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
		} else {
			return lbAbiAmd64SysV_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
		}
	case TargetArchI386:
		return lbAbi386_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	case TargetArchArm32:
		return lbAbiArm32_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	case TargetArchArm64:
		return lbAbiArm64_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	case TargetArchWasm32:
		return lbAbiWasm_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	case TargetArchWasm64p32:
		return lbAbiWasm_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	case TargetArchRiscv64:
		return lbAbiRiscv64_abi_info(m, arg_types, arg_count, return_type, return_is_defined, return_is_tuple, calling_convention, original_type)
	}

	gb_assert_handler("Panic", 0, "llvm_abi_dispatch.go", 0, "Unsupported ABI")
	return nil
}

func lb_get_abi_info(m *lbModule, arg_types []LLVMTypeRef, arg_count uint, return_type LLVMTypeRef, return_is_defined bool, return_is_tuple bool, calling_convention ProcCallingConvention, original_type *Type) *lbFunctionType {
	ft := lb_get_abi_info_internal(
		m,
		arg_types, arg_count,
		return_type, return_is_defined,
		true && return_is_tuple && is_calling_convention_odin(calling_convention),
		calling_convention,
		base_type(original_type),
	)
	if calling_convention == ProcCC_Odin {
		context_param := lb_arg_type_direct(LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0))
		ft.Args = append(ft.Args, context_param)
	}
	return ft
}
