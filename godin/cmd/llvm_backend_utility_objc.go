package cmd

import (
	"fmt"
	"unsafe"
)

func lb_handle_objc_find_or_register_selector(p *lbProcedure, name String) lbAddr {
	m := p.Module
	found := string_map_get(&m.ObjCSelectors, name)
	if found != nil {
		return *found
	}
	defaultModule := &p.Module.Gen.DefaultModule
	globalName := gb_string_make(permanent_allocator(), "__$objc_SEL::")
	globalName = gb_string_append_length(globalName, name.Text, name.Len)
	t := lb_type(m, t_objc_SEL)
	g := lbValue{
		Value: LLVMAddGlobal(m.Mod, t, unsafe.String(globalName, gb_string_length(globalName))),
		Type:  alloc_type_pointer(t_objc_SEL),
	}
	if defaultModule == m {
		LLVMSetInitializer(g.Value, LLVMConstNull(t))
		lb_add_member(m, make_string_c(unsafe.String(globalName, gb_string_length(globalName))), g)
	} else {
		LLVMSetLinkage(g.Value, LLVMExternalLinkage)
	}
	mpsc_enqueue(&m.Gen.ObjCSelectors, lbObjCGlobal{Module: m, GlobalName: globalName, Name: name, Type: t_objc_SEL})
	addr := lb_addr(g)
	string_map_set(&m.ObjCSelectors, name, addr)
	return addr
}

func lb_handle_objc_find_or_register_class(p *lbProcedure, name String, classImplType *Type) lbAddr {
	m := p.Module
	found := string_map_get(&m.ObjCClasses, name)
	if found != nil {
		return *found
	}
	defaultModule := &p.Module.Gen.DefaultModule
	globalName := gb_string_make(permanent_allocator(), "__$objc_Class::")
	globalName = gb_string_append_length(globalName, name.Text, name.Len)
	t := lb_type(m, t_objc_Class)
	g := lbValue{
		Value: LLVMAddGlobal(m.Mod, t, unsafe.String(globalName, gb_string_length(globalName))),
		Type:  alloc_type_pointer(t_objc_Class),
	}
	if defaultModule == m {
		LLVMSetInitializer(g.Value, LLVMConstNull(t))
		lb_add_member(m, make_string_c(unsafe.String(globalName, gb_string_length(globalName))), g)
	} else {
		LLVMSetLinkage(g.Value, LLVMExternalLinkage)
	}
	mpsc_enqueue(&m.Gen.ObjCClasses, lbObjCGlobal{Module: m, GlobalName: globalName, Name: name, Type: t_objc_Class, ClassImplType: classImplType})
	addr := lb_addr(g)
	string_map_set(&m.ObjCClasses, name, addr)
	return addr
}

func lb_handle_objc_find_or_register_ivar(m *lbModule, selfType *Type) lbAddr {
	name := selfType.Named.TypeName.TypeName.ObjcClassName
	if name == "" {
		gb_assert_handler("Assertion Failure", `name != ""`, "llvm_backend_utility_objc.go", 0)
	}
	found := string_map_get(&m.ObjCIvars, name)
	if found != nil {
		return *found
	}
	defaultModule := &m.Gen.DefaultModule
	globalName := gb_string_make(permanent_allocator(), "__$objc_ivar::")
	globalName = gb_string_append_length(globalName, name.Text, name.Len)
	t := lb_type(m, t_int)
	g := lbValue{
		Value: LLVMAddGlobal(m.Mod, t, unsafe.String(globalName, gb_string_length(globalName))),
		Type:  t_int_ptr,
	}
	if defaultModule == m {
		LLVMSetInitializer(g.Value, LLVMConstInt(t, 0, true))
		lb_add_member(m, make_string_c(unsafe.String(globalName, gb_string_length(globalName))), g)
	} else {
		LLVMSetLinkage(g.Value, LLVMExternalLinkage)
	}
	mpsc_enqueue(&m.Gen.ObjCIvars, lbObjCGlobal{Module: m, GlobalName: globalName, Name: name, Type: t_int, ClassImplType: selfType})
	addr := lb_addr(g)
	string_map_set(&m.ObjCIvars, name, addr)
	return addr
}

