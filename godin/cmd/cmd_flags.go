package cmd

import (
	"strings"
	"unsafe"
)

const DEFAULT_DID_YOU_MEAN_LIMIT isize = 10
const lb_use_new_pass_system = true

type CmdDocFlag uint32

const (
	CmdDocFlag_Short         CmdDocFlag = 1 << 0
	CmdDocFlag_InSourceOrder CmdDocFlag = 1 << 1
	CmdDocFlag_AllPackages   CmdDocFlag = 1 << 2
	CmdDocFlag_DocFormat     CmdDocFlag = 1 << 3
)

type String_Iterator struct {
	Str String
	Pos isize
}

func add_flag(build_flags *[]BuildFlag, kind BuildFlagKind, name string, param_kind BuildFlagParamKind, command_support uint64, allow_multiple ...bool) {
	am := false
	if len(allow_multiple) > 0 {
		am = allow_multiple[0]
	}
	*build_flags = append(*build_flags, BuildFlag{
		Kind:           kind,
		Name:           name,
		ParamKind:      param_kind,
		CommandSupport: command_support,
		AllowMultiple:  am,
	})
}

func build_param_to_exact_value(name String, param String) ExactValue {
	var value ExactValue
	if param.Len == 0 {
		gb_printf_err("Invalid flag parameter for '%s' = '%s'\n", goStr(name), goStr(param))
		return value
	}
	if str_eq_ignore_case(param, S("t")) || str_eq_ignore_case(param, S("true")) {
		return exact_value_bool(true)
	}
	if str_eq_ignore_case(param, S("f")) || str_eq_ignore_case(param, S("false")) {
		return exact_value_bool(false)
	}
	if param.Data[0] == '-' || param.Data[0] == '+' || (param.Data[0] >= '0' && param.Data[0] <= '9') {
		if string_contains_char(param, '.') {
			value = exact_value_float_from_string(param)
		} else {
			value = exact_value_integer_from_string(param)
		}
		if value.Kind != ExactValue_Invalid {
			return value
		}
	}
	value = exact_value_string(param)
	if param.Data[0] == '\'' && value.Kind == ExactValue_String {
		s := value.Value_string
		if s.Len > 1 && s.Data[0] == '\'' && s.Data[s.Len-1] == '\'' {
			value.Value_string = substring(s, 1, s.Len-1)
		}
	}
	if value.Kind != ExactValue_String {
		gb_printf_err("Invalid flag parameter for '%s' = '%s'\n", goStr(name), goStr(param))
	}
	return value
}

func S(s string) String {
	return String{Data: unsafe.StringData(s), Len: isize(len(s))}
}

func goStr(s String) string {
	if s.Data == nil {
		return ""
	}
	return unsafe.String(s.Data, s.Len)
}

func did_you_mean_flag(flag String) {
	name := goStr(flag)
	nameLower := strings.ToLower(name)
	if nameLower == "opt" {
		gb_printf_err("`-opt` is an unrecognized option. Did you mean `-o`?\n")
		return
	}
	gb_printf_err("Unknown flag: '%s'\n", goStr(flag))
}

