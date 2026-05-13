package odingo

import "core:strings"

// =============================================================================
// Forward-declared type stubs
// =============================================================================

TokenKind :: enum int {
	Invalid,
	Rune,
	String,
	Integer,
	Float,
}

Token :: struct {
	kind:   TokenKind,
	string: string,
}

CommentGroup :: struct {}

ExactValue_Kind :: enum int { Invalid, String }

ExactValue :: struct {
	kind:         ExactValue_Kind,
	value_string: string,
}

AddressingMode :: enum int { Constant }

ProcCallingConvention :: enum int { Odin }

UnionTypeKind :: enum int { Normal }

InlineAsmDialectKind :: enum int { Att, Intel }

Allocator :: struct {
	procedure: rawptr,
	data:      rawptr,
}

AstKind :: enum int {
	BadExpr,
	TagExpr,
	UnaryExpr,
	BinaryExpr,
	ParenExpr,
	CallExpr,
	SelectorExpr,
	ImplicitSelectorExpr,
	SelectorCallExpr,
	IndexExpr,
	SliceExpr,
	DerefExpr,
	MatrixIndexExpr,
	Ident,
	Implicit,
	Uninit,
	BasicLit,
	BasicDirective,
	Ellipsis,
	ProcGroup,
	ProcLit,
	FieldValue,
	EnumFieldValue,
	CompoundLit,
	TernaryIfExpr,
	TernaryWhenExpr,
	OrElseExpr,
	OrReturnExpr,
	OrBranchExpr,
	TypeAssertion,
	TypeCast,
	AutoCast,
	InlineAsmExpr,
	BadStmt,
	EmptyStmt,
	ExprStmt,
	AssignStmt,
	BlockStmt,
	IfStmt,
	WhenStmt,
	ReturnStmt,
	ForStmt,
	RangeStmt,
	UnrollRangeStmt,
	SwitchStmt,
	TypeSwitchStmt,
	CaseClause,
	DeferStmt,
	BranchStmt,
	UsingStmt,
	BadDecl,
	Field,
	BitFieldField,
	FieldList,
	TypeidType,
	HelperType,
	DistinctType,
	PolyType,
	ProcType,
	RelativeType,
	PointerType,
	MultiPointerType,
	ArrayType,
	DynamicArrayType,
	FixedCapacityDynamicArrayType,
	StructType,
	UnionType,
	EnumType,
	BitSetType,
	BitFieldType,
	MapType,
	MatrixType,
	ForeignBlockDecl,
	Label,
	ValueDecl,
	PackageDecl,
	ImportDecl,
	ForeignImportDecl,
	Attribute,
}

