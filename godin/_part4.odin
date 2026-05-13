// Part 4: remove_temp_files, print_show_help, print_show_unused, check_env,
// StripSemicolonFile, write_file_with_stripped_tokens, strip_semicolons, init_terminal

import "core:fmt"
import "core:os"
import "core:strings"

// ===========================================================================
// Forward declarations (defined in other parts)
// ===========================================================================

BuildContext :: struct {
	keep_temp_files:              bool,
	has_ansi_terminal_colours:    bool,
	gen:                          lbGenerator,
}

lbGenerator :: struct {
	output_temp_paths:    [dynamic]string,
	output_object_paths:  [dynamic]string,
}

Checker :: struct {}
CheckerInfo :: struct {}
Entity :: struct {
	kind: EntityKind,
}
EntityKind :: enum int {
	Invalid,
	Import,
	Library,
	Package,
	// ... other kinds defined in checker
}
ScopeFlag :: enum u32 {
	Unused = 1 << 0,
	// ... other flags
}
AstPackage :: struct {
	name:     string,
	fullpath: string,
	files:    [dynamic]AstFile,
}
AstFile :: struct {
	fullpath: string,
	tokenizer: Tokenizer,
}
Tokenizer :: struct {
	fullpath: string,
}
TokenArray :: struct {
	tokens: [dynamic]Token,
}
Token :: struct {
	kind:  TokenKind,
	flags: u32,
	pos:   TokenPos,
	text:  string,
}
TokenKind :: enum int {
	Invalid,
	Semicolon,
	// ... other kinds
}
TokenPos :: struct {
	file_id: int,
	offset:  int,
	line:    int,
	column:  int,
}
TokenFlags_RemoveSemicolon  :: u32(1 << 0)
TokenFlags_ReplaceSemicolon :: u32(1 << 1)

// ===========================================================================
// External references (defined in other parts)
// ===========================================================================

print_usage_line :: proc(indent: int, fmt_string: string, args: ..any) {
	// Defined in print usage / help part
	_ = indent
	_ = fmt_string
	_ = args
}

build_context: BuildContext

// ===========================================================================
// remove_temp_files (lines 2140-2167)
// ===========================================================================

remove_temp_files :: proc() {
	if build_context.keep_temp_files {
		return
	}

	for path in build_context.gen.output_temp_paths {
		os.remove(path)
	}
	for path in build_context.gen.output_object_paths {
		os.remove(path)
	}
}

// ===========================================================================
// print_show_help (lines 2168-2807)
// ===========================================================================

