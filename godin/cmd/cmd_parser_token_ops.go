package cmd

import (
	"unsafe"
)

func next_token0(f *AstFile) bool {
	if f.CurrTokenIndex+1 < isize(len(f.Tokens)) {
		f.CurrTokenIndex++
		f.CurrToken = f.Tokens[f.CurrTokenIndex]
		return true
	}
	syntax_error_token(f.CurrToken, "Token is EOF")
	return false
}

func consume_comment(f *AstFile, endLine_ *isize) Token {
	tok := f.CurrToken
	_ = tok.Kind == TokenComment
	endLine := tok.Pos.Line
	if tok.String.Len >= 2 && *(*byte)(unsafe.Add(unsafe.Pointer(tok.String.Data), 1)) == '*' {
		for i := isize(2); i < tok.String.Len; i++ {
			if *(*byte)(unsafe.Add(unsafe.Pointer(tok.String.Data), i)) == '\n' {
				endLine++
			}
		}
	}
	if endLine_ != nil {
		*endLine_ = endLine
	}
	next_token0(f)
	return tok
}

func consume_comment_group(f *AstFile, n isize, endLine_ *isize) *CommentGroup {
	list := make([]Token, 0)
	endLine := f.CurrToken.Pos.Line
	if f.CurrTokenIndex == 1 &&
		f.PrevToken.Kind == TokenComment &&
		f.PrevToken.Pos.Line+1 == f.CurrToken.Pos.Line {
		list = append(list, f.PrevToken)
	}
	for f.CurrToken.Kind == TokenComment &&
		f.CurrToken.Pos.Line <= endLine+n {
		var el isize
		tok := consume_comment(f, &el)
		endLine = el
		list = append(list, tok)
	}
	if endLine_ != nil {
		*endLine_ = endLine
	}
	var comments *CommentGroup
	if len(list) > 0 {
		comments = permanent_alloc_item[*CommentGroup]()
		comments.List = list
		f.Comments = append(f.Comments, comments)
	}
	return comments
}

func consume_comment_groups(f *AstFile, prev Token) {
	if f.CurrToken.Kind != TokenComment {
		return
	}
	var comment *CommentGroup
	endLine := isize(0)
	if f.CurrToken.Pos.Line == prev.Pos.Line {
		comment = consume_comment_group(f, 0, &endLine)
		if f.CurrToken.Pos.Line != endLine ||
			f.CurrToken.Pos.Line == prev.Pos.Line+1 ||
			f.CurrToken.Kind == TokenEOF {
			f.LineComment = comment
		}
	}
	endLine = -1
	for f.CurrToken.Kind == TokenComment {
		comment = consume_comment_group(f, 1, &endLine)
	}
	if endLine+1 == f.CurrToken.Pos.Line || endLine < 0 {
		f.LeadComment = comment
	}
	_ = f.CurrToken.Kind != TokenComment
}

func ignore_newlines(f *AstFile) bool {
	return f.ExprLevel > 0
}

func advance_token(f *AstFile) Token {
	f.LeadComment = nil
	f.LineComment = nil
	f.PrevTokenIndex = f.CurrTokenIndex
	f.PrevToken = f.CurrToken
	prev := f.PrevToken
	ok := next_token0(f)
	if ok {
		switch f.CurrToken.Kind {
		case TokenComment:
			consume_comment_groups(f, prev)
		case TokenSemicolon:
			if ignore_newlines(f) && f.CurrToken.String.Len == 1 && *f.CurrToken.String.Data == '\n' {
				advance_token(f)
			}
		}
	}
	return prev
}

func peek_token(f *AstFile) Token {
	for i := f.CurrTokenIndex + 1; i < isize(len(f.Tokens)); i++ {
		tok := f.Tokens[i]
		if tok.Kind == TokenComment {
			continue
		}
		return tok
	}
	return EmptyToken
}

func peek_token_n(f *AstFile, n isize) Token {
	var found Token
	for i := f.CurrTokenIndex + 1; i < isize(len(f.Tokens)); i++ {
		tok := f.Tokens[i]
		if tok.Kind == TokenComment {
			continue
		}
		found = tok
		if n == 0 {
			return found
		}
		n--
	}
	return EmptyToken
}

