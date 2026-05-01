package common

// String is a UTF-8 string with text and length.
type String struct {
	Text []byte
	Len  int
}

// String16 is a UTF-16 string with text and length.
type String16 struct {
	Text []u16
	Len  int
}

// StringIterator tracks position during string splitting.
type StringIterator struct {
	Str String
	Pos int
}

// --- Construction ---

// MakeString creates a String from bytes and length.
func MakeString(text []byte, length int) String {
	if length < 0 {
		length = len(text)
	}
	return String{Text: text[:length], Len: length}
}

// MakeStringC creates a String from a null-terminated byte slice.
func MakeStringC(text []byte) String {
	n := 0
	for n < len(text) && text[n] != 0 {
		n++
	}
	return String{Text: text[:n], Len: n}
}

// MakeString16 creates a String16 from u16 slice and length.
func MakeString16(text []u16, length int) String16 {
	if length < 0 {
		length = len(text)
	}
	return String16{Text: text[:length], Len: length}
}

// String16Len returns the length of a null-terminated u16 slice.
func String16Len(s []u16) int {
	for i, c := range s {
		if c == 0 {
			return i
		}
	}
	return len(s)
}

// MakeString16C creates a String16 from a null-terminated u16 slice.
func MakeString16C(text []u16) String16 {
	n := String16Len(text)
	return String16{Text: text[:n], Len: n}
}

// --- Slice / Substring ---

func (s String) Slice(lo, hi int) String {
	if lo < 0 {
		lo = 0
	}
	if hi > s.Len {
		hi = s.Len
	}
	if lo > hi {
		lo = hi
	}
	if hi == lo {
		return String{}
	}
	return String{Text: s.Text[lo:hi], Len: hi - lo}
}

func Substring(s String, lo, hi int) String   { return s.Slice(lo, hi) }
func Substring16(s String16, lo, hi int) String16 {
	if lo < 0 {
		lo = 0
	}
	if hi > s.Len {
		hi = s.Len
	}
	if lo > hi {
		lo = hi
	}
	if hi == lo {
		return String16{}
	}
	return String16{Text: s.Text[lo:hi], Len: hi - lo}
}

// --- Comparison ---

func StringEq(a, b String) bool {
	if a.Len != b.Len {
		return false
	}
	for i := 0; i < a.Len; i++ {
		if a.Text[i] != b.Text[i] {
			return false
		}
	}
	return true
}

func StringNe(a, b String) bool { return !StringEq(a, b) }

func StringCompare(a, b String) int {
	n := a.Len
	if n > b.Len {
		n = b.Len
	}
	for i := 0; i < n; i++ {
		if a.Text[i] != b.Text[i] {
			return int(a.Text[i]) - int(b.Text[i])
		}
	}
	return a.Len - b.Len
}

func StringLt(a, b String) bool { return StringCompare(a, b) < 0 }
func StringGt(a, b String) bool { return StringCompare(a, b) > 0 }
func StringLe(a, b String) bool { return StringCompare(a, b) <= 0 }
func StringGe(a, b String) bool { return StringCompare(a, b) >= 0 }

func String16Eq(a, b String16) bool {
	if a.Len != b.Len {
		return false
	}
	for i := 0; i < a.Len; i++ {
		if a.Text[i] != b.Text[i] {
			return false
		}
	}
	return true
}

func String16Compare(a, b String16) int {
	n := a.Len
	if n > b.Len {
		n = b.Len
	}
	for i := 0; i < n; i++ {
		if a.Text[i] != b.Text[i] {
			return int(a.Text[i]) - int(b.Text[i])
		}
	}
	return a.Len - b.Len
}

func StrEqIgnoreCase(a, b String) bool {
	if a.Len != b.Len {
		return false
	}
	for i := 0; i < a.Len; i++ {
		ac, bc := a.Text[i], b.Text[i]
		if ac >= 'A' && ac <= 'Z' {
			ac += 32
		}
		if bc >= 'A' && bc <= 'Z' {
			bc += 32
		}
		if ac != bc {
			return false
		}
	}
	return true
}

// --- Search ---

func StringIndexByte(s String, x byte) int {
	for i := 0; i < s.Len; i++ {
		if s.Text[i] == x {
			return i
		}
	}
	return -1
}

const PRIME_RABIN_KARP u32 = 16777619

