package docs

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Format / binary structure tests
// ============================================================================

func TestOdinDocHeaderBinarySize(t *testing.T) {
	// Verify struct sizes match expected binary layout
	t.Run("OdinDocHeaderBase", func(t *testing.T) {
		// Magic(8) + Padding0(4) + Version(4) + TotalSize(4) + HeaderSize(4) + Hash(4) = 28
		// Version is: major(1) + minor(1) + patch(1) + pad0(1) = 4
		assert.Equal(t, 28, binary.Size(OdinDocHeaderBase{}))
	})

	t.Run("OdinDocFile", func(t *testing.T) {
		// Pkg(4) + Name(8) = 12
		assert.Equal(t, 12, binary.Size(OdinDocFile{}))
	})

	t.Run("OdinDocPosition", func(t *testing.T) {
		// File(4) + Line(4) + Column(4) + Offset(4) = 16
		assert.Equal(t, 16, binary.Size(OdinDocPosition{}))
	})

	t.Run("OdinDocEntity", func(t *testing.T) {
		size := binary.Size(OdinDocEntity{})
		t.Logf("OdinDocEntity size: %d", size)
		assert.Greater(t, size, 0)
	})

	t.Run("OdinDocType", func(t *testing.T) {
		size := binary.Size(OdinDocType{})
		t.Logf("OdinDocType size: %d", size)
		assert.Greater(t, size, 0)
	})

	t.Run("OdinDocPkg", func(t *testing.T) {
		size := binary.Size(OdinDocPkg{})
		t.Logf("OdinDocPkg size: %d", size)
		assert.Greater(t, size, 0)
	})

	t.Run("OdinDocHeader", func(t *testing.T) {
		size := binary.Size(OdinDocHeader{})
		t.Logf("OdinDocHeader size: %d", size)
		assert.Greater(t, size, 0)
	})
}

func TestAlignOffset(t *testing.T) {
	tests := []struct {
		offset    int
		alignment int
		expected  int
	}{
		{0, 4, 0},
		{1, 4, 4},
		{3, 4, 4},
		{4, 4, 4},
		{5, 4, 8},
		{0, 8, 0},
		{1, 8, 8},
		{7, 8, 8},
		{8, 8, 8},
		{9, 8, 16},
	}

	for _, tt := range tests {
		result := alignOffset(tt.offset, tt.alignment)
		assert.Equal(t, tt.expected, result,
			"alignOffset(%d, %d) = %d, want %d", tt.offset, tt.alignment, result, tt.expected)
	}
}

func TestCalcFNV1a(t *testing.T) {
	// FNV-1a test vectors
	tests := []struct {
		input    string
		expected uint32
	}{
		{"", 0x811c9dc5},
		{"a", 0xe40c292c},
		{"ab", 0x4d2505ca},
		{"abc", 0x1a47e90b},
	}

	for _, tt := range tests {
		result := calcFNV1a([]byte(tt.input))
		assert.Equal(t, tt.expected, result, "FNV1a(%q)", tt.input)
	}
}

// ============================================================================
// Writer tests
// ============================================================================

func TestWriterPrepare(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()

	assert.Equal(t, WriterStatePreparing, w.state)
	assert.NotNil(t, w.stringCache)
	assert.NotNil(t, w.entityCache)
	assert.NotNil(t, w.typeCache)
	assert.NotNil(t, w.pkgCache)
	assert.NotNil(t, w.fileCache)

	// Initial capacities
	assert.Equal(t, 1, w.files.cap)
	assert.Equal(t, 1, w.pkgs.cap)
	assert.Equal(t, 1, w.entities.cap)
	assert.Equal(t, 1, w.types.cap)
	assert.Equal(t, 16, w.strings_.cap)
	assert.Equal(t, 16, w.blob.cap)
}

func TestWriterCalcTotalSize(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()

	totalSize := w.calcTotalSize()
	assert.Greater(t, totalSize, 0)
	t.Logf("Total buffer size: %d", totalSize)
}

