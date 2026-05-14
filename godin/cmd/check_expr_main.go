package cmd

func gb_string_append_rune(str gbString, r byte) gbString {
	return gb_string_append_length(str, &r, 1)
}

func check_cycle(c *CheckerContext, curr *Entity, report bool) bool {
	if curr.State != EntityState_InProgress {
		return false
	}
	tp := *c.TypePath
	for i := 0; i < len(tp); i++ {
		prev := tp[i]
		if prev == curr {
			if report {
				error(curr.Token, "Illegal declaration cycle of `%.*s`", curr.Token.String.Len, curr.Token.String.Data)
				for j := i; j < len(tp); j++ {
					e := tp[j]
					error(e.Token, "\t%.*s refers to", e.Token.String.Len, e.Token.String.Data)
				}
				error(curr.Token, "\t%.*s", curr.Token.String.Len, curr.Token.String.Data)
				curr.Type = t_invalid
			}
			return true
		}
	}
	return false
}

func check_expr_as_value_for_ternary(c *CheckerContext, o *Operand, e *Ast, type_hint *Type) {
	check_expr_base(c, o, e, type_hint)
	check_not_tuple(c, o)
	error_operand_no_value(o)
	switch o.Mode {
	case Addressing_Type:
		begin_error_block()
		defer end_error_block()
		expr_str := expr_to_string(o.Expr)
		defer gb_string_free(expr_str)
		error(o.Expr, "A type '%s' cannot be used as a runtime value", goStr(expr_str))
		error_line("\tSuggestion: If a runtime 'typeid' is wanted, use 'typeid_of' to convert a type\n")
		o.Mode = Addressing_Invalid
	case Addressing_Builtin:
		begin_error_block()
		defer end_error_block()
		expr_str := expr_to_string(o.Expr)
		defer gb_string_free(expr_str)
		error(o.Expr, "A built-in procedure '%s' cannot be used as a runtime value", goStr(expr_str))
		error_line("\tNote: Built-in procedures are implemented by the compiler and might not be actually instantiated procedures\n")
		o.Mode = Addressing_Invalid
	case Addressing_ProcGroup:
		begin_error_block()
		defer end_error_block()
		expr_str := expr_to_string(o.Expr)
		defer gb_string_free(expr_str)
		error(o.Expr, "Cannot use overloaded procedure '%s' as a runtime value", goStr(expr_str))
		error_line("\tNote: Please specify which procedure in the procedure group to use, via cast or type inference\n")
		o.Mode = Addressing_Invalid
	}
}

func check_ternary_if_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := ExprKind_Expr
	var cond Operand
	cond.Mode = Addressing_Invalid
	gb_assert_handler("Assertion Failure", "node.Kind == Ast_TernaryIfExpr", "cmd_check_expr_main.go", 0)
	te := &node.TernaryIfExpr
	check_expr(c, &cond, te.Cond)
	node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.Cond.ViralStateFlags.Load())
	if cond.Mode != Addressing_Invalid && !is_type_boolean(cond.Type) {
		error(te.Cond, "Non-boolean condition in ternary if expression")
	}
	var x Operand
	x.Mode = Addressing_Invalid
	var y Operand
	y.Mode = Addressing_Invalid
	check_expr_as_value_for_ternary(c, &x, te.X, type_hint)
	node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.X.ViralStateFlags.Load())
	if te.Y != nil {
		th := type_hint
		if type_hint == nil && is_type_typed(x.Type) {
			th = x.Type
		}
		check_expr_as_value_for_ternary(c, &y, te.Y, th)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.Y.ViralStateFlags.Load())
	} else {
		error(node, "A ternary expression must have an else clause")
		return kind
	}
	if x.Mode == Addressing_Type || y.Mode == Addressing_Type {
		type_expr := x.Expr
		if y.Mode == Addressing_Type {
			type_expr = y.Expr
		}
		type_string := expr_to_string(type_expr)
		defer gb_string_free(type_string)
		error(node, "Type %s is invalid operand for ternary if expression", goStr(type_string))
		return kind
	}
	use_type_hint := type_hint != nil && (is_operand_nil(x) || is_operand_nil(y))
	convert_to_typed(c, &x, mapValue(use_type_hint && type_hint != nil, type_hint, y.Type))
	if x.Mode == Addressing_Invalid {
		return kind
	}
	convert_to_typed(c, &y, mapValue(use_type_hint && type_hint != nil, type_hint, x.Type))
	if y.Mode == Addressing_Invalid {
		x.Mode = Addressing_Invalid
		return kind
	}
	if x.Mode == Addressing_Builtin && y.Mode == Addressing_Builtin {
		if type_hint == nil {
			error(node, "Built-in procedures cannot be used within a ternary expression since they have no well-defined signature")
			return kind
		}
	}
	if x.Mode == Addressing_ProcGroup && y.Mode == Addressing_ProcGroup {
		if type_hint == nil {
			error(node, "Procedure groups cannot be used within a ternary expression since they have no well-defined signature that can be inferred without a context")
			return kind
		}
	}
	if type_hint != nil && !is_type_any(type_hint) {
		if check_is_assignable_to(c, &x, type_hint) && check_is_assignable_to(c, &y, type_hint) {
			check_cast(c, &x, type_hint)
			check_cast(c, &y, type_hint)
		}
	}
	if !ternary_compare_types(x.Type, y.Type) {
		its := type_to_string(x.Type)
		defer gb_string_free(its)
		ets := type_to_string(y.Type)
		defer gb_string_free(ets)
		error(node, "Mismatched types in ternary if expression, %s vs %s", goStr(its), goStr(ets))
		return kind
	}
	o.Type = x.Type
	if is_type_untyped_nil(o.Type) || is_type_untyped_uninit(o.Type) {
		o.Type = y.Type
	}
	o.Mode = Addressing_Value
	o.Expr = node
	if type_hint != nil && is_type_untyped(o.Type) && !is_type_any(type_hint) {
		if check_cast_internal(c, &x, type_hint) &&
			check_cast_internal(c, &y, type_hint) {
			convert_to_typed(c, o, type_hint)
			update_untyped_expr_type(c, node, type_hint, !is_type_untyped(type_hint))
			o.Type = type_hint
		}
	}
	return kind
}

func mapValue[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func check_ternary_when_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := ExprKind_Expr
	var cond Operand
	gb_assert_handler("Assertion Failure", "node.Kind == Ast_TernaryWhenExpr", "cmd_check_expr_main.go", 0)
	te := &node.TernaryWhenExpr
	check_expr(c, &cond, te.Cond)
	node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.Cond.ViralStateFlags.Load())
	if cond.Mode != Addressing_Constant || !is_type_boolean(cond.Type) {
		error(te.Cond, "Expected a constant boolean condition in ternary when expression")
		return kind
	}
	if cond.Value.ValueBool {
		check_expr_or_type(c, o, te.X, type_hint)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.X.ViralStateFlags.Load())
	} else {
		if te.Y != nil {
			check_expr_or_type(c, o, te.Y, type_hint)
			node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.Y.ViralStateFlags.Load())
		} else {
			error(node, "A ternary when expression must have an else clause")
			return kind
		}
	}
	return kind
}

