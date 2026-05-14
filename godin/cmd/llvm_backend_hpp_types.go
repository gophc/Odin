package cmd

import "sync/atomic"

type lbValue struct {
	Value LLVMValueRef
	Type  *Type
}

type lbAddrKind int

const (
	lbAddr_Default     lbAddrKind = iota
	lbAddr_Map
	lbAddr_Context
	lbAddr_SoaVariable
	lbAddr_Swizzle
	lbAddr_SwizzleLarge
	lbAddr_BitField
)

type lbAddr struct {
	Kind lbAddrKind
	Addr lbValue

	MapKey    *lbValue
	MapType   *Type
	MapResult *Type

	CtxSel Selection

	SoaIndex     lbValue
	SoaIndexExpr *Ast

	IndexSetIndex lbValue
	IndexSetNode  *Ast

	SwizzleType    *Type
	SwizzleCount   u8
	SwizzleIndices [4]u8

	SwizzleLargeType    *Type
	SwizzleLargeIndices []i32

	BitFieldType  *Type
	BitFieldOffset i64
	BitFieldSize   i64
}

type lbIncompleteDebugType struct {
	Type     *Type
	Metadata LLVMMetadataRef
}

type lbStructFieldRemapping = []i32

type lbFunctionPassManagerKind int

const (
	lbFunctionPassManager_default                lbFunctionPassManagerKind = iota
	lbFunctionPassManager_default_without_memcpy
	lbFunctionPassManager_none
	lbFunctionPassManager_COUNT
)

type lbPadType struct {
	Padding      i64
	PaddingAlign i64
	Type         LLVMTypeRef
}

type lbModule struct {
	Mod                              LLVMModuleRef
	Ctx                              LLVMContextRef
	Checker                          *Checker
	Gen                              *lbGenerator
	TargetMachine                    LLVMTargetMachineRef
	PolymorphicModule                *lbModule
	Info                             *CheckerInfo
	Pkg                              *AstPackage
	File                             *AstFile
	ModuleName                       string
	Types                            PtrMap[uint64, LLVMTypeRef]
	StructFieldRemapping              PtrMap[uintptr, lbStructFieldRemapping]
	FuncRawTypes                     PtrMap[uint64, LLVMTypeRef]
	TypesMutex                       RecursiveMutex
	FuncRawTypesMutex                RecursiveMutex
	InternalTypeLevel                i32
	ValuesMutex                      RwMutex
	GlobalArrayIndex                 atomic.Uint32
	Values                           PtrMap[*Entity, lbValue]
	SoaValues                        PtrMap[*Entity, lbAddr]
	Members                          StringMap[lbValue]
	Procedures                       StringMap[*lbProcedure]
	ProcedureValues                  PtrMap[LLVMValueRef, *Entity]
	MissingProceduresToCheck         MPSCQueue[*lbProcedure]
	ConstStrings                     StringMap[LLVMValueRef]
	ConstString16s                   String16Map[LLVMValueRef]
	FunctionTypeMap                  PtrMap[uint64, *lbFunctionType]
	GenProcs                         StringMap[*lbProcedure]
	ProceduresToGenerate             MPSCQueue[*lbProcedure]
	GlobalProceduresToCreate         Array[*Entity]
	GlobalTypesToCreate              Array[*Entity]
	GeneratedProceduresMutex         BlockingMutex
	GeneratedProcedures              Array[*lbProcedure]
	CurrProcedure                    *lbProcedure
	ConstDummyBuilder                LLVMBuilderRef
	DebugBuilder                     LLVMDIBuilderRef
	DebugCompileUnit                 LLVMMetadataRef
	DebugValuesMutex                 RecursiveMutex
	DebugValues                      PtrMap[uintptr, LLVMMetadataRef]
	ObjCClasses                      StringMap[lbAddr]
	ObjCSelectors                    StringMap[lbAddr]
	ObjCIvars                        StringMap[lbAddr]
	ObjCNextBlockID                  isize
	MapCellInfoMap                   PtrMap[uint64, lbAddr]
	MapInfoMap                       PtrMap[uint64, lbAddr]
	ExactValueCompoundLiteralAddrMap PtrMap[*Ast, lbAddr]
	FunctionPassManagers             [lbFunctionPassManager_COUNT]LLVMPassManagerRef
	PadTypesMutex                    BlockingMutex
	PadTypes                         Array[lbPadType]
}

