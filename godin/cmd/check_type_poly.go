package cmd

func ensure_polymorphic_record_entity_has_gen_types(ctx *CheckerContext, original_type *Type) *GenTypesData {
	mutex_lock(&original_type.Named.GenTypesDataMutex)
	if original_type.Named.GenTypesData == nil {
		gen_types := permanent_alloc_item[GenTypesData]()
		gen_types.Types = make([]*Entity, 0)
		original_type.Named.GenTypesData = gen_types
	}
	found_gen_types := original_type.Named.GenTypesData
	mutex_unlock(&original_type.Named.GenTypesDataMutex)
	return found_gen_types
}

func add_polymorphic_record_entity(ctx *CheckerContext, node *Ast, named_type *Type, original_type *Type) {
	s := ctx.Scope.Parent
	pkg := original_type.Named.TypeName.Pkg
	if pkg == nil {
		pkg = ctx.Pkg
	}

	token := ast_token(node)
	token.Kind = Token_String
	token.String = named_type.Named.Name
	ast_node := ast_ident(nil, token)
	e := alloc_entity_type_name(s, token, named_type)
	e.State = EntityState_Resolved
	e.File = ctx.File
	e.Pkg = pkg
	e.TypeName.OriginalTypeForParapoly = original_type
	add_entity_use(ctx, ast_node, e)

	named_type.Named.TypeName = e
	if original_type.Named.TypeName != nil {
		e.TypeName.ObjcClassName = original_type.Named.TypeName.TypeName.ObjcClassName
		e.TypeName.ObjcMetadata = original_type.Named.TypeName.TypeName.ObjcMetadata
	}

	found_gen_types := ensure_polymorphic_record_entity_has_gen_types(ctx, original_type)
	found_gen_types.Mutex.Lock()
	defer found_gen_types.Mutex.Unlock()
	for _, prev := range found_gen_types.Types {
		if prev == e {
			return
		}
	}
	found_gen_types.Types = append(found_gen_types.Types, e)
}

func check_record_polymorphic_params(ctx *CheckerContext, polymorphic_params *Ast, is_polymorphic_ *bool, poly_operands *[]Operand) *Type {
	if polymorphic_params == nil {
		if !*is_polymorphic_ {
			*is_polymorphic_ = polymorphic_params != nil && poly_operands == nil
		}
		return nil
	}

	var polymorphic_params_type *Type

	field_list := &polymorphic_params.FieldList
	params := field_list.List

	if len(params) != 0 {
		variable_count := isize(0)
		for _, param := range params {
			if param.Kind != Ast_Field {
				continue
			}
			count := isize(len(param.Field.Names))
			if count == 0 {
				count = 1
			}
			variable_count += count
		}

		entities := make([]*Entity, 0, variable_count)
		field_group_index := int32(-1)

		for _, param := range params {
			if param.Kind != Ast_Field {
				continue
			}
			field_group_index++
			p := &param.Field

			type_expr := p.Type
			var type_ *Type
			is_type_param := false

			if type_expr == nil {
				error(param, "A polymorphic parameter must have a type, e.g. `$T: typeid`")
				continue
			}

			if type_expr.Kind == Ast_Ellipsis {
				error(type_expr, "A polymorphic parameter cannot be variadic")
				continue
			}

			if type_expr.Kind == Ast_TypeidType {
				is_type_param = true
				specialization := type_expr.TypeidType.Specialization
				if specialization != nil {
					type_ = check_type(ctx, specialization)
					if !is_type_polymorphic(type_) && !is_type_typeid(type_) {
						error(type_expr, "$T/specialization must be a typeid")
					}
					if poly_operands != nil {
						add_specialization_to_type_if_needed(ctx, type_, poly_operands)
					}
				} else {
					type_ = t_typeid
				}
			} else {
				type_ = check_type(ctx, type_expr)
			}

			if type_ == nil {
				type_ = t_invalid
			}

			default_value := EmptyExactValue
			if p.DefaultValue != nil {
				handle_parameter_value(ctx, param, &type_, &default_value)
			}

			if is_type_param {
				if type_ != t_typeid && !is_type_typeid(type_) && !is_type_polymorphic(type_) {
					error(type_expr, "A type parameter must have a type of 'typeid'")
				}
			} else {
				if type_ != nil && type_.Kind == Type_Generic {
					if type_.Generic.Specialized != nil {
						type_ = type_.Generic.Specialized
					}
				}
			}

			for _, name := range p.Names {
				if is_blank_ident(name) {
					continue
				}

				name_token := ast_token(name)

				if poly_operands != nil {
					i := isize(len(entities))
					var operand Operand
					if i < isize(len(*poly_operands)) {
						operand = (*poly_operands)[i]
					}

					if operand.Expr == nil {
						continue
					}

					if is_type_param {
						if operand.Mode != Addressing_Type {
							error(name, "Expected a type for polymorphic parameter '%.*s', got '%s'", name_token.String.Len, goStr(name_token.String), goStr(type_to_string(operand.Type)))
							continue
						}
						e := alloc_entity_type_name(ctx.Scope, name_token, type_, EntityState_Resolved)
						if type_ != nil && !is_type_typeid(type_) {
							type_to_string(type_)
						}
						e.State = EntityState_Resolved
						add_entity(ctx, ctx.Scope, name, e)
						entities = append(entities, e)
					} else {
						if !is_operand_value(operand) {
							error(name, "Expected a value for polymorphic parameter '%.*s'", name_token.String.Len, goStr(name_token.String))
							continue
						}
						if is_type_polymorphic(type_) {
							type_ = determine_type_from_polymorphic(ctx, type_, operand)
						}
						e := alloc_entity_const_param(ctx.Scope, name_token, type_, operand.Value, true)
						e.State = EntityState_Resolved
						add_entity(ctx, ctx.Scope, name, e)
						entities = append(entities, e)
					}
				} else {
					var e *Entity
					if is_type_param {
						e = alloc_entity_type_name(ctx.Scope, name_token, type_, EntityState_Resolved)
					} else {
						e = alloc_entity_const_param(ctx.Scope, name_token, type_, default_value, true)
					}
					e.State = EntityState_Resolved
					add_entity(ctx, ctx.Scope, name, e)
					entities = append(entities, e)
				}
			}
		}

		if len(entities) > 0 {
			tuple := alloc_type_tuple()
			tuple.Tuple.Variables = entities
			polymorphic_params_type = tuple
		}
	}

	if !*is_polymorphic_ {
		*is_polymorphic_ = polymorphic_params != nil && poly_operands == nil
	}
	return polymorphic_params_type
}

