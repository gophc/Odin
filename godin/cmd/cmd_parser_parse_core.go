package cmd

func parse_ident(f *AstFile, allow_poly_names ...bool) *Ast {
	apn := false
	if len(allow_poly_names) > 0 {
		apn = allow_poly_names[0]
	}
	token := f.CurrToken
	if token.Kind == TokenIdent {
		advance_token(f)
	} else if apn && token.Kind == TokenDollar {
		dollar := expect_token(f, TokenDollar)
		name := ast_ident(f, expect_token(f, TokenIdent))
		if is_blank_ident(name) {
			syntax_error(name.Kind, "Invalid polymorphic type definition with a blank identifier")
		}
		return ast_poly_type(f, dollar, name, nil)
	} else {
		token = BlankToken
		expect_token(f, TokenIdent)
	}
	return ast_ident(f, token)
}

func parse_tag_expr(f *AstFile, expression *Ast) *Ast {
	token := expect_token(f, TokenHash)
	name := expect_token(f, TokenIdent)
	return ast_tag_expr(f, token, name, expression)
}

func parse_value(f *AstFile) *Ast {
	if f.CurrToken.Kind == TokenOpenBrace {
		return parse_literal_value(f, nil)
	}
	prev := f.AllowRange
	f.AllowRange = true
	value := parse_expr(f, false)
	f.AllowRange = prev
	return value
}

func parse_literal_value(f *AstFile, typ *Ast) *Ast {
	open := expect_token(f, TokenOpenBrace)
	prev := f.ExprLevel
	f.ExprLevel = 0
	var elems []*Ast
	if f.CurrToken.Kind != TokenCloseBrace {
		elems = parse_element_list(f)
	}
	f.ExprLevel = prev
	close := expect_closing(f, TokenCloseBrace, String{Data: strData("compound literal"), Len: isize(len("compound literal"))})
	return ast_compound_lit(f, typ, elems, open, close)
}

func parse_element_list(f *AstFile) []*Ast {
	var elems []*Ast
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		elem := parse_value(f)
		if f.CurrToken.Kind == TokenEq {
			eq := expect_token(f, TokenEq)
			value := parse_value(f)
			elem = ast_field_value(f, elem, value, eq)
		}
		elems = append(elems, elem)
		if !allow_field_separator(f) {
			break
		}
	}
	return elems
}

func parse_enum_field_list(f *AstFile) []*Ast {
	var elems []*Ast
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		parse_enforce_tabs(f)
		docs := f.LeadComment
		name := parse_value(f)
		var value *Ast
		if f.CurrToken.Kind == TokenEq {
			eq := expect_token(f, TokenEq)
			value = parse_value(f)
			_ = eq
		}
		comment := consume_line_comment(f)
		elem := ast_enum_field_value(f, name, value, docs, comment)
		elems = append(elems, elem)
		if !allow_field_separator(f) {
			break
		}
		if elem.Kind == AstEnumFieldValue && elem.EnumFieldValue.Comment == nil {
			elem.EnumFieldValue.Comment = consume_line_comment(f)
		}
	}
	return elems
}

func parse_union_variant_list(f *AstFile) []*Ast {
	var variants []*Ast
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		typ := parse_type(f)
		if typ.Kind != AstBadExpr {
			variants = append(variants, typ)
		}
		if !allow_field_separator(f) {
			break
		}
	}
	return variants
}

func token_precedence(f *AstFile, t TokenKind) int32 {
	switch t {
	case TokenQuestion, TokenIf, TokenWhen, TokenOrElse:
		return 1
	case TokenEllipsis, TokenRangeFull, TokenRangeHalf:
		if !f.AllowRange {
			return 0
		}
		return 2
	case TokenCmpOr:
		return 3
	case TokenCmpAnd:
		return 4
	case TokenCmpEq, TokenNotEq, TokenLt, TokenGt, TokenLtEq, TokenGtEq:
		return 5
	case TokenIn, TokenNotIn:
		if f.ExprLevel < 0 && !f.AllowInExpr {
			return 0
		}
		return 5
	case TokenAdd, TokenSub, TokenOr, TokenXor:
		return 6
	case TokenMul, TokenQuo, TokenMod, TokenModMod, TokenAnd, TokenAndNot, TokenShl, TokenShr:
		return 7
	}
	return 0
}

func parse_binary_expr(f *AstFile, lhs bool, prec_in int32) *Ast {
	expr := parse_unary_expr(f, lhs)
	for {
		op := f.CurrToken
		op_prec := token_precedence(f, op.Kind)
		if op_prec < prec_in {
			break
		}
		prev := f.PrevToken
		switch op.Kind {
		case TokenIf, TokenWhen:
			if prev.Pos.Line < op.Pos.Line {
				goto loop_end
			}
		}
		expect_operator(f)
		if op.Kind == TokenQuestion {
			cond := expr
			x := parse_expr(f, lhs)
			token_c := expect_token(f, TokenColon)
			y := parse_expr(f, lhs)
			_ = token_c
			expr = ast_ternary_if_expr(f, x, cond, y)
		} else if op.Kind == TokenIf || op.Kind == TokenWhen {
			x := expr
			cond := parse_expr(f, lhs)
			expect_token(f, TokenElse)
			y := parse_expr(f, lhs)
			switch op.Kind {
			case TokenIf:
				expr = ast_ternary_if_expr(f, x, cond, y)
			case TokenWhen:
				expr = ast_ternary_when_expr(f, x, cond, y)
			}
		} else {
			right := parse_binary_expr(f, false, op_prec+1)
			if right == nil {
				syntax_error(op.Kind, "Expected expression on the right-hand side of the binary operator")
			}
			if op.Kind == TokenOrElse {
				expr = ast_or_else_expr(f, expr, op, right)
			} else {
				expr = ast_binary_expr(f, op, expr, right)
			}
		}
		lhs = false
	}
loop_end:
	return expr
}

func parse_expr(f *AstFile, lhs bool) *Ast {
	return parse_binary_expr(f, lhs, 0+1)
}

func parse_expr_list(f *AstFile, lhs bool) []*Ast {
	prev := f.AllowNewline
	f.AllowNewline = file_allow_newline(f)
	var list []*Ast
	for {
		e := parse_expr(f, lhs)
		list = append(list, e)
		if f.CurrToken.Kind != TokenComma || f.CurrToken.Kind == TokenEOF {
			break
		}
		advance_token(f)
	}
	f.AllowNewline = prev
	return list
}

func parse_lhs_expr_list(f *AstFile) []*Ast {
	return parse_expr_list(f, true)
}

func parse_rhs_expr_list(f *AstFile) []*Ast {
	return parse_expr_list(f, false)
}

func parse_ident_list(f *AstFile, allow_poly_names bool) []*Ast {
	var list []*Ast
	for {
		list = append(list, parse_ident(f, allow_poly_names))
		if f.CurrToken.Kind != TokenComma || f.CurrToken.Kind == TokenEOF {
			break
		}
		advance_token(f)
	}
	return list
}

