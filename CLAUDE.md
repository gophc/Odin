# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Windows (MSVC)
- Debug build: `build.bat` or `build.bat debug`
- Release build: `build.bat release`
- Nightly build: `build.bat debug 1` (second arg enables nightly defines)
- Requires MSVC x64 native tools (vswhere auto-detects installation)

### Unix (Linux/macOS)
- Debug build: `make debug` or `./build_odin.sh debug`
- Release build: `make release` or `./build_odin.sh release`
- Native optimizations: `make release-native`

## Testing

- Run demo: After building, `odin run examples/demo -vet -strict-style` (Windows) or `./odin run examples/demo/demo.odin -file` (Unix)
- Full test suite: Not explicitly defined; tests are in `tests/` directory
- Run specific test: `odin test tests/core/` (if test runner exists)

## Code Architecture

### Compiler (C++)
- Entry: `src/main.cpp`
- Core components in `src/`:
  - `checker.cpp` - Type checking and semantic analysis
  - `llvm_backend.cpp` - LLVM IR generation (LLVM-C API)
  - `parser.cpp` - Source parsing
  - `types.cpp` - Type system implementation
  - `libtommath.cpp` - Big integer arithmetic for compile-time evaluation
- Uses preprocessor for platform abstraction: `src/cipp/` contains preprocessed intermediate files

### Build System Details
- Versioning: Embed git commit hash and date into binary (via `misc/odin.rc`)
- Resource compilation: `rc.exe` generates `.res` file with version info
- Manifest embedding: `mt.exe` embeds `misc/odin.manifest` for modern Windows features
- Dependencies: LLVM-C library (`bin/llvm/windows/LLVM-C.lib` on Windows)

### Additional Tools
- `godin/` - Go implementation of libtommath utilities (used for testing/validation)
- `misc/featuregen/` - Feature generation tool (C++)
- `tests/` - Test suites, including `core/sys/windows/win32gen/` for Windows binding generation

### Key Design Patterns
- Data-oriented design: Heavy use of arrays, pools, and custom allocators
- Single-pass compilation (no persistent AST; builds IR directly)
- Libtommath integration: Used for compile-time constant folding of big integers
- LLVM backend: Generates optimized machine code via LLVM's optimization passes

## Important Constraints
- Do not commit `.env`, credentials, or large binaries
- MSVC 64-bit only (32-bit target not supported)
- Requires C++17 or later for compiler codebase
- LLVM version: Compatible with LLVM 15+ (adjust paths if needed)

## Common Workflows
- Add new builtin: Modify `src/builtin.cpp` and update `src/check_builtin.cpp`
- Add new type: Update `src/types.cpp`, `src/check_type.cpp`, and LLVM backend in `src/llvm_backend_type.cpp`
- Modify code generation: Change `src/llvm_backend_proc.cpp` or `src/llvm_backend_expr.cpp`
- Update libtommath: Replace files in `src/libtommath/` and rebuild
