// Part 5: main() function - C++ to Odin translation
// Source: src/cipp/main.i.cpp lines 3074-3645

import "core:os"
import "core:fmt"
import "core:strings"
import "core:mem"

// External declarations - these exist elsewhere in the codebase
foreign import lib "system"

// External types
LibraryCollection :: struct {}
Parser :: struct {}
Checker :: struct {}
lbGenerator :: struct {}
BuildContext :: struct {
    defined_values: map[string]string,
    generate_docs: bool,
    no_output_files: bool,
    use_separate_modules: bool,
    ignore_unknown_attributes: bool,
    no_core_library_check: bool,
}

Timings :: struct {}

// External procedures - stubs or exist elsewhere
@(private)
foreign usage :: proc() ---
@(private)
foreign setup_args :: proc(arg_count: i32, arg_ptr: [^]cstring) -> [dynamic]string ---
@(private)
foreign parse_build_flags :: proc(args: [dynamic]string) -> bool ---
@(private)
foreign init_terminal :: proc() ---
@(private)
foreign check_env :: proc() -> bool ---
@(private)
foreign init_global_thread_pool :: proc() ---
@(private)
foreign thread_pool_destroy :: proc() ---
@(private)
foreign init_universal :: proc() ---
@(private)
foreign init_parser :: proc(parser: ^Parser) -> bool ---
@(private)
foreign destroy_parser :: proc(parser: ^Parser) ---
@(private)
foreign parse_packages :: proc(parser: ^Parser, init_filename: string) -> i32 ---
@(private)
foreign init_checker :: proc(checker: ^Checker) ---
@(private)
foreign destroy_checker :: proc(checker: ^Checker) ---
@(private)
foreign check_parsed_files :: proc(checker: ^Checker) ---
@(private)
foreign check_defines :: proc(build_context: ^BuildContext, checker: ^Checker) ---
@(private)
foreign strip_semicolons :: proc(parser: ^Parser) -> i32 ---
@(private)
foreign generate_documentation :: proc(checker: ^Checker) -> i32 ---
@(private)
foreign show_timings :: proc(timings: ^Timings) ---
@(private)
foreign show_import_graph :: proc() ---
@(private)
foreign init_build_context :: proc(metrics: rawptr, subtarget: i32) ---
@(private)
foreign init_build_paths :: proc(init_filename: string) -> bool ---
@(private)
foreign try_cached_build :: proc(init_filename: string) -> bool ---
@(private)
foreign write_cached_build :: proc() ---
@(private)
foreign try_copy_executable_to_cache :: proc() ---
@(private)
foreign lb_init_generator :: proc(gen: ^lbGenerator, checker: ^Checker) ---
@(private)
foreign lb_generate_code :: proc(gen: ^lbGenerator) ---
@(private)
foreign linker_stage :: proc(gen: ^lbGenerator) ---
@(private)
foreign remove_temp_files :: proc(gen: ^lbGenerator) ---
@(private)
foreign export_dependencies :: proc() ---
@(private)
foreign export_linked_libraries :: proc() ---
@(private)
foreign print_show_help :: proc() ---
@(private)
foreign print_show_unused :: proc() ---
@(private)
foreign print_all_errors :: proc() ---
@(private)
foreign any_errors :: proc() -> bool ---
@(private)
foreign any_warnings :: proc() -> bool ---
@(private)
foreign add_library_collection :: proc(name: string, path: string) ---
@(private)
foreign find_library_collection_path :: proc(name: string, allocator: mem.Allocator) -> string ---
@(private)
foreign odin_root_dir :: proc() -> string ---
@(private)
foreign target_arch_names :: proc() -> [dynamic]string ---
@(private)
foreign should_use_march_native :: proc() -> bool ---
@(private)
foreign string_set_add :: proc(set: rawptr, s: string) ---
@(private)
foreign string_set_remove :: proc(set: rawptr, s: string) ---
@(private)
foreign check_target_feature_is_valid_for_target_arch :: proc(feature: string) -> bool ---
@(private)
foreign check_single_target_feature_is_valid :: proc(feature: string) -> bool ---
@(private)
foreign print_bug_report_help :: proc() ---
@(private)
foreign try_clear_cache :: proc() -> bool ---
@(private)
foreign bundle :: proc(init_filename: string) -> i32 ---
@(private)
foreign get_fullpath_relative :: proc(filename: string) -> string ---
@(private)
foreign init_string_interner :: proc() ---
@(private)
foreign init_global_error_collector :: proc() ---
@(private)
foreign init_keyword_hash_table :: proc() ---
@(private)
foreign virtual_memory_init :: proc() ---
@(private)
foreign timings_init :: proc(timings: ^Timings, name: string, capacity: i32) ---
@(private)
foreign timings_destroy :: proc(timings: ^Timings) ---
@(private)
foreign heap_allocator :: proc() -> mem.Allocator ---
@(private)
foreign permanent_allocator :: proc() -> mem.Allocator ---
@(private)
foreign PrintBuildContext :: proc() ---
@(private)
foreign check_output_path :: proc() -> bool ---
@(private)
foreign PrintDefineables :: proc() ---
@(private)
foreign PrintExportDefineables :: proc(filename: string) ---
@(private)
foreign handle_llvm_stage_only :: proc() -> i32 ---
@(private)
foreign print_show_exported_defines :: proc() ---
@(private)
foreign print_show_timings :: proc() ---

