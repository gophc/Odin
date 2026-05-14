package cmd

func add_map_get_dependencies(c *CheckerContext) {
	if build_context.dynamic_map_calls {
		add_package_dependency(c, "runtime", "__dynamic_map_get", false)
	} else {
		add_package_dependency(c, "runtime", "map_desired_position", false)
		add_package_dependency(c, "runtime", "map_probe_distance", false)
	}
}

func add_map_set_dependencies(c *CheckerContext) {
	init_core_source_code_location(c.checker)
	if t_map_set_proc == nil {
		map_set_args := [5]*Type{t_rawptr, t_uintptr, t_rawptr, t_rawptr, t_source_code_location}
		t_map_set_proc = alloc_type_proc_from_types(map_set_args[:], t_rawptr, false, ProcCC_Odin)
	}
	if build_context.dynamic_map_calls {
		add_package_dependency(c, "runtime", "__dynamic_map_set", false)
	} else {
		add_package_dependency(c, "runtime", "__dynamic_map_check_grow", false)
		add_package_dependency(c, "runtime", "map_insert_hash_dynamic", false)
	}
}

func add_map_reserve_dependencies(c *CheckerContext) {
	init_core_source_code_location(c.checker)
	add_package_dependency(c, "runtime", "__dynamic_map_reserve", false)
}

func check_scope_decls(c *CheckerContext, nodes []*Ast, reserve_size isize) {
	s := c.scope
	check_collect_entities(c, nodes)
	for _, entry := range s.elements {
		e := entry.value
		switch e.kind {
		case Entity_Constant, Entity_TypeName, Entity_Procedure:
		default:
			continue
		}
		d := decl_info_of_entity(e)
		if d != nil {
			check_entity_decl(c, e, d, nil)
		}
	}
}

func check_unary_op(c *CheckerContext, o *Operand, op Token) bool {
	if o.Type == nil {
		str := expr_to_string(o.Expr)
		defer gb_string_free(str)
		error(o.Expr, "Expression has no value '%s'", goStr(str))
		return false
	}
	if o.Mode == Addressing_Type {
		str := type_to_string(o.Type)
		defer gb_string_free(str)
		error_pos(op.Pos, "Expected an expression for operator '%.*s', got type '%s'", op.String.Len, op.String.Data, goStr(str))
		return false
	}
	type_ := base_type(core_array_type(o.Type))
	switch op.Kind {
	case Token_Add, Token_Sub:
		if !is_type_numeric(type_) {
			str := expr_to_string(o.Expr)
			defer gb_string_free(str)
			error_pos(op.Pos, "Operator '%.*s' is not allowed with '%s'", op.String.Len, op.String.Data, goStr(str))
		}
	case Token_Xor:
		if !is_type_integer(type_) && !is_type_boolean(type_) && !is_type_bit_set(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with integers, booleans, or bit sets", op.String.Len, op.String.Data)
		}
	case Token_Not:
		if !is_type_boolean(type_) || is_type_array_like(o.Type) {
			begin_error_block()
			defer end_error_block()
			error_pos(op.Pos, "Operator '%.*s' is only allowed on boolean expressions", op.String.Len, op.String.Data)
			if is_type_integer(type_) {
				str := expr_to_string(o.Expr)
				defer gb_string_free(str)
				error_line("\tSuggestion: Did you mean to do one of the following?\n")
				error_line("\t\t'%s == 0'?\n", goStr(str))
				error_line("\t\tUse of the bitwise not operator '~'?\n")
			}
		} else {
			o.Type = t_untyped_bool
		}
	case Token_Mul:
		begin_error_block()
		defer end_error_block()
		error_pos(op.Pos, "Operator '%.*s' is not a valid unary operator in Odin", op.String.Len, op.String.Data)
		if is_type_pointer(o.Type) {
			str := expr_to_string(o.Expr)
			defer gb_string_free(str)
			error_line("\tSuggestion: Did you mean '%s^'?\n", goStr(str))
			o.Type = type_deref(o.Type)
		} else if is_type_multi_pointer(o.Type) {
			str := expr_to_string(o.Expr)
			defer gb_string_free(str)
			error_line("\tSuggestion: The value is a multi-pointer, did you mean '%s[0]'?\n", goStr(str))
			o.Type = type_deref(o.Type, true)
		}
	default:
		error_pos(op.Pos, "Unknown operator '%.*s'", op.String.Len, op.String.Data)
		return false
	}
	return true
}

func check_binary_op(c *CheckerContext, o *Operand, op Token) bool {
	main_type := o.Type
	type_ := base_type(core_array_type(main_type))
	ct := core_type(type_)
	switch op.Kind {
	case Token_Sub, Token_SubEq:
		if is_type_bit_set(type_) {
			return true
		} else if !is_type_numeric(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with numeric expressions", op.String.Len, op.String.Data)
			return false
		}
	case Token_Quo, Token_QuoEq:
		if is_type_matrix(main_type) {
			error_pos(op.Pos, "Operator '%.*s' is not allowed with matrix types", op.String.Len, op.String.Data)
			return false
		} else if is_type_simd_vector(main_type) && is_type_integer(type_) {
			error_pos(op.Pos, "Operator '%.*s' is not allowed with #simd types with integer elements", op.String.Len, op.String.Data)
			return false
		}
		fallthrough
	case Token_Mul, Token_MulEq, Token_AddEq:
		if is_type_bit_set(type_) {
			return true
		} else if !is_type_numeric(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with numeric expressions", op.String.Len, op.String.Data)
			return false
		}
	case Token_Add:
		if is_type_string(type_) {
			if o.Mode == Addressing_Constant {
				return true
			}
			error_pos(op.Pos, "String concatenation is only allowed with constant strings")
			return false
		} else if is_type_bit_set(type_) {
			return true
		} else if !is_type_numeric(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with numeric expressions", op.String.Len, op.String.Data)
			return false
		}
	case Token_And, Token_Or, Token_AndEq, Token_OrEq, Token_Xor, Token_XorEq:
		if !is_type_integer(ct) && !is_type_boolean(ct) && !is_type_bit_set(ct) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with integers, booleans, or bit sets", op.String.Len, op.String.Data)
			return false
		}
	case Token_Mod, Token_ModMod, Token_ModEq, Token_ModModEq:
		if is_type_matrix(main_type) {
			error_pos(op.Pos, "Operator '%.*s' is not allowed with matrix types", op.String.Len, op.String.Data)
			return false
		}
		if !is_type_integer(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with integers", op.String.Len, op.String.Data)
			return false
		} else if is_type_simd_vector(main_type) {
			error_pos(op.Pos, "Operator '%.*s' is not allowed with #simd types with integer elements", op.String.Len, op.String.Data)
			return false
		}
	case Token_AndNot, Token_AndNotEq:
		if !is_type_integer(ct) && !is_type_bit_set(ct) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with integers and bit sets", op.String.Len, op.String.Data)
			return false
		}
	case Token_CmpAnd, Token_CmpOr, Token_CmpAndEq, Token_CmpOrEq:
		if !is_type_boolean(type_) {
			error_pos(op.Pos, "Operator '%.*s' is only allowed with boolean expressions", op.String.Len, op.String.Data)
			return false
		}
	default:
		error_pos(op.Pos, "Unknown operator '%.*s'", op.String.Len, op.String.Data)
		return false
	}
	return true
}