func TestWriterStartWriting(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()
	w.startWriting()

	assert.Equal(t, WriterStateWriting, w.state)
	assert.NotNil(t, w.Data)
	assert.Greater(t, len(w.Data), 0)
	assert.NotNil(t, w.Header)

	// Check magic bytes
	assert.Equal(t, MagicBytes[:], w.Header.Base.Magic[:])
}

func TestWriterStringCache(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()
	w.startWriting()

	s1 := w.writeString("hello")
	s2 := w.writeString("hello")

	// Same string should return same offset/length (cached)
	assert.Equal(t, s1, s2)
	assert.Equal(t, uint32(5), s1.Length)

	// Verify string content in buffer
	str := extractString(w.Data, s1)
	assert.Equal(t, "hello", str)
}

func TestWriterStringWithoutCache(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()
	w.startWriting()

	s1 := w.writeStringWithoutCache("hello")
	s2 := w.writeStringWithoutCache("hello")

	// Without cache, different strings will be at different offsets
	assert.NotEqual(t, s1.Offset, s2.Offset,
		"writeStringWithoutCache should not deduplicate")
}

func TestWriterFileItems(t *testing.T) {
	w := NewWriter(nil)

	// Preparing phase
	w.prepare()
	idx, dst := w.writeFile(&OdinDocFile{Name: OdinDocString{Offset: 100, Length: 10}})
	assert.Equal(t, uint32(0), idx)
	assert.Nil(t, dst)
	assert.Equal(t, 2, w.files.cap) // started at 1, incremented

	// Writing phase
	w.startWriting()
	idx, dst = w.writeFile(&OdinDocFile{Name: OdinDocString{Offset: 100, Length: 10}})
	assert.Greater(t, idx, uint32(0))
	assert.Nil(t, dst) // writeTrackedItem returns nil for dst currently
}

// ============================================================================
// LoadDocFile tests
// ============================================================================

func TestWriteAndLoadRoundtrip(t *testing.T) {
	w := NewWriter(nil)

	// Prepare phase: register strings (sizes the string buffer)
	w.prepare()
	w.writeString("test_package")
	w.writeString("test_file.odin")

	// Writing phase: actually write the strings
	w.startWriting()
	w.writeString("test_package")
	w.writeString("test_file.odin")
	w.endWriting()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.odin-doc")

	err := w.WriteToFile(tmpFile)
	require.NoError(t, err)

	// Load it back
	header, data, err := LoadDocFile(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, header)
	require.NotNil(t, data)

	// Verify header fields
	assert.Equal(t, MagicBytes[:], header.Base.Magic[:])
	assert.Equal(t, OdinDocVersion, header.Base.Version)
	assert.Greater(t, header.Base.TotalSize, uint32(0))
}

func TestLoadDocFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent file
	_, _, err := LoadDocFile(filepath.Join(tmpDir, "nonexistent.odin-doc"))
	assert.Error(t, err)

	// Invalid magic
	tmpFile := filepath.Join(tmpDir, "bad.odin-doc")
	os.WriteFile(tmpFile, make([]byte, 256), 0644)
	_, _, err = LoadDocFile(tmpFile)
	assert.Error(t, err)
}

// ============================================================================
// Entity comparison tests
// ============================================================================

func TestCmpEntitiesForPrinting(t *testing.T) {
	pkgA := &AstPackage{Name: MakeString("alpha")}
	pkgB := &AstPackage{Name: MakeString("beta")}

	e1 := &Entity{
		Pkg:   pkgA,
		Kind:  EntityVariable,
		Token: Token{String: MakeString("Var1")},
	}
	e2 := &Entity{
		Pkg:   pkgA,
		Kind:  EntityConstant,
		Token: Token{String: MakeString("Const1")},
	}
	e3 := &Entity{
		Pkg:   pkgB,
		Kind:  EntityVariable,
		Token: Token{String: MakeString("Var2")},
	}

	// Same package, different kinds: Constant (0) < Variable (1)
	assert.Less(t, CmpEntitiesForPrinting(e2, e1), 0)

	// Different packages: alpha < beta
	assert.Less(t, CmpEntitiesForPrinting(e1, e3), 0)

	// Nil package handling
	e4 := &Entity{Pkg: nil, Kind: EntityVariable, Token: Token{String: MakeString("X")}}
	assert.Less(t, CmpEntitiesForPrinting(e4, e1), 0) // nil pkg comes first
}

