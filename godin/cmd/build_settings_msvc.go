package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// --- Internal directory iteration state ---

type mcFindState struct {
	entries []os.DirEntry
	pos     int
}

var mcFindStates = make(map[uintptr]*mcFindState)
var mcFindNextHandle uintptr = 1

// --- mc_wstring_to_string / mc_string_to_wstring ---

func mc_wstring_to_string(str *uint16) String {
	if str == nil {
		return String{}
	}
	s := windows.UTF16PtrToString(str)
	return copy_string(permanent_allocator(), S(s))
}

func mc_string_to_wstring(str String) String16 {
	if str.Data == nil || str.Len == 0 {
		return String16{}
	}
	u16s, err := windows.UTF16FromString(goStr(str))
	if err != nil {
		return String16{}
	}
	alloc := permanent_allocator()
	n := isize(len(u16s))
	data := (*uint16)(gb_alloc(alloc, uintptr(n)*2))
	for i := isize(0); i < n; i++ {
		*(*uint16)(unsafe.Add(unsafe.Pointer(data), i*2)) = u16s[i]
	}
	return String16{Data: data, Len: n}
}

// --- mc_concat (variadic, supports 2, 3, or 4 args) ---

func mc_concat(parts ...String) String {
	alloc := permanent_allocator()
	switch len(parts) {
	case 2:
		return concatenate_strings(alloc, parts[0], parts[1])
	case 3:
		return concatenate3_strings(alloc, parts[0], parts[1], parts[2])
	case 4:
		return concatenate4_strings(alloc, parts[0], parts[1], parts[2], parts[3])
	}
	return String{}
}

// --- mc_get_env ---

func mc_get_env(key String) String {
	val, ok := os.LookupEnv(goStr(key))
	if !ok {
		return String{}
	}
	return copy_string(permanent_allocator(), S(val))
}

// --- mc_free ---

func mc_free(str String) {
	if str.Len != 0 && str.Data != nil {
		gbFree(permanent_allocator(), unsafe.Pointer(str.Data))
	}
}

func mc_free16(str String16) {
	if str.Len != 0 && str.Data != nil {
		gbFree(permanent_allocator(), unsafe.Pointer(str.Data))
	}
}

// --- mc_find_first / mc_find_next / mc_find_close ---

func mc_find_first(wildcard String, find_data *MCFindData) uintptr {
	dir := filepath.Dir(goStr(wildcard))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ^uintptr(0)
	}
	state := &mcFindState{entries: entries, pos: 0}
	handle := mcFindNextHandle
	mcFindNextHandle++
	mcFindStates[handle] = state

	if !mc_find_next(handle, find_data) {
		delete(mcFindStates, handle)
		return ^uintptr(0)
	}
	return handle
}

func mc_find_next(handle uintptr, find_data *MCFindData) bool {
	state, ok := mcFindStates[handle]
	if !ok {
		return false
	}
	for state.pos < len(state.entries) {
		entry := state.entries[state.pos]
		state.pos++
		name := entry.Name()
		if entry.IsDir() {
			find_data.FileAttributes = 0x00000010
		} else {
			find_data.FileAttributes = 0x00000080 // FILE_ATTRIBUTE_NORMAL
		}
		find_data.Filename = S(name)
		return true
	}
	return false
}

func mc_find_close(handle uintptr) {
	delete(mcFindStates, handle)
}

// --- mc_visit_files ---

type mc_visit_proc func(short_name String, full_name String, data *VersionData)

func mc_visit_files(dir_name String, data *VersionData, proc mc_visit_proc) bool {
	wildcard_name := mc_concat(dir_name, S("*"))
	entries, err := os.ReadDir(goStr(wildcard_name))
	mc_free(wildcard_name)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		short_name := S(entry.Name())
		if entry.IsDir() && (short_name.Len == 0 || *(*byte)(short_name.Data) != '.') {
			full_name := mc_concat(dir_name, short_name)
			proc(short_name, full_name, data)
			mc_free(full_name)
		}
	}
	return true
}

// --- string_at helper ---

