package docs

import "sort"

// ============================================================================
// Abstract types for the writer (mirroring the Odin compiler's Entity, Type, etc.)
// ============================================================================

// EntityKind mirrors the Odin compiler's Entity kind enum.
type EntityKind int32

const (
	EntityInvalid     EntityKind = 0
	EntityConstant    EntityKind = 1
	EntityVariable    EntityKind = 2
	EntityTypeName    EntityKind = 3
	EntityProcedure   EntityKind = 4
	EntityProcGroup   EntityKind = 5
	EntityImportName  EntityKind = 6
	EntityLibraryName EntityKind = 7
	EntityBuiltin     EntityKind = 8
	EntityNil         EntityKind = 9
	EntityLabel       EntityKind = 10
	EntityCount       EntityKind = 11
)

// EntityFlag constants
type EntityFlag uint64

const (
	EntityFlagUsing         EntityFlag = 1 << 2
	EntityFlagConstInput    EntityFlag = 1 << 3
	EntityFlagEllipsis      EntityFlag = 1 << 4
	EntityFlagNoAlias       EntityFlag = 1 << 5
	EntityFlagAnyInt        EntityFlag = 1 << 6
	EntityFlagByPtr         EntityFlag = 1 << 7
	EntityFlagNoBroadcast   EntityFlag = 1 << 8
	EntityFlagBitFieldField EntityFlag = 1 << 9
	EntityFlagStatic        EntityFlag = 1 << 10
)

// EntConstantFlag for constant-specific flags.
type EntConstantFlag uint32

const (
	EntConstantFlagImplicitEnumValue EntConstantFlag = 1 << 0
)

// PackageKind
type PackageKind int32

const (
	PackageNormal  PackageKind = 0
	PackageRuntime PackageKind = 1
	PackageInit    PackageKind = 2
	PackageBuiltin PackageKind = 3
)

// TokenPos represents a source position.
type TokenPos struct {
	FileID int32
	Line   int32
	Column int32
	Offset int32
}

// String represents a string with length (mirroring Odin's String type).
type String struct {
	Text []byte
}

// MakeString creates a String from a Go string.
func MakeString(s string) String {
	return String{Text: []byte(s)}
}

// String returns the Go string representation.
func (s String) String() string {
	return string(s.Text)
}

// Len returns the length.
func (s String) Len() int {
	return len(s.Text)
}

// Comment represents a single comment.
type Comment struct {
	String String
}

// CommentGroup represents a group of comments.
type CommentGroup struct {
	List []Comment
}

// AstPackage represents an Odin package (abstract).
type AstPackage struct {
	Name     String
	Fullpath String
	Kind     PackageKind
	IsExtra  bool
	Scope    *Scope
	Files    []*AstFile
}

// AstFile represents a source file.
type AstFile struct {
	Fullpath   String
	PkgDecl    *Ast
	OrderInSrc uint64
}

// Scope represents a scope.
type Scope struct {
	Elements []ScopeElement
	Flags    uint64
}

// ScopeElement is a single element in a scope.
type ScopeElement struct {
	Name  String
	Value *Entity
}

// ScopeFlag
type ScopeFlag uint64

const (
	ScopeFlagFile ScopeFlag = 1 << 0
	ScopeFlagPkg  ScopeFlag = 1 << 1
)

// Entity represents a named entity in Odin.
type Entity struct {
	Kind       EntityKind
	Flags      EntityFlag
	Token      Token
	Pkg        *AstPackage
	File       *AstFile
	OrderInSrc uint64
	Type       *Type
	DeclInfo   *DeclInfo
	Scope      *Scope

	// Union fields — prefixed with Ent to avoid clash with EntityKind consts
	Constant  EntConstantData
	Variable  EntVariableData
	Procedure EntProcedureData
	TypeName  EntTypeNameData
	ProcGroup EntProcGroupData
	Builtin   EntBuiltinData
}

// Token represents a lexical token.
type Token struct {
	String String
	Pos    TokenPos
}

// EntConstantData holds constant-specific data.
type EntConstantData struct {
	Comment         *CommentGroup
	Docs            *CommentGroup
	Value           ExactValueStr
	ParamValue      ParamValue
	FieldGroupIndex int32
	Flags           EntConstantFlag
}

