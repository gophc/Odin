package odingo

import "core:strings"
import "core:fmt"
import "core:os"
import "core:sync"
import "core:mem"
import "core:unicode/utf8"
import "core:time"
import "core:slice"
import "core:path/filepath"

ParseFileError :: enum int {
	None                 = 0,
	WrongExtension       = 1,
	NotFound             = 2,
	Permission           = 3,
	FileTooLarge         = 4,
	InvalidFile          = 5,
	InvalidToken         = 6,
	EmptyFile            = 7,
	GeneralError         = 8,
	DirectoryAlreadyExists = 9,
}

illegal_import_runes := []rune {
	'"', '\'', '`',
	'\t', '\r', '\n', '\v', '\f',
	'\\',
	'!', '$', '%', '~', '&', '*', '(', ')', '=',
	'[', ']', '{', '}',
	';',
	':',
	'#',
	'|', ',', '<', '>', '?',
}

is_import_path_valid :: proc(path: string) -> bool {
	if len(path) > 0 {
		curr := 0
		for curr < len(path) {
			r, width := utf8.decode_rune_in_string(path[curr:])
			if r == utf8.RUNE_ERROR && width == 1 {
				return false
			} else if r == utf8.RUNE_BOM && curr > 0 {
				return false
			}
			for illegal_rune in illegal_import_runes {
				if r == illegal_rune {
					return false
				}
			}
			curr += width
		}
		return true
	}
	return false
}

is_build_flag_path_valid :: proc(path: string) -> bool {
	if len(path) > 0 {
		curr := 0
		for curr < len(path) {
			r, width := utf8.decode_rune_in_string(path[curr:])
			if r == utf8.RUNE_ERROR && width == 1 {
				return false
			} else if r == utf8.RUNE_BOM && curr > 0 {
				return false
			}
			for illegal_rune in illegal_import_runes {
				if r == '\\' || r == ':' {
					break
				}
				if r == illegal_rune {
					return false
				}
			}
			curr += width
		}
		return true
	}
	return false
}

is_package_name_reserved :: proc(name: string) -> bool {
	return name == "builtin" || name == "intrinsics"
}

parse_enforce_tabs :: proc(f: ^AstFile) {
	if (ast_file_vet_flags(f) & VetFlag_Tabs) == 0 {
		return
	}
	prev := f.prev_token
	curr := f.curr_token
	if prev.pos.line < curr.pos.line {
		start_ptr := raw_data(f.tokenizer.source)
		start_idx := prev.pos.offset
		end_idx := curr.pos.offset
		it := end_idx
		for it > start_idx {
			if start_ptr[it] == '\n' {
				it += 1
				break
			}
			it -= 1
		}
		seg_len := end_idx - it
		for i in 0 ..< seg_len {
			if start_ptr[it + i] == '/' {
				break
			}
			if start_ptr[it + i] == ' ' {
				syntax_error(curr, "With '-vet-tabs', tabs must be used for indentation")
				break
			}
		}
	}
}

parse_stmt_list :: proc(f: ^AstFile, allocator := context.allocator) -> [dynamic]^Ast {
	list := make([dynamic]^Ast, allocator)
	for f.curr_token.kind != .Case &&
	    f.curr_token.kind != .CloseBrace &&
	    f.curr_token.kind != .EOF {
		parse_enforce_tabs(f)
		stmt := parse_stmt(f)
		if stmt != nil && stmt.kind != .EmptyStmt {
			append(&list, stmt)
			if stmt.kind == .ExprStmt &&
			   stmt.ExprStmt.expr != nil &&
			   stmt.ExprStmt.expr.kind == .ProcLit {
				syntax_error(stmt, "Procedure literal evaluated but not used")
			}
		}
	}
	return list
}

init_ast_file :: proc(f: ^AstFile, fullpath: string, err_pos: ^TokenPos) -> ParseFileError {
	assert(f != nil)

	f.fullpath  = strings.trim_space(fullpath)
	f.filename  = remove_directory_from_path(f.fullpath)
	f.directory = directory_from_path(f.fullpath)
	set_file_path_string(f.id, f.fullpath)
	thread_safe_set_ast_file_from_id(f.id, f)

	if !strings.has_suffix(f.fullpath, ".odin") {
		return .WrongExtension
	}

	mem.zero(&f.tokenizer, size_of(f.tokenizer))
	f.tokenizer.curr_file_id = f.id

	err := init_tokenizer_from_fullpath(&f.tokenizer, f.fullpath, build_context.copy_file_contents)
	if err != .None {
		switch err {
		case .Empty:
		case .NotExists:
			return .NotFound
		case .Permission:
			return .Permission
		case .FileTooLarge:
			return .FileTooLarge
		case:
			return .InvalidFile
		}
	}

	file_size := f.tokenizer.end - f.tokenizer.start
	token_cap := file_size / 3
	pow2_cap := prev_pow2(token_cap) / 2
	if pow2_cap < 16 {
		pow2_cap = 16
	}
	token_cap = ((token_cap + pow2_cap - 1) / pow2_cap) * pow2_cap
	init_token_cap := token_cap if token_cap > 16 else 16
	f.tokens = make([dynamic]Token, context.allocator, 0, init_token_cap if init_token_cap > 16 else 16)

	if err == .Empty {
		token := Token{kind = .EOF}
		token.pos.file_id = f.id
		token.pos.line = 1
		token.pos.column = 1
		append(&f.tokens, token)
		return .None
	}

	start_time := time.tick_now()
	for {
		token: Token
		tokenizer_get_token(&f.tokenizer, &token)
		append(&f.tokens, token)
		if token.kind == .Invalid {
			err_pos.line   = token.pos.line
			err_pos.column = token.pos.column
			return .InvalidToken
		}
		if token.kind == .EOF {
			break
		}
	}
	end_time := time.tick_now()
	f.time_to_tokenize = time.duration_seconds(time.tick_diff(start_time, end_time))
	f.prev_token_index = 0
	f.curr_token_index = 0
	f.prev_token = f.tokens[f.prev_token_index]
	f.curr_token = f.tokens[f.curr_token_index]
	f.comments = make([dynamic]CommentGroup, context.allocator)
	f.imports  = make([dynamic]AstImportEntry, context.allocator)
	f.curr_proc = nil
	return .None
}

