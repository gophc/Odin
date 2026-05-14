package cmd

// Set file/line/col in the first 3 elements of arr based on SourceCodeLocationInfo
func lb_set_file_line_col(p *lbProcedure, arr []lbValue, pos TokenPos) {
	file := get_file_path_string(pos.FileID)
	line := pos.Line
	col := pos.Column

	switch buildContext.SourceCodeLocationInfo {
	case SourceCodeLocationInfoNormal:
		// nothing
	case SourceCodeLocationInfoObfuscated:
		file = obfuscate_string(file, "F")
		line = obfuscate_i32(line)
		col = obfuscate_i32(col)
	case SourceCodeLocationInfoFilename:
		file = last_path_element(file)
	case SourceCodeLocationInfoNone:
		file = ""
		line = 0
		col = 0
	}

	arr[0] = lb_find_or_add_entity_string(p.Module, file, false)
	arr[1] = lb_const_int(p.Module, t_i32, int64(line))
	arr[2] = lb_const_int(p.Module, t_i32, int64(col))
}

func lb_bounds_check_disabled(p *lbProcedure) bool {
	if buildContext.NoBoundsCheck {
		return true
	}
	if (p.StateFlags & uint16(StateFlag_NoBoundsCheck)) != 0 {
		return true
	}
	return false
}

func lb_bounds_check_short_circuit(p *lbProcedure, index lbValue, len_ lbValue) bool {
	if lb_bounds_check_disabled(p) {
		return true
	}

	if LLVMIsConstant(index.Value) && LLVMIsConstant(len_.Value) {
		i := LLVMConstIntGetSExtValue(index.Value)
		n := LLVMConstIntGetSExtValue(len_.Value)
		if 0 <= i && i < n {
			return true
		}
	}

	if LLVMIsAInstruction(index.Value) {
		op := LLVMGetInstructionOpcode(index.Value)
		if op == LLVMURem {
			divisor := LLVMGetOperand(index.Value, 1)
			if divisor == len_.Value {
				return true
			}
		} else if op == LLVMAnd {
			mask := LLVMGetOperand(index.Value, 1)
			if LLVMIsConstant(mask) && LLVMIsConstant(len_.Value) {
				m := LLVMConstIntGetSExtValue(mask)
				l := LLVMConstIntGetSExtValue(len_.Value)
				if l > 0 && (l&(l-1)) == 0 && m == l-1 {
					return true
				}
			}
		}
	}
	return false
}

func lb_emit_bounds_check(p *lbProcedure, token Token, index lbValue, len_ lbValue) {
	if lb_bounds_check_short_circuit(p, index, len_) {
		return
	}

	index = lb_emit_conv(p, index, t_int)
	len_ = lb_emit_conv(p, len_, t_int)

	args := make([]lbValue, 5)
	lb_set_file_line_col(p, args, token.Pos)
	args[3] = index
	args[4] = len_

	lb_emit_runtime_call(p, "bounds_check_error", args)
}

func lb_emit_matrix_bounds_check(p *lbProcedure, token Token, row_index lbValue, column_index lbValue, row_count lbValue, column_count lbValue) {
	if lb_bounds_check_disabled(p) {
		return
	}

	row_index = lb_emit_conv(p, row_index, t_int)
	column_index = lb_emit_conv(p, column_index, t_int)
	row_count = lb_emit_conv(p, row_count, t_int)
	column_count = lb_emit_conv(p, column_count, t_int)

	args := make([]lbValue, 7)
	lb_set_file_line_col(p, args, token.Pos)
	args[3] = row_index
	args[4] = column_index
	args[5] = row_count
	args[6] = column_count

	lb_emit_runtime_call(p, "matrix_bounds_check_error", args)
}

func lb_emit_multi_pointer_slice_bounds_check(p *lbProcedure, token Token, low lbValue, high lbValue) {
	if lb_bounds_check_disabled(p) {
		return
	}

	if LLVMIsConstant(low.Value) && LLVMIsConstant(high.Value) {
		i := LLVMConstIntGetSExtValue(low.Value)
		n := LLVMConstIntGetSExtValue(high.Value)
		if i < n {
			return
		}
	}

	low = lb_emit_conv(p, low, t_int)
	high = lb_emit_conv(p, high, t_int)

	args := make([]lbValue, 5)
	lb_set_file_line_col(p, args, token.Pos)
	args[3] = low
	args[4] = high

	lb_emit_runtime_call(p, "multi_pointer_slice_expr_error", args)
}

