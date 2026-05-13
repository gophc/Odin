package odingo

import "core:unicode/utf8"

// =============================================================================
// Forward declarations: procs defined in other translation units
// =============================================================================

// parse_expr is defined in parser_expr.odin
// parse_proc_type is defined in parser_type.odin
// parse_stmt_list is defined in parser_stmt.odin
// parse_stmt is defined in parser_stmt.odin
// parse_body is defined in parser_stmt.odin
// parse_do_body is defined in parser_stmt.odin
// parse_block_stmt is defined in parser_stmt.odin
// parse_type_or_ident is defined in parser_type.odin
// parse_lhs_expr_list is defined in parser_expr.odin
// parse_rhs_expr_list is defined in parser_expr.odin
// parse_simple_stmt is defined in parser_stmt.odin
// parse_type is defined in parser_type.odin
// parse_call_expr is defined in parser_expr.odin
// parse_struct_field_list is defined in parser_field.odin
// parse_field_list is defined in parser_field.odin
// parse_unary_expr is defined in parser_expr.odin

// =============================================================================
// Forward declarations for helpers used in this file
// =============================================================================

expect_token :: proc(f: ^AstFile, kind: TokenKind) -> Token {
	// Defined in parser_token.odin or similar
	// Consumes current token if it matches `kind`, otherwise reports syntax error
	if f.curr_token.kind == kind {
		return advance_token(f)
	}
	syntax_error(f.curr_token, "Expected %s, got %s", token_kind_string(kind), token_kind_string(f.curr_token.kind))
	return f.curr_token
}

expect_token_after :: proc(f: ^AstFile, kind: TokenKind, context_name: string) -> Token {
	// Defined in parser_token.odin
	// Advances past the current token, then expects `kind`
	advance_token(f)
	if f.curr_token.kind == kind {
		return advance_token(f)
	}
	syntax_error(f.curr_token, "Expected %s after %s, got %s", token_kind_string(kind), context_name, token_kind_string(f.curr_token.kind))
	return f.curr_token
}

expect_closing :: proc(f: ^AstFile, kind: TokenKind, context_name: string) -> Token {
	// Defined in parser_token.odin
	if f.curr_token.kind == kind {
		return advance_token(f)
	}
	syntax_error(f.curr_token, "Expected closing %s for %s, got %s", token_kind_string(kind), context_name, token_kind_string(f.curr_token.kind))
	return f.curr_token
}

expect_closing_brace_of_field_list :: proc(f: ^AstFile) -> Token {
	return expect_closing(f, .CloseBrace, "field list")
}

allow_token :: proc(f: ^AstFile, kind: TokenKind) -> bool {
	// Defined in parser_token.odin
	if f.curr_token.kind == kind {
		advance_token(f)
		return true
	}
	return false
}

allow_field_separator :: proc(f: ^AstFile) -> bool {
	// Defined in parser_token.odin
	// Allows comma, semicolon, or newline as field separators
	if f.curr_token.kind == .Comma || f.curr_token.kind == .Semicolon {
		advance_token(f)
		return true
	}
	if token_is_newline(f.curr_token) {
		advance_token(f)
		return true
	}
	return false
}

token_is_newline :: proc(tok: Token) -> bool {
	// Defined in parser_token.odin
	return tok.kind == .Semicolon && tok.string == "\n"
}

token_is_keyword :: proc(tok: Token, kind: TokenKind) -> bool {
	// Defined in parser_token.odin
	return tok.kind == kind
}

token_end_of_line :: proc(f: ^AstFile, tok: Token) -> Token {
	// Defined in parser_token.odin
	// Returns the last token on the same line as `tok`
	return tok
}

token_to_string :: proc(tok: Token) -> string {
	// Defined in parser_token.odin
	return tok.string
}

token_kind_string :: proc(kind: TokenKind) -> string {
	// Placeholder: defined in parser_token.odin
	return "<token>"
}

is_blank_ident :: proc(ast: ^Ast) -> bool {
	// Defined in parser_ast_create.odin
	if ast != nil && ast.kind == .Ident {
		return ast.Ident.token.string == "_"
	}
	return false
}

skip_possible_newline_for_literal :: proc(f: ^AstFile, had_where_clause := false) {
	// Defined in parser_token.odin
	// Skips a newline if it precedes an opening brace for a literal
	_ = had_where_clause
}

ast_token :: proc(ast: ^Ast) -> Token {
	// Defined in parser_ast_create.odin
	// Returns the representative token for an AST node
	if ast == nil {
		return Token{}
	}
	#partial switch ast.kind {
	case .Ident:
		return ast.Ident.token
	case .BadExpr:
		return ast.BadExpr.begin
	case .BasicLit:
		return ast.BasicLit.token
	case .ProcLit:
		return ast.ProcLit.token
	case .BlockStmt:
		return ast.BlockStmt.open
	case .EmptyStmt:
		return ast.EmptyStmt.token
	}
	return Token{}
}

is_ast_stmt :: proc(ast: ^Ast) -> bool {
	// Defined in parser_ast_create.odin
	if ast == nil { return false }
	#partial switch ast.kind {
	case .BadStmt, .EmptyStmt, .ExprStmt, .AssignStmt, .BlockStmt,
	     .IfStmt, .WhenStmt, .ReturnStmt, .ForStmt, .RangeStmt,
	     .UnrollRangeStmt, .SwitchStmt, .TypeSwitchStmt, .CaseClause,
	     .DeferStmt, .BranchStmt, .UsingStmt:
		return true
	}
	return false
}

is_ast_decl :: proc(ast: ^Ast) -> bool {
	// Defined in parser_ast_create.odin
	if ast == nil { return false }
	#partial switch ast.kind {
	case .BadDecl, .ValueDecl, .PackageDecl, .ImportDecl,
	     .ForeignImportDecl, .ForeignBlockDecl:
		return true
	}
	return false
}

ast_allocator :: proc(f: ^AstFile) -> Allocator {
	// Defined in parser_file.odin; returns the AST allocator for the file
	_ = f
	return context.allocator
}

ast_file_vet_flags :: proc(f: ^AstFile) -> u64 {
	// Defined in parser_file.odin
	_ = f
	return 0
}

syntax_error :: proc(pos: union { Token, ^Ast }, fmt: string, args: ..any) {
	// Defined in parser_error.odin
	_ = pos
	_ = fmt
}

syntax_warning :: proc(pos: union { Token, ^Ast }, fmt: string, args: ..any) {
	// Defined in parser_error.odin
	_ = pos
	_ = fmt
}

