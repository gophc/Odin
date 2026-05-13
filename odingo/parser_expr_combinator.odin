package odingo

// =============================================================================
// External function declarations (defined elsewhere in odingo package)
// =============================================================================

unparen_expr              :: proc(node: ^Ast) -> ^Ast
parse_expr                :: proc(f: ^AstFile, lhs: bool) -> ^Ast
parse_ident               :: proc(f: ^AstFile, allow_poly_names: bool = false) -> ^Ast
parse_value               :: proc(f: ^AstFile) -> ^Ast
parse_type                :: proc(f: ^AstFile) -> ^Ast
parse_operand             :: proc(f: ^AstFile, lhs: bool) -> ^Ast
parse_literal_value       :: proc(f: ^AstFile, operand: ^Ast) -> ^Ast
allow_field_separator     :: proc(f: ^AstFile) -> bool
file_allow_newline        :: proc(f: ^AstFile) -> bool
token_is_newline          :: proc(tok: Token) -> bool
expr_to_string            :: proc(expr: ^Ast) -> string
ast_token                 :: proc(node: ^Ast) -> Token
error_line                :: proc(format: string, args: ..any)
begin_error_block         :: proc()
end_error_block           :: proc()

// =============================================================================
// is_literal_type
// =============================================================================

is_literal_type :: proc(node: ^Ast) -> bool {
	node = unparen_expr(node)
	if node == nil {
		return false
	}
	#partial switch node.kind {
	case .BadExpr:
		return true
	case .Ident:
		return true
	case .SelectorExpr:
		return true
	case .ArrayType:
		return true
	case .StructType:
		return true
	case .UnionType:
		return true
	case .EnumType:
		return true
	case .FixedCapacityDynamicArrayType:
		return true
	case .DynamicArrayType:
		return true
	case .MapType:
		return true
	case .BitSetType:
		return true
	case .MatrixType:
		return true
	case .CallExpr:
		return true
	case .MultiPointerType:
		return true
	}
	return false
}

// =============================================================================
// parse_check_or_return
// =============================================================================

parse_check_or_return :: proc(operand: ^Ast, msg: string) {
	if operand == nil {
		return
	}
	#partial switch operand.kind {
	case .OrReturnExpr:
		syntax_error_with_verbose(operand, "'or_return' use within %s is not wrapped in parentheses (...)", msg)
	case .OrBranchExpr:
		tok_str := operand.OrBranchExpr.token.string
		syntax_error_with_verbose(operand, "'%s' use within %s is not wrapped in parentheses (...)", tok_str, msg)
	}
}

// =============================================================================
// parse_call_expr
// =============================================================================

parse_call_expr :: proc(f: ^AstFile, operand: ^Ast) -> ^Ast {
	allocator := ast_allocator(f)
	args := make([dynamic]^Ast, allocator)
	open_paren, close_paren: Token
	ellipsis: Token
	prev_expr_level := f.expr_level
	prev_allow_newline := f.allow_newline
	f.expr_level = 0
	f.allow_newline = file_allow_newline(f)
	open_paren = expect_token(f, .OpenParen)
	seen_ellipsis := false
	for f.curr_token.kind != .CloseParen &&
	    f.curr_token.kind != .EOF {
		if f.curr_token.kind == .Comma {
			syntax_error(f.curr_token, "Expected an expression not ,")
		} else if f.curr_token.kind == .Eq {
			syntax_error(f.curr_token, "Expected an expression not =")
		}
		prefix_ellipsis := false
		if f.curr_token.kind == .Ellipsis {
			prefix_ellipsis = true
			ellipsis = expect_token(f, .Ellipsis)
		}
		arg := parse_expr(f, false)
		if f.curr_token.kind == .Eq {
			eq := expect_token(f, .Eq)
			if prefix_ellipsis {
				syntax_error(ellipsis, "'..' must be applied to value rather than the field name")
			}
			value := parse_value(f)
			arg = ast_field_value(f, arg, value, eq)
		} else if seen_ellipsis {
			syntax_error(arg, "Positional arguments are not allowed after '..'")
		}
		append(&args, arg)
		if ellipsis.pos.line != 0 {
			seen_ellipsis = true
		}
		if !allow_field_separator(f) {
			break
		}
	}
	f.allow_newline = prev_allow_newline
	f.expr_level = prev_expr_level
	close_paren = expect_closing(f, .CloseParen, "argument list")
	call := ast_call_expr(f, operand, args[:], open_paren, close_paren, ellipsis)
	o := unparen_expr(operand)
	if o != nil && o.kind == .SelectorExpr && o.SelectorExpr.token.kind == .ArrowRight {
		return ast_selector_call_expr(f, o.SelectorExpr.token, o, call)
	}
	return call
}

// =============================================================================
// parse_atom_expr
// =============================================================================