func str_at(s String, i isize) byte {
	return *(*byte)(unsafe.Add(unsafe.Pointer(s.Data), i))
}

func str_last(s String) byte {
	return str_at(s, s.Len-1)
}

// --- find_windows_kit_root ---

func find_windows_kit_root(key registry.Key, version String) String {
	val, _, err := key.GetStringValue(goStr(version))
	if err != nil {
		return String{}
	}
	return copy_string(permanent_allocator(), S(val))
}

// --- win10_best ---

func win10_best(short_name String, full_name String, data *VersionData) {
	if short_name.Len == 0 {
		return
	}
	s := goStr(short_name)
	parts := strings.Split(s, ".")
	if len(parts) < 4 {
		return
	}
	i0, err0 := strconv.Atoi(parts[0])
	i1, err1 := strconv.Atoi(parts[1])
	i2, err2 := strconv.Atoi(parts[2])
	i3, err3 := strconv.Atoi(parts[3])
	if err0 != nil || err1 != nil || err2 != nil || err3 != nil {
		return
	}
	if i0 < int(data.BestVersion[0]) {
		return
	} else if i0 == int(data.BestVersion[0]) {
		if i1 < int(data.BestVersion[1]) {
			return
		} else if i1 == int(data.BestVersion[1]) {
			if i2 < int(data.BestVersion[2]) {
				return
			} else if i2 == int(data.BestVersion[2]) {
				if i3 < int(data.BestVersion[3]) {
					return
				}
			}
		}
	}
	if data.BestName.Len != 0 {
		mc_free(data.BestName)
	}
	data.BestName = copy_string(permanent_allocator(), full_name)
	if data.BestName.Len != 0 {
		data.BestVersion[0] = int32(i0)
		data.BestVersion[1] = int32(i1)
		data.BestVersion[2] = int32(i2)
		data.BestVersion[3] = int32(i3)
	}
}

// --- find_windows_kit_paths ---

func find_windows_kit_paths(result *FindResult) {
	sdk_found := false

	main_key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows Kits\Installed Roots`, registry.READ)
	if err != nil {
		return
	}
	defer main_key.Close()

	windows10_root := find_windows_kit_root(main_key, S("KitsRoot10"))
	if windows10_root.Len != 0 {
		windows10_lib := mc_concat(windows10_root, S("Lib\\"))
		var data_lib VersionData
		mc_visit_files(windows10_lib, &data_lib, win10_best)

		windows10_bin := mc_concat(windows10_root, S("bin\\"))
		var data_bin VersionData
		mc_visit_files(windows10_bin, &data_bin, win10_best)

		if data_lib.BestName.Len != 0 && data_bin.BestName.Len != 0 {
			if buildContext.Metrics.Arch == TargetArchAmd64 {
				result.WindowsSDKUMLibraryPath = mc_concat(data_lib.BestName, S("\\um\\x64\\"))
				result.WindowsSDKUCRTLibraryPath = mc_concat(data_lib.BestName, S("\\ucrt\\x64\\"))
				result.WindowsSDKBinPath = mc_concat(data_bin.BestName, S("\\x64\\"))
				sdk_found = true
			} else if buildContext.Metrics.Arch == TargetArchI386 {
				result.WindowsSDKUMLibraryPath = mc_concat(data_lib.BestName, S("\\um\\x86\\"))
				result.WindowsSDKUCRTLibraryPath = mc_concat(data_lib.BestName, S("\\ucrt\\x86\\"))
				result.WindowsSDKBinPath = mc_concat(data_bin.BestName, S("\\x86\\"))
				sdk_found = true
			}
		}

		mc_free(data_bin.BestName)
		mc_free(windows10_bin)
		mc_free(data_lib.BestName)
		mc_free(windows10_lib)
		mc_free(windows10_root)
	}

	if sdk_found {
		result.WindowsSDKVersion = 10
	}
}

// --- find_visual_studio_by_fighting_through_microsoft_craziness ---

func find_visual_studio_by_fighting_through_microsoft_craziness(result *FindResult) bool {
	vs7_key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\VisualStudio\SxS\VS7`, registry.READ)
	if err != nil {
		return false
	}
	defer vs7_key.Close()

	versions := []string{"14.0", "13.0", "12.0", "11.0", "10.0", "9.0"}
	for _, v := range versions {
		val, _, err := vs7_key.GetStringValue(v)
		if err != nil {
			continue
		}
		base_path := copy_string(permanent_allocator(), S(val))

		lib_path := String{}
		if buildContext.Metrics.Arch == TargetArchAmd64 {
			lib_path = mc_concat(base_path, S("VC\\Lib\\amd64\\"))
		} else if buildContext.Metrics.Arch == TargetArchI386 {
			lib_path = mc_concat(base_path, S("VC\\Lib\\"))
		} else {
			mc_free(base_path)
			continue
		}

		vcruntime_filename := mc_concat(lib_path, S("vcruntime.lib"))
		exe_path := String{}
		vs_found := false
		if gb_file_exists(vcruntime_filename) {
			if buildContext.Metrics.Arch == TargetArchAmd64 {
				exe_path = mc_concat(base_path, S("VC\\bin\\"))
			} else if buildContext.Metrics.Arch == TargetArchI386 {
				exe_path = mc_concat(base_path, S("VC\\bin\\x86_amd64\\"))
			}
			result.VSExePath = exe_path
			result.VSLibraryPath = lib_path
			vs_found = true
		}
		mc_free(vcruntime_filename)
		mc_free(base_path)
		if vs_found {
			return true
		}
		mc_free(lib_path)
	}
	return false
}