Ast :: struct {
	kind: AstKind,
	tav: struct {
		mode:  AddressingMode,
		value: ExactValue,
	},
	BadExpr:              struct { begin, end: Token },
	TagExpr:              struct { token, name: Token, expr: ^Ast },
	UnaryExpr:            struct { op: Token, expr: ^Ast },
	BinaryExpr:           struct { op: Token, left, right: ^Ast },
	ParenExpr:            struct { open, close: Token, expr: ^Ast },
	CallExpr:             struct { open, close, ellipsis: Token, proc: ^Ast, args: []^Ast },
	SelectorExpr:         struct { token: Token, expr, selector: ^Ast },
	ImplicitSelectorExpr: struct { token: Token, selector: ^Ast },
	SelectorCallExpr:     struct { token: Token, expr, call: ^Ast },
	IndexExpr:            struct { open, close: Token, expr, index: ^Ast },
	SliceExpr:            struct { open, close, interval: Token, expr, low, high: ^Ast },
	DerefExpr:            struct { op: Token, expr: ^Ast },
	MatrixIndexExpr: struct {
		expr, row_index, column_index: ^Ast,
		open, close:                  Token,
	},
	Ident:     struct { token: Token },
	Implicit:  Token,
	Uninit:    Token,
	BasicLit:  struct { token: Token },
	BasicDirective: struct { token, name: Token },
	Ellipsis:       struct { token: Token, expr: ^Ast },
	ProcGroup:      struct { token, open, close: Token, args: []^Ast },
	ProcLit:        struct { type, body: ^Ast, tags: u64, where_token: Token, where_clauses: []^Ast },
	FieldValue:     struct { field, value: ^Ast, eq: Token },
	EnumFieldValue: struct { name, value: ^Ast, docs, comment: ^CommentGroup },
	CompoundLit:    struct { type: ^Ast, elems: []^Ast, open, close: Token },
	TernaryIfExpr:   struct { x, cond, y: ^Ast },
	TernaryWhenExpr: struct { x, cond, y: ^Ast },
	OrElseExpr:      struct { x, y: ^Ast, token: Token },
	OrReturnExpr:    struct { expr: ^Ast, token: Token },
	OrBranchExpr:    struct { expr, label: ^Ast, token: Token },
	TypeAssertion:   struct { expr, type: ^Ast, dot: Token },
	TypeCast:        struct { token: Token, type, expr: ^Ast },
	AutoCast:        struct { token: Token, expr: ^Ast },
	InlineAsmExpr: struct {
		token, open, close:     Token,
		param_types:            []^Ast,
		return_type:            ^Ast,
		asm_string:             ^Ast,
		constraints_string:     ^Ast,
		has_side_effects:       bool,
		is_align_stack:         bool,
		dialect:                InlineAsmDialectKind,
	},
	BadStmt:      struct { begin, end: Token },
	EmptyStmt:    struct { token: Token },
	ExprStmt:     struct { expr: ^Ast },
	AssignStmt:   struct { op: Token, lhs, rhs: []^Ast },
	BlockStmt:    struct { stmts: []^Ast, open, close: Token },
	IfStmt:       struct { token: Token, init, cond, body, else_stmt: ^Ast },
	WhenStmt:     struct { token: Token, cond, body, else_stmt: ^Ast },
	ReturnStmt:   struct { token: Token, results: []^Ast },
	ForStmt:      struct { token: Token, init, cond, post, body: ^Ast },
	RangeStmt: struct {
		token, in_token: Token,
		init:             ^Ast,
		vals:             []^Ast,
		expr, body:       ^Ast,
	},
	UnrollRangeStmt: struct {
		unroll_token, for_token, in_token: Token,
		init:                              ^Ast,
		args:                              []^Ast,
		val0, val1, expr, body:           ^Ast,
	},
	SwitchStmt:     struct { token: Token, init, tag, body: ^Ast, partial: bool },
	TypeSwitchStmt: struct { token: Token, tag, body: ^Ast, partial: bool },
	CaseClause:     struct { token: Token, list, stmts: []^Ast },
	DeferStmt:      struct { token: Token, stmt: ^Ast },
	BranchStmt:     struct { token: Token, label: ^Ast },
	UsingStmt:      struct { token: Token, list: []^Ast },
	BadDecl:        struct { begin, end: Token },
	Field: struct {
		names:                    []^Ast,
		type, default_value:      ^Ast,
		flags:                    u32,
		tag:                      Token,
		docs, comment:            ^CommentGroup,
	},
	BitFieldField: struct {
		name, type, bit_size: ^Ast,
		tag:                   Token,
		docs, comment:         ^CommentGroup,
	},
	FieldList:       struct { token: Token, list: []^Ast },
	TypeidType:      struct { token: Token, specialization: ^Ast },
	HelperType:      struct { token: Token, type: ^Ast },
	DistinctType:    struct { token: Token, type: ^Ast },
	PolyType:        struct { token: Token, type, specialization: ^Ast },
	ProcType: struct {
		token:                       Token,
		params, results:             ^Ast,
		tags:                        u64,
		calling_convention:          ProcCallingConvention,
		generic, diverging:          bool,
	},
	RelativeType:       struct { tag, type: ^Ast },
	PointerType:        struct { token: Token, type: ^Ast },
	MultiPointerType:   struct { token: Token, type: ^Ast },
	ArrayType:          struct { token: Token, count, elem: ^Ast },
	DynamicArrayType:   struct { token: Token, elem: ^Ast },
	FixedCapacityDynamicArrayType: struct { token: Token, capacity, elem: ^Ast },
	StructType: struct {
		token:                           Token,
		fields:                          []^Ast,
		field_count:                     int,
		polymorphic_params:              ^Ast,
		is_packed:                       bool,
		is_raw_union:                    bool,
		is_all_or_none:                  bool,
		is_simple:                       bool,
		align, min_field_align:          ^Ast,
		max_field_align:                 ^Ast,
		where_token:                     Token,
		where_clauses:                   []^Ast,
	},
	UnionType: struct {
		token:               Token,
		variants:            []^Ast,
		polymorphic_params:  ^Ast,
		align:               ^Ast,
		kind:                UnionTypeKind,
		where_token:         Token,
		where_clauses:       []^Ast,
	},
	EnumType:       struct { token: Token, base_type: ^Ast, fields: []^Ast },
	BitSetType:     struct { token: Token, elem, underlying: ^Ast },
	BitFieldType: struct {
		token, open, close: Token,
		backing_type:        ^Ast,
		fields:              []^Ast,
	},
	MapType:    struct { token: Token, key, value: ^Ast },
	MatrixType: struct { token: Token, row_count, column_count, elem: ^Ast },
	ForeignBlockDecl: struct {
		token:                   Token,
		foreign_library, body:   ^Ast,
		docs:                    ^CommentGroup,
		attributes: struct { allocator: Allocator },
	},
	Label: struct { token: Token, name: ^Ast },
	ValueDecl: struct {
		names, values: []^Ast,
		type:                  ^Ast,
		is_mutable:            bool,
		docs, comment:         ^CommentGroup,
		attributes: struct { allocator: Allocator },
	},
	PackageDecl: struct {
		token, name:        Token,
		docs, comment:      ^CommentGroup,
	},
	ImportDecl: struct {
		token, relpath, import_name: Token,
		docs, comment:               ^CommentGroup,
		attributes: struct { allocator: Allocator },
	},
	ForeignImportDecl: struct {
		token, library_name: Token,
		filepaths:                   []^Ast,
		docs, comment:               ^CommentGroup,
		multiple_filepaths:          bool,
		attributes: struct { allocator: Allocator },
	},
	Attribute: struct { token, open, close: Token, elems: []^Ast },
}