func exact_bit_set_all_set_mask(type_ *Type) ExactValue {
	type_ = base_type(type_)
	gb_assert_handler("Assertion Failure", "type_->kind == Type_BitSet", "cmd_check_expr_expr.go", 0)
	lower := type_.BitSet.lower
	upper := type_.BitSet.upper
	elem := type_.BitSet.elem
	underlying := type_.BitSet.underlying
	is_backed := underlying != nil
	_ = is_backed

	var b_lower BigInt
	big_int_from_i64(&b_lower, lower)
	var b_upper BigInt
	big_int_from_i64(&b_upper, upper)
	var one BigInt
	big_int_from_u64(&one, 1)
	var mask BigInt
	if elem == nil {
		big_int_from_i64(&mask, -1)
	} else if is_type_enum(elem) {
		e := base_type(elem)
		gb_assert_handler("Assertion Failure", "e->kind == Type_Enum", "cmd_check_expr_expr.go", 0)
		if (big_int_cmp(&e.Enum.min_value.value_integer, &b_lower) == 0 || is_backed) &&
			big_int_cmp(&e.Enum.max_value.value_integer, &b_upper) == 0 {
			var lower_base int64
			if is_backed {
				if lower > 0 {
					lower_base = 0
				} else {
					lower_base = lower
				}
			} else {
				lower_base = lower
			}
			var b_lower_base BigInt
			big_int_from_i64(&b_lower_base, lower_base)
			for _, f := range e.Enum.fields {
				if f.kind != Entity_Constant {
					continue
				}
				if f.Constant.value.kind != ExactValue_Integer {
					continue
				}
				var shift_amount BigInt
				big_int_sub(&shift_amount, &f.Constant.value.value_integer, &b_lower_base)
				var value BigInt
				big_int_shl(&value, &one, &shift_amount)
				big_int_or(&mask, &mask, &value)
			}
		} else {
			big_int_from_i64(&mask, -1)
		}
	} else {
		lower_base := lower
		for x := lower; x <= upper; x++ {
			var shift_amount BigInt
			big_int_from_i64(&shift_amount, x-lower_base)
			var value BigInt
			big_int_shl(&value, &one, &shift_amount)
			big_int_or(&mask, &mask, &value)
		}
	}
	res := ExactValue{}
	res.Kind = ExactValue_Integer
	res.ValueInteger = mask
	return res
}

func check_unary_expr(c *CheckerContext, o *Operand, op Token, node *Ast) {
	switch op.Kind {
	case Token_And:
		if check_is_not_addressable(c, o) {
			if ast_node_expect(node, Ast_UnaryExpr) {
				ue := &node.UnaryExpr
				str := expr_to_string(ue.Expr)
				defer gb_string_free(str)
				e := entity_of_node(ue.Expr)
				if e != nil && (e.flags&EntityFlag_Param) != 0 {
					error_pos(op.Pos, "Cannot take the pointer address of '%s' which is a procedure parameter", goStr(str))
				} else if e != nil && (e.flags&EntityFlag_BitFieldField) != 0 {
					error_pos(op.Pos, "Cannot take the pointer address of '%s' which is a bit_field's field", goStr(str))
				} else {
					switch o.Mode {
					case Addressing_Constant:
						error_pos(op.Pos, "Cannot take the pointer address of '%s' which is a constant", goStr(str))
					case Addressing_SwizzleValue, Addressing_SwizzleVariable:
						error_pos(op.Pos, "Cannot take the pointer address of '%s' which is a swizzle intermediate array value", goStr(str))
					default:
						begin_error_block()
						defer end_error_block()
						error_pos(op.Pos, "Cannot take the pointer address of '%s'", goStr(str))
						if e == nil {
							break
						}
						if (e.flags&EntityFlag_ForValue) != 0 {
							parent_type := type_deref(e.Variable.for_loop_parent_type)
							if parent_type != nil && is_type_string(parent_type) {
								error_line("\tSuggestion: Iterating over a string produces an intermediate 'rune' value which cannot be addressed.\n")
							} else if parent_type != nil && is_type_tuple(parent_type) {
								error_line("\tSuggestion: Iterating over a procedure does not produce values which are addressable.\n")
							} else {
								error_line("\tSuggestion: Did you want to pass the iterable value to the for statement by pointer to get addressable semantics?\n")
							}
							if parent_type != nil && is_type_map(parent_type) {
								error_line("\t            Prefer doing 'for key, &%.*s in ...'\n", e.token.string.Len, e.token.string.Data)
							} else {
								error_line("\t            Prefer doing 'for &%.*s in ...'\n", e.token.string.Len, e.token.string.Data)
							}
						}
						if (e.flags&EntityFlag_SwitchValue) != 0 {
							error_line("\tSuggestion: Did you want to pass the value to the switch statement by pointer to get addressable semantics?\n")
							error_line("\t            Prefer doing 'switch &%.*s in ...'\n", e.token.string.Len, e.token.string.Data)
						}
					}
				}
			}
			o.Mode = Addressing_Invalid
			return
		}
		if o.Mode == Addressing_SoaVariable {
			ue := &node.UnaryExpr
			if ast_node_expect(ue.Expr, Ast_IndexExpr) {
				ie := &ue.Expr.IndexExpr
				soa_type := type_deref(type_of_expr(ie.Expr))
				gb_assert_handler("Assertion Failure", "is_type_soa_struct(soa_type)", "cmd_check_expr_expr.go", 0)
				o.Type = alloc_type_soa_pointer(soa_type)
			} else {
				o.Type = alloc_type_pointer(o.Type)
			}
		} else {
			o.Type = alloc_type_pointer(o.Type)
		}
		switch o.Mode {
		case Addressing_OptionalOk, Addressing_MapIndex:
			o.Mode = Addressing_OptionalOkPtr
		default:
			o.Mode = Addressing_Value
		}
		return
	}
	if !check_unary_op(c, o, op) {
		o.Mode = Addressing_Invalid
		return
	}
	if o.Mode == Addressing_Constant {
		type_ := base_type(o.Type)
		if !is_type_constant_type(o.Type) {
			if is_type_array_like(o.Type) {
				o.Mode = Addressing_Value
				return
			}
			xt := type_to_string(o.Type)
			defer gb_string_free(xt)
			err_str := expr_to_string(node)
			defer gb_string_free(err_str)
			error_pos(op.Pos, "Invalid type, '%s', for constant unary expression '%s'", goStr(xt), goStr(err_str))
			o.Mode = Addressing_Invalid
			return
		}
		if op.Kind == Token_Xor && is_type_untyped(type_) {
			err_str := expr_to_string(node)
			defer gb_string_free(err_str)
			error_pos(op.Pos, "Bitwise not cannot be applied to untyped constants '%s'", goStr(err_str))
			o.Mode = Addressing_Invalid
			return
		}
		if op.Kind == Token_Sub && is_type_unsigned(type_) {
			err_str := expr_to_string(node)
			defer gb_string_free(err_str)
			error_pos(op.Pos, "A unsigned constant cannot be negated '%s'", goStr(err_str))
			o.Mode = Addressing_Invalid
			return
		}
		var precision int32 = 0
		if is_type_typed(type_) {
			precision = int32(8 * type_size_of(type_))
		}
		is_unsigned := is_type_unsigned(type_)
		if is_type_rune(type_) {
			gb_assert_handler("Assertion Failure", "!is_unsigned", "cmd_check_expr_expr.go", 0)
		}
		o.Value = exact_unary_operator_value(op.Kind, o.Value, precision, is_unsigned)
		if op.Kind == Token_Xor && is_type_bit_set(type_) {
			mask := exact_bit_set_all_set_mask(type_)
			o.Value = exact_binary_operator_value(Token_And, o.Value, mask)
		}
		if is_type_typed(type_) {
			if node != nil {
				o.Expr = node
			}
			check_is_expressible(c, o, type_)
		}
		return
	}
	o.Mode = Addressing_Value
}

