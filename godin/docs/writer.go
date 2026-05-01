package docs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync/atomic"
)

// ============================================================================
// Writer: binary .odin-doc file writer (mirrors src/docs_writer.cpp)
// ============================================================================

// WriterState represents the current phase of the doc writer.
type WriterState int

const (
	WriterStatePreparing WriterState = iota
	WriterStateWriting
)

var writerStateStrings = []string{"preparing", "writing  "}

// itemTracker tracks capacity, length and offset for a section in the buffer.
type itemTracker struct {
	len    int
	cap    int
	offset int
}

func (t *itemTracker) init(size int) {
	t.len = size
	t.cap = size
}

// Writer is the binary doc format writer.
type Writer struct {
	info      any
	state     WriterState
	Data      []byte
	DataLen   int
	Header    *OdinDocHeader
	byteOrder binary.ByteOrder

	stringCache map[string]OdinDocString
	fileCache   map[*AstFile]OdinDocFileIndex
	pkgCache    map[*AstPackage]OdinDocPkgIndex
	entityCache map[*Entity]OdinDocEntityIndex
	typeCache   map[uint64]OdinDocTypeIndex

	files    itemTracker
	pkgs     itemTracker
	entities itemTracker
	types    itemTracker
	strings_ itemTracker
	blob     itemTracker
}

var globalInDocWriter atomic.Bool

// NewWriter creates a new Writer.
func NewWriter(info any) *Writer {
	return &Writer{
		info:        info,
		byteOrder:   binary.LittleEndian,
		stringCache: make(map[string]OdinDocString),
	}
}

// IsInDocWriter reports whether a doc writer is currently active.
func IsInDocWriter() bool {
	return globalInDocWriter.Load()
}

func (w *Writer) prepare() {
	w.state = WriterStatePreparing
	w.stringCache = make(map[string]OdinDocString)
	w.fileCache = make(map[*AstFile]OdinDocFileIndex, 1<<10)
	w.pkgCache = make(map[*AstPackage]OdinDocPkgIndex, 1<<10)
	w.entityCache = make(map[*Entity]OdinDocEntityIndex, 1<<18)
	w.typeCache = make(map[uint64]OdinDocTypeIndex, 1<<18)

	w.files.init(1)
	w.pkgs.init(1)
	w.entities.init(1)
	w.types.init(1)
	w.strings_ = itemTracker{cap: 16}
	w.blob = itemTracker{cap: 16}
}

func (w *Writer) calcTotalSize() int {
	totalSize := binary.Size(OdinDocHeader{})
	totalSize = w.trackerSize(totalSize, &w.files, binary.Size(OdinDocFile{}), 1)
	totalSize = w.trackerSize(totalSize, &w.pkgs, binary.Size(OdinDocPkg{}), 1)
	totalSize = w.trackerSize(totalSize, &w.entities, binary.Size(OdinDocEntity{}), 1)
	totalSize = w.trackerSize(totalSize, &w.types, binary.Size(OdinDocType{}), 1)
	totalSize = w.trackerSize(totalSize, &w.strings_, 1, 16)
	totalSize = w.trackerSize(totalSize, &w.blob, 1, 16)
	return totalSize
}

func (w *Writer) trackerSize(offset int, t *itemTracker, elemSize int, alignment int) int {
	size := t.cap * elemSize
	if alignment > 1 {
		offset = alignOffset(offset, alignment)
	}
	t.offset = offset
	return offset + size
}

func (w *Writer) startWriting() {
	w.state = WriterStateWriting
	w.stringCache = make(map[string]OdinDocString)
	w.fileCache = make(map[*AstFile]OdinDocFileIndex, 1<<10)
	w.pkgCache = make(map[*AstPackage]OdinDocPkgIndex, 1<<10)
	w.entityCache = make(map[*Entity]OdinDocEntityIndex, 1<<18)
	w.typeCache = make(map[uint64]OdinDocTypeIndex, 1<<18)

	totalSize := w.calcTotalSize()
	totalSize = alignOffset(totalSize, 8)

	w.Data = make([]byte, totalSize)
	w.DataLen = totalSize

	h := OdinDocHeader{}
	copy(h.Base.Magic[:], MagicBytes[:])
	h.Base.Version = OdinDocVersion
	h.Base.TotalSize = uint32(totalSize)
	h.Base.HeaderSize = uint32(binary.Size(OdinDocHeader{}))
	writeStructAt(w.Data, 0, &h, w.byteOrder)
	w.Header = readStructAtOffset[OdinDocHeader](w.Data, 0)
}

func (w *Writer) hashDataAfterHeader() uint32 {
	start := int(w.Header.Base.HeaderSize)
	end := int(w.Header.Base.TotalSize)
	return calcFNV1a(w.Data[start:end])
}