destroy_ast_file :: proc(f: ^AstFile) {
	assert(f != nil)
	delete(f.tokens)
	delete(f.comments)
	delete(f.imports)
}

init_parser :: proc(p: ^Parser) -> bool {
	assert(p != nil)
	p.imported_files = make(map[string]struct{})
	p.packages = make([dynamic]^AstPackage, permanent_allocator())
	return true
}

destroy_parser :: proc(p: ^Parser) {
	assert(p != nil)
	for pkg in p.packages {
		for file in pkg.files {
			destroy_ast_file(file)
		}
		delete(pkg.files)
		delete(pkg.foreign_files)
	}
	delete(p.packages)
	delete(p.imported_files)
}

parser_add_package :: proc(p: ^Parser, pkg: ^AstPackage) {
	sync.mutex_lock(&p.packages_mutex)
	defer sync.mutex_unlock(&p.packages_mutex)
	pkg.id = len(p.packages) + 1
	append(&p.packages, pkg)
}

parser_worker_proc :: proc(data: rawptr) -> int {
	wd := cast(^ParserWorkerData)data
	err := process_imported_file(wd.parser, wd.imported_file)
	if err != .None {
		node := new_clone(ParseFileErrorNode{err = err})
		sync.mutex_lock(&wd.parser.file_error_mutex)
		defer sync.mutex_unlock(&wd.parser.file_error_mutex)
		if wd.parser.file_error_tail != nil {
			wd.parser.file_error_tail.next = node
		}
		wd.parser.file_error_tail = node
		if wd.parser.file_error_head == nil {
			wd.parser.file_error_head = node
		}
	}
	return int(err)
}

parser_add_file_to_process :: proc(p: ^Parser, pkg: ^AstPackage, fi: FileInfo, pos: TokenPos) {
	f := ImportedFile{pkg = pkg, fi = fi, pos = pos, index = p.file_to_process_count}
	p.file_to_process_count += 1
	f.pos.file_id = i32(f.index + 1)
	wd := new_clone(ParserWorkerData{parser = p, imported_file = f})
	thread_pool_add_task(parser_worker_proc, wd)
}

foreign_file_worker_proc :: proc(data: rawptr) -> int {
	wd := cast(^ForeignFileWorkerData)data
	imp := &wd.imported_file
	pkg := imp.pkg
	foreign_file := AstForeignFile{kind = wd.foreign_kind}
	fullpath := strings.trim_space(imp.fi.fullpath)
	c_str := strings.clone_to_cstring(fullpath, context.temp_allocator)
	fc, ok := os.read_entire_file_from_filename(c_str, context.allocator)
	if ok {
		foreign_file.source.text = transmute(^u8)raw_data(fc)
		foreign_file.source.len  = len(fc)
	}
	sync.mutex_lock(&pkg.foreign_files_mutex)
	append(&pkg.foreign_files, foreign_file)
	sync.mutex_unlock(&pkg.foreign_files_mutex)
	return 0
}

parser_add_foreign_file_to_process :: proc(p: ^Parser, pkg: ^AstPackage, kind: AstForeignFileKind, fi: FileInfo, pos: TokenPos) {
	f := ImportedFile{pkg = pkg, fi = fi, pos = pos, index = p.file_to_process_count}
	p.file_to_process_count += 1
	f.pos.file_id = i32(f.index + 1)
	wd := new_clone(ForeignFileWorkerData{parser = p, imported_file = f, foreign_kind = kind})
	thread_pool_add_task(foreign_file_worker_proc, wd)
}

