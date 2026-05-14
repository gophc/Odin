package cmd

import "unsafe"

func lb_simple_compare_hash(p *lbProcedure, typ *Type, data lbValue, seed lbValue) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_simple_compare(type)", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0, "%s", type_to_string(typ))
	args := make([]lbValue, 3)
	args[0] = data
	args[1] = seed
	args[2] = lb_const_int(p.Module, t_int, u64(type_size_of(typ)))
	return lb_emit_runtime_call(p, "default_hasher", args)
}

func lb_add_callsite_force_inline(p *lbProcedure, ret_value lbValue) {
	LLVMAddCallSiteAttribute(ret_value.Value, LLVMAttributeIndex_FunctionIndex, lb_create_enum_attribute(p.Module.Ctx, "alwaysinline"))
}

func lb_hasher_proc_for_type(m *lbModule, typ *Type) lbValue {
	typ = core_type(typ)
	gb_assert_handler("Assertion Failure", "is_type_comparable(type)", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0, "%s", type_to_string(typ))
	pt := alloc_type_pointer(typ)
	proc_name := lb_internal_gen_name_from_type("__$hasher", typ)
	if found, ok := m.GenProcs[proc_name]; ok {
		gb_assert_handler("Assertion Failure", "found != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		return lbValue{Value: found.Value, Type: found.Type}
	}
	p := lb_create_dummy_procedure(m, proc_name, t_hasher_proc)
	m.GenProcs[proc_name] = p
	lb_begin_procedure_body(p)
	defer lb_end_procedure_body(p)
	LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	lb_add_attribute_to_proc(p, "nounwind")
	x := LLVMGetParam(p.Value, 0)
	y := LLVMGetParam(p.Value, 1)
	data := lbValue{Value: x, Type: t_rawptr}
	seed := lbValue{Value: y, Type: t_uintptr}
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+0), "nonnull")
	if is_type_simple_compare(typ) {
		res := lb_simple_compare_hash(p, typ, data, seed)
		lb_add_callsite_force_inline(p, res)
		LLVMBuildRet(p.Builder, res.Value)
		return lbValue{Value: p.Value, Type: p.Type}
	}
	switch typ.Kind {
	case Type_Struct:
		type_set_offsets(typ)
		data = lb_emit_conv(p, data, t_u8_ptr)
		args := make([]lbValue, 2)
		for i := isize(0); i < isize(len(typ.Struct.Fields)); i++ {
			gb_assert_handler("Assertion Failure", "type.Struct.Offsets != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
			offset := typ.Struct.Offsets[i]
			field := typ.Struct.Fields[i]
			field_hasher := lb_hasher_proc_for_type(m, field.Type)
			ptr := lb_emit_ptr_offset(p, data, lb_const_int(m, t_uintptr, u64(offset)))
			args[0] = ptr
			args[1] = seed
			seed = lb_emit_call(p, field_hasher, args)
		}
		LLVMBuildRet(p.Builder, seed.Value)
	case Type_Union:
		args := make([]lbValue, 2)
		if is_type_union_maybe_pointer(typ) {
			v := typ.Union.Variants[0]
			variant_hasher := lb_hasher_proc_for_type(m, v)
			args[0] = data
			args[1] = seed
			res := lb_emit_call(p, variant_hasher, args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else {
			end_block := lb_create_block(p, "bend")
			data = lb_emit_conv(p, data, pt)
			tag_ptr := lb_emit_union_tag_ptr(p, data)
			tag := lb_emit_load(p, tag_ptr)
			v_switch := LLVMBuildSwitch(p.Builder, tag.Value, end_block.Block, uint(len(typ.Union.Variants)))
			for _, v := range typ.Union.Variants {
				case_block := lb_create_block(p, "bcase")
				lb_start_block(p, case_block)
				case_tag := lb_const_union_tag(p.Module, typ, v)
				variant_hasher := lb_hasher_proc_for_type(m, v)
				args[0] = data
				args[1] = seed
				res := lb_emit_call(p, variant_hasher, args)
				LLVMBuildRet(p.Builder, res.Value)
				LLVMAddCase(v_switch, case_tag.Value, case_block.Block)
			}
			lb_start_block(p, end_block)
			LLVMBuildRet(p.Builder, seed.Value)
		}
	case Type_Array:
		pres := lb_add_local_generated(p, t_uintptr, false)
		lb_addr_store(p, pres, seed)
		args := make([]lbValue, 2)
		elem_hasher := lb_hasher_proc_for_type(m, typ.Array.Elem)
		loop_data := lb_loop_start(p, isize(typ.Array.Count), t_i32)
		data = lb_emit_conv(p, data, pt)
		ptr := lb_emit_array_ep(p, data, loop_data.Idx)
		args[0] = ptr
		args[1] = lb_addr_load(p, pres)
		new_seed := lb_emit_call(p, elem_hasher, args)
		lb_addr_store(p, pres, new_seed)
		lb_loop_end(p, loop_data)
		res := lb_addr_load(p, pres)
		LLVMBuildRet(p.Builder, res.Value)
	case Type_EnumeratedArray:
		pres := lb_add_local_generated(p, t_uintptr, false)
		lb_addr_store(p, pres, seed)
		args := make([]lbValue, 2)
		elem_hasher := lb_hasher_proc_for_type(m, typ.EnumeratedArray.Elem)
		loop_data := lb_loop_start(p, isize(typ.EnumeratedArray.Count), t_i32)
		data = lb_emit_conv(p, data, pt)
		ptr := lb_emit_array_ep(p, data, loop_data.Idx)
		args[0] = ptr
		args[1] = lb_addr_load(p, pres)
		new_seed := lb_emit_call(p, elem_hasher, args)
		lb_addr_store(p, pres, new_seed)
		lb_loop_end(p, loop_data)
		vres := lb_addr_load(p, pres)
		LLVMBuildRet(p.Builder, vres.Value)
	default:
		if is_type_cstring(typ) {
			args := make([]lbValue, 2)
			args[0] = data
			args[1] = seed
			res := lb_emit_runtime_call(p, "default_hasher_cstring", args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else if is_type_string(typ) {
			args := make([]lbValue, 2)
			args[0] = data
			args[1] = seed
			res := lb_emit_runtime_call(p, "default_hasher_string", args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else if is_type_float(typ) {
			ptr := lb_emit_conv(p, data, pt)
			v := lb_emit_load(p, ptr)
			v = lb_emit_conv(p, v, t_f64)
			args := make([]lbValue, 2)
			args[0] = v
			args[1] = seed
			res := lb_emit_runtime_call(p, "default_hasher_f64", args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else if is_type_complex(typ) {
			ptr := lb_emit_conv(p, data, pt)
			xp := lb_emit_struct_ep(p, ptr, 0)
			yp := lb_emit_struct_ep(p, ptr, 1)
			x := lb_emit_conv(p, lb_emit_load(p, xp), t_f64)
			y := lb_emit_conv(p, lb_emit_load(p, yp), t_f64)
			args := make([]lbValue, 3)
			args[0] = x
			args[1] = y
			args[2] = seed
			res := lb_emit_runtime_call(p, "default_hasher_complex128", args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else if is_type_quaternion(typ) {
			ptr := lb_emit_conv(p, data, pt)
			xp := lb_emit_struct_ep(p, ptr, 0)
			yp := lb_emit_struct_ep(p, ptr, 1)
			zp := lb_emit_struct_ep(p, ptr, 2)
			wp := lb_emit_struct_ep(p, ptr, 3)
			x := lb_emit_conv(p, lb_emit_load(p, xp), t_f64)
			y := lb_emit_conv(p, lb_emit_load(p, yp), t_f64)
			z := lb_emit_conv(p, lb_emit_load(p, zp), t_f64)
			w := lb_emit_conv(p, lb_emit_load(p, wp), t_f64)
			args := make([]lbValue, 5)
			args[0] = x
			args[1] = y
			args[2] = z
			args[3] = w
			args[4] = seed
			res := lb_emit_runtime_call(p, "default_hasher_quaternion256", args)
			lb_add_callsite_force_inline(p, res)
			LLVMBuildRet(p.Builder, res.Value)
		} else {
			gb_assert_handler("Panic", 0, "llvm_backend_map_verify_global_init_startup_proc_types.go", 0, "Unhandled type for hasher: %s", type_to_string(typ))
		}
	}
	return lbValue{Value: p.Value, Type: p.Type}
}

func lb_map_get_proc_for_type(m *lbModule, typ *Type) lbValue {
	gb_assert_handler("Assertion Failure", "!buildContext.DynamicMapCalls", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	typ = base_type(typ)
	gb_assert_handler("Assertion Failure", "typ.Kind == Type_Map", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	proc_name := lb_internal_gen_name_from_type("__$map_get", typ)
	if found, ok := m.GenProcs[proc_name]; ok {
		gb_assert_handler("Assertion Failure", "found != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		return lbValue{Value: found.Value, Type: found.Type}
	}
	p := lb_create_dummy_procedure(m, proc_name, t_map_get_proc)
	m.GenProcs[proc_name] = p
	p.InternalGenType = typ
	lb_begin_procedure_body(p)
	defer lb_end_procedure_body(p)
	LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	lb_add_attribute_to_proc(p, "nounwind")
	if buildContext.ODIN_DEBUG {
		lb_add_attribute_to_proc(p, "noinline")
	}
	x := LLVMGetParam(p.Value, 0)
	y := LLVMGetParam(p.Value, 1)
	z := LLVMGetParam(p.Value, 2)
	map_ptr := lbValue{Value: x, Type: t_rawptr}
	h := lbValue{Value: y, Type: t_uintptr}
	key_ptr := lbValue{Value: z, Type: t_rawptr}
	LLVMSetValueName2(h.Value, "hash", uint(len("hash")))
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+0), "nonnull")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+0), "readonly")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+2), "nonnull")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+2), "readonly")
	loop_block := lb_create_block(p, "loop")
	hash_block := lb_create_block(p, "hash")
	probe_block := lb_create_block(p, "probe")
	increment_block := lb_create_block(p, "increment")
	hash_compare_block := lb_create_block(p, "hash_compare")
	key_compare_block := lb_create_block(p, "key_compare")
	value_block := lb_create_block(p, "value")
	nil_block := lb_create_block(p, "nil")
	map_ptr = lb_emit_conv(p, map_ptr, t_raw_map_ptr)
	LLVMSetValueName2(map_ptr.Value, "map_ptr", uint(len("map_ptr")))
	map_val := lb_emit_load(p, map_ptr)
	LLVMSetValueName2(map_val.Value, "map", uint(len("map")))
	length := lb_map_len(p, map_val)
	LLVMSetValueName2(length.Value, "length", uint(len("length")))
	lb_emit_if(p, lb_emit_comp(p, Token_CmpEq, length, lb_const_nil(m, t_int)), nil_block, hash_block)
	lb_start_block(p, hash_block)
	key_ptr = lb_emit_conv(p, key_ptr, alloc_type_pointer(typ.Map.Key))
	LLVMSetValueName2(key_ptr.Value, "key_ptr", uint(len("key_ptr")))
	key := lb_emit_load(p, key_ptr)
	LLVMSetValueName2(key.Value, "key", uint(len("key")))
	pos := lb_add_local_generated(p, t_uintptr, false)
	distance := lb_add_local_generated(p, t_uintptr, true)
	LLVMSetValueName2(pos.Addr.Value, "pos", uint(len("pos")))
	LLVMSetValueName2(distance.Addr.Value, "distance", uint(len("distance")))
	capacity := lb_map_cap(p, map_val)
	LLVMSetValueName2(capacity.Value, "capacity", uint(len("capacity")))
	cap_minus_1 := lb_emit_arith(p, Token_Sub, capacity, lb_const_int(m, t_int, 1), t_int)
	mask := lb_emit_conv(p, cap_minus_1, t_uintptr)
	LLVMSetValueName2(mask.Value, "mask", uint(len("mask")))
	{
		the_pos := lb_emit_arith(p, Token_And, h, mask, t_uintptr)
		the_pos = lb_emit_conv(p, the_pos, t_uintptr)
		lb_addr_store(p, pos, the_pos)
	}
	zero_uintptr := lb_const_int(m, t_uintptr, 0)
	one_uintptr := lb_const_int(m, t_uintptr, 1)
	ks := lb_map_data_uintptr(p, map_val)
	vs := lb_map_cell_index_static(p, typ.Map.Value, ks, capacity)
	hs := lb_map_cell_index_static(p, typ.Map.Value, vs, capacity)
	ks = lb_emit_conv(p, ks, alloc_type_pointer(typ.Map.Key))
	vs = lb_emit_conv(p, vs, alloc_type_pointer(typ.Map.Value))
	hs = lb_emit_conv(p, hs, alloc_type_pointer(t_uintptr))
	LLVMSetValueName2(ks.Value, "ks", uint(len("ks")))
	LLVMSetValueName2(vs.Value, "vs", uint(len("vs")))
	LLVMSetValueName2(hs.Value, "hs", uint(len("hs")))
	lb_emit_jump(p, loop_block)
	lb_start_block(p, loop_block)
	element_hash := lb_emit_load(p, lb_emit_ptr_offset(p, hs, lb_addr_load(p, pos)))
	LLVMSetValueName2(element_hash.Value, "element_hash", uint(len("element_hash")))
	{
		lb_emit_if(p, lb_emit_comp(p, Token_CmpEq, element_hash, zero_uintptr), nil_block, probe_block)
	}
	lb_start_block(p, probe_block)
	{
		probe_distance := lb_emit_arith(p, Token_And, h, mask, t_uintptr)
		probe_distance = lb_emit_conv(p, probe_distance, t_uintptr)
		cap_val := lb_emit_conv(p, capacity, t_uintptr)
		base := lb_emit_arith(p, Token_Add, lb_addr_load(p, pos), cap_val, t_uintptr)
		probe_distance = lb_emit_arith(p, Token_Sub, base, probe_distance, t_uintptr)
		probe_distance = lb_emit_arith(p, Token_And, probe_distance, mask, t_uintptr)
		LLVMSetValueName2(probe_distance.Value, "probe_distance", uint(len("probe_distance")))
		cond := lb_emit_comp(p, Token_Gt, lb_addr_load(p, distance), probe_distance)
		lb_emit_if(p, cond, nil_block, hash_compare_block)
	}
	lb_start_block(p, hash_compare_block)
	{
		lb_emit_if(p, lb_emit_comp(p, Token_CmpEq, element_hash, h), key_compare_block, increment_block)
	}
	lb_start_block(p, key_compare_block)
	{
		element_key := lb_map_cell_index_static(p, typ.Map.Key, ks, lb_addr_load(p, pos))
		element_key = lb_emit_conv(p, element_key, ks.Type)
		LLVMSetValueName2(element_key.Value, "element_key_ptr", uint(len("element_key_ptr")))
		cond := lb_emit_comp(p, Token_CmpEq, lb_emit_load(p, element_key), key)
		lb_emit_if(p, cond, value_block, increment_block)
	}
	lb_start_block(p, value_block)
	{
		element_value := lb_map_cell_index_static(p, typ.Map.Value, vs, lb_addr_load(p, pos))
		LLVMSetValueName2(element_value.Value, "element_value_ptr", uint(len("element_value_ptr")))
		element_value = lb_emit_conv(p, element_value, t_rawptr)
		LLVMBuildRet(p.Builder, element_value.Value)
	}
	lb_start_block(p, increment_block)
	{
		pp := lb_addr_load(p, pos)
		pp = lb_emit_arith(p, Token_Add, pp, one_uintptr, t_uintptr)
		pp = lb_emit_arith(p, Token_And, pp, mask, t_uintptr)
		lb_addr_store(p, pos, pp)
		lb_emit_increment(p, distance.Addr)
	}
	lb_emit_jump(p, loop_block)
	lb_start_block(p, nil_block)
	{
		res := lb_const_nil(m, t_rawptr)
		LLVMBuildRet(p.Builder, res.Value)
	}
	return lbValue{Value: p.Value, Type: p.Type}
}

func lb_map_set_proc_for_type(m *lbModule, typ *Type) lbValue {
	gb_assert_handler("Assertion Failure", "!buildContext.DynamicMapCalls", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	typ = base_type(typ)
	gb_assert_handler("Assertion Failure", "typ.Kind == Type_Map", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	proc_name := lb_internal_gen_name_from_type("__$map_set", typ)
	if found, ok := m.GenProcs[proc_name]; ok {
		gb_assert_handler("Assertion Failure", "found != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		return lbValue{Value: found.Value, Type: found.Type}
	}
	p := lb_create_dummy_procedure(m, proc_name, t_map_set_proc)
	m.GenProcs[proc_name] = p
	lb_begin_procedure_body(p)
	defer lb_end_procedure_body(p)
	LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	lb_add_attribute_to_proc(p, "nounwind")
	if buildContext.ODIN_DEBUG {
		lb_add_attribute_to_proc(p, "noinline")
	}
	map_ptr := lbValue{Value: LLVMGetParam(p.Value, 0), Type: t_rawptr}
	hash_param := lbValue{Value: LLVMGetParam(p.Value, 1), Type: t_uintptr}
	key_ptr := lbValue{Value: LLVMGetParam(p.Value, 2), Type: t_rawptr}
	value_ptr := lbValue{Value: LLVMGetParam(p.Value, 3), Type: t_rawptr}
	location_ptr := lbValue{Value: LLVMGetParam(p.Value, 4), Type: t_source_code_location_ptr}
	map_ptr = lb_emit_conv(p, map_ptr, alloc_type_pointer(typ))
	key_ptr = lb_emit_conv(p, key_ptr, alloc_type_pointer(typ.Map.Key))
	LLVMSetValueName2(map_ptr.Value, "map_ptr", uint(len("map_ptr")))
	LLVMSetValueName2(hash_param.Value, "hash_param", uint(len("hash_param")))
	LLVMSetValueName2(key_ptr.Value, "key_ptr", uint(len("key_ptr")))
	LLVMSetValueName2(value_ptr.Value, "value_ptr", uint(len("value_ptr")))
	LLVMSetValueName2(location_ptr.Value, "location", uint(len("location")))
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+0), "nonnull")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+0), "noalias")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+2), "nonnull")
	if !are_types_identical(typ.Map.Key, typ.Map.Value) {
		lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+2), "noalias")
	}
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+2), "readonly")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+3), "nonnull")
	if !are_types_identical(typ.Map.Key, typ.Map.Value) {
		lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+3), "noalias")
	}
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+3), "readonly")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+4), "nonnull")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+4), "noalias")
	lb_add_proc_attribute_at_index(p, isize(LLVMAttributeIndex_FirstArgIndex+4), "readonly")
	hash_addr := lb_add_local_generated(p, t_uintptr, false)
	lb_addr_store(p, hash_addr, hash_param)
	LLVMSetValueName2(hash_addr.Addr.Value, "hash", uint(len("hash")))
	var found_ptr lbValue
	{
		map_get_proc := lb_map_get_proc_for_type(m, typ)
		args := make([]lbValue, 3)
		args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
		args[1] = lb_addr_load(p, hash_addr)
		args[2] = key_ptr
		found_ptr = lb_emit_call(p, map_get_proc, args)
	}
	LLVMSetValueName2(found_ptr.Value, "found_ptr", uint(len("found_ptr")))
	found_block := lb_create_block(p, "found")
	check_grow_block := lb_create_block(p, "check-grow")
	grow_fail_block := lb_create_block(p, "grow-fail")
	insert_block := lb_create_block(p, "insert")
	check_has_grown_block := lb_create_block(p, "check-has-grown")
	rehash_block := lb_create_block(p, "rehash")
	lb_emit_if(p, lb_emit_comp_against_nil(p, Token_NotEq, found_ptr), found_block, check_grow_block)
	lb_start_block(p, found_block)
	{
		lb_mem_copy_non_overlapping(p, found_ptr, value_ptr, lb_const_int(m, t_int, u64(type_size_of(typ.Map.Value))))
		LLVMBuildRet(p.Builder, lb_emit_conv(p, found_ptr, t_rawptr).Value)
	}
	lb_start_block(p, check_grow_block)
	map_info := lb_gen_map_info_ptr(p.Module, typ)
	LLVMSetValueName2(map_info.Value, "map_info", uint(len("map_info")))
	{
		args := make([]lbValue, 3)
		args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
		args[1] = map_info
		args[2] = lb_emit_load(p, location_ptr)
		grow_err_and_has_grown := lb_emit_runtime_call(p, "__dynamic_map_check_grow", args)
		grow_err := lb_emit_struct_ev(p, grow_err_and_has_grown, 0)
		has_grown := lb_emit_struct_ev(p, grow_err_and_has_grown, 1)
		LLVMSetValueName2(grow_err.Value, "grow_err", uint(len("grow_err")))
		LLVMSetValueName2(has_grown.Value, "has_grown", uint(len("has_grown")))
		lb_emit_if(p, lb_emit_comp_against_nil(p, Token_NotEq, grow_err), grow_fail_block, check_has_grown_block)
		lb_start_block(p, grow_fail_block)
		LLVMBuildRet(p.Builder, LLVMConstNull(lb_type(p.Module, t_rawptr)))
		lb_start_block(p, check_has_grown_block)
		lb_emit_if(p, has_grown, rehash_block, insert_block)
		lb_start_block(p, rehash_block)
		key := lb_emit_load(p, key_ptr)
		new_hash := lb_gen_map_key_hash(p, map_ptr, key)
		LLVMSetValueName2(new_hash.Value, "new_hash", uint(len("new_hash")))
		lb_addr_store(p, hash_addr, new_hash)
		lb_emit_jump(p, insert_block)
	}
	lb_start_block(p, insert_block)
	{
		args := make([]lbValue, 5)
		args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
		args[1] = map_info
		args[2] = lb_addr_load(p, hash_addr)
		args[3] = lb_emit_conv(p, key_ptr, t_uintptr)
		args[4] = lb_emit_conv(p, value_ptr, t_uintptr)
		result := lb_emit_runtime_call(p, "map_insert_hash_dynamic", args)
		lb_emit_increment(p, lb_map_len_ptr(p, map_ptr))
		LLVMBuildRet(p.Builder, lb_emit_conv(p, result, t_rawptr).Value)
	}
	return lbValue{Value: p.Value, Type: p.Type}
}

