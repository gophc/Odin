// unicode_test.odin - tests for godin/unicode
package godin

import "core:testing"

// ---------------------------------------------------------------------------
// utf8proc_iterate tests
// ---------------------------------------------------------------------------
@test
test_utf8proc_iterate_ascii :: proc(t: ^testing.T) {
    using testing
    s : []u8 = []u8{ 'A', 'B', 'C' }
    cp : utf8proc_int32_t
    // null terminated mode not needed; just test with strlen
    n := utf8proc_iterate(s, 1, &cp)
    expect(t, n == 1, "width of 'A' should be 1")
    expect(t, cp == 0x41, "code point should be 0x41")
    // advance
    n = utf8proc_iterate(s[1:], 1, &cp)
    expect(t, n == 1)
    expect(t, cp == 0x42)
}

@test
test_utf8proc_iterate_2byte :: proc(t: ^testing.T) {
    using testing
    s : []u8 = []u8{ 0xC2, 0xA9 } // ©
    cp : utf8proc_int32_t
    n := utf8proc_iterate(s, 2, &cp)
    expect(t, n == 2)
    expect(t, cp == 0xA9)
}

@test
test_utf8proc_iterate_3byte :: proc(t: ^testing.T) {
    using testing
    s : []u8 = []u8{ 0xE2, 0x82, 0xAC } // €
    cp : utf8proc_int32_t
    n := utf8proc_iterate(s, 3, &cp)
    expect(t, n == 3)
    expect(t, cp == 0x20AC)
}

@test
test_utf8proc_iterate_4byte :: proc(t: ^testing.T) {
    using testing
    s : []u8 = []u8{ 0xF0, 0x9F, 0x98, 0x81 } // 😁
    cp : utf8proc_int32_t
    n := utf8proc_iterate(s, 4, &cp)
    expect(t, n == 4)
    expect(t, cp == 0x1F601)
}

@test
test_utf8proc_iterate_invalid :: proc(t: ^testing.T) {
    using testing
    s : []u8 = []u8{ 0x80 } // invalid
    cp : utf8proc_int32_t
    n := utf8proc_iterate(s, 1, &cp)
    expect(t, n == -3, "invalid UTF-8 should return -3")
}

// ---------------------------------------------------------------------------
// utf8proc_encode_char tests
// ---------------------------------------------------------------------------
@test
test_utf8proc_encode_char :: proc(t: ^testing.T) {
    using testing
    buf: [4]u8
    n := utf8proc_encode_char(0x41, buf[:])
    expect(t, n == 1)
    expect(t, buf[0] == 0x41)

    n = utf8proc_encode_char(0xA9, buf[:])
    expect(t, n == 2)
    expect(t, buf[0] == 0xC2 && buf[1] == 0xA9)
}

// ---------------------------------------------------------------------------
// utf8_decode tests (our custom fast decoder)
// ---------------------------------------------------------------------------
@test
test_utf8_decode_ascii :: proc(t: ^testing.T) {
    using testing
    s := []u8{ 'X' }
    cp : rune
    w := utf8_decode(s, 1, &cp)
    expect(t, w == 1)
    expect(t, cp == 'X')
}

@test
test_utf8_decode_multibyte :: proc(t: ^testing.T) {
    using testing
    s := []u8{ 0xE2, 0x82, 0xAC } // €
    cp : rune
    w := utf8_decode(s, 3, &cp)
    expect(t, w == 3)
    expect(t, cp == 0x20AC)
}

@test
test_utf8_decode_invalid :: proc(t: ^testing.T) {
    using testing
    s := []u8{ 0x80 }
    cp : rune
    w := utf8_decode(s, 1, &cp)
    expect(t, w == 1)
    expect(t, cp == 0xFFFD)
}

// ---------------------------------------------------------------------------
// rune classification tests
// ---------------------------------------------------------------------------
@test
test_rune_is_letter :: proc(t: ^testing.T) {
    using testing
    expect(t, rune_is_letter('a'))
    expect(t, rune_is_letter('Z'))
    expect(t, rune_is_letter('_'))
    expect(t, !rune_is_letter('1'))
    // test some unicode letter
    expect(t, rune_is_letter(0x03B1)) // Greek alpha
}

@test
test_rune_is_digit :: proc(t: ^testing.T) {
    using testing
    expect(t, rune_is_digit('5'))
    expect(t, !rune_is_digit('a'))
    expect(t, rune_is_digit(0x0660)) // Arabic-Indic digit 0
}

@test
test_rune_is_whitespace :: proc(t: ^testing.T) {
    using testing
    expect(t, rune_is_whitespace(' '))
    expect(t, rune_is_whitespace('\t'))
    expect(t, rune_is_whitespace('\n'))
    expect(t, rune_is_whitespace('\r'))
    expect(t, !rune_is_whitespace('x'))
}

