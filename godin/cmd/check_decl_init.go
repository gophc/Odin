package cmd

func check_init_variable(ctx *CheckerContext, e *Entity, operand *Operand, context_name String) *Type {
	if operand.Mode == Addressing_Invalid ||
		operand.Type == t_invalid ||
		e.Type == t_invalid {
		if operand.Mode == Addressing_Builtin {
			begin_error_block()
			defer end_error_block()
			expr_str := expr_to_string(operand.Expr)
			defer gb_string_free(expr_str)
			error(operand.Expr, "Cannot assign built-in procedure '%s' in %.*s", expr_str, context_name.Len, context_name.Data)
			error_line("\tBuilt-in procedures are implemented by the compiler and might not be actually instantiated procedure\n")
			operand.Mode = Addressing_Invalid
		}
		if operand.Mode == Addressing_ProcGroup {
			if e.Type == nil {
				error(operand.Expr, "Cannot determine type from overloaded procedure '%.*s'", operand.ProcGroup.Token.String.Len, operand.ProcGroup.Token.String.Data)
			} else {
				check_assignment(ctx, operand, e.Type, String{Data: strData("variable assignment"), Len: isize(len("variable assignment"))})
				if operand.Mode != Addressing_Type {
					return operand.Type
				}
			}
		}
		if e.Type == nil {
			e.Type = t_invalid
		}
		return nil
	}
	if e.Kind == Entity_Variable {
		e.Variable.InitExpr = operand.Expr
	}
	if operand.Mode == Addressing_Type {
		if e.Type != nil && is_type_typeid(e.Type) && !is_type_polymorphic(operand.Type) {
			add_type_info_type(ctx, operand.Type)
			add_type_and_value(ctx, operand.Expr, Addressing_Value, e.Type, exact_value_typeid(operand.Type))
			return e.Type
		} else {
			begin_error_block()
			defer end_error_block()
			t := type_to_string(operand.Type)
			defer gb_string_free(t)
			if is_type_polymorphic(operand.Type) {
				error(operand.Expr, "Cannot assign a non-specialized polymorphic type '%s' to variable '%.*s'", t, e.Token.String.Len, e.Token.String.Data)
			} else {
				error(operand.Expr, "Cannot assign a type '%s' to variable '%.*s'", t, e.Token.String.Len, e.Token.String.Data)
			}
			if e.Type == nil {
				error_line("\tThe type of the variable '%.*s' cannot be inferred as a type and does not have a default type\n", e.Token.String.Len, e.Token.String.Data)
			}
			e.Type = operand.Type
			return nil
		}
	}
	if e.Type == nil {
		t := operand.Type
		if is_type_untyped(t) {
			if is_type_untyped_uninit(t) {
				error(e.Token, "Invalid use of --- in %.*s", context_name.Len, context_name.Data)
				e.Type = t_invalid
				return nil
			} else if t == t_invalid || is_type_untyped_nil(t) {
				error(e.Token, "Invalid use of untyped nil in %.*s", context_name.Len, context_name.Data)
				e.Type = t_invalid
				return nil
			}
			t = default_type(t)
		}
		if is_type_asm_proc(t) {
			error(e.Token, "Invalid use of inline asm in %.*s", context_name.Len, context_name.Data)
			e.Type = t_invalid
			return nil
		} else if is_type_polymorphic(t) {
			e2 := entity_of_node(operand.Expr)
			if e2 == nil {
				e.Type = t_invalid
				return nil
			}
			if e2.State != EntityState_Resolved {
				e.Type = t
				return nil
			}
			str := type_to_string(t)
			defer gb_string_free(str)
			error(operand.Expr, "Invalid use of a non-specialized polymorphic type '%s' in %.*s", str, context_name.Len, context_name.Data)
			e.Type = t_invalid
			return nil
		} else if is_type_empty_union(t) {
			str := type_to_string(t)
			defer gb_string_free(str)
			error(e.Token, "An empty union '%s' cannot be instantiated in %.*s", str, context_name.Len, context_name.Data)
			e.Type = t_invalid
			return nil
		}
		gb_assert_handler("Assertion Failure", "is_type_typed(t)", "check_decl_init.go", 0, 0)
		e.Type = t
	}
	e.ParentProcDecl = ctx.CurrProcDecl
	check_assignment(ctx, operand, e.Type, context_name)
	if operand.Mode == Addressing_Invalid {
		return nil
	}
	return e.Type
}

func check_init_variables(ctx *CheckerContext, lhs []*Entity, inits []*Ast, context_name String) {
	if len(lhs) == 0 && len(inits) == 0 {
		return
	}
	operands := make([]Operand, 0, 2*len(lhs))
	check_unpack_arguments(ctx, lhs, len(lhs), &operands, inits, UnpackFlag_AllowOk|UnpackFlag_AllowUndef)
	rhs_count := len(operands)
	max := min(len(lhs), rhs_count)
	for i := 0; i < max; i++ {
		e := lhs[i]
		d := decl_info_of_entity(e)
		o := &operands[i]
		check_init_variable(ctx, e, o, context_name)
		if d != nil {
			d.InitExpr = o.Expr
		}
	}
	if rhs_count > 0 && len(lhs) != rhs_count {
		error(lhs[0].Token, "Assignment count mismatch '%d' = '%d'", len(lhs), rhs_count)
	}
}

