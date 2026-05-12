package godin

import "core:mem"
import "core:sync"

// ============================================================================
// Forward declarations — types from other already-rewritten Odin files
// ============================================================================

Token        :: distinct u64       // placeholder token index / bit-packed token
TokenPos     :: distinct int       // placeholder source position
Tokenizer    :: struct {}          // placeholder tokenizer (defined in tokenizer.odin)
FileInfo     :: struct {}          // placeholder file info (defined in common.odin)
ExactValue   :: struct {}          // placeholder exact compile-time value
Scope        :: struct {}          // placeholder scope
Type         :: struct {}          // placeholder type
Entity       :: struct {}          // placeholder entity
DeclInfo     :: struct {}          // placeholder declaration info
InternedString :: distinct string  // placeholder interned string

// Opaque LLVM types
LLVMOpaqueMetadata :: rawptr

// ============================================================================
// Constants
// ============================================================================

// Repeated-string constants used in ast_strings
AST_NODE_INVALID_STR       :: "invalid node"
AST_NODE_IDENT_STR         :: "identifier"
AST_NODE_IMPLICIT_STR      :: "implicit"
AST_NODE_UNINIT_STR        :: "uninitialized value"
AST_NODE_BASIC_LIT_STR     :: "basic literal"
AST_NODE_BASIC_DIR_STR     :: "basic directive"
AST_NODE_ELLIPSIS_STR      :: "ellipsis"
AST_NODE_PROC_GROUP_STR    :: "procedure group"
AST_NODE_PROC_LIT_STR      :: "procedure literal"
AST_NODE_COMPOUND_LIT_STR  :: "compound literal"
AST_NODE_BAD_EXPR_STR      :: "bad expression"
AST_NODE_TAG_EXPR_STR      :: "tag expression"
AST_NODE_UNARY_EXPR_STR    :: "unary expression"
AST_NODE_BINARY_EXPR_STR   :: "binary expression"
AST_NODE_PAREN_EXPR_STR    :: "parentheses expression"
AST_NODE_SELECTOR_EXPR_STR :: "selector expression"
AST_NODE_IMP_SEL_EXPR_STR  :: "implicit selector expression"
AST_NODE_SEL_CALL_EXPR_STR :: "selector call expression"
AST_NODE_INDEX_EXPR_STR    :: "index expression"
AST_NODE_DEREF_EXPR_STR    :: "dereference expression"
AST_NODE_SLICE_EXPR_STR    :: "slice expression"
AST_NODE_CALL_EXPR_STR     :: "call expression"
AST_NODE_FIELD_VAL_STR     :: "field value"
AST_NODE_ENUM_FIELD_STR    :: "enum field value"
AST_NODE_TERN_IF_STR       :: "ternary if expression"
AST_NODE_TERN_WHEN_STR     :: "ternary when expression"
AST_NODE_OR_ELSE_STR       :: "or_else expression"
AST_NODE_OR_RETURN_STR     :: "or_return expression"
AST_NODE_OR_BRANCH_STR     :: "or branch expression"
AST_NODE_TYPE_ASSERT_STR   :: "type assertion"
AST_NODE_TYPE_CAST_STR     :: "type cast"
AST_NODE_AUTO_CAST_STR     :: "auto_cast"
AST_NODE_INLINE_ASM_STR    :: "inline asm expression"
AST_NODE_MATRIX_INDEX_STR  :: "matrix index expression"
AST_NODE_BAD_STMT_STR      :: "bad statement"
AST_NODE_EMPTY_STMT_STR    :: "empty statement"
AST_NODE_EXPR_STMT_STR     :: "expression statement"
AST_NODE_ASSIGN_STMT_STR   :: "assign statement"
AST_NODE_BLOCK_STMT_STR    :: "block statement"
AST_NODE_IF_STMT_STR       :: "if statement"
AST_NODE_WHEN_STMT_STR     :: "when statement"
AST_NODE_RETURN_STMT_STR   :: "return statement"
AST_NODE_FOR_STMT_STR      :: "for statement"
AST_NODE_RANGE_STMT_STR    :: "range statement"
AST_NODE_UNROLL_RANGE_STR  :: "#unroll range statement"
AST_NODE_CASE_CLAUSE_STR   :: "case clause"
AST_NODE_SWITCH_STMT_STR   :: "switch statement"
AST_NODE_TYPE_SWITCH_STR   :: "type switch statement"
AST_NODE_DEFER_STMT_STR    :: "defer statement"
AST_NODE_BRANCH_STMT_STR   :: "branch statement"
AST_NODE_USING_STMT_STR    :: "using statement"
AST_NODE_BAD_DECL_STR      :: "bad declaration"
AST_NODE_FOREIGN_BLOCK_STR :: "foreign block declaration"
AST_NODE_LABEL_STR         :: "label"
AST_NODE_VALUE_DECL_STR    :: "value declaration"
AST_NODE_PACKAGE_DECL_STR  :: "package declaration"
AST_NODE_IMPORT_DECL_STR   :: "import declaration"
AST_NODE_FOREIGN_IMP_STR   :: "foreign import declaration"
AST_NODE_ATTRIBUTE_STR     :: "attribute"
AST_NODE_FIELD_STR         :: "field"
AST_NODE_BIT_FIELD_STR     :: "bit field field"
AST_NODE_FIELD_LIST_STR    :: "field list"
AST_NODE_TYPEID_STR        :: "typeid"
AST_NODE_HELPER_TYPE_STR   :: "helper type"
AST_NODE_DISTINCT_TYPE_STR :: "distinct type"
AST_NODE_POLY_TYPE_STR     :: "polymorphic type"
AST_NODE_PROC_TYPE_STR     :: "procedure type"
AST_NODE_POINTER_TYPE_STR  :: "pointer type"
AST_NODE_RELATIVE_TYPE_STR :: "relative type"
AST_NODE_MULTI_PTR_STR     :: "multi pointer type"
AST_NODE_ARRAY_TYPE_STR    :: "array type"
AST_NODE_DYN_ARRAY_STR     :: "dynamic array type"
AST_NODE_FIXED_DYN_ARR_STR :: "fixed capacity dynamic array type"
AST_NODE_STRUCT_TYPE_STR   :: "struct type"
AST_NODE_UNION_TYPE_STR    :: "union type"
AST_NODE_ENUM_TYPE_STR     :: "enum type"
AST_NODE_BIT_SET_TYPE_STR  :: "bit set type"
AST_NODE_BIT_FIELD_TYPE_STR:: "bit field type"
AST_NODE_MAP_TYPE_STR      :: "map type"
AST_NODE_MATRIX_TYPE_STR   :: "matrix type"

