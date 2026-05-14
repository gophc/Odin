package cmd

func handle_parameter_value(ctx *CheckerContext, in_type *Type, out_type_ **Type, expr *Ast, allow_caller_location bool) ParameterValue {
	param_value := ParameterValue{}
	param_value.OriginalAstExpr = expr
	if expr == nil {
		return param_value
	}

	o := Operand{}
	if allow_caller_location && expr.Kind == Ast_BasicDirective && expr.BasicDirective.Name.String == "caller_location" {
		init_core_source_code_location(ctx.Checker)
		param_value.Kind = ParameterValue_Location
		o.Type = t_source_code_location
		o.Mode = Addressing_Value
		o.Expr = expr
		if in_type != nil {
			check_assignment(ctx, &o, in_type, String{Data: strData("parameter value"), Len: 15})
		}
	} else if is_caller_expression(expr) {
		if expr.Kind != Ast_BasicDirective {
			check_builtin_procedure_directive(ctx, &o, expr, t_string)
		}
		param_value.Kind = ParameterValue_Expression
		o.Type = t_string
		o.Mode = Addressing_Value
		o.Expr = expr
		if in_type != nil {
			check_assignment(ctx, &o, in_type, String{Data: strData("parameter value"), Len: 15})
		}
	} else {
		if in_type != nil {
			check_expr_with_type_hint(ctx, &o, expr, in_type)
		} else {
			check_expr(ctx, &o, expr)
		}
		if in_type != nil {
			check_assignment(ctx, &o, in_type, String{Data: strData("parameter value"), Len: 15})
		}
		if is_operand_nil(o) {
			param_value.Kind = ParameterValue_Nil
		} else if o.Mode != Addressing_Constant {
			if expr.Kind == Ast_ProcLit {
				param_value.Kind = ParameterValue_Constant
				param_value.Value = exact_value_procedure(expr)
			} else {
				e := entity_from_expr(o.Expr)
				if e != nil {
					if e.Kind == Entity_Procedure {
						id := e.Identifier.Load()
						param_value.Kind = ParameterValue_Constant
						param_value.Value = exact_value_procedure(id)
						add_entity_use(ctx, id, e)
					} else {
						if e.Flags&EntityFlag_Param != 0 {
							error(expr, "Default parameter cannot be another parameter")
						} else if is_expr_from_a_parameter(ctx, expr) {
							error(expr, "Default parameter cannot be another parameter")
						} else {
							param_value.Kind = ParameterValue_Value
							param_value.AstValue = expr
							add_entity_use(ctx, e.Identifier.Load(), e)
						}
					}
				} else if allow_caller_location && o.Mode == Addressing_Context {
					param_value.Kind = ParameterValue_Value
					param_value.AstValue = expr
				} else if o.Value.Kind != ExactValue_Invalid {
					param_value.Kind = ParameterValue_Constant
					param_value.Value = o.Value
				} else {
					error(expr, "Default parameter must be a constant, got %s", expr_to_string(o.Expr))
				}
			}
		} else {
			if o.Value.Kind != ExactValue_Invalid {
				param_value.Kind = ParameterValue_Constant
				param_value.Value = o.Value
			} else {
				error(o.Expr, "Invalid constant parameter, got '%s'", expr_to_string(o.Expr))
			}
		}
	}
	if out_type_ != nil {
		if in_type != nil {
			*out_type_ = in_type
		} else {
			*out_type_ = default_type(o.Type)
		}
	}
	return param_value
}

