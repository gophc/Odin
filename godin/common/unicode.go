package common

import "unicode"

// Unicode category constants matching utf8proc categories.
const (
	UTF8PROC_CATEGORY_CN = 0
	UTF8PROC_CATEGORY_LU = 1
	UTF8PROC_CATEGORY_LL = 2
	UTF8PROC_CATEGORY_LT = 3
	UTF8PROC_CATEGORY_LM = 4
	UTF8PROC_CATEGORY_LO = 5
	UTF8PROC_CATEGORY_MN = 6
	UTF8PROC_CATEGORY_MC = 7
	UTF8PROC_CATEGORY_ME = 8
	UTF8PROC_CATEGORY_ND = 9
	UTF8PROC_CATEGORY_NL = 10
	UTF8PROC_CATEGORY_NO = 11
	UTF8PROC_CATEGORY_PC = 12
	UTF8PROC_CATEGORY_PD = 13
	UTF8PROC_CATEGORY_PS = 14
	UTF8PROC_CATEGORY_PE = 15
	UTF8PROC_CATEGORY_PI = 16
	UTF8PROC_CATEGORY_PF = 17
	UTF8PROC_CATEGORY_PO = 18
	UTF8PROC_CATEGORY_SM = 19
	UTF8PROC_CATEGORY_SC = 20
	UTF8PROC_CATEGORY_SK = 21
	UTF8PROC_CATEGORY_SO = 22
	UTF8PROC_CATEGORY_ZS = 23
	UTF8PROC_CATEGORY_ZL = 24
	UTF8PROC_CATEGORY_ZP = 25
	UTF8PROC_CATEGORY_CC = 26
	UTF8PROC_CATEGORY_CF = 27
	UTF8PROC_CATEGORY_CS = 28
	UTF8PROC_CATEGORY_CO = 29
)

// Boundary class constants.
const (
	UTF8PROC_BOUNDCLASS_START       = 0
	UTF8PROC_BOUNDCLASS_OTHER       = 1
	UTF8PROC_BOUNDCLASS_CR          = 2
	UTF8PROC_BOUNDCLASS_LF          = 3
	UTF8PROC_BOUNDCLASS_EXTEND      = 4
	UTF8PROC_BOUNDCLASS_EXTENDED    = 5
	UTF8PROC_BOUNDCLASS_PREFIX      = 6
	UTF8PROC_BOUNDCLASS_POSTFIX     = 7
	UTF8PROC_BOUNDCLASS_COMPLEX     = 8
	UTF8PROC_BOUNDCLASS_L           = 9
	UTF8PROC_BOUNDCLASS_V           = 10
	UTF8PROC_BOUNDCLASS_T           = 11
	UTF8PROC_BOUNDCLASS_LV          = 12
	UTF8PROC_BOUNDCLASS_LVT         = 13
	UTF8PROC_BOUNDCLASS_CR_LF       = 14
)

// Byte-level character classification.

func IsWhiteSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

func ByteToLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

func ByteToUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

func DigitToInt(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	return -1
}

func HexDigitToInt(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'f' {
		return int(c - 'a' + 10)
	}
	if c >= 'A' && c <= 'F' {
		return int(c - 'A' + 10)
	}
	return -1
}

// --- Rune functions ---

func RuneIsControl(r Rune) bool {
	return unicode.IsControl(rune(r))
}

func RuneIsWhitespace(r Rune) bool {
	return unicode.IsSpace(rune(r))
}

func RuneIsLetter(r Rune) bool {
	return unicode.IsLetter(rune(r))
}

func RuneIsDigit(r Rune) bool {
	return unicode.IsDigit(rune(r))
}

func RuneIsUpper(r Rune) bool {
	return unicode.IsUpper(rune(r))
}

func RuneIsLower(r Rune) bool {
	return unicode.IsLower(rune(r))
}

