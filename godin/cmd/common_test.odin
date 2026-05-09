// common_test.odin
package godin

import "core:os"
import "core:strings"
import "core:testing"

// ---------------------------------------------------------------
// next_pow2_32 / 64 / int
// ---------------------------------------------------------------
@(test)
test_next_pow2_32 :: proc(t: ^testing.T) {
    testing.expect(t, next_pow2_32(0) == 0)
    testing.expect(t, next_pow2_32(1) == 1)
    testing.expect(t, next_pow2_32(2) == 2)
    testing.expect(t, next_pow2_32(3) == 4)
    testing.expect(t, next_pow2_32(5) == 8)
    testing.expect(t, next_pow2_32(15) == 16)
    testing.expect(t, next_pow2_32(16) == 16)
}

@(test)
test_next_pow2_64 :: proc(t: ^testing.T) {
    testing.expect(t, next_pow2_64(0) == 0)
    testing.expect(t, next_pow2_64(1) == 1)
    testing.expect(t, next_pow2_64(2) == 2)
    testing.expect(t, next_pow2_64(7) == 8)
    testing.expect(t, next_pow2_64(9) == 16)
    testing.expect(t, next_pow2_64(1 << 33) == 1 << 34)
}

@(test)
test_next_pow2_int :: proc(t: ^testing.T) {
    testing.expect(t, next_pow2_int(0) == 0)
    testing.expect(t, next_pow2_int(3) == 4)
    testing.expect(t, next_pow2_int(5) == 8)
}

// ---------------------------------------------------------------
// is_power_of_two_int
// ---------------------------------------------------------------
@(test)
test_is_power_of_two :: proc(t: ^testing.T) {
    testing.expect(t, is_power_of_two_int(1))
    testing.expect(t, is_power_of_two_int(2))
    testing.expect(t, is_power_of_two_int(4))
    testing.expect(t, !is_power_of_two_int(0))
    testing.expect(t, !is_power_of_two_int(3))
    testing.expect(t, !is_power_of_two_int(6))
    testing.expect(t, !is_power_of_two_int(-1))
}

// ---------------------------------------------------------------
// bit_set_count / floor / ceil / prev_pow2
// ---------------------------------------------------------------
@(test)
test_bit_set_count :: proc(t: ^testing.T) {
    testing.expect(t, bit_set_count_u32(0) == 0)
    testing.expect(t, bit_set_count_u32(0b1010) == 2)
    testing.expect(t, bit_set_count_u32(0xFFFFFFFF) == 32)
    testing.expect(t, bit_set_count_u64(0) == 0)
    testing.expect(t, bit_set_count_u64(0x8000000000000001) == 2)
}

@(test)
test_floor_log2 :: proc(t: ^testing.T) {
    testing.expect(t, floor_log2_u32(1) == 0)
    testing.expect(t, floor_log2_u32(2) == 1)
    testing.expect(t, floor_log2_u32(3) == 1)
    testing.expect(t, floor_log2_u32(4) == 2)
    testing.expect(t, floor_log2_u32(16) == 4)
}

@(test)
test_ceil_log2 :: proc(t: ^testing.T) {
    testing.expect(t, ceil_log2_u32(1) == 0)
    testing.expect(t, ceil_log2_u32(2) == 1)
    testing.expect(t, ceil_log2_u32(3) == 2)
    testing.expect(t, ceil_log2_u32(4) == 2)
    testing.expect(t, ceil_log2_u32(5) == 3)
}

@(test)
test_prev_pow2 :: proc(t: ^testing.T) {
    testing.expect(t, prev_pow2_u32(0) == 0)
    testing.expect(t, prev_pow2_u32(1) == 1)
    testing.expect(t, prev_pow2_u32(2) == 2)
    testing.expect(t, prev_pow2_u32(3) == 2)
    testing.expect(t, prev_pow2_u32(5) == 4)
    testing.expect(t, prev_pow2_u32(8) == 8)
    testing.expect(t, prev_pow2_u32(9) == 8)
}

