package cmd

import (
	"unsafe"
)

type ObjcMethodData struct {
	Ac     AttributeContext
	Entity *Entity
}

type lbObjCGlobalClass struct {
	g           lbObjCGlobal
	ClassValue  lbValue
	ClassGlobal lbAddr
}

func lb_create_objc_names(main_module *lbModule) *lbProcedure {
	if buildContext.Metrics.Os != TargetOs_darwin {
		return nil
	}
	procType := alloc_type_proc(nil, nil, 0, nil, 0, false, ProcCC_CDecl)
	p := lb_create_dummy_procedure(main_module, String{Data: unsafe.SliceData(unsafe.StringData("__$init_objc_names")), Len: isize(len("__$init_objc_names"))}, procType)
	lb_add_attribute_to_proc(p.Module, p.Value, "nounwind")
	p.IsStartup = true
	return p
}

func lb_get_objc_type_encoding(t *Type, pointer_depth ...isize) String {
	pd := isize(0)
	if len(pointer_depth) > 0 {
		pd = pointer_depth[0]
	}

	switch t.Kind {
	case Type_Basic:
		switch t.Basic.Kind {
		case Basic_Invalid:
			return String{Data: unsafe.SliceData(unsafe.StringData("?")), Len: 1}
		case Basic_llvm_bool, Basic_bool, Basic_b8:
			return String{Data: unsafe.SliceData(unsafe.StringData("B")), Len: 1}
		case Basic_b16, Basic_u8:
			return String{Data: unsafe.SliceData(unsafe.StringData("C")), Len: 1}
		case Basic_b32, Basic_u32le, Basic_u32, Basic_u32be, Basic_rune:
			return String{Data: unsafe.SliceData(unsafe.StringData("I")), Len: 1}
		case Basic_b64, Basic_i64, Basic_i64le, Basic_i64be:
			return String{Data: unsafe.SliceData(unsafe.StringData("q")), Len: 1}
		case Basic_i8:
			return String{Data: unsafe.SliceData(unsafe.StringData("c")), Len: 1}
		case Basic_i16, Basic_i16le, Basic_i16be, Basic_f16, Basic_f16le, Basic_f16be:
			return String{Data: unsafe.SliceData(unsafe.StringData("s")), Len: 1}
		case Basic_u16, Basic_u16le, Basic_u16be:
			return String{Data: unsafe.SliceData(unsafe.StringData("S")), Len: 1}
		case Basic_i32, Basic_i32le, Basic_i32be:
			return String{Data: unsafe.SliceData(unsafe.StringData("i")), Len: 1}
		case Basic_u64, Basic_u64le, Basic_u64be:
			return String{Data: unsafe.SliceData(unsafe.StringData("Q")), Len: 1}
		case Basic_i128, Basic_i128le, Basic_i128be:
			return String{Data: unsafe.SliceData(unsafe.StringData("t")), Len: 1}
		case Basic_u128, Basic_u128le, Basic_u128be:
			return String{Data: unsafe.SliceData(unsafe.StringData("T")), Len: 1}
		case Basic_f32, Basic_f32le, Basic_f32be:
			return String{Data: unsafe.SliceData(unsafe.StringData("f")), Len: 1}
		case Basic_f64, Basic_f64le, Basic_f64be:
			return String{Data: unsafe.SliceData(unsafe.StringData("d")), Len: 1}
		case Basic_complex32:
			return String{Data: unsafe.SliceData(unsafe.StringData("{complex32=ss}")), Len: isize(len("{complex32=ss}"))}
		case Basic_complex64:
			return String{Data: unsafe.SliceData(unsafe.StringData("{complex64=ff}")), Len: isize(len("{complex64=ff}"))}
		case Basic_complex128:
			return String{Data: unsafe.SliceData(unsafe.StringData("{complex128=dd}")), Len: isize(len("{complex128=dd}"))}
		case Basic_quaternion64:
			return String{Data: unsafe.SliceData(unsafe.StringData("{quaternion64=ssss}")), Len: isize(len("{quaternion64=ssss}"))}
		case Basic_quaternion128:
			return String{Data: unsafe.SliceData(unsafe.StringData("{quaternion128=ffff}")), Len: isize(len("{quaternion128=ffff}"))}
		case Basic_quaternion256:
			return String{Data: unsafe.SliceData(unsafe.StringData("{quaternion256=dddd}")), Len: isize(len("{quaternion256=dddd}"))}
		case Basic_int:
			if buildContext.Metrics.IntSize == 4 {
				return String{Data: unsafe.SliceData(unsafe.StringData("i")), Len: 1}
			}
			return String{Data: unsafe.SliceData(unsafe.StringData("q")), Len: 1}
		case Basic_uint:
			if buildContext.Metrics.IntSize == 4 {
				return String{Data: unsafe.SliceData(unsafe.StringData("I")), Len: 1}
			}
			return String{Data: unsafe.SliceData(unsafe.StringData("Q")), Len: 1}
		case Basic_uintptr, Basic_rawptr:
			return String{Data: unsafe.SliceData(unsafe.StringData("^v")), Len: 2}
		case Basic_string:
			if buildContext.Metrics.IntSize == 4 {
				return String{Data: unsafe.SliceData(unsafe.StringData("{string=*i}")), Len: isize(len("{string=*i}"))}
			}
			return String{Data: unsafe.SliceData(unsafe.StringData("{string=*q}")), Len: isize(len("{string=*q}"))}
		case Basic_string16:
			if buildContext.Metrics.IntSize == 4 {
				return String{Data: unsafe.SliceData(unsafe.StringData("{string16=*i}")), Len: isize(len("{string16=*i}"))}
			}
			return String{Data: unsafe.SliceData(unsafe.StringData("{string16=*q}")), Len: isize(len("{string16=*q}"))}
		case Basic_cstring, Basic_cstring16:
			return String{Data: unsafe.SliceData(unsafe.StringData("*")), Len: 1}
		case Basic_any:
			return String{Data: unsafe.SliceData(unsafe.StringData("{any=^v^v}")), Len: isize(len("{any=^v^v}"))}
		case Basic_typeid:
			if !(t.Basic.Size == 8) {
				gb_assert_handler("Assertion Failure", "t.Basic.Size == 8", "llvm_backend_objc.go", 0)
			}
			return String{Data: unsafe.SliceData(unsafe.StringData("q")), Len: 1}
		case Basic_UntypedBool, Basic_UntypedInteger, Basic_UntypedFloat,
			Basic_UntypedComplex, Basic_UntypedQuaternion, Basic_UntypedString,
			Basic_UntypedRune, Basic_UntypedNil, Basic_UntypedUninit:
			gb_assert_handler("Panic", 0, "llvm_backend_objc.go", 0, "Untyped types cannot be @encoded()")
			return String{Data: unsafe.SliceData(unsafe.StringData("?")), Len: 1}
		}

	case Type_Named, Type_Struct, Type_Union:
		base := t
		if base.Kind == Type_Named {
			base = base_type(base)
			if base.Kind != Type_Struct && base.Kind != Type_Union {
				return lb_get_objc_type_encoding(base, pd)
			}
		}
		isUnion := base.Kind == Type_Union
		if !isUnion {
			if has_type_got_objc_class_attribute(t) && pd == 0 {
				return String{Data: unsafe.SliceData(unsafe.StringData("#")), Len: 1}
			}
		}
		if is_type_objc_object(base) {
			return String{Data: unsafe.SliceData(unsafe.StringData("@")), Len: 1}
		}
		s := gb_string_make_reserve(temporary_allocator(), 16)
		if isUnion {
			s = gb_string_append_length(s, "(", 1)
		} else {
			s = gb_string_append_length(s, "{", 1)
		}
		if t.Kind == Type_Named {
			s = gb_string_append_length(s, t.Named.Name.Text, t.Named.Name.Len)
		}
		if pd < 2 {
			s = gb_string_append_length(s, "=", 1)
			if !isUnion {
				for _, f := range base.Struct.Fields {
					fieldType := lb_get_objc_type_encoding(f.Type, pd)
					s = gb_string_append_length(s, fieldType.Text, fieldType.Len)
				}
			} else {
				for _, v := range base.Union.Variants {
					variantType := lb_get_objc_type_encoding(v, pd)
					s = gb_string_append_length(s, variantType.Text, variantType.Len)
				}
			}
		}
		if isUnion {
			s = gb_string_append_length(s, ")", 1)
		} else {
			s = gb_string_append_length(s, "}", 1)
		}
		return make_string_c(s)

	case Type_Generic:
		gb_assert_handler("Panic", 0, "llvm_backend_objc.go", 0, "Generic types cannot be @encoded()")
		return String{Data: unsafe.SliceData(unsafe.StringData("?")), Len: 1}

	case Type_Pointer:
		if internal_check_is_assignable_to(t, t_objc_SEL) {
			return String{Data: unsafe.SliceData(unsafe.StringData(":")), Len: 1}
		}
		if internal_check_is_assignable_to(t, t_objc_Class) {
			return String{Data: unsafe.SliceData(unsafe.StringData("#")), Len: 1}
		}
		pointee := lb_get_objc_type_encoding(t.Pointer.Elem, pd+1)
		if pd == 0 && pointee.Len == 1 && *pointee.Text == '@' {
			return pointee
		}
		caret := String{Data: unsafe.SliceData(unsafe.StringData("^")), Len: 1}
		return concatenate_strings(temporary_allocator(), caret, pointee)

	case Type_MultiPointer:
		caret := String{Data: unsafe.SliceData(unsafe.StringData("^")), Len: 1}
		return concatenate_strings(temporary_allocator(), caret, lb_get_objc_type_encoding(t.MultiPointer.Elem, pd+1))

	case Type_Array:
		typeStr := lb_get_objc_type_encoding(t.Array.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+8)
		s = gb_string_append_fmt(s, "[%lld%.*s]", t.Array.Count, int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len))
		return make_string_c(s)

	case Type_EnumeratedArray:
		typeStr := lb_get_objc_type_encoding(t.EnumeratedArray.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+8)
		s = gb_string_append_fmt(s, "[%lld%.*s]", t.EnumeratedArray.Count, int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len))
		return make_string_c(s)

	case Type_Slice:
		typeStr := lb_get_objc_type_encoding(t.Slice.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+8)
		intSuffix := "q"
		if buildContext.Metrics.IntSize == 4 {
			intSuffix = "i"
		}
		s = gb_string_append_fmt(s, "{slice=^%.*s%s}", int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len), intSuffix)
		return make_string_c(s)

	case Type_DynamicArray:
		typeStr := lb_get_objc_type_encoding(t.DynamicArray.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+8)
		intSuffix := "q"
		if buildContext.Metrics.IntSize == 4 {
			intSuffix = "i"
		}
		s = gb_string_append_fmt(s, "{dynamic=^%.*s%s%sAllocator={?^v}}", int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len), intSuffix, intSuffix)
		return make_string_c(s)

	case Type_Map:
		return String{Data: unsafe.SliceData(unsafe.StringData("{^v^v{Allocator=?^v}}")), Len: isize(len("{^v^v{Allocator=?^v}}"))}

	case Type_Enum:
		return lb_get_objc_type_encoding(t.Enum.BaseType, pd)

	case Type_Tuple:
		return String{Data: unsafe.SliceData(unsafe.StringData("?")), Len: 1}

	case Type_Proc:
		return String{Data: unsafe.SliceData(unsafe.StringData("?")), Len: 1}

	case Type_BitSet:
		bitsetIntegerType := t.BitSet.Underlying
		if bitsetIntegerType == nil {
			switch t.CachedSize {
			case 1:
				bitsetIntegerType = t_u8
			case 2:
				bitsetIntegerType = t_u16
			case 4:
				bitsetIntegerType = t_u32
			case 8:
				bitsetIntegerType = t_u64
			case 16:
				bitsetIntegerType = t_u128
			}
		}
		if bitsetIntegerType == nil {
			gb_assert_handler("Assertion Failure", "bitsetIntegerType", "llvm_backend_objc.go", 0,
				"Could not determine bit_set integer size for objc_type_encoding")
		}
		return lb_get_objc_type_encoding(bitsetIntegerType, pd)

	case Type_SimdVector:
		typeStr := lb_get_objc_type_encoding(t.SimdVector.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+5)
		s = gb_string_append_fmt(s, "[%lld%.*s]", t.SimdVector.Count, int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len))
		return make_string_c(s)

	case Type_Matrix:
		typeStr := lb_get_objc_type_encoding(t.Matrix.Elem, pd)
		s := gb_string_make_reserve(temporary_allocator(), typeStr.Len+5)
		elementCount := t.Matrix.ColumnCount * t.Matrix.RowCount
		s = gb_string_append_fmt(s, "[%lld%.*s]", elementCount, int(typeStr.Len), unsafe.String(typeStr.Text, typeStr.Len))
		return make_string_c(s)

	case Type_BitField:
		return lb_get_objc_type_encoding(t.BitField.BackingType, pd)

	case Type_SoaPointer:
		s := gb_string_make_reserve(temporary_allocator(), 8)
		intSuffix := "q"
		if buildContext.Metrics.IntSize == 4 {
			intSuffix = "i"
		}
		s = gb_string_append_fmt(s, "{=^v%s}", intSuffix)
		return make_string_c(s)
	}

	gb_assert_handler("Panic", 0, "llvm_backend_objc.go", 0, "Unreachable")
	return String{}
}