func add_comparison_procedures_for_fields(c *CheckerContext, t *Type) {
	if t == nil {
		return
	}
	t = base_type(t)
	if !is_type_comparable(t) {
		return
	}
	switch t.Kind {
	case Type_Basic:
		switch t.Basic.kind {
		case Basic_complex32:
			add_package_dependency(c, "runtime", "complex32_eq", false)
			add_package_dependency(c, "runtime", "complex32_ne", false)
		case Basic_complex64:
			add_package_dependency(c, "runtime", "complex64_eq", false)
			add_package_dependency(c, "runtime", "complex64_ne", false)
		case Basic_complex128:
			add_package_dependency(c, "runtime", "complex128_eq", false)
			add_package_dependency(c, "runtime", "complex128_ne", false)
		case Basic_quaternion64:
			add_package_dependency(c, "runtime", "quaternion64_eq", false)
			add_package_dependency(c, "runtime", "quaternion64_ne", false)
		case Basic_quaternion128:
			add_package_dependency(c, "runtime", "quaternion128_eq", false)
			add_package_dependency(c, "runtime", "quaternion128_ne", false)
		case Basic_quaternion256:
			add_package_dependency(c, "runtime", "quaternion256_eq", false)
			add_package_dependency(c, "runtime", "quaternion256_ne", false)
		case Basic_cstring:
			add_package_dependency(c, "runtime", "cstring_eq", false)
			add_package_dependency(c, "runtime", "cstring_ne", false)
		case Basic_string:
			add_package_dependency(c, "runtime", "string_eq", false)
			add_package_dependency(c, "runtime", "string_ne", false)
		case Basic_cstring16:
			add_package_dependency(c, "runtime", "cstring16_eq", false)
			add_package_dependency(c, "runtime", "cstring16_ne", false)
		case Basic_string16:
			add_package_dependency(c, "runtime", "string16_eq", false)
			add_package_dependency(c, "runtime", "string16_ne", false)
		}
	case Type_Struct:
		for _, field := range t.Struct.fields {
			add_comparison_procedures_for_fields(c, field.Type)
		}
	}
}

func check_comparison(c *CheckerContext, node *Ast, x *Operand, y *Operand, op TokenKind) {
	if x.Mode == Addressing_Type && y.Mode == Addressing_Type {
		comp := are_types_identical(x.Type, y.Type)
		switch op {
		case Token_CmpEq:
		case Token_NotEq:
			comp = !comp
		}
		x.Mode = Addressing_Constant
		x.Type = t_untyped_bool
		x.Value = exact_value_bool(comp)
		return
	}
	if x.Mode == Addressing_Type && is_type_typeid(y.Type) {
		add_type_info_type(c, x.Type)
		add_type_info_type(c, y.Type)
		add_type_and_value(c, x.Expr, Addressing_Value, y.Type, exact_value_typeid(x.Type))
		x.Mode = Addressing_Value
		x.Type = t_untyped_bool
		return
	} else if is_type_typeid(x.Type) && y.Mode == Addressing_Type {
		add_type_info_type(c, x.Type)
		add_type_info_type(c, y.Type)
		add_type_and_value(c, y.Expr, Addressing_Value, x.Type, exact_value_typeid(y.Type))
		x.Mode = Addressing_Value
		x.Type = t_untyped_bool
		return
	}
	var err_str String
	if check_is_assignable_to(c, x, y.Type) ||
		check_is_assignable_to(c, y, x.Type) {
		if x.Type.failure || y.Type.failure {
			x.Mode = Addressing_Value
			x.Type = t_untyped_bool
			return
		}
		err_type := x.Type
		defined := false
		switch op {
		case Token_CmpEq, Token_NotEq:
			defined = (is_operand_nil(*x) && type_has_nil(y.Type)) ||
				(is_operand_nil(*y) && type_has_nil(x.Type)) ||
				(is_type_comparable(x.Type) && is_type_comparable(y.Type))
		case Token_Lt, Token_Gt, Token_LtEq, Token_GtEq:
			if are_types_identical(x.Type, y.Type) && is_type_bit_set(x.Type) {
				defined = true
			} else {
				defined = is_type_ordered(x.Type) && is_type_ordered(y.Type)
			}
		}
		if !defined {
			xs := type_to_string(x.Type, temporary_allocator())
			defer gb_string_free(xs)
			ys := type_to_string(y.Type, temporary_allocator())
			defer gb_string_free(ys)
			if !is_type_comparable(x.Type) {
				err_str = gb_string_make(temporary_allocator(),
					gb_bprintf("Type '%s' is not simply comparable, so operator '%.*s' is not defined for it", goStr(xs), token_strings[op].Len, token_strings[op].Data))
			} else if !is_type_comparable(y.Type) {
				err_str = gb_string_make(temporary_allocator(),
					gb_bprintf("Type '%s' is not simply comparable, so operator '%.*s' is not defined for it", goStr(ys), token_strings[op].Len, token_strings[op].Data))
			} else {
				err_str = gb_string_make(temporary_allocator(),
					gb_bprintf("Operator '%.*s' not defined between the types '%s' and '%s'", token_strings[op].Len, token_strings[op].Data, goStr(xs), goStr(ys)))
			}
		} else {
			comparison_type := x.Type
			if x.Type == err_type && is_operand_nil(*x) {
				comparison_type = y.Type
			}
			add_comparison_procedures_for_fields(c, comparison_type)
		}
	} else {
		var xt, yt String
		if x.Mode == Addressing_ProcGroup {
			xt = gb_string_make(temporary_allocator(), "procedure group")
			defer gb_string_free(xt)
		} else {
			xt = type_to_string(x.Type)
			defer gb_string_free(xt)
		}
		if y.Mode == Addressing_ProcGroup {
			yt = gb_string_make(temporary_allocator(), "procedure group")
			defer gb_string_free(yt)
		} else {
			yt = type_to_string(y.Type)
			defer gb_string_free(yt)
		}
		err_str = gb_string_make(temporary_allocator(), gb_bprintf("Mismatched types '%s' and '%s'", goStr(xt), goStr(yt)))
		defer gb_string_free(err_str)
	}
	if err_str != nil {
		error(node, "Cannot compare expression. %s.", goStr(err_str))
		x.Type = t_untyped_bool
	} else {
		if x.Mode == Addressing_Constant &&
			y.Mode == Addressing_Constant {
			if is_type_constant_type(x.Type) {
				if is_type_bit_set(x.Type) {
					switch op {
					case Token_CmpEq, Token_NotEq:
						x.Value = exact_value_bool(compare_exact_values(op, x.Value, y.Value))
					case Token_Lt, Token_LtEq:
						lhs := x.Value
						rhs := y.Value
						res := exact_binary_operator_value(Token_And, lhs, rhs)
						res = exact_value_bool(compare_exact_values(op, res, lhs))
						if op == Token_Lt {
							res = exact_binary_operator_value(Token_And, res, exact_value_bool(compare_exact_values(op, lhs, rhs)))
						}
						x.Value = res
					case Token_Gt, Token_GtEq:
						lhs := x.Value
						rhs := y.Value
						res := exact_binary_operator_value(Token_And, lhs, rhs)
						res = exact_value_bool(compare_exact_values(op, res, rhs))
						if op == Token_Gt {
							res = exact_binary_operator_value(Token_And, res, exact_value_bool(compare_exact_values(op, lhs, rhs)))
						}
						x.Value = res
					}
				} else {
					x.Value = exact_value_bool(compare_exact_values(op, x.Value, y.Value))
				}
			} else {
				x.Mode = Addressing_Value
			}
		} else {
			x.Mode = Addressing_Value
			update_untyped_expr_type(c, x.Expr, default_type(x.Type), true)
			update_untyped_expr_type(c, y.Expr, default_type(y.Type), true)
			var size int64 = 0
			if !is_type_untyped(x.Type) {
				s := type_size_of(x.Type)
				if s > size {
					size = s
				}
			}
			if !is_type_untyped(y.Type) {
				s := type_size_of(y.Type)
				if s > size {
					size = s
				}
			}
			if is_type_cstring(x.Type) && is_type_cstring(y.Type) {
				switch op {
				case Token_CmpEq:
					add_package_dependency(c, "runtime", "cstring_eq", false)
				case Token_NotEq:
					add_package_dependency(c, "runtime", "cstring_ne", false)
				case Token_Lt:
					add_package_dependency(c, "runtime", "cstring_lt", false)
				case Token_Gt:
					add_package_dependency(c, "runtime", "cstring_gt", false)
				case Token_LtEq:
					add_package_dependency(c, "runtime", "cstring_le", false)
				case Token_GtEq:
					add_package_dependency(c, "runtime", "cstring_ge", false)
				}
			} else if is_type_cstring16(x.Type) && is_type_cstring16(y.Type) {
				switch op {
				case Token_CmpEq:
					add_package_dependency(c, "runtime", "cstring16_eq", false)
				case Token_NotEq:
					add_package_dependency(c, "runtime", "cstring16_ne", false)
				case Token_Lt:
					add_package_dependency(c, "runtime", "cstring16_lt", false)
				case Token_Gt:
					add_package_dependency(c, "runtime", "cstring16_gt", false)
				case Token_LtEq:
					add_package_dependency(c, "runtime", "cstring16_le", false)
				case Token_GtEq:
					add_package_dependency(c, "runtime", "cstring16_ge", false)
				}
			} else if is_type_string16(x.Type) || is_type_string16(y.Type) {
				switch op {
				case Token_CmpEq:
					add_package_dependency(c, "runtime", "string16_eq", false)
				case Token_NotEq:
					add_package_dependency(c, "runtime", "string16_ne", false)
				case Token_Lt:
					add_package_dependency(c, "runtime", "string16_lt", false)
				case Token_Gt:
					add_package_dependency(c, "runtime", "string16_gt", false)
				case Token_LtEq:
					add_package_dependency(c, "runtime", "string16_le", false)
				case Token_GtEq:
					add_package_dependency(c, "runtime", "string16_ge", false)
				}
			} else if is_type_string(x.Type) || is_type_string(y.Type) {
				switch op {
				case Token_CmpEq:
					add_package_dependency(c, "runtime", "string_eq", false)
				case Token_NotEq:
					add_package_dependency(c, "runtime", "string_ne", false)
				case Token_Lt:
					add_package_dependency(c, "runtime", "string_lt", false)
				case Token_Gt:
					add_package_dependency(c, "runtime", "string_gt", false)
				case Token_LtEq:
					add_package_dependency(c, "runtime", "string_le", false)
				case Token_GtEq:
					add_package_dependency(c, "runtime", "string_ge", false)
				}
			} else if is_type_complex(x.Type) || is_type_complex(y.Type) {
				switch op {
				case Token_CmpEq:
					switch 8 * size {
					case 64:
						add_package_dependency(c, "runtime", "complex64_eq", false)
					case 128:
						add_package_dependency(c, "runtime", "complex128_eq", false)
					}
				case Token_NotEq:
					switch 8 * size {
					case 64:
						add_package_dependency(c, "runtime", "complex64_ne", false)
					case 128:
						add_package_dependency(c, "runtime", "complex128_ne", false)
					}
				}
			} else if is_type_quaternion(x.Type) || is_type_quaternion(y.Type) {
				switch op {
				case Token_CmpEq:
					switch 8 * size {
					case 128:
						add_package_dependency(c, "runtime", "quaternion128_eq", false)
					case 256:
						add_package_dependency(c, "runtime", "quaternion256_eq", false)
					}
				case Token_NotEq:
					switch 8 * size {
					case 128:
						add_package_dependency(c, "runtime", "quaternion128_ne", false)
					case 256:
						add_package_dependency(c, "runtime", "quaternion256_ne", false)
					}
				}
			}
		}
		x.Type = t_untyped_bool
	}
}