// ============================================================================
// ProcCallingConvention strings
// ============================================================================

PROC_CC_STRING_ODIN        :: "odin"
PROC_CC_STRING_CONTEXTLESS :: "contextless"
PROC_CC_STRING_CDECL       :: "cdecl"
PROC_CC_STRING_STDCALL     :: "stdcall"
PROC_CC_STRING_FASTCALL    :: "fastcall"
PROC_CC_STRING_NONE        :: "none"
PROC_CC_STRING_NAKED       :: "naked"
PROC_CC_STRING_INLINEASM   :: "inlineasm"
PROC_CC_STRING_WIN64       :: "win64"
PROC_CC_STRING_SYSV        :: "sysv"
PROC_CC_STRING_PRESERVE_N  :: "preserve/none"
PROC_CC_STRING_PRESERVE_M  :: "preserve/most"
PROC_CC_STRING_PRESERVE_A  :: "preserve/all"

// ============================================================================
// InlineAsmDialect strings
// ============================================================================

INLINE_ASM_DIALECT_ATT  :: "att"
INLINE_ASM_DIALECT_INTEL :: "intel"

// ============================================================================
// UnionTypeKind strings
// ============================================================================

UNION_TYPE_NORMAL_STR     :: "(normal)"
UNION_TYPE_MAYBE_STR      :: "#maybe"
UNION_TYPE_NO_NIL_STR     :: "#no_nil"
UNION_TYPE_SHARED_NIL_STR :: "#shared_nil"

// ============================================================================
// Enums
// ============================================================================

AddressingMode :: enum u8 {
	Invalid        = 0,
	NoValue        = 1,
	Value          = 2,
	Context        = 3,
	Variable       = 4,
	Constant       = 5,
	Type           = 6,
	Builtin        = 7,
	ProcGroup      = 8,
	MapIndex       = 9,
	OptionalOk     = 10,
	OptionalOkPtr  = 11,
	SoaVariable    = 12,
	SwizzleValue   = 13,
	SwizzleVariable = 14,
}

ParseFileError :: enum {
	None,
	WrongExtension,
	InvalidFile,
	EmptyFile,
	Permission,
	NotFound,
	InvalidToken,
	GeneralError,
	FileTooLarge,
	DirectoryAlreadyExists,
	COUNT,
}

PackageKind :: enum {
	Normal,
	Runtime,
	Init,
	Builtin,
}

AstFileFlag :: enum u32 {
	IsPrivatePkg        = 0,
	IsPrivateFile       = 1,
	IsLazy              = 4,
	NoInstrumentation   = 5,
}

AstFileFlag_Bits :: distinct bit_set[AstFileFlag; u32]

AstDelayQueueKind :: enum {
	Import,
	Expr,
	ForeignBlock,
	COUNT,
}

AstForeignFileKind :: enum {
	Invalid,
	S,
	COUNT,
}

ProcInlining :: enum u8 {
	None      = 0,
	Inline    = 1,
	NoInline  = 2,
}

ProcTailing :: enum u8 {
	None     = 0,
	MustTail = 1,
}

ProcTag :: enum u32 {
	BoundsCheck          = 0,
	NoBoundsCheck        = 1,
	TypeAssert           = 2,
	NoTypeAssert         = 3,
	RequireResults       = 4,
	OptionalOk           = 5,
	OptionalAllocError   = 6,
}

ProcTag_Bits :: distinct bit_set[ProcTag; u32]

ProcCallingConvention :: enum i32 {
	Invalid              = 0,
	Odin                 = 1,
	Contextless          = 2,
	CDecl                = 3,
	StdCall              = 4,
	FastCall             = 5,
	None                 = 6,
	Naked                = 7,
	InlineAsm            = 8,
	Win64                = 9,
	SysV                 = 10,
	PreserveNone         = 11,
	PreserveMost         = 12,
	PreserveAll          = 13,
	MAX,
	ForeignBlockDefault  = -1,
}

StateFlag :: enum u8 {
	BoundsCheck        = 0,
	NoBoundsCheck      = 1,
	TypeAssert         = 2,
	NoTypeAssert       = 3,
	SelectorCallExpr   = 5,
	DirectiveWasFalse  = 6,
	BeenHandled        = 7,
}

StateFlag_Bits :: distinct bit_set[StateFlag; u8]

ViralStateFlag :: enum u8 {
	ContainsDeferredProcedure = 0,
	ContainsOrBreak           = 1,
	ContainsOrReturn          = 2,
}

ViralStateFlag_Bits :: distinct bit_set[ViralStateFlag; u8]

FieldFlag :: enum u32 {
	NONE         = 0,
	Ellipsis     = 0,
	Using        = 1,
	NoAlias      = 2,
	CVarArg      = 3,
	Const        = 5,
	AnyInt       = 6,
	Subtype      = 7,
	ByPtr        = 8,
	NoBroadcast  = 9,
	NoCapture    = 11,
	Tags         = 15,
	Results      = 16,
	Unknown      = 30,
	Invalid      = 31,
}

FieldFlag_Bits :: distinct bit_set[FieldFlag; u32]

// Composite flag sets
FIELD_FLAG_SIGNATURE :: FieldFlag_Bits{
	.Ellipsis, .Using, .NoAlias, .CVarArg,
	.Const, .AnyInt, .ByPtr, .NoBroadcast,
	.NoCapture,
}
FIELD_FLAG_STRUCT :: FieldFlag_Bits{
	.Using, .Subtype, .Tags,
}

StmtAllowFlag :: enum {
	None  = 0,
	In    = 0,
	Label = 1,
}

StmtAllowFlag_Bits :: distinct bit_set[StmtAllowFlag; u8]

InlineAsmDialectKind :: enum u8 {
	Default,
	ATT,
	Intel,
	COUNT,
}

