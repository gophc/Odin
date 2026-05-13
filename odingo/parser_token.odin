package odingo

next_token0 :: proc(f: ^AstFile) -> bool {
	if f.curr_token_index+1 < len(f.tokens) {
		f.curr_token_index += 1
		f.curr_token = f.tokens[f.curr_token_index]
		return true
	}
	syntax_error(f.curr_token, "Token is EOF")
	return false
}

consume_comment :: proc(f: ^AstFile, end_line_: ^int) -> Token {
	tok := f.curr_token
	if tok.kind != .Comment {
		panic("consume_comment called on non-comment token")
	}
	end_line := tok.pos.line
	if len(tok.string) > 1 && tok.string[1] == '*' {
		for i := 2; i < len(tok.string); i += 1 {
			if tok.string[i] == '\n' {
				end_line += 1
			}
		}
	}
	if end_line_ != nil {
		end_line_^ = end_line
	}
	next_token0(f)
	return tok
}

consume_comment_group :: proc(f: ^AstFile, n: int, end_line_: ^int) -> ^CommentGroup {
	allocator := context.allocator
	list := make([dynamic]Token, allocator)
	end_line := f.curr_token.pos.line
	if f.curr_token_index == 1 &&
	   f.prev_token.kind == .Comment &&
	   f.prev_token.pos.line+1 == f.curr_token.pos.line {
		append(&list, f.prev_token)
	}
	for f.curr_token.kind == .Comment &&
	    f.curr_token.pos.line <= end_line+n {
		append(&list, consume_comment(f, &end_line))
	}
	if end_line_ != nil {
		end_line_^ = end_line
	}
	if len(list) > 0 {
		comments := new(CommentGroup, allocator)
		comments.list = list[:]
		append(&f.comments, comments)
		return comments
	}
	return nil
}

consume_comment_groups :: proc(f: ^AstFile, prev: Token) {
	if f.curr_token.kind != .Comment {
		return
	}
	comment: ^CommentGroup = nil
	end_line: int = 0
	if f.curr_token.pos.line == prev.pos.line {
		comment = consume_comment_group(f, 0, &end_line)
		if f.curr_token.pos.line != end_line ||
		   f.curr_token.pos.line == prev.pos.line+1 ||
		   f.curr_token.kind == .EOF {
			f.line_comment = comment
		}
	}
	end_line = -1
	for f.curr_token.kind == .Comment {
		comment = consume_comment_group(f, 1, &end_line)
	}
	if end_line+1 == f.curr_token.pos.line || end_line < 0 {
		f.lead_comment = comment
	}
	if f.curr_token.kind == .Comment {
		panic("f.curr_token.kind != Token.Comment")
	}
}

ignore_newlines :: proc(f: ^AstFile) -> bool {
	return f.expr_level > 0
}

advance_token :: proc(f: ^AstFile) -> Token {
	f.lead_comment = nil
	f.line_comment = nil
	f.prev_token_index = f.curr_token_index
	prev := f.prev_token
	f.prev_token = f.curr_token
	ok := next_token0(f)
	if ok {
		switch f.curr_token.kind {
		case .Comment:
			consume_comment_groups(f, prev)
		case .Semicolon:
			if ignore_newlines(f) && f.curr_token.string == "\n" {
				advance_token(f)
			}
		}
	}
	return prev
}

peek_token :: proc(f: ^AstFile) -> Token {
	for i := f.curr_token_index+1; i < len(f.tokens); i += 1 {
		tok := f.tokens[i]
		if tok.kind == .Comment {
			continue
		}
		return tok
	}
	return {}
}

peek_token_n :: proc(f: ^AstFile, n: int) -> Token {
	found: Token = {}
	count := n
	for i := f.curr_token_index+1; i < len(f.tokens); i += 1 {
		tok := f.tokens[i]
		if tok.kind == .Comment {
			continue
		}
		found = tok
		if count == 0 {
			return found
		}
		count -= 1
	}
	return {}
}

skip_possible_newline :: proc(f: ^AstFile) -> bool {
	if token_is_newline(f.curr_token) {
		advance_token(f)
		return true
	}
	return false
}

