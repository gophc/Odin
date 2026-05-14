package cmd

func check_valid_type_switch_type(typ *Type) TypeSwitchKind {
	typ = type_deref(typ, false)
	if is_type_union(typ) {
		return TypeSwitchUnion
	}
	if is_type_any(typ) {
		return TypeSwitchAny
	}
	return TypeSwitchInvalid
}

func check_switch_stmt(c *CheckerContext, node *Ast, mod_flags uint32) {
	ss := &node.SwitchStmt
	x := Operand{}

	mod_flags |= StmtFlag_BreakAllowed | StmtFlag_FallthroughAllowed
	check_open_scope(c, node)
	defer check_close_scope(c)

	check_label(c, ss.label, node)

	if ss.init != nil {
		check_stmt(c, ss.init, 0)
	}

	if ss.tag != nil {
		check_expr(c, &x, ss.tag)
		check_assignment(c, &x, nil, make_string_c("switch expression"))
		if x.Type == nil {
			return
		}
	} else {
		x.Mode = AddressingConstant
		x.Type = t_bool
		x.Value = exact_value_bool(true)
		token := Token{}
		token.Pos = ast_token(ss.body).Pos
		token.String = make_string_c("true")
		x.Expr = alloc_ast_node(nil, AstIdent)
		x.Expr.Ident.Token = token
	}

	var first_default *Ast = nil
	bs := &ss.body.BlockStmt

	for _, stmt := range bs.stmts {
		var default_stmt *Ast = nil
		if stmt.Kind == AstCaseClause {
			cc := &stmt.CaseClause
			if len(cc.list) == 0 {
				default_stmt = stmt
			}
		} else {
			error(stmt, "Invalid AST - expected case clause")
		}
		if default_stmt != nil {
			if first_default != nil {
				pos := ast_token(first_default).Pos
				error(stmt,
					"multiple default clauses\n"+
						"\tfirst at %s", token_pos_to_string(pos))
			} else {
				first_default = default_stmt
			}
		}
	}

	is_partial := ss.partial
	if is_partial {
		if !is_type_enum(x.Type) {
			error(x.Expr, "#partial switch statement can be only used with an enum type")
		}
	}

	seen := SeenMap{}
	defer seen_map_destroy(&seen)

	for _, stmt := range bs.stmts {
		if stmt.Kind != AstCaseClause {
			continue
		}
		cc := &stmt.CaseClause
		for _, expr := range cc.list {
			expr = unparen_expr(expr)
			if is_ast_range(expr) {
				be := &expr.BinaryExpr
				lhs := Operand{}
				rhs := Operand{}

				check_expr_with_type_hint(c, &lhs, be.left, x.Type)
				if x.Mode == AddressingInvalid {
					continue
				}
				if lhs.Mode == AddressingInvalid {
					continue
				}
				check_expr_with_type_hint(c, &rhs, be.right, x.Type)
				if rhs.Mode == AddressingInvalid {
					continue
				}

				if !is_type_ordered(x.Type) {
					str := type_to_string(x.Type)
					error(expr, "Unordered type '%s', is invalid for an interval expression", str)
					gb_string_free(str)
					continue
				}

				upper_op := TokenInvalid
				switch be.op.Kind {
				case TokenEllipsis:
					upper_op = TokenLtEq
				case TokenRangeFull:
					upper_op = TokenLtEq
				case TokenRangeHalf:
					upper_op = TokenLt
				default:
					gb_assert_handler("Panic", 0, "cmd_check_stmt_switch.go", 0, "Invalid range operator")
				}

				a := lhs
				b := rhs
				check_comparison(c, expr, &a, &x, TokenLtEq)
				if a.Mode == AddressingInvalid {
					continue
				}
				check_comparison(c, expr, &b, &x, upper_op)
				if b.Mode == AddressingInvalid {
					continue
				}
				a1 := lhs
				b1 := rhs
				check_comparison(c, expr, &a1, &b1, TokenLtEq)

				add_to_seen_map(c, &seen, upper_op, x, lhs, rhs)

				if is_type_string16(x.Type) {
					add_package_dependency(c, "runtime", "string16_le", true)
					add_package_dependency(c, "runtime", "string16_lt", true)
				} else if is_type_string(x.Type) {
					add_package_dependency(c, "runtime", "string_le", true)
					add_package_dependency(c, "runtime", "string_lt", true)
				}
			} else {
				y := Operand{}
				if is_type_typeid(x.Type) {
					check_expr_or_type(c, &y, expr, x.Type)
				} else {
					check_expr_with_type_hint(c, &y, expr, x.Type)
				}
				if x.Mode == AddressingInvalid ||
					y.Mode == AddressingInvalid {
					continue
				}
				if y.Mode == AddressingType {
					t := y.Type
					if t == nil || t == t_invalid || is_type_polymorphic(t) {
						error(y.Expr, "Invalid type for case clause")
						continue
					}
					t = default_type(t)
					add_type_info_type(c, t)
				} else {
					convert_to_typed(c, &y, x.Type)
					if y.Mode == AddressingInvalid {
						continue
					}
					z := y
					check_comparison(c, expr, &z, &x, TokenCmpEq)
					if z.Mode == AddressingInvalid {
						continue
					}
					if y.Mode != AddressingConstant {
						continue
					}
					update_untyped_expr_type(c, z.Expr, x.Type, !is_type_untyped(x.Type))
					add_to_seen_map(c, &seen, y)
				}
			}
		}
		check_open_scope(c, stmt)
		check_stmt_list(c, cc.stmts, mod_flags)
		check_close_scope(c)
	}

	if !is_partial && is_type_enum(x.Type) {
		et := base_type(x.Type)
		gb_assert_handler("Assertion Failure", "is_type_enum(et)", "cmd_check_stmt_switch.go", 0, "")
		fields := et.Enum.Fields
		unhandled := make([]*Entity, 0, len(fields))
		for _, f := range fields {
			if f.Kind != EntityConstant {
				continue
			}
			v := f.Constant.Value
			found := seen_map_get(&seen, hash_exact_value(v))
			if !found {
				unhandled = append(unhandled, f)
			}
		}
		if len(unhandled) > 0 {
			begin_error_block()
			if len(unhandled) == 1 {
				error_no_newline(node, "Unhandled switch case: %s", goStr(unhandled[0].Token.String))
			} else {
				error(node, "Unhandled switch cases:")
				for _, f := range unhandled {
					error_line("\t%s\n", goStr(f.Token.String))
				}
			}
			error_line("\tSuggestion: Was '#partial switch' wanted?\n")
			end_error_block()
		}
	}

	if buildContext.StrictStyle {
		stok := ss.token
		for _, stmt := range bs.stmts {
			if stmt.Kind != AstCaseClause {
				continue
			}
			ctok := stmt.CaseClause.token
			if ctok.Pos.Column > stok.Pos.Column {
				error(ctok, "With '-strict-style', 'case' statements must share the same column as the 'switch' token")
			}
		}
	}
}

