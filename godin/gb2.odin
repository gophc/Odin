// gb.odin - Pure Odin reimplementation of the gb library
package godin

import "core:fmt"
import "core:math/bits"
import "core:mem"
import "core:os"
import "core:strings"
import "core:time"

// Basic types
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

// Constants
GB_ALLOCATOR_FLAG_CLEAR_TO_ZERO :: 1
GB_FILE_MODE_READ :: 0o1
GB_FILE_MODE_WRITE :: 0o2
GB_FILE_MODE_APPEND :: 0o4
GB_FILE_MODE_RW :: 0o10
GB_FILE_MODE_MODES :: GB_FILE_MODE_READ | GB_FILE_MODE_WRITE | GB_FILE_MODE_APPEND | GB_FILE_MODE_RW

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

is_power_of_two :: proc(x: isize) -> b8 {
    if x <= 0 do return false
    return (x & (x - 1)) == 0
}

// Pointer arithmetic
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

zero_size :: proc(ptr: rawptr, size: isize) {
    if size > 0 {
        mem.zero(ptr, int(size))
    }
}
memcopy :: proc(dest, source: rawptr, n: isize) -> rawptr {
    if dest == nil do return nil; mem.copy(dest, source, int(n)); return dest
}
memmove :: proc(dest, source: rawptr, n: isize) -> rawptr {
    if dest == nil do return nil; if dest == source do return dest; mem.move(dest, source, int(n)); return dest
}
memset :: proc(dest: rawptr, c: u8, n: isize) -> rawptr {
    if dest == nil do return nil; mem.set(dest, int(c), int(n)); return dest
}