parse_atom_expr :: proc(f: ^AstFile, operand: ^Ast, lhs: bool) -> ^Ast {
	operand := operand
	lhs := lhs
	if operand == nil {
		if f.allow_type {
			return nil
		}
		begin := f.curr_token
		syntax_error(begin, "Expected an operand")
		fix_advance_to_next_stmt(f)
		operand = ast_bad_expr(f, begin, f.curr_token)
	}
	loop := true
	for loop {
		switch f.curr_token.kind {
		case .OpenParen:
			parse_check_or_return(operand, "call expression")
			operand = parse_call_expr(f, operand)

		case .Period:
			{
				token := advance_token(f)
				switch f.curr_token.kind {
				case .Ident:
					parse_check_or_return(operand, "selector expression")
					operand = ast_selector_expr(f, token, operand, parse_ident(f))

				case .OpenParen:
					parse_check_or_return(operand, "type assertion")
					open := expect_token(f, .OpenParen)
					type := parse_type(f)
					_ = expect_token(f, .CloseParen)
					operand = ast_type_assertion(f, operand, token, type)

				case .Question:
					parse_check_or_return(operand, ".? based type assertion")
					question := expect_token(f, .Question)
					type := ast_unary_expr(f, question, nil)
					operand = ast_type_assertion(f, operand, token, type)

				case:
					syntax_error(f.curr_token, "Expected a selector")
					advance_token(f)
					operand = ast_bad_expr(f, ast_token(operand), f.curr_token)
				}
			}

		case .ArrowRight:
			parse_check_or_return(operand, "-> based call expression")
			token := advance_token(f)
			operand = ast_selector_expr(f, token, operand, parse_ident(f))

		case .OpenBracket:
			{
				prev_allow_range := f.allow_range
				f.allow_range = false
				defer f.allow_range = prev_allow_range

				open, close, interval: Token
				indices: [2]^Ast
				is_interval := false
				f.expr_level += 1
				open = expect_token(f, .OpenBracket)
				if f.curr_token.kind == .CloseBracket {
					begin_error_block()
					defer end_error_block()
					syntax_error(f.curr_token, "Expected an operand, got ]")
					close = expect_token(f, .CloseBracket)
					if f.allow_type {
						s := expr_to_string(operand)
						error_line("\tSuggestion: If a type was wanted, did you mean '[]%s'?", s)
					}
					operand = ast_index_expr(f, operand, nil, open, close)
					break
				}
				switch f.curr_token.kind {
				case .Ellipsis, .RangeFull, .RangeHalf, .Colon:
					// nothing
				case:
					indices[0] = parse_expr(f, false)
				}
				switch f.curr_token.kind {
				case .Ellipsis, .RangeFull, .RangeHalf:
					syntax_error(f.curr_token, "Expected a colon, not a range")
				case .Comma, .Colon:
					interval = advance_token(f)
					is_interval = true
					if f.curr_token.kind != .CloseBracket &&
					   f.curr_token.kind != .EOF {
						indices[1] = parse_expr(f, false)
					}
				}
				f.expr_level -= 1
				close = expect_token(f, .CloseBracket)
				if is_interval {
					if interval.kind == .Comma {
						if indices[0] == nil || indices[1] == nil {
							syntax_error(open, "Matrix index expressions require both row and column indices")
						}
						parse_check_or_return(operand, "matrix index expression")
						operand = ast_matrix_index_expr(f, operand, open, close, indices[0], indices[1])
					} else {
						parse_check_or_return(operand, "slice expression")
						operand = ast_slice_expr(f, operand, open, close, interval, indices[0], indices[1])
					}
				} else {
					parse_check_or_return(operand, "index expression")
					operand = ast_index_expr(f, operand, indices[0], open, close)
				}
			}

		case .Pointer:
			parse_check_or_return(operand, "dereference")
			operand = ast_deref_expr(f, operand, expect_token(f, .Pointer))

		case .or_return:
			operand = ast_or_return_expr(f, operand, expect_token(f, .or_return))

		case .or_break, .or_continue:
			{
				token := advance_token(f)
				label: ^Ast
				if f.curr_token.kind == .Ident {
					label = parse_ident(f)
				}
				operand = ast_or_branch_expr(f, operand, token, label)
			}

		case .OpenBrace:
			if !lhs && is_literal_type(operand) && f.expr_level >= 0 {
				operand = parse_literal_value(f, operand)
			} else {
				loop = false
			}

		case .Increment, .Decrement:
			if !lhs {
				token := advance_token(f)
				str := token.string
				syntax_error(token, "Postfix '%s' operator is not supported", str)
			} else {
				loop = false
			}

		case:
			loop = false
		}
		lhs = false
	}
	return operand
}

// =============================================================================
// parse_unary_expr
// =============================================================================

parse_unary_expr :: proc(f: ^AstFile, lhs: bool) -> ^Ast {
	switch f.curr_token.kind {
	case .transmute, .cast:
		{
			token := advance_token(f)
			_ = expect_token(f, .OpenParen)
			type := parse_type(f)
			_ = expect_token(f, .CloseParen)
			expr := parse_unary_expr(f, lhs)
			return ast_type_cast(f, token, type, expr)
		}

	case .auto_cast:
		{
			token := advance_token(f)
			expr := parse_unary_expr(f, lhs)
			return ast_auto_cast(f, token, expr)
		}

	case .Add, .Sub, .Xor, .And, .Not, .Mul:
		{
			token := advance_token(f)
			if token.kind == .Not {
				skip_possible_newline(f)
			}
			expr := parse_unary_expr(f, lhs)
			return ast_unary_expr(f, token, expr)
		}

	case .Increment, .Decrement:
		{
			token := advance_token(f)
			str := token.string
			syntax_error(token, "Unary '%s' operator is not supported", str)
			expr := parse_unary_expr(f, lhs)
			return ast_unary_expr(f, token, expr)
		}

	case .Period:
		{
			token := expect_token(f, .Period)
			ident := parse_ident(f)
			return ast_implicit_selector_expr(f, token, ident)
		}
	}
	return parse_atom_expr(f, parse_operand(f, lhs), lhs)
}

