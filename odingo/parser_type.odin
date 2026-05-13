package odingo

import "core:fmt"

// ============================================================
// Data
// ============================================================

ParseFieldPrefixMapping :: struct {
	name:       string,
	token_kind: TokenKind,
	flag:       FieldFlag,
}

parse_field_prefix_mappings := [?]ParseFieldPrefixMapping{
	{"using",        .Using,     .using},
	{"no_alias",     .Hash,      .no_alias},
	{"no_capture",   .Hash,      .no_capture},
	{"c_vararg",     .Hash,      .c_vararg},
	{"const",        .Hash,      .const},
	{"any_int",      .Hash,      .any_int},
	{"subtype",      .Hash,      .subtype},
	{"by_ptr",       .Hash,      .by_ptr},
	{"no_broadcast", .Hash,      .no_broadcast},
}

// ============================================================
// parse_type
// ============================================================

parse_type :: proc(f: ^AstFile) -> ^Ast {
	type := parse_type_or_ident(f)
	if type == nil {
		prev_token := f.curr_token
		token: Token
		if f.curr_token.kind == .OpenBrace {
			token = f.curr_token
		} else {
			token = advance_token(f)
		}
		prev_text := token_text(prev_token)
		if prev_text == "\n" {
			syntax_error(token, "Expected a type, got newline")
		} else {
			syntax_error(token, "Expected a type, got '%s'", prev_text)
		}
		return ast_bad_expr(f, token, f.curr_token)
	} else if type.kind == .ParenExpr && unparen_expr(type) == nil {
		syntax_error(type, "Expected a type within the parentheses")
		return ast_bad_expr(f, type.ParenExpr.open, type.ParenExpr.close)
	}
	return type
}

// ============================================================
// parse_foreign_block_decl
// ============================================================

parse_foreign_block_decl :: proc(f: ^AstFile, decls: ^[dynamic]^Ast) {
	decl := parse_stmt(f)
	switch decl.kind {
	case .EmptyStmt, .BadStmt, .BadDecl:
		return
	case .WhenStmt, .ValueDecl:
		append(decls, decl)
		return
	case:
		syntax_error(decl, "Foreign blocks only allow procedure and variable declarations")
		return
	}
}

// ============================================================
// parse_foreign_block
// ============================================================

parse_foreign_block :: proc(f: ^AstFile, token: Token) -> ^Ast {
	docs := f.lead_comment
	foreign_library: ^Ast
	if f.curr_token.kind == .OpenBrace {
		foreign_library = ast_ident(f, blank_token())
	} else {
		foreign_library = parse_ident(f)
	}
	open, close: Token
	decls := make([dynamic]^Ast, context.allocator)
	prev_in_foreign_block := f.in_foreign_block
	defer f.in_foreign_block = prev_in_foreign_block
	f.in_foreign_block = true
	skip_possible_newline_for_literal(f)
	open = expect_token(f, .OpenBrace)
	for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
		parse_foreign_block_decl(f, &decls)
	}
	close = expect_token(f, .CloseBrace)
	body := ast_block_stmt(f, decls[:], open, close)
	decl := ast_foreign_block_decl(f, token, foreign_library, body, docs)
	expect_semicolon(f)
	return decl
}

// ============================================================
// print_comment_group
// ============================================================

print_comment_group :: proc(group: ^CommentGroup) {
	if group != nil {
		for token in group.list {
			fmt.eprintf("%s\n", token.text)
		}
		fmt.eprintf("\n")
	}
}

// ============================================================
// parse_value_decl
// ============================================================