func hashRabinKarp(s String) u32 {
	h := u32(0)
	for i := 0; i < s.Len; i++ {
		h = h*PRIME_RABIN_KARP + u32(s.Text[i])
	}
	return h
}

func hashRabinKarpPow(sLen int) u32 {
	pow := u32(1)
	pq := PRIME_RABIN_KARP
	for i := sLen; i > 0; i >>= 1 {
		if (i & 1) != 0 {
			pow *= pq
		}
		pq *= pq
	}
	return pow
}

func HashStrRabinKarp(s String) (u32, u32) {
	return hashRabinKarp(s), hashRabinKarpPow(s.Len)
}

func StringIndex(s, substr String) int {
	if substr.Len == 0 {
		return 0
	}
	if substr.Len == 1 {
		return StringIndexByte(s, substr.Text[0])
	}
	if substr.Len > s.Len {
		return -1
	}
	firstHash := hashRabinKarp(substr)
	pow := hashRabinKarpPow(substr.Len)
	h := hashRabinKarp(s.Slice(0, substr.Len))
	if h == firstHash && StringEq(s.Slice(0, substr.Len), substr) {
		return 0
	}
	for i := substr.Len; i < s.Len; i++ {
		h = h*PRIME_RABIN_KARP + u32(s.Text[i])
		h -= pow * u32(s.Text[i-substr.Len])
		if h == firstHash && StringEq(s.Slice(i-substr.Len+1, i+1), substr) {
			return i - substr.Len + 1
		}
	}
	return -1
}

func StringContainsChar(s String, c byte) bool { return StringIndexByte(s, c) >= 0 }
func StringContainsString(haystack, needle String) bool { return StringIndex(haystack, needle) >= 0 }

// --- Prefix / Suffix ---

func StringStartsWith(s, prefix String) bool {
	if prefix.Len > s.Len {
		return false
	}
	return StringEq(s.Slice(0, prefix.Len), prefix)
}

func StringEndsWith(s, suffix String) bool {
	if suffix.Len > s.Len {
		return false
	}
	return StringEq(s.Slice(s.Len-suffix.Len, s.Len), suffix)
}

func StringTrimStartsWith(s, prefix String) String {
	if StringStartsWith(s, prefix) {
		return s.Slice(prefix.Len, s.Len)
	}
	return s
}

// --- Partition ---

type StringPartition struct {
	Head  String
	Match String
	Tail  String
}

func StringPartitionFunc(str, sep String) StringPartition {
	i := StringIndex(str, sep)
	if i < 0 {
		return StringPartition{Head: str}
	}
	return StringPartition{
		Head:  str.Slice(0, i),
		Match: str.Slice(i, i+sep.Len),
		Tail:  str.Slice(i+sep.Len, str.Len),
	}
}

// --- Split ---

func StringSplitIterator(it *StringIterator, sep byte) String {
	start := it.Pos
	if start >= it.Str.Len {
		return String{}
	}
	for i := start; i < it.Str.Len; i++ {
		if it.Str.Text[i] == sep {
			res := it.Str.Slice(start, i)
			it.Pos = i + 1
			return res
		}
	}
	it.Pos = it.Str.Len
	return it.Str.Slice(start, it.Str.Len)
}

// --- Cloning and Concatenation ---

func CloneString(a Allocator, x String) String {
	if x.Len == 0 {
		return String{}
	}
	data := make([]byte, x.Len+1)
	copy(data, x.Text[:x.Len])
	data[x.Len] = 0
	_ = a
	return String{Text: data[:x.Len], Len: x.Len}
}

func ConcatenateStrings(a Allocator, x, y String) String {
	length := x.Len + y.Len
	data := make([]byte, length+1)
	copy(data, x.Text[:x.Len])
	copy(data[x.Len:], y.Text[:y.Len])
	data[length] = 0
	_ = a
	return String{Text: data[:length], Len: length}
}

func Concatenate3Strings(a Allocator, x, y, z String) String {
	length := x.Len + y.Len + z.Len
	data := make([]byte, length+1)
	copy(data, x.Text[:x.Len])
	copy(data[x.Len:], y.Text[:y.Len])
	copy(data[x.Len+y.Len:], z.Text[:z.Len])
	data[length] = 0
	_ = a
	return String{Text: data[:length], Len: length}
}

