// Depends on: common.odin, types, exact_value
package cmd

func check_distance_between_types(c *CheckerContext, operand *Operand, type_ *Type, allow_array_programming bool) int64 {
	if c == nil {
		gb_assert_handler("Assertion Failure", "operand.Mode == Addressing_Value", "cmd_check_expr_assign.go", 0)
		gb_assert_handler("Assertion Failure", "is_type_typed(operand.Type)", "cmd_check_expr_assign.go", 0)
	}
	if operand.Mode == Addressing_Invalid ||
		type_ == t_invalid {
		return -1
	}
	if operand.Mode == Addressing_Builtin {
		return -1
	}
	if operand.Mode == Addressing_Type {
		if is_type_typeid(type_) {
			if is_type_polymorphic(operand.Type) {
				return -1
			}
			add_type_info_type(c, operand.Type)
			return 4
		}
		return -1
	}
	if operand.Mode == Addressing_ProcGroup && !is_type_proc(type_) {
		return -1
	}
	s := operand.Type
	if are_types_identical(s, type_) {
		return 0
	}
	src := base_type(s)
	dst := base_type(type_)
	if is_type_untyped_uninit(src) {
		return 1
	}
	if is_type_untyped_nil(src) {
		if type_has_nil(dst) {
			return 1
		}
		return -1
	}
	if is_type_untyped(src) {
		if is_type_any(dst) {
			add_type_info_type(c, s)
			return 10
		}
		if dst.Kind == TypeBasic {
			if operand.Mode == Addressing_Constant {
				if check_representable_as_constant(c, operand.Value, dst, nil) {
					if is_type_typed(dst) && src.Kind == TypeBasic {
						switch src.Basic.Kind {
						case BasicUntypedBool:
							if is_type_boolean(dst) {
								return 1
							}
						case BasicUntypedRune:
							if is_type_integer(dst) || is_type_rune(dst) {
								return 1
							}
						case BasicUntypedInteger:
							if is_type_integer(dst) || is_type_rune(dst) {
								return 1
							}
						case BasicUntypedString:
							if is_type_string(dst) {
								return 1
							}
						case BasicUntypedFloat:
							if is_type_float(dst) {
								return 1
							}
						case BasicUntypedComplex:
							if is_type_complex(dst) {
								return 1
							}
							if is_type_quaternion(dst) {
								return 2
							}
						case BasicUntypedQuaternion:
							if is_type_quaternion(dst) {
								return 1
							}
						}
					}
					return 2
				}
				return -1
			}
			if src.Kind == TypeBasic {
				d := base_array_type(dst)
				score := int64(-1)
				switch src.Basic.Kind {
				case BasicUntypedBool:
					if is_type_boolean(d) {
						score = 1
					}
				case BasicUntypedRune:
					if is_type_integer(d) || is_type_rune(d) {
						score = 1
					}
				case BasicUntypedInteger:
					if is_type_integer(d) || is_type_rune(d) {
						score = 1
					}
				case BasicUntypedString:
					if is_type_string(d) {
						score = 1
					}
				case BasicUntypedFloat:
					if is_type_float(d) {
						score = 1
					}
				case BasicUntypedComplex:
					if is_type_complex(d) {
						score = 1
					}
					if is_type_quaternion(d) {
						score = 2
					}
				case BasicUntypedQuaternion:
					if is_type_quaternion(d) {
						score = 1
					}
				}
				if score > 0 {
					if is_type_typed(d) {
						score += 1
					}
					if d != dst {
						score += 6
					}
				}
				return score
			}
		}
	}
	if c != nil {
		if is_type_enum(dst) && are_types_identical(dst.Enum.BaseType, operand.Type) {
			if c.in_enum_type {
				return 3
			}
		}
	}
	{
		subtype_level := check_is_assignable_to_using_subtype(operand.Type, type_)
		if subtype_level > 0 {
			return 4 + subtype_level
		}
	}
	if are_types_identical(type_, t_rawptr) && is_type_pointer(src) {
		return 5
	}
	if are_types_identical(type_, t_rawptr) && is_type_multi_pointer(src) {
		return 5
	}
	if dst.Kind == TypePointer && src.Kind == TypeMultiPointer {
		if are_types_identical(dst.Pointer.Elem, src.MultiPointer.Elem) {
			return 4
		}
	}
	if dst.Kind == TypeMultiPointer && src.Kind == TypePointer {
		if are_types_identical(dst.MultiPointer.Elem, src.Pointer.Elem) {
			return 4
		}
	}
	if is_type_polymorphic(dst) && !is_type_polymorphic(src) {
		modify_type := !c.NoPolymorphicErrors
		if is_polymorphic_type_assignable(c, type_, s, false, modify_type) {
			return 2
		}
	}
	if is_type_union(dst) {
		for _, vt := range dst.Union.Variants {
			if are_types_identical(vt, s) {
				return 1
			}
			if is_type_proc(vt) {
				if are_types_identical(base_type(vt), src) {
					return 1
				}
			}
		}
		if len(dst.Union.Variants) == 1 {
			vt := dst.Union.Variants[0]
			score := check_distance_between_types(c, operand, vt, allow_array_programming)
			if score >= 0 {
				return score + 2
			}
		} else if is_type_untyped(src) || is_type_struct(type_deref(src, false)) {

			prev_lowest_score := int64(-1)
			lowest_score := int64(-1)
			for _, vt := range dst.Union.Variants {
				score := check_distance_between_types(c, operand, vt, allow_array_programming)
				if score >= 0 {
					if lowest_score < 0 {
						lowest_score = score
					} else {
						if prev_lowest_score < 0 {
							prev_lowest_score = lowest_score
						} else {
							if prev_lowest_score > lowest_score {
								prev_lowest_score = lowest_score
							}
						}
						if lowest_score > score {
							lowest_score = score
						}
					}
				}
			}
			if lowest_score >= 0 {
				if prev_lowest_score != lowest_score {
					return lowest_score + 2
				}
			}
		}
	}
	if is_type_proc(dst) {
		if are_types_identical(src, dst) {
			return 3
		}
		poly_proc_data := &PolyProcData{}
		if check_polymorphic_procedure_assignment(c, operand, type_, operand.Expr, poly_proc_data) {
			e := poly_proc_data.GenEntity
			add_type_and_value(c, operand.Expr, Addressing_Value, e.Type, ExactValue{})
			add_entity_use(c, operand.Expr, e)
			return 4
		}
		if is_type_proc(src) && are_proc_properties_identical(dst, src) && check_proc_params_assignable(c, dst, src) {
			return 4
		}
	}
	if is_type_complex_or_quaternion(dst) {
		elem := base_complex_elem_type(dst)
		if are_types_identical(elem, base_type(src)) {
			return 5
		}
	}
	if allow_array_programming {
		if is_type_array(dst) {
			elem := base_array_type(dst)
			distance := check_distance_between_types(c, operand, elem, allow_array_programming)
			if distance >= 0 {
				return distance + 6
			}
		}
		if is_type_simd_vector(dst) {
			dst_elem := base_array_type(dst)
			distance := check_distance_between_types(c, operand, dst_elem, allow_array_programming)
			if distance >= 0 {
				return distance + 6
			}
		}
	}
	if is_type_matrix(dst) {
		if are_types_identical(src, dst) {
			return 5
		}
		if dst.Matrix.RowCount == dst.Matrix.ColumnCount {
			dst_elem := base_array_type(dst)
			distance := check_distance_between_types(c, operand, dst_elem, allow_array_programming)
			if distance >= 0 {
				return distance + 7
			}
		}
	}
	if is_type_any(dst) {
		if !is_type_polymorphic(src) {
			if operand.Mode == Addressing_Context && operand.Type == t_context {
				return -1
			} else {
				add_type_info_type(c, s)
				return 10
			}
		}
	}
	expr := unparen_expr(operand.Expr)
	if expr != nil {
		if expr.Kind == AstAutoCast {
			x := *operand
			x.Expr = expr.AutoCast.Expr
			if check_cast_internal(c, &x, type_) {
				return 10
			}
		}
	}
	return -1
}

