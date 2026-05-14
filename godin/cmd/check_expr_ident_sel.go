package cmd

// ---------------------------------------------------------------------------
// DidYouMean helpers
// ---------------------------------------------------------------------------

type DidYouMeanResult struct {
	Target String
}

type DidYouMeanAnswers struct {
	Results []DidYouMeanResult
	Name    String
}

func did_you_mean_make(allocator gbAllocator, capacity isize, name String) DidYouMeanAnswers {
	return DidYouMeanAnswers{
		Results: make([]DidYouMeanResult, 0, capacity),
		Name:    name,
	}
}

func did_you_mean_append(d *DidYouMeanAnswers, target String) {
	d.Results = append(d.Results, DidYouMeanResult{Target: target})
}

func did_you_mean_destroy(d *DidYouMeanAnswers) {
	d.Results = nil
}

func did_you_mean_results(d *DidYouMeanAnswers) []DidYouMeanResult {
	return d.Results
}

func check_did_you_mean_print(d *DidYouMeanAnswers, prefix string) {
	limit := buildContext.DidYouMeanLimit
	results := did_you_mean_results(d)
	count := 0
	if len(results) != 0 {
		error_line("\tSuggestion: Did you mean?\n")
		for _, result := range results {
			target := result.Target
			error_line("\t\t%s%.*s\n", prefix, isize(target.Len), target.Text)
			if limit > 0 {
				count++
				if count == limit {
					error_line("\t\t... and %td more ...", isize(len(results))-isize(limit))
					break
				}
			}
		}
	}
}

func check_did_you_mean_type(name String, fields []*Entity, prefix string) {
	if buildContext.TerseErrors {
		return
	}
	begin_error_block()
	defer end_error_block()
	d := did_you_mean_make(heap_allocator(), isize(len(fields)), name)
	defer did_you_mean_destroy(&d)
	for _, e := range fields {
		did_you_mean_append(&d, e.Token.String)
	}
	check_did_you_mean_print(&d, prefix)
}

func check_did_you_mean_scope(name String, scope *Scope, prefix string) {
	if buildContext.TerseErrors {
		return
	}
	begin_error_block()
	defer end_error_block()
	d := did_you_mean_make(heap_allocator(), isize(len(scope.Elements)), name)
	defer did_you_mean_destroy(&d)
	for _, e := range scope.Elements {
		did_you_mean_append(&d, e.Token.String)
	}
	check_did_you_mean_print(&d, prefix)
}

func populate_check_did_you_mean_objc_entity(set map[String]struct{}, e *Entity, is_type bool) {
	if e.Kind != Entity_TypeName {
		return
	}
	if e.TypeName.ObjcMetadata == nil {
		return
	}
	objc_metadata := e.TypeName.ObjcMetadata
	t := base_type(e.Type)
	_ = t
	entries := objc_metadata.ValueEntries
	if is_type {
		entries = objc_metadata.TypeEntries
	}
	for _, entry := range entries {
		set[entry.Interned.string()] = struct{}{}
	}
	for _, f := range t.Struct.Fields {
		if f.Flags&EntityFlag_Using != 0 && f.Type != nil {
			if f.Type.Kind == Type_Named && f.Type.Named.TypeName != nil {
				populate_check_did_you_mean_objc_entity(set, f.Type.Named.TypeName, is_type)
			}
		}
	}
}

func check_did_you_mean_objc_entity(name String, e *Entity, is_type bool, prefix string) {
	if buildContext.TerseErrors {
		return
	}
	begin_error_block()
	defer end_error_block()

	if e.Kind != Entity_TypeName {
		return
	}
	if e.TypeName.ObjcMetadata == nil {
		return
	}

	seen := make(map[String]struct{})
	populate_check_did_you_mean_objc_entity(seen, e, is_type)

	names := make([]String, 0, len(seen))
	for k := range seen {
		names = append(names, k)
	}

	d := did_you_mean_make(heap_allocator(), isize(len(names)), name)
	defer did_you_mean_destroy(&d)
	for _, target := range names {
		did_you_mean_append(&d, target)
	}
	check_did_you_mean_print(&d, prefix)
}

// ---------------------------------------------------------------------------
// entity_from_expr
// ---------------------------------------------------------------------------

