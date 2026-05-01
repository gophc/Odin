package common

// Path represents a decomposed file path.
type Path struct {
	Basename String
	Name     String
	Ext      String
}

// PathToString converts a Path to a full path string.
func PathToString(a Allocator, p Path) String {
	if p.Basename.Len+p.Name.Len+p.Ext.Len == 0 {
		return String{}
	}
	length := p.Basename.Len + 1 + p.Name.Len
	if p.Ext.Len > 0 {
		length += 1 + p.Ext.Len
	}
	data := make([]byte, length+1)
	pos := 0
	copy(data[pos:], p.Basename.Text[:p.Basename.Len])
	pos += p.Basename.Len
	data[pos] = '/'
	pos++
	copy(data[pos:], p.Name.Text[:p.Name.Len])
	pos += p.Name.Len
	if p.Ext.Len > 0 {
		data[pos] = '.'
		pos++
		copy(data[pos:], p.Ext.Text[:p.Ext.Len])
		pos += p.Ext.Len
	}
	data[length] = 0
	_ = a
	return String{Text: data[:length], Len: length}
}

// QuotePath returns a quoted path string.
func QuotePath(a Allocator, p Path) String {
	temp := PathToString(a, p)
	quoted := Concatenate3Strings(a, MakeString([]byte{34}, 1), temp, MakeString([]byte{34}, 1))
	return quoted
}

// PathFromString decomposes a path string into its parts.
func PathFromString(a Allocator, path String) Path {
	res := Path{}
	if path.Len == 0 {
		return res
	}
	res.Basename = DirectoryFromPath(path)
	res.Basename = CloneString(a, res.Basename)
	nameStart := res.Basename.Len + 1
	if nameStart > path.Len {
		nameStart = path.Len
	}
	res.Name = path.Slice(nameStart, path.Len)
	res.Name = PathRemoveExtension(res.Name)
	res.Name = CloneString(a, res.Name)
	res.Ext = PathExtension(path, false)
	res.Ext = CloneString(a, res.Ext)
	return res
}

// GetWorkingDirectory returns the current working directory.
func GetWorkingDirectory(a Allocator) String {
	// Stub: platform-specific implementation needed.
	_ = a
	return MakeString([]byte{'.'}, 1)
}

// SetWorkingDirectory changes the current working directory.
func SetWorkingDirectory(dir String) bool {
	// Stub: platform-specific implementation needed.
	_ = dir
	return false
}

// PathIsDirectory checks if a path points to a directory.
func PathIsDirectory(path String) bool {
	// Stub: platform-specific implementation needed.
	_ = path
	return false
}

// PathToFullPath converts a relative path to an absolute path.
func PathToFullPath(a Allocator, path String) String {
	_ = a
	return path
}
