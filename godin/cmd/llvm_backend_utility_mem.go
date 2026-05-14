package cmd

import (
	"sort"
	"unsafe"
)

const (
	ODIN_METADATA_IS_PACKED  = "odin-is-packed"
	ODIN_METADATA_MIN_ALIGN  = "odin-min-align"
	ODIN_METADATA_MAX_ALIGN  = "odin-max-align"
	MAP_CACHE_LINE_SIZE      = 64
)

func lb_get_struct_remapping(m *lbModule, t *Type) lbStructFieldRemapping {
	t = base_type(t)
	struct_type := lb_type(m, t)

	m.TypesMutex.Lock()
	field_remapping, ok := m.StructFieldRemapping[uintptr(unsafe.Pointer(t))]
	if !ok {
		field_remapping = m.StructFieldRemapping[uintptr(struct_type)]
	}
	m.TypesMutex.Unlock()

	gb_assert_handler("Assertion Failure", "field_remapping != nil", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(t))
	return field_remapping
}

func lb_convert_struct_index(m *lbModule, t *Type, index i32) i32 {
	if t.Kind == Type_Struct {
		if t.Struct.IsRawUnion {
			return 0
		}
		field_remapping := lb_get_struct_remapping(m, t)
		return field_remapping[index]
	} else if is_type_any(t) && buildContext.PtrSize == 4 {
		gb_assert_handler("Assertion Failure", "t->kind == Type_Basic", "llvm_backend_utility_mem.go", 0)
		gb_assert_handler("Assertion Failure", "t->Basic.kind == Basic_any", "llvm_backend_utility_mem.go", 0)
		switch index {
		case 0:
			return 0
		case 1:
			return 2
		default:
			gb_assert_handler("Panic", "index > 1", "llvm_backend_utility_mem.go", 0)
		}
	} else if buildContext.PtrSize != buildContext.IntSize {
		switch t.Kind {
		case Type_Basic:
			if t.Basic.Kind != BasicString &&
				t.Basic.Kind != BasicString16 {
				break
			}
			fallthrough
		case Type_Slice:
			gb_assert_handler("Assertion Failure", "buildContext.PtrSize*2 == buildContext.IntSize", "llvm_backend_utility_mem.go", 0)
			switch index {
			case 0:
				return 0
			case 1:
				return 2
			}
			break
		case Type_DynamicArray:
			gb_assert_handler("Assertion Failure", "buildContext.PtrSize*2 == buildContext.IntSize", "llvm_backend_utility_mem.go", 0)
			switch index {
			case 0:
				return 0
			case 1:
				return 2
			case 2:
				return 3
			case 3:
				return 4
			}
			break
		case Type_SoaPointer:
			gb_assert_handler("Assertion Failure", "buildContext.PtrSize*2 == buildContext.IntSize", "llvm_backend_utility_mem.go", 0)
			switch index {
			case 0:
				return 0
			case 1:
				return 2
			}
			break
		}
	}
	if t.Kind == Type_FixedCapacityDynamicArray {
		switch index {
		case 0:
			return 0
		case 1:
			if t.FixedCapacityDynamicArray.PaddingNeeded > 0 {
				return 2
			}
			return 1
		}
	}
	return index
}

func lb_type_padding_filler(m *lbModule, padding i64, padding_align i64) LLVMTypeRef {
	mutex_lock(&m.PadTypesMutex)
	if padding%padding_align == 0 {
		for _, pd := range m.PadTypes {
			if pd.Padding == padding && pd.PaddingAlign == padding_align {
				mutex_unlock(&m.PadTypesMutex)
				return pd.Type
			}
		}
	} else {
		for _, pd := range m.PadTypes {
			if pd.Padding == padding && pd.PaddingAlign == 1 {
				mutex_unlock(&m.PadTypesMutex)
				return pd.Type
			}
		}
	}

	if padding_align < 1 {
		padding_align = 1
	} else if padding_align > 8 {
		padding_align = 8
	}
	if padding%padding_align == 0 {
		var elem LLVMTypeRef
		llen := padding / padding_align
		switch padding_align {
		case 1:
			elem = lb_type(m, t_u8)
		case 2:
			elem = lb_type(m, t_u16)
		case 4:
			elem = lb_type(m, t_u32)
		case 8:
			elem = lb_type(m, t_u64)
		}
		gb_assert_handler("Assertion Failure", "elem != nil", "llvm_backend_utility_mem.go", 0, "Invalid lb_type_padding_filler padding and padding_align: %d", padding_align)

		var llvm_type LLVMTypeRef
		if llen != 1 {
			llvm_type = llvm_array_type(elem, uint64(llen))
		} else {
			llvm_type = elem
		}
		m.PadTypes = append(m.PadTypes, lbPadType{Padding: padding, PaddingAlign: padding_align, Type: llvm_type})
		mutex_unlock(&m.PadTypesMutex)
		return llvm_type
	} else {
		llvm_type := llvm_array_type(lb_type(m, t_u8), uint64(padding))
		m.PadTypes = append(m.PadTypes, lbPadType{Padding: padding, PaddingAlign: 1, Type: llvm_type})
		mutex_unlock(&m.PadTypesMutex)
		return llvm_type
	}
}