UnionTypeKind :: enum u8 {
	Normal     = 0,
	// index 1 left for Maybe variant (referenced in string table)
	NoNil     = 2,
	SharedNil = 3,
	COUNT,
}

AstKind :: enum u16 {
	Invalid,
	// --- expressions ---
	Ident,
	Implicit,
	Uninit,
	BasicLit,
	BasicDirective,
	Ellipsis,
	ProcGroup,
	ProcLit,
	CompoundLit,
	ExprBegin,      // marker — sentinel, not a real node
	BadExpr,
	TagExpr,
	UnaryExpr,
	BinaryExpr,
	ParenExpr,
	SelectorExpr,
	ImplicitSelectorExpr,
	SelectorCallExpr,
	IndexExpr,
	DerefExpr,
	SliceExpr,
	CallExpr,
	FieldValue,
	EnumFieldValue,
	TernaryIfExpr,
	TernaryWhenExpr,
	OrElseExpr,
	OrReturnExpr,
	OrBranchExpr,
	TypeAssertion,
	TypeCast,
	AutoCast,
	InlineAsmExpr,
	MatrixIndexExpr,
	ExprEnd,        // marker — sentinel
	// --- statements ---
	StmtBegin,      // marker
	BadStmt,
	EmptyStmt,
	ExprStmt,
	AssignStmt,
	ComplexStmtBegin,
	BlockStmt,
	IfStmt,
	WhenStmt,
	ReturnStmt,
	ForStmt,
	RangeStmt,
	UnrollRangeStmt,
	CaseClause,
	SwitchStmt,
	TypeSwitchStmt,
	DeferStmt,
	BranchStmt,
	UsingStmt,
	ComplexStmtEnd,
	StmtEnd,        // marker
	// --- declarations ---
	DeclBegin,      // marker
	BadDecl,
	ForeignBlockDecl,
	Label,
	ValueDecl,
	PackageDecl,
	ImportDecl,
	ForeignImportDecl,
	DeclEnd,        // marker
	// --- misc ---
	Attribute,
	Field,
	BitFieldField,
	FieldList,
	// --- types ---
	TypeBegin,      // marker
	TypeidType,
	HelperType,
	DistinctType,
	PolyType,
	ProcType,
	PointerType,
	RelativeType,
	MultiPointerType,
	ArrayType,
	DynamicArrayType,
	FixedCapacityDynamicArrayType,
	StructType,
	UnionType,
	EnumType,
	BitSetType,
	BitFieldType,
	MapType,
	MatrixType,
	TypeEnd,        // marker
	COUNT,
}

// ============================================================================
// Constant arrays (public runtime strings)
// ============================================================================

addressing_mode_strings := [?]string{
	"Invalid",
	"NoValue",
	"Value",
	"Context",
	"Variable",
	"Constant",
	"Type",
	"Builtin",
	"ProcGroup",
	"MapIndex",
	"OptionalOk",
	"OptionalOkPtr",
	"SoaVariable",
	"SwizzleValue",
	"SwizzleVariable",
}

proc_calling_convention_strings := [ProcCallingConvention.MAX]string{
	"",
	PROC_CC_STRING_ODIN,
	PROC_CC_STRING_CONTEXTLESS,
	PROC_CC_STRING_CDECL,
	PROC_CC_STRING_STDCALL,
	PROC_CC_STRING_FASTCALL,
	PROC_CC_STRING_NONE,
	PROC_CC_STRING_NAKED,
	PROC_CC_STRING_INLINEASM,
	PROC_CC_STRING_WIN64,
	PROC_CC_STRING_SYSV,
	PROC_CC_STRING_PRESERVE_N,
	PROC_CC_STRING_PRESERVE_M,
	PROC_CC_STRING_PRESERVE_A,
}

inline_asm_dialect_strings := [InlineAsmDialectKind.COUNT]string{
	"",
	INLINE_ASM_DIALECT_ATT,
	INLINE_ASM_DIALECT_INTEL,
}

union_type_kind_strings := [UnionTypeKind.COUNT]string{
	UNION_TYPE_NORMAL_STR,
	UNION_TYPE_MAYBE_STR,     // corresponds to the un-named value 1
	UNION_TYPE_NO_NIL_STR,
	UNION_TYPE_SHARED_NIL_STR,
}

