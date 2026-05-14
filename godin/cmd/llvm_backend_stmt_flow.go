package cmd

func lb_build_stmt_list(p *lbProcedure, stmts []*Ast) {
	for _, stmt := range stmts {
		switch stmt.Kind {
		case Ast_ValueDecl:
			lb_build_constant_value_decl(p, &stmt.ValueDecl)
		case Ast_ForeignBlockDecl:
			fb := &stmt.ForeignBlockDecl
			block := &fb.Body.BlockStmt
			lb_build_stmt_list(p, block.Stmts)
		}
	}
	for _, stmt := range stmts {
		lb_build_stmt(p, stmt)
	}
}

func lb_lookup_branch_blocks(p *lbProcedure, ident *Ast) lbBranchBlocks {
	e := entity_of_node(ident)
	for _, b := range p.BranchBlocks {
		if b.Label == e.Label.Node {
			return b
		}
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 257, "Unreachable")
	return lbBranchBlocks{}
}

func lb_push_target_list(p *lbProcedure, label *Ast, break_, continue_, fallthrough_ *lbBlock) *lbTargetList {
	tl := &lbTargetList{
		Prev:        p.TargetList,
		Break_:      break_,
		Continue_:   continue_,
		Fallthrough: fallthrough_,
	}
	p.TargetList = tl
	if label != nil {
		for i := range p.BranchBlocks {
			b := &p.BranchBlocks[i]
			if b.Label == label {
				b.Break_ = break_
				b.Continue = continue_
				return tl
			}
		}
		gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 283, "Unreachable")
	}
	return tl
}

func lb_pop_target_list(p *lbProcedure) {
	p.TargetList = p.TargetList.Prev
}

func lb_open_scope(p *lbProcedure, s *Scope) {
	m := p.Module
	if m.DebugBuilder != 0 {
		curr_metadata := lb_get_llvm_metadata(m, unsafe.Pointer(s))
		if s != nil && s.Node != nil && curr_metadata == 0 {
			token := ast_token(s.Node)
			line := uint(token.Pos.Line)
			column := uint(token.Pos.Column)
			var file LLVMMetadataRef
			astFile := s.Node.file()
			if astFile != nil {
				file = lb_get_llvm_metadata(m, unsafe.Pointer(astFile))
			}
			var scope LLVMMetadataRef
			if len(p.ScopeStack) > 0 {
				scope = lb_get_llvm_metadata(m, unsafe.Pointer(p.ScopeStack[len(p.ScopeStack)-1]))
			}
			if scope == 0 {
				scope = lb_get_llvm_metadata(m, unsafe.Pointer(p))
			}
			if m.DebugBuilder != 0 {
				res := LLVMDIBuilderCreateLexicalBlock(m.DebugBuilder, scope, file, line, column)
				lb_set_llvm_metadata(m, unsafe.Pointer(s), res)
			}
		}
	}
	p.CurrScope = s
	p.ScopeIndex += 1
	p.ScopeStack = append(p.ScopeStack, s)
}

func lb_close_scope(p *lbProcedure, kind lbDeferExitKind, block *lbBlock, node *Ast, pop_stack ...bool) {
	popStk := true
	if len(pop_stack) > 0 {
		popStk = pop_stack[0]
	}
	lb_emit_defer_stmts(p, kind, block, node)
	for len(p.ContextStack) > 0 {
		ctx := &p.ContextStack[len(p.ContextStack)-1]
		if ctx.ScopeIndex >= isize(p.ScopeIndex) {
			p.ContextStack = p.ContextStack[:len(p.ContextStack)-1]
		} else {
			break
		}
	}
	if p.CurrScope != nil {
		p.CurrScope = p.CurrScope.Parent
	}
	p.ScopeIndex -= 1
	if popStk {
		p.ScopeStack = p.ScopeStack[:len(p.ScopeStack)-1]
	}
}

