package odingo

import "core:fmt"
import "core:os"
import "core:strings"
import "core:sync"
import "core:slice"
import "core:unicode"
import "core:unicode/utf8"


TokenKind :: enum u8 {
    Token_Invalid,
    Token_EOF,
    Token_Comment,
    Token_FileTag,
    Token__LiteralBegin,
    Token_Ident,
    Token_Integer,
    Token_Float,
    Token_Imag,
    Token_Rune,
    Token_String,
    Token__LiteralEnd,
    Token__OperatorBegin,
    Token_Eq,
    Token_Not,
    Token_Hash,
    Token_At,
    Token_Dollar,
    Token_Pointer,
    Token_Question,
    Token_Add,
    Token_Sub,
    Token_Mul,
    Token_Quo,
    Token_Mod,
    Token_ModMod,
    Token_And,
    Token_Or,
    Token_Xor,
    Token_AndNot,
    Token_Shl,
    Token_Shr,
    Token_CmpAnd,
    Token_CmpOr,
    Token__AssignOpBegin,
    Token_AddEq,
    Token_SubEq,
    Token_MulEq,
    Token_QuoEq,
    Token_ModEq,
    Token_ModModEq,
    Token_AndEq,
    Token_OrEq,
    Token_XorEq,
    Token_AndNotEq,
    Token_ShlEq,
    Token_ShrEq,
    Token_CmpAndEq,
    Token_CmpOrEq,
    Token__AssignOpEnd,
    Token_Increment,
    Token_Decrement,
    Token_ArrowRight,
    Token_Uninit,
    Token__ComparisonBegin,
    Token_CmpEq,
    Token_NotEq,
    Token_Lt,
    Token_Gt,
    Token_LtEq,
    Token_GtEq,
    Token__ComparisonEnd,
    Token_OpenParen,
    Token_CloseParen,
    Token_OpenBracket,
    Token_CloseBracket,
    Token_OpenBrace,
    Token_CloseBrace,
    Token_Colon,
    Token_Semicolon,
    Token_Period,
    Token_Comma,
    Token_Ellipsis,
    Token_RangeFull,
    Token_RangeHalf,
    Token_BackSlash,
    Token__OperatorEnd,
    Token__KeywordBegin,
    Token_Import,
    Token_Foreign,
    Token_Package,
    Token_Typeid,
    Token_When,
    Token_Where,
    Token_If,
    Token_Else,
    Token_For,
    Token_Switch,
    Token_In,
    Token_Not_In,
    Token_Do,
    Token_Case,
    Token_Break,
    Token_Continue,
    Token_Fallthrough,
    Token_Defer,
    Token_Return,
    Token_Proc,
    Token_Struct,
    Token_Union,
    Token_Enum,
    Token_Bit_Set,
    Token_Bit_Field,
    Token_Map,
    Token_Dynamic,
    Token_Auto_Cast,
    Token_Cast,
    Token_Transmute,
    Token_Distinct,
    Token_Using,
    Token_Context,
    Token_Or_Else,
    Token_Or_Return,
    Token_Or_Break,
    Token_Or_Continue,
    Token_Asm,
    Token_Matrix,
    Token__KeywordEnd,
    Token_Count,
}

@(rodata)
token_strings := [Token_Count]string{
    .Token_Eq            = "=",
    .Token_Not           = "!",
    .Token_Hash          = "#",
    .Token_At            = "@",
    .Token_Dollar        = "$",
    .Token_Pointer       = "^",
    .Token_Question      = "?",
    .Token_Add           = "+",
    .Token_Sub           = "-",
    .Token_Mul           = "*",
    .Token_Quo           = "/",
    .Token_Mod           = "%",
    .Token_ModMod        = "%%",
    .Token_And           = "&",
    .Token_Or            = "|",
    .Token_Xor           = "~",
    .Token_AndNot        = "&~",
    .Token_Shl           = "<<",
    .Token_Shr           = ">>",
    .Token_CmpAnd        = "&&",
    .Token_CmpOr         = "||",
    .Token_AddEq         = "+=",
    .Token_SubEq         = "-=",
    .Token_MulEq         = "*=",
    .Token_QuoEq         = "/=",
    .Token_ModEq         = "%=",
    .Token_ModModEq      = "%%=",
    .Token_AndEq         = "&=",
    .Token_OrEq          = "|=",
    .Token_XorEq         = "~=",
    .Token_AndNotEq      = "&~=",
    .Token_ShlEq         = "<<=",
    .Token_ShrEq         = ">>=",
    .Token_CmpAndEq      = "&&=",
    .Token_CmpOrEq       = "||=",
    .Token_Increment     = "++",
    .Token_Decrement     = "--",
    .Token_ArrowRight    = "->",
    .Token_Uninit        = "---",
    .Token_CmpEq         = "==",
    .Token_NotEq         = "!=",
    .Token_Lt            = "<",
    .Token_Gt            = ">",
    .Token_LtEq          = "<=",
    .Token_GtEq          = ">=",
    .Token_OpenParen     = "(",
    .Token_CloseParen    = ")",
    .Token_OpenBracket   = "[",
    .Token_CloseBracket  = "]",
    .Token_OpenBrace     = "{",
    .Token_CloseBrace    = "}",
    .Token_Colon         = ":",
    .Token_Semicolon     = ";",
    .Token_Period        = ".",
    .Token_Comma         = ",",
    .Token_Ellipsis      = "..",
    .Token_RangeFull     = "..=",
    .Token_RangeHalf     = "..<",
    .Token_BackSlash     = "\\",
    .Token_Import        = "import",
    .Token_Foreign       = "foreign",
    .Token_Package       = "package",
    .Token_Typeid        = "typeid",
    .Token_When          = "when",
    .Token_Where         = "where",
    .Token_If            = "if",
    .Token_Else          = "else",
    .Token_For           = "for",
    .Token_Switch        = "switch",
    .Token_In            = "in",
    .Token_Not_In        = "not_in",
    .Token_Do            = "do",
    .Token_Case          = "case",
    .Token_Break         = "break",
    .Token_Continue      = "continue",
    .Token_Fallthrough   = "fallthrough",
    .Token_Defer         = "defer",
    .Token_Return        = "return",
    .Token_Proc          = "proc",
    .Token_Struct        = "struct",
    .Token_Union         = "union",
    .Token_Enum          = "enum",
    .Token_Bit_Set       = "bit_set",
    .Token_Bit_Field     = "bit_field",
    .Token_Map           = "map",
    .Token_Dynamic       = "dynamic",
    .Token_Auto_Cast     = "auto_cast",
    .Token_Cast          = "cast",
    .Token_Transmute     = "transmute",
    .Token_Distinct      = "distinct",
    .Token_Using         = "using",
    .Token_Context       = "context",
    .Token_Or_Else       = "or_else",
    .Token_Or_Return     = "or_return",
    .Token_Or_Break      = "or_break",
    .Token_Or_Continue   = "or_continue",
    .Token_Asm           = "asm",
    .Token_Matrix        = "matrix",
}

TokenPos :: struct {
    file_id: i32,
    offset:  i32,
    line:    i32,
    column:  i32,
}

token_pos_cmp :: proc(a, b: TokenPos) -> int {
    if a.file_id != b.file_id {
        return int(a.file_id) - int(b.file_id)
    }
    return int(a.offset) - int(b.offset)
}

token_pos_eq :: proc(a, b: TokenPos) -> bool {
    return a.file_id == b.file_id && a.offset == b.offset
}

token_pos_lt :: proc(a, b: TokenPos) -> bool {
    return token_pos_cmp(a, b) < 0
}

token_pos_gt :: proc(a, b: TokenPos) -> bool {
    return token_pos_cmp(a, b) > 0
}

token_pos_lte :: proc(a, b: TokenPos) -> bool {
    return token_pos_cmp(a, b) <= 0
}

token_pos_gte :: proc(a, b: TokenPos) -> bool {
    return token_pos_cmp(a, b) >= 0
}

token_pos_ne :: proc(a, b: TokenPos) -> bool {
    return !token_pos_eq(a, b)
}

TokenFlag :: enum u8 {
    Remove  = 1,
    Replace = 2,
}

TokenFlags :: bit_set[TokenFlag; u8]

Token :: struct {
    kind:   TokenKind,
    flags:  TokenFlags,
    string: string,
    pos:    TokenPos,
}

make_token_ident :: proc(s: string) -> Token {
    return Token{kind = .Token_Ident, string = s}
}

make_token_ident_cstring :: proc(s: cstring) -> Token {
    return Token{kind = .Token_Ident, string = string(s)}
}

