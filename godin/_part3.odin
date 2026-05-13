// Part 3: timings_export_all, check_defines, defineables, import_graph, show_timings,
// export_dependencies, export_linked_libraries

import "core:fmt"
import "core:os"
import "core:strings"
import "core:slice"
import "core:mem"
import "core:time"
import "core:path/filepath"

// =============================================================================
// Forward-declared external types referenced throughout this part
// =============================================================================

AstPackage :: struct {}
AstFile :: struct {}
Entity :: struct {}
EntityKind :: enum int {}
TimeStamp :: struct {}
Timings :: struct {}
Checker :: struct {}
Parser :: struct {}
ParserStats :: struct {}
Defineable :: struct {}
Token :: struct {}
TokenPos :: struct {}
LoadFileCache :: struct {}
lbGenerator :: struct {}
LinkerData :: struct {}
AstForeignImportDecl :: struct {}
LibraryCollections :: struct {}

// =============================================================================
// timings_export_all (lines 1662-1723)
// =============================================================================

timings_export_all :: proc(t: ^Timings, filename: string) {
	fd, open_err := os.open(filename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)
	if open_err != os.ERROR_NONE {
		fmt.eprintf("Failed to open timing export file: %s\n", filename)
		return
	}
	defer os.close(fd)

	is_json := strings.has_suffix(filename, ".json")

	if is_json {
		os.write(fd, transmute([]u8)"{\n")
		os.write(fd, transmute([]u8)"  \"sections\": [\n")
	}

	// Placeholder for timing section iteration.
	// Original code iterates t.sections and writes formatted timing data.
	// Each section has: name, total_time, count, bytes_processed, etc.
	// TODO: iterate t.sections when Timings struct is fully defined

	if is_json {
		os.write(fd, transmute([]u8)"  ]\n")
		os.write(fd, transmute([]u8)"}\n")
	} else {
		// CSV header
		fmt.fprintf(fd, "section,total_ms,count,bytes_per_sec\n")
		// TODO: write CSV rows from t.sections
	}
}

// =============================================================================
// check_defines (lines 1724-1745)
// =============================================================================

check_defines :: proc(c: ^Checker) {
	warned := false

	// Iterate defined values map and check against info.defineables
	// Original code:
	//   for each entry in bc.defined_values:
	//     found = false
	//     for each d in c.info.defineables:
	//       if d.used: found = true; break
	//     if !found: warn about unused define

	// TODO: iterate c.defined_values when Checker struct is fully defined
	// TODO: iterate c.info.defineables

	_ = warned
}

// =============================================================================
// temp_alloc_defineable_strings (lines 1746-1752)
// =============================================================================

temp_alloc_defineable_strings :: proc(c: ^Checker) {
	// Caches string representations of defineable values into a temporary allocator.
	// Original code:
	//   for each d in c.info.defineables:
	//     d.cached_string = write_value_to_string(...)

	// TODO: iterate c.info.defineables and cache string representations
}

// =============================================================================
// defineables_cmp (lines 1753-1764)
// =============================================================================

@(private="file")
defineables_cmp :: proc(a, b: ^Defineable) -> int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Compare by position (token_pos_cmp logic)
	// Original compares a.pos.file_id, a.pos.offset, a.pos.line, a.pos.column
	// TODO: implement token_pos_cmp(a.pos, b.pos) when TokenPos is defined
	return 0
}

// =============================================================================
// sort_defineables_and_remove_duplicates (lines 1765-1780)
// =============================================================================

sort_defineables_and_remove_duplicates :: proc(defineables: ^[dynamic]Defineable) {
	if len(defineables) <= 1 {
		return
	}

	// Sort by position
	slice.sort_by(defineables[:], proc(a, b: Defineable) -> bool {
		return defineables_cmp(&a, &b) < 0
	})

	// Remove duplicates (same position)
	n := 1
	for i in 1 ..< len(defineables) {
		if defineables_cmp(&defineables[i], &defineables[n-1]) != 0 {
			defineables[n] = defineables[i]
			n += 1
		}
	}

	// Resize to remove duplicates
	resize(&defineables, n)
}

