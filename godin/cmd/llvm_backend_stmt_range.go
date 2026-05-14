package cmd

import "unicode/utf8"

func is_power_of_two(x uint64) bool {
	return x != 0 && (x&(x-1)) == 0
}

func floor_log2(x uint64) uint64 {
	r := uint64(0)
	for x > 1 {
		x >>= 1
		r++
	}
	return r
}

func bit_set_to_int(t *Type) *Type {
	return t.BitSet.Underlying
}

type rangeIndexedResult struct {
	Val  lbValue
	Idx  lbValue
	Loop *lbBlock
	Done *lbBlock
}

type rangeMapResult struct {
	Val  lbValue
	Key  lbValue
	Loop *lbBlock
	Done *lbBlock
}

type rangeStringResult struct {
	Val  lbValue
	Idx  lbValue
	Loop *lbBlock
	Done *lbBlock
}

type rangeEnumResult struct {
	Val  lbValue
	Idx  lbValue
	Loop *lbBlock
	Done *lbBlock
}

func lb_build_range_indexed(p *lbProcedure, expr lbValue, val_type *Type, count_ptr lbValue, is_reverse bool, unroll_count ...int64) rangeIndexedResult {
	m := p.Module
	expr_type := base_type(type_deref(expr.Type))
	var count lbValue
	switch expr_type.Kind {
	case Type_Array:
		count = lb_const_int(m, t_int, uint64(expr_type.Array.Count))
	}
	var val lbValue
	var idx lbValue
	loop := lb_create_block(p, "for.index.loop")
	body := lb_create_block(p, "for.index.body")
	done := lb_create_block(p, "for.index.done")
	index := lb_add_local_generated(p, t_int, false)
	if !is_reverse {
		lb_addr_store(p, index, lb_const_int(m, t_int, ^uint64(0)))
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		incr := lb_emit_arith(p, Token_Add, lb_addr_load(p, index), lb_const_int(m, t_int, 1), t_int)
		lb_addr_store(p, index, incr)
		if count.Value == 0 {
			if count_ptr.Value == 0 {
				gb_assert_handler("Assertion Failure", "count_ptr.value != nullptr", "llvm_backend_stmt_range.go", 0, 0)
			}
			count = lb_emit_load(p, count_ptr)
		}
		cond := lb_emit_comp(p, Token_Lt, incr, count)
		lb_emit_if(p, cond, body, done)
	} else {
		if count.Value == 0 {
			if count_ptr.Value == 0 {
				gb_assert_handler("Assertion Failure", "count_ptr.value != nullptr", "llvm_backend_stmt_range.go", 0, 0)
			}
			count = lb_emit_load(p, count_ptr)
		}
		count = lb_emit_conv(p, count, t_int)
		lb_addr_store(p, index, count)
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		incr := lb_emit_arith(p, Token_Sub, lb_addr_load(p, index), lb_const_int(m, t_int, 1), t_int)
		lb_addr_store(p, index, incr)
		anti_cond := lb_emit_comp(p, Token_Lt, incr, lb_const_int(m, t_int, 0))
		lb_emit_if(p, anti_cond, done, body)
	}
	lb_start_block(p, body)
	idx = lb_addr_load(p, index)
	switch expr_type.Kind {
	case Type_Array:
		if val_type != nil {
			val = lb_emit_load(p, lb_emit_array_ep(p, expr, idx))
		}
	case Type_EnumeratedArray:
		if val_type != nil {
			val = lb_emit_load(p, lb_emit_array_ep(p, expr, idx))
			index_type := expr_type.EnumeratedArray.Index
			if compare_exact_values(Token_NotEq, *expr_type.EnumeratedArray.MinValue, exact_value_u64(0)) {
				idx = lb_emit_arith(p, Token_Add, idx, lb_const_value(m, index_type, *expr_type.EnumeratedArray.MinValue), index_type)
			}
		}
	case Type_FixedCapacityDynamicArray:
		if val_type != nil {
			data := lb_emit_struct_ep(p, expr, 0)
			val = lb_emit_load(p, lb_emit_array_ep(p, data, idx))
		}
	case Type_Slice:
		if val_type != nil {
			elem := lb_slice_elem(p, expr)
			val = lb_emit_load(p, lb_emit_ptr_offset(p, elem, idx))
		}
	case Type_DynamicArray:
		if val_type != nil {
			elem := lb_emit_struct_ep(p, expr, 0)
			elem = lb_emit_load(p, elem)
			val = lb_emit_load(p, lb_emit_ptr_offset(p, elem, idx))
		}
	case Type_Struct:
		if !is_type_soa_struct(expr_type) {
			gb_assert_handler("Assertion Failure", "is_type_soa_struct(expr_type)", "llvm_backend_stmt_range.go", 0, 0)
		}
	default:
		gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Cannot do range_indexed")
	}
	return rangeIndexedResult{Val: val, Idx: idx, Loop: loop, Done: done}
}

func lb_map_cell_index_static(p *lbProcedure, typ *Type, cells_ptr lbValue, index lbValue) lbValue {
	var size, length int64
	elem_sz := type_size_of(typ)
	map_cell_size_and_len(typ, &size, &length)
	index = lb_emit_conv(p, index, t_uintptr)
	if size == length*elem_sz {
		elems_ptr := lb_emit_conv(p, cells_ptr, alloc_type_pointer(typ))
		return lb_emit_ptr_offset(p, elems_ptr, index)
	}
	var cell_index lbValue
	var data_index lbValue
	size_const := lb_const_int(p.Module, t_uintptr, uint64(size))
	len_const := lb_const_int(p.Module, t_uintptr, uint64(length))
	if is_power_of_two(uint64(length)) {
		log2_len := floor_log2(uint64(length))
		if log2_len == 0 {
			cell_index = index
		} else {
			cell_index = lb_emit_arith(p, Token_Shr, index, lb_const_int(p.Module, t_uintptr, log2_len), t_uintptr)
		}
		data_index = lb_emit_arith(p, Token_And, index, lb_const_int(p.Module, t_uintptr, uint64(length-1)), t_uintptr)
	} else {
		cell_index = lb_emit_arith(p, Token_Quo, index, len_const, t_uintptr)
		data_index = lb_emit_arith(p, Token_Mod, index, len_const, t_uintptr)
	}
	elems_ptr := lb_emit_conv(p, cells_ptr, t_uintptr)
	cell_offset := lb_emit_arith(p, Token_Mul, size_const, cell_index, t_uintptr)
	elems_ptr = lb_emit_arith(p, Token_Add, elems_ptr, cell_offset, t_uintptr)
	elems_ptr = lb_emit_conv(p, elems_ptr, alloc_type_pointer(typ))
	return lb_emit_ptr_offset(p, elems_ptr, data_index)
}

