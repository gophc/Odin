package cmd

import (
	"sort"
)

func lb_is_empty_string_constant(expr *Ast) bool {
	if expr.TAV.Value.Kind == ExactValue_String && is_type_string(expr.TAV.Type) {
		s := expr.TAV.Value.ValueString
		return s.Len == 0
	}
	return false
}

func reduce_tuple_to_single_type(t *Type) *Type {
	if t != nil {
		ct := core_type(t)
		if ct.Kind == Type_Tuple && len(ct.Tuple.Variables) == 1 {
			return ct.Tuple.Variables[0].Type
		}
	}
	return t
}

func lb_emit_conv(p *lbProcedure, value lbValue, t *Type) lbValue {
	m := p.Module
	t = reduce_tuple_to_single_type(t)

	src_type := value.Type
	if are_types_identical(t, src_type) {
		return value
	}

	src := core_type(src_type)
	dst := core_type(t)
	if src == nil {
		panic("src == nil")
	}
	if dst == nil {
		panic("dst == nil")
	}

	if is_type_untyped_uninit(src) {
		return lb_const_undef(m, t)
	}
	if is_type_untyped_nil(src) {
		return lb_const_nil(m, t)
	}

	if LLVMIsConstant(value.Value) != 0 {
		if is_type_any(dst) {
			st := default_type(src_type)
			default_value := lb_add_local_generated(p, st, false)
			lb_addr_store(p, default_value, value)
			data := lb_emit_conv(p, default_value.Addr, t_rawptr)
			id := lb_typeid(m, st)

			res := lb_add_local_generated(p, t, false)
			a0 := lb_emit_struct_ep(p, res.Addr, 0)
			a1 := lb_emit_struct_ep(p, res.Addr, 1)
			lb_emit_store(p, a0, data)
			lb_emit_store(p, a1, id)
			return lb_addr_load(p, res)
		} else if dst.Kind == Type_Basic {
			if src.Kind == Type_Basic && src.Basic.Kind == Basic_string && dst.Basic.Kind == Basic_cstring {
				str := lb_get_const_string(m, value)
				res := lbValue{}
				res.Type = t
				res.Value = llvm_cstring(m, str)
				return res
			} else if src.Kind == Type_Basic && src.Basic.Kind == Basic_string16 && dst.Basic.Kind == Basic_cstring16 {
				panic("TODO(bill): UTF-16 string")
			}
		}
	}

	if are_types_identical(src, dst) {
		if !are_types_identical(src_type, t) {
			return lb_emit_transmute(p, value, t)
		}
		return value
	}

	// bool <-> llvm bool
	if is_type_boolean(src) && dst == t_llvm_bool {
		res := lbValue{}
		res.Value = LLVMBuildICmp(p.Builder, LLVMIntNE, value.Value, LLVMConstNull(lb_type(m, src)), "")
		res.Type = t
		return res
	}
	if src == t_llvm_bool && is_type_boolean(dst) {
		res := lbValue{}
		res.Value = LLVMBuildZExt(p.Builder, value.Value, lb_type(m, dst), "")
		res.Type = t
		return res
	}

	// integer -> integer
	if is_type_integer(src) && is_type_integer(dst) {
		if !(src.Kind == Type_Basic && dst.Kind == Type_Basic) {
			panic("integer types must be Type_Basic")
		}
		sz := type_size_of(default_type(src))
		dz := type_size_of(default_type(dst))

		if sz == dz {
			if dz > 1 && !types_have_same_internal_endian(src, dst) {
				return lb_emit_byte_swap(p, value, t)
			}
			res := lbValue{}
			res.Value = value.Value
			res.Type = t
			return res
		}

		if sz > 1 && is_type_different_to_arch_endianness(src) {
			platform_src_type := integer_endian_type_to_platform_type(src)
			value = lb_emit_byte_swap(p, value, platform_src_type)
		}
		op := LLVMTrunc

		if dz < sz {
			op = LLVMTrunc
		} else if dz == sz {
			op = LLVMBitCast
		} else if dz > sz {
			if is_type_unsigned(src) {
				op = LLVMZExt
			} else {
				op = LLVMSExt
			}
		}

		if dz > 1 && is_type_different_to_arch_endianness(dst) {
			platform_dst_type := integer_endian_type_to_platform_type(dst)
			res := lbValue{}
			res.Value = LLVMBuildCast(p.Builder, op, value.Value, lb_type(m, platform_dst_type), "")
			res.Type = t
			return lb_emit_byte_swap(p, res, t)
		} else {
			res := lbValue{}
			res.Value = LLVMBuildCast(p.Builder, op, value.Value, lb_type(m, t), "")
			res.Type = t
			return res
		}
	}

	// boolean -> boolean/integer
	if is_type_boolean(src) && (is_type_boolean(dst) || is_type_integer(dst)) {
		b := LLVMBuildICmp(p.Builder, LLVMIntNE, value.Value, LLVMConstNull(lb_type(m, value.Type)), "")
		res := lbValue{}
		res.Value = LLVMBuildIntCast2(p.Builder, b, lb_type(m, t), false, "")
		res.Type = t
		return res
	}

	if is_type_cstring(src) && is_type_u8_ptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_u8_ptr(src) && is_type_cstring(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_cstring(src) && is_type_u8_multi_ptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_u8_multi_ptr(src) && is_type_cstring(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_cstring(src) && is_type_rawptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_rawptr(src) && is_type_cstring(dst) {
		return lb_emit_transmute(p, value, dst)
	}

	if are_types_identical(src, t_cstring) && are_types_identical(dst, t_string) {
		c := lb_emit_conv(p, value, t_cstring)
		args := []lbValue{c}
		s := lb_emit_runtime_call(p, "cstring_to_string", args)
		return lb_emit_conv(p, s, dst)
	}

	if is_type_cstring16(src) && is_type_u16_ptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_u16_ptr(src) && is_type_cstring16(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_cstring16(src) && is_type_u16_multi_ptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_u16_multi_ptr(src) && is_type_cstring16(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_cstring16(src) && is_type_rawptr(dst) {
		return lb_emit_transmute(p, value, dst)
	}
	if is_type_rawptr(src) && is_type_cstring16(dst) {
		return lb_emit_transmute(p, value, dst)
	}

	if are_types_identical(src, t_cstring16) && are_types_identical(dst, t_string16) {
		c := lb_emit_conv(p, value, t_cstring16)
		args := []lbValue{c}
		s := lb_emit_runtime_call(p, "cstring16_to_string16", args)
		return lb_emit_conv(p, s, dst)
	}

	// integer -> boolean
	if is_type_integer(src) && is_type_boolean(dst) {
		res := lbValue{}
		res.Value = LLVMBuildICmp(p.Builder, LLVMIntNE, value.Value, LLVMConstNull(lb_type(m, value.Type)), "")
		res.Type = t_llvm_bool
		return lb_emit_conv(p, res, t)
	}

	// float -> float
	if is_type_float(src) && is_type_float(dst) {
		sz := type_size_of(src)
		dz := type_size_of(dst)

		if dz == sz {
			if types_have_same_internal_endian(src, dst) {
				res := lbValue{}
				res.Type = t
				res.Value = value.Value
				return res
			} else {
				return lb_emit_byte_swap(p, value, t)
			}
		}

		if is_type_different_to_arch_endianness(src) || is_type_different_to_arch_endianness(dst) {
			platform_src_type := integer_endian_type_to_platform_type(src)
			platform_dst_type := integer_endian_type_to_platform_type(dst)
			res := lb_emit_conv(p, value, platform_src_type)
			res = lb_emit_conv(p, res, platform_dst_type)
			if is_type_different_to_arch_endianness(dst) {
				res = lb_emit_byte_swap(p, res, t)
			}
			return lb_emit_conv(p, res, t)
		}

		res := lbValue{}
		res.Type = t

		if dz >= sz {
			res.Value = LLVMBuildFPExt(p.Builder, value.Value, lb_type(m, t), "")
		} else {
			res.Value = LLVMBuildFPTrunc(p.Builder, value.Value, lb_type(m, t), "")
		}
		return res
	}

	if is_type_complex(src) && is_type_complex(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, lb_emit_struct_ev(p, value, 0), ft)
		imag := lb_emit_conv(p, lb_emit_struct_ev(p, value, 1), ft)
		fields := [2]lbValue{real, imag}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}

	if is_type_quaternion(src) && is_type_quaternion(dst) {
		ft := base_complex_elem_type(dst)
		q0 := lb_emit_conv(p, lb_emit_struct_ev(p, value, 0), ft)
		q1 := lb_emit_conv(p, lb_emit_struct_ev(p, value, 1), ft)
		q2 := lb_emit_conv(p, lb_emit_struct_ev(p, value, 2), ft)
		q3 := lb_emit_conv(p, lb_emit_struct_ev(p, value, 3), ft)
		fields := [4]lbValue{q0, q1, q2, q3}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}

	if is_type_integer(src) && is_type_complex(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, value, ft)
		fields := [2]lbValue{real, lb_const_nil(m, ft)}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}
	if is_type_float(src) && is_type_complex(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, value, ft)
		fields := [2]lbValue{real, lb_const_nil(m, ft)}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}

	if is_type_integer(src) && is_type_quaternion(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, value, ft)
		zero := lb_const_nil(m, ft)
		fields := [4]lbValue{zero, zero, zero, real}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}
	if is_type_float(src) && is_type_quaternion(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, value, ft)
		zero := lb_const_nil(m, ft)
		fields := [4]lbValue{zero, zero, zero, real}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}
	if is_type_complex(src) && is_type_quaternion(dst) {
		ft := base_complex_elem_type(dst)
		real := lb_emit_conv(p, lb_emit_struct_ev(p, value, 0), ft)
		imag := lb_emit_conv(p, lb_emit_struct_ev(p, value, 1), ft)
		zero := lb_const_nil(m, ft)
		fields := [4]lbValue{imag, zero, zero, real}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}

	// float <-> integer
	if is_type_float(src) && is_type_integer(dst) {
		if is_type_different_to_arch_endianness(src) || is_type_different_to_arch_endianness(dst) {
			platform_src_type := integer_endian_type_to_platform_type(src)
			platform_dst_type := integer_endian_type_to_platform_type(dst)
			res := lb_emit_conv(p, value, platform_src_type)
			res = lb_emit_conv(p, res, platform_dst_type)
			return lb_emit_conv(p, res, t)
		}

		if is_type_integer_128bit(dst) {
			args := []lbValue{value}
			call := "fixunsdfdi"
			if is_type_unsigned(dst) {
				call = "fixunsdfti"
			}
			res_i128 := lb_emit_runtime_call(p, call, args)
			return lb_emit_conv(p, res_i128, t)
		}
		sz := type_size_of(src)

		res := lbValue{}
		res.Type = t
		if is_type_unsigned(dst) {
			switch sz {
			case 2, 4:
				res.Value = LLVMBuildFPToUI(p.Builder, value.Value, lb_type(m, t_u32), "")
				res.Value = LLVMBuildIntCast2(p.Builder, res.Value, lb_type(m, t), false, "")
			case 8:
				res.Value = LLVMBuildFPToUI(p.Builder, value.Value, lb_type(m, t_u64), "")
				res.Value = LLVMBuildIntCast2(p.Builder, res.Value, lb_type(m, t), false, "")
			default:
				panic("Unhandled float type")
			}
		} else {
			switch sz {
			case 2, 4:
				res.Value = LLVMBuildFPToSI(p.Builder, value.Value, lb_type(m, t_i32), "")
				res.Value = LLVMBuildIntCast2(p.Builder, res.Value, lb_type(m, t), true, "")
			case 8:
				res.Value = LLVMBuildFPToSI(p.Builder, value.Value, lb_type(m, t_i64), "")
				res.Value = LLVMBuildIntCast2(p.Builder, res.Value, lb_type(m, t), true, "")
			default:
				panic("Unhandled float type")
			}
		}
		return res
	}
	if is_type_integer(src) && is_type_float(dst) {
		if is_type_different_to_arch_endianness(src) || is_type_different_to_arch_endianness(dst) {
			platform_src_type := integer_endian_type_to_platform_type(src)
			platform_dst_type := integer_endian_type_to_platform_type(dst)
			res := lb_emit_conv(p, value, platform_src_type)
			res = lb_emit_conv(p, res, platform_dst_type)
			if is_type_different_to_arch_endianness(dst) {
				res = lb_emit_byte_swap(p, res, t)
			}
			return lb_emit_conv(p, res, t)
		}

		if is_type_integer_128bit(src) {
			args := []lbValue{value}
			call := "floattidf"
			if is_type_unsigned(src) {
				call = "floattidf_unsigned"
			}
			res_f64 := lb_emit_runtime_call(p, call, args)
			return lb_emit_conv(p, res_f64, t)
		}

		res := lbValue{}
		res.Type = t
		if is_type_unsigned(src) {
			res.Value = LLVMBuildUIToFP(p.Builder, value.Value, lb_type(m, t), "")
		} else {
			res.Value = LLVMBuildSIToFP(p.Builder, value.Value, lb_type(m, t), "")
		}
		return res
	}

	if is_type_simd_vector(dst) {
		et := base_array_type(dst)
		if is_type_simd_vector(src) {
			src_elem := core_array_type(src)
			dst_elem := core_array_type(dst)

			if src.SimdVector.Count != dst.SimdVector.Count {
				panic("SIMD vector count mismatch")
			}

			res := lbValue{}
			res.Type = t
			if are_types_identical(src_elem, dst_elem) {
				res.Value = value.Value
			} else if is_type_float(src_elem) && is_type_integer(dst_elem) {
				if is_type_unsigned(dst_elem) {
					res.Value = LLVMBuildFPToUI(p.Builder, value.Value, lb_type(m, t), "")
				} else {
					res.Value = LLVMBuildFPToSI(p.Builder, value.Value, lb_type(m, t), "")
				}
			} else if is_type_integer(src_elem) && is_type_float(dst_elem) {
				if is_type_unsigned(src_elem) {
					res.Value = LLVMBuildUIToFP(p.Builder, value.Value, lb_type(m, t), "")
				} else {
					res.Value = LLVMBuildSIToFP(p.Builder, value.Value, lb_type(m, t), "")
				}
			} else if (is_type_integer(src_elem) || is_type_boolean(src_elem)) && is_type_integer(dst_elem) {
				res.Value = LLVMBuildIntCast2(p.Builder, value.Value, lb_type(m, t), !is_type_unsigned(src_elem), "")
			} else if is_type_float(src_elem) && is_type_float(dst_elem) {
				res.Value = LLVMBuildFPCast(p.Builder, value.Value, lb_type(m, t), "")
			} else if is_type_integer(src_elem) && is_type_boolean(dst_elem) {
				i1vector := LLVMBuildICmp(p.Builder, LLVMIntNE, value.Value, LLVMConstNull(LLVMTypeOf(value.Value)), "")
				res.Value = LLVMBuildIntCast2(p.Builder, i1vector, lb_type(m, t), !is_type_unsigned(src_elem), "")
			} else if is_type_pointer(src_elem) && is_type_integer(dst_elem) {
				res.Value = LLVMBuildPtrToInt(p.Builder, value.Value, lb_type(m, t), "")
			} else if is_type_integer(src_elem) && is_type_pointer(dst_elem) {
				res.Value = LLVMBuildIntToPtr(p.Builder, value.Value, lb_type(m, t), "")
			} else {
				panic("Unhandled simd vector conversion")
			}
			return res
		} else {
			count := get_array_type_count(dst)
			vt := lb_type(m, t)
			llvm_u32 := lb_type(m, t_u32)
			elem := lb_emit_conv(p, value, et).Value
			vector := LLVMConstNull(vt)
			for i := i64(0); i < count; i++ {
				idx := LLVMConstInt(llvm_u32, uint64(i), false)
				vector = LLVMBuildInsertElement(p.Builder, vector, elem, idx, "")
			}
			res := lbValue{}
			res.Type = t
			res.Value = vector
			return res
		}
	}

	// bit_field <-> backing type
	if is_type_bit_field(src) {
		if are_types_identical(src.BitField.BackingType, dst) {
			res := lbValue{}
			res.Type = t
			res.Value = value.Value
			return res
		}
	}
	if is_type_bit_field(dst) {
		if are_types_identical(src, dst.BitField.BackingType) {
			res := lbValue{}
			res.Type = t
			res.Value = value.Value
			return res
		}
	}

	// bit_set <-> backing type
	if is_type_bit_set(src) {
		backing := bit_set_to_int(src)
		if are_types_identical(backing, dst) {
			res := lbValue{}
			res.Type = t
			res.Value = value.Value
			return res
		}
	}
	if is_type_bit_set(dst) {
		backing := bit_set_to_int(dst)
		if are_types_identical(src, backing) {
			res := lbValue{}
			res.Type = t
			res.Value = value.Value
			return res
		}
	}

	// Pointer <-> uintptr
	if is_type_pointer(src) && is_type_uintptr(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPtrToInt(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_uintptr(src) && is_type_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildIntToPtr(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_multi_pointer(src) && is_type_uintptr(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPtrToInt(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_uintptr(src) && is_type_multi_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildIntToPtr(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}

	if is_type_union(dst) {
		if len(dst.Union.Variants) == 1 {
			vt := dst.Union.Variants[0]
			if internal_check_is_assignable_to(src_type, vt) {
				value = lb_emit_conv(p, value, vt)
				parent := lb_add_local_generated(p, t, true)
				lb_emit_store_union_variant(p, parent.Addr, value, vt)
				return lb_addr_load(p, parent)
			}
		}
		for _, vt := range dst.Union.Variants {
			if src_type == t_llvm_bool && is_type_boolean(vt) {
				parent := lb_add_local_generated(p, t, true)
				lb_emit_store_union_variant(p, parent.Addr, value, vt)
				return lb_addr_load(p, parent)
			}
			if are_types_identical(src_type, vt) {
				parent := lb_add_local_generated(p, t, true)
				lb_emit_store_union_variant(p, parent.Addr, value, vt)
				return lb_addr_load(p, parent)
			}
		}
		valids := make([]ValidIndexAndScore, len(dst.Union.Variants))
		valid_count := 0
		first_success_index := -1
		for i, vt := range dst.Union.Variants {
			score := i64(0)
			if internal_check_is_assignable_to(src_type, vt) {
				valids[valid_count].Index = i
				valids[valid_count].Score = score
				valid_count++
				if first_success_index < 0 {
					first_success_index = i
				}
			}
		}
		if valid_count > 1 {
			sort.Slice(valids[:valid_count], func(i, j int) bool {
				return valids[i].Score > valids[j].Score
			})
			best_score := valids[0].Score
			for i := 1; i < valid_count; i++ {
				v := valids[i]
				if best_score > v.Score {
					valid_count = i
					break
				}
				best_score = v.Score
			}
			first_success_index = valids[0].Index
		}

		if valid_count == 1 {
			vt := dst.Union.Variants[first_success_index]
			value = lb_emit_conv(p, value, vt)
			parent := lb_add_local_generated(p, t, true)
			lb_emit_store_union_variant(p, parent.Addr, value, vt)
			return lb_addr_load(p, parent)
		}
	}

	// subtype polymorphism casting
	if check_is_assignable_to_using_subtype(src_type, t) {
		st := type_deref(src_type)
		st = type_deref(st)

		st_is_ptr := is_type_pointer(src_type)
		st = base_type(st)

		dt := t

		if !(is_type_struct(st) || is_type_raw_union(st)) {
			panic("invalid subtype cast")
		}
		sel := Selection{}
		if lookup_subtype_polymorphic_selection(t, src_type, &sel) {
			if sel.Entity == nil {
				panic("invalid subtype cast")
			}
			if st_is_ptr {
				res := lb_emit_deep_field_gep(p, value, sel)
				rt := res.Type
				if !are_types_identical(rt, dt) && are_types_identical(type_deref(rt), dt) {
					res = lb_emit_load(p, res)
				}
				return res
			} else {
				if is_type_pointer(value.Type) {
					rt := value.Type
					if !are_types_identical(rt, dt) && are_types_identical(type_deref(rt), dt) {
						value = lb_emit_load(p, value)
					} else {
						value = lb_emit_deep_field_gep(p, value, sel)
						return lb_emit_load(p, value)
					}
				}
				return lb_emit_deep_field_ev(p, value, sel)
			}
		}
	}

	// Pointer <-> Pointer
	if is_type_pointer(src) && is_type_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_multi_pointer(src) && is_type_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_pointer(src) && is_type_multi_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	if is_type_multi_pointer(src) && is_type_multi_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}

	// proc <-> proc
	if is_type_proc(src) && is_type_proc(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}

	// pointer -> proc
	if is_type_pointer(src) && is_type_proc(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}
	// proc -> pointer
	if is_type_proc(src) && is_type_pointer(dst) {
		res := lbValue{}
		res.Type = t
		res.Value = LLVMBuildPointerCast(p.Builder, value.Value, lb_type(m, t), "")
		return res
	}

	// [^]u16 <-> cstring16
	if is_type_u16_multi_ptr(src) && is_type_cstring16(dst) {
		return lb_emit_transmute(p, value, t)
	}
	if is_type_cstring16(src) && is_type_u16_multi_ptr(dst) {
		return lb_emit_transmute(p, value, t)
	}
	if is_type_u16_ptr(src) && is_type_cstring16(dst) {
		return lb_emit_transmute(p, value, t)
	}
	if is_type_cstring16(src) && is_type_u16_ptr(dst) {
		return lb_emit_transmute(p, value, t)
	}

	// []u16 <-> string16
	if is_type_u16_slice(src) && is_type_string16(dst) {
		return lb_emit_transmute(p, value, t)
	}
	if is_type_string16(src) && is_type_u16_slice(dst) {
		return lb_emit_transmute(p, value, t)
	}

	// []byte/[]u8 <-> string
	if is_type_u8_slice(src) && is_type_string(dst) {
		return lb_emit_transmute(p, value, t)
	}
	if is_type_string(src) && is_type_u8_slice(dst) {
		return lb_emit_transmute(p, value, t)
	}

	if is_type_array(dst) && is_type_array(src) {
		dst_elem := base_array_type(dst)
		src_elem := base_array_type(src)
		if dst.Array.Count == src.Array.Count &&
			!is_type_array_like(dst.Array.Elem) &&
			!is_type_array_like(src.Array.Elem) {
			if are_types_identical(dst_elem, src_elem) {
				v := value
				v.Type = t
				return v
			}
			count := dst.Array.Count

			de := core_type(dst_elem)
			se := core_type(src_elem)
			if count <= MAX_SIMD_ARRAY_COUNT_FOR_INLINING &&
				is_simd_able_type(se) {
				if is_simd_able_type(de) {
					v := lb_add_local_generated(p, t, false)

					psrc := lb_address_from_load_or_generate_local(p, value)
					pdst := v.Addr

					de_sz := type_size_of(de)
					se_sz := type_size_of(se)

					op := LLVMOpcode(0)

					if is_type_float(se) {
						if is_type_float(de) {
							if de_sz == se_sz {
								op = LLVMBitCast
							} else if de_sz < se_sz {
								op = LLVMFPTrunc
							} else {
								op = LLVMFPExt
							}
						} else {
							if !(is_type_integer_like(de) || is_type_rune(de)) {
								panic("expected integer-like or rune")
							}
							if is_type_unsigned(de) || is_type_boolean(de) {
								op = LLVMFPToUI
							} else {
								op = LLVMFPToSI
							}
						}
					} else {
						if !(is_type_integer_like(se) || is_type_rune(se)) {
							panic("expected integer-like or rune")
						}
						if is_type_float(de) {
							if is_type_unsigned(se) || is_type_boolean(se) {
								op = LLVMUIToFP
							} else {
								op = LLVMSIToFP
							}
						} else {
							if de_sz == se_sz {
								op = LLVMBitCast
							} else if de_sz < se_sz {
								op = LLVMTrunc
							} else {
								if is_type_unsigned(se) || is_type_boolean(se) {
									op = LLVMZExt
								} else {
									op = LLVMSExt
								}
							}
						}
					}

					if op == 0 {
						panic("unhandled array conversion op")
					}

					src_vector_type := LLVMVectorType(lb_type(p.Module, se), uint(count))
					dst_vector_type := LLVMVectorType(lb_type(p.Module, de), uint(count))

					src_ptr := LLVMBuildPointerCast(p.Builder, psrc.Value, LLVMPointerType(src_vector_type, 0), "")
					dst_ptr := LLVMBuildPointerCast(p.Builder, pdst.Value, LLVMPointerType(dst_vector_type, 0), "")

					src_vector := LLVMBuildLoad2(p.Builder, src_vector_type, src_ptr, "")
					LLVMSetAlignment(src_vector, uint(type_align_of(se)))

					dst_vector := LLVMBuildCast(p.Builder, op, src_vector, dst_vector_type, "")

					store := LLVMBuildStore(p.Builder, dst_vector, dst_ptr)
					LLVMSetAlignment(store, uint(type_align_of(de)))

					return lb_addr_load(p, v)
				} else if is_type_complex(de) {
					if !is_type_float(se) {
						panic("expected float src for complex dst")
					}

					cde := base_complex_elem_type(de)

					v := lb_add_local_generated(p, t, false)

					cde_sz := type_size_of(cde)
					se_sz := type_size_of(se)

					psrc := lb_address_from_load_or_generate_local(p, value)
					pdst := v.Addr

					op := LLVMOpcode(0)

					if cde_sz == se_sz {
						op = LLVMBitCast
					} else if cde_sz < se_sz {
						op = LLVMFPTrunc
					} else {
						op = LLVMFPExt
					}

					src_vector_type := LLVMVectorType(lb_type(p.Module, se), uint(count))
					dst_vector_type := LLVMVectorType(lb_type(p.Module, cde), uint(count))

					src_ptr := LLVMBuildPointerCast(p.Builder, psrc.Value, LLVMPointerType(src_vector_type, 0), "")
					dst_ptr := LLVMBuildPointerCast(p.Builder, pdst.Value, LLVMPointerType(dst_vector_type, 0), "")

					src_vector := LLVMBuildLoad2(p.Builder, src_vector_type, src_ptr, "")
					LLVMSetAlignment(src_vector, uint(type_align_of(se)))

					dst_vector := LLVMBuildCast(p.Builder, op, src_vector, dst_vector_type, "")
					dst_zero := LLVMConstNull(dst_vector_type)

					dst_mask_scalars := make([]LLVMValueRef, count*2)
					llvm_i32 := lb_type(p.Module, t_i32)
					for i := i64(0); i < count; i++ {
						dst_mask_scalars[i*2+0] = LLVMConstInt(llvm_i32, uint64(i), false)
						dst_mask_scalars[i*2+1] = LLVMConstInt(llvm_i32, uint64(count+i), false)
					}

					dst_mask := LLVMConstVector(dst_mask_scalars, uint(count*2))
					dst_vector = LLVMBuildShuffleVector(p.Builder, dst_vector, dst_zero, dst_mask, "")

					store := LLVMBuildStore(p.Builder, dst_vector, dst_ptr)
					LLVMSetAlignment(store, uint(type_align_of(de)))

					return lb_addr_load(p, v)
				}
			}

			v := lb_add_local_generated(p, t, true)

			psrc := lb_address_from_load_or_generate_local(p, value)
			pdst := v.Addr

			if count > MAX_SIMD_ARRAY_COUNT_FOR_INLINING {
				loop_data := lb_loop_start(p, count, t_int)

				sp := lb_emit_array_ep(p, psrc, loop_data.Idx)
				dp := lb_emit_array_ep(p, pdst, loop_data.Idx)

				s := lb_emit_load(p, sp)
				s = lb_emit_conv(p, s, dst_elem)
				lb_emit_store(p, dp, s)

				lb_loop_end(p, loop_data)
			} else {
				for i := i64(0); i < count; i++ {
					sp := lb_emit_array_epi(p.Module, psrc, i)
					dp := lb_emit_array_epi(p.Module, pdst, i)

					s := lb_emit_load(p, sp)
					s = lb_emit_conv(p, s, dst_elem)
					lb_emit_store(p, dp, s)
				}
			}

			return lb_addr_load(p, v)
		}
	}

	if is_type_array_like(dst) {
		elem := base_array_type(dst)
		index_count := get_array_type_count(dst)

		inlineable := type_size_of(dst) <= build_context.MaxSimdAlign
		e := lb_emit_conv(p, value, elem)
		if inlineable && lb_is_const(e) {
			v := lbAddr{}
			if e.Value != 0 {
				values := make([]LLVMValueRef, index_count)
				for i := isize(0); i < index_count; i++ {
					values[i] = e.Value
				}
				array_const_value := lbValue{}
				array_const_value.Type = t
				array_const_value.Value = LLVMConstArray(lb_type(m, elem), values, uint(index_count))
				v = lb_add_global_generated_from_procedure(p, t, array_const_value)
			} else {
				v = lb_add_global_generated_from_procedure(p, t)
			}

			lb_make_global_private_const(v)
			return lb_addr_load(p, v)
		}

		v := lb_add_local_generated(p, t, false)

		if !inlineable {
			loop_data := lb_loop_start(p, index_count, t_int)

			elem_ := lb_emit_array_ep(p, v.Addr, loop_data.Idx)
			lb_emit_store(p, elem_, e)

			lb_loop_end(p, loop_data)
		} else {
			for i := isize(0); i < index_count; i++ {
				elem_ := lb_emit_array_epi(p.Module, v.Addr, i)
				lb_emit_store(p, elem_, e)
			}
		}
		return lb_addr_load(p, v)
	}

	if is_type_matrix(dst) && !is_type_matrix(src) {
		if dst.Matrix.RowCount != dst.Matrix.ColumnCount {
			panic("matrix must be square for scalar to matrix conversion")
		}

		elem := base_array_type(dst)
		e := lb_emit_conv(p, value, elem)
		v := lb_add_local_generated(p, t, false)
		zero := lb_const_value(p.Module, elem, exact_value_i64(0), LB_CONST_CONTEXT_DEFAULT_ALLOW_LOCAL)
		for j := i64(0); j < dst.Matrix.ColumnCount; j++ {
			for i := i64(0); i < dst.Matrix.RowCount; i++ {
				ptr := lb_emit_matrix_epi(p, v.Addr, i, j)
				if i == j {
					lb_emit_store(p, ptr, e)
				} else {
					lb_emit_store(p, ptr, zero)
				}
			}
		}

		return lb_addr_load(p, v)
	}

	if is_type_matrix(dst) && is_type_matrix(src) {
		if !(dst.Kind == Type_Matrix) {
			panic("dst must be Type_Matrix")
		}
		if !(src.Kind == Type_Matrix) {
			panic("src must be Type_Matrix")
		}
		v := lb_add_local_generated(p, t, true)

		if dst.Matrix.RowCount == src.Matrix.RowCount &&
			dst.Matrix.ColumnCount == src.Matrix.ColumnCount {
			for j := i64(0); j < dst.Matrix.ColumnCount; j++ {
				for i := i64(0); i < dst.Matrix.RowCount; i++ {
					d := lb_emit_matrix_epi(p, v.Addr, i, j)
					s := lb_emit_matrix_ev(p, value, i, j)
					s = lb_emit_conv(p, s, dst.Matrix.Elem)
					lb_emit_store(p, d, s)
				}
			}
		} else if is_matrix_square(dst) && is_matrix_square(src) {
			for j := i64(0); j < dst.Matrix.ColumnCount; j++ {
				for i := i64(0); i < dst.Matrix.RowCount; i++ {
					if i < src.Matrix.RowCount && j < src.Matrix.ColumnCount {
						d := lb_emit_matrix_epi(p, v.Addr, i, j)
						s := lb_emit_matrix_ev(p, value, i, j)
						s = lb_emit_conv(p, s, dst.Matrix.Elem)
						lb_emit_store(p, d, s)
					} else if i == j {
						d := lb_emit_matrix_epi(p, v.Addr, i, j)
						s := lb_const_value(p.Module, dst.Matrix.Elem, exact_value_i64(1), LB_CONST_CONTEXT_DEFAULT_ALLOW_LOCAL)
						lb_emit_store(p, d, s)
					}
				}
			}
		} else {
			dst_count := dst.Matrix.RowCount * dst.Matrix.ColumnCount
			src_count := src.Matrix.RowCount * src.Matrix.ColumnCount
			if dst_count != src_count {
				panic("matrix element count mismatch")
			}

			pdst := v.Addr
			psrc := lb_address_from_load_or_generate_local(p, value)

			same_elem_base_types := are_types_identical(
				base_type(dst.Matrix.Elem),
				base_type(src.Matrix.Elem),
			)

			if same_elem_base_types && type_size_of(dst) == type_size_of(src) {
				lb_mem_copy_overlapping(p, v.Addr, psrc, lb_const_int(p.Module, t_int, uint64(type_size_of(dst))))
			} else {
				for i := i64(0); i < src_count; i++ {
					dp := lb_emit_array_epi(p.Module, v.Addr, matrix_column_major_index_to_offset(dst, i))
					sp := lb_emit_array_epi(p.Module, psrc, matrix_column_major_index_to_offset(src, i))
					s := lb_emit_load(p, sp)
					s = lb_emit_conv(p, s, dst.Matrix.Elem)
					lb_emit_store(p, dp, s)
				}
			}
		}
		return lb_addr_load(p, v)
	}

	if is_type_any(dst) {
		if is_type_untyped_uninit(src) {
			return lb_const_undef(p.Module, t)
		}
		if is_type_untyped_nil(src) {
			return lb_const_nil(p.Module, t)
		}

		st := default_type(src_type)

		data := lb_address_from_load_or_generate_local(p, value)
		if !is_type_pointer(data.Type) {
			panic("expected pointer type for data")
		}
		if !is_type_typed(st) {
			panic("expected typed type")
		}
		data = lb_emit_conv(p, data, t_rawptr)

		id := lb_typeid(p.Module, st)
		fields := [2]lbValue{data, id}
		return lb_build_struct_value(p, t, fields[:], len(fields))
	}

	src_sz := type_size_of(src)
	dst_sz := type_size_of(dst)

	if src_sz == dst_sz {
		// bit_set <-> integer
		if is_type_integer(src) && is_type_bit_set(dst) {
			res := lb_emit_conv(p, value, bit_set_to_int(dst))
			res.Type = t
			return res
		}
		if is_type_bit_set(src) && is_type_integer(dst) {
			bs := value
			bs.Type = bit_set_to_int(src)
			return lb_emit_conv(p, bs, dst)
		}

		// typeid <-> integer
		if is_type_integer(src) && is_type_typeid(dst) {
			return lb_emit_transmute(p, value, dst)
		}
		if is_type_typeid(src) && is_type_integer(dst) {
			return lb_emit_transmute(p, value, dst)
		}
	}

	if is_type_untyped(src) {
		if is_type_string(src) && is_type_string(dst) {
			result := lb_add_local_generated(p, t, false)
			lb_addr_store(p, result, value)
			return lb_addr_load(p, result)
		}
	}

	panic("Invalid type conversion")
}
