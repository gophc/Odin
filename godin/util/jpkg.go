package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gophc/gophc/gophc-core/zstd"
	"hash/adler32"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

var JsV8Ver = [4]byte{11, 9, 169, 6}

type PkgAble interface {
	Mate() []byte // (pre: {JPK[G|Z|E]}, hash: {4-7})(help: {0-3}, meta len: {4}, data len: {5,6,7})(...)
	Data() []byte
}

func FileList(base string) (map[string]*FileItem, error) {
	m := make(map[string]*FileItem)
	err := filepath.Walk(base, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		var buf []byte
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			buf, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		}
		path = strings.ReplaceAll(path, base, "")
		path = strings.ReplaceAll(path, `\`, "/")

		m[path] = &FileItem{
			Size:    uint32(f.Size()),
			ModTime: uint32(f.ModTime().Unix()),
			Ver:     JsV8Ver,
			path:    path,
			buf:     buf,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return m, nil
}

type FileItem struct {
	path    string
	Size    uint32
	ModTime uint32
	Ver     [4]byte
	buf     []byte
}

func (fi *FileItem) Mate() []byte {
	mate := make([]byte, 30)
	mate[0], mate[1], mate[2], mate[3] = 'J', 'P', 'K', 'Z'
	// mate[4] ... mate[15] auto fill
	*(*uint32)(unsafe.Pointer(&mate[16])) = fi.Size
	*(*uint32)(unsafe.Pointer(&mate[20])) = fi.ModTime
	mate[24], mate[25], mate[26], mate[27] = fi.Ver[0], fi.Ver[1], fi.Ver[2], fi.Ver[3]
	mate[28], mate[29] = 'f', 'e'
	return mate
}

func (fi *FileItem) Data() []byte {
	return fi.buf
}

type PkgItem struct {
	name string
	head []byte
	mate []byte
	data []byte
}

type PkgEncoder struct {
	w         *zstd.Encoder
	r         *zstd.Decoder
	enc       *Base32768
	bigEndian bool
}

func NewPkgEncoder(opts ...zstd.EOption) *PkgEncoder {
	if len(opts) == 0 {
		opts = append(opts, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	}
	// Ordinary compression, for reference.
	w, err := zstd.NewWriter(nil, opts...)
	if err != nil {
		panic(err)
	}
	r, _ := zstd.NewReader(nil)

	pe := &PkgEncoder{
		enc:       NewBase32768(),
		w:         w,
		r:         r,
		bigEndian: false,
	}
	return pe
}

func (pe *PkgEncoder) Item2FileItem(obj *PkgItem) (*FileItem, error) {
	if len(obj.head) != 32 {
		return nil, errors.New("error head [32]byte but len(obj.head): " + convItoa(len(obj.head)))
	}
	head, err := pe.enc.Decode(obj.head, pe.bigEndian)
	if err != nil {
		return nil, err
	}
	if len(head) != 30 || head[0] != 'J' || head[1] != 'P' || head[2] != 'K' {
		return nil, errors.New("mate not start with JPK[X]: " + string(head[:4]))
	}
	if len(head) != 30 || head[28] != 'f' || head[29] != 'e' {
		return nil, errors.New("error mate[28:30] != 'fe' :" + string(head[28:30]))
	}
	hm := uint32(head[4])*16777216 + uint32(head[5])*65536 + uint32(head[6])*256 + uint32(head[7])
	// fake fill head hash
	head[4], head[5], head[6], head[7] = head[0], head[1], head[2], head[3]
	hn, err := PkgAdler32(head, obj.mate, obj.data)
	if err != nil {
		return nil, err
	}
	if hm != hn {
		return nil, errors.New("error adler32 hm: " + convItoa(int(hm)) + " but hn: " + convItoa(int(hn)))
	}

	lm := int(head[12])
	if len(obj.mate) != lm*32 {
		return nil, errors.New("error len(obj.mate): " + convItoa(len(obj.mate)) + " but lm: " + convItoa(lm))
	}
	ld := int(head[13])*65536 + int(head[14])*256 + int(head[15]) // byte(ld/65536), byte(ld/256), byte(ld%256)
	if len(obj.data) <= 0 || len(obj.data) != ld*2 {
		return nil, errors.New("error len(obj.data): " + convItoa(len(obj.data)) + " but ld: " + convItoa(ld))
	}

	mate, err := pe.enc.Decode(obj.mate, pe.bigEndian)
	if err != nil {
		return nil, err
	}
	mate = append(head, mate...)

	data, err := pe.enc.Decode(obj.data, pe.bigEndian)
	if err != nil {
		return nil, err
	}
	var buf []byte
	if mate[3] == 'Z' {
		buf, err = pe.r.DecodeAll(data, nil)
		if err != nil {
			return nil, err
		}
	} else if mate[3] == 'E' {
		if len(data) != 1 || data[0] != '\x00' {
			return nil, errors.New("mate mate[3] is E but len: (0d" + convItoa(int(data[0])) + ") " + convItoa(len(data)))
		}
		buf = nil
	} else if mate[3] == 'G' {
		buf = data
	} else {
		return nil, errors.New("error mate[3]: " + string([]byte{mate[3]}))
	}

	fi := &FileItem{
		path: obj.name,
		buf:  buf,

		Size:    *(*uint32)(unsafe.Pointer(&mate[16])),
		ModTime: *(*uint32)(unsafe.Pointer(&mate[20])),
		Ver:     *(*[4]byte)(unsafe.Pointer(&mate[24])),
	}

	return fi, nil
}

func (pe *PkgEncoder) BuildJPKG(m map[string]PkgAble) ([]*PkgItem, error) {
	mm := make([]*PkgItem, 0, len(m))
	for key, item := range m {
		mate := item.Mate()
		if len(mate) <= 0 {
			mate = []byte{'J', 'P', 'K', 'G'}
		}
		for len(mate)%30 != 0 {
			mate = append(mate, '\x00')
		}
		if mate[0] != 'J' || mate[1] != 'P' || mate[2] != 'K' {
			return nil, errors.New("mate not start with JPK[X] for key: (" + string(mate[:4]) + ") " + key)
		}
		lm := len(mate)/30 - 1
		if lm > 256 {
			return nil, errors.New("mate out max len 256 * 30 for key: (" + convItoa(lm) + " * 30) " + key)
		}

		data := item.Data()
		if len(data) <= 0 {
			mate[3] = 'E'
		}

		var val []byte
		if mate[3] == 'Z' {
			val = pe.w.EncodeAll(data, nil)
		} else if mate[3] == 'E' {
			val = []byte{'\x00'}
		} else if mate[3] == 'G' {
			val = data
		} else {
			mate[3] = 'G'
			val = data
		}

		udata := pe.enc.Encode(val, pe.bigEndian, false)
		ld := len(udata) / 2
		if ld > 16777216 {
			return nil, errors.New("data out max len 16777216 * 2 for key: " + key)
		}

		umate := pe.enc.Encode(mate[30:], pe.bigEndian, false)

		head := mate[:30]
		// TODO zstd dict ??
		head[8], head[9], head[10], head[11] = 'j', 's', '0', '1' // help: {0...4}

		head[12] = byte(lm) //  meta len: {4...5}
		head[13], head[14], head[15] = byte(ld/65536), byte(ld/256), byte(ld%256)

		// fake fill head hash
		head[4], head[5], head[6], head[7] = head[0], head[1], head[2], head[3]

		hn, err := PkgAdler32(head, umate, udata)
		if err != nil {
			return nil, fmt.Errorf("PkgAdler32 err %v for key: %s", err, key)
		}

		head[4], head[5], head[6], head[7] = byte(hn/16777216), byte(hn/65536), byte(hn/256), byte(hn%256)
		uhead := pe.enc.Encode(head, pe.bigEndian, false)

		tmp := &PkgItem{
			name: key,
			head: uhead,
			mate: umate,
			data: udata,
		}

		mm = append(mm, tmp)
	}

	return mm, nil
}

func FileExist(filename string) (bool, bool) {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, false
	}
	return f.IsDir(), true
}

func RemoveFileIfExist(filename string) error {
	isDir, exist := FileExist(filename)
	if isDir {
		return errors.New("path is dir: " + filename)
	}
	if exist {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFileIfExist(filename string) ([]byte, error) {
	isDir, exist := FileExist(filename)
	if isDir {
		return nil, errors.New("path is dir: " + filename)
	}
	if exist {
		buf, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	return nil, os.ErrNotExist
}

func PkgLoad(path string, enc *Base32768) ([]*PkgItem, error) {
	buf, err := ReadFileIfExist(path)
	if err != nil {
		return nil, err
	}
	u8s := Ut2Utf8(buf)
	ida := bytes.IndexByte(u8s, '{')
	idb := bytes.IndexByte(u8s, '}')
	if ida < 0 || idb < 0 {
		return nil, errors.New("not vaild json ida: " + convItoa(ida) + ", idb: " + convItoa(idb))
	}

	u8s = u8s[ida : idb+1]
	ms := make(map[string]string)
	err = json.Unmarshal(u8s, &ms)
	if err != nil {
		return nil, err
	}

	mm := make([]*PkgItem, 0, len(ms))
	for key, val := range ms {
		if len(val) >= 49 {
			buf = Utf16FromStr(val[:49], false)
		} else {
			buf = Utf16FromStr(val, false)
		}

		if len(buf) < 32 {
			return nil, errors.New("utf16 chars lt 16 for key: " + key)
		}
		head, err := enc.Decode(buf[:32], false)
		if err != nil {
			return nil, err
		}

		if len(head) != 30 || head[0] != 'J' || head[1] != 'P' || head[2] != 'K' {
			return nil, errors.New("mate not start with JPK[X]: " + string(head[:4]))
		}

		hm := uint32(head[4])*16777216 + uint32(head[5])*65536 + uint32(head[6])*256 + uint32(head[7])

		buf = Utf16FromStr(val, false)
		im := (int(head[12]) + 1) * 32
		uhead, mate, data := buf[0:32], buf[32:im], buf[im:]
		(*_uSliceH)(unsafe.Pointer(&uhead)).Cap = len(uhead)
		(*_uSliceH)(unsafe.Pointer(&mate)).Cap = len(mate)
		(*_uSliceH)(unsafe.Pointer(&data)).Cap = len(data)

		ld := int(head[13])*65536 + int(head[14])*256 + int(head[15]) // byte(ld/65536), byte(ld/256), byte(ld%256)
		if len(data) <= 0 || len(data) != ld*2 {
			return nil, errors.New("error len(obj.data): " + convItoa(len(data)) + " but ld: " + convItoa(ld))
		}

		// fake fill head hash
		head[4], head[5], head[6], head[7] = head[0], head[1], head[2], head[3]
		hn, err := PkgAdler32(head, mate, data)
		if err != nil {
			return nil, err
		}
		if hm != hn {
			return nil, errors.New("error adler32 hm: " + convItoa(int(hm)) + " but hn: " + convItoa(int(hn)))
		}

		obj := &PkgItem{
			name: key,
			head: uhead,
			mate: mate,
			data: data,
		}

		mm = append(mm, obj)
	}
	return mm, nil
}

func PkgDumpUtf8(name string, path string, mm []*PkgItem) error {
	err := RemoveFileIfExist(path)
	if err != nil {
		return err
	}

	name = PkgFixedName(name)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666) //创建文件
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	buf := ([]byte)("var " + name + " = {")
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	first := true
	quote := ([]byte)(`"`)
	bufs := [5][]byte{nil, nil, nil, nil, quote}
	for _, item := range mm {
		if first {
			buf = ([]byte)("\n" + `  "` + item.name + `": "`)
			first = false
		} else {
			buf = ([]byte)(",\n" + `  "` + item.name + `": "`)
		}
		bufs[0], bufs[1], bufs[2], bufs[3] = buf, Utf16ToUtf8(item.head, false),
			Utf16ToUtf8(item.mate, false), Utf16ToUtf8(item.data, false)

		for _, data := range bufs {
			if len(data) > 0 {
				_, err = f.Write(data)
				if err != nil {
					return err
				}
			}
		}
	}

	buf = ([]byte)("\n};\n")
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}

func PkgDumpU16LE(name string, path string, mm []*PkgItem) error {
	err := RemoveFileIfExist(path)
	if err != nil {
		return err
	}

	name = PkgFixedName(name)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666) //创建文件
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	buf := []byte{'\xff', '\xfe'}
	buf = append(buf, Utf16FromStr("var "+name+" = {", false)...)
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	first := true
	quote := Utf16FromStr(`"`, false)
	bufs := [5][]byte{nil, nil, nil, nil, quote}
	for _, item := range mm {
		if first {
			buf = Utf16FromStr("\n"+`  "`+item.name+`": "`, false)
			first = false
		} else {
			buf = Utf16FromStr(",\n"+`  "`+item.name+`": "`, false)
		}
		bufs[0], bufs[1], bufs[2], bufs[3] = buf, item.head, item.mate, item.data
		for _, data := range bufs {
			if len(data) > 0 {
				_, err = f.Write(data)
				if err != nil {
					return err
				}
			}
		}
	}

	buf = Utf16FromStr("\n};\n", false)
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}
	return nil
}

func PkgFixedName(name string) string {
	for _, ru := range ".{}[]()-" {
		name = strings.ReplaceAll(name, string([]rune{ru}), "_")
	}
	for i := strings.Index(name, "//"); i >= 0; {
		name = strings.ReplaceAll(name, "//", "/")
	}
	for i := strings.Index(name, `\\`); i >= 0; {
		name = strings.ReplaceAll(name, `\\`, "/")
	}
	name = strings.ReplaceAll(name, "/", "$")
	return name
}

func PkgAdler32(head, umate, udata []byte) (uint32, error) {
	if len(head) != 30 {
		return 0, errors.New("PkgAdler32 len(head) must 30")
	}
	if len(udata) <= 0 {
		return 0, errors.New("PkgAdler32 len(udata) must gt 0")
	}
	if len(udata)%2 != 0 {
		return 0, errors.New("PkgAdler32 len(udata) must even")
	}
	if len(umate)%32 != 0 {
		return 0, errors.New("PkgAdler32 len(umate) must mod 32 eq 0")
	}

	h := adler32.New()
	_, _ = h.Write(head)

	if len(umate) > 0 {
		_, _ = h.Write(umate)
	}
	_, _ = h.Write(udata)
	return h.Sum32(), nil
}