func check_type_switch_stmt(c *CheckerContext, node *Ast, mod_flags uint32) {
	ss := &node.TypeSwitchStmt
	x := Operand{}

	mod_flags |= StmtFlag_BreakAllowed | StmtFlag_TypeSwitch
	check_open_scope(c, node)
	defer check_close_scope(c)

	check_label(c, ss.label, node)

	if ss.tag.Kind != AstAssignStmt {
		error(ss.tag, "Expected an 'in' assignment for this type switch statement")
		return
	}

	as := &ss.tag.AssignStmt
	as_token := ast_token(ss.tag)
	if len(as.lhs) != 1 {
		syntax_error(as_token, "Expected 1 name before 'in'")
		return
	}
	if len(as.rhs) != 1 {
		syntax_error(as_token, "Expected 1 expression after 'in'")
		return
	}

	is_addressed := false
	lhs := as.lhs[0]
	rhs := as.rhs[0]
	if lhs.Kind == AstUnaryExpr && lhs.UnaryExpr.op.Kind == TokenAnd {
		is_addressed = true
		lhs = lhs.UnaryExpr.expr
	}

	check_expr(c, &x, rhs)
	check_assignment(c, &x, nil, make_string_c("type switch expression"))
	add_type_info_type(c, x.Type)

	switch_kind := check_valid_type_switch_type(x.Type)
	if switch_kind == TypeSwitchInvalid {
		str := type_to_string(x.Type)
		error(x.Expr, "Invalid type for this type switch expression, got '%s'", str)
		gb_string_free(str)
		return
	}

	is_partial := ss.partial
	if is_partial {
		if switch_kind != TypeSwitchUnion {
			error(node, "#partial switch statement may only be used with a union")
		}
	}

	var first_default *Ast = nil
	bs := &ss.body.BlockStmt
	for _, stmt := range bs.stmts {
		var default_stmt *Ast = nil
		if stmt.Kind == AstCaseClause {
			cc := &stmt.CaseClause
			if len(cc.list) == 0 {
				default_stmt = stmt
			}
		} else {
			error(stmt, "Invalid AST - expected case clause")
		}
		if default_stmt != nil {
			if first_default != nil {
				pos := ast_token(first_default).Pos
				error(stmt,
					"Multiple default clauses\n"+
						"\tfirst at %s", token_pos_to_string(pos))
			} else {
				first_default = default_stmt
			}
		}
	}

	if lhs.Kind != AstIdent {
		error(rhs, "Expected an identifier, got '%s'", ast_kind_string(rhs.Kind))
		return
	}

	var nil_seen *Ast = nil
	seen := TypeSet{}
	type_set_init(&seen)
	defer type_set_destroy(&seen)

	for _, stmt := range bs.stmts {
		if stmt.Kind != AstCaseClause {
			continue
		}
		cc := &stmt.CaseClause
		saw_nil := false
		bt := base_type(type_deref(x.Type, false))
		var case_type *Type = nil

		for _, type_expr := range cc.list {
			if type_expr != nil {
				y := Operand{}
				check_expr_or_type(c, &y, type_expr)
				if is_operand_nil(y) {
					if !type_has_nil(type_deref(x.Type, false)) {
						error(type_expr, "'nil' case is not allowed for the type '%s'", type_to_string(type_deref(x.Type, false)))
						continue
					}
					saw_nil = true
					if nil_seen != nil {
						begin_error_block()
						error(type_expr, "'nil' case has already been handled previously")
						error_line("\t 'nil' was already previously seen at %s", token_pos_to_string(ast_token(nil_seen).Pos))
						end_error_block()
					} else {
						nil_seen = type_expr
					}
					case_type = y.Type
					continue
				}
				if y.Mode != AddressingType {
					str := expr_to_string(type_expr)
					error(type_expr, "Expected a type as a case, got %s", str)
					gb_string_free(str)
					continue
				}
				if switch_kind == TypeSwitchUnion {
					gb_assert_handler("Assertion Failure", "is_type_union(bt)", "cmd_check_stmt_switch.go", 0, "")
					tag_type_found := false
					for _, vt := range bt.Union.Variants {
						if are_types_identical(vt, y.Type) {
							tag_type_found = true
							break
						}
					}
					if !tag_type_found {
						type_str := type_to_string(y.Type)
						error(y.Expr, "Unknown variant type, got '%s'", type_str)
						gb_string_free(type_str)
						continue
					}
					case_type = y.Type
					add_type_info_type(c, y.Type)
				} else if switch_kind == TypeSwitchAny {
					case_type = y.Type
					add_type_info_type(c, y.Type)
				} else {
					gb_assert_handler("Panic", 0, "cmd_check_stmt_switch.go", 0, "Unknown type to type switch statement")
				}
				if type_set_update(&seen, y.Type) {
					pos := cc.token.Pos
					expr_str := expr_to_string(y.Expr)
					error(y.Expr,
						"Duplicate type case '%s'\n"+
							"\tprevious type case at %s",
						expr_str,
						token_pos_to_string(pos))
					gb_string_free(expr_str)
					break
				}
			}
		}

		is_reference := is_addressed
		if len(cc.list) > 1 || saw_nil {
			case_type = nil
		}
		if case_type == nil {
			case_type = type_deref(x.Type, false)
		}
		if switch_kind == TypeSwitchAny {
			if !is_type_untyped(case_type) {
				add_type_info_type(c, case_type)
			}
		}

		check_open_scope(c, stmt)
		{
			tag_var := alloc_entity_variable(c.scope, lhs.Ident.Token, case_type, EntityStateResolved)
			tag_var.Flags |= EntityFlag_Used
			tag_var.Flags |= EntityFlag_SwitchValue
			if !is_reference {
				tag_var.Flags |= EntityFlag_Value
			}
			add_entity(c, c.scope, lhs, tag_var)
			add_entity_use(c, lhs, tag_var)
			add_implicit_entity(c, stmt, tag_var)
		}
		check_stmt_list(c, cc.stmts, mod_flags)
		check_close_scope(c)
	}

	if !is_partial && is_type_union(type_deref(x.Type, false)) {
		ut := base_type(type_deref(x.Type, false))
		gb_assert_handler("Assertion Failure", "is_type_union(ut)", "cmd_check_stmt_switch.go", 0, "")
		variants := ut.Union.Variants
		unhandled := make([]*Type, 0, len(variants))
		for _, t := range variants {
			if !type_set_exists(&seen, t) {
				unhandled = append(unhandled, t)
			}
		}
		if len(unhandled) > 0 {
			begin_error_block()
			if len(unhandled) == 1 {
				s := type_to_string(unhandled[0])
				error_no_newline(node, "Unhandled switch case: %s", s)
				gb_string_free(s)
			} else {
				error_no_newline(node, "Unhandled switch cases:\n")
				for _, t := range unhandled {
					s := type_to_string(t)
					error_line("\t%s\n", s)
					gb_string_free(s)
				}
			}
			error_line("\n")
			error_line("\tSuggestion: Was '#partial switch' wanted?\n")
			end_error_block()
		}
	}

	if buildContext.StrictStyle {
		stok := ss.token
		for _, stmt := range bs.stmts {
			if stmt.Kind != AstCaseClause {
				continue
			}
			ctok := stmt.CaseClause.token
			if ctok.Pos.Column > stok.Pos.Column {
				error(ctok, "With '-strict-style', 'case' statements must share the same column as the 'switch' token")
			}
		}
	}
}