func check_or_else_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	gb_assert_handler("Assertion Failure", "node.Kind == Ast_OrElseExpr", "cmd_check_expr_main.go", 0)
	oe := &node.OrElseExpr
	name := oe.Token.String
	arg := oe.X
	default_value := oe.Y
	var x Operand
	var y Operand

	if is_load_directive_call(arg) {
		res := check_load_directive(c, &x, arg, type_hint, false)
		if !(is_load_directive_call(default_value) && res == LoadDirective_Success) {
			y_is_diverging := false
			check_expr_base(c, &y, default_value, x.Type)
			switch y.Mode {
			case Addressing_NoValue:
				if is_diverging_expr(y.Expr) {
					y.Mode = Addressing_Value
					y_is_diverging = true
				} else {
					error_operand_no_value(&y)
					y.Mode = Addressing_Invalid
				}
			case Addressing_Type:
				error_operand_not_expression(&y)
				y.Mode = Addressing_Invalid
			}
			if y.Mode == Addressing_Invalid {
				o.Mode = Addressing_Value
				o.Type = t_invalid
				o.Expr = node
				return ExprKind_Expr
			}
			if !y_is_diverging {
				check_assignment(c, &y, x.Type, name)
				if y.Mode != Addressing_Constant {
					error(y.Expr, "expected a constant expression on the right-hand side of 'or_else' in conjuction with '#load'")
				}
			}
		}
		if res == LoadDirective_Success {
			*o = x
		} else {
			*o = y
		}
		o.Expr = node
		return ExprKind_Expr
	}

	check_multi_expr_with_type_hint(c, &x, arg, type_hint)
	if x.Mode == Addressing_Invalid {
		o.Mode = Addressing_Value
		o.Type = t_invalid
		o.Expr = node
		return ExprKind_Expr
	}
	var left_type *Type
	var right_type *Type
	check_or_else_split_types(c, &x, name, &left_type, &right_type)
	add_type_and_value(c, arg, x.Mode, x.Type, x.Value)
	y_is_diverging := false
	check_expr_base(c, &y, default_value, left_type)
	switch y.Mode {
	case Addressing_NoValue:
		if is_diverging_expr(y.Expr) {
			y.Mode = Addressing_Value
			y_is_diverging = true
		} else {
			error_operand_no_value(&y)
			y.Mode = Addressing_Invalid
		}
	case Addressing_Type:
		error_operand_not_expression(&y)
		y.Mode = Addressing_Invalid
	}
	if y.Mode == Addressing_Invalid {
		o.Mode = Addressing_Value
		o.Type = t_invalid
		o.Expr = node
		return ExprKind_Expr
	}
	if left_type != nil {
		if !y_is_diverging {
			if is_type_tuple(left_type) {
				if !is_type_tuple(y.Type) {
					error(y.Expr, "Found a single value where a %td-valued expression was expected", len(left_type.Tuple.Variables))
				} else if !are_types_identical(left_type, y.Type) {
					xt := type_to_string(left_type)
					defer gb_string_free(xt)
					yt := type_to_string(y.Type)
					defer gb_string_free(yt)
					error(y.Expr, "Mismatched types, expected (%s), got (%s)", goStr(xt), goStr(yt))
				}
			} else {
				check_assignment(c, &y, left_type, name)
			}
		}
	} else {
		check_or_else_expr_no_value_error(c, name, x, type_hint)
	}
	if left_type == nil {
		left_type = t_invalid
	}
	o.Mode = Addressing_Value
	o.Type = left_type
	o.Expr = node
	return ExprKind_Expr
}

func check_or_return_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	gb_assert_handler("Assertion Failure", "node.Kind == Ast_OrReturnExpr", "cmd_check_expr_main.go", 0)
	re := &node.OrReturnExpr
	name := re.Token.String
	var x Operand
	check_multi_expr_with_type_hint(c, &x, re.Expr, type_hint)
	if x.Mode == Addressing_Invalid {
		o.Mode = Addressing_Value
		o.Type = t_invalid
		o.Expr = node
		return ExprKind_Expr
	}
	var left_type *Type
	var right_type *Type
	check_or_return_split_types(c, &x, name, &left_type, &right_type)
	add_type_and_value(c, re.Expr, x.Mode, x.Type, x.Value)
	if right_type == nil {
		check_or_else_expr_no_value_error(c, name, x, type_hint)
	} else {
		proc_type := base_type(c.CurrProcSig)
		gb_assert_handler("Assertion Failure", "proc_type.Kind == Type_Proc", "cmd_check_expr_main.go", 0)
		result_type := proc_type.Proc.Results
		if result_type == nil {
			error(node, "'%.*s' requires the current procedure to have at least one return value", name.Len, name.Data)
		} else {
			gb_assert_handler("Assertion Failure", "result_type.Kind == Type_Tuple", "cmd_check_expr_main.go", 0)
			vars := result_type.Tuple.Variables
			end_type := vars[len(vars)-1].Type
			if len(vars) > 1 {
				if !proc_type.Proc.HasNamedResults {
					error(node, "'%.*s' within a procedure with more than 1 return value requires that the return values are named, allowing for early return", name.Len, name.Data)
				}
			}
			var rhs Operand
			rhs.Type = right_type
			rhs.Mode = Addressing_Value
			if is_type_boolean(right_type) && is_type_boolean(end_type) {
			} else if !check_is_assignable_to(c, &rhs, end_type) {
				begin_error_block()
				defer end_error_block()
				a := type_to_string(right_type)
				defer gb_string_free(a)
				b := type_to_string(end_type)
				defer gb_string_free(b)
				ret_type := type_to_string(result_type)
				defer gb_string_free(ret_type)
				error(node, "Cannot assign end value of type '%s' to '%s' in '%.*s'", goStr(a), goStr(b), name.Len, name.Data)
				if len(vars) == 1 {
					error_line("\tProcedure return value type: %s\n", goStr(ret_type))
				} else {
					error_line("\tProcedure return value types: (%s)\n", goStr(ret_type))
				}
			}
		}
	}
	o.Expr = node
	o.Type = left_type
	if left_type != nil {
		o.Mode = Addressing_Value
	} else {
		o.Mode = Addressing_NoValue
	}
	if c.CurrProcSig == nil {
		error(node, "'%.*s' can only be used within a procedure", name.Len, name.Data)
	}
	if c.InDefer {
		error(node, "'or_return' cannot be used within a defer statement")
	}
	return ExprKind_Expr
}

