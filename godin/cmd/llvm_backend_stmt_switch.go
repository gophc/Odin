package cmd

func lb_switch_stmt_can_be_trivial_jump_table(ss *Ast, defaultFound *bool) bool {
	if ss.SwitchStmt.Tag == nil {
		return false
	}
	isTypeid := false
	tv := type_and_value_of_expr(ss.SwitchStmt.Tag)
	if is_type_integer(core_type(tv.Type)) {
	} else if is_type_typeid(tv.Type) {
		isTypeid = true
	} else {
		return false
	}
	body := ss.SwitchStmt.Body
	if body.Kind == AstBlockStmt {
		_ = body.BlockStmt
	} else {
		gb_assert_handler("Assertion Failure", "(ss->body)->kind == Ast_BlockStmt", "src/llvm_backend_stmt.cpp", 1837, "expected BlockStmt got %s", ast_strings[body.Kind].text)
	}
	for _, clause := range body.BlockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 1839, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		if len(cc.List) == 0 {
			if defaultFound != nil {
				*defaultFound = true
			}
			continue
		}
		for _, expr := range cc.List {
			expr = unparen_expr(expr)
			if is_ast_range(expr) {
				return false
			}
			if expr.TAV.Mode == Addressing_Type {
				if !isTypeid {
					gb_assert_handler("Assertion Failure", "is_typeid", "src/llvm_backend_stmt.cpp", 1852, 0)
				}
				continue
			}
			tv = type_and_value_of_expr(expr)
			if tv.Mode != Addressing_Constant {
				return false
			}
			if !is_type_integer(core_type(tv.Type)) {
				return false
			}
		}
	}
	if isTypeid {
		return false
	}
	return true
}

