package cmd

type RegClass int

const (
	RegClass_NoClass    RegClass = iota
	RegClass_Int
	RegClass_SSEHs
	RegClass_SSEHv
	RegClass_SSEFs
	RegClass_SSEFv
	RegClass_SSEDs
	RegClass_SSEDv
	RegClass_SSEInt8
	RegClass_SSEInt16
	RegClass_SSEInt32
	RegClass_SSEInt64
	RegClass_SSEInt128
	RegClass_SSEUp
	RegClass_X87
	RegClass_X87Up
	RegClass_ComplexX87
	RegClass_Memory
)

type Amd64TypeAttributeKind int

const (
	Amd64TypeAttribute_None       Amd64TypeAttributeKind = iota
	Amd64TypeAttribute_ByVal
	Amd64TypeAttribute_StructRect
)

func lbAbiAmd64SysV_is_sse(regClass RegClass) bool {
	switch regClass {
	case RegClass_SSEHs, RegClass_SSEHv, RegClass_SSEFs, RegClass_SSEFv,
		RegClass_SSEDs, RegClass_SSEDv,
		RegClass_SSEInt8, RegClass_SSEInt16, RegClass_SSEInt32, RegClass_SSEInt64:
		return true
	}
	return false
}

func lbAbiAmd64SysV_all_mem(cls []RegClass) {
	for i := range cls {
		cls[i] = RegClass_Memory
	}
}

func lbAbiAmd64SysV_classify_with(t LLVMTypeRef, cls []RegClass, ix i64, off i64) {
	tAlign := lb_alignof(t)
	tSize := lb_sizeof(t)
	misalign := off % tAlign
	if misalign != 0 {
		e := (off + tSize + 7) / 8
		for i := off / 8; i < e; i++ {
			lbAbiAmd64SysV_unify(cls, ix+i, RegClass_Memory)
		}
		return
	}

	switch LLVMGetTypeKind(t) {
	case LLVMIntegerTypeKind:
		s := tSize
		for s > 0 {
			lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_Int)
			off += 8
			s -= 8
		}
	case LLVMPointerTypeKind:
		lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_Int)
	case LLVMHalfTypeKind:
		if off%8 != 0 {
			lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_SSEHv)
		} else {
			lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_SSEHs)
		}
	case LLVMFloatTypeKind:
		if off%8 == 4 {
			lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_SSEFv)
		} else {
			lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_SSEFs)
		}
	case LLVMDoubleTypeKind:
		lbAbiAmd64SysV_unify(cls, ix+off/8, RegClass_SSEDs)
	case LLVMStructTypeKind:
		packed := LLVMIsPackedStruct(t)
		fieldCount := LLVMCountStructElementTypes(t)
		fieldOff := off
		for fieldIndex := uint(0); fieldIndex < fieldCount; fieldIndex++ {
			fieldType := LLVMStructGetTypeAtIndex(t, fieldIndex)
			if packed == 0 {
				fieldOff = llvm_align_formula(fieldOff, lb_alignof(fieldType))
			}
			lbAbiAmd64SysV_classify_with(fieldType, cls, ix, fieldOff)
			fieldOff += lb_sizeof(fieldType)
		}
	case LLVMArrayTypeKind:
		alen := int64(LLVMGetArrayLength(t))
		elem := OdinLLVMGetArrayElementType(t)
		elemSz := lb_sizeof(elem)
		for i := int64(0); i < alen; i++ {
			lbAbiAmd64SysV_classify_with(elem, cls, ix, off+i*elemSz)
		}
	case LLVMVectorTypeKind:
		vlen := int64(LLVMGetVectorSize(t))
		elem := OdinLLVMGetVectorElementType(t)
		elemSz := lb_sizeof(elem)
		elemKind := LLVMGetTypeKind(elem)
		reg := RegClass_NoClass
		switch elemKind {
		case LLVMIntegerTypeKind:
			elemWidth := LLVMGetIntTypeWidth(elem)
			switch elemWidth {
			case 8:
				reg = RegClass_SSEInt8
			case 16:
				reg = RegClass_SSEInt16
			case 32:
				reg = RegClass_SSEInt32
			case 64:
				reg = RegClass_SSEInt64
			default:
				if elemWidth > 64 {
					for i := int64(0); i < vlen; i++ {
						lbAbiAmd64SysV_classify_with(elem, cls, ix, off+i*elemSz)
					}
				}
			}
		case LLVMHalfTypeKind:
			reg = RegClass_SSEHv
		case LLVMFloatTypeKind:
			reg = RegClass_SSEFv
		case LLVMDoubleTypeKind:
			reg = RegClass_SSEDv
		}
		for i := int64(0); i < vlen; i++ {
			lbAbiAmd64SysV_unify(cls, ix+(off+i*elemSz)/8, reg)
			reg = RegClass_SSEUp
		}
	}
}

