package odingo

// ============================================================
// Statement Parsing
// ============================================================
// Translates src/cipp/parser.i.cpp lines 4338-5153
//
// External procedure references (defined in other odingo files):
//   parser_token:    expect_token, allow_token, peek_token, advance_token,
//                    consume_comment_group, consume_line_comment,
//                    is_blank_ident, expect_token_after, expect_semicolon,
//                    expect_closing, expect_closing_brace_of_field_list,
//                    allow_field_separator, skip_possible_newline,
//                    skip_possible_newline_for_literal, fix_advance_to_next_stmt,
//                    token_end_of_line, token_to_string, token_kind_string
//   parser_expr:     parse_expr, parse_simple_stmt, parse_ident, parse_value,
//                    parse_call_expr, parse_rhs_expr_list, parse_lhs_expr_list,
//                    parse_ident_list, parse_stmt_list, parse_block_stmt,
//                    parse_foreign_block, parse_value_decl, parse_inlining_or_tailing_operand,
//                    parse_check_directive_for_statement, convert_stmt_to_body,
//                    convert_stmt_to_expr, ast_on_same_line, expr_to_string
//   parser_error:    syntax_error, error_line, begin_error_block, end_error_block
//   parser_ast_create: ast_block_stmt, ast_if_stmt, ast_when_stmt, ast_return_stmt,
//                    ast_for_stmt, ast_range_stmt, ast_case_clause, ast_switch_stmt,
//                    ast_type_switch_stmt, ast_defer_stmt, ast_import_decl,
//                    ast_foreign_import_decl, ast_attribute, ast_unroll_range_stmt,
//                    ast_bad_stmt, ast_bad_decl, ast_branch_stmt, ast_using_stmt,
//                    ast_empty_stmt, ast_ident, ast_basic_lit, ast_assign_stmt,
//                    ast_field_value, ast_basic_directive, ast_expr_stmt,
//                    ast_token, ast_end_token

// ============================================================
// Import Declaration Kind
// ============================================================

ImportDeclKind :: enum {
	Standard,
	Using,
}

// ============================================================
// Statement Allow Flags (bit set)
// ============================================================

StmtAllowFlag :: enum {
	None  = 0,
	In    = 1,
	Label = 2,
}

// ============================================================
// State Flags for check directives
// ============================================================

StateFlag :: enum {
	bounds_check    = 0,
	no_bounds_check = 1,
	type_assert     = 2,
	no_type_assert  = 3,
}

// ============================================================
// parse_type_or_ident (line 4338)
// ============================================================

parse_type_or_ident :: proc(f: ^AstFile) -> ^Ast {
	prev_allow_type := f.allow_type
	prev_expr_level := f.expr_level
	defer {
		f.allow_type = prev_allow_type
		f.expr_level = prev_expr_level
	}
	f.allow_type = true
	f.expr_level = -1
	lhs := true
	operand := parse_operand(f, lhs)
	type_expr := parse_atom_expr(f, operand, lhs)
	return type_expr
}

// ============================================================
// parse_body (line 4349)
// ============================================================

parse_body :: proc(f: ^AstFile) -> ^Ast {
	stmts: [dynamic]^Ast
	prev_expr_level := f.expr_level
	prev_allow_newline := f.allow_newline
	f.expr_level = 0
	open := expect_token(f, .OpenBrace)
	stmts = parse_stmt_list(f)
	close := expect_token(f, .CloseBrace)
	f.expr_level = prev_expr_level
	f.allow_newline = prev_allow_newline
	return ast_block_stmt(f, stmts, open, close)
}

// ============================================================
// parse_do_body (line 4362)
// ============================================================

parse_do_body :: proc(f: ^AstFile, token: Token, msg: string) -> ^Ast {
	prev_expr_level := f.expr_level
	prev_allow_newline := f.allow_newline
	f.expr_level = 0
	f.allow_newline = false
	body := convert_stmt_to_body(f, parse_stmt(f))
	if build_context.disallow_do {
		syntax_error(body, "'do' has been disallowed")
	} else if token.pos.file_id != 0 && !ast_on_same_line(token, body) {
		syntax_error(body, "The body of a 'do' must be on the same line as %s", msg)
	}
	f.expr_level = prev_expr_level
	f.allow_newline = prev_allow_newline
	return body
}