AstFile :: struct {
	seen_load_directive_count: int,
}

// =============================================================================
// External function declarations (defined elsewhere in odingo package)
// =============================================================================

alloc_ast_node              :: proc(f: ^AstFile, kind: AstKind) -> ^Ast
ast_allocator               :: proc(f: ^AstFile) -> Allocator
exact_value_from_basic_literal :: proc(kind: TokenKind, s: string) -> ExactValue
unquote_string              :: proc(allocator: Allocator, s: ^string, flags: int, raw_string: bool = false) -> bool
syntax_error                :: proc(tok: Token, format: string, args: ..any)
syntax_error_with_verbose   :: proc(node: ^Ast, format: string, args: ..any)

// =============================================================================
// Expression constructors
// =============================================================================

ast_bad_expr :: proc(f: ^AstFile, begin, end: Token) -> ^Ast {
	result := alloc_ast_node(f, .BadExpr)
	result.BadExpr.begin = begin
	result.BadExpr.end = end
	return result
}

ast_tag_expr :: proc(f: ^AstFile, token, name: Token, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TagExpr)
	result.TagExpr.token = token
	result.TagExpr.name = name
	result.TagExpr.expr = expr
	return result
}

ast_unary_expr :: proc(f: ^AstFile, op: Token, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .UnaryExpr)
	if expr != nil {
		#partial switch expr.kind {
		case .OrReturnExpr:
			syntax_error_with_verbose(expr, "'or_return' within an unary expression not wrapped in parentheses (...)")
		case .OrBranchExpr:
			syntax_error_with_verbose(expr, "'%s' within an unary expression not wrapped in parentheses (...)", expr.OrBranchExpr.token.string)
		}
	}
	result.UnaryExpr.op = op
	result.UnaryExpr.expr = expr
	return result
}

ast_binary_expr :: proc(f: ^AstFile, op: Token, left, right: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .BinaryExpr)
	if left == nil {
		syntax_error(op, "No lhs expression for binary expression '%s'", op.string)
		left = ast_bad_expr(f, op, op)
	}
	if right == nil {
		syntax_error(op, "No rhs expression for binary expression '%s'", op.string)
		right = ast_bad_expr(f, op, op)
	}
	if left != nil {
		#partial switch left.kind {
		case .OrReturnExpr:
			syntax_error_with_verbose(left, "'or_return' within a binary expression not wrapped in parentheses (...)")
		case .OrBranchExpr:
			syntax_error_with_verbose(left, "'%s' within a binary expression not wrapped in parentheses (...)", left.OrBranchExpr.token.string)
		}
	}
	if right != nil {
		#partial switch right.kind {
		case .OrReturnExpr:
			syntax_error_with_verbose(right, "'or_return' within a binary expression not wrapped in parentheses (...)")
		case .OrBranchExpr:
			syntax_error_with_verbose(right, "'%s' within a binary expression not wrapped in parentheses (...)", right.OrBranchExpr.token.string)
		}
	}
	result.BinaryExpr.op = op
	result.BinaryExpr.left = left
	result.BinaryExpr.right = right
	return result
}

ast_paren_expr :: proc(f: ^AstFile, expr: ^Ast, open, close: Token) -> ^Ast {
	result := alloc_ast_node(f, .ParenExpr)
	result.ParenExpr.expr = expr
	result.ParenExpr.open = open
	result.ParenExpr.close = close
	return result
}

ast_call_expr :: proc(f: ^AstFile, proc_expr: ^Ast, args: []^Ast, open, close, ellipsis: Token) -> ^Ast {
	result := alloc_ast_node(f, .CallExpr)
	result.CallExpr.proc = proc_expr
	result.CallExpr.args = args
	result.CallExpr.open = open
	result.CallExpr.close = close
	result.CallExpr.ellipsis = ellipsis
	return result
}

ast_selector_expr :: proc(f: ^AstFile, token: Token, expr, selector: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .SelectorExpr)
	result.SelectorExpr.token = token
	result.SelectorExpr.expr = expr
	result.SelectorExpr.selector = selector
	return result
}

ast_implicit_selector_expr :: proc(f: ^AstFile, token: Token, selector: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ImplicitSelectorExpr)
	result.ImplicitSelectorExpr.token = token
	result.ImplicitSelectorExpr.selector = selector
	return result
}