func lb_handle_objc_ivar_for_objc_object_pointer(p *lbProcedure, self lbValue) lbValue {
	if !(self.Type.Kind == Type_Pointer && self.Type.Pointer.Elem.Kind == Type_Named) {
		gb_assert_handler("Assertion Failure", "self.type->kind == Type_Pointer && self.type->Pointer.elem->kind == Type_Named", "llvm_backend_utility_objc.go", 0)
	}
	selfType := self.Type.Pointer.Elem
	selfUPtr := lb_emit_conv(p, self, t_uintptr)
	ivarOffset := lb_addr_load(p, lb_handle_objc_find_or_register_ivar(p.Module, selfType))
	ivarOffsetUPtr := lb_emit_conv(p, ivarOffset, t_uintptr)
	ivarUPtr := lb_emit_arith(p, Token_Add, selfUPtr, ivarOffsetUPtr, t_uintptr)
	ivarType := selfType.Named.TypeName.TypeName.ObjcIvar
	return lb_emit_conv(p, ivarUPtr, alloc_type_pointer(ivarType))
}

func lb_handle_objc_ivar_get(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	if ce.Args[0].TAV.Type.Kind != Type_Pointer {
		gb_assert_handler("Assertion Failure", "ce->args[0]->tav.type->kind == Type_Pointer", "llvm_backend_utility_objc.go", 0)
	}
	self := lb_build_expr(p, ce.Args[0])
	return lb_handle_objc_ivar_for_objc_object_pointer(p, self)
}

func lb_create_objc_block_helper_procs(
	m *lbModule, blockLitType LLVMTypeRef, captureFieldOffset isize, blockID isize,
	captureValues []lbValue, objcObjectIndices []isize,
) (outCopyHelper *lbProcedure, outDisposeHelper *lbProcedure) {
	copyHelperName := fmt.Sprintf("__$%s::objc_block_copy_helper_%d", m.ModuleName, blockID)
	disposeHelperName := fmt.Sprintf("__$%s::objc_block_dispose_helper_%d", m.ModuleName, blockID)

	types := [3]*Type{t_rawptr, t_rawptr, t_i32}
	copyTuple := alloc_type_tuple_from_field_types(types[:], 3, false, true)
	disposeTuple := alloc_type_tuple_from_field_types(types[1:], 2, false, true)
	copyProcType := alloc_type_proc(nil, copyTuple, 3, nil, 0, false, ProcCC_CDecl)
	disposeProcType := alloc_type_proc(nil, disposeTuple, 2, nil, 0, false, ProcCC_CDecl)

	copyProc := lb_create_dummy_procedure(m, copyHelperName, copyProcType)
	disposeProc := lb_create_dummy_procedure(m, disposeHelperName, disposeProcType)

	LLVMSetLinkage(copyProc.Value, LLVMPrivateLinkage)
	LLVMSetLinkage(disposeProc.Value, LLVMPrivateLinkage)

	const BLOCK_FIELD_IS_OBJECT = 3
	const BLOCK_FIELD_IS_BLOCK = 7

	blockBaseType := find_core_type(m.Info.Checker, make_string_c("Objc_Block"))

	is_object_objc_block := func(type_ *Type, blockBaseType *Type) bool {
		base := base_type(type_deref(type_))
		if base.Kind != Type_Struct {
			gb_assert_handler("Assertion Failure", "base->kind == Type_Struct", "llvm_backend_utility_objc.go", 0)
		}
		for is_type_polymorphic_record_specialized(base) {
			if base.Struct.PolymorphicParent != nil {
				base = base.Struct.PolymorphicParent
				if base == blockBaseType {
					return true
				}
				base = base_type(base)
				if base.Kind != Type_Struct {
					gb_assert_handler("Assertion Failure", "base->kind == Type_Struct", "llvm_backend_utility_objc.go", 0)
				}
			}
		}
		return false
	}

	lb_begin_procedure_body(copyProc)
	lb_begin_procedure_body(disposeProc)

	for _, objectIndex := range objcObjectIndices {
		fieldOffset := uint(captureFieldOffset + objectIndex)
		fieldType := captureValues[objectIndex].Type
		fieldRawType := lb_type(m, fieldType)
		if !is_type_objc_object(fieldType) {
			gb_assert_handler("Assertion Failure", "is_type_objc_object(field_type)", "llvm_backend_utility_objc.go", 0)
		}
		isBlockObj := is_object_objc_block(fieldType, blockBaseType)

		{
			dstField := LLVMBuildStructGEP2(copyProc.Builder, blockLitType, copyProc.RawInputParameters[0], fieldOffset, "")
			srcField := LLVMBuildStructGEP2(copyProc.Builder, blockLitType, copyProc.RawInputParameters[1], fieldOffset, "")
			dstValue := lbValue{Type: alloc_type_pointer(fieldType), Value: dstField}
			srcValue := lbValue{Type: fieldType, Value: LLVMBuildLoad2(copyProc.Builder, fieldRawType, srcField, "")}
			fieldFlag := BLOCK_FIELD_IS_OBJECT
			if isBlockObj {
				fieldFlag = BLOCK_FIELD_IS_BLOCK
			}
			copyArgs := []lbValue{dstValue, srcValue, lb_const_int(m, t_i32, uint64(fieldFlag))}
			lb_emit_runtime_call(copyProc, "_Block_object_assign", copyArgs)
		}
		{
			srcField := LLVMBuildStructGEP2(disposeProc.Builder, blockLitType, disposeProc.RawInputParameters[0], fieldOffset, "")
			srcValue := lbValue{Type: fieldType, Value: LLVMBuildLoad2(disposeProc.Builder, fieldRawType, srcField, "")}
			fieldFlag := BLOCK_FIELD_IS_OBJECT
			if isBlockObj {
				fieldFlag = BLOCK_FIELD_IS_BLOCK
			}
			disposeArgs := []lbValue{srcValue, lb_const_int(m, t_i32, uint64(fieldFlag))}
			lb_emit_runtime_call(disposeProc, "_Block_object_dispose", disposeArgs)
		}
	}

	lb_end_procedure_body(copyProc)
	lb_end_procedure_body(disposeProc)

	return copyProc, disposeProc
}

