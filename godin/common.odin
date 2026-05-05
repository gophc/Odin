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
import "core:unicode/utf8"

// ==============================================
// Basic type aliases (C++ → Odin)
// ==============================================
i8 :: i8; i16 :: i16; i32 :: i32; i64 :: i64; isize :: int
u8 :: u8; u16 :: u16; u32 :: u32; u64 :: u64; usize :: uint
f32 :: f32; f64 :: f64
uintptr :: uintptr
b32 :: bool
Rune :: rune

// ==============================================
// Utility math / bit operations
// ==============================================
next_pow2_32 :: proc(x: u32) -> u32 {
    if x == 0 do return 0; x -= 1; x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16; return x + 1
}
next_pow2_64 :: proc(x: u64) -> u64 {
    if x == 0 do return 0; x -= 1; x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16; x |= x >> 32; return x + 1
}
next_pow2_int :: proc(x: int) -> int {
    return int(next_pow2_64(u64(x)))
}
is_power_of_two_int :: proc(x: int) -> bool {
    return x > 0 && (x & (x - 1)) == 0
}
bit_set_count_u32 :: proc(x: u32) -> u32 {
    x -= (x >> 1) & 0x55555555; x = (x & 0x33333333) + ((x >> 2) & 0x33333333); x = (x + (x >> 4)) & 0x0f0f0f0f; x += x >> 8; x += x >> 16; return x & 0x3f
}
bit_set_count_u64 :: proc(x: u64) -> u64 {
    lo := u32(x); hi := u32(x >> 32); return u64(bit_set_count_u32(lo) + bit_set_count_u32(hi))
}
floor_log2_u32 :: proc(x: u32) -> u32 {
    y := x; y |= y >> 1; y |= y >> 2; y |= y >> 4; y |= y >> 8; y |= y >> 16; return bit_set_count_u32(y) - 1
}
ceil_log2_u32 :: proc(x: u32) -> u32 {
    y := i32(x & (x - 1)); y |= -y; y >>= 31; z :=x; z |= z >> 1; z |= z >> 2; z |= z >> 4; z |= z >> 8; z |= z >> 16; return u32(bit_set_count_u32(z) - 1 - u32(y))
}
prev_pow2_u32 :: proc(x: u32) -> u32 {
    if x == 0 do return 0; x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16; return x - (x >> 1)
}
// ... 类似 64 位版本省略

isize_cmp :: proc(a, b: int) -> int {
    if a < b do return -1; if a > b do return 1; return 0
}
i64_cmp :: proc(a, b: i64) -> int {
    if a < b do return -1; if a > b do return 1; return 0
}
i32_cmp :: proc(a, b: i32) -> int {
    if a < b do return -1; if a > b do return 1; return 0
}
u64_cmp :: proc(a, b: u64) -> int {
    if a < b do return -1; if a > b do return 1; return 0
}

// ==============================================
// Unicode helpers
// ==============================================
rune_is_whitespace :: proc(r: rune) -> bool {
    return unicode.is_space(r)
}
gb_char_to_lower :: proc(c: u8) -> u8 {
    return u8(unicode.to_lower(rune(c)))
}
rune_is_letter :: proc(r: rune) -> bool {
    return unicode.is_letter(r)
}
rune_is_digit :: proc(r: rune) -> bool {
    return unicode.is_digit(r)
}
utf8_decode :: proc(data: []u8, n: int, r: ^rune) -> int {
    if len(data) == 0 do return 0
    rr, size := utf8.decode_rune(data)
    if r != nil do r^ = rr
    return size
}
gb_utf8_encode_rune :: proc(buf: [4]u8, r: rune) -> int {
    n := utf8.encode_rune(buf[:], r)
    return n
}

// ==============================================
// Debug output
// ==============================================
debugf :: proc(fmt_str: string, args: ..any) {
    fmt.eprintf(fmt_str, ..args)
}

// ==============================================
// String types and simple constructors
// ==============================================
String :: string
String16 :: []u16

