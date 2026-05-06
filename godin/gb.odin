// gb.odin - Pure Odin reimplementation of the gb library
package godin

import "core:fmt"
import "core:math/bits"
import "core:mem"
import "core:os"
import "core:slice"
import "core:strconv"
import "core:strings"
import "core:time"
import "core:unicode/utf8"

// =========================================================================
// Basic types
// =========================================================================
b8 :: bool
b16 :: i16
b32 :: i32
usize :: uint
isize :: int
uintptr :: uintptr
intptr :: int

Rune :: rune
f32 :: f32
f64 :: f64

// =========================================================================
// Constants
// =========================================================================
GB_ALLOCATOR_FLAG_CLEAR_TO_ZERO :: 1

GB_FILE_MODE_READ :: 0o1
GB_FILE_MODE_WRITE :: 0o2
GB_FILE_MODE_APPEND :: 0o4
GB_FILE_MODE_RW :: 0o10
GB_FILE_MODE_MODES :: GB_FILE_MODE_READ | GB_FILE_MODE_WRITE | GB_FILE_MODE_APPEND | GB_FILE_MODE_RW

GB_SEEK_WHENCE_BEGIN :: 0
GB_SEEK_WHENCE_CURRENT :: 1
GB_SEEK_WHENCE_END :: 2

// =========================================================================
// Assert (simplified)
// =========================================================================
gb_assert_handler :: proc(prefix, condition: string, loc := #caller_location, msg: string = "", args: ..any) {
    fmt.eprintf("%s(%d): %s: ", loc.file_path, loc.line, prefix)
    if condition != "" {
        fmt.eprintf("`%s` ", condition)
    }
    if msg != "" {
        fmt.eprintf(msg, ..args)
    }
    fmt.eprintln()
    fmt.eprintln("This is a compiler error. Please report this.")
    runtime.debug_trap()
}

// =========================================================================
// Math helpers
// =========================================================================
is_power_of_two :: proc(x: isize) -> b8 {
    if x <= 0 do return false
    return (x & (x - 1)) == 0
}

// =========================================================================
// Pointer arithmetic
// =========================================================================
align_forward :: proc(ptr: rawptr, alignment: isize) -> rawptr {
    assert(is_power_of_two(alignment))
    p := uintptr(ptr)
    return rawptr(uintptr((p + uintptr(alignment - 1)) & ~uintptr(alignment - 1)))
}

pointer_add :: proc(ptr: rawptr, bytes: isize) -> rawptr {
    return rawptr(uintptr(ptr) + uintptr(bytes))
}

pointer_sub :: proc(ptr: rawptr, bytes: isize) -> rawptr {
    return rawptr(uintptr(ptr) - uintptr(bytes))
}

pointer_diff :: proc(begin, end: rawptr) -> isize {
    return isize(uintptr(end) - uintptr(begin))
}

// =========================================================================
// Memory operations
// =========================================================================
zero_size :: proc(ptr: rawptr, size: isize) {
    if size > 0 {
        mem.zero(ptr, int(size))
    }
}

memcopy :: proc(dest, source: rawptr, n: isize) -> rawptr {
    if dest == nil do return nil
    mem.copy(dest, source, int(n))
    return dest
}

memmove :: proc(dest, source: rawptr, n: isize) -> rawptr {
    if dest == nil do return nil
    if dest == source do return dest
    mem.move(dest, source, int(n))
    return dest
}

memset :: proc(dest: rawptr, c: u8, n: isize) -> rawptr {
    if dest == nil do return nil
    mem.set(dest, int(c), int(n))
    return dest
}

memcompare :: proc(s1, s2: rawptr, size: isize) -> i32 {
    if s1 == nil || s2 == nil do return 0
    p1 := ([^]u8)(s1)
    p2 := ([^]u8)(s2)
    for i : isize = 0; i < size; i += 1 {
        if p1[i] != p2[i] {
            return i32(p1[i]) - i32(p2[i])
        }
    }
    return 0
}

memswap :: proc(i, j: rawptr, size: isize) {
    if i == j do return
    switch size {
    case 4:
        a := cast(^u32)i
        b := cast(^u32)j
        a^, b^ = b^, a^
    case 8:
        a := cast(^u64)i
        b := cast(^u64)j
        a^, b^ = b^, a^
    case:
        if size < 8 {
            a := ([^]u8)(i)
            b := ([^]u8)(j)
            for k : isize = 0; k < size; k += 1 {
                a[k], b[k] = b[k], a[k]
            }
        } else {
            buf := make([]u8, 256, context.temp_allocator)
            for off : isize = 0; off < size; {
                chunk := min(isize(len(buf)), size - off)
                src := pointer_add(i, off)
                dst := pointer_add(j, off)
                mem.copy(raw_data(buf), src, int(chunk))
                mem.copy(src, dst, int(chunk))
                mem.copy(dst, raw_data(buf), int(chunk))
                off += chunk
            }
        }
    }
}