ast_strings := [AstKind.COUNT]string{
	AST_NODE_INVALID_STR,
	AST_NODE_IDENT_STR, AST_NODE_IMPLICIT_STR, AST_NODE_UNINIT_STR,
	AST_NODE_BASIC_LIT_STR, AST_NODE_BASIC_DIR_STR, AST_NODE_ELLIPSIS_STR,
	AST_NODE_PROC_GROUP_STR, AST_NODE_PROC_LIT_STR, AST_NODE_COMPOUND_LIT_STR,
	"", // ExprBegin
	AST_NODE_BAD_EXPR_STR, AST_NODE_TAG_EXPR_STR, AST_NODE_UNARY_EXPR_STR,
	AST_NODE_BINARY_EXPR_STR, AST_NODE_PAREN_EXPR_STR, AST_NODE_SELECTOR_EXPR_STR,
	AST_NODE_IMP_SEL_EXPR_STR, AST_NODE_SEL_CALL_EXPR_STR, AST_NODE_INDEX_EXPR_STR,
	AST_NODE_DEREF_EXPR_STR, AST_NODE_SLICE_EXPR_STR, AST_NODE_CALL_EXPR_STR,
	AST_NODE_FIELD_VAL_STR, AST_NODE_ENUM_FIELD_STR,
	AST_NODE_TERN_IF_STR, AST_NODE_TERN_WHEN_STR, AST_NODE_OR_ELSE_STR,
	AST_NODE_OR_RETURN_STR, AST_NODE_OR_BRANCH_STR, AST_NODE_TYPE_ASSERT_STR,
	AST_NODE_TYPE_CAST_STR, AST_NODE_AUTO_CAST_STR, AST_NODE_INLINE_ASM_STR,
	AST_NODE_MATRIX_INDEX_STR,
	"", // ExprEnd
	"", // StmtBegin
	AST_NODE_BAD_STMT_STR, AST_NODE_EMPTY_STMT_STR, AST_NODE_EXPR_STMT_STR,
	AST_NODE_ASSIGN_STMT_STR,
	"", // ComplexStmtBegin
	AST_NODE_BLOCK_STMT_STR, AST_NODE_IF_STMT_STR, AST_NODE_WHEN_STMT_STR,
	AST_NODE_RETURN_STMT_STR, AST_NODE_FOR_STMT_STR, AST_NODE_RANGE_STMT_STR,
	AST_NODE_UNROLL_RANGE_STR, AST_NODE_CASE_CLAUSE_STR, AST_NODE_SWITCH_STMT_STR,
	AST_NODE_TYPE_SWITCH_STR, AST_NODE_DEFER_STMT_STR, AST_NODE_BRANCH_STMT_STR,
	AST_NODE_USING_STMT_STR,
	"", // ComplexStmtEnd
	"", // StmtEnd
	"", // DeclBegin
	AST_NODE_BAD_DECL_STR, AST_NODE_FOREIGN_BLOCK_STR, AST_NODE_LABEL_STR,
	AST_NODE_VALUE_DECL_STR, AST_NODE_PACKAGE_DECL_STR, AST_NODE_IMPORT_DECL_STR,
	AST_NODE_FOREIGN_IMP_STR,
	"", // DeclEnd
	AST_NODE_ATTRIBUTE_STR, AST_NODE_FIELD_STR, AST_NODE_BIT_FIELD_STR,
	AST_NODE_FIELD_LIST_STR,
	"", // TypeBegin
	AST_NODE_TYPEID_STR, AST_NODE_HELPER_TYPE_STR, AST_NODE_DISTINCT_TYPE_STR,
	AST_NODE_POLY_TYPE_STR, AST_NODE_PROC_TYPE_STR, AST_NODE_POINTER_TYPE_STR,
	AST_NODE_RELATIVE_TYPE_STR, AST_NODE_MULTI_PTR_STR, AST_NODE_ARRAY_TYPE_STR,
	AST_NODE_DYN_ARRAY_STR, AST_NODE_FIXED_DYN_ARR_STR,
	AST_NODE_STRUCT_TYPE_STR, AST_NODE_UNION_TYPE_STR, AST_NODE_ENUM_TYPE_STR,
	AST_NODE_BIT_SET_TYPE_STR, AST_NODE_BIT_FIELD_TYPE_STR, AST_NODE_MAP_TYPE_STR,
	AST_NODE_MATRIX_TYPE_STR,
	"", // TypeEnd
}

// ============================================================================
// Struct types — small / leaf types first, then composites
// ============================================================================

TypeAndValue :: struct {
	type:              ^Type,
	mode:              AddressingMode,
	is_lhs:            bool,
	objc_super_target: ^Type,
	value:             ExactValue,
}

CommentGroup :: struct {
	list: []Token,
}

ImportedFile :: struct {
	pkg:   ^AstPackage,
	fi:    FileInfo,
	pos:   TokenPos,
	index: int,
}

AstPackageExportedEntity :: struct {
	identifier: ^Ast,
	entity:     ^Entity,
}

ParseFileErrorNode :: struct {
	next, prev: ^ParseFileErrorNode,
	err:        ParseFileError,
}

// MPMCQueue — simple generic multi-producer multi-consumer queue placeholder
MPMCQueue :: struct(T: typeid) {
	// Placeholder: actual implementation lives in a separate file.
	// Uses Odin's sync primitives internally.
}

// ============================================================================
// Ast variant sub-types (used inside the Ast union)
// ============================================================================

AstExprBegin    :: struct {}
AstExprEnd      :: struct {}
AstStmtBegin    :: struct {}
AstStmtEnd      :: struct {}
AstComplexStmtBegin :: struct {}
AstComplexStmtEnd   :: struct {}
AstDeclBegin    :: struct {}
AstDeclEnd      :: struct {}
AstTypeBegin    :: struct {}
AstTypeEnd      :: struct {}

