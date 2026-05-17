package sys

import (
	"io/fs"
	"math"
)

var (
	walltime Walltime
	info     fs.FileInfo
	st       Stat_t
)

// This shows typical conversions to EpochNanos type, for Stat_t fields.
func Example_epochNanos() {
	// Convert an adapted fs.File's fs.FileInfo to Mtim.
	st.Mtim = info.ModTime().UnixNano()

	// Generate a fake Atim using Walltime passed to wazero.ModuleConfig.
	sec, nsec := walltime()
	st.Atim = sec*1e9 + int64(nsec)
}

type fileInfoWithSys struct {
	fs.FileInfo
	st Stat_t
}

func (f *fileInfoWithSys) Sys() any { return &f.st }

// This shows how to return data not defined in fs.FileInfo, notably Inode.
func Example_inode() {
	st := NewStat_t(info)
	st.Ino = math.MaxUint64 // arbitrary non-zero value
	info = &fileInfoWithSys{info, st}
}
