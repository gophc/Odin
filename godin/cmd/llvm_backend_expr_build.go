package cmd

func lb_build_binary_in(p *lbProcedure, left lbValue, right lbValue, op TokenKind) lbValue {
	rt := base_type(right.Type)
	if is_type_pointer(rt) {
		right = lb_emit_load(p, right)
		rt = base_type(type_deref(rt))
	}
	switch rt.Kind {
	case TypeMap:
		map_ptr := lb_address_from_load_or_generate_local(p, right)
		key := left
		ptr := lb_internal_dynamic_map_get_ptr(p, map_ptr, key)
		if op == Token_in {
			return lb_emit_conv(p, lb_emit_comp_against_nil(p, Token_NotEq, ptr), t_bool)
		} else {
			return lb_emit_conv(p, lb_emit_comp_against_nil(p, Token_CmpEq, ptr), t_bool)
		}

	case TypeBitSet:
		key_type := rt.BitSet.Elem
		_ = key_type
		it := bit_set_to_int(rt)
		left = lb_emit_conv(p, left, it)
		if is_type_different_to_arch_endianness(it) {
			left = lb_emit_byte_swap(p, left, integer_endian_type_to_platform_type(it))
		}
		lower := lb_const_value(p.Module, left.Type, exact_value_i64(rt.BitSet.Lower))
		key := lb_emit_arith(p, Token_Sub, left, lower, left.Type)
		bit := lb_emit_arith(p, Token_Shl, lb_const_int(p.Module, left.Type, 1), key, left.Type)
		bit = lb_emit_conv(p, bit, it)
		old_value := lb_emit_transmute(p, right, it)
		new_value := lb_emit_arith(p, Token_And, old_value, bit, it)
		if op == Token_in {
			return lb_emit_conv(p, lb_emit_comp(p, Token_NotEq, new_value, lb_const_int(p.Module, new_value.Type, 0)), t_bool)
		} else {
			return lb_emit_conv(p, lb_emit_comp(p, Token_CmpEq, new_value, lb_const_int(p.Module, new_value.Type, 0)), t_bool)
		}
	}

	gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 1966, "Invalid 'in' type")
	return lbValue{}
}

