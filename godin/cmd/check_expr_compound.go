package cmd

func get_constant_field_single(c *CheckerContext, value ExactValue, index int32, success_ *bool, finish_ *bool) ExactValue {
	if value.Kind == ExactValue_String {
		gb_assert_handler("Assertion Failure", "0 <= index && index < value.value_string.len", "cmd_check_expr_compound.go", 0)
		val := value.ValueString[index]
		if success_ != nil {
			*success_ = true
		}
		if finish_ != nil {
			*finish_ = true
		}
		return exact_value_u64(uint64(val))
	} else if value.Kind == ExactValue_String16 {
		gb_assert_handler("Assertion Failure", "0 <= index && index < value.value_string16.len", "cmd_check_expr_compound.go", 0)
		val := value.ValueString16[index]
		if success_ != nil {
			*success_ = true
		}
		if finish_ != nil {
			*finish_ = true
		}
		return exact_value_u64(uint64(val))
	}
	if value.Kind != ExactValue_Compound {
		if success_ != nil {
			*success_ = true
		}
		if finish_ != nil {
			*finish_ = true
		}
		return value
	}
	node := value.ValueCompound
	switch node.Kind {
	case Ast_CompoundLit:
		cl := &node.CompoundLit
		if len(cl.Elems) == 0 {
			if success_ != nil {
				*success_ = true
			}
			if finish_ != nil {
				*finish_ = true
			}
			return empty_exact_value
		}
		if cl.Elems[0].Kind == Ast_FieldValue {
			if is_type_raw_union(node.TAV.Type) {
				if success_ != nil {
					*success_ = false
				}
				if finish_ != nil {
					*finish_ = true
				}
				return empty_exact_value
			} else if is_type_struct(node.TAV.Type) {
				found := false
				for _, elem := range cl.Elems {
					if elem.Kind != Ast_FieldValue {
						continue
					}
					fv := &elem.FieldValue
					name := fv.Field.Ident.Interned
					sub_sel := lookup_field(node.TAV.Type, name, false)
					if len(sub_sel.Index) > 0 && sub_sel.Index[0] == index {
						value = fv.Value.TAV.Value
						found = true
						break
					}
				}
				if !found {
					value = empty_exact_value
				}
			} else if is_type_array(node.TAV.Type) || is_type_enumerated_array(node.TAV.Type) {
				for _, elem := range cl.Elems {
					if elem.Kind != Ast_FieldValue {
						continue
					}
					fv := &elem.FieldValue
					if is_ast_range(fv.Field) {
						ie := &fv.Field.BinaryExpr
						lo_tav := ie.Left.TAV
						hi_tav := ie.Right.TAV
						gb_assert_handler("Assertion Failure", "lo_tav.Mode == Addressing_Constant", "cmd_check_expr_compound.go", 0)
						gb_assert_handler("Assertion Failure", "hi_tav.Mode == Addressing_Constant", "cmd_check_expr_compound.go", 0)
						op := ie.Op.Kind
						lo := exact_value_to_i64(lo_tav.Value)
						hi := exact_value_to_i64(hi_tav.Value)
						corrected_index := int64(index)
						if is_type_enumerated_array(node.TAV.Type) {
							bt := base_type(node.TAV.Type)
							gb_assert_handler("Assertion Failure", "bt.Kind == Type_EnumeratedArray", "cmd_check_expr_compound.go", 0)
							corrected_index = int64(index) + exact_value_to_i64(*bt.EnumeratedArray.MinValue)
						}
						if op != Token_RangeHalf {
							if lo <= corrected_index && corrected_index <= hi {
								tav := fv.Value.TAV
								if success_ != nil {
									*success_ = true
								}
								if finish_ != nil {
									*finish_ = false
								}
								return tav.Value
							}
						} else {
							if lo <= corrected_index && corrected_index < hi {
								tav := fv.Value.TAV
								if success_ != nil {
									*success_ = true
								}
								if finish_ != nil {
									*finish_ = false
								}
								return tav.Value
							}
						}
					} else {
						index_tav := fv.Field.TAV
						if index_tav.Mode != Addressing_Constant {
							if success_ != nil {
								*success_ = false
							}
							if finish_ != nil {
								*finish_ = true
							}
							return empty_exact_value
						}
						gb_assert_handler("Assertion Failure", "index_tav.Mode == Addressing_Constant", "cmd_check_expr_compound.go", 0)
						index_value := index_tav.Value
						if is_type_enumerated_array(node.TAV.Type) {
							bt := base_type(node.TAV.Type)
							gb_assert_handler("Assertion Failure", "bt.Kind == Type_EnumeratedArray", "cmd_check_expr_compound.go", 0)
							index_value = exact_value_sub(index_value, *bt.EnumeratedArray.MinValue)
						}
						field_index := exact_value_to_i64(index_value)
						if int64(index) == field_index {
							tav := fv.Value.TAV
							if success_ != nil {
								*success_ = true
							}
							if finish_ != nil {
								*finish_ = false
							}
							return tav.Value
						}
					}
				}
			}
		} else {
			count := int32(len(cl.Elems))
			if count < index {
				if success_ != nil {
					*success_ = false
				}
				if finish_ != nil {
					*finish_ = true
				}
				return empty_exact_value
			}
			if int32(len(cl.Elems)) <= index {
				if success_ != nil {
					*success_ = false
				}
				if finish_ != nil {
					*finish_ = false
				}
				return value
			}
			tav := cl.Elems[index].TAV
			if tav.Mode == Addressing_Constant {
				if success_ != nil {
					*success_ = true
				}
				if finish_ != nil {
					*finish_ = false
				}
				return tav.Value
			} else if is_type_proc(tav.Type) {
				if success_ != nil {
					*success_ = true
				}
				if finish_ != nil {
					*finish_ = false
				}
				return tav.Value
			} else {
				gb_assert_handler("Assertion Failure", "is_type_untyped_nil(tav.Type)", "cmd_check_expr_compound.go", 0)
				if success_ != nil {
					*success_ = true
				}
				if finish_ != nil {
					*finish_ = false
				}
				return tav.Value
			}
		}
	default:
		if success_ != nil {
			*success_ = true
		}
		if finish_ != nil {
			*finish_ = true
		}
		return empty_exact_value
	}
	if finish_ != nil {
		*finish_ = false
	}
	return value
}

func get_constant_field(c *CheckerContext, operand *Operand, sel Selection, success_ *bool) ExactValue {
	if operand.Mode != Addressing_Constant {
		if success_ != nil {
			*success_ = false
		}
		return empty_exact_value
	}
	if sel.Indirect {
		if success_ != nil {
			*success_ = false
		}
		return empty_exact_value
	}
	if len(sel.Index) == 0 {
		if success_ != nil {
			*success_ = false
		}
		return empty_exact_value
	}
	value := operand.Value
	if value.Kind == ExactValue_Compound {
		for len(sel.Index) > 0 {
			index := sel.Index[0]
			sel = sub_selection(sel, 1)
			finish := false
			value = get_constant_field_single(c, value, index, success_, &finish)
			if finish {
				return value
			}
		}
		if success_ != nil {
			*success_ = true
		}
		return value
	} else if value.Kind == ExactValue_Quaternion {
		q := *value.ValueQuaternion
		gb_assert_handler("Assertion Failure", "len(sel.Index) == 1", "cmd_check_expr_compound.go", 0)
		switch sel.Index[0] {
		case 3:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(q.Real)
		case 0:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(q.Imag)
		case 1:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(q.Jmag)
		case 2:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(q.Kmag)
		}
		if success_ != nil {
			*success_ = false
		}
		return empty_exact_value
	} else if value.Kind == ExactValue_Complex {
		cmplx := *value.ValueComplex
		gb_assert_handler("Assertion Failure", "len(sel.Index) == 1", "cmd_check_expr_compound.go", 0)
		switch sel.Index[0] {
		case 0:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(cmplx.Real)
		case 1:
			if success_ != nil {
				*success_ = true
			}
			return exact_value_float(cmplx.Imag)
		}
		if success_ != nil {
			*success_ = false
		}
		return empty_exact_value
	}
	if success_ != nil {
		*success_ = true
	}
	return empty_exact_value
}