func lb_build_switch_stmt(p *lbProcedure, ss *Ast, scope *Scope) {
	lb_open_scope(p, scope)
	if ss.SwitchStmt.Label != nil && p.DebugInfo != nil {
		label := lb_create_block(p, "switch.label")
		lb_emit_jump(p, label)
		lb_start_block(p, label)
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, ss.SwitchStmt.Label))
		lb_add_debug_label(p, ss.SwitchStmt.Label, label)
	}
	if ss.SwitchStmt.Init != nil {
		lb_build_stmt(p, ss.SwitchStmt.Init)
	}
	tag := lb_const_bool(p.Module, t_llvm_bool, true)
	if ss.SwitchStmt.Tag != nil {
		tag = lb_build_expr(p, ss.SwitchStmt.Tag)
	}
	done := lb_create_block(p, "switch.done")
	bodyNode := ss.SwitchStmt.Body
	if bodyNode.Kind == AstBlockStmt {
		_ = bodyNode.BlockStmt
	} else {
		gb_assert_handler("Assertion Failure", "(ss->body)->kind == Ast_BlockStmt", "src/llvm_backend_stmt.cpp", 1894, "expected BlockStmt got %s", ast_strings[bodyNode.Kind].text)
	}
	blockStmt := &bodyNode.BlockStmt
	caseCount := len(blockStmt.Stmts)
	var defaultClause *Ast
	var defaultStmts []*Ast
	var defaultFall *lbBlock
	var defaultBlock *lbBlock
	var fall *lbBlock
	defaultFound := false
	isTrivial := lb_switch_stmt_can_be_trivial_jump_table(ss, &defaultFound)
	bodyBlocks := make([]*lbBlock, len(blockStmt.Stmts))
	for i, clause := range blockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 1909, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		blockName := "switch.default.body"
		if len(cc.List) >= 1 {
			blockName = "switch.case.body"
		}
		if isTrivial && len(cc.List) >= 1 {
			bn := gb_string_make(heap_allocator(), "switch.case.")
			first := cc.List[0]
			if first.TAV.Mode == Addressing_Type {
				bn = gb_string_appendc(bn, "type.")
			} else if is_type_rune(first.TAV.Type) {
				bn = gb_string_appendc(bn, "rune.")
			} else {
				bn = gb_string_appendc(bn, "value.")
			}
			for j, expr := range cc.List {
				if j > 0 {
					bn = gb_string_appendc(bn, "..")
				}
				if expr.TAV.Mode == Addressing_Type {
					bn = write_type_to_string(bn, expr.TAV.Type, false)
				} else {
					value := expr.TAV.Value
					if is_type_rune(expr.TAV.Type) && value.Kind == ExactValue_Integer {
						r := Rune(exact_value_to_i64(value))
						var runeTemp [6]u8
						size := gb_utf8_encode_rune(runeTemp[:], r)
						bn = gb_string_append_length(bn, runeTemp[:], size)
					} else {
						bn = write_exact_value_to_string(bn, value, 1024)
					}
				}
			}
			blockName = string(bn)
		}
		bodyBlocks[i] = lb_create_block(p, blockName)
		if len(cc.List) == 0 {
			defaultBlock = bodyBlocks[i]
		}
	}
	var switchInstr LLVMValueRef
	if isTrivial {
		numCases := 0
		for _, clause := range blockStmt.Stmts {
			if clause.Kind == AstCaseClause {
				_ = clause.CaseClause
			} else {
				gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 1960, "expected CaseClause got %s", ast_strings[clause.Kind].text)
			}
			cc := &clause.CaseClause
			numCases += len(cc.List)
		}
		endBlock := done.Block
		if defaultBlock != nil {
			endBlock = defaultBlock.Block
		}
		switchInstr = LLVMBuildSwitch(p.Builder, tag.Value, endBlock, uint(numCases))
	}
	for i, clause := range blockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 1975, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		bodyBlock := bodyBlocks[i]
		fall = done
		if i+1 < caseCount {
			fall = bodyBlocks[i+1]
		}
		if len(cc.List) == 0 {
			defaultClause = clause
			defaultStmts = cc.Stmts
			defaultFall = fall
			if switchInstr == 0 {
				defaultBlock = bodyBlock
			} else {
				if defaultBlock == nil {
					gb_assert_handler("Assertion Failure", "default_block != nil", "src/llvm_backend_stmt.cpp", 1991, 0)
				}
			}
			continue
		}
		var nextCond *lbBlock
		if p.DebugInfo != nil {
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, clause))
		}
		for _, expr := range cc.List {
			expr = unparen_expr(expr)
			if switchInstr != 0 {
				var onVal lbValue
				if expr.TAV.Mode == Addressing_Type {
					if !is_type_typeid(tag.Type) {
						gb_assert_handler("Assertion Failure", "is_type_typeid(tag.type)", "src/llvm_backend_stmt.cpp", 2007, 0)
					}
					e := lb_typeid(p.Module, expr.TAV.Type)
					onVal = lb_emit_conv(p, e, tag.Type)
				} else {
					if expr.TAV.Mode != Addressing_Constant {
						gb_assert_handler("Assertion Failure", "expr->tav.mode == Addressing_Constant", "src/llvm_backend_stmt.cpp", 2011, 0)
					}
					if is_ast_range(expr) {
						gb_assert_handler("Assertion Failure", "!is_ast_range(expr)", "src/llvm_backend_stmt.cpp", 2012, 0)
					}
					onVal = lb_build_expr(p, expr)
					onVal = lb_emit_conv(p, onVal, tag.Type)
				}
				if LLVMIsConstant(onVal.Value) == 0 {
					gb_assert_handler("Assertion Failure", "LLVMIsConstant(on_val.value)", "src/llvm_backend_stmt.cpp", 2018, 0)
				}
				LLVMAddCase(switchInstr, onVal.Value, bodyBlock.Block)
				continue
			}
			nextCond = lb_create_block(p, "switch.case.next")
			var cond lbValue
			if is_ast_range(expr) {
				if expr.Kind == AstBinaryExpr {
					_ = expr.BinaryExpr
				} else {
					gb_assert_handler("Assertion Failure", "(expr)->kind == Ast_BinaryExpr", "src/llvm_backend_stmt.cpp", 2027, "expected BinaryExpr got %s", ast_strings[expr.Kind].text)
				}
				ie := &expr.BinaryExpr
				op := Token_Invalid
				switch ie.Op.Kind {
				case Token_Ellipsis:
					op = Token_LtEq
				case Token_RangeFull:
					op = Token_LtEq
				case Token_RangeHalf:
					op = Token_Lt
				default:
					gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 2033, "Invalid interval operator")
				}
				lhs := lb_build_expr(p, ie.Left)
				rhs := lb_build_expr(p, ie.Right)
				condLhs := lb_emit_comp(p, Token_LtEq, lhs, tag)
				condRhs := lb_emit_comp(p, op, tag, rhs)
				cond = lb_emit_arith(p, Token_And, condLhs, condRhs, t_bool)
			} else {
				if expr.TAV.Mode == Addressing_Type {
					if !is_type_typeid(tag.Type) {
						gb_assert_handler("Assertion Failure", "is_type_typeid(tag.type)", "src/llvm_backend_stmt.cpp", 2043, 0)
					}
					e := lb_typeid(p.Module, expr.TAV.Type)
					e = lb_emit_conv(p, e, tag.Type)
					cond = lb_emit_comp(p, Token_CmpEq, tag, e)
				} else {
					cond = lb_emit_comp(p, Token_CmpEq, tag, lb_build_expr(p, expr))
				}
			}
			lb_emit_if(p, cond, bodyBlock, nextCond)
			lb_start_block(p, nextCond)
		}
		lb_start_block(p, bodyBlock)
		lb_push_target_list(p, ss.SwitchStmt.Label, done, nil, fall)
		lb_open_scope(p, bodyBlock.Scope)
		lb_build_stmt_list(p, cc.Stmts)
		lb_close_scope(p, lbDeferExit_Default, bodyBlock, clause)
		lb_pop_target_list(p)
		lb_emit_jump(p, done)
		if switchInstr == 0 {
			lb_start_block(p, nextCond)
		}
	}
	if defaultBlock != nil {
		if switchInstr == 0 {
			lb_emit_jump(p, defaultBlock)
		}
		lb_start_block(p, defaultBlock)
		lb_push_target_list(p, ss.SwitchStmt.Label, done, nil, defaultFall)
		lb_open_scope(p, defaultBlock.Scope)
		lb_build_stmt_list(p, defaultStmts)
		lb_close_scope(p, lbDeferExit_Default, defaultBlock, defaultClause)
		lb_pop_target_list(p)
	}
	lb_emit_jump(p, done)
	lb_start_block(p, done)
	lb_close_scope(p, lbDeferExit_Default, done, ss.SwitchStmt.Body)
}

