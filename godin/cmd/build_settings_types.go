// Depends on: common.odin (String, StringSet, Path, Array, PtrMap, ExactValue, BigInt, gbAffinity, BlockingMutex, RecursiveMutex, isize, i64, i32, u8, u16, u32, u64, CommandKind)
package cmd

type TargetOsKind uint16

const (
	TargetOsInvalid      TargetOsKind = 0
	TargetOsWindows      TargetOsKind = 1
	TargetOsDarwin       TargetOsKind = 2
	TargetOsLinux        TargetOsKind = 3
	TargetOsFreeBSD      TargetOsKind = 4
	TargetOsOpenBSD      TargetOsKind = 5
	TargetOsNetBSD       TargetOsKind = 6
	TargetOsHaiku        TargetOsKind = 7
	TargetOsWasi         TargetOsKind = 8
	TargetOsJs           TargetOsKind = 9
	TargetOsOrca         TargetOsKind = 10
	TargetOsFreestanding TargetOsKind = 11
	TargetOsCOUNT
)

var targetOsNames = []String{
	{},
	{Data: strData("windows"), Len: isize(len("windows"))},
	{Data: strData("darwin"), Len: isize(len("darwin"))},
	{Data: strData("linux"), Len: isize(len("linux"))},
	{Data: strData("freebsd"), Len: isize(len("freebsd"))},
	{Data: strData("openbsd"), Len: isize(len("openbsd"))},
	{Data: strData("netbsd"), Len: isize(len("netbsd"))},
	{Data: strData("haiku"), Len: isize(len("haiku"))},
	{Data: strData("wasi"), Len: isize(len("wasi"))},
	{Data: strData("js"), Len: isize(len("js"))},
	{Data: strData("orca"), Len: isize(len("orca"))},
	{Data: strData("freestanding"), Len: isize(len("freestanding"))},
}

type TargetArchKind uint16

const (
	TargetArchInvalid   TargetArchKind = 0
	TargetArchAmd64     TargetArchKind = 1
	TargetArchI386      TargetArchKind = 2
	TargetArchArm32     TargetArchKind = 3
	TargetArchArm64     TargetArchKind = 4
	TargetArchWasm32    TargetArchKind = 5
	TargetArchWasm64p32 TargetArchKind = 6
	TargetArchRiscv64   TargetArchKind = 7
	TargetArchCOUNT
)

var targetArchNames = []String{
	{},
	{Data: strData("amd64"), Len: isize(len("amd64"))},
	{Data: strData("i386"), Len: isize(len("i386"))},
	{Data: strData("arm32"), Len: isize(len("arm32"))},
	{Data: strData("arm64"), Len: isize(len("arm64"))},
	{Data: strData("wasm32"), Len: isize(len("wasm32"))},
	{Data: strData("wasm64p32"), Len: isize(len("wasm64p32"))},
	{Data: strData("riscv64"), Len: isize(len("riscv64"))},
}

type TargetEndianKind uint8

const (
	TargetEndianLittle TargetEndianKind = 0
	TargetEndianBig    TargetEndianKind = 1
	TargetEndianCOUNT
)

var targetEndianNames = []String{
	{Data: strData("little"), Len: isize(len("little"))},
	{Data: strData("big"), Len: isize(len("big"))},
}

type TargetABIKind uint16

const (
	TargetABIDefault TargetABIKind = 0
	TargetABIWin64   TargetABIKind = 1
	TargetABISysV    TargetABIKind = 2
	TargetABICOUNT
)

var targetABINames = []String{
	{},
	{Data: strData("win64"), Len: isize(len("win64"))},
	{Data: strData("sysv"), Len: isize(len("sysv"))},
}

type WindowsSubsystem uint8

const (
	WindowsSubsystemUNKNOWN              WindowsSubsystem = 0
	WindowsSubsystemBOOTAPPLICATION      WindowsSubsystem = 1
	WindowsSubsystemCONSOLE              WindowsSubsystem = 2
	WindowsSubsystemEFIAPPLICATION       WindowsSubsystem = 3
	WindowsSubsystemEFIBOOTSERVICEDRIVER WindowsSubsystem = 4
	WindowsSubsystemEFIROM               WindowsSubsystem = 5
	WindowsSubsystemEFIRUNTIMEDRIVER     WindowsSubsystem = 6
	WindowsSubsystemNATIVE               WindowsSubsystem = 7
	WindowsSubsystemPOSIX                WindowsSubsystem = 8
	WindowsSubsystemWINDOWS              WindowsSubsystem = 9
	WindowsSubsystemWINDOWSCE            WindowsSubsystem = 10
	WindowsSubsystemCOUNT
)