func lb_register_objc_thing(
	handled map[string]struct{},
	m *lbModule,
	classImpls *[]lbObjCGlobalClass,
	classMap map[string]lbObjCGlobalClass,
	p *lbProcedure,
	g lbObjCGlobal,
	call string,
) {
	key := string(g.Name.Text[:g.Name.Len])
	if _, ok := handled[key]; ok {
		return
	}
	handled[key] = struct{}{}

	var addr lbAddr
	found := string_map_get(&m.Members, g.GlobalName)
	if found != nil {
		addr = lb_addr(*found)
	} else {
		v := lbValue{}
		t := lb_type(m, g.Type)
		v.Value = LLVMAddGlobal(m.Mod, t, unsafe.String(g.GlobalName, gb_string_length(g.GlobalName)))
		v.Type = alloc_type_pointer(g.Type)
		addr = lb_addr(v)
		LLVMSetInitializer(v.Value, LLVMConstNull(t))
	}

	if g.ClassImplType != nil {
		tn := &g.ClassImplType.Named.TypeName.TypeName
		superclass := tn.ObjcSuperclass
		if superclass != nil {
			superKey := string(superclass.Named.TypeName.TypeName.ObjcClassName.Text[:superclass.Named.TypeName.TypeName.ObjcClassName.Len])
			superclassGlobal := classMap[superKey]
			lb_register_objc_thing(handled, m, classImpls, classMap, p, superclassGlobal.g, call)
			if superclassGlobal.ClassGlobal.Addr.Value == 0 {
				gb_assert_handler("Assertion Failure", "superclassGlobal.ClassGlobal.Addr.Value", "llvm_backend_objc.go", 0)
			}
		}
		implGlobal := lbObjCGlobalClass{
			g:           g,
			ClassGlobal: addr,
		}
		*classImpls = append(*classImpls, implGlobal)
		if classGlobal, ok := classMap[key]; ok {
			classGlobal.ClassGlobal = addr
			classMap[key] = classGlobal
		}
	} else {
		className := lb_const_value(m, t_cstring, exact_value_string(g.Name))
		callArgs := []lbValue{className}
		classPtr := lb_emit_runtime_call(p, call, callArgs)
		lb_addr_store(p, addr, classPtr)
		if classGlobal, ok := classMap[key]; ok {
			classGlobal.ClassValue = classPtr
			classMap[key] = classGlobal
		}
	}
}