parse_value_decl :: proc(f: ^AstFile, names: [dynamic]^Ast, docs: ^CommentGroup) -> ^Ast {
	is_mutable := true
	values: [dynamic]^Ast
	type := parse_type_or_ident(f)

	if f.curr_token.kind == .Eq || f.curr_token.kind == .Colon {
		sep: Token
		if !is_mutable {
			sep = expect_token_after(f, .Colon, "type")
		} else {
			sep = advance_token(f)
			is_mutable = sep.kind != .Colon
		}
		values = parse_rhs_expr_list(f)
		if len(values) > len(names) {
			syntax_error(f.curr_token, "Too many values on the right hand side of the declaration")
		} else if len(values) < len(names) && !is_mutable {
			syntax_error(f.curr_token, "All constant declarations must be defined")
		} else if len(values) == 0 {
			syntax_error(f.curr_token, "Expected an expression for this declaration")
		}
	}

	if is_mutable {
		if type == nil && len(values) == 0 {
			syntax_error(f.curr_token, "Missing variable type or initialization")
			return ast_bad_decl(f, f.curr_token, f.curr_token)
		}
	} else {
		if type == nil && len(values) == 0 && len(names) > 0 {
			syntax_error(f.curr_token, "Missing constant value")
			return ast_bad_decl(f, f.curr_token, f.curr_token)
		}
	}

	if values == nil {
		values = make([dynamic]^Ast, context.allocator)
	}
	end_comment := f.lead_comment

	if f.expr_level >= 0 {
		if f.curr_token.kind == .CloseBrace && f.curr_token.pos.line == f.prev_token.pos.line {
			// allow missing semicolon before closing brace on same line
		} else {
			expect_semicolon(f)
		}
	}

	if f.curr_proc == nil {
		if len(values) > 0 && len(names) != len(values) {
			syntax_error(
				values[0],
				"Expected %d expressions on the right hand side, got %d\n" +
				"\tNote: Global declarations do not allow for multi-valued expressions",
				len(names), len(values),
			)
		}
	}

	return ast_value_decl(f, names, type, values[:], is_mutable, docs, end_comment)
}

// ============================================================
// parse_simple_stmt
// ============================================================

parse_simple_stmt :: proc(f: ^AstFile, flags: u32) -> ^Ast {
	token := f.curr_token
	docs := f.lead_comment
	lhs := parse_lhs_expr_list(f)
	token = f.curr_token

	switch token.kind {
	case .Eq, .AddEq, .SubEq, .MulEq, .QuoEq, .ModEq, .ModModEq,
	     .AndEq, .OrEq, .XorEq, .ShlEq, .ShrEq, .AndNotEq,
	     .CmpAndEq, .CmpOrEq:
		if f.curr_proc == nil {
			syntax_error(f.curr_token, "You cannot use a simple statement in the file scope")
			return ast_bad_stmt(f, f.curr_token, f.curr_token)
		}
		advance_token(f)
		rhs := parse_rhs_expr_list(f)
		if len(rhs) == 0 {
			syntax_error(token, "No right-hand side in assignment statement.")
			return ast_bad_stmt(f, token, f.curr_token)
		}
		return ast_assign_stmt(f, token, lhs, rhs)

	case .In:
		if flags & u32(StmtAllowFlag.In) != 0 {
			allow_token(f, .In)
			prev_allow_range := f.allow_range
			f.allow_range = true
			expr := parse_expr(f, true)
			f.allow_range = prev_allow_range
			rhs := make([dynamic]^Ast, 0, 1, context.allocator)
			append(&rhs, expr)
			return ast_assign_stmt(f, token, lhs, rhs[:])
		}

	case .Colon:
		expect_token_after(f, .Colon, "identifier list")
		if flags & u32(StmtAllowFlag.Label) != 0 && len(lhs) == 1 {
			is_partial := false
			is_reverse := false
			partial_token: Token

			if f.curr_token.kind == .Hash {
				name := peek_token(f, 0)
				if name.kind == .Ident && token_text(name) == "partial" &&
				   peek_token(f, 1).kind == .Switch {
					partial_token = expect_token(f, .Hash)
					expect_token(f, .Ident)
					is_partial = true
				} else if name.kind == .Ident && token_text(name) == "reverse" &&
				   peek_token(f, 1).kind == .For {
					partial_token = expect_token(f, .Hash)
					expect_token(f, .Ident)
					is_reverse = true
				}
			}

			switch f.curr_token.kind {
			case .OpenBrace, .If, .For, .Switch:
				name := lhs[0]
				label := ast_label_decl(f, ast_token(name), name)
				stmt := parse_stmt(f)
				switch stmt.kind {
				case .BlockStmt:
					stmt.BlockStmt.label = label
				case .IfStmt:
					stmt.IfStmt.label = label
				case .ForStmt:
					stmt.ForStmt.label = label
				case .RangeStmt:
					stmt.RangeStmt.label = label
				case .SwitchStmt:
					stmt.SwitchStmt.label = label
				case .TypeSwitchStmt:
					stmt.TypeSwitchStmt.label = label
				case:
					syntax_error(token, "Labels can only be applied to a loop or switch statement")
				}

				if is_partial {
					switch stmt.kind {
					case .SwitchStmt:
						stmt.SwitchStmt.partial = true
					case .TypeSwitchStmt:
						stmt.TypeSwitchStmt.partial = true
					case:
						name_text := token_text(ast_token(name))
						syntax_error(partial_token, "Incorrect use of directive, use '%s: #partial switch'", name_text)
					}
				} else if is_reverse {
					switch stmt.kind {
					case .RangeStmt:
						if stmt.RangeStmt.reverse {
							syntax_error(token, "#reverse already applied to a 'for in' statement")
						}
						stmt.RangeStmt.reverse = true
					case:
						syntax_error(token, "#reverse can only be applied to a 'for in' statement")
					}
				}
				return stmt
			}
		}
		return parse_value_decl(f, lhs, docs)
	}

	if len(lhs) > 1 {
		syntax_error(token, "Expected 1 expression")
		return ast_bad_stmt(f, token, f.curr_token)
	}

	switch token.kind {
	case .Increment, .Decrement:
		advance_token(f)
		syntax_error(token, "Postfix '%s' statement is not supported", token_text(token))
	}

	return ast_expr_stmt(f, lhs[0])
}

