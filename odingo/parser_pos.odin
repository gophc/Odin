package odingo

import "core:mem"
import "core:strings"
import "core:unicode/utf8"

// --- Stub types (forward declarations) ---

Token :: struct {
	pos:    TokenPos,
	string: string,
	kind:   TokenKind,
}

TokenPos :: struct {
	offset:  int,
	line:    i32,
	column:  i32,
	file_id: u64,
}

TokenKind :: enum i32 {
	Invalid = 0,
}

AstKind :: enum i32 {
	Invalid = 0,
	Ident,
	Implicit,
	Uninit,
	BasicLit,
	BasicDirective,
	ProcGroup,
	ProcLit,
	CompoundLit,
	TagExpr,
	BadExpr,
	UnaryExpr,
	BinaryExpr,
	ParenExpr,
	CallExpr,
	SelectorExpr,
	SelectorCallExpr,
	ImplicitSelectorExpr,
	IndexExpr,
	MatrixIndexExpr,
	SliceExpr,
	Ellipsis,
	FieldValue,
	EnumFieldValue,
	DerefExpr,
	TernaryIfExpr,
	TernaryWhenExpr,
	OrElseExpr,
	OrReturnExpr,
	OrBranchExpr,
	TypeAssertion,
	TypeCast,
	AutoCast,
	InlineAsmExpr,
	BadStmt,
	EmptyStmt,
	ExprStmt,
	AssignStmt,
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
	BadDecl,
	Label,
	ValueDecl,
	PackageDecl,
	ImportDecl,
	ForeignImportDecl,
	ForeignBlockDecl,
	Attribute,
	Field,
	FieldList,
	TypeidType,
	HelperType,
	DistinctType,
	PolyType,
	ProcType,
	RelativeType,
	PointerType,
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
}

Ast :: struct {
	kind:    AstKind,
	file_id: u64,
	using _: Ast_Union,
}

Ast_Union :: struct #raw_union {
	Ident:                        Ast_Ident,
	Implicit:                     Token,
	Uninit:                       Token,
	BasicLit:                     Ast_BasicLit,
	BasicDirective:               Ast_BasicDirective,
	ProcGroup:                    Ast_ProcGroup,
	ProcLit:                      Ast_ProcLit,
	CompoundLit:                  Ast_CompoundLit,
	TagExpr:                      Ast_TagExpr,
	BadExpr:                      Ast_BadExpr,
	UnaryExpr:                    Ast_UnaryExpr,
	BinaryExpr:                   Ast_BinaryExpr,
	ParenExpr:                    Ast_ParenExpr,
	CallExpr:                     Ast_CallExpr,
	SelectorExpr:                 Ast_SelectorExpr,
	SelectorCallExpr:             Ast_SelectorCallExpr,
	ImplicitSelectorExpr:         Ast_ImplicitSelectorExpr,
	IndexExpr:                    Ast_IndexExpr,
	MatrixIndexExpr:              Ast_MatrixIndexExpr,
	SliceExpr:                    Ast_SliceExpr,
	Ellipsis:                     Ast_Ellipsis,
	FieldValue:                   Ast_FieldValue,
	EnumFieldValue:               Ast_EnumFieldValue,
	DerefExpr:                    Ast_DerefExpr,
	TernaryIfExpr:                Ast_TernaryIfExpr,
	TernaryWhenExpr:              Ast_TernaryWhenExpr,
	OrElseExpr:                   Ast_OrElseExpr,
	OrReturnExpr:                 Ast_OrReturnExpr,
	OrBranchExpr:                 Ast_OrBranchExpr,
	TypeAssertion:                Ast_TypeAssertion,
	TypeCast:                     Ast_TypeCast,
	AutoCast:                     Ast_AutoCast,
	InlineAsmExpr:                Ast_InlineAsmExpr,
	BadStmt:                      Ast_BadStmt,
	EmptyStmt:                    Ast_EmptyStmt,
	ExprStmt:                     Ast_ExprStmt,
	AssignStmt:                   Ast_AssignStmt,
	BlockStmt:                    Ast_BlockStmt,
	IfStmt:                       Ast_IfStmt,
	WhenStmt:                     Ast_WhenStmt,
	ReturnStmt:                   Ast_ReturnStmt,
	ForStmt:                      Ast_ForStmt,
	RangeStmt:                    Ast_RangeStmt,
	UnrollRangeStmt:              Ast_UnrollRangeStmt,
	CaseClause:                   Ast_CaseClause,
	SwitchStmt:                   Ast_SwitchStmt,
	TypeSwitchStmt:               Ast_TypeSwitchStmt,
	DeferStmt:                    Ast_DeferStmt,
	BranchStmt:                   Ast_BranchStmt,
	UsingStmt:                    Ast_UsingStmt,
	BadDecl:                      Ast_BadDecl,
	Label:                        Ast_Label,
	ValueDecl:                    Ast_ValueDecl,
	PackageDecl:                  Ast_PackageDecl,
	ImportDecl:                   Ast_ImportDecl,
	ForeignImportDecl:            Ast_ForeignImportDecl,
	ForeignBlockDecl:             Ast_ForeignBlockDecl,
	Attribute:                    Ast_Attribute,
	Field:                        Ast_Field,
	FieldList:                    Ast_FieldList,
	TypeidType:                   Ast_TypeidType,
	HelperType:                   Ast_HelperType,
	DistinctType:                 Ast_DistinctType,
	PolyType:                     Ast_PolyType,
	ProcType:                     Ast_ProcType,
	RelativeType:                 Ast_RelativeType,
	PointerType:                  Ast_PointerType,
	MultiPointerType:             Ast_MultiPointerType,
	ArrayType:                    Ast_ArrayType,
	DynamicArrayType:             Ast_DynamicArrayType,
	FixedCapacityDynamicArrayType: Ast_FixedCapacityDynamicArrayType,
	StructType:                   Ast_StructType,
	UnionType:                    Ast_UnionType,
	EnumType:                     Ast_EnumType,
	BitSetType:                   Ast_BitSetType,
	BitFieldType:                 Ast_BitFieldType,
	MapType:                      Ast_MapType,
	MatrixType:                   Ast_MatrixType,
}

