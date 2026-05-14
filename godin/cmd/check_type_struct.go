package cmd

func populate_using_array_index(ctx *CheckerContext, node *Ast, field *AstField, t *Type, name String, idx int32) {
	t = base_type(t)
	gbAssertHandler("Assertion Failure", "t.Kind == Type_Array", "check_type_struct.go", 0)
	interned := stringInternerInsert(name)

	e := scope_lookup_current(ctx.Scope, interned, 0)
	if e != nil {
		var str gbString
		if node != nil {
			str = expr_to_string(node)
		}
		if str != nil {
			error_pos(e.Token.Pos, "'%s' is already declared in '%s'", goStr(name), goStr(str))
		} else {
			error_pos(e.Token.Pos, "'%s' is already declared", goStr(name))
		}
	} else {
		tok := makeTokenIdent(name)
		if field != nil {
			if len(field.Names) > 0 {
				tok.Pos = ast_token(field.Names[0]).Pos
			} else {
				tok.Pos = ast_token(field.Type).Pos
			}
		}
		f := alloc_entity_array_elem(nil, tok, t.Array.Elem, idx)
		add_entity(ctx, ctx.Scope, nil, f)
	}
}

func populate_using_entity_scope(ctx *CheckerContext, node *Ast, field *AstField, t *Type, level isize) {
	if t == nil {
		return
	}
	original_type := t
	t = base_type(type_deref(t, false))

	if t.Kind == Type_Struct {
		for _, f := range t.Struct.Fields {
			gbAssertHandler("Assertion Failure", "f.Kind == Entity_Variable", "check_type_struct.go", 0)
			name := f.Token.String
			interned := entity_interned_name(f)
			e := scope_lookup_current(ctx.Scope, interned, 0)
			if e != nil && name != "_" {
				ot := type_to_string(original_type)
				if node != nil {
					str := expr_to_string(node)
					error_pos(e.Token.Pos, "'%s' is already declared in '%s', through 'using' from '%s'", goStr(name), goStr(str), goStr(ot))
					gb_string_free(str)
				} else {
					error_pos(e.Token.Pos, "'%s' is already declared, through 'using' from '%s'", goStr(name), goStr(ot))
				}
				gb_string_free(ot)
			} else {
				add_entity(ctx, ctx.Scope, nil, f)
				if f.Flags&EntityFlag_Using != 0 {
					populate_using_entity_scope(ctx, node, field, f.Type, level+1)
				}
			}
		}
	} else if t.Kind == Type_Array && t.Array.Count <= 4 {
		switch t.Array.Count {
		case 4:
			populate_using_array_index(ctx, node, field, t, makeString("w"), 3)
			populate_using_array_index(ctx, node, field, t, makeString("a"), 3)
			fallthrough
		case 3:
			populate_using_array_index(ctx, node, field, t, makeString("z"), 2)
			populate_using_array_index(ctx, node, field, t, makeString("b"), 2)
			fallthrough
		case 2:
			populate_using_array_index(ctx, node, field, t, makeString("y"), 1)
			populate_using_array_index(ctx, node, field, t, makeString("g"), 1)
			fallthrough
		case 1:
			populate_using_array_index(ctx, node, field, t, makeString("x"), 0)
			populate_using_array_index(ctx, node, field, t, makeString("r"), 0)
			fallthrough
		default:
		}
	}
}

