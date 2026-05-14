package cmd

const (
	Stmt_BreakAllowed       = 1 << 0
	Stmt_ContinueAllowed    = 1 << 1
	Stmt_FallthroughAllowed = 1 << 2
	Stmt_TypeSwitch         = 1 << 4
	Stmt_CheckScopeDecls    = 1 << 5
)

func check_stmt_list(ctx *CheckerContext, stmts []*Ast, flags u32) {
	if len(stmts) == 0 {
		return
	}
	if flags&Stmt_CheckScopeDecls != 0 {
		check_scope_decls(ctx, stmts, isize(1.2*float64(len(stmts))))
	}
	ft_ok := flags&Stmt_FallthroughAllowed != 0
	flags &= ^Stmt_FallthroughAllowed
	max := len(stmts)
	for i := len(stmts) - 1; i >= 0; i-- {
		if stmts[i].Kind != Ast_EmptyStmt {
			break
		}
		max--
	}
	max_non_constant_declaration := len(stmts)
	for i := len(stmts) - 1; i >= 0; i-- {
		if stmts[i].Kind == Ast_EmptyStmt {
		} else if stmts[i].Kind == Ast_ValueDecl && !stmts[i].ValueDecl.is_mutable {
		} else {
			break
		}
		max_non_constant_declaration--
	}
	for i := 0; i < max; i++ {
		n := stmts[i]
		if n.Kind == Ast_EmptyStmt {
			continue
		}
		new_flags := flags
		if ft_ok && i+1 == max {
			new_flags |= Stmt_FallthroughAllowed
		}
		prev := ctx.stmt_flags
		ctx.stmt_flags = new_flags
		check_stmt(ctx, n, new_flags)
		ctx.stmt_flags = prev
		if i+1 < max_non_constant_declaration {
			switch n.Kind {
			case Ast_ReturnStmt:
				error(n, "Statements after this 'return' are never executed")
			case Ast_BranchStmt:
				error(n, "Statements after this '%s' are never executed", n.BranchStmt.token.string)
			case Ast_ExprStmt:
				if is_diverging_stmt(n) {
					error(n, "Statements after a diverging procedure call are never executed")
				}
			}
		} else if i+1 == max_non_constant_declaration {
			if is_diverging_stmt(n) {
				for j := 0; j < i; j++ {
					stmt := stmts[j]
					if stmt.Kind == Ast_ValueDecl && !stmt.ValueDecl.is_mutable {
					} else if stmt.Kind == Ast_DeferStmt {
						error(stmt, "Unreachable defer statement due to diverging procedure call at the end of the current scope")
					} else if contains_deferred_call(stmt) {
						error(stmt, "Unreachable deferred procedure call due to a diverging procedure call at the end of the current scope")
					}
				}
			}
		}
	}
}

func check_stmt(ctx *CheckerContext, node *Ast, flags u32) {
	prev_state_flags := ctx.state_flags
	if node.state_flags != 0 {
		in := u32(node.state_flags)
		out := u32(ctx.state_flags)
		if in&StateFlag_NoBoundsCheck != 0 {
			out |= StateFlag_NoBoundsCheck
			out &^= StateFlag_BoundsCheck
		} else if in&StateFlag_BoundsCheck != 0 {
			out |= StateFlag_BoundsCheck
			out &^= StateFlag_NoBoundsCheck
		}
		if in&StateFlag_NoTypeAssert != 0 {
			out |= StateFlag_NoTypeAssert
			out &^= StateFlag_TypeAssert
		} else if in&StateFlag_TypeAssert != 0 {
			out |= StateFlag_TypeAssert
			out &^= StateFlag_NoTypeAssert
		}
		ctx.state_flags = StateFlag(out)
	}
	check_stmt_internal(ctx, node, flags)
	ctx.state_flags = prev_state_flags
}

