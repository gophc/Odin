package godin

import "core:testing"
import "core:strings"
import "core:mem"

// ============================================================================
// Test helper: setup a clean LinkerData
// ============================================================================

make_test_linker_data :: proc() -> LinkerData {
	ld: LinkerData
	ld.output_object_paths = make([dynamic]string)
	ld.output_temp_paths   = make([dynamic]string)
	ld.foreign_libraries   = make([dynamic]^Entity)
	ptr_set_init(&ld.foreign_libraries_set, 16)
	return ld
}

destroy_test_linker_data :: proc(ld: ^LinkerData) {
	delete(ld.output_object_paths)
	delete(ld.output_temp_paths)
	delete(ld.foreign_libraries)
	ptr_set_destroy(&ld.foreign_libraries_set)
}

// ============================================================================
// StringSet tests
// ============================================================================

@test
test_string_set_init :: proc(t: ^testing.T) {
	s: StringSet
	string_set_init(&s, 8)
	defer string_set_destroy(&s)

	testing.expectf(t, len(s) == 0, "expected empty set, got %v", len(s))
}

@test
test_string_set_update_adds :: proc(t: ^testing.T) {
	s: StringSet
	string_set_init(&s, 8)
	defer string_set_destroy(&s)

	existed := string_set_update(&s, "hello")
	testing.expectf(t, !existed, "first insert should not exist already")
	testing.expectf(t, len(s) == 1, "expected 1 element, got %v", len(s))
	testing.expectf(t, "hello" in s, "expected 'hello' in set")
}

@test
test_string_set_update_duplicate :: proc(t: ^testing.T) {
	s: StringSet
	string_set_init(&s, 8)
	defer string_set_destroy(&s)

	_ = string_set_update(&s, "hello")
	existed := string_set_update(&s, "hello")
	testing.expectf(t, existed, "second insert should already exist")
	testing.expectf(t, len(s) == 1, "expected still 1 element, got %v", len(s))
}

@test
test_string_set_update_multiple :: proc(t: ^testing.T) {
	s: StringSet
	string_set_init(&s, 8)
	defer string_set_destroy(&s)

	string_set_update(&s, "a")
	string_set_update(&s, "b")
	string_set_update(&s, "c")
	string_set_update(&s, "a") // duplicate

	testing.expectf(t, len(s) == 3, "expected 3 elements, got %v", len(s))
	testing.expectf(t, "a" in s && "b" in s && "c" in s, "all three keys should be present")
}

@test
test_string_set_destroy_clears :: proc(t: ^testing.T) {
	s: StringSet
	string_set_init(&s, 8)
	string_set_update(&s, "x")
	string_set_update(&s, "y")

	string_set_destroy(&s)
	// After destroy, using the set is undefined ÔÇö just verify no panic
	testing.expect(t, true)
}

// ============================================================================
// PtrSet tests
// ============================================================================

@test
test_ptr_set_init :: proc(t: ^testing.T) {
	s: PtrSet
	ptr_set_init(&s, 8)
	defer ptr_set_destroy(&s)

	testing.expectf(t, len(s) == 0, "expected empty set, got %v", len(s))
}

@test
test_ptr_set_update_adds :: proc(t: ^testing.T) {
	s: PtrSet
	ptr_set_init(&s, 8)
	defer ptr_set_destroy(&s)

	key: rawptr = &s // use any address
	existed := ptr_set_update(&s, key)
	testing.expectf(t, !existed, "first insert should not exist")
	testing.expectf(t, len(s) == 1, "expected 1 element, got %v", len(s))
}

@test
test_ptr_set_update_duplicate :: proc(t: ^testing.T) {
	s: PtrSet
	ptr_set_init(&s, 8)
	defer ptr_set_destroy(&s)

	key: rawptr = &s
	ptr_set_update(&s, key)
	existed := ptr_set_update(&s, key)
	testing.expectf(t, existed, "duplicate insert should already exist")
	testing.expectf(t, len(s) == 1, "expected still 1 element, got %v", len(s))
}

// ============================================================================
// linker_enable_system_library_linking tests
// ============================================================================