try_add_import_path :: proc(p: ^Parser, path: string, rel_path: string, pos: TokenPos, kind: PackageKind = .Normal) -> ^AstPackage {
	FILE_EXT :: ".odin"

	{
		sync.mutex_lock(&p.imported_files_mutex)
		defer sync.mutex_unlock(&p.imported_files_mutex)
		if path in p.imported_files {
			return nil
		}
		p.imported_files[path] = {}
	}

	perm_path := strings.clone(path, permanent_allocator())
	pkg := new_clone(AstPackage{})
	pkg.kind = kind
	pkg.fullpath = perm_path
	pkg.files = make([dynamic]^AstFile, permanent_allocator())
	pkg.foreign_files = make([dynamic]AstForeignFile, permanent_allocator())

	if kind == .Init && !path_is_directory(perm_path) && strings.has_suffix(perm_path, FILE_EXT) {
		fi := FileInfo{
			name     = filename_from_path(perm_path),
			fullpath = perm_path,
			size     = get_file_size(perm_path),
			is_dir   = false,
		}
		reserve(&pkg.files, 1)
		pkg.is_single_file = true
		parser_add_package(p, pkg)
		parser_add_file_to_process(p, pkg, fi, pos)
		return pkg
	}

	list := make([dynamic]FileInfo, context.temp_allocator)
	defer delete(list)
	rd_err := read_directory(path, &list)

	if len(list) == 1 {
		assert(path != list[0].fullpath)
	}

	switch rd_err {
	case .InvalidPath:
		syntax_error(pos, "Invalid path: %.*s", len(rel_path), rel_path)
		return nil
	case .NotExists:
		syntax_error(pos, "Path does not exist: %.*s", len(rel_path), rel_path)
		return nil
	case .Permission:
		syntax_error(pos, "Unknown error whilst reading path %.*s", len(rel_path), rel_path)
		return nil
	case .NotDir:
		syntax_error(pos, "Expected a directory for a package, got a file: %.*s", len(rel_path), rel_path)
		return nil
	case .Empty:
		syntax_error(pos, "Empty directory: %.*s", len(rel_path), rel_path)
		return nil
	case .Unknown:
		syntax_error(pos, "Unknown error whilst reading path %.*s", len(rel_path), rel_path)
		return nil
	}

	if strings.has_suffix(perm_path, FILE_EXT) {
		error(pos, "'import' declarations cannot import directories with a .odin extension/suffix")
		return nil
	}

	files_with_ext := 0
	files_to_reserve := 1
	for fi in list {
		ext := filepath.ext(fi.name)
		if ext == FILE_EXT && !fi.is_dir {
			files_with_ext += 1
		}
		if ext == FILE_EXT && !is_excluded_target_filename(fi.name) {
			files_to_reserve += 1
		}
	}

	if files_with_ext == 0 || files_to_reserve == 1 {
		begin_error_block()
		defer end_error_block()
		if files_with_ext != 0 {
			syntax_error(pos, "Directory contains no .odin files for the specified platform: %.*s", len(rel_path), rel_path)
		} else {
			syntax_error(pos, "Empty directory that contains no .odin files: %.*s", len(rel_path), rel_path)
		}
		if build_context.command_kind == .Test {
			error_line("\tSuggestion: Make an .odin file that imports packages to test and use the `-all-packages` flag.")
		}
		return nil
	}

	reserve(&pkg.files, files_to_reserve)
	for fi in list {
		ext := filepath.ext(fi.name)
		if ext == FILE_EXT && !fi.is_dir {
			if is_excluded_target_filename(fi.name) {
				continue
			}
			parser_add_file_to_process(p, pkg, fi, pos)
		} else if ext == ".S" || ext == ".s" {
			if is_excluded_target_filename(fi.name) {
				continue
			}
			parser_add_foreign_file_to_process(p, pkg, .S, fi, pos)
		}
	}
	parser_add_package(p, pkg)
	return pkg
}

determine_path_from_string :: proc(file_mutex: ^sync.Mutex, node: ^Ast, base_dir: string, original_string: string, path: ^string, use_check_errors: bool = false) -> bool {
	assert(path != nil)

	collection_name := ""
	colon_pos := -1
	for r, j in original_string {
		if r == ':' {
			colon_pos = j
			break
		}
	}

	has_windows_drive := false
	if file_mutex == nil {
		if colon_pos == 1 && len(original_string) > 2 {
			if original_string[2] == '/' || original_string[2] == '\\' {
				colon_pos = -1
				has_windows_drive = true
			}
		}
	}

	file_str := ""
	if colon_pos == 0 {
		syntax_error(node, "Expected a collection name")
		return false
	}
	if len(original_string) > 0 && colon_pos > 0 {
		collection_name = original_string[:colon_pos]
		file_str = original_string[colon_pos + 1:]
	} else {
		file_str = original_string
	}

	if has_windows_drive {
		sub_file_path := file_str[3:]
		if !is_import_path_valid(sub_file_path) {
			syntax_error(node, "Invalid import path: '%.*s'", len(file_str), file_str)
			return false
		}
	} else if !is_import_path_valid(file_str) {
		syntax_error(node, "Invalid import path: '%.*s'", len(file_str), file_str)
		return false
	}

	if len(collection_name) > 0 {
		if collection_name == "core" {
			replace_with_base := false
			if strings.starts_with(file_str, "runtime") {
				replace_with_base = true
			} else if strings.starts_with(file_str, "intrinsics") {
				replace_with_base = true
			} else if strings.starts_with(file_str, "builtin") {
				replace_with_base = true
			}
			if replace_with_base {
				collection_name = "base"
			}
			if replace_with_base {
				if ast_file_vet_deprecated(node.file()) {
					syntax_error(node, "import \"core:%.*s\" has been deprecated in favour of \"base:%.*s\"", len(file_str), file_str, len(file_str), file_str)
				} else {
					syntax_warning(ast_token(node), "import \"core:%.*s\" has been deprecated in favour of \"base:%.*s\"", len(file_str), file_str, len(file_str), file_str)
				}
			}
		}

		if collection_name == "system" {
			if node.kind != .ForeignImportDecl {
				syntax_error(node, "The library collection 'system' is restricted for 'foreign import'")
				return false
			} else {
				path^ = file_str
				return true
			}
		} else if !find_library_collection_path(collection_name, &base_dir) {
			syntax_error(node, "Unknown library collection: '%.*s'", len(collection_name), collection_name)
			return false
		}
	}

	if is_package_name_reserved(file_str) {
		path^ = file_str
		if collection_name == "core" || collection_name == "base" {
			return true
		} else {
			syntax_error(node, "The package '%.*s' must be imported with the 'base' library collection: 'base:%.*s'", len(file_str), file_str, len(file_str), file_str)
			return false
		}
	}

	if file_mutex != nil {
		sync.mutex_lock(file_mutex)
		defer sync.mutex_unlock(file_mutex)
	}

	if node.kind == .ForeignImportDecl {
		node.ForeignImportDecl.collection_name = collection_name
	}

	if has_windows_drive {
		path^ = file_str
	} else {
		ok: bool
		fullpath := strings.trim_space(get_fullpath_relative(permanent_allocator(), base_dir, file_str, &ok))
		path^ = fullpath
	}
	return true
}