func check_shift(c *CheckerContext, x *Operand, y *Operand, node *Ast, type_hint *Type) {
	gb_assert_handler("Assertion Failure", "node->kind == Ast_BinaryExpr", "cmd_check_expr_expr.go", 0)
	be := &node.BinaryExpr
	_ = be

	y_is_untyped := is_type_untyped(y.Type)
	if y_is_untyped {
		convert_to_typed(c, y, t_untyped_integer)
		if y.Mode == Addressing_Invalid {
			x.Mode = Addressing_Invalid
			return
		}
	} else if !is_type_unsigned(y.Type) {
		y_str := expr_to_string(y.Expr)
		defer gb_string_free(y_str)
		error(y.Expr, "Shift amount '%s' must be an unsigned integer", goStr(y_str))
		x.Mode = Addressing_Invalid
		return
	}
	x_is_untyped := is_type_untyped(x.Type)
	if !(x_is_untyped || is_type_integer(x.Type)) {
		x_str := expr_to_string(x.Expr)
		defer gb_string_free(x_str)
		error(x.Expr, "Shifted operand '%s' must be an integer", goStr(x_str))
		x.Mode = Addressing_Invalid
		return
	}
	if y.Mode == Addressing_Constant {
		if big_int_is_neg(&y.Value.ValueInteger) {
			y_str := expr_to_string(y.Expr)
			defer gb_string_free(y_str)
			error(y.Expr, "Shift amount '%s' cannot be negative", goStr(y_str))
			x.Mode = Addressing_Invalid
			return
		}
		var max_shift BigInt
		big_int_from_u64(&max_shift, 1024)
		if big_int_cmp(&y.Value.ValueInteger, &max_shift) > 0 {
			y_str := expr_to_string(y.Expr)
			defer gb_string_free(y_str)
			error(y.Expr, "Shift amount '%s' must be <= %u", goStr(y_str), 1024)
			x.Mode = Addressing_Invalid
			return
		}
		if x.Mode == Addressing_Constant {
			if x_is_untyped {
				convert_to_typed(c, x, t_untyped_integer)
				if x.Mode == Addressing_Invalid {
					return
				}
				x.Expr = node
				x.Value = exact_value_shift(be.Op.Kind, exact_value_to_integer(x.Value), exact_value_to_integer(y.Value))
				return
			}
			x.Expr = node
			x.Value = exact_value_shift(be.Op.Kind, x.Value, y.Value)
			check_is_expressible(c, x, x.Type)
			return
		}
		if y_is_untyped {
			convert_to_typed(c, y, t_uint)
		}
		return
	}
	if x.Mode == Addressing_Constant {
		if x_is_untyped {
			if type_hint != nil {
				if is_type_integer(type_hint) {
					convert_to_typed(c, x, type_hint)
				} else if is_type_any(type_hint) {
					convert_to_typed(c, x, default_type(t_untyped_integer))
				} else {
					x_str := expr_to_string(x.Expr)
					defer gb_string_free(x_str)
					type_str := type_to_string(type_hint)
					defer gb_string_free(type_str)
					error(x.Expr, "Shifted operand '%s' cannot convert to non-integer type '%s'", goStr(x_str), goStr(type_str))
					x.Mode = Addressing_Invalid
					return
				}
			} else {
				check_is_expressible(c, x, default_type(t_untyped_integer))
			}
			if x.Mode == Addressing_Invalid {
				return
			}
		}
		x.Mode = Addressing_Value
	}
}