func skip_possible_newline(f *AstFile) bool {
	if token_is_newline(f.CurrToken) {
		advance_token(f)
		return true
	}
	return false
}

func skip_possible_newline_for_literal(f *AstFile, ignoreStrictStyle bool) bool {
	curr := f.CurrToken
	if token_is_newline(curr) {
		next := peek_token(f)
		if curr.Pos.Line+1 >= next.Pos.Line {
			switch next.Kind {
			case TokenOpenBrace, TokenElse:
				if buildContext.StrictStyle && !ignoreStrictStyle {
					syntax_error_token(next, "With '-strict-style' the attached brace style (1TBS) is enforced")
				}
				fallthrough
			case TokenWhere:
				advance_token(f)
				return true
			}
		}
	}
	return false
}

func token_to_string(tok Token) String {
	p := tokenStrings[tok.Kind]
	if token_is_newline(tok) {
		p = S("newline")
	}
	return p
}

func expect_token(f *AstFile, kind TokenKind) Token {
	prev := f.CurrToken
	if prev.Kind != kind {
		c := tokenStrings[kind]
		p := token_to_string(prev)
		begin_error_block()
		syntax_error_token(f.CurrToken, "Expected '%s', got '%s'", goStr(c), goStr(p))
		if kind == TokenIdent {
			switch prev.Kind {
			case TokenContext:
				error_line("\tSuggestion: '%s' is a keyword, would 'ctx' suffice?\n", goStr(prev.String))
			case TokenPackage:
				error_line("\tSuggestion: '%s' is a keyword, would 'pkg' suffice?\n", goStr(prev.String))
			default:
				if token_is_keyword(prev.Kind) {
					error_line("\tNote: '%s' is a keyword\n", goStr(prev.String))
				}
			}
		}
		end_error_block()
		if prev.Kind == TokenEOF {
			exit_with_errors()
		}
	}
	advance_token(f)
	return prev
}

func expect_token_after(f *AstFile, kind TokenKind, msg string) Token {
	prev := f.PrevToken
	curr := f.CurrToken
	if curr.Kind != kind {
		p := token_to_string(curr)
		token := f.CurrToken
		if token_is_newline(curr) {
			token = curr
			token.Pos.Column -= 1
			skip_possible_newline(f)
		}
		syntax_error_token(token, "Expected '%s' after %s, got '%s'", goStr(tokenStrings[kind]), msg, goStr(p))
	}
	advance_token(f)
	if ast_file_vet_style(f) &&
		prev.Kind == TokenComma &&
		prev.Pos.Line == curr.Pos.Line {
		syntax_error_token(prev, "No need for a trailing comma followed by a %s on the same line", goStr(tokenStrings[kind]))
	}
	return curr
}

func is_token_range(kind TokenKind) bool {
	switch kind {
	case TokenEllipsis, TokenRangeFull, TokenRangeHalf:
		return true
	}
	return false
}

func is_token_range_tok(tok Token) bool {
	return is_token_range(tok.Kind)
}

func expect_operator(f *AstFile) Token {
	prev := f.CurrToken
	if (prev.Kind == TokenIn || prev.Kind == TokenNotIn) && (f.ExprLevel >= 0 || f.AllowInExpr) {
	} else if prev.Kind == TokenIf || prev.Kind == TokenWhen {
	} else if prev.Kind == TokenOrElse || prev.Kind == TokenOrReturn ||
		prev.Kind == TokenOrBreak || prev.Kind == TokenOrContinue {
	} else if !(prev.Kind > TokenOperatorBegin && prev.Kind < TokenOperatorEnd) {
		p := token_to_string(prev)
		syntax_error_token(prev, "Expected an operator, got '%s'", goStr(p))
	} else if !f.AllowRange && is_token_range(prev) {
		p := token_to_string(prev)
		syntax_error_token(prev, "Expected an non-range operator, got '%s'", goStr(p))
	}
	if prev.Kind == TokenEllipsis {
		syntax_error_token(prev, "'..' for ranges are not allowed, did you mean '..<' or '..='?")
		f.Tokens[f.CurrTokenIndex].Flags |= TokenFlagReplace
	}
	advance_token(f)
	return prev
}

func allow_token(f *AstFile, kind TokenKind) bool {
	prev := f.CurrToken
	if prev.Kind == kind {
		advance_token(f)
		return true
	}
	return false
}

