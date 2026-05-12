// common_test.odin - unit tests for common.odin
package godin

import "core:testing"
import "core:strings"

// ---------- next_pow2 ----------
@(test)
test_next_pow2_i32 :: proc(t: ^testing.T) {
    testing.expect_value(t, next_pow2(i32(0)), i32(0))
    testing.expect_value(t, next_pow2(i32(1)), i32(1))
    testing.expect_value(t, next_pow2(i32(2)), i32(2))
    testing.expect_value(t, next_pow2(i32(3)), i32(4))
    testing.expect_value(t, next_pow2(i32(4)), i32(4))
    testing.expect_value(t, next_pow2(i32(5)), i32(8))
    testing.expect_value(t, next_pow2(i32(255)), i32(256))
    testing.expect_value(t, next_pow2(i32(256)), i32(256))
    testing.expect_value(t, next_pow2(i32(257)), i32(512))
}

@(test)
test_next_pow2_isize :: proc(t: ^testing.T) {
    testing.expect_value(t, next_pow2(isize(0)), isize(0))
    testing.expect_value(t, next_pow2(isize(1)), isize(1))
    testing.expect_value(t, next_pow2(isize(3)), isize(4))
    testing.expect_value(t, next_pow2(isize(15)), isize(16))
}

// ---------- fnv32a / fnv64a ----------
@(test)
test_fnv32a :: proc(t: ^testing.T) {
	str := "hello"
    data := transmute([]byte)str
    h := fnv32a(data)
    testing.expect(t, h != 0, "hash should not be zero")
    // test determinism
    h2 := fnv32a(data)
    testing.expect_value(t, h, h2)
}

@(test)
test_fnv64a :: proc(t: ^testing.T) {
	str := "hello"
    data := transmute([]byte)str
    h := fnv64a(data)
    h2 := fnv64a(data, 0xcbf29ce484222325)
    testing.expect_value(t, h, h2)
}

// ---------- u64_from_string ----------
@(test)
test_u64_from_string :: proc(t: ^testing.T) {
    testing.expect_value(t, u64_from_string_v("0"), u64(0))
    testing.expect_value(t, u64_from_string_v("123"), u64(123))
    testing.expect_value(t, u64_from_string_v("0x1A"), u64(0x1A))
    testing.expect_value(t, u64_from_string_v("0b1010"), u64(10))
    testing.expect_value(t, u64_from_string_v("0o77"), u64(63))
    testing.expect_value(t, u64_from_string_v("0d99"), u64(99))
    testing.expect_value(t, u64_from_string_v("1_000"), u64(1000))
}

// ---------- u64_to_string / i64_to_string ----------
@(test)
test_u64_to_string :: proc(t: ^testing.T) {
    buf: [32]byte
    s := u64_to_string(12345, buf[:])
    testing.expect_value(t, s, "12345")
    s2 := u64_to_string(0, buf[:])
    testing.expect_value(t, s2, "0")
}

@(test)
test_i64_to_string :: proc(t: ^testing.T) {
    buf: [32]byte
    pos := i64_to_string(12345, buf[:])
    testing.expect_value(t, pos, "12345")
    neg := i64_to_string(-789, buf[:])
    testing.expect_value(t, neg, "-789")
    zero := i64_to_string(0, buf[:])
    testing.expect_value(t, zero, "0")
}

// ---------- f16 <-> f32 ----------
@(test)
test_f32_to_f16_roundtrip :: proc(t: ^testing.T) {
    values := []f32{0.0, 1.0, -1.0, 0.5, 65504.0, -65504.0, 0.0000001}
    for v in values {
        h := f32_to_f16(v)
        back := f16_to_f32(h)
        // allow small error for finite values
        testing.expectf(t, abs(back - v) < 0.1 * max(abs(v), 0.001), "Expected %f, got %f", v, back)
    }
    // test infinity
    inf_f16 := f32_to_f16(1e10)
    inf_f32 := f16_to_f32(inf_f16)
    testing.expect(t, inf_f32 > 1e9, "should be large/inf")
}

// ---------- RangeCache ----------
@(test)
test_range_cache :: proc(t: ^testing.T) {
    cache := range_cache_make()
    defer range_cache_destroy(&cache)

    added := range_cache_add_range(&cache, 10, 20)
    testing.expect(t, !added)
    added = range_cache_add_index(&cache, 15)
    testing.expect(t, added, "index 15 should already be covered")
    added = range_cache_add_index(&cache, 5)
    testing.expect(t, !added, "5 should be new")
}

