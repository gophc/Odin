package cmd

import (
	"fmt"
	"sync"
	"unsafe"
)

const (
	keywordHashTableCount = 1 << 9
	keywordHashTableMask  = keywordHashTableCount - 1
)

var (
	keywordHashTable [keywordHashTableCount]KeywordHashEntry
	minKeywordSize   isize = 2
	maxKeywordSize   isize = 11
	keywordIndices   [16]bool

	globalFilePathStrings []String
	globalFiles           []*AstFile
	globalFilesMutex      sync.Mutex
)

var loadedFileErrorMapToTokenizer = []TokenizerInitError{
	TokenizerInitNone,         // LoadedFile_None (0)
	TokenizerInitEmpty,        // LoadedFile_Empty (1)
	TokenizerInitFileTooLarge, // LoadedFile_FileTooLarge (2)
	TokenizerInitInvalid,      // LoadedFile_Invalid (3)
	TokenizerInitNotExists,    // LoadedFile_NotExists (4)
	TokenizerInitPermission,   // LoadedFile_Permission (5)
}

func keywordHash(text *byte, len isize) uint32 {
	return fnv32a(unsafe.Slice(text, len))
}

func addKeywordHashEntry(s String, kind TokenKind) {
	maxKeywordSize = max(maxKeywordSize, s.Len)
	keywordIndices[s.Len] = true
	hash := keywordHash(s.Data, s.Len)
	index := int(hash) & keywordHashTableMask
	entry := &keywordHashTable[index]
	if entry.Kind != TokenInvalid {
		gb_assert_handler("Assertion Failure",
			"Keyword hash table initialization collision",
			"cmd_tokenizer_core.go", int64(0),
			"entry->kind == Token_Invalid")
	}
	entry.Hash = hash
	entry.Kind = kind
	entry.Text = s
}

func initKeywordHashTable() {
	for kind := TokenKeywordBegin + 1; kind < TokenKeywordEnd; kind++ {
		addKeywordHashEntry(tokenStrings[kind], kind)
	}
	addKeywordHashEntry(String{Data: unsafe.StringData("notin"), Len: 5}, TokenNotIn)
	if maxKeywordSize >= 16 {
		gb_assert_handler("Assertion Failure", "",
			"cmd_tokenizer_core.go", int64(0), "max_keyword_size < 16")
	}
}

func tokenPosCmp(a, b TokenPos) int {
	if a.Offset != b.Offset {
		if a.Offset < b.Offset {
			return -1
		}
		return +1
	}
	if a.Line != b.Line {
		if a.Line < b.Line {
			return -1
		}
		return +1
	}
	if a.Column != b.Column {
		if a.Column < b.Column {
			return -1
		}
		return +1
	}
	pathA := getFilePathString(a.FileID)
	pathB := getFilePathString(b.FileID)
	sa := unsafe.String(pathA.Data, pathA.Len)
	sb := unsafe.String(pathB.Data, pathB.Len)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return +1
	}
	return 0
}

func tokenPosEq(a, b TokenPos) bool { return tokenPosCmp(a, b) == 0 }
func tokenPosNe(a, b TokenPos) bool { return tokenPosCmp(a, b) != 0 }
func tokenPosLt(a, b TokenPos) bool { return tokenPosCmp(a, b) < 0 }
func tokenPosLe(a, b TokenPos) bool { return tokenPosCmp(a, b) <= 0 }
func tokenPosGt(a, b TokenPos) bool { return tokenPosCmp(a, b) > 0 }
func tokenPosGe(a, b TokenPos) bool { return tokenPosCmp(a, b) >= 0 }

func tokenPosAddColumn(pos TokenPos) TokenPos {
	pos.Column += 1
	pos.Offset += 1
	return pos
}

func makeTokenIdent(s String) Token {
	return Token{Kind: TokenIdent, String: s}
}

func makeTokenIdentC(s string) Token {
	return Token{Kind: TokenIdent, String: makeStringC(s)}
}