// --- find_windows_kit_paths_from_env_vars ---

func find_windows_kit_paths_from_env_vars(result *FindResult) {
	if buildContext.Metrics.Arch != TargetArchAmd64 && buildContext.Metrics.Arch != TargetArchI386 {
		return
	}

	sdk_lib_found := false
	sdk_bin_found := false

	win_sdk_ver_env := mc_get_env(S("WindowsSDKVersion"))
	win_sdk_lib_ver_env := mc_get_env(S("WindowsSDKLibVersion"))
	win_sdk_dir_env := mc_get_env(S("WindowsSdkDir"))
	crt_sdk_dir_env := mc_get_env(S("UniversalCRTSdkDir"))
	win_sdk_bin_path_env := mc_get_env(S("WindowsSdkBinPath"))
	win_sdk_ver_bin_path_env := mc_get_env(S("WindowsSdkVerBinPath"))

	// --- Bin path ---
	if win_sdk_ver_bin_path_env.Len != 0 ||
		((win_sdk_bin_path_env.Len != 0 || win_sdk_dir_env.Len != 0 || crt_sdk_dir_env.Len != 0) &&
			(win_sdk_ver_env.Len != 0 || win_sdk_lib_ver_env.Len != 0)) {

		bin := String{}
		if win_sdk_ver_bin_path_env.Len != 0 {
			dir := win_sdk_ver_bin_path_env
			if str_last(dir) != '\\' {
				bin = mc_concat(dir, S("\\"))
			} else {
				bin = mc_concat(dir, S(""))
			}
		} else {
			dir := String{}
			if win_sdk_bin_path_env.Len != 0 {
				dir = win_sdk_bin_path_env
			} else if win_sdk_dir_env.Len != 0 {
				dir = win_sdk_dir_env
			} else {
				dir = crt_sdk_dir_env
			}
			ver := String{}
			if win_sdk_ver_env.Len != 0 {
				ver = win_sdk_ver_env
			} else {
				ver = win_sdk_lib_ver_env
			}

			dir_tmp := String{}
			if str_last(dir) != '\\' {
				dir_tmp = mc_concat(dir, S("\\"))
			} else {
				dir_tmp = mc_concat(dir, S(""))
			}
			ver_tmp := String{}
			if str_last(ver) != '\\' {
				ver_tmp = mc_concat(ver, S("\\"))
			} else {
				ver_tmp = mc_concat(ver, S(""))
			}

			dir_bin := String{}
			if win_sdk_bin_path_env.Len != 0 {
				dir_bin = mc_concat(dir_tmp, S(""))
			} else {
				dir_bin = mc_concat(dir_tmp, S("bin\\"))
			}
			bin = mc_concat(dir_bin, ver_tmp)

			mc_free(dir_bin)
			mc_free(ver_tmp)
			mc_free(dir_tmp)
		}

		if buildContext.Metrics.Arch == TargetArchAmd64 {
			result.WindowsSDKBinPath = mc_concat(bin, S("x64\\"))
			sdk_bin_found = true
		} else if buildContext.Metrics.Arch == TargetArchI386 {
			result.WindowsSDKBinPath = mc_concat(bin, S("x86\\"))
			sdk_bin_found = true
		}
		mc_free(bin)
	}

	// --- Lib path ---
	if (win_sdk_ver_env.Len != 0 || win_sdk_lib_ver_env.Len != 0) &&
		(win_sdk_dir_env.Len != 0 || crt_sdk_dir_env.Len != 0) {

		dir := String{}
		if win_sdk_dir_env.Len != 0 {
			dir = win_sdk_dir_env
		} else {
			dir = crt_sdk_dir_env
		}
		ver := String{}
		if win_sdk_ver_env.Len != 0 {
			ver = win_sdk_ver_env
		} else {
			ver = win_sdk_lib_ver_env
		}

		dir_tmp := String{}
		if str_last(dir) != '\\' {
			dir_tmp = mc_concat(dir, S("\\"))
		} else {
			dir_tmp = mc_concat(dir, S(""))
		}
		ver_tmp := String{}
		if str_last(ver) != '\\' {
			ver_tmp = mc_concat(ver, S("\\"))
		} else {
			ver_tmp = mc_concat(ver, S(""))
		}

		if buildContext.Metrics.Arch == TargetArchAmd64 {
			result.WindowsSDKUMLibraryPath = mc_concat(dir_tmp, S("Lib\\"), ver_tmp, S("um\\x64\\"))
			result.WindowsSDKUCRTLibraryPath = mc_concat(dir_tmp, S("Lib\\"), ver_tmp, S("ucrt\\x64\\"))
			sdk_lib_found = true
		} else if buildContext.Metrics.Arch == TargetArchI386 {
			result.WindowsSDKUMLibraryPath = mc_concat(dir_tmp, S("Lib\\"), ver_tmp, S("um\\x86\\"))
			result.WindowsSDKUCRTLibraryPath = mc_concat(dir_tmp, S("Lib\\"), ver_tmp, S("ucrt\\x86\\"))
			sdk_lib_found = true
		}

		mc_free(ver_tmp)
		mc_free(dir_tmp)
	}

	// --- LIB env var fallback ---
	if !sdk_lib_found {
		lib := mc_get_env(S("LIB"))
		if lib.Len != 0 {
			um_dir := S("um\\x64")
			ucrt_dir := S("ucrt\\x64")
			if buildContext.Metrics.Arch == TargetArchI386 {
				um_dir = S("um\\x86")
				ucrt_dir = S("ucrt\\x86")
			}
			lo := isize(0)
			hi := isize(0)
			for c := isize(0); c <= lib.Len; c++ {
				if c != lib.Len && str_at(lib, c) != ';' {
					continue
				}
				hi = c
				if lo == hi {
					lo = hi + 1
					continue
				}
				dir := substring(lib, lo, hi)
				end := String{}
				if str_last(dir) == '\\' {
					end = substring(dir, 0, dir.Len-1)
				} else {
					end = substring(dir, 0, dir.Len)
				}
				if string_ends_with(end, um_dir) {
					result.WindowsSDKUMLibraryPath = mc_concat(end, S("\\"))
				} else if string_ends_with(end, ucrt_dir) {
					result.WindowsSDKUCRTLibraryPath = mc_concat(end, S("\\"))
				}
				if result.WindowsSDKUMLibraryPath.Len != 0 && result.WindowsSDKUCRTLibraryPath.Len != 0 {
					sdk_lib_found = true
					break
				}
				lo = hi + 1
			}
		}
		mc_free(lib)
	}

	// --- Cleanup env strings ---
	mc_free(win_sdk_ver_bin_path_env)
	mc_free(win_sdk_bin_path_env)
	mc_free(crt_sdk_dir_env)
	mc_free(win_sdk_dir_env)
	mc_free(win_sdk_lib_ver_env)
	mc_free(win_sdk_ver_env)

	if sdk_bin_found && sdk_lib_found {
		result.WindowsSDKVersion = 10
	}
}

