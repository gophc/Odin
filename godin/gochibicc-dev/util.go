package gochibicc

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"
)

//region MAX MIN

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MaxFloat32(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}

func MinFloat32(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}

//endregion

//region CharPtr

type CharPtr struct {
	ptr unsafe.Pointer
}

func (p *CharPtr) copy() *CharPtr {
	return &CharPtr{
		ptr: p.ptr,
	}
}

func (p *CharPtr) str() string {
	l, pp := 0, p.copy()
	for {
		if pp.xipp() == '\x00' {
			break
		}
		l += 1
	}
	bh := reflect.SliceHeader{
		Data: uintptr(p.ptr),
		Len:  l,
		Cap:  l,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return Bytes2String(buf)
}

func (p *CharPtr) xi() byte {
	return *((*uint8)(p.ptr))
}

func (p *CharPtr) xipp() byte {
	chr := p.xi()
	p.ptr = unsafe.Add(p.ptr, 1)
	return chr
}

func (p *CharPtr) xppi() byte {
	p.ptr = unsafe.Add(p.ptr, 1)
	chr := p.xi()
	return chr
}

func (p *CharPtr) xiss() byte {
	chr := p.xi()
	p.ptr = unsafe.Add(p.ptr, -1)
	return chr
}

func (p *CharPtr) xssi() byte {
	p.ptr = unsafe.Add(p.ptr, -1)
	chr := p.xi()
	return chr
}

func String2CharPtr(s string) *CharPtr {
	return &CharPtr{
		ptr: String2Pointer(s),
	}
}

func Bytes2CharPtr(b []byte) *CharPtr {
	return &CharPtr{
		ptr: Bytes2Pointer(b),
	}
}

//endregion

// String2Bytes return GoString's buffer slice(enable modify string)
func String2Bytes(s string) ([]byte, int) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh)), sh.Len
}

// Bytes2String convert b to string without copy
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Pointer returns &s[0], which is not allowed in go
func String2Pointer(s string) unsafe.Pointer {
	p := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Pointer(p.Data)
}

// Bytes2Pointer returns &b[0], which is not allowed in go
func Bytes2Pointer(b []byte) unsafe.Pointer {
	p := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return unsafe.Pointer(p.Data)
}

type FileProvider interface {
	file_exists(path string) bool
	open_file(path string) string
}

// Reports an error and exit.
func error_(ap ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, ap...)
	os.Exit(1)
}

// Reports an error message in the following format.
//
// foo.c:10: x = y + 1;
//               ^ <error message here>
func verror_at(filename string, input string, line_no int, loc int, ap ...interface{}) {
	input_, input_l := String2Bytes(input)
	// Find a line containing `loc`.
	line := loc
	for {
		if line > 0 && input_[line] != '\n' {
			line--
		}
		break
	}

	end := loc
	for {
		if end < input_l && input_[end] != '\n' {
			end++
		}
		break
	}

	// Print out the line.
	indent, _ := fmt.Fprintf(os.Stderr, "%s:%d: \n", filename, line_no)
	_, _ = fmt.Fprintf(os.Stderr, "%.*s\n", end-line, input_[line:end])

	// Show the error message.
	pos := display_width(input_, line, loc-line) + indent

	_, _ = fmt.Fprintf(os.Stderr, "%*s", pos, "") // print pos spaces.
	_, _ = fmt.Fprintln(os.Stderr, "^ ")
	_, _ = fmt.Fprintln(os.Stderr, ap...)
	_, _ = fmt.Fprintln(os.Stderr, "")
}

func error_tok(tok *Token, fmts string, ap ...interface{}) {
	verror_at(tok.file.name, tok.file.contents, tok.line_no, tok.loc, ap...)
	os.Exit(1)
}

func warn_tok(tok *Token, fmts string, ap ...interface{}) {
	verror_at(tok.file.name, tok.file.contents, tok.line_no, tok.loc, ap...)
}

// Consumes the current token if it matches `op`.
func equal(tok *Token, op string) bool {
	a := tok.getBytes()
	b, _ := String2Bytes(op)
	return bytes.Equal(a, b)
}

// Ensure that the current token is `op`.
func skip(tok *Token, op string) *Token {
	if !equal(tok, op) {
		error_tok(tok, "expected '%s'", op)
	}
	return tok.next
}

func consume(tok *Token, str string) (bool, *Token) {
	if equal(tok, str) {
		return true, tok.next
	}
	return false, tok
}

func startswith(p string, q string) bool {
	return strings.HasPrefix(p, q)
}

func startswith_(p []byte, q string) bool {
	q_, l := String2Bytes(q)
	return len(p) >= l && bytes.Equal(p[:l], q_)
}

func (tok *Token) getBytes() []byte {
	buf, _ := String2Bytes(tok.file.contents)
	return buf[tok.loc : tok.loc+tok.len]
}

func (tok *Token) getStr() string {
	return string(tok.getBytes())
}
