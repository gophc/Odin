package godin

import "base:runtime"
import "core:math"
import "core:mem"
import "core:os"
import "core:slice"
import "core:sort"
import "core:strings"
import "core:sync/mutex"
import "core:unicode"
import "core:path/filepath"

// ============================================================================
// Block 1: Basic Utilities
// Translated from src/common.cpp
// ============================================================================

// --- heap_allocator ---

heap_allocator :: proc() -> runtime.Allocator {
	return context.allocator
}

// --- next_pow2 / prev_pow2 ---

next_pow2_i32 :: proc(n: i32) -> i32 {
	if n <= 0 do return 0
	x := n - 1
	x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16
	return x + 1
}

next_pow2_i64 :: proc(n: i64) -> i64 {
	if n <= 0 do return 0
	x := n - 1
	x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16; x |= x >> 32
	return x + 1
}

next_pow2_int :: proc(n: int) -> int {
	if n <= 0 do return 0
	x := n - 1
	x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16
	when size_of(int) == 8 { x |= x >> 32 }
	return x + 1
}

next_pow2_u32 :: proc(n: u32) -> u32 {
	if n == 0 do return 0
	x := n - 1
	x |= x >> 1; x |= x >> 2; x |= x >> 4; x |= x >> 8; x |= x >> 16
	return x + 1
}

prev_pow2_u32 :: proc(n: u32) -> u32 {
	if n == 0 do return 0
	return next_pow2_u32(n) >> 1
}

prev_pow2_i32 :: proc(n: i32) -> i32 {
	if n <= 0 do return 0
	return cast(i32)(cast(u32)next_pow2_i32(n) >> 1)
}

prev_pow2_i64 :: proc(n: i64) -> i64 {
	if n <= 0 do return 0
	return cast(i64)(cast(u64)next_pow2_i64(n) >> 1)
}

// --- Type trait helpers (approximate C++ template metaprogramming) ---

TYPE_IS_PTR_SIZED_INTEGER :: proc($T: typeid) -> bool {
	return T == int || T == uint
}

// --- is_power_of_two ---

is_power_of_two :: proc(x: i64) -> bool {
	return x > 0 && (x & (x - 1)) == 0
}

is_power_of_two_u64 :: proc(x: u64) -> bool {
	return x > 0 && (x & (x - 1)) == 0
}

// --- Comparison functions ---

isize_cmp :: proc(a, b: int) -> int {
	if a < b do return -1
	if a > b do return 1
	return 0
}