build_tag_get_token :: proc(s: string, out: ^string) -> string {
	s = strings.trim_space(s)
	n := 0
	for n < len(s) {
		r, width := utf8.decode_rune_in_string(s[n:])
		if n == 0 && r == '!' {
			// allow leading !
		} else if !unicode.is_letter(r) && !unicode.is_digit(r) && r != ':' {
			k := max(max(n, width), 1)
			out^ = s[k:]
			return s[:k]
		}
		n += width
	}
	out^ = ""
	return s
}

build_require_space_after :: proc(s: string, prefix: string) -> bool {
	assert(strings.starts_with(s, prefix))
	if len(s) == len(prefix) {
		return false
	}
	stripped := strings.trim_space(s[len(prefix):])
	if s[len(prefix)] != ' ' && len(stripped) != 0 {
		return true
	}
	return false
}

parse_build_tag :: proc(token_for_pos: Token, s: string) -> bool {
	prefix :: "build"
	assert(strings.starts_with(s, prefix))

	if build_require_space_after(s, prefix) {
		syntax_error(token_for_pos, "Expected a space after #+%.*s", len(prefix), prefix)
		return true
	}

	s = strings.trim_space(s[len(prefix):])
	if len(s) == 0 {
		return true
	}

	any_correct := false
	for len(s) > 0 {
		this_kind_correct := true
		this_kind_os_seen := false
		this_kind_arch_seen := false
		num_tokens := 0

		for {
			p := strings.trim_space(build_tag_get_token(s, &s))
			if len(p) == 0 {break}
			if p == "," {break}

			is_notted := false
			p_ref := p
			if p_ref[0] == '!' {
				is_notted = true
				p_ref = p_ref[1:]
				if len(p_ref) == 0 {
					syntax_error(token_for_pos, "Expected a build platform after '!'")
					break
				}
			}
			if len(p_ref) == 0 {
				continue
			}

			if p_ref == "ignore" {
				this_kind_correct = false
				continue
			}

			subtarget: Subtarget = .Invalid
			subtarget_str: string
			os_kind  := get_target_os_from_string(p_ref, &subtarget, &subtarget_str)
			arch_kind := get_target_arch_from_string(p_ref)
			num_tokens += 1

			if num_tokens > 2 || (this_kind_os_seen && os_kind != .Invalid) || (this_kind_arch_seen && arch_kind != .Invalid) {
				syntax_error(token_for_pos, "Invalid build tag: Missing ',' before '%.*s'. Format: '#+build linux, windows amd64, darwin'", len(p_ref), p_ref)
				break
			}

			is_ios_subtarget := false
			if subtarget == .Invalid {
				if !str_eq_ignore_case(subtarget_str, "ios") {
					syntax_error(token_for_pos, "Invalid subtarget '%.*s'.", len(subtarget_str), subtarget_str)
					break
				}
				is_ios_subtarget = true
			}

			if os_kind != .Invalid {
				this_kind_os_seen = true
				is_explicit_default_subtarget := str_eq_ignore_case(subtarget_str, "default")
				same_subtarget := (subtarget == .Default && !is_explicit_default_subtarget) || (subtarget == selected_subtarget)
				if is_ios_subtarget && (selected_subtarget == .iPhone || selected_subtarget == .iPhoneSimulator) {
					same_subtarget = true
				}
				assert(arch_kind == .Invalid)
				if is_notted {
					this_kind_correct = this_kind_correct && (os_kind != build_context.metrics.os || !same_subtarget)
				} else {
					this_kind_correct = this_kind_correct && (os_kind == build_context.metrics.os && same_subtarget)
				}
			} else if arch_kind != .Invalid {
				this_kind_arch_seen = true
				if is_notted {
					this_kind_correct = this_kind_correct && (arch_kind != build_context.metrics.arch)
				} else {
					this_kind_correct = this_kind_correct && (arch_kind == build_context.metrics.arch)
				}
			}

			if os_kind == .Invalid && arch_kind == .Invalid {
				syntax_error(token_for_pos, "Invalid build tag platform: %.*s", len(p_ref), p_ref)
				break
			}
		}
		any_correct = any_correct || this_kind_correct
	}
	return any_correct
}

vet_tag_get_token :: proc(s: string, out: ^string, allow_colon: bool) -> string {
	s = strings.trim_space(s)
	n := 0
	for n < len(s) {
		r, width := utf8.decode_rune_in_string(s[n:])
		if n == 0 && r == '!' {
			// allow leading !
		} else if !unicode.is_letter(r) && !unicode.is_digit(r) && r != '-' && !(allow_colon && r == ':') {
			k := max(max(n, width), 1)
			out^ = s[k:]
			return s[:k]
		}
		n += width
	}
	out^ = ""
	return s
}

