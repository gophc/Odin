package cmd

func check_unroll_range_stmt(ctx *CheckerContext, node *Ast, mod_flags u32) {
	irs := &node.UnrollRangeStmt
	if node.Kind != Ast_UnrollRangeStmt {
		gb_assert_handler("Assertion Failure", "(node)->kind == Ast_UnrollRangeStmt", "check_stmt.cpp", 903, "expected 'Ast_UnrollRangeStmt' got '%s'", ast_strings[node.Kind].text)
	}
	check_open_scope(ctx, node)
	defer check_close_scope(ctx)

	var val0, val1 *Type
	var entities [2]*Entity
	entity_count := 0
	unroll_count := int64(-1)

	if len(irs.args) > 0 {
		if len(irs.args) > 1 {
			error(irs.args[1], "#unroll only supports a single argument for the unroll per loop amount")
		}
		arg := irs.args[0]
		if arg.Kind == Ast_FieldValue {
			error(arg, "#unroll does not yet support named arguments")
			arg = arg.FieldValue.value
		}
		var x Operand
		check_expr(ctx, &x, arg)
		if x.Mode != Addressing_Constant || !is_type_integer(x.Type) {
			s := expr_to_string(x.Expr)
			error(x.Expr, "Expected a constant integer for #unroll, got '%s'", goStr(s))
			gb_string_free(s)
		} else {
			value := exact_value_to_integer(x.Value)
			v := exact_value_to_i64(value)
			if v < 1 {
				error(x.Expr, "Expected a constant integer >= 1 for #unroll, got %d", v)
			} else {
				unroll_count = v
				if v > 1024 {
					error(x.Expr, "Too large of a value for #unroll, got %d, expected <= 1024", v)
				}
			}
		}
	}

	expr := unparen_expr(irs.expr)
	inline_for_depth := exact_value_i64(0)

	if is_ast_range(expr) {
		var x, y Operand
		ok := check_range(ctx, expr, true, &x, &y, &inline_for_depth)
		if !ok {
			goto skip_expr
		}
		val0 = x.Type
		val1 = t_int
	} else {
		var operand Operand
		operand.Mode = Addressing_Invalid
		check_expr_or_type(ctx, &operand, irs.expr)
		if operand.Mode == Addressing_Type {
			if !is_type_enum(operand.Type) {
				t := type_to_string(operand.Type)
				error(operand.Expr, "Cannot iterate over the type '%s'", goStr(t))
				gb_string_free(t)
				goto skip_expr
			} else {
				val0 = operand.Type
				val1 = t_int
				add_type_info_type(ctx, operand.Type)
				bt := base_type(operand.Type)
				inline_for_depth = exact_value_i64(int64(len(bt.Enum.Fields)))
				goto skip_expr
			}
		} else if operand.Mode != Addressing_Invalid {
			t := base_type(operand.Type)
			switch t.Kind {
			case Type_Basic:
				if is_type_string16(t) && t.Basic.Kind != BasicCstring {
					val0 = t_rune
					val1 = t_int
					inline_for_depth = exact_value_i64(int64(len(operand.Value.ValueString)))
					if unroll_count > 0 {
						error(node, "#unroll(%d) does not support strings", unroll_count)
					}
				} else if is_type_string(t) && t.Basic.Kind != BasicCstring {
					val0 = t_rune
					val1 = t_int
					inline_for_depth = exact_value_i64(int64(len(operand.Value.ValueString)))
					if unroll_count > 0 {
						error(node, "#unroll(%d) does not support strings", unroll_count)
					}
				}
			case Type_Array:
				val0 = t.Array.Elem
				val1 = t_int
				if unroll_count > 0 {
					inline_for_depth = exact_value_i64(unroll_count)
				} else {
					inline_for_depth = exact_value_i64(t.Array.Count)
				}
			case Type_EnumeratedArray:
				val0 = t.EnumeratedArray.Elem
				val1 = t.EnumeratedArray.Index
				if unroll_count > 0 {
					error(node, "#unroll(%d) does not support enumerated arrays", unroll_count)
				}
				inline_for_depth = exact_value_i64(t.EnumeratedArray.Count)
			case Type_Slice:
				if unroll_count > 0 {
					val0 = t.Slice.Elem
					val1 = t_int
					inline_for_depth = exact_value_i64(unroll_count)
				}
			case Type_DynamicArray:
				if unroll_count > 0 {
					val0 = t.DynamicArray.Elem
					val1 = t_int
					inline_for_depth = exact_value_i64(unroll_count)
				}
			case Type_FixedCapacityDynamicArray:
				if unroll_count > 0 {
					val0 = t.FixedCapacityDynamicArray.Elem
					val1 = t_int
					inline_for_depth = exact_value_i64(unroll_count)
				}
			}
		}
		if val0 == nil {
			begin_error_block()
			s := expr_to_string(operand.Expr)
			t := type_to_string(operand.Type)
			error(operand.Expr, "Cannot iterate over '%s' of type '%s' in an '#unroll for' statement", goStr(s), goStr(t))
			if is_type_slice(operand.Type) || is_type_dynamic_array(operand.Type) || is_type_fixed_capacity_dynamic_array(operand.Type) {
				error_line("\tSuggestion: An unroll count `#unroll(N)` must be specified with an array of a runtime-known length\n")
			}
			gb_string_free(s)
			gb_string_free(t)
			end_error_block()
		} else if operand.Mode != Addressing_Constant &&
			unroll_count <= 0 &&
			compare_exact_values(Token_CmpEq, inline_for_depth, exact_value_i64(0)) {
			error(operand.Expr, "An '#unroll for' expression must be known at compile time")
		}
	}

skip_expr:
	lhs := [2]*Ast{irs.val0, irs.val1}
	rhs := [2]*Type{val0, val1}
	for i := 0; i < 2; i++ {
		if lhs[i] == nil {
			continue
		}
		name := lhs[i]
		type_ := rhs[i]
		var entity *Entity
		if name.Kind == Ast_Ident {
			token := name.Ident.token
			str := token.String
			var found *Entity
			if !is_blank_ident_str(str) {
				found = scope_lookup_current(ctx.scope, name.Ident.interned, name.Ident.hash)
			}
			if found == nil {
				entity = alloc_entity_variable(ctx.scope, token, type_, EntityState_Resolved)
				entity.Flags |= EntityFlag_Value
				add_entity_definition(&ctx.checker.info, name, entity)
			} else {
				pos := found.Token.Pos
				error(token, "Redeclaration of '%s' in this scope\n"+
					"\tat %s", goStr(str), goStr(token_pos_to_string(pos)))
				entity = found
			}
		} else {
			error_var_decl_identifier(name)
		}
		if entity == nil {
			entity = alloc_entity_dummy_variable(builtin_pkg.scope, ast_token(name))
		}
		entities[entity_count] = entity
		entity_count++
		if type_ == nil {
			entity.Type = t_invalid
			entity.Flags |= EntityFlag_Used
		}
	}
	for i := 0; i < entity_count; i++ {
		add_entity(ctx, ctx.scope, entities[i].identifier, entities[i])
	}

	prev_inline_for_depth := ctx.inline_for_depth
	defer func() { ctx.inline_for_depth = prev_inline_for_depth }()
	{
		v := exact_value_to_i64(inline_for_depth)
		if v > 0 {
			if ctx.inline_for_depth > 1 {
				ctx.inline_for_depth = ctx.inline_for_depth * v
			} else {
				ctx.inline_for_depth = 1 * v
			}
		}
		if ctx.inline_for_depth >= 1024 && prev_inline_for_depth < 1024 {
			begin_error_block()
			if prev_inline_for_depth > 0 {
				error(node, "Nested '#unroll for' loop cannot be inlined as it exceeds the maximum '#unroll for' depth (%d levels >= %d maximum levels)", v, 1024)
			} else {
				error(node, "'#unroll for' loop cannot be inlined as it exceeds the maximum '#unroll for' depth (%d levels >= %d maximum levels)", v, 1024)
			}
			error_line("\tUse a normal 'for' loop instead by removing the 'inline' prefix\n")
			ctx.inline_for_depth = 1024
			end_error_block()
		}
	}
	new_flags := mod_flags &^ Stmt_BreakAllowed &^ Stmt_ContinueAllowed
	check_stmt(ctx, irs.body, new_flags)
}

