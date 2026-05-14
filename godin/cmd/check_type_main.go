package cmd

func check_type_internal(ctx *CheckerContext, e *Ast, typeOut **Type, named_type *Type) bool {
	if typeOut == nil {
		return false
	}
	if e == nil {
		*typeOut = t_invalid
		return true
	}

	switch e.Kind {
	case Ast_Ident:
		o := Operand{}
		check_ident(ctx, &o, e, named_type, nil, false)
		var err_str gbString
		defer func() { gb_string_free(err_str) }()

		switch o.Mode {
		case Addressing_Invalid:
		case Addressing_Type:
			*typeOut = o.Type
			if !ctx.InPolymorphicSpecialization {
				t := base_type(o.Type)
				if t != nil && is_type_polymorphic_record_unspecialized(t) {
					err_str = expr_to_string(e)
					error(e, "Invalid use of a non-specialized polymorphic type '%s'", goStr(err_str))
					return true
				}
			}
			return true
		case Addressing_NoValue:
			err_str = expr_to_string(e)
			error(e, "'%s' used as a type", goStr(err_str))
		default:
			err_str = expr_to_string(e)
			error(e, "'%s' used as a type when not a type", goStr(err_str))
		}

	case Ast_HelperType:
		return check_type_internal(ctx, e.HelperType.Type, typeOut, named_type)

	case Ast_DistinctType:
		error(e, "Invalid use of a distinct type")
		return check_type_internal(ctx, e.DistinctType.Type, typeOut, named_type)

	case Ast_TypeidType:
		e.TAV.Mode = Addressing_Type
		e.TAV.Type = t_typeid
		*typeOut = t_typeid
		set_base_type(named_type, *typeOut)
		return true

	case Ast_PolyType:
		ident := e.PolyType.Type
		if ident.Kind != Ast_Ident {
			error(ident, "Expected an identifier after the $")
			*typeOut = t_invalid
			return false
		}
		token := ident.Ident.Token
		var specific *Type = nil
		if e.PolyType.Specialization != nil {
			c := *ctx
			c.InPolymorphicSpecialization = true
			specific = check_type(&c, e.PolyType.Specialization)
		}
		t := alloc_type_generic(ctx.Scope, 0, ident.Ident.Interned, specific)
		if ctx.AllowPolymorphicTypes {
			if ctx.DisallowPolymorphicReturnTypes {
				error(ident, "Undeclared polymorphic parameter '%s' in return type", goStr(token.String))
			}
			e_entity := alloc_entity_type_name(nil, token, t)
			t.Generic.Entity = e_entity
			e_entity.TypeName.IsTypeAlias = true
			e_entity.State = EntityState_Resolved
			add_entity(ctx, ctx.PolymorphicScope, ident, e_entity)
			add_entity(ctx, ctx.Scope, ident, e_entity)
		} else {
			error(ident, "Invalid use of a polymorphic parameter '$%s'", goStr(token.String))
			*typeOut = t_invalid
			return false
		}
		*typeOut = t
		set_base_type(named_type, *typeOut)
		return true

	case Ast_SelectorExpr:
		o := Operand{}
		check_selector(ctx, &o, e, nil)
		switch o.Mode {
		case Addressing_Invalid:
		case Addressing_Type:
			*typeOut = o.Type
			if o.Type != nil {
				return true
			}
		case Addressing_NoValue:
			err_str := expr_to_string(e)
			defer gb_string_free(err_str)
			error(e, "'%s' used as a type", goStr(err_str))
		default:
			err_str := expr_to_string(e)
			defer gb_string_free(err_str)
			error(e, "'%s' is not a type", goStr(err_str))
		}

	case Ast_ParenExpr:
		*typeOut = check_type_expr(ctx, e.ParenExpr.Expr, named_type)
		set_base_type(named_type, *typeOut)
		return true

	case Ast_UnaryExpr:
		if e.UnaryExpr.Op.Kind == TokenPointer {
			elem := check_type(ctx, e.UnaryExpr.Expr)
			*typeOut = alloc_type_pointer(elem)
			set_base_type(named_type, *typeOut)
			return true
		}

	case Ast_PointerType:
		{
			c := *ctx
			c.TypePath = new_checker_type_path()
			defer destroy_checker_type_path(c.TypePath)

			elem := check_type_expr(&c, e.PointerType.Type, nil)

			if c.DisallowPolymorphicReturnTypes && is_type_polymorphic(elem, true) {
				err_str := expr_to_string(e.PointerType.Type)
				defer gb_string_free(err_str)
				error(e.PointerType.Type, "Undeclared polymorphic parameter '%s' in return type", goStr(err_str))
			}

			tag := e.PointerType.Tag
			if tag != nil {
				if tag.Kind == Ast_Ident && tag.Ident.Token.String == "soa" {
					*typeOut = alloc_type_soa_pointer(elem)
					set_base_type(named_type, *typeOut)
					return true
				}
				error(tag, "Invalid pointer type tag")
			}

			*typeOut = alloc_type_pointer(elem)
			set_base_type(named_type, *typeOut)
			return true
		}

	case Ast_MultiPointerType:
		*typeOut = alloc_type_multi_pointer(check_type(ctx, e.MultiPointerType.Type))
		set_base_type(named_type, *typeOut)
		return true

	case Ast_RelativeType:
		error(e, "#relative types have been removed; use #soa instead")
		*typeOut = t_invalid
		set_base_type(named_type, *typeOut)
		return true

	case Ast_ArrayType:
		check_array_type_internal(ctx, e, typeOut, named_type)
		set_base_type(named_type, *typeOut)
		return true

	case Ast_DynamicArrayType:
		{
			tag := e.DynamicArrayType.Tag
			if tag != nil {
				if tag.Kind == Ast_Ident && tag.Ident.Token.String == "soa" {
					elem := check_type(ctx, e.DynamicArrayType.Elem)
					t := alloc_type_struct()
					t.Struct.SoaKind = StructSoaDynamic
					t.Struct.SoaElem = elem
					*typeOut = t
					set_base_type(named_type, *typeOut)
					return true
				}
				error(tag, "Invalid dynamic array type tag")
			}
			elem := check_type(ctx, e.DynamicArrayType.Elem)
			*typeOut = alloc_type_dynamic_array(elem)
			set_base_type(named_type, *typeOut)
			return true
		}

	case Ast_FixedCapacityDynamicArrayType:
		{
			tag := e.FixedCapacityDynamicArrayType.Tag
			if tag != nil {
				if tag.Kind == Ast_Ident && tag.Ident.Token.String == "soa" {
					elem := check_type(ctx, e.FixedCapacityDynamicArrayType.Elem)
					t := alloc_type_struct()
					t.Struct.SoaKind = StructSoaFixed
					t.Struct.SoaElem = elem
					*typeOut = t
					set_base_type(named_type, *typeOut)
					return true
				}
				error(tag, "Invalid fixed capacity dynamic array type tag")
			}
			elem := check_type(ctx, e.FixedCapacityDynamicArrayType.Elem)
			*typeOut = alloc_type_fixed_capacity_dynamic_array(elem, 0, nil)
			set_base_type(named_type, *typeOut)
			return true
		}

	case Ast_StructType:
		{
			c := *ctx
			c.InPolymorphicSpecialization = false
			c.TypeLevel += 1
			*typeOut = alloc_type_struct()
			set_base_type(named_type, *typeOut)
			check_open_scope(&c, e)
			check_struct_type(&c, *typeOut, e, nil, named_type)
			check_close_scope(&c)
			return true
		}

	case Ast_UnionType:
		{
			c := *ctx
			c.InPolymorphicSpecialization = false
			c.TypeLevel += 1
			*typeOut = alloc_type_union()
			set_base_type(named_type, *typeOut)
			check_open_scope(&c, e)
			check_union_type(&c, *typeOut, e, nil, named_type)
			check_close_scope(&c)
			return true
		}

	case Ast_EnumType:
		{
			c := *ctx
			c.InPolymorphicSpecialization = false
			c.TypeLevel += 1
			*typeOut = alloc_type_enum()
			set_base_type(named_type, *typeOut)
			check_open_scope(&c, e)
			check_enum_type(&c, *typeOut, e, nil, named_type)
			check_close_scope(&c)
			return true
		}

	case Ast_BitSetType:
		*typeOut = alloc_type_bit_set()
		set_base_type(named_type, *typeOut)
		check_bit_set_type(ctx, *typeOut, named_type, e)
		return true

	case Ast_BitFieldType:
		{
			*typeOut = alloc_type_bit_field()
			set_base_type(named_type, *typeOut)
			check_open_scope(ctx, e)
			check_bit_field_type(ctx, *typeOut, e, nil)
			check_close_scope(ctx)
			return true
		}

	case Ast_ProcType:
		{
			c := *ctx
			c.TypeLevel += 1
			*typeOut = alloc_type(TypeProc)
			set_base_type(named_type, *typeOut)
			check_procedure_type(&c, *typeOut, e, nil)
			return true
		}

	case Ast_MapType:
		*typeOut = alloc_type(TypeMap)
		set_base_type(named_type, *typeOut)
		check_map_type(ctx, *typeOut, e)
		return true

	case Ast_CallExpr:
		{
			o := Operand{}
			check_expr_or_type(ctx, &o, e, nil)
			if o.Mode == Addressing_Type {
				*typeOut = o.Type
				set_base_type(named_type, *typeOut)
				return true
			}
		}

	case Ast_TernaryIfExpr:
		{
			o := Operand{}
			check_expr_or_type(ctx, &o, e, nil)
			if o.Mode == Addressing_Type {
				*typeOut = o.Type
				set_base_type(named_type, *typeOut)
				return true
			}
		}

	case Ast_TernaryWhenExpr:
		{
			o := Operand{}
			check_expr_or_type(ctx, &o, e, nil)
			if o.Mode == Addressing_Type {
				*typeOut = o.Type
				set_base_type(named_type, *typeOut)
				return true
			}
		}

	case Ast_MatrixType:
		check_matrix_type(ctx, typeOut, e)
		set_base_type(named_type, *typeOut)
		return true

	default:
		{
			o := Operand{}
			check_expr_base(ctx, &o, e, nil)
			if o.Mode == Addressing_Type || (o.Mode == Addressing_Constant && o.Value.Kind == ExactValue_Typeid) {
				*typeOut = o.Type
				set_base_type(named_type, *typeOut)
				return true
			}
		}
	}

	*typeOut = t_invalid
	return false
}