parse_vet_tag :: proc(token_for_pos: Token, s: string, base_vet_flags: u64) -> u64 {
	prefix :: "vet"
	assert(strings.starts_with(s, prefix))

	if build_require_space_after(s, prefix) {
		syntax_error(token_for_pos, "Expected a space after #+%.*s", len(prefix), prefix)
		return 0
	}

	s = strings.trim_space(s[len(prefix):])
	vet_flags := base_vet_flags
	if len(s) == 0 {
		vet_flags |= VetFlag_All
	}
	for len(s) > 0 {
		p := strings.trim_space(vet_tag_get_token(s, &s, false))
		if len(p) == 0 {
			break
		}
		is_notted := false
		p_ref := p
		if p_ref[0] == '!' {
			is_notted = true
			p_ref = p_ref[1:]
			if len(p_ref) == 0 {
				syntax_error(token_for_pos, "Expected a vet flag name after '!'")
				return vet_flags
			}
		}
		flag := get_vet_flag_from_name(p_ref)
		if flag != VetFlag_NONE {
			if is_notted {
				vet_flags &~= flag
			} else {
				vet_flags |= flag
			}
		} else {
			begin_error_block()
			defer end_error_block()
			syntax_error(token_for_pos, "Invalid vet flag name: %.*s", len(p_ref), p_ref)
			error_line("\tExpected one of the following")
			error_line("\tunused")
			error_line("\tunused-variables")
			error_line("\tunused-imports")
			error_line("\tunused-procedures")
			error_line("\tshadowing")
			error_line("\tusing-stmt")
			error_line("\tusing-param")
			error_line("\tstyle")
			error_line("\textra")
			error_line("\tcast")
			error_line("\ttabs")
			error_line("\texplicit-allocators")
			return vet_flags
		}
	}
	return vet_flags
}

parse_feature_tag :: proc(token_for_pos: Token, s: string) -> u64 {
	prefix :: "feature"
	assert(strings.starts_with(s, prefix))

	if build_require_space_after(s, prefix) {
		syntax_error(token_for_pos, "Expected a space after #+%.*s", len(prefix), prefix)
		return 0
	}

	s = strings.trim_space(s[len(prefix):])
	if len(s) == 0 {
		return OptInFeatureFlag_NONE
	}

	feature_flags: u64 = 0
	feature_not_flags: u64 = 0

	for len(s) > 0 {
		p := strings.trim_space(vet_tag_get_token(s, &s, true))
		if len(p) == 0 {
			break
		}
		is_notted := false
		p_ref := p
		if p_ref[0] == '!' {
			is_notted = true
			p_ref = p_ref[1:]
			if len(p_ref) == 0 {
				syntax_error(token_for_pos, "Expected a feature flag name after '!'")
				return OptInFeatureFlag_NONE
			}
		}

		flag := get_feature_flag_from_name(p_ref)
		if flag != OptInFeatureFlag_NONE {
			if is_notted {
				feature_not_flags |= flag
			} else {
				feature_flags |= flag
			}
			if is_notted {
				switch flag {
				case .IntegerDivisionByZero_Trap, .IntegerDivisionByZero_Zero, .IntegerDivisionByZero_AllBits:
					syntax_error(token_for_pos, "Feature flag does not support notting with '!' - '%.*s'", len(p_ref), p_ref)
				}
			}
		} else {
			begin_error_block()
			defer end_error_block()
			syntax_error(token_for_pos, "Invalid feature flag name: %.*s", len(p_ref), p_ref)
			error_line("\tExpected one of the following")
			error_line("\tdynamic-literals")
			error_line("\tglobal-context")
			error_line("\tusing-stmt")
			error_line("\tinteger-division-by-zero:trap")
			error_line("\tinteger-division-by-zero:zero")
			error_line("\tinteger-division-by-zero:self")
			error_line("\tinteger-division-by-zero:all-bits")
			return OptInFeatureFlag_NONE
		}
	}

	res: u64 = OptInFeatureFlag_NONE
	if feature_flags == 0 && feature_not_flags == 0 {
		res = OptInFeatureFlag_NONE
	} else if feature_flags == 0 && feature_not_flags != 0 {
		res = OptInFeatureFlag_NONE &~ feature_not_flags
	} else if feature_flags != 0 && feature_not_flags == 0 {
		res = feature_flags
	} else {
		assert(feature_flags != 0 && feature_not_flags != 0)
		res = feature_flags &~ feature_not_flags
	}

	idbz_count := count_set_bits(res & OptInFeatureFlag_IntegerDivisionByZero_ALL)
	if idbz_count > 1 {
		syntax_error(token_for_pos, "Only one integer-division-by-zero feature flag can be enabled")
	}
	return res
}

dir_from_path :: proc(path: string) -> string {
	base_dir := path
	for i := len(path) - 1; i >= 0; i -= 1 {
		if base_dir[i] == '\\' || base_dir[i] == '/' {
			break
		}
		base_dir = base_dir[:len(base_dir) - 1]
	}
	return base_dir
}

