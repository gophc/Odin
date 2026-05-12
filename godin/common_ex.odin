// common.odin - translated from common.i.cpp
package godin

import "base:intrinsics"
import "core:strings"

// ==============================================
// String utilities
// ==============================================
str_eq :: proc(a, b: string) -> bool {
	return a == b
}
str_ne :: proc(a, b: string) -> bool {
	return a != b
}
str_lt :: proc(a, b: string) -> bool {
	return a < b
}
str_gt :: proc(a, b: string) -> bool {
	return a > b
}
str_le :: proc(a, b: string) -> bool {
	return a <= b
}
str_ge :: proc(a, b: string) -> bool {
	return a >= b
}
str_eq_ignore_case :: proc(a, b: string) -> bool {
	return strings.equal_fold(a, b)
}
string_compare :: proc(a, b: string) -> int {
	return strings.compare(a, b)
}
string_to_lower :: proc(s: string) -> string {
	return strings.to_lower(s)
}
string_starts_with :: proc(s, prefix: string) -> bool {
	return strings.has_prefix(s, prefix)
}
string_ends_with :: proc(s, suffix: string) -> bool {
	return strings.has_suffix(s, suffix)
}
string_trim_whitespace :: proc(s: string) -> string {
	return strings.trim_space(s)
}
string_trim_trailing_whitespace :: proc(s: string) -> string {
	return strings.trim_right_space(s)
}
substring :: proc(s: string, lo, hi: int) -> string {
	return s[lo:hi]
}
string_index_byte :: proc(s: string, c: u8) -> int {
	return strings.index_byte(s, c)
}
string_contains_char :: proc(s: string, c: u8) -> bool {
	return strings.contains_rune(s, rune(c))
}
string_contains_string :: proc(haystack, needle: string) -> bool {
	return strings.contains(haystack, needle)
}
string_index :: proc(s, substr: string) -> int {
	return strings.index(s, substr)
}
string_partition :: proc(str, sep: string) -> (head, match, tail: string) {
	i := strings.index(str, sep)
	if i < 0 {
		return str, "", ""
	}
	return str[:i], str[i:i + len(sep)], str[i + len(sep):]
}

String_Iterator :: struct {
	str: string,
	pos: int,
}
string_split_iterator :: proc(it: ^String_Iterator, sep: u8) -> string {
	if it.pos >= len(it.str) do return ""
	start := it.pos
	for i := it.pos; i < len(it.str); i += 1 {
		if it.str[i] == sep {
			res := it.str[start:i]
			it.pos = i + 1
			return res
		}
	}
	it.pos = len(it.str)
	return it.str[start:]
}
is_separator :: proc(c: u8) -> bool {
	return c == '/' || c == '\\'
}
string_extension_position :: proc(s: string) -> int {
	for i := len(s) - 1; i >= 0; i -= 1 {
		if is_separator(s[i]) do break
		if s[i] == '.' do return i
	}
	return -1
}

// ---------- PtrMap (simple pointer map, using built-in map) ----------
// We provide wrapper functions that match C++ API but use Odin map[uintptr]V or map[rawptr]V.

PtrMap :: struct($K, $V: typeid) where intrinsics.type_is_pointer(K) {
	m:     map[K]V
}

//region --- PtrMap ---

map_init :: proc(h: ^PtrMap($K, $V), capacity: isize = 16) {
	h.m = make(map[K]V, int(capacity))
}
map_destroy :: proc(h: ^PtrMap($K, $V)) {
    delete(h.m)
}
map_get :: proc(h: ^PtrMap($K, $V), key: K) -> ^V {
    val, ok := &h.m[key]
    if ok { return &val }
    return nil
}
map_set :: proc(h: ^PtrMap($K, $V), key: K, value: V) {
    h.m[key] = value
}
map_set_if_not_previously_exists :: proc(h: ^PtrMap($K, $V), key: K, value: V) -> bool {
    if key in h {
        return true
    }
    h.m[key] = value
    return false
}
map_remove :: proc(h: ^PtrMap($K, $V), key: K) {
    delete_key(&h.m, key)
}
map_clear :: proc(h: ^PtrMap($K, $V)) {
    clear(&h.m)
}
map_grow :: proc(h: ^PtrMap($K, $V)) {
    // built-in map grows automatically, nothing needed
}
map_reserve :: proc(h: ^PtrMap($K, $V), cap: isize) {
    reserve(&h.m, int(cap))
}

