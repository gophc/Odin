package common

import "unsafe"

// Arena is a linear/region-based memory allocator.
// It allocates from fixed-size blocks, growing as needed.
type Arena struct {
	currBlock        *MemoryBlock
	minimumBlockSize isize
	tempCount        isize
}

// MemoryBlock is a single allocation block within an arena.
type MemoryBlock struct {
	prev      *MemoryBlock
	base      []byte
	size      isize
	used      isize
	committed isize
}

const (
	DefaultMinimumBlockSize = 8 << 20 // 8 MB
	DefaultPageSize         = 4096
)

func init() {
	ArenaInit(&defaultPermanentArena)
	ArenaInit(&defaultTemporaryArena)
}

// ArenaInit initializes an arena.
func ArenaInit(a *Arena) {
	a.minimumBlockSize = DefaultMinimumBlockSize
	a.currBlock = nil
	a.tempCount = 0
}

// arenaAlignForwardOffset returns the padding needed to align the current block position.
func ArenaAlignForwardOffset(a *Arena, alignment isize) isize {
	if a.currBlock == nil {
		return 0
	}
	return AlignFormula_isize(a.currBlock.used, alignment) - a.currBlock.used
}

// ArenaAlloc allocates memory from the arena with the given size and alignment.
func ArenaAlloc(a *Arena, size, alignment isize) []byte {
	if size <= 0 {
		return nil
	}
	if alignment <= 0 {
		alignment = 8
	}

	// Try to fit in the current block.
	if a.currBlock != nil {
		offset := ArenaAlignForwardOffset(a, alignment)
		if a.currBlock.used+offset+size <= a.currBlock.size {
			start := a.currBlock.used + offset
			a.currBlock.used = start + size
			return a.currBlock.base[start : start+size]
		}
	}

	// Need new block.
	blockSize := a.minimumBlockSize
	needed := size + alignment
	if needed > blockSize {
		blockSize = nextPow2_isize(needed)
	}
	block := &MemoryBlock{
		prev:      a.currBlock,
		base:      make([]byte, blockSize),
		size:      blockSize,
		used:      0,
		committed: blockSize,
	}
	a.currBlock = block

	offset := AlignFormula_isize(0, alignment)
	block.used = offset + size
	return block.base[offset : offset+size]
}

// ArenaFreeAll frees all blocks in the arena.
func ArenaFreeAll(a *Arena) {
	for block := a.currBlock; block != nil; {
		prev := block.prev
		block.prev = nil
		block.base = nil
		a.currBlock = prev
		block = a.currBlock
	}
	a.tempCount = 0
}

// ArenaAllocItem allocates a single item of type T from the arena.
func ArenaAllocItem[T any](a *Arena) *T {
	var zero T
	size := isize(unsafe.Sizeof(zero))
	if size == 0 {
		size = 1
	}
	align := isize(unsafe.Alignof(zero))
	if align == 0 {
		align = 8
	}
	ptr := ArenaAlloc(a, size, align)
	return (*T)(unsafe.Pointer(unsafe.SliceData(ptr)))
}

// ArenaAllocArray allocates count items of type T from the arena.
func ArenaAllocArray[T any](a *Arena, count int) []T {
	if count <= 0 {
		return nil
	}
	var zero T
	elemSize := isize(unsafe.Sizeof(zero))
	if elemSize == 0 {
		elemSize = 1
	}
	align := isize(unsafe.Alignof(zero))
	if align == 0 {
		align = 8
	}
	totalSize := elemSize * isize(count)
	ptr := ArenaAlloc(a, totalSize, align)
	return unsafe.Slice((*T)(unsafe.Pointer(unsafe.SliceData(ptr))), count)
}

// --- ArenaTemp enables scoped temporary allocations ---

type ArenaTemp struct {
	arena *Arena
	block *MemoryBlock
	used  isize
}

// ArenaTempBegin saves the current arena state.
func ArenaTempBegin(a *Arena) ArenaTemp {
	at := ArenaTemp{arena: a}
	if a.currBlock != nil {
		at.block = a.currBlock
		at.used = a.currBlock.used
	}
	a.tempCount++
	return at
}

// ArenaTempEnd restores the arena to the saved state.
func ArenaTempEnd(at ArenaTemp) {
	a := at.arena
	if at.block != nil {
		// Free blocks allocated after the saved block.
		for a.currBlock != at.block {
			prev := a.currBlock.prev
			a.currBlock.prev = nil
			a.currBlock.base = nil
			a.currBlock = prev
		}
		a.currBlock.used = at.used
	}
	a.tempCount--
}