func check_range_stmt(ctx *CheckerContext, node *Ast, mod_flags u32) {
	rs := &node.RangeStmt
	if node.Kind != Ast_RangeStmt {
		gb_assert_handler("Assertion Failure", "(node)->kind == Ast_RangeStmt", "check_stmt.cpp", 1724, "expected 'Ast_RangeStmt' got '%s'", ast_strings[node.Kind].text)
	}
	check_open_scope(ctx, node)
	check_label(ctx, rs.label, node)
	var init Operand
	if rs.init != nil {
		check_stmt(ctx, rs.init, mod_flags)
	}
	new_flags := mod_flags | Stmt_BreakAllowed | Stmt_ContinueAllowed

	vals := make([]*Type, 0, 2)
	entities := make([]*Entity, 0, 2)
	is_map := false
	is_bit_set := false
	is_soa := false
	is_reverse := rs.reverse
	expr := unparen_expr(rs.expr)
	rhs_operand := Operand{}
	is_range := false
	is_possibly_addressable := true
	max_val_count := 2

	if is_ast_range(expr) {
		var x, y Operand
		is_possibly_addressable = false
		is_range = true
		ok := check_range(ctx, expr, true, &x, &y, nil)
		if !ok {
			goto skip_expr_range_stmt
		}
		vals = append(vals, x.Type)
		vals = append(vals, t_int)
		if is_reverse {
			error(node, "#reverse for is not supported with ranges, prefer an explicit for loop with init, condition, and post arguments")
		}
	} else {
		var operand Operand
		operand.Mode = Addressing_Invalid
		check_expr_base(ctx, &operand, expr, nil)
		error_operand_no_value(&operand)
		if operand.Mode == Addressing_Type {
			if !is_type_enum(operand.Type) {
				t := type_to_string(operand.Type)
				error(operand.Expr, "Cannot iterate over the type '%s'", goStr(t))
				gb_string_free(t)
				goto skip_expr_range_stmt
			} else {
				is_possibly_addressable = false
				if is_reverse {
					error(node, "#reverse for is not supported for enum types")
				}
				vals = append(vals, operand.Type)
				vals = append(vals, t_int)
				add_type_info_type(ctx, operand.Type)
				if buildContext.NoRTTI {
					error(node, "Iteration over an enum type is not allowed runtime type information (RTTI) has been disallowed")
				}
				goto skip_expr_range_stmt
			}
		} else if operand.Mode != Addressing_Invalid {
			if operand.Mode == Addressing_OptionalOk || operand.Mode == Addressing_OptionalOkPtr {
				expr2 := unparen_expr(operand.Expr)
				if expr2.Kind != Ast_TypeAssertion {
					var end_type *Type
					check_promote_optional_ok(ctx, &operand, nil, &end_type, false)
					if is_type_boolean(end_type) {
						check_promote_optional_ok(ctx, &operand, nil, &end_type, true)
					}
				}
			}
			is_ptr := is_type_pointer(operand.Type)
			t := base_type(type_deref(operand.Type))
			switch t.Kind {
			case Type_Basic:
				if t.Basic.Kind == BasicString16 {
					is_possibly_addressable = false
					vals = append(vals, t_rune)
					vals = append(vals, t_int)
					if is_reverse {
						add_package_dependency(ctx, "runtime", "string16_decode_last_rune")
					} else {
						add_package_dependency(ctx, "runtime", "string16_decode_rune")
					}
				} else if t.Basic.Kind == BasicString || t.Basic.Kind == BasicUntypedString {
					is_possibly_addressable = false
					vals = append(vals, t_rune)
					vals = append(vals, t_int)
					if is_reverse {
						add_package_dependency(ctx, "runtime", "string_decode_last_rune")
					} else {
						add_package_dependency(ctx, "runtime", "string_decode_rune")
					}
				}
			case Type_BitSet:
				vals = append(vals, t.BitSet.Elem)
				max_val_count = 1
				is_bit_set = true
				is_possibly_addressable = false
				add_type_info_type(ctx, operand.Type)
				if buildContext.NoRTTI && is_type_enum(t.BitSet.Elem) {
					error(node, "Iteration over a bit_set of an enum is not allowed runtime type information (RTTI) has been disallowed")
				}
				if len(rs.vals) == 1 && rs.vals[0] != nil && rs.vals[0].Kind == Ast_Ident {
					ident := &rs.vals[0].Ident
					found := scope_lookup(ctx.scope, ident.interned, ident.hash)
					if found != nil && are_types_identical(found.Type, t.BitSet.Elem) {
						begin_error_block()
						name := ident.Token.String
						s := expr_to_string(expr)
						error(rs.vals[0], "'%s' shadows a previous declaration which might be ambiguous with 'for (%s in %s)'", goStr(name), goStr(name), goStr(s))
						error_line("\tSuggestion: Use a different identifier if iteration is wanted, or surround in parentheses if a normal for loop is wanted\n")
						gb_string_free(s)
						end_error_block()
					}
				}
			case Type_EnumeratedArray:
				is_possibly_addressable = operand.Mode == Addressing_Variable || is_ptr
				vals = append(vals, t.EnumeratedArray.Elem)
				vals = append(vals, t.EnumeratedArray.Index)
			case Type_Array:
				is_possibly_addressable = operand.Mode == Addressing_Variable || is_ptr
				vals = append(vals, t.Array.Elem)
				vals = append(vals, t_int)
			case Type_FixedCapacityDynamicArray:
				is_possibly_addressable = operand.Mode == Addressing_Variable || is_ptr
				vals = append(vals, t.FixedCapacityDynamicArray.Elem)
				vals = append(vals, t_int)
			case Type_DynamicArray:
				is_possibly_addressable = true
				vals = append(vals, t.DynamicArray.Elem)
				vals = append(vals, t_int)
			case Type_Slice:
				is_possibly_addressable = true
				vals = append(vals, t.Slice.Elem)
				vals = append(vals, t_int)
			case Type_Map:
				is_possibly_addressable = true
				is_map = true
				vals = append(vals, t.Map.Key)
				vals = append(vals, t.Map.Value)
				if is_reverse {
					error(node, "#reverse for is not supported for map types, as maps are unordered")
				}
				if len(rs.vals) == 1 && rs.vals[0] != nil && rs.vals[0].Kind == Ast_Ident {
					ident := &rs.vals[0].Ident
					found := scope_lookup(ctx.scope, ident.interned, ident.hash)
					if found != nil && are_types_identical(found.Type, t.Map.Key) {
						begin_error_block()
						name := ident.Token.String
						s := expr_to_string(expr)
						error(rs.vals[0], "'%s' shadows a previous declaration which might be ambiguous with 'for (%s in %s)'", goStr(name), goStr(name), goStr(s))
						error_line("\tSuggestion: Use a different identifier if iteration is wanted, or surround in parentheses if a normal for loop is wanted\n")
						gb_string_free(s)
						end_error_block()
					}
				}
			case Type_Tuple:
				is_possibly_addressable = false
				count := len(t.Tuple.Variables)
				if count < 1 {
					begin_error_block()
					check_not_tuple(ctx, &operand)
					error_line("\tMultiple return valued parameters in a range statement are limited to a minimum of 1 usable values with a trailing boolean for the conditional, got %d\n", count)
					end_error_block()
					break
				}
				const MAXIMUM_COUNT = 20
				if count > MAXIMUM_COUNT {
					begin_error_block()
					check_not_tuple(ctx, &operand)
					error_line("\tMultiple return valued parameters in a range statement are limited to a maximum of %d usable values with a trailing boolean for the conditional, got %d\n", MAXIMUM_COUNT, count)
					end_error_block()
					break
				}
				cond_type := t.Tuple.Variables[count-1].Type
				if !is_type_boolean(cond_type) {
					s := type_to_string(cond_type)
					error(operand.Expr, "The final type of %d-valued expression must be a boolean, got %s", count, goStr(s))
					gb_string_free(s)
					break
				}
				max_val_count = count - 1
				for _, e := range t.Tuple.Variables {
					vals = append(vals, e.Type)
				}
				do_break := false
				for j := len(rs.vals) - 1; j >= 0; j-- {
					if rs.vals[j] != nil && count < j+2 {
						s := type_to_string(t)
						error(operand.Expr, "Expected a %d-valued expression on the rhs, got (%s)", j+2, goStr(s))
						gb_string_free(s)
						do_break = true
						break
					}
				}
				if do_break {
					break
				}
				if is_reverse {
					error(node, "#reverse for is not supported for multiple return valued parameters")
				}
				expr2 := unparen_expr(operand.Expr)
				if expr2.Kind == Ast_CallExpr {
					p := base_type(type_of_expr(expr2.CallExpr.proc))
					if p != nil && p.Kind == Type_Proc {
						if p.Proc.RequireResults {
							if len(rs.vals) < max_val_count {
								start := ast_token(rs.vals[0]).Pos
								end := ast_end_pos(rs.vals[len(rs.vals)-1])
								plural := ""
								if max_val_count != 1 {
									plural = "s"
								}
								error_range(start, end, "Expected %d identifier%s, got %d", max_val_count, plural, len(rs.vals))
							}
						}
					}
				}
			case Type_Struct:
				if t.Struct.SoaKind != StructSoaNone {
					if t.Struct.SoaKind == StructSoaFixed {
						is_possibly_addressable = operand.Mode == Addressing_Variable || is_ptr
					} else {
						is_possibly_addressable = true
					}
					is_soa = true
					vals = append(vals, t.Struct.SoaElem)
					vals = append(vals, t_int)
				}
			}
		}
		if len(vals) == 0 || vals[0] == nil {
			s := expr_to_string(operand.Expr)
			t2 := type_to_string(operand.Type)
			begin_error_block()
			error(operand.Expr, "Cannot iterate over '%s' of type '%s'", goStr(s), goStr(t2))
			if len(rs.vals) == 1 {
				t3 := type_deref(operand.Type)
				if t3 != nil && (is_type_map(t3) || is_type_bit_set(t3)) {
					v := expr_to_string(rs.vals[0])
					error_line("\tSuggestion: place parentheses around the expression\n")
					error_line("\t            for (%s in %s) {\n", goStr(v), goStr(s))
					gb_string_free(v)
				}
			}
			gb_string_free(s)
			gb_string_free(t2)
			end_error_block()
		}
	}

skip_expr_range_stmt:
	if len(rs.vals) > max_val_count {
		plural := ""
		if max_val_count != 1 {
			plural = "s"
		}
		error(rs.vals[max_val_count], "Expected a maximum of %d identifier%s, got %d", max_val_count, plural, len(rs.vals))
	}

	rhs := vals
	lhs := make([]*Ast, len(rhs))
	copy(lhs, rs.vals)
	addressable_index := isize(0)
	if is_map {
		addressable_index = 1
	}
	for i := 0; i < len(rhs); i++ {
		if lhs[i] == nil {
			continue
		}
		name := lhs[i]
		type_ := rhs[i]
		var entity *Entity
		is_addressed := false
		if name.Kind == Ast_UnaryExpr && name.UnaryExpr.op.Kind == Token_And {
			is_addressed = true
			name = name.UnaryExpr.expr
		}
		if name.Kind == Ast_Ident {
			token := name.Ident.token
			str := token.String
			var found *Entity
			if !is_blank_ident_str(str) {
				found = scope_lookup_current(ctx.scope, name.Ident.interned, name.Ident.hash)
			}
			if found == nil {
				entity = alloc_entity_variable(ctx.scope, token, type_, EntityState_Resolved)
				if !is_range {
					entity.Flags |= EntityFlag_ForValue
				}
				entity.Flags |= EntityFlag_Value
				entity.identifier = name
				entity.Variable.for_loop_parent_type = type_of_expr(expr)
				if is_addressed {
					if is_possibly_addressable && i == addressable_index {
						entity.Flags &^= EntityFlag_Value
					} else {
						idx_name := "element"
						if is_map {
							idx_name = "key"
						} else if is_bit_set || i == 0 {
							idx_name = "element"
						} else {
							idx_name = "index"
						}
						error(token, "The %s variable '%s' cannot be made addressable", idx_name, goStr(str))
					}
				}
				if is_soa {
					if i == 0 {
						entity.Flags |= EntityFlag_SoaPtrField
					}
				}
				add_entity_definition(&ctx.checker.info, name, entity)
			} else {
				pos := found.Token.Pos
				error(token, "Redeclaration of '%s' in this scope\n"+
					"\tat %s", goStr(str), goStr(token_pos_to_string(pos)))
				entity = found
			}
		} else {
			error_var_decl_identifier(name)
		}
		if entity == nil {
			entity = alloc_entity_dummy_variable(builtin_pkg.scope, ast_token(name))
			entity.identifier = name
		}
		entities = append(entities, entity)
		if type_ == nil {
			entity.Type = t_invalid
			entity.Flags |= EntityFlag_Used
		}
	}
	for _, e := range entities {
		d := decl_info_of_entity(e)
		if d != nil {
			gb_assert_handler("Assertion Failure", "d == nil", "check_stmt.cpp", 2094, 0)
		}
		add_entity(ctx, ctx.scope, e.identifier, e)
		d = make_decl_info(ctx.scope, ctx.decl)
		add_entity_and_decl_info(ctx, e.identifier, e, d)
	}
	check_stmt(ctx, rs.body, new_flags)
	check_close_scope(ctx)
}