func lb_store_type_case_implicit(p *lbProcedure, clause *Ast, value lbValue, is_default_case bool) {
	e := implicit_entity_of_node(clause)
	if e == nil {
		gb_assert_handler("Assertion Failure", "e != nil", "src/llvm_backend_stmt.cpp", 2089, 0)
	}
	if e.Flags&EntityFlag_Value != 0 {
		if are_types_identical(e.Type, value.Type) {
			x := lb_add_local(p, e.Type, e, false, false)
			lb_addr_store(p, x, value)
		} else {
			if !are_types_identical(e.Type, type_deref(value.Type)) {
				gb_assert_handler("Assertion Failure", "are_types_identical(e->type, type_deref(value.type))", "src/llvm_backend_stmt.cpp", 2096, "%s", type_to_string(value.Type))
			}
			x := lb_add_local(p, e.Type, e, false, false)
			lb_addr_store(p, x, lb_emit_load(p, value))
		}
	} else {
		if !is_default_case {
			clauseType := e.Type
			if !are_types_identical(type_deref(clauseType), type_deref(value.Type)) {
				gb_assert_handler("Assertion Failure", "are_types_identical(type_deref(clause_type), type_deref(value.type))", "src/llvm_backend_stmt.cpp", 2103, "%s %s", type_to_string(clauseType), type_to_string(value.Type))
			}
		}
		lb_add_entity(p.Module, e, value)
	}
}