// EntVariableData holds variable-specific data.
type EntVariableData struct {
	Comment          *CommentGroup
	Docs             *CommentGroup
	IsForeign        bool
	IsExport         bool
	ThreadLocalModel string
	LinkName         String
	InitExpr         *Ast
	ForeignLibrary   *Entity
	FieldGroupIndex  int32
	BitFieldBitSize  uint32
	ParamValue       ParamValue
}

// EntProcedureData holds procedure-specific data.
type EntProcedureData struct {
	IsForeign      bool
	IsExport       bool
	LinkName       String
	ForeignLibrary *Entity
}

// EntTypeNameData holds type-name-specific data.
type EntTypeNameData struct {
	IsTypeAlias bool
}

// EntProcGroupData holds proc-group-specific data.
type EntProcGroupData struct {
	Entities []*Entity
}

// EntBuiltinData holds builtin-specific data.
type EntBuiltinData struct {
	ID int32
}

// DeclInfo holds declaration info.
type DeclInfo struct {
	TypeExpr   *Ast
	InitExpr   *Ast
	DeclNode   *Ast
	Comment    *CommentGroup
	Docs       *CommentGroup
	Attributes []*Ast
}

// ExactValueStr is a string representation of an exact value.
type ExactValueStr struct {
	Str string
}

// ParamValue holds parameter value data.
type ParamValue struct {
	OriginalAstExpr *Ast
}

// Ast is an abstract syntax tree node (opaque).
type Ast struct {
	Kind AstKind
}

// AstKind enumerates AST node kinds.
type AstKind int32

const (
	AstInvalid     AstKind = 0
	AstIdent       AstKind = 1
	AstImplicit    AstKind = 2
	AstFieldValue  AstKind = 3
	AstPackageDecl AstKind = 4
	AstStructType  AstKind = 5
	AstUnionType   AstKind = 6
	AstAttribute   AstKind = 7
)

// AstIdentNode represents an identifier node.
type AstIdentNode struct {
	Token Token
}

// AstImplicitNode represents an implicit selector.
type AstImplicitNode struct {
	String String
}

// AstFieldValueNode represents a field value.
type AstFieldValueNode struct {
	Field *Ast
	Value *Ast
}

// AstAttributeNode represents an attribute node.
type AstAttributeNode struct {
	Elems []*Ast
}

// AstPackageDeclNode represents a package declaration.
type AstPackageDeclNode struct {
	Docs *CommentGroup
}

// AstStructTypeNode represents a struct type node.
type AstStructTypeNode struct {
	Align        *Ast
	WhereClauses []*Ast
}

// AstUnionTypeNode represents a union type node.
type AstUnionTypeNode struct {
	Align        *Ast
	WhereClauses []*Ast
}

// BuiltinProcPkg
type BuiltinProcPkg int32

const (
	BuiltinProcPkgBuiltin    BuiltinProcPkg = 0
	BuiltinProcPkgIntrinsics BuiltinProcPkg = 1
)

// BuiltinProc represents a builtin procedure definition.
type BuiltinProc struct {
	Name String
	Pkg  BuiltinProcPkg
}

// BuiltinProcs is the table of builtin procedures.
var BuiltinProcs []BuiltinProc

// Type represents a type in the Odin type system.
type Type struct {
	Kind TypeKind

	// Union fields
	Basic                 TypBasic
	Named                 TypNamed
	Generic               TypGeneric
	Pointer               TypPointer
	MultiPointer          TypMultiPointer
	SoaPointer            TypSoaPointer
	Array                 TypArray
	EnumeratedArray       TypEnumeratedArray
	Slice                 TypSlice
	DynamicArray          TypDynamicArray
	FixedCapacityDynArray TypFixedCapacityDynArray
	Map                   TypMap
	BitField              TypBitField
	Struct                TypStruct
	Union                 TypUnion
	Enum                  TypEnum
	Tuple                 TypTuple
	Proc                  TypProc
	BitSet                TypBitSet
	SimdVector            TypSimdVector
	Matrix                TypMatrix
}

// TypeKind enumerates type kinds.
type TypeKind int32

