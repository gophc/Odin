// Depends on: common.odin (String, isize, i64, i32, u8, u16, u32, Rune, Array, AstFile, RecursiveMutex, BlockingMutex)
package cmd

import "sync/atomic"

type TokenKind uint8

const (
	TokenInvalid TokenKind = iota
	TokenEOF
	TokenComment
	TokenFileTag
	TokenLiteralBegin
	TokenIdent
	TokenInteger
	TokenFloat
	TokenImag
	TokenRune
	TokenString
	TokenLiteralEnd
	TokenOperatorBegin
	TokenEq
	TokenNot
	TokenHash
	TokenAt
	TokenDollar
	TokenPointer
	TokenQuestion
	TokenAdd
	TokenSub
	TokenMul
	TokenQuo
	TokenMod
	TokenModMod
	TokenAnd
	TokenOr
	TokenXor
	TokenAndNot
	TokenShl
	TokenShr
	TokenCmpAnd
	TokenCmpOr
	TokenAssignOpBegin
	TokenAddEq
	TokenSubEq
	TokenMulEq
	TokenQuoEq
	TokenModEq
	TokenModModEq
	TokenAndEq
	TokenOrEq
	TokenXorEq
	TokenAndNotEq
	TokenShlEq
	TokenShrEq
	TokenCmpAndEq
	TokenCmpOrEq
	TokenAssignOpEnd
	TokenIncrement
	TokenDecrement
	TokenArrowRight
	TokenUninit
	TokenComparisonBegin
	TokenCmpEq
	TokenNotEq
	TokenLt
	TokenGt
	TokenLtEq
	TokenGtEq
	TokenComparisonEnd
	TokenOpenParen
	TokenCloseParen
	TokenOpenBracket
	TokenCloseBracket
	TokenOpenBrace
	TokenCloseBrace
	TokenColon
	TokenSemicolon
	TokenPeriod
	TokenComma
	TokenEllipsis
	TokenRangeFull
	TokenRangeHalf
	TokenBackSlash
	TokenOperatorEnd
	TokenKeywordBegin
	TokenImport
	TokenForeign
	TokenPackage
	TokenTypeid
	TokenWhen
	TokenWhere
	TokenIf
	TokenElse
	TokenFor
	TokenSwitch
	TokenIn
	TokenNotIn
	TokenDo
	TokenCase
	TokenBreak
	TokenContinue
	TokenFallthrough
	TokenDefer
	TokenReturn
	TokenProc
	TokenStruct
	TokenUnion
	TokenEnum
	TokenBitSet
	TokenBitField
	TokenMap
	TokenDynamic
	TokenAutoCast
	TokenCast
	TokenTransmute
	TokenDistinct
	TokenUsing
	TokenContext
	TokenOrElse
	TokenOrReturn
	TokenOrBreak
	TokenOrContinue
	TokenAsm
	TokenMatrix
	TokenKeywordEnd
	TokenCount
)