func lb_map_hash_is_valid(p *lbProcedure, hash lbValue) lbValue {
	top_bit_index := uint64(type_size_of(t_uintptr)*8 - 1)
	shift_amount := lb_const_int(p.Module, t_uintptr, top_bit_index)
	zero := lb_const_int(p.Module, t_uintptr, 0)
	not_empty := lb_emit_comp(p, Token_NotEq, hash, zero)
	not_deleted := lb_emit_arith(p, Token_Shr, hash, shift_amount, t_uintptr)
	not_deleted = lb_emit_comp(p, Token_CmpEq, not_deleted, zero)
	return lb_emit_arith(p, Token_And, not_deleted, not_empty, t_uintptr)
}

func lb_build_range_map(p *lbProcedure, expr lbValue, val_type *Type) rangeMapResult {
	m := p.Module
	typ := base_type(type_deref(expr.Type))
	if typ.Kind != Type_Map {
		gb_assert_handler("Assertion Failure", "type->kind == Type_Map", "llvm_backend_stmt_range.go", 0, 0)
	}
	var idx lbValue
	var loop, done, body, hash_check *lbBlock
	index := lb_add_local_generated(p, t_int, false)
	lb_addr_store(p, index, lb_const_int(m, t_int, ^uint64(0)))
	loop = lb_create_block(p, "for.index.loop")
	lb_emit_jump(p, loop)
	lb_start_block(p, loop)
	incr := lb_emit_arith(p, Token_Add, lb_addr_load(p, index), lb_const_int(m, t_int, 1), t_int)
	lb_addr_store(p, index, incr)
	hash_check = lb_create_block(p, "for.index.hash_check")
	body = lb_create_block(p, "for.index.body")
	done = lb_create_block(p, "for.index.done")
	map_value := lb_emit_load(p, expr)
	capacity := lb_map_cap(p, map_value)
	cond := lb_emit_comp(p, Token_Lt, incr, capacity)
	lb_emit_if(p, cond, hash_check, done)
	lb_start_block(p, hash_check)
	idx = lb_addr_load(p, index)
	ks := lb_map_data_uintptr(p, map_value)
	vs := lb_emit_conv(p, lb_map_cell_index_static(p, typ.Map.Key, ks, capacity), alloc_type_pointer(typ.Map.Value))
	hs := lb_emit_conv(p, lb_map_cell_index_static(p, typ.Map.Value, vs, capacity), alloc_type_pointer(t_uintptr))
	hash := lb_emit_load(p, lb_emit_ptr_offset(p, hs, idx))
	hash_cond := lb_map_hash_is_valid(p, hash)
	lb_emit_if(p, hash_cond, body, loop)
	lb_start_block(p, body)
	key_ptr := lb_map_cell_index_static(p, typ.Map.Key, ks, idx)
	val_ptr := lb_map_cell_index_static(p, typ.Map.Value, vs, idx)
	key := lb_emit_load(p, key_ptr)
	val := lb_emit_load(p, val_ptr)
	return rangeMapResult{Val: val, Key: key, Loop: loop, Done: done}
}

func lb_build_range_string(p *lbProcedure, expr lbValue, val_type *Type, is_reverse bool) rangeStringResult {
	m := p.Module
	count := lb_const_int(m, t_int, 0)
	expr_type := base_type(expr.Type)
	switch expr_type.Kind {
	case Type_Basic:
		count = lb_string_len(p, expr)
	default:
		gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Cannot do range_string")
	}
	loop := lb_create_block(p, "for.string.loop")
	body := lb_create_block(p, "for.string.body")
	done := lb_create_block(p, "for.string.done")
	offset_ := lb_add_local_generated(p, t_int, false)
	var offset lbValue
	var cond lbValue
	if !is_reverse {
		lb_addr_store(p, offset_, lb_const_int(m, t_int, 0))
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		offset = lb_addr_load(p, offset_)
		cond = lb_emit_comp(p, Token_Lt, offset, count)
	} else {
		lb_addr_store(p, offset_, count)
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		offset = lb_addr_load(p, offset_)
		cond = lb_emit_comp(p, Token_Gt, offset, lb_const_int(m, t_int, 0))
	}
	lb_emit_if(p, cond, body, done)
	lb_start_block(p, body)
	var rune_and_len lbValue
	var idx lbValue
	var val lbValue
	if !is_reverse {
		str_elem := lb_emit_ptr_offset(p, lb_string_elem(p, expr), offset)
		str_len := lb_emit_arith(p, Token_Sub, count, offset, t_int)
		args := []lbValue{lb_emit_string(p, str_elem, str_len)}
		rune_and_len = lb_emit_runtime_call(p, "string_decode_rune", args)
		len_ := lb_emit_struct_ev(p, rune_and_len, 1)
		lb_addr_store(p, offset_, lb_emit_arith(p, Token_Add, offset, len_, t_int))
		idx = offset
	} else {
		str_elem := lb_string_elem(p, expr)
		str_len := offset
		args := []lbValue{lb_emit_string(p, str_elem, str_len)}
		rune_and_len = lb_emit_runtime_call(p, "string_decode_last_rune", args)
		len_ := lb_emit_struct_ev(p, rune_and_len, 1)
		lb_addr_store(p, offset_, lb_emit_arith(p, Token_Sub, offset, len_, t_int))
		idx = lb_addr_load(p, offset_)
	}
	if val_type != nil {
		val = lb_emit_struct_ev(p, rune_and_len, 0)
	}
	return rangeStringResult{Val: val, Idx: idx, Loop: loop, Done: done}
}