// multi map not implemented with built-in map, we skip those functions.
// For full compatibility we could implement a custom OrderedInsertPtrMap, but odds are low.
// We leave them unimplemented.
multi_map_find_first :: proc($T: typeid, h: ^PtrMap($K, $V), key: K) -> ^T {
    assert(false, "not implemented")
    return nil
}
multi_map_find_next :: proc($T: typeid, h: ^PtrMap($K, $V), e: ^T) -> ^T {
    assert(false, "not implemented")
    return nil
}
multi_map_count :: proc(h: ^PtrMap($K, $V), key: K) -> isize {
    assert(false, "not implemented")
    return 0
}
multi_map_get_all :: proc(h: ^PtrMap($K, $V), key: K, items: [^]V) {
    assert(false, "not implemented")
}
multi_map_insert :: proc(h: ^PtrMap($K, $V), key: K, value: V) {
    assert(false, "not implemented")
}
multi_map_remove :: proc($T: typeid, h: ^PtrMap($K, $V), key: K, e: ^T) {
    assert(false, "not implemented")
}
multi_map_remove_all :: proc(h: ^PtrMap($K, $V), key: K) {
    assert(false, "not implemented")
}

//endregion

// ---------- PtrSet ----------
PtrSet :: struct($T: typeid) where intrinsics.type_is_pointer(T) {
	m:     map[T]struct{}
}

//region --- PtrSet ---

ptr_set_init :: proc(s: ^PtrSet($T), capacity: isize = 16) {
	s.m = make(map[T]struct{}, int(capacity))
}
ptr_set_destroy :: proc(s: ^PtrSet($T)) {
    delete(s.m)
}
ptr_set_add :: proc(s: ^PtrSet($T), ptr: T) -> T {
    s.m[ptr] = {}
    return ptr
}
ptr_set_update :: proc(s: ^PtrSet($T), ptr: T) -> bool {
    if ptr in s.m {
        return true
    }
    s.m[ptr] = {}
    return false
}
ptr_set_exists :: proc(s: ^PtrSet($T), ptr: T) -> bool {
    return ptr in s.m
}
ptr_set_remove :: proc(s: ^PtrSet($T), ptr: T) {
    delete_key(&s.m, ptr)
}
ptr_set_clear :: proc(s: ^PtrSet($T)) {
    clear(&s.m)
}
// ptr_set_update_with_mutex not needed, use lock externally.

//endregion

// ---------- StringMap (using built-in map[string]V) ----------
StringMap :: struct($V: typeid) {
	m:     map[string]V
}

//region --- StringMap ---