make_string :: proc(data: []u8) -> String {
    return string(data)
}
make_string_c :: proc(cstr: cstring) -> String {
    return string(cstr)
}
make_string16 :: proc(data: []u16) -> String16 {
    return data
}
make_string16_c :: proc(data: []u16) -> String16 {
    for i, c in data {
        if c == 0 {
            return data[:i]
        }
    }
    return data
}

// ==============================================
// Dynamic array / slice wrappers (Array<T>, Slice<T> compatibility)
// ==============================================
array_init :: proc($T: typeid, arr: ^[dynamic]T, capacity := 0) {
    if capacity > 0 do reserve(arr, capacity)
}
array_make :: proc($T: typeid, capacity := 0) -> [dynamic]T {
    return make([dynamic]T, 0, capacity)
}
array_free :: proc(arr: ^[dynamic]$T) {
    delete(arr^)
}
array_add :: proc(arr: ^[dynamic]$T, item: T) {
    append(arr, item)
}
array_pop :: proc(arr: ^[dynamic]$T) -> T {
    return pop(arr)
}
array_clear :: proc(arr: ^[dynamic]$T) {
    clear(arr)
}
array_reserve :: proc(arr: ^[dynamic]$T, cap: int) {
    reserve(arr, cap)
}
array_resize :: proc(arr: ^[dynamic]$T, count: int) {
    resize(arr, count)
}
array_ordered_remove :: proc(arr: ^[dynamic]$T, idx: int) {
    ordered_remove(arr, idx)
}
array_unordered_remove :: proc(arr: ^[dynamic]$T, idx: int) {
    unordered_remove(arr, idx)
}
array_end_ptr :: proc(arr: ^[dynamic]$T) -> ^T {
    if len(arr) > 0 do return &arr[len(arr) - 1]; return nil
}

slice_from_array :: proc($T: typeid, arr: [dynamic]T) -> []T {
    return arr[:]
}
slice_make :: proc($T: typeid, count: int) -> []T {
    return make([]T, count)
}
slice_clone :: proc(s: []$T) -> []T {
    c := make([]T, len(s)); copy(c, s); return c
}
slice_free :: proc(s: []$T) {
    delete(s)
}
slice_resize :: proc(s: ^[]$T, new_len: int) {
    resize(s^, new_len)
}
slice_copy :: proc(dst, src: []$T) {
    copy(dst, src)
}
slice_ordered_remove :: proc(s: ^[]$T, idx: int) {
    ordered_remove(s, idx)
}
slice_unordered_remove :: proc(s: ^[]$T, idx: int) {
    unordered_remove(s, idx)
}

// ==============================================
// Threading primitives
// ==============================================
BlockingMutex :: sync.Mutex
RecursiveMutex :: sync.RecursiveMutex
RwMutex :: sync.RW_Mutex
Condition :: sync.Cond

mutex_lock :: proc(m: ^BlockingMutex) {
    sync.lock(m)
}
mutex_try_lock :: proc(m: ^BlockingMutex) -> bool {
    return sync.try_lock(m)
}
mutex_unlock :: proc(m: ^BlockingMutex) {
    sync.unlock(m)
}
// RecursiveMutex wrappers similar...
rw_mutex_lock :: proc(m: ^RwMutex) {
    sync.lock(m)
}
rw_mutex_unlock :: proc(m: ^RwMutex) {
    sync.unlock(m)
}
rw_mutex_shared_lock :: proc(m: ^RwMutex) {
    sync.shared_lock(m)
}
rw_mutex_shared_unlock :: proc(m: ^RwMutex) {
    sync.shared_unlock(m)
}

condition_broadcast :: proc(c: ^Condition) {
    sync.broadcast(c)
}
condition_signal :: proc(c: ^Condition) {
    sync.signal(c)
}
condition_wait :: proc(c: ^Condition, m: ^BlockingMutex) {
    sync.wait(c, m)
}

