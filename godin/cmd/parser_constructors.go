package cmd

import "unsafe"

func ast_token(node *Ast) Token {
	switch node.Kind {
	case AstIdent:
		return node.Ident.Token
	case AstImplicit:
		return node.Implicit
	case AstUninit:
		return node.Uninit
	case AstBasicLit:
		return node.BasicLit.Token
	case AstBasicDirective:
		return node.BasicDirective.Token
	case AstProcGroup:
		return node.ProcGroup.Token
	case AstProcLit:
		return ast_token(node.ProcLit.Type)
	case AstCompoundLit:
		if node.CompoundLit.Type != nil {
			return ast_token(node.CompoundLit.Type)
		}
		return node.CompoundLit.Open
	case AstTagExpr:
		return node.TagExpr.Token
	case AstBadExpr:
		return node.BadExpr.Begin
	case AstUnaryExpr:
		return node.UnaryExpr.Op
	case AstBinaryExpr:
		return ast_token(node.BinaryExpr.Left)
	case AstParenExpr:
		return node.ParenExpr.Open
	case AstCallExpr:
		return ast_token(node.CallExpr.Proc)
	case AstSelectorExpr:
		if node.SelectorExpr.Expr != nil {
			return ast_token(node.SelectorExpr.Expr)
		}
		if node.SelectorExpr.Selector != nil {
			return ast_token(node.SelectorExpr.Selector)
		}
		return node.SelectorExpr.Token
	case AstSelectorCallExpr:
		if node.SelectorCallExpr.Expr != nil {
			return ast_token(node.SelectorCallExpr.Expr)
		}
		return node.SelectorCallExpr.Token
	case AstImplicitSelectorExpr:
		if node.ImplicitSelectorExpr.Selector != nil {
			return ast_token(node.ImplicitSelectorExpr.Selector)
		}
		return node.ImplicitSelectorExpr.Token
	case AstIndexExpr:
		return ast_token(node.IndexExpr.Expr)
	case AstMatrixIndexExpr:
		return ast_token(node.MatrixIndexExpr.Expr)
	case AstSliceExpr:
		return ast_token(node.SliceExpr.Expr)
	case AstEllipsis:
		return node.Ellipsis.Token
	case AstFieldValue:
		if node.FieldValue.Field != nil {
			return ast_token(node.FieldValue.Field)
		}
		return node.FieldValue.Eq
	case AstEnumFieldValue:
		return ast_token(node.EnumFieldValue.Name)
	case AstDerefExpr:
		return node.DerefExpr.Op
	case AstTernaryIfExpr:
		return ast_token(node.TernaryIfExpr.X)
	case AstTernaryWhenExpr:
		return ast_token(node.TernaryWhenExpr.X)
	case AstOrElseExpr:
		return ast_token(node.OrElseExpr.X)
	case AstOrReturnExpr:
		return ast_token(node.OrReturnExpr.Expr)
	case AstOrBranchExpr:
		return ast_token(node.OrBranchExpr.Expr)
	case AstTypeAssertion:
		return ast_token(node.TypeAssertion.Expr)
	case AstTypeCast:
		return node.TypeCast.Token
	case AstAutoCast:
		return node.AutoCast.Token
	case AstInlineAsmExpr:
		return node.InlineAsmExpr.Token
	case AstBadStmt:
		return node.BadStmt.Begin
	case AstEmptyStmt:
		return node.EmptyStmt.Token
	case AstExprStmt:
		return ast_token(node.ExprStmt.Expr)
	case AstAssignStmt:
		return node.AssignStmt.Op
	case AstBlockStmt:
		return node.BlockStmt.Open
	case AstIfStmt:
		return node.IfStmt.Token
	case AstWhenStmt:
		return node.WhenStmt.Token
	case AstReturnStmt:
		return node.ReturnStmt.Token
	case AstForStmt:
		return node.ForStmt.Token
	case AstRangeStmt:
		return node.RangeStmt.Token
	case AstUnrollRangeStmt:
		return node.UnrollRangeStmt.UnrollToken
	case AstCaseClause:
		return node.CaseClause.Token
	case AstSwitchStmt:
		return node.SwitchStmt.Token
	case AstTypeSwitchStmt:
		return node.TypeSwitchStmt.Token
	case AstDeferStmt:
		return node.DeferStmt.Token
	case AstBranchStmt:
		return node.BranchStmt.Token
	case AstUsingStmt:
		return node.UsingStmt.Token
	case AstBadDecl:
		return node.BadDecl.Begin
	case AstLabel:
		return node.Label.Token
	case AstValueDecl:
		if len(node.ValueDecl.Names) > 0 {
			return ast_token(node.ValueDecl.Names[0])
		}
		return EmptyToken
	case AstPackageDecl:
		return node.PackageDecl.Token
	case AstImportDecl:
		return node.ImportDecl.Token
	case AstForeignImportDecl:
		return node.ForeignImportDecl.Token
	case AstForeignBlockDecl:
		return node.ForeignBlockDecl.Token
	case AstAttribute:
		return node.Attribute.Token
	case AstField:
		if len(node.Field.Names) > 0 {
			return ast_token(node.Field.Names[0])
		}
		return ast_token(node.Field.Type)
	case AstFieldList:
		return node.FieldList.Token
	case AstTypeidType:
		return node.TypeidType.Token
	case AstHelperType:
		return node.HelperType.Token
	case AstDistinctType:
		return node.DistinctType.Token
	case AstPolyType:
		return node.PolyType.Token
	case AstProcType:
		return node.ProcType.Token
	case AstRelativeType:
		return ast_token(node.RelativeType.Tag)
	case AstPointerType:
		return node.PointerType.Token
	case AstMultiPointerType:
		return node.MultiPointerType.Token
	case AstArrayType:
		return node.ArrayType.Token
	case AstDynamicArrayType:
		return node.DynamicArrayType.Token
	case AstFixedCapacityDynamicArrayType:
		return node.FixedCapacityDynamicArrayType.Token
	case AstStructType:
		return node.StructType.Token
	case AstUnionType:
		return node.UnionType.Token
	case AstEnumType:
		return node.EnumType.Token
	case AstBitSetType:
		return node.BitSetType.Token
	case AstBitFieldType:
		return node.BitFieldType.Token
	case AstMapType:
		return node.MapType.Token
	case AstMatrixType:
		return node.MatrixType.Token
	}
	return EmptyToken
}