func check_binary_array_expr(c *CheckerContext, op Token, x *Operand, y *Operand) bool {
	if is_type_array_like(x.Type) || is_type_array_like(y.Type) {
		if op.Kind == Token_CmpAnd || op.Kind == Token_CmpOr {
			error_pos(op.Pos, "Array programming is not allowed with the operator '%.*s'", op.String.Len, op.String.Data)
		}
	}
	if is_type_array(x.Type) && !is_type_array(y.Type) {
		if check_is_assignable_to(c, y, x.Type) {
			if check_binary_op(c, x, op) {
				return true
			}
		}
	}
	if is_type_simd_vector(x.Type) && !is_type_simd_vector(y.Type) {
		if check_is_assignable_to(c, y, x.Type) {
			if check_binary_op(c, x, op) {
				return true
			}
		}
	}
	return false
}

func is_ise_expr(node *Ast) bool {
	node = unparen_expr(node)
	return node.Kind == Ast_ImplicitSelectorExpr
}

func can_use_other_type_as_type_hint(use_lhs_as_type_hint bool, other_type *Type) bool {
	if use_lhs_as_type_hint {
		return other_type != nil && other_type != t_invalid && is_type_typed(other_type)
	}
	return false
}

func check_matrix_type_hint(matrix *Type, type_hint *Type) *Type {
	xt := base_type(matrix)
	if type_hint != nil {
		th := base_type(type_hint)
		if are_types_identical(th, xt) {
			return type_hint
		} else if xt.Kind == Type_Matrix && th.Kind == Type_Matrix {
			if xt.Matrix.row_count == th.Matrix.row_count &&
				xt.Matrix.column_count == th.Matrix.column_count {
				return type_hint
			}
		} else if xt.Kind == Type_Matrix && th.Kind == Type_Array {
			if xt.Matrix.row_count == 1 && xt.Matrix.column_count == th.Array.count {
				return type_hint
			} else if xt.Matrix.column_count == 1 && xt.Matrix.row_count == th.Array.count {
				return type_hint
			}
		}
	}
	return matrix
}

func check_binary_matrix(c *CheckerContext, op Token, x *Operand, y *Operand, type_hint *Type, use_lhs_as_type_hint bool) {
	if !check_binary_op(c, x, op) {
		x.Mode = Addressing_Invalid
		return
	}
	xt := base_type(x.Type)
	yt := base_type(y.Type)
	if is_type_matrix(x.Type) {
		gb_assert_handler("Assertion Failure", "xt->kind == Type_Matrix", "cmd_check_expr_expr.go", 0)
		if op.Kind == Token_Mul {
			if yt.Kind == Type_Matrix {
				if !are_types_identical(xt.Matrix.elem, yt.Matrix.elem) {
					goto matrix_error
				}
				if xt.Matrix.column_count != yt.Matrix.row_count {
					goto matrix_error
				}
				if xt.Matrix.is_row_major != yt.Matrix.is_row_major {
					goto matrix_error
				}
				x.Mode = Addressing_Value
				if are_types_identical(xt, yt) {
					if are_types_identical(x.Type, y.Type) {
						return
					}
					if !is_type_named(x.Type) && is_type_named(y.Type) {
						x.Type = y.Type
					}
				} else {
					is_row_major := xt.Matrix.is_row_major && yt.Matrix.is_row_major
					x.Type = alloc_type_matrix(xt.Matrix.elem, xt.Matrix.row_count, yt.Matrix.column_count, nil, nil, is_row_major)
				}
				goto matrix_success
			} else if yt.Kind == Type_Array {
				if !are_types_identical(xt.Matrix.elem, yt.Array.elem) {
					goto matrix_error
				}
				if xt.Matrix.column_count != yt.Array.count {
					goto matrix_error
				}
				x.Mode = Addressing_Value
				if xt.Matrix.row_count == yt.Array.count {
					x.Type = y.Type
				} else {
					x.Type = alloc_type_matrix(xt.Matrix.elem, xt.Matrix.row_count, 1, nil, nil, xt.Matrix.is_row_major)
				}
				goto matrix_success
			}
		}
		if !are_types_identical(xt, yt) {
			goto matrix_error
		}
		x.Mode = Addressing_Value
		x.Type = xt
		goto matrix_success
	} else {
		gb_assert_handler("Assertion Failure", "!is_type_matrix(xt)", "cmd_check_expr_expr.go", 0)
		gb_assert_handler("Assertion Failure", "is_type_matrix(yt)", "cmd_check_expr_expr.go", 0)
		if op.Kind == Token_Mul {
			if xt.Kind == Type_Array {
				if !are_types_identical(yt.Matrix.elem, xt.Array.elem) {
					goto matrix_error
				}
				if xt.Array.count != yt.Matrix.row_count {
					goto matrix_error
				}
				x.Mode = Addressing_Value
				if yt.Matrix.column_count == xt.Array.count {
					x.Type = x.Type
				} else {
					x.Type = alloc_type_matrix(yt.Matrix.elem, 1, yt.Matrix.column_count, nil, nil, yt.Matrix.is_row_major)
				}
				goto matrix_success
			} else if are_types_identical(yt.Matrix.elem, xt) {
				x.Type = check_matrix_type_hint(y.Type, type_hint)
				return
			}
		}
		if !are_types_identical(xt, yt) {
			goto matrix_error
		}
		x.Mode = Addressing_Value
		x.Type = xt
		goto matrix_success
	}

matrix_success:
	x.Type = check_matrix_type_hint(x.Type, type_hint)
	return

matrix_error:
	xts := type_to_string(x.Type)
	defer gb_string_free(xts)
	yts := type_to_string(y.Type)
	defer gb_string_free(yts)
	expr_str := expr_to_string(x.Expr)
	defer gb_string_free(expr_str)
	error_pos(op.Pos, "Mismatched types in binary matrix expression '%s' for operator '%.*s' : '%s' vs '%s'", goStr(expr_str), op.String.Len, op.String.Data, goStr(xts), goStr(yts))
	x.Type = t_invalid
	x.Mode = Addressing_Invalid
}