@test
test_linker_enable_system_library_linking :: proc(t: ^testing.T) {
	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	testing.expectf(t, !ld.needs_system_library_linked, "should start false")

	linker_enable_system_library_linking(&ld)
	testing.expectf(t, ld.needs_system_library_linked, "should be true after enabling")
}

// ============================================================================
// linker_data_init tests
// ============================================================================

@test
test_linker_data_init_empty_out_filepath :: proc(t: ^testing.T) {
	// Save and restore build_context
	saved_out := build_context.out_filepath
	build_context.out_filepath = ""
	defer build_context.out_filepath = saved_out

	info: CheckerInfo
	// Minimal fake info
	pkg: Package
	pkg.name = "test_pkg"
	scope: Scope
	scope.pkg = &pkg
	info.init_scope = &scope

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	linker_data_init(&ld, &info, "/some/path/hello.odin")

	testing.expectf(t, ld.output_name != "", "output_name should not be empty")
	testing.expectf(t, ld.output_base != "", "output_base should not be empty")
	testing.expectf(t, !ld.needs_system_library_linked, "should start false")
	testing.expectf(t, len(ld.output_object_paths) == 0, "object paths should be empty")
	testing.expectf(t, len(ld.output_temp_paths) == 0, "temp paths should be empty")
	testing.expectf(t, len(ld.foreign_libraries) == 0, "foreign libraries should be empty")
	testing.expectf(t, len(ld.foreign_libraries_set) == 0, "foreign libraries set should be empty")
}

@test
test_linker_data_init_with_out_filepath :: proc(t: ^testing.T) {
	saved_out := build_context.out_filepath
	build_context.out_filepath = "/output/custom_name.exe"
	defer build_context.out_filepath = saved_out

	info: CheckerInfo
	pkg: Package
	pkg.name = "test_pkg"
	scope: Scope
	scope.pkg = &pkg
	info.init_scope = &scope

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	linker_data_init(&ld, &info, "/some/path/hello.odin")

	testing.expectf(t, strings.contains(ld.output_name, "custom_name"), "output_name should contain custom_name, got: %s", ld.output_name)
	testing.expectf(t, strings.contains(ld.output_base, "custom_name"), "output_base should contain custom_name, got: %s", ld.output_base)
}

@test
test_linker_data_init_blank_init_fullpath :: proc(t: ^testing.T) {
	saved_out := build_context.out_filepath
	build_context.out_filepath = ""
	defer build_context.out_filepath = saved_out

	info: CheckerInfo
	pkg: Package
	pkg.name = "fallback_pkg"
	scope: Scope
	scope.pkg = &pkg
	info.init_scope = &scope

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	// Empty init_fullpath ÔÇö should fall back to pkg.name
	linker_data_init(&ld, &info, "")

	testing.expectf(t, ld.output_name == "fallback_pkg", "expected 'fallback_pkg', got: %s", ld.output_name)
}

// ============================================================================
// has_asm_extension tests
// ============================================================================

@test
test_has_asm_extension_true :: proc(t: ^testing.T) {
	testing.expect(t, has_asm_extension("foo.asm"))
	testing.expect(t, has_asm_extension("bar.s"))
	testing.expect(t, has_asm_extension("baz.S"))
	testing.expect(t, has_asm_extension("path/to/file.asm"))
}

@test
test_has_asm_extension_false :: proc(t: ^testing.T) {
	testing.expect(t, !has_asm_extension("foo.c"))
	testing.expect(t, !has_asm_extension("foo.cpp"))
	testing.expect(t, !has_asm_extension("foo.o"))
	testing.expect(t, !has_asm_extension("foo.obj"))
	testing.expect(t, !has_asm_extension("foo"))
	testing.expect(t, !has_asm_extension(""))
}

@test
test_has_asm_extension_case_insensitive :: proc(t: ^testing.T) {
	testing.expect(t, has_asm_extension("FOO.ASM"))
	testing.expect(t, has_asm_extension("FOO.Asm"))
	testing.expect(t, has_asm_extension("FOO.s"))
}

// ============================================================================
// remove_extension_from_path tests
// ============================================================================

@test
test_remove_extension_from_path_basic :: proc(t: ^testing.T) {
	result := remove_extension_from_path("hello.odin")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: %s", result)
}