// --- Node sub-types (stubs) ---

Ast_Ident :: struct { token: Token }
Ast_BasicLit :: struct { token: Token }
Ast_BasicDirective :: struct { token: Token }
Ast_ProcGroup :: struct { token, close: Token }
Ast_ProcLit :: struct { type: ^Ast, body: ^Ast }
Ast_CompoundLit :: struct { type: ^Ast, open, close: Token }
Ast_TagExpr :: struct { token, name: Token, expr: ^Ast }
Ast_BadExpr :: struct { begin, end: Token }
Ast_UnaryExpr :: struct { op: Token, expr: ^Ast }
Ast_BinaryExpr :: struct { left, right: ^Ast }
Ast_ParenExpr :: struct { open, close: Token }
Ast_CallExpr :: struct { proc: ^Ast, close: Token }
Ast_SelectorExpr :: struct { expr, selector: ^Ast, token: Token }
Ast_SelectorCallExpr :: struct { expr: ^Ast, call: ^Ast, token: Token }
Ast_ImplicitSelectorExpr :: struct { selector: ^Ast, token: Token }
Ast_IndexExpr :: struct { expr: ^Ast, close: Token }
Ast_MatrixIndexExpr :: struct { expr: ^Ast, close: Token }
Ast_SliceExpr :: struct { expr: ^Ast, close: Token }
Ast_Ellipsis :: struct { token: Token, expr: ^Ast }
Ast_FieldValue :: struct { field, value: ^Ast, eq: Token }
Ast_EnumFieldValue :: struct { name, value: ^Ast }
Ast_DerefExpr :: struct { op: Token }
Ast_TernaryIfExpr :: struct { x, y: ^Ast }
Ast_TernaryWhenExpr :: struct { x, y: ^Ast }
Ast_OrElseExpr :: struct { x, y: ^Ast }
Ast_OrReturnExpr :: struct { expr: ^Ast, token: Token }
Ast_OrBranchExpr :: struct { expr, label: ^Ast, token: Token }
Ast_TypeAssertion :: struct { expr, type: ^Ast }
Ast_TypeCast :: struct { token: Token, expr: ^Ast }
Ast_AutoCast :: struct { token: Token, expr: ^Ast }
Ast_InlineAsmExpr :: struct { token, close: Token }
Ast_BadStmt :: struct { begin, end: Token }
Ast_EmptyStmt :: struct { token: Token }
Ast_ExprStmt :: struct { expr: ^Ast }
Ast_AssignStmt :: struct { op: Token, rhs: [dynamic]^Ast }
Ast_BlockStmt :: struct { open, close: Token }
Ast_IfStmt :: struct { token: Token, body, else_stmt: ^Ast }
Ast_WhenStmt :: struct { token: Token, body, else_stmt: ^Ast }
Ast_ReturnStmt :: struct { token: Token, results: [dynamic]^Ast }
Ast_ForStmt :: struct { token: Token, body: ^Ast }
Ast_RangeStmt :: struct { token: Token, body: ^Ast }
Ast_UnrollRangeStmt :: struct { unroll_token: Token, body: ^Ast }
Ast_CaseClause :: struct { token: Token, list, stmts: [dynamic]^Ast }
Ast_SwitchStmt :: struct { token: Token, body: ^Ast }
Ast_TypeSwitchStmt :: struct { token: Token, body: ^Ast }
Ast_DeferStmt :: struct { token: Token, stmt: ^Ast }
Ast_BranchStmt :: struct { token: Token, label: ^Ast }
Ast_UsingStmt :: struct { token: Token, list: [dynamic]^Ast }
Ast_BadDecl :: struct { begin, end: Token }
Ast_Label :: struct { token: Token, name: ^Ast }
Ast_ValueDecl :: struct { names, values: [dynamic]^Ast, type: ^Ast }
Ast_PackageDecl :: struct { token, name: Token }
Ast_ImportDecl :: struct { token, relpath: Token }
Ast_ForeignImportDecl :: struct { token, library_name: Token, filepaths: [dynamic]^Ast }
Ast_ForeignBlockDecl :: struct { token: Token, body: ^Ast }
Ast_Attribute :: struct { token, open, close: Token, elems: [dynamic]^Ast }
Ast_Field :: struct { names: [dynamic]^Ast, type, default_value: ^Ast, tag: Token }
Ast_FieldList :: struct { token: Token, list: [dynamic]^Ast }
Ast_TypeidType :: struct { token: Token, specialization: ^Ast }
Ast_HelperType :: struct { token: Token, type: ^Ast }
Ast_DistinctType :: struct { token: Token, type: ^Ast }
Ast_PolyType :: struct { token: Token, type, specialization: ^Ast }
Ast_ProcType :: struct { token: Token, params, results: ^Ast }
Ast_RelativeType :: struct { tag, type: ^Ast }
Ast_PointerType :: struct { token: Token, type: ^Ast }
Ast_MultiPointerType :: struct { token: Token, type: ^Ast }
Ast_ArrayType :: struct { token: Token, elem: ^Ast }
Ast_DynamicArrayType :: struct { token: Token, elem: ^Ast }
Ast_FixedCapacityDynamicArrayType :: struct { token: Token, elem: ^Ast }
Ast_StructType :: struct { token: Token, fields: [dynamic]^Ast }
Ast_UnionType :: struct { token: Token, variants: [dynamic]^Ast }
Ast_EnumType :: struct { token: Token, base_type: ^Ast, fields: [dynamic]^Ast }
Ast_BitSetType :: struct { token: Token, elem, underlying: ^Ast }
Ast_BitFieldType :: struct { token, close: Token }
Ast_MapType :: struct { token: Token, value: ^Ast }
Ast_MatrixType :: struct { token: Token, elem: ^Ast }