func check_compound_literal_field_values(c *CheckerContext, elems []*Ast, o *Operand, typ *Type, is_constant *bool) {
	bt := base_type(typ)
	fields_visited := make(StringSet)
	defer string_set_destroy(&fields_visited)
	fields_visited_through_raw_union := make(StringMap)
	defer string_map_destroy(&fields_visited_through_raw_union)
	assignment_str := String{Text: strPtr("structure literal"), Len: isize(len("structure literal"))}
	if bt.Kind == Type_BitField {
		assignment_str = String{Text: strPtr("bit_field literal"), Len: isize(len("bit_field literal"))}
	}
	for _, elem := range elems {
		if elem.Kind != Ast_FieldValue {
			error(elem, "Mixture of 'field = value' and value elements in a literal is not allowed")
			continue
		}
		fv := &elem.FieldValue
		ident := fv.Field
		if ident.Kind == Ast_ImplicitSelectorExpr {
			expr_str := expr_to_string(ident)
			error(ident, "Field names do not start with a '.', remove the '.' in structure literal")
			gb_string_free(expr_str)
			ident = ident.ImplicitSelectorExpr.Selector
		}
		if ident.Kind != Ast_Ident {
			expr_str := expr_to_string(ident)
			error(elem, "Invalid field name '%s' in structure literal", expr_str)
			gb_string_free(expr_str)
			continue
		}
		name := ident.Ident.Token.String
		interned := ident.Ident.Interned
		sel := lookup_field(typ, interned, o.Mode == Addressing_Type)
		is_unknown := sel.Entity == nil
		if is_unknown {
			error(ident, "Unknown field '%.*s' in structure literal", isize(name.Len), name.Text)
			continue
		}
		var field *Entity
		if bt.Kind == Type_Struct {
			field = bt.Struct.Fields[sel.Index[0]]
		} else if bt.Kind == Type_BitField {
			field = bt.BitField.Fields[sel.Index[0]]
		} else {
			gb_assert_handler("Panic", 0, "cmd_check_expr_compound.go", 0)
		}
		add_entity_use(c, ident, field)
		if string_set_update(&fields_visited, name) {
			if len(sel.Index) > 1 {
				if found := string_map_get(&fields_visited_through_raw_union, sel.Entity.Token.String); found != nil {
					error(ident, "Field '%.*s' is already initialized due to a previously assigned struct #raw_union field '%.*s'", isize(sel.Entity.Token.String.Len), sel.Entity.Token.String.Text, isize(found.Len), found.Text)
				} else {
					error(ident, "Duplicate or reused field '%.*s' in %.*s", isize(sel.Entity.Token.String.Len), sel.Entity.Token.String.Text, isize(assignment_str.Len), assignment_str.Text)
				}
			} else {
				error(ident, "Duplicate field '%.*s' in %.*s", isize(field.Token.String.Len), field.Token.String.Text, isize(assignment_str.Len), assignment_str.Text)
			}
			continue
		} else if found := string_map_get(&fields_visited_through_raw_union, sel.Entity.Token.String); found != nil {
			error(ident, "Field '%.*s' is already initialized due to a previously assigned struct #raw_union field '%.*s'", isize(sel.Entity.Token.String.Len), sel.Entity.Token.String.Text, isize(found.Len), found.Text)
			continue
		}
		if sel.Indirect {
			error(ident, "Cannot assign to the %d-nested anonymous indirect field '%.*s' in a %.*s", isize(len(sel.Index))-1, isize(name.Len), name.Text, isize(assignment_str.Len), assignment_str.Text)
			continue
		}
		if len(sel.Index) > 1 {
			gb_assert_handler("Assertion Failure", "bt.Kind == Type_Struct", "cmd_check_expr_compound.go", 0)
			if *is_constant {
				ft := typ
				for _, idx := range sel.Index {
					bt2 := base_type(ft)
					switch bt2.Kind {
					case Type_Struct:
						if bt2.Struct.IsRawUnion {
							*is_constant = false
							break
						}
						ft = bt2.Struct.Fields[idx].Type
					case Type_Array:
						ft = bt2.Array.Elem
					case Type_BitField:
						*is_constant = false
						ft = bt2.BitField.Fields[idx].Type
					default:
						gb_assert_handler("Panic", 0, "cmd_check_expr_compound.go", 0)
					}
				}
				if *is_constant && elem_cannot_be_constant(ft) {
					*is_constant = false
				}
			}
			nested_ft := bt
			for _, idx := range sel.Index {
				bt2 := base_type(nested_ft)
				switch bt2.Kind {
				case Type_Struct:
					if bt2.Struct.IsRawUnion {
						for _, re := range bt2.Struct.Fields {
							string_map_set(&fields_visited_through_raw_union, re.Token.String, sel.Entity.Token.String)
						}
					}
					nested_ft = bt2.Struct.Fields[idx].Type
				case Type_Array:
					nested_ft = bt2.Array.Elem
				case Type_BitField:
					nested_ft = bt2.BitField.Fields[idx].Type
				default:
					gb_assert_handler("Panic", 0, "cmd_check_expr_compound.go", 0)
				}
			}
			field = sel.Entity
		}
		operand := &Operand{}
		check_expr_or_type(c, operand, fv.Value, field.Type)
		if elem_cannot_be_constant(field.Type) {
			*is_constant = false
		}
		if *is_constant {
			*is_constant = check_is_operand_compound_lit_constant(c, operand, field.Type)
		}
		prev_bit_field_bit_size := c.bit_field_bit_size
		if field.Kind == Entity_Variable && field.Variable.bit_field_bit_size != 0 {
			c.bit_field_bit_size = field.Variable.bit_field_bit_size
		}
		check_assignment(c, operand, field.Type, assignment_str)
		c.bit_field_bit_size = prev_bit_field_bit_size
	}
	if bt.Kind == Type_Struct && bt.Struct.IsAllOrNone && len(elems) > 0 && len(bt.Struct.Fields) > 0 {
		missing_fields := &PtrSet{}
		defer ptr_set_destroy(missing_fields)
		for i := 0; i < len(bt.Struct.Fields); i++ {
			field := bt.Struct.Fields[i]
			name := field.Token.String
			if is_blank_ident(name) || name.Len == 0 {
				continue
			}
			found := string_set_exists(&fields_visited, name)
			raw_union := string_map_get(&fields_visited_through_raw_union, name)
			if !found && raw_union == nil {
				ptr_set_add(missing_fields, field)
			}
		}
		if missing_fields.count > 0 {
			expr := o.Expr
			if expr == nil {
				gb_assert_handler("Assertion Failure", "len(elems) > 0", "cmd_check_expr_compound.go", 0)
				expr = elems[len(elems)-1]
			}
			begin_error_block()
			if build_context.terse_errors {
				fields_string := gb_string_make(heap_allocator(), "")
				defer gb_string_free(fields_string)
				i := 0
				for _, field := range missing_fields.keys {
					if field == nil {
						continue
					}
					if i > 0 {
						fields_string = gb_string_appendc(fields_string, ", ")
					}
					name := field.Token.String
					fields_string = gb_string_append_length(fields_string, name.Text, name.Len)
					i++
				}
				error(expr, "All or none of the fields must be assigned to a struct with '#all_or_none' applied, missing fields: %s", fields_string)
			} else {
				error(expr, "All or none of the fields must be assigned to a struct with '#all_or_none' applied, missing fields:")
				for _, field := range missing_fields.keys {
					if field == nil {
						continue
					}
					s := type_to_string(field.Type)
					error_line("\t%.*s: %s\n", isize(field.Token.String.Len), field.Token.String.Text, s)
					gb_string_free(s)
				}
			}
			end_error_block()
		}
	}
}