AstIdent :: struct {
	token:    Token,
	// entity:  ^Entity,           // atomic — handled externally via helper procs
	hash:     u32,
	interned: InternedString,
}
AstImplicit           :: Token
AstUninit             :: Token
AstBasicLit           :: struct { token: Token }
AstBasicDirective     :: struct { token, name: Token }
AstEllipsis           :: struct { token: Token; expr: ^Ast }
AstProcGroup          :: struct { token, open, close: Token; args: []^Ast }
AstProcLit            :: struct {
	type:         ^Ast,
	body:         ^Ast,
	tags:         u64,
	inlining:     ProcInlining,
	tailing:      ProcTailing,
	where_token:  Token,
	where_clauses: []^Ast,
	decl:         ^DeclInfo,
}
AstCompoundLit :: struct {
	type:      ^Ast,
	elems:     []^Ast,
	open:      Token,
	close:     Token,
	max_count: i64,
	tag:       ^Ast,
}
AstBadExpr     :: struct { begin, end: Token }
AstTagExpr     :: struct { token, name: Token; expr: ^Ast }
AstUnaryExpr   :: struct { op: Token; expr: ^Ast }
AstBinaryExpr  :: struct { op: Token; left, right: ^Ast }
AstParenExpr   :: struct { expr: ^Ast; open, close: Token }
AstSelectorExpr :: struct {
	token:          Token,
	expr:           ^Ast,
	selector:       ^Ast,
	swizzle_count:  u8,
	swizzle_indices: u8,
	is_bit_field:   bool,
}
AstImplicitSelectorExpr :: struct { token: Token; selector: ^Ast }
AstSelectorCallExpr :: struct {
	token:         Token,
	expr:          ^Ast,
	call:          ^Ast,
	modified_call: bool,
}
AstIndexExpr      :: struct { expr, index: ^Ast; open, close: Token }
AstDerefExpr      :: struct { expr: ^Ast; op: Token }
AstSliceExpr      :: struct {
	expr:     ^Ast,
	open:     Token,
	close:    Token,
	interval: Token,
	low:      ^Ast,
	high:     ^Ast,
}
AstCallExpr :: struct {
	proc:               ^Ast,
	args:               []^Ast,
	open:               Token,
	close:              Token,
	ellipsis:           Token,
	inlining:           ProcInlining,
	tailing:            ProcTailing,
	optional_ok_one:    bool,
	was_selector:       bool,
	split_args:         ^AstSplitArgs,
	entity_procedure_of: ^Entity,
}
AstFieldValue     :: struct { eq: Token; field, value: ^Ast }
AstEnumFieldValue :: struct {
	name:    ^Ast,
	value:   ^Ast,
	docs:    ^CommentGroup,
	comment: ^CommentGroup,
}
AstTernaryIfExpr  :: struct { x, cond, y: ^Ast }
AstTernaryWhenExpr :: struct { x, cond, y: ^Ast }
AstOrElseExpr     :: struct { x: ^Ast; token: Token; y: ^Ast }
AstOrReturnExpr   :: struct { expr: ^Ast; token: Token }
AstOrBranchExpr   :: struct { expr: ^Ast; token: Token; label: ^Ast }
AstTypeAssertion  :: struct {
	expr:       ^Ast,
	dot:        Token,
	type:       ^Ast,
	type_hint:  ^Type,
	ignores:    [2]bool,
}
AstTypeCast  :: struct { token: Token; type, expr: ^Ast }
AstAutoCast  :: struct { token: Token; expr: ^Ast }
AstInlineAsmExpr :: struct {
	token:            Token,
	open:             Token,
	close:            Token,
	param_types:      []^Ast,
	return_type:      ^Ast,
	asm_string:       ^Ast,
	constraints_string: ^Ast,
	has_side_effects: bool,
	is_align_stack:   bool,
	dialect:          InlineAsmDialectKind,
}
AstMatrixIndexExpr :: struct {
	expr:         ^Ast,
	row_index:    ^Ast,
	column_index: ^Ast,
	open:         Token,
	close:        Token,
}
AstBadStmt    :: struct { begin, end: Token }
AstEmptyStmt  :: struct { token: Token }
AstExprStmt   :: struct { expr: ^Ast }
AstAssignStmt :: struct { op: Token; lhs, rhs: []^Ast }
AstBlockStmt  :: struct {
	scope: ^Scope,
	stmts: []^Ast,
	label: ^Ast,
	open:  Token,
	close: Token,
}
AstIfStmt :: struct {
	scope:     ^Scope,
	token:     Token,
	label:     ^Ast,
	init:      ^Ast,
	cond:      ^Ast,
	body:      ^Ast,
	else_stmt: ^Ast,
}
AstWhenStmt :: struct {
	token:             Token,
	cond:              ^Ast,
	body:              ^Ast,
	else_stmt:         ^Ast,
	is_cond_determined: bool,
	determined_cond:   bool,
}
AstReturnStmt :: struct { token: Token; results: []^Ast }
AstForStmt :: struct {
	scope: ^Scope,
	token: Token,
	label: ^Ast,
	init:  ^Ast,
	cond:  ^Ast,
	post:  ^Ast,
	body:  ^Ast,
}
AstRangeStmt :: struct {
	scope:    ^Scope,
	token:    Token,
	label:    ^Ast,
	init:     ^Ast,
	vals:     []^Ast,
	in_token: Token,
	expr:     ^Ast,
	body:     ^Ast,
	reverse:  bool,
}
AstUnrollRangeStmt :: struct {
	scope:        ^Scope,
	unroll_token: Token,
	init:         ^Ast,
	args:         []^Ast,
	for_token:    Token,
	val0:         ^Ast,
	val1:         ^Ast,
	in_token:     Token,
	expr:         ^Ast,
	body:         ^Ast,
}
AstCaseClause :: struct {
	scope:           ^Scope,
	token:           Token,
	list:            []^Ast,
	stmts:           []^Ast,
	implicit_entity: ^Entity,
}
AstSwitchStmt :: struct {
	scope:   ^Scope,
	token:   Token,
	label:   ^Ast,
	init:    ^Ast,
	tag:     ^Ast,
	body:    ^Ast,
	partial: bool,
}
AstTypeSwitchStmt :: struct {
	scope:   ^Scope,
	token:   Token,
	label:   ^Ast,
	tag:     ^Ast,
	body:    ^Ast,
	partial: bool,
}
AstDeferStmt  :: struct { token: Token; stmt: ^Ast }
AstBranchStmt :: struct { token: Token; label: ^Ast }
AstUsingStmt  :: struct { token: Token; list: []^Ast }
AstBadDecl    :: struct { begin, end: Token }
AstForeignBlockDecl :: struct {
	token:           Token,
	foreign_library: ^Ast,
	body:            ^Ast,
	attributes:      [dynamic]^Ast,
	docs:            ^CommentGroup,
}
AstLabel :: struct { token: Token; name: ^Ast }
AstValueDecl :: struct {
	names:      []^Ast,
	type:       ^Ast,
	values:     []^Ast,
	attributes: [dynamic]^Ast,
	docs:       ^CommentGroup,
	comment:    ^CommentGroup,
	is_using:   bool,
	is_mutable: bool,
}
AstPackageDecl :: struct {
	token:   Token,
	name:    Token,
	docs:    ^CommentGroup,
	comment: ^CommentGroup,
}
AstImportDecl :: struct {
	package:    ^AstPackage,
	token:      Token,
	relpath:    Token,
	fullpath:   string,
	import_name: Token,
	attributes: [dynamic]^Ast,
	docs:       ^CommentGroup,
	comment:    ^CommentGroup,
}
AstForeignImportDecl :: struct {
	token:              Token,
	filepaths:          []^Ast,
	multiple_filepaths: bool,
	library_name:       Token,
	collection_name:    string,
	fullpaths:          []string,
	attributes:         [dynamic]^Ast,
	docs:               ^CommentGroup,
	comment:            ^CommentGroup,
}
AstAttribute :: struct {
	token: Token,
	elems: []^Ast,
	open:  Token,
	close: Token,
}
AstField :: struct {
	names:         []^Ast,
	type:          ^Ast,
	default_value: ^Ast,
	tag:           Token,
	flags:         u32,
	docs:          ^CommentGroup,
	comment:       ^CommentGroup,
}
AstBitFieldField :: struct {
	name:     ^Ast,
	type:     ^Ast,
	bit_size: ^Ast,
	tag:      Token,
	docs:     ^CommentGroup,
	comment:  ^CommentGroup,
}
AstFieldList :: struct { token: Token; list: []^Ast }
AstTypeidType :: struct { token: Token; specialization: ^Ast }
AstHelperType :: struct { token: Token; type: ^Ast }
AstDistinctType :: struct { token: Token; type: ^Ast }
AstPolyType :: struct {
	token:          Token,
	type:           ^Ast,
	specialization: ^Ast,
}
AstProcType :: struct {
	scope:              ^Scope,
	token:              Token,
	params:             ^Ast,
	results:            ^Ast,
	tags:               u64,
	calling_convention: ProcCallingConvention,
	generic:            bool,
	diverging:          bool,
}
AstPointerType :: struct { token: Token; type: ^Ast; tag: ^Ast }
AstRelativeType :: struct { tag: ^Ast; type: ^Ast }
AstMultiPointerType :: struct { token: Token; type: ^Ast }
AstArrayType :: struct { token: Token; count: ^Ast; elem: ^Ast; tag: ^Ast }
AstDynamicArrayType :: struct { token: Token; elem: ^Ast; tag: ^Ast }
AstFixedCapacityDynamicArrayType :: struct {
	token:    Token,
	capacity: ^Ast,
	elem:     ^Ast,
	tag:      ^Ast,
}
AstStructType :: struct {
	scope:              ^Scope,
	token:              Token,
	fields:             []^Ast,
	field_count:        int,
	polymorphic_params: ^Ast,
	align:              ^Ast,
	min_field_align:    ^Ast,
	max_field_align:    ^Ast,
	where_token:        Token,
	where_clauses:      []^Ast,
	is_packed:          bool,
	is_raw_union:       bool,
	is_no_copy:         bool,
	is_all_or_none:     bool,
	is_simple:          bool,
}
AstUnionType :: struct {
	scope:              ^Scope,
	token:              Token,
	variants:           []^Ast,
	polymorphic_params: ^Ast,
	align:              ^Ast,
	kind:               UnionTypeKind,
	where_token:        Token,
	where_clauses:      []^Ast,
}
AstEnumType :: struct {
	scope:     ^Scope,
	token:     Token,
	base_type: ^Ast,
	fields:    []^Ast,
	is_using:  bool,
}
AstBitSetType :: struct { token: Token; elem: ^Ast; underlying: ^Ast }
AstBitFieldType :: struct {
	scope:        ^Scope,
	token:        Token,
	backing_type: ^Ast,
	open:         Token,
	fields:       []^Ast,
	close:        Token,
}
AstMapType :: struct {
	token: Token,
	count: ^Ast,
	key:   ^Ast,
	value: ^Ast,
}
AstMatrixType :: struct {
	token:        Token,
	row_count:    ^Ast,
	column_count: ^Ast,
	elem:         ^Ast,
	is_row_major: bool,
}

