package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func add_library_collection(name String, path String) {
	lc := LibraryCollections{Name: name, Path: string_trim_whitespace(path)}
	libraryCollections = append(libraryCollections, lc)
}

func find_library_collection_path(name String, path *String) bool {
	for _, lc := range libraryCollections {
		if str_eq(lc.Name, name) {
			if path != nil {
				*path = lc.Path
			}
			return true
		}
	}
	return false
}

func odin_root_dir() String {
	if globalModulePathSet {
		return globalModulePath
	}
	found := gb_get_env("ODIN_ROOT", permanent_allocator())
	if found.Len > 0 {
		path := path_to_fullpath(permanent_allocator(), found, nil)
		path = normalize_path(permanent_allocator(), path, WIN32_SEPARATOR_STRING)
		globalModulePath = path
		globalModulePathSet = true
		return globalModulePath
	}
	return internal_odin_root_dir()
}

func internal_odin_root_dir() String {
	if globalModulePathSet {
		return globalModulePath
	}
	exe, err := os.Executable()
	if err != nil {
		return String{}
	}
	exePath := make_string(heap_allocator(), exe)
	exePath = path_to_fullpath(heap_allocator(), exePath, nil)
	for i := exePath.Len - 1; i >= 0; i-- {
		c := *(*byte)(unsafe.Add(unsafe.Pointer(exePath.Data), i))
		if c == '/' || c == '\\' {
			break
		}
		exePath.Len--
	}
	globalModulePath = exePath
	globalModulePathSet = true
	return exePath
}

func path_to_fullpath(a gbAllocator, s String, ok_ *bool) String {
	goPath := goStr(s)
	absPath, err := filepath.Abs(goPath)
	if err != nil {
		if ok_ != nil {
			*ok_ = false
		}
		return String{}
	}
	absPath = strings.ReplaceAll(absPath, "\\", "/")
	if ok_ != nil {
		*ok_ = true
	}
	return make_string(a, absPath)
}

func get_fullpath_relative(a gbAllocator, base_dir String, path String, ok_ *bool) String {
	result := concatenate3_strings(a, base_dir, S("/"), path)
	result = string_trim_whitespace(result)
	return path_to_fullpath(a, result, ok_)
}

func get_fullpath_base_collection(a gbAllocator, path String, ok_ *bool) String {
	module_dir := odin_root_dir()
	result := concatenate3_strings(a, module_dir, S("base/"), path)
	result = string_trim_whitespace(result)
	return path_to_fullpath(a, result, ok_)
}

func get_fullpath_core_collection(a gbAllocator, path String, ok_ *bool) String {
	module_dir := odin_root_dir()
	result := concatenate3_strings(a, module_dir, S("core/"), path)
	result = string_trim_whitespace(result)
	return path_to_fullpath(a, result, ok_)
}

func is_arch_wasm() bool {
	switch buildContext.Metrics.Arch {
	case TargetArchWasm32, TargetArchWasm64p32:
		return true
	default:
		return false
	}
}

func infer_object_extension_from_build_context() String {
	if is_arch_wasm() {
		return S("wasm.o")
	}
	switch buildContext.Metrics.Os {
	case TargetOsWindows:
		return S("obj")
	case TargetOsDarwin, TargetOsLinux, TargetOsFreeBSD,
		TargetOsOpenBSD, TargetOsNetBSD, TargetOsHaiku,
		TargetOsFreestanding:
		switch buildContext.Metrics.ABI {
		case TargetABIDefault, TargetABISysV:
			return S("o")
		case TargetABIWin64:
			return S("obj")
		}
	}
	return S("o")
}

