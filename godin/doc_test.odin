package godin

import "core:testing"
import "core:mem"
import "core:fmt"

// ============================================================================
// align_formula_int tests
// ============================================================================

@(test)
test_align_formula_int :: proc(t: ^testing.T) {
	expect := testing.expect_value
	expect(t, align_formula_int(0, 4), 0)
	expect(t, align_formula_int(1, 4), 4)
	expect(t, align_formula_int(3, 4), 4)
	expect(t, align_formula_int(4, 4), 4)
	expect(t, align_formula_int(5, 4), 8)
	expect(t, align_formula_int(0, 8), 0)
	expect(t, align_formula_int(1, 8), 8)
	expect(t, align_formula_int(7, 8), 8)
	expect(t, align_formula_int(8, 8), 8)
	expect(t, align_formula_int(9, 8), 16)
}

// ============================================================================
// from_array / from_string tests
// ============================================================================

@(test)
test_from_array :: proc(t: ^testing.T) {
	buf := make([]u8, 256)
	defer delete(buf)

	vals := []u32{10, 20, 30, 40, 50}
	mem.copy(&buf[100], raw_data(vals), size_of(u32) * len(vals))

	arr := OdinDocArray(u32){offset = 100, length = 5}
	result := from_array((^OdinDocHeaderBase)(raw_data(buf)), arr)

	testing.expect(t, result != nil, "result should not be nil")
	testing.expect_value(t, len(result), 5)
	testing.expect_value(t, result[0], u32(10))
	testing.expect_value(t, result[1], u32(20))
	testing.expect_value(t, result[4], u32(50))
}

@(test)
test_from_array_empty :: proc(t: ^testing.T) {
	buf := make([]u8, 64)
	defer delete(buf)

	arr := OdinDocArray(u32){offset = 0, length = 0}
	result := from_array((^OdinDocHeaderBase)(raw_data(buf)), arr)
	testing.expect(t, result == nil, "zero-length array should return nil")
}

@(test)
test_from_string :: proc(t: ^testing.T) {
	buf := make([]u8, 256)
	defer delete(buf)

	test_str := "Hello, Odin Docs!"
	mem.copy(&buf[50], raw_data(test_str), len(test_str))

	s := OdinDocString{offset = 50, length = u32(len(test_str))}
	result := from_string((^OdinDocHeaderBase)(raw_data(buf)), s)
	testing.expect_value(t, result, test_str)
}

@(test)
test_from_string_empty :: proc(t: ^testing.T) {
	buf := make([]u8, 64)
	defer delete(buf)
	s := OdinDocString{offset = 0, length = 0}
	result := from_string((^OdinDocHeaderBase)(raw_data(buf)), s)
	testing.expect_value(t, result, "")
}

@(test)
test_from_string_boundary :: proc(t: ^testing.T) {
	buf := make([]u8, 512)
	defer delete(buf)

	test_data := "edge"
	offset := len(buf) - len(test_data)
	mem.copy(&buf[offset], raw_data(test_data), len(test_data))

	s := OdinDocString{offset = u32(offset), length = u32(len(test_data))}
	result := from_string((^OdinDocHeaderBase)(raw_data(buf)), s)
	testing.expect_value(t, result, test_data)
}

// ============================================================================
// hash_data_after_header tests
// ============================================================================

@(test)
test_hash_data_after_header :: proc(t: ^testing.T) {
	data := make([]u8, 128)
	defer delete(data)

	base := (^OdinDocHeaderBase)(raw_data(data))
	base.header_size = 32
	base.total_size = 128

	for i in 32 ..< 128 {
		data[i] = u8(i - 32)
	}

	h1 := hash_data_after_header(base, raw_data(data), 128)
	testing.expect(t, h1 != 0, "hash should be non-zero")
	h2 := hash_data_after_header(base, raw_data(data), 128)
	testing.expect_value(t, h1, h2)
}

