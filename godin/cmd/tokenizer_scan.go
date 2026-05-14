package cmd

import "unsafe"

func scan_number_to_token(t *Tokenizer, token *Token, seen_decimal_point bool) {
	token.Kind = TokenInteger
	token.Pos.FileID = t.CurrFileID
	token.Pos.Line = t.LineCount
	token.Pos.Column = t.ColumnMinusOne + 1

	if seen_decimal_point {
		token.Pos.Column--
		token.Kind = TokenFloat
		scan_mantissa(t, 10, true)
		goto exponent
	}

	if t.CurrRune == '0' {
		prev := t.Curr
		advance_to_next_rune(t)
		switch t.CurrRune {
		case 'b':
			advance_to_next_rune(t)
			scan_mantissa(t, 2, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid binary integer")
				token.Kind = TokenInvalid
			}
			goto end
		case 'o':
			advance_to_next_rune(t)
			scan_mantissa(t, 8, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid octal integer")
				token.Kind = TokenInvalid
			}
			goto end
		case 'd':
			advance_to_next_rune(t)
			scan_mantissa(t, 10, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid explicitly decimal integer")
				token.Kind = TokenInvalid
			}
			goto end
		case 'z':
			advance_to_next_rune(t)
			scan_mantissa(t, 12, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid dozenal integer")
				token.Kind = TokenInvalid
			}
			goto end
		case 'x':
			advance_to_next_rune(t)
			scan_mantissa(t, 16, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid hexadecimal integer")
				token.Kind = TokenInvalid
			}
			goto end
		case 'h':
			token.Kind = TokenFloat
			advance_to_next_rune(t)
			scan_mantissa(t, 16, false)
			if t.Curr-prev <= 2 {
				tokenizer_err(t, TokenPos{}, "Invalid hexadecimal float")
				token.Kind = TokenInvalid
			} else {
				start := prev + 2
				n := t.Curr - start
				digit_count := isize(0)
				for i := isize(0); i < n; i++ {
					if t.Data[start+i] != '_' {
						digit_count++
					}
				}
				switch digit_count {
				case 4, 8, 16:
				default:
					tokenizer_err(t, TokenPos{}, "Invalid hexadecimal float, expected 4, 8, or 16 digits, got %d", digit_count)
				}
			}
			goto end
		default:
			scan_mantissa(t, 10, true)
			goto fraction
		}
	}
	scan_mantissa(t, 10, true)

fraction:
	if t.CurrRune == '.' {
		if peek_byte(t, 0) == '.' {
			goto end
		}
		advance_to_next_rune(t)
		token.Kind = TokenFloat
		scan_mantissa(t, 10, true)
	}

exponent:
	if t.CurrRune == 'e' || t.CurrRune == 'E' {
		token.Kind = TokenFloat
		advance_to_next_rune(t)
		if t.CurrRune == '-' || t.CurrRune == '+' {
			advance_to_next_rune(t)
		}
		scan_mantissa(t, 10, false)
	}
	switch t.CurrRune {
	case 'i', 'j', 'k':
		token.Kind = TokenImag
		advance_to_next_rune(t)
	}

end:
	token.String = String{Data: &t.Data[token.Pos.Offset], Len: isize(t.Curr - int(token.Pos.Offset))}
}

