# Odin Programming Language - Agent Guide

This is the Odin compiler and core library repository.

## Build

- **Windows**: `build.bat` (debug) or `build.bat 1` (release)
- **Unix**: `make` or `./build_odin.sh` (debug/release/nightly)

## Common Commands

```bash
# Run demo
./odin run examples/demo

# Run with debug
./odin run examples/demo -debug

# Check all examples (strict lint)
./odin check examples/all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do

# Check SDL3 examples (no entry point)
./odin check examples/all/sdl3 -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -no-entry-point

# Run core library tests
./odin test tests/core/normal.odin -file -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Run optimized tests
./odin test tests/core/speed.odin -o:speed -file -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Run vendor tests
./odin test tests/vendor -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Run internal tests
./odin test tests/internal -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address
```

## Structure

- `core/` - Standard library packages
- `vendor/` - Third-party library bindings (stb, cgltf, miniaudio, SDL2/3, etc.)
- `base/` - Runtime internals (allocator, thread, profiler)
- `src/` - Compiler source (C++)
- `examples/` - Example code
- `tests/` - Test suites (core, vendor, internal, issues, benchmark)

## Key Flags

- `-vet` - Enable vet linter
- `-strict-style` / `-vet-style` - Enforce strict code style
- `-warnings-as-errors` - Treat warnings as errors
- `-disallow-do` - Disallow `do` statements
- `-file` - Treat input as single file (not package)
- `-all-packages` - Process all packages in directory
- `-o:speed` - Optimized compilation
- `-target:<triple>` - Cross-compile (e.g., `windows_amd64`, `linux_i386`, `js_wasm32`)
- `-no-entry-point` - Skip entry point requirement (for libraries)
- `-sanitize:address` - Enable address sanitizer

## Windows Build Notes

Run from Visual Studio x64 Native Tools Command Prompt or ensure `cl.exe` is in PATH.