func check_when_stmt(ctx *CheckerContext, node *Ast, flags u32) {
	ws := &node.WhenStmt
	var operand Operand
	check_expr(ctx, &operand, ws.cond)
	if operand.Mode != Addressing_Constant || !is_type_boolean(operand.Type) {
		error(ws.cond, "Non-constant boolean 'when' condition")
		return
	}
	if ws.body == nil || ws.body.Kind != Ast_BlockStmt {
		error(ws.cond, "Invalid body for 'when' statement")
		return
	}
	if operand.Value.Kind == ExactValue_Bool && operand.Value.ValueBool {
		check_stmt_list(ctx, ws.body.BlockStmt.stmts, flags)
	} else if ws.else_stmt != nil {
		switch ws.else_stmt.Kind {
		case Ast_BlockStmt:
			check_stmt_list(ctx, ws.else_stmt.BlockStmt.stmts, flags)
		case Ast_WhenStmt:
			check_when_stmt(ctx, ws.else_stmt, flags)
		default:
			error(ws.else_stmt, "Invalid 'else' statement in 'when' statement")
		}
	}
}

func check_label(ctx *CheckerContext, label *Ast, parent *Ast) {
	if label == nil {
		return
	}
	if label.Kind != Ast_Label {
		return
	}
	l := &label.Label
	if l.name.Kind != Ast_Ident {
		error(l.name, "A label's name must be an identifier")
		return
	}
	name := l.name.Ident.token.string
	if is_blank_ident_str(name) {
		error(l.name, "A label's name cannot be a blank identifier")
		return
	}
	if ctx.curr_proc_decl == nil {
		error(l.name, "A label is only allowed within a procedure")
		return
	}
	ok := true
	for i := 0; i < len(ctx.decl.labels); i++ {
		bl := ctx.decl.labels[i]
		if bl.Name == name {
			error(label, "Duplicate label with the name '%.*s'", name.Len, name.Data)
			ok = false
			break
		}
	}
	e := alloc_entity_label(ctx.scope, l.name.Ident.token, t_invalid, label, parent)
	add_entity(ctx, ctx.scope, l.name, e)
	e.parent_proc_decl = ctx.curr_proc_decl
	if ok {
		bl := BlockLabel{Name: name, Label: label}
		ctx.decl.labels = append(ctx.decl.labels, bl)
	}
}

func check_block_stmt_for_errors(ctx *CheckerContext, body *Ast) {
	if body.Kind != Ast_BlockStmt {
		return
	}
	bs := &body.BlockStmt
	if bs.scope != nil && len(bs.scope.elements) > 0 {
		if bs.scope.parent.node != nil {
			switch bs.scope.parent.node.Kind {
			case Ast_IfStmt, Ast_ForStmt, Ast_RangeStmt, Ast_UnrollRangeStmt, Ast_SwitchStmt, Ast_TypeSwitchStmt:
			default:
				return
			}
		}
		stmt_count := 0
		var the_stmt *Ast
		for _, stmt := range bs.stmts {
			switch stmt.Kind {
			case Ast_EmptyStmt, Ast_BadStmt, Ast_BadDecl:
			default:
				the_stmt = stmt
				stmt_count++
			}
		}
		if stmt_count == 1 && the_stmt.Kind == Ast_ValueDecl {
			for _, name := range the_stmt.ValueDecl.names {
				if name.Kind != Ast_Ident {
					continue
				}
				n := name.Ident.token.string
				if n != "_" {
					var s string
					s = string(n.Data[:n.Len])
					_ = s
					error(name, "'%.*s' declared but not used", n.Len, n.Data)
				}
			}
		}
	}
}