func expect_closing_brace_of_field_list(f *AstFile) Token {
	token := f.CurrToken
	if allow_token(f, TokenCloseBrace) {
		return token
	}
	ok := true
	if f.AllowNewline {
		ok = !skip_possible_newline(f)
	}
	if ok && allow_token(f, TokenSemicolon) {
		p := token_to_string(token)
		syntax_error_token(token_end_of_line(f, f.PrevToken), "Expected a comma, got a %s", goStr(p))
	}
	return expect_token(f, TokenCloseBrace)
}

func fix_advance_to_next_stmt(f *AstFile) {
	for {
		t := f.CurrToken
		switch t.Kind {
		case TokenEOF, TokenSemicolon:
			return
		case TokenPackage, TokenForeign, TokenImport,
			TokenIf, TokenFor, TokenWhen, TokenReturn,
			TokenSwitch, TokenDefer, TokenUsing,
			TokenBreak, TokenContinue, TokenFallthrough, TokenHash:
			if t.Pos == f.FixPrevPos &&
				f.FixCount < 6 {
				f.FixCount++
				return
			}
			if tokenPosLt(f.FixPrevPos, t.Pos) {
				f.FixPrevPos = t.Pos
				f.FixCount = 0
				return
			}
		}
		advance_token(f)
	}
}

func expect_closing(f *AstFile, kind TokenKind, context String) Token {
	if f.CurrToken.Kind != kind &&
		f.CurrToken.Kind == TokenSemicolon &&
		((f.CurrToken.String.Len == 1 && *f.CurrToken.String.Data == '\n') || f.CurrToken.Kind == TokenEOF) {
		if f.AllowNewline {
			tok := f.PrevToken
			tok.Pos.Column += int32(tok.String.Len)
			syntax_error_token(tok, "Missing ',' before newline in %s", goStr(context))
		}
		advance_token(f)
	}
	return expect_token(f, kind)
}

func assign_removal_flag_to_semicolon(f *AstFile) {
	prevToken := &f.Tokens[f.PrevTokenIndex]
	currToken := &f.Tokens[f.CurrTokenIndex]
	_ = prevToken.Kind == TokenSemicolon
	if !(prevToken.String.Len == 1 && *prevToken.String.Data == ';') {
		return
	}
	ok := false
	if currToken.Pos.Line > prevToken.Pos.Line {
		ok = true
	} else if currToken.Pos.Line == prevToken.Pos.Line {
		switch currToken.Kind {
		case TokenCloseBrace, TokenCloseParen, TokenEOF:
			ok = true
		}
	}
	if !ok {
		return
	}
	if buildContext.StrictStyle || (ast_file_vet_flags(f)&uint64(VetFlagSemicolon)) != 0 {
		syntax_error_token(*prevToken, "Found unneeded semicolon")
	}
	prevToken.Flags |= TokenFlagRemove
}

func expect_semicolon(f *AstFile) {
	var prevToken Token
	if allow_token(f, TokenSemicolon) {
		assign_removal_flag_to_semicolon(f)
		return
	}
	switch f.CurrToken.Kind {
	case TokenCloseBrace, TokenCloseParen:
		if f.CurrToken.Pos.Line == f.PrevToken.Pos.Line {
			return
		}
	}
	prevToken = f.PrevToken
	if prevToken.Kind == TokenSemicolon {
		assign_removal_flag_to_semicolon(f)
		return
	}
	if f.CurrToken.Kind == TokenEOF {
		return
	}
	switch f.CurrToken.Kind {
	case TokenEOF:
		return
	}
	if f.CurrToken.Pos.Line == f.PrevToken.Pos.Line {
		p := token_to_string(f.CurrToken)
		prevToken.Pos = token_pos_end(prevToken)
		syntax_error_token(prevToken, "Expected ';', got %s", goStr(p))
		fix_advance_to_next_stmt(f)
	}
}

func clone_ast_array(array []*Ast, f *AstFile) []*Ast {
	if len(array) == 0 {
		return nil
	}
	result := make([]*Ast, len(array))
	for i := range array {
		result[i] = clone_ast(array[i], f)
	}
	return result
}