// ============================================================
// parse_control_statement_semicolon_separator (line 4378)
// ============================================================

parse_control_statement_semicolon_separator :: proc(f: ^AstFile) -> bool {
	tok := peek_token(f)
	if tok.kind != .OpenBrace {
		if f.curr_token.kind == .Semicolon && f.curr_token.string != ";" {
			syntax_error(token_end_of_line(f, f.prev_token), "Expected ';', got newline")
		}
		return allow_token(f, .Semicolon)
	}
	if f.curr_token.string == ";" {
		return allow_token(f, .Semicolon)
	}
	return false
}

// ============================================================
// parse_if_stmt (line 4391)
// ============================================================

parse_if_stmt :: proc(f: ^AstFile) -> ^Ast {
	if f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use an if statement in the file scope")
		return ast_bad_stmt(f, f.curr_token, f.curr_token)
	}
	top_if_stmt: ^Ast
	prev_if_stmt: ^Ast
	curr_if_stmt: ^Ast
	for {
		token := expect_token(f, .If)
		init: ^Ast
		cond: ^Ast
		body: ^Ast
		else_stmt: ^Ast
		prev_level := f.expr_level
		f.expr_level = -1
		prev_allow_in_expr := f.allow_in_expr
		f.allow_in_expr = true
		if allow_token(f, .Semicolon) {
			cond = parse_expr(f, false)
		} else {
			init = parse_simple_stmt(f, .None)
			if parse_control_statement_semicolon_separator(f) {
				cond = parse_expr(f, false)
			} else {
				cond = convert_stmt_to_expr(f, init, "boolean expression")
				init = nil
			}
		}
		f.expr_level = prev_level
		f.allow_in_expr = prev_allow_in_expr
		if cond == nil {
			syntax_error(f.curr_token, "Expected condition for if statement")
		}
		cond_token := token
		if cond != nil {
			cond_token = ast_token(cond)
		}
		if allow_token(f, .Do) {
			body = parse_do_body(f, cond_token, "the if statement")
		} else {
			body = parse_block_stmt(f, false)
		}
		ignore_strict_style := false
		if token.pos.line == ast_end_token(body).pos.line {
			ignore_strict_style = true
		}
		skip_possible_newline_for_literal(f, ignore_strict_style)
		curr_if_stmt = ast_if_stmt(f, token, init, cond, body, nil)
		if top_if_stmt == nil {
			top_if_stmt = curr_if_stmt
		}
		if prev_if_stmt != nil {
			prev_if_stmt.IfStmt.else_stmt = curr_if_stmt
		}
		if f.curr_token.kind != .Else {
			break
		}
		// Consume else token
		else_token := expect_token(f, .Else)
		if f.curr_token.kind == .If {
			prev_if_stmt = curr_if_stmt
			continue
		}
		// Handle else body
		switch f.curr_token.kind {
		case .OpenBrace:
			else_stmt = parse_block_stmt(f, false)
		case .Do:
			expect_token(f, .Do)
			else_stmt = parse_do_body(f, else_token, "'else'")
		case:
			syntax_error(f.curr_token, "Expected if statement block statement")
			else_stmt = ast_bad_stmt(f, f.curr_token, f.tokens[f.curr_token_index + 1])
		}
		curr_if_stmt.IfStmt.else_stmt = else_stmt
		break
	}
	return top_if_stmt
}

// ============================================================
// parse_when_stmt (line 4463)
// ============================================================

