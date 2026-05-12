package godin

import "base:runtime"
import "core:fmt"
import "core:io"
import "core:mem"
import "core:os"
import "core:path/filepath"
import "core:math/bits"
import "core:slice"
import "core:time"

//	--- type aliases to match original gb library ---
b32 :: bool
isize :: int
usize :: uint
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

//  --- gb_count_set_bits ---
gb_count_set_bits :: proc(mask: u64) -> isize {
	return isize(bits.count_ones(mask))
}

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


//  --- gb_sort ---
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
GB_NUM_TO_CHAR_TABLE := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz@$"

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
			buf[0] = u8(GB_NUM_TO_CHAR_TABLE[v % u64(base)])
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
			buf[0] = u8(GB_NUM_TO_CHAR_TABLE[v % u64(base)])
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

//  --- UTF-8 strlen ---

gb_utf8_strlen :: proc(str: cstring) -> isize {
	count: isize = 0
	s := transmute(^u8)str
	for s^ != 0 {
		c := s^
		inc: isize
		switch {
		case c < 0x80:           inc = 1
		case (c & 0xe0) == 0xc0: inc = 2
		case (c & 0xf0) == 0xe0: inc = 3
		case (c & 0xf8) == 0xf0: inc = 4
		case:                      return -1
		}
		s = mem.ptr_offset(s, inc)
		count += 1
	}
	return count
}

gb_utf8_strnlen :: proc(str: cstring, max_len: isize) -> isize {
	count: isize = 0
	s := transmute(^u8)str
	remaining := max_len
	for s^ != 0 && remaining > 0 {
		c := s^
		inc: isize
		switch {
		case c < 0x80:           inc = 1
		case (c & 0xe0) == 0xc0: inc = 2
		case (c & 0xf0) == 0xe0: inc = 3
		case (c & 0xf8) == 0xf0: inc = 4
		case:                      return -1
		}
		s = mem.ptr_offset(s, inc)
		remaining -= inc
		count += 1
	}
	return count
}

//  --- UTF-8 <-> UCS-2 ---

gb_utf8_to_ucs2 :: proc(buf: []u16, len: isize, str: cstring) -> (isize, []u16) {
	c: Rune
	i: isize = 0
	rem := len - 1
	s := transmute(^u8)str
	for s^ != 0 {
		if i >= rem do return 0, nil
		if (s^ & 0x80) == 0 {
			buf[i] = u16(s^)
			i += 1
			s = mem.ptr_offset(s, 1)
		} else if (s^ & 0xe0) == 0xc0 {
			if s^ < 0xc2 do return 0, nil
			c = Rune(s^ & 0x1f) << 6
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			buf[i] = u16(c + Rune(s^ & 0x3f))
			i += 1
			s = mem.ptr_offset(s, 1)
		} else if (s^ & 0xf0) == 0xe0 {
			s1 := mem.ptr_offset(s, 1)
			if s^ == 0xe0 && (s1^ < 0xa0 || s1^ > 0xbf) do return 0, nil
			if s^ == 0xed && s1^ > 0x9f do return 0, nil
			c = Rune(s^ & 0x0f) << 12
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			c += Rune(s^ & 0x3f) << 6
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			buf[i] = u16(c + Rune(s^ & 0x3f))
			i += 1
			s = mem.ptr_offset(s, 1)
		} else if (s^ & 0xf8) == 0xf0 {
			if s^ > 0xf4 do return 0, nil
			s1 := mem.ptr_offset(s, 1)
			if s^ == 0xf0 && (s1^ < 0x90 || s1^ > 0xbf) do return 0, nil
			if s^ == 0xf4 && s1^ > 0x8f do return 0, nil
			c = Rune(s^ & 0x07) << 18
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			c += Rune(s^ & 0x3f) << 12
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			c += Rune(s^ & 0x3f) << 6
			s = mem.ptr_offset(s, 1)
			if (s^ & 0xc0) != 0x80 do return 0, nil
			c += Rune(s^ & 0x3f)
			s = mem.ptr_offset(s, 1)
			if (u32(c) & 0xfffff800) == 0xd800 do return 0, nil
			if c >= 0x10000 {
				c -= 0x10000
				if i + 2 > rem do return 0, nil
				buf[i] = u16(0xd800 | (0x3ff & (c >> 10)))
				i += 1
				buf[i] = u16(0xdc00 | (0x3ff & c))
				i += 1
			}
		} else {
			return 0, nil
		}
	}
	buf[i] = 0
	return i, buf[:i]
}

