package cmd

import "sort"

func cache_load_file_directive(c *CheckerContext, call *Ast, original_string String, err_on_not_found bool, cache_ **LoadFileCache, tier LoadFileTier, use_mutex bool) bool {
	ce := call.CallExpr
	bd := ce.Proc.BasicDirective
	builtin_name := bd.Name.String

	var path String
	if gb_path_is_absolute(goStr(original_string)) {
		path = original_string
	} else {
		base_dir := dir_from_path(get_file_path_string(call.FileID))
		var ignore_mutex *BlockingMutex
		var ok bool
		ok = determine_path_from_string(ignore_mutex, call, base_dir, original_string, &path)
		if !ok {
			if err_on_not_found {
				error(ce.Proc, "Failed to `#%.*s` file: %.*s; invalid file or cannot be found", builtin_name.Len, builtin_name.Data, original_string.Len, original_string.Data)
			}
			call.StateFlags |= StateFlag_DirectiveWasFalse
			return false
		}
	}
	if use_mutex {
		mutex_lock(&c.Info.LoadFileMutex)
	}
	if use_mutex {
		defer mutex_unlock(&c.Info.LoadFileMutex)
	}

	file_error := gbFileError_None
	var data String
	exists := false
	cache_tier := LoadFileTier_Invalid
	cache_ptr := string_map_get(&c.Info.LoadFileCache, path)
	var cache *LoadFileCache
	if cache_ptr != nil {
		cache = *cache_ptr
	}
	if cache != nil {
		file_error = cache.FileError
		data = cache.Data
		exists = cache.Exists
		cache_tier = cache.Tier
	}

	defer func() {
		if cache == nil {
			new_cache := permanent_alloc_item[*LoadFileCache]()
			*new_cache = LoadFileCache{
				Path:      path,
				Data:      data,
				FileError: file_error,
				Exists:    exists,
				Tier:      cache_tier,
			}
			string_map_init(&new_cache.Hashes, 32)
			string_map_set(&c.Info.LoadFileCache, path, new_cache)
			if cache_ != nil {
				*cache_ = new_cache
			}
		} else {
			cache.Data = data
			cache.FileError = file_error
			cache.Exists = exists
			cache.Tier = cache_tier
			if cache_ != nil {
				*cache_ = cache
			}
		}
	}()

	if tier > cache_tier {
		cache_tier = tier
		c_str := alloc_cstring(temporary_allocator(), path)
		var f gbFile
		file_error = gb_file_open(&f, c_str)
		if file_error == gbFileError_None {
			exists = true
			switch tier {
			case LoadFileTier_Exists:
			case LoadFileTier_Contents:
				file_size := isize(gb_file_size(&f))
				if file_size > 0 {
					ptr := permanent_alloc_array[byte](file_size + 1)
					gb_file_read_at(&f, ptr, file_size, 0)
					ptr[file_size] = 0
					data = String{Data: ptr, Len: file_size}
				}
			default:
				gb_assert_handler("Panic", "Unhandled LoadFileTier", "cmd_check_builtin_load.go", 0)
			}
		}
		gb_file_close(&f)
	}

	switch file_error {
	default:
		fallthrough
	case gbFileError_Invalid:
		if err_on_not_found {
			error(ce.Proc, "Failed to `#%.*s` file: %.*s; invalid file or cannot be found", builtin_name.Len, builtin_name.Data, path.Len, path.Data)
		}
		call.StateFlags |= StateFlag_DirectiveWasFalse
		return false
	case gbFileError_NotExists:
		if err_on_not_found {
			error(ce.Proc, "Failed to `#%.*s` file: %.*s; file cannot be found", builtin_name.Len, builtin_name.Data, path.Len, path.Data)
		}
		call.StateFlags |= StateFlag_DirectiveWasFalse
		return false
	case gbFileError_Permission:
		if err_on_not_found {
			error(ce.Proc, "Failed to `#%.*s` file: %.*s; file permissions problem", builtin_name.Len, builtin_name.Data, path.Len, path.Data)
		}
		call.StateFlags |= StateFlag_DirectiveWasFalse
		return false
	case gbFileError_None:
	}
	return true
}

func is_valid_type_for_load(type_ *Type) bool {
	if type_ == t_invalid {
		return false
	} else if is_type_string(type_) {
		return true
	} else if is_type_slice(type_) {
		bt := base_type(type_)
		var elem *Type
		switch bt.Kind {
		case Type_Slice:
			elem = bt.Slice.Elem
		case Type_Array:
			elem = bt.Array.Elem
		case Type_EnumeratedArray:
			elem = bt.EnumeratedArray.Elem
		}
		gb_assert_handler("Assertion Failure", "elem != nil", "cmd_check_builtin_load.go", 0)
		return is_type_load_safe(elem)
	}
	return false
}