ast_selector_call_expr :: proc(f: ^AstFile, token: Token, expr, call: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .SelectorCallExpr)
	result.SelectorCallExpr.token = token
	result.SelectorCallExpr.expr = expr
	result.SelectorCallExpr.call = call
	return result
}

ast_index_expr :: proc(f: ^AstFile, expr, index: ^Ast, open, close: Token) -> ^Ast {
	result := alloc_ast_node(f, .IndexExpr)
	result.IndexExpr.expr = expr
	result.IndexExpr.index = index
	result.IndexExpr.open = open
	result.IndexExpr.close = close
	return result
}

ast_slice_expr :: proc(f: ^AstFile, expr: ^Ast, open, close, interval: Token, low, high: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .SliceExpr)
	result.SliceExpr.expr = expr
	result.SliceExpr.open = open
	result.SliceExpr.close = close
	result.SliceExpr.interval = interval
	result.SliceExpr.low = low
	result.SliceExpr.high = high
	return result
}

ast_deref_expr :: proc(f: ^AstFile, expr: ^Ast, op: Token) -> ^Ast {
	result := alloc_ast_node(f, .DerefExpr)
	result.DerefExpr.expr = expr
	result.DerefExpr.op = op
	return result
}

ast_matrix_index_expr :: proc(f: ^AstFile, expr: ^Ast, open, close: Token, row, column: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .MatrixIndexExpr)
	result.MatrixIndexExpr.expr = expr
	result.MatrixIndexExpr.row_index = row
	result.MatrixIndexExpr.column_index = column
	result.MatrixIndexExpr.open = open
	result.MatrixIndexExpr.close = close
	return result
}

ast_ident :: proc(f: ^AstFile, token: Token) -> ^Ast {
	result := alloc_ast_node(f, .Ident)
	result.Ident.token = token
	// result.Ident.hash = string_hash(token.string)   -- ported elsewhere
	// result.Ident.interned = string_interner_insert(token.string) -- skipped (string interner not yet ported)
	return result
}

ast_implicit :: proc(f: ^AstFile, token: Token) -> ^Ast {
	result := alloc_ast_node(f, .Implicit)
	result.Implicit = token
	return result
}

ast_uninit :: proc(f: ^AstFile, token: Token) -> ^Ast {
	result := alloc_ast_node(f, .Uninit)
	result.Uninit = token
	return result
}

exact_value_from_token :: proc(f: ^AstFile, token: Token) -> ExactValue {
	s := token.string
	// string_interner_insert(s)  -- skipped (string interner not yet ported)
	#partial switch token.kind {
	case .Rune:
		if !unquote_string(ast_allocator(f), &s, 0) {
			syntax_error(token, "Invalid rune literal")
		}
	case .String:
		raw := len(s) > 0 && s[0] == '`'
		if !unquote_string(ast_allocator(f), &s, 0, raw) {
			syntax_error(token, "Invalid string literal")
		}
	}
	value := exact_value_from_basic_literal(token.kind, s)
	if value.kind == .Invalid {
		#partial switch token.kind {
		case .Integer:
			syntax_error(token, "Invalid integer literal")
		case .Float:
			if !strings.contains_rune(s, '.') && !strings.contains_rune(s, '-') {
				syntax_error(token, "Invalid integer literal")
			} else {
				syntax_error(token, "Invalid float literal")
			}
		case:
			syntax_error(token, "Invalid token literal")
		}
	}
	return value
}

string_value_from_token :: proc(f: ^AstFile, token: Token) -> string {
	value := exact_value_from_token(f, token)
	if value.kind == .String {
		return value.value_string
	}
	return ""
}

ast_basic_lit :: proc(f: ^AstFile, basic_lit: Token) -> ^Ast {
	result := alloc_ast_node(f, .BasicLit)
	result.BasicLit.token = basic_lit
	result.tav.mode = .Constant
	result.tav.value = exact_value_from_token(f, basic_lit)
	return result
}

ast_basic_directive :: proc(f: ^AstFile, token, name: Token) -> ^Ast {
	result := alloc_ast_node(f, .BasicDirective)
	result.BasicDirective.token = token
	result.BasicDirective.name = name
	// string_interner_insert(name.string)  -- skipped (string interner not yet ported)
	if strings.starts_with(name.string, "load") {
		f.seen_load_directive_count += 1
	}
	return result
}

ast_ellipsis :: proc(f: ^AstFile, token: Token, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .Ellipsis)
	result.Ellipsis.token = token
	result.Ellipsis.expr = expr
	return result
}

ast_proc_group :: proc(f: ^AstFile, token, open, close: Token, args: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ProcGroup)
	result.ProcGroup.token = token
	result.ProcGroup.open = open
	result.ProcGroup.close = close
	result.ProcGroup.args = args
	return result
}

