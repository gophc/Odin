package odingo

import "base:runtime"

// Forward references — defined in other odingo files:
//   Ast, AstFile, AstKind, Token, ast_node_size, alloc_ast_node

clone_ast_array :: proc(array: [dynamic]^Ast, f: ^AstFile) -> [dynamic]^Ast {
	result: [dynamic]^Ast
	if len(array) > 0 {
		result = make([dynamic]^Ast, len(array))
		for v, i in array {
			result[i] = clone_ast(v, f)
		}
	}
	return result
}

clone_ast_slice :: proc(array: []^Ast, f: ^AstFile) -> []^Ast {
	result: []^Ast
	if len(array) > 0 {
		result = make([]^Ast, len(array))
		for v, i in array {
			result[i] = clone_ast(v, f)
		}
	}
	return result
}

clone_ast :: proc(node: ^Ast, f: ^AstFile) -> ^Ast {
	if node == nil {
		return nil
	}
	if f == nil {
		f = node.file()
	}
	n := alloc_ast_node(f, node.kind)
	runtime.mem_copy(n, node, ast_node_size(node.kind))
	switch n.kind {
	case .Invalid:
	case .Ident:
		n.Ident.entity = nil
	case .Implicit:
	case .Uninit:
	case .BasicLit:
	case .BasicDirective:
	case .PolyType:
		n.PolyType.type           = clone_ast(n.PolyType.type, f)
		n.PolyType.specialization = clone_ast(n.PolyType.specialization, f)
	case .Ellipsis:
		n.Ellipsis.expr = clone_ast(n.Ellipsis.expr, f)
	case .ProcGroup:
		n.ProcGroup.args = clone_ast_slice(n.ProcGroup.args, f)
	case .ProcLit:
		n.ProcLit.type = clone_ast(n.ProcLit.type, f)
		n.ProcLit.body = clone_ast(n.ProcLit.body, f)
		n.ProcLit.where_clauses = clone_ast_slice(n.ProcLit.where_clauses, f)
	case .CompoundLit:
		n.CompoundLit.type  = clone_ast(n.CompoundLit.type, f)
		n.CompoundLit.elems = clone_ast_slice(n.CompoundLit.elems, f)
	case .BadExpr:
	case .TagExpr:
		n.TagExpr.expr = clone_ast(n.TagExpr.expr, f)
	case .UnaryExpr:
		n.UnaryExpr.expr = clone_ast(n.UnaryExpr.expr, f)
	case .BinaryExpr:
		n.BinaryExpr.left  = clone_ast(n.BinaryExpr.left, f)
		n.BinaryExpr.right = clone_ast(n.BinaryExpr.right, f)
	case .ParenExpr:
		n.ParenExpr.expr = clone_ast(n.ParenExpr.expr, f)
	case .SelectorExpr:
		n.SelectorExpr.expr = clone_ast(n.SelectorExpr.expr, f)
		n.SelectorExpr.selector = clone_ast(n.SelectorExpr.selector, f)
	case .ImplicitSelectorExpr:
		n.ImplicitSelectorExpr.selector = clone_ast(n.ImplicitSelectorExpr.selector, f)
	case .SelectorCallExpr:
		n.SelectorCallExpr.expr = clone_ast(n.SelectorCallExpr.expr, f)
		n.SelectorCallExpr.call = clone_ast(n.SelectorCallExpr.call, f)
	case .IndexExpr:
		n.IndexExpr.expr  = clone_ast(n.IndexExpr.expr, f)
		n.IndexExpr.index = clone_ast(n.IndexExpr.index, f)
	case .MatrixIndexExpr:
		n.MatrixIndexExpr.expr         = clone_ast(n.MatrixIndexExpr.expr, f)
		n.MatrixIndexExpr.row_index    = clone_ast(n.MatrixIndexExpr.row_index, f)
		n.MatrixIndexExpr.column_index = clone_ast(n.MatrixIndexExpr.column_index, f)
	case .DerefExpr:
		n.DerefExpr.expr = clone_ast(n.DerefExpr.expr, f)
	case .SliceExpr:
		n.SliceExpr.expr = clone_ast(n.SliceExpr.expr, f)
		n.SliceExpr.low  = clone_ast(n.SliceExpr.low, f)
		n.SliceExpr.high = clone_ast(n.SliceExpr.high, f)
	case .CallExpr:
		n.CallExpr.proc = clone_ast(n.CallExpr.proc, f)
		n.CallExpr.args = clone_ast_slice(n.CallExpr.args, f)
	case .FieldValue:
		n.FieldValue.field = clone_ast(n.FieldValue.field, f)
		n.FieldValue.value = clone_ast(n.FieldValue.value, f)
	case .EnumFieldValue:
		n.EnumFieldValue.name  = clone_ast(n.EnumFieldValue.name, f)
		n.EnumFieldValue.value = clone_ast(n.EnumFieldValue.value, f)
	case .TernaryIfExpr:
		n.TernaryIfExpr.x    = clone_ast(n.TernaryIfExpr.x, f)
		n.TernaryIfExpr.cond = clone_ast(n.TernaryIfExpr.cond, f)
		n.TernaryIfExpr.y    = clone_ast(n.TernaryIfExpr.y, f)
	case .TernaryWhenExpr:
		n.TernaryWhenExpr.x    = clone_ast(n.TernaryWhenExpr.x, f)
		n.TernaryWhenExpr.cond = clone_ast(n.TernaryWhenExpr.cond, f)
		n.TernaryWhenExpr.y    = clone_ast(n.TernaryWhenExpr.y, f)
	case .OrElseExpr:
		n.OrElseExpr.x = clone_ast(n.OrElseExpr.x, f)
		n.OrElseExpr.y = clone_ast(n.OrElseExpr.y, f)
	case .OrReturnExpr:
		n.OrReturnExpr.expr = clone_ast(n.OrReturnExpr.expr, f)
	case .OrBranchExpr:
		n.OrBranchExpr.label = clone_ast(n.OrBranchExpr.label, f)
		n.OrBranchExpr.expr  = clone_ast(n.OrBranchExpr.expr, f)
	case .TypeAssertion:
		n.TypeAssertion.expr = clone_ast(n.TypeAssertion.expr, f)
		n.TypeAssertion.type = clone_ast(n.TypeAssertion.type, f)
	case .TypeCast:
		n.TypeCast.type = clone_ast(n.TypeCast.type, f)
		n.TypeCast.expr = clone_ast(n.TypeCast.expr, f)
	case .AutoCast:
		n.AutoCast.expr = clone_ast(n.AutoCast.expr, f)
	case .InlineAsmExpr:
		n.InlineAsmExpr.param_types        = clone_ast_slice(n.InlineAsmExpr.param_types, f)
		n.InlineAsmExpr.return_type        = clone_ast(n.InlineAsmExpr.return_type, f)
		n.InlineAsmExpr.asm_string         = clone_ast(n.InlineAsmExpr.asm_string, f)
		n.InlineAsmExpr.constraints_string = clone_ast(n.InlineAsmExpr.constraints_string, f)
	case .BadStmt:
	case .EmptyStmt:
	case .ExprStmt:
		n.ExprStmt.expr = clone_ast(n.ExprStmt.expr, f)
	case .AssignStmt:
		n.AssignStmt.lhs = clone_ast_slice(n.AssignStmt.lhs, f)
		n.AssignStmt.rhs = clone_ast_slice(n.AssignStmt.rhs, f)
	case .BlockStmt:
		n.BlockStmt.label = clone_ast(n.BlockStmt.label, f)
		n.BlockStmt.stmts = clone_ast_slice(n.BlockStmt.stmts, f)
	case .IfStmt:
		n.IfStmt.label     = clone_ast(n.IfStmt.label, f)
		n.IfStmt.init      = clone_ast(n.IfStmt.init, f)
		n.IfStmt.cond      = clone_ast(n.IfStmt.cond, f)
		n.IfStmt.body      = clone_ast(n.IfStmt.body, f)
		n.IfStmt.else_stmt = clone_ast(n.IfStmt.else_stmt, f)
	case .WhenStmt:
		n.WhenStmt.cond      = clone_ast(n.WhenStmt.cond, f)
		n.WhenStmt.body      = clone_ast(n.WhenStmt.body, f)
		n.WhenStmt.else_stmt = clone_ast(n.WhenStmt.else_stmt, f)
	case .ReturnStmt:
		n.ReturnStmt.results = clone_ast_slice(n.ReturnStmt.results, f)
	case .ForStmt:
		n.ForStmt.label = clone_ast(n.ForStmt.label, f)
		n.ForStmt.init  = clone_ast(n.ForStmt.init, f)
		n.ForStmt.cond  = clone_ast(n.ForStmt.cond, f)
		n.ForStmt.post  = clone_ast(n.ForStmt.post, f)
		n.ForStmt.body  = clone_ast(n.ForStmt.body, f)
	case .RangeStmt:
		n.RangeStmt.label = clone_ast(n.RangeStmt.label, f)
		n.RangeStmt.init  = clone_ast(n.RangeStmt.init, f)
		n.RangeStmt.vals  = clone_ast_slice(n.RangeStmt.vals, f)
		n.RangeStmt.expr  = clone_ast(n.RangeStmt.expr, f)
		n.RangeStmt.body  = clone_ast(n.RangeStmt.body, f)
	case .UnrollRangeStmt:
		n.UnrollRangeStmt.args = clone_ast_slice(n.UnrollRangeStmt.args, f)
		n.UnrollRangeStmt.init = clone_ast(n.UnrollRangeStmt.init, f)
		n.UnrollRangeStmt.val0 = clone_ast(n.UnrollRangeStmt.val0, f)
		n.UnrollRangeStmt.val1 = clone_ast(n.UnrollRangeStmt.val1, f)
		n.UnrollRangeStmt.expr = clone_ast(n.UnrollRangeStmt.expr, f)
		n.UnrollRangeStmt.body = clone_ast(n.UnrollRangeStmt.body, f)
	case .CaseClause:
		n.CaseClause.list            = clone_ast_slice(n.CaseClause.list, f)
		n.CaseClause.stmts           = clone_ast_slice(n.CaseClause.stmts, f)
		n.CaseClause.implicit_entity = nil
	case .SwitchStmt:
		n.SwitchStmt.label = clone_ast(n.SwitchStmt.label, f)
		n.SwitchStmt.init  = clone_ast(n.SwitchStmt.init, f)
		n.SwitchStmt.tag   = clone_ast(n.SwitchStmt.tag, f)
		n.SwitchStmt.body  = clone_ast(n.SwitchStmt.body, f)
	case .TypeSwitchStmt:
		n.TypeSwitchStmt.label = clone_ast(n.TypeSwitchStmt.label, f)
		n.TypeSwitchStmt.tag   = clone_ast(n.TypeSwitchStmt.tag, f)
		n.TypeSwitchStmt.body  = clone_ast(n.TypeSwitchStmt.body, f)
	case .DeferStmt:
		n.DeferStmt.stmt = clone_ast(n.DeferStmt.stmt, f)
	case .BranchStmt:
		n.BranchStmt.label = clone_ast(n.BranchStmt.label, f)
	case .UsingStmt:
		n.UsingStmt.list = clone_ast_slice(n.UsingStmt.list, f)
	case .BadDecl:
	case .ForeignBlockDecl:
		n.ForeignBlockDecl.foreign_library = clone_ast(n.ForeignBlockDecl.foreign_library, f)
		n.ForeignBlockDecl.body            = clone_ast(n.ForeignBlockDecl.body, f)
		n.ForeignBlockDecl.attributes      = clone_ast_slice(n.ForeignBlockDecl.attributes, f)
	case .Label:
		n.Label.name = clone_ast(n.Label.name, f)
	case .ValueDecl:
		n.ValueDecl.names      = clone_ast_slice(n.ValueDecl.names, f)
		n.ValueDecl.type       = clone_ast(n.ValueDecl.type, f)
		n.ValueDecl.values     = clone_ast_slice(n.ValueDecl.values, f)
		n.ValueDecl.attributes = clone_ast_slice(n.ValueDecl.attributes, f)
	case .Attribute:
		n.Attribute.elems = clone_ast_slice(n.Attribute.elems, f)
	case .Field:
		n.Field.names = clone_ast_slice(n.Field.names, f)
		n.Field.type  = clone_ast(n.Field.type, f)
	case .BitFieldField:
		n.BitFieldField.name     = clone_ast(n.BitFieldField.name, f)
		n.BitFieldField.type     = clone_ast(n.BitFieldField.type, f)
		n.BitFieldField.bit_size = clone_ast(n.BitFieldField.bit_size, f)
	case .FieldList:
		n.FieldList.list = clone_ast_slice(n.FieldList.list, f)
	case .TypeidType:
		n.TypeidType.specialization = clone_ast(n.TypeidType.specialization, f)
	case .HelperType:
		n.HelperType.type = clone_ast(n.HelperType.type, f)
	case .DistinctType:
		n.DistinctType.type = clone_ast(n.DistinctType.type, f)
	case .ProcType:
		n.ProcType.params  = clone_ast(n.ProcType.params, f)
		n.ProcType.results = clone_ast(n.ProcType.results, f)
	case .RelativeType:
		n.RelativeType.tag  = clone_ast(n.RelativeType.tag, f)
		n.RelativeType.type = clone_ast(n.RelativeType.type, f)
	case .PointerType:
		n.PointerType.type = clone_ast(n.PointerType.type, f)
		n.PointerType.tag  = clone_ast(n.PointerType.tag, f)
	case .MultiPointerType:
		n.MultiPointerType.type = clone_ast(n.MultiPointerType.type, f)
	case .ArrayType:
		n.ArrayType.count = clone_ast(n.ArrayType.count, f)
		n.ArrayType.elem  = clone_ast(n.ArrayType.elem, f)
		n.ArrayType.tag   = clone_ast(n.ArrayType.tag, f)
	case .DynamicArrayType:
		n.DynamicArrayType.elem = clone_ast(n.DynamicArrayType.elem, f)
		n.DynamicArrayType.tag  = clone_ast(n.DynamicArrayType.tag, f)
	case .FixedCapacityDynamicArrayType:
		n.FixedCapacityDynamicArrayType.elem     = clone_ast(n.FixedCapacityDynamicArrayType.elem, f)
		n.FixedCapacityDynamicArrayType.capacity = clone_ast(n.FixedCapacityDynamicArrayType.capacity, f)
		n.FixedCapacityDynamicArrayType.tag      = clone_ast(n.FixedCapacityDynamicArrayType.tag, f)
	case .StructType:
		n.StructType.fields             = clone_ast_slice(n.StructType.fields, f)
		n.StructType.polymorphic_params = clone_ast(n.StructType.polymorphic_params, f)
		n.StructType.align              = clone_ast(n.StructType.align, f)
		n.StructType.min_field_align    = clone_ast(n.StructType.min_field_align, f)
		n.StructType.max_field_align    = clone_ast(n.StructType.max_field_align, f)
		n.StructType.where_clauses      = clone_ast_slice(n.StructType.where_clauses, f)
	case .UnionType:
		n.UnionType.variants           = clone_ast_slice(n.UnionType.variants, f)
		n.UnionType.polymorphic_params = clone_ast(n.UnionType.polymorphic_params, f)
		n.UnionType.where_clauses      = clone_ast_slice(n.UnionType.where_clauses, f)
	case .EnumType:
		n.EnumType.base_type = clone_ast(n.EnumType.base_type, f)
		n.EnumType.fields    = clone_ast_slice(n.EnumType.fields, f)
	case .BitSetType:
		n.BitSetType.elem       = clone_ast(n.BitSetType.elem, f)
		n.BitSetType.underlying = clone_ast(n.BitSetType.underlying, f)
	case .BitFieldType:
		n.BitFieldType.backing_type = clone_ast(n.BitFieldType.backing_type, f)
		n.BitFieldType.fields       = clone_ast_slice(n.BitFieldType.fields, f)
	case .MapType:
		n.MapType.count = clone_ast(n.MapType.count, f)
		n.MapType.key   = clone_ast(n.MapType.key, f)
		n.MapType.value = clone_ast(n.MapType.value, f)
	case .MatrixType:
		n.MatrixType.row_count    = clone_ast(n.MatrixType.row_count, f)
		n.MatrixType.column_count = clone_ast(n.MatrixType.column_count, f)
		n.MatrixType.elem         = clone_ast(n.MatrixType.elem, f)
	case:
		panic("clone_ast: unhandled Ast kind")
	}
	return n
}
