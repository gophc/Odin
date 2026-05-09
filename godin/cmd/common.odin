// common.odin - godin common library
package godin

import "core:fmt"
import "core:os"
import "core:path/filepath"
import "core:strings"
import "core:unicode"


// ==============================================
// Utility math / bit operations
// ==============================================
next_pow2_32 :: proc(x: u32) -> u32 {
	if x == 0 do return 0
	x -= 1; x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16
	return x + 1
}
next_pow2_64 :: proc(x: u64) -> u64 {
	if x == 0 do return 0
	x -= 1; x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16; x |= x >> 32
	return x + 1
}
next_pow2_int :: proc(x: int) -> int {
	return int(next_pow2_64(u64(x)))
}
is_power_of_two_int :: proc(x: int) -> bool {
	return x > 0 && (x & (x - 1)) == 0
}
bit_set_count_u32 :: proc(x: u32) -> u32 {
	x -= (x >> 1) & 0x55555555
	x = (x & 0x33333333) + ((x >> 2) & 0x33333333)
	x = (x + (x >> 4)) & 0x0f0f0f0f
	x += x >> 8; x += x >> 16
	return x & 0x3f
}
bit_set_count_u64 :: proc(x: u64) -> u64 {
	lo := u32(x); hi := u32(x >> 32)
	return u64(bit_set_count_u32(lo) + bit_set_count_u32(hi))
}
floor_log2_u32 :: proc(x: u32) -> u32 {
	y := x; y |= y >> 1; y |= y >> 2; y |= y >> 4; y |= y >> 8; y |= y >> 16
	return bit_set_count_u32(y) - 1
}
ceil_log2_u32 :: proc(x: u32) -> u32 {
	y := i32(x & (x - 1)); y |= -y; y >>= 31
	z := x; z |= z >> 1; z |= z >> 2; z |= z >> 4; z |= z >> 8; z |= z >> 16
	return u32(bit_set_count_u32(z) - 1 - u32(y))
}
prev_pow2_u32 :: proc(x: u32) -> u32 {
	if x == 0 do return 0
	x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16
	return x - (x >> 1)
}
isize_cmp :: proc(a, b: int) -> int {
	if a < b do return -1
	if a > b do return 1
	return 0
}
i64_cmp :: proc(a, b: i64) -> int {
	if a < b do return -1
	if a > b do return 1
	return 0
}
i32_cmp :: proc(a, b: i32) -> int {
	if a < b do return -1
	if a > b do return 1
	return 0
}
u64_cmp :: proc(a, b: u64) -> int {
	if a < b do return -1
	if a > b do return 1
	return 0
}

