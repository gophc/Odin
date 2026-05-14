package cmd

func lb_build_stmt(p *lbProcedure, node *Ast) {
	prevStmt := p.CurrStmt
	defer func() {
		p.CurrStmt = prevStmt
	}()
	p.CurrStmt = node

	if p.CurrBlock != nil {
		lastInstr := LLVMGetLastInstruction(p.CurrBlock.Block)
		if lb_is_instr_terminating(lastInstr) {
			return
		}
	}

	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, node))
	}

	prevStateFlags := p.StateFlags
	defer func() {
		p.StateFlags = prevStateFlags
	}()
	if node.StateFlags != 0 {
		in := uint16(node.StateFlags)
		out := p.StateFlags
		if in&uint16(StateFlag_BoundsCheck) != 0 {
			out |= uint16(StateFlag_BoundsCheck)
			out &^= uint16(StateFlag_NoBoundsCheck)
		} else if in&uint16(StateFlag_NoBoundsCheck) != 0 {
			out |= uint16(StateFlag_NoBoundsCheck)
			out &^= uint16(StateFlag_BoundsCheck)
		}
		if in&uint16(StateFlag_NoTypeAssert) != 0 {
			out |= uint16(StateFlag_NoTypeAssert)
			out &^= uint16(StateFlag_TypeAssert)
		} else if in&uint16(StateFlag_TypeAssert) != 0 {
			out |= uint16(StateFlag_TypeAssert)
			out &^= uint16(StateFlag_NoTypeAssert)
		}
		p.StateFlags = out
	}

	switch node.Kind {
	case Ast_EmptyStmt:

	case Ast_UsingStmt:

	case Ast_WhenStmt:
		ws := &node.WhenStmt
		lb_build_when_stmt(p, ws)

	case Ast_BlockStmt:
		bs := &node.BlockStmt
		var body, done *lbBlock
		if bs.Label != nil {
			if p.DebugInfo != 0 {
				label := lb_create_block(p, "block.label")
				lb_emit_jump(p, label)
				lb_start_block(p, label)
				LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, bs.Label))
				lb_add_debug_label(p, bs.Label, label)
			}
			body = lb_create_block(p, "block.body")
			done = lb_create_block(p, "block.done")
			lb_emit_jump(p, body)
			lb_start_block(p, body)
			tl := lb_push_target_list(p, bs.Label, done, nil, nil)
			tl.IsBlock = true
		}
		lb_open_scope(p, bs.Scope)
		lb_build_stmt_list(p, bs.Stmts)
		lb_close_scope(p, lbDeferExit_Default, nil, node)
		if done != nil {
			lb_emit_jump(p, done)
			lb_start_block(p, done)
		}
		if bs.Label != nil {
			lb_pop_target_list(p)
		}

	case Ast_ValueDecl:
		vd := &node.ValueDecl
		if !vd.IsMutable {
			return
		}
		isStatic := false
		if len(vd.Names) > 0 {
			for _, name := range vd.Names {
				if !is_blank_ident(name) {
					e := entity_of_node(name)
					if e.Flags&EntityFlag_Static != 0 {
						isStatic = true
						break
					}
				}
			}
		}
		if isStatic {
			lb_build_static_variables(p, vd)
			return
		}
		values := vd.Values
		if len(values) == 0 {
			for i := 0; i < len(vd.Names); i++ {
				name := vd.Names[i]
				if !is_blank_ident(name) {
					e := entity_of_node(name)
					zeroInit := len(values) == 0
					lb_add_local(p, e.Type, e, zeroInit)
				}
			}
		} else {
			if buildContext.EnableRVO {
				if len(vd.Names) == 1 && len(values) == 1 && !is_blank_ident(vd.Names[0]) {
					rhsExpr := unparen_expr(values[0])
					e := entity_of_node(vd.Names[0])
					if rhsExpr.Kind == Ast_CallExpr && e != nil && lb_call_sret_eligible(p, rhsExpr, e.Type) != nil {
						var dst lbValue
						if e == p.SretRvoEntity {
							dst = p.ReturnPtr.Addr
							lb_add_entity(p.Module, e, dst)
							lb_add_debug_local_variable(p, dst.Value, e.Type, e.Token)
						} else {
							local := lb_add_local(p, e.Type, e, true)
							dst = local.Addr
						}
						lb_build_call_expr(p, rhsExpr, &dst)
						break
					}
				}
			}
			lvalsPreused := make([]bool, len(vd.Names))
			lvals := make([]lbAddr, len(vd.Names))
			inits := make([]lbValue, 0, len(lvals))
			lvalIndex := 0
			for _, rhs := range values {
				rhs = unparen_expr(rhs)
				init := lb_build_expr(p, rhs)
				if rhs.Kind == Ast_CompoundLit {
					compLitAddr, ok := p.Module.ExactValueCompoundLiteralAddrMap[rhs]
					if ok {
						if e := entity_of_node(vd.Names[lvalIndex]); e != nil {
							val := compLitAddr.Addr
							lb_add_entity(p.Module, e, val)
							lb_add_debug_local_variable(p, val.Value, e.Type, e.Token)
							lvalsPreused[lvalIndex] = true
							lvals[lvalIndex] = compLitAddr
						}
					}
				}
				lvalIndex += lb_append_tuple_values(p, &inits, init)
			}
			for i := 0; i < len(vd.Names); i++ {
				name := vd.Names[i]
				if !is_blank_ident(name) && !lvalsPreused[i] {
					e := entity_of_node(name)
					zeroInit := len(values) == 0
					lvals[i] = lb_add_local(p, e.Type, e, zeroInit)
				}
			}
			for i := 0; i < len(inits); i++ {
				lval := lvals[i]
				init := inits[i]
				lb_addr_store(p, lval, init)
			}
		}

	case Ast_AssignStmt:
		as := &node.AssignStmt
		lb_build_assign_stmt(p, as)

	case Ast_ExprStmt:
		es := &node.ExprStmt
		lb_build_expr(p, es.Expr)

	case Ast_DeferStmt:
		ds := &node.DeferStmt
		lb_add_defer_node(p, isize(p.ScopeIndex), ds.Stmt)

	case Ast_ReturnStmt:
		rs := &node.ReturnStmt
		lb_build_return_stmt(p, rs.Results, ast_token(node).Pos)

	case Ast_IfStmt:
		lb_build_if_stmt(p, node)

	case Ast_ForStmt:
		lb_build_for_stmt(p, node)

	case Ast_RangeStmt:
		rs := &node.RangeStmt
		lb_build_range_stmt(p, rs, rs.Scope)

	case Ast_UnrollRangeStmt:
		rs := &node.UnrollRangeStmt
		lb_build_unroll_range_stmt(p, rs, rs.Scope)

	case Ast_SwitchStmt:
		ss := &node.SwitchStmt
		lb_build_switch_stmt(p, node, ss.Scope)

	case Ast_TypeSwitchStmt:
		lb_build_type_switch_stmt(p, node)

	case Ast_BranchStmt:
		bs := &node.BranchStmt
		var block *lbBlock
		if bs.Label != nil {
			bb := lb_lookup_branch_blocks(p, bs.Label)
			switch bs.Token.Kind {
			case Token_break:
				block = bb.Break_
			case Token_continue:
				block = bb.Continue
			case Token_fallthrough:
				gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 3381, "fallthrough cannot have a label")
			}
		} else {
			for t := p.TargetList; t != nil && block == nil; t = t.Prev {
				if t.IsBlock {
					continue
				}
				switch bs.Token.Kind {
				case Token_break:
					block = t.Break_
				case Token_continue:
					block = t.Continue_
				case Token_fallthrough:
					block = t.Fallthrough
				}
			}
		}
		if block != nil {
			lb_emit_defer_stmts(p, lbDeferExit_Branch, block, node)
		}
		lb_emit_jump(p, block)
		lb_start_block(p, lb_create_block(p, "unreachable"))
	}
}
