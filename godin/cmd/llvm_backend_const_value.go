package cmd

import (
	"fmt"
	"math"
	"unicode/utf8"
	"unsafe"
)

func lb_const_value(m *lbModule, typ *Type, value ExactValue, cc ...lbConstContext) lbValue {
	ctx := LB_CONST_CONTEXT_DEFAULT
	if len(cc) > 0 {
		ctx = cc[0]
	}
	return lb_const_value_internal(m, typ, value, ctx, nil)
}

func lb_const_value_internal(m *lbModule, type_ *Type, value ExactValue, cc lbConstContext, value_type *Type) lbValue {
	if cc.AllowLocal {
		cc.IsRodata = false
	}
	ctx := m.Ctx
	type_ = default_type(type_)
	original_type := type_
	res := lbValue{}
	res.Type = original_type
	type_ = core_type(type_)
	value = convert_exact_value_for_type(value, type_)
	is_local := cc.AllowLocal && m.CurrProcedure != nil

	if is_type_union(type_) && is_type_union_constantable(type_) {
		bt := base_type(type_)
		if len(bt.Union.Variants) == 0 {
			return lb_const_nil(m, original_type)
		} else if len(bt.Union.Variants) == 1 {
			if value.Kind == ExactValue_Compound {
				cl := value.ValueCompound.CompoundLit
				if len(cl.Elems) == 0 {
					if cl.Type == nil {
						return lb_const_nil(m, original_type)
					}
					if are_types_identical(type_of_expr(cl.Type), original_type) {
						return lb_const_nil(m, original_type)
					}
				}
			}
			if value_type == t_untyped_nil {
				return lb_const_nil(m, original_type)
			}
			t := bt.Union.Variants[0]
			cv := lb_const_value_internal(m, t, value, cc, value_type)
			llvm_type := lb_type(m, original_type)
			if is_type_union_maybe_pointer(type_) {
				values := [1]LLVMValueRef{cv.Value}
				res.Value = llvm_const_named_struct_internal(m, llvm_type, values[:], 1)
				res.Type = original_type
				return res
			} else {
				tag_value := uint64(1)
				if bt.Union.Kind == UnionType_NoNil {
					tag_value = 0
				}
				tag := LLVMConstInt(LLVMStructGetTypeAtIndex(llvm_type, 1), tag_value, false)
				var padding LLVMValueRef
				value_count := isize(2)
				if LLVMCountStructElementTypes(llvm_type) > 2 {
					value_count = 3
					padding = LLVMConstNull(LLVMStructGetTypeAtIndex(llvm_type, 2))
				}
				values := [3]LLVMValueRef{cv.Value, tag, padding}
				res.Value = llvm_const_named_struct_internal(m, llvm_type, values[:], value_count)
				res.Type = original_type
				return res
			}
		} else {
			if value_type == nil {
				if value.Kind == ExactValue_Compound {
					cl := value.ValueCompound.CompoundLit
					if len(cl.Elems) == 0 {
						return lb_const_nil(m, original_type)
					}
				} else if value.Kind == ExactValue_Invalid {
					return lb_const_nil(m, original_type)
				}
			} else if value_type == t_untyped_nil {
				return lb_const_nil(m, original_type)
			}
			block_size := bt.Union.VariantBlockSize
			if are_types_identical(value_type, original_type) {
				if value.Kind == ExactValue_Compound {
					cl := value.ValueCompound.CompoundLit
					if len(cl.Elems) == 0 {
						return lb_const_nil(m, original_type)
					}
				} else if value.Kind == ExactValue_Invalid {
					return lb_const_nil(m, original_type)
				}
			}
			cv := lb_const_value_internal(m, value_type, value, cc, value_type)
			variant_type := cv.Type
			var values [4]LLVMValueRef
			value_count := 0
			values[value_count] = cv.Value
			value_count++
			if type_size_of(variant_type) != block_size {
				padding_type := lb_type_padding_filler(m, block_size-type_size_of(variant_type), 1)
				values[value_count] = LLVMConstNull(padding_type)
				value_count++
			}
			tag_type := union_tag_type(bt)
			llvm_tag_type := lb_type(m, tag_type)
			tag_index := union_variant_index_checked(bt, variant_type)
			values[value_count] = LLVMConstInt(llvm_tag_type, uint64(tag_index), false)
			value_count++
			used_size := block_size + type_size_of(tag_type)
			union_size := type_size_of(bt)
			padding_size := union_size - used_size
			if padding_size > 0 {
				padding_type := lb_type_padding_filler(m, padding_size, 1)
				values[value_count] = LLVMConstNull(padding_type)
				value_count++
			}
			res.Value = LLVMConstStructInContext(m.Ctx, values[:], uint(value_count), true)
			return res
		}
	}

	if value.Kind == ExactValue_Procedure {
		res_proc := lbValue{}
		for {
			expr := unparen_expr(value.ValueProcedure)
			if expr.Kind == Ast_ProcLit {
				res_proc = lb_generate_anonymous_proc_lit(m, "_proclit", expr)
				break
			}
			e := entity_from_expr(expr)
			if e.Kind != Entity_Constant {
				res_proc = lb_find_procedure_value_from_entity(m, e)
				break
			}
			value = e.Constant.Value
		}
		if res_proc.Value == 0 {
			return lb_const_nil(m, original_type)
		}
		if LLVMGetIntrinsicID(res_proc.Value) == 0 {
			res_proc.Value = LLVMConstPointerCast(res_proc.Value, lb_type(m, res_proc.Type))
		}
		return res_proc
	}

	if value.Kind == ExactValue_Invalid {
		return lb_const_nil(m, original_type)
	}

	if value.Kind == ExactValue_Typeid {
		return lb_typeid(m, value.ValueTypeid)
	}

	if value.Kind == ExactValue_Compound {
		cl := value.ValueCompound.CompoundLit
		if len(cl.Elems) == 0 {
			return lb_const_nil(m, original_type)
		}
	}

	if is_type_slice(type_) {
		if value.Kind == ExactValue_String {
			res.Value = lb_find_or_add_entity_string_byte_slice_with_type(m, goStr(value.ValueString), original_type).Value
			return res
		} else if value.Kind == ExactValue_String16 {
			s16_str := string16_to_string(temporary_allocator(), value.ValueString16)
			res.Value = lb_find_or_add_entity_string16_slice_with_type(m, goStr(s16_str), original_type).Value
			return res
		} else {
			cl := value.ValueCompound.CompoundLit
			count := isize(len(cl.Elems))
			if count == 0 {
				return lb_const_nil(m, type_)
			}
			if cl.MaxCount > int64(count) {
				count = isize(cl.MaxCount)
			}
			elem := base_type(type_).Slice.Elem
			t := alloc_type_array(elem, int64(count))
			backing_array := lb_const_value_internal(m, t, value, cc, nil)
			var array_data LLVMValueRef
			if is_local {
				p := m.CurrProcedure
				llvm_type := lb_type(m, t)
				alignment := uint(max64(type_align_of(t), 16))
				array_data = llvm_alloca(p, llvm_type, int64(alignment))
				LLVMBuildStore(p.Builder, backing_array.Value, array_data)
				array_data = LLVMBuildPointerCast(p.Builder, array_data, LLVMPointerType(llvm_type, 0), "")
				{
					indices := [2]LLVMValueRef{llvm_zero(m), llvm_zero(m)}
					ptr := LLVMBuildInBoundsGEP2(p.Builder, llvm_type, array_data, indices[:], 2, "")
					len_val := LLVMConstInt(lb_type(m, t_int), uint64(count), true)
					slice := lb_add_local_generated(p, original_type, false)
					m.ExactValueCompoundLiteralAddrMap[value.ValueCompound] = slice
					lb_fill_slice(p, slice, lbValue{Value: ptr, Type: alloc_type_pointer(elem)}, lbValue{Value: len_val, Type: t_int})
					return lb_addr_load(p, slice)
				}
			} else {
				id := m.GlobalArrayIndex.Add(1)
				name := fmt.Sprintf("csba$%s$%x", m.ModuleName, id)
				e := alloc_entity_constant(nil, make_token_ident(name), t, value)
				array_data = LLVMAddGlobal(m.Mod, LLVMTypeOf(backing_array.Value), name)
				LLVMSetInitializer(array_data, backing_array.Value)
				if cc.LinkSection.Len > 0 {
					LLVMSetSection(array_data, alloc_cstring(permanent_allocator(), goStr(cc.LinkSection)))
				}
				if cc.IsRodata {
					LLVMSetGlobalConstant(array_data, true)
				}
				g := lbValue{}
				g.Value = LLVMConstPointerCast(array_data, LLVMPointerType(lb_type(m, t), 0))
				g.Type = t
				m.Values[unsafe.Pointer(e)] = g
				lb_add_member(m, name, g)
				{
					ptr := g.Value
					len_val := LLVMConstInt(lb_type(m, t_int), uint64(count), true)
					values := [2]LLVMValueRef{ptr, len_val}
					res.Value = llvm_const_named_struct(m, original_type, values[:], 2)
					return res
				}
			}
		}
	} else if is_type_rune_array(type_) && value.Kind == ExactValue_String && !is_type_u8(core_array_type(type_)) {
		count := type_.Array.Count
		elem := type_.Array.Elem
		et := lb_type(m, elem)
		s := goStr(value.ValueString)
		elems := make([]LLVMValueRef, count)
		offset := 0
		for i := int64(0); i < count && offset < len(s); i++ {
			r, width := utf8.DecodeRuneInString(s[offset:])
			offset += width
			elems[i] = LLVMConstInt(et, uint64(r), true)
		}
		res.Value = llvm_const_array(m, et, elems, isize(count))
		return res
	} else if is_type_u8_array(type_) && value.Kind == ExactValue_String {
		s := goStr(value.ValueString)
		if int64(len(s)) != type_.Array.Count {
			return lb_const_nil(m, original_type)
		}
		data := LLVMConstStringInContext(ctx, s, uint(len(s)), true)
		res.Value = data
		return res
	} else if is_type_array(type_) &&
		value.Kind != ExactValue_Invalid &&
		value.Kind != ExactValue_Compound {
		lb_const_array_spread(m, cc, type_, value, &res)
		return res
	} else if is_type_matrix(type_) &&
		value.Kind != ExactValue_Invalid &&
		value.Kind != ExactValue_Compound {
		row := type_.Matrix.RowCount
		column := type_.Matrix.ColumnCount
		elem := type_.Matrix.Elem
		single_elem := lb_const_value_internal(m, elem, value, cc, nil)
		single_elem.Value = llvm_const_cast(m, single_elem.Value, lb_type(m, elem), nil)
		total_elem_count := matrixTypeTotalInternalElems(type_)
		elems := make([]LLVMValueRef, total_elem_count)
		for i := int64(0); i < row; i++ {
			elems[matrixRowMajorIndexToOffset(type_, i*column+i)] = single_elem.Value
		}
		for i := int64(0); i < total_elem_count; i++ {
			if elems[i] == 0 {
				elems[i] = LLVMConstNull(lb_type(m, elem))
			}
		}
		res.Value = LLVMConstArray(lb_type(m, elem), elems, uint(total_elem_count))
		return res
	} else if is_type_simd_vector(type_) &&
		value.Kind != ExactValue_Invalid &&
		value.Kind != ExactValue_Compound {
		count := type_.SimdVector.Count
		elem := type_.SimdVector.Elem
		single_elem := lb_const_value_internal(m, elem, value, cc, nil)
		single_elem.Value = llvm_const_cast(m, single_elem.Value, lb_type(m, elem), nil)
		elems := make([]LLVMValueRef, count)
		for i := int64(0); i < count; i++ {
			elems[i] = single_elem.Value
		}
		res.Value = LLVMConstVector(elems, uint(count))
		return res
	}

	switch value.Kind {
	case ExactValue_Invalid:
		res.Value = LLVMConstNull(lb_type(m, original_type))
		return res

	case ExactValue_Bool:
		{
			v := uint64(0)
			if value.ValueBool {
				v = 1
			}
			res.Value = LLVMConstInt(lb_type(m, original_type), v, false)
			return res
		}

	case ExactValue_String:
		{
			custom_link_section := cc.LinkSection.Len > 0
			var ptr LLVMValueRef
			res = lbValue{}
			res.Type = default_type(original_type)
			str := goStr(value.ValueString)
			len_val := isize(len(str))
			if is_type_string16(res.Type) || is_type_cstring16(res.Type) {
				s16 := string_to_string16(temporary_allocator(), value.ValueString)
				len_val = s16.Len
				s16_str := string16_to_string(temporary_allocator(), s16)
				ptr = lb_find_or_add_entity_string16_ptr(m, goStr(s16_str), custom_link_section)
			} else {
				ptr = lb_find_or_add_entity_string_ptr(m, str, custom_link_section)
			}
			if custom_link_section {
				LLVMSetSection(ptr, alloc_cstring(permanent_allocator(), goStr(cc.LinkSection)))
			}
			if is_type_cstring(res.Type) || is_type_cstring16(res.Type) {
				res.Value = ptr
			} else {
				if len_val == 0 {
					if is_type_string16(res.Type) {
						ptr = LLVMConstNull(lb_type(m, t_u16_ptr))
					} else {
						ptr = LLVMConstNull(lb_type(m, t_u8_ptr))
					}
				}
				str_len := LLVMConstInt(lb_type(m, t_int), uint64(len_val), true)
				if is_type_string16(res.Type) {
					res.Value = llvm_const_string16_internal(m, original_type, ptr, str_len)
				} else {
					res.Value = llvm_const_string_internal(m, original_type, ptr, str_len)
				}
			}
			return res
		}

	case ExactValue_String16:
		{
			custom_link_section := cc.LinkSection.Len > 0
			s8 := string16_to_string(temporary_allocator(), value.ValueString16)
			str := goStr(s8)
			ptr := lb_find_or_add_entity_string16_ptr(m, str, custom_link_section)
			res = lbValue{}
			res.Type = default_type(original_type)
			if custom_link_section {
				LLVMSetSection(ptr, alloc_cstring(permanent_allocator(), goStr(cc.LinkSection)))
			}
			if is_type_cstring16(res.Type) {
				res.Value = ptr
			} else {
				if value.ValueString16.Len == 0 {
					ptr = LLVMConstNull(lb_type(m, t_u8_ptr))
				}
				str_len := LLVMConstInt(lb_type(m, t_int), uint64(value.ValueString16.Len), true)
				res.Value = llvm_const_string16_internal(m, original_type, ptr, str_len)
			}
			return res
		}

	case ExactValue_Integer:
		{
			if is_type_pointer(type_) || is_type_multi_pointer(type_) || is_type_proc(type_) {
				t := lb_type(m, original_type)
				i := lb_big_int_to_llvm(m, t_uintptr, &value.ValueInteger)
				res.Value = LLVMConstIntToPtr(i, t)
			} else {
				res.Value = lb_big_int_to_llvm(m, original_type, &value.ValueInteger)
			}
			return res
		}

	case ExactValue_Float:
		{
			if is_type_different_to_arch_endianness(type_) {
				bk := type_.Basic.Kind
				if bk == Basic_f32le || bk == Basic_f32be {
					f := float32(value.ValueFloat)
					u := math.Float32bits(f)
					u = gb_endian_swap32(u)
					res.Value = LLVMConstReal(lb_type(m, original_type), float64(math.Float32frombits(u)))
				} else if bk == Basic_f16le || bk == Basic_f16be {
					f := float32(value.ValueFloat)
					u := f32_to_f16(f)
					u = gb_endian_swap16(u)
					res.Value = LLVMConstReal(lb_type(m, original_type), float64(f16_to_f32(u)))
				} else {
					u := math.Float64bits(value.ValueFloat)
					u = gb_endian_swap64(u)
					res.Value = LLVMConstReal(lb_type(m, original_type), math.Float64frombits(u))
				}
			} else {
				res.Value = LLVMConstReal(lb_type(m, original_type), value.ValueFloat)
			}
			return res
		}

	case ExactValue_Complex:
		{
			var values [2]LLVMValueRef
			switch 8 * type_size_of(type_) {
			case 32:
				values[0] = lb_const_f16(m, float32(value.ValueComplex.Real))
				values[1] = lb_const_f16(m, float32(value.ValueComplex.Imag))
			case 64:
				values[0] = lb_const_f32(m, float32(value.ValueComplex.Real))
				values[1] = lb_const_f32(m, float32(value.ValueComplex.Imag))
			case 128:
				values[0] = LLVMConstReal(lb_type(m, t_f64), value.ValueComplex.Real)
				values[1] = LLVMConstReal(lb_type(m, t_f64), value.ValueComplex.Imag)
			}
			res.Value = llvm_const_named_struct(m, original_type, values[:], 2)
			return res
		}

	case ExactValue_Quaternion:
		{
			var values [4]LLVMValueRef
			switch 8 * type_size_of(type_) {
			case 64:
				values[3] = lb_const_f16(m, float32(value.ValueQuaternion.Real))
				values[0] = lb_const_f16(m, float32(value.ValueQuaternion.Imag))
				values[1] = lb_const_f16(m, float32(value.ValueQuaternion.Jmag))
				values[2] = lb_const_f16(m, float32(value.ValueQuaternion.Kmag))
			case 128:
				values[3] = lb_const_f32(m, float32(value.ValueQuaternion.Real))
				values[0] = lb_const_f32(m, float32(value.ValueQuaternion.Imag))
				values[1] = lb_const_f32(m, float32(value.ValueQuaternion.Jmag))
				values[2] = lb_const_f32(m, float32(value.ValueQuaternion.Kmag))
			case 256:
				values[3] = LLVMConstReal(lb_type(m, t_f64), value.ValueQuaternion.Real)
				values[0] = LLVMConstReal(lb_type(m, t_f64), value.ValueQuaternion.Imag)
				values[1] = LLVMConstReal(lb_type(m, t_f64), value.ValueQuaternion.Jmag)
				values[2] = LLVMConstReal(lb_type(m, t_f64), value.ValueQuaternion.Kmag)
			}
			res.Value = llvm_const_named_struct(m, original_type, values[:], 4)
			return res
		}

	case ExactValue_Pointer:
		{
			res.Value = LLVMConstIntToPtr(LLVMConstInt(lb_type(m, t_uintptr), uint64(value.ValuePointer), false), lb_type(m, original_type))
			return res
		}

	case ExactValue_Compound:
		{
			if is_type_slice(type_) {
				return lb_const_value_internal(m, type_, value, cc, nil)
			} else if is_type_soa_struct(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.Struct.SoaElem
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 || !elem_type_can_be_constant(elem_type) {
					return lb_const_nil(m, original_type)
				}
				if cl.Elems[0].Kind == Ast_FieldValue {
					elem_count = isize(type_.Struct.SoaCount)
					aos_values := make([]LLVMValueRef, elem_count)
					value_index := 0
					for i := int64(0); i < int64(elem_count); i++ {
						found := false
						for j := isize(0); j < isize(len(cl.Elems)); j++ {
							elem := cl.Elems[j]
							fv := elem.FieldValue
							if is_ast_range(fv.Field) {
								ie := fv.Field.BinaryExpr
								lo_tav := ie.Left.TAV
								hi_tav := ie.Right.TAV
								op := ie.Op.Kind
								lo := exact_value_to_i64(lo_tav.Value)
								hi := exact_value_to_i64(hi_tav.Value)
								if op != Token_RangeHalf {
									hi += 1
								}
								if lo == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									for k := lo; k < hi; k++ {
										aos_values[value_index] = val
										value_index++
									}
									found = true
									i += (hi - lo - 1)
									break
								}
							} else {
								index_tav := fv.Field.TAV
								index := exact_value_to_i64(index_tav.Value)
								if index == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									aos_values[value_index] = val
									value_index++
									found = true
									break
								}
							}
						}
						if !found {
							aos_values[value_index] = 0
							value_index++
						}
					}
					field_count := isize(len(type_.Struct.Fields))
					soa_values := make([]LLVMValueRef, field_count)
					for i := isize(0); i < field_count; i++ {
						values := make([]LLVMValueRef, elem_count)
						f := type_.Struct.Fields[i]
						array_type := f.Type
						field_type := array_type.Array.Elem
						for j := isize(0); j < elem_count; j++ {
							v := aos_values[j]
							if v != 0 {
								values[j] = llvm_const_extract_value(m, v, []uint{uint(i)}, 1)
							} else {
								values[j] = LLVMConstNull(lb_type(m, field_type))
							}
						}
						soa_values[i] = lb_build_constant_array_values(m, array_type, field_type, elem_count, values, cc)
					}
					res.Value = llvm_const_named_struct(m, type_, soa_values, field_count)
					return res
				} else {
					elem_count = isize(type_.Struct.SoaCount)
					aos_values := make([]LLVMValueRef, elem_count)
					cl_elem_count := isize(len(cl.Elems))
					for i := isize(0); i < cl_elem_count; i++ {
						tav := cl.Elems[i].TAV
						aos_values[i] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
					}
					for i := cl_elem_count; i < elem_count; i++ {
						aos_values[i] = 0
					}
					field_count := isize(len(type_.Struct.Fields))
					soa_values := make([]LLVMValueRef, field_count)
					for i := isize(0); i < field_count; i++ {
						values := make([]LLVMValueRef, elem_count)
						f := type_.Struct.Fields[i]
						array_type := f.Type
						field_type := array_type.Array.Elem
						for j := isize(0); j < elem_count; j++ {
							v := aos_values[j]
							if v != 0 {
								values[j] = llvm_const_extract_value(m, v, []uint{uint(i)}, 1)
							} else {
								values[j] = LLVMConstNull(lb_type(m, field_type))
							}
						}
						soa_values[i] = lb_build_constant_array_values(m, array_type, field_type, elem_count, values, cc)
					}
					res.Value = llvm_const_named_struct(m, type_, soa_values, field_count)
					return res
				}
			} else if is_type_array(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.Array.Elem
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 || !elem_type_can_be_constant(elem_type) {
					return lb_const_nil(m, original_type)
				}
				if cl.Elems[0].Kind == Ast_FieldValue {
					values := make([]LLVMValueRef, type_.Array.Count)
					value_index := 0
					for i := int64(0); i < type_.Array.Count; i++ {
						found := false
						for j := isize(0); j < elem_count; j++ {
							elem := cl.Elems[j]
							fv := elem.FieldValue
							if is_ast_range(fv.Field) {
								ie := fv.Field.BinaryExpr
								lo_tav := ie.Left.TAV
								hi_tav := ie.Right.TAV
								op := ie.Op.Kind
								lo := exact_value_to_i64(lo_tav.Value)
								hi := exact_value_to_i64(hi_tav.Value)
								if op != Token_RangeHalf {
									hi += 1
								}
								if lo == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									for k := lo; k < hi; k++ {
										values[value_index] = val
										value_index++
									}
									found = true
									i += (hi - lo - 1)
									break
								}
							} else {
								index_tav := fv.Field.TAV
								index := exact_value_to_i64(index_tav.Value)
								if index == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									values[value_index] = val
									value_index++
									found = true
									break
								}
							}
						}
						if !found {
							values[value_index] = LLVMConstNull(lb_type(m, elem_type))
							value_index++
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(type_.Array.Count), values, cc)
					return res
				} else if are_types_identical(value.ValueCompound.TAV.Type, elem_type) {
					values := make([]LLVMValueRef, type_.Array.Count)
					for i := int64(0); i < type_.Array.Count; i++ {
						values[i] = lb_const_value_internal(m, elem_type, value, cc, elem_type).Value
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(type_.Array.Count), values, cc)
					return res
				} else {
					values := make([]LLVMValueRef, type_.Array.Count)
					elem_index := 0
					for i := isize(0); i < elem_count; i++ {
						tav := cl.Elems[i].TAV
						if is_type_tuple(tav.Type) {
							elem_index += len(base_type(tav.Type).Tuple.Variables)
						} else {
							values[elem_index] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
							elem_index++
						}
					}
					for i := int64(0); i < type_.Array.Count; i++ {
						if values[i] == 0 {
							values[i] = LLVMConstNull(lb_type(m, elem_type))
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(type_.Array.Count), values, cc)
					return res
				}
			} else if is_type_enumerated_array(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.EnumeratedArray.Elem
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 || !elem_type_can_be_constant(elem_type) {
					return lb_const_nil(m, original_type)
				}
				if cl.Elems[0].Kind == Ast_FieldValue {
					values := make([]LLVMValueRef, type_.EnumeratedArray.Count)
					value_index := 0
					total_lo := exact_value_to_i64(*type_.EnumeratedArray.MinValue)
					total_hi := exact_value_to_i64(*type_.EnumeratedArray.MaxValue)
					for i := total_lo; i <= total_hi; i++ {
						found := false
						for j := isize(0); j < elem_count; j++ {
							elem := cl.Elems[j]
							fv := elem.FieldValue
							if is_ast_range(fv.Field) {
								ie := fv.Field.BinaryExpr
								lo_tav := ie.Left.TAV
								hi_tav := ie.Right.TAV
								op := ie.Op.Kind
								lo := exact_value_to_i64(lo_tav.Value)
								hi := exact_value_to_i64(hi_tav.Value)
								if op != Token_RangeHalf {
									hi += 1
								}
								if lo == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									for k := lo; k < hi; k++ {
										values[value_index] = val
										value_index++
									}
									found = true
									i += (hi - lo - 1)
									break
								}
							} else {
								index_tav := fv.Field.TAV
								index := exact_value_to_i64(index_tav.Value)
								if index == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									values[value_index] = val
									value_index++
									found = true
									break
								}
							}
						}
						if !found {
							values[value_index] = LLVMConstNull(lb_type(m, elem_type))
							value_index++
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(type_.EnumeratedArray.Count), values, cc)
					return res
				} else {
					values := make([]LLVMValueRef, type_.EnumeratedArray.Count)
					elem_index := 0
					for i := isize(0); i < elem_count; i++ {
						tav := cl.Elems[i].TAV
						if is_type_tuple(tav.Type) {
							elem_index += len(base_type(tav.Type).Tuple.Variables)
						} else {
							values[elem_index] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
							elem_index++
						}
					}
					for i := int64(0); i < type_.EnumeratedArray.Count; i++ {
						if values[i] == 0 {
							values[i] = LLVMConstNull(lb_type(m, elem_type))
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(type_.EnumeratedArray.Count), values, cc)
					return res
				}
			} else if is_type_fixed_capacity_dynamic_array(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.FixedCapacityDynamicArray.Elem
				capacity := type_.FixedCapacityDynamicArray.Capacity
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 || !elem_type_can_be_constant(elem_type) {
					return lb_const_nil(m, original_type)
				}
				if cl.Elems[0].Kind == Ast_FieldValue {
					values := make([]LLVMValueRef, capacity)
					max_index := int64(-1)
					value_index := 0
					for i := int64(0); i < capacity; i++ {
						found := false
						for j := isize(0); j < elem_count; j++ {
							elem := cl.Elems[j]
							fv := elem.FieldValue
							if is_ast_range(fv.Field) {
								ie := fv.Field.BinaryExpr
								lo_tav := ie.Left.TAV
								hi_tav := ie.Right.TAV
								op := ie.Op.Kind
								lo := exact_value_to_i64(lo_tav.Value)
								hi := exact_value_to_i64(hi_tav.Value)
								if op != Token_RangeHalf {
									hi += 1
								}
								if hi-1 > max_index {
									max_index = hi - 1
								}
								if lo == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									for k := lo; k < hi; k++ {
										values[value_index] = val
										value_index++
									}
									found = true
									i += (hi - lo - 1)
									break
								}
							} else {
								index_tav := fv.Field.TAV
								index := exact_value_to_i64(index_tav.Value)
								if index > max_index {
									max_index = index
								}
								if index == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									values[value_index] = val
									value_index++
									found = true
									break
								}
							}
						}
						if !found {
							values[value_index] = LLVMConstNull(lb_type(m, elem_type))
							value_index++
						}
					}
					count := max_index + 1
					res.Value = lb_fill_fixed_capacity_dynamic_array(m, count, original_type, values, cc)
					return res
				} else if are_types_identical(value.ValueCompound.TAV.Type, elem_type) {
					values := make([]LLVMValueRef, capacity)
					for i := int64(0); i < capacity; i++ {
						values[i] = lb_const_value_internal(m, elem_type, value, cc, elem_type).Value
					}
					res.Value = lb_fill_fixed_capacity_dynamic_array(m, capacity, original_type, values, cc)
					return res
				} else {
					values := make([]LLVMValueRef, capacity)
					elem_index := 0
					for i := isize(0); i < elem_count; i++ {
						tav := cl.Elems[i].TAV
						if is_type_tuple(tav.Type) {
							elem_index += len(base_type(tav.Type).Tuple.Variables)
						} else {
							values[elem_index] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
							elem_index++
						}
					}
					for i := int64(0); i < capacity; i++ {
						if values[i] == 0 {
							values[i] = LLVMConstNull(lb_type(m, elem_type))
						}
					}
					res.Value = lb_fill_fixed_capacity_dynamic_array(m, int64(elem_index), original_type, values, cc)
					return res
				}
			} else if is_type_simd_vector(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.SimdVector.Elem
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 {
					return lb_const_nil(m, original_type)
				}
				total_elem_count := isize(type_.SimdVector.Count)
				values := make([]LLVMValueRef, total_elem_count)
				if cl.Elems[0].Kind == Ast_FieldValue {
					value_index := 0
					for i := int64(0); i < int64(total_elem_count); i++ {
						found := false
						for j := isize(0); j < elem_count; j++ {
							elem := cl.Elems[j]
							fv := elem.FieldValue
							if is_ast_range(fv.Field) {
								ie := fv.Field.BinaryExpr
								lo_tav := ie.Left.TAV
								hi_tav := ie.Right.TAV
								op := ie.Op.Kind
								lo := exact_value_to_i64(lo_tav.Value)
								hi := exact_value_to_i64(hi_tav.Value)
								if op != Token_RangeHalf {
									hi += 1
								}
								if lo == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									for k := lo; k < hi; k++ {
										values[value_index] = val
										value_index++
									}
									found = true
									i += (hi - lo - 1)
									break
								}
							} else {
								index_tav := fv.Field.TAV
								index := exact_value_to_i64(index_tav.Value)
								if index == i {
									tav := fv.Value.TAV
									val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
									values[value_index] = val
									value_index++
									found = true
									break
								}
							}
						}
						if !found {
							values[value_index] = LLVMConstNull(lb_type(m, elem_type))
							value_index++
						}
					}
					res.Value = LLVMConstVector(values, uint(total_elem_count))
					return res
				} else {
					for i := isize(0); i < elem_count; i++ {
						tav := cl.Elems[i].TAV
						values[i] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
					}
					et := lb_type(m, elem_type)
					for i := elem_count; i < total_elem_count; i++ {
						values[i] = LLVMConstNull(et)
					}
					for i := 0; i < total_elem_count; i++ {
						values[i] = llvm_const_cast(m, values[i], et, nil)
					}
					res.Value = LLVMConstVector(values, uint(total_elem_count))
					return res
				}
			} else if is_type_struct(type_) {
				cl := value.ValueCompound.CompoundLit
				if len(cl.Elems) == 0 {
					return lb_const_nil(m, original_type)
				}
				if is_type_raw_union(type_) {
					if is_type_raw_union_constantable(type_) {
						fv := cl.Elems[0].FieldValue
						f := entity_of_node(fv.Field)
						tav := fv.Value.TAV
						if tav.Value.Kind != ExactValue_Invalid {
							val := lb_const_value_internal(m, f.Type, tav.Value, cc, f.Type)
							var values [2]LLVMValueRef
							value_count := 0
							values[value_count] = val.Value
							value_count++
							union_alignment := type_align_of(type_)
							value_alignment := type_align_of(f.Type)
							alignment := min64(value_alignment, union_alignment)
							if alignment < 1 {
								alignment = 1
							}
							union_size := type_size_of(type_)
							value_size := lb_sizeof(LLVMTypeOf(val.Value))
							padding_size := union_size - value_size
							if padding_size > 0 {
								padding_type := lb_type_padding_filler(m, padding_size, alignment)
								values[value_count] = LLVMConstNull(padding_type)
								value_count++
							}
							res.Value = LLVMConstStructInContext(m.Ctx, values[:], uint(value_count), padding_size > 0)
							res.Type = original_type
							return res
						}
					}
					return lb_const_nil(m, original_type)
				}
				struct_type := lb_type(m, original_type)
				field_remapping := lb_get_struct_remapping(m, type_)
				value_count := int(LLVMCountStructElementTypes(struct_type))
				values := make([]LLVMValueRef, value_count)
				visited := make([]bool, value_count)
				if cl.Elems[0].Kind == Ast_FieldValue {
					elem_count := isize(len(cl.Elems))
					for i := isize(0); i < elem_count; i++ {
						fv := cl.Elems[i].FieldValue
						name := fv.Field.Ident.Token.String
						interned := fv.Field.Ident.Interned
						tav := fv.Value.TAV
						sel := lookup_field(type_, interned, false)
						f := type_.Struct.Fields[sel.Index[0]]
						index := field_remapping[f.Variable.FieldIndex]
						if elem_type_can_be_constant(f.Type) {
							if len(sel.Index) == 1 {
								val := lb_const_value_internal(m, f.Type, tav.Value, cc, tav.Type)
								values[index] = val.Value
								visited[index] = true
							} else {
								if !visited[index] {
									new_cc := cc
									new_cc.AllowLocal = false
									values[index] = lb_const_value_internal(m, f.Type, ExactValue{}, new_cc, nil).Value
									visited[index] = true
								}
								idx_list_len := len(sel.Index) - 1
								idx_list := make([]uint, idx_list_len)
								if lb_is_nested_possibly_constant(type_, sel, fv.Value) {
									is_constant := true
									cv_type := f.Type
									for j := 1; j < len(sel.Index); j++ {
										index_j := sel.Index[j]
										cvt := base_type(cv_type)
										if cvt.Kind == Type_Struct {
											if cvt.Struct.IsRawUnion {
												is_constant = false
												break
											}
											cv_type = cvt.Struct.Fields[index_j].Type
											if is_type_struct(cvt) {
												cv_field_remapping := lb_get_struct_remapping(m, cvt)
												remapped_index := cv_field_remapping[index_j]
												idx_list[j-1] = uint(remapped_index)
											} else {
												idx_list[j-1] = uint(index_j)
											}
										} else if cvt.Kind == Type_Array {
											cv_type = cvt.Array.Elem
											idx_list[j-1] = uint(index_j)
										} else {
											is_constant = false
											break
										}
									}
									if is_constant {
										elem_value := lb_const_value_internal(m, tav.Type, tav.Value, cc, tav.Type).Value
										if LLVMIsConstant(elem_value) != 0 && LLVMIsConstant(values[index]) != 0 {
											values[index] = llvm_const_insert_value(m, values[index], elem_value, idx_list, uint(idx_list_len))
										} else if is_local {
											p := m.CurrProcedure
											if LLVMIsConstant(values[index]) != 0 {
												addr := lb_add_local_generated(p, f.Type, false)
												lb_addr_store(p, addr, lbValue{Value: values[index], Type: f.Type})
												values[index] = lb_addr_load(p, addr).Value
											}
											ptr := LLVMGetOperand(values[index], 0)
											indices := make([]LLVMValueRef, idx_list_len)
											lt_u32 := lb_type(m, t_u32)
											for k := 0; k < idx_list_len; k++ {
												indices[k] = LLVMConstInt(lt_u32, uint64(idx_list[k]), false)
											}
											ptr = LLVMBuildGEP2(p.Builder, lb_type(m, f.Type), ptr, indices, uint(idx_list_len), "")
											ptr = LLVMBuildPointerCast(p.Builder, ptr, lb_type(m, alloc_type_pointer(tav.Type)), "")
											if LLVMIsALoadInst(elem_value) != 0 {
												sz := type_size_of(tav.Type)
												src := LLVMGetOperand(elem_value, 0)
												lb_mem_copy_non_overlapping(p, lbValue{Value: ptr, Type: t_rawptr}, lbValue{Value: src, Type: t_rawptr}, lb_const_int(m, t_int, uint64(sz)), false)
											} else {
												LLVMBuildStore(p.Builder, elem_value, ptr)
											}
											is_constant = false
										} else {
											is_constant = false
										}
									}
								}
							}
						}
					}
				} else {
					for i := isize(0); i < isize(len(cl.Elems)); i++ {
						f := type_.Struct.Fields[i]
						tav := cl.Elems[i].TAV
						index := field_remapping[f.Variable.FieldIndex]
						if elem_type_can_be_constant(f.Type) {
							val := lb_const_value_internal(m, f.Type, tav.Value, cc, tav.Type)
							values[index] = val.Value
							visited[index] = true
						}
					}
				}
				for i := 0; i < value_count; i++ {
					if !visited[i] {
						typ := LLVMStructGetTypeAtIndex(struct_type, uint(i))
						values[i] = LLVMConstNull(typ)
					}
				}
				is_constant := true
				for i := 0; i < value_count; i++ {
					val := values[i]
					if LLVMIsConstant(val) == 0 {
						is_constant = false
						break
					}
				}
				if is_constant {
					res.Value = llvm_const_named_struct_internal(m, struct_type, values, isize(value_count))
					return res
				} else {
					new_values := make([]LLVMValueRef, value_count)
					for i := 0; i < value_count; i++ {
						old_value := values[i]
						if LLVMIsConstant(old_value) != 0 {
							new_values[i] = old_value
						} else {
							new_values[i] = LLVMConstNull(LLVMTypeOf(old_value))
						}
					}
					constant_value := llvm_const_named_struct_internal(m, struct_type, new_values, isize(value_count))
					p := m.CurrProcedure
					v := lb_add_local_generated(p, res.Type, true)
					m.ExactValueCompoundLiteralAddrMap[value.ValueCompound] = v
					LLVMBuildStore(p.Builder, constant_value, v.Addr.Value)
					for i := 0; i < value_count; i++ {
						val := values[i]
						if LLVMIsConstant(val) == 0 {
							dst := LLVMBuildStructGEP2(p.Builder, llvm_addr_type(m, v.Addr), v.Addr.Value, uint(i), "")
							LLVMBuildStore(p.Builder, val, dst)
						}
					}
					return lb_addr_load(p, v)
				}
			} else if is_type_bit_set(type_) {
				cl := value.ValueCompound.CompoundLit
				if len(cl.Elems) == 0 {
					return lb_const_nil(m, original_type)
				}
				sz := type_size_of(type_)
				if sz == 0 {
					return lb_const_nil(m, original_type)
				}
				bits := BigInt{}
				one := BigInt{}
				big_int_from_u64(&one, 1)
				for i := isize(0); i < isize(len(cl.Elems)); i++ {
					e := cl.Elems[i]
					tav := e.TAV
					if tav.Mode != Addressing_Constant {
						continue
					}
					v := big_int_to_i64(&tav.Value.ValueInteger)
					lower := type_.BitSet.Lower
					index := uint64(v - lower)
					bit := BigInt{}
					big_int_from_u64(&bit, index)
					big_int_shl(&bit, &one, &bit)
					big_int_or(&bits, &bits, &bit)
				}
				res.Value = lb_big_int_to_llvm(m, original_type, &bits)
				return res
			} else if is_type_matrix(type_) {
				cl := value.ValueCompound.CompoundLit
				elem_type := type_.Matrix.Elem
				elem_count := isize(len(cl.Elems))
				if elem_count == 0 || !elem_type_can_be_constant(elem_type) {
					return lb_const_nil(m, original_type)
				}
				max_count := type_.Matrix.RowCount * type_.Matrix.ColumnCount
				total_count := matrixTypeTotalInternalElems(type_)
				values := make([]LLVMValueRef, total_count)
				if cl.Elems[0].Kind == Ast_FieldValue {
					for j := isize(0); j < isize(len(cl.Elems)); j++ {
						elem := cl.Elems[j]
						fv := elem.FieldValue
						if is_ast_range(fv.Field) {
							ie := fv.Field.BinaryExpr
							lo_tav := ie.Left.TAV
							hi_tav := ie.Right.TAV
							op := ie.Op.Kind
							lo := exact_value_to_i64(lo_tav.Value)
							hi := exact_value_to_i64(hi_tav.Value)
							if op != Token_RangeHalf {
								hi += 1
							}
							tav := fv.Value.TAV
							val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
							for k := lo; k < hi; k++ {
								offset := matrixRowMajorIndexToOffset(type_, k)
								values[offset] = val
							}
						} else {
							index_tav := fv.Field.TAV
							index := exact_value_to_i64(index_tav.Value)
							tav := fv.Value.TAV
							val := lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
							offset := matrixRowMajorIndexToOffset(type_, index)
							values[offset] = val
						}
					}
					for i := int64(0); i < total_count; i++ {
						if values[i] == 0 {
							values[i] = LLVMConstNull(lb_type(m, elem_type))
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(total_count), values, cc)
					return res
				} else {
					values := make([]LLVMValueRef, total_count)
					for i := isize(0); i < isize(len(cl.Elems)); i++ {
						tav := cl.Elems[i].TAV
						offset := matrixRowMajorIndexToOffset(type_, int64(i))
						values[offset] = lb_const_value_internal(m, elem_type, tav.Value, cc, tav.Type).Value
					}
					for i := int64(0); i < total_count; i++ {
						if values[i] == 0 {
							values[i] = LLVMConstNull(lb_type(m, elem_type))
						}
					}
					res.Value = lb_build_constant_array_values(m, type_, elem_type, isize(total_count), values, cc)
					return res
				}
			} else {
				return lb_const_nil(m, original_type)
			}
		}

	case ExactValue_Procedure:
		break

	case ExactValue_Typeid:
		return lb_typeid(m, value.ValueTypeid)
	}

	return lb_const_nil(m, original_type)
}
