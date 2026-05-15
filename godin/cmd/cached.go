// Depends on: common.odin, cmd_build_settings_types.go, cmd_parser_types.go, cmd_tokenizer_core.go
// (build_context, BuildPath_Output, BuildPath_RC, BuildPath_RES, selected_target_metrics,
//
//	selected_subtarget, subtarget_strings, Checker, AstPackage, AstFile, Parser, LoadedFile,
//	LoadedFileError, LoadedFile_Empty, load_file_32, alloc_cstring, gbFile, gbFileError,
//	gbFileError_None, gb_file_open_mode, gbFileMode_Write, gb_file_create, gb_file_close,
//	gb_file_remove, gb_file_copy, gb_file_exists, gb_file_last_write_time, gbFileTime,
//	gb_fprintf, gb_printf_err, gb_printf, gb_string_make, gb_string_make_reserve, gb_string_free,
//	gb_string_append_length, gb_string_appendc, gb_string_append_fmt, gb_string_length, gbString,
//	heap_allocator, permanent_allocator, temporary_allocator,
//	string_to_string16, make_string16_c, string16_to_string, alloc_wstring,
//	path_to_string, concatenate_strings, concatenate3_strings, substring, string_trim_whitespace,
//	string_starts_with, string_compare, string_split_iterator, string_index_byte,
//	exact_value_to_u64, exact_value_integer_from_string,
//	array_make, array_add, array_sort, array_free, Array,
//	String, String16, isize, i64, u32, u8, u16, isize,
//	debugf, gb__defer_func, FileInfo,
//	String_Iterator, ReadDirectoryError, ReadDirectory_*)
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func string_cmp(a, b String) int {
	return string_compare(a, b)
}

func recursively_delete_directory_w(wpath_c *uint16) bool {
	is_dots_w := func(str *uint16) bool {
		if str == nil {
			return false
		}
		s := windows.UTF16PtrToString(str)
		return s == "." || s == ".."
	}

	var dirPath [260]uint16
	var filename [260]uint16
	wcscpy(&dirPath, wpath_c)
	wcscat(&dirPath, windows.StringToUTF16Ptr("\\*"))
	wcscpy(&filename, wpath_c)
	wcscat(&filename, windows.StringToUTF16Ptr("\\"))

	findFileData := &windows.Win32finddata{}
	hFind, err := windows.FindFirstFile(windows.UTF16ToString(dirPath[:]), findFileData)
	if hFind == windows.InvalidHandle {
		return false
	}
	defer windows.FindClose(hFind)
	wcscpy(&dirPath, &filename)

	for {
		err = windows.FindNextFile(hFind, findFileData)
		if err != nil {
			if err == windows.ERROR_NO_MORE_FILES {
				break
			}
			return false
		}
		if is_dots_w(&findFileData.FileName[0]) {
			continue
		}
		name := windows.UTF16ToString(findFileData.FileName[:])
		filenameCopy := windows.UTF16ToString(filename[:])
		fullPath := filepath.Join(filenameCopy, name)
		fullPathW := windows.StringToUTF16Ptr(fullPath)

		if findFileData.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
			if !recursively_delete_directory_w(fullPathW) {
				return false
			}
			windows.RemoveDirectory(fullPathW)
		} else {
			if findFileData.FileAttributes&windows.FILE_ATTRIBUTE_READONLY != 0 {
				os.Chmod(fullPath, 0200)
			}
			if err := windows.DeleteFile(fullPathW); err != nil {
				return false
			}
		}
	}
	return windows.RemoveDirectory(wpath_c) == nil
}

func wcscpy(dst *[260]uint16, src *uint16) {
	i := 0
	for src[i] != 0 {
		dst[i] = src[i]
		i++
	}
	dst[i] = 0
}

func wcscat(dst *[260]uint16, src *uint16) {
	i := 0
	for dst[i] != 0 {
		i++
	}
	j := 0
	for src[j] != 0 {
		dst[i] = src[j]
		i++
		j++
	}
	dst[i] = 0
}

func recursively_delete_directory(path String) bool {
	wpath := string_to_string16(permanent_allocator(), path)
	wpath_c := alloc_wstring(permanent_allocator(), wpath)
	return recursively_delete_directory_w(wpath_c)
}

func try_clear_cache() bool {
	return recursively_delete_directory(".odin-cache")
}

var gb_crc64_table [256]uint64

func crc64_with_seed(data unsafe.Pointer, length isize, seed uint64) uint64 {
	result := ^seed
	c := (*[1 << 30]uint8)(data)[:length]
	for _, b := range c {
		result = (result >> 8) ^ gb_crc64_table[(result^uint64(b))&0xff]
	}
	return ^result
}

