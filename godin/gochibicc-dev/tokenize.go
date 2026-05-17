package gochibicc

import "os"

func (app *ChibiccApp) error_at(loc int, ap ...interface{}) {
	input_, _ := String2Bytes(app.current_file.contents)

	line_no := 1
	for i := 0; i < loc; i++ {
		if input_[i] == '\n' {
			line_no++
		}
	}

	verror_at(app.current_file.name, app.current_file.contents, line_no, loc, ap...)
	os.Exit(1)
}

// Create a new token.
func (app *ChibiccApp) new_token(kind TokenKind, start int, end int) *Token {

	tok := &Token{}
	tok.kind = kind
	tok.loc = start
	tok.len = end - start
	tok.file = app.current_file
	tok.filename = app.current_file.display_name
	tok.at_bol = app.at_bol
	tok.has_space = app.has_space

	app.at_bol = false
	app.has_space = false
	return tok
}

// Read an identifier and returns the length of it.
// If p does not point to a valid identifier, 0 is returned.
func read_ident(buf []byte, start int) int {

	c, p, _ := decode_utf8(buf, start)
	if !is_ident1(c) {
		return 0
	}

	for {
		c, q, _ := decode_utf8(buf, p)
		if !is_ident2(c) {
			return p - start
		}

		p = q
	}

}

func from_hex(c byte) uint8 {
	if '0' <= c && c <= '9' {
		return c - '0'
	}

	if 'a' <= c && c <= 'f' {
		return c - 'a' + 10
	}

	return c - 'A' + 10
}

var kw_op = []string{
	"<<=", ">>=", "...", "==", "!=", "<=", ">=", "->", "+=",
	"-=", "*=", "/=", "++", "--", "%=", "&=", "|=", "^=", "&&",
	"||", "<<", ">>", "##",
}

// Read a punctuator token from p and returns its length.
func read_punct(buf []byte, p int) int {

	kwl := len(kw_op)
	for i := 0; i < kwl; i++ {
		if startswith_(buf[p:], kw_op[i]) {
			return len(kw_op[i])
		}
	}

	if ispunct(buf, p) {
		return 1
	}
	return 0
}

func ispunct(buf []byte, p int) bool {
	return false
}

var kw_word = []string{
	"return", "if", "else", "for", "while", "int", "sizeof", "char",
	"struct", "union", "short", "long", "void", "typedef", "_Bool",
	"enum", "static", "goto", "break", "continue", "switch", "case",
	"default", "extern", "_Alignof", "_Alignas", "do", "signed",
	"unsigned", "const", "volatile", "auto", "register", "restrict",
	"__restrict", "__restrict__", "_Noreturn", "float", "double",
	"typeof", "asm", "_Thread_local", "__thread", "_Atomic",
	"__attribute__",
}

func is_keyword(tok *Token) bool {
	kwl := len(kw_word)
	for i := 0; i < kwl; i++ {
		if equal(tok, kw_word[i]) {
			return true
		}
	}
	return false
}
