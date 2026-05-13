// Depends on: common.odin, cmd_tokenizer_types.go, cmd_build_settings_types.go
// (String, Token, TokenPos, Tokenizer, TokenKind, ExactValue, BuildContext, BlockingMutex, MPMCQueue, Scope, DeclInfo, Entity, Type, FileInfo, isize, i64, i32, u8, u16, u32, u64, Rune, Slice, Array, gbString, gbAllocator, InternedString, std::atomic → sync/atomic)
package cmd

import "sync/atomic"

type AddressingMode uint8

const (
	AddressingInvalid       AddressingMode = 0
	AddressingNoValue       AddressingMode = 1
	AddressingValue         AddressingMode = 2
	AddressingContext       AddressingMode = 3
	AddressingVariable      AddressingMode = 4
	AddressingConstant      AddressingMode = 5
	AddressingType          AddressingMode = 6
	AddressingBuiltin       AddressingMode = 7
	AddressingProcGroup     AddressingMode = 8
	AddressingMapIndex      AddressingMode = 9
	AddressingOptionalOk    AddressingMode = 10
	AddressingOptionalOkPtr AddressingMode = 11
	AddressingSoaVariable   AddressingMode = 12
	AddressingSwizzleValue    AddressingMode = 13
	AddressingSwizzleVariable AddressingMode = 14
)

var addressingModeStrings = []String{
	{Data: strData("Invalid"), Len: 7}, {Data: strData("NoValue"), Len: 7}, {Data: strData("Value"), Len: 5}, {Data: strData("Context"), Len: 7}, {Data: strData("Variable"), Len: 8}, {Data: strData("Constant"), Len: 8}, {Data: strData("Type"), Len: 4}, {Data: strData("Builtin"), Len: 7}, {Data: strData("ProcGroup"), Len: 9}, {Data: strData("MapIndex"), Len: 8}, {Data: strData("OptionalOk"), Len: 10}, {Data: strData("OptionalOkPtr"), Len: 13}, {Data: strData("SoaVariable"), Len: 11}, {Data: strData("SwizzleValue"), Len: 12}, {Data: strData("SwizzleVariable"), Len: 15},
}

type TypeAndValue struct {
	Type           *Type
	Mode           AddressingMode
	IsLHS          bool
	ObjCSuperTarget *Type
	Value          ExactValue
}

type ParseFileError int

const (
	ParseFileNone                ParseFileError = 0
	ParseFileWrongExtension      ParseFileError = 1
	ParseFileInvalidFile         ParseFileError = 2
	ParseFileEmptyFile           ParseFileError = 3
	ParseFilePermission          ParseFileError = 4
	ParseFileNotFound            ParseFileError = 5
	ParseFileInvalidToken        ParseFileError = 6
	ParseFileGeneralError        ParseFileError = 7
	ParseFileFileTooLarge        ParseFileError = 8
	ParseFileDirectoryAlreadyExists ParseFileError = 9
	ParseFileCount
)

type CommentGroup struct {
	List []Token
}

type PackageKind int

const (
	PackageNormal  PackageKind = 0
	PackageRuntime PackageKind = 1
	PackageInit    PackageKind = 2
	PackageBuiltin PackageKind = 3
)

type ImportedFile struct {
	Pkg   *AstPackage
	Fi    FileInfo
	Pos   TokenPos
	Index isize
}

type AstFileFlag uint32

const (
	AstFileIsPrivatePkg   AstFileFlag = 1 << 0
	AstFileIsPrivateFile  AstFileFlag = 1 << 1
	AstFileIsLazy         AstFileFlag = 1 << 4
	AstFileNoInstrumentation AstFileFlag = 1 << 5
)

type AstDelayQueueKind int

const (
	AstDelayQueueImport      AstDelayQueueKind = 0
	AstDelayQueueExpr        AstDelayQueueKind = 1
	AstDelayQueueForeignBlock AstDelayQueueKind = 2
	AstDelayQueueCOUNT
)

