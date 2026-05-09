// gb_test.odin - Tests for the godin package
package godin

import "core:mem"
import "core:testing"

// Test helper
expectf :: proc(t: ^testing.T, condition: bool, msg: string, args: ..any, loc := #caller_location) {
    testing.expectf(t, condition, msg, ..args, loc = loc)
}

@(test)
test_is_power_of_two :: proc(t: ^testing.T) {
    testing.expect(t, is_power_of_two(1))
    testing.expect(t, is_power_of_two(2))
    testing.expect(t, is_power_of_two(4))
    testing.expect(t, is_power_of_two(1024))
    testing.expect(t, !is_power_of_two(0))
    testing.expect(t, !is_power_of_two(3))
    testing.expect(t, !is_power_of_two(-1))
}

@(test)
test_pointer_arithmetic :: proc(t: ^testing.T) {
    arr : [4]u32 = { 10, 20, 30, 40 }
    ptr := rawptr(&arr[0])
    ptr2 := pointer_add(ptr, size_of(u32))
    testing.expect(t, (^u32)(ptr2)^ == 20)
    ptr3 := pointer_sub(ptr2, size_of(u32))
    testing.expect(t, ptr3 == ptr)
    diff := pointer_diff(ptr, pointer_add(ptr, 3 * size_of(u32)))
    testing.expect(t, diff == 3 * isize(size_of(u32)))

    aligned := align_forward(ptr, 4)
    testing.expect(t, uintptr(aligned) % 4 == 0)
}

@(test)
test_memory_ops :: proc(t: ^testing.T) {
// memcopy
    a : [10]u8 = { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 }
    b: [10]u8
    memcopy(&b[0], &a[0], 10)
    testing.expect(t, mem.compare(&a[0], &b[0], 10) == 0)

    // memset
    memset(&b[0], 0xAB, 5)
    testing.expect(t, b[0] == 0xAB && b[4] == 0xAB)
    testing.expect(t, b[5] == 5) // untouched

    // memmove (overlapping)
    memmove(&b[1], &b[0], 4)
    testing.expect(t, b[1] == 0xAB && b[4] == 0xAB)

    // memcompare
    c := [4]u8  {
        1, 2, 3, 4 }
    d := [4]u8  {
        1, 2, 3, 4 }
    testing.expect(t, memcompare(&c, &d, 4) == 0)
    d[2] = 99
    testing.expect(t, memcompare(&c, &d, 4) != 0)

    // memswap
    swap_a := u32(0x1234)
    swap_b := u32(0x5678)
    memswap(&swap_a, &swap_b, 4)
    testing.expect(t, swap_a == 0x5678 && swap_b == 0x1234)

    // memchr
    e := [6]u8  {
        1, 2, 3, 0, 5, 6 }
    testing.expect(t, memchr(&e[0], 0, 6) == &e[3])
    testing.expect(t, memchr(&e[0], 7, 6) == nil)
    testing.expect(t, memrchr(&e[0], 6, 6) == &e[5])
}

@(test)
test_thread_id :: proc(t: ^testing.T) {
    id := thread_current_id()
    testing.expect(t, id > 0, "thread id should be positive")
}

@(test)
test_affinity :: proc(t: ^testing.T) {
    a: gbAffinity
    affinity_init(&a)
    testing.expect(t, a.core_count == 1 && a.thread_count == 1)
    testing.expect(t, affinity_set(&a, 0, 0) == false)
    testing.expect(t, affinity_thread_count_for_core(&a, 0) == 1)
    affinity_destroy(&a) // no-op, just coverage
}

@(test)
test_allocator :: proc(t: ^testing.T) {
    alloc := heap_allocator()
    // alloc and zero
    mem := alloc_align(alloc, 16, 8)
    defer free(alloc, mem)
    testing.expect(t, mem != nil)
    bytes := ([^]u8)(mem)
    testing.expect(t, bytes[0] == 0 && bytes[15] == 0)

    // alloc_copy
    src := "hello"
    copied := alloc_str(alloc, src)
    defer delete(copied)
    testing.expect(t, copied == src)

    // resize
    p := alloc(alloc, 4)
    defer free(alloc, p)
    (^u32)(p)^ = 0xDEAD
    p2 := resize(alloc, p, 4, 8)
    testing.expect(t, (^u32)(p2)^ == 0xDEAD)
    free(alloc, p2)
}

@(test)
test_comparators :: proc(t: ^testing.T) {
    arr := [5]i32{ 3, 1, 4, 1, 5 }
    sort(&arr[0], 5, size_of(i32), gb_i32_cmp(0))
    testing.expect(t, arr == [5]i32{ 1, 1, 3, 4, 5 })

    // radix sort u8
    u8_arr := []u8{ 5, 2, 7, 1, 9 }
    u8_tmp := make([]u8, len(u8_arr), context.temp_allocator)
    radix_sort_u8(u8_arr, u8_tmp)
    testing.expect(t, u8_arr[0] == 1 && u8_arr[4] == 9)

    // binary search
    pos := binary_search(&arr[0], 5, size_of(i32), &arr[2], gb_i32_cmp(0))
    testing.expect(t, pos == 2)
}