func ast_end_token(node *Ast) Token {
	switch node.Kind {
	case AstInvalid:
		return EmptyToken
	case AstIdent:
		return node.Ident.Token
	case AstImplicit:
		return node.Implicit
	case AstUninit:
		return node.Uninit
	case AstBasicLit:
		return node.BasicLit.Token
	case AstBasicDirective:
		return node.BasicDirective.Token
	case AstProcGroup:
		return node.ProcGroup.Close
	case AstProcLit:
		if node.ProcLit.Body != nil {
			return ast_end_token(node.ProcLit.Body)
		}
		return ast_end_token(node.ProcLit.Type)
	case AstCompoundLit:
		return node.CompoundLit.Close
	case AstBadExpr:
		return node.BadExpr.End
	case AstTagExpr:
		if node.TagExpr.Expr != nil {
			return ast_end_token(node.TagExpr.Expr)
		}
		return node.TagExpr.Name
	case AstUnaryExpr:
		if node.UnaryExpr.Expr != nil {
			return ast_end_token(node.UnaryExpr.Expr)
		}
		return node.UnaryExpr.Op
	case AstBinaryExpr:
		return ast_end_token(node.BinaryExpr.Right)
	case AstParenExpr:
		return node.ParenExpr.Close
	case AstCallExpr:
		return node.CallExpr.Close
	case AstSelectorExpr:
		return ast_end_token(node.SelectorExpr.Selector)
	case AstSelectorCallExpr:
		return ast_end_token(node.SelectorCallExpr.Call)
	case AstImplicitSelectorExpr:
		if node.ImplicitSelectorExpr.Selector != nil {
			return ast_end_token(node.ImplicitSelectorExpr.Selector)
		}
		return node.ImplicitSelectorExpr.Token
	case AstIndexExpr:
		return node.IndexExpr.Close
	case AstMatrixIndexExpr:
		return node.MatrixIndexExpr.Close
	case AstSliceExpr:
		return node.SliceExpr.Close
	case AstEllipsis:
		if node.Ellipsis.Expr != nil {
			return ast_end_token(node.Ellipsis.Expr)
		}
		return node.Ellipsis.Token
	case AstFieldValue:
		return ast_end_token(node.FieldValue.Value)
	case AstEnumFieldValue:
		if node.EnumFieldValue.Value != nil {
			return ast_end_token(node.EnumFieldValue.Value)
		}
		return ast_end_token(node.EnumFieldValue.Name)
	case AstDerefExpr:
		return node.DerefExpr.Op
	case AstTernaryIfExpr:
		return ast_end_token(node.TernaryIfExpr.Y)
	case AstTernaryWhenExpr:
		return ast_end_token(node.TernaryWhenExpr.Y)
	case AstOrElseExpr:
		return ast_end_token(node.OrElseExpr.Y)
	case AstOrReturnExpr:
		return node.OrReturnExpr.Token
	case AstOrBranchExpr:
		if node.OrBranchExpr.Label != nil {
			return ast_end_token(node.OrBranchExpr.Label)
		}
		return node.OrBranchExpr.Token
	case AstTypeAssertion:
		return ast_end_token(node.TypeAssertion.Type)
	case AstTypeCast:
		return ast_end_token(node.TypeCast.Expr)
	case AstAutoCast:
		return ast_end_token(node.AutoCast.Expr)
	case AstInlineAsmExpr:
		return node.InlineAsmExpr.Close
	case AstBadStmt:
		return node.BadStmt.End
	case AstEmptyStmt:
		return node.EmptyStmt.Token
	case AstExprStmt:
		return ast_end_token(node.ExprStmt.Expr)
	case AstAssignStmt:
		if len(node.AssignStmt.RHS) > 0 {
			return ast_end_token(node.AssignStmt.RHS[len(node.AssignStmt.RHS)-1])
		}
		return node.AssignStmt.Op
	case AstBlockStmt:
		return node.BlockStmt.Close
	case AstIfStmt:
		if node.IfStmt.ElseStmt != nil {
			return ast_end_token(node.IfStmt.ElseStmt)
		}
		return ast_end_token(node.IfStmt.Body)
	case AstWhenStmt:
		if node.WhenStmt.ElseStmt != nil {
			return ast_end_token(node.WhenStmt.ElseStmt)
		}
		return ast_end_token(node.WhenStmt.Body)
	case AstReturnStmt:
		if len(node.ReturnStmt.Results) > 0 {
			return ast_end_token(node.ReturnStmt.Results[len(node.ReturnStmt.Results)-1])
		}
		return node.ReturnStmt.Token
	case AstForStmt:
		return ast_end_token(node.ForStmt.Body)
	case AstRangeStmt:
		return ast_end_token(node.RangeStmt.Body)
	case AstUnrollRangeStmt:
		return ast_end_token(node.UnrollRangeStmt.Body)
	case AstCaseClause:
		if len(node.CaseClause.Stmts) > 0 {
			return ast_end_token(node.CaseClause.Stmts[len(node.CaseClause.Stmts)-1])
		} else if len(node.CaseClause.List) > 0 {
			return ast_end_token(node.CaseClause.List[len(node.CaseClause.List)-1])
		}
		return node.CaseClause.Token
	case AstSwitchStmt:
		return ast_end_token(node.SwitchStmt.Body)
	case AstTypeSwitchStmt:
		return ast_end_token(node.TypeSwitchStmt.Body)
	case AstDeferStmt:
		return ast_end_token(node.DeferStmt.Stmt)
	case AstBranchStmt:
		if node.BranchStmt.Label != nil {
			return ast_end_token(node.BranchStmt.Label)
		}
		return node.BranchStmt.Token
	case AstUsingStmt:
		if len(node.UsingStmt.List) > 0 {
			return ast_end_token(node.UsingStmt.List[len(node.UsingStmt.List)-1])
		}
		return node.UsingStmt.Token
	case AstBadDecl:
		return node.BadDecl.End
	case AstLabel:
		if node.Label.Name != nil {
			return ast_end_token(node.Label.Name)
		}
		return node.Label.Token
	case AstValueDecl:
		if len(node.ValueDecl.Values) > 0 {
			return ast_end_token(node.ValueDecl.Values[len(node.ValueDecl.Values)-1])
		}
		if node.ValueDecl.Type != nil {
			return ast_end_token(node.ValueDecl.Type)
		}
		if len(node.ValueDecl.Names) > 0 {
			return ast_end_token(node.ValueDecl.Names[len(node.ValueDecl.Names)-1])
		}
		return EmptyToken
	case AstPackageDecl:
		return node.PackageDecl.Name
	case AstImportDecl:
		return node.ImportDecl.Relpath
	case AstForeignImportDecl:
		if len(node.ForeignImportDecl.Filepaths) > 0 {
			return ast_end_token(node.ForeignImportDecl.Filepaths[len(node.ForeignImportDecl.Filepaths)-1])
		}
		if node.ForeignImportDecl.LibraryName.Kind != TokenInvalid {
			return node.ForeignImportDecl.LibraryName
		}
		return node.ForeignImportDecl.Token
	case AstForeignBlockDecl:
		return ast_end_token(node.ForeignBlockDecl.Body)
	case AstAttribute:
		if node.Attribute.Close.Kind != TokenInvalid {
			return node.Attribute.Close
		}
		if len(node.Attribute.Elems) > 0 {
			return ast_end_token(node.Attribute.Elems[len(node.Attribute.Elems)-1])
		}
		if node.Attribute.Open.Kind != TokenInvalid {
			return node.Attribute.Open
		}
		return node.Attribute.Token
	case AstField:
		if node.Field.Tag.Kind != TokenInvalid {
			return node.Field.Tag
		}
		if node.Field.DefaultValue != nil {
			return ast_end_token(node.Field.DefaultValue)
		}
		if node.Field.Type != nil {
			return ast_end_token(node.Field.Type)
		}
		return ast_end_token(node.Field.Names[len(node.Field.Names)-1])
	case AstFieldList:
		if len(node.FieldList.List) > 0 {
			return ast_end_token(node.FieldList.List[len(node.FieldList.List)-1])
		}
		return node.FieldList.Token
	case AstTypeidType:
		if node.TypeidType.Specialization != nil {
			return ast_end_token(node.TypeidType.Specialization)
		}
		return node.TypeidType.Token
	case AstHelperType:
		return ast_end_token(node.HelperType.Type)
	case AstDistinctType:
		return ast_end_token(node.DistinctType.Type)
	case AstPolyType:
		if node.PolyType.Specialization != nil {
			return ast_end_token(node.PolyType.Specialization)
		}
		return ast_end_token(node.PolyType.Type)
	case AstProcType:
		if node.ProcType.Results != nil {
			return ast_end_token(node.ProcType.Results)
		}
		if node.ProcType.Params != nil {
			return ast_end_token(node.ProcType.Params)
		}
		return node.ProcType.Token
	case AstRelativeType:
		return ast_end_token(node.RelativeType.Type)
	case AstPointerType:
		return ast_end_token(node.PointerType.Type)
	case AstMultiPointerType:
		return ast_end_token(node.MultiPointerType.Type)
	case AstArrayType:
		return ast_end_token(node.ArrayType.Elem)
	case AstDynamicArrayType:
		return ast_end_token(node.DynamicArrayType.Elem)
	case AstFixedCapacityDynamicArrayType:
		return ast_end_token(node.FixedCapacityDynamicArrayType.Elem)
	case AstStructType:
		if len(node.StructType.Fields) > 0 {
			return ast_end_token(node.StructType.Fields[len(node.StructType.Fields)-1])
		}
		return node.StructType.Token
	case AstUnionType:
		if len(node.UnionType.Variants) > 0 {
			return ast_end_token(node.UnionType.Variants[len(node.UnionType.Variants)-1])
		}
		return node.UnionType.Token
	case AstEnumType:
		if len(node.EnumType.Fields) > 0 {
			return ast_end_token(node.EnumType.Fields[len(node.EnumType.Fields)-1])
		}
		if node.EnumType.BaseType != nil {
			return ast_end_token(node.EnumType.BaseType)
		}
		return node.EnumType.Token
	case AstBitSetType:
		if node.BitSetType.Underlying != nil {
			return ast_end_token(node.BitSetType.Underlying)
		}
		return ast_end_token(node.BitSetType.Elem)
	case AstBitFieldType:
		return node.BitFieldType.Close
	case AstMapType:
		return ast_end_token(node.MapType.Value)
	case AstMatrixType:
		return ast_end_token(node.MatrixType.Elem)
	}
	return EmptyToken
}