var tokenStrings = []String{
	/*TokenInvalid*/    {Data: strData("Invalid"), Len: isize(len("Invalid"))},
	/*TokenEOF*/        {Data: strData("EOF"), Len: isize(len("EOF"))},
	/*TokenComment*/    {Data: strData("Comment"), Len: isize(len("Comment"))},
	/*TokenFileTag*/    {Data: strData("FileTag"), Len: isize(len("FileTag"))},
	/*TokenLiteralBegin*/ {},
	/*TokenIdent*/      {Data: strData("identifier"), Len: isize(len("identifier"))},
	/*TokenInteger*/    {Data: strData("integer"), Len: isize(len("integer"))},
	/*TokenFloat*/      {Data: strData("float"), Len: isize(len("float"))},
	/*TokenImag*/       {Data: strData("imaginary"), Len: isize(len("imaginary"))},
	/*TokenRune*/       {Data: strData("rune"), Len: isize(len("rune"))},
	/*TokenString*/     {Data: strData("string"), Len: isize(len("string"))},
	{},
	{},
	{Data: strData("="), Len: isize(len("="))},
	{Data: strData("!"), Len: isize(len("!"))},
	{Data: strData("#"), Len: isize(len("#"))},
	{Data: strData("@"), Len: isize(len("@"))},
	{Data: strData("$"), Len: isize(len("$"))},
	{Data: strData("^"), Len: isize(len("^"))},
	{Data: strData("?"), Len: isize(len("?"))},
	{Data: strData("+"), Len: isize(len("+"))},
	{Data: strData("-"), Len: isize(len("-"))},
	{Data: strData("*"), Len: isize(len("*"))},
	{Data: strData("/"), Len: isize(len("/"))},
	{Data: strData("%"), Len: isize(len("%"))},
	{Data: strData("%%"), Len: isize(len("%%"))},
	{Data: strData("&"), Len: isize(len("&"))},
	{Data: strData("|"), Len: isize(len("|"))},
	{Data: strData("~"), Len: isize(len("~"))},
	{Data: strData("&~"), Len: isize(len("&~"))},
	{Data: strData("<<"), Len: isize(len("<<"))},
	{Data: strData(">>"), Len: isize(len(">>"))},
	{Data: strData("&&"), Len: isize(len("&&"))},
	{Data: strData("||"), Len: isize(len("||"))},
	{},
	{Data: strData("+="), Len: isize(len("+="))},
	{Data: strData("-="), Len: isize(len("-="))},
	{Data: strData("*="), Len: isize(len("*="))},
	{Data: strData("/="), Len: isize(len("/="))},
	{Data: strData("%="), Len: isize(len("%="))},
	{Data: strData("%%="), Len: isize(len("%%="))},
	{Data: strData("&="), Len: isize(len("&="))},
	{Data: strData("|="), Len: isize(len("|="))},
	{Data: strData("~="), Len: isize(len("~="))},
	{Data: strData("&~="), Len: isize(len("&~="))},
	{Data: strData("<<="), Len: isize(len("<<="))},
	{Data: strData(">>="), Len: isize(len(">>="))},
	{Data: strData("&&="), Len: isize(len("&&="))},
	{Data: strData("||="), Len: isize(len("||="))},
	{},
	{Data: strData("++"), Len: isize(len("++"))},
	{Data: strData("--"), Len: isize(len("--"))},
	{Data: strData("->"), Len: isize(len("->"))},
	{Data: strData("---"), Len: isize(len("---"))},
	{},
	{Data: strData("=="), Len: isize(len("=="))},
	{Data: strData("!="), Len: isize(len("!="))},
	{Data: strData("<"), Len: isize(len("<"))},
	{Data: strData(">"), Len: isize(len(">"))},
	{Data: strData("<="), Len: isize(len("<="))},
	{Data: strData(">="), Len: isize(len(">="))},
	{},
	{Data: strData("("), Len: isize(len("("))},
	{Data: strData(")"), Len: isize(len(")"))},
	{Data: strData("["), Len: isize(len("["))},
	{Data: strData("]"), Len: isize(len("]"))},
	{Data: strData("{"), Len: isize(len("{"))},
	{Data: strData("}"), Len: isize(len("}"))},
	{Data: strData(":"), Len: isize(len(":"))},
	{Data: strData(";"), Len: isize(len(";"))},
	{Data: strData("."), Len: isize(len("."))},
	{Data: strData(","), Len: isize(len(","))},
	{Data: strData(".."), Len: isize(len(".."))},
	{Data: strData("..="), Len: isize(len("..="))},
	{Data: strData("..<"), Len: isize(len("..<"))},
	{Data: strData("\\"), Len: isize(len("\\"))},
	{},
	{},
	{Data: strData("import"), Len: isize(len("import"))},
	{Data: strData("foreign"), Len: isize(len("foreign"))},
	{Data: strData("package"), Len: isize(len("package"))},
	{Data: strData("typeid"), Len: isize(len("typeid"))},
	{Data: strData("when"), Len: isize(len("when"))},
	{Data: strData("where"), Len: isize(len("where"))},
	{Data: strData("if"), Len: isize(len("if"))},
	{Data: strData("else"), Len: isize(len("else"))},
	{Data: strData("for"), Len: isize(len("for"))},
	{Data: strData("switch"), Len: isize(len("switch"))},
	{Data: strData("in"), Len: isize(len("in"))},
	{Data: strData("not_in"), Len: isize(len("not_in"))},
	{Data: strData("do"), Len: isize(len("do"))},
	{Data: strData("case"), Len: isize(len("case"))},
	{Data: strData("break"), Len: isize(len("break"))},
	{Data: strData("continue"), Len: isize(len("continue"))},
	{Data: strData("fallthrough"), Len: isize(len("fallthrough"))},
	{Data: strData("defer"), Len: isize(len("defer"))},
	{Data: strData("return"), Len: isize(len("return"))},
	{Data: strData("proc"), Len: isize(len("proc"))},
	{Data: strData("struct"), Len: isize(len("struct"))},
	{Data: strData("union"), Len: isize(len("union"))},
	{Data: strData("enum"), Len: isize(len("enum"))},
	{Data: strData("bit_set"), Len: isize(len("bit_set"))},
	{Data: strData("bit_field"), Len: isize(len("bit_field"))},
	{Data: strData("map"), Len: isize(len("map"))},
	{Data: strData("dynamic"), Len: isize(len("dynamic"))},
	{Data: strData("auto_cast"), Len: isize(len("auto_cast"))},
	{Data: strData("cast"), Len: isize(len("cast"))},
	{Data: strData("transmute"), Len: isize(len("transmute"))},
	{Data: strData("distinct"), Len: isize(len("distinct"))},
	{Data: strData("using"), Len: isize(len("using"))},
	{Data: strData("context"), Len: isize(len("context"))},
	{Data: strData("or_else"), Len: isize(len("or_else"))},
	{Data: strData("or_return"), Len: isize(len("or_return"))},
	{Data: strData("or_break"), Len: isize(len("or_break"))},
	{Data: strData("or_continue"), Len: isize(len("or_continue"))},
	{Data: strData("asm"), Len: isize(len("asm"))},
	{Data: strData("matrix"), Len: isize(len("matrix"))},
	{},
}