// ==============================================
// Debug output
// ==============================================
debugf :: proc(fmt_str: string, args: ..any) {
	fmt.eprintf(fmt_str, ..args)
}

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
remove_directory_from_path :: proc(s: string) -> string {
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
string16_len :: proc(data: []u16) -> int {
	return len(data)
}
string16_to_string :: proc(s: []u16) -> string {
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
	return strings.replace_all(abs, "\\", "/")
}

Path :: struct {
	basename, name, ext: string,
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

FileInfo :: struct {
	name, fullpath: string,
	size: i64,
	is_dir: bool,
}
ReadDirectoryError :: enum {
	None, InvalidPath, NotExists, Permission, NotDir, Empty, Unknown,
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
// Levenshtein distance (Damerau-Levenshtein, case insensitive)
// ==============================================
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
	max_dist := 2
	count := 0
	for dist in d.distances {
		if dist.distance > max_dist do continue
		count += 1
	}
	if count == 0 do return nil
	result := make([]DistanceAndTarget, count)
	idx := 0
	for dist in d.distances {
		if dist.distance <= max_dist {
			result[idx] = dist
			idx += 1
		}
	}
	return result
}

// ==============================================
// Priority queue
// ==============================================
PriorityQueue :: struct($T: typeid) {
	queue: [dynamic]T,
	cmp: proc(a, b: T) -> bool,
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
priority_queue_peek :: proc(pq: ^PriorityQueue($T)) -> T {
	assert(len(pq.queue) > 0)
	return pq.queue[0]
}
priority_queue_len :: proc(pq: ^PriorityQueue($T)) -> int {
	return len(pq.queue)
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

// ==============================================
// Hash maps / sets
// ==============================================
PtrMap :: struct($K, $V: typeid) {
	m: map[K]V,
}
map_init :: proc(h: ^PtrMap($K, $V)) {
	h.m = make(map[K]V)
}
map_get :: proc(h: ^PtrMap($K, $V), key: K) -> ^V {
	if v, ok := &h.m[key]; ok do return v
	return nil
}
map_set :: proc(h: ^PtrMap($K, $V), key: K, val: V) {
	h.m[key] = val
}

StringMap :: struct($T: typeid) {
	m: map[string]T,
}
string_map_init :: proc(h: ^StringMap($T)) {
	h.m = make(map[string]T)
}
string_map_get :: proc(h: ^StringMap($T), key: string) -> ^T {
	if v, ok := &h.m[key]; ok do return v
	return nil
}
string_map_set :: proc(h: ^StringMap($T), key: string, val: T) {
	h.m[key] = val
}

StringSet :: struct {
	m: map[string]struct{},
}
string_set_init :: proc(s: ^StringSet) {
	s.m = make(map[string]struct{})
}
string_set_add :: proc(s: ^StringSet, str: string) {
	s.m[str] = {}
}
string_set_exists :: proc(s: ^StringSet, str: string) -> bool {
	_, ok := s.m[str]; return ok
}

// ==============================================
// String interner
// ==============================================
InternedString :: int
StringInterner :: struct {
	store: map[string]InternedString,
	rev: [dynamic]string,
}
string_interner_insert :: proc(s: string) -> InternedString {
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
	lo, hi: i64,
}
RangeCache :: struct {
	ranges: [dynamic]RangeValue,
}
range_cache_make :: proc() -> RangeCache {
	return RangeCache{}
}
range_cache_add_index :: proc(c: ^RangeCache, idx: i64) -> bool {
	for _, i in c.ranges {
		r := &c.ranges[i]
		if r.lo - 1 == idx {
			r.lo = idx
			return true
		}
		if r.hi + 1 == idx {
			r.hi = idx
			return true
		}
		if r.lo <= idx && idx <= r.hi {
			return true
		}
	}
	append(&c.ranges, RangeValue{ lo = idx, hi = idx })
	return false
}
range_cache_add_range :: proc(c: ^RangeCache, lo, hi: i64) -> bool {
	if lo > hi do return false
	for _, i in c.ranges {
		r := &c.ranges[i]
		if r.lo - 1 == hi {
			r.lo = lo
			return true
		}
		if r.hi + 1 == lo {
			r.hi = hi
			return true
		}
		if r.lo <= lo && hi <= r.hi {
			return true
		}
	}
	append(&c.ranges, RangeValue{ lo = lo, hi = hi })
	return false
}

// ==============================================
// FP16 conversions
// ==============================================
f32_to_f16 :: proc(value: f32) -> u16 {
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
	}
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
f16_to_f32 :: proc(value: u16) -> f32 {
	s := u32(value & 0x8000) << 16
	e := i32((value >> 10) & 0x1f)
	m := u32(value & 0x03ff)
	if e == 0 {
		if m == 0 do return transmute(f32)s
		for m & 0x0400 == 0 {
			m <<= 1
			e -= 1
		}
		m &= 0x03ff
		e += 1
	} else if e == 31 {
		return transmute(f32)(s | 0x7f800000 | (m << 13))
	}
	e += 127 - 15
	m <<= 13
	return transmute(f32)(s | (u32(e) << 23) | m)
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
fnv64a :: proc(data: []u8) -> u64 {
	h : u64 = 0xcbf29ce484222325
	for b in data {
		h = (h ~ u64(b)) * 0x100000001b3
	}
	return h
}
obfuscate_string :: proc(s: string, prefix: string) -> string {
	if len(s) == 0 do return s
	assert(prefix != "")
	hash := fnv64a(transmute([]u8)s)
	return fmt.aprintf("%sx%x", prefix, hash)
}
obfuscate_i32 :: proc(i: i32) -> i32 {
	x := i32(fnv64a(transmute([]u8)[]i32{ i }))
	if x < 0 do x = 1 - x
	return x
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
	handle: rawptr,
	data: []byte,
}
LoadedFileError :: enum {
	None, Empty, FileTooLarge, Invalid, NotExists, Permission,
}
load_file :: proc(fullpath: string, copy_contents: bool) -> (LoadedFile, LoadedFileError) {
	data, ok := os.read_entire_file(fullpath)
	if !ok do return LoadedFile{}, .NotExists
	return LoadedFile{ data = data }, .None
}

// ==============================================
// Global initialization
// ==============================================
@(init)
init_common :: proc() {
	g_interner.store = make(map[string]InternedString)
    g_interned_blank = string_interner_insert("_")
}

g_interner: StringInterner
g_interned_blank: InternedString
