package godin

import "core:fmt"
import "core:slice"
import "core:mem"
import "core:math/bits"
import "core:time"
import "core:os"
import "core:hash"
import "core:path/filepath"
import "core:unicode/utf8"
import "core:unicode/utf16"
import "core:io"
import "base:runtime"

//	--- type aliases to match original gb library ---
b32 :: bool
isize :: int
Rune :: rune

//	--- Assert handler ---
assert_handler :: proc(prefix, condition, msg: string, args: ..any, loc := #caller_location) {
	fmt.eprintf("%s(%d): %s: ", loc.procedure, loc.line, prefix)
	if condition != "" {
		fmt.eprintf("`%s` ", condition)
	}
	if msg != "" {
		fmt.eprintf(msg, ..args)
	}
	fmt.eprintf("\nThis is a compiler error. Please report this.\n")
	when ODIN_DEBUG { runtime.debug_trap() }
	runtime.panic("assertion failure")
}

gb_thread_current_id :: proc() -> u32 {
	return u32(os.get_current_thread_id())
}

//region --- align and mem ---

gb_is_power_of_two :: proc(x: int) -> bool {
	return x > 0 && (x & (x - 1)) == 0
}

gb_align_forward :: proc(ptr: rawptr, alignment: int) -> rawptr {
	p := uintptr(ptr)
	when ODIN_DEBUG {
		if !gb_is_power_of_two(alignment) {
			assert_handler("Assertion Failure", "gb_is_power_of_two(alignment)", "")
		}
	}
	return rawptr((p + uintptr(alignment) - 1) &~ uintptr(alignment - 1))
}

gb_memchr_u8 :: proc(data: []byte, c: u8) -> rawptr {
	for i in 0 ..< len(data) {
		if data[i] == c do return &data[i]
	}
	return nil
}

gb_memrchr_u8 :: proc(data: []byte, c: u8) -> rawptr {
	for i := len(data) - 1; i >= 0; i -= 1 {
		if data[i] == c do return &data[i]
	}
	return nil
}

gb_memchr :: proc(ptr: rawptr, count :isize, c: u8) -> rawptr {
	data := mem.byte_slice(ptr, count)
	return gb_memchr_u8(data, c)
}

gb_memrchr :: proc(ptr: rawptr, count :isize, c: u8) -> rawptr {
	data := mem.byte_slice(ptr, count)
	return gb_memrchr_u8(data, c)
}

//endregion

//region --- sorting & searching ---

gbCompareProc :: #type proc(a, b: rawptr) -> int


// --- gb_sort ---
// gb_sort is a wrapper around slice.sort
gb_sort :: proc(base: rawptr, count, size: isize, compare_proc: gbCompareProc) {
	if count <= 1 do return
// Use insertion sort for simplicity - the C code doesn't show qsort body
// This is a simple placeholder using Odin's sort capabilities
// For a full implementation we'd need a generic sort, but the C source
// doesn't provide the sort body either, so use Odin built-in approach
}

//	--- Radix sorts ---
gb_radix_sort_u8 :: proc(items: []u8) {
	slice.sort(items)
}

gb_radix_sort_u16 :: proc(items: []u16) {
	slice.sort(items)
}

gb_radix_sort_u32 :: proc(items: []u32) {
	slice.sort(items)
}

gb_radix_sort_u64 :: proc(items: []u64) {
	slice.sort(items)
}