token_is_newline :: proc(tok: Token) -> bool {
    return tok.kind == .Token_Semicolon && tok.string == "\n"
}

token_is_literal :: proc(kind: TokenKind) -> bool {
    return kind > .Token__LiteralBegin && kind < .Token__LiteralEnd
}

token_is_operator :: proc(kind: TokenKind) -> bool {
    return kind > .Token__OperatorBegin && kind < .Token__OperatorEnd
}

token_is_keyword :: proc(kind: TokenKind) -> bool {
    return kind > .Token__KeywordBegin && kind < .Token__KeywordEnd
}

token_is_comparison :: proc(kind: TokenKind) -> bool {
    return kind > .Token__ComparisonBegin && kind < .Token__ComparisonEnd
}

token_is_shift :: proc(kind: TokenKind) -> bool {
    return kind == .Token_Shl || kind == .Token_Shr
}

fnv32a :: proc(data: []u8) -> u32 {
    FNV_OFFSET :: 2166136261
    FNV_PRIME  :: 16777619
    hash := FNV_OFFSET
    for b in data {
        hash = (hash ~ u32(b)) * FNV_PRIME
    }
    return hash
}

KeywordHashEntry :: struct {
    hash: u32,
    kind: TokenKind,
    text: string,
}

KEYWORD_HASH_TABLE_COUNT :: 512
KEYWORD_HASH_TABLE_MASK  :: 511

keyword_hash_table: [KEYWORD_HASH_TABLE_COUNT]KeywordHashEntry
min_keyword_size: int = 2
max_keyword_size: int = 11
keyword_indices: [16]bool

add_keyword_hash_entry :: proc(s: string, kind: TokenKind) {
    max_keyword_size = max(max_keyword_size, len(s))
    keyword_indices[len(s)] = true
    hash := fnv32a(transmute([]u8)s)
    index := int(hash & KEYWORD_HASH_TABLE_MASK)
    for keyword_hash_table[index].hash != 0 {
        index = (index + 1) & KEYWORD_HASH_TABLE_MASK
    }
    keyword_hash_table[index].hash = hash
    keyword_hash_table[index].kind = kind
    keyword_hash_table[index].text = s
}

@(init)
init_keyword_hash_table :: proc() {
    for i := int(Token__KeywordBegin) + 1; i < int(Token__KeywordEnd); i += 1 {
        kind := TokenKind(i)
        add_keyword_hash_entry(token_strings[kind], kind)
    }
    // legacy alias: "notin" -> Token_Not_In
    add_keyword_hash_entry("notin", .Token_Not_In)
    assert(max_keyword_size < 16)
}

Tokenizer :: struct {
    curr_file_id:      i32,
    fullpath:          string,
    start:             [^]u8,
    end:               [^]u8,
    curr_rune:         rune,
    curr:              [^]u8,
    read_curr:         [^]u8,
    column_minus_one:  i32,
    line_count:        i32,
    error_count:       i32,
    insert_semicolon:  bool,
}

AstFile :: struct {}  // opaque placeholder - defined in the full compiler

// ============================================================================
// Error value types
// ============================================================================

ErrorValueKind :: enum u32 {
    ErrorValue_Error,
    ErrorValue_Warning,
}

ErrorValue :: struct {
    kind:         ErrorValueKind,
    pos:          TokenPos,
    end:          TokenPos,
    msg:          [dynamic]u8,
    seen_newline: bool,
}

// ============================================================================
// Terminal style and colour enums
// ============================================================================

TerminalStyle :: enum u8 {
    TerminalStyle_Normal,
    TerminalStyle_Bold,
    TerminalStyle_Underline,
}

TerminalColour :: enum u8 {
    TerminalColour_White,
    TerminalColour_Red,
    TerminalColour_Yellow,
    TerminalColour_Green,
    TerminalColour_Cyan,
    TerminalColour_Blue,
    TerminalColour_Purple,
    TerminalColour_Black,
    TerminalColour_Grey,
}

// ============================================================================
// Configuration flags (package-level, can be set by the compiler driver)
// ============================================================================

_global_warnings_as_errors: bool
_global_ignore_warnings: bool
_show_error_line: bool
_terse_errors: bool
_json_errors: bool
_has_ansi_terminal_colours: bool

MAX_ERROR_COLLECTOR_COUNT :: 36

// ============================================================================
// Error collector and global state
// ============================================================================

ErrorCollector :: struct {
    mutex:                sync.Mutex,
    path_mutex:           sync.Mutex,
    count:                i64,
    warning_count:        i64,
    in_block:             bool,
    error_values:         [dynamic]ErrorValue,
    curr_error_value:     ErrorValue,
    curr_error_value_set: bool,
}

global_error_collector: ErrorCollector

// File path and file data storage indexed by file_id
global_file_path_strings: [dynamic]string
global_files:             [dynamic]^AstFile
global_file_data:         [dynamic][]u8
global_files_mutex:       sync.Mutex

// Already-printed guard
errors_already_printed: bool

// ============================================================================
// Configuration accessors
// ============================================================================

global_warnings_as_errors :: proc() -> bool {
    return _global_warnings_as_errors
}

global_ignore_warnings :: proc() -> bool {
    return _global_ignore_warnings
}

show_error_line :: proc() -> bool {
    return _show_error_line
}

terse_errors :: proc() -> bool {
    return _terse_errors
}

json_errors :: proc() -> bool {
    return _json_errors
}

has_ansi_terminal_colours :: proc() -> bool {
    return _has_ansi_terminal_colours
}

set_global_warnings_as_errors :: proc(val: bool) { _global_warnings_as_errors = val }
set_global_ignore_warnings :: proc(val: bool)    { _global_ignore_warnings = val }
set_show_error_line :: proc(val: bool)           { _show_error_line = val }
set_terse_errors :: proc(val: bool)               { _terse_errors = val }
set_json_errors :: proc(val: bool)                { _json_errors = val }
set_has_ansi_terminal_colours :: proc(val: bool)  { _has_ansi_terminal_colours = val }

// ============================================================================
// Core error value push/pop/get
// ============================================================================

push_error_value :: proc(pos: TokenPos, kind := ErrorValueKind.ErrorValue_Error) {
    assert(global_error_collector.curr_error_value_set == false,
           "Possible race condition in error handling system, please report this with an issue")
    global_error_collector.curr_error_value.kind = kind
    global_error_collector.curr_error_value.pos = pos
    global_error_collector.curr_error_value.end = {}
    global_error_collector.curr_error_value.seen_newline = false
    clear(&global_error_collector.curr_error_value.msg)
    global_error_collector.curr_error_value_set = true
}

pop_error_value :: proc() {
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)
    if global_error_collector.curr_error_value_set {
        append(&global_error_collector.error_values, global_error_collector.curr_error_value)
        global_error_collector.curr_error_value = ErrorValue{}
        global_error_collector.curr_error_value_set = false
    }
}

try_pop_error_value :: proc() {
    if !global_error_collector.in_block {
        pop_error_value()
    }
}

get_error_value :: proc() -> ^ErrorValue {
    assert(global_error_collector.curr_error_value_set == true,
           "Possible race condition in error handling system, please report this with an issue")
    return &global_error_collector.curr_error_value
}

any_errors :: proc() -> bool {
    return global_error_collector.count != 0
}

any_warnings :: proc() -> bool {
    return global_error_collector.warning_count != 0
}

// ============================================================================
// Error block management
// ============================================================================

begin_error_block :: proc() {
    sync.lock(&global_error_collector.mutex)
    global_error_collector.in_block = true
}

end_error_block :: proc() {
    pop_error_value()
    global_error_collector.in_block = false
    sync.unlock(&global_error_collector.mutex)
}

// ============================================================================
// Initialization
// ============================================================================

init_global_error_collector :: proc() {
    global_error_collector.error_values = make([dynamic]ErrorValue)
    global_error_collector.curr_error_value = ErrorValue{}
    global_error_collector.curr_error_value_set = false
    global_error_collector.count = 0
    global_error_collector.warning_count = 0
    global_error_collector.in_block = false

    global_file_path_strings = make([dynamic]string)
    resize(&global_file_path_strings, 4096)
    clear(&global_file_path_strings)

    global_files = make([dynamic]^AstFile)
    resize(&global_files, 4096)
    clear(&global_files)

    global_file_data = make([dynamic][]u8)
    resize(&global_file_data, 4096)
    clear(&global_file_data)
}

@(init)
_auto_init_error_collector :: proc() {
    init_global_error_collector()
}

// ============================================================================
// File path management
// ============================================================================