func check_or_branch_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	gb_assert_handler("Assertion Failure", "node.Kind == Ast_OrBranchExpr", "cmd_check_expr_main.go", 0)
	be := &node.OrBranchExpr
	name := be.Token.String
	var x Operand
	check_multi_expr_with_type_hint(c, &x, be.Expr, type_hint)
	if x.Mode == Addressing_Invalid {
		o.Mode = Addressing_Value
		o.Type = t_invalid
		o.Expr = node
		return ExprKind_Expr
	}
	var left_type *Type
	var right_type *Type
	check_or_return_split_types(c, &x, name, &left_type, &right_type)
	add_type_and_value(c, be.Expr, x.Mode, x.Type, x.Value)
	if right_type == nil {
		check_or_else_expr_no_value_error(c, name, x, type_hint)
	} else {
		if is_type_boolean(right_type) || type_has_nil(right_type) {
		} else {
			s := type_to_string(right_type)
			defer gb_string_free(s)
			error(node, "'%.*s' requires a boolean or nil-able type, got %s", name.Len, name.Data, goStr(s))
		}
	}
	o.Expr = node
	o.Type = left_type
	if left_type != nil {
		o.Mode = Addressing_Value
	} else {
		o.Mode = Addressing_NoValue
	}
	if c.CurrProcSig == nil {
		error(node, "'%.*s' can only be used within a procedure", name.Len, name.Data)
	}
	label := be.Label
	switch be.Token.Kind {
	case Token_or_break:
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | uint32(ViralStateFlag_ContainsOrBreak))
		if (c.StmtFlags&StmtFlag_BreakAllowed) == 0 && label == nil {
			error(be.Token, "'%.*s' only allowed in non-inline loops or 'switch' statements", name.Len, name.Data)
		}
	case Token_or_continue:
		if (c.StmtFlags&StmtFlag_ContinueAllowed) == 0 && label == nil {
			error(be.Token, "'%.*s' only allowed in non-inline loops", name.Len, name.Data)
		}
	}
	if label != nil {
		if c.InDefer {
			error(label, "A labelled '%.*s' cannot be used within a 'defer'", name.Len, name.Data)
			return ExprKind_Expr
		}
		if label.Kind != Ast_Ident {
			error(label, "A branch statement's label name must be an identifier")
			return ExprKind_Expr
		}
		ident := label
		label_name := ident.Ident.Token.String
		var ident_op Operand
		e := check_ident(c, &ident_op, ident, nil, nil, false)
		if e == nil {
			error(ident, "Undeclared label name: %.*s", label_name.Len, label_name.Data)
			return ExprKind_Expr
		}
		add_entity_use(c, ident, e)
		if e.Kind != Entity_Label {
			error(ident, "'%.*s' is not a label", label_name.Len, label_name.Data)
			return ExprKind_Expr
		}
		parent := e.Label.Parent
		gb_assert_handler("Assertion Failure", "parent != nil", "cmd_check_expr_main.go", 0)
		switch parent.Kind {
		case Ast_BlockStmt, Ast_IfStmt, Ast_SwitchStmt:
			if be.Token.Kind != Token_or_break {
				error(label, "Label '%.*s' can only be used with 'or_break'", e.Token.String.Len, e.Token.String.Data)
			}
		case Ast_RangeStmt, Ast_ForStmt:
			if (be.Token.Kind != Token_or_break) && (be.Token.Kind != Token_or_continue) {
				error(label, "Label '%.*s' can only be used with 'or_break' and 'or_continue'", e.Token.String.Len, e.Token.String.Data)
			}
		}
	}
	return ExprKind_Expr
}

