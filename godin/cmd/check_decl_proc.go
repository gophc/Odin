package cmd

func init_entity_foreign_library(ctx *CheckerContext, e *Entity) *Entity {
	var ident *Ast
	var foreign_library **Entity

	switch e.Kind {
	case Entity_Procedure:
		ident = e.Procedure.ForeignLibraryIdent
		foreign_library = &e.Procedure.ForeignLibrary
	case Entity_Variable:
		ident = e.Variable.ForeignLibraryIdent
		foreign_library = &e.Variable.ForeignLibrary
	default:
		return nil
	}

	if ident == nil {
		error(e.Token, "foreign entities must declare which library they are from")
	} else if ident.Kind != Ast_Ident {
		error(ident, "foreign library names must be an identifier")
	} else {
		name := ident.Ident.Token.String
		found := scope_lookup(ctx.Scope, ident.Ident.Interned, ident.Ident.Hash)

		if found == nil {
			if is_blank_ident(name) {
			} else {
				error(ident, "Undeclared name: %.*s", int(name.Len), goStr(name))
			}
		} else if found.Kind != Entity_LibraryName {
			error(ident, "'%.*s' cannot be used as a library name", int(name.Len), goStr(name))
		} else {
			*foreign_library = found
			found.Flags |= EntityFlag_Used
			add_entity_use(ctx, ident, found)
			return found
		}
	}
	return nil
}