// ============================================================
// parse_block_stmt
// ============================================================

parse_block_stmt :: proc(f: ^AstFile, is_when: bool) -> ^Ast {
	skip_possible_newline_for_literal(f)
	if !is_when && f.curr_proc == nil {
		syntax_error(f.curr_token, "You cannot use a block statement in the file scope")
		return ast_bad_stmt(f, f.curr_token, f.curr_token)
	}
	return parse_body(f)
}

// ============================================================
// parse_results
// ============================================================

parse_results :: proc(f: ^AstFile, diverging: ^bool) -> ^Ast {
	if !allow_token(f, .ArrowRight) {
		return nil
	}
	if allow_token(f, .Not) {
		if diverging != nil { diverging^ = true }
		return nil
	}
	prev_level := f.expr_level
	defer f.expr_level = prev_level

	if f.curr_token.kind != .OpenParen {
		begin_token := f.curr_token
		empty_names: [dynamic]^Ast
		list := make([dynamic]^Ast, 0, 1, context.allocator)
		type := parse_type(f)
		tag: Token
		append(&list, ast_field(f, empty_names[:], type, nil, 0, tag, nil, nil))
		return ast_field_list(f, begin_token, list[:])
	}

	list: ^Ast
	expect_token(f, .OpenParen)
	list = parse_field_list(f, nil, FieldFlag.Results, .CloseParen, true, false)
	if file_allow_newline(f) {
		skip_possible_newline(f)
	}
	expect_token_after(f, .CloseParen, "parameter list")
	return list
}

// ============================================================
// string_to_calling_convention
// ============================================================

string_to_calling_convention :: proc(s: string) -> ProcCallingConvention {
	switch s {
	case "odin":        return .Odin
	case "contextless": return .Contextless
	case "cdecl":       return .CDecl
	case "c":           return .CDecl
	case "stdcall":     return .StdCall
	case "std":         return .StdCall
	case "fastcall":    return .FastCall
	case "fast":        return .FastCall
	case "none":        return .None
	case "naked":       return .Naked
	case "win64":       return .Win64
	case "sysv":        return .SysV
	case "preserve/none": return .PreserveNone
	case "preserve/most": return .PreserveMost
	case "preserve/all":  return .PreserveAll
	case "system":
		if build_context.metrics.os == .Windows {
			return .StdCall
		}
		return .CDecl
	}
	return .Invalid
}

// ============================================================
// parse_proc_type
// ============================================================