func assign_score_function(distance int64, is_variadic bool) int64 {
	const c = 3*10*10 + 1
	d := distance * distance
	if is_variadic && d >= 0 {
		d += distance + 1
	}
	if c-d > 0 {
		return c - d
	}
	return 0
}

func check_is_assignable_to_with_score(c *CheckerContext, operand *Operand, type_ *Type, score_ *int64, is_variadic bool, allow_array_programming bool) bool {
	if c == nil {
		gb_assert_handler("Assertion Failure", "operand.Mode == Addressing_Value", "cmd_check_expr_assign.go", 0)
		gb_assert_handler("Assertion Failure", "is_type_typed(operand.Type)", "cmd_check_expr_assign.go", 0)
	}
	if operand.Mode == Addressing_Invalid || type_ == t_invalid {
		if score_ != nil {
			*score_ = 0
		}
		return false
	}
	if operand.Mode == Addressing_Value && is_type_proc(type_) && is_type_proc(operand.Type) {
		e := entity_from_expr(operand.Expr)
		if e != nil && e.Kind == Entity_Procedure && is_type_polymorphic(e.Type, false) && !is_type_polymorphic(type_, false) {
			if score_ != nil {
				*score_ = assign_score_function(1, false)
			}
			return true
		}
	}
	score := check_distance_between_types(c, operand, type_, allow_array_programming)
	if score >= 0 {
		if score_ != nil {
			*score_ = assign_score_function(score, is_variadic)
		}
		return true
	}
	if score_ != nil {
		*score_ = 0
	}
	return false
}

func check_is_assignable_to(c *CheckerContext, operand *Operand, type_ *Type) bool {
	var score int64
	return check_is_assignable_to_with_score(c, operand, type_, &score, false, true)
}

func internal_check_is_assignable_to(src, dst *Type) bool {
	x := &Operand{}
	x.Type = src
	x.Mode = Addressing_Value
	return check_is_assignable_to(nil, x, dst)
}

