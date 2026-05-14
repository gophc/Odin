package cmd

func lb_call_intrinsic(p *lbProcedure, name string, args []LLVMValueRef, arg_count uint, types []LLVMTypeRef, type_count uint) LLVMValueRef {
	id := LLVMLookupIntrinsicID(name, uint(len(name)))
	gb_assert_handler("Assertion Failure", id != 0, "llvm_backend_proc_mem.go", 0, "Unable to find %s", name)
	ip := LLVMGetIntrinsicDeclaration(p.Module.Mod, id, types, type_count)
	call_type := LLVMIntrinsicGetType(p.Module.Ctx, id, types, type_count)
	return LLVMBuildCall2(p.Builder, call_type, ip, args, arg_count, "")
}

func lb_mem_copy_overlapping(p *lbProcedure, dst lbValue, src lbValue, len_ lbValue, is_volatile ...bool) {
	is_vol := false
	if len(is_volatile) > 0 {
		is_vol = is_volatile[0]
	}
	dst = lb_emit_conv(p, dst, t_rawptr)
	src = lb_emit_conv(p, src, t_rawptr)
	len_ = lb_emit_conv(p, len_, t_int)

	name := "llvm.memmove"
	if !p.IsStartup && LLVMIsConstant(len_.Value) != 0 {
		const_len := LLVMConstIntGetSExtValue(len_.Value)
		if const_len <= lb_max_zero_init_size() {
			name = "llvm.memmove.inline"
		}
	}

	types := [3]LLVMTypeRef{
		lb_type(p.Module, t_rawptr),
		lb_type(p.Module, t_rawptr),
		lb_type(p.Module, t_int),
	}
	args := [4]LLVMValueRef{
		dst.Value,
		src.Value,
		len_.Value,
		LLVMConstInt(LLVMInt1TypeInContext(p.Module.Ctx), uint64(boolToLLVM(is_vol)), false),
	}
	lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
}

func lb_mem_copy_non_overlapping(p *lbProcedure, dst lbValue, src lbValue, len_ lbValue, is_volatile ...bool) {
	is_vol := false
	if len(is_volatile) > 0 {
		is_vol = is_volatile[0]
	}
	dst = lb_emit_conv(p, dst, t_rawptr)
	src = lb_emit_conv(p, src, t_rawptr)
	len_ = lb_emit_conv(p, len_, t_int)

	name := "llvm.memcpy"
	if !p.IsStartup && LLVMIsConstant(len_.Value) != 0 {
		const_len := LLVMConstIntGetSExtValue(len_.Value)
		if const_len <= lb_max_zero_init_size() {
			name = "llvm.memcpy.inline"
		}
	}

	types := [3]LLVMTypeRef{
		lb_type(p.Module, t_rawptr),
		lb_type(p.Module, t_rawptr),
		lb_type(p.Module, t_int),
	}
	args := [4]LLVMValueRef{
		dst.Value,
		src.Value,
		len_.Value,
		LLVMConstInt(LLVMInt1TypeInContext(p.Module.Ctx), uint64(boolToLLVM(is_vol)), false),
	}
	lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
}
