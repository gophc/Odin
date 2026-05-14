package cmd

type LLVMMemoryBufferRef uintptr
type LLVMContextRef uintptr
type LLVMModuleRef uintptr
type LLVMTypeRef uintptr
type LLVMValueRef uintptr
type LLVMBasicBlockRef uintptr
type LLVMMetadataRef uintptr
type LLVMNamedMDNodeRef uintptr
type LLVMBuilderRef uintptr
type LLVMDIBuilderRef uintptr
type LLVMModuleProviderRef uintptr
type LLVMPassManagerRef uintptr
type LLVMUseRef uintptr
type LLVMOperandBundleRef uintptr
type LLVMAttributeRef uintptr
type LLVMDiagnosticInfoRef uintptr
type LLVMComdatRef uintptr
type LLVMJITEventListenerRef uintptr
type LLVMBinaryRef uintptr
type LLVMDbgRecordRef uintptr
type LLVMTargetDataRef uintptr
type LLVMTargetLibraryInfoRef uintptr
type LLVMTargetMachineOptionsRef uintptr
type LLVMTargetMachineRef uintptr
type LLVMTargetRef uintptr
type LLVMGenericValueRef uintptr
type LLVMExecutionEngineRef uintptr
type LLVMMCJITMemoryManagerRef uintptr
type LLVMSectionIteratorRef uintptr
type LLVMSymbolIteratorRef uintptr
type LLVMRelocationIteratorRef uintptr
type LLVMObjectFileRef uintptr
type LLVMPassBuilderOptionsRef uintptr
type LLVMErrorRef uintptr
type LLVMErrorTypeId uintptr

type LLVMBool int
type LLVMAttributeIndex uint
type LLVMFastMathFlags uint
type LLVMGEPNoWrapFlags uint
type LLVMMetadataKind uint
type LLVMDWARFTypeEncoding uint

type LLVMFatalErrorHandler uintptr
type LLVMDiagnosticHandler uintptr
type LLVMYieldCallback uintptr
type LLVMMemoryManagerAllocateCodeSectionCallback uintptr
type LLVMMemoryManagerAllocateDataSectionCallback uintptr
type LLVMMemoryManagerFinalizeMemoryCallback uintptr
type LLVMMemoryManagerDestroyCallback uintptr

type LLVMValueMetadataEntry struct {
	_ [0]byte
}

type LLVMMCJITCompilerOptions struct {
	OptLevel           uint
	CodeModel          LLVMCodeModel
	NoFramePointerElim LLVMBool
	EnableFastISel     LLVMBool
	MCJMM              LLVMMCJITMemoryManagerRef
}

type LLVMOpcode int

const (
	LLVMRet            LLVMOpcode = 1
	LLVMBr             LLVMOpcode = 2
	LLVMSwitch         LLVMOpcode = 3
	LLVMIndirectBr     LLVMOpcode = 4
	LLVMInvoke         LLVMOpcode = 5
	LLVMUnreachable    LLVMOpcode = 7
	LLVMCleanupRet     LLVMOpcode = 8
	LLVMCatchRet       LLVMOpcode = 9
	LLVMCatchPad       LLVMOpcode = 10
	LLVMCleanupPad     LLVMOpcode = 11
	LLVMCatchSwitch    LLVMOpcode = 12
	LLVMAdd            LLVMOpcode = 13
	LLVMFAdd           LLVMOpcode = 14
	LLVMSub            LLVMOpcode = 15
	LLVMFSub           LLVMOpcode = 16
	LLVMMul            LLVMOpcode = 17
	LLVMFMul           LLVMOpcode = 18
	LLVMUDiv           LLVMOpcode = 19
	LLVMSDiv           LLVMOpcode = 20
	LLVMFDiv           LLVMOpcode = 21
	LLVMURem           LLVMOpcode = 22
	LLVMSRem           LLVMOpcode = 23
	LLVMFRem           LLVMOpcode = 24
	LLVMShl            LLVMOpcode = 25
	LLVMLShr           LLVMOpcode = 26
	LLVMAShr           LLVMOpcode = 27
	LLVMAnd            LLVMOpcode = 28
	LLVMOr             LLVMOpcode = 29
	LLVMXor            LLVMOpcode = 30
	LLVMAlloca         LLVMOpcode = 31
	LLVMLoad           LLVMOpcode = 32
	LLVMStore          LLVMOpcode = 33
	LLVMGetElementPtr  LLVMOpcode = 34
	LLVMTrunc          LLVMOpcode = 35
	LLVMZExt           LLVMOpcode = 36
	LLVMSExt           LLVMOpcode = 37
	LLVMFPToUI         LLVMOpcode = 38
	LLVMFPToSI         LLVMOpcode = 39
	LLVMUIToFP         LLVMOpcode = 40
	LLVMSIToFP         LLVMOpcode = 41
	LLVMFPTrunc        LLVMOpcode = 42
	LLVMFPExt          LLVMOpcode = 43
	LLVMPtrToInt       LLVMOpcode = 44
	LLVMIntToPtr       LLVMOpcode = 45
	LLVMBitCast        LLVMOpcode = 46
	LLVMAddrSpaceCast  LLVMOpcode = 47
	LLVMICmp           LLVMOpcode = 50
	LLVMFCmp           LLVMOpcode = 51
	LLVMPHI            LLVMOpcode = 52
	LLVMCall           LLVMOpcode = 53
	LLVMSelect         LLVMOpcode = 54
	LLVMUserOp1        LLVMOpcode = 55
	LLVMUserOp2        LLVMOpcode = 56
	LLVMVAArg          LLVMOpcode = 57
	LLVMExtractElement LLVMOpcode = 58
	LLVMInsertElement  LLVMOpcode = 59
	LLVMShuffleVector  LLVMOpcode = 60
	LLVMExtractValue   LLVMOpcode = 61
	LLVMInsertValue    LLVMOpcode = 62
	LLVMFreeze         LLVMOpcode = 63
	LLVMCallBr         LLVMOpcode = 64
	LLVMFNeg           LLVMOpcode = 66
	LLVMAtomicCmpXchg  LLVMOpcode = 67
	LLVMAtomicRMW      LLVMOpcode = 68
	LLVMFence          LLVMOpcode = 69
)

type LLVMTypeKind int

const (
	LLVMVoidTypeKind          LLVMTypeKind = 0
	LLVMHalfTypeKind          LLVMTypeKind = 1
	LLVMFloatTypeKind         LLVMTypeKind = 2
	LLVMDoubleTypeKind        LLVMTypeKind = 3
	LLVMX86_FP80TypeKind      LLVMTypeKind = 4
	LLVMFP128TypeKind         LLVMTypeKind = 5
	LLVMPPC_FP128TypeKind     LLVMTypeKind = 6
	LLVMLabelTypeKind         LLVMTypeKind = 7
	LLVMIntegerTypeKind       LLVMTypeKind = 8
	LLVMFunctionTypeKind      LLVMTypeKind = 9
	LLVMStructTypeKind        LLVMTypeKind = 10
	LLVMArrayTypeKind         LLVMTypeKind = 11
	LLVMPointerTypeKind       LLVMTypeKind = 12
	LLVMVectorTypeKind        LLVMTypeKind = 13
	LLVMMetadataTypeKind      LLVMTypeKind = 14
	LLVMX86_MMXTypeKind       LLVMTypeKind = 15
	LLVMTokenTypeKind         LLVMTypeKind = 16
	LLVMScalableVectorTypeKind LLVMTypeKind = 17
	LLVMBFloatTypeKind        LLVMTypeKind = 18
	LLVMX86_AMXTypeKind       LLVMTypeKind = 19
	LLVMTargetExtTypeKind     LLVMTypeKind = 20
)

type LLVMLinkage int

const (
	LLVMExternalLinkage               LLVMLinkage = 0
	LLVMAvailableExternallyLinkage    LLVMLinkage = 1
	LLVMLinkOnceAnyLinkage            LLVMLinkage = 2
	LLVMLinkOnceODRLinkage            LLVMLinkage = 3
	LLVMLinkOnceODRAutoHideLinkage    LLVMLinkage = 4
	LLVMWeakAnyLinkage                LLVMLinkage = 5
	LLVMWeakODRLinkage                LLVMLinkage = 6
	LLVMAppendingLinkage              LLVMLinkage = 7
	LLVMInternalLinkage               LLVMLinkage = 8
	LLVMPrivateLinkage                LLVMLinkage = 9
	LLVMDLLImportLinkage              LLVMLinkage = 10
	LLVMDLLExportLinkage              LLVMLinkage = 11
	LLVMExternalWeakLinkage           LLVMLinkage = 12
	LLVMGhostLinkage                  LLVMLinkage = 13
	LLVMCommonLinkage                 LLVMLinkage = 14
	LLVMLinkerPrivateLinkage          LLVMLinkage = 15
	LLVMLinkerPrivateWeakLinkage      LLVMLinkage = 16
)

type LLVMVisibility int

const (
	LLVMDefaultVisibility   LLVMVisibility = 0
	LLVMHiddenVisibility    LLVMVisibility = 1
	LLVMProtectedVisibility LLVMVisibility = 2
)

type LLVMUnnamedAddr int

const (
	LLVMNoUnnamedAddr     LLVMUnnamedAddr = 0
	LLVMLocalUnnamedAddr  LLVMUnnamedAddr = 1
	LLVMGlobalUnnamedAddr LLVMUnnamedAddr = 2
)

type LLVMDLLStorageClass int

const (
	LLVMDefaultStorageClass      LLVMDLLStorageClass = 0
	LLVMDLLImportStorageClass    LLVMDLLStorageClass = 1
	LLVMDLLExportStorageClass    LLVMDLLStorageClass = 2
)

type LLVMCallConv int

const (
	LLVMCCallConv             LLVMCallConv = 0
	LLVMFastCallConv          LLVMCallConv = 8
	LLVMColdCallConv          LLVMCallConv = 9
	LLVMGHCCallConv           LLVMCallConv = 10
	LLVMHiPECallConv          LLVMCallConv = 11
	LLVMWebKitJSCallConv      LLVMCallConv = 12
	LLVMAnyRegCallConv        LLVMCallConv = 13
	LLVMPreserveMostCallConv  LLVMCallConv = 14
	LLVMPreserveAllCallConv   LLVMCallConv = 15
	LLVMSwiftCallConv         LLVMCallConv = 16
	LLVMCXXFASTTLSCallConv    LLVMCallConv = 17
	LLVMX86StdcallCallConv    LLVMCallConv = 64
	LLVMX86FastcallCallConv   LLVMCallConv = 65
	LLVMARMAPCSCallConv       LLVMCallConv = 66
	LLVMARMAAPCSCallConv      LLVMCallConv = 67
	LLVMARMAAPCSVFPCallConv   LLVMCallConv = 68
	LLVMMSP430IntrCallConv    LLVMCallConv = 69
	LLVMX86ThisCallCallConv   LLVMCallConv = 70
	LLVMPTXKernelCallConv     LLVMCallConv = 71
	LLVMPTXDeviceCallConv     LLVMCallConv = 72
	LLVMSPIRKernelCallConv    LLVMCallConv = 73
	LLVMSPIRFuncCallConv      LLVMCallConv = 74
	LLVMIntelOclBiccCallConv  LLVMCallConv = 75
	LLVMX86_64SysVCallConv    LLVMCallConv = 76
	LLVMWin64CallConv         LLVMCallConv = 77
	LLVMCFGuardCheckCallConv  LLVMCallConv = 78
	LLVMWebRTCJsCallConv      LLVMCallConv = 79
	LLVMWebRTCJsFastCallConv  LLVMCallConv = 80
	LLVMAVRBuiltinCallConv    LLVMCallConv = 81
)

type LLVMValueKind int

const (
	LLVMArgumentValueKind             LLVMValueKind = 0
	LLVMBasicBlockValueKind           LLVMValueKind = 1
	LLVMMemoryUseValueKind            LLVMValueKind = 2
	LLVMMemoryDefValueKind            LLVMValueKind = 3
	LLVMMemoryPhiValueKind            LLVMValueKind = 4
	LLVMFunctionValueKind             LLVMValueKind = 5
	LLVMGlobalAliasValueKind          LLVMValueKind = 6
	LLVMGlobalIFuncValueKind          LLVMValueKind = 7
	LLVMGlobalVariableValueKind       LLVMValueKind = 8
	LLVMBlockAddressValueKind         LLVMValueKind = 9
	LLVMConstantExprValueKind         LLVMValueKind = 10
	LLVMConstantArrayValueKind        LLVMValueKind = 11
	LLVMConstantStructValueKind       LLVMValueKind = 12
	LLVMConstantVectorValueKind       LLVMValueKind = 13
	LLVMUndefValueValueKind           LLVMValueKind = 14
	LLVMConstantAggregateZeroValueKind LLVMValueKind = 15
	LLVMConstantDataArrayValueKind    LLVMValueKind = 16
	LLVMConstantDataVectorValueKind   LLVMValueKind = 17
	LLVMConstantIntValueKind          LLVMValueKind = 18
	LLVMConstantFPValueKind           LLVMValueKind = 19
	LLVMConstantPointerNullValueKind  LLVMValueKind = 20
	LLVMConstantTokenNoneValueKind    LLVMValueKind = 21
	LLVMMetadataAsValueValueKind      LLVMValueKind = 22
	LLVMInlineAsmValueKind            LLVMValueKind = 23
	LLVMInstructionValueKind          LLVMValueKind = 24
	LLVMPoisonValueValueKind          LLVMValueKind = 25
	LLVMConstantTargetNoneValueKind   LLVMValueKind = 26
	LLVMConstantPtrAuthValueKind      LLVMValueKind = 27
	LLVMDSOLocalEquivalentValueKind   LLVMValueKind = 28
	LLVMNoCFIValueValueKind           LLVMValueKind = 29
)

type LLVMIntPredicate int

const (
	LLVMIntEQ  LLVMIntPredicate = 32
	LLVMIntNE  LLVMIntPredicate = 33
	LLVMIntUGT LLVMIntPredicate = 34
	LLVMIntUGE LLVMIntPredicate = 35
	LLVMIntULT LLVMIntPredicate = 36
	LLVMIntULE LLVMIntPredicate = 37
	LLVMIntSGT LLVMIntPredicate = 38
	LLVMIntSGE LLVMIntPredicate = 39
	LLVMIntSLT LLVMIntPredicate = 40
	LLVMIntSLE LLVMIntPredicate = 41
)

type LLVMRealPredicate int