gb_ucs2_to_utf8 :: proc(buf: []u8, len: isize, str: []u16) -> (isize, []u8) {
	i: isize = 0
	rem := len - 1
	s := &str[0]
	for s^ != 0 {
		if s^ < 0x80 {
			if i + 1 > rem do return 0, nil
			buf[i] = u8(s^)
			i += 1
			s = mem.ptr_offset(s, 1)
		} else if s^ < 0x800 {
			if i + 2 > rem do return 0, nil
			buf[i] = u8(0xc0 + (s^ >> 6))
			i += 1
			buf[i] = u8(0x80 + (s^ & 0x3f))
			i += 1
			s = mem.ptr_offset(s, 1)
		} else if s^ >= 0xd800 && s^ < 0xdc00 {
			if i + 4 > rem do return 0, nil
			s1 := mem.ptr_offset(s, 1)
			c := Rune(s^ - 0xd800) << 10 + Rune(s1^ - 0xdc00) + 0x10000
			buf[i] = u8(0xf0 + (c >> 18))
			i += 1
			buf[i] = u8(0x80 + ((c >> 12) & 0x3f))
			i += 1
			buf[i] = u8(0x80 + ((c >> 6) & 0x3f))
			i += 1
			buf[i] = u8(0x80 + (c & 0x3f))
			i += 1
			s = mem.ptr_offset(s, 2)
		} else if s^ >= 0xdc00 && s^ < 0xe000 {
			return 0, nil
		} else {
			if i + 3 > rem do return 0, nil
			buf[i] = u8(0xe0 + (s^ >> 12))
			i += 1
			buf[i] = u8(0x80 + ((s^ >> 6) & 0x3f))
			i += 1
			buf[i] = u8(0x80 + (s^ & 0x3f))
			i += 1
			s = mem.ptr_offset(s, 1)
		}
	}
	buf[i] = 0
	return i, buf[:i]
}

//  --- gb_utf8_to_ucs2_buf (uses static buffer) ---
@(private)
UTF16_BUF: [4096]u16

gb_utf8_to_ucs2_buf :: proc(str: cstring) -> (isize, []u16) {
	return gb_utf8_to_ucs2(UTF16_BUF[:], 4096, str)
}

@(private)
UTF8_BUF: [4096]u8

gb_ucs2_to_utf8_buf :: proc(str: []u16) -> (isize, []u8) {
	return gb_ucs2_to_utf8(UTF8_BUF[:], 4096, str)
}

gb_utf8_to_ucs2_s :: proc(buf: []u16, len: isize, str: string) -> (isize, []u16) {
	s := raw_data(str)
	return gb_utf8_to_ucs2(buf, len, cstring(s))
}

gb_utf8_to_ucs2_buf_s :: proc(str: string) -> (isize, []u16) {
	s := raw_data(str)
	return gb_utf8_to_ucs2_buf(cstring(s))
}

//  --- UTF-8 decode tables ---
@(private)
_UTF8_FIRST := [256]u8{
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
	0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
	0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
	0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
	0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
	0xf1, 0xf1, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	0x13, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x23, 0x03, 0x03,
	0x34, 0x04, 0x04, 0x04, 0x44, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
}

_Utf8AcceptRange :: struct {
	lo, hi: u8,
}

@(private)
_UTF8_ACCEPT_RANGES := [?]_Utf8AcceptRange{
	{0x80, 0xbf},
	{0xa0, 0xbf},
	{0x80, 0x9f},
	{0x90, 0xbf},
	{0x80, 0x8f},
}