AstFile :: struct {
	id:           u64,
	pkg:          ^AstPackage,
	tokenizer:    Tokenizer,
	vet_flags:    u64,
	vet_flags_set: bool,
}

AstPackage :: struct {
	name: string,
}

Tokenizer :: struct {
	start: [^]u8,
	end:   [^]u8,
}

BuildContext :: struct {
	vet_packages:  StringSet,
	vet_flags:     u64,
	strict_style:  bool,
}

StringSet :: struct {
	entries: struct {
		count: int,
	},
}

// --- Vet flags ---

VetFlag_Style             :: 1 << 0
VetFlag_Deprecated        :: 1 << 1
VetFlag_ExplicitAllocators :: 1 << 2

// --- Globals ---

empty_token: Token

build_context: BuildContext

g_parsing_done: bool

// was string_interner_insert – not used in Odin port

// --- Helpers ---

ast_token :: proc(node: ^Ast) -> Token {
	#partial switch node.kind {
	case .Ident:                        return node.Ident.token
	case .Implicit:                     return node.Implicit
	case .Uninit:                       return node.Uninit
	case .BasicLit:                     return node.BasicLit.token
	case .BasicDirective:               return node.BasicDirective.token
	case .ProcGroup:                    return node.ProcGroup.token
	case .ProcLit:                      return ast_token(node.ProcLit.type)
	case .CompoundLit:
		if node.CompoundLit.type != nil {
			return ast_token(node.CompoundLit.type)
		}
		return node.CompoundLit.open
	case .TagExpr:                      return node.TagExpr.token
	case .BadExpr:                      return node.BadExpr.begin
	case .UnaryExpr:                    return node.UnaryExpr.op
	case .BinaryExpr:                   return ast_token(node.BinaryExpr.left)
	case .ParenExpr:                    return node.ParenExpr.open
	case .CallExpr:                     return ast_token(node.CallExpr.proc)
	case .SelectorExpr:
		if node.SelectorExpr.expr != nil {
			return ast_token(node.SelectorExpr.expr)
		}
		if node.SelectorExpr.selector != nil {
			return ast_token(node.SelectorExpr.selector)
		}
		return node.SelectorExpr.token
	case .SelectorCallExpr:
		if node.SelectorCallExpr.expr != nil {
			return ast_token(node.SelectorCallExpr.expr)
		}
		return node.SelectorCallExpr.token
	case .ImplicitSelectorExpr:
		if node.ImplicitSelectorExpr.selector != nil {
			return ast_token(node.ImplicitSelectorExpr.selector)
		}
		return node.ImplicitSelectorExpr.token
	case .IndexExpr:                    return ast_token(node.IndexExpr.expr)
	case .MatrixIndexExpr:              return ast_token(node.MatrixIndexExpr.expr)
	case .SliceExpr:                    return ast_token(node.SliceExpr.expr)
	case .Ellipsis:                     return node.Ellipsis.token
	case .FieldValue:
		if node.FieldValue.field != nil {
			return ast_token(node.FieldValue.field)
		}
		return node.FieldValue.eq
	case .EnumFieldValue:               return ast_token(node.EnumFieldValue.name)
	case .DerefExpr:                    return node.DerefExpr.op
	case .TernaryIfExpr:                return ast_token(node.TernaryIfExpr.x)
	case .TernaryWhenExpr:              return ast_token(node.TernaryWhenExpr.x)
	case .OrElseExpr:                   return ast_token(node.OrElseExpr.x)
	case .OrReturnExpr:                 return ast_token(node.OrReturnExpr.expr)
	case .OrBranchExpr:                 return ast_token(node.OrBranchExpr.expr)
	case .TypeAssertion:                return ast_token(node.TypeAssertion.expr)
	case .TypeCast:                     return node.TypeCast.token
	case .AutoCast:                     return node.AutoCast.token
	case .InlineAsmExpr:                return node.InlineAsmExpr.token
	case .BadStmt:                      return node.BadStmt.begin
	case .EmptyStmt:                    return node.EmptyStmt.token
	case .ExprStmt:                     return ast_token(node.ExprStmt.expr)
	case .AssignStmt:                   return node.AssignStmt.op
	case .BlockStmt:                    return node.BlockStmt.open
	case .IfStmt:                       return node.IfStmt.token
	case .WhenStmt:                     return node.WhenStmt.token
	case .ReturnStmt:                   return node.ReturnStmt.token
	case .ForStmt:                      return node.ForStmt.token
	case .RangeStmt:                    return node.RangeStmt.token
	case .UnrollRangeStmt:              return node.UnrollRangeStmt.unroll_token
	case .CaseClause:                   return node.CaseClause.token
	case .SwitchStmt:                   return node.SwitchStmt.token
	case .TypeSwitchStmt:               return node.TypeSwitchStmt.token
	case .DeferStmt:                    return node.DeferStmt.token
	case .BranchStmt:                   return node.BranchStmt.token
	case .UsingStmt:                    return node.UsingStmt.token
	case .BadDecl:                      return node.BadDecl.begin
	case .Label:                        return node.Label.token
	case .ValueDecl:                    return ast_token(node.ValueDecl.names[0])
	case .PackageDecl:                  return node.PackageDecl.token
	case .ImportDecl:                   return node.ImportDecl.token
	case .ForeignImportDecl:            return node.ForeignImportDecl.token
	case .ForeignBlockDecl:             return node.ForeignBlockDecl.token
	case .Attribute:                    return node.Attribute.token
	case .Field:
		if len(node.Field.names) > 0 {
			return ast_token(node.Field.names[0])
		}
		return ast_token(node.Field.type)
	case .FieldList:                    return node.FieldList.token
	case .TypeidType:                   return node.TypeidType.token
	case .HelperType:                   return node.HelperType.token
	case .DistinctType:                 return node.DistinctType.token
	case .PolyType:                     return node.PolyType.token
	case .ProcType:                     return node.ProcType.token
	case .RelativeType:                 return ast_token(node.RelativeType.tag)
	case .PointerType:                  return node.PointerType.token
	case .MultiPointerType:             return node.MultiPointerType.token
	case .ArrayType:                    return node.ArrayType.token
	case .DynamicArrayType:             return node.DynamicArrayType.token
	case .FixedCapacityDynamicArrayType: return node.FixedCapacityDynamicArrayType.token
	case .StructType:                   return node.StructType.token
	case .UnionType:                    return node.UnionType.token
	case .EnumType:                     return node.EnumType.token
	case .BitSetType:                   return node.BitSetType.token
	case .BitFieldType:                 return node.BitFieldType.token
	case .MapType:                      return node.MapType.token
	case .MatrixType:                   return node.MatrixType.token
	}
	return empty_token
}