type AstFile struct {
	ID                     int32
	Flags                  uint32
	Pkg                    *AstPackage
	Scope                  *Scope
	PkgDecl                *Ast
	Fullpath               String
	Filename               String
	Directory              String
	Tokenizer              Tokenizer
	Tokens                 []Token
	CurrTokenIndex         isize
	PrevTokenIndex         isize
	CurrToken              Token
	PrevToken              Token
	PackageToken           Token
	PackageName            String
	VetFlags               uint64
	FeatureFlags           uint64
	VetFlagsSet            bool
	FeatureFlagsSet        bool
	ExprLevel              isize
	AllowNewline           bool
	AllowRange             bool
	AllowInExpr            bool
	InForeignBlock         bool
	AllowType              bool
	InWhenStatement        bool
	TotalFileDeclCount     isize
	DelayedDeclCount       isize
	Decls                  []*Ast
	Imports                []*Ast
	DirectiveCount         isize
	CurrProc               *Ast
	ErrorCount             isize
	LastError              ParseFileError
	TimeToTokenize         float64
	TimeToParse            float64
	LeadComment            *CommentGroup
	LineComment            *CommentGroup
	Docs                   *CommentGroup
	Comments               []*CommentGroup
	DelayedDeclsQueues     [AstDelayQueueCOUNT][]*Ast
	SeenLoadDirectiveCount atomic.Int64
	FixCount               isize
	FixPrevPos             TokenPos
	LLVMMetadata           uintptr
	LLVMMetadataScope      uintptr
}

type AstForeignFileKind int

const (
	AstForeignFileInvalid AstForeignFileKind = 0
	AstForeignFileS       AstForeignFileKind = 1
	AstForeignFileCOUNT
)

type AstForeignFile struct {
	Kind   AstForeignFileKind
	Source String
}

type AstPackageExportedEntity struct {
	Identifier *Ast
	Entity     *Entity
}

type AstPackage struct {
	Kind                   PackageKind
	ID                     isize
	Name                   String
	Fullpath               String
	Files                  []*AstFile
	ForeignFiles           []AstForeignFile
	IsSingleFile           bool
	Order                  isize
	FilesMutex             BlockingMutex
	ForeignFilesMutex      BlockingMutex
	TypeAndValueMutex      BlockingMutex
	NameMutex              BlockingMutex
	ExportedEntityQueue    MPMCQueue[AstPackageExportedEntity]
	Scope                  *Scope
	DeclInfo               *DeclInfo
	IsExtra                bool
}

type ParseFileErrorNode struct {
	Next *ParseFileErrorNode
	Prev *ParseFileErrorNode
	Err  ParseFileError
}

type Parser struct {
	InitFullpath              String
	ImportedFiles             StringSet
	ImportedFilesMutex        BlockingMutex
	Packages                  []*AstPackage
	PackagesMutex             BlockingMutex
	FileToProcessCount        atomic.Int64
	TotalTokenCount           atomic.Int64
	TotalLineCount            atomic.Int64
	TotalSeenLoadDirectiveCount atomic.Int64
	FileDeclMutex             BlockingMutex
	FileErrorMutex            BlockingMutex
	FileErrorHead             *ParseFileErrorNode
	FileErrorTail             *ParseFileErrorNode
}

type ParserWorkerData struct {
	Parser        *Parser
	ImportedFile  ImportedFile
}

type ForeignFileWorkerData struct {
	Parser       *Parser
	ImportedFile ImportedFile
	ForeignKind  AstForeignFileKind
}

type AstKind uint16