func check_objc_methods(ctx *CheckerContext, e *Entity, ac *AttributeContext) {
	if ac.ObjcType == nil {
		return
	}

	t := ac.ObjcType
	if t.Kind != Type_Named {
		return
	}

	if ac.ObjcName.Len == 0 {
		proc_name := goStr(e.Token.String)
		type_name := t.Named.Name

		if len(proc_name) > len(type_name)+1 &&
			proc_name[len(type_name)] == '_' &&
			type_name == proc_name[:len(type_name)] {
			ac.ObjcName = S(proc_name[len(type_name)+1:])
		} else {
			error(e.Token, "@(objc_name) requires that @(objc_type) be set or inferred by prefixing the proc name with the type and underscore: MyObjcType_myProcName :: proc().")
		}
	}

	tn := t.Named.TypeName
	if tn.Kind != Entity_TypeName {
		return
	}

	if tn.Scope != e.Scope {
		error(e.Token, "@(objc_name) attribute may only be applied to procedures and types within the same scope")
	} else {
		implement := tn.TypeName.ObjcIsImplementation
		if ac.ObjcIsImplementation && !tn.TypeName.ObjcIsImplementation {
			error(e.Token, "Cannot apply @(objc_is_implement) to a procedure whose type does not also have @(objc_is_implement) set")
		}
		if ac.ObjcIsDisabledImplement {
			implement = false
		}

		objc_selector := ac.ObjcSelector
		if objc_selector.Len == 0 {
			objc_selector = ac.ObjcName
		}

		if e.Kind == Entity_Procedure {
			has_body := e.DeclInfo.ProcLit.ProcLit.Body != nil
			e.Procedure.IsObjcImplOrImport = implement || !has_body
			e.Procedure.IsObjcClassMethod = ac.ObjcIsClassMethod
			e.Procedure.ObjcSelectorName = objc_selector
			e.Procedure.ObjcClass = tn

			pt := &e.Type.Proc
			var first_param *Type
			if pt.ParamCount > 0 {
				first_param = pt.Params.Tuple.Variables[0].Type
			} else {
				first_param = t_untyped_nil
			}

			if implement {
				if !has_body {
					error(e.Token, "Procedures with @(objc_is_implement) must have a body")
				} else if !tn.TypeName.ObjcIsImplementation {
					error(e.Token, "@(objc_is_implement) attribute may only be applied to procedures whose class also have @(objc_is_implement) applied")
				} else if !ac.ObjcIsClassMethod && !(first_param.Kind == Type_Pointer && internal_check_is_assignable_to(t, first_param.Pointer.Elem)) {
					error(e.Token, "Objective-C instance methods implementations require the first parameter to be a pointer to the class type set by @(objc_type)")
				} else if pt.CallingConvention == ProcCC_Odin && !tn.TypeName.ObjcContextProvider {
					error(e.Token, "Objective-C methods with Odin calling convention can only be used with classes that have @(objc_context_provider) set")
				} else if ac.ObjcIsClassMethod && pt.CallingConvention != ProcCC_CDecl {
					error(e.Token, "Objective-C class methods (objc_is_class_method=true) that have @objc_is_implementation can only use \"c\" calling convention")
				} else if pt.ResultCount > 1 {
					error(e.Token, "Objective-C method implementations may return at most 1 value")
				} else {
					if ac.IsExport {
						error(e.Token, "Explicit export not allowed when @(objc_implement) is set. It set exported implicitly")
					}
					if ac.LinkName.Len != 0 {
						error(e.Token, "Explicit linkage not allowed when @(objc_implement) is set. It set to \"strong\" implicitly")
					}
					ac.IsExport = true
					ac.Linkage = S("strong")

					method := ObjcMethodData{Ac: *ac, Entity: e}
					method.Ac.ObjcSelector = objc_selector

					info := ctx.Info
					mutex_lock(&info.ObjcMethodMutex)
					methodListPtr := map_get(&info.ObjcMethodImplementations, t)
					if methodListPtr != nil {
						methodList := *methodListPtr
						methodList = append(methodList, method)
						map_set(&info.ObjcMethodImplementations, t, methodList)
					} else {
						methodList := []ObjcMethodData{method}
						map_set(&info.ObjcMethodImplementations, t, methodList)
					}
					mutex_unlock(&info.ObjcMethodMutex)
				}
			} else if !has_body {
				if goStr(ac.ObjcSelector) == "The @(objc_selector) attribute is required for imported Objective-C methods." {
					return
				} else if pt.CallingConvention != ProcCC_CDecl {
					error(e.Token, "Imported Objective-C methods must use the \"c\" calling convention")
					return
				} else if tn.TypeName.ObjcContextProvider {
					error(e.Token, "Imported Objective-C class '%s' must not declare context providers.", t.Named.Name)
					return
				} else if tn.TypeName.ObjcIsImplementation {
					error(e.Token, "Imported Objective-C methods used in a class with @(objc_implement) is not allowed.")
					return
				} else if !ac.ObjcIsClassMethod && !(first_param.Kind == Type_Pointer && internal_check_is_assignable_to(t, first_param.Pointer.Elem)) {
					error(e.Token, "Objective-C instance methods require the first parameter to be a pointer to the class type set by @(objc_type)")
					return
				}
			} else if ac.ObjcSelector.Len != 0 {
				error(e.Token, "@(objc_selector) may only be applied to procedures that are Objective-C method implementations or are imported.")
				return
			}
		} else {
			if tn.TypeName.ObjcIsImplementation {
				error(e.Token, "Objective-C procedure groups cannot use the @(objc_implement) attribute.")
				return
			}
		}

		mutex_lock(&global_type_name_objc_metadata_mutex)
		if tn.TypeName.ObjcMetadata == nil {
			tn.TypeName.ObjcMetadata = create_type_name_obj_c_metadata()
		}
		md := tn.TypeName.ObjcMetadata
		mutex_lock(md.Mutex)

		nameKey := string_interner_insert(ac.ObjcName)
		if !ac.ObjcIsClassMethod {
			ok := true
			for _, entry := range md.ValueEntries {
				if entry.Interned == nameKey {
					error(e.Token, "Previous declaration of @(objc_name=\"%.*s\")", int(ac.ObjcName.Len), goStr(ac.ObjcName))
					ok = false
					break
				}
			}
			if ok {
				md.ValueEntries = append(md.ValueEntries, TypeNameObjCMetadataEntry{Interned: nameKey, Entity: e})
			}
		} else {
			ok := true
			for _, entry := range md.TypeEntries {
				if entry.Interned == nameKey {
					error(e.Token, "Previous declaration of @(objc_name=\"%.*s\")", int(ac.ObjcName.Len), goStr(ac.ObjcName))
					ok = false
					break
				}
			}
			if ok {
				md.TypeEntries = append(md.TypeEntries, TypeNameObjCMetadataEntry{Interned: nameKey, Entity: e})
			}
		}
		mutex_unlock(md.Mutex)
		mutex_unlock(&global_type_name_objc_metadata_mutex)
	}
}