@(test)
test_char_functions :: proc(t: ^testing.T) {
    testing.expect(t, char_is_digit('5'))
    testing.expect(t, !char_is_digit('x'))
    testing.expect(t, hex_digit_to_int('F') == 15)
    testing.expect(t, char_to_lower('Q') == 'q')
    testing.expect(t, char_to_upper('m') == 'M')
}

@(test)
test_string_basics :: proc(t: ^testing.T) {
    s := "Hello, World!"
    testing.expect(t, strlen(s) == 13)
    testing.expect(t, strnlen(s, 5) == 5)
    testing.expect(t, strcmp("abc", "abc") == 0)
    testing.expect(t, strncmp("abcd", "abce", 3) == 0)
    testing.expect(t, str_has_prefix("foobar", "foo"))
    testing.expect(t, str_has_suffix("foobar", "bar"))
    testing.expect(t, char_first_occurence(s, 'W') == 7)
    testing.expect(t, char_last_occurence(s, 'l') == 10)

    rev := strrev("dog")
    testing.expect(t, rev == "god")
    delete(rev) // if allocated by strrev

    token: string
    rest := strtok(&token, "foo bar", " ")
    testing.expect(t, token == "foo" && rest == "bar")
}

@(test)
test_string_to_number :: proc(t: ^testing.T) {
    testing.expect(t, str_to_u64("42") == 42)
    testing.expect(t, str_to_i64("-17") == -17)
    testing.expectf(t, abs(str_to_f64("3.14") - 3.14) < 0.0001, "float conversion error")
    s := i64_to_str(123)
    defer delete(s)
    testing.expect(t, s == "123")
    u := u64_to_str(0xFF)
    defer delete(u)
    testing.expect(t, u == "255")
}

@(test)
test_utf8 :: proc(t: ^testing.T) {
    str := "こんにちは" // 5 runes
    testing.expect(t, utf8_strlen(str) == 5)

    r, sz := utf8_decode(str)
    testing.expect(t, sz > 0 && r == 'こ')

    buf: [4]u8
    n := utf8_encode_rune(buf, 'A')
    testing.expect(t, n == 1 && buf[0] == 'A')

    // UCS-2 roundtrip (basic)
    ucs := utf8_to_ucs2("test")
    defer delete(ucs)
    back := ucs2_to_utf8(ucs)
    defer delete(back)
    testing.expect(t, back == "test")
}

@(test)
test_hashes :: proc(t: ^testing.T) {
    data := []u8("The quick brown fox jumps over the lazy dog")
    a := adler32(data)
    testing.expect(t, a == 0x5BDC0F87, "adler32 mismatch") // known value

    m32 := murmur32(data)
    testing.expect(t, m32 == 0x2E4FF723, "murmur32 mismatch")

    m64 := murmur64(data)
    testing.expect(t, m64 == 0x021C8EF739E1ACD2, "murmur64 mismatch")

    // crc32 known check
    c32 := crc32(data)
    testing.expect(t, c32 == 0xB7B41263, "crc32 mismatch")

    // fnv32a
    f32a := fnv32a(data)
    testing.expect(t, f32a == 0xD87F7E0C, "fnv32a mismatch")
}

@(test)
test_path_utilities :: proc(t: ^testing.T) {
    testing.expect(t, path_is_absolute("/usr/bin"))
    testing.expect(t, !path_is_relative("/usr/bin"))
    name := path_base_name("/usr/bin/test.exe")
    testing.expect(t, name == "test.exe")
    ext := path_extension("test.exe")
    testing.expect(t, ext == "exe")
}

@(test)
test_format :: proc(t: ^testing.T) {
    buf: [64]u8
    n := snprintf(buf[:], "%s %d", "answer", 42)
    testing.expect(t, n > 0 && string(buf[:n - 1]) == "answer 42") // snprintf includes null
}

@(test)
test_time :: proc(t: ^testing.T) {
    t1 := time_now()
    testing.expect(t, t1 > 0.0)
    t2 := utc_time_now()
    testing.expect(t, t2 > 0)
    sleep_ms(1) // just ensure no crash
}

@(test)
test_env :: proc(t: ^testing.T) {
    set_env("GB_TEST_VAR", "odin")
    val := get_env("GB_TEST_VAR")
    defer delete(val)
    testing.expect(t, val == "odin")
    unset_env("GB_TEST_VAR")
    testing.expect(t, get_env("GB_TEST_VAR") == "")
}

@(test)
test_endian_and_bits :: proc(t: ^testing.T) {
    testing.expect(t, endian_swap16(0xABCD) == 0xCDAB)
    testing.expect(t, endian_swap32(0x12345678) == 0x78563412)
    testing.expect(t, endian_swap64(0x1122334455667788) == 0x8877665544332211)
    testing.expect(t, count_set_bits(0xF0F0) == 8)
}