func ast_end_pos(node *Ast) TokenPos {
	return token_pos_end(ast_end_token(node))
}

func in_vet_packages(file *AstFile) bool {
	if file == nil {
		return true
	}
	if file.Pkg == nil {
		return true
	}
	if len(build_context.VetPackages) == 0 {
		return true
	}
	return string_set_exists(&build_context.VetPackages, file.Pkg.Name)
}

func ast_file_vet_flags(f *AstFile) uint64 {
	if f != nil && f.VetFlagsSet {
		return f.VetFlags
	}
	found := in_vet_packages(f)
	if found {
		return build_context.VetFlags
	}
	return 0
}

func ast_file_vet_style(f *AstFile) bool {
	return (ast_file_vet_flags(f) & uint64(VetFlagStyle)) != 0
}

func ast_file_vet_deprecated(f *AstFile) bool {
	return (ast_file_vet_flags(f) & uint64(VetFlagDeprecated)) != 0
}

func ast_file_vet_explicit_allocators(f *AstFile) bool {
	return (ast_file_vet_flags(f) & uint64(VetFlagExplicitAllocators)) != 0
}

func file_allow_newline(f *AstFile) bool {
	is_strict := build_context.StrictStyle || ast_file_vet_style(f)
	return !is_strict
}

func token_end_of_line(f *AstFile, tok Token) Token {
	start := f.Tokenizer.Start + tok.Pos.Offset
	s := start
	end := f.Tokenizer.End
	for s != end && *s != '\n' {
		s = unsafe.Add(s, 1)
	}
	tok.Pos.Column += int32(uintptr(unsafe.Pointer(s))-uintptr(unsafe.Pointer(start))) - 1
	return tok
}

