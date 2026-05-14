package cmd

const (
	ProcedureOptimizationMode_None     uint32 = 0
	ProcedureOptimizationMode_FavorSize uint32 = 1
)

func check_decl_attribute_value(c *CheckerContext, value *Ast) ExactValue {
	ev := ExactValue{}
	if value != nil {
		op := Operand{}
		check_expr(c, &op, value)
		if op.Mode != Addressing_Invalid {
			if op.Mode == Addressing_Constant {
				ev = op.Value
			} else {
				error(value, "Expected a constant attribute element")
			}
		}
	}
	return ev
}

func foreign_block_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	ev := check_decl_attribute_value(c, value)
	if name == "tag" {
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "default_calling_convention" {
		if ev.Kind == ExactValue_String {
			cc := string_to_calling_convention(goStr(ev.ValueString))
			if cc == ProcCC_Invalid {
				error(elem, "Unknown procedure calling convention: '%.*s'", int(ev.ValueString.Len), goStr(ev.ValueString))
			} else {
				c.ForeignContext.DefaultCC = cc
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_prefix" {
		if ev.Kind == ExactValue_String {
			link_prefix := string_trim_whitespace(ev.ValueString)
			if link_prefix.Len != 0 && !is_foreign_name_valid(link_prefix) {
				error(elem, "Invalid link prefix: '%.*s'", int(link_prefix.Len), goStr(link_prefix))
			} else {
				c.ForeignContext.LinkPrefix = link_prefix
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_suffix" {
		if ev.Kind == ExactValue_String {
			link_suffix := string_trim_whitespace(ev.ValueString)
			if link_suffix.Len != 0 && !is_foreign_name_valid(link_suffix) {
				error(elem, "Invalid link suffix: '%.*s'", int(link_suffix.Len), goStr(link_suffix))
			} else {
				c.ForeignContext.LinkSuffix = link_suffix
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "private" {
		kind := EntityVisiblity_PrivateToPackage
		if ev.Kind == ExactValue_Invalid {
		} else if ev.Kind == ExactValue_String {
			v := ev.ValueString
			if v == "file" {
				kind = EntityVisiblity_PrivateToFile
			} else if v == "package" {
				kind = EntityVisiblity_PrivateToPackage
			} else {
				error(value, "'%.*s' expects no parameter, or a string literal containing \"file\" or \"package\"", int(name.Len), goStr(name))
			}
		} else {
			error(value, "'%.*s' expects no parameter, or a string literal containing \"file\" or \"package\"", int(name.Len), goStr(name))
		}
		c.ForeignContext.VisibilityKind = kind
		return true
	} else if name == "require_results" {
		if value != nil {
			error(elem, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		c.ForeignContext.RequireResults = true
		return true
	}
	return false
}

func proc_group_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_name" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			if string_is_valid_identifier(ev.ValueString) {
				ac.ObjcName = ev.ValueString
			} else {
				error(elem, "Invalid identifier for '%.*s', got '%.*s'", int(name.Len), goStr(name), int(ev.ValueString.Len), goStr(ev.ValueString))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_is_class_method" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Bool {
			ac.ObjcIsClassMethod = ev.ValueBool
		} else {
			error(elem, "Expected a boolean value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_type" {
		if value == nil {
			error(elem, "Expected a type for '%.*s'", int(name.Len), goStr(name))
		} else {
			objc_type := check_type(c, value)
			if objc_type != nil {
				if !has_type_got_objc_class_attribute(objc_type) {
					t := type_to_string(objc_type)
					error(value, "'%.*s' expected a named type with the attribute @(obj_class=<string>), got type %s", int(name.Len), goStr(name), t)
					gb_string_free(t)
				} else {
					ac.ObjcType = objc_type
				}
			}
		}
		return true
	} else if name == "require_results" {
		if value != nil {
			error(elem, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		ac.RequireResults = true
		return true
	}
	return false
}

func proc_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "test" {
		if value != nil {
			error(value, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		ac.Test = true
		return true
	} else if name == "export" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Invalid {
			ac.IsExport = true
		} else if ev.Kind == ExactValue_Bool {
			ac.IsExport = ev.ValueBool
		} else {
			error(value, "Expected either a boolean or no parameter for 'export'")
			return false
		}
		return true
	} else if name == "linkage" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(value, "Expected either a string 'linkage'")
			return false
		}
		linkage := ev.ValueString
		if linkage == "internal" ||
			linkage == "strong" ||
			linkage == "weak" ||
			linkage == "link_once" {
			ac.Linkage = linkage
		} else {
			begin_error_block()
			error(elem, "Invalid linkage '%.*s'. Valid kinds:", int(linkage.Len), goStr(linkage))
			error_line("\tinternal\n")
			error_line("\tstrong\n")
			error_line("\tweak\n")
			error_line("\tlink_once\n")
			end_error_block()
		}
		return true
	} else if name == "require" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Invalid {
			ac.RequireDeclaration = true
		} else if ev.Kind == ExactValue_Bool {
			ac.RequireDeclaration = ev.ValueBool
		} else {
			error(value, "Expected either a boolean or no parameter for 'require'")
		}
		return true
	} else if name == "init" {
		if value != nil {
			error(value, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		ac.Init = true
		return true
	} else if name == "fini" {
		if value != nil {
			error(value, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		ac.Fini = true
		return true
	} else if name == "deferred" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				error(elem, "'%.*s' is not allowed any more, please use one of the following instead: 'deferred_none', 'deferred_in', 'deferred_out'", int(name.Len), goStr(name))
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_out
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_none" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				ac.DeferredProcedure.Kind = DeferredProcedure_none
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_in" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_in
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_out" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_out
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_in_out" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_in_out
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_in_by_ptr" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_in_by_ptr
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_out_by_ptr" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_out_by_ptr
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "deferred_in_out_by_ptr" {
		if value != nil {
			o := Operand{}
			check_expr(c, &o, value)
			e := entity_of_node(o.Expr)
			if e != nil && e.Kind == Entity_Procedure {
				if ac.DeferredProcedure.Entity != nil {
					error(elem, "Previous usage of a 'deferred_*' attribute")
				}
				ac.DeferredProcedure.Kind = DeferredProcedure_in_out_by_ptr
				ac.DeferredProcedure.Entity = e
				return true
			}
		}
		error(elem, "Expected a procedure entity for '%.*s'", int(name.Len), goStr(name))
		return false
	} else if name == "link_name" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkName = ev.ValueString
			if !is_foreign_name_valid(ac.LinkName) {
				error(elem, "Invalid link name: %.*s", int(ac.LinkName.Len), goStr(ac.LinkName))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_prefix" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkPrefix = ev.ValueString
			if ac.LinkPrefix.Len != 0 && !is_foreign_name_valid(ac.LinkPrefix) {
				error(elem, "Invalid link prefix: %.*s", int(ac.LinkPrefix.Len), goStr(ac.LinkPrefix))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_suffix" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkSuffix = ev.ValueString
			if ac.LinkSuffix.Len != 0 && !is_foreign_name_valid(ac.LinkSuffix) {
				error(elem, "Invalid link suffix: %.*s", int(ac.LinkSuffix.Len), goStr(ac.LinkSuffix))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "deprecated" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			msg := ev.ValueString
			if msg.Len == 0 {
				error(elem, "Deprecation message cannot be an empty string")
			} else {
				ac.DeprecatedMessage = msg
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "require_results" {
		if value != nil {
			error(elem, "Expected no value for '%.*s'", int(name.Len), goStr(name))
		}
		ac.RequireResults = true
		return true
	} else if name == "disabled" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Bool {
			ac.HasDisabledProc = true
			ac.DisabledProc = ev.ValueBool
		} else {
			error(elem, "Expected a boolean value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "cold" {
		if value == nil {
			ac.SetCold = true
		} else {
			ev := check_decl_attribute_value(c, value)
			if ev.Kind == ExactValue_Bool {
				ac.SetCold = ev.ValueBool
			} else {
				error(elem, "Expected a boolean value for '%.*s' or no value whatsoever", int(name.Len), goStr(name))
			}
		}
		return true
	} else if name == "optimization_mode" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			mode := ev.ValueString
			if mode == "none" {
				ac.OptimizationMode = ProcedureOptimizationMode_None
			} else if mode == "favor_size" {
				ac.OptimizationMode = ProcedureOptimizationMode_FavorSize
			} else if mode == "minimal" {
				error(elem, "Invalid optimization_mode 'minimal' for '%.*s', mode has been removed due to confusion, but 'none' has the same behaviour", int(name.Len), goStr(name))
			} else if mode == "size" {
				error(elem, "Invalid optimization_mode 'size' for '%.*s', mode has been removed due to confusion, but 'favor_size' has the same behaviour", int(name.Len), goStr(name))
			} else if mode == "speed" {
				error(elem, "Invalid optimization_mode 'speed' for '%.*s', mode has been removed due to confusion, but 'favor_size' has the same behaviour", int(name.Len), goStr(name))
			} else {
				begin_error_block()
				error(elem, "Invalid optimization_mode for '%.*s'. Valid modes:", int(name.Len), goStr(name))
				error_line("\tnone\n")
				error_line("\tfavor_size\n")
				end_error_block()
			}
		} else {
			error(elem, "Expected a string for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_name" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			if string_is_valid_identifier(ev.ValueString) {
				ac.ObjcName = ev.ValueString
			} else {
				error(elem, "Invalid identifier for '%.*s', got '%.*s'", int(name.Len), goStr(name), int(ev.ValueString.Len), goStr(ev.ValueString))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_is_class_method" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Bool {
			ac.ObjcIsClassMethod = ev.ValueBool
		} else {
			error(elem, "Expected a boolean value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_type" {
		if value == nil {
			error(elem, "Expected a type for '%.*s'", int(name.Len), goStr(name))
		} else {
			objc_type := check_type(c, value)
			if objc_type != nil {
				if !has_type_got_objc_class_attribute(objc_type) {
					t := type_to_string(objc_type)
					error(value, "'%.*s' expected a named type with the attribute @(obj_class=<string>), got type %s", int(name.Len), goStr(name), t)
					gb_string_free(t)
				} else {
					ac.ObjcType = objc_type
				}
			}
		}
		return true
	} else if name == "objc_implement" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Bool {
			ac.ObjcIsImplementation = ev.ValueBool
			if !ac.ObjcIsImplementation {
				ac.ObjcIsDisabledImplement = true
			}
		} else if ev.Kind == ExactValue_Invalid {
			ac.ObjcIsImplementation = true
		} else {
			error(elem, "Expected a boolean value, or no value, for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_selector" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			if string_is_valid_identifier(ev.ValueString) {
				ac.ObjcSelector = ev.ValueString
			} else {
				error(elem, "Invalid identifier for '%.*s', got '%.*s'", int(name.Len), goStr(name), int(ev.ValueString.Len), goStr(ev.ValueString))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "require_target_feature" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.RequireTargetFeature = ev.ValueString
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "enable_target_feature" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.EnableTargetFeature = ev.ValueString
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "entry_point_only" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.EntryPointOnly = true
		return true
	} else if name == "no_instrumentation" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Invalid {
			ac.NoInstrumentation = Instrumentation_Disabled
		} else if ev.Kind == ExactValue_Bool {
			if ev.ValueBool {
				ac.NoInstrumentation = Instrumentation_Disabled
			} else {
				ac.NoInstrumentation = Instrumentation_Enabled
			}
		} else {
			error(value, "Expected either a boolean or no parameter for '%.*s'", int(name.Len), goStr(name))
			return false
		}
		return true
	} else if name == "instrumentation_enter" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.InstrumentationEnter = true
		return true
	} else if name == "instrumentation_exit" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.InstrumentationExit = true
		return true
	} else if name == "no_sanitize_address" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.NoSanitizeAddress = true
		return true
	} else if name == "no_sanitize_memory" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.NoSanitizeMemory = true
		return true
	} else if name == "no_sanitize_thread" {
		if value != nil {
			error(value, "'%.*s' expects no parameter", int(name.Len), goStr(name))
		}
		ac.NoSanitizeThread = true
		return true
	}
	return false
}

func var_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "static" {
		if value != nil {
			error(elem, "'static' does not have any parameters")
		}
		ac.IsStatic = true
		return true
	} else if name == "rodata" {
		if value != nil {
			error(elem, "'rodata' does not have any parameters")
		}
		ac.Rodata = true
		return true
	} else if name == "thread_local" {
		ev := check_decl_attribute_value(c, value)
		if ac.InitExprListCount > 0 {
			error(elem, "A thread local variable declaration cannot have initialization values")
		} else if c.ForeignContext.CurrLibrary != nil {
			error(elem, "A foreign block variable cannot be thread local")
		} else if ac.IsExport {
			error(elem, "An exported variable cannot be thread local")
		} else if ev.Kind == ExactValue_Invalid {
			ac.ThreadLocalModel = S("default")
		} else if ev.Kind == ExactValue_String {
			model := ev.ValueString
			if model == "default" ||
				model == "globaldynamic" ||
				model == "localdynamic" ||
				model == "initialexec" ||
				model == "localexec" {
				ac.ThreadLocalModel = model
			} else {
				begin_error_block()
				error(elem, "Invalid thread local model '%.*s'. Valid models:", int(model.Len), goStr(model))
				error_line("\tdefault\n")
				error_line("\tglobaldynamic\n")
				error_line("\tlocaldynamic\n")
				error_line("\tinitialexec\n")
				error_line("\tlocalexec\n")
				end_error_block()
			}
		} else {
			error(elem, "Expected either no value or a string for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	}
	if c.CurrProcDecl != nil {
		error(elem, "Only a variable at file scope can have a '%.*s'", int(name.Len), goStr(name))
		return true
	}
	if name == "require" {
		if value != nil {
			error(elem, "'require' does not have any parameters")
		}
		ac.RequireDeclaration = true
		return true
	} else if name == "export" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Invalid {
			ac.IsExport = true
		} else if ev.Kind == ExactValue_Bool {
			ac.IsExport = ev.ValueBool
		} else {
			error(value, "Expected either a boolean or no parameter for 'export'")
			return false
		}
		if ac.ThreadLocalModel.Len != 0 {
			error(elem, "An exported variable cannot be thread local")
		}
		return true
	} else if name == "linkage" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(value, "Expected either a string 'linkage'")
			return false
		}
		linkage := ev.ValueString
		if linkage == "internal" ||
			linkage == "strong" ||
			linkage == "weak" ||
			linkage == "link_once" {
			ac.Linkage = linkage
		} else {
			begin_error_block()
			error(elem, "Invalid linkage '%.*s'. Valid kinds:", int(linkage.Len), goStr(linkage))
			error_line("\tinternal\n")
			error_line("\tstrong\n")
			error_line("\tweak\n")
			error_line("\tlink_once\n")
			end_error_block()
		}
		return true
	} else if name == "link_name" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkName = ev.ValueString
			if !is_foreign_name_valid(ac.LinkName) {
				error(elem, "Invalid link name: %.*s", int(ac.LinkName.Len), goStr(ac.LinkName))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_prefix" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkPrefix = ev.ValueString
			if ac.LinkPrefix.Len != 0 && !is_foreign_name_valid(ac.LinkPrefix) {
				error(elem, "Invalid link prefix: %.*s", int(ac.LinkPrefix.Len), goStr(ac.LinkPrefix))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_suffix" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkSuffix = ev.ValueString
			if ac.LinkSuffix.Len != 0 && !is_foreign_name_valid(ac.LinkSuffix) {
				error(elem, "Invalid link suffix: %.*s", int(ac.LinkSuffix.Len), goStr(ac.LinkSuffix))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "link_section" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			ac.LinkSection = ev.ValueString
			if !is_foreign_name_valid(ac.LinkSection) {
				error(elem, "Invalid link section: %.*s", int(ac.LinkSection.Len), goStr(ac.LinkSection))
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	}
	return false
}

func const_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "private" {
		return true
	} else if name == "static" ||
		name == "thread_local" ||
		name == "require" ||
		name == "linkage" ||
		name == "link_name" ||
		name == "link_prefix" ||
		name == "link_suffix" ||
		name == "rodata" {
		error(elem, "@(%.*s) is not supported for compile time constant value declarations", int(name.Len), goStr(name))
		return true
	}
	return false
}

func type_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "private" {
		return true
	} else if name == "objc_class" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String || ev.ValueString.Len == 0 {
			error(elem, "Expected a non-empty string value for '%.*s'", int(name.Len), goStr(name))
		} else {
			ac.ObjcClass = ev.ValueString
		}
		return true
	} else if name == "objc_implement" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Bool {
			ac.ObjcIsImplementation = ev.ValueBool
		} else if ev.Kind == ExactValue_Invalid {
			ac.ObjcIsImplementation = true
		} else {
			error(elem, "Expected a boolean value, or no value, for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_superclass" {
		objc_superclass := check_type(c, value)
		if objc_superclass != nil {
			ac.ObjcSuperclass = objc_superclass
		} else {
			error(value, "'%.*s' expected a named type", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_ivar" {
		objc_ivar := check_type(c, value)
		if objc_ivar != nil && objc_ivar.Kind == Type_Named {
			ac.ObjcIvar = objc_ivar
		} else {
			error(value, "'%.*s' expected a named type", int(name.Len), goStr(name))
		}
		return true
	} else if name == "objc_context_provider" {
		o := Operand{}
		check_expr(c, &o, value)
		e := entity_of_node(o.Expr)
		if e != nil {
			if ac.ObjcContextProvider != nil {
				error(elem, "Previous usage of a 'objc_context_provider' attribute")
			}
			if e.Kind != Entity_Procedure {
				error(elem, "'objc_context_provider' must refer to a procedure")
			} else {
				ac.ObjcContextProvider = e
			}
			return true
		}
	} else if name == "raddbg_type_view" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_Invalid {
			ac.RaddbgTypeView = true
		} else if ev.Kind == ExactValue_String {
			ac.RaddbgTypeView = true
			ac.RaddbgTypeViewString = ev.ValueString
			if ev.ValueString.Len == 0 {
				error(elem, "Expected a non-empty string for '%.*s'", int(name.Len), goStr(name))
			}
		} else {
			error(elem, "Expected a string or no value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	} else if name == "deprecated" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind == ExactValue_String {
			msg := ev.ValueString
			if msg.Len == 0 {
				error(elem, "Deprecation message cannot be an empty string")
			} else {
				ac.DeprecatedMessage = msg
			}
		} else {
			error(elem, "Expected a string value for '%.*s'", int(name.Len), goStr(name))
		}
		return true
	}
	return false
}

func check_decl_attributes(c *CheckerContext, attributes []*Ast, proc DeclAttributeProc, ac *AttributeContext) {
	if len(attributes) == 0 {
		return
	}
	original_link_prefix := String{}
	original_link_suffix := String{}
	if ac != nil {
		original_link_prefix = ac.LinkPrefix
		original_link_suffix = ac.LinkSuffix
	}
	set := make(map[string]struct{})
	is_runtime := false
	if c.Scope != nil && c.Scope.File != nil && (c.Scope.Flags&int32(ScopeFlag_File)) != 0 &&
		c.Scope.File.Pkg != nil &&
		c.Scope.File.Pkg.Kind == PackageRuntime {
		is_runtime = true
	} else if c.Scope != nil && c.Scope.Parent != nil &&
		(c.Scope.Flags&int32(ScopeFlag_Proc)) != 0 &&
		(c.Scope.Parent.Flags&int32(ScopeFlag_File)) != 0 &&
		c.Scope.Parent.File != nil && c.Scope.Parent.File.Pkg != nil &&
		c.Scope.Parent.File.Pkg.Kind == PackageRuntime {
		is_runtime = true
	}
	for i := 0; i < len(attributes); i++ {
		attr := attributes[i]
		if attr.Kind != Ast_Attribute {
			continue
		}
		for j := 0; j < len(attr.Attribute.Elems); j++ {
			elem := attr.Attribute.Elems[j]
			var name String
			var value *Ast
			switch elem.Kind {
			case Ast_Ident:
				name = elem.Ident.Token.String
			case Ast_Implicit:
				name = elem.Implicit.String
			case Ast_FieldValue:
				fv := &elem.FieldValue
				if fv.Field.Kind == Ast_Ident {
					name = fv.Field.Ident.Token.String
				} else if fv.Field.Kind == Ast_Implicit {
					name = fv.Field.Implicit.String
				} else {
					gb_assert_handler("Panic", "0", "checker_attr.go", 0, "Unknown Field Value name")
				}
				value = fv.Value
			default:
				error(elem, "Invalid attribute element")
				continue
			}
			key := goStr(name)
			if _, exists := set[key]; exists {
				error(elem, "Previous declaration of '%.*s'", int(name.Len), goStr(name))
				continue
			}
			set[key] = struct{}{}
			if name == "builtin" && is_runtime {
				continue
			}
			if !proc(c, elem, name, value, ac) {
				if !buildContext.IgnoreUnknownAttributes {
					if _, custExists := buildContext.CustomAttributes[key]; !custExists {
						begin_error_block()
						error(elem, "Unknown attribute element name '%.*s'", int(name.Len), goStr(name))
						error_line("\tDid you forget to use the build flag '-ignore-unknown-attributes' or '-custom-attribute:%.*s'?\n", int(name.Len), goStr(name))
						end_error_block()
					}
				}
			}
		}
	}
	if ac != nil {
		if ac.LinkPrefix.Data == original_link_prefix.Data && ac.LinkPrefix.Len == original_link_prefix.Len {
			if ac.LinkName.Len > 0 {
				ac.LinkPrefix = String{}
			}
		}
		if ac.LinkSuffix.Data == original_link_suffix.Data && ac.LinkSuffix.Len == original_link_suffix.Len {
			if ac.LinkName.Len > 0 {
				ac.LinkSuffix = String{}
			}
		}
	}
}

func get_total_value_count(values []*Ast) isize {
	count := isize(0)
	for i := 0; i < len(values); i++ {
		t := type_of_expr(values[i])
		if t == nil {
			count += 1
			continue
		}
		t = core_type(t)
		if t.Kind == Type_Tuple {
			count += isize(len(t.Tuple.Variables))
		} else {
			count += 1
		}
	}
	return count
}

func check_arity_match(c *CheckerContext, vd *Ast, is_global bool) bool {
	lhs := isize(len(vd.ValueDecl.Names))
	rhs := isize(0)
	if is_global {
		rhs = isize(len(vd.ValueDecl.Values))
	} else {
		rhs = get_total_value_count(vd.ValueDecl.Values)
	}
	if rhs == 0 {
		if vd.ValueDecl.Type == nil {
			error(vd.ValueDecl.Names[0], "Missing type or initial expression")
			return false
		}
	} else if lhs < rhs {
		if lhs < isize(len(vd.ValueDecl.Values)) {
			n := vd.ValueDecl.Values[lhs]
			str := expr_to_string(n)
			error(n, "Extra initial expression '%s'", str)
			gb_string_free(str)
		} else {
			error(vd.ValueDecl.Names[0], "Extra initial expression")
		}
		return false
	} else if lhs > rhs {
		if !is_global && rhs != 1 {
			n := vd.ValueDecl.Names[rhs]
			str := expr_to_string(n)
			error(n, "Missing expression for '%s'", str)
			gb_string_free(str)
			return false
		} else if is_global {
			begin_error_block()
			n := vd.ValueDecl.Values[rhs-1]
			error(n, "Expected %d expressions on the right hand side, got %d", lhs, rhs)
			error_line("Note: Global declarations do not allow for multi-valued expressions")
			end_error_block()
			return false
		}
	}
	return true
}

func check_collect_entities_from_when_stmt(c *CheckerContext, ws *Ast) {
	operand := Operand{Mode: Addressing_Invalid}
	ws2 := &ws.WhenStmt
	if !ws2.IsConditionDetermined {
		check_expr(c, &operand, ws2.Cond)
		if operand.Mode != Addressing_Invalid && !is_type_boolean(operand.Type) {
			error(ws2.Cond, "Non-boolean condition in 'when' statement")
		}
		if operand.Mode != Addressing_Constant {
			error(ws2.Cond, "Non-constant condition in 'when' statement")
		}
		ws2.IsConditionDetermined = true
		ws2.DeterminedCond = operand.Value.Kind == ExactValue_Bool && operand.Value.ValueBool
	}
	if ws2.Body == nil || ws2.Body.Kind != Ast_BlockStmt {
		error(ws2.Cond, "Invalid body for 'when' statement")
	} else {
		if ws2.DeterminedCond {
			CheckCollectEntities(c, ws2.Body.BlockStmt.Stmts)
		} else if ws2.ElseStmt != nil {
			switch ws2.ElseStmt.Kind {
			case Ast_BlockStmt:
				CheckCollectEntities(c, ws2.ElseStmt.BlockStmt.Stmts)
			case Ast_WhenStmt:
				check_collect_entities_from_when_stmt(c, ws2.ElseStmt)
			default:
				error(ws2.ElseStmt, "Invalid 'else' statement in 'when' statement")
			}
		}
	}
}

func check_builtin_attributes(ctx *CheckerContext, e *Entity, attributes *[]*Ast) {
	switch e.Kind {
	case Entity_ProcGroup, Entity_Procedure, Entity_TypeName, Entity_Constant:
	default:
		return
	}
	if ctx.Scope.Flags&int32(ScopeFlag_File) == 0 || ctx.Scope.File == nil || ctx.Scope.File.Pkg == nil || ctx.Scope.File.Pkg.Kind != PackageRuntime {
		return
	}
	for j := 0; j < len(*attributes); j++ {
		attr := (*attributes)[j]
		if attr.Kind != Ast_Attribute {
			continue
		}
		for k := 0; k < len(attr.Attribute.Elems); k++ {
			elem := attr.Attribute.Elems[k]
			var name String
			var value *Ast
			switch elem.Kind {
			case Ast_Ident:
				name = elem.Ident.Token.String
			case Ast_FieldValue:
				fv := &elem.FieldValue
				if fv.Field.Kind != Ast_Ident {
					continue
				}
				name = fv.Field.Ident.Token.String
				value = fv.Value
			default:
				continue
			}
			if name == "builtin" {
				mutex_lock(&ctx.Info.BuiltinMutex)
				add_entity(ctx, builtinPkg.Scope, nil, e)
				interned := entity_interned_name(e)
				if ScopeLookup(builtinPkg.Scope, interned, interned.Hash()) == nil {
					gb_assert_handler("Assertion Failure", "scope_lookup(builtinPkg.Scope, interned, hash) != nil", "checker_attr.go", 0, "")
				}
				if value != nil {
					error(value, "'builtin' cannot have a field value")
				}
				mutex_unlock(&ctx.Info.BuiltinMutex)
			}
		}
	}
	for i := 0; i < len(*attributes); i++ {
		attr := (*attributes)[i]
		if attr.Kind != Ast_Attribute {
			continue
		}
		if len(attr.Attribute.Elems) == 0 {
			(*attributes)[i] = (*attributes)[len(*attributes)-1]
			*attributes = (*attributes)[:len(*attributes)-1]
			i--
		}
	}
}