func lb_build_when_stmt(p *lbProcedure, ws *AstWhenStmt) {
	tv := type_and_value_of_expr(ws.Cond)
	if tv.Value.Kind == ExactValue_Bool && tv.Value.ValueBool {
		lb_build_stmt_list(p, ws.Body.BlockStmt.Stmts)
	} else if ws.ElseStmt != nil {
		switch ws.ElseStmt.Kind {
		case Ast_BlockStmt:
			lb_build_stmt_list(p, ws.ElseStmt.BlockStmt.Stmts)
		case Ast_WhenStmt:
			lb_build_when_stmt(p, &ws.ElseStmt.WhenStmt)
		default:
			gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 370, "Invalid 'else' statement in 'when' statement")
		}
	}
}

func lb_build_if_stmt(p *lbProcedure, node *Ast) {
	is := &node.IfStmt
	lb_open_scope(p, is.Scope)
	defer lb_close_scope(p, lbDeferExit_Default, nil, node)
	thenBlock := lb_create_block(p, "if.then")
	doneBlock := lb_create_block(p, "if.done")
	elseBlock := doneBlock
	if is.ElseStmt != nil {
		elseBlock = lb_create_block(p, "if.else")
	}
	if is.Label != nil {
		if p.DebugInfo != 0 {
			label := lb_create_block(p, "if.label")
			lb_emit_jump(p, label)
			lb_start_block(p, label)
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, is.Label))
			lb_add_debug_label(p, is.Label, label)
		}
		tl := lb_push_target_list(p, is.Label, doneBlock, nil, nil)
		tl.IsBlock = true
	}
	if is.Init != nil {
		initBlock := lb_create_block(p, "if.init")
		lb_emit_jump(p, initBlock)
		lb_start_block(p, initBlock)
		lb_build_stmt(p, is.Init)
	}
	cond := lb_build_cond(p, is.Cond, thenBlock, elseBlock)
	if cond.Value != 0 && LLVMIsAConstantInt(cond.Value) != 0 {
		constCond := LLVMConstIntGetZExtValue(cond.Value) != 0
		ifInstr := LLVMGetLastInstruction(p.CurrBlock.Block)
		LLVMInstructionEraseFromParent(ifInstr)
		if constCond {
			lb_emit_jump(p, thenBlock)
			lb_start_block(p, thenBlock)
			lb_build_stmt(p, is.Body)
			lb_emit_jump(p, doneBlock)
		} else {
			if is.ElseStmt != nil {
				lb_emit_jump(p, elseBlock)
				lb_start_block(p, elseBlock)
				lb_open_scope(p, scope_of_node(is.ElseStmt))
				lb_build_stmt(p, is.ElseStmt)
				lb_close_scope(p, lbDeferExit_Default, nil, is.ElseStmt)
			}
			lb_emit_jump(p, doneBlock)
		}
	} else {
		lb_start_block(p, thenBlock)
		lb_build_stmt(p, is.Body)
		lb_emit_jump(p, doneBlock)
		if is.ElseStmt != nil {
			lb_start_block(p, elseBlock)
			lb_open_scope(p, scope_of_node(is.ElseStmt))
			lb_build_stmt(p, is.ElseStmt)
			lb_close_scope(p, lbDeferExit_Default, nil, is.ElseStmt)
			lb_emit_jump(p, doneBlock)
		}
	}
	if is.Label != nil {
		lb_pop_target_list(p)
	}
	lb_start_block(p, doneBlock)
}

func lb_build_for_stmt(p *lbProcedure, node *Ast) {
	fs := &node.ForStmt
	lb_open_scope(p, fs.Scope)
	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, node))
	}
	bodyBlock := lb_create_block(p, "for.body")
	doneBlock := lb_create_block(p, "for.done")
	loopBlock := bodyBlock
	if fs.Cond != nil {
		loopBlock = lb_create_block(p, "for.loop")
	}
	postBlock := loopBlock
	if fs.Post != nil {
		postBlock = lb_create_block(p, "for.post")
	}
	lb_push_target_list(p, fs.Label, doneBlock, postBlock, nil)
	if fs.Label != nil && p.DebugInfo != 0 {
		label := lb_create_block(p, "for.label")
		lb_emit_jump(p, label)
		lb_start_block(p, label)
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, fs.Label))
		lb_add_debug_label(p, fs.Label, label)
	}
	if fs.Init != nil {
		initBlock := lb_create_block(p, "for.init")
		lb_emit_jump(p, initBlock)
		lb_start_block(p, initBlock)
		lb_build_stmt(p, fs.Init)
	}
	lb_emit_jump(p, loopBlock)
	lb_start_block(p, loopBlock)
	if loopBlock != bodyBlock {
		if p.DebugInfo != 0 {
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, fs.Cond))
		}
		lb_build_cond(p, fs.Cond, bodyBlock, doneBlock)
		lb_start_block(p, bodyBlock)
	}
	lb_build_stmt(p, fs.Body)
	lb_pop_target_list(p)
	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, fs.Body))
	}
	lb_emit_jump(p, postBlock)
	if fs.Post != nil {
		lb_start_block(p, postBlock)
		lb_build_stmt(p, fs.Post)
		lb_emit_jump(p, loopBlock)
	}
	lb_start_block(p, doneBlock)
	lb_close_scope(p, lbDeferExit_Default, nil, node)
}