func get_file_line_as_string(pos TokenPos, offset_ *int32) gbString {
	file := thread_safe_get_ast_file_from_id(pos.FileID)
	if file == nil {
		return nil
	}
	start := file.Tokenizer.Start
	end := file.Tokenizer.End
	if start == end {
		return nil
	}
	offset := isize(pos.Offset)
	if pos.Line != 0 && offset == 0 {
		for i := int32(1); i < pos.Line; i++ {
			for start+offset < end {
				c := *(*byte)(unsafe.Add(unsafe.Pointer(start), offset))
				offset++
				if c == '\n' {
					break
				}
			}
		}
		for i := int32(1); i < pos.Column; i++ {
			ptr := unsafe.Add(unsafe.Pointer(start), offset)
			c := *(*byte)(ptr)
			if c&0x80 != 0 {
				offset += isize(utf8_decode(ptr, end-start-offset, nil))
			} else {
				offset++
			}
		}
	}
	len_ := isize(uintptr(unsafe.Pointer(end)) - uintptr(unsafe.Pointer(start)))
	if len_ < offset {
		return nil
	}
	pos_offset := unsafe.Add(unsafe.Pointer(start), offset)
	line_start := pos_offset
	line_end := pos_offset
	if offset > 0 && *(*byte)(line_start) == '\n' {
		line_start = unsafe.Add(line_start, -1)
	}
	for uintptr(unsafe.Pointer(line_start)) >= uintptr(unsafe.Pointer(start)) {
		if *(*byte)(line_start) == '\n' {
			line_start = unsafe.Add(line_start, 1)
			break
		}
		line_start = unsafe.Add(line_start, -1)
	}
	if uintptr(unsafe.Pointer(line_start)) == uintptr(unsafe.Pointer(start))-1 {
		line_start = unsafe.Add(line_start, 1)
	}
	for uintptr(unsafe.Pointer(line_end)) < uintptr(unsafe.Pointer(end)) {
		if *(*byte)(line_end) == '\n' {
			break
		}
		line_end = unsafe.Add(line_end, 1)
	}
	the_line := make_string(line_start, isize(uintptr(unsafe.Pointer(line_end))-uintptr(unsafe.Pointer(line_start))))
	the_line = string_trim_whitespace(the_line)
	if offset_ != nil {
		*offset_ = int32(uintptr(unsafe.Pointer(pos_offset)) - uintptr(unsafe.Pointer(the_line.Data)))
	}
	return gb_string_make_length(heap_allocator(), the_line.Data, the_line.Len)
}

func ast_node_size(kind AstKind) isize {
	return isize(unsafe.Sizeof(AstCommonStuff{})) + ast_variant_sizes[kind]
}

func is_blank_ident(str String) bool {
	if str.Len == 1 {
		return *(*byte)(unsafe.Pointer(str.Data)) == '_'
	}
	return false
}

func is_blank_ident(token Token) bool {
	if token.Kind == TokenIdent {
		return is_blank_ident(token.String)
	}
	return false
}

func is_blank_ident(node *Ast) bool {
	if node.Kind == AstIdent {
		return is_blank_ident(node.Ident.Token.String)
	}
	return false
}

func ast_ident(f *AstFile, token Token) *Ast {
	result := alloc_ast_node(f, AstIdent)
	result.Ident.Token = token
	result.Ident.Hash = string_hash(token.String)
	result.Ident.Interned = string_interner_insert(token.String)
	return result
}

func ast_implicit(f *AstFile, token Token) *Ast {
	result := alloc_ast_node(f, AstImplicit)
	result.Implicit = token
	return result
}

func ast_uninit(f *AstFile, token Token) *Ast {
	result := alloc_ast_node(f, AstUninit)
	result.Uninit = token
	return result
}

func exact_value_from_token(f *AstFile, token Token) ExactValue {
	s := token.String
	string_interner_insert(s)
	switch token.Kind {
	case TokenRune:
		if !unquote_string(ast_allocator(f), &s, 0) {
			syntax_error(token, "Invalid rune literal")
		}
	case TokenString:
		if !unquote_string(ast_allocator(f), &s, 0, s.Data[0] == '`') {
			syntax_error(token, "Invalid string literal")
		}
	}
	value := exact_value_from_basic_literal(token.Kind, s)
	if value.Kind == ExactValue_Invalid {
		switch token.Kind {
		case TokenInteger:
			syntax_error(token, "Invalid integer literal")
		case TokenFloat:
			if !string_contains_char(s, '.') && !string_contains_char(s, '-') {
				syntax_error(token, "Invalid integer literal")
			} else {
				syntax_error(token, "Invalid float literal")
			}
		default:
			syntax_error(token, "Invalid token literal")
		}
	}
	return value
}

func string_value_from_token(f *AstFile, token Token) String {
	value := exact_value_from_token(f, token)
	var str String
	if value.Kind == ExactValue_String {
		str = value.ValueString
	}
	return str
}

func ast_basic_lit(f *AstFile, basic_lit Token) *Ast {
	result := alloc_ast_node(f, AstBasicLit)
	result.BasicLit.Token = basic_lit
	result.TAV.Mode = AddressingConstant
	result.TAV.Value = exact_value_from_token(f, basic_lit)
	return result
}

func ast_basic_directive(f *AstFile, token Token, name Token) *Ast {
	result := alloc_ast_node(f, AstBasicDirective)
	result.BasicDirective.Token = token
	result.BasicDirective.Name = name
	string_interner_insert(name.String)
	if string_starts_with(name.String, S("load")) {
		f.SeenLoadDirectiveCount.Add(1)
	}
	return result
}

func ast_ellipsis(f *AstFile, token Token, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstEllipsis)
	result.Ellipsis.Token = token
	result.Ellipsis.Expr = expr
	return result
}

func ast_proc_group(f *AstFile, token Token, open Token, close Token, args []*Ast) *Ast {
	result := alloc_ast_node(f, AstProcGroup)
	result.ProcGroup.Token = token
	result.ProcGroup.Open = open
	result.ProcGroup.Close = close
	result.ProcGroup.Args = slice_from_array(args)
	return result
}

func ast_proc_lit(f *AstFile, typ *Ast, body *Ast, tags uint64, where_token Token, where_clauses []*Ast) *Ast {
	result := alloc_ast_node(f, AstProcLit)
	result.ProcLit.Type = typ
	result.ProcLit.Body = body
	result.ProcLit.Tags = tags
	result.ProcLit.WhereToken = where_token
	result.ProcLit.WhereClauses = slice_from_array(where_clauses)
	return result
}

func ast_field_value(f *AstFile, field *Ast, value *Ast, eq Token) *Ast {
	result := alloc_ast_node(f, AstFieldValue)
	result.FieldValue.Field = field
	result.FieldValue.Value = value
	result.FieldValue.Eq = eq
	return result
}