calc_decl_count :: proc(decl: ^Ast) -> int {
	count := 0
	switch decl.kind {
	case .BlockStmt:
		for stmt in decl.BlockStmt.stmts {
			count += calc_decl_count(stmt)
		}
	case .WhenStmt:
		inner_count := calc_decl_count(decl.WhenStmt.body)
		if decl.WhenStmt.else_stmt != nil {
			else_count := calc_decl_count(decl.WhenStmt.else_stmt)
			if else_count > inner_count {
				inner_count = else_count
			}
		}
		count += inner_count
	case .ValueDecl:
		count = len(decl.ValueDecl.names)
	case .ForeignBlockDecl:
		count = calc_decl_count(decl.ForeignBlockDecl.body)
	case .ImportDecl, .ForeignImportDecl:
		count = 1
	}
	return count
}

parse_build_project_directory_tag :: proc(token_for_pos: Token, s: string) -> bool {
	prefix :: "build-project-name"
	assert(strings.starts_with(s, prefix))

	s = strings.trim_space(s[len(prefix):])
	if len(s) == 0 {
		return true
	}

	any_correct := false
	for len(s) > 0 {
		this_kind_correct := true
		for {
			p := strings.trim_space(build_tag_get_token(s, &s))
			if len(p) == 0 {break}
			if p == "," {break}

			is_notted := false
			p_ref := p
			if p_ref[0] == '!' {
				is_notted = true
				p_ref = p_ref[1:]
				if len(p_ref) == 0 {
					syntax_error(token_for_pos, "Expected a build-project-name after '!'")
					break
				}
			}
			if len(p_ref) == 0 {
				continue
			}
			if is_notted {
				this_kind_correct = this_kind_correct && (p_ref != build_context.ODIN_BUILD_PROJECT_NAME)
			} else {
				this_kind_correct = this_kind_correct && (p_ref == build_context.ODIN_BUILD_PROJECT_NAME)
			}
		}
		any_correct = any_correct || this_kind_correct
	}
	return any_correct
}

parse_file_tag :: proc(lc: string, tok: Token, f: ^AstFile) -> bool {
	if strings.starts_with(lc, "build-project-name") {
		if !parse_build_project_directory_tag(tok, lc) {
			return false
		}
	} else if strings.starts_with(lc, "build") {
		if !parse_build_tag(tok, lc) {
			return false
		}
	} else if strings.starts_with(lc, "vet") {
		f.vet_flags = parse_vet_tag(tok, lc, ast_file_vet_flags(f))
		f.vet_flags_set = true
	} else if strings.starts_with(lc, "test") {
		if (build_context.command_kind & .Test) == {} {
			return false
		}
	} else if strings.starts_with(lc, "ignore") {
		return false
	} else if strings.starts_with(lc, "private") {
		f.flags |= {.IsPrivatePkg}
		command := strings.trim_left_space(lc[len("private"):])
		command = strings.trim_space(command)
		if lc == "private" {
			f.flags |= {.IsPrivatePkg}
		} else if command == "package" {
			f.flags |= {.IsPrivatePkg}
		} else if command == "file" {
			f.flags |= {.IsPrivateFile}
		}
	} else if strings.starts_with(lc, "feature") {
		f.feature_flags |= parse_feature_tag(tok, lc)
		f.feature_flags_set = true
	} else if lc == "lazy" {
		if build_context.ignore_lazy {
			// skip
		} else if f.pkg.kind == .Init && build_context.command_kind == .Doc {
			// skip
		} else {
			f.flags |= {.IsLazy}
		}
	} else if lc == "no-instrumentation" {
		f.flags |= {.NoInstrumentation}
	} else {
		syntax_error(tok, "Unknown tag '%.*s'", len(lc), lc)
	}
	return true
}

parse_setup_file_when_stmt :: proc(p: ^Parser, f: ^AstFile, base_dir: string, ws: ^AstWhenStmt) {
	if ws.body != nil {
		stmts := ws.body.BlockStmt.stmts
		parse_setup_file_decls(p, f, base_dir, stmts)
	}
	if ws.else_stmt != nil {
		switch ws.else_stmt.kind {
		case .BlockStmt:
			stmts := ws.else_stmt.BlockStmt.stmts
			parse_setup_file_decls(p, f, base_dir, stmts)
		case .WhenStmt:
			parse_setup_file_when_stmt(p, f, base_dir, &ws.else_stmt.WhenStmt)
		}
	}
}

parse_setup_file_decls :: proc(p: ^Parser, f: ^AstFile, base_dir: string, decls: []^Ast) {
	for &node, i in decls {
		if !is_ast_decl(node) &&
		   node.kind != .WhenStmt &&
		   node.kind != .BadStmt &&
		   node.kind != .EmptyStmt {
			if node.kind == .ExprStmt {
				expr := node.ExprStmt.expr
				if expr.kind == .CallExpr &&
				   expr.CallExpr.proc.kind == .BasicDirective {
					f.directive_count += 1
					continue
				}
			}
			syntax_error(node, "Only declarations are allowed at file scope, got %v", ast_strings[node.kind])
		} else if node.kind == .ImportDecl {
			id := &node.ImportDecl
			original_string := strings.trim_space(string_value_from_token(f, id.relpath))
			import_path: string
			ok := determine_path_from_string(&p.file_decl_mutex, node, base_dir, original_string, &import_path)
			if !ok {
				decls[i] = ast_bad_decl(f, id.relpath, id.relpath)
				continue
			}
			import_path = strings.trim_space(import_path)
			id.fullpath = import_path
			if is_package_name_reserved(import_path) {
				continue
			}
			try_add_import_path(p, import_path, original_string, ast_token(node).pos)
		} else if node.kind == .ForeignImportDecl {
			fl := &node.ForeignImportDecl
			if len(fl.filepaths) == 0 {
				syntax_error(decls[i], "No foreign paths found")
				decls[i] = ast_bad_decl(f, ast_token(fl.filepaths[0]), ast_end_token(fl.filepaths[len(fl.filepaths) - 1]))
			} else if !fl.multiple_filepaths && len(fl.filepaths) == 1 {
				fp := fl.filepaths[0]
				assert(fp.kind == .BasicLit)
				fp_token := fp.BasicLit.token
				file_str := strings.trim_space(string_value_from_token(f, fp_token))
				fullpath := file_str
				if !is_arch_wasm() || strings.has_suffix(fullpath, ".o") {
					foreign_path: string
					ok := determine_path_from_string(&p.file_decl_mutex, node, base_dir, file_str, &foreign_path)
					if !ok {
						decls[i] = ast_bad_decl(f, fp_token, fp_token)
						continue
					}
					fullpath = foreign_path
				}
				fl.fullpaths = make([]string, 1, permanent_allocator())
				fl.fullpaths[0] = fullpath
			}
		} else if node.kind == .WhenStmt {
			ws := &node.WhenStmt
			parse_setup_file_when_stmt(p, f, base_dir, ws)
		}
	}
}