func check_get_params(ctx *CheckerContext, scope *Scope, _params *Ast, is_variadic_ *bool, variadic_index_ *isize, success_ *bool, specialization_count_ *isize, operands []Operand) *Type {
	success := true
	if _params == nil {
		if success_ != nil {
			*success_ = success
		}
		return nil
	}

	field_list := &_params.FieldList
	params := field_list.List
	if len(params) == 0 {
		if success_ != nil {
			*success_ = success
		}
		return nil
	}

	variable_count := isize(0)
	for _, param := range params {
		if param.Kind != Ast_Field {
			continue
		}
		p := &param.Field
		names := p.Names
		if len(names) > 0 {
			variable_count += isize(len(names))
		} else {
			variable_count += 1
		}
	}

	is_variadic := false
	variadic_index := isize(-1)
	is_c_vararg := false

	variables := make([]*Entity, 0, variable_count)

	for _, param := range params {
		if param.Kind != Ast_Field {
			continue
		}

		p := &param.Field
		names := p.Names

		type_expr := unparen_expr(p.Type)

		if type_expr != nil && type_expr.Kind == Ast_Ellipsis {
			e := &type_expr.Ellipsis
			type_expr = unparen_expr(e.Expr)
			is_variadic = true
			variadic_index = isize(len(variables))

			if is_type_polymorphic(type_of_expr(type_expr)) {
				type_expr = e.Expr
			}
		}

		if type_expr != nil && type_expr.Kind == Ast_TypeidType {
			spec := type_expr.TypeidType.Specialization
			if spec != nil {
				type_expr = spec
			}
		}

		if type_expr != nil && type_expr.Kind == Ast_PolyType {
			pt := &type_expr.PolyType
			if p.DefaultValue == nil && !is_variadic {
				p.DefaultValue = pt.Specialization
				if p.DefaultValue != nil {
					p.DefaultValue = unparen_expr(p.DefaultValue)
				}
			}
		}

		if p.Flags&uint32(FieldFlagCVararg) != 0 {
			is_c_vararg = true
		}

		default_value := ParameterValue{}
		if p.DefaultValue != nil {
			allow_caller_location := false
			if !is_variadic && !is_c_vararg {
				allow_caller_location = true
			}
			default_value = handle_parameter_value(ctx, nil, nil, p.DefaultValue, allow_caller_location)
		}

		if type_expr == nil {
			error(param, "Expected a parameter type")
			success = false
			break
		}

		is_type_param := false

		for name_index, name := range names {
			token := astToken(name)
			if name == nil {
				token = makeTokenIdent("_")
			}
			is_poly_name := name != nil && name.Kind == Ast_PolyType

			if is_poly_name {
				if type_expr.Kind == Ast_TypeidType {
					is_type_param = true
				} else if is_variadic {
					error(name, "Variadic polymorphic parameters are not allowed")
				} else {
					if has_parameter_value(default_value) {
						error(name, "A constant polymorphic parameter cannot have a default value")
					}
				}
			}

			if has_parameter_value(default_value) && !is_poly_name && name_index >= 1 {
				default_value = ParameterValue{}
			}

			var param_entity *Entity

			if is_type_param {
				type_ := type_of_expr(type_expr)
				if operands != nil && isize(len(operands)) > 0 {
					index := isize(len(variables))
					if index < isize(len(operands)) {
						o := operands[index]
						if type_ == nil || type_ == t_invalid {
							type_ = o.Type
						} else if o.Type != nil {
							if o.Mode == Addressing_Constant && o.Type.Kind == Type_Generic {
							} else if !are_types_identical(type_, o.Type) {
								if !is_type_polymorphic(type_) {
									error(name, "Expected type of '%s', got '%s'", type_to_string(type_), type_to_string(o.Type))
								} else {
									type_ = o.Type
								}
							}
						}
					}
				}
				param_entity = alloc_entity_type_name(scope, token, type_)
				param_entity.TypeName.IsTypeAlias = true
				param_entity.State = EntityState_Resolved
			} else {
				poly_const := is_poly_name && !is_type_param

				if is_poly_name {
					ev := ExactValue{}
					param_entity = alloc_entity_const_param(scope, token, type_of_expr(type_expr), ev, true)
				} else {
					is_using := p.Flags&uint32(FieldFlagUsing) != 0
					is_value := p.Flags&uint32(FieldFlagConst) != 0
					param_entity = alloc_entity_param(scope, token, type_of_expr(type_expr), is_using, is_value)
				}

				if poly_const {
					if type_expr != nil && type_expr.Kind == Ast_TypeidType {
						param_entity.Kind = Entity_TypeName
						param_entity.TypeName.IsTypeAlias = true
					}
				}

				if p.DefaultValue != nil {
					if has_parameter_value(default_value) {
						if param_entity.Kind == Entity_Constant {
							if default_value.Kind == ParameterValue_Constant {
								param_entity.Constant.Value = default_value.Value
							}
						}
					}
				}
			}

			if is_variadic && isize(len(variables)) == variadic_index {
				param_entity.Flags |= EntityFlag_Ellipsis
			}
			if p.Flags&uint32(FieldFlagNoAlias) != 0 {
				param_entity.Flags |= EntityFlag_NoAlias
			}
			if p.Flags&uint32(FieldFlagByPtr) != 0 {
				param_entity.Flags |= EntityFlag_ByPtr
			}
			if p.Flags&uint32(FieldFlagNoBroadcast) != 0 {
				param_entity.Flags |= EntityFlag_NoBroadcast
			}
			if p.Flags&uint32(FieldFlagNoCapture) != 0 {
				param_entity.Flags |= EntityFlag_NoCapture
			}
			if p.Flags&uint32(FieldFlagConst) != 0 {
				param_entity.Flags |= EntityFlag_ConstInput
			}
			if p.Flags&uint32(FieldFlagAnyInt) != 0 {
				param_entity.Flags |= EntityFlag_AnyInt
			}

			param_entity.State = EntityState_Resolved
			add_entity(ctx, scope, name, param_entity)
			variables = append(variables, param_entity)
		}

		if len(names) == 0 {
			token := makeTokenIdent("_")
			is_using := p.Flags&uint32(FieldFlagUsing) != 0
			is_value := p.Flags&uint32(FieldFlagConst) != 0
			param_entity := alloc_entity_param(scope, token, type_of_expr(type_expr), is_using, is_value)
			param_entity.Flags &= ^EntityFlag_Used

			if is_variadic && isize(len(variables)) == variadic_index {
				param_entity.Flags |= EntityFlag_Ellipsis
			}
			if p.Flags&uint32(FieldFlagNoAlias) != 0 {
				param_entity.Flags |= EntityFlag_NoAlias
			}
			if p.Flags&uint32(FieldFlagByPtr) != 0 {
				param_entity.Flags |= EntityFlag_ByPtr
			}
			if p.Flags&uint32(FieldFlagNoBroadcast) != 0 {
				param_entity.Flags |= EntityFlag_NoBroadcast
			}
			if p.Flags&uint32(FieldFlagNoCapture) != 0 {
				param_entity.Flags |= EntityFlag_NoCapture
			}
			if p.Flags&uint32(FieldFlagConst) != 0 {
				param_entity.Flags |= EntityFlag_ConstInput
			}
			if p.Flags&uint32(FieldFlagAnyInt) != 0 {
				param_entity.Flags |= EntityFlag_AnyInt
			}

			param_entity.State = EntityState_Resolved
			add_entity(ctx, scope, nil, param_entity)
			variables = append(variables, param_entity)
		}
	}

	if !success {
		if success_ != nil {
			*success_ = success
		}
		if is_variadic_ != nil {
			*is_variadic_ = is_variadic
		}
		if variadic_index_ != nil {
			*variadic_index_ = variadic_index
		}
		if specialization_count_ != nil {
			*specialization_count_ = 0
		}
		return nil
	}

	specialization_count := isize(0)
	for _, v := range variables {
		if v.Kind == Entity_TypeName {
			specialization_count++
		} else if v.Kind == Entity_Constant {
			specialization_count++
		}
	}

	tuple := alloc_type_tuple()
	tuple.Tuple.Variables = variables

	if success_ != nil {
		*success_ = success
	}
	if specialization_count_ != nil {
		*specialization_count_ = specialization_count
	}
	if is_variadic_ != nil {
		*is_variadic_ = is_variadic
	}
	if variadic_index_ != nil {
		*variadic_index_ = variadic_index
	}
	return tuple
}