func ast_enum_field_value(f *AstFile, name *Ast, value *Ast, docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstEnumFieldValue)
	result.EnumFieldValue.Name = name
	result.EnumFieldValue.Value = value
	result.EnumFieldValue.Docs = docs
	result.EnumFieldValue.Comment = comment
	return result
}

func ast_compound_lit(f *AstFile, typ *Ast, elems []*Ast, open Token, close Token) *Ast {
	result := alloc_ast_node(f, AstCompoundLit)
	result.CompoundLit.Type = typ
	result.CompoundLit.Elems = slice_from_array(elems)
	result.CompoundLit.Open = open
	result.CompoundLit.Close = close
	return result
}

func ast_ternary_if_expr(f *AstFile, x *Ast, cond *Ast, y *Ast) *Ast {
	result := alloc_ast_node(f, AstTernaryIfExpr)
	result.TernaryIfExpr.X = x
	result.TernaryIfExpr.Cond = cond
	result.TernaryIfExpr.Y = y
	return result
}

func ast_ternary_when_expr(f *AstFile, x *Ast, cond *Ast, y *Ast) *Ast {
	result := alloc_ast_node(f, AstTernaryWhenExpr)
	result.TernaryWhenExpr.X = x
	result.TernaryWhenExpr.Cond = cond
	result.TernaryWhenExpr.Y = y
	return result
}

func ast_or_else_expr(f *AstFile, x *Ast, token Token, y *Ast) *Ast {
	result := alloc_ast_node(f, AstOrElseExpr)
	result.OrElseExpr.X = x
	result.OrElseExpr.Token = token
	result.OrElseExpr.Y = y
	return result
}

func ast_or_return_expr(f *AstFile, expr *Ast, token Token) *Ast {
	result := alloc_ast_node(f, AstOrReturnExpr)
	result.OrReturnExpr.Expr = expr
	result.OrReturnExpr.Token = token
	return result
}

func ast_or_branch_expr(f *AstFile, expr *Ast, token Token, label *Ast) *Ast {
	result := alloc_ast_node(f, AstOrBranchExpr)
	result.OrBranchExpr.Expr = expr
	result.OrBranchExpr.Token = token
	result.OrBranchExpr.Label = label
	return result
}

func ast_type_assertion(f *AstFile, expr *Ast, dot Token, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstTypeAssertion)
	result.TypeAssertion.Expr = expr
	result.TypeAssertion.Dot = dot
	result.TypeAssertion.Type = typ
	return result
}

func ast_type_cast(f *AstFile, token Token, typ *Ast, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstTypeCast)
	result.TypeCast.Token = token
	result.TypeCast.Type = typ
	result.TypeCast.Expr = expr
	return result
}

func ast_auto_cast(f *AstFile, token Token, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstAutoCast)
	result.AutoCast.Token = token
	result.AutoCast.Expr = expr
	return result
}

func ast_inline_asm_expr(f *AstFile, token Token, open Token, close Token,
	param_types []*Ast,
	return_type *Ast,
	asm_string *Ast,
	constraints_string *Ast,
	has_side_effects bool,
	is_align_stack bool,
	dialect InlineAsmDialectKind) *Ast {
	result := alloc_ast_node(f, AstInlineAsmExpr)
	result.InlineAsmExpr.Token = token
	result.InlineAsmExpr.Open = open
	result.InlineAsmExpr.Close = close
	result.InlineAsmExpr.ParamTypes = slice_from_array(param_types)
	result.InlineAsmExpr.ReturnType = return_type
	result.InlineAsmExpr.AsmString = asm_string
	result.InlineAsmExpr.ConstraintsString = constraints_string
	result.InlineAsmExpr.HasSideEffects = has_side_effects
	result.InlineAsmExpr.IsAlignStack = is_align_stack
	result.InlineAsmExpr.Dialect = dialect
	return result
}

func ast_bad_stmt(f *AstFile, begin Token, end Token) *Ast {
	result := alloc_ast_node(f, AstBadStmt)
	result.BadStmt.Begin = begin
	result.BadStmt.End = end
	return result
}

func ast_empty_stmt(f *AstFile, token Token) *Ast {
	result := alloc_ast_node(f, AstEmptyStmt)
	result.EmptyStmt.Token = token
	return result
}

func ast_expr_stmt(f *AstFile, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstExprStmt)
	result.ExprStmt.Expr = expr
	return result
}

func ast_assign_stmt(f *AstFile, op Token, lhs []*Ast, rhs []*Ast) *Ast {
	result := alloc_ast_node(f, AstAssignStmt)
	result.AssignStmt.Op = op
	result.AssignStmt.LHS = slice_from_array(lhs)
	result.AssignStmt.RHS = slice_from_array(rhs)
	return result
}

func ast_block_stmt(f *AstFile, stmts []*Ast, open Token, close Token) *Ast {
	result := alloc_ast_node(f, AstBlockStmt)
	result.BlockStmt.Stmts = slice_from_array(stmts)
	result.BlockStmt.Open = open
	result.BlockStmt.Close = close
	return result
}

func ast_if_stmt(f *AstFile, token Token, init *Ast, cond *Ast, body *Ast, else_stmt *Ast) *Ast {
	result := alloc_ast_node(f, AstIfStmt)
	result.IfStmt.Token = token
	result.IfStmt.Init = init
	result.IfStmt.Cond = cond
	result.IfStmt.Body = body
	result.IfStmt.ElseStmt = else_stmt
	return result
}

func ast_when_stmt(f *AstFile, token Token, cond *Ast, body *Ast, else_stmt *Ast) *Ast {
	result := alloc_ast_node(f, AstWhenStmt)
	result.WhenStmt.Token = token
	result.WhenStmt.Cond = cond
	result.WhenStmt.Body = body
	result.WhenStmt.ElseStmt = else_stmt
	return result
}

func ast_return_stmt(f *AstFile, token Token, results []*Ast) *Ast {
	result := alloc_ast_node(f, AstReturnStmt)
	result.ReturnStmt.Token = token
	result.ReturnStmt.Results = slice_from_array(results)
	return result
}

func ast_for_stmt(f *AstFile, token Token, init *Ast, cond *Ast, post *Ast, body *Ast) *Ast {
	result := alloc_ast_node(f, AstForStmt)
	result.ForStmt.Token = token
	result.ForStmt.Init = init
	result.ForStmt.Cond = cond
	result.ForStmt.Post = post
	result.ForStmt.Body = body
	return result
}

