package cmd

func is_diverging_expr(expr *Ast) bool {
	expr = unparen_expr(expr)
	if expr.Kind != Ast_CallExpr {
		return false
	}
	if expr.CallExpr.proc.Kind == Ast_BasicDirective {
		name := expr.CallExpr.proc.BasicDirective.name.string
		return name == "panic"
	}
	proc := unparen_expr(expr.CallExpr.proc)
	tv := proc.tav
	if tv.mode == Addressing_Builtin {
		e := entity_of_node(proc)
		id := BuiltinProc_Invalid
		if e != nil {
			id = BuiltinProcId(e.Builtin.id)
		} else {
			id = BuiltinProc_DIRECTIVE
		}
		return builtin_procs[id].diverging
	}
	t := base_type(tv.type)
	return t != nil && t.Kind == Type_Proc && t.Proc.diverging
}

func is_diverging_stmt(stmt *Ast) bool {
	if stmt.Kind != Ast_ExprStmt {
		return false
	}
	return is_diverging_expr(stmt.ExprStmt.expr)
}

func contains_deferred_call(node *Ast) bool {
	if node.viral_state_flags&ViralStateFlag_ContainsDeferredProcedure != 0 {
		return true
	}
	switch node.Kind {
	case Ast_ExprStmt:
		return contains_deferred_call(node.ExprStmt.expr)
	case Ast_AssignStmt:
		for i := 0; i < len(node.AssignStmt.rhs); i++ {
			if contains_deferred_call(node.AssignStmt.rhs[i]) {
				return true
			}
		}
		for i := 0; i < len(node.AssignStmt.lhs); i++ {
			if contains_deferred_call(node.AssignStmt.lhs[i]) {
				return true
			}
		}
	case Ast_ValueDecl:
		for i := 0; i < len(node.ValueDecl.values); i++ {
			if contains_deferred_call(node.ValueDecl.values[i]) {
				return true
			}
		}
	}
	return false
}

func check_is_terminating_list(stmts []*Ast, label string) bool {
	for n := len(stmts) - 1; n >= 0; n-- {
		stmt := stmts[n]
		if stmt.Kind == Ast_EmptyStmt {
		} else if stmt.Kind == Ast_ValueDecl && !stmt.ValueDecl.is_mutable {
		} else if is_diverging_stmt(stmt) {
			return true
		} else {
			return check_is_terminating(stmt, label)
		}
	}
	return false
}

func check_has_break_list(stmts []*Ast, label string, implicit bool) bool {
	for _, stmt := range stmts {
		if check_has_break(stmt, label, implicit) {
			return true
		}
	}
	return false
}

func check_has_break_expr(expr *Ast, label string) bool {
	if expr != nil && expr.viral_state_flags&ViralStateFlag_ContainsOrBreak != 0 {
		return true
	}
	return false
}

func check_has_break_expr_list(exprs []*Ast, label string) bool {
	for _, expr := range exprs {
		if check_has_break_expr(expr, label) {
			return true
		}
	}
	return false
}

func check_has_break(stmt *Ast, label string, implicit bool) bool {
	switch stmt.Kind {
	case Ast_BranchStmt:
		if stmt.BranchStmt.token.kind == Token_break {
			if stmt.BranchStmt.label == nil {
				return implicit
			}
			if stmt.BranchStmt.label.Kind == Ast_Ident && stmt.BranchStmt.label.Ident.token.string == label {
				return true
			}
		}
	case Ast_DeferStmt:
		return check_has_break(stmt.DeferStmt.stmt, label, implicit)
	case Ast_BlockStmt:
		return check_has_break_list(stmt.BlockStmt.stmts, label, implicit)
	case Ast_IfStmt:
		if stmt.IfStmt.init != nil && check_has_break(stmt.IfStmt.init, label, implicit) {
			return true
		}
		if stmt.IfStmt.cond != nil && check_has_break_expr(stmt.IfStmt.cond, label) {
			return true
		}
		if check_has_break(stmt.IfStmt.body, label, implicit) || (stmt.IfStmt.else_stmt != nil && check_has_break(stmt.IfStmt.else_stmt, label, implicit)) {
			return true
		}
	case Ast_CaseClause:
		return check_has_break_list(stmt.CaseClause.stmts, label, implicit)
	case Ast_SwitchStmt:
		if stmt.SwitchStmt.init != nil && check_has_break_expr(stmt.SwitchStmt.init, label) {
			return true
		}
		if label != "" && check_has_break(stmt.SwitchStmt.body, label, false) {
			return true
		}
	case Ast_TypeSwitchStmt:
		if label != "" && check_has_break(stmt.TypeSwitchStmt.body, label, false) {
			return true
		}
	case Ast_ForStmt:
		if stmt.ForStmt.init != nil && check_has_break(stmt.ForStmt.init, label, implicit) {
			return true
		}
		if stmt.ForStmt.cond != nil && check_has_break_expr(stmt.ForStmt.cond, label) {
			return true
		}
		if stmt.ForStmt.post != nil && check_has_break(stmt.ForStmt.post, label, implicit) {
			return true
		}
		if label != "" && check_has_break(stmt.ForStmt.body, label, false) {
			return true
		}
	case Ast_RangeStmt:
		if label != "" && check_has_break(stmt.RangeStmt.body, label, false) {
			return true
		}
	case Ast_ExprStmt:
		if stmt.ExprStmt.expr.viral_state_flags&ViralStateFlag_ContainsOrBreak != 0 {
			return true
		}
	case Ast_ValueDecl:
		if stmt.ValueDecl.is_mutable && check_has_break_expr_list(stmt.ValueDecl.values, label) {
			return true
		}
	case Ast_AssignStmt:
		if check_has_break_expr_list(stmt.AssignStmt.lhs, label) {
			return true
		}
		if check_has_break_expr_list(stmt.AssignStmt.rhs, label) {
			return true
		}
	}
	return false
}