func lb_handle_objc_block(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	if !(len(ce.Args) > 0) {
		gb_assert_handler("Assertion Failure", "ce->args.count > 0", "llvm_backend_utility_objc.go", 0)
	}
	m := p.Module
	blockID := m.ObjCNextBlockID
	m.ObjCNextBlockID++

	captureArgCount := len(ce.Args) - 1
	blockResultType := type_of_expr(expr)
	if blockResultType == nil || blockResultType.Kind != Type_Pointer {
		gb_assert_handler("Assertion Failure", "block_result_type != nullptr && block_result_type->kind == Type_Pointer", "llvm_backend_utility_objc.go", 0)
	}

	lbTypeRawptr := lb_type(m, t_rawptr)
	lbTypeI32 := lb_type(m, t_i32)
	lbTypeInt := lb_type(m, t_int)

	userProcValue := lb_build_expr(p, ce.Args[captureArgCount])
	userProc := userProcValue.Type.Proc
	if userProcValue.Type.Kind != Type_Proc {
		gb_assert_handler("Assertion Failure", "user_proc_value.type->kind == Type_Proc", "llvm_backend_utility_objc.go", 0)
	}

	isGlobal := captureArgCount == 0 && userProc.CallingConvention != ProcCC_Odin
	blockForwardArgs := int(userProc.ParamCount) - captureArgCount
	captureFieldsOffset := isize(5)
	if userProc.CallingConvention != ProcCC_Odin {
		captureFieldsOffset = 5
	} else {
		captureFieldsOffset = 6
	}

	procLit := unparen_expr(ce.Args[captureArgCount])
	if procLit.Kind == Ast_Ident {
		procLit = procLit.Ident.Entity.Load().DeclInfo.ProcLit
	}
	if procLit.Kind != Ast_ProcLit {
		gb_assert_handler("Assertion Failure", "proc_lit->kind == Ast_ProcLit", "llvm_backend_utility_objc.go", 0)
	}

	var copyHelper, disposeHelper *lbProcedure
	capturedValues := make([]lbValue, captureArgCount)
	objcCaptures := make([]isize, 0)
	for i := 0; i < captureArgCount; i++ {
		capturedValues[i] = lb_build_expr(p, ce.Args[i])
		if is_type_pointer(capturedValues[i].Type) && is_type_objc_object(capturedValues[i].Type) {
			objcCaptures = append(objcCaptures, isize(i))
		}
	}
	hasObjcFields := len(objcCaptures) > 0

	blockInvokerName := fmt.Sprintf("__$%s::objc_block_invoker_%d", m.ModuleName, blockID)
	invokerArgs := make([]*Type, blockForwardArgs+1)
	invokerArgs[0] = t_rawptr
	if !(blockForwardArgs <= int(userProc.ParamCount)) {
		gb_assert_handler("Assertion Failure", "block_forward_args <= user_proc.param_count", "llvm_backend_utility_objc.go", 0)
	}
	if userProc.ParamCount > 0 {
		userProcParamTypes := userProc.Params.Tuple.Variables
		for i := 0; i < blockForwardArgs; i++ {
			invokerArgs[i+1] = userProcParamTypes[i].Type
		}
	}
	if !(int(userProc.ResultCount) <= 1) {
		gb_assert_handler("Assertion Failure", "user_proc.result_count <= 1", "llvm_backend_utility_objc.go", 0)
	}
	invokerArgsTuple := alloc_type_tuple_from_field_types(invokerArgs, len(invokerArgs), false, true)
	var invokerResultsTuple *Type
	if userProc.ResultCount > 0 {
		invokerResultsTuple = alloc_type_tuple_from_field_types([]*Type{userProc.Results.Tuple.Variables[0].Type}, 1, false, true)
	}
	invokerProcType := alloc_type_proc(nil, invokerArgsTuple, len(invokerArgs),
		invokerResultsTuple, int(userProc.ResultCount), false, ProcCC_CDecl)
	invokerProc := lb_create_dummy_procedure(m, blockInvokerName, invokerProcType)
	LLVMSetLinkage(invokerProc.Value, LLVMPrivateLinkage)
	lb_add_function_type_attributes(invokerProc.Value, lb_get_function_type(m, invokerProcType), ProcCC_CDecl)

	blockLitTypeName := fmt.Sprintf("__$%s::ObjC_Block_Literal_%d", m.ModuleName, blockID)
	blockDescTypeName := fmt.Sprintf("__$%s::ObjC_Block_Descriptor_%d", m.ModuleName, blockID)

	var blockLitType LLVMTypeRef
	var blockDescType LLVMTypeRef
	var blockDescInitializer LLVMValueRef

	{
		blockDescType = LLVMStructCreateNamed(m.Ctx, blockDescTypeName)
		var fieldsTypes [4]LLVMTypeRef
		fieldsTypes[0] = lbTypeInt
		fieldsTypes[1] = lbTypeInt
		fieldsTypes[2] = lbTypeRawptr
		fieldsTypes[3] = lbTypeRawptr
		if hasObjcFields {
			LLVMStructSetBody(blockDescType, fieldsTypes[:], 4, false)
		} else {
			LLVMStructSetBody(blockDescType, fieldsTypes[:], 2, false)
		}
	}
	{
		blockLitType = LLVMStructCreateNamed(m.Ctx, blockLitTypeName)
		fields := make([]LLVMTypeRef, 0)
		fields = append(fields, lbTypeRawptr)
		fields = append(fields, lbTypeI32)
		fields = append(fields, lbTypeI32)
		fields = append(fields, lbTypeRawptr)
		fields = append(fields, blockDescType)
		if userProc.CallingConvention == ProcCC_Odin {
			fields = append(fields, lb_type(m, t_context))
		}
		for _, capArg := range capturedValues {
			fields = append(fields, lb_type(m, capArg.Type))
		}
		LLVMStructSetBody(blockLitType, fields, uint(len(fields)), false)
	}

	if hasObjcFields {
		copyHelper, disposeHelper = lb_create_objc_block_helper_procs(m, blockLitType, captureFieldsOffset, blockID,
			capturedValues, objcCaptures)
	}

	{
		fieldsValues := [4]LLVMValueRef{
			lb_const_int(m, t_int, 0).Value,
			lb_const_int(m, t_int, uint64(lb_sizeof(blockLitType))).Value,
			LLVMValueRef(0),
			LLVMValueRef(0),
		}
		if hasObjcFields {
			fieldsValues[2] = copyHelper.Value
			fieldsValues[3] = disposeHelper.Value
		}
		if hasObjcFields {
			blockDescInitializer = LLVMConstNamedStruct(blockDescType, fieldsValues[:], 4)
		} else {
			blockDescInitializer = LLVMConstNamedStruct(blockDescType, fieldsValues[:], 2)
		}
	}

	descGlobalName := fmt.Sprintf("__$%s::objc_block_desc_%d", m.ModuleName, blockID)
	pDescriptor := LLVMAddGlobal(m.Mod, blockDescType, descGlobalName)
	LLVMSetInitializer(pDescriptor, blockDescInitializer)

	lb_begin_procedure_body(invokerProc)
	{
		callArgs := make([]LLVMValueRef, 0, int(userProc.ParamCount)+2)
		blockLiteralArgIndex := 0
		userProcFT := lb_get_function_type(m, userProcValue.Type)
		var returnKind lbArgKind
		if !(int(userProc.ResultCount) <= 1) {
			gb_assert_handler("Assertion Failure", "user_proc.result_count <= 1", "llvm_backend_utility_objc.go", 0)
		}
		if userProc.ResultCount > 0 {
			returnKind = userProcFT.Return.Kind
			if returnKind == lbArg_Indirect {
				callArgs = append(callArgs, invokerProc.RawInputParameters[0])
				blockLiteralArgIndex = 1
			}
		}
		for i := blockLiteralArgIndex + 1; i < len(invokerProc.RawInputParameters); i++ {
			callArgs = append(callArgs, invokerProc.RawInputParameters[i])
		}
		blockLiteral := invokerProc.RawInputParameters[blockLiteralArgIndex]
		captureArgInUserProcStartIndex := len(userProcFT.Args) - captureArgCount
		if userProc.CallingConvention == ProcCC_Odin {
			captureArgInUserProcStartIndex--
		}
		for i := 0; i < captureArgCount; i++ {
			capValue := LLVMBuildStructGEP2(invokerProc.Builder, blockLitType, blockLiteral, uint(captureFieldsOffset+isize(i)), "")
			capArgIndexInUserProc := captureArgInUserProcStartIndex + i
			if userProcFT.Args[capArgIndexInUserProc].Kind != lbArg_Indirect {
				capValue = OdinLLVMBuildLoad(invokerProc, lb_type(invokerProc.Module, capturedValues[i].Type), capValue)
			}
			callArgs = append(callArgs, capValue)
		}
		if userProc.CallingConvention == ProcCC_Odin {
			pContext := LLVMBuildStructGEP2(invokerProc.Builder, blockLitType, blockLiteral, 5, "context")
			callArgs = append(callArgs, pContext)
		}
		fnp := lb_type_internal_for_procedures_raw(m, userProcValue.Type)
		retVal := LLVMBuildCall2(invokerProc.Builder, fnp, userProcValue.Value, callArgs, uint(len(callArgs)), "")
		if userProc.ResultCount > 0 && returnKind != lbArg_Indirect {
			LLVMBuildRet(invokerProc.Builder, retVal)
		} else {
			LLVMBuildRetVoid(invokerProc.Builder)
		}
	}
	lb_end_procedure_body(invokerProc)

	const BLOCK_HAS_COPY_DISPOSE = 1 << 25
	const BLOCK_IS_GLOBAL = 1 << 28
	rawFlags := 0
	if isGlobal {
		rawFlags = BLOCK_IS_GLOBAL
	}
	if hasObjcFields {
		rawFlags |= BLOCK_HAS_COPY_DISPOSE
	}
	blockVarName := fmt.Sprintf("__$objc_block_literal_%d", blockID)
	blockResult := lbValue{}
	blockResult.Type = blockResultType

	isaName := "_NSConcreteGlobalBlock"
	if !isGlobal {
		isaName = "_NSConcreteStackBlock"
	}
	isaVal := lb_find_runtime_value(m, isaName)
	flagsVal := lb_const_int(m, t_i32, uint64(rawFlags))
	reservedVal := lb_const_int(m, t_i32, 0)

	if isGlobal {
		pBlockLit := LLVMAddGlobal(m.Mod, blockLitType, blockVarName)
		blockResult.Value = pBlockLit
		fieldsValues := [5]LLVMValueRef{
			isaVal.Value,
			flagsVal.Value,
			reservedVal.Value,
			invokerProc.Value,
			pDescriptor,
		}
		gBlockLitInitializer := LLVMConstNamedStruct(blockLitType, fieldsValues[:], 5)
		LLVMSetInitializer(pBlockLit, gBlockLitInitializer)
	} else {
		pBlockLit := llvm_alloca(p, blockLitType, isize(lb_alignof(blockLitType)), blockVarName)
		blockResult.Value = pBlockLit
		fIsa := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 0, "isa")
		fFlags := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 1, "flags")
		fReserved := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 2, "reserved")
		fInvoke := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 3, "invoke")
		fDescriptor := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 4, "descriptor")
		LLVMBuildStore(p.Builder, isaVal.Value, fIsa)
		LLVMBuildStore(p.Builder, flagsVal.Value, fFlags)
		LLVMBuildStore(p.Builder, reservedVal.Value, fReserved)
		LLVMBuildStore(p.Builder, invokerProc.Value, fInvoke)
		LLVMBuildStore(p.Builder, pDescriptor, fDescriptor)

		if userProc.CallingConvention == ProcCC_Odin {
			fContext := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, 5, "context")
			pCurrentContext := lb_find_or_generate_context_ptr(p)
			contextSize := LLVMConstInt(LLVMInt64TypeInContext(m.Ctx), uint64(lb_sizeof(lb_type(m, t_context))), false)
			LLVMBuildMemCpy(p.Builder, fContext, lb_try_get_alignment(fContext, 1),
				pCurrentContext.Addr.Value, lb_try_get_alignment(pCurrentContext.Addr.Value, 1), contextSize)
		}
		for i := 0; i < len(capturedValues); i++ {
			captureArg := capturedValues[i]
			fieldIndex := uint(captureFieldsOffset + isize(i))
			fCapture := LLVMBuildStructGEP2(p.Builder, blockLitType, pBlockLit, fieldIndex, "capture_arg")
			fCaptureVal := lbValue{Type: alloc_type_pointer(captureArg.Type), Value: fCapture}
			lb_emit_store(p, fCaptureVal, captureArg)
		}
	}
	return blockResult
}