func ast_range_stmt(f *AstFile, token Token, init *Ast, vals []*Ast, in_token Token, expr *Ast, body *Ast) *Ast {
	result := alloc_ast_node(f, AstRangeStmt)
	result.RangeStmt.Token = token
	result.RangeStmt.Init = init
	result.RangeStmt.Vals = vals
	result.RangeStmt.InToken = in_token
	result.RangeStmt.Expr = expr
	result.RangeStmt.Body = body
	return result
}

func ast_unroll_range_stmt(f *AstFile, unroll_token Token, init *Ast, args []*Ast, for_token Token, val0 *Ast, val1 *Ast, in_token Token, expr *Ast, body *Ast) *Ast {
	result := alloc_ast_node(f, AstUnrollRangeStmt)
	result.UnrollRangeStmt.UnrollToken = unroll_token
	result.UnrollRangeStmt.Init = init
	result.UnrollRangeStmt.Args = args
	result.UnrollRangeStmt.ForToken = for_token
	result.UnrollRangeStmt.Val0 = val0
	result.UnrollRangeStmt.Val1 = val1
	result.UnrollRangeStmt.InToken = in_token
	result.UnrollRangeStmt.Expr = expr
	result.UnrollRangeStmt.Body = body
	return result
}

func ast_switch_stmt(f *AstFile, token Token, init *Ast, tag *Ast, body *Ast) *Ast {
	result := alloc_ast_node(f, AstSwitchStmt)
	result.SwitchStmt.Token = token
	result.SwitchStmt.Init = init
	result.SwitchStmt.Tag = tag
	result.SwitchStmt.Body = body
	result.SwitchStmt.Partial = false
	return result
}

func ast_type_switch_stmt(f *AstFile, token Token, tag *Ast, body *Ast) *Ast {
	result := alloc_ast_node(f, AstTypeSwitchStmt)
	result.TypeSwitchStmt.Token = token
	result.TypeSwitchStmt.Tag = tag
	result.TypeSwitchStmt.Body = body
	result.TypeSwitchStmt.Partial = false
	return result
}

func ast_case_clause(f *AstFile, token Token, list []*Ast, stmts []*Ast) *Ast {
	result := alloc_ast_node(f, AstCaseClause)
	result.CaseClause.Token = token
	result.CaseClause.List = slice_from_array(list)
	result.CaseClause.Stmts = slice_from_array(stmts)
	return result
}

func ast_defer_stmt(f *AstFile, token Token, stmt *Ast) *Ast {
	result := alloc_ast_node(f, AstDeferStmt)
	result.DeferStmt.Token = token
	result.DeferStmt.Stmt = stmt
	return result
}

func ast_branch_stmt(f *AstFile, token Token, label *Ast) *Ast {
	result := alloc_ast_node(f, AstBranchStmt)
	result.BranchStmt.Token = token
	result.BranchStmt.Label = label
	return result
}

func ast_using_stmt(f *AstFile, token Token, list []*Ast) *Ast {
	result := alloc_ast_node(f, AstUsingStmt)
	result.UsingStmt.Token = token
	result.UsingStmt.List = slice_from_array(list)
	return result
}

func ast_bad_decl(f *AstFile, begin Token, end Token) *Ast {
	result := alloc_ast_node(f, AstBadDecl)
	result.BadDecl.Begin = begin
	result.BadDecl.End = end
	return result
}

func ast_field(f *AstFile, names []*Ast, typ *Ast, default_value *Ast, flags uint32, tag Token,
	docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstField)
	result.Field.Names = slice_from_array(names)
	result.Field.Type = typ
	result.Field.DefaultValue = default_value
	result.Field.Flags = flags
	result.Field.Tag = tag
	result.Field.Docs = docs
	result.Field.Comment = comment
	return result
}

func ast_bit_field_field(f *AstFile, name *Ast, typ *Ast, bit_size *Ast, tag Token,
	docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstBitFieldField)
	result.BitFieldField.Name = name
	result.BitFieldField.Type = typ
	result.BitFieldField.BitSize = bit_size
	result.BitFieldField.Tag = tag
	result.BitFieldField.Docs = docs
	result.BitFieldField.Comment = comment
	return result
}

func ast_field_list(f *AstFile, token Token, list []*Ast) *Ast {
	result := alloc_ast_node(f, AstFieldList)
	result.FieldList.Token = token
	result.FieldList.List = slice_from_array(list)
	return result
}

func ast_typeid_type(f *AstFile, token Token, specialization *Ast) *Ast {
	result := alloc_ast_node(f, AstTypeidType)
	result.TypeidType.Token = token
	result.TypeidType.Specialization = specialization
	return result
}

func ast_helper_type(f *AstFile, token Token, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstHelperType)
	result.HelperType.Token = token
	result.HelperType.Type = typ
	return result
}

func ast_distinct_type(f *AstFile, token Token, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstDistinctType)
	result.DistinctType.Token = token
	result.DistinctType.Type = typ
	return result
}

func ast_poly_type(f *AstFile, token Token, typ *Ast, specialization *Ast) *Ast {
	result := alloc_ast_node(f, AstPolyType)
	result.PolyType.Token = token
	result.PolyType.Type = typ
	result.PolyType.Specialization = specialization
	return result
}

func ast_proc_type(f *AstFile, token Token, params *Ast, results *Ast, tags uint64, calling_convention ProcCallingConvention, generic bool, diverging bool) *Ast {
	result := alloc_ast_node(f, AstProcType)
	result.ProcType.Token = token
	result.ProcType.Params = params
	result.ProcType.Results = results
	result.ProcType.Tags = tags
	result.ProcType.CallingConvention = calling_convention
	result.ProcType.Generic = generic
	result.ProcType.Diverging = diverging
	return result
}

func ast_relative_type(f *AstFile, tag *Ast, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstRelativeType)
	result.RelativeType.Tag = tag
	result.RelativeType.Type = typ
	return result
}

func ast_pointer_type(f *AstFile, token Token, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstPointerType)
	result.PointerType.Token = token
	result.PointerType.Type = typ
	return result
}

