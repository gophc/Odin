package util

import (
	"unsafe"
)

//goland:noinspection GoSnakeCaseUsage
const (
	BOM_UTF32_BE = "\x00\x00\xfe\xff"
	BOM_UTF32_LE = "\xff\xfe\x00\x00"
	BOM_UTF16_BE = "\xfe\xff"
	BOM_UTF16_LE = "\xff\xfe"
	BOM_UTF8     = "\xef\xbb\xbf"

	BOM_UTF8_RUNE = rune(65279)
)

type _uSliceH struct {
	Ptr uintptr
	Len int
	Cap int
}

//goland:noinspection GoUnusedExportedFunction
func IsSysBigEndian() bool {
	dap := uintptr(1)
	return *(*byte)(unsafe.Pointer(&dap)) != 1
}

//region utf helper

//go:linkname fastRand64 runtime.fastrand64
func fastRand64() uint64

func convItoa(i int) string {
	buf := make([]byte, 0, 22)
	neg := i < 0
	if neg {
		i = -i
	}
	if i == 0 {
		return "0"
	}
	for i > 0 {
		buf = append(buf, byte(i%10+48))
		i = i / 10
	}
	if neg {
		buf = append(buf, '-')
	}
	l := len(buf) - 1
	for j := 0; j <= l/2; j++ {
		buf[j], buf[l-j] = buf[l-j], buf[j]
	}
	return string(buf)
}

func errNew(text string) error {
	return &errString{s: text}
}

type errString struct {
	s string
}

func (e *errString) Error() string {
	return e.s
}

//goland:noinspection GoUnusedConst
const (
	replacementChar = '\uFFFD'     // Unicode replacement character
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.

	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000

	t1 = 0b00000000
	tx = 0b10000000
	t2 = 0b11000000
	t3 = 0b11100000
	t4 = 0b11110000
	t5 = 0b11111000

	maskx = 0b00111111
	mask2 = 0b00011111
	mask3 = 0b00001111
	mask4 = 0b00000111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1

	// The default lowest and highest continuation byte.
	locb = 0b10000000
	hicb = 0b10111111

	// These names of these constants are chosen to give nice alignment in the
	// table below. The first nibble is an index into acceptRanges or F for
	// special one-byte cases. The second nibble is the Rune length or the
	// Status for the special one-byte case.
	xx = 0xF1 // invalid: size 1
	as = 0xF0 // ASCII: size 1
	s1 = 0x02 // accept 0, size 2
	s2 = 0x13 // accept 1, size 3
	s3 = 0x03 // accept 0, size 3
	s4 = 0x23 // accept 2, size 3
	s5 = 0x34 // accept 3, size 4
	s6 = 0x04 // accept 0, size 4
	s7 = 0x44 // accept 4, size 4

	RuneError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"
	RuneSelf  = 0x80         // characters below RuneSelf are represented as themselves in a single byte.
	MaxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	UTFMax    = 4            // maximum number of bytes of a UTF-8 encoded Unicode character.
)

// Code points in the surrogate range are not valid for UTF-8.
const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
)

func Utf16FromStr(str string, bigEndian bool) []byte {
	if len(str) <= 0 {
		return []byte{}
	}

	buf := make([]uint16, len(str)/2+4)
	bup := (*_uSliceH)(unsafe.Pointer(&buf)).Ptr

	r1, r2, i16 := int32(0), int32(0), 0
	for idx, v := range str {
		if idx == 0 && v == BOM_UTF8_RUNE { // utf8 bom
			continue
		}

		if i16+2 > len(buf) {
			buf = append(buf, 0, 0)
			(*_uSliceH)(unsafe.Pointer(&buf)).Len = cap(buf)
			bup = (*_uSliceH)(unsafe.Pointer(&buf)).Ptr
		}

		//goland:noinspection GoVetUnsafePointer
		u16 := (*[2]uint16)(unsafe.Pointer(bup + uintptr(i16*2)))

		switch {
		case 0 <= v && v < surr1, surr3 <= v && v < surrSelf:
			// normal rune
			u16[0] = uint16(v)
			i16 += 1
		case surrSelf <= v && v <= maxRune:
			// needs surrogate sequence
			if v < surrSelf || v > maxRune {
				r1, r2 = replacementChar, replacementChar
			} else {
				v -= surrSelf
				r1, r2 = surr1+(v>>10)&0x3ff, surr2+v&0x3ff
			}
			u16[0] = uint16(r1)
			u16[1] = uint16(r2)
			i16 += 2
		default:
			u16[0] = uint16(replacementChar)
			i16 += 1
		}
	}
	buf = buf[:i16]

	bup = uintptr(1)
	if (*(*byte)(unsafe.Pointer(&bup)) != 1) != bigEndian {
		ChgEndianU16((*_uSliceH)(unsafe.Pointer(&buf)).Ptr, len(buf)*2, bigEndian)
	}

	data := *(*[]byte)(unsafe.Pointer(&buf))
	(*_uSliceH)(unsafe.Pointer(&data)).Len = len(buf) * 2
	(*_uSliceH)(unsafe.Pointer(&data)).Cap = cap(buf) * 2
	return data
}