Semaphore :: struct {
    m: sync.Mutex,
    cond: sync.Cond,
    cnt: int,
}
semaphore_post :: proc(s: ^Semaphore, n: int = 1) {
    sync.lock(&s.m); defer sync.unlock(&s.m)
    s.cnt += n
    sync.broadcast(&s.cond)
}
semaphore_wait :: proc(s: ^Semaphore) {
    sync.lock(&s.m); defer sync.unlock(&s.m)
    for s.cnt == 0 {
        sync.wait(&s.cond, &s.m)
    }
    s.cnt -= 1
}

// Thread (encapsulates an Odin thread)
Thread :: struct {
    sys_handle: ^thread.Thread,
    idx: int,
    pool: ^ThreadPool,
}
thread_current_id :: proc() -> u32 {
    return u32(os.current_thread_id())
}
yield_thread :: proc() {
    thread.yield()
}
yield_process :: proc() {
    thread.yield()
}
thread_init :: proc(t: ^Thread, idx: int, pool: ^ThreadPool) {
    t.idx = idx; t.pool = pool
}
thread_init_and_start :: proc(t: ^Thread, idx: int, pool: ^ThreadPool, proc_: thread.Proc) {
    thread_init(t, idx, pool)
    t.sys_handle = thread.create_with_data(proc_, t)
    assert(t.sys_handle != nil)
}
thread_join_and_destroy :: proc(t: ^Thread) {
    if t.sys_handle != nil {
        thread.join(t.sys_handle)
        thread.destroy(t.sys_handle)
        t.sys_handle = nil
    }
}
thread_set_name :: proc(t: ^Thread, name: string) {
// Not natively supported in Odin, just ignore
}

// Thread pool (simple channel‑based)
WorkerTaskProc :: proc(data: rawptr)
WorkerTask :: struct {
    do_work: WorkerTaskProc,
    data: rawptr,
}
ThreadPool :: struct {
    tasks: chan(WorkerTask),
    wg: sync.WaitGroup,
    threads: [dynamic]^thread.Thread,
    running: bool,
}
thread_pool_init :: proc(pool: ^ThreadPool, worker_count: int, name: string) {
    pool.tasks = make(chan(WorkerTask), 128)
    pool.running = true
    for i in 0 ..< worker_count {
        t := thread.create_with_data((proc(data: rawptr) {
            pool := (^ThreadPool)(data)
            for {
                task, ok := recv(pool.tasks)
                if !ok do break
                task.do_work(task.data)
                sync.wg_done(&pool.wg)
            }
        }), pool)
        assert(t != nil)
        append(&pool.threads, t)
    }
}
thread_pool_destroy :: proc(pool: ^ThreadPool) {
    close(pool.tasks)
    for t in pool.threads {
        thread.join(t)
        thread.destroy(t)
    }
    delete(pool.threads)
    delete(pool.tasks)
}
thread_pool_add_task :: proc(pool: ^ThreadPool, proc_: WorkerTaskProc, data: rawptr) -> bool {
    sync.wg_add(&pool.wg, 1)
    send(pool.tasks, WorkerTask{ proc_, data })
    return true
}
thread_pool_wait :: proc(pool: ^ThreadPool) {
    sync.wg_wait(&pool.wg)
}

string16_len :: proc(data: []u16) -> int {

}

// ==============================================
// Memory / Arena
// ==============================================
MemoryBlock :: struct {
    prev: ^MemoryBlock,
    data: []u8,
    used: int,
}
Arena :: struct {
    curr_block: ^MemoryBlock,
    minimum_block_size: int,
    temp_count: int,
}
DEFAULT_MINIMUM_BLOCK_SIZE :: 8 * mem.Megabyte
DEFAULT_PAGE_SIZE :: 4096