func lb_build_range_string16(p *lbProcedure, expr lbValue, val_type *Type, is_reverse bool) rangeStringResult {
	m := p.Module
	count := lb_const_int(m, t_int, 0)
	expr_type := base_type(expr.Type)
	switch expr_type.Kind {
	case Type_Basic:
		count = lb_string_len(p, expr)
	default:
		gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Cannot do range_string16")
	}
	loop := lb_create_block(p, "for.string16.loop")
	body := lb_create_block(p, "for.string16.body")
	done := lb_create_block(p, "for.string16.done")
	offset_ := lb_add_local_generated(p, t_int, false)
	var offset lbValue
	var cond lbValue
	if !is_reverse {
		lb_addr_store(p, offset_, lb_const_int(m, t_int, 0))
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		offset = lb_addr_load(p, offset_)
		cond = lb_emit_comp(p, Token_Lt, offset, count)
	} else {
		lb_addr_store(p, offset_, count)
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		offset = lb_addr_load(p, offset_)
		cond = lb_emit_comp(p, Token_Gt, offset, lb_const_int(m, t_int, 0))
	}
	lb_emit_if(p, cond, body, done)
	lb_start_block(p, body)
	var rune_and_len lbValue
	var idx lbValue
	var val lbValue
	if !is_reverse {
		str_elem := lb_emit_ptr_offset(p, lb_string_elem(p, expr), offset)
		str_len := lb_emit_arith(p, Token_Sub, count, offset, t_int)
		args := []lbValue{lb_emit_string16(p, str_elem, str_len)}
		rune_and_len = lb_emit_runtime_call(p, "string16_decode_rune", args)
		len_ := lb_emit_struct_ev(p, rune_and_len, 1)
		lb_addr_store(p, offset_, lb_emit_arith(p, Token_Add, offset, len_, t_int))
		idx = offset
	} else {
		str_elem := lb_string_elem(p, expr)
		str_len := offset
		args := []lbValue{lb_emit_string16(p, str_elem, str_len)}
		rune_and_len = lb_emit_runtime_call(p, "string16_decode_last_rune", args)
		len_ := lb_emit_struct_ev(p, rune_and_len, 1)
		lb_addr_store(p, offset_, lb_emit_arith(p, Token_Sub, offset, len_, t_int))
		idx = lb_addr_load(p, offset_)
	}
	if val_type != nil {
		val = lb_emit_struct_ev(p, rune_and_len, 0)
	}
	return rangeStringResult{Val: val, Idx: idx, Loop: loop, Done: done}
}

func lb_strip_and_prefix(ident *Ast) *Ast {
	if ident != nil {
		if ident.Kind == Ast_UnaryExpr && ident.UnaryExpr.Op.Kind == Token_And {
			ident = ident.UnaryExpr.Expr
		}
		if ident.Kind != Ast_Ident {
			gb_assert_handler("Assertion Failure", "ident->kind == Ast_Ident", "llvm_backend_stmt_range.go", 0, 0)
		}
	}
	return ident
}

func lb_build_range_interval(p *lbProcedure, node *AstBinaryExpr, rs *AstRangeStmt, scope *Scope) {
	ADD_EXTRA_WRAPPING_CHECK := true
	m := p.Module
	lb_open_scope(p, scope)
	if rs.Init != nil {
		lb_build_stmt(p, rs.Init)
	}
	var val0 *Ast
	if len(rs.Vals) > 0 {
		val0 = lb_strip_and_prefix(rs.Vals[0])
	}
	var val1 *Ast
	if len(rs.Vals) > 1 {
		val1 = lb_strip_and_prefix(rs.Vals[1])
	}
	var val0_type *Type
	var val1_type *Type
	if val0 != nil && !is_blank_ident(val0) {
		val0_type = type_of_expr(val0)
	}
	if val1 != nil && !is_blank_ident(val1) {
		val1_type = type_of_expr(val1)
	}
	op := Token_Lt
	switch node.Op.Kind {
	case Token_Ellipsis:
		op = Token_LtEq
	case Token_RangeFull:
		op = Token_LtEq
	case Token_RangeHalf:
		op = Token_Lt
	default:
		gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Invalid interval operator")
	}
	lower := lb_build_expr(p, node.Left)
	var upper lbValue
	var value lbAddr
	if val0_type != nil {
		e := entity_of_node(val0)
		value = lb_add_local(p, val0_type, e)
	} else {
		value = lb_add_local_generated(p, lower.Type, false)
	}
	lb_addr_store(p, value, lower)
	var index lbAddr
	if val1_type != nil {
		e := entity_of_node(val1)
		index = lb_add_local(p, val1_type, e)
	} else {
		index = lb_add_local_generated(p, t_int, false)
	}
	lb_addr_store(p, index, lb_const_int(m, t_int, 0))
	loop := lb_create_block(p, "for.interval.loop")
	body := lb_create_block(p, "for.interval.body")
	done := lb_create_block(p, "for.interval.done")
	if rs.Label != nil && p.DebugInfo != 0 {
		label := lb_create_block(p, "for.interval.label")
		lb_emit_jump(p, label)
		lb_start_block(p, label)
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, rs.Label))
		lb_add_debug_label(p, rs.Label, label)
	}
	lb_emit_jump(p, loop)
	lb_start_block(p, loop)
	upper = lb_build_expr(p, node.Right)
	curr_value := lb_addr_load(p, value)
	cond := lb_emit_comp(p, op, curr_value, upper)
	lb_emit_if(p, cond, body, done)
	lb_start_block(p, body)
	val := lb_addr_load(p, value)
	idx := lb_addr_load(p, index)
	if val0_type != nil {
		lb_store_range_stmt_val(p, val0, val)
	}
	if val1_type != nil {
		lb_store_range_stmt_val(p, val1, idx)
	}
	{
		var check *lbBlock
		post := lb_create_block(p, "for.interval.post")
		continue_block := post
		if ADD_EXTRA_WRAPPING_CHECK && op == Token_LtEq {
			check = lb_create_block(p, "for.interval.check")
			continue_block = check
		}
		lb_push_target_list(p, rs.Label, done, continue_block, nil)
		lb_build_stmt(p, rs.Body)
		lb_close_scope(p, lbDeferExit_Default, nil, node.Left)
		lb_pop_target_list(p)
		if p.DebugInfo != 0 {
			LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, rs.Body))
		}
		if check != nil {
			lb_emit_jump(p, check)
			lb_start_block(p, check)
			check_cond := lb_emit_comp(p, Token_NotEq, curr_value, upper)
			lb_emit_if(p, check_cond, post, done)
		} else {
			lb_emit_jump(p, post)
		}
		lb_start_block(p, post)
		lb_emit_increment(p, value.Addr)
		lb_emit_increment(p, index.Addr)
		lb_emit_jump(p, loop)
	}
	lb_start_block(p, done)
}

