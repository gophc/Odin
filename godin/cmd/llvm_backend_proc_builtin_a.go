package cmd

func lb_build_builtin_proc(p *lbProcedure, expr *Ast, tv TypeAndValue, id BuiltinProcId) lbValue {
	if BuiltinProc__simd_begin < id && id < BuiltinProc__simd_end {
		return lb_build_builtin_simd_proc(p, expr, tv, id)
	}

	m := p.Module
	ce := &expr.CallExpr
	_ = ce

	switch id {
	case BuiltinProc_DIRECTIVE:
		bd := &ce.Proc.BasicDirective
		_ = bd
		name := bd.Name.String
		if name == "location" {
			procedure := p.Entity.Token.String
			pos := ast_token(ce.Proc).Pos
			if len(ce.Args) > 0 {
				ident := unselector_expr(ce.Args[0])
				e := entity_of_node(ident)
				parent_proc_decl := e.ParentProcDecl.Load()
				if parent_proc_decl != nil && parent_proc_decl.Entity != nil {
					procedure = parent_proc_decl.Entity.Load().Token.String
				} else {
					procedure = String{}
				}
				pos = e.Token.Pos
			}
			return lb_emit_source_code_location_as_global(p, procedure, pos)
		} else if name == "load_directory" {
			cache_ptr := map_must_get(&m.Info.LoadDirectoryMap, expr)
			cache := *cache_ptr
			count := len(cache.Files)
			elements := make([]LLVMValueRef, count)
			for i, file := range cache.Files {
				file_name := filename_without_directory(file.Path)
				values := [2]LLVMValueRef{
					lb_const_string(m, file_name).Value,
					lb_const_value(m, t_u8_slice, exact_value_string(file.Data)).Value,
				}
				elements[i] = llvm_const_named_struct(m, t_load_directory_file, values[:], isize(len(values)))
			}
			backing_array := llvm_const_array(m, lb_type(m, t_load_directory_file), elements, isize(count))
			array_type := alloc_type_array(t_load_directory_file, int64(count), nil)
			backing_array_addr := lb_add_global_generated_from_procedure(p, array_type, lbValue{Value: backing_array, Type: array_type})
			lb_make_global_private_const(backing_array_addr)
			backing_array_ptr := backing_array_addr.Addr.Value
			backing_array_ptr = LLVMConstPointerCast(backing_array_ptr, lb_type(m, t_load_directory_file_ptr))
			const_slice := llvm_const_slice_internal(m, backing_array_ptr, LLVMConstInt(lb_type(m, t_int), uint64(count), false))
			addr := lb_add_global_generated_from_procedure(p, tv.Type, lbValue{Value: const_slice, Type: t_load_directory_file_slice})
			lb_make_global_private_const(addr)
			return lb_addr_load(p, addr)
		} else {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "UNKNOWN DIRECTIVE: %.*s", int(name.Len), name.Data)
		}

	case BuiltinProc_type_info_of:
		arg := ce.Args[0]
		tav := type_and_value_of_expr(arg)
		if tav.Mode == Addressing_Type {
			t := default_type(type_of_expr(arg))
			return lb_type_info(p, t)
		}
		args := []lbValue{lb_build_expr(p, arg)}
		return lb_emit_runtime_call(p, "__type_info_of", args)

	case BuiltinProc_typeid_of:
		arg := ce.Args[0]
		tav := type_and_value_of_expr(arg)
		t := default_type(type_of_expr(arg))
		return lb_typeid(p.Module, t)

	case BuiltinProc_len:
		v := lb_build_expr(p, ce.Args[0])
		t := base_type(v.Type)
		if is_type_pointer(t) {
			v = lb_emit_load(p, v)
			t = type_deref(t)
		}
		if is_type_cstring(t) {
			return lb_cstring_len(p, v)
		} else if is_type_cstring16(t) {
			return lb_cstring16_len(p, v)
		} else if is_type_string16(t) {
			return lb_string_len(p, v)
		} else if is_type_string(t) {
			return lb_string_len(p, v)
		} else if is_type_array(t) {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Array lengths are constant")
		} else if is_type_slice(t) {
			return lb_slice_len(p, v)
		} else if is_type_dynamic_array(t) {
			return lb_dynamic_array_len(p, v)
		} else if is_type_fixed_capacity_dynamic_array(t) {
			return lb_fixed_capacity_dynamic_array_len(p, v)
		} else if is_type_map(t) {
			return lb_map_len(p, v)
		} else if is_type_soa_struct(t) {
			return lb_soa_struct_len(p, v)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unreachable")

	case BuiltinProc_cap:
		v := lb_build_expr(p, ce.Args[0])
		t := base_type(v.Type)
		if is_type_pointer(t) {
			v = lb_emit_load(p, v)
			t = type_deref(t)
		}
		if is_type_string(t) {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unreachable")
		} else if is_type_array(t) {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Array lengths are constant")
		} else if is_type_fixed_capacity_dynamic_array(t) {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Fixed capacity dynamic array capacities are constant")
		} else if is_type_slice(t) {
			return lb_slice_len(p, v)
		} else if is_type_dynamic_array(t) {
			return lb_dynamic_array_cap(p, v)
		} else if is_type_map(t) {
			return lb_map_cap(p, v)
		} else if is_type_soa_struct(t) {
			return lb_soa_struct_cap(p, v)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unreachable")

	case BuiltinProc_swizzle:
		index_count := len(ce.Args) - 1
		if is_type_simd_vector(tv.Type) {
			vec := lb_build_expr(p, ce.Args[0])
			if index_count == 0 {
				return vec
			}
			mask_len := uint(index_count)
			mask_elems := make([]LLVMValueRef, index_count)
			for i := 1; i < len(ce.Args); i++ {
				tv_arg := type_and_value_of_expr(ce.Args[i])
				index := uint32(big_int_to_i64(&tv_arg.Value.ValueInteger))
				mask_elems[i-1] = LLVMConstInt(lb_type(m, t_u32), uint64(index), false)
			}
			mask := LLVMConstVector(mask_elems, mask_len)
			v1 := vec.Value
			v2 := vec.Value
			var res lbValue
			res.Type = tv.Type
			res.Value = LLVMBuildShuffleVector(p.Builder, v1, v2, mask, "")
			return res
		}
		addr := lb_build_array_swizzle_addr(p, ce, tv)
		return lb_addr_load(p, addr)

	case BuiltinProc_complex:
		real := lb_build_expr(p, ce.Args[0])
		imag := lb_build_expr(p, ce.Args[1])
		dst_addr := lb_add_local_generated(p, tv.Type, false)
		dst := lb_addr_get_ptr(p, dst_addr)
		ft := base_complex_elem_type(tv.Type)
		real = lb_emit_conv(p, real, ft)
		imag = lb_emit_conv(p, imag, ft)
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 0), real)
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 1), imag)
		return lb_emit_load(p, dst)

	case BuiltinProc_quaternion:
		var xyzw [4]lbValue
		for i := 0; i < 4; i++ {
			f := &ce.Args[i].FieldValue
			field_name := f.Field.Ident.Token.String
			var index int32 = -1
			if field_name == "x" || field_name == "imag" {
				index = 0
			} else if field_name == "y" || field_name == "jmag" {
				index = 1
			} else if field_name == "z" || field_name == "kmag" {
				index = 2
			} else if field_name == "w" || field_name == "real" {
				index = 3
			}
			xyzw[index] = lb_build_expr(p, f.Value)
		}
		dst_addr := lb_add_local_generated(p, tv.Type, false)
		dst := lb_addr_get_ptr(p, dst_addr)
		ft := base_complex_elem_type(tv.Type)
		xyzw[0] = lb_emit_conv(p, xyzw[0], ft)
		xyzw[1] = lb_emit_conv(p, xyzw[1], ft)
		xyzw[2] = lb_emit_conv(p, xyzw[2], ft)
		xyzw[3] = lb_emit_conv(p, xyzw[3], ft)
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 0), xyzw[0])
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 1), xyzw[1])
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 2), xyzw[2])
		lb_emit_store(p, lb_emit_struct_ep(p, dst, 3), xyzw[3])
		return lb_emit_load(p, dst)

	case BuiltinProc_real:
		val := lb_build_expr(p, ce.Args[0])
		if is_type_complex(val.Type) {
			real := lb_emit_struct_ev(p, val, 0)
			return lb_emit_conv(p, real, tv.Type)
		} else if is_type_quaternion(val.Type) {
			real := lb_emit_struct_ev(p, val, 3)
			return lb_emit_conv(p, real, tv.Type)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "invalid type for real")
		return lbValue{}

	case BuiltinProc_imag:
		val := lb_build_expr(p, ce.Args[0])
		if is_type_complex(val.Type) {
			imag := lb_emit_struct_ev(p, val, 1)
			return lb_emit_conv(p, imag, tv.Type)
		} else if is_type_quaternion(val.Type) {
			imag := lb_emit_struct_ev(p, val, 0)
			return lb_emit_conv(p, imag, tv.Type)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "invalid type for imag")
		return lbValue{}

	case BuiltinProc_jmag:
		val := lb_build_expr(p, ce.Args[0])
		if is_type_quaternion(val.Type) {
			imag := lb_emit_struct_ev(p, val, 1)
			return lb_emit_conv(p, imag, tv.Type)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "invalid type for jmag")
		return lbValue{}

	case BuiltinProc_kmag:
		val := lb_build_expr(p, ce.Args[0])
		if is_type_quaternion(val.Type) {
			imag := lb_emit_struct_ev(p, val, 2)
			return lb_emit_conv(p, imag, tv.Type)
		}
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "invalid type for kmag")
		return lbValue{}

	case BuiltinProc_conj:
		val := lb_build_expr(p, ce.Args[0])
		return lb_emit_conjugate(p, val, tv.Type)

	case BuiltinProc_expand_values:
		val := lb_build_expr(p, ce.Args[0])
		t := base_type(val.Type)
		if !is_type_tuple(tv.Type) {
			if t.Kind == Type_Struct {
				return lb_emit_struct_ev(p, val, 0)
			} else if t.Kind == Type_Array {
				return lb_emit_struct_ev(p, val, 0)
			} else {
				gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unknown type of expand_values")
			}
		}
		tuple := lb_addr_get_ptr(p, lb_add_local_generated(p, tv.Type, false))
		if t.Kind == Type_Struct {
			for src_index, field := range t.Struct.Fields {
				field_index := field.Variable.FieldIndex
				f := lb_emit_struct_ev(p, val, field_index)
				ep := lb_emit_struct_ep(p, tuple, int32(src_index))
				lb_emit_store(p, ep, f)
			}
		} else if is_type_array_like(t) {
			ap := lb_address_from_load_or_generate_local(p, val)
			n := int32(get_array_type_count(t))
			for i := int32(0); i < n; i++ {
				f := lb_emit_load(p, lb_emit_array_epi(m, ap, isize(i)))
				ep := lb_emit_struct_ep(p, tuple, i)
				lb_emit_store(p, ep, f)
			}
		} else {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unknown type of expand_values")
		}
		return lb_emit_load(p, tuple)

	case BuiltinProc_compress_values:
		value_count := 0
		for _, arg := range ce.Args {
			t := arg.TAV.Type
			if is_type_tuple(t) {
				value_count += len(t.Tuple.Variables)
			} else {
				value_count += 1
			}
		}
		if value_count == 1 {
			x := lb_build_expr(p, ce.Args[0])
			x = lb_emit_conv(p, x, tv.Type)
			return x
		}
		dt := base_type(tv.Type)
		addr := lb_add_local_generated(p, tv.Type, true)
		if is_type_struct(dt) || is_type_tuple(dt) {
			index := int32(0)
			for _, arg := range ce.Args {
				x := lb_build_expr(p, arg)
				if is_type_tuple(x.Type) {
					for i := 0; i < len(x.Type.Tuple.Variables); i++ {
						y := lb_emit_tuple_ev(p, x, int32(i))
						ptr := lb_emit_struct_ep(p, addr.Addr, index)
						index++
						y = lb_emit_conv(p, y, type_deref(ptr.Type))
						lb_emit_store(p, ptr, y)
					}
				} else {
					ptr := lb_emit_struct_ep(p, addr.Addr, index)
					index++
					x = lb_emit_conv(p, x, type_deref(ptr.Type))
					lb_emit_store(p, ptr, x)
				}
			}
		} else if is_type_array_like(dt) {
			index := int32(0)
			for _, arg := range ce.Args {
				x := lb_build_expr(p, arg)
				if is_type_tuple(x.Type) {
					for i := 0; i < len(x.Type.Tuple.Variables); i++ {
						y := lb_emit_tuple_ev(p, x, int32(i))
						ptr := lb_emit_array_epi(m, addr.Addr, isize(index))
						index++
						y = lb_emit_conv(p, y, type_deref(ptr.Type))
						lb_emit_store(p, ptr, y)
					}
				} else {
					ptr := lb_emit_array_epi(m, addr.Addr, isize(index))
					index++
					x = lb_emit_conv(p, x, type_deref(ptr.Type))
					lb_emit_store(p, ptr, x)
				}
			}
		} else {
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "TODO(bill): compress_values -> %s", type_to_string(tv.Type))
		}
		return lb_addr_load(p, addr)

	case BuiltinProc_min:
		t := type_of_expr(expr)
		if len(ce.Args) == 2 {
			return lb_emit_min(p, t, lb_build_expr(p, ce.Args[0]), lb_build_expr(p, ce.Args[1]))
		} else {
			x := lb_build_expr(p, ce.Args[0])
			for i := 1; i < len(ce.Args); i++ {
				x = lb_emit_min(p, t, x, lb_build_expr(p, ce.Args[i]))
			}
			return x
		}

	case BuiltinProc_max:
		t := type_of_expr(expr)
		if len(ce.Args) == 2 {
			return lb_emit_max(p, t, lb_build_expr(p, ce.Args[0]), lb_build_expr(p, ce.Args[1]))
		} else {
			x := lb_build_expr(p, ce.Args[0])
			for i := 1; i < len(ce.Args); i++ {
				x = lb_emit_max(p, t, x, lb_build_expr(p, ce.Args[i]))
			}
			return x
		}

	case BuiltinProc_abs:
		x := lb_build_expr(p, ce.Args[0])
		t := x.Type
		if is_type_unsigned(t) {
			return x
		}
		if is_type_quaternion(t) {
			sz := int64(8) * type_size_of(t)
			args := []lbValue{x}
			switch sz {
			case 64:
				return lb_emit_runtime_call(p, "abs_quaternion64", args)
			case 128:
				return lb_emit_runtime_call(p, "abs_quaternion128", args)
			case 256:
				return lb_emit_runtime_call(p, "abs_quaternion256", args)
			}
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unknown complex type")
		} else if is_type_complex(t) {
			sz := int64(8) * type_size_of(t)
			args := []lbValue{x}
			switch sz {
			case 32:
				return lb_emit_runtime_call(p, "abs_complex32", args)
			case 64:
				return lb_emit_runtime_call(p, "abs_complex64", args)
			case 128:
				return lb_emit_runtime_call(p, "abs_complex128", args)
			}
			gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unknown complex type")
		} else if is_type_float(t) {
			little := is_type_endian_little(t) || (is_type_endian_platform(t) && build_context.EndianKind == TargetEndian_Little)
			var t_unsigned *Type
			var mask lbValue
			switch type_size_of(t) {
			case 2:
				t_unsigned = t_u16
				if little {
					mask = lb_const_int(m, t_unsigned, 0x7FFF)
				} else {
					mask = lb_const_int(m, t_unsigned, 0xFF7F)
				}
			case 4:
				t_unsigned = t_u32
				if little {
					mask = lb_const_int(m, t_unsigned, 0x7FFFFFFF)
				} else {
					mask = lb_const_int(m, t_unsigned, 0xFFFFFF7F)
				}
			case 8:
				t_unsigned = t_u64
				if little {
					mask = lb_const_int(m, t_unsigned, 0x7FFFFFFFFFFFFFFF)
				} else {
					mask = lb_const_int(m, t_unsigned, 0xFFFFFFFFFFFFFF7F)
				}
			default:
				gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "abs: unhandled float size")
			}
			as_unsigned := lb_emit_transmute(p, x, t_unsigned)
			abs := lb_emit_arith(p, Token_And, as_unsigned, mask, t_unsigned)
			return lb_emit_transmute(p, abs, t)
		}
		zero := lb_const_nil(m, t)
		cond := lb_emit_comp(p, Token_Lt, x, zero)
		neg := lb_emit_unary_arith(p, Token_Sub, x, t)
		return lb_emit_select(p, cond, neg, x)

	case BuiltinProc_clamp:
		return lb_emit_clamp(p, type_of_expr(expr),
			lb_build_expr(p, ce.Args[0]),
			lb_build_expr(p, ce.Args[1]),
			lb_build_expr(p, ce.Args[2]))

	case BuiltinProc_soa_zip:
		return lb_soa_zip(p, ce, tv)

	case BuiltinProc_soa_unzip:
		return lb_soa_unzip(p, ce, tv)

	case BuiltinProc_transpose:
		m_val := lb_build_expr(p, ce.Args[0])
		return lb_emit_matrix_transpose(p, m_val, tv.Type)

	case BuiltinProc_outer_product:
		a := lb_build_expr(p, ce.Args[0])
		b := lb_build_expr(p, ce.Args[1])
		return lb_emit_outer_product(p, a, b, tv.Type)

	case BuiltinProc_hadamard_product:
		a := lb_build_expr(p, ce.Args[0])
		b := lb_build_expr(p, ce.Args[1])
		if is_type_array(tv.Type) {
			return lb_emit_arith(p, Token_Mul, a, b, tv.Type)
		}
		return lb_emit_arith_matrix(p, Token_Mul, a, b, tv.Type, true)

	case BuiltinProc_matrix_flatten:
		m_val := lb_build_expr(p, ce.Args[0])
		return lb_emit_matrix_flatten(p, m_val, tv.Type)

	case BuiltinProc_unreachable:
		lb_emit_unreachable(p)
		return lbValue{}

	case BuiltinProc_raw_data:
		x := lb_build_expr(p, ce.Args[0])
		t := base_type(x.Type)
		var res lbValue
		switch t.Kind {
		case Type_Slice:
			res = lb_slice_elem(p, x)
			res = lb_emit_conv(p, res, tv.Type)
		case Type_DynamicArray:
			res = lb_dynamic_array_elem(p, x)
			res = lb_emit_conv(p, res, tv.Type)
		case Type_Basic:
			if t.Basic.Kind == Basic_string {
				res = lb_string_elem(p, x)
				res = lb_emit_conv(p, res, tv.Type)
			} else if t.Basic.Kind == Basic_cstring {
				res = lb_emit_conv(p, x, tv.Type)
			} else if t.Basic.Kind == Basic_string16 {
				res = lb_string_elem(p, x)
				res = lb_emit_conv(p, res, tv.Type)
			} else if t.Basic.Kind == Basic_cstring16 {
				res = lb_emit_conv(p, x, tv.Type)
			}
		case Type_Pointer:
			fallthrough
		case Type_MultiPointer:
			res = lb_emit_conv(p, x, tv.Type)
		}
		return res

	case BuiltinProc_alloca:
		sz := lb_build_expr(p, ce.Args[0])
		al := exact_value_to_i64(type_and_value_of_expr(ce.Args[1]).Value)
		var res lbValue
		res.Type = alloc_type_multi_pointer(t_u8)
		res.Value = LLVMBuildArrayAlloca(p.Builder, lb_type(m, t_u8), sz.Value, "")
		LLVMSetAlignment(res.Value, uint(al))
		return res

	case BuiltinProc_cpu_relax:
		if build_context.Metrics.Arch == TargetArch_i386 ||
			build_context.Metrics.Arch == TargetArch_amd64 {
			func_type := LLVMFunctionType(LLVMVoidTypeInContext(m.Ctx), nil, 0, false)
			the_asm := llvm_get_inline_asm(func_type, "pause", "", true, false, LLVMInlineAsmDialectAT_T)
			LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
		} else if build_context.Metrics.Arch == TargetArch_arm64 {
			func_type := LLVMFunctionType(LLVMVoidTypeInContext(m.Ctx), nil, 0, false)
			the_asm := llvm_get_inline_asm(func_type, "isb", "", true, false, LLVMInlineAsmDialectAT_T)
			LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
		} else {
			func_type := LLVMFunctionType(LLVMVoidTypeInContext(m.Ctx), nil, 0, false)
			the_asm := llvm_get_inline_asm(func_type, "", "", true, false, LLVMInlineAsmDialectAT_T)
			LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
		}
		return lbValue{}

	case BuiltinProc_debug_trap:
		fallthrough
	case BuiltinProc_trap:
		var name string
		switch id {
		case BuiltinProc_debug_trap:
			name = "llvm.debugtrap"
		case BuiltinProc_trap:
			name = "llvm.trap"
		}
		lb_call_intrinsic(p, name, nil, 0, nil, 0)
		if id == BuiltinProc_trap {
			LLVMBuildUnreachable(p.Builder)
		}
		return lbValue{}

	case BuiltinProc_read_cycle_counter:
		var res lbValue
		res.Type = tv.Type
		if build_context.Metrics.Arch == TargetArch_arm64 {
			func_type := LLVMFunctionType(LLVMInt64TypeInContext(m.Ctx), nil, 0, false)
			the_asm := llvm_get_inline_asm(func_type, "mrs $0, cntvct_el0", "=r", false, false, LLVMInlineAsmDialectAT_T)
			res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
		} else {
			res.Value = lb_call_intrinsic(p, "llvm.readcyclecounter", nil, 0, nil, 0)
		}
		return res

	case BuiltinProc_read_cycle_counter_frequency:
		var res lbValue
		res.Type = tv.Type
		if build_context.Metrics.Arch == TargetArch_arm64 {
			func_type := LLVMFunctionType(LLVMInt64TypeInContext(m.Ctx), nil, 0, false)
			the_asm := llvm_get_inline_asm(func_type, "mrs $0, cntfrq_el0", "=r", false, false, LLVMInlineAsmDialectAT_T)
			res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
		}
		return res

	case BuiltinProc_count_trailing_zeros:
		return lb_emit_count_trailing_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_count_leading_zeros:
		return lb_emit_count_leading_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_count_trailing_ones:
		return lb_emit_count_trailing_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_count_leading_ones:
		return lb_emit_count_leading_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_count_ones:
		return lb_emit_count_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_count_zeros:
		return lb_emit_count_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_reverse_bits:
		return lb_emit_reverse_bits(p, lb_build_expr(p, ce.Args[0]), tv.Type)
	case BuiltinProc_byte_swap:
		x := lb_build_expr(p, ce.Args[0])
		x = lb_emit_conv(p, x, tv.Type)
		return lb_emit_byte_swap(p, x, tv.Type)

	case BuiltinProc_overflow_add:
		fallthrough
	case BuiltinProc_overflow_sub:
		fallthrough
	case BuiltinProc_overflow_mul:
		main_type := tv.Type
		type_ := main_type
		if is_type_tuple(main_type) {
			type_ = main_type.Tuple.Variables[0].Type
		}
		x := lb_build_expr(p, ce.Args[0])
		y := lb_build_expr(p, ce.Args[1])
		x = lb_emit_conv(p, x, type_)
		y = lb_emit_conv(p, y, type_)
		var name string
		if is_type_unsigned(type_) {
			switch id {
			case BuiltinProc_overflow_add:
				name = "llvm.uadd.with.overflow"
			case BuiltinProc_overflow_sub:
				name = "llvm.usub.with.overflow"
			case BuiltinProc_overflow_mul:
				name = "llvm.umul.with.overflow"
			}
		} else {
			switch id {
			case BuiltinProc_overflow_add:
				name = "llvm.sadd.with.overflow"
			case BuiltinProc_overflow_sub:
				name = "llvm.ssub.with.overflow"
			case BuiltinProc_overflow_mul:
				name = "llvm.smul.with.overflow"
			}
		}
		types := [1]LLVMTypeRef{lb_type(m, type_)}
		args := [2]LLVMValueRef{x.Value, y.Value}
		var res lbValue
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
		if is_type_tuple(main_type) {
			res_type := alloc_type_tuple()
			res_type.Tuple.Variables = make([]*Entity, 2)
			res_type.Tuple.Variables[0] = alloc_entity_field(nil, BlankToken, type_, false, 0)
			res_type.Tuple.Variables[1] = alloc_entity_field(nil, BlankToken, t_llvm_bool, false, 1)
			res.Type = res_type
		} else {
			res.Value = LLVMBuildExtractValue(p.Builder, res.Value, 0, "")
			res.Type = type_
		}
		return res

	case BuiltinProc_mem_copy:
		dst := lb_build_expr(p, ce.Args[0])
		src := lb_build_expr(p, ce.Args[1])
		len_ := lb_build_expr(p, ce.Args[2])
		lb_mem_copy_overlapping(p, dst, src, len_, false)
		return lbValue{}

	case BuiltinProc_mem_copy_non_overlapping:
		dst := lb_build_expr(p, ce.Args[0])
		src := lb_build_expr(p, ce.Args[1])
		len_ := lb_build_expr(p, ce.Args[2])
		lb_mem_copy_non_overlapping(p, dst, src, len_, false)
		return lbValue{}

	case BuiltinProc_mem_zero:
		ptr := lb_build_expr(p, ce.Args[0])
		len_ := lb_build_expr(p, ce.Args[1])
		ptr = lb_emit_conv(p, ptr, t_rawptr)
		len_ = lb_emit_conv(p, len_, t_int)
		alignment := uint(1)
		lb_mem_zero_ptr_internal(p, ptr.Value, len_.Value, alignment, false)
		return lbValue{}

	case BuiltinProc_mem_zero_volatile:
		ptr := lb_build_expr(p, ce.Args[0])
		len_ := lb_build_expr(p, ce.Args[1])
		ptr = lb_emit_conv(p, ptr, t_rawptr)
		len_ = lb_emit_conv(p, len_, t_int)
		alignment := uint(1)
		lb_mem_zero_ptr_internal(p, ptr.Value, len_.Value, alignment, true)
		return lbValue{}

	case BuiltinProc_ptr_offset:
		ptr := lb_build_expr(p, ce.Args[0])
		len_ := lb_build_expr(p, ce.Args[1])
		len_ = lb_emit_conv(p, len_, t_int)
		return lb_emit_ptr_offset(p, ptr, len_)

	case BuiltinProc_ptr_sub:
		elem0 := type_deref(type_of_expr(ce.Args[0]), true)
		elem1 := type_deref(type_of_expr(ce.Args[1]), true)
		elem := elem0
		ptr0 := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_uintptr)
		ptr1 := lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_uintptr)
		ptr0 = lb_emit_conv(p, ptr0, t_int)
		ptr1 = lb_emit_conv(p, ptr1, t_int)
		diff := lb_emit_arith(p, Token_Sub, ptr0, ptr1, t_int)
		return lb_emit_arith(p, Token_Quo, diff, lb_const_int(m, t_int, uint64(type_size_of(elem))), t_int)

	case BuiltinProc_atomic_thread_fence:
		LLVMBuildFence(p.Builder, llvm_atomic_ordering_from_odin(ce.Args[0].TAV.Value), false, "")
		return lbValue{}

	case BuiltinProc_atomic_signal_fence:
		LLVMBuildFence(p.Builder, llvm_atomic_ordering_from_odin(ce.Args[0].TAV.Value), true, "")
		return lbValue{}

	case BuiltinProc_volatile_store:
		fallthrough
	case BuiltinProc_non_temporal_store:
		fallthrough
	case BuiltinProc_atomic_store:
		fallthrough
	case BuiltinProc_atomic_store_explicit:
		dst := lb_build_expr(p, ce.Args[0])
		val := lb_build_expr(p, ce.Args[1])
		val = lb_emit_conv(p, val, type_deref(dst.Type))
		instr := LLVMBuildStore(p.Builder, val.Value, dst.Value)
		switch id {
		case BuiltinProc_non_temporal_store:
			kind_id := LLVMGetMDKindIDInContext(m.Ctx, "nontemporal", 11)
			node := LLVMValueAsMetadata(LLVMConstInt(lb_type(m, t_u32), 1, false))
			LLVMSetMetadata(instr, kind_id, LLVMMetadataAsValue(m.Ctx, node))
		case BuiltinProc_volatile_store:
			LLVMSetVolatile(instr, LLVMBool(true))
		case BuiltinProc_atomic_store:
			LLVMSetOrdering(instr, LLVMAtomicOrderingSequentiallyConsistent)
			LLVMSetVolatile(instr, LLVMBool(true))
		case BuiltinProc_atomic_store_explicit:
			ordering := llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
			LLVMSetOrdering(instr, ordering)
			LLVMSetVolatile(instr, LLVMBool(true))
		}
		LLVMSetAlignment(instr, uint(type_align_of(type_deref(dst.Type))))
		return lbValue{}

	case BuiltinProc_volatile_load:
		fallthrough
	case BuiltinProc_non_temporal_load:
		fallthrough
	case BuiltinProc_atomic_load:
		fallthrough
	case BuiltinProc_atomic_load_explicit:
		dst := lb_build_expr(p, ce.Args[0])
		instr := OdinLLVMBuildLoad(p, lb_type(m, type_deref(dst.Type)), dst.Value)
		switch id {
		case BuiltinProc_non_temporal_load:
			kind_id := LLVMGetMDKindIDInContext(m.Ctx, "nontemporal", 11)
			node := LLVMValueAsMetadata(LLVMConstInt(lb_type(m, t_u32), 1, false))
			LLVMSetMetadata(instr, kind_id, LLVMMetadataAsValue(m.Ctx, node))
		case BuiltinProc_volatile_load:
			LLVMSetVolatile(instr, LLVMBool(true))
		case BuiltinProc_atomic_load:
			LLVMSetOrdering(instr, LLVMAtomicOrderingSequentiallyConsistent)
			LLVMSetVolatile(instr, LLVMBool(true))
		case BuiltinProc_atomic_load_explicit:
			ordering := llvm_atomic_ordering_from_odin(ce.Args[1].TAV.Value)
			LLVMSetOrdering(instr, ordering)
			LLVMSetVolatile(instr, LLVMBool(true))
		}
		LLVMSetAlignment(instr, uint(type_align_of(type_deref(dst.Type))))
		var res lbValue
		res.Value = instr
		res.Type = type_deref(dst.Type)
		return res

	case BuiltinProc_unaligned_store:
		dst := lb_build_expr(p, ce.Args[0])
		src := lb_build_expr(p, ce.Args[1])
		t := type_deref(dst.Type)
		if is_type_simd_vector(t) {
			store := LLVMBuildStore(p.Builder, src.Value, dst.Value)
			LLVMSetAlignment(store, 1)
		} else {
			src = lb_address_from_load_or_generate_local(p, src)
			lb_mem_copy_non_overlapping(p, dst, src, lb_const_int(m, t_int, uint64(type_size_of(t))), false)
		}
		return lbValue{}

	case BuiltinProc_unaligned_load:
		src := lb_build_expr(p, ce.Args[0])
		t := type_deref(src.Type)
		if is_type_simd_vector(t) {
			var res lbValue
			res.Type = t
			res.Value = OdinLLVMBuildLoadAligned(p, lb_type(m, t), src.Value, 1)
			return res
		} else {
			dst := lb_add_local_generated(p, t, false)
			lb_mem_copy_non_overlapping(p, dst.Addr, src, lb_const_int(m, t_int, uint64(type_size_of(t))), false)
			return lb_addr_load(p, dst)
		}

	case BuiltinProc_atomic_add:
		fallthrough
	case BuiltinProc_atomic_sub:
		fallthrough
	case BuiltinProc_atomic_and:
		fallthrough
	case BuiltinProc_atomic_nand:
		fallthrough
	case BuiltinProc_atomic_or:
		fallthrough
	case BuiltinProc_atomic_xor:
		fallthrough
	case BuiltinProc_atomic_exchange:
		fallthrough
	case BuiltinProc_atomic_add_explicit:
		fallthrough
	case BuiltinProc_atomic_sub_explicit:
		fallthrough
	case BuiltinProc_atomic_and_explicit:
		fallthrough
	case BuiltinProc_atomic_nand_explicit:
		fallthrough
	case BuiltinProc_atomic_or_explicit:
		fallthrough
	case BuiltinProc_atomic_xor_explicit:
		fallthrough
	case BuiltinProc_atomic_exchange_explicit:
		dst := lb_build_expr(p, ce.Args[0])
		val := lb_build_expr(p, ce.Args[1])
		val = lb_emit_conv(p, val, type_deref(dst.Type))
		var op LLVMAtomicRMWBinOp
		var ordering LLVMAtomicOrdering
		switch id {
		case BuiltinProc_atomic_add:
			op = LLVMAtomicRMWBinOpAdd
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_sub:
			op = LLVMAtomicRMWBinOpSub
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_and:
			op = LLVMAtomicRMWBinOpAnd
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_nand:
			op = LLVMAtomicRMWBinOpNand
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_or:
			op = LLVMAtomicRMWBinOpOr
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_xor:
			op = LLVMAtomicRMWBinOpXor
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_exchange:
			op = LLVMAtomicRMWBinOpXchg
			ordering = LLVMAtomicOrderingSequentiallyConsistent
		case BuiltinProc_atomic_add_explicit:
			op = LLVMAtomicRMWBinOpAdd
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_sub_explicit:
			op = LLVMAtomicRMWBinOpSub
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_and_explicit:
			op = LLVMAtomicRMWBinOpAnd
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_nand_explicit:
			op = LLVMAtomicRMWBinOpNand
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_or_explicit:
			op = LLVMAtomicRMWBinOpOr
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_xor_explicit:
			op = LLVMAtomicRMWBinOpXor
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		case BuiltinProc_atomic_exchange_explicit:
			op = LLVMAtomicRMWBinOpXchg
			ordering = llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		}
		var res lbValue
		res.Value = LLVMBuildAtomicRMW(p.Builder, op, dst.Value, val.Value, ordering, false)
		res.Type = tv.Type
		LLVMSetVolatile(res.Value, LLVMBool(true))
		return res

	case BuiltinProc_atomic_compare_exchange_strong:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_weak:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_strong_explicit:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_weak_explicit:
		address := lb_build_expr(p, ce.Args[0])
		elem := type_deref(address.Type)
		old_value := lb_build_expr(p, ce.Args[1])
		new_value := lb_build_expr(p, ce.Args[2])
		old_value = lb_emit_conv(p, old_value, elem)
		new_value = lb_emit_conv(p, new_value, elem)
		var success_ordering LLVMAtomicOrdering
		var failure_ordering LLVMAtomicOrdering
		var weak LLVMBool
		switch id {
		case BuiltinProc_atomic_compare_exchange_strong:
			success_ordering = LLVMAtomicOrderingSequentiallyConsistent
			failure_ordering = LLVMAtomicOrderingSequentiallyConsistent
			weak = LLVMBool(false)
		case BuiltinProc_atomic_compare_exchange_weak:
			success_ordering = LLVMAtomicOrderingSequentiallyConsistent
			failure_ordering = LLVMAtomicOrderingSequentiallyConsistent
			weak = LLVMBool(true)
		case BuiltinProc_atomic_compare_exchange_strong_explicit:
			success_ordering = llvm_atomic_ordering_from_odin(ce.Args[3].TAV.Value)
			failure_ordering = llvm_atomic_ordering_from_odin(ce.Args[4].TAV.Value)
			weak = LLVMBool(false)
		case BuiltinProc_atomic_compare_exchange_weak_explicit:
			success_ordering = llvm_atomic_ordering_from_odin(ce.Args[3].TAV.Value)
			failure_ordering = llvm_atomic_ordering_from_odin(ce.Args[4].TAV.Value)
			weak = LLVMBool(true)
		}
		single_threaded := LLVMBool(false)
		value := LLVMBuildAtomicCmpXchg(
			p.Builder, address.Value,
			old_value.Value, new_value.Value,
			success_ordering,
			failure_ordering,
			single_threaded,
		)
		LLVMSetWeak(value, weak)
		LLVMSetVolatile(value, LLVMBool(true))
		if is_type_tuple(tv.Type) {
			fix_typed := alloc_type_tuple()
			fix_typed.Tuple.Variables = make([]*Entity, 2)
			fix_typed.Tuple.Variables[0] = tv.Type.Tuple.Variables[0]
			fix_typed.Tuple.Variables[1] = alloc_entity_field(nil, BlankToken, t_llvm_bool, false, 1)
			var res lbValue
			res.Value = value
			res.Type = fix_typed
			return res
		} else {
			var res lbValue
			res.Value = LLVMBuildExtractValue(p.Builder, value, 0, "")
			res.Type = tv.Type
			return res
		}

	case BuiltinProc_type_equal_proc:
		return lb_equal_proc_for_type(m, ce.Args[0].TAV.Type)
	case BuiltinProc_type_hasher_proc:
		return lb_hasher_proc_for_type(m, ce.Args[0].TAV.Type)
	case BuiltinProc_type_map_info:
		return lb_gen_map_info_ptr(m, ce.Args[0].TAV.Type)
	case BuiltinProc_type_map_cell_info:
		return lb_gen_map_cell_info_ptr(m, ce.Args[0].TAV.Type)

	default:
		gb_assert_handler("Panic", "", "llvm_backend_proc_builtin_a.go", 0, "Unhandled built-in procedure %.*s", int(builtin_procs[id].Name.Len), builtin_procs[id].Name.Data)
		return lbValue{}
	}
}