print_show_help :: proc(optional_command: string, optional_flag: string) {
	// Determine which commands are active
	has_doc_cmd := optional_command == "" || optional_command == "doc"
	has_version_cmd := optional_command == "" || optional_command == "version"
	has_test_cmd := optional_command == "" || optional_command == "test"
	has_build_cmd := optional_command == "" || optional_command == "build"
	has_check_cmd := optional_command == "" || optional_command == "check"
	has_run_cmd := optional_command == "" || optional_command == "run"
	has_report_cmd := optional_command == "" || optional_command == "report"
	has_strip_semicolons_cmd := optional_command == "" || optional_command == "strip-semicolons"

	// Helper for flag printing with optional filtering
	print_flag :: proc(active: bool, flag_line: string, description: string, flag_type: string = "") {
		if !active {
			return
		}
		if optional_flag != "" && !strings.contains(flag_line, optional_flag) {
			return
		}
		print_usage_line(0, flag_line)
		if description != "" {
			print_usage_line(1, description)
		}
		if flag_type != "" {
			print_usage_line(1, flag_type)
		}
	}

	// DOC command help
	if has_doc_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin doc <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_doc_cmd, "-all-packages", "Generates documentation for all packages in the directory.")
		print_flag(has_doc_cmd, "-doc-format", "Defines the output format for documentation. Default is .odin.")
		print_flag(has_doc_cmd, "-short", "Generates shorter documentation output, only showing declarations.")
		print_flag(has_doc_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes for documentation generation.")
		print_flag(has_doc_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
	}

	// VERSION command help
	if has_version_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin version [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")
		print_usage_line(0, "-version")
		print_usage_line(1, "Prints the compiler version.")
		print_usage_line(0, "")
	}

	// TEST command help
	if has_test_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin test <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_test_cmd, "-file", "Treats the input as a single file rather than a package.")
		print_flag(has_test_cmd, "-all-packages", "Runs tests for all packages in the directory.")
		print_flag(has_test_cmd, "-o:<string>", "Sets the optimization mode for tests.", "Options: minimal, size, speed")
		print_flag(has_test_cmd, "-define:<string>", "Defines a compile-time constant (e.g., -define:ODIN_TEST_FANCY=false).")
		print_flag(has_test_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes during test compilation.")
		print_flag(has_test_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays in tests.")
		print_flag(has_test_cmd, "-strict-style", "Enforces strict style guidelines on tests.")
	}

	// BUILD command help
	if has_build_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin build <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_build_cmd, "-file", "Treats the input as a single file rather than a package.")
		print_flag(has_build_cmd, "-out:<string>", "Sets the output file path for the compiled binary.")
		print_flag(has_build_cmd, "-o:<string>", "Sets the optimization mode.", "Options: minimal, size, speed")
		print_flag(has_build_cmd, "-target:<string>", "Sets the target platform triple (e.g., windows_amd64, linux_arm64).")
		print_flag(has_build_cmd, "-define:<string>", "Defines a compile-time constant.")
		print_flag(has_build_cmd, "-build-mode:<string>", "Sets the build mode.", "Options: exe, shared, static, object, asm, llvm-ir")
		print_flag(has_build_cmd, "-debug", "Enables debug information in the output binary.")
		print_flag(has_build_cmd, "-disable-assert", "Disables runtime assertions.")
		print_flag(has_build_cmd, "-no-bounds-check", "Disables bounds checking for arrays and slices.")
		print_flag(has_build_cmd, "-no-type-assert", "Disables type assertion safety checks.")
		print_flag(has_build_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
		print_flag(has_build_cmd, "-no-thread-local", "Disables thread-local storage usage.")
		print_flag(has_build_cmd, "-lld", "Uses the LLD linker instead of the system linker.")
		print_flag(has_build_cmd, "-opt:<string>", "Sets LLVM optimization level.", "Options: none, less, default, aggressive")
		print_flag(has_build_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes during compilation.")
		print_flag(has_build_cmd, "-vet", "Enables code style vetting.")
		print_flag(has_build_cmd, "-vet-tabs", "Vets that tabs are used for indentation.")
		print_flag(has_build_cmd, "-vet-style", "Vets code style conventions.")
		print_flag(has_build_cmd, "-strict-style", "Treats style vet issues as errors.")
		print_flag(has_build_cmd, "-warnings-as-errors", "Treats all warnings as errors.")
		print_flag(has_build_cmd, "-disallow-do", "Disallows the use of the 'do' keyword.")
		print_flag(has_build_cmd, "-max-error-count:<int>", "Sets the maximum number of errors before aborting compilation.")
		print_flag(has_build_cmd, "-thread-count:<int>", "Sets the number of threads to use for compilation.")
		print_flag(has_build_cmd, "-keep-temp-files", "Keeps temporary files after compilation.")
		print_flag(has_build_cmd, "-extra-linker-flags:<string>", "Passes extra flags to the linker.")
		print_flag(has_build_cmd, "-extra-assembler-flags:<string>", "Passes extra flags to the assembler.")
		print_flag(has_build_cmd, "-pdb-name:<string>", "Sets the name for the PDB debug file (Windows only).")
		print_flag(has_build_cmd, "-subsystem:<string>", "Sets the Windows subsystem.", "Options: console, windows")
		print_flag(has_build_cmd, "-no-crt", "Disables linking against the C runtime library.")
		print_flag(has_build_cmd, "-no-entry-point", "Disables the entry point requirement (for libraries).")
		print_flag(has_build_cmd, "-ignore-microsoft-magic", "Ignores Microsoft-specific magic numbers in PDB files.")
		print_flag(has_build_cmd, "-reloc-mode:<string>", "Sets the relocation mode.", "Options: default, static, pic, dynamic-no-pic")
		print_flag(has_build_cmd, "-sanitize:<string>", "Enables the specified sanitizer.", "Options: address, memory, thread")
		print_flag(has_build_cmd, "-microarch:<string>", "Sets the target microarchitecture for CPU-specific optimizations.")
		print_flag(has_build_cmd, "-show-system-calls", "Shows system calls made by the compiler.")
		print_flag(has_build_cmd, "-export-dependencies:<string>", "Exports dependency information in the specified format.")
		print_flag(has_build_cmd, "-export-timings:<string>", "Exports timing information in the specified format.")
		print_flag(has_build_cmd, "-export-defineables:<string>", "Exports defineable information to the specified file.")
		print_flag(has_build_cmd, "-linked-libraries:<string>", "Exports linked library information to the specified file.")
		print_flag(has_build_cmd, "-error-pos-style:<string>", "Sets the style for error position display.", "Options: default, unix")
		print_flag(has_build_cmd, "-src-code-info:<string>", "Sets the source code location information style.", "Options: none, lines, full")
		print_flag(has_build_cmd, "-linker:<string>", "Selects the linker to use.", "Options: default, lld, system")
		print_flag(has_build_cmd, "-lto", "Enables link-time optimization.")
	}

	// CHECK command help
	if has_check_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin check <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_check_cmd, "-file", "Treats the input as a single file rather than a package.")
		print_flag(has_check_cmd, "-all-packages", "Checks all packages in the directory.")
		print_flag(has_check_cmd, "-target:<string>", "Sets the target platform triple for cross-compilation checking.")
		print_flag(has_check_cmd, "-define:<string>", "Defines a compile-time constant for checking.")
		print_flag(has_check_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes during checking.")
		print_flag(has_check_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays during checking.")
		print_flag(has_check_cmd, "-vet", "Enables code style vetting.")
		print_flag(has_check_cmd, "-vet-tabs", "Vets that tabs are used for indentation.")
		print_flag(has_check_cmd, "-vet-style", "Vets code style conventions.")
		print_flag(has_check_cmd, "-strict-style", "Treats style vet issues as errors.")
		print_flag(has_check_cmd, "-warnings-as-errors", "Treats all warnings as errors.")
		print_flag(has_check_cmd, "-disallow-do", "Disallows the 'do' statement.")
	}

	// RUN command help
	if has_run_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin run <directory> [flags] [-- args]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_run_cmd, "-file", "Treats the input as a single file rather than a package.")
		print_flag(has_run_cmd, "-out:<string>", "Sets the output file path for the compiled binary before running.")
		print_flag(has_run_cmd, "-o:<string>", "Sets the optimization mode.", "Options: minimal, size, speed")
		print_flag(has_run_cmd, "-target:<string>", "Sets the target platform triple.")
		print_flag(has_run_cmd, "-define:<string>", "Defines a compile-time constant.")
		print_flag(has_run_cmd, "-debug", "Enables debug information.")
		print_flag(has_run_cmd, "-disable-assert", "Disables runtime assertions.")
		print_flag(has_run_cmd, "-no-bounds-check", "Disables bounds checking.")
		print_flag(has_run_cmd, "-no-type-assert", "Disables type assertion safety checks.")
		print_flag(has_run_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes.")
		print_flag(has_run_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
		print_flag(has_run_cmd, "-strict-style", "Enforces strict style guidelines.")
		print_flag(has_run_cmd, "-disallow-do", "Disallows the 'do' statement.")
		print_flag(has_run_cmd, "-no-crt", "Disables linking against the C runtime library.")
		print_flag(has_run_cmd, "-sanitize:<string>", "Enables the specified sanitizer.", "Options: address, memory, thread")
	}

	// REPORT command help
	if has_report_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin report <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_report_cmd, "-all-packages", "Generates a report for all packages in the directory.")
		print_flag(has_report_cmd, "-define:<string>", "Defines a compile-time constant.")
		print_flag(has_report_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes.")
		print_flag(has_report_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
	}

	// STRIP-SEMICOLONS command help
	if has_strip_semicolons_cmd {
		print_usage_line(0, "USAGE:")
		print_usage_line(1, "odin strip-semicolons <directory> [flags]")
		print_usage_line(0, "")
		print_usage_line(0, "FLAGS:")
		print_usage_line(0, "")

		print_flag(has_strip_semicolons_cmd, "-all-packages", "Processes all packages in the directory.")
		print_flag(has_strip_semicolons_cmd, "-ignore-unknown-attributes", "Ignores unknown attributes.")
		print_flag(has_strip_semicolons_cmd, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
	}

	// Common flags section (shown when no specific command is filtered)
	if optional_command == "" {
		print_usage_line(0, "")
		print_usage_line(0, "COMMON FLAGS (build, run, check, test, doc, report, strip-semicolons):")
		print_usage_line(0, "")

		print_flag(true, "-file", "Treats the input as a single file rather than a package.")
		print_flag(true, "-all-packages", "Processes all packages in the directory.")
		print_flag(true, "-o:<string>", "Sets the optimization mode: minimal, size, speed.")
		print_flag(true, "-target:<string>", "Sets the target platform triple.")
		print_flag(true, "-define:<string>", "Defines a compile-time constant.")
		print_flag(true, "-debug", "Enables debug information.")
		print_flag(true, "-disable-assert", "Disables runtime assertions.")
		print_flag(true, "-no-bounds-check", "Disables bounds checking.")
		print_flag(true, "-vet", "Enables code style vetting.")
		print_flag(true, "-vet-tabs", "Vets that tabs are used for indentation.")
		print_flag(true, "-vet-style", "Vets code style conventions.")
		print_flag(true, "-strict-style", "Treats style vet issues as errors.")
		print_flag(true, "-warnings-as-errors", "Treats all warnings as errors.")
		print_flag(true, "-disallow-do", "Disallows the 'do' keyword.")
		print_flag(true, "-ignore-unknown-attributes", "Ignores unknown attributes.")
		print_flag(true, "-no-dynamic-literals", "Treats dynamic map/array literals as fixed arrays.")
		print_flag(true, "-max-error-count:<int>", "Sets the maximum number of errors before aborting.")
		print_flag(true, "-thread-count:<int>", "Sets the number of threads for compilation.")
		print_flag(true, "-keep-temp-files", "Keeps temporary files after compilation.")
		print_flag(true, "-export-dependencies:<string>", "Exports dependency information.")
		print_flag(true, "-export-timings:<string>", "Exports timing information.")
		print_flag(true, "-export-defineables:<string>", "Exports defineable information.")
		print_flag(true, "-error-pos-style:<string>", "Sets the style for error position display.", "Options: default, unix")
		print_flag(true, "-src-code-info:<string>", "Sets source code location information style.", "Options: none, lines, full")
	}
}

// ===========================================================================
// print_show_unused (lines 2808-2873)
// ===========================================================================

print_show_unused :: proc(c: ^Checker) {
	// Iterates checker info entities and prints unused package declarations.
	// Filters by EntityKind (Import, Library, etc.) and ScopeFlag.Unused,
	// grouping output by package.
	//
	// Since Checker and CheckerInfo internals are defined in other parts,
	// this is a functional stub that outlines the algorithm.

	_ = c
	// TODO: implement when CheckerInfo entity list is available

	// Original C++ algorithm:
	//   for each entity in checker->info.entities:
	//       if entity.kind == Entity_Import || entity.kind == Entity_Library etc.:
	//           if entity has scope flag Unused:
	//               print the unused import/library grouped by package
}

// ===========================================================================
// check_env (lines 2874-2891)
// ===========================================================================

check_env :: proc() {
	odin_root, found_root := os.get_env("ODIN_ROOT")
	if !found_root || odin_root == "" {
		fmt.eprintf("ODIN_ROOT environment variable is not set or is empty.\n")
		fmt.eprintf("Please set ODIN_ROOT to the root directory of the Odin compiler.\n")
		os.exit(1)
	}
}

// ===========================================================================
// StripSemicolonFile struct (lines 2892-2898)
// ===========================================================================

StripSemicolonFile :: struct {
	old_fullpath:        string,
	old_fullpath_backup: string,
	new_fullpath:        string,
	file:                os.Handle,
	written:             bool,
}

// ===========================================================================
// write_file_with_stripped_tokens (lines 2899-2935)
// ===========================================================================

write_file_with_stripped_tokens :: proc(f: ^StripSemicolonFile, file_data: []u8, tok_array: ^TokenArray) -> bool {
	// Writes file content to f.file with tokens flagged for semicolon
	// removal or replacement skipped/modified.
	//
	// Token flags checked:
	//   TokenFlags_RemoveSemicolon  — skip the semicolon token entirely
	//   TokenFlags_ReplaceSemicolon — replace with a newline
	//
	// This is a stub because TokenArray and token flag definitions are
	// provided by the tokenizer in other parts.

	_ = f
	_ = file_data
	_ = tok_array

	// Original C++ logic (pseudocode):
	//   for each token in tok_array.tokens:
	//       if token.flags & TokenFlags_RemoveSemicolon:
	//           // skip writing this token's source text
	//       else if token.flags & TokenFlags_ReplaceSemicolon:
	//           // write a newline instead
	//       else:
	//           // write the token's source text from file_data

	return false
}

// ===========================================================================
// strip_semicolons (lines 2936-3040)
// ===========================================================================

strip_semicolons :: proc(files: ^[dynamic]string) -> bool {
	// Full semicolon stripping pipeline:
	// 1. Iterate packages/files checking for tokens with semicolon flags
	// 2. Create backup filenames (append .bak to original)
	// 3. Write stripped versions to temp files
	// 4. Copy original -> backup (preserve original)
	// 5. Copy temp -> original (replace with stripped version)
	// 6. Clean up temp files

	_ = files

	// Stub — requires full tokenizer, parser infrastructure and TokenFlags
	// definitions from other parts. The actual implementation would:
	//
	// for fullpath in files:
	//     parse and tokenize the file
	//     if no semicolon flags set in any token, continue
	//
	//     sf: StripSemicolonFile
	//     sf.old_fullpath = fullpath
	//     sf.old_fullpath_backup = strings.concatenate({fullpath, ".bak"})
	//     sf.new_fullpath = strings.concatenate({fullpath, ".tmp"})
	//     sf.file = os.open(sf.new_fullpath, O_WRONLY|O_CREATE|O_TRUNC, 0o644)
	//     defer os.close(sf.file)
	//
	//     file_ok := write_file_with_stripped_tokens(&sf, file_data, tok_array)
	//     if file_ok:
	//         sf.written = true
	//         os.remove(sf.old_fullpath_backup)  // clean old backup
	//         os.rename(sf.old_fullpath, sf.old_fullpath_backup)  // backup original
	//         os.rename(sf.new_fullpath, sf.old_fullpath)         // replace original

	return false
}

// ===========================================================================
// init_terminal (lines 3041-3073)
// ===========================================================================

init_terminal :: proc() {
	// 1. Check NO_COLOR environment variable
	no_color, has_no_color := os.get_env("NO_COLOR")
	if has_no_color && no_color != "" && no_color != "0" {
		build_context.has_ansi_terminal_colours = false
		return
	}

	// 2. Check FORCE_COLOR environment variable
	force_color, has_force_color := os.get_env("FORCE_COLOR")
	if has_force_color && force_color != "" && force_color != "0" {
		build_context.has_ansi_terminal_colours = true
		return
	}

	// 3. Windows: try to enable virtual terminal processing via SetConsoleMode
	when ODIN_OS == .Windows {
		import "core:sys/windows"

		hnd := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
		if hnd != windows.INVALID_HANDLE_VALUE {
			mode: windows.DWORD
			if windows.GetConsoleMode(hnd, &mode) {
				// ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
				if windows.SetConsoleMode(hnd, mode | 0x0004) {
					build_context.has_ansi_terminal_colours = true
					return
				}
			}
		}
	}

	// 4. Check ODIN_TERMINAL environment variable
	odin_terminal, has_odin_terminal := os.get_env("ODIN_TERMINAL")
	if has_odin_terminal {
		if odin_terminal == "ansi" {
			build_context.has_ansi_terminal_colours = true
		} else if odin_terminal == "none" {
			build_context.has_ansi_terminal_colours = false
		}
	}
}