token_pos_end :: proc(token: Token) -> TokenPos {
	pos := token.pos
	pos.offset += i32(len(token.string))
	for c in token.string {
		if c == '\n' {
			pos.line += 1
			pos.column = 1
		} else {
			pos.column += 1
		}
	}
	return pos
}

ast_end_token :: proc(node: ^Ast) -> Token {
	assert(node != nil)
	#partial switch node.kind {
	case .Invalid:
		return empty_token
	case .Ident:                        return node.Ident.token
	case .Implicit:                     return node.Implicit
	case .Uninit:                       return node.Uninit
	case .BasicLit:                     return node.BasicLit.token
	case .BasicDirective:               return node.BasicDirective.token
	case .ProcGroup:                    return node.ProcGroup.close
	case .ProcLit:
		if node.ProcLit.body != nil {
			return ast_end_token(node.ProcLit.body)
		}
		return ast_end_token(node.ProcLit.type)
	case .CompoundLit:                  return node.CompoundLit.close
	case .BadExpr:                      return node.BadExpr.end
	case .TagExpr:
		if node.TagExpr.expr != nil {
			return ast_end_token(node.TagExpr.expr)
		}
		return node.TagExpr.name
	case .UnaryExpr:
		if node.UnaryExpr.expr != nil {
			return ast_end_token(node.UnaryExpr.expr)
		}
		return node.UnaryExpr.op
	case .BinaryExpr:                   return ast_end_token(node.BinaryExpr.right)
	case .ParenExpr:                    return node.ParenExpr.close
	case .CallExpr:                     return node.CallExpr.close
	case .SelectorExpr:
		return ast_end_token(node.SelectorExpr.selector)
	case .SelectorCallExpr:
		return ast_end_token(node.SelectorCallExpr.call)
	case .ImplicitSelectorExpr:
		if node.ImplicitSelectorExpr.selector != nil {
			return ast_end_token(node.ImplicitSelectorExpr.selector)
		}
		return node.ImplicitSelectorExpr.token
	case .IndexExpr:                    return node.IndexExpr.close
	case .MatrixIndexExpr:              return node.MatrixIndexExpr.close
	case .SliceExpr:                    return node.SliceExpr.close
	case .Ellipsis:
		if node.Ellipsis.expr != nil {
			return ast_end_token(node.Ellipsis.expr)
		}
		return node.Ellipsis.token
	case .FieldValue:                   return ast_end_token(node.FieldValue.value)
	case .EnumFieldValue:
		if node.EnumFieldValue.value != nil {
			return ast_end_token(node.EnumFieldValue.value)
		}
		return ast_end_token(node.EnumFieldValue.name)
	case .DerefExpr:                    return node.DerefExpr.op
	case .TernaryIfExpr:                return ast_end_token(node.TernaryIfExpr.y)
	case .TernaryWhenExpr:              return ast_end_token(node.TernaryWhenExpr.y)
	case .OrElseExpr:                   return ast_end_token(node.OrElseExpr.y)
	case .OrReturnExpr:                 return node.OrReturnExpr.token
	case .OrBranchExpr:
		if node.OrBranchExpr.label != nil {
			return ast_end_token(node.OrBranchExpr.label)
		}
		return node.OrBranchExpr.token
	case .TypeAssertion:                return ast_end_token(node.TypeAssertion.type)
	case .TypeCast:                     return ast_end_token(node.TypeCast.expr)
	case .AutoCast:                     return ast_end_token(node.AutoCast.expr)
	case .InlineAsmExpr:                return node.InlineAsmExpr.close
	case .BadStmt:                      return node.BadStmt.end
	case .EmptyStmt:                    return node.EmptyStmt.token
	case .ExprStmt:                     return ast_end_token(node.ExprStmt.expr)
	case .AssignStmt:
		if len(node.AssignStmt.rhs) > 0 {
			return ast_end_token(node.AssignStmt.rhs[len(node.AssignStmt.rhs)-1])
		}
		return node.AssignStmt.op
	case .BlockStmt:                    return node.BlockStmt.close
	case .IfStmt:
		if node.IfStmt.else_stmt != nil {
			return ast_end_token(node.IfStmt.else_stmt)
		}
		return ast_end_token(node.IfStmt.body)
	case .WhenStmt:
		if node.WhenStmt.else_stmt != nil {
			return ast_end_token(node.WhenStmt.else_stmt)
		}
		return ast_end_token(node.WhenStmt.body)
	case .ReturnStmt:
		if len(node.ReturnStmt.results) > 0 {
			return ast_end_token(node.ReturnStmt.results[len(node.ReturnStmt.results)-1])
		}
		return node.ReturnStmt.token
	case .ForStmt:                      return ast_end_token(node.ForStmt.body)
	case .RangeStmt:                    return ast_end_token(node.RangeStmt.body)
	case .UnrollRangeStmt:              return ast_end_token(node.UnrollRangeStmt.body)
	case .CaseClause:
		if len(node.CaseClause.stmts) > 0 {
			return ast_end_token(node.CaseClause.stmts[len(node.CaseClause.stmts)-1])
		} else if len(node.CaseClause.list) > 0 {
			return ast_end_token(node.CaseClause.list[len(node.CaseClause.list)-1])
		}
		return node.CaseClause.token
	case .SwitchStmt:                   return ast_end_token(node.SwitchStmt.body)
	case .TypeSwitchStmt:               return ast_end_token(node.TypeSwitchStmt.body)
	case .DeferStmt:                    return ast_end_token(node.DeferStmt.stmt)
	case .BranchStmt:
		if node.BranchStmt.label != nil {
			return ast_end_token(node.BranchStmt.label)
		}
		return node.BranchStmt.token
	case .UsingStmt:
		if len(node.UsingStmt.list) > 0 {
			return ast_end_token(node.UsingStmt.list[len(node.UsingStmt.list)-1])
		}
		return node.UsingStmt.token
	case .BadDecl:                      return node.BadDecl.end
	case .Label:
		if node.Label.name != nil {
			return ast_end_token(node.Label.name)
		}
		return node.Label.token
	case .ValueDecl:
		if len(node.ValueDecl.values) > 0 {
			return ast_end_token(node.ValueDecl.values[len(node.ValueDecl.values)-1])
		}
		if node.ValueDecl.type != nil {
			return ast_end_token(node.ValueDecl.type)
		}
		if len(node.ValueDecl.names) > 0 {
			return ast_end_token(node.ValueDecl.names[len(node.ValueDecl.names)-1])
		}
		return {}
	case .PackageDecl:                  return node.PackageDecl.name
	case .ImportDecl:                   return node.ImportDecl.relpath
	case .ForeignImportDecl:
		if len(node.ForeignImportDecl.filepaths) > 0 {
			return ast_end_token(node.ForeignImportDecl.filepaths[len(node.ForeignImportDecl.filepaths)-1])
		}
		if node.ForeignImportDecl.library_name.kind != .Invalid {
			return node.ForeignImportDecl.library_name
		}
		return node.ForeignImportDecl.token
	case .ForeignBlockDecl:
		return ast_end_token(node.ForeignBlockDecl.body)
	case .Attribute:
		if node.Attribute.close.kind != .Invalid {
			return node.Attribute.close
		}
		if len(node.Attribute.elems) > 0 {
			return ast_end_token(node.Attribute.elems[len(node.Attribute.elems)-1])
		}
		if node.Attribute.open.kind != .Invalid {
			return node.Attribute.open
		}
		return node.Attribute.token
	case .Field:
		if node.Field.tag.kind != .Invalid {
			return node.Field.tag
		}
		if node.Field.default_value != nil {
			return ast_end_token(node.Field.default_value)
		}
		if node.Field.type != nil {
			return ast_end_token(node.Field.type)
		}
		return ast_end_token(node.Field.names[len(node.Field.names)-1])
	case .FieldList:
		if len(node.FieldList.list) > 0 {
			return ast_end_token(node.FieldList.list[len(node.FieldList.list)-1])
		}
		return node.FieldList.token
	case .TypeidType:
		if node.TypeidType.specialization != nil {
			return ast_end_token(node.TypeidType.specialization)
		}
		return node.TypeidType.token
	case .HelperType:                   return ast_end_token(node.HelperType.type)
	case .DistinctType:                 return ast_end_token(node.DistinctType.type)
	case .PolyType:
		if node.PolyType.specialization != nil {
			return ast_end_token(node.PolyType.specialization)
		}
		return ast_end_token(node.PolyType.type)
	case .ProcType:
		if node.ProcType.results != nil {
			return ast_end_token(node.ProcType.results)
		}
		if node.ProcType.params != nil {
			return ast_end_token(node.ProcType.params)
		}
		return node.ProcType.token
	case .RelativeType:
		return ast_end_token(node.RelativeType.type)
	case .PointerType:                  return ast_end_token(node.PointerType.type)
	case .MultiPointerType:             return ast_end_token(node.MultiPointerType.type)
	case .ArrayType:                    return ast_end_token(node.ArrayType.elem)
	case .DynamicArrayType:             return ast_end_token(node.DynamicArrayType.elem)
	case .FixedCapacityDynamicArrayType: return ast_end_token(node.FixedCapacityDynamicArrayType.elem)
	case .StructType:
		if len(node.StructType.fields) > 0 {
			return ast_end_token(node.StructType.fields[len(node.StructType.fields)-1])
		}
		return node.StructType.token
	case .UnionType:
		if len(node.UnionType.variants) > 0 {
			return ast_end_token(node.UnionType.variants[len(node.UnionType.variants)-1])
		}
		return node.UnionType.token
	case .EnumType:
		if len(node.EnumType.fields) > 0 {
			return ast_end_token(node.EnumType.fields[len(node.EnumType.fields)-1])
		}
		if node.EnumType.base_type != nil {
			return ast_end_token(node.EnumType.base_type)
		}
		return node.EnumType.token
	case .BitSetType:
		if node.BitSetType.underlying != nil {
			return ast_end_token(node.BitSetType.underlying)
		}
		return ast_end_token(node.BitSetType.elem)
	case .BitFieldType:                 return node.BitFieldType.close
	case .MapType:                      return ast_end_token(node.MapType.value)
	case .MatrixType:                   return ast_end_token(node.MatrixType.elem)
	}
	return empty_token
}