parse_file :: proc(p: ^Parser, f: ^AstFile) -> bool {
	if len(f.tokens) == 0 {
		return true
	}
	if len(f.tokens) > 0 && f.tokens[0].kind == .EOF {
		return true
	}

	start_time := time.tick_now()
	filepath := f.tokenizer.fullpath
	base_dir := dir_from_path(filepath)

	tags := make([dynamic]Token, context.temp_allocator)
	first_invalid_token_set := false
	first_invalid_token: Token

	for f.curr_token.kind != .Package && f.curr_token.kind != .EOF {
		if f.curr_token.kind == .Comment {
			consume_comment_groups(f, f.prev_token)
		} else if f.curr_token.kind == .FileTag {
			append(&tags, f.curr_token)
			advance_token(f)
		} else {
			if !first_invalid_token_set {
				first_invalid_token_set = true
				first_invalid_token = f.curr_token
			}
			advance_token(f)
		}
	}

	docs := f.lead_comment
	if f.curr_token.kind != .Package {
		begin_error_block()
		defer end_error_block()
		t := first_invalid_token if first_invalid_token_set else f.curr_token
		syntax_error(t, "Expected a package declaration at the beginning of the file")
		if f.pkg != nil && f.pkg.name != "" {
			error_line("\tSuggestion: Add 'package %v' to the top of the file", f.pkg.name)
		}
		return false
	}

	if first_invalid_token_set {
		syntax_error(first_invalid_token, "Expected only comments or lines starting with '#+' before the package declaration")
		return false
	}

	f.package_token = expect_token(f, .Package)
	if f.package_token.kind != .Package {
		return false
	}

	package_name := expect_token_after(f, .Ident, "package")
	if package_name.kind == .Ident {
		if package_name.string == "_" {
			syntax_error(package_name, "Invalid package name '_'")
		} else if f.pkg.kind != .Runtime && package_name.string == "runtime" {
			syntax_error(package_name, "Use of reserved package name '%v'", package_name.string)
		} else if is_package_name_reserved(package_name.string) {
			syntax_error(package_name, "Use of reserved package name '%v'", package_name.string)
		}
	}
	f.package_name = package_name.string

	for tok in tags {
		assert(tok.kind == .FileTag)
		assert(strings.starts_with(tok.string, "#+"))
		lt := strings.trim_space(tok.string[2:])
		if parse_file_tag(lt, tok, f) == false {
			return false
		}
	}

	pd := ast_package_decl(f, f.package_token, package_name, docs, f.line_comment)
	expect_semicolon(f)
	f.pkg_decl = pd

	if f.error_count == 0 {
		decls := make([dynamic]^Ast, context.allocator)
		for f.curr_token.kind != .EOF {
			stmt := parse_stmt(f)
			if stmt != nil && stmt.kind != .EmptyStmt {
				append(&decls, stmt)
				if stmt.kind == .ExprStmt &&
				   stmt.ExprStmt.expr != nil &&
				   stmt.ExprStmt.expr.kind == .ProcLit {
					syntax_error(stmt, "Procedure literal evaluated but not used")
				}
				f.total_file_decl_count += calc_decl_count(stmt)
				if stmt.kind == .WhenStmt || stmt.kind == .ExprStmt || stmt.kind == .ImportDecl || stmt.kind == .ForeignBlockDecl {
					f.delayed_decl_count += 1
				}
			}
		}
		f.decls = decls[:]
		parse_setup_file_decls(p, f, base_dir, f.decls)
	}

	end_time := time.tick_now()
	f.time_to_parse = time.duration_seconds(time.tick_diff(start_time, end_time))

	for q in 0 ..< AstDelayQueue_COUNT {
		f.delayed_decls_queues[q] = make([dynamic]^Ast, context.allocator, 0, f.delayed_decl_count)
	}

	return f.error_count == 0
}