func tokenIsNewline(tok Token) bool {
	return tok.Kind == TokenSemicolon && tok.String.Len == 1 && *tok.String.Data == '\n'
}

func tokenIsLiteral(t TokenKind) bool {
	return t > TokenLiteralBegin && t < TokenLiteralEnd
}

func tokenIsOperator(t TokenKind) bool {
	return t > TokenOperatorBegin && t < TokenOperatorEnd
}

func tokenIsKeyword(t TokenKind) bool {
	return t > TokenKeywordBegin && t < TokenKeywordEnd
}

func tokenIsComparison(t TokenKind) bool {
	return t > TokenComparisonBegin && t < TokenComparisonEnd
}

func tokenIsShift(t TokenKind) bool {
	return t == TokenShl || t == TokenShr
}

func printToken(t Token) {
	s := unsafe.String(t.String.Data, t.String.Len)
	fmt.Printf("%s\n", s)
}

func setFilePathString(index i32, path String) bool {
	if index < 0 {
		gb_assert_handler("Assertion Failure", "index >= 0",
			"cmd_tokenizer_core.go", int64(0), "")
	}
	mutex_lock(&globalFilesMutex)
	defer mutex_unlock(&globalFilesMutex)

	if int(index) >= len(globalFilePathStrings) {
		globalFilePathStrings = append(globalFilePathStrings,
			make([]String, int(index)+1-len(globalFilePathStrings))...)
	}
	prev := globalFilePathStrings[index]
	if prev.Len == 0 {
		globalFilePathStrings[index] = path
		return true
	}
	return false
}

func threadSafeSetAstFileFromId(index i32, file *AstFile) bool {
	if index < 0 {
		gb_assert_handler("Assertion Failure", "index >= 0",
			"cmd_tokenizer_core.go", int64(0), "")
	}
	mutex_lock(&globalFilesMutex)
	defer mutex_unlock(&globalFilesMutex)

	if int(index) >= len(globalFiles) {
		globalFiles = append(globalFiles,
			make([]*AstFile, int(index)+1-len(globalFiles))...)
	}
	prev := globalFiles[index]
	if prev == nil {
		globalFiles[index] = file
		return true
	}
	return false
}

func getFilePathString(index i32) String {
	if index < 0 {
		gb_assert_handler("Assertion Failure", "index >= 0",
			"cmd_tokenizer_core.go", int64(0), "")
	}
	mutex_lock(&globalFilesMutex)
	defer mutex_unlock(&globalFilesMutex)

	if int(index) < len(globalFilePathStrings) {
		return globalFilePathStrings[index]
	}
	return String{}
}

func threadSafeGetAstFileFromId(index i32) *AstFile {
	if index < 0 {
		gb_assert_handler("Assertion Failure", "index >= 0",
			"cmd_tokenizer_core.go", int64(0), "")
	}
	mutex_lock(&globalFilesMutex)
	defer mutex_unlock(&globalFilesMutex)

	if int(index) < len(globalFiles) {
		return globalFiles[index]
	}
	return nil
}

func threadUnsafeGetAstFileFromId(index i32) *AstFile {
	if index < 0 {
		gb_assert_handler("Assertion Failure", "index >= 0",
			"cmd_tokenizer_core.go", int64(0), "")
	}
	if int(index) < len(globalFiles) {
		return globalFiles[index]
	}
	return nil
}

func initGlobalErrorCollector() {
	globalErrorCollector.ErrorValues = make([]ErrorValue, 0)
	globalFilePathStrings = make([]String, 0, 4096)
	globalFiles = make([]*AstFile, 0, 4096)
}

func digitValue(r Rune) i32 {
	switch {
	case r >= '0' && r <= '9':
		return i32(r) - i32('0')
	case r >= 'a' && r <= 'f':
		return i32(r) - i32('a') + 10
	case r >= 'A' && r <= 'F':
		return i32(r) - i32('A') + 10
	}
	return 16
}