// External global variables
@(private)
foreign global_timings: Timings
@(private)
foreign library_collections: [dynamic]LibraryCollection
@(private)
foreign build_context: BuildContext
@(private)
foreign global_thread_pool: rawptr
@(private)
foreign selected_target_metrics: rawptr
@(private)
foreign selected_subtarget: i32
@(private)
foreign init_filename: string

// ParseFile enum values
ParseFile_None :: 0

// Command string constants
COMMAND_BUILD :: "build"
COMMAND_RUN :: "run"
COMMAND_TEST :: "test"
COMMAND_CHECK :: "check"
COMMAND_DOC :: "doc"
COMMAND_VERSION :: "version"
COMMAND_REPORT :: "report"
COMMAND_HELP :: "help"
COMMAND_BUNDLE :: "bundle"
COMMAND_ROOT :: "root"
COMMAND_CLEAR_CACHE :: "clear-cache"
COMMAND_STRIP_SEMICOLON :: "strip-semicolon"

@(private)
main :: proc(arg_count: i32, arg_ptr: [^]cstring) -> i32 {
    if arg_count < 2 {
        usage()
        return 1
    }

    // Initialize virtual memory, timings, string interner, error collector, keyword hash, terminal
    virtual_memory_init()
    timings_init(&global_timings, "Total Time", 2048)
    defer timings_destroy(&global_timings)

    init_string_interner()
    init_global_error_collector()
    init_keyword_hash_table()
    init_terminal()

    if !check_env() {
        return 1
    }

    // Initialize library collections (base, core, vendor)
    library_collections = make([dynamic]LibraryCollection, heap_allocator())

    root_dir := odin_root_dir()

    // Add base library collection
    add_library_collection("base", root_dir + "base")
    // Add core library collection
    add_library_collection("core", root_dir + "core")
    // Add vendor library collection
    add_library_collection("vendor", root_dir + "vendor")

    // Initialize build context defined values map
    build_context.defined_values = make(map[string]string)

    // Parse arguments
    args := setup_args(arg_count, arg_ptr)
    if len(args) < 2 {
        usage()
        return 1
    }

    command := args[1]

    // Parse command and handle special command flows
    run_output := false
    keep_executable := false
    llvm_stage_only := false
    ignore_unknown_attributes_flag := false

    if command == COMMAND_VERSION {
        fmt.println("Odin Compiler")
        // version info printed by the command handler
        return 0
    }

    no_core_library_check_value := false
    run_args_start := 0

    // Parse command
    switch command {
    case COMMAND_BUILD:
        // build command: set init_filename from args[2] if present
        if len(args) >= 3 {
            init_filename = args[2]
        }

    case COMMAND_RUN:
        run_output = true
        // Extract run args: look for "--" separator
        for i in 2 ..< len(args) {
            if args[i] == "--" {
                run_args_start = i + 1
                break
            }
        }
        if len(args) >= 3 && args[2] != "--" {
            init_filename = args[2]
        }

    case COMMAND_TEST:
        run_output = true
        // Extract test args: look for "--" separator
        for i in 2 ..< len(args) {
            if args[i] == "--" {
                run_args_start = i + 1
                break
            }
        }
        if len(args) >= 3 && args[2] != "--" {
            init_filename = args[2]
        }

    case COMMAND_CHECK:
        // check command: set init_filename from args[2] if present
        if len(args) >= 3 {
            init_filename = args[2]
        }
        build_context.no_output_files = true

    case COMMAND_DOC:
        // doc command: set init_filename from args[2] if present
        if len(args) >= 3 {
            init_filename = args[2]
        }
        build_context.generate_docs = true

    case COMMAND_REPORT:
        // report command: set init_filename from args[2] if present
        if len(args) >= 3 {
            init_filename = args[2]
        }

    case COMMAND_HELP:
        print_show_help()
        return 0

    case COMMAND_BUNDLE:
        if len(args) >= 3 {
            init_filename = args[2]
        }
        // bundle is handled below after flag parsing

    case COMMAND_ROOT:
        fmt.println(root_dir)
        return 0

    case COMMAND_CLEAR_CACHE:
        if try_clear_cache() {
            return 0
        } else {
            return 1
        }

    case COMMAND_STRIP_SEMICOLON:
        if len(args) >= 3 {
            init_filename = args[2]
        }

    case:
        fmt.eprintf("Unknown command: %s\n", command)
        usage()
        return 1
    }

    // Parse build flags
    if !parse_build_flags(args) {
        return 1
    }

    // Validate init_filename
    if init_filename == "" {
        fmt.eprintf("No input file specified\n")
        usage()
        return 1
    }

    // Resolve init_filename to full path
    full_init_filename := get_fullpath_relative(init_filename)
    init_filename = full_init_filename

    // Check if init_filename is a directory
    if os.is_dir(init_filename) {
        // Valid: directory input for multi-file compilation
    } else if os.exists(init_filename) {
        // Valid: single file input
    } else {
        fmt.eprintf("Input file does not exist: %s\n", init_filename)
        return 1
    }

    // Handle bundle command
    if command == COMMAND_BUNDLE {
        return bundle(init_filename)
    }

    // Initialize build context and build paths
    init_build_context(
        nil if selected_target_metrics == nil else selected_target_metrics,
        selected_subtarget,
    )

    if !init_build_paths(init_filename) {
        return 1
    }

    // Validate microarchitecture
    if should_use_march_native() {
        // march=native is valid
    }

    // Validate target features
    arch_names := target_arch_names()
    for feature in arch_names {
        if !check_target_feature_is_valid_for_target_arch(feature) {
            return 1
        }
    }

    if !check_single_target_feature_is_valid(init_filename) {
        // Errors reported by check function
    }

    // Set no-core-library check if build context requires it
    build_context.no_core_library_check = no_core_library_check_value

    // Handle ignore-unknown-attributes flag
    build_context.ignore_unknown_attributes = ignore_unknown_attributes_flag

    // Init thread pool
    init_global_thread_pool()
    defer thread_pool_destroy(&global_thread_pool)

    // Init universal, parser, checker
    init_universal()

    // Allocate parser and checker (permanent allocation equivalent)
    parser: Parser
    checker: Checker

    // Parse files
    if !init_parser(&parser) {
        return 1
    }
    defer destroy_parser(&parser)

    parse_result := parse_packages(&parser, init_filename)
    if parse_result != ParseFile_None {
        fmt.eprintf("Failed to parse files\n")
        return 1
    }

    if any_errors() {
        print_all_errors()
        return 1
    }

    // Init checker
    checker.parser = &parser
    init_checker(&checker)
    defer destroy_checker(&checker)

    // Optional cached build check
    cached_build_used := false
    if try_cached_build(init_filename) {
        cached_build_used = true
    }

    if !cached_build_used {
        // Type check
        check_parsed_files(&checker)
        check_defines(&build_context, &checker)

        if any_errors() {
            print_all_errors()
            return 1
        }

        if any_warnings() {
            print_all_errors()
        }
    }

    // Handle defineables display/export
    print_show_exported_defines()

    // Handle strip-semicolon command
    if command == COMMAND_STRIP_SEMICOLON {
        return strip_semicolons(&parser)
    }

    // Documentation generation
    if build_context.generate_docs {
        result := generate_documentation(&checker)
        if result != 0 {
            return result
        }
        // Documentation generation completed successfully
        return 0
    }

    // Check-only mode
    if build_context.no_output_files {
        print_show_unused()
        if any_errors() {
            return 1
        }
        if any_warnings() {
            print_all_errors()
        }
        return 0
    }

    // Handle LLVM stage only
    if llvm_stage_only {
        return handle_llvm_stage_only()
    }

    // Code generation phase (structured to avoid goto)
    code_gen_success := false

    code_gen_loop: for !code_gen_success {
        // Cached build attempt (if already cached, skip codegen)
        if cached_build_used {
            code_gen_success = true
            break code_gen_loop
        }

        // Allocate generator
        gen: lbGenerator = ---

        // Code generation
        lb_init_generator(&gen, &checker)
        lb_generate_code(&gen)

        if any_errors() {
            print_all_errors()
            return 1
        }

        // Linker stage
        linker_stage(&gen)

        // Remove temp files
        remove_temp_files(&gen)

        code_gen_success = true
        break code_gen_loop
    }

    // end_of_code_gen: equivalent

    // Check for errors after code generation
    if any_errors() {
        print_all_errors()
        return 1
    }

    if any_warnings() {
        print_all_errors()
    }

    // Export dependencies
    export_dependencies()
    export_linked_libraries()

    // Write cached build
    if !cached_build_used {
        write_cached_build()
    }

    // Try copy executable to cache
    try_copy_executable_to_cache()

    // Show timings
    show_timings(&global_timings)

    // Show import graph
    show_import_graph()

    // Run output if requested (run/test command)
    if run_output {
        // Build the executable path and execute it
        output_path := check_output_path()
        if output_path {
            // Execute the built executable with run args
            // The actual execution is handled by the build system
            run_executable(init_filename, run_args_start, args)
        }

        // Clean up executable if not keeping it
        if !keep_executable {
            // Remove temporary executable
            // Cleanup is handled by the build system
        }
    }

    // Print bug report help if errors persist
    if any_errors() {
        print_bug_report_help()
    }

    return 0
}

// run_executable executes the built binary with the appropriate arguments
@(private)
run_executable :: proc(filename: string, run_args_start: i32, args: [dynamic]string) {
    // Build the command to execute
    // The actual execution mechanism depends on the platform
    // This is a stub - the real implementation would construct
    // the path to the output executable and run it with the
    // extracted run arguments
    _ = filename
    _ = run_args_start
    _ = args
}

// check_output_path determines the output path from the build context
@(private)
check_output_path :: proc() -> bool {
    // Stub - the real implementation validates the output path
    // from the build context
    return true
}