func lbAbiAmd64SysV_fixup(t LLVMTypeRef, cls []RegClass) {
	i := i64(0)
	e := i64(len(cls))
	if e > 2 && (lb_is_type_kind(t, LLVMStructTypeKind) ||
		lb_is_type_kind(t, LLVMArrayTypeKind) ||
		lb_is_type_kind(t, LLVMVectorTypeKind)) {
		oldv := cls[i]
		if lbAbiAmd64SysV_is_sse(oldv) {
			for i++; i < e; i++ {
				if oldv != RegClass_SSEUp {
					lbAbiAmd64SysV_all_mem(cls)
					return
				}
			}
		} else {
			lbAbiAmd64SysV_all_mem(cls)
			return
		}
	} else {
		for i < e {
			oldv := cls[i]
			if oldv == RegClass_Memory {
				lbAbiAmd64SysV_all_mem(cls)
				return
			} else if oldv == RegClass_X87Up {
				lbAbiAmd64SysV_all_mem(cls)
				return
			} else if oldv == RegClass_SSEUp {
				cls[i] = RegClass_SSEDv
				i++
			} else if lbAbiAmd64SysV_is_sse(oldv) {
				for i++; i < e; i++ {
					v := cls[i]
					if v != RegClass_SSEUp {
						break
					}
				}
			} else if oldv == RegClass_X87 {
				for i++; i < e; i++ {
					v := cls[i]
					if v != RegClass_X87Up {
						break
					}
				}
			} else {
				i++
			}
		}
	}
}

func lbAbiAmd64SysV_unify(cls []RegClass, i i64, newv RegClass) {
	oldv := cls[i]
	if oldv == newv {
		return
	}
	toWrite := newv
	if oldv == RegClass_NoClass {
		toWrite = newv
	} else if newv == RegClass_NoClass {
		return
	} else if oldv == RegClass_Memory || newv == RegClass_Memory {
		toWrite = RegClass_Memory
	} else if oldv == RegClass_Int || newv == RegClass_Int {
		toWrite = RegClass_Int
	} else if oldv == RegClass_X87 || oldv == RegClass_X87Up || oldv == RegClass_ComplexX87 {
		toWrite = RegClass_Memory
	} else if newv == RegClass_X87 || newv == RegClass_X87Up || newv == RegClass_ComplexX87 {
		toWrite = RegClass_Memory
	} else if newv == RegClass_SSEUp {
		switch oldv {
		case RegClass_SSEHv, RegClass_SSEHs, RegClass_SSEFv, RegClass_SSEFs,
			RegClass_SSEDv, RegClass_SSEDs,
			RegClass_SSEInt8, RegClass_SSEInt16, RegClass_SSEInt32, RegClass_SSEInt64:
			return
		}
	}
	cls[i] = toWrite
}

func lbAbiAmd64SysV_classify(t LLVMTypeRef) []RegClass {
	sz := lb_sizeof(t)
	words := (sz + 7) / 8
	regClasses := make([]RegClass, words)
	if words > 4 {
		lbAbiAmd64SysV_all_mem(regClasses)
	} else {
		lbAbiAmd64SysV_classify_with(t, regClasses, 0, 0)
		lbAbiAmd64SysV_fixup(t, regClasses)
	}
	return regClasses
}

func lbAbiAmd64SysV_llvec_len(regClasses []RegClass, offset isize) uint32 {
	llen := uint32(1)
	for i := offset; i < isize(len(regClasses)); i++ {
		if regClasses[i] != RegClass_SSEUp {
			break
		}
		llen++
	}
	return llen
}