parse_proc_type :: proc(f: ^AstFile, proc_token: Token) -> ^Ast {
	params: ^Ast
	results: ^Ast
	diverging := false
	cc := ProcCallingConvention.Invalid

	if f.curr_token.kind == .String {
		token := expect_token(f, .String)
		c := string_to_calling_convention(string_value_from_token(f, token))
		if c == .Invalid {
			syntax_error(token, "Unknown procedure calling convention: '%s'", token_text(token))
		} else {
			cc = c
		}
	}

	if cc == .Invalid {
		if f.in_foreign_block {
			cc = .ForeignBlockDefault
		} else {
			cc = default_calling_convention()
		}
	}

	expect_token(f, .OpenParen)
	f.expr_level += 1
	params = parse_field_list(f, nil, FieldFlag.Signature, .CloseParen, true, true)
	if file_allow_newline(f) {
		skip_possible_newline(f)
	}
	f.expr_level -= 1
	expect_token_after(f, .CloseParen, "parameter list")

	results = parse_results(f, &diverging)

	tags: u64
	is_generic := false

	if params != nil {
		check_generic: for param in params.FieldList.list {
			assert(param.kind == .Field)
			field := param.Field
			if field.type != nil {
				if field.type.kind == .PolyType {
					is_generic = true
					break check_generic
				}
				for name in field.names {
					if name.kind == .PolyType {
						is_generic = true
						break check_generic
					}
				}
			}
		}
	}

	return ast_proc_type(f, proc_token, params, results, tags, cc, is_generic, diverging)
}

// ============================================================
// parse_var_type
// ============================================================

parse_var_type :: proc(f: ^AstFile, allow_ellipsis: bool, allow_typeid_token: bool) -> ^Ast {
	if allow_ellipsis && f.curr_token.kind == .Ellipsis {
		tok := advance_token(f)
		type := parse_type_or_ident(f)
		if type == nil {
			syntax_error(tok, "variadic field missing type after '..'")
			type = ast_bad_expr(f, tok, f.curr_token)
		}
		return ast_ellipsis(f, tok, type)
	}

	type: ^Ast
	if allow_typeid_token && f.curr_token.kind == .Typeid {
		token := expect_token(f, .Typeid)
		specialization: ^Ast
		if allow_token(f, .Quo) {
			specialization = parse_type(f)
		}
		type = ast_typeid_type(f, token, specialization)
	} else {
		type = parse_type(f)
	}
	return type
}

// ============================================================
// is_token_field_prefix
// ============================================================

is_token_field_prefix :: proc(f: ^AstFile) -> FieldFlag {
	switch f.curr_token.kind {
	case .EOF:
		return .Invalid
	case .Using:
		return .using
	case .Hash:
		advance_token(f)
		switch f.curr_token.kind {
		case .Ident:
			for mapping in parse_field_prefix_mappings {
				if mapping.token_kind == .Hash {
					if f.curr_token.text == mapping.name {
						return mapping.flag
					}
				}
			}
		case:
		}
		return .Unknown
	}
	return .Invalid
}

// ============================================================
// parse_field_prefixes
// ============================================================

parse_field_prefixes :: proc(f: ^AstFile) -> u32 {
	counts: [len(parse_field_prefix_mappings)]int
	for {
		flag := is_token_field_prefix(f)
		if flag & .Invalid != 0 {
			break
		}
		if flag & .Unknown != 0 {
			syntax_error(f.curr_token, "Unknown prefix kind '#%s'", f.curr_token.text)
			advance_token(f)
			continue
		}
		for mapping, i in parse_field_prefix_mappings {
			if mapping.flag == flag {
				counts[i] += 1
				advance_token(f)
				break
			}
		}
	}

	field_flags: u32
	for mapping, i in parse_field_prefix_mappings {
		if counts[i] > 0 {
			field_flags |= u32(mapping.flag)
			if counts[i] != 1 {
				prefix := ""
				if mapping.token_kind == .Hash {
					prefix = "#"
				}
				syntax_error(f.curr_token, "Multiple '%s%s' in this field list", prefix, mapping.name)
			}
		}
	}
	return field_flags
}

// ============================================================
// check_field_prefixes
// ============================================================

check_field_prefixes :: proc(f: ^AstFile, name_count: int, allowed_flags: u32, set_flags: u32) -> u32 {
	for mapping in parse_field_prefix_mappings {
		err := false
		flag := u32(mapping.flag)
		if set_flags & flag != 0 {
			if allowed_flags & flag == 0 {
				err = true
			} else if name_count > 0 {
				if mapping.flag == .using {
					err = true
				}
			} else if name_count == 0 {
				// anonymous fields can use this flag
			}
		}
		if err {
			prefix := ""
			if mapping.token_kind == .Hash {
				prefix = "#"
			}
			syntax_error(f.curr_token, "'%s%s' is not allowed in this field list", prefix, mapping.name)
		}
	}
	return set_flags
}