var windowsSubsystemNames = []String{
	{},
	{Data: strData("BOOT_APPLICATION"), Len: isize(len("BOOT_APPLICATION"))},
	{Data: strData("CONSOLE"), Len: isize(len("CONSOLE"))},
	{Data: strData("EFI_APPLICATION"), Len: isize(len("EFI_APPLICATION"))},
	{Data: strData("EFI_BOOT_SERVICE_DRIVER"), Len: isize(len("EFI_BOOT_SERVICE_DRIVER"))},
	{Data: strData("EFI_ROM"), Len: isize(len("EFI_ROM"))},
	{Data: strData("EFI_RUNTIME_DRIVER"), Len: isize(len("EFI_RUNTIME_DRIVER"))},
	{Data: strData("NATIVE"), Len: isize(len("NATIVE"))},
	{Data: strData("POSIX"), Len: isize(len("POSIX"))},
	{Data: strData("WINDOWS"), Len: isize(len("WINDOWS"))},
	{Data: strData("WINDOWSCE"), Len: isize(len("WINDOWSCE"))},
}

type MicroarchFeatureList struct {
	Microarch String
	Features  String
}

var ODIN_VERSION = String{Data: strData("dev-2026-05"), Len: isize(len("dev-2026-05"))}

type TargetMetrics struct {
	Os            TargetOsKind
	Arch          TargetArchKind
	PtrSize       isize
	IntSize       isize
	MaxAlign      isize
	MaxSimdAlign  isize
	TargetTriplet String
	ABI           TargetABIKind
}

type Subtarget uint32

const (
	SubtargetDefault         Subtarget = 0
	SubtargetIPhone          Subtarget = 1
	SubtargetIPhoneSimulator Subtarget = 2
	SubtargetAndroid         Subtarget = 3
	SubtargetCOUNT
	SubtargetInvalid Subtarget = 0xFFFFFFFF
)

var subtargetStrings = []String{
	{},
	{Data: strData("iphone"), Len: isize(len("iphone"))},
	{Data: strData("iphonesimulator"), Len: isize(len("iphonesimulator"))},
	{Data: strData("android"), Len: isize(len("android"))},
}

type QueryDataSetKind int

const (
	QueryDataSetInvalid           QueryDataSetKind = 0
	QueryDataSetGlobalDefinitions QueryDataSetKind = 1
	QueryDataSetGoToDefinitions   QueryDataSetKind = 2
)

type QueryDataSetSettings struct {
	Kind    QueryDataSetKind
	Ok      bool
	Compact bool
}

type BuildModeKind int

const (
	BuildModeExecutable     BuildModeKind = 0
	BuildModeDynamicLibrary BuildModeKind = 1
	BuildModeStaticLibrary  BuildModeKind = 2
	BuildModeObject         BuildModeKind = 3
	BuildModeAssembly       BuildModeKind = 4
	BuildModeLLVMIR         BuildModeKind = 5
	BuildModeCOUNT
)

type CommandKind uint64

const (
	CommandRun            CommandKind = 1 << 0
	CommandBuild          CommandKind = 1 << 1
	CommandCheck          CommandKind = 1 << 2
	CommandDoc            CommandKind = 1 << 3
	CommandVersion        CommandKind = 1 << 4
	CommandTest           CommandKind = 1 << 5
	CommandStripSemicolon CommandKind = 1 << 6
	CommandBugReport      CommandKind = 1 << 7
	CommandBundleAndroid  CommandKind = 1 << 8
	CommandBundleMacOS    CommandKind = 1 << 9
	CommandBundleIOS      CommandKind = 1 << 10
	CommandBundleOrca     CommandKind = 1 << 11
	CommandDoesCheck      CommandKind = CommandRun | CommandBuild | CommandCheck | CommandDoc | CommandTest | CommandStripSemicolon
	CommandDoesBuild      CommandKind = CommandRun | CommandBuild | CommandTest
	CommandAll            CommandKind = 0xFFFFFFFF_FFFFFFFF
)