func lb_gen_map_cell_info_ptr(m *lbModule, typ *Type) lbValue {
	key := uint64(uintptr(unsafe.Pointer(typ)))
	if found, ok := m.MapCellInfoMap[key]; ok {
		return found.Addr
	}
	var size, len_ int64
	map_cell_size_and_len(typ, &size, &len_)
	const_values := make([]LLVMValueRef, 4)
	const_values[0] = lb_const_int(m, t_uintptr, u64(type_size_of(typ))).Value
	const_values[1] = lb_const_int(m, t_uintptr, u64(type_align_of(typ))).Value
	const_values[2] = lb_const_int(m, t_uintptr, u64(size)).Value
	const_values[3] = lb_const_int(m, t_uintptr, u64(len_)).Value
	llvm_res := llvm_const_named_struct(m, t_map_cell_info, const_values, isize(len(const_values)))
	res := lbValue{Value: llvm_res, Type: t_map_cell_info}
	addr := lb_add_global_generated_with_name(m, t_map_cell_info, res, lb_internal_gen_name_from_type("ggv$map_cell_info", typ))
	lb_make_global_private_const(addr)
	m.MapCellInfoMap[key] = addr
	return addr.Addr
}

func lb_gen_map_info_ptr(m *lbModule, map_type *Type) lbValue {
	map_type = base_type(map_type)
	gb_assert_handler("Assertion Failure", "map_type.Kind == Type_Map", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	key := uint64(uintptr(unsafe.Pointer(map_type)))
	if found, ok := m.MapInfoMap[key]; ok {
		return found.Addr
	}
	gb_assert_handler("Assertion Failure", "t_map_info != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	gb_assert_handler("Assertion Failure", "t_map_cell_info != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	key_cell_info := lb_gen_map_cell_info_ptr(m, map_type.Map.Key).Value
	value_cell_info := lb_gen_map_cell_info_ptr(m, map_type.Map.Value).Value
	const_values := make([]LLVMValueRef, 4)
	const_values[0] = key_cell_info
	const_values[1] = value_cell_info
	const_values[2] = lb_hasher_proc_for_type(m, map_type.Map.Key).Value
	const_values[3] = lb_equal_proc_for_type(m, map_type.Map.Key).Value
	llvm_res := llvm_const_named_struct(m, t_map_info, const_values, isize(len(const_values)))
	res := lbValue{Value: llvm_res, Type: t_map_info}
	addr := lb_add_global_generated_with_name(m, t_map_info, res, lb_internal_gen_name_from_type("ggv$map_info", map_type))
	lb_make_global_private_const(addr)
	m.MapInfoMap[key] = addr
	return addr.Addr
}

func lb_const_hash(m *lbModule, key lbValue, key_type *Type) lbValue {
	return lbValue{}
}

func lb_gen_map_key_hash(p *lbProcedure, map_ptr lbValue, key lbValue, key_ptr ...*lbValue) lbValue {
	key_type := base_type(type_deref(map_ptr.Type)).Map.Key
	real_key := lb_emit_conv(p, key, key_type)
	kp := lb_address_from_load_or_generate_local(p, real_key)
	kp = lb_emit_conv(p, kp, t_rawptr)
	if len(key_ptr) > 0 && key_ptr[0] != nil {
		*key_ptr[0] = kp
	}
	hashed_key := lb_const_hash(p.Module, real_key, key_type)
	if hashed_key.Value == 0 {
		hasher := lb_hasher_proc_for_type(p.Module, key_type)
		var seed lbValue
		{
			args := make([]lbValue, 1)
			args[0] = lb_map_data_uintptr(p, lb_emit_load(p, map_ptr))
			seed = lb_emit_runtime_call(p, "map_seed_from_map_data", args)
		}
		args := make([]lbValue, 2)
		args[0] = kp
		args[1] = seed
		hashed_key = lb_emit_call(p, hasher, args)
	}
	return hashed_key
}

func lb_internal_dynamic_map_get_ptr(p *lbProcedure, map_ptr lbValue, key lbValue) lbValue {
	map_type := base_type(type_deref(map_ptr.Type))
	gb_assert_handler("Assertion Failure", "map_type.Kind == Type_Map", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	var ptr lbValue
	var kp lbValue
	hash := lb_gen_map_key_hash(p, map_ptr, key, &kp)
	if buildContext.DynamicMapCalls {
		args := make([]lbValue, 4)
		args[0] = lb_emit_transmute(p, map_ptr, t_raw_map_ptr)
		args[1] = lb_gen_map_info_ptr(p.Module, map_type)
		args[2] = hash
		args[3] = kp
		ptr = lb_emit_runtime_call(p, "__dynamic_map_get", args)
	} else {
		map_get_proc := lb_map_get_proc_for_type(p.Module, map_type)
		args := make([]lbValue, 3)
		args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
		args[1] = hash
		args[2] = kp
		ptr = lb_emit_call(p, map_get_proc, args)
	}
	return lb_emit_conv(p, ptr, alloc_type_pointer(map_type.Map.Value))
}

func lb_internal_dynamic_map_set(p *lbProcedure, map_ptr lbValue, map_type *Type, map_key lbValue, map_value lbValue, node *Ast) {
	map_type = base_type(map_type)
	gb_assert_handler("Assertion Failure", "map_type.Kind == Type_Map", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
	var kp lbValue
	hash := lb_gen_map_key_hash(p, map_ptr, map_key, &kp)
	v := lb_emit_conv(p, map_value, map_type.Map.Value)
	value_ptr := lb_address_from_load_or_generate_local(p, v)
	if buildContext.DynamicMapCalls {
		args := make([]lbValue, 6)
		args[0] = lb_emit_conv(p, map_ptr, t_raw_map_ptr)
		args[1] = lb_gen_map_info_ptr(p.Module, map_type)
		args[2] = hash
		args[3] = lb_emit_conv(p, kp, t_rawptr)
		args[4] = lb_emit_conv(p, value_ptr, t_rawptr)
		args[5] = lb_emit_source_code_location_as_global_from_node(p, node)
		lb_emit_runtime_call(p, "__dynamic_map_set", args)
	} else {
		map_set_proc := lb_map_set_proc_for_type(p.Module, map_type)
		args := make([]lbValue, 5)
		args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
		args[1] = hash
		args[2] = lb_emit_conv(p, kp, t_rawptr)
		args[3] = lb_emit_conv(p, value_ptr, t_rawptr)
		args[4] = lb_emit_source_code_location_as_global_from_node(p, node)
		lb_emit_call(p, map_set_proc, args)
	}
}

func lb_dynamic_map_reserve(p *lbProcedure, map_ptr lbValue, capacity isize, pos TokenPos) lbValue {
	proc_name := ""
	if p.Entity != nil {
		proc_name = p.Entity.Token.String
	}
	args := make([]lbValue, 4)
	args[0] = lb_emit_conv(p, map_ptr, t_rawptr)
	args[1] = lb_gen_map_info_ptr(p.Module, type_deref(map_ptr.Type))
	args[2] = lb_const_int(p.Module, t_uint, u64(capacity))
	args[3] = lb_emit_source_code_location_as_global(p, proc_name, pos)
	return lb_emit_runtime_call(p, "__dynamic_map_reserve", args)
}

func lb_verify_function(m *lbModule, p *lbProcedure, dump_ll ...bool) {
	if buildContext.InternalIgnoreLLVMVerification {
		return
	}
	dump := len(dump_ll) > 0 && dump_ll[0]
	if !m.DebugBuilder && LLVMVerifyFunction(p.Value, LLVMReturnStatusAction) != 0 {
		var llvm_error *i8
		gb_printf_err("LLVM CODE GEN FAILED FOR PROCEDURE: %s\n", goStr(p.Name))
		LLVMDumpValue(p.Value)
		gb_printf_err("\n")
		if dump {
			gb_printf_err("\n\n\n")
			filepath_ll := lb_filepath_ll_for_module(m)
			if LLVMPrintModuleToFile(m.Mod, filepath_ll, &llvm_error) != 0 {
				gb_printf_err("LLVM Error: %s\n", llvm_error)
			}
		}
		LLVMVerifyFunction(p.Value, LLVMPrintMessageAction)
		exit_with_errors()
	}
}

func lb_llvm_module_verification_worker_proc(data unsafe.Pointer) isize {
	if buildContext.InternalIgnoreLLVMVerification {
		return 0
	}
	var llvm_error *i8
	defer LLVMDisposeMessage(llvm_error)
	m := (*lbModule)(data)
	if LLVMVerifyModule(m.Mod, LLVMReturnStatusAction, &llvm_error) != 0 {
		gb_printf_err("LLVM Error in module %s:\n%s\n", m.ModuleName, llvm_error)
		if buildContext.KeepTempFiles {
			debugf("[Section] %s\n", "LLVM Print Module to File")
			if buildContext.ShowMoreTimings {
				timings_start_section(&global_timings, "LLVM Print Module to File")
			}
			filepath_ll := lb_filepath_ll_for_module(m)
			if LLVMPrintModuleToFile(m.Mod, filepath_ll, &llvm_error) != 0 {
				gb_printf_err("LLVM Error: %s\n", llvm_error)
				exit_with_errors()
				return 0
			}
		}
		exit_with_errors()
		return 1
	}
	return 0
}

func lb_init_global_var(m *lbModule, p *lbProcedure, e *Entity, init_expr *Ast, v *lbGlobalVariable) bool {
	if init_expr != nil {
		init := lb_build_expr(p, init_expr)
		if init.Value == 0 {
			global_type := llvm_addr_type(p.Module, v.Var)
			if is_type_untyped_nil(init.Type) {
				LLVMSetInitializer(v.Var.Value, LLVMConstNull(global_type))
				v.IsInitialized = true
				if e.Variable.IsRodata {
					LLVMSetGlobalConstant(v.Var.Value, true)
				}
				return true
			}
			gb_assert_handler("Panic", 0, "llvm_backend_map_verify_global_init_startup_proc_types.go", 0, "Invalid init value, got %s", expr_to_string(init_expr))
		}
		if is_type_any(e.Type) {
			v.Init = init
		} else if lb_is_const_or_global(init) {
			if !v.IsInitialized {
				if is_type_proc(init.Type) {
					init.Value = LLVMConstPointerCast(init.Value, lb_type(p.Module, init.Type))
				}
				LLVMSetInitializer(v.Var.Value, init.Value)
				v.IsInitialized = true
				if e.Variable.IsRodata {
					LLVMSetGlobalConstant(v.Var.Value, true)
				}
				return true
			}
		} else {
			v.Init = init
		}
	}
	if v.Init.Value != 0 {
		gb_assert_handler("Assertion Failure", "!v.IsInitialized", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		t := type_deref(v.Var.Type)
		if is_type_any(t) {
			var_type := default_type(v.Init.Type)
			var_name := gb_string_make(permanent_allocator(), "__$global_any::")
			e_str := string_canonical_entity_name(temporary_allocator(), e)
			var_name = gb_string_append_length(var_name, e_str.Text, e_str.Len)
			g := lb_add_global_generated_with_name(m, var_type, lbValue{}, goStr(var_name))
			lb_addr_store(p, g, v.Init)
			gp := lb_addr_get_ptr(p, g)
			data := lb_emit_struct_ep(p, v.Var, 0)
			ti := lb_emit_struct_ep(p, v.Var, 1)
			lb_emit_store(p, data, lb_emit_conv(p, gp, t_rawptr))
			lb_emit_store(p, ti, lb_typeid(p.Module, var_type))
		} else {
			sz := type_size_of(e.Type)
			if sz >= 4*1024 {
				warning(init_expr, "[Possible Code Generation Issue] Non-constant initialization is large (%d bytes), and might cause problems with LLVM", sz)
			}
			vt := llvm_addr_type(p.Module, v.Var)
			src0 := lb_emit_conv(p, v.Init, t)
			src := OdinLLVMBuildTransmute(p, src0.Value, vt)
			dst := v.Var.Value
			LLVMBuildStore(p.Builder, src, dst)
		}
		v.IsInitialized = true
	}
	return false
}

func lb_create_startup_runtime_generate_body(m *lbModule, p *lbProcedure) {
	lb_begin_procedure_body(p)
	defer lb_end_procedure_body(p)
	lb_setup_type_info_data(m)
	if p.ObjCNames != nil {
		LLVMBuildCall2(p.Builder, lb_type_internal_for_procedures_raw(m, p.ObjCNames.Type), p.ObjCNames.Value, nil, 0, "")
	}
	dummy_type := alloc_type_proc(nil, nil, 0, nil, 0, false, ProcCC_Odin)
	raw_dummy_type := lb_type_internal_for_procedures_raw(m, dummy_type)
	for _, v := range p.GlobalVariables {
		if v.IsInitialized {
			continue
		}
		entity_module := m
		e := v.Decl.Entity
		gb_assert_handler("Assertion Failure", "e.Kind == Entity_Variable", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		e.CodeGenModule = entity_module
		init_expr := v.Decl.InitExpr
		if init_expr == nil && v.Init.Value == 0 {
			continue
		}
		if false && type_size_of(e.Type) > 8 {
			ename := lb_get_entity_name(m, e)
			name := "__$startup$" + ename
			dummy := lb_create_dummy_procedure(m, name, dummy_type)
			dummy.IsStartup = true
			LLVMSetVisibility(dummy.Value, LLVMHiddenVisibility)
			if buildContext.UseSeparateModules {
				LLVMSetLinkage(p.Value, LLVMWeakAnyLinkage)
			} else {
				LLVMSetLinkage(p.Value, LLVMInternalLinkage)
			}
			lb_begin_procedure_body(dummy)
			lb_init_global_var(m, dummy, e, init_expr, v)
			lb_end_procedure_body(dummy)
			context_ptr := lb_find_or_generate_context_ptr(p).Addr.Value
			cast_ctx := LLVMBuildBitCast(p.Builder, context_ptr, LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0), "")
			ctx_arr := []LLVMValueRef{cast_ctx}
			LLVMBuildCall2(p.Builder, raw_dummy_type, dummy.Value, &ctx_arr[0], 1, "")
		} else {
			lb_init_global_var(m, p, e, init_expr, v)
		}
	}
	info := m.Gen.Info
	for _, e := range info.InitProcedures {
		value := lb_find_procedure_value_from_entity(m, e)
		lb_emit_call(p, value, nil, ProcInlining_none, ProcTailing_none)
	}
}

func lb_create_startup_runtime(main_module *lbModule, objc_names *lbProcedure, global_variables *Array[*lbGlobalVariable]) *lbProcedure {
	proc_type := alloc_type_proc(nil, nil, 0, nil, 0, false, ProcCC_Odin)
	p := lb_create_dummy_procedure(main_module, "__$startup_runtime", proc_type)
	p.IsStartup = true
	lb_add_attribute_to_proc(p, "optnone")
	lb_add_attribute_to_proc(p, "noinline")
	LLVMSetVisibility(p.Value, LLVMHiddenVisibility)
	if buildContext.UseSeparateModules {
		LLVMSetLinkage(p.Value, LLVMWeakAnyLinkage)
	} else {
		LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	}
	p.GlobalVariables = global_variables
	p.ObjCNames = objc_names
	lb_create_startup_runtime_generate_body(main_module, p)
	return p
}

func lb_create_cleanup_runtime(main_module *lbModule) *lbProcedure {
	proc_type := alloc_type_proc(nil, nil, 0, nil, 0, false, ProcCC_Odin)
	p := lb_create_dummy_procedure(main_module, "__$cleanup_runtime", proc_type)
	p.IsStartup = true
	lb_add_attribute_to_proc(p, "optnone")
	lb_add_attribute_to_proc(p, "noinline")
	LLVMSetVisibility(p.Value, LLVMHiddenVisibility)
	if buildContext.UseSeparateModules {
		LLVMSetLinkage(p.Value, LLVMWeakAnyLinkage)
	} else {
		LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	}
	lb_begin_procedure_body(p)
	defer lb_end_procedure_body(p)
	info := main_module.Gen.Info
	for _, e := range info.FiniProcedures {
		value := lb_find_procedure_value_from_entity(main_module, e)
		lb_emit_call(p, value, nil, ProcInlining_none, ProcTailing_none)
	}
	lb_verify_function(main_module, p)
	return p
}

func lb_generate_procedures_and_types_per_module(data unsafe.Pointer) isize {
	m := (*lbModule)(data)
	for _, e := range m.GlobalTypesToCreate {
		lb_get_entity_name(m, e)
		lb_type(m, e.Type)
	}
	for _, e := range m.GlobalProceduresToCreate {
		lb_get_entity_name(m, e)
		m.ProceduresToGenerate.Enqueue(lb_create_procedure(m, e))
	}
	return 0
}

func llvm_global_entity_cmp(a, b unsafe.Pointer) int {
	x := *(*unsafe.Pointer)(a)
	y := *(*unsafe.Pointer)(b)
	if x == y {
		return 0
	}
	xe := (*Entity)(x)
	ye := (*Entity)(y)
	if xe.Kind != ye.Kind {
		return int(xe.Kind) - int(ye.Kind)
	}
	cmp := token_pos_cmp(xe.Token.Pos, ye.Token.Pos)
	return cmp
}

func lb_create_global_procedures_and_types(gen *lbGenerator, info *CheckerInfo, do_threading bool) {
	for _, e := range info.Entities {
		_ = e.Token.String
		scope := e.Scope
		if scope.Flags&ScopeFlag_File == 0 {
			continue
		}
		package_scope := scope.Parent
		gb_assert_handler("Assertion Failure", "package_scope.Flags&ScopeFlag_Pkg != 0", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		switch e.Kind {
		case Entity_Variable:
			continue
		case Entity_ProcGroup:
			continue
		case Entity_TypeName, Entity_Procedure:
		case Entity_Constant:
			if buildContext.ODIN_DEBUG {
				lb_add_debug_info_for_global_constant_from_entity(gen, e)
			}
		}
		polymorphic_struct := false
		if e.Type != nil && e.Kind == Entity_TypeName {
			bt := base_type(e.Type)
			if bt.Kind == Type_Struct {
				polymorphic_struct = is_type_polymorphic(bt)
			}
		}
		if !polymorphic_struct && e.MinDepCount.Load() == 0 {
			continue
		}
		m := &gen.DefaultModule
		if buildContext.UseSeparateModules {
			m = lb_module_of_entity(gen, e)
		}
		gb_assert_handler("Assertion Failure", "m != nil", "llvm_backend_map_verify_global_init_startup_proc_types.go", 0)
		if e.Kind == Entity_Procedure {
			if e.Procedure.IsForeign && e.Procedure.IsObjCImplOrImport {
				continue
			}
			m.GlobalProceduresToCreate.Add(e)
		} else if e.Kind == Entity_TypeName {
			m.GlobalTypesToCreate.Add(e)
		}
	}
	for _, m := range gen.Modules {
		m.GlobalTypesToCreate.Sort(llvm_global_entity_cmp)
		m.GlobalProceduresToCreate.Sort(llvm_global_entity_cmp)
	}
	if do_threading {
		for _, m := range gen.Modules {
			thread_pool_add_task(lb_generate_procedures_and_types_per_module, unsafe.Pointer(m))
		}
	} else {
		for _, m := range gen.Modules {
			lb_generate_procedures_and_types_per_module(unsafe.Pointer(m))
		}
	}
	thread_pool_wait()
}
