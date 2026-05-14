package cmd

import "fmt"

func lb_create_procedure(m *lbModule, entity *Entity, ignore_body ...bool) *lbProcedure {
	gb_assert_handler("Assertion Failure", "entity != nil", "llvm_backend_proc_create.go", 0)
	gb_assert_handler("Assertion Failure", "entity.Kind == Entity_Procedure", "llvm_backend_proc_create.go", 0)

	ignoreBody := false
	if len(ignore_body) > 0 {
		ignoreBody = ignore_body[0]
	}

	if is_type_polymorphic(entity.Type) && !entity.Procedure.IsForeign {
		bt := base_type(entity.Type)
		if bt.Kind == Type_Proc && bt.Proc.IsPolymorphic && !bt.Proc.IsPolySpecialized {
			return nil
		}
	}
	if !entity.Procedure.IsForeign {
		if entity.Flags&EntityFlag_ProcBodyChecked == 0 {
			gb_assert_handler("Panic", "proc body not checked", "llvm_backend_proc_create.go", 0,
				"%.*s :: %s (was parapoly: %d %d)",
				len(entity.Token.String), goStr(entity.Token.String), type_to_string(entity.Type),
				bToInt(is_type_polymorphic(entity.Type, true)), bToInt(is_type_polymorphic(entity.Type, false)))
		}
	}

	var link_name string
	if ignoreBody {
		other_module := lb_module_of_entity(m.Gen, entity, m)
		link_name = lb_get_entity_name(other_module, entity)
	} else {
		link_name = lb_get_entity_name(m, entity)
	}

	{
		if found, ok := m.Members[link_name]; ok {
			lb_add_entity(m, entity, found)
			return m.Procedures[link_name]
		}
	}

	p := &lbProcedure{}
	p.Module = m
	entity.CodeGenModule.Store(m)
	entity.CodeGenProcedure = p
	p.Entity = entity
	p.Name = make_string_c(link_name)

	decl := entity.DeclInfo
	gb_assert_handler("Assertion Failure", "decl.ProcLit.Kind == Ast_ProcLit", "llvm_backend_proc_create.go", 0)
	pl := &decl.ProcLit.ProcLit
	pt := base_type(entity.Type)
	gb_assert_handler("Assertion Failure", "pt.Kind == Type_Proc", "llvm_backend_proc_create.go", 0)

	p.Type = entity.Type
	p.TypeExpr = decl.TypeExpr
	p.Body = pl.Body
	p.Inlining = pl.Inlining
	p.Tailing = pl.Tailing
	p.IsForeign = entity.Procedure.IsForeign
	p.IsExport = entity.Procedure.IsExport
	p.IsEntryPoint = false

	if p.Entity != nil && p.Entity.Procedure.UsesBranchLocation {
		p.UsesBranchLocation = true
	}

	if p.IsForeign {
		lb_add_foreign_library_path(p.Module, entity.Procedure.ForeignLibrary)
	}

	func_type := lb_get_procedure_raw_type(m, p.Type)
	p.Value = LLVMAddFunction(m.Mod, link_name, func_type)

	lb_ensure_abi_function_type(m, p)
	lb_add_function_type_attributes(p.Value, p.AbiFunctionType, p.AbiFunctionType.CallingConvention)

	if buildContext.DisableUnwind {
		lb_add_attribute_to_proc(m, p.Value, "nounwind")
	}
	if pt.Proc.Diverging {
		lb_add_attribute_to_proc(m, p.Value, "noreturn")
	}
	if pt.Proc.CallingConvention == ProcCC_Naked {
		lb_add_attribute_to_proc(m, p.Value, "naked")
	}
	if !entity.Procedure.IsForeign && buildContext.DisableRedZone {
		lb_add_attribute_to_proc(m, p.Value, "noredzone")
	}

	switch p.Inlining {
	case ProcInlining_inline:
		lb_add_attribute_to_proc(m, p.Value, "alwaysinline")
	case ProcInlining_no_inline:
		lb_add_attribute_to_proc(m, p.Value, "noinline")
	default:
		if buildContext.InternalNoInline {
			lb_add_attribute_to_proc(m, p.Value, "noinline")
		}
	}

	switch entity.Procedure.OptimizationMode {
	case ProcedureOptimizationMode_None:
		lb_add_attribute_to_proc(m, p.Value, "optnone")
		lb_add_attribute_to_proc(m, p.Value, "noinline")
	case ProcedureOptimizationMode_FavorSize:
		lb_add_attribute_to_proc(m, p.Value, "optsize")
	}

	if pt.Proc.EnableTargetFeature != "" {
		var feature_buf string
		it := String_Iterator{Text: make_string_c(pt.Proc.EnableTargetFeature), Pos: 0}
		first := true
		for {
			str := string_split_iterator(&it, ',')
			if str.Len == 0 {
				break
			}
			add_prefix := !(string_starts_with(str, S("+")) || string_starts_with(str, S("-")))
			if !first {
				feature_buf += ","
			}
			first = false
			if add_prefix {
				feature_buf += "+"
			}
			feature_buf += goStr(str)
		}
		lb_add_attribute_to_proc_with_string(m, p.Value, make_string_c("target-features"), make_string_c(feature_buf))
	}

	if entity.Flags&EntityFlag_Cold != 0 {
		lb_add_attribute_to_proc(m, p.Value, "cold")
	}

	if p.IsExport {
		LLVMSetLinkage(p.Value, LLVMDLLExportLinkage)
		LLVMSetDLLStorageClass(p.Value, LLVMDLLExportStorageClass)
		LLVMSetVisibility(p.Value, LLVMDefaultVisibility)
		lb_set_wasm_export_attributes(p.Value, p.Name)
	} else if !p.IsForeign {
		if USE_SEPARATE_MODULES {
			LLVMSetLinkage(p.Value, LLVMExternalLinkage)
		} else {
			LLVMSetLinkage(p.Value, LLVMInternalLinkage)
			if entity.Pkg != nil && entity.Pkg.Kind == Package_Runtime && p.Body != nil {
				gb_assert_handler("Assertion Failure", "entity.Kind == Entity_Procedure", "llvm_backend_proc_create.go", 0)
				ln := entity.Procedure.LinkName
				if entity.Flags&EntityFlag_CustomLinkName != 0 && ln.Len > 0 {
					if string_starts_with(ln, S("__")) {
						LLVMSetLinkage(p.Value, LLVMExternalLinkage)
					} else {
						LLVMSetLinkage(p.Value, LLVMInternalLinkage)
					}
				}
			}
		}
	}

	if p.IsForeign {
		lb_set_wasm_procedure_import_attributes(p.Value, entity, p.Name)
	}

	offset := isize(1)
	if pt.Proc.ReturnByPointer {
		offset = 2
	}

	parameter_index := isize(0)
	if pt.Proc.ParamCount != 0 {
		params := &pt.Proc.Params.Tuple
		for i := isize(0); i < isize(pt.Proc.ParamCount); i++ {
			e := params.Variables[i]
			if e.Kind != Entity_Variable {
				continue
			}
			if i+1 == isize(len(params.Variables)) && pt.Proc.CVararg {
				continue
			}
			if e.Flags&EntityFlag_NoAlias != 0 {
				lb_add_proc_attribute_at_index(p, offset+parameter_index, "noalias")
			}
			if e.Flags&EntityFlag_NoCapture != 0 {
				if is_type_internally_pointer_like(e.Type) {
					lb_add_nocapture_proc_attribute_at_index(p, offset+parameter_index)
				}
			}
			parameter_index++
		}
	}

	if ignoreBody {
		p.Body = nil
		LLVMSetLinkage(p.Value, LLVMExternalLinkage)
	}

	lb_set_linkage_from_entity_flags(p.Module, p.Value, entity.Flags)

	if buildContext.LTOKind != LTO_None {
		if entity.Flags&EntityFlag_Require != 0 {
			linkage := LLVMGetLinkage(p.Value)
			if linkage != LLVMInternalLinkage {
				lb_append_to_used(m, p.Value)
			}
		}
	}

	if m.DebugBuilder != 0 {
		bt := base_type(p.Type)
		line := uint(entity.Token.Pos.Line)

		var scope LLVMMetadataRef
		var file LLVMMetadataRef
		var typ LLVMMetadataRef
		scope = m.DebugCompileUnit
		typ = lb_debug_type_internal_proc(m, bt)

		ident := entity.Identifier.Load()
		if entity.File != nil {
			file = lb_get_llvm_metadata(m, entity.File)
			scope = file
		} else if ident != nil && ident.FileID != 0 {
			file = lb_get_llvm_metadata(m, ident.File())
			scope = file
		} else if entity.Scope != nil {
			file = lb_get_llvm_metadata(m, entity.Scope.File)
			scope = file
		}
		gb_assert_handler("Assertion Failure", "file != nil", "llvm_backend_proc_create.go", 0, "%.*s", len(entity.Token.String), goStr(entity.Token.String))

		is_local_to_unit := LLVMBool(0)
		is_definition := LLVMBool(0)
		if p.Body != nil {
			is_definition = 1
		}
		scope_line := line
		flags := LLVMDIFlagStaticMember
		is_optimized := LLVMBool(0)
		if bt.Proc.Diverging {
			flags |= LLVMDIFlagNoReturn
		}
		if p.Body == nil {
			flags |= LLVMDIFlagPrototyped
			is_optimized = 0
		}

		if p.Body != nil {
			debug_name := p.Name

			p.DebugInfo = LLVMDIBuilderCreateFunction(m.DebugBuilder, scope,
				goStr(debug_name), uint(len(goStr(debug_name))),
				goStr(p.Name), uint(len(goStr(p.Name))),
				file, line, typ,
				is_local_to_unit, is_definition,
				scope_line, LLVMDIFlags(flags), is_optimized)
			gb_assert_handler("Assertion Failure", "p.DebugInfo != nil", "llvm_backend_proc_create.go", 0)
			LLVMSetSubprogram(p.Value, p.DebugInfo)
			lb_set_llvm_metadata(m, p, p.DebugInfo)
		}
	}

	if p.Body != nil && entity.Pkg != nil && (entity.Pkg.Kind == Package_Normal || entity.Pkg.Kind == Package_Init) {
		if buildContext.SanitizerFlags&SanitizerFlag_Address != 0 && !entity.Procedure.NoSanitizeAddress {
			lb_add_attribute_to_proc(m, p.Value, "sanitize_address")
		}
		if buildContext.SanitizerFlags&SanitizerFlag_Memory != 0 && !entity.Procedure.NoSanitizeMemory {
			lb_add_attribute_to_proc(m, p.Value, "sanitize_memory")
		}
		if buildContext.SanitizerFlags&SanitizerFlag_Thread != 0 && !entity.Procedure.NoSanitizeThread {
			lb_add_attribute_to_proc(m, p.Value, "sanitize_thread")
		}
	}

	if p.Body != nil && entity.Procedure.HasInstrumentation {
		instrumentation_enter := m.Info.InstrumentationEnterEntity
		instrumentation_exit := m.Info.InstrumentationExitEntity
		if instrumentation_enter != nil && instrumentation_exit != nil {
			enter := lb_get_entity_name(m, instrumentation_enter)
			exit := lb_get_entity_name(m, instrumentation_exit)
			lb_add_attribute_to_proc_with_string(m, p.Value, make_string_c("instrument-function-entry"), make_string_c(enter))
			lb_add_attribute_to_proc_with_string(m, p.Value, make_string_c("instrument-function-exit"), make_string_c(exit))
		}
	}

	proc_value := lbValue{Value: p.Value, Type: p.Type}
	lb_add_entity(m, entity, proc_value)
	m.Members[link_name] = proc_value
	m.Procedures[link_name] = p

	return p
}

