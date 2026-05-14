package cmd

type lbCompoundLitElemTempData struct {
	Value      lbValue
	Expr       *Ast
	ElemIndex  i64
	ElemLength i64
	Gep        lbValue
}

func lb_build_addr_compound_lit_populate(p *lbProcedure, elems []*Ast, temp_data *[]lbCompoundLitElemTempData, compound_type *Type) {
	bt := base_type(compound_type)
	var et *Type
	switch bt.Kind {
	case TypeArray:
		et = bt.Array.Elem
	case TypeEnumeratedArray:
		et = bt.EnumeratedArray.Elem
	case TypeSlice:
		et = bt.Slice.Elem
	case TypeBitSet:
		et = bt.BitSet.Elem
	case TypeDynamicArray:
		et = bt.DynamicArray.Elem
	case TypeSimdVector:
		et = bt.SimdVector.Elem
	case TypeMatrix:
		et = bt.Matrix.Elem
	case TypeFixedCapacityDynamicArray:
		et = bt.FixedCapacityDynamicArray.Elem
	}

	elem_index := isize(0)
	for _, elem := range elems {
		if elem.Kind == Ast_FieldValue {
			fv := &elem.FieldValue
			if bt.Kind != TypeDynamicArray && lb_is_elem_const(fv.Value, et) {
				continue
			}
			if is_ast_range(fv.Field) {
				ie := &fv.Field.BinaryExpr
				lo_tav := ie.Left.Tav
				hi_tav := ie.Right.Tav
				op := ie.Op.Kind
				lo := exact_value_to_i64(lo_tav.Value)
				hi := exact_value_to_i64(hi_tav.Value)
				if op != Token_RangeHalf {
					hi += 1
				}

				value := lb_emit_conv(p, lb_build_expr(p, fv.Value), et)

				if (hi - lo) > 0 {
					if bt.Kind == TypeMatrix {
						for k := lo; k < hi; k++ {
							data := lbCompoundLitElemTempData{}
							data.Value = value
							data.ElemIndex = matrix_row_major_index_to_offset(bt, k)
							*temp_data = append(*temp_data, data)
						}
					} else {
						const MAX_ELEMENT_AMOUNT = 32
						if (hi - lo) <= MAX_ELEMENT_AMOUNT {
							for k := lo; k < hi; k++ {
								data := lbCompoundLitElemTempData{}
								data.Value = value
								data.ElemIndex = k
								*temp_data = append(*temp_data, data)
							}
						} else {
							data := lbCompoundLitElemTempData{}
							data.Value = value
							data.ElemIndex = lo
							data.ElemLength = hi - lo
							*temp_data = append(*temp_data, data)
						}
					}
				}
			} else {
				tav := fv.Field.Tav
				index := exact_value_to_i64(tav.Value)

				value := lb_emit_conv(p, lb_build_expr(p, fv.Value), et)

				data := lbCompoundLitElemTempData{}
				data.Value = value
				data.Expr = fv.Value
				if bt.Kind == TypeMatrix {
					data.ElemIndex = matrix_row_major_index_to_offset(bt, index)
				} else {
					data.ElemIndex = index
				}
				*temp_data = append(*temp_data, data)
			}
		} else {
			if bt.Kind != TypeDynamicArray && lb_is_elem_const(elem, et) {
				elem_index++
				continue
			}

			field_expr := lb_build_expr(p, elem)

			if is_type_tuple(field_expr.Type) {
				tt := &field_expr.Type.Tuple
				for jj := 0; jj < len(tt.Variables); jj++ {
					sub_field_expr := lb_emit_struct_ev(p, field_expr, i32(jj))
					ev := lb_emit_conv(p, sub_field_expr, et)

					data := lbCompoundLitElemTempData{}
					data.Value = ev
					if bt.Kind == TypeMatrix {
						data.ElemIndex = matrix_row_major_index_to_offset(bt, elem_index)
					} else {
						data.ElemIndex = i64(elem_index)
					}
					*temp_data = append(*temp_data, data)
					elem_index++
				}
			} else {
				ev := lb_emit_conv(p, field_expr, et)

				data := lbCompoundLitElemTempData{}
				data.Value = ev
				if bt.Kind == TypeMatrix {
					data.ElemIndex = matrix_row_major_index_to_offset(bt, elem_index)
				} else {
					data.ElemIndex = i64(elem_index)
				}
				*temp_data = append(*temp_data, data)
				elem_index++
			}
		}
	}
}

