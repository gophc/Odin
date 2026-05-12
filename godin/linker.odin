// linker.odin - Pure Odin rewrite of src/cipp/linker.i.cpp
// Package: godin
// compile linker value representation

package godin

import "core:fmt"
import "core:strings"
import "core:os"
import "core:path/filepath"
import "core:sync"
import "core:mem"

// ============================================================================
// Constants
// ============================================================================

LINKER_NAME_MSVC   :: "msvc-link"
LINKER_NAME_LLD	:: "lld-link"
LINKER_NAME_RAD	:: "rad-link"
LINKER_NAME_WASM   :: "wasm-ld"
LINKER_NAME_DSYM   :: "dsymutil"
LINKER_NAME_AR	 :: "ar"
LINKER_NAME_NASM   :: "nasm"

DOT_O_STR		:: ".o"
DOT_OBJ_STR	  :: ".obj"
DOT_A_STR		:: ".a"
DOT_SO_STR	   :: ".so"
DOT_DYLIB_STR	:: ".dylib"
DOT_FRAMEWORK_STR:: ".framework"
DOT_SO_DOT_STR   :: ".so."

NASM_WIN64_FMT   :: "win64"
NASM_ELF64_FMT   :: "elf64"
NASM_ELF32_FMT   :: "elf32"
NASM_MACHO64_FMT :: "macho64"
NASM_MACHO32_FMT :: "macho32"

NIX_SEPARATOR :: "/"

// ============================================================================
// Foreign / external type declarations (defined in other odingo3 files)
// ============================================================================

// These types are assumed to exist elsewhere in the odingo3 package.
// They mirror the C++ types used in linker.i.cpp.

Entity_Kind :: enum i32 {
	Invalid,
	LibraryName,
	// ... other values defined in entity.odin
}

LibraryName_Data :: struct {
	paths:			  [dynamic]string,
	extra_linker_flags: string,
	ignore_duplicates:  bool,
}

Entity :: struct {
	kind:		Entity_Kind,
	LibraryName: LibraryName_Data,
	// ... other fields
}

Build_Path_Kind :: enum i32 {
	Output,
	VS_LIB,
	VS_EXE,
	Win_SDK_UM_Lib,
	Win_SDK_UCRT_Lib,
	Win_SDK_Bin_Path,
	Symbols,
	RES,
	RC,
}

Build_Path :: struct {
	basename: string,
	name:	 string,
}

Target_Os :: enum i32 {
	windows,
	darwin,
	linux,
	freebsd,
	openbsd,
	haiku,
	orca,
	wasm,
	// ...
}

Target_Arch :: enum i32 {
	amd64,
	i386,
	arm64,
	riscv64,
	// ...
}

Subtarget :: enum i32 {
	Default,
	Android,
	iPhone,
	iPhoneSimulator,
}

Linker_Choice :: enum i32 {
	Default,
	lld,
	radlink,
	mold,
}

Build_Mode :: enum i32 {
	Executable,
	DynamicLibrary,
	StaticLibrary,
}

Reloc_Mode :: enum i32 {
	Default,
	PIC,
	// ...
}

LTO_Kind :: enum i32 {
	None,
	Thin,
	// ...
}

Sanitizer_Flag :: enum i32 {
	Address = 1,
	Memory  = 2,
	// ...
}

Metrics :: struct {
	os:			Target_Os,
	arch:		  Target_Arch,
	ptr_size:	  int,
	target_triplet: string,
}

Build_Context :: struct {
	ODIN_ROOT:					 string,
	ODIN_DEBUG:					bool,
	ODIN_WINDOWS_SUBSYSTEM:		int,
	ODIN_ANDROID_API_LEVEL:		int,
	ODIN_ANDROID_NDK:			  string,
	ODIN_ANDROID_NDK_TOOLCHAIN:	string,
	ODIN_ANDROID_NDK_TOOLCHAIN_LIB: string,
	ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL: string,
	ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT: string,
	metrics:					   Metrics,
	out_filepath:				  string,
	link_flags:					string,
	extra_linker_flags:			string,
	extra_assembler_flags:		 string,
	build_paths:				   [Build_Path_Kind]Build_Path,
	cross_compiling:			   bool,
	different_os:				  bool,
	linker_choice:				 Linker_Choice,
	build_mode:					Build_Mode,
	reloc_mode:					Reloc_Mode,
	lto_kind:					  LTO_Kind,
	sanitizer_flags:			   bit_set[Sanitizer_Flag],
	keep_object_files:			 bool,
	has_resource:				  bool,
	no_crt:						bool,
	no_entry_point:				bool,
	min_link_libs:				 bool,
	no_rpath:					  bool,
	show_more_timings:			 bool,
	thread_count:				  int,
	minimum_os_version_string:	 string,
	minimum_os_version_string_given: bool,
}

Timings :: struct {
	// defined in timings.odin
}

CheckerInfo :: struct {
	init_scope: ^Scope,
	// ...
}

Scope :: struct {
	pkg: ^Package,
	// ...
}

Package :: struct {
	name: string,
	// ...
}

// ============================================================================
// Globals (externally defined)
// ============================================================================

build_context: Build_Context
global_timings: Timings
selected_subtarget: Subtarget
target_os_names: [Target_Os]string
target_arch_names: [Target_Arch]string
linker_choices: [Linker_Choice]string
windows_subsystem_names: [6]string // indexed by ODIN_WINDOWS_SUBSYSTEM

// ============================================================================
// Set helper types (Odin maps as sets)
// ============================================================================

StringSet :: distinct map[string]struct{}

string_set_init :: proc(s: ^StringSet, capacity: int = 64, allocator := context.allocator) {
	s^ = make(map[string]struct{}, capacity, allocator)
}

string_set_destroy :: proc(s: ^StringSet) {
	delete(s^)
}

