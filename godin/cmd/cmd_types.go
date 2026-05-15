// Depends on: common.odin (ExactValue, String, isize, BuildContext, etc.)
package cmd

type BuildFlagKind int

const (
	BuildFlagInvalid BuildFlagKind = iota
	BuildFlagHelp
	BuildFlagSingleFile
	BuildFlagOutFile
	BuildFlagOptimizationMode
	BuildFlagShowTimings
	BuildFlagShowUnused
	BuildFlagShowUnusedWithLocation
	BuildFlagShowMoreTimings
	BuildFlagShowImportGraph
	BuildFlagExportTimings
	BuildFlagExportTimingsFile
	BuildFlagExportDependencies
	BuildFlagExportDependenciesFile
	BuildFlagShowSystemCalls
	BuildFlagThreadCount
	BuildFlagKeepTempFiles
	BuildFlagCollection
	BuildFlagDefine
	BuildFlagBuildMode
	BuildFlagKeepExecutable
	BuildFlagTarget
	BuildFlagSubtarget
	BuildFlagDebug
	BuildFlagDisableAssert
	BuildFlagNoBoundsCheck
	BuildFlagNoTypeAssert
	BuildFlagNoDynamicLiterals
	BuildFlagDynamicLiterals
	BuildFlagNoCRT
	BuildFlagNoRPath
	BuildFlagNoEntryPoint
	BuildFlagLinker
	BuildFlagUseSeparateModules
	BuildFlagUseSingleModule
	BuildFlagNoThreadedChecker
	BuildFlagShowDebugMessages
	BuildFlagDidYouMeanLimit
	BuildFlagShowDefineables
	BuildFlagExportDefineables
	BuildFlagIgnoreUnusedDefineables
	BuildFlagVet
	BuildFlagVetShadowing
	BuildFlagVetUnused
	BuildFlagVetUnusedImports
	BuildFlagVetUnusedVariables
	BuildFlagVetUnusedProcedures
	BuildFlagVetUsingStmt
	BuildFlagVetUsingParam
	BuildFlagVetStyle
	BuildFlagVetSemicolon
	BuildFlagVetCast
	BuildFlagVetTabs
	BuildFlagVetPackages
	BuildFlagCustomAttribute
	BuildFlagIgnoreUnknownAttributes
	BuildFlagExtraLinkerFlags
	BuildFlagExtraAssemblerFlags
	BuildFlagMicroarch
	BuildFlagTargetFeatures
	BuildFlagStrictTargetFeatures
	BuildFlagMinimumOSVersion
	BuildFlagNoThreadLocal
	BuildFlagRelocMode
	BuildFlagDisableRedZone
	BuildFlagDisableUnwind
	BuildFlagDisallowDo
	BuildFlagDefaultToNilAllocator
	BuildFlagDefaultToPanicAllocator
	BuildFlagStrictStyle
	BuildFlagForeignKeyErrorProcedures
	BuildFlagNoRTTI
	BuildFlagDynamicMapCalls
	BuildFlagObfuscateSourceCodeLocations
	BuildFlagSourceCodeLocations
	BuildFlagCompact
	BuildFlagGlobalDefinitions
	BuildFlagGoToDefinitions
	BuildFlagShort
	BuildFlagInSourceOrder
	BuildFlagAllPackages
	BuildFlagDocFormat
	BuildFlagIgnoreWarnings
	BuildFlagWarningsAsErrors
	BuildFlagTerseErrors
	BuildFlagVerboseErrors
	BuildFlagJSONErrors
	BuildFlagErrorPosStyle
	BuildFlagMaxErrorCount
	BuildFlagMinLinkLibs
	BuildFlagPrintLinkerFlags
	BuildFlagExportLinkedLibraries
	BuildFlagIntegerDivisionByZero
	BuildFlagBuildDiagnostics
	BuildFlagInternalFastISel
	BuildFlagInternalIgnoreLazy
	BuildFlagInternalIgnoreLLVMBuild
	BuildFlagInternalIgnorePanic
	BuildFlagInternalModulePerFile
	BuildFlagInternalCached
	BuildFlagInternalNoInline
	BuildFlagInternalByValue
	BuildFlagInternalWeakMonomorphization
	BuildFlagInternalLLVMVerification
	BuildFlagInternalLLVMNoSROA
	BuildFlagInternalEnableRVO
	BuildFlagSanitize
	BuildFlagLTO
	BuildFlagIgnoreVsSearch
	BuildFlagResourceFile
	BuildFlagWindowsPdbName
	BuildFlagSubsystem
	BuildFlagAndroidKeystore
	BuildFlagAndroidKeystoreAlias
	BuildFlagAndroidKeystorePassword
	BuildFlagCOUNT
)

type BuildFlagParamKind int

const (
	BuildFlagParamNone BuildFlagParamKind = iota
	BuildFlagParamBoolean
	BuildFlagParamInteger
	BuildFlagParamFloat
	BuildFlagParamString
	BuildFlagParamCOUNT
)

type BuildFlag struct {
	Kind           BuildFlagKind
	Name           string
	ParamKind      BuildFlagParamKind
	CommandSupport uint64
	AllowMultiple  bool
}

const (
	CommandAll uint64 = 0xFFFFFFFF_FFFFFFFF
)

type StripSemicolonFile struct {
	OldFullpath       string
	OldFullpathBackup string
	NewFullpath       string
	File              *AstFile
	Written           int64
}