process_imported_file :: proc(p: ^Parser, imported_file: ImportedFile) -> ParseFileError {
	pkg := imported_file.pkg
	fi  := imported_file.fi
	pos := imported_file.pos

	file := new_clone(AstFile{})
	file.pkg = pkg
	file.id  = i32(imported_file.index + 1)

	err_pos: TokenPos
	err := init_ast_file(file, fi.fullpath, &err_pos)
	err_pos.file_id = file.id
	file.last_error = err

	if err != .None {
		if err == .EmptyFile {
			if fi.fullpath == p.init_fullpath {
				syntax_error(pos, "Initial file is empty - %v", p.init_fullpath)
				os.exit(1)
			}
		} else {
			switch err {
			case .WrongExtension:
				syntax_error(pos, "Failed to parse file: %v; invalid file extension: File must have the extension '.odin'", fi.name)
			case .InvalidFile:
				syntax_error(pos, "Failed to parse file: %v; invalid file or cannot be found", fi.name)
			case .Permission:
				syntax_error(pos, "Failed to parse file: %v; file permissions problem", fi.name)
			case .NotFound:
				syntax_error(pos, "Failed to parse file: %v; file cannot be found ('%v')", fi.name, fi.fullpath)
			case .InvalidToken:
				syntax_error(err_pos, "Failed to parse file: %v; invalid token found in file", fi.name)
			case .EmptyFile:
				syntax_error(pos, "Failed to parse file: %v; file contains no tokens", fi.name)
			case .FileTooLarge:
				syntax_error(pos, "Failed to parse file: %v; file is too large, exceeds maximum file size of 2 GiB", fi.name)
			}
			return err
		}
	}

	{
		name := file.fullpath
		name = remove_directory_from_path(name)
		name = remove_extension_from_path(name)
		if strings.starts_with(name, "_") {
			syntax_error(pos, "Files cannot start with '_', got '%v'", file.fullpath)
		}
	}

	if parse_file(p, file) {
		sync.mutex_lock(&pkg.files_mutex)
		append(&pkg.files, file)
		sync.mutex_unlock(&pkg.files_mutex)

		sync.mutex_lock(&pkg.name_mutex)
		if len(pkg.name) == 0 {
			pkg.name = file.package_name
		} else if pkg.name != file.package_name {
			if len(file.tokens) > 0 && file.tokens[0].kind != .EOF {
				tok := file.package_token
				tok.pos.file_id = file.id
				tok.pos.line    = max(tok.pos.line, 1)
				tok.pos.column  = max(tok.pos.column, 1)
				syntax_error(tok, "Different package name, expected '%v', got '%v'", pkg.name, file.package_name)
			}
		}
		sync.mutex_unlock(&pkg.name_mutex)

		atomic_add(&p.total_line_count, file.tokenizer.line_count)
		atomic_add(&p.total_token_count, len(file.tokens))
	}

	return .None
}

parse_packages :: proc(p: ^Parser, init_filename: string) -> ParseFileError {
	init_fullpath := path_to_full_path(permanent_allocator(), init_filename)

	if !path_is_directory(init_fullpath) {
		if !strings.has_suffix(init_fullpath, ".odin") {
			error({}, "Expected either a directory or a .odin file, got '%v'", init_filename)
			return .WrongExtension
		}
	} else if len(init_fullpath) != 0 {
		path_str := init_fullpath
		if path_str[len(path_str) - 1] == '/' {
			path_str = path_str[:len(path_str) - 1]
		}
		if (build_context.command_kind & Command__does_build) != {} &&
		   build_context.build_mode == .Executable {
			output_path := path_to_string(context.temp_allocator, build_context.build_paths[8])
			if path_is_directory(output_path) {
				error({}, "Please specify the executable name with -out:<string> as a directory exists with the same name in the current working directory")
				return .DirectoryAlreadyExists
			}
		}
	}

	init_pos: TokenPos
	{
		ok: bool
		s := get_fullpath_base_collection(permanent_allocator(), "runtime", &ok)
		if !ok {
			compiler_error("Unable to find The 'base:runtime' package. Is the ODIN_ROOT set up correctly?")
		}
		try_add_import_path(p, s, s, init_pos, .Runtime)
	}
	try_add_import_path(p, init_fullpath, init_fullpath, init_pos, .Init)
	p.init_fullpath = init_fullpath

	if build_context.command_kind == .Test {
		ok: bool
		s := get_fullpath_core_collection(permanent_allocator(), "testing", &ok)
		if !ok {
			compiler_error("Unable to find The 'core:testing' package. Is the ODIN_ROOT set up correctly?")
		}
		try_add_import_path(p, s, s, init_pos, .Normal)
	}

	for path_str in build_context.extra_packages {
		fullpath := path_to_full_path(permanent_allocator(), path_str)
		if !path_is_directory(fullpath) {
			if !strings.has_suffix(fullpath, ".odin") {
				error({}, "Expected either a directory or a .odin file, got '%v'", fullpath)
				return .WrongExtension
			}
		}
		pkg := try_add_import_path(p, fullpath, fullpath, init_pos, .Normal)
		if pkg != nil {
			pkg.is_extra = true
		}
	}

	thread_pool_wait()

	for node := p.file_error_head; node != nil; node = node.next {
		if node.err != .None {
			return node.err
		}
	}

	for i := len(p.packages) - 1; i >= 0; i -= 1 {
		pkg := p.packages[i]
		for j := len(pkg.files) - 1; j >= 0; j -= 1 {
			file := pkg.files[j]
			if file.error_count != 0 {
				if file.last_error != .None {
					return file.last_error
				}
				return .GeneralError
			}
		}
	}

	for pkg in p.packages {
		for file in pkg.files {
			p.total_seen_load_directive_count += file.seen_load_directive_count
		}
	}

	g_parsing_done = true
	return .None
}