func lb_enum_values_slice(p *lbProcedure, enum_type *Type, enum_count *int64) lbValue {
	t := enum_type
	if !is_type_enum(t) {
		gb_assert_handler("Assertion Failure", "is_type_enum(t)", "llvm_backend_stmt_range.go", 0, 0)
	}
	t = base_type(t)
	if t.Kind != Type_Enum {
		gb_assert_handler("Assertion Failure", "t->kind == Type_Enum", "llvm_backend_stmt_range.go", 0, 0)
	}
	enum_count_val := int64(len(t.Enum.Fields))
	if enum_count != nil {
		*enum_count = enum_count_val
	}
	ti := lb_type_info(p, t)
	variant := lb_emit_struct_ep(p, ti, 4)
	eti_ptr := lb_emit_conv(p, variant, t_type_info_enum_ptr)
	values := lb_emit_load(p, lb_emit_struct_ep(p, eti_ptr, 2))
	return values
}

func lb_build_range_enum(p *lbProcedure, enum_type *Type, val_type *Type) rangeEnumResult {
	m := p.Module
	t := enum_type
	if !is_type_enum(t) {
		gb_assert_handler("Assertion Failure", "is_type_enum(t)", "llvm_backend_stmt_range.go", 0, 0)
	}
	t = base_type(t)
	core_elem := core_type(t)
	var enum_count int64
	values := lb_enum_values_slice(p, enum_type, &enum_count)
	values_data := lb_slice_elem(p, values)
	max_count := lb_const_int(m, t_int, uint64(enum_count))
	offset_ := lb_add_local_generated(p, t_int, false)
	lb_addr_store(p, offset_, lb_const_int(m, t_int, 0))
	loop := lb_create_block(p, "for.enum.loop")
	lb_emit_jump(p, loop)
	lb_start_block(p, loop)
	body := lb_create_block(p, "for.enum.body")
	done := lb_create_block(p, "for.enum.done")
	offset := lb_addr_load(p, offset_)
	cond := lb_emit_comp(p, Token_Lt, offset, max_count)
	lb_emit_if(p, cond, body, done)
	lb_start_block(p, body)
	val_ptr := lb_emit_ptr_offset(p, values_data, offset)
	lb_emit_increment(p, offset_.Addr)
	var val lbValue
	if val_type != nil {
		if !are_types_identical(enum_type, val_type) {
			gb_assert_handler("Assertion Failure", "are_types_identical(enum_type, val_type)", "llvm_backend_stmt_range.go", 0, 0)
		}
		if is_type_integer(core_elem) {
			i_ := lb_emit_load(p, lb_emit_conv(p, val_ptr, t_i64_ptr))
			val = lb_emit_conv(p, i_, enum_type)
		} else {
			gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "TODO enum core type")
		}
	}
	return rangeEnumResult{Val: val, Idx: offset, Loop: loop, Done: done}
}

func lb_build_range_tuple(p *lbProcedure, rs *AstRangeStmt, scope *Scope) {
	expr := unparen_expr(rs.Expr)
	et := base_type(type_deref(type_of_expr(expr)))
	if et.Kind != Type_Tuple {
		gb_assert_handler("Assertion Failure", "et->kind == Type_Tuple", "llvm_backend_stmt_range.go", 0, 0)
	}
	value_count := len(et.Tuple.Variables)
	values := make([]lbValue, value_count)
	lb_open_scope(p, scope)
	if rs.Init != nil {
		lb_build_stmt(p, rs.Init)
	}
	loop := lb_create_block(p, "for.tuple.loop")
	lb_emit_jump(p, loop)
	lb_start_block(p, loop)
	body := lb_create_block(p, "for.tuple.body")
	done := lb_create_block(p, "for.tuple.done")
	tuple_value := lb_build_expr(p, expr)
	tuple := tuple_value.Type
	if tuple.Kind != Type_Tuple {
		gb_assert_handler("Assertion Failure", "tuple->kind == Type_Tuple", "llvm_backend_stmt_range.go", 0, 0)
	}
	tuple_count := len(tuple.Tuple.Variables)
	cond_index := tuple_count - 1
	cond := lb_emit_tuple_ev(p, tuple_value, int32(cond_index))
	lb_emit_if(p, cond, body, done)
	lb_start_block(p, body)
	for i := 0; i < value_count; i++ {
		values[i] = lb_emit_tuple_ev(p, tuple_value, int32(i))
	}
	if len(rs.Vals) > value_count {
		gb_assert_handler("Assertion Failure", "rs->vals.count <= value_count", "llvm_backend_stmt_range.go", 0, 0)
	}
	for i := 0; i < len(rs.Vals); i++ {
		val := rs.Vals[i]
		if val != nil {
			lb_store_range_stmt_val(p, val, values[i])
		}
	}
	lb_push_target_list(p, rs.Label, done, loop, nil)
	lb_build_stmt(p, rs.Body)
	lb_close_scope(p, lbDeferExit_Default, nil, rs.Body)
	lb_pop_target_list(p)
	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, rs.Body))
	}
	lb_emit_jump(p, loop)
	lb_start_block(p, done)
}