func lb_store_range_stmt_val(p *lbProcedure, stmtVal *Ast, value lbValue) lbAddr {
	e := entity_of_node(stmtVal)
	if e == nil {
		return lbAddr{}
	}
	if e.Flags&EntityFlag_Value == 0 {
		if LLVMIsALoadInst(value.Value) != 0 {
			ptr := lb_address_from_load_or_generate_local(p, value)
			lb_add_entity(p.Module, e, ptr)
			lb_add_debug_local_variable(p, ptr.Value, e.Type, e.Token)
			return lb_addr(ptr)
		}
	}
	addr := lb_add_local(p, e.Type, e, false, false)
	lb_addr_store(p, addr, value)
	return addr
}

func lb_type_case_body(p *lbProcedure, label *Ast, clause *Ast, body *lbBlock, done *lbBlock) {
	if clause.Kind == AstCaseClause {
		_ = clause.CaseClause
	} else {
		gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 2131, "expected CaseClause got %s", ast_strings[clause.Kind].text)
	}
	cc := &clause.CaseClause
	lb_push_target_list(p, label, done, nil, nil)
	lb_build_stmt_list(p, cc.Stmts)
	lb_close_scope(p, lbDeferExit_Default, body, clause)
	lb_pop_target_list(p)
	lb_emit_jump(p, done)
}

