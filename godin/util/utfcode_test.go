package util

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"unsafe"
)

func TestUt2Utf8(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"", ""},
		{"a", "a"},
		{"abc", "abc"},
		{"你好", "你好"},
		{"你好你好", "你好你好"},
		{"\xE4\xBD\xA0\xE5\xA5\xBD\x61\x62\x63", "你好abc"},
		{"\xE4\xBD\xA0\xE5\xA5\xBD\xE4\xBD\xA0\xE5\xA5\xBD", "你好你好"},
		{"\xE4\xBD\xA0\xE5\xA5\xBD\x61\x62\x63", "你好abc"},
		{"\xE4\xBD\xA0\xE5\xA5\xBDabc", "你好abc"},
		{"\x61\x62\x63\xE4\xBD\xA0\xE5\xA5\xBD", "abc你好"},
		{"abc\xE4\xBD\xA0\xE5\xA5\xBD", "abc你好"},
	}
	for i, tt := range tests {
		t.Run("Ut2Utf8_"+strconv.Itoa(i), func(t *testing.T) {
			data := ([]byte)(tt.args)
			got := Ut2Utf8(data)
			if string(got) != tt.want {
				t.Errorf("Ut2Utf8() = %v, want %v", got, tt.want)
			}
		})
	}
	for i, tt := range tests {
		t.Run("Ut2Utf8_2_"+strconv.Itoa(i), func(t *testing.T) {
			data := ([]byte)(BOM_UTF8 + tt.args)
			got := Ut2Utf8(data)
			if string(got) != tt.want {
				t.Errorf("Ut2Utf8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtf32ToUtf8(t *testing.T) {
	tests := []struct {
		bigEndian bool
		data      string
		want      string
	}{
		{true, "\x00\x00\x4F\x60\x00\x00\x59\x7D", "你好"},

		{false, "\x60\x4F\x00\x00\x7D\x59\x00\x00\x60\x4F\x00\x00\x7D\x59\x00\x00", "你好你好"},
		{false, "\x60\x4F\x00\x00\x7D\x59\x00\x00", "你好"},
		{false, "\x61\x00\x00\x00\x62\x00\x00\x00\x63\x00\x00\x00", "abc"},
		{false, "\x61\x00\x00\x00\x62\x00\x00\x00\x63\x00\x00\x00\x60\x4F\x00\x00\x7D\x59\x00\x00", "abc你好"},
		{false, "\x60\x4F\x00\x00\x7D\x59\x00\x00\x61\x00\x00\x00\x62\x00\x00\x00\x63\x00\x00\x00", "你好abc"},
		{false, "", ""},

		{true, "\x00\x00\x4F\x60\x00\x00\x59\x7D\x00\x00\x4F\x60\x00\x00\x59\x7D", "你好你好"},
		{true, "\x00\x00\x4F\x60\x00\x00\x59\x7D", "你好"},
		{true, "\x00\x00\x00\x61\x00\x00\x00\x62\x00\x00\x00\x63", "abc"},
		{true, "\x00\x00\x00\x61\x00\x00\x00\x62\x00\x00\x00\x63\x00\x00\x4F\x60\x00\x00\x59\x7D", "abc你好"},
		{true, "\x00\x00\x4F\x60\x00\x00\x59\x7D\x00\x00\x00\x61\x00\x00\x00\x62\x00\x00\x00\x63", "你好abc"},
		{true, "", ""},
	}
	for i, tt := range tests {
		t.Run("u32_"+strconv.Itoa(i), func(t *testing.T) {
			data := ([]byte)(tt.data)
			got := Utf32ToUtf8(data, tt.bigEndian)
			if string(got) != tt.want {
				t.Errorf("Utf32ToUtf8() = %v, want %v", got, tt.want)
			}
		})
	}

	for i, tt := range tests {
		t.Run("ut2utf8_"+strconv.Itoa(i), func(t *testing.T) {
			_data := BOM_UTF32_LE + tt.data
			if tt.bigEndian {
				_data = BOM_UTF32_BE + tt.data
			}
			data := ([]byte)(_data)
			got := Ut2Utf8(data)
			if string(got) != tt.want {
				t.Errorf("Ut2Utf8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUtf16ToUtf8(t *testing.T) {
	tests := []struct {
		bigEndian bool
		data      string
		want      string
	}{
		{true, "\x4F\x60\x59\x7D", "你好"},

		{false, "\x60\x4F\x7D\x59\x60\x4F\x7D\x59", "你好你好"},
		{false, "\x60\x4F\x7D\x59", "你好"},
		{false, "\x61\x00\x62\x00\x63\x00", "abc"},
		{false, "\x61\x00\x62\x00\x63\x00\x60\x4F\x7D\x59", "abc你好"},
		{false, "\x60\x4F\x7D\x59\x61\x00\x62\x00\x63\x00", "你好abc"},
		{false, "", ""},

		{true, "\x4F\x60\x59\x7D\x4F\x60\x59\x7D", "你好你好"},
		{true, "\x4F\x60\x59\x7D", "你好"},
		{true, "\x00\x61\x00\x62\x00\x63", "abc"},
		{true, "\x00\x61\x00\x62\x00\x63\x4F\x60\x59\x7D", "abc你好"},
		{true, "\x4F\x60\x59\x7D\x00\x61\x00\x62\x00\x63", "你好abc"},
		{true, "", ""},
	}
	for i, tt := range tests {
		t.Run("u16_"+strconv.Itoa(i), func(t *testing.T) {
			data := ([]byte)(tt.data)
			got := Utf16ToUtf8(data, tt.bigEndian)
			if string(got) != tt.want {
				t.Errorf("Utf16ToUtf8() = %v, want %v", got, tt.want)
			}
		})
	}

	for i, tt := range tests {
		t.Run("ut2utf8_"+strconv.Itoa(i), func(t *testing.T) {
			_data := BOM_UTF16_LE + tt.data
			if tt.bigEndian {
				_data = BOM_UTF16_BE + tt.data
			}
			data := ([]byte)(_data)
			got := Ut2Utf8(data)
			if string(got) != tt.want {
				t.Errorf("Ut2Utf8() = %v, want %v", got, tt.want)
			}
		})
	}

	for i, tt := range tests {
		t.Run("Utf16FromStr_"+strconv.Itoa(i), func(t *testing.T) {
			_data := BOM_UTF16_LE + tt.data
			if tt.bigEndian {
				_data = BOM_UTF16_BE + tt.data
			}
			data := ([]byte)(_data)
			got := Ut2Utf8(data)

			data = Utf16FromStr(string(got), tt.bigEndian)
			got = Utf16ToUtf8(data, tt.bigEndian)
			if string(got) != tt.want {
				t.Errorf("Ut2Utf8() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testBase32768 struct {
	data []byte
	want string
}

func testBase32768Decode(bigEndian bool) (*Base32768, []*testBase32768, bool) {
	tests := []*testBase32768{
		{nil, ""},      // "媒腻㐤┖ꈳ埳"
		{[]byte{}, ""}, // [104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100]
		{Utf16FromStr("媒腻㐤┖ꈳ埳", bigEndian), "hello world"},
		{Utf16FromStr("媒腻㐤┖ꈳ埰ɣ", bigEndian), "hello world!"},
		{Utf16FromStr("妚臸铭㴶鞣挅鴬闒㩻㺜橮䨖ꀓ埵茮髈㱲疨檥浃迡堰ꄀ勁䓘㔘䷮䧶⛹包猾颈憐苨ⴇ絕腪鬑⡺梩傉縴♊䰴㤊睴胂憥䕞牮輅礲爻彔眼抜㻔焪ⴇ詄侑欄纶凔䓞疰澊殕ꄢ㮉擦溡佳✪Ⰵ娶觛損酀殻悎紻爅浃迡堰ꄀ保嬓⠌ᖧ稷䞋鏹蕄些㛔䍈ⴍ衦觛損酀殉䏔臯䱥壦顁霤碸骋㯝羪踍浖ꇻ簡⢶骽吗㶟咬綧凳廡煛", false), "function(o){for(var r=o.length,e=\"\",n=0,t=0,a=0;a<r;a++)for(var i=o[a],E=BITS_PER_BYTE-1;E>=0;E--){n=(n<<1)+(i>>E&1),++t===BITS_PER_CHAR&&(e+=lookupE[t][n],n=0,t=0)}if(0!==t){for(;!(t in lookupE);)n=1+(n<<1),t++;e+=lookupE[t][n]}return e}"},
	}

	enc := NewBase32768()
	return enc, tests, bigEndian
}

func TestBase32768Decode(t *testing.T) {
	enc, tests, bigEndian := testBase32768Decode(false)

	for i, tt := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			got, err := enc.Decode(tt.data, bigEndian)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}
			gots := string(got)
			if gots != tt.want {
				t.Errorf("Decode() got = %v, want %v", gots, tt.want)
			}

			got, err = enc.Decode(([]byte)(BOM_UTF16_LE+string(tt.data)), bigEndian)
			if err != nil {
				t.Errorf("BOM Decode() error = %v", err)
				return
			}
			gots = string(got)
			if gots != tt.want {
				t.Errorf("BOM Decode() got = %v, want %v", gots, tt.want)
			}

			got, err = enc.Decode(([]byte)(BOM_UTF16_LE+string(tt.data)), !bigEndian)
			if err != nil {
				t.Errorf("BOM Decode() error = %v", err)
				return
			}
			gots = string(got)
			if gots != tt.want {
				t.Errorf("BOM Decode() got = %v, want %v", gots, tt.want)
			}
		})
	}
}

func TestBase32768DecodeStr(t *testing.T) {
	enc, tests, bigEndian := testBase32768Decode(false)

	for i, tt := range tests {
		t.Run("test_str_"+strconv.Itoa(i), func(t *testing.T) {
			str := string(Utf16ToUtf8(tt.data, bigEndian))
			got, err := enc.DecodeStr(str)
			if err != nil {
				t.Errorf("DecodeStr() error = %v", err)
				return
			}
			gots := string(got)
			if gots != tt.want {
				t.Errorf("DecodeStr() got = %v, want %v", gots, tt.want)
			}

			got, err = enc.DecodeStr(BOM_UTF8 + str)
			if err != nil {
				t.Errorf("BOM DecodeStr() error = %v", err)
				return
			}
			gots = string(got)
			if gots != tt.want {
				t.Errorf("BOM DecodeStr() got = %v, want %v", gots, tt.want)
			}
		})
	}
}

func TestBase32768DecodeBOM(t *testing.T) {
	enc, tests, bigEndian := testBase32768Decode(false)

	for i, tt := range tests {
		t.Run("test_chg_"+strconv.Itoa(i), func(t *testing.T) {
			data := append([]byte{}, tt.data...)
			ChgEndianU16((*_uSliceH)(unsafe.Pointer(&data)).Ptr, len(data), !bigEndian)
			got, err := enc.Decode(data, !bigEndian)
			if err != nil {
				t.Errorf("ChgEndianU16 Decode() error = %v", err)
				return
			}
			gots := string(got)
			if gots != tt.want {
				t.Errorf("ChgEndianU16 Decode() got = %v, want %v", gots, tt.want)
			}

			data = append([]byte{'\xfe', '\xff'}, data...)
			got, err = enc.Decode(data, !bigEndian)
			if err != nil {
				t.Errorf("BOM ChgEndianU16 Decode() error = %v", err)
				return
			}
			gots = string(got)
			if gots != tt.want {
				t.Errorf("BOM ChgEndianU16 Decode() got = %v, want %v", gots, tt.want)
			}

			got, err = enc.Decode(data, bigEndian)
			if err != nil {
				t.Errorf("BOM ChgEndianU16 fix Decode() error = %v", err)
				return
			}
			gots = string(got)
			if gots != tt.want {
				t.Errorf("BOM ChgEndianU16 fix Decode() got = %v, want %v", gots, tt.want)
			}
		})
	}
}

func Test_rand_convItoa(t *testing.T) {
	num := 1000
	for i := 0; i < num; i++ {
		n1, n2 := fastRand64(), fastRand64()
		n := (int(n1) - int(n2)) / 2
		got := convItoa(n)
		want := strconv.Itoa(n)
		if got != want {
			t.Errorf("convItoa() = %v, want %v", got, want)
		}
	}
}

func TestEncoding_Base32768Encode(t *testing.T) {
	tests := []struct {
		data []byte
		want string
	}{
		{[]byte{}, ""}, // "媒腻㐤┖ꈳ埳"
		{[]byte{}, ""}, // [104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100]
		{Utf16FromStr("媒腻㐤┖ꈳ埳", false), "hello world"},
		{Utf16FromStr("媒腻㐤┖ꈳ埰ɣ", false), "hello world!"},
		{Utf16FromStr("妚臸铭㴶鞣挅鴬闒㩻㺜橮䨖ꀓ埵茮髈㱲疨檥浃迡堰ꄀ勁䓘㔘䷮䧶⛹包猾颈憐苨ⴇ絕腪鬑⡺梩傉縴♊䰴㤊睴胂憥䕞牮輅礲爻彔眼抜㻔焪ⴇ詄侑欄纶凔䓞疰澊殕ꄢ㮉擦溡佳✪Ⰵ娶觛損酀殻悎紻爅浃迡堰ꄀ保嬓⠌ᖧ稷䞋鏹蕄些㛔䍈ⴍ衦觛損酀殉䏔臯䱥壦顁霤碸骋㯝羪踍浖ꇻ簡⢶骽吗㶟咬綧凳廡煛", false), "function(o){for(var r=o.length,e=\"\",n=0,t=0,a=0;a<r;a++)for(var i=o[a],E=BITS_PER_BYTE-1;E>=0;E--){n=(n<<1)+(i>>E&1),++t===BITS_PER_CHAR&&(e+=lookupE[t][n],n=0,t=0)}if(0!==t){for(;!(t in lookupE);)n=1+(n<<1),t++;e+=lookupE[t][n]}return e}"},
	}

	bigEndian := false
	enc := NewBase32768()
	for i, tt := range tests {
		t.Run("test_"+strconv.Itoa(i), func(t *testing.T) {
			got := enc.Encode(([]byte)(tt.want), bigEndian, false)

			if !assert.Equal(t, tt.data, got) {
				return
			}

			data2, err := enc.Decode(got, bigEndian)
			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}
			assert.Equal(t, tt.want, string(data2))
		})
	}

	for i, tt := range tests {
		t.Run("test_str_"+strconv.Itoa(i), func(t *testing.T) {
			data := ([]byte)(tt.want)
			got := enc.EncodeStr(data)

			data2, err := enc.DecodeStr(got)
			if err != nil {
				t.Errorf("DecodeStr() error = %v", err)
				return
			}
			assert.Equal(t, data, data2)
		})
	}
}

type benchmarksItem struct {
	name  string
	n     uint64
	in    []byte
	out   []byte
	outU8 string
	eq    bool
}

func BenchmarkBase32768Encode(b *testing.B) {
	var benchmarks = []*benchmarksItem{
		{name: "16B", n: 16},
		{name: "128B", n: 128},
		{name: "256B", n: 256},
		{name: "1KB", n: 1024},
		{name: "4KB", n: 4 * 1024},
		{name: "10MB", n: 10 * 1024 * 1024},
	}

	n := uint64(65537)
	for _, bb := range benchmarks {
		in := make([]byte, bb.n)
		n = n * bb.n * 23
		for i := range in {
			n = n>>(i%17) + uint64(i)*257
			in[i] = byte(n)
		}
		bb.in = in
	}
	benchBase32768Encode(b, benchmarks)
}

func benchBase32768Encode(b *testing.B, elems []*benchmarksItem) {
	enc := NewBase32768()

	eqf := func(k string, got []byte, elem *benchmarksItem) bool {
		want, err := enc.Decode(got, false)
		if err != nil {
			b.Errorf("Decode %v %v err: %v", k, elem.name, err)
		}
		if bytes.Compare(elem.in, want) != 0 {
			b.Errorf("Decode %v not eq: %v", k, elem.name)
		}

		elem.out = got
		elem.outU8 = string(Utf16ToUtf8(got, false))
		want, err = enc.DecodeStr(elem.outU8)
		if err != nil {
			b.Errorf("DecodeStr %v %v err: %v", k, elem.name, err)
		}
		if bytes.Compare(elem.in, want) != 0 {
			b.Errorf("DecodeStr %v not eq: %v", k, elem.name)
		}
		return true
	}

	for _, elem := range elems {
		got := enc.Encode(elem.in, false, false)
		elem.eq = eqf(b.Name(), got, elem)
	}

	b.ResetTimer()

	b.Run("Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, elem := range elems {
				_ = enc.Encode(elem.in, false, false)
			}
		}
	})

	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, elem := range elems {
				_, _ = enc.Decode(elem.out, false)
			}
		}
	})

	b.Run("DecodeStr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, elem := range elems {
				_, _ = enc.DecodeStr(elem.outU8)
			}
		}
	})
}