func label_string(node *Ast) string {
	if node == nil {
		return ""
	}
	if node.Kind == Ast_Ident {
		return node.Ident.token.string
	} else if node.Kind == Ast_Label {
		return label_string(node.Label.name)
	}
	return ""
}

func check_is_terminating(node *Ast, label string) bool {
	switch node.Kind {
	case Ast_ReturnStmt:
		return true
	case Ast_BlockStmt:
		if check_is_terminating_list(node.BlockStmt.stmts, label) {
			if node.BlockStmt.label != nil {
				return check_is_terminating_list(node.BlockStmt.stmts, label_string(node.BlockStmt.label))
			}
			return true
		}
	case Ast_ExprStmt:
		return check_is_terminating(unparen_expr(node.ExprStmt.expr), label)
	case Ast_ValueDecl:
		return check_has_break_expr_list(node.ValueDecl.values, label)
	case Ast_AssignStmt:
		return check_has_break_expr_list(node.AssignStmt.lhs, label) || check_has_break_expr_list(node.AssignStmt.rhs, label)
	case Ast_BranchStmt:
		return node.BranchStmt.token.kind == Token_fallthrough
	case Ast_IfStmt:
		if node.IfStmt.else_stmt != nil {
			if check_is_terminating(node.IfStmt.body, label) && check_is_terminating(node.IfStmt.else_stmt, label) {
				return true
			}
		}
	case Ast_WhenStmt:
		tv := node.WhenStmt.cond.tav
		if tv.mode != Addressing_Constant {
			if node.WhenStmt.else_stmt != nil {
				if check_is_terminating(node.WhenStmt.body, label) && check_is_terminating(node.WhenStmt.else_stmt, label) {
					return true
				}
			}
			return false
		}
		if tv.value.kind == ExactValue_Bool {
			if tv.value.value_bool {
				return check_is_terminating(node.WhenStmt.body, label)
			} else {
				if node.WhenStmt.else_stmt == nil {
					return false
				}
				return check_is_terminating(node.WhenStmt.else_stmt, label)
			}
		}
	case Ast_ForStmt:
		if node.ForStmt.cond == nil && !check_has_break(node.ForStmt.body, label, true) {
			if node.ForStmt.label != nil {
				return !check_has_break(node.ForStmt.body, label_string(node.ForStmt.label), false)
			}
			return true
		}
	case Ast_UnrollRangeStmt:
		return false
	case Ast_RangeStmt:
		return false
	case Ast_SwitchStmt:
		has_default := false
		for i := 0; i < len(node.SwitchStmt.body.BlockStmt.stmts); i++ {
			clause := node.SwitchStmt.body.BlockStmt.stmts[i]
			if len(clause.CaseClause.list) == 0 {
				has_default = true
			}
			if !check_is_terminating_list(clause.CaseClause.stmts, label) || check_has_break_list(clause.CaseClause.stmts, label, true) {
				return false
			}
		}
		return has_default
	case Ast_TypeSwitchStmt:
		has_default := false
		for i := 0; i < len(node.TypeSwitchStmt.body.BlockStmt.stmts); i++ {
			clause := node.TypeSwitchStmt.body.BlockStmt.stmts[i]
			if len(clause.CaseClause.list) == 0 {
				has_default = true
			}
			if !check_is_terminating_list(clause.CaseClause.stmts, label) || check_has_break_list(clause.CaseClause.stmts, label, true) {
				return false
			}
		}
		return has_default
	}
	return false
}

func all_operands_valid(operands []Operand) bool {
	if any_errors() {
		for _, o := range operands {
			if o.type == t_invalid {
				return false
			}
		}
	}
	return true
}

func check_stmt_internal_builtin_proc_id(expr *Ast, id_ *BuiltinProcId) bool {
	id := BuiltinProc_Invalid
	e := entity_of_node(expr)
	if e != nil && e.Kind == Entity_Builtin {
		if e.Builtin.id != 0 && e.Builtin.id != BuiltinProc_DIRECTIVE {
			id = BuiltinProcId(e.Builtin.id)
		}
	}
	if id_ != nil {
		*id_ = id
	}
	return id != BuiltinProc_Invalid
}