func lb_build_defer_stmt(p *lbProcedure, d lbDefer) {
	if p.CurrBlock == nil {
		return
	}
	lastInstr := LLVMGetLastInstruction(p.CurrBlock.Block)
	if lastInstr != 0 && LLVMIsAReturnInst(lastInstr) != 0 {
		return
	}
	prevContextStackCount := isize(len(p.ContextStack))
	defer func() {
		p.ContextStack = p.ContextStack[:prevContextStackCount]
	}()
	p.ContextStack = p.ContextStack[:d.ContextStackCount]
	b := lb_create_block(p, "defer")
	if lastInstr == 0 || LLVMIsATerminatorInst(lastInstr) == 0 {
		lb_emit_jump(p, b)
	}
	lb_start_block(p, b)
	if d.Kind == lbDefer_Node {
		lb_build_stmt(p, d.Stmt)
	} else if d.Kind == lbDefer_Proc {
		if p.DebugInfo != 0 && d.Pos.Line > 0 {
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_token_pos(p, d.Pos))
		}
		lb_emit_call(p, d.ProcDeferred, d.ProcResultAsArgs)
	}
}

func lb_emit_defer_stmts(p *lbProcedure, kind lbDeferExitKind, block *lbBlock, pos_or_node any) {
	prevTokenPos := p.BranchLocationPos
	if p.UsesBranchLocation {
		switch v := pos_or_node.(type) {
		case TokenPos:
			p.BranchLocationPos = v
		case *Ast:
			if v != nil {
				if v.Kind == Ast_BlockStmt || v.Kind == Ast_CaseClause {
					p.BranchLocationPos = ast_end_token(v).Pos
				} else {
					p.BranchLocationPos = ast_token(v).Pos
				}
			}
		}
	}
	defer func() {
		p.BranchLocationPos = prevTokenPos
	}()
	if kind == lbDeferExit_Return {
		for i := 0; i < len(p.AsanStackLocals); i++ {
			local := p.AsanStackLocals[i]
			args := []lbValue{
				lb_emit_conv(p, local, t_rawptr),
				lb_const_int(p.Module, t_int, uint64(type_size_of(local.Type.Pointer.Elem))),
			}
			lb_emit_runtime_call(p, "__asan_unpoison_memory_region", args)
		}
	}
	count := len(p.DeferStmts)
	for i := count - 1; i >= 0; i-- {
		d := p.DeferStmts[i]
		if kind == lbDeferExit_Default {
			if isize(p.ScopeIndex) == d.ScopeIndex && d.ScopeIndex > 0 {
				lb_build_defer_stmt(p, d)
				p.DeferStmts = append(p.DeferStmts[:i], p.DeferStmts[i+1:]...)
				continue
			} else {
				break
			}
		} else if kind == lbDeferExit_Return {
			lb_build_defer_stmt(p, d)
		} else if kind == lbDeferExit_Branch {
			lowerLimit := block.ScopeIndex
			if lowerLimit < d.ScopeIndex {
				lb_build_defer_stmt(p, d)
			}
		}
	}
}