// ============================================================================
// AstSplitArgs — helper for CallExpr
// ============================================================================

AstSplitArgs :: struct {
	positional: []^Ast,
	named:      []^Ast,
}

// ============================================================================
// AstCommonStuff / Ast — the main tagged-union node
// ============================================================================

Ast :: struct {
	kind:              AstKind,
	state_flags:       u8,
	viral_state_flags: u8,    // NOTE — originally std::atomic<u8>; use sync procs for atomic ops
	file_id:           i32,
	tav:               TypeAndValue,

	// Discriminated union: active member is determined by `kind`
	v: union {
	    Ident:                 AstIdent,
	    Implicit:              AstImplicit,
	    Uninit:                AstUninit,
	    BasicLit:              AstBasicLit,
	    BasicDirective:        AstBasicDirective,
	    Ellipsis:              AstEllipsis,
	    ProcGroup:             AstProcGroup,
	    ProcLit:               AstProcLit,
	    CompoundLit:           AstCompoundLit,
	    _ExprBegin:            AstExprBegin,
	    BadExpr:               AstBadExpr,
	    TagExpr:               AstTagExpr,
	    UnaryExpr:             AstUnaryExpr,
	    BinaryExpr:            AstBinaryExpr,
	    ParenExpr:             AstParenExpr,
	    SelectorExpr:          AstSelectorExpr,
	    ImplicitSelectorExpr:  AstImplicitSelectorExpr,
	    SelectorCallExpr:      AstSelectorCallExpr,
	    IndexExpr:             AstIndexExpr,
	    DerefExpr:             AstDerefExpr,
	    SliceExpr:             AstSliceExpr,
	    CallExpr:              AstCallExpr,
	    FieldValue:            AstFieldValue,
	    EnumFieldValue:        AstEnumFieldValue,
	    TernaryIfExpr:         AstTernaryIfExpr,
	    TernaryWhenExpr:       AstTernaryWhenExpr,
	    OrElseExpr:            AstOrElseExpr,
	    OrReturnExpr:          AstOrReturnExpr,
	    OrBranchExpr:          AstOrBranchExpr,
	    TypeAssertion:         AstTypeAssertion,
	    TypeCast:              AstTypeCast,
	    AutoCast:              AstAutoCast,
	    InlineAsmExpr:         AstInlineAsmExpr,
	    MatrixIndexExpr:       AstMatrixIndexExpr,
	    _ExprEnd:              AstExprEnd,
	    _StmtBegin:            AstStmtBegin,
	    BadStmt:               AstBadStmt,
	    EmptyStmt:             AstEmptyStmt,
	    ExprStmt:              AstExprStmt,
	    AssignStmt:            AstAssignStmt,
	    _ComplexStmtBegin:     AstComplexStmtBegin,
	    BlockStmt:             AstBlockStmt,
	    IfStmt:                AstIfStmt,
	    WhenStmt:              AstWhenStmt,
	    ReturnStmt:            AstReturnStmt,
	    ForStmt:               AstForStmt,
	    RangeStmt:             AstRangeStmt,
	    UnrollRangeStmt:       AstUnrollRangeStmt,
	    CaseClause:            AstCaseClause,
	    SwitchStmt:            AstSwitchStmt,
	    TypeSwitchStmt:        AstTypeSwitchStmt,
	    DeferStmt:             AstDeferStmt,
	    BranchStmt:            AstBranchStmt,
	    UsingStmt:             AstUsingStmt,
	    _ComplexStmtEnd:       AstComplexStmtEnd,
	    _StmtEnd:              AstStmtEnd,
	    _DeclBegin:            AstDeclBegin,
	    BadDecl:               AstBadDecl,
	    ForeignBlockDecl:      AstForeignBlockDecl,
	    Label:                 AstLabel,
	    ValueDecl:             AstValueDecl,
	    PackageDecl:           AstPackageDecl,
	    ImportDecl:            AstImportDecl,
	    ForeignImportDecl:     AstForeignImportDecl,
	    _DeclEnd:              AstDeclEnd,
	    Attribute:             AstAttribute,
	    Field:                 AstField,
	    BitFieldField:         AstBitFieldField,
	    FieldList:             AstFieldList,
	    _TypeBegin:            AstTypeBegin,
	    TypeidType:            AstTypeidType,
	    HelperType:            AstHelperType,
	    DistinctType:          AstDistinctType,
	    PolyType:              AstPolyType,
	    ProcType:              AstProcType,
	    PointerType:           AstPointerType,
	    RelativeType:          AstRelativeType,
	    MultiPointerType:      AstMultiPointerType,
	    ArrayType:             AstArrayType,
	    DynamicArrayType:      AstDynamicArrayType,
	    FixedCapacityDynamicArrayType: AstFixedCapacityDynamicArrayType,
	    StructType:            AstStructType,
	    UnionType:             AstUnionType,
	    EnumType:              AstEnumType,
	    BitSetType:            AstBitSetType,
	    BitFieldType:          AstBitFieldType,
	    MapType:               AstMapType,
	    MatrixType:            AstMatrixType,
	    _TypeEnd:              AstTypeEnd,
	},
}