func ast_multi_pointer_type(f *AstFile, token Token, typ *Ast) *Ast {
	result := alloc_ast_node(f, AstMultiPointerType)
	result.MultiPointerType.Token = token
	result.MultiPointerType.Type = typ
	return result
}

func ast_array_type(f *AstFile, token Token, count *Ast, elem *Ast) *Ast {
	result := alloc_ast_node(f, AstArrayType)
	result.ArrayType.Token = token
	result.ArrayType.Count = count
	result.ArrayType.Elem = elem
	return result
}

func ast_dynamic_array_type(f *AstFile, token Token, elem *Ast) *Ast {
	result := alloc_ast_node(f, AstDynamicArrayType)
	result.DynamicArrayType.Token = token
	result.DynamicArrayType.Elem = elem
	return result
}

func ast_fixed_capacity_dynamic_array_type(f *AstFile, token Token, capacity *Ast, elem *Ast) *Ast {
	result := alloc_ast_node(f, AstFixedCapacityDynamicArrayType)
	result.FixedCapacityDynamicArrayType.Token = token
	result.FixedCapacityDynamicArrayType.Capacity = capacity
	result.FixedCapacityDynamicArrayType.Elem = elem
	return result
}

func ast_struct_type(f *AstFile, token Token, fields []*Ast, field_count isize,
	polymorphic_params *Ast, is_packed bool, is_raw_union bool, is_all_or_none bool, is_simple bool,
	align *Ast, min_field_align *Ast, max_field_align *Ast,
	where_token Token, where_clauses []*Ast) *Ast {
	result := alloc_ast_node(f, AstStructType)
	result.StructType.Token = token
	result.StructType.Fields = fields
	result.StructType.FieldCount = field_count
	result.StructType.PolymorphicParams = polymorphic_params
	result.StructType.IsPacked = is_packed
	result.StructType.IsRawUnion = is_raw_union
	result.StructType.IsAllOrNone = is_all_or_none
	result.StructType.IsSimple = is_simple
	result.StructType.Align = align
	result.StructType.MinFieldAlign = min_field_align
	result.StructType.MaxFieldAlign = max_field_align
	result.StructType.WhereToken = where_token
	result.StructType.WhereClauses = slice_from_array(where_clauses)
	return result
}

func ast_union_type(f *AstFile, token Token, variants []*Ast, polymorphic_params *Ast, align *Ast, kind UnionTypeKind,
	where_token Token, where_clauses []*Ast) *Ast {
	result := alloc_ast_node(f, AstUnionType)
	result.UnionType.Token = token
	result.UnionType.Variants = slice_from_array(variants)
	result.UnionType.PolymorphicParams = polymorphic_params
	result.UnionType.Align = align
	result.UnionType.Kind = kind
	result.UnionType.WhereToken = where_token
	result.UnionType.WhereClauses = slice_from_array(where_clauses)
	return result
}

func ast_enum_type(f *AstFile, token Token, base_type *Ast, fields []*Ast) *Ast {
	result := alloc_ast_node(f, AstEnumType)
	result.EnumType.Token = token
	result.EnumType.BaseType = base_type
	result.EnumType.Fields = slice_from_array(fields)
	return result
}

func ast_bit_set_type(f *AstFile, token Token, elem *Ast, underlying *Ast) *Ast {
	result := alloc_ast_node(f, AstBitSetType)
	result.BitSetType.Token = token
	result.BitSetType.Elem = elem
	result.BitSetType.Underlying = underlying
	return result
}

func ast_bit_field_type(f *AstFile, token Token, backing_type *Ast, open Token, fields []*Ast, close Token) *Ast {
	result := alloc_ast_node(f, AstBitFieldType)
	result.BitFieldType.Token = token
	result.BitFieldType.BackingType = backing_type
	result.BitFieldType.Open = open
	result.BitFieldType.Fields = slice_from_array(fields)
	result.BitFieldType.Close = close
	return result
}

func ast_map_type(f *AstFile, token Token, key *Ast, value *Ast) *Ast {
	result := alloc_ast_node(f, AstMapType)
	result.MapType.Token = token
	result.MapType.Key = key
	result.MapType.Value = value
	return result
}

func ast_matrix_type(f *AstFile, token Token, row_count *Ast, column_count *Ast, elem *Ast) *Ast {
	result := alloc_ast_node(f, AstMatrixType)
	result.MatrixType.Token = token
	result.MatrixType.RowCount = row_count
	result.MatrixType.ColumnCount = column_count
	result.MatrixType.Elem = elem
	return result
}

func ast_foreign_block_decl(f *AstFile, token Token, foreign_library *Ast, body *Ast,
	docs *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstForeignBlockDecl)
	result.ForeignBlockDecl.Token = token
	result.ForeignBlockDecl.ForeignLibrary = foreign_library
	result.ForeignBlockDecl.Body = body
	result.ForeignBlockDecl.Docs = docs
	return result
}

func ast_label_decl(f *AstFile, token Token, name *Ast) *Ast {
	result := alloc_ast_node(f, AstLabel)
	result.Label.Token = token
	result.Label.Name = name
	return result
}

func ast_value_decl(f *AstFile, names []*Ast, typ *Ast, values []*Ast, is_mutable bool,
	docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstValueDecl)
	result.ValueDecl.Names = slice_from_array(names)
	result.ValueDecl.Type = typ
	result.ValueDecl.Values = slice_from_array(values)
	result.ValueDecl.IsMutable = is_mutable
	result.ValueDecl.Docs = docs
	result.ValueDecl.Comment = comment
	return result
}

func ast_package_decl(f *AstFile, token Token, name Token, docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstPackageDecl)
	result.PackageDecl.Token = token
	result.PackageDecl.Name = name
	result.PackageDecl.Docs = docs
	result.PackageDecl.Comment = comment
	return result
}

func ast_import_decl(f *AstFile, token Token, relpath Token, import_name Token,
	docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstImportDecl)
	result.ImportDecl.Token = token
	result.ImportDecl.Relpath = relpath
	result.ImportDecl.ImportName = import_name
	result.ImportDecl.Docs = docs
	result.ImportDecl.Comment = comment
	return result
}

func ast_foreign_import_decl(f *AstFile, token Token, filepaths []*Ast, library_name Token,
	multiple_filepaths bool,
	docs *CommentGroup, comment *CommentGroup) *Ast {
	result := alloc_ast_node(f, AstForeignImportDecl)
	result.ForeignImportDecl.Token = token
	result.ForeignImportDecl.Filepaths = slice_from_array(filepaths)
	result.ForeignImportDecl.LibraryName = library_name
	result.ForeignImportDecl.Docs = docs
	result.ForeignImportDecl.Comment = comment
	result.ForeignImportDecl.MultipleFilepaths = multiple_filepaths
	return result
}

