// Depends on: common.odin (String, isize, u8, u32, u64, i64, Slice, Array, Entity, AstFile, AstPackage, Type, Token, TokenPos, CommentGroup, Ast, CheckerInfo, Checker)
package cmd

type OdinDocFileIndex uint32
type OdinDocPkgIndex uint32
type OdinDocEntityIndex uint32
type OdinDocTypeIndex uint32

type OdinDocArray[T any] struct {
	Offset uint32
	Length uint32
}

type OdinDocString = OdinDocArray[uint8]

type OdinDocVersionType struct {
	Major uint8
	Minor uint8
	Patch uint8
	Pad0  uint8
}

type OdinDocHeaderBase struct {
	Magic      [8]uint8
	Padding0   uint32
	Version    OdinDocVersionType
	TotalSize  uint32
	HeaderSize uint32
	Hash       uint32
}

type OdinDocFile struct {
	Pkg  OdinDocPkgIndex
	Name OdinDocString
}

type OdinDocPosition struct {
	File   OdinDocFileIndex
	Line   uint32
	Column uint32
	Offset uint32
}

type OdinDocTypeKind uint32

const (
	OdinDocTypeInvalid                   OdinDocTypeKind = 0
	OdinDocTypeBasic                     OdinDocTypeKind = 1
	OdinDocTypeNamed                     OdinDocTypeKind = 2
	OdinDocTypeGeneric                   OdinDocTypeKind = 3
	OdinDocTypePointer                   OdinDocTypeKind = 4
	OdinDocTypeArray                     OdinDocTypeKind = 5
	OdinDocTypeEnumeratedArray           OdinDocTypeKind = 6
	OdinDocTypeSlice                     OdinDocTypeKind = 7
	OdinDocTypeDynamicArray              OdinDocTypeKind = 8
	OdinDocTypeMap                       OdinDocTypeKind = 9
	OdinDocTypeStruct                    OdinDocTypeKind = 10
	OdinDocTypeUnion                     OdinDocTypeKind = 11
	OdinDocTypeEnum                      OdinDocTypeKind = 12
	OdinDocTypeTuple                     OdinDocTypeKind = 13
	OdinDocTypeProc                      OdinDocTypeKind = 14
	OdinDocTypeBitSet                    OdinDocTypeKind = 15
	OdinDocTypeSimdVector                OdinDocTypeKind = 16
	OdinDocTypeSOAStructFixed            OdinDocTypeKind = 17
	OdinDocTypeSOAStructSlice            OdinDocTypeKind = 18
	OdinDocTypeSOAStructDynamic          OdinDocTypeKind = 19
	OdinDocTypeMultiPointer              OdinDocTypeKind = 22
	OdinDocTypeMatrix                    OdinDocTypeKind = 23
	OdinDocTypeSoaPointer                OdinDocTypeKind = 24
	OdinDocTypeBitField                  OdinDocTypeKind = 25
	OdinDocTypeFixedCapacityDynamicArray OdinDocTypeKind = 26
)

const OdinDocTypeElemsCap = 4

type OdinDocType struct {
	Kind              OdinDocTypeKind
	Flags             uint32
	Name              OdinDocString
	CustomAlign       OdinDocString
	ElemCountLen      uint32
	ElemCounts        [OdinDocTypeElemsCap]int64
	CallingConvention OdinDocString
	Types             OdinDocArray[OdinDocTypeIndex]
	Entities          OdinDocArray[OdinDocEntityIndex]
	PolymorphicParams OdinDocTypeIndex
	WhereClauses      OdinDocArray[OdinDocString]
	Tags              OdinDocArray[OdinDocString]
}

type OdinDocAttribute struct {
	Name  OdinDocString
	Value OdinDocString
}

type OdinDocEntityKind uint32

const (
	OdinDocEntityInvalid     OdinDocEntityKind = 0
	OdinDocEntityConstant    OdinDocEntityKind = 1
	OdinDocEntityVariable    OdinDocEntityKind = 2
	OdinDocEntityTypeName    OdinDocEntityKind = 3
	OdinDocEntityProcedure   OdinDocEntityKind = 4
	OdinDocEntityProcGroup   OdinDocEntityKind = 5
	OdinDocEntityImportName  OdinDocEntityKind = 6
	OdinDocEntityLibraryName OdinDocEntityKind = 7
	OdinDocEntityBuiltin     OdinDocEntityKind = 8
)