var odinCommandStrings = [32]string{
	"run", "build", "check", "doc", "version", "test", "strip-semicolon", "",
	"bundle android", "bundle macos", "bundle ios", "bundle orca",
}

type CmdDocFlag uint32

const (
	CmdDocFlagShort         CmdDocFlag = 1 << 0
	CmdDocFlagInSourceOrder CmdDocFlag = 1 << 1
	CmdDocFlagAllPackages   CmdDocFlag = 1 << 2
	CmdDocFlagDocFormat     CmdDocFlag = 1 << 3
)

type TimingsExportFormat int32

const (
	TimingsExportUnspecified TimingsExportFormat = 0
	TimingsExportJSON        TimingsExportFormat = 1
	TimingsExportCSV         TimingsExportFormat = 2
)

type DependenciesExportFormat int32

const (
	DependenciesExportUnspecified DependenciesExportFormat = 0
	DependenciesExportMake        DependenciesExportFormat = 1
	DependenciesExportJSON        DependenciesExportFormat = 2
)

type ErrorPosStyle int

const (
	ErrorPosStyleDefault ErrorPosStyle = 0
	ErrorPosStyleUnix    ErrorPosStyle = 1
	ErrorPosStyleCOUNT
)

type RelocMode uint8

const (
	RelocModeDefault      RelocMode = 0
	RelocModeStatic       RelocMode = 1
	RelocModePIC          RelocMode = 2
	RelocModeDynamicNoPIC RelocMode = 3
)

type BuildPath uint8

const (
	BuildPathMainPackage   BuildPath = 0
	BuildPathRC            BuildPath = 1
	BuildPathRES           BuildPath = 2
	BuildPathWinSDKBinPath BuildPath = 3
	BuildPathWinSDKUMLib   BuildPath = 4
	BuildPathWinSDKUCRTLib BuildPath = 5
	BuildPathVSEXE         BuildPath = 6
	BuildPathVSLIB         BuildPath = 7
	BuildPathOutput        BuildPath = 8
	BuildPathSymbols       BuildPath = 9
	BuildPathCOUNT
)

type VetFlags uint64

const (
	VetFlagNONE               VetFlags = 0
	VetFlagShadowing          VetFlags = 1 << 0
	VetFlagUsingStmt          VetFlags = 1 << 1
	VetFlagUsingParam         VetFlags = 1 << 2
	VetFlagStyle              VetFlags = 1 << 3
	VetFlagSemicolon          VetFlags = 1 << 4
	VetFlagUnusedVariables    VetFlags = 1 << 5
	VetFlagUnusedImports      VetFlags = 1 << 6
	VetFlagDeprecated         VetFlags = 1 << 7
	VetFlagCast               VetFlags = 1 << 8
	VetFlagTabs               VetFlags = 1 << 9
	VetFlagUnusedProcedures   VetFlags = 1 << 10
	VetFlagExplicitAllocators VetFlags = 1 << 11
	VetFlagUnused             VetFlags = VetFlagUnusedVariables | VetFlagUnusedImports
	VetFlagAll                VetFlags = VetFlagUnused | VetFlagShadowing | VetFlagUsingStmt | VetFlagDeprecated | VetFlagCast
	VetFlagUsing              VetFlags = VetFlagUsingStmt | VetFlagUsingParam
)

type OptInFeatureFlags uint64

const (
	OptInFeatureFlagNONE                         OptInFeatureFlags = 0
	OptInFeatureFlagDynamicLiterals              OptInFeatureFlags = 1 << 0
	OptInFeatureFlagGlobalContext                OptInFeatureFlags = 1 << 1
	OptInFeatureFlagIntegerDivisionByZeroTrap    OptInFeatureFlags = 1 << 2
	OptInFeatureFlagIntegerDivisionByZeroZero    OptInFeatureFlags = 1 << 3
	OptInFeatureFlagIntegerDivisionByZeroSelf    OptInFeatureFlags = 1 << 4
	OptInFeatureFlagIntegerDivisionByZeroAllBits OptInFeatureFlags = 1 << 5
	OptInFeatureFlagIntegerDivisionByZeroALL     OptInFeatureFlags = OptInFeatureFlagIntegerDivisionByZeroTrap | OptInFeatureFlagIntegerDivisionByZeroZero | OptInFeatureFlagIntegerDivisionByZeroSelf | OptInFeatureFlagIntegerDivisionByZeroAllBits
	OptInFeatureFlagForceTypeAssert              OptInFeatureFlags = 1 << 6
	OptInFeatureFlagUsingStmt                    OptInFeatureFlags = 1 << 7
)