ast_end_pos :: proc(node: ^Ast) -> TokenPos {
	return token_pos_end(ast_end_token(node))
}

// --- Vet flag helpers ---

in_vet_packages :: proc(file: ^AstFile) -> bool {
	if file == nil {
		return true
	}
	if file.pkg == nil {
		return true
	}
	if build_context.vet_packages.entries.count == 0 {
		return true
	}
	// was string_set_exists(&build_context.vet_packages, file.pkg.name)
	return true
}

ast_file_vet_flags :: proc(f: ^AstFile) -> u64 {
	if f != nil && f.vet_flags_set {
		return f.vet_flags
	}
	found := in_vet_packages(f)
	if found {
		return build_context.vet_flags
	}
	return 0
}

ast_file_vet_style :: proc(f: ^AstFile) -> bool {
	return (ast_file_vet_flags(f) & VetFlag_Style) != 0
}

ast_file_vet_deprecated :: proc(f: ^AstFile) -> bool {
	return (ast_file_vet_flags(f) & VetFlag_Deprecated) != 0
}

ast_file_vet_explicit_allocators :: proc(f: ^AstFile) -> bool {
	return (ast_file_vet_flags(f) & VetFlag_ExplicitAllocators) != 0
}

file_allow_newline :: proc(f: ^AstFile) -> bool {
	is_strict := build_context.strict_style || ast_file_vet_style(f)
	return !is_strict
}