memchr :: proc(data: rawptr, c: u8, n: isize) -> rawptr {
    p := ([^]u8)(data)
    for i : isize = 0; i < n; i += 1 {
        if p[i] == c {
            return &p[i]
        }
    }
    return nil
}

memrchr :: proc(data: rawptr, c: u8, n: isize) -> rawptr {
    p := ([^]u8)(data)
    for i := n - 1; i >= 0; i -= 1 {
        if p[i] == c {
            return &p[i]
        }
    }
    return nil
}

// =========================================================================
// Thread ID (simplified, not available in Odin portably)
// =========================================================================
thread_current_id :: proc() -> u32 {
    return u32(os.current_thread_id())
}

// =========================================================================
// Affinity (not implemented)
// =========================================================================
gbAffinity :: struct {
    is_accurate: b32,
    core_count: isize,
    thread_count: isize,
    core_masks: [64]usize,
}

affinity_init :: proc(a: ^gbAffinity) {
    a.is_accurate = false
    a.core_count = 1
    a.thread_count = 1
    a.core_masks[0] = 1
}
affinity_destroy :: proc(a: ^gbAffinity) {
}
affinity_set :: proc(a: ^gbAffinity, core, thread: isize) -> b32 {
    return false
}
affinity_thread_count_for_core :: proc(a: ^gbAffinity, core: isize) -> isize {
    return a.thread_count
}

// =========================================================================
// Allocator
// =========================================================================
gbAllocator :: mem.Allocator

alloc_align :: proc(a: gbAllocator, size: isize, alignment: isize) -> rawptr {
    if size <= 0 do return nil
    ptr := mem.alloc_bytes(int(size), int(alignment), a)
    if ptr != nil {
    // The original gb library always allocates with ClearToZero
        mem.zero(ptr, int(size))
    }
    return ptr
}

alloc :: proc(a: gbAllocator, size: isize) -> rawptr {
    return alloc_align(a, size, 2 * size_of(rawptr))
}

free :: proc(a: gbAllocator, ptr: rawptr) {
    if ptr != nil {
        mem.free(ptr, a)
    }
}

free_all :: proc(a: gbAllocator) {
    mem.free_all(a)
}

resize :: proc(a: gbAllocator, ptr: rawptr, old_size, new_size: isize) -> rawptr {
    return resize_align(a, ptr, old_size, new_size, 2 * size_of(rawptr))
}

resize_align :: proc(a: gbAllocator, ptr: rawptr, old_size, new_size, alignment: isize) -> rawptr {
    if new_size == 0 {
        if ptr != nil {
            mem.free(ptr, a)
        }
        return nil
    }
    if ptr == nil {
        return alloc_align(a, new_size, alignment)
    }
    if old_size == new_size && uintptr(ptr) % uintptr(alignment) == 0 {
        return ptr
    }
    // Allocate new block, copy old data, free old
    new_ptr := alloc_align(a, new_size, alignment)
    if new_ptr != nil {
        copy_sz := min(old_size, new_size)
        mem.copy(new_ptr, ptr, int(copy_sz))
        mem.free(ptr, a)
    }
    return new_ptr
}

alloc_copy :: proc(a: gbAllocator, src: rawptr, size: isize) -> rawptr {
    ptr := alloc(a, size)
    if ptr != nil {
        mem.copy(ptr, src, int(size))
    }
    return ptr
}

alloc_copy_align :: proc(a: gbAllocator, src: rawptr, size, alignment: isize) -> rawptr {
    ptr := alloc_align(a, size, alignment)
    if ptr != nil {
        mem.copy(ptr, src, int(size))
    }
    return ptr
}

alloc_str :: proc(a: gbAllocator, str: string) -> string {
    return strings.clone(str, a)
}

alloc_str_len :: proc(a: gbAllocator, data: rawptr, len: isize) -> string {
    buf := make([]u8, len, a)
    mem.copy(raw_data(buf), data, int(len))
    return string(buf)
}

default_resize_align :: proc(a: gbAllocator, old_memory: rawptr, old_size, new_size, alignment: isize) -> rawptr {
    if old_memory == nil do return alloc_align(a, new_size, alignment)
    if new_size == 0 {
        free(a, old_memory)
        return nil
    }
    if new_size < old_size {
        new_size = old_size
    }
    if old_size == new_size {
        return old_memory
    }
    new_memory := alloc_align(a, new_size, alignment)
    if new_memory == nil do return nil
    mem.copy(new_memory, old_memory, int(min(old_size, new_size)))
    free(a, old_memory)
    return new_memory
}