func Concatenate4Strings(a Allocator, x, y, z, w String) String {
	length := x.Len + y.Len + z.Len + w.Len
	data := make([]byte, length+1)
	copy(data, x.Text[:x.Len])
	copy(data[x.Len:], y.Text[:y.Len])
	copy(data[x.Len+y.Len:], z.Text[:z.Len])
	copy(data[x.Len+y.Len+z.Len:], w.Text[:w.Len])
	data[length] = 0
	_ = a
	return String{Text: data[:length], Len: length}
}

func CopyString(a Allocator, s String) String {
	if s.Len == 0 {
		return String{}
	}
	data := make([]byte, s.Len+1)
	copy(data, s.Text[:s.Len])
	data[s.Len] = 0
	_ = a
	return String{Text: data[:s.Len], Len: s.Len}
}

// --- Mutation ---

func StringToLower(s *String) {
	for i := 0; i < s.Len; i++ {
		if c := s.Text[i]; 'A' <= c && c <= 'Z' {
			s.Text[i] = c + 32
		}
	}
}

// --- White space trimming ---

func StringTrimWhitespace(s String) String {
	start := 0
	for start < s.Len && IsWhiteSpace(s.Text[start]) {
		start++
	}
	end := s.Len
	for end > start && IsWhiteSpace(s.Text[end-1]) {
		end--
	}
	return s.Slice(start, end)
}

func StringTrimTrailingWhitespace(s String) String {
	end := s.Len
	for end > 0 && (IsWhiteSpace(s.Text[end-1]) || s.Text[end-1] == 0) {
		end--
	}
	return s.Slice(0, end)
}

// --- Path helpers ---

func IsSeparator(c byte) bool { return c == '/' || c == '\\' }

func StringExtensionPosition(str String) int {
	for i := str.Len - 1; i >= 0; i-- {
		if IsSeparator(str.Text[i]) {
			break
		}
		if str.Text[i] == '.' {
			return i
		}
	}
	return -1
}

func PathExtension(str String, includeDot bool) String {
	pos := StringExtensionPosition(str)
	if pos < 0 {
		return String{}
	}
	if includeDot {
		return str.Slice(pos, str.Len)
	}
	return str.Slice(pos+1, str.Len)
}

func PathRemoveExtension(str String) String {
	pos := StringExtensionPosition(str)
	if pos < 0 {
		return str
	}
	return str.Slice(0, pos)
}

func FilenameFromPath(s String) String {
	pos := StringExtensionPosition(s)
	if pos >= 0 {
		return s.Slice(0, pos)
	}
	return String{}
}

func FilenameWithoutDirectory(s String) String {
	i := s.Len - 1
	for i >= 0 && !IsSeparator(s.Text[i]) {
		i--
	}
	return s.Slice(i+1, s.Len)
}

func LastPathElement(path String) String {
	count := 0
	for i := path.Len - 1; i >= 0 && path.Text[i] != '/'; i-- {
		count++
	}
	if count > 0 {
		return path.Slice(path.Len-count, path.Len)
	}
	return String{}
}

func DirectoryFromPath(s String) String {
	i := s.Len - 1
	for i >= 0 && !IsSeparator(s.Text[i]) {
		i--
	}
	if i >= 0 {
		return s.Slice(0, i)
	}
	return String{}
}

func RemoveExtensionFromPath(s String) String {
	if s.Len > 0 && s.Text[s.Len-1] == '.' {
		return s
	}
	for i := s.Len - 1; i >= 0; i-- {
		if s.Text[i] == '.' {
			return s.Slice(0, i)
		}
	}
	return s
}

func RemoveDirectoryFromPath(s String) String {
	length := 0
	for i := s.Len - 1; i >= 0 && !IsSeparator(s.Text[i]); i-- {
		length++
	}
	return s.Slice(s.Len-length, s.Len)
}

// --- Lines ---

func SplitLinesFirstLineFromArray(array Array[byte], allocator Allocator) String {
	for i := 0; i < array.count; i++ {
		c := array.data[i]
		if c == '\n' || c == '\r' {
			return MakeString(array.data[:i], i)
		}
	}
	return MakeString(array.data[:array.count], array.count)
}