arena_alloc :: proc(a: ^Arena, size, alignment: int) -> rawptr {
    if a.curr_block == nil || a.curr_block.used + size > len(a.curr_block.data) {
        block := new(MemoryBlock)
        block_size := max(size, a.minimum_block_size)
        block.data = make([]byte, block_size)
        block.prev = a.curr_block
        a.curr_block = block
    }
    block := a.curr_block
    ptr := uintptr(raw_data(block.data)) + uintptr(block.used)
    aligned := ptr + uintptr(alignment - 1)
    aligned -= aligned % uintptr(alignment)
    offset := int(aligned - ptr)
    block.used += offset + size
    return rawptr(aligned)
}
arena_free_all :: proc(a: ^Arena) {
    for a.curr_block != nil {
        block := a.curr_block
        a.curr_block = block.prev
        delete(block.data)
        free(block)
    }
}
ArenaTemp :: struct {
    arena: ^Arena,
    block: ^MemoryBlock,
    used: int,
}
arena_temp_begin :: proc(a: ^Arena) -> ArenaTemp {
    temp := ArenaTemp{ arena = a }
    if a.curr_block != nil {
        temp.block = a.curr_block
        temp.used = a.curr_block.used
    }
    a.temp_count += 1
    return temp
}
arena_temp_end :: proc(temp: ArenaTemp) {
    arena := temp.arena
    assert(arena.temp_count > 0)
    arena.temp_count -= 1
    if temp.block != nil {
    // Roll back current block to temp.used
        for arena.curr_block != temp.block {
            block := arena.curr_block
            arena.curr_block = block.prev
            delete(block.data)
            free(block)
        }
        arena.curr_block.used = temp.used
    }
}
ArenaTempGuard :: struct {
    temp: ArenaTemp
}
arena_allocator :: proc(a: ^Arena) -> runtime.Allocator {
    return runtime.Allocator{
        procedure = (proc(data: rawptr, mode: mem.Allocator_Mode, size, alignment: int, old_memory: rawptr, old_size: int, flags: u64) -> rawptr {
            switch mode {
            case .Alloc: return arena_alloc((^Arena)(data), size, alignment)
            case .Free: return nil
            case .Free_All: arena_free_all((^Arena)(data)); return nil
            case .Resize:
                if size <= old_size do return old_memory
                new_mem := arena_alloc((^Arena)(data), size, alignment)
                mem.copy(new_mem, old_memory, old_size)
                return new_mem
            }
            return nil
        }),
        data = a,
    }
}
// Default thread arenas
default_permanent_arena: Arena
default_temporary_arena: Arena
permanent_allocator :: proc() -> runtime.Allocator {
    return arena_allocator(&default_permanent_arena)
}
temporary_allocator :: proc() -> runtime.Allocator {
    return arena_allocator(&default_temporary_arena)
}