func check_binary_expr_dependency(c *CheckerContext, op Token, bt *Type, REQUIRE bool) {
	if op.Kind == Token_Mod || op.Kind == Token_ModEq ||
		op.Kind == Token_ModMod || op.Kind == Token_ModModEq {
		if bt.Kind == Type_Basic {
			switch bt.Basic.kind {
			case Basic_u128:
				add_package_dependency(c, "runtime", "umodti3", REQUIRE)
			case Basic_i128:
				add_package_dependency(c, "runtime", "modti3", REQUIRE)
			}
		}
	} else if op.Kind == Token_Quo || op.Kind == Token_QuoEq {
		if bt.Kind == Type_Basic {
			switch bt.Basic.kind {
			case Basic_complex32:
				add_package_dependency(c, "runtime", "quo_complex32", false)
			case Basic_complex64:
				add_package_dependency(c, "runtime", "quo_complex64", false)
			case Basic_complex128:
				add_package_dependency(c, "runtime", "quo_complex128", false)
			case Basic_quaternion64:
				add_package_dependency(c, "runtime", "quo_quaternion64", false)
			case Basic_quaternion128:
				add_package_dependency(c, "runtime", "quo_quaternion128", false)
			case Basic_quaternion256:
				add_package_dependency(c, "runtime", "quo_quaternion256", false)
			case Basic_u128:
				add_package_dependency(c, "runtime", "udivti3", REQUIRE)
			case Basic_i128:
				add_package_dependency(c, "runtime", "divti3", REQUIRE)
			}
		}
	} else if op.Kind == Token_Mul || op.Kind == Token_MulEq {
		if bt.Kind == Type_Basic {
			switch bt.Basic.kind {
			case Basic_quaternion64:
				add_package_dependency(c, "runtime", "mul_quaternion64", false)
			case Basic_quaternion128:
				add_package_dependency(c, "runtime", "mul_quaternion128", false)
			case Basic_quaternion256:
				add_package_dependency(c, "runtime", "mul_quaternion256", false)
			case Basic_u128, Basic_i128:
				if is_arch_wasm() {
					add_package_dependency(c, "runtime", "__multi3", REQUIRE)
				}
			}
		}
	} else if op.Kind == Token_Shl || op.Kind == Token_ShlEq {
		if bt.Kind == Type_Basic {
			switch bt.Basic.kind {
			case Basic_u128, Basic_i128:
				if is_arch_wasm() {
					add_package_dependency(c, "runtime", "__ashlti3", REQUIRE)
				}
			}
		}
	} else if op.Kind == Token_Shr || op.Kind == Token_ShrEq {
		if bt.Kind == Type_Basic {
			switch bt.Basic.kind {
			case Basic_u128, Basic_i128:
				if is_arch_wasm() {
					add_package_dependency(c, "runtime", "__lshrti3", REQUIRE)
				}
			}
		}
	}
}