func lb_create_dummy_procedure(m *lbModule, link_name string, typ *Type) *lbProcedure {
	{
		_, ok := m.Members[link_name]
		gb_assert_handler("Assertion Failure", !ok, "llvm_backend_proc_create.go", 0, "failed to create dummy procedure for: %s", link_name)
	}

	p := &lbProcedure{}
	p.Module = m
	p.Name = make_string_c(link_name)

	p.Type = typ
	p.TypeExpr = nil
	p.Body = nil
	p.Tags = 0
	p.Inlining = ProcInlining_none
	p.Tailing = ProcTailing_none
	p.IsForeign = false
	p.IsExport = false
	p.IsEntryPoint = false

	p.TupleFixMap = make(PtrMap[LLVMValueRef, lbTupleFix])

	func_type := lb_get_procedure_raw_type(m, p.Type)
	p.Value = LLVMAddFunction(m.Mod, link_name, func_type)

	pt := p.Type
	cc_kind := lbCallingConvention_C
	if !is_arch_wasm() {
		cc_kind = lb_calling_convention_map[pt.Proc.CallingConvention]
	}
	LLVMSetFunctionCallConv(p.Value, cc_kind)
	proc_value := lbValue{Value: p.Value, Type: p.Type}
	m.Members[link_name] = proc_value
	m.Procedures[link_name] = p

	offset := isize(1)
	if pt.Proc.ReturnByPointer {
		lb_add_proc_attribute_at_index(p, 1, "sret")
		lb_add_proc_attribute_at_index(p, 1, "noalias")
		offset = 2
	}

	parameter_index := isize(0)
	if pt.Proc.CallingConvention == ProcCC_Odin {
		lb_add_proc_attribute_at_index(p, offset+parameter_index, "noalias")
		lb_add_proc_attribute_at_index(p, offset+parameter_index, "nonnull")
		lb_add_nocapture_proc_attribute_at_index(p, offset+parameter_index)
	}
	return p
}

func lb_build_nested_proc(p *lbProcedure, pd *AstProcLit, e *Entity) {
	gb_assert_handler("Assertion Failure", "pd.Body != nil", "llvm_backend_proc_create.go", 0)
	m := p.Module

	if e.MinDepCount.Load() == 0 {
		return
	}

	original_name := e.Token.String
	pd_name := original_name
	if e.Procedure.LinkName.Len > 0 {
		pd_name = e.Procedure.LinkName
	}

	guid := i32(len(p.Children))
	name_str := fmt.Sprintf("%s%s%s-%d", goStr(p.Name), ABI_PKG_NAME_SEPARATOR, goStr(pd_name), guid)
	name := make_string_c(name_str)

	e.Procedure.LinkName = name

	nested_proc := lb_create_procedure(p.Module, e)
	if nested_proc == nil {
		return
	}
	e.CodeGenProcedure = nested_proc

	value := lbValue{Value: nested_proc.Value, Type: nested_proc.Type}
	lb_add_entity(m, e, value)
	p.Children = append(p.Children, nested_proc)
	mpsc_enqueue(&m.ProceduresToGenerate, nested_proc)
}
