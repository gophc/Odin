package docs

// ============================================================================
// Binary format type definitions (mirrors src/docs_format.cpp)
// ============================================================================

// OdinDocArray is a slice reference within the binary doc buffer.
type OdinDocArray struct {
	Offset uint32
	Length uint32
}

// OdinDocString is a string reference within the binary doc buffer.
type OdinDocString struct {
	Offset uint32
	Length uint32
}

// OdinDocVersionType stores the version of the odin-doc binary format.
type OdinDocVersionType struct {
	Major uint8
	Minor uint8
	Patch uint8
	Pad0  uint8
}

// OdinDocHeaderBase is the base header for the odin-doc binary format.
type OdinDocHeaderBase struct {
	Magic      [8]byte
	Padding0   uint32
	Version    OdinDocVersionType
	TotalSize  uint32
	HeaderSize uint32
	Hash       uint32
}

// Index types used in the binary format.
type (
	OdinDocFileIndex   = uint32
	OdinDocPkgIndex    = uint32
	OdinDocEntityIndex = uint32
	OdinDocTypeIndex   = uint32
)

// OdinDocFile represents a source file entry.
type OdinDocFile struct {
	Pkg  OdinDocPkgIndex
	Name OdinDocString
}

// OdinDocPosition stores source location information.
type OdinDocPosition struct {
	File   OdinDocFileIndex
	Line   uint32
	Column uint32
	Offset uint32
}

// OdinDocTypeKind enumerates type kinds in the doc format.
type OdinDocTypeKind uint32

const (
	OdinDocTypeInvalid               OdinDocTypeKind = 0
	OdinDocTypeBasic                 OdinDocTypeKind = 1
	OdinDocTypeNamed                 OdinDocTypeKind = 2
	OdinDocTypeGeneric               OdinDocTypeKind = 3
	OdinDocTypePointer               OdinDocTypeKind = 4
	OdinDocTypeArray                 OdinDocTypeKind = 5
	OdinDocTypeEnumeratedArray       OdinDocTypeKind = 6
	OdinDocTypeSlice                 OdinDocTypeKind = 7
	OdinDocTypeDynamicArray          OdinDocTypeKind = 8
	OdinDocTypeMap                   OdinDocTypeKind = 9
	OdinDocTypeStruct                OdinDocTypeKind = 10
	OdinDocTypeUnion                 OdinDocTypeKind = 11
	OdinDocTypeEnum                  OdinDocTypeKind = 12
	OdinDocTypeTuple                 OdinDocTypeKind = 13
	OdinDocTypeProc                  OdinDocTypeKind = 14
	OdinDocTypeBitSet                OdinDocTypeKind = 15
	OdinDocTypeSimdVector            OdinDocTypeKind = 16
	OdinDocTypeSOAStructFixed        OdinDocTypeKind = 17
	OdinDocTypeSOAStructSlice        OdinDocTypeKind = 18
	OdinDocTypeSOAStructDynamic      OdinDocTypeKind = 19
	OdinDocTypeMultiPointer          OdinDocTypeKind = 22
	OdinDocTypeMatrix                OdinDocTypeKind = 23
	OdinDocTypeSoaPointer            OdinDocTypeKind = 24
	OdinDocTypeBitField              OdinDocTypeKind = 25
	OdinDocTypeFixedCapacityDynArray OdinDocTypeKind = 26
)

