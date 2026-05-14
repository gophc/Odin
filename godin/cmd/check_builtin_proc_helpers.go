// Depends on: common.odin, checker types
package cmd

func check_builtin_procedure_len_cap(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name String) bool {
	ce := call.CallExpr
	check_expr_or_type(c, operand, ce.args[0])
	if operand.mode == Addressing_Invalid {
		return false
	}
	op_type := type_deref(operand.type)
	type_ := t_int
	if type_hint != nil {
		bt := type_hint
		if bt == t_int || bt == t_uint {
			type_ = type_hint
		}
	}
	mode := Addressing_Invalid
	value := ExactValue{}
	if is_type_string(op_type) && id == BuiltinProc_len {
		if operand.mode == Addressing_Constant {
			mode = Addressing_Constant
			if operand.value.kind == ExactValue_String {
				value = exact_value_i64(operand.value.value_string.len)
			} else if operand.value.kind == ExactValue_String16 {
				value = exact_value_i64(operand.value.value_string16.len)
			} else {
				gb_assert_handler("Panic", 0, "check_builtin.cpp", 2866, "Unhandled value kind: %d", operand.value.kind)
			}
			type_ = t_untyped_integer
		} else {
			mode = Addressing_Value
			if is_type_cstring(op_type) {
				add_package_dependency(c, "runtime", "cstring_len")
			} else if is_type_cstring16(op_type) {
				add_package_dependency(c, "runtime", "cstring16_len")
			}
		}
	} else if is_type_array(op_type) {
		at := core_type(op_type)
		mode = Addressing_Constant
		value = exact_value_i64(at.Array.count)
		type_ = t_untyped_integer
	} else if is_type_fixed_capacity_dynamic_array(op_type) {
		at := core_type(op_type)
		if id == BuiltinProc_cap {
			mode = Addressing_Constant
			value = exact_value_i64(at.FixedCapacityDynamicArray.capacity)
			type_ = t_untyped_integer
		} else {
			mode = Addressing_Value
		}
	} else if is_type_enumerated_array(op_type) && id == BuiltinProc_len {
		at := core_type(op_type)
		mode = Addressing_Constant
		value = exact_value_i64(at.EnumeratedArray.count)
		type_ = t_untyped_integer
	} else if is_type_slice(op_type) && id == BuiltinProc_len {
		mode = Addressing_Value
	} else if is_type_dynamic_array(op_type) {
		mode = Addressing_Value
	} else if is_type_map(op_type) {
		mode = Addressing_Value
	} else if operand.mode == Addressing_Type && is_type_enum(op_type) {
		bt := base_type(op_type)
		mode = Addressing_Constant
		type_ = t_untyped_integer
		if id == BuiltinProc_len {
			value = exact_value_i64(bt.Enum.fields.count)
		} else {
			value = exact_value_sub(*bt.Enum.max_value, *bt.Enum.min_value)
			value = exact_value_increment_one(value)
		}
	} else if is_type_struct(op_type) {
		bt := base_type(op_type)
		if bt.Struct.soa_kind == StructSoa_Fixed {
			mode = Addressing_Constant
			value = exact_value_i64(bt.Struct.soa_count)
			type_ = t_untyped_integer
		} else if (bt.Struct.soa_kind == StructSoa_Slice && id == BuiltinProc_len) ||
			bt.Struct.soa_kind == StructSoa_Dynamic {
			mode = Addressing_Value
		}
	} else if is_type_simd_vector(op_type) {
		bt := base_type(op_type)
		mode = Addressing_Constant
		value = exact_value_i64(bt.SimdVector.count)
		type_ = t_untyped_integer
	}
	if operand.mode == Addressing_Type && mode != Addressing_Constant {
		mode = Addressing_Invalid
	}
	if mode == Addressing_Invalid {
		t := type_to_string(operand.type)
		if is_type_bit_set(op_type) && id == BuiltinProc_len {
			error(call, "'%.*s' is not supported for '%s', did you mean 'card'?", builtin_name.len, builtin_name.data, t)
		} else {
			error(call, "'%.*s' is not supported for '%s'", builtin_name.len, builtin_name.data, t)
		}
		gb_string_free(t)
		return false
	}
	operand.mode = mode
	operand.value = value
	operand.type = type_
	return true
}

func check_builtin_procedure_size_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	if ce.args[0].kind == Ast_UnaryExpr {
		arg := ce.args[0].UnaryExpr
		if arg.op.kind == Token_And {
			begin_error_block()
			warning(ce.args[0], "'size_of(&x)' returns the size of a pointer, not the size of x")
			error_line("\tSuggestion: Use 'size_of(rawptr)' if you want the size of the pointer")
			end_error_block()
		}
	}
	o := Operand{}
	check_expr_or_type(c, &o, ce.args[0])
	if o.mode == Addressing_Invalid {
		return false
	}
	t := o.type
	if t == nil || t == t_invalid {
		error(ce.args[0], "Invalid argument for 'size_of'")
		return false
	}
	t = default_type(t)
	operand.mode = Addressing_Constant
	operand.value = exact_value_i64(type_size_of(t))
	operand.type = t_untyped_integer
	return true
}

