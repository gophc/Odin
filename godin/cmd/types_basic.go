// Depends on: common.odin (String, isize, i64, u8, u32, ExactValue, Entity, Scope, Ast, InternedString, BlockingMutex, RecursiveMutex, Wait_Signal, ProcCallingConvention, UnionTypeKind, Slice, TokenKind, Token)
package cmd

type BasicKind int

const (
	BasicInvalid BasicKind = iota
	BasicLLVMBool
	BasicBool
	BasicB8
	BasicB16
	BasicB32
	BasicB64
	BasicI8
	BasicU8
	BasicI16
	BasicU16
	BasicI32
	BasicU32
	BasicI64
	BasicU64
	BasicI128
	BasicU128
	BasicRune
	BasicF16
	BasicF32
	BasicF64
	BasicComplex32
	BasicComplex64
	BasicComplex128
	BasicQuaternion64
	BasicQuaternion128
	BasicQuaternion256
	BasicInt
	BasicUint
	BasicUintptr
	BasicRawptr
	BasicString
	BasicCstring
	BasicString16
	BasicCstring16
	BasicAny
	BasicTypeid
	BasicI16le
	BasicU16le
	BasicI32le
	BasicU32le
	BasicI64le
	BasicU64le
	BasicI128le
	BasicU128le
	BasicI16be
	BasicU16be
	BasicI32be
	BasicU32be
	BasicI64be
	BasicU64be
	BasicI128be
	BasicU128be
	BasicF16le
	BasicF32le
	BasicF64le
	BasicF16be
	BasicF32be
	BasicF64be
	BasicUntypedBool
	BasicUntypedInteger
	BasicUntypedFloat
	BasicUntypedComplex
	BasicUntypedQuaternion
	BasicUntypedString
	BasicUntypedRune
	BasicUntypedNil
	BasicUntypedUninit
	BasicCOUNT

	BasicByte = BasicU8
)

const (
	BasicFlagBoolean    uint32 = 1 << 0
	BasicFlagInteger    uint32 = 1 << 1
	BasicFlagUnsigned   uint32 = 1 << 2
	BasicFlagFloat      uint32 = 1 << 3
	BasicFlagComplex    uint32 = 1 << 4
	BasicFlagQuaternion uint32 = 1 << 5
	BasicFlagPointer    uint32 = 1 << 6
	BasicFlagString     uint32 = 1 << 7
	BasicFlagRune       uint32 = 1 << 8
	BasicFlagUntyped    uint32 = 1 << 9
	BasicFlagLLVM       uint32 = 1 << 11
	BasicFlagEndianLittle uint32 = 1 << 13
	BasicFlagEndianBig    uint32 = 1 << 14

	BasicFlagNumeric        = BasicFlagInteger | BasicFlagFloat | BasicFlagComplex | BasicFlagQuaternion
	BasicFlagOrdered        = BasicFlagInteger | BasicFlagFloat | BasicFlagString | BasicFlagPointer | BasicFlagRune
	BasicFlagOrderedNumeric = BasicFlagInteger | BasicFlagFloat | BasicFlagRune
	BasicFlagConstantType   = BasicFlagBoolean | BasicFlagNumeric | BasicFlagString | BasicFlagPointer | BasicFlagRune
	BasicFlagSimpleCompare  = BasicFlagBoolean | BasicFlagInteger | BasicFlagPointer | BasicFlagRune
)

type BasicType struct {
	Kind  BasicKind
	Flags uint32
	Size  int64
	Name  string
}

type StructSoaKind uint8

const (
	StructSoaNone    StructSoaKind = 0
	StructSoaFixed   StructSoaKind = 1
	StructSoaSlice   StructSoaKind = 2
	StructSoaDynamic StructSoaKind = 3
)

type TypeStruct struct {
	Fields                 []*Entity
	Tags                   []string
	Offsets                []int64
	Node                   *Ast
	Scope                  *Scope
	CustomAlign            int64
	CustomMinFieldAlign    int64
	CustomMaxFieldAlign    int64
	PolymorphicParams      *Type
	PolymorphicParent      *Type
	PolymorphicWaitSignal  WaitSignal
	SoaElem                *Type
	SoaCount               int32
	SoaKind                StructSoaKind
	FieldsWaitSignal       WaitSignal
	SoaMutex               BlockingMutex
	OffsetMutex            BlockingMutex
	IsPolymorphic          bool
	AreOffsetsSet          bool
	IsPacked               bool
	IsRawUnion             bool
	IsAllOrNone            bool
	IsSimple               bool
	IsPolySpecialized      bool
	AreOffsetsBeingProcessed bool
}