// Type flag constants
const (
	OdinDocTypeFlagBasicUntyped uint32 = 1 << 1

	OdinDocTypeFlagStructPolymorphic uint32 = 1 << 0
	OdinDocTypeFlagStructPacked      uint32 = 1 << 1
	OdinDocTypeFlagStructRawUnion    uint32 = 1 << 2
	OdinDocTypeFlagStructAllOrNone   uint32 = 1 << 3

	OdinDocTypeFlagUnionPolymorphic uint32 = 1 << 0
	OdinDocTypeFlagUnionNoNil       uint32 = 1 << 1
	OdinDocTypeFlagUnionSharedNil   uint32 = 1 << 3

	OdinDocTypeFlagProcPolymorphic uint32 = 1 << 0
	OdinDocTypeFlagProcDiverging   uint32 = 1 << 1
	OdinDocTypeFlagProcOptionalOk  uint32 = 1 << 2
	OdinDocTypeFlagProcVariadic    uint32 = 1 << 3
	OdinDocTypeFlagProcCVararg     uint32 = 1 << 4

	OdinDocTypeFlagBitSetRange          uint32 = 1 << 1
	OdinDocTypeFlagBitSetOpLt           uint32 = 1 << 2
	OdinDocTypeFlagBitSetOpLtEq         uint32 = 1 << 3
	OdinDocTypeFlagBitSetUnderlyingType uint32 = 1 << 4
)

const OdinDocTypeElemsCap = 4

// OdinDocType represents a type entry in the binary doc format.
type OdinDocType struct {
	Kind              OdinDocTypeKind
	Flags             uint32
	Name              OdinDocString
	CustomAlign       OdinDocString
	ElemCountLen      uint32
	ElemCounts        [OdinDocTypeElemsCap]int64
	CallingConvention OdinDocString
	Types             OdinDocArray
	Entities          OdinDocArray
	PolymorphicParams OdinDocTypeIndex
	WhereClauses      OdinDocArray
	Tags              OdinDocArray
}

// OdinDocAttribute represents an attribute entry.
type OdinDocAttribute struct {
	Name  OdinDocString
	Value OdinDocString
}

// OdinDocEntityKind enumerates entity kinds.
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

// Entity flag constants (bit positions match C++)
type OdinDocEntityFlag uint64

const (
	OdinDocEntityFlagForeign              OdinDocEntityFlag = 1 << 0
	OdinDocEntityFlagExport               OdinDocEntityFlag = 1 << 1
	OdinDocEntityFlagParamUsing           OdinDocEntityFlag = 1 << 2
	OdinDocEntityFlagParamConst           OdinDocEntityFlag = 1 << 3
	OdinDocEntityFlagParamAutoCast        OdinDocEntityFlag = 1 << 4
	OdinDocEntityFlagParamEllipsis        OdinDocEntityFlag = 1 << 5
	OdinDocEntityFlagParamCVararg         OdinDocEntityFlag = 1 << 6
	OdinDocEntityFlagParamNoAlias         OdinDocEntityFlag = 1 << 7
	OdinDocEntityFlagParamAnyInt          OdinDocEntityFlag = 1 << 8
	OdinDocEntityFlagParamByPtr           OdinDocEntityFlag = 1 << 9
	OdinDocEntityFlagParamNoBroadcast     OdinDocEntityFlag = 1 << 10
	OdinDocEntityFlagBitFieldField        OdinDocEntityFlag = 1 << 19
	OdinDocEntityFlagTypeAlias            OdinDocEntityFlag = 1 << 20
	OdinDocEntityFlagBuiltinPkgBuiltin    OdinDocEntityFlag = 1 << 30
	OdinDocEntityFlagBuiltinPkgIntrinsics OdinDocEntityFlag = 1 << 31
	OdinDocEntityFlagVarThreadLocal       OdinDocEntityFlag = 1 << 40
	OdinDocEntityFlagVarStatic            OdinDocEntityFlag = 1 << 41
	OdinDocEntityFlagPrivate              OdinDocEntityFlag = 1 << 50
)

// OdinDocEntity represents an entity entry in the binary doc format.
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
	Attributes      OdinDocArray
	GroupedEntities OdinDocArray
	WhereClauses    OdinDocArray
}

// OdinDocPkgFlags for package metadata.
type OdinDocPkgFlags uint32

const (
	OdinDocPkgFlagBuiltin OdinDocPkgFlags = 1 << 0
	OdinDocPkgFlagRuntime OdinDocPkgFlags = 1 << 1
	OdinDocPkgFlagInit    OdinDocPkgFlags = 1 << 2
)

// OdinDocScopeEntry maps a scope name to an entity index.
type OdinDocScopeEntry struct {
	Name   OdinDocString
	Entity OdinDocEntityIndex
}