func check_proc_params_assignable(c *CheckerContext, dst, src *Type) bool {
	gb_assert_handler("Assertion Failure", "dst.Kind == Type_Proc", "cmd_check_expr_assign.go", 0)
	gb_assert_handler("Assertion Failure", "src.Kind == Type_Proc", "cmd_check_expr_assign.go", 0)
	if dst.Proc.Params == nil || src.Proc.Params == nil {
		return false
	}
	if !are_types_identical(src.Proc.Results, dst.Proc.Results) {
		return false
	}
	dst_tuple := dst.Proc.Params.Tuple
	src_tuple := src.Proc.Params.Tuple
	if len(dst_tuple.Variables) == len(src_tuple.Variables) && dst_tuple.IsPacked == src_tuple.IsPacked {
		for i := range dst_tuple.Variables {
			edst := dst_tuple.Variables[i]
			esrc := src_tuple.Variables[i]
			if edst.Kind != esrc.Kind || !are_types_identical(edst.Type, esrc.Type) {
				if edst.Type.Kind == TypePointer && esrc.Type.Kind == TypePointer &&
					is_type_struct(esrc.Type.Pointer.Elem) &&
					check_is_assignable_to_using_offset_zero_subtype(esrc.Type.Pointer.Elem, edst.Type.Pointer.Elem) {
					continue
				}
				return false
			}
			if edst.Kind == Entity_Constant && !compare_exact_values(Token_CmpEq, edst.Constant.Value, esrc.Constant.Value) {
				return false
			}
		}
		return true
	}
	return false
}

func get_package_of_type(type_ *Type) *AstPackage {
	for {
		if type_ == nil {
			return nil
		}
		switch type_.Kind {
		case TypeBasic:
			return builtin_pkg
		case TypeNamed:
			if type_.Named.TypeName != nil {
				return type_.Named.TypeName.Pkg
			}
			return nil
		case TypePointer:
			type_ = type_.Pointer.Elem
			continue
		case TypeArray:
			type_ = type_.Array.Elem
			continue
		case TypeSlice:
			type_ = type_.Slice.Elem
			continue
		case TypeDynamicArray:
			type_ = type_.DynamicArray.Elem
			continue
		}
		return nil
	}
}

