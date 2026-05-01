package common

import "unsafe"

// Integer type aliases matching Odin's type system.
type (
	i8    = int8
	i16   = int16
	i32   = int32
	i64   = int64
	isize = int
	u8    = uint8
	u16   = uint16
	u32   = uint32
	u64   = uint64
	usize = uint
)

// Integer boundary constants.
const (
	I8_MIN  int8  = -128
	I8_MAX  int8  = 127
	I16_MIN int16 = -32768
	I16_MAX int16 = 32767
	I32_MIN int32 = -2147483648
	I32_MAX int32 = 2147483647
	I64_MIN int64 = -9223372036854775808
	I64_MAX int64 = 9223372036854775807
	U8_MAX  uint8  = 255
	U16_MAX uint16 = 65535
	U32_MAX uint32 = 4294967295
	U64_MAX uint64 = 18446744073709551615
)

var SignedIntegerMins = [9]i64{0, int64(I8_MIN), int64(I16_MIN), 0, int64(I32_MIN), 0, 0, 0, I64_MIN}
var SignedIntegerMaxs = [9]i64{0, int64(I8_MAX), int64(I16_MAX), 0, int64(I32_MAX), 0, 0, 0, I64_MAX}
var UnsignedIntegerMaxs = [9]u64{0, u64(U8_MAX), u64(U16_MAX), 0, u64(U32_MAX), 0, 0, 0, U64_MAX}

// Float type aliases matching Odin's type system.
type (
	f32 = float32
	f64 = float64
)

// Rune represents a Unicode code point (alias for i32 as in Odin).
type Rune = i32

// Allocator is a general-purpose memory allocator handle.
type Allocator struct {
	proc func(mode AllocatorMode, size, alignment isize, oldMemory unsafe.Pointer, oldSize isize, flags u64) unsafe.Pointer
	data unsafe.Pointer
}

// AllocatorMode describes the allocation operation.
type AllocatorMode i32

const (
	AllocatorMode_Alloc   AllocatorMode = 0
	AllocatorMode_Free    AllocatorMode = 1
	AllocatorMode_FreeAll AllocatorMode = 2
	AllocatorMode_Resize  AllocatorMode = 3
	AllocatorMode_Query   AllocatorMode = 4
)

// MapIndex is the index type used in hash table structures.
type MapIndex u32

// MAP_SENTINEL marks an empty hash slot.
const MAP_SENTINEL MapIndex = ^MapIndex(0)

// MapFindResult groups the result of a hash table lookup.
type MapFindResult struct {
	HashIndex  MapIndex
	EntryPrev  MapIndex
	EntryIndex MapIndex
}

// Futex is a fast userspace mutex — an atomic i32.
type Futex struct {
	val i32
}