const (
	AstInvalid AstKind = iota
	AstIdent
	AstImplicit
	AstUninit
	AstBasicLit
	AstBasicDirective
	AstEllipsis
	AstProcGroup
	AstProcLit
	AstCompoundLit
	AstExprBegin
	AstBadExpr
	AstTagExpr
	AstUnaryExpr
	AstBinaryExpr
	AstParenExpr
	AstSelectorExpr
	AstImplicitSelectorExpr
	AstSelectorCallExpr
	AstIndexExpr
	AstDerefExpr
	AstSliceExpr
	AstCallExpr
	AstFieldValue
	AstEnumFieldValue
	AstTernaryIfExpr
	AstTernaryWhenExpr
	AstOrElseExpr
	AstOrReturnExpr
	AstOrBranchExpr
	AstTypeAssertion
	AstTypeCast
	AstAutoCast
	AstInlineAsmExpr
	AstMatrixIndexExpr
	AstExprEnd
	AstStmtBegin
	AstBadStmt
	AstEmptyStmt
	AstExprStmt
	AstAssignStmt
	AstComplexStmtBegin
	AstBlockStmt
	AstIfStmt
	AstWhenStmt
	AstReturnStmt
	AstForStmt
	AstRangeStmt
	AstUnrollRangeStmt
	AstCaseClause
	AstSwitchStmt
	AstTypeSwitchStmt
	AstDeferStmt
	AstBranchStmt
	AstUsingStmt
	AstComplexStmtEnd
	AstStmtEnd
	AstDeclBegin
	AstBadDecl
	AstForeignBlockDecl
	AstLabel
	AstValueDecl
	AstPackageDecl
	AstImportDecl
	AstForeignImportDecl
	AstDeclEnd
	AstAttribute
	AstField
	AstBitFieldField
	AstFieldList
	AstTypeBegin
	AstTypeidType
	AstHelperType
	AstDistinctType
	AstPolyType
	AstProcType
	AstPointerType
	AstRelativeType
	AstMultiPointerType
	AstArrayType
	AstDynamicArrayType
	AstFixedCapacityDynamicArrayType
	AstStructType
	AstUnionType
	AstEnumType
	AstBitSetType
	AstBitFieldType
	AstMapType
	AstMatrixType
	AstTypeEnd
	AstCOUNT
)

var astStrings []String // will be initialized inline as a large init function

// AstIdent through AstMatrixIndexExpr sub-structs
type AstIdent struct {
	Token    Token
	Entity   atomic.Pointer[Entity]
	Hash     uint32
	Interned InternedString
}

// Implicit, Uninit, BasicLit: just a Token
// BasicDirective: Token token, name
type AstBasicDirective struct {
	Token Token
	Name  Token
}

type AstEllipsis struct {
	Token Token
	Expr  *Ast
}

type AstProcGroup struct {
	Token Token
	Open  Token
	Close Token
	Args  []*Ast
}

type AstProcLit struct {
	Type         *Ast
	Body         *Ast
	Tags         uint64
	Inlining     ProcInlining
	Tailing      ProcTailing
	WhereToken   Token
	WhereClauses []*Ast
	Decl         *DeclInfo
}

type AstCompoundLit struct {
	Type     *Ast
	Elems    []*Ast
	Open     Token
	Close    Token
	MaxCount int64
	Tag      *Ast
}

type AstBadExpr struct {
	Begin Token
	End   Token
}

type AstTagExpr struct {
	Token Token
	Name  Token
	Expr  *Ast
}

type AstUnaryExpr struct {
	Op   Token
	Expr *Ast
}

type AstBinaryExpr struct {
	Op    Token
	Left  *Ast
	Right *Ast
}

type AstParenExpr struct {
	Expr  *Ast
	Open  Token
	Close Token
}

type AstSelectorExpr struct {
	Token          Token
	Expr           *Ast
	Selector       *Ast
	SwizzleCount   uint8
	SwizzleIndices uint8
	IsBitField     bool
}

type AstImplicitSelectorExpr struct {
	Token    Token
	Selector *Ast
}

type AstSelectorCallExpr struct {
	Token        Token
	Expr         *Ast
	Call         *Ast
	ModifiedCall bool
}

type AstIndexExpr struct {
	Expr  *Ast
	Index *Ast
	Open  Token
	Close Token
}

type AstDerefExpr struct {
	Expr *Ast
	Op   Token
}

type AstSliceExpr struct {
	Expr     *Ast
	Open     Token
	Close    Token
	Interval Token
	Low      *Ast
	High     *Ast
}

type AstCallExpr struct {
	Proc               *Ast
	Args               []*Ast
	Open               Token
	Close              Token
	Ellipsis           Token
	Inlining           ProcInlining
	Tailing            ProcTailing
	OptionalOkOne      bool
	WasSelector        bool
	SplitArgs          *AstSplitArgs
	EntityProcedureOf  atomic.Pointer[Entity]
}

type AstFieldValue struct {
	Eq     Token
	Field  *Ast
	Value  *Ast
}

type AstEnumFieldValue struct {
	Name    *Ast
	Value   *Ast
	Docs    *CommentGroup
	Comment *CommentGroup
}

type AstTernaryIfExpr struct {
	X    *Ast
	Cond *Ast
	Y    *Ast
}

