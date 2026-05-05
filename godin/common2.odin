// common.odin - godin common library
package godin

import "core:fmt"
import "core:mem"
import "core:os"
import "core:path/filepath"
import "core:strings"
import "core:sync"
import "core:thread"
import "core:unicode"

// ——————————————————————————————————————————————
// Basic types (Odin's built‑in `int`, `uint`, etc. already provide equivalent sizes)
// ——————————————————————————————————————————————

// Odin already has i32, i64, isize, usize, u8, u16, u32, u64, f32, f64, uintptr, bool, etc.
// We'll use them directly.

Rune :: rune
String :: string
String16 :: []u16

// ——————————————————————————————————————————————
// Global module path (placeholder)
// ——————————————————————————————————————————————

global_module_path: string
global_module_path_set := false

// ——————————————————————————————————————————————
// Power of two helpers (same logic, using Odin's bit operations)
// ——————————————————————————————————————————————

next_pow2_32 :: proc(n: u32) -> u32 {
    if n == 0 do return 0
    n -= 1
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    return n + 1
}

next_pow2_64 :: proc(n: u64) -> u64 {
    if n == 0 do return 0
    n -= 1
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    n |= n >> 32
    return n + 1
}

next_pow2_int :: proc(n: int) -> int {
// Odin's int is 64‑bit on our target, delegate to u64 version.
    return int(next_pow2_64(u64(n)))
}

is_power_of_two_int :: proc(x: int) -> bool {
    return x > 0 && (x & (x - 1)) == 0
}

// ——————————————————————————————————————————————
// String helpers (using Odin's string / rune facilities)
// ——————————————————————————————————————————————

// make a string from C‑style pointer + length (not needed in pure Odin, but kept for similarity)
make_string :: proc(data: []u8) -> string {
    return string(data)
}

make_string_c :: proc(cstr: cstring) -> string {
    return string(cstr)
}

// substring (using slicing)
substring :: proc(s: string, lo, hi: int) -> string {
// Odin slices already do bounds checking when compiling with -o:speed, but we replicate the assert
    assert(lo <= hi && hi <= len(s), "substring out of bounds")
    return s[lo:hi]
}

str_eq :: proc(a, b: string) -> bool {
    return a == b
}

str_ne :: proc(a, b: string) -> bool {
    return a != b
}

str_lt :: proc(a, b: string) -> bool {
    return a < b
}

// … (gt, le, ge follow similarly)