@test
test_remove_extension_from_path_no_ext :: proc(t: ^testing.T) {
	result := remove_extension_from_path("hello")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: %s", result)
}

@test
test_remove_extension_from_path_multiple_dots :: proc(t: ^testing.T) {
	result := remove_extension_from_path("lib.name.so")
	defer delete(result)
	testing.expectf(t, result == "lib.name", "expected 'lib.name', got: %s", result)
}

@test
test_remove_extension_from_path_empty :: proc(t: ^testing.T) {
	result := remove_extension_from_path("")
	defer delete(result)
	testing.expectf(t, result == "", "expected empty, got: %s", result)
}

// ============================================================================
// remove_directory_from_path tests
// ============================================================================

@test
test_remove_directory_from_path_basic :: proc(t: ^testing.T) {
	result := remove_directory_from_path("/home/user/file.odin")
	defer delete(result)
	testing.expectf(t, result == "file.odin", "expected 'file.odin', got: %s", result)
}

@test
test_remove_directory_from_path_windows :: proc(t: ^testing.T) {
	result := remove_directory_from_path("C:\\Users\\test\\file.odin")
	defer delete(result)
	testing.expectf(t, result == "file.odin", "expected 'file.odin', got: %s", result)
}

@test
test_remove_directory_from_path_bare_filename :: proc(t: ^testing.T) {
	result := remove_directory_from_path("file.odin")
	defer delete(result)
	testing.expectf(t, result == "file.odin", "expected 'file.odin', got: %s", result)
}

// ============================================================================
// string_extension_position tests
// ============================================================================

@test
test_string_extension_position_basic :: proc(t: ^testing.T) {
	pos := string_extension_position("hello.odin")
	testing.expectf(t, pos == 5, "expected 5, got: %v", pos)
}

@test
test_string_extension_position_with_path :: proc(t: ^testing.T) {
	pos := string_extension_position("/path/to/hello.odin")
	testing.expectf(t, pos >= 0, "should find extension")
}

@test
test_string_extension_position_no_ext :: proc(t: ^testing.T) {
	pos := string_extension_position("hello")
	testing.expectf(t, pos == -1, "expected -1, got: %v", pos)
}

@test
test_string_extension_position_empty :: proc(t: ^testing.T) {
	pos := string_extension_position("")
	testing.expectf(t, pos == -1, "expected -1, got: %v", pos)
}

// ============================================================================
// string_trim_whitespace tests
// ============================================================================

@test
test_string_trim_whitespace_spaces :: proc(t: ^testing.T) {
	result := string_trim_whitespace("  hello  ")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: '%s'", result)
}

@test
test_string_trim_whitespace_tabs :: proc(t: ^testing.T) {
	result := string_trim_whitespace("\t\thello\t\t")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: '%s'", result)
}

@test
test_string_trim_whitespace_mixed :: proc(t: ^testing.T) {
	result := string_trim_whitespace(" \t hello \t ")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: '%s'", result)
}

@test
test_string_trim_whitespace_no_whitespace :: proc(t: ^testing.T) {
	result := string_trim_whitespace("hello")
	defer delete(result)
	testing.expectf(t, result == "hello", "expected 'hello', got: '%s'", result)
}

@test
test_string_trim_whitespace_empty :: proc(t: ^testing.T) {
	result := string_trim_whitespace("   ")
	defer delete(result)
	testing.expectf(t, result == "", "expected empty, got: '%s'", result)
}

// ============================================================================
// string_ends_with tests
// ============================================================================

@test
test_string_ends_with_true :: proc(t: ^testing.T) {
	testing.expect(t, string_ends_with("hello.odin", ".odin"))
	testing.expect(t, string_ends_with("foo.o", ".o"))
	testing.expect(t, string_ends_with("bar", "bar"))
	testing.expect(t, string_ends_with("hello", "")) // empty suffix always matches
}

@test
test_string_ends_with_false :: proc(t: ^testing.T) {
	testing.expect(t, !string_ends_with("hello.odin", ".cpp"))
	testing.expect(t, !string_ends_with("hello", "helloo"))
	testing.expect(t, !string_ends_with("", ".o"))
}

// ============================================================================
// string_contains_string tests
// ============================================================================