func check_struct_fields(ctx *CheckerContext, node *Ast, fields *[]*Entity, tags *[]string, params []*Ast, init_field_capacity isize, struct_type *Type, context String) {
	gbAssertHandler("Assertion Failure", "node.Kind == Ast_StructType", "check_type_struct.go", 0)
	gbAssertHandler("Assertion Failure", "struct_type.Kind == Type_Struct", "check_type_struct.go", 0)

	variable_count := isize(0)
	for _, field := range params {
		if ast_node_expect(field, Ast_Field) {
			f := &field.Field
			variable_count += max(len(f.Names), 1)
		}
	}

	init_field_capacity = max(init_field_capacity, variable_count)
	fields_array := make([]*Entity, 0, init_field_capacity)
	tags_array := make([]string, 0, init_field_capacity)

	entities_to_use := make([]Entity, variable_count)
	entities_to_use_index := isize(0)

	field_src_index := int32(0)
	field_group_index := int32(-1)

	for _, param := range params {
		if param.Kind != Ast_Field {
			continue
		}
		field_group_index += 1

		p := &param.Field
		type_expr := p.Type
		var type_ *Type = nil
		docs := p.Docs
		comment := p.Comment

		if type_expr != nil {
			type_ = check_type_expr(ctx, type_expr, nil)
			if is_type_polymorphic(type_) {
				struct_type.Struct.IsPolymorphic = true
				type_ = nil
			}
		}
		if type_ == nil {
			error(param, "Invalid parameter type")
			type_ = t_invalid
		}
		if is_type_untyped(type_) {
			if is_type_untyped_uninit(type_) {
				error(param, "Cannot determine parameter type from ---")
			} else {
				error(param, "Cannot determine parameter type from a nil")
			}
			type_ = t_invalid
		}

		is_using := (p.Flags & FieldFlagUsing) != 0
		is_subtype := (p.Flags & FieldFlagSubtype) != 0

		for _, name := range p.Names {
			if !ast_node_expect2(name, Ast_Ident, Ast_PolyType) {
				continue
			}
			if name.Kind == Ast_PolyType {
				name = name.PolyType.Type
			}
			name_token := name.Ident.Token

			field := &entities_to_use[entities_to_use_index]
			entities_to_use_index++
			field.Kind = Entity_Variable
			field.State = EntityState_Unresolved
			field.Scope = ctx.Scope
			field.Token = name_token
			field.Type = type_
			field.Id = 1 + globalEntityId.Add(1)
			if name_token.Pos.FileID != 0 {
				field.File = threadUnsafeGetAstFileFromId(name_token.Pos.FileID)
			}

			field.InternedName = name.Ident.Interned
			field.InternedNameHash = name.Ident.Hash

			field.Flags |= EntityFlag_Field
			if is_using {
				field.Flags |= EntityFlag_Using
			}
			field.Variable.FieldIndex = field_src_index

			add_entity(ctx, ctx.Scope, name, field)
			field.Variable.FieldGroupIndex = field_group_index
			if is_subtype {
				field.Flags |= EntityFlag_Subtype
			}

			if &name == &p.Names[0] {
				field.Variable.Docs = docs
			}
			if &name == &p.Names[len(p.Names)-1] {
				field.Variable.Comment = comment
			}

			fields_array = append(fields_array, field)
			tag := p.Tag.String
			if tag.Len != 0 && !unquote_string(permanent_allocator(), &tag, 0, tag.Data[0] == '`') {
				error(&p.Tag, "Invalid string literal")
				tag = String{}
			}
			tags_array = append(tags_array, goStr(tag))

			field_src_index += 1
		}

		if is_using && len(p.Names) > 0 {
			first_type := fields_array[len(fields_array)-1].Type
			soa_ptr := is_type_soa_pointer(first_type)
			t := base_type(type_deref(first_type, false))

			if (soa_ptr || !does_field_type_allow_using(t)) &&
				len(p.Names) >= 1 &&
				p.Names[0].Kind == Ast_Ident {
				name_token := p.Names[0].Ident.Token
				type_str := type_to_string(first_type)
				error_pos(name_token.Pos, "'using' cannot be applied to the field '%s' of type '%s'", goStr(name_token.String), goStr(type_str))
				gb_string_free(type_str)
				continue
			}

			populate_using_entity_scope(ctx, node, p, type_, 1)
		}

		if is_subtype && len(p.Names) > 0 {
			first_type := fields_array[len(fields_array)-1].Type
			t := base_type(type_deref(first_type, false))

			if !does_field_type_allow_using(t) &&
				len(p.Names) >= 1 &&
				p.Names[0].Kind == Ast_Ident {
				name_token := p.Names[0].Ident.Token
				type_str := type_to_string(first_type)
				error_pos(name_token.Pos, "'subtype' cannot be applied to the field '%s' of type '%s'", goStr(name_token.String), goStr(type_str))
				gb_string_free(type_str)
			}
		}
	}

	*fields = fields_array
	*tags = tags_array
}