parse_when_stmt :: proc(f: ^AstFile) -> ^Ast {
	token := expect_token(f, .When)
	cond: ^Ast
	body: ^Ast
	else_stmt: ^Ast
	prev_level := f.expr_level
	f.expr_level = -1
	prev_allow_in_expr := f.allow_in_expr
	f.allow_in_expr = true
	cond = parse_expr(f, false)
	f.allow_in_expr = prev_allow_in_expr
	f.expr_level = prev_level
	if cond == nil {
		syntax_error(f.curr_token, "Expected condition for when statement")
	}
	was_in_when_statement := f.in_when_statement
	f.in_when_statement = true
	cond_token := token
	if cond != nil {
		cond_token = ast_token(cond)
	}
	if allow_token(f, .Do) {
		body = parse_do_body(f, cond_token, "then when statement")
	} else {
		body = parse_block_stmt(f, true)
	}
	ignore_strict_style := false
	if token.pos.line == ast_end_token(body).pos.line {
		ignore_strict_style = true
	}
	skip_possible_newline_for_literal(f, ignore_strict_style)
	if f.curr_token.kind == .Else {
		else_token := expect_token(f, .Else)
		switch f.curr_token.kind {
		case .When:
			else_stmt = parse_when_stmt(f)
		case .OpenBrace:
			else_stmt = parse_block_stmt(f, true)
		case .Do:
			expect_token(f, .Do)
			else_stmt = parse_do_body(f, else_token, "'else'")
		case:
			syntax_error(f.curr_token, "Expected when statement block statement")
			else_stmt = ast_bad_stmt(f, f.curr_token, f.tokens[f.curr_token_index + 1])
		}
	}
	f.in_when_statement = was_in_when_statement
	return ast_when_stmt(f, token, cond, body, else_stmt)
}

// ============================================================
// parse_return_stmt (line 4512)
// ============================================================

parse_return_stmt :: proc(f: ^AstFile) -> ^Ast {
	token := expect_token(f, .Return)
	if f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use a return statement in the file scope")
		return ast_bad_stmt(f, token, f.curr_token)
	}
	if f.expr_level > 0 {
		syntax_error(f.curr_token, "You cannot use a return statement within an expression")
		return ast_bad_stmt(f, token, f.curr_token)
	}
	results := make([dynamic]^Ast, context.allocator)
	for f.curr_token.kind != .Semicolon && f.curr_token.kind != .CloseBrace {
		arg := parse_expr(f, false)
		append(&results, arg)
		if f.curr_token.kind != .Comma || f.curr_token.kind == .EOF {
			break
		}
		advance_token(f)
	}
	expect_semicolon(f)
	return ast_return_stmt(f, token, results)
}

// ============================================================
// parse_for_stmt (line 4535)
// ============================================================