type SanitizerFlags uint32

const (
	SanitizerFlagNONE    SanitizerFlags = 0
	SanitizerFlagAddress SanitizerFlags = 1 << 0
	SanitizerFlagMemory  SanitizerFlags = 1 << 1
	SanitizerFlagThread  SanitizerFlags = 1 << 2
)

type BuildCacheData struct {
	CRC             uint64
	CacheDir        String
	FilesPath       String
	ArgsPath        String
	EnvPath         String
	CopyAlreadyDone bool
}

type LTOKind int32

const (
	LTONone      LTOKind = 0
	LTOThin      LTOKind = 1
	LTOThinFiles LTOKind = 2
)

type LinkerChoice int32

const (
	LinkerInvalid LinkerChoice = -1
	LinkerDefault LinkerChoice = 0
	LinkerLld     LinkerChoice = 1
	LinkerRadlink LinkerChoice = 2
	LinkerMold    LinkerChoice = 3
	LinkerCOUNT
)

var linkerChoices = []String{
	{Data: strData("default"), Len: 7},
	{Data: strData("lld"), Len: 3},
	{Data: strData("radlink"), Len: 7},
	{Data: strData("mold"), Len: 4},
}

type SourceCodeLocationInfo uint8

const (
	SourceCodeLocationInfoNormal     SourceCodeLocationInfo = 0
	SourceCodeLocationInfoObfuscated SourceCodeLocationInfo = 1
	SourceCodeLocationInfoFilename   SourceCodeLocationInfo = 2
	SourceCodeLocationInfoNone       SourceCodeLocationInfo = 3
)

type IntegerDivisionByZeroKind uint8

const (
	IntegerDivisionByZeroTrap    IntegerDivisionByZeroKind = 0
	IntegerDivisionByZeroZero    IntegerDivisionByZeroKind = 1
	IntegerDivisionByZeroSelf    IntegerDivisionByZeroKind = 2
	IntegerDivisionByZeroAllBits IntegerDivisionByZeroKind = 3
)