func entity_from_expr(expr *Ast) *Entity {
	expr = unparen_expr(expr)
	if expr == nil {
		return nil
	}
	switch expr.Kind {
	case AstIdent:
		return expr.Ident.Entity.Load()
	case AstSelectorExpr:
		return entity_from_expr(expr.SelectorExpr.Selector)
	}
	return nil
}

// ---------------------------------------------------------------------------
// error_operand_not_expression
// ---------------------------------------------------------------------------

func error_operand_not_expression(o *Operand) {
	if o.Mode == Addressing_Type {
		err := expr_to_string(o.Expr)
		error(o.Expr, "'%s' is not an expression but a type", err)
		gb_string_free(err)
		o.Mode = Addressing_Invalid
	}
}

// ---------------------------------------------------------------------------
// error_operand_no_value
// ---------------------------------------------------------------------------

func error_operand_no_value(o *Operand) {
	if o.Mode == Addressing_NoValue {
		x := unparen_expr(o.Expr)
		if x != nil && x.Kind == AstCallExpr {
			p := unparen_expr(x.CallExpr.Proc)
			if p.Kind == AstBasicDirective {
				tag := p.BasicDirective.Name.String
				if tag == "panic" ||
					tag == "assert" {
					return
				}
			}
		}
		err := expr_to_string(o.Expr)
		if x != nil && x.Kind == AstCallExpr {
			error(o.Expr, "'%s' call does not return a value and cannot be used as a value", err)
		} else {
			error(o.Expr, "'%s' used as a value", err)
		}
		gb_string_free(err)
		o.Mode = Addressing_Invalid
	}
}

// ---------------------------------------------------------------------------
// check_ident
// ---------------------------------------------------------------------------

func check_ident(c *CheckerContext, o *Operand, n *Ast, named_type *Type, type_hint *Type, allow_import_name bool) *Entity {
	o.Mode = Addressing_Invalid
	o.Expr = n
	name := n.Ident.Token.String
	e := scope_lookup(c.Scope, n.Ident.Interned, n.Ident.Hash)
	if e == nil {
		if is_blank_ident(n) {
			error(n, "'_' cannot be used as a value")
		} else {
			begin_error_block()
			error(n, "Undeclared name: %.*s", isize(name.Len), name.Text)
			for _, suggestion := range CIdentSuggestions {
				if name == suggestion.Name {
					error_line("\tSuggestion: Did you mean %.*s\n", isize(suggestion.Msg.Len), suggestion.Msg.Text)
				}
			}
			end_error_block()
		}
		o.Type = t_invalid
		o.Mode = Addressing_Invalid
		if named_type != nil {
			set_base_type(named_type, t_invalid)
		}
		return nil
	}

	if e.ParentProcDecl != nil &&
		e.ParentProcDecl != c.CurrProcDecl {
		if e.Kind == Entity_Variable {
			if e.Flags&EntityFlag_Static == 0 {
				error(n, "Nested procedures do not capture its parent's variables: %.*s", isize(name.Len), name.Text)
				return nil
			}
		} else if e.Kind == Entity_Label {
			error(n, "Nested procedures do not capture its parent's labels: %.*s", isize(name.Len), name.Text)
			return nil
		}
	}

	if e.Kind == Entity_ProcGroup {
		pge := &e.ProcGroup
		d := decl_info_of_entity(e)
		check_entity_decl(c, e, d, nil)

		procs := pge.Entities
		skip := false
		if type_hint != nil && is_type_proc(type_hint) {
			for _, proc := range procs {
				t := base_type(proc.Type)
				if t == t_invalid {
					continue
				}
				x := Operand{}
				x.Mode = Addressing_Value
				x.Type = t
				if check_is_assignable_to(c, &x, type_hint) {
					e = proc
					add_entity_use(c, n, e)
					skip = true
					break
				}
			}
		}
		if !skip {
			o.Mode = Addressing_ProcGroup
			o.Type = t_invalid
			o.ProcGroup = e
			return nil
		}
	}

	add_entity_use(c, n, e)
	if e.State == EntityState_Unresolved {
		check_entity_decl(c, e, nil, named_type)
	}

	switch e.Kind {
	case Entity_Constant, Entity_Variable, Entity_TypeName:
		if check_cycle(c, e, true) {
			o.Type = t_invalid
			return e
		}
	}

	if e.Type == nil {
		return nil
	}

	e.Flags |= EntityFlag_Used
	type_ := e.Type
	o.Type = type_

	switch e.Kind {
	case Entity_Constant:
		if type_ == t_invalid {
			o.Type = t_invalid
			return e
		}
		o.Value = e.Constant.Value
		if o.Value.Kind == ExactValue_Invalid {
			return e
		}
		if o.Value.Kind == ExactValue_Procedure {
			proc := strip_entity_wrapping(o.Value.ValueProcedure)
			if proc != nil {
				o.Mode = Addressing_Value
				o.Type = proc.Type
				return proc
			}
		}
		o.Mode = Addressing_Constant

	case Entity_Variable:
		e.Flags |= EntityFlag_Used
		if type_ == t_invalid {
			o.Type = t_invalid
			return e
		}
		o.Mode = Addressing_Variable
		if e.Flags&EntityFlag_Value != 0 {
			o.Mode = Addressing_Value
		}

	case Entity_Procedure:
		o.Mode = Addressing_Value
		o.Value = exact_value_procedure(n)

	case Entity_Builtin:
		o.BuiltinID = (BuiltinProcId)(e.Builtin.ID)
		o.Mode = Addressing_Builtin

	case Entity_TypeName:
		o.Mode = Addressing_Type
		if check_cycle(c, e, true) {
			o.Type = t_invalid
		}
		if o.Type != nil && o.Type.Kind == Type_Named && o.Type.Named.TypeName.TypeName.IsTypeAlias {
			bt := base_type(o.Type)
			if bt != nil {
				o.Type = bt
			}
		}

	case Entity_ImportName:
		if !allow_import_name {
			error(n, "Use of import name '%.*s' not in the form of 'x.y'", isize(name.Len), name.Text)
		}
		return e

	case Entity_LibraryName:
		if !allow_import_name {
			error(n, "Use of library '%.*s' not in foreign block", isize(name.Len), name.Text)
		}
		return e

	case Entity_Label:
		o.Mode = Addressing_NoValue

	case Entity_Nil:
		o.Mode = Addressing_Value

	default:
		compiler_error("Unknown EntityKind %.*s", isize(entityStrings[e.Kind].Len), entityStrings[e.Kind].Data)
	}

	return e
}

