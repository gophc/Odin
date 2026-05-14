package cmd

import (
	"sync/atomic"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Forward declarations of parser functions defined in other files
// ---------------------------------------------------------------------------

func parse_expr(f *AstFile, lhs bool) *Ast
func parse_type(f *AstFile) *Ast
func parse_unary_expr(f *AstFile, lhs bool) *Ast
func parse_simple_stmt(f *AstFile, flags uint32) *Ast
func parse_lhs_expr_list(f *AstFile) []*Ast
func parse_rhs_expr_list(f *AstFile) []*Ast
func parse_call_expr(f *AstFile, operand *Ast) *Ast
func parse_body(f *AstFile) *Ast
func parse_do_body(f *AstFile, token Token, msg string) *Ast
func parse_block_stmt(f *AstFile, is_when bool) *Ast
func parse_stmt(f *AstFile) *Ast
func parse_stmt_list(f *AstFile) []*Ast
func parse_struct_field_list(f *AstFile, nameCount *isize) *Ast
func parse_field_list(f *AstFile, nameCount *isize, allowedFlags FieldFlag, follow TokenKind, allowDefaultParams bool, allowTypeidToken bool) *Ast
func parse_proc_type(f *AstFile, procToken Token) *Ast
func parse_value_decl(f *AstFile, names []*Ast, docs *CommentGroup) *Ast
func parse_var_type(f *AstFile, allowEllipsis bool, allowTypeidToken bool) *Ast
func parse_operand(f *AstFile, lhs bool) *Ast
func parse_type_or_ident(f *AstFile) *Ast
func parse_ident_list(f *AstFile, allowPolyNames bool) []*Ast
func parse_foreign_block(f *AstFile, token Token) *Ast
func parse_attribute(f *AstFile, token Token, openKind TokenKind, closeKind TokenKind, docs *CommentGroup) *Ast
func parse_unrolled_for_loop(f *AstFile, unrollToken Token) *Ast
func parse_check_directive_for_statement(s *Ast, name Token, stateFlag StateFlag) *Ast
func parse_inlining_or_tailing_operand(f *AstFile, name Token) *Ast

func ast_file_vet_flags(f *AstFile) uint64
func ast_file_vet_style(f *AstFile) bool
func ast_end_token(node *Ast) Token
func ast_end_pos(node *Ast) TokenPos
func ast_token(node *Ast) Token
func ast_allocator(f *AstFile) gbAllocator
func string_value_from_token(f *AstFile, token Token) String
func token_to_string(tok Token) String
func token_end_of_line(f *AstFile, token Token) TokenPos
func token_pos_end(tok Token) TokenPos

func is_blank_ident(tok Token) bool
func file_allow_newline(f *AstFile) bool
func consume_comment_groups(f *AstFile, prev Token)

func string_value_from_token_t(tok Token) String
func token_is_keyword(kind TokenKind) bool

func alloc_ast_node(f *AstFile, kind AstKind) *Ast

func goStr(s String) string

func syntax_error_pos(pos TokenPos, format string, args ...any)
func syntax_error(node *Ast, format string, args ...any)
func error_line(format string, args ...any)
func error(node *Ast, format string, args ...any)
func begin_error_block()
func end_error_block()
func warning(tok Token, format string, args ...any)

func thread_safe_set_ast_file_from_id(id int32, f *AstFile)
func thread_safe_get_ast_file_from_id(id int32) *AstFile

func prev_pow2(x i64) i64
func time_stamp_time_now() u64
func time_stamp__freq() u64

func find_library_collection_path(name String, baseDir *String) bool
func is_excluded_target_filename(name String) bool
func is_arch_wasm() bool

func set_file_path_string(id int32, path String)
func permanent_alloc_item[T any]() *T

func thread_pool_add_task(proc WorkerTaskProc, data unsafe.Pointer) bool
func thread_pool_wait()

type WorkerTaskProc func(unsafe.Pointer) isize

func copy_string(a gbAllocator, s String) String
func string_trim_whitespace(s String) String
func string_starts_with(s, prefix String) bool
func string_ends_with(s, suffix String) bool
func string_contains_char(s String, c byte) bool
func string_eq_ignore_case(a, b String) bool
func substring(s String, lo, hi isize) String
func remove_directory_from_path(s String) String
func remove_extension_from_path(s String) String
func string_is_valid_identifier(s String) bool
func path_to_full_path(a gbAllocator, path String) String
func path_is_directory(path String) bool
func path_extension(path String) String
func get_fullpath_base_collection(a gbAllocator, name String, ok *bool) String
func get_fullpath_core_collection(a gbAllocator, name String, ok *bool) String
func get_fullpath_relative(a gbAllocator, base, path String, ok *bool) String
func odin_root_dir() String
func directory_from_path(path String) String
func filename_from_path(path String) String
func get_file_size(path String) i64
func read_directory(path String, list *[]FileInfo) ReadDirectoryError
func make_string_c(s string) String

func heap_allocator() gbAllocator
func permanent_allocator() gbAllocator
func temporary_allocator() gbAllocator

func mutex_lock(m *BlockingMutex)
func mutex_unlock(m *BlockingMutex)

type gbFileContents struct {
	Data unsafe.Pointer
	Size isize
}

func gb_file_read_contents(a gbAllocator, text bool, path string) gbFileContents

type StringSet struct{}

func string_set_init(s *StringSet)
func string_set_destroy(s *StringSet)
func string_set_update(s *StringSet, str String) bool

type ReadDirectoryError int

const (
	ReadDirectory_OK          ReadDirectoryError = 0
	ReadDirectory_InvalidPath ReadDirectoryError = 1
	ReadDirectory_NotExists   ReadDirectoryError = 2
	ReadDirectory_Permission  ReadDirectoryError = 3
	ReadDirectory_NotDir      ReadDirectoryError = 4
	ReadDirectory_Empty       ReadDirectoryError = 5
	ReadDirectory_Unknown     ReadDirectoryError = 6
)

// ---------------------------------------------------------------------------
// Helper predicates
// ---------------------------------------------------------------------------

func is_ast_expr(node *Ast) bool {
	if node == nil {
		return false
	}
	return node.Kind > AstExprBegin && node.Kind < AstExprEnd
}

func is_ast_stmt(node *Ast) bool {
	if node == nil {
		return false
	}
	return node.Kind > AstStmtBegin && node.Kind < AstStmtEnd
}

func is_ast_complex_stmt(node *Ast) bool {
	if node == nil {
		return false
	}
	return node.Kind > AstComplexStmtBegin && node.Kind < AstComplexStmtEnd
}

func is_ast_decl(node *Ast) bool {
	if node == nil {
		return false
	}
	return node.Kind > AstDeclBegin && node.Kind < AstDeclEnd
}

func is_ast_type(node *Ast) bool {
	if node == nil {
		return false
	}
	return node.Kind > AstTypeBegin && node.Kind < AstTypeEnd
}

func is_when_stmt(node *Ast) bool {
	return node != nil && node.Kind == AstWhenStmt
}

// ---------------------------------------------------------------------------
// parse_enforce_tabs
// ---------------------------------------------------------------------------

func parse_enforce_tabs(f *AstFile) {
	if (ast_file_vet_flags(f) & uint64(VetFlagTabs)) == 0 {
		return
	}
	prev := f.PrevToken
	curr := f.CurrToken
	if prev.Pos.Line < curr.Pos.Line {
		start := f.Tokenizer.Start + uintptr(prev.Pos.Offset)
		end := f.Tokenizer.Start + uintptr(curr.Pos.Offset)
		it := end
		for it > start {
			if *(*byte)(unsafe.Pointer(it)) == '\n' {
				it++
				break
			}
			it--
		}
		length := end - it
		for i := uintptr(0); i < length; i++ {
			if *(*byte)(unsafe.Pointer(it + i)) == '/' {
				break
			}
			if *(*byte)(unsafe.Pointer(it + i)) == ' ' {
				syntax_error_pos(curr.Pos, "With '-vet-tabs', tabs must be used for indentation")
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// allow_field_separator
// ---------------------------------------------------------------------------

func allow_field_separator(f *AstFile) bool {
	tok := f.CurrToken
	if allow_token(f, TokenComma) {
		return true
	}
	if tok.Kind == TokenSemicolon {
		ok := false
		if file_allow_newline(f) && token_is_newline(tok) {
			next := peek_token(f).Kind
			switch next {
			case TokenCloseBrace, TokenCloseParen:
				ok = true
			}
		}
		if !ok {
			p := token_to_string(tok)
			syntax_error_pos(token_end_of_line(f, f.PrevToken), "Expected a comma, got a %s", goStr(p))
		}
		advance_token(f)
		return true
	}
	return false
}

func token_is_newline(tok Token) bool {
	return tok.Kind == TokenSemicolon && tok.String.Len == 1 && *tok.String.Data == '\n'
}

// ---------------------------------------------------------------------------
// is_foreign_name_valid
// ---------------------------------------------------------------------------

func is_foreign_name_valid(name String) bool {
	if name.Len == 0 {
		return false
	}
	if name.Data[0] == '_' {
		return true
	}
	return string_is_valid_identifier(name)
}

// ---------------------------------------------------------------------------
// parse_proc_tags
// ---------------------------------------------------------------------------

func parse_proc_tags(f *AstFile, tags *uint64) {
	for f.CurrToken.Kind == TokenHash {
		hash := expect_token(f, TokenHash)
		name := expect_token(f, TokenIdent)
		tag := name.String
		var tag_flag uint64 = 0
		if tag == "require_results" {
			tag_flag = uint64(ProcTagRequireResults)
		} else if tag == "optional_ok" {
			tag_flag = uint64(ProcTagOptionalOk)
		} else if tag == "optional_allocator_error" {
			tag_flag = uint64(ProcTagOptionalAllocatorError)
		} else if tag == "bounds_check" {
			tag_flag = uint64(ProcTagBoundsCheck)
		} else if tag == "no_bounds_check" {
			tag_flag = uint64(ProcTagNoBoundsCheck)
		} else if tag == "type_assert" {
			tag_flag = uint64(ProcTagTypeAssert)
		} else if tag == "no_type_assert" {
			tag_flag = uint64(ProcTagNoTypeAssert)
		} else {
			syntax_error_pos(hash.Pos, "Unknown procedure tag '#%.*s'", tag.Len, goStr(tag))
			continue
		}
		if (*tags & tag_flag) != 0 {
			syntax_error_pos(hash.Pos, "Duplicate procedure tag '#%.*s'", tag.Len, goStr(tag))
		}
		*tags |= tag_flag
	}
}

// ---------------------------------------------------------------------------
// check_polymorphic_params_for_type
// ---------------------------------------------------------------------------

func check_polymorphic_params_for_type(f *AstFile, params *Ast, tok Token) {
	if params == nil {
		return
	}
	if params.Kind != AstFieldList {
		return
	}
	for _, field := range params.FieldList.List {
		fld := &field.Field
		for _, name := range fld.Names {
			if name.Kind == AstPolyType {
				if name.PolyType.Specialization != nil {
					syntax_error(name, "Polymorphic type parameters for types cannot have specializations")
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// ast_on_same_line
// ---------------------------------------------------------------------------

func ast_on_same_line(a, b *Ast) bool {
	if a == nil || b == nil {
		return false
	}
	return ast_token(a).Pos.Line == ast_end_token(b).Pos.Line
}

// ---------------------------------------------------------------------------
// convert_stmt_to_expr
// ---------------------------------------------------------------------------

func convert_stmt_to_expr(f *AstFile, stmt *Ast, msg string) *Ast {
	if stmt == nil {
		return nil
	}
	if stmt.Kind == AstEmptyStmt {
		return nil
	}
	if stmt.Kind == AstBadStmt {
		return nil
	}
	if stmt.Kind == AstExprStmt {
		return stmt.ExprStmt.Expr
	}
	if stmt.Kind == AstAssignStmt {
		opStr := stmt.AssignStmt.Op.String
		syntax_error(stmt, "Expected %s, got an assignment '%s'", msg, goStr(opStr))
		return nil
	}
	if stmt.Kind == AstBlockStmt {
		syntax_error(stmt, "Expected %s, got a block statement", msg)
		return nil
	}
	syntax_error(stmt, "Expected %s, got '%s'", msg, goStr(astStrings[stmt.Kind]))
	return nil
}

// ---------------------------------------------------------------------------
// convert_stmt_to_body
// ---------------------------------------------------------------------------

func convert_stmt_to_body(f *AstFile, stmt *Ast) *Ast {
	if stmt == nil {
		return ast_block_stmt(f, nil, Token{}, Token{})
	}
	if stmt.Kind == AstBlockStmt {
		return stmt
	}
	stmts := []*Ast{stmt}
	return ast_block_stmt(f, stmts, Token{}, Token{})
}

// ---------------------------------------------------------------------------
// unparen_expr
// ---------------------------------------------------------------------------

func unparen_expr(node *Ast) *Ast {
	if node == nil {
		return nil
	}
	if node.Kind == AstParenExpr {
		return unparen_expr(node.ParenExpr.Expr)
	}
	return node
}

// ---------------------------------------------------------------------------
// unselector_expr
// ---------------------------------------------------------------------------

func unselector_expr(node *Ast) *Ast {
	if node == nil {
		return nil
	}
	if node.Kind == AstSelectorExpr {
		return unparen_expr(node.SelectorExpr.Expr)
	}
	return node
}

// ---------------------------------------------------------------------------
// strip_or_return_expr
// ---------------------------------------------------------------------------

func strip_or_return_expr(node *Ast) *Ast {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case AstOrReturnExpr:
		return node.OrReturnExpr.Expr
	case AstOrBranchExpr:
		return node.OrBranchExpr.Expr
	}
	return node
}

// ---------------------------------------------------------------------------
// parser_check_polymorphic_record_parameters
// ---------------------------------------------------------------------------

func parser_check_polymorphic_record_parameters(f *AstFile, params *Ast) {
	if params == nil {
		return
	}
	if params.Kind != AstFieldList {
		return
	}
	for _, field := range params.FieldList.List {
		fld := &field.Field
		for _, name := range fld.Names {
			if name.Kind == AstPolyType {
				pt := &name.PolyType
				if pt.Type != nil {
					syntax_error(name, "Polymorphic parameter type has a type but only specializations are allowed in polymorphic parameter lists for records")
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// parse_ident
// ---------------------------------------------------------------------------

func parse_ident(f *AstFile, allowPolyNames ...bool) *Ast {
	allowPoly := len(allowPolyNames) > 0 && allowPolyNames[0]
	tok := f.CurrToken
	switch tok.Kind {
	case TokenIdent:
		advance_token(f)
		return ast_ident(f, tok)
	case TokenContext:
		advance_token(f)
		return ast_implicit(f, tok)
	case TokenDollar:
		if allowPoly {
			dollar := expect_token(f, TokenDollar)
			name := expect_token(f, TokenIdent)
			spec := parse_type(f)
			if spec != nil && spec.Kind == AstBadExpr {
				spec = nil
			}
			if spec != nil {
				return ast_poly_type(f, dollar, name, spec, nil)
			}
			return ast_poly_type(f, dollar, name, nil, nil)
		}
		fallthrough
	default:
		return ast_ident(f, blank_token)
	}
}

// ---------------------------------------------------------------------------
// parse_tag_expr
// ---------------------------------------------------------------------------

func parse_tag_expr(f *AstFile, expression *Ast) *Ast {
	if f.CurrToken.Kind != TokenHash {
		return expression
	}
	hash := expect_token(f, TokenHash)
	name := expect_token(f, TokenIdent)
	return ast_tag_expr(f, hash, name, expression)
}

// ---------------------------------------------------------------------------
// parse_literal_value
// ---------------------------------------------------------------------------

func parse_literal_value(f *AstFile, typ *Ast) *Ast {
	prev_allow_newline := f.AllowNewline
	f.AllowNewline = file_allow_newline(f)
	prev_expr_level := f.ExprLevel
	f.ExprLevel = 0
	if f.CurrToken.Kind != TokenOpenBrace {
		f.AllowNewline = prev_allow_newline
		f.ExprLevel = prev_expr_level
		return nil
	}
	open := expect_token(f, TokenOpenBrace)
	if f.CurrToken.Kind == TokenCloseBrace {
		close := expect_token(f, TokenCloseBrace)
		f.AllowNewline = prev_allow_newline
		f.ExprLevel = prev_expr_level
		return ast_compound_lit(f, typ, nil, open, close, nil)
	}
	elems := parse_element_list(f)
	close := expect_closing_brace_of_field_list(f)
	f.AllowNewline = prev_allow_newline
	f.ExprLevel = prev_expr_level
	return ast_compound_lit(f, typ, elems, open, close, nil)
}

// ---------------------------------------------------------------------------
// parse_value
// ---------------------------------------------------------------------------

func parse_value(f *AstFile) *Ast {
	return parse_expr(f, false)
}

// ---------------------------------------------------------------------------
// parse_element_list
// ---------------------------------------------------------------------------

func parse_element_list(f *AstFile) []*Ast {
	prev_allow_newline := f.AllowNewline
	f.AllowNewline = file_allow_newline(f)
	list := make([]*Ast, 0)
	for {
		if f.CurrToken.Kind == TokenCloseBrace || f.CurrToken.Kind == TokenCloseParen || f.CurrToken.Kind == TokenEOF {
			break
		}
		parse_enforce_tabs(f)
		elem := parse_expr(f, false)
		list = append(list, elem)
		if !allow_field_separator(f) {
			break
		}
	}
	f.AllowNewline = prev_allow_newline
	return list
}

// ---------------------------------------------------------------------------
// consume_line_comment
// ---------------------------------------------------------------------------

func consume_line_comment(f *AstFile) *CommentGroup {
	cg := f.LineComment
	f.LineComment = nil
	return cg
}

// ---------------------------------------------------------------------------
// parse_enum_field_list
// ---------------------------------------------------------------------------

func parse_enum_field_list(f *AstFile) []*Ast {
	list := make([]*Ast, 0)
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		parse_enforce_tabs(f)
		docs := f.LeadComment
		name := parse_ident(f)
		for allow_token(f, TokenComma) {
			_ = parse_ident(f)
			syntax_error(name, "'enum' fields do not support multiple names per field")
		}
		var value *Ast = nil
		if allow_token(f, TokenEq) {
			value = parse_expr(f, true)
		}
		comment := consume_line_comment(f)
		efv := ast_enum_field_value(f, name, value, docs, comment)
		list = append(list, efv)
		if !allow_field_separator(f) {
			break
		}
	}
	return list
}

// ---------------------------------------------------------------------------
// parse_union_variant_list
// ---------------------------------------------------------------------------

func parse_union_variant_list(f *AstFile) []*Ast {
	list := make([]*Ast, 0)
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		parse_enforce_tabs(f)
		docs := f.LeadComment
		name := parse_ident(f)
		for allow_token(f, TokenComma) {
			_ = parse_ident(f)
			syntax_error(name, "'union' variants do not support multiple names per variant")
		}
		var typ *Ast = nil
		if allow_token(f, TokenColon) {
			typ = parse_type(f)
		}
		var tag Token
		if f.CurrToken.Kind == TokenString {
			tag = expect_token(f, TokenString)
		}
		comment := consume_line_comment(f)
		field := ast_field(f, []*Ast{name}, typ, nil, 0, tag, docs, comment)
		list = append(list, field)
		if !allow_field_separator(f) {
			break
		}
	}
	return list
}

// ---------------------------------------------------------------------------
// init_ast_file
// ---------------------------------------------------------------------------

func init_ast_file(f *AstFile, fullpath String, errPos *TokenPos) ParseFileError {
	f.Fullpath = string_trim_whitespace(fullpath)
	f.Filename = remove_directory_from_path(f.Fullpath)
	f.Directory = directory_from_path(f.Fullpath)
	set_file_path_string(f.ID, f.Fullpath)
	thread_safe_set_ast_file_from_id(f.ID, f)

	ext := S(".odin")
	if !string_ends_with(f.Fullpath, ext) {
		return ParseFileWrongExtension
	}

	var tok Tokenizer
	tok.CurrFileID = f.ID
	f.Tokenizer = tok

	err := init_tokenizer_from_fullpath(&f.Tokenizer, f.Fullpath, buildContext.CopyFileContents)
	if err != TokenizerInitNone {
		switch err {
		case TokenizerInitEmpty:
			break
		case TokenizerInitNotExists:
			return ParseFileNotFound
		case TokenizerInitPermission:
			return ParseFilePermission
		case TokenizerInitFileTooLarge:
			return ParseFileFileTooLarge
		default:
			return ParseFileInvalidFile
		}
	}

	file_size := f.Tokenizer.End - f.Tokenizer.Start
	token_cap := file_size / 3
	pow2_cap := prev_pow2(i64(token_cap)) / 2
	if pow2_cap < 16 {
		pow2_cap = 16
	}
	token_cap = ((token_cap + pow2_cap - 1) / pow2_cap) * pow2_cap
	init_token_cap := token_cap
	if init_token_cap < 16 {
		init_token_cap = 16
	}
	f.Tokens = make([]Token, 0, init_token_cap)

	if err == TokenizerInitEmpty {
		tok0 := Token{Kind: TokenEOF}
		tok0.Pos.FileID = f.ID
		tok0.Pos.Line = 1
		tok0.Pos.Column = 1
		f.Tokens = append(f.Tokens, tok0)
		return ParseFileNone
	}

	start_ts := time_stamp_time_now()
	for {
		var tok0 Token
		tokenizer_get_token(&f.Tokenizer, &tok0)
		if tok0.Kind == TokenInvalid {
			errPos.Line = tok0.Pos.Line
			errPos.Column = tok0.Pos.Column
			return ParseFileInvalidToken
		}
		f.Tokens = append(f.Tokens, tok0)
		if tok0.Kind == TokenEOF {
			break
		}
	}
	end_ts := time_stamp_time_now()
	f.TimeToTokenize = float64(end_ts-start_ts) / float64(time_stamp__freq())

	f.PrevTokenIndex = 0
	f.CurrTokenIndex = 0
	f.PrevToken = f.Tokens[f.PrevTokenIndex]
	f.CurrToken = f.Tokens[f.CurrTokenIndex]
	f.Comments = make([]*CommentGroup, 0)
	f.Imports = make([]*Ast, 0)
	f.CurrProc = nil
	return ParseFileNone
}

// ---------------------------------------------------------------------------
// init_parser / destroy_parser
// ---------------------------------------------------------------------------

func init_parser(p *Parser) bool {
	string_set_init(&p.ImportedFiles)
	p.Packages = make([]*AstPackage, 0)
	return true
}

func destroy_parser(p *Parser) {
	for _, pkg := range p.Packages {
		for _, file := range pkg.Files {
			destroy_ast_file(file)
		}
		pkg.Files = nil
		pkg.ForeignFiles = nil
	}
	p.Packages = nil
	string_set_destroy(&p.ImportedFiles)
}

func destroy_ast_file(f *AstFile) {
	f.Tokens = nil
	f.Comments = nil
	f.Imports = nil
}

// ---------------------------------------------------------------------------
// parser_add_package
// ---------------------------------------------------------------------------

func parser_add_package(p *Parser, pkg *AstPackage) {
	mutex_lock(&p.PackagesMutex)
	pkg.ID = isize(len(p.Packages) + 1)
	p.Packages = append(p.Packages, pkg)
	mutex_unlock(&p.PackagesMutex)
}

// ---------------------------------------------------------------------------
// try_add_import_path
// ---------------------------------------------------------------------------

func try_add_import_path(p *Parser, path String, rel_path String, pos TokenPos, kind ...PackageKind) *AstPackage {
	pkgKind := PackageNormal
	if len(kind) > 0 {
		pkgKind = kind[0]
	}
	FILE_EXT := S(".odin")

	mutex_lock(&p.ImportedFilesMutex)
	if string_set_update(&p.ImportedFiles, path) {
		mutex_unlock(&p.ImportedFilesMutex)
		return nil
	}
	mutex_unlock(&p.ImportedFilesMutex)

	path = copy_string(permanent_allocator(), path)
	pkg := permanent_alloc_item[AstPackage]()
	pkg.Kind = pkgKind
	pkg.Fullpath = path
	pkg.Files = make([]*AstFile, 0)
	pkg.ForeignFiles = make([]AstForeignFile, 0)

	if pkgKind == PackageInit && !path_is_directory(path) && string_ends_with(path, FILE_EXT) {
		var fi FileInfo
		fi.Name = filename_from_path(path)
		fi.Fullpath = path
		fi.Size = get_file_size(path)
		fi.IsDir = false
		pkg.Files = make([]*AstFile, 0, 1)
		pkg.IsSingleFile = true
		parser_add_package(p, pkg)
		parser_add_file_to_process(p, pkg, fi, pos)
		return pkg
	}

	list := make([]FileInfo, 0)
	rd_err := read_directory(path, &list)

	switch rd_err {
	case ReadDirectory_OK:
		break
	case ReadDirectory_InvalidPath:
		syntax_error_pos(pos, "Invalid path: %.*s", rel_path.Len, goStr(rel_path))
		return nil
	case ReadDirectory_NotExists:
		syntax_error_pos(pos, "Path does not exist: %.*s", rel_path.Len, goStr(rel_path))
		return nil
	case ReadDirectory_Permission:
		syntax_error_pos(pos, "Unknown error whilst reading path %.*s", rel_path.Len, goStr(rel_path))
		return nil
	case ReadDirectory_NotDir:
		syntax_error_pos(pos, "Expected a directory for a package, got a file: %.*s", rel_path.Len, goStr(rel_path))
		return nil
	case ReadDirectory_Empty:
		syntax_error_pos(pos, "Empty directory: %.*s", rel_path.Len, goStr(rel_path))
		return nil
	default:
		syntax_error_pos(pos, "Unknown error whilst reading path %.*s", rel_path.Len, goStr(rel_path))
		return nil
	}

	if string_ends_with(path, FILE_EXT) {
		error_pos(pos, "'import' declarations cannot import directories with a .odin extension/suffix")
		return nil
	}

	files_with_ext := isize(0)
	files_to_reserve := isize(1)
	for _, fi := range list {
		name := fi.Name
		ext := path_extension(name)
		if ext == FILE_EXT && !fi.IsDir {
			files_with_ext++
		}
		if ext == FILE_EXT && !is_excluded_target_filename(name) {
			files_to_reserve++
		}
	}

	if files_with_ext == 0 || files_to_reserve == 1 {
		begin_error_block()
		if files_with_ext != 0 {
			syntax_error_pos(pos, "Directory contains no .odin files for the specified platform: %.*s", rel_path.Len, goStr(rel_path))
		} else {
			syntax_error_pos(pos, "Empty directory that contains no .odin files: %.*s", rel_path.Len, goStr(rel_path))
		}
		if build_context.CommandKind&CommandTest != 0 {
			error_line("\tSuggestion: Make an .odin file that imports packages to test and use the `-all-packages` flag.")
		}
		end_error_block()
		return nil
	}

	pkg.Files = make([]*AstFile, 0, files_to_reserve)
	for _, fi := range list {
		name := fi.Name
		ext := path_extension(name)
		if ext == FILE_EXT && !fi.IsDir {
			if is_excluded_target_filename(name) {
				continue
			}
			parser_add_file_to_process(p, pkg, fi, pos)
		} else if ext == ".S" || ext == ".s" {
			if is_excluded_target_filename(name) {
				continue
			}
			parser_add_foreign_file_to_process(p, pkg, AstForeignFileS, fi, pos)
		}
	}
	parser_add_package(p, pkg)
	return pkg
}

// ---------------------------------------------------------------------------
// parser_add_file_to_process
// ---------------------------------------------------------------------------

func parser_add_file_to_process(p *Parser, pkg *AstPackage, fi FileInfo, pos TokenPos) {
	f := ImportedFile{Pkg: pkg, Fi: fi, Pos: pos, Index: p.FileToProcessCount.Add(1) - 1}
	f.Pos.FileID = int32(f.Index + 1)
	wd := permanent_alloc_item[ParserWorkerData]()
	wd.Parser = p
	wd.ImportedFile = f
	thread_pool_add_task(parser_worker_proc, unsafe.Pointer(wd))
}

// ---------------------------------------------------------------------------
// parser_add_foreign_file_to_process
// ---------------------------------------------------------------------------

func parser_add_foreign_file_to_process(p *Parser, pkg *AstPackage, kind AstForeignFileKind, fi FileInfo, pos TokenPos) {
	f := ImportedFile{Pkg: pkg, Fi: fi, Pos: pos, Index: p.FileToProcessCount.Add(1) - 1}
	f.Pos.FileID = int32(f.Index + 1)
	wd := permanent_alloc_item[ForeignFileWorkerData]()
	wd.Parser = p
	wd.ImportedFile = f
	wd.ForeignKind = kind
	thread_pool_add_task(foreign_file_worker_proc, unsafe.Pointer(wd))
}

// ---------------------------------------------------------------------------
// parser_worker_proc
// ---------------------------------------------------------------------------

func parser_worker_proc(data unsafe.Pointer) isize {
	wd := (*ParserWorkerData)(data)
	err := process_imported_file(wd.Parser, wd.ImportedFile)
	if err != ParseFileNone {
		node := permanent_alloc_item[ParseFileErrorNode]()
		node.Err = err
		mutex_lock(&wd.Parser.FileErrorMutex)
		if wd.Parser.FileErrorTail != nil {
			wd.Parser.FileErrorTail.Next = node
			node.Prev = wd.Parser.FileErrorTail
		}
		wd.Parser.FileErrorTail = node
		if wd.Parser.FileErrorHead == nil {
			wd.Parser.FileErrorHead = node
		}
		mutex_unlock(&wd.Parser.FileErrorMutex)
	}
	return isize(err)
}

// ---------------------------------------------------------------------------
// foreign_file_worker_proc
// ---------------------------------------------------------------------------

func foreign_file_worker_proc(data unsafe.Pointer) isize {
	wd := (*ForeignFileWorkerData)(data)
	imp := wd.ImportedFile
	pkg := imp.Pkg
	foreign_file := AstForeignFile{Kind: wd.ForeignKind}
	fullpath := string_trim_whitespace(imp.Fi.Fullpath)
	c_str := alloc_cstring(temporary_allocator(), fullpath)
	fc := gb_file_read_contents(permanent_allocator(), true, c_str)
	foreign_file.Source = String{Data: (*byte)(fc.Data), Len: fc.Size}

	mutex_lock(&pkg.ForeignFilesMutex)
	pkg.ForeignFiles = append(pkg.ForeignFiles, foreign_file)
	mutex_unlock(&pkg.ForeignFilesMutex)
	return 0
}

func alloc_cstring(a gbAllocator, s String) string {
	return string(s.Data[:s.Len])
}

// ---------------------------------------------------------------------------
// process_imported_file
// ---------------------------------------------------------------------------

func process_imported_file(p *Parser, imported_file ImportedFile) ParseFileError {
	pkg := imported_file.Pkg
	fi := imported_file.Fi
	pos := imported_file.Pos
	file := permanent_alloc_item[AstFile]()
	file.Pkg = pkg
	file.ID = int32(imported_file.Index + 1)
	err_pos := TokenPos{}
	err := init_ast_file(file, fi.Fullpath, &err_pos)
	err_pos.FileID = file.ID
	file.LastError = err

	if err != ParseFileNone {
		if err == ParseFileEmptyFile {
			if fi.Fullpath == p.InitFullpath {
				syntax_error_pos(pos, "Initial file is empty - %.*s\n", p.InitFullpath.Len, goStr(p.InitFullpath))
				exit_with_errors()
			}
		} else {
			switch err {
			case ParseFileWrongExtension:
				syntax_error_pos(pos, "Failed to parse file: %.*s; invalid file extension: File must have the extension '.odin'", fi.Name.Len, goStr(fi.Name))
			case ParseFileInvalidFile:
				syntax_error_pos(pos, "Failed to parse file: %.*s; invalid file or cannot be found", fi.Name.Len, goStr(fi.Name))
			case ParseFilePermission:
				syntax_error_pos(pos, "Failed to parse file: %.*s; file permissions problem", fi.Name.Len, goStr(fi.Name))
			case ParseFileNotFound:
				syntax_error_pos(pos, "Failed to parse file: %.*s; file cannot be found ('%s')", fi.Name.Len, goStr(fi.Name), fi.Fullpath.Len, goStr(fi.Fullpath))
			case ParseFileInvalidToken:
				syntax_error_pos(err_pos, "Failed to parse file: %.*s; invalid token found in file", fi.Name.Len, goStr(fi.Name))
			case ParseFileEmptyFile:
				syntax_error_pos(pos, "Failed to parse file: %.*s; file contains no tokens", fi.Name.Len, goStr(fi.Name))
			case ParseFileFileTooLarge:
				syntax_error_pos(pos, "Failed to parse file: %.*s; file is too large, exceeds maximum file size of 2 GiB", fi.Name.Len, goStr(fi.Name))
			}
			return err
		}
	}

	{
		name := file.Fullpath
		name = remove_directory_from_path(name)
		name = remove_extension_from_path(name)
		if string_starts_with(name, S("_")) {
			syntax_error_pos(pos, "Files cannot start with '_', got '%s'", file.Fullpath.Len, goStr(file.Fullpath))
		}
	}

	if build_context.CommandKind&CommandTest != 0 {
		// consume the name (C++ had empty body but kept the variable reference)
	}

	if parse_file(p, file) {
		mutex_lock(&pkg.FilesMutex)
		pkg.Files = append(pkg.Files, file)
		mutex_unlock(&pkg.FilesMutex)

		mutex_lock(&pkg.NameMutex)
		if pkg.Name.Len == 0 {
			pkg.Name = file.PackageName
		} else if pkg.Name != file.PackageName {
			if len(file.Tokens) > 0 && file.Tokens[0].Kind != TokenEOF {
				tok := file.PackageToken
				tok.Pos.FileID = file.ID
				if tok.Pos.Line < 1 {
					tok.Pos.Line = 1
				}
				if tok.Pos.Column < 1 {
					tok.Pos.Column = 1
				}
				syntax_error_pos(tok.Pos, "Different package name, expected '%.*s', got '%.*s'", pkg.Name.Len, goStr(pkg.Name), file.PackageName.Len, goStr(file.PackageName))
			}
		}
		mutex_unlock(&pkg.NameMutex)

		p.TotalLineCount.Add(int64(file.Tokenizer.LineCount))
		p.TotalTokenCount.Add(int64(len(file.Tokens)))
	}
	return ParseFileNone
}

// ---------------------------------------------------------------------------
// parse_file
// ---------------------------------------------------------------------------

func parse_file(p *Parser, f *AstFile) bool {
	if len(f.Tokens) == 0 {
		return true
	}
	if len(f.Tokens) > 0 && f.Tokens[0].Kind == TokenEOF {
		return true
	}

	start := time_stamp_time_now()
	filepath := f.Tokenizer.Fullpath
	base_dir := dir_from_path(filepath)

	tags := make([]Token, 0)
	first_invalid_token_set := false
	var first_invalid_token Token

	for f.CurrToken.Kind != TokenPackage && f.CurrToken.Kind != TokenEOF {
		if f.CurrToken.Kind == TokenComment {
			consume_comment_groups(f, f.PrevToken)
		} else if f.CurrToken.Kind == TokenFileTag {
			tags = append(tags, f.CurrToken)
			advance_token(f)
		} else {
			if !first_invalid_token_set {
				first_invalid_token_set = true
				first_invalid_token = f.CurrToken
			}
			advance_token(f)
		}
	}

	if f.CurrToken.Kind != TokenPackage {
		begin_error_block()
		t := first_invalid_token
		if !first_invalid_token_set {
			t = f.CurrToken
		}
		syntax_error_pos(t.Pos, "Expected a package declaration at the beginning of the file")
		if f.Pkg != nil && f.Pkg.Name != "" {
			error_line("\tSuggestion: Add 'package %.*s' to the top of the file\n", f.Pkg.Name.Len, goStr(f.Pkg.Name))
		}
		end_error_block()
		return false
	}

	if first_invalid_token_set {
		syntax_error_pos(first_invalid_token.Pos, "Expected only comments or lines starting with '#+' before the package declaration")
		return false
	}

	f.PackageToken = expect_token(f, TokenPackage)
	if f.PackageToken.Kind != TokenPackage {
		return false
	}

	package_name := expect_token_after(f, TokenIdent, "package")
	if package_name.Kind == TokenIdent {
		if package_name.String == "_" {
			syntax_error_pos(package_name.Pos, "Invalid package name '_'")
		} else if f.Pkg.Kind != PackageRuntime && package_name.String == "runtime" {
			syntax_error_pos(package_name.Pos, "Use of reserved package name '%.*s'", package_name.String.Len, goStr(package_name.String))
		} else if is_package_name_reserved(package_name.String) {
			syntax_error_pos(package_name.Pos, "Use of reserved package name '%.*s'", package_name.String.Len, goStr(package_name.String))
		}
	}
	f.PackageName = package_name.String

	for _, tok := range tags {
		lt := string_trim_whitespace(substring(tok.String, 2, tok.String.Len))
		if parse_file_tag(lt, tok, f) == false {
			return false
		}
	}

	pd := ast_package_decl(f, f.PackageToken, package_name, f.LeadComment, f.LineComment)
	expect_semicolon(f)
	f.PkgDecl = pd

	if f.ErrorCount == 0 {
		decls := make([]*Ast, 0)
		for f.CurrToken.Kind != TokenEOF {
			stmt := parse_stmt(f)
			if stmt != nil && stmt.Kind != AstEmptyStmt {
				decls = append(decls, stmt)
				if stmt.Kind == AstExprStmt && stmt.ExprStmt.Expr != nil && stmt.ExprStmt.Expr.Kind == AstProcLit {
					syntax_error(stmt, "Procedure literal evaluated but not used")
				}
				f.TotalFileDeclCount += calc_decl_count(stmt)
				if stmt.Kind == AstWhenStmt || stmt.Kind == AstExprStmt || stmt.Kind == AstImportDecl || stmt.Kind == AstForeignBlockDecl {
					f.DelayedDeclCount++
				}
			}
		}
		f.Decls = decls
		parse_setup_file_decls(p, f, base_dir, f.Decls)
	}

	end := time_stamp_time_now()
	f.TimeToParse = float64(end-start) / float64(time_stamp__freq())

	for i := 0; i < int(AstDelayQueueCOUNT); i++ {
		f.DelayedDeclsQueues[i] = make([]*Ast, 0, f.DelayedDeclCount)
	}
	return f.ErrorCount == 0
}

// ---------------------------------------------------------------------------
// parse_file_tag
// ---------------------------------------------------------------------------

func parse_file_tag(lc String, tok Token, f *AstFile) bool {
	if string_starts_with(lc, S("build-project-name")) {
		if !parse_build_project_directory_tag(tok, lc) {
			return false
		}
	} else if string_starts_with(lc, S("build")) {
		if !parse_build_tag(tok, lc) {
			return false
		}
	} else if string_starts_with(lc, S("vet")) {
		f.VetFlags = parse_vet_tag(tok, lc, ast_file_vet_flags(f))
		f.VetFlagsSet = true
	} else if string_starts_with(lc, S("test")) {
		if (build_context.CommandKind & CommandTest) == 0 {
			return false
		}
	} else if string_starts_with(lc, S("ignore")) {
		return false
	} else if string_starts_with(lc, S("private")) {
		f.Flags |= uint32(AstFileIsPrivatePkg)
		rest := string_trim_starts_with(lc, S("private "))
		rest = string_trim_whitespace(rest)
		if lc == "private" {
			f.Flags |= uint32(AstFileIsPrivatePkg)
		} else if rest == "package" {
			f.Flags |= uint32(AstFileIsPrivatePkg)
		} else if rest == "file" {
			f.Flags |= uint32(AstFileIsPrivateFile)
		}
	} else if string_starts_with(lc, S("feature")) {
		f.FeatureFlags |= parse_feature_tag(tok, lc)
		f.FeatureFlagsSet = true
	} else if lc == "lazy" {
		if buildContext.IgnoreLazy {
		} else if f.Pkg.Kind == PackageInit && build_context.CommandKind == CommandDoc {
		} else {
			f.Flags |= uint32(AstFileIsLazy)
		}
	} else if lc == "no-instrumentation" {
		f.Flags |= uint32(AstFileNoInstrumentation)
	} else {
		syntax_error_pos(tok.Pos, "Unknown tag '%s'", goStr(lc))
	}
	return true
}

func string_trim_starts_with(s String, prefix String) String {
	if string_starts_with(s, prefix) {
		return substring(s, prefix.Len, s.Len)
	}
	return s
}

// ---------------------------------------------------------------------------
// parse_setup_file_decls
// ---------------------------------------------------------------------------

func parse_setup_file_decls(p *Parser, f *AstFile, base_dir String, decls []*Ast) {
	for i := 0; i < len(decls); i++ {
		node := decls[i]
		if !is_ast_decl(node) && node.Kind != AstWhenStmt && node.Kind != AstBadStmt && node.Kind != AstEmptyStmt {
			if node.Kind == AstExprStmt {
				expr := node.ExprStmt.Expr
				if expr != nil && expr.Kind == AstCallExpr && expr.CallExpr.Proc != nil && expr.CallExpr.Proc.Kind == AstBasicDirective {
					f.DirectiveCount++
					continue
				}
			}
			syntax_error(node, "Only declarations are allowed at file scope, got '%s'", goStr(astStrings[node.Kind]))
		} else if node.Kind == AstImportDecl {
			id := &node.ImportDecl
			original_string := string_trim_whitespace(string_value_from_token(f, id.Relpath))
			import_path := String{}
			ok := determine_path_from_string(&p.FileDeclMutex, node, base_dir, original_string, &import_path)
			if !ok {
				decls[i] = ast_bad_decl(f, id.Relpath, id.Relpath)
				continue
			}
			import_path = string_trim_whitespace(import_path)
			id.Fullpath = import_path
			if is_package_name_reserved(import_path) {
				continue
			}
			try_add_import_path(p, import_path, original_string, ast_token(node).Pos)
		} else if node.Kind == AstForeignImportDecl {
			fl := &node.ForeignImportDecl
			if len(fl.Filepaths) == 0 {
				syntax_error(decls[i], "No foreign paths found")
				if len(fl.Filepaths) > 0 {
					decls[i] = ast_bad_decl(f, ast_token(fl.Filepaths[0]).Pos, ast_end_token(fl.Filepaths[len(fl.Filepaths)-1]).Pos)
				} else {
					decls[i] = ast_bad_decl(f, Token{}, Token{})
				}
				goto end_label
			} else if !fl.MultipleFilepaths && len(fl.Filepaths) == 1 {
				fp := fl.Filepaths[0]
				fp_token := fp.BasicLit.Token
				file_str := string_trim_whitespace(string_value_from_token(f, fp_token))
				fullpath := file_str
				if !is_arch_wasm() || string_ends_with(fullpath, S(".o")) {
					foreign_path := String{}
					ok := determine_path_from_string(&p.FileDeclMutex, node, base_dir, file_str, &foreign_path)
					if !ok {
						decls[i] = ast_bad_decl(f, fp_token, fp_token)
						goto end_label
					}
					fullpath = foreign_path
				}
				fl.Paths = make([]String, 1)
				fl.Paths[0] = fullpath
			}
		} else if node.Kind == AstWhenStmt {
			ws := &node.WhenStmt
			parse_setup_file_when_stmt(p, f, base_dir, ws)
		}
	end_label:
	}
}

// ---------------------------------------------------------------------------
// parse_setup_file_when_stmt
// ---------------------------------------------------------------------------

func parse_setup_file_when_stmt(p *Parser, f *AstFile, base_dir String, ws *AstWhenStmt) {
	if ws.Body != nil {
		stmts := ws.Body.BlockStmt.Stmts
		parse_setup_file_decls(p, f, base_dir, stmts)
	}
	if ws.ElseStmt != nil {
		switch ws.ElseStmt.Kind {
		case AstBlockStmt:
			stmts := ws.ElseStmt.BlockStmt.Stmts
			parse_setup_file_decls(p, f, base_dir, stmts)
		case AstWhenStmt:
			parse_setup_file_when_stmt(p, f, base_dir, &ws.ElseStmt.WhenStmt)
		}
	}
}

// ---------------------------------------------------------------------------
// dir_from_path
// ---------------------------------------------------------------------------

func dir_from_path(path String) String {
	base_dir := path
	for i := path.Len - 1; i >= 0; i-- {
		if base_dir.Data[i] == '\\' || base_dir.Data[i] == '/' {
			break
		}
		base_dir.Len--
	}
	return base_dir
}

// ---------------------------------------------------------------------------
// calc_decl_count
// ---------------------------------------------------------------------------

func calc_decl_count(decl *Ast) isize {
	count := isize(0)
	switch decl.Kind {
	case AstBlockStmt:
		for _, stmt := range decl.BlockStmt.Stmts {
			count += calc_decl_count(stmt)
		}
	case AstWhenStmt:
		inner_count := calc_decl_count(decl.WhenStmt.Body)
		if decl.WhenStmt.ElseStmt != nil {
			other := calc_decl_count(decl.WhenStmt.ElseStmt)
			if other > inner_count {
				inner_count = other
			}
		}
		count += inner_count
	case AstValueDecl:
		count = isize(len(decl.ValueDecl.Names))
	case AstForeignBlockDecl:
		count = calc_decl_count(decl.ForeignBlockDecl.Body)
	case AstImportDecl, AstForeignImportDecl:
		count = 1
	}
	return count
}

// ---------------------------------------------------------------------------
// determine_path_from_string
// ---------------------------------------------------------------------------

func determine_path_from_string(file_mutex *BlockingMutex, node *Ast, base_dir String, original_string String, path *String, use_check_errors ...bool) bool {
	check_errors := false
	if len(use_check_errors) > 0 {
		check_errors = use_check_errors[0]
	}

	do_error := syntax_error
	do_warning := func(tok Token, format string, args ...any) {
		warning(tok, format, args...)
	}
	if check_errors {
		do_error = func(node *Ast, format string, args ...any) {
			error(node, format, args...)
		}
		do_warning = func(tok Token, format string, args ...any) {
			warning(tok, format, args...)
		}
	}

	collection_name := String{}
	colon_pos := isize(-1)
	for j := isize(0); j < original_string.Len; j++ {
		if original_string.Data[j] == ':' {
			colon_pos = j
			break
		}
	}

	has_windows_drive := false
	if file_mutex == nil {
		if colon_pos == 1 && original_string.Len > 2 {
			if original_string.Data[2] == '/' || original_string.Data[2] == '\\' {
				colon_pos = -1
				has_windows_drive = true
			}
		}
	}

	file_str := String{}
	if colon_pos == 0 {
		do_error(node, "Expected a collection name")
		return false
	}
	if original_string.Len > 0 && colon_pos > 0 {
		collection_name = substring(original_string, 0, colon_pos)
		file_str = substring(original_string, colon_pos+1, original_string.Len)
	} else {
		file_str = original_string
	}

	if !is_import_path_valid(file_str) {
		do_error(node, "Invalid import path: '%s'", goStr(file_str))
		return false
	}

	if collection_name.Len > 0 {
		if collection_name == "core" {
			replace_with_base := false
			if string_starts_with(file_str, S("runtime")) {
				replace_with_base = true
			} else if string_starts_with(file_str, S("intrinsics")) {
				replace_with_base = true
			} else if string_starts_with(file_str, S("builtin")) {
				replace_with_base = true
			}
			if replace_with_base {
				collection_name = S("base")
			}
			if replace_with_base {
				if false { // ast_file_vet_deprecated placeholder
					do_error(node, "import \"core:%.*s\" has been deprecated in favour of \"base:%.*s\"", file_str.Len, goStr(file_str), file_str.Len, goStr(file_str))
				} else {
					do_warning(ast_token(node), "import \"core:%.*s\" has been deprecated in favour of \"base:%.*s\"", file_str.Len, goStr(file_str), file_str.Len, goStr(file_str))
				}
			}
		}
		if collection_name == "system" {
			if node.Kind != AstForeignImportDecl {
				do_error(node, "The library collection 'system' is restrict for 'foreign import'")
				return false
			} else {
				*path = file_str
				return true
			}
		} else if !find_library_collection_path(collection_name, &base_dir) {
			do_error(node, "Unknown library collection: '%.*s'", collection_name.Len, goStr(collection_name))
			return false
		}
	}

	if is_package_name_reserved(file_str) {
		*path = file_str
		if collection_name == "core" || collection_name == "base" {
			return true
		} else {
			do_error(node, "The package '%.*s' must be imported with the 'base' library collection: 'base:%.*s'", file_str.Len, goStr(file_str), file_str.Len, goStr(file_str))
			return false
		}
	}

	if file_mutex != nil {
		mutex_lock(file_mutex)
	}
	if has_windows_drive {
		*path = file_str
	} else {
		ok := false
		fullpath := string_trim_whitespace(get_fullpath_relative(permanent_allocator(), base_dir, file_str, &ok))
		*path = fullpath
	}
	if file_mutex != nil {
		mutex_unlock(file_mutex)
	}
	return true
}

// ---------------------------------------------------------------------------
// is_import_path_valid
// ---------------------------------------------------------------------------

var illegal_import_runes = []rune{
	'"', '\'', '`',
	'\t', '\r', '\n', '\v', '\f',
	'\\',
	'!', '$', '%', '^', '&', '*', '(', ')', '=',
	'[', ']', '{', '}',
	';', ':', '#',
	'|', ',', '<', '>', '?',
}

func is_import_path_valid(path String) bool {
	if path.Len > 0 {
		curr := isize(0)
		for curr < path.Len {
			r := rune(path.Data[curr])
			width := isize(1)
			if r >= 0x80 {
				width = utf8_decode(path.Data[curr:], path.Len-curr, &r)
				if r == 0xfffd && width == 1 {
					return false
				} else if r == 0xfeff && curr > 0 {
					return false
				}
			}
			for _, illegal := range illegal_import_runes {
				if r == illegal {
					return false
				}
			}
			curr += width
		}
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// parse_build_tag
// ---------------------------------------------------------------------------

func parse_build_tag(token_for_pos Token, s String) bool {
	prefix := S("build")
	if build_require_space_after(s, prefix) {
		syntax_error_pos(token_for_pos.Pos, "Expected a space after #+%.*s", prefix.Len, goStr(prefix))
		return true
	}
	s = string_trim_whitespace(substring(s, prefix.Len, s.Len))
	if s.Len == 0 {
		return true
	}
	any_correct := false
	for s.Len > 0 {
		this_kind_correct := true
		this_kind_os_seen := false
		this_kind_arch_seen := false
		num_tokens := 0
		for {
			p := string_trim_whitespace(build_tag_get_token(s, &s))
			if p.Len == 0 {
				break
			}
			if p == "," {
				break
			}
			is_notted := false
			if p.Data[0] == '!' {
				is_notted = true
				p = substring(p, 1, p.Len)
				if p.Len == 0 {
					syntax_error_pos(token_for_pos.Pos, "Expected a build platform after '!'")
					break
				}
			}
			if p.Len == 0 {
				continue
			}
			if p == "ignore" {
				this_kind_correct = false
				continue
			}
			os := get_target_os_from_string(p, nil, nil)
			arch := get_target_arch_from_string(p)
			num_tokens++
			if num_tokens > 2 || (this_kind_os_seen && os != TargetOsInvalid) || (this_kind_arch_seen && arch != TargetArchInvalid) {
				syntax_error_pos(token_for_pos.Pos, "Invalid build tag: Missing ',' before '%s'. Format: '#+build linux, windows amd64, darwin'", goStr(p))
				break
			}
			if os != TargetOsInvalid {
				this_kind_os_seen = true
				if is_notted {
					this_kind_correct = this_kind_correct && (os != buildContext.Metrics.Os)
				} else {
					this_kind_correct = this_kind_correct && (os == buildContext.Metrics.Os)
				}
			} else if arch != TargetArchInvalid {
				this_kind_arch_seen = true
				if is_notted {
					this_kind_correct = this_kind_correct && (arch != buildContext.Metrics.Arch)
				} else {
					this_kind_correct = this_kind_correct && (arch == buildContext.Metrics.Arch)
				}
			}
			if os == TargetOsInvalid && arch == TargetArchInvalid {
				syntax_error_pos(token_for_pos.Pos, "Invalid build tag platform: %s", goStr(p))
				break
			}
		}
		any_correct = any_correct || this_kind_correct
	}
	return any_correct
}

func build_tag_get_token(s String, out *String) String {
	s = string_trim_whitespace(s)
	n := isize(0)
	for n < s.Len {
		r := rune(s.Data[n])
		width := isize(1)
		if r >= 0x80 {
			width = utf8_decode(s.Data[n:], s.Len-n, &r)
		}
		if n == 0 && r == '!' {
		} else if !rune_is_letter(r) && !rune_is_digit(r) && r != ':' {
			k := n
			if k < 1 {
				k = 1
			} else if n > width {
				k = n
			} else {
				k = width
				if k < 1 {
					k = 1
				}
			}
			*out = substring(s, k, s.Len)
			return substring(s, 0, k)
		}
		n += width
	}
	*out = String{}
	return s
}

func build_require_space_after(s String, prefix String) bool {
	if s.Len == prefix.Len {
		return false
	}
	stripped := string_trim_whitespace(substring(s, prefix.Len, s.Len))
	if s.Data[prefix.Len] != ' ' && stripped.Len != 0 {
		return true
	}
	return false
}

func rune_is_letter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func rune_is_digit(r rune) bool {
	return r >= '0' && r <= '9'
}

type (
	TargetOsKind   int32
	TargetArchKind int32
)

const (
	TargetOsInvalid   TargetOsKind   = 0
	TargetArchInvalid TargetArchKind = 0
)

func get_target_os_from_string(s String, subtarget *Subtarget, subtarget_str *String) TargetOsKind {
	return TargetOsInvalid // stub
}

func get_target_arch_from_string(s String) TargetArchKind {
	return TargetArchInvalid // stub
}

// ---------------------------------------------------------------------------
// parse_vet_tag
// ---------------------------------------------------------------------------

func parse_vet_tag(token_for_pos Token, s String, base_vet_flags uint64) uint64 {
	prefix := S("vet")
	if build_require_space_after(s, prefix) {
		syntax_error_pos(token_for_pos.Pos, "Expected a space after #+%.*s", prefix.Len, goStr(prefix))
		return uint64(1)
	}
	s = string_trim_whitespace(substring(s, prefix.Len, s.Len))
	vet_flags := base_vet_flags
	if s.Len == 0 {
		vet_flags |= uint64(VetFlagAll)
	}
	for s.Len > 0 {
		p := string_trim_whitespace(vet_tag_get_token(s, &s, false))
		if p.Len == 0 {
			break
		}
		is_notted := false
		if p.Data[0] == '!' {
			is_notted = true
			p = substring(p, 1, p.Len)
			if p.Len == 0 {
				syntax_error_pos(token_for_pos.Pos, "Expected a vet flag name after '!'")
				return vet_flags
			}
		}
		flag := get_vet_flag_from_name(p)
		if flag != uint64(VetFlagNONE) {
			if is_notted {
				vet_flags &^= flag
			} else {
				vet_flags |= flag
			}
		} else {
			begin_error_block()
			syntax_error_pos(token_for_pos.Pos, "Invalid vet flag name: %s", goStr(p))
			error_line("\tExpected one of the following\n")
			error_line("\tunused\n")
			error_line("\tunused-variables\n")
			error_line("\tunused-imports\n")
			error_line("\tunused-procedures\n")
			error_line("\tshadowing\n")
			error_line("\tusing-stmt\n")
			error_line("\tusing-param\n")
			error_line("\tstyle\n")
			error_line("\textra\n")
			error_line("\tcast\n")
			error_line("\ttabs\n")
			error_line("\texplicit-allocators\n")
			end_error_block()
			return vet_flags
		}
	}
	return vet_flags
}

func vet_tag_get_token(s String, out *String, allow_colon bool) String {
	s = string_trim_whitespace(s)
	n := isize(0)
	for n < s.Len {
		r := rune(s.Data[n])
		width := isize(1)
		if r >= 0x80 {
			width = utf8_decode(s.Data[n:], s.Len-n, &r)
		}
		if n == 0 && r == '!' {
		} else if !rune_is_letter(r) && !rune_is_digit(r) && r != '-' && !(allow_colon && r == ':') {
			k := n
			if k < 1 {
				k = 1
			} else if n > width {
				k = n
			} else {
				k = width
				if k < 1 {
					k = 1
				}
			}
			*out = substring(s, k, s.Len)
			return substring(s, 0, k)
		}
		n += width
	}
	*out = String{}
	return s
}

type VetFlag uint64

const (
	VetFlagNONE VetFlag = 0
	VetFlagAll  VetFlag = 0xFFFFFFFFFFFFFFFF
	VetFlagTabs VetFlag = 1 << 0
)

func get_vet_flag_from_name(s String) uint64 {
	return 0 // stub
}

// ---------------------------------------------------------------------------
// parse_feature_tag
// ---------------------------------------------------------------------------

func parse_feature_tag(token_for_pos Token, s String) uint64 {
	prefix := S("feature")
	if build_require_space_after(s, prefix) {
		syntax_error_pos(token_for_pos.Pos, "Expected a space after #+%.*s", prefix.Len, goStr(prefix))
		return uint64(1)
	}
	s = string_trim_whitespace(substring(s, prefix.Len, s.Len))
	if s.Len == 0 {
		return 0
	}
	feature_flags := uint64(0)
	feature_not_flags := uint64(0)
	for s.Len > 0 {
		p := string_trim_whitespace(vet_tag_get_token(s, &s, true))
		if p.Len == 0 {
			break
		}
		is_notted := false
		if p.Data[0] == '!' {
			is_notted = true
			p = substring(p, 1, p.Len)
			if p.Len == 0 {
				syntax_error_pos(token_for_pos.Pos, "Expected a feature flag name after '!'")
				return 0
			}
		}
		flag := get_feature_flag_from_name(p)
		if flag != 0 {
			if is_notted {
				feature_not_flags |= flag
			} else {
				feature_flags |= flag
			}
		} else {
			begin_error_block()
			syntax_error_pos(token_for_pos.Pos, "Invalid feature flag name: %s", goStr(p))
			error_line("\tExpected one of the following\n")
			error_line("\tdynamic-literals\n")
			error_line("\tglobal-context\n")
			error_line("\tusing-stmt\n")
			error_line("\tinteger-division-by-zero:trap\n")
			error_line("\tinteger-division-by-zero:zero\n")
			error_line("\tinteger-division-by-zero:self\n")
			error_line("\tinteger-division-by-zero:all-bits\n")
			end_error_block()
			return 0
		}
	}
	res := uint64(0)
	if feature_flags == 0 && feature_not_flags == 0 {
		res = 0
	} else if feature_flags == 0 && feature_not_flags != 0 {
		res = 0 &^ feature_not_flags
	} else if feature_flags != 0 && feature_not_flags == 0 {
		res = feature_flags
	} else {
		res = feature_flags &^ feature_not_flags
	}
	return res
}

func get_feature_flag_from_name(s String) uint64 {
	return 0 // stub
}

// ---------------------------------------------------------------------------
// parse_build_project_directory_tag
// ---------------------------------------------------------------------------

func parse_build_project_directory_tag(token_for_pos Token, s String) bool {
	prefix := S("build-project-name")
	s = string_trim_whitespace(substring(s, prefix.Len, s.Len))
	if s.Len == 0 {
		return true
	}
	any_correct := false
	for s.Len > 0 {
		this_kind_correct := true
		for {
			p := string_trim_whitespace(build_tag_get_token(s, &s))
			if p.Len == 0 {
				break
			}
			if p == "," {
				break
			}
			is_notted := false
			if p.Data[0] == '!' {
				is_notted = true
				p = substring(p, 1, p.Len)
				if p.Len == 0 {
					syntax_error_pos(token_for_pos.Pos, "Expected a build-project-name after '!'")
					break
				}
			}
			if p.Len == 0 {
				continue
			}
			if is_notted {
				this_kind_correct = this_kind_correct && (p != buildContext.ODIN_BUILD_PROJECT_NAME)
			} else {
				this_kind_correct = this_kind_correct && (p == buildContext.ODIN_BUILD_PROJECT_NAME)
			}
		}
		any_correct = any_correct || this_kind_correct
	}
	return any_correct
}

// ---------------------------------------------------------------------------
// parse_packages
// ---------------------------------------------------------------------------

func parse_packages(p *Parser, initFilename String) ParseFileError {
	init_fullpath := path_to_full_path(permanent_allocator(), initFilename)

	if !path_is_directory(init_fullpath) {
		ext := S(".odin")
		if !string_ends_with(init_fullpath, ext) {
			error(nil, "Expected either a directory or a .odin file, got '%s'\n", goStr(initFilename))
			return ParseFileWrongExtension
		}
	} else if init_fullpath.Len != 0 {
		path := init_fullpath
		if path.Data[path.Len-1] == '/' {
			path.Len--
		}
		if (buildContext.CommandKind&CommandDoesBuild) != 0 && buildContext.BuildMode == BuildModeExecutable {
			output_path := path_to_string(temporary_allocator(), buildContext.BuildPaths[BuildPath_Output])
			if path_is_directory(output_path) {
				error(nil, "Please specify the executable name with -out:<string> as a directory exists with the same name in the current working directory")
				return ParseFileDirectoryAlreadyExists
			}
		}
	}

	{
		init_pos := TokenPos{}
		{
			ok := false
			s := get_fullpath_base_collection(permanent_allocator(), S("runtime"), &ok)
			if !ok {
				compiler_error("Unable to find The 'base:runtime' package. Is the ODIN_ROOT set up correctly?")
			}
			try_add_import_path(p, s, s, init_pos, PackageRuntime)
		}
		try_add_import_path(p, init_fullpath, init_fullpath, init_pos, PackageInit)
		p.InitFullpath = init_fullpath

		if buildContext.CommandKind&CommandTest != 0 {
			ok := false
			s := get_fullpath_core_collection(permanent_allocator(), S("testing"), &ok)
			if !ok {
				compiler_error("Unable to find The 'core:testing' package. Is the ODIN_ROOT set up correctly?")
			}
			try_add_import_path(p, s, s, init_pos, PackageNormal)
		}
		for _, extraPath := range buildContext.ExtraPackages {
			fullpath := path_to_full_path(permanent_allocator(), extraPath)
			if !path_is_directory(fullpath) {
				ext := S(".odin")
				if !string_ends_with(fullpath, ext) {
					error(nil, "Expected either a directory or a .odin file, got '%s'\n", goStr(fullpath))
					return ParseFileWrongExtension
				}
			}
			pkg := try_add_import_path(p, fullpath, fullpath, init_pos, PackageNormal)
			if pkg != nil {
				pkg.IsExtra = true
			}
		}
	}

	thread_pool_wait()

	for node := p.FileErrorHead; node != nil; node = node.Next {
		if node.Err != ParseFileNone {
			return node.Err
		}
	}

	for i := isize(len(p.Packages) - 1); i >= 0; i-- {
		pkg := p.Packages[i]
		for j := isize(len(pkg.Files) - 1); j >= 0; j-- {
			file := pkg.Files[j]
			if file.ErrorCount != 0 {
				if file.LastError != ParseFileNone {
					return file.LastError
				}
				return ParseFileGeneralError
			}
		}
	}

	for _, pkg := range p.Packages {
		for _, file := range pkg.Files {
			p.TotalSeenLoadDirectiveCount.Add(file.SeenLoadDirectiveCount.Load())
		}
	}

	gParsingDone.Store(true)
	return ParseFileNone
}

// ---------------------------------------------------------------------------
// path_to_string
// ---------------------------------------------------------------------------

func path_to_string(a gbAllocator, path String) String {
	return copy_string(a, path)
}

// ---------------------------------------------------------------------------
// BuildPath constants
// ---------------------------------------------------------------------------

const BuildPath_Output int = 8

// ---------------------------------------------------------------------------
// Ast constructor stubs
// ---------------------------------------------------------------------------

func ast_ident(f *AstFile, token Token) *Ast
func ast_implicit(f *AstFile, token Token) *Ast
func ast_poly_type(f *AstFile, dollar Token, name Token, typ *Ast, specialization *Ast) *Ast
func ast_bad_expr(f *AstFile, begin Token, end Token) *Ast
func ast_bad_decl(f *AstFile, begin Token, end Token) *Ast
func ast_tag_expr(f *AstFile, hash Token, name Token, expr *Ast) *Ast
func ast_compound_lit(f *AstFile, typ *Ast, elems []*Ast, open Token, close Token, tag *Ast) *Ast
func ast_enum_field_value(f *AstFile, name *Ast, value *Ast, docs *CommentGroup, comment *CommentGroup) *Ast
func ast_field(f *AstFile, names []*Ast, typ *Ast, defaultValue *Ast, flags uint32, tag Token, docs *CommentGroup, comment *CommentGroup) *Ast
func ast_block_stmt(f *AstFile, stmts []*Ast, open Token, close Token) *Ast
func ast_package_decl(f *AstFile, token Token, name Token, docs *CommentGroup, comment *CommentGroup) *Ast

// ---------------------------------------------------------------------------
// Tokenizer functions
// ---------------------------------------------------------------------------

func init_tokenizer_from_fullpath(t *Tokenizer, fullpath String, copyFileContents bool) TokenizerInitError
func init_tokenizer_with_data(t *Tokenizer, data []byte, filename String) TokenizerInitError
func tokenizer_get_token(t *Tokenizer, token *Token)

type TokenizerInitError int

const (
	TokenizerInitNone         TokenizerInitError = 0
	TokenizerInitNotExists    TokenizerInitError = 1
	TokenizerInitInvalid      TokenizerInitError = 2
	TokenizerInitPermission   TokenizerInitError = 3
	TokenizerInitEmpty        TokenizerInitError = 4
	TokenizerInitFileTooLarge TokenizerInitError = 5
)

var loadedFileErrorMapToTokenizer [ParseFileCOUNT]TokenizerInitError

// ---------------------------------------------------------------------------
// FileInfo
// ---------------------------------------------------------------------------

type FileInfo struct {
	Name     String
	Fullpath String
	Size     i64
	IsDir    bool
}

// ---------------------------------------------------------------------------
// Subtarget
// ---------------------------------------------------------------------------

type Subtarget int32

const (
	SubtargetDefault Subtarget = 0
	SubtargetInvalid Subtarget = 0xFFFFFFFF
)

// ---------------------------------------------------------------------------
// String helper
// ---------------------------------------------------------------------------

func S(s string) String {
	data := []byte(s)
	if len(data) == 0 {
		return String{}
	}
	return String{Data: &data[0], Len: isize(len(data))}
}

// ---------------------------------------------------------------------------
// Type aliases
// ---------------------------------------------------------------------------

type (
	u8    = byte
	isize = int64
	i64   = int64
	i32   = int32
	u64   = uint64
	u32   = uint32
)

var EmptyToken Token

// ---------------------------------------------------------------------------
// Error functions for Token
// ---------------------------------------------------------------------------

func error_pos(pos TokenPos, format string, args ...any) {
	_ = pos
	_ = format
	_ = args
	// stub - will be implemented elsewhere
}

func compiler_error(format string, args ...any) {
	_ = format
	_ = args
	// stub
}

func exit_with_errors() {
	// stub
}

func utf8_decode(data []byte, max_len isize, r *rune) isize {
	if max_len <= 0 || len(data) == 0 {
		*r = 0
		return 1
	}
	if data[0] < 0x80 {
		*r = rune(data[0])
		return 1
	}
	var rune_val rune
	var width isize
	if data[0]&0xE0 == 0xC0 {
		rune_val = rune(data[0] & 0x1F)
		width = 2
	} else if data[0]&0xF0 == 0xE0 {
		rune_val = rune(data[0] & 0x0F)
		width = 3
	} else if data[0]&0xF8 == 0xF0 {
		rune_val = rune(data[0] & 0x07)
		width = 4
	} else {
		*r = 0xfffd
		return 1
	}
	for i := isize(1); i < width && i < isize(len(data)); i++ {
		rune_val = (rune_val << 6) | rune(data[i]&0x3F)
	}
	*r = rune_val
	return width
}

func goStr(s String) string {
	if s.Data == nil || s.Len == 0 {
		return ""
	}
	return string(s.Data[:s.Len])
}