str_eq_ignore_case :: proc(a, b: string) -> bool {
    if len(a) != len(b) do return false
    return strings.equal_fold(a, b)   // Odin provides case‑insensitive compare
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

string_index_byte :: proc(s: string, c: u8) -> int {
    return strings.index_byte(s, c)
}

string_contains_string :: proc(haystack, needle: string) -> bool {
    return strings.contains(haystack, needle)
}

// string_split_iterator not directly needed; Odin has `strings.split_iterator`.
// We keep the function as a simple wrapper.
String_Iterator :: struct {
    str: string,
    pos: int,
}

string_split_iterator :: proc(it: ^String_Iterator, sep: rune) -> string {
    if it.pos >= len(it.str) do return ""
    start := it.pos
    for i := it.pos; i < len(it.str); i += 1 {
        if rune(it.str[i]) == sep {
            res := it.str[start:i]
            it.pos = i + 1
            return res
        }
    }
    it.pos = len(it.str)
    return it.str[start:]
}

// path_extension, filename_from_path, etc. – use filepath package in real code.
// We reproduce the simplest forms for compatibility.
is_separator :: proc(c: u8) -> bool {
    return c == '/' || c == '\\'
}

filename_without_directory :: proc(s: string) -> string {
    return filepath.base(s)
}

// ——————————————————————————————————————————————
// Dynamic array / slice – Odin built‑in, no need for custom Array/Slice templates.
// Old API wrappers can be added if needed, but we skip them.
// ——————————————————————————————————————————————

// ——————————————————————————————————————————————
// Threading primitives (using Odin's sync and OS threads)
// ——————————————————————————————————————————————

// Mutexes – Odin's sync.Mutex is a blocking mutex (SRWLock on Windows)
Mutex :: sync.Mutex
Recursive_Mutex :: sync.Recursive_Mutex
Rw_Mutex :: sync.RW_Mutex

// Semaphore using atomic (simplified)
Semaphore :: struct {
    count: atomic.Int,
}

semaphore_post :: proc(s: ^Semaphore, n: i32 = 1) {
// Odin's atomic add
    prev := atomic.add(&s.count, n)
// Wake waiting threads – not implemented here, would need futex/cond
// TODO: implement futex‑like wake
}

semaphore_wait :: proc(s: ^Semaphore) {
    for {
        c := atomic.load(&s.count)
        if c > 0 {
            if atomic.compare_and_swap(&s.count, c, c - 1) == c {
                return
            }
        } else {
        // TODO: wait (futex / sync.Cond)
            thread.yield()
        }
    }
}

// Parker (from C++ code) not directly translated, use sync.Cond if needed.
// We skip detailed futex implementation.

// Thread representation
Thread :: struct {
    sys_handle: rawptr, // OS handle
    idx: int,
    stack_size: int,
    pool: ^ThreadPool,
    permanent_arena: ^Arena,
    temporary_arena: ^Arena,
}

ThreadPool :: struct {
    threads: [dynamic]Thread,
    running: atomic.Bool,
    tasks_available: atomic.Int, // used as a futex replacement
    tasks_left: atomic.Int,
}

current_thread_index :: proc() -> int {
// Not easily exposed in Odin without TLS, skip.
// TODO: implement TLS getter.
    return 0
}

get_current_thread :: proc() -> ^Thread {
// TODO: thread‑local storage
    return nil
}

thread_init :: proc(t: ^Thread, idx: int, pool: ^ThreadPool) {
    t.idx = idx
    t.pool = pool
// Arenas will be created later
}

thread_init_and_start :: proc(t: ^Thread, idx: int, pool: ^ThreadPool) {
    thread_init(t, idx, pool)
// Create OS thread – not implemented
// t.sys_handle = os.thread_create(...)
}

thread_join_and_destroy :: proc(t: ^Thread) {
    if t.sys_handle != nil {
    // os.thread_join(t.sys_handle)
    }
}

// Thread pool functions – minimal skeleton
thread_pool_init :: proc(pool: ^ThreadPool, worker_count: int, name: string) {
    pool.threads = make([dynamic]Thread, 0, worker_count + 1, context.allocator)
    atomic.store(&pool.running, true)
    // create main thread dummy
    append(&pool.threads, Thread{ idx = 0 })
    for i in 1 ..< worker_count + 1 {
    // TODO: start worker thread
    }
}

thread_pool_destroy :: proc(pool: ^ThreadPool) {
    atomic.store(&pool.running, false)
    // signal workers, join...
    delete(pool.threads)
}

thread_pool_add_task :: proc(pool: ^ThreadPool, proc_: rawptr, data: rawptr) -> bool {
// TODO: task queue insertion
    return false
}

thread_pool_wait :: proc(pool: ^ThreadPool) {
// TODO: wait until tasks_left == 0
}

// ——————————————————————————————————————————————
// Memory / Arena (simplified bump allocator using context.allocator)
// ——————————————————————————————————————————————

Arena :: struct {
    curr_block: ^MemoryBlock,
    minimum_block_size: int,
    temp_count: int,
}

MemoryBlock :: struct {
    prev: ^MemoryBlock,
    data: []u8,
    used: int,
}

DEFAULT_MINIMUM_BLOCK_SIZE :: 8 * mem.Megabyte

arena_alloc :: proc(a: ^Arena, size, alignment: int) -> rawptr {
// Odin's mem.Arena can be used, but we implement a basic one.
    if a.curr_block == nil || (a.curr_block.used + size > len(a.curr_block.data)) {
        block := new(MemoryBlock)
        block_size := max(size, a.minimum_block_size)
        block.data = make([]u8, block_size, context.allocator)
        block.prev = a.curr_block
        a.curr_block = block
    }
    block := a.curr_block
    // align
    ptr := (&block.data[0]) + uintptr(block.used)
    aligned_ptr := mem.align_forward(ptr, uintptr(alignment))
    offset := int(aligned_ptr - ptr)
    block.used += offset + size
    return rawptr(aligned_ptr)
}

arena_free_all :: proc(a: ^Arena) {
    for a.curr_block != nil {
        block := a.curr_block
        a.curr_block = block.prev
        delete(block.data)
        free(block)
    }
}

arena_temp_begin :: proc(a: ^Arena) -> Arena_Temp {
    temp := Arena_Temp{
        arena = a,
        block = a.curr_block,
        used  = a.curr_block.used if a.curr_block != nil else 0,
    }
    a.temp_count += 1
    return temp
}

arena_temp_end :: proc(temp: Arena_Temp) {
// similar logic to C++ code, using block chains and zeroing
// Not fully implemented, but sketch:
    arena := temp.arena
    assert(arena.temp_count > 0)
    arena.temp_count -= 1
    // if block still exists, roll back used
    if temp.block != nil {
        for arena.curr_block != temp.block {
        // free intermediate blocks
        }
        arena.curr_block.used = temp.used
    }
}

Arena_Temp :: struct {
    arena: ^Arena,
    block: ^MemoryBlock,
    used: int,
}

// Arena allocator procedurals
arena_allocator_proc :: proc(allocator_data: rawptr, type: mem.Allocator_Mode, size, alignment: int, old_memory: rawptr, old_size: int, flags: u64) -> rawptr {
    arena := cast(^Arena)allocator_data
    switch type {
    case .Alloc:
        return arena_alloc(arena, size, alignment)
    case .Free:
        return nil
    case .Resize:
        if size == 0 do return nil
        if size <= old_size do return old_memory
        new_mem := arena_alloc(arena, size, alignment)
        mem.copy(new_mem, old_memory, old_size)
        return new_mem
    case .Free_All:
        arena_free_all(arena)
        return nil
    }
    return nil
}

// Helper to turn Arena into a context allocator
arena_allocator :: proc(a: ^Arena) -> runtime.Allocator {
    return runtime.Allocator{ procedure = arena_allocator_proc, data = a }
}

// ——————————————————————————————————————————————
// Priority queue (generic using Odin's dynamic array and comparator)
// ——————————————————————————————————————————————

PriorityQueue :: struct($T: typeid) {
    queue: [dynamic]T,
    cmp: proc(a, b: T) -> bool, // true if a < b (for min‑heap)
}

priority_queue_push :: proc(pq: ^PriorityQueue($T), value: T) {
    append(&pq.queue, value)
    priority_queue_shift_up(pq, len(pq.queue) - 1)
}

priority_queue_pop :: proc(pq: ^PriorityQueue($T)) -> T {
    assert(len(pq.queue) > 0)
    n := len(pq.queue) - 1
    pq.queue[0], pq.queue[n] = pq.queue[n], pq.queue[0]
    priority_queue_shift_down(pq, 0, n)
    item := pop(&pq.queue)
    return item
}

priority_queue_shift_up :: proc(pq: ^PriorityQueue($T), j: int) {
    for j > 0 {
        i := (j - 1) / 2
        if !pq.cmp(pq.queue[j], pq.queue[i]) do break
        pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
        j = i
    }
}

priority_queue_shift_down :: proc(pq: ^PriorityQueue($T), i0, n: int) {
    i := i0
    for {
        j1 := 2 * i + 1
        if j1 >= n do break
        j := j1
        if j2 := j1 + 1; j2 < n && pq.cmp(pq.queue[j2], pq.queue[j1]) {
            j = j2
        }
        if !pq.cmp(pq.queue[j], pq.queue[i]) do break
        pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
        i = j
    }
}

// ——————————————————————————————————————————————
// String interner (simplified, using Odin map for permanent strings)
// ——————————————————————————————————————————————

InternedString :: distinct u32

string_interner_load :: proc(interned: InternedString) -> string {
// In a real interner you would look up an offset.
// Here we store strings in a global map keyed by hash.
// Not implemented fully, returns empty.
// TODO: implement persistent string storage.
    return ""
}

string_interner_insert :: proc(s: string) -> InternedString {
// placeholder
    return InternedString(0)
}

// ——————————————————————————————————————————————
// Path and file utilities
// ——————————————————————————————————————————————

path_is_directory :: proc(path: string) -> bool {
    info, err := os.lstat(path)
    if err != nil do return false
    return info.is_dir
}

path_to_full_path :: proc(path: string, allocator := context.allocator) -> string {
    abs, err := filepath.abs(path)
    if err != nil do return path
    // normalize slashes
    return strings.replace_all(abs, "\\", "/")
}

remove_directory_from_path :: proc(s: string) -> string {
    return filepath.base(s)
}

read_directory :: proc(path: string, allocator := context.allocator) -> ([]FileInfo, ReadDirectoryError) {
    if !path_is_directory(path) do return nil, .NotDir
    entries, err := os.read_dir(path)
    if err != nil do return nil, .InvalidPath
    info_list := make([dynamic]FileInfo, 0, len(entries), allocator)
    for entry in entries {
        full := filepath.join(path, entry.name)
        info := FileInfo{
            name     = entry.name,
            fullpath = full,
            size     = entry.size,
            is_dir   = entry.is_dir,
        }
        append(&info_list, info)
    }
    return info_list[:], .None
}

FileInfo :: struct {
    name: string,
    fullpath: string,
    size: i64,
    is_dir: bool,
}

ReadDirectoryError :: enum {
    None,
    InvalidPath,
    NotExists,
    Permission,
    NotDir,
    Empty,
    Unknown,
}

// Loaded file (simplified)
LoadedFile :: struct {
    handle: rawptr,
    data: []u8,
}

load_file :: proc(fullpath: string, copy_contents: bool) -> (LoadedFile, LoadedFileError) {
    data, ok := os.read_entire_file(fullpath)
    if !ok do return LoadedFile{ }, .NotExists
    return LoadedFile{ data = data }, .None
}

LoadedFileError :: enum {
    None,
    Empty,
    FileTooLarge,
    Invalid,
    NotExists,
    Permission,
}

// ——————————————————————————————————————————————
// Levenshtein distance (case insensitive)
// ——————————————————————————————————————————————

levenstein_distance_case_insensitive :: proc(a, b: string) -> int {
    h := len(a) + 1
    w := len(b) + 1
    matrix_ := make([]int, h * w, context.temp_allocator)
    defer delete(matrix_, context.temp_allocator)
    for i in 0 ..= len(a) {
        matrix_[i * w] = i
    }
    for j in 0 ..= len(b) {
        matrix_[j] = j
    }
    for i in 1 ..= len(a) {
        ac := unicode.to_lower(rune(a[i - 1]))
        for j in 1 ..= len(b) {
            bc := unicode.to_lower(rune(b[j - 1]))
            cost := 1
            if ac == bc {
                cost = 0
            }
            sub := matrix_[(i - 1) * w + j - 1] + cost
            del := matrix_[(i - 1) * w + j] + 1
            ins := matrix_[i * w + j - 1] + 1
            min_cell := min(del, ins, sub)
            if i > 1 && j > 1 && ac == unicode.to_lower(rune(b[j - 2])) && unicode.to_lower(rune(a[i - 2])) == bc {
                transpose := matrix_[(i - 2) * w + j - 2] + 1
                min_cell = min(min_cell, transpose)
            }
            matrix_[i * w + j] = min_cell
        }
    }
    return matrix_[len(a) * w + len(b)]
}

// Did you mean helpers
DistanceAndTarget :: struct {
    distance: int,
    target: string,
}

DidYouMeanAnswers :: struct {
    distances: [dynamic]DistanceAndTarget,
    key: string,
}

did_you_mean_make :: proc(key: string, cap: int) -> DidYouMeanAnswers {
    return DidYouMeanAnswers{
        distances = make([dynamic]DistanceAndTarget, 0, cap),
        key = key,
    }
}

did_you_mean_append :: proc(d: ^DidYouMeanAnswers, target: string) {
    if len(target) == 0 || target == "_" do return
    dat := DistanceAndTarget{
        target   = target,
        distance = levenstein_distance_case_insensitive(d.key, target),
    }
    append(&d.distances, dat)
}

did_you_mean_results :: proc(d: ^DidYouMeanAnswers) -> []DistanceAndTarget {
// sort by distance
    sort_proc :: proc(i, j: int, slice: []DistanceAndTarget) -> bool {
        return slice[i].distance < slice[j].distance
    }
    // Note: Odin does not have a built-in sort for slices of structs with custom comparator in std;
    // you would normally use `sort.slice` from `core:sort`. We'll keep a stub.
    // TODO: sort d.distances
    max_dist := 2 // MAX_SMALLEST_DID_YOU_MEAN_DISTANCE = 3-1
    count := 0
    for dist in d.distances {
        if dist.distance > max_dist do break
        count += 1
    }
    return d.distances[:count]
}

// ——————————————————————————————————————————————
// Common utilities from original C++ code
// ——————————————————————————————————————————————

f32_to_f16 :: proc(value: f32) -> u16 {
// use the same bit‑manipulation, Odin can do union transmute
    i := transmute(u32)value
    s := u16((i >> 16) & 0x8000)
    e := i32(((i >> 23) & 0xff)) - 127 + 15
    m := i & 0x007fffff
    if e <= 0 {
        if e < -10 do return s
        m = (m | 0x00800000) >> u32(1 - e)
        if m & 0x1000 != 0 {
            m += 0x2000
        }
        return s | u16(m >> 13)
    } else if e == 0xff - (127 - 15) {
        if m == 0 do return s | 0x7c00
        m >>= 13
        return s | 0x7c00 | u16(m) | u16(u32(m == 0 ? 1 : 0))
    } else {
        if m & 0x1000 != 0 {
            m += 0x2000
            if m & 0x00800000 != 0 {
                m = 0
                e += 1
            }
        }
        if e > 30 {
            return s | 0x7c00
        }
        return s | u16(e << 10) | u16(m >> 13)
    }
}

f16_to_f32 :: proc(value: u16) -> f32 {
// not implemented
    return 0.0
}

// ——————————————————————————————————————————————
// Obfuscate helpers (optional)
// ——————————————————————————————————————————————

obfuscate_string :: proc(s: string, prefix: string) -> string {
    if len(s) == 0 do return s
    assert(prefix != "")
    hash := fnv64a(s)
    res := fmt.aprintf("%sx%x", prefix, hash)  // allocated with context allocator
    // memory leak if not freed – but in command line tools it's okay
    return res
}

obfuscate_i32 :: proc(i: i32) -> i32 {
    x := i32(fnv64a( transmute([]u8)[]i32{ i } ))
    if x < 0 do x = 1 - x
    return x
}

// FNV‑1a hash (64‑bit) – using Odin's core:hash or manual implementation
fnv64a :: proc(data: []u8) -> u64 {
    h := 0xcbf29ce484222325
    for b in data {
        h = (h ~ u64(b)) * 0x100000001b3
    }
    return h
}

// ——————————————————————————————————————————————
// Debug output (placeholder)
// ——————————————————————————————————————————————

debugf :: proc(fmt_str: string, args: ..any) {
// Output to stderr, uses context.allocator for formatting
    fmt.eprintf(fmt_str, ..args)
}

// ——————————————————————————————————————————————
// Init / global state
// ——————————————————————————————————————————————

@(init)
init_common :: proc() {
// Initialize any required global state, e.g., string interner, default arenas.
// Since we dropped custom memory, none needed.
}

// ——————————————————————————————————————————————
// The end of common.odin
// ——————————————————————————————————————————————