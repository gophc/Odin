// gb_test.odin
package godin

import "core:os"
import "core:strings"
import "core:testing"

// -----------------------------------------------------------------------------
// gb_is_power_of_two
// -----------------------------------------------------------------------------

@(test)
test_gb_is_power_of_two :: proc(t: ^testing.T) {
    testing.expect(t, !gb_is_power_of_two(0), "0 is not power of two")
    testing.expect(t, !gb_is_power_of_two(-1), "negative not power of two")
    testing.expect(t, gb_is_power_of_two(1), "1 is power of two")
    testing.expect(t, gb_is_power_of_two(2), "2 is power of two")
    testing.expect(t, !gb_is_power_of_two(3), "3 not power of two")
    testing.expect(t, gb_is_power_of_two(4), "4 is power of two")
    testing.expect(t, gb_is_power_of_two(64), "64 power of two")
    testing.expect(t, !gb_is_power_of_two(63), "63 not")
}

// -----------------------------------------------------------------------------
// gb_align_forward
// -----------------------------------------------------------------------------

@(test)
test_gb_align_forward :: proc(t: ^testing.T) {
    buf: [32]byte
    ptr := rawptr(&buf[0])

    p := gb_align_forward(ptr, 4)
    testing.expect(t, uintptr(p) % 4 == 0, "align 4")
    testing.expect(t, uintptr(p) >= uintptr(ptr), "not lost")

    // unaligned start
    unaligned := rawptr(uintptr(ptr) + 1)
    q := gb_align_forward(unaligned, 4)
    testing.expect(t, uintptr(q) % 4 == 0, "align from unaligned")
    testing.expect(t, uintptr(q) >= uintptr(unaligned), "not lost")
    testing.expect(t, uintptr(q) - uintptr(unaligned) < 4, "gap less than alignment")
}

// -----------------------------------------------------------------------------
// gb_memchr / gb_memrchr
// -----------------------------------------------------------------------------

@(test)
test_gb_memchr :: proc(t: ^testing.T) {
    data := []byte{1, 2, 3, 4, 2, 5}
    r := gb_memchr_u8(data, 2)
    testing.expect(t, r != nil, "found 2")
    if r != nil {
        testing.expect(t, (^byte)(r)^ == 2, "value is 2")
        testing.expect(t, uintptr(r) == uintptr(&data[1]), "at index 1")
    }
    r2 := gb_memchr_u8(data, 9)
    testing.expect(t, r2 == nil, "not found")
}

@(test)
test_gb_memrchr :: proc(t: ^testing.T) {
    data := []byte{1, 2, 3, 2, 5}
    r := gb_memrchr_u8(data, 2)
    testing.expect(t, r != nil, "found last 2")
    if r != nil {
        testing.expect(t, (^byte)(r)^ == 2, "value is 2")
        testing.expect(t, uintptr(r) == uintptr(&data[3]), "at index 3")
    }
    r2 := gb_memrchr_u8(data, 9)
    testing.expect(t, r2 == nil, "not found")
}

// -----------------------------------------------------------------------------
// gb_reverse
// -----------------------------------------------------------------------------

@(test)
test_gb_reverse :: proc(t: ^testing.T) {
    buf: [5]i32 = {1, 2, 3, 4, 5}
    gb_reverse(&buf[0], 5, size_of(i32))
    testing.expect(t, buf == [5]i32{5, 4, 3, 2, 1}, "reversed")

    // reverse back
    gb_reverse(&buf[0], 5, size_of(i32))
    testing.expect(t, buf == [5]i32{1, 2, 3, 4, 5}, "reversed twice")
}

// -----------------------------------------------------------------------------
// character classification (with custom logic)
// -----------------------------------------------------------------------------

@(test)
test_char_is_hex_digit :: proc(t: ^testing.T) {
    testing.expect(t, gb_char_is_hex_digit('0'), "'0' hex")
    testing.expect(t, gb_char_is_hex_digit('9'), "'9' hex")
    testing.expect(t, gb_char_is_hex_digit('a'), "'a' hex")
    testing.expect(t, gb_char_is_hex_digit('f'), "'f' hex")
    testing.expect(t, gb_char_is_hex_digit('A'), "'A' hex")
    testing.expect(t, gb_char_is_hex_digit('F'), "'F' hex")
    testing.expect(t, !gb_char_is_hex_digit('g'), "'g' not hex")
}

@(test)
test_digit_to_int :: proc(t: ^testing.T) {
    testing.expect(t, gb_digit_to_int('0') == 0, "0 -> 0")
    testing.expect(t, gb_digit_to_int('9') == 9, "9 -> 9")
    // non-digit returns c - 'W', so 'A' - 'W' = -22
    vv := gb_digit_to_int('A')
    testing.expect_value(t, vv, -22)
}

@(test)
test_hex_digit_to_int :: proc(t: ^testing.T) {
    testing.expect(t, gb_hex_digit_to_int('0') == 0, "'0'")
    testing.expect(t, gb_hex_digit_to_int('a') == 10, "'a'")
    testing.expect(t, gb_hex_digit_to_int('f') == 15, "'f'")
    testing.expect(t, gb_hex_digit_to_int('F') == 15, "'F'")
    testing.expect(t, gb_hex_digit_to_int('g') == -1, "'g' invalid")
}

// -----------------------------------------------------------------------------
// string to number conversions (base handling)
// -----------------------------------------------------------------------------