// ==============================================
// String utilities (extensive)
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
// String split iterator (as in C++)
String_Iterator :: struct {
    str: string, pos: int
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
path_extension :: proc(s: string, include_dot := true) -> string {
    pos := string_extension_position(s)
    if pos < 0 do return ""
    start := pos if include_dot else pos + 1
    return s[start:]
}
path_remove_extension :: proc(s: string) -> string {
    pos := string_extension_position(s)
    if pos < 0 do return s
    return s[:pos]
}
filename_from_path :: proc(s: string) -> string {
    pos := string_extension_position(s)
    if pos >= 0 do return s[:pos]
    return ""
}
filename_without_directory :: proc(s: string) -> string {
    return filepath.base(s)
}
last_path_element :: proc(s: string) -> string {
    return filepath.base(s)
}
clone_string :: proc(s: string) -> string {
    return strings.clone(s)
}
alloc_cstring :: proc(s: string) -> cstring {
    c := make([]byte, len(s) + 1, context.temp_allocator)
    copy(c, s)
    c[len(s)] = 0
    return cstring(raw_data(c))
}
alloc_wstring :: proc(s: []u16) -> []u16 {
    c := make([]u16, len(s) + 1, context.temp_allocator)
    copy(c, s)
    c[len(s)] = 0
    return c
}
string16_to_string :: proc(s: []u16) -> string {
    runtime.DEFAULT_TEMP_ALLOCATOR_TEMP_GUARD()
    buf := make([]byte, len(s) * 3)
    n := unicode.utf16_to_utf8(buf, s)
    return string(buf[:n])
}
string_to_string16 :: proc(s: string) -> []u16 {
    return unicode.utf8_to_utf16(s)
}
get_working_directory :: proc() -> string {
    wd, err := os.get_current_directory()
    if err != nil {
        return ""
    }
    return wd
}
set_working_directory :: proc(dir: string) -> bool {
    return os.set_current_directory(dir) == nil
}
path_is_directory :: proc(path: string) -> bool {
    info, err := os.lstat(path)
    return err == nil && info.is_dir
}
directory_from_path :: proc(s: string) -> string {
    if path_is_directory(s) do return s
    return filepath.dir(s)
}
path_to_full_path :: proc(path: string) -> string {
    abs, _ := filepath.abs(path)
    return abs
}
// Path struct
Path :: struct {
    basename, name, ext: string
}
path_from_string :: proc(fullpath: string) -> (p: Path, ok: bool) {
    clean := path_to_full_path(fullpath)
    p.basename = filepath.dir(clean)
    base := filepath.base(clean)
    if path_is_directory(clean) {
        return p, true
    }
    ext := filepath.ext(base)
    p.ext = ext
    p.name = strings.trim_suffix(base, ext)
    return p, true
}
// Read directory
FileInfo :: struct {
    name, fullpath: string; size: i64; is_dir: bool
}
ReadDirectoryError :: enum {
    None, InvalidPath, NotExists, Permission, NotDir, Empty, Unknown
}
read_directory :: proc(path: string, allocator := context.allocator) -> ([]FileInfo, ReadDirectoryError) {
    entries, err := os.read_dir(path)
    if err != nil {
        return nil, .InvalidPath
    }
    list := make([dynamic]FileInfo, 0, len(entries), allocator)
    for e in entries {
        info := FileInfo{
            name     = e.name,
            fullpath = filepath.join(path, e.name),
            size     = e.size,
            is_dir   = e.is_dir,
        }
        append(&list, info)
    }
    return list[:], .None
}
write_directory :: proc(path: string) -> bool {
    return os.make_directory(path) == nil
}

// ==============================================
// Levenshtein / Did you mean
// ==============================================
levenstein_distance_case_insensitive :: proc(a, b: string) -> int {
// implementation (as before, omitted for brevity) ...
}
DistanceAndTarget :: struct {
    distance: int; target: string
}
DidYouMeanAnswers :: struct {
    distances: [dynamic]DistanceAndTarget; key: string
}
did_you_mean_make :: proc(key: string, cap: int) -> DidYouMeanAnswers {
    ...
}
did_you_mean_append :: proc(d: ^DidYouMeanAnswers, target: string) {
    ...
}
did_you_mean_results :: proc(d: ^DidYouMeanAnswers) -> []DistanceAndTarget {
    ...
}

// ==============================================
// Priority queue (generic)
// ==============================================
PriorityQueue :: struct($T: typeid) {
    queue: [dynamic]T,
    less: proc(a, b: T) -> bool,
}
priority_queue_push :: proc(pq: ^PriorityQueue($T), v: T) {
    ...
}
// ... rest of implementation

// ==============================================
// Hash maps / sets (PtrMap, StringMap, StringSet, etc.)
// ==============================================
PtrMap :: struct($K, $V: typeid) {
    m: map[K]V
}
map_init :: proc(h: ^PtrMap($K, $V)) {
    h.m = make(map[K]V)
}
map_get :: proc(h: ^PtrMap($K, $V), key: K) -> ^V {
    if v, ok := &h.m[key]; ok do return v; return nil
}
map_set :: proc(h: ^PtrMap($K, $V), key: K, val: V) {
    h.m[key] = val
}
// OrderedInsertPtrMap can be emulated with a [dynamic]K and map[K]V

StringMap :: struct($T: typeid) {
    m: map[string]T
}
string_map_init :: proc(h: ^StringMap($T)) {
    h.m = make(map[string]T)
}
string_map_get :: proc(h: ^StringMap($T), key: string) -> ^T {
    if v, ok := &h.m[key]; ok do return v; return nil
}
string_map_set :: proc(h: ^StringMap($T), key: string, val: T) {
    h.m[key] = val
}

StringSet :: struct {
    m: map[string]struct{
    }
}
string_set_init :: proc(s: ^StringSet) {
    s.m = make(map[string]struct{
    })
}
string_set_add :: proc(s: ^StringSet, str: string) {
    s.m[str] = { }
}
string_set_exists :: proc(s: ^StringSet, str: string) -> bool {
    _, ok := s.m[str]; return ok
}

// String interner
StringInterner :: struct {
    store: map[string]InternedString,
    rev: [dynamic]string,
}
InternedString :: int
string_interner_insert :: proc(s: string) -> InternedString {
// lock not shown for brevity
    if idx, ok := g_interner.store[s]; ok do return idx
    idx := InternedString(len(g_interner.rev))
    append(&g_interner.rev, s)
    g_interner.store[s] = idx
    return idx
}
string_interner_load :: proc(id: InternedString) -> string {
    return g_interner.rev[id]
}

// ==============================================
// Range cache
// ==============================================
RangeValue :: struct {
    lo, hi: i64
}
RangeCache :: struct {
    ranges: [dynamic]RangeValue
}
range_cache_make :: proc() -> RangeCache {
    return RangeCache{ }
}
range_cache_add_index :: proc(c: ^RangeCache, idx: i64) -> bool {
    ...
}
range_cache_add_range :: proc(c: ^RangeCache, lo, hi: i64) -> bool {
    ...
}

// ==============================================
// FP16 conversions
// ==============================================
f32_to_f16 :: proc(val: f32) -> u16 {
/* todo ... */
}
f16_to_f32 :: proc(val: u16) -> f32 {
/* todo ... */
}

// ==============================================
// Hashing, obfuscation
// ==============================================
fnv32a :: proc(data: []byte) -> u32 {
    h : u32 = 0x811c9dc5
    for b in data {
        h = (h ~ u32(b)) * 0x01000193
    }
    return h
}
fnv64a :: proc(data: []byte) -> u64 {
    h : u64 = 0xcbf29ce484222325
    for b in data {
        h = (h ~ u64(b)) * 0x100000001b3
    }
    return h
}
obfuscate_string :: proc(s: string, prefix: string) -> string {
    ...
}
obfuscate_i32 :: proc(i: i32) -> i32 {
    ...
}
u64_to_string :: proc(v: u64) -> string {
    return fmt.aprintf("%v", v)
}
i64_to_string :: proc(v: i64) -> string {
    return fmt.aprintf("%v", v)
}

// ==============================================
// Loaded file
// ==============================================
LoadedFile :: struct {
    handle: rawptr; data: []byte
}
LoadedFileError :: enum {
    None, Empty, FileTooLarge, Invalid, NotExists, Permission
}
load_file :: proc(fullpath: string, copy_contents: bool) -> (LoadedFile, LoadedFileError) {
    data, ok := os.read_entire_file(fullpath)
    if !ok do return LoadedFile{ }, .NotExists
    return LoadedFile{ data = data }, .None
}

// ==============================================
// Global initialization
// ==============================================
@(init)
init_common :: proc() {
// Initialize default arenas, string interner, etc.
    default_permanent_arena.minimum_block_size = DEFAULT_MINIMUM_BLOCK_SIZE
    default_temporary_arena.minimum_block_size = DEFAULT_MINIMUM_BLOCK_SIZE
    g_interner.store = make(map[string]InternedString)
}

g_interner: StringInterner
g_interned_blank: InternedString

// ==============================================
// End
// ==============================================