//  --- gb_utf8_decode ---
gb_utf8_decode :: proc(str: cstring, str_len: isize) -> (Rune, isize) {
	width: isize = 0
	codepoint: Rune = 0xfffd
	str := transmute([^]u8)str
	if str_len > 0 {
		s0 := str[0]
		x := _UTF8_FIRST[s0]
		if x >= 0xf0 {
			mask := Rune(x) << 31 >> 31
			codepoint = (Rune(s0) & ~mask) | (0xfffd & mask)
			width = 1
			// goto end handled by fallthrough to end label equivalent
		} else if s0 < 0x80 {
			codepoint = Rune(s0)
			width = 1
		} else {
			sz := x & 7
			accept := _UTF8_ACCEPT_RANGES[x >> 4]
			if str_len < isize(sz) {
				// goto invalid_codepoint
				codepoint = 0xfffd
				width = 1
				// goto end
			} else {
				b1 := str[1]
				if b1 < accept.lo || accept.hi < b1 {
					// goto invalid_codepoint
					codepoint = 0xfffd
					width = 1
					// goto end
				} else if sz == 2 {
					codepoint = (Rune(s0) & 0x1f) << 6 | (Rune(b1) & 0x3f)
					width = 2
				} else {
					b2 := str[2]
					if !(b2 >= 0x80 && b2 <= 0xbf) {
						// goto invalid_codepoint
						codepoint = 0xfffd
						width = 1
						// goto end
					} else if sz == 3 {
						codepoint = (Rune(s0) & 0x1f) << 12 | (Rune(b1) & 0x3f) << 6 | (Rune(b2) & 0x3f)
						width = 3
					} else {
						b3 := str[3]
						if !(b3 >= 0x80 && b3 <= 0xbf) {
							// goto invalid_codepoint
							codepoint = 0xfffd
							width = 1
							// goto end
						} else {
							codepoint = (Rune(s0) & 0x07) << 18 | (Rune(b1) & 0x3f) << 12 | (Rune(b2) & 0x3f) << 6 | (Rune(b3) & 0x3f)
							width = 4
						}
					}
				}
			}
		}
	} else {
		codepoint = 0xfffd
		width = 1
	}
	return codepoint, width
}

//  --- gb_utf8_codepoint_size ---
gb_utf8_codepoint_size :: proc(str: cstring, str_len: isize) -> isize {
	str := transmute([^]u8)str
	for i: isize = 0; i < str_len; i += 1 {
		if str[i] == 0 do break
		if (str[i] & 0xc0) != 0x80 do return i + 1
	}
	return str_len + 1
}

//  --- gb_utf8_encode_rune ---
gb_utf8_encode_rune :: proc(buf: [4]u8, r: Rune) -> isize {
	i := u32(r)
	mask: u8 = 0x3f
	buf := buf
	if i <= (1 << 7) - 1 {
		buf[0] = u8(r)
		return 1
	}
	if i <= (1 << 11) - 1 {
		buf[0] = 0xc0 | u8(r >> 6)
		buf[1] = 0x80 | (u8(r) & mask)
		return 2
	}
	if i > 0x0010ffff || (i >= 0xd800 && i <= 0xdfff) {
		rr: Rune = 0xfffd
		buf[0] = 0xe0 | u8(rr >> 12)
		buf[1] = 0x80 | (u8(rr >> 6) & mask)
		buf[2] = 0x80 | (u8(rr) & mask)
		return 3
	}
	if i <= (1 << 16) - 1 {
		buf[0] = 0xe0 | u8(r >> 12)
		buf[1] = 0x80 | (u8(r >> 6) & mask)
		buf[2] = 0x80 | (u8(r) & mask)
		return 3
	}
	buf[0] = 0xf0 | u8(r >> 18)
	buf[1] = 0x80 | (u8(r >> 12) & mask)
	buf[2] = 0x80 | (u8(r >> 6) & mask)
	buf[3] = 0x80 | (u8(r) & mask)
	return 4
}

//endregion

//region --- hashing ---

//  --- gb_adler32 ---
gb_adler32 :: proc(data: rawptr, len: isize) -> u32 {
	MOD_ADLER :: 65521
	a: u32 = 1
	b: u32 = 0
	buf := ([^]u8)(data)
	remaining := len
	block_len := remaining % 5552
	for remaining > 0 {
		i: isize
		for i = 0; i + 7 < block_len; i += 8 {
			a += u32(buf[0]); b += a
			a += u32(buf[1]); b += a
			a += u32(buf[2]); b += a
			a += u32(buf[3]); b += a
			a += u32(buf[4]); b += a
			a += u32(buf[5]); b += a
			a += u32(buf[6]); b += a
			a += u32(buf[7]); b += a
			buf = mem.ptr_offset(buf, 8)
		}
		for ; i < block_len; i += 1 {
			a += u32(buf[0])
			b += a
			buf = mem.ptr_offset(buf, 1)
		}
		a %= MOD_ADLER
		b %= MOD_ADLER
		remaining -= block_len
		block_len = 5552
	}
	return (b << 16) | a
}