func lb_build_addr_compound_lit_assign_array(p *lbProcedure, temp_data []lbCompoundLitElemTempData) {
	for _, td := range temp_data {
		if td.Value.Value != 0 {
			if td.ElemLength > 0 {
				loop_data := lb_loop_start(p, isize(td.ElemLength), t_i32)
				{
					dst := td.Gep
					dst = lb_emit_ptr_offset(p, dst, loop_data.Idx)
					lb_emit_store(p, dst, td.Value)
				}
				lb_loop_end(p, loop_data)
			} else {
				lb_emit_store(p, td.Gep, td.Value)
			}
		}
	}
}

func lb_build_addr_compound_lit(p *lbProcedure, expr *Ast) lbAddr {
	cl := &expr.CompoundLit

	type_ := type_of_expr(expr)
	bt := base_type(type_)

	v := lb_add_local_generated(p, type_, true)

	TEMPORARY_ALLOCATOR_GUARD()

	var et *Type
	switch bt.Kind {
	case TypeArray:
		et = bt.Array.Elem
	case TypeEnumeratedArray:
		et = bt.EnumeratedArray.Elem
	case TypeSlice:
		et = bt.Slice.Elem
	case TypeBitSet:
		et = bt.BitSet.Elem
	case TypeSimdVector:
		et = bt.SimdVector.Elem
	case TypeMatrix:
		et = bt.Matrix.Elem
	case TypeFixedCapacityDynamicArray:
		et = bt.FixedCapacityDynamicArray.Elem
	}
	_ = et

	proc_name := String{}
	if p.Entity != nil {
		proc_name = p.Entity.Token.String
	}
	pos := ast_token(expr).Pos

	switch bt.Kind {
	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_expr.cpp", 5812,
			"Unknown CompoundLit type: %s", type_to_string(type_))

	case TypeBitField:
		{
			TEMPORARY_ALLOCATOR_GUARD()

			type FieldData struct {
				FieldType *Type
				BitOffset u64
				BitSize   u64
			}

			values := make([]lbValue, 0, len(cl.Elems))
			fields := make([]FieldData, 0, len(cl.Elems))

			for _, elem := range cl.Elems {
				fv := &elem.FieldValue
				interned := fv.Field.Ident.Interned
				sel := lookup_field(bt, interned, false)

				index := sel.Index.Data[0]
				f := bt.BitField.Fields[index]
				bit_offset := bt.BitField.BitOffsets[index]
				bit_size := bt.BitField.BitSizes[index]

				field_type := sel.Entity.Type
				field_expr := lb_build_expr(p, fv.Value)
				field_expr = lb_emit_conv(p, field_expr, field_type)
				values = append(values, field_expr)
				fields = append(fields, FieldData{field_type, u64(bit_offset), u64(bit_size)})
			}

			for i := 1; i < len(values); i++ {
				for j := i; j > 0 && fields[i].BitOffset < fields[j].BitOffset; j-- {
					values[j], values[j-1] = values[j-1], values[j]
					fields[j], fields[j-1] = fields[j-1], fields[j]
				}
			}

			any_fields_different_endian := false
			for _, f := range fields {
				if is_type_different_to_arch_endianness(f.FieldType) {
					any_fields_different_endian = true
					break
				}
			}

			if !any_fields_different_endian && len(fields) == len(bt.BitField.Fields) {
				backing_type := core_type(bt.BitField.BackingType)

				dst_byte_ptr := lb_emit_conv(p, v.Addr, t_u8_ptr)
				total_bit_size := u64(8 * type_size_of(bt))

				if is_type_integer(backing_type) {
					lit := lb_type(p.Module, backing_type)

					res := LLVMConstInt(lit, 0, false)

					for i := 0; i < len(fields); i++ {
						f := fields[i]

						mask := lb_const_low_bits_mask(lit, f.BitSize)

						elem := values[i].Value
						if lb_sizeof(lit) < lb_sizeof(LLVMTypeOf(elem)) {
							elem = LLVMBuildTrunc(p.Builder, elem, lit, "")
						} else {
							elem = LLVMBuildZExt(p.Builder, elem, lit, "")
						}
						elem = LLVMBuildAnd(p.Builder, elem, mask, "")

						elem = LLVMBuildShl(p.Builder, elem, LLVMConstInt(lit, f.BitOffset, false), "")

						res = LLVMBuildOr(p.Builder, res, elem, "")
					}

					LLVMBuildStore(p.Builder, res, v.Addr.Value)
				} else if is_type_array(backing_type) {
					array_count := backing_type.Array.Count
					lit := lb_type(p.Module, core_type(backing_type.Array.Elem))

					elems := make([]LLVMValueRef, array_count)
					for i := int64(0); i < array_count; i++ {
						elems[i] = LLVMConstInt(lit, 0, false)
					}

					elem_bit_size := u64(8 * type_size_of(backing_type.Array.Elem))
					curr_bit_offset := u64(0)
					for i := 0; i < len(fields); i++ {
						f := fields[i]
						val := values[i].Value
						vt := lb_type(p.Module, values[i].Type)
						for bits_to_set := f.BitSize; bits_to_set > 0; {
							elem_idx := curr_bit_offset / elem_bit_size
							elem_bit_offset := curr_bit_offset % elem_bit_size

							mask_width := bits_to_set
							if elem_bit_size-elem_bit_offset < mask_width {
								mask_width = elem_bit_size - elem_bit_offset
							}

							bits_to_set -= mask_width

							mask := lb_const_low_bits_mask(vt, mask_width)

							to_set := LLVMBuildAnd(p.Builder, val, mask, "")

							if elem_bit_offset != 0 {
								to_set = LLVMBuildShl(p.Builder, to_set, LLVMConstInt(vt, elem_bit_offset, false), "")
							}
							to_set = LLVMBuildTrunc(p.Builder, to_set, lit, "")

							if LLVMIsNull(elems[elem_idx]) {
								elems[elem_idx] = to_set
							} else {
								elems[elem_idx] = LLVMBuildOr(p.Builder, elems[elem_idx], to_set, "")
							}

							if mask_width != 0 {
								val = LLVMBuildLShr(p.Builder, val, LLVMConstInt(vt, mask_width, false), "")
							}
							curr_bit_offset += mask_width
						}

						if curr_bit_offset != f.BitOffset+f.BitSize {
							gb_assert_handler("Assertion Failure", nil, "", 0, "bit field offset mismatch")
						}
					}

					for i := int64(0); i < array_count; i++ {
						elem_ptr := LLVMBuildStructGEP2(p.Builder, lb_type(p.Module, backing_type), v.Addr.Value, unsigned(i), "")
						LLVMBuildStore(p.Builder, elems[i], elem_ptr)
					}
				} else {
					for i := 0; i < len(fields); i++ {
						f := fields[i]
						if (f.BitOffset & 7) == 0 {
							unpacked_bit_size := u64(8 * type_size_of(f.FieldType))
							byte_size := (f.BitSize + 7) / 8

							if f.BitOffset+unpacked_bit_size <= total_bit_size {
								byte_size = unpacked_bit_size / 8
							}
							dst := lb_emit_ptr_offset(p, dst_byte_ptr, lb_const_int(p.Module, t_int, int64(f.BitOffset/8)))
							src := lb_address_from_load_or_generate_local(p, values[i])
							lb_mem_copy_non_overlapping(p, dst, src, lb_const_int(p.Module, t_uintptr, byte_size))
						} else {
							dst := lb_addr_bit_field(v.Addr, f.FieldType, int64(f.BitOffset), int64(f.BitSize))
							lb_addr_store(p, dst, values[i])
						}
					}
				}
			} else {
				for i := 0; i < len(values); i++ {
					f := fields[i]
					dst := lb_addr_bit_field(v.Addr, f.FieldType, int64(f.BitOffset), int64(f.BitSize))
					lb_addr_store(p, dst, values[i])
				}
			}

			return v
		}

	case TypeStruct:
		lb_build_addr_struct_compound_lit_populate(p, expr, type_, v)

	case TypeMap:
		if len(cl.Elems) > 0 {
			err := lb_dynamic_map_reserve(p, v.Addr, 2*len(cl.Elems), pos)
			_ = err

			for _, elem := range cl.Elems {
				fv := &elem.FieldValue

				key := lb_build_expr(p, fv.Field)
				value := lb_build_expr(p, fv.Value)
				lb_internal_dynamic_map_set(p, v.Addr, type_, key, value, elem)
			}
		}

	case TypeArray:
		if len(cl.Elems) > 0 {
			lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			dst_ptr := lb_addr_get_ptr(p, v)
			for i := 0; i < len(temp_data); i++ {
				index := i32(temp_data[i].ElemIndex)
				temp_data[i].Gep = lb_emit_array_epi(p, dst_ptr, index)
			}

			lb_build_addr_compound_lit_assign_array(p, temp_data)
		}

	case TypeEnumeratedArray:
		if len(cl.Elems) > 0 {
			lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			dst_ptr := lb_addr_get_ptr(p, v)
			index_offset := exact_value_to_i64(*bt.EnumeratedArray.MinValue)
			for i := 0; i < len(temp_data); i++ {
				index := i32(temp_data[i].ElemIndex - index_offset)
				temp_data[i].Gep = lb_emit_array_epi(p, dst_ptr, index)
			}

			lb_build_addr_compound_lit_assign_array(p, temp_data)
		}

	case TypeSlice:
		if len(cl.Elems) > 0 {
			slice := lb_const_value(p.Module, type_, exact_value_compound(expr))

			data := lb_slice_elem(p, slice)

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			for i := 0; i < len(temp_data); i++ {
				temp_data[i].Gep = lb_emit_ptr_offset(p, data, lb_const_int(p.Module, t_int, temp_data[i].ElemIndex))
			}

			lb_build_addr_compound_lit_assign_array(p, temp_data)

			{
				count := lbValue{}
				count.Type = t_int

				len_index := lb_convert_struct_index(p.Module, type_, 1)
				if lb_is_const(slice) {
					indices := [1]unsigned{len_index}
					count.Value = llvm_const_extract_value(p.Module, slice.Value, indices[:], 1)
				} else {
					count.Value = LLVMBuildExtractValue(p.Builder, slice.Value, len_index, "")
				}
				lb_fill_slice(p, v, data, count)
			}
		}

	case TypeFixedCapacityDynamicArray:
		if len(cl.Elems) > 0 {
			lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			dst_ptr := lb_addr_get_ptr(p, v)
			for i := 0; i < len(temp_data); i++ {
				index := i32(temp_data[i].ElemIndex)
				temp_data[i].Gep = lb_emit_array_epi(p, dst_ptr, index)
			}

			lb_build_addr_compound_lit_assign_array(p, temp_data)
		}

	case TypeDynamicArray:
		if len(cl.Elems) == 0 {
			break
		}

		et := bt.DynamicArray.Elem
		size := lb_const_int(p.Module, t_int, type_size_of(et))
		align := lb_const_int(p.Module, t_int, type_align_of(et))

		item_count := max(int64(cl.MaxCount), int64(len(cl.Elems)))
		{
			args := make([]lbValue, 5)
			args[0] = lb_emit_conv(p, lb_addr_get_ptr(p, v), t_rawptr)
			args[1] = size
			args[2] = align
			args[3] = lb_const_int(p.Module, t_int, item_count)
			args[4] = lb_emit_source_code_location_as_global(p, proc_name, pos)
			lb_emit_runtime_call(p, "__dynamic_array_reserve", args)
		}

		items := lb_generate_local_array(p, et, item_count)

		temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))
		lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

		for i := 0; i < len(temp_data); i++ {
			temp_data[i].Gep = lb_emit_array_epi(p, items, temp_data[i].ElemIndex)
		}
		lb_build_addr_compound_lit_assign_array(p, temp_data)

		{
			args := make([]lbValue, 6)
			args[0] = lb_emit_conv(p, v.Addr, t_rawptr)
			args[1] = size
			args[2] = align
			args[3] = lb_emit_conv(p, items, t_rawptr)
			args[4] = lb_const_int(p.Module, t_int, item_count)
			args[5] = lb_emit_source_code_location_as_global(p, proc_name, pos)
			lb_emit_runtime_call(p, "__dynamic_array_append", args)
		}

	case TypeBasic:
		_ = bt
		if len(cl.Elems) > 0 {
			lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))
			field_names := [2]string{"data", "id"}
			field_types := [2]*Type{t_rawptr, t_typeid}

			for field_index := 0; field_index < len(cl.Elems); field_index++ {
				elem := cl.Elems[field_index]

				field_expr := lbValue{}
				index := isize(field_index)

				if elem.Kind == Ast_FieldValue {
					fv := &elem.FieldValue
					sel := lookup_field(bt, fv.Field.Ident.Interned, false)
					index = sel.Index.Data[0]
					elem = fv.Value
				} else {
					tav := type_and_value_of_expr(elem)
					_ = tav
					sel := lookup_field(bt, string_interner_insert(field_names[field_index]), false)
					index = sel.Index.Data[0]
				}

				field_expr = lb_build_expr(p, elem)

				ft := field_types[index]
				fv := lb_emit_conv(p, field_expr, ft)
				gep := lb_emit_struct_ep(p, lb_addr_get_ptr(p, v), i32(index))
				lb_emit_store(p, gep, fv)
			}
		}

	case TypeBitSet:
		sz := type_size_of(type_)
		if len(cl.Elems) > 0 && sz > 0 {
			lower := lb_const_value(p.Module, t_int, exact_value_i64(bt.BitSet.Lower))

			backing := bit_set_to_int(type_)
			if is_type_array(backing) {
				base_it := core_array_type(backing)
				bits_per_elem := 8 * type_size_of(base_it)
				_ = bits_per_elem
				one := lb_const_value(p.Module, t_i64, exact_value_i64(1))
				for _, elem := range cl.Elems {
					if elem.Kind == Ast_FieldValue {
						continue
					}
					expr := lb_build_expr(p, elem)
					e := lb_emit_conv(p, expr, t_i64)
					e = lb_emit_arith(p, Token_Sub, e, lower, t_i64)
					_ = one
				}
			} else {
				it := bit_set_to_int(bt)
				one := lb_const_value(p.Module, it, exact_value_i64(1))
				for _, elem := range cl.Elems {
					if elem.Kind == Ast_FieldValue {
						continue
					}

					expr := lb_build_expr(p, elem)

					e := lb_emit_conv(p, expr, it)
					e = lb_emit_arith(p, Token_Sub, e, lower, it)
					e = lb_emit_arith(p, Token_Shl, one, e, it)

					old_value := lb_emit_transmute(p, lb_addr_load(p, v), it)
					new_value := lb_emit_arith(p, Token_Or, old_value, e, it)
					new_value = lb_emit_transmute(p, new_value, type_)
					lb_addr_store(p, v, new_value)
				}
			}
		}

	case TypeMatrix:
		if len(cl.Elems) > 0 {
			lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			dst_ptr := lb_addr_get_ptr(p, v)
			for i := 0; i < len(temp_data); i++ {
				temp_data[i].Gep = lb_emit_array_epi(p, dst_ptr, temp_data[i].ElemIndex)
			}

			lb_build_addr_compound_lit_assign_array(p, temp_data)
		}

	case TypeSimdVector:
		if len(cl.Elems) > 0 {
			vector_value := lb_const_value(p.Module, type_, exact_value_compound(expr))
			defer func() { lb_addr_store(p, v, vector_value) }()

			temp_data := make([]lbCompoundLitElemTempData, 0, len(cl.Elems))

			lb_build_addr_compound_lit_populate(p, cl.Elems, &temp_data, type_)

			for _, td := range temp_data {
				if td.Value.Value != 0 {
					if td.ElemLength > 0 {
						for k := int64(0); k < td.ElemLength; k++ {
							index := lb_const_int(p.Module, t_u32, td.ElemIndex+k).Value
							vector_value.Value = LLVMBuildInsertElement(p.Builder, vector_value.Value, td.Value.Value, index, "")
						}
					} else {
						index := lb_const_int(p.Module, t_u32, td.ElemIndex).Value
						vector_value.Value = LLVMBuildInsertElement(p.Builder, vector_value.Value, td.Value.Value, index, "")
					}
				}
			}
		}
	}

	return v
}