parse_for_stmt :: proc(f: ^AstFile) -> ^Ast {
	if f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use a for statement in the file scope")
		return ast_bad_stmt(f, f.curr_token, f.curr_token)
	}
	token := expect_token(f, .For)
	init: ^Ast
	cond: ^Ast
	post: ^Ast
	body: ^Ast
	is_range := false
	if f.curr_token.kind != .OpenBrace && f.curr_token.kind != .Do {
		prev_level := f.expr_level
		defer f.expr_level = prev_level
		f.expr_level = -1
		if f.curr_token.kind == .In {
			in_token := expect_token(f, .In)
			syntax_error(in_token, "Prefer 'for _ in' over 'for in'")
			rhs: ^Ast
			prev_allow_range := f.allow_range
			f.allow_range = true
			rhs = parse_expr(f, false)
			f.allow_range = prev_allow_range
			if allow_token(f, .Do) {
				body = parse_do_body(f, token, "the for statement")
			} else {
				body = parse_block_stmt(f, false)
			}
			return ast_range_stmt(f, token, init, nil, in_token, rhs, body)
		}
		if f.curr_token.kind != .Semicolon {
			cond = parse_simple_stmt(f, .In)
			if cond.kind == .AssignStmt && cond.AssignStmt.op.kind == .In {
				is_range = true
			}
		}
		if !is_range && parse_control_statement_semicolon_separator(f) {
			init = cond
			cond = nil
			if f.curr_token.kind == .OpenBrace || f.curr_token.kind == .Do {
				syntax_error(f.curr_token, "Expected ';', followed by a condition expression and post statement, or 'x in y' style loop, got %s", token_kind_string(f.curr_token.kind))
			} else {
				if f.curr_token.kind != .Semicolon {
					if f.curr_token.kind == .Ident {
						next_token := peek_token(f)
						if next_token.kind == .In || next_token.kind == .Comma {
							cond = parse_simple_stmt(f, .In)
							if cond.kind == .AssignStmt && cond.AssignStmt.op.kind == .In {
								is_range = true
							}
							// range_skip: jump to body parsing (at end of outer if)
						} else {
							goto_range_skip := is_range
							if !goto_range_skip {
								cond = parse_simple_stmt(f, .None)
							}
							if f.curr_token.string != ";" {
								syntax_error(f.curr_token, "Expected ';', got %s", token_to_string(f.curr_token))
							} else {
								expect_token(f, .Semicolon)
							}
							if !goto_range_skip &&
							   f.curr_token.kind != .OpenBrace &&
							   f.curr_token.kind != .Do {
								post = parse_simple_stmt(f, .None)
							}
						}
					} else {
						cond = parse_simple_stmt(f, .None)
						if f.curr_token.string != ";" {
							syntax_error(f.curr_token, "Expected ';', got %s", token_to_string(f.curr_token))
						} else {
							expect_token(f, .Semicolon)
						}
						if f.curr_token.kind != .OpenBrace && f.curr_token.kind != .Do {
							post = parse_simple_stmt(f, .None)
						}
					}
				} else {
					if f.curr_token.kind != .OpenBrace && f.curr_token.kind != .Do {
						post = parse_simple_stmt(f, .None)
					}
				}
			}
		}
	}
	// range_skip target: parse body
	if allow_token(f, .Do) {
		body = parse_do_body(f, token, "the for statement")
	} else {
		body = parse_block_stmt(f, false)
	}
	if is_range {
		#assert(cond.kind == .AssignStmt)
		in_token := cond.AssignStmt.op
		vals := cond.AssignStmt.lhs
		rhs: ^Ast
		if len(cond.AssignStmt.rhs) > 0 {
			rhs = cond.AssignStmt.rhs[0]
		}
		return ast_range_stmt(f, token, init, vals, in_token, rhs, body)
	}
	cond = convert_stmt_to_expr(f, cond, "boolean expression")
	if init != nil && cond == nil && post == nil {
		syntax_error(init, "'for init; ; {' without an explicit condition nor post statement is not allowed, please prefer something like 'for init; true; /**/{'")
	}
	return ast_for_stmt(f, token, init, cond, post, body)
}

// ============================================================
// parse_case_clause (line 4627)
// ============================================================

parse_case_clause :: proc(f: ^AstFile, is_type: bool) -> ^Ast {
	token := f.curr_token
	list: [dynamic]^Ast
	expect_token(f, .Case)
	prev_allow_range := f.allow_range
	prev_allow_in_expr := f.allow_in_expr
	f.allow_range = !is_type
	f.allow_in_expr = !is_type
	if f.curr_token.kind != .Colon {
		list = parse_rhs_expr_list(f)
	}
	f.allow_range = prev_allow_range
	f.allow_in_expr = prev_allow_in_expr
	expect_token(f, .Colon)
	stmts := parse_stmt_list(f)
	return ast_case_clause(f, token, list, stmts)
}

// ============================================================
// parse_switch_stmt (line 4644)
// ============================================================

parse_switch_stmt :: proc(f: ^AstFile) -> ^Ast {
	if f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use a switch statement in the file scope")
		return ast_bad_stmt(f, f.curr_token, f.curr_token)
	}
	token := expect_token(f, .Switch)
	init: ^Ast
	tag: ^Ast
	body: ^Ast
	open, close: Token
	is_type_switch := false
	list := make([dynamic]^Ast, context.allocator)
	if f.curr_token.kind != .OpenBrace {
		prev_level := f.expr_level
		defer f.expr_level = prev_level
		f.expr_level = -1
		if f.curr_token.kind == .In {
			in_token := expect_token(f, .In)
			syntax_error(in_token, "Prefer 'switch _ in' over 'switch in'")
			lhs := make([dynamic]^Ast, 0, 1, context.allocator)
			rhs := make([dynamic]^Ast, 0, 1, context.allocator)
			blank_ident := token
			blank_ident.kind = .Ident
			blank_ident.string = "_"
			blank := ast_ident(f, blank_ident)
			append(&lhs, blank)
			append(&rhs, parse_expr(f, true))
			tag = ast_assign_stmt(f, token, lhs, rhs)
			is_type_switch = true
		} else {
			tag = parse_simple_stmt(f, .In)
			if tag.kind == .AssignStmt && tag.AssignStmt.op.kind == .In {
				is_type_switch = true
			} else if parse_control_statement_semicolon_separator(f) {
				init = tag
				tag = nil
				if f.curr_token.kind != .OpenBrace {
					tag = parse_simple_stmt(f, .None)
				}
			}
		}
	}
	skip_possible_newline(f)
	open = expect_token(f, .OpenBrace)
	for f.curr_token.kind == .Case {
		append(&list, parse_case_clause(f, is_type_switch))
	}
	close = expect_token(f, .CloseBrace)
	body = ast_block_stmt(f, list, open, close)
	if is_type_switch {
		return ast_type_switch_stmt(f, token, tag, body)
	}
	tag = convert_stmt_to_expr(f, tag, "switch expression")
	return ast_switch_stmt(f, token, init, tag, body)
}