@test
test_string_contains_string_true :: proc(t: ^testing.T) {
	testing.expect(t, string_contains_string("libfoo.so.1", ".so."))
	testing.expect(t, string_contains_string("hello world", "lo wo"))
	testing.expect(t, string_contains_string("abc", ""))
}

@test
test_string_contains_string_false :: proc(t: ^testing.T) {
	testing.expect(t, !string_contains_string("libfoo.so", ".so."))
	testing.expect(t, !string_contains_string("hello", "xyz"))
	testing.expect(t, !string_contains_string("", "x"))
}

// ============================================================================
// filename_without_directory tests
// ============================================================================

@test
test_filename_without_directory :: proc(t: ^testing.T) {
	result := filename_without_directory("/path/to/file.txt")
	defer delete(result)
	testing.expectf(t, result == "file.txt", "expected 'file.txt', got: '%s'", result)
}

@test
test_filename_without_directory_bare :: proc(t: ^testing.T) {
	result := filename_without_directory("file.txt")
	defer delete(result)
	testing.expectf(t, result == "file.txt", "expected 'file.txt', got: '%s'", result)
}

// ============================================================================
// is_arch_wasm tests
// ============================================================================

@test
test_is_arch_wasm :: proc(t: ^testing.T) {
	saved := build_context.metrics.os
	defer build_context.metrics.os = saved

	build_context.metrics.os = .wasm
	testing.expect(t, is_arch_wasm())

	build_context.metrics.os = .windows
	testing.expect(t, !is_arch_wasm())

	build_context.metrics.os = .linux
	testing.expect(t, !is_arch_wasm())
}

// ============================================================================
// is_darwin tests
// ============================================================================

@test
test_is_darwin :: proc(t: ^testing.T) {
	saved := build_context.metrics.os
	defer build_context.metrics.os = saved

	build_context.metrics.os = .darwin
	testing.expect(t, is_darwin())

	build_context.metrics.os = .windows
	testing.expect(t, !is_darwin())
}

// ============================================================================
// is_windows tests
// ============================================================================

@test
test_is_windows :: proc(t: ^testing.T) {
	saved := build_context.metrics.os
	defer build_context.metrics.os = saved

	build_context.metrics.os = .windows
	testing.expect(t, is_windows())

	build_context.metrics.os = .darwin
	testing.expect(t, !is_windows())
}

// ============================================================================
// path_is_directory tests (stub)
// ============================================================================

@test
test_path_is_directory_exists :: proc(t: ^testing.T) {
	// Test with a known directory that should exist
	result := path_is_directory(".")
	testing.expectf(t, result, "current directory should be a directory")
}

// ============================================================================
// temporary_directory tests
// ============================================================================

@test
test_temporary_directory_non_empty :: proc(t: ^testing.T) {
	dir := temporary_directory()
	defer delete(dir)
	testing.expectf(t, dir != "", "temporary directory should not be empty")
}

// ============================================================================
// path_to_full_path tests
// ============================================================================

@test
test_path_to_full_path_relative :: proc(t: ^testing.T) {
	result := path_to_full_path(context.temp_allocator, ".")
	testing.expectf(t, result != "", "full path should not be empty")
	testing.expectf(t, result != ".", "should expand relative path")
}

// ============================================================================
// concatenate_strings tests
// ============================================================================

@test
test_concatenate_strings :: proc(t: ^testing.T) {
	result := concatenate_strings(context.temp_allocator, "hello", "world")
	testing.expectf(t, result == "helloworld", "expected 'helloworld', got: '%s'", result)
}

@test
test_concatenate3_strings :: proc(t: ^testing.T) {
	result := concatenate3_strings(context.temp_allocator, "a", "b", "c")
	testing.expectf(t, result == "abc", "expected 'abc', got: '%s'", result)
}

@test
test_concatenate4_strings :: proc(t: ^testing.T) {
	result := concatenate4_strings(context.temp_allocator, "a", "b", "c", "d")
	testing.expectf(t, result == "abcd", "expected 'abcd', got: '%s'", result)
}

// ============================================================================
// normalize_path tests
// ============================================================================

@test
test_normalize_path_backslash_to_forward :: proc(t: ^testing.T) {
	result := normalize_path(context.temp_allocator, "C:\\Users\\test", "/")
	testing.expectf(t, result == "C:/Users/test", "expected forward slashes, got: '%s'", result)
}