set_file_path_string :: proc(index: i32, path: string) -> bool {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) >= len(global_file_path_strings) {
        resize(&global_file_path_strings, index + 1)
    }
    prev := global_file_path_strings[index]
    if len(prev) == 0 {
        global_file_path_strings[index] = path
        return true
    }
    return false
}

get_file_path_string :: proc(index: i32) -> string {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) < len(global_file_path_strings) {
        return global_file_path_strings[index]
    }
    return ""
}

thread_safe_set_ast_file_from_id :: proc(index: i32, file: ^AstFile) -> bool {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) >= len(global_files) {
        resize(&global_files, index + 1)
    }
    prev := global_files[index]
    if prev == nil {
        global_files[index] = file
        return true
    }
    return false
}

thread_safe_get_ast_file_from_id :: proc(index: i32) -> ^AstFile {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) < len(global_files) {
        return global_files[index]
    }
    return nil
}

thread_unsafe_get_ast_file_from_id :: proc(index: i32) -> ^AstFile {
    assert(index >= 0)
    if int(index) < len(global_files) {
        return global_files[index]
    }
    return nil
}

// ============================================================================
// File data management (for error line display)
// ============================================================================

set_file_data :: proc(index: i32, data: []u8) -> bool {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) >= len(global_file_data) {
        resize(&global_file_data, index + 1)
    }
    if global_file_data[index] == nil {
        global_file_data[index] = data
        return true
    }
    return false
}

get_file_data :: proc(index: i32) -> []u8 {
    assert(index >= 0)
    sync.lock(&global_files_mutex)
    defer sync.unlock(&global_files_mutex)

    if int(index) < len(global_file_data) {
        return global_file_data[index]
    }
    return nil
}

// ============================================================================
// Token position formatting
// ============================================================================

token_pos_to_string :: proc(pos: TokenPos) -> string {
    path := get_file_path_string(pos.file_id)
    return fmt.tprintf("%s(%d:%d)", path, pos.line, pos.column)
}

// ============================================================================
// Terminal colour output
// ============================================================================

terminal_set_colours :: proc(style: TerminalStyle, foreground: TerminalColour) {
    if !has_ansi_terminal_colours() {
        return
    }
    ss: string
    switch style {
    case .TerminalStyle_Normal:    ss = "0"
    case .TerminalStyle_Bold:      ss = "1"
    case .TerminalStyle_Underline: ss = "4"
    }
    switch foreground {
    case .TerminalColour_White:  error_out("\x1b[%s;37m", ss)
    case .TerminalColour_Red:    error_out("\x1b[%s;31m", ss)
    case .TerminalColour_Yellow: error_out("\x1b[%s;33m", ss)
    case .TerminalColour_Green:  error_out("\x1b[%s;32m", ss)
    case .TerminalColour_Cyan:   error_out("\x1b[%s;36m", ss)
    case .TerminalColour_Blue:   error_out("\x1b[%s;34m", ss)
    case .TerminalColour_Purple: error_out("\x1b[%s;35m", ss)
    case .TerminalColour_Black:  error_out("\x1b[%s;30m", ss)
    case .TerminalColour_Grey:   error_out("\x1b[%s;90m", ss)
    }
}

terminal_reset_colours :: proc() {
    if has_ansi_terminal_colours() {
        error_out("\x1b[0m")
    }
}

// ============================================================================
// Error output helpers - write to the current error value's message buffer
// ============================================================================

error_out_raw :: proc(s: string) {
    ev := get_error_value()
    if terse_errors() {
        for c in s {
            if c == '\n' {
                ev.seen_newline = true
                break
            }
            append(&ev.msg, u8(c))
        }
    } else {
        append(&ev.msg, ..transmute([]u8)s)
    }
}

error_out :: proc(format: string, args: ..any) {
    s := fmt.tprintf(format, ..args)
    error_out_raw(s)
}

error_out_empty :: proc() {
    error_out("")
}

error_out_pos :: proc(pos: TokenPos) {
    terminal_set_colours(.TerminalStyle_Bold, .TerminalColour_White)
    error_out("%s ", token_pos_to_string(pos))
    terminal_reset_colours()
}

error_out_coloured :: proc(str: string, style: TerminalStyle, foreground: TerminalColour) {
    terminal_set_colours(style, foreground)
    error_out("%s", str)
    terminal_reset_colours()
}

// ============================================================================
// Get a source line as a string from file data
// Returns the line text and sets the byte offset of pos within the returned
// string. Used by show_error_on_line.
// ============================================================================

get_file_line_as_string :: proc(pos: TokenPos, error_start_index_bytes: ^int) -> string {
    data := get_file_data(pos.file_id)
    if data == nil || len(data) == 0 {
        error_start_index_bytes^ = 0
        return ""
    }

    // Find the start of the target line
    // pos.offset is the error position in the file (byte offset)
    // We need to find the line containing this offset
    line_start: int
    current_line: i32 = 1

    for i := 0; i < len(data) && i <= int(pos.offset); i += 1 {
        if data[i] == '\n' {
            current_line += 1
            if current_line > pos.line {
                break
            }
            line_start = i + 1
        }
    }

    // If the line count doesn't match exactly, scan from beginning
    if current_line < pos.line {
        line_start = 0
        current_line = 1
        for i := 0; i < len(data); i += 1 {
            if current_line == pos.line {
                line_start = i
                break
            }
            if data[i] == '\n' {
                current_line += 1
            }
        }
    }

    // Find the end of the line
    line_end := len(data)
    for i := line_start; i < len(data); i += 1 {
        if data[i] == '\n' {
            line_end = i
            break
        }
    }

    error_start_index_bytes^ = int(pos.offset) - line_start
    if error_start_index_bytes^ < 0 {
        error_start_index_bytes^ = 0
    }
    if error_start_index_bytes^ > line_end - line_start {
        error_start_index_bytes^ = line_end - line_start
    }

    line_data := data[line_start:line_end]
    return string(line_data)
}

// ============================================================================
// Show error context on a source line
// Simplified version - uses character-based display instead of
// grapheme clusters. Odin does not have the ucg grapheme library.
// ============================================================================

