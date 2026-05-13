// Forward declarations - defined in other parts
BuildFlagKind :: enum int {
	Invalid,
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

BuildFlagParamKind :: enum int {
	None,
	Boolean,
	Integer,
	Float,
	String,
	COUNT,
}

BuildFlag :: struct {
	kind:           BuildFlagKind,
	name:           string,
	param_kind:     BuildFlagParamKind,
	command_support: u64,
	allow_multiple: bool,
}

ExactValue :: distinct rawptr
ExactValueKind :: enum int { Invalid, Bool, Integer, Float, String }

// External procs from other parts
add_flag :: proc(build_flags: ^[dynamic]BuildFlag, kind: BuildFlagKind, name: string, param_kind: BuildFlagParamKind, command_support: u64, allow_multiple: bool = false)
build_param_to_exact_value :: proc(name: string, param: string) -> ExactValue
did_you_mean_flag :: proc(flag: string)
exact_value_bool :: proc(b: bool) -> ExactValue
exact_value_integer_from_string :: proc(s: string) -> ExactValue
exact_value_float_from_string :: proc(s: string) -> ExactValue
exact_value_string :: proc(s: string) -> ExactValue
heap_allocator :: proc() -> runtime.Allocator
permanent_allocator :: proc() -> runtime.Allocator
find_library_collection_path :: proc(name: string) -> (string, bool)
add_library_collection :: proc(name: string, path: string)
path_to_fullpath :: proc(allocator: runtime.Allocator, path: string) -> (string, bool)
path_is_directory :: proc(path: string) -> bool
is_build_flag_path_valid :: proc(path: string) -> bool
path_to_full_path :: proc(allocator: runtime.Allocator, path: string) -> string
levenstein_distance_case_insensitive :: proc(a: string, b: string) -> int

// External data
build_context: BuildContext_stub
selected_target_metrics: _
selected_subtarget: _

// Stub types
BuildContext_stub :: struct {}
Subtarget :: distinct int
Subtarget_COUNT :: 16

// Command bitmask constants
Command_build          :: 1 << 0
Command_run            :: 1 << 1
Command_check          :: 1 << 2
Command_doc            :: 1 << 3
Command_test           :: 1 << 4
Command_strip_semicolon :: 1 << 5
Command_bundle_android :: 1 << 6
Command_version        :: 1 << 7
Command_bug_report     :: 1 << 8
Command_bundle         :: Command_bundle_android
Command_all            :: 0xFFFFFFFFFFFFFFFF
Command__does_build    :: Command_build | Command_run | Command_test
Command__does_check    :: Command_build | Command_run | Command_check | Command_strip_semicolon | Command_doc | Command_test

// External arrays referenced
linker_choices: []string
windows_subsystem_names: []string
named_targets: []string
subtarget_strings: []string

@(require)
import "core:fmt"
@(require)
import "core:strings"
@(require)
import "core:os"
@(require)
import "core:slice"
@(require)
import "core:runtime"

parse_build_flags :: proc(args: []string) {
	build_flags := make([dynamic]BuildFlag, heap_allocator())

	// --- add_flag calls (lines 448-561) ---

	add_flag(&build_flags, .Help,                     "help",                      .None,    Command_all)
	add_flag(&build_flags, .Help,                     "h",                         .None,    Command_all)
	add_flag(&build_flags, .SingleFile,               "file",                      .None,    Command_all)
	add_flag(&build_flags, .SingleFile,               "f",                         .None,    Command_all)
	add_flag(&build_flags, .OutFile,                  "out",                       .String,  Command__does_build)
	add_flag(&build_flags, .OutFile,                  "o",                         .String,  Command__does_build)
	add_flag(&build_flags, .OptimizationMode,         "optimization-mode",         .String,  Command__does_build)
	add_flag(&build_flags, .OptimizationMode,         "o",                         .String,  Command__does_build, true)
	add_flag(&build_flags, .ShowTimings,              "show-timings",              .None,    Command__does_build)
	add_flag(&build_flags, .ShowUnused,               "show-unused",               .None,    Command__does_check)
	add_flag(&build_flags, .ShowUnusedWithLocation,   "show-unused-with-location", .None,    Command__does_check)
	add_flag(&build_flags, .ShowMoreTimings,          "show-more-timings",         .None,    Command__does_build)
	add_flag(&build_flags, .ShowImportGraph,          "show-import-graph",         .None,    Command__does_check)
	add_flag(&build_flags, .ExportTimings,            "export-timings",            .None,    Command__does_build)
	add_flag(&build_flags, .ExportTimingsFile,        "export-timings-file",       .String,  Command__does_build)
	add_flag(&build_flags, .ExportDependencies,       "export-dependencies",       .None,    Command__does_build)
	add_flag(&build_flags, .ExportDependenciesFile,   "export-dependencies-file",  .String,  Command__does_build)
	add_flag(&build_flags, .ShowSystemCalls,          "show-system-calls",         .None,    Command__does_build)
	add_flag(&build_flags, .ThreadCount,              "thread-count",              .Integer, Command__does_build)
	add_flag(&build_flags, .KeepTempFiles,            "keep-temp-files",           .None,    Command__does_build)
	add_flag(&build_flags, .Collection,               "collection",                .String,  Command__does_check, true)
	add_flag(&build_flags, .Collection,               "c",                         .String,  Command__does_check, true)
	add_flag(&build_flags, .Define,                   "define",                    .String,  Command__does_check, true)
	add_flag(&build_flags, .Define,                   "d",                         .String,  Command__does_check, true)
	add_flag(&build_flags, .BuildMode,                "build-mode",                .String,  Command__does_build)
	add_flag(&build_flags, .KeepExecutable,           "keep-executable",           .None,    Command__does_build)
	add_flag(&build_flags, .Target,                   "target",                    .String,  Command__does_check)
	add_flag(&build_flags, .Subtarget,                "subtarget",                 .String,  Command__does_build)
	add_flag(&build_flags, .Debug,                    "debug",                     .None,    Command__does_build)
	add_flag(&build_flags, .DisableAssert,            "disable-assert",            .None,    Command__does_build)
	add_flag(&build_flags, .NoBoundsCheck,            "no-bounds-check",           .None,    Command__does_build)
	add_flag(&build_flags, .NoTypeAssert,             "no-type-assert",            .None,    Command__does_build)
	add_flag(&build_flags, .NoDynamicLiterals,        "no-dynamic-literals",       .None,    Command__does_build)
	add_flag(&build_flags, .DynamicLiterals,          "dynamic-literals",          .None,    Command__does_build)
	add_flag(&build_flags, .NoCRT,                    "no-crt",                    .None,    Command__does_build)
	add_flag(&build_flags, .NoRPath,                  "no-rpath",                  .None,    Command__does_build)
	add_flag(&build_flags, .NoEntryPoint,             "no-entry-point",            .None,    Command__does_check)
	add_flag(&build_flags, .Linker,                   "linker",                    .String,  Command__does_build)
	add_flag(&build_flags, .UseSeparateModules,       "use-separate-modules",      .None,    Command__does_build)
	add_flag(&build_flags, .UseSingleModule,          "use-single-module",         .None,    Command__does_build)
	add_flag(&build_flags, .NoThreadedChecker,        "no-threaded-checker",       .None,    Command__does_check)
	add_flag(&build_flags, .ShowDebugMessages,        "show-debug-messages",       .None,    Command__does_check)
	add_flag(&build_flags, .DidYouMeanLimit,          "did-you-mean-limit",        .Integer, Command__does_check)
	add_flag(&build_flags, .ShowDefineables,          "show-defineables",          .None,    Command__does_check)
	add_flag(&build_flags, .ExportDefineables,        "export-defineables",        .None,    Command__does_check)
	add_flag(&build_flags, .IgnoreUnusedDefineables,  "ignore-unused-defineables", .None,    Command__does_check)
	add_flag(&build_flags, .Vet,                      "vet",                       .None,    Command__does_check)
	add_flag(&build_flags, .VetShadowing,             "vet-shadowing",             .None,    Command__does_check)
	add_flag(&build_flags, .VetUnused,                "vet-unused",                .None,    Command__does_check)
	add_flag(&build_flags, .VetUnusedImports,         "vet-unused-imports",        .None,    Command__does_check)
	add_flag(&build_flags, .VetUnusedVariables,       "vet-unused-variables",      .None,    Command__does_check)
	add_flag(&build_flags, .VetUnusedProcedures,      "vet-unused-procedures",     .None,    Command__does_check)
	add_flag(&build_flags, .VetUsingStmt,             "vet-using-stmt",            .None,    Command__does_check)
	add_flag(&build_flags, .VetUsingParam,            "vet-using-param",           .None,    Command__does_check)
	add_flag(&build_flags, .VetStyle,                 "vet-style",                 .None,    Command__does_check)
	add_flag(&build_flags, .VetSemicolon,             "vet-semicolon",             .None,    Command__does_check)
	add_flag(&build_flags, .VetCast,                  "vet-cast",                  .None,    Command__does_check)
	add_flag(&build_flags, .VetTabs,                  "vet-tabs",                  .None,    Command__does_check)
	add_flag(&build_flags, .VetPackages,              "vet-packages",              .None,    Command__does_check)
	add_flag(&build_flags, .CustomAttribute,          "custom-attribute",          .String,  Command__does_check, true)
	add_flag(&build_flags, .IgnoreUnknownAttributes,  "ignore-unknown-attributes", .None,    Command__does_check)
	add_flag(&build_flags, .ExtraLinkerFlags,         "extra-linker-flags",        .String,  Command__does_build, true)
	add_flag(&build_flags, .ExtraAssemblerFlags,      "extra-assembler-flags",     .String,  Command__does_build, true)
	add_flag(&build_flags, .Microarch,                "microarch",                 .String,  Command__does_build)
	add_flag(&build_flags, .TargetFeatures,           "target-features",           .String,  Command__does_build)
	add_flag(&build_flags, .StrictTargetFeatures,     "strict-target-features",    .None,    Command__does_build)
	add_flag(&build_flags, .MinimumOSVersion,         "minimum-os-version",        .String,  Command__does_build)
	add_flag(&build_flags, .NoThreadLocal,            "no-thread-local",           .None,    Command__does_build)
	add_flag(&build_flags, .RelocMode,                "reloc-mode",                .String,  Command__does_build)
	add_flag(&build_flags, .DisableRedZone,           "disable-red-zone",          .None,    Command__does_build)
	add_flag(&build_flags, .DisableUnwind,            "disable-unwind",            .None,    Command__does_build)
	add_flag(&build_flags, .DisallowDo,               "disallow-do",               .None,    Command__does_check)
	add_flag(&build_flags, .DefaultToNilAllocator,    "default-to-nil-allocator",  .None,    Command__does_build)
	add_flag(&build_flags, .DefaultToPanicAllocator,  "default-to-panic-allocator",.None,    Command__does_build)
	add_flag(&build_flags, .StrictStyle,              "strict-style",              .None,    Command__does_check)
	add_flag(&build_flags, .ForeignErrorProcedures,   "foreign-error-procedures",  .None,    Command__does_check)
	add_flag(&build_flags, .NoRTTI,                   "no-rtti",                   .None,    Command__does_build)
	add_flag(&build_flags, .DynamicMapCalls,          "dynamic-map-calls",         .None,    Command__does_build)
	add_flag(&build_flags, .ObfuscateSourceCodeLocations, "obfuscate-source-code-locations", .None, Command__does_build)
	add_flag(&build_flags, .SourceCodeLocations,      "source-code-locations",     .String,  Command__does_build)
	add_flag(&build_flags, .Compact,                  "compact",                   .None,    Command__does_build)
	add_flag(&build_flags, .GlobalDefinitions,        "global-definitions",        .None,    Command__does_check)
	add_flag(&build_flags, .GoToDefinitions,          "go-to-definitions",         .None,    Command__does_check)
	add_flag(&build_flags, .Short,                    "short",                     .None,    Command__does_check)
	add_flag(&build_flags, .InSourceOrder,            "in-source-order",           .None,    Command__does_check)
	add_flag(&build_flags, .AllPackages,              "all-packages",              .None,    Command_test)
	add_flag(&build_flags, .DocFormat,                "doc-format",                .String,  Command_doc)
	add_flag(&build_flags, .IgnoreWarnings,           "ignore-warnings",           .None,    Command__does_check)
	add_flag(&build_flags, .WarningsAsErrors,         "warnings-as-errors",        .None,    Command__does_check)
	add_flag(&build_flags, .TerseErrors,              "terse-errors",              .None,    Command__does_check)
	add_flag(&build_flags, .VerboseErrors,            "verbose-errors",            .None,    Command__does_check)
	add_flag(&build_flags, .JsonErrors,               "json-errors",               .None,    Command__does_check)
	add_flag(&build_flags, .ErrorPosStyle,            "error-pos-style",           .String,  Command__does_check)
	add_flag(&build_flags, .MaxErrorCount,            "max-error-count",           .Integer, Command__does_check)
	add_flag(&build_flags, .MinLinkLibs,              "min-link-libs",             .None,    Command__does_build)
	add_flag(&build_flags, .PrintLinkerFlags,         "print-linker-flags",        .None,    Command__does_build)
	add_flag(&build_flags, .ExportLinkedLibraries,    "export-linked-libraries",   .None,    Command__does_build)
	add_flag(&build_flags, .IntegerDivisionByZero,    "integer-division-by-zero",  .None,    Command__does_check)
	add_flag(&build_flags, .BuildDiagnostics,         "build-diagnostics",         .None,    Command__does_check)
	add_flag(&build_flags, .InternalFastISel,         "internal-fast-isel",        .None,    Command__does_build)
	add_flag(&build_flags, .InternalIgnoreLazy,       "internal-ignore-lazy",      .None,    Command__does_build)
	add_flag(&build_flags, .InternalIgnoreLLVMBuild,  "internal-ignore-llvm-build",.None,    Command__does_build)
	add_flag(&build_flags, .InternalIgnorePanic,      "internal-ignore-panic",     .None,    Command__does_build)
	add_flag(&build_flags, .InternalModulePerFile,    "internal-module-per-file",  .None,    Command__does_build)
	add_flag(&build_flags, .InternalCached,           "internal-cached",           .None,    Command__does_build)
	add_flag(&build_flags, .InternalNoInline,         "internal-no-inline",        .None,    Command__does_build)
	add_flag(&build_flags, .InternalByValue,          "internal-by-value",         .None,    Command__does_build)
	add_flag(&build_flags, .InternalWeakMonomorphization, "internal-weak-monomorphization", .None, Command__does_build)
	add_flag(&build_flags, .InternalLLVMVerification, "internal-llvm-verification",.None,    Command__does_build)
	add_flag(&build_flags, .InternalLLVMNoSROA,       "internal-llvm-no-sroa",     .None,    Command__does_build)
	add_flag(&build_flags, .InternalEnableRVO,        "internal-enable-rvo",       .None,    Command__does_build)
	add_flag(&build_flags, .Sanitize,                 "sanitize",                  .String,  Command__does_build)
	add_flag(&build_flags, .LTO,                      "lto",                       .String,  Command__does_build)
	add_flag(&build_flags, .IgnoreVsSearch,           "ignore-vs-search",          .None,    Command__does_build)
	add_flag(&build_flags, .ResourceFile,             "resource",                  .String,  Command__does_build)
	add_flag(&build_flags, .WindowsPdbName,           "pdb-name",                  .String,  Command__does_build)
	add_flag(&build_flags, .Subsystem,                "subsystem",                 .String,  Command__does_build)
	add_flag(&build_flags, .AndroidKeystore,          "android-keystore",          .String,  Command__does_build)
	add_flag(&build_flags, .AndroidKeystoreAlias,     "android-keystore-alias",    .String,  Command__does_build)
	add_flag(&build_flags, .AndroidKeystorePassword,  "android-keystore-password", .String,  Command__does_build)

	// --- Determine flag_args (lines 562-569) ---
	flag_args := args[3:]
	// bundle_android handling would be: args[4:] but we skip that for now

	set_flags := make(map[string]BuildFlag, heap_allocator())

	// --- Parse each flag argument (lines 570-1632) ---
	for flag_idx := 0; flag_idx < len(flag_args); flag_idx += 1 {
		flag := flag_args[flag_idx]

		if !strings.has_prefix(flag, "-") {
			continue
		}

		// Handle '--' prefix
		if strings.has_prefix(flag, "--") {
			flag = flag[2:]
		} else {
			flag = flag[1:]
		}

		// Extract name and param
		name: string
		param: string
		has_param: bool

		if idx := strings.index_byte(flag, ':'); idx >= 0 {
			name = flag[:idx]
			param = flag[idx+1:]
			has_param = true
		} else if idx := strings.index_byte(flag, '='); idx >= 0 {
			name = flag[:idx]
			param = flag[idx+1:]
			has_param = true
		} else {
			name = flag
			param = ""
			has_param = false
		}

		// Find matching BuildFlag
		bf: Maybe(BuildFlag)
		for i in 0 ..< len(build_flags) {
			if build_flags[i].name == name {
				bf = build_flags[i]
				break
			}
		}

		if bf == nil {
			did_you_mean_flag(name)
			continue
		}

		b := bf.(BuildFlag)

		// Check command_support
		if b.command_support & Command__does_check == 0 {
			fmt.eprintf("Flag '-%s' is not supported for this command\n", b.name)
			continue
		}

		// Check for duplicate
		if name in set_flags && !b.allow_multiple {
			fmt.eprintf("Flag '-%s' specified multiple times\n", b.name)
			continue
		}
		set_flags[name] = b

		// Validate and build ExactValue
		value: ExactValue

		switch b.param_kind {
		case .None:
			if has_param {
				fmt.eprintf("Flag '-%s' does not take a parameter\n", b.name)
				continue
			}
			value = exact_value_bool(true)

		case .Boolean:
			if !has_param {
				value = exact_value_bool(true)
			} else {
				if param == "true" || param == "1" || param == "yes" {
					value = exact_value_bool(true)
				} else if param == "false" || param == "0" || param == "no" {
					value = exact_value_bool(false)
				} else {
					fmt.eprintf("Flag '-%s' expects a boolean parameter (true/false)\n", b.name)
					continue
				}
			}

		case .Integer:
			if !has_param {
				fmt.eprintf("Flag '-%s' requires an integer parameter\n", b.name)
				continue
			}
			value = exact_value_integer_from_string(param)

		case .Float:
			if !has_param {
				fmt.eprintf("Flag '-%s' requires a float parameter\n", b.name)
				continue
			}
			value = exact_value_float_from_string(param)

		case .String:
			if !has_param {
				fmt.eprintf("Flag '-%s' requires a string parameter\n", b.name)
				continue
			}
			value = exact_value_string(param)

		case .COUNT:
			// unreachable
		}

		// --- Switch on b.kind to set build_context fields ---
		switch b.kind {
		case .Help:
			build_context.show_help = true

		case .SingleFile:
			build_context.use_single_file = true

		case .OutFile:
			build_context.out_file = param

		case .OptimizationMode:
			build_context.optimization_mode = param

		case .ShowTimings:
			build_context.show_timings = true

		case .ShowUnused:
			build_context.show_unused = true

		case .ShowUnusedWithLocation:
			build_context.show_unused_with_location = true

		case .ShowMoreTimings:
			build_context.show_more_timings = true

		case .ShowImportGraph:
			build_context.show_import_graph = true

		case .ExportTimings:
			build_context.export_timings = true

		case .ExportTimingsFile:
			build_context.export_timings_file = param

		case .ExportDependencies:
			build_context.export_dependencies = true

		case .ExportDependenciesFile:
			build_context.export_dependencies_file = param

		case .ShowSystemCalls:
			build_context.show_system_calls = true

		case .ThreadCount:
			// thread count is an integer, set from param
			build_context.thread_count = 0 // TODO: parse from value

		case .KeepTempFiles:
			build_context.keep_temp_files = true

		case .Collection:
			if !is_build_flag_path_valid(param) {
				fmt.eprintf("Invalid collection path: '%s'\n", param)
				continue
			}
			fullpath, ok := path_to_fullpath(heap_allocator(), param)
			if !ok || !path_is_directory(fullpath) {
				fmt.eprintf("Collection path is not a valid directory: '%s'\n", param)
				continue
			}
			add_library_collection(name, fullpath)

		case .Define:
			build_context.defines = param

		case .BuildMode:
			build_context.build_mode = param

		case .KeepExecutable:
			build_context.keep_executable = true

		case .Target:
			build_context.target = param

		case .Subtarget:
			// Look up subtarget from subtarget_strings
			found_subtarget := false
			param_lower := strings.to_lower(param)
			for i in 0 ..< len(subtarget_strings) {
				if strings.equal_fold(subtarget_strings[i], param) {
					selected_subtarget = Subtarget(i)
					found_subtarget = true
					break
				}
			}
			if !found_subtarget {
				fmt.eprintf("Unknown subtarget: '%s'\n", param)
			}

		case .Debug:
			build_context.debug = true

		case .DisableAssert:
			build_context.disable_assert = true

		case .NoBoundsCheck:
			build_context.no_bounds_check = true

		case .NoTypeAssert:
			build_context.no_type_assert = true

		case .NoDynamicLiterals:
			build_context.no_dynamic_literals = true

		case .DynamicLiterals:
			build_context.dynamic_literals = true

		case .NoCRT:
			build_context.no_crt = true

		case .NoRPath:
			build_context.no_rpath = true

		case .NoEntryPoint:
			build_context.no_entry_point = true

		case .Linker:
			param_lower := strings.to_lower(param)
			found_linker := false
			for l in linker_choices {
				if strings.equal_fold(l, param) {
					build_context.linker = param_lower
					found_linker = true
					break
				}
			}
			if !found_linker {
				// Levenshtein distance did-you-mean for linker
				best_dist := max(int)
				best_name: string
				for l in linker_choices {
					dist := levenstein_distance_case_insensitive(param, l)
					if dist < best_dist {
						best_dist = dist
						best_name = l
					}
				}
				if best_name != "" {
					fmt.eprintf("Unknown linker: '%s'. Did you mean '%s'?\n", param, best_name)
				} else {
					fmt.eprintf("Unknown linker: '%s'\n", param)
				}
			}

		case .UseSeparateModules:
			build_context.use_separate_modules = true

		case .UseSingleModule:
			build_context.use_single_module = true

		case .NoThreadedChecker:
			build_context.no_threaded_checker = true

		case .ShowDebugMessages:
			build_context.show_debug_messages = true

		case .DidYouMeanLimit:
			build_context.did_you_mean_limit = 0 // TODO: parse

		case .ShowDefineables:
			build_context.show_defineables = true

		case .ExportDefineables:
			build_context.export_defineables = true

		case .IgnoreUnusedDefineables:
			build_context.ignore_unused_defineables = true

		case .Vet:
			build_context.vet = true

		case .VetShadowing:
			build_context.vet_shadowing = true

		case .VetUnused:
			build_context.vet_unused = true

		case .VetUnusedImports:
			build_context.vet_unused_imports = true

		case .VetUnusedVariables:
			build_context.vet_unused_variables = true

		case .VetUnusedProcedures:
			build_context.vet_unused_procedures = true

		case .VetUsingStmt:
			build_context.vet_using_stmt = true

		case .VetUsingParam:
			build_context.vet_using_param = true

		case .VetStyle:
			build_context.vet_style = true

		case .VetSemicolon:
			build_context.vet_semicolon = true

		case .VetCast:
			build_context.vet_cast = true

		case .VetTabs:
			build_context.vet_tabs = true

		case .VetPackages:
			build_context.vet_packages = true

		case .CustomAttribute:
			// TODO: append custom attribute

		case .IgnoreUnknownAttributes:
			build_context.ignore_unknown_attributes = true

		case .ExtraLinkerFlags:
			append(&build_context.extra_linker_flags, param)

		case .ExtraAssemblerFlags:
			append(&build_context.extra_assembler_flags, param)

		case .Microarch:
			build_context.microarch = param

		case .TargetFeatures:
			build_context.target_features = param

		case .StrictTargetFeatures:
			build_context.strict_target_features = true

		case .MinimumOSVersion:
			build_context.minimum_os_version = param

		case .NoThreadLocal:
			build_context.no_thread_local = true

		case .RelocMode:
			build_context.reloc_mode = param

		case .DisableRedZone:
			build_context.disable_red_zone = true

		case .DisableUnwind:
			build_context.disable_unwind = true

		case .DisallowDo:
			build_context.disallow_do = true

		case .DefaultToNilAllocator:
			build_context.default_to_nil_allocator = true

		case .DefaultToPanicAllocator:
			build_context.default_to_panic_allocator = true

		case .StrictStyle:
			build_context.strict_style = true

		case .ForeignErrorProcedures:
			build_context.foreign_error_procedures = true

		case .NoRTTI:
			build_context.no_rtti = true

		case .DynamicMapCalls:
			build_context.dynamic_map_calls = true

		case .ObfuscateSourceCodeLocations:
			build_context.obfuscate_source_code_locations = true

		case .SourceCodeLocations:
			build_context.source_code_locations = param

		case .Compact:
			build_context.compact = true

		case .GlobalDefinitions:
			build_context.global_definitions = true

		case .GoToDefinitions:
			build_context.go_to_definitions = true

		case .Short:
			build_context.short_output = true

		case .InSourceOrder:
			build_context.in_source_order = true

		case .AllPackages:
			build_context.all_packages = true

		case .DocFormat:
			build_context.doc_format = param

		case .IgnoreWarnings:
			build_context.ignore_warnings = true

		case .WarningsAsErrors:
			build_context.warnings_as_errors = true

		case .TerseErrors:
			build_context.terse_errors = true

		case .VerboseErrors:
			build_context.verbose_errors = true

		case .JsonErrors:
			build_context.json_errors = true

		case .ErrorPosStyle:
			build_context.error_pos_style = param

		case .MaxErrorCount:
			build_context.max_error_count = 0 // TODO: parse

		case .MinLinkLibs:
			build_context.min_link_libs = true

		case .PrintLinkerFlags:
			build_context.print_linker_flags = true

		case .ExportLinkedLibraries:
			build_context.export_linked_libraries = true

		case .IntegerDivisionByZero:
			build_context.integer_division_by_zero = true

		case .BuildDiagnostics:
			build_context.build_diagnostics = true

		case .InternalFastISel:
			build_context.internal_fast_isel = true

		case .InternalIgnoreLazy:
			build_context.internal_ignore_lazy = true

		case .InternalIgnoreLLVMBuild:
			build_context.internal_ignore_llvm_build = true

		case .InternalIgnorePanic:
			build_context.internal_ignore_panic = true

		case .InternalModulePerFile:
			build_context.internal_module_per_file = true

		case .InternalCached:
			build_context.internal_cached = true

		case .InternalNoInline:
			build_context.internal_no_inline = true

		case .InternalByValue:
			build_context.internal_by_value = true

		case .InternalWeakMonomorphization:
			build_context.internal_weak_monomorphization = true

		case .InternalLLVMVerification:
			build_context.internal_llvm_verification = true

		case .InternalLLVMNoSROA:
			build_context.internal_llvm_no_sroa = true

		case .InternalEnableRVO:
			build_context.internal_enable_rvo = true

		case .Sanitize:
			build_context.sanitize = param

		case .LTO:
			build_context.lto = param

		case .IgnoreVsSearch:
			build_context.ignore_vs_search = true

		case .ResourceFile:
			build_context.resource_file = param

		case .WindowsPdbName:
			build_context.windows_pdb_name = param

		case .Subsystem:
			param_lower := strings.to_lower(param)
			found_subsystem := false
			for s in windows_subsystem_names {
				if strings.equal_fold(s, param) {
					build_context.subsystem = param_lower
					found_subsystem = true
					break
				}
			}
			if !found_subsystem {
				best_dist := max(int)
				best_name: string
				for s in windows_subsystem_names {
					dist := levenstein_distance_case_insensitive(param, s)
					if dist < best_dist {
						best_dist = dist
						best_name = s
					}
				}
				if best_name != "" {
					fmt.eprintf("Unknown subsystem: '%s'. Did you mean '%s'?\n", param, best_name)
				} else {
					fmt.eprintf("Unknown subsystem: '%s'\n", param)
				}
			}

		case .AndroidKeystore:
			build_context.android_keystore = param

		case .AndroidKeystoreAlias:
			build_context.android_keystore_alias = param

		case .AndroidKeystorePassword:
			build_context.android_keystore_password = param

		case .Invalid, .COUNT:
			// Unreachable
		}
	}

	// --- Post-validation constraints (lines 1633-1660) ---

	// vet_unused_procedures requires vet_packages
	if build_context.vet_unused_procedures && !build_context.vet_packages {
		fmt.eprintln("Flag '-vet-unused-procedures' requires '-vet-packages'")
	}

	// export_timings requires export_timings_file and vice versa
	if build_context.export_timings && build_context.export_timings_file == "" {
		fmt.eprintln("Flag '-export-timings' requires '-export-timings-file:<path>'")
	}
	if build_context.export_timings_file != "" && !build_context.export_timings {
		fmt.eprintln("Flag '-export-timings-file' requires '-export-timings'")
	}

	// export_timings requires show_timings or show_more_timings
	if build_context.export_timings && !build_context.show_timings && !build_context.show_more_timings {
		fmt.eprintln("Flag '-export-timings' requires '-show-timings' or '-show-more-timings'")
	}

	// export_dependencies incompatibility checks
	if build_context.export_dependencies && build_context.export_dependencies_file != "" {
		fmt.eprintln("Cannot use both '-export-dependencies' and '-export-dependencies-file'")
	}

	// test_all_packages compatibility
	if build_context.all_packages {
		// valid only with test command
	}
}
