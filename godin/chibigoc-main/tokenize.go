package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

// helpers

// base10
func strtol(s *[]rune) int64 {

	var val int64 = 0

	for i, r := range *s {
		if !unicode.IsDigit(r) {
			*s = (*s)[i:]
			return val
		}

		digit := int64(r) - 48
		val = 10*val + digit
	}

	*s = []rune("")
	return val
}

// Read a punctuator token from p and returns its length
//
// unicode.isPunct behaves differently then `ctype.h => int ispunct(char)`
//
// All punctuations in C:
// ! " # $ % & ' ( ) * + , - . / : ;
// < = > ? @ [ \ ] ^ _ ` { | } ~
func readPunct(p []rune) int {

	var s string
	// no need to stringify the []rune,
	// more than max. of possible punctuation length
	if len(p) > 2 {
		s = string(p[:2])
	} else {
		s = string(p[:len(p)])
	}

	if strings.HasPrefix(s, "==") ||
		strings.HasPrefix(s, "!=") ||
		strings.HasPrefix(s, "<=") ||
		strings.HasPrefix(s, ">=") {
		return 2
	}

	if unicode.IsPunct(p[0]) ||
		strings.HasPrefix(s, "+") ||
		strings.HasPrefix(s, "-") ||
		strings.HasPrefix(s, "<") ||
		strings.HasPrefix(s, ">") ||
		strings.HasPrefix(s, "=") ||
		strings.HasPrefix(s, ";") {
		return 1
	}
	return 0
}

// Returns true if c is valid as the first character of an identifier.
func isIdent1(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == '_'
}

// Returns true if c is valid as a non-first character of an identifier.
func isIdent2(r rune) bool {
	return isIdent1(r) || ('0' <= r && r <= '9')
}

//
// Tokenizer
//

// token data type
// it is implemented as a linked list

type TokenKind int

const (
	PUNCT   TokenKind = iota // punctuators
	IDENT                    // identifiers
	KEYWORD                  // keywords
	NUM                      // numeric literals
	EOF                      // end-of-file markers
)

func (tk TokenKind) String() string {
	switch tk {
	case PUNCT:
		return "punctuators"
	case IDENT:
		return "identifiers"
	case KEYWORD:
		return "keywords"
	case NUM:
		return "numeric literals"
	case EOF:
		return "end-of-file markers"
	default:
		return "token kind is not known"
	}
}

type Token struct {
	Kind TokenKind // token kind
	Next *Token    // next token
	val  int       // if kind is NUM, its value
	loc  []rune    // the rune slice, underlying the the token val
	// not needed, use len(loc)
	// len  int
}

// Consumes the current token, if it matches
func (t *Token) equal(s string) bool {
	return string(t.loc) == s
}

// consumes the token, even if it's not equal
func (t *Token) consume(rest **Token, str string) (isEqual bool) {
	if t.equal(str) {
		*rest = t.Next
		return true
	}
	*rest = t
	return false
}

func (t *Token) isKeyword() bool {
	kw := [...]string{"return", "if", "else", "for", "while", "int"}

	for _, v := range kw {
		if t.equal(v) {
			return true
		}
	}
	return false
}

// ensure token->Kind == NUM
func (t *Token) number() int {
	if t.Kind != NUM {
		errorTok(t, "expected a number")
	}
	return t.val
}

// functions related to token type

func NewToken(kind TokenKind, text []rune) *Token {
	t := new(Token)
	t.Kind = kind
	t.loc = text
	return t
}

// Ensures that the current token is `s`
func skip(t *Token, s string) *Token {
	if !t.equal(s) {
		errorTok(t, "expected '%s'", s)
	}
	return t.Next
}

// the return string's token type is specified here
func convertKeywords(tok *Token) {
	for t := tok; t.Kind != EOF; t = t.Next {
		if t.isKeyword() {
			t.Kind = KEYWORD
		}
	}
}

// Tokenize `p` and returns new tokens.
func tokenize() *Token {
	p := currentInput
	// start node of the linked list
	head := new(Token)
	cur := head

	for len(p) > 0 {

		// skip whitespace characters
		if unicode.IsSpace(p[0]) {
			p = p[1:]
			continue
		}

		// numeric literal
		if unicode.IsDigit(p[0]) {
			q := p
			num := strtol(&p)
			l := len(q) - len(p)

			cur.Next = NewToken(NUM, q[0:l])
			cur = cur.Next
			cur.val = int(num)

			continue
		}

		// identifier or keyword
		if isIdent1(p[0]) {
			fin := 1
			for isIdent2(p[fin]) {
				fin++
			}
			cur.Next = NewToken(IDENT, p[0:fin])
			cur = cur.Next
			p = p[fin:]
			continue
		}

		// punctuators
		if punctLen := readPunct(p); punctLen > 0 {
			cur.Next = NewToken(PUNCT, p[0:punctLen])
			cur = cur.Next
			p = p[punctLen:]
			continue
		}

		errorAt(p, "invalid token")
	}

	cur, cur.Next = cur.Next, NewToken(EOF, p)
	convertKeywords(head.Next)
	return head.Next
}

// walks and prints the linked list
func (t *Token) String() string {
	s := ""
	n := new(Token)
	*n = *t
	for ; n != nil; n = n.Next {
		s += fmt.Sprintln("[rune loc: ", (n.loc), "str: ", string(n.loc), " kind: ", n.Kind, " val:", n.val, "],")
	}
	return s
}

func errorAt(loc []rune, format string, v ...interface{}) {
	pos := len(currentInput) - cap(loc)
	fmt.Fprintln(os.Stderr, string(currentInput))
	fmt.Fprintf(os.Stderr, "%*s", pos, "") // print pos spaces
	fmt.Fprintln(os.Stderr, "^ ")
	fmt.Fprintln(os.Stderr, "token: ", string(loc))

	log.Fatalf(format, v...)
}

func errorTok(tok *Token, fmt string, v ...interface{}) {
	errorAt(tok.loc, fmt, v...)
}