func check_builtin_procedure_align_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	o := Operand{}
	check_expr_or_type(c, &o, ce.args[0])
	if o.mode == Addressing_Invalid {
		return false
	}
	t := o.type
	if t == nil || t == t_invalid {
		error(ce.args[0], "Invalid argument for 'align_of'")
		return false
	}
	t = default_type(t)
	operand.mode = Addressing_Constant
	operand.value = exact_value_i64(type_align_of(t))
	operand.type = t_untyped_integer
	return true
}

func check_builtin_procedure_offset_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	var type_ *Type
	var field_arg *Ast
	if ce.args.count == 1 {
		arg0 := unparen_expr(ce.args[0])
		if arg0.kind != Ast_SelectorExpr {
			x := expr_to_string(arg0)
			error(ce.args[0], "Invalid expression for '%.*s', '%s' is not a selector expression", builtin_name.len, builtin_name.data, x)
			gb_string_free(x)
			return false
		}
		se := arg0.SelectorExpr
		x := Operand{}
		check_expr(c, &x, se.expr)
		if x.mode == Addressing_Invalid {
			return false
		}
		type_ = type_deref(x.type)
		bt := base_type(type_)
		if bt == nil || bt == t_invalid {
			error(ce.args[0], "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
			return false
		}
		field_arg = unparen_expr(se.selector)
	} else if ce.args.count == 2 {
		type_ = check_type(c, ce.args[0])
		bt := base_type(type_)
		if bt == nil || bt == t_invalid {
			error(ce.args[0], "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
			return false
		}
		field_arg = unparen_expr(ce.args[1])
	} else {
		error(ce.args[0], "Expected either 1 or 2 arguments to '%.*s'", builtin_name.len, builtin_name.data)
		return false
	}
	field_name := InternedString{}
	if field_arg == nil {
		error(call, "Expected an identifier for field argument")
		return false
	}
	if field_arg.kind == Ast_Ident {
		field_name = field_arg.Ident.interned
	}
	if field_name.value == 0 {
		error(field_arg, "Expected an identifier for field argument")
		return false
	}
	if is_type_array(type_) || is_type_bit_field(type_) {
		t := type_to_string(type_)
		error(field_arg, "Expected a struct type for '%.*s', got '%s'", builtin_name.len, builtin_name.data, t)
		gb_string_free(t)
		return false
	}
	bt := base_type(type_)
	if bt.kind == Type_Struct && bt.Struct.scope != nil {
		if is_type_polymorphic(bt) {
			t := type_to_string(type_)
			error(field_arg, "Cannot use '%.*s' on an unspecialized polymorphic struct type, got '%s'", builtin_name.len, builtin_name.data, t)
			gb_string_free(t)
			return false
		} else if bt.Struct.fields.count == 0 && bt.Struct.node == nil {
			t := type_to_string(type_)
			error(field_arg, "Cannot use '%.*s' on incomplete struct declaration, got '%s'", builtin_name.len, builtin_name.data, t)
			gb_string_free(t)
			return false
		}
	}
	sel := lookup_field(type_, field_name, false)
	if sel.entity == nil {
		begin_error_block()
		type_str := type_to_string_shorthand(type_)
		error(ce.args[0], "'%s' has no field named '%s'", type_str, field_name.cstring())
		gb_string_free(type_str)
		bt := base_type(type_)
		if bt.kind == Type_Struct {
			check_did_you_mean_type(field_name.string(), bt.Struct.fields)
		}
		end_error_block()
		return false
	}
	if sel.indirect {
		type_str := type_to_string_shorthand(type_)
		error(ce.args[0], "Field '%s' is embedded via a pointer in '%s'", field_name.cstring(), type_str)
		gb_string_free(type_str)
		return false
	}
	operand.mode = Addressing_Constant
	operand.value = exact_value_i64(type_offset_of_from_selection(type_, sel))
	operand.type = t_uintptr
	return true
}

func check_builtin_procedure_offset_of_by_string(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	var type_ *Type
	var field_arg *Ast
	if ce.args.count == 2 {
		type_ = check_type(c, ce.args[0])
		bt := base_type(type_)
		if bt == nil || bt == t_invalid {
			error(ce.args[0], "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
			return false
		}
		field_arg = unparen_expr(ce.args[1])
	} else {
		error(ce.args[0], "Expected either 2 arguments to '%.*s'", builtin_name.len, builtin_name.data)
		return false
	}
	field_name := InternedString{}
	if field_arg == nil {
		error(call, "Expected a constant (not-empty) string for field argument")
		return false
	}
	x := Operand{}
	check_expr(c, &x, field_arg)
	if x.mode == Addressing_Constant && x.value.kind == ExactValue_String {
		field_name = string_interner_insert(x.value.value_string)
	}
	if field_name.value == 0 {
		error(field_arg, "Expected a constant (non-empty) string for field argument")
		return false
	}
	if is_type_array(type_) {
		t := type_to_string(type_)
		error(field_arg, "Invalid a struct type for '%.*s', got '%s'", builtin_name.len, builtin_name.data, t)
		gb_string_free(t)
		return false
	}
	sel := lookup_field(type_, field_name, false)
	if sel.entity == nil {
		begin_error_block()
		type_str := type_to_string_shorthand(type_)
		error(ce.args[0], "'%s' has no field named '%s'", type_str, field_name.cstring())
		gb_string_free(type_str)
		bt := base_type(type_)
		if bt.kind == Type_Struct {
			check_did_you_mean_type(field_name.string(), bt.Struct.fields)
		}
		end_error_block()
		return false
	}
	if sel.indirect {
		type_str := type_to_string_shorthand(type_)
		error(ce.args[0], "Field '%s' is embedded via a pointer in '%s'", field_name.string(), type_str)
		gb_string_free(type_str)
		return false
	}
	operand.mode = Addressing_Constant
	operand.value = exact_value_i64(type_offset_of_from_selection(type_, sel))
	operand.type = t_uintptr
	return true
}

func check_builtin_procedure_type_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	expr := ce.args[0]
	o := Operand{}
	check_expr_or_type(c, &o, expr)
	if o.mode == Addressing_Invalid || o.mode == Addressing_Builtin {
		e := entity_of_node(expr)
		if e != nil && e.state == EntityState_InProgress && e.type == nil {
			s := expr_to_string(expr)
			error(expr, "Invalid cyclic type usage from 'type_of', got '%s'", s)
			gb_string_free(s)
		}
		return false
	}
	if o.type == nil || o.type == t_invalid || is_type_asm_proc(o.type) {
		error(o.expr, "Invalid argument to 'type_of'")
		return false
	}
	if is_type_untyped(o.type) {
		t := type_to_string(o.type)
		error(o.expr, "'type_of' of %s cannot be determined", t)
		gb_string_free(t)
		return false
	}
	if c.curr_proc_sig == o.type {
		s := expr_to_string(o.expr)
		error(o.expr, "Invalid cyclic type usage from 'type_of', got '%s'", s)
		gb_string_free(s)
		return false
	}
	if is_type_polymorphic(o.type) {
		error(o.expr, "'type_of' of polymorphic type cannot be determined")
		return false
	}
	operand.mode = Addressing_Type
	operand.type = o.type
	return true
}

func check_builtin_procedure_type_info_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	if c.scope.flags&ScopeFlag_Global != 0 {
		compiler_error("'type_info_of' Cannot be declared within the runtime package")
	}
	if build_context.no_rtti {
		error(call, "'%.*s' has been disallowed", builtin_name.len, builtin_name.data)
		return false
	}
	init_core_type_info(c.checker)
	expr := ce.args[0]
	o := Operand{}
	check_expr_or_type(c, &o, expr)
	if o.mode == Addressing_Invalid {
		return false
	}
	t := o.type
	if t == nil || t == t_invalid || is_type_asm_proc(o.type) || is_type_polymorphic(t) {
		if is_type_polymorphic(t) {
			error(ce.args[0], "Invalid argument for '%.*s', unspecialized polymorphic type", builtin_name.len, builtin_name.data)
		} else {
			error(ce.args[0], "Invalid argument for '%.*s'", builtin_name.len, builtin_name.data)
		}
		return false
	}
	t = default_type(t)
	add_type_info_type(c, t)
	if t_type_info_ptr == nil {
		gb_assert_handler("Assertion Failure", "t_type_info_ptr != nil", "check_builtin.cpp", 3269, "")
	}
	add_type_info_type(c, t_type_info_ptr)
	if is_operand_value(o) && is_type_typeid(t) {
		add_package_dependency(c, "runtime", "__type_info_of")
	} else if o.mode != Addressing_Type {
		error(expr, "Expected a type or typeid for '%.*s'", builtin_name.len, builtin_name.data)
		return false
	}
	operand.mode = Addressing_Value
	operand.type = t_type_info_ptr
	return true
}

func check_builtin_procedure_typeid_of(c *CheckerContext, operand *Operand, call *Ast, builtin_name String) bool {
	ce := call.CallExpr
	if c.scope.flags&ScopeFlag_Global != 0 {
		compiler_error("'typeid_of' Cannot be declared within the runtime package")
	}
	if build_context.no_rtti {
		error(call, "'%.*s' has been disallowed", builtin_name.len, builtin_name.data)
		return false
	}
	init_core_type_info(c.checker)
	expr := ce.args[0]
	o := Operand{}
	check_expr_or_type(c, &o, expr)
	if o.mode == Addressing_Invalid {
		return false
	}
	t := o.type
	if t == nil || t == t_invalid || is_type_asm_proc(t) || is_type_polymorphic(t) {
		error(ce.args[0], "Invalid argument for '%.*s'", builtin_name.len, builtin_name.data)
		return false
	}
	t = default_type(t)
	add_type_info_type(c, t)
	if o.mode != Addressing_Type {
		error(expr, "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
		return false
	}
	operand.mode = Addressing_Value
	operand.type = t_typeid
	operand.value = exact_value_typeid(t)
	return true
}