const (
	OdinDocEntityFlagForeign              uint64 = 1 << 0
	OdinDocEntityFlagExport               uint64 = 1 << 1
	OdinDocEntityFlagParamUsing           uint64 = 1 << 2
	OdinDocEntityFlagParamConst           uint64 = 1 << 3
	OdinDocEntityFlagParamAutoCast        uint64 = 1 << 4
	OdinDocEntityFlagParamEllipsis        uint64 = 1 << 5
	OdinDocEntityFlagParamCVararg         uint64 = 1 << 6
	OdinDocEntityFlagParamNoAlias         uint64 = 1 << 7
	OdinDocEntityFlagParamAnyInt          uint64 = 1 << 8
	OdinDocEntityFlagParamByPtr           uint64 = 1 << 9
	OdinDocEntityFlagParamNoBroadcast     uint64 = 1 << 10
	OdinDocEntityFlagBitFieldField        uint64 = 1 << 19
	OdinDocEntityFlagTypeAlias            uint64 = 1 << 20
	OdinDocEntityFlagBuiltinPkgBuiltin    uint64 = 1 << 30
	OdinDocEntityFlagBuiltinPkgIntrinsics uint64 = 1 << 31
	OdinDocEntityFlagVarThreadLocal       uint64 = 1 << 40
	OdinDocEntityFlagVarStatic            uint64 = 1 << 41
	OdinDocEntityFlagPrivate              uint64 = 1 << 50
)

type OdinDocEntity struct {
	Kind            OdinDocEntityKind
	Reserved        uint32
	Flags           uint64
	Pos             OdinDocPosition
	Name            OdinDocString
	Type            OdinDocTypeIndex
	InitString      OdinDocString
	ReservedForInit uint32
	Comment         OdinDocString
	Docs            OdinDocString
	FieldGroupIndex int32
	ForeignLibrary  OdinDocEntityIndex
	LinkName        OdinDocString
	Attributes      OdinDocArray[OdinDocAttribute]
	GroupedEntities OdinDocArray[OdinDocEntityIndex]
	WhereClauses    OdinDocArray[OdinDocString]
}

type OdinDocScopeEntry struct {
	Name   OdinDocString
	Entity OdinDocEntityIndex
}

const (
	OdinDocPkgFlagBuiltin uint32 = 1 << 0
	OdinDocPkgFlagRuntime uint32 = 1 << 1
	OdinDocPkgFlagInit    uint32 = 1 << 2
)

type OdinDocPkg struct {
	Fullpath OdinDocString
	Name     OdinDocString
	Flags    uint32
	Docs     OdinDocString
	Files    OdinDocArray[OdinDocFileIndex]
	Entries  OdinDocArray[OdinDocScopeEntry]
}

type OdinDocHeader struct {
	Base     OdinDocHeaderBase
	Files    OdinDocArray[OdinDocFile]
	Pkgs     OdinDocArray[OdinDocPkg]
	Entities OdinDocArray[OdinDocEntity]
	Types    OdinDocArray[OdinDocType]
}

type OdinDocWriterItemTracker[T any] struct {
	Len    isize
	Cap    isize
	Offset isize
}

type OdinDocWriterState int

const (
	OdinDocWriterStatePreparing OdinDocWriterState = iota
	OdinDocWriterStateWriting
)

var OdinDocWriterStateStrings = []string{"preparing", "writing  "}

var gInDocWriter bool

type OdinDocWriter struct {
	Info        *CheckerInfo
	State       OdinDocWriterState
	Data        []byte
	DataLen     isize
	Header      *OdinDocHeader
	StringCache map[OdinDocString]OdinDocString
	FileCache   map[*AstFile]OdinDocFileIndex
	PkgCache    map[*AstPackage]OdinDocPkgIndex
	EntityCache map[*Entity]OdinDocEntityIndex
	TypeCache   map[uint64]OdinDocTypeIndex
	Files       OdinDocWriterItemTracker[OdinDocFile]
	Pkgs        OdinDocWriterItemTracker[OdinDocPkg]
	Entities    OdinDocWriterItemTracker[OdinDocEntity]
	Types       OdinDocWriterItemTracker[OdinDocType]
	Strings     OdinDocWriterItemTracker[uint8]
	Blob        OdinDocWriterItemTracker[uint8]
}