func parse_operand(f *AstFile, lhs bool) *Ast {
	switch f.CurrToken.Kind {
	case TokenIdent:
		return parse_ident(f)
	case TokenUninit:
		return ast_uninit(f, expect_token(f, TokenUninit))
	case TokenContext:
		return ast_implicit(f, expect_token(f, TokenContext))
	case TokenInteger, TokenFloat, TokenImag, TokenRune:
		return ast_basic_lit(f, advance_token(f))
	case TokenString:
		return ast_basic_lit(f, advance_token(f))
	case TokenOpenBrace:
		if !lhs {
			return parse_literal_value(f, nil)
		}
	case TokenOpenParen:
		open := expect_token(f, TokenOpenParen)
		if f.PrevToken.Kind == TokenCloseParen {
			close := expect_token(f, TokenCloseParen)
			syntax_error(open.Kind, "Invalid parentheses expression with no inside expression")
			return ast_bad_expr(f, open, close)
		}
		prev_expr_level := f.ExprLevel
		prev_allow_newline := f.AllowNewline
		if f.ExprLevel < 0 {
			f.AllowNewline = false
		}
		if f.ExprLevel > 0 {
			f.ExprLevel = f.ExprLevel
		} else {
			f.ExprLevel = 0
		}
		f.ExprLevel++
		operand := parse_expr(f, false)
		f.AllowNewline = prev_allow_newline
		f.ExprLevel = prev_expr_level
		close := expect_token(f, TokenCloseParen)
		return ast_paren_expr(f, operand, open, close)
	case TokenDistinct:
		token := expect_token(f, TokenDistinct)
		typ := parse_type(f)
		return ast_distinct_type(f, token, typ)
	case TokenHash:
		token := expect_token(f, TokenHash)
		name := expect_token(f, TokenIdent)
		switch goStr(name.String) {
		case "type":
			return ast_helper_type(f, token, parse_type(f))
		case "simd":
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			typ := unparen_expr(original_type)
			switch typ.Kind {
			case AstArrayType:
				typ.ArrayType.Tag = tag
			default:
				syntax_error(typ.Kind, "Expected a fixed array type after #simd")
			}
			return original_type
		case "soa":
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			typ := unparen_expr(original_type)
			switch typ.Kind {
			case AstArrayType:
				typ.ArrayType.Tag = tag
			case AstDynamicArrayType:
				typ.DynamicArrayType.Tag = tag
			case AstPointerType:
				typ.PointerType.Tag = tag
			case AstFixedCapacityDynamicArrayType:
				typ.FixedCapacityDynamicArrayType.Tag = tag
			default:
				syntax_error(typ.Kind, "Expected an array or pointer type after #soa")
			}
			return original_type
		case "row_major", "column_major":
			original_type := parse_type(f)
			typ := unparen_expr(original_type)
			switch typ.Kind {
			case AstMatrixType:
				typ.MatrixType.IsRowMajor = goStr(name.String) == "row_major"
			default:
				syntax_error(typ.Kind, "Expected a matrix type after #row_major/#column_major")
			}
			return original_type
		case "partial":
			tag := ast_basic_directive(f, token, name)
			original_expr := parse_expr(f, lhs)
			expr := unparen_expr(original_expr)
			if expr == nil {
				syntax_error(name.Kind, "Expected a compound literal after #partial")
				return ast_bad_expr(f, token, name)
			}
			switch expr.Kind {
			case AstCompoundLit:
				expr.CompoundLit.Tag = tag
			default:
				syntax_error(expr.Kind, "Expected a compound literal after #partial")
			}
			return original_expr
		case "sparse":
			tag := ast_basic_directive(f, token, name)
			original_type := parse_type(f)
			typ := unparen_expr(original_type)
			switch typ.Kind {
			case AstArrayType:
				typ.ArrayType.Tag = tag
			default:
				syntax_error(typ.Kind, "Expected an enumerated array type after #sparse")
			}
			return original_type
		case "bounds_check":
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlagBoundsCheck)
		case "no_bounds_check":
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlagNoBoundsCheck)
		case "type_assert":
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlagTypeAssert)
		case "no_type_assert":
			operand := parse_expr(f, lhs)
			return parse_check_directive_for_statement(operand, name, StateFlagNoTypeAssert)
		case "relative":
			tag := ast_basic_directive(f, token, name)
			if f.CurrToken.Kind != TokenOpenParen {
				syntax_error(tag.Kind, "expected #relative(<integer type>) <type>")
			} else {
				tag = parse_call_expr(f, tag)
			}
			typ := parse_type(f)
			syntax_error(tag.Kind, "#relative types have now been removed in favour of \"core:relative\"")
			return ast_relative_type(f, tag, typ)
		case "force_inline", "force_no_inline", "must_tail":
			return parse_inlining_or_tailing_operand(f, name)
		}
		return ast_basic_directive(f, token, name)
	case TokenProc:
		token := expect_token(f, TokenProc)
		if f.CurrToken.Kind == TokenOpenBrace {
			open := expect_token(f, TokenOpenBrace)
			var args []*Ast
			for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
				elem := parse_expr(f, false)
				args = append(args, elem)
				if !allow_field_separator(f) {
					break
				}
			}
			close := expect_token(f, TokenCloseBrace)
			if len(args) == 0 {
				syntax_error(token.Kind, "Expected at least 1 argument in a procedure group")
			}
			return ast_proc_group(f, token, open, close, args)
		}
		typ := parse_proc_type(f, token)
		where_token := Token{}
		var where_clauses []*Ast
		var tags uint64
		skip_possible_newline_for_literal(f)
		if f.CurrToken.Kind == TokenWhere {
			where_token = expect_token(f, TokenWhere)
			prev_level := f.ExprLevel
			f.ExprLevel = -1
			where_clauses = parse_rhs_expr_list(f)
			f.ExprLevel = prev_level
		}
		parse_proc_tags(f, &tags)
		if tags&uint64(ProcTagRequireResults) != 0 {
			syntax_error(f.CurrToken.Kind, "#require_results has now been replaced as an attribute @(require_results) on the declaration")
			tags &^= uint64(ProcTagRequireResults)
		}
	if typ != nil && typ.Kind == AstProcType {
		typ.ProcType.Tags = tags
	}
		if f.AllowType && f.ExprLevel < 0 {
			if tags != 0 {
				syntax_error(token.Kind, "A procedure type cannot have suffix tags")
			}
			if where_token.Kind != TokenInvalid {
				syntax_error(where_token.Kind, "'where' clauses are not allowed on procedure types")
			}
			return typ
		}
		skip_possible_newline_for_literal(f, where_token.Kind == TokenWhere)
		if allow_token(f, TokenUninit) {
			if where_token.Kind != TokenInvalid {
				syntax_error(where_token.Kind, "'where' clauses are not allowed on procedure literals without a defined body (replaced with ---)")
			}
			return ast_proc_lit(f, typ, nil, tags, where_token, where_clauses)
		} else if f.CurrToken.Kind == TokenOpenBrace {
			curr_proc := f.CurrProc
			var body *Ast
			f.CurrProc = typ
			body = parse_body(f)
			f.CurrProc = curr_proc
			if tags&uint64(ProcTagNoBoundsCheck) != 0 {
				body.StateFlags |= uint8(StateFlagNoBoundsCheck)
			}
			if tags&uint64(ProcTagBoundsCheck) != 0 {
				body.StateFlags |= uint8(StateFlagBoundsCheck)
			}
			if tags&uint64(ProcTagNoTypeAssert) != 0 {
				body.StateFlags |= uint8(StateFlagNoTypeAssert)
			}
			if tags&uint64(ProcTagTypeAssert) != 0 {
				body.StateFlags |= uint8(StateFlagTypeAssert)
			}
			return ast_proc_lit(f, typ, body, tags, where_token, where_clauses)
		} else if allow_token(f, TokenDo) {
			curr_proc := f.CurrProc
			var body *Ast
			f.CurrProc = typ
			body = convert_stmt_to_body(f, parse_stmt(f))
			f.CurrProc = curr_proc
			syntax_error(body.Kind, "'do' for procedure bodies is not allowed, prefer {}")
			return ast_proc_lit(f, typ, body, tags, where_token, where_clauses)
		}
		if tags != 0 {
			syntax_error(token.Kind, "A procedure type cannot have suffix tags")
		}
		if where_token.Kind != TokenInvalid {
			syntax_error(where_token.Kind, "'where' clauses are not allowed on procedure types")
		}
		return typ
	case TokenDollar:
		token := expect_token(f, TokenDollar)
		typ := parse_ident(f)
		if is_blank_ident(typ) {
			syntax_error(typ.Kind, "Invalid polymorphic type definition with a blank identifier")
		}
		var specialization *Ast
		if allow_token(f, TokenQuo) {
			specialization = parse_type(f)
		}
		return ast_poly_type(f, token, typ, specialization)
	case TokenTypeid:
		token := expect_token(f, TokenTypeid)
		return ast_typeid_type(f, token, nil)
	case TokenPointer:
		token := expect_token(f, TokenPointer)
		elem := parse_type(f)
		return ast_pointer_type(f, token, elem)
	case TokenMul:
		return parse_unary_expr(f, true)
	case TokenOpenBracket:
		token := expect_token(f, TokenOpenBracket)
		if f.CurrToken.Kind == TokenPointer {
			expect_token(f, TokenPointer)
			expect_token(f, TokenCloseBracket)
			return ast_multi_pointer_type(f, token, parse_type(f))
		} else if f.CurrToken.Kind == TokenQuestion {
			count_expr := ast_unary_expr(f, expect_token(f, TokenQuestion), nil)
			expect_token(f, TokenCloseBracket)
			return ast_array_type(f, token, count_expr, parse_type(f))
		} else if allow_token(f, TokenDynamic) {
			var capacity *Ast
			if f.CurrToken.Kind == TokenSemicolon && goStr(f.CurrToken.String) == ";" {
				expect_token(f, TokenSemicolon)
				capacity = parse_expr(f, false)
			} else if allow_token(f, TokenComma) || allow_token(f, TokenSemicolon) {
				capacity = parse_expr(f, false)
			}
			expect_token(f, TokenCloseBracket)
			elem := parse_type(f)
			if capacity == nil {
				return ast_dynamic_array_type(f, token, elem)
			}
			return ast_fixed_capacity_dynamic_array_type(f, token, capacity, elem)
		} else if f.CurrToken.Kind != TokenCloseBracket {
			f.ExprLevel++
			count_expr := parse_expr(f, false)
			f.ExprLevel--
			expect_token(f, TokenCloseBracket)
			return ast_array_type(f, token, count_expr, parse_type(f))
		}
		expect_token(f, TokenCloseBracket)
		return ast_array_type(f, token, nil, parse_type(f))
	case TokenMap:
		token := expect_token(f, TokenMap)
		open := expect_token_after(f, TokenOpenBracket, "map")
		key := parse_expr(f, true)
		close := expect_token(f, TokenCloseBracket)
		value := parse_type(f)
		return ast_map_type(f, token, key, value)
	case TokenMatrix:
		token := expect_token(f, TokenMatrix)
		open := expect_token_after(f, TokenOpenBracket, "matrix")
		row_count := parse_expr(f, true)
		expect_token(f, TokenComma)
		column_count := parse_expr(f, true)
		close := expect_token(f, TokenCloseBracket)
		typ := parse_type(f)
		return ast_matrix_type(f, token, row_count, column_count, typ)
	case TokenBitField:
		token := expect_token(f, TokenBitField)
		prev_level := f.ExprLevel
		f.ExprLevel = -1
		backing_type := parse_type_or_ident(f)
		if backing_type == nil {
			tok := advance_token(f)
			syntax_error(tok.Kind, "Expected a backing type for a 'bit_field'")
			backing_type = ast_bad_expr(f, tok, f.CurrToken)
		}
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, TokenOpenBrace, "bit_field")
		var fields []*Ast
		for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
			docs := f.LeadComment
			comment := f.LineComment
			name := parse_ident(f)
			err_once := false
			for allow_token(f, TokenComma) {
				dummy_name := parse_ident(f)
				if !err_once {
					syntax_error(dummy_name.Kind, "'bit_field' fields do not support multiple names per field")
					err_once = true
				}
				_ = dummy_name
			}
			expect_token(f, TokenColon)
			typ := parse_type(f)
			expect_token(f, TokenOr)
			bit_size := parse_expr(f, true)
			tag := Token{}
			if f.CurrToken.Kind == TokenString {
				tag = expect_token(f, TokenString)
			}
			bf_field := ast_bit_field_field(f, name, typ, bit_size, tag, docs, comment)
			fields = append(fields, bf_field)
			if !allow_field_separator(f) {
				break
			}
		}
		close := expect_closing(f, TokenCloseBrace, String{Data: strData("bit_field"), Len: isize(len("bit_field"))})
		f.ExprLevel = prev_level
		return ast_bit_field_type(f, token, backing_type, open, fields, close)
	case TokenStruct:
		token := expect_token(f, TokenStruct)
		var polymorphic_params *Ast
		is_packed := false
		is_all_or_none := false
		is_raw_union := false
		is_simple := false
		var align *Ast
		var min_field_align *Ast
		var max_field_align *Ast

		if allow_token(f, TokenOpenParen) {
			param_count := isize(0)
			polymorphic_params = parse_field_list(f, &param_count, 0, TokenCloseParen, true, true)
			if param_count == 0 {
				syntax_error(polymorphic_params.Kind, "Expected at least 1 polymorphic parameter")
				polymorphic_params = nil
			}
			expect_token_after(f, TokenCloseParen, "parameter list")
			check_polymorphic_params_for_type(f, polymorphic_params, token)
		}
		prev_level := f.ExprLevel
		f.ExprLevel = -1
		for allow_token(f, TokenHash) {
			tag := expect_token_after(f, TokenIdent, "#")
			switch goStr(tag.String) {
			case "packed":
				if is_packed {
					syntax_error(tag.Kind, "Duplicate struct tag '#packed'")
				}
				is_packed = true
			case "all_or_none":
				if is_all_or_none {
					syntax_error(tag.Kind, "Duplicate struct tag '#all_or_none'")
				}
				is_all_or_none = true
			case "align":
				if align != nil {
					syntax_error(tag.Kind, "Duplicate struct tag '#align'")
				}
				align = parse_expr(f, true)
				if align != nil && align.Kind != AstParenExpr {
					syntax_error(tag.Kind, "#align requires parentheses around the expression")
				}
			case "field_align":
				if min_field_align != nil {
					syntax_error(tag.Kind, "Duplicate struct tag '#field_align'")
				}
				syntax_error(tag.Kind, "#field_align has been deprecated in favour of #min_field_align")
				min_field_align = parse_expr(f, true)
			case "min_field_align":
				if min_field_align != nil {
					syntax_error(tag.Kind, "Duplicate struct tag '#min_field_align'")
				}
				min_field_align = parse_expr(f, true)
			case "max_field_align":
				if max_field_align != nil {
					syntax_error(tag.Kind, "Duplicate struct tag '#max_field_align'")
				}
				max_field_align = parse_expr(f, true)
			case "raw_union":
				if is_raw_union {
					syntax_error(tag.Kind, "Duplicate struct tag '#raw_union'")
				}
				is_raw_union = true
			case "simple":
				if is_simple {
					syntax_error(tag.Kind, "Duplicate struct tag '#simple'")
				}
				is_simple = true
			default:
				syntax_error(tag.Kind, "Invalid struct tag")
			}
		}
		f.ExprLevel = prev_level
		if is_raw_union && is_packed {
			is_packed = false
			syntax_error(token.Kind, "'#raw_union' cannot also be '#packed'")
		}
		if is_raw_union && is_all_or_none {
			is_all_or_none = false
			syntax_error(token.Kind, "'#raw_union' cannot also be '#all_or_none'")
		}
		where_token := Token{}
		var where_clauses []*Ast
		skip_possible_newline_for_literal(f)
		if f.CurrToken.Kind == TokenWhere {
			where_token = expect_token(f, TokenWhere)
			prev_level = f.ExprLevel
			f.ExprLevel = -1
			where_clauses = parse_rhs_expr_list(f)
			f.ExprLevel = prev_level
		}
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, TokenOpenBrace, "struct")
		name_count := isize(0)
		fields := parse_struct_field_list(f, &name_count)
		close := expect_closing(f, TokenCloseBrace, String{Data: strData("struct"), Len: isize(len("struct"))})
		var decls []*Ast
		if fields != nil && fields.Kind == AstFieldList {
			decls = fields.FieldList.List
		}
		parser_check_polymorphic_record_parameters(f, polymorphic_params)
		return ast_struct_type(f, token, decls, name_count,
			polymorphic_params, is_packed, is_raw_union, is_all_or_none, is_simple,
			align, min_field_align, max_field_align,
			where_token, where_clauses)
	case TokenUnion:
		token := expect_token(f, TokenUnion)
		var polymorphic_params *Ast
		var align *Ast
		no_nil := false
		maybe := false
		shared_nil := false
		union_kind := UnionTypeNormal

		if allow_token(f, TokenOpenParen) {
			param_count := isize(0)
			polymorphic_params = parse_field_list(f, &param_count, 0, TokenCloseParen, true, true)
			if param_count == 0 {
				syntax_error(polymorphic_params.Kind, "Expected at least 1 polymorphic parametric")
				polymorphic_params = nil
			}
			expect_token_after(f, TokenCloseParen, "parameter list")
			check_polymorphic_params_for_type(f, polymorphic_params, token)
		}
		for allow_token(f, TokenHash) {
			tag := expect_token_after(f, TokenIdent, "#")
			switch goStr(tag.String) {
			case "align":
				if align != nil {
					syntax_error(tag.Kind, "Duplicate union tag '#align'")
				}
				align = parse_expr(f, true)
			case "no_nil":
				if no_nil {
					syntax_error(tag.Kind, "Duplicate union tag '#no_nil'")
				}
				no_nil = true
			case "shared_nil":
				if shared_nil {
					syntax_error(tag.Kind, "Duplicate union tag '#shared_nil'")
				}
				shared_nil = true
			case "maybe":
				if maybe {
					syntax_error(tag.Kind, "Duplicate union tag '#maybe'")
				}
				maybe = true
			default:
				syntax_error(tag.Kind, "Invalid union tag")
			}
		}
		if no_nil && shared_nil {
			syntax_error(f.CurrToken.Kind, "#shared_nil and #no_nil cannot be applied together")
		}
		if maybe {
			syntax_error(f.CurrToken.Kind, "#maybe functionality has now been merged with standard 'union' functionality")
		}
		if no_nil {
			union_kind = UnionTypeNoNil
		} else if shared_nil {
			union_kind = UnionTypeSharedNil
		}
		skip_possible_newline_for_literal(f)
		where_token := Token{}
		var where_clauses []*Ast
		if f.CurrToken.Kind == TokenWhere {
			where_token = expect_token(f, TokenWhere)
			prev_level := f.ExprLevel
			f.ExprLevel = -1
			where_clauses = parse_rhs_expr_list(f)
			f.ExprLevel = prev_level
		}
		skip_possible_newline_for_literal(f)
		open := expect_token_after(f, TokenOpenBrace, "union")
		variants := parse_union_variant_list(f)
		close := expect_closing(f, TokenCloseBrace, String{Data: strData("union"), Len: isize(len("union"))})
		parser_check_polymorphic_record_parameters(f, polymorphic_params)
		return ast_union_type(f, token, variants, polymorphic_params, align, union_kind, where_token, where_clauses)
	case TokenEnum:
		token := expect_token(f, TokenEnum)
		var base_type *Ast
		if f.CurrToken.Kind != TokenOpenBrace {
			base_type = parse_type(f)
		}
		skip_possible_newline_for_literal(f)
		open := expect_token(f, TokenOpenBrace)
		values := parse_enum_field_list(f)
		close := expect_closing(f, TokenCloseBrace, String{Data: strData("enum"), Len: isize(len("enum"))})
		return ast_enum_type(f, token, base_type, values)
	case TokenBitSet:
		token := expect_token(f, TokenBitSet)
		expect_token(f, TokenOpenBracket)
		var elem *Ast
		var underlying *Ast
		prev_allow_range := f.AllowRange
		f.AllowRange = true
		elem = parse_expr(f, true)
		f.AllowRange = prev_allow_range
		if elem == nil {
			syntax_error(token.Kind, "Expected a type or range, got nothing")
		}
		if f.CurrToken.Kind == TokenSemicolon {
			expect_token(f, TokenSemicolon)
			underlying = parse_type(f)
		} else if allow_token(f, TokenComma) || allow_token(f, TokenSemicolon) {
			underlying = parse_type(f)
		}
		expect_token(f, TokenCloseBracket)
		return ast_bit_set_type(f, token, elem, underlying)
	case TokenAsm:
		token := expect_token(f, TokenAsm)
		var param_types []*Ast
		var return_type *Ast
		if allow_token(f, TokenOpenParen) {
			param_types = make([]*Ast, 0)
			for f.CurrToken.Kind != TokenCloseParen && f.CurrToken.Kind != TokenEOF {
				t := parse_type(f)
				param_types = append(param_types, t)
				if f.CurrToken.Kind != TokenComma || f.CurrToken.Kind == TokenEOF {
					break
				}
				advance_token(f)
			}
			expect_token(f, TokenCloseParen)
			if allow_token(f, TokenArrowRight) {
				return_type = parse_type(f)
			}
		}
		has_side_effects := false
		is_align_stack := false
		dialect := InlineAsmDialectDefault
		for f.CurrToken.Kind == TokenHash {
			advance_token(f)
			if f.CurrToken.Kind == TokenIdent {
				tok := advance_token(f)
				switch goStr(tok.String) {
				case "side_effects":
					if has_side_effects {
						syntax_error(tok.Kind, "Duplicate directive on inline asm expression: '#side_effects'")
					}
					has_side_effects = true
				case "align_stack":
					if is_align_stack {
						syntax_error(tok.Kind, "Duplicate directive on inline asm expression: '#align_stack'")
					}
					is_align_stack = true
				case "att":
					if dialect == InlineAsmDialectATT {
						syntax_error(tok.Kind, "Duplicate directive on inline asm expression: '#att'")
					} else if dialect != InlineAsmDialectDefault {
						syntax_error(tok.Kind, "Conflicting asm dialects")
					} else {
						dialect = InlineAsmDialectATT
					}
				case "intel":
					if dialect == InlineAsmDialectIntel {
						syntax_error(tok.Kind, "Duplicate directive on inline asm expression: '#intel'")
					} else if dialect != InlineAsmDialectDefault {
						syntax_error(tok.Kind, "Conflicting asm dialects")
					} else {
						dialect = InlineAsmDialectIntel
					}
				default:
					syntax_error(tok.Kind, "Invalid directive on inline asm expression")
				}
			} else {
				syntax_error(f.CurrToken.Kind, "Expected an identifier after hash")
			}
		}
		skip_possible_newline_for_literal(f)
		open := expect_token(f, TokenOpenBrace)
		asm_string := parse_expr(f, false)
		expect_token(f, TokenComma)
		constraints_string := parse_expr(f, false)
		allow_token(f, TokenComma)
		close := expect_closing(f, TokenCloseBrace, String{Data: strData("inline asm"), Len: isize(len("inline asm"))})
		return ast_inline_asm_expr(f, token, open, close, param_types, return_type, asm_string, constraints_string, has_side_effects, is_align_stack, dialect)
	}
	return nil
}