func check_type(ctx *CheckerContext, e *Ast) *Type {
	c := *ctx
	c.TypePath = new_checker_type_path()
	defer destroy_checker_type_path(c.TypePath)
	return check_type_expr(&c, e, nil)
}

func check_type_expr(ctx *CheckerContext, e *Ast, named_type *Type) *Type {
	var type_ *Type = nil
	ok := check_type_internal(ctx, e, &type_, named_type)
	if !ok {
		err_str := expr_to_string(e)
		defer gb_string_free(err_str)
		error(e, "'%s' is not a type", goStr(err_str))

		if e.Kind == Ast_IndexExpr {
			index_str := expr_to_string(e.IndexExpr.Index)
			typ_str := expr_to_string(e.IndexExpr.Expr)
			error_line("\tSuggestion: Did you mean '[%s]%s'?\n", goStr(index_str), goStr(typ_str))
			gb_string_free(index_str)
			gb_string_free(typ_str)

			elem := check_type(ctx, e.IndexExpr.Expr)
			type_ = alloc_type_array(elem, 0, nil)
		} else if e.Kind == Ast_UnaryExpr && e.UnaryExpr.Op.Kind == TokenMul {
			typ_str := expr_to_string(e.UnaryExpr.Expr)
			error_line("\tSuggestion: Did you mean '^%s'?\n", goStr(typ_str))
			gb_string_free(typ_str)

			elem := check_type(ctx, e.UnaryExpr.Expr)
			type_ = alloc_type_pointer(elem)
		}
	}

	if type_ == nil {
		type_ = t_invalid
	}

	if named_type != nil {
		set_base_type(named_type, type_)
		if is_type_polymorphic(type_, true) {
			type_.Flags |= TypeFlagPolymorphic
		}
	}

	if type_ != t_invalid {
		add_type_and_value(ctx, e, Addressing_Type, type_, ExactValue{})
		check_rtti_type_disallowed_expr(e, type_, "type with RTTI is used in a context which does not allow RTTI")
	}

	return type_
}