func check_assignment(c *CheckerContext, operand *Operand, type_ *Type, context_name String) {
	check_not_tuple(c, operand)
	if operand.Mode == Addressing_Invalid {
		return
	}
	if is_type_untyped(operand.Type) {
		target_type := type_
		if type_ == nil || is_type_any(type_) {
			if type_ == nil && is_type_untyped_uninit(operand.Type) {
				article := error_article(context_name)
				error(operand.Expr, "Use of --- in %.*s%.*s", article.Len, article.Data, context_name.Len, context_name.Data)
				operand.Mode = Addressing_Invalid
				return
			}
			if type_ == nil && is_type_untyped_nil(operand.Type) {
				article := error_article(context_name)
				error(operand.Expr, "Use of untyped nil in %.*s%.*s", article.Len, article.Data, context_name.Len, context_name.Data)
				operand.Mode = Addressing_Invalid
				return
			}
			target_type = default_type(operand.Type)
			if type_ != nil && !is_type_any(type_) {
				gb_assert_handler("Assertion Failure", "is_type_typed(target_type)", "cmd_check_expr_assign.go", 0)
			}
			add_type_info_type(c, type_)
			add_type_info_type(c, target_type)
		}
		convert_to_typed(c, operand, target_type)
		if operand.Mode == Addressing_Invalid {
			return
		}
	}
	if type_ == nil {
		return
	}
	if operand.Mode == Addressing_ProcGroup {
		good := false
		if type_ != nil && is_type_proc(type_) {
			procs := proc_group_entities(c, *operand)
			for _, e := range procs {
				t := base_type(e.Type)
				if t == t_invalid {
					continue
				}
				x := &Operand{}
				x.Mode = Addressing_Value
				x.Type = t
				if check_is_assignable_to(c, x, type_) {
					if operand.Expr.Kind == AstSelectorExpr {
						add_entity_use(c, operand.Expr.SelectorExpr.Selector, e)
					} else {
						add_entity_use(c, operand.Expr, e)
					}
					good = true
					break
				}
			}
		}
		if !good {
			expr_str := expr_to_string(operand.Expr)
			op_type_str := type_to_string(operand.Type)
			type_str := type_to_string(type_)
			defer gb_string_free(type_str)
			defer gb_string_free(op_type_str)
			defer gb_string_free(expr_str)
			article := error_article(context_name)
			error(operand.Expr,
				"Cannot assign overloaded procedure group '%s' to '%s' in %.*s%.*s",
				expr_str,
				op_type_str,
				article.Len, article.Data,
				context_name.Len, context_name.Data)
			operand.Mode = Addressing_Invalid
		}
		convert_to_typed(c, operand, type_)
		return
	}
	if check_is_assignable_to(c, operand, type_) {
		if operand.Mode == Addressing_Type && is_type_typeid(type_) {
			add_type_info_type(c, operand.Type)
			add_type_and_value(c, operand.Expr, Addressing_Value, type_, exact_value_typeid(operand.Type))
		}
	} else {
		expr_str := expr_to_string(operand.Expr)
		op_type_str := type_to_string(operand.Type)
		type_str := type_to_string(type_)
		defer gb_string_free(type_str)
		defer gb_string_free(op_type_str)
		defer gb_string_free(expr_str)
		article := error_article(context_name)
		switch operand.Mode {
		case Addressing_Builtin:
			error(operand.Expr,
				"Cannot assign built-in procedure '%s' to %.*s%.*s",
				expr_str,
				article.Len, article.Data,
				context_name.Len, context_name.Data)
		case Addressing_Type:
			if is_type_polymorphic(operand.Type, false) {
				error(operand.Expr,
					"Cannot assign '%s', a polymorphic type, to %.*s%.*s",
					op_type_str,
					article.Len, article.Data,
					context_name.Len, context_name.Data)
			} else {
				begin_error_block()
				defer end_error_block()
				error(operand.Expr,
					"Cannot assign '%s', a type, to %.*s%.*s",
					op_type_str,
					article.Len, article.Data,
					context_name.Len, context_name.Data)
				if type_ != nil && are_types_identical(type_, t_any) {
					error_line("\tSuggestion: 'typeid_of(%s)'", expr_str)
				}
			}
		default:
			op_type_extra := gb_string_make(heap_allocator(), "")
			type_extra := gb_string_make(heap_allocator(), "")
			defer gb_string_free(op_type_extra)
			defer gb_string_free(type_extra)
			on := gb_string_length(op_type_str)
			tn := gb_string_length(type_str)
			if on == tn && gb_strncmp(op_type_str, type_str, on) == 0 {
				op_pkg := get_package_of_type(operand.Type)
				type_pkg := get_package_of_type(type_)
				if op_pkg != nil {
					op_type_extra = gb_string_append_fmt(op_type_extra, " (package %.*s)", op_pkg.Name.Len, op_pkg.Name.Data)
				}
				if type_pkg != nil {
					type_extra = gb_string_append_fmt(type_extra, " (package %.*s)", type_pkg.Name.Len, type_pkg.Name.Data)
				}
			}
			begin_error_block()
			defer end_error_block()
			error(operand.Expr,
				"Cannot assign value '%s' of type '%s%s' to '%s%s' in %.*s%.*s",
				expr_str,
				op_type_str, op_type_extra,
				type_str, type_extra,
				article.Len, article.Data,
				context_name.Len, context_name.Data)
			check_assignment_error_suggestion(c, operand, type_)
			src := base_type(operand.Type)
			dst := base_type(type_)
			if string_compare(context_name, make_string_c("procedure argument")) == 0 {
				if is_type_slice(src) && are_types_identical(src.Slice.Elem, dst) {
					a := expr_to_string(operand.Expr)
					error_line("\tSuggestion: Did you mean to pass the slice into the variadic parameter with ..%s?\n\n", a)
					gb_string_free(a)
				}
			}
			if src.Kind == dst.Kind && src.Kind == TypeProc {
				x := src
				y := dst
				same_inputs := areTypesIdenticalInternal(x.Proc.Params, y.Proc.Params, false)
				same_outputs := areTypesIdenticalInternal(x.Proc.Results, y.Proc.Results, false)
				if same_inputs && same_outputs &&
					x.Proc.CallingConvention != y.Proc.CallingConvention {
					s_expected := type_to_string(y)
					s_got := type_to_string(x)
					error_line("\tNote: The calling conventions differ between the procedure signature types\n")
					error_line("\t      Expected \"%s\", got \"%s\"\n",
						proc_calling_convention_strings[y.Proc.CallingConvention],
						proc_calling_convention_strings[x.Proc.CallingConvention])
					error_line("\t      Expected: %s\n", s_expected)
					error_line("\t      Got:      %s\n", s_got)
					gb_string_free(s_got)
					gb_string_free(s_expected)
				} else if same_inputs && same_outputs &&
					x.Proc.Diverging != y.Proc.Diverging {
					s_expected := type_to_string(y)
					if y.Proc.Diverging {
						s_expected = gb_string_appendc(s_expected, " -> !")
					}
					s_got := type_to_string(x)
					if x.Proc.Diverging {
						s_got = gb_string_appendc(s_got, " -> !")
					}
					error_line("\tNote: One of the procedures is diverging while the other isn't\n")
					error_line("\t      Expected: %s\n", s_expected)
					error_line("\t      Got:      %s\n", s_got)
					gb_string_free(s_got)
					gb_string_free(s_expected)
				} else if same_inputs && !same_outputs {
					s_expected := type_to_string(y.Proc.Results)
					s_got := type_to_string(x.Proc.Results)
					error_line("\tNote: The return types differ between the procedure signature types\n")
					error_line("\t      Expected: %s\n", s_expected)
					error_line("\t      Got:      %s\n", s_got)
					gb_string_free(s_got)
					gb_string_free(s_expected)
				} else if !same_inputs && same_outputs {
					s_expected := type_to_string(y.Proc.Params)
					s_got := type_to_string(x.Proc.Params)
					error_line("\tNote: The input parameter types differ between the procedure signature types\n")
					error_line("\t      Expected: %s\n", s_expected)
					error_line("\t      Got:      %s\n", s_got)
					gb_string_free(s_got)
					gb_string_free(s_expected)
				} else {
					s_expected := type_to_string(y)
					s_got := type_to_string(x)
					error_line("\tNote: The signature type do not match whatsoever\n")
					error_line("\t      Expected: %s\n", s_expected)
					error_line("\t      Got:      %s\n", s_got)
					gb_string_free(s_got)
					gb_string_free(s_expected)
				}
			}
		}
		operand.Mode = Addressing_Invalid
		return
	}
}

