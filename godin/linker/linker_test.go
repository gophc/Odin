package linker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLinkerData(t *testing.T) {
	ld := NewLinkerData()
	assert.NotNil(t, ld)
	assert.NotNil(t, ld.ForeignLibrariesSet)
	assert.NotNil(t, ld.ForeignLibraries)
	assert.NotNil(t, ld.OutputObjectPaths)
	assert.NotNil(t, ld.OutputTempPaths)
	assert.False(t, ld.NeedsSystemLibraryLinked)
}

func TestLinkerData_EnableSystemLibraryLinking(t *testing.T) {
	ld := NewLinkerData()
	assert.False(t, ld.NeedsSystemLibraryLinked)
	ld.EnableSystemLibraryLinking()
	assert.True(t, ld.NeedsSystemLibraryLinked)
}

func TestLinkerData_AddForeignLibrary(t *testing.T) {
	ld := NewLinkerData()
	lib1 := &LibraryNameEntity{
		Paths:            []string{"lib1.a"},
		ExtraLinkerFlags: "-lstdc++",
		IgnoreDuplicates: false,
	}
	lib2 := &LibraryNameEntity{
		Paths:            []string{"lib2.a"},
		ExtraLinkerFlags: "-lm",
		IgnoreDuplicates: true,
	}

	ld.AddForeignLibrary(lib1)
	assert.Len(t, ld.ForeignLibraries, 1)
	assert.Len(t, ld.ForeignLibrariesSet, 1)

	// Adding duplicate should not increase count
	ld.AddForeignLibrary(lib1)
	assert.Len(t, ld.ForeignLibraries, 1)
	assert.Len(t, ld.ForeignLibrariesSet, 1)

	ld.AddForeignLibrary(lib2)
	assert.Len(t, ld.ForeignLibraries, 2)
	assert.Len(t, ld.ForeignLibrariesSet, 2)
}

func TestLinkerData_AddOutputObjectPath(t *testing.T) {
	ld := NewLinkerData()
	ld.AddOutputObjectPath("obj1.o")
	ld.AddOutputObjectPath("obj2.o")
	assert.Equal(t, []string{"obj1.o", "obj2.o"}, ld.OutputObjectPaths)
}

func TestLinkerData_AddOutputTempPath(t *testing.T) {
	ld := NewLinkerData()
	ld.AddOutputTempPath("temp1")
	ld.AddOutputTempPath("temp2")
	assert.Equal(t, []string{"temp1", "temp2"}, ld.OutputTempPaths)
}

func TestLinkerData_Init(t *testing.T) {
	ld := NewLinkerData()
	info := &CheckerInfo{
		InitScope: &Scope{
			Pkg: &Package{Name: "testpkg"},
		},
	}

	// Test with empty OutFilepath
	buildCtx := &BuildContext{OutFilepath: ""}
	ld.Init(info, "/path/to/main.odin", buildCtx)
	assert.Contains(t, ld.OutputName, "main")
	assert.NotEmpty(t, ld.OutputBase)

	// Test with custom OutFilepath
	buildCtx.OutFilepath = "output.exe"
	ld.Init(info, "/path/to/main.odin", buildCtx)
	assert.Equal(t, "output.exe", ld.OutputName)
	assert.Contains(t, ld.OutputBase, "output")
}

func TestStringSet(t *testing.T) {
	s := NewStringSet(0)
	assert.True(t, s.Add("a"))
	assert.False(t, s.Add("a")) // duplicate
	assert.True(t, s.Contains("a"))
	assert.False(t, s.Contains("b"))
	s.Remove("a")
	assert.False(t, s.Contains("a"))
}

func TestArray(t *testing.T) {
	a := NewArray[int](0)
	a.Add(1)
	a.Add(2)
	assert.Equal(t, 2, a.Len())
	assert.Equal(t, 1, a.Get(0))
	assert.Equal(t, 2, a.Get(1))
	assert.Equal(t, []int{1, 2}, a.Slice())
}

func TestHelperFunctions(t *testing.T) {
	assert.True(t, stringEndsWith("file.o", ".o"))
	assert.False(t, stringEndsWith("file.c", ".o"))
	assert.Equal(t, "hello", stringToLower("HELLO"))
	assert.True(t, stringContainsString("hello world", "world"))
	assert.True(t, hasASMExtension("file.s"))
	assert.True(t, hasASMExtension("file.asm"))
	assert.False(t, hasASMExtension("file.c"))
	assert.Equal(t, "file", filenameWithoutDirectory("/path/to/file"))
	assert.Equal(t, "\"spaced path\"", quotePath("spaced path"))
	assert.Equal(t, "noquote", quotePath("noquote"))
	assert.Equal(t, "ab", concatenate4Strings("a", "b", "", ""))
	assert.Equal(t, "ab", concatenateStrings("a", "b"))
	assert.Equal(t, "a/b", normalizePath("a//b"))
}

func TestGBString(t *testing.T) {
	s := gbStringMake()
	s = gbStringAppendc(s, "hello")
	s = gbStringAppendFmt(s, " world %d", 42)
	assert.Equal(t, "hello world 42", s.String())

	s2 := gbStringMakeReserve(100)
	s2.WriteString("test")
	assert.Equal(t, "test", s2.String())

	gbStringClear(s2)
	assert.Equal(t, "", s2.String())
}