func check_load_directive(c *CheckerContext, operand *Operand, call *Ast, type_hint *Type, err_on_not_found bool) LoadDirectiveResult {
	ce := call.CallExpr
	bd := ce.Proc.BasicDirective
	name := bd.Name.String

	gb_assert_handler("Assertion Failure", "name == \"load\"", "cmd_check_builtin_load.go", 0)

	if len(ce.Args) != 1 && len(ce.Args) != 2 {
		if len(ce.Args) == 0 {
			error(ce.Close, "'#%.*s' expects 1 or 2 arguments, got 0", name.Len, name.Data)
		} else {
			error(ce.Args[0], "'#%.*s' expects 1 or 2 arguments, got %d", name.Len, name.Data, len(ce.Args))
		}
		return LoadDirective_Error
	}
	arg := ce.Args[0]
	var o Operand
	check_expr(c, &o, arg)
	if o.Mode != Addressing_Constant {
		error(arg, "'#%.*s' expected a constant string argument", name.Len, name.Data)
		return LoadDirective_Error
	}
	if !is_type_string(o.Type) {
		str := type_to_string(o.Type)
		error(arg, "'#%.*s' expected a constant string, got %s", name.Len, name.Data, str)
		gb_string_free(str)
		return LoadDirective_Error
	}
	gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String", "cmd_check_builtin_load.go", 0)

	operand.Type = t_u8_slice
	if len(ce.Args) == 1 {
		if type_hint != nil && is_valid_type_for_load(type_hint) {
			operand.Type = type_hint
		}
	} else if len(ce.Args) == 2 {
		arg_type := ce.Args[1]
		type_ := check_type(c, arg_type)
		if type_ != nil {
			if is_valid_type_for_load(type_) {
				operand.Type = type_
			} else {
				type_str := type_to_string(type_)
				error(arg_type, "'#%.*s' invalid type, expected a string, or slice of simple types, got %s", name.Len, name.Data, type_str)
				gb_string_free(type_str)
			}
		}
	} else {
		gb_assert_handler("Panic", "unreachable", "cmd_check_builtin_load.go", 0)
	}
	operand.Mode = Addressing_Constant
	var cache *LoadFileCache
	if cache_load_file_directive(c, call, o.Value.ValueString, err_on_not_found, &cache, LoadFileTier_Contents, true) {
		operand.Value = exact_value_string(cache.Data)
		return LoadDirective_Success
	}
	return LoadDirective_NotFound
}

func file_cache_sort_cmp(x, y *LoadFileCache) int {
	if x == y {
		return 0
	}
	return string_compare(x.Path, y.Path)
}