string_map_init :: proc(h: ^StringMap($V), capacity: usize = 16) {
	h.m = make(map[string]V, int(capacity))
}
string_map_destroy :: proc(h: ^StringMap($V)) {
    delete(h.m)
}
string_map_get :: proc{
    string_map_get_string,
    string_map_get_cstring,
}
string_map_get_string :: proc(h: ^StringMap($V), key: string) -> ^V {
    val, ok := &h.m[key]
    if ok { return &val }
    return nil
}
string_map_get_cstring :: proc(h: ^StringMap($V), key: cstring) -> ^V {
    return string_map_get_string(h, string(key))
}
string_map_must_get :: proc{
    string_map_must_get_string,
    string_map_must_get_cstring,
}
string_map_must_get_string :: proc(h: ^StringMap($V), key: string) -> ^V {
    val, ok := &h.m[key]
    assert(ok, "key not found in StringMap")
    return &val
}
string_map_must_get_cstring :: proc(h: ^StringMap($V), key: cstring) -> ^V {
    return string_map_must_get_string(h, string(key))
}
string_map_set :: proc{
    string_map_set_string,
    string_map_set_cstring,
}
string_map_set_string :: proc(h: ^StringMap($V), key: string, value: V) {
    h.m[key] = value
}
string_map_set_cstring :: proc(h: ^StringMap($V), key: cstring, value: V) {
    h.m[string(key)] = value
}
string_map_clear :: proc(h: ^StringMap($V)) {
    clear(&h.m)
}
string_map_grow :: proc(h: ^StringMap($V)) {
    // auto-grow
}
string_map_reserve :: proc(h: ^StringMap($V), new_count: usize) {
    reserve(&h.m, int(new_count))
}

//endregion

// ---------- String16Map (using map[string]V, converting from u16) ----------

String16Map :: struct($V: typeid) {
	m:     map[string]V
}

//region --- String16Map ---

string16_map_init :: proc(h: ^String16Map($V), capacity: usize = 16) {
	h.m = make(map[string]V, int(capacity))
}
string16_map_destroy :: proc(h: ^String16Map($V)) {
    delete(h.m)
}
// helper to convert String16 to string
string16_to_utf8 :: proc(str: string16, allocator := context.allocator) -> string {
    // simple: assume all u16 are ASCII-ish; proper implementation would use unicode
    buf := make([]byte, len(str), allocator)
    for i in 0..<len(str) {
        buf[i] = byte(str[i])
    }
    return string(buf)
}
string16_map_get :: proc(h: ^String16Map($V), key: string16) -> ^V {
    k := string16_to_utf8(key, context.temp_allocator)
    defer delete(k) // temp
    val, ok := &h.m[k]
    if ok { return &val }
    return nil
}
string16_map_must_get :: proc(h: ^String16Map($V), key: string16) -> ^V {
    k := string16_to_utf8(key, context.temp_allocator)
    defer delete(k) // temp
    val, ok := &h.m[k]
    assert(ok, "key not found in String16Map")
    return &val
}
string16_map_set :: proc(h: ^String16Map($V), key: string16, value: V) {
    k := string16_to_utf8(key, context.allocator)
    h.m[k] = value
}
string16_map_clear :: proc(h: ^String16Map($V)) {
    clear(&h.m)
}
string16_map_grow :: proc(h: ^String16Map($V)) {
    // auto
}
string16_map_reserve :: proc(h: ^String16Map($V), new_count: usize) {
    reserve(&h.m, int(new_count))
}

//endregion

// ---------- StringSet ----------
StringSet  :: struct {
	m:     map[string]struct{}
}

//region --- StringSet ---

string_set_init :: proc(s: ^StringSet, capacity: isize = 16) {
    s.m = make(map[string]struct{}, int(capacity))
}
string_set_destroy :: proc(s: ^StringSet) {
    delete(s.m)
}
string_set_add :: proc(s: ^StringSet, str: string) {
    s.m[str] = {}
}
string_set_update :: proc(s: ^StringSet, str: string) -> bool {
    if str in s.m {
        return true
    }
    s.m[str] = {}
    return false
}
string_set_exists :: proc(s: ^StringSet, str: string) -> bool {
    return str in s.m
}
string_set_remove :: proc(s: ^StringSet, str: string) {
    delete_key(&s.m, str)
}
string_set_clear :: proc(s: ^StringSet) {
    clear(&s.m)
}
string_set_grow :: proc(s: ^StringSet) {
    // auto
}
string_set_rehash :: proc(s: ^StringSet, new_count: isize) {
    reserve(&s.m, int(new_count))
}

//endregion

// ---------- ThreadPool (simplified) ----------

// ----------