func Utf16EncodeRune(v rune) (r1, r2 rune) {
	if v < surrSelf || v > maxRune {
		return replacementChar, replacementChar
	}
	v -= surrSelf
	return surr1 + (v>>10)&0x3ff, surr2 + v&0x3ff
}

//goland:noinspection GoUnusedExportedFunction
func Utf16Encode(s []rune) []uint16 {
	n := len(s)
	for _, v := range s {
		if v >= surrSelf {
			n++
		}
	}

	a := make([]uint16, n)
	n = 0
	for _, v := range s {
		switch {
		case 0 <= v && v < surr1, surr3 <= v && v < surrSelf:
			// normal rune
			a[n] = uint16(v)
			n++
		case surrSelf <= v && v <= maxRune:
			// needs surrogate sequence
			r1, r2 := Utf16EncodeRune(v)
			a[n] = uint16(r1)
			a[n+1] = uint16(r2)
			n += 2
		default:
			a[n] = uint16(replacementChar)
			n++
		}
	}
	return a[:n]
}

func Utf16DecodeRune(r1, r2 rune) rune {
	if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
		return (r1-surr1)<<10 | (r2 - surr2) + surrSelf
	}
	return replacementChar
}

func Utf16Decode(s []uint16) []rune {
	a := make([]rune, len(s))
	n := 0
	for i := 0; i < len(s); i++ {
		switch r := s[i]; {
		case r < surr1, surr3 <= r:
			// normal rune
			a[n] = rune(r)
		case surr1 <= r && r < surr2 && i+1 < len(s) &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			a[n] = Utf16DecodeRune(rune(r), rune(s[i+1]))
			i++
		default:
			// invalid surrogate sequence
			a[n] = replacementChar
		}
		n++
	}
	return a[:n]
}

func ChgEndianU16(dap uintptr, ln int, bigEndian bool) {
	if ln <= 0 {
		return
	}

	if ln%2 != 0 {
		panic("ChgEndianU16 must len() % 2 == 0")
	}

	if IsSysBigEndian() != bigEndian {
		end := dap + uintptr(ln)
		if bigEndian {
			for ; dap < end; dap += 2 {
				//goland:noinspection GoVetUnsafePointer
				*(*uint16)(unsafe.Pointer(dap)) = bigEndianUint16(*(*[2]byte)(unsafe.Pointer(dap)))
			}
		} else {
			for ; dap < end; dap += 2 {
				//goland:noinspection GoVetUnsafePointer
				*(*uint16)(unsafe.Pointer(dap)) = littleEndianUint16(*(*[2]byte)(unsafe.Pointer(dap)))
			}
		}
	}
}

//goland:noinspection GoUnusedExportedFunction
func ChgEndianU32(dap uintptr, ln int, bigEndian bool) {
	if ln <= 0 {
		return
	}

	if ln%4 != 0 {
		panic("ChgEndianU32 must len() % 4 == 0")
	}

	if IsSysBigEndian() != bigEndian {
		end := dap + uintptr(ln)
		if bigEndian {
			for ; dap < end; dap += 4 {
				//goland:noinspection GoVetUnsafePointer
				*(*uint32)(unsafe.Pointer(dap)) = bigEndianUint32(*(*[4]byte)(unsafe.Pointer(dap)))
			}
		} else {
			for ; dap < end; dap += 4 {
				//goland:noinspection GoVetUnsafePointer
				*(*uint32)(unsafe.Pointer(dap)) = littleEndianUint32(*(*[4]byte)(unsafe.Pointer(dap)))
			}
		}
	}
}