type TypeUnion struct {
	Variants             []*Type
	Node                 *Ast
	Scope                *Scope
	VariantBlockSize     int64
	CustomAlign          int64
	PolymorphicParams    *Type
	PolymorphicParent    *Type
	PolymorphicWaitSignal WaitSignal
	TagSize              int16
	IsPolymorphic        bool
	IsPolySpecialized    bool
	Kind                 UnionTypeKind
}

type TypeProc struct {
	Node                   *Ast
	Scope                  *Scope
	Params                 *Type
	Results                *Type
	ParamCount             int32
	ResultCount            int32
	SpecializationCount    isize
	CallingConvention      ProcCallingConvention
	VariadicIndex          int32
	RequireTargetFeature   string
	EnableTargetFeature    string
	Variadic               bool
	RequireResults         bool
	CVararg                bool
	IsPolymorphic          bool
	IsPolySpecialized      bool
	HasNamedResults        bool
	Diverging              bool
	ReturnByPointer        bool
	OptionalOK             bool
}

type TypeNamed struct {
	Name         string
	Base         *Type
	TypeName     *Entity
	GenTypesDataMutex BlockingMutex
	GenTypesData *GenTypesData
}

type TypeKind int

const (
	TypeInvalid TypeKind = iota
	TypeBasic
	TypeNamed
	TypeGeneric
	TypePointer
	TypeMultiPointer
	TypeArray
	TypeEnumeratedArray
	TypeSlice
	TypeDynamicArray
	TypeFixedCapacityDynamicArray
	TypeMap
	TypeStruct
	TypeUnion
	TypeEnum
	TypeTuple
	TypeProc
	TypeBitSet
	TypeSimdVector
	TypeMatrix
	TypeBitField
	TypeSoaPointer
	TypeCount
)

var TypeStrings = []string{
	"Invalid", "Basic", "Named", "Generic", "Pointer", "MultiPointer",
	"Array", "EnumeratedArray", "Slice", "DynamicArray",
	"FixedCapacityDynamicArray", "Map", "Struct", "Union", "Enum",
	"Tuple", "Proc", "BitSet", "SimdVector", "Matrix", "BitField",
	"SoaPointer",
}

type TypeGeneric struct {
	ID           int64
	Name         string
	InternedName InternedString
	Specialized  *Type
	Scope        *Scope
	Entity       *Entity
}

type TypePointer struct {
	Elem *Type
}

type TypeMultiPointer struct {
	Elem *Type
}

type TypeArray struct {
	Elem         *Type
	Count        int64
	GenericCount *Type
}

type TypeEnumeratedArray struct {
	Elem      *Type
	Index     *Type
	MinValue  *ExactValue
	MaxValue  *ExactValue
	Count     int64
	Op        TokenKind
	IsSparse  bool
}

type TypeSlice struct {
	Elem *Type
}

type TypeDynamicArray struct {
	Elem *Type
}

type TypeFixedCapacityDynamicArray struct {
	Capacity        int64
	GenericCapacity *Type
	Elem            *Type
	PaddingNeeded   int64
}

type TypeMap struct {
	Key                *Type
	Value              *Type
	LookupResultType   *Type
	DebugMetadataType  *Type
}

type TypeEnum struct {
	Fields         []*Entity
	Node           *Ast
	Scope          *Scope
	BaseType       *Type
	MinValue       *ExactValue
	MaxValue       *ExactValue
	MinValueIndex  isize
	MaxValueIndex  isize
}

type TypeTuple struct {
	Variables                  []*Entity
	Offsets                    []int64
	Mutex                      BlockingMutex
	AreOffsetsBeingProcessed   bool
	AreOffsetsSet              bool
	IsPacked                   bool
}

type TypeBitSet struct {
	Elem       *Type
	Underlying *Type
	Lower      int64
	Upper      int64
	Node       *Ast
}

type TypeSimdVector struct {
	Count         int64
	Elem          *Type
	GenericCount  *Type
}

type TypeMatrix struct {
	Elem                *Type
	RowCount            int64
	ColumnCount         int64
	GenericRowCount     *Type
	GenericColumnCount  *Type
	StrideInBytes       int64
	IsRowMajor          bool
}

type TypeBitField struct {
	Scope        *Scope
	BackingType  *Type
	Fields       []*Entity
	Tags         []string
	BitSizes     []uint8
	BitOffsets   []int64
	Node         *Ast
}

type TypeSoaPointer struct {
	Elem *Type
}

const (
	TypeFlagPolymorphic                    uint32 = 1 << 1
	TypeFlagPolySpecialized                uint32 = 1 << 2
	TypeFlagInProcessOfCheckingPolymorphic uint32 = 1 << 3
)