func check_foreign_procedure(ctx *CheckerContext, e *Entity, d *DeclInfo) {
	if e.Kind != Entity_Procedure {
		return
	}
	name := e.Procedure.LinkName

	mutex_lock(&ctx.Info.ForeignMutex)

	fp := &ctx.Info.Foreigns
	key := string_hash_string(name)
	found := string_map_get(fp, key)
	if found != nil && e != *found {
		f := *found
		pos := f.Token.Pos
		this_type := base_type(e.Type)
		other_type := base_type(f.Type)
		if is_type_proc(this_type) && is_type_proc(other_type) {
			if !are_signatures_similar_enough(this_type, other_type) {
				error(d.ProcLit,
					"Redeclaration of foreign procedure '%.*s' with different type signatures\n\tat %s",
					int(name.Len), goStr(name), token_pos_to_string(pos))
			}
		} else if !signature_parameter_similar_enough(this_type, other_type) {
			error(d.ProcLit,
				"Foreign entity '%.*s' previously declared elsewhere with a different type\n\tat %s",
				int(name.Len), goStr(name), token_pos_to_string(pos))
		}
	} else if goStr(name) == "main" {
		error(d.ProcLit, "The link name 'main' is reserved for internal use")
	} else {
		string_map_set(fp, key, e)
	}

	mutex_unlock(&ctx.Info.ForeignMutex)
}