func littleEndianUint32(b [4]byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func bigEndianUint32(b [4]byte) uint32 {
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func littleEndianUint16(b [2]byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func bigEndianUint16(b [2]byte) uint16 {
	return uint16(b[1]) | uint16(b[0])<<8
}

func bigEndianUint64(b [8]byte) uint64 {
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

//endregion

//region Ut2Utf8

func Ut2Utf8(data []byte) []byte {
	ln := len(data)
	if ln < 2 {
		return data
	}

	c0 := int(data[0])
	if c0 == '\xef' || c0 == '\xfe' || c0 == '\xff' || c0 == '\x00' {
		c1, c2, c3 := int(data[1]), 65536, 65536
		if ln >= 4 {
			c2, c3 = int(data[2]), int(data[3])
		} else if ln >= 3 {
			c2 = int(data[2])
		}

		if c0 == '\xef' && c1 == '\xbb' && c2 == '\xbf' {
			return data[3:]
		}

		if c0 == '\xff' && c1 == '\xfe' {
			if c2 == '\x00' && c3 == '\x00' {
				return Utf32ToUtf8(data[4:], false)
			} else {
				return Utf16ToUtf8(data[2:], false)
			}
		} else if c0 == '\xfe' && c1 == '\xff' {
			return Utf16ToUtf8(data[2:], true)
		} else if c0 == '\x00' && c1 == '\x00' && c2 == '\xfe' && c3 == '\xff' {
			return Utf32ToUtf8(data[4:], true)
		}
	}

	return data
}

func Utf32ToUtf8(src []byte, bigEndian bool) []byte {
	ln := len(src)
	if ln <= 0 {
		return []byte{}
	}
	if ln%4 != 0 {
		panic("utf-32 decode err len() % 4 != 0")
	}

	buf := *(*[]uint32)(unsafe.Pointer(&src))
	(*_uSliceH)(unsafe.Pointer(&buf)).Len = ln >> 2
	(*_uSliceH)(unsafe.Pointer(&buf)).Cap = ln >> 2

	bup := uintptr(1)
	chg := (*(*byte)(unsafe.Pointer(&bup)) != 1) != bigEndian
	bup = (*_uSliceH)(unsafe.Pointer(&buf)).Ptr

	data := make([]byte, ln)
	(*_uSliceH)(unsafe.Pointer(&data)).Len = cap(data)
	dap := (*_uSliceH)(unsafe.Pointer(&data)).Ptr

	end, i8, r := bup+uintptr(ln), 0, uint32(0)
	for ; bup < end; bup += 4 {
		//goland:noinspection GoVetUnsafePointer
		if chg {
			if bigEndian {
				r = bigEndianUint32(*(*[4]byte)(unsafe.Pointer(bup)))
			} else {
				r = littleEndianUint32(*(*[4]byte)(unsafe.Pointer(bup)))
			}
		} else {
			r = *(*uint32)(unsafe.Pointer(bup))
		}

		//goland:noinspection GoVetUnsafePointer
		u8 := (*[4]byte)(unsafe.Pointer(dap + uintptr(i8)))

		switch {
		case r <= rune1Max:
			u8[0] = byte(r)
			i8 += 1
		case r <= rune2Max:
			u8[0] = t2 | byte(r>>6)
			u8[1] = tx | byte(r)&maskx
			i8 += 2
		case r > MaxRune, surrogateMin <= r && r <= surrogateMax:
			r = RuneError
			fallthrough
		case r <= rune3Max:
			u8[0] = t3 | byte(r>>12)
			u8[1] = tx | byte(r>>6)&maskx
			u8[2] = tx | byte(r)&maskx
			i8 += 3
		default:
			u8[0] = t4 | byte(r>>18)
			u8[1] = tx | byte(r>>12)&maskx
			u8[2] = tx | byte(r>>6)&maskx
			u8[3] = tx | byte(r)&maskx
			i8 += 4
		}
	}

	return data[:i8]
}

func Utf16ToUtf8(src []byte, bigEndian bool) []byte {
	ln := len(src)
	if ln <= 0 {
		return []byte{}
	}
	if ln%2 != 0 {
		panic("utf-16 decode err len() % 2 != 0")
	}

	buf := *(*[]uint16)(unsafe.Pointer(&src))
	(*_uSliceH)(unsafe.Pointer(&buf)).Len = ln >> 1
	(*_uSliceH)(unsafe.Pointer(&buf)).Cap = ln >> 1

	bup := uintptr(1)
	chg := (*(*byte)(unsafe.Pointer(&bup)) != 1) != bigEndian
	bup = (*_uSliceH)(unsafe.Pointer(&buf)).Ptr

	data := make([]byte, 3*ln>>1+4)
	(*_uSliceH)(unsafe.Pointer(&data)).Len = cap(data)
	dap := (*_uSliceH)(unsafe.Pointer(&data)).Ptr

	end, i8, ru, r16 := bup+uintptr(ln), 0, rune(0), uint16(0)
	for ; bup < end; bup += 2 {
		//goland:noinspection GoVetUnsafePointer
		if chg {
			if bigEndian {
				r16 = bigEndianUint16(*(*[2]byte)(unsafe.Pointer(bup)))
			} else {
				r16 = littleEndianUint16(*(*[2]byte)(unsafe.Pointer(bup)))
			}
		} else {
			r16 = *(*uint16)(unsafe.Pointer(bup))
		}

		switch {
		case r16 < surr1, surr3 <= r16:
			// normal rune
			ru = rune(r16)
		case surr1 <= r16 && r16 < surr2 && bup+2 < end:
			//goland:noinspection GoVetUnsafePointer
			r16n := *(*uint16)(unsafe.Pointer(bup + 2))
			if surr2 <= r16n && r16n < surr3 {
				// valid surrogate sequence
				r1, r2 := rune(r16), rune(r16n)
				if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
					ru = (r1-surr1)<<10 | (r2 - surr2) + surrSelf
				} else {
					ru = replacementChar
				}
				bup += 2
			}
		default:
			// invalid surrogate sequence
			ru = replacementChar
		}

		r := uint32(ru)
		if i8+4 > len(data) {
			data = append(data, '\x00', '\x00', '\x00', '\x00')
			(*_uSliceH)(unsafe.Pointer(&data)).Len = cap(data)
			dap = (*_uSliceH)(unsafe.Pointer(&data)).Ptr
		}

		//goland:noinspection GoVetUnsafePointer
		u8 := (*[4]byte)(unsafe.Pointer(dap + uintptr(i8)))
		switch {
		case r <= rune1Max:
			u8[0] = byte(r)
			i8 += 1
		case r <= rune2Max:
			u8[0] = t2 | byte(r>>6)
			u8[1] = tx | byte(r)&maskx
			i8 += 2
		case r > MaxRune, surrogateMin <= r && r <= surrogateMax:
			r = RuneError
			fallthrough
		case r <= rune3Max:
			u8[0] = t3 | byte(r>>12)
			u8[1] = tx | byte(r>>6)&maskx
			u8[2] = tx | byte(r)&maskx
			i8 += 3
		default:
			u8[0] = t4 | byte(r>>18)
			u8[1] = tx | byte(r>>12)&maskx
			u8[2] = tx | byte(r>>6)&maskx
			u8[3] = tx | byte(r)&maskx
			i8 += 4
		}
	}

	return data[:i8]
}

//endregion

//region base32768

//goland:noinspection GoUnusedConst,GoSnakeCaseUsage
const (
	BITS_PER_LAST_CHAR = 7
	BITS_PER_CHAR      = 15 // Base32768 is a 15-bit encoding
	BITS_PER_BYTE      = 8

	lookupM   = uint32(0xff00_0000)
	lookupM7  = uint32(0x0700_0000)
	lookupM15 = uint32(0x0f00_0000)

	pairString0 = `ҠҿԀԟڀڿݠޟ߀ߟကဟႠႿᄀᅟᆀᆟᇠሿበቿዠዿጠጿᎠᏟᐠᙟᚠᛟកសᠠᡟᣀᣟᦀᦟ᧠᧿ᨠᨿᯀᯟᰀᰟᴀᴟ⇠⇿⋀⋟⍀⏟␀␟─❟➀➿⠀⥿⦠⦿⨠⩟⪀⪿⫠⭟ⰀⰟⲀⳟⴀⴟⵀⵟ⺠⻟㇀㇟㐀䶟䷀龿ꀀꑿ꒠꒿ꔀꗿꙀꙟꚠꛟ꜀ꝟꞀꞟꡀꡟ`
	pairString1 = `ƀƟɀʟ`
)

type Base32768 struct {
	lookupE7  uintptr
	lookupE15 uintptr
	lookupE   [32896]uint16
	lookupD   [65536]uint32
}

func NewBase32768() *Base32768 {
	enc := &Base32768{
		lookupE7:  uintptr(0),
		lookupE15: uintptr(0),
		lookupE:   [32896]uint16{},
		lookupD:   [65536]uint32{},
	}

	n, n1 := 0, 0
	for r, pairString := range []string{pairString0, pairString1} {
		// Decompression
		first, last := rune(0), rune(0)
		for _, ru := range pairString {
			if first == 0 {
				first = ru
			} else {
				last = ru
			}
			if first > 0 && last > 0 {
				for codePoint := first; codePoint <= last; codePoint++ {
					enc.lookupE[n] = uint16(codePoint)
					n += 1
				}
				first, last = 0, 0
			}
		}

		numZBits := BITS_PER_CHAR - BITS_PER_BYTE*r // 0 -> 15, 1 -> 7
		if r == 0 {
			n1 = n
			encodeRepertoire := enc.lookupE[0:n]
			enc.lookupE15 = (*_uSliceH)(unsafe.Pointer(&encodeRepertoire)).Ptr
			for z, chr := range encodeRepertoire {
				enc.lookupD[chr] = uint32(numZBits<<24 + z)
			}
		} else {
			encodeRepertoire := enc.lookupE[n1:n]
			enc.lookupE7 = (*_uSliceH)(unsafe.Pointer(&encodeRepertoire)).Ptr
			for z, chr := range encodeRepertoire {
				enc.lookupD[chr] = uint32(numZBits<<24 + z)
			}
		}
	}
	return enc
}

func (enc *Base32768) Encode(src []byte, bigEndian bool, withBom bool) []byte {
	ln := len(src)
	if ln <= 0 {
		return []byte{}
	}

	bup := uintptr(1)
	sysBigEndian := *(*byte)(unsafe.Pointer(&bup)) != 1

	nn := uint64(ln*BITS_PER_BYTE+BITS_PER_CHAR-1) / BITS_PER_CHAR
	if withBom {
		nn += 2
	}
	buf := make([]uint16, nn)
	bup = (*_uSliceH)(unsafe.Pointer(&buf)).Ptr

	dap := (*_uSliceH)(unsafe.Pointer(&src)).Ptr
	z, numZBits, endp := uint64(0), 0, dap+uintptr(ln)

	//goland:noinspection GoVetUnsafePointer
	for ; dap < endp-30; dap += 30 {
		m0 := bigEndianUint64(*(*[8]byte)(unsafe.Pointer(dap)))
		m1 := bigEndianUint64(*(*[8]byte)(unsafe.Pointer(dap + 8)))
		m2 := bigEndianUint64(*(*[8]byte)(unsafe.Pointer(dap + 16)))
		m3 := bigEndianUint64(*(*[8]byte)(unsafe.Pointer(dap + 24)))

		u16 := (*[16]uint16)(unsafe.Pointer(bup))
		u16[0] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(m0>>48)&0b1111_1111_1111_1110))
		u16[1] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>33)&0b1111_1111_1111_1110)))
		u16[2] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>18)&0b1111_1111_1111_1110)))
		u16[3] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>3)&0b1111_1111_1111_1110)))

		m0 = m0<<60 | m1>>4
		u16[4] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(m0>>48)&0b1111_1111_1111_1110))
		u16[5] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>33)&0b1111_1111_1111_1110)))
		u16[6] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>18)&0b1111_1111_1111_1110)))
		u16[7] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>3)&0b1111_1111_1111_1110)))

		m0 = m1<<56 | m2>>8
		u16[8] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(m0>>48)&0b1111_1111_1111_1110))
		u16[9] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>33)&0b1111_1111_1111_1110)))
		u16[10] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>18)&0b1111_1111_1111_1110)))
		u16[11] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>3)&0b1111_1111_1111_1110)))

		m0 = m2<<52 | m3>>12
		u16[12] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(m0>>48)&0b1111_1111_1111_1110))
		u16[13] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>33)&0b1111_1111_1111_1110)))
		u16[14] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>18)&0b1111_1111_1111_1110)))
		u16[15] = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr((m0>>3)&0b1111_1111_1111_1110)))

		bup += 32
	}

	for ; dap < endp; dap++ {
		//goland:noinspection GoVetUnsafePointer
		u8 := *(*byte)(unsafe.Pointer(dap))
		z = (z << 8) + uint64(u8)
		numZBits += 8
		// Take most significant bit first
		if numZBits >= BITS_PER_CHAR {
			numZBits -= BITS_PER_CHAR
			nn = z >> numZBits
			z = z - (nn << numZBits)

			//goland:noinspection GoVetUnsafePointer
			*(*uint16)(unsafe.Pointer(bup)) = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(nn*2)))
			bup += 2
		}
	}

	if numZBits != 0 {
		// Final bits require special treatment.

		// z = bbbbbbcccccccc, numZBits = 14, padBits = 1
		// z = bbbbbcccccccc, numZBits = 13, padBits = 2
		// z = bbbbcccccccc, numZBits = 12, padBits = 3
		// z = bbbcccccccc, numZBits = 11, padBits = 4
		// z = bbcccccccc, numZBits = 10, padBits = 5
		// z = bcccccccc, numZBits = 9, padBits = 6
		// z = cccccccc, numZBits = 8, padBits = 7
		// => Pad `z` out to 15 bits using 1s, then encode as normal (r = 0)

		// z = ccccccc, numZBits = 7, padBits = 0
		// z = cccccc, numZBits = 6, padBits = 1
		// z = ccccc, numZBits = 5, padBits = 2
		// z = cccc, numZBits = 4, padBits = 3
		// z = ccc, numZBits = 3, padBits = 4
		// z = cc, numZBits = 2, padBits = 5
		// z = c, numZBits = 1, padBits = 6
		// => Pad `z` out to 7 bits using 1s, then encode specially (r = 1)
		if numZBits <= BITS_PER_LAST_CHAR {
			numZBits = BITS_PER_LAST_CHAR - numZBits
			z = (z << numZBits) | (0x7f >> (BITS_PER_LAST_CHAR - numZBits))
			//goland:noinspection GoVetUnsafePointer
			*(*uint16)(unsafe.Pointer(bup)) = *(*uint16)(unsafe.Pointer(enc.lookupE7 + uintptr(z*2)))
		} else {
			numZBits = BITS_PER_CHAR - numZBits
			z = (z << numZBits) | (0x7fff >> (BITS_PER_CHAR - numZBits))
			//goland:noinspection GoVetUnsafePointer
			*(*uint16)(unsafe.Pointer(bup)) = *(*uint16)(unsafe.Pointer(enc.lookupE15 + uintptr(z*2)))
		}
	}

	ln = len(buf) * 2
	data := *(*[]byte)(unsafe.Pointer(&buf))
	(*_uSliceH)(unsafe.Pointer(&data)).Len = ln
	(*_uSliceH)(unsafe.Pointer(&data)).Cap = cap(buf) * 2

	if sysBigEndian != bigEndian {
		dap = (*_uSliceH)(unsafe.Pointer(&data)).Ptr
		if withBom {
			if bigEndian {
				data[0], data[1] = '\xfe', '\xff'
			} else {
				data[0], data[1] = '\xff', '\xfe'
			}
			dap += 2
			ln -= 2
		}
		ChgEndianU16(dap, ln, bigEndian)
	}
	return data
}