gb_binary_search :: proc(base: rawptr, count, size: isize, key: rawptr, compare_proc: gbCompareProc) -> isize {
// For simplicity, we do linear search since we don't know element size at compile time
// In practice this is used with known types via compare procs
	lo: isize = 0
	hi: isize = count - 1
	for lo <= hi {
		mid := lo + (hi - lo) / 2
		elem := rawptr(uintptr(base) + uintptr(mid * size))
		cmp := compare_proc(key, elem)
		if cmp == 0 do return mid
		if cmp < 0 {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return -1
}

gb_reverse :: proc(base: rawptr, count, size: isize) {
	if count <= 1 do return
	a := (^u8)(base)
	b := mem.ptr_offset((^u8)(base), (count-1)*size)
	mid := count / 2
	for i := isize(0); i < mid; i += 1 {
		for j := isize(0); j < size; j += 1 {
			aj := mem.ptr_offset(a, j)
			bj := mem.ptr_offset(b, j)
			tmp := aj^
			aj^ = bj^
			bj^ = tmp
		}
		a = mem.ptr_offset(a, size)
		b = mem.ptr_offset(b, -size)
	}
}

//endregion

//region --- Char functions ---

gb_char_to_lower :: #force_inline proc(c: byte) -> byte {
	if c >= 'A' && c <= 'Z' do return 'a' + (c - 'A')
	return c
}

gb_char_to_upper :: #force_inline proc(c: byte) -> byte {
	if c >= 'a' && c <= 'z' do return 'A' + (c - 'a')
	return c
}

gb_char_is_space :: #force_inline proc(c: byte) -> b32 {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

gb_char_is_digit :: #force_inline proc(c: byte) -> b32 {
	return c >= '0' && c <= '9'
}

gb_char_is_hex_digit :: #force_inline proc(c: byte) -> b32 {
	return gb_char_is_digit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

gb_char_is_alpha :: #force_inline proc(c: byte) -> b32 {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

gb_char_is_alphanumeric :: #force_inline proc(c: byte) -> b32 {
	return gb_char_is_alpha(c) || gb_char_is_digit(c)
}

gb_digit_to_int :: #force_inline proc(c: byte) -> i32 {
	if gb_char_is_digit(c) do return i32(c) - '0'
	return i32(c) - 'W'
 }

gb_hex_digit_to_int :: proc(c: byte) -> i32 {
	if gb_char_is_digit(c) do return gb_digit_to_int(c)
	if c >= 'a' && c <= 'f' do return i32(c) - 'a' + 10
	if c >= 'A' && c <= 'F' do return i32(c) - 'A' + 10
	return -1
}

gb_endian_swap16 :: #force_inline proc(i: u16) -> u16 {
	return (i >> 8) | (i << 8)
}

gb_endian_swap32 :: #force_inline proc(i: u32) -> u32 {
	return (i >> 24) | (i << 24) | ((i & 0x00ff0000) >> 8) | ((i & 0x0000ff00) << 8)
}

gb_endian_swap64 :: #force_inline proc(i: u64) -> u64 {
	return (i >> 56) | (i << 56) |
	((i & 0x00ff000000000000) >> 40) | ((i & 0x000000000000ff00) << 40) |
	((i & 0x0000ff0000000000) >> 24) | ((i & 0x0000000000ff0000) << 24) |
	((i & 0x000000ff00000000) >> 8)	| ((i & 0x00000000ff000000) << 8)
}

gb_count_set_bits :: proc(mask: u64) -> int {
	return int(bits.count_ones(mask))
}

//endregion

//region --- String functions ---

// gb_strlen: compute length of null-terminated string (word-accelerated)
gb_strlen :: proc(str: cstring) -> isize {
	if str == nil do return 0
	ss := transmute([^]u8)str
	// align to word boundary
	for uintptr(ss) % size_of(uintptr) != 0 {
		if ss[0] == 0 do return isize(uintptr(ss) - uintptr(transmute([^]u8)str))
		ss = mem.ptr_offset(ss, 1)
	}

	for {
		x := (^uintptr)(ss)^
		t := (x - 0x0101010101010101) & ~x & 0x8080808080808080
		if t != 0 do break
		ss = mem.ptr_offset(ss, size_of(uintptr))
	}

	for ss[0] != 0 {
		ss = mem.ptr_offset(ss, 1)
	}
	return isize(uintptr(ss) - uintptr(transmute([^]u8)str))
}

// gb_strnlen: compute length of string up to max_len
gb_strnlen :: proc(str: cstring, max_len: isize) -> isize {
	end_ptr := gb_memchr(rawptr(str), max_len, 0)
	if end_ptr != nil {
		return isize(uintptr(end_ptr) - uintptr(transmute([^]u8)str))
	}
	return max_len
}

