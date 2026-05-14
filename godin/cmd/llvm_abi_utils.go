package cmd

func lb_function_type_to_llvm_raw(ft *lbFunctionType, is_var_arg bool) LLVMTypeRef {
	arg_count := uint(len(ft.Args))
	offset := uint(0)
	var ret LLVMTypeRef

	if ft.Ret.Kind == lbArg_Direct {
		if ft.Ret.CastType != 0 {
			ret = ft.Ret.CastType
		} else {
			ret = ft.Ret.Type
		}
	} else if ft.Ret.Kind == lbArg_Indirect {
		offset += 1
		ret = LLVMVoidTypeInContext(ft.Ctx)
	} else if ft.Ret.Kind == lbArg_Ignore {
		ret = LLVMVoidTypeInContext(ft.Ctx)
	}

	if ret == 0 {
		gb_assert_handler("Assertion Failure", "ret != nullptr", "llvm_abi_utils.go", 0, "%d", ft.Ret.Kind)
	}

	maximum_arg_count := offset + arg_count
	args := make([]LLVMTypeRef, maximum_arg_count)

	if offset == 1 {
		if ft.Ret.Kind != lbArg_Indirect {
			gb_assert_handler("Assertion Failure", "ft.ret.kind == lbArg_Indirect", "llvm_abi_utils.go", 0)
		}
		args[0] = LLVMPointerType(ft.Ret.Type, 0)
	}

	arg_index := offset
	for i := uint(0); i < arg_count; i++ {
		arg := &ft.Args[i]
		if arg.Kind == lbArg_Direct {
			var arg_type LLVMTypeRef
			if ft.Args[i].CastType != 0 {
				arg_type = arg.CastType
			} else {
				arg_type = arg.Type
			}
			args[arg_index] = arg_type
			arg_index++
		} else if arg.Kind == lbArg_Indirect {
			if ft.MultipleReturnOriginalType == 0 || i < uint(ft.OriginalArgCount) {
				if lb_is_type_kind(arg.Type, LLVMPointerTypeKind) {
					gb_assert_handler("Assertion Failure", "!lb_is_type_kind(arg.type, LLVMPointerTypeKind)", "llvm_abi_utils.go", 0)
				}
			}
			args[arg_index] = LLVMPointerType(arg.Type, 0)
			arg_index++
		} else if arg.Kind == lbArg_Ignore {
		}
	}

	total_arg_count := arg_index
	func_type := LLVMFunctionType(ret, args[:total_arg_count], total_arg_count, is_var_arg)
	return func_type
}

func lb_add_function_type_attributes(fn LLVMValueRef, ft *lbFunctionType, calling_convention ProcCallingConvention) {
	if ft == nil {
		return
	}
	arg_count := uint(len(ft.Args))
	offset := uint(0)
	if ft.Ret.Kind == lbArg_Indirect {
		offset += 1
	}
	c := ft.Ctx
	noalias_attr := lb_create_enum_attribute(c, "noalias")
	nonnull_attr := lb_create_enum_attribute(c, "nonnull")
	nocapture_attr := lb_create_enum_attribute(c, "nocapture")
	arg_index := offset
	for i := uint(0); i < arg_count; i++ {
		arg := &ft.Args[i]
		if arg.Kind == lbArg_Ignore {
			continue
		}
		if arg.Attribute != 0 {
			LLVMAddAttributeAtIndex(fn, arg_index+1, arg.Attribute)
		}
		if arg.AlignAttribute != 0 {
			LLVMAddAttributeAtIndex(fn, arg_index+1, arg.AlignAttribute)
		}
		if arg.NoCapture {
			LLVMAddAttributeAtIndex(fn, arg_index+1, nocapture_attr)
		}
		if ft.MultipleReturnOriginalType != 0 {
			if int(i) >= ft.OriginalArgCount {
				LLVMAddAttributeAtIndex(fn, arg_index+1, noalias_attr)
				LLVMAddAttributeAtIndex(fn, arg_index+1, nonnull_attr)
			}
		}
		arg_index++
	}
	if offset != 0 && ft.Ret.Kind == lbArg_Indirect && ft.Ret.Attribute != 0 {
		LLVMAddAttributeAtIndex(fn, offset, ft.Ret.Attribute)
		LLVMAddAttributeAtIndex(fn, offset, noalias_attr)
	}
	cc_kind := lbCallingConvention_C
	if !is_arch_wasm() {
		cc_kind = lb_calling_convention_map[calling_convention]
	}
	LLVMSetFunctionCallConv(fn, uint(cc_kind))
	if calling_convention == ProcCC_Odin {
		context_index := arg_index
		LLVMAddAttributeAtIndex(fn, context_index, noalias_attr)
		LLVMAddAttributeAtIndex(fn, context_index, nonnull_attr)
		LLVMAddAttributeAtIndex(fn, context_index, nocapture_attr)
	}
}