heap_allocator :: proc() -> gbAllocator {
    return context.allocator
}

// =========================================================================
// Compare procedures
// =========================================================================
gbCompareProc :: #type proc(a: rawptr, b: rawptr) -> i32

@(private)
_cmp_offset: isize

_i16_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^i16)(pointer_add(a, _cmp_offset))
    pb := (^i16)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_i32_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^i32)(pointer_add(a, _cmp_offset))
    pb := (^i32)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_i64_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^i64)(pointer_add(a, _cmp_offset))
    pb := (^i64)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_isize_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^isize)(pointer_add(a, _cmp_offset))
    pb := (^isize)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_f32_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^f32)(pointer_add(a, _cmp_offset))
    pb := (^f32)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_f64_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^f64)(pointer_add(a, _cmp_offset))
    pb := (^f64)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_char_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^u8)(pointer_add(a, _cmp_offset))
    pb := (^u8)(pointer_add(b, _cmp_offset))
    if pa^ < pb^ do return -1
    if pa^ > pb^ do return 1
    return 0
}

_str_cmp :: proc(a, b: rawptr) -> i32 {
    sa := (^string)(pointer_add(a, _cmp_offset))
    sb := (^string)(pointer_add(b, _cmp_offset))
    return i32(strings.compare(sa^, sb^))
}

// These generator functions are not directly translatable;
// Odin code can capture the offset in a closure if needed.
// We omit the factory functions and use the comparators directly.
// (Original C had: gb_i16_cmp(isize offset) returning a function pointer)
// For compatibility, we provide a placeholder.
gb_i16_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _i16_cmp
}
gb_i32_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _i32_cmp
}
gb_i64_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _i64_cmp
}
gb_isize_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _isize_cmp
}
gb_f32_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _f32_cmp
}
gb_f64_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _f64_cmp
}
gb_char_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _char_cmp
}
gb_str_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset
    return _str_cmp
}

// =========================================================================
// Sorting
// =========================================================================
// Simple quicksort implementation
@(private)
_quicksort :: proc(base: rawptr, lo, hi: isize, size: isize, cmp: gbCompareProc) {
    if lo >= hi do return
    p := _partition(base, lo, hi, size, cmp)
    _quicksort(base, lo, p - 1, size, cmp)
    _quicksort(base, p + 1, hi, size, cmp)
}

@(private)
_partition :: proc(base: rawptr, lo, hi: isize, size: isize, cmp: gbCompareProc) -> isize {
// Use middle element as pivot to avoid worst-case on sorted arrays
    mid := lo + (hi - lo) / 2
    pivot_ptr := pointer_add(base, mid * size)
    // Move pivot to the end temporarily
    memswap(pivot_ptr, pointer_add(base, hi * size), size)

    pivot_val := pointer_add(base, hi * size)
    i := lo
    for j := lo; j < hi; j += 1 {
        if cmp(pointer_add(base, j * size), pivot_val) < 0 {
            memswap(pointer_add(base, i * size), pointer_add(base, j * size), size)
            i += 1
        }
    }
    // Move pivot to its final place
    memswap(pointer_add(base, i * size), pointer_add(base, hi * size), size)
    return i
}

sort :: proc(base: rawptr, count, size: isize, cmp: gbCompareProc) {
    if count < 2 do return
    _quicksort(base, 0, count - 1, size, cmp)
}

// Radix sorts are not implemented in Odin; leave as stubs
radix_sort_u8 :: proc(items: []u8, temp: []u8) {
// Not implemented
}
radix_sort_u16 :: proc(items: []u16, temp: []u16) {
// Not implemented
}
radix_sort_u32 :: proc(items: []u32, temp: []u32) {
// Not implemented
}
radix_sort_u64 :: proc(items: []u64, temp: []u64) {
// Not implemented
}

binary_search :: proc(base: rawptr, count, size: isize, key: rawptr, cmp: gbCompareProc) -> isize {
    lo : isize = 0
    hi := count
    for lo < hi {
        mid := lo + (hi - lo) / 2
        res := cmp(key, pointer_add(base, mid * size))
        if res < 0 {
            hi = mid
        } else if res > 0 {
            lo = mid + 1
        } else {
            return mid
        }
    }
    return -1
}

reverse :: proc(base: rawptr, count, size: isize) {
    for i, j := isize(0), count - 1; i < j; i, j = i + 1, j - 1 {
        memswap(pointer_add(base, i * size), pointer_add(base, j * size), size)
    }
}

