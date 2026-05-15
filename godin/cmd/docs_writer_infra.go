package cmd

import (
	"unsafe"
)

func odin_doc_writer_item_tracker_init[T any](t *OdinDocWriterItemTracker[T], size isize) {
	t.Len = size
	t.Cap = size
}

func odin_doc_writer_prepare(w *OdinDocWriter) {
	debugf("odin_doc_writer_prepare\n")
	w.State = OdinDocWriterStatePreparing
	string_map_init(&w.StringCache)
	map_init(&w.FileCache, 1<<10)
	map_init(&w.PkgCache, 1<<10)
	map_init(&w.EntityCache, 1<<18)
	map_init(&w.TypeCache, 1<<18)
	odin_doc_writer_item_tracker_init(&w.Files, 1)
	odin_doc_writer_item_tracker_init(&w.Pkgs, 1)
	odin_doc_writer_item_tracker_init(&w.Entities, 1)
	odin_doc_writer_item_tracker_init(&w.Types, 1)
	odin_doc_writer_item_tracker_init(&w.Strings, 16)
	odin_doc_writer_item_tracker_init(&w.Blob, 16)
}

func odin_doc_writer_destroy(w *OdinDocWriter) {
	debugf("odin_doc_writer_destroy\n")
	w.Data = nil
	string_map_destroy(&w.StringCache)
	map_destroy(&w.FileCache)
	map_destroy(&w.PkgCache)
	map_destroy(&w.EntityCache)
	map_destroy(&w.TypeCache)
}

func odin_doc_writer_tracker_size[T any](offset *isize, t *OdinDocWriterItemTracker[T], alignment isize) {
	size := t.Cap * isize(unsafe.Sizeof(*new(T)))
	align := max(alignment, isize(unsafe.Alignof(*new(T))))
	*offset = align_formula_isize(*offset, align)
	t.Offset = *offset
	*offset += size
}

func odin_doc_writer_calc_total_size(w *OdinDocWriter) isize {
	total_size := isize(unsafe.Sizeof(OdinDocHeader{}))
	odin_doc_writer_tracker_size(&total_size, &w.Files)
	odin_doc_writer_tracker_size(&total_size, &w.Pkgs)
	odin_doc_writer_tracker_size(&total_size, &w.Entities)
	odin_doc_writer_tracker_size(&total_size, &w.Types)
	odin_doc_writer_tracker_size(&total_size, &w.Strings, 16)
	odin_doc_writer_tracker_size(&total_size, &w.Blob, 16)
	return total_size
}

func odin_doc_writer_start_writing(w *OdinDocWriter) {
	debugf("odin_doc_writer_start_writing\n")
	w.State = OdinDocWriterStateWriting
	string_map_clear(&w.StringCache)
	map_clear(&w.FileCache)
	map_clear(&w.PkgCache)
	map_clear(&w.EntityCache)
	map_clear(&w.TypeCache)
	total_size := odin_doc_writer_calc_total_size(w)
	total_size = align_formula_isize(total_size, 8)
	w.Data = make([]byte, total_size)
	w.DataLen = total_size
	w.Header = (*OdinDocHeader)(unsafe.Pointer(&w.Data[0]))
}

func hash_data_after_header(base *OdinDocHeaderBase, data []byte, data_len isize) u32 {
	h := u32(0x811c9dc5)
	for i := isize(base.HeaderSize); i < isize(base.TotalSize); i++ {
		h = (h ^ u32(data[i])) * u32(0x01000193)
	}
	return h
}

func odin_doc_writer_assign_tracker[T any](array *OdinDocArray[T], t *OdinDocWriterItemTracker[T]) {
	array.Offset = u32(t.Offset)
	array.Length = u32(t.Len)
}

func odin_doc_writer_end_writing(w *OdinDocWriter) {
	debugf("odin_doc_writer_end_writing\n")
	h := w.Header
	copy(h.Base.Magic[:], "odindoc")
	h.Base.Version.Major = 0
	h.Base.Version.Minor = 3
	h.Base.Version.Patch = 2
	h.Base.TotalSize = u32(w.DataLen)
	h.Base.HeaderSize = u32(unsafe.Sizeof(OdinDocHeader{}))
	h.Base.Hash = hash_data_after_header(&h.Base, w.Data, w.DataLen)
	odin_doc_writer_assign_tracker(&h.Files, &w.Files)
	odin_doc_writer_assign_tracker(&h.Pkgs, &w.Pkgs)
	odin_doc_writer_assign_tracker(&h.Entities, &w.Entities)
	odin_doc_writer_assign_tracker(&h.Types, &w.Types)
}