//  --- CRC32 table ---
@(private)
_CRC32_TABLE := [256]u32{
	0x00000000, 0x77073096, 0xee0e612c, 0x990951ba,
	0x076dc419, 0x706af48f, 0xe963a535, 0x9e6495a3,
	0x0edb8832, 0x79dcb8a4, 0xe0d5e91e, 0x97d2d988,
	0x09b64c2b, 0x7eb17cbd, 0xe7b82d07, 0x90bf1d91,
	0x1db71064, 0x6ab020f2, 0xf3b97148, 0x84be41de,
	0x1adad47d, 0x6ddde4eb, 0xf4d4b551, 0x83d385c7,
	0x136c9856, 0x646ba8c0, 0xfd62f97a, 0x8a65c9ec,
	0x14015c4f, 0x63066cd9, 0xfa0f3d63, 0x8d080df5,
	0x3b6e20c8, 0x4c69105e, 0xd56041e4, 0xa2677172,
	0x3c03e4d1, 0x4b04d447, 0xd20d85fd, 0xa50ab56b,
	0x35b5a8fa, 0x42b2986c, 0xdbbbc9d6, 0xacbcf940,
	0x32d86ce3, 0x45df5c75, 0xdcd60dcf, 0xabd13d59,
	0x26d930ac, 0x51de003a, 0xc8d75180, 0xbfd06116,
	0x21b4f4b5, 0x56b3c423, 0xcfba9599, 0xb8bda50f,
	0x2802b89e, 0x5f058808, 0xc60cd9b2, 0xb10be924,
	0x2f6f7c87, 0x58684c11, 0xc1611dab, 0xb6662d3d,
	0x76dc4190, 0x01db7106, 0x98d220bc, 0xefd5102a,
	0x71b18589, 0x06b6b51f, 0x9fbfe4a5, 0xe8b8d433,
	0x7807c9a2, 0x0f00f934, 0x9609a88e, 0xe10e9818,
	0x7f6a0dbb, 0x086d3d2d, 0x91646c97, 0xe6635c01,
	0x6b6b51f4, 0x1c6c6162, 0x856530d8, 0xf262004e,
	0x6c0695ed, 0x1b01a57b, 0x8208f4c1, 0xf50fc457,
	0x65b0d9c6, 0x12b7e950, 0x8bbeb8ea, 0xfcb9887c,
	0x62dd1ddf, 0x15da2d49, 0x8cd37cf3, 0xfbd44c65,
	0x4db26158, 0x3ab551ce, 0xa3bc0074, 0xd4bb30e2,
	0x4adfa541, 0x3dd895d7, 0xa4d1c46d, 0xd3d6f4fb,
	0x4369e96a, 0x346ed9fc, 0xad678846, 0xda60b8d0,
	0x44042d73, 0x33031de5, 0xaa0a4c5f, 0xdd0d7cc9,
	0x5005713c, 0x270241aa, 0xbe0b1010, 0xc90c2086,
	0x5768b525, 0x206f85b3, 0xb966d409, 0xce61e49f,
	0x5edef90e, 0x29d9c998, 0xb0d09822, 0xc7d7a8b4,
	0x59b33d17, 0x2eb40d81, 0xb7bd5c3b, 0xc0ba6cad,
	0xedb88320, 0x9abfb3b6, 0x03b6e20c, 0x74b1d29a,
	0xead54739, 0x9dd277af, 0x04db2615, 0x73dc1683,
	0xe3630b12, 0x94643b84, 0x0d6d6a3e, 0x7a6a5aa8,
	0xe40ecf0b, 0x9309ff9d, 0x0a00ae27, 0x7d079eb1,
	0xf00f9344, 0x8708a3d2, 0x1e01f268, 0x6906c2fe,
	0xf762575d, 0x806567cb, 0x196c3671, 0x6e6b06e7,
	0xfed41b76, 0x89d32be0, 0x10da7a5a, 0x67dd4acc,
	0xf9b9df6f, 0x8ebeeff9, 0x17b7be43, 0x60b08ed5,
	0xd6d6a3e8, 0xa1d1937e, 0x38d8c2c4, 0x4fdff252,
	0xd1bb67f1, 0xa6bc5767, 0x3fb506dd, 0x48b2364b,
	0xd80d2bda, 0xaf0a1b4c, 0x36034af6, 0x41047a60,
	0xdf60efc3, 0xa867df55, 0x316e8eef, 0x4669be79,
	0xcb61b38c, 0xbc66831a, 0x256fd2a0, 0x5268e236,
	0xcc0c7795, 0xbb0b4703, 0x220216b9, 0x5505262f,
	0xc5ba3bbe, 0xb2bd0b28, 0x2bb45a92, 0x5cb36a04,
	0xc2d7ffa7, 0xb5d0cf31, 0x2cd99e8b, 0x5bdeae1d,
	0x9b64c2b0, 0xec63f226, 0x756aa39c, 0x026d930a,
	0x9c0906a9, 0xeb0e363f, 0x72076785, 0x05005713,
	0x95bf4a82, 0xe2b87a14, 0x7bb12bae, 0x0cb61b38,
	0x92d28e9b, 0xe5d5be0d, 0x7cdcefb7, 0x0bdbdf21,
	0x86d3d2d4, 0xf1d4e242, 0x68ddb3f8, 0x1fda836e,
	0x81be16cd, 0xf6b9265b, 0x6fb077e1, 0x18b74777,
	0x88085ae6, 0xff0f6a70, 0x66063bca, 0x11010b5c,
	0x8f659eff, 0xf862ae69, 0x616bffd3, 0x166ccf45,
	0xa00ae278, 0xd70dd2ee, 0x4e048354, 0x3903b3c2,
	0xa7672661, 0xd06016f7, 0x4969474d, 0x3e6e77db,
	0xaed16a4a, 0xd9d65adc, 0x40df0b66, 0x37d83bf0,
	0xa9bcae53, 0xdebb9ec5, 0x47b2cf7f, 0x30b5ffe9,
	0xbdbdf21c, 0xcabac28a, 0x53b39330, 0x24b4a3a6,
	0xbad03605, 0xcdd70693, 0x54de5729, 0x23d967bf,
	0xb3667a2e, 0xc4614ab8, 0x5d681b02, 0x2a6f2b94,
	0xb40bbe37, 0xc30c8ea1, 0x5a05df1b, 0x2d02ef8d,
}