func clone_ast(node *Ast, f *AstFile) *Ast {
	if node == nil {
		return nil
	}
	if f == nil {
		f = thread_safe_get_ast_file_from_id(node.FileID)
	}
	n := alloc_ast_node(f, node.Kind)
	*n = *node
	switch n.Kind {
	case AstIdent:
		n.Ident.Entity.Store(nil)
	case AstPolyType:
		n.PolyType.Type = clone_ast(n.PolyType.Type, f)
		n.PolyType.Specialization = clone_ast(n.PolyType.Specialization, f)
	case AstEllipsis:
		n.Ellipsis.Expr = clone_ast(n.Ellipsis.Expr, f)
	case AstProcGroup:
		n.ProcGroup.Args = clone_ast_array(n.ProcGroup.Args, f)
	case AstProcLit:
		n.ProcLit.Type = clone_ast(n.ProcLit.Type, f)
		n.ProcLit.Body = clone_ast(n.ProcLit.Body, f)
		n.ProcLit.WhereClauses = clone_ast_array(n.ProcLit.WhereClauses, f)
	case AstCompoundLit:
		n.CompoundLit.Type = clone_ast(n.CompoundLit.Type, f)
		n.CompoundLit.Elems = clone_ast_array(n.CompoundLit.Elems, f)
	case AstTagExpr:
		n.TagExpr.Expr = clone_ast(n.TagExpr.Expr, f)
	case AstUnaryExpr:
		n.UnaryExpr.Expr = clone_ast(n.UnaryExpr.Expr, f)
	case AstBinaryExpr:
		n.BinaryExpr.Left = clone_ast(n.BinaryExpr.Left, f)
		n.BinaryExpr.Right = clone_ast(n.BinaryExpr.Right, f)
	case AstParenExpr:
		n.ParenExpr.Expr = clone_ast(n.ParenExpr.Expr, f)
	case AstSelectorExpr:
		n.SelectorExpr.Expr = clone_ast(n.SelectorExpr.Expr, f)
		n.SelectorExpr.Selector = clone_ast(n.SelectorExpr.Selector, f)
	case AstImplicitSelectorExpr:
		n.ImplicitSelectorExpr.Selector = clone_ast(n.ImplicitSelectorExpr.Selector, f)
	case AstSelectorCallExpr:
		n.SelectorCallExpr.Expr = clone_ast(n.SelectorCallExpr.Expr, f)
		n.SelectorCallExpr.Call = clone_ast(n.SelectorCallExpr.Call, f)
	case AstIndexExpr:
		n.IndexExpr.Expr = clone_ast(n.IndexExpr.Expr, f)
		n.IndexExpr.Index = clone_ast(n.IndexExpr.Index, f)
	case AstMatrixIndexExpr:
		n.MatrixIndexExpr.Expr = clone_ast(n.MatrixIndexExpr.Expr, f)
		n.MatrixIndexExpr.RowIndex = clone_ast(n.MatrixIndexExpr.RowIndex, f)
		n.MatrixIndexExpr.ColumnIndex = clone_ast(n.MatrixIndexExpr.ColumnIndex, f)
	case AstDerefExpr:
		n.DerefExpr.Expr = clone_ast(n.DerefExpr.Expr, f)
	case AstSliceExpr:
		n.SliceExpr.Expr = clone_ast(n.SliceExpr.Expr, f)
		n.SliceExpr.Low = clone_ast(n.SliceExpr.Low, f)
		n.SliceExpr.High = clone_ast(n.SliceExpr.High, f)
	case AstCallExpr:
		n.CallExpr.Proc = clone_ast(n.CallExpr.Proc, f)
		n.CallExpr.Args = clone_ast_array(n.CallExpr.Args, f)
	case AstFieldValue:
		n.FieldValue.Field = clone_ast(n.FieldValue.Field, f)
		n.FieldValue.Value = clone_ast(n.FieldValue.Value, f)
	case AstEnumFieldValue:
		n.EnumFieldValue.Name = clone_ast(n.EnumFieldValue.Name, f)
		n.EnumFieldValue.Value = clone_ast(n.EnumFieldValue.Value, f)
	case AstTernaryIfExpr:
		n.TernaryIfExpr.X = clone_ast(n.TernaryIfExpr.X, f)
		n.TernaryIfExpr.Cond = clone_ast(n.TernaryIfExpr.Cond, f)
		n.TernaryIfExpr.Y = clone_ast(n.TernaryIfExpr.Y, f)
	case AstTernaryWhenExpr:
		n.TernaryWhenExpr.X = clone_ast(n.TernaryWhenExpr.X, f)
		n.TernaryWhenExpr.Cond = clone_ast(n.TernaryWhenExpr.Cond, f)
		n.TernaryWhenExpr.Y = clone_ast(n.TernaryWhenExpr.Y, f)
	case AstOrElseExpr:
		n.OrElseExpr.X = clone_ast(n.OrElseExpr.X, f)
		n.OrElseExpr.Y = clone_ast(n.OrElseExpr.Y, f)
	case AstOrReturnExpr:
		n.OrReturnExpr.Expr = clone_ast(n.OrReturnExpr.Expr, f)
	case AstOrBranchExpr:
		n.OrBranchExpr.Label = clone_ast(n.OrBranchExpr.Label, f)
		n.OrBranchExpr.Expr = clone_ast(n.OrBranchExpr.Expr, f)
	case AstTypeAssertion:
		n.TypeAssertion.Expr = clone_ast(n.TypeAssertion.Expr, f)
		n.TypeAssertion.Type = clone_ast(n.TypeAssertion.Type, f)
	case AstTypeCast:
		n.TypeCast.Type = clone_ast(n.TypeCast.Type, f)
		n.TypeCast.Expr = clone_ast(n.TypeCast.Expr, f)
	case AstAutoCast:
		n.AutoCast.Expr = clone_ast(n.AutoCast.Expr, f)
	case AstInlineAsmExpr:
		n.InlineAsmExpr.ParamTypes = clone_ast_array(n.InlineAsmExpr.ParamTypes, f)
		n.InlineAsmExpr.ReturnType = clone_ast(n.InlineAsmExpr.ReturnType, f)
		n.InlineAsmExpr.AsmString = clone_ast(n.InlineAsmExpr.AsmString, f)
		n.InlineAsmExpr.ConstraintsString = clone_ast(n.InlineAsmExpr.ConstraintsString, f)
	case AstExprStmt:
		n.ExprStmt.Expr = clone_ast(n.ExprStmt.Expr, f)
	case AstAssignStmt:
		n.AssignStmt.LHS = clone_ast_array(n.AssignStmt.LHS, f)
		n.AssignStmt.RHS = clone_ast_array(n.AssignStmt.RHS, f)
	case AstBlockStmt:
		n.BlockStmt.Label = clone_ast(n.BlockStmt.Label, f)
		n.BlockStmt.Stmts = clone_ast_array(n.BlockStmt.Stmts, f)
	case AstIfStmt:
		n.IfStmt.Label = clone_ast(n.IfStmt.Label, f)
		n.IfStmt.Init = clone_ast(n.IfStmt.Init, f)
		n.IfStmt.Cond = clone_ast(n.IfStmt.Cond, f)
		n.IfStmt.Body = clone_ast(n.IfStmt.Body, f)
		n.IfStmt.ElseStmt = clone_ast(n.IfStmt.ElseStmt, f)
	case AstWhenStmt:
		n.WhenStmt.Cond = clone_ast(n.WhenStmt.Cond, f)
		n.WhenStmt.Body = clone_ast(n.WhenStmt.Body, f)
		n.WhenStmt.ElseStmt = clone_ast(n.WhenStmt.ElseStmt, f)
	case AstReturnStmt:
		n.ReturnStmt.Results = clone_ast_array(n.ReturnStmt.Results, f)
	case AstForStmt:
		n.ForStmt.Label = clone_ast(n.ForStmt.Label, f)
		n.ForStmt.Init = clone_ast(n.ForStmt.Init, f)
		n.ForStmt.Cond = clone_ast(n.ForStmt.Cond, f)
		n.ForStmt.Post = clone_ast(n.ForStmt.Post, f)
		n.ForStmt.Body = clone_ast(n.ForStmt.Body, f)
	case AstRangeStmt:
		n.RangeStmt.Label = clone_ast(n.RangeStmt.Label, f)
		n.RangeStmt.Init = clone_ast(n.RangeStmt.Init, f)
		n.RangeStmt.Vals = clone_ast_array(n.RangeStmt.Vals, f)
		n.RangeStmt.Expr = clone_ast(n.RangeStmt.Expr, f)
		n.RangeStmt.Body = clone_ast(n.RangeStmt.Body, f)
	case AstUnrollRangeStmt:
		n.UnrollRangeStmt.Args = clone_ast_array(n.UnrollRangeStmt.Args, f)
		n.UnrollRangeStmt.Init = clone_ast(n.UnrollRangeStmt.Init, f)
		n.UnrollRangeStmt.Val0 = clone_ast(n.UnrollRangeStmt.Val0, f)
		n.UnrollRangeStmt.Val1 = clone_ast(n.UnrollRangeStmt.Val1, f)
		n.UnrollRangeStmt.Expr = clone_ast(n.UnrollRangeStmt.Expr, f)
		n.UnrollRangeStmt.Body = clone_ast(n.UnrollRangeStmt.Body, f)
	case AstCaseClause:
		n.CaseClause.List = clone_ast_array(n.CaseClause.List, f)
		n.CaseClause.Stmts = clone_ast_array(n.CaseClause.Stmts, f)
		n.CaseClause.ImplicitEntity.Store(nil)
	case AstSwitchStmt:
		n.SwitchStmt.Label = clone_ast(n.SwitchStmt.Label, f)
		n.SwitchStmt.Init = clone_ast(n.SwitchStmt.Init, f)
		n.SwitchStmt.Tag = clone_ast(n.SwitchStmt.Tag, f)
		n.SwitchStmt.Body = clone_ast(n.SwitchStmt.Body, f)
	case AstTypeSwitchStmt:
		n.TypeSwitchStmt.Label = clone_ast(n.TypeSwitchStmt.Label, f)
		n.TypeSwitchStmt.Tag = clone_ast(n.TypeSwitchStmt.Tag, f)
		n.TypeSwitchStmt.Body = clone_ast(n.TypeSwitchStmt.Body, f)
	case AstDeferStmt:
		n.DeferStmt.Stmt = clone_ast(n.DeferStmt.Stmt, f)
	case AstBranchStmt:
		n.BranchStmt.Label = clone_ast(n.BranchStmt.Label, f)
	case AstUsingStmt:
		n.UsingStmt.List = clone_ast_array(n.UsingStmt.List, f)
	case AstForeignBlockDecl:
		n.ForeignBlockDecl.ForeignLibrary = clone_ast(n.ForeignBlockDecl.ForeignLibrary, f)
		n.ForeignBlockDecl.Body = clone_ast(n.ForeignBlockDecl.Body, f)
		n.ForeignBlockDecl.Attributes = clone_ast_array(n.ForeignBlockDecl.Attributes, f)
	case AstLabel:
		n.Label.Name = clone_ast(n.Label.Name, f)
	case AstValueDecl:
		n.ValueDecl.Names = clone_ast_array(n.ValueDecl.Names, f)
		n.ValueDecl.Type = clone_ast(n.ValueDecl.Type, f)
		n.ValueDecl.Values = clone_ast_array(n.ValueDecl.Values, f)
		n.ValueDecl.Attributes = clone_ast_array(n.ValueDecl.Attributes, f)
	case AstAttribute:
		n.Attribute.Elems = clone_ast_array(n.Attribute.Elems, f)
	case AstField:
		n.Field.Names = clone_ast_array(n.Field.Names, f)
		n.Field.Type = clone_ast(n.Field.Type, f)
	case AstBitFieldField:
		n.BitFieldField.Name = clone_ast(n.BitFieldField.Name, f)
		n.BitFieldField.Type = clone_ast(n.BitFieldField.Type, f)
		n.BitFieldField.BitSize = clone_ast(n.BitFieldField.BitSize, f)
	case AstFieldList:
		n.FieldList.List = clone_ast_array(n.FieldList.List, f)
	case AstTypeidType:
		n.TypeidType.Specialization = clone_ast(n.TypeidType.Specialization, f)
	case AstHelperType:
		n.HelperType.Type = clone_ast(n.HelperType.Type, f)
	case AstDistinctType:
		n.DistinctType.Type = clone_ast(n.DistinctType.Type, f)
	case AstProcType:
		n.ProcType.Params = clone_ast(n.ProcType.Params, f)
		n.ProcType.Results = clone_ast(n.ProcType.Results, f)
	case AstRelativeType:
		n.RelativeType.Tag = clone_ast(n.RelativeType.Tag, f)
		n.RelativeType.Type = clone_ast(n.RelativeType.Type, f)
	case AstPointerType:
		n.PointerType.Type = clone_ast(n.PointerType.Type, f)
		n.PointerType.Tag = clone_ast(n.PointerType.Tag, f)
	case AstMultiPointerType:
		n.MultiPointerType.Type = clone_ast(n.MultiPointerType.Type, f)
	case AstArrayType:
		n.ArrayType.Count = clone_ast(n.ArrayType.Count, f)
		n.ArrayType.Elem = clone_ast(n.ArrayType.Elem, f)
		n.ArrayType.Tag = clone_ast(n.ArrayType.Tag, f)
	case AstDynamicArrayType:
		n.DynamicArrayType.Elem = clone_ast(n.DynamicArrayType.Elem, f)
		n.DynamicArrayType.Tag = clone_ast(n.DynamicArrayType.Tag, f)
	case AstFixedCapacityDynamicArrayType:
		n.FixedCapacityDynamicArrayType.Elem = clone_ast(n.FixedCapacityDynamicArrayType.Elem, f)
		n.FixedCapacityDynamicArrayType.Capacity = clone_ast(n.FixedCapacityDynamicArrayType.Capacity, f)
		n.FixedCapacityDynamicArrayType.Tag = clone_ast(n.FixedCapacityDynamicArrayType.Tag, f)
	case AstStructType:
		n.StructType.Fields = clone_ast_array(n.StructType.Fields, f)
		n.StructType.PolymorphicParams = clone_ast(n.StructType.PolymorphicParams, f)
		n.StructType.Align = clone_ast(n.StructType.Align, f)
		n.StructType.MinFieldAlign = clone_ast(n.StructType.MinFieldAlign, f)
		n.StructType.MaxFieldAlign = clone_ast(n.StructType.MaxFieldAlign, f)
		n.StructType.WhereClauses = clone_ast_array(n.StructType.WhereClauses, f)
	case AstUnionType:
		n.UnionType.Variants = clone_ast_array(n.UnionType.Variants, f)
		n.UnionType.PolymorphicParams = clone_ast(n.UnionType.PolymorphicParams, f)
		n.UnionType.WhereClauses = clone_ast_array(n.UnionType.WhereClauses, f)
	case AstEnumType:
		n.EnumType.BaseType = clone_ast(n.EnumType.BaseType, f)
		n.EnumType.Fields = clone_ast_array(n.EnumType.Fields, f)
	case AstBitSetType:
		n.BitSetType.Elem = clone_ast(n.BitSetType.Elem, f)
		n.BitSetType.Underlying = clone_ast(n.BitSetType.Underlying, f)
	case AstBitFieldType:
		n.BitFieldType.BackingType = clone_ast(n.BitFieldType.BackingType, f)
		n.BitFieldType.Fields = clone_ast_array(n.BitFieldType.Fields, f)
	case AstMapType:
		n.MapType.Count = clone_ast(n.MapType.Count, f)
		n.MapType.Key = clone_ast(n.MapType.Key, f)
		n.MapType.Value = clone_ast(n.MapType.Value, f)
	case AstMatrixType:
		n.MatrixType.RowCount = clone_ast(n.MatrixType.RowCount, f)
		n.MatrixType.ColumnCount = clone_ast(n.MatrixType.ColumnCount, f)
		n.MatrixType.Elem = clone_ast(n.MatrixType.Elem, f)
	}
	return n
}