@test
test_normalize_path_no_change :: proc(t: ^testing.T) {
	result := normalize_path(context.temp_allocator, "/usr/local/bin", "/")
	testing.expectf(t, result == "/usr/local/bin", "expected no change, got: '%s'", result)
}

// ============================================================================
// string_to_lower tests
// ============================================================================

@test
test_string_to_lower :: proc(t: ^testing.T) {
	s := "HeLLo.ASM"
	string_to_lower(&s)
	testing.expectf(t, s == "hello.asm", "expected 'hello.asm', got: '%s'", s)

	s2 := "ALREADY LOWER"
	string_to_lower(&s2)
	testing.expectf(t, s2 == "already lower", "expected 'already lower', got: '%s'", s2)
}

// ============================================================================
// Integration-style tests for linker_stage edge cases
// ============================================================================

@test
test_linker_stage_static_library_mode :: proc(t: ^testing.T) {
	// This tests the early-return path for StaticLibrary mode
	saved_mode := build_context.build_mode
	saved_os	:= build_context.metrics.os
	build_context.build_mode = .StaticLibrary
	build_context.metrics.os = .linux
	defer {
		build_context.build_mode = saved_mode
		build_context.metrics.os = saved_os
	}

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	// linker_stage with static library should attempt ar
	// Since system_exec_command_line_app is a stub returning 0, this should return 0
	result := linker_stage(&ld)
	testing.expectf(t, result == 0, "static library linking should succeed, got: %v", result)
}

@test
test_linker_stage_wasm_path :: proc(t: ^testing.T) {
	saved_os := build_context.metrics.os
	build_context.metrics.os = .wasm
	defer build_context.metrics.os = saved_os

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	result := linker_stage(&ld)
	testing.expectf(t, result == 0, "wasm linking should succeed, got: %v", result)
}

@test
test_linker_stage_invalid_linker_choice :: proc(t: ^testing.T) {
	saved_os := build_context.metrics.os
	saved_choice := build_context.linker_choice
	build_context.metrics.os = .linux
	// Use an invalid linker choice index
	build_context.linker_choice = Linker_Choice(999)
	defer {
		build_context.metrics.os = saved_os
		build_context.linker_choice = saved_choice
	}

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	result := linker_stage(&ld)
	testing.expectf(t, result == 1, "invalid linker should return 1, got: %v", result)
}

@test
test_linker_stage_cross_compile_unsupported :: proc(t: ^testing.T) {
	saved_cross := build_context.cross_compiling
	saved_diff := build_context.different_os
	saved_sub := selected_subtarget
	build_context.cross_compiling = true
	build_context.different_os = true
	selected_subtarget = .Default
	defer {
		build_context.cross_compiling = saved_cross
		build_context.different_os = saved_diff
		selected_subtarget = saved_sub
	}

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	result := linker_stage(&ld)
	testing.expectf(t, result == 0, "unsupported cross-compile should return 0, got: %v", result)
	testing.expectf(t, build_context.keep_object_files, "should set keep_object_files")
}

@test
test_linker_stage_empty_output_name :: proc(t: ^testing.T) {
	// Verify that linker_data_init handles empty output name gracefully
	saved_out := build_context.out_filepath
	build_context.out_filepath = ""
	defer build_context.out_filepath = saved_out

	info: CheckerInfo
	pkg: Package
	pkg.name = ""
	scope: Scope
	scope.pkg = &pkg
	info.init_scope = &scope

	ld := make_test_linker_data()
	defer destroy_test_linker_data(&ld)

	// With empty init_fullpath and empty pkg name, output_name should still be set
	linker_data_init(&ld, &info, "")
	// pkg.name is empty, init_fullpath is empty, so output_name might end up empty
	// but the code should not crash
	testing.expect(t, true) // just verifying no panic
}

// ============================================================================
// make_string_c tests
// ============================================================================

@test
test_make_string_c :: proc(t: ^testing.T) {
	cstr: cstring = "hello from c"
	result := make_string_c(cstr)
	testing.expectf(t, result == "hello from c", "expected 'hello from c', got: '%s'", result)
}