// ============================================================
// parse_defer_stmt (line 4699)
// ============================================================

parse_defer_stmt :: proc(f: ^AstFile) -> ^Ast {
	if f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use a defer statement in the file scope")
		return ast_bad_stmt(f, f.curr_token, f.curr_token)
	}
	token := expect_token(f, .Defer)
	stmt := parse_stmt(f)
	switch stmt.kind {
	case .EmptyStmt:
		syntax_error(token, "Empty statement after defer (e.g. ';')")
	case .DeferStmt:
		syntax_error(token, "You cannot defer a defer statement")
		stmt = stmt.DeferStmt.stmt
	case .ReturnStmt:
		syntax_error(token, "You cannot defer a return statement")
	}
	return ast_defer_stmt(f, token, stmt)
}

// ============================================================
// parse_import_decl (line 4724)
// ============================================================

parse_import_decl :: proc(f: ^AstFile, kind: ImportDeclKind) -> ^Ast {
	docs := f.lead_comment
	token := expect_token(f, .Import)
	import_name: Token
	switch f.curr_token.kind {
	case .Ident:
		import_name = advance_token(f)
	case:
		import_name.pos = f.curr_token.pos
	}
	file_path := expect_token_after(f, .String, "import")
	s: ^Ast
	if f.curr_proc != nil {
		syntax_error(import_name, "Cannot use 'import' within a procedure. This must be done at the file scope")
		s = ast_bad_decl(f, import_name, file_path)
	} else {
		s = ast_import_decl(f, token, file_path, import_name, docs, f.line_comment)
		append(&f.imports, s)
	}
	if f.in_when_statement {
		syntax_error(import_name, "Cannot use 'import' within a 'when' statement. Prefer using the file suffixes (e.g. foo_windows.odin) or '#+build' tags")
	}
	if kind != .Standard {
		syntax_error(import_name, "'using import' is not allowed, please use the import name explicitly")
	}
	if file_path.string == "\".\"" {
		// intentionally empty: self-import check handled elsewhere
	}
	expect_semicolon(f)
	return s
}

// ============================================================
// parse_foreign_decl (line 4756)
// ============================================================

parse_foreign_decl :: proc(f: ^AstFile) -> ^Ast {
	docs := f.lead_comment
	token := expect_token(f, .Foreign)
	switch f.curr_token.kind {
	case .Ident, .OpenBrace:
		return parse_foreign_block(f, token)
	case .Import:
		import_token := expect_token(f, .Import)
		lib_name: Token
		switch f.curr_token.kind {
		case .Ident:
			lib_name = advance_token(f)
		case:
			lib_name.pos = token.pos
		}
		if is_blank_ident(lib_name) {
			syntax_error(lib_name, "Illegal foreign import name: '_'")
		}
		multiple_filepaths := false
		filepaths: [dynamic]^Ast
		if allow_token(f, .OpenBrace) {
			multiple_filepaths = true
			filepaths = make([dynamic]^Ast, context.allocator)
			for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
				path := parse_expr(f, false)
				append(&filepaths, path)
				if !allow_field_separator(f) {
					break
				}
			}
			expect_closing_brace_of_field_list(f)
		} else {
			filepaths = make([dynamic]^Ast, 0, 1, context.allocator)
			path := expect_token(f, .String)
			lit := ast_basic_lit(f, path)
			append(&filepaths, lit)
		}
		s: ^Ast
		if len(filepaths) == 0 {
			syntax_error(lib_name, "foreign import without any paths")
			s = ast_bad_decl(f, lib_name, f.curr_token)
		} else if f.curr_proc != nil {
			syntax_error(lib_name, "You cannot use foreign import within a procedure. This must be done at the file scope")
			s = ast_bad_decl(f, lib_name, ast_token(filepaths[0]))
		} else {
			s = ast_foreign_import_decl(f, token, filepaths, lib_name, multiple_filepaths, docs, f.line_comment)
		}
		expect_semicolon(f)
		return s
	}
	syntax_error(token, "Invalid foreign declaration")
	return ast_bad_decl(f, token, f.curr_token)
}