func is_literal_type(node *Ast) bool {
	node = unparen_expr(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case AstBadExpr, AstIdent, AstSelectorExpr, AstArrayType, AstStructType,
		AstUnionType, AstEnumType, AstFixedCapacityDynamicArrayType,
		AstDynamicArrayType, AstMapType, AstBitSetType, AstMatrixType,
		AstCallExpr, AstMultiPointerType:
		return true
	}
	return false
}

func parse_call_expr(f *AstFile, operand *Ast) *Ast {
	prev_expr_level := f.ExprLevel
	prev_newline := f.AllowNewline
	f.ExprLevel = 0
	f.AllowNewline = file_allow_newline(f)
	open_paren := expect_token(f, TokenOpenParen)
	var args []*Ast
	ellipsis := Token{}
	seen_ellipsis := false
	for f.CurrToken.Kind != TokenCloseParen && f.CurrToken.Kind != TokenEOF {
		if f.CurrToken.Kind == TokenComma {
			syntax_error(f.CurrToken.Kind, "Expected an expression not ,")
		} else if f.CurrToken.Kind == TokenEq {
			syntax_error(f.CurrToken.Kind, "Expected an expression not =")
		}
		prefix_ellipsis := false
		if f.CurrToken.Kind == TokenEllipsis {
			prefix_ellipsis = true
			ellipsis = expect_token(f, TokenEllipsis)
		}
		arg := parse_expr(f, false)
		if f.CurrToken.Kind == TokenEq {
			eq := expect_token(f, TokenEq)
			if prefix_ellipsis {
				syntax_error(ellipsis.Kind, "'..' must be applied to value rather than the field name")
			}
			value := parse_value(f)
			arg = ast_field_value(f, arg, value, eq)
		} else if seen_ellipsis {
			syntax_error(arg.Kind, "Positional arguments are not allowed after '..'")
		}
		args = append(args, arg)
		if ellipsis.Pos.Line != 0 {
			seen_ellipsis = true
		}
		if !allow_field_separator(f) {
			break
		}
	}
	f.AllowNewline = prev_newline
	f.ExprLevel = prev_expr_level
	close_paren := expect_closing(f, TokenCloseParen, String{Data: strData("argument list"), Len: isize(len("argument list"))})
	call := ast_call_expr(f, operand, args, open_paren, close_paren, ellipsis)
	o := unparen_expr(operand)
	if o != nil && o.Kind == AstSelectorExpr && o.SelectorExpr.Token.Kind == TokenArrowRight {
		return ast_selector_call_expr(f, o.SelectorExpr.Token, o, call)
	}
	return call
}

func is_ast_range(expr *Ast) bool {
	if expr == nil {
		return false
	}
	if expr.Kind != AstBinaryExpr {
		return false
	}
	return is_token_range(expr.BinaryExpr.Op.Kind)
}

func parse_atom_expr(f *AstFile, operand *Ast, lhs bool) *Ast {
	if operand == nil {
		if f.AllowType {
			return nil
		}
		begin := f.CurrToken
		syntax_error(begin.Kind, "Expected an operand")
		fix_advance_to_next_stmt(f)
		operand = ast_bad_expr(f, begin, f.CurrToken)
	}
	loop := true
	for loop {
		switch f.CurrToken.Kind {
		case TokenOpenParen:
			parse_check_or_return(operand, "call expression")
			operand = parse_call_expr(f, operand)
		case TokenPeriod:
			token := advance_token(f)
			switch f.CurrToken.Kind {
			case TokenIdent:
				parse_check_or_return(operand, "selector expression")
				operand = ast_selector_expr(f, token, operand, parse_ident(f))
			case TokenOpenParen:
				parse_check_or_return(operand, "type assertion")
				open := expect_token(f, TokenOpenParen)
				typ := parse_type(f)
				close := expect_token(f, TokenCloseParen)
				_ = open
				_ = close
				operand = ast_type_assertion(f, operand, token, typ)
			case TokenQuestion:
				parse_check_or_return(operand, ".? based type assertion")
				question := expect_token(f, TokenQuestion)
				typ := ast_unary_expr(f, question, nil)
				operand = ast_type_assertion(f, operand, token, typ)
			default:
				syntax_error(f.CurrToken.Kind, "Expected a selector")
				advance_token(f)
				operand = ast_bad_expr(f, ast_token(operand), f.CurrToken)
			}
		case TokenArrowRight:
			parse_check_or_return(operand, "-> based call expression")
			token := advance_token(f)
			operand = ast_selector_expr(f, token, operand, parse_ident(f))
		case TokenOpenBracket:
			prev_allow_range := f.AllowRange
			f.AllowRange = false
			open := Token{}
			close := Token{}
			interval := Token{}
			var indices [2]*Ast
			is_interval := false
			f.ExprLevel++
			open = expect_token(f, TokenOpenBracket)
			if f.CurrToken.Kind == TokenCloseBracket {
				syntax_error(f.CurrToken.Kind, "Expected an operand, got ]")
				close = expect_token(f, TokenCloseBracket)
				if f.AllowType {
				}
				operand = ast_index_expr(f, operand, nil, open, close)
				f.AllowRange = prev_allow_range
				break
			}
			switch f.CurrToken.Kind {
			case TokenEllipsis, TokenRangeFull, TokenRangeHalf, TokenColon:
			default:
				indices[0] = parse_expr(f, false)
			}
			switch f.CurrToken.Kind {
			case TokenEllipsis, TokenRangeFull, TokenRangeHalf:
				syntax_error(f.CurrToken.Kind, "Expected a colon, not a range")
				fallthrough
			case TokenComma, TokenColon:
				interval = advance_token(f)
				is_interval = true
				if f.CurrToken.Kind != TokenCloseBracket && f.CurrToken.Kind != TokenEOF {
					indices[1] = parse_expr(f, false)
				}
			}
			f.ExprLevel--
			close = expect_token(f, TokenCloseBracket)
			if is_interval {
				if interval.Kind == TokenComma {
					if indices[0] == nil || indices[1] == nil {
						syntax_error(open.Kind, "Matrix index expressions require both row and column indices")
					}
					parse_check_or_return(operand, "matrix index expression")
					operand = ast_matrix_index_expr(f, operand, open, close, interval, indices[0], indices[1])
				} else {
					parse_check_or_return(operand, "slice expression")
					operand = ast_slice_expr(f, operand, open, close, interval, indices[0], indices[1])
				}
			} else {
				parse_check_or_return(operand, "index expression")
				operand = ast_index_expr(f, operand, indices[0], open, close)
			}
			f.AllowRange = prev_allow_range
		case TokenPointer:
			parse_check_or_return(operand, "dereference")
			operand = ast_deref_expr(f, operand, expect_token(f, TokenPointer))
		case TokenOrReturn:
			operand = ast_or_return_expr(f, operand, expect_token(f, TokenOrReturn))
		case TokenOrBreak, TokenOrContinue:
			token := advance_token(f)
			var label *Ast
			if f.CurrToken.Kind == TokenIdent {
				label = parse_ident(f)
			}
			operand = ast_or_branch_expr(f, operand, token, label)
		case TokenOpenBrace:
			if !lhs && is_literal_type(operand) && f.ExprLevel >= 0 {
				operand = parse_literal_value(f, operand)
			} else {
				loop = false
			}
		case TokenIncrement, TokenDecrement:
			if !lhs {
				token := advance_token(f)
				syntax_error(token.Kind, "Postfix '++'/'--' operator is not supported")
			} else {
				loop = false
			}
		default:
			loop = false
		}
		lhs = false
	}
	return operand
}

