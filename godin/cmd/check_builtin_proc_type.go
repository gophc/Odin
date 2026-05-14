package cmd

import "sort"

func checkBuiltinProc_type_is_boolean(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_bool(false)
	if operand.Mode != Addressing_Type {
		str := expr_to_string(operand.Expr)
		error(operand.Expr, "Expected a type for '%s', got '%s'", builtin_name, str)
		gb_string_free(str)
	} else {
		var ok bool
		switch BuiltinProcId(id) {
		case BuiltinProc_type_is_boolean:
			ok = is_type_boolean(operand.Type)
		case BuiltinProc_type_is_bit_field:
			ok = is_type_bit_field(operand.Type)
		case BuiltinProc_type_is_integer:
			ok = is_type_integer(operand.Type)
		case BuiltinProc_type_is_rune:
			ok = is_type_rune(operand.Type)
		case BuiltinProc_type_is_float:
			ok = is_type_float(operand.Type)
		case BuiltinProc_type_is_complex:
			ok = is_type_complex(operand.Type)
		case BuiltinProc_type_is_quaternion:
			ok = is_type_quaternion(operand.Type)
		case BuiltinProc_type_is_string:
			ok = is_type_string(operand.Type)
		case BuiltinProc_type_is_string16:
			ok = is_type_string16(operand.Type)
		case BuiltinProc_type_is_cstring:
			ok = is_type_cstring(operand.Type)
		case BuiltinProc_type_is_cstring16:
			ok = is_type_cstring16(operand.Type)
		case BuiltinProc_type_is_typeid:
			ok = is_type_typeid(operand.Type)
		case BuiltinProc_type_is_any:
			ok = is_type_any(operand.Type)
		case BuiltinProc_type_is_endian_platform:
			ok = is_type_endian_platform(operand.Type)
		case BuiltinProc_type_is_endian_little:
			ok = is_type_endian_little(operand.Type)
		case BuiltinProc_type_is_endian_big:
			ok = is_type_endian_big(operand.Type)
		case BuiltinProc_type_is_unsigned:
			ok = is_type_unsigned(operand.Type)
		case BuiltinProc_type_is_numeric:
			ok = is_type_numeric(operand.Type)
		case BuiltinProc_type_is_ordered:
			ok = is_type_ordered(operand.Type)
		case BuiltinProc_type_is_ordered_numeric:
			ok = is_type_ordered_numeric(operand.Type)
		case BuiltinProc_type_is_indexable:
			ok = is_type_indexable(operand.Type)
		case BuiltinProc_type_is_sliceable:
			ok = is_type_sliceable(operand.Type)
		case BuiltinProc_type_is_comparable:
			ok = is_type_comparable(operand.Type)
		case BuiltinProc_type_is_simple_compare:
			ok = is_type_simple_compare(operand.Type)
		case BuiltinProc_type_is_nearly_simple_compare:
			ok = is_type_nearly_simple_compare(operand.Type)
		case BuiltinProc_type_is_dereferenceable:
			ok = is_type_dereferenceable(operand.Type)
		case BuiltinProc_type_is_valid_map_key:
			ok = is_type_valid_for_keys(operand.Type)
		case BuiltinProc_type_is_valid_matrix_elements:
			ok = isTypeValidForMatrixElems(operand.Type)
		case BuiltinProc_type_is_named:
			ok = is_type_named(operand.Type)
		case BuiltinProc_type_is_pointer:
			ok = is_type_pointer(operand.Type)
		case BuiltinProc_type_is_multi_pointer:
			ok = is_type_multi_pointer(operand.Type)
		case BuiltinProc_type_is_array:
			ok = is_type_array(operand.Type)
		case BuiltinProc_type_is_enumerated_array:
			ok = is_type_enumerated_array(operand.Type)
		case BuiltinProc_type_is_slice:
			ok = is_type_slice(operand.Type)
		case BuiltinProc_type_is_dynamic_array:
			ok = is_type_dynamic_array(operand.Type)
		case BuiltinProc_type_is_map:
			ok = is_type_map(operand.Type)
		case BuiltinProc_type_is_struct:
			ok = is_type_struct(operand.Type)
		case BuiltinProc_type_is_union:
			ok = is_type_union(operand.Type)
		case BuiltinProc_type_is_enum:
			ok = is_type_enum(operand.Type)
		case BuiltinProc_type_is_proc:
			ok = is_type_proc(operand.Type)
		case BuiltinProc_type_is_bit_set:
			ok = is_type_bit_set(operand.Type)
		case BuiltinProc_type_is_simd_vector:
			ok = is_type_simd_vector(operand.Type)
		case BuiltinProc_type_is_matrix:
			ok = is_type_matrix(operand.Type)
		case BuiltinProc_type_is_raw_union:
			ok = is_type_raw_union(operand.Type)
		case BuiltinProc_type_is_specialized_polymorphic_record:
			ok = is_type_polymorphic_record_specialized(operand.Type)
		case BuiltinProc_type_is_unspecialized_polymorphic_record:
			ok = is_type_polymorphic_record_unspecialized(operand.Type)
		case BuiltinProc_type_has_nil:
			ok = type_has_nil(operand.Type)
		}
		operand.Value = exact_value_bool(ok)
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_is_matrix_row_major(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if type_.Kind != Type_Matrix {
		s := type_to_string(bt)
		error(ce.Args[0], "Expected a matrix type for '%s', got '%s'", builtin_name, s)
		gb_string_free(s)
		return false
	}
	if BuiltinProcId(id) == BuiltinProc_type_is_matrix_row_major {
		operand.Value = exact_value_bool(type_.Matrix.IsRowMajor)
	} else {
		operand.Value = exact_value_bool(!type_.Matrix.IsRowMajor)
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_has_field(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	var x Operand
	check_expr(c, &x, ce.Args[1])
	if !is_type_string(x.Type) || x.Mode != Addressing_Constant || x.Value.Kind != ExactValue_String {
		error(ce.Args[1], "Expected a constant string for field argument")
		return false
	}
	field_name := string_interner_insert(x.Value.ValueString)
	sel := lookup_field(type_, field_name, false)
	operand.Mode = Addressing_Constant
	operand.Value = exact_value_bool(sel.Entity != nil)
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_has_shared_fields(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	u := check_type(c, ce.Args[0])
	ut := base_type(u)
	if ut == nil || ut == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if ut.Kind != Type_Struct || ut.Struct.SoaKind != StructSoaNone {
		t := type_to_string(ut)
		error(ce.Args[0], "Expected a struct type for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	v := check_type(c, ce.Args[1])
	vt := base_type(v)
	if vt == nil || vt == t_invalid {
		error(ce.Args[1], "Expected a type for '%s'", builtin_name)
		return false
	}
	if vt.Kind != Type_Struct || vt.Struct.SoaKind != StructSoaNone {
		t := type_to_string(vt)
		error(ce.Args[1], "Expected a struct type for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	is_shared := true
	for _, v_field := range vt.Struct.Fields {
		found := false
		for _, u_field := range ut.Struct.Fields {
			if v_field.Token.String == u_field.Token.String &&
				are_types_identical(v_field.Type, u_field.Type) {
				found = true
				break
			}
		}
		if !found {
			is_shared = false
			break
		}
	}
	operand.Mode = Addressing_Constant
	operand.Value = exact_value_bool(is_shared)
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_field_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	var x Operand
	check_expr(c, &x, ce.Args[1])
	if !is_type_string(x.Type) || x.Mode != Addressing_Constant || x.Value.Kind != ExactValue_String {
		error(ce.Args[1], "Expected a constant string for field argument")
		return false
	}
	field_name := string_interner_insert(x.Value.ValueString)
	sel := lookup_field(type_, field_name, false)
	if sel.Entity == nil {
		t := type_to_string(type_)
		error(ce.Args[1], "'%s' is not a field of type %s", field_name, t)
		gb_string_free(t)
		return false
	}
	operand.Mode = Addressing_Type
	operand.Type = sel.Entity.Type
	return true
}

func checkBuiltinProc_type_field_bit_offset(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if !is_type_bit_field(type_) {
		error(operand.Expr, "Expected a bit field type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	var x Operand
	check_expr(c, &x, ce.Args[1])
	if !is_type_string(x.Type) || x.Mode != Addressing_Constant || x.Value.Kind != ExactValue_String {
		error(ce.Args[1], "Expected a constant string for field argument")
		return false
	}
	field_name := string_interner_insert(x.Value.ValueString)
	var bit_offset, bit_size int64
	for i, f := range type_.BitField.Fields {
		if f.Kind != Entity_Variable || (f.Flags&EntityFlag_Field) == 0 {
			continue
		}
		str := entity_interned_name(f)
		if field_name == str {
			bit_offset = type_.BitField.BitOffsets[i]
			bit_size = int64(type_.BitField.BitSizes[i])
			break
		}
	}
	var value int64
	switch BuiltinProcId(id) {
	case BuiltinProc_type_field_bit_offset:
		value = bit_offset
	case BuiltinProc_type_field_bit_size:
		value = bit_size
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	operand.Value = exact_value_i64(value)
	return true
}

func checkBuiltinProc_type_is_specialization_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	t := operand.Type
	var s *Type
	prev_ips := c.InPolymorphicSpecialization
	c.InPolymorphicSpecialization = true
	s = check_type(c, ce.Args[1])
	c.InPolymorphicSpecialization = prev_ips
	if s == t_invalid {
		error(ce.Args[1], "Invalid specialization type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	operand.Value = exact_value_bool(check_type_specialization_to(c, s, t, false, false))
	return true
}

func checkBuiltinProc_type_is_variant_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	v := check_type(c, ce.Args[1])
	u = base_type(u)
	is_variant := false
	for _, vt := range u.Union.Variants {
		if are_types_identical(v, vt) {
			is_variant = true
			break
		}
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	operand.Value = exact_value_bool(is_variant)
	return true
}

func checkBuiltinProc_type_union_tag_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u = base_type(u)
	operand.Mode = Addressing_Type
	operand.Type = union_tag_type(u)
	return true
}

func checkBuiltinProc_type_union_tag_offset(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u = base_type(u)
	type_size_of(u)
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	operand.Value = exact_value_i64(u.Union.VariantBlockSize)
	return true
}

func checkBuiltinProc_type_union_base_tag_value(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u = base_type(u)
	var tag_value int64
	if u.Union.Kind == UnionTypeNoNil {
		tag_value = 0
	} else {
		tag_value = 1
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	operand.Value = exact_value_i64(tag_value)
	return true
}

func checkBuiltinProc_type_bit_set_elem_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	bs := operand.Type
	if !is_type_bit_set(bs) {
		error(operand.Expr, "Expected a bit_set type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	bs = base_type(bs)
	operand.Mode = Addressing_Type
	operand.Type = bs.BitSet.Elem
	return true
}

func checkBuiltinProc_type_bit_set_underlying_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	bs := operand.Type
	if !is_type_bit_set(bs) {
		error(operand.Expr, "Expected a bit_set type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	bs = base_type(bs)
	operand.Mode = Addressing_Type
	operand.Type = bit_set_to_int(bs)
	return true
}

func checkBuiltinProc_type_union_variant_count(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u = base_type(u)
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	operand.Value = exact_value_i64(int64(len(u.Union.Variants)))
	return true
}

func checkBuiltinProc_type_variant_type_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u = base_type(u)
	var x Operand
	check_expr_or_type(c, &x, ce.Args[1])
	if !is_type_integer(x.Type) || x.Mode != Addressing_Constant {
		error(call, "Expected a constant integer for '%s'", builtin_name)
		operand.Mode = Addressing_Type
		operand.Type = t_invalid
		return false
	}
	index := big_int_to_i64(&x.Value.ValueInteger)
	if index < 0 || index >= int64(len(u.Union.Variants)) {
		error(call, "Variant tag out of bounds index for '%s'", builtin_name)
		operand.Mode = Addressing_Type
		operand.Type = t_invalid
		return false
	}
	operand.Mode = Addressing_Type
	operand.Type = u.Union.Variants[index]
	return true
}

func checkBuiltinProc_type_variant_index_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	u := operand.Type
	if !is_type_union(u) {
		error(operand.Expr, "Expected a union type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	v := check_type(c, ce.Args[1])
	u = base_type(u)
	index := int64(-1)
	for i, vt := range u.Union.Variants {
		if unionVariantIndexTypesEqual(v, vt) {
			index = int64(i)
			break
		}
	}
	if index < 0 {
		error(operand.Expr, "Expected a variant type for '%s'", builtin_name)
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	operand.Value = exact_value_i64(index)
	return true
}

func checkBuiltinProc_type_struct_field_count(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_i64(0)
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a struct type for '%s'", builtin_name)
	} else if !is_type_struct(operand.Type) {
		error(operand.Expr, "Expected a struct type for '%s'", builtin_name)
	} else {
		bt := base_type(operand.Type)
		operand.Value = exact_value_i64(int64(len(bt.Struct.Fields)))
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	return true
}

func checkBuiltinProc_type_struct_has_implicit_padding(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_bool(false)
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a struct type for '%s'", builtin_name)
	} else if !is_type_struct(operand.Type) && !is_type_soa_struct(operand.Type) {
		error(operand.Expr, "Expected a struct type for '%s'", builtin_name)
	} else {
		bt := base_type(operand.Type)
		if bt.Struct.IsPacked {
			operand.Value = exact_value_bool(false)
		} else if len(bt.Struct.Fields) != 0 {
			size := type_size_of(bt)
			var field_type *Type
			last_offset := type_offset_of(bt, int64(len(bt.Struct.Fields)-1), &field_type)
			if last_offset+type_size_of(field_type) < size {
				operand.Value = exact_value_bool(true)
			} else {
				packed_size := type_size_of_struct_pretend_is_packed(bt)
				operand.Value = exact_value_bool(packed_size < size)
			}
		}
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_proc_parameter_count(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_i64(0)
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
	} else if !is_type_proc(operand.Type) {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
	} else {
		bt := base_type(operand.Type)
		operand.Value = exact_value_i64(int64(bt.Proc.ParamCount))
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	return true
}

func checkBuiltinProc_type_proc_return_count(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_i64(0)
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
	} else if !is_type_proc(operand.Type) {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
	} else {
		bt := base_type(operand.Type)
		operand.Value = exact_value_i64(int64(bt.Proc.ResultCount))
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	return true
}

func checkBuiltinProc_type_proc_parameter_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type || !is_type_proc(operand.Type) {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
		return false
	} else {
		if is_type_polymorphic(operand.Type) {
			error(operand.Expr, "Expected a non-polymorphic procedure type for '%s'", builtin_name)
			return false
		}
		var op Operand
		check_expr(c, &op, ce.Args[1])
		if op.Mode != Addressing_Constant || !is_type_integer(op.Type) {
			error(op.Expr, "Expected a constant integer for the index of procedure parameter value")
			return false
		}
		index := exact_value_to_i64(op.Value)
		if index < 0 {
			error(op.Expr, "Expected a non-negative integer for the index of procedure parameter value, got %d", index)
			return false
		}
		var param *Entity
		count := int64(0)
		bt := base_type(operand.Type)
		if bt.Kind == Type_Proc {
			count = int64(bt.Proc.ParamCount)
			if index < count {
				param = bt.Proc.Params.Tuple.Variables[index]
			}
		}
		if index >= count {
			error(op.Expr, "Index of procedure parameter value out of bounds, expected 0..<%d, got %d", count, index)
			return false
		}
		switch param.Kind {
		case Entity_Constant:
			operand.Mode = Addressing_Constant
			operand.Type = param.Type
			operand.Value = param.Constant.Value
		case Entity_TypeName, Entity_Variable:
			operand.Mode = Addressing_Type
			operand.Type = param.Type
		default:
		}
	}
	return true
}

func checkBuiltinProc_type_proc_return_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type || !is_type_proc(operand.Type) {
		error(operand.Expr, "Expected a procedure type for '%s'", builtin_name)
		return false
	} else {
		if is_type_polymorphic(operand.Type) {
			error(operand.Expr, "Expected a non-polymorphic procedure type for '%s'", builtin_name)
			return false
		}
		var op Operand
		check_expr(c, &op, ce.Args[1])
		if op.Mode != Addressing_Constant || !is_type_integer(op.Type) {
			error(op.Expr, "Expected a constant integer for the index of procedure parameter value")
			return false
		}
		index := exact_value_to_i64(op.Value)
		if index < 0 {
			error(op.Expr, "Expected a non-negative integer for the index of procedure parameter value, got %d", index)
			return false
		}
		var param *Entity
		count := int64(0)
		bt := base_type(operand.Type)
		if bt.Kind == Type_Proc {
			count = int64(bt.Proc.ResultCount)
			if index < count {
				param = bt.Proc.Results.Tuple.Variables[index]
			}
		}
		if index >= count {
			error(op.Expr, "Index of procedure parameter value out of bounds, expected 0..<%d, got %d", count, index)
			return false
		}
		switch param.Kind {
		case Entity_Constant:
			operand.Mode = Addressing_Constant
			operand.Type = param.Type
			operand.Value = param.Constant.Value
		case Entity_TypeName, Entity_Variable:
			operand.Mode = Addressing_Type
			operand.Type = param.Type
		default:
		}
	}
	return true
}

func checkBuiltinProc_type_polymorphic_record_parameter_count(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	operand.Value = exact_value_i64(0)
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a record type for '%s'", builtin_name)
	} else {
		tuple := get_record_polymorphic_params(operand.Type)
		if tuple != nil {
			operand.Value = exact_value_i64(int64(len(tuple.Variables)))
		} else {
			error(operand.Expr, "Expected a record type for '%s'", builtin_name)
		}
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_integer
	return true
}

func checkBuiltinProc_type_polymorphic_record_parameter_value(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	if operand.Mode != Addressing_Type {
		error(operand.Expr, "Expected a record type for '%s'", builtin_name)
		return false
	} else if !is_type_polymorphic_record_specialized(operand.Type) {
		error(operand.Expr, "Expected a specialized polymorphic record type for '%s'", builtin_name)
		return false
	} else {
		var op Operand
		check_expr(c, &op, ce.Args[1])
		if op.Mode != Addressing_Constant || !is_type_integer(op.Type) {
			error(op.Expr, "Expected a constant integer for the index of record parameter value")
			return false
		}
		index := exact_value_to_i64(op.Value)
		if index < 0 {
			error(op.Expr, "Expected a non-negative integer for the index of record parameter value, got %d", index)
			return false
		}
		var param *Entity
		count := int64(0)
		tuple := get_record_polymorphic_params(operand.Type)
		if tuple != nil {
			count = int64(len(tuple.Variables))
			if index < count {
				param = tuple.Variables[index]
			}
		} else {
			error(operand.Expr, "Expected a specialized polymorphic record type for '%s'", builtin_name)
			return false
		}
		if index >= count {
			error(op.Expr, "Index of record parameter value out of bounds, expected 0..<%d, got %d", count, index)
			return false
		}
		switch param.Kind {
		case Entity_Constant:
			operand.Mode = Addressing_Constant
			operand.Type = param.Type
			operand.Value = param.Constant.Value
		case Entity_TypeName:
			operand.Mode = Addressing_Type
			operand.Type = param.Type
		default:
		}
	}
	return true
}

func checkBuiltinProc_type_is_subtype_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var op_src, op_dst Operand
	check_expr_or_type(c, &op_src, ce.Args[0])
	if op_src.Mode != Addressing_Type {
		e := expr_to_string(op_src.Expr)
		error(op_src.Expr, "'%s' expects a type, got %s", builtin_name, e)
		gb_string_free(e)
		return false
	}
	check_expr_or_type(c, &op_dst, ce.Args[1])
	if op_dst.Mode != Addressing_Type {
		e := expr_to_string(op_dst.Expr)
		error(op_dst.Expr, "'%s' expects a type, got %s", builtin_name, e)
		gb_string_free(e)
		return false
	}
	operand.Value = exact_value_bool(isTypeSubtypeOfAndAllowPolymorphic(op_src.Type, op_dst.Type))
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	return true
}

func checkBuiltinProc_type_is_superset_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var op_super, op_sub Operand
	check_expr_or_type(c, &op_super, ce.Args[0])
	if op_super.Mode != Addressing_Type {
		e := expr_to_string(op_super.Expr)
		error(op_super.Expr, "'%s' expects a type, got %s", builtin_name, e)
		gb_string_free(e)
		return false
	}
	check_expr_or_type(c, &op_sub, ce.Args[1])
	if op_sub.Mode != Addressing_Type {
		e := expr_to_string(op_sub.Expr)
		error(op_sub.Expr, "'%s' expects a type, got %s", builtin_name, e)
		gb_string_free(e)
		return false
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	super := op_super.Type
	sub := op_sub.Type
	if are_types_identical(super, sub) {
		operand.Value = exact_value_bool(true)
		return true
	}
	super = base_type(super)
	sub = base_type(sub)
	if are_types_identical(super, sub) {
		operand.Value = exact_value_bool(true)
		return true
	}
	if super.Kind != sub.Kind {
		a := type_to_string(op_super.Type)
		b := type_to_string(op_sub.Type)
		error(op_super.Expr, "'%s' expects types of the same kind, got %s vs %s", builtin_name, a, b)
		gb_string_free(b)
		gb_string_free(a)
		return false
	}
	if super.Kind == Type_Enum {
		if len(sub.Enum.Fields) > len(super.Enum.Fields) {
			operand.Value = exact_value_bool(false)
			return true
		}
		base_super := base_enum_type(super)
		base_sub := base_enum_type(sub)
		if base_super == base_sub && base_super == nil {
		} else if !are_types_identical(base_type(base_super), base_type(base_sub)) {
			operand.Value = exact_value_bool(false)
			return true
		}
		for _, f_sub := range sub.Enum.Fields {
			found := false
			if f_sub.Kind != Entity_Constant {
				continue
			}
			for _, f_super := range super.Enum.Fields {
				if f_super.Kind != Entity_Constant {
					continue
				}
				if f_sub.Token.String == f_super.Token.String {
					if compare_exact_values(Token_CmpEq, f_sub.Constant.Value, f_super.Constant.Value) {
						found = true
						break
					}
				}
			}
			if !found {
				operand.Value = exact_value_bool(false)
				return true
			}
		}
		operand.Value = exact_value_bool(true)
		return true
	} else if super.Kind == Type_Union {
		if len(sub.Union.Variants) > len(super.Union.Variants) {
			operand.Value = exact_value_bool(false)
			return true
		}
		if sub.Union.Kind != super.Union.Kind {
			operand.Value = exact_value_bool(false)
			return true
		}
		for i := range sub.Union.Variants {
			t_sub := sub.Union.Variants[i]
			t_super := super.Union.Variants[i]
			if !are_types_identical(t_sub, t_super) {
				operand.Value = exact_value_bool(false)
				return true
			}
		}
		operand.Value = exact_value_bool(true)
		return true
	}
	a := type_to_string(op_super.Type)
	b := type_to_string(op_sub.Type)
	error(op_super.Expr, "'%s' expects types of the same kind and either an enum or union, got %s vs %s", builtin_name, a, b)
	gb_string_free(b)
	gb_string_free(a)
	return false
}

func checkBuiltinProc_type_field_index_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	var x Operand
	check_expr(c, &x, ce.Args[1])
	if !is_type_string(x.Type) || x.Mode != Addressing_Constant || x.Value.Kind != ExactValue_String {
		error(ce.Args[1], "Expected a constant string for field argument")
		return false
	}
	field_name := string_interner_insert(x.Value.ValueString)
	sel := lookup_field(type_, field_name, false)
	if sel.Entity == nil {
		begin_error_block()
		type_str := type_to_string(bt)
		error(ce.Args[0], "'%s' has no field named '%s'", type_str, field_name)
		gb_string_free(type_str)
		if bt.Kind == Type_Struct {
			check_did_you_mean_type(field_name, bt.Struct.Fields)
		}
		end_error_block()
		return false
	}
	if sel.Indirect {
		type_str := type_to_string(bt)
		error(ce.Args[0], "Field '%s' is embedded via a pointer in '%s'", field_name, type_str)
		gb_string_free(type_str)
		return false
	}
	operand.Mode = Addressing_Constant
	operand.Value = exact_value_u64(uint64(sel.Index[0]))
	operand.Type = t_uintptr
	return true
}

func checkBuiltinProc_type_fixed_capacity_dynamic_array_len_offset(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a fixed capacity dynamic array type for '%s'", builtin_name)
		return false
	}
	if !is_type_fixed_capacity_dynamic_array(type_) {
		error(ce.Args[0], "Expected a fixed capacity dynamic array type for '%s'", builtin_name)
		return false
	}
	offset := type_offset_of(type_, 1, nil)
	operand.Mode = Addressing_Constant
	operand.Value = exact_value_u64(uint64(offset))
	operand.Type = t_uintptr
	return true
}

func checkBuiltinProc_type_bit_set_backing_type(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	type_ := check_type(c, ce.Args[0])
	bt := base_type(type_)
	if bt == nil || bt == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if bt.Kind != Type_BitSet {
		s := type_to_string(type_)
		error(ce.Args[0], "Expected a bit_set type for '%s', got %s", builtin_name, s)
		gb_string_free(s)
		return false
	}
	operand.Mode = Addressing_Type
	operand.Type = bit_set_to_int(bt)
	return true
}

func checkBuiltinProc_type_enum_is_contiguous(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if !is_type_enum(type_) {
		t := type_to_string(type_)
		error(ce.Args[0], "Expected an enum type for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	enum_constants := make([]*Entity, len(type_.Enum.Fields))
	copy(enum_constants, type_.Enum.Fields)
	sort.Slice(enum_constants, func(i, j int) bool {
		return enum_constant_entity_cmp(enum_constants[i], enum_constants[j]) < 0
	})
	minus_one := big_int_make_i64(-1)
	defer big_int_dealloc(&minus_one)
	contiguous := true
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	for i := 0; i < len(enum_constants)-1; i++ {
		curr := enum_constants[i].Constant.Value.ValueInteger
		next := enum_constants[i+1].Constant.Value.ValueInteger
		var diff BigInt
		big_int_sub(&diff, &curr, &next)
		if !big_int_is_zero(&diff) && big_int_cmp(&diff, &minus_one) != 0 {
			contiguous = false
			big_int_dealloc(&diff)
			break
		}
		big_int_dealloc(&diff)
	}
	operand.Value = exact_value_bool(contiguous)
	return true
}

func checkBuiltinProc_type_equal_proc(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if !is_type_comparable(type_) {
		t := type_to_string(type_)
		error(ce.Args[0], "Expected a comparable type for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = t_equal_proc
	return true
}

func checkBuiltinProc_type_hasher_proc(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if !is_type_valid_for_keys(type_) {
		t := type_to_string(type_)
		error(ce.Args[0], "Expected a valid type for map keys for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	add_map_key_type_dependencies(c, type_)
	operand.Mode = Addressing_Value
	operand.Type = t_hasher_proc
	return true
}

func checkBuiltinProc_type_map_info(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	if !is_type_map(type_) {
		t := type_to_string(type_)
		error(ce.Args[0], "Expected a map type for '%s', got %s", builtin_name, t)
		gb_string_free(t)
		return false
	}
	add_map_key_type_dependencies(c, type_)
	operand.Mode = Addressing_Value
	operand.Type = t_map_info_ptr
	return true
}

func checkBuiltinProc_type_map_cell_info(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	bt := check_type(c, ce.Args[0])
	type_ := base_type(bt)
	if type_ == nil || type_ == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = t_map_cell_info_ptr
	return true
}

func checkBuiltinProc_type_canonical_name(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	type_ := check_type(c, ce.Args[0])
	bt := base_type(type_)
	if bt == nil || bt == t_invalid {
		error(ce.Args[0], "Expected a type for '%s'", builtin_name)
		return false
	}
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_string
	operand.Value = exact_value_string(type_to_canonical_string(permanent_allocator(), type_))
	return true
}

func checkBuiltinProc_procedure_of(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	call_expr := unparen_expr(ce.Args[0])
	var op Operand
	check_expr_base(c, &op, ce.Args[0], nil)
	if op.Mode != Addressing_Value && !(call_expr != nil && call_expr.Kind == Ast_CallExpr) {
		error(ce.Args[0], "Expected a call expression for '%s'", builtin_name)
		return false
	}
	proc := call_expr.CallExpr.Proc
	e := entity_of_node(proc)
	if e == nil {
		error(ce.Args[0], "Invalid procedure value, expected a regular/specialized procedure")
		return false
	}
	tav := proc.TAV
	operand.Type = e.Type
	operand.Mode = Addressing_Value
	operand.Value = tav.Value
	operand.BuiltinID = BuiltinProc_Invalid
	operand.ProcGroup = nil
	if tav.Mode == Addressing_Builtin {
		operand.Mode = tav.Mode
		operand.BuiltinID = BuiltinProcId(e.Builtin.ID)
		return true
	}
	if !is_type_proc(e.Type) {
		s := type_to_string(e.Type)
		error(ce.Args[0], "Expected a procedure value, got '%s'", s)
		gb_string_free(s)
		return false
	}
	ce.EntityProcedureOf.Store(e)
	return true
}
