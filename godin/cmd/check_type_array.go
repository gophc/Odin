package cmd

import (
	"unsafe"
)

type SoaTypeWorkerData struct {
	Ctx          CheckerContext
	Type         *Type
	WaitToFinish bool
}

func check_array_count(ctx *CheckerContext, o *Operand, e *Ast) int64 {
	if e == nil {
		return 0
	}
	if e.Kind == Ast_UnaryExpr {
		op := e.UnaryExpr.Op
		if op.Kind == TokenQuestion {
			return -1
		}
		if e.UnaryExpr.Expr == nil {
			error(e, "Invalid array count '[%.*s]'", op.String.Len, op.String.Data)
			return 0
		}
	}
	check_expr_or_type(ctx, o, e)
	if o.Mode == Addressing_Type {
		ot := base_type(o.Type)
		if ot != nil && ot.Kind == Type_Generic {
			if ctx.allow_polymorphic_types {
				if ot.Generic.Specialized != nil {
					error(o.Expr, "Polymorphic array length cannot have a specialization")
				}
				return 0
			}
		}
		if is_type_enum(ot) {
			return -1
		}
	}
	if o.Mode != Addressing_Constant {
		if o.Mode != Addressing_Invalid {
			entity := entity_of_node(o.Expr)
			is_poly_type := entity != nil && entity.Kind == Entity_TypeName && entity.Type == t_typeid && (entity.Flags&EntityFlag_PolyConst) != 0
			if ctx.allow_polymorphic_types && is_poly_type {
				return 0
			}
			begin_error_block()
			defer end_error_block()
			err_str := expr_to_string(o.Expr)
			defer gb_string_free(err_str)
			error(e, "Array count must be a constant integer, got %s", goStr(err_str))
			if is_poly_type {
				error_line("\tSuggestion: 'where' clause may be required...")
			}
			o.Mode = Addressing_Invalid
			o.Type = t_invalid
		}
		return 0
	}
	type_ := core_type(o.Type)
	if is_type_untyped(type_) || is_type_integer(type_) {
		if o.Value.Kind == ExactValue_Integer {
			count := &o.Value.ValueInteger
			if big_int_is_neg(count) {
				error(e, "Negative array count")
				return 0
			}
			switch count.Used {
			case 0:
				return 0
			case 1:
				return int64(big_int_to_u64(count))
			}
			error(e, "Array count too large")
			return 0
		} else if o.Value.Kind == ExactValue_Float {
			u := uint64(o.Value.ValueFloat)
			if float64(u) == o.Value.ValueFloat {
				return int64(u)
			}
		}
	}
	error(e, "Array count must be a constant integer")
	return 0
}

func check_matrix_type(ctx *CheckerContext, type_ **Type, node *Ast) {
	mt := &node.MatrixType
	elem := check_type(ctx, mt.Elem)
	if elem == nil || elem == t_invalid {
		*type_ = t_invalid
		return
	}

	var row_count int64
	var column_count int64
	var generic_row_count *Type
	var generic_column_count *Type

	if mt.RowCount != nil {
		var o Operand
		rc := check_array_count(ctx, &o, mt.RowCount)
		if o.Mode == Addressing_Invalid {
			*type_ = t_invalid
			return
		}
		if o.Mode == Addressing_Type {
			ot := base_type(o.Type)
			if ot != nil && ot.Kind == Type_Generic {
				generic_row_count = o.Type
			}
		}
		row_count = rc
	}

	if mt.ColumnCount != nil {
		var o Operand
		cc := check_array_count(ctx, &o, mt.ColumnCount)
		if o.Mode == Addressing_Invalid {
			*type_ = t_invalid
			return
		}
		if o.Mode == Addressing_Type {
			ot := base_type(o.Type)
			if ot != nil && ot.Kind == Type_Generic {
				generic_column_count = o.Type
			}
		}
		column_count = cc
	}

	if generic_row_count == nil && generic_column_count == nil {
		if row_count < MatrixElementCountMin || row_count > MatrixElementCountMax {
			error(node, "Invalid matrix row count %d, expected a value in the range [%d, %d]", row_count, MatrixElementCountMin, MatrixElementCountMax)
			*type_ = t_invalid
			return
		}
		if column_count < MatrixElementCountMin || column_count > MatrixElementCountMax {
			error(node, "Invalid matrix column count %d, expected a value in the range [%d, %d]", column_count, MatrixElementCountMin, MatrixElementCountMax)
			*type_ = t_invalid
			return
		}
	}

	bt := base_type(elem)
	if !is_type_polymorphic(bt) {
		switch bt.Kind {
		case Type_Basic:
			if (bt.Basic.Flags & (BasicFlagInteger | BasicFlagFloat | BasicFlagComplex)) == 0 {
				error(node, "Invalid matrix element type, expected an integer, float, or complex type")
				*type_ = t_invalid
				return
			}
		default:
			error(node, "Invalid matrix element type, expected an integer, float, or complex type")
			*type_ = t_invalid
			return
		}
	}

	*type_ = alloc_type_matrix(elem, row_count, column_count, generic_row_count, generic_column_count, mt.IsRowMajor)
}