func error(node *Ast, format string, args ...any) {
	var tok Token
	var endPos TokenPos
	if node != nil {
		tok = ast_token(node)
		endPos = ast_end_pos(node)
	}
	error_va(tok.Pos, endPos, format, args...)
	if node != nil && node.FileID != 0 {
		f := thread_safe_get_ast_file_from_id(node.FileID)
		f.ErrorCount++
	}
}

func error_range(start, end TokenPos, format string, args ...any) {
	_ = start.FileID == end.FileID
	_ = start.Line == end.Line
	_ = start.Column <= end.Column
	_ = start.Offset <= end.Offset
	error_va(start, end, format, args...)
	if start.FileID != 0 {
		f := thread_safe_get_ast_file_from_id(start.FileID)
		f.ErrorCount++
	}
}

func syntax_error_with_verbose(node *Ast, format string, args ...any) {
	var tok Token
	var endPos TokenPos
	if node != nil {
		tok = ast_token(node)
		endPos = ast_end_pos(node)
	}
	syntax_error_with_verbose_va(tok.Pos, endPos, format, args...)
	if node != nil && node.FileID != 0 {
		f := thread_safe_get_ast_file_from_id(node.FileID)
		f.ErrorCount++
	}
}

func error_no_newline(node *Ast, format string, args ...any) {
	var tok Token
	if node != nil {
		tok = ast_token(node)
	}
	error_no_newline_va(tok.Pos, format, args...)
	if node != nil && node.FileID != 0 {
		f := thread_safe_get_ast_file_from_id(node.FileID)
		f.ErrorCount++
	}
}