func lb_emit_slice_bounds_check(p *lbProcedure, token Token, low lbValue, high lbValue, len_ lbValue, lower_value_used bool) {
	if lb_bounds_check_disabled(p) {
		return
	}
	if !lower_value_used && lb_bounds_check_short_circuit(p, high, len_) {
		return
	}

	high = lb_emit_conv(p, high, t_int)

	if !lower_value_used {
		args := make([]lbValue, 5)
		lb_set_file_line_col(p, args, token.Pos)
		args[3] = high
		args[4] = len_
		lb_emit_runtime_call(p, "slice_expr_error_hi", args)
	} else {
		low = lb_emit_conv(p, low, t_int)

		args := make([]lbValue, 6)
		lb_set_file_line_col(p, args, token.Pos)
		args[3] = low
		args[4] = high
		args[5] = len_
		lb_emit_runtime_call(p, "slice_expr_error_lo_hi", args)
	}
}

func lb_try_get_alignment(addr_ptr LLVMValueRef, default_alignment uint) uint {
	if LLVMIsAGlobalValue(addr_ptr) || LLVMIsAAllocaInst(addr_ptr) || LLVMIsALoadInst(addr_ptr) {
		return LLVMGetAlignment(addr_ptr)
	}
	return default_alignment
}

func lb_try_update_alignment(addr_ptr LLVMValueRef, alignment uint) bool {
	if LLVMIsAGlobalValue(addr_ptr) || LLVMIsAAllocaInst(addr_ptr) || LLVMIsALoadInst(addr_ptr) {
		if LLVMGetAlignment(addr_ptr) < alignment {
			if LLVMIsAAllocaInst(addr_ptr) {
				LLVMSetAlignment(addr_ptr, alignment)
			} else if LLVMIsAGlobalValue(addr_ptr) && LLVMGetLinkage(addr_ptr) != LLVMExternalLinkage {
				LLVMSetAlignment(addr_ptr, alignment)
			}
		}
		return LLVMGetAlignment(addr_ptr) >= alignment
	}
	return false
}

// convenience wrapper
func lb_try_update_alignment_ptr(ptr lbValue, alignment uint) bool {
	return lb_try_update_alignment(ptr.Value, alignment)
}

func lb_can_try_to_inline_array_arith(t *Type) bool {
	return type_size_of(t) <= buildContext.MaxSimdAlign
}

func lb_try_vector_cast(m *lbModule, ptr lbValue, vector_type_ *LLVMTypeRef) bool {
	array_type := base_type(type_deref(ptr.Type))
	if !is_type_array_like(array_type) {
		panic("is_type_array_like(array_type)")
	}
	count := get_array_type_count(array_type)
	elem_type := base_array_type(array_type)

	if lb_can_try_to_inline_array_arith(array_type) &&
		is_type_valid_vector_elem(elem_type) {
		possible := false
		vector_type := LLVMVectorType(lb_type(m, elem_type), uint(count))
		vector_alignment := uint(lb_alignof(vector_type))

		addr_ptr := ptr.Value
		if LLVMIsAAllocaInst(addr_ptr) || LLVMIsAGlobalValue(addr_ptr) {
			possible = lb_try_update_alignment(addr_ptr, vector_alignment)
		} else if LLVMIsALoadInst(addr_ptr) {
			alignment := LLVMGetAlignment(addr_ptr)
			possible = alignment >= vector_alignment
		}

		if possible {
			if vector_type_ != nil {
				*vector_type_ = vector_type
			}
			return true
		}
	}
	return false
}

func OdinLLVMBuildLoad(p *lbProcedure, typ LLVMTypeRef, value LLVMValueRef) LLVMValueRef {
	result := LLVMBuildLoad2(p.Builder, typ, value, "")

	if LLVMIsAInstruction(value) {
		is_packed := lb_get_metadata_custom_u64(p.Module, value, "odin-is-packed")
		if is_packed != 0 {
			LLVMSetAlignment(result, 1)
		}
		align := uint64(LLVMGetAlignment(result))
		align_min := lb_get_metadata_custom_u64(p.Module, value, "odin-min-align")
		align_max := lb_get_metadata_custom_u64(p.Module, value, "odin-max-align")
		if align_min != 0 && align < align_min {
			align = align_min
		}
		if align_max != 0 && align > align_max {
			align = align_max
		}
		LLVMSetAlignment(result, uint(align))
	}

	return result
}

func OdinLLVMBuildLoadAligned(p *lbProcedure, typ LLVMTypeRef, value LLVMValueRef, alignment int64) LLVMValueRef {
	result := LLVMBuildLoad2(p.Builder, typ, value, "")
	LLVMSetAlignment(result, uint(alignment))

	if LLVMIsAInstruction(value) {
		is_packed := lb_get_metadata_custom_u64(p.Module, value, "odin-is-packed")
		if is_packed != 0 {
			LLVMSetAlignment(result, 1)
		}
	}

	return result
}