func check_binary_expr(c *CheckerContext, x *Operand, node *Ast, type_hint *Type, use_lhs_as_type_hint bool) {
	gb_assert_handler("Assertion Failure", "node->kind == Ast_BinaryExpr", "cmd_check_expr_expr.go", 0)
	var y_ Operand
	y := &y_
	be := &node.BinaryExpr
	_ = be
	defer func() {
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | be.Left.ViralStateFlags.Load() | be.Right.ViralStateFlags.Load())
	}()
	op := be.Op

	switch op.Kind {
	case Token_CmpEq, Token_NotEq:
		if is_ise_expr(be.Left) {
			check_expr_or_type(c, y, be.Right, nil)
			check_expr_or_type(c, x, be.Left, y.Type)
		} else {
			check_expr_or_type(c, x, be.Left, nil)
			check_expr_or_type(c, y, be.Right, x.Type)
		}
		xt := x.Mode == Addressing_Type
		yt := y.Mode == Addressing_Type
		if xt != yt {
			if xt {
				if !is_type_typeid(y.Type) {
					error_operand_not_expression(x)
				}
			}
			if yt {
				if !is_type_typeid(x.Type) {
					error_operand_not_expression(y)
				}
			}
		}
	case Token_in, Token_not_in:
		check_expr(c, y, be.Right)
		rhs_type := type_deref(y.Type)
		if rhs_type == nil {
			error(y.Expr, "Cannot use '%.*s' on an expression with no value", op.String.Len, op.String.Data)
			x.Mode = Addressing_Invalid
			x.Expr = node
			return
		}
		if is_type_bit_set(rhs_type) {
			elem := base_type(rhs_type).BitSet.elem
			check_expr_with_type_hint(c, x, be.Left, elem)
		} else if is_type_map(rhs_type) {
			key := base_type(rhs_type).Map.key
			check_expr_with_type_hint(c, x, be.Left, key)
		} else {
			check_expr(c, x, be.Left)
		}
		if x.Mode == Addressing_Invalid {
			return
		}
		if y.Mode == Addressing_Invalid {
			x.Mode = Addressing_Invalid
			x.Expr = y.Expr
			return
		}
		if is_type_map(rhs_type) {
			yt := base_type(rhs_type)
			if op.Kind == Token_in {
				check_assignment(c, x, yt.Map.key, make_string_c("map 'in'"))
			} else {
				check_assignment(c, x, yt.Map.key, make_string_c("map 'not_in'"))
			}
			add_map_get_dependencies(c)
		} else if is_type_bit_set(rhs_type) {
			yt := base_type(rhs_type)
			if op.Kind == Token_in {
				check_assignment(c, x, yt.BitSet.elem, make_string_c("bit_set 'in'"))
			} else {
				check_assignment(c, x, yt.BitSet.elem, make_string_c("bit_set 'not_in'"))
			}
			if x.Mode == Addressing_Constant && y.Mode == Addressing_Constant {
				k := exact_value_to_integer(x.Value)
				v := exact_value_to_integer(y.Value)
				gb_assert_handler("Assertion Failure", "k.kind == ExactValue_Integer", "cmd_check_expr_expr.go", 0)
				gb_assert_handler("Assertion Failure", "v.kind == ExactValue_Integer", "cmd_check_expr_expr.go", 0)
				key := big_int_to_i64(&k.ValueInteger)
				lower := yt.BitSet.lower
				upper := yt.BitSet.upper
				if lower <= key && key <= upper {
					var idx BigInt
					big_int_from_i64(&idx, key-lower)
					var bit BigInt
					big_int_from_i64(&bit, 1)
					big_int_shl(&bit, &bit, &idx)
					var mask BigInt
					big_int_and(&mask, &bit, &v.ValueInteger)
					x.Mode = Addressing_Constant
					x.Type = t_untyped_bool
					if op.Kind == Token_in {
						x.Value = exact_value_bool(!big_int_is_zero(&mask))
					} else {
						x.Value = exact_value_bool(big_int_is_zero(&mask))
					}
					x.Expr = node
					return
				} else {
					error(x.Expr, "key '%lld' out of range of bit set, %lld..%lld", key, lower, upper)
					x.Mode = Addressing_Invalid
				}
			}
		} else {
			t := type_to_string(y.Type)
			defer gb_string_free(t)
			error(x.Expr, "expected either a map or bitset for 'in', got %s", goStr(t))
			x.Expr = node
			x.Mode = Addressing_Invalid
			return
		}
		if x.Mode != Addressing_Invalid {
			x.Mode = Addressing_Value
			x.Type = t_untyped_bool
		}
		x.Expr = node
		return
	default:
		if is_ise_expr(be.Left) {
			hint := type_hint
			if token_is_comparison(op.Kind) {
				hint = nil
			}
			check_expr_or_type(c, y, be.Right, hint)
			if can_use_other_type_as_type_hint(use_lhs_as_type_hint, y.Type) {
				check_expr_or_type(c, x, be.Left, y.Type)
			} else {
				check_expr_with_type_hint(c, x, be.Left, type_hint)
			}
		} else {
			check_expr_with_type_hint(c, x, be.Left, type_hint)
			if can_use_other_type_as_type_hint(use_lhs_as_type_hint, x.Type) {
				check_expr_with_type_hint(c, y, be.Right, x.Type)
			} else {
				hint := type_hint
				if token_is_comparison(op.Kind) {
					hint = nil
				}
				check_expr_with_type_hint(c, y, be.Right, hint)
			}
		}
	}
	if x.Mode == Addressing_Invalid {
		return
	}
	if y.Mode == Addressing_Invalid {
		x.Mode = Addressing_Invalid
		x.Expr = y.Expr
		return
	}
	if x.Mode == Addressing_Builtin {
		x.Mode = Addressing_Invalid
		error(x.Expr, "built-in expression in binary expression")
		return
	}
	if y.Mode == Addressing_Builtin {
		x.Mode = Addressing_Invalid
		error(y.Expr, "built-in expression in binary expression")
		return
	}
	if x.Mode == Addressing_ProcGroup {
		x.Mode = Addressing_Invalid
		if x.ProcGroup != nil {
			error(x.Expr, "procedure group '%.*s' used in binary expression", x.ProcGroup.token.string.Len, x.ProcGroup.token.string.Data)
		} else {
			error(x.Expr, "procedure group used in binary expression")
		}
		return
	}
	if y.Mode == Addressing_ProcGroup {
		x.Mode = Addressing_Invalid
		if y.ProcGroup != nil {
			error(y.Expr, "procedure group '%.*s' used in binary expression", y.ProcGroup.token.string.Len, y.ProcGroup.token.string.Data)
		} else {
			error(y.Expr, "procedure group used in binary expression")
		}
		return
	}
	REQUIRE := true
	btx := base_type(x.Type)
	bty := base_type(y.Type)
	check_binary_expr_dependency(c, op, btx, REQUIRE)
	check_binary_expr_dependency(c, op, bty, REQUIRE)
	if token_is_shift(op.Kind) {
		check_shift(c, x, y, node, type_hint)
		return
	}
	switch op.Kind {
	case Token_Quo, Token_Mod, Token_ModMod, Token_QuoEq, Token_ModEq, Token_ModModEq:
		if is_type_integer(y.Type) && !is_type_untyped(y.Type) &&
			is_type_float(x.Type) && is_type_untyped(x.Type) {
			suggestion := "\tSuggestion: Try explicitly casting the constant value for clarity"
			t := type_to_string(y.Type)
			defer gb_string_free(t)
			if x.Value.Kind != ExactValue_Invalid {
				s := exact_value_to_string(x.Value)
				defer gb_string_free(s)
				warning(node, "Dividing an untyped float '%s' by '%s' will perform integer division\n%s", goStr(s), goStr(t), suggestion)
			} else {
				warning(node, "Dividing an untyped float by '%s' will perform integer division\n%s", goStr(t), suggestion)
			}
		}
	}
	convert_to_typed(c, x, y.Type)
	if x.Mode == Addressing_Invalid {
		return
	}
	convert_to_typed(c, y, x.Type)
	if y.Mode == Addressing_Invalid {
		x.Mode = Addressing_Invalid
		return
	}
	if token_is_comparison(op.Kind) {
		check_comparison(c, node, x, y, op.Kind)
		return
	}
	if check_binary_array_expr(c, op, x, y) {
		x.Mode = Addressing_Value
		x.Type = x.Type
		return
	}
	if check_binary_array_expr(c, op, y, x) {
		x.Mode = Addressing_Value
		x.Type = y.Type
		return
	}
	if is_type_matrix(x.Type) || is_type_matrix(y.Type) {
		check_binary_matrix(c, op, x, y, type_hint, use_lhs_as_type_hint)
		x.Expr = node
		return
	}
	if (op.Kind == Token_CmpAnd || op.Kind == Token_CmpOr) &&
		is_type_boolean(x.Type) && is_type_boolean(y.Type) {
	} else if !are_types_identical(x.Type, y.Type) {
		if x.Type != t_invalid &&
			y.Type != t_invalid {
			xt := type_to_string(x.Type)
			defer gb_string_free(xt)
			yt := type_to_string(y.Type)
			defer gb_string_free(yt)
			expr_str := expr_to_string(node)
			defer gb_string_free(expr_str)
			error_pos(op.Pos, "Mismatched types in binary expression '%s' : '%s' vs '%s'", goStr(expr_str), goStr(xt), goStr(yt))
		}
		x.Mode = Addressing_Invalid
		return
	}
	if !check_binary_op(c, x, op) {
		x.Mode = Addressing_Invalid
		return
	}
	switch op.Kind {
	case Token_Quo, Token_Mod, Token_ModMod, Token_QuoEq, Token_ModEq, Token_ModModEq:
		if (x.Mode == Addressing_Constant || is_type_integer(x.Type)) &&
			y.Mode == Addressing_Constant {
			fail := false
			switch y.Value.Kind {
			case ExactValue_Integer:
				if big_int_is_zero(&y.Value.ValueInteger) {
					fail = true
				}
			case ExactValue_Float:
				if y.Value.ValueFloat == 0.0 {
					fail = true
				}
			}
			if fail {
				if is_type_integer(x.Type) || (x.Mode == Addressing_Constant && x.Value.Kind == ExactValue_Integer) {
					if check_for_integer_division_by_zero(c, node) != IntegerDivisionByZero_Trap {
						break
					}
				}
				switch op.Kind {
				case Token_Mod, Token_ModMod, Token_ModEq, Token_ModModEq:
					error(y.Expr, "Division by zero through '%.*s' not allowed", token_strings[op.Kind].Len, token_strings[op.Kind].Data)
				case Token_Quo, Token_QuoEq:
					error(y.Expr, "Division by zero not allowed")
				}
				x.Mode = Addressing_Invalid
				return
			}
		}
	case Token_CmpAnd, Token_CmpOr:
		if be.Left.ViralStateFlags.Load()&ViralStateFlag_ContainsDeferredProcedure != 0 {
			error(be.Left, "Procedure calls that have an associated deferred procedure are not allowed within logical binary expressions")
		}
		if be.Right.ViralStateFlags.Load()&ViralStateFlag_ContainsDeferredProcedure != 0 {
			error(be.Right, "Procedure calls that have an associated deferred procedure are not allowed within logical binary expressions")
		}
	}
	if x.Mode == Addressing_Constant &&
		y.Mode == Addressing_Constant {
		a := x.Value
		b := y.Value
		if !is_type_constant_type(x.Type) {
			x.Mode = Addressing_Value
			return
		}
		actualOp := op.Kind
		if actualOp == Token_Quo && is_type_integer(x.Type) {
			actualOp = Token_QuoEq
		}
		if is_type_bit_set(x.Type) {
			switch actualOp {
			case Token_Add:
				actualOp = Token_Or
			case Token_Sub:
				actualOp = Token_AndNot
			}
		}
		match_exact_values(&a, &b)
		zero_behaviour := check_for_integer_division_by_zero(c, node)
		if zero_behaviour != IntegerDivisionByZero_Trap &&
			b.Kind == ExactValue_Integer && big_int_is_zero(&b.ValueInteger) &&
			(actualOp == Token_QuoEq || actualOp == Token_Mod || actualOp == Token_ModMod) {
			if actualOp == Token_QuoEq {
				switch zero_behaviour {
				case IntegerDivisionByZero_Zero:
					x.Value = b
				case IntegerDivisionByZero_Self:
					x.Value = a
				case IntegerDivisionByZero_AllBits:
					if is_type_untyped(x.Type) {
						x.Value = exact_value_i64(-1)
					} else {
						x.Value = exact_unary_operator_value(Token_Xor, b, int32(8*type_size_of(x.Type)), is_type_unsigned(x.Type))
					}
				}
			} else {
				switch zero_behaviour {
				case IntegerDivisionByZero_Zero, IntegerDivisionByZero_AllBits:
					x.Value = a
				case IntegerDivisionByZero_Self:
					x.Value = b
				}
			}
		} else if is_type_array_like(x.Type) {
			x.Mode = Addressing_Value
			return
		} else {
			x.Value = exact_binary_operator_value(actualOp, a, b)
		}
		if is_type_typed(x.Type) {
			if node != nil {
				x.Expr = node
			}
			check_is_expressible(c, x, x.Type)
		}
		return
	} else if is_type_string(x.Type) {
		error(node, "String concatenation is only allowed with constant strings")
		x.Mode = Addressing_Invalid
		return
	}
	x.Mode = Addressing_Value
}