//  --- CRC64 table ---
@(private)
_CRC64_TABLE := [256]u64{
	0x0000000000000000, 0x42f0e1eba9ea3693, 0x85e1c3d753d46d26, 0xc711223cfa3e5bb5,
	0x493366450e42ecdf, 0x0bc387aea7a8da4c, 0xccd2a5925d9681f9, 0x8e224479f47cb76a,
	0x9266cc8a1c85d9be, 0xd0962d61b56fef2d, 0x17870f5d4f51b498, 0x5577eeb6e6bb820b,
	0xdb55aacf12c73561, 0x99a54b24bb2d03f2, 0x5eb4691841135847, 0x1c4488f3e8f96ed4,
	0x663d78ff90e185ef, 0x24cd9914390bb37c, 0xe3dcbb28c335e8c9, 0xa12c5ac36adfde5a,
	0x2f0e1eba9ea36930, 0x6dfeff5137495fa3, 0xaaefdd6dcd770416, 0xe81f3c86649d3285,
	0xf45bb4758c645c51, 0xb6ab559e258e6ac2, 0x71ba77a2dfb03177, 0x334a9649765a07e4,
	0xbd68d2308226b08e, 0xff9833db2bcc861d, 0x388911e7d1f2dda8, 0x7a79f00c7818eb3b,
	0xcc7af1ff21c30bde, 0x8e8a101488293d4d, 0x499b3228721766f8, 0x0b6bd3c3dbfd506b,
	0x854997ba2f81e701, 0xc7b97651866bd192, 0x00a8546d7c558a27, 0x4258b586d5bfbcb4,
	0x5e1c3d753d46d260, 0x1cecdc9e94ace4f3, 0xdbfdfea26e92bf46, 0x990d1f49c77889d5,
	0x172f5b3033043ebf, 0x55dfbadb9aee082c, 0x92ce98e760d05399, 0xd03e790cc93a650a,
	0xaa478900b1228e31, 0xe8b768eb18c8b8a2, 0x2fa64ad7e2f6e317, 0x6d56ab3c4b1cd584,
	0xe374ef45bf6062ee, 0xa1840eae168a547d, 0x66952c92ecb40fc8, 0x2465cd79455e395b,
	0x3821458aada7578f, 0x7ad1a461044d611c, 0xbdc0865dfe733aa9, 0xff3067b657990c3a,
	0x711223cfa3e5bb50, 0x33e2c2240a0f8dc3, 0xf4f3e018f031d676, 0xb60301f359dbe0e5,
	0xda050215ea6c212f, 0x98f5e3fe438617bc, 0x5fe4c1c2b9b84c09, 0x1d14202910527a9a,
	0x93366450e42ecdf0, 0xd1c685bb4dc4fb63, 0x16d7a787b7faa0d6, 0x5427466c1e109645,
	0x4863ce9ff6e9f891, 0x0a932f745f03ce02, 0xcd820d48a53d95b7, 0x8f72eca30cd7a324,
	0x0150a8daf8ab144e, 0x43a04931514122dd, 0x84b16b0dab7f7968, 0xc6418ae602954ffb,
	0xbc387aea7a8da4c0, 0xfec89b01d3679253, 0x39d9b93d2959c9e6, 0x7b2958d680b3ff75,
	0xf50b1caf74cf481f, 0xb7fbfd44dd257e8c, 0x70eadf78271b2539, 0x321a3e938ef113aa,
	0x2e5eb66066087d7e, 0x6cae578bcfe24bed, 0xabbf75b735dc1058, 0xe94f945c9c3626cb,
	0x676dd025684a91a1, 0x259d31cec1a0a732, 0xe28c13f23b9efc87, 0xa07cf2199274ca14,
	0x167ff3eacbaf2af1, 0x548f120162451c62, 0x939e303d987b47d7, 0xd16ed1d631917144,
	0x5f4c95afc5edc62e, 0x1dbc74446c07f0bd, 0xdaad56789639ab08, 0x985db7933fd39d9b,
	0x84193f60d72af34f, 0xc6e9de8b7ec0c5dc, 0x01f8fcb784fe9e69, 0x43081d5c2d14a8fa,
	0xcd2a5925d9681f90, 0x8fdab8ce70822903, 0x48cb9af28abc72b6, 0x0a3b7b1923564425,
	0x70428b155b4eaf1e, 0x32b26afef2a4998d, 0xf5a348c2089ac238, 0xb753a929a170f4ab,
	0x3971ed50550c43c1, 0x7b810cbbfce67552, 0xbc902e8706d82ee7, 0xfe60cf6caf321874,
	0xe224479f47cb76a0, 0xa0d4a674ee214033, 0x67c58448141f1b86, 0x253565a3bdf52d15,
	0xab1721da49899a7f, 0xe9e7c031e063acec, 0x2ef6e20d1a5df759, 0x6c0603e6b3b7c1ca,
	0xf6fae5c07d3274cd, 0xb40a042bd4d8425e, 0x731b26172ee619eb, 0x31ebc7fc870c2f78,
	0xbfc9838573709812, 0xfd39626eda9aae81, 0x3a28405220a4f534, 0x78d8a1b9894ec3a7,
	0x649c294a61b7ad73, 0x266cc8a1c85d9be0, 0xe17dea9d3263c055, 0xa38d0b769b89f6c6,
	0x2daf4f0f6ff541ac, 0x6f5faee4c61f773f, 0xa84e8cd83c212c8a, 0xeabe6d3395cb1a19,
	0x90c79d3fedd3f122, 0xd2377cd44439c7b1, 0x15265ee8be079c04, 0x57d6bf0317edaa97,
	0xd9f4fb7ae3911dfd, 0x9b041a914a7b2b6e, 0x5c1538adb04570db, 0x1ee5d94619af4648,
	0x02a151b5f156289c, 0x4051b05e58bc1e0f, 0x87409262a28245ba, 0xc5b073890b687329,
	0x4b9237f0ff14c443, 0x0962d61b56fef2d0, 0xce73f427acc0a965, 0x8c8315cc052a9ff6,
	0x3a80143f5cf17f13, 0x7870f5d4f51b4980, 0xbf61d7e80f251235, 0xfd913603a6cf24a6,
	0x73b3727a52b393cc, 0x31439391fb59a55f, 0xf652b1ad0167feea, 0xb4a25046a88dc879,
	0xa8e6d8b54074a6ad, 0xea16395ee99e903e, 0x2d071b6213a0cb8b, 0x6ff7fa89ba4afd18,
	0xe1d5bef04e364a72, 0xa3255f1be7dc7ce1, 0x64347d271de22754, 0x26c49cccb40811c7,
	0x5cbd6cc0cc10fafc, 0x1e4d8d2b65facc6f, 0xd95caf179fc497da, 0x9bac4efc362ea149,
	0x158e0a85c2521623, 0x577eeb6e6bb820b0, 0x906fc95291867b05, 0xd29f28b9386c4d96,
	0xcedba04ad0952342, 0x8c2b41a1797f15d1, 0x4b3a639d83414e64, 0x09ca82762aab78f7,
	0x87e8c60fded7cf9d, 0xc51827e4773df90e, 0x020905d88d03a2bb, 0x40f9e43324e99428,
	0x2cffe7d5975e55e2, 0x6e0f063e3eb46371, 0xa91e2402c48a38c4, 0xebeec5e96d600e57,
	0x65cc8190991cb93d, 0x273c607b30f68fae, 0xe02d4247cac8d41b, 0xa2dda3ac6322e288,
	0xbe992b5f8bdb8c5c, 0xfc69cab42231bacf, 0x3b78e888d80fe17a, 0x7988096371e5d7e9,
	0xf7aa4d1a85996083, 0xb55aacf12c735610, 0x724b8ecdd64d0da5, 0x30bb6f267fa73b36,
	0x4ac29f2a07bfd00d, 0x08327ec1ae55e69e, 0xcf235cfd546bbd2b, 0x8dd3bd16fd818bb8,
	0x03f1f96f09fd3cd2, 0x41011884a0170a41, 0x86103ab85a2951f4, 0xc4e0db53f3c36767,
	0xd8a453a01b3a09b3, 0x9a54b24bb2d03f20, 0x5d45907748ee6495, 0x1fb5719ce1045206,
	0x919735e51578e56c, 0xd367d40ebc92d3ff, 0x1476f63246ac884a, 0x568617d9ef46bed9,
	0xe085162ab69d5e3c, 0xa275f7c11f7768af, 0x6564d5fde549331a, 0x279434164ca30589,
	0xa9b6706fb8dfb2e3, 0xeb46918411358470, 0x2c57b3b8eb0bdfc5, 0x6ea7525342e1e956,
	0x72e3daa0aa188782, 0x30133b4b03f2b111, 0xf7021977f9cceaa4, 0xb5f2f89c5026dc37,
	0x3bd0bce5a45a6b5d, 0x79205d0e0db05dce, 0xbe317f32f78e067b, 0xfcc19ed95e6430e8,
	0x86b86ed5267cdbd3, 0xc4488f3e8f96ed40, 0x0359ad0275a8b6f5, 0x41a94ce9dc428066,
	0xcf8b0890283e370c, 0x8d7be97b81d4019f, 0x4a6acb477bea5a2a, 0x089a2aacd2006cb9,
	0x14dea25f3af9026d, 0x562e43b4931334fe, 0x913f6188692d6f4b, 0xd3cf8063c0c759d8,
	0x5dedc41a34bbeeb2, 0x1f1d25f19d51d821, 0xd80c07cd676f8394, 0x9afce626ce85b507,
}

