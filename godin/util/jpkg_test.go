package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func testGetFileListPkg(t *testing.T, pe *PkgEncoder, path string) (map[string]*FileItem, []*PkgItem) {
	m, err := FileList(path)
	if err != nil {
		t.Errorf("FileList() err: %v", err)
		return nil, nil
	}

	mi := make(map[string]PkgAble)
	for s, item := range m {
		mi[s] = item
	}

	got, err := pe.BuildJPKG(mi)
	if err != nil {
		t.Errorf("BuildJPKG() err: %v", err)
		return nil, nil
	}
	if len(got) != len(m) {
		t.Errorf("BuildJPKG() err len: %v %v", len(got), len(m))
		return nil, nil
	}
	return m, got
}

func testEqFileListPkg(t *testing.T, pe *PkgEncoder, m map[string]*FileItem, got []*PkgItem) {
	for _, data := range got {
		fi, err := pe.Item2FileItem(data)
		if err != nil {
			t.Errorf("Item2FileItem() err: %v", err)
			return
		}
		ff, ok := m[data.name]
		if !ok {
			t.Errorf("FileItem not found key: %v", data.name)
			return
		}
		assert.Equal(t, ff, fi)
	}
}

func TestBuildJPKG(t *testing.T) {
	if SkipTestOnlyDbg(t, os.Args[1:]) {
		return
	}

	name, pe := "../parser", NewPkgEncoder()
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, name)

	m, got := testGetFileListPkg(t, pe, path)

	testEqFileListPkg(t, pe, m, got)
}

func TestPkgDumpUtf8(t *testing.T) {
	if SkipTestOnlyDbg(t, os.Args[1:]) {
		return
	}

	name, pe := "../parser", NewPkgEncoder()
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, name)

	m, got := testGetFileListPkg(t, pe, path)

	testEqFileListPkg(t, pe, m, got)

	path = filepath.Join(cwd, "../tools/tmp/parser_u8.js")
	err := PkgDumpUtf8(name, path, got)
	if err != nil {
		t.Errorf("PkgDumpUtf8() err: %v", err)
		return
	}

	got2, err := PkgLoad(path, pe.enc)
	if err != nil {
		t.Errorf("PkgLoad() err: %v", err)
		return
	}
	testEqFileListPkg(t, pe, m, got2)

	for _, item := range got {
		for _, item2 := range got2 {
			if item.name == item2.name {
				assert.Equal(t, item, item2)
			}
		}
	}
}

func TestPkgDumpU16LE(t *testing.T) {
	if SkipTestOnlyDbg(t, os.Args[1:]) {
		return
	}

	name, pe := "../parser", NewPkgEncoder()
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, name)

	m, got := testGetFileListPkg(t, pe, path)

	testEqFileListPkg(t, pe, m, got)

	path = filepath.Join(cwd, "../tools/tmp/parser.js")
	err := PkgDumpU16LE(name, path, got)
	if err != nil {
		t.Errorf("PkgDumpU16LE() err: %v", err)
		return
	}

	got2, err := PkgLoad(path, pe.enc)
	if err != nil {
		t.Errorf("PkgLoad() err: %v", err)
		return
	}
	testEqFileListPkg(t, pe, m, got2)

	for _, item := range got {
		for _, item2 := range got2 {
			if item.name == item2.name {
				assert.Equal(t, item, item2)
			}
		}
	}
}