show_error_on_line :: proc(pos: TokenPos, end: TokenPos) -> int {
    get_error_value().end = end
    if !show_error_line() {
        return -1
    }

    error_start_index_bytes: int
    the_line := get_file_line_as_string(pos, &error_start_index_bytes)

    if len(the_line) == 0 {
        terminal_set_colours(.TerminalStyle_Normal, .TerminalColour_Grey)
        error_out("\t( empty line )\n")
        terminal_reset_colours()
        return 0
    }

    MAX_LINE_LENGTH   :: 80
    MAX_TAB_WIDTH     :: 8
    ELLIPSIS_PADDING  :: 8
    MIN_LEFT_VIEW     :: 8
    MAX_INSERTED_WIDTH     := MAX_TAB_WIDTH + ELLIPSIS_PADDING
    MAX_LINE_LENGTH_PADDED := MAX_LINE_LENGTH - MAX_INSERTED_WIDTH

    // Compute display width (tabs count as MAX_TAB_WIDTH, other chars as 1)
    // and build a rune slice for positioning
    LineChar :: struct {
        byte_offset: int,
        width:       int,
    }
    line_chars: [dynamic]LineChar
    defer delete(line_chars)

    line_bytes := transmute([]u8)the_line
    i := 0
    for i < len(line_bytes) {
        lc := LineChar{byte_offset = i, width = 1}
        if line_bytes[i] == '\t' {
            lc.width = MAX_TAB_WIDTH
        }
        append(&line_chars, lc)
        // Advance by one UTF-8 character
        if line_bytes[i] < 0x80 {
            i += 1
        } else if line_bytes[i] < 0xE0 {
            i += 2
        } else if line_bytes[i] < 0xF0 {
            i += 3
        } else {
            i += 4
        }
    }

    // Find the grapheme index for the error position
    error_start_index_graphemes := 0
    for lc, idx in line_chars {
        if lc.byte_offset == error_start_index_bytes {
            error_start_index_graphemes = idx
            break
        }
    }
    if error_start_index_graphemes == 0 && error_start_index_bytes != 0 && len(line_chars) > 0 {
        error_start_index_graphemes = len(line_chars)
    }

    // Compute total line display width
    total_width := 0
    for lc in line_chars {
        total_width += lc.width
    }

    error_out("\t")
    show_right_ellipsis := false
    squiggle_padding := 0
    window_open_bytes := 0
    window_close_bytes := len(the_line)

    display_start := 0
    display_end := len(the_line)

    if total_width > MAX_LINE_LENGTH_PADDED {
        // Need to truncate the line display
        window_size_left := 0
        window_size_right := 0
        window_open_graphemes := 0

        // Find left window boundary
        for idx := error_start_index_graphemes - 1; idx > 0; idx -= 1 {
            window_size_left += line_chars[idx].width
            if window_size_left >= MIN_LEFT_VIEW {
                window_open_graphemes = idx
                window_open_bytes = line_chars[idx].byte_offset
                break
            }
        }

        // Find right window boundary
        for idx := error_start_index_graphemes; idx < len(line_chars); idx += 1 {
            window_size_right += line_chars[idx].width
            if window_size_right >= MAX_LINE_LENGTH_PADDED - MIN_LEFT_VIEW {
                window_close_bytes = line_chars[idx].byte_offset
                break
            }
        }
        if window_close_bytes == 0 {
            window_close_bytes = len(the_line)
        }

        // If window is too small on the right, expand left
        if window_size_right < MAX_LINE_LENGTH_PADDED - MIN_LEFT_VIEW {
            for idx := window_open_graphemes - 1; idx > 0; idx -= 1 {
                window_size_left += line_chars[idx].width
                if window_size_left + window_size_right >= MAX_LINE_LENGTH_PADDED {
                    window_open_graphemes = idx
                    window_open_bytes = line_chars[idx].byte_offset
                    break
                }
            }
        }

        assert(window_close_bytes >= window_open_bytes,
               "Error line truncation window has wrong byte indices")

        if window_close_bytes != len(the_line) {
            show_right_ellipsis = true
        }

        display_start = window_open_bytes
        display_end = window_close_bytes

        if window_open_bytes > 0 {
            error_out("... ")
            squiggle_padding += 4
        }
    }

    // Compute squiggle padding (visual offset before the error caret)
    for idx := error_start_index_graphemes - 1; idx >= 0; idx -= 1 {
        if line_chars[idx].byte_offset == display_start {
            break
        }
        squiggle_padding += line_chars[idx].width
    }

    // Print the line
    terminal_set_colours(.TerminalStyle_Normal, .TerminalColour_White)
    line_segment := the_line[display_start:display_end]
    error_out("%s", line_segment)

    // Compute squiggle length for the error indicator
    squiggle_length := 0
    trailing_squiggle := false

    if end.file_id == pos.file_id {
        if end.line > pos.line {
            // Error spans to next line - underline to end of display
            show_right_ellipsis = true
            for idx := error_start_index_graphemes; idx < len(line_chars); idx += 1 {
                squiggle_length += line_chars[idx].width
                trailing_squiggle = true
            }
        } else if end.line == pos.line && end.column > pos.column {
            // Error spans within the same line
            adjusted_end_index := line_chars[error_start_index_graphemes].byte_offset + int(end.column) - int(pos.column)
            for idx := error_start_index_graphemes; idx < len(line_chars); idx += 1 {
                if line_chars[idx].byte_offset >= adjusted_end_index {
                    break
                }
                if line_chars[idx].byte_offset >= display_end {
                    trailing_squiggle = true
                    break
                }
                squiggle_length += line_chars[idx].width
            }
        }
    } else {
        squiggle_length = 1
    }

    if show_right_ellipsis {
        error_out(" ...")
    }
    error_out("\n\t")

    // Print padding before the squiggle
    for _ in 0 ..< squiggle_padding {
        error_out(" ")
    }

    // Print the squiggle indicator
    terminal_set_colours(.TerminalStyle_Bold, .TerminalColour_Green)
    if squiggle_length > 0 {
        error_out("^")
        squiggle_length -= 1
    }
    for squiggle_length > 1 {
        error_out("~")
        squiggle_length -= 1
    }
    if squiggle_length > 0 {
        if trailing_squiggle {
            error_out("~ ...")
        } else {
            error_out("^")
        }
    }
    error_out("\n")
    terminal_reset_colours()

    return squiggle_padding
}

// ============================================================================
// Core error/warning output functions (the _va variants)
// These take already-formatted strings since Odin handles variadic
// formatting at the call site via fmt.tprintf.
// ============================================================================

error_va :: proc(pos: TokenPos, end: TokenPos, msg: string) {
    global_error_collector.count += 1
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    if global_error_collector.count > MAX_ERROR_COLLECTOR_COUNT {
        print_all_errors()
        os.exit(1)
    }

    push_error_value(pos, .ErrorValue_Error)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        error_out("%s", msg)
        error_out("\n")
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
            error_out_coloured("Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        }
        error_out("%s", msg)
        error_out("\n")
        show_error_on_line(pos, end)
    }

    try_pop_error_value()
}

warning_va :: proc(pos: TokenPos, end: TokenPos, msg: string) {
    if global_warnings_as_errors() {
        error_va(pos, end, msg)
        return
    }
    if global_ignore_warnings() {
        return
    }

    global_error_collector.warning_count += 1
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    push_error_value(pos, .ErrorValue_Warning)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Warning: ", .TerminalStyle_Normal, .TerminalColour_Yellow)
        error_out("%s", msg)
        error_out("\n")
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
            error_out_coloured("Warning: ", .TerminalStyle_Normal, .TerminalColour_Yellow)
        }
        error_out("%s", msg)
        error_out("\n")
        show_error_on_line(pos, end)
    }

    try_pop_error_value()
}

error_line_va :: proc(msg: string) {
    error_out("%s", msg)
}

error_no_newline_va :: proc(pos: TokenPos, msg: string) {
    global_error_collector.count += 1
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    if global_error_collector.count > MAX_ERROR_COLLECTOR_COUNT {
        print_all_errors()
        os.exit(1)
    }

    push_error_value(pos, .ErrorValue_Error)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        error_out("%s", msg)
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
        }
        if has_ansi_terminal_colours() {
            error_out_coloured("Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        }
        error_out("%s", msg)
    }

    try_pop_error_value()
}

// ============================================================================
// Syntax error functions
// ============================================================================

syntax_error_va :: proc(pos: TokenPos, end: TokenPos, msg: string) {
    global_error_collector.count += 1
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    if global_error_collector.count > MAX_ERROR_COLLECTOR_COUNT {
        print_all_errors()
        os.exit(1)
    }

    push_error_value(pos, .ErrorValue_Warning)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Syntax Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        error_out("%s", msg)
        error_out("\n")
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
        }
        error_out_coloured("Syntax Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        error_out("%s", msg)
        error_out("\n")
        show_error_on_line(pos, end)
    }

    try_pop_error_value()
}

syntax_error_with_verbose_va :: proc(pos: TokenPos, end: TokenPos, msg: string) {
    global_error_collector.count += 1
    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    if global_error_collector.count > MAX_ERROR_COLLECTOR_COUNT {
        print_all_errors()
        os.exit(1)
    }

    push_error_value(pos, .ErrorValue_Warning)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Syntax Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        error_out("%s", msg)
        error_out("\n")
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
        }
        if has_ansi_terminal_colours() {
            error_out_coloured("Syntax Error: ", .TerminalStyle_Normal, .TerminalColour_Red)
        }
        error_out("%s", msg)
        error_out("\n")
        show_error_on_line(pos, end)
    }

    try_pop_error_value()
}

syntax_warning_va :: proc(pos: TokenPos, end: TokenPos, msg: string) {
    if global_warnings_as_errors() {
        syntax_error_va(pos, end, msg)
        return
    }
    if global_ignore_warnings() {
        return
    }

    sync.lock(&global_error_collector.mutex)
    defer sync.unlock(&global_error_collector.mutex)

    global_error_collector.warning_count += 1
    push_error_value(pos, .ErrorValue_Warning)

    if pos.line == 0 {
        error_out_empty()
        error_out_coloured("Syntax Warning: ", .TerminalStyle_Normal, .TerminalColour_Yellow)
        error_out("%s", msg)
        error_out("\n")
    } else {
        if json_errors() {
            error_out_empty()
        } else {
            error_out_pos(pos)
        }
        error_out_coloured("Syntax Warning: ", .TerminalStyle_Normal, .TerminalColour_Yellow)
        error_out("%s", msg)
        error_out("\n")
    }

    try_pop_error_value()
}

// ============================================================================
// Wrapper functions - these format varargs and call the _va functions
// These are the primary API for reporting errors and warnings.
// ============================================================================

warning :: proc(tok: Token, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    warning_va(tok.pos, {}, msg)
}

error :: proc{error_token, error_pos}

error_token :: proc(tok: Token, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    error_va(tok.pos, {}, msg)
}

error_pos :: proc(pos: TokenPos, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    error_va(pos, {}, msg)
}

error_line :: proc(format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    error_line_va(msg)
}

syntax_error :: proc{syntax_error_token, syntax_error_pos}