func check_expr_base_internal(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	prev_state_flags := c.StateFlags
	defer func() { c.StateFlags = prev_state_flags }()
	if node.StateFlags != 0 {
		in_flags := node.StateFlags
		out_flags := c.StateFlags
		if in_flags&uint8(StateFlag_NoBoundsCheck) != 0 {
			out_flags |= uint8(StateFlag_NoBoundsCheck)
			out_flags &^= uint8(StateFlag_BoundsCheck)
		} else if in_flags&uint8(StateFlag_BoundsCheck) != 0 {
			out_flags |= uint8(StateFlag_BoundsCheck)
			out_flags &^= uint8(StateFlag_NoBoundsCheck)
		}
		if in_flags&uint8(StateFlag_NoTypeAssert) != 0 {
			out_flags |= uint8(StateFlag_NoTypeAssert)
			out_flags &^= uint8(StateFlag_TypeAssert)
		} else if in_flags&uint8(StateFlag_TypeAssert) != 0 {
			out_flags |= uint8(StateFlag_TypeAssert)
			out_flags &^= uint8(StateFlag_NoTypeAssert)
		}
		c.StateFlags = out_flags
	}
	kind := ExprKind_Stmt
	o.Mode = Addressing_Invalid
	o.Type = t_invalid
	o.Value = ExactValue{Kind: ExactValue_Invalid}
	switch node.Kind {
	default:
		return kind
	case Ast_BadExpr:
		return kind
	case Ast_HelperType:
		ht := &node.HelperType
		type_ := check_type(c, ht.Type)
		if type_ != nil && type_ != t_invalid {
			o.Mode = Addressing_Type
			o.Type = type_
		}
		return kind
	case Ast_Implicit:
		i := &node.Implicit
		switch i.Kind {
		case Token_context:
			if c.ProcName.Len == 0 && c.CurrProcSig == nil {
				error(node, "'context' is only allowed within procedures")
				return kind
			}
			if unparen_expr(c.AssignmentLHSHint) == node {
				c.Scope.Flags |= ScopeFlag_ContextDefined
			}
			if (c.Scope.Flags & ScopeFlag_ContextDefined) == 0 {
				error(node, "'context' has not been defined within this scope")
			}
			init_core_context(c.Checker)
			o.Mode = Addressing_Context
			o.Type = t_context
		default:
			error(node, "Illegal implicit name '%.*s'", i.String.Len, i.String.Data)
			return kind
		}
	case Ast_Ident:
		check_ident(c, o, node, nil, type_hint, false)
	case Ast_Uninit:
		o.Mode = Addressing_Value
		o.Type = t_untyped_uninit
		error(node, "Global variables will always be zeroed if left unassigned, --- is disallowed")
	case Ast_BasicLit:
		bl := &node.BasicLit
		t := t_invalid
		switch node.TAV.Value.Kind {
		case ExactValue_String:
			t = t_untyped_string
		case ExactValue_String16:
			t = t_string16
		case ExactValue_Float:
			t = t_untyped_float
		case ExactValue_Complex:
			t = t_untyped_complex
		case ExactValue_Quaternion:
			t = t_untyped_quaternion
		case ExactValue_Integer:
			t = t_untyped_integer
			if bl.Token.Kind == Token_Rune {
				t = t_untyped_rune
			}
		default:
			gb_assert_handler("Panic", "Unhandled value type for basic literal", "cmd_check_expr_main.go", 0)
		}
		o.Mode = Addressing_Constant
		o.Type = t
		o.Value = node.TAV.Value
	case Ast_BasicDirective:
		kind = check_basic_directive_expr(c, o, node, type_hint)
	case Ast_ProcGroup:
		error(node, "Illegal use of a procedure group")
		o.Mode = Addressing_Invalid
	case Ast_ProcLit:
		pl := &node.ProcLit
		ctx := *c
		var decl *DeclInfo
		type_ := alloc_type(Type_Proc)
		check_open_scope(&ctx, pl.Type)
		{
			decl = make_decl_info(ctx.Scope, ctx.Decl)
			decl.ProcLit = node
			ctx.Decl = decl
			defer func() { ctx.Decl = ctx.Decl.Parent }()
			if pl.Tags != 0 {
				error(node, "A procedure literal cannot have tags")
				pl.Tags = 0
			}
			check_procedure_type(&ctx, type_, pl.Type)
			if !is_type_proc(type_) {
				str := expr_to_string(node)
				defer gb_string_free(str)
				error(node, "Invalid procedure literal '%s'", goStr(str))
				check_close_scope(&ctx)
				return kind
			}
			if pl.Body == nil {
				error(node, "A procedure literal must have a body")
				return kind
			}
			pl.Decl = decl
			check_procedure_later_full(ctx.Checker, ctx.File, empty_token, decl, type_, pl.Body, pl.Tags)
			mutex_lock(&ctx.Checker.NestedProcLitsMutex)
			ctx.Checker.NestedProcLits = append(ctx.Checker.NestedProcLits, decl)
			mutex_unlock(&ctx.Checker.NestedProcLitsMutex)
		}
		check_close_scope(&ctx)
		o.Mode = Addressing_Value
		o.Type = type_
		o.Value = exact_value_procedure(node)
	case Ast_TernaryIfExpr:
		kind = check_ternary_if_expr(c, o, node, type_hint)
	case Ast_TernaryWhenExpr:
		kind = check_ternary_when_expr(c, o, node, type_hint)
	case Ast_OrElseExpr:
		return check_or_else_expr(c, o, node, type_hint)
	case Ast_OrReturnExpr:
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | uint32(ViralStateFlag_ContainsOrReturn))
		return check_or_return_expr(c, o, node, type_hint)
	case Ast_OrBranchExpr:
		return check_or_branch_expr(c, o, node, type_hint)
	case Ast_CompoundLit:
		kind = check_compound_literal(c, o, node, type_hint)
	case Ast_ParenExpr:
		pe := &node.ParenExpr
		kind = check_expr_base(c, o, pe.Expr, type_hint)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | pe.Expr.ViralStateFlags.Load())
		o.Expr = node
	case Ast_TagExpr:
		te := &node.TagExpr
		name := te.Name.String
		error(node, "Unknown tag expression, #%.*s", name.Len, name.Data)
		if te.Expr != nil {
			kind = check_expr_base(c, o, te.Expr, type_hint)
			node.ViralStateFlags.Store(node.ViralStateFlags.Load() | te.Expr.ViralStateFlags.Load())
		}
		o.Expr = node
	case Ast_TypeAssertion:
		kind = check_type_assertion(c, o, node, type_hint)
	case Ast_TypeCast:
		tc := &node.TypeCast
		check_expr_or_type(c, o, tc.Type)
		if o.Mode != Addressing_Type {
			str := expr_to_string(tc.Type)
			defer gb_string_free(str)
			error(tc.Type, "Expected a type, got %s", goStr(str))
			o.Mode = Addressing_Invalid
		}
		if o.Mode == Addressing_Invalid {
			o.Expr = node
			return kind
		}
		type_ := o.Type
		check_expr_base(c, o, tc.Expr, type_)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | tc.Expr.ViralStateFlags.Load())
		if o.Mode != Addressing_Invalid {
			switch tc.Token.Kind {
			case Token_transmute:
				check_transmute(c, node, o, type_, true)
			case Token_cast:
				check_cast(c, o, type_, true)
			default:
				error(node, "Invalid AST: Invalid casting expression")
				o.Mode = Addressing_Invalid
			}
		}
		return ExprKind_Expr
	case Ast_AutoCast:
		ac := &node.AutoCast
		check_expr_base(c, o, ac.Expr, type_hint)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | ac.Expr.ViralStateFlags.Load())
		if o.Mode == Addressing_Invalid {
			o.Expr = node
			return kind
		}
		if type_hint != nil {
			check_cast(c, o, type_hint)
		}
		o.Expr = node
		return ExprKind_Expr
	case Ast_UnaryExpr:
		ue := &node.UnaryExpr
		th := type_hint
		if ue.Op.Kind == Token_And {
			th = type_deref(th)
		}
		check_expr_base(c, o, ue.Expr, th)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | ue.Expr.ViralStateFlags.Load())
		if o.Mode != Addressing_Invalid {
			check_unary_expr(c, o, ue.Op, node)
		} else {
			begin_error_block()
			defer end_error_block()
			s := expr_to_string(ue.Expr)
			defer gb_string_free(s)
			error(node, "Cannot address value '%s' as it has not got a determined type yet", goStr(s))
			e := entity_of_node(ue.Expr)
			if e != nil && e.Kind == Entity_Variable {
				error_line("\tSuggestion: Add an explicit type to the declaration of '%.*s' rather than relying on type inference", e.Token.String.Len, e.Token.String.Data)
			}
		}
		o.Expr = node
		return ExprKind_Expr
	case Ast_BinaryExpr:
		be := &node.BinaryExpr
		check_binary_expr(c, o, node, type_hint, true)
		if o.Mode == Addressing_Invalid {
			o.Expr = node
			return kind
		}
	case Ast_SelectorExpr:
		se := &node.SelectorExpr
		check_selector(c, o, node, type_hint)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | se.Expr.ViralStateFlags.Load())
	case Ast_SelectorCallExpr:
		return check_selector_call_expr(c, o, node, type_hint)
	case Ast_ImplicitSelectorExpr:
		return check_implicit_selector_expr(c, o, node, type_hint)
	case Ast_IndexExpr:
		kind = check_index_expr(c, o, node, type_hint)
	case Ast_SliceExpr:
		kind = check_slice_expr(c, o, node, type_hint)
	case Ast_MatrixIndexExpr:
		mie := &node.MatrixIndexExpr
		check_matrix_index_expr(c, o, node, type_hint)
		o.Expr = node
		return ExprKind_Expr
	case Ast_CallExpr:
		ce := &node.CallExpr
		return check_call_expr(c, o, node, ce.Proc, ce.Args, ce.Inlining, ce.Tailing, type_hint)
	case Ast_DerefExpr:
		de := &node.DerefExpr
		check_expr_or_type(c, o, de.Expr)
		node.ViralStateFlags.Store(node.ViralStateFlags.Load() | de.Expr.ViralStateFlags.Load())
		if o.Mode == Addressing_Invalid {
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		} else if o.Mode == Addressing_Type {
			str := expr_to_string(o.Expr)
			defer gb_string_free(str)
			error(o.Expr, "Cannot dereference '%s' because it is a type", goStr(str))
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		} else {
			t := base_type(o.Type)
			if t.Kind == Type_Pointer && !is_type_empty_union(t.Pointer.Elem) {
				o.Mode = Addressing_Variable
				o.Type = t.Pointer.Elem
			} else if t.Kind == Type_SoaPointer {
				o.Mode = Addressing_SoaVariable
				o.Type = type_deref(t)
			} else {
				str := expr_to_string(o.Expr)
				defer gb_string_free(str)
				typ := type_to_string(o.Type)
				defer gb_string_free(typ)
				begin_error_block()
				defer end_error_block()
				error(o.Expr, "Cannot dereference '%s' of type '%s'", goStr(str), goStr(typ))
				if o.Type != nil && is_type_multi_pointer(o.Type) {
					if !buildContext.TerseErrors {
						error_line("\tDid you mean '%s[0]'?\n", goStr(str))
					}
				}
				o.Mode = Addressing_Invalid
				o.Expr = node
				return kind
			}
		}
	case Ast_InlineAsmExpr:
		ia := &node.InlineAsmExpr
		if c.CurrProcDecl == nil {
			error(node, "Inline asm expressions are only allowed within a procedure body")
		}
		param_types := make([]*Type, len(ia.ParamTypes))
		var return_type *Type
		for i := 0; i < len(ia.ParamTypes); i++ {
			param_types[i] = check_type(c, ia.ParamTypes[i])
		}
		if ia.ReturnType != nil {
			return_type = check_type(c, ia.ReturnType)
		}
		var x Operand
		check_expr(c, &x, ia.AsmString)
		if x.Mode != Addressing_Constant || !is_type_string(x.Type) {
			error(x.Expr, "Expected a constant string for the inline asm main parameter")
		}
		check_expr(c, &x, ia.ConstraintsString)
		if x.Mode != Addressing_Constant || !is_type_string(x.Type) {
			error(x.Expr, "Expected a constant string for the inline asm constraints parameter")
		}
		scope := create_scope(c.Info, c.Scope)
		scope.Flags |= ScopeFlag_Proc
		params := alloc_type_tuple()
		results := alloc_type_tuple()
		if len(param_types) != 0 {
			params.Tuple.Variables = make([]*Entity, len(param_types))
			for i := 0; i < len(param_types); i++ {
				params.Tuple.Variables[i] = alloc_entity_param(scope, blank_token, param_types[i], false, true)
			}
		}
		if return_type != nil {
			results.Tuple.Variables = make([]*Entity, 1)
			results.Tuple.Variables[0] = alloc_entity_param(scope, blank_token, return_type, false, true)
		}
		pt := alloc_type_proc(scope, params, isize(len(param_types)), results, mapValue(return_type != nil, 1, 0), false, ProcCC_InlineAsm)
		o.Type = pt
		o.Mode = Addressing_Value
		o.Expr = node
		return ExprKind_Expr
	case Ast_DistinctType, Ast_TypeidType, Ast_PolyType, Ast_ProcType,
		Ast_PointerType, Ast_MultiPointerType, Ast_ArrayType, Ast_DynamicArrayType,
		Ast_FixedCapacityDynamicArrayType, Ast_StructType, Ast_UnionType, Ast_EnumType,
		Ast_MapType, Ast_BitSetType, Ast_MatrixType, Ast_RelativeType:
		o.Mode = Addressing_Type
		o.Type = check_type(c, node)
	}
	kind = ExprKind_Expr
	o.Expr = node
	return kind
}