func check_compound_literal(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := Expr_Expr
	cl := &node.CompoundLit
	typ := type_hint
	if typ != nil && is_type_untyped(typ) {
		typ = nil
	}
	is_to_be_determined_array_count := false
	is_constant := true
	is_soa := false
	type_expr := cl.Type
	used_type_hint_expr := false
	if type_expr == nil && c.type_hint_expr != nil {
		if is_expr_inferred_fixed_array(c.type_hint_expr) {
			type_expr = clone_ast(c.type_hint_expr, nil)
			used_type_hint_expr = true
		}
	}
	if type_expr != nil {
		typ = nil
		if type_expr.Kind == Ast_ArrayType {
			count := type_expr.ArrayType.Count
			if count != nil {
				if count.Kind == Ast_UnaryExpr && count.UnaryExpr.Op.Kind == Token_Question {
					typ = alloc_type_array(check_type(c, type_expr.ArrayType.Elem), -1)
					is_to_be_determined_array_count = true
				}
			} else {
				typ = alloc_type_slice(check_type(c, type_expr.ArrayType.Elem))
			}
			if len(cl.Elems) > 0 {
				if type_expr.ArrayType.Tag != nil {
					tag := type_expr.ArrayType.Tag
					gb_assert_handler("Assertion Failure", "tag.Kind == Ast_BasicDirective", "cmd_check_expr_compound.go", 0)
					name := tag.BasicDirective.Name.String
					if name == "soa" {
						is_soa = true
						if count == nil {
							error(node, "#soa slices are not supported for compound literals")
							return kind
						} else if count.Kind == Ast_UnaryExpr && count.UnaryExpr.Op.Kind == Token_Question {
							error(node, "#soa fixed length arrays must specify their length and cannot use ?")
						}
					}
				}
			}
		} else if type_expr.Kind == Ast_DynamicArrayType && type_expr.DynamicArrayType.Tag != nil {
			if len(cl.Elems) > 0 {
				tag := type_expr.DynamicArrayType.Tag
				gb_assert_handler("Assertion Failure", "tag.Kind == Ast_BasicDirective", "cmd_check_expr_compound.go", 0)
				name := tag.BasicDirective.Name.String
				if name == "soa" {
					is_soa = true
					error(node, "#soa dynamic arrays are not supported for compound literals")
					return kind
				}
			}
		}
		if typ == nil {
			typ = check_type(c, type_expr)
		}
	}
	if typ == nil {
		error(node, "Missing type in compound literal")
		return kind
	}
	t := base_type(typ)
	if is_type_polymorphic(t) {
		str := type_to_string(typ)
		error(node, "Cannot use a polymorphic type for a compound literal, got '%s'", str)
		o.Expr = node
		o.Type = typ
		gb_string_free(str)
		return kind
	}
	switch t.Kind {
	case Type_Struct:
		if len(cl.Elems) == 0 {
			break
		}
		if t.Struct.SoaKind == StructSoaNone {
			if t.Struct.IsRawUnion {
				if len(cl.Elems) > 0 {
					is_constant = elem_type_can_be_constant(t)
					if cl.Elems[0].Kind != Ast_FieldValue {
						type_str := type_to_string(typ)
						error(node, "%s ('struct #raw_union') compound literals are only allowed to contain 'field = value' elements", type_str)
						gb_string_free(type_str)
					} else {
						if len(cl.Elems) != 1 {
							type_str := type_to_string(typ)
							error(node, "%s ('struct #raw_union') compound literals are only allowed to contain up to 1 'field = value' element, got %d", type_str, isize(len(cl.Elems)))
							gb_string_free(type_str)
						} else {
							check_compound_literal_field_values(c, cl.Elems, o, typ, &is_constant)
						}
					}
				}
				break
			}
			wait_signal_until_available(&t.Struct.FieldsWaitSignal)
			field_count := isize(len(t.Struct.Fields))
			min_field_count := field_count
			for i := min_field_count - 1; i >= 0; i-- {
				e := t.Struct.Fields[i]
				gb_assert_handler("Assertion Failure", "e.Kind == Entity_Variable", "cmd_check_expr_compound.go", 0)
				if e.Variable.param_value.Kind != ParameterValue_Invalid {
					min_field_count--
				} else {
					break
				}
			}
			if len(cl.Elems) > 0 && cl.Elems[0].Kind == Ast_FieldValue {
				check_compound_literal_field_values(c, cl.Elems, o, typ, &is_constant)
			} else {
				seen_field_value := false
				handled_elem_count := isize(0)
				index := isize(0)
				for _, elem := range cl.Elems {
					var field *Entity
					if elem.Kind == Ast_FieldValue {
						seen_field_value = true
						error(elem, "Mixture of 'field = value' and value elements in a literal is not allowed")
						index++
						continue
					} else if seen_field_value {
						error(elem, "Value elements cannot be used after a 'field = value'")
						index++
						continue
					}
					if index >= field_count {
						error(elem, "Too many values in structure literal, expected %d, got %d", isize(field_count), isize(len(cl.Elems)))
						index++
						break
					}
					if field == nil {
						field = t.Struct.Fields[index]
					}
					operand_o := &Operand{}
					check_multi_expr_with_type_hint(c, operand_o, elem, field.Type)
					if is_type_tuple(operand_o.Type) {
						is_constant = false
						tt := &operand_o.Type.Tuple
						jj := isize(0)
						for _, src_field := range tt.Variables {
							src_o := *operand_o
							src_o.Type = src_field.Type
							field = t.Struct.Fields[index+jj]
							check_assignment(c, &src_o, field.Type, String{Text: strPtr("structure literal"), Len: isize(len("structure literal"))})
							jj++
						}
						index += isize(len(tt.Variables)) - 1
						handled_elem_count += isize(len(tt.Variables))
					} else {
						check_not_tuple(c, operand_o)
						if elem_cannot_be_constant(field.Type) {
							is_constant = false
						}
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand_o, field.Type)
						}
						check_assignment(c, operand_o, field.Type, String{Text: strPtr("structure literal"), Len: isize(len("structure literal"))})
						handled_elem_count++
					}
					index++
				}
				if isize(len(cl.Elems)) < field_count {
					if min_field_count < field_count {
						if isize(len(cl.Elems)) < min_field_count {
							error(cl.Close, "Too few values in structure literal, expected at least %d, got %d", isize(min_field_count), isize(len(cl.Elems)))
						}
					} else if handled_elem_count != field_count {
						error(cl.Close, "Too few values in structure literal, expected %d, got %d", isize(field_count), isize(len(cl.Elems)))
					}
				}
			}
			break
		} else if t.Struct.SoaKind != StructSoaFixed {
			error(node, "#soa slices and dynamic arrays are not supported for compound literals")
			break
		}
		fallthrough
	case Type_Slice:
		fallthrough
	case Type_Array:
		fallthrough
	case Type_DynamicArray:
		fallthrough
	case Type_SimdVector:
		fallthrough
	case Type_Matrix:
		fallthrough
	case Type_FixedCapacityDynamicArray:
		{
			var elem_type *Type
			context_name := String{}
			max_type_count := int64(-1)
			if t.Kind == Type_Struct {
				gb_assert_handler("Assertion Failure", "t.Struct.SoaKind == StructSoaFixed", "cmd_check_expr_compound.go", 0)
				elem_type = t.Struct.SoaElem
				context_name = String{Text: strPtr("#soa array literal"), Len: isize(len("#soa array literal"))}
				if !is_to_be_determined_array_count {
					max_type_count = int64(t.Struct.SoaCount)
				}
			} else if t.Kind == Type_Slice {
				elem_type = t.Slice.Elem
				context_name = String{Text: strPtr("slice literal"), Len: isize(len("slice literal"))}
			} else if t.Kind == Type_Array {
				elem_type = t.Array.Elem
				context_name = String{Text: strPtr("array literal"), Len: isize(len("array literal"))}
				if !is_to_be_determined_array_count {
					max_type_count = t.Array.Count
				}
			} else if t.Kind == Type_DynamicArray {
				elem_type = t.DynamicArray.Elem
				context_name = String{Text: strPtr("dynamic array literal"), Len: isize(len("dynamic array literal"))}
				is_constant = false
			} else if t.Kind == Type_FixedCapacityDynamicArray {
				elem_type = t.FixedCapacityDynamicArray.Elem
				context_name = String{Text: strPtr("fixed capacity dynamic array literal"), Len: isize(len("fixed capacity dynamic array literal"))}
				max_type_count = t.FixedCapacityDynamicArray.Capacity
			} else if t.Kind == Type_SimdVector {
				elem_type = t.SimdVector.Elem
				context_name = String{Text: strPtr("simd vector literal"), Len: isize(len("simd vector literal"))}
				max_type_count = t.SimdVector.Count
			} else if t.Kind == Type_Matrix {
				elem_type = t.Matrix.Elem
				context_name = String{Text: strPtr("matrix literal"), Len: isize(len("matrix literal"))}
				max_type_count = t.Matrix.RowCount * t.Matrix.ColumnCount
			} else {
				gb_assert_handler("Panic", 0, "cmd_check_expr_compound.go", 0)
			}
			max_val := int64(0)
			bet := base_type(elem_type)
			if !elem_type_can_be_constant(bet) {
				is_constant = false
			}
			if bet == t_invalid {
				break
			}
			if len(cl.Elems) > 0 && cl.Elems[0].Kind == Ast_FieldValue {
				rc := range_cache_make(heap_allocator())
				defer range_cache_destroy(&rc)
				for _, elem := range cl.Elems {
					if elem.Kind != Ast_FieldValue {
						error(elem, "Mixture of 'field = value' and value elements in a literal is not allowed")
						continue
					}
					fv := &elem.FieldValue
					if is_ast_range(fv.Field) {
						op := fv.Field.BinaryExpr.Op
						x := &Operand{}
						y := &Operand{}
						ok := check_range(c, fv.Field, false, x, y, nil)
						if !ok {
							continue
						}
						if x.Mode != Addressing_Constant || !is_type_integer(core_type(x.Type)) {
							error(x.Expr, "Expected a constant integer as an array field")
							continue
						}
						if y.Mode != Addressing_Constant || !is_type_integer(core_type(y.Type)) {
							error(y.Expr, "Expected a constant integer as an array field")
							continue
						}
						lo := exact_value_to_i64(x.Value)
						hi := exact_value_to_i64(y.Value)
						max_index := hi
						if op.Kind == Token_RangeHalf {
							hi--
						} else {
							max_index++
						}
						new_range := range_cache_add_range(&rc, lo, hi)
						if !new_range {
							error(elem, "Overlapping field range index %lld %.*s %lld for %.*s", lo, isize(op.String.Len), op.String.Text, hi, isize(context_name.Len), context_name.Text)
							continue
						}
						if max_type_count >= 0 && (lo < 0 || lo >= max_type_count) {
							error(elem, "Index %lld is out of bounds (0..<%lld) for %.*s", lo, max_type_count, isize(context_name.Len), context_name.Text)
							continue
						}
						if max_type_count >= 0 && (hi < 0 || hi >= max_type_count) {
							error(elem, "Index %lld is out of bounds (0..<%lld) for %.*s", hi, max_type_count, isize(context_name.Len), context_name.Text)
							continue
						}
						if max_val < hi {
							max_val = max_index
						}
						operand := &Operand{}
						check_expr_with_type_hint(c, operand, fv.Value, elem_type)
						check_assignment(c, operand, elem_type, context_name)
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
						}
					} else {
						op_index := &Operand{}
						check_expr(c, op_index, fv.Field)
						if op_index.Mode != Addressing_Constant || !is_type_integer(core_type(op_index.Type)) {
							error(elem, "Expected a constant integer as an array field")
							continue
						}
						idx := exact_value_to_i64(op_index.Value)
						if max_type_count >= 0 && (idx < 0 || idx >= max_type_count) {
							error(elem, "Index %lld is out of bounds (0..<%lld) for %.*s", idx, max_type_count, isize(context_name.Len), context_name.Text)
							continue
						}
						new_index := range_cache_add_index(&rc, idx)
						if !new_index {
							error(elem, "Duplicate field index %lld for %.*s", idx, isize(context_name.Len), context_name.Text)
							continue
						}
						if max_val < idx+1 {
							max_val = idx + 1
						}
						operand := &Operand{}
						check_expr_with_type_hint(c, operand, fv.Value, elem_type)
						check_assignment(c, operand, elem_type, context_name)
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
						}
					}
				}
				cl.MaxCount = max_val
			} else {
				for idx := isize(0); idx < isize(len(cl.Elems)); idx++ {
					e := cl.Elems[idx]
					if e == nil {
						error(node, "Invalid literal element")
						max_val++
						continue
					}
					if e.Kind == Ast_FieldValue {
						error(e, "Mixture of 'field = value' and value elements in a literal is not allowed")
						max_val++
						continue
					}
					if max_type_count >= 0 && max_type_count <= int64(idx) {
						error(e, "Index %lld is out of bounds (>= %lld) for %.*s", int64(idx), max_type_count, isize(context_name.Len), context_name.Text)
					}
					operand := &Operand{}
					check_multi_expr_with_type_hint(c, operand, e, elem_type)
					if is_type_tuple(operand.Type) {
						is_constant = false
						tt := &operand.Type.Tuple
						for jj := 0; jj < len(tt.Variables); jj++ {
							oo := *operand
							oo.Type = tt.Variables[jj].Type
							check_assignment(c, &oo, elem_type, context_name)
						}
						max_val += int64(len(tt.Variables)) - 1
					} else {
						check_assignment(c, operand, elem_type, context_name)
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
						}
					}
					max_val++
				}
			}
			if t.Kind == Type_Array {
				if is_to_be_determined_array_count {
					t.Array.Count = max_val
				} else if len(cl.Elems) > 0 && cl.Elems[0].Kind != Ast_FieldValue {
					if 0 < max_val && max_val < t.Array.Count {
						error(node, "Expected %lld values for this array literal, got %lld", t.Array.Count, max_val)
					}
				}
			} else if t.Kind == Type_Struct {
				gb_assert_handler("Assertion Failure", "t.Struct.SoaKind == StructSoaFixed", "cmd_check_expr_compound.go", 0)
				if is_to_be_determined_array_count {
					t.Struct.SoaCount = int32(max_val)
				} else if len(cl.Elems) > 0 && cl.Elems[0].Kind != Ast_FieldValue {
					if 0 < max_val && max_val < int64(t.Struct.SoaCount) {
						error(node, "Expected %lld values for this #soa array literal, got %lld", int64(t.Struct.SoaCount), max_val)
					}
				}
			} else if t.Kind == Type_FixedCapacityDynamicArray {
				if max_val > t.FixedCapacityDynamicArray.Capacity {
					error(node, "Expected a maximum of %lld values for this fixed capacity dynamic array, got %lld", t.FixedCapacityDynamicArray.Capacity, max_val)
				}
			}
			if t.Kind == Type_DynamicArray {
				if check_for_dynamic_literals(c, node, cl) {
					add_package_dependency(c, "runtime", "__dynamic_array_reserve", false)
					add_package_dependency(c, "runtime", "__dynamic_array_append", false)
				}
			}
			if t.Kind == Type_Matrix {
				if len(cl.Elems) > 0 && cl.Elems[0].Kind != Ast_FieldValue {
					if 0 < max_val && max_val < max_type_count {
						error(node, "Expected %lld values for this matrix literal, got %lld", max_type_count, max_val)
					}
				}
			}
			cl.MaxCount = max_val
		}
	case Type_EnumeratedArray:
		{
			elem_type := t.EnumeratedArray.Elem
			index_type := t.EnumeratedArray.Index
			context_name := String{Text: strPtr("enumerated array literal"), Len: isize(len("enumerated array literal"))}
			max_type_count := t.EnumeratedArray.Count
			index_type_str := type_to_string(index_type)
			defer gb_string_free(index_type_str)
			total_lo := exact_value_to_i64(*t.EnumeratedArray.MinValue)
			total_hi := exact_value_to_i64(*t.EnumeratedArray.MaxValue)
			total_lo_string := String{}
			total_hi_string := String{}
			gb_assert_handler("Assertion Failure", "is_type_enum(index_type)", "cmd_check_expr_compound.go", 0)
			{
				bt := base_type(index_type)
				gb_assert_handler("Assertion Failure", "bt.Kind == Type_Enum", "cmd_check_expr_compound.go", 0)
				for _, f := range bt.Enum.Fields {
					if f.Kind != Entity_Constant {
						continue
					}
					if total_lo_string.Len == 0 && compare_exact_values(Token_CmpEq, f.Constant.Value, *t.EnumeratedArray.MinValue) {
						total_lo_string = f.Token.String
					}
					if total_hi_string.Len == 0 && compare_exact_values(Token_CmpEq, f.Constant.Value, *t.EnumeratedArray.MaxValue) {
						total_hi_string = f.Token.String
					}
					if total_lo_string.Len != 0 && total_hi_string.Len != 0 {
						break
					}
				}
			}
			max_val := int64(0)
			bet := base_type(elem_type)
			if !elem_type_can_be_constant(bet) {
				is_constant = false
			}
			if bet == t_invalid {
				break
			}
			is_partial := false
			if cl.Tag != nil && cl.Tag.BasicDirective.Name.String == "partial" {
				is_partial = true
			}
			seen := make(SeenMap)
			defer func() {
				for k := range seen {
					delete(seen, k)
				}
			}()
			if len(cl.Elems) > 0 && cl.Elems[0].Kind == Ast_FieldValue {
				rc := range_cache_make(heap_allocator())
				defer range_cache_destroy(&rc)
				for _, elem := range cl.Elems {
					if elem.Kind != Ast_FieldValue {
						error(elem, "Mixture of 'field = value' and value elements in a literal is not allowed")
						continue
					}
					fv := &elem.FieldValue
					if is_ast_range(fv.Field) {
						op := fv.Field.BinaryExpr.Op
						x := &Operand{}
						y := &Operand{}
						ok := check_range(c, fv.Field, false, x, y, nil, index_type)
						if !ok {
							continue
						}
						if x.Mode != Addressing_Constant || !are_types_identical(x.Type, index_type) {
							error(x.Expr, "Expected a constant enum of type '%s' as an array field", index_type_str)
							continue
						}
						if y.Mode != Addressing_Constant || !are_types_identical(x.Type, index_type) {
							error(y.Expr, "Expected a constant enum of type '%s' as an array field", index_type_str)
							continue
						}
						lo := exact_value_to_i64(x.Value)
						hi := exact_value_to_i64(y.Value)
						max_index := hi
						if op.Kind == Token_RangeHalf {
							hi--
						}
						new_range := range_cache_add_range(&rc, lo, hi)
						if !new_range {
							lo_str := expr_to_string(x.Expr)
							hi_str := expr_to_string(y.Expr)
							error(elem, "Overlapping field range index %s %.*s %s for %.*s", lo_str, isize(op.String.Len), op.String.Text, hi_str, isize(context_name.Len), context_name.Text)
							gb_string_free(hi_str)
							gb_string_free(lo_str)
							continue
						}
						if max_type_count >= 0 && (lo < total_lo || lo > total_hi) {
							lo_str := expr_to_string(x.Expr)
							error(elem, "Index %s is out of bounds (%.*s .. %.*s) for %.*s", lo_str, isize(total_lo_string.Len), total_lo_string.Text, isize(total_hi_string.Len), total_hi_string.Text, isize(context_name.Len), context_name.Text)
							gb_string_free(lo_str)
							continue
						}
						if max_type_count >= 0 && (hi < 0 || hi > total_hi) {
							hi_str := expr_to_string(y.Expr)
							error(elem, "Index %s is out of bounds (%.*s .. %.*s) for %.*s", hi_str, isize(total_lo_string.Len), total_lo_string.Text, isize(total_hi_string.Len), total_hi_string.Text, isize(context_name.Len), context_name.Text)
							gb_string_free(hi_str)
							continue
						}
						if max_val < hi {
							max_val = max_index
						}
						operand := &Operand{}
						check_expr_with_type_hint(c, operand, fv.Value, elem_type)
						check_assignment(c, operand, elem_type, context_name)
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
						}
						upper_op := Token_LtEq
						if op.Kind == Token_RangeHalf {
							upper_op = Token_Lt
						}
						add_to_seen_map(c, &seen, upper_op, *x, *x, *y)
					} else {
						op_index := &Operand{}
						check_expr_with_type_hint(c, op_index, fv.Field, index_type)
						if op_index.Mode != Addressing_Constant || !are_types_identical(op_index.Type, index_type) {
							error(op_index.Expr, "Expected a constant enum of type '%s' as an array field", index_type_str)
							continue
						}
						idx := exact_value_to_i64(op_index.Value)
						if max_type_count >= 0 && (idx < total_lo || idx > total_hi) {
							idx_str := expr_to_string(op_index.Expr)
							error(elem, "Index %s is out of bounds (%.*s .. %.*s) for %.*s", idx_str, isize(total_lo_string.Len), total_lo_string.Text, isize(total_hi_string.Len), total_hi_string.Text, isize(context_name.Len), context_name.Text)
							gb_string_free(idx_str)
							continue
						}
						new_index := range_cache_add_index(&rc, idx)
						if !new_index {
							idx_str := expr_to_string(op_index.Expr)
							error(elem, "Duplicate field index %s for %.*s", idx_str, isize(context_name.Len), context_name.Text)
							gb_string_free(idx_str)
							continue
						}
						if max_val < idx+1 {
							max_val = idx + 1
						}
						operand := &Operand{}
						check_expr_with_type_hint(c, operand, fv.Value, elem_type)
						check_assignment(c, operand, elem_type, context_name)
						if is_constant {
							is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
						}
						add_to_seen_map_single(c, &seen, *op_index)
					}
				}
				cl.MaxCount = max_val
			} else {
				idx := isize(0)
				for ; idx < isize(len(cl.Elems)); idx++ {
					e := cl.Elems[idx]
					if e == nil {
						error(node, "Invalid literal element")
						continue
					}
					if e.Kind == Ast_FieldValue {
						error(e, "Mixture of 'field = value' and value elements in a literal is not allowed")
						continue
					}
					if max_type_count >= 0 && max_type_count <= int64(idx) {
						error(e, "Index %lld is out of bounds (>= %lld) for %.*s", int64(idx), max_type_count, isize(context_name.Len), context_name.Text)
					}
					operand := &Operand{}
					check_expr_with_type_hint(c, operand, e, elem_type)
					check_assignment(c, operand, elem_type, context_name)
					if is_constant {
						is_constant = check_is_operand_compound_lit_constant(c, operand, elem_type)
					}
				}
				if max_val < int64(idx) {
					max_val = int64(idx)
				}
			}
			was_error := false
			if len(cl.Elems) > 0 && cl.Elems[0].Kind != Ast_FieldValue {
				if 0 < max_val && max_val < t.EnumeratedArray.Count {
					error(node, "Expected %lld values for this enumerated array literal, got %lld", t.EnumeratedArray.Count, max_val)
					was_error = true
				} else {
					error(node, "Enumerated array literals must only have 'field = value' elements, bare elements are not allowed")
					was_error = true
				}
			}
			if len(cl.Elems) > 0 && !was_error && !is_partial {
				et := base_type(index_type)
				gb_assert_handler("Assertion Failure", "et.Kind == Type_Enum", "cmd_check_expr_compound.go", 0)
				fields := et.Enum.Fields
				unhandled := make([]*Entity, 0, len(fields))
				for _, f := range fields {
					if f.Kind != Entity_Constant {
						continue
					}
					v := f.Constant.Value
					hash := hash_exact_value(v)
					if _, ok := seen[hash]; !ok {
						unhandled = append(unhandled, f)
					}
				}
				if len(unhandled) > 0 {
					begin_error_block()
					if len(unhandled) == 1 {
						error_no_newline(node, "Unhandled enumerated array case: %.*s", isize(unhandled[0].Token.String.Len), unhandled[0].Token.String.Text)
					} else {
						error(node, "Unhandled enumerated array cases:")
						for _, f := range unhandled {
							error_line("\t%.*s\n", isize(f.Token.String.Len), f.Token.String.Text)
						}
					}
					if !build_context.terse_errors {
						error_line("\n")
						error_line("\tSuggestion: Was '#partial %s{...}' wanted?\n", type_to_string(typ))
					}
					end_error_block()
				}
			}
		}
	case Type_Basic:
		if !is_type_any(t) {
			if len(cl.Elems) != 0 {
				s := type_to_string(t)
				error(node, "Illegal compound literal, %s cannot be used as a compound literal with fields", s)
				gb_string_free(s)
				is_constant = false
			}
			break
		}
		if len(cl.Elems) == 0 {
			break
		}
		{
			field_types := [2]*Type{t_rawptr, t_typeid}
			field_count := isize(2)
			if cl.Elems[0].Kind == Ast_FieldValue {
				fields_visited := [2]bool{}
				for _, elem := range cl.Elems {
					if elem.Kind != Ast_FieldValue {
						error(elem, "Mixture of 'field = value' and value elements in a 'any' literal is not allowed")
						continue
					}
					fv := &elem.FieldValue
					if fv.Field.Kind != Ast_Ident {
						expr_str := expr_to_string(fv.Field)
						error(elem, "Invalid field name '%s' in 'any' literal", expr_str)
						gb_string_free(expr_str)
						continue
					}
					name := fv.Field.Ident.Token.String
					interned := fv.Field.Ident.Interned
					sel := lookup_field(typ, interned, o.Mode == Addressing_Type)
					if sel.Entity == nil {
						error(elem, "Unknown field '%.*s' in 'any' literal", isize(name.Len), name.Text)
						continue
					}
					idx := sel.Index[0]
					if fields_visited[idx] {
						error(elem, "Duplicate field '%.*s' in 'any' literal", isize(name.Len), name.Text)
						continue
					}
					fields_visited[idx] = true
					check_expr(c, o, fv.Value)
					is_constant = false
					check_assignment(c, o, field_types[idx], String{Text: strPtr("'any' literal"), Len: isize(len("'any' literal"))})
				}
			} else {
				for index := isize(0); index < isize(len(cl.Elems)); index++ {
					elem := cl.Elems[index]
					if elem.Kind == Ast_FieldValue {
						error(elem, "Mixture of 'field = value' and value elements in a 'any' literal is not allowed")
						continue
					}
					check_expr(c, o, elem)
					if index >= field_count {
						error(o.Expr, "Too many values in 'any' literal, expected %d", isize(field_count))
						break
					}
					is_constant = false
					check_assignment(c, o, field_types[index], String{Text: strPtr("'any' literal"), Len: isize(len("'any' literal"))})
				}
				if isize(len(cl.Elems)) < field_count {
					error(cl.Close, "Too few values in 'any' literal, expected %d, got %d", isize(field_count), isize(len(cl.Elems)))
				}
			}
		}
	case Type_Map:
		if len(cl.Elems) == 0 {
			break
		}
		is_constant = false
		{
			key_is_typeid := is_type_typeid(t.Map.Key)
			value_is_typeid := is_type_typeid(t.Map.Value)
			for _, elem := range cl.Elems {
				if elem.Kind != Ast_FieldValue {
					error(elem, "Only 'field = value' elements are allowed in a map literal")
					continue
				}
				fv := &elem.FieldValue
				if key_is_typeid {
					check_expr_or_type(c, o, fv.Field, t.Map.Key)
				} else {
					check_expr_with_type_hint(c, o, fv.Field, t.Map.Key)
				}
				check_assignment(c, o, t.Map.Key, String{Text: strPtr("map literal"), Len: isize(len("map literal"))})
				if o.Mode == Addressing_Invalid {
					continue
				}
				if value_is_typeid {
					check_expr_or_type(c, o, fv.Value, t.Map.Value)
				} else {
					check_expr_with_type_hint(c, o, fv.Value, t.Map.Value)
				}
				check_assignment(c, o, t.Map.Value, String{Text: strPtr("map literal"), Len: isize(len("map literal"))})
			}
		}
		if check_for_dynamic_literals(c, node, cl) {
			add_map_reserve_dependencies(c)
			add_map_set_dependencies(c)
		}
	case Type_BitSet:
		if len(cl.Elems) == 0 {
			break
		}
		et := base_type(t.BitSet.Elem)
		field_count := isize(0)
		if et != nil && et.Kind == Type_Enum {
			field_count = isize(len(et.Enum.Fields))
		}
		if is_type_array(bit_set_to_int(t)) {
			is_constant = false
		}
		for _, elem := range cl.Elems {
			if elem.Kind == Ast_FieldValue {
				error(elem, "'field = value' in a bit_set literal is not allowed")
				is_constant = false
				continue
			}
			check_expr_with_type_hint(c, o, elem, et)
			if is_constant {
				is_constant = o.Mode == Addressing_Constant
			}
			if elem.Kind == Ast_BinaryExpr {
				switch elem.BinaryExpr.Op.Kind {
				case Token_Or:
					x := expr_to_string(elem.BinaryExpr.Left)
					y := expr_to_string(elem.BinaryExpr.Right)
					e := expr_to_string(elem)
					error(elem, "Was the following intended? '%s, %s'; if not, surround the expression with parentheses '(%s)'", x, y, e)
					gb_string_free(e)
					gb_string_free(y)
					gb_string_free(x)
				}
			}
			check_assignment(c, o, t.BitSet.Elem, String{Text: strPtr("bit_set literal"), Len: isize(len("bit_set literal"))})
			if o.Mode == Addressing_Constant {
				lower := t.BitSet.Lower
				upper := t.BitSet.Upper
				v := exact_value_to_i64(o.Value)
				if lower <= v && v <= upper {
				} else {
					s := expr_to_string(o.Expr)
					error(elem, "Bit field value out of bounds, %s (%lld) not in the range %lld .. %lld", s, v, lower, upper)
					gb_string_free(s)
					continue
				}
			}
		}
	case Type_BitField:
		if len(cl.Elems) == 0 {
			break
		}
		is_constant = false
		if cl.Elems[0].Kind != Ast_FieldValue {
			type_str := type_to_string(typ)
			error(node, "%s ('bit_field') compound literals are only allowed to contain 'field = value' elements", type_str)
			gb_string_free(type_str)
		} else {
			check_compound_literal_field_values(c, cl.Elems, o, typ, &is_constant)
		}
	default:
		if len(cl.Elems) == 0 {
			break
		}
		str := type_to_string(typ)
		error(node, "Invalid compound literal type '%s'", str)
		gb_string_free(str)
		return kind
	}
	if is_constant {
		o.Mode = Addressing_Constant
		if is_type_bit_set(typ) {
			bt := base_type(typ)
			bits := BigInt{}
			one := BigInt{}
			big_int_from_u64(&one, 1)
			for _, e := range cl.Elems {
				gb_assert_handler("Assertion Failure", "e.Kind != Ast_FieldValue", "cmd_check_expr_compound.go", 0)
				tav := e.TAV
				if tav.Mode != Addressing_Constant {
					continue
				}
				if tav.Value.Kind != ExactValue_Integer {
					continue
				}
				v := big_int_to_i64(&tav.Value.ValueInteger)
				lower := bt.BitSet.Lower
				index := uint64(v - lower)
				bit := BigInt{}
				big_int_from_u64(&bit, index)
				big_int_shl(&bit, &one, &bit)
				big_int_or(&bits, &bits, &bit)
			}
			o.Value.Kind = ExactValue_Integer
			o.Value.ValueInteger = bits
		} else if is_type_constant_type(typ) && len(cl.Elems) == 0 {
			value := exact_value_compound(node)
			bt := core_type(typ)
			if bt.Kind == Type_Basic {
				if bt.Basic.Flags&BasicFlag_Boolean != 0 {
					value = exact_value_bool(false)
				} else if bt.Basic.Flags&BasicFlag_Integer != 0 {
					value = exact_value_i64(0)
				} else if bt.Basic.Flags&BasicFlag_Unsigned != 0 {
					value = exact_value_i64(0)
				} else if bt.Basic.Flags&BasicFlag_Float != 0 {
					value = exact_value_float(0)
				} else if bt.Basic.Flags&BasicFlag_Complex != 0 {
					value = exact_value_complex(0, 0)
				} else if bt.Basic.Flags&BasicFlag_Quaternion != 0 {
					value = exact_value_quaternion(0, 0, 0, 0)
				} else if bt.Basic.Flags&BasicFlag_Pointer != 0 {
					value = exact_value_pointer(0)
				} else if bt.Basic.Flags&BasicFlag_String != 0 {
					empty_string := String{}
					value = exact_value_string(empty_string)
				} else if bt.Basic.Flags&BasicFlag_Rune != 0 {
					value = exact_value_i64(0)
				}
			}
			o.Value = value
		} else {
			o.Value = exact_value_compound(node)
		}
	} else {
		o.Mode = Addressing_Value
	}
	o.Type = typ
	return kind
}