type AstTernaryWhenExpr struct {
	X    *Ast
	Cond *Ast
	Y    *Ast
}

type AstOrElseExpr struct {
	X     *Ast
	Token Token
	Y     *Ast
}

type AstOrReturnExpr struct {
	Expr  *Ast
	Token Token
}

type AstOrBranchExpr struct {
	Expr  *Ast
	Token Token
	Label *Ast
}

type AstTypeAssertion struct {
	Expr *Ast
	Dot  Token
	Type *Ast
}

type AstTypeCast struct {
	Token Token
	Type  *Ast
	Expr  *Ast
}

type AstAutoCast struct {
	Token Token
	Expr  *Ast
}

type AstInlineAsmExpr struct {
	Token            Token
	Open             Token
	Close            Token
	ParamTypes       []*Ast
	ReturnType       *Ast
	AsmString        *Ast
	ConstraintsString *Ast
	HasSideEffects   bool
	IsAlignStack     bool
	Dialect          InlineAsmDialectKind
}

type AstMatrixIndexExpr struct {
	Expr         *Ast
	RowIndex     *Ast
	ColumnIndex  *Ast
	Open         Token
	Close        Token
}

// Stmt types
type AstBadStmt struct {
	Begin Token
	End   Token
}

type AstEmptyStmt struct {
	Token Token
}

type AstExprStmt struct {
	Expr *Ast
}

type AstAssignStmt struct {
	Op  Token
	LHS []*Ast
	RHS []*Ast
}

type AstBlockStmt struct {
	Label *Ast
	Stmts []*Ast
	Open  Token
	Close Token
}

type AstIfStmt struct {
	Token     Token
	Label     *Ast
	Init      *Ast
	Cond      *Ast
	Body      *Ast
	ElseStmt  *Ast
}

type AstWhenStmt struct {
	Token     Token
	Cond      *Ast
	Body      *Ast
	ElseStmt  *Ast
}

type AstReturnStmt struct {
	Token   Token
	Results []*Ast
}

type AstForStmt struct {
	Token Token
	Label *Ast
	Init  *Ast
	Cond  *Ast
	Post  *Ast
	Body  *Ast
}

type AstRangeStmt struct {
	Token   Token
	Label   *Ast
	Init    *Ast
	Vals    []*Ast
	InToken Token
	Expr    *Ast
	Body    *Ast
}

type AstUnrollRangeStmt struct {
	UnrollToken Token
	Init        *Ast
	Args        []*Ast
	ForToken    Token
	Val0        *Ast
	Val1        *Ast
	InToken     Token
	Expr        *Ast
	Body        *Ast
}

type AstCaseClause struct {
	Token           Token
	List            []*Ast
	Stmts           []*Ast
	ImplicitEntity  atomic.Pointer[Entity]
}

type AstSwitchStmt struct {
	Token   Token
	Label   *Ast
	Init    *Ast
	Tag     *Ast
	Body    *Ast
	Partial bool
}

type AstTypeSwitchStmt struct {
	Token   Token
	Label   *Ast
	Tag     *Ast
	Body    *Ast
	Partial bool
}

type AstDeferStmt struct {
	Token Token
	Stmt  *Ast
}

type AstBranchStmt struct {
	Token Token
	Label *Ast
}

type AstUsingStmt struct {
	Token Token
	List  []*Ast
}

// Decl types
type AstBadDecl struct {
	Begin Token
	End   Token
}

type AstForeignBlockDecl struct {
	Token           Token
	ForeignLibrary  *Ast
	Body            *Ast
	Docs            *CommentGroup
	Attributes      []*Ast
}

type AstLabel struct {
	Token Token
	Name  *Ast
}

type AstValueDecl struct {
	Names      []*Ast
	Type       *Ast
	Values     []*Ast
	IsMutable  bool
	Docs       *CommentGroup
	Comment    *CommentGroup
	Attributes []*Ast
}

type AstPackageDecl struct {
	Token   Token
	Name    Token
	Docs    *CommentGroup
	Comment *CommentGroup
}

type AstImportDecl struct {
	Token       Token
	Relpath     Token
	ImportName  Token
	Docs        *CommentGroup
	Comment     *CommentGroup
	Attributes  []*Ast
	Fullpath    String
	Package     *AstPackage
}