func TestCmpEntitiesForPrintingByOrderInSrc(t *testing.T) {
	e1 := &Entity{
		Pkg:        &AstPackage{Name: MakeString("pkg")},
		OrderInSrc: 100,
		Token:      Token{Pos: TokenPos{Offset: 50}},
	}
	e2 := &Entity{
		Pkg:        &AstPackage{Name: MakeString("pkg")},
		OrderInSrc: 200,
		Token:      Token{Pos: TokenPos{Offset: 10}},
	}
	e3 := &Entity{
		Pkg:        &AstPackage{Name: MakeString("pkg")},
		OrderInSrc: 100,
		Token:      Token{Pos: TokenPos{Offset: 100}},
	}

	// Different order_in_src
	assert.Less(t, CmpEntitiesForPrintingByOrderInSrc(e1, e2), 0)

	// Same order_in_src, compare by offset
	assert.Less(t, CmpEntitiesForPrintingByOrderInSrc(e1, e3), 0)
}

func TestCmpAstPackageByName(t *testing.T) {
	pkgA := &AstPackage{Name: MakeString("alpha")}
	pkgB := &AstPackage{Name: MakeString("beta")}
	pkgA2 := &AstPackage{Name: MakeString("alpha")}

	assert.Less(t, CmpAstPackageByName(pkgA, pkgB), 0)
	assert.Equal(t, 0, CmpAstPackageByName(pkgA, pkgA2))
	assert.Greater(t, CmpAstPackageByName(pkgB, pkgA), 0)
}

// ============================================================================
// Entity sorting tests
// ============================================================================

func TestSortEntitiesByKind(t *testing.T) {
	pkg := &AstPackage{Name: MakeString("test")}

	entities := []*Entity{
		{Pkg: pkg, Kind: EntityProcedure, Token: Token{String: MakeString("DoSomething")}},
		{Pkg: pkg, Kind: EntityTypeName, Token: Token{String: MakeString("MyType")}},
		{Pkg: pkg, Kind: EntityConstant, Token: Token{String: MakeString("MAX_SIZE")}},
		{Pkg: pkg, Kind: EntityVariable, Token: Token{String: MakeString("counter")}},
	}

	SortEntitiesByKind(entities)

	// Expected order: Constant(0), Variable(1), Procedure(2), TypeName(4)
	assert.Equal(t, EntityConstant, entities[0].Kind)
	assert.Equal(t, EntityVariable, entities[1].Kind)
	assert.Equal(t, EntityProcedure, entities[2].Kind)
	assert.Equal(t, EntityTypeName, entities[3].Kind)
}

func TestSortPackagesByName(t *testing.T) {
	pkgs := []*AstPackage{
		{Name: MakeString("zulu")},
		{Name: MakeString("alpha")},
		{Name: MakeString("mike")},
	}

	SortPackagesByName(pkgs)

	assert.Equal(t, "alpha", pkgs[0].Name.String())
	assert.Equal(t, "mike", pkgs[1].Name.String())
	assert.Equal(t, "zulu", pkgs[2].Name.String())
}

// ============================================================================
// String type tests
// ============================================================================

func TestMakeString(t *testing.T) {
	s := MakeString("hello")
	assert.Equal(t, "hello", s.String())
	assert.Equal(t, 5, s.Len())
}

func TestStringLen(t *testing.T) {
	s := String{Text: []byte("test")}
	assert.Equal(t, 4, s.Len())

	empty := String{}
	assert.Equal(t, 0, empty.Len())
}

// ============================================================================
// Type helpers tests
// ============================================================================

func TestIsTypeUntyped(t *testing.T) {
	untypedTests := []string{
		"untyped integer", "untyped float", "untyped complex",
		"untyped string", "untyped bool", "untyped nil", "untyped rune",
	}

	for _, name := range untypedTests {
		typ := &Type{
			Kind:  TypeBasic,
			Basic: TypBasic{Name: MakeString(name)},
		}
		assert.True(t, isTypeUntyped(typ), "%s should be untyped", name)
	}

	typed := &Type{
		Kind:  TypeBasic,
		Basic: TypBasic{Name: MakeString("int")},
	}
	assert.False(t, isTypeUntyped(typed))

	assert.False(t, isTypeUntyped(nil))

	nonBasic := &Type{Kind: TypePointer}
	assert.False(t, isTypeUntyped(nonBasic))
}