func check_record_poly_operand_specialization(ctx *CheckerContext, record_type *Type, poly_operands *[]Operand, is_polymorphic_ *bool) bool {
	if poly_operands == nil {
		return false
	}
	for _, o := range *poly_operands {
		if is_type_polymorphic(o.Type) {
			return false
		}
		if record_type == o.Type {
			return false
		}
		if o.Mode == Addressing_Type {
			entity := entity_of_node(o.Expr)
			if entity != nil && entity.Kind == Entity_TypeName && entity.Type == t_typeid {
				*is_polymorphic_ = true
				return false
			}
		}
	}
	return true
}

func find_polymorphic_record_entity(found_gen_types *GenTypesData, param_count isize, ordered_operands []Operand) *Entity {
	for _, e := range found_gen_types.Types {
		t := base_type(e.Type)
		tuple := get_record_polymorphic_params(t)
		if tuple == nil {
			continue
		}
		if param_count != isize(len(tuple.Variables)) {
			continue
		}
		skip := false
		for j := isize(0); j < param_count; j++ {
			var o Operand
			if j < isize(len(ordered_operands)) {
				o = ordered_operands[j]
			}
			if o.Expr == nil {
				continue
			}
			oe := entity_of_node(o.Expr)
			p := tuple.Variables[j]
			if p == oe {
				continue
			}
			if p.Kind == Entity_TypeName {
				if is_type_polymorphic(o.Type) {
					skip = true
					break
				}
				if !are_types_identical(o.Type, p.Type) {
					skip = true
					break
				}
			} else if p.Kind == Entity_Constant {
				if !compare_exact_values(TokenCmpEq, o.Value, p.Constant.Value) {
					skip = true
					break
				}
				if !are_types_identical(o.Type, p.Type) {
					skip = true
					break
				}
			}
		}
		if !skip {
			return e
		}
	}
	return nil
}