// ArenaTempIgnore consumes the temp state without restoring.
func ArenaTempIgnore(at ArenaTemp) {
	at.arena.tempCount--
}

// --- StaticArena is a pre-reserved arena that commits pages on demand. ---

type StaticArena struct {
	data            []byte
	used            isize
	committed       isize
	reserved        isize
	commitBlockSize isize
}

const StaticArenaDefaultCommitBlockSize = 8 << 20

// StaticArenaInit initializes a static arena with reserved and commit sizes.
func StaticArenaInit(a *StaticArena, reserveSize, commitBlockSize isize) {
	if commitBlockSize <= 0 {
		commitBlockSize = StaticArenaDefaultCommitBlockSize
	}
	*a = StaticArena{
		data:            make([]byte, reserveSize),
		reserved:        reserveSize,
		commitBlockSize: commitBlockSize,
	}
}

// StaticArenaCommitMemory is a no-op in Go (make already commits).
func StaticArenaCommitMemory(a *StaticArena, amount isize) {}

// StaticArenaAlloc allocates from a static arena.
func StaticArenaAlloc(a *StaticArena, size, alignment isize) []byte {
	if alignment <= 0 {
		alignment = 1
	}
	curr := uintptr(a.used)
	alignMask := uintptr(alignment - 1)
	alignOffset := (alignMask - (curr+alignMask)%alignMask) % alignMask
	totalSize := size + isize(alignOffset)
	end := a.used + totalSize

	if end > a.committed {
		needed := end - a.committed
		blocks := (needed + a.commitBlockSize - 1) / a.commitBlockSize
		a.committed += blocks * a.commitBlockSize
		if a.committed > a.reserved {
			a.committed = a.reserved
		}
	}
	if end > a.committed {
		panic("out of memory for static arena")
	}
	a.used = end
	ptr := a.data[a.used-totalSize : a.used]
	return ptr[alignOffset:]
}

// StaticArenaReset resets the static arena used marker.
func StaticArenaReset(a *StaticArena) {
	a.used = 0
}

// --- ThreadArenaKind ---

type ThreadArenaKind int

const (
	ThreadArena_Permanent ThreadArenaKind = iota
	ThreadArena_Temporary
)

var defaultPermanentArena Arena
var defaultTemporaryArena Arena

// GetArena returns the thread-local arena for the given kind.
func GetArena(kind ThreadArenaKind) *Arena {
	switch kind {
	case ThreadArena_Temporary:
		return &defaultTemporaryArena
	default:
		return &defaultPermanentArena
	}
}

// PermanentAllocItem allocates a single item from the permanent arena.
func PermanentAllocItem[T any]() *T {
	return ArenaAllocItem[T](&defaultPermanentArena)
}

// PermanentAllocArray allocates an array from the permanent arena.
func PermanentAllocArray[T any](count int) []T {
	return ArenaAllocArray[T](&defaultPermanentArena, count)
}

// PermanentSliceMake creates a Slice from the permanent arena.
func PermanentSliceMake[T any](count int) Slice[T] {
	return Slice[T]{
		data:  PermanentAllocArray[T](count),
		count: count,
	}
}

// TemporaryAllocItem allocates a single item from the temporary arena.
func TemporaryAllocItem[T any]() *T {
	return ArenaAllocItem[T](&defaultTemporaryArena)
}

// TemporaryAllocArray allocates an array from the temporary arena.
func TemporaryAllocArray[T any](count int) []T {
	return ArenaAllocArray[T](&defaultTemporaryArena, count)
}

// TemporarySliceMake creates a Slice from the temporary arena.
func TemporarySliceMake[T any](count int) Slice[T] {
	return Slice[T]{
		data:  TemporaryAllocArray[T](count),
		count: count,
	}
}

// --- ResizeArrayRaw dynamically resizes a raw slice (pointer to slice). ---

func ResizeArrayRaw[T any](array *[]T, oldCount, newCount, elementSize int) int {
	if newCount == 0 {
		*array = nil
		return 0
	}
	if newCount <= oldCount {
		return newCount
	}
	newSlice := make([]T, newCount)
	if *array != nil && oldCount > 0 {
		copy(newSlice, (*array)[:oldCount])
	}
	*array = newSlice
	return newCount
}