// ast_variant_sizes — rough sizes for each AstKind (used by pool allocators)
// NOTE: uses `size_of` on each variant type for the Odin equivalent.
ast_variant_sizes := [?]int{
	0,
	size_of(AstIdent),
	size_of(AstImplicit),
	size_of(AstUninit),
	size_of(AstBasicLit),
	size_of(AstBasicDirective),
	size_of(AstEllipsis),
	size_of(AstProcGroup),
	size_of(AstProcLit),
	size_of(AstCompoundLit),
	size_of(AstExprBegin),
	size_of(AstBadExpr),
	size_of(AstTagExpr),
	size_of(AstUnaryExpr),
	size_of(AstBinaryExpr),
	size_of(AstParenExpr),
	size_of(AstSelectorExpr),
	size_of(AstImplicitSelectorExpr),
	size_of(AstSelectorCallExpr),
	size_of(AstIndexExpr),
	size_of(AstDerefExpr),
	size_of(AstSliceExpr),
	size_of(AstCallExpr),
	size_of(AstFieldValue),
	size_of(AstEnumFieldValue),
	size_of(AstTernaryIfExpr),
	size_of(AstTernaryWhenExpr),
	size_of(AstOrElseExpr),
	size_of(AstOrReturnExpr),
	size_of(AstOrBranchExpr),
	size_of(AstTypeAssertion),
	size_of(AstTypeCast),
	size_of(AstAutoCast),
	size_of(AstInlineAsmExpr),
	size_of(AstMatrixIndexExpr),
	size_of(AstExprEnd),
	size_of(AstStmtBegin),
	size_of(AstBadStmt),
	size_of(AstEmptyStmt),
	size_of(AstExprStmt),
	size_of(AstAssignStmt),
	size_of(AstComplexStmtBegin),
	size_of(AstBlockStmt),
	size_of(AstIfStmt),
	size_of(AstWhenStmt),
	size_of(AstReturnStmt),
	size_of(AstForStmt),
	size_of(AstRangeStmt),
	size_of(AstUnrollRangeStmt),
	size_of(AstCaseClause),
	size_of(AstSwitchStmt),
	size_of(AstTypeSwitchStmt),
	size_of(AstDeferStmt),
	size_of(AstBranchStmt),
	size_of(AstUsingStmt),
	size_of(AstComplexStmtEnd),
	size_of(AstStmtEnd),
	size_of(AstDeclBegin),
	size_of(AstBadDecl),
	size_of(AstForeignBlockDecl),
	size_of(AstLabel),
	size_of(AstValueDecl),
	size_of(AstPackageDecl),
	size_of(AstImportDecl),
	size_of(AstForeignImportDecl),
	size_of(AstDeclEnd),
	size_of(AstAttribute),
	size_of(AstField),
	size_of(AstBitFieldField),
	size_of(AstFieldList),
	size_of(AstTypeBegin),
	size_of(AstTypeidType),
	size_of(AstHelperType),
	size_of(AstDistinctType),
	size_of(AstPolyType),
	size_of(AstProcType),
	size_of(AstPointerType),
	size_of(AstRelativeType),
	size_of(AstMultiPointerType),
	size_of(AstArrayType),
	size_of(AstDynamicArrayType),
	size_of(AstFixedCapacityDynamicArrayType),
	size_of(AstStructType),
	size_of(AstUnionType),
	size_of(AstEnumType),
	size_of(AstBitSetType),
	size_of(AstBitFieldType),
	size_of(AstMapType),
	size_of(AstMatrixType),
	size_of(AstTypeEnd),
}

// ============================================================================
// AstFile — per-file parser state
// ============================================================================

AstFile :: struct {
	id:                        i32,
	flags:                     u32,
	pkg:                       ^AstPackage,
	scope:                     ^Scope,
	pkg_decl:                  ^Ast,
	fullpath:                  string,
	filename:                  string,
	directory:                 string,
	tokenizer:                 Tokenizer,
	tokens:                    [dynamic]Token,
	curr_token_index:          int,
	prev_token_index:          int,
	curr_token:                Token,
	prev_token:                Token,
	package_token:             Token,
	package_name:              string,
	vet_flags:                 u64,
	feature_flags:             u64,
	vet_flags_set:             bool,
	feature_flags_set:         bool,
	expr_level:                int,
	allow_newline:             bool,
	allow_range:               bool,
	allow_in_expr:             bool,
	in_foreign_block:          bool,
	allow_type:                bool,
	in_when_statement:         bool,
	total_file_decl_count:     int,
	delayed_decl_count:        int,
	decls:                     []^Ast,
	imports:                   [dynamic]^Ast,
	directive_count:           int,
	curr_proc:                 ^Ast,
	error_count:               int,
	last_error:                ParseFileError,
	time_to_tokenize:          f64,
	time_to_parse:             f64,
	lead_comment:              ^CommentGroup,
	line_comment:              ^CommentGroup,
	docs:                      ^CommentGroup,
	comments:                  [dynamic]^CommentGroup,
	delayed_decls_queues:      [AstDelayQueueKind.COUNT][dynamic]^Ast,
	// NOTE — originally std::atomic<isize>; needs atomic ops in threaded mode
	seen_load_directive_count: int,
	fix_count:                 int,
	fix_prev_pos:              TokenPos,
	llvm_metadata:             LLVMOpaqueMetadata,
	llvm_metadata_scope:       LLVMOpaqueMetadata,
}

