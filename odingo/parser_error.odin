package odingo

import "core:fmt"

// =============================================================================
// External function declarations (defined elsewhere in odingo package)
// =============================================================================

error_va                     :: proc(pos, end: TokenPos, msg: string)
syntax_error_with_verbose_va :: proc(pos, end: TokenPos, msg: string)
error_no_newline_va          :: proc(pos: TokenPos, msg: string)
warning_va                   :: proc(pos, end: TokenPos, msg: string)
syntax_error_va              :: proc(pos, end: TokenPos, msg: string)

// Forward-declared (defined in other files within this package):
//   Ast, Token, TokenPos, AstKind, AstFile
//   ast_token      :: proc(node: ^Ast) -> Token
//   ast_end_pos    :: proc(node: ^Ast) -> TokenPos
//   thread_safe_get_ast_file_from_id :: proc(id: u64) -> ^AstFile
//   ast_strings: []string — maps AstKind enum values to their name strings

// =============================================================================
// Error / Warning reporting
// =============================================================================

error :: proc(node: ^Ast, fmt_str: string, args: ..any) {
	token: Token = {}
	end_pos: TokenPos = {}
	if node != nil {
		token = ast_token(node)
		end_pos = ast_end_pos(node)
	}
	formatted := fmt.aprintf(fmt_str, ..args)
	error_va(token.pos, end_pos, formatted)
	if node != nil && node.file_id != 0 {
		f := thread_safe_get_ast_file_from_id(node.file_id)
		if f != nil {
			f.error_count += 1
		}
	}
}

error_range :: proc(start, end: TokenPos, fmt_str: string, args: ..any) {
	if start.file_id != end.file_id { panic("start.file_id == end.file_id") }
	if start.line != end.line       { panic("start.line == end.line") }
	if start.column > end.column    { panic("start.column <= end.column") }
	if start.offset > end.offset    { panic("start.offset <= end.offset") }
	formatted := fmt.aprintf(fmt_str, ..args)
	error_va(start, end, formatted)
	if start.file_id != 0 {
		f := thread_safe_get_ast_file_from_id(start.file_id)
		if f != nil {
			f.error_count += 1
		}
	}
}

syntax_error_with_verbose :: proc(node: ^Ast, fmt_str: string, args: ..any) {
	token: Token = {}
	end_pos: TokenPos = {}
	if node != nil {
		token = ast_token(node)
		end_pos = ast_end_pos(node)
	}
	formatted := fmt.aprintf(fmt_str, ..args)
	syntax_error_with_verbose_va(token.pos, end_pos, formatted)
	if node != nil && node.file_id != 0 {
		f := thread_safe_get_ast_file_from_id(node.file_id)
		if f != nil {
			f.error_count += 1
		}
	}
}

error_no_newline :: proc(node: ^Ast, fmt_str: string, args: ..any) {
	token: Token = {}
	if node != nil {
		token = ast_token(node)
	}
	formatted := fmt.aprintf(fmt_str, ..args)
	error_no_newline_va(token.pos, formatted)
	if node != nil && node.file_id != 0 {
		f := thread_safe_get_ast_file_from_id(node.file_id)
		if f != nil {
			f.error_count += 1
		}
	}
}

warning :: proc(node: ^Ast, fmt_str: string, args: ..any) {
	token: Token = {}
	end_pos: TokenPos = {}
	if node != nil {
		token = ast_token(node)
		end_pos = ast_end_pos(node)
	}
	formatted := fmt.aprintf(fmt_str, ..args)
	warning_va(token.pos, end_pos, formatted)
}

syntax_error :: proc(node: ^Ast, fmt_str: string, args: ..any) {
	token: Token = {}
	end_pos: TokenPos = {}
	if node != nil {
		token = ast_token(node)
		end_pos = ast_end_pos(node)
	}
	formatted := fmt.aprintf(fmt_str, ..args)
	syntax_error_va(token.pos, end_pos, formatted)
	if node != nil && node.file_id != 0 {
		f := thread_safe_get_ast_file_from_id(node.file_id)
		if f != nil {
			f.error_count += 1
		}
	}
}

// =============================================================================
// AST node kind checking
// =============================================================================

ast_node_expect :: proc(node: ^Ast, kind: AstKind) -> bool {
	if node.kind != kind {
		syntax_error(node, "Expected %s, got %s", ast_strings[kind], ast_strings[node.kind])
		return false
	}
	return true
}

ast_node_expect2 :: proc(node: ^Ast, kind0, kind1: AstKind) -> bool {
	if node.kind != kind0 && node.kind != kind1 {
		syntax_error(node, "Expected %s or %s, got %s", ast_strings[kind0], ast_strings[kind1], ast_strings[node.kind])
		return false
	}
	return true
}