// ---------------------------------------------------------------------------
// determine_swizzle_array_type
// ---------------------------------------------------------------------------

func determine_swizzle_array_type(original_type *Type, type_hint *Type, new_count isize) *Type {
	array_type := base_type(type_deref(original_type))

	if array_type.Kind == Type_SimdVector {
		elem_type := array_type.SimdVector.Elem
		return alloc_type_simd_vector(int64(new_count), elem_type, nil)
	}

	elem_type := array_type.Array.Elem
	var swizzle_array_type *Type
	bth := base_type(type_deref(type_hint))
	if bth != nil && bth.Kind == Type_Array &&
		bth.Array.Count == int64(new_count) &&
		are_types_identical(bth.Array.Elem, elem_type) {
		swizzle_array_type = type_hint
	} else {
		max_count := array_type.Array.Count
		if new_count == isize(max_count) {
			swizzle_array_type = original_type
		} else {
			swizzle_array_type = alloc_type_array(elem_type, int64(new_count), nil)
		}
	}
	return swizzle_array_type
}

// ---------------------------------------------------------------------------
// is_entity_declared_for_selector
// ---------------------------------------------------------------------------

func is_entity_declared_for_selector(entity *Entity, import_scope *Scope, allow_builtin *bool) bool {
	is_declared := entity != nil
	if is_declared {
		if entity.Kind == Entity_Builtin {
			*allow_builtin = entity.Scope == import_scope ||
				(entity.Scope != builtinPkg.Scope && entity.Scope != intrinsicsPkg.Scope)
		} else if (entity.Scope.Flags&ScopeFlag_Global) == ScopeFlag_Global && (import_scope.Flags&ScopeFlag_Global) == 0 {
			is_declared = false
		}
	}
	return is_declared
}

// ---------------------------------------------------------------------------
// check_entity_from_ident_or_selector
// ---------------------------------------------------------------------------

