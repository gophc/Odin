// Part 2a: BuildFlag enums, struct, add_flag, build_param_to_exact_value, did_you_mean_flag
// Translated from src/cipp/main.i.cpp lines 264-446

import "core:strings"
import "core:fmt"

// ============================================================================
// Command bitmask constants (used by parse_build_flags)
// ============================================================================

Command_build :: u64(1) << 0
Command_run :: u64(1) << 1
Command_check :: u64(1) << 2
Command_doc :: u64(1) << 3
Command_test :: u64(1) << 4
Command_strip_semicolon :: u64(1) << 5
Command_bundle_android :: u64(1) << 6
Command_version :: u64(1) << 7
Command_bug_report :: u64(1) << 8
Command_all :: u64(0xFFFF_FFFF_FFFF_FFFF)
Command__does_build :: Command_build | Command_run | Command_test
Command__does_check :: Command_build | Command_run | Command_check | Command_strip_semicolon | Command_doc | Command_test

// ============================================================================
// ExactValue stubs (will be replaced by exact_value.odin)
// ============================================================================

ExactValueKind :: enum int {
    Invalid = 0,
    Bool,
    Integer,
    Float,
    String,
}

ExactValue :: struct {
    kind: ExactValueKind,
}

exact_value_bool :: proc(b: bool) -> ExactValue {
    return ExactValue{ExactValueKind.Bool}
}

exact_value_integer_from_string :: proc(s: string) -> ExactValue {
    return ExactValue{ExactValueKind.Integer}
}

exact_value_float_from_string :: proc(s: string) -> ExactValue {
    return ExactValue{ExactValueKind.Float}
}

exact_value_string :: proc(s: string) -> ExactValue {
    return ExactValue{ExactValueKind.String}
}

// ============================================================================
// BuildFlagKind enum (lines 264-381)
// ============================================================================

BuildFlagKind :: enum int {
    Invalid = 0,
    Help,
    SingleFile,
    OutFile,
    OptimizationMode,
    ShowTimings,
    ShowUnused,
    ShowUnusedWithLocation,
    ShowMoreTimings,
    ShowImportGraph,
    ExportTimings,
    ExportTimingsFile,
    ExportDependencies,
    ExportDependenciesFile,
    ShowSystemCalls,
    ThreadCount,
    KeepTempFiles,
    Collection,
    Define,
    BuildMode,
    KeepExecutable,
    Target,
    Subtarget,
    Debug,
    DisableAssert,
    NoBoundsCheck,
    NoTypeAssert,
    NoDynamicLiterals,
    DynamicLiterals,
    NoCRT,
    NoRPath,
    NoEntryPoint,
    Linker,
    UseSeparateModules,
    UseSingleModule,
    NoThreadedChecker,
    ShowDebugMessages,
    DidYouMeanLimit,
    ShowDefineables,
    ExportDefineables,
    IgnoreUnusedDefineables,
    Vet,
    VetShadowing,
    VetUnused,
    VetUnusedImports,
    VetUnusedVariables,
    VetUnusedProcedures,
    VetUsingStmt,
    VetUsingParam,
    VetStyle,
    VetSemicolon,
    VetCast,
    VetTabs,
    VetPackages,
    CustomAttribute,
    IgnoreUnknownAttributes,
    ExtraLinkerFlags,
    ExtraAssemblerFlags,
    Microarch,
    TargetFeatures,
    StrictTargetFeatures,
    MinimumOSVersion,
    NoThreadLocal,
    RelocMode,
    DisableRedZone,
    DisableUnwind,
    DisallowDo,
    DefaultToNilAllocator,
    DefaultToPanicAllocator,
    StrictStyle,
    ForeignErrorProcedures,
    NoRTTI,
    DynamicMapCalls,
    ObfuscateSourceCodeLocations,
    SourceCodeLocations,
    Compact,
    GlobalDefinitions,
    GoToDefinitions,
    Short,
    InSourceOrder,
    AllPackages,
    DocFormat,
    IgnoreWarnings,
    WarningsAsErrors,
    TerseErrors,
    VerboseErrors,
    JsonErrors,
    ErrorPosStyle,
    MaxErrorCount,
    MinLinkLibs,
    PrintLinkerFlags,
    ExportLinkedLibraries,
    IntegerDivisionByZero,
    BuildDiagnostics,
    InternalFastISel,
    InternalIgnoreLazy,
    InternalIgnoreLLVMBuild,
    InternalIgnorePanic,
    InternalModulePerFile,
    InternalCached,
    InternalNoInline,
    InternalByValue,
    InternalWeakMonomorphization,
    InternalLLVMVerification,
    InternalLLVMNoSROA,
    InternalEnableRVO,
    Sanitize,
    LTO,
    IgnoreVsSearch,
    ResourceFile,
    WindowsPdbName,
    Subsystem,
    AndroidKeystore,
    AndroidKeystoreAlias,
    AndroidKeystorePassword,
    COUNT,
}

// ============================================================================
// BuildFlagParamKind enum (lines 383-389)
// ============================================================================

BuildFlagParamKind :: enum int {
    None = 0,
    Boolean,
    Integer,
    Float,
    String,
    COUNT,
}

// ============================================================================
// BuildFlag struct (lines 391-397)
// ============================================================================

BuildFlag :: struct {
    kind: BuildFlagKind,
    name: string,
    param_kind: BuildFlagParamKind,
    command_support: u64,
    allow_multiple: bool,
}

// ============================================================================
// add_flag (lines 398-401)
// ============================================================================

add_flag :: proc(build_flags: ^[dynamic]BuildFlag, kind: BuildFlagKind, name: string, param_kind: BuildFlagParamKind, command_support: u64, allow_multiple: bool = false) {
    append(build_flags, BuildFlag{kind, name, param_kind, command_support, allow_multiple})
}

// ============================================================================
// build_param_to_exact_value (lines 402-435)
// ============================================================================

build_param_to_exact_value :: proc(param: string) -> ExactValue {
    if strings.equal_fold(param, "t") || strings.equal_fold(param, "true") {
        return exact_value_bool(true)
    }
    if strings.equal_fold(param, "f") || strings.equal_fold(param, "false") {
        return exact_value_bool(false)
    }

    if len(param) > 0 {
        first := param[0]
        if first == '+' || first == '-' || (first >= '0' && first <= '9') {
            if strings.contains_rune(param, '.') {
                return exact_value_float_from_string(param)
            }
            return exact_value_integer_from_string(param)
        }
    }

    // Strip surrounding quotes if present
    s := param
    if len(s) >= 2 {
        if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
            s = s[1:len(s)-1]
        }
    }

    return exact_value_string(s)
}

// ============================================================================
// did_you_mean_flag (lines 436-446)
// ============================================================================

@(private)
did_you_mean_flag :: proc(name: string) {
    lower := strings.to_lower(name, context.temp_allocator)

    if strings.has_prefix(lower, "opt") {
        fmt.eprintf("Did you mean '-%s' for the command?\n", lower)
        fmt.eprintf("The flag '-%s' does not exist. To set optimization mode, use '-o:speed', '-o:size', '-o:minimal', or '-o:none'.\n", name)
    } else {
        fmt.eprintf("Unknown flag: '-%s'\n", name)
    }
}
