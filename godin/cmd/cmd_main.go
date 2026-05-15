package cmd

import (
	"fmt"
	"os"
	"strings"
)

func stringJoinAndQuote(args []string) string {
	var sb strings.Builder
	for i, arg := range args {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if strings.ContainsAny(arg, " \t") {
			sb.WriteByte('"')
			sb.WriteString(strings.ReplaceAll(arg, `\`, `\\`))
			sb.WriteString(strings.ReplaceAll(arg, `"`, `\"`))
			sb.WriteByte('"')
		} else {
			sb.WriteString(arg)
		}
	}
	return sb.String()
}

func Main() int {
	args := setup_args()

	if len(args) < 2 {
		usage(args[0])
		return 1
	}

	virtual_memory_init()

	timings_init(&global_timings, "Total Time", 2048)
	defer timings_destroy(&global_timings)

	debugf("[Section] %s\n", "initialization")
	timings_start_section(&global_timings, "initialization")

	init_string_interner()
	init_global_error_collector()
	init_keyword_hash_table()
	init_terminal()

	if !check_env() {
		return 1
	}

	debugf("[Section] %s\n", "init default library collections")
	if build_context.show_more_timings {
		timings_start_section(&global_timings, "init default library collections")
	}
	library_collections = make([]LibraryCollection, 0)

	addCollection := func(name string) {
		var ok bool
		path := get_fullpath_relative(heap_allocator(), odin_root_dir(), name, &ok)
		if !ok {
			compiler_error("Cannot find the library collection '%s'. Is the ODIN_ROOT set up correctly?", name)
		}
		add_library_collection(name, path)
	}

	addCollection("base")
	addCollection("core")
	addCollection("vendor")

	debugf("[Section] %s\n", "init args")
	if build_context.show_more_timings {
		timings_start_section(&global_timings, "init args")
	}
	build_context.defined_values = make(map[string]string)
	build_context.extra_packages = []string{}

	command := args[1]
	initFilename := ""
	runArgsString := ""
	lastNonRunArg := len(args)

	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			break
		}
		if args[i] == "-help" || args[i] == "--help" {
			build_context.show_help = true
			return print_show_help(args[0], command)
		}
	}

	runOutput := false

	if command == "run" || command == "test" {
		if len(args) < 3 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_run
		if command == "test" {
			build_context.command_kind = Command_test
		}

		runArgs := make([]string, 0, len(args))

		runArgsStartIdx := -1
		for i := 0; i < len(args); i++ {
			if args[i] == "--" {
				runArgsStartIdx = i
				break
			}
		}
		if runArgsStartIdx != -1 {
			lastNonRunArg = runArgsStartIdx

			if runArgsStartIdx == 2 {
				usage(args[0])
				return 1
			}

			for i := runArgsStartIdx + 1; i < len(args); i++ {
				runArgs = append(runArgs, args[i])
			}
		}
		args = args[:lastNonRunArg]
		runArgsString = stringJoinAndQuote(runArgs)

		initFilename = args[2]
		runOutput = true

	} else if command == "build" {
		if len(args) < 3 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_build
		initFilename = args[2]

	} else if command == "check" {
		if len(args) < 3 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_check
		build_context.no_output_files = true
		initFilename = args[2]

	} else if command == "strip-semicolon" {
		if len(args) < 3 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_strip_semicolon
		build_context.no_output_files = true
		initFilename = args[2]

	} else if command == "doc" {
		if len(args) < 3 {
			usage(args[0])
			return 1
		}

		build_context.command_kind = Command_doc
		initFilename = args[2]
		for i := 3; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-") {
				break
			}
			build_context.extra_packages = append(build_context.extra_packages, arg)
		}
		extraCount := len(build_context.extra_packages)
		if extraCount > 0 {
			copy(args[3:], args[3+extraCount:])
			args = args[:len(args)-extraCount]
		}

		build_context.no_output_files = true
		build_context.generate_docs = true
		build_context.no_entry_point = true

	} else if command == "version" {
		if len(args) != 2 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_version
		gb_printf("%s version %s", args[0], ODIN_VERSION)
		if len(NIGHTLY) > 0 {
			gb_printf("-nightly")
		}
		if len(GIT_SHA) > 0 {
			gb_printf(":%s", GIT_SHA)
		}
		gb_printf("\n")
		return 0

	} else if command == "report" {
		if len(args) != 2 {
			usage(args[0])
			return 1
		}
		build_context.command_kind = Command_bug_report
		print_bug_report_help()
		return 0

	} else if command == "help" {
		if len(args) <= 2 {
			usage(args[0])
			return 1
		} else {
			return print_show_help(args[0], args[1], args[2])
		}

	} else if command == "bundle" {
		if len(args) < 4 {
			usage(args[0])
			return 1
		}
		if args[2] == "android" {
			build_context.command_kind = Command_bundle_android
		} else {
			gb_printf_err("Unknown package command: '%s'\n", args[2])
			usage(args[0])
			return 1
		}
		initFilename = args[3]

	} else if command == "root" {
		if len(args) != 2 {
			usage(args[0])
			return 1
		}
		gb_printf("%s", odin_root_dir())
		return 0

	} else if command == "clear-cache" {
		if try_clear_cache() {
			return 0
		}
		return 1

	} else {
		argv1 := ""
		if len(args) > 1 {
			argv1 = args[1]
		}
		usage(args[0], argv1)
		return 1
	}

	initFilename = copy_string(permanent_allocator(), initFilename)

	build_context.command = command

	build_context.custom_attributes = make(map[string]struct{})
	build_context.vet_packages = make(map[string]struct{})

	if !parse_build_flags(args) {
		return 1
	}

	if build_context.show_help {
		return print_show_help(args[0], command)
	}

	if len(initFilename) > 0 && !build_context.show_help {
		if !path_is_directory(initFilename) {
			singleFilePackage := false
			for i := 0; i < len(args); i++ {
				if i >= 3 && i <= lastNonRunArg && args[i] == "-file" {
					singleFilePackage = true
					break
				}
			}
			if !singleFilePackage {
				colonPos := -1
				for j := 0; j < len(initFilename); j++ {
					if initFilename[j] == ':' {
						colonPos = j
						break
					}
				}
				if colonPos > 0 {
					collectionName := initFilename[:colonPos]
					fileStr := initFilename[colonPos+1:]

					if collectionName == "core" {
						replaceWithBase := false
						if strings.HasPrefix(fileStr, "runtime") {
							replaceWithBase = true
						} else if strings.HasPrefix(fileStr, "intrinsics") {
							replaceWithBase = true
						} else if strings.HasPrefix(fileStr, "builtin") {
							replaceWithBase = true
						}
						if replaceWithBase {
							collectionName = "base"
						}
					}

					var baseDir string
					if find_library_collection_path(collectionName, &baseDir) {
						var ok bool
						fullpath := string_trim_whitespace(get_fullpath_relative(permanent_allocator(), baseDir, fileStr, &ok))
						if ok {
							initFilename = fullpath
							if path_is_directory(initFilename) {
								goto filenameCheckSuccess
							}
						}
					}
				}

				if !singleFilePackage {
					gb_printf_err("ERROR: `%s %s` takes a package/directory as its first argument.\n", args[0], command)
					if initFilename == "-file" {
						gb_printf_err("Did you mean `%s %s <filename.odin> -file`?\n", args[0], command)
					} else {
						if !gb_file_exists(initFilename) {
							gb_printf_err("The file '%s' was not found.\n", initFilename)
							return 1
						}
						gb_printf_err("Did you mean `%s %s %s -file`?\n", args[0], command, initFilename)
					}
					gb_printf_err("The `-file` flag tells it to treat a file as a self-contained package.\n")
					return 1
				} else {
					if !strings.HasSuffix(initFilename, ".odin") {
						gb_printf_err("Expected either a directory or a .odin file, got '%s'\n", initFilename)
						return 1
					}
					if !gb_file_exists(initFilename) {
						gb_printf_err("The file '%s' was not found.\n", initFilename)
						return 1
					}
				}
			}
		}
	filenameCheckSuccess:
	}

	if command == "bundle" {
		return bundle(initFilename)
	}

	if !find_library_collection_path("shared", nil) {
		add_library_collection("shared",
			get_fullpath_relative(heap_allocator(), odin_root_dir(), "shared", nil))
	}

	init_build_context(selected_target_metrics, selected_subtarget)

	if build_context.metrics.arch == TargetArch_i386 && build_context.metrics.os == TargetOs_windows {
		gb_printf_err("Warning: Thread-local storage is disabled on Windows i386.\n")
	}

	printMicroarchList := true
	if len(build_context.microarch) == 0 || build_context.microarch == "native" {
		printMicroarchList = false
	} else {
		marchList := target_microarch_list[build_context.metrics.arch]
		it := String_Iterator{List: marchList, Pos: 0}
		for {
			str := string_split_iterator(&it, ',')
			if str == "" {
				break
			}
			if str == build_context.microarch {
				printMicroarchList = false
				break
			}
		}
	}

	if !init_build_paths(initFilename) {
		return 1
	}

	defaultMarch := get_default_microarchitecture()
	if printMicroarchList {
		if build_context.microarch != "?" {
			gb_printf("Unknown microarchitecture '%s'.\n", build_context.microarch)
		}
		gb_printf("Possible -microarch values for target %s are:\n", target_arch_names[build_context.metrics.arch])
		gb_printf("\n")

		marchList := target_microarch_list[build_context.metrics.arch]
		it := String_Iterator{List: marchList, Pos: 0}
		for {
			str := string_split_iterator(&it, ',')
			if str == "" {
				break
			}
			if str == defaultMarch {
				gb_printf("\t%s (default)\n", str)
			} else {
				gb_printf("\t%s\n", str)
			}
		}
		return 0
	}

	march := get_final_microarchitecture()
	defaultFeatures := get_default_features()
	{
		it := String_Iterator{List: defaultFeatures, Pos: 0}
		for {
			str := string_split_iterator(&it, ',')
			if str == "" {
				break
			}
			string_set_add(&build_context.target_features_set, str)
		}
	}

	if should_use_march_native() && march == get_default_microarchitecture() {
		if command == "run" || command == "test" {
			gb_printf_err("Error: Try using '-microarch:native' as Odin defaults to %s (close to Nehalem), and your CPU seems to be older.\n", march)
			os.Exit(1)
		} else if command == "build" {
			gb_printf("Suggestion: Try using '-microarch:native' as Odin defaults to %s (close to Nehalem), and your CPU seems to be older.\n", march)
		}
	}

	if len(build_context.target_features_string) != 0 {
		targetIt := String_Iterator{List: build_context.target_features_string, Pos: 0}
		for {
			item := string_split_iterator(&targetIt, ',')
			if item == "" {
				break
			}

			strippedItem := item
			if len(strippedItem) > 0 && (strippedItem[0] == '+' || strippedItem[0] == '-') {
				strippedItem = strippedItem[1:]
			}

			var invalid string
			if !check_target_feature_is_valid_for_target_arch(strippedItem, &invalid) && strippedItem != "help" {
				if strippedItem != "?" {
					gb_printf_err("Unknown target feature '%s'.\n", invalid)
				}
				gb_printf("Possible -target-features for target %s are:\n", target_arch_names[build_context.metrics.arch])
				gb_printf("\n")

				featureList := target_features_list[build_context.metrics.arch]
				it := String_Iterator{List: featureList, Pos: 0}
				for {
					str := string_split_iterator(&it, ',')
					if str == "" {
						break
					}
					if check_single_target_feature_is_valid(defaultFeatures, str) {
						if build_context.has_ansi_terminal_colours {
							gb_printf("\t%s\x1b[38;5;244m (implied by target microarch %s)\x1b[0m\n", str, march)
						} else {
							gb_printf("\t%s (implied by current microarch %s)\n", str, march)
						}
					} else {
						gb_printf("\t%s\n", str)
					}
				}
				return 1
			}

			featureStr := item
			if len(featureStr) > 0 && featureStr[0] != '+' && featureStr[0] != '-' {
				featureStr = "+" + featureStr
			}

			negFeatureBytes := []byte(featureStr)
			switch negFeatureBytes[0] {
			case '+':
				negFeatureBytes[0] = '-'
			case '-':
				negFeatureBytes[0] = '+'
			default:
				gb_printf_err("unexpected feature prefix\n")
				return 1
			}
			negFeatureStr := string(negFeatureBytes)

			string_set_remove(&build_context.target_features_set, negFeatureStr)
			string_set_add(&build_context.target_features_set, featureStr)
		}
	}

	if build_context.metrics.arch == TargetArch_riscv64 {
		var disabled string
		if !check_target_feature_is_enabled("64bit,f,d,m", &disabled) {
			gb_printf_err("missing required target feature: \"%s\", enable it by setting a different -microarch or explicitly adding it through -target-features\n", disabled)
			os.Exit(1)
		}

		if LLVM_VERSION_MAJOR < 17 {
			gb_printf_err("Invalid LLVM version %s, RISC-V targets require at least LLVM 17\n", LLVM_VERSION_STRING)
			os.Exit(1)
		}
	}

	if build_context.show_debug_messages {
		debugf("Selected microarch: %s\n", march)
		debugf("Default microarch features: %s\n", defaultFeatures)
		debugf("Target triplet: %s\n", build_context.metrics.target_triplet)
		for i := range build_context.build_paths {
			bp := path_to_string(heap_allocator(), build_context.build_paths[i])
			debugf("build_paths[%d]: %s\n", i, bp)
		}
	}

	debugf("[Section] %s\n", "init thread pool")
	if build_context.show_more_timings {
		timings_start_section(&global_timings, "init thread pool")
	}
	init_global_thread_pool()
	defer thread_pool_destroy(&globalThreadPool)

	debugf("[Section] %s\n", "init universal")
	if build_context.show_more_timings {
		timings_start_section(&global_timings, "init universal")
	}
	init_universal()

	parser := &Parser{}
	var checker *Checker
	failedToCacheParsing := false

	debugf("[Section] %s\n", "parse files")
	timings_start_section(&global_timings, "parse files")

	if !init_parser(parser) {
		return 1
	}
	defer destroy_parser(parser)

	if parse_packages(parser, initFilename) != ParseFile_None {
		if !any_errors() {
			panic("parse_packages failed but no error was reported.")
		}
	}

	if any_errors() {
		print_all_errors()
		return 1
	}

	checker = &Checker{}
	checker.Parser = parser
	init_checker(checker)
	defer destroy_checker(checker)

	cachedBuildSuccess := false

	if build_context.cached && parser.TotalSeenLoadDirectiveCount == 0 {
		if build_context.show_more_timings {
			debugf("[Section] %s\n", "check cached build (pre-semantic check)")
			timings_start_section(&global_timings, "check cached build (pre-semantic check)")
		}
		if try_cached_build(checker, args) {
			cachedBuildSuccess = true
		}
	}

	if !cachedBuildSuccess {
		debugf("[Section] %s\n", "type check")
		timings_start_section(&global_timings, "type check")

		check_parsed_files(checker)
		if !build_context.ignore_unused_defineables {
			check_defines(&build_context, checker)
		}
		if any_errors() {
			print_all_errors()
			return 1
		}
		if any_warnings() {
			print_all_errors()
		}

		if build_context.show_defineables || build_context.export_defineables_file != "" {
			temp_alloc_defineable_strings(checker)
			sort_defineables_and_remove_duplicates(checker)

			if build_context.show_defineables {
				show_defineables(checker)
			}
			if build_context.export_defineables_file != "" {
				export_defineables(checker, build_context.export_defineables_file)
			}
		}

		if build_context.command_kind == Command_strip_semicolon {
			return strip_semicolons(parser)
		}

		if build_context.generate_docs {
			if build_context.show_more_timings {
				debugf("[Section] %s\n", "generate documentation")
				timings_start_section(&global_timings, "generate documentation")
			}
			if global_error_collector.count != 0 {
				return 1
			}
			generate_documentation(checker)

			if build_context.show_timings {
				show_timings(checker, &global_timings)
			}
			if build_context.show_import_graph {
				show_import_graph(checker)
			}
			return 0
		}

		if build_context.no_output_files {
			if build_context.show_unused {
				print_show_unused(checker)
			}
			if build_context.show_timings {
				show_timings(checker, &global_timings)
			}
			if build_context.show_import_graph {
				show_import_graph(checker)
			}
			if global_error_collector.count != 0 {
				return 1
			}
			return 0
		}

		if build_context.cached {
			if build_context.show_more_timings {
				debugf("[Section] %s\n", "check cached build")
				timings_start_section(&global_timings, "check cached build")
			}
			if try_cached_build(checker, args) {
				cachedBuildSuccess = true
			}
			if !cachedBuildSuccess {
				failedToCacheParsing = true
			}
		}

		if !cachedBuildSuccess {
			gen := &lbGenerator{}
			if !lb_init_generator(gen, checker) {
				return 1
			}

			labelCodeGen := "LLVM API Code Gen"
			if len(gen.Modules) > 1 {
				labelCodeGen = fmt.Sprintf("%s ( %4d modules )", labelCodeGen, len(gen.Modules))
			}
			debugf("[Section] %s\n", labelCodeGen)
			timings_start_section(&global_timings, labelCodeGen)

			if lb_generate_code(gen) {
				switch build_context.build_mode {
				case BuildMode_Executable, BuildMode_StaticLibrary, BuildMode_DynamicLibrary:
					result := linker_stage(gen)
					if result != 0 {
						if build_context.show_timings {
							show_timings(checker, &global_timings)
						}
						if build_context.show_import_graph {
							show_import_graph(checker)
						}
						if build_context.export_dependencies_format != DependenciesExportUnspecified {
							export_dependencies(checker)
						}
						return result
					} else {
						if build_context.export_linked_libs_path != "" {
							export_linked_libraries(gen)
						}
					}
				}
			}

			remove_temp_files(gen)

			if any_errors() {
				print_all_errors()
				return 1
			}
			if any_warnings() {
				print_all_errors()
			}
		}
	}

	if build_context.export_dependencies_format != DependenciesExportUnspecified {
		export_dependencies(checker)
	}

	if build_context.cached {
		if build_context.show_more_timings {
			debugf("[Section] %s\n", "write cached build")
			timings_start_section(&global_timings, "write cached build")
		}
		if !build_context.build_cache_data.copy_already_done {
			try_copy_executable_to_cache()
		}
		if failedToCacheParsing {
			write_cached_build(checker, args)
		}
	}

	if build_context.show_timings {
		show_timings(checker, &global_timings)
	}
	if build_context.show_import_graph {
		show_import_graph(checker)
	}

	if runOutput {
		exeName := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Output])

		system_must_exec_command_line_app("odin run", fmt.Sprintf(`"%s" %s`, exeName, runArgsString))

		if !build_context.keep_executable {
			gb_file_remove(exeName)

			if build_context.odin_debug {
				if build_context.metrics.os == TargetOs_windows || build_context.metrics.os == TargetOs_darwin {
					symbolPath := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Symbols])
					gb_file_remove(symbolPath)
				}
			}
		}
	}

	if build_context.show_more_timings {
		debugf("[Section] %s\n", "cleanup")
		timings_start_section(&global_timings, "cleanup")
	}

	return 0
}
