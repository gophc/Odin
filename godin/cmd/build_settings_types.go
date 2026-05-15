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
	"",
	"windows",
	"darwin",
	"linux",
	"freebsd",
	"openbsd",
	"netbsd",
	"haiku",
	"wasi",
	"js",
	"orca",
	"freestanding",
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
	"",
	"amd64",
	"i386",
	"arm32",
	"arm64",
	"wasm32",
	"wasm64p32",
	"riscv64",
}

type TargetEndianKind uint8

const (
	TargetEndianLittle TargetEndianKind = 0
	TargetEndianBig    TargetEndianKind = 1
	TargetEndianCOUNT
)

var targetEndianNames = []String{
	"little",
	"big",
}

type TargetABIKind uint16

const (
	TargetABIDefault TargetABIKind = 0
	TargetABIWin64   TargetABIKind = 1
	TargetABISysV    TargetABIKind = 2
	TargetABICOUNT
)

var targetABINames = []String{
	"",
	"win64",
	"sysv",
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
	"",
	"BOOT_APPLICATION",
	"CONSOLE",
	"EFI_APPLICATION",
	"EFI_BOOT_SERVICE_DRIVER",
	"EFI_ROM",
	"EFI_RUNTIME_DRIVER",
	"NATIVE",
	"POSIX",
	"WINDOWS",
	"WINDOWSCE",
}

type MicroarchFeatureList struct {
	Microarch String
	Features  String
}

var ODIN_VERSION = "dev-2026-05"

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
	"",
	"iphone",
	"iphonesimulator",
	"android",
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
	"default",
	"lld",
	"radlink",
	"mold",
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

var WIN32_SEPARATOR_STRING = "\\"
var NIX_SEPARATOR_STRING = "/"
var SEPARATOR_STRING = WIN32_SEPARATOR_STRING
var WASM_MODULE_NAME_SEPARATOR = ".."

var globalModulePathSet bool
var globalModulePath String

var fullpathMutex BlockingMutex
