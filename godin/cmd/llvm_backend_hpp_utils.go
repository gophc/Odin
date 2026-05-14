package cmd

func lb_use_new_pass_system() bool {
	return true
}

func lb_max_zero_init_size() i64 {
	if buildContext.Metrics.Os == TargetOsDarwin && buildContext.Metrics.Arch == TargetArchArm64 {
		return i64(4 * buildContext.Metrics.IntSize)
	}
	return i64(8)
}

func llvm_array_type(ElementType LLVMTypeRef, ElementCount uint64) LLVMTypeRef {
	return LLVMArrayType2(ElementType, ElementCount)
}