func check_struct_type(ctx *CheckerContext, struct_type *Type, node *Ast, poly_operands []Operand, named_type *Type, original_type_for_poly *Type) {
	gbAssertHandler("Assertion Failure", "is_type_struct(struct_type)", "check_type_struct.go", 0)
	st := &node.StructType

	context := makeString("struct")

	min_field_count := isize(0)
	for _, field := range st.Fields {
		switch field.Kind {
		case Ast_ValueDecl:
			min_field_count += len(field.ValueDecl.Names)
		case Ast_Field:
			min_field_count += len(field.Field.Names)
		}
	}

	scope_reserve(ctx.Scope, min_field_count)

	if st.IsRawUnion && min_field_count > 1 {
		struct_type.Struct.IsRawUnion = true
		context = makeString("struct #raw_union")
	}

	struct_type.Struct.Node = node
	struct_type.Struct.Scope = ctx.Scope
	struct_type.Struct.IsPacked = st.IsPacked
	struct_type.Struct.IsAllOrNone = st.IsAllOrNone

	is_polymorphic := false
	struct_type.Struct.PolymorphicParams = check_record_polymorphic_params(ctx, st.PolymorphicParams, &is_polymorphic, poly_operands)
	wait_signal_set(&struct_type.Struct.PolymorphicWaitSignal)
	struct_type.Struct.IsPolymorphic = is_polymorphic

	struct_type.Struct.IsPolySpecialized = check_record_poly_operand_specialization(ctx, struct_type, poly_operands, &struct_type.Struct.IsPolymorphic)
	if original_type_for_poly != nil {
		gbAssertHandler("Assertion Failure", "named_type != nil", "check_type_struct.go", 0)
		add_polymorphic_record_entity(ctx, node, named_type, original_type_for_poly)
	}

	if !struct_type.Struct.IsPolymorphic {
		if len(st.WhereClauses) > 0 && st.PolymorphicParams == nil {
			error(st.WhereClauses[0], "'where' clauses can only be used on structures with polymorphic parameters")
		} else {
			where_clause_ok := evaluate_where_clauses(ctx, node, ctx.Scope, &st.WhereClauses, true)
			_ = where_clause_ok
		}
		check_struct_fields(ctx, node, &struct_type.Struct.Fields, &struct_type.Struct.Tags, st.Fields, min_field_count, struct_type, context)

		if st.IsSimple {
			success := true
			for _, f := range struct_type.Struct.Fields {
				if !is_type_nearly_simple_compare(f.Type) {
					s := type_to_string(f.Type)
					error_pos(f.Token.Pos, "'struct #simple' requires all fields to be at least 'nearly simple compare', got %s", goStr(s))
					gb_string_free(s)
				}
			}
			if success {
				struct_type.Struct.IsSimple = true
			}
		}

		wait_signal_set(&struct_type.Struct.FieldsWaitSignal)
	}

	if st.Align != nil {
		if st.IsPacked {
			error(st.Align, "'#align' cannot be applied with '#packed'")
			return
		}
		align := int64(1)
		if check_custom_align(ctx, st.Align, &align, "align") {
			struct_type.Struct.CustomAlign = align
		}
	}
	if st.MinFieldAlign != nil {
		if st.IsPacked {
			error(st.MinFieldAlign, "'#min_field_align' cannot be applied with '#packed'")
			return
		}
		align := int64(1)
		if check_custom_align(ctx, st.MinFieldAlign, &align, "min_field_align") {
			struct_type.Struct.CustomMinFieldAlign = align
		}
	}
	if st.MaxFieldAlign != nil {
		if st.IsPacked {
			error(st.MaxFieldAlign, "'#max_field_align' cannot be applied with '#packed'")
			return
		}
		align := int64(1)
		if check_custom_align(ctx, st.MaxFieldAlign, &align, "max_field_align") {
			struct_type.Struct.CustomMaxFieldAlign = align
		}
	}
	if struct_type.Struct.CustomAlign != 0 &&
		struct_type.Struct.CustomAlign < struct_type.Struct.CustomMinFieldAlign {
		error(st.Align, "#align(%lld) is defined to be less than #min_field_align(%lld)", struct_type.Struct.CustomAlign, struct_type.Struct.CustomMinFieldAlign)
	}
	if struct_type.Struct.CustomMaxFieldAlign != 0 &&
		struct_type.Struct.CustomAlign > struct_type.Struct.CustomMaxFieldAlign {
		error(st.Align, "#align(%lld) is defined to be greater than #max_field_align(%lld)", struct_type.Struct.CustomAlign, struct_type.Struct.CustomMaxFieldAlign)
	}
	if struct_type.Struct.CustomMaxFieldAlign != 0 &&
		struct_type.Struct.CustomMinFieldAlign > struct_type.Struct.CustomMaxFieldAlign {
		error(st.Align, "#min_field_align(%lld) is defined to be greater than #max_field_align(%lld)", struct_type.Struct.CustomMinFieldAlign, struct_type.Struct.CustomMaxFieldAlign)

		a := min(struct_type.Struct.CustomMinFieldAlign, struct_type.Struct.CustomMaxFieldAlign)
		b := max(struct_type.Struct.CustomMinFieldAlign, struct_type.Struct.CustomMaxFieldAlign)
		struct_type.Struct.CustomMinFieldAlign = a
		struct_type.Struct.CustomMaxFieldAlign = b
	}
}