func lb_add_defer_node(p *lbProcedure, scopeIndex isize, stmt *Ast) {
	pt := base_type(p.Type)
	if pt.Proc.CallingConvention == ProcCC_Odin {
		if len(p.ContextStack) == 0 {
			gb_assert_handler("Assertion Failure", "p.context_stack.count != 0", "src/llvm_backend_stmt.cpp", 3505, 0)
		}
	}
	d := lbDefer{
		Kind:              lbDefer_Node,
		ScopeIndex:        scopeIndex,
		ContextStackCount: isize(len(p.ContextStack)),
		Block:             p.CurrBlock,
		Pos:               ast_token(stmt).Pos,
		Stmt:              stmt,
	}
	p.DeferStmts = append(p.DeferStmts, d)
}

func lb_add_defer_proc(p *lbProcedure, scopeIndex isize, deferred lbValue, resultAsArgs []lbValue, pos TokenPos) {
	pt := base_type(p.Type)
	if pt.Proc.CallingConvention == ProcCC_Odin {
		if len(p.ContextStack) == 0 {
			gb_assert_handler("Assertion Failure", "p.context_stack.count != 0", "src/llvm_backend_stmt.cpp", 3521, 0)
		}
	}
	d := lbDefer{
		Kind:              lbDefer_Proc,
		ScopeIndex:        scopeIndex,
		ContextStackCount: isize(len(p.ContextStack)),
		Block:             p.CurrBlock,
		Pos:               pos,
		ProcDeferred:      deferred,
		ProcResultAsArgs:  resultAsArgs,
	}
	p.DeferStmts = append(p.DeferStmts, d)
}

func lb_map_cell_index_static(p *lbProcedure, typ *Type, cellsPtr lbValue, index lbValue) lbValue {
	size, len_ := i64(0), i64(0)
	elemSz := type_size_of(typ)
	map_cell_size_and_len(typ, &size, &len_)
	index = lb_emit_conv(p, index, t_uintptr)
	if size == len_*elemSz {
		elemsPtr := lb_emit_conv(p, cellsPtr, alloc_type_pointer(typ))
		return lb_emit_ptr_offset(p, elemsPtr, index)
	}
	sizeConst := lb_const_int(p.Module, t_uintptr, uint64(size))
	lenConst := lb_const_int(p.Module, t_uintptr, uint64(len_))
	var cellIndex, dataIndex lbValue
	if is_power_of_two(uint64(len_)) {
		log2Len := floor_log2(uint64(len_))
		if log2Len == 0 {
			cellIndex = index
		} else {
			cellIndex = lb_emit_arith(p, Token_Shr, index, lb_const_int(p.Module, t_uintptr, log2Len), t_uintptr)
		}
		dataIndex = lb_emit_arith(p, Token_And, index, lb_const_int(p.Module, t_uintptr, uint64(len_-1)), t_uintptr)
	} else {
		cellIndex = lb_emit_arith(p, Token_Quo, index, lenConst, t_uintptr)
		dataIndex = lb_emit_arith(p, Token_Mod, index, lenConst, t_uintptr)
	}
	elemsPtr := lb_emit_conv(p, cellsPtr, t_uintptr)
	cellOffset := lb_emit_arith(p, Token_Mul, sizeConst, cellIndex, t_uintptr)
	elemsPtr = lb_emit_arith(p, Token_Add, elemsPtr, cellOffset, t_uintptr)
	elemsPtr = lb_emit_conv(p, elemsPtr, alloc_type_pointer(typ))
	return lb_emit_ptr_offset(p, elemsPtr, dataIndex)
}

func lb_map_hash_is_valid(p *lbProcedure, hash lbValue) lbValue {
	topBitIndex := uint64(type_size_of(t_uintptr)*8 - 1)
	shiftAmount := lb_const_int(p.Module, t_uintptr, topBitIndex)
	zero := lb_const_int(p.Module, t_uintptr, 0)
	notEmpty := lb_emit_comp(p, Token_NotEq, hash, zero)
	notDeleted := lb_emit_arith(p, Token_Shr, hash, shiftAmount, t_uintptr)
	notDeleted = lb_emit_comp(p, Token_CmpEq, notDeleted, zero)
	return lb_emit_arith(p, Token_And, notDeleted, notEmpty, t_bool)
}