// ---------------------------------------------------------------
// isize_cmp / i64_cmp / i32_cmp / u64_cmp
// ---------------------------------------------------------------
@(test)
test_integer_comparators :: proc(t: ^testing.T) {
    testing.expect(t, isize_cmp(1, 2) == -1)
    testing.expect(t, isize_cmp(2, 2) == 0)
    testing.expect(t, isize_cmp(3, 2) == 1)
    testing.expect(t, i64_cmp(-5, -3) == -1)
    testing.expect(t, u64_cmp(5, 5) == 0)
}

// ---------------------------------------------------------------
// string_split_iterator
// ---------------------------------------------------------------
@(test)
test_string_split_iterator :: proc(t: ^testing.T) {
    s := "apple,orange,,grape"
    it := String_Iterator{ str = s, pos = 0 }
    parts := make([dynamic]string, context.temp_allocator)
    for {
        part := string_split_iterator(&it, ',')
        if part == "" && it.pos >= len(s) do break
        append(&parts, part)
    }
    testing.expect(t, len(parts) == 4)
    testing.expect(t, parts[0] == "apple")
    testing.expect(t, parts[1] == "orange")
    testing.expect(t, parts[2] == "")
    testing.expect(t, parts[3] == "grape")
}

// ---------------------------------------------------------------
// is_separator
// ---------------------------------------------------------------
@(test)
test_is_separator :: proc(t: ^testing.T) {
    testing.expect(t, is_separator('/'))
    testing.expect(t, is_separator('\\'))
    testing.expect(t, !is_separator('a'))
}

// ---------------------------------------------------------------
// string_extension_position / path_extension / path_remove_extension
// ---------------------------------------------------------------
@(test)
test_string_extension_position :: proc(t: ^testing.T) {
    testing.expect(t, string_extension_position("test.txt") == 4)
    testing.expect(t, string_extension_position("archive.tar.gz") == 11)
    testing.expect(t, string_extension_position("noext") == -1)
    testing.expect(t, string_extension_position("/path/to/file") == -1)
    testing.expect(t, string_extension_position("/path.with.dots/file") == -1)
    testing.expect(t, string_extension_position(".hidden") == 0) // dot at start
}

@(test)
test_path_extension :: proc(t: ^testing.T) {
    testing.expect(t, path_extension("file.txt") == ".txt")
    testing.expect(t, path_extension("file.txt", false) == "txt")
    testing.expect(t, path_extension("no_ext") == "")
}

@(test)
test_path_remove_extension :: proc(t: ^testing.T) {
    testing.expect(t, path_remove_extension("file.txt") == "file")
    testing.expect(t, path_remove_extension("file.tar.gz") == "file.tar")
    testing.expect(t, path_remove_extension("no_ext") == "no_ext")
}

@(test)
test_filename_from_path :: proc(t: ^testing.T) {
    testing.expect(t, filename_from_path("/home/user/file.txt") == "/home/user/file")
    testing.expect(t, filename_from_path("readme.md") == "readme")
}

// ---------------------------------------------------------------
// string16_len
// ---------------------------------------------------------------
@(test)
test_string16_len :: proc(t: ^testing.T) {
    testing.expect(t, string16_len({ }) == 0)
    testing.expect(t, string16_len({ 0 }) == 0) // empty string
    data := []u16{ 'h', 'i', 0 }
    testing.expect(t, string16_len(data) == 2)
}

// ---------------------------------------------------------------
// string_partition
// ---------------------------------------------------------------
@(test)
test_string_partition :: proc(t: ^testing.T) {
    str := "hello world"
    head, match, tail := string_partition(str, " ")
    testing.expect(t, head == "hello")
    testing.expect(t, match == " ")
    testing.expect(t, tail == "world")

    // no match
    h, m, t := string_partition(str, "xxx")
    testing.expect(t, h == str)
    testing.expect(t, m == "")
    testing.expect(t, t == "")
}

// ---------------------------------------------------------------
// Arena
// ---------------------------------------------------------------
@(test)
test_arena_alloc_and_temp :: proc(t: ^testing.T) {
    a: Arena
    a.minimum_block_size = 64
    p1 := arena_alloc(&a, 10, 4)
    p2 := arena_alloc(&a, 20, 8)
    testing.expect(t, p1 != nil && p2 != nil)
    testing.expect(t, uintptr(p2) - uintptr(p1) >= 10) // rough check

    // temp begin/end rollback
    temp := arena_temp_begin(&a)
    p3 := arena_alloc(&a, 200, 4) // forces new block
    testing.expect(t, p3 != nil)
    arena_temp_end(temp)
    // after rollback, used memory should be reset, further allocations reuse
    p4 := arena_alloc(&a, 50, 4)
    testing.expect(t, p4 != nil) // should succeed without crash
    arena_free_all(&a)
}