func lb_addr_store(p *lbProcedure, addr lbAddr, value lbValue) {
	if addr.Addr.Value == 0 {
		return
	}
	if value.Type == nil {
		panic("value.type != nil")
	}

	if is_type_untyped_uninit(value.Type) {
		t := lb_addr_type(addr)
		value.Type = t
		value.Value = LLVMGetUndef(lb_type(p.Module, t))
	} else if is_type_untyped_nil(value.Type) {
		t := lb_addr_type(addr)
		value.Type = t
		value.Value = LLVMConstNull(lb_type(p.Module, t))
	}

	if addr.Kind == lbAddr_BitField {
		dst := addr.Addr
		if is_type_endian_big(addr.BitFieldType) {
			shift_amount := 8*type_size_of(value.Type) - addr.BitFieldSize
			shifted_value := value
			shifted_value.Value = LLVMBuildLShr(p.Builder,
				shifted_value.Value,
				LLVMConstInt(LLVMTypeOf(shifted_value.Value), uint64(shift_amount), false), "")

			src := lb_address_from_load_or_generate_local(p, shifted_value)

			args := make([]lbValue, 4)
			args[0] = dst
			args[1] = src
			args[2] = lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldOffset))
			args[3] = lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldSize))
			lb_emit_runtime_call(p, "__write_bits", args)
		} else if (addr.BitFieldOffset%8) == 0 &&
			(addr.BitFieldSize%8) == 0 {
			src := lb_address_from_load_or_generate_local(p, value)

			byte_offset := lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldOffset/8))
			byte_size := lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldSize/8))
			dst_offset := lb_emit_conv(p, dst, t_u8_ptr)
			dst_offset = lb_emit_ptr_offset(p, dst_offset, byte_offset)
			lb_mem_copy_non_overlapping(p, dst_offset, src, byte_size)
		} else {
			src := lb_address_from_load_or_generate_local(p, value)

			args := make([]lbValue, 4)
			args[0] = dst
			args[1] = src
			args[2] = lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldOffset))
			args[3] = lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldSize))
			lb_emit_runtime_call(p, "__write_bits", args)
		}
		return
	} else if addr.Kind == lbAddr_Map {
		lb_internal_dynamic_map_set(p, addr.Addr, addr.MapType, *addr.MapKey, value, p.CurrStmt)
		return
	} else if addr.Kind == lbAddr_Context {
		old_addr := lb_find_or_generate_context_ptr(p)

		create_new := true
		for i := range p.ContextStack {
			ctx_data := &p.ContextStack[i]
			if ctx_data.Ctx.Addr.Value == old_addr.Addr.Value {
				if ctx_data.Uses > 0 {
					create_new = true
				} else if p.ScopeIndex > ctx_data.ScopeIndex {
					create_new = true
				} else {
					create_new = false
				}
				break
			}
		}

		var next lbValue
		if create_new {
			old := lb_addr_load(p, old_addr)
			next_addr := lb_add_local_generated(p, t_context, true)
			lb_addr_store(p, next_addr, old)
			lb_push_context_onto_stack(p, next_addr)
			next = next_addr.Addr
		} else {
			next = old_addr.Addr
		}

		if len(addr.CtxSel.Index) > 0 {
			lhs := lb_emit_deep_field_gep(p, next, addr.CtxSel)
			rhs := lb_emit_conv(p, value, type_deref(lhs.Type))
			lb_emit_store(p, lhs, rhs)
		} else {
			lhs := next
			rhs := lb_emit_conv(p, value, lb_addr_type(addr))
			lb_emit_store(p, lhs, rhs)
		}
		return
	} else if addr.Kind == lbAddr_SoaVariable {
		t := type_deref(addr.Addr.Type)
		t = base_type(t)
		if !(t.Kind == Type_Struct && t.Struct.SoaKind != StructSoa_None) {
			panic("t->kind == Type_Struct && t->Struct.soa_kind != StructSoa_None")
		}
		elem_type := t.Struct.SoaElem
		value = lb_emit_conv(p, value, elem_type)
		elem_type = base_type(elem_type)

		index := addr.SoaIndex
		if !lb_is_const(index) || t.Struct.SoaKind != StructSoa_Fixed {
			t2 := base_type(type_deref(addr.Addr.Type))
			if !(t2.Kind == Type_Struct && t2.Struct.SoaKind != StructSoa_None) {
				panic("t->kind == Type_Struct && t->Struct.soa_kind != StructSoa_None")
			}
			len_ := lb_soa_struct_len(p, addr.Addr)
			if addr.SoaIndexExpr != nil && (!lb_is_const(addr.SoaIndex) || t2.Struct.SoaKind != StructSoa_Fixed) {
				lb_emit_bounds_check(p, ast_token(addr.SoaIndexExpr), addr.SoaIndex, len_)
			}
		}

		field_count := isize(0)
		switch elem_type.Kind {
		case Type_Struct:
			field_count = isize(len(elem_type.Struct.Fields))
		case Type_Array:
			field_count = isize(elem_type.Array.Count)
		}

		for i := isize(0); i < field_count; i++ {
			dst := lb_emit_struct_ep(p, addr.Addr, int32(i))
			src := lb_emit_struct_ev(p, value, int32(i))
			if t.Struct.SoaKind == StructSoa_Fixed {
				dst = lb_emit_array_ep(p, dst, index)
				lb_emit_store(p, dst, src)
			} else {
				field := lb_emit_load(p, dst)
				dst = lb_emit_ptr_offset(p, field, index)
				lb_emit_store(p, dst, src)
			}
		}
		return
	} else if addr.Kind == lbAddr_Swizzle {
		if addr.SwizzleCount > 4 {
			panic("addr.swizzle.count <= 4")
		}
		if value.Value == 0 {
			panic("value.value != nil")
		}
		value = lb_emit_conv(p, value, lb_addr_type(addr))

		dst := addr.Addr
		src := lb_address_from_load_or_generate_local(p, value)

		var src_ptrs [4]lbValue
		var src_loads [4]lbValue
		var dst_ptrs [4]lbValue

		for i := u8(0); i < addr.SwizzleCount; i++ {
			src_ptrs[i] = lb_emit_array_epi(p, src, isize(i))
		}
		for i := u8(0); i < addr.SwizzleCount; i++ {
			dst_ptrs[i] = lb_emit_array_epi(p, dst, isize(addr.SwizzleIndices[i]))
		}
		for i := u8(0); i < addr.SwizzleCount; i++ {
			src_loads[i] = lb_emit_load(p, src_ptrs[i])
		}
		for i := u8(0); i < addr.SwizzleCount; i++ {
			lb_emit_store(p, dst_ptrs[i], src_loads[i])
		}
		return
	} else if addr.Kind == lbAddr_SwizzleLarge {
		if value.Value == 0 {
			panic("value.value != nil")
		}
		value = lb_emit_conv(p, value, lb_addr_type(addr))

		dst := addr.Addr
		src := lb_address_from_load_or_generate_local(p, value)
		for i := range addr.SwizzleLargeIndices {
			src_ptr := lb_emit_array_epi(p, src, isize(i))
			dst_ptr := lb_emit_array_epi(p, dst, isize(addr.SwizzleLargeIndices[i]))
			src_load := lb_emit_load(p, src_ptr)
			lb_emit_store(p, dst_ptr, src_load)
		}
		return
	}

	if value.Value == 0 {
		panic("value.value != nil")
	}
	value = lb_emit_conv(p, value, lb_addr_type(addr))
	lb_emit_store(p, addr.Addr, value)
}