func lb_handle_objc_block_invoke(p *lbProcedure, expr *Ast) lbValue {
	return lbValue{}
}

func lb_handle_objc_super(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	if !(len(ce.Args) == 1) {
		gb_assert_handler("Assertion Failure", "ce->args.count == 1", "llvm_backend_utility_objc.go", 0)
	}
	return lb_build_expr(p, ce.Args[0])
}

func lb_handle_objc_find_selector(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	tav := ce.Args[0].TAV
	if tav.Value.Kind != ExactValue_String {
		gb_assert_handler("Assertion Failure", "tav.value.kind == ExactValue_String", "llvm_backend_utility_objc.go", 0)
	}
	name := tav.Value.ValueString
	return lb_addr_load(p, lb_handle_objc_find_or_register_selector(p, name))
}

func lb_handle_objc_register_selector(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	m := p.Module
	tav := ce.Args[0].TAV
	if tav.Value.Kind != ExactValue_String {
		gb_assert_handler("Assertion Failure", "tav.value.kind == ExactValue_String", "llvm_backend_utility_objc.go", 0)
	}
	name := tav.Value.ValueString
	dst := lb_handle_objc_find_or_register_selector(p, name)
	args := make([]lbValue, 1)
	args[0] = lb_const_value(m, t_cstring, exact_value_string(name))
	ptr := lb_emit_runtime_call(p, "sel_registerName", args)
	lb_addr_store(p, dst, ptr)
	return lb_addr_load(p, dst)
}