func scan_escape(t *Tokenizer) bool {
	length := isize(0)
	base := uint32(0)
	max := uint32(0)
	x := uint32(0)
	r := t.CurrRune
	switch r {
	case 'a', 'b', 'e', 'f', 'n', 'r', 't', 'v', '\\', '\'', '"':
		advance_to_next_rune(t)
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		length = 3
		base = 8
		max = 255
	case 'x':
		advance_to_next_rune(t)
		length = 2
		base = 16
		max = 255
	case 'u':
		advance_to_next_rune(t)
		length = 4
		base = 16
		max = 0x0010ffff
	case 'U':
		advance_to_next_rune(t)
		length = 8
		base = 16
		max = 0x0010ffff
	default:
		if t.CurrRune < 0 {
			tokenizer_err(t, TokenPos{}, "Escape sequence was not terminated")
		} else {
			tokenizer_err(t, TokenPos{}, "Unknown escape sequence")
		}
		return false
	}
	for length > 0 {
		length--
		d := uint32(digit_value(t.CurrRune))
		if d >= base {
			if t.CurrRune < 0 {
				tokenizer_err(t, TokenPos{}, "Escape sequence was not terminated")
			} else {
				tokenizer_err(t, TokenPos{}, "Illegal character %d in escape sequence", t.CurrRune)
			}
			return false
		}
		x = x*base + d
		advance_to_next_rune(t)
	}
	return true
}

func tokenizer_skip_line(t *Tokenizer) {
	for t.CurrRune != '\n' && t.CurrRune != Rune(-1) {
		advance_to_next_rune(t)
	}
}

func tokenizer_skip_whitespace(t *Tokenizer, on_newline bool) {
	if on_newline {
		for {
			switch t.CurrRune {
			case ' ', '\t', '\r':
				advance_to_next_rune(t)
				continue
			}
			break
		}
	} else {
		for {
			switch t.CurrRune {
			case '\n', ' ', '\t', '\r':
				advance_to_next_rune(t)
				continue
			}
			break
		}
	}
}

func token_pos_add_column(pos TokenPos) TokenPos {
	pos.Column++
	pos.Offset++
	return pos
}