func lbAbiAmd64SysV_llreg(c LLVMContextRef, regClasses []RegClass, type_ LLVMTypeRef) LLVMTypeRef {
	types := make([]LLVMTypeRef, 0, len(regClasses))
	allInts := true
	for _, regClass := range regClasses {
		if regClass != RegClass_Int {
			allInts = false
			break
		}
	}
	sz := lb_sizeof(type_)
	if allInts {
		for i := 0; i < len(regClasses); i++ {
			if sz >= 8 {
				types = append(types, LLVMIntTypeInContext(c, 64))
				sz -= 8
			} else {
				types = append(types, LLVMIntTypeInContext(c, uint(sz*8)))
				sz = 0
			}
		}
	} else {
		for i := 0; i < len(regClasses); {
			regClass := regClasses[i]
			switch regClass {
			case RegClass_Int:
				rs := sz
				if rs > 8 {
					rs = 8
				}
				types = append(types, LLVMIntTypeInContext(c, uint(rs*8)))
				sz -= rs
				i++
			case RegClass_SSEHv, RegClass_SSEFv, RegClass_SSEDv,
				RegClass_SSEInt8, RegClass_SSEInt16, RegClass_SSEInt32, RegClass_SSEInt64:
				var elemsPerWord uint32
				var elemType LLVMTypeRef
				switch regClass {
				case RegClass_SSEHv:
					elemsPerWord = 4
					elemType = LLVMHalfTypeInContext(c)
				case RegClass_SSEFv:
					elemsPerWord = 2
					elemType = LLVMFloatTypeInContext(c)
				case RegClass_SSEDv:
					elemsPerWord = 1
					elemType = LLVMDoubleTypeInContext(c)
				case RegClass_SSEInt8:
					elemsPerWord = 64 / 8
					elemType = LLVMIntTypeInContext(c, 8)
				case RegClass_SSEInt16:
					elemsPerWord = 64 / 16
					elemType = LLVMIntTypeInContext(c, 16)
				case RegClass_SSEInt32:
					elemsPerWord = 64 / 32
					elemType = LLVMIntTypeInContext(c, 32)
				case RegClass_SSEInt64:
					elemsPerWord = 64 / 64
					elemType = LLVMIntTypeInContext(c, 64)
				}
				vecLen := lbAbiAmd64SysV_llvec_len(regClasses, isize(i+1))
				vecType := LLVMVectorType(elemType, uint(vecLen*elemsPerWord))
				types = append(types, vecType)
				sz -= lb_sizeof(vecType)
				i += int(vecLen)
			case RegClass_SSEHs:
				types = append(types, LLVMHalfTypeInContext(c))
				sz -= 2
				i++
			case RegClass_SSEFs:
				types = append(types, LLVMFloatTypeInContext(c))
				sz -= 4
				i++
			case RegClass_SSEDs:
				types = append(types, LLVMDoubleTypeInContext(c))
				sz -= 8
				i++
			}
		}
	}
	if len(types) == 1 {
		return types[0]
	}
	return LLVMStructTypeInContext(c, types, uint(len(types)), 0)
}

func lbAbiAmd64SysV_is_mem_cls(cls []RegClass, attributeKind Amd64TypeAttributeKind) bool {
	if attributeKind == Amd64TypeAttribute_ByVal {
		if len(cls) == 0 {
			return false
		}
		first := cls[0]
		return first == RegClass_Memory || first == RegClass_X87 || first == RegClass_ComplexX87
	} else if attributeKind == Amd64TypeAttribute_StructRect {
		if len(cls) == 0 {
			return false
		}
		return cls[0] == RegClass_Memory
	}
	return false
}

func lbAbiAmd64SysV_is_register(type_ LLVMTypeRef) bool {
	kind := LLVMGetTypeKind(type_)
	sz := lb_sizeof(type_)
	if sz == 0 {
		return false
	}
	switch kind {
	case LLVMIntegerTypeKind:
		if 20 >= 18 && sz >= 16 {
			return true
		}
		return false
	case LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind, LLVMPointerTypeKind:
		return true
	}
	return false
}

func lbAbiAmd64SysV_is_aggregate(type_ LLVMTypeRef) bool {
	kind := LLVMGetTypeKind(type_)
	switch kind {
	case LLVMStructTypeKind:
		if LLVMCountStructElementTypes(type_) == 1 {
			return lbAbiAmd64SysV_is_aggregate(LLVMStructGetTypeAtIndex(type_, 0))
		}
		return true
	case LLVMArrayTypeKind:
		if LLVMGetArrayLength(type_) == 1 {
			return lbAbiAmd64SysV_is_aggregate(LLVMGetElementType(type_))
		}
		return true
	}
	return false
}

func lbAbiAmd64SysV_non_struct(c LLVMContextRef, type_ LLVMTypeRef) lbArgType {
	attr := LLVMAttributeRef(0)
	i1 := LLVMInt1TypeInContext(c)
	if type_ == i1 {
		attr = lb_create_enum_attribute(c, "zeroext")
	}
	return lb_arg_type_direct(type_, nil, nil, attr)
}

