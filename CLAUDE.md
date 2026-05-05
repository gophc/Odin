# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Windows (MSVC)
- Debug build: `build.bat`
- Release build: `build.bat release`
- Nightly build: `build.bat debug 1` (second arg enables NIGHTLY define)
- Vendor libs (stb, miniaudio, cgltf): `build_vendor.bat` (called automatically by `build.bat`)
- Requires MSVC x64 native tools (`vswhere.exe` auto-detects; set `ODIN_IGNORE_MSVC_CHECK` to bypass arch check)

### Unix (Linux/macOS/BSD)
- Debug build: `make` or `./build_odin.sh debug`
- Release build: `make release` or `./build_odin.sh release`
- Native optimizations: `make release-native`
- Supported LLVM versions: 14, 17, 18, 19, 20, 21, 22 (15 and 16 are explicitly unsupported)

## Testing

All test commands use consistent vet flags: `-vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do`

```bash
# Core library tests
./odin test tests/core/normal.odin -file -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Core speed tests (optimized)
./odin test tests/core/speed.odin -o:speed -file -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Vendor tests
./odin test tests/vendor -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Internal tests
./odin test tests/internal -all-packages -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -define:ODIN_TEST_FANCY=false -define:ODIN_TEST_FAIL_ON_BAD_MEMORY=true -sanitize:address

# Run demo (smoke test)
./odin run examples/demo -vet -strict-style

# Cross-target check all examples
./odin check examples/all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do
```

## Key Compiler Flags

- `-vet` / `-vet-tabs` / `-vet-style` / `-strict-style` — Style and correctness linters
- `-warnings-as-errors` — Treat all warnings as errors
- `-disallow-do` — Disallow `do` statements
- `-file` — Treat input as single file (not package)
- `-all-packages` — Process all packages in directory
- `-o:speed` — Optimized compilation
- `-target:<triple>` — Cross-compile (e.g., `windows_amd64`, `linux_arm64`, `js_wasm32`)
- `-no-entry-point` — Skip entry point requirement (for libraries)
- `-sanitize:address` — Enable address sanitizer
- `-define:NAME=VALUE` — Set compile-time defines

## Code Architecture

### Compiler (C++) — Single Translation Unit

The entire compiler compiles as a single translation unit. `src/main.cpp` is the entry point that `#include`s every other `.cpp` file:

```
main.cpp → common.cpp → tokenizer.cpp → parser.cpp → checker.cpp →
           linker.cpp → llvm_backend.cpp → ...
```

This means: no header/impl separation needed; static functions are file-local; be mindful of include order.

Key source files:
- `src/main.cpp` — Entry point, `main()`, includes all other sources
- `src/checker.cpp` — Type checking and semantic analysis (largest file at ~240K)
- `src/check_expr.cpp` — Expression type checking (~411K)
- `src/llvm_backend.cpp` — LLVM IR code generation orchestration
- `src/llvm_backend_expr.cpp` — Expression codegen (~222K)
- `src/llvm_backend_proc.cpp` — Procedure codegen (~174K)
- `src/parser.cpp` — Recursive-descent parser (~217K)
- `src/build_settings.cpp` — Build flags, target triples, microarch definitions
- `src/big_int.cpp` — Big integer arithmetic (arbitrary precision)
- `src/libtommath.cpp` — LibTomMath wrapper for compile-time big-int evaluation
- `src/gb/` — Internal utility library (allocators, strings, threading, file I/O)
- `src/libtommath/` — LibTomMath C sources (big integer math library)

### Standard Library (Odin)
- `base/` — Runtime internals: `builtin`, `intrinsics`, `runtime`, `sanitizer`
- `core/` — Standard library packages (`fmt`, `os`, `mem`, `math`, `strings`, `sync`, `thread`, `reflect`, etc.)
- `vendor/` — Third-party library bindings (SDL2/3, glfw, stb, miniaudio, cgltf, OpenGL, Vulkan, DirectX, freetype, etc.)

### Test Structure
- `tests/core/` — Core library tests (entry: `normal.odin` and `speed.odin`)
- `tests/vendor/` — Vendor library tests
- `tests/internal/` — Internal compiler tests
- `tests/issues/` — Regression tests for fixed issues
- `tests/benchmark/` — Performance benchmarks

### Build System Details
- Versioning: Git commit hash and date embedded via `misc/odin.rc` → `odin.res`
- Manifest: `mt.exe` embeds `misc/odin.manifest` for Windows DPI awareness etc.
- Dependencies: LLVM-C shared library (`LLVM-C.dll` at root, `bin/llvm/windows/LLVM-C.lib` for linking)
- On Linux, the built `odin` binary links LLVM dynamically with `-Wl,-rpath=$ORIGIN`

### Key Design Patterns
- Data-oriented design: Heavy use of arrays, pools, and custom allocators (via `src/gb/`)
- Single-pass compilation: No persistent AST; parser feeds directly into checker which builds IR
- Single translation unit: All `.cpp` files `#include`d from `main.cpp`
- Big integer compile-time evaluation: LibTomMath for constant folding
- LLVM-C API: Uses the C API (not C++ API) for LLVM interaction

## Important Constraints
- Do not commit `.env`, credentials, or large binaries
- MSVC 64-bit only (32-bit target not supported)
- C++14 minimum for compiler (build scripts use `-std=c++14`)
- LLVM versions: 14, 17-22 supported; **15 and 16 are explicitly unsupported**
- Vendor `.lib` files on Windows must be prebuilt (see `build_vendor.bat`)

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