func override_entity_in_scope(original_entity *Entity, new_entity *Entity) {
	original_name := original_entity.Token.String
	original_intern := entity_interned_name(original_entity)
	hash := original_entity.InternedNameHash
	var found_scope *Scope
	var found_entity *Entity
	scope_lookup_parent(original_entity.Scope, original_intern, &found_scope, &found_entity, hash)
	if found_scope == nil {
		return
	}
	rw_mutex_lock(&found_scope.mutex)
	scope_map_insert(&found_scope.elements, original_intern, hash, new_entity)
	rw_mutex_unlock(&found_scope.mutex)
	original_entity.Flags |= EntityFlag_Overridden
	original_entity.Type = new_entity.Type
	original_entity.Kind = new_entity.Kind
	original_entity.DeclInfo = new_entity.DeclInfo
	original_entity.AliasedOf = new_entity
	original_entity.Identifier.Store(new_entity.Identifier.Load())
	if original_entity.Identifier.Load() != nil &&
		original_entity.Identifier.Load().Kind == Ast_Ident {
		original_entity.Identifier.Load().Ident.Entity = new_entity
	}
}

func check_override_as_type_due_to_aliasing(ctx *CheckerContext, e *Entity, entity *Entity, init *Ast, named_type *Type) bool {
	if entity != nil && entity.Kind == Entity_TypeName {
		if e.Type != nil && is_type_typed(e.Type) {
			return false
		}
		e.Kind = Entity_TypeName
		check_type_decl(ctx, e, init, named_type)
		return true
	}
	return false
}

func check_try_override_const_decl(ctx *CheckerContext, e *Entity, entity *Entity, init *Ast, named_type *Type) bool {
	if entity == nil {
	retry_proc_lit:
		init = unparen_expr(init)
		if init == nil {
			return false
		}
		if init.Kind == Ast_TernaryWhenExpr {
			we := &init.TernaryWhenExpr
			if we.Cond == nil {
				return false
			}
			if we.Cond.TAV.Value.Kind != ExactValue_Bool {
				return false
			}
			if we.Cond.TAV.Value.ValueBool {
				init = we.X
			} else {
				init = we.Y
			}
			goto retry_proc_lit
		}
		if init.Kind == Ast_ProcLit {
			e.Kind = Entity_Procedure
			e.Type = nil
			d := decl_info_of_entity(e)
			d.ProcLit = init
			check_proc_decl(ctx, e, d)
			return true
		}
		return false
	}
	switch entity.Kind {
	case Entity_TypeName:
		if check_override_as_type_due_to_aliasing(ctx, e, entity, init, named_type) {
			return true
		}
	case Entity_Builtin:
		if e.Type != nil {
			return false
		}
		e.Kind = Entity_Builtin
		e.Builtin.ID = entity.Builtin.ID
		e.Type = t_invalid
		return true
	}
	if e.Type != nil && entity.Type != nil {
		x := Operand{}
		x.Type = entity.Type
		x.Mode = Addressing_Variable
		if !check_is_assignable_to(ctx, &x, e.Type) {
			return false
		}
	}
	switch entity.Kind {
	case Entity_ProcGroup, Entity_Procedure:
		override_entity_in_scope(e, entity)
		return true
	}
	return false
}

func check_init_constant(ctx *CheckerContext, e *Entity, operand *Operand) {
	if operand.Mode == Addressing_Invalid ||
		operand.Type == t_invalid ||
		e.Type == t_invalid {
		if e.Type == nil {
			e.Type = t_invalid
		}
		return
	}
	if operand.Mode != Addressing_Constant {
		entity := entity_of_node(operand.Expr)
		if check_try_override_const_decl(ctx, e, entity, operand.Expr, nil) {
			return
		}
	}
	if operand.Mode != Addressing_Constant {
		str := expr_to_string(operand.Expr)
		defer gb_string_free(str)
		error(operand.Expr, "'%s' is not a compile-time known constant", str)
		if e.Type == nil {
			e.Type = t_invalid
		}
		return
	}
	if e.Type == nil {
		e.Type = operand.Type
	}
	check_assignment(ctx, operand, e.Type, String{Data: strData("constant declaration"), Len: isize(len("constant declaration"))})
	if operand.Mode == Addressing_Invalid {
		return
	}
	if is_type_proc(e.Type) {
		error(e.Token, "Illegal declaration of a constant procedure value")
	}
	e.ParentProcDecl = ctx.CurrProcDecl
	e.Constant.Value = operand.Value
}