func check_if_stmt(ctx *CheckerContext, node *Ast, mod_flags u32) {
	is := &node.IfStmt
	check_open_scope(ctx, node)
	check_label(ctx, is.label, node)
	if is.init != nil {
		check_stmt(ctx, is.init, 0)
	}
	var operand Operand
	check_expr(ctx, &operand, is.cond)
	if operand.Mode != Addressing_Invalid && !is_type_boolean(operand.Type) {
		error(is.cond, "Non-boolean condition in 'if' statement")
	}
	check_stmt(ctx, is.body, mod_flags)
	if is.else_stmt != nil {
		switch is.else_stmt.Kind {
		case Ast_IfStmt, Ast_BlockStmt:
			check_stmt(ctx, is.else_stmt, mod_flags)
		default:
			error(is.else_stmt, "Invalid 'else' statement in 'if' statement")
		}
	}
	check_close_scope(ctx)
}

func check_for_stmt(ctx *CheckerContext, node *Ast, mod_flags u32) {
	fs := &node.ForStmt
	mod_flags |= Stmt_BreakAllowed | Stmt_ContinueAllowed
	check_open_scope(ctx, node)
	check_label(ctx, fs.label, node)
	if fs.init != nil {
		check_stmt(ctx, fs.init, 0)
	}
	if fs.cond != nil {
		var o Operand
		check_expr(ctx, &o, fs.cond)
		if o.Mode != Addressing_Invalid && !is_type_boolean(o.Type) {
			error(fs.cond, "Non-boolean condition in 'for' statement")
		} else {
			cond := unparen_expr(o.Expr)
			if cond != nil && cond.Kind == Ast_BinaryExpr &&
				cond.BinaryExpr.left != nil && cond.BinaryExpr.right != nil {
				if cond.BinaryExpr.op.Kind == Token_GtEq &&
					type_of_expr(cond.BinaryExpr.left) != nil &&
					is_type_unsigned(type_of_expr(cond.BinaryExpr.left)) &&
					cond.BinaryExpr.right.TAV.Value.Kind == ExactValue_Integer &&
					is_exact_value_zero(cond.BinaryExpr.right.TAV.Value) {
					warning(cond, "Expression is always true since unsigned numbers are always >= 0")
				} else if cond.BinaryExpr.op.Kind == Token_LtEq &&
					type_of_expr(cond.BinaryExpr.right) != nil &&
					is_type_unsigned(type_of_expr(cond.BinaryExpr.right)) &&
					cond.BinaryExpr.left.TAV.Value.Kind == ExactValue_Integer &&
					is_exact_value_zero(cond.BinaryExpr.left.TAV.Value) {
					warning(cond, "Expression is always true since unsigned numbers are always >= 0")
				}
			}
		}
	}
	if fs.post != nil {
		check_stmt(ctx, fs.post, 0)
		if fs.post.Kind != Ast_AssignStmt {
			error(fs.post, "'for' statement post statement must be a simple statement")
		}
	}
	check_stmt(ctx, fs.body, mod_flags)
	check_close_scope(ctx)
}

func check_return_stmt(ctx *CheckerContext, node *Ast) {
	rs := &node.ReturnStmt
	if ctx.in_defer {
		error(rs.token, "'return' cannot be used within a defer statement")
		return
	}
	proc_type := ctx.curr_proc_sig
	if proc_type.Kind != Type_Proc {
		return
	}
	pt := &proc_type.Proc
	if pt.diverging {
		error(rs.token, "Diverging procedures may not return")
		return
	}
	var result_entities []*Entity
	result_count := 0
	has_named_results := pt.has_named_results
	if pt.results != nil {
		result_entities = pt.results.Tuple.variables
		result_count = len(pt.results.Tuple.variables)
	}
	operands := make([]Operand, 0, 2*len(rs.results))
	check_unpack_arguments(ctx, result_entities, result_count, &operands, rs.results, UnpackFlag_AllowOk)
	if result_count == 0 && len(rs.results) > 0 {
		error(rs.results[0], "No return values expected")
	} else if has_named_results && len(operands) == 0 {
	} else if len(operands) != result_count {
		if all_operands_valid(operands) {
			if len(operands) == 1 {
				t := type_to_string(operands[0].Type)
				error(node, "Expected %td return values, got %td (%s)", result_count, len(operands), t)
				gb_string_free(t)
			} else {
				error(node, "Expected %td return values, got %td", result_count, len(operands))
			}
		}
	} else {
		for i := 0; i < result_count; i++ {
			e := pt.results.Tuple.variables[i]
			o := &operands[i]
			check_assignment(ctx, o, e.Type, S("return statement"))
			if is_type_untyped(o.Type) {
				update_untyped_expr_type(ctx, o.Expr, e.Type, true)
			}
		}
	}
	for _, o := range operands {
		if o.Expr == nil {
			continue
		}
		expr := unparen_expr(o.Expr)
		for expr.Kind == Ast_CallExpr && expr.CallExpr.proc.TAV.Mode == Addressing_Type {
			if len(expr.CallExpr.args) != 1 {
				break
			}
			arg := expr.CallExpr.args[0]
			if arg.Kind == Ast_FieldValue || !are_types_identical(arg.TAV.Type, expr.TAV.Type) {
				break
			}
			expr = unparen_expr(arg)
		}
		check_unsafe_return(o, o.Type, expr)
	}
}