func check_type_assertion(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := Expr_Expr
	ta := &node.TypeAssertion
	check_expr(c, o, ta.Expr)
	node.ViralStateFlags |= ta.Expr.ViralStateFlags
	if o.Mode == Addressing_Invalid {
		o.Expr = node
		return kind
	}
	if o.Mode == Addressing_Constant {
		expr_str := expr_to_string(o.Expr)
		error(o.Expr, "A type assertion cannot be applied to a constant expression: '%s'", expr_str)
		gb_string_free(expr_str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	if is_type_untyped(o.Type) {
		expr_str := expr_to_string(o.Expr)
		error(o.Expr, "A type assertion cannot be applied to an untyped expression: '%s'", expr_str)
		gb_string_free(expr_str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	src := type_deref(o.Type)
	bsrc := base_type(src)
	if ta.Type != nil && ta.Type.Kind == Ast_UnaryExpr && ta.Type.UnaryExpr.Op.Kind == Token_Question {
		if !is_type_union(src) {
			str := type_to_string(o.Type)
			error(o.Expr, "Type assertions with .? can only operate on unions, got %s", str)
			gb_string_free(str)
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
		if len(bsrc.Union.Variants) != 1 && type_hint != nil {
			allowed := false
			for _, vt := range bsrc.Union.Variants {
				if are_types_identical(vt, type_hint) {
					allowed = true
					add_type_info_type(c, vt)
					break
				}
			}
			if allowed {
				add_type_info_type(c, o.Type)
				o.Type = type_hint
				o.Mode = Addressing_OptionalOk
				goto end_label
			}
		}
		if len(bsrc.Union.Variants) != 1 {
			error(o.Expr, "Type assertions with .? can only operate on unions with 1 variant, got %lld", int64(len(bsrc.Union.Variants)))
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
		add_type_info_type(c, o.Type)
		add_type_info_type(c, bsrc.Union.Variants[0])
		o.Type = bsrc.Union.Variants[0]
		o.Mode = Addressing_OptionalOk
	} else {
		t := check_type(c, ta.Type)
		dst := t
		if is_type_union(src) {
			ok := false
			for _, vt := range bsrc.Union.Variants {
				if are_types_identical(vt, dst) {
					ok = true
					break
				}
			}
			if !ok {
				expr_str := expr_to_string(o.Expr)
				dst_type_str := type_to_string(t)
				defer gb_string_free(expr_str)
				defer gb_string_free(dst_type_str)
				if len(bsrc.Union.Variants) == 0 {
					error(o.Expr, "Cannot type assert '%s' to '%s' as this is an empty union", expr_str, dst_type_str)
				} else {
					error(o.Expr, "Cannot type assert '%s' to '%s' as it is not a variant of that union", expr_str, dst_type_str)
				}
				o.Mode = Addressing_Invalid
				o.Expr = node
				return kind
			}
			add_type_info_type(c, o.Type)
			add_type_info_type(c, t)
			o.Type = t
			o.Mode = Addressing_OptionalOk
		} else if is_type_any(src) {
			o.Type = t
			o.Mode = Addressing_OptionalOk
			add_type_info_type(c, o.Type)
			add_type_info_type(c, t)
		} else {
			str := type_to_string(o.Type)
			error(o.Expr, "Type assertions can only operate on unions and 'any', got %s", str)
			gb_string_free(str)
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
	}
end_label:
	if (c.state_flags & StateFlag_no_type_assert) == 0 {
		has_context := true
		if c.proc_name.Len == 0 && c.curr_proc_sig == nil {
			has_context = false
		} else if (c.scope.Flags & ScopeFlag_ContextDefined) == 0 {
			has_context = false
		}
		if has_context {
			add_package_dependency(c, "runtime", "type_assertion_check_with_context", false)
			add_package_dependency(c, "runtime", "type_assertion_check2_with_context", false)
		} else {
			add_package_dependency(c, "runtime", "type_assertion_check_contextless", false)
			add_package_dependency(c, "runtime", "type_assertion_check2_contextless", false)
		}
	}
	return kind
}

func check_selector_call_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	se := &node.SelectorCallExpr
	if se.ModifiedCall {
		o.Expr = node
		o.Type = node.TAV.Type
		o.Value = node.TAV.Value
		o.Mode = node.TAV.Mode
		return Expr_Expr
	}
	allow_arrow_right_selector_expr := c.allow_arrow_right_selector_expr
	c.allow_arrow_right_selector_expr = true
	x := &Operand{}
	kind := check_expr_base(c, x, se.Expr, nil)
	c.allow_arrow_right_selector_expr = allow_arrow_right_selector_expr
	if x.Mode == Addressing_Invalid || (x.Type == t_invalid && x.Mode != Addressing_ProcGroup) {
		o.Mode = Addressing_Invalid
		o.Type = t_invalid
		o.Expr = node
		return kind
	}
	if !is_type_proc(x.Type) && x.Mode != Addressing_ProcGroup {
		type_str := type_to_string(x.Type)
		error(se.Call, "Selector call expressions expect a procedure type for the call, got '%s'", type_str)
		gb_string_free(type_str)
		o.Mode = Addressing_Invalid
		o.Type = t_invalid
		o.Expr = node
		return Expr_Stmt
	}
	ce := &se.Call.CallExpr
	gb_assert_handler("Assertion Failure", "x.Expr.Kind == Ast_SelectorExpr", "cmd_check_expr_compound.go", 0)
	first_arg := x.Expr.SelectorExpr.Expr
	gb_assert_handler("Assertion Failure", "first_arg != nil", "cmd_check_expr_compound.go", 0)
	e := entity_of_node(se.Expr)
	if !(e != nil && (e.Kind == Entity_Procedure || e.Kind == Entity_ProcGroup)) {
		first_arg.StateFlags |= StateFlag_SelectorCallExpr
	}
	if e.Kind != Entity_ProcGroup {
		pt := base_type(x.Type)
		gb_assert_handler("Assertion Failure", "pt.Kind == Type_Proc", "cmd_check_expr_compound.go", 0)
		var first_type *Type
		first_arg_name := String{}
		if pt.Proc.ParamCount > 0 {
			f := pt.Proc.Params.Tuple.Variables[0]
			first_type = f.Type
			first_arg_name = f.Token.String
		}
		if first_arg_name.Len == 0 {
			first_arg_name = String{Text: strPtr("_"), Len: isize(len("_"))}
		}
		if first_type == nil {
			error(se.Call, "Selector call expressions expect a procedure type for the call with at least 1 parameter")
			o.Mode = Addressing_Invalid
			o.Type = t_invalid
			o.Expr = node
			return Expr_Stmt
		}
		y := &Operand{}
		y.Mode = first_arg.TAV.Mode
		y.Type = first_arg.TAV.Type
		y.Value = first_arg.TAV.Value
		if check_is_assignable_to(c, y, first_type) {
		} else {
			z := *y
			z.Type = type_deref(y.Type)
			if check_is_assignable_to(c, &z, first_type) {
				op := Token{Kind: Token_Pointer}
				first_arg = ast_deref_expr(first_arg.File(), first_arg, op)
			} else if y.Mode == Addressing_Variable {
				w := *y
				w.Type = alloc_type_pointer(y.Type)
				if check_is_assignable_to(c, &w, first_type) {
					op := Token{Kind: Token_And}
					first_arg = ast_unary_expr(first_arg.File(), op, first_arg)
				}
			}
		}
		if len(ce.Args) > 0 {
			fail := false
			first_is_field_value := ce.Args[0].Kind == Ast_FieldValue
			for _, arg := range ce.Args {
				mix := false
				if first_is_field_value {
					mix = arg.Kind != Ast_FieldValue
				} else {
					mix = arg.Kind == Ast_FieldValue
				}
				if mix {
					fail = true
					break
				}
			}
			if !fail && first_is_field_value {
				op := Token{Kind: Token_Eq}
				f := first_arg.File()
				first_arg = ast_field_value(f, ast_ident(f, make_token_ident(first_arg_name)), first_arg, op)
			}
		}
	}
	modified_args := make([]*Ast, len(ce.Args)+1)
	modified_args[0] = first_arg
	copy(modified_args[1:], ce.Args)
	ce.Args = modified_args
	se.ModifiedCall = true
	allow_arrow_right_selector_expr = c.allow_arrow_right_selector_expr
	c.allow_arrow_right_selector_expr = true
	check_expr_base(c, o, se.Call, type_hint)
	c.allow_arrow_right_selector_expr = allow_arrow_right_selector_expr
	o.Expr = node
	return Expr_Expr
}

func check_index_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := Expr_Expr
	ie := &node.IndexExpr
	check_expr(c, o, ie.Expr)
	node.ViralStateFlags |= ie.Expr.ViralStateFlags
	if o.Mode == Addressing_Invalid {
		o.Expr = node
		return kind
	}
	t := base_type(type_deref(o.Type))
	is_ptr := is_type_pointer(o.Type)
	is_const := o.Mode == Addressing_Constant
	if is_type_map(t) {
		key := &Operand{}
		if is_type_typeid(t.Map.Key) {
			check_expr_or_type(c, key, ie.Index, t.Map.Key)
		} else {
			check_expr_with_type_hint(c, key, ie.Index, t.Map.Key)
		}
		check_assignment(c, key, t.Map.Key, String{Text: strPtr("map index"), Len: isize(len("map index"))})
		if key.Mode == Addressing_Invalid {
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
		o.Mode = Addressing_MapIndex
		o.Type = t.Map.Value
		o.Expr = node
		add_map_get_dependencies(c)
		add_map_set_dependencies(c)
		return Expr_Expr
	}
	max_count := int64(-1)
	valid := check_set_index_data(o, t, is_ptr, &max_count, o.Type)
	if is_const {
		if is_type_array(t) {
		} else if is_type_slice(t) {
		} else if is_type_enumerated_array(t) {
		} else if is_type_fixed_capacity_dynamic_array(t) {
		} else if is_type_string(t) {
		} else if is_type_matrix(t) {
		} else {
			valid = false
		}
	}
	if !valid {
		str := expr_to_string(o.Expr)
		type_str := type_to_string(o.Type)
		defer gb_string_free(str)
		defer gb_string_free(type_str)
		if is_const {
			error(o.Expr, "Cannot index constant '%s' of type '%s'", str, type_str)
		} else {
			error(o.Expr, "Cannot index '%s' of type '%s'", str, type_str)
		}
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	if ie.Index == nil {
		str := expr_to_string(o.Expr)
		error(o.Expr, "Missing index for '%s'", str)
		gb_string_free(str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	var index_type_hint *Type
	if is_type_enumerated_array(t) {
		bt := base_type(t)
		gb_assert_handler("Assertion Failure", "bt.Kind == Type_EnumeratedArray", "cmd_check_expr_compound.go", 0)
		index_type_hint = bt.EnumeratedArray.Index
	}
	index := int64(0)
	ok := check_index_value(c, t, false, ie.Index, max_count, &index, index_type_hint)
	if is_const {
		if index < 0 {
			begin_error_block()
			str := expr_to_string(o.Expr)
			error(o.Expr, "Cannot index a constant '%s'", str)
			if !build_context.terse_errors {
				error_line("\tSuggestion: store the constant into a variable in order to index it with a variable index\n")
			}
			gb_string_free(str)
			end_error_block()
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		} else if ok && !is_type_matrix(t) {
			tav := type_and_value_of_expr(ie.Expr)
			value := tav.Value
			o.Mode = Addressing_Constant
			success := false
			finish := false
			o.Value = get_constant_field_single(c, value, int32(index), &success, &finish)
			if !success {
				begin_error_block()
				str := expr_to_string(o.Expr)
				error(o.Expr, "Cannot index a constant '%s' with index %lld", str, index)
				if !build_context.terse_errors {
					error_line("\tSuggestion: store the constant into a variable in order to index it with a variable index\n")
				}
				gb_string_free(str)
				end_error_block()
				o.Mode = Addressing_Invalid
				o.Expr = node
				return kind
			}
		}
	}
	if type_hint != nil && is_type_matrix(t) {
	}
	return kind
}

func check_slice_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := Expr_Stmt
	se := &node.SliceExpr
	check_expr(c, o, se.Expr)
	node.ViralStateFlags |= se.Expr.ViralStateFlags
	if o.Mode == Addressing_Invalid {
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	valid := false
	max_count := int64(-1)
	t := base_type(type_deref(o.Type))
	switch t.Kind {
	case Type_Basic:
		if t.Basic.Kind == Basic_string || t.Basic.Kind == Basic_UntypedString {
			valid = true
			if o.Mode == Addressing_Constant {
				gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String", "cmd_check_expr_compound.go", 0)
				max_count = int64(o.Value.ValueString.Len)
			}
			o.Type = type_deref(o.Type)
		} else if t.Basic.Kind == Basic_string16 {
			valid = true
			if o.Mode == Addressing_Constant {
				gb_assert_handler("Assertion Failure", "o.Value.Kind == ExactValue_String16", "cmd_check_expr_compound.go", 0)
				max_count = int64(o.Value.ValueString16.Len)
			}
			o.Type = type_deref(o.Type)
		}
	case Type_Array:
		valid = true
		max_count = t.Array.Count
		if o.Mode != Addressing_Variable && !is_type_pointer(o.Type) {
			str := expr_to_string(node)
			error(node, "Cannot slice array '%s', value is not addressable", str)
			gb_string_free(str)
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
		o.Type = alloc_type_slice(t.Array.Elem)
	case Type_MultiPointer:
		valid = true
		o.Type = type_deref(o.Type)
	case Type_Slice:
		valid = true
		o.Type = type_deref(o.Type)
	case Type_DynamicArray:
		valid = true
		o.Type = alloc_type_slice(t.DynamicArray.Elem)
	case Type_FixedCapacityDynamicArray:
		valid = true
		if o.Mode != Addressing_Variable && !is_type_pointer(o.Type) {
			str := expr_to_string(node)
			error(node, "Cannot slice a fixed capacity dynamic array '%s', value is not addressable", str)
			gb_string_free(str)
			o.Mode = Addressing_Invalid
			o.Expr = node
			return kind
		}
		o.Type = alloc_type_slice(t.FixedCapacityDynamicArray.Elem)
	case Type_Struct:
		if is_type_soa_struct(t) {
			valid = true
			if t.Struct.SoaKind == StructSoaFixed {
				max_count = int64(t.Struct.SoaCount)
				if o.Mode != Addressing_Variable && !is_type_pointer(o.Type) {
					str := expr_to_string(node)
					error(node, "Cannot slice #soa array '%s', value is not addressable", str)
					gb_string_free(str)
					o.Mode = Addressing_Invalid
					o.Expr = node
					return kind
				}
			}
			o.Type = make_soa_struct_slice(c, nil, nil, t.Struct.SoaElem)
		}
	case Type_EnumeratedArray:
		begin_error_block()
		str := expr_to_string(o.Expr)
		type_str := type_to_string(o.Type)
		error(o.Expr, "Cannot slice '%s' of type '%s', as enumerated arrays cannot be sliced", str, type_str)
		error_line("\tSuggestion: Slicing an enumerated array does not make much sense, but if you need such a construct, use 'slice.enumerated_array'\n")
		gb_string_free(type_str)
		gb_string_free(str)
		end_error_block()
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	if !valid {
		str := expr_to_string(o.Expr)
		type_str := type_to_string(o.Type)
		error(o.Expr, "Cannot slice '%s' of type '%s'", str, type_str)
		gb_string_free(type_str)
		gb_string_free(str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return kind
	}
	indices := [2]int64{}
	nodes := [2]*Ast{se.Low, se.High}
	for i := 0; i < 2; i++ {
		idx := max_count
		if nodes[i] != nil {
			capacity := int64(-1)
			if max_count >= 0 {
				capacity = max_count
			}
			j := int64(0)
			if check_index_value(c, t, true, nodes[i], capacity, &j, nil) {
				idx = j
			}
			node.ViralStateFlags |= nodes[i].ViralStateFlags
		} else if i == 0 {
			idx = 0
		}
		indices[i] = idx
	}
	for i := 0; i < 2; i++ {
		a := indices[i]
		for j := i + 1; j < 2; j++ {
			b := indices[j]
			if a > b && b >= 0 {
				error(se.Close, "Invalid slice indices: [%d > %d]", a, b)
			}
		}
	}
	if max_count < 0 {
		if o.Mode == Addressing_Constant {
			s := expr_to_string(se.Expr)
			error(se.Expr, "Cannot slice constant value '%s'", s)
			gb_string_free(s)
		}
	}
	if t.Kind == Type_MultiPointer && se.High != nil {
		o.Type = alloc_type_slice(t.MultiPointer.Elem)
	}
	o.Mode = Addressing_Value
	if is_type_string(t) && max_count >= 0 {
		all_constant := true
		for i := 0; i < 2; i++ {
			if nodes[i] != nil {
				tav := type_and_value_of_expr(nodes[i])
				if tav.Mode != Addressing_Constant {
					all_constant = false
					break
				}
			}
		}
		if !all_constant {
			begin_error_block()
			str := expr_to_string(o.Expr)
			error(o.Expr, "Cannot slice '%s' with non-constant indices", str)
			if !build_context.terse_errors {
				error_line("\tSuggestion: store the constant into a variable in order to index it with a variable index\n")
			}
			gb_string_free(str)
			end_error_block()
			o.Mode = Addressing_Value
			o.Expr = node
			return kind
		}
		o.Mode = Addressing_Constant
		o.Type = t
		if o.Value.Kind == ExactValue_String16 {
			s16 := o.Value.ValueString16
			o.Value = exact_value_string16(substring(s16, isize(indices[0]), isize(indices[1])))
		} else {
			s := String{}
			if o.Value.Kind == ExactValue_String {
				s = o.Value.ValueString
			}
			o.Value = exact_value_string(substring(s, isize(indices[0]), isize(indices[1])))
		}
	}
	return kind
}

func check_matrix_index_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) {
	ie := &node.MatrixIndexExpr
	check_expr(c, o, ie.Expr)
	node.ViralStateFlags |= ie.Expr.ViralStateFlags
	if o.Mode == Addressing_Invalid {
		o.Expr = node
		return
	}
	t := base_type(type_deref(o.Type))
	is_ptr := is_type_pointer(o.Type)
	is_const := o.Mode == Addressing_Constant
	if t.Kind != Type_Matrix {
		str := expr_to_string(o.Expr)
		type_str := type_to_string(o.Type)
		defer gb_string_free(str)
		defer gb_string_free(type_str)
		if is_const {
			error(o.Expr, "Cannot use matrix indexing on constant '%s' of type '%s'", str, type_str)
		} else {
			error(o.Expr, "Cannot use matrix indexing on '%s' of type '%s'", str, type_str)
		}
		o.Mode = Addressing_Invalid
		o.Expr = node
		return
	}
	o.Type = t.Matrix.Elem
	if is_ptr {
		o.Mode = Addressing_Variable
	} else if o.Mode != Addressing_Variable {
		o.Mode = Addressing_Value
	}
	if ie.RowIndex == nil {
		str := expr_to_string(o.Expr)
		error(o.Expr, "Missing row index for '%s'", str)
		gb_string_free(str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return
	}
	if ie.ColumnIndex == nil {
		str := expr_to_string(o.Expr)
		error(o.Expr, "Missing column index for '%s'", str)
		gb_string_free(str)
		o.Mode = Addressing_Invalid
		o.Expr = node
		return
	}
	row_count := t.Matrix.RowCount
	column_count := t.Matrix.ColumnCount
	row_index := int64(0)
	column_index := int64(0)
	row_ok := check_index_value(c, t, false, ie.RowIndex, row_count, &row_index, nil)
	column_ok := check_index_value(c, t, false, ie.ColumnIndex, column_count, &column_index, nil)
	if is_const && (ie.RowIndex.TAV.Mode != Addressing_Constant || ie.ColumnIndex.TAV.Mode != Addressing_Constant) {
		error(o.Expr, "Cannot index constant matrix with non-constant indices '%s'", expr_to_string(node))
	}
	_ = row_ok
	_ = column_ok
}

func attempt_implicit_selector_expr(c *CheckerContext, o *Operand, ise *AstImplicitSelectorExpr, th *Type) bool {
	if is_type_enum(th) {
		enum_type := base_type(th)
		gb_assert_handler("Assertion Failure", "enum_type.Kind == Type_Enum", "cmd_check_expr_compound.go", 0)
		name := ise.Selector.Ident.Interned
		e := scope_lookup_current(enum_type.Enum.Scope, name, 0)
		if e == nil {
			return false
		}
		gb_assert_handler("Assertion Failure", "are_types_identical(base_type(e.Type), enum_type)", "cmd_check_expr_compound.go", 0)
		gb_assert_handler("Assertion Failure", "e.Kind == Entity_Constant", "cmd_check_expr_compound.go", 0)
		o.Value = e.Constant.Value
		o.Mode = Addressing_Constant
		o.Type = e.Type
		return true
	}
	if is_type_union(th) {
		union_type := base_type(th)
		operands := make([]Operand, 0, len(union_type.Union.Variants))
		for _, vt := range union_type.Union.Variants {
			var x Operand
			if attempt_implicit_selector_expr(c, &x, ise, vt) {
				operands = append(operands, x)
			}
		}
		if len(operands) == 1 {
			*o = operands[0]
			return true
		}
	}
	return false
}

func check_implicit_selector_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	ise := &node.ImplicitSelectorExpr
	o.Type = t_invalid
	o.Expr = node
	o.Mode = Addressing_Invalid
	th := type_hint
	if th == nil {
		str := expr_to_string(node)
		error(node, "Cannot determine type for implicit selector expression '%s'", str)
		gb_string_free(str)
		return Expr_Expr
	}
	o.Type = th
	ok := attempt_implicit_selector_expr(c, o, ise, th)
	if !ok {
		name := ise.Selector.Ident.Token.String
		if is_type_enum(th) {
			begin_error_block()
			bt := base_type(th)
			gb_assert_handler("Assertion Failure", "bt.Kind == Type_Enum", "cmd_check_expr_compound.go", 0)
			typ := type_to_string(th)
			defer gb_string_free(typ)
			error(node, "Undeclared name '%.*s' for type '%s'", isize(name.Len), name.Text, typ)
			check_did_you_mean_type(name, bt.Enum.Fields)
			end_error_block()
		} else if is_type_bit_set(th) && is_type_enum(th.BitSet.Elem) {
			begin_error_block()
			typ := type_to_string(th)
			str := expr_to_string(node)
			error(node, "Cannot convert enum value to '%s'", typ)
			error_line("\tSuggestion: Did you mean '{ %s }'?\n", str)
			gb_string_free(typ)
			gb_string_free(str)
			end_error_block()
		} else {
			typ := type_to_string(th)
			str := expr_to_string(node)
			error(node, "Invalid type '%s' for implicit selector expression '%s'", typ, str)
			gb_string_free(str)
			gb_string_free(typ)
		}
	}
	o.Expr = node
	return Expr_Expr
}

func check_promote_optional_ok(c *CheckerContext, x *Operand, val_type_ **Type, ok_type_ **Type, change_operand bool) {
	switch x.Mode {
	case Addressing_MapIndex:
	case Addressing_OptionalOk:
	case Addressing_OptionalOkPtr:
		if val_type_ != nil {
			*val_type_ = x.Type
		}
	default:
		if ok_type_ != nil {
			*ok_type_ = x.Type
		}
		return
	}
	expr := unparen_expr(x.Expr)
	if expr.Kind == Ast_CallExpr {
		pt := base_type(type_of_expr(expr.CallExpr.Proc))
		if is_type_proc(pt) {
			tuple := pt.Proc.Results
			if pt.Proc.ResultCount >= 2 {
				if ok_type_ != nil {
					*ok_type_ = tuple.Tuple.Variables[1].Type
				}
				if change_operand {
					expr.CallExpr.OptionalOkOne = false
					x.Type = tuple
					add_type_and_value(c, x.Expr, x.Mode, tuple, x.Value)
				}
				return
			}
		}
	}
	tuple := make_optional_ok_type(x.Type)
	if ok_type_ != nil {
		*ok_type_ = tuple.Tuple.Variables[1].Type
	}
	if change_operand {
		add_type_and_value(c, x.Expr, x.Mode, tuple, x.Value)
		x.Type = tuple
		gb_assert_handler("Assertion Failure", "is_type_tuple(type_of_expr(x.Expr))", "cmd_check_expr_compound.go", 0)
	}
}

func add_constant_switch_case(ctx *CheckerContext, seen *SeenMap, operand Operand, use_expr ...bool) {
	useExpr := true
	if len(use_expr) > 0 {
		useExpr = use_expr[0]
	}
	if operand.Mode != Addressing_Constant {
		return
	}
	if operand.Value.Kind == ExactValue_Invalid {
		return
	}
	key := hash_exact_value(operand.Value)
	gb_assert_handler("Assertion Failure", "key != 0", "cmd_check_expr_compound.go", 0)
	count := len((*seen)[key])
	if count > 0 {
		taps := (*seen)[key]
		for _, tap := range taps {
			to := &Operand{}
			to.Mode = Addressing_Value
			to.Type = &tap.Type
			if !check_is_assignable_to_with_score(ctx, to, operand.Type, nil) {
				continue
			}
			pos := tap.Token.Pos
			if useExpr {
				expr_str := expr_to_string(operand.Expr)
				error(operand.Expr,
					"Duplicate case '%s'\n"+
						"\tprevious case at %s",
					expr_str,
					token_pos_to_string(pos))
				gb_string_free(expr_str)
			} else {
				error(operand.Expr, "Duplicate case found with previous case at %s", token_pos_to_string(pos))
			}
			return
		}
	}
	tap := TypeAndToken{Type: *operand.Type, Token: ast_token(operand.Expr)}
	(*seen)[key] = append((*seen)[key], tap)
}

func add_to_seen_map(ctx *CheckerContext, seen *SeenMap, upper_op TokenKind, x Operand, lhs Operand, rhs Operand) {
	if is_type_enum(x.Type) {
		v0 := exact_value_to_i64(lhs.Value)
		v1 := exact_value_to_i64(rhs.Value)
		v := Operand{}
		v.Mode = Addressing_Constant
		v.Type = x.Type
		v.Expr = x.Expr
		bt := base_type(x.Type)
		gb_assert_handler("Assertion Failure", "bt.Kind == Type_Enum", "cmd_check_expr_compound.go", 0)
		for vi := v0; vi <= v1; vi++ {
			if upper_op != Token_LtEq && vi == v1 {
				break
			}
			v.Value = exact_value_i64(vi)
			add_constant_switch_case(ctx, seen, v)
		}
	} else {
		add_constant_switch_case(ctx, seen, lhs)
		if upper_op == Token_LtEq {
			add_constant_switch_case(ctx, seen, rhs)
		}
	}
}

func add_to_seen_map_single(ctx *CheckerContext, seen *SeenMap, x Operand) {
	add_constant_switch_case(ctx, seen, x)
}

func check_basic_directive_expr(c *CheckerContext, o *Operand, node *Ast, type_hint *Type) ExprKind {
	kind := Expr_Expr
	o.Mode = Addressing_Constant
	bd := &node.BasicDirective
	name := bd.Name.String
	if name == "file" {
		file := get_file_path_string(bd.Token.Pos.FileId)
		switch build_context.source_code_location_info {
		case SourceCodeLocationInfo_Normal:
		case SourceCodeLocationInfo_Obfuscated:
			file = obfuscate_string(file, "F")
		case SourceCodeLocationInfo_Filename:
			file = last_path_element(file)
		case SourceCodeLocationInfo_None:
			file = String{Text: strPtr(""), Len: 0}
		}
		o.Type = t_untyped_string
		o.Value = exact_value_string(file)
	} else if name == "directory" {
		file := get_file_path_string(bd.Token.Pos.FileId)
		path := dir_from_path(file)
		switch build_context.source_code_location_info {
		case SourceCodeLocationInfo_Normal:
		case SourceCodeLocationInfo_Obfuscated:
			path = obfuscate_string(path, "D")
		case SourceCodeLocationInfo_Filename:
			path = last_path_element(path)
		case SourceCodeLocationInfo_None:
			path = String{Text: strPtr(""), Len: 0}
		}
		o.Type = t_untyped_string
		o.Value = exact_value_string(path)
	} else if name == "line" {
		line := bd.Token.Pos.Line
		switch build_context.source_code_location_info {
		case SourceCodeLocationInfo_Normal:
		case SourceCodeLocationInfo_Obfuscated:
			line = obfuscate_i32(line)
		case SourceCodeLocationInfo_Filename:
		case SourceCodeLocationInfo_None:
			line = 0
		}
		o.Type = t_untyped_integer
		o.Value = exact_value_i64(int64(line))
	} else if name == "procedure" {
		if c.curr_proc_decl == nil {
			error(node, "#procedure may only be used within procedures")
			o.Type = t_untyped_string
			o.Value = exact_value_string(String{Text: strPtr(""), Len: 0})
		} else {
			p := c.proc_name
			switch build_context.source_code_location_info {
			case SourceCodeLocationInfo_Normal:
			case SourceCodeLocationInfo_Obfuscated:
				p = obfuscate_string(p, "P")
			case SourceCodeLocationInfo_Filename:
			case SourceCodeLocationInfo_None:
				p = String{Text: strPtr(""), Len: 0}
			}
			o.Type = t_untyped_string
			o.Value = exact_value_string(p)
		}
	} else if name == "caller_location" {
		init_core_source_code_location(c.checker)
		error(node, "#caller_location may only be used as a default argument parameter")
		o.Type = t_source_code_location
		o.Mode = Addressing_Value
	} else if name == "caller_expression" {
		error(node, "#caller_expression may only be used as a default argument parameter")
		o.Type = t_string
		o.Mode = Addressing_Value
	} else if name == "branch_location" {
		if !c.in_defer {
			error(node, "#branch_location may only be used within a 'defer' statement")
		} else if c.curr_proc_decl != nil {
			e := c.curr_proc_decl.Entity
			if e != nil {
				gb_assert_handler("Assertion Failure", "e.Kind == Entity_Procedure", "cmd_check_expr_compound.go", 0)
				e.Procedure.UsesBranchLocation = true
			}
		}
		o.Type = t_source_code_location
		o.Mode = Addressing_Value
	} else {
		if name == "location" {
			init_core_source_code_location(c.checker)
			error(node, "'#location' must be used as a call, i.e. #location(proc), where #location() defaults to the procedure in which it was used.")
			o.Type = t_source_code_location
			o.Mode = Addressing_Value
		} else if name == "assert" ||
			name == "defined" ||
			name == "config" ||
			name == "exists" ||
			name == "load" ||
			name == "load_hash" ||
			name == "load_directory" ||
			name == "load_or" {
			error(node, "'#%.*s' must be used as a call", isize(name.Len), name.Text)
			o.Type = t_invalid
			o.Mode = Addressing_Invalid
		} else {
			error(node, "Unknown directive: #%.*s", isize(name.Len), name.Text)
			o.Type = t_invalid
			o.Mode = Addressing_Invalid
		}
	}
	return kind
}