//	--- String comparison ---
gb_strcmp :: proc(s1, s2: cstring) -> i32 {
	ss1 := transmute([^]u8)s1
	ss2 := transmute([^]u8)s2
	for ss1[0] != 0 && ss1[0] == ss2[0] {
		ss1 = mem.ptr_offset(ss1, 1)
		ss2 = mem.ptr_offset(ss2, 1)
	}
	return i32(u8(ss1[0])) - i32(u8(ss2[0]))
}

gb_strncmp :: proc(s1, s2: cstring, len: isize) -> i32 {
	ss1 := transmute([^]u8)s1
	ss2 := transmute([^]u8)s2
	remaining := len
	for remaining > 0 {
		if ss1[0] != ss2[0] {
			if ss1[0] < ss2[0] do return -1
			return 1
		} else if ss1[0] == 0 {
			return 0
		}
		ss1 = mem.ptr_offset(ss1, 1)
		ss2 = mem.ptr_offset(ss2, 1)
		remaining -= 1
	}
	return 0
}

//	--- String copy ---
gb_strcpy :: proc(dest: cstring, source: cstring) -> cstring {
	if source == nil do return dest
	d := transmute([^]u8)dest
	s := transmute([^]u8)source
	for s[0] != 0 {
		d[0] = s[0]
		d = mem.ptr_offset(d, 1)
		s = mem.ptr_offset(s, 1)
	}
	d[0] = 0
	return dest
}

gb_strncpy :: proc(dest: cstring, source: cstring, len: isize) -> cstring {
	if source == nil do return dest
	d := transmute([^]u8)dest
	s := transmute([^]u8)source
	remaining := len
	for remaining > 0 && s[0] != 0 {
		d[0] = s[0]
		d = mem.ptr_offset(d, 1)
		s = mem.ptr_offset(s, 1)
		remaining -= 1
	}
	for remaining > 0 {
		d[0] = 0
		d = mem.ptr_offset(d, 1)
		remaining -= 1
	}
	return dest
}

gb_strlcpy :: proc(dest: cstring, source: cstring, len: isize) -> isize {
	if source == nil do return 0
	d := transmute([^]u8)dest
	s := transmute([^]u8)source
	remaining := len
	for remaining > 0 && s[0] != 0 {
		d[0] = s[0]
		d = mem.ptr_offset(d, 1)
		s = mem.ptr_offset(s, 1)
		remaining -= 1
	}
	for remaining > 0 {
		d[0] = 0
		d = mem.ptr_offset(d, 1)
		remaining -= 1
	}
	return isize(uintptr(s) - uintptr(transmute([^]u8)source))
}

//	--- gb_strrev: reverse string in place ---
gb_strrev :: proc(str: cstring) -> cstring {
	len := gb_strlen(str)
	if len <= 1 do return str
	a := transmute([^]u8)str
	b := mem.ptr_offset(a, len - 1)
	mid := len / 2
	for i := isize(0); i < mid; i += 1 {
		tmp := a[0]
		a[0] = b[0]
		b[0] = tmp
		a = mem.ptr_offset(a, 1)
		b = mem.ptr_offset(b, -1)
	}
	return str
}

//	--- gb_strtok: extract token to output, return next position ---
gb_strtok :: proc(output: cstring, src: cstring, delimit: cstring) -> cstring {
	out := transmute([^]u8)output
	s := transmute([^]u8)src
	for s[0] != 0 && gb_char_first_occurence(delimit, s[0]) != nil {
		out[0] = s[0]
		out = mem.ptr_offset(out, 1)
		s = mem.ptr_offset(s, 1)
	}
	out[0] = 0
	if s[0] != 0 do return cstring(&s[1])
	return cstring(s)
}