error_line :: proc(fmt: string, args: ..any) {
	// Defined in parser_error.odin
	_ = fmt
}

expr_to_string :: proc(expr: ^Ast) -> string {
	// Defined in parser_error.odin
	_ = expr
	return ""
}

// =============================================================================
// Global string table for AST kind names (defined in parser_ast_create.odin)
// =============================================================================
ast_strings: [AstKind]string

// =============================================================================
// Bit flag constants (defined elsewhere, declared here for reference)
// =============================================================================

ProcTag_optional_ok             :: u64(1 << 0)
ProcTag_optional_allocator_error :: u64(1 << 1)
ProcTag_require_results         :: u64(1 << 2)
ProcTag_bounds_check            :: u64(1 << 3)
ProcTag_no_bounds_check         :: u64(1 << 4)
ProcTag_type_assert             :: u64(1 << 5)
ProcTag_no_type_assert          :: u64(1 << 6)

StateFlag_bounds_check    :: u16(1 << 0)
StateFlag_no_bounds_check :: u16(1 << 1)
StateFlag_type_assert     :: u16(1 << 2)
StateFlag_no_type_assert   :: u16(1 << 3)

ProcInlining :: enum int {
	None     = 0,
	Inline   = 1,
	No_Inline = 2,
}

ProcTailing :: enum int {
	None      = 0,
	Must_Tail = 1,
}

InlineAsmDialectKind :: enum int {
	Default = 0,
	ATT     = 1,
	Intel   = 2,
}

UnionTypeKind :: enum int {
	Normal     = 0,
	No_Nil     = 1,
	Shared_Nil = 2,
}

VetFlag_Tabs :: u64(1 << 0)

// =============================================================================
// parse_ident — parse an identifier token, optionally allowing $ polymorphic names
// C++ lines 2009-2025
// =============================================================================

parse_ident :: proc(f: ^AstFile, allow_poly_names := false) -> ^Ast {
	token := f.curr_token
	if token.kind == .Ident {
		advance_token(f)
	} else if allow_poly_names && token.kind == .Dollar {
		dollar := expect_token(f, .Dollar)
		name := ast_ident(f, expect_token(f, .Ident))
		if is_blank_ident(name) {
			syntax_error(name, "Invalid polymorphic type definition with a blank identifier")
		}
		return ast_poly_type(f, dollar, name, nil)
	} else {
		// Override token string to "_" for error recovery
		token.string = "_"
		expect_token(f, .Ident)
	}
	return ast_ident(f, token)
}

// =============================================================================
// parse_tag_expr — parse a #identifier tag expression
// C++ lines 2026-2030
// =============================================================================

parse_tag_expr :: proc(f: ^AstFile, expression: ^Ast) -> ^Ast {
	token := expect_token(f, .Hash)
	name := expect_token(f, .Ident)
	return ast_tag_expr(f, token, name, expression)
}

// =============================================================================
// unparen_expr — unwrap parenthesized expressions to get the inner expression
// C++ lines 2031-2041
// =============================================================================

unparen_expr :: proc(node: ^Ast) -> ^Ast {
	n := node
	for n != nil && n.kind == .ParenExpr {
		n = n.ParenExpr.expr
	}
	return n
}

// =============================================================================
// unselector_expr — unwrap selector expression chains to their base
// C++ lines 2042-2051
// =============================================================================

unselector_expr :: proc(node: ^Ast) -> ^Ast {
	n := unparen_expr(node)
	if n == nil {
		return nil
	}
	for n.kind == .SelectorExpr {
		n = n.SelectorExpr.selector
	}
	return n
}

// =============================================================================
// strip_or_return_expr — strip or_return/or_branch/paren wrappers
// C++ lines 2052-2067
// =============================================================================

strip_or_return_expr :: proc(node: ^Ast) -> ^Ast {
	n := node
	for {
		if n == nil {
			return n
		}
		switch n.kind {
		case .OrReturnExpr:
			n = n.OrReturnExpr.expr
		case .OrBranchExpr:
			n = n.OrBranchExpr.expr
		case .ParenExpr:
			n = n.ParenExpr.expr
		case:
			return n
		}
	}
}

// =============================================================================
// parse_element_list — parse elements inside a compound literal { ... }
// C++ lines 2069-2085
// =============================================================================

parse_element_list :: proc(f: ^AstFile) -> [dynamic]^Ast {
	allocator := ast_allocator(f)
	elems := make([dynamic]^Ast, allocator)
	for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
		elem := parse_value(f)
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
	return elems
}

// =============================================================================
// consume_line_comment — consume and clear the current line comment
// C++ lines 2086-2093
// =============================================================================

consume_line_comment :: proc(f: ^AstFile) -> ^CommentGroup {
	comment := f.line_comment
	if f.line_comment == f.lead_comment {
		f.lead_comment = nil
	}
	f.line_comment = nil
	return comment
}

// =============================================================================
// parse_enum_field_list — parse the field list inside an enum { ... }
// C++ lines 2094-2118
// =============================================================================

parse_enum_field_list :: proc(f: ^AstFile) -> [dynamic]^Ast {
	allocator := ast_allocator(f)
	elems := make([dynamic]^Ast, allocator)
	for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
		docs := f.lead_comment
		comment: ^CommentGroup = nil
		parse_enforce_tabs(f)
		name := parse_value(f)
		value: ^Ast = nil
		if f.curr_token.kind == .Eq {
			eq := expect_token(f, .Eq)
			value = parse_value(f)
			_ = eq
		}
		comment = consume_line_comment(f)
		elem := ast_enum_field_value(f, name, value, docs, comment)
		append(&elems, elem)
		if !allow_field_separator(f) {
			break
		}
		if elem.EnumFieldValue.comment == nil {
			elem.EnumFieldValue.comment = consume_line_comment(f)
		}
	}
	return elems
}

// =============================================================================
// parse_literal_value — parse a compound literal value { ... }
// C++ lines 2119-2130
// =============================================================================

parse_literal_value :: proc(f: ^AstFile, type: ^Ast) -> ^Ast {
	elems: [dynamic]^Ast
	open := expect_token(f, .OpenBrace)
	expr_level := f.expr_level
	f.expr_level = 0
	if f.curr_token.kind != .CloseBrace {
		elems = parse_element_list(f)
	}
	f.expr_level = expr_level
	close := expect_closing(f, .CloseBrace, "compound literal")
	return ast_compound_lit(f, type, elems, open, close)
}