@(test)
test_str_to_u64 :: proc(t: ^testing.T) {
    str, val1 := gb_str_to_u64_s("123", 0)
    testing.expect(t, str == "" && val1 == 123, "decimal 123")

    str2, val2 := gb_str_to_u64_s("0x1A", 0)
    testing.expect(t, str2 == "" && val2 == 0x1A, "hex 0x1A")

    str3, val3 := gb_str_to_u64_s("1A", 16)
    testing.expect(t, str3 == "" && val3 == 0x1A, "forced base 16")
}

@(test)
test_str_to_i64 :: proc(t: ^testing.T) {
    str, val1 := gb_str_to_i64_s("-42", 0)
    testing.expect(t, str == "" && val1 == -42, "negative decimal")
    str2, val2 := gb_str_to_i64_s("0xFF", 0)
    testing.expect(t, str2 == "" && val2 == 255, "hex positive")
}

// -----------------------------------------------------------------------------
// UTF-8 <-> UCS-2 conversion (complex logic)
// -----------------------------------------------------------------------------

@(test)
test_utf8_to_ucs2 :: proc(t: ^testing.T) {
    buf: [32]u16

    // plain ASCII
	_, result := gb_utf8_to_ucs2_s(buf[:], len(buf), "Hello")
    expected: []u16 = {'H','e','l','l','o', 0}
    testing.expect(t, result[0] == expected[0] && result[1] == expected[1], "ascii")

    // two-byte char (U+00E9 é)
    buf2: [32]u16
	_, result2 := gb_utf8_to_ucs2_s(buf2[:], len(buf2), "é")
    testing.expect(t, len(result2) >= 1 && result2[0] == 0x00E9, "é U+00E9")

    // surrogate pair (U+1F600 😀)
    buf3: [32]u16
	_, result3 := gb_utf8_to_ucs2_s(buf3[:], len(buf3), "😀")
    // high surrogate 0xD83D, low 0xDE00
    testing.expect(t, len(result3) >= 2, "emoji length")
    testing.expect(t, result3[0] == 0xD83D && result3[1] == 0xDE00, "emoji surrogate pair")
}

@(test)
test_ucs2_to_utf8 :: proc(t: ^testing.T) {
    buf: [32]byte
    src: []u16 = {'H','i',0}
    _, result := gb_ucs2_to_utf8(buf[:], len(buf), src)
	testing.expect_value(t, len(result), 2)
    testing.expect_value(t, string(result), "Hi")

    // emoji surrogate
    src2: []u16 = {0xD83D, 0xDE00, 0} // 😀
    buf2: [32]byte
	_, result2 := gb_ucs2_to_utf8(buf2[:], len(buf2), src2)
    testing.expect_value(t, len(result2), 4)
    testing.expect_value(t, string(result2), "😀")
}

// -----------------------------------------------------------------------------
// gb_path_is_root (Windows‑style logic)
// -----------------------------------------------------------------------------

@(test)
test_path_is_root :: proc(t: ^testing.T) {
    // note: this logic is Windows‑only (has letter colon backslash)
    testing.expect(t, !gb_path_is_root("C:"), "just drive, not root")
    testing.expect(t, gb_path_is_root("C:\\"), "C:\\ is root")
    testing.expect(t, !gb_path_is_root("C:\\sub"), "subdir not root")
    testing.expect(t, gb_path_is_root("/"), "unix root not detected")
}

// -----------------------------------------------------------------------------
// gb_snprintf (truncation behaviour)
// -----------------------------------------------------------------------------

@(test)
test_snprintf :: proc(t: ^testing.T) {
    buf: [8]byte

    res2 := gb_snprintf(buf[:], "123456789011111")
    // expects -1 because output was truncated
    testing.expect(t, res2 == -1, "truncated returns -1")
    testing.expect(t, string(buf[:7]) == "1234567", "first 7 chars copied")
    testing.expect(t, buf[7] == '8', "null terminated")

    res3 := gb_snprintf(buf[:], "1234567")
    // expects -1 because output was truncated
    testing.expect(t, res3 == 7, "truncated returns -1")
    testing.expect(t, string(buf[:7]) == "1234567", "first 7 chars copied")
    testing.expect(t, buf[7] == 0, "not null terminated")
}

// -----------------------------------------------------------------------------
// gb_get_env / gb_set_env / gb_unset_env
// -----------------------------------------------------------------------------

@(test)
test_env :: proc(t: ^testing.T) {
    test_key := "GB_TEST_VAR_DO_NOT_CLASH"
    defer os.unset_env(test_key)

    gb_set_env(test_key, "odin_rocks")
    val := gb_get_env(test_key); defer delete(val)
    // The returned string may be a copy; compare
    testing.expect(t, strings.compare(val, "odin_rocks") == 0, "env value matches")

    gb_unset_env(test_key)
    val2 := gb_get_env(test_key)
    testing.expect(t, val2 == "", "unset yields empty")
}

// -----------------------------------------------------------------------------
// endian swaps
// -----------------------------------------------------------------------------

@(test)
test_endian_swap :: proc(t: ^testing.T) {
    testing.expect(t, gb_endian_swap16(0x1234) == 0x3412, "swap16")
    testing.expect(t, gb_endian_swap32(0x12345678) == 0x78563412, "swap32")
    testing.expect(t, gb_endian_swap64(0x0123456789ABCDEF) == 0xEFCDAB8967452301, "swap64")
}