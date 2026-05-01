package common

import "unsafe"

// nextPow2 returns the smallest power of 2 >= n.
func nextPow2_i32(n i32) i32 {
	if n <= 0 {
		return 0
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

func nextPow2_i64(n i64) i64 {
	if n <= 0 {
		return 0
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

func nextPow2_isize(n isize) isize {
	if n <= 0 {
		return 0
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if unsafe.Sizeof(n) > 4 {
		n |= n >> 32
	}
	return n + 1
}

func nextPow2_u32(n u32) u32 {
	if n <= 0 {
		return 0
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

func nextPow2_u64(n u64) u64 {
	if n <= 0 {
		return 0
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// nextPow2Int is the common int version used internally.
func nextPow2Int(n int) int {
	return int(nextPow2_isize(isize(n)))
}

// IsPowerOfTwo checks if x is a positive power of two.
func IsPowerOfTwo(x i64) bool {
	return x > 0 && (x&(x-1)) == 0
}

func IsPowerOfTwo_u64(x u64) bool {
	return x > 0 && (x&(x-1)) == 0
}

// BitSetCount counts set bits (popcount).
func BitSetCount_u32(x u32) i32 {
	x = x - ((x >> 1) & 0x55555555)
	x = (x & 0x33333333) + ((x >> 2) & 0x33333333)
	x = (x + (x >> 4)) & 0x0F0F0F0F
	x = x + (x >> 8)
	x = x + (x >> 16)
	return i32(x & 0x3F)
}

func BitSetCount_u64(x u64) i64 {
	lo := u32(x)
	hi := u32(x >> 32)
	return i64(BitSetCount_u32(lo)) + i64(BitSetCount_u32(hi))
}

// FloorLog2 returns floor(log2(x)). x must be > 0.
func FloorLog2_u32(x u32) u32 {
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return u32(BitSetCount_u32(x)) - 1
}

func FloorLog2_u64(x u64) u64 {
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	return u64(BitSetCount_u64(x)) - 1
}

// CeilLog2 returns ceil(log2(x)).
func CeilLog2_u32(x u32) u32 {
	y := i32(x & (x - 1))
	y |= -y
	y >>= 31
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return u32(BitSetCount_u32(x) - 1 - i32(y))
}

func CeilLog2_u64(x u64) u64 {
	y := i64(x & (x - 1))
	y |= -y
	y >>= 63
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	return u64(BitSetCount_u64(x)) - u64(1) - u64(y)
}

// PrevPow2 returns the largest power of two <= n.
func PrevPow2_u32(n u32) u32 {
	if n == 0 {
		return 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n - (n >> 1)
}

func PrevPow2_i32(n i32) i32 {
	if n <= 0 {
		return 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n - (n >> 1)
}

func PrevPow2_i64(n i64) i64 {
	if n <= 0 {
		return 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n - (n >> 1)
}

// Comparison functions.
func IsizeCmp(x, y isize) i32 {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

func U64Cmp(x, y u64) i32 {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

func I64Cmp(x, y i64) i32 {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

func I32Cmp(x, y i32) i32 {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

// AlignFormula aligns size to the given alignment.
func AlignFormula_i64(size, align i64) i64 {
	result := size + align - 1
	return result - (result % align)
}

func AlignFormula_isize(size, align isize) isize {
	result := size + align - 1
	return result - result%align
}

func AlignFormula_ptr(ptr uintptr, align isize) uintptr {
	result := ptr + uintptr(align) - 1
	return result - result%uintptr(align)
}

// Overflow checks.
func AddOverflow_u64(x, y u64) (u64, bool) {
	result := x + y
	return result, result < x || result < y
}

func SubOverflow_u64(x, y u64) (u64, bool) {
	result := x - y
	return result, result > x
}

func MulOverflow_u64(x, y u64) (lo, hi u64) {
	// Returns lo and hi for 128-bit multiplication.
	lo = x * y
	a := u32(x)
	b := u32(x >> 32)
	c := u32(y)
	d := u32(y >> 32)
	hi = u64(a)*u64(c)>>32 + u64(b)*u64(c)>>32 + u64(a)*u64(d)>>32 + u64(b)*u64(d)>>32
	ac := u64(a) * u64(c)
	bc := u64(b) * u64(c)
	ad := u64(a) * u64(d)
	bd := u64(b) * u64(d)
	carry := ((ac>>32)+u64(u32(bc))+u64(u32(ad))) >> 32
	hi = (bc>>32)+(ad>>32)+bd + carry
	carry2 := ((ac>>32)+u64(a)*u64(c)>>32) >> 32
	_ = carry2
	return
}

// FNV hashing.
func FNV32a(data []byte) u32 {
	h := u32(0x811c9dc5)
	for _, b := range data {
		h = (h ^ u32(b)) * 0x01000193
	}
	return h
}

func FNV64a(data []byte, seed u64) u64 {
	h := seed
	for _, b := range data {
		h = (h ^ u64(b)) * 0x100000001b3
	}
	return h
}

// U64DigitValue returns the numeric value of a rune, or 16 if invalid.
func U64DigitValue(r Rune) u64 {
	switch {
	case r >= '0' && r <= '9':
		return u64(r - '0')
	case r >= 'a' && r <= 'z':
		return u64(r - 'a' + 10)
	case r >= 'A' && r <= 'Z':
		return u64(r - 'A' + 10)
	}
	return 16
}

// GlobalNumToCharTable maps digit values to characters.
const GlobalNumToCharTable = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz@$"

// U64ToString converts a uint64 to a String using base 10.
func U64ToString(v u64) String {
	var buf [32]byte
	i := 32
	b := u64(10)
	for v >= b {
		i--
		buf[i] = GlobalNumToCharTable[v%b]
		v /= b
	}
	i--
	buf[i] = GlobalNumToCharTable[v%b]
	t := buf[i:]
	return MakeString(t[:], len(t))
}

// I64ToString converts an int64 to a String using base 10.
func I64ToString(a i64) String {
	var buf [32]byte
	i := 32
	negative := false
	if a < 0 {
		negative = true
		a = -a
	}
	v := u64(a)
	b := u64(10)
	for v >= b {
		i--
		buf[i] = GlobalNumToCharTable[v%b]
		v /= b
	}
	i--
	buf[i] = GlobalNumToCharTable[v%b]
	if negative {
		i--
		buf[i] = '-'
	}
	t := buf[i:]
	return MakeString(t[:], len(t))
}

// U64FromString parses a u64 from a String, supporting 0b/0o/0d/0z/0x prefixes.
func U64FromString(str String) u64 {
	base := u64(10)
	hasPrefix := false
	if str.Len > 2 && str.Text[0] == '0' {
		switch str.Text[1] {
		case 'b':
			base = 2
			hasPrefix = true
		case 'o':
			base = 8
			hasPrefix = true
		case 'd':
			base = 10
			hasPrefix = true
		case 'z':
			base = 12
			hasPrefix = true
		case 'x', 'h':
			base = 16
			hasPrefix = true
		}
	}
	text := str.Text
	length := str.Len
	if hasPrefix {
		text = text[2:]
		length -= 2
	}
	result := u64(0)
	for i := 0; i < length; i++ {
		if text[i] == '_' {
			continue
		}
		v := U64DigitValue(Rune(text[i]))
		if v >= base {
			break
		}
		result = result*base + v
	}
	return result
}

// GlobalModulePath is the global module path (set at startup).
var GlobalModulePath String
var GlobalModulePathSet bool