type lbEntityCorrection struct {
	OtherModule *lbModule
	E           *Entity
	Cname       string
}

type lbObjCGlobal struct {
	Module        *lbModule
	GlobalName    gbString
	Name          String
	Type          *Type
	ClassImplType *Type
}

type lbGenerator struct {
	LinkerData
	Info                      *CheckerInfo
	Modules                   PtrMap[uintptr, *lbModule]
	ModulesThroughCtx         PtrMap[LLVMContextRef, *lbModule]
	DefaultModule             lbModule
	EqualModule               *lbModule
	UsedModuleCount           isize
	StartupRuntime            *lbProcedure
	CleanupRuntime            *lbProcedure
	ObjCNames                 *lbProcedure
	EntitiesToCorrectLinkage  MPSCQueue[lbEntityCorrection]
	ObjCSelectors             MPSCQueue[lbObjCGlobal]
	ObjCClasses               MPSCQueue[lbObjCGlobal]
	ObjCIvars                 MPSCQueue[lbObjCGlobal]
	RadDebugSectionStrings    MPSCQueue[String]
}

type lbBlock struct {
	Block      LLVMBasicBlockRef
	Scope      *Scope
	ScopeIndex isize
	Appended   bool
	Preds      Array[*lbBlock]
	Succs      Array[*lbBlock]
}

type lbBranchBlocks struct {
	Label    *Ast
	Break_   *lbBlock
	Continue *lbBlock
}

type lbContextData struct {
	Ctx        lbAddr
	ScopeIndex isize
	Uses       isize
}

type lbParamPassKind int

const (
	lbParamPass_Value    lbParamPassKind = iota
	lbParamPass_Pointer
	lbParamPass_Integer
	lbParamPass_ConstRef
	lbParamPass_BitCast
	lbParamPass_Tuple
)

type lbDeferExitKind int

const (
	lbDeferExit_Default lbDeferExitKind = iota
	lbDeferExit_Return
	lbDeferExit_Branch
)

type lbDeferKind int

const (
	lbDefer_Node lbDeferKind = iota
	lbDefer_Proc
)

type lbDefer struct {
	Kind              lbDeferKind
	ScopeIndex        isize
	ContextStackCount isize
	Block             *lbBlock
	Pos               TokenPos
	Stmt              *Ast
	ProcDeferred      lbValue
	ProcResultAsArgs  Array[lbValue]
}

type lbTargetList struct {
	Prev        *lbTargetList
	IsBlock     bool
	Break_      *lbBlock
	Continue_   *lbBlock
	Fallthrough *lbBlock
}

type lbTupleFix struct {
	Values []lbValue
}

const (
	lbProcedureFlag_WithoutMemcpyPass uint32 = 1 << 0
	lbProcedureFlag_DebugAllocaCopy   uint32 = 1 << 1
)

type lbVariadicReuseSlices struct {
	SliceType *Type
	SliceAddr lbAddr
}

type lbGlobalVariable struct {
	Var           lbValue
	Init          lbValue
	Decl          *DeclInfo
	IsInitialized bool
}