string_set_update :: proc(s: ^StringSet, key: string) -> bool {
	if key in s^ {
		return true
	}
	s^[key] = {}
	return false
}

PtrSet :: distinct map[rawptr]struct{}

ptr_set_init :: proc(s: ^PtrSet, capacity: int = 1024, allocator := context.allocator) {
	s^ = make(map[rawptr]struct{}, capacity, allocator)
}

ptr_set_destroy :: proc(s: ^PtrSet) {
	delete(s^)
}

ptr_set_update :: proc(s: ^PtrSet, key: rawptr) -> bool {
	if key in s^ {
		return true
	}
	s^[key] = {}
	return false
}

// ============================================================================
// LinkerData
// ============================================================================

LinkerData :: struct {
	foreign_mutex:			   sync.Mutex,
	foreign_libraries_set:	   PtrSet,
	foreign_libraries:		   [dynamic]^Entity,
	output_object_paths:		 [dynamic]string,
	output_temp_paths:		   [dynamic]string,
	output_base:				 string,
	output_name:				 string,
	needs_system_library_linked: bool,
}

// ============================================================================
// Forward declarations of helper functions
// ============================================================================

// system_exec_command_line_app: executes a command line application.
// Body is defined elsewhere (likely in os/process utilities).
system_exec_command_line_app :: proc(name: string, allocator := context.allocator, fmt_str: string, args: ..any) -> i32 {
	// TODO: implement — calls external process and returns exit code
	_ = name
	_ = allocator
	_ = fmt_str
	_ = args
	return 0
}

// system_exec_command_line_app_output: executes a command and captures its output.
system_exec_command_line_app_output :: proc(command: string, output: ^string, allocator := context.allocator) -> bool {
	// TODO: implement — calls external process and captures stdout
	_ = command
	_ = output
	_ = allocator
	return false
}

// ============================================================================
// Path and string utility functions
// ============================================================================

has_asm_extension :: proc(s: string) -> bool {
	// TODO: check known assembly extensions (.asm, .s, .S)
	lowered := strings.to_lower(s, context.temp_allocator)
	return strings.has_suffix(lowered, ".asm") ||
		   strings.has_suffix(lowered, ".s") ||
		   strings.has_suffix(lowered, ".S")
}

remove_directory_from_path :: proc(s: string, allocator := context.allocator) -> string {
	// Return just the filename portion
	return filepath.base(s)
}

remove_extension_from_path :: proc(s: string, allocator := context.allocator) -> string {
	ext := filepath.ext(s)
	if ext == "" {
		return strings.clone(s, allocator)
	}
	return strings.clone(s[:len(s)-len(ext)], allocator)
}

string_trim_whitespace :: proc(s: string, allocator := context.allocator) -> string {
	return strings.trim_space(s)
}

string_extension_position :: proc(s: string) -> int {
	// Returns the index of the last '.' in the filename portion, or -1
	base := filepath.base(s)
	last_dot := strings.last_index_byte(base, '.')
	if last_dot == -1 {
		return -1
	}
	return len(s) - len(base) + last_dot
}

string_ends_with :: proc(s: string, suffix: string) -> bool {
	return strings.has_suffix(s, suffix)
}

string_contains_string :: proc(s: string, substr: string) -> bool {
	return strings.contains(s, substr)
}

string_to_lower :: proc(s: ^string) {
	s^ = strings.to_lower(s^)
}

filename_without_directory :: proc(s: string, allocator := context.allocator) -> string {
	return filepath.base(s)
}

temporary_directory :: proc(allocator := context.allocator) -> string {
	return os.temp_dir(allocator)
}

path_to_full_path :: proc(allocator: mem.Allocator, s: string) -> string {
	return filepath.abs(s, allocator)
}

path_to_string :: proc(allocator: mem.Allocator, p: Build_Path) -> string {
	// Concatenate Build_Path components into a single path string
	// TODO: proper implementation based on Build_Path structure
	return strings.clone(p.basename, allocator)
}

quote_path :: proc(allocator: mem.Allocator, p: Build_Path) -> string {
	s := path_to_string(allocator, p)
	// already quoted in the original? return as-is for now
	return s
}

path_is_directory :: proc(s: string) -> bool {
	return os.is_dir(s, context.temp_allocator)
}

normalize_path :: proc(allocator: mem.Allocator, s: string, separator: string) -> string {
	// Normalize path separators
	result := strings.clone(s, allocator)
	result = strings.replace_all(result, "\\", separator)
	return result
}

concatenate_strings :: proc(allocator: mem.Allocator, a: string, b: string) -> string {
	return strings.concatenate({a, b}, allocator)
}

concatenate3_strings :: proc(allocator: mem.Allocator, a: string, b: string, c: string) -> string {
	return strings.concatenate({a, b, c}, allocator)
}

concatenate4_strings :: proc(allocator: mem.Allocator, a: string, b: string, c: string, d: string) -> string {
	return strings.concatenate({a, b, c, d}, allocator)
}

make_string_c :: proc(cstr: cstring) -> string {
	return string(cstr)
}

// ============================================================================
// Linker data management
// ============================================================================

linker_enable_system_library_linking :: proc(ld: ^LinkerData) {
	ld.needs_system_library_linked = true
}