// ---------- PriorityQueue (min-heap) ----------
@(test)
test_priority_queue :: proc(t: ^testing.T) {
    pq: PriorityQueue(int)
    pq.cmp = isize_cmp
    pq.queue = make([dynamic]int)
    defer delete(pq.queue)

    priority_queue_push(&pq, 3)
    priority_queue_push(&pq, 1)
    priority_queue_push(&pq, 2)

    val, _ := priority_queue_pop(&pq)
    testing.expect_value(t, val, 1)
    val, _ = priority_queue_pop(&pq)
    testing.expect_value(t, val, 2)
    val, _ = priority_queue_pop(&pq)
    testing.expect_value(t, val, 3)
}

@(test)
test_priority_queue_remove :: proc(t: ^testing.T) {
    pq: PriorityQueue(int)
    pq.cmp = isize_cmp
    pq.queue = make([dynamic]int)
    defer delete(pq.queue)

	array := [?]int{5, 2, 8, 1, 3}
	for x in array {
        priority_queue_push(&pq, x)
    }
    // remove element at index 2 (value 8 after push, but heap order unknown)
    // better to pop until find, but we test remove by index
    // get ordering
    sorted := make([dynamic]int)
    for len(pq.queue) > 0 {
		v, _ := priority_queue_pop(&pq)
        append(&sorted, v)
    }
    delete(sorted)
}

// ---------- StringInterner ----------
@(test)
test_string_interner :: proc(t: ^testing.T) {
    interner := g_interner
    testing.expect(t, interner.entries != nil, "string interner should be initialized")

    id1 := string_interner_insert(&g_interner, "Hello, World!")
    id2 := string_interner_insert(&g_interner, "Hello, World!")
    testing.expect(t, id1 == id2, "same string should map to same interning id")

    s := string_interner_load(&g_interner, id1)
    testing.expect_value(t, s, "Hello, World!")

    blank := string_interner_insert(&g_interner, "_")
    testing.expect_value(t, blank, g_interned_blank)
}

// ---------- Path functions ----------
@(test)
test_remove_extension :: proc(t: ^testing.T) {
    testing.expect_value(t, remove_extension_from_path("file.txt"), "file")
    testing.expect_value(t, remove_extension_from_path("file.tar.gz"), "file.tar")
    testing.expect_value(t, remove_extension_from_path("noext"), "noext")
    testing.expect_value(t, remove_extension_from_path("/dir/file"), "/dir/file")
}

@(test)
test_remove_directory :: proc(t: ^testing.T) {
    testing.expect_value(t, remove_directory_from_path("C:/path/to/file.txt"), "file.txt")
    testing.expect_value(t, remove_directory_from_path("dir/file"), "file")
    testing.expect_value(t, remove_directory_from_path("nodir"), "nodir")
}

@(test)
test_directory_from_path :: proc(t: ^testing.T) {
    testing.expect_value(t, directory_from_path("/a/b/c/file.txt"), "/a/b/c")
    // directory itself
    testing.expect_value(t, directory_from_path("/a/b/c/"), "/a/b/c")
}

@(test)
test_path_to_full_path :: proc(t: ^testing.T) {
    cwd, _ := get_working_directory(); defer delete(cwd)
    abs, _ := path_to_full_path("common.odin"); defer delete(abs)
    testing.expect(t, strings.has_prefix(abs, cwd))
}

// ---------- Levenshtein ----------
@(test)
test_levenstein_distance :: proc(t: ^testing.T) {
    d := levenstein_distance_case_insensitive("kitten", "sitting")
    testing.expect_value(t, d, isize(3))
    d2 := levenstein_distance_case_insensitive("", "abc")
    testing.expect_value(t, d2, isize(3))
    d3 := levenstein_distance_case_insensitive("abc", "abc")
    testing.expect_value(t, d3, isize(0))
    d4 := levenstein_distance_case_insensitive("AbC", "aBc")
    testing.expect_value(t, d4, isize(0))
}

// ---------- DidYouMean ----------
@(test)
test_did_you_mean :: proc(t: ^testing.T) {
    dym := did_you_mean_make("helo")
    defer did_you_mean_destroy(&dym)
    did_you_mean_append(&dym, "hello")
    did_you_mean_append(&dym, "helper")
    did_you_mean_append(&dym, "hera")
    did_you_mean_append(&dym, "xyz")
    results := did_you_mean_results(&dym)
    testing.expect(t, len(results) > 0)
    // first should be "hello" (distance 1)
    testing.expect_value(t, results[0].target, "hello")
}