// ============================================================
// parse_attribute (line 4814)
// ============================================================

parse_attribute :: proc(f: ^AstFile, token: Token, open_kind: TokenKind, close_kind: TokenKind, docs: ^CommentGroup) -> ^Ast {
	elems: [dynamic]^Ast
	open, close: Token
	if f.curr_token.kind == .Ident {
		elems = make([dynamic]^Ast, 0, 1, context.allocator)
		elem := parse_ident(f)
		append(&elems, elem)
	} else {
		open = expect_token(f, open_kind)
		f.expr_level += 1
		if f.curr_token.kind != close_kind {
			elems = make([dynamic]^Ast, context.allocator)
			for f.curr_token.kind != close_kind && f.curr_token.kind != .EOF {
				elem: ^Ast
				elem = parse_ident(f)
				if f.curr_token.kind == .Eq {
					eq := expect_token(f, .Eq)
					value := parse_value(f)
					elem = ast_field_value(f, elem, value, eq)
				}
				append(&elems, elem)
				if !allow_field_separator(f) {
					break
				}
			}
		}
		f.expr_level -= 1
		close = expect_closing(f, close_kind, "attribute")
	}
	attribute := ast_attribute(f, token, open, close, elems)
	skip_possible_newline(f)
	decl := parse_stmt(f)
	if decl.kind == .ValueDecl {
		if decl.ValueDecl.docs == nil && docs != nil {
			decl.ValueDecl.docs = docs
		}
		append(&decl.ValueDecl.attributes, attribute)
	} else if decl.kind == .ForeignBlockDecl {
		append(&decl.ForeignBlockDecl.attributes, attribute)
	} else if decl.kind == .ForeignImportDecl {
		append(&decl.ForeignImportDecl.attributes, attribute)
	} else if decl.kind == .ImportDecl {
		append(&decl.ImportDecl.attributes, attribute)
	} else {
		syntax_error(decl, "Expected a value or foreign declaration after an attribute, got %s", ast_kind_string(decl.kind))
		return ast_bad_stmt(f, token, f.curr_token)
	}
	return decl
}

// ============================================================
// parse_unrolled_for_loop (line 4865)
// ============================================================

parse_unrolled_for_loop :: proc(f: ^AstFile, unroll_token: Token) -> ^Ast {
	args: [dynamic]^Ast
	if allow_token(f, .OpenParen) {
		f.expr_level += 1
		if f.curr_token.kind == .CloseParen {
			syntax_error(f.curr_token, "#unroll expected at least 1 argument, got 0")
		} else {
			args = make([dynamic]^Ast, context.allocator)
			for f.curr_token.kind != .CloseParen && f.curr_token.kind != .EOF {
				arg: ^Ast
				arg = parse_value(f)
				if f.curr_token.kind == .Eq {
					eq := expect_token(f, .Eq)
					if arg != nil && arg.kind != .Ident {
						syntax_error(arg, "Expected an identifier for 'key=value'")
					}
					value := parse_value(f)
					arg = ast_field_value(f, arg, value, eq)
				}
				append(&args, arg)
				if !allow_field_separator(f) {
					break
				}
			}
		}
		f.expr_level -= 1
		close := expect_closing(f, .CloseParen, "#unroll")
		_ = close
	}
	for_token := expect_token(f, .For)
	init: ^Ast
	val0, val1: ^Ast
	in_token: Token
	expr: ^Ast
	body: ^Ast
	bad_stmt := false
	if f.curr_token.kind != .In {
		idents := parse_ident_list(f, false)
		switch len(idents) {
		case 1:
			val0 = idents[0]
		case 2:
			val0 = idents[0]
			val1 = idents[1]
		case:
			syntax_error(for_token, "Expected either 1 or 2 identifiers")
			bad_stmt = true
		}
	}
	in_token = expect_token(f, .In)
	prev_allow_range := f.allow_range
	prev_level := f.expr_level
	f.allow_range = true
	f.expr_level = -1
	expr = parse_expr(f, false)
	f.expr_level = prev_level
	f.allow_range = prev_allow_range
	if allow_token(f, .Do) {
		body = parse_do_body(f, for_token, "the for statement")
	} else {
		body = parse_block_stmt(f, false)
	}
	if bad_stmt {
		return ast_bad_stmt(f, unroll_token, f.curr_token)
	}
	return ast_unroll_range_stmt(f, unroll_token, init, args[:], for_token, val0, val1, in_token, expr, body)
}