func lb_is_type_proc_recursive(t *Type) bool {
	for {
		if t == nil {
			return false
		}
		switch t.Kind {
		case Type_Named:
			t = t.Named.Base
		case Type_Pointer:
			t = t.Pointer.Elem
		case Type_Proc:
			return true
		default:
			return false
		}
	}
}

func lb_emit_store(p *lbProcedure, ptr lbValue, value lbValue) {
	if value.Value == 0 {
		panic("value.value != nil")
	}

	if LLVMIsUndef(value.Value) {
		return
	}

	a := type_deref(ptr.Type, true)
	if LLVMIsNull(value.Value) {
		src_t := llvm_addr_type(p.Module, ptr)
		if is_type_proc(a) {
			rawptr_type := lb_type(p.Module, t_rawptr)
			rawptr_ptr_type := LLVMPointerType(rawptr_type, 0)
			LLVMBuildStore(p.Builder, LLVMConstNull(rawptr_type), LLVMBuildBitCast(p.Builder, ptr.Value, rawptr_ptr_type, ""))
		} else if is_type_bit_set(a) {
			lb_mem_zero_ptr(p, ptr.Value, a, 1)
		} else if lb_sizeof(src_t) <= lb_max_zero_init_size() {
			LLVMBuildStore(p.Builder, LLVMConstNull(src_t), ptr.Value)
		} else {
			lb_mem_zero_ptr(p, ptr.Value, a, 1)
		}
		return
	}

	if is_type_boolean(a) {
		value = lb_emit_conv(p, value, a)
	}
	ca := core_type(a)
	if ca.Kind == Type_Basic {
		if !are_types_identical(ca, core_type(value.Type)) {
			panic("are_types_identical(ca, core_type(value.type))")
		}
	}

	const MAX_STORE_SIZE = 64

	if lb_sizeof(LLVMTypeOf(value.Value)) > MAX_STORE_SIZE {
		if !p.InMultiAssignment && LLVMIsALoadInst(value.Value) {
			dst_ptr := ptr.Value
			src_ptr_original := LLVMGetOperand(value.Value, 0)
			src_ptr := LLVMBuildPointerCast(p.Builder, src_ptr_original, LLVMTypeOf(dst_ptr), "")

			LLVMBuildMemMove(p.Builder,
				dst_ptr, lb_try_get_alignment(dst_ptr, 1),
				src_ptr, lb_try_get_alignment(src_ptr_original, 1),
				LLVMConstInt(LLVMInt64TypeInContext(p.Module.Ctx), uint64(lb_sizeof(LLVMTypeOf(value.Value))), false))
			return
		} else if LLVMIsConstant(value.Value) {
			addr2 := lb_add_global_generated_from_procedure(p, value.Type, value)
			lb_make_global_private_const(addr2)

			dst_ptr := ptr.Value
			src_ptr := addr2.Addr.Value
			src_ptr = LLVMBuildPointerCast(p.Builder, src_ptr, LLVMTypeOf(dst_ptr), "")

			LLVMBuildMemMove(p.Builder,
				dst_ptr, lb_try_get_alignment(dst_ptr, 1),
				src_ptr, lb_try_get_alignment(src_ptr, 1),
				LLVMConstInt(LLVMInt64TypeInContext(p.Module.Ctx), uint64(lb_sizeof(LLVMTypeOf(value.Value))), false))
			return
		}
	}

	var instr LLVMValueRef
	if lb_is_type_proc_recursive(a) {
		rawptr_type := lb_type(p.Module, t_rawptr)
		rawptr_ptr_type := LLVMPointerType(rawptr_type, 0)
		instr = LLVMBuildStore(p.Builder,
			LLVMBuildPointerCast(p.Builder, value.Value, rawptr_type, ""),
			LLVMBuildPointerCast(p.Builder, ptr.Value, rawptr_ptr_type, ""))
	} else {
		ca := core_type(a)
		if ca.Kind == Type_Basic || ca.Kind == Type_Proc {
			if !are_types_identical(ca, core_type(value.Type)) {
				panic("are_types_identical(ca, core_type(value.type))")
			}
		} else {
			if !are_types_identical(a, value.Type) {
				panic("are_types_identical(a, value.type)")
			}
		}
		instr = LLVMBuildStore(p.Builder, value.Value, ptr.Value)
	}
	_ = instr
}