var llvm_type_kinds = [...]string{
	"LLVMVoidTypeKind",
	"LLVMHalfTypeKind",
	"LLVMFloatTypeKind",
	"LLVMDoubleTypeKind",
	"LLVMX86_FP80TypeKind",
	"LLVMFP128TypeKind",
	"LLVMPPC_FP128TypeKind",
	"LLVMLabelTypeKind",
	"LLVMIntegerTypeKind",
	"LLVMFunctionTypeKind",
	"LLVMStructTypeKind",
	"LLVMArrayTypeKind",
	"LLVMPointerTypeKind",
	"LLVMVectorTypeKind",
	"LLVMMetadataTypeKind",
	"LLVMX86_MMXTypeKind",
	"LLVMTokenTypeKind",
	"LLVMScalableVectorTypeKind",
	"LLVMBFloatTypeKind",
}

func lb_emit_struct_ep_internal(p *lbProcedure, s lbValue, index i32, result_type *Type) lbValue {
	t := base_type(type_deref(s.Type))

	original_index := index
	index = lb_convert_struct_index(p.Module, t, index)

	if lb_is_const(s) {
		m := p.Module
		var res lbValue
		indices := [2]LLVMValueRef{
			llvm_zero(m),
			LLVMConstInt(lb_type(m, t_i32), uint64(index), 0),
		}
		res.Value = LLVMConstGEP2(lb_type(m, type_deref(s.Type)), s.Value, indices[:], 2)
		res.Type = alloc_type_pointer(result_type)
		return res
	} else {
		var res lbValue
		st := lb_type(p.Module, type_deref(s.Type))
		gb_assert_handler("Assertion Failure", "LLVMGetTypeKind(st) == LLVMStructTypeKind", "llvm_backend_utility_mem.go", 0, "%s", llvm_type_kinds[LLVMGetTypeKind(st)])
		count := LLVMCountStructElementTypes(st)
		gb_assert_handler("Assertion Failure", "count >= uint(index)", "llvm_backend_utility_mem.go", 0, "%d %d %d", count, index, original_index)

		res.Value = LLVMBuildStructGEP2(p.Builder, st, s.Value, uint(index), "")
		res.Type = alloc_type_pointer(result_type)
		return res
	}
}

func lb_emit_tuple_ep(p *lbProcedure, ptr lbValue, index i32) lbValue {
	t := type_deref(ptr.Type)
	gb_assert_handler("Assertion Failure", "is_type_tuple(t)", "llvm_backend_utility_mem.go", 0)
	result_type := t.Tuple.Variables[index].Type

	var res lbValue
	tf, ok := p.TupleFixMap[ptr.Value]
	if ok {
		res = tf.Values[index]
		gb_assert_handler("Assertion Failure", "are_types_identical(res.type, result_type)", "llvm_backend_utility_mem.go", 0)
		res = lb_address_from_load_or_generate_local(p, res)
	} else {
		res = lb_emit_struct_ep_internal(p, ptr, index, result_type)
	}
	return res
}

func lb_emit_struct_ep(p *lbProcedure, s lbValue, index i32) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_pointer(s.type)", "llvm_backend_utility_mem.go", 0)
	t := base_type(type_deref(s.Type))
	var result_type *Type

	if is_type_struct(t) {
		result_type = get_struct_field_type(t, index)
	} else if is_type_union(t) {
		gb_assert_handler("Assertion Failure", "index == -1", "llvm_backend_utility_mem.go", 0)
		return lb_emit_union_tag_ptr(p, s)
	} else if is_type_tuple(t) {
		return lb_emit_tuple_ep(p, s, index)
	} else if is_type_complex(t) {
		ft := base_complex_elem_type(t)
		switch index {
		case 0:
			result_type = ft
		case 1:
			result_type = ft
		}
	} else if is_type_quaternion(t) {
		ft := base_complex_elem_type(t)
		switch index {
		case 0:
			result_type = ft
		case 1:
			result_type = ft
		case 2:
			result_type = ft
		case 3:
			result_type = ft
		}
	} else if is_type_slice(t) {
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.Slice.Elem)
		case 1:
			result_type = t_int
		}
	} else if is_type_string16(t) {
		switch index {
		case 0:
			result_type = t_u16_ptr
		case 1:
			result_type = t_int
		}
	} else if is_type_string(t) {
		switch index {
		case 0:
			result_type = t_u8_ptr
		case 1:
			result_type = t_int
		}
	} else if is_type_any(t) {
		switch index {
		case 0:
			result_type = t_rawptr
		case 1:
			result_type = t_typeid
		default:
			gb_assert_handler("Panic", "index > 1", "llvm_backend_utility_mem.go", 0)
		}
	} else if is_type_dynamic_array(t) {
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.DynamicArray.Elem)
		case 1:
			result_type = t_int
		case 2:
			result_type = t_int
		case 3:
			result_type = t_allocator
		}
	} else if is_type_fixed_capacity_dynamic_array(t) {
		switch index {
		case 0:
			result_type = alloc_type_array(t.FixedCapacityDynamicArray.Elem, t.FixedCapacityDynamicArray.Capacity, nil)
		case 1:
			result_type = t_int
		}
	} else if is_type_map(t) {
		init_map_internal_debug_types(t)
		itp := alloc_type_pointer(t_raw_map)
		s = lb_emit_transmute(p, s, itp)

		switch index {
		case 0:
			result_type = get_struct_field_type(t_raw_map, 0)
		case 1:
			result_type = get_struct_field_type(t_raw_map, 1)
		case 2:
			result_type = get_struct_field_type(t_raw_map, 2)
		}
	} else if is_type_array(t) {
		return lb_emit_array_epi_proc(p, s, isize(index))
	} else if is_type_soa_pointer(t) {
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.SoaPointer.Elem)
		case 1:
			result_type = t_int
		}
	} else {
		gb_assert_handler("Panic", "TODO: struct_gep type", "llvm_backend_utility_mem.go", 0, "%s, %d", type_to_string(s.Type), index)
	}

	gb_assert_handler("Assertion Failure", "result_type != nil", "llvm_backend_utility_mem.go", 0, "%s %d", type_to_string(t), index)

	gep := lb_emit_struct_ep_internal(p, s, index, result_type)

	bt := base_type(t)
	if bt.Kind == Type_Struct {
		if bt.Struct.IsPacked {
			lb_set_metadata_custom_u64(p.Module, gep.Value, ODIN_METADATA_IS_PACKED, 1)
			gb_assert_handler("Assertion Failure", "lb_get_metadata_custom_u64(...) == 1", "llvm_backend_utility_mem.go", 0)
		}
		align_max := bt.Struct.CustomMaxFieldAlign
		align_min := bt.Struct.CustomMinFieldAlign
		gb_assert_handler("Assertion Failure", "align_min == 0 || align_max == 0 || align_min <= align_max", "llvm_backend_utility_mem.go", 0)
		if align_max > 0 {
			lb_set_metadata_custom_u64(p.Module, gep.Value, ODIN_METADATA_MAX_ALIGN, u64(align_max))
			gb_assert_handler("Assertion Failure", "lb_get_metadata_custom_u64(...) == align_max", "llvm_backend_utility_mem.go", 0)
		}
		if align_min > 0 {
			lb_set_metadata_custom_u64(p.Module, gep.Value, ODIN_METADATA_MIN_ALIGN, u64(align_min))
			gb_assert_handler("Assertion Failure", "lb_get_metadata_custom_u64(...) == align_min", "llvm_backend_utility_mem.go", 0)
		}
	}

	return gep
}