// ============================================================
// parse_stmt (line 4937) - Main statement dispatch
// ============================================================

parse_stmt :: proc(f: ^AstFile) -> ^Ast {
	token := f.curr_token
	s: ^Ast
	switch token.kind {
	case .Context, .Proc, .Ident, .Integer, .Float, .Imag, .Rune, .String,
	     .OpenParen, .Pointer, .Asm, .Add, .Sub, .Xor, .Not, .And, .Mul:
		s = parse_simple_stmt(f, .Label)
		expect_semicolon(f)
		return s

	case .Foreign:
		return parse_foreign_decl(f)

	case .Import:
		return parse_import_decl(f, .Standard)

	case .If:
		return parse_if_stmt(f)
	case .When:
		return parse_when_stmt(f)
	case .For:
		return parse_for_stmt(f)
	case .Switch:
		return parse_switch_stmt(f)
	case .Defer:
		return parse_defer_stmt(f)
	case .Return:
		return parse_return_stmt(f)

	case .Break, .Continue, .Fallthrough:
		branch_token := advance_token(f)
		label: ^Ast
		if branch_token.kind != .Fallthrough && f.curr_token.kind == .Ident {
			label = parse_ident(f)
		}
		s = ast_branch_stmt(f, branch_token, label)
		expect_semicolon(f)
		return s

	case .Using:
		docs := f.lead_comment
		using_token := expect_token(f, .Using)
		if f.curr_token.kind == .Import {
			return parse_import_decl(f, .Using)
		}
		decl: ^Ast
		list := parse_lhs_expr_list(f)
		if len(list) == 0 {
			syntax_error(using_token, "Illegal use of 'using' statement")
			expect_semicolon(f)
			return ast_bad_stmt(f, using_token, f.curr_token)
		}
		if f.curr_token.kind != .Colon {
			expect_semicolon(f)
			return ast_using_stmt(f, using_token, list)
		}
		expect_token_after(f, .Colon, "identifier list")
		decl = parse_value_decl(f, list, docs)
		if decl != nil && decl.kind == .ValueDecl {
			decl.ValueDecl.is_using = true
			return decl
		}
		syntax_error(using_token, "Illegal use of 'using' statement")
		return ast_bad_stmt(f, using_token, f.curr_token)

	case .At:
		docs := f.lead_comment
		at_token := expect_token(f, .At)
		return parse_attribute(f, at_token, .OpenParen, .CloseParen, docs)

	case .Hash:
		s = nil
		hash_token := expect_token(f, .Hash)
		name := expect_token(f, .Ident)
		tag := name.string
		switch tag {
		case "bounds_check":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, .bounds_check)
		case "no_bounds_check":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, .no_bounds_check)
		case "type_assert":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, .type_assert)
		case "no_type_assert":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, .no_type_assert)
		case "partial":
			s = parse_stmt(f)
			switch s.kind {
			case .SwitchStmt:
				if s.SwitchStmt.partial {
					syntax_error(token, "#partial already applied to a switch statement")
				}
				s.SwitchStmt.partial = true
			case .TypeSwitchStmt:
				if s.TypeSwitchStmt.partial {
					syntax_error(token, "#partial already applied to a switch statement")
				}
				s.TypeSwitchStmt.partial = true
			case .EmptyStmt:
				return parse_check_directive_for_statement(s, name, nil)
			case:
				syntax_error(token, "#partial can only be applied to a switch statement")
			}
			return s
		case "assert", "panic":
			t := ast_basic_directive(f, hash_token, name)
			stmt := ast_expr_stmt(f, parse_call_expr(f, t))
			expect_semicolon(f)
			return stmt
		case "force_inline", "force_no_inline", "must_tail":
			expr := parse_inlining_or_tailing_operand(f, name)
			stmt := ast_expr_stmt(f, expr)
			expect_semicolon(f)
			return stmt
		case "unroll":
			return parse_unrolled_for_loop(f, name)
		case "reverse":
			for_stmt := parse_stmt(f)
			if for_stmt.kind == .RangeStmt {
				if for_stmt.RangeStmt.reverse {
					syntax_error(token, "#reverse already applied to a 'for in' statement")
				}
				for_stmt.RangeStmt.reverse = true
			} else {
				syntax_error(token, "#reverse can only be applied to a 'for in' statement")
			}
			return for_stmt
		case "include":
			syntax_error(token, "#include is not a valid import declaration kind. Did you mean 'import'?")
			s = ast_bad_stmt(f, token, f.curr_token)
		case "define":
			s = ast_bad_stmt(f, token, f.curr_token)
			if name.pos.line == f.curr_token.pos.line {
				call_like := false
				macro_expr: ^Ast
				ident := f.curr_token
				if allow_token(f, .Ident) && name.pos.line == f.curr_token.pos.line {
					if f.curr_token.kind == .OpenParen && f.curr_token.pos.column == ident.pos.column + len(ident.string) {
						call_like = true
						_ = parse_call_expr(f, nil)
					}
					if name.pos.line == f.curr_token.pos.line && f.curr_token.kind != .Semicolon {
						macro_expr = parse_expr(f, false)
					}
				}
				begin_error_block()
				defer end_error_block()
				syntax_error(ident, "#define is not a valid declaration, Odin does not have a C-like preprocessor.")
				if macro_expr == nil || call_like {
					error_line("\tNote: Odin does not support macros\n")
				} else {
					expr_str := expr_to_string(macro_expr)
					error_line("\tSuggestion: Did you mean '%s :: %s'?\n", ident.string, expr_str)
					delete(expr_str)
				}
			} else {
				syntax_error(token, "#define is not a valid declaration, Odin does not have a C-like preprocessor.")
			}
		case:
			syntax_error(token, "Unknown tag directive used: '%s'", tag)
			s = ast_bad_stmt(f, token, f.curr_token)
		}
		fix_advance_to_next_stmt(f)
		return s

	case .OpenBrace:
		return parse_block_stmt(f, false)

	case .Semicolon:
		s = ast_empty_stmt(f, token)
		expect_semicolon(f)
		return s

	case .FileTag:
		syntax_error(token, "Lines starting with #+ (file tags) are only allowed before the package line.")
		return ast_bad_stmt(f, token, f.curr_token)
	}

	// Handle dangling else
	switch token.kind {
	case .Else:
		expect_token(f, .Else)
		syntax_error(token, "'else' unattached to an 'if' statement")
		switch f.curr_token.kind {
		case .If:
			return parse_if_stmt(f)
		case .When:
			return parse_when_stmt(f)
		case .OpenBrace:
			return parse_block_stmt(f, true)
		case .Do:
			expect_token(f, .Do)
			stmt := parse_do_body(f, Token{}, "the for statement")
			if build_context.disallow_do {
				syntax_error(stmt, "'do' has been disallowed")
			}
			return stmt
		case:
			fix_advance_to_next_stmt(f)
			return ast_bad_stmt(f, token, f.curr_token)
		}
	}

	syntax_error(token, "Expected a statement, got '%s'", token_kind_string(token.kind))
	fix_advance_to_next_stmt(f)
	return ast_bad_stmt(f, token, f.curr_token)
}