func init_build_paths(init_filename String) bool {
	ha := heap_allocator()
	bc := &buildContext

	bc.BuildPaths = make([]Path, BuildPathCOUNT)
	bc.TargetFeaturesSet = string_set_init(1024)
	bc.BuildPaths[BuildPathMainPackage] = path_from_string(ha, init_filename)

	{
		build_project_name := last_path_element(bc.BuildPaths[BuildPathMainPackage].Basename)
		bc.ODINBUILDPROJECTNAME = build_project_name
	}

	produces_output_file := false
	if bc.CommandKind == CommandDoc && bc.CmdDocFlags&CmdDocFlagDocFormat != 0 {
		produces_output_file = true
	} else if bc.CommandKind&CommandDoesBuild != 0 {
		produces_output_file = true
	}
	if !produces_output_file {
		return true
	}

	if bc.Metrics.Os == TargetOsWindows {
		if bc.ResourceFilepath.Len > 0 {
			bc.BuildPaths[BuildPathRES] = path_from_string(ha, bc.ResourceFilepath)
			if !string_ends_with(bc.ResourceFilepath, S(".res")) {
				bc.BuildPaths[BuildPathRES].Ext = copy_string(ha, S("res"))
				bc.BuildPaths[BuildPathRC] = path_from_string(ha, bc.ResourceFilepath)
				bc.BuildPaths[BuildPathRC].Ext = copy_string(ha, S("rc"))
			}
		}
		if (bc.CommandKind&CommandDoesBuild) != 0 && !bc.IgnoreMicrosoftMagic {
			find_result := find_visual_studio_and_windows_sdk()
			if find_result.WindowsSDKVersion == 0 {
				gb_printf_err("Windows SDK not found.\n")
				return false
			}
			if bc.LinkerChoice == LinkerDefault && find_result.VSExePath.Len == 0 {
				gb_printf_err("link.exe not found.\n")
				return false
			}
			if find_result.VSLibraryPath.Len == 0 {
				gb_printf_err("VS library path not found.\n")
				return false
			}
			if find_result.WindowsSDKUMLibraryPath.Len > 0 {
				if find_result.WindowsSDKBinPath.Len > 0 {
					bc.BuildPaths[BuildPathWinSDKBinPath] = path_from_string(ha, find_result.WindowsSDKBinPath)
				}
				if find_result.WindowsSDKUMLibraryPath.Len > 0 {
					bc.BuildPaths[BuildPathWinSDKUMLib] = path_from_string(ha, find_result.WindowsSDKUMLibraryPath)
				}
				if find_result.WindowsSDKUCRTLibraryPath.Len > 0 {
					bc.BuildPaths[BuildPathWinSDKUCRTLib] = path_from_string(ha, find_result.WindowsSDKUCRTLibraryPath)
				}
				if find_result.VSExePath.Len > 0 {
					bc.BuildPaths[BuildPathVSEXE] = path_from_string(ha, find_result.VSExePath)
				}
				if find_result.VSLibraryPath.Len > 0 {
					bc.BuildPaths[BuildPathVSLIB] = path_from_string(ha, find_result.VSLibraryPath)
				}
			}
		}
	}

	var output_extension String
	if bc.CommandKind == CommandDoc && bc.CmdDocFlags&CmdDocFlagDocFormat != 0 {
		output_extension = S("odin-doc")
	} else if is_arch_wasm() {
		output_extension = S("wasm")
	} else if bc.BuildMode == BuildModeExecutable {
		output_extension = String{}
		single_file_extension := S(".odin")
		if selectedSubtarget == SubtargetAndroid {
			output_extension = S("so")
		} else if bc.Metrics.Os == TargetOsWindows {
			output_extension = S("exe")
		} else if path_is_directory(last_path_element(bc.BuildPaths[BuildPathMainPackage].Basename)) {
			output_extension = S("bin")
		} else if string_ends_with(init_filename, single_file_extension) && path_is_directory(remove_extension_from_path(init_filename)) {
			output_extension = S("bin")
		}
	} else if bc.BuildMode == BuildModeDynamicLibrary {
		output_extension = S("so")
		if bc.Metrics.Os == TargetOsWindows {
			output_extension = S("dll")
		} else if bc.Metrics.Os == TargetOsDarwin {
			output_extension = S("dylib")
		}
	} else if bc.BuildMode == BuildModeStaticLibrary {
		output_extension = S("a")
		if bc.Metrics.Os == TargetOsWindows {
			output_extension = S("lib")
		}
	} else if bc.BuildMode == BuildModeObject {
		output_extension = infer_object_extension_from_build_context()
	} else if bc.BuildMode == BuildModeAssembly {
		output_extension = S("S")
	} else if bc.BuildMode == BuildModeLLVMIR {
		output_extension = S("ll")
	}

	if bc.OutFilepath.Len > 0 {
		bc.BuildPaths[BuildPathOutput] = path_from_string(ha, bc.OutFilepath)
		if bc.Metrics.Os == TargetOsWindows {
			output_file := path_to_string(ha, bc.BuildPaths[BuildPathOutput])
			if path_is_directory(bc.BuildPaths[BuildPathOutput]) {
				gb_printf_err("Output path %s is a directory.\n", goStr(output_file))
				return false
			} else if bc.BuildPaths[BuildPathOutput].Ext.Len == 0 {
				gb_printf_err("Output path %s must have an appropriate extension.\n", goStr(output_file))
				return false
			}
		}
	} else {
		var output_path Path
		if str_eq(init_filename, S(".")) {
			debugf("Output name will be created from current base name %s.\n", goStr(bc.BuildPaths[BuildPathMainPackage].Basename))
			last_element := last_path_element(bc.BuildPaths[BuildPathMainPackage].Basename)
			if last_element.Len == 0 {
				gb_printf_err("The output name is created from the last path element. `%s` has none. Use `-out:output_name.ext` to set it.\n", goStr(bc.BuildPaths[BuildPathMainPackage].Basename))
				return false
			}
			output_path.Basename = copy_string(ha, bc.BuildPaths[BuildPathMainPackage].Basename)
			output_path.Name = copy_string(ha, last_element)
		} else {
			output_name := init_filename
			for output_name.Len > 0 {
				c := *(*byte)(unsafe.Add(unsafe.Pointer(output_name.Data), output_name.Len-1))
				if c == '/' || c == '\\' {
					output_name.Len--
				} else {
					break
				}
			}
			if str_eq(path_extension(output_name), S(".odin")) && !path_is_directory(output_name) {
				output_name = remove_extension_from_path(output_name)
			}
			output_name = remove_directory_from_path(output_name)
			output_name = copy_string(ha, string_trim_whitespace(output_name))

			var res Path
			if output_name.Len > 0 {
				fullpath := path_to_fullpath(ha, output_name, nil)
				res.Basename = directory_from_path(fullpath)
				res.Basename = copy_string(ha, res.Basename)
				if path_is_directory(fullpath) {
					if res.Basename.Len > 0 {
						c := *(*byte)(unsafe.Add(unsafe.Pointer(res.Basename.Data), res.Basename.Len-1))
						if c == '/' {
							res.Basename.Len--
						}
					}
				} else {
					name_start := isize(0)
					if res.Basename.Len > 0 {
						name_start = res.Basename.Len + 1
					}
					res.Name = substring(fullpath, name_start, fullpath.Len)
					res.Name = copy_string(ha, res.Name)
				}
			}
			output_path = res
			if output_path.Name.Len == 0 {
				l := output_path.Basename.Len
				for l > 1 {
					c := *(*byte)(unsafe.Add(unsafe.Pointer(output_path.Basename.Data), l-1))
					if c == '/' {
						break
					}
					l--
				}
				old_basename := output_path.Basename
				output_path.Basename.Len = l - 1
				output_path.Name = substring(old_basename, l, old_basename.Len)
				output_path.Basename = copy_string(ha, output_path.Basename)
				output_path.Name = copy_string(ha, output_path.Name)
			}
		}
		output_path.Ext = copy_string(ha, output_extension)
		bc.BuildPaths[BuildPathOutput] = output_path
	}

	if bc.ODINDEBUG {
		if bc.Metrics.Os == TargetOsWindows {
			if bc.PdbFilepath.Len > 0 {
				bc.BuildPaths[BuildPathSymbols] = path_from_string(ha, bc.PdbFilepath)
			} else {
				var symbol_path Path
				symbol_path.Basename = copy_string(ha, bc.BuildPaths[BuildPathOutput].Basename)
				symbol_path.Name = copy_string(ha, bc.BuildPaths[BuildPathOutput].Name)
				symbol_path.Ext = copy_string(ha, S("pdb"))
				bc.BuildPaths[BuildPathSymbols] = symbol_path
			}
		} else if bc.Metrics.Os == TargetOsDarwin {
			var symbol_path Path
			symbol_path.Basename = copy_string(ha, bc.BuildPaths[BuildPathOutput].Basename)
			symbol_path.Name = copy_string(ha, bc.BuildPaths[BuildPathOutput].Name)
			symbol_path.Ext = copy_string(ha, S("dSYM"))
			bc.BuildPaths[BuildPathSymbols] = symbol_path
		}
	}

	if bc.BuildPaths[BuildPathOutput].Ext.Len == 0 {
		if bc.Metrics.Os == TargetOsWindows || is_arch_wasm() || bc.BuildMode != BuildModeExecutable {
			bc.BuildPaths[BuildPathOutput].Ext = copy_string(ha, output_extension)
		}
	}

	output_file := path_to_string(ha, bc.BuildPaths[BuildPathOutput])
	if path_is_directory(bc.BuildPaths[BuildPathOutput]) {
		gb_printf_err("Output path %s is a directory.\n", goStr(output_file))
		return false
	}

	if bc.SanitizerFlags&SanitizerFlagAddress != 0 {
		switch bc.Metrics.Os {
		case TargetOsWindows, TargetOsLinux, TargetOsDarwin, TargetOsFreeBSD:
		default:
			gb_printf_err("-sanitize:address is only supported on Windows, Linux, Darwin, and FreeBSD\n")
			return false
		}
	}
	if bc.SanitizerFlags&SanitizerFlagMemory != 0 {
		switch bc.Metrics.Os {
		case TargetOsLinux, TargetOsFreeBSD:
		default:
			gb_printf_err("-sanitize:memory is only supported on Linux and FreeBSD\n")
			return false
		}
	}
	if bc.SanitizerFlags&SanitizerFlagThread != 0 {
		switch bc.Metrics.Os {
		case TargetOsLinux, TargetOsDarwin, TargetOsFreeBSD:
		default:
			gb_printf_err("-sanitize:thread is only supported on Linux, Darwin, and FreeBSD\n")
			return false
		}
	}

	no_crt_checks_failed := false
	if bc.NoCRT && !bc.ODINDEFAULTTONILALLOCATOR && !bc.ODINDEFAULTOPANICALLOCATOR {
		switch bc.Metrics.Os {
		case TargetOsLinux, TargetOsDarwin, TargetOsFreeBSD,
			TargetOsOpenBSD, TargetOsNetBSD, TargetOsHaiku:
			gb_printf_err("-no-crt on Unix systems requires either -default-to-nil-allocator or -default-to-panic-allocator to also be present, because the default allocator requires CRT\n")
			no_crt_checks_failed = true
		}
	}
	if bc.NoCRT && !bc.NoThreadLocal {
		switch bc.Metrics.Os {
		case TargetOsLinux, TargetOsDarwin, TargetOsFreeBSD,
			TargetOsOpenBSD, TargetOsNetBSD, TargetOsHaiku:
			gb_printf_err("-no-crt on Unix systems requires the -no-thread-local flag to also be present, because the TLS is inaccessible without CRT\n")
			no_crt_checks_failed = true
		}
	}
	if no_crt_checks_failed {
		return false
	}

	return true
}