func check_unsafe_return(o Operand, type_ *Type, expr *Ast) {
	unsafe_return_error := func(o Operand, msg string, extra_type ...*Type) {
		s := expr_to_string(o.Expr)
		if len(extra_type) > 0 && extra_type[0] != nil {
			t := type_to_string(extra_type[0])
			error(o.Expr, "It is unsafe to return %s ('%s') of type ('%s') from a procedure, as it uses the current stack frame's memory", msg, s, t)
			gb_string_free(t)
		} else {
			error(o.Expr, "It is unsafe to return %s ('%s') from a procedure, as it uses the current stack frame's memory", msg, s)
		}
		gb_string_free(s)
	}
	if type_ == nil || expr == nil {
		return
	}
	if expr.Kind == Ast_CompoundLit && is_type_slice(type_) {
		cl := &expr.CompoundLit
		if len(cl.elems) == 0 {
			return
		}
		unsafe_return_error(o, "a compound literal of a slice")
	} else if expr.Kind == Ast_UnaryExpr && expr.UnaryExpr.op.Kind == Token_And {
		x := unparen_expr(expr.UnaryExpr.expr)
		e := entity_of_node(x)
		if is_entity_local_variable(e) {
			unsafe_return_error(o, "the address of a local variable")
		} else if x.Kind == Ast_CompoundLit {
			unsafe_return_error(o, "the address of a compound literal")
		} else if x.Kind == Ast_IndexExpr {
			f := entity_of_node(x.IndexExpr.expr)
			if f != nil && (is_type_array_like(f.Type) || is_type_matrix(f.Type)) {
				if is_entity_local_variable(f) {
					unsafe_return_error(o, "the address of an indexed variable", f.Type)
				}
			}
		} else if x.Kind == Ast_MatrixIndexExpr {
			f := entity_of_node(x.MatrixIndexExpr.expr)
			if f != nil && is_type_matrix(f.Type) && is_entity_local_variable(f) {
				unsafe_return_error(o, "the address of an indexed variable", f.Type)
			}
		}
	} else if expr.Kind == Ast_SliceExpr {
		x := unparen_expr(expr.SliceExpr.expr)
		e := entity_of_node(x)
		if is_entity_local_variable(e) && is_type_array(e.Type) {
			unsafe_return_error(o, "a slice of a local variable")
		} else if x.Kind == Ast_CompoundLit {
			unsafe_return_error(o, "a slice of a compound literal")
		}
	} else if o.Mode == Addressing_Constant && is_type_slice(type_) {
		if is_load_directive_call(o.Expr) {
			return
		}
		begin_error_block()
		unsafe_return_error(o, "a compound literal of a slice")
		error_line("\tNote: A constant slice value will use the memory of the current stack frame\n")
		end_error_block()
	} else if expr.Kind == Ast_CompoundLit {
		cl := &expr.CompoundLit
		for _, elem := range cl.elems {
			if elem.Kind == Ast_FieldValue {
				fv := &elem.FieldValue
				e := entity_of_node(fv.field)
				if e != nil {
					check_unsafe_return(o, e.Type, fv.value)
				}
			}
		}
	}
}