func check_entity_from_ident_or_selector(c *CheckerContext, node *Ast, ident_only bool) *Entity {
	if node == nil {
		return nil
	}
	if node.Kind == AstIdent {
		e := node.Ident.Entity.Load()
		if e != nil {
			return e
		}
		return scope_lookup(c.Scope, node.Ident.Interned, node.Ident.Hash)
	} else if !ident_only && node.Kind == AstSelectorExpr {
		se := &node.SelectorExpr
		if se.Token.Kind == Token_ArrowRight {
			return nil
		}
		op_expr := se.Expr
		selector := unparen_expr(se.Selector)
		if selector == nil {
			return nil
		}
		if selector.Kind != AstIdent {
			return nil
		}
		var entity *Entity
		var expr_entity *Entity
		check_op_expr := true

		if op_expr.Kind == AstIdent {
			e := scope_lookup(c.Scope, op_expr.Ident.Interned, op_expr.Ident.Hash)
			if e == nil {
				return nil
			}
			add_entity_use(c, op_expr, e)
			expr_entity = e
			if e.Kind == Entity_ImportName && selector.Kind == AstIdent {
				import_scope := e.ImportName.Scope
				entity_name := selector.Ident.Interned
				check_op_expr = false
				entity = scope_lookup_current(import_scope, entity_name, entity_name.Hash())
				allow_builtin := false
				if !is_entity_declared_for_selector(entity, import_scope, &allow_builtin) {
					return nil
				}
				check_entity_decl(c, entity, nil, nil)
				if entity.Kind == Entity_ProcGroup {
					return entity
				}
			}
		}

		operand := Operand{}
		if check_op_expr {
			check_expr_base(c, &operand, op_expr, nil)
			if operand.Mode == Addressing_Invalid {
				return nil
			}
		}
		if entity == nil && selector.Kind == AstIdent {
			field_name := selector.Ident.Interned
			if is_type_dynamic_array(type_deref(operand.Type)) {
				init_mem_allocator(c.Checker)
			}
			sel := lookup_field(operand.Type, field_name, operand.Mode == Addressing_Type)
			entity = sel.Entity
		}
		if entity != nil {
			return entity
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// check_selector
// ---------------------------------------------------------------------------

func check_selector(c *CheckerContext, operand *Operand, node *Ast, type_hint *Type) *Entity {
	se := &node.SelectorExpr
	check_op_expr := true
	var expr_entity *Entity
	var entity *Entity
	sel := Selection{}

	if !c.AllowArrowRightSelectorExpr && se.Token.Kind == Token_ArrowRight {
		begin_error_block()
		error(node, "Illegal use of -> selector shorthand outside of a call")
		x := expr_to_string(se.Expr)
		y := expr_to_string(se.Selector)
		error_line("\tSuggestion: Did you mean '%s.%s'?\n", x, y)
		gb_string_free(y)
		gb_string_free(x)
		end_error_block()
	}

	operand.Expr = node
	op_expr := se.Expr
	selector := unparen_expr(se.Selector)

	if selector == nil {
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	if selector.Kind != AstIdent {
		error(selector, "Illegal selector kind: '%.*s'", isize(astStrings[selector.Kind].Len), astStrings[selector.Kind].Data)
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	if op_expr.Kind == AstIdent {
		op_name := op_expr.Ident.Token.String
		e := scope_lookup(c.Scope, op_expr.Ident.Interned, op_expr.Ident.Hash)
		add_entity_use(c, op_expr, e)
		expr_entity = e

		if e != nil && (e.Kind == Entity_Procedure || e.Kind == Entity_ProcGroup) && selector.Kind == AstIdent {
			sel_str := expr_to_string(selector)
			error(node, "'%s' is not declared by by '%.*s'", sel_str, isize(e.Token.String.Len), e.Token.String.Text)
			gb_string_free(sel_str)
			operand.Mode = Addressing_Invalid
			operand.Expr = node
			return nil
		} else if e != nil && e.Kind == Entity_ImportName && selector.Kind == AstIdent {
			import_name := op_name
			import_scope := e.ImportName.Scope
			entity_name := selector.Ident.Token.String
			entity_name_interned := selector.Ident.Interned

			if import_scope == nil {
				begin_error_block()
				error(node, "'%.*s' is not imported in this file, '%.*s' is unavailable", isize(import_name.Len), import_name.Text, isize(entity_name.Len), entity_name.Text)
				operand.Mode = Addressing_Invalid
				operand.Expr = node
				end_error_block()
				return nil
			}

			check_op_expr = false
			entity = scope_lookup_current(import_scope, entity_name_interned, entity_name_interned.Hash())
			allow_builtin := false
			if !is_entity_declared_for_selector(entity, import_scope, &allow_builtin) {
				begin_error_block()
				error(node, "'%.*s' is not declared by '%.*s'", isize(entity_name.Len), entity_name.Text, isize(import_name.Len), import_name.Text)
				operand.Mode = Addressing_Invalid
				operand.Expr = node
				check_did_you_mean_scope(entity_name, import_scope)
				end_error_block()
				return nil
			}

			if !is_entity_exported(entity, allow_builtin) {
				sel_str := expr_to_string(selector)
				error(node, "'%s' is not exported by '%.*s'", sel_str, isize(import_name.Len), import_name.Text)
				gb_string_free(sel_str)
			}

			check_entity_decl(c, entity, nil, nil)
			if entity.Kind == Entity_ProcGroup {
				operand.Mode = Addressing_ProcGroup
				operand.ProcGroup = entity
				add_type_and_value(c, operand.Expr, operand.Mode, operand.Type, operand.Value)
				return entity
			}
		}
	}

	if check_op_expr {
		check_expr_base(c, operand, op_expr, nil)
		if operand.Mode == Addressing_Invalid {
			operand.Mode = Addressing_Invalid
			operand.Expr = node
			return nil
		}
	}

	if operand.Type != nil && is_type_soa_struct(type_deref(operand.Type)) {
		complete_soa_type(c.Checker, type_deref(operand.Type), false)
	}

	if entity == nil && selector.Kind == AstIdent {
		field_name := selector.Ident.Interned
		t := type_deref(operand.Type)
		if t == nil {
			error(operand.Expr, "Cannot use a selector expression on 0-value expression")
		} else {
			if is_type_dynamic_array(t) {
				init_mem_allocator(c.Checker)
			}
			sel = lookup_field(operand.Type, field_name, operand.Mode == Addressing_Type)
			entity = sel.Entity
			if entity != nil && (entity.Flags&EntityFlag_TypeField) != 0 {
				add_type_info_type(c, operand.Type)
			}
			if is_type_enum(operand.Type) {
				add_type_info_type(c, operand.Type)
			}
		}
	}

	if operand.Type != nil && is_type_simd_vector(type_deref(operand.Type)) {
		field_name := selector.Ident.Token.String
		if field_name.Len == 1 {
			error(op_expr, "Extracting an element from a #simd array using .%.*s syntax is disallowed, prefer `simd.extract`", isize(field_name.Len), field_name.Text)
		} else {
			error(op_expr, "Extracting elements from a #simd array using .%.*s syntax is disallowed, prefer `swizzle`", isize(field_name.Len), field_name.Text)
		}
		return nil
	}

	if entity == nil && selector.Kind == AstIdent && operand.Type != nil &&
		is_type_array(type_deref(operand.Type)) {
		field_name := selector.Ident.Token.String
		if 1 < field_name.Len && field_name.Len <= 4 {
			swizzles_xyzw := [4]byte{'x', 'y', 'z', 'w'}
			swizzles_rgba := [4]byte{'r', 'g', 'b', 'a'}
			found_xyzw := false
			found_rgba := false

			for i := isize(0); i < field_name.Len; i++ {
				valid := false
				for j := isize(0); j < 4; j++ {
					if field_name.Text[i] == swizzles_xyzw[j] {
						found_xyzw = true
						valid = true
						break
					}
					if field_name.Text[i] == swizzles_rgba[j] {
						found_rgba = true
						valid = true
						break
					}
				}
				if !valid {
					goto end_of_array_selector_swizzle
				}
			}

			var swizzles []byte
			index_count := uint8(field_name.Len)
			if found_xyzw && found_rgba {
				op_str := expr_to_string(op_expr)
				error(op_expr, "Mixture of swizzle kinds for field index, got %s", op_str)
				gb_string_free(op_str)
				operand.Mode = Addressing_Invalid
				operand.Expr = node
				return nil
			}

			var indices uint8 = 0
			if found_xyzw {
				swizzles = swizzles_xyzw[:]
			} else if found_rgba {
				swizzles = swizzles_rgba[:]
			}

			for i := isize(0); i < field_name.Len; i++ {
				for j := isize(0); j < 4; j++ {
					if field_name.Text[i] == swizzles[j] {
						indices |= uint8(j) << (i * 2)
						break
					}
				}
			}

			original_type := operand.Type
			array_type := base_type(type_deref(original_type))
			array_count := get_array_type_count(array_type)

			for i := uint8(0); i < index_count; i++ {
				idx := (indices >> (i * 2)) & 3
				if int64(idx) >= array_count {
					var c byte = 0
					if found_xyzw {
						c = swizzles_xyzw[idx]
					} else if found_rgba {
						c = swizzles_rgba[idx]
					}
					error(selector.Ident.Token, "Swizzle value is out of bounds, got %c, max count %lld", c, array_count)
					break
				}
			}

			se.SwizzleCount = index_count
			se.SwizzleIndices = indices
			prev_mode := operand.Mode
			operand.Mode = Addressing_SwizzleValue
			operand.Type = determine_swizzle_array_type(original_type, type_hint, isize(index_count))
			operand.Expr = node

			switch prev_mode {
			case Addressing_Variable, Addressing_SoaVariable, Addressing_SwizzleVariable:
				operand.Mode = Addressing_SwizzleVariable
			case Addressing_Value:
				if is_type_pointer(original_type) {
					operand.Mode = Addressing_SwizzleVariable
				}
			}

			swizzle_entity := alloc_entity_variable(nil, make_token_ident(field_name), operand.Type, EntityState_Resolved)
			add_type_and_value(c, operand.Expr, operand.Mode, operand.Type, operand.Value)
			return swizzle_entity
		}
	}
end_of_array_selector_swizzle:

	if entity == nil {
		op_str := expr_to_string(op_expr)
		type_str := type_to_string_shorthand(operand.Type)
		sel_str := expr_to_string(selector)

		if operand.Mode == Addressing_Type {
			if is_type_polymorphic(operand.Type, true) {
				error(op_expr, "Type '%s' has no field nor polymorphic parameter '%s'", op_str, sel_str)
			} else {
				error(op_expr, "Type '%s' has no field '%s'", op_str, sel_str)
			}
		} else {
			begin_error_block()
			error(op_expr, "'%s' of type '%s' has no field '%s'", op_str, type_str, sel_str)
			if operand.Type != nil && selector.Kind == AstIdent {
				name := selector.Ident.Token.String
				bt := base_type(operand.Type)
				if operand.Type.Kind == Type_Named &&
					operand.Type.Named.TypeName != nil &&
					operand.Type.Named.TypeName.Kind == Entity_TypeName &&
					operand.Type.Named.TypeName.TypeName.ObjcMetadata != nil {
					check_did_you_mean_objc_entity(name, operand.Type.Named.TypeName, operand.Mode == Addressing_Type)
				} else if bt.Kind == Type_Struct {
					check_did_you_mean_type(name, bt.Struct.Fields, "")
				} else if bt.Kind == Type_Enum {
					check_did_you_mean_type(name, bt.Enum.Fields, "")
				}
			}
			end_error_block()
		}

		gb_string_free(sel_str)
		gb_string_free(type_str)
		gb_string_free(op_str)
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	if expr_entity != nil && expr_entity.Kind == Entity_Constant && entity.Kind != Entity_Constant {
		success := false
		field_value := get_constant_field(c, operand, sel, &success)
		if success {
			operand.Mode = Addressing_Constant
			operand.Expr = node
			operand.Value = field_value
			operand.Type = entity.Type
			add_entity_use(c, selector, entity)
			add_type_and_value(c, operand.Expr, operand.Mode, operand.Type, operand.Value)
			return entity
		}
		op_str := expr_to_string(op_expr)
		type_str := type_to_string_shorthand(operand.Type)
		sel_str := expr_to_string(selector)
		error(op_expr, "Cannot access non-constant field '%s' from '%s'", sel_str, op_str)
		gb_string_free(sel_str)
		gb_string_free(type_str)
		gb_string_free(op_str)
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	if operand.Mode == Addressing_Constant && entity.Kind != Entity_Constant {
		success := false
		field_value := get_constant_field(c, operand, sel, &success)
		if success {
			operand.Mode = Addressing_Constant
			operand.Expr = node
			operand.Value = field_value
			operand.Type = entity.Type
			add_entity_use(c, selector, entity)
			add_type_and_value(c, operand.Expr, operand.Mode, operand.Type, operand.Value)
			return entity
		}
		op_str := expr_to_string(op_expr)
		type_str := type_to_string_shorthand(operand.Type)
		sel_str := expr_to_string(selector)
		error(op_expr, "Cannot access non-constant field '%s' from '%s'", sel_str, op_str)
		gb_string_free(sel_str)
		gb_string_free(type_str)
		gb_string_free(op_str)
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	if expr_entity != nil && is_type_polymorphic(expr_entity.Type) {
		op_str := expr_to_string(op_expr)
		type_str := type_to_string_shorthand(operand.Type)
		sel_str := expr_to_string(selector)
		error(op_expr, "Cannot access field '%s' from non-specialized polymorphic type '%s'", sel_str, op_str)
		gb_string_free(sel_str)
		gb_string_free(type_str)
		gb_string_free(op_str)
		operand.Mode = Addressing_Invalid
		operand.Expr = node
		return nil
	}

	add_entity_use(c, selector, entity)
	operand.Type = entity.Type
	operand.Expr = node

	if entity.Flags&EntityFlag_BitFieldField != 0 {
		add_package_dependency(c, "runtime", "__write_bits")
		add_package_dependency(c, "runtime", "__read_bits")
	}

	switch entity.Kind {
	case Entity_Constant:
		operand.Value = entity.Constant.Value
		operand.Mode = Addressing_Constant
		if operand.Value.Kind == ExactValue_Procedure {
			proc := strip_entity_wrapping(operand.Value.ValueProcedure)
			if proc != nil {
				operand.Mode = Addressing_Value
				operand.Type = proc.Type
			}
		}

	case Entity_Variable:
		if sel.IsBitField {
			se.IsBitField = true
		}
		if sel.Indirect {
			operand.Mode = Addressing_Variable
		} else if operand.Mode == Addressing_Context {
			// keep context
		} else if operand.Mode == Addressing_MapIndex {
			operand.Mode = Addressing_Value
		} else if entity.Flags&EntityFlag_SoaPtrField != 0 {
			operand.Mode = Addressing_SoaVariable
		} else if operand.Mode == Addressing_OptionalOk || operand.Mode == Addressing_OptionalOkPtr {
			operand.Mode = Addressing_Value
		} else if operand.Mode == Addressing_SoaVariable {
			operand.Mode = Addressing_Variable
		} else if operand.Mode != Addressing_Value {
			operand.Mode = Addressing_Variable
		} else {
			operand.Mode = Addressing_Value
		}

	case Entity_TypeName:
		operand.Mode = Addressing_Type

	case Entity_Procedure:
		operand.Mode = Addressing_Value
		operand.Value = exact_value_procedure(node)

	case Entity_Builtin:
		operand.Mode = Addressing_Builtin
		operand.BuiltinID = (BuiltinProcId)(entity.Builtin.ID)

	case Entity_ProcGroup:
		operand.Mode = Addressing_ProcGroup
		operand.ProcGroup = entity

	case Entity_Nil:
		operand.Mode = Addressing_Value
	}

	add_type_and_value(c, operand.Expr, operand.Mode, operand.Type, operand.Value)
	return entity
}

// ---------------------------------------------------------------------------
// is_type_normal_pointer
// ---------------------------------------------------------------------------

func is_type_normal_pointer(ptr *Type, elem **Type) bool {
	ptr = base_type(ptr)
	if is_type_pointer(ptr) {
		if is_type_rawptr(ptr) {
			return false
		}
		if elem != nil {
			*elem = ptr.Pointer.Elem
		}
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// is_type_valid_atomic_type
// ---------------------------------------------------------------------------

func is_type_valid_atomic_type(elem *Type) bool {
	elem = core_type(elem)
	if is_type_internally_pointer_like(elem) {
		return true
	}
	if elem.Kind == Type_BitSet {
		elem = bit_set_to_int(elem)
	}
	if elem.Kind != Type_Basic {
		return false
	}
	return (elem.Basic.Flags & (BasicFlag_Boolean | BasicFlag_OrderedNumeric)) != 0
}

// ---------------------------------------------------------------------------
// check_identifier_exists
// ---------------------------------------------------------------------------

func check_identifier_exists(s *Scope, node *Ast, nested bool, out_scope **Scope) bool {
	switch node.Kind {
	case AstIdent:
		i := &node.Ident
		if nested {
			e := scope_lookup_current(s, i.Interned, i.Hash)
			if e != nil {
				if out_scope != nil {
					*out_scope = e.Scope
				}
				return true
			}
		} else {
			e := scope_lookup(s, i.Interned, i.Hash)
			if e != nil {
				if out_scope != nil {
					*out_scope = e.Scope
				}
				return true
			}
		}

	case AstSelectorExpr:
		se := &node.SelectorExpr
		lhs := se.Expr
		rhs := se.Selector
		var lhs_scope *Scope
		if check_identifier_exists(s, lhs, nested, &lhs_scope) {
			return check_identifier_exists(lhs_scope, rhs, true)
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// check_is_operand_compound_lit_constant
// ---------------------------------------------------------------------------

func check_is_operand_compound_lit_constant(c *CheckerContext, o *Operand, field_type *Type) bool {
	if is_operand_nil(*o) {
		return true
	}
	if is_type_any(field_type) {
		return false
	}
	if field_type != nil && is_type_typeid(field_type) && o.Mode == Addressing_Type {
		add_type_info_type(c, o.Type)
		return true
	}

	expr := unparen_expr(o.Expr)
	if expr != nil {
		e := strip_entity_wrapping(entity_from_expr(expr))
		if e != nil && e.Kind == Entity_Procedure {
			return true
		}
		if expr.Kind == AstProcLit {
			add_type_and_value(c, expr, Addressing_Constant, type_of_expr(expr), exact_value_procedure(expr))
			return true
		}
		if e != nil && e.Kind == Entity_Variable && e.Variable.IsRodata {
			// fallthrough
		}
	}
	return o.Mode == Addressing_Constant
}

// ---------------------------------------------------------------------------
// is_expr_inferred_fixed_array
// ---------------------------------------------------------------------------

func is_expr_inferred_fixed_array(type_expr *Ast) bool {
	type_expr = unparen_expr(type_expr)
	if type_expr == nil {
		return false
	}
	if type_expr.Kind == AstArrayType && type_expr.ArrayType.Count != nil {
		count := type_expr.ArrayType.Count
		if count.Kind == AstUnaryExpr &&
			count.UnaryExpr.Op.Kind == Token_Question {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// check_for_dynamic_literals
// ---------------------------------------------------------------------------

func check_for_dynamic_literals(c *CheckerContext, node *Ast, cl *AstCompoundLit) bool {
	if len(cl.Elems) == 0 {
		return false
	}

	if (check_feature_flags(c, node)&OptInFeatureFlag_DynamicLiterals) == 0 && !buildContext.DynamicLiterals {
		begin_error_block()
		error(node, "Compound literals of dynamic types are disabled by default")
		error_line("\tSuggestion: If you want to enable them for this specific file, add '#+feature dynamic-literals' at the top of the file\n")
		error_line("\tWarning: Please understand that dynamic literals will implicitly allocate using the current 'context.allocator' in that scope\n")
		if buildContext.ODIN_DEFAULT_TO_NIL_ALLOCATOR {
			error_line("\tWarning: As '-default-to-panic-allocator' has been set, the dynamic compound literal may not be initialized as expected\n")
		} else if buildContext.ODIN_DEFAULT_TO_PANIC_ALLOCATOR {
			error_line("\tWarning: As '-default-to-panic-allocator' has been set, the dynamic compound literal may not be initialized as expected\n")
		}
		end_error_block()
		return false
	} else if c.CurrProcDecl != nil && c.CurrProcCallingConvention != ProcCC_Odin {
		if c.Scope != nil && (c.Scope.Flags&ScopeFlag_ContextDefined) == 0 {
			error(node, "Compound literals of dynamic types require a 'context' to defined")
		}
	}
	return true
}
