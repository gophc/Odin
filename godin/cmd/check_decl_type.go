package cmd

func is_type_distinct(node *Ast) bool {
	for {
		if node == nil {
			return false
		}
		if node.Kind == Ast_ParenExpr {
			node = node.ParenExpr.Expr
		} else if node.Kind == Ast_HelperType {
			node = node.HelperType.Type
		} else {
			break
		}
	}
	switch node.Kind {
	case Ast_DistinctType:
		return true
	case Ast_StructType, Ast_UnionType, Ast_EnumType, Ast_ProcType, Ast_BitFieldType:
		return true
	case Ast_PointerType, Ast_ArrayType, Ast_DynamicArrayType, Ast_MapType:
		return false
	}
	return false
}

func remove_type_alias_clutter(node *Ast) *Ast {
	for {
		if node == nil {
			return nil
		}
		if node.Kind == Ast_ParenExpr {
			node = node.ParenExpr.Expr
		} else if node.Kind == Ast_DistinctType {
			node = node.DistinctType.Type
		} else {
			return node
		}
	}
}

func clone_enum_type(ctx *CheckerContext, original_enum_type *Type, named_type *Type) *Type {
	gb_assert_handler(original_enum_type != nil, "original_enum_type != nil", "cmd_check_decl_type.go", 0)
	gb_assert_handler(named_type != nil, "named_type != nil", "cmd_check_decl_type.go", 0)
	gb_assert_handler(original_enum_type.Kind == Type_Enum, "original_enum_type.Kind == Type_Enum", "cmd_check_decl_type.go", 0)
	gb_assert_handler(named_type.Kind == Type_Named, "named_type.Kind == Type_Named", "cmd_check_decl_type.go", 0)

	parent := original_enum_type.Enum.Scope.Parent
	scope := create_scope(nil, parent)

	et := alloc_type_enum()
	et.Enum.BaseType = original_enum_type.Enum.BaseType
	*et.Enum.MinValue = *original_enum_type.Enum.MinValue
	*et.Enum.MaxValue = *original_enum_type.Enum.MaxValue
	et.Enum.MinValueIndex = original_enum_type.Enum.MinValueIndex
	et.Enum.MaxValueIndex = original_enum_type.Enum.MaxValueIndex
	et.Enum.Scope = scope

	fields := make([]*Entity, len(original_enum_type.Enum.Fields))
	for i, old := range original_enum_type.Enum.Fields {
		e := alloc_entity_constant(scope, old.Token, named_type, old.Constant.Value)
		e.File = old.File
		e.Identifier = clone_ast(old.Identifier, nil)
		e.Flags |= EntityFlag_Visited
		e.State = EntityState_Resolved
		e.Constant.Flags = old.Constant.Flags
		e.Constant.Docs = old.Constant.Docs
		e.Constant.Comment = old.Constant.Comment

		fields[i] = e
		add_entity(ctx, scope, nil, e)
		add_entity_use(ctx, e.Identifier, e)
	}
	et.Enum.Fields = fields
	return et
}