func lb_build_range_stmt_struct_soa(p *lbProcedure, rs *AstRangeStmt, scope *Scope) {
	expr := unparen_expr(rs.Expr)
	is_reverse := rs.Reverse
	lb_open_scope(p, scope)
	if rs.Init != nil {
		lb_build_stmt(p, rs.Init)
	}
	var val0 *Ast
	if len(rs.Vals) > 0 {
		val0 = lb_strip_and_prefix(rs.Vals[0])
	}
	var val1 *Ast
	if len(rs.Vals) > 1 {
		val1 = lb_strip_and_prefix(rs.Vals[1])
	}
	var val_types [2]*Type
	if val0 != nil && !is_blank_ident(val0) {
		val_types[0] = type_of_expr(val0)
	}
	if val1 != nil && !is_blank_ident(val1) {
		val_types[1] = type_of_expr(val1)
	}
	array := lb_build_addr(p, expr)
	if is_type_pointer(lb_addr_type(array)) {
		array = lb_addr(lb_addr_load(p, array))
	}
	count := lb_soa_struct_len(p, lb_addr_load(p, array))
	index := lb_add_local_generated(p, t_int, false)
	if rs.Label != nil && p.DebugInfo != 0 {
		label := lb_create_block(p, "for.soa.label")
		lb_emit_jump(p, label)
		lb_start_block(p, label)
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, rs.Label))
		lb_add_debug_label(p, rs.Label, label)
	}
	var loop, body, done *lbBlock
	if !is_reverse {
		lb_addr_store(p, index, lb_const_int(p.Module, t_int, ^uint64(0)))
		loop = lb_create_block(p, "for.soa.loop")
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		incr := lb_emit_arith(p, Token_Add, lb_addr_load(p, index), lb_const_int(p.Module, t_int, 1), t_int)
		lb_addr_store(p, index, incr)
		body = lb_create_block(p, "for.soa.body")
		done = lb_create_block(p, "for.soa.done")
		cond := lb_emit_comp(p, Token_Lt, incr, count)
		lb_emit_if(p, cond, body, done)
	} else {
		lb_addr_store(p, index, count)
		loop = lb_create_block(p, "for.soa.loop")
		lb_emit_jump(p, loop)
		lb_start_block(p, loop)
		incr := lb_emit_arith(p, Token_Sub, lb_addr_load(p, index), lb_const_int(p.Module, t_int, 1), t_int)
		lb_addr_store(p, index, incr)
		body = lb_create_block(p, "for.soa.body")
		done = lb_create_block(p, "for.soa.done")
		cond := lb_emit_comp(p, Token_Lt, incr, lb_const_int(p.Module, t_int, 0))
		lb_emit_if(p, cond, done, body)
	}
	lb_start_block(p, body)
	if val_types[0] != nil {
		e := entity_of_node(val0)
		if e != nil {
			soa_val := lb_addr_soa_variable(array.Addr, lb_addr_load(p, index), nil)
			p.Module.SoaValues[e] = soa_val
		}
	}
	if val_types[1] != nil {
		lb_store_range_stmt_val(p, val1, lb_addr_load(p, index))
	}
	lb_push_target_list(p, rs.Label, done, loop, nil)
	lb_build_stmt(p, rs.Body)
	lb_close_scope(p, lbDeferExit_Default, nil, rs.Body)
	lb_pop_target_list(p)
	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, rs.Body))
	}
	lb_emit_jump(p, loop)
	lb_start_block(p, done)
}