type AstForeignImportDecl struct {
	Token            Token
	Filepaths        []*Ast
	LibraryName      Token
	Docs             *CommentGroup
	Comment          *CommentGroup
	MultipleFilepaths bool
	Attributes       []*Ast
	Paths            []String
	ExtraLinkerFlags String
	IgnoreDuplicates bool
}

type AstAttribute struct {
	Token Token
	Open  Token
	Close Token
	Elems []*Ast
}

type AstField struct {
	Names        []*Ast
	Type         *Ast
	DefaultValue *Ast
	Flags        uint32
	Tag          Token
	Docs         *CommentGroup
	Comment      *CommentGroup
}

type AstBitFieldField struct {
	Name    *Ast
	Type    *Ast
	BitSize *Ast
	Tag     Token
	Docs    *CommentGroup
	Comment *CommentGroup
}

type AstFieldList struct {
	Token Token
	List  []*Ast
}

// Type AST nodes
type AstTypeidType struct {
	Token          Token
	Specialization *Ast
}

type AstHelperType struct {
	Token Token
	Type  *Ast
}

type AstDistinctType struct {
	Token Token
	Type  *Ast
}

type AstPolyType struct {
	Token          Token
	Type           *Ast
	Specialization *Ast
}

type AstProcType struct {
	Token              Token
	Params             *Ast
	Results            *Ast
	Tags               uint64
	CallingConvention  ProcCallingConvention
	Generic            bool
	Diverging          bool
}

type AstPointerType struct {
	Token Token
	Type  *Ast
	Tag   *Ast
}

type AstRelativeType struct {
	Tag  *Ast
	Type *Ast
}

type AstMultiPointerType struct {
	Token Token
	Type  *Ast
}

type AstArrayType struct {
	Token Token
	Count *Ast
	Elem  *Ast
	Tag   *Ast
}

type AstDynamicArrayType struct {
	Token Token
	Elem  *Ast
	Tag   *Ast
}

type AstFixedCapacityDynamicArrayType struct {
	Token    Token
	Capacity *Ast
	Elem     *Ast
	Tag      *Ast
}

type AstStructType struct {
	Token              Token
	Fields             []*Ast
	FieldCount         isize
	PolymorphicParams *Ast
	IsPacked           bool
	IsRawUnion         bool
	IsAllOrNone        bool
	IsSimple           bool
	Align              *Ast
	MinFieldAlign      *Ast
	MaxFieldAlign      *Ast
	WhereToken         Token
	WhereClauses       []*Ast
}

type AstUnionType struct {
	Token              Token
	Variants           []*Ast
	PolymorphicParams *Ast
	Align              *Ast
	Kind               UnionTypeKind
	WhereToken         Token
	WhereClauses       []*Ast
}

type AstEnumType struct {
	Token    Token
	BaseType *Ast
	Fields   []*Ast
}

type AstBitSetType struct {
	Token      Token
	Elem       *Ast
	Underlying *Ast
}

type AstBitFieldType struct {
	Token       Token
	BackingType *Ast
	Open        Token
	Fields      []*Ast
	Close       Token
}

type AstMapType struct {
	Token Token
	Count *Ast
	Key   *Ast
	Value *Ast
}

type AstMatrixType struct {
	Token       Token
	RowCount    *Ast
	ColumnCount *Ast
	Elem        *Ast
	IsRowMajor  bool
}

type AstCommonStuff struct {
	Kind            AstKind
	StateFlags      uint8
	ViralStateFlags atomic.Uint32
	FileID          int32
	TAV             TypeAndValue
}