// ============================================================================
// AstForeignFile
// ============================================================================

AstForeignFile :: struct {
	kind:   AstForeignFileKind,
	source: string,
}

// ============================================================================
// AstPackage
// ============================================================================

AstPackage :: struct {
	kind:                  PackageKind,
	id:                    int,
	name:                  string,
	fullpath:              string,
	files:                 [dynamic]^AstFile,
	foreign_files:         [dynamic]AstForeignFile,
	is_single_file:        bool,
	order:                 int,
	files_mutex:           sync.Mutex,
	foreign_files_mutex:   sync.Mutex,
	type_and_value_mutex:  sync.Mutex,
	name_mutex:            sync.Mutex,
	exported_entity_queue: MPMCQueue(AstPackageExportedEntity),
	scope:                 ^Scope,
	decl_info:             ^DeclInfo,
	is_extra:              bool,
}

// ============================================================================
// Parser — top-level parser state (shared across worker threads)
// ============================================================================

// StringSet — Odin equivalent of a hash-set of strings
StringSet :: map[string]struct{}

Parser :: struct {
	init_fullpath:                 string,
	imported_files:                StringSet,
	imported_files_mutex:          sync.Mutex,
	packages:                      [dynamic]^AstPackage,
	packages_mutex:                sync.Mutex,
	// NOTE — originally std::atomic<isize>; use sync atomics for cross-thread access
	file_to_process_count:         int,
	total_token_count:             int,
	total_line_count:              int,
	total_seen_load_directive_count: int,
	file_decl_mutex:               sync.Mutex,
	file_error_mutex:              sync.Mutex,
	file_error_head:               ^ParseFileErrorNode,
	file_error_tail:               ^ParseFileErrorNode,
}

// ============================================================================
// Worker data structs
// ============================================================================

ParserWorkerData :: struct {
	parser:        ^Parser,
	imported_file: ImportedFile,
}

ForeignFileWorkerData :: struct {
	parser:        ^Parser,
	imported_file: ImportedFile,
	foreign_kind:  AstForeignFileKind,
}

// ============================================================================
// Helper procedures — range checks on AstKind
// ============================================================================

is_ast_expr :: proc(node: ^Ast) -> bool {
	k := int(node.kind)
	return int(AstKind.ExprBegin) < k && k < int(AstKind.ExprEnd)
}

is_ast_stmt :: proc(node: ^Ast) -> bool {
	k := int(node.kind)
	return int(AstKind.StmtBegin) < k && k < int(AstKind.StmtEnd)
}

is_ast_complex_stmt :: proc(node: ^Ast) -> bool {
	k := int(node.kind)
	return int(AstKind.ComplexStmtBegin) < k && k < int(AstKind.ComplexStmtEnd)
}

is_ast_decl :: proc(node: ^Ast) -> bool {
	k := int(node.kind)
	return int(AstKind.DeclBegin) < k && k < int(AstKind.DeclEnd)
}

is_ast_type :: proc(node: ^Ast) -> bool {
	k := int(node.kind)
	return int(AstKind.TypeBegin) < k && k < int(AstKind.TypeEnd)
}

is_ast_when_stmt :: proc(node: ^Ast) -> bool {
	return node != nil && node.kind == .WhenStmt
}

// ============================================================================
// default_calling_convention
// ============================================================================

default_calling_convention :: proc() -> ProcCallingConvention {
	return .Odin
}

// ============================================================================
// ast_allocator — returns the allocator to use for AST nodes in a given file
// ============================================================================

ast_allocator :: proc(f: ^AstFile, allocator := context.allocator) -> mem.Allocator {
	_ = f
	return allocator
}

// ============================================================================
// Stub / declaration-only procedures — defined in other Odin files
// ============================================================================

// alloc_ast_node — allocate a new Ast node of the given kind in the file's arena
// (implemented in the allocator / AST construction module)
alloc_ast_node :: proc(f: ^AstFile, kind: AstKind, allocator := context.allocator) -> ^Ast {
	_ = f
	_ = kind
	_ = allocator
	// TODO: implementation lives in the AST allocator module
	return nil
}

// expr_to_string — convert an expression AST to a debug string
// (implemented in the pretty-printer / debug module)
expr_to_string :: proc(expression: ^Ast, allocator := context.allocator) -> string {
	_ = expression
	_ = allocator
	// TODO: implementation lives in the expression stringifier module
	return ""
}

// allow_field_separator — check whether a field separator (comma / newline) is allowed
// (implemented in the parser module)
allow_field_separator :: proc(f: ^AstFile) -> bool {
	_ = f
	// TODO: implementation lives in the parser logic module
	return false
}

// parse_enforce_tabs — verify that the current token uses tabs for indentation
// (implemented in the parser-style enforcement module)
parse_enforce_tabs :: proc(f: ^AstFile) {
	_ = f
	// TODO: implementation lives in the parser style-check module
}

// ============================================================================
// thread_safe_file helpers — delegated to the rewritten AstFile module
// ============================================================================

// These were __forceinline in C++ and delegated to global arrays / thread-safe
// accessors that are defined in other files.  Stub them here; the real
// implementations are in the rewritten module.

@(require_results)
thread_safe_file :: proc(node: ^Ast) -> ^AstFile {
	// TODO: delegates to thread_safe_get_ast_file_from_id(node.file_id)
	_ = node
	return nil
}