func scanMantissa(t *Tokenizer, base i32, forceBase bool) {
	if !forceBase {
		base = 16
	}
	for digitValue(t.currRune) < base || t.currRune == '_' {
		advanceToNextRune(t)
	}
}

func peekByte(t *Tokenizer, offset isize) u8 {
	endOff := uintptr(unsafe.Pointer(t.end))
	readOff := uintptr(unsafe.Pointer(t.readCurr))
	if readOff+uintptr(offset) < endOff {
		return *(*u8)(unsafe.Add(unsafe.Pointer(t.readCurr), int(offset)))
	}
	return 0
}

func tokenizerErr(t *Tokenizer, msg string) {
	column := t.columnMinusOne + 1
	if column < 1 {
		column = 1
	}
	pos := TokenPos{
		FileID: t.currFileID,
		Line:   t.lineCount,
		Column: column,
		Offset: i32(uintptr(unsafe.Pointer(t.readCurr)) - uintptr(unsafe.Pointer(t.start))),
	}
	syntaxErrorVa(pos, TokenPos{}, msg)
	t.errorCount++
}

func tokenizerErrPos(t *Tokenizer, pos TokenPos, msg string) {
	syntaxErrorVa(pos, TokenPos{}, msg)
	t.errorCount++
}

func advanceToNextRune(t *Tokenizer) {
	if t.currRune == '\n' {
		t.columnMinusOne = -1
		t.lineCount++
	}
	if uintptr(unsafe.Pointer(t.readCurr)) < uintptr(unsafe.Pointer(t.end)) {
		t.curr = t.readCurr
		first := *t.readCurr
		if first == 0 {
			tokenizerErr(t, "Illegal character NUL")
			t.readCurr = (*u8)(unsafe.Add(unsafe.Pointer(t.readCurr), 1))
			t.currRune = 0
		} else if first&0x80 != 0 {
			remaining := int(uintptr(unsafe.Pointer(t.end)) - uintptr(unsafe.Pointer(t.readCurr)))
			runeVal := Rune(first)
			width := utf8_decode(t.readCurr, remaining, &runeVal)
			t.readCurr = (*u8)(unsafe.Add(unsafe.Pointer(t.readCurr), width))
			if runeVal == 0xfffd && width == 1 {
				tokenizerErr(t, "Illegal UTF-8 encoding")
			} else if runeVal == 0xfeff && int(uintptr(unsafe.Pointer(t.curr))-uintptr(unsafe.Pointer(t.start))) > 0 {
				tokenizerErr(t, "Illegal byte order mark")
			}
			t.currRune = runeVal
		} else {
			t.readCurr = (*u8)(unsafe.Add(unsafe.Pointer(t.readCurr), 1))
			t.currRune = Rune(first)
		}
		t.columnMinusOne++
	} else {
		t.curr = t.end
		t.currRune = -1
	}
}

func initTokenizerWithData(t *Tokenizer, fullpath String, data unsafe.Pointer, size isize) {
	t.fullpath = fullpath
	t.columnMinusOne = -1
	t.lineCount = 1
	t.start = (*u8)(data)
	t.readCurr = t.start
	t.curr = t.start
	t.end = (*u8)(unsafe.Add(data, int(size)))
	advanceToNextRune(t)
	if t.currRune == 0xfeff {
		advanceToNextRune(t)
	}
}

func initTokenizerFromFullpath(t *Tokenizer, fullpath String, copyFileContents bool) TokenizerInitError {
	cpath := allocCstring(temporaryAllocator(), fullpath)
	var lf LoadedFile
	fileErr := loadFile32(cpath, &lf, copyFileContents)
	err := loadedFileErrorMapToTokenizer[int(fileErr)]
	switch fileErr {
	case LoadedFile_None:
		t.loadedFile = lf
		initTokenizerWithData(t, fullpath, lf.Data, isize(lf.Size))
	case LoadedFile_FileTooLarge, LoadedFile_Empty:
		t.fullpath = fullpath
		t.lineCount = 1
	}
	return err
}