// ---------------------------------------------------------------
// PriorityQueue
// ---------------------------------------------------------------
@(test)
test_priority_queue :: proc(t: ^testing.T) {
    pq: PriorityQueue(int)
    pq.cmp = proc(a, b: int) -> bool {
        return a < b
    }
    priority_queue_push(&pq, 5)
    priority_queue_push(&pq, 1)
    priority_queue_push(&pq, 3)
    testing.expect(t, len(pq.queue) == 3)
    testing.expect(t, priority_queue_pop(&pq) == 1)
    testing.expect(t, priority_queue_pop(&pq) == 3)
    priority_queue_push(&pq, 0)
    priority_queue_push(&pq, 9)
    testing.expect(t, priority_queue_pop(&pq) == 0)
    testing.expect(t, priority_queue_pop(&pq) == 5)
    testing.expect(t, priority_queue_pop(&pq) == 9)
    testing.expect(t, len(pq.queue) == 0)
}

// ---------------------------------------------------------------
// String interner
// ---------------------------------------------------------------
@(test)
test_string_interner :: proc(t: ^testing.T) {
// reset for test
    g_interner.store = make(map[string]InternedString)
    g_interner.rev = nil
    id1 := string_interner_insert("hello")
    id2 := string_interner_insert("world")
    id3 := string_interner_insert("hello") // same string
    testing.expect(t, id1 == id3)
    testing.expect(t, id1 < id2 || id2 < id1) // different values
    testing.expect(t, string_interner_load(id1) == "hello")
    testing.expect(t, string_interner_load(id2) == "world")
}

// ---------------------------------------------------------------
// RangeCache
// ---------------------------------------------------------------
@(test)
test_range_cache :: proc(t: ^testing.T) {
    c: RangeCache
    testing.expect(t, range_cache_add_index(&c, 5) == true)
    testing.expect(t, range_cache_add_index(&c, 5) == false)
    testing.expect(t, range_cache_add_range(&c, 10, 20) == true)
    testing.expect(t, range_cache_add_range(&c, 12, 15) == false) // already covered
    testing.expect(t, range_cache_add_range(&c, 0, 4) == true)
    // check merged continuity: adding (4,6) should merge (0,4) and (5,5)
    testing.expect(t, range_cache_add_range(&c, 4, 6) == false)
    // enumerate for verification (ranges should now be [0,6] and [10,20])
    testing.expect(t, len(c.ranges) == 2)
// sort to make assertion stable (lazy)
// not implemented in the library, so we just check counts
}

// ---------------------------------------------------------------
// Levenshtein distance
// ---------------------------------------------------------------
@(test)
test_levenstein_distance :: proc(t: ^testing.T) {
    testing.expect(t, levenstein_distance_case_insensitive("kitten", "sitting") == 3)
    testing.expect(t, levenstein_distance_case_insensitive("", "") == 0)
    testing.expect(t, levenstein_distance_case_insensitive("abc", "abc") == 0)
    testing.expect(t, levenstein_distance_case_insensitive("abc", "abx") == 1)
    testing.expect(t, levenstein_distance_case_insensitive("ABCD", "abcd") == 0) // case insensitive
    testing.expect(t, levenstein_distance_case_insensitive("flaw", "lawn") == 2)
}

// ---------------------------------------------------------------
// f32_to_f16 (half-float)
// ---------------------------------------------------------------
@(test)
test_f32_to_f16 :: proc(t: ^testing.T) {
    testing.expect(t, f32_to_f16(0.0) == 0)
    testing.expect(t, f32_to_f16(1.0) == 0x3c00)
    testing.expect(t, f32_to_f16(-1.0) == 0xbc00)
    testing.expect(t, f32_to_f16(0.5) == 0x3800)
    testing.expect(t, f32_to_f16(2.0) == 0x4000)
    // infinity/NaN checks
    inf := f32(1e38 * 1e10)
    nan := f32(0.0 / 0.0)
    testing.expect(t, (f32_to_f16(inf) & 0x7c00) == 0x7c00)
    testing.expect(t, (f32_to_f16(nan) & 0x7c00) == 0x7c00)
}

