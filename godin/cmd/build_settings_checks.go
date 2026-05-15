package cmd

import (
	"unsafe"
)

func string_equals(a, b String) bool {
	return a == b
}

func odin_cpuid(leaf int, result []int) {
}

func should_use_march_native() bool {
	return false
}

func init_android_values(with_sdk bool) {
	bc := &build_context
	{
		defaultLevel := S("34")
		if !bc.MinimumOSVersionStringGiven {
			bc.MinimumOSVersionString = defaultLevel
		}
		var level BigInt
		success := false
		big_int_from_string(&level, bc.MinimumOSVersionString, &success)
		if !success {
			gb_printf_err("Warning: Invalid -minimum-os-version:%s for -subtarget:Android, defaulting to %s\n", goStr(bc.MinimumOSVersionString), goStr(defaultLevel))
			bc.MinimumOSVersionString = defaultLevel
			big_int_from_string(&level, bc.MinimumOSVersionString, &success)
			gb_assert_handler("Assertion Failure", "success", "build_settings.cpp", 1623)
		}
		newLevel := big_int_to_i64(&level)
		if newLevel >= 21 {
			bc.ODINANDROIDAPIVALUE = int(newLevel)
		} else {
			gb_printf_err("Warning: Invalid -minimum-os-version:%s for -subtarget:Android, defaulting to %s\n", goStr(bc.MinimumOSVersionString), goStr(defaultLevel))
			bc.ODINANDROIDAPIVALUE = 34
		}
	}
	bc.ODINANDROIDNDK = normalize_path(permanent_allocator(), make_string_c(gb_get_env("ODIN_ANDROID_NDK", permanent_allocator())), NIX_SEPARATOR_STRING)
	bc.ODINANDROIDNDKTOOLCHAIN = normalize_path(permanent_allocator(), make_string_c(gb_get_env("ODIN_ANDROID_NDK_TOOLCHAIN", permanent_allocator())), NIX_SEPARATOR_STRING)
	bc.ODINANDROIDSDK = normalize_path(permanent_allocator(), make_string_c(gb_get_env("ODIN_ANDROID_SDK", permanent_allocator())), NIX_SEPARATOR_STRING)
	if len(bc.ODINANDROIDSDK) == 0 {
		bc.ODINANDROIDSDK = normalize_path(permanent_allocator(),
			path_to_fullpath(permanent_allocator(), S("%LocalAppData%/Android/Sdk"), nil),
			NIX_SEPARATOR_STRING)
	}
	if len(bc.ODINANDROIDNDK) != 0 && len(bc.ODINANDROIDNDKTOOLCHAIN) == 0 {
		arch := S("x86_64")
		bc.ODINANDROIDNDKTOOLCHAIN = concatenate4_strings(temporary_allocator(),
			bc.ODINANDROIDNDK,
			S("toolchains/llvm/prebuilt/"),
			S("windows-"),
			arch)
		bc.ODINANDROIDNDKTOOLCHAIN = normalize_path(permanent_allocator(), bc.ODINANDROIDNDKTOOLCHAIN, NIX_SEPARATOR_STRING)
	}
	if len(bc.ODINANDROIDNDK) == 0 && !with_sdk {
		gb_printf_err("Error: ODIN_ANDROID_NDK not set")
		gb_exit(1)
	}
	if len(bc.ODINANDROIDNDKTOOLCHAIN) == 0 && !with_sdk {
		gb_printf_err("Error: ODIN_ANDROID_NDK not set")
		gb_exit(1)
	}
	switch bc.Metrics.Arch {
	case TargetArchArm64:
		bc.ODINANDROIDNDKTOOLCHAINLIB = S("aarch64-linux-android")
	case TargetArchArm32:
		bc.ODINANDROIDNDKTOOLCHAINLIB = S("arm-linux-androideabi")
	case TargetArchAmd64:
		bc.ODINANDROIDNDKTOOLCHAINLIB = S("x86_64-linux-android")
	case TargetArchI386:
		bc.ODINANDROIDNDKTOOLCHAINLIB = S("i686-linux-android")
	}
	buf := [32]byte{}
	gb_snprintf(buf[:], "%d/", bc.ODINANDROIDAPIVALUE)
	bc.ODINANDROIDNDKTOOLCHAINLIBVALUE = concatenate_strings(permanent_allocator(), bc.ODINANDROIDNDKTOOLCHAINLIB, make_string_c(&buf[0]))
	bc.ODINANDROIDNDKTOOLCHAINSYSROOT = concatenate_strings(permanent_allocator(), bc.ODINANDROIDNDKTOOLCHAIN, S("sysroot/"))
	if with_sdk {
		if len(bc.ODINANDROIDSDK) == 0 {
			gb_printf_err("Error: ODIN_ANDROID_SDK not set, which is required for -build-mode:executable for -subtarget:android")
			gb_exit(1)
		}
		if len(bc.AndroidKeystore) == 0 {
			gb_printf_err("Error: -android-keystore:<string> has not been set\n")
			gb_exit(1)
		}
	}
}