func TestBaseType(t *testing.T) {
	inner := &Type{Kind: TypeBasic, Basic: TypBasic{Name: MakeString("int")}}
	named := &Type{
		Kind:  TypeNamed,
		Named: TypNamed{Name: MakeString("MyInt"), Base: inner},
	}

	assert.Equal(t, inner, baseType(named))
	assert.Equal(t, inner, baseType(inner)) // non-named returns self
	assert.Nil(t, baseType(nil))
}

func TestHashTypeCanonical(t *testing.T) {
	typ := &Type{Kind: TypeBasic, Basic: TypBasic{Name: MakeString("int")}}
	h := hashTypeCanonical(typ)
	assert.NotZero(t, h)

	assert.Zero(t, hashTypeCanonical(nil))
}

// ============================================================================
// Doc printer / text output tests
// ============================================================================

func TestDocPrinterPrintDocs(t *testing.T) {
	var buf bytes.Buffer
	printer := NewDocPrinterTo(&buf, 0)

	pkg := &AstPackage{
		Name:     MakeString("test_pkg"),
		Fullpath: MakeString("/path/to/test_pkg"),
		Kind:     PackageNormal,
		Scope: &Scope{
			Elements: []ScopeElement{
				{
					Name: MakeString("ExportedVar"),
					Value: &Entity{
						Kind:  EntityVariable,
						Pkg:   nil, // Will be set below
						Token: Token{String: MakeString("ExportedVar")},
						Variable: EntVariableData{
							IsExport: true,
						},
						DeclInfo: &DeclInfo{
							TypeExpr: &Ast{Kind: AstIdent},
						},
					},
				},
			},
		},
		Files: []*AstFile{
			{Fullpath: MakeString("/path/to/test_pkg/test.odin")},
		},
	}

	// Fix up entity pkg reference
	pkg.Scope.Elements[0].Value.Pkg = pkg
	pkg.Scope.Elements[0].Value.File = pkg.Files[0]

	printer.PrintPackage(pkg)

	output := buf.String()
	assert.Contains(t, output, "package test_pkg")
}

func TestDocPrinterPrintCommentGroup(t *testing.T) {
	t.Run("nil comment group", func(t *testing.T) {
		var localBuf bytes.Buffer
		p := NewDocPrinterTo(&localBuf, 0)
		assert.False(t, p.PrintCommentGroup(1, nil))
	})

	t.Run("slash slash comment", func(t *testing.T) {
		var localBuf bytes.Buffer
		p := NewDocPrinterTo(&localBuf, 0)
		g := &CommentGroup{
			List: []Comment{
				{String: MakeString("// This is a comment")},
			},
		}
		assert.True(t, p.PrintCommentGroup(1, g))
		output := localBuf.String()
		assert.Contains(t, output, "This is a comment")
	})

	t.Run("slash slash with special prefix", func(t *testing.T) {
		var localBuf bytes.Buffer
		p := NewDocPrinterTo(&localBuf, 0)
		g := &CommentGroup{
			List: []Comment{
				{String: MakeString("// +private")},
			},
		}
		assert.False(t, p.PrintCommentGroup(1, g))
	})

	t.Run("block comment", func(t *testing.T) {
		var localBuf bytes.Buffer
		p := NewDocPrinterTo(&localBuf, 0)
		g := &CommentGroup{
			List: []Comment{
				{String: MakeString("/* This is a\n   block comment */")},
			},
		}
		assert.True(t, p.PrintCommentGroup(1, g))
		output := localBuf.String()
		assert.Contains(t, output, "block comment")
	})
}

// ============================================================================
// IsEntityExported tests
// ============================================================================