token_end_of_line :: proc(f: ^AstFile, tok: Token) -> Token {
	tok := tok
	start := mem.ptr_offset(f.tokenizer.start, tok.pos.offset)
	s := start
	end_ptr := f.tokenizer.end
	for s^ != 0 && s^ != '\n' && s < end_ptr {
		s = mem.ptr_offset(s, 1)
	}
	tok.pos.column += i32(mem.ptr_sub(s, start)) - 1
	return tok
}

get_file_line_as_string :: proc(pos: TokenPos, offset_: ^i32) -> string {
	file := thread_safe_get_ast_file_from_id(pos.file_id)
	if file == nil {
		return ""
	}
	start := file.tokenizer.start
	end_ptr := file.tokenizer.end
	if start == end_ptr {
		return ""
	}

	offset := pos.offset
	if pos.line != 0 && offset == 0 {
		for i in 1 ..< pos.line {
			for {
				ptr := mem.ptr_offset(start, offset)
				if ptr >= end_ptr {
					break
				}
				c := ptr^
				offset += 1
				if c == '\n' {
					break
				}
			}
		}
		for i in 1 ..< pos.column {
			ptr := mem.ptr_offset(start, offset)
			c := ptr^
			if (c & 0x80) != 0 {
				rune_len, _ := utf8.decode_rune(mem.byte_slice(ptr, int(mem.ptr_sub(end_ptr, ptr))))
				offset += rune_len
			} else {
				offset += 1
			}
		}
	}

	len_total := int(mem.ptr_sub(end_ptr, start))
	if len_total < offset {
		return ""
	}

	pos_offset := mem.ptr_offset(start, offset)
	line_start := pos_offset
	line_end := pos_offset

	if offset > 0 && line_start^ == '\n' {
		line_start = mem.ptr_offset(line_start, -1)
	}
	for line_start >= start {
		if line_start^ == '\n' {
			line_start = mem.ptr_offset(line_start, 1)
			break
		}
		line_start = mem.ptr_offset(line_start, -1)
	}
	if line_start == mem.ptr_offset(start, -1) {
		line_start = mem.ptr_offset(line_start, 1)
	}
	for line_end < end_ptr {
		if line_end^ == '\n' {
			break
		}
		line_end = mem.ptr_offset(line_end, 1)
	}

	line_len := int(mem.ptr_sub(line_end, line_start))
	the_line := strings.trim_space(string(line_start[:line_len]))
	if offset_ != nil {
		offset_^ = i32(mem.ptr_sub(pos_offset, raw_data(the_line)))
	}
	return strings.clone(the_line)
}

// --- AST node allocation ---

ast_node_size :: proc(kind: AstKind) -> int {
	return 0 // stub: depends on ast_variant_sizes table
}

alloc_ast_node :: proc(f: ^AstFile, kind: AstKind, allocator := context.allocator) -> ^Ast {
	size := ast_node_size(kind)
	node := cast(^Ast)mem.alloc(size, 16, allocator)
	node.kind = kind
	if f != nil {
		node.file_id = f.id
	}
	return node
}

// --- External dependency stubs ---

thread_safe_get_ast_file_from_id :: proc(id: u64) -> ^AstFile {
	return nil // stub
}
