package cmd

type lbArgKind int

const (
	lbArg_Direct  lbArgKind = iota
	lbArg_Indirect
	lbArg_Ignore
)

type lbArgType struct {
	Kind           lbArgKind
	Type           LLVMTypeRef
	CastType       LLVMTypeRef
	PadType        LLVMTypeRef
	Attribute      LLVMAttributeRef
	AlignAttribute LLVMAttributeRef
	ByvalAlignment i64
	IsByval        bool
	NoCapture      bool
}

func lb_arg_type_direct(type_ LLVMTypeRef, extra ...LLVMTypeRef) lbArgType {
	r := lbArgType{Kind: lbArg_Direct, Type: type_}
	if len(extra) > 0 {
		r.CastType = extra[0]
	}
	if len(extra) > 1 {
		r.PadType = extra[1]
	}
	return r
}

func lb_arg_type_indirect(type_ LLVMTypeRef, attr LLVMAttributeRef) lbArgType {
	return lbArgType{Kind: lbArg_Indirect, Type: type_, Attribute: attr}
}

func lb_arg_type_indirect_byval(c LLVMContextRef, type_ LLVMTypeRef) lbArgType {
	alignment := lb_alignof(type_)
	if alignment < 8 {
		alignment = 8
	}
	byvalAttr := lb_create_enum_attribute_with_type(c, "byval", type_)
	alignAttr := lb_create_enum_attribute(c, "align", u64(alignment))
	return lbArgType{Kind: lbArg_Indirect, Type: type_, Attribute: byvalAttr, AlignAttribute: alignAttr, ByvalAlignment: alignment, IsByval: true}
}

func lb_arg_type_ignore(type_ LLVMTypeRef) lbArgType {
	return lbArgType{Kind: lbArg_Ignore, Type: type_}
}

type lbFunctionType struct {
	Ctx                       LLVMContextRef
	CallingConvention         ProcCallingConvention
	Args                      []lbArgType
	Ret                       lbArgType
	MultipleReturnOriginalType LLVMTypeRef
	OriginalArgCount          isize
}

func lb_function_type_args_allocator() gbAllocator {
	return heap_allocator()
}

func llvm_align_formula(off i64, a i64) i64 {
	return (off + a - 1) / a * a
}

func lb_is_type_kind(type_ LLVMTypeRef, kind LLVMTypeKind) bool {
	if type_ == 0 {
		return false
	}
	return LLVMGetTypeKind(type_) == kind
}

func lb_sizeof(type_ LLVMTypeRef) i64 {
	kind := LLVMGetTypeKind(type_)
	switch kind {
	case LLVMVoidTypeKind:
		return 0
	case LLVMIntegerTypeKind:
		w := uint64(LLVMGetIntTypeWidth(type_))
		return i64(w+7) / 8
	case LLVMHalfTypeKind:
		return 2
	case LLVMFloatTypeKind:
		return 4
	case LLVMDoubleTypeKind:
		return 8
	case LLVMPointerTypeKind:
		return buildContext.PtrSize
	case LLVMStructTypeKind:
		fieldCount := LLVMCountStructElementTypes(type_)
		var offset i64
		if LLVMIsPackedStruct(type_) != 0 {
			for i := uint(0); i < fieldCount; i++ {
				offset += lb_sizeof(LLVMStructGetTypeAtIndex(type_, i))
			}
		} else {
			for i := uint(0); i < fieldCount; i++ {
				field := LLVMStructGetTypeAtIndex(type_, i)
				align := lb_alignof(field)
				offset = llvm_align_formula(offset, align)
				offset += lb_sizeof(field)
			}
			offset = llvm_align_formula(offset, lb_alignof(type_))
		}
		return offset
	case LLVMArrayTypeKind:
		elem := OdinLLVMGetArrayElementType(type_)
		elemSize := lb_sizeof(elem)
		count := i64(LLVMGetArrayLength(type_))
		return count * elemSize
	case LLVMVectorTypeKind:
		elem := OdinLLVMGetVectorElementType(type_)
		size := lb_sizeof(elem) * i64(LLVMGetVectorSize(type_))
		return next_pow2(size)
	}
	panic("unhandled type for lb_sizeof")
}

func lb_alignof(type_ LLVMTypeRef) i64 {
	kind := LLVMGetTypeKind(type_)
	switch kind {
	case LLVMVoidTypeKind:
		return 1
	case LLVMIntegerTypeKind:
		w := uint64(LLVMGetIntTypeWidth(type_))
		a := i64(w+7) / 8
		if a < 1 {
			a = 1
		}
		if a > buildContext.MaxAlign {
			a = buildContext.MaxAlign
		}
		return a
	case LLVMHalfTypeKind:
		return 2
	case LLVMFloatTypeKind:
		return 4
	case LLVMDoubleTypeKind:
		return 8
	case LLVMPointerTypeKind:
		return buildContext.PtrSize
	case LLVMStructTypeKind:
		if LLVMIsPackedStruct(type_) != 0 {
			return 1
		}
		fieldCount := LLVMCountStructElementTypes(type_)
		maxAlign := i64(1)
		for i := uint(0); i < fieldCount; i++ {
			fieldAlign := lb_alignof(LLVMStructGetTypeAtIndex(type_, i))
			if fieldAlign > maxAlign {
				maxAlign = fieldAlign
			}
		}
		return maxAlign
	case LLVMArrayTypeKind:
		return lb_alignof(OdinLLVMGetArrayElementType(type_))
	case LLVMVectorTypeKind:
		elem := OdinLLVMGetVectorElementType(type_)
		size := next_pow2(lb_sizeof(elem) * i64(LLVMGetVectorSize(type_)))
		if size < 1 {
			size = 1
		}
		if size > buildContext.MaxSimdAlign {
			size = buildContext.MaxSimdAlign
		}
		return size
	}
	panic("unhandled type for lb_alignof")
}

func next_pow2(v i64) i64 {
	if v <= 0 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	return v + 1
}

type lbAbiInfoType func(m *lbModule, argTypes []LLVMTypeRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool, callingConvention ProcCallingConvention, originalType *Type) *lbFunctionType

type lbAbiComputeReturnType func(ft *lbFunctionType, c LLVMContextRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool) lbArgType