func TestIsEntityExported(t *testing.T) {
	t.Run("nil entity", func(t *testing.T) {
		assert.False(t, IsEntityExported(nil, false))
	})

	t.Run("builtin entity", func(t *testing.T) {
		e := &Entity{Kind: EntityBuiltin}
		assert.True(t, IsEntityExported(e, true))
		assert.False(t, IsEntityExported(e, false))
	})

	t.Run("exported variable", func(t *testing.T) {
		e := &Entity{
			Kind:     EntityVariable,
			Variable: EntVariableData{IsExport: true},
		}
		assert.True(t, IsEntityExported(e, false))
	})

	t.Run("exported procedure", func(t *testing.T) {
		e := &Entity{
			Kind:      EntityProcedure,
			Procedure: EntProcedureData{IsExport: true},
		}
		assert.True(t, IsEntityExported(e, false))
	})

	t.Run("uppercase name", func(t *testing.T) {
		e := &Entity{
			Kind:  EntityVariable,
			Token: Token{String: MakeString("Exported")},
		}
		assert.True(t, IsEntityExported(e, false))
	})

	t.Run("lowercase name", func(t *testing.T) {
		e := &Entity{
			Kind:  EntityVariable,
			Token: Token{String: MakeString("private")},
		}
		assert.False(t, IsEntityExported(e, false))
	})
}

// ============================================================================
// Utility tests
// ============================================================================

func TestRemoveDirectoryFromPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/file.odin", "file.odin"},
		{"file.odin", "file.odin"},
		{"C:\\Users\\test\\file.odin", "file.odin"},
		{"relative/path/file.odin", "file.odin"},
		{"file", "file"},
		{"", ""},
	}

	for _, tt := range tests {
		result := removeDirectoryFromPath(tt.input)
		assert.Equal(t, tt.expected, result, "removeDirectoryFromPath(%q)", tt.input)
	}
}

func TestRemoveExtension(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.odin", "test"},
		{"test.odin-doc", "test"},
		{"/path/to/test.odin", "/path/to/test"},
		{"noext", "noext"},
		{"", ""},
	}

	for _, tt := range tests {
		result := removeExtension(tt.input)
		assert.Equal(t, tt.expected, result, "removeExtension(%q)", tt.input)
	}
}

func TestStringExtensionPosition(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"test.odin", 4},
		{"noext", -1},
		{"/path/to/file.odin", 13},
		{"", -1},
	}

	for _, tt := range tests {
		result := stringExtensionPosition(tt.input)
		assert.Equal(t, tt.expected, result, "stringExtensionPosition(%q)", tt.input)
	}
}

// ============================================================================
// IsInDocWriter tests
// ============================================================================

func TestIsInDocWriter(t *testing.T) {
	// Initially false
	assert.False(t, IsInDocWriter())
}

// ============================================================================
// Comment group string writer tests
// ============================================================================

func TestWriteCommentGroupString(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()

	// Pre-flight during prepare phase to size the string buffer
	t.Run("single slash-slash comment", func(t *testing.T) {
		g := &CommentGroup{
			List: []Comment{
				{String: MakeString("// A simple comment")},
			},
		}
		w.writeCommentGroupString(g)
	})

	w.startWriting()

	t.Run("nil comment group", func(t *testing.T) {
		result := w.writeCommentGroupString(nil)
		assert.Equal(t, OdinDocString{}, result)
	})

	t.Run("empty comment group", func(t *testing.T) {
		result := w.writeCommentGroupString(&CommentGroup{})
		assert.Equal(t, OdinDocString{}, result)
	})

	t.Run("single slash-slash comment", func(t *testing.T) {
		g := &CommentGroup{
			List: []Comment{
				{String: MakeString("// A simple comment")},
			},
		}
		result := w.writeCommentGroupString(g)
		str := extractString(w.Data, result)
		assert.Contains(t, str, "A simple comment")
	})
}

// ============================================================================
// Token position casting tests
// ============================================================================

func TestTokenPosCast(t *testing.T) {
	w := NewWriter(nil)

	pos := TokenPos{
		FileID: 0,
		Line:   42,
		Column: 10,
		Offset: 150,
	}

	docPos := w.tokenPosCast(pos)

	assert.Equal(t, uint32(0), docPos.File)
	assert.Equal(t, uint32(42), docPos.Line)
	assert.Equal(t, uint32(10), docPos.Column)
	assert.Equal(t, uint32(150), docPos.Offset)
}