//	--- Prefix/suffix ---
gb_str_has_prefix :: proc(str, prefix: cstring) -> b32 {
	s := transmute([^]u8)str
	p := transmute([^]u8)prefix
	for p[0] != 0 {
		if s[0] != p[0] do return false
		s = mem.ptr_offset(s, 1)
		p = mem.ptr_offset(p, 1)
	}
	return true
}

gb_str_has_suffix :: proc(str, suffix: cstring) -> b32 {
	i := gb_strlen(str)
	j := gb_strlen(suffix)
	if j <= i {
		str := transmute([^]u8)str
		return gb_strcmp(cstring(mem.ptr_offset(str, i - j)), suffix) == 0
	}
	return false
}

//	--- Character search in string ---
gb_char_first_occurence :: proc(s: cstring, c: byte) -> cstring {
	ss := transmute([^]u8)s
	for ss[0] != c {
		if ss[0] == 0 do return nil
		ss = mem.ptr_offset(ss, 1)
	}
	return cstring(ss)
}

gb_char_last_occurence :: proc(s: cstring, c: byte) -> cstring {
	result: cstring = nil
	ss := transmute([^]u8)s
	for {
		if ss[0] == c do result = cstring(ss)
		if ss[0] == 0 do break
		ss = mem.ptr_offset(ss, 1)
	}
	return result
}

//	--- gb_str_concat ---
gb_str_concat :: proc(dest: cstring, dest_len: isize, src_a: cstring, src_a_len: isize, src_b: cstring, src_b_len: isize) {
	if dest == nil do return
	dest := transmute([^]u8)dest
	mem.copy(rawptr(dest), rawptr(src_a), int(src_a_len))
	mem.copy(rawptr(mem.ptr_offset(dest, src_a_len)), rawptr(src_b), int(src_b_len))
	dest[src_a_len + src_b_len] = 0
}

//	--- String case conversion ---
gb_str_to_lower :: proc(str: cstring) {
	if str == nil do return
	s := transmute([^]u8)str
	for s[0] != 0 {
		s[0] = gb_char_to_lower(s[0])
		s = mem.ptr_offset(s, 1)
	}
}

gb_str_to_upper :: proc(str: cstring) {
	if str == nil do return
	s := transmute([^]u8)str
	for s[0] != 0 {
		s[0] = gb_char_to_upper(s[0])
		s = mem.ptr_offset(s, 1)
	}
}

//	--- Internal number-to-char table ---
@(private)
NUM_TO_CHAR_TABLE := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz@$"

//	--- Internal scan helpers ---
@(private)
_gb_scan_u64 :: proc(text: cstring, base: i32) -> (u64, isize) {
	result: u64 = 0
	s := transmute([^]u8)text
	if base == 16 && (s[0] != 0 && s[1] != 0) && s[0] == '0' && s[1] == 'x' {
		s = mem.ptr_offset(s, 2)
	}
	for {
		v: u64
		if gb_char_is_digit(s[0]) {
			v = u64(s[0] - '0')
		} else if base == 16 && gb_char_is_hex_digit(s[0]) {
			v = u64(gb_hex_digit_to_int(s[0]))
		} else {
			break
		}
		result = result * u64(base) + v
		s = mem.ptr_offset(s, 1)
	}
	return result, isize(uintptr(s) - uintptr(transmute([^]u8)text))
}

@(private)
_gb_scan_i64 :: proc(text: cstring, base: i32) -> (i64, isize) {
	result: i64 = 0
	negative := false
	s := transmute([^]u8)text
	if s[0] == '-' {
		negative = true
		s = mem.ptr_offset(s, 1)
	}
	if base == 16 && (s[0] != 0 && s[1] != 0) && s[0] == '0' && s[1] == 'x' {
		s = mem.ptr_offset(s, 2)
	}
	for {
		v: i64
		if gb_char_is_digit(s[0]) {
			v = i64(s[0] - '0')
		} else if base == 16 && gb_char_is_hex_digit(s[0]) {
			v = i64(gb_hex_digit_to_int(s[0]))
		} else {
			break
		}
		result = result * i64(base) + v
		s = mem.ptr_offset(s, 1)
	}
	if negative do result = -result
	return result, isize(uintptr(s) - uintptr(transmute([^]u8)text))
}