@(test)
test_hash_data_after_header_empty_body :: proc(t: ^testing.T) {
	data := make([]u8, 32)
	defer delete(data)

	base := (^OdinDocHeaderBase)(raw_data(data))
	base.header_size = 32
	base.total_size = 32

	result := hash_data_after_header(base, raw_data(data), 32)
	testing.expect_value(t, result, u32(FNV1A_INIT))
}

// ============================================================================
// Writer: ItemTracker init/size tests
// ============================================================================

@(test)
test_writer_item_tracker_init :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(u32)
	odin_doc_writer_item_tracker_init(&tracker, 100)
	testing.expect_value(t, tracker.len, 100)
	testing.expect_value(t, tracker.cap, 100)
	testing.expect_value(t, tracker.offset, 0)
}

@(test)
test_writer_item_tracker_init_zero :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(OdinDocString)
	odin_doc_writer_item_tracker_init(&tracker, 0)
	testing.expect_value(t, tracker.len, 0)
	testing.expect_value(t, tracker.cap, 0)
	testing.expect_value(t, tracker.offset, 0)
}

@(test)
test_writer_tracker_size :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(u32)
	tracker.cap = 10
	tracker.len = 10
	offset := 64
	odin_doc_writer_tracker_size(&offset, &tracker)
	testing.expect_value(t, offset, 104)
	testing.expect_value(t, tracker.offset, 64)
}

@(test)
test_writer_tracker_size_with_alignment :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(u8)
	tracker.cap = 20
	tracker.len = 20
	offset := 5
	odin_doc_writer_tracker_size(&offset, &tracker, 16)
	testing.expect_value(t, offset, 36)
	testing.expect_value(t, tracker.offset, 16)
}

@(test)
test_writer_tracker_size_struct_align :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(OdinDocEntity)
	tracker.cap = 3
	tracker.len = 3
	offset := 100
	odin_doc_writer_tracker_size(&offset, &tracker)
	testing.expect(t, offset > 100)
	testing.expect(t, tracker.offset >= 100)
	testing.expect(t, tracker.offset % align_of(OdinDocEntity) == 0)
}

// ============================================================================
// Writer: calc_total_size test
// ============================================================================

@(test)
test_calc_total_size :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_item_tracker_init(&w.files, 1)
	odin_doc_writer_item_tracker_init(&w.pkgs, 1)
	odin_doc_writer_item_tracker_init(&w.entities, 1)
	odin_doc_writer_item_tracker_init(&w.types, 1)
	odin_doc_writer_item_tracker_init(&w.strings, 16)
	odin_doc_writer_item_tracker_init(&w.blob, 16)
	total := odin_doc_writer_calc_total_size(&w)
	testing.expect(t, total >= size_of(OdinDocHeader))
}

// ============================================================================
// Writer: prepare / destroy tests
// ============================================================================

@(test)
test_writer_prepare :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	testing.expect_value(t, w.state, OdinDocWriterState.Preparing)
	testing.expect(t, w.string_cache != nil)
	testing.expect(t, w.file_cache != nil)
	testing.expect(t, w.pkg_cache != nil)
	testing.expect(t, w.entity_cache != nil)
	testing.expect(t, w.type_cache != nil)
	testing.expect_value(t, w.files.cap, 1)
	testing.expect_value(t, w.strings.cap, 16)
	testing.expect_value(t, w.blob.cap, 16)
}

// ============================================================================
// Writer: full cycle (prepare → start → end) test
// ============================================================================