// ============================================================================
// Entity kind ordering tests
// ============================================================================

func TestPrintEntityKindOrdering(t *testing.T) {
	// Verify ordering is consistent with C++ source
	assert.Equal(t, -1, printEntityKindOrdering[EntityInvalid])
	assert.Equal(t, 0, printEntityKindOrdering[EntityConstant])
	assert.Equal(t, 1, printEntityKindOrdering[EntityVariable])
	assert.Equal(t, 4, printEntityKindOrdering[EntityTypeName])
	assert.Equal(t, 2, printEntityKindOrdering[EntityProcedure])
	assert.Equal(t, 3, printEntityKindOrdering[EntityProcGroup])
}

func TestPrintEntityNames(t *testing.T) {
	assert.Equal(t, "constants", printEntityNames[EntityConstant])
	assert.Equal(t, "variables", printEntityNames[EntityVariable])
	assert.Equal(t, "types", printEntityNames[EntityTypeName])
	assert.Equal(t, "procedures", printEntityNames[EntityProcedure])
	assert.Equal(t, "proc_group", printEntityNames[EntityProcGroup])
	assert.Equal(t, "import names", printEntityNames[EntityImportName])
	assert.Equal(t, "library names", printEntityNames[EntityLibraryName])
}

// ============================================================================
// Format type constants tests
// ============================================================================

func TestOdinDocVersionConstants(t *testing.T) {
	assert.Equal(t, uint8(0), OdinDocVersionMajor)
	assert.Equal(t, uint8(3), OdinDocVersionMinor)
	assert.Equal(t, uint8(2), OdinDocVersionPatch)

	assert.Equal(t, OdinDocVersionMajor, OdinDocVersion.Major)
	assert.Equal(t, OdinDocVersionMinor, OdinDocVersion.Minor)
	assert.Equal(t, OdinDocVersionPatch, OdinDocVersion.Patch)
}

func TestMagicBytes(t *testing.T) {
	expected := "odindoc"
	assert.Equal(t, expected, string(MagicBytes[:7]))
	assert.Equal(t, byte(0), MagicBytes[7])
}

func TestOdinDocTypeKindStrings(t *testing.T) {
	assert.NotEmpty(t, OdinDocTypeKindStrings[OdinDocTypeBasic])
	assert.NotEmpty(t, OdinDocTypeKindStrings[OdinDocTypeStruct])
	assert.NotEmpty(t, OdinDocTypeKindStrings[OdinDocTypeProc])
}

func TestOdinDocEntityKindStrings(t *testing.T) {
	assert.NotEmpty(t, OdinDocEntityKindStrings[OdinDocEntityConstant])
	assert.NotEmpty(t, OdinDocEntityKindStrings[OdinDocEntityVariable])
	assert.NotEmpty(t, OdinDocEntityKindStrings[OdinDocEntityProcedure])
}

// ============================================================================
// Writer slice/blob tests
// ============================================================================

func TestWriteBytes(t *testing.T) {
	w := NewWriter(nil)

	// Preparing phase
	w.prepare()
	result := w.writeBytes([]byte("test"))
	assert.Equal(t, OdinDocArray{}, result)
	assert.Greater(t, w.blob.cap, 16) // grew beyond initial 16

	// Writing phase
	w.startWriting()
	result = w.writeBytes([]byte("test"))
	assert.NotEqual(t, OdinDocArray{}, result)
	assert.Equal(t, uint32(4), result.Length)

	// Verify data in buffer
	data := w.Data[result.Offset : result.Offset+result.Length]
	assert.Equal(t, []byte("test"), data)
}

func TestWriteUint32Slice(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()
	w.startWriting()

	result := w.writeUint32Slice([]uint32{1, 2, 3, 4})
	assert.Equal(t, uint32(4), result.Length)

	// Read back
	data := readSlice[uint32](w.Data, int(result.Offset), int(result.Length))
	assert.Equal(t, []uint32{1, 2, 3, 4}, data)
}

