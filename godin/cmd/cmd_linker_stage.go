package cmd

import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

func linker_stage(gen *LinkerData) int32 {
	result := int32(0)
	output_filename := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Output])
	debugf("Linking %s\n", goStr(output_filename))

	if is_arch_wasm() {
		timings_start_section(&global_timings, S("wasm-ld"))

		var libStrBuilder strings.Builder
		var extraOrcaFlagsBuilder strings.Builder
		var inputsBuilder strings.Builder

		inputsBuilder.WriteString(fmt.Sprintf("\"%s.o\"", goStr(output_filename)))

		for _, e := range gen.ForeignLibraries {
			extraLinkerFlags := string_trim_whitespace(e.LibraryName.ExtraLinkerFlags)
			if extraLinkerFlags.Len != 0 {
				libStrBuilder.WriteString(fmt.Sprintf(" %s", goStr(extraLinkerFlags)))
			}
			for _, lib := range e.LibraryName.Paths {
				if lib.Len == 0 {
					continue
				}
				if !string_ends_with(lib, S(".o")) {
					continue
				}
				inputsBuilder.WriteString(fmt.Sprintf(" \"%s\"", goStr(lib)))
			}
		}

		if build_context.metrics.os == TargetOs_orca {
			var orcaSdkOutput string
			if !system_exec_command_line_app_output("orca sdk-path", &orcaSdkOutput) {
				gb_printf_err("executing `orca sdk-path` failed, make sure Orca is installed and added to your path\n")
				return 1
			}
			orcaSdkOutput = strings.TrimSpace(orcaSdkOutput)
			if len(orcaSdkOutput) == 0 {
				gb_printf_err("executing `orca sdk-path` did not produce output\n")
				return 1
			}
			inputsBuilder.WriteString(fmt.Sprintf(" \"%s/orca-libc/lib/crt1.o\" \"%s/orca-libc/lib/libc.a\"", orcaSdkOutput, orcaSdkOutput))
			extraOrcaFlagsBuilder.WriteString(fmt.Sprintf(" -L \"%s/bin\" -lorca_wasm --export-dynamic", orcaSdkOutput))
		}

		result = system_exec_command_line_app("wasm-ld",
			"\"%s\\bin\\wasm-ld\" %s -o \"%s\" %s %s %s %s",
			goStr(build_context.ODIN_ROOT),
			inputsBuilder.String(),
			goStr(output_filename),
			goStr(build_context.link_flags),
			goStr(build_context.extra_linker_flags),
			libStrBuilder.String(),
			extraOrcaFlagsBuilder.String())
		return result
	}

	is_cross_linking := false
	is_android := false

	if build_context.cross_compiling && (build_context.different_os || selected_subtarget != Subtarget_Default) {
		switch selected_subtarget {
		case Subtarget_Android:
			is_cross_linking = true
			is_android = true
		default:
			gb_printf_err("Linking for cross compilation for this platform is not yet supported (%s %s)\n",
				goStr(target_os_names[build_context.metrics.os]),
				goStr(target_arch_names[build_context.metrics.arch]))
			build_context.keep_object_files = true
			return result
		}
	}

	section_name := S("msvc-link")
	is_windows := build_context.metrics.os == TargetOs_windows
	is_osx := build_context.metrics.os == TargetOs_darwin

	switch build_context.linker_choice {
	case Linker_Default:
	case Linker_lld:
		section_name = S("lld-link")
	case Linker_radlink:
		section_name = S("rad-link")
	default:
		gb_printf_err("'%s' linker is not supported on this platform\n", goStr(linker_choices[build_context.linker_choice]))
		return 1
	}

	if is_windows {
		timings_start_section(&global_timings, section_name)

		var libStr strings.Builder
		var linkSettings strings.Builder
		linkSettings.Grow(256)

		if build_context.build_paths[BuildPath_VS_LIB].Basename.Len > 0 {
			addPath := func(path String) {
				if path.Data[path.Len-1] == '\\' {
					path.Len--
				}
				linkSettings.WriteString(fmt.Sprintf(" /LIBPATH:\"%s\"", goStr(path)))
			}
			addPath(build_context.build_paths[BuildPath_Win_SDK_UM_Lib].Basename)
			addPath(build_context.build_paths[BuildPath_Win_SDK_UCRT_Lib].Basename)
			addPath(build_context.build_paths[BuildPath_VS_LIB].Basename)
		}

		var minLibsSet StringSet
		string_set_init(&minLibsSet, 64)
		defer string_set_destroy(&minLibsSet)

		var prevLib String

		var asmFiles StringSet
		string_set_init(&asmFiles, 64)
		defer string_set_destroy(&asmFiles)

		var loweredStrings []string

		for _, e := range gen.ForeignLibraries {
			extraLinkerFlags := string_trim_whitespace(e.LibraryName.ExtraLinkerFlags)
			if extraLinkerFlags.Len != 0 {
				libStr.WriteString(fmt.Sprintf(" %s", goStr(extraLinkerFlags)))
			}
			for _, lib := range e.LibraryName.Paths {
				lib = string_trim_whitespace(lib)
				libLowerStr := strings.ToLower(goStr(lib))
				loweredStrings = append(loweredStrings, libLowerStr)
				lib = S(libLowerStr)
				if lib.Len == 0 {
					continue
				}
				if has_asm_extension(lib) {
					if !string_set_update(&asmFiles, lib) {
						asmFile := lib
						var objFile String
						tempDir := temporary_directory(temporary_allocator())
						if tempDir.Len != 0 {
							filename := filename_without_directory(asmFile)
							var strBuilder strings.Builder
							strBuilder.WriteString(goStr(tempDir))
							strBuilder.WriteString("/")
							strBuilder.WriteString(goStr(filename))
							dataPtr := unsafe.StringData(goStr(asmFile))
							strBuilder.WriteString(fmt.Sprintf("-%p.obj", unsafe.Pointer(dataPtr)))
							objFile = make_string_c(S(strBuilder.String()))
						} else {
							objFile = concatenate_strings(permanent_allocator(), asmFile, S(".obj"))
						}
						objFormat := S("win64")
						result = system_exec_command_line_app("nasm",
							"\"%s\\bin\\nasm\\windows\\nasm.exe\" \"%s\" -f \"%s\" -o \"%s\" %s",
							goStr(build_context.ODIN_ROOT),
							goStr(asmFile),
							goStr(objFormat),
							goStr(objFile),
							goStr(build_context.extra_assembler_flags))
						if result != 0 {
							return result
						}
						gen.OutputObjectPaths = append(gen.OutputObjectPaths, objFile)
					}
				} else if !string_set_update(&minLibsSet, lib) || !build_context.min_link_libs {
					if goStr(prevLib) != goStr(lib) {
						libStr.WriteString(fmt.Sprintf(" \"%s\"", goStr(lib)))
					}
					prevLib = lib
				}
			}
		}

		if build_context.build_mode == BuildMode_DynamicLibrary {
			linkSettings.WriteString(" /DLL")
			if build_context.no_entry_point {
				linkSettings.WriteString(" /NOENTRY")
			}
		} else {
			if !(build_context.metrics.arch == TargetArch_i386 && !build_context.no_crt) {
				linkSettings.WriteString(" /ENTRY:mainCRTStartup")
			}
		}

		if build_context.build_paths[BuildPath_Symbols].Name.Len != 0 {
			symbolPath := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Symbols])
			linkSettings.WriteString(fmt.Sprintf(" /PDB:\"%s\"", goStr(symbolPath)))
		}

		if build_context.build_mode != BuildMode_StaticLibrary {
			if build_context.no_crt {
				linkSettings.WriteString(" /nodefaultlib")
			} else {
				linkSettings.WriteString(" /defaultlib:libcmt")
			}
		}

		if build_context.ODIN_DEBUG {
			linkSettings.WriteString(" /DEBUG")
		}

		var objectFiles strings.Builder
		for _, objectPath := range gen.OutputObjectPaths {
			objectFiles.WriteString(fmt.Sprintf("\"%s\" ", goStr(objectPath)))
		}

		vsExePath := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_VS_EXE])
		windowsSdkBinPath := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Win_SDK_Bin_Path])

		var lldLtoFlags strings.Builder
		if build_context.lto_kind != LTO_None {
			lldLtoFlags.WriteString(fmt.Sprintf("/opt:lldltojobs=%d ", build_context.thread_count))
		}

		switch build_context.linker_choice {
		case Linker_lld:
			result = system_exec_command_line_app("msvc-lld-link",
				"\"%s\\bin\\lld-link\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s %s",
				goStr(build_context.ODIN_ROOT),
				objectFiles.String(),
				goStr(output_filename),
				linkSettings.String(),
				goStr(windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM]),
				goStr(build_context.link_flags),
				goStr(build_context.extra_linker_flags),
				libStr.String(),
				lldLtoFlags.String())
			if result != 0 {
				return result
			}
		case Linker_radlink:
			result = system_exec_command_line_app("msvc-rad-link",
				"\"%s\\bin\\radlink\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s",
				goStr(build_context.ODIN_ROOT),
				objectFiles.String(),
				goStr(output_filename),
				linkSettings.String(),
				goStr(windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM]),
				goStr(build_context.link_flags),
				goStr(build_context.extra_linker_flags),
				libStr.String())
			if result != 0 {
				return result
			}
		default:
			resPath := quote_path(heap_allocator(), build_context.build_paths[BuildPath_RES])
			rcPath := quote_path(heap_allocator(), build_context.build_paths[BuildPath_RC])

			if build_context.has_resource {
				if build_context.build_paths[BuildPath_RC].Basename.Len == 0 {
					debugf("Using precompiled resource %s\n", goStr(resPath))
				} else {
					debugf("Compiling resource %s\n", goStr(resPath))
					result = system_exec_command_line_app("msvc-link",
						"\"%src.exe\" /nologo /fo %s %s",
						goStr(windowsSdkBinPath),
						goStr(resPath),
						goStr(rcPath))
					if result != 0 {
						return result
					}
				}
			} else {
				resPath = String{}
			}

			linkerName := S("link.exe")
			switch build_context.build_mode {
			case BuildMode_Executable:
				linkSettings.WriteString(" /NOIMPLIB /NOEXP")
			}

			switch build_context.build_mode {
			case BuildMode_StaticLibrary:
				linkerName = S("lib.exe")
			default:
				linkSettings.WriteString(" /incremental:no /opt:ref")
			}

			result = system_exec_command_line_app("msvc-link",
				"\"%s%s\" %s %s -OUT:\"%s\" %s /nologo /subsystem:%s %s %s %s",
				goStr(vsExePath),
				goStr(linkerName),
				objectFiles.String(),
				goStr(resPath),
				goStr(output_filename),
				linkSettings.String(),
				goStr(windows_subsystem_names[build_context.ODIN_WINDOWS_SUBSYSTEM]),
				goStr(build_context.link_flags),
				goStr(build_context.extra_linker_flags),
				libStr.String())
			if result != 0 {
				return result
			}
		}
	} else {
		timings_start_section(&global_timings, section_name)

		ODIN_ANDROID_API_LEVEL := build_context.ODIN_ANDROID_API_LEVEL
		ODIN_ANDROID_NDK := build_context.ODIN_ANDROID_NDK
		ODIN_ANDROID_NDK_TOOLCHAIN := build_context.ODIN_ANDROID_NDK_TOOLCHAIN
		ODIN_ANDROID_NDK_TOOLCHAIN_LIB := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_LIB
		ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_LIB_LEVEL
		ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT := build_context.ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT

		clangPath := os.Getenv("ODIN_CLANG_PATH")
		hasOdinClangPathEnv := true
		if clangPath == "" {
			clangPath = "clang"
			hasOdinClangPathEnv = false
		}

		var libStr strings.Builder

		var asmFiles StringSet
		string_set_init(&asmFiles, 64)
		defer string_set_destroy(&asmFiles)

		var minLibsSet StringSet
		string_set_init(&minLibsSet, 64)
		defer string_set_destroy(&minLibsSet)

		var prevLib String

		for _, e := range gen.ForeignLibraries {
			extraLinkerFlags := string_trim_whitespace(e.LibraryName.ExtraLinkerFlags)
			if extraLinkerFlags.Len != 0 {
				libStr.WriteString(fmt.Sprintf(" %s", goStr(extraLinkerFlags)))
			}

			if is_osx {
				for _, lib := range e.LibraryName.Paths {
					lib = string_trim_whitespace(lib)
					if lib.Len == 0 {
						continue
					}
					if string_ends_with(lib, S(".framework")) {
						if string_set_update(&minLibsSet, lib) {
							continue
						}
						libName := lib
						libName = remove_extension_from_path(libName)
						libStr.WriteString(fmt.Sprintf(" -framework %s ", goStr(libName)))
					}
				}
			}

			for _, lib := range e.LibraryName.Paths {
				lib = string_trim_whitespace(lib)
				if lib.Len == 0 {
					continue
				}
				if has_asm_extension(lib) {
					if string_set_update(&asmFiles, lib) {
						continue
					}
					asmFile := lib
					var objFile String
					tempDir := temporary_directory(temporary_allocator())
					if tempDir.Len != 0 {
						filename := filename_without_directory(asmFile)
						var strBuilder strings.Builder
						strBuilder.WriteString(goStr(tempDir))
						strBuilder.WriteString("/")
						strBuilder.WriteString(goStr(filename))
						dataPtr := unsafe.StringData(goStr(asmFile))
						strBuilder.WriteString(fmt.Sprintf("-%p.o", unsafe.Pointer(dataPtr)))
						objFile = make_string_c(S(strBuilder.String()))
					} else {
						objFile = concatenate_strings(permanent_allocator(), asmFile, S(".o"))
					}

					var objFormat String
					if build_context.metrics.ptr_size == 8 {
						if is_osx {
							objFormat = S("macho64")
						} else {
							objFormat = S("elf64")
						}
					} else {
						if is_osx {
							objFormat = S("macho32")
						} else {
							objFormat = S("elf32")
						}
					}

					if build_context.metrics.arch == TargetArch_riscv64 {
						result = system_exec_command_line_app("clang",
							"%s \"%s\" -c -o \"%s\" -target %s -march=rv64gc %s",
							clangPath,
							goStr(asmFile),
							goStr(objFile),
							goStr(build_context.metrics.target_triplet),
							goStr(build_context.extra_assembler_flags))
					} else if is_osx {
						result = system_exec_command_line_app("as",
							"as \"%s\" -o \"%s\" %s",
							goStr(asmFile),
							goStr(objFile),
							goStr(build_context.extra_assembler_flags))
					} else {
						result = system_exec_command_line_app("nasm",
							"nasm \"%s\" -f \"%s\" -o \"%s\" %s",
							goStr(asmFile),
							goStr(objFormat),
							goStr(objFile),
							goStr(build_context.extra_assembler_flags))
						if result != 0 {
							gb_printf_err("executing `nasm` to assemble foreign import of %s failed.\n\tSuggestion: `nasm` does not ship with the compiler and should be installed with your system's package manager.\n", goStr(asmFile))
							return result
						}
					}
					gen.OutputObjectPaths = append(gen.OutputObjectPaths, objFile)
				} else {
					shortCircuit := false
					if string_ends_with(lib, S(".framework")) {
						shortCircuit = true
					} else if string_ends_with(lib, S(".dylib")) {
						shortCircuit = true
					} else if string_ends_with(lib, S(".so")) {
						shortCircuit = true
					} else if e.LibraryName.IgnoreDuplicates {
						shortCircuit = true
					}

					if string_set_update(&minLibsSet, lib) && (build_context.min_link_libs || shortCircuit) {
						continue
					}

					if goStr(prevLib) == goStr(lib) {
						continue
					}
					prevLib = lib

					libStr2 := goStr(lib)
					if libStr2 == "System.framework" || libStr2 == "System" || libStr2 == "c" {
						continue
					}

					if is_osx {
						if string_ends_with(lib, S(".framework")) {
							libName := lib
							libName = remove_extension_from_path(libName)
							libStr.WriteString(fmt.Sprintf(" -framework %s ", goStr(libName)))
						} else if string_ends_with(lib, S(".a")) || string_ends_with(lib, S(".o")) || string_ends_with(lib, S(".dylib")) {
							libStr.WriteString(fmt.Sprintf(" \"%s\" ", goStr(lib)))
						} else {
							libStr.WriteString(fmt.Sprintf(" -l%s ", goStr(lib)))
						}
					} else {
						if string_ends_with(lib, S(".a")) || string_ends_with(lib, S(".o")) || string_ends_with(lib, S(".so")) || string_contains_string(lib, S(".so.")) {
							libStr.WriteString(fmt.Sprintf(" -l:\"%s\" ", goStr(lib)))
						} else {
							libStr.WriteString(fmt.Sprintf(" -l%s ", goStr(lib)))
						}
					}
				}
			}
		}

		var objectFiles strings.Builder

		if is_android {
			debugf("[Section] %s\n", "Android Native App Glue Compile")
			if build_context.show_more_timings {
				timings_start_section(&global_timings, S("Android Native App Glue Compile"))
			}
			var androidGlueObject String
			var androidGlueStaticLib String
			var hashBuf [64]byte
			hashStr := fmt.Sprintf("%p", &hashBuf)
			hash := make_string_c(S(hashStr))
			tempDir := normalize_path(temporary_allocator(), temporary_directory(temporary_allocator()), NIX_SEPARATOR_STRING)
			androidGlueObject = concatenate4_strings(temporary_allocator(), tempDir, S("android_native_app_glue-"), hash, S(".o"))
			androidGlueStaticLib = concatenate4_strings(permanent_allocator(), tempDir, S("libandroid_native_app_glue-"), hash, S(".a"))

			var glue strings.Builder
			glue.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			glue.WriteString(fmt.Sprintf("bin/clang --target=%s%d ", goStr(build_context.metrics.target_triplet), ODIN_ANDROID_API_LEVEL))
			glue.WriteString("-c \"")
			glue.WriteString(goStr(ODIN_ANDROID_NDK))
			glue.WriteString("sources/android/native_app_glue/android_native_app_glue.c")
			glue.WriteString("\" ")
			glue.WriteString("-o \"")
			glue.WriteString(goStr(androidGlueObject))
			glue.WriteString("\" ")
			glue.WriteString("--sysroot \"")
			glue.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			glue.WriteString("sysroot")
			glue.WriteString("\" ")
			glue.WriteString("\"-I")
			glue.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			glue.WriteString("sysroot/usr/include/")
			glue.WriteString("\" ")
			glue.WriteString("\"-I")
			glue.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			glue.WriteString("sysroot/usr/include/")
			glue.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN_LIB))
			glue.WriteString("/\" ")
			glue.WriteString("-Wno-macro-redefined ")
			result = system_exec_command_line_app("android-native-app-glue-compile", glue.String())
			if result != 0 {
				return result
			}

			debugf("[Section] %s\n", "Android Native App Glue ar")
			if build_context.show_more_timings {
				timings_start_section(&global_timings, S("Android Native App Glue ar"))
			}
			var ar strings.Builder
			ar.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			ar.WriteString("bin/llvm-ar rcs \"")
			ar.WriteString(goStr(androidGlueStaticLib))
			ar.WriteString("\" \"")
			ar.WriteString(goStr(androidGlueObject))
			ar.WriteString("\" ")
			result = system_exec_command_line_app("android-native-app-glue-ar", ar.String())
			if result != 0 {
				return result
			}
			objectFiles.WriteString(fmt.Sprintf("\"%s\" ", goStr(androidGlueStaticLib)))
		}

		for _, objectPath := range gen.OutputObjectPaths {
			objectFiles.WriteString(fmt.Sprintf("\"%s\" ", goStr(objectPath)))
		}

		var linkSettings strings.Builder
		linkSettings.Grow(32)
		if build_context.no_crt {
			linkSettings.WriteString("-nostdlib ")
		}

		if build_context.build_mode == BuildMode_StaticLibrary {
			debugf("[Section] %s\n", "Static Library Creation")
			if build_context.show_more_timings {
				timings_start_section(&global_timings, S("Static Library Creation"))
			}
			var arCommand strings.Builder
			arCommand.WriteString("ar rcs ")
			arCommand.WriteString(fmt.Sprintf("\"%s\" ", goStr(output_filename)))
			arCommand.WriteString(objectFiles.String())
			result = system_exec_command_line_app("ar", arCommand.String())
			if result != 0 {
				return result
			}
			return result
		}

		if build_context.build_mode == BuildMode_DynamicLibrary {
			linkSettings.WriteString("-shared ")
			if is_osx {
				linkSettings.WriteString("-Wl,-init,'__odin_entry_point' ")
			} else {
				linkSettings.WriteString("-Wl,-init,'_odin_entry_point' ")
				linkSettings.WriteString("-Wl,-fini,'_odin_exit_point' ")
			}
		} else if is_android {
			linkSettings.WriteString("-shared ")
		}

		if build_context.build_mode == BuildMode_Executable && build_context.reloc_mode == RelocMode_PIC {
		} else if build_context.build_mode != BuildMode_DynamicLibrary {
			if build_context.metrics.os != TargetOs_openbsd &&
				build_context.metrics.os != TargetOs_haiku &&
				build_context.metrics.arch != TargetArch_riscv64 &&
				!is_android {
				linkSettings.WriteString("-no-pie ")
			}
		}

		var platformLibStr strings.Builder

		if is_osx {
			var darwinSdkPathBuf strings.Builder
			darwinPlatformName := "MacOSX"
			darwinXcrunSdkName := "macosx"
			darwinMinVersionId := "macosx"
			originalClangPath := clangPath

			switch selected_subtarget {
			case Subtarget_iPhone:
				darwinPlatformName = "iPhoneOS"
				darwinXcrunSdkName = "iphoneos"
				darwinMinVersionId = "ios"
				if !hasOdinClangPathEnv {
					clangPath = "xcrun --sdk iphoneos clang"
				}
			case Subtarget_iPhoneSimulator:
				darwinPlatformName = "iPhoneSimulator"
				darwinXcrunSdkName = "iphonesimulator"
				darwinMinVersionId = "ios-simulator"
				if !hasOdinClangPathEnv {
					clangPath = "xcrun --sdk iphonesimulator clang"
				}
			}

			darwinFindSdkCmd := fmt.Sprintf("xcrun --sdk %s --show-sdk-path", darwinXcrunSdkName)
			var darwinSdkPath string
			if !system_exec_command_line_app_output(darwinFindSdkCmd, &darwinSdkPath) {
				clangPath = originalClangPath
				darwinSdkPathBuf.WriteString(fmt.Sprintf("/Library/Developer/CommandLineTools/SDKs/%s.sdk", darwinPlatformName))
				if !path_is_directory(make_string_c(S(darwinSdkPathBuf.String()))) {
					darwinSdkPathBuf.Reset()
					darwinSdkPathBuf.WriteString(fmt.Sprintf("/Applications/Xcode.app/Contents/Developer/Platforms/%s.platform/Developer/SDKs/%s.sdk", darwinPlatformName, darwinPlatformName))
					if !path_is_directory(make_string_c(S(darwinSdkPathBuf.String()))) {
						gb_printf_err("Failed to find %s SDK\n", darwinPlatformName)
						return -1
					}
				}
				darwinSdkPath = darwinSdkPathBuf.String()
			} else {
				darwinSdkPath = strings.TrimSpace(darwinSdkPath)
			}

			platformLibStr.WriteString(fmt.Sprintf("--sysroot %s ", darwinSdkPath))
			platformLibStr.WriteString("-L/usr/local/lib ")
			if gb_file_exists("/opt/homebrew/lib") {
				platformLibStr.WriteString("-L/opt/homebrew/lib ")
			}
			if gb_file_exists("/opt/local/lib") {
				platformLibStr.WriteString("-L/opt/local/lib ")
			}

			if build_context.minimum_os_version_string_given || selected_subtarget != Subtarget_Default {
				linkSettings.WriteString(fmt.Sprintf("-m%s-version-min=%s ", darwinMinVersionId, goStr(build_context.minimum_os_version_string)))
			}

			if build_context.build_mode != BuildMode_DynamicLibrary {
				linkSettings.WriteString("-e _main ")
			}
		} else if build_context.metrics.os == TargetOs_freebsd {
			if build_context.sanitizer_flags&(SanitizerFlag_Address|SanitizerFlag_Memory) != 0 {
				platformLibStr.WriteString("-lpthread ")
			}
			platformLibStr.WriteString("-Wl,-L/usr/local/lib ")
		} else if build_context.metrics.os == TargetOs_openbsd {
			platformLibStr.WriteString("-lpthread -Wl,-L/usr/local/lib ")
			platformLibStr.WriteString("-Wl,-z,nobtcfi ")
		}

		if is_android {
			platformLibStr.WriteString("\"-L")
			platformLibStr.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT))
			platformLibStr.WriteString("usr/lib/")
			platformLibStr.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN_LIB))
			platformLibStr.WriteString(fmt.Sprintf("/%d", ODIN_ANDROID_API_LEVEL))
			platformLibStr.WriteString("\" ")
			platformLibStr.WriteString("-landroid ")
			platformLibStr.WriteString("-llog ")
			platformLibStr.WriteString("\"--sysroot=")
			platformLibStr.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN_SYSROOT))
			platformLibStr.WriteString("\" ")
			linkSettings.WriteString("-u ANativeActivity_onCreate ")
		}

		if !build_context.no_rpath {
			if is_osx {
				linkSettings.WriteString("-Wl,-rpath,@loader_path ")
			} else {
				if !is_android {
					linkSettings.WriteString("-Wl,-rpath,$ORIGIN ")
				}
			}
		}

		if !build_context.no_crt {
			libStr.WriteString("-lm ")
			if !is_osx {
				libStr.WriteString("-lc ")
			}
		}

		var linkCommandLine strings.Builder

		if is_android {
			linkCommandLine.WriteString(goStr(ODIN_ANDROID_NDK_TOOLCHAIN))
			linkCommandLine.WriteString(fmt.Sprintf("bin/clang --target=%s%d ", goStr(build_context.metrics.target_triplet), ODIN_ANDROID_API_LEVEL))
		} else {
			linkCommandLine.WriteString(clangPath)
		}

		linkCommandLine.WriteString(" -Wno-unused-command-line-argument ")

		if build_context.lto_kind != LTO_None {
			linkCommandLine.WriteString(" -flto=thin")
			linkCommandLine.WriteString(fmt.Sprintf(" -flto-jobs=%d ", build_context.thread_count))
			if build_context.ODIN_DEBUG {
				linkCommandLine.WriteString(" -g ")
			}
			if is_osx && !build_context.minimum_os_version_string_given {
				linkCommandLine.WriteString(" -Wno-override-module ")
			}
		}

		linkCommandLine.WriteString(objectFiles.String())
		linkCommandLine.WriteString(fmt.Sprintf(" -o \"%s\" ", goStr(output_filename)))
		linkCommandLine.WriteString(fmt.Sprintf(" %s ", platformLibStr.String()))
		linkCommandLine.WriteString(fmt.Sprintf(" %s ", libStr.String()))
		linkCommandLine.WriteString(fmt.Sprintf(" %s ", goStr(build_context.link_flags)))
		linkCommandLine.WriteString(fmt.Sprintf(" %s ", goStr(build_context.extra_linker_flags)))
		linkCommandLine.WriteString(fmt.Sprintf(" %s ", linkSettings.String()))

		if is_android {
			debugf("[Section] %s\n", "Linking")
			if build_context.show_more_timings {
				timings_start_section(&global_timings, S("Linking"))
			}
		}

		if build_context.linker_choice == Linker_lld {
			linkCommandLine.WriteString(" -fuse-ld=lld")
			result = system_exec_command_line_app("lld-link", linkCommandLine.String())
		} else if build_context.linker_choice == Linker_mold {
			linkCommandLine.WriteString(" -fuse-ld=mold")
			result = system_exec_command_line_app("mold-link", linkCommandLine.String())
		} else {
			result = system_exec_command_line_app("ld-link", linkCommandLine.String())
		}

		if result != 0 {
			return result
		}

		if is_osx && build_context.ODIN_DEBUG {
			result = system_exec_command_line_app("dsymutil", "dsymutil \"%s\"", goStr(output_filename))
			if result != 0 {
				return result
			}
		}
	}

	return result
}