func check_load_directory_directive(c *CheckerContext, operand *Operand, call *Ast, type_hint *Type, err_on_not_found bool) LoadDirectiveResult {
	ce := call.CallExpr
	bd := ce.Proc.BasicDirective
	name := bd.Name.String

	gb_assert_handler("Assertion Failure", "name == \"load_directory\"", "cmd_check_builtin_load.go", 0)

	if len(ce.Args) != 1 {
		error(ce.Args[0], "'#%.*s' expects 1 argument, got %d", name.Len, name.Data, len(ce.Args))
		return LoadDirective_Error
	}
	arg := ce.Args[0]
	var o Operand
	check_expr(c, &o, arg)
	if o.Mode != Addressing_Constant {
		error(arg, "'#%.*s' expected a constant string argument", name.Len, name.Data)
		return LoadDirective_Error
	}
	if !is_type_string(o.Type) {
		str := type_to_string(o.Type)
		error(arg, "'#%.*s' expected a constant string, got %s", name.Len, name.Data, str)
		gb_string_free(str)
		return LoadDirective_Error
	}
	gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String", "cmd_check_builtin_load.go", 0)

	init_core_load_directory_file(c.Checker)
	operand.Type = t_load_directory_file_slice
	operand.Mode = Addressing_Value
	original_string := o.Value.ValueString

	var path String
	if gb_path_is_absolute(goStr(original_string)) {
		path = original_string
	} else {
		base_dir := dir_from_path(get_file_path_string(call.FileID))
		var ignore_mutex *BlockingMutex
		determine_path_from_string(ignore_mutex, call, base_dir, original_string, &path)
	}

	mutex_lock(&c.Info.LoadDirectoryMutex)
	defer mutex_unlock(&c.Info.LoadDirectoryMutex)

	file_error := gbFileError_None
	var file_caches []*LoadFileCache
	cache_ptr := string_map_get(&c.Info.LoadDirectoryCache, path)
	var cache *LoadDirectoryCache
	if cache_ptr != nil {
		cache = *cache_ptr
	}
	if cache != nil {
		file_error = cache.FileError
	}

	defer func() {
		if cache == nil {
			new_cache := permanent_alloc_item[*LoadDirectoryCache]()
			*new_cache = LoadDirectoryCache{
				Path:      path,
				Files:     file_caches,
				FileError: file_error,
			}
			string_map_set(&c.Info.LoadDirectoryCache, path, new_cache)
			map_set(&c.Info.LoadDirectoryMap, call, new_cache)
		} else {
			cache.FileError = file_error
			map_set(&c.Info.LoadDirectoryMap, call, cache)
		}
	}()

	result := LoadDirective_Success
	if cache == nil {
		var list []FileInfo
		rd_err := read_directory(path, &list)
		defer array_free(&list)

		if len(list) == 1 {
			gb_assert_handler("Assertion Failure", "path != list[0].fullpath", "cmd_check_builtin_load.go", 0)
		}
		switch rd_err {
		case ReadDirectory_InvalidPath:
			error(call, "%.*s error - invalid path: %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_NotFound
		case ReadDirectory_NotExists:
			error(call, "%.*s error - path does not exist: %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_NotFound
		case ReadDirectory_Permission:
			error(call, "%.*s error - unknown error whilst reading path, %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_Error
		case ReadDirectory_NotDir:
			error(call, "%.*s error - expected a directory, got a file: %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_Error
		case ReadDirectory_Empty:
			error(call, "%.*s error - empty directory: %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_NotFound
		case ReadDirectory_Unknown:
			error(call, "%.*s error - unknown error whilst reading path %.*s", name.Len, name.Data, original_string.Len, original_string.Data)
			return LoadDirective_Error
		}
		files_to_reserve := len(list) + 1
		file_caches = make([]*LoadFileCache, 0, files_to_reserve)

		mutex_lock(&c.Info.LoadFileMutex)
		defer mutex_unlock(&c.Info.LoadFileMutex)

		for _, fi := range list {
			if fi.IsDir {
				continue
			}
			var fn_cache *LoadFileCache
			if cache_load_file_directive(c, call, fi.Fullpath, err_on_not_found, &fn_cache, LoadFileTier_Contents, false) {
				file_caches = append(file_caches, fn_cache)
			} else {
				result = LoadDirective_Error
			}
		}
		sort.Slice(file_caches, func(i, j int) bool {
			return file_cache_sort_cmp(file_caches[i], file_caches[j]) < 0
		})
	}
	return result
}

func check_hash_kind(c *CheckerContext, call *Ast, hash_kind String, data []byte, hash_value *uint64) bool {
	ce := call.CallExpr
	bd := ce.Proc.BasicDirective
	name := bd.Name.String

	gb_assert_handler("Assertion Failure", "name == \"load_hash\" || name == \"hash\"", "cmd_check_builtin_load.go", 0)

	supported_hashes := []String{
		{Data: strData("adler32"), Len: 7},
		{Data: strData("crc32"), Len: 5},
		{Data: strData("crc64"), Len: 5},
		{Data: strData("fnv32"), Len: 5},
		{Data: strData("fnv64"), Len: 5},
		{Data: strData("fnv32a"), Len: 6},
		{Data: strData("fnv64a"), Len: 6},
		{Data: strData("murmur32"), Len: 8},
		{Data: strData("murmur64"), Len: 8},
	}

	hash_found := false
	for _, h := range supported_hashes {
		if string_eq(h, hash_kind) {
			hash_found = true
			break
		}
	}
	if !hash_found {
		begin_error_block()
		defer end_error_block()
		error(ce.Proc, "Invalid hash kind passed to `#%.*s`, got: %.*s", name.Len, name.Data, hash_kind.Len, hash_kind.Data)
		error_line("\tAvailable hash kinds:\n")
		for _, h := range supported_hashes {
			error_line("\t%.*s\n", h.Len, h.Data)
		}
		return false
	}

	if string_eq(hash_kind, S("adler32")) {
		*hash_value = gb_adler32(data)
	} else if string_eq(hash_kind, S("crc32")) {
		*hash_value = gb_crc32(data)
	} else if string_eq(hash_kind, S("crc64")) {
		*hash_value = gb_crc64(data)
	} else if string_eq(hash_kind, S("fnv32")) {
		*hash_value = gb_fnv32(data)
	} else if string_eq(hash_kind, S("fnv64")) {
		*hash_value = gb_fnv64(data)
	} else if string_eq(hash_kind, S("fnv32a")) {
		*hash_value = fnv32a(data)
	} else if string_eq(hash_kind, S("fnv64a")) {
		*hash_value = fnv64a(data)
	} else if string_eq(hash_kind, S("murmur32")) {
		*hash_value = gb_murmur32(data)
	} else if string_eq(hash_kind, S("murmur64")) {
		*hash_value = gb_murmur64(data)
	} else {
		compiler_error("unhandled hash kind: %.*s", hash_kind.Len, hash_kind.Data)
	}
	return true
}