// OdinDocPkg represents a package entry in the binary doc format.
type OdinDocPkg struct {
	Fullpath OdinDocString
	Name     OdinDocString
	Flags    uint32
	Docs     OdinDocString
	Files    OdinDocArray
	Entries  OdinDocArray
}

// OdinDocHeader is the top-level structure of the binary doc format.
type OdinDocHeader struct {
	Base     OdinDocHeaderBase
	Files    OdinDocArray
	Pkgs     OdinDocArray
	Entities OdinDocArray
	Types    OdinDocArray
}

// Version constants
const (
	OdinDocVersionMajor uint8 = 0
	OdinDocVersionMinor uint8 = 3
	OdinDocVersionPatch uint8 = 2
)

var (
	// MagicBytes is the magic identifier for odin-doc files.
	MagicBytes = [8]byte{'o', 'd', 'i', 'n', 'd', 'o', 'c', 0}

	// OdinDocVersion is the current binary format version.
	OdinDocVersion = OdinDocVersionType{
		Major: OdinDocVersionMajor,
		Minor: OdinDocVersionMinor,
		Patch: OdinDocVersionPatch,
	}
)

// OdinDocTypeKindStrings maps type kinds to strings.
var OdinDocTypeKindStrings = map[OdinDocTypeKind]string{
	OdinDocTypeInvalid:               "Invalid",
	OdinDocTypeBasic:                 "Basic",
	OdinDocTypeNamed:                 "Named",
	OdinDocTypeGeneric:               "Generic",
	OdinDocTypePointer:               "Pointer",
	OdinDocTypeArray:                 "Array",
	OdinDocTypeEnumeratedArray:       "EnumeratedArray",
	OdinDocTypeSlice:                 "Slice",
	OdinDocTypeDynamicArray:          "DynamicArray",
	OdinDocTypeMap:                   "Map",
	OdinDocTypeStruct:                "Struct",
	OdinDocTypeUnion:                 "Union",
	OdinDocTypeEnum:                  "Enum",
	OdinDocTypeTuple:                 "Tuple",
	OdinDocTypeProc:                  "Proc",
	OdinDocTypeBitSet:                "BitSet",
	OdinDocTypeSimdVector:            "SimdVector",
	OdinDocTypeSOAStructFixed:        "SOAStructFixed",
	OdinDocTypeSOAStructSlice:        "SOAStructSlice",
	OdinDocTypeSOAStructDynamic:      "SOAStructDynamic",
	OdinDocTypeMultiPointer:          "MultiPointer",
	OdinDocTypeMatrix:                "Matrix",
	OdinDocTypeSoaPointer:            "SoaPointer",
	OdinDocTypeBitField:              "BitField",
	OdinDocTypeFixedCapacityDynArray: "FixedCapacityDynamicArray",
}

// OdinDocEntityKindStrings maps entity kinds to strings.
var OdinDocEntityKindStrings = map[OdinDocEntityKind]string{
	OdinDocEntityInvalid:     "Invalid",
	OdinDocEntityConstant:    "Constant",
	OdinDocEntityVariable:    "Variable",
	OdinDocEntityTypeName:    "TypeName",
	OdinDocEntityProcedure:   "Procedure",
	OdinDocEntityProcGroup:   "ProcGroup",
	OdinDocEntityImportName:  "ImportName",
	OdinDocEntityLibraryName: "LibraryName",
	OdinDocEntityBuiltin:     "Builtin",
}

// alignOffset aligns an offset to the given boundary.
func alignOffset(offset int, alignment int) int {
	return (offset + alignment - 1) &^ (alignment - 1)
}

// fnv constants
const (
	fnvOffsetBasis uint32 = 0x811c9dc5
	fnvPrime       uint32 = 0x01000193
)

// calcFNV1a computes the FNV-1a hash of data.
func calcFNV1a(data []byte) uint32 {
	h := uint32(fnvOffsetBasis)
	for _, b := range data {
		h = (h ^ uint32(b)) * fnvPrime
	}
	return h
}