// =============================================================================
// parse_value — parse a value (compound literal or expression)
// C++ lines 2131-2141
// =============================================================================

parse_value :: proc(f: ^AstFile) -> ^Ast {
	if f.curr_token.kind == .OpenBrace {
		return parse_literal_value(f, nil)
	}
	prev_allow_range := f.allow_range
	f.allow_range = true
	value := parse_expr(f, false)
	f.allow_range = prev_allow_range
	return value
}

// =============================================================================
// check_proc_add_tag — check for duplicate procedure tags and add if unique
// C++ lines 2143-2148
// =============================================================================

check_proc_add_tag :: proc(f: ^AstFile, tag_expr: ^Ast, tags: ^u64, tag: u64, tag_name: string) {
	if tags^ & tag != 0 {
		syntax_error(tag_expr, "Procedure tag already used: %s", tag_name)
	}
	tags^ |= tag
}

// =============================================================================
// is_foreign_name_valid — validate a foreign name string
// C++ lines 2149-2192
// =============================================================================

is_foreign_name_valid :: proc(name: string) -> bool {
	if len(name) == 0 {
		return false
	}
	offset := 0
	for offset < len(name) {
		r: rune
		remaining := len(name) - offset
		r, width := utf8.decode_rune_in_string(name[offset:])
		if r == utf8.RUNE_ERROR && width == 1 {
			return false
		} else if r == utf8.RUNE_BOM && remaining > 0 {
			return false
		}
		if offset == 0 {
			switch r {
			case '-', '$', '.', '_':
				// Valid first characters
			case:
				if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') {
					return false
				}
			}
		} else {
			switch r {
			case '-', '$', '.', '_':
				// Valid subsequent characters
			case:
				if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
					return false
				}
			}
		}
		offset += width
	}
	return true
}

// =============================================================================
// parse_proc_tags — parse #procedure tags (bounds_check, no_bounds_check, etc.)
// C++ lines 2193-2217
// =============================================================================

parse_proc_tags :: proc(f: ^AstFile, tags: ^u64) {
	assert(tags != nil)
	for f.curr_token.kind == .Hash {
		tag_expr := parse_tag_expr(f, nil)
		assert(tag_expr.kind == .TagExpr)
		te := &tag_expr.TagExpr
		tag_name := te.name.string

		switch tag_name {
		case "optional_ok":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_optional_ok, tag_name)
		case "optional_allocator_error":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_optional_allocator_error, tag_name)
		case "require_results":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_require_results, tag_name)
		case "bounds_check":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_bounds_check, tag_name)
		case "no_bounds_check":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_no_bounds_check, tag_name)
		case "type_assert":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_type_assert, tag_name)
		case "no_type_assert":
			check_proc_add_tag(f, tag_expr, tags, ProcTag_no_type_assert, tag_name)
		case:
			syntax_error(tag_expr, "Unknown procedure type tag #%s", tag_name)
		}
	}
	if tags^ & ProcTag_bounds_check != 0 && tags^ & ProcTag_no_bounds_check != 0 {
		syntax_error(f.curr_token, "You cannot apply both #bounds_check and #no_bounds_check to a procedure")
	}
	if tags^ & ProcTag_type_assert != 0 && tags^ & ProcTag_no_type_assert != 0 {
		syntax_error(f.curr_token, "You cannot apply both #type_assert and #no_type_assert to a procedure")
	}
}

// =============================================================================
// convert_stmt_to_expr — convert a statement to an expression, error if not possible
// C++ lines 2226-2239
// =============================================================================

convert_stmt_to_expr :: proc(f: ^AstFile, statement: ^Ast, kind: string) -> ^Ast {
	if statement == nil {
		return nil
	}
	if statement.kind == .ExprStmt {
		return statement.ExprStmt.expr
	}
	syntax_error(f.curr_token, "Expected '%s', found a simple statement.", kind)
	end := f.curr_token
	// If there is a next token, extend the bad expression range to it
	if len(f.tokens) > f.curr_token_index + 1 {
		end = f.tokens[f.curr_token_index + 1]
	}
	return ast_bad_expr(f, f.curr_token, end)
}

// =============================================================================
// convert_stmt_to_body — convert a statement to a block body
// C++ lines 2240-2254
// =============================================================================

convert_stmt_to_body :: proc(f: ^AstFile, stmt: ^Ast) -> ^Ast {
	if stmt.kind == .BlockStmt {
		syntax_error(stmt, "Expected a normal statement rather than a block statement")
		return stmt
	}
	if stmt.kind == .EmptyStmt {
		syntax_error(stmt, "Expected a non-empty statement")
	}
	assert(is_ast_stmt(stmt) || is_ast_decl(stmt))
	open := ast_token(stmt)
	close := ast_token(stmt)
	allocator := ast_allocator(f)
	stmts := make([dynamic]^Ast, allocator, 1)
	append(&stmts, stmt)
	return ast_block_stmt(f, stmts[:], open, close)
}

// =============================================================================
// check_polymorphic_params_for_type — validate polymorphic parameter consistency
// C++ lines 2255-2274
// =============================================================================

check_polymorphic_params_for_type :: proc(f: ^AstFile, polymorphic_params: ^Ast, token: Token) {
	if polymorphic_params == nil {
		return
	}
	if polymorphic_params.kind != .FieldList {
		return
	}
	fl := &polymorphic_params.FieldList
	assert(polymorphic_params.kind == .FieldList)
	for field in fl.list {
		if field.kind != .Field {
			continue
		}
		for name in field.Field.names {
			if name.kind != field.Field.names[0].kind {
				syntax_error(name, "Mixture of polymorphic names using both $ and not for %s parameters", token.string)
				return
			}
		}
	}
}

// =============================================================================
// ast_on_same_line — check if a token and AST node are on the same source line
// C++ lines 2275-2278
// =============================================================================

ast_on_same_line :: proc(x: Token, y: ^Ast) -> bool {
	y_token := ast_token(y)
	return x.pos.line == y_token.pos.line
}

// =============================================================================
// parse_inlining_or_tailing_operand — parse #force_inline/#force_no_inline/#must_tail
// C++ lines 2279-2324
// =============================================================================