func tokenizer_get_token(t *Tokenizer, token *Token, repeat int) {
	tokenizer_skip_whitespace(t, t.InsertSemicolon)
	token.Kind = TokenInvalid
	token.Pos.FileID = t.CurrFileID
	token.Pos.Line = t.LineCount
	token.Pos.Offset = int32(t.Curr - t.Start)
	token.Pos.Column = t.ColumnMinusOne + 1
	currentPos := token.Pos
	currRune := t.CurrRune

	if rune_is_letter(currRune) {
		token.Kind = TokenIdent
		for rune_is_letter_or_digit(t.CurrRune) {
			advance_to_next_rune(t)
		}
		token.String = String{Data: &t.Data[token.Pos.Offset], Len: isize(t.Curr - int(token.Pos.Offset))}
		if 1 < token.String.Len && token.String.Len <= max_keyword_size && keyword_indices[token.String.Len] {
			hash := keyword_hash(token.String.Data, token.String.Len)
			index := hash & KEYWORD_HASH_TABLE_MASK
			entry := &keyword_hash_table[index]
			if entry.Kind != TokenInvalid && entry.Hash == hash {
				if str_eq(entry.Text, token.String) {
					token.Kind = entry.Kind
					if token.Kind == TokenNotIn && entry.Text.Len == 5 {
						syntax_error(*token, "Did you mean 'not_in'?")
					}
				}
			}
		}
		goto semicolon_check
	}

	switch currRune {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		scan_number_to_token(t, token, false)
		goto semicolon_check
	}

	advance_to_next_rune(t)

	switch currRune {
	case Rune(-1):
		token.Kind = TokenEOF
		if t.InsertSemicolon {
			t.InsertSemicolon = false
			token.String = String{Data: unsafe.StringData("\n"), Len: 1}
			token.Kind = TokenSemicolon
			return
		}

	case '\n':
		t.InsertSemicolon = false
		token.String = String{Data: unsafe.StringData("\n"), Len: 1}
		token.Kind = TokenSemicolon
		return

	case '\\':
		t.InsertSemicolon = false
		tokenizer_get_token(t, token, 0)
		if token.Pos.Line == currentPos.Line {
			tokenizer_err(t, token_pos_add_column(currentPos), "Expected a newline after \\")
		}
		return

	case '\'':
		token.Kind = TokenRune
		quote := currRune
		valid := true
		n := 0
		for {
			r := t.CurrRune
			if r == '\n' || r < 0 {
				tokenizer_err(t, TokenPos{}, "Rune literal not terminated")
				break
			}
			advance_to_next_rune(t)
			if r == quote {
				break
			}
			n++
			if r == '\\' {
				if !scan_escape(t) {
					valid = false
				}
			}
		}
		if valid && n != 1 {
			tokenizer_err(t, token.Pos, "Invalid rune literal")
		}
		token.String = String{Data: &t.Data[token.Pos.Offset], Len: isize(t.Curr - int(token.Pos.Offset))}
		goto semicolon_check

	case '`', '"':
		quote := currRune
		token.Kind = TokenString
		if currRune == '"' {
			for {
				r := t.CurrRune
				if r == '\n' || r < 0 {
					tokenizer_err(t, TokenPos{}, "String literal not terminated")
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
			for {
				r := t.CurrRune
				if r < 0 {
					tokenizer_err(t, TokenPos{}, "String literal not terminated")
					break
				}
				advance_to_next_rune(t)
				if r == quote {
					break
				}
			}
		}
		token.String = String{Data: &t.Data[token.Pos.Offset], Len: isize(t.Curr - int(token.Pos.Offset))}
		goto semicolon_check

	case '.':
		token.Kind = TokenPeriod
		switch t.CurrRune {
		case '.':
			advance_to_next_rune(t)
			token.Kind = TokenEllipsis
			if t.CurrRune == '<' {
				advance_to_next_rune(t)
				token.Kind = TokenRangeHalf
			} else if t.CurrRune == '=' {
				advance_to_next_rune(t)
				token.Kind = TokenRangeFull
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			scan_number_to_token(t, token, true)
		}

	case '@':
		token.Kind = TokenAt
	case '$':
		token.Kind = TokenDollar
	case '?':
		token.Kind = TokenQuestion
	case '^':
		token.Kind = TokenPointer
	case ';':
		token.Kind = TokenSemicolon
	case ',':
		token.Kind = TokenComma
	case ':':
		token.Kind = TokenColon
	case '(':
		token.Kind = TokenOpenParen
	case ')':
		token.Kind = TokenCloseParen
	case '[':
		token.Kind = TokenOpenBracket
	case ']':
		token.Kind = TokenCloseBracket
	case '{':
		token.Kind = TokenOpenBrace
	case '}':
		token.Kind = TokenCloseBrace

	case '%':
		token.Kind = TokenMod
		switch t.CurrRune {
		case '=':
			advance_to_next_rune(t)
			token.Kind = TokenModEq
		case '%':
			token.Kind = TokenModMod
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenModModEq
				advance_to_next_rune(t)
			}
		}

	case '*':
		token.Kind = TokenMul
		if t.CurrRune == '=' {
			advance_to_next_rune(t)
			token.Kind = TokenMulEq
		}

	case '=':
		token.Kind = TokenEq
		if t.CurrRune == '=' {
			advance_to_next_rune(t)
			token.Kind = TokenCmpEq
		}

	case '~':
		token.Kind = TokenXor
		if t.CurrRune == '=' {
			advance_to_next_rune(t)
			token.Kind = TokenXorEq
		}

	case '!':
		token.Kind = TokenNot
		if t.CurrRune == '=' {
			advance_to_next_rune(t)
			token.Kind = TokenNotEq
		}

	case '+':
		token.Kind = TokenAdd
		switch t.CurrRune {
		case '=':
			advance_to_next_rune(t)
			token.Kind = TokenAddEq
		case '+':
			advance_to_next_rune(t)
			token.Kind = TokenIncrement
		}

	case '-':
		token.Kind = TokenSub
		switch t.CurrRune {
		case '=':
			advance_to_next_rune(t)
			token.Kind = TokenSubEq
		case '-':
			advance_to_next_rune(t)
			token.Kind = TokenDecrement
			if t.CurrRune == '-' {
				advance_to_next_rune(t)
				token.Kind = TokenUninit
			}
		case '>':
			advance_to_next_rune(t)
			token.Kind = TokenArrowRight
		}

	case '#':
		token.Kind = TokenHash
		if t.CurrRune == '!' {
			token.Kind = TokenComment
			tokenizer_skip_line(t)
		} else if t.CurrRune == '+' {
			token.Kind = TokenFileTag
			for t.CurrRune != Rune(-1) {
				if t.CurrRune == '\n' {
					break
				}
				if t.CurrRune == '/' {
					break
				}
				advance_to_next_rune(t)
			}
		}

	case '/':
		token.Kind = TokenQuo
		switch t.CurrRune {
		case '/':
			token.Kind = TokenComment
			tokenizer_skip_line(t)
		case '*':
			token.Kind = TokenComment
			advance_to_next_rune(t)
			for comment_scope := isize(1); comment_scope > 0; {
				if t.CurrRune == Rune(-1) {
					tokenizer_err(t, TokenPos{}, "Multi-line comment not terminated")
					break
				} else if t.CurrRune == '/' {
					advance_to_next_rune(t)
					if t.CurrRune == '*' {
						advance_to_next_rune(t)
						comment_scope++
					}
				} else if t.CurrRune == '*' {
					advance_to_next_rune(t)
					if t.CurrRune == '/' {
						advance_to_next_rune(t)
						comment_scope--
					}
				} else {
					advance_to_next_rune(t)
				}
			}
		case '=':
			advance_to_next_rune(t)
			token.Kind = TokenQuoEq
		}

	case '<':
		token.Kind = TokenLt
		switch t.CurrRune {
		case '=':
			token.Kind = TokenLtEq
			advance_to_next_rune(t)
		case '<':
			token.Kind = TokenShl
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenShlEq
				advance_to_next_rune(t)
			}
		}

	case '>':
		token.Kind = TokenGt
		switch t.CurrRune {
		case '=':
			token.Kind = TokenGtEq
			advance_to_next_rune(t)
		case '>':
			token.Kind = TokenShr
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenShrEq
				advance_to_next_rune(t)
			}
		}

	case '&':
		token.Kind = TokenAnd
		switch t.CurrRune {
		case '~':
			token.Kind = TokenAndNot
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenAndNotEq
				advance_to_next_rune(t)
			}
		case '=':
			token.Kind = TokenAndEq
			advance_to_next_rune(t)
		case '&':
			token.Kind = TokenCmpAnd
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenCmpAndEq
				advance_to_next_rune(t)
			}
		}

	case '|':
		token.Kind = TokenOr
		switch t.CurrRune {
		case '=':
			token.Kind = TokenOrEq
			advance_to_next_rune(t)
		case '|':
			token.Kind = TokenCmpOr
			advance_to_next_rune(t)
			if t.CurrRune == '=' {
				token.Kind = TokenCmpOrEq
				advance_to_next_rune(t)
			}
		}

	default:
		token.Kind = TokenInvalid
		if currRune != Rune(0xfeff) {
			var str [4]byte
			length := gb_utf8_encode_rune(str[:], currRune)
			tokenizer_err(t, TokenPos{}, "Illegal character: %.*s (%d)", length, unsafe.String(&str[0], length), currRune)
		}
	}

	token.String = String{Data: &t.Data[token.Pos.Offset], Len: isize(t.Curr - int(token.Pos.Offset))}

semicolon_check:
	switch token.Kind {
	case TokenInvalid, TokenComment:
	case TokenIdent, TokenContext, TokenTypeid, TokenBreak, TokenContinue,
		TokenFallthrough, TokenReturn, TokenOrReturn, TokenOrBreak, TokenOrContinue,
		TokenInteger, TokenFloat, TokenImag, TokenRune, TokenString,
		TokenUninit, TokenQuestion, TokenPointer,
		TokenCloseParen, TokenCloseBracket, TokenCloseBrace,
		TokenIncrement, TokenDecrement, TokenNot:
		t.InsertSemicolon = true
	default:
		t.InsertSemicolon = false
	}
}