ast_proc_lit :: proc(f: ^AstFile, type, body: ^Ast, tags: u64, where_token: Token, where_clauses: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ProcLit)
	result.ProcLit.type = type
	result.ProcLit.body = body
	result.ProcLit.tags = tags
	result.ProcLit.where_token = where_token
	result.ProcLit.where_clauses = where_clauses
	return result
}

ast_field_value :: proc(f: ^AstFile, field, value: ^Ast, eq: Token) -> ^Ast {
	result := alloc_ast_node(f, .FieldValue)
	result.FieldValue.field = field
	result.FieldValue.value = value
	result.FieldValue.eq = eq
	return result
}

ast_enum_field_value :: proc(f: ^AstFile, name, value: ^Ast, docs, comment: ^CommentGroup) -> ^Ast {
	result := alloc_ast_node(f, .EnumFieldValue)
	result.EnumFieldValue.name = name
	result.EnumFieldValue.value = value
	result.EnumFieldValue.docs = docs
	result.EnumFieldValue.comment = comment
	return result
}

ast_compound_lit :: proc(f: ^AstFile, type: ^Ast, elems: []^Ast, open, close: Token) -> ^Ast {
	result := alloc_ast_node(f, .CompoundLit)
	result.CompoundLit.type = type
	result.CompoundLit.elems = elems
	result.CompoundLit.open = open
	result.CompoundLit.close = close
	return result
}

ast_ternary_if_expr :: proc(f: ^AstFile, x, cond, y: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TernaryIfExpr)
	result.TernaryIfExpr.x = x
	result.TernaryIfExpr.cond = cond
	result.TernaryIfExpr.y = y
	return result
}

ast_ternary_when_expr :: proc(f: ^AstFile, x, cond, y: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TernaryWhenExpr)
	result.TernaryWhenExpr.x = x
	result.TernaryWhenExpr.cond = cond
	result.TernaryWhenExpr.y = y
	return result
}

ast_or_else_expr :: proc(f: ^AstFile, x: ^Ast, token: Token, y: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .OrElseExpr)
	result.OrElseExpr.x = x
	result.OrElseExpr.token = token
	result.OrElseExpr.y = y
	return result
}

ast_or_return_expr :: proc(f: ^AstFile, expr: ^Ast, token: Token) -> ^Ast {
	result := alloc_ast_node(f, .OrReturnExpr)
	result.OrReturnExpr.expr = expr
	result.OrReturnExpr.token = token
	return result
}

ast_or_branch_expr :: proc(f: ^AstFile, expr: ^Ast, token: Token, label: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .OrBranchExpr)
	result.OrBranchExpr.expr = expr
	result.OrBranchExpr.token = token
	result.OrBranchExpr.label = label
	return result
}

ast_type_assertion :: proc(f: ^AstFile, expr: ^Ast, dot: Token, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TypeAssertion)
	result.TypeAssertion.expr = expr
	result.TypeAssertion.dot = dot
	result.TypeAssertion.type = type
	return result
}

ast_type_cast :: proc(f: ^AstFile, token: Token, type, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TypeCast)
	result.TypeCast.token = token
	result.TypeCast.type = type
	result.TypeCast.expr = expr
	return result
}

ast_auto_cast :: proc(f: ^AstFile, token: Token, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .AutoCast)
	result.AutoCast.token = token
	result.AutoCast.expr = expr
	return result
}

ast_inline_asm_expr :: proc(
	f: ^AstFile,
	token, open, close: Token,
	param_types: []^Ast,
	return_type: ^Ast,
	asm_string: ^Ast,
	constraints_string: ^Ast,
	has_side_effects: bool,
	is_align_stack: bool,
	dialect: InlineAsmDialectKind,
) -> ^Ast {
	result := alloc_ast_node(f, .InlineAsmExpr)
	result.InlineAsmExpr.token = token
	result.InlineAsmExpr.open = open
	result.InlineAsmExpr.close = close
	result.InlineAsmExpr.param_types = param_types
	result.InlineAsmExpr.return_type = return_type
	result.InlineAsmExpr.asm_string = asm_string
	result.InlineAsmExpr.constraints_string = constraints_string
	result.InlineAsmExpr.has_side_effects = has_side_effects
	result.InlineAsmExpr.is_align_stack = is_align_stack
	result.InlineAsmExpr.dialect = dialect
	return result
}

// =============================================================================
// Statement constructors
// =============================================================================

ast_bad_stmt :: proc(f: ^AstFile, begin, end: Token) -> ^Ast {
	result := alloc_ast_node(f, .BadStmt)
	result.BadStmt.begin = begin
	result.BadStmt.end = end
	return result
}

ast_empty_stmt :: proc(f: ^AstFile, token: Token) -> ^Ast {
	result := alloc_ast_node(f, .EmptyStmt)
	result.EmptyStmt.token = token
	return result
}

ast_expr_stmt :: proc(f: ^AstFile, expr: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ExprStmt)
	result.ExprStmt.expr = expr
	return result
}