type Type struct {
	Kind         TypeKind
	Basic        BasicType
	Named        TypeNamed
	Generic      TypeGeneric
	Pointer      TypePointer
	MultiPointer TypeMultiPointer
	Array        TypeArray
	EnumeratedArray TypeEnumeratedArray
	Slice        TypeSlice
	DynamicArray TypeDynamicArray
	FixedCapacityDynamicArray TypeFixedCapacityDynamicArray
	Map          TypeMap
	Struct       TypeStruct
	Union        TypeUnion
	Enum         TypeEnum
	Tuple        TypeTuple
	Proc         TypeProc
	BitSet       TypeBitSet
	SimdVector   TypeSimdVector
	Matrix       TypeMatrix
	BitField     TypeBitField
	SoaPointer   TypeSoaPointer
	CachedSize   int64
	CachedAlign  int64
	CanonicalHash uint64
	Flags        uint32
	Failure      bool
}

type TypeidKind uint8

const (
	TypeidInvalid TypeidKind = iota
	TypeidInteger
	TypeidRune
	TypeidFloat
	TypeidComplex
	TypeidQuaternion
	TypeidString
	TypeidBoolean
	TypeidAny
	TypeidTypeId
	TypeidPointer
	TypeidMultiPointer
	TypeidProcedure
	TypeidArray
	TypeidEnumeratedArray
	TypeidDynamicArray
	TypeidSlice
	TypeidTuple
	TypeidStruct
	TypeidUnion
	TypeidEnum
	TypeidMap
	TypeidBitSet
	TypeidSimdVector
	TypeidMatrix
	TypeidSoaPointer
	TypeidBitField
	TypeidFixedCapacityDynamicArray
	TypeidCOUNT
)

const (
	TypeInfoFlagComparable     uint32 = 1 << 0
	TypeInfoFlagSimpleCompare  uint32 = 1 << 1
)

const (
	MatrixElementCountMin = 1
	MatrixElementCountMax = 64
	MatrixElementMaxSize  = MatrixElementCountMax * (2 * 8)
	SimdElementCountMin   = 1
	SimdElementCountMax   = 64
)

type Selection struct {
	Entity        *Entity
	Index         []int32
	Indirect      bool
	SwizzleCount  uint8
	SwizzleIndices uint8
	IsBitField    bool
	PseudoField   bool
}

func makeSelection(entity *Entity, index []int32, indirect bool) Selection {
	return Selection{Entity: entity, Index: index, Indirect: indirect}
}

func selectionAddIndex(s *Selection, index isize) {
	if s.Index == nil {
		s.Index = make([]int32, 0)
	}
	s.Index = append(s.Index, int32(index))
}

func selectionCombine(lhs, rhs Selection) Selection {
	ns := lhs
	ns.Indirect = lhs.Indirect || rhs.Indirect
	ns.Index = make([]int32, len(lhs.Index)+len(rhs.Index))
	copy(ns.Index, lhs.Index)
	copy(ns.Index[len(lhs.Index):], rhs.Index)
	return ns
}

func subSelection(sel Selection, offset isize) Selection {
	var res Selection
	if offset < isize(len(sel.Index)) {
		res.Index = sel.Index[offset:]
	} else {
		res.Index = nil
	}
	return res
}

func trimSelection(sel Selection) Selection {
	var res Selection
	if len(sel.Index) > 0 {
		res.Index = sel.Index[:len(sel.Index)-1]
	} else {
		res.Index = nil
	}
	return res
}

var basicTypes []Type
var emptySelection Selection