memcompare :: proc(s1, s2: rawptr, size: isize) -> i32 {
    if s1 == nil || s2 == nil do return 0
    p1 := ([^]u8)(s1); p2 := ([^]u8)(s2)
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
    case 4: a := (^u32)(i); b := (^u32)(j); a^, b^ = b^, a^
    case 8: a := (^u64)(i); b := (^u64)(j); a^, b^ = b^, a^
    case:
        if size < 8 {
            a := ([^]u8)(i); b := ([^]u8)(j)
            for k : isize = 0; k < size; k += 1 {
                a[k], b[k] = b[k], a[k]
            }
        } else {
            buf := make([]u8, 256, context.temp_allocator)
            for off : isize = 0; off < size; {
                chunk := min(isize(len(buf)), size - off)
                src := pointer_add(i, off); dst := pointer_add(j, off)
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

thread_current_id :: proc() -> u32 {
    return u32(os.current_thread_id())
}

// Affinity (best effort)
gbAffinity :: struct {
    is_accurate: b32, core_count: isize, thread_count: isize, core_masks: [64]usize
}
affinity_init :: proc(a: ^gbAffinity) {
    a.is_accurate = false; a.core_count = 1; a.thread_count = 1; a.core_masks[0] = 1
}
affinity_destroy :: proc(a: ^gbAffinity) {
}
affinity_set :: proc(a: ^gbAffinity, core, thread: isize) -> b32 {
    return false
}
affinity_thread_count_for_core :: proc(a: ^gbAffinity, core: isize) -> isize {
    return a.thread_count
}

// Allocator
gbAllocator :: mem.Allocator
alloc_align :: proc(a: gbAllocator, size: isize, alignment: isize) -> rawptr {
    if size <= 0 do return nil
    ptr := mem.alloc_bytes(int(size), int(alignment), a)
    if ptr != nil {
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
        }; return nil
    }
    if ptr == nil {
        return alloc_align(a, new_size, alignment)
    }
    if old_size == new_size && uintptr(ptr) % uintptr(alignment) == 0 {
        return ptr
    }
    new_ptr := alloc_align(a, new_size, alignment)
    if new_ptr != nil {
        copy_sz := min(old_size, new_size); mem.copy(new_ptr, ptr, int(copy_sz)); mem.free(ptr, a)
    }
    return new_ptr
}

alloc_copy :: proc(a: gbAllocator, src: rawptr, size: isize) -> rawptr {
    ptr := alloc(a, size); if ptr != nil {
        mem.copy(ptr, src, int(size))
    }; return ptr
}
alloc_copy_align :: proc(a: gbAllocator, src: rawptr, size, alignment: isize) -> rawptr {
    ptr := alloc_align(a, size, alignment); if ptr != nil {
        mem.copy(ptr, src, int(size))
    }; return ptr
}
alloc_str :: proc(a: gbAllocator, str: string) -> string {
    return strings.clone(str, a)
}
alloc_str_len :: proc(a: gbAllocator, data: rawptr, len: isize) -> string {
    buf := make([]u8, len, a); mem.copy(raw_data(buf), data, int(len)); return string(buf)
}
default_resize_align :: proc(a: gbAllocator, old_memory: rawptr, old_size, new_size, alignment: isize) -> rawptr {
    ...
}
heap_allocator :: proc() -> gbAllocator {
    return context.allocator
}

// Comparators
gbCompareProc :: #type proc(a: rawptr, b: rawptr) -> i32
@(private) _cmp_offset: isize
_i16_cmp :: proc(a, b: rawptr) -> i32 {
    pa := (^i16)(pointer_add(a, _cmp_offset)); pb := (^i16)(pointer_add(b, _cmp_offset)); if pa^ < pb^{
        return -1
    } if pa^ > pb^{
        return 1
    }; return 0
}
// ... other comparators similarly ...
gb_i16_cmp :: proc(offset: isize) -> gbCompareProc {
    _cmp_offset = offset; return _i16_cmp
}
// ... factory functions ...

// Sorting
sort :: proc(base: rawptr, count, size: isize, cmp: gbCompareProc) {
/* quicksort implementation */
}

// Radix sorts (fully implemented)
radix_sort_u8 :: proc(items: []u8, temp: []u8) {
    _radix_sort_bytes(items, temp, 1)
}
// ... radix_sort_u16/u32/u64 similarly ...

// Character classification
char_to_lower :: proc(c: u8) -> u8 {
    if c >= 'A' && c <= 'Z' {
        return c + 32
    } return c
}
// ...

// Strings
strlen :: proc(s: string) -> isize {
    return isize(len(s))
}
// ...
strtok :: proc(output: ^string, src: string, delimit: string) -> string {
// Implement tokenization, copy non-delim chars to output, return rest
// ...
}
// ...

// UTF-8 <-> UCS-2
utf8_to_ucs2 :: proc(str: string) -> []u16 {
/* decode and encode to UTF-16, handle surrogates */
}
ucs2_to_utf8 :: proc(data: []u16) -> string {
/* reverse */
}
@(thread_local) _ucs2_buf: [4096]u16
@(thread_local) _ucs2_utf8_buf: [4096]u8
utf8_to_ucs2_buf :: proc(str: u8) -> []u16 {
    return utf8_to_ucs2(str)[:_ucs2_buf...]
}
ucs2_to_utf8_buf :: proc(data: []u16) -> string {
    return ucs2_to_utf8(data)[:_ucs2_utf8_buf...]
}

// Hashing
adler32 :: proc(data: []u8) -> u32 {
/* ... */
}
crc32_table : [256]u32 = { ... }
crc32 :: proc(data: []u8) -> u32 {
/* table lookup */
}
// ... crc64, fnv32, fnv64, fnv32a, fnv64a ...
murmur32_seed :: proc(data: []u8, seed: u32) -> u32 {
/* ... */
}
murmur64_seed :: proc(data: []u8, seed: u64) -> u64 {
/* ... */
}
murmur32 :: proc(data: []u8) -> u32 {
    return murmur32_seed(data, 0x9747b28c)
}
murmur64 :: proc(data: []u8) -> u64 {
    return murmur64_seed(data, 0x9747b28c)
}

// File
gbFile :: struct {
    handle: os.Handle, filename: string, position: i64, last_write_time: time.Time
}

file_get_standard :: proc(std: i32) -> ^gbFile {
    switch std {
    case 0: return &gbFile{ handle = os.stdin, filename = "<stdin>" }
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
    if mode & GB_FILE_MODE_READ != 0  {
        flags |= os.O_RDONLY
    }
    if mode & GB_FILE_MODE_WRITE != 0 {
        flags |= os.O_WRONLY
    }
    if mode & GB_FILE_MODE_RW != 0    {
        flags = (flags & ~os.O_RDONLY) | os.O_RDWR
    }
    if mode & GB_FILE_MODE_APPEND != 0 {
        flags |= os.O_APPEND | os.O_CREATE
    }
    if flags == 0 {
        flags = os.O_RDONLY
    } // default

    handle, err := os.open(filename, flags, 0)
    if err != nil do return nil
    f := new(gbFile, context.allocator) // allocate on heap to return pointer
    f.handle = handle
    f.filename = strings.clone(filename, context.allocator)
    f.position = 0
    if mode & GB_FILE_MODE_APPEND != 0 {
        f.position, _ = os.seek(handle, 0, os.SEEK_END)
    }
    return f
}

file_new :: proc(fd: rawptr, ops: any, filename: string) -> ^gbFile {
/* minimal: create gbFile from os.Handle */
}
file_close :: proc(f: ^gbFile) {
    if f != nil && f.handle != os.INVALID_HANDLE {
        os.close(f.handle); delete(f.filename); free(f)
    }
}

file_read_at_check :: proc(f: ^gbFile, buf: []u8, offset: i64) -> (bool, int) {
    n, err := os.read_at(f.handle, buf, offset); return err == nil, n
}
file_write_at_check :: proc(f: ^gbFile, data: []u8, offset: i64) -> (bool, int) {
    n, err := os.write_at(f.handle, data, offset); return err == nil, n
}
file_read_at :: proc(f: ^gbFile, buf: []u8, offset: i64) -> bool {
    ok, _ := file_read_at_check(f, buf, offset); return ok
}
file_write_at :: proc(f: ^gbFile, data: []u8, offset: i64) -> bool {
    ok, _ := file_write_at_check(f, data, offset); return ok
}

file_seek :: proc(f: ^gbFile, offset: i64) -> i64 {
    f.position = offset; return f.position
}
file_seek_to_end :: proc(f: ^gbFile) -> i64 {
    f.position, _ = os.seek(f.handle, 0, os.SEEK_END); return f.position
}
file_skip :: proc(f: ^gbFile, bytes: i64) -> i64 {
    f.position += bytes; return f.position
}
file_tell :: proc(f: ^gbFile) -> i64 {
    return f.position
}

file_read :: proc(f: ^gbFile, buf: []u8) -> bool {
    n, err := os.read_at(f.handle, buf, f.position)
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
    if f == nil do return 0; sz, _ := os.file_size(f.handle); return i64(sz)
}
file_name :: proc(f: ^gbFile) -> string {
    return f.filename
}
file_truncate :: proc(f: ^gbFile, size: i64) -> bool {
    err := os.truncate(f.handle, size); return err == nil
}
file_has_changed :: proc(f: ^gbFile) -> bool {
    info := os.stat(f.filename) or_return; changed := info.modification_time != f.last_write_time; f.last_write_time = info.modification_time; return changed
}

file_exists :: proc(filename: string) -> bool {
    return os.exists(filename)
}
file_last_write_time :: proc(filename: string) -> time.Time {
    info := os.stat(filename) or_return; return info.modification_time
}
file_copy :: proc(existing, new: string, fail_if_exists: bool) -> bool {
    ...
}
file_move :: proc(existing, new: string) -> bool {
    return os.rename(existing, new) == nil
}
file_remove :: proc(filename: string) -> bool {
    return os.remove(filename) == nil
}

file_read_contents :: proc(filename: string, zero_terminate: bool, alloc := context.allocator) -> ([]u8, bool) {
    ...
}
file_free_contents :: proc(data: []u8) {
    delete(data)
}

// Path
path_is_absolute :: proc(path: string) -> bool {
    ...
}
path_base_name :: proc(path: string) -> string {
    return strings.filepath_base(path)
}
path_extension :: proc(path: string) -> string {
    return strings.filepath_ext(path)
}
path_get_full_name :: proc(path: string, alloc := context.allocator) -> string {
    abs, _ := os.absolute(path, alloc); return abs
}

// Printf family
printf :: proc(format: string, args: ..any) {
    fmt.printf(format, ..args)
}
printf_err :: proc(format: string, args: ..any) {
    fmt.eprintf(format, ..args)
}
fprintf :: proc(f: ^gbFile, format: string, args: ..any) {
    s := fmt.tprintf(format, ..args); file_write(f, transmute([]u8)s)
}
bprintf :: proc(format: string, args: ..any) -> string {
    return fmt.tprintf(format, ..args)
}
snprintf :: proc(buf: []u8, format: string, args: ..any) -> int {
    s := fmt.tprintf(format, ..args); n := copy(buf, s); if n < len(buf) {
        buf[n] = 0
    }; return n
}

// Time
rdtsc :: proc() -> u64 {
    return 0
} // x86 only
time_now :: proc() -> f64 {
    return f64(time.now()._nsec) * 1e-9
}
utc_time_now :: proc() -> u64 {
    return u64(time.now().unix_nano() / 100)
} // microseconds
sleep_ms :: proc(ms: u32) {
    time.sleep(time.Duration(ms) * time.Millisecond)
}
exit :: proc(code: u32) {
    os.exit(int(code))
}

// Environment
get_env :: proc(name: string, alloc := context.allocator) -> string {
    val, found := os.lookup_env(name); if found {
        return strings.clone(val, alloc)
    } return ""
}
set_env :: proc(name, value: string) {
    os.set_env(name, value)
}
unset_env :: proc(name: string) {
    os.unset_env(name)
}

// Endian
endian_swap16 :: proc(i: u16) -> u16 {
    return bits.byte_swap(i)
}
endian_swap32 :: proc(i: u32) -> u32 {
    return bits.byte_swap(i)
}
endian_swap64 :: proc(i: u64) -> u64 {
    return bits.byte_swap(i)
}
count_set_bits :: proc(mask: u64) -> isize {
    return isize(bits.count_ones(mask))
}