// =============================================================================
// export_defineables (lines 1781-1814)
// =============================================================================

export_defineables :: proc(c: ^Checker, filename: string) {
	fd, open_err := os.open(filename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)
	if open_err != os.ERROR_NONE {
		fmt.eprintf("Failed to open defineables export file: %s\n", filename)
		return
	}
	defer os.close(fd)

	temp_alloc_defineable_strings(c)

	// CSV header
	os.write(fd, transmute([]u8)"name,value,file,line,column,used\n")

	// TODO: iterate c.info.defineables
	// For each defineable d:
	//   pos_str := token_pos_to_string(d.pos)  // TODO stub
	//   value_str := d.cached_string
	//   fmt.fprintf(fd, "%s,%s,%s,%t\n", d.name, value_str, pos_str, d.used)
}

// =============================================================================
// show_defineables (lines 1815-1832)
// =============================================================================

show_defineables :: proc(c: ^Checker) {
	temp_alloc_defineable_strings(c)

	has_color := has_ansi_terminal_colours()

	// TODO: iterate c.info.defineables
	// For each defineable d:
	//   pos_str := token_pos_to_string(d.pos)  // TODO stub
	//   value_str := d.cached_string
	//   if has_color: use ANSI color codes for formatting
	//   fmt.printf("  %-40s = %-40s // %s\n", d.name, value_str, pos_str)

	_ = has_color
}

// =============================================================================
// has_ansi_terminal_colours (stub - referenced by show_defineables)
// =============================================================================

has_ansi_terminal_colours :: proc() -> bool {
	// Check NO_COLOR env var and whether stderr is a terminal
	// Original checks gb_is_terminal and NO_COLOR environment variable
	// TODO: implement proper terminal/color detection
	term := os.get_env("TERM")
	no_color := os.get_env("NO_COLOR")
	if no_color != "" {
		return false
	}
	if term == "" {
		return false
	}
	return true
}

// =============================================================================
// show_import_graph (lines 1833-1889)
// =============================================================================

show_import_graph :: proc(gen: ^lbGenerator, c: ^Checker, collections: ^LibraryCollections) {
	fd, open_err := os.open("import_graph.dot", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)
	if open_err != os.ERROR_NONE {
		fmt.eprintf("Failed to open import_graph.dot\n")
		return
	}
	defer os.close(fd)

	fmt.fprintf(fd, "digraph ImportGraph {\n")
	fmt.fprintf(fd, "  rankdir=LR;\n")
	fmt.fprintf(fd, "  node [shape=box, style=filled, fillcolor=lightyellow];\n")

	// TODO: iterate p.packages (gen.info.packages)
	// For each package pkg:
	//   pkg_name := pkg.name (with collection prefix)
	//   fmt.fprintf(fd, "  \"%s\";\n", pkg_name)
	//   For each import imp in pkg.imports:
	//     imp_name := imp.name
	//     fmt.fprintf(fd, "  \"%s\" -> \"%s\";\n", pkg_name, imp_name)

	fmt.fprintf(fd, "}\n")

	fmt.printf("Import graph written to import_graph.dot\n")

	_ = gen
	_ = c
	_ = collections
}

// =============================================================================
// show_timings (lines 1890-2006)
// =============================================================================

show_timings :: proc(t: ^Timings, parser: ^Parser, linker_data: ^LinkerData) {
	if t == nil {
		return
	}

	// Placeholder: original has detailed sections for tokenization, parsing, checker, total.
	// It uses multiple time sections from t.sections, parser stats, linker timing info.

	has_color := has_ansi_terminal_colours()

	fmt.printf("\n=== Timings ===\n\n")

	// Section format (original does per-section timing output):
	// Name | Duration | Bytes | Count | Throughput
	// e.g. "Tokenizer: 1.234s (100000 loc/s, 500000 tokens/s, 50 MB/s)"

	// TODO: iterate t.sections to print each timing section
	// Each section contains: name, total_time, total_bytes, total_count
	// Compute throughput: loc/s, tokens/s, bytes/s

	// Parser stats placeholder
	// TODO: print parser.total_lines, parser.total_tokens when ParserStats is defined
	_ = parser
	_ = linker_data
	_ = has_color

	fmt.printf("\n")
}

