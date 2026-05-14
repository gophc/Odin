package cmd

import (
	"math"
	"unsafe"
)

func lb_is_const(value lbValue) bool {
	v := value.Value
	if is_type_untyped_nil(value.Type) {
		return true
	}
	if LLVMIsConstant(v) != 0 {
		return true
	}
	return false
}

func lb_is_const_or_global(value lbValue) bool {
	if lb_is_const(value) {
		return true
	}
	return false
}

func lb_is_elem_const(elem *Ast, elem_type *Type) bool {
	if !elem_type_can_be_constant(elem_type) {
		return false
	}
	if elem.Kind == Ast_FieldValue {
		elem = elem.FieldValue.Value
	}
	tav := type_and_value_of_expr(elem)
	if tav.Mode == Addressing_Invalid {
		gb_assert_handler("Assertion Failure", "tav.mode != Addressing_Invalid", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 29, "%s %s", expr_to_string(elem), type_to_string(tav.Type))
	}
	return tav.Value.Kind != ExactValue_Invalid
}

func lb_is_const_nil(value lbValue) bool {
	v := value.Value
	if v != 0 && LLVMIsConstant(v) != 0 {
		if LLVMIsAConstantAggregateZero(v) != 0 {
			return true
		} else if LLVMIsAConstantPointerNull(v) != 0 {
			return true
		}
	}
	return false
}