@(test)
test_writer_prepare_start_end_cycle :: proc(t: ^testing.T) {
	w: OdinDocWriter
	defer {
		if w.data != nil {
			odin_doc_writer_destroy(&w, context.allocator)
		}
	}

	odin_doc_writer_prepare(&w, context.allocator)
	testing.expect_value(t, w.state, OdinDocWriterState.Preparing)

	odin_doc_writer_start_writing(&w, context.allocator)
	testing.expect_value(t, w.state, OdinDocWriterState.Writing)
	testing.expect(t, w.data != nil)
	testing.expect(t, w.header != nil)
	testing.expect(t, w.data_len > 0)

	odin_doc_writer_end_writing(&w)

	h := w.header
	magic_str := transmute(string)h.base.magic[:]
	testing.expect_value(t, magic_str, ODIN_DOC_MAGIC)
	testing.expect_value(t, h.base.version.major, u8(ODIN_DOC_VERSION_MAJOR))
	testing.expect_value(t, h.base.version.minor, u8(ODIN_DOC_VERSION_MINOR))
	testing.expect_value(t, h.base.version.patch, u8(ODIN_DOC_VERSION_PATCH))
	testing.expect_value(t, h.base.total_size, u32(w.data_len))
	testing.expect_value(t, h.base.header_size, u32(size_of(OdinDocHeader)))

	testing.expect_value(t, h.files.offset, u32(w.files.offset))
	testing.expect_value(t, h.files.length, u32(w.files.len))
	testing.expect_value(t, h.pkgs.offset, u32(w.pkgs.offset))
	testing.expect_value(t, h.pkgs.length, u32(w.pkgs.len))
	testing.expect_value(t, h.entities.offset, u32(w.entities.offset))
	testing.expect_value(t, h.entities.length, u32(w.entities.len))
	testing.expect_value(t, h.types.offset, u32(w.types.offset))
	testing.expect_value(t, h.types.length, u32(w.types.len))
}

// ============================================================================
// Writer: write_item tests
// ============================================================================

@(test)
test_write_item_preparing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	tracker: OdinDocWriterItemTracker(u32)
	odin_doc_writer_item_tracker_init(&tracker, 5)
	initial_cap := tracker.cap

	val := u32(42)
	dst: ^u32
	index := odin_doc_write_item(&w, &tracker, &val, &dst)

	testing.expect_value(t, index, u32(0))
	testing.expect_value(t, tracker.cap, initial_cap + 1)
	testing.expect(t, dst == nil)
}

@(test)
test_write_item_writing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	tracker: OdinDocWriterItemTracker(u32)
	odin_doc_write_item(&w, &tracker, nil)
	odin_doc_write_item(&w, &tracker, nil)
	odin_doc_write_item(&w, &tracker, nil)

	odin_doc_writer_start_writing(&w, context.allocator)
	tracker.len = 3
	tracker.cap = 6
	tracker.offset = 100

	val1, val2 := u32(111), u32(222)
	dst1, dst2: ^u32
	idx1 := odin_doc_write_item(&w, &tracker, &val1, &dst1)
	idx2 := odin_doc_write_item(&w, &tracker, &val2, &dst2)

	testing.expect_value(t, idx1, u32(3))
	testing.expect_value(t, idx2, u32(4))
	testing.expect(t, dst1 != nil)
	testing.expect(t, dst2 != nil)
	testing.expect_value(t, dst1^, u32(111))
	testing.expect_value(t, dst2^, u32(222))
	testing.expect_value(t, tracker.len, 5)
}

@(test)
test_write_item_to_nil :: proc(t: ^testing.T) {
	// Passing nil item is allowed (just zeroes memory)
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	tracker: OdinDocWriterItemTracker(u64)
	_ = odin_doc_write_item(&w, &tracker, nil)

	odin_doc_writer_start_writing(&w, context.allocator)
	tracker.len = 1
	tracker.cap = 3
	tracker.offset = 64

	dst: ^u64
	idx := odin_doc_write_item(&w, &tracker, nil, &dst)
	testing.expect_value(t, idx, u32(1))
	testing.expect(t, dst != nil)
	testing.expect_value(t, dst^, u64(0))
}

// ============================================================================
// Writer: get_item tests
// ============================================================================

@(test)
test_get_item_writing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	tracker: OdinDocWriterItemTracker(u32)
	odin_doc_write_item(&w, &tracker, nil)
	odin_doc_write_item(&w, &tracker, nil)

	odin_doc_writer_start_writing(&w, context.allocator)
	tracker.len = 2
	tracker.cap = 4
	tracker.offset = 200

	v1, v2 := u32(777), u32(888)
	odin_doc_write_item(&w, &tracker, &v1)
	odin_doc_write_item(&w, &tracker, &v2)

	item0 := odin_doc_get_item(&w, &tracker, 2)
	item1 := odin_doc_get_item(&w, &tracker, 3)

	testing.expect(t, item0 != nil)
	testing.expect(t, item1 != nil)
	testing.expect_value(t, item0^, u32(777))
	testing.expect_value(t, item1^, u32(888))
}