func RuneIsTitle(r Rune) bool {
	return unicode.IsTitle(rune(r))
}

func RuneIsPrint(r Rune) bool {
	return unicode.IsPrint(rune(r))
}

func RuneIsPunct(r Rune) bool {
	return unicode.IsPunct(rune(r))
}

func RuneIsSymbol(r Rune) bool {
	return unicode.IsSymbol(rune(r))
}

func RuneIsMark(r Rune) bool {
	return unicode.IsMark(rune(r))
}

func RuneIsNumber(r Rune) bool {
	return unicode.IsNumber(rune(r))
}

func IsPrintable(r Rune) bool {
	return unicode.IsPrint(rune(r))
}

func RuneToLower(r Rune) Rune {
	return Rune(unicode.ToLower(rune(r)))
}

func RuneToUpper(r Rune) Rune {
	return Rune(unicode.ToUpper(rune(r)))
}

func RuneToTitle(r Rune) Rune {
	return Rune(unicode.ToTitle(rune(r)))
}

func RuneLength(r Rune) int {
	c := rune(r)
	if c < 0x80 {
		return 1
	}
	if c < 0x800 {
		return 2
	}
	if c < 0x10000 {
		return 3
	}
	return 4
}

// --- UTF-8 decode/encode ---

// Utf8Decode decodes the first UTF-8 codepoint from s, returning the codepoint and bytes consumed.
func Utf8Decode(s []byte) (Rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	c := s[0]
	if c < 0x80 {
		return Rune(c), 1
	}
	if c < 0xC0 {
		return 0xFFFD, 1
	}
	if c < 0xE0 {
		if len(s) < 2 {
			return 0xFFFD, 1
		}
		return Rune(c&0x1F)<<6 | Rune(s[1]&0x3F), 2
	}
	if c < 0xF0 {
		if len(s) < 3 {
			return 0xFFFD, 1
		}
		return Rune(c&0x0F)<<12 | Rune(s[1]&0x3F)<<6 | Rune(s[2]&0x3F), 3
	}
	if len(s) < 4 {
		return 0xFFFD, 1
	}
	r := Rune(c&0x07)<<18 | Rune(s[1]&0x3F)<<12 | Rune(s[2]&0x3F)<<6 | Rune(s[3]&0x3F)
	if r > 0x10FFFF {
		return 0xFFFD, 4
	}
	return r, 4
}

// Utf8EncodeRune encodes a rune into p, returning bytes written or -1 on error.
func Utf8EncodeRune(p []byte, r Rune) int {
	if r < 0 || r > 0x10FFFF {
		return -1
	}
	if r <= 0x7F {
		if len(p) < 1 {
			return 0
		}
		p[0] = byte(r)
		return 1
	}
	if r <= 0x7FF {
		if len(p) < 2 {
			return 0
		}
		p[0] = 0xC0 | byte(r>>6)
		p[1] = 0x80 | byte(r&0x3F)
		return 2
	}
	if r <= 0xFFFF {
		if len(p) < 3 {
			return 0
		}
		p[0] = 0xE0 | byte(r>>12)
		p[1] = 0x80 | byte(r>>6&0x3F)
		p[2] = 0x80 | byte(r&0x3F)
		return 3
	}
	if len(p) < 4 {
		return 0
	}
	p[0] = 0xF0 | byte(r>>18)
	p[1] = 0x80 | byte(r>>12&0x3F)
	p[2] = 0x80 | byte(r>>6&0x3F)
	p[3] = 0x80 | byte(r&0x3F)
	return 4
}

// --- Surrogate pair decoding ---

func DecodeSurrogatePair(r1, r2 u16) Rune {
	const surr1 = 0xd800
	const surr2 = 0xdc00
	const surr3 = 0xe000
	const surrSelf = 0x10000
	if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
		return (Rune(r1-surr1)<<10 | Rune(r2-surr2)) + surrSelf
	}
	return 0xFFFD
}
