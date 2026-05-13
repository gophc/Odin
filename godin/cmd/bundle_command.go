// Depends on: common.odin, cmd_build_settings_types.go, cmd_system.go
// (build_context, Command_bundle_android, init_android_values, path_to_fullpath, normalize_path,
//  concatenate3_strings, concatenate_strings, read_directory, ReadDirectoryError,
//  ReadDirectory_InvalidPath, ReadDirectory_NotExists, ReadDirectory_Permission,
//  ReadDirectory_NotDir, ReadDirectory_Empty, ReadDirectory_Unknown,
//  FileInfo, heap_allocator, temporary_allocator, array_make, array_add, array_free, Array,
//  gb_string_make, gb_string_free, gb_string_clear, gb_string_appendc, gb_string_append_length,
//  gb_string_append_fmt, gb_bprintf, gb_printf_err, gb_file_exists, system_exec_command_line_app,
//  debugf, timings_start_section, global_timings, NIX_SEPARATOR_STRING, String, isize, i32, u8, make_string_c,
//  path_remove_extension, substring, gbFileError, gbFileError_None)
package cmd

import "fmt"

func bundle(init_directory String) i32 {
	switch build_context.CommandKind {
	case Command_bundle_android:
		return bundle_android(init_directory)
	}
	gb_printf_err("Unknown odin package <platform>\n")
	return 1
}