func parse_unary_expr(f *AstFile, lhs bool) *Ast {
	switch f.CurrToken.Kind {
	case TokenTransmute, TokenCast:
		token := advance_token(f)
		expect_token(f, TokenOpenParen)
		typ := parse_type(f)
		expect_token(f, TokenCloseParen)
		expr := parse_unary_expr(f, lhs)
		return ast_type_cast(f, token, typ, expr)
	case TokenAutoCast:
		token := advance_token(f)
		expr := parse_unary_expr(f, lhs)
		return ast_auto_cast(f, token, expr)
	case TokenAdd, TokenSub, TokenXor, TokenAnd, TokenNot, TokenMul:
		token := advance_token(f)
		if token.Kind == TokenNot {
			skip_possible_newline(f)
		}
		expr := parse_unary_expr(f, lhs)
		return ast_unary_expr(f, token, expr)
	case TokenIncrement, TokenDecrement:
		token := advance_token(f)
		syntax_error(token.Kind, "Unary '++'/'--' operator is not supported")
		expr := parse_unary_expr(f, lhs)
		return ast_unary_expr(f, token, expr)
	case TokenPeriod:
		token := expect_token(f, TokenPeriod)
		ident := parse_ident(f)
		return ast_implicit_selector_expr(f, token, ident)
	}
	return parse_atom_expr(f, parse_operand(f, lhs), lhs)
}

func parse_type_or_ident(f *AstFile) *Ast {
	prev_allow_type := f.AllowType
	prev_expr_level := f.ExprLevel
	f.AllowType = true
	f.ExprLevel = -1
	lhs := true
	operand := parse_operand(f, lhs)
	typ := parse_atom_expr(f, operand, lhs)
	f.AllowType = prev_allow_type
	f.ExprLevel = prev_expr_level
	return typ
}

func parse_type(f *AstFile) *Ast {
	typ := parse_type_or_ident(f)
	if typ == nil {
		prev_token := f.CurrToken
		var token Token
		if f.CurrToken.Kind == TokenOpenBrace {
			token = f.CurrToken
		} else {
			token = advance_token(f)
		}
		prev_token_str := prev_token.String
		if goStr(prev_token_str) == "\n" {
			syntax_error(token.Kind, "Expected a type, got newline")
		} else {
			syntax_error(token.Kind, "Expected a type, got...")
		}
		return ast_bad_expr(f, token, f.CurrToken)
	} else if typ.Kind == AstParenExpr && unparen_expr(typ) == nil {
		syntax_error(typ.Kind, "Expected a type within the parentheses")
		return ast_bad_expr(f, typ.ParenExpr.Open, typ.ParenExpr.Close)
	}
	return typ
}

func parse_results(f *AstFile, diverging *bool) *Ast {
	if !allow_token(f, TokenArrowRight) {
		return nil
	}
	if allow_token(f, TokenNot) {
		if diverging != nil {
			*diverging = true
		}
		return nil
	}
	prev_level := f.ExprLevel
	defer func() { f.ExprLevel = prev_level }()
	if f.CurrToken.Kind != TokenOpenParen {
		begin_token := f.CurrToken
		var empty_names []*Ast
		var list []*Ast
		typ := parse_type(f)
		tag := Token{}
		list = append(list, ast_field(f, empty_names, typ, nil, 0, tag, nil, nil))
		_ = begin_token
		return ast_field_list(f, begin_token, list)
	}
	var list *Ast
	expect_token(f, TokenOpenParen)
	list = parse_field_list(f, nil, uint32(FieldFlagResults), TokenCloseParen, true, false)
	if file_allow_newline(f) {
		skip_possible_newline(f)
	}
	expect_token_after(f, TokenCloseParen, "parameter list")
	return list
}

func string_to_calling_convention(s string) ProcCallingConvention {
	switch s {
	case "odin":
		return ProcCCOdin
	case "contextless":
		return ProcCCContextless
	case "cdecl", "c":
		return ProcCCCDecl
	case "stdcall", "std":
		return ProcCCStdCall
	case "fastcall", "fast":
		return ProcCCFastCall
	case "none":
		return ProcCCNone
	case "naked":
		return ProcCCNaked
	case "win64":
		return ProcCCWin64
	case "sysv":
		return ProcCCSysV
	case "preserve/none":
		return ProcCCPreserveNone
	case "preserve/most":
		return ProcCCPreserveMost
	case "preserve/all":
		return ProcCCPreserveAll
	case "system":
		if buildContext.Metrics.Os == TargetOsWindows {
			return ProcCCStdCall
		}
		return ProcCCCDecl
	}
	return ProcCCInvalid
}

func parse_proc_type(f *AstFile, proc_token Token) *Ast {
	var params *Ast
	var results *Ast
	diverging := false
	cc := ProcCCInvalid
	if f.CurrToken.Kind == TokenString {
		token := expect_token(f, TokenString)
		c := string_to_calling_convention(goStr(token.String))
		if c == ProcCCInvalid {
			syntax_error(token.Kind, "Unknown procedure calling convention")
		} else {
			cc = c
		}
	}
	if cc == ProcCCInvalid {
		if f.InForeignBlock {
			cc = ProcCCForeignBlockDefault
		} else {
			cc = default_calling_convention()
		}
	}
	expect_token(f, TokenOpenParen)
	f.ExprLevel++
	params = parse_field_list(f, nil, uint32(FieldFlagSignature), TokenCloseParen, true, true)
	if file_allow_newline(f) {
		skip_possible_newline(f)
	}
	f.ExprLevel--
	expect_token_after(f, TokenCloseParen, "parameter list")
	results = parse_results(f, &diverging)
	tags := uint64(0)
	is_generic := false
	if params != nil && params.Kind == AstFieldList {
		for _, param := range params.FieldList.List {
			if param.Kind != AstField {
				continue
			}
			field := param.Field
			if field.Type != nil {
				if field.Type.Kind == AstPolyType {
					is_generic = true
					break
				}
				for _, name := range field.Names {
					if name.Kind == AstPolyType {
						is_generic = true
						break
					}
				}
				if is_generic {
					break
				}
			}
		}
	}
	return ast_proc_type(f, proc_token, params, results, tags, cc, is_generic, diverging)
}

func parse_var_type(f *AstFile, allow_ellipsis bool, allow_typeid_token bool) *Ast {
	if allow_ellipsis && f.CurrToken.Kind == TokenEllipsis {
		tok := advance_token(f)
		typ := parse_type_or_ident(f)
		if typ == nil {
			syntax_error(tok.Kind, "variadic field missing type after '..'")
			typ = ast_bad_expr(f, tok, f.CurrToken)
		}
		return ast_ellipsis(f, tok, typ)
	}
	var typ *Ast
	if allow_typeid_token && f.CurrToken.Kind == TokenTypeid {
		token := expect_token(f, TokenTypeid)
		var specialization *Ast
		if allow_token(f, TokenQuo) {
			specialization = parse_type(f)
		}
		typ = ast_typeid_type(f, token, specialization)
	} else {
		typ = parse_type(f)
	}
	return typ
}

type AstAndFlags struct {
	Node  *Ast
	Flags uint32
}

var parse_field_prefix_mappings = []struct {
	Name      string
	TokenKind TokenKind
	Flag      FieldFlag
}{
	{"using", TokenUsing, FieldFlagUsing},
	{"no_alias", TokenHash, FieldFlagNoAlias},
	{"no_capture", TokenHash, FieldFlagNoCapture},
	{"c_vararg", TokenHash, FieldFlagCVararg},
	{"const", TokenHash, FieldFlagConst},
	{"any_int", TokenHash, FieldFlagAnyInt},
	{"subtype", TokenHash, FieldFlagSubtype},
	{"by_ptr", TokenHash, FieldFlagByPtr},
	{"no_broadcast", TokenHash, FieldFlagNoBroadcast},
}

func is_token_field_prefix(f *AstFile) FieldFlag {
	switch f.CurrToken.Kind {
	case TokenEOF:
		return FieldFlagInvalid
	case TokenUsing:
		return FieldFlagUsing
	case TokenHash:
		advance_token(f)
		if f.CurrToken.Kind == TokenIdent {
			for _, m := range parse_field_prefix_mappings {
				if m.TokenKind == TokenHash && goStr(f.CurrToken.String) == m.Name {
					return m.Flag
				}
			}
		}
		return FieldFlagUnknown
	}
	return FieldFlagInvalid
}

func parse_field_prefixes(f *AstFile) uint32 {
	counts := make([]int, len(parse_field_prefix_mappings))
	for {
		flag := is_token_field_prefix(f)
		if flag&FieldFlagInvalid != 0 {
			break
		}
		if flag&FieldFlagUnknown != 0 {
			syntax_error(f.CurrToken.Kind, "Unknown prefix kind")
			advance_token(f)
			continue
		}
		for i, m := range parse_field_prefix_mappings {
			if m.Flag == flag {
				counts[i]++
				advance_token(f)
				break
			}
		}
	}
	field_flags := uint32(0)
	for i, m := range parse_field_prefix_mappings {
		if counts[i] > 0 {
			field_flags |= uint32(m.Flag)
			if counts[i] != 1 {
				prefix := ""
				if m.TokenKind == TokenHash {
					prefix = "#"
				}
				syntax_error(f.CurrToken.Kind, "Multiple '"+prefix+m.Name+"' in this field list")
			}
		}
	}
	return field_flags
}

func check_field_prefixes(f *AstFile, name_count isize, allowed_flags uint32, set_flags uint32) uint32 {
	for _, m := range parse_field_prefix_mappings {
		err := false
		if set_flags&uint32(m.Flag) != 0 {
			if m.Flag == FieldFlagUsing && name_count > 1 {
				err = true
				syntax_error(f.CurrToken.Kind, "Cannot apply 'using' to more than one of the same type")
			}
			if allowed_flags&uint32(m.Flag) == 0 {
				err = true
				prefix := ""
				if m.TokenKind == TokenHash {
					prefix = "#"
				}
				syntax_error(f.CurrToken.Kind, "'"+prefix+m.Name+"' in not allowed within this field list")
			}
		}
		if err {
			set_flags &^= uint32(m.Flag)
		}
	}
	return set_flags
}