skip_possible_newline_for_literal :: proc(f: ^AstFile, ignore_strict_style: bool = false) -> bool {
	curr := f.curr_token
	if token_is_newline(curr) {
		next := peek_token(f)
		if curr.pos.line+1 >= next.pos.line {
			switch next.kind {
			case .OpenBrace:
				fallthrough
			case .else:
				if build_context.strict_style && !ignore_strict_style {
					syntax_error(next, "With '-strict-style' the attached brace style (1TBS) is enforced")
				}
				fallthrough
			case .where:
				advance_token(f)
				return true
			}
		}
	}
	return false
}

token_to_string :: proc(tok: Token) -> string {
	p := token_strings[tok.kind]
	if token_is_newline(tok) {
		p = "newline"
	}
	return p
}

expect_token :: proc(f: ^AstFile, kind: TokenKind) -> Token {
	prev := f.curr_token
	if prev.kind != kind {
		c := token_strings[kind]
		p := token_to_string(prev)
		syntax_error(f.curr_token, "Expected '%v', got '%v'", c, p)
		if kind == .Ident {
			switch prev.kind {
			case .context:
				error_line("\tSuggestion: '%v' is a keyword, would 'ctx' suffice?\n", prev.string)
			case .package:
				error_line("\tSuggestion: '%v' is a keyword, would 'pkg' suffice?\n", prev.string)
			case:
				if token_is_keyword(prev.kind) {
					error_line("\tNote: '%v' is a keyword\n", prev.string)
				}
			}
		}
		if prev.kind == .EOF {
			exit_with_errors()
		}
	}
	advance_token(f)
	return prev
}

expect_token_after :: proc(f: ^AstFile, kind: TokenKind, msg: string) -> Token {
	prev := f.prev_token
	curr := f.curr_token
	if curr.kind != kind {
		p := token_to_string(curr)
		token := f.curr_token
		if token_is_newline(curr) {
			token = curr
			token.pos.column -= 1
			skip_possible_newline(f)
		}
		tk_str := token_strings[kind]
		syntax_error(token, "Expected '%v' after %v, got '%v'", tk_str, msg, p)
	}
	advance_token(f)
	if ast_file_vet_style(f) &&
	   prev.kind == .Comma &&
	   prev.pos.line == curr.pos.line {
		tk_str := token_strings[kind]
		syntax_error(prev, "No need for a trailing comma followed by a %v on the same line", tk_str)
	}
	return curr
}

is_token_range :: proc {
	is_token_range_kind :: proc(kind: TokenKind) -> bool {
		switch kind {
		case .Ellipsis, .RangeFull, .RangeHalf:
			return true
		}
		return false
	},
	is_token_range_token :: proc(tok: Token) -> bool {
		return is_token_range_kind(tok.kind)
	},
}

expect_operator :: proc(f: ^AstFile) -> Token {
	prev := f.curr_token
	if (prev.kind == .in || prev.kind == .not_in) && (f.expr_level >= 0 || f.allow_in_expr) {
		// ok
	} else if prev.kind == .if_ || prev.kind == .when {
		// ok
	} else if prev.kind == .or_else || prev.kind == .or_return ||
	          prev.kind == .or_break || prev.kind == .or_continue {
		// ok
	} else if !is_within_operator_range(prev.kind) {
		p := token_to_string(prev)
		syntax_error(prev, "Expected an operator, got '%v'", p)
	} else if !f.allow_range && is_token_range.is_token_range_token(prev) {
		p := token_to_string(prev)
		syntax_error(prev, "Expected an non-range operator, got '%v'", p)
	}
	if prev.kind == .Ellipsis {
		syntax_error(prev, "'..' for ranges are not allowed, did you mean '..<' or '..='?")
		f.tokens[f.curr_token_index].flags |= {.Replace}
	}
	advance_token(f)
	return prev
}

allow_token :: proc(f: ^AstFile, kind: TokenKind) -> bool {
	prev := f.curr_token
	if prev.kind == kind {
		advance_token(f)
		return true
	}
	return false
}