func ast_attribute(f *AstFile, token Token, open Token, close Token, elems []*Ast) *Ast {
	result := alloc_ast_node(f, AstAttribute)
	result.Attribute.Token = token
	result.Attribute.Open = open
	result.Attribute.Elems = slice_from_array(elems)
	result.Attribute.Close = close
	return result
}

func ast_bad_expr(f *AstFile, begin Token, end Token) *Ast {
	result := alloc_ast_node(f, AstBadExpr)
	result.BadExpr.Begin = begin
	result.BadExpr.End = end
	return result
}

func ast_tag_expr(f *AstFile, token Token, name Token, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstTagExpr)
	result.TagExpr.Token = token
	result.TagExpr.Name = name
	result.TagExpr.Expr = expr
	return result
}

func ast_unary_expr(f *AstFile, op Token, expr *Ast) *Ast {
	result := alloc_ast_node(f, AstUnaryExpr)
	if expr != nil {
		switch expr.Kind {
		case AstOrReturnExpr:
			syntax_error_with_verbose(expr, "'or_return' within an unary expression not wrapped in parentheses (...)")
		case AstOrBranchExpr:
			syntax_error_with_verbose(expr, "'%.*s' within an unary expression not wrapped in parentheses (...)", expr.OrBranchExpr.Token.String)
		}
	}
	result.UnaryExpr.Op = op
	result.UnaryExpr.Expr = expr
	return result
}

func ast_binary_expr(f *AstFile, op Token, left *Ast, right *Ast) *Ast {
	result := alloc_ast_node(f, AstBinaryExpr)
	if left == nil {
		syntax_error(op, "No lhs expression for binary expression '"+
			goStr(String{Data: op.String.Data, Len: op.String.Len})+"'")
		left = ast_bad_expr(f, op, op)
	}
	if right == nil {
		syntax_error(op, "No rhs expression for binary expression '"+
			goStr(String{Data: op.String.Data, Len: op.String.Len})+"'")
		right = ast_bad_expr(f, op, op)
	}
	if left != nil {
		switch left.Kind {
		case AstOrReturnExpr:
			syntax_error_with_verbose(left, "'or_return' within a binary expression not wrapped in parentheses (...)")
		case AstOrBranchExpr:
			syntax_error_with_verbose(left, "'%.*s' within a binary expression not wrapped in parentheses (...)", left.OrBranchExpr.Token.String)
		}
	}
	if right != nil {
		switch right.Kind {
		case AstOrReturnExpr:
			syntax_error_with_verbose(right, "'or_return' within a binary expression not wrapped in parentheses (...)")
		case AstOrBranchExpr:
			syntax_error_with_verbose(right, "'%.*s' within a binary expression not wrapped in parentheses (...)", right.OrBranchExpr.Token.String)
		}
	}
	result.BinaryExpr.Op = op
	result.BinaryExpr.Left = left
	result.BinaryExpr.Right = right
	return result
}

func ast_paren_expr(f *AstFile, expr *Ast, open Token, close Token) *Ast {
	result := alloc_ast_node(f, AstParenExpr)
	result.ParenExpr.Expr = expr
	result.ParenExpr.Open = open
	result.ParenExpr.Close = close
	return result
}

func ast_call_expr(f *AstFile, proc *Ast, args []*Ast, open Token, close Token, ellipsis Token) *Ast {
	result := alloc_ast_node(f, AstCallExpr)
	result.CallExpr.Proc = proc
	result.CallExpr.Args = slice_from_array(args)
	result.CallExpr.Open = open
	result.CallExpr.Close = close
	result.CallExpr.Ellipsis = ellipsis
	return result
}

func ast_selector_expr(f *AstFile, token Token, expr *Ast, selector *Ast) *Ast {
	result := alloc_ast_node(f, AstSelectorExpr)
	result.SelectorExpr.Token = token
	result.SelectorExpr.Expr = expr
	result.SelectorExpr.Selector = selector
	return result
}

func ast_implicit_selector_expr(f *AstFile, token Token, selector *Ast) *Ast {
	result := alloc_ast_node(f, AstImplicitSelectorExpr)
	result.ImplicitSelectorExpr.Token = token
	result.ImplicitSelectorExpr.Selector = selector
	return result
}

func ast_selector_call_expr(f *AstFile, token Token, expr *Ast, call *Ast) *Ast {
	result := alloc_ast_node(f, AstSelectorCallExpr)
	result.SelectorCallExpr.Token = token
	result.SelectorCallExpr.Expr = expr
	result.SelectorCallExpr.Call = call
	return result
}

func ast_index_expr(f *AstFile, expr *Ast, index *Ast, open Token, close Token) *Ast {
	result := alloc_ast_node(f, AstIndexExpr)
	result.IndexExpr.Expr = expr
	result.IndexExpr.Index = index
	result.IndexExpr.Open = open
	result.IndexExpr.Close = close
	return result
}

func ast_slice_expr(f *AstFile, expr *Ast, open Token, close Token, interval Token, low *Ast, high *Ast) *Ast {
	result := alloc_ast_node(f, AstSliceExpr)
	result.SliceExpr.Expr = expr
	result.SliceExpr.Open = open
	result.SliceExpr.Close = close
	result.SliceExpr.Interval = interval
	result.SliceExpr.Low = low
	result.SliceExpr.High = high
	return result
}

func ast_deref_expr(f *AstFile, expr *Ast, op Token) *Ast {
	result := alloc_ast_node(f, AstDerefExpr)
	result.DerefExpr.Expr = expr
	result.DerefExpr.Op = op
	return result
}

func ast_matrix_index_expr(f *AstFile, expr *Ast, open Token, close Token, interval Token, row *Ast, column *Ast) *Ast {
	result := alloc_ast_node(f, AstMatrixIndexExpr)
	result.MatrixIndexExpr.Expr = expr
	result.MatrixIndexExpr.RowIndex = row
	result.MatrixIndexExpr.ColumnIndex = column
	result.MatrixIndexExpr.Open = open
	result.MatrixIndexExpr.Close = close
	return result
}