// =========================================================================
// Character classification
// =========================================================================
char_to_lower :: proc(c: u8) -> u8 {
    if c >= 'A' && c <= 'Z' do return c + 32
    return c
}
char_to_upper :: proc(c: u8) -> u8 {
    if c >= 'a' && c <= 'z' do return c - 32
    return c
}
char_is_space :: proc(c: u8) -> b8 {
    return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
char_is_digit :: proc(c: u8) -> b8 {
    return c >= '0' && c <= '9'
}
char_is_hex_digit :: proc(c: u8) -> b8 {
    return char_is_digit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
char_is_alpha :: proc(c: u8) -> b8 {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}
char_is_alphanumeric :: proc(c: u8) -> b8 {
    return char_is_alpha(c) || char_is_digit(c)
}
digit_to_int :: proc(c: u8) -> i32 {
    if char_is_digit(c) do return i32(c - '0')
    return i32(c - 'W') // 'W' is used in the original for letters? This is unusual but kept as is.
}
hex_digit_to_int :: proc(c: u8) -> i32 {
    if char_is_digit(c) do return i32(c - '0')
    if c >= 'a' && c <= 'f' do return i32(c - 'a' + 10)
    if c >= 'A' && c <= 'F' do return i32(c - 'A' + 10)
    return -1
}

// =========================================================================
// String utilities
// =========================================================================
str_to_lower :: proc(s: string) -> string {
    buf := make([]u8, len(s), context.allocator)
    for r, i in s {
        buf[i] = char_to_lower(u8(r))
    }
    return string(buf)
}
str_to_upper :: proc(s: string) -> string {
    buf := make([]u8, len(s), context.allocator)
    for r, i in s {
        buf[i] = char_to_upper(u8(r))
    }
    return string(buf)
}

strlen :: proc(s: string) -> isize {
    return isize(len(s))
}

strnlen :: proc(s: string, max_len: isize) -> isize {
    return min(isize(len(s)), max_len)
}

strcmp :: proc(s1, s2: string) -> i32 {
    return i32(strings.compare(s1, s2))
}

strncmp :: proc(s1, s2: string, n: isize) -> i32 {
    lhs := s1[:min(isize(len(s1)), n)]
    rhs := s2[:min(isize(len(s2)), n)]
    return i32(strings.compare(lhs, rhs))
}

// strcpy, strncpy are awkward in Odin because strings are immutable.
// We provide versions that write into a provided buffer.
strcpy :: proc(dest: []u8, source: string) {
    copy(dest, source)
    if len(dest) > len(source) {
        dest[len(source)] = 0
    }
}
strncpy :: proc(dest: []u8, source: string, max_len: isize) {
    n := min(max_len, isize(len(dest)))
    m := min(n, isize(len(source)))
    copy(dest, source[:m])
    if m < n {
        dest[m] = 0
    }
}
strlcpy :: proc(dest: []u8, source: string, max_len: isize) -> isize {
    n := min(max_len - 1, isize(len(source)))
    copy(dest, source[:n])
    dest[n] = 0
    return n
}
strrev :: proc(s: string) -> string {
// Reverse runes, not bytes, to handle UTF-8 correctly
    runes: [dynamic]rune
    defer delete(runes)
    for r in s {
        append(&runes, r)
    }
    slice.reverse(runes[:])
    // Encode runes back to UTF-8
    buf := make([]u8, 0, len(s), context.allocator)
    for r in runes {
        arr: [4]u8
        n := utf8.encode_rune(r, arr[:])
        for i := 0; i < n; i += 1 {
            append(&buf, arr[i])
        }
    }
    return string(buf)
}
strtok :: proc(output: ^string, src: string, delimit: string) -> string {
// Not fully implemented; returns empty string
// Original: iterates through src copying non-delim chars to output, then sets *output and returns next position.
    return ""
}
str_has_prefix :: proc(s, prefix: string) -> b8 {
    return strings.has_prefix(s, prefix)
}
str_has_suffix :: proc(s, suffix: string) -> b8 {
    return strings.has_suffix(s, suffix)
}
char_first_occurence :: proc(s: string, c: u8) -> int {
    return strings.index_byte(s, c)
}
char_last_occurence :: proc(s: string, c: u8) -> int {
    return strings.last_index_byte(s, c)
}
str_concat :: proc(a, b: string, allocator := context.allocator) -> string {
    return strings.concatenate({ a, b }, allocator)
}

// =========================================================================
// String to number conversions
// =========================================================================
str_to_u64 :: proc(s: string) -> u64 {
    val, _ := strconv.parse_u64(s)
    return val
}
str_to_i64 :: proc(s: string) -> i64 {
    val, _ := strconv.parse_i64(s)
    return val
}
str_to_f32 :: proc(s: string) -> f32 {
    f, _ := strconv.parse_f64(s)
    return f32(f)
}
str_to_f64 :: proc(s: string) -> f64 {
    f, _ := strconv.parse_f64(s)
    return f
}
i64_to_str :: proc(value: i64) -> string {
    buf: [64]u8
    n := fmt.bprintf(buf[:], "%d", value)
    return strings.clone(string(buf[:n]), context.allocator)
}
u64_to_str :: proc(value: u64) -> string {
    buf: [64]u8
    n := fmt.bprintf(buf[:], "%d", value)
    return strings.clone(string(buf[:n]), context.allocator)
}

// =========================================================================
// UTF-8 manipulation
// =========================================================================
utf8_strlen :: proc(str: string) -> isize {
    return isize(utf8.rune_count(str))
}
utf8_strnlen :: proc(str: string, max_len: isize) -> isize {
    return isize(utf8.rune_count(str[:max_len]))
}

utf8_to_ucs2 :: proc(str: string) -> []u16 {
// Not implemented (Odin uses runes instead of UCS-2)
    return nil
}
ucs2_to_utf8 :: proc(data: []u16) -> string {
// Not implemented
    return ""
}
utf8_to_ucs2_buf :: proc(str: u8) -> []u16 {
// Not implemented
    return nil
}
ucs2_to_utf8_buf :: proc(data: []u16) -> string {
// Not implemented
    return ""
}

utf8_decode :: proc(str: string) -> (Rune, int) {
    return utf8.decode_rune(str)
}
utf8_codepoint_size :: proc(str: string) -> int {
    _, sz := utf8.decode_rune(str)
    return sz
}
utf8_encode_rune :: proc(buf: [4]u8, r: Rune) -> int {
    return utf8.encode_rune(r, buf[:])
}

// =========================================================================
// gbString type (dynamic string)
// =========================================================================
gbString :: struct {
    data: []u8,
    alloc: gbAllocator,
}

make_string_reserve :: proc(capacity: isize, alloc := context.allocator) -> gbString {
    return gbString{ alloc = alloc, data = make([]u8, 0, capacity, alloc) }
}

make_string :: proc(str: string, alloc := context.allocator) -> gbString {
    return gbString{ alloc = alloc, data = strings.clone_bytes(transmute([]u8)str, alloc) }
}

make_string_length :: proc(data: rawptr, len: isize, alloc := context.allocator) -> gbString {
    buf := make([]u8, len, alloc)
    if data != nil {
        mem.copy(raw_data(buf), data, int(len))
    }
    return gbString{ alloc = alloc, data = buf }
}

string_free :: proc(s: gbString) {
    delete(s.data)
}

string_duplicate :: proc(s: gbString, alloc := context.allocator) -> gbString {
    return make_string(string(s.data), alloc)
}

string_length :: proc(s: gbString) -> isize {
    return isize(len(s.data))
}

string_capacity :: proc(s: gbString) -> isize {
    return isize(cap(s.data))
}

string_available_space :: proc(s: gbString) -> isize {
    return string_capacity(s) - string_length(s)
}

string_clear :: proc(s: ^gbString) {
    s.data = s.data[:0]
}

string_append :: proc(s: ^gbString, other: gbString) {
    string_append_length(s, other.data)
}

string_append_length :: proc(s: ^gbString, other: []u8) {
    if len(other) == 0 do return
    old_len := len(s.data)
    new_len := old_len + len(other)
    if new_len > cap(s.data) {
        new_cap := max(cap(s.data) * 2, new_len)
        reserve := make([]u8, new_cap, s.alloc)
        copy(reserve, s.data)
        delete(s.data)
        s.data = reserve
    }
    s.data = s.data[:new_len]
    copy(s.data[old_len:], other)
}

string_appendc :: proc(s: ^gbString, cstr: string) {
    string_append_length(s, transmute([]u8)cstr)
}

string_append_rune :: proc(s: ^gbString, r: Rune) {
    buf: [4]u8
    n := utf8.encode_rune(r, buf[:])
    string_append_length(s, buf[:n])
}

string_append_fmt :: proc(s: ^gbString, format: string, args: ..any) {
// Using temporary buffer for formatting then append
    str := fmt.tprintf(format, ..args)
    bytes := make([]u8, len(str), s.alloc)
    copy(bytes, str)
    string_append_length(s, bytes)
}

string_set :: proc(s: ^gbString, cstr: string) {
    s.data = strings.clone_bytes(transmute([]u8)cstr, s.alloc)
}

string_make_space_for :: proc(s: ^gbString, add_len: isize) {
    required := string_length(s^) + add_len
    if required <= string_capacity(s^) do return
    new_cap := max(cap(s.data) * 2, required)
    reserve := make([]u8, new_cap, s.alloc)
    copy(reserve, s.data)
    delete(s.data)
    s.data = reserve[:len(s.data)]
}

string_allocation_size :: proc(s: gbString) -> isize {
    return isize(cap(s.data))
}

string_are_equal :: proc(a, b: gbString) -> b8 {
    return string(a.data) == string(b.data)
}

string_trim :: proc(s: gbString, cutset: string) -> gbString {
    trimmed := strings.trim(string(s.data), cutset)
    return make_string(trimmed, s.alloc)
}

string_trim_space :: proc(s: gbString) -> gbString {
    return string_trim(s, " \t\r\n\v\f")
}

// =========================================================================
// Hashing
// =========================================================================
adler32 :: proc(data: []u8) -> u32 {
    MOD_ALDER : u32 : 65521
    a : u32 = 1
    b : u32 = 0
    off: isize
    chunk_size := isize(5552)
    for off < isize(len(data)) {
        end := min(off + chunk_size, isize(len(data)))
        for i := off; i + 7 < end; i += 8 {
            a += u32(data[i]); b += a
            a += u32(data[i + 1]); b += a
            a += u32(data[i + 2]); b += a
            a += u32(data[i + 3]); b += a
            a += u32(data[i + 4]); b += a
            a += u32(data[i + 5]); b += a
            a += u32(data[i + 6]); b += a
            a += u32(data[i + 7]); b += a
        }
        for i := off; i < end; i += 1 {
            a += u32(data[i]); b += a
        }
        a %= MOD_ALDER
        b %= MOD_ALDER
        off = end
        chunk_size = 5552
    }
    return (b << 16) | a
}

// CRC32/64 tables from the original would be huge; we provide stubs.
crc32 :: proc(data: []u8) -> u32 {
// Not implemented
    return 0
}
crc64 :: proc(data: []u8) -> u64 {
// Not implemented
    return 0
}
fnv32 :: proc(data: []u8) -> u32 {
// Not implemented
    return 0
}
fnv64 :: proc(data: []u8) -> u64 {
// Not implemented
    return 0
}
fnv32a :: proc(data: []u8) -> u32 {
// Not implemented
    return 0
}
fnv64a :: proc(data: []u8) -> u64 {
// Not implemented
    return 0
}

murmur32 :: proc(data: []u8) -> u32 {
    return murmur32_seed(data, 0x9747_b28c)
}
murmur64 :: proc(data: []u8) -> u64 {
    return murmur64_seed(data, 0x9747_b28c)
}

murmur32_seed :: proc(data: []u8, seed: u32) -> u32 {
    c1 : u32 = 0xcc9e_2d51
    c2 : u32 = 0x1b87_3593
    r1 : u32 = 15
    r2 : u32 = 13
    m :  u32 = 5
    n :  u32 = 0xe654_6b64

    hash := seed
    nblocks := len(data) / 4

    for i := 0; i < nblocks; i += 1 {
        k := (^u32)(&data[i * 4])^
        k *= c1
        k = (k << r1) | (k >> (32 - r1))
        k *= c2
        hash ^ = k
        hash = ((hash << r2) | (hash >> (32 - r2))) * m + n
    }

    tail := data[nblocks * 4:]
    k1: u32
    switch len(tail) {
    case 3:
        k1 ^ = u32(tail[2]) << 16
        fallthrough
    case 2:
        k1 ^ = u32(tail[1]) << 8
        fallthrough
    case 1:
        k1 ^ = u32(tail[0])
        k1 *= c1
        k1 = (k1 << r1) | (k1 >> (32 - r1))
        k1 *= c2
        hash ^ = k1
    }

    hash ^ = u32(len(data))
    hash ^ = (hash >> 16)
    hash *= 0x85eb_ca6b
    hash ^ = (hash >> 13)
    hash *= 0xc2b2_ae35
    hash ^ = (hash >> 16)
    return hash
}

murmur64_seed :: proc(data: []u8, seed: u64) -> u64 {
    m : u64 = 0xc6a4_a793_5bd1_e995
    r : i32 = 47

    h := seed ^ (u64(len(data)) * m)
    nblocks := len(data) / 8
    for i := 0; i < nblocks; i += 1 {
        k := (^u64)(&data[i * 8])^
        k *= m
        k ^ = k >> u32(r)
        k *= m
        h ^ = k
        h *= m
    }

    tail := data[nblocks * 8:]
    switch len(tail) {
    case 7: h ^ = u64(tail[6]) << 48; fallthrough
    case 6: h ^ = u64(tail[5]) << 40; fallthrough
    case 5: h ^ = u64(tail[4]) << 32; fallthrough
    case 4: h ^ = u64(tail[3]) << 24; fallthrough
    case 3: h ^ = u64(tail[2]) << 16; fallthrough
    case 2: h ^ = u64(tail[1]) << 8;  fallthrough
    case 1: h ^ = u64(tail[0]); h *= m
    }

    h ^ = h >> u32(r)
    h *= m
    h ^ = h >> u32(r)
    return h
}

// =========================================================================
// File operations
// =========================================================================
gbFile :: struct {
    handle: os.Handle,
    filename: string,
    position: i64,
    last_write_time: time.Time,
}
gbFileMode :: u32
gbSeekWhenceType :: i32

gbDefaultFileOperations := struct{
}{ } // placeholder

file_get_standard :: proc(std: i32) -> ^gbFile {
    switch std {
    case 0: return &gbFile{ handle = os.stdin,  filename = "<stdin>"  }
    case 1: return &gbFile{ handle = os.stdout, filename = "<stdout>" }
    case 2: return &gbFile{ handle = os.stderr, filename = "<stderr>" }
    }
    return nil
}

file_create :: proc(filename: string) -> ^gbFile {
    return file_open_mode(filename, GB_FILE_MODE_WRITE | GB_FILE_MODE_RW)
}

file_open :: proc(filename: string) -> ^gbFile {
    return file_open_mode(filename, GB_FILE_MODE_READ)
}

file_open_mode :: proc(filename: string, mode: gbFileMode) -> ^gbFile {
    flags: int
    if mode & GB_FILE_MODE_READ != 0  { flags |= os.O_RDONLY }
    if mode & GB_FILE_MODE_WRITE != 0 { flags |= os.O_WRONLY }
    if mode & GB_FILE_MODE_RW != 0    { flags = (flags & ~os.O_RDONLY) | os.O_RDWR }
    if mode & GB_FILE_MODE_APPEND != 0 { flags |= os.O_APPEND | os.O_CREATE }
    if flags == 0 { flags = os.O_RDONLY }

    handle, err := os.open(filename, flags, 0)
    if err != nil do return nil
    f := new(gbFile, context.allocator)
    f.handle = handle
    f.filename = strings.clone(filename, context.allocator)
    f.position = 0
    if mode & GB_FILE_MODE_APPEND != 0 {
        f.position, _ = os.seek(handle, 0, os.SEEK_END)
    }
    return f
}

file_new :: proc(fd: rawptr, ops: any, filename: string) -> ^gbFile {
    f := new(gbFile, context.allocator)
    f.handle = cast(os.Handle)fd
    f.filename = strings.clone(filename, context.allocator)
    return f
}

file_close :: proc(f: ^gbFile) {
    if f != nil && f.handle != os.INVALID_HANDLE {
        os.close(f.handle)
        delete(f.filename)
        free(f)
    }
}

file_read_at_check :: proc(f: ^gbFile, buffer: []u8, offset: i64) -> (bool, int) {
    n, err := os.read_at(f.handle, buffer, offset)
    return err == nil, n
}

file_write_at_check :: proc(f: ^gbFile, data: []u8, offset: i64) -> (bool, int) {
    n, err := os.write_at(f.handle, data, offset)
    return err == nil, n
}

file_read_at :: proc(f: ^gbFile, buffer: []u8, offset: i64) -> bool {
    ok, _ := file_read_at_check(f, buffer, offset)
    return ok
}

file_write_at :: proc(f: ^gbFile, data: []u8, offset: i64) -> bool {
    ok, _ := file_write_at_check(f, data, offset)
    return ok
}

file_seek :: proc(f: ^gbFile, offset: i64) -> i64 {
    f.position = offset
    return f.position
}
file_seek_to_end :: proc(f: ^gbFile) -> i64 {
    f.position, _ = os.seek(f.handle, 0, os.SEEK_END)
    return f.position
}
file_skip :: proc(f: ^gbFile, bytes: i64) -> i64 {
    f.position += bytes
    return f.position
}
file_tell :: proc(f: ^gbFile) -> i64 {
    return f.position
}
file_read :: proc(f: ^gbFile, buffer: []u8) -> bool {
    n, err := os.read_at(f.handle, buffer, f.position)
    if err != nil do return false
    f.position += i64(n)
    return true
}
file_write :: proc(f: ^gbFile, data: []u8) -> bool {
    n, err := os.write_at(f.handle, data, f.position)
    if err != nil do return false
    f.position += i64(n)
    return true
}
file_size :: proc(f: ^gbFile) -> i64 {
    if f == nil do return 0
    sz, _ := os.file_size(f.handle)
    return i64(sz)
}
file_name :: proc(f: ^gbFile) -> string {
    return f.filename
}
file_truncate :: proc(f: ^gbFile, size: i64) -> bool {
    return os.truncate(f.handle, size) == nil
}
file_has_changed :: proc(f: ^gbFile) -> bool {
    info, err := os.stat(f.filename)
    if err != nil do return false
    changed := info.modification_time != f.last_write_time
    f.last_write_time = info.modification_time
    return changed
}

file_exists :: proc(filename: string) -> bool {
    return os.exists(filename)
}
file_last_write_time :: proc(filename: string) -> time.Time {
    info, err := os.stat(filename)
    if err != nil do return time.Time{}
    return info.modification_time
}
file_copy :: proc(existing, new: string, fail_if_exists: bool) -> bool {
    if fail_if_exists && os.exists(new) do return false
    data, ok := os.read_entire_file(existing, context.temp_allocator)
    if !ok do return false
    return os.write_entire_file(new, data, false)
}
file_move :: proc(existing, new: string) -> bool {
    return os.rename(existing, new) == nil
}
file_remove :: proc(filename: string) -> bool {
    return os.remove(filename) == nil
}

file_read_contents :: proc(filename: string, zero_terminate: bool, alloc := context.allocator) -> ([]u8, bool) {
    data, ok := os.read_entire_file(filename, alloc)
    if !ok do return nil, false
    if zero_terminate {
        new_data := make([]u8, len(data) + 1, alloc)
        copy(new_data, data)
        new_data[len(data)] = 0
        delete(data)
        return new_data, true
    }
    return data, true
}
file_free_contents :: proc(data: []u8) {
    delete(data)
}

// =========================================================================
// Path utilities
// =========================================================================
path_is_absolute :: proc(path: string) -> bool {
// Windows: starts with drive letter or '\', Unix: starts with '/'
    return len(path) > 0 && (path[0] == '/' || (len(path) >= 3 && path[1] == ':'))
}
path_is_relative :: proc(path: string) -> bool {
    return !path_is_absolute(path)
}
path_is_root :: proc(path: string) -> bool {
    return path_is_absolute(path) && len(strings.trim_right(path, "/\\")) == 0
}
path_base_name :: proc(path: string) -> string {
    return strings.filepath_base(path)
}
path_extension :: proc(path: string) -> string {
    return strings.filepath_ext(path)
}
path_get_full_name :: proc(path: string, alloc := context.allocator) -> string {
    abs, err := os.absolute(path, alloc)
    if err != nil do return path
    return abs
}

// =========================================================================
// Printf family
// =========================================================================
printf :: proc(format: string, args: ..any) {
    fmt.printf(format, ..args)
}
printf_err :: proc(format: string, args: ..any) {
    fmt.eprintf(format, ..args)
}
fprintf :: proc(f: ^gbFile, format: string, args: ..any) {
    if f == nil do return
    s := fmt.tprintf(format, ..args)
    file_write(f, transmute([]u8)s)
}
bprintf :: proc(format: string, args: ..any) -> string {
    return fmt.tprintf(format, ..args)
}
snprintf :: proc(buf: []u8, format: string, args: ..any) -> int {
    s := fmt.tprintf(format, ..args)
    n := copy(buf, s)
    if n < len(buf) {
        buf[n] = 0
    }
    return n
}

// =========================================================================
// Time
// =========================================================================
rdtsc :: proc() -> u64 {
// x86-specific, not available
    return 0
}
time_now :: proc() -> f64 {
    now := time.now()
    // Convert to seconds as a double
    return f64(now._nsec) * 1e-9 // approximate
}
utc_time_now :: proc() -> u64 {
    return u64(time.now().unix_nano() / 100) // convert to microseconds as in original?
}
sleep_ms :: proc(ms: u32) {
    time.sleep(time.Duration(ms) * time.Millisecond)
}
exit :: proc(code: u32) {
    os.exit(int(code))
}

// =========================================================================
// Environment
// =========================================================================
get_env :: proc(name: string, alloc := context.allocator) -> string {
    val, found := os.lookup_env(name)
    if found {
        return strings.clone(val, alloc)
    }
    return ""
}
set_env :: proc(name, value: string) {
    os.set_env(name, value)
}
unset_env :: proc(name: string) {
    os.unset_env(name)
}

// =========================================================================
// Endian swaps
// =========================================================================
endian_swap16 :: proc(i: u16) -> u16 {
    return bits.byte_swap(i)
}
endian_swap32 :: proc(i: u32) -> u32 {
    return bits.byte_swap(i)
}
endian_swap64 :: proc(i: u64) -> u64 {
    return bits.byte_swap(i)
}

// =========================================================================
// Bit count
// =========================================================================
count_set_bits :: proc(mask: u64) -> isize {
    return isize(bits.count_ones(mask))
}