func lb_build_range_stmt(p *lbProcedure, rs *AstRangeStmt, scope *Scope) {
	expr := unparen_expr(rs.Expr)
	if is_ast_range(expr) {
		lb_build_range_interval(p, &expr.BinaryExpr, rs, scope)
		return
	}
	expr_type := type_of_expr(expr)
	if expr_type != nil {
		et := base_type(type_deref(expr_type))
		if is_type_soa_struct(et) {
			lb_build_range_stmt_struct_soa(p, rs, scope)
			return
		}
	}
	tav := type_and_value_of_expr(expr)
	if tav.Mode != Addressing_Type {
		expr_type = type_of_expr(expr)
		et := base_type(type_deref(expr_type))
		if et.Kind == Type_Tuple {
			lb_build_range_tuple(p, rs, scope)
			return
		}
	}
	lb_open_scope(p, scope)
	if rs.Init != nil {
		lb_build_stmt(p, rs.Init)
	}
	var val0 *Ast
	if len(rs.Vals) > 0 {
		val0 = lb_strip_and_prefix(rs.Vals[0])
	}
	var val1 *Ast
	if len(rs.Vals) > 1 {
		val1 = lb_strip_and_prefix(rs.Vals[1])
	}
	var val0_type *Type
	var val1_type *Type
	if val0 != nil && !is_blank_ident(val0) {
		val0_type = type_of_expr(val0)
	}
	if val1 != nil && !is_blank_ident(val1) {
		val1_type = type_of_expr(val1)
	}
	var val lbValue
	var key lbValue
	var loop, done *lbBlock
	is_map := false
	if rs.Label != nil && p.DebugInfo != 0 {
		label := lb_create_block(p, "for.range.label")
		lb_emit_jump(p, label)
		lb_start_block(p, label)
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_ast(p, rs.Label))
		lb_add_debug_label(p, rs.Label, label)
	}
	if tav.Mode == Addressing_Type {
		r := lb_build_range_enum(p, type_deref(tav.Type), val0_type)
		val = r.Val
		key = r.Idx
		loop = r.Loop
		done = r.Done
	} else {
		expr_type = type_of_expr(expr)
		et := base_type(type_deref(expr_type))
		switch et.Kind {
		case Type_Map:
			is_map = true
			map_val := lb_build_addr_ptr(p, expr)
			if is_type_pointer(type_deref(map_val.Type)) {
				map_val = lb_emit_load(p, map_val)
			}
			r := lb_build_range_map(p, map_val, val1_type)
			val = r.Val
			key = r.Key
			loop = r.Loop
			done = r.Done

		case Type_Array:
			var array lbValue
			addr := lb_build_addr(p, expr)
			switch addr.Kind {
			case lbAddr_Swizzle, lbAddr_SwizzleLarge:
				array = lb_address_from_load(p, lb_addr_load(p, addr))
			default:
				array = lb_addr_get_ptr(p, addr)
				if is_type_pointer(type_deref(array.Type)) {
					array = lb_emit_load(p, array)
				}
			}
			count_ptr := lb_add_local_generated(p, t_int, false)
			lb_addr_store(p, count_ptr, lb_const_int(p.Module, t_int, uint64(et.Array.Count)))
			r := lb_build_range_indexed(p, array, val0_type, count_ptr.Addr, rs.Reverse)
			val = r.Val
			key = r.Idx
			loop = r.Loop
			done = r.Done

		case Type_EnumeratedArray:
			array := lb_build_addr_ptr(p, expr)
			if is_type_pointer(type_deref(array.Type)) {
				array = lb_emit_load(p, array)
			}
			count_ptr := lb_add_local_generated(p, t_int, false)
			lb_addr_store(p, count_ptr, lb_const_int(p.Module, t_int, uint64(et.EnumeratedArray.Count)))
			r := lb_build_range_indexed(p, array, val0_type, count_ptr.Addr, rs.Reverse)
			val = r.Val
			key = r.Idx
			loop = r.Loop
			done = r.Done

		case Type_FixedCapacityDynamicArray:
			array := lb_build_addr_ptr(p, expr)
			if is_type_pointer(type_deref(array.Type)) {
				array = lb_emit_load(p, array)
			}
			count_ptr := lb_emit_struct_ep(p, array, 1)
			r := lb_build_range_indexed(p, array, val0_type, count_ptr, rs.Reverse)
			val = r.Val
			key = r.Idx
			loop = r.Loop
			done = r.Done

		case Type_DynamicArray:
			var count_ptr lbValue
			array := lb_build_addr_ptr(p, expr)
			if is_type_pointer(type_deref(array.Type)) {
				array = lb_emit_load(p, array)
			}
			count_ptr = lb_emit_struct_ep(p, array, 1)
			r := lb_build_range_indexed(p, array, val0_type, count_ptr, rs.Reverse)
			val = r.Val
			key = r.Idx
			loop = r.Loop
			done = r.Done

		case Type_Slice:
			var count_ptr lbValue
			slice := lb_build_expr(p, expr)
			if is_type_pointer(slice.Type) {
				count_ptr = lb_emit_struct_ep(p, slice, 1)
				slice = lb_emit_load(p, slice)
			} else {
				count_ptr = lb_add_local_generated(p, t_int, false).Addr
				lb_emit_store(p, count_ptr, lb_slice_len(p, slice))
			}
			r := lb_build_range_indexed(p, slice, val0_type, count_ptr, rs.Reverse)
			val = r.Val
			key = r.Idx
			loop = r.Loop
			done = r.Done

		case Type_Basic:
			str := lb_build_expr(p, expr)
			if is_type_pointer(str.Type) {
				str = lb_emit_load(p, str)
			}
			if is_type_untyped(expr_type) {
				s := lb_add_local_generated(p, default_type(str.Type), false)
				lb_addr_store(p, s, str)
				str = lb_addr_load(p, s)
			}
			t := base_type(str.Type)
			if is_type_cstring(t) {
				gb_assert_handler("Assertion Failure", "!is_type_cstring(t)", "llvm_backend_stmt_range.go", 0, 0)
			}
			if is_type_string16(t) {
				r := lb_build_range_string16(p, str, val0_type, rs.Reverse)
				val = r.Val
				key = r.Idx
				loop = r.Loop
				done = r.Done
			} else {
				r := lb_build_range_string(p, str, val0_type, rs.Reverse)
				val = r.Val
				key = r.Idx
				loop = r.Loop
				done = r.Done
			}

		case Type_Tuple:
			gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Tuple should be handled already")

		case Type_BitSet:
			m := p.Module
			the_set := lb_build_expr(p, expr)
			if is_type_pointer(type_deref(the_set.Type)) {
				the_set = lb_emit_load(p, the_set)
			}
			elem := et.BitSet.Elem
			mask := bit_set_to_int(et)
			all_mask := lb_const_value(p.Module, mask, exact_bit_set_all_set_mask(et))
			initial_mask := lb_emit_arith(p, Token_And, the_set, all_mask, mask)
			if rs.Reverse {
				initial_mask = lb_emit_reverse_bits(p, initial_mask, mask)
			}
			remaining := lb_add_local_generated(p, mask, false)
			lb_addr_store(p, remaining, initial_mask)
			loop = lb_create_block(p, "for.bit_set.loop")
			body := lb_create_block(p, "for.bit_set.body")
			done = lb_create_block(p, "for.bit_set.done")
			lb_emit_jump(p, loop)
			lb_start_block(p, loop)
			remaining_val := lb_addr_load(p, remaining)
			cond := lb_emit_comp(p, Token_NotEq, remaining_val, lb_zero(m, mask))
			lb_emit_if(p, cond, body, done)
			lb_start_block(p, body)
			val = lb_emit_count_trailing_zeros(p, remaining_val, mask)
			val = lb_emit_conv(p, val, elem)
			if rs.Reverse {
				val = lb_emit_arith(p, Token_Sub, lb_const_int(m, elem, uint64(et.BitSet.Lower+8*type_size_of(mask)-1)), val, elem)
			} else {
				val = lb_emit_arith(p, Token_Add, val, lb_const_int(m, elem, uint64(et.BitSet.Lower)), elem)
			}
			reduce_val := lb_emit_arith(p, Token_Sub, remaining_val, lb_const_int(m, mask, 1), mask)
			remaining_val = lb_emit_arith(p, Token_And, remaining_val, reduce_val, mask)
			lb_addr_store(p, remaining, remaining_val)

		default:
			gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Cannot range over type")
		}
	}
	if is_map {
		if val0_type != nil {
			lb_store_range_stmt_val(p, val0, key)
		}
		if val1_type != nil {
			lb_store_range_stmt_val(p, val1, val)
		}
	} else {
		if val0_type != nil {
			lb_store_range_stmt_val(p, val0, val)
		}
		if val1_type != nil {
			lb_store_range_stmt_val(p, val1, key)
		}
	}
	lb_push_target_list(p, rs.Label, done, loop, nil)
	lb_build_stmt(p, rs.Body)
	lb_close_scope(p, lbDeferExit_Default, nil, rs.Body)
	lb_pop_target_list(p)
	if p.DebugInfo != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_end_location_from_ast(p, rs.Body))
	}
	lb_emit_jump(p, loop)
	lb_start_block(p, done)
}