@(test)
test_get_item_not_writing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	tracker: OdinDocWriterItemTracker(u32)
	tracker.len = 5
	tracker.cap = 10

	item := odin_doc_get_item(&w, &tracker, 2)
	testing.expect(t, item == nil)
}

// ============================================================================
// Writer: string write tests
// ============================================================================

@(test)
test_write_string_without_cache_preparing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	res := odin_doc_write_string_without_cache(&w, "hello")
	testing.expect_value(t, res.offset, u32(0))
	testing.expect_value(t, res.length, u32(0))
	testing.expect(t, w.strings.cap > 16, "strings cap should have grown")
}

@(test)
test_write_string_without_cache_writing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	_ = odin_doc_write_string_without_cache(&w, "hello")
	_ = odin_doc_write_string_without_cache(&w, "world")

	odin_doc_writer_start_writing(&w, context.allocator)
	w.strings.offset = 100
	w.strings.len = 0
	w.strings.cap = 32

	r1 := odin_doc_write_string_without_cache(&w, "hello")
	r2 := odin_doc_write_string_without_cache(&w, "world")

	testing.expect_value(t, r1.offset, u32(100))
	testing.expect_value(t, r1.length, u32(5))
	testing.expect_value(t, r2.offset, u32(100 + 5 + 1))
	testing.expect_value(t, r2.length, u32(5))

	s1 := from_string(&w.header.base, r1)
	s2 := from_string(&w.header.base, r2)
	testing.expect_value(t, s1, "hello")
	testing.expect_value(t, s2, "world")
}

@(test)
test_write_string_with_cache :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	_ = odin_doc_write_string(&w, "cached")
	_ = odin_doc_write_string(&w, "cached")

	odin_doc_writer_start_writing(&w, context.allocator)
	w.strings.offset = 0
	w.strings.len = 0
	w.strings.cap = 64

	r1 := odin_doc_write_string(&w, "cached")
	r2 := odin_doc_write_string(&w, "cached")

	testing.expect_value(t, r1.offset, r2.offset)
	testing.expect_value(t, r1.length, r2.length)
	testing.expect_value(t, len(w.string_cache), 1)
}

// ============================================================================
// Writer: blob/slice write tests
// ============================================================================

@(test)
test_write_slice_preparing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	res := odin_write_slice(&w, []u32{1, 2, 3, 4, 5})
	testing.expect_value(t, res.offset, u32(0))
	testing.expect_value(t, res.length, u32(0))
	testing.expect(t, w.blob.cap > 16, "blob cap should have grown")
}

@(test)
test_write_slice_writing :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	data := []u32{10, 20, 30}
	_ = odin_write_slice(&w, data)

	odin_doc_writer_start_writing(&w, context.allocator)
	w.blob.offset = 50
	w.blob.len = 0
	w.blob.cap = 64

	res := odin_write_slice(&w, data)
	testing.expect_value(t, res.length, u32(3))

	arr := from_array(&w.header.base, res)
	testing.expect(t, arr != nil)
	testing.expect_value(t, len(arr), 3)
	testing.expect_value(t, arr[0], u32(10))
	testing.expect_value(t, arr[1], u32(20))
	testing.expect_value(t, arr[2], u32(30))
}

@(test)
test_write_slice_empty :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	odin_doc_writer_start_writing(&w, context.allocator)

	res := odin_write_slice(&w, []u32{})
	testing.expect_value(t, res.offset, u32(0))
	testing.expect_value(t, res.length, u32(0))
}

@(test)
test_write_slice_nil :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	odin_doc_writer_start_writing(&w, context.allocator)

	empty: []u32 = nil
	res := odin_write_slice(&w, empty)
	testing.expect_value(t, res.offset, u32(0))
	testing.expect_value(t, res.length, u32(0))
}