func odin_doc_write_item[T any](w *OdinDocWriter, t *OdinDocWriterItemTracker[T], item *T, dst **T) u32 {
	if w.State == OdinDocWriterStatePreparing {
		t.Cap += 1
		if dst != nil {
			*dst = nil
		}
		return 0
	}
	item_index := t.Len
	t.Len++
	basePtr := unsafe.Pointer(&w.Data[0])
	offset := t.Offset + unsafe.Sizeof(*new(T))*item_index
	dataPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	if item != nil {
		gb_memmove(dataPtr, unsafe.Pointer(item), isize(unsafe.Sizeof(*new(T))))
	}
	if dst != nil {
		*dst = (*T)(dataPtr)
	}
	return u32(item_index)
}

func odin_doc_get_item[T any](w *OdinDocWriter, t *OdinDocWriterItemTracker[T], index u32) *T {
	if w.State != OdinDocWriterStateWriting {
		return nil
	}
	basePtr := unsafe.Pointer(&w.Data[0])
	offset := t.Offset + unsafe.Sizeof(*new(T))*isize(index)
	return (*T)(unsafe.Pointer(uintptr(basePtr) + uintptr(offset)))
}

func odin_doc_write_string_without_cache(w *OdinDocWriter, str String) OdinDocString {
	var res OdinDocString
	if w.State == OdinDocWriterStatePreparing {
		w.Strings.Cap += len(str) + 1
	} else {
		offset := w.Strings.Offset + w.Strings.Len
		basePtr := unsafe.Pointer(&w.Data[0])
		dataPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
		gb_memmove(dataPtr, unsafe.Pointer(unsafe.StringData(str)), len(str))
		*(*byte)(unsafe.Pointer(uintptr(dataPtr) + uintptr(len(str)))) = 0
		w.Strings.Len += len(str) + 1
		res.Offset = u32(offset)
		res.Length = u32(len(str))
	}
	return res
}

func odin_doc_write_string(w *OdinDocWriter, str String) OdinDocString {
	c := string_map_get(&w.StringCache, str)
	if c != nil {
		return *c
	}
	res := odin_doc_write_string_without_cache(w, str)
	string_map_set(&w.StringCache, str, res)
	return res
}

func odin_write_slice[T any](w *OdinDocWriter, data *T, len isize) OdinDocArray[T] {
	if len <= 0 {
		return OdinDocArray[T]{}
	}
	alignment := isize(4)
	if w.State == OdinDocWriterStatePreparing {
		w.Blob.Cap = align_formula_isize(w.Blob.Cap, alignment)
		w.Blob.Cap += len * isize(unsafe.Sizeof(*new(T)))
		return OdinDocArray[T]{}
	}
	w.Blob.Len = align_formula_isize(w.Blob.Len, alignment)
	offset := w.Blob.Offset + w.Blob.Len
	basePtr := unsafe.Pointer(&w.Data[0])
	dstPtr := unsafe.Pointer(uintptr(basePtr) + uintptr(offset))
	gb_memmove(dstPtr, unsafe.Pointer(data), len*isize(unsafe.Sizeof(*new(T))))
	w.Blob.Len += len * isize(unsafe.Sizeof(*new(T)))
	return OdinDocArray[T]{Offset: u32(offset), Length: u32(len)}
}

func odin_write_item_as_slice[T any](w *OdinDocWriter, data T) OdinDocArray[T] {
	return odin_write_slice(w, &data, 1)
}

func from_array[T any](base *OdinDocHeaderBase, a OdinDocArray[T]) Slice[T] {
	var s Slice[T]
	s.Data = (*T)(unsafe.Pointer(uintptr(unsafe.Pointer(base)) + uintptr(a.Offset)))
	s.Count = isize(a.Length)
	return s
}

func from_string(base *OdinDocHeaderBase, s OdinDocString) String {
	ptr := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(base)) + uintptr(s.Offset)))
	return unsafe.String(ptr, int(s.Length))
}