func lb_build_binary_expr(p *lbProcedure, expr *Ast) lbValue {
	be := &expr.BinaryExpr
	tv := type_and_value_of_expr(expr)
	if is_type_matrix(be.Left.Tav.Type) || is_type_matrix(be.Right.Tav.Type) {
		left := lb_build_expr(p, be.Left)
		right := lb_build_expr(p, be.Right)
		return lb_emit_arith_matrix(p, be.Op.Kind, left, right, default_type(tv.Type), false)
	}
	switch be.Op.Kind {
	case Token_Add, Token_Sub, Token_Mul, Token_Quo, Token_Mod, Token_ModMod, Token_And, Token_Or, Token_Xor, Token_AndNot:
		type_ := default_type(tv.Type)
		left := lb_build_expr(p, be.Left)
		right := lb_build_expr(p, be.Right)
		return lb_emit_arith(p, be.Op.Kind, left, right, type_)

	case Token_Shl, Token_Shr:
		type_ := default_type(tv.Type)
		left := lb_build_expr(p, be.Left)
		var right lbValue
		if lb_is_expr_untyped_const(be.Right) {
			right = lb_expr_untyped_const_to_typed(p.Module, be.Right, type_)
		} else {
			right = lb_build_expr(p, be.Right)
		}
		return lb_emit_arith(p, be.Op.Kind, left, right, type_)

	case Token_CmpEq, Token_NotEq:
		if is_type_untyped_nil(be.Right.Tav.Type) {
			left := lb_build_expr(p, be.Left)
			cmp := lb_emit_comp_against_nil(p, be.Op.Kind, left)
			type_ := default_type(tv.Type)
			return lb_emit_conv(p, cmp, type_)
		} else if is_type_untyped_nil(be.Left.Tav.Type) {
			right := lb_build_expr(p, be.Right)
			cmp := lb_emit_comp_against_nil(p, be.Op.Kind, right)
			type_ := default_type(tv.Type)
			return lb_emit_conv(p, cmp, type_)
		} else if lb_is_empty_string_constant(be.Right) && !is_type_union(be.Left.Tav.Type) {
			str_type := t_string
			if is_type_string16(be.Left.Tav.Type) || is_type_cstring16(be.Left.Tav.Type) {
				str_type = t_string16
			}
			s := lb_build_expr(p, be.Left)
			s = lb_emit_conv(p, s, str_type)
			len_ := lb_string_len(p, s)
			cmp := lb_emit_comp(p, be.Op.Kind, len_, lb_const_int(p.Module, t_int, 0))
			type_ := default_type(tv.Type)
			return lb_emit_conv(p, cmp, type_)
		} else if lb_is_empty_string_constant(be.Left) && !is_type_union(be.Right.Tav.Type) {
			str_type := t_string
			if is_type_string16(be.Right.Tav.Type) || is_type_cstring16(be.Right.Tav.Type) {
				str_type = t_string16
			}
			s := lb_build_expr(p, be.Right)
			s = lb_emit_conv(p, s, str_type)
			len_ := lb_string_len(p, s)
			cmp := lb_emit_comp(p, be.Op.Kind, len_, lb_const_int(p.Module, t_int, 0))
			type_ := default_type(tv.Type)
			return lb_emit_conv(p, cmp, type_)
		}
		fallthrough

	case Token_Lt, Token_LtEq, Token_Gt, Token_GtEq:
		var left lbValue
		var right lbValue
		if be.Left.Tav.Mode == Addressing_Type {
			left = lb_typeid(p.Module, be.Left.Tav.Type)
		}
		if be.Right.Tav.Mode == Addressing_Type {
			right = lb_typeid(p.Module, be.Right.Tav.Type)
		}
		if left.Value == 0 {
			left = lb_build_expr(p, be.Left)
		}
		if right.Value == 0 {
			right = lb_build_expr(p, be.Right)
		}
		cmp := lb_emit_comp(p, be.Op.Kind, left, right)
		type_ := default_type(tv.Type)
		return lb_emit_conv(p, cmp, type_)

	case Token_CmpAnd, Token_CmpOr:
		return lb_emit_logical_binary_expr(p, be.Op.Kind, be.Left, be.Right, tv.Type)

	case Token_in, Token_not_in:
		left := lb_build_expr(p, be.Left)
		right := lb_build_expr(p, be.Right)
		return lb_build_binary_in(p, left, right, be.Op.Kind)

	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 2088, "Invalid binary expression")
	}
	return lbValue{}
}