func check_proc_decl(ctx *CheckerContext, e *Entity, d *DeclInfo) {
	if e.Type != nil {
		return
	}
	if d.ProcLit.Kind != Ast_ProcLit {
		error(d.ProcLit, "Expected a procedure to check")
		return
	}

	proc_type := e.Type
	if d.GenProcType != nil {
		proc_type = d.GenProcType
	} else {
		proc_type = alloc_type_proc(e.Scope, nil, 0, nil, 0, false, default_calling_convention())
	}
	e.Type = proc_type
	pl := &d.ProcLit.ProcLit

	check_open_scope(ctx, pl.Type)
	defer check_close_scope(ctx)
	ctx.Scope.ProcedureEntity = e

	var decl_type *Type = nil

	if d.TypeExpr != nil {
		decl_type = check_type(ctx, d.TypeExpr)
		if !is_type_proc(decl_type) {
			str := type_to_string(decl_type)
			error(d.TypeExpr, "Expected a procedure type, got '%s'", str)
			gb_string_free(str)
		}
	}

	tmp_ctx := *ctx
	tmp_ctx.AllowPolymorphicTypes = true
	if decl_type != nil {
		tmp_ctx.TypeHint = decl_type
	}
	check_procedure_type(&tmp_ctx, proc_type, pl.Type, nil)

	if decl_type != nil {
		var x Operand
		x.Type = e.Type
		x.Mode = Addressing_Variable
		if !check_is_assignable_to(ctx, &x, decl_type) {
			expr_str := expr_to_string(d.ProcLit)
			op_type_str := type_to_string(e.Type)
			type_str := type_to_string(decl_type)
			error(e.Token,
				"Cannot assign '%s' of type '%s' to '%s'",
				expr_str,
				op_type_str,
				type_str)
			gb_string_free(type_str)
			gb_string_free(op_type_str)
			gb_string_free(expr_str)
		}
	}

	pt := &proc_type.Proc
	ac := make_attribute_context(e.Procedure.LinkPrefix, e.Procedure.LinkSuffix)

	if d != nil {
		check_decl_attributes(ctx, d.Attributes, proc_decl_attribute, &ac)
	}

	if ac.Test {
		e.Flags |= EntityFlag_Test
	}
	if ac.Init && ac.Fini {
		error(e.Token, "A procedure cannot be both declared as @(init) and @(fini)")
	} else if ac.Init {
		e.Flags |= EntityFlag_Init
	} else if ac.Fini {
		e.Flags |= EntityFlag_Fini
	}

	if ac.SetCold {
		e.Flags |= EntityFlag_Cold
	}
	e.Procedure.OptimizationMode = ac.OptimizationMode

	check_objc_methods(ctx, e, &ac)

	{
		if ac.RequireTargetFeature.Len != 0 && ac.EnableTargetFeature.Len != 0 {
			error(e.Token, "A procedure cannot have both @(require_target_feature=\"...\") and @(enable_target_feature=\"...\")")
		}

		if build_context.StrictTargetFeatures && ac.EnableTargetFeature.Len != 0 {
			ac.RequireTargetFeature = ac.EnableTargetFeature
			ac.EnableTargetFeature.Len = 0
		}

		if ac.RequireTargetFeature.Len != 0 {
			pt.RequireTargetFeature = goStr(ac.RequireTargetFeature)
			var invalid String
			if !check_target_feature_is_valid_globally(ac.RequireTargetFeature, &invalid) {
				error(e.Token, "Required target feature '%.*s' is not a valid target feature", int(invalid.Len), goStr(invalid))
			} else if !check_target_feature_is_enabled(ac.RequireTargetFeature, nil) {
				e.Flags |= EntityFlag_Disabled
			}
		} else if ac.EnableTargetFeature.Len != 0 {
			if is_arch_wasm() {
				error(e.Token, "@(enable_target_feature=\"...\") is not allowed on wasm, features for wasm must be declared globally")
			}
			pt.EnableTargetFeature = goStr(ac.EnableTargetFeature)
			var invalid String
			if !check_target_feature_is_valid_globally(ac.EnableTargetFeature, &invalid) {
				error(e.Token, "Procedure enabled target feature '%.*s' is not a valid target feature", int(invalid.Len), goStr(invalid))
			}
		}
	}

	switch e.Procedure.OptimizationMode {
	case ProcedureOptimizationMode_None:
		if pl.Inlining == ProcInlining_inline {
			error(e.Token, "#force_inline cannot be used in conjunction with the attribute 'optimization_mode' with neither \"none\" nor \"minimal\"")
		}
	}

	e.Procedure.EntryPointOnly = ac.EntryPointOnly
	e.Procedure.IsExport = ac.IsExport

	has_instrumentation := false
	if pl.Body == nil {
		has_instrumentation = false
		if ac.NoInstrumentation != Instrumentation_Default {
			error(e.Token, "@(no_instrumentation) is not allowed on foreign procedures")
		}
	} else {
		var file *AstFile
		if e.Token.Pos.FileID != 0 {
			file = threadUnsafeGetAstFileFromId(e.Token.Pos.FileID)
		}
		if file != nil {
			has_instrumentation = (file.Flags & AstFile_NoInstrumentation) == 0
		}

		switch ac.NoInstrumentation {
		case Instrumentation_Enabled:
			has_instrumentation = true
		case Instrumentation_Default:
		case Instrumentation_Disabled:
			has_instrumentation = false
		}
	}

	is_valid_instrumentation_call := func(type_ *Type) bool {
		if type_ == nil || type_.Kind != Type_Proc {
			return false
		}
		if type_.Proc.CallingConvention != ProcCC_Contextless {
			return false
		}
		if type_.Proc.ResultCount != 0 {
			return false
		}
		if type_.Proc.ParamCount != 3 {
			return false
		}
		p0 := type_.Proc.Params.Tuple.Variables[0].Type
		p1 := type_.Proc.Params.Tuple.Variables[1].Type
		p3 := type_.Proc.Params.Tuple.Variables[2].Type
		return is_type_rawptr(p0) && is_type_rawptr(p1) && are_types_identical(p3, t_source_code_location)
	}

	instrumentation_proc_type_str := "proc \"contextless\" (proc_address: rawptr, call_site_return_address: rawptr, loc: runtime.Source_Code_Location)"

	if ac.InstrumentationEnter && ac.InstrumentationExit {
		error(e.Token, "A procedure cannot be marked with both @(instrumentation_enter) and @(instrumentation_exit)")
		has_instrumentation = false
		e.Flags |= EntityFlag_Require
	} else if ac.InstrumentationEnter {
		init_core_source_code_location(ctx.Checker)
		if !is_valid_instrumentation_call(e.Type) {
			init_core_source_code_location(ctx.Checker)
			s := type_to_string(e.Type)
			error(e.Token, "@(instrumentation_enter) procedures must have the type '%s', got %s", instrumentation_proc_type_str, s)
			gb_string_free(s)
		}
		if (e.Scope.Flags & (ScopeFlag_File | ScopeFlag_Pkg)) == 0 {
			error(e.Token, "@(instrumentation_enter) procedures must be declared at the file scope")
		}
		mutex_lock(&ctx.Info.InstrumentationMutex)
		if ctx.Info.InstrumentationEnterEntity != nil {
			error(e.Token, "@(instrumentation_enter) has already been set")
		} else {
			ctx.Info.InstrumentationEnterEntity = e
		}
		mutex_unlock(&ctx.Info.InstrumentationMutex)
		has_instrumentation = false
		e.Flags |= EntityFlag_Require
	} else if ac.InstrumentationExit {
		init_core_source_code_location(ctx.Checker)
		if !is_valid_instrumentation_call(e.Type) {
			s := type_to_string(e.Type)
			error(e.Token, "@(instrumentation_exit) procedures must have the type '%s', got %s", instrumentation_proc_type_str, s)
			gb_string_free(s)
		}
		if (e.Scope.Flags & (ScopeFlag_File | ScopeFlag_Pkg)) == 0 {
			error(e.Token, "@(instrumentation_exit) procedures must be declared at the file scope")
		}
		mutex_lock(&ctx.Info.InstrumentationMutex)
		if ctx.Info.InstrumentationExitEntity != nil {
			error(e.Token, "@(instrumentation_exit) has already been set")
		} else {
			ctx.Info.InstrumentationExitEntity = e
		}
		mutex_unlock(&ctx.Info.InstrumentationMutex)
		has_instrumentation = false
		e.Flags |= EntityFlag_Require
	}

	e.Procedure.HasInstrumentation = has_instrumentation

	e.Procedure.NoSanitizeAddress = ac.NoSanitizeAddress
	e.Procedure.NoSanitizeMemory = ac.NoSanitizeMemory
	e.Procedure.NoSanitizeThread = ac.NoSanitizeThread

	e.DeprecatedMessage = ac.DeprecatedMessage
	e.WarningMessage = ac.WarningMessage
	ac.LinkName = handle_link_name(ctx, e.Token, ac.LinkName, ac.LinkPrefix, ac.LinkSuffix)
	if ac.HasDisabledProc {
		if ac.DisabledProc {
			e.Flags |= EntityFlag_Disabled
		}
		t := base_type(e.Type)
		if t.Kind == Type_Proc && t.Proc.ResultCount != 0 {
			error(e.Token, "Procedure with the 'disabled' attribute may not have any return values")
		}
	}

	is_foreign := e.Procedure.IsForeign
	is_export := e.Procedure.IsExport

	if ac.Linkage.Len != 0 {
		switch goStr(ac.Linkage) {
		case "internal":
			e.Flags |= EntityFlag_CustomLinkage_Internal
		case "strong":
			e.Flags |= EntityFlag_CustomLinkage_Strong
		case "weak":
			e.Flags |= EntityFlag_CustomLinkage_Weak
		case "link_once":
			e.Flags |= EntityFlag_CustomLinkage_LinkOnce
		}

		if is_foreign && (e.Flags&EntityFlag_CustomLinkage_Internal) != 0 {
			error(e.Token, "A foreign procedure may not have an \"internal\" linkage")
		}
	}

	if ac.RequireDeclaration {
		e.Flags |= EntityFlag_Require
		pl.Inlining = ProcInlining_no_inline
	}

	if e.Pkg != nil && goStr(e.Token.String) == "main" && !build_context.NoEntryPoint {
		if e.Pkg.Kind != Package_Runtime {
			if pt.ParamCount != 0 ||
				pt.ResultCount != 0 {
				str := type_to_string(proc_type)
				error(e.Token, "Procedure type of 'main' was expected to be 'proc()', got %s", str)
				gb_string_free(str)
			}
			if pt.CallingConvention != default_calling_convention() {
				error(e.Token, "Procedure 'main' cannot have a custom calling convention")
			}
			pt.CallingConvention = default_calling_convention()
			if e.Pkg.Kind == Package_Init {
				if ctx.Info.EntryPoint != nil {
					error(e.Token, "Redeclaration of the entry pointer procedure 'main'")
				} else {
					ctx.Info.EntryPoint = e
				}
			}
		}
	}

	if is_foreign && is_export {
		error(pl.Type, "A foreign procedure cannot have an 'export' tag")
	}

	if pt.IsPolymorphic {
		if pl.Body == nil {
			error(e.Token, "Polymorphic procedures must have a body")
		}
		if is_foreign {
			error(e.Token, "A foreign procedure cannot be a polymorphic")
			return
		}
	}

	if pl.Body != nil {
		if is_foreign {
			error(pl.Body, "A foreign procedure cannot have a body")
		}

		d.Scope = ctx.Scope

		if pl.Body.Kind == Ast_BlockStmt && !pt.IsPolymorphic {
			check_procedure_later_full(ctx.Checker, ctx.File, e.Token, d, proc_type, pl.Body, pl.Tags)
		}
	} else if !is_foreign && !e.Procedure.IsObjcImplOrImport {
		if e.Procedure.IsExport {
			error(e.Token, "Foreign export procedures must have a body")
		} else {
			error(e.Token, "Only a foreign procedure cannot have a body")
		}
	}

	if ac.RequireResults {
		if pt.ResultCount == 0 {
			error(pl.Type, "'require_results' is not needed on a procedure with no results")
		} else {
			pt.RequireResults = true
		}
	} else if d.ForeignRequireResults && pt.ResultCount != 0 {
		pt.RequireResults = true
	}

	if ac.LinkName.Len > 0 {
		ln := ac.LinkName
		e.Procedure.LinkName = ln
		if goStr(ln) == "memcpy" ||
			goStr(ln) == "memmove" ||
			goStr(ln) == "mem_copy" ||
			goStr(ln) == "mem_copy_non_overlapping" {
			e.Procedure.IsMemcpyLike = true
		}
	}

	if ac.DeferredProcedure.Entity != nil {
		e.Procedure.DeferredProcedure = ac.DeferredProcedure
		mpsc_enqueue(&ctx.Checker.ProcsWithDeferredToCheck, e)
	}

	if is_foreign {
		name := e.Token.String
		if e.Procedure.LinkName.Len > 0 {
			name = e.Procedure.LinkName
		}
		foreign_library := init_entity_foreign_library(ctx, e)
		e.Procedure.IsForeign = true
		e.Procedure.LinkName = name
		e.Procedure.ForeignLibrary = foreign_library

		if is_arch_wasm() && foreign_library != nil {
			mpsc_enqueue(&ctx.Info.ForeignDeclsToCheck, e)
		} else if !e.Procedure.IsObjcImplOrImport {
			check_foreign_procedure(ctx, e, d)
		}
	} else {
		name := e.Token.String
		if e.Procedure.LinkName.Len > 0 {
			name = e.Procedure.LinkName
		}
		if e.Procedure.LinkName.Len > 0 || is_export {
			mutex_lock(&ctx.Info.ForeignMutex)

			fp := &ctx.Info.Foreigns
			key := string_hash_string(name)
			found := string_map_get(fp, key)
			if found != nil {
				f := *found
				pos := f.Token.Pos
				error(d.ProcLit,
					"Non unique linking name for procedure '%.*s'\n\tother at %s",
					int(name.Len), goStr(name), token_pos_to_string(pos))
			} else if goStr(name) == "main" {
				if e.Pkg.Kind != Package_Runtime {
					error(d.ProcLit, "The link name 'main' is reserved for internal use")
				}
			} else {
				string_map_set(fp, key, e)
			}

			mutex_unlock(&ctx.Info.ForeignMutex)
		}
	}

	if e.Procedure.LinkName.Len > 0 {
		e.Flags |= EntityFlag_CustomLinkName
	}
}