func TestWriteEmptySlice(t *testing.T) {
	w := NewWriter(nil)
	w.prepare()
	w.startWriting()

	result := w.writeBytes(nil)
	assert.Equal(t, OdinDocArray{}, result)

	result = w.writeUint32Slice(nil)
	assert.Equal(t, OdinDocArray{}, result)
}

// ============================================================================
// Writer item tracking tests
// ============================================================================

func TestItemTrackerInit(t *testing.T) {
	var tr itemTracker
	tr.init(5)
	assert.Equal(t, 5, tr.len)
	assert.Equal(t, 5, tr.cap)
	assert.Equal(t, 0, tr.offset)
}

func TestWriterWriteTrackedItem(t *testing.T) {
	t.Run("preparing phase increments cap", func(t *testing.T) {
		w2 := NewWriter(nil)
		w2.prepare()
		initialCap := w2.files.cap
		w2.writeTrackedItem(&w2.files, 12, nil)
		assert.Equal(t, initialCap+1, w2.files.cap)
	})

	t.Run("writing phase increments len", func(t *testing.T) {
		w2 := NewWriter(nil)
		w2.prepare()
		// Pre-size the cap by calling writeTrackedItem during prepare phase
		w2.writeTrackedItem(&w2.files, 12, nil)
		w2.startWriting()
		initialLen := w2.files.len
		idx, _ := w2.writeTrackedItem(&w2.files, 12, &OdinDocFile{})
		assert.Equal(t, initialLen+1, w2.files.len)
		assert.Greater(t, idx, uint32(0))
	})
}

// ============================================================================
// GenerateDocumentation tests
// ============================================================================

func TestGenerateDocumentation(t *testing.T) {
	pkg := &AstPackage{
		Name: MakeString("test_pkg"),
		Kind: PackageInit,
		Scope: &Scope{
			Elements: []ScopeElement{
				{
					Name: MakeString("MyVar"),
					Value: &Entity{
						Kind:  EntityVariable,
						Pkg:   nil,
						Token: Token{String: MakeString("MyVar")},
						Variable: EntVariableData{
							IsExport: true,
						},
						DeclInfo: &DeclInfo{
							TypeExpr: &Ast{Kind: AstIdent},
						},
					},
				},
			},
		},
	}
	pkg.Scope.Elements[0].Value.Pkg = pkg

	output := GenerateDocumentation([]*AstPackage{pkg}, 0)

	assert.Contains(t, output, "package test_pkg")
	assert.Contains(t, output, "MyVar")
}

// ============================================================================
// Edge case tests
// ============================================================================

func TestWriterEmptyPackage(t *testing.T) {
	w := NewWriter(nil)

	// Should not panic with empty package list
	w.Write([]*AstPackage{}, "")
}

func TestLoadDocFileEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.odin-doc")
	os.WriteFile(tmpFile, []byte{}, 0644)

	_, _, err := LoadDocFile(tmpFile)
	assert.Error(t, err)
}

func TestSortEmptySlices(t *testing.T) {
	// Should not panic
	SortEntitiesByKind(nil)
	SortEntitiesByKind([]*Entity{})
	SortEntitiesBySrcOrder(nil)
	SortPackagesByName(nil)
}

func TestCmpWithNilPackage(t *testing.T) {
	e1 := &Entity{Pkg: nil, Kind: EntityVariable}
	e2 := &Entity{Pkg: &AstPackage{Name: MakeString("test")}, Kind: EntityVariable}

	// nil package should come before non-nil
	assert.Less(t, CmpEntitiesForPrinting(e1, e2), 0)
	assert.Greater(t, CmpEntitiesForPrinting(e2, e1), 0)
}

func TestDocPrinterShortMode(t *testing.T) {
	var buf bytes.Buffer
	printer := NewDocPrinterTo(&buf, CmdDocFlagShort)

	assert.True(t, printer.isShort())

	pkg := &AstPackage{
		Name: MakeString("short_pkg"),
		Kind: PackageInit,
	}

	printer.PrintPackage(pkg)
	output := buf.String()
	assert.NotContains(t, output, "fullpath:")
}