func lb_build_expr_internal(p *lbProcedure, expr *Ast) lbValue {
	m := p.Module

	expr = unparen_expr(expr)

	expr_pos := ast_token(expr).Pos
	tv := type_and_value_of_expr(expr)
	type_ := type_of_expr(expr)
	if tv.Mode == Addressing_Invalid {
		gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4304,
			"invalid expression '%s' (tv.mode = %d, tv.type = %s) @ %s\n Current Proc: %.*s : %s",
			expr_to_string(expr), int(tv.Mode), type_to_string(tv.Type), token_pos_to_string(expr_pos),
			len(p.Name), p.Name, type_to_string(p.Type))
		_ = expr_pos
	}

	if tv.Value.Kind != ExactValue_Invalid {
		original_type := lb_build_expr_original_const_type(expr)
		return lb_const_value(p.Module, type_, tv.Value, LB_CONST_CONTEXT_DEFAULT_ALLOW_LOCAL, original_type)
	} else if tv.Mode == Addressing_Type {
		return lb_typeid(m, tv.Type)
	}

	switch expr.Kind {
	case Ast_BasicLit:
		if type_ != nil && type_.Named.Name == "Error" {
			e := type_.Named.TypeName
			if e.Pkg != nil && e.Pkg.Name == "os" {
				return lb_const_nil(p.Module, type_)
			}
		}
		pos := expr.BasicLit.Token.Pos
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4326,
			"Non-constant basic literal %s - %.*s (%s)",
			token_pos_to_string(pos), len(token_strings[expr.BasicLit.Token.Kind]), token_strings[expr.BasicLit.Token.Kind], type_to_string(type_))

	case Ast_BasicDirective:
		pos := expr.BasicDirective.Token.Pos
		name := expr.BasicDirective.Name.String
		if name == "branch_location" {
			proc_name := p.Entity.Token.String
			return lb_emit_source_code_location_as_global(p, proc_name, p.BranchLocationPos)
		}
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4337,
			"Non-constant basic literal %s - %.*s", token_pos_to_string(pos), len(name), name)

	case Ast_Implicit:
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_Uninit:
		res := lbValue{}
		if is_type_untyped(type_) {
			res.Value = 0
			res.Type = t_untyped_uninit
		} else {
			res.Value = LLVMGetUndef(lb_type(m, type_))
			res.Type = type_
		}
		return res

	case Ast_Ident:
		e := entity_from_expr(expr)
		e = strip_entity_wrapping(e)
		if e == nil {
			gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4360,
				"%s in %.*s %p", expr_to_string(expr), len(p.Name), p.Name, expr)
		}
		if e.Kind == Entity_Builtin {
			token := ast_token(expr)
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4363,
				"TODO(bill): lb_build_expr Entity_Builtin '%.*s'\n\t at %s",
				len(builtin_procs[e.Builtin.Id].Name), builtin_procs[e.Builtin.Id].Name,
				token_pos_to_string(token.Pos))
			return lbValue{}
		} else if e.Kind == Entity_Nil {
			res := lbValue{}
			res.Value = 0
			res.Type = e.Type
			return res
		}
		return lb_find_ident(p, m, e, expr)

	case Ast_DerefExpr:
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_SelectorExpr:
		tav := type_and_value_of_expr(expr)
		_ = tav
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_ImplicitSelectorExpr:
		tav := type_and_value_of_expr(expr)
		_ = tav
		return lb_const_value(p.Module, type_, tv.Value, LB_CONST_CONTEXT_DEFAULT_ALLOW_LOCAL, tv.Type)

	case Ast_SelectorCallExpr:
		return lb_build_call_expr(p, expr.SelectorCallExpr.Call)

	case Ast_TernaryIfExpr:
		te := &expr.TernaryIfExpr
		type_ := default_type(type_of_expr(expr))
		if lb_is_expr_trivial(te.X) && lb_is_expr_trivial(te.Y) {
			cond := lb_build_expr(p, te.Cond)
			x := lb_emit_conv(p, lb_build_expr(p, te.X), type_)
			y := lb_emit_conv(p, lb_build_expr(p, te.Y), type_)
			return lb_emit_select(p, cond, x, y)
		}

		var incoming_values [2]LLVMValueRef
		var incoming_blocks [2]LLVMBasicBlockRef

		then_ := lb_create_block(p, "if.then")
		done := lb_create_block(p, "if.done")
		else_ := lb_create_block(p, "if.else")

		lb_build_cond(p, te.Cond, then_, else_)
		lb_start_block(p, then_)

		llvm_type := lb_type(p.Module, type_)

		incoming_values[0] = lb_emit_conv(p, lb_build_expr(p, te.X), type_).Value
		if is_type_internally_pointer_like(type_) {
			incoming_values[0] = LLVMBuildBitCast(p.Builder, incoming_values[0], llvm_type, "")
		}

		lb_emit_jump(p, done)
		lb_start_block(p, else_)

		incoming_values[1] = lb_emit_conv(p, lb_build_expr(p, te.Y), type_).Value
		if is_type_internally_pointer_like(type_) {
			incoming_values[1] = LLVMBuildBitCast(p.Builder, incoming_values[1], llvm_type, "")
		}

		lb_emit_jump(p, done)
		lb_start_block(p, done)

		res := lbValue{}
		res.Value = LLVMBuildPhi(p.Builder, llvm_type, "")
		res.Type = type_

		incoming_blocks[0] = p.CurrBlock.Preds[0].Block
		incoming_blocks[1] = p.CurrBlock.Preds[1].Block

		LLVMAddIncoming(res.Value, incoming_values[:], incoming_blocks[:], 2)
		return res

	case Ast_TernaryWhenExpr:
		te := &expr.TernaryWhenExpr
		tav := type_and_value_of_expr(te.Cond)
		if tav.Value.Kind == ExactValue_Bool {
			if tav.Value.ValueBool {
				return lb_build_expr(p, te.X)
			} else {
				return lb_build_expr(p, te.Y)
			}
		}
		gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4454,
			"ternary when condition must be constant bool")

	case Ast_OrElseExpr:
		oe := &expr.OrElseExpr
		return lb_emit_or_else(p, oe.X, oe.Y, tv)

	case Ast_OrReturnExpr:
		oe := &expr.OrReturnExpr
		return lb_emit_or_return(p, oe.Expr, tv)

	case Ast_OrBranchExpr:
		be := &expr.OrBranchExpr
		var block *lbBlock

		if be.Label != nil {
			bb := lb_lookup_branch_blocks(p, be.Label)
			switch be.Token.Kind {
			case Token_or_break:
				block = bb.Break_
			case Token_or_continue:
				block = bb.Continue_
			}
		} else {
			for t := p.TargetList; t != nil && block == nil; t = t.Prev {
				if t.IsBlock {
					continue
				}
				switch be.Token.Kind {
				case Token_or_break:
					block = t.Break_
				case Token_or_continue:
					block = t.Continue_
				}
			}
		}

		if block == nil {
			gb_assert_handler("Assertion Failure", nil, "src/llvm_backend_expr.cpp", 4493, "or_branch block not found")
		}

		var lhs lbValue
		var rhs lbValue
		lb_emit_try_lhs_rhs(p, be.Expr, tv, &lhs, &rhs)
		type_ = default_type(tv.Type)
		if lhs.Value != 0 {
			lhs = lb_emit_conv(p, lhs, type_)
		} else if type_ != nil && type_ != t_invalid {
			lhs = lb_const_nil(p.Module, type_)
		}

		then_ := lb_create_block(p, "or_branch.then")
		else_ := lb_create_block(p, "or_branch.else")

		lb_emit_if(p, lb_emit_try_has_value(p, rhs), then_, else_)
		lb_start_block(p, else_)
		lb_emit_defer_stmts(p, lbDeferExit_Branch, block, expr)
		lb_emit_jump(p, block)
		lb_start_block(p, then_)

		return lhs

	case Ast_TypeAssertion:
		ta := &expr.TypeAssertion
		pos := ast_token(expr).Pos
		e := lb_build_expr(p, ta.Expr)
		t := type_deref(e.Type)
		if is_type_union(t) {
			if ta.Ignores[0] {
				return lb_emit_union_cast_only_ok_check(p, e, type_, pos)
			}
			return lb_emit_union_cast(p, e, type_, pos)
		} else if is_type_any(t) {
			return lb_emit_any_cast(p, e, type_, pos)
		} else {
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4530,
				"TODO(bill): type assertion %s", type_to_string(e.Type))
		}

	case Ast_TypeCast:
		tc := &expr.TypeCast
		e := lb_build_expr(p, tc.Expr)
		switch tc.Token.Kind {
		case Token_cast:
			return lb_emit_conv(p, e, type_)
		case Token_transmute:
			return lb_emit_transmute(p, e, type_)
		}
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4542, "Invalid AST TypeCast")

	case Ast_AutoCast:
		ac := &expr.AutoCast
		value := lb_build_expr(p, ac.Expr)
		return lb_emit_conv(p, value, type_)

	case Ast_UnaryExpr:
		ue := &expr.UnaryExpr
		switch ue.Op.Kind {
		case Token_And:
			return lb_build_unary_and(p, expr)
		default:
			v := lb_build_expr(p, ue.Expr)
			return lb_emit_unary_arith(p, ue.Op.Kind, v, type_)
		}

	case Ast_BinaryExpr:
		return lb_build_binary_expr(p, expr)

	case Ast_ProcLit:
		return lb_generate_anonymous_proc_lit(p.Module, p.Name, expr, p)

	case Ast_CompoundLit:
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_CallExpr:
		return lb_build_call_expr(p, expr)

	case Ast_SliceExpr:
		se := &expr.SliceExpr
		if is_type_slice(type_of_expr(se.Expr)) {
			if se.High == nil && (se.Low == nil || lb_is_expr_constant_zero(se.Low)) {
				return lb_build_expr(p, se.Expr)
			}
		}
		{
			src_type := base_type(type_of_expr(se.Expr))
			if is_type_pointer(src_type) {
				src_type = base_type(type_deref(src_type))
			}
			if is_type_slice(src_type) ||
				is_type_dynamic_array(src_type) ||
				is_type_string(src_type) ||
				is_type_string16(src_type) {
				return lb_build_slice_expr_value(p, expr)
			}
			if is_type_multi_pointer(src_type) && se.High != nil {
				return lb_build_slice_expr_value(p, expr)
			}
		}
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_IndexExpr:
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_MatrixIndexExpr:
		return lb_addr_load(p, lb_build_addr(p, expr))

	case Ast_InlineAsmExpr:
		ia := &expr.InlineAsmExpr
		t := type_of_expr(expr)
		_ = t

		var asm_string String
		var constraints_string String

		tav := type_and_value_of_expr(ia.AsmString)
		if tav.Value.Kind == ExactValue_String {
			asm_string = tav.Value.ValueString
		}

		tav = type_and_value_of_expr(ia.ConstraintsString)
		if tav.Value.Kind == ExactValue_String {
			constraints_string = tav.Value.ValueString
		}

		dialect := LLVMInlineAsmDialectATT
		switch ia.Dialect {
		case InlineAsmDialect_Default:
			dialect = LLVMInlineAsmDialectATT
		case InlineAsmDialect_ATT:
			dialect = LLVMInlineAsmDialectATT
		case InlineAsmDialect_Intel:
			dialect = LLVMInlineAsmDialectIntel
		default:
			gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4639, "Unhandled inline asm dialect")
		}

		func_type := lb_type_internal_for_procedures_raw(p.Module, t)
		the_asm := llvm_get_inline_asm(func_type, asm_string, constraints_string, ia.HasSideEffects, ia.HasSideEffects, dialect)
		return lbValue{Value: the_asm, Type: t}
	}

	gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 4649,
		"lb_build_expr: %.*s", len(ast_strings[expr.Kind]), ast_strings[expr.Kind])
	return lbValue{}
}