func check_expr_base(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := check_expr_base_internal(c, o, node, type_hint)
	if o.Type != nil && core_type(o.Type) == nil {
		o.Type = t_invalid
		xs := expr_to_string(o.Expr)
		defer gb_string_free(xs)
		if o.Mode == Addressing_Type {
			error(o.Expr, "Invalid type usage '%s'", goStr(xs))
		} else {
			error(o.Expr, "Invalid expression '%s'", goStr(xs))
		}
	}
	if o.Type != nil && is_type_untyped(o.Type) {
		add_untyped(c, node, o.Mode, o.Type, o.Value)
	}
	check_rtti_type_disallowed_expr(node, o.Type, "An expression is using a type, %s, which has been disallowed")
	add_type_and_value(c, node, o.Mode, o.Type, o.Value)
	return kind
}

func check_multi_expr_or_type(c *CheckerContext, o *Operand, e *Ast) {
	check_expr_base(c, o, e, nil)
	switch o.Mode {
	default:
		return
	case Addressing_NoValue:
		error_operand_no_value(o)
	}
	o.Mode = Addressing_Invalid
}

func check_multi_expr(c *CheckerContext, o *Operand, e *Ast) {
	check_expr_base(c, o, e, nil)
	switch o.Mode {
	default:
		return
	case Addressing_NoValue:
		error_operand_no_value(o)
	case Addressing_Type:
		error_operand_not_expression(o)
	}
	o.Mode = Addressing_Invalid
}

func check_multi_expr_with_type_hint(c *CheckerContext, o *Operand, e *Ast, type_hint *Type) {
	check_expr_base(c, o, e, type_hint)
	switch o.Mode {
	default:
		return
	case Addressing_NoValue:
		error_operand_no_value(o)
	case Addressing_Type:
		if type_hint != nil && is_type_typeid(type_hint) {
			add_type_info_type(c, o.Type)
			break
		}
		error_operand_not_expression(o)
	}
	o.Mode = Addressing_Invalid
}

func check_not_tuple(c *CheckerContext, o *Operand) {
	if o.Mode == Addressing_Value {
		if o.Type.Kind == Type_Tuple {
			count := len(o.Type.Tuple.Variables)
			error(o.Expr, "%td-valued expression found where single value expected", count)
			o.Mode = Addressing_Invalid
			gb_assert_handler("Assertion Failure", "count != 1", "cmd_check_expr_main.go", 0)
		}
	}
}

func check_expr(c *CheckerContext, o *Operand, e *Ast) {
	check_multi_expr(c, o, e)
	check_not_tuple(c, o)
}

func check_expr_or_type(c *CheckerContext, o *Operand, e *Ast, type_hint *Type) {
	check_expr_base(c, o, e, type_hint)
	check_not_tuple(c, o)
	error_operand_no_value(o)
}

func check_expr_with_type_hint(c *CheckerContext, o *Operand, e *Ast, t *Type) {
	check_expr_base(c, o, e, t)
	check_not_tuple(c, o)
	err_str := ""
	switch o.Mode {
	case Addressing_NoValue:
		err_str = "used as a value"
	case Addressing_Type:
		if t == nil || !is_type_typeid(t) {
			err_str = "is not an expression but a type, in this context it is ambiguous"
		}
	case Addressing_Builtin:
		err_str = "must be called"
	}
	if err_str != "" {
		str := expr_to_string(e)
		defer gb_string_free(str)
		error(e, "'%s' %s", goStr(str), err_str)
		o.Mode = Addressing_Invalid
	}
}

// ---------------------------------------------------------------------------
// write_expr_to_string helpers
// ---------------------------------------------------------------------------

func write_struct_fields_to_string(str gbString, params []*Ast) gbString {
	for i := 0; i < len(params); i++ {
		if i > 0 {
			str = gb_string_appendc(str, ", ")
		}
		str = write_expr_to_string(str, params[i], false)
	}
	return str
}

func string_append_string(str gbString, s String) gbString {
	if s.Len > 0 {
		return gb_string_append_length(str, s.Data, s.Len)
	}
	return str
}

func string_append_token(str gbString, tok Token) gbString {
	return string_append_string(str, tok.String)
}