// =============================================================================
// file_path_cmp (lines 2007-2011)
// =============================================================================

file_path_cmp :: proc(a, b: ^AstFile) -> int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Compare by fullpath
	// TODO: a.fullpath and b.fullpath are string fields on AstFile
	return strings.compare("", "") // Placeholder
}

// =============================================================================
// export_dependencies (lines 2012-2093)
// =============================================================================

export_dependencies :: proc(gen: ^lbGenerator, c: ^Checker, filename: string) {
	fd, open_err := os.open(filename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)
	if open_err != os.ERROR_NONE {
		fmt.eprintf("Failed to open dependency export file: %s\n", filename)
		return
	}
	defer os.close(fd)

	is_make := strings.has_suffix(filename, ".d") || strings.has_suffix(filename, ".make")
	is_json := strings.has_suffix(filename, ".json")

	if is_make {
		// Makefile dependency format
		// Iterate gen.info.packages and c.info.load_file_cache
		// Output: target: dep1 dep2 dep3 ...
		// TODO: iterate p.packages
		// TODO: iterate c.info.load_file_cache entries
		fmt.fprintf(fd, "# Dependency export (Make format)\n")
		fmt.fprintf(fd, "# TODO: full implementation\n")
	} else if is_json {
		// JSON dependency format
		fmt.fprintf(fd, "{\n")
		fmt.fprintf(fd, "  \"dependencies\": [\n")
		// TODO: iterate packages and their imports, write JSON entries
		fmt.fprintf(fd, "  ]\n")
		fmt.fprintf(fd, "}\n")
	} else {
		fmt.eprintf("Unknown dependency export format. Use .d, .make, or .json extension.\n")
	}

	_ = gen
	_ = c
}

// =============================================================================
// export_linked_libraries (lines 2094-2139)
// =============================================================================

export_linked_libraries :: proc(gen: ^lbGenerator, c: ^Checker, filename: string) {
	fd, open_err := os.open(filename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0o644)
	if open_err != os.ERROR_NONE {
		fmt.eprintf("Failed to open linked libraries export file: %s\n", filename)
		return
	}
	defer os.close(fd)

	is_json := strings.has_suffix(filename, ".json")

	if is_json {
		fmt.fprintf(fd, "{\n")
		fmt.fprintf(fd, "  \"libraries\": [\n")
	} else {
		fmt.fprintf(fd, "# Linked Libraries\n")
		fmt.fprintf(fd, "# Format: type | scope | name\n")
		fmt.fprintf(fd, "# type: static or dynamic\n")
		fmt.fprintf(fd, "# scope: system or user\n\n")
	}

	// TODO: iterate gen.foreign_libraries
	// Original code iterates gen.foreign_libraries which contains AstForeignImportDecl entries
	// For each lib:
	//   lib_type := lib.is_dynamic ? "dynamic" : "static"
	//   lib_scope := lib.is_system ? "system" : "user"
	//   lib_name := lib.name
	//   if is_json:
	//     fmt.fprintf(fd, "    {\"name\": \"%s\", \"type\": \"%s\", \"scope\": \"%s\"},\n", lib_name, lib_type, lib_scope)
	//   else:
	//     fmt.fprintf(fd, "%s | %s | %s\n", lib_type, lib_scope, lib_name)

	if is_json {
		fmt.fprintf(fd, "  ]\n")
		fmt.fprintf(fd, "}\n")
	}

	_ = gen
	_ = c
}