type Ast struct {
	Kind            AstKind
	StateFlags      uint8
	ViralStateFlags atomic.Uint32
	FileID          int32
	TAV             TypeAndValue
	// Embedded sub-structs
	Ident                AstIdent
	Implicit             Token
	Uninit               Token
	BasicLit             AstBasicLit
	BasicDirective       AstBasicDirective
	Ellipsis             AstEllipsis
	ProcGroup            AstProcGroup
	ProcLit              AstProcLit
	CompoundLit          AstCompoundLit
	BadExpr              AstBadExpr
	TagExpr              AstTagExpr
	UnaryExpr            AstUnaryExpr
	BinaryExpr           AstBinaryExpr
	ParenExpr            AstParenExpr
	SelectorExpr         AstSelectorExpr
	ImplicitSelectorExpr AstImplicitSelectorExpr
	SelectorCallExpr     AstSelectorCallExpr
	IndexExpr            AstIndexExpr
	DerefExpr            AstDerefExpr
	SliceExpr            AstSliceExpr
	CallExpr             AstCallExpr
	FieldValue           AstFieldValue
	EnumFieldValue       AstEnumFieldValue
	TernaryIfExpr        AstTernaryIfExpr
	TernaryWhenExpr      AstTernaryWhenExpr
	OrElseExpr           AstOrElseExpr
	OrReturnExpr         AstOrReturnExpr
	OrBranchExpr         AstOrBranchExpr
	TypeAssertion        AstTypeAssertion
	TypeCast             AstTypeCast
	AutoCast             AstAutoCast
	InlineAsmExpr        AstInlineAsmExpr
	MatrixIndexExpr      AstMatrixIndexExpr
	BadStmt              AstBadStmt
	EmptyStmt            AstEmptyStmt
	ExprStmt             AstExprStmt
	AssignStmt           AstAssignStmt
	BlockStmt            AstBlockStmt
	IfStmt               AstIfStmt
	WhenStmt             AstWhenStmt
	ReturnStmt           AstReturnStmt
	ForStmt              AstForStmt
	RangeStmt            AstRangeStmt
	UnrollRangeStmt      AstUnrollRangeStmt
	CaseClause           AstCaseClause
	SwitchStmt           AstSwitchStmt
	TypeSwitchStmt       AstTypeSwitchStmt
	DeferStmt            AstDeferStmt
	BranchStmt           AstBranchStmt
	UsingStmt            AstUsingStmt
	BadDecl              AstBadDecl
	ForeignBlockDecl     AstForeignBlockDecl
	Label                AstLabel
	ValueDecl            AstValueDecl
	PackageDecl          AstPackageDecl
	ImportDecl           AstImportDecl
	ForeignImportDecl    AstForeignImportDecl
	Attribute            AstAttribute
	Field                AstField
	BitFieldField        AstBitFieldField
	FieldList            AstFieldList
	TypeidType           AstTypeidType
	HelperType           AstHelperType
	DistinctType         AstDistinctType
	PolyType             AstPolyType
	ProcType             AstProcType
	PointerType          AstPointerType
	RelativeType         AstRelativeType
	MultiPointerType     AstMultiPointerType
	ArrayType            AstArrayType
	DynamicArrayType     AstDynamicArrayType
	FixedCapacityDynamicArrayType AstFixedCapacityDynamicArrayType
	StructType           AstStructType
	UnionType            AstUnionType
	EnumType             AstEnumType
	BitSetType           AstBitSetType
	BitFieldType         AstBitFieldType
	MapType              AstMapType
	MatrixType           AstMatrixType
}

// Basic lit
type AstBasicLit struct {
	Token Token
}

// Split args
type AstSplitArgs struct {
	Positional []*Ast
	Named      []*Ast
}

type ProcInlining uint8

const (
	ProcInliningNone     ProcInlining = 0
	ProcInliningInline   ProcInlining = 1
	ProcInliningNoInline ProcInlining = 2
)

type ProcTailing uint8

const (
	ProcTailingNone     ProcTailing = 0
	ProcTailingMustTail ProcTailing = 1
)

type ProcTag uint64

const (
	ProcTagBoundsCheck            ProcTag = 1 << 0
	ProcTagNoBoundsCheck          ProcTag = 1 << 1
	ProcTagTypeAssert             ProcTag = 1 << 2
	ProcTagNoTypeAssert           ProcTag = 1 << 3
	ProcTagRequireResults         ProcTag = 1 << 4
	ProcTagOptionalOk             ProcTag = 1 << 5
	ProcTagOptionalAllocatorError ProcTag = 1 << 6
)

type ProcCallingConvention int32