func lb_build_expr_original_const_type(expr *Ast) *Type {
	expr = unparen_expr(expr)
	type_ := type_of_expr(expr)
	if is_type_union(type_) {
		if expr.Kind == Ast_CallExpr {
			if expr.CallExpr.Proc.Tav.Mode == Addressing_Type {
				res := lb_build_expr_original_const_type(expr.CallExpr.Args[0])
				return res
			}
		}
	}
	return type_of_expr(expr)
}

func lb_build_expr(p *lbProcedure, expr *Ast) lbValue {
	prev_state_flags := p.StateFlags
	defer func() { p.StateFlags = prev_state_flags }()

	if expr.StateFlags != 0 {
		in_ := expr.StateFlags
		out := p.StateFlags

		if in_&uint16(StateFlag_BoundsCheck) != 0 {
			out |= uint16(StateFlag_BoundsCheck)
			out &^= uint16(StateFlag_NoBoundsCheck)
		} else if in_&uint16(StateFlag_NoBoundsCheck) != 0 {
			out |= uint16(StateFlag_NoBoundsCheck)
			out &^= uint16(StateFlag_BoundsCheck)
		}

		if in_&uint16(StateFlag_TypeAssert) != 0 {
			out |= uint16(StateFlag_TypeAssert)
			out &^= uint16(StateFlag_NoTypeAssert)
		} else if in_&uint16(StateFlag_NoTypeAssert) != 0 {
			out |= uint16(StateFlag_NoTypeAssert)
			out &^= uint16(StateFlag_TypeAssert)
		}

		p.StateFlags = out
	}

	if expr.StateFlags&uint16(StateFlag_SelectorCallExpr) != 0 {
		pp, ok := p.SelectorValues[expr]
		if ok {
			res := pp
			delete(p.SelectorValues, expr)
			return res
		}
		pa, ok := p.SelectorAddr[expr]
		if ok {
			res := pa
			delete(p.SelectorAddr, expr)
			return lb_addr_load(p, res)
		}
	}
	res := lb_build_expr_internal(p, expr)
	if expr.StateFlags&uint16(StateFlag_SelectorCallExpr) != 0 {
		p.SelectorValues[expr] = res
	}
	return res
}