ast_assign_stmt :: proc(f: ^AstFile, op: Token, lhs, rhs: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .AssignStmt)
	result.AssignStmt.op = op
	result.AssignStmt.lhs = lhs
	result.AssignStmt.rhs = rhs
	return result
}

ast_block_stmt :: proc(f: ^AstFile, stmts: []^Ast, open, close: Token) -> ^Ast {
	result := alloc_ast_node(f, .BlockStmt)
	result.BlockStmt.stmts = stmts
	result.BlockStmt.open = open
	result.BlockStmt.close = close
	return result
}

ast_if_stmt :: proc(f: ^AstFile, token: Token, init, cond, body, else_stmt: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .IfStmt)
	result.IfStmt.token = token
	result.IfStmt.init = init
	result.IfStmt.cond = cond
	result.IfStmt.body = body
	result.IfStmt.else_stmt = else_stmt
	return result
}

ast_when_stmt :: proc(f: ^AstFile, token: Token, cond, body, else_stmt: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .WhenStmt)
	result.WhenStmt.token = token
	result.WhenStmt.cond = cond
	result.WhenStmt.body = body
	result.WhenStmt.else_stmt = else_stmt
	return result
}

ast_return_stmt :: proc(f: ^AstFile, token: Token, results: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ReturnStmt)
	result.ReturnStmt.token = token
	result.ReturnStmt.results = results
	return result
}

ast_for_stmt :: proc(f: ^AstFile, token: Token, init, cond, post, body: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ForStmt)
	result.ForStmt.token = token
	result.ForStmt.init = init
	result.ForStmt.cond = cond
	result.ForStmt.post = post
	result.ForStmt.body = body
	return result
}

ast_range_stmt :: proc(f: ^AstFile, token: Token, init: ^Ast, vals: []^Ast, in_token: Token, expr, body: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .RangeStmt)
	result.RangeStmt.token = token
	result.RangeStmt.init = init
	result.RangeStmt.vals = vals
	result.RangeStmt.in_token = in_token
	result.RangeStmt.expr = expr
	result.RangeStmt.body = body
	return result
}

ast_unroll_range_stmt :: proc(
	f: ^AstFile,
	unroll_token: Token,
	init: ^Ast,
	args: []^Ast,
	for_token: Token,
	val0, val1: ^Ast,
	in_token: Token,
	expr, body: ^Ast,
) -> ^Ast {
	result := alloc_ast_node(f, .UnrollRangeStmt)
	result.UnrollRangeStmt.unroll_token = unroll_token
	result.UnrollRangeStmt.init = init
	result.UnrollRangeStmt.args = args
	result.UnrollRangeStmt.for_token = for_token
	result.UnrollRangeStmt.val0 = val0
	result.UnrollRangeStmt.val1 = val1
	result.UnrollRangeStmt.in_token = in_token
	result.UnrollRangeStmt.expr = expr
	result.UnrollRangeStmt.body = body
	return result
}

ast_switch_stmt :: proc(f: ^AstFile, token: Token, init, tag, body: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .SwitchStmt)
	result.SwitchStmt.token = token
	result.SwitchStmt.init = init
	result.SwitchStmt.tag = tag
	result.SwitchStmt.body = body
	result.SwitchStmt.partial = false
	return result
}

ast_type_switch_stmt :: proc(f: ^AstFile, token: Token, tag, body: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TypeSwitchStmt)
	result.TypeSwitchStmt.token = token
	result.TypeSwitchStmt.tag = tag
	result.TypeSwitchStmt.body = body
	result.TypeSwitchStmt.partial = false
	return result
}

ast_case_clause :: proc(f: ^AstFile, token: Token, list, stmts: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .CaseClause)
	result.CaseClause.token = token
	result.CaseClause.list = list
	result.CaseClause.stmts = stmts
	return result
}

ast_defer_stmt :: proc(f: ^AstFile, token: Token, stmt: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .DeferStmt)
	result.DeferStmt.token = token
	result.DeferStmt.stmt = stmt
	return result
}

ast_branch_stmt :: proc(f: ^AstFile, token: Token, label: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .BranchStmt)
	result.BranchStmt.token = token
	result.BranchStmt.label = label
	return result
}

ast_using_stmt :: proc(f: ^AstFile, token: Token, list: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .UsingStmt)
	result.UsingStmt.token = token
	result.UsingStmt.list = list
	return result
}

// =============================================================================
// Declaration and type constructors
// =============================================================================

ast_bad_decl :: proc(f: ^AstFile, begin, end: Token) -> ^Ast {
	result := alloc_ast_node(f, .BadDecl)
	result.BadDecl.begin = begin
	result.BadDecl.end = end
	return result
}

ast_field :: proc(
	f: ^AstFile,
	names: []^Ast,
	type, default_value: ^Ast,
	flags: u32,
	tag: Token,
	docs, comment: ^CommentGroup,
) -> ^Ast {
	result := alloc_ast_node(f, .Field)
	result.Field.names = names
	result.Field.type = type
	result.Field.default_value = default_value
	result.Field.flags = flags
	result.Field.tag = tag
	result.Field.docs = docs
	result.Field.comment = comment
	return result
}