syntax_error_token :: proc(tok: Token, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    syntax_error_va(tok.pos, {}, msg)
}

syntax_error_pos :: proc(pos: TokenPos, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    syntax_error_va(pos, {}, msg)
}

syntax_warning :: proc(tok: Token, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    syntax_warning_va(tok.pos, {}, msg)
}

syntax_error_with_verbose :: proc(pos: TokenPos, end: TokenPos, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    syntax_error_with_verbose_va(pos, end, msg)
}

// ============================================================================
// Compiler error and exit
// ============================================================================

compiler_error :: proc(format: string, args: ..any) {
    if any_errors() || any_warnings() {
        print_all_errors()
    }
    msg := fmt.tprintf(format, ..args)
    fmt.eprintf("Internal Compiler Error: %s\n", msg)
    os.exit(1)
}

exit_with_errors :: proc() {
    if any_errors() || any_warnings() {
        print_all_errors()
    }
    os.exit(1)
}

// ============================================================================
// Error value comparison for sorting/deduplication
// ============================================================================

error_value_cmp :: proc(a, b: ErrorValue) -> int {
    return token_pos_cmp(a.pos, b.pos)
}

// ============================================================================
// Error article table for grammar ("a" vs "an")
// ============================================================================

error_article_table := [][2]string{
    {"a ", "bit_set literal"},
    {"a ", "constant declaration"},
    {"a ", "dynamic array literal"},
    {"a ", "map index"},
    {"a ", "map literal"},
    {"a ", "matrix literal"},
    {"a ", "polymorphic type argument"},
    {"a ", "procedure argument"},
    {"a ", "simd vector literal"},
    {"a ", "slice literal"},
    {"a ", "structure literal"},
    {"a ", "variable declaration"},
    {"an ", "'any' literal"},
    {"an ", "array literal"},
    {"an ", "enumerated array literal"},
}

error_article :: proc(context_name: string) -> string {
    for entry in error_article_table {
        if context_name == entry[1] {
            return entry[0]
        }
    }
    return ""
}

// ============================================================================
// JSON escape helper for print_all_errors
// ============================================================================

_escape_json_char :: proc(buf: ^[dynamic]u8, c: u8) {
    switch c {
    case '\n': append(buf, ..transmute([]u8)"\\n")
    case '"':  append(buf, ..transmute([]u8)"\\\"")
    case '\\': append(buf, ..transmute([]u8)"\\\\")
    case '\b': append(buf, ..transmute([]u8)"\\b")
    case '\f': append(buf, ..transmute([]u8)"\\f")
    case '\r': append(buf, ..transmute([]u8)"\\r")
    case '\t': append(buf, ..transmute([]u8)"\\t")
    case:
        if c <= 0x1f {
            s := fmt.tprintf("\\u%04x", c)
            append(buf, ..transmute([]u8)s)
        } else {
            append(buf, c)
        }
    }
}

_escape_json_string :: proc(s: string) -> string {
    buf: [dynamic]u8
    for c in transmute([]u8)s {
        _escape_json_char(&buf, c)
    }
    return string(buf[:])
}

// ============================================================================
// Split error message into lines
// ============================================================================

_split_lines :: proc(data: []u8, allocator := context.allocator) -> [dynamic]string {
    lines: [dynamic]string
    line_start := 0
    for i := 0; i < len(data); i += 1 {
        if data[i] == '\n' {
            append(&lines, string(data[line_start:i]))
            line_start = i + 1
        }
    }
    if line_start < len(data) {
        append(&lines, string(data[line_start:]))
    }
    return lines
}

// ============================================================================
// Print all collected errors
// Handles both JSON and plain text output
// Deduplicates errors at the same position
// ============================================================================

print_all_errors :: proc() {
    if errors_already_printed {
        if global_error_collector.warning_count == i64(len(global_error_collector.error_values)) {
            for &ev in global_error_collector.error_values {
                delete(ev.msg)
            }
            clear(&global_error_collector.error_values)
            errors_already_printed = false
        }
        return
    }

    assert(any_errors() || any_warnings())

    // Sort error values by position
    slice.sort_by(global_error_collector.error_values[:], proc(a, b: ErrorValue) -> bool {
        return error_value_cmp(a, b) < 0
    })

    // Deduplicate errors at the same position
    {
        default_lines_to_skip := 1
        if show_error_line() {
            default_lines_to_skip += 2
        }

        prev_idx := -1
        i := 0
        for i < len(global_error_collector.error_values) {
            ev := &global_error_collector.error_values[i]
            if prev_idx >= 0 {
                prev_ev := &global_error_collector.error_values[prev_idx]
                if token_pos_eq(prev_ev.pos, ev.pos) {
                    // Same position - merge messages, skip duplicate lines
                    lines := _split_lines(ev.msg[:])
                    defer delete(lines)

                    skip_count := default_lines_to_skip
                    msg_start_line := 0
                    for line_idx in 0 ..< len(lines) {
                        if skip_count <= 0 {
                            msg_start_line = line_idx
                            break
                        }
                        if len(lines[line_idx]) == 0 {
                            break
                        }
                        skip_count -= 1
                    }

                    if msg_start_line < len(lines) {
                        addition_lines := lines[msg_start_line:]
                        for line in addition_lines {
                            if len(line) > 0 {
                                prev_line := string(prev_ev.msg[:])
                                if !strings.contains(prev_line, line) {
                                    append(&prev_ev.msg, '\n')
                                    append(&prev_ev.msg, ..transmute([]u8)line)
                                }
                            }
                        }
                    }

                    delete(ev.msg)
                    unordered_remove(&global_error_collector.error_values, i)
                } else {
                    prev_idx = i
                    i += 1
                }
            } else {
                prev_idx = i
                i += 1
            }
        }
    }

    // Build output
    res: [dynamic]u8
    defer delete(res)

    if json_errors() {
        append(&res, ..transmute([]u8)"{\n")
        cnt_str := fmt.tprintf("\t\"error_count\": %d,\n", len(global_error_collector.error_values))
        append(&res, ..transmute([]u8)cnt_str)
        append(&res, ..transmute([]u8)"\t\"errors\": [\n")

        for ev, ev_idx in global_error_collector.error_values {
            append(&res, ..transmute([]u8)"\t\t{\n")
            append(&res, ..transmute([]u8)"\t\t\t\"type\": \"")
            if ev.kind == .ErrorValue_Warning {
                append(&res, ..transmute([]u8)"warning")
            } else {
                append(&res, ..transmute([]u8)"error")
            }
            append(&res, ..transmute([]u8)"\",\n")

            if ev.pos.file_id != 0 {
                append(&res, ..transmute([]u8)"\t\t\t\"pos\": {\n")
                append(&res, ..transmute([]u8)"\t\t\t\t\"file\": \"")
                file := get_file_path_string(ev.pos.file_id)
                escaped := _escape_json_string(file)
                append(&res, ..transmute([]u8)escaped)
                append(&res, ..transmute([]u8)"\",\n")

                offset_str := fmt.tprintf("\t\t\t\t\"offset\": %d,\n", ev.pos.offset)
                line_str   := fmt.tprintf("\t\t\t\t\"line\": %d,\n", ev.pos.line)
                col_str    := fmt.tprintf("\t\t\t\t\"column\": %d,\n", ev.pos.column)
                end_col := max(ev.end.column, ev.pos.column)
                end_col_str := fmt.tprintf("\t\t\t\t\"end_column\": %d\n", end_col)

                append(&res, ..transmute([]u8)offset_str)
                append(&res, ..transmute([]u8)line_str)
                append(&res, ..transmute([]u8)col_str)
                append(&res, ..transmute([]u8)end_col_str)
                append(&res, ..transmute([]u8)"\t\t\t},\n")
            } else {
                append(&res, ..transmute([]u8)"\t\t\t\"pos\": null,\n")
            }

            append(&res, ..transmute([]u8)"\t\t\t\"msgs\": [\n")
            lines := _split_lines(ev.msg[:])
            if len(lines) > 0 {
                for line, j in lines {
                    append(&res, ..transmute([]u8)"\t\t\t\t\"")
                    escaped_line := _escape_json_string(line)
                    append(&res, ..transmute([]u8)escaped_line)
                    if j + 1 < len(lines) {
                        append(&res, ..transmute([]u8)"\",\n")
                    } else {
                        append(&res, ..transmute([]u8)"\"\n")
                    }
                }
            }
            delete(lines)
            append(&res, ..transmute([]u8)"\t\t\t]\n")
            append(&res, ..transmute([]u8)"\t\t}")
            if ev_idx + 1 != len(global_error_collector.error_values) {
                append(&res, ..transmute([]u8)",")
            }
            append(&res, ..transmute([]u8)"\n")
        }

        append(&res, ..transmute([]u8)"\t]\n")
        append(&res, ..transmute([]u8)"}\n")
    } else {
        // Plain text output
        for ev in global_error_collector.error_values {
            lines := _split_lines(ev.msg[:])
            for line, line_idx in lines {
                trimmed := strings.trim_right(line, " \t\r")
                append(&res, ..transmute([]u8)trimmed)
                append(&res, ..transmute([]u8)" \n")
                if line_idx == 0 && terse_errors() {
                    break
                }
            }
            delete(lines)
        }
    }

    // Write to stderr
    os.write(os.stderr, res[:])
    errors_already_printed = true
}