func check_using_stmt_entity(ctx *CheckerContext, us *Ast, expr *Ast, is_selector bool, e *Entity) bool {
	if e == nil {
		if is_blank_ident(expr) {
			error(us.UsingStmt.token, "'using' in a statement is not allowed with the blank identifier '_'")
		} else {
			error(us.UsingStmt.token, "'using' applied to an unknown entity")
		}
		return true
	}
	add_entity_use(ctx, expr, e)
	begin_error_block()
	defer end_error_block()

	switch e.Kind {
	case Entity_TypeName:
		t := base_type(e.Type)
		if t.Kind == Type_Enum {
			for _, f := range t.Enum.Fields {
				if !is_entity_exported(f) {
					continue
				}
				found := scope_insert(ctx.scope, f)
				if found != nil {
					expr_str := expr_to_string(expr)
					error(us.UsingStmt.token, "Namespace collision while 'using' enum '%s' of: %s", goStr(expr_str), goStr(found.Token.String))
					gb_string_free(expr_str)
					return false
				}
				f.UsingParent = e
			}
		} else {
			error(us.UsingStmt.token, "'using' can be only applied to enum type entities")
		}

	case Entity_ImportName:
		scope := e.ImportName.Scope
		rw_mutex_lock(&scope.mutex)
		for i := u32(0); i < scope.elements.cap; i++ {
			if scope.elements.slots[i].hash == 0 {
				continue
			}
			decl := scope.elements.slots[i].value
			if !is_entity_exported(decl, true) {
				continue
			}
			hash := scope.elements.slots[i].hash
			interned := scope.elements.keys[i]
			found := scope_insert_with_name(ctx.scope, interned, hash, decl)
			if found != nil {
				expr_str := expr_to_string(expr)
				error(us.UsingStmt.token,
					"Namespace collision while 'using' import name '%s' of: %s\n"+
						"\tat %s\n"+
						"\tat %s",
					goStr(expr_str), goStr(found.Token.String),
					goStr(token_pos_to_string(found.Token.Pos)),
					goStr(token_pos_to_string(decl.Token.Pos)))
				gb_string_free(expr_str)
				rw_mutex_unlock(&scope.mutex)
				return false
			}
		}
		rw_mutex_unlock(&scope.mutex)

	case Entity_Variable:
		is_ptr := is_type_pointer(e.Type)
		t := base_type(type_deref(e.Type))
		if t.Kind == Type_Struct {
			wait_signal_until_available(&t.Struct.FieldsWaitSignal)
			found := t.Struct.Scope
			if found == nil {
				gb_assert_handler("Assertion Failure", "found != nil", "check_stmt.cpp", 829, 0)
			}
			for iter := beginScopeMap(&found.elements); ; {
				_, f, ok := ScopeMapIteratorNext(&iter)
				if !ok {
					break
				}
				if f.Kind == Entity_Variable {
					uvar := alloc_entity_using_variable(e, f.Token, f.Type, expr)
					if !is_ptr && e.Flags&EntityFlag_Value != 0 {
						uvar.Flags |= EntityFlag_Value
					}
					if e.Flags&EntityFlag_Param != 0 {
						uvar.Flags |= EntityFlag_Param
					}
					if e.Flags&EntityFlag_SoaPtrField != 0 {
						uvar.Flags |= EntityFlag_SoaPtrField
					}
					prev := scope_insert(ctx.scope, uvar)
					if prev != nil {
						expr_str := expr_to_string(expr)
						error(us.UsingStmt.token, "Namespace collision while using '%s' of: '%s'", goStr(expr_str), goStr(prev.Token.String))
						gb_string_free(expr_str)
						return false
					}
				}
			}
		} else {
			error(us.UsingStmt.token, "'using' can only be applied to variables of type 'struct'")
			return false
		}

	case Entity_Constant:
		error(us.UsingStmt.token, "'using' cannot be applied to a constant")

	case Entity_Procedure, Entity_ProcGroup, Entity_Builtin:
		error(us.UsingStmt.token, "'using' cannot be applied to a procedure")

	case Entity_Nil:
		error(us.UsingStmt.token, "'using' cannot be applied to 'nil'")

	case Entity_Label:
		error(us.UsingStmt.token, "'using' cannot be applied to a label")

	case Entity_Invalid:
		error(us.UsingStmt.token, "'using' cannot be applied to an invalid entity")

	default:
		gb_assert_handler("Panic", "0", "check_stmt.cpp", 877, "TODO(bill): 'using' other expressions?")
	}

	return true
}