// =============================================================================
// is_ast_range
// =============================================================================

is_ast_range :: proc(expr: ^Ast) -> bool {
	if expr == nil {
		return false
	}
	if expr.kind != .BinaryExpr {
		return false
	}
	return is_token_range(expr.BinaryExpr.op.kind)
}

// =============================================================================
// token_precedence
// =============================================================================

token_precedence :: proc(f: ^AstFile, t: TokenKind) -> i32 {
	switch t {
	case .Question, .if_, .when, .or_else:
		return 1

	case .Ellipsis, .RangeFull, .RangeHalf:
		if !f.allow_range {
			return 0
		}
		return 2

	case .CmpOr:
		return 3

	case .CmpAnd:
		return 4

	case .CmpEq, .NotEq, .Lt, .Gt, .LtEq, .GtEq:
		return 5

	case .in, .not_in:
		if f.expr_level < 0 && !f.allow_in_expr {
			return 0
		}
		return 6

	case .Add, .Sub, .Or, .Xor:
		return 6

	case .Mul, .Quo, .Mod, .ModMod, .And, .AndNot, .Shl, .Shr:
		return 7
	}
	return 0
}

// =============================================================================
// parse_binary_expr
// =============================================================================

parse_binary_expr :: proc(f: ^AstFile, lhs: bool, prec_in: i32) -> ^Ast {
	expr := parse_unary_expr(f, lhs)
	loop: for {
		op := f.curr_token
		op_prec := token_precedence(f, op.kind)
		if op_prec < prec_in {
			break
		}
		prev := f.prev_token
		switch op.kind {
		case .if_, .when:
			if prev.pos.line < op.pos.line {
				break loop
			}
		}
		_ = expect_operator(f)
		if op.kind == .Question {
			cond := expr
			x := parse_expr(f, lhs)
			token_c := expect_token(f, .Colon)
			y := parse_expr(f, lhs)
			expr = ast_ternary_if_expr(f, x, cond, y)
		} else if op.kind == .if_ || op.kind == .when {
			x := expr
			cond := parse_expr(f, lhs)
			_ = expect_token(f, .else)
			y := parse_expr(f, lhs)
			switch op.kind {
			case .if_:
				expr = ast_ternary_if_expr(f, x, cond, y)
			case .when:
				expr = ast_ternary_when_expr(f, x, cond, y)
			}
		} else {
			right := parse_binary_expr(f, false, op_prec + 1)
			if right == nil {
				str := op.string
				syntax_error(op, "Expected expression on the right-hand side of the binary operator '%s'", str)
			}
			if op.kind == .or_else {
				expr = ast_or_else_expr(f, expr, op, right)
			} else {
				expr = ast_binary_expr(f, op, expr, right)
			}
		}
		lhs = false
	}
	return expr
}

// =============================================================================
// parse_expr (entry point)
// =============================================================================

parse_expr :: proc(f: ^AstFile, lhs: bool) -> ^Ast {
	return parse_binary_expr(f, lhs, 1)
}

// =============================================================================
// parse_expr_list
// =============================================================================

parse_expr_list :: proc(f: ^AstFile, lhs: bool) -> []^Ast {
	allocator := ast_allocator(f)
	allow_newline := f.allow_newline
	f.allow_newline = file_allow_newline(f)
	list := make([dynamic]^Ast, allocator)
	for {
		e := parse_expr(f, lhs)
		append(&list, e)
		if f.curr_token.kind != .Comma ||
		   f.curr_token.kind == .EOF {
			break
		}
		advance_token(f)
	}
	f.allow_newline = allow_newline
	return list[:]
}

// =============================================================================
// parse_lhs_expr_list
// =============================================================================

parse_lhs_expr_list :: proc(f: ^AstFile) -> []^Ast {
	return parse_expr_list(f, true)
}

// =============================================================================
// parse_rhs_expr_list
// =============================================================================

parse_rhs_expr_list :: proc(f: ^AstFile) -> []^Ast {
	return parse_expr_list(f, false)
}

// =============================================================================
// parse_ident_list
// =============================================================================

parse_ident_list :: proc(f: ^AstFile, allow_poly_names: bool) -> []^Ast {
	allocator := ast_allocator(f)
	list := make([dynamic]^Ast, allocator)
	for {
		append(&list, parse_ident(f, allow_poly_names))
		if f.curr_token.kind != .Comma ||
		   f.curr_token.kind == .EOF {
			break
		}
		advance_token(f)
	}
	return list[:]
}