ast_bit_field_field :: proc(
	f: ^AstFile,
	name, type, bit_size: ^Ast,
	tag: Token,
	docs, comment: ^CommentGroup,
) -> ^Ast {
	result := alloc_ast_node(f, .BitFieldField)
	result.BitFieldField.name = name
	result.BitFieldField.type = type
	result.BitFieldField.bit_size = bit_size
	result.BitFieldField.tag = tag
	result.BitFieldField.docs = docs
	result.BitFieldField.comment = comment
	return result
}

ast_field_list :: proc(f: ^AstFile, token: Token, list: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .FieldList)
	result.FieldList.token = token
	result.FieldList.list = list
	return result
}

ast_typeid_type :: proc(f: ^AstFile, token: Token, specialization: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .TypeidType)
	result.TypeidType.token = token
	result.TypeidType.specialization = specialization
	return result
}

ast_helper_type :: proc(f: ^AstFile, token: Token, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .HelperType)
	result.HelperType.token = token
	result.HelperType.type = type
	return result
}

ast_distinct_type :: proc(f: ^AstFile, token: Token, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .DistinctType)
	result.DistinctType.token = token
	result.DistinctType.type = type
	return result
}

ast_poly_type :: proc(f: ^AstFile, token: Token, type, specialization: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .PolyType)
	result.PolyType.token = token
	result.PolyType.type = type
	result.PolyType.specialization = specialization
	return result
}

ast_proc_type :: proc(
	f: ^AstFile,
	token: Token,
	params, results: ^Ast,
	tags: u64,
	calling_convention: ProcCallingConvention,
	generic, diverging: bool,
) -> ^Ast {
	result := alloc_ast_node(f, .ProcType)
	result.ProcType.token = token
	result.ProcType.params = params
	result.ProcType.results = results
	result.ProcType.tags = tags
	result.ProcType.calling_convention = calling_convention
	result.ProcType.generic = generic
	result.ProcType.diverging = diverging
	return result
}

ast_relative_type :: proc(f: ^AstFile, tag, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .RelativeType)
	result.RelativeType.tag = tag
	result.RelativeType.type = type
	return result
}

ast_pointer_type :: proc(f: ^AstFile, token: Token, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .PointerType)
	result.PointerType.token = token
	result.PointerType.type = type
	return result
}

ast_multi_pointer_type :: proc(f: ^AstFile, token: Token, type: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .MultiPointerType)
	result.MultiPointerType.token = token
	result.MultiPointerType.type = type
	return result
}

ast_array_type :: proc(f: ^AstFile, token: Token, count, elem: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .ArrayType)
	result.ArrayType.token = token
	result.ArrayType.count = count
	result.ArrayType.elem = elem
	return result
}

ast_dynamic_array_type :: proc(f: ^AstFile, token: Token, elem: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .DynamicArrayType)
	result.DynamicArrayType.token = token
	result.DynamicArrayType.elem = elem
	return result
}

ast_fixed_capacity_dynamic_array_type :: proc(f: ^AstFile, token: Token, capacity, elem: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .FixedCapacityDynamicArrayType)
	result.FixedCapacityDynamicArrayType.token = token
	result.FixedCapacityDynamicArrayType.capacity = capacity
	result.FixedCapacityDynamicArrayType.elem = elem
	return result
}

ast_struct_type :: proc(
	f: ^AstFile,
	token: Token,
	fields: []^Ast,
	field_count: int,
	polymorphic_params: ^Ast,
	is_packed, is_raw_union, is_all_or_none, is_simple: bool,
	align, min_field_align, max_field_align: ^Ast,
	where_token: Token,
	where_clauses: []^Ast,
) -> ^Ast {
	result := alloc_ast_node(f, .StructType)
	result.StructType.token = token
	result.StructType.fields = fields
	result.StructType.field_count = field_count
	result.StructType.polymorphic_params = polymorphic_params
	result.StructType.is_packed = is_packed
	result.StructType.is_raw_union = is_raw_union
	result.StructType.is_all_or_none = is_all_or_none
	result.StructType.is_simple = is_simple
	result.StructType.align = align
	result.StructType.min_field_align = min_field_align
	result.StructType.max_field_align = max_field_align
	result.StructType.where_token = where_token
	result.StructType.where_clauses = where_clauses
	return result
}

ast_union_type :: proc(
	f: ^AstFile,
	token: Token,
	variants: []^Ast,
	polymorphic_params, align: ^Ast,
	kind: UnionTypeKind,
	where_token: Token,
	where_clauses: []^Ast,
) -> ^Ast {
	result := alloc_ast_node(f, .UnionType)
	result.UnionType.token = token
	result.UnionType.variants = variants
	result.UnionType.polymorphic_params = polymorphic_params
	result.UnionType.align = align
	result.UnionType.kind = kind
	result.UnionType.where_token = where_token
	result.UnionType.where_clauses = where_clauses
	return result
}