func parse_build_flags(args []String) bool {
	build_flags := make([]BuildFlag, 0, BuildFlagCOUNT)

	add_flag(&build_flags, BuildFlagHelp, "help", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagSingleFile, "file", BuildFlagParamNone, Command__does_build|Command__does_check)
	add_flag(&build_flags, BuildFlagOutFile, "out", BuildFlagParamString, Command__does_build|Command_test|Command_doc)
	add_flag(&build_flags, BuildFlagOptimizationMode, "o", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagShowTimings, "show-timings", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagShowMoreTimings, "show-more-timings", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagShowImportGraph, "show-import-graph", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagExportTimings, "export-timings", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagExportTimingsFile, "export-timings-file", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagExportDependencies, "export-dependencies", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagExportDependenciesFile, "export-dependencies-file", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagShowUnused, "show-unused", BuildFlagParamNone, Command_check)
	add_flag(&build_flags, BuildFlagShowUnusedWithLocation, "show-unused-with-location", BuildFlagParamNone, Command_check)
	add_flag(&build_flags, BuildFlagShowSystemCalls, "show-system-calls", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagThreadCount, "thread-count", BuildFlagParamInteger, CommandAll)
	add_flag(&build_flags, BuildFlagKeepTempFiles, "keep-temp-files", BuildFlagParamNone, Command__does_build|Command_strip_semicolon)
	add_flag(&build_flags, BuildFlagCollection, "collection", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagDefine, "define", BuildFlagParamString, Command__does_check, true)
	add_flag(&build_flags, BuildFlagBuildMode, "build-mode", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagKeepExecutable, "keep-executable", BuildFlagParamNone, Command__does_build|Command_test)
	add_flag(&build_flags, BuildFlagTarget, "target", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagSubtarget, "subtarget", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagDebug, "debug", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagDisableAssert, "disable-assert", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoBoundsCheck, "no-bounds-check", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoTypeAssert, "no-type-assert", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoThreadLocal, "no-thread-local", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoDynamicLiterals, "no-dynamic-literals", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagDynamicLiterals, "dynamic-literals", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoCRT, "no-crt", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagNoRPath, "no-rpath", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagNoEntryPoint, "no-entry-point", BuildFlagParamNone, Command__does_check & ^Command_test)
	add_flag(&build_flags, BuildFlagLinker, "linker", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagUseSeparateModules, "use-separate-modules", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagUseSingleModule, "use-single-module", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagNoThreadedChecker, "no-threaded-checker", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagShowDebugMessages, "show-debug-messages", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagDidYouMeanLimit, "did-you-mean-limit", BuildFlagParamInteger, Command__does_check)

	add_flag(&build_flags, BuildFlagShowDefineables, "show-defineables", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagExportDefineables, "export-defineables", BuildFlagParamString, Command__does_check)
	add_flag(&build_flags, BuildFlagIgnoreUnusedDefineables, "ignore-unused-defineables", BuildFlagParamNone, Command__does_check)

	add_flag(&build_flags, BuildFlagVet, "vet", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUnused, "vet-unused", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUnusedVariables, "vet-unused-variables", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUnusedProcedures, "vet-unused-procedures", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUnusedImports, "vet-unused-imports", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetShadowing, "vet-shadowing", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUsingStmt, "vet-using-stmt", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetUsingParam, "vet-using-param", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetStyle, "vet-style", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetSemicolon, "vet-semicolon", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetCast, "vet-cast", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetTabs, "vet-tabs", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagVetPackages, "vet-packages", BuildFlagParamString, Command__does_check)

	add_flag(&build_flags, BuildFlagCustomAttribute, "custom-attribute", BuildFlagParamString, Command__does_check, true)
	add_flag(&build_flags, BuildFlagIgnoreUnknownAttributes, "ignore-unknown-attributes", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagExtraLinkerFlags, "extra-linker-flags", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagExtraAssemblerFlags, "extra-assembler-flags", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagMicroarch, "microarch", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagTargetFeatures, "target-features", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagStrictTargetFeatures, "strict-target-features", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagMinimumOSVersion, "minimum-os-version", BuildFlagParamString, Command__does_build|Command_bundle_android)

	add_flag(&build_flags, BuildFlagRelocMode, "reloc-mode", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagDisableRedZone, "disable-red-zone", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagDisableUnwind, "disable-unwind", BuildFlagParamNone, Command__does_build)

	add_flag(&build_flags, BuildFlagDisallowDo, "disallow-do", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagDefaultToNilAllocator, "default-to-nil-allocator", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagDefaultToPanicAllocator, "default-to-panic-allocator", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagStrictStyle, "strict-style", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagForeignKeyErrorProcedures, "foreign-error-procedures", BuildFlagParamNone, Command__does_check)

	add_flag(&build_flags, BuildFlagNoRTTI, "no-rtti", BuildFlagParamNone, Command__does_check)
	add_flag(&build_flags, BuildFlagNoRTTI, "disallow-rtti", BuildFlagParamNone, Command__does_check)

	add_flag(&build_flags, BuildFlagDynamicMapCalls, "dynamic-map-calls", BuildFlagParamNone, Command__does_check)

	add_flag(&build_flags, BuildFlagObfuscateSourceCodeLocations, "obfuscate-source-code-locations", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagSourceCodeLocations, "source-code-locations", BuildFlagParamString, Command__does_build)

	add_flag(&build_flags, BuildFlagShort, "short", BuildFlagParamNone, Command_doc)
	add_flag(&build_flags, BuildFlagInSourceOrder, "in-source-order", BuildFlagParamNone, Command_doc)
	add_flag(&build_flags, BuildFlagAllPackages, "all-packages", BuildFlagParamNone, Command_doc|Command_test|Command_build)
	add_flag(&build_flags, BuildFlagDocFormat, "doc-format", BuildFlagParamNone, Command_doc)

	add_flag(&build_flags, BuildFlagIgnoreWarnings, "ignore-warnings", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagWarningsAsErrors, "warnings-as-errors", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagTerseErrors, "terse-errors", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagVerboseErrors, "verbose-errors", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagJsonErrors, "json-errors", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagErrorPosStyle, "error-pos-style", BuildFlagParamString, CommandAll)
	add_flag(&build_flags, BuildFlagMaxErrorCount, "max-error-count", BuildFlagParamInteger, CommandAll)

	add_flag(&build_flags, BuildFlagMinLinkLibs, "min-link-libs", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagExportLinkedLibraries, "export-linked-libs-file", BuildFlagParamString, Command__does_check)

	add_flag(&build_flags, BuildFlagPrintLinkerFlags, "print-linker-flags", BuildFlagParamNone, Command_build)

	add_flag(&build_flags, BuildFlagIntegerDivisionByZero, "integer-division-by-zero", BuildFlagParamString, Command__does_check)

	add_flag(&build_flags, BuildFlagBuildDiagnostics, "build-diagnostics", BuildFlagParamNone, Command__does_build)

	add_flag(&build_flags, BuildFlagInternalFastISel, "internal-fast-isel", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalIgnoreLazy, "internal-ignore-lazy", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalIgnoreLLVMBuild, "internal-ignore-llvm-build", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalIgnorePanic, "internal-ignore-panic", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalModulePerFile, "internal-module-per-file", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalCached, "internal-cached", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalNoInline, "internal-no-inline", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalByValue, "internal-by-value", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalWeakMonomorphization, "internal-weak-monomorphization", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalLLVMVerification, "internal-ignore-llvm-verification", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalLLVMNoSROA, "internal-llvm-no-sroa", BuildFlagParamNone, CommandAll)
	add_flag(&build_flags, BuildFlagInternalEnableRVO, "internal-enable-rvo", BuildFlagParamNone, CommandAll)

	add_flag(&build_flags, BuildFlagSanitize, "sanitize", BuildFlagParamString, Command__does_build, true)
	add_flag(&build_flags, BuildFlagLTO, "lto", BuildFlagParamString, Command__does_build)

	add_flag(&build_flags, BuildFlagIgnoreVsSearch, "ignore-vs-search", BuildFlagParamNone, Command__does_build)
	add_flag(&build_flags, BuildFlagResourceFile, "resource", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagWindowsPdbName, "pdb-name", BuildFlagParamString, Command__does_build)
	add_flag(&build_flags, BuildFlagSubsystem, "subsystem", BuildFlagParamString, Command__does_build)

	add_flag(&build_flags, BuildFlagAndroidKeystore, "android-keystore", BuildFlagParamString, Command_bundle_android)
	add_flag(&build_flags, BuildFlagAndroidKeystoreAlias, "android-keystore-alias", BuildFlagParamString, Command_bundle_android)
	add_flag(&build_flags, BuildFlagAndroidKeystorePassword, "android-keystore-password", BuildFlagParamString, Command_bundle_android)

	var flag_args []String
	if build_context.command_kind == Command_bundle_android {
		flag_args = args[4:]
	} else {
		flag_args = args[3:]
	}

	set_flags := make([]bool, BuildFlagCOUNT)
	bad_flags := false

	for _, flag := range flag_args {
		if flag.Len == 0 || flag.Data[0] != '-' {
			gb_printf_err("Invalid flag: %s\n", goStr(flag))
			continue
		}
		if flag.Len >= 2 && flag.Data[0] == '-' && flag.Data[1] == '-' {
			flag = substring(flag, 1, flag.Len)
		}
		name := substring(flag, 1, flag.Len)
		end := isize(0)
		for ; end < name.Len; end++ {
			if name.Data[end] == ':' {
				break
			}
			if name.Data[end] == '=' {
				break
			}
		}
		name = substring(name, 0, end)
		var param String
		if end < flag.Len-1 {
			param = substring(flag, 2+end, flag.Len)
		}

		is_supported := true
		found := false
		var found_bf BuildFlag

		for _, bf := range build_flags {
			if goStr(name) != bf.Name {
				continue
			}
			found = true
			found_bf = bf
			if (bf.CommandSupport & build_context.command_kind) == 0 {
				is_supported = false
				break
			}

			if set_flags[bf.Kind] {
				gb_printf_err("Previous flag set: '%s'\n", goStr(name))
				bad_flags = true
			} else {
				var value ExactValue
				ok := false
				if bf.ParamKind == BuildFlagParamNone {
					if param.Len == 0 {
						ok = true
					} else {
						gb_printf_err("Flag '%s' was not expecting a parameter '%s'\n", goStr(name), goStr(param))
						bad_flags = true
					}
				} else if param.Len == 0 {
					gb_printf_err("Flag missing for '%s'\n", goStr(name))
					bad_flags = true
				} else {
					ok = true
					switch bf.ParamKind {
					default:
						ok = false
					case BuildFlagParamBoolean:
						if str_eq_ignore_case(param, S("t")) || str_eq_ignore_case(param, S("true")) || (param.Len == 1 && param.Data[0] == '1') {
							value = exact_value_bool(true)
						} else if str_eq_ignore_case(param, S("f")) || str_eq_ignore_case(param, S("false")) || (param.Len == 1 && param.Data[0] == '0') {
							value = exact_value_bool(false)
						} else {
							gb_printf_err("Invalid flag parameter for '%s' : '%s'\n", goStr(name), goStr(param))
						}
					case BuildFlagParamInteger:
						value = exact_value_integer_from_string(param)
					case BuildFlagParamFloat:
						value = exact_value_float_from_string(param)
					case BuildFlagParamString:
						value = exact_value_string(param)
						if value.Kind == ExactValue_String {
							s := value.Value_string
							if s.Len > 1 && s.Data[0] == '"' && s.Data[s.Len-1] == '"' {
								value.Value_string = substring(s, 1, s.Len-1)
							}
						}
					}
				}
				if ok {
					switch bf.ParamKind {
					case BuildFlagParamNone:
						if value.Kind != ExactValue_Invalid {
							gb_printf_err("%s expected no value, got %s\n", goStr(name), goStr(param))
							bad_flags = true
							ok = false
						}
					case BuildFlagParamBoolean:
						if value.Kind != ExactValue_Bool {
							gb_printf_err("%s expected a boolean, got %s\n", goStr(name), goStr(param))
							bad_flags = true
							ok = false
						}
					case BuildFlagParamInteger:
						if value.Kind != ExactValue_Integer {
							gb_printf_err("%s expected an integer, got %s\n", goStr(name), goStr(param))
							bad_flags = true
							ok = false
						}
					case BuildFlagParamFloat:
						if value.Kind != ExactValue_Float {
							gb_printf_err("%s expected a floating pointer number, got %s\n", goStr(name), goStr(param))
							bad_flags = true
							ok = false
						}
					case BuildFlagParamString:
						if value.Kind != ExactValue_String {
							gb_printf_err("%s expected a string, got %s\n", goStr(name), goStr(param))
							bad_flags = true
							ok = false
						}
					}
					if ok {
						switch bf.Kind {
						case BuildFlagHelp:
							build_context.show_help = true
						case BuildFlagOutFile:
							path := string_trim_whitespace(value.Value_string)
							if is_build_flag_path_valid(path) {
								build_context.out_filepath = path_to_full_path(heap_allocator(), path)
							} else {
								gb_printf_err("Invalid -out path, got %s\n", goStr(path))
								bad_flags = true
							}
						case BuildFlagOptimizationMode:
							valStr := goStr(value.Value_string)
							switch valStr {
							case "none":
								build_context.custom_optimization_level = true
								build_context.optimization_level = -1
							case "minimal":
								build_context.custom_optimization_level = true
								build_context.optimization_level = 0
							case "size":
								build_context.custom_optimization_level = true
								build_context.optimization_level = 1
							case "speed":
								build_context.custom_optimization_level = true
								build_context.optimization_level = 2
							case "aggressive":
								if lb_use_new_pass_system {
									build_context.custom_optimization_level = true
									build_context.optimization_level = 3
								} else {
									gb_printf_err("Invalid optimization mode for -o:<string>, got %s\n", goStr(value.Value_string))
									gb_printf_err("Valid optimization modes:\n")
									gb_printf_err("\tminimal\n")
									gb_printf_err("\tsize\n")
									gb_printf_err("\tspeed\n")
									gb_printf_err("\tnone (useful for -debug builds)\n")
									bad_flags = true
								}
							default:
								gb_printf_err("Invalid optimization mode for -o:<string>, got %s\n", goStr(value.Value_string))
								gb_printf_err("Valid optimization modes:\n")
								gb_printf_err("\tminimal\n")
								gb_printf_err("\tsize\n")
								gb_printf_err("\tspeed\n")
								if lb_use_new_pass_system {
									gb_printf_err("\taggressive\n")
								}
								gb_printf_err("\tnone (useful for -debug builds)\n")
								bad_flags = true
							}
						case BuildFlagShowTimings:
							build_context.show_timings = true
						case BuildFlagShowUnused:
							build_context.show_unused = true
						case BuildFlagShowUnusedWithLocation:
							build_context.show_unused = true
							build_context.show_unused_with_location = true
						case BuildFlagShowMoreTimings:
							build_context.show_timings = true
							build_context.show_more_timings = true
						case BuildFlagShowImportGraph:
							build_context.show_import_graph = true
						case BuildFlagExportTimings:
							valStr := goStr(value.Value_string)
							switch valStr {
							case "json":
								build_context.export_timings_format = TimingsExportJson
							case "csv":
								build_context.export_timings_format = TimingsExportCSV
							default:
								gb_printf_err("Invalid export format for -export-timings:<string>, got %s\n", goStr(value.Value_string))
								gb_printf_err("Valid export formats:\n")
								gb_printf_err("\tjson\n")
								gb_printf_err("\tcsv\n")
								bad_flags = true
							}
						case BuildFlagExportTimingsFile:
							export_path := string_trim_whitespace(value.Value_string)
							if is_build_flag_path_valid(export_path) {
								build_context.export_timings_file = path_to_full_path(heap_allocator(), export_path)
							} else {
								gb_printf_err("Invalid -export-timings-file path, got %s\n", goStr(export_path))
								bad_flags = true
							}
						case BuildFlagExportDependencies:
							valStr := goStr(value.Value_string)
							switch valStr {
							case "make":
								build_context.export_dependencies_format = DependenciesExportMake
							case "json":
								build_context.export_dependencies_format = DependenciesExportJson
							default:
								gb_printf_err("Invalid export format for -export-dependencies:<string>, got %s\n", goStr(value.Value_string))
								gb_printf_err("Valid export formats:\n")
								gb_printf_err("\tmake\n")
								gb_printf_err("\tjson\n")
								bad_flags = true
							}
						case BuildFlagExportDependenciesFile:
							export_path := string_trim_whitespace(value.Value_string)
							if is_build_flag_path_valid(export_path) {
								build_context.export_dependencies_file = path_to_full_path(heap_allocator(), export_path)
							} else {
								gb_printf_err("Invalid -export-dependencies path, got %s\n", goStr(export_path))
								bad_flags = true
							}
						case BuildFlagShowDefineables:
							build_context.show_defineables = true
						case BuildFlagExportDefineables:
							export_path := string_trim_whitespace(value.Value_string)
							if is_build_flag_path_valid(export_path) {
								build_context.export_defineables_file = path_to_full_path(heap_allocator(), export_path)
							} else {
								gb_printf_err("Invalid -export-defineables path, got %s\n", goStr(export_path))
								bad_flags = true
							}
						case BuildFlagIgnoreUnusedDefineables:
							build_context.ignore_unused_defineables = true
						case BuildFlagShowSystemCalls:
							build_context.show_system_calls = true
						case BuildFlagThreadCount:
							count := isize(big_int_to_i64(&value.Value_integer))
							if count <= 0 {
								gb_printf_err("%s expected a positive non-zero number, got %s\n", goStr(name), goStr(param))
								build_context.thread_count = 1
							} else {
								build_context.thread_count = count
							}
						case BuildFlagKeepTempFiles:
							build_context.keep_temp_files = true
						case BuildFlagCollection:
							str := value.Value_string
							eq_pos := isize(-1)
							for i := isize(0); i < str.Len; i++ {
								if str.Data[i] == '=' {
									eq_pos = i
									break
								}
							}
							if eq_pos < 0 {
								gb_printf_err("Expected 'name=path', got '%s'\n", goStr(param))
								bad_flags = true
								break
							}
							coll_name := substring(str, 0, eq_pos)
							coll_path := substring(str, eq_pos+1, str.Len)
							if coll_name.Len == 0 || coll_path.Len == 0 {
								gb_printf_err("Expected 'name=path', got '%s'\n", goStr(param))
								bad_flags = true
								break
							}
							if !string_is_valid_identifier(coll_name) {
								gb_printf_err("Library collection name '%s' must be a valid identifier\n", goStr(coll_name))
								bad_flags = true
								break
							}
							if goStr(coll_name) == "_" {
								gb_printf_err("Library collection name cannot be an underscore\n")
								bad_flags = true
								break
							}
							if goStr(coll_name) == "system" {
								gb_printf_err("Library collection name 'system' is reserved\n")
								bad_flags = true
								break
							}
							var prev_path String
							found_path := find_library_collection_path(coll_name, &prev_path)
							if found_path {
								gb_printf_err("Library collection '%s' already exists with path '%s'\n", goStr(coll_name), goStr(prev_path))
								bad_flags = true
								break
							}
							a := heap_allocator()
							path_ok := false
							fullpath := path_to_fullpath(a, coll_path, &path_ok)
							if !path_ok || !path_is_directory(fullpath) {
								display_path := coll_path
								if path_ok {
									display_path = fullpath
								}
								gb_printf_err("Library collection '%s' path must be a directory, got '%s'\n", goStr(coll_name), goStr(display_path))
								bad_flags = true
								break
							}
							add_library_collection(coll_name, coll_path)
							continue
						case BuildFlagDefine:
							str := value.Value_string
							eq_pos := isize(-1)
							for i := isize(0); i < str.Len; i++ {
								if str.Data[i] == '=' {
									eq_pos = i
									break
								}
							}
							if eq_pos < 0 {
								gb_printf_err("Expected 'name=value', got '%s'\n", goStr(param))
								bad_flags = true
								break
							}
							def_name := substring(str, 0, eq_pos)
							def_value := substring(str, eq_pos+1, str.Len)
							if def_name.Len == 0 || def_value.Len == 0 {
								gb_printf_err("Expected 'name=value', got '%s'\n", goStr(param))
								bad_flags = true
								break
							}
							if !string_is_valid_identifier(def_name) {
								gb_printf_err("Defined constant name '%s' must be a valid identifier\n", goStr(def_name))
								bad_flags = true
								break
							}
							if goStr(def_name) == "_" {
								gb_printf_err("Defined constant name cannot be an underscore\n")
								bad_flags = true
								break
							}
							key := string_intern_cstring(def_name)
							if map_get(&build_context.defined_values, key) != nil {
								gb_printf_err("Defined constant '%s' already exists\n", goStr(def_name))
								bad_flags = true
								break
							}
							v := build_param_to_exact_value(def_name, def_value)
							if v.Kind != ExactValue_Invalid {
								map_set(&build_context.defined_values, key, v)
							} else {
								gb_printf_err("Invalid define constant value: '%s'. Define constants must be a valid Odin literal.\n", goStr(def_value))
								bad_flags = true
							}
						case BuildFlagTarget:
							str := value.Value_string
							targetFound := false
							for i := range named_targets {
								if str_eq_ignore_case(str, named_targets[i].name) {
									targetFound = true
									selected_target_metrics = &named_targets[i]
									break
								}
							}
							if !targetFound {
								strStr := goStr(str)
								if strStr != "?" {
									type distanceAndTargetIndex struct {
										distance    isize
										targetIndex isize
									}
									distances := make([]distanceAndTargetIndex, len(named_targets))
									for i := range named_targets {
										distances[i].targetIndex = isize(i)
										distances[i].distance = levenstein_distance_case_insensitive(str, named_targets[i].name)
									}
									gb_sort(distances, func(a, b distanceAndTargetIndex) int {
										if a.distance < b.distance {
											return -1
										} else if a.distance > b.distance {
											return 1
										}
										return 0
									})
									gb_printf_err("Unknown target '%s'\n", goStr(str))
									if distances[0].distance <= MAX_SMALLEST_DID_YOU_MEAN_DISTANCE {
										gb_printf_err("Did you mean:\n")
										for i := 0; i < len(distances); i++ {
											if distances[i].distance > MAX_SMALLEST_DID_YOU_MEAN_DISTANCE {
												break
											}
											gb_printf_err("\t%s\n", goStr(named_targets[distances[i].targetIndex].name))
										}
									}
								}
								gb_printf_err("All supported targets:\n")
								for i := range named_targets {
									gb_printf_err("\t%s\n", goStr(named_targets[i].name))
								}
								bad_flags = true
							}
						case BuildFlagSubtarget:
							if selected_target_metrics == nil {
								gb_printf_err("-target must be set before -subtarget is used\n")
								bad_flags = true
							} else {
								str := value.Value_string
								osStr := goStr(selected_target_metrics.metrics.Metrics_os)
								osStrLower := strings.ToLower(osStr)
								if osStrLower != goStr(TargetOs_darwin) && osStrLower != goStr(TargetOs_linux) {
									gb_printf_err("-subtarget can only be used with darwin and linux based targets at the moment\n")
									bad_flags = true
									break
								}
								subFound := false
								for i := uint32(1); i < uint32(SubtargetCOUNT); i++ {
									if str_eq_ignore_case(str, subtarget_strings[i]) {
										selected_subtarget = Subtarget(i)
										subFound = true
										break
									}
								}
								if !subFound {
									gb_printf_err("Unknown subtarget '%s'\n", goStr(str))
									gb_printf_err("All supported subtargets:\n")
									for i := uint32(1); i < uint32(SubtargetCOUNT); i++ {
										gb_printf_err("\t%s\n", goStr(subtarget_strings[i]))
									}
									bad_flags = true
								}
							}
						case BuildFlagBuildMode:
							str := value.Value_string
							if goStr(build_context.command) != "build" {
								gb_printf_err("'build-mode' can only be used with the 'build' command\n")
								bad_flags = true
								break
							}
							modeStr := goStr(str)
							switch modeStr {
							case "dll", "shared", "dynamic":
								build_context.build_mode = BuildMode_DynamicLibrary
							case "obj", "object":
								build_context.build_mode = BuildMode_Object
							case "static", "lib":
								build_context.build_mode = BuildMode_StaticLibrary
							case "exe":
								build_context.build_mode = BuildMode_Executable
							case "asm", "assembly", "assembler":
								build_context.build_mode = BuildMode_Assembly
							case "llvm", "llvm-ir":
								build_context.build_mode = BuildMode_LLVM_IR
							case "test":
								build_context.build_mode = BuildMode_Executable
								build_context.command_kind = Command_test
							default:
								gb_printf_err("Unknown build mode '%s'\n", goStr(str))
								gb_printf_err("Valid build modes:\n")
								gb_printf_err("\tdll, shared, dynamic\n")
								gb_printf_err("\tlib, static\n")
								gb_printf_err("\tobj, object\n")
								gb_printf_err("\texe\n")
								gb_printf_err("\tasm, assembly, assembler\n")
								gb_printf_err("\tllvm, llvm-ir\n")
								gb_printf_err("\ttest\n")
								bad_flags = true
							}
						case BuildFlagKeepExecutable:
							build_context.keep_executable = true
						case BuildFlagDebug:
							build_context.ODIN_DEBUG = true
						case BuildFlagDisableAssert:
							build_context.ODIN_DISABLE_ASSERT = true
						case BuildFlagNoBoundsCheck:
							build_context.no_bounds_check = true
						case BuildFlagNoTypeAssert:
							build_context.no_type_assert = true
						case BuildFlagNoDynamicLiterals:
							gb_printf_err("Warning: Use of -no-dynamic-literals is now redundant\n")
						case BuildFlagDynamicLiterals:
							build_context.dynamic_literals = true
						case BuildFlagNoCRT:
							build_context.no_crt = true
						case BuildFlagNoRPath:
							build_context.no_rpath = true
						case BuildFlagNoEntryPoint:
							build_context.no_entry_point = true
						case BuildFlagNoThreadLocal:
							build_context.no_thread_local = true
						case BuildFlagLinker:
							linker_choice := Linker_Invalid
							valStr := goStr(value.Value_string)
							for i := 0; i < LinkerCOUNT; i++ {
								if goStr(linker_choices[i]) == valStr {
									linker_choice = LinkerChoice(i)
									break
								}
							}
							if linker_choice == Linker_Invalid {
								gb_printf_err("Invalid option for -linker:<string>. Expected one of the following\n")
								for i := 0; i < LinkerCOUNT; i++ {
									gb_printf_err("\t%s\n", goStr(linker_choices[i]))
								}
								bad_flags = true
							} else {
								build_context.linker_choice = linker_choice
							}
						case BuildFlagUseSeparateModules:
							if build_context.use_single_module {
								gb_printf_err("-use-separate-modules cannot be used with -use-single-module\n")
								bad_flags = true
							}
							build_context.use_separate_modules = true
						case BuildFlagUseSingleModule:
							if build_context.use_separate_modules {
								gb_printf_err("-use-single-module cannot be used with -use-separate-modules\n")
								bad_flags = true
							}
							build_context.use_single_module = true
						case BuildFlagNoThreadedChecker:
							build_context.no_threaded_checker = true
						case BuildFlagShowDebugMessages:
							build_context.show_debug_messages = true
						case BuildFlagDidYouMeanLimit:
							count := isize(big_int_to_i64(&value.Value_integer))
							if count <= 0 {
								gb_printf_err("%s expected a positive non-zero number, got %s\n", goStr(name), goStr(param))
								build_context.did_you_mean_limit = DEFAULT_DID_YOU_MEAN_LIMIT
							} else {
								build_context.did_you_mean_limit = int(count)
							}
						case BuildFlagVet:
							build_context.vet_flags |= VetFlag_All
						case BuildFlagVetUnusedVariables:
							build_context.vet_flags |= VetFlag_UnusedVariables
						case BuildFlagVetUnusedImports:
							build_context.vet_flags |= VetFlag_UnusedImports
						case BuildFlagVetUnused:
							build_context.vet_flags |= VetFlag_Unused
						case BuildFlagVetShadowing:
							build_context.vet_flags |= VetFlag_Shadowing
						case BuildFlagVetUsingStmt:
							build_context.vet_flags |= VetFlag_UsingStmt
						case BuildFlagVetUsingParam:
							build_context.vet_flags |= VetFlag_UsingParam
						case BuildFlagVetStyle:
							build_context.vet_flags |= VetFlag_Style
						case BuildFlagVetSemicolon:
							build_context.vet_flags |= VetFlag_Semicolon
						case BuildFlagVetCast:
							build_context.vet_flags |= VetFlag_Cast
						case BuildFlagVetTabs:
							build_context.vet_flags |= VetFlag_Tabs
						case BuildFlagVetUnusedProcedures:
							build_context.vet_flags |= VetFlag_UnusedProcedures
						case BuildFlagVetPackages:
							{
								val := value.Value_string
								it := String_Iterator{Str: val, Pos: 0}
								for {
									pkg := string_split_iterator(&it, ',')
									if pkg.Len == 0 {
										break
									}
									pkg = string_trim_whitespace(pkg)
									if !string_is_valid_identifier(pkg) {
										gb_printf_err("-%s '%s' must be a valid identifier\n", goStr(name), goStr(pkg))
										bad_flags = true
										continue
									}
									string_set_add(&build_context.vet_packages, pkg)
								}
							}
						case BuildFlagCustomAttribute:
							{
								val := value.Value_string
								it := String_Iterator{Str: val, Pos: 0}
								for {
									attr := string_split_iterator(&it, ',')
									if attr.Len == 0 {
										break
									}
									attr = string_trim_whitespace(attr)
									if !string_is_valid_identifier(attr) {
										gb_printf_err("-%s '%s' must be a valid identifier\n", goStr(name), goStr(attr))
										bad_flags = true
										continue
									}
									string_set_add(&build_context.custom_attributes, attr)
								}
							}
						case BuildFlagIgnoreUnknownAttributes:
							build_context.ignore_unknown_attributes = true
						case BuildFlagExtraLinkerFlags:
							build_context.extra_linker_flags = value.Value_string
						case BuildFlagExtraAssemblerFlags:
							build_context.extra_assembler_flags = value.Value_string
						case BuildFlagMicroarch:
							build_context.microarch = value.Value_string
							string_to_lower(&build_context.microarch)
						case BuildFlagTargetFeatures:
							build_context.target_features_string = value.Value_string
							string_to_lower(&build_context.target_features_string)
						case BuildFlagStrictTargetFeatures:
							build_context.strict_target_features = true
						case BuildFlagMinimumOSVersion:
							build_context.minimum_os_version_string = value.Value_string
							build_context.minimum_os_version_string_given = true
						case BuildFlagRelocMode:
							v := value.Value_string
							vStr := goStr(v)
							switch vStr {
							case "default":
								build_context.reloc_mode = RelocMode_Default
							case "static":
								build_context.reloc_mode = RelocMode_Static
							case "pic":
								build_context.reloc_mode = RelocMode_PIC
							case "dynamic-no-pic":
								build_context.reloc_mode = RelocMode_DynamicNoPIC
							default:
								gb_printf_err("-reloc-mode flag expected one of the following\n")
								gb_printf_err("\tdefault\n")
								gb_printf_err("\tstatic\n")
								gb_printf_err("\tpic\n")
								gb_printf_err("\tdynamic-no-pic\n")
								bad_flags = true
							}
						case BuildFlagDisableRedZone:
							build_context.disable_red_zone = true
						case BuildFlagDisableUnwind:
							build_context.disable_unwind = true
						case BuildFlagDisallowDo:
							build_context.disallow_do = true
						case BuildFlagNoRTTI:
							if goStr(name) == "disallow-rtti" {
								gb_printf_err("'-disallow-rtti' has been replaced with '-no-rtti'\n")
								bad_flags = true
							}
							build_context.no_rtti = true
						case BuildFlagDynamicMapCalls:
							build_context.dynamic_map_calls = true
						case BuildFlagObfuscateSourceCodeLocations:
							gb_printf_err("'-obfuscate-source-code-locations' is now deprecated in favor of '-source-code-locations:obfuscated'\n")
							build_context.source_code_location_info = SourceCodeLocationInfo_Obfuscated
						case BuildFlagSourceCodeLocations:
							{
								valStr := value.Value_string
								if str_eq_ignore_case(valStr, S("normal")) {
									build_context.source_code_location_info = SourceCodeLocationInfo_Normal
								} else if str_eq_ignore_case(valStr, S("obfuscated")) {
									build_context.source_code_location_info = SourceCodeLocationInfo_Obfuscated
								} else if str_eq_ignore_case(valStr, S("filename")) {
									build_context.source_code_location_info = SourceCodeLocationInfo_Filename
								} else if str_eq_ignore_case(valStr, S("none")) {
									build_context.source_code_location_info = SourceCodeLocationInfo_None
								} else {
									gb_printf_err("-source-code-locations:<string> options are 'normal', 'obfuscated', 'filename', and 'none'\n")
									bad_flags = true
								}
							}
						case BuildFlagDefaultToNilAllocator:
							if build_context.ODIN_DEFAULT_TO_PANIC_ALLOCATOR {
								gb_printf_err("'-default-to-panic-allocator' cannot be used with '-default-to-nil-allocator'\n")
								bad_flags = true
							}
							build_context.ODIN_DEFAULT_TO_NIL_ALLOCATOR = true
						case BuildFlagDefaultToPanicAllocator:
							if build_context.ODIN_DEFAULT_TO_NIL_ALLOCATOR {
								gb_printf_err("'-default-to-nil-allocator' cannot be used with '-default-to-panic-allocator'\n")
								bad_flags = true
							}
							build_context.ODIN_DEFAULT_TO_PANIC_ALLOCATOR = true
						case BuildFlagForeignKeyErrorProcedures:
							build_context.ODIN_FOREIGN_ERROR_PROCEDURES = true
						case BuildFlagStrictStyle:
							build_context.strict_style = true
						case BuildFlagShort:
							build_context.cmd_doc_flags |= CmdDocFlag_Short
						case BuildFlagInSourceOrder:
							build_context.cmd_doc_flags |= CmdDocFlag_InSourceOrder
						case BuildFlagAllPackages:
							build_context.cmd_doc_flags |= CmdDocFlag_AllPackages
							build_context.test_all_packages = true
						case BuildFlagDocFormat:
							build_context.cmd_doc_flags |= CmdDocFlag_DocFormat
						case BuildFlagIgnoreWarnings:
							if build_context.warnings_as_errors {
								gb_printf_err("-ignore-warnings cannot be used with -warnings-as-errors\n")
								bad_flags = true
							} else {
								build_context.ignore_warnings = true
							}
						case BuildFlagWarningsAsErrors:
							if build_context.ignore_warnings {
								gb_printf_err("-warnings-as-errors cannot be used with -ignore-warnings\n")
								bad_flags = true
							} else {
								build_context.warnings_as_errors = true
							}
						case BuildFlagTerseErrors:
							build_context.hide_error_line = true
							build_context.terse_errors = true
						case BuildFlagVerboseErrors:
							gb_printf_err("-verbose-errors is now the default, -terse-errors can disable it\n")
							build_context.hide_error_line = false
							build_context.terse_errors = false
						case BuildFlagJsonErrors:
							build_context.json_errors = true
						case BuildFlagErrorPosStyle:
							valStr := value.Value_string
							if str_eq_ignore_case(valStr, S("odin")) || str_eq_ignore_case(valStr, S("default")) {
								build_context.ODIN_ERROR_POS_STYLE = ErrorPosStyle_Default
							} else if str_eq_ignore_case(valStr, S("unix")) {
								build_context.ODIN_ERROR_POS_STYLE = ErrorPosStyle_Unix
							} else {
								gb_printf_err("-error-pos-style options are 'unix', 'odin', and 'default' (odin)\n")
								bad_flags = true
							}
						case BuildFlagMaxErrorCount:
							count := big_int_to_i64(&value.Value_integer)
							if count <= 0 {
								gb_printf_err("-%s must be greater than 0", bf.Name)
								bad_flags = true
							} else {
								build_context.max_error_count = isize(count)
							}
						case BuildFlagMinLinkLibs:
							build_context.min_link_libs = true
						case BuildFlagExportLinkedLibraries:
							build_context.export_linked_libs_path = string_trim_whitespace(value.Value_string)
							if build_context.export_linked_libs_path.Len == 0 {
								gb_printf_err("-%s specified an empty path\n", goStr(name))
								bad_flags = true
							}
						case BuildFlagPrintLinkerFlags:
							build_context.print_linker_flags = true
						case BuildFlagIntegerDivisionByZero:
							{
								v := value.Value_string
								if str_eq_ignore_case(v, S("trap")) {
									build_context.integer_division_by_zero_behaviour = IntegerDivisionByZero_Trap
								} else if str_eq_ignore_case(v, S("zero")) {
									build_context.integer_division_by_zero_behaviour = IntegerDivisionByZero_Zero
								} else if str_eq_ignore_case(v, S("self")) {
									build_context.integer_division_by_zero_behaviour = IntegerDivisionByZero_Self
								} else if str_eq_ignore_case(v, S("all-bits")) {
									build_context.integer_division_by_zero_behaviour = IntegerDivisionByZero_AllBits
								} else {
									gb_printf_err("-integer-division-by-zero options are 'trap', 'zero', 'self', and 'all-bits'.\n")
									bad_flags = true
								}
							}
						case BuildFlagBuildDiagnostics:
							build_context.build_diagnostics = true
						case BuildFlagInternalFastISel:
							build_context.fast_isel = true
						case BuildFlagInternalIgnoreLazy:
							build_context.ignore_lazy = true
						case BuildFlagInternalIgnoreLLVMBuild:
							build_context.ignore_llvm_build = true
						case BuildFlagInternalIgnorePanic:
							build_context.ignore_panic = true
						case BuildFlagInternalModulePerFile:
							build_context.module_per_file = true
							build_context.use_separate_modules = true
						case BuildFlagInternalCached:
							build_context.cached = true
							build_context.use_separate_modules = true
						case BuildFlagInternalNoInline:
							build_context.internal_no_inline = true
						case BuildFlagInternalByValue:
							build_context.internal_by_value = true
						case BuildFlagInternalWeakMonomorphization:
							build_context.internal_weak_monomorphization = true
						case BuildFlagInternalLLVMVerification:
							build_context.internal_ignore_llvm_verification = true
						case BuildFlagInternalLLVMNoSROA:
							build_context.internal_llvm_no_sroa = true
						case BuildFlagInternalEnableRVO:
							build_context.enable_rvo = true
						case BuildFlagSanitize:
							if build_context.sanitizer_flags != 0 {
								gb_printf_err("-sanitize:<string> may only be used once\n")
								bad_flags = true
							}
							{
								v := value.Value_string
								if str_eq_ignore_case(v, S("address")) {
									build_context.sanitizer_flags |= SanitizerFlag_Address
								} else if str_eq_ignore_case(v, S("memory")) {
									build_context.sanitizer_flags |= SanitizerFlag_Memory
								} else if str_eq_ignore_case(v, S("thread")) {
									build_context.sanitizer_flags |= SanitizerFlag_Thread
								} else {
									gb_printf_err("-sanitize:<string> options are 'address', 'memory', and 'thread'\n")
									bad_flags = true
								}
							}
						case BuildFlagLTO:
							{
								v := value.Value_string
								if str_eq_ignore_case(v, S("thin")) {
									build_context.lto_kind = LTO_Thin
									if build_context.linker_choice == Linker_Invalid || build_context.linker_choice == Linker_Default {
										build_context.linker_choice = Linker_lld
									}
									if !build_context.use_separate_modules {
										build_context.use_separate_modules = true
										if build_context.use_single_module {
											gb_printf_err("-linker:<string> cannot be used with -use-single-module\n")
											bad_flags = true
										}
									}
								} else if str_eq_ignore_case(v, S("thin-files")) {
									build_context.lto_kind = LTO_Thin_Files
									if build_context.linker_choice == Linker_Invalid {
										build_context.linker_choice = Linker_lld
									}
									if !build_context.use_separate_modules {
										build_context.use_separate_modules = true
										if build_context.use_single_module {
											gb_printf_err("-linker:<string> cannot be used with -use-single-module\n")
											bad_flags = true
										}
									}
								} else {
									gb_printf_err("-lto:<string> options are 'thin' and 'thin-files'\n")
									bad_flags = true
								}
							}
						case BuildFlagIgnoreVsSearch:
							build_context.ignore_microsoft_magic = true
						case BuildFlagResourceFile:
							path := value.Value_string
							path = string_trim_whitespace(path)
							if is_build_flag_path_valid(path) {
								pathStr := goStr(path)
								is_resource := strings.HasSuffix(pathStr, ".rc") || strings.HasSuffix(pathStr, ".res")
								if !is_resource {
									gb_printf_err("Invalid -resource path %s, missing .rc or .res file\n", goStr(path))
									bad_flags = true
									break
								} else if !gb_file_exists(goStr(path)) {
									gb_printf_err("Invalid -resource path %s, file does not exist.\n", goStr(path))
									bad_flags = true
									break
								}
								build_context.resource_filepath = path
								build_context.has_resource = true
							} else {
								gb_printf_err("Invalid -resource path, got %s\n", goStr(path))
								bad_flags = true
							}
						case BuildFlagWindowsPdbName:
							path := value.Value_string
							path = string_trim_whitespace(path)
							if is_build_flag_path_valid(path) {
								if path_is_directory(path) {
									gb_printf_err("Invalid -pdb-name path. %s, is a directory.\n", goStr(path))
									bad_flags = true
									break
								}
								build_context.pdb_filepath = path
							} else {
								gb_printf_err("Invalid -pdb-name path, got %s\n", goStr(path))
								bad_flags = true
							}
						case BuildFlagSubsystem:
							subsystem := value.Value_string
							subsystem_found := false
							for i := 1; i < Windows_SubsystemCOUNT; i++ {
								if str_eq_ignore_case(subsystem, windows_subsystem_names[i]) {
									build_context.ODIN_WINDOWS_SUBSYSTEM = Windows_Subsystem(i)
									subsystem_found = true
									break
								}
							}
							if !subsystem_found {
								if str_eq_ignore_case(subsystem, S("WINDOW")) {
									build_context.ODIN_WINDOWS_SUBSYSTEM = Windows_Subsystem_WINDOWS
									subsystem_found = true
									break
								}
							}
							if !subsystem_found {
								gb_printf_err("Invalid -subsystem string, got %s. Expected one of:\n", goStr(subsystem))
								gb_printf_err("\t")
								for i := 1; i < Windows_SubsystemCOUNT; i++ {
									if i > 1 {
										gb_printf_err(", ")
									}
									gb_printf_err("%s", goStr(windows_subsystem_names[i]))
									if Windows_Subsystem(i) == Windows_Subsystem_CONSOLE {
										gb_printf_err(" (default)")
									}
									if Windows_Subsystem(i) == Windows_Subsystem_WINDOWS {
										gb_printf_err(" (or WINDOW)")
									}
								}
								gb_printf_err("\n")
								bad_flags = true
							}
						case BuildFlagAndroidKeystore:
							build_context.android_keystore = value.Value_string
						case BuildFlagAndroidKeystoreAlias:
							build_context.android_keystore_alias = value.Value_string
						case BuildFlagAndroidKeystorePassword:
							build_context.android_keystore_password = value.Value_string
						}
					}
				}
				if !bf.AllowMultiple {
					set_flags[bf.Kind] = ok
				}
			}
			break
		}
		if found && !is_supported {
			gb_printf_err("Unknown flag for 'odin %s': '%s'\n", goStr(build_context.command), goStr(name))
			gb_printf_err("'%s' is supported with the following commands:\n", goStr(name))
			gb_printf_err("\t")
			count := 0
			for i := uint64(0); i < 64; i++ {
				if found_bf.CommandSupport&(1<<i) != 0 {
					if count > 0 {
						gb_printf_err(", ")
					}
					gb_printf_err("%s", odin_command_strings[i])
					count++
				}
			}
			gb_printf_err("\n")
			bad_flags = true
		} else if !found {
			did_you_mean_flag(name)
			bad_flags = true
		}

	}

	if set_flags[BuildFlagVetUnusedProcedures] && !set_flags[BuildFlagVetPackages] {
		gb_printf_err("-vet-unused-procedures must be used with -vet-packages\n")
		bad_flags = true
	}

	if build_context.export_timings_format != TimingsExportUnspecified && build_context.export_timings_file.Len == 0 {
		gb_printf_err("`-export-timings:<format>` requires `-export-timings-file:<filename>` to be specified as well\n")
		bad_flags = true
	} else if build_context.export_timings_format == TimingsExportUnspecified && build_context.export_timings_file.Len > 0 {
		gb_printf_err("`-export-timings-file:<filename>` requires `-export-timings:<format>` to be specified as well\n")
		bad_flags = true
	}

	if build_context.export_timings_format != TimingsExportUnspecified && !(build_context.show_timings || build_context.show_more_timings) {
		gb_printf_err("`-export-timings:<format>` requires `-show-timings` or `-show-more-timings` to be present\n")
		bad_flags = true
	}

	if build_context.export_dependencies_format != DependenciesExportUnspecified && build_context.print_linker_flags {
		gb_printf_err("-export-dependencies cannot be used with -print-linker-flags\n")
		bad_flags = true
	} else if build_context.show_timings && build_context.print_linker_flags {
		gb_printf_err("-show-timings/-show-more-timings cannot be used with -print-linker-flags\n")
		bad_flags = true
	}

	if (build_context.command_kind&(Command_doc|Command_test)) == 0 && build_context.test_all_packages {
		gb_printf_err("`-test-all-packages` can only be used with `odin build -build-mode:test`, `odin test`, or `odin doc`.\n")
		bad_flags = true
	}

	if build_context.did_you_mean_limit == 0 {
		build_context.did_you_mean_limit = DEFAULT_DID_YOU_MEAN_LIMIT
	}

	return !bad_flags
}