type TokenFlag uint8

const (
	TokenFlagRemove  TokenFlag = 1 << 1
	TokenFlagReplace TokenFlag = 1 << 2
)

type TokenPos struct {
	FileID int32
	Offset int32
	Line   int32
	Column int32
}

type Token struct {
	Kind   TokenKind
	Flags  TokenFlag
	String String
	Pos    TokenPos
}

var EmptyToken = Token{Kind: TokenInvalid}
var BlankToken = Token{Kind: TokenIdent, String: String{Data: strData("_"), Len: 1}}

type KeywordHashEntry struct {
	Hash uint32
	Kind TokenKind
	Text String
}

type TokenizerInitError int

const (
	TokenizerInitNone        TokenizerInitError = 0
	TokenizerInitInvalid     TokenizerInitError = 1
	TokenizerInitNotExists   TokenizerInitError = 2
	TokenizerInitPermission  TokenizerInitError = 3
	TokenizerInitEmpty       TokenizerInitError = 4
	TokenizerInitFileTooLarge TokenizerInitError = 5
	TokenizerInitCount
)

type ErrorValueKind uint32

const (
	ErrorValueError   ErrorValueKind = 0
	ErrorValueWarning ErrorValueKind = 1
)

type ErrorValue struct {
	Kind         ErrorValueKind
	Pos          TokenPos
	End          TokenPos
	Msg          []byte
	SeenNewline  bool
}

type ErrorCollector struct {
	Count            atomic.Int64
	WarningCount     atomic.Int64
	InBlock          atomic.Bool
	Mutex            RecursiveMutex
	PathMutex        BlockingMutex
	ErrorValues      []ErrorValue
	CurrErrorValue   ErrorValue
	CurrErrorValueSet atomic.Bool
}

var globalErrorCollector ErrorCollector

var errorsAlreadyPrinted bool