func lb_handle_objc_find_class(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	tav := ce.Args[0].TAV
	if tav.Value.Kind != ExactValue_String {
		gb_assert_handler("Assertion Failure", "tav.value.kind == ExactValue_String", "llvm_backend_utility_objc.go", 0)
	}
	name := tav.Value.ValueString
	return lb_addr_load(p, lb_handle_objc_find_or_register_class(p, name, nil))
}

func lb_handle_objc_register_class(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	m := p.Module
	tav := ce.Args[0].TAV
	if tav.Value.Kind != ExactValue_String {
		gb_assert_handler("Assertion Failure", "tav.value.kind == ExactValue_String", "llvm_backend_utility_objc.go", 0)
	}
	name := tav.Value.ValueString
	dst := lb_handle_objc_find_or_register_class(p, name, nil)
	args := make([]lbValue, 3)
	args[0] = lb_const_nil(m, t_objc_Class)
	args[1] = lb_const_nil(m, t_objc_Class)
	args[2] = lb_const_int(m, t_uint, 0)
	ptr := lb_emit_runtime_call(p, "objc_allocateClassPair", args)
	lb_addr_store(p, dst, ptr)
	return lb_addr_load(p, dst)
}

func lb_handle_objc_id(p *lbProcedure, expr *Ast) lbValue {
	tav := type_and_value_of_expr(expr)
	if tav.Mode == Addressing_Type {
		type_ := tav.Type
		if type_.Kind != Type_Named {
			gb_assert_handler("Assertion Failure", "type->kind == Type_Named", "llvm_backend_utility_objc.go", 0)
		}
		e := type_.Named.TypeName
		if e.Kind != Entity_TypeName {
			gb_assert_handler("Assertion Failure", "e->kind == Entity_TypeName", "llvm_backend_utility_objc.go", 0)
		}
		name := e.TypeName.ObjcClassName
		var classImplType *Type
		if e.TypeName.ObjcIsImplementation {
			classImplType = type_
		}
		return lb_addr_load(p, lb_handle_objc_find_or_register_class(p, name, classImplType))
	}
	return lb_build_expr(p, expr)
}