const (
	LLVMRealPredicateFalse LLVMRealPredicate = 0
	LLVMRealOEQ            LLVMRealPredicate = 1
	LLVMRealOGT            LLVMRealPredicate = 2
	LLVMRealOGE            LLVMRealPredicate = 3
	LLVMRealOLT            LLVMRealPredicate = 4
	LLVMRealOLE            LLVMRealPredicate = 5
	LLVMRealONE            LLVMRealPredicate = 6
	LLVMRealORD            LLVMRealPredicate = 7
	LLVMRealUNO            LLVMRealPredicate = 8
	LLVMRealUEQ            LLVMRealPredicate = 9
	LLVMRealUGT            LLVMRealPredicate = 10
	LLVMRealUGE            LLVMRealPredicate = 11
	LLVMRealULT            LLVMRealPredicate = 12
	LLVMRealULE            LLVMRealPredicate = 13
	LLVMRealUNE            LLVMRealPredicate = 14
	LLVMRealPredicateTrue  LLVMRealPredicate = 15
)

type LLVMLandingPadClauseTy int

const (
	LLVMLandingPadCatch  LLVMLandingPadClauseTy = 0
	LLVMLandingPadFilter LLVMLandingPadClauseTy = 1
)

type LLVMThreadLocalMode int

const (
	LLVMNotThreadLocal          LLVMThreadLocalMode = 0
	LLVMGeneralDynamicTLSModel  LLVMThreadLocalMode = 1
	LLVMLocalDynamicTLSModel    LLVMThreadLocalMode = 2
	LLVMInitialExecTLSModel     LLVMThreadLocalMode = 3
	LLVMLocalExecTLSModel       LLVMThreadLocalMode = 4
)

type LLVMAtomicOrdering int

const (
	LLVMAtomicOrderingNotAtomic            LLVMAtomicOrdering = 0
	LLVMAtomicOrderingUnordered            LLVMAtomicOrdering = 1
	LLVMAtomicOrderingMonotonic            LLVMAtomicOrdering = 2
	LLVMAtomicOrderingAcquire              LLVMAtomicOrdering = 4
	LLVMAtomicOrderingRelease              LLVMAtomicOrdering = 5
	LLVMAtomicOrderingAcquireRelease       LLVMAtomicOrdering = 6
	LLVMAtomicOrderingSequentiallyConsistent LLVMAtomicOrdering = 7
)

type LLVMAtomicRMWBinOp int

const (
	LLVMAtomicRMWBinOpXchg    LLVMAtomicRMWBinOp = 0
	LLVMAtomicRMWBinOpAdd     LLVMAtomicRMWBinOp = 1
	LLVMAtomicRMWBinOpSub     LLVMAtomicRMWBinOp = 2
	LLVMAtomicRMWBinOpAnd     LLVMAtomicRMWBinOp = 3
	LLVMAtomicRMWBinOpNand    LLVMAtomicRMWBinOp = 4
	LLVMAtomicRMWBinOpOr      LLVMAtomicRMWBinOp = 5
	LLVMAtomicRMWBinOpXor     LLVMAtomicRMWBinOp = 6
	LLVMAtomicRMWBinOpMax     LLVMAtomicRMWBinOp = 7
	LLVMAtomicRMWBinOpMin     LLVMAtomicRMWBinOp = 8
	LLVMAtomicRMWBinOpUMax    LLVMAtomicRMWBinOp = 9
	LLVMAtomicRMWBinOpUMin    LLVMAtomicRMWBinOp = 10
	LLVMAtomicRMWBinOpFAdd    LLVMAtomicRMWBinOp = 11
	LLVMAtomicRMWBinOpFSub    LLVMAtomicRMWBinOp = 12
	LLVMAtomicRMWBinOpFMax    LLVMAtomicRMWBinOp = 13
	LLVMAtomicRMWBinOpFMin    LLVMAtomicRMWBinOp = 14
	LLVMAtomicRMWBinOpUIncWrap LLVMAtomicRMWBinOp = 15
	LLVMAtomicRMWBinOpUDecWrap LLVMAtomicRMWBinOp = 16
	LLVMAtomicRMWBinOpUSubCond LLVMAtomicRMWBinOp = 17
	LLVMAtomicRMWBinOpUSubSat  LLVMAtomicRMWBinOp = 18
	LLVMAtomicRMWBinOpUAddSat  LLVMAtomicRMWBinOp = 19
)

type LLVMDiagnosticSeverity int

const (
	LLVMDSError   LLVMDiagnosticSeverity = 0
	LLVMDSWarning LLVMDiagnosticSeverity = 1
	LLVMDSRemark  LLVMDiagnosticSeverity = 2
	LLVMDSNote    LLVMDiagnosticSeverity = 3
)

type LLVMInlineAsmDialect int

const (
	LLVMInlineAsmDialectATT   LLVMInlineAsmDialect = 0
	LLVMInlineAsmDialectIntel LLVMInlineAsmDialect = 1
)

type LLVMModuleFlagBehavior int

const (
	LLVMModuleFlagBehaviorError       LLVMModuleFlagBehavior = 0
	LLVMModuleFlagBehaviorWarning     LLVMModuleFlagBehavior = 1
	LLVMModuleFlagBehaviorRequire     LLVMModuleFlagBehavior = 2
	LLVMModuleFlagBehaviorOverride    LLVMModuleFlagBehavior = 3
	LLVMModuleFlagBehaviorAppend      LLVMModuleFlagBehavior = 4
	LLVMModuleFlagBehaviorAppendUnique LLVMModuleFlagBehavior = 5
)

type LLVMTailCallKind int

const (
	LLVMTailCallKindNone    LLVMTailCallKind = 0
	LLVMTailCallKindTail    LLVMTailCallKind = 1
	LLVMTailCallKindMustTail LLVMTailCallKind = 2
	LLVMTailCallKindNoTail  LLVMTailCallKind = 3
)

type LLVMByteOrdering int

const (
	LLVMBigEndian    LLVMByteOrdering = 0
	LLVMLittleEndian LLVMByteOrdering = 1
)

type LLVMCodeGenOptLevel int

const (
	LLVMCodeGenLevelNone       LLVMCodeGenOptLevel = 0
	LLVMCodeGenLevelLess       LLVMCodeGenOptLevel = 1
	LLVMCodeGenLevelDefault    LLVMCodeGenOptLevel = 2
	LLVMCodeGenLevelAggressive LLVMCodeGenOptLevel = 3
)

type LLVMRelocMode int

const (
	LLVMRelocDefault       LLVMRelocMode = 0
	LLVMRelocStatic        LLVMRelocMode = 1
	LLVMRelocPIC           LLVMRelocMode = 2
	LLVMRelocDynamicNoPic  LLVMRelocMode = 3
	LLVMRelocROPI          LLVMRelocMode = 4
	LLVMRelocRWPI          LLVMRelocMode = 5
	LLVMRelocROPI_RWPI     LLVMRelocMode = 6
)

type LLVMCodeModel int

const (
	LLVMCodeModelDefault    LLVMCodeModel = 0
	LLVMCodeModelJITDefault LLVMCodeModel = 1
	LLVMCodeModelTiny       LLVMCodeModel = 2
	LLVMCodeModelSmall      LLVMCodeModel = 3
	LLVMCodeModelKernel     LLVMCodeModel = 4
	LLVMCodeModelMedium     LLVMCodeModel = 5
	LLVMCodeModelLarge      LLVMCodeModel = 6
)

type LLVMCodeGenFileType int

const (
	LLVMCodeGenFileTypeAssemblySource LLVMCodeGenFileType = 0
	LLVMCodeGenFileTypeObject         LLVMCodeGenFileType = 1
)

type LLVMGlobalISelAbortMode int

const (
	LLVMGlobalISelAbortModeNever   LLVMGlobalISelAbortMode = 0
	LLVMGlobalISelAbortModeError   LLVMGlobalISelAbortMode = 1
	LLVMGlobalISelAbortModeDisable LLVMGlobalISelAbortMode = 2
)

type LLVMVerifierFailureAction int

const (
	LLVMVerifierAbortProcessAction LLVMVerifierFailureAction = 0
	LLVMVerifierPrintMessageAction LLVMVerifierFailureAction = 1
	LLVMVerifierReturnStatusAction LLVMVerifierFailureAction = 2
)

type LLVMBinaryType int

const (
	LLVMBinaryTypeArchive              LLVMBinaryType = 0
	LLVMBinaryTypeMachOUniversalBinary LLVMBinaryType = 1
	LLVMBinaryTypeCOFFImportFile       LLVMBinaryType = 2
	LLVMBinaryTypeIR                   LLVMBinaryType = 3
	LLVMBinaryTypeWinRes               LLVMBinaryType = 4
	LLVMBinaryTypeCOFF                 LLVMBinaryType = 5
	LLVMBinaryTypeELF32L               LLVMBinaryType = 6
	LLVMBinaryTypeELF32B               LLVMBinaryType = 7
	LLVMBinaryTypeELF64L               LLVMBinaryType = 8
	LLVMBinaryTypeELF64B               LLVMBinaryType = 9
	LLVMBinaryTypeMachO32L             LLVMBinaryType = 10
	LLVMBinaryTypeMachO32B             LLVMBinaryType = 11
	LLVMBinaryTypeMachO64L             LLVMBinaryType = 12
	LLVMBinaryTypeMachO64B             LLVMBinaryType = 13
	LLVMBinaryTypeWasm                 LLVMBinaryType = 14
	LLVMBinaryTypeOffloadFatBinary     LLVMBinaryType = 15
	LLVMBinaryTypeAssembly             LLVMBinaryType = 16
	LLVMBinaryTypeDXContainer          LLVMBinaryType = 17
)

type LLVMDIFlags uint

const (
	LLVMDIFlagZero                LLVMDIFlags = 0
	LLVMDIFlagPrivate             LLVMDIFlags = 1 << 0
	LLVMDIFlagProtected           LLVMDIFlags = 1 << 1
	LLVMDIFlagPublic              LLVMDIFlags = 1 << 2
	LLVMDIFlagFwdDecl             LLVMDIFlags = 1 << 3
	LLVMDIFlagAppleBlock          LLVMDIFlags = 1 << 4
	LLVMDIFlagReservedBit5        LLVMDIFlags = 1 << 5
	LLVMDIFlagVirtual             LLVMDIFlags = 1 << 6
	LLVMDIFlagArtificial          LLVMDIFlags = 1 << 7
	LLVMDIFlagExplicit            LLVMDIFlags = 1 << 8
	LLVMDIFlagPrototyped          LLVMDIFlags = 1 << 9
	LLVMDIFlagObjcClassComplete   LLVMDIFlags = 1 << 10
	LLVMDIFlagObjectPointer       LLVMDIFlags = 1 << 11
	LLVMDIFlagVector              LLVMDIFlags = 1 << 12
	LLVMDIFlagStaticMember        LLVMDIFlags = 1 << 13
	LLVMDIFlagLValueReference     LLVMDIFlags = 1 << 14
	LLVMDIFlagRValueReference     LLVMDIFlags = 1 << 15
	LLVMDIFlagReservedBit16       LLVMDIFlags = 1 << 16
	LLVMDIFlagSingleInheritance   LLVMDIFlags = 1 << 17
	LLVMDIFlagMultipleInheritance  LLVMDIFlags = 1 << 18
	LLVMDIFlagVirtualInheritance  LLVMDIFlags = 1 << 19
	LLVMDIFlagIntroducedVirtual   LLVMDIFlags = 1 << 20
	LLVMDIFlagBitField            LLVMDIFlags = 1 << 21
	LLVMDIFlagNoReturn            LLVMDIFlags = 1 << 22
	LLVMDIFlagTypePassByValue     LLVMDIFlags = 1 << 23
	LLVMDIFlagTypePassByReference LLVMDIFlags = 1 << 24
	LLVMDIFlagEnumClass           LLVMDIFlags = 1 << 25
	LLVMDIFlagFixedEnum           LLVMDIFlags = 1 << 26
	LLVMDIFlagThunk               LLVMDIFlags = 1 << 27
	LLVMDIFlagTrivial             LLVMDIFlags = 1 << 28
	LLVMDIFlagBigEndian           LLVMDIFlags = 1 << 29
	LLVMDIFlagLittleEndian        LLVMDIFlags = 1 << 30
	LLVMDIFlagAllCallsDescribed   LLVMDIFlags = 1 << 31
	LLVMDIFlagExportSymbols       LLVMDIFlags = 1 << 32
	LLVMDIFlagDeleted             LLVMDIFlags = 1 << 33
	LLVMDIFlagAccessibility       LLVMDIFlags = LLVMDIFlagPrivate | LLVMDIFlagProtected | LLVMDIFlagPublic
	LLVMDIFlagPtrToMemberRep      LLVMDIFlags = LLVMDIFlagSingleInheritance | LLVMDIFlagMultipleInheritance | LLVMDIFlagVirtualInheritance
)

type LLVMDWARFSourceLanguage int