expect_closing_brace_of_field_list :: proc(f: ^AstFile) -> Token {
	token := f.curr_token
	if allow_token(f, .CloseBrace) {
		return token
	}
	ok := true
	if f.allow_newline {
		ok = !skip_possible_newline(f)
	}
	if ok && allow_token(f, .Semicolon) {
		p := token_to_string(token)
		syntax_error(token_end_of_line(f, f.prev_token), "Expected a comma, got a %v", p)
	}
	return expect_token(f, .CloseBrace)
}

is_blank_ident :: proc {
	is_blank_ident_string :: proc(str: string) -> bool {
		if len(str) == 1 {
			return str[0] == '_'
		}
		return false
	},
	is_blank_ident_token :: proc(tok: Token) -> bool {
		if tok.kind == .Ident {
			return is_blank_ident_string(tok.string)
		}
		return false
	},
	is_blank_ident_ast :: proc(node: ^Ast) -> bool {
		if node.kind == .Ident {
			i := cast(^AstIdent) node
			return is_blank_ident_string(i.token.string)
		}
		return false
	},
}

fix_advance_to_next_stmt :: proc(f: ^AstFile) {
	for {
		t := f.curr_token
		switch t.kind {
		case .EOF, .Semicolon:
			return
		case .package, .foreign, .import, .if_, .for, .when, .return,
		     .switch, .defer, .using, .break, .continue, .fallthrough, .Hash:
			if t.pos == f.fix_prev_pos && f.fix_count < 6 {
				f.fix_count += 1
				return
			}
			if f.fix_prev_pos < t.pos {
				f.fix_prev_pos = t.pos
				f.fix_count = 0
				return
			}
		}
		advance_token(f)
	}
}

expect_closing :: proc(f: ^AstFile, kind: TokenKind, context_str: string) -> Token {
	if f.curr_token.kind != kind &&
	   f.curr_token.kind == .Semicolon &&
	   (f.curr_token.string == "\n" || f.curr_token.kind == .EOF) {
		if f.allow_newline {
			tok := f.prev_token
			tok.pos.column += i32(len(tok.string))
			syntax_error(tok, "Missing ',' before newline in %v", context_str)
		}
		advance_token(f)
	}
	return expect_token(f, kind)
}

assign_removal_flag_to_semicolon :: proc(f: ^AstFile) {
	prev_token := &f.tokens[f.prev_token_index]
	curr_token := &f.tokens[f.curr_token_index]
	if prev_token.kind != .Semicolon {
		panic("prev_token.kind == Token.Semicolon")
	}
	if prev_token.string != ";" {
		return
	}
	ok := false
	if curr_token.pos.line > prev_token.pos.line {
		ok = true
	} else if curr_token.pos.line == prev_token.pos.line {
		switch curr_token.kind {
		case .CloseBrace, .CloseParen, .EOF:
			ok = true
		}
	}
	if !ok {
		return
	}
	if build_context.strict_style || (ast_file_vet_flags(f) & VetFlag_Semicolon) != 0 {
		syntax_error(prev_token^, "Found unneeded semicolon")
	}
	prev_token.flags |= {.Remove}
}

expect_semicolon :: proc(f: ^AstFile) {
	prev_token: Token = {}
	if allow_token(f, .Semicolon) {
		assign_removal_flag_to_semicolon(f)
		return
	}
	switch f.curr_token.kind {
	case .CloseBrace, .CloseParen:
		if f.curr_token.pos.line == f.prev_token.pos.line {
			return
		}
	}
	prev_token = f.prev_token
	if prev_token.kind == .Semicolon {
		assign_removal_flag_to_semicolon(f)
		return
	}
	if f.curr_token.kind == .EOF {
		return
	}
	switch f.curr_token.kind {
	case .EOF:
		return
	}
	if f.curr_token.pos.line == f.prev_token.pos.line {
		p := token_to_string(f.curr_token)
		prev_token.pos = token_pos_end(prev_token)
		syntax_error(prev_token, "Expected ';', got %v", p)
		fix_advance_to_next_stmt(f)
	}
}