linker_data_init :: proc(ld: ^LinkerData, info: ^CheckerInfo, init_fullpath: string, allocator := context.allocator) {
	ld.output_object_paths = make([dynamic]string, 0, 256, allocator)
	ld.output_temp_paths   = make([dynamic]string, 0, 16, allocator)
	ld.foreign_libraries   = make([dynamic]^Entity, 0, 1024, allocator)
	ptr_set_init(&ld.foreign_libraries_set, 1024)
	ld.needs_system_library_linked = false

	if build_context.out_filepath == "" {
		ld.output_name = remove_directory_from_path(init_fullpath, allocator)
		ld.output_name = remove_extension_from_path(ld.output_name, allocator)
		ld.output_name = string_trim_whitespace(ld.output_name, allocator)
		if ld.output_name == "" {
			ld.output_name = info.init_scope.pkg.name
		}
		ld.output_base = ld.output_name
	} else {
		ld.output_name = build_context.out_filepath
		ld.output_name = string_trim_whitespace(ld.output_name, allocator)
		if ld.output_name == "" {
			ld.output_name = info.init_scope.pkg.name
		}
		pos := string_extension_position(ld.output_name)
		if pos < 0 {
			ld.output_base = ld.output_name
		} else {
			ld.output_base = ld.output_name[:pos]
		}
	}
	ld.output_base = path_to_full_path(allocator, ld.output_base)
}

// ============================================================================
// is_arch_wasm helper
// ============================================================================

is_arch_wasm :: proc() -> bool {
	return build_context.metrics.os == .wasm
}

is_darwin :: proc() -> bool {
	return build_context.metrics.os == .darwin
}

is_windows :: proc() -> bool {
	return build_context.metrics.os == .windows
}

// ============================================================================
// sb_printf helper — like gb_string_append_fmt using a strings.Builder
// ============================================================================

sb_printf :: proc(sb: ^strings.Builder, format: string, args: ..any) {
	fmt.sbprintf(sb, format, ..args)
}

sb_write :: proc(sb: ^strings.Builder, s: string) {
	strings.write_string(sb, s)
}

// ============================================================================
// gb_get_env
// ============================================================================

gb_get_env :: proc(name: string, allocator := context.allocator) -> (string, bool) {
	return os.lookup_env(name, allocator)
}

// ============================================================================
// debugf / gb_printf_err stubs
// ============================================================================

debugf :: proc(format: string, args: ..any) {
	fmt.printf(format, ..args)
}

gb_printf_err :: proc(format: string, args: ..any) {
	fmt.eprintf(format, ..args)
}

// ============================================================================
// timings helpers
// ============================================================================

timings_start_section :: proc(t: ^Timings, name: string) {
	// TODO: implement timing section start
	_ = t
	_ = name
}

// ============================================================================
// linker_stage — the main linking function
// ============================================================================