// Helper stub - makes a global constant private
func lb_make_global_private_const(addr lbAddr) {
	_ = addr
}

func llvm_addr_type(module *lbModule, addr_val lbValue) LLVMTypeRef {
	return lb_type(module, type_deref(addr_val.Type))
}

func lb_emit_load(p *lbProcedure, value lbValue) lbValue {
	if value.Value == 0 {
		panic("value.value != nil")
	}
	if is_type_multi_pointer(value.Type) {
		vt := base_type(value.Type)
		if vt.Kind != Type_MultiPointer {
			panic("vt->kind == Type_MultiPointer")
		}
		t := vt.MultiPointer.Elem
		v := OdinLLVMBuildLoad(p, lb_type(p.Module, t), value.Value)
		return lbValue{Value: v, Type: t}
	} else if is_type_soa_pointer(value.Type) {
		ptr := lb_emit_struct_ev(p, value, 0)
		idx := lb_emit_struct_ev(p, value, 1)
		addr := lb_addr_soa_variable(ptr, idx, nil)
		return lb_addr_load(p, addr)
	}

	if !is_type_pointer(value.Type) {
		panic("is_type_pointer(value.type)")
	}
	t := type_deref(value.Type)
	v := OdinLLVMBuildLoad(p, lb_type(p.Module, t), value.Value)
	return lbValue{Value: v, Type: t}
}