func has_asm_extension(path String) bool {
	ext := path_extension(path)
	return string_equals(ext, S(".asm")) || string_equals(ext, S(".s")) || string_equals(ext, S(".S"))
}

func token_pos_to_string(pos TokenPos) gbString {
	s := gb_string_make(temporary_allocator(), "")
	file := get_file_path_string(pos.FileID)
	fileStr := goStr(file)
	switch build_context.ODINERRORPOSSTYLE {
	case ErrorPosStyleUnix:
		s = gb_string_append_fmt(s, "%s:%d:%d:", fileStr, pos.Line, pos.Column)
	default:
		s = gb_string_append_fmt(s, "%s(%d:%d)", fileStr, pos.Line, pos.Column)
	}
	return s
}

func normalize_minimum_os_version_string(version String) String {
	if len(version) <= 0 {
		gb_assert_handler("Assertion Failure", "version.Len > 0", "build_settings.cpp", 1739)
	}
	normalized := gb_string_make(permanent_allocator(), "")
	granularity := 0
	it := String_Iterator{Str: version, Pos: 0}
	for {
		str := string_split_iterator(&it, '.')
		if len(str) == 0 {
			break
		}
		if granularity > 0 {
			normalized = gb_string_appendc(normalized, ".")
		}
		normalized = gb_string_append_length(normalized, unsafe.StringData(str), len(str))
		granularity++
	}
	for ; granularity < 3; granularity++ {
		normalized = gb_string_appendc(normalized, ".0")
	}
	return make_string_c(normalized)
}

func check_single_target_feature_is_valid(feature_list String, feature String) bool {
	it := String_Iterator{Str: feature_list, Pos: 0}
	for {
		str := string_split_iterator(&it, ',')
		if len(str) == 0 {
			break
		}
		if string_equals(str, feature) {
			return true
		}
	}
	return false
}

func check_target_feature_is_valid(feature String, arch TargetArchKind, invalid *String) bool {
	feature_list := target_features_list[arch]
	it := String_Iterator{Str: feature, Pos: 0}
	for {
		str := string_split_iterator(&it, ',')
		feature_str := str
		if string_starts_with(feature_str, S("+")) || string_starts_with(feature_str, S("-")) {
			feature_str = substring(feature_str, 1, len(feature_str))
			if len(feature_str) == 0 {
				if invalid != nil {
					*invalid = str
				}
				return false
			}
		}
		if len(feature_str) == 0 {
			break
		}
		if !check_single_target_feature_is_valid(feature_list, feature_str) {
			if invalid != nil {
				*invalid = str
			}
			return false
		}
	}
	return true
}

func check_target_feature_is_valid_globally(feature String, invalid *String) bool {
	it := String_Iterator{Str: feature, Pos: 0}
	for {
		str := string_split_iterator(&it, ',')
		if len(str) == 0 {
			break
		}
		valid := false
		for arch := TargetArchInvalid; arch < TargetArchCOUNT; arch++ {
			if check_target_feature_is_valid(str, arch, invalid) {
				valid = true
				break
			}
		}
		if !valid {
			if invalid != nil {
				*invalid = str
			}
			return false
		}
	}
	return true
}

func check_target_feature_is_valid_for_target_arch(feature String, invalid *String) bool {
	return check_target_feature_is_valid(feature, build_context.Metrics.Arch, invalid)
}

func check_target_feature_is_enabled(feature String, not_enabled *String) bool {
	it := String_Iterator{Str: feature, Pos: 0}
	for {
		str := string_split_iterator(&it, ',')
		feature_str := str
		want_enabled := true
		if string_starts_with(feature_str, S("+")) || string_starts_with(feature_str, S("-")) {
			want_enabled = feature_str[0] == '+'
			feature_str = substring(feature_str, 1, len(feature_str))
		}
		if len(feature_str) == 0 {
			break
		}
		plus_str := concatenate_strings(temporary_allocator(), S("+"), feature_str)
		minus_str := concatenate_strings(temporary_allocator(), S("-"), feature_str)
		has_raw := string_set_exists(&build_context.TargetFeaturesSet, feature_str)
		has_plus := string_set_exists(&build_context.TargetFeaturesSet, plus_str)
		has_minus := string_set_exists(&build_context.TargetFeaturesSet, minus_str)
		is_enabled := (has_plus || has_raw) && !has_minus
		if want_enabled != is_enabled {
			if not_enabled != nil {
				*not_enabled = str
			}
			return false
		}
	}
	return true
}

func check_target_feature_is_superset_of(superset String, of String, missing *String) bool {
	it := String_Iterator{Str: of, Pos: 0}
	for {
		str := string_split_iterator(&it, ',')
		if len(str) == 0 {
			break
		}
		if !check_single_target_feature_is_valid(superset, str) {
			if missing != nil {
				*missing = str
			}
			return false
		}
	}
	return true
}
