package cmd

func check_builtin_c_procedure(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type) bool {
	builtin_name := builtin_procs[id].Name
	ce := call.CallExpr
	switch id {
	default:
		gb_assert_handler("Panic", 0, "cmd_check_builtin_cproc.go", 0, "Implement C built-in procedure: %.*s", builtin_name.Len, builtin_name.Data)
		return false

	case BuiltinProc_c_va_start:
		var list Operand
		check_expr(c, &list, ce.Args[0])
		if list.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(list.Type, t_c_va_list_ptr) {
			lpt := type_to_string(t_c_va_list_ptr)
			t := type_to_string(list.Type)
			error(list.Expr, "'%.*s' expected a value of type %s, got type %s", builtin_name.Len, builtin_name.Data, lpt, t)
			gb_string_free(t)
			gb_string_free(lpt)
			return false
		}
		var args Operand
		check_expr(c, &args, ce.Args[1])
		if args.Mode == Addressing_Invalid {
			return false
		}
		e := entity_of_node(args.Expr)
		if e == nil || (e.Flags&EntityFlag_CVarArg) == 0 {
			error(list.Expr, "'%.*s' expected a `#c_vararg` parameter", builtin_name.Len, builtin_name.Data)
		}
		operand.Mode = Addressing_NoValue
		operand.Type = nil
		return true

	case BuiltinProc_c_va_end:
		var list Operand
		check_expr(c, &list, ce.Args[0])
		if list.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(list.Type, t_c_va_list_ptr) {
			lpt := type_to_string(t_c_va_list_ptr)
			t := type_to_string(list.Type)
			error(list.Expr, "'%.*s' expected a value of type %s, got type %s", builtin_name.Len, builtin_name.Data, lpt, t)
			gb_string_free(t)
			gb_string_free(lpt)
			return false
		}
		operand.Mode = Addressing_NoValue
		operand.Type = nil
		return true

	case BuiltinProc_c_va_copy:
		var dst Operand
		check_expr(c, &dst, ce.Args[0])
		if dst.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(dst.Type, t_c_va_list_ptr) {
			lpt := type_to_string(t_c_va_list_ptr)
			t := type_to_string(dst.Type)
			error(dst.Expr, "'%.*s' expected a value of type %s, got type %s", builtin_name.Len, builtin_name.Data, lpt, t)
			gb_string_free(t)
			gb_string_free(lpt)
			return false
		}
		var src Operand
		check_expr(c, &src, ce.Args[1])
		if src.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(src.Type, t_c_va_list_ptr) {
			lpt := type_to_string(t_c_va_list_ptr)
			t := type_to_string(src.Type)
			error(src.Expr, "'%.*s' expected a value of type %s, got type %s", builtin_name.Len, builtin_name.Data, lpt, t)
			gb_string_free(t)
			gb_string_free(lpt)
			return false
		}
		operand.Mode = Addressing_NoValue
		operand.Type = nil
		return true

	case BuiltinProc_c_va_arg:
		var list Operand
		check_expr(c, &list, ce.Args[0])
		if list.Mode == Addressing_Invalid {
			return false
		}
		if !are_types_identical(list.Type, t_c_va_list_ptr) {
			lpt := type_to_string(t_c_va_list_ptr)
			t := type_to_string(list.Type)
			error(list.Expr, "'%.*s' expected a value of type %s, got type %s", builtin_name.Len, builtin_name.Data, lpt, t)
			gb_string_free(t)
			gb_string_free(lpt)
			return false
		}
		type_ := check_type(c, ce.Args[1])
		if type_ == nil || type_ == t_invalid {
			error(ce.Args[1], "'%.*s' expected a type as the second parameter to intrinsics.%.*s", builtin_name.Len, builtin_name.Data)
			return false
		}
		operand.Mode = Addressing_Value
		operand.Type = type_
		return true
	}
}