func lb_addr_load(p *lbProcedure, addr lbAddr) lbValue {
	if addr.Addr.Value == 0 {
		panic("addr.addr.value != nil")
	}

	if addr.Kind == lbAddr_BitField {
		ct := core_type(addr.BitFieldType)
		do_mask := false
		if is_type_unsigned(ct) || is_type_boolean(ct) {
			if addr.BitFieldSize != 8*type_size_of(ct) {
				do_mask = true
			}
		}

		total_bitfield_bit_size := 8 * type_size_of(lb_addr_type(addr))
		dst_byte_size := type_size_of(addr.BitFieldType)
		dst := lb_add_local_generated(p, addr.BitFieldType, true)
		src := addr.Addr

		bit_offset := lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldOffset))
		bit_size := lb_const_int(p.Module, t_uintptr, uint64(addr.BitFieldSize))
		byte_offset := lb_const_int(p.Module, t_uintptr, uint64((addr.BitFieldOffset+7)/8))
		byte_size := lb_const_int(p.Module, t_uintptr, uint64((addr.BitFieldSize+7)/8))

		if !(type_size_of(addr.BitFieldType) >= ((addr.BitFieldSize + 7) / 8)) {
			panic("type_size_of(addr.bitfield.type) >= ((addr.bitfield.bit_size+7)/8)")
		}

		var r lbValue
		if is_type_endian_big(addr.BitFieldType) {
			args := make([]lbValue, 4)
			args[0] = dst.Addr
			args[1] = src
			args[2] = bit_offset
			args[3] = bit_size
			lb_emit_runtime_call(p, "__read_bits", args)

			shift_amount := LLVMConstInt(
				lb_type(p.Module, lb_addr_type(dst)),
				uint64(8*dst_byte_size-addr.BitFieldSize),
				false,
			)
			r = lb_addr_load(p, dst)
			r.Value = LLVMBuildShl(p.Builder, r.Value, shift_amount, "")
		} else if (addr.BitFieldOffset % 8) == 0 {
			do_mask = 8*dst_byte_size != addr.BitFieldSize

			copy_size := byte_size
			src_offset := lb_emit_conv(p, src, t_u8_ptr)
			src_offset = lb_emit_ptr_offset(p, src_offset, byte_offset)
			if addr.BitFieldOffset+8*dst_byte_size <= total_bitfield_bit_size {
				copy_size = lb_const_int(p.Module, t_uintptr, uint64(dst_byte_size))
			}
			lb_mem_copy_non_overlapping(p, dst.Addr, src_offset, copy_size, false)
			r = lb_addr_load(p, dst)
		} else {
			args := make([]lbValue, 4)
			args[0] = dst.Addr
			args[1] = src
			args[2] = bit_offset
			args[3] = bit_size
			lb_emit_runtime_call(p, "__read_bits", args)
			r = lb_addr_load(p, dst)
		}

		t := addr.BitFieldType

		if do_mask {
			if !(addr.BitFieldSize <= 8*type_size_of(ct)) {
				panic("addr.bitfield.bit_size <= 8*type_size_of(ct)")
			}
			mask := lb_const_int(p.Module, t, (uint64(1)<<uint64(addr.BitFieldSize))-1)
			r = lb_emit_arith(p, Token_And, r, mask, t)
		}

		if !is_type_unsigned(ct) && !is_type_boolean(ct) {
			m := lb_const_int(p.Module, t, uint64(1)<<(addr.BitFieldSize-1))
			r = lb_emit_arith(p, Token_Xor, r, m, t)
			r = lb_emit_arith(p, Token_Sub, r, m, t)
		}

		return r
	} else if addr.Kind == lbAddr_Map {
		map_type := base_type(type_deref(addr.Addr.Type))
		if map_type.Kind != Type_Map {
			panic("map_type->kind == Type_Map")
		}
		v := lb_add_local_generated(p, map_type.Map.LookupResultType, true)

		ptr := lb_internal_dynamic_map_get_ptr(p, addr.Addr, *addr.MapKey)
		ok := lb_emit_conv(p, lb_emit_comp_against_nil(p, Token_NotEq, ptr), t_bool)
		lb_emit_store(p, lb_emit_struct_ep(p, v.Addr, 1), ok)

		then_blk := lb_create_block(p, "map.get.then")
		done_blk := lb_create_block(p, "map.get.done")
		lb_emit_if(p, ok, then_blk, done_blk)
		lb_start_block(p, then_blk)
		{
			gep0 := lb_emit_struct_ep(p, v.Addr, 0)
			val := lb_emit_conv(p, ptr, gep0.Type)
			lb_emit_store(p, gep0, lb_emit_load(p, val))
		}
		lb_emit_jump(p, done_blk)
		lb_start_block(p, done_blk)

		if is_type_tuple(addr.MapResult) {
			return lb_addr_load(p, v)
		} else {
			single := lb_emit_struct_ep(p, v.Addr, 0)
			return lb_emit_load(p, single)
		}
	} else if addr.Kind == lbAddr_Context {
		a := addr.Addr
		for i := range p.ContextStack {
			ctx_data := &p.ContextStack[i]
			if ctx_data.Ctx.Addr.Value == a.Value {
				ctx_data.Uses += 1
				break
			}
		}
		a.Value = LLVMBuildPointerCast(p.Builder, a.Value, lb_type(p.Module, t_context_ptr), "")

		if len(addr.CtxSel.Index) > 0 {
			b := lb_emit_deep_field_gep(p, a, addr.CtxSel)
			return lb_emit_load(p, b)
		} else {
			return lb_emit_load(p, a)
		}
	} else if addr.Kind == lbAddr_SoaVariable {
		t := type_deref(addr.Addr.Type)
		t = base_type(t)
		if !(t.Kind == Type_Struct && t.Struct.SoaKind != StructSoa_None) {
			panic("t->kind == Type_Struct && t->Struct.soa_kind != StructSoa_None")
		}
		elem := t.Struct.SoaElem

		var len_ lbValue
		if t.Struct.SoaKind == StructSoa_Fixed {
			len_ = lb_const_int(p.Module, t_int, t.Struct.SoaCount)
		} else {
			v := lb_emit_load(p, addr.Addr)
			len_ = lb_soa_struct_len(p, v)
		}

		res := lb_add_local_generated(p, elem, true)

		if addr.SoaIndexExpr != nil && (!lb_is_const(addr.SoaIndex) || t.Struct.SoaKind != StructSoa_Fixed) {
			lb_emit_bounds_check(p, ast_token(addr.SoaIndexExpr), addr.SoaIndex, len_)
		}

		if t.Struct.SoaKind == StructSoa_Fixed {
			for i, field := range t.Struct.Fields {
				base_type := field.Type
				if base_type.Kind != Type_Array {
					panic("base_type->kind == Type_Array")
				}

				dst := lb_emit_struct_ep(p, res.Addr, int32(i))
				src_ptr := lb_emit_struct_ep(p, addr.Addr, int32(i))
				src_ptr = lb_emit_array_ep(p, src_ptr, addr.SoaIndex)
				src := lb_emit_load(p, src_ptr)
				lb_emit_store(p, dst, src)
			}
		} else {
			field_count := isize(len(t.Struct.Fields))
			if t.Struct.SoaKind == StructSoa_Slice {
				field_count -= 1
			} else if t.Struct.SoaKind == StructSoa_Dynamic {
				field_count -= 3
			}
			for i := isize(0); i < field_count; i++ {
				field := t.Struct.Fields[i]
				base_type := field.Type
				if base_type.Kind != Type_MultiPointer {
					panic("base_type->kind == Type_MultiPointer")
				}

				dst := lb_emit_struct_ep(p, res.Addr, int32(i))
				src_ptr := lb_emit_struct_ep(p, addr.Addr, int32(i))
				src := lb_emit_load(p, src_ptr)
				src = lb_emit_ptr_offset(p, src, addr.SoaIndex)
				src = lb_emit_load(p, src)
				lb_emit_store(p, dst, src)
			}
		}

		return lb_addr_load(p, res)
	} else if addr.Kind == lbAddr_Swizzle {
		array_type := base_type(addr.SwizzleType)
		if array_type.Kind == Type_SimdVector {
			vec := lb_emit_load(p, addr.Addr)
			index_count := addr.SwizzleCount
			if index_count == 0 {
				return vec
			}

			mask_len := uint(index_count)
			mask_elems := make([]LLVMValueRef, index_count)
			for i := u8(0); i < index_count; i++ {
				mask_elems[i] = LLVMConstInt(lb_type(p.Module, t_u32), uint64(addr.SwizzleIndices[i]), false)
			}

			mask := LLVMConstVector(mask_elems, mask_len)

			v1 := vec.Value
			v2 := vec.Value

			var res lbValue
			res.Type = addr.SwizzleType
			res.Value = LLVMBuildShuffleVector(p.Builder, v1, v2, mask, "")
			return res
		}

		if array_type.Kind != Type_Array {
			panic("array_type->kind == Type_Array")
		}

		res_align := uint(type_align_of(addr.SwizzleType))

		ordered_indices := [4]u8{0, 1, 2, 3}
		ordered_match := true
		for i := u8(0); i < addr.SwizzleCount; i++ {
			if ordered_indices[i] != addr.SwizzleIndices[i] {
				ordered_match = false
				break
			}
		}
		if ordered_match {
			if lb_try_update_alignment_ptr(addr.Addr, res_align) {
				pt := alloc_type_pointer(addr.SwizzleType)
				var res lbValue
				res.Value = LLVMBuildPointerCast(p.Builder, addr.Addr.Value, lb_type(p.Module, pt), "")
				res.Type = pt
				return lb_emit_load(p, res)
			}
		}

		res := lb_add_local_generated(p, addr.SwizzleType, false)
		ptr := lb_addr_get_ptr(p, res)
		if !is_type_pointer(ptr.Type) {
			panic("is_type_pointer(ptr.type)")
		}

		var vector_type LLVMTypeRef
		if lb_try_vector_cast(p.Module, addr.Addr, &vector_type) {
			vp := LLVMBuildPointerCast(p.Builder, addr.Addr.Value, LLVMPointerType(vector_type, 0), "")
			v := OdinLLVMBuildLoad(p, vector_type, vp)
			var scalars [4]LLVMValueRef
			for i := u8(0); i < addr.SwizzleCount; i++ {
				scalars[i] = LLVMConstInt(lb_type(p.Module, t_u32), uint64(addr.SwizzleIndices[i]), false)
			}
			mask := LLVMConstVector(scalars[:], addr.SwizzleCount)
			sv := llvm_basic_shuffle(p, v, mask)

			LLVMSetAlignment(res.Addr.Value, uint(lb_alignof(LLVMTypeOf(sv))))

			dst := LLVMBuildPointerCast(p.Builder, ptr.Value, LLVMPointerType(LLVMTypeOf(sv), 0), "")
			LLVMBuildStore(p.Builder, sv, dst)
		} else {
			for i := u8(0); i < addr.SwizzleCount; i++ {
				index := addr.SwizzleIndices[i]
				dst := lb_emit_array_epi(p, ptr, isize(i))
				src := lb_emit_array_epi(p, addr.Addr, isize(index))
				lb_emit_store(p, dst, lb_emit_load(p, src))
			}
		}
		return lb_addr_load(p, res)
	} else if addr.Kind == lbAddr_SwizzleLarge {
		array_type := base_type(addr.SwizzleLargeType)
		if array_type.Kind != Type_Array {
			panic("array_type->kind == Type_Array")
		}

		res := lb_add_local_generated(p, addr.SwizzleLargeType, false)
		ptr := lb_addr_get_ptr(p, res)
		if !is_type_pointer(ptr.Type) {
			panic("is_type_pointer(ptr.type)")
		}

		for i, index := range addr.SwizzleLargeIndices {
			dst := lb_emit_array_epi(p, ptr, isize(i))
			src := lb_emit_array_epi(p, addr.Addr, isize(index))
			lb_emit_store(p, dst, lb_emit_load(p, src))
		}

		return lb_addr_load(p, res)
	}

	if is_type_proc(addr.Addr.Type) {
		return addr.Addr
	}
	return lb_emit_load(p, addr.Addr)
}