func lb_is_expr_constant_zero(expr *Ast) bool {
	if expr == nil {
		gb_assert_handler("Assertion Failure", "expr != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 48)
	}
	v := exact_value_to_integer(expr.TAV.Value)
	if v.Kind == ExactValue_Integer {
		return big_int_cmp_zero(&v.ValueInteger) == 0
	}
	return false
}

func lb_get_const_string(m *lbModule, value lbValue) string {
	gb_assert_handler("Assertion Failure", "lb_is_const(value)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 57)
	gb_assert_handler("Assertion Failure", "LLVMIsConstant(value.value)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 58)
	t := base_type(value.Type)
	gb_assert_handler("Assertion Failure", "are_types_identical(t, t_string)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 61)
	ptr_indices := [1]uint{0}
	len_indices := [1]uint{1}
	underlying_ptr := llvm_const_extract_value(m, value.Value, ptr_indices[:], 1)
	underlying_len := llvm_const_extract_value(m, value.Value, len_indices[:], 1)
	gb_assert_handler("Assertion Failure", "LLVMGetConstOpcode(underlying_ptr) == LLVMGetElementPtr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 70)
	underlying_ptr = LLVMGetOperand(underlying_ptr, 0)
	gb_assert_handler("Assertion Failure", "LLVMIsAGlobalVariable(underlying_ptr)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 72)
	underlying_ptr = LLVMGetInitializer(underlying_ptr)
	var length uintptr
	text := LLVMGetAsString(underlying_ptr, &length)
	real_length := int64(LLVMConstIntGetSExtValue(underlying_len))
	return unsafe.String(text, int(real_length))
}

func llvm_const_cast(m *lbModule, val LLVMValueRef, dst LLVMTypeRef, failure_ *bool) LLVMValueRef {
	src := LLVMTypeOf(val)
	if src == dst {
		return val
	}
	if LLVMIsNull(val) != 0 {
		return LLVMConstNull(dst)
	}
	gb_assert_handler("Assertion Failure", "lb_sizeof(dst) == lb_sizeof(src)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 93, "%s vs %s", LLVMPrintTypeToString(dst), LLVMPrintTypeToString(src))
	kind := LLVMGetTypeKind(dst)
	switch kind {
	case LLVMPointerTypeKind:
		return LLVMConstPointerCast(val, dst)
	case LLVMStructTypeKind:
		src_n := LLVMCountStructElementTypes(src)
		dst_n := LLVMCountStructElementTypes(dst)
		if src_n != dst_n {
			goto failure
		}
		field_vals := make([]LLVMValueRef, dst_n)
		for i := uint(0); i < dst_n; i++ {
			field_val := llvm_const_extract_value(m, val, []uint{i}, 1)
			if field_val == 0 {
				goto failure
			}
			dst_elem_ty := LLVMStructGetTypeAtIndex(dst, i)
			field_vals[i] = llvm_const_cast(m, field_val, dst_elem_ty, failure_)
			if failure_ != nil && *failure_ {
				goto failure
			}
		}
		if LLVMIsLiteralStruct(dst) == 0 {
			return LLVMConstNamedStruct(dst, field_vals, dst_n)
		} else {
			return LLVMConstStructInContext(m.Ctx, field_vals, dst_n, LLVMIsPackedStruct(dst))
		}
	}
failure:
	if failure_ != nil {
		*failure_ = true
	}
	return val
}

func lb_const_ptr_cast(m *lbModule, value lbValue, t *Type) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_internally_pointer_like(value.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 129)
	gb_assert_handler("Assertion Failure", "is_type_internally_pointer_like(t)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 130)
	gb_assert_handler("Assertion Failure", "lb_is_const(value)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 131)
	res := lbValue{}
	res.Value = LLVMConstPointerCast(value.Value, lb_type(m, t))
	res.Type = t
	return res
}

func llvm_const_string16_internal(m *lbModule, t *Type, data LLVMValueRef, len LLVMValueRef) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "is_type_string16(t)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 159)
	if buildContext.PtrSize < buildContext.IntSize {
		values := [3]LLVMValueRef{
			data,
			LLVMConstNull(lb_type(m, t_i32)),
			len,
		}
		return llvm_const_named_struct_internal(m, lb_type(m, t), values[:], 3)
	} else {
		values := [2]LLVMValueRef{
			data,
			len,
		}
		return llvm_const_named_struct_internal(m, lb_type(m, t), values[:], 2)
	}
}

func llvm_const_named_struct(m *lbModule, t *Type, values []LLVMValueRef, value_count isize) LLVMValueRef {
	struct_type := lb_type(m, t)
	gb_assert_handler("Assertion Failure", "LLVMGetTypeKind(struct_type) == LLVMStructTypeKind", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 179)
	value_count_u := uint(value_count)
	elem_count := LLVMCountStructElementTypes(struct_type)
	if elem_count == value_count_u {
		return llvm_const_named_struct_internal(m, struct_type, values, value_count)
	}
	bt := base_type(t)
	gb_assert_handler("Assertion Failure", "bt->kind == Type_Struct || bt->kind == Type_Union", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 187)
	gb_assert_handler("Assertion Failure", "bt->kind != Type_Struct || value_count_ == bt->Struct.fields.count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 188)

	field_remapping := lb_get_struct_remapping(m, t)
	values_with_padding_count := elem_count
	values_with_padding := make([]LLVMValueRef, values_with_padding_count)
	for i := uint(0); i < value_count_u; i++ {
		values_with_padding[field_remapping[i]] = values[i]
	}
	for i := uint(0); i < values_with_padding_count; i++ {
		if values_with_padding[i] == 0 {
			values_with_padding[i] = LLVMConstNull(LLVMStructGetTypeAtIndex(struct_type, i))
		}
	}
	return llvm_const_named_struct_internal(m, struct_type, values_with_padding, isize(values_with_padding_count))
}

func llvm_const_named_struct_internal(m *lbModule, t LLVMTypeRef, values []LLVMValueRef, value_count isize) LLVMValueRef {
	value_count_u := uint(value_count)
	elem_count := LLVMCountStructElementTypes(t)
	gb_assert_handler("Assertion Failure", "value_count == elem_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 209, "%s %u %u", LLVMPrintTypeToString(t), value_count_u, elem_count)
	failure := false
	for i := uint(0); i < elem_count; i++ {
		elem_type := LLVMStructGetTypeAtIndex(t, i)
		values[i] = llvm_const_cast(m, values[i], elem_type, &failure)
	}
	if failure {
		return LLVMConstStructInContext(m.Ctx, values, value_count_u, 1)
	}
	return LLVMConstNamedStruct(t, values, value_count_u)
}

func llvm_const_array(m *lbModule, elem_type LLVMTypeRef, values []LLVMValueRef, value_count isize) LLVMValueRef {
	value_count_u := uint(value_count)
	failure := false
	for i := uint(0); i < value_count_u; i++ {
		values[i] = llvm_const_cast(m, values[i], elem_type, &failure)
	}
	if failure {
		return LLVMConstStructInContext(m.Ctx, values, value_count_u, 0)
	}
	for i := uint(0); i < value_count_u; i++ {
		if elem_type != LLVMTypeOf(values[i]) {
			return LLVMConstStructInContext(m.Ctx, values, value_count_u, 0)
		}
	}
	return LLVMConstArray(elem_type, values, value_count_u)
}

func llvm_const_slice_internal(m *lbModule, data LLVMValueRef, len LLVMValueRef) LLVMValueRef {
	if buildContext.PtrSize < buildContext.IntSize {
		gb_assert_handler("Assertion Failure", "build_context.metrics.ptr_size == 4", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 242)
		gb_assert_handler("Assertion Failure", "build_context.metrics.int_size == 8", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 243)
		vals := [3]LLVMValueRef{
			data,
			LLVMConstNull(lb_type(m, t_u32)),
			len,
		}
		return LLVMConstStructInContext(m.Ctx, vals[:], 3, 0)
	} else {
		vals := [2]LLVMValueRef{
			data,
			len,
		}
		return LLVMConstStructInContext(m.Ctx, vals[:], 2, 0)
	}
}

func llvm_const_slice(m *lbModule, data lbValue, len lbValue) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "is_type_pointer(data.type) || is_type_multi_pointer(data.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 259)
	gb_assert_handler("Assertion Failure", "are_types_identical(len.type, t_int)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 260)
	return llvm_const_slice_internal(m, data.Value, len.Value)
}

func lb_const_nil(m *lbModule, typ *Type) lbValue {
	v := LLVMConstNull(lb_type(m, typ))
	return lbValue{Value: v, Type: typ}
}

func lb_const_undef(m *lbModule, typ *Type) lbValue {
	v := LLVMGetUndef(lb_type(m, typ))
	return lbValue{Value: v, Type: typ}
}

func lb_const_int(m *lbModule, typ *Type, value u64) lbValue {
	res := lbValue{}
	signed := 1
	if is_type_unsigned(typ) {
		signed = 0
	}
	res.Value = LLVMConstInt(lb_type(m, typ), uint64(value), signed)
	res.Type = typ
	return res
}

func lb_const_string(m *lbModule, value String) lbValue {
	return lb_const_value(m, t_string, exact_value_string(value))
}

func lb_const_string16(m *lbModule, value String16) lbValue {
	return lb_const_value(m, t_string16, exact_value_string16(value))
}

func lb_const_bool(m *lbModule, typ *Type, value bool) lbValue {
	res := lbValue{}
	v := 0
	if value {
		v = 1
	}
	res.Value = LLVMConstInt(lb_type(m, typ), uint64(v), 0)
	res.Type = typ
	return res
}

func lb_const_f16(m *lbModule, f f32, typ ...*Type) LLVMValueRef {
	t := t_f16
	if len(typ) > 0 {
		t = typ[0]
	}
	gb_assert_handler("Assertion Failure", "type_size_of(type) == 2", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 303)
	u := f32_to_f16(f)
	if is_type_different_to_arch_endianness(t) {
		u = gb_endian_swap16(u)
	}
	i := LLVMConstInt(LLVMInt16TypeInContext(m.Ctx), uint64(u), 0)
	return LLVMConstBitCast(i, lb_type(m, t))
}

func lb_const_f32(m *lbModule, f f32, typ ...*Type) LLVMValueRef {
	t := t_f32
	if len(typ) > 0 {
		t = typ[0]
	}
	gb_assert_handler("Assertion Failure", "type_size_of(type) == 4", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 314)
	u := math.Float32bits(f)
	if is_type_different_to_arch_endianness(t) {
		u = gb_endian_swap32(u)
	}
	i := LLVMConstInt(LLVMInt32TypeInContext(m.Ctx), uint64(u), 0)
	return LLVMConstBitCast(i, lb_type(m, t))
}

func lb_is_expr_untyped_const(expr *Ast) bool {
	tv := type_and_value_of_expr(expr)
	if is_type_untyped(tv.Type) {
		return tv.Value.Kind != ExactValue_Invalid
	}
	return false
}

func lb_expr_untyped_const_to_typed(m *lbModule, expr *Ast, t *Type) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_typed(t)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 335)
	tv := type_and_value_of_expr(expr)
	return lb_const_value(m, t, tv.Value)
}

func lb_const_source_code_location_const(m *lbModule, procedure string, pos TokenPos) lbValue {
	file := get_file_path_string(pos.FileId)
	line := pos.Line
	column := pos.Column
	var procStr String
	if len(procedure) > 0 {
		procStr = String{Text: unsafe.StringData(procedure), Len: isize(len(procedure))}
	}
	switch buildContext.SourceCodeLocationInfo {
	case SourceCodeLocationInfo_Normal:
	case SourceCodeLocationInfo_Obfuscated:
		file = obfuscate_string(file, "F")
		procStr = obfuscate_string(procStr, "P")
		line = obfuscate_i32(line)
		column = obfuscate_i32(column)
	case SourceCodeLocationInfo_Filename:
		file = last_path_element(file)
	case SourceCodeLocationInfo_None:
		file = String{Text: strPtr(""), Len: 0}
		procStr = String{Text: strPtr(""), Len: 0}
		line = 0
		column = 0
	}
	fields := [4]LLVMValueRef{}
	fields[0] = lb_find_or_add_entity_string(m, goStr(file), false).Value
	fields[1] = lb_const_int(m, t_i32, u64(line)).Value
	fields[2] = lb_const_int(m, t_i32, u64(column)).Value
	fields[3] = lb_find_or_add_entity_string(m, goStr(procStr), false).Value
	res := lbValue{}
	res.Value = llvm_const_named_struct(m, t_source_code_location, fields[:], 4)
	res.Type = t_source_code_location
	return res
}

func lb_emit_source_code_location_const(p *lbProcedure, procedure string, pos TokenPos) lbValue {
	return lb_const_source_code_location_const(p.Module, procedure, pos)
}

func lb_emit_source_code_location_const_from_node(p *lbProcedure, node *Ast) lbValue {
	proc_name := ""
	if p.Entity != nil {
		proc_name = goStr(p.Entity.Token.String)
	}
	pos := TokenPos{}
	if node != nil {
		pos = ast_token(node).Pos
	}
	return lb_emit_source_code_location_const(p, proc_name, pos)
}

func lb_source_code_location_gen_name_from_pos(procedure string, pos TokenPos) String {
	s := gb_string_make(permanent_allocator(), "scl$[")
	if len(procedure) > 0 {
		s = gb_string_append_length(s, unsafe.StringData(procedure), isize(len(procedure)))
	}
	if pos.Offset != 0 {
		s = gb_string_append_fmt(s, "%d", pos.Offset)
	} else {
		s = gb_string_append_fmt(s, "%d_%d", pos.Line, pos.Column)
	}
	s = gb_string_appendc(s, "]")
	return make_string(s, gb_string_length(s))
}

func lb_source_code_location_gen_name(p *lbProcedure, node *Ast) String {
	proc_name := ""
	if p.Entity != nil {
		proc_name = goStr(p.Entity.Token.String)
	}
	pos := TokenPos{}
	if node != nil {
		pos = ast_token(node).Pos
	}
	return lb_source_code_location_gen_name_from_pos(proc_name, pos)
}

func lb_emit_source_code_location_as_global_ptr(p *lbProcedure, procedure string, pos TokenPos) lbValue {
	loc := lb_emit_source_code_location_const(p, procedure, pos)
	name := lb_source_code_location_gen_name_from_pos(procedure, pos)
	addr := lb_add_global_generated_with_name(p.Module, loc.Type, loc, goStr(name))
	lb_make_global_private_const(addr.Addr.Value)
	return addr.Addr
}

func lb_const_source_code_location_as_global_ptr(m *lbModule, procedure string, pos TokenPos) lbValue {
	loc := lb_const_source_code_location_const(m, procedure, pos)
	name := lb_source_code_location_gen_name_from_pos(procedure, pos)
	addr := lb_add_global_generated_with_name(m, loc.Type, loc, goStr(name))
	lb_make_global_private_const(addr.Addr.Value)
	return addr.Addr
}

func lb_emit_source_code_location_as_global_ptr_from_node(p *lbProcedure, node *Ast) lbValue {
	loc := lb_emit_source_code_location_const_from_node(p, node)
	name := lb_source_code_location_gen_name(p, node)
	addr := lb_add_global_generated_with_name(p.Module, loc.Type, loc, goStr(name))
	lb_make_global_private_const(addr.Addr.Value)
	return addr.Addr
}

func lb_emit_source_code_location_as_global(p *lbProcedure, procedure string, pos TokenPos) lbValue {
	return lb_emit_load(p, lb_emit_source_code_location_as_global_ptr(p, procedure, pos))
}

func lb_emit_source_code_location_as_global_from_node(p *lbProcedure, node *Ast) lbValue {
	return lb_emit_load(p, lb_emit_source_code_location_as_global_ptr_from_node(p, node))
}

func lb_build_constant_array_values(m *lbModule, typ *Type, elem_type *Type, count isize, values []LLVMValueRef, cc lbConstContext) LLVMValueRef {
	if cc.AllowLocal {
		cc.IsRodata = false
	}
	is_local := cc.AllowLocal && m.CurrProcedure != nil
	is_const := true
	if is_local {
		for i := isize(0); i < count; i++ {
			gb_assert_handler("Assertion Failure", "values[i] != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 470)
			if LLVMIsConstant(values[i]) == 0 {
				is_const = false
				break
			}
		}
	}
	if !is_const {
		llvm_elem_type := lb_type(m, elem_type)
		p := m.CurrProcedure
		gb_assert_handler("Assertion Failure", "p != nullptr", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 481)
		v := lb_add_local_generated(p, typ, false)
		ptr := lb_addr_get_ptr(p, v)
		for i := isize(0); i < count; i++ {
			elem := lb_emit_array_epi(p, ptr, i)
			if is_type_proc(elem_type) {
				values[i] = LLVMConstPointerCast(values[i], llvm_elem_type)
			}
			LLVMBuildStore(p.Builder, values[i], elem.Value)
		}
		return lb_addr_load(p, v).Value
	}
	return llvm_const_array(m, lb_type(m, elem_type), values, count)
}

func lb_big_int_to_llvm(m *lbModule, original_type *Type, a *BigInt) LLVMValueRef {
	if big_int_is_zero(a) {
		return LLVMConstNull(lb_type(m, original_type))
	}
	val := BigInt{}
	big_int_init(&val, a)
	if big_int_is_neg(&val) {
		mp_add_d(&val, 1, &val)
	}
	sz := uintptr(type_size_of(original_type))
	rop64 := [4]u64{}
	rop := (*byte)(unsafe.Pointer(&rop64))
	var written uintptr
	var max_count uintptr
	var size uintptr = 1
	nails := uintptr(0)
	endian := MP_LITTLE_ENDIAN
	max_count = mp_pack_count(&val, nails, size)
	if sz < max_count {
		debug_print_big_int(a)
		gb_printf_err("%s -> %tu\n", type_to_string(original_type), sz)
	}
	gb_assert_handler("Assertion Failure", "sz >= max_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 524, "max_count: %tu, sz: %tu, written: %tu, type %s", max_count, sz, written, type_to_string(original_type))
	gb_assert_handler("Assertion Failure", "(isize)(sizeof(rop64)) >= sz", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 525)
	err := mp_pack(rop, sz, &written,
		MP_LSB_FIRST,
		size, endian, nails,
		&val)
	gb_assert_handler("Assertion Failure", "err == MP_OKAY", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 531)
	if !is_type_endian_little(original_type) {
		for i := uintptr(0); i < sz/2; i++ {
			tmp := rop64[i]
			rop64[i] = rop64[sz-1-i]
			rop64[sz-1-i] = tmp
		}
	}
	if big_int_is_neg(a) {
		for i := uintptr(0); i < uintptr(len(rop64)); i++ {
			rop64[i] = ^rop64[i]
		}
	}
	big_int_dealloc(&val)
	gb_assert_handler("Assertion Failure", "!is_type_array(original_type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 550)
	value := LLVMConstIntOfArbitraryPrecision(lb_type(m, original_type), uint((sz+7)/8), (*u64)(unsafe.Pointer(rop)))
	return value
}

func lb_is_nested_possibly_constant(ft *Type, sel Selection, elem *Ast) bool {
	gb_assert_handler("Assertion Failure", "!sel.indirect", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 558)
	for _, index := range sel.Index {
		bt := base_type(ft)
		switch bt.Kind {
		case Type_Struct:
			if bt.Struct.IsRawUnion {
				return false
			}
			ft = bt.Struct.Fields[index].Type
		case Type_Array:
			ft = bt.Array.Elem
		default:
			return false
		}
	}
	if is_type_raw_union(ft) {
		return false
	}
	return lb_is_elem_const(elem, ft)
}

func lb_const_array_spread(m *lbModule, cc lbConstContext, array *Type, value ExactValue, res *lbValue) {
	gb_assert_handler("Assertion Failure", "array->kind == Type_Array", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 682)
	count := array.Array.Count
	elem := array.Array.Elem
	single_elem := lb_const_value(m, elem, value, cc)
	elems := make([]LLVMValueRef, count)
	for i := int64(0); i < count; i++ {
		elems[i] = single_elem.Value
	}
	res.Value = llvm_const_array(m, lb_type(m, elem), elems, isize(count))
}

func lb_fill_fixed_capacity_dynamic_array(m *lbModule, elem_count int64, original_type *Type, values []LLVMValueRef, cc lbConstContext) LLVMValueRef {
	bt := base_type(original_type)
	gb_assert_handler("Assertion Failure", "bt->kind == Type_FixedCapacityDynamicArray", "G:\\b0pass-win\\Odin\\src\\llvm_backend_const.cpp", 699)
	elem_type := bt.FixedCapacityDynamicArray.Elem
	capacity := bt.FixedCapacityDynamicArray.Capacity
	array_backing_type := alloc_type_array(elem_type, capacity, nil)
	array_backing := lb_build_constant_array_values(m, array_backing_type, elem_type, isize(capacity), values, cc)
	array_len := lb_const_int(m, t_int, u64(elem_count)).Value
	svalue_count := 0
	svalues := [3]LLVMValueRef{}
	svalues[svalue_count] = array_backing
	svalue_count++
	padding := bt.FixedCapacityDynamicArray.PaddingNeeded
	if padding > 0 {
		svalues[svalue_count] = LLVMConstNull(lb_type_padding_filler(m, padding, 1))
		svalue_count++
	}
	svalues[svalue_count] = array_len
	svalue_count++
	return llvm_const_named_struct(m, original_type, svalues[:], isize(svalue_count))
}
