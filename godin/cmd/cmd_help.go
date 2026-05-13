package cmd

import (
	"fmt"
	"unsafe"
)

const (
	defaultMaxErrorCollectorCount = 36
	defaultDidYouMeanLimit        = 10
)

func gostr(s String) string {
	if s.Len == 0 {
		return ""
	}
	return unsafe.String(s.Data, int(s.Len))
}

func print_usage_line(indent i32, format string, args ...any) {
	for i := i32(0); i < indent; i++ {
		fmt.Print("\t")
	}
	fmt.Printf(format, args...)
	fmt.Println()
}

func usage(argv0 String, argv1 ...String) {
	var av1 String
	if len(argv1) > 0 {
		av1 = argv1[0]
	}
	s1 := gostr(av1)
	if s1 == "run." {
		print_usage_line(0, "Did you mean 'odin run .'?")
	} else if s1 == "build." {
		print_usage_line(0, "Did you mean 'odin build .'?")
	}
	print_usage_line(0, "%s is a tool for managing Odin source code.", gostr(argv0))
	print_usage_line(0, "Usage:")
	print_usage_line(1, "%s command [arguments]", gostr(argv0))
	print_usage_line(0, "Commands:")
	print_usage_line(1, "build             Compiles directory of .odin files, as an executable.")
	print_usage_line(1, "                  One must contain the program's entry point, all must be in the same package.")
	print_usage_line(1, "run               Same as 'build', but also then runs the newly compiled executable.")
	print_usage_line(1, "bundle            Bundles a directory in a specific layout for that platform.")
	print_usage_line(1, "check             Parses and type checks a directory of .odin files.")
	print_usage_line(1, "strip-semicolon   Parses, type checks, and removes unneeded semicolons from the entire program.")
	print_usage_line(1, "test              Builds and runs procedures with the attribute @(test) in the initial package.")
	print_usage_line(1, "doc               Generates documentation from a directory of .odin files.")
	print_usage_line(1, "version           Prints version.")
	print_usage_line(1, "report            Prints information useful to reporting a bug.")
	print_usage_line(1, "root              Prints the root path where Odin looks for the builtin collections.")
	print_usage_line(0, "")
	print_usage_line(0, "For further details on a command, invoke command help:")
	print_usage_line(1, "e.g. `odin build -help` or `odin help build`")
}