const (
	LLVMDWARFSourceLanguageC89            LLVMDWARFSourceLanguage = 0
	LLVMDWARFSourceLanguageC              LLVMDWARFSourceLanguage = 1
	LLVMDWARFSourceLanguageAda83          LLVMDWARFSourceLanguage = 2
	LLVMDWARFSourceLanguageC_PLUS_PLUS    LLVMDWARFSourceLanguage = 3
	LLVMDWARFSourceLanguageCobol74        LLVMDWARFSourceLanguage = 4
	LLVMDWARFSourceLanguageCobol85        LLVMDWARFSourceLanguage = 5
	LLVMDWARFSourceLanguageFortran77      LLVMDWARFSourceLanguage = 6
	LLVMDWARFSourceLanguageFortran90      LLVMDWARFSourceLanguage = 7
	LLVMDWARFSourceLanguagePascal83       LLVMDWARFSourceLanguage = 8
	LLVMDWARFSourceLanguageModula2        LLVMDWARFSourceLanguage = 9
	LLVMDWARFSourceLanguageJava           LLVMDWARFSourceLanguage = 10
	LLVMDWARFSourceLanguageC99            LLVMDWARFSourceLanguage = 11
	LLVMDWARFSourceLanguageAda95          LLVMDWARFSourceLanguage = 12
	LLVMDWARFSourceLanguageFortran95      LLVMDWARFSourceLanguage = 13
	LLVMDWARFSourceLanguagePLI            LLVMDWARFSourceLanguage = 14
	LLVMDWARFSourceLanguageObjC           LLVMDWARFSourceLanguage = 15
	LLVMDWARFSourceLanguageObjC_PLUS_PLUS LLVMDWARFSourceLanguage = 16
	LLVMDWARFSourceLanguageUPC            LLVMDWARFSourceLanguage = 17
	LLVMDWARFSourceLanguageD              LLVMDWARFSourceLanguage = 18
	LLVMDWARFSourceLanguagePython         LLVMDWARFSourceLanguage = 19
	LLVMDWARFSourceLanguageOpenCL         LLVMDWARFSourceLanguage = 20
	LLVMDWARFSourceLanguageGo             LLVMDWARFSourceLanguage = 21
	LLVMDWARFSourceLanguageModula3        LLVMDWARFSourceLanguage = 22
	LLVMDWARFSourceLanguageHaskell        LLVMDWARFSourceLanguage = 23
	LLVMDWARFSourceLanguageC_PLUS_PLUS_03 LLVMDWARFSourceLanguage = 24
	LLVMDWARFSourceLanguageC_PLUS_PLUS_11 LLVMDWARFSourceLanguage = 25
	LLVMDWARFSourceLanguageOCaml          LLVMDWARFSourceLanguage = 26
	LLVMDWARFSourceLanguageRust           LLVMDWARFSourceLanguage = 27
	LLVMDWARFSourceLanguageC11            LLVMDWARFSourceLanguage = 28
	LLVMDWARFSourceLanguageSwift          LLVMDWARFSourceLanguage = 29
	LLVMDWARFSourceLanguageJulia          LLVMDWARFSourceLanguage = 30
	LLVMDWARFSourceLanguageDylan          LLVMDWARFSourceLanguage = 31
	LLVMDWARFSourceLanguageC_PLUS_PLUS_14 LLVMDWARFSourceLanguage = 32
	LLVMDWARFSourceLanguageFortran03      LLVMDWARFSourceLanguage = 33
	LLVMDWARFSourceLanguageFortran08      LLVMDWARFSourceLanguage = 34
	LLVMDWARFSourceLanguageRenderScript   LLVMDWARFSourceLanguage = 35
	LLVMDWARFSourceLanguageBLISS          LLVMDWARFSourceLanguage = 36
	LLVMDWARFSourceLanguageKotlin         LLVMDWARFSourceLanguage = 37
	LLVMDWARFSourceLanguageZig            LLVMDWARFSourceLanguage = 38
	LLVMDWARFSourceLanguageCrystal        LLVMDWARFSourceLanguage = 39
	LLVMDWARFSourceLanguageC_PLUS_PLUS_17 LLVMDWARFSourceLanguage = 40
	LLVMDWARFSourceLanguageC_PLUS_PLUS_20 LLVMDWARFSourceLanguage = 41
	LLVMDWARFSourceLanguageC17            LLVMDWARFSourceLanguage = 42
	LLVMDWARFSourceLanguageFortran18      LLVMDWARFSourceLanguage = 43
	LLVMDWARFSourceLanguageAda2005        LLVMDWARFSourceLanguage = 44
	LLVMDWARFSourceLanguageAda2012        LLVMDWARFSourceLanguage = 45
	LLVMDWARFSourceLanguageAda2022        LLVMDWARFSourceLanguage = 46
	LLVMDWARFSourceLanguageNim            LLVMDWARFSourceLanguage = 47
	LLVMDWARFSourceLanguageHIP            LLVMDWARFSourceLanguage = 48
	LLVMDWARFSourceLanguageC_PLUS_PLUS_23 LLVMDWARFSourceLanguage = 49
	LLVMDWARFSourceLanguageC23            LLVMDWARFSourceLanguage = 50
	LLVMDWARFSourceLanguageHylo           LLVMDWARFSourceLanguage = 51
	LLVMDWARFSourceLanguageMojo           LLVMDWARFSourceLanguage = 52
	LLVMDWARFSourceLanguageOpenCLPP       LLVMDWARFSourceLanguage = 54
	LLVMDWARFSourceLanguageCUDA           LLVMDWARFSourceLanguage = 55
	LLVMDWARFSourceLanguageHI             LLVMDWARFSourceLanguage = 56
	LLVMDWARFSourceLanguageCommonLisp     LLVMDWARFSourceLanguage = 57
	LLVMDWARFSourceLanguageVerve          LLVMDWARFSourceLanguage = 58
	LLVMDWARFSourceLanguageLua            LLVMDWARFSourceLanguage = 59
	LLVMDWARFSourceLanguageCobra          LLVMDWARFSourceLanguage = 60
	LLVMDWARFSourceLanguageLLVM           LLVMDWARFSourceLanguage = 61
	LLVMDWARFSourceLanguageLLVMIR         LLVMDWARFSourceLanguage = 62
	LLVMDWARFSourceLanguageHLSL           LLVMDWARFSourceLanguage = 63
	LLVMDWARFSourceLanguageGLSL           LLVMDWARFSourceLanguage = 64
	LLVMDWARFSourceLanguageGLSLES         LLVMDWARFSourceLanguage = 65
	LLVMDWARFSourceLanguageWGSL           LLVMDWARFSourceLanguage = 66
	LLVMDWARFSourceLanguageCIR            LLVMDWARFSourceLanguage = 67
	LLVMDWARFSourceLanguageSPIRV          LLVMDWARFSourceLanguage = 68
	LLVMDWARFSourceLanguageCPPForOpenCL   LLVMDWARFSourceLanguage = 69
	LLVMDWARFSourceLanguageSYCL           LLVMDWARFSourceLanguage = 70
	LLVMDWARFSourceLanguagePPL            LLVMDWARFSourceLanguage = 71
)

type LLVMDWARFEmissionKind int

const (
	LLVMDWARFEmissionNone          LLVMDWARFEmissionKind = 0
	LLVMDWARFEmissionFull          LLVMDWARFEmissionKind = 1
	LLVMDWARFEmissionLineTablesOnly LLVMDWARFEmissionKind = 2
)

type LLVMDWARFMacinfoRecordType int

const (
	LLVMDWARFMacinfoRecordTypeDefine    LLVMDWARFMacinfoRecordType = 1
	LLVMDWARFMacinfoRecordTypeUndef     LLVMDWARFMacinfoRecordType = 2
	LLVMDWARFMacinfoRecordTypeStartFile LLVMDWARFMacinfoRecordType = 3
	LLVMDWARFMacinfoRecordTypeEndFile   LLVMDWARFMacinfoRecordType = 4
	LLVMDWARFMacinfoRecordTypeVendorExt LLVMDWARFMacinfoRecordType = 255
)

const (
	LLVMFastMathAllowReassoc  LLVMFastMathFlags = 1 << 0
	LLVMFastMathNoNaNs        LLVMFastMathFlags = 1 << 1
	LLVMFastMathNoInfs        LLVMFastMathFlags = 1 << 2
	LLVMFastMathNoSignedZeros LLVMFastMathFlags = 1 << 3
	LLVMFastMathAllowReciprocal LLVMFastMathFlags = 1 << 4
	LLVMFastMathAllowContract  LLVMFastMathFlags = 1 << 5
	LLVMFastMathApproxFunc     LLVMFastMathFlags = 1 << 6
	LLVMFastMathAllowAll       LLVMFastMathFlags = LLVMFastMathAllowReassoc | LLVMFastMathNoNaNs | LLVMFastMathNoInfs | LLVMFastMathNoSignedZeros | LLVMFastMathAllowReciprocal | LLVMFastMathAllowContract | LLVMFastMathApproxFunc
)

const (
	LLVMGEPFlagInBounds LLVMGEPNoWrapFlags = 1 << 0
	LLVMGEPFlagNUSW     LLVMGEPNoWrapFlags = 1 << 1
	LLVMGEPFlagNUW      LLVMGEPNoWrapFlags = 1 << 2
)

func LLVMInstallFatalErrorHandler(Handler LLVMFatalErrorHandler) {}
func LLVMResetFatalErrorHandler() {}
func LLVMEnablePrettyStackTrace() {}
func LLVMFatalError(ErrorMsg string) {}
func LLVMYield(Callback LLVMYieldCallback, OpaqueHandle uintptr) {}

func LLVMShutdown() {}
func LLVMGetVersion(Major, Minor, Patch *uint) { *Major = 0; *Minor = 0; *Patch = 0 }
func LLVMCreateMessage(Message string) *byte { return nil }
func LLVMDisposeMessage(Message *byte) {}
func LLVMContextCreate() LLVMContextRef { return 0 }
func LLVMGetGlobalContext() LLVMContextRef { return 0 }
func LLVMContextSetDiagnosticHandler(C LLVMContextRef, Handler LLVMDiagnosticHandler, DiagnosticContext uintptr) {}
func LLVMContextGetDiagnosticHandler(C LLVMContextRef) LLVMDiagnosticHandler { return 0 }
func LLVMContextGetDiagnosticContext(C LLVMContextRef) uintptr { return 0 }
func LLVMContextSetYieldCallback(C LLVMContextRef, Callback LLVMYieldCallback, OpaqueHandle uintptr) {}
func LLVMContextShouldDiscardValueNames(C LLVMContextRef) LLVMBool { return 0 }
func LLVMContextSetDiscardValueNames(C LLVMContextRef, Discard LLVMBool) {}
func LLVMContextDispose(C LLVMContextRef) {}
func LLVMGetDiagInfoDescription(DI LLVMDiagnosticInfoRef) *byte { return nil }
func LLVMGetDiagInfoSeverity(DI LLVMDiagnosticInfoRef) LLVMDiagnosticSeverity { return 0 }
func LLVMGetMDKindIDInContext(C LLVMContextRef, Name string, SLen uint) uint { return 0 }
func LLVMGetMDKindID(Name string, SLen uint) uint { return 0 }
func LLVMGetMDNodeIDInContext(C LLVMContextRef, Name string, SLen uint) uint { return 0 }
func LLVMGetMDNodeID(Name string, SLen uint) uint { return 0 }
func LLVMGetEnumAttributeKindForName(Name string, SLen uint) uint { return 0 }
func LLVMGetLastEnumAttributeKind() uint { return 0 }
func LLVMCreateAttribute(C LLVMContextRef, KindID uint, Val uint64) LLVMAttributeRef { return 0 }
func LLVMCreateEnumAttribute(C LLVMContextRef, KindID uint, Val uint64) LLVMAttributeRef { return 0 }
func LLVMCreateTypeAttribute(C LLVMContextRef, KindID uint, Ty LLVMTypeRef) LLVMAttributeRef { return 0 }
func LLVMCreateStringAttribute(C LLVMContextRef, K string, KLength uint, V string, VLength uint) LLVMAttributeRef { return 0 }
func LLVMAttributeGetEnumKind(A LLVMAttributeRef) uint { return 0 }
func LLVMAttributeGetEnumValue(A LLVMAttributeRef) uint64 { return 0 }
func LLVMAttributeGetStringKind(A LLVMAttributeRef) *byte { return nil }
func LLVMAttributeGetStringKindLength(A LLVMAttributeRef) uint { return 0 }
func LLVMAttributeGetStringValue(A LLVMAttributeRef) *byte { return nil }
func LLVMAttributeGetStringValueLength(A LLVMAttributeRef) uint { return 0 }
func LLVMGetNamedMetadataNumOperands(M LLVMModuleRef, Name string) uint { return 0 }
func LLVMGetNamedMetadataOperands(M LLVMModuleRef, Name string, Dest *LLVMValueRef) {}
func LLVMAddNamedMetadataOperand(M LLVMModuleRef, Name string, Val LLVMValueRef) {}
func LLVMIsAMDGPU(C LLVMContextRef) LLVMBool { return 0 }
func LLVMIsSPIRV(C LLVMContextRef) LLVMBool { return 0 }