func lb_const_union_tag(m *lbModule, u *Type, v *Type) lbValue {
	return lb_const_value(m, union_tag_type(u), exact_value_i64(union_variant_index_checked(u, v)))
}

func lb_emit_union_tag_ptr(p *lbProcedure, u lbValue) lbValue {
	t := u.Type
	if !(is_type_pointer(t) &&
		is_type_union(type_deref(t))) {
		panic("is_type_pointer(t) && is_type_union(type_deref(t))")
	}
	ut := type_deref(t)

	if is_type_union_maybe_pointer_original_alignment(ut) {
		panic("!is_type_union_maybe_pointer_original_alignment(ut)")
	}
	if is_type_union_maybe_pointer(ut) {
		panic("!is_type_union_maybe_pointer(ut)")
	}
	if !(type_size_of(ut) > 0) {
		panic("type_size_of(ut) > 0")
	}

	tag_type := union_tag_type(ut)

	uvt := llvm_addr_type(p.Module, u)
	element_count := LLVMCountStructElementTypes(uvt)
	if !(element_count >= 2) {
		panic("element_count >= 2")
	}

	ptr := u.Value
	ptr = LLVMBuildPointerCast(p.Builder, ptr, LLVMPointerType(uvt, 0), "")

	var tag_ptr lbValue
	tag_ptr.Value = LLVMBuildStructGEP2(p.Builder, uvt, ptr, 1, "")
	tag_ptr.Type = alloc_type_pointer(tag_type)
	return tag_ptr
}