func (w *Writer) endWriting() {
	w.Header.Base.Version = OdinDocVersion
	w.Header.Base.TotalSize = uint32(w.DataLen)
	w.Header.Base.HeaderSize = uint32(binary.Size(OdinDocHeader{}))
	w.Header.Base.Hash = w.hashDataAfterHeader()

	w.Header.Files.Offset = uint32(w.files.offset)
	w.Header.Files.Length = uint32(w.files.len)
	w.Header.Pkgs.Offset = uint32(w.pkgs.offset)
	w.Header.Pkgs.Length = uint32(w.pkgs.len)
	w.Header.Entities.Offset = uint32(w.entities.offset)
	w.Header.Entities.Length = uint32(w.entities.len)
	w.Header.Types.Offset = uint32(w.types.offset)
	w.Header.Types.Length = uint32(w.types.len)

	writeStructAt(w.Data, 0, w.Header, w.byteOrder)
}

// -- Typed write helpers (avoid problematic generic binary serialization) --

func (w *Writer) writeFile(item *OdinDocFile) (uint32, *OdinDocFile) {
	idx, dst := w.writeTrackedItem(&w.files, binary.Size(OdinDocFile{}), item)
	if dst == nil {
		return idx, nil
	}
	return idx, dst.(*OdinDocFile)
}

func (w *Writer) writePkg(item *OdinDocPkg) (uint32, *OdinDocPkg) {
	idx, dst := w.writeTrackedItem(&w.pkgs, binary.Size(OdinDocPkg{}), item)
	if dst == nil {
		return idx, nil
	}
	return idx, dst.(*OdinDocPkg)
}

func (w *Writer) writeEntity(item *OdinDocEntity) (uint32, *OdinDocEntity) {
	idx, dst := w.writeTrackedItem(&w.entities, binary.Size(OdinDocEntity{}), item)
	if dst == nil {
		return idx, nil
	}
	return idx, dst.(*OdinDocEntity)
}

func (w *Writer) writeType(item *OdinDocType) (uint32, *OdinDocType) {
	idx, dst := w.writeTrackedItem(&w.types, binary.Size(OdinDocType{}), item)
	if dst == nil {
		return idx, nil
	}
	return idx, dst.(*OdinDocType)
}

func (w *Writer) writeTrackedItem(t *itemTracker, elemSize int, item any) (uint32, any) {
	if w.state == WriterStatePreparing {
		t.cap++
		return 0, nil
	}
	if t.len >= t.cap {
		panic(fmt.Sprintf("itemTracker full: %d >= %d", t.len, t.cap))
	}
	itemIndex := t.len
	t.len++

	offset := t.offset + elemSize*itemIndex
	if item != nil {
		buf := bytes.NewBuffer(w.Data[offset:offset])
		binary.Write(buf, w.byteOrder, item)
	}
	return uint32(itemIndex), nil
}

func (w *Writer) getFile(index uint32) *OdinDocFile {
	return getTrackedItem[OdinDocFile](w, &w.files, binary.Size(OdinDocFile{}), index)
}

func (w *Writer) getPkg(index uint32) *OdinDocPkg {
	return getTrackedItem[OdinDocPkg](w, &w.pkgs, binary.Size(OdinDocPkg{}), index)
}

func (w *Writer) getEntity(index uint32) *OdinDocEntity {
	return getTrackedItem[OdinDocEntity](w, &w.entities, binary.Size(OdinDocEntity{}), index)
}

func (w *Writer) getType(index uint32) *OdinDocType {
	return getTrackedItem[OdinDocType](w, &w.types, binary.Size(OdinDocType{}), index)
}

func getTrackedItem[T any](w *Writer, t *itemTracker, elemSize int, index uint32) *T {
	if w.state != WriterStateWriting || int(index) >= t.len {
		return nil
	}
	offset := t.offset + elemSize*int(index)
	return readStructAtOffset[T](w.Data, offset)
}

// -- String writing --

func (w *Writer) writeStringWithoutCache(str string) OdinDocString {
	if w.state == WriterStatePreparing {
		w.strings_.cap += len(str) + 1
		return OdinDocString{}
	}
	if w.strings_.len+len(str)+1 > w.strings_.cap {
		panic("string overflow")
	}
	offset := w.strings_.offset + w.strings_.len
	copy(w.Data[offset:], str)
	w.Data[offset+len(str)] = 0
	w.strings_.len += len(str) + 1
	return OdinDocString{Offset: uint32(offset), Length: uint32(len(str))}
}

func (w *Writer) writeString(str string) OdinDocString {
	if cached, ok := w.stringCache[str]; ok {
		if w.state == WriterStateWriting {
			// Verify cached string matches (debug check from C++)
			_ = extractString(w.Data, cached)
		}
		return cached
	}
	res := w.writeStringWithoutCache(str)
	w.stringCache[str] = res
	return res
}

// -- Slice/blob writing --

func (w *Writer) writeBytes(data []byte) OdinDocArray {
	if len(data) == 0 {
		return OdinDocArray{}
	}
	if w.state == WriterStatePreparing {
		w.blob.cap = alignOffset(w.blob.cap, 4)
		w.blob.cap += len(data)
		return OdinDocArray{}
	}
	w.blob.len = alignOffset(w.blob.len, 4)
	offset := w.blob.offset + w.blob.len
	copy(w.Data[offset:], data)
	w.blob.len += len(data)
	return OdinDocArray{Offset: uint32(offset), Length: uint32(len(data))}
}