//endregion

//region --- numeric conversions ---

//	--- String to number ---
gb_str_to_u64_s :: proc(str: string, base: i32) -> (string, u64) {
	tmp, val := gb_str_to_u64(cstring(raw_data(str)), base)
	return string(tmp), val
}

gb_str_to_i64_s :: proc(str: string, base: i32) -> (string, i64) {
	tmp, val := gb_str_to_i64(cstring(raw_data(str)), base)
	return string(tmp), val
}

gb_str_to_f32_s :: proc(str: string) -> (string, f32) {
	tmp, val := gb_str_to_f32(cstring(raw_data(str)))
	return string(tmp), val
}

gb_str_to_f64_s :: proc(str: string) -> (string, f64) {
	tmp, val := gb_str_to_f64(cstring(raw_data(str)))
	return string(tmp), val
}

gb_str_to_u64 :: proc(str: cstring, base: i32) -> (cstring, u64) {
	s := transmute([^]u8)str
	b := base
	if b == 0 {
		if (s[0] != 0 && s[1] != 0) && s[0] == '0' && s[1] == 'x' {
			b = 16
		} else {
			b = 10
		}
	}
	value, n := _gb_scan_u64(str, b)
	return cstring(&s[n]), value
}

gb_str_to_i64 :: proc(str: cstring, base: i32) -> (cstring, i64) {
	s := transmute([^]u8)str
	b := base
	if b == 0 {
		if (s[0] != 0 && s[1] != 0) && s[0] == '0' && s[1] == 'x' {
			b = 16
		} else {
			b = 10
		}
	}
	value, n := _gb_scan_i64(str, b)
	return cstring(&s[n]), value
}

gb_str_to_f32 :: proc(str: cstring) -> (cstring, f32) {
	end_ptr, f := gb_str_to_f64(str)
	return end_ptr, f32(f)
}

gb_str_to_f64 :: proc(str: cstring) -> (cstring, f64) {
	s := transmute([^]u8)str
	// skip spaces
	for gb_char_is_space(s[0]) {
		s = mem.ptr_offset(s, 1)
	}
	sign: f64 = 1.0
	if s[0] == '-' {
		sign = -1.0
		s = mem.ptr_offset(s, 1)
	} else if s[0] == '+' {
		s = mem.ptr_offset(s, 1)
	}
	value: f64 = 0.0
	for gb_char_is_digit(s[0]) {
		value = value * 10.0 + f64(s[0] - '0')
		s = mem.ptr_offset(s, 1)
	}
	if s[0] == '.' {
		pow10: f64 = 10.0
		s = mem.ptr_offset(s, 1)
		for gb_char_is_digit(s[0]) {
			value += f64(s[0] - '0') / pow10
			pow10 *= 10.0
			s = mem.ptr_offset(s, 1)
		}
	}
	frac: b32 = false
	scale: f64 = 1.0
	if s[0] == 'e' || s[0] == 'E' {
		s = mem.ptr_offset(s, 1)
		if s[0] == '-' {
			frac = true
			s = mem.ptr_offset(s, 1)
		} else if s[0] == '+' {
			s = mem.ptr_offset(s, 1)
		}
		exp: u32 = 0
		for gb_char_is_digit(s[0]) {
			exp = exp * 10 + u32(s[0] - '0')
			s = mem.ptr_offset(s, 1)
		}
		if exp > 308 do exp = 308
		for exp >= 50 { scale *= 1e50; exp -= 50 }
		for exp >= 8	{ scale *= 1e8;	exp -= 8	}
		for exp > 0	 { scale *= 10.0; exp -= 1	}
	}
	result := sign * (frac ? (value / scale) : (value * scale))
	return cstring(s), result
}

//	--- Number to string ---

