package util

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

func MustAtoI(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

//region init

var NumChars [256]int8

func init() {
	for i := 0; i < len(NumChars); i++ {
		NumChars[i] = int8(-1)
	}
	charsStr := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i, c := range charsStr {
		k := uint8(c)
		NumChars[k] = int8(i)
	}
	charsStr = strings.ToLower(charsStr)
	for i, c := range charsStr {
		k := uint8(c)
		NumChars[k] = int8(i)
	}
}

//goland:noinspection GoUnusedGlobalVariable
var (
	errEmptyInt                    = errors.New("empty integer")
	errUnexpectedFirstChar         = errors.New("unexpected first char found. Expecting 0-9")
	errUnexpectedRadix             = errors.New("unexpected radix found. Expecting 0-36")
	errUnexpectedTrailingChar      = errors.New("unexpected trailing char found. Expecting 0-9")
	errUnexpectedFloatTrailingChar = errors.New("unexpected float trailing char found. Expecting 0-9.e")
	errTooLongInt                  = errors.New("too long int over 2**64 - 1")
)

//goland:noinspection GoUnusedGlobalVariable
var (
	errEmptyFloat           = errors.New("empty float number")
	errDuplicateFloatPoint  = errors.New("duplicate point found in float number")
	errUnexpectedFloatEnd   = errors.New("unexpected end of float number")
	errInvalidFloatExponent = errors.New("invalid float number exponent")
	errUnexpectedFloatChar  = errors.New("unexpected char found in float number")
)

//endregion

//region helper

//goland:noinspection GoUnusedConst,GoSnakeCaseUsage
const (
	MAX_LENGTH_OF_LONG = 20

	LONG_MIN_DIGITS = "9223372036854775808"

	Int64Max uint64 = 9223372036854775808
)

//go:nosplit
func PtrAdd(p uintptr, i int) unsafe.Pointer {
	//goland:noinspection GoVetUnsafePointer
	return unsafe.Pointer(uintptr(int(p) + i))
}

//goland:noinspection GoUnusedExportedFunction
func PtrStrcmp(p uintptr, s string, i int) int {
	o := StrStruct{
		Ptr: p,
		Len: i,
	}
	s1 := *(*string)(unsafe.Pointer(&o))
	return strings.Compare(s1, s)
}

//goland:noinspection GoUnusedExportedFunction
func NextLine(ptr uintptr, end uintptr) string {
	cursor := ptr
	//goland:noinspection GoVetUnsafePointer
	c := *(*byte)(unsafe.Pointer(cursor))
	for cursor < end {
		if c == '\n' || (c == '\r' && *(*byte)(PtrAdd(cursor, 1)) != '\n') {
			l := int(cursor - ptr)
			if l < 0 {
				l = 0
			}
			return PtrToStr(ptr, l)
		}
		cursor += 1
		//goland:noinspection GoVetUnsafePointer
		c = *(*byte)(unsafe.Pointer(cursor))
	}
	return PtrToStr(ptr, int(end-ptr))
}

//goland:noinspection GoUnusedExportedFunction
func PtrNextMemchr(p uintptr, c byte, end uintptr) uintptr {
	i := int(end - p)
	o := StrStruct{
		Ptr: p,
		Len: i,
	}
	idx := strings.IndexByte(*(*string)(unsafe.Pointer(&o)), c)
	if idx < 0 {
		return end
	}
	return p + uintptr(idx+1)
}

//goland:noinspection GoUnusedExportedFunction
func ByteToStr(c ...byte) string {
	return string(c[:])
}

//goland:noinspection GoUnusedExportedFunction
func SliceUintPtr[T any](buf []T) uintptr {
	return (*SliceStruct)(unsafe.Pointer(&buf)).Ptr
}

//goland:noinspection GoUnusedExportedFunction
func SlicePtr[T any](buf []T) unsafe.Pointer {
	//goland:noinspection GoVetUnsafePointer
	return unsafe.Pointer(SliceUintPtr(buf))
}

//goland:noinspection GoUnusedExportedFunction
func StrUintPtr(s string) uintptr {
	return (*SliceStruct)(unsafe.Pointer(&s)).Ptr
}

//goland:noinspection GoUnusedExportedFunction
func StrPtr(s string) unsafe.Pointer {
	//goland:noinspection GoVetUnsafePointer
	return unsafe.Pointer(StrUintPtr(s))
}

//goland:noinspection GoUnusedExportedFunction
func StrToBytes(s string) []byte {
	return PtrToSlice((*StrStruct)(unsafe.Pointer(&s)).Ptr, len(s))
}

//goland:noinspection GoUnusedExportedFunction
func BytesToStr(buf []byte) string {
	return *(*string)(unsafe.Pointer(&buf))
}

//goland:noinspection GoUnusedExportedFunction
func PtrToSlice(p uintptr, i int) []byte {
	if i < 0 {
		i = 0
	}

	o := SliceStruct{
		Ptr: p,
		Len: i,
		Cap: i,
	}
	return *(*[]byte)(unsafe.Pointer(&o))
}

// PtrToStr fast read as a string
// Don't use as string convert
// use string(PtrToSlice(p, i)) inside
//
//goland:noinspection GoUnusedExportedFunction
func PtrToStr(p uintptr, i int) string {
	if i < 0 {
		i = 0
	}

	o := StrStruct{
		Ptr: p,
		Len: i,
	}
	return *(*string)(unsafe.Pointer(&o))
}

//endregion

// ParseUint parses uint from buf.
//
//goland:noinspection GoUnusedExportedFunction
func ParseUint(buf []byte) (int, error) {
	v, n, err := ParseUintBuf(buf)
	if n != len(buf) {
		return -1, errUnexpectedTrailingChar
	}
	return v, err
}

// ParseUfloat parses unsigned float from buf.
//
//goland:noinspection GoUnusedExportedFunction
func ParseUfloat(buf []byte) (float64, error) {
	v, n, err := ParseUfloatBuf(buf)
	if n != len(buf) {
		return -1, errUnexpectedFloatTrailingChar
	}
	if err != nil {
		return -1, err
	}
	return v, err
}

func BinStrtod(p uintptr, n int, radix int) (float64, error) {
	/*
	 * Verify the validity of the current character as a base-radix digit.  In
	 * the event that an invalid digit is found, halt the conversion and
	 * return the portion which has been converted thus far.
	 */
	if radix == 0 {
		radix = 10
	}
	if radix > 36 || radix < 2 {
		return 0.0, errUnexpectedRadix
	}

	if n <= 0 {
		return 0.0, errEmptyInt
	}

	val := 0.0
	numFound := false
	i := 0
	radix_ := float64(radix)
	for ; i < n; i++ {
		c := *(*byte)(PtrAdd(p, i))
		if c == '_' {
			if !numFound {
				return 0.0, errUnexpectedFirstChar
			}
			continue
		}

		v := int(NumChars[c])
		if v >= radix || v < 0 {
			if !numFound {
				return 0.0, errUnexpectedFirstChar
			}
			return val, nil
		}
		numFound = true
		val = val*radix_ + float64(v)
	}
	return val, nil
}

//goland:noinspection GoUnusedExportedFunction
func Strtold(p uintptr, n int) (float64, uintptr, error) {
	buf := PtrToSlice(p, n)
	v, n, err := ParseUfloatBuf(buf)
	if err != nil {
		return 0.0, p, err
	}

	return v, p + uintptr(n), nil
}

//goland:noinspection GoUnusedExportedFunction
func Strtoul(p uintptr, n int, radix int) (uint64, bool, uintptr, error) {
	if radix == 0 {
		radix = 10
	}
	if radix > 36 || radix < 2 {
		return 0, false, p, errUnexpectedRadix
	}

	if n <= 0 {
		return 0, false, p, errEmptyInt
	}

	val := uint64(0)
	numFound := false
	i := 0
	for ; i < n; i++ {
		c := *(*byte)(PtrAdd(p, i))
		if c == '_' {
			if !numFound {
				return 0, false, p, errUnexpectedFirstChar
			}
			continue
		}

		v := int(NumChars[c])
		if v >= radix || v < 0 {
			if !numFound {
				return 0, false, p, errUnexpectedFirstChar
			}
			return val, false, p + uintptr(i), nil
		}
		numFound = true
		tmp := val*uint64(radix) + uint64(v)
		if val > 0 && val >= tmp {
			return 0, true, p, errTooLongInt
		}
		val = tmp
	}
	return val, false, p + uintptr(i), nil
}

//goland:noinspection GoUnusedExportedFunction
func ParseUintBuf(buf []byte) (int, int, error) {
	n := len(buf)
	if n == 0 {
		return -1, 0, errEmptyInt
	}
	v := 0
	for i := 0; i < n; i++ {
		c := buf[i]
		if c >= '0' && c <= '9' {
			k := c - '0'
			vNew := 10*v + int(k)
			// Test for overflow.
			if vNew < v {
				return -1, i, errTooLongInt
			}
			v = vNew
			continue
		}

		if i == 0 {
			return -1, i, errUnexpectedFirstChar
		}
		return v, i, nil
	}
	return v, n, nil
}

//goland:noinspection GoUnusedExportedFunction
func ParseUfloatBuf(buf []byte) (float64, int, error) {
	n := len(buf)
	if n == 0 {
		return 0.0, 0, errEmptyFloat
	}

	var v uint64
	offset := 1.0
	pointFound := false
	numFound := false
	i := 0
	for ; i < n; i++ {
		c := buf[i]
		if c == '_' {
			continue
		}

		if c >= '0' && c <= '9' {
			v = 10*v + uint64(c-'0')
			if pointFound {
				offset /= 10
			}
			numFound = true
			continue
		}

		if c == '.' {
			if pointFound {
				return 0.0, i, errDuplicateFloatPoint
			}
			pointFound = true
			continue
		}

		if c == 'e' || c == 'E' {
			if i+1 >= n {
				return 0.0, i, errUnexpectedFloatEnd
			}
			tmp := buf[i+1:]
			i += 1
			minus := -1
			switch tmp[0] {
			case '+':
				tmp = tmp[1:]
				i += 1
				minus = 1
			case '-':
				tmp = tmp[1:]
				i += 1
			default:
				minus = 1
			}
			vv, nn, err := ParseUintBuf(tmp)
			if err != nil {
				return 0.0, i + nn, errInvalidFloatExponent
			}
			tmpv := float64(v) * offset * math.Pow10(minus*vv)
			return tmpv, i + nn, nil
		}

		break
	}

	if !numFound {
		return 0.0, 0, errUnexpectedFirstChar
	}
	return float64(v) * offset, i, nil
}