@(test)
test_write_slice_alignment :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	_ = odin_write_slice(&w, []u32{1})
	odin_doc_writer_start_writing(&w, context.allocator)

	w.blob.offset = 7 // deliberately unaligned
	w.blob.len = 0
	w.blob.cap = 64

	res := odin_write_slice(&w, []u32{1})
	testing.expect_value(t, res.offset % 4, u32(0))
}

// ============================================================================
// Writer: write_item_as_slice test
// ============================================================================

@(test)
test_write_item_as_slice :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	_ = odin_write_item_as_slice(&w, u32(999))
	odin_doc_writer_start_writing(&w, context.allocator)
	w.blob.offset = 10
	w.blob.len = 0
	w.blob.cap = 64

	res := odin_write_item_as_slice(&w, u32(999))
	testing.expect_value(t, res.length, u32(1))
	arr := from_array(&w.header.base, res)
	testing.expect_value(t, arr[0], u32(999))
}

// ============================================================================
// Writer: assign_tracker test
// ============================================================================

@(test)
test_assign_tracker :: proc(t: ^testing.T) {
	tracker: OdinDocWriterItemTracker(OdinDocFile)
	tracker.offset = 128
	tracker.len = 5

	arr: OdinDocArray(OdinDocFile)
	odin_doc_writer_assign_tracker(&arr, tracker)

	testing.expect_value(t, arr.offset, u32(128))
	testing.expect_value(t, arr.length, u32(5))
}

// ============================================================================
// Writer: multiple items in sequence
// ============================================================================

@(test)
test_writer_multiple_items_sequence :: proc(t: ^testing.T) {
	w: OdinDocWriter
	odin_doc_writer_prepare(&w, context.allocator)
	defer odin_doc_writer_destroy(&w, context.allocator)

	t1: OdinDocWriterItemTracker(u32)
	t2: OdinDocWriterItemTracker(u64)

	for _ in 0 ..< 5 { odin_doc_write_item(&w, &t1, nil) }
	for _ in 0 ..< 3 { odin_doc_write_item(&w, &t2, nil) }

	odin_doc_writer_start_writing(&w, context.allocator)
	t1.len, t1.cap, t1.offset = 5, 7, 100
	t2.len, t2.cap, t2.offset = 3, 5, 200

	for i in 0 ..< 5 {
		val := u32(i * 10)
		odin_doc_write_item(&w, &t1, &val)
	}
	for i in 0 ..< 3 {
		val := u64(i * 100)
		odin_doc_write_item(&w, &t2, &val)
	}

	it0 := odin_doc_get_item(&w, &t1, 5)
	it4 := odin_doc_get_item(&w, &t1, 9)
	testing.expect_value(t, it0^, u32(0))
	testing.expect_value(t, it4^, u32(40))

	it2_0 := odin_doc_get_item(&w, &t2, 3)
	it2_2 := odin_doc_get_item(&w, &t2, 5)
	testing.expect_value(t, it2_0^, u64(0))
	testing.expect_value(t, it2_2^, u64(200))
}

// ============================================================================
// Structure layout / enum value tests
// ============================================================================

@(test)
test_odin_doc_version_type_size :: proc(t: ^testing.T) {
	testing.expect_value(t, size_of(OdinDocVersionType), 4)
}

@(test)
test_odin_doc_position_size :: proc(t: ^testing.T) {
	testing.expect_value(t, size_of(OdinDocPosition), 16)
}

@(test)
test_odin_doc_type_kind_values :: proc(t: ^testing.T) {
	expect := testing.expect_value
	expect(t, int(OdinDocTypeKind.Invalid), 0)
	expect(t, int(OdinDocTypeKind.Basic), 1)
	expect(t, int(OdinDocTypeKind.Named), 2)
	expect(t, int(OdinDocTypeKind.Generic), 3)
	expect(t, int(OdinDocTypeKind.Pointer), 4)
	expect(t, int(OdinDocTypeKind.Struct), 10)
	expect(t, int(OdinDocTypeKind.Union), 11)
	expect(t, int(OdinDocTypeKind.Enum), 12)
	expect(t, int(OdinDocTypeKind.Proc), 14)
	expect(t, int(OdinDocTypeKind.MultiPointer), 22)
	expect(t, int(OdinDocTypeKind.Matrix), 23)
	expect(t, int(OdinDocTypeKind.BitField), 25)
	expect(t, int(OdinDocTypeKind.FixedCapacityDynamicArray), 26)
}

