package cmd

func check_assignment_variable(ctx *CheckerContext, lhs *Operand, rhs *Operand, context_name String) *Type {
	if rhs.Mode == Addressing_Invalid {
		return nil
	}
	if rhs.Type == t_invalid &&
		rhs.Mode != Addressing_ProcGroup &&
		rhs.Mode != Addressing_Builtin {
		return nil
	}

	node := unparen_expr(lhs.Expr)
	if is_blank_ident(node) {
		check_assignment(ctx, rhs, nil, S("assignment to '_' identifier"))
		if rhs.Mode == Addressing_Invalid {
			return nil
		}
		return rhs.Type
	}

	var e *Entity
	used := false

	if lhs.Mode == Addressing_Invalid ||
		(lhs.Type == t_invalid &&
			lhs.Mode != Addressing_ProcGroup &&
			lhs.Mode != Addressing_Builtin) {
		return nil
	}

	if rhs.Mode == Addressing_ProcGroup {
		procs := proc_group_entities(ctx, *rhs)
		if len(procs) > 0 {
			for i := 0; i < len(procs); i++ {
				t := base_type(procs[i].Type)
				if t == t_invalid {
					continue
				}
				x := Operand{Mode: Addressing_Value, Type: t}
				if check_is_assignable_to(ctx, &x, lhs.Type) {
					e = procs[i]
					add_entity_use(ctx, rhs.Expr, e)
					break
				}
			}
		}
		if e != nil {
			rhs.Mode = Addressing_Value
			rhs.Type = e.Type
			rhs.ProcGroup = nil
		}
	} else {
		var ident_node *Ast
		if node.Kind == Ast_Ident {
			ident_node = node
		} else if node.Kind == Ast_IndexExpr && node.IndexExpr.expr.Kind == Ast_Ident {
			ident_node = node.IndexExpr.expr
		}
		if ident_node != nil {
			e = scope_lookup(ctx.scope, ident_node.Ident.interned, ident_node.Ident.hash)
			if e != nil && e.Kind == Entity_Variable {
				used = (e.Flags & EntityFlag_Used) != 0
			}
		}
	}

	if e != nil && used {
		e.Flags |= EntityFlag_Used
	}

	assignment_type := lhs.Type

	if rhs.Mode == Addressing_Type && is_type_polymorphic(rhs.Type) {
		t := type_to_string(rhs.Type)
		error(rhs.Expr, "Invalid use of a non-specialized polymorphic type '%s'", t)
		gb_string_free(t)
	}

	switch lhs.Mode {
	case Addressing_Invalid:
		return nil
	case Addressing_Variable:
		if e != nil && e.Kind == Entity_Variable && e.Variable.is_rodata {
			error(lhs.Expr, "Assignment to variable '%.*s' marked as @(rodata) is not allowed", e.Token.String.Len, e.Token.String.Data)
		}
	case Addressing_MapIndex:
		ln := unparen_expr(lhs.Expr)
		if ln.Kind == Ast_IndexExpr {
			x := ln.IndexExpr.expr
			tav := x.TAV
			if tav.Mode != Addressing_Variable {
				if !is_type_pointer(tav.Type) {
					str := expr_to_string(lhs.Expr)
					error(lhs.Expr, "Cannot assign to the value of a map '%s'", str)
					gb_string_free(str)
					return nil
				}
			}
		}
	case Addressing_Context:
	case Addressing_SoaVariable:
	case Addressing_SwizzleVariable:
	default:
		if lhs.Expr.Kind == Ast_SelectorExpr {
			se := &lhs.Expr.SelectorExpr
			op_c := Operand{Mode: Addressing_Invalid}
			check_expr(ctx, &op_c, se.expr)
			if op_c.Mode == Addressing_MapIndex {
				str := expr_to_string(lhs.Expr)
				error(lhs.Expr, "Cannot assign to struct field '%s' in map", str)
				gb_string_free(str)
				return nil
			}
		}
		le := entity_of_node(lhs.Expr)
		original_e := le
		name := unparen_expr(lhs.Expr)
		for name.Kind == Ast_SelectorExpr {
			name = name.SelectorExpr.expr
			le = entity_of_node(name)
		}
		if le == nil {
			le = original_e
		}
		str := expr_to_string(lhs.Expr)
		if le != nil && le.Flags&EntityFlag_Param != 0 {
			begin_error_block()
			if le.Flags&EntityFlag_Using != 0 {
				error(lhs.Expr, "Cannot assign to '%s' which is from a 'using' procedure parameter", str)
			} else {
				error(lhs.Expr, "Cannot assign to '%s' which is a procedure parameter", str)
			}
			if is_type_pointer(le.Type) {
				error_line("\tSuggestion: Did you mean to shadow it? '%.*s := %.*s'?\n", le.Token.String.Len, le.Token.String.Data, le.Token.String.Len, le.Token.String.Data)
			} else {
				error_line("\tSuggestion: Did you mean to pass '%.*s' by pointer?\n", le.Token.String.Len, le.Token.String.Data)
			}
			show_error_on_line(le.Token.Pos, token_pos_end(le.Token))
			end_error_block()
		} else {
			begin_error_block()
			error(lhs.Expr, "Cannot assign to '%s'", str)
			if le != nil && le.Flags&EntityFlag_ForValue != 0 {
				offset := show_error_on_line(le.Token.Pos, token_pos_end(le.Token))
				if offset < 0 {
					if is_type_map(le.Type) {
						error_line("\tSuggestion: Did you mean? 'for key, &%.*s in ...'\n", le.Token.String.Len, le.Token.String.Data)
					} else {
						error_line("\tSuggestion: Did you mean? 'for &%.*s in ...'\n", le.Token.String.Len, le.Token.String.Data)
					}
				} else {
					error_line("\t")
					for i := isize(0); i < offset-1; i++ {
						error_line(" ")
					}
					error_line("'%.*s' is immutable, declare it as '&%.*s' to make it mutable\n", le.Token.String.Len, le.Token.String.Data, le.Token.String.Len, le.Token.String.Data)
				}
			} else if le != nil && le.Flags&EntityFlag_SwitchValue != 0 {
				offset := show_error_on_line(le.Token.Pos, token_pos_end(le.Token))
				if offset < 0 {
					error_line("\tSuggestion: Did you mean? 'switch &%.*s in ...'\n", le.Token.String.Len, le.Token.String.Data)
				} else {
					error_line("\t")
					for i := isize(0); i < offset-1; i++ {
						error_line(" ")
					}
					error_line("'%.*s' is immutable, declare it as '&%.*s' to make it mutable\n", le.Token.String.Len, le.Token.String.Data, le.Token.String.Len, le.Token.String.Data)
				}
			}
			end_error_block()
		}
		gb_string_free(str)
	}

	lhs_e := entity_of_node(lhs.Expr)
	prev_bit_field_bit_size := ctx.bit_field_bit_size
	if lhs_e != nil && lhs_e.Kind == Entity_Variable && lhs_e.Variable.bit_field_bit_size != 0 {
		ctx.bit_field_bit_size = lhs_e.Variable.bit_field_bit_size
	}
	check_assignment(ctx, rhs, assignment_type, context_name)
	ctx.bit_field_bit_size = prev_bit_field_bit_size
	if rhs.Mode == Addressing_Invalid {
		return nil
	}
	return rhs.Type
}