const (
	TypeInvalid               TypeKind = 0
	TypeBasic                 TypeKind = 1
	TypeNamed                 TypeKind = 2
	TypeGeneric               TypeKind = 3
	TypePointer               TypeKind = 4
	TypeMultiPointer          TypeKind = 5
	TypeSoaPointer            TypeKind = 6
	TypeArray                 TypeKind = 7
	TypeEnumeratedArray       TypeKind = 8
	TypeSlice                 TypeKind = 9
	TypeDynamicArray          TypeKind = 10
	TypeFixedCapacityDynArray TypeKind = 11
	TypeMap                   TypeKind = 12
	TypeBitField              TypeKind = 13
	TypeStruct                TypeKind = 14
	TypeUnion                 TypeKind = 15
	TypeEnum                  TypeKind = 16
	TypeTuple                 TypeKind = 17
	TypeProc                  TypeKind = 18
	TypeBitSet                TypeKind = 19
	TypeSimdVector            TypeKind = 20
	TypeMatrix                TypeKind = 21
)

// StructSoaKind
type StructSoaKind int32

const (
	StructSoaNone    StructSoaKind = 0
	StructSoaFixed   StructSoaKind = 1
	StructSoaSlice   StructSoaKind = 2
	StructSoaDynamic StructSoaKind = 3
)

// UnionTypeKind
type UnionTypeKind int32

const (
	UnionTypeNormal    UnionTypeKind = 0
	UnionTypeNoNil     UnionTypeKind = 1
	UnionTypeSharedNil UnionTypeKind = 2
)

// CallingConvention
type CallingConvention int32

// ProcCallingConventionStrings maps calling convention to string.
var ProcCallingConventionStrings []string

// TypBasic
type TypBasic struct {
	Name String
}

// TypNamed
type TypNamed struct {
	Name     String
	Base     *Type
	TypeName *Entity
}

// TypGeneric
type TypGeneric struct {
	Name        String
	Entity      *Entity
	Specialized *Type
}

// TypPointer
type TypPointer struct {
	Elem *Type
}

// TypMultiPointer
type TypMultiPointer struct {
	Elem *Type
}

// TypSoaPointer
type TypSoaPointer struct {
	Elem *Type
}

// TypArray
type TypArray struct {
	Count        int64
	Elem         *Type
	GenericCount *Type
}

// TypEnumeratedArray
type TypEnumeratedArray struct {
	Count int64
	Index *Type
	Elem  *Type
}

// TypSlice
type TypSlice struct {
	Elem *Type
}

// TypDynamicArray
type TypDynamicArray struct {
	Elem *Type
}

// TypFixedCapacityDynArray
type TypFixedCapacityDynArray struct {
	Capacity        int64
	Elem            *Type
	GenericCapacity *Type
}

// TypMap
type TypMap struct {
	Key   *Type
	Value *Type
}

// TypBitField
type TypBitField struct {
	Fields      []*Entity
	BackingType *Type
}

// TypStruct
type TypStruct struct {
	SoaKind             StructSoaKind
	SoaCount            int64
	SoaElem             *Type
	IsPolymorphic       bool
	IsPacked            bool
	IsRawUnion          bool
	IsAllOrNone         bool
	CustomMinFieldAlign int32
	CustomMaxFieldAlign int32
	Fields              []*Entity
	PolymorphicParams   *Type
	Node                *Ast
	Tags                []String
}

// TypUnion
type TypUnion struct {
	IsPolymorphic     bool
	Kind              UnionTypeKind
	Variants          []*Type
	PolymorphicParams *Type
	Node              *Ast
}

// TypEnum
type TypEnum struct {
	Fields   []*Entity
	BaseType *Type
}

// TypTuple
type TypTuple struct {
	Variables []*Entity
}

// TypProc
type TypProc struct {
	IsPolymorphic     bool
	Diverging         bool
	OptionalOk        bool
	Variadic          bool
	CVararg           bool
	Params            *Type
	Results           *Type
	CallingConvention CallingConvention
}

// TypBitSet
type TypBitSet struct {
	Elem       *Type
	Underlying *Type
	Lower      int64
	Upper      int64
}

// TypSimdVector
type TypSimdVector struct {
	Count int64
	Elem  *Type
}

// TypMatrix
type TypMatrix struct {
	RowCount    int64
	ColumnCount int64
	Elem        *Type
}