// t_* type globals
var (
	tInvalid                      *Type
	tLLVMBool                     *Type
	tBool                         *Type
	tI8, tU8, tI16, tU16         *Type
	tI32, tU32, tI64, tU64       *Type
	tI128, tU128                  *Type
	tRune                         *Type
	tF16, tF32, tF64             *Type
	tF16be, tF32be, tF64be       *Type
	tF16le, tF32le, tF64le       *Type
	tComplex32, tComplex64, tComplex128 *Type
	tQuaternion64, tQuaternion128, tQuaternion256 *Type
	tInt, tUint, tUintptr        *Type
	tRawptr                       *Type
	tString, tCstring            *Type
	tString16, tCstring16        *Type
	tAny, tTypeid                *Type
	tI16le, tU16le, tI32le, tU32le *Type
	tI64le, tU64le, tI128le, tU128le *Type
	tI16be, tU16be, tI32be, tU32be *Type
	tI64be, tU64be, tI128be, tU128be *Type
	tUntypedBool, tUntypedInteger  *Type
	tUntypedFloat, tUntypedComplex *Type
	tUntypedQuaternion, tUntypedString *Type
	tUntypedRune, tUntypedNil, tUntypedUninit *Type
	tU8Ptr, tU8MultiPtr          *Type
	tU16Ptr, tU16MultiPtr        *Type
	tIntPtr, tI64Ptr, tF64Ptr    *Type
	tU8Slice, tStringSlice        *Type
	tTypeInfo, tTypeInfoEnumValue *Type
	tTypeInfoPtr, tTypeInfoEnumValuePtr *Type
	tTypeInfoStringEncodingKind   *Type
	tTypeInfoNamed                *Type
	tTypeInfoInteger              *Type
	tTypeInfoRune, tTypeInfoFloat *Type
	tTypeInfoComplex, tTypeInfoQuaternion *Type
	tTypeInfoAny, tTypeInfoTypeid *Type
	tTypeInfoString, tTypeInfoBoolean *Type
	tTypeInfoPointer, tTypeInfoMultiPointer *Type
	tTypeInfoProcedure            *Type
	tTypeInfoArray, tTypeInfoEnumeratedArray *Type
	tTypeInfoDynamicArray, tTypeInfoSlice *Type
	tTypeInfoParameters, tTypeInfoStruct *Type
	tTypeInfoUnion, tTypeInfoEnum *Type
	tTypeInfoMap, tTypeInfoBitSet *Type
	tTypeInfoSimdVector, tTypeInfoMatrix *Type
	tTypeInfoSoaPointer, tTypeInfoBitField *Type
	tTypeInfoFixedCapacityDynamicArray *Type
	tTypeInfoNamedPtr, tTypeInfoIntegerPtr *Type
	tTypeInfoRunePtr, tTypeInfoFloatPtr *Type
	tTypeInfoComplexPtr, tTypeInfoQuaternionPtr *Type
	tTypeInfoAnyPtr, tTypeInfoTypeidPtr *Type
	tTypeInfoStringPtr, tTypeInfoBooleanPtr *Type
	tTypeInfoPointerPtr, tTypeInfoMultiPointerPtr *Type
	tTypeInfoProcedurePtr         *Type
	tTypeInfoArrayPtr, tTypeInfoEnumeratedArrayPtr *Type
	tTypeInfoDynamicArrayPtr, tTypeInfoSlicePtr *Type
	tTypeInfoParametersPtr, tTypeInfoStructPtr *Type
	tTypeInfoUnionPtr, tTypeInfoEnumPtr *Type
	tTypeInfoMapPtr, tTypeInfoBitSetPtr *Type
	tTypeInfoSimdVectorPtr, tTypeInfoMatrixPtr *Type
	tTypeInfoSoaPointerPtr, tTypeInfoBitFieldPtr *Type
	tTypeInfoFixedCapacityDynamicArrayPtr *Type
	tAllocator, tAllocatorPtr     *Type
	tContext, tContextPtr         *Type
	tAllocatorError               *Type
	tSourceCodeLocation, tSourceCodeLocationPtr *Type
	tLoadDirectoryFile, tLoadDirectoryFilePtr *Type
	tLoadDirectoryFileSlice       *Type
	tMapInfo, tMapCellInfo        *Type
	tRawMap                       *Type
	tMapInfoPtr, tMapCellInfoPtr *Type
	tRawMapPtr                    *Type
	tEqualProc, tHasherProc       *Type
	tMapGetProc, tMapSetProc      *Type
	tObjcObject, tObjcSelector    *Type
	tObjcClass, tObjcIvar, tObjcSuper *Type
	tObjcSuperPtr, tObjcID        *Type
	tObjcSEL, tObjcClass2         *Type
	tObjcIvar2, tObjcInstancetype *Type
	tCValist, tCValistPtr         *Type
	tAtomicMemoryOrder            *Type
)

type OdinAtomicMemoryOrder int32

const (
	OdinAtomicMemoryOrderRelaxed OdinAtomicMemoryOrder = iota
	OdinAtomicMemoryOrderConsume
	OdinAtomicMemoryOrderAcquire
	OdinAtomicMemoryOrderRelease
	OdinAtomicMemoryOrderAcqRel
	OdinAtomicMemoryOrderSeqCst
	OdinAtomicMemoryOrderCOUNT
)

var OdinAtomicMemoryOrderStrings = [...]string{
	"Relaxed", "Consume", "Acquire", "Release", "Acq_Rel", "Seq_Cst",
}

var gTypeMutex RecursiveMutex

// TypePath - cycle detection for type graph traversal
type TypePath struct {
	Mutex   RecursiveMutex
	Path    []*Entity
	Failure bool
}

type TypeEndianKind int

const (
	TypeEndianPlatform TypeEndianKind = iota
	TypeEndianLittle
	TypeEndianBig
)

type ProcTypeOverloadKind int

const (
	ProcOverloadIdentical ProcTypeOverloadKind = iota
	ProcOverloadCallingConvention
	ProcOverloadParamCount
	ProcOverloadParamVariadic
	ProcOverloadParamTypes
	ProcOverloadResultCount
	ProcOverloadResultTypes
	ProcOverloadPolymorphic
	ProcOverloadTargetFeatures
	ProcOverloadNotProcedure
)

type GenTypesData struct{}