func lb_emit_union_tag_value(p *lbProcedure, u lbValue) lbValue {
	ptr := lb_address_from_load_or_generate_local(p, u)
	tag_ptr := lb_emit_union_tag_ptr(p, ptr)
	return lb_emit_load(p, tag_ptr)
}

func lb_emit_store_union_variant_tag(p *lbProcedure, parent lbValue, variant_type *Type) {
	t := type_deref(parent.Type)
	if !is_type_union(t) {
		panic("is_type_union(t)")
	}

	if is_type_union_maybe_pointer(t) || type_size_of(t) == 0 {
		// No tag needed!
	} else {
		tag_ptr := lb_emit_union_tag_ptr(p, parent)
		tag := lb_const_union_tag(p.Module, t, variant_type)
		lb_emit_store(p, tag_ptr, tag)
	}
}

func lb_emit_store_union_variant(p *lbProcedure, parent lbValue, variant lbValue, variant_type *Type) {
	pt := base_type(type_deref(parent.Type))
	if pt.Kind != Type_Union {
		panic("pt->kind == Type_Union")
	}

	if !union_is_variant_of(pt, variant_type) {
		panic("union_is_variant_of(pt, variant_type)")
	}

	if pt.Union.Kind == UnionTypeSharedNil {
		if !(type_size_of(variant_type) != 0) {
			panic("type_size_of(variant_type)")
		}

		if_nil := lb_create_block(p, "shared_nil.if_nil")
		if_not_nil := lb_create_block(p, "shared_nil.if_not_nil")
		done := lb_create_block(p, "shared_nil.done")

		cond_is_nil := lb_emit_comp_against_nil(p, Token_CmpEq, variant)
		lb_emit_if(p, cond_is_nil, if_nil, if_not_nil)

		lb_start_block(p, if_nil)
		lb_emit_store(p, parent, lb_const_nil(p.Module, type_deref(parent.Type)))
		lb_emit_jump(p, done)

		lb_start_block(p, if_not_nil)
		underlying := lb_emit_conv(p, parent, alloc_type_pointer(variant_type))
		lb_emit_store(p, underlying, variant)
		lb_emit_store_union_variant_tag(p, parent, variant_type)
		lb_emit_jump(p, done)

		lb_start_block(p, done)
	} else {
		if type_size_of(variant_type) == 0 {
			alignment := uint(1)
			lb_mem_zero_ptr_internal(p, parent.Value, pt.Union.VariantBlockSize, alignment, false)
		} else {
			underlying := lb_emit_conv(p, parent, alloc_type_pointer(variant_type))
			lb_emit_store(p, underlying, variant)
		}
		lb_emit_store_union_variant_tag(p, parent, variant_type)
	}
}