func check_get_results(ctx *CheckerContext, scope *Scope, _results *Ast) *Type {
	if _results == nil {
		return nil
	}

	field_list := &_results.FieldList
	params := field_list.List
	if len(params) == 0 {
		return nil
	}

	variables := make([]*Entity, 0)

	seen_names := make(map[String]*Entity)

	for _, param := range params {
		if param.Kind != Ast_Field {
			continue
		}
		p := &param.Field
		names := p.Names
		type_expr := unparen_expr(p.Type)

		if type_expr == nil {
			error(param, "Expected a result type")
			continue
		}

		if len(names) == 0 {
			token := makeTokenIdent("_")
			entity := alloc_entity_param(scope, token, type_of_expr(type_expr), false, false)
			entity.Flags |= EntityFlag_Result
			entity.Flags &= ^EntityFlag_Used
			entity.State = EntityState_Resolved
			add_entity(ctx, scope, nil, entity)
			variables = append(variables, entity)
		} else {
			for _, name := range names {
				token := astToken(name)
				entity := alloc_entity_param(scope, token, type_of_expr(type_expr), false, false)
				entity.Flags |= EntityFlag_Result
				entity.Flags &= ^EntityFlag_Used
				entity.State = EntityState_Resolved

				if p.DefaultValue != nil {
					allow_caller_location := true
					dv := handle_parameter_value(ctx, nil, nil, p.DefaultValue, allow_caller_location)
					if has_parameter_value(dv) {
						entity.Constant.Value = dv.Value
					}
				}

				if prev := seen_names[token.String]; prev != nil {
					error(name, "Duplicate return value name '%.*s' in this scope", token.String.Len, token.String.Data)
				} else if !is_blank_ident_str(token.String) {
					seen_names[token.String] = entity
				}

				add_entity(ctx, scope, name, entity)
				variables = append(variables, entity)
			}
		}
	}

	tuple := alloc_type_tuple()
	tuple.Tuple.Variables = variables
	return tuple
}