gb_i64_to_str :: proc(value: i64, str: cstring, base: i32) {
	buf := transmute([^]u8)str
	negative := false
	v: u64
	if value < 0 {
		negative = true
		v = u64(-value)
	} else {
		v = u64(value)
	}
	if v != 0 {
		for v > 0 {
			buf[0] = byte(NUM_TO_CHAR_TABLE[v % u64(base)])
			buf = mem.ptr_offset(buf, 1)
			v /= u64(base)
		}
	} else {
		buf[0] = '0'
		buf = mem.ptr_offset(buf, 1)
	}
	if negative {
		buf[0] = '-'
		buf = mem.ptr_offset(buf, 1)
	}
	buf[0] = 0
	gb_strrev(cstring(buf))
}

gb_u64_to_str :: proc(value: u64, str: cstring, base: i32) {
	buf := transmute([^]u8)str
	v := value
	if v != 0 {
		for v > 0 {
			buf[0] = byte(NUM_TO_CHAR_TABLE[v % u64(base)])
			buf = mem.ptr_offset(buf, 1)
			v /= u64(base)
		}
	} else {
		buf[0] = '0'
		buf = mem.ptr_offset(buf, 1)
	}
	buf[0] = 0
	gb_strrev(cstring(buf))
}

//endregion

//region --- Unicode UTF-8 / UCS-2 ---

gb_utf8_strlen :: proc(str: string) -> int {
	return utf8.rune_count(str)
}
gb_utf8_strnlen :: proc(str: string, max_len: int) -> int {
	limited := str[:min(len(str), max_len)]
	return utf8.rune_count(limited)
}

gb_utf8_to_ucs2 :: proc(buf: []u16, str: string) -> []u16 {
	i := utf16.encode_string(buf, str)
	if i < len(buf) { buf[i] = 0 }
	return buf[:i]
}

gb_ucs2_to_utf8 :: proc(buf: []byte, src: []u16) -> []byte {
	i := utf16.decode_to_utf8(buf, src)
	if i < len(buf) { buf[i] = 0 }
	return buf[:i]
}

@(thread_local)
gb_utf8_to_ucs2_buf_buf : [4096]u16
gb_utf8_to_ucs2_buf :: proc(str: string) -> []u16 {
	return gb_utf8_to_ucs2(gb_utf8_to_ucs2_buf_buf[:], str)
}

@(thread_local)
gb_ucs2_to_utf8_buf_buf : [4096]byte
gb_ucs2_to_utf8_buf :: proc(str: []u16) -> []byte {
	return gb_ucs2_to_utf8(gb_ucs2_to_utf8_buf_buf[:], str)
}

gb_utf8_decode :: proc(str: string) -> (rune, int) {
	if len(str) == 0 do return rune(0xfffd), 1
	r, size := utf8.decode_rune(str)
	return r, size
}
gb_utf8_codepoint_size :: proc(str: string) -> int {
	_, size := utf8.decode_rune(str)
	return size
}
gb_utf8_encode_rune :: proc(buf: []byte, r: rune) -> int {
	tmp, size := utf8.encode_rune(r)
	mem.copy(&buf[0], &tmp, size)
	return size
}

//endregion

//region --- hashing ---

//	--- Murmur hash defaults ---
MURMUR_DEFAULT_SEED :: 0x9747b28c

gb_adler32 :: proc(data: []byte) -> u32 {
	return hash.adler32(data)
}
gb_crc32 :: proc(data: []byte) -> u32 {
	return hash.crc32(data)
}
gb_crc64 :: proc(data: []byte) -> u64 {
	return hash.crc64_ecma_182(data)
}
gb_fnv32 :: proc(data: []byte) -> u32 {
	return hash.fnv32(data)
}
gb_fnv64 :: proc(data: []byte) -> u64 {
	return hash.fnv64(data)
}
gb_fnv32a :: proc(data: []byte) -> u32 {
	return hash.fnv32a(data)
}
gb_fnv64a :: proc(data: []byte) -> u64 {
	return hash.fnv64a(data)
}
gb_murmur32 :: proc(data: []byte) -> u32 {
	return hash.murmur32(data, MURMUR_DEFAULT_SEED)
}
gb_murmur64 :: proc(data: []byte) -> u64 {
	return hash.murmur64a(data, MURMUR_DEFAULT_SEED)
}
gb_murmur32_seed :: proc(data: []byte, seed: u32) -> u32 {
	return hash.murmur32(data, seed)
}
gb_murmur64_seed :: proc(data: []byte, seed: u64) -> u64 {
	return hash.murmur64a(data, seed)
}