// IsEntityExported checks if an entity is exported.
func IsEntityExported(e *Entity, allowBuiltin bool) bool {
	if e == nil {
		return false
	}
	if e.Kind == EntityBuiltin {
		return allowBuiltin
	}
	if e.Variable.IsExport || e.Procedure.IsExport {
		return true
	}
	if e.Token.String.Len() > 0 && e.Token.String.Text[0] >= 'A' && e.Token.String.Text[0] <= 'Z' {
		return true
	}
	return false
}

// printEntityKindOrdering maps EntityKind to printing order.
var printEntityKindOrdering = [EntityCount]int{
	EntityInvalid:     -1,
	EntityConstant:    0,
	EntityVariable:    1,
	EntityTypeName:    4,
	EntityProcedure:   2,
	EntityProcGroup:   3,
	EntityImportName:  -1,
	EntityLibraryName: -1,
	EntityBuiltin:     -1,
	EntityNil:         -1,
	EntityLabel:       -1,
}

// printEntityNames maps EntityKind to display name.
var printEntityNames = [EntityCount]string{
	EntityInvalid:     "",
	EntityConstant:    "constants",
	EntityVariable:    "variables",
	EntityTypeName:    "types",
	EntityProcedure:   "procedures",
	EntityProcGroup:   "proc_group",
	EntityImportName:  "import names",
	EntityLibraryName: "library names",
	EntityBuiltin:     "",
	EntityNil:         "",
	EntityLabel:       "",
}

// CmdDocFlag constants define documentation generation flags.
type CmdDocFlag uint32

const (
	CmdDocFlagShort         CmdDocFlag = 1 << 0
	CmdDocFlagAllPackages   CmdDocFlag = 1 << 1
	CmdDocFlagDocFormat     CmdDocFlag = 1 << 2
	CmdDocFlagInSourceOrder CmdDocFlag = 1 << 3
)

// CmpEntitiesForPrinting compares two entities for sorting in documentation output.
func CmpEntitiesForPrinting(x, y *Entity) int {
	if x.Pkg != y.Pkg {
		if x.Pkg == nil {
			return -1
		}
		if y.Pkg == nil {
			return 1
		}
		res := compareStrings(x.Pkg.Name, y.Pkg.Name)
		if res != 0 {
			return res
		}
	}
	ox := printEntityKindOrdering[x.Kind]
	oy := printEntityKindOrdering[y.Kind]
	res := ox - oy
	if res == 0 {
		res = compareStrings(x.Token.String, y.Token.String)
	}
	return res
}

// CmpEntitiesForPrintingByOrderInSrc compares entities by source order.
func CmpEntitiesForPrintingByOrderInSrc(x, y *Entity) int {
	if x.Pkg != y.Pkg {
		if x.Pkg == nil {
			return -1
		}
		if y.Pkg == nil {
			return 1
		}
		res := compareStrings(x.Pkg.Name, y.Pkg.Name)
		if res != 0 {
			return res
		}
	}
	sx := x.OrderInSrc
	sy := y.OrderInSrc
	if sx < sy {
		return -1
	}
	if sx > sy {
		return 1
	}
	if x.Token.Pos.Offset < y.Token.Pos.Offset {
		return -1
	}
	if x.Token.Pos.Offset > y.Token.Pos.Offset {
		return 1
	}
	return 0
}

// CmpAstPackageByName compares packages by name.
func CmpAstPackageByName(x, y *AstPackage) int {
	return compareStrings(x.Name, y.Name)
}

func compareStrings(a, b String) int {
	aStr := string(a.Text)
	bStr := string(b.Text)
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

// SortEntitiesByKind sorts a slice of entities by kind order.
func SortEntitiesByKind(entities []*Entity) {
	sort.Slice(entities, func(i, j int) bool {
		return CmpEntitiesForPrinting(entities[i], entities[j]) < 0
	})
}

// SortEntitiesBySrcOrder sorts a slice of entities by source order.
func SortEntitiesBySrcOrder(entities []*Entity) {
	sort.Slice(entities, func(i, j int) bool {
		return CmpEntitiesForPrintingByOrderInSrc(entities[i], entities[j]) < 0
	})
}

// SortPackagesByName sorts packages by name.
func SortPackagesByName(pkgs []*AstPackage) {
	sort.Slice(pkgs, func(i, j int) bool {
		return CmpAstPackageByName(pkgs[i], pkgs[j]) < 0
	})
}