// ============================================================================
// Tokenizer error reporting helpers
// These are used by the tokenizer (part 3) to report errors during
// tokenization. They reference the Tokenizer struct from part 1.
// ============================================================================

// These procs will be defined in part 3 since they depend on the full
// Tokenizer struct definition.
tokenizer_err :: proc{tok_err_pos, tok_err_curr}

tok_err_pos :: proc(t: ^Tokenizer, pos: TokenPos, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    syntax_error_va(pos, {}, msg)
    t.error_count += 1
}

tok_err_curr :: proc(t: ^Tokenizer, format: string, args: ..any) {
    msg := fmt.tprintf(format, ..args)
    column := t.column_minus_one + 1
    if column < 1 {
        column = 1
    }
    pos := TokenPos{
        file_id = t.curr_file_id,
        line    = t.line_count,
        column  = column,
        offset  = i32(int(t.read_curr - t.start)),
    }
    syntax_error_va(pos, {}, msg)
    t.error_count += 1
}

TokenizerInitError :: enum u32 {
    TokenizerInit_None,
    TokenizerInit_Invalid,
    TokenizerInit_NotExists,
    TokenizerInit_Permission,
    TokenizerInit_Empty,
    TokenizerInit_FileTooLarge,
    TokenizerInit_Count,
}

// ============================================================================
// LoadedFile placeholder types (from already-rewritten Odin code)
// ============================================================================

LoadedFile :: struct {
    data: rawptr,
    size: u64,
}

LoadedFileError :: enum u32 {
    None,
    Empty,
    FileTooLarge,
    Invalid,
    NotExists,
    Permission,
}

// Forward declaration:
load_file_32 :: proc(path: cstring, file: ^LoadedFile, copy_contents: bool) -> LoadedFileError {
    // Stub - actual implementation is in another file
    _ = path
    _ = file
    _ = copy_contents
    return .NotExists
}

// ============================================================================
// Tokenizer error reporting (uses error functions from part 2)
// ============================================================================

// tokenizer_err is defined in part 2 as a proc group

// ============================================================================
// Core functions
// ============================================================================

advance_to_next_rune :: proc(t: ^Tokenizer) {
    if t.curr_rune == '\n' {
        t.column_minus_one = -1
        t.line_count += 1
    }
    if t.read_curr < t.end {
        t.curr = t.read_curr
        r := t.read_curr^
        if r == 0 {
            tokenizer_err(t, "Illegal character NUL")
            t.read_curr = t.read_curr + 1
        } else if r >= 0x80 {
            remaining := int(t.end - t.read_curr)
            width: int
            rune_val: rune
                        rune_val = rune(t.read_curr[0])
            width = 1
            if rune_val >= 0x80 {
                if rune_val < 0xE0 {
                    rune_val = (rune(t.read_curr[0]) & 0x1F) << 6 | (rune(t.read_curr[1]) & 0x3F)
                    width = 2
                } else if rune_val < 0xF0 {
                    rune_val = (rune(t.read_curr[0]) & 0x0F) << 12 | (rune(t.read_curr[1]) & 0x3F) << 6 | (rune(t.read_curr[2]) & 0x3F)
                    width = 3
                } else {
                    rune_val = (rune(t.read_curr[0]) & 0x07) << 18 | (rune(t.read_curr[1]) & 0x3F) << 12 | (rune(t.read_curr[2]) & 0x3F) << 6 | (rune(t.read_curr[3]) & 0x3F)
                    width = 4
                }
            }
            t.read_curr = t.read_curr + width
            if rune_val == rune(0xfffd) && width == 1 {
                tokenizer_err(t, "Illegal UTF-8 encoding")
            } else if rune_val == rune(0xfeff) && int(t.curr - t.start) > 0 {
                tokenizer_err(t, "Illegal byte order mark")
            }
            r = rune_val
        } else {
            t.read_curr = t.read_curr + 1
        }
        t.curr_rune = r
        t.column_minus_one += 1
    } else {
        t.curr = t.end
        t.curr_rune = rune(-1)
    }
}

init_tokenizer_with_data :: proc(t: ^Tokenizer, fullpath: string, data: rawptr, size: int) {
    t.fullpath = fullpath
    t.column_minus_one = -1
    t.line_count = 1
    t.start = ([^]u8)(data)
    t.read_curr = t.start
    t.curr = t.start
    t.end = t.start + size
    advance_to_next_rune(t)
    if t.curr_rune == rune(0xfeff) {
        advance_to_next_rune(t)
    }
}

loaded_file_error_map_to_tokenizer := [6]TokenizerInitError{
    .TokenizerInit_None,
    .TokenizerInit_Empty,
    .TokenizerInit_FileTooLarge,
    .TokenizerInit_Invalid,
    .TokenizerInit_NotExists,
    .TokenizerInit_Permission,
}

init_tokenizer_from_fullpath :: proc(t: ^Tokenizer, fullpath: string, copy_file_contents: bool) -> TokenizerInitError {
    cpath := strings.clone_to_cstring(fullpath, context.temp_allocator)
    file_err := load_file_32(cpath, &LoadedFile{}, copy_file_contents)
    err := loaded_file_error_map_to_tokenizer[uint(file_err)]
    // Stub - actual file loading handled by load_file_32
    return .TokenizerInit_None
}

// ============================================================================
// Character classification helpers
// ============================================================================

digit_value :: proc(r: rune) -> i32 {
    switch r {
    case '0'..'9': return i32(r) - i32('0')
    case 'a'..'f': return i32(r) - i32('a') + 10
    case 'A'..'F': return i32(r) - i32('A') + 10
    }
    return 16
}

peek_byte :: proc(t: ^Tokenizer, #optional offset := 0) -> u8 {
    ptr := t.read_curr + offset
    if ptr < t.end {
        return ptr^
    }
    return 0
}

// ============================================================================
// scan_mantissa - scan numeric digits for a given base
// ============================================================================

scan_mantissa :: proc(t: ^Tokenizer, base: i32, force_base: bool) {
    effective_base := base
    if !force_base {
        effective_base = 16
    }
    for digit_value(t.curr_rune) < effective_base || t.curr_rune == '_' {
        advance_to_next_rune(t)
    }
}

// ============================================================================
// scan_number_to_token - the big number scanning function
// ============================================================================

scan_number_to_token :: proc(t: ^Tokenizer, token: ^Token, seen_decimal_point: bool) {
    token.kind = .Token_Integer
    token.string = string(t.curr[0:1])
    token.pos.file_id = t.curr_file_id
    token.pos.line = t.line_count
    token.pos.column = t.column_minus_one + 1

    if seen_decimal_point {
        // Back up to include the '.' before the digits
        prev_ptr := t.curr - 1
        token.string = prev_ptr[0:int(t.curr - prev_ptr)]
        token.pos.column -= 1
        token.kind = .Token_Float
        scan_mantissa(t, 10, true)
        // goto exponent
        _goto_exponent(t, token)
        return
    }

    if t.curr_rune == '0' {
        prev_ptr := t.curr
        advance_to_next_rune(t)
        switch t.curr_rune {
        case 'b':
            advance_to_next_rune(t)
            scan_mantissa(t, 2, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid binary integer")
                token.kind = .Token_Invalid
            }
            _goto_end(t, token)
            return
        case 'o':
            advance_to_next_rune(t)
            scan_mantissa(t, 8, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid octal integer")
                token.kind = .Token_Invalid
            }
            _goto_end(t, token)
            return
        case 'd':
            advance_to_next_rune(t)
            scan_mantissa(t, 10, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid explicitly decimal integer")
                token.kind = .Token_Invalid
            }
            _goto_end(t, token)
            return
        case 'z':
            advance_to_next_rune(t)
            scan_mantissa(t, 12, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid dozenal integer")
                token.kind = .Token_Invalid
            }
            _goto_end(t, token)
            return
        case 'x':
            advance_to_next_rune(t)
            scan_mantissa(t, 16, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid hexadecimal integer")
                token.kind = .Token_Invalid
            }
            _goto_end(t, token)
            return
        case 'h':
            token.kind = .Token_Float
            advance_to_next_rune(t)
            scan_mantissa(t, 16, false)
            if int(t.curr - prev_ptr) <= 2 {
                tokenizer_err(t, "Invalid hexadecimal float")
                token.kind = .Token_Invalid
            } else {
                // Validate hex float digit count: must be 4, 8, or 16
                start_bytes := prev_ptr + 2
                n := int(t.curr - start_bytes)
                digit_count := 0
                for i in 0 ..< n {
                    if start_bytes[i] != '_' {
                        digit_count += 1
                    }
                }
                switch digit_count {
                case 4, 8, 16:
                    // valid
                case:
                    tokenizer_err(t, "Invalid hexadecimal float, expected 4, 8, or 16 digits, got %d", digit_count)
                }
            }
            _goto_end(t, token)
            return
        case:
            scan_mantissa(t, 10, true)
            _goto_fraction(t, token)
            return
        }
    }

    scan_mantissa(t, 10, true)
    _goto_fraction(t, token)
    return
}