func bundle_android(original_init_directory String) i32 {
	var result i32 = 0
	init_android_values(true)
	init_directory_ok := false
	init_directory := path_to_fullpath(temporary_allocator(), original_init_directory, &init_directory_ok)
	if !init_directory_ok {
		gb_printf_err("Error: '%.*s' is not a valid directory", len(original_init_directory.Text), original_init_directory.Text)
		return 1
	}
	init_directory = normalize_path(temporary_allocator(), init_directory, NIX_SEPARATOR_STRING)

	ODIN_ANDROID_API_LEVEL := build_context.ODINANDROIDAPIVALUE
	android_sdk_build_tools := concatenate3_strings(temporary_allocator(),
		build_context.ODINANDROIDSDK,
		func() String { s := "build-tools"; return String{Text: s, Len: isize(len(s))} }(),
		NIX_SEPARATOR_STRING,
	)

	var list []FileInfo
	rd_err := read_directory(android_sdk_build_tools, &list)
	switch rd_err {
	case ReadDirectory_InvalidPath:
		gb_printf_err("Invalid path: %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	case ReadDirectory_NotExists:
		gb_printf_err("Path does not exist: %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	case ReadDirectory_Permission:
		gb_printf_err("Unknown error whilst reading path %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	case ReadDirectory_NotDir:
		gb_printf_err("Expected a directory for a package, got a file: %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	case ReadDirectory_Empty:
		gb_printf_err("Empty directory: %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	case ReadDirectory_Unknown:
		gb_printf_err("Unknown error whilst reading path %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	}

	var possible_valid_dirs []FileInfo
	for _, fi := range list {
		if !fi.is_dir {
			continue
		}
		all_numbers := true
		for i := isize(0); i < fi.name.Len; i++ {
			c := fi.name.Text[i]
			if '0' <= c && c <= '9' {
				continue
			}
			if i == 0 {
				all_numbers = false
				break
			}
			if c == '.' {
				break
			}
			all_numbers = false
			break
		}
		if all_numbers {
			possible_valid_dirs = append(possible_valid_dirs, fi)
		}
	}

	if len(possible_valid_dirs) == 0 {
		gb_printf_err("Unable to find any Android SDK/API Level in %.*s\n", len(android_sdk_build_tools.Text), android_sdk_build_tools.Text)
		return 1
	}

	dir_numbers := make([]int, len(possible_valid_dirs))
	for i, fi := range possible_valid_dirs {
		n := int(fi.name.Len)
		if n > 1023 {
			n = 1023
		}
		buf := make([]byte, n)
		copy(buf, fi.name.Text[:n])
		dir_numbers[i] = int(atoi(string(buf)))
	}

	closest_number_idx := -1
	for i := range possible_valid_dirs {
		if dir_numbers[i] >= ODIN_ANDROID_API_LEVEL {
			if closest_number_idx < 0 || dir_numbers[i] < dir_numbers[closest_number_idx] {
				closest_number_idx = i
			}
		}
	}
	if closest_number_idx < 0 {
		gb_printf_err("Unable to find any Android SDK/API Level in %.*s meeting the minimum API level of %d\n",
			len(android_sdk_build_tools.Text), android_sdk_build_tools.Text, ODIN_ANDROID_API_LEVEL)
		return 1
	}

	api_number := possible_valid_dirs[closest_number_idx].name
	android_sdk_build_tools = concatenate_strings(temporary_allocator(), android_sdk_build_tools, api_number)
	apiLevelStr := fmt.Sprintf("platforms/android-%d/", dir_numbers[closest_number_idx])
	android_sdk_platforms := concatenate_strings(temporary_allocator(),
		build_context.ODINANDROIDSDK,
		func() String { return String{Text: apiLevelStr, Len: isize(len(apiLevelStr))} }(),
	)

	android_sdk_build_tools = normalize_path(temporary_allocator(), android_sdk_build_tools, NIX_SEPARATOR_STRING)
	android_sdk_platforms = normalize_path(temporary_allocator(), android_sdk_platforms, NIX_SEPARATOR_STRING)

	cmd := gb_string_make(heap_allocator(), "")
	defer gb_string_free(cmd)

	output_filename := func() String { s := "test"; return String{Text: s, Len: isize(len(s))} }()
	output_apk := path_remove_extension(output_filename)

	{
		debugf("[Section] %s\n", "Android aapt")
		if build_context.show_more_timings {
			timings_start_section(&global_timings, func() String { s := "Android aapt"; return String{Text: s, Len: isize(len(s))} }())
		}
		gb_string_clear(cmd)
		manifest := concatenate_strings(temporary_allocator(), init_directory, func() String { s := "AndroidManifest.xml"; return String{Text: s, Len: isize(len(s))} }())
		cmd = gb_string_append_length(cmd, android_sdk_build_tools.Text, android_sdk_build_tools.Len)
		cmd = gb_string_appendc(cmd, "aapt")
		cmd = gb_string_appendc(cmd, " package -f")
		cmd = gb_string_append_fmt(cmd, " -M \"%.*s\"", len(manifest.Text), manifest.Text)
		cmd = gb_string_append_fmt(cmd, " -I \"%.*sandroid.jar\"", len(android_sdk_platforms.Text), android_sdk_platforms.Text)
		cmd = gb_string_append_fmt(cmd, " -F \"%.*s.apk-build\"", len(output_apk.Text), output_apk.Text)
		resources_dir := concatenate_strings(temporary_allocator(), init_directory, func() String { s := "res"; return String{Text: s, Len: isize(len(s))} }())
		if gb_file_exists(func() string { return string(resources_dir.Text[:resources_dir.Len]) }()) {
			cmd = gb_string_append_fmt(cmd, " -S \"%.*s\"", len(resources_dir.Text), resources_dir.Text)
		}
		assets_dir := concatenate_strings(temporary_allocator(), init_directory, func() String { s := "assets"; return String{Text: s, Len: isize(len(s))} }())
		if gb_file_exists(string(assets_dir.Text[:assets_dir.Len])) {
			cmd = gb_string_append_fmt(cmd, " -A \"%.*s\"", len(assets_dir.Text), assets_dir.Text)
		}
		lib_dir := concatenate_strings(temporary_allocator(), init_directory, func() String { s := "lib"; return String{Text: s, Len: isize(len(s))} }())
		if gb_file_exists(string(lib_dir.Text[:lib_dir.Len])) {
			cmd = gb_string_append_fmt(cmd, " \"%.*s\"", len(lib_dir.Text), lib_dir.Text)
		}
		result = system_exec_command_line_app("android-aapt", cmd)
		if result != 0 {
			return result
		}
	}

	{
		debugf("[Section] %s\n", "Android zipalign")
		if build_context.show_more_timings {
			timings_start_section(&global_timings, func() String { s := "Android zipalign"; return String{Text: s, Len: isize(len(s))} }())
		}
		gb_string_clear(cmd)
		cmd = gb_string_append_length(cmd, android_sdk_build_tools.Text, android_sdk_build_tools.Len)
		cmd = gb_string_appendc(cmd, "zipalign")
		cmd = gb_string_appendc(cmd, " -f 4")
		cmd = gb_string_append_fmt(cmd, " \"%.*s.apk-build\" \"%.*s.apk\"", len(output_apk.Text), output_apk.Text, len(output_apk.Text), output_apk.Text)
		result = system_exec_command_line_app("android-zipalign", cmd)
		if result != 0 {
			return result
		}
	}

	{
		debugf("[Section] %s\n", "Android apksigner")
		if build_context.show_more_timings {
			timings_start_section(&global_timings, func() String { s := "Android apksigner"; return String{Text: s, Len: isize(len(s))} }())
		}
		gb_string_clear(cmd)
		cmd = gb_string_append_length(cmd, android_sdk_build_tools.Text, android_sdk_build_tools.Len)
		cmd = gb_string_appendc(cmd, "apksigner.bat")
		cmd = gb_string_appendc(cmd, " sign")
		keystore := normalize_path(temporary_allocator(), build_context.android_keystore, NIX_SEPARATOR_STRING)
		keystore = substring(keystore, 0, keystore.Len-1)
		cmd = gb_string_append_fmt(cmd, " --ks \"%.*s\"", len(keystore.Text), keystore.Text)
		if build_context.android_keystore_alias.Len != 0 {
			cmd = gb_string_append_fmt(cmd, " --ks-key-alias \"%.*s\"", len(build_context.android_keystore_alias.Text), build_context.android_keystore_alias.Text)
		}
		if build_context.android_keystore_password.Len != 0 {
			cmd = gb_string_append_fmt(cmd, " --ks-pass pass:\"%.*s\"", len(build_context.android_keystore_password.Text), build_context.android_keystore_password.Text)
		}
		cmd = gb_string_append_fmt(cmd, " \"%.*s.apk\"", len(output_apk.Text), output_apk.Text)
		result = system_exec_command_line_app("android-apksigner", cmd)
		if result != 0 {
			return result
		}
	}

	return 0
}