func LLVMModuleCreateWithName(ModuleID string) LLVMModuleRef { return 0 }
func LLVMModuleCreateWithNameInContext(ModuleID string, C LLVMContextRef) LLVMModuleRef { return 0 }
func LLVMCloneModule(M LLVMModuleRef) LLVMModuleRef { return 0 }
func LLVMDisposeModule(M LLVMModuleRef) {}
func LLVMGetModuleIdentifier(M LLVMModuleRef, Len *uint) *byte { return nil }
func LLVMSetModuleIdentifier(M LLVMModuleRef, Ident string, Len uint) {}
func LLVMGetSourceFileName(M LLVMModuleRef, Len *uint) *byte { return nil }
func LLVMSetSourceFileName(M LLVMModuleRef, Name string, Len uint) {}
func LLVMGetModuleDataLayout(M LLVMModuleRef) *byte { return nil }
func LLVMSetModuleDataLayout(M LLVMModuleRef, Triple string) {}
func LLVMGetModuleTarget(M LLVMModuleRef) *byte { return nil }
func LLVMSetModuleTarget(M LLVMModuleRef, Triple string) {}
func LLVMGetModuleFlag(M LLVMModuleRef, Key string, KeyLen uint) LLVMMetadataRef { return 0 }
func LLVMAddModuleFlag(M LLVMModuleRef, Behavior LLVMModuleFlagBehavior, Key string, KeyLen uint, Val LLVMMetadataRef) {}
func LLVMModuleGetAsm(M LLVMModuleRef) *byte { return nil }
func LLVMSetModuleInlineAsm(M LLVMModuleRef, Asm string, Len uint) {}
func LLVMAppendModuleInlineAsm(M LLVMModuleRef, Asm string, Len uint) {}
func LLVMGetModuleContext(M LLVMModuleRef) LLVMContextRef { return 0 }
func LLVMGetTypeByName(M LLVMModuleRef, Name string) LLVMTypeRef { return 0 }
func LLVMGetTypeByName2(C LLVMContextRef, Name string) LLVMTypeRef { return 0 }
func LLVMGetNamedGlobal(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetNamedFunction(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetNamedGlobalAlias(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetNamedGlobalIFunc(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetFirstGlobal(M LLVMModuleRef) LLVMValueRef { return 0 }
func LLVMGetLastGlobal(M LLVMModuleRef) LLVMValueRef { return 0 }
func LLVMGetNextGlobal(GlobalVar LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetPreviousGlobal(GlobalVar LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGlobalGetValueType(GlobalVar LLVMValueRef) LLVMTypeRef { return 0 }
func LLVMSetInitializer(GlobalVar LLVMValueRef, ConstantVal LLVMValueRef) {}
func LLVMIsDeclaration(GlobalVar LLVMValueRef) LLVMBool { return 0 }
func LLVMSetDSOLocal(GlobalVal LLVMValueRef, DSO LLVMBool) {}
func LLVMIsDSOLocal(GlobalVal LLVMValueRef) LLVMBool { return 0 }
func LLVMGetNamedGlobalAlias2(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetNamedGlobalIFunc2(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMAddAlias(M LLVMModuleRef, T LLVMTypeRef, Aliasee LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMAddAlias2(M LLVMModuleRef, T LLVMTypeRef, AddrSpace uint, Aliasee LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMAddFunction(M LLVMModuleRef, Name string, FunctionTy LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMGetNamedFunction2(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetFirstFunction(M LLVMModuleRef) LLVMValueRef { return 0 }
func LLVMGetLastFunction(M LLVMModuleRef) LLVMValueRef { return 0 }
func LLVMGetNextFunction(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetPreviousFunction(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetNamedGlobal2(M LLVMModuleRef, Name string) LLVMValueRef { return 0 }
func LLVMGetNamedMetadata(M LLVMModuleRef, Name string) LLVMNamedMDNodeRef { return 0 }
func LLVMGetFirstNamedMetadata(M LLVMModuleRef) LLVMNamedMDNodeRef { return 0 }
func LLVMGetLastNamedMetadata(M LLVMModuleRef) LLVMNamedMDNodeRef { return 0 }
func LLVMGetNextNamedMetadata(M LLVMNamedMDNodeRef) LLVMNamedMDNodeRef { return 0 }
func LLVMGetPreviousNamedMetadata(M LLVMNamedMDNodeRef) LLVMNamedMDNodeRef { return 0 }
func LLVMAddGlobal(M LLVMModuleRef, T LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMAddGlobalInAddressSpace(M LLVMModuleRef, T LLVMTypeRef, Name string, AddressSpace uint) LLVMValueRef { return 0 }
func LLVMGetNamedMetadataName(M LLVMNamedMDNodeRef) *byte { return nil }
func LLVMGetNamedMetadataNameLen(M LLVMNamedMDNodeRef) uint { return 0 }
func LLVMNull(M LLVMModuleRef) LLVMValueRef { return 0 }
func LLVMGetModuleDIFlags(M LLVMModuleRef) LLVMDIFlags { return 0 }
func LLVMSetModuleDIFlags(M LLVMModuleRef, Flags LLVMDIFlags) {}

func LLVMTypeIsSized(Ty LLVMTypeRef) LLVMBool { return 0 }
func LLVMGetTypeKind(Ty LLVMTypeRef) LLVMTypeKind { return 0 }
func LLVMTypeGetContext(Ty LLVMTypeRef) LLVMContextRef { return 0 }
func LLVMDumpType(Ty LLVMTypeRef) {}
func LLVMPrintTypeToString(Ty LLVMTypeRef) *byte { return nil }
func LLVMInt1Type() LLVMTypeRef { return 0 }
func LLVMInt1TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMInt8Type() LLVMTypeRef { return 0 }
func LLVMInt8TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMInt16Type() LLVMTypeRef { return 0 }
func LLVMInt16TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMInt32Type() LLVMTypeRef { return 0 }
func LLVMInt32TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMInt64Type() LLVMTypeRef { return 0 }
func LLVMInt64TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMInt128Type() LLVMTypeRef { return 0 }
func LLVMInt128TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMIntType(NumBits uint) LLVMTypeRef { return 0 }
func LLVMIntTypeInContext(C LLVMContextRef, NumBits uint) LLVMTypeRef { return 0 }
func LLVMGetIntTypeWidth(IntegerTy LLVMTypeRef) uint { return 0 }
func LLVMHalfType() LLVMTypeRef { return 0 }
func LLVMHalfTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMBFloatType() LLVMTypeRef { return 0 }
func LLVMBFloatTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMFloatType() LLVMTypeRef { return 0 }
func LLVMFloatTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMDoubleType() LLVMTypeRef { return 0 }
func LLVMDoubleTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMX86FP80Type() LLVMTypeRef { return 0 }
func LLVMX86FP80TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMFP128Type() LLVMTypeRef { return 0 }
func LLVMFP128TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMPPCFP128Type() LLVMTypeRef { return 0 }
func LLVMPPCFP128TypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMVoidType() LLVMTypeRef { return 0 }
func LLVMVoidTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMLabelType() LLVMTypeRef { return 0 }
func LLVMLabelTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMX86MMXType() LLVMTypeRef { return 0 }
func LLVMX86MMXTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMX86AMXType() LLVMTypeRef { return 0 }
func LLVMX86AMXTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMTokenType() LLVMTypeRef { return 0 }
func LLVMTokenTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMMetadataType() LLVMTypeRef { return 0 }
func LLVMMetadataTypeInContext(C LLVMContextRef) LLVMTypeRef { return 0 }
func LLVMArrayType(ElementType LLVMTypeRef, ElementCount uint) LLVMTypeRef { return 0 }
func LLVMArrayType2(ElementType LLVMTypeRef, ElementCount uint64) LLVMTypeRef { return 0 }
func LLVMPointerType(ElementType LLVMTypeRef, AddressSpace uint) LLVMTypeRef { return 0 }
func LLVMGetElementType(Ty LLVMTypeRef) LLVMTypeRef { return 0 }
func LLVMGetPointerAddressSpace(PointerTy LLVMTypeRef) uint { return 0 }
func LLVMStructType(ElementTypes []LLVMTypeRef, ElementCount uint, Packed LLVMBool) LLVMTypeRef { return 0 }
func LLVMStructTypeInContext(C LLVMContextRef, ElementTypes []LLVMTypeRef, ElementCount uint, Packed LLVMBool) LLVMTypeRef { return 0 }
func LLVMStructCreateNamed(C LLVMContextRef, Name string) LLVMTypeRef { return 0 }
func LLVMGetStructName(Ty LLVMTypeRef) *byte { return nil }
func LLVMStructSetBody(StructTy LLVMTypeRef, ElementTypes []LLVMTypeRef, ElementCount uint, Packed LLVMBool) {}
func LLVMCountStructElementTypes(StructTy LLVMTypeRef) uint { return 0 }
func LLVMGetStructElementTypes(StructTy LLVMTypeRef) *LLVMTypeRef { return nil }
func LLVMStructGetTypeAtIndex(StructTy LLVMTypeRef, i uint) LLVMTypeRef { return 0 }
func LLVMIsPackedStruct(StructTy LLVMTypeRef) LLVMBool { return 0 }
func LLVMIsOpaqueStruct(StructTy LLVMTypeRef) LLVMBool { return 0 }
func LLVMIsLiteralStruct(StructTy LLVMTypeRef) LLVMBool { return 0 }
func LLVMGetArrayLength(Ty LLVMTypeRef) uint64 { return 0 }
func LLVMGetArrayLength2(Ty LLVMTypeRef) uint64 { return 0 }
func LLVMVectorType(ElementType LLVMTypeRef, ElementCount uint) LLVMTypeRef { return 0 }
func LLVMScalableVectorType(ElementType LLVMTypeRef, ElementCount uint) LLVMTypeRef { return 0 }
func LLVMGetVectorSize(VectorTy LLVMTypeRef) uint { return 0 }
func LLVMFunctionType(ReturnType LLVMTypeRef, ParamTypes []LLVMTypeRef, ParamCount uint, IsVarArg LLVMBool) LLVMTypeRef { return 0 }
func LLVMIsFunctionVarArg(FunctionTy LLVMTypeRef) LLVMBool { return 0 }
func LLVMGetReturnType(FunctionTy LLVMTypeRef) LLVMTypeRef { return 0 }
func LLVMCountParamTypes(FunctionTy LLVMTypeRef) uint { return 0 }
func LLVMGetParamTypes(FunctionTy LLVMTypeRef, Dest *LLVMTypeRef) {}
func LLVMGetParamTypes2(FunctionTy LLVMTypeRef, Dest *LLVMTypeRef) {}
func LLVMTargetExtType(C LLVMContextRef, Name string, TypeParams []LLVMTypeRef, IntParams []uint) LLVMTypeRef { return 0 }
func LLVMGetTargetExtName(Ty LLVMTypeRef) *byte { return nil }

func LLVMTypeOf(Val LLVMValueRef) LLVMTypeRef { return 0 }
func LLVMGetValueKind(Val LLVMValueRef) LLVMValueKind { return 0 }
func LLVMGetValueName(Val LLVMValueRef) *byte { return nil }
func LLVMGetValueName2(Val LLVMValueRef, Length *uint) *byte { return nil }
func LLVMSetValueName(Val LLVMValueRef, Name string) {}
func LLVMSetValueName2(Val LLVMValueRef, Name string, NameLen uint) {}
func LLVMDumpValue(Val LLVMValueRef) {}
func LLVMPrintValueToString(Val LLVMValueRef) *byte { return nil }
func LLVMReplaceAllUsesWith(OldVal LLVMValueRef, NewVal LLVMValueRef) {}
func LLVMIsConstant(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMIsUndef(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMIsPoison(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMIsAArgument(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsABasicBlock(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAInlineAsm(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAUser(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstant(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsABlockAddress(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantAggregateZero(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantArray(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantDataArray(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantDataVector(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantExpr(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantFP(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantInt(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantPointerNull(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantStruct(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantTokenNone(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAConstantVector(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGlobalValue(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGlobalAlias(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGlobalIFunc(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGlobalObject(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAFunction(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGlobalVariable(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAUndefValue(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAPoisonValue(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAInstruction(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAUnaryInstruction(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAAllocaInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsALoadInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAStoreInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAAtomicCmpXchgInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAAtomicRMWInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAFenceInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAGetElementPtrInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAPHIInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsACallInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsASelectInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAExtractValueInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAInsertValueInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAExtractElementInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAInsertElementInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAShuffleVectorInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsACmpInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAICmpInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAFCmpInst(Val LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsAUnreachableInst(Val LLVMValueRef) LLVMValueRef { return 0 }

func LLVMConstNull(T LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstAllOnes(T LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMGetUndef(T LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMGetPoison(T LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMIsNull(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMConstPointerNull(T LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstInt(IntTy LLVMTypeRef, N uint64, SignExtend LLVMBool) LLVMValueRef { return 0 }
func LLVMConstIntOfArbitraryPrecision(IntTy LLVMTypeRef, Words []uint64) LLVMValueRef { return 0 }
func LLVMConstIntOfString(IntTy LLVMTypeRef, Text string, Radix uint8) LLVMValueRef { return 0 }
func LLVMConstIntOfStringAndSize(IntTy LLVMTypeRef, Text string, SLen uint, Radix uint8) LLVMValueRef { return 0 }
func LLVMConstIntGetZExtValue(ConstantVal LLVMValueRef) uint64 { return 0 }
func LLVMConstIntGetSExtValue(ConstantVal LLVMValueRef) int64 { return 0 }
func LLVMConstReal(RealTy LLVMTypeRef, N float64) LLVMValueRef { return 0 }
func LLVMConstRealOfString(RealTy LLVMTypeRef, Text string) LLVMValueRef { return 0 }
func LLVMConstRealOfStringAndSize(RealTy LLVMTypeRef, Text string, SLen uint) LLVMValueRef { return 0 }
func LLVMConstRealGetDouble(ConstantVal LLVMValueRef, losesInfo *LLVMBool) float64 { return 0 }
func LLVMConstStringInContext(C LLVMContextRef, Str string, Length uint, DontNullTerminate LLVMBool) LLVMValueRef { return 0 }
func LLVMConstString(Str string, Length uint, DontNullTerminate LLVMBool) LLVMValueRef { return 0 }
func LLVMConstStructInContext(C LLVMContextRef, ConstantVals []LLVMValueRef, Count uint, Packed LLVMBool) LLVMValueRef { return 0 }
func LLVMConstStruct(ConstantVals []LLVMValueRef, Count uint, Packed LLVMBool) LLVMValueRef { return 0 }
func LLVMConstArray(ElementTy LLVMTypeRef, ConstantVals []LLVMValueRef, Length uint) LLVMValueRef { return 0 }
func LLVMConstArray2(ElementTy LLVMTypeRef, ConstantVals []LLVMValueRef, Length uint) LLVMValueRef { return 0 }
func LLVMConstNamedStruct(StructTy LLVMTypeRef, ConstantVals []LLVMValueRef, Count uint) LLVMValueRef { return 0 }
func LLVMConstVector(ScalarConstantVals []LLVMValueRef, Size uint) LLVMValueRef { return 0 }
func LLVMGetConstOpcode(ConstantVal LLVMValueRef) LLVMOpcode { return 0 }
func LLVMGetNumOperands(Val LLVMValueRef) uint { return 0 }
func LLVMGetOperand(Val LLVMValueRef, Index uint) LLVMValueRef { return 0 }
func LLVMGetOperandUse(Val LLVMValueRef, Index uint) LLVMUseRef { return 0 }
func LLVMSetOperand(User LLVMValueRef, Index uint, Val LLVMValueRef) {}
func LLVMConstNeg(ConstantVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNSWNeg(ConstantVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNUWNeg(ConstantVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFNeg(ConstantVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNot(ConstantVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstAdd(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNSWAdd(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNUWAdd(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFAdd(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstSub(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNSWSub(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNUWSub(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFSub(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstMul(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNSWMul(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstNUWMul(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFMul(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstUDiv(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstExactUDiv(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstSDiv(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstExactSDiv(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFDiv(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstURem(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstSRem(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFRem(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstAnd(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstOr(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstXor(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstICmp(Predicate LLVMIntPredicate, LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstFCmp(Predicate LLVMRealPredicate, LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstShl(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstLShr(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstAShr(LHSConstant LLVMValueRef, RHSConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstGEP(Ty LLVMTypeRef, ConstantVal LLVMValueRef, ConstantIndices []LLVMValueRef, NumIndices uint) LLVMValueRef { return 0 }
func LLVMConstGEP2(Ty LLVMTypeRef, ConstantVal LLVMValueRef, ConstantIndices []LLVMValueRef, NumIndices uint) LLVMValueRef { return 0 }
func LLVMConstInBoundsGEP(Ty LLVMTypeRef, ConstantVal LLVMValueRef, ConstantIndices []LLVMValueRef, NumIndices uint) LLVMValueRef { return 0 }
func LLVMConstInBoundsGEP2(Ty LLVMTypeRef, ConstantVal LLVMValueRef, ConstantIndices []LLVMValueRef, NumIndices uint) LLVMValueRef { return 0 }
func LLVMConstTrunc(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstSExt(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstZExt(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstFPTrunc(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstFPExt(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstUIToFP(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstSIToFP(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstFPToUI(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstFPToSI(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstPtrToInt(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstIntToPtr(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstBitCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstAddrSpaceCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstZExtOrBitCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstSExtOrBitCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstTruncOrBitCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstPointerCast(ConstantVal LLVMValueRef, ToType LLVMTypeRef) LLVMValueRef { return 0 }
func LLVMConstExtractElement(VectorConstant LLVMValueRef, IndexConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstInsertElement(VectorConstant LLVMValueRef, ElementValueConstant LLVMValueRef, IndexConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstShuffleVector(VectorAConstant LLVMValueRef, VectorBConstant LLVMValueRef, MaskConstant LLVMValueRef) LLVMValueRef { return 0 }
func LLVMConstExtractValue(AggConstant LLVMValueRef, IdxList []uint, NumIdx uint) LLVMValueRef { return 0 }
func LLVMConstInsertValue(AggConstant LLVMValueRef, ElementValueConstant LLVMValueRef, IdxList []uint, NumIdx uint) LLVMValueRef { return 0 }
func LLVMBlockAddress(F LLVMValueRef, BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMConstPtrAuth(ConstantVal LLVMValueRef, Key uint, Discriminator LLVMValueRef, IsaPointer LLVMValueRef, AuthenticatedPointer LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetGlobalParent(GlobalVar LLVMValueRef) LLVMModuleRef { return 0 }
func LLVMIsDeclaration2(GlobalVar LLVMValueRef) LLVMBool { return 0 }
func LLVMGetLinkage(Global LLVMValueRef) LLVMLinkage { return 0 }
func LLVMSetLinkage(Global LLVMValueRef, Linkage LLVMLinkage) {}
func LLVMGetSection(Global LLVMValueRef) *byte { return nil }
func LLVMSetSection(Global LLVMValueRef, Section string) {}
func LLVMGetVisibility(Global LLVMValueRef) LLVMVisibility { return 0 }
func LLVMSetVisibility(Global LLVMValueRef, Viz LLVMVisibility) {}
func LLVMGetUnnamedAddress(Global LLVMValueRef) LLVMUnnamedAddr { return 0 }
func LLVMSetUnnamedAddress(Global LLVMValueRef, UnnamedAddr LLVMUnnamedAddr) {}
func LLVMGetDLLStorageClass(Global LLVMValueRef) LLVMDLLStorageClass { return 0 }
func LLVMSetDLLStorageClass(Global LLVMValueRef, StorageClass LLVMDLLStorageClass) {}
func LLVMHasUnnamedAddr(Global LLVMValueRef) LLVMBool { return 0 }
func LLVMSetUnnamedAddr(Global LLVMValueRef, HasUnnamedAddr LLVMBool) {}
func LLVMGetAlignment(Global LLVMValueRef) uint { return 0 }
func LLVMSetAlignment(Global LLVMValueRef, Bytes uint) {}
func LLVMGlobalGetMetadata(Global LLVMValueRef, Kind uint) LLVMMetadataRef { return 0 }
func LLVMGlobalSetMetadata(Global LLVMValueRef, Kind uint, MD LLVMMetadataRef) {}
func LLVMGlobalEraseMetadata(Global LLVMValueRef, Kind uint) {}
func LLVMGlobalClearMetadata(Global LLVMValueRef) {}
func LLVMGlobalCopyAllMetadata(Global LLVMValueRef, NumEntries *uint) *LLVMValueMetadataEntry { return nil }
func LLVMHasPersonalityFn(Fn LLVMValueRef) LLVMBool { return 0 }
func LLVMGetPersonalityFn(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetPersonalityFn(Fn LLVMValueRef, PersonalityFn LLVMValueRef) {}
func LLVMGetIntrinsicID(Fn LLVMValueRef) uint { return 0 }
func LLVMGetFunctionCallConv(Fn LLVMValueRef) uint { return 0 }
func LLVMSetFunctionCallConv(Fn LLVMValueRef, CallConv uint) {}
func LLVMGetGC(Fn LLVMValueRef) *byte { return nil }
func LLVMSetGC(Fn LLVMValueRef, Name string) {}
func LLVMGetPrefixData(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetPrefixData(Fn LLVMValueRef, PrefixData LLVMValueRef) {}
func LLVMGetPrologueData(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetPrologueData(Fn LLVMValueRef, PrologueData LLVMValueRef) {}
func LLVMAddAttributeAtIndex(F LLVMValueRef, Idx LLVMAttributeIndex, A LLVMAttributeRef) {}
func LLVMGetAttributeAtIndex(F LLVMValueRef, Idx LLVMAttributeIndex) LLVMAttributeRef { return 0 }
func LLVMGetAttributesAtIndex(F LLVMValueRef, Idx LLVMAttributeIndex) *LLVMAttributeRef { return nil }
func LLVMGetAttributesAtIndexCount(F LLVMValueRef, Idx LLVMAttributeIndex) uint { return 0 }
func LLVMRemoveEnumAttributeAtIndex(F LLVMValueRef, Idx LLVMAttributeIndex, KindID uint) {}
func LLVMRemoveStringAttributeAtIndex(F LLVMValueRef, Idx LLVMAttributeIndex, K string, KLength uint) {}
func LLVMAddCallSiteAttribute(C LLVMValueRef, Idx LLVMAttributeIndex, A LLVMAttributeRef) {}
func LLVMGetCallSiteAttribute(C LLVMValueRef, Idx LLVMAttributeIndex) LLVMAttributeRef { return 0 }
func LLVMGetCallSiteAttributes(C LLVMValueRef, Idx LLVMAttributeIndex) *LLVMAttributeRef { return nil }
func LLVMGetCallSiteAttributeCount(C LLVMValueRef, Idx LLVMAttributeIndex) uint { return 0 }
func LLVMRemoveCallSiteEnumAttribute(C LLVMValueRef, Idx LLVMAttributeIndex, KindID uint) {}
func LLVMRemoveCallSiteStringAttribute(C LLVMValueRef, Idx LLVMAttributeIndex, K string, KLength uint) {}
func LLVMGetParamAlignment(Fn LLVMValueRef, ParamIndex uint) uint { return 0 }
func LLVMCountParams(Fn LLVMValueRef) uint { return 0 }
func LLVMGetParams(Fn LLVMValueRef, Params *LLVMValueRef) {}
func LLVMGetParam(Fn LLVMValueRef, Index uint) LLVMValueRef { return 0 }
func LLVMGetParamParent(Inst LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetFirstParam(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetLastParam(Fn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetNextParam(Arg LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetPreviousParam(Arg LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetParamAlignment(Arg LLVMValueRef, Align uint) {}
func LLVMAddGlobalIFunc(M LLVMModuleRef, Name string, AddressSpace uint, Resolver LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetGlobalIFuncResolver(IFunc LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetGlobalIFuncResolver(IFunc LLVMValueRef, Resolver LLVMValueRef) {}
func LLVMGlobalIFuncGetParent(IFunc LLVMValueRef) LLVMModuleRef { return 0 }
func LLVMGlobalAliasGetParent(Alias LLVMValueRef) LLVMModuleRef { return 0 }
func LLVMGlobalAliasGetAliasee(Alias LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetGlobalAliasAliasee(Alias LLVMValueRef, Aliasee LLVMValueRef) {}
func LLVMIsGlobalConstant(GlobalVar LLVMValueRef) LLVMBool { return 0 }
func LLVMSetGlobalConstant(GlobalVar LLVMValueRef, IsConstant LLVMBool) {}
func LLVMGetThreadLocalMode(GlobalVar LLVMValueRef) LLVMThreadLocalMode { return 0 }
func LLVMSetThreadLocalMode(GlobalVar LLVMValueRef, Mode LLVMThreadLocalMode) {}
func LLVMIsExternallyInitialized(GlobalVar LLVMValueRef) LLVMBool { return 0 }
func LLVMSetExternallyInitialized(GlobalVar LLVMValueRef, IsExtInit LLVMBool) {}
func LLVMHasMetadata(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMGetMetadata(Val LLVMValueRef, KindID uint) LLVMValueRef { return 0 }
func LLVMSetMetadata(Val LLVMValueRef, KindID uint, Node LLVMValueRef) {}
func LLVMGetMetadata2(Val LLVMValueRef, KindID uint) LLVMMetadataRef { return 0 }
func LLVMSetMetadata2(Val LLVMValueRef, KindID uint, MDNode LLVMMetadataRef) {}
func LLVMBasicBlockAsValue(BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMValueIsBasicBlock(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMValueAsBasicBlock(Val LLVMValueRef) LLVMBasicBlockRef { return 0 }
func LLVMCreateBasicBlockInContext(C LLVMContextRef, Name string) LLVMBasicBlockRef { return 0 }
func LLVMAppendExistingBasicBlock(Fn LLVMValueRef, BB LLVMBasicBlockRef) {}
func LLVMGetBasicBlockName(BB LLVMBasicBlockRef) *byte { return nil }
func LLVMGetBasicBlockParent(BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMGetBasicBlockTerminator(BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMCountBasicBlocks(Fn LLVMValueRef) uint { return 0 }
func LLVMGetBasicBlocks(Fn LLVMValueRef, BasicBlocks *LLVMBasicBlockRef) {}
func LLVMGetEntryBasicBlock(Fn LLVMValueRef) LLVMBasicBlockRef { return 0 }
func LLVMGetFirstBasicBlock(Fn LLVMValueRef) LLVMBasicBlockRef { return 0 }
func LLVMGetLastBasicBlock(Fn LLVMValueRef) LLVMBasicBlockRef { return 0 }
func LLVMGetNextBasicBlock(BB LLVMBasicBlockRef) LLVMBasicBlockRef { return 0 }
func LLVMGetPreviousBasicBlock(BB LLVMBasicBlockRef) LLVMBasicBlockRef { return 0 }
func LLVMGetFirstInstruction(BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMGetLastInstruction(BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMHasNUses(Val LLVMValueRef) LLVMBool { return 0 }
func LLVMGetNumUses(Val LLVMValueRef) uint { return 0 }
func LLVMGetNextInstruction(Inst LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetPreviousInstruction(Inst LLVMValueRef) LLVMValueRef { return 0 }
func LLVMIsATerminatorInst(Inst LLVMValueRef) LLVMValueRef { return 0 }
func LLVMSetInstructionCallConv(Instr LLVMValueRef, CC uint) {}
func LLVMGetInstructionCallConv(Instr LLVMValueRef) uint { return 0 }
func LLVMGetTailCallKind(Instr LLVMValueRef) LLVMTailCallKind { return 0 }
func LLVMSetTailCallKind(Instr LLVMValueRef, TCK LLVMTailCallKind) {}
func LLVMGetNumArgOperands(Instr LLVMValueRef) uint { return 0 }
func LLVMGetArgOperand(Instr LLVMValueRef, i uint) LLVMValueRef { return 0 }
func LLVMSetArgOperand(Instr LLVMValueRef, i uint, Arg LLVMValueRef) {}
func LLVMGetNumOperandBundles(Instr LLVMValueRef) uint { return 0 }
func LLVMGetOperandBundle(Instr LLVMValueRef, Index uint) LLVMOperandBundleRef { return 0 }
func LLVMOperandBundleGetTag(OpBundle LLVMOperandBundleRef) *byte { return nil }
func LLVMOperandBundleGetTagLen(OpBundle LLVMOperandBundleRef) uint { return 0 }
func LLVMOperandBundleGetNumInputs(OpBundle LLVMOperandBundleRef) uint { return 0 }
func LLVMGetInstrProfWeight(Inst LLVMValueRef) uint64 { return 0 }
func LLVMSetInstrProfWeight(Inst LLVMValueRef, W uint64) {}
func LLVMGetDebugLoc(Inst LLVMValueRef) LLVMMetadataRef { return 0 }
func LLVMSetDebugLoc(Inst LLVMValueRef, Loc LLVMMetadataRef) {}
func LLVMGetFirstDbgRecord(BB LLVMBasicBlockRef) LLVMDbgRecordRef { return 0 }
func LLVMGetLastDbgRecord(BB LLVMBasicBlockRef) LLVMDbgRecordRef { return 0 }
func LLVMGetNextDbgRecord(DR LLVMDbgRecordRef) LLVMDbgRecordRef { return 0 }
func LLVMGetPreviousDbgRecord(DR LLVMDbgRecordRef) LLVMDbgRecordRef { return 0 }
func LLVMGetDbgRecordInst(DR LLVMDbgRecordRef) LLVMValueRef { return 0 }
func LLVMDbgRecordIsFirstOfInst(DR LLVMDbgRecordRef) LLVMBool { return 0 }
func LLVMDbgRecordIsLastOfInst(DR LLVMDbgRecordRef) LLVMBool { return 0 }
func LLVMGetInstructionDebugLoc(Inst LLVMValueRef) LLVMMetadataRef { return 0 }
func LLVMSetInstructionDebugLoc(Inst LLVMValueRef, Loc LLVMMetadataRef) {}
func LLVMGetNumIndices(Inst LLVMValueRef) uint { return 0 }
func LLVMGetIndices(Inst LLVMValueRef) *uint { return nil }
func LLVMGetInstructionOpcode(Inst LLVMValueRef) LLVMOpcode { return 0 }
func LLVMGetICmpPredicate(Inst LLVMValueRef) LLVMIntPredicate { return 0 }
func LLVMGetFCmpPredicate(Inst LLVMValueRef) LLVMRealPredicate { return 0 }
func LLVMInstructionClone(Inst LLVMValueRef) LLVMValueRef { return 0 }
func LLVMGetFirstUse(Val LLVMValueRef) LLVMUseRef { return 0 }
func LLVMGetNextUse(U LLVMUseRef) LLVMUseRef { return 0 }
func LLVMGetUser(U LLVMUseRef) LLVMValueRef { return 0 }
func LLVMGetUsedValue(U LLVMUseRef) LLVMValueRef { return 0 }
func LLVMCreateBuilder() LLVMBuilderRef { return 0 }
func LLVMCreateBuilderInContext(C LLVMContextRef) LLVMBuilderRef { return 0 }
func LLVMPositionBuilder(Builder LLVMBuilderRef, Block LLVMBasicBlockRef, Instr LLVMValueRef) {}
func LLVMPositionBuilderBefore(Builder LLVMBuilderRef, Instr LLVMValueRef) {}
func LLVMPositionBuilderAtEnd(Builder LLVMBuilderRef, Block LLVMBasicBlockRef) {}
func LLVMGetInsertBlock(Builder LLVMBuilderRef) LLVMBasicBlockRef { return 0 }
func LLVMClearInsertionPosition(Builder LLVMBuilderRef) {}
func LLVMInsertIntoBuilder(Builder LLVMBuilderRef, Instr LLVMValueRef) {}
func LLVMInsertIntoBuilderBefore(Builder LLVMBuilderRef, Instr LLVMValueRef, BeforeInst LLVMValueRef) {}
func LLVMBuilderGetCurrentDebugLocation2(Builder LLVMBuilderRef) LLVMMetadataRef { return 0 }
func LLVMSetCurrentDebugLocation2(Builder LLVMBuilderRef, L LLVMMetadataRef) {}
func LLVMBuilderGetDefaultFPMathTag(Builder LLVMBuilderRef) LLVMMetadataRef { return 0 }
func LLVMSetCurrentDefaultFPMathTag(Builder LLVMBuilderRef, FPMathTag LLVMMetadataRef) {}
func LLVMDisposeBuilder(Builder LLVMBuilderRef) {}
func LLVMBuildAlloca(B LLVMBuilderRef, Ty LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildAllocaWithAlignment(B LLVMBuilderRef, Ty LLVMTypeRef, Align uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildArrayAlloca(B LLVMBuilderRef, Ty LLVMTypeRef, Size LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildArrayAllocaWithAlignment(B LLVMBuilderRef, Ty LLVMTypeRef, Size LLVMValueRef, Align uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildFree(B LLVMBuilderRef, PointerVal LLVMValueRef) LLVMValueRef { return 0 }
func LLVMBuildLoad(B LLVMBuilderRef, Ty LLVMTypeRef, PointerVal LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildLoad2(B LLVMBuilderRef, T LLVMTypeRef, PointerVal LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildStore(B LLVMBuilderRef, Val LLVMValueRef, Ptr LLVMValueRef) LLVMValueRef { return 0 }
func LLVMBuildGEP(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Indices []LLVMValueRef, NumIndices uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildGEP2(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Indices []LLVMValueRef, NumIndices uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildInBoundsGEP(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Indices []LLVMValueRef, NumIndices uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildInBoundsGEP2(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Indices []LLVMValueRef, NumIndices uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildStructGEP(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Index uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildStructGEP2(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Index uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildGlobalString(B LLVMBuilderRef, Str string, Name string) LLVMValueRef { return 0 }
func LLVMBuildGlobalStringPtr(B LLVMBuilderRef, Str string, Name string) LLVMValueRef { return 0 }
func LLVMGetVolatile(MemoryAccessInst LLVMValueRef) LLVMBool { return 0 }
func LLVMSetVolatile(MemoryAccessInst LLVMValueRef, IsVolatile LLVMBool) {}
func LLVMGetWeak(Inst LLVMValueRef) LLVMBool { return 0 }
func LLVMSetWeak(Inst LLVMValueRef, IsWeak LLVMBool) {}
func LLVMBuildTrunc(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildZExt(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSExt(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFPToUI(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFPToSI(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildUIToFP(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSIToFP(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFPTrunc(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFPExt(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildPtrToInt(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildIntToPtr(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildBitCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildAddrSpaceCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildZExtOrBitCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSExtOrBitCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildTruncOrBitCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildPointerCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildIntCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildIntCast2(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, IsSigned LLVMBool, Name string) LLVMValueRef { return 0 }
func LLVMBuildFPCast(B LLVMBuilderRef, Val LLVMValueRef, DestTy LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildIsNull(B LLVMBuilderRef, Val LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildIsNotNull(B LLVMBuilderRef, Val LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildPtrDiff(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildPtrDiff2(B LLVMBuilderRef, ElemTy LLVMTypeRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildAdd(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNSWAdd(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNUWAdd(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFAdd(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSub(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNSWSub(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNUWSub(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFSub(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildMul(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNSWMul(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNUWMul(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFMul(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildUDiv(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildExactUDiv(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSDiv(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildExactSDiv(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFDiv(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildURem(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildSRem(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFRem(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildShl(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildLShr(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildAShr(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildAnd(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildOr(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildXor(B LLVMBuilderRef, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildBinOp(B LLVMBuilderRef, Op LLVMOpcode, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNeg(B LLVMBuilderRef, V LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNSWNeg(B LLVMBuilderRef, V LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNUWNeg(B LLVMBuilderRef, V LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFNeg(B LLVMBuilderRef, V LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildNot(B LLVMBuilderRef, V LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildICmp(B LLVMBuilderRef, Op LLVMIntPredicate, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildFCmp(B LLVMBuilderRef, Op LLVMRealPredicate, LHS LLVMValueRef, RHS LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildPhi(B LLVMBuilderRef, Ty LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildCall(B LLVMBuilderRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCall2(B LLVMBuilderRef, FnTy LLVMTypeRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildSelect(B LLVMBuilderRef, If LLVMValueRef, Then LLVMValueRef, Else LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildVAArg(B LLVMBuilderRef, List LLVMValueRef, Ty LLVMTypeRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildExtractElement(B LLVMBuilderRef, VecVal LLVMValueRef, Index LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildInsertElement(B LLVMBuilderRef, VecVal LLVMValueRef, EltVal LLVMValueRef, Index LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildShuffleVector(B LLVMBuilderRef, V1 LLVMValueRef, V2 LLVMValueRef, Mask LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildExtractValue(B LLVMBuilderRef, AggVal LLVMValueRef, Index uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildInsertValue(B LLVMBuilderRef, AggVal LLVMValueRef, EltVal LLVMValueRef, Index uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildFreeze(B LLVMBuilderRef, Val LLVMValueRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildGEPNoWrap(B LLVMBuilderRef, Ty LLVMTypeRef, Pointer LLVMValueRef, Indices []LLVMValueRef, NumIndices uint, NoWrapFlags LLVMGEPNoWrapFlags, Name string) LLVMValueRef { return 0 }
func LLVMBuildRetVoid(B LLVMBuilderRef) LLVMValueRef { return 0 }
func LLVMBuildRet(B LLVMBuilderRef, V LLVMValueRef) LLVMValueRef { return 0 }
func LLVMBuildAggregateRet(B LLVMBuilderRef, RetVals []LLVMValueRef, N uint) LLVMValueRef { return 0 }
func LLVMBuildBr(B LLVMBuilderRef, Dest LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMBuildCondBr(B LLVMBuilderRef, If LLVMValueRef, Then LLVMBasicBlockRef, Else LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMBuildSwitch(B LLVMBuilderRef, V LLVMValueRef, Else LLVMBasicBlockRef, NumCases uint) LLVMValueRef { return 0 }
func LLVMBuildIndirectBr(B LLVMBuilderRef, Addr LLVMValueRef, NumDests uint) LLVMValueRef { return 0 }
func LLVMBuildInvoke(B LLVMBuilderRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Then LLVMBasicBlockRef, Catch LLVMBasicBlockRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildInvoke2(B LLVMBuilderRef, FnTy LLVMTypeRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Then LLVMBasicBlockRef, Catch LLVMBasicBlockRef, Name string) LLVMValueRef { return 0 }
func LLVMBuildUnreachable(B LLVMBuilderRef) LLVMValueRef { return 0 }
func LLVMBuildResume(B LLVMBuilderRef, Exn LLVMValueRef) LLVMValueRef { return 0 }
func LLVMBuildLandingPad(B LLVMBuilderRef, Ty LLVMTypeRef, PersFn LLVMValueRef, NumClauses uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCleanupRet(B LLVMBuilderRef, CleanupPad LLVMValueRef, UnwindBB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMBuildCatchRet(B LLVMBuilderRef, CatchPad LLVMValueRef, BB LLVMBasicBlockRef) LLVMValueRef { return 0 }
func LLVMBuildCatchPad(B LLVMBuilderRef, ParentPad LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCleanupPad(B LLVMBuilderRef, ParentPad LLVMValueRef, Args []LLVMValueRef, NumArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCatchSwitch(B LLVMBuilderRef, ParentPad LLVMValueRef, UnwindBB LLVMBasicBlockRef, NumHandlers uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCallBr(B LLVMBuilderRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, DefaultDest LLVMBasicBlockRef, IndirectDests []LLVMBasicBlockRef, NumIndirectDests uint, IndirectArgs []LLVMValueRef, NumIndirectArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMBuildCallBr2(B LLVMBuilderRef, FnTy LLVMTypeRef, Fn LLVMValueRef, Args []LLVMValueRef, NumArgs uint, DefaultDest LLVMBasicBlockRef, IndirectDests []LLVMBasicBlockRef, NumIndirectDests uint, IndirectArgs []LLVMValueRef, NumIndirectArgs uint, Name string) LLVMValueRef { return 0 }
func LLVMIsCleanupPad(LandingPadInst LLVMValueRef) LLVMBool { return 0 }
func LLVMSetCleanupPad(LandingPadInst LLVMValueRef, Val LLVMBool) {}
func LLVMAddClause(LandingPadInst LLVMValueRef, ClauseVal LLVMValueRef) {}
func LLVMGetLandingPadType(LandingPadInst LLVMValueRef) LLVMTypeRef { return 0 }
func LLVMAddCase(SwitchInst LLVMValueRef, OnVal LLVMValueRef, Dest LLVMBasicBlockRef) {}
func LLVMAddDestination(IndirectBrInst LLVMValueRef, Dest LLVMBasicBlockRef) {}
func LLVMGetNumClauses(LandingPadInst LLVMValueRef) uint { return 0 }
func LLVMGetClause(LandingPadInst LLVMValueRef, Idx uint) LLVMValueRef { return 0 }
func LLVMAddIncoming(PhiNode LLVMValueRef, IncomingValues []LLVMValueRef, IncomingBlocks []LLVMBasicBlockRef, Count uint) {}
func LLVMCountIncoming(PhiNode LLVMValueRef) uint { return 0 }
func LLVMGetIncomingValue(PhiNode LLVMValueRef, Index uint) LLVMValueRef { return 0 }
func LLVMGetIncomingBlock(PhiNode LLVMValueRef, Index uint) LLVMBasicBlockRef { return 0 }
func LLVMBuildAtomicCmpXchg(B LLVMBuilderRef, Ptr LLVMValueRef, Cmp LLVMValueRef, New LLVMValueRef, SuccessOrdering LLVMAtomicOrdering, FailureOrdering LLVMAtomicOrdering, SingleThread LLVMBool) LLVMValueRef { return 0 }
func LLVMBuildAtomicRMW(B LLVMBuilderRef, Op LLVMAtomicRMWBinOp, Ptr LLVMValueRef, Val LLVMValueRef, Ordering LLVMAtomicOrdering, SingleThread LLVMBool) LLVMValueRef { return 0 }
func LLVMBuildFence(B LLVMBuilderRef, Ordering LLVMAtomicOrdering, SingleThread LLVMBool, Name string) LLVMValueRef { return 0 }
func LLVMGetOrdering(Inst LLVMValueRef) LLVMAtomicOrdering { return 0 }
func LLVMSetOrdering(Inst LLVMValueRef, Ordering LLVMAtomicOrdering) {}
func LLVMGetSuccessOrdering(Inst LLVMValueRef) LLVMAtomicOrdering { return 0 }
func LLVMSetSuccessOrdering(Inst LLVMValueRef, Ordering LLVMAtomicOrdering) {}
func LLVMGetFailureOrdering(Inst LLVMValueRef) LLVMAtomicOrdering { return 0 }
func LLVMSetFailureOrdering(Inst LLVMValueRef, Ordering LLVMAtomicOrdering) {}
func LLVMIsAtomicSingleThread(Inst LLVMValueRef) LLVMBool { return 0 }
func LLVMSetAtomicSingleThread(Inst LLVMValueRef, IsSingleThread LLVMBool) {}
func LLVMGetOrInsertComdat(M LLVMModuleRef, Name string) LLVMComdatRef { return 0 }
func LLVMGetComdat(Global LLVMValueRef) LLVMComdatRef { return 0 }
func LLVMSetComdat(Global LLVMValueRef, C LLVMComdatRef) {}
func LLVMCanValueUseFastMathFlags(Inst LLVMValueRef) LLVMBool { return 0 }
func LLVMGetFastMathFlags(Inst LLVMValueRef) LLVMFastMathFlags { return 0 }
func LLVMSetFastMathFlags(Inst LLVMValueRef, FMF LLVMFastMathFlags) {}
func LLVMGetFastMathFlag(Inst LLVMValueRef, Flag LLVMFastMathFlags) LLVMBool { return 0 }
func LLVMSetFastMathFlag(Inst LLVMValueRef, Flag LLVMFastMathFlags) {}
func LLVMMDStringInContext(C LLVMContextRef, Str string, SLen uint) LLVMMetadataRef { return 0 }
func LLVMMDString(Str string, SLen uint) LLVMMetadataRef { return 0 }
func LLVMMDNodeInContext(C LLVMContextRef, MDs []LLVMMetadataRef, Count uint) LLVMMetadataRef { return 0 }
func LLVMMDNode(MDs []LLVMMetadataRef, Count uint) LLVMMetadataRef { return 0 }
func LLVMMetadataAsValue(C LLVMContextRef, MD LLVMMetadataRef) LLVMValueRef { return 0 }
func LLVMValueAsMetadata(Val LLVMValueRef) LLVMMetadataRef { return 0 }
func LLVMGetMDString(MD LLVMMetadataRef, Length *uint) *byte { return nil }
func LLVMGetMDNodeNumOperands(MD LLVMMetadataRef) uint { return 0 }
func LLVMGetMDNodeOperands(MD LLVMMetadataRef, Dest *LLVMMetadataRef) {}
func LLVMReplaceMDNodeOperandWith(MD LLVMMetadataRef, Index uint, Replacement LLVMMetadataRef) {}
func LLVMMetadataReplaceAllUsesWith(MD LLVMMetadataRef, NewMD LLVMMetadataRef) {}
func LLVMInitializeAllTargetInfos() {}
func LLVMInitializeAllTargets() {}
func LLVMInitializeAllTargetMCs() {}
func LLVMInitializeAllAsmPrinters() {}
func LLVMInitializeAllAsmParsers() {}
func LLVMInitializeAllDisassemblers() {}
func LLVMInitializeNativeTarget() LLVMBool { return 0 }
func LLVMInitializeNativeAsmParser() LLVMBool { return 0 }
func LLVMInitializeNativeAsmPrinter() LLVMBool { return 0 }
func LLVMInitializeNativeDisassembler() LLVMBool { return 0 }
func LLVMCreateTargetData(StringRep string) LLVMTargetDataRef { return 0 }
func LLVMDisposeTargetData(TD LLVMTargetDataRef) {}
func LLVMAddTargetLibraryInfo(TLI LLVMTargetLibraryInfoRef, PM LLVMPassManagerRef) {}
func LLVMCopyStringRepOfTargetData(TD LLVMTargetDataRef) *byte { return nil }
func LLVMByteOrder(TD LLVMTargetDataRef) LLVMByteOrdering { return 0 }
func LLVMPointerSize(TD LLVMTargetDataRef) uint { return 0 }
func LLVMPointerSizeForAS(TD LLVMTargetDataRef, AS uint) uint { return 0 }
func LLVMIntPtrType(TD LLVMTargetDataRef) LLVMTypeRef { return 0 }
func LLVMIntPtrTypeForAS(TD LLVMTargetDataRef, AS uint) LLVMTypeRef { return 0 }
func LLVMIntPtrTypeInContext(C LLVMContextRef, TD LLVMTargetDataRef) LLVMTypeRef { return 0 }
func LLVMIntPtrTypeForASInContext(C LLVMContextRef, TD LLVMTargetDataRef, AS uint) LLVMTypeRef { return 0 }
func LLVMSizeOfTypeInBits(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint64 { return 0 }
func LLVMStoreSizeOfType(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint64 { return 0 }
func LLVMABISizeOfType(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint64 { return 0 }
func LLVMABIAlignmentOfType(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint { return 0 }
func LLVMCallFrameAlignmentOfType(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint { return 0 }
func LLVMPreferredAlignmentOfType(TD LLVMTargetDataRef, Ty LLVMTypeRef) uint { return 0 }
func LLVMPreferredAlignmentOfGlobal(TD LLVMTargetDataRef, GlobalVar LLVMValueRef) uint { return 0 }
func LLVMElementAtOffset(TD LLVMTargetDataRef, StructTy LLVMTypeRef, Offset uint64) uint { return 0 }
func LLVMOffsetOfElement(TD LLVMTargetDataRef, StructTy LLVMTypeRef, Element uint) uint64 { return 0 }
func LLVMGetTargetFromName(Name string) LLVMTargetRef { return 0 }
func LLVMGetFirstTarget() LLVMTargetRef { return 0 }
func LLVMGetLastTarget() LLVMTargetRef { return 0 }
func LLVMGetPreviousTarget(T LLVMTargetRef) LLVMTargetRef { return 0 }
func LLVMGetNextTarget(T LLVMTargetRef) LLVMTargetRef { return 0 }
func LLVMGetTargetName(T LLVMTargetRef) *byte { return nil }
func LLVMGetTargetDescription(T LLVMTargetRef) *byte { return nil }
func LLVMTargetHasTargetMachine(T LLVMTargetRef) LLVMBool { return 0 }
func LLVMTargetHasAsmBackend(T LLVMTargetRef) LLVMBool { return 0 }
func LLVMHasDefaultTriple() LLVMBool { return 0 }
func LLVMDefaultTargetTriple() *byte { return nil }
func LLVMNormalizeTargetTriple(triple string) *byte { return nil }
func LLVMGetHostCPUName() *byte { return nil }
func LLVMGetHostCPUFeatures() *byte { return nil }
func LLVMAddTargetData(TD LLVMTargetDataRef, PM LLVMPassManagerRef) {}
func LLVMCreateTargetMachine(T LLVMTargetRef, Triple string, CPU string, Features string, Level LLVMCodeGenOptLevel, Reloc LLVMRelocMode, CodeModel LLVMCodeModel) LLVMTargetMachineRef { return 0 }
func LLVMCreateTargetMachineWithOptions(T LLVMTargetRef, Triple string, CPU string, Features string, Options LLVMTargetMachineOptionsRef) LLVMTargetMachineRef { return 0 }
func LLVMDisposeTargetMachine(TM LLVMTargetMachineRef) {}
func LLVMGetTargetMachineTarget(TM LLVMTargetMachineRef) LLVMTargetRef { return 0 }
func LLVMGetTargetMachineTriple(TM LLVMTargetMachineRef) *byte { return nil }
func LLVMGetTargetMachineCPU(TM LLVMTargetMachineRef) *byte { return nil }
func LLVMGetTargetMachineFeatureString(TM LLVMTargetMachineRef) *byte { return nil }
func LLVMCreateTargetDataLayout(TM LLVMTargetMachineRef) LLVMTargetDataRef { return 0 }
func LLVMSetTargetMachineAsmVerbosity(TM LLVMTargetMachineRef, Verbose LLVMBool) {}
func LLVMTargetMachineEmitToFile(TM LLVMTargetMachineRef, M LLVMModuleRef, Filename string, codegen LLVMCodeGenFileType, ErrorMessage *string) LLVMBool { return 0 }
func LLVMTargetMachineEmitToMemoryBuffer(TM LLVMTargetMachineRef, M LLVMModuleRef, codegen LLVMCodeGenFileType, ErrorMessage *string) LLVMMemoryBufferRef { return 0 }
func LLVMGetDefaultTargetTriple() *byte { return nil }
func LLVMAddAnalysisPasses(TM LLVMTargetMachineRef, PM LLVMPassManagerRef) {}
func LLVMCreateTargetMachineOptions() LLVMTargetMachineOptionsRef { return 0 }
func LLVMTargetMachineOptionsSetCodeGenOptLevel(Options LLVMTargetMachineOptionsRef, Level LLVMCodeGenOptLevel) {}
func LLVMTargetMachineOptionsSetRelocMode(Options LLVMTargetMachineOptionsRef, Reloc LLVMRelocMode) {}
func LLVMTargetMachineOptionsSetCodeModel(Options LLVMTargetMachineOptionsRef, CodeModel LLVMCodeModel) {}
func LLVMTargetMachineOptionsSetGlobalISelAbortMode(Options LLVMTargetMachineOptionsRef, Mode LLVMGlobalISelAbortMode) {}
func LLVMTargetMachineOptionsSetUseJIT(Options LLVMTargetMachineOptionsRef, UseJIT LLVMBool) {}
func LLVMTargetMachineOptionsSetMCJITMemoryManager(Options LLVMTargetMachineOptionsRef, MM LLVMMCJITMemoryManagerRef) {}
func LLVMDisposeTargetMachineOptions(Options LLVMTargetMachineOptionsRef) {}
func LLVMLinkInMCJIT() {}
func LLVMLinkInInterpreter() {}
func LLVMCreateGenericValueOfInt(Ty LLVMTypeRef, N uint64, IsSigned LLVMBool) LLVMGenericValueRef { return 0 }
func LLVMCreateGenericValueOfFloat(Ty LLVMTypeRef, N float64) LLVMGenericValueRef { return 0 }
func LLVMGenericValueIntWidth(GenValRef LLVMGenericValueRef) uint { return 0 }
func LLVMGenericValueToInt(GenVal LLVMGenericValueRef, IsSigned LLVMBool) uint64 { return 0 }
func LLVMGenericValueToFloat(Ty LLVMTypeRef, GenVal LLVMGenericValueRef) float64 { return 0 }
func LLVMGenericValueToPointer(GenVal LLVMGenericValueRef) uintptr { return 0 }
func LLVMDisposeGenericValue(GenVal LLVMGenericValueRef) {}
func LLVMCreateExecutionEngineForModule(OutEE *LLVMExecutionEngineRef, M LLVMModuleRef, OutError *string) LLVMBool { return 0 }
func LLVMCreateInterpreterForModule(OutInterp *LLVMExecutionEngineRef, M LLVMModuleRef, OutError *string) LLVMBool { return 0 }
func LLVMCreateJITCompilerForModule(OutJIT *LLVMExecutionEngineRef, M LLVMModuleRef, OptLevel uint, OutError *string) LLVMBool { return 0 }
func LLVMInitializeMCJITCompilerOptions(Options *LLVMMCJITCompilerOptions, SizeOfOptions uint) {}
func LLVMCreateMCJITCompilerForModule(OutJIT *LLVMExecutionEngineRef, M LLVMModuleRef, Options *LLVMMCJITCompilerOptions, SizeOfOptions uint, OutError *string) LLVMBool { return 0 }
func LLVMDisposeExecutionEngine(EE LLVMExecutionEngineRef) {}
func LLVMRunStaticConstructors(EE LLVMExecutionEngineRef) {}
func LLVMRunStaticDestructors(EE LLVMExecutionEngineRef) {}
func LLVMRunFunctionAsMain(EE LLVMExecutionEngineRef, F LLVMValueRef, ArgC uint, ArgV []*byte, EnvP *byte) int { return 0 }
func LLVMRunFunction(EE LLVMExecutionEngineRef, F LLVMValueRef, NumArgs uint, Args []LLVMGenericValueRef) LLVMGenericValueRef { return 0 }
func LLVMGetPointerToGlobal(EE LLVMExecutionEngineRef, Global LLVMValueRef) uintptr { return 0 }
func LLVMGetGlobalValueAddress(EE LLVMExecutionEngineRef, Name string) uint64 { return 0 }
func LLVMGetFunctionAddress(EE LLVMExecutionEngineRef, Name string) uint64 { return 0 }
func LLVMExecutionEngineGetErrMsg(EE LLVMExecutionEngineRef) *byte { return nil }
func LLVMAddModule(EE LLVMExecutionEngineRef, M LLVMModuleRef) {}
func LLVMRemoveModule(EE LLVMExecutionEngineRef, M LLVMModuleRef, OutMod *LLVMModuleRef, OutError *string) {}
func LLVMFindFunction(EE LLVMExecutionEngineRef, Name string, OutFn *LLVMValueRef) int { return 0 }
func LLVMRecompileAndRelinkFunction(EE LLVMExecutionEngineRef, Fn LLVMValueRef) uintptr { return 0 }
func LLVMGetExecutionEngineTargetData(EE LLVMExecutionEngineRef) LLVMTargetDataRef { return 0 }
func LLVMGetExecutionEngineTargetMachine(EE LLVMExecutionEngineRef) LLVMTargetMachineRef { return 0 }
func LLVMAddGlobalMapping(EE LLVMExecutionEngineRef, Global LLVMValueRef, Addr uintptr) {}
func LLVMDefaultAsmVerbosity() LLVMBool { return 0 }
func LLVMCreateMemoryBufferWithContentsOfFile(Path string, OutMemBuf *LLVMMemoryBufferRef, OutMsg *string) LLVMBool { return 0 }
func LLVMCreateMemoryBufferWithSTDIN(OutMemBuf *LLVMMemoryBufferRef, OutMsg *string) LLVMBool { return 0 }
func LLVMCreateMemoryBufferWithMemoryRange(InputData string, InputDataLength uint, BufferName string, RequiresNullTerminator LLVMBool) LLVMMemoryBufferRef { return 0 }
func LLVMCreateMemoryBufferWithMemoryRangeCopy(InputData string, InputDataLength uint, BufferName string) LLVMMemoryBufferRef { return 0 }
func LLVMGetBufferStart(MemBuf LLVMMemoryBufferRef) *byte { return nil }
func LLVMGetBufferSize(MemBuf LLVMMemoryBufferRef) uint { return 0 }
func LLVMDisposeMemoryBuffer(MemBuf LLVMMemoryBufferRef) {}
func LLVMParseIRInContext(ContextRef LLVMContextRef, MemBuf LLVMMemoryBufferRef, OutM *LLVMModuleRef, OutMsg *string) LLVMBool { return 0 }
func LLVMParseBitcode(MemBuf LLVMMemoryBufferRef, OutModule *LLVMModuleRef, OutMsg *string) LLVMBool { return 0 }
func LLVMParseBitcode2(MemBuf LLVMMemoryBufferRef, OutModule *LLVMModuleRef) LLVMBool { return 0 }
func LLVMParseBitcodeInContext(ContextRef LLVMContextRef, MemBuf LLVMMemoryBufferRef, OutModule *LLVMModuleRef, OutMsg *string) LLVMBool { return 0 }
func LLVMParseBitcodeInContext2(ContextRef LLVMContextRef, MemBuf LLVMMemoryBufferRef, OutModule *LLVMModuleRef) LLVMBool { return 0 }
func LLVMGetBitcodeModuleInContext(ContextRef LLVMContextRef, MemBuf LLVMMemoryBufferRef, OutM *LLVMModuleRef, OutMsg *string) LLVMBool { return 0 }
func LLVMGetBitcodeModuleInContext2(ContextRef LLVMContextRef, MemBuf LLVMMemoryBufferRef, OutM *LLVMModuleRef) LLVMBool { return 0 }
func LLVMGetBitcodeModule(MemBuf LLVMMemoryBufferRef, OutM *LLVMModuleRef, OutMsg *string) LLVMBool { return 0 }
func LLVMGetBitcodeModule2(MemBuf LLVMMemoryBufferRef, OutM *LLVMModuleRef) LLVMBool { return 0 }
func LLVMRemapModule(M LLVMModuleRef, PreallocatedIdentifiedMetadata []LLVMMetadataRef, NumPreallocatedIdentifiedMetadata uint, Flags uintptr) {}
func LLVMCreateModuleProviderForExistingModule(M LLVMModuleRef) LLVMModuleProviderRef { return 0 }
func LLVMDisposeModuleProvider(MP LLVMModuleProviderRef) {}
func LLVMCreatePassManager() LLVMPassManagerRef { return 0 }
func LLVMCreateFunctionPassManager(MP LLVMModuleProviderRef) LLVMPassManagerRef { return 0 }
func LLVMRunPassManager(PM LLVMPassManagerRef, M LLVMModuleRef) LLVMBool { return 0 }
func LLVMRunFunctionPassManager(PM LLVMPassManagerRef, F LLVMValueRef) LLVMBool { return 0 }
func LLVMInitializeFunctionPassManager(PM LLVMPassManagerRef) LLVMBool { return 0 }
func LLVMFinalizeFunctionPassManager(PM LLVMPassManagerRef) LLVMBool { return 0 }
func LLVMDisposePassManager(PM LLVMPassManagerRef) {}
func LLVMStartMultithreaded() LLVMBool { return 0 }
func LLVMStopMultithreaded() {}
func LLVMIsMultithreaded() LLVMBool { return 0 }
func LLVMPrintModuleToString(M LLVMModuleRef) *byte { return nil }
func LLVMPrintModuleToFile(M LLVMModuleRef, Filename string, ErrorMessage *string) LLVMBool { return 0 }
func LLVMGetModuleInlineAsm(M LLVMModuleRef, Len *uint) *byte { return nil }
func LLVMWriteBitcodeToFile(M LLVMModuleRef, Path string) int { return 0 }
func LLVMWriteBitcodeToFD(M LLVMModuleRef, FD int, ShouldClose LLVMBool, Unbuffered LLVMBool) int { return 0 }
func LLVMWriteBitcodeToFileHandle(M LLVMModuleRef, Handle int) int { return 0 }
func LLVMWriteBitcodeToMemoryBuffer(M LLVMModuleRef) LLVMMemoryBufferRef { return 0 }
func LLVMVerifyModule(M LLVMModuleRef, Action LLVMVerifierFailureAction, OutMessage *string) LLVMBool { return 0 }
func LLVMVerifyFunction(Fn LLVMValueRef, Action LLVMVerifierFailureAction) LLVMBool { return 0 }
func LLVMViewFunctionCFG(Fn LLVMValueRef) {}
func LLVMViewFunctionCFGOnly(Fn LLVMValueRef) {}
func LLVMParseCommandLineOptions(argc int, argv []*byte, Overview string) {}
func LLVMSearchForAddressOfSymbol(symbolName string) uintptr { return 0 }
func LLVMAddSymbol(symbolName string, symbolValue uintptr) {}
func LLVMLookupSymbolOfAddress(address uintptr) *byte { return nil }
func LLVMLoadLibraryPermanently(Filename string) LLVMBool { return 0 }
func LLVMGetSectionIterator(Obj LLVMObjectFileRef) LLVMSectionIteratorRef { return 0 }
func LLVMDisposeSectionIterator(Obj LLVMObjectFileRef, SI LLVMSectionIteratorRef) {}
func LLVMMoveToNextSection(SI LLVMSectionIteratorRef) {}
func LLVMIsSectionIteratorAtEnd(Obj LLVMObjectFileRef, SI LLVMSectionIteratorRef) LLVMBool { return 0 }
func LLVMGetSectionName(SI LLVMSectionIteratorRef) *byte { return nil }
func LLVMGetSectionSize(SI LLVMSectionIteratorRef) uint64 { return 0 }
func LLVMGetSectionContents(SI LLVMSectionIteratorRef) *byte { return nil }
func LLVMGetSectionAddress(SI LLVMSectionIteratorRef) uint64 { return 0 }
func LLVMGetSectionContainsSymbol(SI LLVMSectionIteratorRef, Sym LLVMSymbolIteratorRef) LLVMBool { return 0 }
func LLVMGetRelocations(SI LLVMSectionIteratorRef) LLVMRelocationIteratorRef { return 0 }
func LLVMDisposeRelocationIterator(Obj LLVMObjectFileRef, RI LLVMRelocationIteratorRef) {}
func LLVMMoveToNextRelocation(RI LLVMRelocationIteratorRef) {}
func LLVMIsRelocationIteratorAtEnd(SI LLVMSectionIteratorRef, RI LLVMRelocationIteratorRef) LLVMBool { return 0 }
func LLVMGetRelocationOffset(RI LLVMRelocationIteratorRef) uint64 { return 0 }
func LLVMGetRelocationSymbol(RI LLVMRelocationIteratorRef) LLVMSymbolIteratorRef { return 0 }
func LLVMGetRelocationType(RI LLVMRelocationIteratorRef) uint64 { return 0 }
func LLVMGetRelocationTypeName(RI LLVMRelocationIteratorRef) *byte { return nil }
func LLVMGetRelocationValueString(RI LLVMRelocationIteratorRef) *byte { return nil }
func LLVMGetSymbolIterator(Obj LLVMObjectFileRef) LLVMSymbolIteratorRef { return 0 }
func LLVMDisposeSymbolIterator(SI LLVMSymbolIteratorRef) {}
func LLVMMoveToNextSymbol(SI LLVMSymbolIteratorRef) {}
func LLVMIsSymbolIteratorAtEnd(Obj LLVMObjectFileRef, SI LLVMSymbolIteratorRef) LLVMBool { return 0 }
func LLVMGetSymbolName(SI LLVMSymbolIteratorRef) *byte { return nil }
func LLVMGetSymbolAddress(SI LLVMSymbolIteratorRef) uint64 { return 0 }
func LLVMGetSymbolSize(SI LLVMSymbolIteratorRef) uint64 { return 0 }
func LLVMObjectFileCopySectionIterator(Obj LLVMObjectFileRef) LLVMSectionIteratorRef { return 0 }
func LLVMObjectFileCopySymbolIterator(Obj LLVMObjectFileRef) LLVMSymbolIteratorRef { return 0 }
func LLVMObjectFileDispose(Obj LLVMObjectFileRef) {}
func LLVMCreateBinary(MemBuf LLVMMemoryBufferRef, Context LLVMContextRef, ErrorMessage *string) LLVMBinaryRef { return 0 }
func LLVMBinaryCopyMemoryBuffer(Binary LLVMBinaryRef) LLVMMemoryBufferRef { return 0 }
func LLVMBinaryGetType(Binary LLVMBinaryRef) LLVMBinaryType { return 0 }
func LLVMBinaryDispose(Binary LLVMBinaryRef) {}
func LLVMCreateDebugInfoBuilder(C LLVMContextRef) LLVMDIBuilderRef { return 0 }
func LLVMDIBuilderDispose(Builder LLVMDIBuilderRef) {}
func LLVMDIBuilderCreateCompileUnit(Builder LLVMDIBuilderRef, Lang LLVMDWARFSourceLanguage, FileRef LLVMMetadataRef, Producer string, ProducerLen uint, IsOptimized LLVMBool, Flags string, FlagsLen uint, RuntimeVer uint, SplitName string, SplitNameLen uint, Kind LLVMDWARFEmissionKind, DWOId uint, SplitDebugFilename string, SplitDebugFilenameLen uint, DebugInfoForProfiling LLVMBool, SysRoot string, SysRootLen uint, SDK string, SDKLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateFile(Builder LLVMDIBuilderRef, Filename string, FilenameLen uint, Directory string, DirectoryLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateModule(Builder LLVMDIBuilderRef, ParentScope LLVMMetadataRef, Name string, NameLen uint, ConfigMacros string, ConfigMacrosLen uint, IncludePath string, IncludePathLen uint, ISysRoot string, ISysRootLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateNameSpace(Builder LLVMDIBuilderRef, ParentScope LLVMMetadataRef, Name string, NameLen uint, ExportSymbols LLVMBool) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateFunction(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, LinkageName string, LinkageNameLen uint, File LLVMMetadataRef, LineNo uint, Ty LLVMMetadataRef, IsLocalToUnit LLVMBool, IsDefinition LLVMBool, ScopeLine uint, Flags LLVMDIFlags, IsOptimized LLVMBool) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateLexicalBlock(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, File LLVMMetadataRef, Line uint, Column uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateLexicalBlockFile(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, File LLVMMetadataRef, Discriminator uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateImportedModule(Builder LLVMDIBuilderRef, Context LLVMMetadataRef, NS LLVMMetadataRef, File LLVMMetadataRef, Line uint, Name string, NameLen uint, Elements []LLVMMetadataRef, NumElements uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateImportedDeclaration(Builder LLVMDIBuilderRef, Context LLVMMetadataRef, Decl LLVMMetadataRef, File LLVMMetadataRef, Line uint, Name string, NameLen uint, Element LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateEnumerator(Builder LLVMDIBuilderRef, Name string, NameLen uint, Value int64, IsUnsigned LLVMBool) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateEnumerationType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNumber uint, SizeInBits uint64, AlignInBits uint32, Elements []LLVMMetadataRef, NumElements uint, ClassTy LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateUnionType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNumber uint, SizeInBits uint64, AlignInBits uint32, Flags LLVMDIFlags, Elements []LLVMMetadataRef, NumElements uint, RunTimeLang uint, UniqueId string, UniqueIdLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateArrayType(Builder LLVMDIBuilderRef, Size uint64, AlignInBits uint32, Ty LLVMMetadataRef, Subscripts []LLVMMetadataRef, NumSubscripts uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateVectorType(Builder LLVMDIBuilderRef, Size uint64, AlignInBits uint32, Ty LLVMMetadataRef, Subscripts []LLVMMetadataRef, NumSubscripts uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateBasicType(Builder LLVMDIBuilderRef, Name string, NameLen uint, SizeInBits uint64, Encoding LLVMDWARFTypeEncoding, Flags LLVMDIFlags) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreatePointerType(Builder LLVMDIBuilderRef, PointeeTy LLVMMetadataRef, SizeInBits uint64, AlignInBits uint32, AddressSpace uint, Name string, NameLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateStructType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNumber uint, SizeInBits uint64, AlignInBits uint32, Flags LLVMDIFlags, DerivedFrom LLVMMetadataRef, Elements []LLVMMetadataRef, NumElements uint, RunTimeLang uint, VTableHolder LLVMMetadataRef, UniqueId string, UniqueIdLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateMemberType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNo uint, SizeInBits uint64, AlignInBits uint32, OffsetInBits uint64, Flags LLVMDIFlags, Ty LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateUnspecifiedType(Builder LLVMDIBuilderRef, Name string, NameLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateStaticMemberType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNumber uint, Ty LLVMMetadataRef, Flags LLVMDIFlags, ConstantVal LLVMValueRef, AlignInBits uint32) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateObjCIVar(Builder LLVMDIBuilderRef, Name string, NameLen uint, File LLVMMetadataRef, LineNo uint, SizeInBits uint64, AlignInBits uint32, OffsetInBits uint64, Flags LLVMDIFlags, Ty LLVMMetadataRef, PropertyNode LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateObjCProperty(Builder LLVMDIBuilderRef, Name string, NameLen uint, File LLVMMetadataRef, LineNo uint, GetterName string, GetterNameLen uint, SetterName string, SetterNameLen uint, PropertyAttributes uint, Ty LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateObjectPointerType(Builder LLVMDIBuilderRef, Ty LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateQualifiedType(Builder LLVMDIBuilderRef, Tag uint, Ty LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateReferenceType(Builder LLVMDIBuilderRef, Tag uint, Ty LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateNullPtrType(Builder LLVMDIBuilderRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateSubroutineType(Builder LLVMDIBuilderRef, File LLVMMetadataRef, ParameterTypes []LLVMMetadataRef, NumParameterTypes uint, Flags LLVMDIFlags) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateMacro(Builder LLVMDIBuilderRef, ParentMacroFile LLVMMetadataRef, Line uint, RecordType LLVMDWARFMacinfoRecordType, Name string, NameLen uint, Value string, ValueLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateTempMacroFile(Builder LLVMDIBuilderRef, ParentMacroFile LLVMMetadataRef, Line uint, File LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateDebugLocation(Line uint, Column uint, Scope LLVMMetadataRef, InlinedAt LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDILocationGetLine(Location LLVMMetadataRef) uint { return 0 }
func LLVMDILocationGetColumn(Location LLVMMetadataRef) uint { return 0 }
func LLVMDILocationGetScope(Location LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDILocationGetInlinedAt(Location LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIScopeGetFile(Scope LLVMMetadataRef) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateExpression(Builder LLVMDIBuilderRef, Value []uint64, NumElements uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateConstantValueExpression(Builder LLVMDIBuilderRef, Value string, ValueLen uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateGlobalVariableExpression(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, LinkageName string, LinkageNameLen uint, File LLVMMetadataRef, LineNo uint, Ty LLVMMetadataRef, IsLocalToUnit LLVMBool, Val LLVMValueRef, DIExpr LLVMMetadataRef, Decl LLVMMetadataRef, AlignInBits uint32) LLVMMetadataRef { return 0 }
func LLVMGetSubprogram(Func LLVMValueRef) LLVMMetadataRef { return 0 }
func LLVMSetSubprogram(Func LLVMValueRef, SP LLVMMetadataRef) {}
func LLVMDIBuilderGetOrCreateTypeArray(Builder LLVMDIBuilderRef, Types []LLVMMetadataRef, Length uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderGetOrCreateSubrange(Builder LLVMDIBuilderRef, LowerBound int64, Count int64) LLVMMetadataRef { return 0 }
func LLVMDIBuilderGetOrCreateArray(Builder LLVMDIBuilderRef, Data []LLVMMetadataRef, Length uint) LLVMMetadataRef { return 0 }
func LLVMDIBuilderInsertDeclareAtEnd(Builder LLVMDIBuilderRef, Storage LLVMValueRef, VarInfo LLVMMetadataRef, DIExpr LLVMMetadataRef, DILocation LLVMMetadataRef, Block LLVMBasicBlockRef) LLVMDbgRecordRef { return 0 }
func LLVMDIBuilderInsertValueAtEnd(Builder LLVMDIBuilderRef, Val LLVMValueRef, DIExpr LLVMMetadataRef, DILocation LLVMMetadataRef, Block LLVMBasicBlockRef) LLVMDbgRecordRef { return 0 }
func LLVMDIBuilderInsertDbgValueBefore(Builder LLVMDIBuilderRef, Val LLVMValueRef, DIExpr LLVMMetadataRef, DILocation LLVMMetadataRef, Instr LLVMValueRef) LLVMDbgRecordRef { return 0 }
func LLVMDIBuilderInsertDbgValueAtEnd(Builder LLVMDIBuilderRef, Val LLVMValueRef, DIExpr LLVMMetadataRef, DILocation LLVMMetadataRef, Block LLVMBasicBlockRef) LLVMDbgRecordRef { return 0 }
func LLVMDIBuilderInsertDeclareBefore(Builder LLVMDIBuilderRef, Storage LLVMValueRef, DIExpr LLVMMetadataRef, DILocation LLVMMetadataRef, Instr LLVMValueRef) LLVMDbgRecordRef { return 0 }
func LLVMDIBuilderCreateAutoVariable(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNo uint, Ty LLVMMetadataRef, AlwaysPreserve LLVMBool, Flags LLVMDIFlags, AlignInBits uint32) LLVMMetadataRef { return 0 }
func LLVMDIBuilderCreateParameterVariable(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, ArgNo uint, File LLVMMetadataRef, LineNo uint, Ty LLVMMetadataRef, AlwaysPreserve LLVMBool, Flags LLVMDIFlags, AlignInBits uint32) LLVMMetadataRef { return 0 }
func LLVMDIBuilderGetFlag(Builder LLVMDIBuilderRef, Name string, NameLen uint) LLVMDIFlags { return 0 }
func LLVMDIBuilderCreateReplaceableCompositeType(Builder LLVMDIBuilderRef, Tag uint, Name string, NameLen uint, Scope LLVMMetadataRef, File LLVMMetadataRef, Line uint, RuntimeLang uint, SizeInBits uint64, AlignInBits uint32, Flags LLVMDIFlags, UniqueIdentifier string, UniqueIdentifierLen uint) LLVMMetadataRef { return 0 }
func LLVMNewPassManager() LLVMPassManagerRef { return 0 }
func LLVMPassManagerRun(PM LLVMPassManagerRef, M LLVMModuleRef) LLVMBool { return 0 }
func LLVMPassManagerDispose(PM LLVMPassManagerRef) {}
func LLVMCreatePassBuilderOptions() LLVMPassBuilderOptionsRef { return 0 }
func LLVMPassBuilderOptionsSetVerifyEach(Options LLVMPassBuilderOptionsRef, VerifyEach LLVMBool) {}
func LLVMPassBuilderOptionsSetDebugLogging(Options LLVMPassBuilderOptionsRef, DebugLogging LLVMBool) {}
func LLVMPassBuilderOptionsSetLoopInterleaving(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetLoopVectorization(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetSLPVectorization(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetLoopUnrolling(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetForgetAllSCEVInLoopUnroll(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetLicMssaOptCap(Options LLVMPassBuilderOptionsRef, Value uint) {}
func LLVMPassBuilderOptionsSetLicMssaNoAccForPromotion(Options LLVMPassBuilderOptionsRef, Value uint) {}
func LLVMPassBuilderOptionsSetCallGraphProfile(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetMergeFunctions(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetRunPartialInlining(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetRunInliner(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetRunInlinerThreshold(Options LLVMPassBuilderOptionsRef, Threshold uint) {}
func LLVMPassBuilderOptionsSetAAPipeline(Options LLVMPassBuilderOptionsRef, AAPipeline string) {}
func LLVMPassBuilderOptionsSetGlobalISel(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetGlobalISelAbortMode(Options LLVMPassBuilderOptionsRef, Mode LLVMGlobalISelAbortMode) {}
func LLVMPassBuilderOptionsSetLoopRotate(Options LLVMPassBuilderOptionsRef, Value LLVMBool) {}
func LLVMPassBuilderOptionsSetPassPlugins(Options LLVMPassBuilderOptionsRef, Plugins string) {}
func LLVMPassBuilderOptionsSetRemarksFilename(Options LLVMPassBuilderOptionsRef, RemarksFilename string) {}
func LLVMPassBuilderOptionsSetRemarksPasses(Options LLVMPassBuilderOptionsRef, RemarksPasses string) {}
func LLVMPassBuilderOptionsSetRemarksWithHotness(Options LLVMPassBuilderOptionsRef, RemarksWithHotness LLVMBool) {}
func LLVMPassBuilderOptionsSetRemarksHotnessThreshold(Options LLVMPassBuilderOptionsRef, Threshold uint64) {}
func LLVMPassBuilderOptionsSetRemarksFormat(Options LLVMPassBuilderOptionsRef, RemarksFormat string) {}
func LLVMDisposePassBuilderOptions(Options LLVMPassBuilderOptionsRef) {}
func LLVMRunPasses(M LLVMModuleRef, Passes string, MRef LLVMTargetMachineRef, Options LLVMPassBuilderOptionsRef) LLVMBool { return 0 }
func LLVMCreateObjectFile(Obj LLVMBinaryRef) LLVMObjectFileRef { return 0 }
func LLVMDisposeObjectFile(Obj LLVMObjectFileRef) {}
func LLVMGetSections(Obj LLVMObjectFileRef) LLVMSectionIteratorRef { return 0 }
func LLVMGetSymbols(Obj LLVMObjectFileRef) LLVMSymbolIteratorRef { return 0 }