// Internal goto-replacement helpers for scan_number_to_token

_goto_fraction :: proc(t: ^Tokenizer, token: ^Token) {
    if t.curr_rune == '.' {
        if peek_byte(t) == '.' {
            _goto_end(t, token)
            return
        }
        advance_to_next_rune(t)
        token.kind = .Token_Float
        scan_mantissa(t, 10, true)
    }
    _goto_exponent(t, token)
}

_goto_exponent :: proc(t: ^Tokenizer, token: ^Token) {
    if t.curr_rune == 'e' || t.curr_rune == 'E' {
        token.kind = .Token_Float
        advance_to_next_rune(t)
        if t.curr_rune == '-' || t.curr_rune == '+' {
            advance_to_next_rune(t)
        }
        scan_mantissa(t, 10, false)
    }
    switch t.curr_rune {
    case 'i', 'j', 'k':
        token.kind = .Token_Imag
        advance_to_next_rune(t)
    }
    _goto_end(t, token)
}

_goto_end :: proc(t: ^Tokenizer, token: ^Token) {
    // Update token string length
    // token.string was set as a slice into t.curr at the start
    // We need to recalculate the length
    _ = t
    _ = token
    // The caller sets token.string.len at the return site
}

// ============================================================================
// scan_escape - scan escape sequences in string/rune literals
// ============================================================================

scan_escape :: proc(t: ^Tokenizer) -> bool {
    r := t.curr_rune
    switch r {
    case 'a', 'b', 'e', 'f', 'n', 'r', 't', 'v', '\\', '\'', '\"':
        advance_to_next_rune(t)
        return true
    case '0'..'7':
        // Octal escape: up to 3 digits, max 255
        base: u32 = 8
        max_val: u32 = 255
        x: u32
        for _ in 0 ..< 3 {
            d := u32(digit_value(t.curr_rune))
            if d >= base {
                break
            }
            x = x * base + d
            advance_to_next_rune(t)
        }
        return true
    case 'x':
        // Hex escape: exactly 2 digits, max 255
        advance_to_next_rune(t)
        base: u32 = 16
        x: u32
        for _ in 0 ..< 2 {
            d := u32(digit_value(t.curr_rune))
            if d >= base {
                if t.curr_rune < 0 {
                    tokenizer_err(t, "Escape sequence was not terminated")
                } else {
                    tokenizer_err(t, "Illegal character %d in escape sequence", t.curr_rune)
                }
                return false
            }
            x = x * base + d
            advance_to_next_rune(t)
        }
        return true
    case 'u':
        // Unicode escape: exactly 4 hex digits
        advance_to_next_rune(t)
        base: u32 = 16
        x: u32
        for _ in 0 ..< 4 {
            d := u32(digit_value(t.curr_rune))
            if d >= base {
                if t.curr_rune < 0 {
                    tokenizer_err(t, "Escape sequence was not terminated")
                } else {
                    tokenizer_err(t, "Illegal character %d in escape sequence", t.curr_rune)
                }
                return false
            }
            x = x * base + d
            advance_to_next_rune(t)
        }
        return true
    case 'U':
        // Unicode escape: exactly 8 hex digits
        advance_to_next_rune(t)
        base: u32 = 16
        x: u32
        for _ in 0 ..< 8 {
            d := u32(digit_value(t.curr_rune))
            if d >= base {
                if t.curr_rune < 0 {
                    tokenizer_err(t, "Escape sequence was not terminated")
                } else {
                    tokenizer_err(t, "Illegal character %d in escape sequence", t.curr_rune)
                }
                return false
            }
            x = x * base + d
            advance_to_next_rune(t)
        }
        return true
    case:
        if t.curr_rune < 0 {
            tokenizer_err(t, "Escape sequence was not terminated")
        } else {
            tokenizer_err(t, "Unknown escape sequence")
        }
        return false
    }
}

// ============================================================================
// tokenizer_skip_line - skip to end of line
// ============================================================================

tokenizer_skip_line :: proc(t: ^Tokenizer) {
    for t.curr_rune != '\n' && t.curr_rune != rune(-1) {
        advance_to_next_rune(t)
    }
}

// ============================================================================
// tokenizer_skip_whitespace - skip whitespace, respecting semicolon context
// ============================================================================

tokenizer_skip_whitespace :: proc(t: ^Tokenizer, on_newline: bool) {
    if on_newline {
        for {
            switch t.curr_rune {
            case ' ', '\t', '\r':
                advance_to_next_rune(t)
                continue
            }
            break
        }
    } else {
        for {
            switch t.curr_rune {
            case '\n', ' ', '\t', '\r':
                advance_to_next_rune(t)
                continue
            }
            break
        }
    }
}

// ============================================================================
// tokenizer_get_token - the main tokenizer entry point
// ============================================================================