func warning(node *Ast, format string, args ...any) {
	var tok Token
	var endPos TokenPos
	if node != nil {
		tok = ast_token(node)
		endPos = ast_end_pos(node)
	}
	warning_va(tok.Pos, endPos, format, args...)
}

func syntax_error(node *Ast, format string, args ...any) {
	var tok Token
	var endPos TokenPos
	if node != nil {
		tok = ast_token(node)
		endPos = ast_end_pos(node)
	}
	syntax_error_va(tok.Pos, endPos, format, args...)
	if node != nil && node.FileID != 0 {
		f := thread_safe_get_ast_file_from_id(node.FileID)
		f.ErrorCount++
	}
}

func ast_node_expect(node *Ast, kind AstKind) bool {
	if node.Kind != kind {
		syntax_error(node, "Expected %s, got %s", goStr(astStrings[kind]), goStr(astStrings[node.Kind]))
		return false
	}
	return true
}

func ast_node_expect2(node *Ast, kind0, kind1 AstKind) bool {
	if node.Kind != kind0 && node.Kind != kind1 {
		syntax_error(node, "Expected %s or %s, got %s", goStr(astStrings[kind0]), goStr(astStrings[kind1]), goStr(astStrings[node.Kind]))
		return false
	}
	return true
}

func is_blank_ident_str(str String) bool {
	if str.Len == 1 {
		return *str.Data == '_'
	}
	return false
}