func error_var_decl_identifier(name *Ast) {
	if name.Kind != Ast_Ident {
		begin_error_block()
		s := expr_to_string(name)
		error(name, "A variable declaration must be an identifier, got '%s'", s)
		gb_string_free(s)
		if name.Kind == Ast_Implicit {
			imp := name.Implicit.string
			if imp == "context" {
				error_line("\tSuggestion: '%.*s' is a reserved keyword, would 'ctx' suffice?\n", imp.Len, imp.Data)
			} else {
				error_line("\tNote: '%.*s' is a reserved keyword\n", imp.Len, imp.Data)
			}
		}
		end_error_block()
	}
}

func check_value_decl_stmt(ctx *CheckerContext, node *Ast, mod_flags u32) {
	vd := &node.ValueDecl
	if !vd.is_mutable {
		for _, name := range vd.names {
			if is_blank_ident(name) {
				e := name.Ident.entity.Load()
				d := decl_info_of_entity(e)
				if d != nil {
					check_entity_decl(ctx, e, d, nil)
				}
			}
		}
		return
	}

	entities := make([]*Entity, 0, len(vd.names))
	entity_count := isize(0)
	new_name_count := isize(0)

	for _, name := range vd.names {
		var entity *Entity
		if name.Kind != Ast_Ident {
			error_var_decl_identifier(name)
		} else {
			token := name.Ident.token
			str := token.string
			var found *Entity
			if !is_blank_ident_str(str) {
				found = scope_lookup_current(ctx.scope, name.Ident.interned, name.Ident.hash)
				new_name_count++
			}
			if found == nil {
				entity = alloc_entity_variable(ctx.scope, token, nil)
				entity.identifier = name
				fl := ctx.foreign_context.curr_library
				if fl != nil {
					entity.Variable.is_foreign = true
					entity.Variable.foreign_library_ident = fl
				}
			} else {
				pos := found.Token.Pos
				error(token,
					"Redeclaration of '%.*s' in this scope\n"+
						"\tat %s",
					str.Len, str.Data, token_pos_to_string(pos))
				entity = found
			}
		}
		if entity == nil {
			entity = alloc_entity_dummy_variable(builtin_pkg.scope, ast_token(name))
		}
		entity.parent_proc_decl = ctx.curr_proc_decl
		entities = append(entities, entity)
		entity_count++
		if name.Kind == Ast_Ident {
			name.Ident.entity.Store(entity)
		}
	}

	if new_name_count == 0 {
		begin_error_block()
		error(node, "No new declarations on the left hand side")
		all_underscore := true
		for _, name := range vd.names {
			if name.Kind == Ast_Ident {
				if !is_blank_ident(name) {
					all_underscore = false
					break
				}
			} else {
				all_underscore = false
				break
			}
		}
		if all_underscore {
			error_line("\tSuggestion: Try changing the declaration (:=) to an assignment (=)\n")
		}
		end_error_block()
	}

	var init_type *Type
	if vd._type != nil {
		init_type = check_type(ctx, vd._type)
		if init_type == nil {
			init_type = t_invalid
		}
		if init_type == t_invalid && entity_count == 1 && (mod_flags&(Stmt_BreakAllowed|Stmt_FallthroughAllowed)) != 0 {
			e := entities[0]
			if e != nil && e.Token.string == "default" {
				warning(e.Token, "Did you mean 'case:'?")
			}
		}
	}

	ac := &AttributeContext{}
	check_decl_attributes(ctx, vd.attributes, var_decl_attribute, ac)

	for i := isize(0); i < entity_count; i++ {
		e := entities[i]
		if e.Flags&EntityFlag_Visited != 0 {
			e.Type = t_invalid
			continue
		}
		e.Flags |= EntityFlag_Visited
		e.State = EntityState_InProgress
		if e.Type == nil {
			e.Type = init_type
			e.State = EntityState_Resolved
		}
		ac.LinkName = handle_link_name(ctx, e.Token, ac.LinkName, ac.LinkPrefix, ac.LinkSuffix)
		if ac.LinkName.Len > 0 {
			e.Variable.link_name = ac.LinkName
		}
		e.Flags &^= EntityFlag_Static
		if ac.IsStatic {
			name := e.Token.string
			if name == "_" {
				error(e.Token, "The 'static' attribute is not allowed to be applied to '_'")
			} else {
				e.Flags |= EntityFlag_Static
				if ctx.in_defer {
					error(e.Token, "'static' variables cannot be declared within a defer statement")
				}
			}
		}
		if ac.Rodata {
			if ac.IsStatic {
				e.Variable.is_rodata = true
			} else {
				error(e.Token, "Only global or @(static) variables can have @(rodata) applied")
			}
		}
		if ac.ThreadLocalModel.Len > 0 {
			name := e.Token.string
			if name == "_" {
				error(e.Token, "The 'thread_local' attribute is not allowed to be applied to '_'")
			} else {
				e.Flags |= EntityFlag_Static
				if ctx.in_defer {
					error(e.Token, "'thread_local' variables cannot be declared within a defer statement")
				}
			}
			e.Variable.thread_local_model = ac.ThreadLocalModel
		}
		if ac.IsStatic && ac.ThreadLocalModel.Len > 0 {
			error(e.Token, "The 'static' attribute is not needed if 'thread_local' is applied")
		}
	}

	prev_type_hint_expr := ctx.type_hint_expr
	ctx.type_hint_expr = vd._type
	check_init_variables(ctx, entities, entity_count, vd.values, S("variable declaration"))
	ctx.type_hint_expr = prev_type_hint_expr

	check_arity_match(ctx, vd, false)

	for i := isize(0); i < entity_count; i++ {
		e := entities[i]
		if e.Variable.is_foreign {
			if len(vd.values) > 0 {
				error(e.Token, "A foreign variable declaration cannot have a default value")
			}
			name := e.Token.string
			if e.Variable.link_name.Len > 0 {
				name = e.Variable.link_name
			}
			init_entity_foreign_library(ctx, e)
			fp := &ctx.checker.Info.foreigns
			key := string_hash_string(name)
			found := string_map_get(fp, key)
			if found != nil {
				f := found
				pos := f.Token.Pos
				this_type := base_type(e.Type)
				other_type := base_type(f.Type)
				if !signature_parameter_similar_enough(this_type, other_type) {
					error(e.Token,
						"Foreign entity '%.*s' previously declared elsewhere with a different type\n"+
							"\tat %s",
						name.Len, name.Data, token_pos_to_string(pos))
				}
			} else {
				string_map_set(fp, key, e)
			}
		} else if e.Flags&EntityFlag_Static != 0 {
			if len(vd.values) > 0 {
				if entity_count != isize(len(vd.values)) {
					error(e.Token, "A static variable declaration with a default value must be constant")
				} else {
					value := vd.values[i]
					if value.TAV.Mode != Addressing_Constant {
						error(e.Token, "A static variable declaration with a default value must be constant")
					}
				}
			}
		}
		add_entity(ctx, ctx.scope, e.identifier, e)
	}

	if vd.is_using {
		token := ast_token(node)
		if vd._type != nil && entity_count > 1 {
			error(token, "'using' can only be applied to one variable of the same type")
		}
		for entity_index := isize(0); entity_index < 1; entity_index++ {
			e := entities[entity_index]
			if e == nil {
				continue
			}
			if e.Kind != Entity_Variable {
				continue
			}
			name := e.Token.string
			t := base_type(type_deref(e.Type))
			if is_blank_ident_str(name) {
				error(token, "'using' cannot be applied variable declared as '_'")
			} else if is_type_struct(t) || is_type_raw_union(t) {
				begin_error_block()
				scope := t.Struct.scope
				for iter := beginScopeMap(&scope.elements); ; {
					key, f, ok := ScopeMapIteratorNext(&iter)
					if !ok {
						break
					}
					_ = key
					if f.Kind == Entity_Variable {
						uvar := alloc_entity_using_variable(e, f.Token, f.Type, e.identifier)
						uvar.Flags |= (e.Flags & EntityFlag_Value)
						prev := scope_insert(ctx.scope, uvar)
						if prev != nil {
							error(token, "Namespace collision while 'using' '%.*s' of: %.*s", name.Len, name.Data, prev.Token.string.Len, prev.Token.string.Data)
							end_error_block()
							return
						}
					}
				}
				add_entity_use(ctx, nil, e)
				end_error_block()
			} else {
				error(token, "'using' can only be applied to variables of type struct or raw_union")
				return
			}
		}
	}
}