func lb_handle_objc_send(p *lbProcedure, expr *Ast) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	m := p.Module
	info := m.Info
	data := map_must_get(&info.ObjcMsgSendTypes, expr)
	if data.ProcType == nil {
		gb_assert_handler("Assertion Failure", "data.proc_type != nullptr", "llvm_backend_utility_objc.go", 0)
	}
	if !(len(ce.Args) >= 3) {
		gb_assert_handler("Assertion Failure", "ce->args.count >= 3", "llvm_backend_utility_objc.go", 0)
	}
	args := make([]lbValue, 0, len(ce.Args)-1)
	id := lb_handle_objc_id(p, ce.Args[1])
	selExpr := ce.Args[2]
	if selExpr.TAV.Value.Kind != ExactValue_String {
		gb_assert_handler("Assertion Failure", "sel_expr->tav.value.kind == ExactValue_String", "llvm_backend_utility_objc.go", 0)
	}
	sel := lb_addr_load(p, lb_handle_objc_find_or_register_selector(p, selExpr.TAV.Value.ValueString))
	args = append(args, id)
	args = append(args, sel)
	for i := 3; i < len(ce.Args); i++ {
		arg := lb_build_expr(p, ce.Args[i])
		args = append(args, arg)
	}
	var theProc lbValue
	switch data.Kind {
	default:
		gb_assert_handler("Panic", 0, "llvm_backend_utility_objc.go", 0, "unhandled ObjcMsgKind %u", data.Kind)
	case ObjcMsg_normal:
		theProc = lb_lookup_runtime_procedure(m, "objc_msgSend")
	case ObjcMsg_fpret:
		theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_fpret")
	case ObjcMsg_fp2ret:
		theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_fp2ret")
	case ObjcMsg_stret:
		theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_stret")
	}
	theProc = lb_emit_conv(p, theProc, data.ProcType)
	return lb_emit_call(p, theProc, args)
}