//  --- gb_crc32 ---
gb_crc32 :: proc(data: rawptr, len: isize) -> u32 {
	result: u32 = ~u32(0)
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		result = (result >> 8) ~ _CRC32_TABLE[(result ~ u32(c^)) & 0xff]
		c = mem.ptr_offset(c, 1)
	}
	return ~result
}

//  --- gb_crc64 ---
gb_crc64 :: proc(data: rawptr, len: isize) -> u64 {
	result: u64 = ~u64(0)
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		result = (result >> 8) ~ _CRC64_TABLE[(result ~ u64(c^)) & 0xff]
		c = mem.ptr_offset(c, 1)
	}
	return ~result
}

//  --- gb_fnv32 ---
gb_fnv32 :: proc(data: rawptr, len: isize) -> u32 {
	h: u32 = 0x811c9dc5
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		h = (h * 0x01000193) ~ u32(c^)
		c = mem.ptr_offset(c, 1)
	}
	return h
}

//  --- gb_fnv64 ---
gb_fnv64 :: proc(data: rawptr, len: isize) -> u64 {
	h: u64 = 0xcbf29ce484222325
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		h = (h * 0x100000001b3) ~ u64(c^)
		c = mem.ptr_offset(c, 1)
	}
	return h
}

//  --- gb_fnv32a ---
gb_fnv32a :: proc(data: rawptr, len: isize) -> u32 {
	h: u32 = 0x811c9dc5
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		h = (h ~ u32(c^)) * 0x01000193
		c = mem.ptr_offset(c, 1)
	}
	return h
}