linker_stage :: proc(gen: ^LinkerData, allocator := context.allocator) -> i32 {
	result: i32 = 0
	timings := &global_timings

	output_filename := path_to_string(allocator, build_context.build_paths[.Output])
	debugf("Linking %s\n", output_filename)

	// ----  WASM path ----
	if is_arch_wasm() {
		timings_start_section(timings, LINKER_NAME_WASM)

		lib_builder: strings.Builder
		strings.builder_init(&lib_builder, allocator)
		defer strings.builder_destroy(&lib_builder)

		extra_orca_builder: strings.Builder
		strings.builder_init(&extra_orca_builder, context.temp_allocator)

		inputs_builder: strings.Builder
		strings.builder_init(&inputs_builder, context.temp_allocator)

		sb_printf(&inputs_builder, "\"%s.o\"", output_filename)

		for e in gen.foreign_libraries {
			assert(e.kind == .LibraryName)

			extra_flags := string_trim_whitespace(e.LibraryName.extra_linker_flags, context.temp_allocator)
			if extra_flags != "" {
				sb_printf(&lib_builder, " %s", extra_flags)
			}

			for lib in e.LibraryName.paths {
				if lib == "" {
					continue
				}
				if !strings.has_suffix(lib, ".o") {
					continue
				}
				sb_printf(&inputs_builder, " \"%s\"", lib)
			}
		}

		if build_context.metrics.os == .orca {
			orca_sdk_path: string
			if !system_exec_command_line_app_output("orca sdk-path", &orca_sdk_path, context.temp_allocator) {
				gb_printf_err("executing `orca sdk-path` failed, make sure Orca is installed and added to your path\n")
				return 1
			}
			if orca_sdk_path == "" {
				gb_printf_err("executing `orca sdk-path` did not produce output\n")
				return 1
			}
			sb_printf(&inputs_builder, " \"%s/orca-libc/lib/crt1.o\" \"%s/orca-libc/lib/libc.a\"", orca_sdk_path, orca_sdk_path)
			sb_printf(&extra_orca_builder, " -L \"%s/bin\" -lorca_wasm --export-dynamic", orca_sdk_path)
		}

		result = system_exec_command_line_app(
			LINKER_NAME_WASM,
			allocator,
			"\"%s\\bin\\wasm-ld\" %s -o \"%s\" %s %s %s %s",
			build_context.ODIN_ROOT,
			strings.to_string(inputs_builder),
			output_filename,
			build_context.link_flags,
			build_context.extra_linker_flags,
			strings.to_string(lib_builder),
			strings.to_string(extra_orca_builder),
		)
		return result
	}

	// ---- Native paths ----
	is_cross_linking: bool
	is_android: bool

	if build_context.cross_compiling && (build_context.different_os || selected_subtarget != .Default) {
		switch selected_subtarget {
		case .Android:
			is_cross_linking = true
			is_android = true
			// fall through to try_cross_linking
		case:
			gb_printf_err(
				"Linking for cross compilation for this platform is not yet supported (%s %s)\n",
				target_os_names[build_context.metrics.os],
				target_arch_names[build_context.metrics.arch],
			)
			build_context.keep_object_files = true
			return 0
		}
	}

	// --- Label target for cross-linking jump ---
	// (The C++ uses a goto; we restructure with a nested scope)

	{
		section_name: string = LINKER_NAME_MSVC
		osx := is_darwin()
		win := is_windows()

		switch build_context.linker_choice {
		case .Default:
			// keep section_name as LINKER_NAME_MSVC
		case .lld:
			section_name = LINKER_NAME_LLD
		case .radlink:
			section_name = LINKER_NAME_RAD
		case:
			gb_printf_err(
				"'%s' linker is not supported on this platform\n",
				linker_choices[build_context.linker_choice],
			)
			return 1
		}

		if win {
			// ============ WINDOWS (MSVC) LINKING ============
			timings_start_section(timings, section_name)

			lib_builder: strings.Builder
			strings.builder_init(&lib_builder, allocator)
			defer strings.builder_destroy(&lib_builder)

			link_settings_builder: strings.Builder
			strings.builder_init(&link_settings_builder, allocator, 256)
			defer strings.builder_destroy(&link_settings_builder)

			if build_context.build_paths[.VS_LIB].basename != "" {
				add_path :: proc(b: ^strings.Builder, p: string) {
					s := p
					if len(s) > 0 && s[len(s)-1] == '\\' {
						s = s[:len(s)-1]
					}
					sb_printf(b, " /LIBPATH:\"%s\"", s)
				}
				add_path(&link_settings_builder, build_context.build_paths[.Win_SDK_UM_Lib].basename)
				add_path(&link_settings_builder, build_context.build_paths[.Win_SDK_UCRT_Lib].basename)
				add_path(&link_settings_builder, build_context.build_paths[.VS_LIB].basename)
			}

			min_libs_set: StringSet
			string_set_init(&min_libs_set, 64, context.temp_allocator)
			defer string_set_destroy(&min_libs_set)

			prev_lib: string

			asm_files: StringSet
			string_set_init(&asm_files, 64, context.temp_allocator)
			defer string_set_destroy(&asm_files)

			for e in gen.foreign_libraries {
				assert(e.kind == .LibraryName)

				extra_flags := string_trim_whitespace(e.LibraryName.extra_linker_flags, context.temp_allocator)
				if extra_flags != "" {
					sb_printf(&lib_builder, " %s", extra_flags)
				}

				for lib in e.LibraryName.paths {
					lib_trimmed := string_trim_whitespace(lib, context.temp_allocator)
					string_to_lower(&lib_trimmed)
					if lib_trimmed == "" {
						continue
					}

					if has_asm_extension(lib_trimmed) {
						if !string_set_update(&asm_files, lib_trimmed) {
							asm_file := lib_trimmed
							obj_file: string
							temp_dir := temporary_directory(context.temp_allocator)

							if temp_dir != "" {
								filename := filename_without_directory(asm_file, context.temp_allocator)
								obj_builder: strings.Builder
								strings.builder_init(&obj_builder, allocator)
								sb_write(&obj_builder, temp_dir)
								sb_write(&obj_builder, "/")
								sb_write(&obj_builder, filename)
								sb_printf(&obj_builder, "-%p.obj", raw_data(asm_file))
								obj_file = strings.to_string(obj_builder)
							} else {
								obj_file = concatenate_strings(allocator, asm_file, ".obj")
							}

							obj_format := NASM_WIN64_FMT
							result = system_exec_command_line_app(
								LINKER_NAME_NASM,
								allocator,
								"\"%s\\bin\\nasm\\windows\\nasm.exe\" \"%s\" -f \"%s\" -o \"%s\" %s",
								build_context.ODIN_ROOT,
								asm_file,
								obj_format,
								obj_file,
								build_context.extra_assembler_flags,
							)
							if result != 0 {
								return result
							}
							append(&gen.output_object_paths, obj_file)
						}
					} else if !string_set_update(&min_libs_set, lib_trimmed) || !build_context.min_link_libs {
						if prev_lib != lib_trimmed {
							sb_printf(&lib_builder, " \"%s\"", lib_trimmed)
						}
						prev_lib = lib_trimmed
					}
				}
			}

			if build_context.build_mode == .DynamicLibrary {
				sb_write(&link_settings_builder, " /DLL")
				if build_context.no_entry_point {
					sb_write(&link_settings_builder, " /NOENTRY")
				}
			} else {
				if !(build_context.metrics.arch == .i386 && !build_context.no_crt) {
					sb_write(&link_settings_builder, " /ENTRY:mainCRTStartup")
				}
			}

			if build_context.build_paths[.Symbols].name != "" {
				symbol_path := path_to_string(allocator, build_context.build_paths[.Symbols])
				sb_printf(&link_settings_builder, " /PDB:\"%s\"", symbol_path)
			}

			if build_context.build_mode != .StaticLibrary {
				if build_context.no_crt {
					sb_write(&link_settings_builder, " /nodefaultlib")
				} else {
					sb_write(&link_settings_builder, " /defaultlib:libcmt")
				}
			}

			if build_context.ODIN_DEBUG {
				sb_write(&link_settings_builder, " /DEBUG")
			}

			object_files_builder: strings.Builder
			strings.builder_init(&object_files_builder, allocator)
			defer strings.builder_destroy(&object_files_builder)

			for object_path in gen.output_object_paths {
				sb_printf(&object_files_builder, "\"%s\" ", object_path)
			}

			object_files_str := strings.to_string(object_files_builder)
			link_settings_str := strings.to_string(link_settings_builder)
			lib_str := strings.to_string(lib_builder)

			switch build_context.linker_choice {
			case .lld:
				lld_lto_builder: strings.Builder
				strings.builder_init(&lld_lto_builder, allocator)
				defer strings.builder_destroy(&lld_lto_builder)

				if build_context.lto_kind != .None {
					sb_printf(&lld_lto_builder, "/opt:lldltojobs=%d ", build_context.thread_count)
				}

				result = system_exec_command_line_app(
					"msvc-lld-link",
					allocator,
					"\"%s\\bin\\lld-link\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s %s",
					build_context.ODIN_ROOT,
					object_files_str,
					output_filename,
					link_settings_str,
					windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM],
					build_context.link_flags,
					build_context.extra_linker_flags,
					lib_str,
					strings.to_string(lld_lto_builder),
				)
				if result != 0 {
					return result
				}

			case .radlink:
				result = system_exec_command_line_app(
					"msvc-rad-link",
					allocator,
					"\"%s\\bin\\radlink\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s",
					build_context.ODIN_ROOT,
					object_files_str,
					output_filename,
					link_settings_str,
					windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM],
					build_context.link_flags,
					build_context.extra_linker_flags,
					lib_str,
				)
				if result != 0 {
					return result
				}

			case:
				// Default MSVC linker
				res_path := quote_path(allocator, build_context.build_paths[.RES])
				defer delete(res_path)
				rc_path  := quote_path(allocator, build_context.build_paths[.RC])
				defer delete(rc_path)

				res_path_final: string
				if build_context.has_resource {
					if build_context.build_paths[.RC].basename == "" {
						debugf("Using precompiled resource %s\n", res_path)
						res_path_final = res_path
					} else {
						debugf("Compiling resource %s\n", res_path)
						windows_sdk_bin_path := path_to_string(allocator, build_context.build_paths[.Win_SDK_Bin_Path])
						defer delete(windows_sdk_bin_path)

						result = system_exec_command_line_app(
							LINKER_NAME_MSVC,
							allocator,
							"\"%src.exe\" /nologo /fo %s %s",
							windows_sdk_bin_path,
							res_path,
							rc_path,
						)
						if result != 0 {
							return result
						}
						res_path_final = res_path
					}
				} else {
					res_path_final = ""
				}

				vs_exe_path := path_to_string(allocator, build_context.build_paths[.VS_EXE])
				defer delete(vs_exe_path)

				linker_name := "link.exe"
				link_settings_extra_builder: strings.Builder
				strings.builder_init(&link_settings_extra_builder, context.temp_allocator)

				switch build_context.build_mode {
				case .Executable:
					sb_write(&link_settings_extra_builder, " /NOIMPLIB /NOEXP")
				}

				switch build_context.build_mode {
				case .StaticLibrary:
					linker_name = "lib.exe"
				case:
					sb_write(&link_settings_extra_builder, " /incremental:no /opt:ref")
				}

				result = system_exec_command_line_app(
					LINKER_NAME_MSVC,
					allocator,
					"\"%s%s\" %s %s -OUT:\"%s\" %s%s /nologo /subsystem:%s %s %s %s",
					vs_exe_path,
					linker_name,
					object_files_str,
					res_path_final,
					output_filename,
					link_settings_str,
					strings.to_string(link_settings_extra_builder),
					windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM],
					build_context.link_flags,
					build_context.extra_linker_flags,
					lib_str,
				)
				if result != 0 {
					return result
				}
			}

		} else {
			// ============ UNIX / CLANG LINKING ============
			timings_start_section(timings, section_name)

			ODIN_ANDROID_API_LEVEL := build_context.ODIN_ANDROID_API_LEVEL
			ODIN_ANDROID_NDK					 := build_context.ODIN_ANDROID_NDK
			ODIN_ANDROID_NDK_TOOLCHAIN		   := build_context.ODIN_ANDROID_NDK_TOOLCHAIN
			ODIN_ANDROID_NDK_TOOLCHAIN_LIB	   := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_LIB
			ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL
			ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT   := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT

			clang_path, has_odin_clang_path_env := gb_get_env("ODIN_CLANG_PATH", allocator)
			if !has_odin_clang_path_env {
				clang_path = "clang"
			}

			lib_builder: strings.Builder
			strings.builder_init(&lib_builder, allocator)
			defer strings.builder_destroy(&lib_builder)

			asm_files: StringSet
			string_set_init(&asm_files, 64, context.temp_allocator)
			defer string_set_destroy(&asm_files)

			min_libs_set: StringSet
			string_set_init(&min_libs_set, 64, context.temp_allocator)
			defer string_set_destroy(&min_libs_set)

			prev_lib: string

			for e in gen.foreign_libraries {
				assert(e.kind == .LibraryName)

				extra_flags := string_trim_whitespace(e.LibraryName.extra_linker_flags, context.temp_allocator)
				if extra_flags != "" {
					sb_printf(&lib_builder, " %s", extra_flags)
				}

				// macOS framework preprocessing pass
				if is_darwin() {
					for lib in e.LibraryName.paths {
						lib = string_trim_whitespace(lib, context.temp_allocator)
						if lib == "" {
							continue
						}
						if strings.has_suffix(lib, ".framework") {
							if string_set_update(&min_libs_set, lib) {
								continue
							}
							lib_name := remove_extension_from_path(lib, context.temp_allocator)
							sb_printf(&lib_builder, " -framework %s ", lib_name)
						}
					}
				}

				for lib in e.LibraryName.paths {
					lib = string_trim_whitespace(lib, context.temp_allocator)
					if lib == "" {
						continue
					}

					if has_asm_extension(lib) {
						if string_set_update(&asm_files, lib) {
							continue
						}
						asm_file := lib
						obj_file: string
						temp_dir := temporary_directory(context.temp_allocator)

						if temp_dir != "" {
							filename := filename_without_directory(asm_file, context.temp_allocator)
							obj_builder: strings.Builder
							strings.builder_init(&obj_builder, allocator)
							sb_write(&obj_builder, temp_dir)
							sb_write(&obj_builder, "/")
							sb_write(&obj_builder, filename)
							sb_printf(&obj_builder, "-%p.o", raw_data(asm_file))
							obj_file = strings.to_string(obj_builder)
						} else {
							obj_file = concatenate_strings(allocator, asm_file, ".o")
						}

						obj_format: string
						if build_context.metrics.ptr_size == 8 {
							if osx {
								obj_format = NASM_MACHO64_FMT
							} else {
								obj_format = NASM_ELF64_FMT
							}
						} else {
							assert(build_context.metrics.ptr_size == 4)
							if osx {
								obj_format = NASM_MACHO32_FMT
							} else {
								obj_format = NASM_ELF32_FMT
							}
						}

						if build_context.metrics.arch == .riscv64 {
							result = system_exec_command_line_app(
								"clang",
								allocator,
								"%s \"%s\" -c -o \"%s\" -target %s -march=rv64gc %s",
								clang_path,
								asm_file,
								obj_file,
								build_context.metrics.target_triplet,
								build_context.extra_assembler_flags,
							)
						} else if osx {
							result = system_exec_command_line_app(
								"as",
								allocator,
								"as \"%s\" -o \"%s\" %s",
								asm_file,
								obj_file,
								build_context.extra_assembler_flags,
							)
						} else {
							result = system_exec_command_line_app(
								LINKER_NAME_NASM,
								allocator,
								"nasm \"%s\" -f \"%s\" -o \"%s\" %s",
								asm_file,
								obj_format,
								obj_file,
								build_context.extra_assembler_flags,
							)
							if result != 0 {
								gb_printf_err(
									"executing `nasm` to assemble foreign import of %s failed.\n\tSuggestion: `nasm` does not ship with the compiler and should be installed with your system's package manager.\n",
									asm_file,
								)
								return result
							}
						}
						append(&gen.output_object_paths, obj_file)
					} else {
						short_circuit := false
						if strings.has_suffix(lib, ".framework") {
							short_circuit = true
						} else if strings.has_suffix(lib, ".dylib") {
							short_circuit = true
						} else if strings.has_suffix(lib, ".so") {
							short_circuit = true
						} else if e.LibraryName.ignore_duplicates {
							short_circuit = true
						}

						if string_set_update(&min_libs_set, lib) && (build_context.min_link_libs || short_circuit) {
							continue
						}
						if prev_lib == lib {
							continue
						}
						prev_lib = lib

						// Skip system framework/c for macOS
						if lib == "System.framework" || lib == "System" || lib == "c" {
							continue
						}

						if is_darwin() {
							if strings.has_suffix(lib, ".framework") {
								lib_name := remove_extension_from_path(lib, context.temp_allocator)
								sb_printf(&lib_builder, " -framework %s ", lib_name)
							} else if strings.has_suffix(lib, ".a") || strings.has_suffix(lib, ".o") || strings.has_suffix(lib, ".dylib") {
								sb_printf(&lib_builder, " \"%s\" ", lib)
							} else {
								sb_printf(&lib_builder, " -l%s ", lib)
							}
						} else {
							if strings.has_suffix(lib, ".a") || strings.has_suffix(lib, ".o") || strings.has_suffix(lib, ".so") || strings.contains(lib, ".so.") {
								sb_printf(&lib_builder, " -l:\"%s\" ", lib)
							} else {
								sb_printf(&lib_builder, " -l%s ", lib)
							}
						}
					}
				}
			}

			// ---- Object files string ----
			object_files_builder: strings.Builder
			strings.builder_init(&object_files_builder, allocator)
			defer strings.builder_destroy(&object_files_builder)

			// ---- Android Native App Glue compile ----
			if is_android {
				debugf("[Section] Android Native App Glue Compile\n")
				if build_context.show_more_timings {
					timings_start_section(&global_timings, "Android Native App Glue Compile")
				}

				android_glue_object: string
				android_glue_static_lib: string

				hash_buf: [64]byte
				hash_str := fmt.aprintf("%p", &hash_buf, context.temp_allocator)

				temp_dir := normalize_path(
					context.temp_allocator,
					temporary_directory(context.temp_allocator),
					NIX_SEPARATOR,
				)

				android_glue_object = concatenate4_strings(
					context.temp_allocator,
					temp_dir,
					"android_native_app_glue-",
					hash_str,
					".o",
				)
				android_glue_static_lib = concatenate4_strings(
					allocator,
					temp_dir,
					"libandroid_native_app_glue-",
					hash_str,
					".a",
				)

				glue_builder: strings.Builder
				strings.builder_init(&glue_builder, allocator, len(ODIN_ANDROID_NDK_TOOLCHAIN))
				defer strings.builder_destroy(&glue_builder)

				sb_write(&glue_builder, ODIN_ANDROID_NDK_TOOLCHAIN)
				sb_write(&glue_builder, "bin/clang")
				sb_printf(&glue_builder, " --target=%s%d ", build_context.metrics.target_triplet, ODIN_ANDROID_API_LEVEL)
				sb_write(&glue_builder, "-c \"")
				sb_write(&glue_builder, ODIN_ANDROID_NDK)
				sb_write(&glue_builder, "sources/android/native_app_glue/android_native_app_glue.c")
				sb_write(&glue_builder, "\" ")
				sb_write(&glue_builder, "-o \"")
				sb_write(&glue_builder, android_glue_object)
				sb_write(&glue_builder, "\" ")
				sb_write(&glue_builder, "--sysroot \"")
				sb_write(&glue_builder, ODIN_ANDROID_NDK_TOOLCHAIN)
				sb_write(&glue_builder, "sysroot")
				sb_write(&glue_builder, "\" ")
				sb_write(&glue_builder, "\"-I")
				sb_write(&glue_builder, ODIN_ANDROID_NDK_TOOLCHAIN)
				sb_write(&glue_builder, "sysroot/usr/include/")
				sb_write(&glue_builder, "\" ")
				sb_write(&glue_builder, "\"-I")
				sb_write(&glue_builder, ODIN_ANDROID_NDK_TOOLCHAIN)
				sb_write(&glue_builder, "sysroot/usr/include/")
				sb_write(&glue_builder, ODIN_ANDROID_NDK_TOOLCHAIN_LIB)
				sb_write(&glue_builder, "/\" ")
				sb_write(&glue_builder, "-Wno-macro-redefined ")

				result = system_exec_command_line_app("android-native-app-glue-compile", allocator, strings.to_string(glue_builder))
				if result != 0 {
					return result
				}

				debugf("[Section] Android Native App Glue ar\n")
				if build_context.show_more_timings {
					timings_start_section(&global_timings, "Android Native App Glue ar")
				}

				ar_builder: strings.Builder
				strings.builder_init(&ar_builder, allocator, len(ODIN_ANDROID_NDK_TOOLCHAIN))
				defer strings.builder_destroy(&ar_builder)

				sb_write(&ar_builder, ODIN_ANDROID_NDK_TOOLCHAIN)
				sb_write(&ar_builder, "bin/llvm-ar")
				sb_write(&ar_builder, " rcs ")
				sb_write(&ar_builder, "\"")
				sb_write(&ar_builder, android_glue_static_lib)
				sb_write(&ar_builder, "\" ")
				sb_write(&ar_builder, "\"")
				sb_write(&ar_builder, android_glue_object)
				sb_write(&ar_builder, "\" ")

				result = system_exec_command_line_app("android-native-app-glue-ar", allocator, strings.to_string(ar_builder))
				if result != 0 {
					return result
				}

				sb_printf(&object_files_builder, "\"%s\" ", android_glue_static_lib)
			}

			for object_path in gen.output_object_paths {
				sb_printf(&object_files_builder, "\"%s\" ", object_path)
			}

			// ---- Link settings ----
			link_settings_builder: strings.Builder
			strings.builder_init(&link_settings_builder, allocator, 32)

			if build_context.no_crt {
				sb_write(&link_settings_builder, "-nostdlib ")
			}

			if build_context.build_mode == .StaticLibrary {
				debugf("[Section] Static Library Creation\n")
				if build_context.show_more_timings {
					timings_start_section(&global_timings, "Static Library Creation")
				}

				ar_command_builder: strings.Builder
				strings.builder_init(&ar_command_builder, allocator)
				defer strings.builder_destroy(&ar_command_builder)

				sb_write(&ar_command_builder, "ar rcs ")
				sb_printf(&ar_command_builder, "\"%s\" ", output_filename)
				sb_write(&ar_command_builder, strings.to_string(object_files_builder))

				result = system_exec_command_line_app(LINKER_NAME_AR, allocator, strings.to_string(ar_command_builder))
				if result != 0 {
					return result
				}
				return result
			}

			if build_context.build_mode == .DynamicLibrary {
				sb_write(&link_settings_builder, "-shared ")
				if is_darwin() {
					sb_write(&link_settings_builder, "-Wl,-init,'__odin_entry_point' ")
				} else {
					sb_write(&link_settings_builder, "-Wl,-init,'_odin_entry_point' ")
					sb_write(&link_settings_builder, "-Wl,-fini,'_odin_exit_point' ")
				}
			} else if is_android {
				sb_write(&link_settings_builder, "-shared ")
			}

			if build_context.build_mode == .Executable && build_context.reloc_mode == .PIC {
				// no -no-pie
			} else if build_context.build_mode != .DynamicLibrary {
				if build_context.metrics.os != .openbsd &&
				   build_context.metrics.os != .haiku &&
				   build_context.metrics.arch != .riscv64 &&
				   !is_android {
					sb_write(&link_settings_builder, "-no-pie ")
				}
			}

			// ---- Platform library strings ----
			platform_lib_builder: strings.Builder
			strings.builder_init(&platform_lib_builder, allocator)
			defer strings.builder_destroy(&platform_lib_builder)

			if is_darwin() {
				darwin_sdk_path_builder: strings.Builder
				strings.builder_init(&darwin_sdk_path_builder, context.temp_allocator)

				darwin_platform_name  := "MacOSX"
				darwin_xcrun_sdk_name := "macosx"
				darwin_min_version_id := "macosx"
				original_clang_path   := clang_path

				switch selected_subtarget {
				case .iPhone:
					darwin_platform_name  = "iPhoneOS"
					darwin_xcrun_sdk_name = "iphoneos"
					darwin_min_version_id = "ios"
					if !has_odin_clang_path_env {
						clang_path = "xcrun --sdk iphoneos clang"
					}
				case .iPhoneSimulator:
					darwin_platform_name  = "iPhoneSimulator"
					darwin_xcrun_sdk_name = "iphonesimulator"
					darwin_min_version_id = "ios-simulator"
					if !has_odin_clang_path_env {
						clang_path = "xcrun --sdk iphonesimulator clang"
					}
				}

				darwin_find_sdk_cmd := fmt.aprintf("xcrun --sdk %s --show-sdk-path", darwin_xcrun_sdk_name, context.temp_allocator)

				if !system_exec_command_line_app_output(darwin_find_sdk_cmd, &strings.to_string(darwin_sdk_path_builder), context.temp_allocator) {
					clang_path = original_clang_path
					strings.builder_reset(&darwin_sdk_path_builder)
					sb_printf(&darwin_sdk_path_builder, "/Library/Developer/CommandLineTools/SDKs/%s.sdk", darwin_platform_name)
					if !path_is_directory(strings.to_string(darwin_sdk_path_builder)) {
						strings.builder_reset(&darwin_sdk_path_builder)
						sb_printf(&darwin_sdk_path_builder, "/Applications/Xcode.app/Contents/Developer/Platforms/%s.platform/Developer/SDKs/%s.sdk", darwin_platform_name)
						if !path_is_directory(strings.to_string(darwin_sdk_path_builder)) {
							gb_printf_err("Failed to find %s SDK\n", darwin_platform_name)
							return -1
						}
					}
				} else {
					// trim space
					trimmed := strings.trim_space(strings.to_string(darwin_sdk_path_builder))
					strings.builder_reset(&darwin_sdk_path_builder)
					sb_write(&darwin_sdk_path_builder, trimmed)
				}

				darwin_sdk_path := strings.to_string(darwin_sdk_path_builder)
				sb_printf(&platform_lib_builder, "--sysroot %s ", darwin_sdk_path)
				sb_write(&platform_lib_builder, "-L/usr/local/lib ")

				if os.exists("/opt/homebrew/lib") {
					sb_write(&platform_lib_builder, "-L/opt/homebrew/lib ")
				}
				if os.exists("/opt/local/lib") {
					sb_write(&platform_lib_builder, "-L/opt/local/lib ")
				}
				if build_context.minimum_os_version_string_given || selected_subtarget != .Default {
					sb_printf(&link_settings_builder, "-m%s-version-min=%s ", darwin_min_version_id, build_context.minimum_os_version_string)
				}
				if build_context.build_mode != .DynamicLibrary {
					sb_write(&link_settings_builder, "-e _main ")
				}
			} else if build_context.metrics.os == .freebsd {
				if .Address in build_context.sanitizer_flags || .Memory in build_context.sanitizer_flags {
					sb_write(&platform_lib_builder, "-lpthread ")
				}
				sb_write(&platform_lib_builder, "-Wl,-L/usr/local/lib ")
			} else if build_context.metrics.os == .openbsd {
				sb_write(&platform_lib_builder, "-lpthread -Wl,-L/usr/local/lib ")
				sb_write(&platform_lib_builder, "-Wl,-z,nobtcfi ")
			}

			if is_android {
				assert(ODIN_ANDROID_NDK_TOOLCHAIN_LIB != "")
				assert(ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL != "")
				assert(ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT != "")

				sb_write(&platform_lib_builder, "\"-L")
				sb_write(&platform_lib_builder, ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT)
				sb_write(&platform_lib_builder, "usr/lib/")
				sb_write(&platform_lib_builder, ODIN_ANDROID_NDK_TOOLCHAIN_LIB)
				sb_printf(&platform_lib_builder, "/%d", ODIN_ANDROID_API_LEVEL)
				sb_write(&platform_lib_builder, "\" ")
				sb_write(&platform_lib_builder, "-landroid ")
				sb_write(&platform_lib_builder, "-llog ")
				sb_write(&platform_lib_builder, "\"--sysroot=")
				sb_write(&platform_lib_builder, ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT)
				sb_write(&platform_lib_builder, "\" ")
				sb_write(&link_settings_builder, "-u ANativeActivity_onCreate ")
			}

			if !build_context.no_rpath {
				if is_darwin() {
					sb_write(&link_settings_builder, "-Wl,-rpath,@loader_path ")
				} else {
					if !is_android {
						sb_write(&link_settings_builder, "-Wl,-rpath,\\$ORIGIN ")
					}
				}
			}

			lib_str_final := strings.to_string(lib_builder)
			if !build_context.no_crt {
				sb_write(&lib_builder, "-lm ")
				if !is_darwin() {
					sb_write(&lib_builder, "-lc ")
				}
				lib_str_final = strings.to_string(lib_builder)
			}

			// ---- Build link command line ----
			link_cmd_builder: strings.Builder
			strings.builder_init(&link_cmd_builder, allocator)
			defer strings.builder_destroy(&link_cmd_builder)

			if is_android {
				ndk_bin_dir := strings.concatenate({ODIN_ANDROID_NDK_TOOLCHAIN, "bin/clang"}, context.temp_allocator)
				sb_write(&link_cmd_builder, ndk_bin_dir)
				sb_printf(&link_cmd_builder, " --target=%s%d ", build_context.metrics.target_triplet, ODIN_ANDROID_API_LEVEL)
			} else {
				sb_write(&link_cmd_builder, clang_path)
			}

			sb_write(&link_cmd_builder, " -Wno-unused-command-line-argument ")

			if build_context.lto_kind != .None {
				sb_write(&link_cmd_builder, " -flto=thin")
				sb_printf(&link_cmd_builder, " -flto-jobs=%d ", build_context.thread_count)
				if build_context.ODIN_DEBUG {
					sb_write(&link_cmd_builder, " -g ")
				}
				if osx && !build_context.minimum_os_version_string_given {
					sb_write(&link_cmd_builder, " -Wno-override-module ")
				}
			}

			sb_write(&link_cmd_builder, strings.to_string(object_files_builder))
			sb_printf(&link_cmd_builder, " -o \"%s\" ", output_filename)
			sb_printf(&link_cmd_builder, " %s ", strings.to_string(platform_lib_builder))
			sb_printf(&link_cmd_builder, " %s ", lib_str_final)
			sb_printf(&link_cmd_builder, " %s ", build_context.link_flags)
			sb_printf(&link_cmd_builder, " %s ", build_context.extra_linker_flags)
			sb_printf(&link_cmd_builder, " %s ", strings.to_string(link_settings_builder))

			if is_android {
				debugf("[Section] Linking\n")
				if build_context.show_more_timings {
					timings_start_section(&global_timings, "Linking")
				}
			}

			if build_context.linker_choice == .lld {
				sb_write(&link_cmd_builder, " -fuse-ld=lld")
				result = system_exec_command_line_app(LINKER_NAME_LLD, allocator, strings.to_string(link_cmd_builder))
			} else if build_context.linker_choice == .mold {
				sb_write(&link_cmd_builder, " -fuse-ld=mold")
				result = system_exec_command_line_app("mold-link", allocator, strings.to_string(link_cmd_builder))
			} else {
				result = system_exec_command_line_app("ld-link", allocator, strings.to_string(link_cmd_builder))
			}

			if result != 0 {
				return result
			}

			// ---- dsymutil on macOS debug ----
			if osx && build_context.ODIN_DEBUG {
				result = system_exec_command_line_app(
					LINKER_NAME_DSYM,
					allocator,
					"dsymutil \"%s\"",
					output_filename,
				)
				if result != 0 {
					return result
				}
			}
		}
	}

	return result
}