func lb_build_type_switch_stmt(p *lbProcedure, ss *Ast) {
	m := p.Module
	lb_open_scope(p, ss.TypeSwitchStmt.scope)
	if ss.TypeSwitchStmt.Tag.Kind == AstAssignStmt {
		_ = ss.TypeSwitchStmt.Tag.AssignStmt
	} else {
		gb_assert_handler("Assertion Failure", "(ss->tag)->kind == Ast_AssignStmt", "src/llvm_backend_stmt.cpp", 2146, "expected AssignStmt got %s", ast_strings[ss.TypeSwitchStmt.Tag.Kind].text)
	}
	as := &ss.TypeSwitchStmt.Tag.AssignStmt
	if len(as.LHS) != 1 {
		gb_assert_handler("Assertion Failure", "as->lhs.count == 1", "src/llvm_backend_stmt.cpp", 2147, 0)
	}
	if len(as.RHS) != 1 {
		gb_assert_handler("Assertion Failure", "as->rhs.count == 1", "src/llvm_backend_stmt.cpp", 2148, 0)
	}
	parent := lb_build_expr(p, as.RHS[0])
	isParentPtr := is_type_pointer(parent.Type)
	parentBaseType := type_deref(parent.Type)
	switchKind := check_valid_type_switch_type(parent.Type)
	if switchKind == TypeSwitchInvalid {
		gb_assert_handler("Assertion Failure", "switch_kind != TypeSwitch_Invalid", "src/llvm_backend_stmt.cpp", 2155, 0)
	}
	parentValue := parent
	parentPtr := parent
	if !isParentPtr {
		parentPtr = lb_address_from_load_or_generate_local(p, parent)
	}
	var tag lbValue
	var unionData lbValue
	if switchKind == TypeSwitchUnion {
		unionData = lb_emit_conv(p, parentPtr, t_rawptr)
		unionType := type_deref(parentPtr.Type)
		if is_type_union_maybe_pointer(unionType) {
			tag = lb_emit_conv(p, lb_emit_comp_against_nil(p, Token_NotEq, parentValue), t_int)
		} else if union_tag_size(unionType) == 0 {
			tag = lbValue{}
		} else {
			tagPtr := lb_emit_union_tag_ptr(p, parentPtr)
			tag = lb_emit_load(p, tagPtr)
		}
	} else if switchKind == TypeSwitchAny {
		tag = lb_emit_load(p, lb_emit_struct_ep(p, parentPtr, 1))
	} else {
		gb_assert_handler("Panic", nil, "src/llvm_backend_stmt.cpp", 2180, "Unknown switch kind")
	}
	bodyNode := ss.TypeSwitchStmt.Body
	if bodyNode.Kind == AstBlockStmt {
		_ = bodyNode.BlockStmt
	} else {
		gb_assert_handler("Assertion Failure", "(ss->body)->kind == Ast_BlockStmt", "src/llvm_backend_stmt.cpp", 2183, "expected BlockStmt got %s", ast_strings[bodyNode.Kind].text)
	}
	blockStmt := &bodyNode.BlockStmt
	done := lb_create_block(p, "typeswitch.done")
	elseBlock := done
	var defaultBlock *lbBlock
	numCases := 0
	for _, clause := range blockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 2191, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		numCases += len(cc.List)
		if len(cc.List) == 0 {
			if defaultBlock != nil {
				gb_assert_handler("Assertion Failure", "default_block == nil", "src/llvm_backend_stmt.cpp", 2194, 0)
			}
			defaultBlock = lb_create_block(p, "typeswitch.case.default")
			elseBlock = defaultBlock
		}
	}
	var switchInstr LLVMValueRef
	if type_size_of(parentBaseType) == 0 {
		if tag.Value != 0 {
			gb_assert_handler("Assertion Failure", "tag.value == nil", "src/llvm_backend_stmt.cpp", 2203, 0)
		}
		switchInstr = LLVMBuildSwitch(p.Builder, lb_const_bool(p.Module, t_llvm_bool, false).Value, elseBlock.Block, uint(numCases))
	} else {
		if tag.Value == 0 {
			gb_assert_handler("Assertion Failure", "tag.value != nil", "src/llvm_backend_stmt.cpp", 2206, 0)
		}
		switchInstr = LLVMBuildSwitch(p.Builder, tag.Value, elseBlock.Block, uint(numCases))
	}
	allByReference := false
	for _, clause := range blockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 2212, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		if len(cc.List) != 1 {
			continue
		}
		caseEntity := implicit_entity_of_node(clause)
		allByReference = allByReference || (caseEntity.Flags&EntityFlag_Value == 0)
		break
	}
	backingData := lbAddr{}
	if !allByReference {
		variantsFound := false
		maxSize := i64(0)
		maxAlign := i64(1)
		for _, clause := range blockStmt.Stmts {
			if clause.Kind == AstCaseClause {
				_ = clause.CaseClause
			} else {
				gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 2239, "expected CaseClause got %s", ast_strings[clause.Kind].text)
			}
			cc := &clause.CaseClause
			if len(cc.List) != 1 {
				continue
			}
			caseEntity := implicit_entity_of_node(clause)
			if !is_type_untyped_nil(caseEntity.Type) {
				if type_size_of(caseEntity.Type) > maxSize {
					maxSize = type_size_of(caseEntity.Type)
				}
				if type_align_of(caseEntity.Type) > maxAlign {
					maxAlign = type_align_of(caseEntity.Type)
				}
				variantsFound = true
			}
		}
		if variantsFound {
			t := alloc_type_array(t_u8, maxSize, nil)
			backingData = lb_add_local(p, t, nil, false, true)
			if !lb_try_update_alignment(backingData.Addr, uint(maxAlign)) {
				gb_assert_handler("Assertion Failure", "lb_try_update_alignment(backing_data.addr, (unsigned)max_align)", "src/llvm_backend_stmt.cpp", 2253, 0)
			}
		}
	}
	backingPtr := backingData.Addr
	for _, clause := range blockStmt.Stmts {
		if clause.Kind == AstCaseClause {
			_ = clause.CaseClause
		} else {
			gb_assert_handler("Assertion Failure", "(clause)->kind == Ast_CaseClause", "src/llvm_backend_stmt.cpp", 2259, "expected CaseClause got %s", ast_strings[clause.Kind].text)
		}
		cc := &clause.CaseClause
		caseEntity := implicit_entity_of_node(clause)
		lb_open_scope(p, cc.scope)
		if len(cc.List) == 0 {
			lb_start_block(p, defaultBlock)
			if caseEntity.Flags&EntityFlag_Value != 0 {
				lb_store_type_case_implicit(p, clause, parentValue, true)
			} else {
				lb_store_type_case_implicit(p, clause, parentPtr, true)
			}
			lb_type_case_body(p, ss.TypeSwitchStmt.Label, clause, p.CurrBlock, done)
			continue
		}
		bodyName := "typeswitch.case"
		if !are_types_identical(caseEntity.Type, parentBaseType) {
			canonicalName := temp_canonical_string(caseEntity.Type)
			bn := gb_string_make(heap_allocator(), "typeswitch.case.")
			bn = gb_string_append_length(bn, canonicalName, gb_string_length(canonicalName))
			bodyName = string(bn)
		}
		bodyBlock := lb_create_block(p, bodyName)
		if p.DebugInfo != nil {
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, clause))
		}
		sawNil := false
		for _, typeExpr := range cc.List {
			caseType := type_of_expr(typeExpr)
			var onVal lbValue
			if switchKind == TypeSwitchUnion {
				ut := base_type(type_deref(parent.Type))
				if is_type_untyped_nil(caseType) {
					if !type_has_nil(ut) {
						gb_assert_handler("Assertion Failure", "type_has_nil(ut)", "src/llvm_backend_stmt.cpp", 2296, 0)
					}
					sawNil = true
					onVal = lb_const_int(m, union_tag_type(ut), 0)
				} else {
					onVal = lb_const_union_tag(m, ut, caseType)
				}
			} else if switchKind == TypeSwitchAny {
				if is_type_untyped_nil(caseType) {
					sawNil = true
					onVal = lb_const_nil(m, t_typeid)
				} else {
					onVal = lb_typeid(m, caseType)
				}
			}
			if onVal.Value == 0 {
				gb_assert_handler("Assertion Failure", "on_val.value != nil", "src/llvm_backend_stmt.cpp", 2311, 0)
			}
			LLVMAddCase(switchInstr, onVal.Value, bodyBlock.Block)
		}
		lb_start_block(p, bodyBlock)
		byReference := (caseEntity.Flags & EntityFlag_Value) == 0
		if len(cc.List) == 1 && !sawNil {
			var data lbValue
			if switchKind == TypeSwitchUnion {
				data = unionData
			} else if switchKind == TypeSwitchAny {
				data = lb_emit_load(p, lb_emit_struct_ep(p, parentPtr, 0))
			}
			if !is_type_pointer(data.Type) {
				gb_assert_handler("Assertion Failure", "is_type_pointer(data.type)", "src/llvm_backend_stmt.cpp", 2327, 0)
			}
			ct := caseEntity.Type
			ctPtr := alloc_type_pointer(ct)
			var ptr lbValue
			if backingData.Addr.Value != 0 {
				if byReference {
					gb_assert_handler("Assertion Failure", "!by_reference", "src/llvm_backend_stmt.cpp", 2335, 0)
				}
				lb_mem_copy_non_overlapping(p,
					backingPtr,
					data,
					lb_const_int(p.Module, t_int, type_size_of(caseEntity.Type)))
				ptr = lb_emit_conv(p, backingPtr, ctPtr)
			} else {
				if !byReference {
					gb_assert_handler("Assertion Failure", "by_reference", "src/llvm_backend_stmt.cpp", 2344, 0)
				}
				ptr = lb_emit_conv(p, data, ctPtr)
			}
			if !are_types_identical(caseEntity.Type, type_deref(ptr.Type)) {
				gb_assert_handler("Assertion Failure", "are_types_identical(case_entity->type, type_deref(ptr.type))", "src/llvm_backend_stmt.cpp", 2347, 0)
			}
			lb_add_entity(p.Module, caseEntity, ptr)
			lb_add_debug_local_variable(p, ptr.Value, caseEntity.Type, caseEntity.Token)
		} else {
			if byReference {
				lb_store_type_case_implicit(p, clause, parentPtr, false)
			} else {
				lb_store_type_case_implicit(p, clause, parentValue, false)
			}
		}
		lb_type_case_body(p, ss.TypeSwitchStmt.Label, clause, bodyBlock, done)
	}
	lb_emit_jump(p, done)
	lb_start_block(p, done)
	lb_close_scope(p, lbDeferExit_Default, done, ss.TypeSwitchStmt.Body)
}