type BuildContext struct {
	ODINOS                          String
	ODINARCH                        String
	ODINVENDOR                      String
	ODINVERSION                     String
	ODINROOT                        String
	ODINBUILDPROJECTNAME            String
	ODINWINDOWSSUBSYSTEM            WindowsSubsystem
	ODINDEBUG                       bool
	ODINDISABLEASSERT               bool
	ODINDEFAULTTONILALLOCATOR       bool
	ODINDEFAULTOPANICALLOCATOR      bool
	ODINFOREIGNERRORPROCEDURES      bool
	ODINVALGRINDSUPPORT             bool
	ODINERRORPOSSTYLE               ErrorPosStyle
	EndianKind                      TargetEndianKind
	PtrSize                         int64
	IntSize                         int64
	MaxAlign                        int64
	MaxSimdAlign                    int64
	CommandKind                     CommandKind
	Command                         String
	Metrics                         TargetMetrics
	ShowHelp                        bool
	BuildPaths                      []Path
	OutFilepath                     String
	ResourceFilepath                String
	PdbFilepath                     String
	VetFlags                        uint64
	SanitizerFlags                  uint32
	VetPackages                     StringSet
	HasResource                     bool
	LinkFlags                       String
	ExtraLinkerFlags                String
	ExtraAssemblerFlags             String
	Microarch                       String
	BuildMode                       BuildModeKind
	KeepExecutable                  bool
	GenerateDocs                    bool
	CustomOptimizationLevel         bool
	OptimizationLevel               int32
	ShowTimings                     bool
	ExportTimingsFormat             TimingsExportFormat
	ExportTimingsFile               String
	ExportDependenciesFormat        DependenciesExportFormat
	ExportDependenciesFile          String
	ShowUnused                      bool
	ShowUnusedWithLocation          bool
	ShowMoreTimings                 bool
	ShowDefineables                 bool
	ExportDefineablesFile           String
	IgnoreUnusedDefineables         bool
	ShowSystemCalls                 bool
	KeepTempFiles                   bool
	IgnoreUnknownAttributes         bool
	NoBoundsCheck                   bool
	NoTypeAssert                    bool
	DynamicLiterals                 bool
	NoOutputFiles                   bool
	NoCRT                           bool
	NoRPath                         bool
	NoEntryPoint                    bool
	NoThreadLocal                   bool
	CrossCompiling                  bool
	DifferentOS                     bool
	KeepObjectFiles                 bool
	DisallowDo                      bool
	ShowImportGraph                 bool
	IntegerDivisionByZeroBehaviour  IntegerDivisionByZeroKind
	LinkerChoice                    LinkerChoice
	CustomAttributes                StringSet
	StrictStyle                     bool
	IgnoreWarnings                  bool
	WarningsAsErrors                bool
	HideErrorLine                   bool
	TerseErrors                     bool
	JSONErrors                      bool
	HasANSITerminalColours          bool
	FastISel                        bool
	IgnoreLazy                      bool
	IgnoreLLVMBuild                 bool
	IgnorePanic                     bool
	IgnoreMicrosoftMagic            bool
	LinkerMapFile                   bool
	BuildDiagnostics                bool
	UseSingleModule                 bool
	UseSeparateModules              bool
	LTOKind                         LTOKind
	ModulePerFile                   bool
	Cached                          bool
	BuildCacheData                  BuildCacheData
	InternalNoInline                bool
	InternalByValue                 bool
	InternalWeakMonomorphization    bool
	InternalIgnoreLLVMVerification  bool
	InternalLLVMNoSROA              bool
	EnableRVO                       bool
	NoThreadedChecker               bool
	ShowDebugMessages               bool
	DidYouMeanLimit                 int
	CopyFileContents                bool
	NoRTTI                          bool
	DynamicMapCalls                 bool
	SourceCodeLocationInfo          SourceCodeLocationInfo
	MinLinkLibs                     bool
	ExportLinkedLibsPath            String
	PrintLinkerFlags                bool
	RelocMode                       RelocMode
	DisableRedZone                  bool
	DisableUnwind                   bool
	MaxErrorCount                   isize
	CmdDocFlags                     uint32
	ExtraPackages                   []String
	TestAllPackages                 bool
	Affinity                        gbAffinity
	ThreadCount                     isize
	DefinedValues                   PtrMap[string, ExactValue]
	TargetFeaturesSet               StringSet
	TargetFeaturesString            String
	StrictTargetFeatures            bool
	MinimumOSVersionString          String
	MinimumOSVersionStringGiven     bool
	ODINANDROIDAPIVALUE             int
	ODINANDROIDSDK                  String
	ODINANDROIDNDK                  String
	ODINANDROIDNDKTOOLCHAIN         String
	ODINANDROIDNDKTOOLCHAINLIB      String
	ODINANDROIDNDKTOOLCHAINLIBVALUE String
	ODINANDROIDNDKTOOLCHAINSYSROOT  String
	AndroidKeystore                 String
	AndroidKeystoreAlias            String
	AndroidKeystorePassword         String
}

var buildContext BuildContext

var selectedTargetMetrics *NamedTargetMetrics
var selectedSubtarget Subtarget

var libraryCollections []LibraryCollections

type LibraryCollections struct {
	Name String
	Path String
}

type NamedTargetMetrics struct {
	Name    String
	Metrics *TargetMetrics
}

type FindResult struct {
	WindowsSDKVersion         int
	WindowsSDKBinPath         String
	WindowsSDKUMLibraryPath   String
	WindowsSDKUCRTLibraryPath String
	VSExePath                 String
	VSLibraryPath             String
}

type VersionData struct {
	BestVersion [4]int32
	BestName    String
}

type MCFindData struct {
	FileAttributes uint32
	Filename       String
}

var WIN32_SEPARATOR_STRING = String{Data: strData("\\"), Len: 1}
var NIX_SEPARATOR_STRING = String{Data: strData("/"), Len: 1}
var SEPARATOR_STRING = WIN32_SEPARATOR_STRING
var WASM_MODULE_NAME_SEPARATOR = String{Data: strData(".."), Len: 2}

var globalModulePathSet bool
var globalModulePath String

var fullpathMutex BlockingMutex