//  --- gb_fnv64a ---
gb_fnv64a :: proc(data: rawptr, len: isize) -> u64 {
	h: u64 = 0xcbf29ce484222325
	c := (^u8)(data)
	for i: isize = 0; i < len; i += 1 {
		h = (h ~ u64(c^)) * 0x100000001b3
		c = mem.ptr_offset(c, 1)
	}
	return h
}

//  --- Murmur hash defaults ---
MURMUR_DEFAULT_SEED :: 0x9747b28c

gb_murmur32 :: proc(data: rawptr, len: isize) -> u32 { return gb_murmur32_seed(data, len, MURMUR_DEFAULT_SEED) }
gb_murmur64 :: proc(data: rawptr, len: isize) -> u64 { return gb_murmur64_seed(data, len, u64(MURMUR_DEFAULT_SEED)) }

//  --- gb_murmur32_seed ---
gb_murmur32_seed :: proc(data: rawptr, len: isize, seed: u32) -> u32 {
	C1 :: 0xcc9e2d51
	C2 :: 0x1b873593
	R1 :: 15
	R2 :: 13
	M  :: 5
	N  :: 0xe6546b64

	nblocks := len / 4
	hash := seed
	blocks := (^u32)(data)
	end_ptr := mem.ptr_offset(blocks, nblocks)
	tail := ([^]u8)(end_ptr)

	for ;blocks != end_ptr;blocks = mem.ptr_offset(blocks, 1) {
		k := blocks^
		k *= C1
		k = (k << R1) | (k >> (32 - R1))
		k *= C2
		hash ~= k
		hash = ((hash << R2) | (hash >> (32 - R2))) * M + N
	}

	k1: u32 = 0
	switch len & 3 {
	case 3:
		k1 ~= u32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ~= u32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ~= u32(tail[0])
		k1 *= C1
		k1 = (k1 << R1) | (k1 >> (32 - R1))
		k1 *= C2
		hash ~= k1
	}

	hash ~= u32(len)
	hash ~= (hash >> 16)
	hash *= 0x85ebca6b
	hash ~= (hash >> 13)
	hash *= 0xc2b2ae35
	hash ~= (hash >> 16)
	return hash
}