func update_untyped_expr_type(c *CheckerContext, e *Ast, type_ *Type, final bool) {
	gb_assert_handler("Assertion Failure", "e != nil", "cmd_check_expr_assign.go", 0)
	old := check_get_expr_info(c, e)
	if old == nil {
		if type_ != nil && type_ != t_invalid {
			if e.TAV.Type == nil || e.TAV.Type == t_invalid {
				add_type_and_value(c, e, e.TAV.Mode, type_, e.TAV.Value)
				if e.Kind == AstTernaryIfExpr {
					update_untyped_expr_type(c, e.TernaryIfExpr.X, type_, final)
					update_untyped_expr_type(c, e.TernaryIfExpr.Y, type_, final)
				}
			}
		}
		return
	}
	switch e.Kind {
	case AstUnaryExpr:
		ue := &e.UnaryExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		update_untyped_expr_type(c, ue.Expr, type_, final)
	case AstBinaryExpr:
		be := &e.BinaryExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		if token_is_comparison(be.Op.Kind) {
		} else if token_is_shift(be.Op.Kind) {
			update_untyped_expr_type(c, be.Left, type_, final)
		} else {
			update_untyped_expr_type(c, be.Left, type_, final)
			update_untyped_expr_type(c, be.Right, type_, final)
		}
	case AstTernaryIfExpr:
		te := &e.TernaryIfExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		x := make_operand_from_node(te.X)
		y := make_operand_from_node(te.Y)
		if x.Mode != Addressing_Constant || check_is_expressible(c, &x, type_) {
			update_untyped_expr_type(c, te.X, type_, final)
		}
		if y.Mode != Addressing_Constant || check_is_expressible(c, &y, type_) {
			update_untyped_expr_type(c, te.Y, type_, final)
		}
	case AstTernaryWhenExpr:
		te := &e.TernaryWhenExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		update_untyped_expr_type(c, te.X, type_, final)
		update_untyped_expr_type(c, te.Y, type_, final)
	case AstOrReturnExpr:
		ore := &e.OrReturnExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		update_untyped_expr_type(c, ore.Expr, type_, final)
	case AstOrBranchExpr:
		obe := &e.OrBranchExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		update_untyped_expr_type(c, obe.Expr, type_, final)
	case AstOrElseExpr:
		oee := &e.OrElseExpr
		if old.Value.Kind != ExactValue_Invalid {
			break
		}
		update_untyped_expr_type(c, oee.X, type_, final)
		update_untyped_expr_type(c, oee.Y, type_, final)
	case AstParenExpr:
		pe := &e.ParenExpr
		update_untyped_expr_type(c, pe.Expr, type_, final)
	}
	if !final && is_type_untyped(type_) {
		old.Type = base_type(type_)
		return
	}
	check_remove_expr_info(c, e)
	if old.IsLhs && !is_type_integer(type_) {
		expr_str := expr_to_string(e)
		type_str := type_to_string(type_)
		error(e, "Shifted operand %s must be an integer, got %s", expr_str, type_str)
		gb_string_free(type_str)
		gb_string_free(expr_str)
		return
	}
	add_type_and_value(c, e, old.Mode, type_, old.Value)
}

func update_untyped_expr_value(c *CheckerContext, e *Ast, value ExactValue) {
	gb_assert_handler("Assertion Failure", "e != nil", "cmd_check_expr_assign.go", 0)
	found := check_get_expr_info(c, e)
	if found != nil {
		found.Value = value
	}
}

func convert_untyped_error(c *CheckerContext, operand *Operand, target_type *Type, ignore_error_block bool) {
	expr_str := expr_to_string(operand.Expr)
	type_str := type_to_string(target_type)
	from_type_str := type_to_string(operand.Type)
	extra_text := ""
	if operand.Mode == Addressing_Constant {
		if big_int_is_zero(&operand.Value.ValueInteger) {
			// Note: original C++ checks if expression is "nil" to skip suggestion text
			// Cannot easily compare gbString with Go string in pure Go
			extra_text = " - Did you want 'nil'?"
		}
	}
	if !ignore_error_block {
		begin_error_block()
	}
	error(operand.Expr, "Cannot convert untyped value '%s' to '%s' from '%s'%s", expr_str, type_str, from_type_str, extra_text)
	if operand.Value.Kind == ExactValue_String {
		key := operand.Value.ValueString
		if is_type_string(operand.Type) && is_type_enum(target_type) {
			et := base_type(target_type)
			check_did_you_mean_type(key, et.Enum.Fields, ".")
		}
	}
	gb_string_free(from_type_str)
	gb_string_free(type_str)
	gb_string_free(expr_str)
	operand.Mode = Addressing_Invalid
	if !ignore_error_block {
		end_error_block()
	}
}