func make_soa_struct_internal(ctx *CheckerContext, array_typ_expr *Ast, elem_expr *Ast, elem *Type, count int64, generic_type *Type, soa_kind StructSoaKind) *Type {
	bt := base_type(elem)
	if bt == nil || bt == t_invalid {
		return t_invalid
	}

	if bt.Kind == Type_Struct && bt.Struct.IsRawUnion {
		error(elem_expr, "#soa may not be used with a raw_union type")
		return t_invalid
	}
	if bt.Kind == Type_Array && bt.Array.Count > 4 {
		error(elem_expr, "#soa may only be used with an array of length <= 4")
		return t_invalid
	}
	if bt.Kind != Type_Struct && bt.Kind != Type_Array {
		if bt.Kind != Type_Basic {
			error(elem_expr, "#soa may only be used with a struct, array, or a basic type")
			return t_invalid
		}
	}

	soa_type := alloc_type_struct()
	soa_type.Struct.Scope = create_scope(ctx.info, nil)
	soa_type.Struct.Node = array_typ_expr
	soa_type.Struct.SoaKind = soa_kind
	soa_type.Struct.SoaElem = elem
	soa_type.Struct.SoaCount = int32(count)

	if is_type_polymorphic(bt) {
		if bt.Kind == Type_Struct {
			soa_type.Struct.IsPolymorphic = true
		}
		return soa_type
	}

	if bt.Kind == Type_Struct {
		base_struct := bt
		for !wait_signal_is_set(&base_struct.Struct.FieldsWaitSignal) {
			wait_signal_wait(&base_struct.Struct.FieldsWaitSignal)
		}
	}

	extra_field_count := 0
	if soa_kind == StructSoaSlice {
		extra_field_count = 1
	} else if soa_kind == StructSoaDynamic {
		extra_field_count = 3
	}

	if bt.Kind == Type_Array {
		elem_count := bt.Array.Count
		soa_type.Struct.Fields = make([]*Entity, 0, int(elem_count)+extra_field_count)
		for i := int64(0); i < elem_count; i++ {
			ft := bt.Array.Elem
			var new_type *Type
			switch soa_kind {
			case StructSoaFixed:
				new_type = alloc_type_array(ft, count, nil)
			case StructSoaSlice:
				new_type = alloc_type_multi_pointer(ft)
			case StructSoaDynamic:
				new_type = alloc_type_multi_pointer(ft)
			}
			var token Token
			token.Kind = TokenIdent
			token.String = S("0")
			token.Pos = ast_token(array_typ_expr).Pos
			nf := alloc_entity_field(soa_type.Struct.Scope, token, new_type, false, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
			nf.Flags |= EntityFlag_Field
			soa_type.Struct.Fields = append(soa_type.Struct.Fields, nf)
		}
	} else if bt.Kind == Type_Struct {
		base_struct := bt
		field_count := len(base_struct.Struct.Fields)
		soa_type.Struct.Fields = make([]*Entity, 0, field_count+extra_field_count)
		for _, f := range base_struct.Struct.Fields {
			ft := f.Type
			is_using := (f.Flags & EntityFlag_Using) != 0
			var new_type *Type
			switch soa_kind {
			case StructSoaFixed:
				new_type = alloc_type_array(ft, count, nil)
			case StructSoaSlice:
				new_type = alloc_type_multi_pointer(ft)
			case StructSoaDynamic:
				new_type = alloc_type_multi_pointer(ft)
			}
			nf := alloc_entity_field(soa_type.Struct.Scope, f.Token, new_type, is_using, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
			nf.Flags |= f.Flags & EntityFlag_Field
			soa_type.Struct.Fields = append(soa_type.Struct.Fields, nf)
		}
	} else {
		soa_type.Struct.Fields = make([]*Entity, 0, 1+extra_field_count)
		ft := bt
		var new_type *Type
		switch soa_kind {
		case StructSoaFixed:
			new_type = alloc_type_array(ft, count, nil)
		case StructSoaSlice:
			new_type = alloc_type_multi_pointer(ft)
		case StructSoaDynamic:
			new_type = alloc_type_multi_pointer(ft)
		}
		var token Token
		token.Kind = TokenIdent
		token.String = S("0")
		token.Pos = ast_token(array_typ_expr).Pos
		nf := alloc_entity_field(soa_type.Struct.Scope, token, new_type, false, 0, EntityState_Resolved)
		nf.Flags |= EntityFlag_Field
		soa_type.Struct.Fields = append(soa_type.Struct.Fields, nf)
	}

	if soa_kind == StructSoaSlice {
		len_token := make_token_ident_str(S("len"))
		len_type := t_int
		len_field := alloc_entity_field(soa_type.Struct.Scope, len_token, len_type, false, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
		len_field.Flags |= EntityFlag_Field
		soa_type.Struct.Fields = append(soa_type.Struct.Fields, len_field)
	} else if soa_kind == StructSoaDynamic {
		len_token := make_token_ident_str(S("__$len"))
		len_field := alloc_entity_field(soa_type.Struct.Scope, len_token, t_int, false, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
		len_field.Flags |= EntityFlag_Field
		soa_type.Struct.Fields = append(soa_type.Struct.Fields, len_field)

		cap_token := make_token_ident_str(S("__$cap"))
		cap_field := alloc_entity_field(soa_type.Struct.Scope, cap_token, t_int, false, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
		cap_field.Flags |= EntityFlag_Field
		soa_type.Struct.Fields = append(soa_type.Struct.Fields, cap_field)

		if t_allocator != nil {
			alloc_token := make_token_ident_str(S("allocator"))
			alloc_field := alloc_entity_field(soa_type.Struct.Scope, alloc_token, t_allocator, false, int32(len(soa_type.Struct.Fields)), EntityState_Resolved)
			alloc_field.Flags |= EntityFlag_Field
			soa_type.Struct.Fields = append(soa_type.Struct.Fields, alloc_field)
		}
	}

	soa_type.Struct.Tags = make([]string, len(soa_type.Struct.Fields))

	base_type_entity := entity_of_node(elem_expr)
	if base_type_entity != nil && base_type_entity.Kind == Entity_TypeName {
		type_name := base_type_entity.Token.String
		soa_name_data := make([]byte, type_name.Len+5)
		copy(soa_name_data, "$soa_")
		copy(soa_name_data[5:], unsafe.Slice(type_name.Data, type_name.Len))
		soa_name := S(string(soa_name_data))

		soa_scope := soa_type.Struct.Scope
		if soa_scope == nil {
			soa_scope = scope_of_node(array_typ_expr)
		}
		soa_type_entity := alloc_entity_type_name(soa_scope, make_token_ident(soa_name), soa_type, EntityState_Resolved)
		_ = soa_type_entity
	}

	if soa_kind != StructSoaSlice || count > 0 {
		wd := SoaTypeWorkerData{
			Ctx:          *ctx,
			Type:         soa_type,
			WaitToFinish: true,
		}
		mpsc_enqueue(&ctx.checker.soa_types_to_complete, wd)
	}

	return soa_type
}

func complete_soa_type(checker *Checker, t *Type, wait_to_finish bool) bool {
	t = base_type(t)
	if t == nil || !is_type_soa_struct(t) {
		return true
	}

	mutex_lock(&t.Struct.SoaMutex)
	defer mutex_unlock(&t.Struct.SoaMutex)

	fs := &t.Struct.FieldsWaitSignal
	if wait_signal_is_set(fs) {
		return true
	}
	if !wait_to_finish && wait_signal_is_initialized(fs) {
		wait_signal_wait(fs)
		return true
	}
	wait_signal_initialize(fs)

	soa_kind := t.Struct.SoaKind
	field_count := 0
	extra_field_count := 0
	if soa_kind == StructSoaSlice {
		extra_field_count = 1
	} else if soa_kind == StructSoaDynamic {
		extra_field_count = 3
	}

	elem := t.Struct.SoaElem
	if elem == nil {
		wait_signal_set(fs)
		return false
	}
	bt := base_type(elem)
	if bt == nil || bt.Kind != Type_Struct {
		wait_signal_set(fs)
		return false
	}
	if bt.Struct.IsRawUnion {
		wait_signal_set(fs)
		return false
	}

	for !wait_signal_is_set(&bt.Struct.FieldsWaitSignal) {
		if wait_to_finish {
			wait_signal_wait(&bt.Struct.FieldsWaitSignal)
		} else {
			mutex_unlock(&t.Struct.SoaMutex)
			wait_signal_wait(&bt.Struct.FieldsWaitSignal)
			mutex_lock(&t.Struct.SoaMutex)
		}
	}

	field_count = len(bt.Struct.Fields)
	new_fields := make([]*Entity, 0, field_count+extra_field_count)
	count := int64(t.Struct.SoaCount)

	for _, f := range bt.Struct.Fields {
		ft := f.Type
		is_using := (f.Flags & EntityFlag_Using) != 0
		var new_type *Type
		switch soa_kind {
		case StructSoaFixed:
			new_type = alloc_type_array(ft, count, nil)
		case StructSoaSlice:
			new_type = alloc_type_multi_pointer(ft)
		case StructSoaDynamic:
			new_type = alloc_type_multi_pointer(ft)
		}
		nf := alloc_entity_field(t.Struct.Scope, f.Token, new_type, is_using, int32(len(new_fields)), EntityState_Resolved)
		nf.Flags |= f.Flags & EntityFlag_Field
		new_fields = append(new_fields, nf)
	}

	if soa_kind == StructSoaSlice {
		len_token := make_token_ident_str(S("len"))
		len_type := t_int
		len_field := alloc_entity_field(t.Struct.Scope, len_token, len_type, false, int32(len(new_fields)), EntityState_Resolved)
		len_field.Flags |= EntityFlag_Field
		new_fields = append(new_fields, len_field)
	} else if soa_kind == StructSoaDynamic {
		len_token := make_token_ident_str(S("__$len"))
		len_field := alloc_entity_field(t.Struct.Scope, len_token, t_int, false, int32(len(new_fields)), EntityState_Resolved)
		len_field.Flags |= EntityFlag_Field
		new_fields = append(new_fields, len_field)

		cap_token := make_token_ident_str(S("__$cap"))
		cap_field := alloc_entity_field(t.Struct.Scope, cap_token, t_int, false, int32(len(new_fields)), EntityState_Resolved)
		cap_field.Flags |= EntityFlag_Field
		new_fields = append(new_fields, cap_field)

		if t_allocator != nil {
			alloc_token := make_token_ident_str(S("allocator"))
			alloc_field := alloc_entity_field(t.Struct.Scope, alloc_token, t_allocator, false, int32(len(new_fields)), EntityState_Resolved)
			alloc_field.Flags |= EntityFlag_Field
			new_fields = append(new_fields, alloc_field)
		}
	}

	t.Struct.Fields = new_fields
	t.Struct.Tags = make([]string, len(new_fields))
	wait_signal_set(fs)
	return true
}

func make_soa_struct_fixed(ctx *CheckerContext, array_typ_expr *Ast, elem_expr *Ast, elem *Type, count int64, generic_type *Type) *Type {
	return make_soa_struct_internal(ctx, array_typ_expr, elem_expr, elem, count, generic_type, StructSoaFixed)
}

func make_soa_struct_slice(ctx *CheckerContext, array_typ_expr *Ast, elem_expr *Ast, elem *Type) *Type {
	return make_soa_struct_internal(ctx, array_typ_expr, elem_expr, elem, 0, nil, StructSoaSlice)
}

func make_soa_struct_dynamic_array(ctx *CheckerContext, array_typ_expr *Ast, elem_expr *Ast, elem *Type) *Type {
	return make_soa_struct_internal(ctx, array_typ_expr, elem_expr, elem, 0, nil, StructSoaDynamic)
}

func check_array_type_internal(ctx *CheckerContext, e *Ast, type_ **Type, named_type *Type) {
	at := &e.ArrayType

	if at.Count != nil {
		var o Operand
		count := check_array_count(ctx, &o, at.Count)
		if o.Mode == Addressing_Invalid {
			*type_ = t_invalid
			return
		}

		elem := check_type(ctx, at.Elem)
		if elem == nil || elem == t_invalid {
			*type_ = t_invalid
			return
		}

		if o.Mode == Addressing_Type {
			ot := base_type(o.Type)
			if ot != nil && ot.Kind == Type_Generic {
				*type_ = alloc_type_array(elem, 0, o.Type)
				return
			}
		}

		if count < 0 && o.Mode == Addressing_Constant {
			elem_bt := base_type(elem)
			if is_type_enum(elem_bt) {
				index_type := elem
				if elem_bt.Kind == Type_Enum {
					index_type = elem_bt.Enum.BaseType
				}
				count = -count
				if count > 0 {
					*type_ = alloc_type_enumerated_array(elem, index_type, nil, nil, isize(count), TokenRangeHalf)
					return
				}
			}
			count = 0
		}

		if count <= 0 {
			error(at.Count, "Invalid array count %d", count)
			*type_ = t_invalid
			return
		}

		if at.Tag != nil {
			tag := at.Tag
			if tag.Kind == Ast_Ident {
				name := tag.Ident.Token.String
				if name == "soa" {
					is_poly := is_type_polymorphic(elem) || (o.Mode == Addressing_Type && o.Type != nil && base_type(o.Type).Kind == Type_Generic)
					if is_poly {
						*type_ = make_soa_struct_fixed(ctx, e, at.Elem, elem, count, nil)
						return
					}
					*type_ = make_soa_struct_fixed(ctx, e, at.Elem, elem, count, nil)
					return
				} else if name == "simd" {
					*type_ = alloc_type_simd_vector(count, elem, nil)
					return
				}
			}
		}

		*type_ = alloc_type_array(elem, count, nil)
	} else {
		elem := check_type(ctx, at.Elem)
		if elem == nil || elem == t_invalid {
			*type_ = t_invalid
			return
		}

		if at.Tag != nil {
			tag := at.Tag
			if tag.Kind == Ast_Ident {
				name := tag.Ident.Token.String
				if name == "soa" {
					is_poly := is_type_polymorphic(elem)
					if is_poly {
						*type_ = make_soa_struct_slice(ctx, e, at.Elem, elem)
						return
					}
					*type_ = make_soa_struct_slice(ctx, e, at.Elem, elem)
					return
				}
			}
		}

		if is_type_polymorphic(elem) {
			*type_ = alloc_type_slice(elem)
			return
		}

		*type_ = alloc_type_slice(elem)
	}
}