parse_inlining_or_tailing_operand :: proc(f: ^AstFile, token: Token) -> ^Ast {
	expr := parse_unary_expr(f, false)
	e := strip_or_return_expr(expr)
	if e == nil {
		return expr
	}
	if e.kind != .ProcLit && e.kind != .CallExpr {
		syntax_error(expr, "%s must be followed by a procedure literal or call, got %s", token.string, ast_strings[e.kind])
		return ast_bad_expr(f, token, f.curr_token)
	}
	pi: ProcInlining = .None
	pt: ProcTailing   = .None
	if token.kind == .Ident {
		switch token.string {
		case "force_inline":
			pi = .Inline
		case "force_no_inline":
			pi = .No_Inline
		case "must_tail":
			pt = .Must_Tail
		}
	}
	if pi != .None {
		if e.kind == .ProcLit {
			if e.ProcLit.inlining != .None && e.ProcLit.inlining != pi {
				syntax_error(expr, "Cannot apply both '#force_inline' and '#force_no_inline' to a procedure literal")
			}
			e.ProcLit.inlining = pi
		} else if e.kind == .CallExpr {
			if e.CallExpr.inlining != .None && e.CallExpr.inlining != pi {
				syntax_error(expr, "Cannot apply both '#force_inline' and '#force_no_inline' to a procedure call")
			}
			e.CallExpr.inlining = pi
		}
	}
	if pt != .None {
		if e.kind == .ProcLit {
			syntax_error(expr, "'#must_tail' can only be applied to a procedure call, not the procedure literal")
			e.ProcLit.tailing = pt
		} else if e.kind == .CallExpr {
			e.CallExpr.tailing = pt
		}
	}
	return expr
}

// =============================================================================
// parse_check_directive_for_statement — apply #bounds_check/#no_bounds_check etc.
// C++ lines 2325-2394
// =============================================================================

parse_check_directive_for_statement :: proc(s: ^Ast, tag_token: Token, state_flag: u16) -> ^Ast {
	name := tag_token.string
	if s == nil {
		syntax_error(tag_token, "Invalid operand for #%s", name)
		return nil
	}
	if s.kind == .EmptyStmt {
		if s.EmptyStmt.token.string == "\n" {
			syntax_error(tag_token, "#%s cannot be followed by a newline", name)
		} else {
			syntax_error(tag_token, "#%s cannot be applied to an empty statement ';'", name)
		}
	}
	if s.state_flags & state_flag != 0 {
		syntax_error(tag_token, "#%s has been applied multiple times", name)
	}
	s.state_flags |= state_flag

	// Check for conflicting flags
	switch state_flag {
	case StateFlag_bounds_check:
		if s.state_flags & StateFlag_no_bounds_check != 0 {
			syntax_error(tag_token, "#bounds_check and #no_bounds_check cannot be applied together")
		}
	case StateFlag_no_bounds_check:
		if s.state_flags & StateFlag_bounds_check != 0 {
			syntax_error(tag_token, "#bounds_check and #no_bounds_check cannot be applied together")
		}
	case StateFlag_type_assert:
		if s.state_flags & StateFlag_no_type_assert != 0 {
			syntax_error(tag_token, "#type_assert and #no_type_assert cannot be applied together")
		}
	case StateFlag_no_type_assert:
		if s.state_flags & StateFlag_type_assert != 0 {
			syntax_error(tag_token, "#type_assert and #no_type_assert cannot be applied together")
		}
	}

	// Validate that the directive is applied to an allowed statement type
	switch state_flag {
	case StateFlag_bounds_check,
	     StateFlag_no_bounds_check,
	     StateFlag_type_assert,
	     StateFlag_no_type_assert:
		switch s.kind {
		case .BlockStmt, .IfStmt, .WhenStmt, .ForStmt, .RangeStmt,
		     .UnrollRangeStmt, .SwitchStmt, .TypeSwitchStmt,
		     .ReturnStmt, .DeferStmt, .AssignStmt:
			// Allowed
		case .ValueDecl:
			if !s.ValueDecl.is_mutable {
				syntax_error(tag_token, "#%s may only be applied to a variable declaration, and not a constant value declaration", name)
			}
		case:
			syntax_error(tag_token, "#%s may only be applied to the following statements: '{}', 'if', 'when', 'for', 'switch', 'return', 'defer', assignment, variable declaration", name)
		}
	}
	return s
}

// =============================================================================
// parse_union_variant_list — parse union variant types inside { ... }
// C++ lines 2395-2409
// =============================================================================

parse_union_variant_list :: proc(f: ^AstFile) -> [dynamic]^Ast {
	allocator := ast_allocator(f)
	variants := make([dynamic]^Ast, allocator)
	for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
		parse_enforce_tabs(f)
		type := parse_type(f)
		if type.kind != .BadExpr {
			append(&variants, type)
		}
		if !allow_field_separator(f) {
			break
		}
	}
	return variants
}

// =============================================================================
// parser_check_polymorphic_record_parameters — validate $ vs bare name consistency
// C++ lines 2410-2446
// =============================================================================

Prefix_State :: enum int {
	Unknown = 0,
	Dollar  = 1,
	Bare    = 2,
}

parser_check_polymorphic_record_parameters :: proc(f: ^AstFile, polymorphic_params: ^Ast) {
	if polymorphic_params == nil {
		return
	}
	if polymorphic_params.kind != .FieldList {
		return
	}
	prefix: Prefix_State = .Unknown
	for field in polymorphic_params.FieldList.list {
		if field == nil || field.kind != .Field {
			continue
		}
		for name in field.Field.names {
			if name == nil {
				continue
			}
			error := false
			if name.kind == .Ident {
				switch prefix {
				case .Unknown: prefix = .Bare
				case .Dollar:  error = true
				case .Bare:
					// OK
				}
			} else if name.kind == .PolyType {
				switch prefix {
				case .Unknown: prefix = .Dollar
				case .Dollar:
					// OK
				case .Bare:    error = true
				}
			}
			if error {
				syntax_error(name, "Mixture of polymorphic $ names and normal identifiers are not allowed within record parameters")
			}
		}
	}
}

// =============================================================================
// parse_operand — THE BIG ONE: parse primary operand expressions and types
// C++ lines 2447-3129 (~680 lines)
// =============================================================================