u64_cmp :: proc(a, b: u64) -> int {
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

// --- FNV hash functions ---

FNV32_OFFSET :: u32(2166136261)
FNV32_PRIME  :: u32(16777619)

fnv32a :: proc(data: []byte, seed: u32 = FNV32_OFFSET) -> u32 {
	h := seed
	for b in data {
		h ~= u32(b)
		h *= FNV32_PRIME
	}
	return h
}

FNV64_OFFSET :: u64(14695981039346656037)
FNV64_PRIME  :: u64(1099511628211)

fnv64a :: proc(data: []byte, seed: u64 = FNV64_OFFSET) -> u64 {
	h := seed
	for b in data {
		h ~= u64(b)
		h *= FNV64_PRIME
	}
	return h
}

fnv64a_string :: proc(s: string, seed: u64 = FNV64_OFFSET) -> u64 {
	return fnv64a(transmute([]byte)s, seed)
}

// --- u64_digit_value and u64_from_string ---

u64_digit_value :: proc(r: rune) -> (u64, bool) {
	switch {
	case r >= '0' && r <= '9': return cast(u64)(r - '0'), true
	case r >= 'a' && r <= 'z': return cast(u64)(r - 'a' + 10), true
	case r >= 'A' && r <= 'Z': return cast(u64)(r - 'A' + 10), true
	}
	return 0, false
}

u64_from_string :: proc(s: string) -> (u64, bool) {
	if len(s) == 0 do return 0, false
	str := s
	base := u64(10)
	if len(str) >= 2 && str[0] == '0' {
		switch str[1] {
		case 'b': base = 2;  str = str[2:]
		case 'o': base = 8;  str = str[2:]
		case 'd': base = 10; str = str[2:]
		case 'z': base = 12; str = str[2:]
		case 'x': base = 16; str = str[2:]
		case 'h': base = 16; str = str[2:]
		}
	}
	if len(str) == 0 do return 0, false
	result: u64 = 0
	for r in str {
		if r == '_' do continue
		digit, ok := u64_digit_value(r)
		if !ok || digit >= base do return 0, false
		new_result := result * base + digit
		if new_result < result do return 0, false // overflow
		result = new_result
	}
	return result, true
}

// --- u64_to_string and i64_to_string ---

NUM_TO_CHAR_TABLE :: [16]byte{'0','1','2','3','4','5','6','7','8','9','a','b','c','d','e','f'}

u64_to_string :: proc(value: u64, buf: []byte) -> string {
	if len(buf) == 0 do return ""
	v := value
	if v == 0 {
		buf[0] = '0'
		return string(buf[:1])
	}
	i := len(buf)
	for v > 0 && i > 0 {
		i -= 1
		buf[i] = NUM_TO_CHAR_TABLE[v % 10]
		v /= 10
	}
	return string(buf[i:])
}

i64_to_string :: proc(value: i64, buf: []byte) -> string {
	if len(buf) == 0 do return ""
	if value >= 0 do return u64_to_string(cast(u64)value, buf)
	u := cast(u64)(-value)
	s := u64_to_string(u, buf[1:])
	buf[0] = '-'
	return string(buf[:len(s) + 1])
}

// --- Integer min/max compile-time constants ---

I8_MIN  :: i32(-128)
I8_MAX  :: i32(127)
U8_MAX  :: u32(255)
I16_MIN :: i32(-32768)
I16_MAX :: i32(32767)
U16_MAX :: u32(65535)
I32_MIN :: i32(-2147483648)
I32_MAX :: i32(2147483647)
U32_MAX :: u32(4294967295)
I64_MIN :: i64(-9223372036854775808)
I64_MAX :: i64(9223372036854775807)
U64_MAX :: u64(18446744073709551615)

// --- Overflow checks ---

add_overflow_u64 :: proc(x, y: u64) -> (result: u64, overflow: bool) {
	result = x + y
	overflow = result < x
	return
}

sub_overflow_u64 :: proc(x, y: u64) -> (result: u64, overflow: bool) {
	result = x - y
	overflow = result > x
	return
}

mul_overflow_u64 :: proc(x, y: u64) -> (lo: u64, hi: u64) {
	res := u128(x) * u128(y)
	lo = u64(res)
	hi = u64(res >> 64)
	return
}

// --- f32_to_f16 and f16_to_f32 ---

f32_to_f16 :: proc(value: f32) -> u16 {
	f := value
	u := transmute(u32)f
	sign := u16((u >> 16) & 0x8000)
	exponent := i32((u >> 23) & 0xff) - 127
	mantissa := u & 0x007fffff
	if exponent >= 16 {
		return sign | 0x7c00 // Infinity or too large
	} else if exponent >= -14 {
		return sign | u16((u16(exponent + 15) << 10) | u16(mantissa >> 13))
	} else if exponent >= -24 {
		shifted := u16((mantissa | 0x00800000) >> u32(1 - exponent))
		return sign | (shifted >> 13)
	}
	return sign // Underflow to zero
}

f16_to_f32 :: proc(value: u16) -> f32 {
	sign := u32(value & 0x8000) << 16
	exponent := i32((value >> 10) & 0x1f)
	mantissa := u32(value & 0x03ff)
	if exponent == 0 {
		if mantissa == 0 do return transmute(f32)(sign) // Zero
		for mantissa & 0x0400 == 0 { mantissa <<= 1; exponent -= 1 }
		mantissa &= 0x03ff
		biased_exp := u32(cast(i32)(exponent + 1) + 127)
		return transmute(f32)(sign | (biased_exp << 23) | (mantissa << 13))
	} else if exponent == 31 {
		return transmute(f32)(sign | (0xff << 23) | (mantissa << 13))
	}
	biased_exp := u32(cast(i32)(exponent) + 127 - 15)
	return transmute(f32)(sign | (biased_exp << 23) | (mantissa << 13))
}

// --- bit_set_count (popcount) ---

bit_set_count_u32 :: proc(x: u32) -> int {
	v := x
	v = (v & 0x55555555) + ((v >> 1) & 0x55555555)
	v = (v & 0x33333333) + ((v >> 2) & 0x33333333)
	v = (v & 0x0f0f0f0f) + ((v >> 4) & 0x0f0f0f0f)
	v = (v & 0x00ff00ff) + ((v >> 8) & 0x00ff00ff)
	v = (v & 0x0000ffff) + ((v >> 16) & 0x0000ffff)
	return cast(int)v
}

bit_set_count_u64 :: proc(x: u64) -> int {
	v := x
	v = (v & 0x5555555555555555) + ((v >> 1) & 0x5555555555555555)
	v = (v & 0x3333333333333333) + ((v >> 2) & 0x3333333333333333)
	v = (v & 0x0f0f0f0f0f0f0f0f) + ((v >> 4) & 0x0f0f0f0f0f0f0f0f)
	v = (v & 0x00ff00ff00ff00ff) + ((v >> 8) & 0x00ff00ff00ff00ff)
	v = (v & 0x0000ffff0000ffff) + ((v >> 16) & 0x0000ffff0000ffff)
	v = (v & 0x00000000ffffffff) + ((v >> 32) & 0x00000000ffffffff)
	return cast(int)v
}

// --- floor_log2 and ceil_log2 ---

floor_log2_u32 :: proc(x: u32) -> int {
	if x == 0 do return -1
	r: int = 0; v := x
	if v & 0xffff0000 != 0 { v >>= 16; r += 16 }
	if v & 0x0000ff00 != 0 { v >>= 8;  r += 8  }
	if v & 0x000000f0 != 0 { v >>= 4;  r += 4  }
	if v & 0x0000000c != 0 { v >>= 2;  r += 2  }
	if v & 0x00000002 != 0 { r += 1 }
	return r
}

floor_log2_u64 :: proc(x: u64) -> int {
	if x == 0 do return -1
	r: int = 0; v := x
	if v & 0xffffffff00000000 != 0 { v >>= 32; r += 32 }
	if v & 0x00000000ffff0000 != 0 { v >>= 16; r += 16 }
	if v & 0x000000000000ff00 != 0 { v >>= 8;  r += 8  }
	if v & 0x00000000000000f0 != 0 { v >>= 4;  r += 4  }
	if v & 0x000000000000000c != 0 { v >>= 2;  r += 2  }
	if v & 0x0000000000000002 != 0 { r += 1 }
	return r
}

ceil_log2_u32 :: proc(x: u32) -> int {
	if x <= 1 do return 0
	return floor_log2_u32(x - 1) + 1
}

ceil_log2_u64 :: proc(x: u64) -> int {
	if x <= 1 do return 0
	return floor_log2_u64(x - 1) + 1
}

// --- gb_sqrt ---

gb_sqrt :: proc(x: f64) -> f64 {
	return math.sqrt(x)
}

// global_module_path — skipped (application-specific globals)
// debugf — skipped (declaration only in original, no body)


// ============================================================================
// Block 2: Array/Slice wrappers + Queue implementations
// Translated from src/array.cpp and src/queue.cpp
// ============================================================================

// --- Array wrappers over [dynamic]T ---
// Thin wrappers around Odin's built-in [dynamic]T to match C++ Array<T> API.

array_init :: proc(a: ^[dynamic]T, allocator := context.allocator) {
	a^ = make([dynamic]T, allocator)
}

array_init_with_count :: proc(a: ^[dynamic]T, count: int, allocator := context.allocator) {
	a^ = make([dynamic]T, count, allocator)
}

array_init_with_capacity :: proc(a: ^[dynamic]T, count, capacity: int, allocator := context.allocator) {
	a^ = make([dynamic]T, count, capacity, allocator)
}

array_make :: proc($T: typeid, allocator := context.allocator) -> [dynamic]T {
	return make([dynamic]T, allocator)
}

array_make_with_count :: proc($T: typeid, count: int, allocator := context.allocator) -> [dynamic]T {
	return make([dynamic]T, count, allocator)
}

array_make_with_capacity :: proc($T: typeid, count, capacity: int, allocator := context.allocator) -> [dynamic]T {
	return make([dynamic]T, count, capacity, allocator)
}

array_free :: proc(a: ^[dynamic]T) {
	delete(a^)
}

array_add :: proc(a: ^[dynamic]T, item: T) {
	append(a, item)
}

array_add_and_get :: proc(a: ^[dynamic]T) -> ^T {
	n := len(a)
	append(a, T{})
	return &a[n]
}

array_add_elems :: proc(a: ^[dynamic]T, elems: []T) {
	append(a, ..elems)
}

array_pop :: proc(a: ^[dynamic]T) -> T {
	val := a[len(a)-1]
	resize(a, len(a)-1)
	return val
}

array_pop_safe :: proc(a: ^[dynamic]T) -> (T, bool) {
	n := len(a)
	if n == 0 do return T{}, false
	val := a[n-1]
	resize(a, n-1)
	return val, true
}

array_clear :: proc(a: ^[dynamic]T) {
	clear(a)
}

array_reserve :: proc(a: ^[dynamic]T, capacity: int) {
	reserve(a, capacity)
}

array_resize :: proc(a: ^[dynamic]T, count: int) {
	resize(a, count)
}

array_slice :: proc(a: [dynamic]T, lo, hi: int) -> []T {
	return a[lo:hi]
}

array_clone :: proc(a: [dynamic]T, allocator := context.allocator) -> [dynamic]T {
	res := make([dynamic]T, len(a), cap(a), allocator)
	append(&res, ..a[:])
	return res
}

array_ordered_remove :: proc(a: ^[dynamic]T, index: int) {
	ordered_remove(a, index)
}

array_unordered_remove :: proc(a: ^[dynamic]T, index: int) {
	unordered_remove(a, index)
}

// array_copy has two overloads
array_copy :: proc{
	array_copy_with_offset,
	array_copy_with_offset_and_count,
}

array_copy_with_offset :: proc(a: ^[dynamic]T, data: [dynamic]T, offset: int) {
	if offset < 0 do return
	n := len(data)
	for i in 0..<n {
		if offset + i >= len(a) do break
		a[offset + i] = data[i]
	}
}

array_copy_with_offset_and_count :: proc(a: ^[dynamic]T, data: [dynamic]T, offset, count: int) {
	if offset < 0 || count <= 0 do return
	for i in 0..<count {
		if offset + i >= len(a) || i >= len(data) do break
		a[offset + i] = data[i]
	}
}

array_end_ptr :: proc(a: ^[dynamic]T) -> ^T {
	if len(a) == 0 do return nil
	return &a[len(a)-1]
}

array_sort :: proc(a: ^[dynamic]T, cmp: proc(a, b: T) -> int) {
	slice.sort_by(a[:], cmp)
}

// --- Slice wrappers over []T ---

slice_from_array :: proc(a: [dynamic]T) -> []T {
	return a[:]
}

slice_array :: proc(a: [dynamic]T, lo, hi: int) -> []T {
	return a[lo:hi]
}

slice_make :: proc($T: typeid, count: int, allocator := context.allocator) -> []T {
	return make([]T, count, allocator)
}

slice_clone :: proc(s: []T, allocator := context.allocator) -> []T {
	res := make([]T, len(s), allocator)
	copy(res, s)
	return res
}

slice_clone_from_array :: proc(a: [dynamic]T, allocator := context.allocator) -> []T {
	res := make([]T, len(a), allocator)
	copy(res, a[:])
	return res
}

// slice_copy has three overloads
slice_copy :: proc{
	slice_copy_simple,
	slice_copy_with_offset,
	slice_copy_with_offset_and_count,
}

slice_copy_simple :: proc(dst: ^[]T, src: []T) {
	copy(dst^, src)
}

slice_copy_with_offset :: proc(dst: ^[]T, src: []T, offset: int) {
	if offset < 0 || offset >= len(dst) do return
	n := min(len(src), len(dst) - offset)
	copy((dst^)[offset:], src[:n])
}

slice_copy_with_offset_and_count :: proc(dst: ^[]T, src: []T, offset, count: int) {
	if offset < 0 || count <= 0 || offset >= len(dst) do return
	n := min(count, min(len(src), len(dst) - offset))
	copy((dst^)[offset:], src[:n])
}

slice_ordered_remove :: proc(s: ^[]T, index: int) {
	n := len(s)
	if index < 0 || index >= n do return
	if index < n-1 {
		mem.move(&s[index], &s[index+1], (n - index - 1) * size_of(T))
	}
	s^ = s^[:n-1]
}

slice_unordered_remove :: proc(s: ^[]T, index: int) {
	n := len(s)
	if index < 0 || index >= n do return
	s[index] = s[n-1]
	s^ = s^[:n-1]
}

// --- Queue implementations ---
// NOTE: The C++ MPSC and MPMC queues use lock-free std::atomic algorithms.
// These Odin versions are simplified mutex-based implementations.

// MPSCQueue — Multi-Producer Single-Consumer (simplified)

MPSCQueue :: struct($T: typeid) {
	mutex: mutex.Mutex,
	items: [dynamic]T,
}

mpsc_init :: proc(q: ^MPSCQueue($T), allocator := context.allocator) {
	q.items = make([dynamic]T, allocator)
}

mpsc_destroy :: proc(q: ^MPSCQueue($T)) {
	delete(q.items)
}

mpsc_enqueue :: proc(q: ^MPSCQueue($T), value: T) -> int {
	mutex.lock(&q.mutex)
	append(&q.items, value)
	mutex.unlock(&q.mutex)
	return 0
}

mpsc_dequeue :: proc(q: ^MPSCQueue($T)) -> (T, bool) {
	mutex.lock(&q.mutex)
	defer mutex.unlock(&q.mutex)
	if len(q.items) == 0 do return T{}, false
	val := q.items[0]
	ordered_remove(&q.items, 0)
	return val, true
}

// MPMCQueue — Multi-Producer Multi-Consumer (simplified, bounded)

MPMCQueue :: struct($T: typeid) {
	mutex: mutex.Mutex,
	buf:   []T,
	head:  int,
	tail:  int,
	count: int,
	cap:   int,
}

mpmc_init :: proc(q: ^MPMCQueue($T), size: int) {
	q.buf = make([]T, size)
	q.cap = size
	q.head = 0; q.tail = 0; q.count = 0
}

mpmc_destroy :: proc(q: ^MPMCQueue($T)) {
	delete(q.buf)
	q.cap = 0; q.head = 0; q.tail = 0; q.count = 0
}

mpmc_enqueue :: proc(q: ^MPMCQueue($T), data: T) -> i32 {
	mutex.lock(&q.mutex)
	defer mutex.unlock(&q.mutex)
	if q.count >= q.cap do return -1
	q.buf[q.tail] = data
	q.tail = (q.tail + 1) % q.cap
	q.count += 1
	return 0
}

mpmc_dequeue :: proc(q: ^MPMCQueue($T)) -> (T, bool) {
	mutex.lock(&q.mutex)
	defer mutex.unlock(&q.mutex)
	if q.count == 0 do return T{}, false
	val := q.buf[q.head]
	q.head = (q.head + 1) % q.cap
	q.count -= 1
	return val, true
}


// ============================================================================
// Block 3: Containers (RangeCache, PriorityQueue, StringInterner)
// Translated from src/range_cache.cpp, src/priority_queue.cpp, src/string_interner.cpp
// ============================================================================

// --- RangeCache ---

RangeValue :: struct {
	lo: i64,
	hi: i64,
}

RangeCache :: struct {
	ranges: [dynamic]RangeValue,
}

range_cache_make :: proc(allocator := context.allocator) -> RangeCache {
	return RangeCache{ranges = make([dynamic]RangeValue, allocator)}
}

range_cache_destroy :: proc(rc: ^RangeCache) {
	delete(rc.ranges)
	rc^ = {}
}

range_cache_add_index :: proc(rc: ^RangeCache, index: i64) {
	for i in 0..<len(rc.ranges) {
		r := &rc.ranges[i]
		if index == r.lo - 1 {
			r.lo = index
			if i > 0 && rc.ranges[i-1].hi + 1 == r.lo {
				rc.ranges[i-1].hi = r.hi
				ordered_remove(&rc.ranges, i)
			}
			return
		}
		if index == r.hi + 1 {
			r.hi = index
			if i + 1 < len(rc.ranges) && r.hi + 1 == rc.ranges[i+1].lo {
				r.hi = rc.ranges[i+1].hi
				ordered_remove(&rc.ranges, i + 1)
			}
			return
		}
		if index >= r.lo && index <= r.hi do return
		if index < r.lo - 1 {
			inject_at(&rc.ranges, i, RangeValue{lo = index, hi = index})
			return
		}
	}
	append(&rc.ranges, RangeValue{lo = index, hi = index})
}

range_cache_add_range :: proc(rc: ^RangeCache, lo, hi: i64) {
	for index in lo..=hi {
		range_cache_add_index(rc, index)
	}
}

// --- PriorityQueue ---

PriorityQueue :: struct($T: typeid) {
	queue: [dynamic]T,
	cmp:   proc(q: ^T, i, j: int) -> int,
	swap:  proc(q: ^T, i, j: int),
}

@(private)
priority_queue_swap :: proc(pq: ^$P/PriorityQueue($T), i, j: int) {
	if i == j do return
	if pq.swap != nil {
		pq.swap(&pq.queue[0], i, j)
	} else {
		pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	}
}

@(private)
priority_queue_cmp :: proc(pq: ^$P/PriorityQueue($T), i, j: int) -> int {
	if pq.cmp != nil do return pq.cmp(&pq.queue[0], i, j)
	return 0
}

priority_queue_shift_down :: proc(pq: ^$P/PriorityQueue($T), i0: int, n: int) -> bool {
	i := i0
	if i < 0 || i > n do return false
	changed := false
	for {
		j1 := 2*i + 1
		if j1 < 0 || j1 >= n do break
		j := j1
		j2 := j1 + 1
		if j2 < n && priority_queue_cmp(pq, j2, j1) < 0 do j = j2
		if priority_queue_cmp(pq, j, i) >= 0 do break
		priority_queue_swap(pq, i, j)
		i = j
		changed = true
	}
	return changed
}

priority_queue_shift_up :: proc(pq: ^$P/PriorityQueue($T), j: int) {
	for j >= 0 && j < len(pq.queue) {
		i := (j - 1) / 2
		if i == j || priority_queue_cmp(pq, j, i) >= 0 do break
		priority_queue_swap(pq, i, j)
		j = i
	}
}

priority_queue_fix :: proc(pq: ^$P/PriorityQueue($T), i: int) {
	if !priority_queue_shift_down(pq, i, len(pq.queue)) {
		priority_queue_shift_up(pq, i)
	}
}

priority_queue_push :: proc(pq: ^$P/PriorityQueue($T), value: T) {
	append(&pq.queue, value)
	priority_queue_shift_up(pq, len(pq.queue) - 1)
}

priority_queue_pop :: proc(pq: ^$P/PriorityQueue($T)) -> (T, bool) {
	n := len(pq.queue)
	if n == 0 do return T{}, false
	result := pq.queue[0]
	n -= 1
	pq.queue[0] = pq.queue[n]
	priority_queue_shift_down(pq, 0, n)
	pop(&pq.queue)
	return result, true
}

priority_queue_remove :: proc(pq: ^$P/PriorityQueue($T), i: int) -> bool {
	n := len(pq.queue)
	if i < 0 || i >= n do return false
	if i == n - 1 {
		pop(&pq.queue)
		return true
	}
	pq.queue[i] = pq.queue[n - 1]
	pop(&pq.queue)
	priority_queue_fix(pq, i)
	return true
}

priority_queue_create :: proc(cmp: proc(q: ^$T, i, j: int) -> int, swap: proc(q: ^$T, i, j: int) = nil, allocator := context.allocator) -> PriorityQueue(T) {
	return PriorityQueue(T){
		queue = make([dynamic]T, allocator),
		cmp   = cmp,
		swap  = swap,
	}
}

priority_queue_destroy :: proc(pq: ^$P/PriorityQueue($T)) {
	delete(pq.queue)
	pq^ = {}
}

// --- StringInterner ---
// NOTE: Simplified from the lock-free C++ version.
// Uses a mutex-protected map[string]string for string interning.

InternedString :: struct {
	value: u32,
}

INTERN_CELL_CAP :: 8

StringInternCell :: struct {
	hashes:  [INTERN_CELL_CAP]u64,
	offsets: [INTERN_CELL_CAP]u32,
}

StringInterner :: struct {
	entries:     map[string]u32, // string -> unique id
	reverse:     [dynamic]string, // id -> string
	mutex:       mutex.Mutex,
	track_count: bool,
	count:       i64,
}

init_string_interner :: proc(si: ^StringInterner, track_count: bool = false, allocator := context.allocator) {
	si.entries = make(map[string]u32, allocator)
	si.reverse = make([dynamic]string, allocator)
	si.track_count = track_count
	si.count = 0
}

destroy_string_interner :: proc(si: ^StringInterner) {
	delete(si.entries)
	delete(si.reverse)
	si^ = {}
}

string_interner_load :: proc(si: ^StringInterner, s: string) -> InternedString {
	return string_interner_insert(si, s)
}

string_interner_load_cstring :: proc(si: ^StringInterner, s: string) -> InternedString {
	return string_interner_insert(si, s)
}

string_interner_insert :: proc(si: ^StringInterner, s: string) -> InternedString {
	if len(s) == 0 do return InternedString{value = 0}

	mutex.lock(&si.mutex)
	defer mutex.unlock(&si.mutex)

	if id, found := si.entries[s]; found {
		return InternedString{value = id}
	}

	id := u32(len(si.reverse) + 1)
	si.entries[s] = id
	append(&si.reverse, s)

	if si.track_count do si.count += 1
	return InternedString{value = id}
}

string_interner_lookup :: proc(si: ^StringInterner, is: InternedString) -> string {
	mutex.lock(&si.mutex)
	defer mutex.unlock(&si.mutex)
	if is.value == 0 || is.value > u32(len(si.reverse)) do return ""
	return si.reverse[is.value - 1]
}

string_intern_cstring :: proc(si: ^StringInterner, s: string) -> InternedString {
	return string_interner_insert(si, s)
}

string_intern_string :: proc(si: ^StringInterner, s: string) -> InternedString {
	return string_interner_insert(si, s)
}

// --- Thread-local arena helpers (simplified) ---

ThreadLocalArena :: struct {
	data: [dynamic]byte,
	used: int,
}

string_interner_thread_local_arena_init :: proc(arena: ^ThreadLocalArena, size: int, allocator := context.allocator) {
	arena.data = make([dynamic]byte, size, allocator)
	arena.used = 0
}

string_interner_thread_local_arena_alloc :: proc(arena: ^ThreadLocalArena, size: int) -> []byte {
	if arena.used + size > len(arena.data) do return nil
	result := arena.data[arena.used:arena.used + size]
	arena.used += size
	return result
}

string_interner_thread_local_arena_reset :: proc(arena: ^ThreadLocalArena) {
	arena.used = 0
}

string_interner_thread_local_arena_destroy :: proc(arena: ^ThreadLocalArena) {
	delete(arena.data)
	arena^ = {}
}

// --- Obfuscate helpers ---

obfuscate_string :: proc(s: string, prefix: string) -> string {
	if len(s) == 0 do return s
	hash := fnv64a_string(s)
	// Return formatted: prefix + hex hash
	// NOTE: simplified — original uses gb_string_append_fmt
	return strings.concatenate({prefix, "x", u64_to_hex(hash)}, context.temp_allocator)
}

@(private)
u64_to_hex :: proc(v: u64) -> string {
	buf: [18]byte
	i := len(buf)
	if v == 0 {
		i -= 1; buf[i] = '0'
	} else {
		for v > 0 && i > 1 {
			i -= 1; buf[i] = NUM_TO_CHAR_TABLE[v & 0xf]; v >>= 4
		}
	}
	return string(buf[i:])
}

obfuscate_i32 :: proc(v: i32) -> i32 {
	x := cast(i32)fnv64a(transmute([]byte)mem.byte_slice(&v, size_of(v)))
	if x < 0 do return 1 - x
	return x
}


// ============================================================================
// Block 4: Path utilities, File loading, Levenshtein, DidYouMean
// Translated from src/path.cpp
// ============================================================================

// --- Path utilities ---

remove_extension_from_path :: proc(s: string) -> string {
	for i := len(s) - 1; i >= 0; i -= 1 {
		if s[i] == '.' do return s[:i]
	}
	return s
}

remove_directory_from_path :: proc(s: string) -> string {
	for i := len(s) - 1; i >= 0; i -= 1 {
		c := s[i]
		if c == '/' || c == '\\' do return s[i+1:]
	}
	return s
}

directory_from_path :: proc(path: string) -> string {
	for i := len(path) - 1; i >= 0; i -= 1 {
		c := path[i]
		if c == '/' || c == '\\' do return path[:i]
	}
	return ""
}

get_working_directory :: proc(allocator := context.allocator) -> (string, bool) {
	return os.get_current_directory(allocator)
}

set_working_directory :: proc(dir: string) -> bool {
	return os.set_current_directory(dir) == os.ERROR_NONE
}

path_is_directory :: proc(path: string) -> bool {
	return os.is_dir(path)
}

path_to_full_path :: proc(path: string, allocator := context.allocator) -> (string, bool) {
	return os.absolute_path(path, allocator)
}

// --- Path struct ---

Path :: struct {
	basename: string,
	name:     string,
	ext:      string,
}

path_to_string :: proc(p: Path) -> string {
	return p.basename
}

quote_path :: proc(path: string, allocator := context.allocator) -> string {
	if len(path) > 0 && path[0] == '"' do return path
	if strings.contains_rune(path, ' ') {
		return strings.concatenate({"\"", path, "\""}, allocator)
	}
	return path
}

path_from_string :: proc(s: string) -> Path {
	p: Path
	p.basename = s
	p.name = remove_extension_from_path(remove_directory_from_path(s))
	// Find extension
	last_sep := -1
	last_dot := -1
	for i := len(s) - 1; i >= 0; i -= 1 {
		c := s[i]
		if c == '/' || c == '\\' { last_sep = i; break }
		if c == '.' && last_dot < 0 { last_dot = i }
	}
	if last_dot >= 0 && last_dot > last_sep {
		p.ext = s[last_dot:]
	}
	return p
}

last_path_element :: proc(path: string) -> string {
	return remove_directory_from_path(path)
}

// --- FileInfo and ReadDirectoryError ---

FileInfo :: struct {
	name:     string,
	fullpath: string,
	size:     i64,
	is_dir:   bool,
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

read_directory :: proc(path: string, allocator := context.allocator) -> ([dynamic]FileInfo, ReadDirectoryError) {
	entries, err := os.read_dir(path, allocator)
	if err != os.ERROR_NONE {
		switch err {
		case os.ERROR_FILE_NOT_FOUND: return nil, .NotExists
		case os.ERROR_PATH_NOT_FOUND: return nil, .InvalidPath
		case os.ERROR_ACCESS_DENIED:  return nil, .Permission
		case:                         return nil, .Unknown
		}
	}
	result := make([dynamic]FileInfo, 0, len(entries), allocator)
	for entry in entries {
		append(&result, FileInfo{
			name     = entry.name,
			fullpath = filepath.join({path, entry.name}, allocator),
			size     = entry.size,
			is_dir   = entry.is_dir,
		})
	}
	if len(result) == 0 do return result, .Empty
	return result, .None
}

get_file_size :: proc(path: string) -> (i64, bool) {
	s, err := os.stat(path, context.temp_allocator)
	if err != os.ERROR_NONE do return 0, false
	return s.size, true
}

write_directory :: proc(path: string) -> bool {
	return os.make_directory(path) == os.ERROR_NONE
}

// --- LoadedFile ---

LoadedFile :: struct {
	handle: rawptr,
	data:   rawptr,
	size:   i32,
}

LoadedFileError :: enum {
	None,
	Empty,
	FileTooLarge,
	Invalid,
	NotExists,
	Permission,
}

// Simplified file loading using os.read_entire_file.
// Original C++ used memory-mapped files (MapViewOfFile).

load_file :: proc(path: string, allocator := context.allocator) -> (LoadedFile, LoadedFileError) {
	data, ok := os.read_entire_file(path, allocator)
	if !ok {
		if !os.is_file(path) do return LoadedFile{}, .NotExists
		return LoadedFile{}, .Invalid
	}
	if len(data) == 0 do return LoadedFile{}, .Empty
	max_i32: i64 = max(i32)
	if cast(i64)len(data) > max_i32 do return LoadedFile{}, .FileTooLarge
	return LoadedFile{data = raw_data(data), size = cast(i32)len(data)}, .None
}

loaded_file_free :: proc(lf: ^LoadedFile, allocator := context.allocator) {
	if lf.data != nil && lf.size > 0 {
		delete(cast([^]byte)lf.data[:lf.size], allocator)
	}
	lf.data = nil; lf.size = 0; lf.handle = nil
}

// --- Levenshtein distance ---

levenstein_distance_case_insensitive :: proc(a, b: string) -> int {
	a_runes := make([]rune, len(a), context.temp_allocator)
	b_runes := make([]rune, len(b), context.temp_allocator)
	for ch, i in a do a_runes[i] = unicode.to_lower(ch)
	for ch, i in b do b_runes[i] = unicode.to_lower(ch)
	a_len := len(a_runes)
	b_len := len(b_runes)
	if a_len == 0 do return b_len
	if b_len == 0 do return a_len
	prev_row := make([]int, b_len + 1, context.temp_allocator)
	curr_row := make([]int, b_len + 1, context.temp_allocator)
	for j in 0..=b_len do prev_row[j] = j
	for i in 1..=a_len {
		curr_row[0] = i
		for j in 1..=b_len {
			cost := 1 if a_runes[i-1] != b_runes[j-1] else 0
			delete_cost := prev_row[j] + 1
			insert_cost := curr_row[j-1] + 1
			substitute_cost := prev_row[j-1] + cost
			curr_row[j] = min(min(delete_cost, insert_cost), substitute_cost)
		}
		prev_row, curr_row = curr_row, prev_row
	}
	return prev_row[b_len]
}

// --- DidYouMean ---

MAX_SMALLEST_DID_YOU_MEAN_DISTANCE :: 2

DistanceAndTarget :: struct {
	distance: int,
	target:   string,
}

DidYouMeanAnswers :: struct {
	distances: [dynamic]DistanceAndTarget,
	key:       string,
}

did_you_mean_make :: proc(key: string, allocator := context.allocator) -> DidYouMeanAnswers {
	return DidYouMeanAnswers{
		distances = make([dynamic]DistanceAndTarget, allocator),
		key = key,
	}
}

did_you_mean_destroy :: proc(dym: ^DidYouMeanAnswers) {
	delete(dym.distances)
	dym.key = ""
}

did_you_mean_append :: proc(dym: ^DidYouMeanAnswers, target: string) {
	dist := levenstein_distance_case_insensitive(dym.key, target)
	append(&dym.distances, DistanceAndTarget{distance = dist, target = target})
}

did_you_mean_results :: proc(dym: ^DidYouMeanAnswers) -> []DistanceAndTarget {
	if len(dym.distances) == 0 do return nil
	sort.quick_sort(dym.distances[:], proc(i, j: DistanceAndTarget) -> bool {
		return i.distance < j.distance
	})
	smallest := dym.distances[0].distance
	limit := smallest + MAX_SMALLEST_DID_YOU_MEAN_DISTANCE
	count := 0
	for dist in dym.distances {
		if dist.distance <= limit { count += 1 } else { break }
	}
	return dym.distances[:count]
}

// --- Command-line parsing ---
// Simplified Windows CommandLineToArgvW-style parsing.

command_line_to_wargv :: proc(cmd_line: string, allocator := context.allocator) -> []string {
	args := make([dynamic]string, allocator)
	i := 0; n := len(cmd_line)
	for i < n {
		for i < n && (cmd_line[i] == ' ' || cmd_line[i] == '\t') { i += 1 }
		if i >= n do break
		in_quotes := false
		arg_buf := make([dynamic]byte, allocator)
		for i < n {
			c := cmd_line[i]
			if c == '"' {
				// Count backslashes before this quote
				slashes := 0
				k := i - 1
				for k >= 0 && cmd_line[k] == '\\' { slashes += 1; k -= 1 }
				// Add half the backslashes (integer division)
				for _ in 0..<slashes/2 { append(&arg_buf, '\\') }
				if slashes & 1 == 1 {
					// Odd: escaped quote is literal
					append(&arg_buf, '"')
				} else if in_quotes {
					// Even: could be end of quoted block or double-quote escape
					if i + 1 < n && cmd_line[i + 1] == '"' {
						append(&arg_buf, '"'); i += 1
					} else {
						in_quotes = false
					}
				} else {
					in_quotes = true
				}
				i += 1
			} else if !in_quotes && (c == ' ' || c == '\t') {
				break
			} else {
				append(&arg_buf, c)
				i += 1
			}
		}
		append(&args, string(arg_buf[:]))
	}
	return args[:]
}