func check_stmt_internal(ctx *CheckerContext, node *Ast, flags u32) {
	mod_flags := flags & ^Stmt_FallthroughAllowed
	switch node.Kind {
	case Ast_EmptyStmt, Ast_BadStmt, Ast_BadDecl:
	case Ast_ExprStmt:
		check_expr_stmt(ctx, node)
	case Ast_AssignStmt:
		check_assign_stmt(ctx, node)
	case Ast_BlockStmt:
		bs := &node.BlockStmt
		check_open_scope(ctx, node)
		check_label(ctx, bs.label, node)
		check_stmt_list(ctx, bs.stmts, flags)
		check_block_stmt_for_errors(ctx, node)
		check_close_scope(ctx)
	case Ast_IfStmt:
		check_if_stmt(ctx, node, mod_flags)
	case Ast_WhenStmt:
		check_when_stmt(ctx, node, flags)
	case Ast_ReturnStmt:
		check_return_stmt(ctx, node)
	case Ast_ForStmt:
		check_for_stmt(ctx, node, mod_flags)
	case Ast_RangeStmt:
		check_range_stmt(ctx, node, mod_flags)
	case Ast_UnrollRangeStmt:
		check_unroll_range_stmt(ctx, node, mod_flags)
	case Ast_SwitchStmt:
		check_switch_stmt(ctx, node, mod_flags)
	case Ast_TypeSwitchStmt:
		check_type_switch_stmt(ctx, node, mod_flags)
	case Ast_DeferStmt:
		ds := &node.DeferStmt
		if is_ast_decl(ds.stmt) {
			error(ds.token, "You cannot defer a declaration")
		} else {
			out_in_defer := ctx.in_defer
			ctx.in_defer = true
			check_stmt(ctx, ds.stmt, 0)
			ctx.in_defer = out_in_defer
			if ctx.decl != nil {
				ctx.decl.defer_used++
			}
			stmt := ds.stmt
			original_stmt := stmt
			if stmt.Kind == Ast_BlockStmt && len(stmt.BlockStmt.stmts) == 0 {
				break
			}
			is_singular := true
			for is_singular && stmt.Kind == Ast_BlockStmt {
				var inner_stmt *Ast
				for _, s := range stmt.BlockStmt.stmts {
					if s.Kind == Ast_EmptyStmt {
						continue
					}
					if inner_stmt != nil {
						is_singular = false
						break
					}
					inner_stmt = s
				}
				if inner_stmt != nil {
					stmt = inner_stmt
				}
			}
			if !is_singular {
				stmt = original_stmt
			}
			if stmt.Kind == Ast_AssignStmt {
				as := &stmt.AssignStmt
				if as.op.Kind == Token_Eq {
					for _, lhs := range as.lhs {
						e := entity_of_node(lhs)
						if e != nil && e.Flags&EntityFlag_Result != 0 {
							error(lhs, "Assignments to named return values within 'defer' will not affect the value that is returned")
						}
					}
				}
			}
		}
	case Ast_BranchStmt:
		bs := &node.BranchStmt
		token := bs.token
		switch token.Kind {
		case Token_break:
			if flags&Stmt_BreakAllowed == 0 && bs.label == nil {
				error(token, "'break' only allowed in non-inline loops or 'switch' statements")
			}
		case Token_continue:
			if flags&Stmt_ContinueAllowed == 0 && bs.label == nil {
				error(token, "'continue' only allowed in non-inline loops")
			}
		case Token_fallthrough:
			if flags&Stmt_FallthroughAllowed == 0 {
				if flags&Stmt_TypeSwitch != 0 {
					error(token, "'fallthrough' statement not allowed within a type switch statement")
				} else {
					error(token, "'fallthrough' statement in illegal position, expected at the end of a 'case' block")
				}
			} else if bs.label != nil {
				error(token, "'fallthrough' cannot have a label")
			}
		default:
			error(token, "Invalid AST: Branch Statement '%.*s'", token.String.Len, token.String.Data)
		}
		if bs.label != nil {
			if bs.label.Kind != Ast_Ident {
				error(bs.label, "A branch statement's label name must be an identifier")
				return
			}
			ident := bs.label
			name := ident.Ident.token.string
			var o Operand
			e := check_ident(ctx, &o, ident, nil, nil, false)
			if e == nil {
				error(ident, "Undeclared label name: %.*s", name.Len, name.Data)
				return
			}
			add_entity_use(ctx, ident, e)
			if e.Kind != Entity_Label {
				error(ident, "'%.*s' is not a label", name.Len, name.Data)
				return
			}
			parent := e.Label.parent
			if parent != nil {
				switch parent.Kind {
				case Ast_BlockStmt, Ast_IfStmt, Ast_SwitchStmt, Ast_TypeSwitchStmt:
					if token.Kind != Token_break {
						error(bs.label, "Label '%.*s' can only be used with 'break'", e.Token.String.Len, e.Token.String.Data)
					}
				case Ast_RangeStmt, Ast_ForStmt:
					if token.Kind != Token_break && token.Kind != Token_continue {
						error(bs.label, "Label '%.*s' can only be used with 'break' and 'continue'", e.Token.String.Len, e.Token.String.Data)
					}
				}
			}
			if ctx.in_defer {
				error(bs.label, "A labelled '%.*s' cannot be used within a 'defer'", token.String.Len, token.String.Data)
			}
		}
	case Ast_UsingStmt:
		us := &node.UsingStmt
		if len(us.list) == 0 {
			error(us.token, "Empty 'using' list")
			return
		}
		feature_flags := check_feature_flags(ctx, node)
		if feature_flags&OptInFeatureFlag_UsingStmt == 0 {
			begin_error_block()
			error(node, "'using' has been disallowed as it is considered bad practice to use as a statement outside of immediate refactoring")
			error_line("\tIf you do require it for refactoring purposes or legacy code, it can be enabled on a per-file basis with '#+feature using-stmt'\n")
			end_error_block()
		}
		for _, expr := range us.list {
			expr = unparen_expr(expr)
			var e *Entity
			is_selector := false
			var o Operand
			switch expr.Kind {
			case Ast_Ident:
				e = check_ident(ctx, &o, expr, nil, nil, true)
			case Ast_SelectorExpr:
				e = check_selector(ctx, &o, expr, nil)
				is_selector = true
			case Ast_Implicit:
				error(us.token, "'using' applied to an implicit value")
				continue
			default:
				error(us.token, "'using' can only be applied to an entity, got %.*s", astStrings[expr.Kind].Len, astStrings[expr.Kind].Data)
				continue
			}
			if !check_using_stmt_entity(ctx, us, expr, is_selector, e) {
				return
			}
		}
	case Ast_ForeignBlockDecl:
		fb := &node.ForeignBlockDecl
		foreign_library := fb.foreign_library
		c := *ctx
		if foreign_library.Kind != Ast_Ident {
			error(foreign_library, "foreign library name must be an identifier")
		} else {
			c.foreign_context.curr_library = foreign_library
			c.foreign_context.default_cc = ProcCC_CDecl
		}
		check_decl_attributes(&c, fb.attributes, foreign_block_decl_attribute, nil)
		block := &fb.body.BlockStmt
		for _, decl := range block.stmts {
			if decl.Kind == Ast_ValueDecl && decl.ValueDecl.is_mutable {
				check_stmt(&c, decl, flags)
			}
		}
	case Ast_ValueDecl:
		check_value_decl_stmt(ctx, node, mod_flags)
	}
}