tokenizer_get_token :: proc(t: ^Tokenizer, token: ^Token, #optional repeat := 0) {
    tokenizer_skip_whitespace(t, t.insert_semicolon)
    token.kind = .Token_Invalid
    token.string = string(t.curr[0:1])
    token.pos.file_id = t.curr_file_id
    token.pos.line = t.line_count
    token.pos.offset = i32(int(t.curr - t.start))
    token.pos.column = t.column_minus_one + 1
    current_pos := token.pos
    curr_rune := t.curr_rune

    if unicode.is_alpha(curr_rune) {
        token.kind = .Token_Ident
        for unicode.is_alpha(t.curr_rune) || unicode.is_digit(t.curr_rune) {
            advance_to_next_rune(t)
        }
        ident_start :: t.start + int(token.pos.offset); token_len := int(t.curr - ident_start)
        if 1 < token_len && token_len <= max_keyword_size && keyword_indices[token_len] {
            token_text := token.string[:token_len]
            hash := fnv32a(transmute([]u8)token_text)
            index := int(hash & KEYWORD_HASH_TABLE_MASK)
            entry := &keyword_hash_table[index]
            if entry.kind != .Token_Invalid && entry.hash == hash {
                if entry.text == token_text {
                    token.kind = entry.kind
                    if token.kind == .Token_Not_In && len(entry.text) == 5 {
                        syntax_error(token^, "Did you mean 'not_in'?")
                    }
                }
            }
        }
        _goto_semicolon_check(t, token, current_pos)
        return
    } else {
        switch curr_rune {
        case '0'..'9':
            scan_number_to_token(t, token, false)
            _goto_semicolon_check(t, token, current_pos)
            return
        }
        advance_to_next_rune(t)
        switch curr_rune {
        case rune(-1):
            token.kind = .Token_EOF
            if t.insert_semicolon {
                t.insert_semicolon = false
                token.string = "\n"
                token.kind = .Token_Semicolon
                return
            }
            break
        case '\n':
            t.insert_semicolon = false
            token.string = "\n"
            token.kind = .Token_Semicolon
            return
        case '\\':
            t.insert_semicolon = false
            tokenizer_get_token(t, token)
            if token.pos.line == current_pos.line {
                new_pos := current_pos
                new_pos.column += 1
                new_pos.offset += 1
                tokenizer_err(t, new_pos, "Expected a newline after \\")
            }
            return
        case '\'':
            token.kind = .Token_Rune
            quote := curr_rune
            valid := true
            n := 0
            for {
                r := t.curr_rune
                if r == '\n' || r < 0 {
                    tokenizer_err(t, "Rune literal not terminated")
                    break
                }
                advance_to_next_rune(t)
                if r == quote {
                    break
                }
                n += 1
                if r == '\\' {
                    if !scan_escape(t) {
                        valid = false
                    }
                }
            }
            if valid && n != 1 {
                tokenizer_err(t, token.pos, "Invalid rune literal")
            }
            _goto_semicolon_check(t, token, current_pos)
            return
        case '`', '"':
            token.kind = .Token_String
            quote := curr_rune
            if curr_rune == '"' {
                for {
                    r := t.curr_rune
                    if r == '\n' || r < 0 {
                        tokenizer_err(t, "String literal not terminated")
                        break
                    }
                    advance_to_next_rune(t)
                    if r == quote {
                        break
                    }
                    if r == '\\' {
                        scan_escape(t)
                    }
                }
            } else {
                // Raw string: backtick-quoted
                for {
                    r := t.curr_rune
                    if r < 0 {
                        tokenizer_err(t, "String literal not terminated")
                        break
                    }
                    advance_to_next_rune(t)
                    if r == quote {
                        break
                    }
                }
            }
            _goto_semicolon_check(t, token, current_pos)
            return
        case '.':
            token.kind = .Token_Period
            switch t.curr_rune {
            case '.':
                advance_to_next_rune(t)
                token.kind = .Token_Ellipsis
                if t.curr_rune == '<' {
                    advance_to_next_rune(t)
                    token.kind = .Token_RangeHalf
                } else if t.curr_rune == '=' {
                    advance_to_next_rune(t)
                    token.kind = .Token_RangeFull
                }
                break
            case '0'..'9':
                scan_number_to_token(t, token, true)
                _goto_semicolon_check(t, token, current_pos)
                return
            }
            break
        case '@': token.kind = .Token_At;           break
        case '$': token.kind = .Token_Dollar;       break
        case '?': token.kind = .Token_Question;     break
        case '^': token.kind = .Token_Pointer;      break
        case ';': token.kind = .Token_Semicolon;    break
        case ',': token.kind = .Token_Comma;        break
        case ':': token.kind = .Token_Colon;        break
        case '(': token.kind = .Token_OpenParen;    break
        case ')': token.kind = .Token_CloseParen;   break
        case '[': token.kind = .Token_OpenBracket;  break
        case ']': token.kind = .Token_CloseBracket; break
        case '{': token.kind = .Token_OpenBrace;    break
        case '}': token.kind = .Token_CloseBrace;   break
        case '%':
            token.kind = .Token_Mod
            switch t.curr_rune {
            case '=':
                advance_to_next_rune(t)
                token.kind = .Token_ModEq
                break
            case '%':
                token.kind = .Token_ModMod
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_ModModEq
                    advance_to_next_rune(t)
                }
                break
            }
            break
        case '*':
            token.kind = .Token_Mul
            if t.curr_rune == '=' {
                advance_to_next_rune(t)
                token.kind = .Token_MulEq
            }
            break
        case '=':
            token.kind = .Token_Eq
            if t.curr_rune == '=' {
                advance_to_next_rune(t)
                token.kind = .Token_CmpEq
            }
            break
        case '~':
            token.kind = .Token_Xor
            if t.curr_rune == '=' {
                advance_to_next_rune(t)
                token.kind = .Token_XorEq
            }
            break
        case '!':
            token.kind = .Token_Not
            if t.curr_rune == '=' {
                advance_to_next_rune(t)
                token.kind = .Token_NotEq
            }
            break
        case '+':
            token.kind = .Token_Add
            switch t.curr_rune {
            case '=':
                advance_to_next_rune(t)
                token.kind = .Token_AddEq
                break
            case '+':
                advance_to_next_rune(t)
                token.kind = .Token_Increment
                break
            }
            break
        case '-':
            token.kind = .Token_Sub
            switch t.curr_rune {
            case '=':
                advance_to_next_rune(t)
                token.kind = .Token_SubEq
                break
            case '-':
                advance_to_next_rune(t)
                token.kind = .Token_Decrement
                if t.curr_rune == '-' {
                    advance_to_next_rune(t)
                    token.kind = .Token_Uninit
                }
                break
            case '>':
                advance_to_next_rune(t)
                token.kind = .Token_ArrowRight
                break
            }
            break
        case '#':
            token.kind = .Token_Hash
            if t.curr_rune == '!' {
                token.kind = .Token_Comment
                tokenizer_skip_line(t)
            } else if t.curr_rune == '+' {
                token.kind = .Token_FileTag
                for t.curr_rune != rune(-1) {
                    if t.curr_rune == '\n' {
                        break
                    }
                    if t.curr_rune == '/' {
                        break
                    }
                    advance_to_next_rune(t)
                }
            }
            break
        case '/':
            token.kind = .Token_Quo
            switch t.curr_rune {
            case '/':
                token.kind = .Token_Comment
                tokenizer_skip_line(t)
                break
            case '*':
                token.kind = .Token_Comment
                advance_to_next_rune(t)
                comment_scope := 1
                for comment_scope > 0 {
                    if t.curr_rune == rune(-1) {
                        tokenizer_err(t, "Multi-line comment not terminated")
                        break
                    } else if t.curr_rune == '/' {
                        advance_to_next_rune(t)
                        if t.curr_rune == '*' {
                            advance_to_next_rune(t)
                            comment_scope += 1
                        }
                    } else if t.curr_rune == '*' {
                        advance_to_next_rune(t)
                        if t.curr_rune == '/' {
                            advance_to_next_rune(t)
                            comment_scope -= 1
                        }
                    } else {
                        advance_to_next_rune(t)
                    }
                }
                break
            case '=':
                advance_to_next_rune(t)
                token.kind = .Token_QuoEq
                break
            }
            break
        case '<':
            token.kind = .Token_Lt
            switch t.curr_rune {
            case '=':
                token.kind = .Token_LtEq
                advance_to_next_rune(t)
                break
            case '<':
                token.kind = .Token_Shl
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_ShlEq
                    advance_to_next_rune(t)
                }
                break
            }
            break
        case '>':
            token.kind = .Token_Gt
            switch t.curr_rune {
            case '=':
                token.kind = .Token_GtEq
                advance_to_next_rune(t)
                break
            case '>':
                token.kind = .Token_Shr
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_ShrEq
                    advance_to_next_rune(t)
                }
                break
            }
            break
        case '&':
            token.kind = .Token_And
            switch t.curr_rune {
            case '~':
                token.kind = .Token_AndNot
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_AndNotEq
                    advance_to_next_rune(t)
                }
                break
            case '=':
                token.kind = .Token_AndEq
                advance_to_next_rune(t)
                break
            case '&':
                token.kind = .Token_CmpAnd
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_CmpAndEq
                    advance_to_next_rune(t)
                }
                break
            }
            break
        case '|':
            token.kind = .Token_Or
            switch t.curr_rune {
            case '=':
                token.kind = .Token_OrEq
                advance_to_next_rune(t)
                break
            case '|':
                token.kind = .Token_CmpOr
                advance_to_next_rune(t)
                if t.curr_rune == '=' {
                    token.kind = .Token_CmpOrEq
                    advance_to_next_rune(t)
                }
                break
            }
            break
        case:
            token.kind = .Token_Invalid
            if curr_rune != rune(0xfeff) {
                str: [4]u8
                len := utf8.encode_rune(str[:], curr_rune)
                tokenizer_err(t, "Illegal character: %.*s (%d) ", len, string(str[:len]), curr_rune)
            }
            break
        }
    }

    _goto_semicolon_check(t, token, current_pos)
    return
}

// ============================================================================
// Semicolon insertion logic
// ============================================================================

_goto_semicolon_check :: proc(t: ^Tokenizer, token: ^Token, current_pos: TokenPos) {
    // Update token string length based on current position
    start_ptr := t.start + int(token.pos.offset)
    token_len := max(int(t.curr - start_ptr), 1)
    token.string = string(start_ptr[:token_len])

    switch token.kind {
    case .Token_Invalid, .Token_Comment:
        // No semicolon insertion
    case .Token_Ident, .Token_Context, .Token_Typeid, .Token_Break,
         .Token_Continue, .Token_Fallthrough, .Token_Return,
         .Token_Or_Return, .Token_Or_Break, .Token_Or_Continue,
         .Token_Integer, .Token_Float, .Token_Imag,
         .Token_Rune, .Token_String, .Token_Uninit,
         .Token_Question, .Token_Pointer,
         .Token_CloseParen, .Token_CloseBracket, .Token_CloseBrace,
         .Token_Increment, .Token_Decrement, .Token_Not:
        t.insert_semicolon = true
    case:
        t.insert_semicolon = false
    }
}