func lb_emit_tuple_ev(p *lbProcedure, value lbValue, index i32) lbValue {
	t := value.Type
	gb_assert_handler("Assertion Failure", "is_type_tuple(t)", "llvm_backend_utility_mem.go", 0)
	result_type := t.Tuple.Variables[index].Type

	var res lbValue
	tf, ok := p.TupleFixMap[value.Value]
	if ok {
		res = tf.Values[index]
		gb_assert_handler("Assertion Failure", "are_types_identical(res.type, result_type)", "llvm_backend_utility_mem.go", 0)
	} else {
		if t.Tuple.Variables.Count == 1 {
			gb_assert_handler("Assertion Failure", "index == 0", "llvm_backend_utility_mem.go", 0)
			return value
		}
		if LLVMIsALoadInst(value.Value) != 0 {
			var res_ lbValue
			res_.Value = LLVMGetOperand(value.Value, 0)
			res_.Type = alloc_type_pointer(value.Type)
			ptr := lb_emit_struct_ep(p, res_, index)
			return lb_emit_load(p, ptr)
		}

		res.Value = LLVMBuildExtractValue(p.Builder, value.Value, uint(index), "")
		res.Type = result_type
	}
	return res
}

func lb_emit_struct_ev(p *lbProcedure, s lbValue, index i32) lbValue {
	t := base_type(s.Type)
	if is_type_tuple(t) {
		return lb_emit_tuple_ev(p, s, index)
	}

	if LLVMIsALoadInst(s.Value) != 0 {
		var res lbValue
		res.Value = LLVMGetOperand(s.Value, 0)
		res.Type = alloc_type_pointer(s.Type)
		ptr := lb_emit_struct_ep(p, res, index)
		return lb_emit_load(p, ptr)
	}

	var result_type *Type

	switch t.Kind {
	case Type_Basic:
		switch t.Basic.Kind {
		case BasicString16:
			switch index {
			case 0:
				result_type = t_u16_ptr
			case 1:
				result_type = t_int
			}
		case BasicString:
			switch index {
			case 0:
				result_type = t_u8_ptr
			case 1:
				result_type = t_int
			}
		case BasicAny:
			switch index {
			case 0:
				result_type = t_rawptr
			case 1:
				result_type = t_typeid
			}
		case BasicComplex32, BasicComplex64, BasicComplex128:
			ft := base_complex_elem_type(t)
			switch index {
			case 0:
				result_type = ft
			case 1:
				result_type = ft
			}
		case BasicQuaternion64, BasicQuaternion128, BasicQuaternion256:
			ft := base_complex_elem_type(t)
			switch index {
			case 0:
				result_type = ft
			case 1:
				result_type = ft
			case 2:
				result_type = ft
			case 3:
				result_type = ft
			}
		}
	case Type_Struct:
		result_type = get_struct_field_type(t, index)
	case Type_Union:
		gb_assert_handler("Assertion Failure", "index == -1", "llvm_backend_utility_mem.go", 0)
		gb_assert_handler("Panic", "lb_emit_union_tag_value", "llvm_backend_utility_mem.go", 0)
	case Type_Tuple:
		return lb_emit_tuple_ev(p, s, index)
	case Type_Slice:
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.Slice.Elem)
		case 1:
			result_type = t_int
		}
	case Type_DynamicArray:
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.DynamicArray.Elem)
		case 1:
			result_type = t_int
		case 2:
			result_type = t_int
		case 3:
			result_type = t_allocator
		}
	case Type_FixedCapacityDynamicArray:
		switch index {
		case 0:
			result_type = alloc_type_array(t.FixedCapacityDynamicArray.Elem, t.FixedCapacityDynamicArray.Capacity, nil)
		case 1:
			result_type = t_int
		}
	case Type_Map:
		init_map_internal_debug_types(t)
		switch index {
		case 0:
			result_type = get_struct_field_type(t_raw_map, 0)
		case 1:
			result_type = get_struct_field_type(t_raw_map, 1)
		case 2:
			result_type = get_struct_field_type(t_raw_map, 2)
		}
	case Type_Array:
		result_type = t.Array.Elem
	case Type_SoaPointer:
		switch index {
		case 0:
			result_type = alloc_type_pointer(t.SoaPointer.Elem)
		case 1:
			result_type = t_int
		}
	default:
		gb_assert_handler("Panic", "TODO: struct_ev type", "llvm_backend_utility_mem.go", 0, "%s, %d", type_to_string(s.Type), index)
	}

	gb_assert_handler("Assertion Failure", "result_type != nil", "llvm_backend_utility_mem.go", 0, "%s, %d", type_to_string(s.Type), index)

	index = lb_convert_struct_index(p.Module, t, index)

	var res lbValue
	res.Value = LLVMBuildExtractValue(p.Builder, s.Value, uint(index), "")
	res.Type = result_type
	return res
}