func check_expr_stmt(ctx *CheckerContext, node *Ast) {
	es := &node.ExprStmt
	operand := Operand{Mode: Addressing_Invalid}
	kind := check_expr_base(ctx, &operand, es.expr, nil)
	switch operand.Mode {
	case Addressing_Type:
		str := type_to_string(operand.Type)
		error(node, "'%s' is not an expression but a type and cannot be used as a statement", str)
		gb_string_free(str)
	case Addressing_NoValue:
		return
	}
	if kind == ExprKind_Stmt {
		return
	}
	expr := strip_or_return_expr(operand.Expr)
	if expr != nil && expr.Kind == Ast_CallExpr {
		builtin_id := BuiltinProc_Invalid
		do_require := false
		ce := &expr.CallExpr
		t := base_type(type_of_expr(ce.proc))
		if t != nil && t.Kind == Type_Proc {
			do_require = t.Proc.require_results
		} else if check_stmt_internal_builtin_proc_id(ce.proc, &builtin_id) {
			bp := builtin_procs[builtin_id]
			do_require = bp.Kind == ExprKind_Expr && !bp.IgnoreResults
		}
		if do_require {
			expr_str := expr_to_string(ce.proc)
			if builtin_id != BuiltinProc_Invalid {
				real_name := builtin_procs[builtin_id].Name
				if real_name != S(expr_str) {
					error(node, "'%s' ('%.*s.%s') requires that its results must be handled", expr_str,
						builtin_proc_pkg_name[builtin_procs[builtin_id].Pkg].Len, builtin_proc_pkg_name[builtin_procs[builtin_id].Pkg].Data,
						goStr(real_name))
					return
				}
			}
			error(node, "'%s' requires that its results must be handled", expr_str)
		}
		return
	} else if expr != nil && expr.Kind == Ast_SelectorCallExpr {
		builtin_id := BuiltinProc_Invalid
		do_require := false
		se := &expr.SelectorCallExpr
		ce := &se.call.CallExpr
		t := base_type(type_of_expr(ce.proc))
		if t == nil {
			expr_str := expr_to_string(ce.proc)
			error(node, "'%s' is not a value field nor procedure", expr_str)
			gb_string_free(expr_str)
			return
		}
		if t.Kind == Type_Proc {
			do_require = t.Proc.require_results
		} else if check_stmt_internal_builtin_proc_id(ce.proc, &builtin_id) {
			bp := builtin_procs[builtin_id]
			do_require = bp.Kind == ExprKind_Expr && !bp.IgnoreResults
		}
		if do_require {
			expr_str := expr_to_string(ce.proc)
			error(node, "'%s' requires that its results must be handled", expr_str)
			gb_string_free(expr_str)
		}
		return
	}

	begin_error_block()
	expr_str := expr_to_string(operand.Expr)
	error(node, "Expression is not used: '%s'", expr_str)
	gb_string_free(expr_str)
	if operand.Expr.Kind == Ast_BinaryExpr {
		be := &operand.Expr.BinaryExpr
		if be.op.Kind != Token_CmpEq {
			end_error_block()
			return
		}
		switch be.left.TAV.Mode {
		case Addressing_Context, Addressing_Variable, Addressing_MapIndex, Addressing_SoaVariable:
			lhs := expr_to_string(be.left)
			rhs := expr_to_string(be.right)
			error_line("\tSuggestion: Did you mean to do an assignment?\n")
			error_line("\t            '%s = %s;'\n", lhs, rhs)
			gb_string_free(rhs)
			gb_string_free(lhs)
		}
	}
	end_error_block()
}