func check_if_exists_file_otherwise_create(str String) bool {
	str_c := alloc_cstring(permanent_allocator(), str)
	if !gb_file_exists(str_c) {
		f, err := os.Create(str_c)
		if err != nil {
			return false
		}
		f.Close()
		return true
	}
	return false
}

func check_if_exists_directory_otherwise_create(str String) bool {
	wstr := string_to_string16(permanent_allocator(), str)
	wstr_c := alloc_wstring(permanent_allocator(), wstr)
	return windows.CreateDirectory(wstr_c, nil) == nil
}

func try_copy_executable_cache_internal(to_cache bool) bool {
	exe_name := path_to_string(heap_allocator(), build_context.build_paths[BuildPath_Output])

	cache_name := gb_string_make(heap_allocator(), "")
	defer gb_string_free(cache_name)

	cache_dir := build_context.build_cache_data.cache_dir
	cache_name = gb_string_append_length(cache_name, unsafe.StringData(cache_dir), len(cache_dir))
	cache_name = gb_string_appendc(cache_name, "/")
	cache_name = gb_string_appendc(cache_name, "cached-exe")
	if selected_target_metrics != nil {
		cache_name = gb_string_appendc(cache_name, "-")
		cache_name = gb_string_append_length(cache_name, unsafe.StringData(selected_target_metrics.name), len(selected_target_metrics.name))
	}
	if selected_subtarget != 0 {
		st := subtarget_strings[selected_subtarget]
		cache_name = gb_string_appendc(cache_name, "-")
		cache_name = gb_string_append_length(cache_name, unsafe.StringData(st), len(st))
	}
	cache_name = gb_string_appendc(cache_name, ".bin")

	if to_cache {
		return gb_file_copy(
			alloc_cstring(temporary_allocator(), exe_name),
			cache_name,
			false,
		)
	}
	return gb_file_copy(
		cache_name,
		alloc_cstring(temporary_allocator(), exe_name),
		false,
	)
}

func try_copy_executable_to_cache() bool {
	debugf("Cache: try_copy_executable_to_cache\n")
	if try_copy_executable_cache_internal(true) {
		build_context.build_cache_data.copy_already_done = true
		return true
	}
	return false
}

func try_copy_executable_from_cache() bool {
	debugf("Cache: try_copy_executable_from_cache\n")
	if try_copy_executable_cache_internal(false) {
		build_context.build_cache_data.copy_already_done = true
		return true
	}
	return false
}

func cache_gather_files(c *Checker) []String {
	p := c.Parser
	var files []String
	for _, pkg := range p.Packages {
		for _, f := range pkg.Files {
			files = append(files, f.Fullpath)
		}
	}
	if build_context.has_resource {
		var res_path String
		if len(build_context.build_paths[BuildPath_RC].Basename) == 0 {
			res_path = path_to_string(permanent_allocator(), build_context.build_paths[BuildPath_RES])
		} else {
			res_path = path_to_string(permanent_allocator(), build_context.build_paths[BuildPath_RC])
		}
		files = append(files, res_path)
	}
	for _, entry := range c.Info.LoadFileCache {
		cache := entry.Value
		if cache == nil || !cache.Exists {
			continue
		}
		files = append(files, cache.Path)
	}
	sort.Slice(files, func(i, j int) bool {
		return string_compare(files[i], files[j]) < 0
	})
	return files
}

func cache_gather_envs() []String {
	var envs []String
	envBlock := windows.GetEnvironmentStrings()
	if envBlock == nil {
		return envs
	}
	defer windows.FreeEnvironmentStrings(envBlock)

	for p := envBlock; *p != 0; {
		wstr := make_string16_c((*uint16)(unsafe.Pointer(p)))
		p += uintptr(len(wstr) + 1)
		str := string16_to_string(temporary_allocator(), wstr)
		if strings.HasPrefix(
			str,
			"CURR_DATE_TIME=",
		) {
			continue
		}
		envs = append(envs, str)
	}
	sort.Slice(envs, func(i, j int) bool {
		return string_compare(envs[i], envs[j]) < 0
	})
	return envs
}

