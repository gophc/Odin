package cmd

func enum_constant_entity_cmp(a, b *Entity) int {
	gb_assert_handler("Assertion Failure", "a.Kind == Entity_Constant && b.Kind == Entity_Constant", "cmd_check_builtin_base.go", 0)
	gb_assert_handler("Assertion Failure", "a.Constant.Value.Kind == ExactValue_Integer && b.Constant.Value.Kind == ExactValue_Integer", "cmd_check_builtin_base.go", 0)
	return big_int_cmp(&a.Constant.Value.ValueInteger, &b.Constant.Value.ValueInteger)
}

var builtin_type_is_procs = []func(*Type) bool{
	nil,                     // 0 (offset)
	is_type_boolean,
	is_type_bit_field,
	is_type_integer,
	is_type_rune,
	is_type_float,
	is_type_complex,
	is_type_quaternion,
	is_type_string,
	is_type_string16,
	is_type_cstring,
	is_type_cstring16,
	is_type_typeid,
	is_type_any,
	is_type_endian_platform,
	is_type_endian_little,
	is_type_endian_big,
	is_type_unsigned,
	is_type_numeric,
	is_type_ordered,
	is_type_ordered_numeric,
	is_type_indexable,
	is_type_sliceable,
	is_type_comparable,
	is_type_simple_compare,
	is_type_nearly_simple_compare,
	is_type_dereferenceable,
	is_type_valid_for_keys,
	is_type_valid_for_matrix_elems,
	is_type_named,
	is_type_pointer,
	is_type_multi_pointer,
	is_type_array,
	is_type_enumerated_array,
	is_type_slice,
	is_type_dynamic_array,
	is_type_map,
	is_type_struct,
	is_type_union,
	is_type_enum,
	is_type_proc,
	is_type_bit_set,
	is_type_simd_vector,
	is_type_matrix,
	is_type_raw_union,
	is_type_fixed_capacity_dynamic_array,
	is_type_polymorphic_record_specialized,
	is_type_polymorphic_record_unspecialized,
	type_has_nil,
}

func check_or_else_right_type(c *CheckerContext, expr *Ast, name String, right_type *Type) {
	if right_type == nil {
		return
	}
	if !is_type_boolean(right_type) && !type_has_nil(right_type) {
		str := type_to_string(right_type)
		error(expr, "'%.*s' expects an \"optional ok\" like value, or an n-valued expression where the last value is either a boolean or can be compared against 'nil', got %s", name.Len, name.Data, str)
		gb_string_free(str)
	}
}

func check_or_else_split_types(c *CheckerContext, x *Operand, name String, left_type_, right_type_ **Type) {
	var left_type *Type
	var right_type *Type
	if x.Type.Kind == Type_Tuple {
		vars := x.Type.Tuple.Variables
		lhs := vars[:len(vars)-1]
		rhs := vars[len(vars)-1]
		if len(lhs) == 1 {
			left_type = lhs[0].Type
		} else if len(lhs) != 0 {
			left_type = alloc_type_tuple()
			left_type.Tuple.Variables = lhs
		}
		right_type = rhs.Type
	} else {
		check_promote_optional_ok(c, x, &left_type, &right_type)
	}
	if left_type_ != nil {
		*left_type_ = left_type
	}
	if right_type_ != nil {
		*right_type_ = right_type
	}
	check_or_else_right_type(c, x.Expr, name, right_type)
}

func check_or_else_expr_no_value_error(c *CheckerContext, name String, x Operand, type_hint *Type) {
	begin_error_block()
	defer end_error_block()
	t := type_to_string(x.Type)
	error(x.Expr, "'%.*s' does not return a value, value is of type %s", name.Len, name.Data, t)
	if is_type_union(type_deref(x.Type)) {
		bsrc := base_type(type_deref(x.Type))
		var th gbString
		if type_hint != nil {
			gb_assert_handler("Assertion Failure", "bsrc.Kind == Type_Union", "cmd_check_builtin_base.go", 0)
			for _, vt := range bsrc.Union.Variants {
				if are_types_identical(vt, type_hint) {
					th = type_to_string(type_hint)
					break
				}
			}
		}
		expr_str := expr_to_string(x.Expr)
		if th != nil {
			error_line("\tSuggestion: was a type assertion such as %s.(%s) or %s.? wanted?\n", expr_str, th, expr_str)
		} else {
			error_line("\tSuggestion: was a type assertion such as %s.(T) or %s.? wanted?\n", expr_str, expr_str)
		}
		gb_string_free(th)
		gb_string_free(expr_str)
	}
	gb_string_free(t)
}

func check_or_return_split_types(c *CheckerContext, x *Operand, name String, left_type_, right_type_ **Type) {
	var left_type *Type
	var right_type *Type
	if x.Type.Kind == Type_Tuple {
		vars := x.Type.Tuple.Variables
		lhs := vars[:len(vars)-1]
		rhs := vars[len(vars)-1]
		if len(lhs) == 1 {
			left_type = lhs[0].Type
		} else if len(lhs) != 0 {
			left_type = alloc_type_tuple()
			left_type.Tuple.Variables = lhs
		}
		right_type = rhs.Type
	} else {
		check_promote_optional_ok(c, x, &left_type, &right_type)
	}
	if left_type_ != nil {
		*left_type_ = left_type
	}
	if right_type_ != nil {
		*right_type_ = right_type
	}
	check_or_else_right_type(c, x.Expr, name, right_type)
}

func is_constant_string(c *CheckerContext, builtin_name String, expr *Ast, name_ *String) bool {
	var op Operand
	check_expr(c, &op, expr)
	if op.Mode == Addressing_Constant && op.Value.Kind == ExactValue_String {
		if name_ != nil {
			*name_ = op.Value.ValueString
		}
		return true
	}
	e := expr_to_string(op.Expr)
	t := type_to_string(op.Type)
	error(op.Expr, "'%.*s' expected a constant string value, got %s of type %s", builtin_name.Len, builtin_name.Data, e, t)
	gb_string_free(t)
	gb_string_free(e)
	return false
}