func check_type_specialization_to(ctx *CheckerContext, specialization *Type, type_ *Type, compound bool, modify_type bool) bool {
	if type_ == nil || type_ == t_invalid {
		return true
	}
	t := base_type(type_)
	s := base_type(specialization)
	if t.Kind != s.Kind {
		if t.Kind == Type_EnumeratedArray && s.Kind == Type_Array {
		} else {
			return false
		}
	}
	if is_type_untyped(t) {
		o := Operand{Mode: Addressing_Value, Type: default_type(type_)}
		return check_cast_internal(ctx, &o, specialization)
	}
	if t.Kind == Type_Struct {
		if t.Struct.PolymorphicParent == nil && t == s {
			return true
		}
		if t.Struct.PolymorphicParent == specialization {
			return true
		}
		if t.Struct.PolymorphicParent != nil && t.Struct.PolymorphicParent == s.Struct.PolymorphicParent {
			sp := get_record_polymorphic_params(s)
			tp := get_record_polymorphic_params(t)
			if sp != nil && tp != nil {
				count := isize(len(sp.Variables))
				if count != isize(len(tp.Variables)) {
					return false
				}
				for i := isize(0); i < count; i++ {
					s_e := sp.Variables[i]
					t_e := tp.Variables[i]
					if s_e.Kind == Entity_TypeName {
						if !are_types_identical(s_e.Type, t_e.Type) {
							return false
						}
					} else if s_e.Kind == Entity_Constant {
						if !are_types_identical(s_e.Type, t_e.Type) {
							return false
						}
					} else {
						return false
					}
				}
				if modify_type {
					*specialization = *type_
				}
				return true
			}
		}
	}
	if t.Kind == Type_Union {
		if t.Union.PolymorphicParent == nil && t == s {
			return true
		}
		if t.Union.PolymorphicParent == specialization {
			return true
		}
		if t.Union.PolymorphicParent != nil && t.Union.PolymorphicParent == s.Union.PolymorphicParent {
			sp := get_record_polymorphic_params(s)
			tp := get_record_polymorphic_params(t)
			if sp != nil && tp != nil {
				count := isize(len(sp.Variables))
				if count != isize(len(tp.Variables)) {
					return false
				}
				for i := isize(0); i < count; i++ {
					s_e := sp.Variables[i]
					t_e := tp.Variables[i]
					if s_e.Kind == Entity_TypeName {
						if !are_types_identical(s_e.Type, t_e.Type) {
							return false
						}
					} else if s_e.Kind == Entity_Constant {
						if !are_types_identical(s_e.Type, t_e.Type) {
							return false
						}
					} else {
						return false
					}
				}
				if modify_type {
					*specialization = *type_
				}
				return true
			}
		}
	}
	if specialization.Kind == Type_Named && type_.Kind != Type_Named {
		return false
	}
	return is_polymorphic_type_assignable(ctx, base_type(specialization), base_type(type_), compound, modify_type)
}

func determine_type_from_polymorphic(ctx *CheckerContext, poly_type *Type, operand Operand) *Type {
	modify_type := !ctx.NoPolymorphicErrors
	show_error := modify_type && !ctx.HidePolymorphicErrors

	if !is_operand_value(operand) {
		if show_error {
			begin_error_block()
			pts := type_to_string(poly_type)
			ots := type_to_string(operand.Type)
			error(operand.Expr, "Cannot determine polymorphic type from parameter: '%s' to '%s'", goStr(ots), goStr(pts))
			if operand.Mode == Addressing_Type {
				error_line("\tSuggestion: Are you trying to pass a type to a value parameter?\n")
			}
			end_error_block()
		}
		return t_invalid
	}

	if is_polymorphic_type_assignable(ctx, poly_type, operand.Type, false, modify_type) {
		return poly_type
	}

	if show_error {
		pts := type_to_string(poly_type)
		ots := type_to_string(operand.Type)
		begin_error_block()
		error(operand.Expr, "Cannot determine polymorphic type from parameter: '%s' to '%s'", goStr(ots), goStr(pts))
		if operand.Type.Kind == Type_Slice && poly_type.Kind == Type_Pointer {
			error_line("\tSuggestion: Did you mean to use a slice type rather than a pointer type?\n")
		} else if operand.Type.Kind == Type_Pointer && poly_type.Kind == Type_Slice {
			error_line("\tSuggestion: Did you mean to use a pointer type rather than a slice type?\n")
		}
		end_error_block()
	}

	return t_invalid
}