type lbProcedure struct {
	Flags                         uint32
	StateFlags                    uint16
	Parent                        *lbProcedure
	Children                      Array[*lbProcedure]
	Entity                        *Entity
	Module                        *lbModule
	Name                          String
	Type                          *Type
	TypeExpr                      *Ast
	Body                          *Ast
	Tags                          uint64
	Inlining                      ProcInlining
	Tailing                       ProcTailing
	IsForeign                     bool
	IsExport                      bool
	IsEntryPoint                  bool
	IsStartup                     bool
	AbiFunctionType               *lbFunctionType
	Value                         LLVMValueRef
	Builder                       LLVMBuilderRef
	IsDone                        atomic.Bool
	ReturnPtr                     lbAddr
	SretRvoEntity                 *Entity
	DeferStmts                    Array[lbDefer]
	Blocks                        Array[*lbBlock]
	BranchBlocks                  Array[lbBranchBlocks]
	CurrScope                     *Scope
	ScopeIndex                    i32
	DeclBlock                     *lbBlock
	EntryBlock                    *lbBlock
	CurrBlock                     *lbBlock
	TargetList                    *lbTargetList
	DirectParameters              PtrMap[*Entity, lbValue]
	InMultiAssignment             bool
	RawInputParameters            Array[LLVMValueRef]
	GlobalGeneratedIndex          uint32
	UsesBranchLocation            bool
	BranchLocationPos             TokenPos
	CurrTokenPos                  TokenPos
	VariadicReuses                Array[lbVariadicReuseSlices]
	VariadicReuseBaseArrayPtr     lbAddr
	TempCalleeReturnStructMemory  LLVMValueRef
	CurrStmt                      *Ast
	ScopeStack                    Array[*Scope]
	ContextStack                  Array[lbContextData]
	DebugInfo                     LLVMMetadataRef
	SelectorValues                PtrMap[*Ast, lbValue]
	SelectorAddr                  PtrMap[*Ast, lbAddr]
	TupleFixMap                   PtrMap[LLVMValueRef, lbTupleFix]
	AsanStackLocals               Array[lbValue]
	GenerateBody                  func(m *lbModule, p *lbProcedure)
	GlobalVariables               Array[*lbGlobalVariable]
	ObjCNames                     *lbProcedure
	InternalGenType               *Type
}

type lbConstContext struct {
	AllowLocal   bool
	IsRodata     bool
	LinkSection  String
}

var LB_CONST_CONTEXT_DEFAULT = lbConstContext{AllowLocal: true, IsRodata: false}
var LB_CONST_CONTEXT_DEFAULT_ALLOW_LOCAL = lbConstContext{AllowLocal: true, IsRodata: false}
var LB_CONST_CONTEXT_DEFAULT_NO_LOCAL = lbConstContext{AllowLocal: false, IsRodata: false}

type lbCallingConventionKind uint32

const (
	lbCallingConvention_C                    lbCallingConventionKind = 0
	lbCallingConvention_Fast                 lbCallingConventionKind = 8
	lbCallingConvention_Cold                 lbCallingConventionKind = 9
	lbCallingConvention_GHC                  lbCallingConventionKind = 10
	lbCallingConvention_HiPE                 lbCallingConventionKind = 11
	lbCallingConvention_WebKit_JS            lbCallingConventionKind = 12
	lbCallingConvention_AnyReg               lbCallingConventionKind = 13
	lbCallingConvention_PreserveMost         lbCallingConventionKind = 14
	lbCallingConvention_PreserveAll          lbCallingConventionKind = 15
	lbCallingConvention_Swift                lbCallingConventionKind = 16
	lbCallingConvention_CXX_FAST_TLS         lbCallingConventionKind = 17
	lbCallingConvention_PreserveNone         lbCallingConventionKind = 21
	lbCallingConvention_FirstTargetCC        lbCallingConventionKind = 64
	lbCallingConvention_X86_StdCall          lbCallingConventionKind = 64
	lbCallingConvention_X86_FastCall         lbCallingConventionKind = 65
	lbCallingConvention_ARM_APCS             lbCallingConventionKind = 66
	lbCallingConvention_ARM_AAPCS            lbCallingConventionKind = 67
	lbCallingConvention_ARM_AAPCS_VFP        lbCallingConventionKind = 68
	lbCallingConvention_MSP430_INTR          lbCallingConventionKind = 69
	lbCallingConvention_X86_ThisCall         lbCallingConventionKind = 70
	lbCallingConvention_PTX_Kernel           lbCallingConventionKind = 71
	lbCallingConvention_PTX_Device           lbCallingConventionKind = 72
	lbCallingConvention_SPIR_FUNC            lbCallingConventionKind = 75
	lbCallingConvention_SPIR_KERNEL          lbCallingConventionKind = 76
	lbCallingConvention_Intel_OCL_BI         lbCallingConventionKind = 77
	lbCallingConvention_X86_64_SysV          lbCallingConventionKind = 78
	lbCallingConvention_Win64                lbCallingConventionKind = 79
	lbCallingConvention_X86_VectorCall       lbCallingConventionKind = 80
	lbCallingConvention_HHVM                 lbCallingConventionKind = 81
	lbCallingConvention_HHVM_C               lbCallingConventionKind = 82
	lbCallingConvention_X86_INTR             lbCallingConventionKind = 83
	lbCallingConvention_AVR_INTR             lbCallingConventionKind = 84
	lbCallingConvention_AVR_SIGNAL           lbCallingConventionKind = 85
	lbCallingConvention_AVR_BUILTIN          lbCallingConventionKind = 86
	lbCallingConvention_AMDGPU_VS            lbCallingConventionKind = 87
	lbCallingConvention_AMDGPU_GS            lbCallingConventionKind = 88
	lbCallingConvention_AMDGPU_PS            lbCallingConventionKind = 89
	lbCallingConvention_AMDGPU_CS            lbCallingConventionKind = 90
	lbCallingConvention_AMDGPU_KERNEL        lbCallingConventionKind = 91
	lbCallingConvention_X86_RegCall          lbCallingConventionKind = 92
	lbCallingConvention_AMDGPU_HS            lbCallingConventionKind = 93
	lbCallingConvention_MSP430_BUILTIN       lbCallingConventionKind = 94
	lbCallingConvention_AMDGPU_LS            lbCallingConventionKind = 95
	lbCallingConvention_AMDGPU_ES            lbCallingConventionKind = 96
	lbCallingConvention_AArch64_VectorCall        lbCallingConventionKind = 97
	lbCallingConvention_AArch64_SVE_VectorCall    lbCallingConventionKind = 98
	lbCallingConvention_WASM_EmscriptenInvoke     lbCallingConventionKind = 99
	lbCallingConvention_MaxID               lbCallingConventionKind = 1023
)