func lb_emit_deep_field_gep(p *lbProcedure, e lbValue, sel Selection) lbValue {
	gb_assert_handler("Assertion Failure", "len(sel.index) > 0", "llvm_backend_utility_mem.go", 0)
	type_ := type_deref(e.Type)

	for i := range sel.Index {
		index := sel.Index[i]
		if is_type_pointer(type_) {
			type_ = type_deref(type_)
			e = lb_emit_load(p, e)
		}
		type_ = core_type(type_)

		if type_.Kind == Type_SoaPointer {
			addr := lb_emit_struct_ep(p, e, 0)
			idx := lb_emit_struct_ep(p, e, 1)
			addr = lb_emit_load(p, addr)
			idx = lb_emit_load(p, idx)

			first_index := sel.Index[0]
			sub_sel := sel
			sub_sel.Index = sub_sel.Index[1:]

			arr := lb_emit_struct_ep(p, addr, first_index)

			t2 := base_type(type_deref(addr.Type))
			gb_assert_handler("Assertion Failure", "is_type_soa_struct(t2)", "llvm_backend_utility_mem.go", 0)

			if t2.Struct.SoaKind == StructSoaFixed {
				e = lb_emit_array_ep(p, arr, idx)
			} else {
				e = lb_emit_ptr_offset(p, lb_emit_load(p, arr), idx)
			}
			e.Type = alloc_type_multi_pointer_to_pointer(e.Type)

		} else if is_type_quaternion(type_) {
			e = lb_emit_struct_ep(p, e, index)
		} else if is_type_raw_union(type_) {
			type_ = get_struct_field_type(type_, index)
			gb_assert_handler("Assertion Failure", "is_type_pointer(e.type)", "llvm_backend_utility_mem.go", 0)
			e = lb_emit_transmute(p, e, alloc_type_pointer(type_))
		} else if is_type_struct(type_) {
			type_ = get_struct_field_type(type_, index)
			e = lb_emit_struct_ep(p, e, index)
		} else if type_.Kind == Type_Union {
			gb_assert_handler("Assertion Failure", "index == -1", "llvm_backend_utility_mem.go", 0)
			type_ = t_type_info_ptr
			e = lb_emit_struct_ep(p, e, index)
		} else if type_.Kind == Type_Tuple {
			type_ = type_.Tuple.Variables[index].Type
			e = lb_emit_struct_ep(p, e, index)
		} else if type_.Kind == Type_Basic {
			switch type_.Basic.Kind {
			case BasicAny:
				if index == 0 {
					type_ = t_rawptr
				} else if index == 1 {
					type_ = t_typeid
				}
				e = lb_emit_struct_ep(p, e, index)
			case BasicString:
				e = lb_emit_struct_ep(p, e, index)
			case BasicString16:
				e = lb_emit_struct_ep(p, e, index)
			default:
				gb_assert_handler("Panic", "un-gep-able type", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(type_))
			}
		} else if type_.Kind == Type_Slice {
			e = lb_emit_struct_ep(p, e, index)
		} else if type_.Kind == Type_DynamicArray {
			e = lb_emit_struct_ep(p, e, index)
		} else if type_.Kind == Type_Array {
			e = lb_emit_array_epi_proc(p, e, isize(index))
		} else if type_.Kind == Type_Map {
			e = lb_emit_struct_ep(p, e, index)
		} else {
			gb_assert_handler("Panic", "un-gep-able type", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(type_))
		}
	}

	return e
}