func convert_to_ident_list(f *AstFile, list []AstAndFlags, ignore_flags bool, allow_poly_names bool) []*Ast {
	var idents []*Ast
	for i, item := range list {
		ident := item.Node
		if !ignore_flags && i != 0 {
			syntax_error(ident.Kind, "Illegal use of prefixes in parameter list")
		}
		switch ident.Kind {
		case AstIdent, AstBadExpr:
		case AstImplicit:
			syntax_error(ident.Kind, "Expected an identifier, keyword used")
			ident = ast_ident(f, BlankToken)
		case AstPolyType:
			if allow_poly_names {
				if ident.PolyType.Specialization == nil {
					break
				}
				syntax_error(ident.Kind, "Expected a polymorphic identifier without any specialization")
			} else {
				syntax_error(ident.Kind, "Expected a non-polymorphic identifier")
			}
		default:
			syntax_error(ident.Kind, "Expected an identifier")
			ident = ast_ident(f, BlankToken)
		}
		idents = append(idents, ident)
	}
	return idents
}

func check_procedure_name_list(names []*Ast) bool {
	if len(names) == 0 {
		return false
	}
	first_is_polymorphic := names[0].Kind == AstPolyType
	any_polymorphic_names := first_is_polymorphic
	for i := 1; i < len(names); i++ {
		name := names[i]
		if first_is_polymorphic {
			if name.Kind == AstPolyType {
				any_polymorphic_names = true
			} else {
				syntax_error(name.Kind, "Mixture of polymorphic and non-polymorphic identifiers")
				return any_polymorphic_names
			}
		} else {
			if name.Kind == AstPolyType {
				any_polymorphic_names = true
				syntax_error(name.Kind, "Mixture of polymorphic and non-polymorphic identifiers")
				return any_polymorphic_names
			}
		}
	}
	return any_polymorphic_names
}

func parse_field_list(f *AstFile, name_count_ *isize, allowed_flags uint32, follow TokenKind, allow_default_parameters bool, allow_typeid_token bool) *Ast {
	prev_allow_newline := f.AllowNewline
	defer func() { f.AllowNewline = prev_allow_newline }()
	f.AllowNewline = file_allow_newline(f)
	start_token := f.CurrToken
	docs := f.LeadComment
	var params []*Ast
	var list []AstAndFlags
	allow_poly_names := allow_typeid_token
	total_name_count := isize(0)
	allow_ellipsis := allowed_flags&uint32(FieldFlagEllipsis) != 0
	seen_ellipsis := false
	is_signature := allowed_flags&uint32(FieldFlagSignature) == uint32(FieldFlagSignature)

	for f.CurrToken.Kind != follow &&
		f.CurrToken.Kind != TokenColon &&
		f.CurrToken.Kind != TokenEOF {
		if !is_signature {
			parse_enforce_tabs(f)
		}
		flags := parse_field_prefixes(f)
		param := parse_var_type(f, allow_ellipsis, allow_typeid_token)
		if param.Kind == AstEllipsis {
			if seen_ellipsis {
				syntax_error(param.Kind, "Extra variadic parameter after ellipsis")
			}
			seen_ellipsis = true
		} else if seen_ellipsis {
			syntax_error(param.Kind, "Extra parameter after ellipsis")
		}
		list = append(list, AstAndFlags{Node: param, Flags: flags})
		if !allow_field_separator(f) {
			break
		}
	}

	if f.CurrToken.Kind != TokenColon {
		for _, item := range list {
			typ := item.Node
			token := Token{Kind: TokenIdent}
			if allowed_flags&uint32(FieldFlagResults) != 0 {
				token.String = String{Data: strData(""), Len: 0}
			}
			var names []*Ast = make([]*Ast, 1)
			token.Pos = ast_token(typ).Pos
			names[0] = ast_ident(f, token)
			flags := check_field_prefixes(f, isize(len(list)), allowed_flags, item.Flags)
			tag := Token{}
			param := ast_field(f, names, item.Node, nil, flags, tag, docs, f.LineComment)
			params = append(params, param)
		}
		if name_count_ != nil {
			*name_count_ = total_name_count
		}
		return ast_field_list(f, start_token, params)
	}

	if f.PrevToken.Kind == TokenComma {
		syntax_error(f.PrevToken.Kind, "Trailing comma before a colon is not allowed")
	}

	names := convert_to_ident_list(f, list, true, allow_poly_names)
	if len(names) == 0 {
		syntax_error(f.CurrToken.Kind, "Empty field declaration")
	}
	any_polymorphic_names := check_procedure_name_list(names)
	set_flags := uint32(0)
	if len(list) > 0 {
		set_flags = list[0].Flags
	}
	set_flags = check_field_prefixes(f, isize(len(names)), allowed_flags, set_flags)
	total_name_count += isize(len(names))
	var typ *Ast
	var default_value *Ast
	tag := Token{}
	expect_token_after(f, TokenColon, "field list")
	if f.CurrToken.Kind != TokenEq {
		typ = parse_var_type(f, allow_ellipsis, allow_typeid_token)
		tt := unparen_expr(typ)
		if tt == nil {
			syntax_error(f.PrevToken.Kind, "Invalid type expression in field list")
		} else if is_signature && !any_polymorphic_names && tt.Kind == AstTypeidType && tt.TypeidType.Specialization != nil {
			syntax_error(typ.Kind, "Specialization of typeid is not allowed without polymorphic names")
		}
	}
	if allow_token(f, TokenEq) {
		default_value = parse_expr(f, false)
		if !allow_default_parameters {
			syntax_error(f.CurrToken.Kind, "Default parameters are only allowed for procedures")
			default_value = nil
		}
	}
	if default_value != nil && len(names) > 1 {
		syntax_error(f.CurrToken.Kind, "Default parameters can only be applied to single values")
	}
	if allowed_flags == uint32(FieldFlagStruct) && default_value != nil {
		syntax_error(default_value.Kind, "Default parameters are not allowed for structs")
		default_value = nil
	}
	if typ != nil && typ.Kind == AstEllipsis {
		if seen_ellipsis {
			syntax_error(typ.Kind, "Extra variadic parameter after ellipsis")
		}
		seen_ellipsis = true
		if len(names) != 1 {
			syntax_error(typ.Kind, "Variadic parameters can only have one field name")
		}
	} else if seen_ellipsis && default_value == nil {
		syntax_error(f.CurrToken.Kind, "Extra parameter after ellipsis without a default value")
	}
	if typ != nil && default_value == nil {
		if f.CurrToken.Kind == TokenString {
			tag = expect_token(f, TokenString)
			if allowed_flags&uint32(FieldFlagTags) == 0 {
				syntax_error(tag.Kind, "Field tags are only allowed within structures")
			}
		}
	}
	more_fields := allow_field_separator(f)
	param := ast_field(f, names, typ, default_value, set_flags, tag, docs, f.LineComment)
	params = append(params, param)
	if !more_fields {
		if name_count_ != nil {
			*name_count_ = total_name_count
		}
		return ast_field_list(f, start_token, params)
	}

	for f.CurrToken.Kind != follow &&
		f.CurrToken.Kind != TokenEOF &&
		f.CurrToken.Kind != TokenSemicolon {
		docs = f.LeadComment
		if !is_signature {
			parse_enforce_tabs(f)
		}
		set_flags = parse_field_prefixes(f)
		tag = Token{}
		names = parse_ident_list(f, allow_poly_names)
		if len(names) == 0 {
			syntax_error(f.CurrToken.Kind, "Empty field declaration")
			break
		}
		any_polymorphic_names = check_procedure_name_list(names)
		set_flags = check_field_prefixes(f, isize(len(names)), allowed_flags, set_flags)
		total_name_count += isize(len(names))
		typ = nil
		default_value = nil
		expect_token_after(f, TokenColon, "field list")
		if f.CurrToken.Kind != TokenEq {
			typ = parse_var_type(f, allow_ellipsis, allow_typeid_token)
			tt := unparen_expr(typ)
			if is_signature && !any_polymorphic_names &&
				tt != nil &&
				tt.Kind == AstTypeidType && tt.TypeidType.Specialization != nil {
				syntax_error(typ.Kind, "Specialization of typeid is not allowed without polymorphic names")
			}
		}
		if allow_token(f, TokenEq) {
			default_value = parse_expr(f, false)
			if !allow_default_parameters {
				syntax_error(f.CurrToken.Kind, "Default parameters are only allowed for procedures")
				default_value = nil
			}
		}
		if default_value != nil && len(names) > 1 {
			syntax_error(f.CurrToken.Kind, "Default parameters can only be applied to single values")
		}
		if typ != nil && typ.Kind == AstEllipsis {
			if seen_ellipsis {
				syntax_error(typ.Kind, "Extra variadic parameter after ellipsis")
			}
			seen_ellipsis = true
			if len(names) != 1 {
				syntax_error(typ.Kind, "Variadic parameters can only have one field name")
			}
		} else if seen_ellipsis && default_value == nil {
			syntax_error(f.CurrToken.Kind, "Extra parameter after ellipsis without a default value")
		}
		if typ != nil && default_value == nil {
			if f.CurrToken.Kind == TokenString {
				tag = expect_token(f, TokenString)
				if allowed_flags&uint32(FieldFlagTags) == 0 {
					syntax_error(tag.Kind, "Field tags are only allowed within structures")
				}
			}
		}
		ok := allow_field_separator(f)
		param = ast_field(f, names, typ, default_value, set_flags, tag, docs, f.LineComment)
		params = append(params, param)
		if !ok {
			break
		}
	}
	if name_count_ != nil {
		*name_count_ = total_name_count
	}
	return ast_field_list(f, start_token, params)
}

func parse_struct_field_list(f *AstFile, name_count_ *isize) *Ast {
	start_token := f.CurrToken
	total_name_count := isize(0)
	params := parse_field_list(f, &total_name_count, uint32(FieldFlagStruct), TokenCloseBrace, false, false)
	if name_count_ != nil {
		*name_count_ = total_name_count
	}
	_ = start_token
	return params
}

func parse_value_decl(f *AstFile, names []*Ast, docs *CommentGroup) *Ast {
	is_mutable := true
	var values []*Ast
	typ := parse_type_or_ident(f)
	if f.CurrToken.Kind == TokenEq || f.CurrToken.Kind == TokenColon {
		if !is_mutable {
			expect_token_after(f, TokenColon, "type")
		} else {
			sep := advance_token(f)
			is_mutable = sep.Kind != TokenColon
		}
		values = parse_rhs_expr_list(f)
		if len(values) > len(names) {
			syntax_error(f.CurrToken.Kind, "Too many values on the right hand side of the declaration")
		} else if len(values) < len(names) && !is_mutable {
			syntax_error(f.CurrToken.Kind, "All constant declarations must be defined")
		} else if len(values) == 0 {
			syntax_error(f.CurrToken.Kind, "Expected an expression for this declaration")
		}
	}
	if is_mutable {
		if typ == nil && len(values) == 0 {
			syntax_error(f.CurrToken.Kind, "Missing variable type or initialization")
			return ast_bad_decl(f, f.CurrToken, f.CurrToken)
		}
	} else {
		if typ == nil && len(values) == 0 && len(names) > 0 {
			syntax_error(f.CurrToken.Kind, "Missing constant value")
			return ast_bad_decl(f, f.CurrToken, f.CurrToken)
		}
	}
	if values == nil {
		values = make([]*Ast, 0)
	}
	end_comment := f.LeadComment
	if f.ExprLevel >= 0 {
		if f.CurrToken.Kind == TokenCloseBrace &&
			f.CurrToken.Pos.Line == f.PrevToken.Pos.Line {
		} else {
			expect_semicolon(f)
		}
	}
	if f.CurrProc == nil {
		if len(values) > 0 && len(names) != len(values) {
			syntax_error(values[0].Kind, "Expected N expressions on the right hand side, got M")
		}
	}
	return ast_value_decl(f, names, typ, values, is_mutable, docs, end_comment)
}