func check_type_decl(ctx *CheckerContext, e *Entity, init_expr *Ast, def *Type) {
	gb_assert_handler(e.Type == nil, "e.Type == nil", "cmd_check_decl_type.go", 0)

	decl := decl_info_of_entity(e)

	is_distinct := is_type_distinct(init_expr)
	te := remove_type_alias_clutter(init_expr)
	e.Type = t_invalid
	name := e.Token.String
	named := alloc_type_named(name, nil, e)
	if def != nil && def.Kind == Type_Named {
		def.Named.Base = named
	}
	e.Type = named

	if !is_distinct {
		e.TypeName.IsTypeAlias = true
	}

	check_type_path_push(ctx, e)
	bt := check_type_expr(ctx, te, named)
	check_type_path_pop(ctx)

	base := base_type(bt)
	if is_distinct && bt.Kind == Type_Named && base.Kind == Type_Enum {
		base = clone_enum_type(ctx, base, named)
	}
	named.Named.Base = base

	if is_distinct {
		if is_type_typeid(e.Type) {
			error(init_expr, "'distinct' cannot be applied to 'typeid'")
			is_distinct = false
		} else if is_type_any(e.Type) {
			error(init_expr, "'distinct' cannot be applied to 'any'")
			is_distinct = false
		} else if is_type_simd_vector(e.Type) || is_type_soa_pointer(e.Type) {
			str := type_to_string(e.Type)
			error(init_expr, "'distinct' cannot be applied to '%s'", goStr(str))
			gb_string_free(str)
			is_distinct = false
		}
	} else {
		if is_type_typeid(e.Type) {
			error(init_expr, "'typeid' cannot be aliased")
		} else if is_type_any(e.Type) {
			error(init_expr, "'any' cannot be aliased")
		}
	}

	if !is_distinct {
		e.Type = bt
		named.Named.Base = bt
	}

	e.TypeName.IsTypeAlias = !is_distinct

	if decl != nil && decl.TypeExpr != nil {
		t := check_type(ctx, decl.TypeExpr)
		if t != nil && !is_type_typeid(t) {
			operand := Operand{}
			operand.Mode = Addressing_Type
			operand.Type = e.Type
			operand.Expr = init_expr
			check_assignment(ctx, &operand, t, S("constant declaration"))
		}
	}

	if decl != nil {
		ac := AttributeContext{}
		check_decl_attributes(ctx, decl.Attributes, type_decl_attribute, &ac)

		e.DeprecatedMessage = ac.DeprecatedMessage

		if e.Kind == Entity_TypeName && ac.ObjcClass != "" {
			e.TypeName.ObjcClassName = ac.ObjcClass

			if ac.ObjcIsImplementation {
				e.TypeName.ObjcIsImplementation = ac.ObjcIsImplementation
				e.TypeName.ObjcSuperclass = ac.ObjcSuperclass
				e.TypeName.ObjcIvar = ac.ObjcIvar
				e.TypeName.ObjcContextProvider = ac.ObjcContextProvider

				mutex_lock(&ctx.info.objc_class_name_mutex)
				class_exists := string_set_update(&ctx.info.obcj_class_name_set, ac.ObjcClass)
				mutex_unlock(&ctx.info.objc_class_name_mutex)
				if class_exists {
					error(e.Token, "@(objc_class) name '%.*s' has already been used elsewhere", int(ac.ObjcClass.Len), goStr(ac.ObjcClass))
				}

				mpsc_enqueue(&ctx.info.objc_class_implementations, e)

				gb_assert_handler(e.TypeName.ObjcIvar == nil || e.TypeName.ObjcIvar.Kind == Type_Named,
					"e.TypeName.ObjcIvar == nil || e.TypeName.ObjcIvar.Kind == Type_Named",
					"cmd_check_decl_type.go", 0)

				if e.TypeName.ObjcContextProvider != nil {
					mpsc_enqueue(&ctx.checker.ProcsWithObjcContextProviderToCheck, e)
				}

				super_set := TypeSet{}
				type_set_init(&super_set, 8)
				defer type_set_destroy(&super_set)

				type_set_update(&super_set, e.Type)

				super := ac.ObjcSuperclass
				for super != nil {
					if super.Kind != Type_Named {
						error(e.Token, "@(objc_superclass) Referenced type must be a named struct")
						break
					}

					if type_set_update(&super_set, super) {
						error(e.Token, "@(objc_superclass) Superclass hierarchy cycle encountered")
						break
					}

					check_single_global_entity(ctx.checker, super.Named.TypeName, super.Named.TypeName.DeclInfo)

					named_type_2 := base_named_type(super)
					gb_assert_handler(named_type_2.Kind == Type_Named,
						"named_type_2.Kind == Type_Named",
						"cmd_check_decl_type.go", 0)

					if !is_type_objc_object(named_type_2) {
						error(e.Token, "@(objc_superclass) Superclass '%.*s' must be an Objective-C class", int(named_type_2.Named.Name.Len), goStr(named_type_2.Named.Name))
						break
					}

					if named_type_2.Named.TypeName.TypeName.ObjcClassName == "" {
						error(e.Token, "@(objc_superclass) Superclass '%.*s' must have a valid @(objc_class) attribute", int(named_type_2.Named.Name.Len), goStr(named_type_2.Named.Name))
						break
					}

					super = named_type_2.Named.TypeName.TypeName.ObjcSuperclass
				}
			} else {
				if ac.ObjcIvar != nil {
					error(e.Token, "@(objc_ivar) may only be applied when the @(obj_implement) attribute is also applied")
				} else if ac.ObjcContextProvider != nil {
					error(e.Token, "@(objc_context_provider) may only be applied when the @(obj_implement) attribute is also applied")
				}
			}

			if type_size_of(e.Type) > 0 {
				error(e.Token, "@(objc_class) marked type must be of zero size")
			}
		} else if ac.ObjcIsImplementation {
			error(e.Token, "@(objc_implement) may only be applied when the @(objc_class) attribute is also applied")
		}

		if ac.RaddbgTypeView {
			type_view := RaddbgTypeView{
				Type: e.Type,
				View: ac.RaddbgTypeViewString,
			}
			mpsc_enqueue(&ctx.info.RaddbgTypeViewsQueue, type_view)
		}
	}

	if decl != nil && decl.IsUsing {
		error(init_expr, "'using' an enum declaration is not allowed, prefer using implicit selector expressions e.g. '.A'")
	}
}