func lb_handle_objc_auto_send(p *lbProcedure, expr *Ast, argValues []lbValue) lbValue {
	ce := &expr.CallExpr
	if expr.Kind != Ast_CallExpr {
		gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_CallExpr", "llvm_backend_utility_objc.go", 0)
	}
	m := p.Module
	info := m.Info
	data := map_must_get(&info.ObjcMsgSendTypes, expr)
	procType := data.ProcType
	if procType == nil {
		gb_assert_handler("Assertion Failure", "proc_type != nullptr", "llvm_backend_utility_objc.go", 0)
	}
	objcMethodEnt := entity_of_node(ce.Proc)
	if objcMethodEnt == nil {
		gb_assert_handler("Assertion Failure", "objc_method_ent != nullptr", "llvm_backend_utility_objc.go", 0)
	}
	if objcMethodEnt.Kind != Entity_Procedure {
		gb_assert_handler("Assertion Failure", "objc_method_ent->kind == Entity_Procedure", "llvm_backend_utility_objc.go", 0)
	}
	if objcMethodEnt.Procedure.ObjcSelectorName.Len <= 0 {
		gb_assert_handler("Assertion Failure", "objc_method_ent->Procedure.objc_selector_name.len > 0", "llvm_backend_utility_objc.go", 0)
	}
	proc := procType.Proc
	if !(int(proc.ParamCount) >= 2) {
		gb_assert_handler("Assertion Failure", "proc.param_count >= 2", "llvm_backend_utility_objc.go", 0)
	}
	var objcSuperOrigType *Type
	if len(ce.Args) > 0 {
		objcSuperOrigType = unparen_expr(ce.Args[0]).TAV.ObjCSuperTarget
	}
	argOffset := isize(1)
	id := lbValue{}
	if !objcMethodEnt.Procedure.IsObjcClassMethod {
		if !(len(ce.Args) > 0) {
			gb_assert_handler("Assertion Failure", "ce->args.count > 0", "llvm_backend_utility_objc.go", 0)
		}
		id = argValues[0]
		if objcSuperOrigType != nil {
			if objcSuperOrigType.Kind != Type_Named {
				gb_assert_handler("Assertion Failure", "objc_super_orig_type->kind == Type_Named", "llvm_backend_utility_objc.go", 0)
			}
			tn := objcSuperOrigType.Named.TypeName.TypeName
			var clsImplType *Type
			if tn.ObjcIsImplementation {
				clsImplType = objcSuperOrigType
			}
			pSupercls := lb_handle_objc_find_or_register_class(p, tn.ObjcClassName, clsImplType)
			supercls := lb_addr_load(p, pSupercls)
			pObjcSuper := lb_add_local_generated(p, t_objc_super, false)
			fID := lb_emit_struct_ep(p, pObjcSuper.Addr, 0)
			fSuperclass := lb_emit_struct_ep(p, pObjcSuper.Addr, 1)
			id = lb_emit_conv(p, id, t_objc_id)
			lb_emit_store(p, fID, id)
			lb_emit_store(p, fSuperclass, supercls)
			id = pObjcSuper.Addr
		}
	} else {
		objcClass := objcMethodEnt.Procedure.ObjcClass
		if ce.Proc.Kind == Ast_SelectorExpr {
			se := &ce.Proc.SelectorExpr
			if ce.Proc.Kind != Ast_SelectorExpr {
				gb_assert_handler("Assertion Failure", "(ce->proc)->kind == Ast_SelectorExpr", "llvm_backend_utility_objc.go", 0)
			}
			if !(se.Expr.TAV.Mode == Addressing_Type && se.Expr.TAV.Type.Kind == Type_Named) {
				gb_assert_handler("Assertion Failure", `se->expr->tav.mode == Addressing_Type && se->expr->tav.type->kind == Type_Named`, "llvm_backend_utility_objc.go", 0)
			}
			objcClass = entity_from_expr(se.Expr)
			if objcClass == nil {
				gb_assert_handler("Assertion Failure", "objc_class", "llvm_backend_utility_objc.go", 0)
			}
			if objcClass.Kind != Entity_TypeName {
				gb_assert_handler("Assertion Failure", "objc_class->kind == Entity_TypeName", "llvm_backend_utility_objc.go", 0)
			}
			if objcClass.TypeName.IsTypeAlias {
				objcClass = objcClass.Type.Named.TypeName
			}
			if objcClass.TypeName.ObjcClassName == "" {
				gb_assert_handler("Assertion Failure", `objc_class->TypeName.objc_class_name != ""`, "llvm_backend_utility_objc.go", 0)
			}
		}
		var classImplType *Type
		if objcClass.TypeName.ObjcIsImplementation {
			classImplType = objcClass.Type
		}
		id = lb_addr_load(p, lb_handle_objc_find_or_register_class(p, objcClass.TypeName.ObjcClassName, classImplType))
		argOffset = 0
	}
	sel := lb_addr_load(p, lb_handle_objc_find_or_register_selector(p, objcMethodEnt.Procedure.ObjcSelectorName))
	args := make([]lbValue, 0, len(argValues)+2-int(argOffset))
	args = append(args, id)
	args = append(args, sel)
	for i := argOffset; i < isize(len(ce.Args)); i++ {
		args = append(args, argValues[i])
	}
	var theProc lbValue
	if objcSuperOrigType == nil {
		switch data.Kind {
		default:
			gb_assert_handler("Panic", 0, "llvm_backend_utility_objc.go", 0, "unhandled ObjcMsgKind %u", data.Kind)
		case ObjcMsg_normal:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSend")
		case ObjcMsg_fpret:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_fpret")
		case ObjcMsg_fp2ret:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_fp2ret")
		case ObjcMsg_stret:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSend_stret")
		}
	} else {
		switch data.Kind {
		default:
			gb_assert_handler("Panic", 0, "llvm_backend_utility_objc.go", 0, "unhandled ObjcMsgKind %u", data.Kind)
		case ObjcMsg_normal, ObjcMsg_fpret, ObjcMsg_fp2ret:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSendSuper2")
		case ObjcMsg_stret:
			theProc = lb_lookup_runtime_procedure(m, "objc_msgSendSuper2_stret")
		}
	}
	theProc = lb_emit_conv(p, theProc, data.ProcType)
	return lb_emit_call(p, theProc, args)
}
