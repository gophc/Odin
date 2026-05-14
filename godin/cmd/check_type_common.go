package cmd

import "unsafe"

func does_field_type_allow_using(t *Type) bool {
	t = base_type(t)
	if is_type_struct(t) {
		return true
	} else if is_type_array(t) {
		return t.Array.Count <= 4
	} else if is_type_bit_field(t) {
		return true
	}
	return false
}

func check_custom_align(ctx *CheckerContext, node *Ast, align_ *int64, msg string) bool {
	var o Operand
	check_expr(ctx, &o, node)
	if o.Mode != Addressing_Constant {
		if o.Mode != Addressing_Invalid {
			error(node, "#%s must be a constant", msg)
		}
		return false
	}
	typ := base_type(o.Type)
	if is_type_untyped(typ) || is_type_integer(typ) {
		if o.Value.Kind == ExactValue_Integer {
			v := o.Value.ValueInteger
			if v.Used > 1 {
				str := big_int_to_string(heap_allocator(), &v)
				error(node, "#%s too large, %.*s", msg, str.Len, str.Text)
				gb_string_free(&str)
				return false
			}
			align := big_int_to_i64(&v)
			if align < 1 || !gb_is_power_of_two(isize(align)) {
				error(node, "#%s must be a power of 2, got %lld", msg, align)
				return false
			}
			*align_ = align
			return true
		}
	}
	error(node, "#%s must be an integer", msg)
	return false
}

func check_constant_parameter_value(typ *Type, expr *Ast) bool {
	if !is_type_constant_type(typ) {
		str := type_to_string(typ)
		defer gb_string_free(&str)
		error(expr, "A parameter must be a valid constant type, got %s", str)
		return true
	}
	return false
}

func is_type_valid_bit_set_range(t *Type) bool {
	if is_type_integer(t) {
		return true
	}
	if is_type_rune(t) {
		return true
	}
	return false
}

func is_expr_from_a_parameter(ctx *CheckerContext, expr *Ast) bool {
	if expr == nil {
		return false
	}
	expr = unparen_expr(expr)
	if expr.Kind == Ast_SelectorExpr {
		return is_expr_from_a_parameter(ctx, expr.SelectorExpr.Expr)
	} else if expr.Kind == Ast_Ident {
		var x Operand
		e := check_ident(ctx, &x, expr, nil, nil, true)
		if e != nil && (e.Flags&EntityFlag_Param) != 0 {
			return true
		}
	}
	return false
}

func is_caller_expression(expr *Ast) bool {
	if expr.Kind == Ast_BasicDirective && expr.BasicDirective.Name.String == "caller_expression" {
		return true
	}
	call := unparen_expr(expr)
	if call.Kind != Ast_CallExpr {
		return false
	}
	ce := &call.CallExpr
	if ce.Proc.Kind != Ast_BasicDirective {
		return false
	}
	bd := &ce.Proc.BasicDirective
	return bd.Name.String == "caller_expression"
}

func ast_references_poly_params(scope *Scope, node *Ast) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case Ast_Ident:
		e := scope_lookup(scope, node.Ident.Interned, node.Ident.Hash)
		if e != nil && e.Kind == Entity_TypeName && e.Type != nil && e.Type.Kind == Type_Generic {
			return true
		}
		return false
	case Ast_SelectorExpr:
		return ast_references_poly_params(scope, node.SelectorExpr.Expr)
	case Ast_IndexExpr:
		return ast_references_poly_params(scope, node.IndexExpr.Expr)
	case Ast_CallExpr:
		for _, arg := range node.CallExpr.Args {
			if ast_references_poly_params(scope, arg) {
				return true
			}
		}
		return ast_references_poly_params(scope, node.CallExpr.Proc)
	case Ast_CompoundLit:
		return ast_references_poly_params(scope, node.CompoundLit.Type)
	case Ast_UnaryExpr:
		return ast_references_poly_params(scope, node.UnaryExpr.Expr)
	case Ast_ParenExpr:
		return ast_references_poly_params(scope, node.ParenExpr.Expr)
	case Ast_DerefExpr:
		return ast_references_poly_params(scope, node.DerefExpr.Expr)
	case Ast_PointerType:
		return ast_references_poly_params(scope, node.PointerType.Type)
	case Ast_ArrayType:
		return ast_references_poly_params(scope, node.ArrayType.Elem) || ast_references_poly_params(scope, node.ArrayType.Count)
	case Ast_DynamicArrayType:
		return ast_references_poly_params(scope, node.DynamicArrayType.Elem)
	case Ast_FixedCapacityDynamicArrayType:
		return ast_references_poly_params(scope, node.FixedCapacityDynamicArrayType.Elem) || ast_references_poly_params(scope, node.FixedCapacityDynamicArrayType.Capacity)
	case Ast_MapType:
		return ast_references_poly_params(scope, node.MapType.Key) || ast_references_poly_params(scope, node.MapType.Value)
	}
	return false
}

func make_optional_ok_type(value *Type, typed ...bool) *Type {
	t := alloc_type_tuple()
	isTyped := true
	if len(typed) > 0 {
		isTyped = typed[0]
	}
	t.Tuple.Variables = make([]*Entity, 2)
	t.Tuple.Variables[0] = alloc_entity_field(nil, BlankToken, value, false, 0)
	boolType := tBool
	if !isTyped {
		boolType = tUntypedBool
	}
	t.Tuple.Variables[1] = alloc_entity_field(nil, BlankToken, boolType, false, 1)
	return t
}

func alignFormula(size, alignment int64) int64 {
	return (size + alignment - 1) & ^(alignment - 1)
}

func map_cell_size_and_len(type_ *Type, size_, len_ *int64) {
	elem_sz := type_size_of(type_)
	len_val := int64(1)
	if 0 < elem_sz && elem_sz < MAP_CELL_CACHE_LINE_SIZE {
		len_val = MAP_CELL_CACHE_LINE_SIZE / elem_sz
	}
	size := alignFormula(elem_sz*len_val, MAP_CELL_CACHE_LINE_SIZE)
	if size_ != nil {
		*size_ = size
	}
	if len_ != nil {
		*len_ = len_val
	}
}

type SoaTypeWorkerData struct {
	Ctx          CheckerContext
	Type         *Type
	WaitToFinish bool
}

func complete_soa_type_worker(data unsafe.Pointer) isize {
	wd := (*SoaTypeWorkerData)(data)
	complete_soa_type(wd.Ctx.checker, wd.Type, wd.WaitToFinish)
	return 0
}