func lb_finalize_objc_names(gen *lbGenerator, p *lbProcedure) {
	if p == nil {
		return
	}
	m := p.Module
	if m != &p.Module.Gen.DefaultModule {
		gb_assert_handler("Assertion Failure", "m == &p.Module.Gen.DefaultModule", "llvm_backend_objc.go", 0)
	}

	handled := make(map[string]struct{})
	classImpls := make([]lbObjCGlobalClass, 0, 16)

	for {
		e := mpsc_dequeue[*Entity](&gen.Info.objc_class_implementations)
		if e == nil {
			break
		}
		if !(e.Kind == Entity_TypeName && e.TypeName.ObjcIsImplementation) {
			gb_assert_handler("Assertion Failure", "e->kind == Entity_TypeName && e->TypeName.objc_is_implementation", "llvm_backend_objc.go", 0)
		}
		lb_handle_objc_find_or_register_class(p, e.TypeName.ObjcClassName, e.Type)
	}

	classSet := make(map[*Type]struct{})
	referencedClasses := make([]lbObjCGlobal, 0, gen.ObjCClasses.Count+16)

	for {
		g := mpsc_dequeue[lbObjCGlobal](&gen.ObjCClasses)
		if g == nil {
			break
		}
		referencedClasses = append(referencedClasses, *g)
		cls := g.ClassImplType
		for cls != nil {
			if _, ok := classSet[cls]; ok {
				break
			}
			classSet[cls] = struct{}{}
			if !(cls.Kind == Type_Named) {
				gb_assert_handler("Assertion Failure", "cls->kind == Type_Named", "llvm_backend_objc.go", 0)
			}
			cls = cls.Named.TypeName.TypeName.ObjcSuperclass
		}
	}

	for cls := range classSet {
		tn := &cls.Named.TypeName.TypeName
		var classImpl *Type
		if tn.ObjcIsImplementation {
			classImpl = cls
		}
		lb_handle_objc_find_or_register_class(p, tn.ObjcClassName, classImpl)
	}

	for {
		g := mpsc_dequeue[lbObjCGlobal](&gen.ObjCClasses)
		if g == nil {
			break
		}
		referencedClasses = append(referencedClasses, *g)
	}

	globalClassMap := make(map[string]lbObjCGlobalClass, gen.ObjCClasses.Count)
	for _, g := range referencedClasses {
		key := string(g.Name.Text[:g.Name.Len])
		globalClassMap[key] = lbObjCGlobalClass{g: g}
	}

	LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	lb_begin_procedure_body(p)

	for _, kv := range globalClassMap {
		lb_register_objc_thing(handled, m, &classImpls, globalClassMap, p, kv.g, "objc_lookUpClass")
	}

	for _, cd := range classImpls {
		g := cd.g
		classType := g.ClassImplType
		methods := map_get(&m.Info.objc_method_implementations, classType)
		if methods == nil {
			continue
		}
		for _, md := range *methods {
			lb_handle_objc_find_or_register_selector(p, md.Ac.ObjcSelector)
		}
	}

	for {
		g := mpsc_dequeue[lbObjCGlobal](&gen.ObjCSelectors)
		if g == nil {
			break
		}
		lb_register_objc_thing(handled, m, &classImpls, globalClassMap, p, *g, "sel_registerName")
	}

	ivarMap := make(map[*Type]lbObjCGlobal)

	for {
		g := mpsc_dequeue[lbObjCGlobal](&gen.ObjCIvars)
		if g == nil {
			break
		}
		ivarMap[g.ClassImplType] = *g
	}

	for _, cd := range classImpls {
		g := cd.g
		classType := g.ClassImplType
		classPtrType := alloc_type_pointer(classType)
		classValue := lbValue{}

		{
			superclassValue := lb_const_nil(m, t_objc_Class)
			tn := &classType.Named.TypeName.TypeName
			superclass := tn.ObjcSuperclass
			if superclass != nil {
				superKey := string(superclass.Named.TypeName.TypeName.ObjcClassName.Text[:superclass.Named.TypeName.TypeName.ObjcClassName.Len])
				superclassValue = globalClassMap[superKey].ClassValue
			}
			classArgs := []lbValue{superclassValue, lb_const_value(m, t_cstring, exact_value_string(g.Name)), lb_const_int(m, t_uint, 0)}
			classValue = lb_emit_runtime_call(p, "objc_allocateClassPair", classArgs)
			classKey := string(tn.ObjcClassName.Text[:tn.ObjcClassName.Len])
			mappedGlobal := globalClassMap[classKey]
			lb_addr_store(p, mappedGlobal.ClassGlobal, classValue)
			mappedGlobal.ClassValue = classValue
			globalClassMap[classKey] = mappedGlobal
		}

		ivarType := classType.Named.TypeName.TypeName.ObjcIvar
		contextProvider := classType.Named.TypeName.TypeName.ObjcContextProvider
		var contextProviderSelfPtrType *Type
		var contextProviderSelfNamedType *Type
		isContextProviderIvar := false
		contextProviderProcValue := lbValue{}
		if contextProvider != nil {
			contextProviderProcValue = lb_find_procedure_value_from_entity(m, contextProvider)
			contextProviderSelfPtrType = base_type(contextProvider.Type.Proc.Params.Tuple.Variables[0].Type)
			if !(contextProviderSelfPtrType.Kind == Type_Pointer) {
				gb_assert_handler("Assertion Failure", "contextProviderSelfPtrType->kind == Type_Pointer", "llvm_backend_objc.go", 0)
			}
			contextProviderSelfNamedType = base_named_type(type_deref(contextProviderSelfPtrType))
			isContextProviderIvar = ivarType != nil && internal_check_is_assignable_to(contextProviderSelfNamedType, ivarType)
		}

		methods := map_get(&m.Info.objc_method_implementations, classType)
		if methods == nil {
			continue
		}

		metaClassValue := lbValue{}
		for _, md := range *methods {
			if !md.Ac.ObjcIsClassMethod {
				continue
			}
			metaClassValue = lb_emit_runtime_call(p, "object_getClass", []lbValue{classValue})
			break
		}

		for _, md := range *methods {
			if !(md.Entity.Kind == Entity_Procedure) {
				gb_assert_handler("Assertion Failure", "md.Entity->kind == Entity_Procedure", "llvm_backend_objc.go", 0)
			}
			methodType := md.Entity.Type
			procName := make_string_c("__$objc_method::")
			procName = concatenate_strings(temporary_allocator(), procName, g.Name)
			colonStr := String{Data: unsafe.SliceData(unsafe.StringData("::")), Len: 2}
			procName = concatenate_strings(temporary_allocator(), procName, colonStr)
			procName = concatenate_strings(permanent_allocator(), procName, md.Ac.ObjcName)

			wrapperArgs := make([]*Type, 2, 8)
			if md.Ac.ObjcIsClassMethod {
				wrapperArgs[0] = t_objc_Class
			} else {
				wrapperArgs[0] = classPtrType
			}
			wrapperArgs[1] = t_objc_SEL

			methodParamCount := methodType.Proc.ParamCount
			methodParamOffset := isize(0)
			if !md.Ac.ObjcIsClassMethod {
				if !(methodParamCount >= 1) {
					gb_assert_handler("Assertion Failure", "methodParamCount >= 1", "llvm_backend_objc.go", 0)
				}
				methodParamCount -= 1
				methodParamOffset = 1
			}
			for i := isize(0); i < methodParamCount; i++ {
				wrapperArgs = append(wrapperArgs, methodType.Proc.Params.Tuple.Variables[methodParamOffset+i].Type)
			}

			wrapperArgsTuple := alloc_type_tuple_from_field_types(wrapperArgs, len(wrapperArgs), false, true)
			var wrapperResultsTuple *Type
			if methodType.Proc.ResultCount > 0 {
				if !(methodType.Proc.ResultCount == 1) {
					gb_assert_handler("Assertion Failure", "methodType.Proc.ResultCount == 1", "llvm_backend_objc.go", 0)
				}
				wrapperResultsTuple = alloc_type_tuple_from_field_types([]*Type{methodType.Proc.Results.Tuple.Variables[0].Type}, 1, false, true)
			}
			wrapperProcType := alloc_type_proc(nil, wrapperArgsTuple, wrapperArgsTuple.Tuple.Variables.Count,
				wrapperResultsTuple, int(methodType.Proc.ResultCount), false, ProcCC_CDecl)
			wrapperProc := lb_create_dummy_procedure(m, procName, wrapperProcType)
			lb_add_function_type_attributes(wrapperProc.Value, lb_get_function_type(m, wrapperProcType), ProcCC_CDecl)
			LLVMSetDLLStorageClass(wrapperProc.Value, LLVMDLLExportStorageClass)
			lb_add_attribute_to_proc(wrapperProc.Module, wrapperProc.Value, "nounwind")
			lb_begin_procedure_body(wrapperProc)

			{
				var contextAddr LLVMValueRef
				if methodType.Proc.CallingConvention == ProcCC_Odin {
					if !(contextProvider != nil) {
						gb_assert_handler("Assertion Failure", "contextProvider", "llvm_backend_objc.go", 0)
					}
					getContextArgs := []lbValue{{Value: wrapperProc.RawInputParameters[0], Type: contextProviderSelfPtrType}}
					if isContextProviderIvar {
						realSelf := lbValue{Value: wrapperProc.RawInputParameters[0], Type: classPtrType}
						getContextArgs[0] = lb_handle_objc_ivar_for_objc_object_pointer(wrapperProc, realSelf)
					}
					context := lb_emit_call(wrapperProc, contextProviderProcValue, getContextArgs)
					contextAddr = lb_address_from_load(wrapperProc, context).Value
				}

				methodForwardArgCount := methodParamCount + methodParamOffset
				methodForwardReturnArgOffset := isize(0)
				rawMethodArgs := make([]LLVMValueRef, 0, methodForwardArgCount+1)
				methodProcValue := lb_find_procedure_value_from_entity(m, md.Entity)
				ft := lb_get_function_type(m, methodType)
				hasReturn := false
				returnKind := lbArg_Direct
				if wrapperResultsTuple != nil {
					hasReturn = true
					returnKind = ft.Return.Kind
					if returnKind == lbArg_Indirect {
						methodForwardReturnArgOffset = 1
						rawMethodArgs = append(rawMethodArgs, wrapperProc.ReturnPtr.Addr.Value)
					}
				}
				if !md.Ac.ObjcIsClassMethod {
					rawMethodArgs = append(rawMethodArgs, wrapperProc.RawInputParameters[methodForwardReturnArgOffset])
				}
				for i := isize(0); i < methodParamCount; i++ {
					rawMethodArgs = append(rawMethodArgs, wrapperProc.RawInputParameters[i+2+methodForwardReturnArgOffset])
				}
				if methodType.Proc.CallingConvention == ProcCC_Odin {
					rawMethodArgs = append(rawMethodArgs, contextAddr)
				}
				fnp := lb_type_internal_for_procedures_raw(m, methodType)
				retValRaw := LLVMBuildCall2(wrapperProc.Builder, fnp, methodProcValue.Value, rawMethodArgs, uint(len(rawMethodArgs)), "")
				if hasReturn && returnKind != lbArg_Indirect {
					LLVMBuildRet(wrapperProc.Builder, retValRaw)
				} else {
					LLVMBuildRetVoid(wrapperProc.Builder)
				}
			}
			lb_end_procedure_body(wrapperProc)

			methodEncoding := String{Data: unsafe.SliceData(unsafe.StringData("v")), Len: 1}
			if !(methodType.Proc.ResultCount <= 1) {
				gb_assert_handler("Assertion Failure", "methodType.Proc.ResultCount <= 1", "llvm_backend_objc.go", 0)
			}
			if methodType.Proc.ResultCount != 0 {
				methodEncoding = lb_get_objc_type_encoding(methodType.Proc.Results.Tuple.Variables[0].Type)
			}
			if !md.Ac.ObjcIsClassMethod {
				atColon := String{Data: unsafe.SliceData(unsafe.StringData("@:")), Len: 2}
				methodEncoding = concatenate_strings(temporary_allocator(), methodEncoding, atColon)
			} else {
				hashColon := String{Data: unsafe.SliceData(unsafe.StringData("#:")), Len: 2}
				methodEncoding = concatenate_strings(temporary_allocator(), methodEncoding, hashColon)
			}
			for i := isize(0); i < methodParamCount; i++ {
				paramType := methodType.Proc.Params.Tuple.Variables[i+methodParamOffset].Type
				paramEncoding := lb_get_objc_type_encoding(paramType)
				methodEncoding = concatenate_strings(temporary_allocator(), methodEncoding, paramEncoding)
			}

			selAddr := string_map_get(&m.ObjCSelectors, md.Ac.ObjcSelector)
			if selAddr == nil {
				gb_assert_handler("Assertion Failure", "selAddr", "llvm_backend_objc.go", 0)
			}
			selectorValue := lb_addr_load(p, *selAddr)
			targetClass := classValue
			if md.Ac.ObjcIsClassMethod {
				targetClass = metaClassValue
			}
			addMethodArgs := []lbValue{targetClass, selectorValue, lbValue{Value: wrapperProc.Value, Type: wrapperProc.Type}, lb_const_value(m, t_cstring, exact_value_string(methodEncoding))}
			lb_emit_runtime_call(p, "class_addMethod", addMethodArgs)
		}

		if ivarType != nil {
			ivarBase := ivarType.Named.Base
			size := type_size_of(ivarBase)
			alignment := isize(floor_log2(uint64(type_align_of(ivarBase))))
			ivarName := String{Data: unsafe.SliceData(unsafe.StringData("__$ivar")), Len: isize(len("__$ivar"))}
			ivarTypes := String{Data: unsafe.SliceData(unsafe.StringData("{= }")), Len: isize(len("{= }"))}
			ivarArgs := []lbValue{
				classValue,
				lb_const_value(m, t_cstring, exact_value_string(ivarName)),
				lb_const_value(m, t_uint, exact_value_u64(uint64(size))),
				lb_const_value(m, t_u8, exact_value_u64(uint64(alignment))),
				lb_const_value(m, t_cstring, exact_value_string(ivarTypes)),
			}
			lb_emit_runtime_call(p, "class_addIvar", ivarArgs)
		}

		lb_emit_runtime_call(p, "objc_registerClassPair", []lbValue{classValue})
	}

	for _, g := range ivarMap {
		ivarAddr := lbAddr{}
		found := string_map_get(&m.Members, g.GlobalName)
		if found != nil {
			ivarAddr = lb_addr(*found)
			if !(ivarAddr.Addr.Type == t_int_ptr) {
				gb_assert_handler("Assertion Failure", "ivarAddr.Addr.Type == t_int_ptr", "llvm_backend_objc.go", 0)
			}
		} else {
			t := lb_type(m, t_int)
			global := lbValue{}
			global.Value = LLVMAddGlobal(m.Mod, t, unsafe.String(g.GlobalName, gb_string_length(g.GlobalName)))
			global.Type = t_int_ptr
			LLVMSetInitializer(global.Value, LLVMConstInt(t, 0, true))
			ivarAddr = lb_addr(global)
		}
		className := g.ClassImplType.Named.TypeName.TypeName.ObjcClassName
		classValue := globalClassMap[string(className.Text[:className.Len])].ClassValue
		ivarN := String{Data: unsafe.SliceData(unsafe.StringData("__$ivar")), Len: isize(len("__$ivar"))}
		ivar := lb_emit_runtime_call(p, "class_getInstanceVariable", []lbValue{classValue, lb_const_value(m, t_cstring, exact_value_string(ivarN))})
		ivarOffset := lb_emit_runtime_call(p, "ivar_getOffset", []lbValue{ivar})
		ivarOffsetInt := lb_emit_conv(p, ivarOffset, t_int)
		lb_addr_store(p, ivarAddr, ivarOffsetInt)
	}

	lb_end_procedure_body(p)
}