@(test)
test_odin_doc_entity_kind_values :: proc(t: ^testing.T) {
	expect := testing.expect_value
	expect(t, int(OdinDocEntityKind.Invalid), 0)
	expect(t, int(OdinDocEntityKind.Constant), 1)
	expect(t, int(OdinDocEntityKind.Variable), 2)
	expect(t, int(OdinDocEntityKind.TypeName), 3)
	expect(t, int(OdinDocEntityKind.Procedure), 4)
	expect(t, int(OdinDocEntityKind.ProcGroup), 5)
	expect(t, int(OdinDocEntityKind.ImportName), 6)
	expect(t, int(OdinDocEntityKind.LibraryName), 7)
	expect(t, int(OdinDocEntityKind.Builtin), 8)
}

@(test)
test_odin_doc_entity_flags_no_overlap :: proc(t: ^testing.T) {
	pairs := [][2]u64{
		{u64(OdinDocEntityFlag.Foreign), u64(OdinDocEntityFlag.Export)},
		{u64(OdinDocEntityFlag.Param_Using), u64(OdinDocEntityFlag.Param_Const)},
		{u64(OdinDocEntityFlag.Param_Ellipsis), u64(OdinDocEntityFlag.Param_CVararg)},
		{u64(OdinDocEntityFlag.Var_Thread_Local), u64(OdinDocEntityFlag.Var_Static)},
		{u64(OdinDocEntityFlag.Builtin_Pkg_Builtin), u64(OdinDocEntityFlag.Builtin_Pkg_Intrinsics)},
	}
	for pair, i in pairs {
		testing.expect(t, (pair[0] & pair[1]) == 0,
			fmt.tprintf("flag collision at pair %d", i))
	}
}

@(test)
test_odin_doc_pkg_flags_values :: proc(t: ^testing.T) {
	testing.expect(t, u32(OdinDocPkgFlags.Builtin) != 0)
	testing.expect(t, u32(OdinDocPkgFlags.Runtime) != 0)
	testing.expect(t, u32(OdinDocPkgFlags.Init) != 0)
	testing.expect(t, (u32(OdinDocPkgFlags.Builtin) & u32(OdinDocPkgFlags.Runtime)) == 0)
	testing.expect(t, (u32(OdinDocPkgFlags.Builtin) & u32(OdinDocPkgFlags.Init)) == 0)
}

// ============================================================================
// Structure sizes (informational — ensures structures have expected non-zero size)
// ============================================================================

@(test)
test_structure_sizes :: proc(t: ^testing.T) {
	testing.expect(t, size_of(OdinDocHeaderBase) > 0)
	testing.expect(t, size_of(OdinDocHeader) > size_of(OdinDocHeaderBase))
	testing.expect(t, size_of(OdinDocFile) > 0)
	testing.expect(t, size_of(OdinDocPkg) > 0)
	testing.expect(t, size_of(OdinDocEntity) > 0)
	testing.expect(t, size_of(OdinDocType) > 0)
	testing.expect(t, size_of(OdinDocScopeEntry) > 0)
	testing.expect(t, size_of(OdinDocAttribute) > 0)
}

// ============================================================================
// Writer state strings
// ============================================================================

@(test)
test_writer_state_strings :: proc(t: ^testing.T) {
	testing.expect_value(t, ODIN_DOC_WRITER_STATE_STRINGS[.Preparing], "preparing")
	testing.expect_value(t, ODIN_DOC_WRITER_STATE_STRINGS[.Writing], "writing  ")
}

// ============================================================================
// is_in_doc_writer
// ============================================================================

@(test)
test_is_in_doc_writer_no_crash :: proc(t: ^testing.T) {
	// Just verify the function doesn't crash
	_ = is_in_doc_writer()
}