func is_blank_ident_token(token Token) bool {
	if token.Kind == TokenIdent {
		return is_blank_ident_str(token.String)
	}
	return false
}

func is_blank_ident(node *Ast) bool {
	if node.Kind == AstIdent {
		return is_blank_ident_str(node.Ident.Token.String)
	}
	return false
}

func token_is_newline(tok Token) bool {
	return tok.Kind == TokenSemicolon && tok.String.Len == 1 && *tok.String.Data == '\n'
}

func token_is_keyword(kind TokenKind) bool {
	return kind > TokenKeywordBegin && kind < TokenKeywordEnd
}

func ast_file_vet_style(f *AstFile) bool {
	return (f.VetFlags & uint64(VetFlagStyle)) != 0
}

func ast_file_vet_flags(f *AstFile) uint64 {
	return f.VetFlags
}

func file_allow_newline(f *AstFile) bool {
	isStrict := buildContext.StrictStyle || ast_file_vet_style(f)
	return !isStrict
}

func token_pos_end(tok Token) TokenPos {
	pos := tok.Pos
	pos.Offset += int32(tok.String.Len)
	for i := isize(0); i < tok.String.Len; i++ {
		if *(*byte)(unsafe.Add(unsafe.Pointer(tok.String.Data), i)) == '\n' {
			pos.Line++
		}
	}
	return pos
}

func token_end_of_line(f *AstFile, tok Token) Token {
	// Walk from token's offset position to end of line (newline or EOF)
	raw := unsafe.Slice(f.Tokenizer.start, int(uintptr(unsafe.Pointer(f.Tokenizer.end))-uintptr(unsafe.Pointer(f.Tokenizer.start))))
	s := raw[tok.Pos.Offset:]
	n := isize(0)
	for n < isize(len(s)) && s[n] != 0 && s[n] != '\n' {
		n++
	}
	tok.Pos.Column += int32(n) - 1
	return tok
}

func syntax_error_token(tok Token, format string, args ...any) {
	syntax_error_va(tok.Pos, TokenPos{}, format, args...)
}