func check_union_type(ctx *CheckerContext, union_type *Type, node *Ast, poly_operands []Operand, named_type *Type, original_type_for_poly *Type) {
	gbAssertHandler("Assertion Failure", "is_type_union(union_type)", "check_type_struct.go", 0)
	ut := &node.UnionType

	union_type.Union.Node = node
	union_type.Union.Scope = ctx.Scope

	is_polymorphic := false
	union_type.Union.PolymorphicParams = check_record_polymorphic_params(ctx, ut.PolymorphicParams, &is_polymorphic, poly_operands)
	wait_signal_set(&union_type.Union.PolymorphicWaitSignal)
	union_type.Union.IsPolymorphic = is_polymorphic

	union_type.Union.IsPolySpecialized = check_record_poly_operand_specialization(ctx, union_type, poly_operands, &union_type.Union.IsPolymorphic)
	if original_type_for_poly != nil {
		gbAssertHandler("Assertion Failure", "named_type != nil", "check_type_struct.go", 0)
		add_polymorphic_record_entity(ctx, node, named_type, original_type_for_poly)
	}

	if !union_type.Union.IsPolymorphic {
		if len(ut.WhereClauses) > 0 && ut.PolymorphicParams == nil {
			error(ut.WhereClauses[0], "'where' clauses can only be used on unions with polymorphic parameters")
		} else {
			where_clause_ok := evaluate_where_clauses(ctx, node, ctx.Scope, &ut.WhereClauses, true)
			_ = where_clause_ok
		}
	}

	variants := make([]*Type, 0, len(ut.Variants))

	for i, variant_node := range ut.Variants {
		t := check_type_expr(ctx, variant_node, nil)
		if union_type.Union.IsPolymorphic && len(poly_operands) == 0 {
			continue
		}
		if t != nil && t != t_invalid {
			ok := true
			t = default_type(t)
			if is_type_untyped(t) {
				ok = false
				str := type_to_string(t)
				error(variant_node, "Invalid variant type in union '%s'", goStr(str))
				gb_string_free(str)
			} else if is_type_empty_union(t) {
				base := base_type(t)
				if base == nil || base.Kind != Type_Union || base.Union.Node == nil {
					ok = false
					str := type_to_string(t)
					error(variant_node, "Invalid variant type in union '%s'", goStr(str))
					gb_string_free(str)
				}
			} else {
				for j := range variants {
					if union_variant_index_types_equal(t, variants[j]) {
						ok = false
						begin_error_block()
						str := type_to_string(t)
						error(variant_node, "Duplicate variant type '%s'", goStr(str))
						if j < len(ut.Variants) {
							error_line("\tPrevious found at %s\n", token_pos_to_string(ast_token(ut.Variants[j]).Pos))
						}
						gb_string_free(str)
						end_error_block()
						break
					}
				}
			}
			if ok {
				variants = append(variants, t)
				if ut.Kind == UnionTypeSharedNil {
					if !type_has_nil(t) {
						s := type_to_string(t)
						error(variant_node, "Each variant of a union with #shared_nil must have a 'nil' value, got %s", goStr(s))
						gb_string_free(s)
					}
				}
			}
		}
	}

	union_type.Union.Variants = variants
	union_type.Union.Kind = ut.Kind
	switch ut.Kind {
	case UnionTypeNoNil:
		if union_type.Union.IsPolymorphic && len(poly_operands) == 0 {
			gbAssertHandler("Assertion Failure", "len(variants) == 0", "check_type_struct.go", 0)
			if len(ut.Variants) != 1 {
				break
			}
		}
		if len(variants) < 2 {
			error(node, "A union with #no_nil must have at least 2 variants")
		}
	}

	if ut.Align != nil {
		custom_align := int64(1)
		if check_custom_align(ctx, ut.Align, &custom_align, "align") {
			if len(variants) == 0 {
				error(ut.Align, "An empty union cannot have a custom alignment")
			} else {
		union_type.Union.CustomAlign = custom_align
		}
	}
}

func check_enum_type(ctx *CheckerContext, enum_type *Type, named_type *Type, node *Ast) {
	et := &node.EnumType
	gbAssertHandler("Assertion Failure", "is_type_enum(enum_type)", "check_type_struct.go", 0)

	enum_type.Enum.BaseType = t_int
	enum_type.Enum.Scope = ctx.Scope

	base_type_val := t_int
	if unparen_expr(et.BaseType) != nil {
		base_type_val = check_type(ctx, et.BaseType)
	}

	if base_type_val == nil || base_type_val == t_invalid || !is_type_integer(base_type_val) {
		error(node, "Base type for enumeration must be an integer")
		return
	}
	if is_type_enum(base_type_val) {
		error(node, "Base type for enumeration cannot be another enumeration")
		return
	}
	if is_type_integer_128bit(base_type_val) {
		error(node, "Base type for enumeration cannot be a 128-bit integer")
		return
	}

	enum_type.Enum.BaseType = base_type_val

	fields := make([]*Entity, 0, len(et.Fields))

	constant_type := enum_type
	if named_type != nil {
		constant_type = named_type
	}

	iota_val := exact_value_i64(-1)
	min_value := exact_value_i64(0)
	max_value := exact_value_i64(0)
	min_value_index := isize(0)
	max_value_index := isize(0)
	min_value_set := false
	max_value_set := false

	scope_reserve(ctx.Scope, len(et.Fields))

	entities_to_use := make([]Entity, len(et.Fields))
	entities_to_use_index := isize(0)

	for i, field := range et.Fields {
		var ident *Ast = nil
		var init *Ast = nil
		entity_flags := uint32(0)
		if field.Kind != Ast_EnumFieldValue {
			error(field, "An enum field's name must be an identifier")
			continue
		}
		ident = field.EnumFieldValue.Name
		init = field.EnumFieldValue.Value
		if ident == nil || ident.Kind != Ast_Ident {
			error(field, "An enum field's name must be an identifier")
			continue
		}
		docs := field.EnumFieldValue.Docs
		comment := field.EnumFieldValue.Comment

		name := ident.Ident.Token.String

		if init != nil {
			o := Operand{}
			check_expr(ctx, &o, init)
			if o.Mode != Addressing_Constant {
				error(init, "Enumeration value must be a constant")
				o.Mode = Addressing_Invalid
			}
			if o.Mode != Addressing_Invalid {
				check_assignment(ctx, &o, constant_type, makeString("enumeration"))
			}
			if o.Mode != Addressing_Invalid {
				iota_val = o.Value
			} else {
				iota_val = exact_binary_operator_value(Token_Add, iota_val, exact_value_i64(1))
			}
		} else {
			iota_val = exact_binary_operator_value(Token_Add, iota_val, exact_value_i64(1))
			entity_flags |= EntityConstantFlag_ImplicitEnumValue
		}

		if is_blank_ident(name) {
			continue
		}

		if min_value_set {
			if compare_exact_values(Token_Gt, min_value, iota_val) {
				min_value_index = isize(i)
				min_value = iota_val
			}
		} else {
			min_value_index = isize(i)
			min_value = iota_val
			min_value_set = true
		}
		if max_value_set {
			if compare_exact_values(Token_Lt, max_value, iota_val) {
				max_value_index = isize(i)
				max_value = iota_val
			}
		} else {
			max_value_index = isize(i)
			max_value = iota_val
			max_value_set = true
		}

		e := &entities_to_use[entities_to_use_index]
		entities_to_use_index++
		token := ident.Ident.Token
		e.Kind = Entity_Constant
		e.State = EntityState_Resolved
		e.Scope = ctx.Scope
		e.Token = token
		e.Type = constant_type
		e.Id = 1 + globalEntityId.Add(1)
		if token.Pos.FileID != 0 {
			e.File = threadUnsafeGetAstFileFromId(token.Pos.FileID)
		}

		e.InternedName = ident.Ident.Interned
		e.InternedNameHash = ident.Ident.Hash

		e.Constant.Value = iota_val
		e.Identifier = ident
		e.Flags |= EntityFlag_Visited
		e.State = EntityState_Resolved
		e.Constant.Flags |= entity_flags
		e.Constant.Docs = docs
		e.Constant.Comment = comment

		interned := entity_interned_name(e)
		if scope_lookup_current(ctx.Scope, interned, 0) != nil {
			error_pos(ident.Ident.Token.Pos, "'%s' is already declared in this enumeration", goStr(name))
		} else {
			add_entity(ctx, ctx.Scope, nil, e)
			fields = append(fields, e)
			add_entity_use(ctx, field, e)
		}
	}

	enum_type.Enum.Fields = fields
	*enum_type.Enum.MinValue = min_value
	*enum_type.Enum.MaxValue = max_value
	enum_type.Enum.MinValueIndex = min_value_index
	enum_type.Enum.MaxValueIndex = max_value_index
}

func check_bit_field_type(ctx *CheckerContext, bit_field_type *Type, named_type *Type, node *Ast) {
	bf := &node.BitFieldType
	gbAssertHandler("Assertion Failure", "is_type_bit_field(bit_field_type)", "check_type_struct.go", 0)

	backing_type := check_type(ctx, bf.BackingType)
	bit_field_type.BitField.BackingType = backing_type
	if backing_type == nil {
		backing_type = t_u8
	}
	bit_field_type.BitField.Scope = ctx.Scope

	if backing_type == nil {
		error(bf.BackingType, "Backing type for a bit_field must be an integer or an array of an integer")
		return
	}
	if !is_valid_bit_field_backing_type(backing_type) {
		error(bf.BackingType, "Backing type for a bit_field must be an integer or an array of an integer")
		return
	}

	fields := make([]*Entity, 0, len(bf.Fields))
	bit_sizes := make([]uint8, 0, len(bf.Fields))
	tags := make([]string, 0, len(bf.Fields))

	maximum_bit_size := uint64(8 * type_size_of(backing_type))
	total_bit_size := uint64(0)

	for i, field := range bf.Fields {
		field_src_index := int32(i)
		if field.Kind != Ast_BitFieldField {
			error(field, "Invalid AST for a bit_field")
			continue
		}
		f := &field.BitFieldField
		if f.Name == nil || f.Name.Kind != Ast_Ident {
			error(field, "A bit_field's field name must be an identifier")
			continue
		}
		docs := f.Docs
		comment := f.Comment

		name := f.Name.Ident.Token.String
		interned := f.Name.Ident.Interned

		if f.Type == nil {
			error(field, "A bit_field's field must have a type")
			continue
		}

		type_ := check_type(ctx, f.Type)
		if type_size_of(type_) > 8 {
			error_pos(f.Type.TAV.Type.Token().Pos, "The type of a bit_field's field must be <= 8 bytes, got %lld", type_size_of(type_))
		}

		if is_type_untyped(type_) {
			s := type_to_string(type_)
			error(f.Type, "The type of a bit_field's field must be a typed integer, enum, or boolean, got %s", goStr(s))
			gb_string_free(s)
		} else if !(is_type_integer(type_) || is_type_enum(type_) || is_type_boolean(type_)) {
			s := type_to_string(type_)
			error(f.Type, "The type of a bit_field's field must be an integer, enum, or boolean, got %s", goStr(s))
			gb_string_free(s)
		}

		if f.BitSize == nil {
			error(field, "A bit_field's field must have a specified bit size")
			continue
		}

		o := Operand{}
		check_expr(ctx, &o, f.BitSize)
		if o.Mode != Addressing_Constant {
			error(f.BitSize, "A bit_field's specified bit size must be a constant")
			o.Mode = Addressing_Invalid
		}
		if o.Value.Kind == ExactValue_Float {
			o.Value = exact_value_to_integer(o.Value)
		}
		if f.BitSize.Kind == Ast_BinaryExpr && f.BitSize.BinaryExpr.Op.Kind == Token_Or {
			s := expr_to_string(f.BitSize)
			error(f.BitSize, "Wrap the expression in parentheses, e.g. (%s)", goStr(s))
			gb_string_free(s)
		}

		bit_size := o.Value

		if bit_size.Kind != ExactValue_Integer {
			s := expr_to_string(f.BitSize)
			error(f.BitSize, "Expected an integer constant value for the specified bit size, got %s", goStr(s))
			gb_string_free(s)
		}

		if scope_lookup_current(ctx.Scope, interned, 0) != nil {
			error_pos(f.Name.Ident.Token.Pos, "'%s' is already declared in this bit_field", goStr(name))
		} else {
			bit_size_i64 := exact_value_to_i64(bit_size)
			var bit_size_u8 uint8 = 0
			if bit_size_i64 <= 0 {
				error(f.BitSize, "A bit_field's specified bit size cannot be <= 0, got %lld", bit_size_i64)
				bit_size_i64 = 1
			}
			if bit_size_i64 > 64 {
				error(f.BitSize, "A bit_field's specified bit size cannot exceed 64 bits, got %lld", bit_size_i64)
				bit_size_i64 = 64
			}
			sz := int64(8 * type_size_of(type_))
			if bit_size_i64 > sz {
				error(f.BitSize, "A bit_field's specified bit size cannot exceed its type, got %lld, expect <=%lld", bit_size_i64, sz)
				bit_size_i64 = sz
			}

			bit_size_u8 = uint8(bit_size_i64)

			e := alloc_entity_field(ctx.Scope, f.Name.Ident.Token, type_, false, field_src_index)
			e.Variable.Docs = docs
			e.Variable.Comment = comment
			e.Variable.BitFieldBitSize = bit_size_u8
			e.Flags |= EntityFlag_BitFieldField

			add_entity(ctx, ctx.Scope, nil, e)
			fields = append(fields, e)
			bit_sizes = append(bit_sizes, bit_size_u8)

			tag := f.Tag.String
			if tag.Len != 0 && !unquote_string(permanent_allocator(), &tag, 0, tag.Data[0] == '`') {
				error(&f.Tag, "Invalid string literal")
				tag = String{}
			}
			tags = append(tags, goStr(tag))

			add_entity_use(ctx, field, e)

			total_bit_size += uint64(bit_size_u8)
		}
	}

	bit_offsets := make([]int64, len(fields))
	curr_offset := int64(0)
	for i := range bit_sizes {
		bit_offsets[i] = curr_offset
		curr_offset += int64(bit_sizes[i])
	}

	if total_bit_size > maximum_bit_size {
		s := type_to_string(backing_type)
		error(node, "The total bit size of a bit_field's fields (%llu) must fit into its backing type's (%s) bit size of %llu",
			total_bit_size, goStr(s), maximum_bit_size)
		gb_string_free(s)
	}

	type EndianKind uint8

	const (
		EndianUnknown EndianKind = iota
		EndianNative
		EndianLittle
		EndianBig
	)

	determine_endian_kind := func(type_ *Type) EndianKind {
		if is_type_boolean(type_) {
			return EndianUnknown
		} else if type_size_of(type_) < 2 {
			return EndianUnknown
		} else if is_type_endian_specific(type_) {
			if is_type_endian_little(type_) {
				return EndianLittle
			} else {
				return EndianBig
			}
		}
		return EndianNative
	}

	backing_type_elem := core_array_type(backing_type)
	backing_type_elem_size := type_size_of(backing_type_elem)
	backing_type_endian_kind := determine_endian_kind(backing_type_elem)
	endian_kind := EndianUnknown
	for _, f := range fields {
		field_kind := determine_endian_kind(f.Type)
		field_size := type_size_of(f.Type)

		if field_kind != 0 && backing_type_endian_kind != field_kind && field_size > 1 && backing_type_elem_size > 1 {
			error_pos(f.Token.Pos, "All 'bit_field' field types must match the same endian kind as the backing type, i.e. all native, all little, or all big")
		}

		if endian_kind == EndianUnknown {
			endian_kind = field_kind
		} else if field_kind != 0 && endian_kind != field_kind && field_size > 1 {
			error_pos(f.Token.Pos, "All 'bit_field' field types must be of the same endian variety, i.e. all native, all little, or all big")
		}
	}

	bit_field_type.BitField.Fields = fields
	bit_field_type.BitField.BitSizes = bit_sizes
	bit_field_type.BitField.BitOffsets = bit_offsets
	bit_field_type.BitField.Tags = tags
}

func check_bit_set_type(ctx *CheckerContext, type_ *Type, named_type *Type, node *Ast) {
	bs := &node.BitSetType
	gbAssertHandler("Assertion Failure", "type_.Kind == Type_BitSet", "check_type_struct.go", 0)
	type_.BitSet.Node = node
	type_.BitSet.Elem = t_invalid

	const MAX_BITS int64 = 128

	base := unparen_expr(bs.Elem)
	if is_ast_range(base) {
		be := &base.BinaryExpr
		lhs := Operand{}
		rhs := Operand{}
		check_expr(ctx, &lhs, be.Left)
		check_expr(ctx, &rhs, be.Right)
		if lhs.Mode == Addressing_Invalid || rhs.Mode == Addressing_Invalid {
			return
		}
		convert_to_typed(ctx, &lhs, rhs.Type)
		if lhs.Mode == Addressing_Invalid {
			return
		}
		convert_to_typed(ctx, &rhs, lhs.Type)
		if rhs.Mode == Addressing_Invalid {
			return
		}
		if !are_types_identical(lhs.Type, rhs.Type) {
			if lhs.Type != t_invalid &&
				rhs.Type != t_invalid {
				xt := type_to_string(lhs.Type)
				yt := type_to_string(rhs.Type)
				expr_str := expr_to_string(bs.Elem)
				error(bs.Elem, "Mismatched types in range '%s' : '%s' vs '%s'", goStr(expr_str), goStr(xt), goStr(yt))
				gb_string_free(expr_str)
				gb_string_free(yt)
				gb_string_free(xt)
			}
			return
		}

		if !is_type_valid_bit_set_range(lhs.Type) {
			str := type_to_string(lhs.Type)
			error(bs.Elem, "'%s' is invalid for an interval expression, expected an integer or rune", goStr(str))
			gb_string_free(str)
			return
		}

		if lhs.Mode != Addressing_Constant || rhs.Mode != Addressing_Constant {
			error(bs.Elem, "Intervals must be constant values")
			return
		}

		iv := exact_value_to_integer(lhs.Value)
		jv := exact_value_to_integer(rhs.Value)
		gbAssertHandler("Assertion Failure", "iv.Kind == ExactValue_Integer", "check_type_struct.go", 0)
		gbAssertHandler("Assertion Failure", "jv.Kind == ExactValue_Integer", "check_type_struct.go", 0)

		i := iv.ValueInteger
		j := jv.ValueInteger
		if big_int_cmp(&i, &j) > 0 {
			a := heap_allocator()
			si := big_int_to_string(a, &i)
			sj := big_int_to_string(a, &j)
			error(bs.Elem, "Lower interval bound larger than upper bound, %s .. %s", goStr(si), goStr(sj))
			gb_free(a, si.Text)
			gb_free(a, sj.Text)
			return
		}

		t := default_type(lhs.Type)
		if bs.Underlying != nil {
			u := check_type(ctx, bs.Underlying)
			if !is_type_integer(u) {
				ts := type_to_string(u)
				error(bs.Underlying, "Expected an underlying integer for the bit set, got %s", goStr(ts))
				gb_string_free(ts)
				if !is_valid_bit_field_backing_type(u) {
					return
				}
			}
			type_.BitSet.Underlying = u
		}

		if !check_representable_as_constant(ctx, iv, t, nil) {
			a := heap_allocator()
			s := big_int_to_string(a, &i)
			ts := type_to_string(t)
			error(bs.Elem, "%s is not representable by %s", goStr(s), goStr(ts))
			gb_string_free(ts)
			gb_free(a, s.Text)
			return
		}
		if !check_representable_as_constant(ctx, jv, t, nil) {
			a := heap_allocator()
			s := big_int_to_string(a, &j)
			ts := type_to_string(t)
			error(bs.Elem, "%s is not representable by %s", goStr(s), goStr(ts))
			gb_string_free(ts)
			gb_free(a, s.Text)
			return
		}
		lower := big_int_to_i64(&i)
		upper := big_int_to_i64(&j)

		actual_lower := lower
		bits := MAX_BITS
		if type_.BitSet.Underlying != nil {
			bits = 8 * type_size_of(type_.BitSet.Underlying)

			if lower > 0 {
				actual_lower = 0
			} else if lower < 0 {
				error(bs.Elem, "bit_set does not allow a negative lower bound (%lld) when an underlying type is set", lower)
			}
		}

		bits_required := upper - actual_lower
		switch be.Op.Kind {
		case Token_Ellipsis, Token_RangeFull:
			bits_required += 1
		}
		is_valid := true

		switch be.Op.Kind {
		case Token_Ellipsis, Token_RangeFull:
			if upper-lower >= bits {
				is_valid = false
			}
		case Token_RangeHalf:
			if upper-lower > bits {
				is_valid = false
			}
			upper -= 1
		}
		if !is_valid {
			if actual_lower != lower {
				error(bs.Elem, "bit_set range is greater than %lld bits, %lld bits are required (internally the lower bound was changed to 0 as an underlying type was set)", bits, bits_required)
			} else {
				error(bs.Elem, "bit_set range is greater than %lld bits, %lld bits are required", bits, bits_required)
			}
		}

		type_.BitSet.Elem = t
		type_.BitSet.Lower = lower
		type_.BitSet.Upper = upper
	} else {
		elem := check_type_expr(ctx, bs.Elem, nil)

		type_.BitSet.Elem = elem
		if !is_type_valid_bit_set_elem(elem) {
			error(bs.Elem, "Expected an enum type for a bit_set")
		} else {
			et := base_type(elem)
			if et.Kind == Type_Enum {
				if !is_type_integer(et.Enum.BaseType) {
					error(bs.Elem, "Enum type for bit_set must be an integer")
					return
				}
				lower := int64(^uint64(0) >> 1) + 1
				upper := int64(^uint64(0) >> 1)

				for _, e := range et.Enum.Fields {
					if e.Kind != Entity_Constant {
						continue
					}
					value := exact_value_to_integer(e.Constant.Value)
					gbAssertHandler("Assertion Failure", "value.Kind == ExactValue_Integer", "check_type_struct.go", 0)
					x := big_int_to_i64(&value.ValueInteger)
					lower = min(lower, x)
					upper = max(upper, x)
				}
				if len(et.Enum.Fields) == 0 {
					lower = 0
					upper = 0
				}

				gbAssertHandler("Assertion Failure", "lower <= upper", "check_type_struct.go", 0)

				lower_changed := false
				bits := MAX_BITS
				if bs.Underlying != nil {
					u := check_type(ctx, bs.Underlying)
					if !is_type_integer(u) {
						ts := type_to_string(u)
						error(bs.Underlying, "Expected an underlying integer for the bit set, got %s", goStr(ts))
						gb_string_free(ts)
						return
					}
					type_.BitSet.Underlying = u
					bits = 8 * type_size_of(u)

					if lower > 0 {
						lower = 0
						lower_changed = true
					} else if lower < 0 {
						s := type_to_string(elem)
						error(bs.Elem, "bit_set does not allow a negative lower bound (%lld) of the element type '%s' when an underlying type is set", lower, goStr(s))
						gb_string_free(s)
					}
				}

				if upper-lower >= bits {
					bits_required := upper - lower + 1
					if lower_changed {
						error(bs.Elem, "bit_set range is greater than %lld bits, %lld bits are required (internally the lower bound was changed to 0 as an underlying type was set)", bits, bits_required)
					} else {
						error(bs.Elem, "bit_set range is greater than %lld bits, %lld bits are required", bits, bits_required)
					}
				}

				type_.BitSet.Lower = lower
				type_.BitSet.Upper = upper
			}
		}
	}
}