func convert_exact_value_for_type(v ExactValue, type_ *Type) ExactValue {
	t := core_type(type_)
	if is_type_boolean(t) {
	} else if is_type_float(t) {
		v = exact_value_to_float(v)
	} else if is_type_integer(t) {
		v = exact_value_to_integer(v)
	} else if is_type_pointer(t) {
		v = exact_value_to_integer(v)
	} else if is_type_complex(t) {
		v = exact_value_to_complex(v)
	} else if is_type_quaternion(t) {
		v = exact_value_to_quaternion(v)
	}
	return v
}

func convert_to_typed(c *CheckerContext, operand *Operand, target_type *Type) {
	if target_type == nil || operand.Mode == Addressing_Invalid ||
		operand.Mode == Addressing_Type ||
		is_type_typed(operand.Type) ||
		target_type == t_invalid {
		return
	}
	if is_type_untyped(target_type) {
		gb_assert_handler("Assertion Failure", "operand.Type.Kind == Type_Basic", "cmd_check_expr_assign.go", 0)
		gb_assert_handler("Assertion Failure", "target_type.Kind == Type_Basic", "cmd_check_expr_assign.go", 0)
		x_kind := operand.Type.Basic.Kind
		y_kind := target_type.Basic.Kind
		if is_type_numeric(operand.Type) && is_type_numeric(target_type) {
			if x_kind < y_kind {
				operand.Type = target_type
				update_untyped_expr_type(c, operand.Expr, target_type, false)
			}
		} else if x_kind != y_kind {
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
		return
	}
	t := base_type(target_type)
	if c.in_enum_type {
		t = core_type(target_type)
	}
	switch t.Kind {
	case TypeBasic:
		if operand.Mode == Addressing_Constant {
			check_is_expressible(c, operand, target_type)
			if operand.Mode == Addressing_Invalid {
				return
			}
			update_untyped_expr_value(c, operand.Expr, operand.Value)
		}
		switch operand.Type.Basic.Kind {
		case BasicUntypedBool:
			if !is_type_boolean(target_type) {
				operand.Mode = Addressing_Invalid
				convert_untyped_error(c, operand, target_type, false)
				return
			}
		case BasicUntypedInteger,
			BasicUntypedFloat,
			BasicUntypedComplex,
			BasicUntypedQuaternion,
			BasicUntypedRune:
			if !is_type_numeric(target_type) {
				operand.Mode = Addressing_Invalid
				convert_untyped_error(c, operand, target_type, false)
				return
			}
		case BasicUntypedNil:
			if is_type_any(target_type) {
			} else if is_type_cstring(target_type) {
			} else if is_type_cstring16(target_type) {
			} else if !type_has_nil(target_type) {
				operand.Mode = Addressing_Invalid
				convert_untyped_error(c, operand, target_type, false)
				return
			}
		}
		switch operand.Type.Basic.Kind {
		case BasicUntypedFloat:
			check_update_float_precision(&operand.Value, t)
		}
	case TypeArray:
		elem := base_array_type(t)
		if check_is_assignable_to(c, operand, elem) {
			operand.Mode = Addressing_Value
		} else {
			if operand.Value.Kind == ExactValue_String {
				s := operand.Value.ValueString
				if is_type_u8_array(t) {
					if s.Len == t.Array.Count {
						break
					}
				} else if is_type_rune_array(t) {
					rune_count := gb_utf8_strnlen(s.Text, s.Len)
					if rune_count == t.Array.Count {
						break
					}
				}
			} else if operand.Value.Kind == ExactValue_String16 {
				s := operand.Value.ValueString16
				if is_type_u16_array(t) {
					if s.Len == t.Array.Count {
						break
					}
				}
			}
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
	case TypeSimdVector:
		elem := base_array_type(t)
		if check_is_assignable_to(c, operand, elem) {
			operand.Mode = Addressing_Value
		} else {
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
	case TypeMatrix:
		elem := base_array_type(t)
		if check_is_assignable_to(c, operand, elem) {
			if t.Matrix.RowCount != t.Matrix.ColumnCount {
				operand.Mode = Addressing_Invalid
				begin_error_block()
				convert_untyped_error(c, operand, target_type, true)
				error_line("\tNote: Only a square matrix types can be initialized with a scalar value\n")
				end_error_block()
				return
			} else {
				operand.Mode = Addressing_Value
			}
		} else {
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
	case TypeUnion:
		if !build_context.StrictStyle &&
			operand.Mode == Addressing_Constant &&
			target_type.Kind == TypeNamed &&
			(c.pkg == nil || string_compare(c.pkg.Name, make_string_c("os")) != 0) &&
			string_compare(target_type.Named.Name, make_string_c("Error")) == 0 {
			e := target_type.Named.TypeName
			if e.Pkg != nil && string_compare(e.Pkg.Name, make_string_c("os")) == 0 {
				if is_exact_value_zero(operand.Value) &&
					(operand.Value.Kind == ExactValue_Integer ||
						operand.Value.Kind == ExactValue_Float) {
					operand.Mode = Addressing_Value
					operand.Value = EmptyExactValue
					update_untyped_expr_value(c, operand.Expr, operand.Value)
					break
				}
			}
		}
		if !is_operand_nil(*operand) && !is_operand_uninit(*operand) {
			count := len(t.Union.Variants)
			valids := make([]ValidIndexAndScore, 0, count)
			first_success_index := -1
			for i, vt := range t.Union.Variants {
				var score int64
				if check_is_assignable_to_with_score(c, operand, vt, &score, false, true) {
					valids = append(valids, ValidIndexAndScore{Index: isize(i), Score: score})
					if first_success_index < 0 {
						first_success_index = i
					}
				}
			}
			valid_count := len(valids)
			if valid_count > 1 {
				sort_valids(valids)
				best_score := valids[0].Score
				for i := 1; i < valid_count; i++ {
					v := valids[i]
					if best_score > v.Score {
						valid_count = i
						break
					}
					best_score = v.Score
				}
				first_success_index = int(valids[0].Index)
			}
			type_str := type_to_string(target_type)
			defer gb_string_free(type_str)
			if valid_count == 1 {
				new_type := t.Union.Variants[first_success_index]
				target_type = new_type
				if is_type_union(new_type) {
					convert_to_typed(c, operand, new_type)
					break
				}
				operand.Type = new_type
				if operand.Mode != Addressing_Constant ||
					!elem_type_can_be_constant(operand.Type) {
					operand.Mode = Addressing_Value
				}
				break
			} else if valid_count > 1 {
				gb_assert_handler("Assertion Failure", "first_success_index >= 0", "cmd_check_expr_assign.go", 0)
				begin_error_block()
				operand.Mode = Addressing_Invalid
				convert_untyped_error(c, operand, target_type, true)
				error_line("Ambiguous type conversion to '%s', which variant did you mean:\n\t", type_str)
				j := 0
				for i := 0; i < valid_count; i++ {
					valid := valids[i]
					if j > 0 && valid_count > 2 {
						error_line(", ")
					}
					if j == valid_count-1 {
						if valid_count == 2 {
							error_line(" ")
						}
						error_line("or ")
					}
					str := type_to_string(t.Union.Variants[valid.Index])
					error_line("'%s'", str)
					gb_string_free(str)
					j++
				}
				error_line("\n\n")
				end_error_block()
				return
			} else if is_type_untyped_uninit(operand.Type) {
				target_type = t_untyped_uninit
			} else if !is_type_untyped_nil(operand.Type) || !type_has_nil(target_type) {
				begin_error_block()
				operand.Mode = Addressing_Invalid
				convert_untyped_error(c, operand, target_type, true)
				if count > 0 {
					error_line("'%s' is a union which only excepts the following types:\n", type_str)
					error_line("\t")
					for i := 0; i < count; i++ {
						v := t.Union.Variants[i]
						if i > 0 && count > 2 {
							error_line(", ")
						}
						if i == count-1 {
							if count == 2 {
								error_line(" ")
							}
							if count > 1 {
								error_line("or ")
							}
						}
						str := type_to_string(v)
						error_line("'%s'", str)
						gb_string_free(str)
					}
					error_line("\n\n")
				}
				end_error_block()
				return
			}
		}
		// Fallthrough to default handling for nil/uninit cases
		if is_type_untyped_uninit(operand.Type) {
			target_type = t_untyped_uninit
		} else if is_type_untyped_nil(operand.Type) && type_has_nil(target_type) {
			target_type = t_untyped_nil
		} else {
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
	default:
		if is_type_untyped_uninit(operand.Type) {
			target_type = t_untyped_uninit
		} else if is_type_untyped_nil(operand.Type) && type_has_nil(target_type) {
			target_type = t_untyped_nil
		} else {
			operand.Mode = Addressing_Invalid
			convert_untyped_error(c, operand, target_type, false)
			return
		}
	}
	if is_type_any(target_type) && is_type_untyped(operand.Type) {
		if is_type_untyped_nil(operand.Type) && is_type_untyped_uninit(operand.Type) {
		} else {
			target_type = default_type(operand.Type)
		}
	}
	update_untyped_expr_type(c, operand.Expr, target_type, true)
	operand.Type = target_type
}

func check_index_value(c *CheckerContext, main_type *Type, open_range bool, index_value *Ast, max_count int64, value *int64, type_hint *Type) bool {
	operand := &Operand{Mode: Addressing_Invalid}
	check_expr_with_type_hint(c, operand, index_value, type_hint)
	if operand.Mode == Addressing_Invalid {
		if value != nil {
			*value = 0
		}
		return true
	}
	index_type := t_int
	if type_hint != nil {
		index_type = type_hint
	}
	convert_to_typed(c, operand, index_type)
	if operand.Mode == Addressing_Invalid {
		if value != nil {
			*value = 0
		}
		return false
	}
	if type_hint != nil {
		if !check_is_assignable_to(c, operand, type_hint) {
			expr_str := expr_to_string(operand.Expr)
			index_type_str := type_to_string(type_hint)
			error(operand.Expr, "Index '%s' must be an enum of type '%s'", expr_str, index_type_str)
			gb_string_free(index_type_str)
			gb_string_free(expr_str)
			if value != nil {
				*value = 0
			}
			return false
		}
	} else if !is_type_integer(operand.Type) && !is_type_enum(operand.Type) {
		expr_str := expr_to_string(operand.Expr)
		type_str := type_to_string(operand.Type)
		error(operand.Expr, "Index '%s' must be an integer, got %s", expr_str, type_str)
		gb_string_free(type_str)
		gb_string_free(expr_str)
		if value != nil {
			*value = 0
		}
		return false
	}
	if operand.Mode == Addressing_Constant &&
		(c.StateFlags&StateFlag_no_bounds_check) == 0 {
		i := exact_value_to_integer(operand.Value).ValueInteger
		if i.Sign && !is_type_enum(index_type) && !is_type_multi_pointer(main_type) {
			idx_str := big_int_to_string(temporary_allocator(), &i)
			expr_str := expr_to_string(operand.Expr, temporary_allocator())
			error(operand.Expr, "Index '%s' cannot be a negative value, got %.*s", expr_str, idx_str.Len, idx_str.Data)
			if value != nil {
				*value = 0
			}
			return false
		}
		if max_count >= 0 {
			if is_type_enum(index_type) {
				bt := base_type(index_type)
				gb_assert_handler("Assertion Failure", "bt.Kind == Type_Enum", "cmd_check_expr_assign.go", 0)
				lo := *bt.Enum.MinValue
				hi := *bt.Enum.MaxValue
				lo_str := String{}
				hi_str := String{}
				if len(bt.Enum.Fields) > 0 {
					lo_idx := bt.Enum.MinValueIndex
					if lo_idx < 0 {
						lo_idx = 0
					}
					if lo_idx > len(bt.Enum.Fields)-1 {
						lo_idx = len(bt.Enum.Fields) - 1
					}
					hi_idx := bt.Enum.MaxValueIndex
					if hi_idx < 0 {
						hi_idx = 0
					}
					if hi_idx > len(bt.Enum.Fields)-1 {
						hi_idx = len(bt.Enum.Fields) - 1
					}
					lo_str = bt.Enum.Fields[lo_idx].Token.String
					hi_str = bt.Enum.Fields[hi_idx].Token.String
				}
				out_of_bounds := false
				if compare_exact_values(Token_Lt, operand.Value, lo) || compare_exact_values(Token_Gt, operand.Value, hi) {
					out_of_bounds = true
				}
				if out_of_bounds {
					expr_str := expr_to_string(operand.Expr)
					if lo_str.Len > 0 {
						error(operand.Expr, "Index '%s' is out of bounds range %.*s ..= %.*s", expr_str, lo_str.Len, lo_str.Data, hi_str.Len, hi_str.Data)
					} else {
						index_type_str := type_to_string(index_type)
						error(operand.Expr, "Index '%s' is out of bounds range of enum type %s", expr_str, index_type_str)
						gb_string_free(index_type_str)
					}
					gb_string_free(expr_str)
					return false
				}
				if value != nil {
					*value = exact_value_to_i64(exact_value_sub(operand.Value, lo))
				}
				return true
			} else {
				var v int64 = -1
				if i.Used <= 1 {
					v = big_int_to_i64(&i)
				}
				if value != nil {
					*value = v
				}
				out_of_bounds := false
				if v < 0 {
					out_of_bounds = true
				} else if open_range {
					out_of_bounds = v > max_count
				} else {
					out_of_bounds = v >= max_count
				}
				if out_of_bounds {
					idx_str := big_int_to_string(temporary_allocator(), &i)
					expr_str := expr_to_string(operand.Expr, temporary_allocator())
					error(operand.Expr, "Index '%s' is out of bounds range 0..<%lld, got %.*s", expr_str, max_count, idx_str.Len, idx_str.Data)
					return false
				}
				return true
			}
		} else {
			if value != nil {
				*value = exact_value_to_i64(operand.Value)
			}
			return true
		}
	}
	if value != nil {
		*value = -1
	}
	return true
}

func ternary_compare_types(x, y *Type) bool {
	if is_type_untyped_uninit(x) {
		return true
	} else if is_type_untyped_nil(x) && type_has_nil(y) {
		return true
	} else if is_type_untyped_uninit(y) {
		return true
	} else if is_type_untyped_nil(y) && type_has_nil(x) {
		return true
	}
	return are_types_identical(x, y)
}