func print_show_help(arg0 String, command String, optional_flag ...String) int {
	var optFlag String
	if len(optional_flag) > 0 {
		optFlag = optional_flag[0]
	}

	help_resolved := false
	printed_usage_header := false
	printed_flags_header := false

	cmd := gostr(command)
	optFlagStr := gostr(optFlag)

	if cmd == "help" && optFlag.Len != 0 && len(optFlagStr) > 0 && optFlagStr[0] != '-' {
		command = optFlag
		cmd = optFlagStr
		optFlag = String{}
		optFlagStr = ""
	}

	sarg0 := gostr(arg0)

	print_usage_header_once := func() {
		if printed_usage_header {
			return
		}
		print_usage_line(0, "%s is a tool for managing Odin source code.", sarg0)
		print_usage_line(0, "Usage:")
		print_usage_line(1, "%s %s [arguments]", sarg0, cmd)
		print_usage_line(0, "")
		help_resolved = true
		printed_usage_header = true
	}

	if cmd == "build" {
		print_usage_header_once()
		print_usage_line(1, "build   Compiles directory of .odin files as an executable.")
		print_usage_line(2, "One must contain the program's entry point, all must be in the same package.")
		print_usage_line(2, "Use `-file` to build a single file instead.")
		print_usage_line(2, "Examples:")
		print_usage_line(3, "odin build .                     Builds package in current directory.")
		print_usage_line(3, "odin build <dir>                 Builds package in <dir>.")
		print_usage_line(3, "odin build filename.odin -file   Builds single-file package, must contain entry point.")
	} else if cmd == "run" {
		print_usage_header_once()
		print_usage_line(1, "run     Same as 'build', but also then runs the newly compiled executable.")
		print_usage_line(2, "Append an empty flag and then the args, '-- <args>', to specify args for the output.")
		print_usage_line(2, "Examples:")
		print_usage_line(3, "odin run .                     Builds and runs package in current directory.")
		print_usage_line(3, "odin run <dir>                 Builds and runs package in <dir>.")
		print_usage_line(3, "odin run filename.odin -file   Builds and runs single-file package, must contain entry point.")
	} else if cmd == "check" {
		print_usage_header_once()
		print_usage_line(1, "check   Parses and type checks directory of .odin files.")
		print_usage_line(2, "Examples:")
		print_usage_line(3, "odin check .                     Type checks package in current directory.")
		print_usage_line(3, "odin check <dir>                 Type checks package in <dir>.")
		print_usage_line(3, "odin check filename.odin -file   Type checks single-file package, must contain entry point.")
	} else if cmd == "test" {
		print_usage_header_once()
		print_usage_line(1, "test    Builds and runs procedures with the attribute @(test) in the initial package.")
	} else if cmd == "doc" {
		print_usage_header_once()
		print_usage_line(1, "doc     Generates documentation from a directory of .odin files.")
		print_usage_line(2, "Examples:")
		print_usage_line(3, "odin doc .                     Generates documentation on package in current directory.")
		print_usage_line(3, "odin doc <dir>                 Generates documentation on package in <dir>.")
		print_usage_line(3, "odin doc filename.odin -file   Generates documentation on single-file package.")
	} else if cmd == "version" {
		print_usage_header_once()
		print_usage_line(1, "version   Prints version.")
	} else if cmd == "strip-semicolon" {
		print_usage_header_once()
		print_usage_line(1, "strip-semicolon")
		print_usage_line(2, "Parses and type checks .odin file(s) and then removes unneeded semicolons from the entire project.")
	} else if cmd == "bundle" {
		print_usage_header_once()
		print_usage_line(1, "bundle <platform>   Bundles a directory in a specific layout for that platform")
		print_usage_line(2, "Supported platforms:")
		print_usage_line(3, "android")
	} else if cmd == "report" {
		print_usage_header_once()
		print_usage_line(1, "report  Prints information useful to reporting a bug.")
	} else if cmd == "root" {
		print_usage_header_once()
		print_usage_line(1, "root    Prints the root path where Odin looks for the builtin collections.")
	}

	doc := cmd == "doc"
	build := cmd == "build"
	run_or_build := cmd == "run" || cmd == "build" || cmd == "test"
	test_only := cmd == "test"
	strip_semicolon := cmd == "strip-semicolon"
	check_only := cmd == "check" || strip_semicolon
	check := run_or_build || check_only
	bundle := cmd == "bundle"

	if cmd == "help" {
		doc = true
		build = true
		run_or_build = true
		test_only = true
		strip_semicolon = true
		check_only = true
		check = true
	}

	print_flag := func(flag string) bool {
		if optFlagStr != "" {
			f := make_string_c(flag)
			i := string_index_byte(f, ':')
			if i >= 0 {
				f.Len = i
			}
			if optFlagStr != gostr(f) {
				return false
			}
		}
		print_usage_header_once()
		if !printed_flags_header {
			print_usage_line(0, "")
			print_usage_line(1, "Flags")
			print_usage_line(0, "")
			printed_flags_header = true
		}
		help_resolved = true
		print_usage_line(0, "")
		print_usage_line(1, flag)
		return true
	}

	if doc {
		if print_flag("-all-packages") {
			print_usage_line(2, "Generates documentation for all packages used in the current project.")
		}
	}
	if test_only {
		if print_flag("-all-packages") {
			print_usage_line(2, "Tests all packages imported into the given initial package.")
		}
	}

	if build {
		if print_flag("-build-mode:<mode>") {
			print_usage_line(2, "Sets the build mode.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-build-mode:exe         Builds as an executable.")
			print_usage_line(3, "-build-mode:test        Builds as an executable that executes tests.")
			print_usage_line(3, "-build-mode:dll         Builds as a dynamically linked library.")
			print_usage_line(3, "-build-mode:shared      Builds as a dynamically linked library.")
			print_usage_line(3, "-build-mode:dynamic     Builds as a dynamically linked library.")
			print_usage_line(3, "-build-mode:lib         Builds as a statically linked library.")
			print_usage_line(3, "-build-mode:static      Builds as a statically linked library.")
			print_usage_line(3, "-build-mode:obj         Builds as an object file.")
			print_usage_line(3, "-build-mode:object      Builds as an object file.")
			print_usage_line(3, "-build-mode:assembly    Builds as an assembly file.")
			print_usage_line(3, "-build-mode:assembler   Builds as an assembly file.")
			print_usage_line(3, "-build-mode:asm         Builds as an assembly file.")
			print_usage_line(3, "-build-mode:llvm-ir     Builds as an LLVM IR file.")
			print_usage_line(3, "-build-mode:llvm        Builds as an LLVM IR file.")
		}
	}

	if check {
		if print_flag("-collection:<name>=<filepath>") {
			print_usage_line(2, "Defines a library collection used for imports.")
			print_usage_line(2, "Example: -collection:shared=dir/to/shared")
			print_usage_line(2, "Usage in Code:")
			print_usage_line(3, "import \"shared:foo\"")
		}

		if print_flag("-custom-attribute:<string>") {
			print_usage_line(2, "Add a custom attribute which will be ignored if it is unknown.")
			print_usage_line(2, "This can be used with metaprogramming tools.")
			print_usage_line(2, "Examples:")
			print_usage_line(3, "-custom-attribute:my_tag")
			print_usage_line(3, "-custom-attribute:my_tag,the_other_thing")
			print_usage_line(3, "-custom-attribute:my_tag -custom-attribute:the_other_thing")
		}
	}

	if run_or_build {
		if print_flag("-debug") {
			print_usage_line(2, "Enables debug information, and defines the global constant ODIN_DEBUG to be 'true'. Sets -o:none by default.")
		}
	}

	if check {
		if print_flag("-default-to-nil-allocator") {
			print_usage_line(2, "Sets the default allocator to be the nil_allocator, an allocator which does nothing.")
		}

		if print_flag("-default-to-panic-allocator") {
			print_usage_line(2, "Sets the default allocator to be the panic_allocator, an allocator which calls panic() on any allocation attempt.")
		}

		if print_flag("-define:<name>=<value>") {
			print_usage_line(2, "Defines a scalar boolean, integer or string as global constant.")
			print_usage_line(2, "Example: -define:SPAM=123")
			print_usage_line(2, "Usage in code:")
			print_usage_line(3, "#config(SPAM, default_value)")
		}
	}

	if run_or_build {
		if print_flag("-disable-assert") {
			print_usage_line(2, "Disables the code generation of the built-in run-time 'assert' procedure, and defines the global constant ODIN_DISABLE_ASSERT to be 'true'.")
		}

		if print_flag("-disable-red-zone") {
			print_usage_line(2, "Disables red zone on a supported freestanding target.")
		}
	}

	if check {
		if print_flag("-disallow-do") {
			print_usage_line(2, "Disallows the 'do' keyword in the project.")
		}
	}

	if doc {
		if print_flag("-doc-format") {
			print_usage_line(2, "Generates documentation as the .odin-doc format (useful for external tooling).")
		}
	}

	if run_or_build {
		if print_flag("-dynamic-map-calls") {
			print_usage_line(2, "Uses dynamic map calls to minimize code generation at the cost of runtime execution.")
		}
	}

	if check {
		if print_flag("-error-pos-style:<string>") {
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-error-pos-style:unix      file/path:45:3:")
			print_usage_line(3, "-error-pos-style:odin      file/path(45:3)")
			print_usage_line(3, "-error-pos-style:default   (Defaults to 'odin'.)")
		}

		if print_flag("-export-defineables:<filename>") {
			print_usage_line(2, "Exports an overview of all the #config/#defined usages in CSV format to the given file path.")
			print_usage_line(2, "Example: -export-defineables:defineables.csv")
		}

		if print_flag("-export-dependencies:<format>") {
			print_usage_line(2, "Exports dependencies to one of a few formats. Requires `-export-dependencies-file`.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-export-dependencies:make   Exports in Makefile format")
			print_usage_line(3, "-export-dependencies:json   Exports in JSON format")
		}

		if print_flag("-export-dependencies-file:<filename>") {
			print_usage_line(2, "Specifies the filename for `-export-dependencies`.")
			print_usage_line(2, "Example: -export-dependencies-file:dependencies.d")
		}

		if print_flag("-export-timings:<format>") {
			print_usage_line(2, "Exports timings to one of a few formats. Requires `-show-timings` or `-show-more-timings`.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-export-timings:json   Exports compile time stats to JSON.")
			print_usage_line(3, "-export-timings:csv    Exports compile time stats to CSV.")
		}

		if print_flag("-export-timings-file:<filename>") {
			print_usage_line(2, "Specifies the filename for `-export-timings`.")
			print_usage_line(2, "Example: -export-timings-file:timings.json")
		}
	}

	if run_or_build {
		if print_flag("-extra-assembler-flags:<string>") {
			print_usage_line(2, "Adds extra assembler specific flags in a string.")
		}

		if print_flag("-extra-linker-flags:<string>") {
			print_usage_line(2, "Adds extra linker specific flags in a string.")
		}
	}

	if check {
		if print_flag("-file") {
			print_usage_line(2, "Tells `%s %s` to treat the given file as a self-contained package.", sarg0, cmd)
			print_usage_line(2, "This means that `<dir>/a.odin` won't have access to `<dir>/b.odin`'s contents.")
		}

		if print_flag("-foreign-error-procedures") {
			print_usage_line(2, "States that the error procedures used in the runtime are defined in a separate translation unit.")
		}

		if print_flag("-ignore-unknown-attributes") {
			print_usage_line(2, "Ignores unknown attributes.")
			print_usage_line(2, "This can be used with metaprogramming tools.")
		}
	}

	if run_or_build {
		if print_flag("-ignore-vs-search") {
			print_usage_line(2, "[Windows only]")
			print_usage_line(2, "Ignores the Visual Studio search for library paths.")
		}
	}

	if check {
		if print_flag("-ignore-warnings") {
			print_usage_line(2, "Ignores warning messages.")
		}

		if print_flag("-integer-division-by-zero:<string>") {
			print_usage_line(2, "Specifies the default behaviour for integer division by zero.")
			print_usage_line(2, "Available Options:")
			print_usage_line(3, "-integer-division-by-zero:trap        Trap on division/modulo/remainder by zero")
			print_usage_line(3, "-integer-division-by-zero:zero        x/0 == 0 and x%%0 == x and x%%%%0 == x")
			print_usage_line(3, "-integer-division-by-zero:self        x/0 == x and x%%0 == 0 and x%%%%0 == 0")
			print_usage_line(3, "-integer-division-by-zero:all-bits    x/0 == ~T(0) and x%%0 == x and x%%%%0 == x")
		}

		if print_flag("-json-errors") {
			print_usage_line(2, "Prints the error messages as json to stderr.")
		}
	}

	if run_or_build {
		if print_flag("-keep-temp-files") {
			print_usage_line(2, "Keeps the temporary files generated during compilation.")
		}
	} else if strip_semicolon {
		if print_flag("-keep-temp-files") {
			print_usage_line(2, "Keeps the temporary files generated during stripping the unneeded semicolons from files.")
		}
	}

	if test_only || run_or_build {
		if print_flag("-keep-executable") {
			print_usage_line(2, "Keep the executable generated by `odin test` or `odin run` after running it. We clean it up by default.")
			print_usage_line(2, "If you build your program or test using `odin build`, the compiler does not automatically execute")
			print_usage_line(2, "the resulting program, and this option is not applicable.")
		}
	}

	if run_or_build {
		if print_flag("-linker:<string>") {
			print_usage_line(2, "Specify the linker to use.")
			print_usage_line(2, "Choices:")
			var i i32
			for i = 0; i < Linker_COUNT; i++ {
				print_usage_line(3, "%s", gostr(linker_choices[i]))
			}
		}

		if print_flag("-lto:<string>") {
			print_usage_line(2, "States that the project is to be build with link-time optimizations.")
			print_usage_line(2, "This also enables '-use-separate-modules' (if not already set) and `-linker:lld")
			print_usage_line(2, "Choices:")
			print_usage_line(3, "thin       (one module per package)")
			print_usage_line(3, "thin-files (one module file)")
		}
	}

	if check {
		if print_flag("-max-error-count:<integer>") {
			print_usage_line(2, "Sets the maximum number of errors that can be displayed before the compiler terminates.")
			print_usage_line(2, "Must be an integer >0.")
			print_usage_line(2, "If not set, the default max error count is %d.", defaultMaxErrorCollectorCount)
		}

		if print_flag("-did-you-mean-limit:<integer>") {
			print_usage_line(2, "Sets the maximum number of suggestions the compiler provides.")
			print_usage_line(2, "Must be an integer >0.")
			print_usage_line(2, "If not set, the default limit is %d.", defaultDidYouMeanLimit)
		}
	}

	if run_or_build {
		if print_flag("-microarch:<string>") {
			print_usage_line(2, "Specifies the specific micro-architecture for the build in a string.")
			print_usage_line(2, "Examples:")
			print_usage_line(3, "-microarch:sandybridge")
			print_usage_line(3, "-microarch:native")
			print_usage_line(3, "-microarch:\"?\" for a list")
		}
	}

	if check {
		if print_flag("-min-link-libs") {
			print_usage_line(2, "If set, the number of linked libraries will be minimized to prevent duplications.")
			print_usage_line(2, "This is useful for so called \"dumb\" linkers compared to \"smart\" linkers.")
		}
	}

	if run_or_build || bundle {
		if print_flag("-minimum-os-version:<string>") {
			print_usage_line(2, "Sets the minimum OS version targeted by the application.")
			print_usage_line(2, "Default: -minimum-os-version:11.0.0")
			print_usage_line(2, "Only used when target is Darwin or subtarget is Android, if given, linking mismatched versions will emit a warning.")
		}
	}

	if run_or_build {
		if print_flag("-no-bounds-check") {
			print_usage_line(2, "Disables bounds checking program wide.")
		}

		if print_flag("-no-crt") {
			print_usage_line(2, "Disables automatic linking with the C Run Time.")
		}
	}

	if check && cmd != "test" {
		if print_flag("-no-entry-point") {
			print_usage_line(2, "Removes default requirement of an entry point (e.g. main procedure).")
		}
	}

	if run_or_build {
		if print_flag("-no-rpath") {
			print_usage_line(2, "Disables automatic addition of an rpath linked to the executable directory.")
		}

		if print_flag("-no-thread-local") {
			print_usage_line(2, "Ignores @thread_local attribute, effectively treating the program as if it is single-threaded.")
		}

		if print_flag("-no-threaded-checker") {
			print_usage_line(2, "Disables multithreading in the semantic checker stage.")
		}

		if print_flag("-no-type-assert") {
			print_usage_line(2, "Disables type assertion checking program wide.")
		}

		if print_flag("-o:<string>") {
			print_usage_line(2, "Sets the optimization mode for compilation.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-o:none")
			print_usage_line(3, "-o:minimal")
			print_usage_line(3, "-o:size")
			print_usage_line(3, "-o:speed")
			print_usage_line(3, "-o:aggressive (use this with caution)")
			print_usage_line(2, "The default is -o:minimal. If -debug is set, the default is -o:none.")
		}

		if print_flag("-source-code-locations:<string>") {
			print_usage_line(2, "Processes the file and procedure strings, and line and column numbers, stored with a 'runtime.Source_Code_Location' value.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-source-code-locations:normal")
			print_usage_line(3, "-source-code-locations:obfuscated")
			print_usage_line(3, "-source-code-locations:filename")
			print_usage_line(3, "-source-code-locations:none")
			print_usage_line(2, "The default is -source-code-locations:normal.")
		}

		if print_flag("-out:<filepath>") {
			print_usage_line(2, "Sets the file name of the outputted executable.")
			print_usage_line(2, "Example: -out:foo.exe")
		}
	}

	if doc {
		if print_flag("-out:<filepath>") {
			print_usage_line(2, "Sets the base name of the resulting .odin-doc file.")
			print_usage_line(2, "The extension can be optionally included; the resulting file will always have an extension of '.odin-doc'.")
			print_usage_line(2, "Example: -out:foo")
		}
	}

	if run_or_build {
		if print_flag("-pdb-name:<filepath>") {
			print_usage_line(2, "[Windows only]")
			print_usage_line(2, "Defines the generated PDB name when -debug is enabled.")
			print_usage_line(2, "Example: -pdb-name:different.pdb")
		}
	}

	if build {
		if print_flag("-print-linker-flags") {
			print_usage_line(2, "Prints the all of the flags/arguments that will be passed to the linker.")
		}
	}

	if run_or_build {
		if print_flag("-reloc-mode:<string>") {
			print_usage_line(2, "Specifies the reloc mode.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-reloc-mode:default")
			print_usage_line(3, "-reloc-mode:static")
			print_usage_line(3, "-reloc-mode:pic")
			print_usage_line(3, "-reloc-mode:dynamic-no-pic")
		}

		if print_flag("-resource:<filepath>") {
			print_usage_line(2, "[Windows only]")
			print_usage_line(2, "Defines the resource file for the executable.")
			print_usage_line(2, "Example: -resource:path/to/file.rc")
			print_usage_line(2, "or:      -resource:path/to/file.res for a precompiled one.")
		}

		if print_flag("-sanitize:<string>") {
			print_usage_line(2, "Enables sanitization analysis.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-sanitize:address")
			print_usage_line(3, "-sanitize:memory")
			print_usage_line(3, "-sanitize:thread")
		}
	}

	if doc {
		if print_flag("-short") {
			print_usage_line(2, "Shows shortened documentation for the packages.")
		}
		if print_flag("-in-source-order") {
			print_usage_line(2, "Shows documentation for the packages in source order within each file.")
		}
	}

	if check {
		if print_flag("-show-defineables") {
			print_usage_line(2, "Shows an overview of all the #config/#defined usages in the project.")
		}

		if print_flag("-ignore-unused-defineables") {
			print_usage_line(2, "Silence warning/error if a -define doesn't have at least one #config/#defined usage.")
		}

		if print_flag("-show-system-calls") {
			print_usage_line(2, "Prints the whole command and arguments for calls to external tools like linker and assembler.")
		}

		if print_flag("-show-import-graph") {
			print_usage_line(2, "Shows dot graph text format of the import graph of a project.")
		}

		if print_flag("-show-timings") {
			print_usage_line(2, "Shows basic overview of the timings of different stages within the compiler in milliseconds.")
		}

		if print_flag("-show-more-timings") {
			print_usage_line(2, "Shows an advanced overview of the timings of different stages within the compiler in milliseconds.")
		}
	}

	if check_only {
		if print_flag("-show-unused") {
			print_usage_line(2, "Shows unused package declarations within the current project.")
		}
		if print_flag("-show-unused-with-location") {
			print_usage_line(2, "Shows unused package declarations within the current project with the declarations source location.")
		}
	}

	if check {
		if print_flag("-strict-style") {
			print_usage_line(2, "This enforces parts of same style as the Odin compiler, prefer '-vet-style -vet-semicolon' if you do not want to match it exactly.")
			print_usage_line(2, "")
			print_usage_line(2, "Errs on unneeded tokens, such as unneeded semicolons.")
			print_usage_line(2, "Errs on missing trailing commas followed by a newline.")
			print_usage_line(2, "Errs on deprecated syntax.")
			print_usage_line(2, "Errs when the attached-brace style is not adhered to (also known as 1TBS).")
			print_usage_line(2, "Errs when 'case' labels are not in the same column as the associated 'switch' token.")
		}
	}

	if run_or_build {
		if print_flag("-strict-target-features") {
			print_usage_line(2, "Makes @(enable_target_features=\"...\") behave the same way as @(require_target_features=\"...\").")
			print_usage_line(2, "This enforces that all generated code uses features supported by the combination of -target, -microarch, and -target-features.")
		}

		if print_flag("-subsystem:<option>") {
			print_usage_line(2, "[Windows only]")
			print_usage_line(2, "Defines the subsystem for the application.")
			print_usage_line(2, "Available options:")
			print_usage_line(3, "-subsystem:console")
			print_usage_line(3, "-subsystem:windows")
		}
	}

	if build {
		if print_flag("-subtarget:<subtarget>") {
			print_usage_line(2, "[Darwin and Linux only]")
			print_usage_line(2, "Available subtargets:")
			var i u32
			for i = 1; i < Subtarget_COUNT; i++ {
				name := subtarget_strings[i]
				prefix := String{Data: unsafe.StringData("-subtarget:"), Len: isize(len("-subtarget:"))}
				help_string := concatenate_strings(temporary_allocator(), prefix, name)
				print_usage_line(3, "%s", gostr(help_string))
			}
		}
	}

	if run_or_build {
		if print_flag("-target-features:<string>") {
			print_usage_line(2, "Specifies CPU features to enable on top of the enabled features implied by -microarch.")
			print_usage_line(2, "Examples:")
			print_usage_line(3, "-target-features:atomics")
			print_usage_line(3, "-target-features:\"sse2,aes\"")
			print_usage_line(3, "-target-features:\"?\" for a list")
		}
	}

	if check {
		if print_flag("-target:<string>") {
			print_usage_line(2, "Sets the target for the executable to be built in.")
			print_usage_line(2, "Examples:")
			print_usage_line(3, "-target:linux_amd64")
			print_usage_line(3, "-target:windows_amd64")
			print_usage_line(3, "-target:\"?\" for a list")
		}

		if print_flag("-terse-errors") {
			print_usage_line(2, "Prints a terse error message without showing the code on that line and the location in that line.")
		}

		if print_flag("-thread-count:<integer>") {
			print_usage_line(2, "Overrides the number of threads the compiler will use to compile with.")
			print_usage_line(2, "Example: -thread-count:2")
		}
	}

	if run_or_build {
		if print_flag("-use-separate-modules") {
			print_usage_line(2, "The backend generates multiple build units which are then linked together.")
			print_usage_line(2, "This is the default behaviour for '-o:none' and '-o:minimal' builds.")
			print_usage_line(2, "Normally, a single build unit is generated for a standard project for '-o:speed' or '-o:size'.")
		}
		if print_flag("-use-single-module") {
			print_usage_line(2, "The backend generates only a single build unit.")
			print_usage_line(2, "This is the default behaviour for '-o:speed' or '-o:size'.")
		}
	}

	if check {
		if print_flag("-vet") {
			print_usage_line(2, "Does extra checks on the code.")
			print_usage_line(2, "Extra checks include:")
			print_usage_line(3, "-vet-unused")
			print_usage_line(3, "-vet-unused-variables")
			print_usage_line(3, "-vet-unused-imports")
			print_usage_line(3, "-vet-shadowing")
			print_usage_line(3, "-vet-using-stmt")
		}

		if print_flag("-vet-cast") {
			print_usage_line(2, "Errs on casting a value to its own type or using `transmute` rather than `cast`.")
		}

		if print_flag("-vet-packages:<comma-separated-strings>") {
			print_usage_line(2, "Sets which packages by name will be vetted.")
			print_usage_line(2, "Files with specific +vet tags will not be ignored if they are not in the packages set.")
		}

		if print_flag("-vet-semicolon") {
			print_usage_line(2, "Errs on unneeded semicolons.")
		}

		if print_flag("-vet-shadowing") {
			print_usage_line(2, "Checks for variable shadowing within procedures.")
		}

		if print_flag("-vet-style") {
			print_usage_line(2, "Errs on missing trailing commas followed by a newline.")
			print_usage_line(2, "Errs on deprecated syntax.")
			print_usage_line(2, "Does not err on unneeded tokens (unlike -strict-style).")
		}

		if print_flag("-vet-tabs") {
			print_usage_line(2, "Errs when the use of tabs has not been used for indentation.")
		}

		if print_flag("-vet-unused") {
			print_usage_line(2, "Checks for unused declarations (variables and imports).")
		}

		if print_flag("-vet-unused-imports") {
			print_usage_line(2, "Checks for unused import declarations.")
		}

		if print_flag("-vet-unused-procedures") {
			print_usage_line(2, "Checks for unused procedures.")
			print_usage_line(2, "Must be used with -vet-packages or specified on a per file with +vet tags.")
		}

		if print_flag("-vet-unused-variables") {
			print_usage_line(2, "Checks for unused variable declarations.")
		}

		if print_flag("-vet-using-param") {
			print_usage_line(2, "Checks for the use of 'using' on procedure parameters.")
			print_usage_line(2, "'using' is considered bad practice outside of immediate refactoring.")
		}

		if print_flag("-vet-using-stmt") {
			print_usage_line(2, "Checks for the use of 'using' as a statement.")
			print_usage_line(2, "'using' is considered bad practice outside of immediate refactoring.")
		}

		if print_flag("-warnings-as-errors") {
			print_usage_line(2, "Treats warning messages as error messages.")
		}
	}

	if bundle {
		print_usage_line(0, "")
		print_usage_line(1, "Android-specific flags")
		print_usage_line(0, "")
		if print_flag("-android-keystore:<string>") {
			print_usage_line(2, "Specifies the keystore file to use to sign the apk.")
		}

		if print_flag("-android-keystore-alias:<string>") {
			print_usage_line(2, "Specifies the key alias to use when signing the apk")
			print_usage_line(2, "Can be omitted if the keystore only contains one key")
		}

		if print_flag("-android-keystore-password:<string>") {
			print_usage_line(2, "Sets the password to use to unlock the keystore")
			print_usage_line(2, "If this is omitted, the terminal will prompt you to provide it.")
		}
	}

	if !help_resolved {
		usage(arg0)
		print_usage_line(0, "")
		if cmd == "help" {
			print_usage_line(0, "'%s' is not a recognized flag.", gostr(optFlag))
		} else {
			print_usage_line(0, "'%s' is not a recognized command.", cmd)
		}
		return 1
	}

	print_usage_line(0, "")
	return 0
}
