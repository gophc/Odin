package linker

import (
	"path/filepath"
	"strings"
)

// BuildMode represents the type of output to produce.
type BuildMode int

const (
	BuildModeExecutable BuildMode = iota
	BuildModeDynamicLibrary
	BuildModeStaticLibrary
)

// LinkerChoice represents the linker to use.
type LinkerChoice int

const (
	LinkerDefault LinkerChoice = iota
	LinkerLLD
	LinkerRadLink
	LinkerMold
)

// LTOKind represents the level of Link-Time Optimization.
type LTOKind int

const (
	LTONone LTOKind = iota
	LTOThin
)

// RelocMode represents relocation mode.
type RelocMode int

const (
	RelocModeDefault RelocMode = iota
	RelocModePIC
)

// TargetOS represents the operating system target.
type TargetOS int

const (
	TargetOSWindows TargetOS = iota
	TargetOSDarwin
	TargetOSLinux
	TargetOSFreeBSD
	TargetOSOpenBSD
	TargetOSAndroid
	TargetOSOrca
	TargetOSHaiku
)

// TargetArch represents the architecture target.
type TargetArch int

const (
	TargetArchI386 TargetArch = iota
	TargetArchAMD64
	TargetArchARM64
	TargetArchRiscV64
)

// WindowsSubsystem represents the subsystem for Windows executables.
type WindowsSubsystem int

const (
	WindowsSubsystemConsole WindowsSubsystem = iota
	WindowsSubsystemWindows
	WindowsSubsystemNative
	WindowsSubsystemEFIApplication
	WindowsSubsystemEFIBootService
	WindowsSubsystemEFIRuntime
	WindowsSubsystemEFIROM
)

// SanitizerFlag represents sanitizer options.
type SanitizerFlag int

const (
	SanitizerFlagNone    SanitizerFlag = 0
	SanitizerFlagAddress SanitizerFlag = 1 << iota
	SanitizerFlagMemory
)

// BuildPaths holds various paths used during building.
type BuildPaths struct {
	Output        string
	VSExe         string
	VSLib         string
	WinSDKUMLib   string
	WinSDKUCRTLib string
	WinSDKBinPath string
	Symbols       string
	RES           string
	RC            string
}

// Metrics holds target metrics.
type Metrics struct {
	OS            TargetOS
	Arch          TargetArch
	PtrSize       int
	TargetTriplet string
}

// BuildContext holds all the build configuration.
type BuildContext struct {
	OutFilepath                     string
	ODINRoot                        string
	LinkFlags                       string
	ExtraLinkerFlags                string
	ExtraAssemblerFlags             string
	ODINDebug                       bool
	LinkerChoice                    LinkerChoice
	BuildMode                       BuildMode
	NoEntryPoint                    bool
	NoCRT                           bool
	MinLinkLibs                     bool
	WindowsSubsystem                WindowsSubsystem
	HasResource                     bool
	KeepObjectFiles                 bool
	CrossCompiling                  bool
	DifferentOS                     bool
	LTOKind                         LTOKind
	ThreadCount                     int
	SanitizerFlags                  SanitizerFlag
	RelocMode                       RelocMode
	MinimumOSVersionString          string
	MinimumOSVersionStringGiven     bool
	NoRPATH                         bool
	ODINAndroidAPILevel             int
	ODINAndroidNDK                  string
	ODINAndroidNDKToolchain         string
	ODINAndroidNDKToolchainLib      string
	ODINAndroidNDKToolchainLibLevel string
	ODINAndroidNDKToolchainSysroot  string
	Metrics                         Metrics
	ShowMoreTimings                 bool
	BuildPaths                      BuildPaths
	SelectedSubtarget               int // corresponds to Subtarget enum
}

// LinkerData holds the state for the linking process.
type LinkerData struct {
	ForeignLibrariesSet      map[string]struct{}
	ForeignLibraries         []*LibraryNameEntity
	OutputObjectPaths        []string
	OutputTempPaths          []string
	OutputBase               string
	OutputName               string
	NeedsSystemLibraryLinked bool
}

// LibraryNameEntity represents a foreign library to link.
type LibraryNameEntity struct {
	Paths            []string
	ExtraLinkerFlags string
	IgnoreDuplicates bool
}

// NewLinkerData creates and initializes a new LinkerData.
func NewLinkerData() *LinkerData {
	return &LinkerData{
		ForeignLibrariesSet: make(map[string]struct{}),
		ForeignLibraries:    make([]*LibraryNameEntity, 0),
		OutputObjectPaths:   make([]string, 0),
		OutputTempPaths:     make([]string, 0),
	}
}

// EnableSystemLibraryLinking marks that system libraries need to be linked.
func (ld *LinkerData) EnableSystemLibraryLinking() {
	ld.NeedsSystemLibraryLinked = true
}

// AddForeignLibrary adds a foreign library to be linked.
func (ld *LinkerData) AddForeignLibrary(lib *LibraryNameEntity) {
	// Check for duplicates using a set based on a key.
	// For simplicity, we use a composite key: paths joined and flags.
	key := strings.Join(lib.Paths, ";") + "|" + lib.ExtraLinkerFlags
	if _, exists := ld.ForeignLibrariesSet[key]; exists {
		return
	}
	ld.ForeignLibrariesSet[key] = struct{}{}
	ld.ForeignLibraries = append(ld.ForeignLibraries, lib)
}

// AddOutputObjectPath adds an object file path to the list.
func (ld *LinkerData) AddOutputObjectPath(path string) {
	ld.OutputObjectPaths = append(ld.OutputObjectPaths, path)
}

// AddOutputTempPath adds a temporary file path to the list.
func (ld *LinkerData) AddOutputTempPath(path string) {
	ld.OutputTempPaths = append(ld.OutputTempPaths, path)
}

// Init initializes the LinkerData from CheckerInfo and initial fullpath.
// This mimics linker_data_init.
func (ld *LinkerData) Init(info *CheckerInfo, initFullpath string, buildCtx *BuildContext) {
	if buildCtx.OutFilepath == "" {
		ld.OutputName = RemoveDirectoryFromPath(initFullpath)
		ld.OutputName = RemoveExtensionFromPath(ld.OutputName)
		ld.OutputName = strings.TrimSpace(ld.OutputName)
		if ld.OutputName == "" {
			ld.OutputName = info.InitScope.Pkg.Name
		}
		ld.OutputBase = ld.OutputName
	} else {
		ld.OutputName = strings.TrimSpace(buildCtx.OutFilepath)
		if ld.OutputName == "" {
			ld.OutputName = info.InitScope.Pkg.Name
		}
		pos := strings.LastIndex(ld.OutputName, ".")
		if pos < 0 {
			ld.OutputBase = ld.OutputName
		} else {
			ld.OutputBase = ld.OutputName[:pos]
		}
	}
	// Convert to full path (simplified: use absolute path)
	abs, err := filepath.Abs(ld.OutputBase)
	if err == nil {
		ld.OutputBase = abs
	}
}

// CheckerInfo is a simplified version of the original CheckerInfo.
type CheckerInfo struct {
	InitScope *Scope
}

// Scope represents a scope in the checker.
type Scope struct {
	Pkg *Package
}

// Package represents a package.
type Package struct {
	Name string
}

// Helper functions (to be implemented in utils.go)

// RemoveDirectoryFromPath removes the directory part from a file path.
func RemoveDirectoryFromPath(path string) string {
	return filepath.Base(path)
}

// RemoveExtensionFromPath removes the file extension.
func RemoveExtensionFromPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path
	}
	return strings.TrimSuffix(path, ext)
}