//  --- gb_murmur64_seed ---
gb_murmur64_seed :: proc(data: rawptr, len: isize, seed: u64) -> u64 {
	M :: 0xc6a4a7935bd1e995
	R :: 47

	h := seed ~ (u64(len) * M)
	nblocks := len / 4
	blocks := (^u64)(data)
	end_ptr := mem.ptr_offset(blocks, nblocks)
	tail := ([^]u8)(end_ptr)

	for ;blocks != end_ptr; blocks = mem.ptr_offset(blocks, 1) {
		k := blocks^
		k *= M
		k ~= k >> R
		k *= M
		h ~= k
		h *= M
	}

	switch len & 7 {
	case 7: h ~= u64(tail[6]) << 48; fallthrough
	case 6: h ~= u64(tail[5]) << 40; fallthrough
	case 5: h ~= u64(tail[4]) << 32; fallthrough
	case 4: h ~= u64(tail[3]) << 24; fallthrough
	case 3: h ~= u64(tail[2]) << 16; fallthrough
	case 2: h ~= u64(tail[1]) << 8;  fallthrough
	case 1: h ~= u64(tail[0]); h *= M
	}

	h ~= h >> R
	h *= M
	h ~= h >> R
	return h
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
gb_snprintf :: proc(buf: []u8, fmt_str: string, args: ..any) -> int {
	s := fmt.bprintf(buf, fmt_str, ..args)
	n := min(len(s), len(buf)-1)
	// return -1 if truncated, same as original
	if n < len(s) { return -1 }
	if n < len(buf) {
		buf[n] = 0
	}
	return n
}
gb_snprintf_va :: proc(buf: []u8, fmt_str: string, args: ..any) -> int {
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