func parse_simple_stmt(f *AstFile, flags StmtAllowFlag) *Ast {
	token := f.CurrToken
	docs := f.LeadComment
	lhs := parse_lhs_expr_list(f)
	token = f.CurrToken
	switch token.Kind {
	case TokenEq, TokenAddEq, TokenSubEq, TokenMulEq, TokenQuoEq,
		TokenModEq, TokenModModEq, TokenAndEq, TokenOrEq, TokenXorEq,
		TokenShlEq, TokenShrEq, TokenAndNotEq, TokenCmpAndEq, TokenCmpOrEq:
		if f.CurrProc == nil {
			syntax_error(f.CurrToken.Kind, "You cannot use a simple statement in the file scope")
			return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
		}
		advance_token(f)
		rhs := parse_rhs_expr_list(f)
		if len(rhs) == 0 {
			syntax_error(token.Kind, "No right-hand side in assignment statement.")
			return ast_bad_stmt(f, token, f.CurrToken)
		}
		return ast_assign_stmt(f, token, lhs, rhs)
	case TokenIn:
		if flags&StmtAllowFlagIn != 0 {
			allow_token(f, TokenIn)
			prev_allow_range := f.AllowRange
			f.AllowRange = true
			expr := parse_expr(f, true)
			f.AllowRange = prev_allow_range
			rhs := make([]*Ast, 0, 1)
			rhs = append(rhs, expr)
			return ast_assign_stmt(f, token, lhs, rhs)
		}
	case TokenColon:
		expect_token_after(f, TokenColon, "identifier list")
		if flags&StmtAllowFlagLabel != 0 && len(lhs) == 1 {
			is_partial := false
			is_reverse := false
			partial_token := Token{}
			if f.CurrToken.Kind == TokenHash {
				name := peek_token_n(f, 0)
				if name.Kind == TokenIdent && goStr(name.String) == "partial" &&
					peek_token_n(f, 1).Kind == TokenSwitch {
					partial_token = expect_token(f, TokenHash)
					expect_token(f, TokenIdent)
					is_partial = true
				} else if name.Kind == TokenIdent && goStr(name.String) == "reverse" &&
					peek_token_n(f, 1).Kind == TokenFor {
					partial_token = expect_token(f, TokenHash)
					expect_token(f, TokenIdent)
					is_reverse = true
				}
			}
			switch f.CurrToken.Kind {
			case TokenOpenBrace, TokenIf, TokenFor, TokenSwitch:
				name := lhs[0]
				label := ast_label_decl(f, ast_token(name), name)
				stmt := parse_stmt(f)
				switch stmt.Kind {
				case AstBlockStmt:
					stmt.BlockStmt.Label = label
				case AstIfStmt:
					stmt.IfStmt.Label = label
				case AstForStmt:
					stmt.ForStmt.Label = label
				case AstRangeStmt:
					stmt.RangeStmt.Label = label
				case AstSwitchStmt:
					stmt.SwitchStmt.Label = label
				case AstTypeSwitchStmt:
					stmt.TypeSwitchStmt.Label = label
				default:
					syntax_error(token.Kind, "Labels can only be applied to a loop or switch statement")
				}
				if is_partial {
					switch stmt.Kind {
					case AstSwitchStmt:
						stmt.SwitchStmt.Partial = true
					case AstTypeSwitchStmt:
						stmt.TypeSwitchStmt.Partial = true
					default:
						syntax_error(partial_token.Kind, "Incorrect use of directive")
					}
				} else if is_reverse {
					switch stmt.Kind {
					case AstRangeStmt:
						if stmt.RangeStmt.Reverse {
							syntax_error(token.Kind, "#reverse already applied to a 'for in' statement")
						}
						stmt.RangeStmt.Reverse = true
					default:
						syntax_error(token.Kind, "#reverse can only be applied to a 'for in' statement")
					}
				}
				return stmt
			}
		}
		return parse_value_decl(f, lhs, docs)
	}
	if len(lhs) > 1 {
		syntax_error(token.Kind, "Expected 1 expression")
		return ast_bad_stmt(f, token, f.CurrToken)
	}
	switch token.Kind {
	case TokenIncrement, TokenDecrement:
		advance_token(f)
		syntax_error(token.Kind, "Postfix '++'/'--' statement is not supported")
	}
	return ast_expr_stmt(f, lhs[0])
}

func parse_block_stmt(f *AstFile, is_when bool) *Ast {
	skip_possible_newline_for_literal(f)
	if !is_when && f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use a block statement in the file scope")
		return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
	}
	return parse_body(f)
}

func parse_body(f *AstFile) *Ast {
	prev_expr_level := f.ExprLevel
	prev_newline := f.AllowNewline
	f.ExprLevel = 0
	open := expect_token(f, TokenOpenBrace)
	stmts := parse_stmt_list(f)
	close := expect_token(f, TokenCloseBrace)
	f.ExprLevel = prev_expr_level
	f.AllowNewline = prev_newline
	return ast_block_stmt(f, stmts, open, close)
}

func parse_do_body(f *AstFile, token Token, msg string) *Ast {
	prev_expr_level := f.ExprLevel
	prev_newline := f.AllowNewline
	f.ExprLevel = 0
	f.AllowNewline = false
	body := convert_stmt_to_body(f, parse_stmt(f))
	if buildContext.DisallowDo {
		syntax_error(body.Kind, "'do' has been disallowed")
	} else if token.Pos.FileID != 0 && !ast_on_same_line(token, body) {
		syntax_error(body.Kind, "The body of a 'do' must be on the same line as "+msg)
	}
	f.ExprLevel = prev_expr_level
	f.AllowNewline = prev_newline
	return body
}

func parse_control_statement_semicolon_separator(f *AstFile) bool {
	tok := peek_token(f)
	if tok.Kind != TokenOpenBrace {
		if f.CurrToken.Kind == TokenSemicolon && goStr(f.CurrToken.String) != ";" {
			syntax_error(f.CurrToken.Kind, "Expected ';', got newline")
		}
		return allow_token(f, TokenSemicolon)
	}
	if goStr(f.CurrToken.String) == ";" {
		return allow_token(f, TokenSemicolon)
	}
	return false
}

func parse_if_stmt(f *AstFile) *Ast {
	if f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use an if statement in the file scope")
		return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
	}
	var top_if_stmt *Ast
	var prev_if_stmt *Ast
if_else_chain:
	token := expect_token(f, TokenIf)
	var init *Ast
	var cond *Ast
	var body *Ast
	var else_stmt *Ast
	prev_level := f.ExprLevel
	f.ExprLevel = -1
	prev_allow_in_expr := f.AllowInExpr
	f.AllowInExpr = true
	if allow_token(f, TokenSemicolon) {
		cond = parse_expr(f, false)
	} else {
		init = parse_simple_stmt(f, StmtAllowFlagNone)
		if parse_control_statement_semicolon_separator(f) {
			cond = parse_expr(f, false)
		} else {
			cond = convert_stmt_to_expr(f, init, "boolean expression")
			init = nil
		}
	}
	f.ExprLevel = prev_level
	f.AllowInExpr = prev_allow_in_expr
	if cond == nil {
		syntax_error(f.CurrToken.Kind, "Expected condition for if statement")
	}
	if allow_token(f, TokenDo) {
		body = parse_do_body(f, ast_token(cond), "the if statement")
	} else {
		body = parse_block_stmt(f, false)
	}
	ignore_strict_style := false
	if token.Pos.Line == ast_end_token(body).Pos.Line {
		ignore_strict_style = true
	}
	skip_possible_newline_for_literal(f, ignore_strict_style)
	curr_if_stmt := ast_if_stmt(f, token, init, cond, body, nil)
	if top_if_stmt == nil {
		top_if_stmt = curr_if_stmt
	}
	if prev_if_stmt != nil {
		prev_if_stmt.IfStmt.ElseStmt = curr_if_stmt
	}
	if f.CurrToken.Kind == TokenElse {
		else_token := expect_token(f, TokenElse)
		switch f.CurrToken.Kind {
		case TokenIf:
			prev_if_stmt = curr_if_stmt
			goto if_else_chain
		case TokenOpenBrace:
			else_stmt = parse_block_stmt(f, false)
		case TokenDo:
			expect_token(f, TokenDo)
			else_stmt = parse_do_body(f, else_token, "'else'")
		default:
			syntax_error(f.CurrToken.Kind, "Expected if statement block statement")
			else_stmt = ast_bad_stmt(f, f.CurrToken, f.CurrToken)
		}
	}
	curr_if_stmt.IfStmt.ElseStmt = else_stmt
	return top_if_stmt
}

func parse_when_stmt(f *AstFile) *Ast {
	token := expect_token(f, TokenWhen)
	var cond *Ast
	var body *Ast
	var else_stmt *Ast
	prev_level := f.ExprLevel
	f.ExprLevel = -1
	prev_allow_in_expr := f.AllowInExpr
	f.AllowInExpr = true
	cond = parse_expr(f, false)
	f.AllowInExpr = prev_allow_in_expr
	f.ExprLevel = prev_level
	if cond == nil {
		syntax_error(f.CurrToken.Kind, "Expected condition for when statement")
	}
	was_in_when_stmt := f.InWhenStatement
	f.InWhenStatement = true
	if allow_token(f, TokenDo) {
		body = parse_do_body(f, ast_token(cond), "the when statement")
	} else {
		body = parse_block_stmt(f, true)
	}
	ignore_strict_style := false
	if token.Pos.Line == ast_end_token(body).Pos.Line {
		ignore_strict_style = true
	}
	skip_possible_newline_for_literal(f, ignore_strict_style)
	if f.CurrToken.Kind == TokenElse {
		else_token := expect_token(f, TokenElse)
		switch f.CurrToken.Kind {
		case TokenWhen:
			else_stmt = parse_when_stmt(f)
		case TokenOpenBrace:
			else_stmt = parse_block_stmt(f, true)
		case TokenDo:
			expect_token(f, TokenDo)
			else_stmt = parse_do_body(f, else_token, "'else'")
		default:
			syntax_error(f.CurrToken.Kind, "Expected when statement block statement")
			else_stmt = ast_bad_stmt(f, f.CurrToken, f.CurrToken)
		}
	}
	f.InWhenStatement = was_in_when_stmt
	return ast_when_stmt(f, token, cond, body, else_stmt)
}

func parse_return_stmt(f *AstFile) *Ast {
	token := expect_token(f, TokenReturn)
	if f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use a return statement in the file scope")
		return ast_bad_stmt(f, token, f.CurrToken)
	}
	if f.ExprLevel > 0 {
		syntax_error(f.CurrToken.Kind, "You cannot use a return statement within an expression")
		return ast_bad_stmt(f, token, f.CurrToken)
	}
	var results []*Ast
	for f.CurrToken.Kind != TokenSemicolon && f.CurrToken.Kind != TokenCloseBrace {
		arg := parse_expr(f, false)
		results = append(results, arg)
		if f.CurrToken.Kind != TokenComma || f.CurrToken.Kind == TokenEOF {
			break
		}
		advance_token(f)
	}
	expect_semicolon(f)
	return ast_return_stmt(f, token, results)
}