func lb_emit_deep_field_ev(p *lbProcedure, e lbValue, sel Selection) lbValue {
	ptr := lb_address_from_load_or_generate_local(p, e)
	res := lb_emit_deep_field_gep(p, ptr, sel)
	return lb_emit_load(p, res)
}

func lb_emit_array_ep(p *lbProcedure, s lbValue, index lbValue) lbValue {
	t := s.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(t)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(t))
	st := base_type(type_deref(t))
	gb_assert_handler("Assertion Failure", "is_type_array(st) || is_type_enumerated_array(st) || is_type_matrix(st)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(st))
	gb_assert_handler("Assertion Failure", "is_type_integer(core_type(index.type))", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(index.Type))

	indices := [2]LLVMValueRef{
		llvm_zero(p.Module),
		lb_emit_conv(p, index, t_int).Value,
	}

	ptr := base_array_type(st)
	var res lbValue

	if LLVMIsConstant(s.Value) != 0 && LLVMIsConstant(index.Value) != 0 {
		res.Value = LLVMConstGEP2(lb_type(p.Module, st), s.Value, indices[:], 2)
	} else {
		res.Value = LLVMBuildGEP2(p.Builder, lb_type(p.Module, st), s.Value, indices[:], 2, "")
	}
	res.Type = alloc_type_pointer(ptr)
	return res
}

