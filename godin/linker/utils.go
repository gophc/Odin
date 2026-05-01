package linker

import (
	"fmt"
	"path/filepath"
	"strings"
)

// StringSet is a simple set of strings.
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet with optional initial capacity.
func NewStringSet(capacity int) StringSet {
	if capacity > 0 {
		return make(StringSet, capacity)
	}
	return make(StringSet)
}

// Add adds a string to the set. Returns true if the string was not already present.
func (s StringSet) Add(key string) bool {
	if _, exists := s[key]; exists {
		return false
	}
	s[key] = struct{}{}
	return true
}

// Contains checks if the set contains the given string.
func (s StringSet) Contains(key string) bool {
	_, exists := s[key]
	return exists
}

// Remove removes a string from the set.
func (s StringSet) Remove(key string) {
	delete(s, key)
}

// Array is a generic slice wrapper for convenience.
type Array[T any] struct {
	items []T
}

// NewArray creates a new Array with optional initial capacity.
func NewArray[T any](capacity int) *Array[T] {
	if capacity > 0 {
		return &Array[T]{items: make([]T, 0, capacity)}
	}
	return &Array[T]{items: make([]T, 0)}
}

// Add appends an item to the array.
func (a *Array[T]) Add(item T) {
	a.items = append(a.items, item)
}

// Get returns the item at index i.
func (a *Array[T]) Get(i int) T {
	return a.items[i]
}

// Len returns the length of the array.
func (a *Array[T]) Len() int {
	return len(a.items)
}

// Slice returns the underlying slice.
func (a *Array[T]) Slice() []T {
	return a.items
}

// Helper string functions

// stringEndsWith checks if s ends with suffix.
func stringEndsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// stringToLower returns the lowercased string.
func stringToLower(s string) string {
	return strings.ToLower(s)
}

// stringContainsString checks if substr is in s.
func stringContainsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// hasASMExtension checks if the file path has a known assembly extension.
func hasASMExtension(path string) bool {
	// Simplified: check for .s or .asm
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".s") || strings.HasSuffix(lower, ".asm")
}

// filenameWithoutDirectory returns the base filename.
func filenameWithoutDirectory(path string) string {
	// Simplified: get last component after slash or backslash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// quotePath wraps a path in quotes if it contains spaces.
func quotePath(path string) string {
	if strings.Contains(path, " ") {
		return "\"" + path + "\""
	}
	return path
}

// pathIsDirectory checks if a path is a directory.
func pathIsDirectory(path string) bool {
	// Stub for now; can be implemented with os.Stat if needed
	return false
}

// concatenateStrings concatenates multiple strings.
func concatenateStrings(ss ...string) string {
	var builder strings.Builder
	for _, s := range ss {
		builder.WriteString(s)
	}
	return builder.String()
}

// concatenate4Strings concatenates 4 strings.
func concatenate4Strings(a, b, c, d string) string {
	return a + b + c + d
}

// normalizePath normalizes a path by cleaning redundant separators.
func normalizePath(path string) string {
	return strings.ReplaceAll(filepath.Clean(path), "\\", "/")
}

// temporaryDirectory returns a temporary directory path.
func temporaryDirectory() string {
	// Stub; can be implemented with os.TempDir() if needed
	return ""
}

// gbString is a simple string builder wrapper.
type gbString struct {
	strings.Builder
}

func gbStringMake() *gbString {
	return &gbString{}
}

func gbStringAppendFmt(s *gbString, format string, args ...interface{}) *gbString {
	fmt.Fprintf(s, format, args...)
	return s
}

func gbStringAppendc(s *gbString, str string) *gbString {
	s.WriteString(str)
	return s
}

func gbStringAppendLength(s *gbString, str string, length int) *gbString {
	if length > len(str) {
		length = len(str)
	}
	s.WriteString(str[:length])
	return s
}

func gbStringAppend(s *gbString, str string) *gbString {
	s.WriteString(str)
	return s
}

func (s *gbString) String() string {
	return s.Builder.String()
}

func gbStringFree(s *gbString) {
	// No-op in Go
}

func gbStringMakeReserve(capacity int) *gbString {
	s := &gbString{}
	s.Grow(capacity)
	return s
}

func gbStringClear(s *gbString) {
	s.Reset()
}

func gbStringTrimSpace(s *gbString) *gbString {
	trimmed := strings.TrimSpace(s.String())
	newStr := gbStringMake()
	newStr.WriteString(trimmed)
	return newStr
}

// debugf prints debug output if requested.
var debugEnabled = false

func debugf(format string, args ...interface{}) {
	if debugEnabled {
		fmt.Printf(format, args...)
	}
}

// Timings struct for profiling.
type Timings struct {
	// Simplified version
}

var globalTimings = &Timings{}

func timingsStartSection(timings *Timings, name string) {
	// Stub
}

// gbAssertHandler simulates the C++ assertion handler.
func gbAssertHandler(file string, line int) {
	panic(fmt.Sprintf("Assertion failure at %s:%d", file, line))
}

// Subtarget constants
const (
	SubtargetDefault         = 0
	SubtargetAndroid         = 1
	SubtargetIPhone          = 2
	SubtargetIPhoneSimulator = 3
)

// Windows subsystem names
var windowsSubsystemNames = []string{
	"console",
	"windows",
	"native",
	"efi_application",
	"efi_boot_service",
	"efi_runtime",
	"efi_rom",
}

// Linker choices strings
var linkerChoices = []string{
	"default",
	"lld",
	"radlink",
	"mold",
}

// Target OS names
var targetOSNames = []string{
	"windows",
	"darwin",
	"linux",
	"freebsd",
	"openbsd",
	"android",
	"orca",
	"haiku",
}

// Target arch names
var targetArchNames = []string{
	"i386",
	"amd64",
	"arm64",
	"riscv64",
}