func lb_build_unroll_range_stmt(p *lbProcedure, rs *AstUnrollRangeStmt, scope *Scope) {
	m := p.Module
	lb_open_scope(p, scope)
	if rs.Init != nil {
		lb_build_stmt(p, rs.Init)
	}
	val0 := lb_strip_and_prefix(rs.Val0)
	val1 := lb_strip_and_prefix(rs.Val1)
	var val0_type *Type
	var val1_type *Type
	if val0 != nil && !is_blank_ident(val0) {
		val0_type = type_of_expr(val0)
	}
	if val1 != nil && !is_blank_ident(val1) {
		val1_type = type_of_expr(val1)
	}
	if val0_type != nil {
		e := entity_of_node(val0)
		lb_add_local(p, e.Type, e)
	}
	if val1_type != nil {
		e := entity_of_node(val1)
		lb_add_local(p, e.Type, e)
	}
	expr := unparen_expr(rs.Expr)
	tav := type_and_value_of_expr(expr)
	if is_ast_range(expr) {
		var val0_addr lbAddr
		var val1_addr lbAddr
		if val0_type != nil {
			val0_addr = lb_build_addr(p, val0)
		}
		if val1_type != nil {
			val1_addr = lb_build_addr(p, val1)
		}
		op := expr.BinaryExpr.Op.Kind
		start_expr := expr.BinaryExpr.Left
		end_expr := expr.BinaryExpr.Right
		if start_expr.TAV.Mode != Addressing_Constant {
			gb_assert_handler("Assertion Failure", "start_expr->tav.mode == Addressing_Constant", "llvm_backend_stmt_range.go", 0, 0)
		}
		if end_expr.TAV.Mode != Addressing_Constant {
			gb_assert_handler("Assertion Failure", "end_expr->tav.mode == Addressing_Constant", "llvm_backend_stmt_range.go", 0, 0)
		}
		start := start_expr.TAV.Value
		end := end_expr.TAV.Value
		if op != Token_RangeHalf {
			index := exact_value_i64(0)
			for v := start; compare_exact_values(Token_LtEq, v, end); v = exact_value_increment_one(v) {
				if val0_type != nil {
					lb_addr_store(p, val0_addr, lb_const_value(m, val0_type, v))
				}
				if val1_type != nil {
					lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, index))
				}
				lb_build_stmt(p, rs.Body)
				index = exact_value_increment_one(index)
			}
		} else {
			index := exact_value_i64(0)
			for v := start; compare_exact_values(Token_Lt, v, end); v = exact_value_increment_one(v) {
				if val0_type != nil {
					lb_addr_store(p, val0_addr, lb_const_value(m, val0_type, v))
				}
				if val1_type != nil {
					lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, index))
				}
				lb_build_stmt(p, rs.Body)
				index = exact_value_increment_one(index)
			}
		}
	} else if tav.Mode == Addressing_Type {
		if !is_type_enum(type_deref(tav.Type)) {
			gb_assert_handler("Assertion Failure", "is_type_enum(type_deref(tav.Type))", "llvm_backend_stmt_range.go", 0, 0)
		}
		et := type_deref(tav.Type)
		bet := base_type(et)
		var val0_addr lbAddr
		var val1_addr lbAddr
		if val0_type != nil {
			val0_addr = lb_build_addr(p, val0)
		}
		if val1_type != nil {
			val1_addr = lb_build_addr(p, val1)
		}
		for i := 0; i < len(bet.Enum.Fields); i++ {
			field := bet.Enum.Fields[i]
			if field.Kind != Entity_Constant {
				gb_assert_handler("Assertion Failure", "field->kind == Entity_Constant", "llvm_backend_stmt_range.go", 0, 0)
			}
			if val0_type != nil {
				lb_addr_store(p, val0_addr, lb_const_value(m, val0_type, field.Constant.Value))
			}
			if val1_type != nil {
				lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, exact_value_i64(int64(i))))
			}
			lb_build_stmt(p, rs.Body)
		}
	} else {
		var val0_addr lbAddr
		var val1_addr lbAddr
		if val0_type != nil {
			val0_addr = lb_build_addr(p, val0)
		}
		if val1_type != nil {
			val1_addr = lb_build_addr(p, val1)
		}
		unroll_count_ev := ExactValue{}
		if len(rs.Args) != 0 {
			unroll_count_ev = rs.Args[0].TAV.Value
		}
		if unroll_count_ev.Kind == ExactValue_Invalid {
			t := base_type(expr.TAV.Type)
			switch t.Kind {
			case Type_Basic:
				if expr.TAV.Mode != Addressing_Constant {
					gb_assert_handler("Assertion Failure", "expr->tav.mode == Addressing_Constant", "llvm_backend_stmt_range.go", 0, 0)
				}
				if !is_type_string(t) {
					gb_assert_handler("Assertion Failure", "is_type_string(t)", "llvm_backend_stmt_range.go", 0, 0)
				}
				value := expr.TAV.Value
				if value.Kind != ExactValue_String {
					gb_assert_handler("Assertion Failure", "value.kind == ExactValue_String", "llvm_backend_stmt_range.go", 0, 0)
				}
				str := goStr(value.ValueString)
				offset := 0
				for offset < len(str) {
					codepoint, width := utf8.DecodeRuneInString(str[offset:])
					if val0_type != nil {
						lb_addr_store(p, val0_addr, lb_const_value(m, val0_type, exact_value_i64(int64(codepoint))))
					}
					if val1_type != nil {
						lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, exact_value_i64(int64(offset))))
					}
					lb_build_stmt(p, rs.Body)
					offset += width
				}

			case Type_Array:
				if t.Array.Count > 0 {
					v := lb_build_expr(p, expr)
					val_addr := lb_address_from_load_or_generate_local(p, v)
					for i := int64(0); i < t.Array.Count; i++ {
						if val0_type != nil {
							elem := lb_emit_array_epi(p, val_addr, isize(i))
							lb_addr_store(p, val0_addr, lb_emit_load(p, elem))
						}
						if val1_type != nil {
							lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, exact_value_i64(i)))
						}
						lb_build_stmt(p, rs.Body)
					}
				}

			case Type_EnumeratedArray:
				if t.EnumeratedArray.Count > 0 {
					v := lb_build_expr(p, expr)
					val_addr := lb_address_from_load_or_generate_local(p, v)
					for i := int64(0); i < t.EnumeratedArray.Count; i++ {
						if val0_type != nil {
							elem := lb_emit_array_epi(p, val_addr, isize(i))
							lb_addr_store(p, val0_addr, lb_emit_load(p, elem))
						}
						if val1_type != nil {
							idx := exact_value_add(exact_value_i64(i), *t.EnumeratedArray.MinValue)
							lb_addr_store(p, val1_addr, lb_const_value(m, val1_type, idx))
						}
						lb_build_stmt(p, rs.Body)
					}
				}

			default:
				gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Invalid '#unroll for' type")
			}
		} else {
			unroll_count := exact_value_to_i64(unroll_count_ev)
			t := base_type(expr.TAV.Type)
			var data_ptr lbValue
			var count_ptr lbValue
			switch t.Kind {
			case Type_Slice, Type_DynamicArray:
				slice := lb_build_expr(p, expr)
				if is_type_pointer(slice.Type) {
					count_ptr = lb_emit_struct_ep(p, slice, 1)
					slice = lb_emit_load(p, slice)
				} else {
					count_ptr = lb_add_local_generated(p, t_int, false).Addr
					lb_emit_store(p, count_ptr, lb_slice_len(p, slice))
				}
				data_ptr = lb_emit_struct_ev(p, slice, 0)

			case Type_FixedCapacityDynamicArray:
				array := lb_build_expr(p, expr)
				if !is_type_pointer(array.Type) {
					array = lb_address_from_load_or_generate_local(p, array)
				}
				if !is_type_pointer(array.Type) {
					gb_assert_handler("Assertion Failure", "is_type_pointer(array.Type)", "llvm_backend_stmt_range.go", 0, 0)
				}
				data_ptr = lb_emit_conv(p, array, alloc_type_pointer(t.FixedCapacityDynamicArray.Elem))

			case Type_Array:
				array := lb_build_expr(p, expr)
				count_ptr = lb_add_local_generated(p, t_int, false).Addr
				lb_emit_store(p, count_ptr, lb_const_int(p.Module, t_int, uint64(t.Array.Count)))
				if !is_type_pointer(array.Type) {
					array = lb_address_from_load_or_generate_local(p, array)
				}
				if !is_type_pointer(array.Type) {
					gb_assert_handler("Assertion Failure", "is_type_pointer(array.Type)", "llvm_backend_stmt_range.go", 0, 0)
				}
				data_ptr = lb_emit_conv(p, array, alloc_type_pointer(t.Array.Elem))

			default:
				gb_assert_handler("Panic", nil, "llvm_backend_stmt_range.go", 0, "Invalid '#unroll for' type")
			}
			data_ptr.Type = alloc_type_multi_pointer_to_pointer(data_ptr.Type)
			loop_top := lb_create_block(p, "for.unroll.loop.top")
			body_top := lb_create_block(p, "for.unroll.body.top")
			body_bot := lb_create_block(p, "for.unroll.body.bot")
			done := lb_create_block(p, "for.unroll.done")
			loop_bot := done
			if unroll_count > 1 {
				loop_bot = lb_create_block(p, "for.unroll.loop.bot")
			}
			var val_entity *Entity
			if val0 != nil {
				val_entity = entity_of_node(val0)
			}
			var idx_entity *Entity
			if val1 != nil {
				idx_entity = entity_of_node(val1)
			}
			val_addr := lb_add_local(p, type_deref(data_ptr.Type, true), val_entity)
			idx_addr := lb_add_local(p, t_int, idx_entity)
			lb_addr_store(p, idx_addr, lb_const_nil(p.Module, t_int))
			lb_emit_jump(p, loop_top)
			lb_start_block(p, loop_top)
			idx_add_n := lb_addr_load(p, idx_addr)
			idx_add_n = lb_emit_arith(p, Token_Add, idx_add_n, lb_const_int(p.Module, t_int, uint64(unroll_count)), t_int)
			cond_top := lb_emit_comp(p, Token_LtEq, idx_add_n, lb_emit_load(p, count_ptr))
			lb_emit_if(p, cond_top, body_top, loop_bot)
			lb_start_block(p, body_top)
			for top := int64(0); top < unroll_count; top++ {
				index := lb_addr_load(p, idx_addr)
				v := lb_emit_load(p, lb_emit_ptr_offset(p, data_ptr, index))
				lb_addr_store(p, val_addr, v)
				lb_build_stmt(p, rs.Body)
				lb_emit_increment(p, lb_addr_get_ptr(p, idx_addr))
			}
			lb_emit_jump(p, loop_top)
			if unroll_count > 1 {
				lb_start_block(p, loop_bot)
				cond_bot := lb_emit_comp(p, Token_Lt, lb_addr_load(p, idx_addr), lb_emit_load(p, count_ptr))
				lb_emit_if(p, cond_bot, body_bot, done)
				lb_start_block(p, body_bot)
				{
					index := lb_addr_load(p, idx_addr)
					v := lb_emit_load(p, lb_emit_ptr_offset(p, data_ptr, index))
					lb_addr_store(p, val_addr, v)
					lb_build_stmt(p, rs.Body)
					lb_emit_increment(p, lb_addr_get_ptr(p, idx_addr))
				}
				lb_emit_jump(p, loop_bot)
			}
			lb_close_scope(p, lbDeferExit_Default, nil, rs.Body)
			lb_emit_jump(p, done)
			lb_start_block(p, done)
			return
		}
	}
	lb_close_scope(p, lbDeferExit_Default, nil, rs.Body)
}
