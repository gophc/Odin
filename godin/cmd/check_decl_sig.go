// Depends on: common.odin (Type, String, isize, i64, core_type, base_type, type_size_of, type_align_of, are_types_identical, bit_set_to_int, gb_assert_handler, error, concatenate_strings, permanent_allocator, is_type_*)

package cmd

func sig_compare(a func(*Type) bool, x, y *Type) bool {
	x = core_type(x)
	y = core_type(y)
	return a(x) && a(y)
}

func sig_compare_pair(a, b func(*Type) bool, x, y *Type) bool {
	x = core_type(x)
	y = core_type(y)
	return (a(x) && b(y)) || (b(x) && a(y))
}

func signature_parameter_similar_enough(x, y *Type) bool {
	if is_type_bit_set(x) {
		x = bit_set_to_int(x)
	}
	if is_type_bit_set(y) {
		y = bit_set_to_int(y)
	}

	if sig_compare(is_type_pointer, x, y) {
		return true
	}
	if sig_compare(is_type_multi_pointer, x, y) {
		return true
	}
	if sig_compare(is_type_proc, x, y) {
		return true
	}

	if sig_compare(is_type_integer, x, y) {
		gb_assert_handler(core_type(x).Kind == Type_Basic, "core_type(x).Kind == Type_Basic")
		gb_assert_handler(core_type(y).Kind == Type_Basic, "core_type(y).Kind == Type_Basic")
		sx := type_size_of(x)
		sy := type_size_of(y)
		if sx == sy {
			return true
		}
	}

	if sig_compare_pair(is_type_integer, is_type_boolean, x, y) {
		gb_assert_handler(core_type(x).Kind == Type_Basic, "core_type(x).Kind == Type_Basic")
		gb_assert_handler(core_type(y).Kind == Type_Basic, "core_type(y).Kind == Type_Basic")
		sx := type_size_of(x)
		sy := type_size_of(y)
		if sx == sy {
			return true
		}
	}

	if sig_compare_pair(is_type_cstring, is_type_u8_ptr, x, y) {
		return true
	}
	if sig_compare_pair(is_type_cstring, is_type_u8_multi_ptr, x, y) {
		return true
	}
	if sig_compare_pair(is_type_cstring16, is_type_u16_ptr, x, y) {
		return true
	}
	if sig_compare_pair(is_type_cstring16, is_type_u16_multi_ptr, x, y) {
		return true
	}

	if sig_compare_pair(is_type_uintptr, is_type_rawptr, x, y) {
		return true
	}

	if sig_compare_pair(is_type_proc, is_type_pointer, x, y) {
		return true
	}
	if sig_compare_pair(is_type_pointer, is_type_multi_pointer, x, y) {
		return true
	}
	if sig_compare_pair(is_type_proc, is_type_multi_pointer, x, y) {
		return true
	}

	if sig_compare(is_type_slice, x, y) {
		s1 := core_type(x)
		s2 := core_type(y)
		if signature_parameter_similar_enough(s1.Slice.Elem, s2.Slice.Elem) {
			return true
		}
	}

	x_base := base_type(x)
	y_base := base_type(y)

	if x_base == y_base {
		return true
	}

	if x_base.Kind == y_base.Kind &&
		x_base.Kind == Type_Struct {
		xs := type_size_of(x_base)
		ys := type_size_of(y_base)

		xa := type_align_of(x_base)
		ya := type_align_of(y_base)

		if x_base.Struct.IsRawUnion == y_base.Struct.IsRawUnion &&
			xs == ys && xa == ya {
			if xs > 16 {
				return true
			}
			if x_base.Struct.IsRawUnion {
				return true
			}
			if len(x_base.Struct.Fields) == len(y_base.Struct.Fields) {
				for i := range x_base.Struct.Fields {
					a := x_base.Struct.Fields[i]
					b := y_base.Struct.Fields[i]
					if !signature_parameter_similar_enough(a.Type, b.Type) {
						goto end
					}
				}
			}
			return true
		}
	}

end:
	return are_types_identical(x, y)
}

func are_signatures_similar_enough(a_, b_ *Type) bool {
	gb_assert_handler(a_.Kind == Type_Proc, "a_.Kind == Type_Proc")
	gb_assert_handler(b_.Kind == Type_Proc, "b_.Kind == Type_Proc")
	a := &a_.Proc
	b := &b_.Proc

	if a.ParamCount != b.ParamCount {
		return false
	}
	if a.ResultCount != b.ResultCount {
		return false
	}

	if a.CVararg != b.CVararg {
		return false
	}

	if a.Variadic != b.Variadic {
		return false
	}

	if a.Variadic && a.VariadicIndex != b.VariadicIndex {
		return false
	}

	for i := int32(0); i < a.ParamCount; i++ {
		x := core_type(a.Params.Tuple.Variables[i].Type)
		y := core_type(b.Params.Tuple.Variables[i].Type)

		if x.Kind == Type_BitSet && x.BitSet.Underlying != nil {
			x = core_type(x.BitSet.Underlying)
		}
		if y.Kind == Type_BitSet && y.BitSet.Underlying != nil {
			y = core_type(y.BitSet.Underlying)
		}

		if a.Variadic && i == a.VariadicIndex {
			gb_assert_handler(x.Kind == Type_Slice, "x.Kind == Type_Slice")
			gb_assert_handler(y.Kind == Type_Slice, "y.Kind == Type_Slice")
			x_elem := core_type(x.Slice.Elem)
			y_elem := core_type(y.Slice.Elem)
			if is_type_any(x_elem) || is_type_any(y_elem) {
				continue
			}
		}

		if !signature_parameter_similar_enough(x, y) {
			return false
		}
	}

	for i := int32(0); i < a.ResultCount; i++ {
		x := core_type(a.Results.Tuple.Variables[i].Type)
		y := core_type(b.Results.Tuple.Variables[i].Type)

		if x.Kind == Type_BitSet && x.BitSet.Underlying != nil {
			x = core_type(x.BitSet.Underlying)
		}
		if y.Kind == Type_BitSet && y.BitSet.Underlying != nil {
			y = core_type(y.BitSet.Underlying)
		}

		if !signature_parameter_similar_enough(x, y) {
			return false
		}
	}

	return true
}

func handle_link_name(ctx *CheckerContext, token Token, link_name String, link_prefix String, link_suffix String) String {
	original_link_name := link_name
	if link_prefix.Len > 0 {
		if original_link_name.Len > 0 {
			error_pos(token.Pos, "'link_name' and 'link_prefix' cannot be used together")
		} else {
			link_name = concatenate_strings(permanent_allocator(), link_prefix, token.String)
		}
	}

	if link_suffix.Len > 0 {
		if original_link_name.Len > 0 {
			error_pos(token.Pos, "'link_name' and 'link_suffix' cannot be used together")
		} else {
			new_name := token.String
			if link_name.Data != original_link_name.Data || link_name.Len != original_link_name.Len {
				new_name = link_name
			}
			link_name = concatenate_strings(permanent_allocator(), new_name, link_suffix)
		}
	}

	return link_name
}