ast_enum_type :: proc(f: ^AstFile, token: Token, base_type: ^Ast, fields: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .EnumType)
	result.EnumType.token = token
	result.EnumType.base_type = base_type
	result.EnumType.fields = fields
	return result
}

ast_bit_set_type :: proc(f: ^AstFile, token: Token, elem, underlying: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .BitSetType)
	result.BitSetType.token = token
	result.BitSetType.elem = elem
	result.BitSetType.underlying = underlying
	return result
}

ast_bit_field_type :: proc(f: ^AstFile, token: Token, backing_type: ^Ast, open: Token, fields: []^Ast, close: Token) -> ^Ast {
	result := alloc_ast_node(f, .BitFieldType)
	result.BitFieldType.token = token
	result.BitFieldType.backing_type = backing_type
	result.BitFieldType.open = open
	result.BitFieldType.fields = fields
	result.BitFieldType.close = close
	return result
}

ast_map_type :: proc(f: ^AstFile, token: Token, key, value: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .MapType)
	result.MapType.token = token
	result.MapType.key = key
	result.MapType.value = value
	return result
}

ast_matrix_type :: proc(f: ^AstFile, token: Token, row_count, column_count, elem: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .MatrixType)
	result.MatrixType.token = token
	result.MatrixType.row_count = row_count
	result.MatrixType.column_count = column_count
	result.MatrixType.elem = elem
	return result
}

ast_foreign_block_decl :: proc(f: ^AstFile, token: Token, foreign_library, body: ^Ast, docs: ^CommentGroup) -> ^Ast {
	result := alloc_ast_node(f, .ForeignBlockDecl)
	result.ForeignBlockDecl.token = token
	result.ForeignBlockDecl.foreign_library = foreign_library
	result.ForeignBlockDecl.body = body
	result.ForeignBlockDecl.docs = docs
	result.ForeignBlockDecl.attributes.allocator = ast_allocator(f)
	return result
}

ast_label_decl :: proc(f: ^AstFile, token: Token, name: ^Ast) -> ^Ast {
	result := alloc_ast_node(f, .Label)
	result.Label.token = token
	result.Label.name = name
	return result
}

ast_value_decl :: proc(
	f: ^AstFile,
	names: []^Ast,
	type: ^Ast,
	values: []^Ast,
	is_mutable: bool,
	docs, comment: ^CommentGroup,
) -> ^Ast {
	result := alloc_ast_node(f, .ValueDecl)
	result.ValueDecl.names = names
	result.ValueDecl.type = type
	result.ValueDecl.values = values
	result.ValueDecl.is_mutable = is_mutable
	result.ValueDecl.docs = docs
	result.ValueDecl.comment = comment
	result.ValueDecl.attributes.allocator = ast_allocator(f)
	return result
}

ast_package_decl :: proc(f: ^AstFile, token, name: Token, docs, comment: ^CommentGroup) -> ^Ast {
	result := alloc_ast_node(f, .PackageDecl)
	result.PackageDecl.token = token
	result.PackageDecl.name = name
	result.PackageDecl.docs = docs
	result.PackageDecl.comment = comment
	return result
}

ast_import_decl :: proc(
	f: ^AstFile,
	token, relpath, import_name: Token,
	docs, comment: ^CommentGroup,
) -> ^Ast {
	result := alloc_ast_node(f, .ImportDecl)
	result.ImportDecl.token = token
	result.ImportDecl.relpath = relpath
	result.ImportDecl.import_name = import_name
	result.ImportDecl.docs = docs
	result.ImportDecl.comment = comment
	result.ImportDecl.attributes.allocator = ast_allocator(f)
	return result
}

ast_foreign_import_decl :: proc(
	f: ^AstFile,
	token: Token,
	filepaths: []^Ast,
	library_name: Token,
	multiple_filepaths: bool,
	docs, comment: ^CommentGroup,
) -> ^Ast {
	result := alloc_ast_node(f, .ForeignImportDecl)
	result.ForeignImportDecl.token = token
	result.ForeignImportDecl.filepaths = filepaths
	result.ForeignImportDecl.library_name = library_name
	result.ForeignImportDecl.docs = docs
	result.ForeignImportDecl.comment = comment
	result.ForeignImportDecl.multiple_filepaths = multiple_filepaths
	result.ForeignImportDecl.attributes.allocator = ast_allocator(f)
	return result
}

ast_attribute :: proc(f: ^AstFile, token, open, close: Token, elems: []^Ast) -> ^Ast {
	result := alloc_ast_node(f, .Attribute)
	result.Attribute.token = token
	result.Attribute.open = open
	result.Attribute.elems = elems
	result.Attribute.close = close
	return result
}