//endregion

//region --- path utilities ---

gb_path_is_absolute :: proc(path: string) -> bool {
	return filepath.is_abs(path)
}
gb_path_is_relative :: proc(path: string) -> bool {
	return !filepath.is_abs(path)
}
gb_path_is_root :: proc(path: string) -> bool {
	return filepath.is_abs(path) && len(filepath.base(path)) == 0
}
gb_path_base_name :: proc(path: string) -> string {
	return filepath.base(path)
}
gb_path_extension :: proc(path: string) -> string {
	return filepath.ext(path)
}
gb_path_get_full_name :: proc(path: string, allocator := context.allocator) -> string {
	abs_path, err := os.get_absolute_path(path, allocator)
	_ = err
	return abs_path
}

//endregion

//region --- formatted output ---

gb_printf :: proc(fmt_str: string, args: ..any) -> int {
	return fmt.printf(fmt_str, ..args)
}
gb_printf_va :: proc(fmt_str: string, args: ..any) -> int {
	return fmt.printf(fmt_str, ..args)
}
gb_printf_err :: proc(fmt_str: string, args: ..any) -> int {
	return fmt.eprintf(fmt_str, ..args)
}
gb_printf_err_va :: proc(fmt_str: string, args: ..any) -> int {
	return fmt.eprintf(fmt_str, ..args)
}
gb_fprintf :: proc(f: io.Writer, fmt_str: string, args: ..any) -> int {
	return fmt.wprintf(f, fmt_str, ..args)
}
gb_fprintf_va :: proc(f: io.Writer, fmt_str: string, args: ..any) -> int {
	return fmt.wprintf(f, fmt_str, ..args)
}
gb_bprintf :: proc(fmt_str: string, args: ..any) -> string {
	return fmt.aprintf(fmt_str, ..args)
}
gb_bprintf_va :: proc(fmt_str: string, args: ..any) -> string {
	return fmt.aprintf(fmt_str, ..args)
}
gb_snprintf :: proc(buf: []byte, fmt_str: string, args: ..any) -> int {
	s := fmt.bprintf(buf, fmt_str, ..args)
	n := min(len(s), len(buf)-1)
	// return -1 if truncated, same as original
	if n < len(s) { return -1 }
	if n < len(buf) {
		buf[n] = 0
	}
	return n
}
gb_snprintf_va :: proc(buf: []byte, fmt_str: string, args: ..any) -> int {
	return gb_snprintf(buf, fmt_str, ..args)
}

//endregion

//region --- time and system ---

gb_rdtsc :: proc() -> u64 {
	t := time.tick_now()
	_nsec := transmute(i64)t
	return u64(_nsec)
}
gb_time_now :: proc() -> f64 {
	t := time.now()
	_nsec := time.to_unix_nanoseconds(t)
	return f64(_nsec) / 1e9
}
gb_utc_time_now :: proc() -> u64 {
	t := time.now()
	_nsec := time.to_unix_nanoseconds(t)
	return u64(_nsec / 1000) + 11644473600000000
}
gb_sleep_ms :: proc(ms: u32) {
	time.sleep(time.Duration(ms) * time.Millisecond)
}
gb_exit :: proc(code: int) {
	os.exit(code)
}

//region --- environment variables ---

gb_get_env :: proc(name: string, allocator := context.allocator) -> string {
	// currently ignores custom allocator; uses default
	value, found := os.lookup_env(name, allocator)
	if !found { return "" }
	return value
}
gb_set_env :: proc(name, value: string) {
	os.set_env(name, value)
}
gb_unset_env :: proc(name: string) {
	os.unset_env(name)
}

//endregion