// write_expr_to_string converts an AST node to its string representation
func write_expr_to_string(str gbString, node *Ast, shorthand bool) gbString {
	if node == nil {
		return str
	}
	if is_ast_stmt(node) {
		gb_assert_handler("Assertion Failure", "stmt passed to write_expr_to_string", "cmd_check_expr_main.go", 0)
	}
	switch node.Kind {
	default:
		str = gb_string_appendc(str, "(BadExpr)")
	case Ast_Ident:
		i := &node.Ident
		str = string_append_token(str, i.Token)
	case Ast_Implicit:
		i := &node.Implicit
		str = string_append_token(str, *i)
	case Ast_BasicLit:
		bl := &node.BasicLit
		str = string_append_token(str, bl.Token)
	case Ast_BasicDirective:
		bd := &node.BasicDirective
		str = gb_string_append_rune(str, '#')
		str = string_append_string(str, bd.Name.String)
	case Ast_Uninit:
		str = gb_string_appendc(str, "---")
	case Ast_ProcGroup:
		pg := &node.ProcGroup
		str = gb_string_appendc(str, "proc{")
		for i := 0; i < len(pg.Args); i++ {
			if i > 0 {
				str = gb_string_appendc(str, ", ")
			}
			str = write_expr_to_string(str, pg.Args[i], shorthand)
		}
		str = gb_string_append_rune(str, '}')
	case Ast_ProcLit:
		pl := &node.ProcLit
		str = write_expr_to_string(str, pl.Type, shorthand)
		if pl.Body != nil {
			str = gb_string_appendc(str, " {...}")
		} else {
			str = gb_string_appendc(str, " ---")
		}
	case Ast_CompoundLit:
		cl := &node.CompoundLit
		str = write_expr_to_string(str, cl.Type, shorthand)
		str = gb_string_append_rune(str, '{')
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			for i := 0; i < len(cl.Elems); i++ {
				if i > 0 {
					str = gb_string_appendc(str, ", ")
				}
				str = write_expr_to_string(str, cl.Elems[i], shorthand)
			}
		}
		str = gb_string_append_rune(str, '}')
	case Ast_TagExpr:
		te := &node.TagExpr
		str = gb_string_append_rune(str, '#')
		str = string_append_token(str, te.Name)
		str = write_expr_to_string(str, te.Expr, shorthand)
	case Ast_UnaryExpr:
		ue := &node.UnaryExpr
		str = string_append_token(str, ue.Op)
		str = write_expr_to_string(str, ue.Expr, shorthand)
	case Ast_DerefExpr:
		de := &node.DerefExpr
		str = write_expr_to_string(str, de.Expr, shorthand)
		str = gb_string_append_rune(str, '^')
	case Ast_BinaryExpr:
		be := &node.BinaryExpr
		str = write_expr_to_string(str, be.Left, shorthand)
		str = gb_string_append_rune(str, ' ')
		str = string_append_token(str, be.Op)
		str = gb_string_append_rune(str, ' ')
		str = write_expr_to_string(str, be.Right, shorthand)
	case Ast_TernaryIfExpr:
		te := &node.TernaryIfExpr
		x_pos := ast_token(te.X).Pos
		cond_pos := ast_token(te.Cond).Pos
		if x_pos.Offset < cond_pos.Offset {
			str = write_expr_to_string(str, te.X, shorthand)
			str = gb_string_appendc(str, " if ")
			str = write_expr_to_string(str, te.Cond, shorthand)
			str = gb_string_appendc(str, " else ")
			str = write_expr_to_string(str, te.Y, shorthand)
		} else {
			str = write_expr_to_string(str, te.Cond, shorthand)
			str = gb_string_appendc(str, " ? ")
			str = write_expr_to_string(str, te.X, shorthand)
			str = gb_string_appendc(str, " : ")
			str = write_expr_to_string(str, te.Y, shorthand)
		}
	case Ast_TernaryWhenExpr:
		te := &node.TernaryWhenExpr
		str = write_expr_to_string(str, te.X, shorthand)
		str = gb_string_appendc(str, " when ")
		str = write_expr_to_string(str, te.Cond, shorthand)
		str = gb_string_appendc(str, " else ")
		str = write_expr_to_string(str, te.Y, shorthand)
	case Ast_OrElseExpr:
		oe := &node.OrElseExpr
		str = write_expr_to_string(str, oe.X, shorthand)
		str = gb_string_appendc(str, " or_else ")
		str = write_expr_to_string(str, oe.Y, shorthand)
	case Ast_OrReturnExpr:
		oe := &node.OrReturnExpr
		str = write_expr_to_string(str, oe.Expr, shorthand)
		str = gb_string_appendc(str, " or_return")
	case Ast_OrBranchExpr:
		oe := &node.OrBranchExpr
		str = write_expr_to_string(str, oe.Expr, shorthand)
		str = gb_string_append_rune(str, ' ')
		str = string_append_token(str, oe.Token)
		if oe.Label != nil {
			str = gb_string_append_rune(str, ' ')
			str = write_expr_to_string(str, oe.Label, shorthand)
		}
	case Ast_ParenExpr:
		pe := &node.ParenExpr
		str = gb_string_append_rune(str, '(')
		str = write_expr_to_string(str, pe.Expr, shorthand)
		str = gb_string_append_rune(str, ')')
	case Ast_SelectorExpr:
		se := &node.SelectorExpr
		str = write_expr_to_string(str, se.Expr, shorthand)
		str = string_append_token(str, se.Token)
		str = write_expr_to_string(str, se.Selector, shorthand)
	case Ast_ImplicitSelectorExpr:
		se := &node.ImplicitSelectorExpr
		str = gb_string_append_rune(str, '.')
		str = write_expr_to_string(str, se.Selector, shorthand)
	case Ast_SelectorCallExpr:
		se := &node.SelectorCallExpr
		str = write_expr_to_string(str, se.Expr, shorthand)
		str = gb_string_appendc(str, "(")
		ce := &se.Call.CallExpr
		start := 0
		if se.ModifiedCall {
			start = 1
		}
		for i := start; i < len(ce.Args); i++ {
			arg := ce.Args[i]
			if i > start {
				str = gb_string_appendc(str, ", ")
			}
			str = write_expr_to_string(str, arg, shorthand)
		}
		str = gb_string_appendc(str, ")")
	case Ast_TypeAssertion:
		ta := &node.TypeAssertion
		str = write_expr_to_string(str, ta.Expr, shorthand)
		if ta.Type != nil &&
			ta.Type.Kind == Ast_UnaryExpr &&
			ta.Type.UnaryExpr.Op.Kind == Token_Question {
			str = gb_string_appendc(str, ".?")
		} else {
			str = gb_string_appendc(str, ".(")
			str = write_expr_to_string(str, ta.Type, shorthand)
			str = gb_string_append_rune(str, ')')
		}
	case Ast_TypeCast:
		tc := &node.TypeCast
		str = string_append_token(str, tc.Token)
		str = gb_string_append_rune(str, '(')
		str = write_expr_to_string(str, tc.Type, shorthand)
		str = gb_string_append_rune(str, ')')
		str = write_expr_to_string(str, tc.Expr, shorthand)
	case Ast_AutoCast:
		ac := &node.AutoCast
		str = string_append_token(str, ac.Token)
		str = gb_string_append_rune(str, ' ')
		str = write_expr_to_string(str, ac.Expr, shorthand)
	case Ast_IndexExpr:
		ie := &node.IndexExpr
		str = write_expr_to_string(str, ie.Expr, shorthand)
		str = gb_string_append_rune(str, '[')
		str = write_expr_to_string(str, ie.Index, shorthand)
		str = gb_string_append_rune(str, ']')
	case Ast_SliceExpr:
		se := &node.SliceExpr
		str = write_expr_to_string(str, se.Expr, shorthand)
		str = gb_string_append_rune(str, '[')
		str = write_expr_to_string(str, se.Low, shorthand)
		str = string_append_token(str, se.Interval)
		str = write_expr_to_string(str, se.High, shorthand)
		str = gb_string_append_rune(str, ']')
	case Ast_MatrixIndexExpr:
		mie := &node.MatrixIndexExpr
		str = write_expr_to_string(str, mie.Expr, shorthand)
		str = gb_string_append_rune(str, '[')
		str = write_expr_to_string(str, mie.RowIndex, shorthand)
		str = gb_string_appendc(str, ", ")
		str = write_expr_to_string(str, mie.ColumnIndex, shorthand)
		str = gb_string_append_rune(str, ']')
	case Ast_Ellipsis:
		e := &node.Ellipsis
		str = gb_string_appendc(str, "..")
		str = write_expr_to_string(str, e.Expr, shorthand)
	case Ast_FieldValue:
		fv := &node.FieldValue
		str = write_expr_to_string(str, fv.Field, shorthand)
		str = gb_string_appendc(str, " = ")
		str = write_expr_to_string(str, fv.Value, shorthand)
	case Ast_EnumFieldValue:
		fv := &node.EnumFieldValue
		str = write_expr_to_string(str, fv.Name, shorthand)
		if fv.Value != nil {
			str = gb_string_appendc(str, " = ")
			str = write_expr_to_string(str, fv.Value, shorthand)
		}
	case Ast_HelperType:
		ht := &node.HelperType
		str = gb_string_appendc(str, "#type ")
		str = write_expr_to_string(str, ht.Type, shorthand)
	case Ast_DistinctType:
		ht := &node.DistinctType
		str = gb_string_appendc(str, "distinct ")
		str = write_expr_to_string(str, ht.Type, shorthand)
	case Ast_PolyType:
		pt := &node.PolyType
		str = gb_string_append_rune(str, '$')
		str = write_expr_to_string(str, pt.Type, shorthand)
		if pt.Specialization != nil {
			str = gb_string_append_rune(str, '/')
			str = write_expr_to_string(str, pt.Specialization, shorthand)
		}
	case Ast_PointerType:
		pt := &node.PointerType
		if pt.Tag != nil {
			str = write_expr_to_string(str, pt.Tag, false)
		}
		str = gb_string_append_rune(str, '^')
		str = write_expr_to_string(str, pt.Type, shorthand)
	case Ast_MultiPointerType:
		pt := &node.MultiPointerType
		str = gb_string_appendc(str, "[^]")
		str = write_expr_to_string(str, pt.Type, shorthand)
	case Ast_ArrayType:
		at := &node.ArrayType
		if at.Tag != nil {
			str = write_expr_to_string(str, at.Tag, false)
		}
		str = gb_string_append_rune(str, '[')
		if at.Count != nil &&
			at.Count.Kind == Ast_UnaryExpr &&
			at.Count.UnaryExpr.Op.Kind == Token_Question {
			str = gb_string_appendc(str, "?")
		} else {
			str = write_expr_to_string(str, at.Count, shorthand)
		}
		str = gb_string_append_rune(str, ']')
		str = write_expr_to_string(str, at.Elem, shorthand)
	case Ast_DynamicArrayType:
		at := &node.DynamicArrayType
		if at.Tag != nil {
			str = write_expr_to_string(str, at.Tag, false)
		}
		str = gb_string_appendc(str, "[dynamic]")
		str = write_expr_to_string(str, at.Elem, shorthand)
	case Ast_BitSetType:
		bs := &node.BitSetType
		str = gb_string_appendc(str, "bit_set[")
		str = write_expr_to_string(str, bs.Elem, shorthand)
		str = gb_string_appendc(str, "]")
	case Ast_MapType:
		mt := &node.MapType
		str = gb_string_appendc(str, "map[")
		str = write_expr_to_string(str, mt.Key, shorthand)
		str = gb_string_append_rune(str, ']')
		str = write_expr_to_string(str, mt.Value, shorthand)
	case Ast_MatrixType:
		mt := &node.MatrixType
		str = gb_string_appendc(str, "matrix[")
		str = write_expr_to_string(str, mt.RowCount, shorthand)
		str = gb_string_appendc(str, ", ")
		str = write_expr_to_string(str, mt.ColumnCount, shorthand)
		str = gb_string_append_rune(str, ']')
		str = write_expr_to_string(str, mt.Elem, shorthand)
	case Ast_Field:
		f := &node.Field
		if f.Flags&uint32(FieldFlagUsing) != 0 {
			str = gb_string_appendc(str, "using ")
		}
		if f.Flags&uint32(FieldFlagNoAlias) != 0 {
			str = gb_string_appendc(str, "#no_alias ")
		}
		if f.Flags&uint32(FieldFlagCVararg) != 0 {
			str = gb_string_appendc(str, "#c_vararg ")
		}
		if f.Flags&uint32(FieldFlagAnyInt) != 0 {
			str = gb_string_appendc(str, "#any_int ")
		}
		if f.Flags&uint32(FieldFlagNoBroadcast) != 0 {
			str = gb_string_appendc(str, "#no_broadcast ")
		}
		if f.Flags&uint32(FieldFlagConst) != 0 {
			str = gb_string_appendc(str, "#const ")
		}
		if f.Flags&uint32(FieldFlagSubtype) != 0 {
			str = gb_string_appendc(str, "#subtype ")
		}
		for i := 0; i < len(f.Names); i++ {
			name := f.Names[i]
			if i > 0 {
				str = gb_string_appendc(str, ", ")
			}
			str = write_expr_to_string(str, name, shorthand)
		}
		if len(f.Names) > 0 {
			if f.Type == nil && f.DefaultValue != nil {
				str = gb_string_append_rune(str, ' ')
			}
			str = gb_string_appendc(str, ":")
		}
		if f.Type != nil {
			str = gb_string_append_rune(str, ' ')
			str = write_expr_to_string(str, f.Type, shorthand)
		}
		if f.DefaultValue != nil {
			if f.Type != nil {
				str = gb_string_append_rune(str, ' ')
			}
			str = gb_string_appendc(str, "= ")
			str = write_expr_to_string(str, f.DefaultValue, shorthand)
		}
	case Ast_FieldList:
		f := &node.FieldList
		has_name := false
		for i := 0; i < len(f.List); i++ {
			field := &f.List[i].Field
			if len(field.Names) > 1 {
				has_name = true
				break
			}
			if len(field.Names) == 0 {
				continue
			}
			if !is_blank_ident(field.Names[0]) {
				has_name = true
				break
			}
		}
		for i := 0; i < len(f.List); i++ {
			if i > 0 {
				str = gb_string_appendc(str, ", ")
			}
			if has_name {
				str = write_expr_to_string(str, f.List[i], shorthand)
			} else {
				field := &f.List[i].Field
				if field.Flags&uint32(FieldFlagUsing) != 0 {
					str = gb_string_appendc(str, "using ")
				}
				if field.Flags&uint32(FieldFlagNoAlias) != 0 {
					str = gb_string_appendc(str, "#no_alias ")
				}
				if field.Flags&uint32(FieldFlagCVararg) != 0 {
					str = gb_string_appendc(str, "#c_vararg ")
				}
				if field.Flags&uint32(FieldFlagAnyInt) != 0 {
					str = gb_string_appendc(str, "#any_int ")
				}
				if field.Flags&uint32(FieldFlagNoBroadcast) != 0 {
					str = gb_string_appendc(str, "#no_broadcast ")
				}
				if field.Flags&uint32(FieldFlagConst) != 0 {
					str = gb_string_appendc(str, "#const ")
				}
				if field.Flags&uint32(FieldFlagSubtype) != 0 {
					str = gb_string_appendc(str, "#subtype ")
				}
				str = write_expr_to_string(str, field.Type, shorthand)
			}
		}
	case Ast_CallExpr:
		ce := &node.CallExpr
		switch ce.Tailing {
		case ProcTailingMustTail:
			str = gb_string_appendc(str, "#must_tail ")
		}
		switch ce.Inlining {
		case ProcInliningInline:
			str = gb_string_appendc(str, "#force_inline ")
		case ProcInliningNoInline:
			str = gb_string_appendc(str, "#force_no_inline ")
		}
		str = write_expr_to_string(str, ce.Proc, shorthand)
		str = gb_string_appendc(str, "(")
		idx0 := 0
		if ce.WasSelector {
			idx0 = 1
		}
		for i := idx0; i < len(ce.Args); i++ {
			arg := ce.Args[i]
			if i > idx0 {
				str = gb_string_appendc(str, ", ")
			}
			str = write_expr_to_string(str, arg, shorthand)
		}
		str = gb_string_appendc(str, ")")
	case Ast_TypeidType:
		tt := &node.TypeidType
		str = gb_string_appendc(str, "typeid")
		if tt.Specialization != nil {
			str = gb_string_appendc(str, "/")
			str = write_expr_to_string(str, tt.Specialization, shorthand)
		}
	case Ast_ProcType:
		pt := &node.ProcType
		str = gb_string_appendc(str, "proc(")
		str = write_expr_to_string(str, pt.Params, shorthand)
		str = gb_string_appendc(str, ")")
		if pt.Results != nil {
			str = gb_string_appendc(str, " -> ")
			parens_needed := false
			if pt.Results != nil && pt.Results.Kind == Ast_FieldList {
				for _, field_ast := range pt.Results.FieldList.List {
					f := &field_ast.Field
					if len(f.Names) != 0 {
						parens_needed = true
						break
					}
				}
			}
			if parens_needed {
				str = gb_string_append_rune(str, '(')
			}
			str = write_expr_to_string(str, pt.Results, shorthand)
			if parens_needed {
				str = gb_string_append_rune(str, ')')
			}
		}
	case Ast_StructType:
		st := &node.StructType
		str = gb_string_appendc(str, "struct ")
		if st.PolymorphicParams != nil {
			str = gb_string_append_rune(str, '(')
			str = write_expr_to_string(str, st.PolymorphicParams, shorthand)
			str = gb_string_appendc(str, ") ")
		}
		if st.IsPacked {
			str = gb_string_appendc(str, "#packed ")
		}
		if st.IsRawUnion {
			str = gb_string_appendc(str, "#raw_union ")
		}
		if st.IsAllOrNone {
			str = gb_string_appendc(str, "#all_or_none ")
		}
		if st.IsSimple {
			str = gb_string_appendc(str, "#simple ")
		}
		if st.Align != nil {
			str = gb_string_appendc(str, "#align ")
			str = write_expr_to_string(str, st.Align, shorthand)
			str = gb_string_append_rune(str, ' ')
		}
		str = gb_string_append_rune(str, '{')
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			str = write_struct_fields_to_string(str, st.Fields)
		}
		str = gb_string_append_rune(str, '}')
	case Ast_UnionType:
		st := &node.UnionType
		str = gb_string_appendc(str, "union ")
		if st.PolymorphicParams != nil {
			str = gb_string_append_rune(str, '(')
			str = write_expr_to_string(str, st.PolymorphicParams, shorthand)
			str = gb_string_appendc(str, ") ")
		}
		switch st.Kind {
		case UnionTypeNoNil:
			str = gb_string_appendc(str, "#no_nil ")
		case UnionTypeSharedNil:
			str = gb_string_appendc(str, "#shared_nil ")
		}
		if st.Align != nil {
			str = gb_string_appendc(str, "#align ")
			str = write_expr_to_string(str, st.Align, shorthand)
			str = gb_string_append_rune(str, ' ')
		}
		str = gb_string_append_rune(str, '{')
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			str = write_struct_fields_to_string(str, st.Variants)
		}
		str = gb_string_append_rune(str, '}')
	case Ast_EnumType:
		et := &node.EnumType
		str = gb_string_appendc(str, "enum ")
		if et.BaseType != nil {
			str = write_expr_to_string(str, et.BaseType, shorthand)
			str = gb_string_append_rune(str, ' ')
		}
		str = gb_string_append_rune(str, '{')
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			for i := 0; i < len(et.Fields); i++ {
				if i > 0 {
					str = gb_string_appendc(str, ", ")
				}
				str = write_expr_to_string(str, et.Fields[i], shorthand)
			}
		}
		str = gb_string_append_rune(str, '}')
	case Ast_RelativeType:
		rt := &node.RelativeType
		str = write_expr_to_string(str, rt.Tag, shorthand)
		str = gb_string_appendc(str, "")
		str = write_expr_to_string(str, rt.Type, shorthand)
	case Ast_BitFieldField:
		f := &node.BitFieldField
		str = write_expr_to_string(str, f.Name, shorthand)
		str = gb_string_appendc(str, ": ")
		str = write_expr_to_string(str, f.Type, shorthand)
		str = gb_string_appendc(str, " | ")
		str = write_expr_to_string(str, f.BitSize, shorthand)
	case Ast_BitFieldType:
		bf := &node.BitFieldType
		str = gb_string_appendc(str, "bit_field ")
		if !shorthand {
			str = write_expr_to_string(str, bf.BackingType, shorthand)
		}
		str = gb_string_appendc(str, " {")
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			for i := 0; i < len(bf.Fields); i++ {
				if i > 0 {
					str = gb_string_appendc(str, ", ")
				}
				str = write_expr_to_string(str, bf.Fields[i], false)
			}
		}
		str = gb_string_appendc(str, "}")
	case Ast_InlineAsmExpr:
		ia := &node.InlineAsmExpr
		str = gb_string_appendc(str, "asm(")
		for i := 0; i < len(ia.ParamTypes); i++ {
			if i > 0 {
				str = gb_string_appendc(str, ", ")
			}
			str = write_expr_to_string(str, ia.ParamTypes[i], shorthand)
		}
		str = gb_string_appendc(str, ")")
		if ia.ReturnType != nil {
			str = gb_string_appendc(str, " -> ")
			str = write_expr_to_string(str, ia.ReturnType, shorthand)
		}
		if ia.HasSideEffects {
			str = gb_string_appendc(str, " #side_effects")
		}
		if ia.IsAlignStack {
			str = gb_string_appendc(str, " #stack_align")
		}
		if ia.Dialect != 0 {
			str = gb_string_appendc(str, " #")
			str = gb_string_appendc(str, inlineAsmDialectStrings[ia.Dialect])
		}
		str = gb_string_appendc(str, " {")
		if shorthand {
			str = gb_string_appendc(str, "...")
		} else {
			str = write_expr_to_string(str, ia.AsmString, shorthand)
			str = gb_string_appendc(str, ", ")
			str = write_expr_to_string(str, ia.ConstraintsString, shorthand)
		}
		str = gb_string_appendc(str, "}")
	}
	return str
}