// ---------------------------------------------------------------------------
// utf8proc_category tests
// ---------------------------------------------------------------------------
@test
test_utf8proc_category :: proc(t: ^testing.T) {
    using testing
    expect(t, utf8proc_category(0x41) == .Lu)
    expect(t, utf8proc_category(0x61) == .Ll)
    expect(t, utf8proc_category(0x30) == .Nd)
    expect(t, utf8proc_category(0x20) == .Zs)
}

// ---------------------------------------------------------------------------
// utf8proc_decompose_char simple test
// ---------------------------------------------------------------------------
@test
test_utf8proc_decompose_char_NFD :: proc(t: ^testing.T) {
    using testing
    // é (U+00E9) decomposes to e + combining acute accent
    buf: [8]utf8proc_int32_t
    last_bc : int = 0
    n := utf8proc_decompose_char(0xE9, buf[:], len(buf), { .DECOMPOSE }, &last_bc)
    expect(t, n == 2)
    expect(t, buf[0] == 0x65) // e
    expect(t, buf[1] == 0x301) // combining acute
}

@test
test_utf8proc_decompose_char_hangul :: proc(t: ^testing.T) {
    using testing
    // Hangul syllable '한' (U+D55C)
    buf: [8]utf8proc_int32_t
    last_bc : int = 0
    n := utf8proc_decompose_char(0xD55C, buf[:], len(buf), { .DECOMPOSE }, &last_bc)
    expect(t, n == 3, "한 should decompose to 3 jamo")
    expect(t, buf[0] == 0x1112) // ᄒ
    expect(t, buf[1] == 0x1161) // ᅡ
    expect(t, buf[2] == 0x11AB) // ᆫ
}

// ---------------------------------------------------------------------------
// grapheme_break tests
// ---------------------------------------------------------------------------
@test
test_grapheme_break_latin :: proc(t: ^testing.T) {
    using testing
    // 'a' + 'b' should break
    expect(t, utf8proc_grapheme_break('a', 'b'))
    // 'a' + combining accent should not break
    expect(t, !utf8proc_grapheme_break('a', 0x301))
}

@test
test_grapheme_break_hangul :: proc(t: ^testing.T) {
    using testing
    // L + V should not break
    expect(t, !utf8proc_grapheme_break(0x1100, 0x1161)) // ᄀ + ᅡ
    // V + T should not break
    expect(t, !utf8proc_grapheme_break(0x1161, 0x11A8)) // ᅡ + ᆨ
    // L + T should break (not valid Hangul sequence)
    expect(t, utf8proc_grapheme_break(0x1100, 0x11A8))
}

// ---------------------------------------------------------------------------
// ucg_decode_grapheme_clusters integration test
// ---------------------------------------------------------------------------
@test
test_ucg_decode_grapheme_simple :: proc(t: ^testing.T) {
    using testing
    // simple "abc"
    str := transmute([]u8)string("abc")
    graphemes, rune_count, grapheme_count, width, err := ucg_decode_grapheme_clusters(str, len(str))
    expect(t, err == 0)
    expect(t, rune_count == 3)
    expect(t, grapheme_count == 3)
    expect(t, width == 3)
    delete(graphemes)
}

@test
test_ucg_decode_grapheme_emoji :: proc(t: ^testing.T) {
    using testing
    // "👨‍👩‍👧" (family: man ZWJ woman ZWJ girl)
    // U+1F468 + ZWJ + U+1F469 + ZWJ + U+1F467
    str : []u8 = { 0xF0, 0x9F, 0x91, 0xA8, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA9, 0xE2, 0x80, 0x8D, 0xF0, 0x9F, 0x91, 0xA7 }
    graphemes, rune_count, grapheme_count, width, err := ucg_decode_grapheme_clusters(str, len(str))
    expect(t, err == 0)
    expect(t, grapheme_count == 1, "family emoji should be a single grapheme cluster")
    delete(graphemes)
}

@test
test_ucg_decode_grapheme_regional :: proc(t: ^testing.T) {
    using testing
    // "🇺🇸" (flag: US) – two regional indicators for one grapheme cluster
    str : []u8 = { 0xF0, 0x9F, 0x87, 0xBA, 0xF0, 0x9F, 0x87, 0xB8 }
    graphemes, rune_count, grapheme_count, width, err := ucg_decode_grapheme_clusters(str, len(str))
    expect(t, err == 0)
    expect(t, grapheme_count == 1, "flag should be one cluster")
    delete(graphemes)
}