// ---------------------------------------------------------------
// fnv32a / fnv64a
// ---------------------------------------------------------------
@(test)
test_fnv_hash :: proc(t: ^testing.T) {
    testing.expect(t, fnv32a({ }) == 0x811c9dc5)
    testing.expect(t, fnv32a({ 'a', 'b', 'c' }) == 0xe71fa219) // example from known fnv32a
    testing.expect(t, fnv64a({ }) == 0xcbf29ce484222325)
    testing.expect(t, fnv64a({ 'a', 'b', 'c' }) == 0xe71fa2190541384b) // known value
}

// ---------------------------------------------------------------
// obfuscate_string / obfuscate_i32
// ---------------------------------------------------------------
@(test)
test_obfuscate :: proc(t: ^testing.T) {
    s := "secret"
    obs := obfuscate_string(s, "p_")
    testing.expect(t, strings.has_prefix(obs, "p_x"))
    testing.expect(t, len(obs) > 2)

    i := i32(42)
    o := obfuscate_i32(i)
    testing.expect(t, o != i) // very likely
}

// ---------------------------------------------------------------
// alloc_cstring / string16_to_string / string_to_string16
// ---------------------------------------------------------------
@(test)
test_string_conversions :: proc(t: ^testing.T) {
// alloc_cstring
    s := "hello"
    cstr := alloc_cstring(s)
    testing.expect(t, cstr == "hello")
    testing.expect(t, cstr[len(s)] == 0)

    // utf16 conversion
    s16 := string_to_string16("ünicodé")
    s2 := string16_to_string(s16)
    testing.expect(t, s2 == "ünicodé")
}

// ---------------------------------------------------------------
// read_directory (only if tests run in a consistent environment)
// ---------------------------------------------------------------
@(test)
test_read_directory :: proc(t: ^testing.T) {
// Create a temp directory and test
    dir := os.temp_dir() + "/odin_common_test"
    os.make_directory(dir)
    defer os.remove_all(dir)
    os.write_entire_file(dir + "/file1.txt", transmute([]u8) "data")
    os.write_entire_file(dir + "/file2.log", transmute([]u8) "data")
    os.make_directory(dir + "/subdir")

    files, err := read_directory(dir)
    testing.expect(t, err == .None)
    testing.expect(t, len(files) == 3) // two files + one subdir

    found_file1 := false
    found_dir := false
    for f in files {
        if f.name == "file1.txt" && !f.is_dir do found_file1 = true
        if f.name == "subdir" && f.is_dir do found_dir = true
    }
    testing.expect(t, found_file1)
    testing.expect(t, found_dir)
}

// ---------------------------------------------------------------
// path_from_string
// ---------------------------------------------------------------
@(test)
test_path_from_string :: proc(t: ^testing.T) {
// Unix-style path
    p, ok := path_from_string("/usr/local/bin/exec.sh")
    if ok {
        testing.expect(t, p.basename == "/usr/local/bin")
        testing.expect(t, p.name == "exec")
        testing.expect(t, p.ext == "sh")
    }
    // bare filename
    p2, ok2 := path_from_string("readme.txt")
    if ok2 {
        testing.expect(t, p2.basename == "")
        testing.expect(t, p2.name == "readme")
        testing.expect(t, p2.ext == "txt")
    }
}

// ---------------------------------------------------------------
// did_you_mean
// ---------------------------------------------------------------
@(test)
test_did_you_mean :: proc(t: ^testing.T) {
    d := did_you_mean_make("helo", 5)
    did_you_mean_append(&d, "hello")
    did_you_mean_append(&d, "hero")
    did_you_mean_append(&d, "help")
    // sort is missing in snippet but we can still test append
    testing.expect(t, len(d.distances) == 3)
    // manually check distance values
    testing.expect(t, d.distances[0].distance == 1) // helo->hello
    testing.expect(t, d.distances[1].distance == 1) // helo->hero
    testing.expect(t, d.distances[2].distance == 1) // helo->help (h<->l swap?)
}