func lbAbiAmd64SysV_amd64_type(c LLVMContextRef, type_ LLVMTypeRef, attributeKind Amd64TypeAttributeKind, callingConvention ProcCallingConvention,
	isArg bool,
	intRegs *int32, sseRegs *int32) lbArgType {
	cls := lbAbiAmd64SysV_classify(type_)
	neededInt := int32(0)
	neededSse := int32(0)
	for _, c := range cls {
		switch c {
		case RegClass_Int:
			neededInt++
		case RegClass_SSEHs, RegClass_SSEHv, RegClass_SSEFs, RegClass_SSEFv,
			RegClass_SSEDs, RegClass_SSEDv,
			RegClass_SSEInt8, RegClass_SSEInt16, RegClass_SSEInt32, RegClass_SSEInt64,
			RegClass_SSEInt128, RegClass_SSEUp:
			neededSse++
		}
	}
	ranOutOfRegs := false
	if intRegs != nil && sseRegs != nil {
		*intRegs -= neededInt
		*sseRegs -= neededSse
		intOk := *intRegs >= 0
		sseOk := *sseRegs >= 0
		if *intRegs < 0 {
			*intRegs = 0
		}
		if *sseRegs < 0 {
			*sseRegs = 0
		}
		if (!intOk || !sseOk) && lbAbiAmd64SysV_is_aggregate(type_) {
			ranOutOfRegs = true
		}
	}
	if lbAbiAmd64SysV_is_register(type_) {
		attribute := LLVMAttributeRef(0)
		if type_ == LLVMInt1TypeInContext(c) {
			attribute = lb_create_enum_attribute(c, "zeroext")
		}
		return lb_arg_type_direct(type_, nil, nil, attribute)
	} else if ranOutOfRegs {
		if isArg {
			return lb_arg_type_indirect_byval(c, type_)
		} else {
			attribute := lb_create_enum_attribute_with_type(c, "sret", type_)
			return lb_arg_type_indirect(type_, attribute)
		}
	} else if lbAbiAmd64SysV_is_mem_cls(cls, attributeKind) {
		attribute := LLVMAttributeRef(0)
		if attributeKind == Amd64TypeAttribute_ByVal {
			if is_calling_convention_odin(callingConvention) {
				return lb_arg_type_indirect(type_, attribute)
			}
			return lb_arg_type_indirect_byval(c, type_)
		} else if attributeKind == Amd64TypeAttribute_StructRect {
			attribute = lb_create_enum_attribute_with_type(c, "sret", type_)
		}
		return lb_arg_type_indirect(type_, attribute)
	} else {
		var regType LLVMTypeRef
		if is_llvm_type_slice_like(type_) {
			regType = type_
		} else {
			regType = lbAbiAmd64SysV_llreg(c, cls, type_)
		}
		return lb_arg_type_direct(type_, regType, nil, nil)
	}
}

func lbAbiAmd64SysV_compute_return_type(ft *lbFunctionType, c LLVMContextRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool) lbArgType {
	if !returnIsDefined {
		return lb_arg_type_direct(LLVMVoidTypeInContext(c))
	}
	if returnIsTuple {
		newReturnType := lb_abi_modify_return_is_tuple(ft, c, returnType, lbAbiAmd64SysV_compute_return_type)
		if newReturnType.typ != 0 {
			return newReturnType
		}
	}
	return lbAbiAmd64SysV_amd64_type(c, returnType, Amd64TypeAttribute_StructRect, ft.calling_convention,
		false,
		nil, nil)
}

func lbAbiAmd64SysV_abi_info(m *lbModule, argTypes []LLVMTypeRef, returnType LLVMTypeRef, returnIsDefined bool, returnIsTuple bool, callingConvention ProcCallingConvention, originalType *Type) *lbFunctionType {
	c := m.Ctx
	ft := permanent_alloc_item[lbFunctionType]()
	ft.ctx = c
	ft.calling_convention = callingConvention
	intRegs := int32(6)
	sseRegs := int32(8)
	ft.args = make([]lbArgType, len(argTypes))
	for i := range argTypes {
		ft.args[i] = lbAbiAmd64SysV_amd64_type(c, argTypes[i], Amd64TypeAttribute_ByVal, callingConvention,
			true,
			&intRegs, &sseRegs)
	}
	ft.ret = lbAbiAmd64SysV_compute_return_type(ft, c, returnType, returnIsDefined, returnIsTuple)
	return ft
}