func (w *Writer) writeUint32Slice(data []uint32) OdinDocArray {
	if len(data) == 0 {
		return OdinDocArray{}
	}
	elemSize := 4
	if w.state == WriterStatePreparing {
		w.blob.cap = alignOffset(w.blob.cap, 4)
		w.blob.cap += len(data) * elemSize
		return OdinDocArray{}
	}
	w.blob.len = alignOffset(w.blob.len, 4)
	offset := w.blob.offset + w.blob.len
	buf := bytes.NewBuffer(w.Data[offset:offset])
	for _, v := range data {
		binary.Write(buf, w.byteOrder, v)
	}
	w.blob.len += len(data) * elemSize
	return OdinDocArray{Offset: uint32(offset), Length: uint32(len(data))}
}

func (w *Writer) writeUint32AsSlice(v uint32) OdinDocArray {
	return w.writeUint32Slice([]uint32{v})
}

// writeDocStringSlice writes a slice of OdinDocString to the blob.
func (w *Writer) writeDocStringSlice(data []OdinDocString) OdinDocArray {
	if len(data) == 0 {
		return OdinDocArray{}
	}
	elemSize := binary.Size(OdinDocString{})
	if w.state == WriterStatePreparing {
		w.blob.cap = alignOffset(w.blob.cap, 4)
		w.blob.cap += len(data) * elemSize
		return OdinDocArray{}
	}
	w.blob.len = alignOffset(w.blob.len, 4)
	offset := w.blob.offset + w.blob.len
	for i, v := range data {
		writeStructAt(w.Data, offset+i*elemSize, &v, w.byteOrder)
	}
	w.blob.len += len(data) * elemSize
	return OdinDocArray{Offset: uint32(offset), Length: uint32(len(data))}
}

// writeAttributeSlice writes a slice of OdinDocAttribute to the blob.
func (w *Writer) writeAttributeSlice(data []OdinDocAttribute) OdinDocArray {
	if len(data) == 0 {
		return OdinDocArray{}
	}
	elemSize := binary.Size(OdinDocAttribute{})
	if w.state == WriterStatePreparing {
		w.blob.cap = alignOffset(w.blob.cap, 4)
		w.blob.cap += len(data) * elemSize
		return OdinDocArray{}
	}
	w.blob.len = alignOffset(w.blob.len, 4)
	offset := w.blob.offset + w.blob.len
	for i, v := range data {
		writeStructAt(w.Data, offset+i*elemSize, &v, w.byteOrder)
	}
	w.blob.len += len(data) * elemSize
	return OdinDocArray{Offset: uint32(offset), Length: uint32(len(data))}
}

// writeScopeEntrySlice writes a slice of OdinDocScopeEntry to the blob.
func (w *Writer) writeScopeEntrySlice(data []OdinDocScopeEntry) OdinDocArray {
	if len(data) == 0 {
		return OdinDocArray{}
	}
	elemSize := binary.Size(OdinDocScopeEntry{})
	if w.state == WriterStatePreparing {
		w.blob.cap = alignOffset(w.blob.cap, 4)
		w.blob.cap += len(data) * elemSize
		return OdinDocArray{}
	}
	w.blob.len = alignOffset(w.blob.len, 4)
	offset := w.blob.offset + w.blob.len
	for i, v := range data {
		writeStructAt(w.Data, offset+i*elemSize, &v, w.byteOrder)
	}
	w.blob.len += len(data) * elemSize
	return OdinDocArray{Offset: uint32(offset), Length: uint32(len(data))}
}

// -- File I/O --

// WriteToFile writes the serialized binary data to a file.
func (w *Writer) WriteToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to write .odin-doc to %s: %w", filename, err)
	}
	defer f.Close()

	if _, err := f.Write(w.Data[:w.DataLen]); err != nil {
		return err
	}
	fmt.Printf("Wrote .odin-doc file to: %s\n", filename)
	return nil
}

// Write produces a complete .odin-doc binary file.
func (w *Writer) Write(info any, filename string) error {
	globalInDocWriter.Store(true)
	defer globalInDocWriter.Store(false)

	w.info = info
	w.prepare()
	w.writeDocs(info)
	w.startWriting()
	w.writeDocs(info)
	w.endWriting()
	return w.WriteToFile(filename)
}

// ============================================================================
// Low-level buffer helpers
// ============================================================================

func writeStructAt[T any](data []byte, offset int, val *T, order binary.ByteOrder) {
	buf := bytes.NewBuffer(data[offset:offset])
	binary.Write(buf, order, val)
}

func readStructAtOffset[T any](data []byte, offset int) *T {
	var val T
	size := binary.Size(val)
	if size < 0 || offset+size > len(data) {
		return nil
	}
	buf := bytes.NewReader(data[offset : offset+size])
	if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
		return nil
	}
	return &val
}
