package docs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// String utility functions used across the docs package.

// hasPrefix checks if s starts with prefix.
func hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// trimWhitespace trims leading and trailing whitespace.
func trimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

// removeDirectoryFromPath strips the directory prefix from a path.
func removeDirectoryFromPath(path string) string {
	if i := strings.LastIndexAny(path, "/\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// removeExtension strips the file extension from a path.
func removeExtension(path string) string {
	if i := strings.LastIndex(path, "."); i >= 0 {
		if j := strings.LastIndexAny(path, "/\\"); j > i {
			return path
		}
		return path[:i]
	}
	return path
}

// stringExtensionPosition returns the position of the last '.' or -1.
func stringExtensionPosition(s string) int {
	dot := strings.LastIndex(s, ".")
	slash := strings.LastIndexAny(s, "/\\")
	if dot < 0 || dot < slash {
		return -1
	}
	return dot
}

// extractString reads a null-terminated string from the binary buffer
// using an OdinDocString reference.
func extractString(data []byte, s OdinDocString) string {
	if s.Length == 0 || int(s.Offset)+int(s.Length) > len(data) {
		return ""
	}
	return string(data[s.Offset : s.Offset+s.Length])
}

// ============================================================================
// Doc file loading (from serialize.go)
// ============================================================================

// LoadDocFile reads and parses a .odin-doc binary file.
func LoadDocFile(filename string) (*OdinDocHeader, []byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read .odin-doc file %s: %w", filename, err)
	}

	header := readStructAtOffset[OdinDocHeader](data, 0)
	if header == nil {
		return nil, nil, fmt.Errorf("failed to parse header from %s", filename)
	}

	magic := string(header.Base.Magic[:7])
	if magic != "odindoc" {
		return nil, nil, fmt.Errorf("invalid magic %q in %s", magic, filename)
	}

	return header, data, nil
}

// readSlice reads a slice of structs from a byte slice.
func readSlice[T any](data []byte, offset int, length int) []T {
	if offset <= 0 || length <= 0 || offset >= len(data) {
		return nil
	}
	var zero T
	elemSize := binary.Size(zero)
	if elemSize <= 0 || offset+length*elemSize > len(data) {
		return nil
	}
	result := make([]T, length)
	buf := bytes.NewReader(data[offset : offset+length*elemSize])
	for i := 0; i < length; i++ {
		if err := binary.Read(buf, binary.LittleEndian, &result[i]); err != nil {
			return nil
		}
	}
	return result
}

// ReadString helper for OdinDocHeader.
func (h *OdinDocHeader) ReadString(data []byte, s OdinDocString) string {
	return extractString(data, s)
}

// ReadFiles returns the file entries from the loaded doc.
func (h *OdinDocHeader) ReadFiles(data []byte) []OdinDocFile {
	return readSlice[OdinDocFile](data, int(h.Files.Offset), int(h.Files.Length))
}

// ReadPkgs returns the package entries.
func (h *OdinDocHeader) ReadPkgs(data []byte) []OdinDocPkg {
	return readSlice[OdinDocPkg](data, int(h.Pkgs.Offset), int(h.Pkgs.Length))
}

// ReadEntities returns the entity entries.
func (h *OdinDocHeader) ReadEntities(data []byte) []OdinDocEntity {
	return readSlice[OdinDocEntity](data, int(h.Entities.Offset), int(h.Entities.Length))
}

// ReadTypes returns the type entries.
func (h *OdinDocHeader) ReadTypes(data []byte) []OdinDocType {
	return readSlice[OdinDocType](data, int(h.Types.Offset), int(h.Types.Length))
}