func lb_build_addr_struct_compound_lit_populate(p *lbProcedure, expr *Ast, type_ *Type, v lbAddr) {
	cl := &expr.CompoundLit

	bt := base_type(type_)
	st := &bt.Struct

	if len(cl.Elems) == 0 {
		return
	}

	is_raw_union := st.IsRawUnion

	lb_addr_store(p, v, lb_const_value(p.Module, type_, exact_value_compound(expr)))
	comp_lit_ptr := lb_addr_get_ptr(p, v)

	if len(cl.Elems) > 0 && cl.Elems[0].Kind == Ast_FieldValue {
		for _, elem := range cl.Elems {
			field_expr := lbValue{}
			var field *Entity

			fv := &elem.FieldValue
			interned := fv.Field.Ident.Interned
			sel := lookup_field(bt, interned, false)

			elem = fv.Value
			if len(sel.Index.Data) > 1 {
				if lb_is_nested_possibly_constant(type_, sel, elem) {
					continue
				}
				field_expr = lb_build_expr(p, elem)
				field_expr = lb_emit_conv(p, field_expr, sel.Entity.Type)
				if sel.IsBitField {
					sub_sel := trim_selection(sel)
					trimmed_dst := lb_emit_deep_field_gep(p, comp_lit_ptr, sub_sel)
					bf := base_type(type_deref(trimmed_dst.Type))
					if is_type_pointer(bf) {
						trimmed_dst = lb_emit_load(p, trimmed_dst)
						bf = base_type(type_deref(trimmed_dst.Type))
					}

					idx := sel.Index.Data[len(sel.Index.Data)-1]
					dst := lb_addr_bit_field(trimmed_dst, bf.BitField.Fields[idx].Type, bf.BitField.BitOffsets[idx], bf.BitField.BitSizes[idx])
					lb_addr_store(p, dst, field_expr)
				} else {
					dst := lb_emit_deep_field_gep(p, comp_lit_ptr, sel)
					lb_emit_store(p, dst, field_expr)
				}
				continue
			}

			index := sel.Index.Data[0]

			field = st.Fields[index]
			ft := field.Type
			if !is_raw_union && !is_type_typeid(ft) && lb_is_elem_const(elem, ft) {
				continue
			}

			field_expr = lb_build_expr(p, elem)

			lb_build_struct_compound_lit_field_assignment(p, comp_lit_ptr, field, isize(index), field_expr, is_raw_union)
		}
		return
	}

	field_index := isize(0)

	for _, elem := range cl.Elems {
		if is_type_tuple(elem.Tav.Type) {
			tuple_field_expr := lb_build_expr(p, elem)

			tt := &tuple_field_expr.Type.Tuple
			for jj := 0; jj < len(tt.Variables); jj++ {
				index := field_index
				field_index++
				sel := lookup_field_from_index(bt, index)
				index = sel.Index.Data[0]

				field := st.Fields[index]
				field_expr := lb_emit_struct_ev(p, tuple_field_expr, i32(jj))

				lb_build_struct_compound_lit_field_assignment(p, comp_lit_ptr, field, index, field_expr, is_raw_union)
			}
			continue
		}

		index := field_index
		field_index++
		sel := lookup_field_from_index(bt, index)
		index = sel.Index.Data[0]

		field := st.Fields[index]
		ft := field.Type
		if !is_type_typeid(ft) && lb_is_elem_const(elem, ft) {
			continue
		}

		field_expr := lb_build_expr(p, elem)

		lb_build_struct_compound_lit_field_assignment(p, comp_lit_ptr, field, index, field_expr, is_raw_union)
	}
}

func lb_build_struct_compound_lit_field_assignment(p *lbProcedure, comp_lit_ptr lbValue, field_entity *Entity, index isize, field_expr lbValue, is_raw_union bool) {
	ft := field_entity.Type

	var gep lbValue
	if is_raw_union {
		gep = lb_emit_conv(p, comp_lit_ptr, alloc_type_pointer(ft))
	} else {
		gep = lb_emit_struct_ep(p, comp_lit_ptr, i32(index))
	}

	fet := field_expr.Type

	if is_type_union(ft) && !are_types_identical(fet, ft) && !is_type_untyped(fet) {
		if union_is_variant_of(ft, fet) {
			fv := lb_emit_conv(p, field_expr, ft)
			lb_emit_store(p, gep, fv)
		} else {
			fv := lb_emit_conv(p, field_expr, ft)
			lb_emit_store(p, gep, fv)
		}
	} else {
		fv := lb_emit_conv(p, field_expr, ft)
		lb_emit_store(p, gep, fv)
	}
}