func make_operand_from_node(node *Ast) Operand {
	gb_assert_handler("Assertion Failure", "node != nullptr", "cmd_check_expr_expr.go", 0)
	var x Operand
	x.Expr = node
	x.Mode = node.TAV.Mode
	x.Type = node.TAV.Type
	x.Value = node.TAV.Value
	return x
}

func check_set_index_data(o *Operand, t *Type, indirection bool, max_count *int64, original_type *Type) bool {
	switch t.Kind {
	case Type_Basic:
		if t.Basic.kind == Basic_string {
			if o.Mode == Addressing_Constant {
				gb_assert_handler("Assertion Failure", "o->value.kind == ExactValue_String", "cmd_check_expr_expr.go", 0)
				*max_count = o.Value.ValueString.Len
			}
			if o.Mode != Addressing_Constant {
				o.Mode = Addressing_Value
			}
			o.Type = t_u8
			return true
		} else if t.Basic.kind == Basic_string16 {
			if o.Mode == Addressing_Constant {
				gb_assert_handler("Assertion Failure", "o->value.kind == ExactValue_String16", "cmd_check_expr_expr.go", 0)
				*max_count = o.Value.ValueString16.Len
			}
			if o.Mode != Addressing_Constant {
				o.Mode = Addressing_Value
			}
			o.Type = t_u16
			return true
		} else if t.Basic.kind == Basic_UntypedString {
			if o.Mode == Addressing_Constant {
				*max_count = o.Value.ValueString.Len
				o.Type = t_u8
				return true
			}
			return false
		}
	case Type_MultiPointer:
		o.Type = t.MultiPointer.elem
		if o.Mode != Addressing_Constant {
			o.Mode = Addressing_Variable
		}
		return true
	case Type_Array:
		*max_count = t.Array.count
		if indirection {
			o.Mode = Addressing_Variable
		} else if o.Mode != Addressing_Variable &&
			o.Mode != Addressing_Constant {
			o.Mode = Addressing_Value
		}
		o.Type = t.Array.elem
		return true
	case Type_EnumeratedArray:
		*max_count = t.EnumeratedArray.count
		if indirection {
			o.Mode = Addressing_Variable
		} else if o.Mode != Addressing_Variable &&
			o.Mode != Addressing_Constant {
			o.Mode = Addressing_Value
		}
		o.Type = t.EnumeratedArray.elem
		return true
	case Type_Matrix:
		if indirection {
			o.Mode = Addressing_Variable
		} else if o.Mode != Addressing_Variable {
			o.Mode = Addressing_Value
		}
		if t.Matrix.is_row_major {
			*max_count = t.Matrix.row_count
			o.Type = alloc_type_array(t.Matrix.elem, t.Matrix.column_count)
		} else {
			*max_count = t.Matrix.column_count
			o.Type = alloc_type_array(t.Matrix.elem, t.Matrix.row_count)
		}
		return true
	case Type_Slice:
		o.Type = t.Slice.elem
		if o.Mode != Addressing_Constant {
			o.Mode = Addressing_Variable
		}
		return true
	case Type_DynamicArray:
		o.Type = t.DynamicArray.elem
		if o.Mode != Addressing_Constant {
			o.Mode = Addressing_Variable
		}
		return true
	case Type_FixedCapacityDynamicArray:
		o.Type = t.FixedCapacityDynamicArray.elem
		if indirection {
			o.Mode = Addressing_Variable
		} else if o.Mode != Addressing_Variable &&
			o.Mode != Addressing_Constant {
			o.Mode = Addressing_Value
		}
		return true
	case Type_Struct:
		if t.Struct.soa_kind != StructSoa_None {
			switch t.Struct.soa_kind {
			case StructSoa_Fixed:
				*max_count = t.Struct.soa_count
			case StructSoa_Slice, StructSoa_Dynamic:
				indirection = o.Mode != Addressing_Constant
			}
			o.Type = t.Struct.soa_elem
			if o.Mode == Addressing_SoaVariable || o.Mode == Addressing_Variable || indirection {
				o.Mode = Addressing_SoaVariable
			} else {
				o.Mode = Addressing_Value
			}
			return true
		}
		return false
	}
	if is_type_pointer(original_type) && indirection {
		ptr := base_type(original_type)
		if ptr.Kind == Type_MultiPointer && o.Mode == Addressing_SoaVariable {
			o.Type = ptr.MultiPointer.elem
			o.Mode = Addressing_Value
			return true
		}
	}
	return false
}

func check_range(c *CheckerContext, node *Ast, is_for_loop bool, x *Operand, y *Operand, inline_for_depth_ *ExactValue, type_hint *Type) bool {
	if !is_ast_range(node) {
		return false
	}
	ie := &node.BinaryExpr
	check_expr_with_type_hint(c, x, ie.Left, type_hint)
	if x.Mode == Addressing_Invalid {
		return false
	}
	check_expr_with_type_hint(c, y, ie.Right, type_hint)
	if y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, x, y.Type)
	if x.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, y, x.Type)
	if y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, x, default_type(y.Type))
	if x.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, y, default_type(x.Type))
	if y.Mode == Addressing_Invalid {
		return false
	}
	if !are_types_identical(x.Type, y.Type) {
		if x.Type != t_invalid &&
			y.Type != t_invalid {
			xt := type_to_string(x.Type)
			defer gb_string_free(xt)
			yt := type_to_string(y.Type)
			defer gb_string_free(yt)
			expr_str := expr_to_string(x.Expr)
			defer gb_string_free(expr_str)
			error(ie.Op, "Mismatched types in interval expression '%s' : '%s' vs '%s'", goStr(expr_str), goStr(xt), goStr(yt))
		}
		return false
	}
	type_ := x.Type
	if is_for_loop {
		if !is_type_integer(type_) && !is_type_float(type_) && !is_type_enum(type_) {
			error(ie.Op, "Only numerical types are allowed within interval expressions")
			return false
		}
	} else {
		if !is_type_integer(type_) && !is_type_float(type_) && !is_type_pointer(type_) && !is_type_enum(type_) {
			error(ie.Op, "Only numerical and pointer types are allowed within interval expressions")
			return false
		}
	}
	if x.Mode == Addressing_Constant &&
		y.Mode == Addressing_Constant {
		a := x.Value
		b := y.Value
		gb_assert_handler("Assertion Failure", "are_types_identical(x->type, y->type)", "cmd_check_expr_expr.go", 0)
		var op TokenKind = Token_Lt
		switch ie.Op.Kind {
		case Token_Ellipsis:
			op = Token_LtEq
		case Token_RangeFull:
			op = Token_LtEq
		case Token_RangeHalf:
			op = Token_Lt
		default:
			error(ie.Op, "Invalid range operator")
		}
		ok := compare_exact_values(op, a, b)
		if !ok {
			error(ie.Op, "Invalid interval range")
			return false
		}
		inline_for_depth := exact_value_sub(b, a)
		if ie.Op.Kind != Token_RangeHalf {
			inline_for_depth = exact_value_increment_one(inline_for_depth)
		}
		if inline_for_depth_ != nil {
			*inline_for_depth_ = inline_for_depth
		}
	} else if inline_for_depth_ != nil {
		error(ie.Op, "Interval expressions must be constant")
		return false
	}
	add_type_and_value(c, ie.Left, x.Mode, x.Type, x.Value)
	add_type_and_value(c, ie.Right, y.Mode, y.Type, y.Value)
	return true
}