// --- find_visual_studio_paths_from_env_vars ---

func find_visual_studio_paths_from_env_vars(result *FindResult) {
	if buildContext.Metrics.Arch != TargetArchAmd64 && buildContext.Metrics.Arch != TargetArchI386 {
		return
	}

	vs_found := false
	vctid := mc_get_env(S("VCToolsInstallDir"))
	if vctid.Len != 0 {
		exe := S("bin\\Hostx64\\x64\\")
		lib := S("lib\\x64\\")
		if buildContext.Metrics.Arch == TargetArchI386 {
			exe = S("bin\\Hostx86\\x86\\")
			lib = S("lib\\x86\\")
		}
		if str_last(vctid) == '\\' {
			result.VSExePath = mc_concat(vctid, exe)
			result.VSLibraryPath = mc_concat(vctid, lib)
		} else {
			result.VSExePath = mc_concat(vctid, S("\\"), exe)
			result.VSLibraryPath = mc_concat(vctid, S("\\"), lib)
		}
		vs_found = true
	}
	mc_free(vctid)

	if !vs_found {
		path := mc_get_env(S("Path"))
		if path.Len != 0 {
			exe := S("bin\\Hostx64\\x64")
			exe2 := S("bin\\HostX64\\x64")
			lib := S("lib\\x64")
			if buildContext.Metrics.Arch == TargetArchI386 {
				exe = S("bin\\Hostx86\\x86")
				exe2 = S("bin\\HostX86\\x86")
				lib = S("lib\\x86")
			}

			lo := isize(0)
			hi := isize(0)
			for c := isize(0); c <= path.Len; c++ {
				if c != path.Len && str_at(path, c) != ';' {
					continue
				}
				hi = c
				if lo == hi {
					lo = hi + 1
					continue
				}
				dir := substring(path, lo, hi)
				end := String{}
				if str_last(dir) == '\\' {
					end = substring(dir, 0, dir.Len-1)
				} else {
					end = substring(dir, 0, dir.Len)
				}
				cl := mc_concat(end, S("\\cl.exe"))
				link := mc_concat(end, S("\\link.exe"))

				if !string_ends_with(end, exe) && !string_ends_with(end, exe2) {
					mc_free(cl)
					mc_free(link)
					lo = hi + 1
					continue
				}
				if !gb_file_exists(cl) || !gb_file_exists(link) {
					mc_free(cl)
					mc_free(link)
					lo = hi + 1
					continue
				}

				root := substring(end, 0, end.Len-exe.Len)
				result.VSExePath = mc_concat(end, S("\\"))
				result.VSLibraryPath = mc_concat(root, lib, S("\\"))
				vs_found = true
				mc_free(cl)
				mc_free(link)
				break
			}
		}
		mc_free(path)
	}
}

// --- find_visual_studio_and_windows_sdk ---

func find_visual_studio_and_windows_sdk() FindResult {
	var r FindResult
	find_windows_kit_paths(&r)
	find_visual_studio_by_fighting_through_microsoft_craziness(&r)

	sdk_found :=
		r.WindowsSDKBinPath.Len != 0 &&
			r.WindowsSDKUMLibraryPath.Len != 0 &&
			r.WindowsSDKUCRTLibraryPath.Len != 0

	vs_found :=
		r.VSExePath.Len != 0 &&
			r.VSLibraryPath.Len != 0

	if !sdk_found {
		find_windows_kit_paths_from_env_vars(&r)
	}
	if !vs_found {
		find_visual_studio_paths_from_env_vars(&r)
	}
	return r
}