const (
	ProcCCInvalid          ProcCallingConvention = 0
	ProcCCOdin             ProcCallingConvention = 1
	ProcCCContextless      ProcCallingConvention = 2
	ProcCCCDecl            ProcCallingConvention = 3
	ProcCCStdCall          ProcCallingConvention = 4
	ProcCCFastCall         ProcCallingConvention = 5
	ProcCCNone             ProcCallingConvention = 6
	ProcCCNaked            ProcCallingConvention = 7
	ProcCCInlineAsm        ProcCallingConvention = 8
	ProcCCWin64            ProcCallingConvention = 9
	ProcCCSysV             ProcCallingConvention = 10
	ProcCCPreserveNone     ProcCallingConvention = 11
	ProcCCPreserveMost     ProcCallingConvention = 12
	ProcCCPreserveAll      ProcCallingConvention = 13
	ProcCCMAX              ProcCallingConvention = 14
	ProcCCForeignBlockDefault ProcCallingConvention = -1
)

var procCallingConventionStrings = []string{
	"", "odin", "contextless", "cdecl", "stdcall", "fastcall", "none", "naked", "inlineasm",
	"win64", "sysv", "preserve/none", "preserve/most", "preserve/all",
}

func default_calling_convention() ProcCallingConvention {
	return ProcCCOdin
}

type StateFlag uint8

const (
	StateFlagBoundsCheck        StateFlag = 1 << 0
	StateFlagNoBoundsCheck      StateFlag = 1 << 1
	StateFlagTypeAssert         StateFlag = 1 << 2
	StateFlagNoTypeAssert       StateFlag = 1 << 3
	StateFlagSelectorCallExpr   StateFlag = 1 << 5
	StateFlagDirectiveWasFalse  StateFlag = 1 << 6
	StateFlagBeenHandled         StateFlag = 1 << 7
)

type ViralStateFlag uint8

const (
	ViralStateFlagContainsDeferredProcedure ViralStateFlag = 1 << 0
	ViralStateFlagContainsOrBreak           ViralStateFlag = 1 << 1
	ViralStateFlagContainsOrReturn          ViralStateFlag = 1 << 2
)

type FieldFlag uint32

const (
	FieldFlagNONE        FieldFlag = 0
	FieldFlagEllipsis    FieldFlag = 1 << 0
	FieldFlagUsing       FieldFlag = 1 << 1
	FieldFlagNoAlias     FieldFlag = 1 << 2
	FieldFlagCVararg     FieldFlag = 1 << 3
	FieldFlagConst       FieldFlag = 1 << 5
	FieldFlagAnyInt      FieldFlag = 1 << 6
	FieldFlagSubtype     FieldFlag = 1 << 7
	FieldFlagByPtr       FieldFlag = 1 << 8
	FieldFlagNoBroadcast FieldFlag = 1 << 9
	FieldFlagNoCapture   FieldFlag = 1 << 11
	FieldFlagTags        FieldFlag = 1 << 15
	FieldFlagResults     FieldFlag = 1 << 16
	FieldFlagUnknown     FieldFlag = 1 << 30
	FieldFlagInvalid     FieldFlag = 1 << 31
	FieldFlagSignature   FieldFlag = FieldFlagEllipsis | FieldFlagUsing | FieldFlagNoAlias | FieldFlagCVararg | FieldFlagConst | FieldFlagAnyInt | FieldFlagByPtr | FieldFlagNoBroadcast | FieldFlagNoCapture
	FieldFlagStruct      FieldFlag = FieldFlagUsing | FieldFlagSubtype | FieldFlagTags
)

type StmtAllowFlag int

const (
	StmtAllowFlagNone  StmtAllowFlag = 0
	StmtAllowFlagIn    StmtAllowFlag = 1 << 0
	StmtAllowFlagLabel StmtAllowFlag = 1 << 1
)

type InlineAsmDialectKind uint8

const (
	InlineAsmDialectDefault InlineAsmDialectKind = 0
	InlineAsmDialectATT     InlineAsmDialectKind = 1
	InlineAsmDialectIntel   InlineAsmDialectKind = 2
	InlineAsmDialectCOUNT
)

var inlineAsmDialectStrings = []string{"", "att", "intel"}

type UnionTypeKind uint8

const (
	UnionTypeNormal     UnionTypeKind = 0
	UnionTypeNoNil      UnionTypeKind = 2
	UnionTypeSharedNil  UnionTypeKind = 3
	UnionTypeCOUNT
)

var unionTypeKindStrings = []string{"(normal)", "#maybe", "#no_nil", "#shared_nil"}

var gParsingDone atomic.Bool