parse_operand :: proc(f: ^AstFile, lhs: bool) -> ^Ast {
	switch f.curr_token.kind {
	// ----------------------------------------------------------------
	// Simple identifiers and literals
	// ----------------------------------------------------------------
	case .Ident:
		return parse_ident(f)

	case .Uninit:
		return ast_uninit(f, expect_token(f, .Uninit))

	case .Context:
		return ast_implicit(f, expect_token(f, .Context))

	case .Integer, .Float, .Imag, .Rune:
		return ast_basic_lit(f, advance_token(f))

	case .String:
		return ast_basic_lit(f, advance_token(f))

	case .OpenBrace:
		if !lhs {
			return parse_literal_value(f, nil)
		}
		// Fall through to nil return on LHS

	// ----------------------------------------------------------------
	// Parenthesized expression
	// ----------------------------------------------------------------
	case .OpenParen: {
		open := expect_token(f, .OpenParen)
		if f.prev_token.kind == .CloseParen {
			close := expect_token(f, .CloseParen)
			syntax_error(open, "Invalid parentheses expression with no inside expression")
			return ast_bad_expr(f, open, close)
		}
		prev_expr_level := f.expr_level
		prev_allow_newline := f.allow_newline
		if f.expr_level < 0 {
			f.allow_newline = false
		}
		f.expr_level = max(f.expr_level, 0) + 1
		operand := parse_expr(f, false)
		f.allow_newline = prev_allow_newline
		f.expr_level = prev_expr_level
		close := expect_token(f, .CloseParen)
		return ast_paren_expr(f, operand, open, close)
	}

	// ----------------------------------------------------------------
	// distinct type
	// ----------------------------------------------------------------
	case .Distinct: {
		token := expect_token(f, .Distinct)
		type := parse_type(f)
		return ast_distinct_type(f, token, type)
	}

	// ----------------------------------------------------------------
	// Hash directives: #type, #simd, #soa, #partial, #sparse, etc.
	// ----------------------------------------------------------------
	case .Hash: {
		token := expect_token(f, .Hash)
		name := expect_token(f, .Ident)

		// #type -- helper type
		if name.string == "type" {
			return ast_helper_type(f, token, parse_type(f))
		}

		// #simd -- SIMD vector type
		if name.string == "simd" {
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			type := unparen_expr(original_type)
			switch type.kind {
			case .ArrayType:
				type.ArrayType.tag = tag
			case:
				syntax_error(type, "Expected a fixed array type after #%s, got %s", name.string, ast_strings[type.kind])
			}
			return original_type
		}

		// #soa -- structure of arrays
		if name.string == "soa" {
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			type := unparen_expr(original_type)
			switch type.kind {
			case .ArrayType:
				type.ArrayType.tag = tag
			case .DynamicArrayType:
				type.DynamicArrayType.tag = tag
			case .PointerType:
				type.PointerType.tag = tag
			case .FixedCapacityDynamicArrayType:
				type.FixedCapacityDynamicArrayType.tag = tag
			case:
				syntax_error(type, "Expected an array or pointer type after #%s, got %s", name.string, ast_strings[type.kind])
			}
			return original_type
		}

		// #row_major / #column_major -- matrix layout
		if name.string == "row_major" || name.string == "column_major" {
			original_type := parse_type(f)
			type := unparen_expr(original_type)
			switch type.kind {
			case .MatrixType:
				type.MatrixType.is_row_major = (name.string == "row_major")
			case:
				syntax_error(type, "Expected a matrix type after #%s, got %s", name.string, ast_strings[type.kind])
			}
			return original_type
		}

		// #partial -- partial compound literal
		if name.string == "partial" {
			tag := ast_basic_directive(f, token, name)
			original_expr := parse_expr(f, lhs)
			expr := unparen_expr(original_expr)
			if expr == nil {
				syntax_error(name, "Expected a compound literal after #%s", name.string)
				return ast_bad_expr(f, token, name)
			}
			switch expr.kind {
			case .CompoundLit:
				expr.CompoundLit.tag = tag
			case:
				syntax_error(expr, "Expected a compound literal after #%s, got %s", name.string, ast_strings[expr.kind])
			}
			return original_expr
		}

		// #sparse -- sparse array type
		if name.string == "sparse" {
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			type := unparen_expr(original_type)
			switch type.kind {
			case .ArrayType:
				type.ArrayType.tag = tag
			case:
				syntax_error(type, "Expected an enumerated array type after #%s, got %s", name.string, ast_strings[type.kind])
			}
			return original_type
		}

		// #bounds_check -- per-statement bounds check
		if name.string == "bounds_check" {
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlag_bounds_check)
		}

		// #no_bounds_check
		if name.string == "no_bounds_check" {
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlag_no_bounds_check)
		}

		// #type_assert
		if name.string == "type_assert" {
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlag_type_assert)
		}

		// #no_type_assert
		if name.string == "no_type_assert" {
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlag_no_type_assert)
		}

		// #relative -- deprecated, replaced by core:relative
		if name.string == "relative" {
			tag := ast_basic_directive(f, token, name)
			if f.curr_token.kind != .OpenParen {
				syntax_error(tag, "expected #relative(<integer type>) <type>")
			} else {
				tag = parse_call_expr(f, tag)
			}
			type := parse_type(f)
			syntax_error(tag, "#relative types have now been removed in favour of \"core:relative\"")
			return ast_relative_type(f, tag, type)
		}

		// #force_inline / #force_no_inline / #must_tail
		if name.string == "force_inline" || name.string == "force_no_inline" || name.string == "must_tail" {
			return parse_inlining_or_tailing_operand(f, name)
		}

		// Unknown #directive → basic directive node
		return ast_basic_directive(f, token, name)
	}

	// ----------------------------------------------------------------
	// proc — procedure type, literal, or group
	// ----------------------------------------------------------------
	case .Proc: {
		token := expect_token(f, .Proc)

		// Procedure group: proc { ... }
		if f.curr_token.kind == .OpenBrace {
			open := expect_token(f, .OpenBrace)
			allocator := ast_allocator(f)
			args := make([dynamic]^Ast, allocator)
			for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
				elem := parse_expr(f, false)
				append(&args, elem)
				if !allow_field_separator(f) {
					break
				}
			}
			close := expect_token(f, .CloseBrace)
			if len(args) == 0 {
				syntax_error(token, "Expected at least 1 argument in a procedure group")
			}
			return ast_proc_group(f, token, open, close, args)
		}

		// Procedure type, literal, or inline body
		type := parse_proc_type(f, token)
		where_token: Token
		where_clauses: [dynamic]^Ast
		tags: u64 = 0

		skip_possible_newline_for_literal(f)

		if f.curr_token.kind == .Where {
			where_token = expect_token(f, .Where)
			prev_level := f.expr_level
			f.expr_level = -1
			where_clauses = parse_rhs_expr_list(f)
			f.expr_level = prev_level
		}

		parse_proc_tags(f, &tags)

		if tags & ProcTag_require_results != 0 {
			syntax_error(f.curr_token, "#require_results has now been replaced as an attribute @(require_results) on the declaration")
			tags &= ~ProcTag_require_results
		}

		assert(type.kind == .ProcType)
		type.ProcType.tags = tags

		// If we're in a type context (expr_level < 0) and allow_type is set,
		// return as a procedure type (not a literal)
		if f.allow_type && f.expr_level < 0 {
			if tags != 0 {
				syntax_error(token, "A procedure type cannot have suffix tags")
			}
			if where_token.kind != .Invalid {
				syntax_error(where_token, "'where' clauses are not allowed on procedure types")
			}
			return type
		}

		skip_possible_newline_for_literal(f, where_token.kind == .Where)

		if allow_token(f, .Uninit) {
			// --- (no body) procedure literal
			if where_token.kind != .Invalid {
				syntax_error(where_token, "'where' clauses are not allowed on procedure literals without a defined body (replaced with ---)")
			}
			return ast_proc_lit(f, type, nil, tags, where_token, where_clauses)
		} else if f.curr_token.kind == .OpenBrace {
			// Procedure literal with body: proc { ... }
			curr_proc := f.curr_proc
			body: ^Ast = nil
			f.curr_proc = type
			body = parse_body(f)
			f.curr_proc = curr_proc
			// Propagate tag states to the body
			if tags & ProcTag_no_bounds_check != 0 {
				body.state_flags |= StateFlag_no_bounds_check
			}
			if tags & ProcTag_bounds_check != 0 {
				body.state_flags |= StateFlag_bounds_check
			}
			if tags & ProcTag_no_type_assert != 0 {
				body.state_flags |= StateFlag_no_type_assert
			}
			if tags & ProcTag_type_assert != 0 {
				body.state_flags |= StateFlag_type_assert
			}
			return ast_proc_lit(f, type, body, tags, where_token, where_clauses)
		} else if allow_token(f, .Do) {
			// Deprecated: proc do stmt
			curr_proc := f.curr_proc
			body: ^Ast = nil
			f.curr_proc = type
			body = convert_stmt_to_body(f, parse_stmt(f))
			f.curr_proc = curr_proc
			syntax_error(body, "'do' for procedure bodies is not allowed, prefer {}")
			return ast_proc_lit(f, type, body, tags, where_token, where_clauses)
		}

		// No body — just a procedure type
		if tags != 0 {
			syntax_error(token, "A procedure type cannot have suffix tags")
		}
		if where_token.kind != .Invalid {
			syntax_error(where_token, "'where' clauses are not allowed on procedure types")
		}
		return type
	}

	// ----------------------------------------------------------------
	// $ — polymorphic type parameter
	// ----------------------------------------------------------------
	case .Dollar: {
		token := expect_token(f, .Dollar)
		type := parse_ident(f)
		if is_blank_ident(type) {
			syntax_error(type, "Invalid polymorphic type definition with a blank identifier")
		}
		specialization: ^Ast = nil
		if allow_token(f, .Quo) {
			specialization = parse_type(f)
		}
		return ast_poly_type(f, token, type, specialization)
	}

	// ----------------------------------------------------------------
	// typeid — type identifier
	// ----------------------------------------------------------------
	case .Typeid: {
		token := expect_token(f, .Typeid)
		return ast_typeid_type(f, token, nil)
	}

	// ----------------------------------------------------------------
	// ^ — pointer type
	// ----------------------------------------------------------------
	case .Pointer: {
		token := expect_token(f, .Pointer)
		elem := parse_type(f)
		return ast_pointer_type(f, token, elem)
	}

	// ----------------------------------------------------------------
	// * — dereference (unary expression, LHS context)
	// ----------------------------------------------------------------
	case .Mul:
		return parse_unary_expr(f, true)

	// ----------------------------------------------------------------
	// [ — array, dynamic array, multi-pointer, or fixed-capacity dynamic array
	// ----------------------------------------------------------------
	case .OpenBracket: {
		token := expect_token(f, .OpenBracket)
		count_expr: ^Ast = nil

		if f.curr_token.kind == .Pointer {
			// [^]T — multi-pointer
			expect_token(f, .Pointer)
			expect_token(f, .CloseBracket)
			return ast_multi_pointer_type(f, token, parse_type(f))
		} else if f.curr_token.kind == .Question {
			// [?]T — array with inferred count
			count_expr = ast_unary_expr(f, expect_token(f, .Question), nil)
		} else if allow_token(f, .Dynamic) {
			// [dynamic]T or [dynamic; cap]T
			capacity: ^Ast = nil
			if f.curr_token.kind == .Semicolon && f.curr_token.string == ";" {
				expect_token(f, .Semicolon)
				capacity = parse_expr(f, false)
			} else if allow_token(f, .Comma) || allow_token(f, .Semicolon) {
				p := token_to_string(f.prev_token)
				syntax_error(token_end_of_line(f, f.prev_token), "Expected a semicolon, got a %s", p)
				capacity = parse_expr(f, false)
			}
			expect_token(f, .CloseBracket)
			elem := parse_type(f)
			if capacity == nil {
				return ast_dynamic_array_type(f, token, elem)
			} else {
				return ast_fixed_capacity_dynamic_array_type(f, token, capacity, elem)
			}
		} else if f.curr_token.kind != .CloseBracket {
			// [N]T or [ ]T — fixed array
			f.expr_level += 1
			count_expr = parse_expr(f, false)
			f.expr_level -= 1
		}

		expect_token(f, .CloseBracket)
		return ast_array_type(f, token, count_expr, parse_type(f))
	}

	// ----------------------------------------------------------------
	// map[K]V — map type
	// ----------------------------------------------------------------
	case .Map: {
		token := expect_token(f, .Map)
		open  := expect_token_after(f, .OpenBracket, "map")
		key   := parse_expr(f, true)
		close := expect_token(f, .CloseBracket)
		value := parse_type(f)
		_ = open
		_ = close
		return ast_map_type(f, token, key, value)
	}

	// ----------------------------------------------------------------
	// matrix[R, C]T — matrix type
	// ----------------------------------------------------------------
	case .Matrix: {
		token := expect_token(f, .Matrix)
		open  := expect_token_after(f, .OpenBracket, "matrix")
		row_count := parse_expr(f, true)
		expect_token(f, .Comma)
		column_count := parse_expr(f, true)
		close := expect_token(f, .CloseBracket)
		type  := parse_type(f)
		_ = open
		_ = close
		return ast_matrix_type(f, token, row_count, column_count, type)
	}

	// ----------------------------------------------------------------
	// bit_field — bit field type
	// ----------------------------------------------------------------
	case .Bit_Field: {
		token := expect_token(f, .Bit_Field)
		prev_level := f.expr_level
		f.expr_level = -1
		backing_type := parse_type_or_ident(f)
		if backing_type == nil {
			bt := advance_token(f)
			syntax_error(bt, "Expected a backing type for a 'bit_field'")
			backing_type = ast_bad_expr(f, bt, f.curr_token)
		}
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, .OpenBrace, "bit_field")
		allocator := ast_allocator(f)
		fields := make([dynamic]^Ast, allocator)
		for f.curr_token.kind != .CloseBrace && f.curr_token.kind != .EOF {
			docs: ^CommentGroup = nil
			comment: ^CommentGroup = nil
			name := parse_ident(f)
			err_once := false
			for allow_token(f, .Comma) {
				dummy_name := parse_ident(f)
				if !err_once {
					syntax_error(dummy_name, "'bit_field' fields do not support multiple names per field")
					err_once = true
				}
			}
			expect_token(f, .Colon)
			type := parse_type(f)
			expect_token(f, .Or)
			bit_size := parse_expr(f, true)
			tag: Token
			if f.curr_token.kind == .String {
				tag = expect_token(f, .String)
			}
			bf_field := ast_bit_field_field(f, name, type, bit_size, tag, docs, comment)
			append(&fields, bf_field)
			if !allow_field_separator(f) {
				break
			}
		}
		close := expect_closing_brace_of_field_list(f)
		f.expr_level = prev_level
		return ast_bit_field_type(f, token, backing_type, open, fields, close)
	}

	// ----------------------------------------------------------------
	// struct — struct type definition
	// ----------------------------------------------------------------
	case .Struct: {
		token := expect_token(f, .Struct)
		polymorphic_params: ^Ast = nil
		is_packed := false
		is_all_or_none := false
		is_raw_union := false
		is_simple := false
		align: ^Ast = nil
		min_field_align: ^Ast = nil
		max_field_align: ^Ast = nil

		// Parse polymorphic parameters: struct($T: typeid) { ... }
		if allow_token(f, .OpenParen) {
			param_count: int = 0
			polymorphic_params = parse_field_list(f, &param_count, 0, .CloseParen, true, true)
			if param_count == 0 {
				syntax_error(polymorphic_params, "Expected at least 1 polymorphic parameter")
				polymorphic_params = nil
			}
			expect_token_after(f, .CloseParen, "parameter list")
			check_polymorphic_params_for_type(f, polymorphic_params, token)
		}

		// Parse struct tags: #packed, #align, #raw_union, etc.
		prev_level := f.expr_level
		f.expr_level = -1
		for allow_token(f, .Hash) {
			tag := expect_token_after(f, .Ident, "#")
			switch tag.string {
			case "packed":
				if is_packed {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				is_packed = true
			case "all_or_none":
				if is_all_or_none {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				is_all_or_none = true
			case "align":
				if align != nil {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				align = parse_expr(f, true)
				if align != nil && align.kind != .ParenExpr {
					// begin_error_block()
					s := expr_to_string(align)
					syntax_warning(tag, "#align requires parentheses around the expression")
					error_line("\tSuggestion: #align(%s)", s)
					// end_error_block()
				}
			case "field_align":
				if min_field_align != nil {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				syntax_warning(tag, "#field_align has been deprecated in favour of #min_field_align")
				min_field_align = parse_expr(f, true)
				if min_field_align != nil && min_field_align.kind != .ParenExpr {
					// begin_error_block()
					s := expr_to_string(min_field_align)
					syntax_warning(tag, "#field_align requires parentheses around the expression")
					error_line("\tSuggestion: #min_field_align(%s)", s)
					// end_error_block()
				}
			case "min_field_align":
				if min_field_align != nil {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				min_field_align = parse_expr(f, true)
				if min_field_align != nil && min_field_align.kind != .ParenExpr {
					// begin_error_block()
					s := expr_to_string(min_field_align)
					syntax_warning(tag, "#min_field_align requires parentheses around the expression")
					error_line("\tSuggestion: #min_field_align(%s)", s)
					// end_error_block()
				}
			case "max_field_align":
				if max_field_align != nil {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				max_field_align = parse_expr(f, true)
				if max_field_align != nil && max_field_align.kind != .ParenExpr {
					// begin_error_block()
					s := expr_to_string(max_field_align)
					syntax_warning(tag, "#max_field_align requires parentheses around the expression")
					error_line("\tSuggestion: #max_field_align(%s)", s)
					// end_error_block()
				}
			case "raw_union":
				if is_raw_union {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				is_raw_union = true
			case "simple":
				if is_simple {
					syntax_error(tag, "Duplicate struct tag '#%s'", tag.string)
				}
				is_simple = true
			case:
				syntax_error(tag, "Invalid struct tag '#%s'", tag.string)
			}
		}
		f.expr_level = prev_level

		// Conflict checks
		if is_raw_union && is_packed {
			is_packed = false
			syntax_error(token, "'#raw_union' cannot also be '#packed'")
		}
		if is_raw_union && is_all_or_none {
			is_all_or_none = false
			syntax_error(token, "'#raw_union' cannot also be '#all_or_none'")
		}

		// where clauses
		where_token: Token
		where_clauses: [dynamic]^Ast
		skip_possible_newline_for_literal(f)
		if f.curr_token.kind == .Where {
			where_token = expect_token(f, .Where)
			prev_level = f.expr_level
			f.expr_level = -1
			where_clauses = parse_rhs_expr_list(f)
			f.expr_level = prev_level
		}

		// Parse struct body { ... }
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, .OpenBrace, "struct")
		name_count: int = 0
		fields := parse_struct_field_list(f, &name_count)
		close := expect_closing_brace_of_field_list(f)

		decls: []^Ast
		if fields != nil {
			assert(fields.kind == .FieldList)
			decls = fields.FieldList.list
		}

		parser_check_polymorphic_record_parameters(f, polymorphic_params)

		return ast_struct_type(f, token, decls, name_count,
			polymorphic_params, is_packed, is_raw_union, is_all_or_none, is_simple,
			align, min_field_align, max_field_align,
			where_token, where_clauses)
	}

	// ----------------------------------------------------------------
	// union — union type definition
	// ----------------------------------------------------------------
	case .Union: {
		token := expect_token(f, .Union)
		polymorphic_params: ^Ast = nil
		align: ^Ast = nil
		no_nil := false
		maybe := false
		shared_nil := false
		union_kind: UnionTypeKind = .Normal

		// Parse polymorphic parameters: union($T: typeid) { ... }
		if allow_token(f, .OpenParen) {
			param_count: int = 0
			polymorphic_params = parse_field_list(f, &param_count, 0, .CloseParen, true, true)
			if param_count == 0 {
				syntax_error(polymorphic_params, "Expected at least 1 polymorphic parameter")
				polymorphic_params = nil
			}
			expect_token_after(f, .CloseParen, "parameter list")
			check_polymorphic_params_for_type(f, polymorphic_params, token)
		}

		// Parse union tags: #align, #no_nil, #shared_nil, #maybe
		for allow_token(f, .Hash) {
			tag := expect_token_after(f, .Ident, "#")
			switch tag.string {
			case "align":
				if align != nil {
					syntax_error(tag, "Duplicate union tag '#%s'", tag.string)
				}
				align = parse_expr(f, true)
				if align != nil && align.kind != .ParenExpr {
					// begin_error_block()
					s := expr_to_string(align)
					syntax_warning(tag, "#align requires parentheses around the expression")
					error_line("\tSuggestion: #align(%s)", s)
					// end_error_block()
				}
			case "no_nil":
				if no_nil {
					syntax_error(tag, "Duplicate union tag '#%s'", tag.string)
				}
				no_nil = true
			case "shared_nil":
				if shared_nil {
					syntax_error(tag, "Duplicate union tag '#%s'", tag.string)
				}
				shared_nil = true
			case "maybe":
				if maybe {
					syntax_error(tag, "Duplicate union tag '#%s'", tag.string)
				}
				maybe = true
			case:
				syntax_error(tag, "Invalid union tag '#%s'", tag.string)
			}
		}

		if no_nil && shared_nil {
			syntax_error(f.curr_token, "#shared_nil and #no_nil cannot be applied together")
		}
		if maybe {
			syntax_error(f.curr_token, "#maybe functionality has now been merged with standard 'union' functionality")
		}

		if no_nil {
			union_kind = .No_Nil
		} else if shared_nil {
			union_kind = .Shared_Nil
		}

		// where clauses
		where_token: Token
		where_clauses: [dynamic]^Ast
		skip_possible_newline_for_literal(f)
		if f.curr_token.kind == .Where {
			where_token = expect_token(f, .Where)
			prev_level := f.expr_level
			f.expr_level = -1
			where_clauses = parse_rhs_expr_list(f)
			f.expr_level = prev_level
		}

		// Parse union body { ... }
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, .OpenBrace, "union")
		variants := parse_union_variant_list(f)
		close := expect_closing_brace_of_field_list(f)

		parser_check_polymorphic_record_parameters(f, polymorphic_params)

		return ast_union_type(f, token, variants, polymorphic_params, align, union_kind, where_token, where_clauses)
	}

	// ----------------------------------------------------------------
	// enum — enum type definition
	// ----------------------------------------------------------------
	case .Enum: {
		token := expect_token(f, .Enum)
		base_type: ^Ast = nil
		if f.curr_token.kind != .OpenBrace {
			base_type = parse_type(f)
		}
		skip_possible_newline_for_literal(f)
		open := expect_token(f, .OpenBrace)
		values := parse_enum_field_list(f)
		close := expect_closing_brace_of_field_list(f)
		_ = open
		_ = close
		return ast_enum_type(f, token, base_type, values)
	}

	// ----------------------------------------------------------------
	// bit_set — bit set type definition
	// ----------------------------------------------------------------
	case .Bit_Set: {
		token := expect_token(f, .Bit_Set)
		expect_token(f, .OpenBracket)
		elem: ^Ast = nil
		underlying: ^Ast = nil
		prev_allow_range := f.allow_range
		f.allow_range = true
		elem = parse_expr(f, true)
		f.allow_range = prev_allow_range
		if elem == nil {
			syntax_error(token, "Expected a type or range, got nothing")
		}
		if f.curr_token.kind == .Semicolon && f.curr_token.string == ";" {
			expect_token(f, .Semicolon)
			underlying = parse_type(f)
		} else if allow_token(f, .Comma) || allow_token(f, .Semicolon) {
			p := token_to_string(f.prev_token)
			syntax_error(token_end_of_line(f, f.prev_token), "Expected a semicolon, got a %s", p)
			underlying = parse_type(f)
		}
		expect_token(f, .CloseBracket)
		return ast_bit_set_type(f, token, elem, underlying)
	}

	// ----------------------------------------------------------------
	// asm — inline assembly expression
	// ----------------------------------------------------------------
	case .Asm: {
		token := expect_token(f, .Asm)
		param_types: [dynamic]^Ast
		return_type: ^Ast = nil

		if allow_token(f, .OpenParen) {
			allocator := ast_allocator(f)
			param_types = make([dynamic]^Ast, allocator)
			for f.curr_token.kind != .CloseParen && f.curr_token.kind != .EOF {
				t := parse_type(f)
				append(&param_types, t)
				if f.curr_token.kind != .Comma || f.curr_token.kind == .EOF {
					break
				}
				advance_token(f)
			}
			expect_token(f, .CloseParen)
			if allow_token(f, .ArrowRight) {
				return_type = parse_type(f)
			}
		}

		has_side_effects := false
		is_align_stack := false
		dialect: InlineAsmDialectKind = .Default

		for f.curr_token.kind == .Hash {
			advance_token(f)
			if f.curr_token.kind == .Ident {
				directive := advance_token(f)
				dname := directive.string
				switch dname {
				case "side_effects":
					if has_side_effects {
						syntax_error(directive, "Duplicate directive on inline asm expression: '#side_effects'")
					}
					has_side_effects = true
				case "align_stack":
					if is_align_stack {
						syntax_error(directive, "Duplicate directive on inline asm expression: '#align_stack'")
					}
					is_align_stack = true
				case "att":
					if dialect == .ATT {
						syntax_error(directive, "Duplicate directive on inline asm expression: '#att'")
					} else if dialect != .Default {
						syntax_error(directive, "Conflicting asm dialects")
					} else {
						dialect = .ATT
					}
				case "intel":
					if dialect == .Intel {
						syntax_error(directive, "Duplicate directive on inline asm expression: '#intel'")
					} else if dialect != .Default {
						syntax_error(directive, "Conflicting asm dialects")
					} else {
						dialect = .Intel
					}
				case:
					syntax_error(directive, "Invalid directive on inline asm expression: '#%s'", directive.string)
				}
			} else {
				syntax_error(f.curr_token, "Expected an identifier after hash")
}