func parse_for_stmt(f *AstFile) *Ast {
	if f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use a for statement in the file scope")
		return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
	}
	token := expect_token(f, TokenFor)
	var init *Ast
	var cond *Ast
	var post *Ast
	var body *Ast
	is_range := false

	if f.CurrToken.Kind != TokenOpenBrace &&
		f.CurrToken.Kind != TokenDo {
		prev_level := f.ExprLevel
		f.ExprLevel = -1
		if f.CurrToken.Kind == TokenIn {
			in_token := expect_token(f, TokenIn)
			syntax_error(in_token.Kind, "Prefer 'for _ in' over 'for in'")
			var rhs *Ast
			prev_allow_range := f.AllowRange
			f.AllowRange = true
			rhs = parse_expr(f, false)
			f.AllowRange = prev_allow_range
			if allow_token(f, TokenDo) {
				body = parse_do_body(f, token, "the for statement")
			} else {
				body = parse_block_stmt(f, false)
			}
			f.ExprLevel = prev_level
			return ast_range_stmt(f, token, init, nil, in_token, rhs, body)
		}
		if f.CurrToken.Kind != TokenSemicolon {
			cond = parse_simple_stmt(f, StmtAllowFlagIn)
			if cond.Kind == AstAssignStmt && cond.AssignStmt.Op.Kind == TokenIn {
				is_range = true
			}
		}
		if !is_range && parse_control_statement_semicolon_separator(f) {
			init = cond
			cond = nil
			if f.CurrToken.Kind == TokenOpenBrace || f.CurrToken.Kind == TokenDo {
				syntax_error(f.CurrToken.Kind, "Expected ';' followed by conditions")
			} else {
				if f.CurrToken.Kind != TokenSemicolon {
					if f.CurrToken.Kind == TokenIdent {
						next_token := peek_token(f)
						if next_token.Kind == TokenIn || next_token.Kind == TokenComma {
							cond = parse_simple_stmt(f, StmtAllowFlagIn)
							if cond.Kind == AstAssignStmt && cond.AssignStmt.Op.Kind == TokenIn {
								is_range = true
							}
							goto range_skip
						}
					}
					cond = parse_simple_stmt(f, StmtAllowFlagNone)
				}
				if goStr(f.CurrToken.String) != ";" {
					syntax_error(f.CurrToken.Kind, "Expected ';'")
				} else {
					expect_token(f, TokenSemicolon)
				}
				if f.CurrToken.Kind != TokenOpenBrace &&
					f.CurrToken.Kind != TokenDo {
					post = parse_simple_stmt(f, StmtAllowFlagNone)
				}
			}
		}
		f.ExprLevel = prev_level
	}
range_skip:
	if allow_token(f, TokenDo) {
		body = parse_do_body(f, token, "the for statement")
	} else {
		body = parse_block_stmt(f, false)
	}
	if is_range && cond != nil && cond.Kind == AstAssignStmt {
		in_token := cond.AssignStmt.Op
		vals := cond.AssignStmt.LHS
		var rhs *Ast
		if len(cond.AssignStmt.RHS) > 0 {
			rhs = cond.AssignStmt.RHS[0]
		}
		return ast_range_stmt(f, token, init, vals, in_token, rhs, body)
	}
	cond = convert_stmt_to_expr(f, cond, "boolean expression")
	if init != nil &&
		cond == nil &&
		post == nil {
		syntax_error(init.Kind, "'for init; ; {' without an explicit condition nor post statement is not allowed")
	}
	return ast_for_stmt(f, token, init, cond, post, body)
}

func parse_case_clause(f *AstFile, is_type bool) *Ast {
	token := f.CurrToken
	var list []*Ast
	expect_token(f, TokenCase)
	prev_allow_range := f.AllowRange
	prev_allow_in_expr := f.AllowInExpr
	f.AllowRange = !is_type
	f.AllowInExpr = !is_type
	if f.CurrToken.Kind != TokenColon {
		list = parse_rhs_expr_list(f)
	}
	f.AllowRange = prev_allow_range
	f.AllowInExpr = prev_allow_in_expr
	expect_token(f, TokenColon)
	stmts := parse_stmt_list(f)
	return ast_case_clause(f, token, list, stmts)
}

func parse_switch_stmt(f *AstFile) *Ast {
	if f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use a switch statement in the file scope")
		return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
	}
	token := expect_token(f, TokenSwitch)
	var init *Ast
	var tag *Ast
	var body *Ast
	open := Token{}
	close := Token{}
	is_type_switch := false
	var list []*Ast

	if f.CurrToken.Kind != TokenOpenBrace {
		prev_level := f.ExprLevel
		f.ExprLevel = -1
		if f.CurrToken.Kind == TokenIn {
			in_token := expect_token(f, TokenIn)
			syntax_error(in_token.Kind, "Prefer 'switch _ in' over 'switch in'")
			var lhs []*Ast = make([]*Ast, 0, 1)
			var rhs []*Ast = make([]*Ast, 0, 1)
		blank_ident := BlankToken
			blank := ast_ident(f, blank_ident)
			lhs = append(lhs, blank)
			rhs = append(rhs, parse_expr(f, true))
			tag = ast_assign_stmt(f, token, lhs, rhs)
			is_type_switch = true
		} else {
			tag = parse_simple_stmt(f, StmtAllowFlagIn)
			if tag.Kind == AstAssignStmt && tag.AssignStmt.Op.Kind == TokenIn {
				is_type_switch = true
			} else if parse_control_statement_semicolon_separator(f) {
				init = tag
				tag = nil
				if f.CurrToken.Kind != TokenOpenBrace {
					tag = parse_simple_stmt(f, StmtAllowFlagNone)
				}
			}
		}
		f.ExprLevel = prev_level
	}
	skip_possible_newline(f)
	open = expect_token(f, TokenOpenBrace)
	for f.CurrToken.Kind == TokenCase {
		list = append(list, parse_case_clause(f, is_type_switch))
	}
	close = expect_token(f, TokenCloseBrace)
	body = ast_block_stmt(f, list, open, close)
	if is_type_switch {
		return ast_type_switch_stmt(f, token, tag, body)
	}
	tag = convert_stmt_to_expr(f, tag, "switch expression")
	return ast_switch_stmt(f, token, init, tag, body)
}

func parse_defer_stmt(f *AstFile) *Ast {
	if f.CurrProc == nil {
		syntax_error(f.CurrToken.Kind, "You cannot use a defer statement in the file scope")
		return ast_bad_stmt(f, f.CurrToken, f.CurrToken)
	}
	token := expect_token(f, TokenDefer)
	stmt := parse_stmt(f)
	switch stmt.Kind {
	case AstEmptyStmt:
		syntax_error(token.Kind, "Empty statement after defer (e.g. ';')")
	case AstDeferStmt:
		syntax_error(token.Kind, "You cannot defer a defer statement")
		stmt = stmt.DeferStmt.Stmt
	case AstReturnStmt:
		syntax_error(token.Kind, "You cannot defer a return statement")
	}
	return ast_defer_stmt(f, token, stmt)
}

func parse_import_decl(f *AstFile, kind int) *Ast {
	docs := f.LeadComment
	token := expect_token(f, TokenImport)
	import_name := Token{}
	switch f.CurrToken.Kind {
	case TokenIdent:
		import_name = advance_token(f)
	default:
		import_name.Pos = f.CurrToken.Pos
	}
	file_path := expect_token_after(f, TokenString, "import")
	var s *Ast
	if f.CurrProc != nil {
		syntax_error(import_name.Kind, "Cannot use 'import' within a procedure. This must be done at the file scope")
		s = ast_bad_decl(f, import_name, file_path)
	} else {
		s = ast_import_decl(f, token, file_path, import_name, docs, f.LineComment)
		f.Imports = append(f.Imports, s)
	}
	if f.InWhenStatement {
		syntax_error(import_name.Kind, "Cannot use 'import' within a 'when' statement.")
	}
	if kind != 0 {
		syntax_error(import_name.Kind, "'using import' is not allowed, please use the import name explicitly")
	}
	expect_semicolon(f)
	return s
}

func parse_foreign_block_decl(f *AstFile, decls *[]*Ast) {
	decl := parse_stmt(f)
	switch decl.Kind {
	case AstEmptyStmt, AstBadStmt, AstBadDecl:
		return
	case AstWhenStmt, AstValueDecl:
		*decls = append(*decls, decl)
		return
	default:
		syntax_error(decl.Kind, "Foreign blocks only allow procedure and variable declarations")
		return
	}
}

func parse_foreign_block(f *AstFile, token Token) *Ast {
	docs := f.LeadComment
	var foreign_library *Ast
	if f.CurrToken.Kind == TokenOpenBrace {
		foreign_library = ast_ident(f, BlankToken)
	} else {
		foreign_library = parse_ident(f)
	}
	var decls []*Ast
	prev_in_foreign_block := f.InForeignBlock
	f.InForeignBlock = true
	defer func() { f.InForeignBlock = prev_in_foreign_block }()
	skip_possible_newline_for_literal(f)
	open := expect_token(f, TokenOpenBrace)
	for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
		parse_foreign_block_decl(f, &decls)
	}
	close := expect_token(f, TokenCloseBrace)
	body := ast_block_stmt(f, decls, open, close)
	decl := ast_foreign_block_decl(f, token, foreign_library, body, docs)
	expect_semicolon(f)
	return decl
}

func parse_foreign_decl(f *AstFile) *Ast {
	docs := f.LeadComment
	_ = docs
	token := expect_token(f, TokenForeign)
	switch f.CurrToken.Kind {
	case TokenIdent, TokenOpenBrace:
		return parse_foreign_block(f, token)
	case TokenImport:
		import_token := expect_token(f, TokenImport)
		lib_name := Token{}
		switch f.CurrToken.Kind {
		case TokenIdent:
			lib_name = advance_token(f)
		default:
			lib_name.Pos = token.Pos
		}
		if is_blank_ident(lib_name) {
			syntax_error(lib_name.Kind, "Illegal foreign import name: '_'")
		}
		multiple_filepaths := false
		var filepaths []*Ast
		if allow_token(f, TokenOpenBrace) {
			multiple_filepaths = true
			filepaths = make([]*Ast, 0)
			for f.CurrToken.Kind != TokenCloseBrace && f.CurrToken.Kind != TokenEOF {
				path := parse_expr(f, false)
				filepaths = append(filepaths, path)
				if !allow_field_separator(f) {
					break
				}
			}
			expect_closing(f, TokenCloseBrace, String{Data: strData("foreign import"), Len: isize(len("foreign import"))})
		} else {
			filepaths = make([]*Ast, 0, 1)
			path := expect_token(f, TokenString)
			lit := ast_basic_lit(f, path)
			filepaths = append(filepaths, lit)
		}
		var s *Ast
		if len(filepaths) == 0 {
			syntax_error(lib_name.Kind, "foreign import without any paths")
			s = ast_bad_decl(f, lib_name, f.CurrToken)
		} else if f.CurrProc != nil {
			syntax_error(lib_name.Kind, "You cannot use foreign import within a procedure. This must be done at the file scope")
			s = ast_bad_decl(f, lib_name, ast_token(filepaths[0]))
		} else {
			s = ast_foreign_import_decl(f, token, filepaths, lib_name, multiple_filepaths, docs, f.LineComment)
		}
		_ = import_token
		expect_semicolon(f)
		return s
	}
	syntax_error(token.Kind, "Invalid foreign declaration")
	return ast_bad_decl(f, token, f.CurrToken)
}

func parse_attribute(f *AstFile, token Token, open_kind TokenKind, close_kind TokenKind, docs *CommentGroup) *Ast {
	var elems []*Ast
	open := Token{}
	close := Token{}
	if f.CurrToken.Kind == TokenIdent {
		elems = make([]*Ast, 0, 1)
		elem := parse_ident(f)
		elems = append(elems, elem)
	} else {
		open = expect_token(f, open_kind)
		f.ExprLevel++
		if f.CurrToken.Kind == close_kind {
			// empty attribute
		} else {
			elems = make([]*Ast, 0)
			for f.CurrToken.Kind != close_kind && f.CurrToken.Kind != TokenEOF {
				elem := parse_ident(f)
				if f.CurrToken.Kind == TokenEq {
					eq := expect_token(f, TokenEq)
					value := parse_value(f)
					elem = ast_field_value(f, elem, value, eq)
				}
				elems = append(elems, elem)
				if !allow_field_separator(f) {
					break
				}
			}
		}
		f.ExprLevel--
		close = expect_closing(f, close_kind, String{Data: strData("attribute"), Len: isize(len("attribute"))})
	}
	attribute := ast_attribute(f, token, open, close, elems)
	skip_possible_newline(f)
	decl := parse_stmt(f)
	if decl.Kind == AstValueDecl {
		if decl.ValueDecl.Docs == nil && docs != nil {
			decl.ValueDecl.Docs = docs
		}
		decl.ValueDecl.Attributes = append(decl.ValueDecl.Attributes, attribute)
	} else if decl.Kind == AstForeignBlockDecl {
		decl.ForeignBlockDecl.Attributes = append(decl.ForeignBlockDecl.Attributes, attribute)
	} else if decl.Kind == AstForeignImportDecl {
		decl.ForeignImportDecl.Attributes = append(decl.ForeignImportDecl.Attributes, attribute)
	} else if decl.Kind == AstImportDecl {
		decl.ImportDecl.Attributes = append(decl.ImportDecl.Attributes, attribute)
	} else {
		syntax_error(decl.Kind, "Expected a value or foreign declaration after an attribute")
		return ast_bad_stmt(f, token, f.CurrToken)
	}
	return decl
}

