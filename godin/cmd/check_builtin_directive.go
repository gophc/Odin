package cmd

import "unsafe"

func check_builtin_procedure_directive(c *CheckerContext, operand *Operand, call *Ast, type_hint *Type) bool {
	ce := call.CallExpr
	bd := ce.Proc.BasicDirective
	name := bd.Name.String

	if string_eq(name, S("location")) {
		if len(ce.Args) > 1 {
			error(ce.Args[0], "'#location' expects either 0 or 1 arguments, got %d", len(ce.Args))
		}
		if len(ce.Args) > 0 {
			arg := ce.Args[0]
			var e *Entity
			var o Operand
			if arg.Kind == Ast_Ident {
				e = check_ident(c, &o, arg, nil, nil, true)
			} else if arg.Kind == Ast_SelectorExpr {
				e = check_selector(c, &o, arg, nil)
			}
			if e == nil {
				error(ce.Args[0], "'#location' expected a valid entity name")
			}
		}
		operand.Type = t_source_code_location
		operand.Mode = Addressing_Value
	} else if string_eq(name, S("caller_expression")) {
		if len(ce.Args) > 1 {
			error(ce.Args[0], "'#caller_expression' expects either 0 or 1 arguments, got %d", len(ce.Args))
		}
		if len(ce.Args) > 0 {
			arg := ce.Args[0]
			if arg.Kind != Ast_Ident {
				error(arg, "'#caller_expression' expected an identifier")
			} else {
				var o Operand
				e := check_ident(c, &o, arg, nil, nil, true)
				if e == nil || (e.Flags&EntityFlag_Param) == 0 {
					error(arg, "'#caller_expression' expected a valid earlier parameter name")
				}
				arg.Ident.Entity = e
			}
		}
		operand.Type = t_string
		operand.Mode = Addressing_Value
	} else if string_eq(name, S("exists")) {
		if len(ce.Args) != 1 {
			error(ce.Close, "'#exists' expects 1 argument, got %d", len(ce.Args))
			return false
		}
		var o Operand
		check_expr(c, &o, ce.Args[0])
		if o.Mode != Addressing_Constant || !is_type_string(o.Type) {
			error(ce.Args[0], "'#exists' expected a constant string argument")
			return false
		}
		operand.Type = t_untyped_bool
		operand.Mode = Addressing_Constant
		original_string := o.Value.ValueString
		var cache *LoadFileCache
		if cache_load_file_directive(c, call, original_string, false, &cache, LoadFileTier_Exists, true) {
			operand.Value = exact_value_bool(cache.Exists)
		} else {
			operand.Value = exact_value_bool(false)
		}
	} else if string_eq(name, S("load")) {
		return check_load_directive(c, operand, call, type_hint, true) == LoadDirective_Success
	} else if string_eq(name, S("load_directory")) {
		return check_load_directory_directive(c, operand, call, type_hint, true) == LoadDirective_Success
	} else if string_eq(name, S("load_hash")) {
		if len(ce.Args) != 2 {
			if len(ce.Args) == 0 {
				error(ce.Close, "'#load_hash' expects 2 argument, got 0")
			} else {
				error(ce.Args[0], "'#load_hash' expects 2 argument, got %d", len(ce.Args))
			}
			return false
		}
		arg0 := ce.Args[0]
		arg1 := ce.Args[1]
		var o Operand
		check_expr(c, &o, arg0)
		if o.Mode != Addressing_Constant {
			error(arg0, "'#load_hash' expected a constant string argument")
			return false
		}
		if !is_type_string(o.Type) {
			str := type_to_string(o.Type)
			error(arg0, "'#load_hash' expected a constant string, got %s", str)
			gb_string_free(str)
			return false
		}
		var o_hash Operand
		check_expr(c, &o_hash, arg1)
		if o_hash.Mode != Addressing_Constant {
			error(arg1, "'#load_hash' expected a constant string argument")
			return false
		}
		if !is_type_string(o_hash.Type) {
			str := type_to_string(o.Type)
			error(arg1, "'#load_hash' expected a constant string, got %s", str)
			gb_string_free(str)
			return false
		}
		gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String", "cmd_check_builtin_directive.go", 0)
		gb_assert_handler("Assertion Failure", "o_hash.Value.Kind == ExactValue_String", "cmd_check_builtin_directive.go", 0)
		original_string := o.Value.ValueString
		hash_kind := o_hash.Value.ValueString
		var cache *LoadFileCache
		if cache_load_file_directive(c, call, original_string, true, &cache, LoadFileTier_Contents, true) {
			mutex_lock(&c.Info.LoadFileMutex)
			defer mutex_unlock(&c.Info.LoadFileMutex)
			var hash_value uint64
			hash_value_ptr := string_map_get(&cache.Hashes, hash_kind)
			if hash_value_ptr != nil {
				hash_value = *hash_value_ptr
			} else {
				data := cache.Data.Data
				file_size := cache.Data.Len
				if !check_hash_kind(c, call, hash_kind, unsafe.Slice(data, file_size), &hash_value) {
					return false
				}
				string_map_set(&cache.Hashes, hash_kind, hash_value)
			}
			operand.Type = t_untyped_integer
			operand.Mode = Addressing_Constant
			operand.Value = exact_value_u64(hash_value)
			return true
		}
		return false
	} else if string_eq(name, S("hash")) {
		if len(ce.Args) != 2 {
			if len(ce.Args) == 0 {
				error(ce.Close, "'#hash' expects 2 argument, got 0")
			} else {
				error(ce.Args[0], "'#hash' expects 2 argument, got %d", len(ce.Args))
			}
			return false
		}
		arg0 := ce.Args[0]
		arg1 := ce.Args[1]
		var o Operand
		check_expr(c, &o, arg0)
		if o.Mode != Addressing_Constant {
			error(arg0, "'#hash' expected a constant string argument")
			return false
		}
		if !is_type_string(o.Type) {
			str := type_to_string(o.Type)
			error(arg0, "'#hash' expected a constant string, got %s", str)
			gb_string_free(str)
			return false
		}
		var o_hash Operand
		check_expr(c, &o_hash, arg1)
		if o_hash.Mode != Addressing_Constant {
			error(arg1, "'#hash' expected a constant string argument")
			return false
		}
		if !is_type_string(o_hash.Type) {
			str := type_to_string(o.Type)
			error(arg1, "'#hash' expected a constant string, got %s", str)
			gb_string_free(str)
			return false
		}
		gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String", "cmd_check_builtin_directive.go", 0)
		gb_assert_handler("Assertion Failure", "o_hash.Value.Kind == ExactValue_String", "cmd_check_builtin_directive.go", 0)
		original_string := o.Value.ValueString
		hash_kind := o_hash.Value.ValueString
		var hash_value uint64
		if check_hash_kind(c, call, hash_kind, unsafe.Slice(original_string.Data, original_string.Len), &hash_value) {
			operand.Type = t_untyped_integer
			operand.Mode = Addressing_Constant
			operand.Value = exact_value_u64(hash_value)
			return true
		}
		return false
	} else if string_eq(name, S("assert")) {
		if len(ce.Args) != 1 && len(ce.Args) != 2 {
			error(call, "'#assert' expects either 1 or 2 arguments, got %d", len(ce.Args))
			return false
		}
		if operand.Type == nil || !is_type_boolean(operand.Type) || operand.Mode != Addressing_Constant {
			str := expr_to_string(ce.Args[0])
			error(call, "'%s' is not a constant boolean", str)
			gb_string_free(str)
			return false
		}
		if len(ce.Args) == 2 {
			arg := unparen_expr(ce.Args[1])
			if arg == nil || arg.Kind != Ast_BasicLit || arg.BasicLit.Token.Kind != Token_String {
				str := expr_to_string(arg)
				error(call, "'%s' is not a constant string", str)
				gb_string_free(str)
				return false
			}
		}
		if !operand.Value.ValueBool {
			begin_error_block()
			defer end_error_block()
			arg1 := expr_to_string(ce.Args[0])
			if len(ce.Args) == 1 {
				error(call, "Compile time assertion: %s", arg1)
			} else {
				arg2 := expr_to_string(ce.Args[1])
				error(call, "Compile time assertion: %s (%s)", arg1, arg2)
				gb_string_free(arg2)
			}
			if c.ProcName != "" {
				str := type_to_string(c.CurrProcSig)
				error_line("\tCalled within '%.*s' :: %s\n", c.ProcName.Len, c.ProcName.Data, str)
				gb_string_free(str)
			}
			gb_string_free(arg1)
		}
		operand.Type = t_untyped_bool
		operand.Mode = Addressing_Constant
	} else if string_eq(name, S("panic")) {
		begin_error_block()
		defer end_error_block()
		if len(ce.Args) != 1 {
			error(call, "'#panic' expects 1 argument, got %d", len(ce.Args))
			return false
		}
		if !is_type_string(operand.Type) && operand.Mode != Addressing_Constant {
			str := expr_to_string(ce.Args[0])
			error(call, "'%s' is not a constant string", str)
			gb_string_free(str)
			return false
		}
		if !build_context.IgnorePanic {
			error(call, "Compile time panic: %.*s", operand.Value.ValueString.Len, operand.Value.ValueString.Data)
			if c.ProcName != "" {
				str := type_to_string(c.CurrProcSig)
				error_line("\tCalled within '%.*s' :: %s\n", c.ProcName.Len, c.ProcName.Data, str)
				gb_string_free(str)
			}
		}
		operand.Type = t_invalid
		operand.Mode = Addressing_NoValue
	} else if string_eq(name, S("defined")) {
		if len(ce.Args) != 1 {
			error(call, "'#defined' expects 1 argument, got %d", len(ce.Args))
			return false
		}
		arg := unparen_expr(ce.Args[0])
		if arg == nil || (arg.Kind != Ast_Ident && arg.Kind != Ast_SelectorExpr) {
			error(call, "'#defined' expects an identifier or selector expression, got %.*s", ast_strings[arg.Kind].Len, ast_strings[arg.Kind].Data)
			return false
		}
		if c.CurrProcDecl == nil {
			error(call, "'#defined' is only allowed within a procedure, prefer the replacement '#config(NAME, default_value)'")
			return false
		}
		is_defined := check_identifier_exists(c.Scope, arg)
		operand.Type = t_untyped_bool
		operand.Mode = Addressing_Constant
		operand.Value = exact_value_bool(is_defined)
		if arg.Kind == Ast_Ident {
			defineable := Defineable{
				Docs:         nil,
				Name:         arg.Ident.Token.String,
				DefaultValue: exact_value_bool(false),
				Pos:          arg.Ident.Token.Pos,
			}
			mutex_lock(&c.Info.DefineablesMutex)
			c.Info.Defineables = append(c.Info.Defineables, defineable)
			mutex_unlock(&c.Info.DefineablesMutex)
		}
	} else if string_eq(name, S("config")) {
		if len(ce.Args) != 2 {
			error(call, "'#config' expects 2 arguments, got %d", len(ce.Args))
			return false
		}
		arg := unparen_expr(ce.Args[0])
		if arg == nil || arg.Kind != Ast_Ident {
			error(call, "'#config' expects an identifier, got %.*s", ast_strings[arg.Kind].Len, ast_strings[arg.Kind].Data)
			return false
		}
		def_arg := unparen_expr(ce.Args[1])
		var def Operand
		check_expr(c, &def, def_arg)
		if def.Mode != Addressing_Constant {
			error(def_arg, "'#config' default value must be a constant")
			return false
		}
		name := arg.Ident.Token.String
		interned := arg.Ident.Interned
		operand.Type = def.Type
		operand.Mode = def.Mode
		operand.Value = def.Value
		found := scope_lookup_current(config_pkg.Scope, interned)
		if found != nil {
			if found.Kind != Entity_Constant {
				error(arg, "'#config' entity '%.*s' found but expected a constant", name.Len, name.Data)
			} else {
				operand.Type = found.Type
				operand.Mode = Addressing_Constant
				operand.Value = found.Constant.Value
			}
		}
		defineable := Defineable{
			Docs:         nil,
			Name:         name,
			DefaultValue: def.Value,
			Pos:          arg.Ident.Token.Pos,
		}
		if c.Decl != nil {
			defineable.Docs = c.Decl.Docs
		}
		mutex_lock(&c.Info.DefineablesMutex)
		c.Info.Defineables = append(c.Info.Defineables, defineable)
		mutex_unlock(&c.Info.DefineablesMutex)
	} else {
		error(call, "Unknown directive call: #%.*s", name.Len, name.Data)
	}
	return true
}