func check_assign_stmt(ctx *CheckerContext, node *Ast) {
	as := &node.AssignStmt
	if as.op.Kind == Token_Eq {
		lhs_count := isize(len(as.lhs))
		if lhs_count == 0 {
			error(as.op, "Missing LHS in assignment statement")
			return
		}
		lhs_operands := make([]Operand, lhs_count)
		for i := isize(0); i < lhs_count; i++ {
			if is_blank_ident(as.lhs[i]) {
				o := &lhs_operands[i]
				o.Expr = as.lhs[i]
				o.Mode = Addressing_Value
			} else {
				ctx.assignment_lhs_hint = unparen_expr(as.lhs[i])
				check_expr(ctx, &lhs_operands[i], as.lhs[i])
			}
		}
		ctx.assignment_lhs_hint = nil
		var rhs_operands []Operand
		check_assignment_arguments(ctx, lhs_operands, &rhs_operands, as.rhs)
		lhs_to_ignore := make([]bool, lhs_count)
		rhs_count := isize(len(rhs_operands))
		max := lhs_count
		if rhs_count < max {
			max = rhs_count
		}
		for i := isize(0); i < max; i++ {
			if lhs_to_ignore[i] {
				continue
			}
			check_assignment_variable(ctx, &lhs_operands[i], &rhs_operands[i], S("assignment"))
		}
		if lhs_count != rhs_count {
			error(as.lhs[0], "Assignment count mismatch '%d' = '%d'", lhs_count, rhs_count)
		}
	} else {
		op := as.op
		if isize(len(as.lhs)) != 1 || isize(len(as.rhs)) != 1 {
			error(op, "Assignment operator '%.*s' requires single-valued operands", op.string.Len, op.string.Data)
			return
		}
		if !(Token_AddEq < op.Kind && op.Kind < Token_SubEq) {
			error(op, "Unknown assignment operator '%.*s'", op.string.Len, op.string.Data)
			return
		}
		lhs := Operand{Mode: Addressing_Invalid}
		rhs := Operand{Mode: Addressing_Invalid}
		binary_expr := alloc_ast_node(node.FileID, Ast_BinaryExpr)
		be := &binary_expr.BinaryExpr
		be.op = op
		be.op.Kind = TokenKind(int32(be.op.Kind) - (int32(Token_AddEq) - int32(Token_Add)))
		be.left = as.lhs[0]
		be.right = as.rhs[0]
		check_expr(ctx, &lhs, as.lhs[0])
		check_binary_expr(ctx, &rhs, binary_expr, nil, true)
		if rhs.Mode != Addressing_Invalid {
			be.op.string = substring(be.op.string, 0, be.op.string.Len-1)
			rhs.Expr = binary_expr
			check_assignment_variable(ctx, &lhs, &rhs, S("assignment operation"))
		}
	}
}