func try_cached_build(c *Checker, args []String) bool {
	files := cache_gather_files(c)
	envs := cache_gather_envs()
	defer array_free(&envs)

	var crc uint64 = 0
	for _, path := range files {
		crc = crc64_with_seed(unsafe.Pointer(unsafe.StringData(path)), len(path), crc)
	}

	base_cache_dir := build_context.build_paths[BuildPath_Output].Basename
	base_cache_dir = concatenate_strings(permanent_allocator(), base_cache_dir, "/.odin-cache")
	check_if_exists_directory_otherwise_create(base_cache_dir)

	crc_str := fmt.Sprintf("%016x", crc)
	cache_dir := concatenate3_strings(permanent_allocator(), base_cache_dir, "/", crc_str)
	files_path := concatenate3_strings(permanent_allocator(), cache_dir, "/", "files.manifest")
	args_path := concatenate3_strings(permanent_allocator(), cache_dir, "/", "args.manifest")
	env_path := concatenate3_strings(permanent_allocator(), cache_dir, "/", "env.manifest")

	build_context.build_cache_data.cache_dir = cache_dir
	build_context.build_cache_data.files_path = files_path
	build_context.build_cache_data.args_path = args_path
	build_context.build_cache_data.env_path = env_path

	if check_if_exists_directory_otherwise_create(cache_dir) {
		return false
	}
	if check_if_exists_file_otherwise_create(files_path) {
		return false
	}
	if check_if_exists_file_otherwise_create(args_path) {
		return false
	}
	if check_if_exists_file_otherwise_create(env_path) {
		return false
	}

	{
		loaded_file := LoadedFile{}
		file_err := load_file_32(alloc_cstring(temporary_allocator(), files_path), &loaded_file, true)
		if file_err > LoadedFile_Empty {
			return false
		}
		data := unsafe.String((*byte)(loaded_file.Data), loaded_file.Size)
		it := String_Iterator{Str: data, Pos: 0}
		file_count := isize(0)
		for ; it.Pos < len(data); file_count++ {
			line := string_split_iterator(&it, '\n')
			if len(line) == 0 {
				break
			}
			sep := string_index_byte(line, ' ')
			if sep < 0 {
				return false
			}
			timestamp_str := substring(line, 0, sep)
			path_str := substring(line, sep+1, len(line))
			timestamp_str = string_trim_whitespace(timestamp_str)
			path_str = string_trim_whitespace(path_str)
			if int(file_count) >= len(files) {
				return false
			}
			if files[file_count] != path_str {
				return false
			}
			timestamp := exact_value_to_u64(exact_value_integer_from_string(timestamp_str))
			last_write_time := gb_file_last_write_time(alloc_cstring(temporary_allocator(), path_str))
			if last_write_time != timestamp {
				return false
			}
		}
		if int(file_count) != len(files) {
			return false
		}
	}

	{
		loaded_file := LoadedFile{}
		file_err := load_file_32(alloc_cstring(temporary_allocator(), args_path), &loaded_file, true)
		if file_err > LoadedFile_Empty {
			return false
		}
		data := unsafe.String((*byte)(loaded_file.Data), loaded_file.Size)
		it := String_Iterator{Str: data, Pos: 0}
		args_count := isize(0)
		for ; it.Pos < len(data); args_count++ {
			line := string_split_iterator(&it, '\n')
			line = string_trim_whitespace(line)
			if len(line) == 0 {
				break
			}
			if int(args_count) >= len(args) {
				return false
			}
			if line != args[args_count] {
				return false
			}
		}
	}

	{
		loaded_file := LoadedFile{}
		file_err := load_file_32(alloc_cstring(temporary_allocator(), env_path), &loaded_file, true)
		if file_err > LoadedFile_Empty {
			return false
		}
		data := unsafe.String((*byte)(loaded_file.Data), loaded_file.Size)
		it := String_Iterator{Str: data, Pos: 0}
		env_count := isize(0)
		for ; it.Pos < len(data); env_count++ {
			line := string_split_iterator(&it, '\n')
			line = string_trim_whitespace(line)
			if len(line) == 0 {
				break
			}
			if int(env_count) >= len(envs) {
				return false
			}
			if line != envs[env_count] {
				return false
			}
		}
	}

	return try_copy_executable_from_cache()
}

func write_cached_build(c *Checker, args []String) {
	files := cache_gather_files(c)
	envs := cache_gather_envs()

	{
		path_c := alloc_cstring(temporary_allocator(), build_context.build_cache_data.files_path)
		gb_file_remove(path_c)
		debugf("Cache: updating %s\n", path_c)
		f, err := os.Create(path_c)
		if err == nil {
			defer f.Close()
			for _, path := range files {
				ft := gb_file_last_write_time(alloc_cstring(temporary_allocator(), path))
				fmt.Fprintf(f, "%d %s\n", ft, path)
			}
		}
	}
	{
		path_c := alloc_cstring(temporary_allocator(), build_context.build_cache_data.args_path)
		gb_file_remove(path_c)
		debugf("Cache: updating %s\n", path_c)
		f, err := os.Create(path_c)
		if err == nil {
			defer f.Close()
			for _, arg := range args {
				targ := string_trim_whitespace(arg)
				fmt.Fprintf(f, "%s\n", targ)
			}
		}
	}
	{
		path_c := alloc_cstring(temporary_allocator(), build_context.build_cache_data.env_path)
		gb_file_remove(path_c)
		debugf("Cache: updating %s\n", path_c)
		f, err := os.Create(path_c)
		if err == nil {
			defer f.Close()
			for _, env := range envs {
				fmt.Fprintf(f, "%s\n", env)
			}
		}
	}
}