func (enc *Base32768) EncodeStr(src []byte) string {
	bigEndian := IsSysBigEndian()
	buf := enc.Encode(src, bigEndian, false)
	u8 := Utf16ToUtf8(buf, bigEndian)
	return *(*string)(unsafe.Pointer(&u8))
}

func (enc *Base32768) Decode(src []byte, bigEndian bool) ([]byte, error) {
	ln := len(src)
	if ln <= 0 {
		return []byte{}, nil
	}
	if ln%2 != 0 {
		panic("Decode utf-16 src must len() % 2 == 0")
	}

	bup := uintptr(1)
	sysBigEndian := *(*byte)(unsafe.Pointer(&bup)) != 1

	buf := *(*[]uint16)(unsafe.Pointer(&src))
	(*_uSliceH)(unsafe.Pointer(&buf)).Len = ln >> 1
	(*_uSliceH)(unsafe.Pointer(&buf)).Cap = ln >> 1
	bup = (*_uSliceH)(unsafe.Pointer(&buf)).Ptr

	c0, c1 := src[0], src[1]
	if c0 == '\xfe' && c1 == '\xff' {
		bigEndian = true
		bup += 2
		ln -= 2
	} else if c0 == '\xff' && c1 == '\xfe' {
		bigEndian = false
		bup += 2
		ln -= 2
	}

	// This length is a guess. There's a chance we allocate one more byte here
	// than we actually need. But we can count and slice it off later
	data := make([]byte, (ln*BITS_PER_CHAR/2+BITS_PER_LAST_CHAR)/BITS_PER_BYTE+2)
	dap := (*_uSliceH)(unsafe.Pointer(&data)).Ptr

	bi, m, bufu := 0, uint64(0), [16]uint16{}
	bufm := [16]uint8{15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15, 15}
	r16, endp := uint16(0), bup+uintptr(ln)

	for ; bup < endp; bup += 2 {
		//goland:noinspection GoVetUnsafePointer
		if sysBigEndian == bigEndian {
			r16 = *(*uint16)(unsafe.Pointer(bup))
		} else {
			if bigEndian {
				r16 = bigEndianUint16(*(*[2]byte)(unsafe.Pointer(bup)))
			} else {
				r16 = littleEndianUint16(*(*[2]byte)(unsafe.Pointer(bup)))
			}
		}
		tmp := enc.lookupD[r16]
		if tmp&lookupM == lookupM7 {
			if bup != endp-2 {
				return nil, errNew("secondary character found before end of input at position " + convItoa(int(endp-(*_uSliceH)(unsafe.Pointer(&buf)).Ptr)/2))
			}

			bufm[bi] = BITS_PER_LAST_CHAR
			bufu[bi] = uint16(tmp) & 0x007f
			bi += 1
			break
		} else if tmp&lookupM == lookupM15 {
			bufu[bi] = uint16(tmp)
			bi += 1

			if bi >= 16 {
				//goland:noinspection GoVetUnsafePointer
				u15 := (*[4]uint64)(unsafe.Pointer(dap))
				m = (uint64(bufu[0]) << 49) | (uint64(bufu[1]) << 34) | (uint64(bufu[2]) << 19) | (uint64(bufu[3]) << 4) | uint64(bufu[4]>>11)
				u15[0] = bigEndianUint64(*(*[8]byte)(unsafe.Pointer(&m)))
				m = uint64(bufu[4] & 0x07FF)

				m = (m << 53) | (uint64(bufu[5]) << 38) | (uint64(bufu[6]) << 23) | (uint64(bufu[7]) << 8) | uint64(bufu[8]>>7)
				u15[1] = bigEndianUint64(*(*[8]byte)(unsafe.Pointer(&m)))
				m = uint64(bufu[8] & 0b01111111)

				m = (m << 57) | (uint64(bufu[9]) << 42) | (uint64(bufu[10]) << 27) | (uint64(bufu[11]) << 12) | uint64(bufu[12]>>3)
				u15[2] = bigEndianUint64(*(*[8]byte)(unsafe.Pointer(&m)))
				m = uint64(bufu[12] & 0b00000111)

				m = (m << 61) | (uint64(bufu[13]) << 46) | (uint64(bufu[14]) << 31) | (uint64(bufu[15]) << 16)
				u15[3] = bigEndianUint64(*(*[8]byte)(unsafe.Pointer(&m)))

				dap += 30
				bi = 0
			}
		} else {
			return nil, errNew("unrecognised Base32768 character: " + string(Utf16Decode([]uint16{r16})))
		}
	}

	uv8, numUint8Bits := uint64(0), 0
	for i := 0; i < bi; i++ {
		z, numZBits := bufu[i], bufm[i]
		// Take most significant bit first
		uv8 = (uv8 << numZBits) | uint64(z)
		numUint8Bits += int(numZBits)
		for numUint8Bits >= BITS_PER_BYTE {
			numUint8Bits -= BITS_PER_BYTE
			m = uv8 >> numUint8Bits
			uv8 = uv8 - (m << numUint8Bits)
			//goland:noinspection GoVetUnsafePointer
			*(*byte)(unsafe.Pointer(dap)) = byte(m)
			dap += 1
		}
	}

	// Final padding bits! Requires special consideration!
	// Remember how we always pad with 1s?
	// Note: there could be 0 such bits, check still works though
	if uv8 != uint64((1<<numUint8Bits)-1) {
		return nil, errNew("padding mismatch")
	}

	(*_uSliceH)(unsafe.Pointer(&data)).Len = int(dap - (*_uSliceH)(unsafe.Pointer(&data)).Ptr)
	return data, nil
}

func (enc *Base32768) DecodeStr(src string) ([]byte, error) {
	data := Utf16FromStr(src, false)
	return enc.Decode(data, false)
}

//endregion