func BenchmarkBase32768Decode(b *testing.B) {
	var benchmarks = []*benchmarksItem{
		{name: "16B", n: 16},
		{name: "128B", n: 128},
		{name: "256B", n: 256},
		{name: "1KB", n: 1024},
		{name: "4KB", n: 4 * 1024},
		{name: "10MB", n: 10 * 1024 * 1024},
	}

	n := uint64(65537)
	for _, bb := range benchmarks {
		in := make([]byte, bb.n)
		n = n * bb.n * 23
		for i := range in {
			n = n>>(i%17) + uint64(i)*257
			in[i] = byte(n)
		}
		bb.in = in
	}
	benchBase32768Decode(b, benchmarks)
}

func benchBase32768Decode(b *testing.B, elems []*benchmarksItem) {
	enc := NewBase32768()

	eqf := func(k string, got []byte, elem *benchmarksItem) bool {
		want, err := enc.Decode(got, false)
		if err != nil {
			b.Errorf("Decode %v %v err: %v", k, elem.name, err)
		}
		if bytes.Compare(elem.in, want) != 0 {
			b.Errorf("Decode %v not eq: %v", k, elem.name)
		}

		elem.out = got
		elem.outU8 = string(Utf16ToUtf8(got, false))
		want, err = enc.DecodeStr(elem.outU8)
		if err != nil {
			b.Errorf("DecodeStr %v %v err: %v", k, elem.name, err)
		}
		if bytes.Compare(elem.in, want) != 0 {
			b.Errorf("DecodeStr %v not eq: %v", k, elem.name)
		}
		return true
	}

	for _, elem := range elems {
		got := enc.Encode(elem.in, false, false)
		elem.eq = eqf(b.Name(), got, elem)
	}

	b.ResetTimer()

	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, elem := range elems {
				_, _ = enc.Decode(elem.out, false)
			}
		}
	})

}