func lb_emit_array_epi_proc(p *lbProcedure, s lbValue, index isize) lbValue {
	t := s.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(t)", "llvm_backend_utility_mem.go", 0)
	st := base_type(type_deref(t))

	gb_assert_handler("Assertion Failure", "0 <= index", "llvm_backend_utility_mem.go", 0)
	if is_type_fixed_capacity_dynamic_array(st) {
		data := lb_emit_struct_ep(p, s, 0)
		return lb_emit_epi(p, data, index)
	}

	gb_assert_handler("Assertion Failure", "is_type_array(st) || is_type_enumerated_array(st) || is_type_matrix(st)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(st))
	return lb_emit_epi(p, s, index)
}

func lb_emit_array_epi(m *lbModule, s lbValue, index isize) lbValue {
	t := s.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(t)", "llvm_backend_utility_mem.go", 0)
	st := base_type(type_deref(t))
	gb_assert_handler("Assertion Failure", "0 <= index", "llvm_backend_utility_mem.go", 0)
	if is_type_fixed_capacity_dynamic_array(st) {
		data := lb_emit_epi_module(m, s, 0)
		return lb_emit_epi_module(m, data, index)
	}

	gb_assert_handler("Assertion Failure", "is_type_array(st) || is_type_enumerated_array(st) || is_type_matrix(st)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(st))
	return lb_emit_epi_module(m, s, index)
}

func lb_emit_ptr_offset(p *lbProcedure, ptr lbValue, index lbValue) lbValue {
	index = lb_emit_conv(p, index, t_int)
	indices := [1]LLVMValueRef{index.Value}
	var res lbValue
	res.Type = ptr.Type
	llvm_type := lb_type(p.Module, type_deref(res.Type, true))

	if lb_is_const(ptr) && lb_is_const(index) {
		res.Value = LLVMConstGEP2(llvm_type, ptr.Value, indices[:], 1)
	} else {
		res.Value = LLVMBuildGEP2(p.Builder, llvm_type, ptr.Value, indices[:], 1, "")
	}
	return res
}

func lb_const_ptr_offset(m *lbModule, ptr lbValue, index lbValue) lbValue {
	indices := [1]LLVMValueRef{index.Value}
	var res lbValue
	res.Type = ptr.Type
	llvm_type := lb_type(m, type_deref(res.Type, true))

	gb_assert_handler("Assertion Failure", "lb_is_const(ptr) && lb_is_const(index)", "llvm_backend_utility_mem.go", 0)
	res.Value = LLVMConstGEP2(llvm_type, ptr.Value, indices[:], 1)
	return res
}

func lb_emit_matrix_epi(p *lbProcedure, s lbValue, row isize, column isize) lbValue {
	t := s.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(t)", "llvm_backend_utility_mem.go", 0)
	mt := base_type(type_deref(t))

	if !mt.Matrix.IsRowMajor {
		if column == 0 {
			gb_assert_handler("Assertion Failure", "is_type_matrix(mt) || is_type_array_like(mt)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(mt))
			return lb_emit_epi(p, s, row)
		} else if row == 0 && is_type_array_like(mt) {
			return lb_emit_epi(p, s, column)
		}
	}

	gb_assert_handler("Assertion Failure", "is_type_matrix(mt)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(mt))

	offset := matrix_indices_to_offset(mt, row, column)
	return lb_emit_epi(p, s, offset)
}

func lb_emit_matrix_ep(p *lbProcedure, s lbValue, row lbValue, column lbValue) lbValue {
	t := s.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(t)", "llvm_backend_utility_mem.go", 0)
	mt := base_type(type_deref(t))
	gb_assert_handler("Assertion Failure", "is_type_matrix(mt)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(mt))

	ptr := base_array_type(mt)

	stride_elems := lb_const_int(p.Module, t_int, u64(matrix_type_stride_in_elems(mt))).Value

	row = lb_emit_conv(p, row, t_int)
	column = lb_emit_conv(p, column, t_int)

	var index LLVMValueRef

	if mt.Matrix.IsRowMajor {
		index = LLVMBuildAdd(p.Builder, column.Value, LLVMBuildMul(p.Builder, row.Value, stride_elems, ""), "")
	} else {
		index = LLVMBuildAdd(p.Builder, row.Value, LLVMBuildMul(p.Builder, column.Value, stride_elems, ""), "")
	}

	indices := [2]LLVMValueRef{
		LLVMConstInt(lb_type(p.Module, t_int), 0, 0),
		index,
	}

	llvm_type := lb_type(p.Module, mt)
	var res lbValue
	if lb_is_const(s) {
		res.Value = LLVMConstGEP2(llvm_type, s.Value, indices[:], 2)
	} else {
		res.Value = LLVMBuildGEP2(p.Builder, llvm_type, s.Value, indices[:], 2, "")
	}
	res.Type = alloc_type_pointer(ptr)
	return res
}

func lb_emit_matrix_ev(p *lbProcedure, s lbValue, row isize, column isize) lbValue {
	t := s.Type
	mt := base_type(t)
	gb_assert_handler("Assertion Failure", "is_type_matrix(mt)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(mt))

	stride_elems := matrix_type_stride_in_elems(mt)

	var index isize = -1

	if mt.Matrix.IsRowMajor {
		index = column + (row * stride_elems)
	} else {
		index = row + (column * stride_elems)
	}

	var res lbValue
	res.Value = LLVMBuildExtractValue(p.Builder, s.Value, uint(index), "")
	res.Type = base_array_type(mt)
	return res
}

func lb_fill_slice(p *lbProcedure, slice lbAddr, base_elem lbValue, len_ lbValue) {
	t := lb_addr_type(slice)
	gb_assert_handler("Assertion Failure", "is_type_slice(t)", "llvm_backend_utility_mem.go", 0)
	ptr := lb_addr_get_ptr(p, slice)
	data := lb_emit_struct_ep(p, ptr, 0)
	if are_types_identical(type_deref(base_elem.Type, true), type_deref(type_deref(data.Type), true)) {
		base_elem = lb_emit_conv(p, base_elem, type_deref(data.Type))
	}
	lb_emit_store(p, data, base_elem)
	lb_emit_store(p, lb_emit_struct_ep(p, ptr, 1), len_)
}

func lb_fill_string(p *lbProcedure, str lbAddr, base_elem lbValue, len_ lbValue) {
	t := lb_addr_type(str)
	gb_assert_handler("Assertion Failure", "is_type_string(t)", "llvm_backend_utility_mem.go", 0)
	ptr := lb_addr_get_ptr(p, str)
	data := lb_emit_struct_ep(p, ptr, 0)
	if are_types_identical(type_deref(base_elem.Type, true), type_deref(type_deref(data.Type), true)) {
		base_elem = lb_emit_conv(p, base_elem, type_deref(data.Type))
	}
	lb_emit_store(p, data, base_elem)
	lb_emit_store(p, lb_emit_struct_ep(p, ptr, 1), len_)
}

func lb_emit_struct_iv(p *lbProcedure, agg lbValue, field lbValue, index i32) lbValue {
	t := base_type(agg.Type)
	mapped_index := lb_convert_struct_index(p.Module, t, index)
	var res lbValue
	res.Value = LLVMBuildInsertValue(p.Builder, agg.Value, field.Value, uint(mapped_index), "")
	res.Type = agg.Type
	return res
}

func lb_build_struct_value(p *lbProcedure, typ *Type, fields []lbValue, count isize) lbValue {
	llvm_type := lb_type(p.Module, typ)
	var agg lbValue
	agg.Value = LLVMConstNull(llvm_type)
	agg.Type = typ
	for i := isize(0); i < count; i++ {
		agg = lb_emit_struct_iv(p, agg, fields[i], i32(i))
	}
	return agg
}

func lb_make_slice_value(p *lbProcedure, slice_type *Type, elem lbValue, len_ lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_slice(slice_type)", "llvm_backend_utility_mem.go", 0)
	fields := [2]lbValue{elem, len_}
	return lb_build_struct_value(p, slice_type, fields[:], 2)
}

func lb_make_string_value(p *lbProcedure, string_type *Type, elem lbValue, len_ lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_string(string_type)", "llvm_backend_utility_mem.go", 0)
	fields := [2]lbValue{elem, len_}
	return lb_build_struct_value(p, string_type, fields[:], 2)
}

func lb_string_elem(p *lbProcedure, str lbValue) lbValue {
	t := base_type(str.Type)
	if t.Kind == Type_Basic && t.Basic.Kind == BasicString16 {
		return lb_emit_struct_ev(p, str, 0)
	}
	gb_assert_handler("Assertion Failure", "t->kind == Type_Basic && t->Basic.kind == BasicString", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, str, 0)
}

func lb_string_len(p *lbProcedure, str lbValue) lbValue {
	t := base_type(str.Type)
	if t.Kind == Type_Basic && t.Basic.Kind == BasicString16 {
		return lb_emit_struct_ev(p, str, 1)
	}
	gb_assert_handler("Assertion Failure", "t->kind == Type_Basic && t->Basic.kind == BasicString", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(t))
	return lb_emit_struct_ev(p, str, 1)
}

func lb_cstring_len(p *lbProcedure, value lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_cstring(value.type)", "llvm_backend_utility_mem.go", 0)
	args := make([]lbValue, 1)
	args[0] = lb_emit_conv(p, value, t_cstring)
	return lb_emit_runtime_call(p, "cstring_len", args)
}

func lb_cstring16_len(p *lbProcedure, value lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_cstring16(value.type)", "llvm_backend_utility_mem.go", 0)
	args := make([]lbValue, 1)
	args[0] = lb_emit_conv(p, value, t_cstring16)
	return lb_emit_runtime_call(p, "cstring16_len", args)
}

func lb_array_elem(p *lbProcedure, array_ptr lbValue) lbValue {
	t := type_deref(array_ptr.Type)
	gb_assert_handler("Assertion Failure", "is_type_array(t)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ep(p, array_ptr, 0)
}

func lb_slice_elem(p *lbProcedure, slice lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_slice(slice.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, slice, 0)
}

func lb_slice_len(p *lbProcedure, slice lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_slice(slice.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, slice, 1)
}

func lb_dynamic_array_elem(p *lbProcedure, da lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_dynamic_array(da.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, da, 0)
}

func lb_dynamic_array_len(p *lbProcedure, da lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_dynamic_array(da.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, da, 1)
}

func lb_dynamic_array_cap(p *lbProcedure, da lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_dynamic_array(da.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, da, 2)
}

func lb_dynamic_array_allocator(p *lbProcedure, da lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_dynamic_array(da.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, da, 3)
}

func lb_fixed_capacity_dynamic_array_len(p *lbProcedure, da lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_fixed_capacity_dynamic_array(da.type)", "llvm_backend_utility_mem.go", 0)
	return lb_emit_struct_ev(p, da, 1)
}

func lb_map_len(p *lbProcedure, value lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_map(value.type) || are_types_identical(value.type, t_raw_map)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(value.Type))
	len_ := lb_emit_struct_ev(p, value, 1)
	return lb_emit_conv(p, len_, t_int)
}

func lb_map_len_ptr(p *lbProcedure, map_ptr lbValue) lbValue {
	type_ := map_ptr.Type
	gb_assert_handler("Assertion Failure", "is_type_pointer(type_)", "llvm_backend_utility_mem.go", 0)
	type_ = type_deref(type_)
	gb_assert_handler("Assertion Failure", "is_type_map(type_) || are_types_identical(type_, t_raw_map)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(type_))
	return lb_emit_struct_ep(p, map_ptr, 1)
}

func lb_map_cap(p *lbProcedure, value lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_map(value.type) || are_types_identical(value.type, t_raw_map)", "llvm_backend_utility_mem.go", 0, "%s", type_to_string(value.Type))
	zero := lb_const_int(p.Module, t_uintptr, 0)
	one := lb_const_int(p.Module, t_uintptr, 1)

	mask := lb_const_int(p.Module, t_uintptr, MAP_CACHE_LINE_SIZE-1)

	data := lb_emit_struct_ev(p, value, 0)
	log2_cap := lb_emit_arith(p, Token_And, data, mask, t_uintptr)
	cap_ := lb_emit_arith(p, Token_Shl, one, log2_cap, t_uintptr)
	cmp := lb_emit_comp(p, Token_CmpEq, data, zero)
	return lb_emit_conv(p, lb_emit_select(p, cmp, zero, cap_), t_int)
}

func lb_map_data_uintptr(p *lbProcedure, value lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_map(value.type) || are_types_identical(value.type, t_raw_map)", "llvm_backend_utility_mem.go", 0)
	data := lb_emit_struct_ev(p, value, 0)
	var mask_value u64
	if buildContext.PtrSize == 4 {
		mask_value = 0xffffffff & ^u64(MAP_CACHE_LINE_SIZE-1)
	} else {
		mask_value = 0xffffffffffffffff & ^u64(MAP_CACHE_LINE_SIZE-1)
	}
	mask := lb_const_int(p.Module, t_uintptr, mask_value)
	return lb_emit_arith(p, Token_And, data, mask, t_uintptr)
}

func lb_soa_struct_len(p *lbProcedure, value lbValue) lbValue {
	t := base_type(value.Type)
	is_ptr := false
	if is_type_pointer(t) {
		is_ptr = true
		t = base_type(type_deref(t))
	}

	if t.Struct.SoaKind == StructSoaFixed {
		return lb_const_int(p.Module, t_int, u64(t.Struct.SoaCount))
	}

	gb_assert_handler("Assertion Failure", "t->Struct.soa_kind == StructSoaSlice || t->Struct.soa_kind == StructSoaDynamic", "llvm_backend_utility_mem.go", 0)

	var n isize = 0
	elem := base_type(t.Struct.SoaElem)
	if elem.Kind == Type_Struct {
		n = isize(len(elem.Struct.Fields))
	} else if elem.Kind == Type_Array {
		n = isize(elem.Array.Count)
	} else {
		gb_assert_handler("Panic", "Unreachable", "llvm_backend_utility_mem.go", 0)
	}

	if is_ptr {
		v := lb_emit_struct_ep(p, value, i32(n))
		return lb_emit_load(p, v)
	}
	return lb_emit_struct_ev(p, value, i32(n))
}

func lb_soa_struct_cap(p *lbProcedure, value lbValue) lbValue {
	t := base_type(value.Type)

	is_ptr := false
	if is_type_pointer(t) {
		is_ptr = true
		t = base_type(type_deref(t))
	}

	if t.Struct.SoaKind == StructSoaFixed {
		return lb_const_int(p.Module, t_int, u64(t.Struct.SoaCount))
	}

	gb_assert_handler("Assertion Failure", "t->Struct.soa_kind == StructSoaDynamic", "llvm_backend_utility_mem.go", 0)

	var n isize = 0
	elem := base_type(t.Struct.SoaElem)
	if elem.Kind == Type_Struct {
		n = isize(len(elem.Struct.Fields)) + 1
	} else if elem.Kind == Type_Array {
		n = isize(elem.Array.Count) + 1
	} else {
		gb_assert_handler("Panic", "Unreachable", "llvm_backend_utility_mem.go", 0)
	}

	if is_ptr {
		v := lb_emit_struct_ep(p, value, i32(n))
		return lb_emit_load(p, v)
	}
	return lb_emit_struct_ev(p, value, i32(n))
}

func lb_emit_mul_add(p *lbProcedure, a lbValue, b lbValue, c lbValue, t *Type) lbValue {
	m := p.Module

	a = lb_emit_conv(p, a, t)
	b = lb_emit_conv(p, b, t)
	c = lb_emit_conv(p, c, t)

	is_possible := !is_type_different_to_arch_endianness(t) && is_type_float(t)

	if is_possible {
		switch buildContext.Metrics.Arch {
		case TargetArchAmd64:
			if type_size_of(t) == 2 || !check_target_feature_is_enabled(String{Data: strData("fma"), Len: 3}, nil) {
				is_possible = false
			}
		case TargetArchArm64:
		case TargetArchI386, TargetArchWasm32, TargetArchWasm64p32:
			is_possible = false
		}
	}

	if is_possible {
		name := "llvm.fma"
		types := [1]LLVMTypeRef{lb_type(m, t)}
		values := [3]LLVMValueRef{a.Value, b.Value, c.Value}
		call := lb_call_intrinsic(p, name, values[:], 3, types[:], 1)
		return lbValue{Value: call, Type: t}
	} else {
		x := lb_emit_arith(p, Token_Mul, a, b, t)
		y := lb_emit_arith(p, Token_Add, x, c, t)
		return y
	}
}

func lb_soa_zip(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	gb_assert_handler("Assertion Failure", "ce->args.count > 0", "llvm_backend_utility_mem.go", 0)

	slices := make([]lbValue, len(ce.Args))
	for i, arg := range ce.Args {
		if arg.Kind == Ast_FieldValue {
			arg = arg.FieldValue.Value
		}
		slices[i] = lb_build_expr(p, arg)
	}

	len_ := lb_slice_len(p, slices[0])
	for i := 1; i < len(slices); i++ {
		other_len := lb_slice_len(p, slices[i])
		len_ = lb_emit_min(p, t_int, len_, other_len)
	}

	gb_assert_handler("Assertion Failure", "is_type_soa_struct(tv.type)", "llvm_backend_utility_mem.go", 0)
	res := lb_add_local_generated(p, tv.Type, true)
	for i, slice := range slices {
		src := lb_slice_elem(p, slice)
		src = lb_emit_conv(p, src, alloc_type_pointer_to_multi_pointer(src.Type))
		dst := lb_emit_struct_ep(p, res.Addr, i32(i))
		lb_emit_store(p, dst, src)
	}
	len_dst := lb_emit_struct_ep(p, res.Addr, i32(len(slices)))
	lb_emit_store(p, len_dst, len_)

	return lb_addr_load(p, res)
}

func lb_soa_unzip(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	gb_assert_handler("Assertion Failure", "ce->args.count == 1", "llvm_backend_utility_mem.go", 0)

	arg := lb_build_expr(p, ce.Args[0])
	t := base_type(arg.Type)
	gb_assert_handler("Assertion Failure", "is_type_soa_struct(t) && t->Struct.soa_kind == StructSoaSlice", "llvm_backend_utility_mem.go", 0)

	len_ := lb_soa_struct_len(p, arg)

	res := lb_add_local_generated(p, tv.Type, true)
	if is_type_tuple(tv.Type) {
		rp := lb_addr_get_ptr(p, res)
		for i := i32(0); i < i32(len(t.Struct.Fields)-1); i++ {
			ptr := lb_emit_struct_ev(p, arg, i)
			dst := lb_addr(lb_emit_struct_ep(p, rp, i))
			lb_fill_slice(p, dst, ptr, len_)
		}
	} else {
		gb_assert_handler("Assertion Failure", "is_type_slice(tv.type)", "llvm_backend_utility_mem.go", 0)
		ptr := lb_emit_struct_ev(p, arg, 0)
		lb_fill_slice(p, res, ptr, len_)
	}

	return lb_addr_load(p, res)
}

func get_struct_field_type(t *Type, index i32) *Type {
	t = base_type(type_deref(t))
	gb_assert_handler("Assertion Failure", "t.Kind == Type_Struct", "llvm_backend_utility_mem.go", 0)
	return t.Struct.Fields[index].Type
}