func parse_unrolled_for_loop(f *AstFile, unroll_token Token) *Ast {
	var args []*Ast
	if allow_token(f, TokenOpenParen) {
		f.ExprLevel++
		if f.CurrToken.Kind == TokenCloseParen {
			syntax_error(f.CurrToken.Kind, "#unroll expected at least 1 argument, got 0")
		} else {
			args = make([]*Ast, 0)
			for f.CurrToken.Kind != TokenCloseParen && f.CurrToken.Kind != TokenEOF {
				arg := parse_value(f)
				if f.CurrToken.Kind == TokenEq {
					eq := expect_token(f, TokenEq)
					if arg != nil && arg.Kind != AstIdent {
						syntax_error(arg.Kind, "Expected an identifier for 'key=value'")
					}
					value := parse_value(f)
					arg = ast_field_value(f, arg, value, eq)
				}
				args = append(args, arg)
				if !allow_field_separator(f) {
					break
				}
			}
		}
		f.ExprLevel--
		close := expect_closing(f, TokenCloseParen, String{Data: strData("#unroll"), Len: isize(len("#unroll"))})
		_ = close
	}
	for_token := expect_token(f, TokenFor)
	var init *Ast
	var val0 *Ast
	var val1 *Ast
	in_token := Token{}
	var expr *Ast
	var body *Ast
	bad_stmt := false

	if f.CurrToken.Kind != TokenIn {
		idents := parse_ident_list(f, false)
		switch len(idents) {
		case 1:
			val0 = idents[0]
		case 2:
			val0 = idents[0]
			val1 = idents[1]
		default:
			syntax_error(for_token.Kind, "Expected either 1 or 2 identifiers")
			bad_stmt = true
		}
	}
	in_token = expect_token(f, TokenIn)
	prev_allow_range := f.AllowRange
	prev_level := f.ExprLevel
	f.AllowRange = true
	f.ExprLevel = -1
	expr = parse_expr(f, false)
	f.ExprLevel = prev_level
	f.AllowRange = prev_allow_range
	if allow_token(f, TokenDo) {
		body = parse_do_body(f, for_token, "the for statement")
	} else {
		body = parse_block_stmt(f, false)
	}
	if bad_stmt {
		return ast_bad_stmt(f, unroll_token, f.CurrToken)
	}
	return ast_unroll_range_stmt(f, unroll_token, init, args, for_token, val0, val1, in_token, expr, body)
}

func parse_foreign_import_decl(f *AstFile, token Token, filepaths []*Ast, lib_name Token, multiple_filepaths bool) *Ast {
	return ast_foreign_import_decl(f, token, filepaths, lib_name, multiple_filepaths, f.LeadComment, f.LineComment)
}

func parse_stmt(f *AstFile) *Ast {
	var s *Ast
	token := f.CurrToken
	switch token.Kind {
	case TokenContext, TokenProc, TokenIdent, TokenInteger, TokenFloat,
		TokenImag, TokenRune, TokenString, TokenOpenParen, TokenPointer,
		TokenAsm, TokenAdd, TokenSub, TokenXor, TokenNot, TokenAnd, TokenMul:
		s = parse_simple_stmt(f, StmtAllowFlagLabel)
		expect_semicolon(f)
		return s
	case TokenForeign:
		return parse_foreign_decl(f)
	case TokenImport:
		return parse_import_decl(f, 0)
	case TokenIf:
		return parse_if_stmt(f)
	case TokenWhen:
		return parse_when_stmt(f)
	case TokenFor:
		return parse_for_stmt(f)
	case TokenSwitch:
		return parse_switch_stmt(f)
	case TokenDefer:
		return parse_defer_stmt(f)
	case TokenReturn:
		return parse_return_stmt(f)
	case TokenBreak, TokenContinue, TokenFallthrough:
		tok := advance_token(f)
		var label *Ast
		if tok.Kind != TokenFallthrough && f.CurrToken.Kind == TokenIdent {
			label = parse_ident(f)
		}
		s = ast_branch_stmt(f, tok, label)
		expect_semicolon(f)
		return s
	case TokenUsing:
		docs := f.LeadComment
		tok := expect_token(f, TokenUsing)
		if f.CurrToken.Kind == TokenImport {
			return parse_import_decl(f, 1)
		}
		list := parse_lhs_expr_list(f)
		if len(list) == 0 {
			syntax_error(tok.Kind, "Illegal use of 'using' statement")
			expect_semicolon(f)
			return ast_bad_stmt(f, tok, f.CurrToken)
		}
		if f.CurrToken.Kind != TokenColon {
			expect_semicolon(f)
			return ast_using_stmt(f, tok, list)
		}
		expect_token_after(f, TokenColon, "identifier list")
		decl := parse_value_decl(f, list, docs)
		if decl != nil && decl.Kind == AstValueDecl {
			return decl
		}
		syntax_error(tok.Kind, "Illegal use of 'using' statement")
		return ast_bad_stmt(f, tok, f.CurrToken)
	case TokenAt:
		docs := f.LeadComment
		tok := expect_token(f, TokenAt)
		return parse_attribute(f, tok, TokenOpenParen, TokenCloseParen, docs)
	case TokenHash:
		hash_token := expect_token(f, TokenHash)
		name := expect_token(f, TokenIdent)
		tag := goStr(name.String)
		switch tag {
		case "bounds_check":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, StateFlagBoundsCheck)
		case "no_bounds_check":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, StateFlagNoBoundsCheck)
		case "type_assert":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, StateFlagTypeAssert)
		case "no_type_assert":
			s = parse_stmt(f)
			return parse_check_directive_for_statement(s, name, StateFlagNoTypeAssert)
		case "partial":
			s = parse_stmt(f)
			switch s.Kind {
			case AstSwitchStmt:
				if s.SwitchStmt.Partial {
					syntax_error(token.Kind, "#partial already applied to a switch statement")
				}
				s.SwitchStmt.Partial = true
			case AstTypeSwitchStmt:
				if s.TypeSwitchStmt.Partial {
					syntax_error(token.Kind, "#partial already applied to a switch statement")
				}
				s.TypeSwitchStmt.Partial = true
			case AstEmptyStmt:
				return parse_check_directive_for_statement(s, name, 0)
			default:
				syntax_error(token.Kind, "#partial can only be applied to a switch statement")
			}
			return s
		case "assert", "panic":
			t := ast_basic_directive(f, hash_token, name)
			stmt := ast_expr_stmt(f, parse_call_expr(f, t))
			expect_semicolon(f)
			return stmt
		case "force_inline", "force_no_inline", "must_tail":
			expr := parse_inlining_or_tailing_operand(f, name)
			stmt := ast_expr_stmt(f, expr)
			expect_semicolon(f)
			return stmt
		case "unroll":
			return parse_unrolled_for_loop(f, name)
		case "reverse":
			for_stmt := parse_stmt(f)
			if for_stmt.Kind == AstRangeStmt {
				if for_stmt.RangeStmt.Reverse {
					syntax_error(token.Kind, "#reverse already applied to a 'for in' statement")
				}
				for_stmt.RangeStmt.Reverse = true
			} else {
				syntax_error(token.Kind, "#reverse can only be applied to a 'for in' statement")
			}
			return for_stmt
		case "include":
			syntax_error(token.Kind, "#include is not a valid import declaration kind. Did you mean 'import'?")
			s = ast_bad_stmt(f, token, f.CurrToken)
		case "define":
			if name.Pos.Line == f.CurrToken.Pos.Line {
				call_like := false
				var macro_expr *Ast
				ident := f.CurrToken
				if allow_token(f, TokenIdent) &&
					name.Pos.Line == f.CurrToken.Pos.Line {
					if f.CurrToken.Kind == TokenOpenParen && f.CurrToken.Pos.Column == ident.Pos.Column+ident.String.Len {
						call_like = true
						parse_call_expr(f, nil)
					}
					if name.Pos.Line == f.CurrToken.Pos.Line && f.CurrToken.Kind != TokenSemicolon {
						macro_expr = parse_expr(f, false)
					}
				}
				syntax_error(ident.Kind, "#define is not a valid declaration, Odin does not have a C-like preprocessor.")
				_ = call_like
				_ = macro_expr
			} else {
				syntax_error(token.Kind, "#define is not a valid declaration, Odin does not have a C-like preprocessor.")
			}
			s = ast_bad_stmt(f, token, f.CurrToken)
		default:
			syntax_error(token.Kind, "Unknown tag directive used")
			s = ast_bad_stmt(f, token, f.CurrToken)
		}
		fix_advance_to_next_stmt(f)
		return s
	case TokenOpenBrace:
		return parse_block_stmt(f, false)
	case TokenSemicolon:
		s = ast_empty_stmt(f, token)
		expect_semicolon(f)
		return s
	case TokenFileTag:
		syntax_error(token.Kind, "Lines starting with #+ (file tags) are only allowed before the package line.")
		return ast_bad_stmt(f, token, f.CurrToken)
	}
	switch token.Kind {
	case TokenElse:
		expect_token(f, TokenElse)
		syntax_error(token.Kind, "'else' unattached to an 'if' statement")
		switch f.CurrToken.Kind {
		case TokenIf:
			return parse_if_stmt(f)
		case TokenWhen:
			return parse_when_stmt(f)
		case TokenOpenBrace:
			return parse_block_stmt(f, true)
		case TokenDo:
			expect_token(f, TokenDo)
			stmt := parse_do_body(f, Token{}, "the for statement")
			if buildContext.DisallowDo {
				syntax_error(stmt.Kind, "'do' has been disallowed")
			}
			return stmt
		default:
			fix_advance_to_next_stmt(f)
			return ast_bad_stmt(f, token, f.CurrToken)
		}
	}
	syntax_error(token.Kind, "Expected a statement, got...")
	fix_advance_to_next_stmt(f)
	return ast_bad_stmt(f, token, f.CurrToken)
}

func parse_stmt_list(f *AstFile) []*Ast {
	var list []*Ast
	for f.CurrToken.Kind != TokenCase &&
		f.CurrToken.Kind != TokenCloseBrace &&
		f.CurrToken.Kind != TokenEOF {
		parse_enforce_tabs(f)
		stmt := parse_stmt(f)
		if stmt != nil && stmt.Kind != AstEmptyStmt {
			list = append(list, stmt)
			if stmt.Kind == AstExprStmt &&
				stmt.ExprStmt.Expr != nil &&
				stmt.ExprStmt.Expr.Kind == AstProcLit {
				syntax_error(stmt.Kind, "Procedure literal evaluated but not used")
			}
		}
	}
	return list
}

func parse_check_or_return(operand *Ast, msg string) {
	if operand == nil {
		return
	}
	switch operand.Kind {
	case AstOrReturnExpr:
		syntax_error_with_verbose(operand.Kind, "'or_return' use within "+msg+" is not wrapped in parentheses (...)")
	case AstOrBranchExpr:
		syntax_error_with_verbose(operand.Kind, "'or_break/or_continue' use within "+msg+" is not wrapped in parentheses (...)")
	}
}