func SplitLinesFromArray(array Array[byte], allocator Allocator) []String {
	var lines []String
	start := 0
	for i := 0; i < array.count; i++ {
		c := array.data[i]
		if c == '\n' || c == '\r' {
			if i > start {
				lines = append(lines, MakeString(array.data[start:i], i-start))
			}
			start = i + 1
			if c == '\r' && i+1 < array.count && array.data[i+1] == '\n' {
				i++
				start = i + 1
			}
		}
	}
	if start < array.count {
		lines = append(lines, MakeString(array.data[start:array.count], array.count-start))
	}
	return lines
}

// --- C string allocation ---

func AllocCString(a Allocator, s String) []byte {
	if s.Len == 0 {
		b := make([]byte, 1)
		b[0] = 0
		return b
	}
	b := make([]byte, s.Len+1)
	copy(b, s.Text[:s.Len])
	b[s.Len] = 0
	return b
}

// --- Conversion ---

func StringToString16(a Allocator, s String) String16 {
	if s.Len == 0 {
		return String16{}
	}
	buf := make([]u16, s.Len)
	count := 0
	pos := 0
	for pos < s.Len {
		r, sz := Utf8Decode(s.Text[pos:])
		if sz <= 0 {
			break
		}
		pos += sz
		if r <= 0xFFFF {
			buf[count] = u16(r)
			count++
		} else {
			r -= 0x10000
			if count+2 <= s.Len {
				buf[count] = u16(r>>10) + 0xD800
				buf[count+1] = u16(r&0x3FF) + 0xDC00
				count += 2
			}
		}
	}
	_ = a
	return String16{Text: buf[:count], Len: count}
}

func String16ToString(a Allocator, s String16) String {
	if s.Len == 0 {
		return String{}
	}
	buf := make([]byte, s.Len*4)
	count := 0
	for i := 0; i < s.Len; i++ {
		r := Rune(s.Text[i])
		if r >= 0xD800 && r <= 0xDBFF && i+1 < s.Len {
			next := Rune(s.Text[i+1])
			if next >= 0xDC00 && next <= 0xDFFF {
				r = DecodeSurrogatePair(u16(r), u16(next))
				i++
			}
		}
		n := Utf8EncodeRune(buf[count:], r)
		if n > 0 {
			count += n
		}
	}
	_ = a
	return String{Text: buf[:count], Len: count}
}

// --- Normalize ---

func NormalizePath(a Allocator, path, sep String) String {
	if path.Len == 0 {
		return path
	}
	var parts []String
	it := StringIterator{Str: path, Pos: 0}
	for part := StringSplitIterator(&it, sep.Text[0]); part.Len > 0; part = StringSplitIterator(&it, sep.Text[0]) {
		if part.Len == 0 || (part.Len == 1 && part.Text[0] == '.') {
			continue
		}
		if part.Len == 2 && part.Text[0] == '.' && part.Text[1] == '.' {
			if len(parts) > 0 {
				parts = parts[:len(parts)-1]
			}
			continue
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return String{}
	}
	// Join
	totalLen := 0
	for _, p := range parts {
		totalLen += p.Len
	}
	totalLen += (len(parts) - 1) * sep.Len
	data := make([]byte, totalLen)
	pos := 0
	for i, p := range parts {
		if i > 0 {
			copy(data[pos:], sep.Text[:sep.Len])
			pos += sep.Len
		}
		copy(data[pos:], p.Text[:p.Len])
		pos += p.Len
	}
	return String{Text: data, Len: totalLen}
}

// --- Escape ---

func EscapeChar(a Allocator, s String, c byte) String {
	var buf []byte
	for i := 0; i < s.Len; i++ {
		ch := s.Text[i]
		if ch == c {
			buf = append(buf, c)
		}
		buf = append(buf, ch)
	}
	_ = a
	return String{Text: buf, Len: len(buf)}
}

func StringJoinAndQuote(a Allocator, strings []String) String {
	var result String
	for i, s := range strings {
		if i > 0 {
			result = ConcatenateStrings(a, result, MakeString([]byte{','}, 1))
		}
		q := Concatenate3Strings(a, MakeString([]byte{34}, 1), s, MakeString([]byte{34}, 1))
		result = ConcatenateStrings(a, result, q)
	}
	return result
}

// --- Valid identifier check ---

func StringIsValidIdentifier(str String) bool {
	if str.Len == 0 {
		return false
	}
	first := str.Text[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	for i := 1; i < str.Len; i++ {
		c := str.Text[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}