var lb_calling_convention_map = [ProcCC_MAX]lbCallingConventionKind{
	lbCallingConvention_C,
	lbCallingConvention_C,
	lbCallingConvention_C,
	lbCallingConvention_C,
	lbCallingConvention_X86_StdCall,
	lbCallingConvention_X86_FastCall,
	lbCallingConvention_C,
	lbCallingConvention_C,
	lbCallingConvention_C,
	lbCallingConvention_Win64,
	lbCallingConvention_X86_64_SysV,
	lbCallingConvention_PreserveNone,
	lbCallingConvention_PreserveMost,
	lbCallingConvention_PreserveAll,
}

const (
	LLVMDWARFTypeEncoding_Address        = 1
	LLVMDWARFTypeEncoding_Boolean        = 2
	LLVMDWARFTypeEncoding_ComplexFloat   = 3
	LLVMDWARFTypeEncoding_Float          = 4
	LLVMDWARFTypeEncoding_Signed         = 5
	LLVMDWARFTypeEncoding_SignedChar     = 6
	LLVMDWARFTypeEncoding_Unsigned       = 7
	LLVMDWARFTypeEncoding_UnsignedChar   = 8
	LLVMDWARFTypeEncoding_ImaginaryFloat = 9
	LLVMDWARFTypeEncoding_PackedDecimal  = 10
	LLVMDWARFTypeEncoding_NumericString  = 11
	LLVMDWARFTypeEncoding_Edited         = 12
	LLVMDWARFTypeEncoding_SignedFixed    = 13
	LLVMDWARFTypeEncoding_UnsignedFixed  = 14
	LLVMDWARFTypeEncoding_DecimalFloat   = 15
	LLVMDWARFTypeEncoding_Utf            = 16
	LLVMDWARFTypeEncoding_LoUser         = 128
	LLVMDWARFTypeEncoding_HiUser         = 255
)

const (
	DW_TAG_array_type       = 1
	DW_TAG_enumeration_type = 4
	DW_TAG_structure_type   = 19
	DW_TAG_union_type       = 23
	DW_TAG_vector_type      = 259
	DW_TAG_subroutine_type  = 21
	DW_TAG_inheritance      = 28
)

const (
	LLVMAttributeIndex_ReturnIndex   = 0
	LLVMAttributeIndex_FunctionIndex = ^uint32(0)
	LLVMAttributeIndex_FirstArgIndex = 1
)

var llvm_linkage_strings = [...]string{
	"external linkage",
	"available externally linkage",
	"link once any linkage",
	"link once odr linkage",
	"link once odr auto hide linkage",
	"weak any linkage",
	"weak odr linkage",
	"appending linkage",
	"internal linkage",
	"private linkage",
	"dllimport linkage",
	"dllexport linkage",
	"external weak linkage",
	"ghost linkage",
	"common linkage",
	"linker private linkage",
	"linker private weak linkage",
}
