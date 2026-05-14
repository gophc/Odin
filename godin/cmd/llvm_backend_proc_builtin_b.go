package cmd

import "fmt"

func lb_build_builtin_proc_typeid_of(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	tav := type_and_value_of_expr(ce.Args[0])
	gb_assert_handler("Assertion Failure", "tav.Mode == Addressing_Type", "src/llvm_backend_proc.cpp", 2778, 0)
	t := default_type(type_of_expr(ce.Args[0]))
	return lb_typeid(p.Module, t)
}

func lb_build_builtin_proc_len(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
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
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2799, "Array lengths are constant")
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
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2812, "Unreachable")
	return lbValue{}
}

func lb_build_builtin_proc_cap(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	v := lb_build_expr(p, ce.Args[0])
	t := base_type(v.Type)
	if is_type_pointer(t) {
		v = lb_emit_load(p, v)
		t = type_deref(t)
	}
	if is_type_string(t) {
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2823, "Unreachable")
	} else if is_type_array(t) {
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2826, "Array lengths are constant")
	} else if is_type_fixed_capacity_dynamic_array(t) {
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2828, "Fixed capacity dynamic array capacities are constant")
	} else if is_type_slice(t) {
		return lb_slice_len(p, v)
	} else if is_type_dynamic_array(t) {
		return lb_dynamic_array_cap(p, v)
	} else if is_type_map(t) {
		return lb_map_cap(p, v)
	} else if is_type_soa_struct(t) {
		return lb_soa_struct_cap(p, v)
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2839, "Unreachable")
	return lbValue{}
}

func lb_build_builtin_proc_swizzle(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	index_count := len(ce.Args) - 1
	if is_type_simd_vector(tv.Type) {
		vec := lb_build_expr(p, ce.Args[0])
		if index_count == 0 {
			return vec
		}

		mask_len := uint(index_count)
		mask_elems := make([]LLVMValueRef, index_count)
		for i := 1; i < len(ce.Args); i++ {
			tav := type_and_value_of_expr(ce.Args[i])
			gb_assert_handler("Assertion Failure", "is_type_integer(tav.Type)", "src/llvm_backend_proc.cpp", 2856, 0)
			gb_assert_handler("Assertion Failure", "tav.Value.Kind == ExactValue_Integer", "src/llvm_backend_proc.cpp", 2857, 0)

			index := u32(big_int_to_i64(&tav.Value.ValueInteger))
			mask_elems[i-1] = LLVMConstInt(lb_type(p.Module, t_u32), u64(index), false)
		}

		mask := LLVMConstVector(mask_elems, mask_len)

		v1 := vec.Value
		v2 := vec.Value

		res := lbValue{}
		res.Type = tv.Type
		res.Value = LLVMBuildShuffleVector(p.Builder, v1, v2, mask, "")
		return res
	}

	addr := lb_build_array_swizzle_addr(p, ce, tv)
	return lb_addr_load(p, addr)
}

func lb_build_builtin_proc_complex(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
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
}

func lb_build_builtin_proc_quaternion(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	var xyzw [4]lbValue
	for i := i32(0); i < 4; i++ {
		f := ce.Args[i].FieldValue
		gb_assert_handler("Assertion Failure", "f.Field.Kind == Ast_Ident", "src/llvm_backend_proc.cpp", 2897, 0)
		name := f.Field.Ident.Token.String
		var index i32 = -1

		if name == "x" || name == "imag" {
			index = 0
		} else if name == "y" || name == "jmag" {
			index = 1
		} else if name == "z" || name == "kmag" {
			index = 2
		} else if name == "w" || name == "real" {
			index = 3
		}
		gb_assert_handler("Assertion Failure", "index >= 0", "src/llvm_backend_proc.cpp", 2911, 0)

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
}

func lb_build_builtin_proc_real(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	if is_type_complex(val.Type) {
		real := lb_emit_struct_ev(p, val, 0)
		return lb_emit_conv(p, real, tv.Type)
	} else if is_type_quaternion(val.Type) {
		real := lb_emit_struct_ev(p, val, 3)
		return lb_emit_conv(p, real, tv.Type)
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2943, "invalid type for real")
	return lbValue{}
}

func lb_build_builtin_proc_imag(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	if is_type_complex(val.Type) {
		imag := lb_emit_struct_ev(p, val, 1)
		return lb_emit_conv(p, imag, tv.Type)
	} else if is_type_quaternion(val.Type) {
		imag := lb_emit_struct_ev(p, val, 0)
		return lb_emit_conv(p, imag, tv.Type)
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2956, "invalid type for imag")
	return lbValue{}
}

func lb_build_builtin_proc_jmag(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	if is_type_quaternion(val.Type) {
		imag := lb_emit_struct_ev(p, val, 1)
		return lb_emit_conv(p, imag, tv.Type)
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2966, "invalid type for jmag")
	return lbValue{}
}

func lb_build_builtin_proc_kmag(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	if is_type_quaternion(val.Type) {
		imag := lb_emit_struct_ev(p, val, 2)
		return lb_emit_conv(p, imag, tv.Type)
	}
	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2976, "invalid type for kmag")
	return lbValue{}
}

func lb_build_builtin_proc_conj(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	return lb_emit_conjugate(p, val, tv.Type)
}

func lb_build_builtin_proc_expand_values(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	val := lb_build_expr(p, ce.Args[0])
	t := base_type(val.Type)

	if !is_type_tuple(tv.Type) {
		if t.Kind == Type_Struct {
			gb_assert_handler("Assertion Failure", "len(t.Struct.Fields) == 1", "src/llvm_backend_proc.cpp", 2991, 0)
			return lb_emit_struct_ev(p, val, 0)
		} else if t.Kind == Type_Array {
			gb_assert_handler("Assertion Failure", "t.Array.Count == 1", "src/llvm_backend_proc.cpp", 2994, 0)
			return lb_emit_struct_ev(p, val, 0)
		} else {
			gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 2997, "Unknown type of expand_values")
		}
	}

	gb_assert_handler("Assertion Failure", "is_type_tuple(tv.Type)", "src/llvm_backend_proc.cpp", 3002, 0)
	tuple := lb_addr_get_ptr(p, lb_add_local_generated(p, tv.Type, false))
	if t.Kind == Type_Struct {
		for src_index, field := range t.Struct.Fields {
			field_index := field.Variable.FieldIndex
			f := lb_emit_struct_ev(p, val, field_index)
			ep := lb_emit_struct_ep(p, tuple, i32(src_index))
			lb_emit_store(p, ep, f)
		}
	} else if is_type_array_like(t) {
		ap := lb_address_from_load_or_generate_local(p, val)
		n := i32(get_array_type_count(t))
		for i := i32(0); i < n; i++ {
			f := lb_emit_load(p, lb_emit_array_epi(p, ap, isize(i)))
			ep := lb_emit_struct_ep(p, tuple, i)
			lb_emit_store(p, ep, f)
		}
	} else {
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 3023, "Unknown type of expand_values")
	}
	return lb_emit_load(p, tuple)
}

func lb_build_builtin_proc_compress_values(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	value_count := isize(0)
	for _, arg := range ce.Args {
		t := arg.TAV.Type
		if is_type_tuple(t) {
			value_count += isize(len(t.Tuple.Variables))
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
		index := i32(0)
		for _, arg := range ce.Args {
			x := lb_build_expr(p, arg)
			if is_type_tuple(x.Type) {
				for i := isize(0); i < isize(len(x.Type.Tuple.Variables)); i++ {
					y := lb_emit_tuple_ev(p, x, i32(i))
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
		gb_assert_handler("Assertion Failure", "i32(value_count) == index", "src/llvm_backend_proc.cpp", 3064, 0)
	} else if is_type_array_like(dt) {
		index := i32(0)
		for _, arg := range ce.Args {
			x := lb_build_expr(p, arg)
			if is_type_tuple(x.Type) {
				for i := isize(0); i < isize(len(x.Type.Tuple.Variables)); i++ {
					y := lb_emit_tuple_ev(p, x, i32(i))
					ptr := lb_emit_array_epi(p, addr.Addr, isize(index))
					index++
					y = lb_emit_conv(p, y, type_deref(ptr.Type))
					lb_emit_store(p, ptr, y)
				}
			} else {
				ptr := lb_emit_array_epi(p, addr.Addr, isize(index))
				index++
				x = lb_emit_conv(p, x, type_deref(ptr.Type))
				lb_emit_store(p, ptr, x)
			}
		}
		gb_assert_handler("Assertion Failure", "i32(value_count) == index", "src/llvm_backend_proc.cpp", 3082, 0)
	} else {
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 3084, "TODO(bill): compress_values -> %s", type_to_string(tv.Type))
	}

	return lb_addr_load(p, addr)
}

func lb_build_builtin_proc_min(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	t := tv.Type
	if len(ce.Args) == 2 {
		return lb_emit_min(p, t, lb_build_expr(p, ce.Args[0]), lb_build_expr(p, ce.Args[1]))
	} else {
		x := lb_build_expr(p, ce.Args[0])
		for i := 1; i < len(ce.Args); i++ {
			x = lb_emit_min(p, t, x, lb_build_expr(p, ce.Args[i]))
		}
		return x
	}
}

func lb_build_builtin_proc_max(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	t := tv.Type
	if len(ce.Args) == 2 {
		return lb_emit_max(p, t, lb_build_expr(p, ce.Args[0]), lb_build_expr(p, ce.Args[1]))
	} else {
		x := lb_build_expr(p, ce.Args[0])
		for i := 1; i < len(ce.Args); i++ {
			x = lb_emit_max(p, t, x, lb_build_expr(p, ce.Args[i]))
		}
		return x
	}
}

func lb_build_builtin_proc_abs(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	x := lb_build_expr(p, ce.Args[0])
	t := x.Type
	if is_type_unsigned(t) {
		return x
	}
	if is_type_quaternion(t) {
		sz := i64(8) * type_size_of(t)
		args := []lbValue{x}
		switch sz {
		case 64:
			return lb_emit_runtime_call(p, "abs_quaternion64", args)
		case 128:
			return lb_emit_runtime_call(p, "abs_quaternion128", args)
		case 256:
			return lb_emit_runtime_call(p, "abs_quaternion256", args)
		}
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 3131, "Unknown complex type")
	} else if is_type_complex(t) {
		sz := i64(8) * type_size_of(t)
		args := []lbValue{x}
		switch sz {
		case 32:
			return lb_emit_runtime_call(p, "abs_complex32", args)
		case 64:
			return lb_emit_runtime_call(p, "abs_complex64", args)
		case 128:
			return lb_emit_runtime_call(p, "abs_complex128", args)
		}
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 3141, "Unknown complex type")
	} else if is_type_float(t) {
		little := is_type_endian_little(t) || (is_type_endian_platform(t) && buildContext.EndianKind == TargetEndian_Little)
		var t_unsigned *Type
		var mask lbValue
		switch type_size_of(t) {
		case 2:
			t_unsigned = t_u16
			if little {
				mask = lb_const_int(p.Module, t_unsigned, 0x7FFF)
			} else {
				mask = lb_const_int(p.Module, t_unsigned, 0xFF7F)
			}
		case 4:
			t_unsigned = t_u32
			if little {
				mask = lb_const_int(p.Module, t_unsigned, 0x7FFFFFFF)
			} else {
				mask = lb_const_int(p.Module, t_unsigned, 0xFFFFFF7F)
			}
		case 8:
			t_unsigned = t_u64
			if little {
				mask = lb_const_int(p.Module, t_unsigned, 0x7FFFFFFFFFFFFFFF)
			} else {
				mask = lb_const_int(p.Module, t_unsigned, 0xFFFFFFFFFFFFFF7F)
			}
		default:
			gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 3160, "abs: unhandled float size")
		}

		as_unsigned := lb_emit_transmute(p, x, t_unsigned)
		abs := lb_emit_arith(p, Token_And, as_unsigned, mask, t_unsigned)
		return lb_emit_transmute(p, abs, t)
	}

	zero := lb_const_nil(p.Module, t)
	cond := lb_emit_comp(p, Token_Lt, x, zero)
	neg := lb_emit_unary_arith(p, Token_Sub, x, t)
	return lb_emit_select(p, cond, neg, x)
}

func lb_build_builtin_proc_clamp(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_clamp(p, tv.Type,
		lb_build_expr(p, ce.Args[0]),
		lb_build_expr(p, ce.Args[1]),
		lb_build_expr(p, ce.Args[2]))
}

func lb_build_builtin_proc_soa_zip(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_soa_zip(p, ce, tv)
}

func lb_build_builtin_proc_soa_unzip(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_soa_unzip(p, ce, tv)
}

func lb_build_builtin_proc_transpose(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	m := lb_build_expr(p, ce.Args[0])
	return lb_emit_matrix_transpose(p, m, tv.Type)
}

func lb_build_builtin_proc_outer_product(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	a := lb_build_expr(p, ce.Args[0])
	b := lb_build_expr(p, ce.Args[1])
	return lb_emit_outer_product(p, a, b, tv.Type)
}

func lb_build_builtin_proc_hadamard_product(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	a := lb_build_expr(p, ce.Args[0])
	b := lb_build_expr(p, ce.Args[1])
	if is_type_array(tv.Type) {
		return lb_emit_arith(p, Token_Mul, a, b, tv.Type)
	}
	gb_assert_handler("Assertion Failure", "is_type_matrix(tv.Type)", "src/llvm_backend_proc.cpp", 3205, 0)
	return lb_emit_arith_matrix(p, Token_Mul, a, b, tv.Type, true)
}

func lb_build_builtin_proc_matrix_flatten(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	m := lb_build_expr(p, ce.Args[0])
	return lb_emit_matrix_flatten(p, m, tv.Type)
}

func lb_build_builtin_proc_unreachable(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	lb_emit_unreachable(p)
	return lbValue{}
}

func lb_build_builtin_proc_raw_data(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	x := lb_build_expr(p, ce.Args[0])
	t := base_type(x.Type)
	res := lbValue{}
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
	gb_assert_handler("Assertion Failure", "res.Value != 0", "src/llvm_backend_proc.cpp", 3251, 0)
	return res
}

func lb_build_builtin_proc_alloca(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	sz := lb_build_expr(p, ce.Args[0])
	al := exact_value_to_i64(type_and_value_of_expr(ce.Args[1]).Value)

	res := lbValue{}
	res.Type = alloc_type_multi_pointer(t_u8)
	res.Value = LLVMBuildArrayAlloca(p.Builder, lb_type(p.Module, t_u8), sz.Value, "")
	LLVMSetAlignment(res.Value, uint(al))
	return res
}

func lb_build_builtin_proc_cpu_relax(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	if buildContext.Metrics.Arch == TargetArch_i386 ||
		buildContext.Metrics.Arch == TargetArch_amd64 {
		func_type := LLVMFunctionType(LLVMVoidTypeInContext(p.Module.Ctx), nil, 0, false)
		the_asm := llvm_get_inline_asm(func_type, "pause", "", true, false, LLVMInlineAsmDialectAT_T)
		gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 3275, 0)
		LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
	} else if buildContext.Metrics.Arch == TargetArch_arm64 {
		func_type := LLVMFunctionType(LLVMVoidTypeInContext(p.Module.Ctx), nil, 0, false)
		the_asm := llvm_get_inline_asm(func_type, "isb", "", true, false, LLVMInlineAsmDialectAT_T)
		gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 3282, 0)
		LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
	} else {
		func_type := LLVMFunctionType(LLVMVoidTypeInContext(p.Module.Ctx), nil, 0, false)
		the_asm := llvm_get_inline_asm(func_type, "", "", true, false, LLVMInlineAsmDialectAT_T)
		gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 3288, 0)
		LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
	}
	return lbValue{}
}

func lb_build_builtin_proc_debug_trap(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
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
}

func lb_build_builtin_proc_read_cycle_counter(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	res := lbValue{}
	res.Type = tv.Type

	if buildContext.Metrics.Arch == TargetArch_arm64 {
		func_type := LLVMFunctionType(LLVMInt64TypeInContext(p.Module.Ctx), nil, 0, false)
		has_side_effects := false
		the_asm := llvm_get_inline_asm(func_type, "mrs $0, cntvct_el0", "=r", has_side_effects, false, LLVMInlineAsmDialectAT_T)
		gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 3319, 0)
		res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
	} else {
		name := "llvm.readcyclecounter"
		res.Value = lb_call_intrinsic(p, name, nil, 0, nil, 0)
	}
	return res
}

func lb_build_builtin_proc_read_cycle_counter_frequency(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	res := lbValue{}
	res.Type = tv.Type

	if buildContext.Metrics.Arch == TargetArch_arm64 {
		func_type := LLVMFunctionType(LLVMInt64TypeInContext(p.Module.Ctx), nil, 0, false)
		has_side_effects := false
		the_asm := llvm_get_inline_asm(func_type, "mrs $0, cntfrq_el0", "=r", has_side_effects, false, LLVMInlineAsmDialectAT_T)
		gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 3336, 0)
		res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, nil, 0, "")
	}

	return res
}

func lb_build_builtin_proc_count_trailing_zeros(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_trailing_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_count_leading_zeros(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_leading_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_count_trailing_ones(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_trailing_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_count_leading_ones(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_leading_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_count_ones(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_ones(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_count_zeros(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_count_zeros(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_reverse_bits(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_emit_reverse_bits(p, lb_build_expr(p, ce.Args[0]), tv.Type)
}

func lb_build_builtin_proc_byte_swap(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	x := lb_build_expr(p, ce.Args[0])
	x = lb_emit_conv(p, x, tv.Type)
	return lb_emit_byte_swap(p, x, tv.Type)
}

func lb_build_builtin_proc_overflow(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
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
	types := [1]LLVMTypeRef{lb_type(p.Module, type_)}
	args := [2]LLVMValueRef{x.Value, y.Value}

	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))

	if is_type_tuple(main_type) {
		var res_type *Type
		res_type = alloc_type_tuple()
		res_type.Tuple.Variables = make([]*Entity, 2)
		res_type.Tuple.Variables[0] = alloc_entity_field(nil, blank_token, type_, false, 0)
		res_type.Tuple.Variables[1] = alloc_entity_field(nil, blank_token, t_llvm_bool, false, 1)
		res.Type = res_type
	} else {
		res.Value = LLVMBuildExtractValue(p.Builder, res.Value, 0, "")
		res.Type = type_
	}
	return res
}

func lb_build_builtin_proc_saturating_add_sub(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	main_type := tv.Type
	type_ := main_type

	x := lb_build_expr(p, ce.Args[0])
	y := lb_build_expr(p, ce.Args[1])
	x = lb_emit_conv(p, x, type_)
	y = lb_emit_conv(p, y, type_)

	var name string
	if is_type_unsigned(type_) {
		switch id {
		case BuiltinProc_saturating_add:
			name = "llvm.uadd.sat"
		case BuiltinProc_saturating_sub:
			name = "llvm.usub.sat"
		}
	} else {
		switch id {
		case BuiltinProc_saturating_add:
			name = "llvm.sadd.sat"
		case BuiltinProc_saturating_sub:
			name = "llvm.ssub.sat"
		}
	}
	types := [1]LLVMTypeRef{lb_type(p.Module, type_)}
	args := [2]LLVMValueRef{x.Value, y.Value}

	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = type_
	return res
}

func lb_build_builtin_proc_sqrt(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	type_ := tv.Type

	x := lb_build_expr(p, ce.Args[0])
	x = lb_emit_conv(p, x, type_)

	name := "llvm.sqrt"
	types := [1]LLVMTypeRef{lb_type(p.Module, type_)}
	args := [1]LLVMValueRef{x.Value}

	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = type_
	return res
}

func lb_build_builtin_proc_fused_mul_add(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	type_ := tv.Type
	x := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), type_)
	y := lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), type_)
	z := lb_emit_conv(p, lb_build_expr(p, ce.Args[2]), type_)

	name := "llvm.fma"
	types := [1]LLVMTypeRef{lb_type(p.Module, type_)}
	args := [3]LLVMValueRef{x.Value, y.Value, z.Value}

	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = type_
	return res
}

func lb_build_builtin_proc_mem_copy(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	dst := lb_build_expr(p, ce.Args[0])
	src := lb_build_expr(p, ce.Args[1])
	len_ := lb_build_expr(p, ce.Args[2])

	lb_mem_copy_overlapping(p, dst, src, len_, false)
	return lbValue{}
}

func lb_build_builtin_proc_mem_copy_non_overlapping(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	dst := lb_build_expr(p, ce.Args[0])
	src := lb_build_expr(p, ce.Args[1])
	len_ := lb_build_expr(p, ce.Args[2])

	lb_mem_copy_non_overlapping(p, dst, src, len_, false)
	return lbValue{}
}

func lb_build_builtin_proc_mem_zero(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])
	len_ := lb_build_expr(p, ce.Args[1])
	ptr = lb_emit_conv(p, ptr, t_rawptr)
	len_ = lb_emit_conv(p, len_, t_int)

	alignment := uint(1)
	lb_mem_zero_ptr_internal(p, ptr.Value, len_.Value, alignment, false)
	return lbValue{}
}

func lb_build_builtin_proc_mem_zero_volatile(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])
	len_ := lb_build_expr(p, ce.Args[1])
	ptr = lb_emit_conv(p, ptr, t_rawptr)
	len_ = lb_emit_conv(p, len_, t_int)

	alignment := uint(1)
	lb_mem_zero_ptr_internal(p, ptr.Value, len_.Value, alignment, true)
	return lbValue{}
}

func lb_build_builtin_proc_ptr_offset(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])
	len_ := lb_build_expr(p, ce.Args[1])
	len_ = lb_emit_conv(p, len_, t_int)
	return lb_emit_ptr_offset(p, ptr, len_)
}

func lb_build_builtin_proc_ptr_sub(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	elem0 := type_deref(type_of_expr(ce.Args[0]), true)
	elem1 := type_deref(type_of_expr(ce.Args[1]), true)
	gb_assert_handler("Assertion Failure", "are_types_identical(elem0, elem1)", "src/llvm_backend_proc.cpp", 3544, 0)
	elem := elem0

	ptr0 := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_uintptr)
	ptr1 := lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_uintptr)
	ptr0 = lb_emit_conv(p, ptr0, t_int)
	ptr1 = lb_emit_conv(p, ptr1, t_int)

	diff := lb_emit_arith(p, Token_Sub, ptr0, ptr1, t_int)
	return lb_emit_arith(p, Token_Quo, diff, lb_const_int(p.Module, t_int, u64(type_size_of(elem))), t_int)
}

func lb_build_builtin_proc_atomic_thread_fence(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	LLVMBuildFence(p.Builder, llvm_atomic_ordering_from_odin(ce.Args[0].TAV.Value), false, "")
	return lbValue{}
}

func lb_build_builtin_proc_atomic_signal_fence(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	LLVMBuildFence(p.Builder, llvm_atomic_ordering_from_odin(ce.Args[0]), true, "")
	return lbValue{}
}

func lb_build_builtin_proc_volatile_non_temporal_atomic_store(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	dst := lb_build_expr(p, ce.Args[0])
	val := lb_build_expr(p, ce.Args[1])
	val = lb_emit_conv(p, val, type_deref(dst.Type))

	instr := LLVMBuildStore(p.Builder, val.Value, dst.Value)
	switch id {
	case BuiltinProc_non_temporal_store:
		kind_id := LLVMGetMDKindIDInContext(p.Module.Ctx, "nontemporal", uint(len("nontemporal")))
		node := LLVMValueAsMetadata(LLVMConstInt(lb_type(p.Module, t_u32), 1, false))
		LLVMSetMetadata(instr, kind_id, LLVMMetadataAsValue(p.Module.Ctx, node))
	case BuiltinProc_volatile_store:
		LLVMSetVolatile(instr, true)
	case BuiltinProc_atomic_store:
		LLVMSetOrdering(instr, LLVMAtomicOrderingSequentiallyConsistent)
		LLVMSetVolatile(instr, true)
	case BuiltinProc_atomic_store_explicit:
		ordering := llvm_atomic_ordering_from_odin(ce.Args[2].TAV.Value)
		LLVMSetOrdering(instr, ordering)
		LLVMSetVolatile(instr, true)
	}

	LLVMSetAlignment(instr, uint(type_align_of(type_deref(dst.Type))))

	return lbValue{}
}

func lb_build_builtin_proc_volatile_non_temporal_atomic_load(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	dst := lb_build_expr(p, ce.Args[0])

	instr := OdinLLVMBuildLoad(p, lb_type(p.Module, type_deref(dst.Type)), dst.Value)
	switch id {
	case BuiltinProc_non_temporal_load:
		kind_id := LLVMGetMDKindIDInContext(p.Module.Ctx, "nontemporal", uint(len("nontemporal")))
		node := LLVMValueAsMetadata(LLVMConstInt(lb_type(p.Module, t_u32), 1, false))
		LLVMSetMetadata(instr, kind_id, LLVMMetadataAsValue(p.Module.Ctx, node))
	case BuiltinProc_volatile_load:
		LLVMSetVolatile(instr, true)
	case BuiltinProc_atomic_load:
		LLVMSetOrdering(instr, LLVMAtomicOrderingSequentiallyConsistent)
		LLVMSetVolatile(instr, true)
	case BuiltinProc_atomic_load_explicit:
		ordering := llvm_atomic_ordering_from_odin(ce.Args[1].TAV.Value)
		LLVMSetOrdering(instr, ordering)
		LLVMSetVolatile(instr, true)
	}
	LLVMSetAlignment(instr, uint(type_align_of(type_deref(dst.Type))))

	res := lbValue{}
	res.Value = instr
	res.Type = type_deref(dst.Type)
	return res
}

func lb_build_builtin_proc_unaligned_store(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	dst := lb_build_expr(p, ce.Args[0])
	src := lb_build_expr(p, ce.Args[1])
	t := type_deref(dst.Type)

	if is_type_simd_vector(t) {
		store := LLVMBuildStore(p.Builder, src.Value, dst.Value)
		LLVMSetAlignment(store, 1)
	} else {
		src = lb_address_from_load_or_generate_local(p, src)
		lb_mem_copy_non_overlapping(p, dst, src, lb_const_int(p.Module, t_int, u64(type_size_of(t))), false)
	}
	return lbValue{}
}

func lb_build_builtin_proc_unaligned_load(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	src := lb_build_expr(p, ce.Args[0])
	t := type_deref(src.Type)
	if is_type_simd_vector(t) {
		res := lbValue{}
		res.Type = t
		res.Value = OdinLLVMBuildLoadAligned(p, lb_type(p.Module, t), src.Value, 1)
		return res
	} else {
		dst := lb_add_local_generated(p, t, false)
		lb_mem_copy_non_overlapping(p, dst.Addr, src, lb_const_int(p.Module, t_int, u64(type_size_of(t))), false)
		return lb_addr_load(p, dst)
	}
}

func lb_build_builtin_proc_atomic_rmw(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
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

	res := lbValue{}
	res.Value = LLVMBuildAtomicRMW(p.Builder, op, dst.Value, val.Value, ordering, false)
	res.Type = tv.Type
	LLVMSetVolatile(res.Value, true)
	return res
}

func lb_build_builtin_proc_atomic_compare_exchange(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	address := lb_build_expr(p, ce.Args[0])
	elem := type_deref(address.Type)
	old_value := lb_build_expr(p, ce.Args[1])
	new_value := lb_build_expr(p, ce.Args[2])
	old_value = lb_emit_conv(p, old_value, elem)
	new_value = lb_emit_conv(p, new_value, elem)

	var success_ordering LLVMAtomicOrdering
	var failure_ordering LLVMAtomicOrdering
	weak := false

	switch id {
	case BuiltinProc_atomic_compare_exchange_strong:
		success_ordering = LLVMAtomicOrderingSequentiallyConsistent
		failure_ordering = LLVMAtomicOrderingSequentiallyConsistent
		weak = false
	case BuiltinProc_atomic_compare_exchange_weak:
		success_ordering = LLVMAtomicOrderingSequentiallyConsistent
		failure_ordering = LLVMAtomicOrderingSequentiallyConsistent
		weak = true
	case BuiltinProc_atomic_compare_exchange_strong_explicit:
		success_ordering = llvm_atomic_ordering_from_odin(ce.Args[3].TAV.Value)
		failure_ordering = llvm_atomic_ordering_from_odin(ce.Args[4].TAV.Value)
		weak = false
	case BuiltinProc_atomic_compare_exchange_weak_explicit:
		success_ordering = llvm_atomic_ordering_from_odin(ce.Args[3].TAV.Value)
		failure_ordering = llvm_atomic_ordering_from_odin(ce.Args[4].TAV.Value)
		weak = true
	}

	single_threaded := false

	value := LLVMBuildAtomicCmpXchg(
		p.Builder, address.Value,
		old_value.Value, new_value.Value,
		success_ordering,
		failure_ordering,
		single_threaded,
	)
	LLVMSetWeak(value, weak)
	LLVMSetVolatile(value, true)

	if is_type_tuple(tv.Type) {
		fix_typed := alloc_type_tuple()
		fix_typed.Tuple.Variables = make([]*Entity, 2)
		fix_typed.Tuple.Variables[0] = tv.Type.Tuple.Variables[0]
		fix_typed.Tuple.Variables[1] = alloc_entity_field(nil, blank_token, t_llvm_bool, false, 1)

		res := lbValue{}
		res.Value = value
		res.Type = fix_typed
		return res
	} else {
		res := lbValue{}
		res.Value = LLVMBuildExtractValue(p.Builder, value, 0, "")
		res.Type = tv.Type
		return res
	}
}

func lb_build_builtin_proc_type_equal_proc(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_equal_proc_for_type(p.Module, ce.Args[0].TAV.Type)
}

func lb_build_builtin_proc_type_hasher_proc(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_hasher_proc_for_type(p.Module, ce.Args[0].TAV.Type)
}

func lb_build_builtin_proc_type_map_info(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_gen_map_info_ptr(p.Module, ce.Args[0].TAV.Type)
}

func lb_build_builtin_proc_type_map_cell_info(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_gen_map_cell_info_ptr(p.Module, ce.Args[0].TAV.Type)
}

func lb_build_builtin_proc_fixed_point(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	platform_type := integer_endian_type_to_platform_type(tv.Type)

	x := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), platform_type)
	y := lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), platform_type)
	scale := lb_emit_conv(p, lb_build_expr(p, ce.Args[2]), t_i32)

	var name string
	if is_type_unsigned(tv.Type) {
		switch id {
		case BuiltinProc_fixed_point_mul:
			name = "llvm.umul.fix"
		case BuiltinProc_fixed_point_div:
			name = "llvm.udiv.fix"
		case BuiltinProc_fixed_point_mul_sat:
			name = "llvm.umul.fix.sat"
		case BuiltinProc_fixed_point_div_sat:
			name = "llvm.udiv.fix.sat"
		}
	} else {
		switch id {
		case BuiltinProc_fixed_point_mul:
			name = "llvm.smul.fix"
		case BuiltinProc_fixed_point_div:
			name = "llvm.sdiv.fix"
		case BuiltinProc_fixed_point_mul_sat:
			name = "llvm.smul.fix.sat"
		case BuiltinProc_fixed_point_div_sat:
			name = "llvm.sdiv.fix.sat"
		}
	}
	gb_assert_handler("Assertion Failure", "name != \"\"", "src/llvm_backend_proc.cpp", 3812, 0)

	res := lbValue{}
	res.Type = platform_type

	if id == BuiltinProc_fixed_point_div ||
		id == BuiltinProc_fixed_point_div_sat {
		res.Value = lb_integer_division_fixed_point_intrinsics(p, x.Value, y.Value, scale.Value, platform_type, name)
	} else {
		types := [1]LLVMTypeRef{lb_type(p.Module, platform_type)}
		args := [3]LLVMValueRef{x.Value, y.Value, scale.Value}
		res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	}
	return lb_emit_conv(p, res, tv.Type)
}

func lb_build_builtin_proc_expect(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	t := default_type(tv.Type)
	x := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t)
	y := lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t)

	name := "llvm.expect"
	types := [1]LLVMTypeRef{lb_type(p.Module, t)}
	res := lbValue{}
	args := [2]LLVMValueRef{x.Value, y.Value}

	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = t
	return lb_emit_conv(p, res, t)
}

func lb_build_builtin_proc_likely_unlikely(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	t := default_type(tv.Type)
	x := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t)
	y := lb_const_bool(p.Module, t, id == BuiltinProc_likely)

	name := "llvm.expect"
	types := [1]LLVMTypeRef{lb_type(p.Module, t)}
	res := lbValue{}
	args := [2]LLVMValueRef{x.Value, y.Value}

	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = t
	return lb_emit_conv(p, res, t)
}

func lb_build_builtin_proc_prefetch(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue, id BuiltinProcId) lbValue {
	ptr := lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_rawptr)
	locality := u64(exact_value_to_i64(ce.Args[1].TAV.Value))
	var rw u64
	var cache u64
	switch id {
	case BuiltinProc_prefetch_read_instruction:
		rw = 0
		cache = 0
	case BuiltinProc_prefetch_read_data:
		rw = 0
		cache = 1
	case BuiltinProc_prefetch_write_instruction:
		rw = 1
		cache = 0
	case BuiltinProc_prefetch_write_data:
		rw = 1
		cache = 1
	}

	name := "llvm.prefetch"

	types := [1]LLVMTypeRef{lb_type(p.Module, t_rawptr)}

	llvm_i32 := lb_type(p.Module, t_i32)
	args := [4]LLVMValueRef{}
	args[0] = ptr.Value
	args[1] = LLVMConstInt(llvm_i32, rw, false)
	args[2] = LLVMConstInt(llvm_i32, locality, false)
	args[3] = LLVMConstInt(llvm_i32, cache, false)

	res := lbValue{}
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	res.Type = nil
	return res
}

func lb_build_builtin_proc___entry_point(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	if p.Module.Info.EntryPoint != nil {
		entry_point := lb_find_procedure_value_from_entity(p.Module, p.Module.Info.EntryPoint)
		gb_assert_handler("Assertion Failure", "entry_point.Value != 0", "src/llvm_backend_proc.cpp", 3916, 0)
		lb_emit_call(p, entry_point, nil)
	}
	return lbValue{}
}

func lb_build_builtin_proc_syscall(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	arg_count := uint(len(ce.Args))
	args := make([]LLVMValueRef, arg_count)
	for i, arg_ast := range ce.Args {
		arg := lb_build_expr(p, arg_ast)
		arg = lb_emit_conv(p, arg, t_uintptr)
		args[i] = arg.Value
	}

	llvm_uintptr := lb_type(p.Module, t_uintptr)
	llvm_arg_types := make([]LLVMTypeRef, arg_count)
	for i := uint(0); i < arg_count; i++ {
		llvm_arg_types[i] = llvm_uintptr
	}

	func_type := LLVMFunctionType(llvm_uintptr, llvm_arg_types, arg_count, false)

	var inline_asm LLVMValueRef

	switch buildContext.Metrics.Arch {
	case TargetArch_riscv64:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 3944, 0)

			asm_string := "ecall"
			constraints := "={a0}"
			regs := []string{"a7", "a0", "a1", "a2", "a3", "a4", "a5", "a6"}
			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}
			constraints += ",~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	case TargetArch_amd64:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 3971, 0)

			asm_string := "syscall"
			constraints := "={rax}"
			regs := []string{"rax", "rdi", "rsi", "rdx", "r10", "r8", "r9"}
			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}

			constraints += ",~{rcx},~{r11},~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	case TargetArch_i386:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 4008, 0)

			asm_string := "int $$0x80"
			constraints := "={eax}"
			regs := []string{"eax", "ebx", "ecx", "edx", "esi", "edi", "ebp"}
			min_count := arg_count
			if min_count > 6 {
				min_count = 6
			}
			for i := uint(0); i < min_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}

			constraints += ",~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	case TargetArch_arm64:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 4035, 0)

			var asm_string string
			var constraints string
			var regs []string

			if buildContext.Metrics.OS == TargetOs_darwin {
				asm_string = "svc #0x80"
				regs = []string{"x16", "x0", "x1", "x2", "x3", "x4", "x5"}
			} else {
				asm_string = "svc #0"
				regs = []string{"x8", "x0", "x1", "x2", "x3", "x4", "x5"}
			}
			constraints = "={x0}"
			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}
			constraints += ",~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	case TargetArch_arm32:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 4084, 0)

			asm_string := "svc #0"
			constraints := "={r0}"
			regs := []string{"r7", "r0", "r1", "r2", "r3", "r4", "r5", "r6"}
			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}
			constraints += ",~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 4110, "Unsupported platform")
	}

	res := lbValue{}
	res.Value = LLVMBuildCall2(p.Builder, func_type, inline_asm, args, arg_count, "")
	res.Type = t_uintptr
	return res
}

func lb_build_builtin_proc_syscall_bsd(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	arg_count := uint(len(ce.Args))
	args := make([]LLVMValueRef, arg_count)
	for i, arg_ast := range ce.Args {
		arg := lb_build_expr(p, arg_ast)
		arg = lb_emit_conv(p, arg, t_uintptr)
		args[i] = arg.Value
	}

	llvm_uintptr := lb_type(p.Module, t_uintptr)
	llvm_arg_types := make([]LLVMTypeRef, arg_count)
	for i := uint(0); i < arg_count; i++ {
		llvm_arg_types[i] = llvm_uintptr
	}

	results := make([]LLVMTypeRef, 2)
	results[0] = lb_type(p.Module, t_uintptr)
	results[1] = lb_type(p.Module, t_bool)
	llvm_results := LLVMStructTypeInContext(p.Module.Ctx, results, 2, false)

	func_type := LLVMFunctionType(llvm_results, llvm_arg_types, arg_count, false)

	var inline_asm LLVMValueRef

	switch buildContext.Metrics.Arch {
	case TargetArch_amd64:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 4152, 0)

			asm_string := "syscall; setnb %cl"
			constraints := "={rax},={cl}"
			regs := []string{"rax", "rdi", "rsi", "rdx", "r10", "r8", "r9"}
			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}

			if buildContext.Metrics.OS == TargetOs_freebsd {
				constraints += ",~{r8},~{r9},~{r10}"
			}

			constraints += ",~{rdx},~{r11},~{cc},~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	case TargetArch_arm64:
		{
			gb_assert_handler("Assertion Failure", "arg_count <= 7", "src/llvm_backend_proc.cpp", 4216, 0)

			var asm_string string
			var constraints string
			var regs []string

			if buildContext.Metrics.OS == TargetOs_netbsd {
				asm_string = "svc #0; cset x17, cc"
				constraints = "={x0},={x17}"
				regs = []string{"x17", "x0", "x1", "x2", "x3", "x4", "x5"}
			} else {
				asm_string = "svc #0; cset x8, cc"
				constraints = "={x0},={x8}"
				regs = []string{"x8", "x0", "x1", "x2", "x3", "x4", "x5"}
				constraints += ",~{x1}"
			}

			for i := uint(0); i < arg_count; i++ {
				constraints += "," + "{" + regs[i] + "}"
			}
			constraints += ",~{cc},~{memory}"

			inline_asm = llvm_get_inline_asm(func_type, asm_string, constraints, true, false, LLVMInlineAsmDialectAT_T)
		}
	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 4265, "Unsupported platform")
	}

	res := lbValue{}
	res.Value = LLVMBuildCall2(p.Builder, func_type, inline_asm, args, arg_count, "")
	res.Type = make_optional_ok_type(t_uintptr, true)

	return res
}

func lb_build_builtin_proc_objc_send(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_send(p, ce)
}

func lb_build_builtin_proc_objc_find_selector(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_find_selector(p, ce)
}

func lb_build_builtin_proc_objc_find_class(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_find_class(p, ce)
}

func lb_build_builtin_proc_objc_register_selector(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_register_selector(p, ce)
}

func lb_build_builtin_proc_objc_register_class(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_register_class(p, ce)
}

func lb_build_builtin_proc_objc_ivar_get(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_ivar_get(p, ce)
}

func lb_build_builtin_proc_objc_block(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_block(p, ce)
}

func lb_build_builtin_proc_objc_super(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	return lb_handle_objc_super(p, ce)
}

func lb_build_builtin_proc_c_va_start(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])
	_ = lb_build_expr(p, ce.Args[1])

	va_start_args := [1]LLVMValueRef{ptr.Value}
	va_start_types := [1]LLVMTypeRef{lb_type(p.Module, ptr.Type)}
	_ = lb_call_intrinsic(p, "llvm.va_start", va_start_args[:], uint(len(va_start_args)), va_start_types[:], uint(len(va_start_types)))

	return lbValue{}
}

func lb_build_builtin_proc_c_va_end(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])

	va_end_args := [1]LLVMValueRef{ptr.Value}
	va_end_types := [1]LLVMTypeRef{lb_type(p.Module, ptr.Type)}
	_ = lb_call_intrinsic(p, "llvm.va_end", va_end_args[:], uint(len(va_end_args)), va_end_types[:], uint(len(va_end_types)))

	return lbValue{}
}

func lb_build_builtin_proc_c_va_copy(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	dst := lb_build_expr(p, ce.Args[0])
	src := lb_build_expr(p, ce.Args[1])

	va_copy_args := [2]LLVMValueRef{dst.Value, src.Value}
	va_copy_types := [1]LLVMTypeRef{lb_type(p.Module, dst.Type)}
	_ = lb_call_intrinsic(p, "llvm.va_copy", va_copy_args[:], uint(len(va_copy_args)), va_copy_types[:], uint(len(va_copy_types)))

	return lbValue{}
}

func lb_build_builtin_proc_c_va_arg(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	ptr := lb_build_expr(p, ce.Args[0])
	type_ := type_of_expr(ce.Args[1])
	value := LLVMBuildVAArg(p.Builder, ptr.Value, lb_type(p.Module, type_), "")

	return lbValue{Value: value, Type: type_}
}

func lb_build_builtin_proc_constant_utf16_cstring(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	m := p.Module

	tav := type_and_value_of_expr(ce.Args[0])
	gb_assert_handler("Assertion Failure", "tav.Value.Kind == ExactValue_String", "src/llvm_backend_proc.cpp", 4350, 0)
	value := tav.Value.ValueString

	llvm_u16 := lb_type(m, t_u16)

	max_len := isize(len(goStr(value)))*2 + 1
	buffer := make([]LLVMValueRef, max_len)
	n := isize(0)
	for _, r := range goStr(value) {
		if (0 <= r && r < 0xd800) || (0xe000 <= r && r < 0x10000) {
			buffer[n] = LLVMConstInt(llvm_u16, u64(u16(r)), false)
			n++
		} else if 0x10000 <= r && r <= 0x10ffff {
			r1 := u16(0xd800 + ((r - 0x10000)>>10)&0x3ff)
			r2 := u16(0xdc00 + (r-0x10000)&0x3ff)
			buffer[n] = LLVMConstInt(llvm_u16, u64(r1), false)
			n++
			buffer[n] = LLVMConstInt(llvm_u16, u64(r2), false)
			n++
		} else {
			buffer[n] = LLVMConstInt(llvm_u16, 0xfffd, false)
			n++
		}
	}

	buffer[n] = LLVMConstInt(llvm_u16, 0, false)
	n++

	array := LLVMConstArray(llvm_u16, buffer, uint(n))

	var name string
	{
		id := m.GlobalArrayIndex.Add(1)
		name = fmt.Sprintf("csbs$%x", id)
	}
	type_ := LLVMTypeOf(array)
	global_data := LLVMAddGlobal(m.Mod, type_, name)
	LLVMSetInitializer(global_data, array)
	LLVMSetUnnamedAddress(global_data, LLVMGlobalUnnamedAddr)
	LLVMSetLinkage(global_data, LLVMInternalLinkage)

	indices := [2]LLVMValueRef{
		LLVMConstInt(lb_type(m, t_u32), 0, false),
		LLVMConstInt(lb_type(m, t_u32), 0, false),
	}
	res := lbValue{}
	res.Type = tv.Type
	res.Value = LLVMBuildInBoundsGEP2(p.Builder, type_, global_data, indices[:], uint(len(indices)), "")
	return res
}

func lb_build_builtin_proc_wasm_memory_grow(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	name := "llvm.wasm.memory.grow"
	types := [1]LLVMTypeRef{
		lb_type(p.Module, t_i32),
	}

	args := [2]LLVMValueRef{}
	args[0] = lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_uintptr).Value
	args[1] = lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_uintptr).Value

	res := lbValue{}
	res.Type = t_i32
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	return lb_emit_conv(p, res, tv.Type)
}

func lb_build_builtin_proc_wasm_memory_size(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	name := "llvm.wasm.memory.size"
	types := [1]LLVMTypeRef{
		lb_type(p.Module, t_i32),
	}

	args := [1]LLVMValueRef{}
	args[0] = lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_uintptr).Value

	res := lbValue{}
	res.Type = t_i32
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), types[:], uint(len(types)))
	return lb_emit_conv(p, res, tv.Type)
}

func lb_build_builtin_proc_wasm_memory_atomic_wait32(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	name := "llvm.wasm.memory.atomic.wait32"

	t_u32_ptr := alloc_type_pointer(t_u32)

	args := [3]LLVMValueRef{}
	args[0] = lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_u32_ptr).Value
	args[1] = lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_u32).Value
	args[2] = lb_emit_conv(p, lb_build_expr(p, ce.Args[2]), t_i64).Value

	res := lbValue{}
	res.Type = tv.Type
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), nil, 0)
	return res
}

func lb_build_builtin_proc_wasm_memory_atomic_notify32(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	name := "llvm.wasm.memory.atomic.notify"

	t_u32_ptr := alloc_type_pointer(t_u32)

	args := [2]LLVMValueRef{
		lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_u32_ptr).Value,
		lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_u32).Value,
	}

	res := lbValue{}
	res.Type = tv.Type
	res.Value = lb_call_intrinsic(p, name, args[:], uint(len(args)), nil, 0)
	return res
}

func lb_build_builtin_proc_x86_cpuid(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	param_types := [2]*Type{t_u32, t_u32}
	type_ := alloc_type_proc_from_types(param_types[:], tv.Type, false, ProcCC_None)
	func_type := lb_get_procedure_raw_type(p.Module, type_)
	the_asm := llvm_get_inline_asm(
		func_type,
		"cpuid",
		"={ax},={bx},={cx},={dx},{ax},{cx}",
		true, false, LLVMInlineAsmDialectAT_T,
	)
	gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 4484, 0)

	asm_args := [2]LLVMValueRef{}
	asm_args[0] = lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_u32).Value
	asm_args[1] = lb_emit_conv(p, lb_build_expr(p, ce.Args[1]), t_u32).Value
	res := lbValue{}
	res.Type = tv.Type
	res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, asm_args[:], uint(len(asm_args)), "")
	return res
}

func lb_build_builtin_proc_x86_xgetbv(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	type_ := alloc_type_proc_from_types([]*Type{t_u32}, tv.Type, false, ProcCC_None)
	func_type := lb_get_procedure_raw_type(p.Module, type_)
	the_asm := llvm_get_inline_asm(
		func_type,
		"xgetbv",
		"={ax},={dx},{cx}",
		true, false, LLVMInlineAsmDialectAT_T,
	)
	gb_assert_handler("Assertion Failure", "the_asm != 0", "src/llvm_backend_proc.cpp", 4504, 0)

	asm_args := [1]LLVMValueRef{}
	asm_args[0] = lb_emit_conv(p, lb_build_expr(p, ce.Args[0]), t_u32).Value
	res := lbValue{}
	res.Type = tv.Type
	res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, asm_args[:], uint(len(asm_args)), "")
	return res
}

func lb_build_builtin_proc_valgrind_client_request(p *lbProcedure, ce *AstCallExpr, tv TypeAndValue) lbValue {
	var args [7]lbValue
	for i := isize(0); i < 7; i++ {
		args[i] = lb_emit_conv(p, lb_build_expr(p, ce.Args[i]), t_uintptr)
	}
	if !buildContext.ODIN_VALGRIND_SUPPORT {
		return args[0]
	}
	array := lb_generate_local_array(p, t_uintptr, 6, false)
	for i := isize(0); i < 6; i++ {
		gep := lb_emit_array_epi(p, array, i)
		lb_emit_store(p, gep, args[i+1])
	}

	switch buildContext.Metrics.Arch {
	case TargetArch_amd64:
		{
			param_types := [2]*Type{}
			param_types[0] = t_uintptr
			param_types[1] = array.Type

			type_ := alloc_type_proc_from_types(param_types[:], t_uintptr, false, ProcCC_None)
			func_type := lb_get_procedure_raw_type(p.Module, type_)
			the_asm := llvm_get_inline_asm(
				func_type,
				"rolq $$3, %rdi; rolq $$13, %rdi\n rolq $$61, %rdi; rolq $$51, %rdi\n xchgq %rbx, %rbx",
				"={rdx},{rdx},{rax},~{cc},~{memory}",
				true, false, LLVMInlineAsmDialectAT_T,
			)

			asm_args := [2]LLVMValueRef{}
			asm_args[0] = args[0].Value
			asm_args[1] = array.Value

			res := lbValue{}
			res.Type = t_uintptr
			res.Value = LLVMBuildCall2(p.Builder, func_type, the_asm, asm_args[:], uint(len(asm_args)), "")
			return res
		}
	default:
		gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 4556, "Unsupported architecture")
	}
	return lbValue{}
}

func lb_build_builtin_proc(p *lbProcedure, expr *Ast, tv TypeAndValue, id BuiltinProcId) lbValue {
	ce := &expr.CallExpr

	switch id {
	case BuiltinProc_typeid_of:
		return lb_build_builtin_proc_typeid_of(p, ce, tv)
	case BuiltinProc_len:
		return lb_build_builtin_proc_len(p, ce, tv)
	case BuiltinProc_cap:
		return lb_build_builtin_proc_cap(p, ce, tv)
	case BuiltinProc_swizzle:
		return lb_build_builtin_proc_swizzle(p, ce, tv)
	case BuiltinProc_complex:
		return lb_build_builtin_proc_complex(p, ce, tv)
	case BuiltinProc_quaternion:
		return lb_build_builtin_proc_quaternion(p, ce, tv)
	case BuiltinProc_real:
		return lb_build_builtin_proc_real(p, ce, tv)
	case BuiltinProc_imag:
		return lb_build_builtin_proc_imag(p, ce, tv)
	case BuiltinProc_jmag:
		return lb_build_builtin_proc_jmag(p, ce, tv)
	case BuiltinProc_kmag:
		return lb_build_builtin_proc_kmag(p, ce, tv)
	case BuiltinProc_conj:
		return lb_build_builtin_proc_conj(p, ce, tv)
	case BuiltinProc_expand_values:
		return lb_build_builtin_proc_expand_values(p, ce, tv)
	case BuiltinProc_compress_values:
		return lb_build_builtin_proc_compress_values(p, ce, tv)
	case BuiltinProc_min:
		return lb_build_builtin_proc_min(p, ce, tv)
	case BuiltinProc_max:
		return lb_build_builtin_proc_max(p, ce, tv)
	case BuiltinProc_abs:
		return lb_build_builtin_proc_abs(p, ce, tv)
	case BuiltinProc_clamp:
		return lb_build_builtin_proc_clamp(p, ce, tv)
	case BuiltinProc_soa_zip:
		return lb_build_builtin_proc_soa_zip(p, ce, tv)
	case BuiltinProc_soa_unzip:
		return lb_build_builtin_proc_soa_unzip(p, ce, tv)
	case BuiltinProc_transpose:
		return lb_build_builtin_proc_transpose(p, ce, tv)
	case BuiltinProc_outer_product:
		return lb_build_builtin_proc_outer_product(p, ce, tv)
	case BuiltinProc_hadamard_product:
		return lb_build_builtin_proc_hadamard_product(p, ce, tv)
	case BuiltinProc_matrix_flatten:
		return lb_build_builtin_proc_matrix_flatten(p, ce, tv)
	case BuiltinProc_unreachable:
		return lb_build_builtin_proc_unreachable(p, ce, tv)
	case BuiltinProc_raw_data:
		return lb_build_builtin_proc_raw_data(p, ce, tv)
	case BuiltinProc_alloca:
		return lb_build_builtin_proc_alloca(p, ce, tv)
	case BuiltinProc_cpu_relax:
		return lb_build_builtin_proc_cpu_relax(p, ce, tv)
	case BuiltinProc_debug_trap:
		fallthrough
	case BuiltinProc_trap:
		return lb_build_builtin_proc_debug_trap(p, ce, tv, id)
	case BuiltinProc_read_cycle_counter:
		return lb_build_builtin_proc_read_cycle_counter(p, ce, tv)
	case BuiltinProc_read_cycle_counter_frequency:
		return lb_build_builtin_proc_read_cycle_counter_frequency(p, ce, tv)
	case BuiltinProc_count_trailing_zeros:
		return lb_build_builtin_proc_count_trailing_zeros(p, ce, tv)
	case BuiltinProc_count_leading_zeros:
		return lb_build_builtin_proc_count_leading_zeros(p, ce, tv)
	case BuiltinProc_count_trailing_ones:
		return lb_build_builtin_proc_count_trailing_ones(p, ce, tv)
	case BuiltinProc_count_leading_ones:
		return lb_build_builtin_proc_count_leading_ones(p, ce, tv)
	case BuiltinProc_count_ones:
		return lb_build_builtin_proc_count_ones(p, ce, tv)
	case BuiltinProc_count_zeros:
		return lb_build_builtin_proc_count_zeros(p, ce, tv)
	case BuiltinProc_reverse_bits:
		return lb_build_builtin_proc_reverse_bits(p, ce, tv)
	case BuiltinProc_byte_swap:
		return lb_build_builtin_proc_byte_swap(p, ce, tv)
	case BuiltinProc_overflow_add:
		fallthrough
	case BuiltinProc_overflow_sub:
		fallthrough
	case BuiltinProc_overflow_mul:
		return lb_build_builtin_proc_overflow(p, ce, tv, id)
	case BuiltinProc_saturating_add:
		fallthrough
	case BuiltinProc_saturating_sub:
		return lb_build_builtin_proc_saturating_add_sub(p, ce, tv, id)
	case BuiltinProc_sqrt:
		return lb_build_builtin_proc_sqrt(p, ce, tv)
	case BuiltinProc_fused_mul_add:
		return lb_build_builtin_proc_fused_mul_add(p, ce, tv)
	case BuiltinProc_mem_copy:
		return lb_build_builtin_proc_mem_copy(p, ce, tv)
	case BuiltinProc_mem_copy_non_overlapping:
		return lb_build_builtin_proc_mem_copy_non_overlapping(p, ce, tv)
	case BuiltinProc_mem_zero:
		return lb_build_builtin_proc_mem_zero(p, ce, tv)
	case BuiltinProc_mem_zero_volatile:
		return lb_build_builtin_proc_mem_zero_volatile(p, ce, tv)
	case BuiltinProc_ptr_offset:
		return lb_build_builtin_proc_ptr_offset(p, ce, tv)
	case BuiltinProc_ptr_sub:
		return lb_build_builtin_proc_ptr_sub(p, ce, tv)
	case BuiltinProc_atomic_thread_fence:
		return lb_build_builtin_proc_atomic_thread_fence(p, ce, tv)
	case BuiltinProc_atomic_signal_fence:
		return lb_build_builtin_proc_atomic_signal_fence(p, ce, tv)
	case BuiltinProc_volatile_store:
		fallthrough
	case BuiltinProc_non_temporal_store:
		fallthrough
	case BuiltinProc_atomic_store:
		fallthrough
	case BuiltinProc_atomic_store_explicit:
		return lb_build_builtin_proc_volatile_non_temporal_atomic_store(p, ce, tv, id)
	case BuiltinProc_volatile_load:
		fallthrough
	case BuiltinProc_non_temporal_load:
		fallthrough
	case BuiltinProc_atomic_load:
		fallthrough
	case BuiltinProc_atomic_load_explicit:
		return lb_build_builtin_proc_volatile_non_temporal_atomic_load(p, ce, tv, id)
	case BuiltinProc_unaligned_store:
		return lb_build_builtin_proc_unaligned_store(p, ce, tv)
	case BuiltinProc_unaligned_load:
		return lb_build_builtin_proc_unaligned_load(p, ce, tv)
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
		return lb_build_builtin_proc_atomic_rmw(p, ce, tv, id)
	case BuiltinProc_atomic_compare_exchange_strong:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_weak:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_strong_explicit:
		fallthrough
	case BuiltinProc_atomic_compare_exchange_weak_explicit:
		return lb_build_builtin_proc_atomic_compare_exchange(p, ce, tv, id)
	case BuiltinProc_type_equal_proc:
		return lb_build_builtin_proc_type_equal_proc(p, ce, tv)
	case BuiltinProc_type_hasher_proc:
		return lb_build_builtin_proc_type_hasher_proc(p, ce, tv)
	case BuiltinProc_type_map_info:
		return lb_build_builtin_proc_type_map_info(p, ce, tv)
	case BuiltinProc_type_map_cell_info:
		return lb_build_builtin_proc_type_map_cell_info(p, ce, tv)
	case BuiltinProc_fixed_point_mul:
		fallthrough
	case BuiltinProc_fixed_point_div:
		fallthrough
	case BuiltinProc_fixed_point_mul_sat:
		fallthrough
	case BuiltinProc_fixed_point_div_sat:
		return lb_build_builtin_proc_fixed_point(p, ce, tv, id)
	case BuiltinProc_expect:
		return lb_build_builtin_proc_expect(p, ce, tv)
	case BuiltinProc_likely:
		fallthrough
	case BuiltinProc_unlikely:
		return lb_build_builtin_proc_likely_unlikely(p, ce, tv, id)
	case BuiltinProc_prefetch_read_instruction:
		fallthrough
	case BuiltinProc_prefetch_read_data:
		fallthrough
	case BuiltinProc_prefetch_write_instruction:
		fallthrough
	case BuiltinProc_prefetch_write_data:
		return lb_build_builtin_proc_prefetch(p, ce, tv, id)
	case BuiltinProc___entry_point:
		return lb_build_builtin_proc___entry_point(p, ce, tv)
	case BuiltinProc_syscall:
		return lb_build_builtin_proc_syscall(p, ce, tv)
	case BuiltinProc_syscall_bsd:
		return lb_build_builtin_proc_syscall_bsd(p, ce, tv)
	case BuiltinProc_objc_send:
		return lb_build_builtin_proc_objc_send(p, ce, tv)
	case BuiltinProc_objc_find_selector:
		return lb_build_builtin_proc_objc_find_selector(p, ce, tv)
	case BuiltinProc_objc_find_class:
		return lb_build_builtin_proc_objc_find_class(p, ce, tv)
	case BuiltinProc_objc_register_selector:
		return lb_build_builtin_proc_objc_register_selector(p, ce, tv)
	case BuiltinProc_objc_register_class:
		return lb_build_builtin_proc_objc_register_class(p, ce, tv)
	case BuiltinProc_objc_ivar_get:
		return lb_build_builtin_proc_objc_ivar_get(p, ce, tv)
	case BuiltinProc_objc_block:
		return lb_build_builtin_proc_objc_block(p, ce, tv)
	case BuiltinProc_objc_super:
		return lb_build_builtin_proc_objc_super(p, ce, tv)
	case BuiltinProc_c_va_start:
		return lb_build_builtin_proc_c_va_start(p, ce, tv)
	case BuiltinProc_c_va_end:
		return lb_build_builtin_proc_c_va_end(p, ce, tv)
	case BuiltinProc_c_va_copy:
		return lb_build_builtin_proc_c_va_copy(p, ce, tv)
	case BuiltinProc_c_va_arg:
		return lb_build_builtin_proc_c_va_arg(p, ce, tv)
	case BuiltinProc_constant_utf16_cstring:
		return lb_build_builtin_proc_constant_utf16_cstring(p, ce, tv)
	case BuiltinProc_wasm_memory_grow:
		return lb_build_builtin_proc_wasm_memory_grow(p, ce, tv)
	case BuiltinProc_wasm_memory_size:
		return lb_build_builtin_proc_wasm_memory_size(p, ce, tv)
	case BuiltinProc_wasm_memory_atomic_wait32:
		return lb_build_builtin_proc_wasm_memory_atomic_wait32(p, ce, tv)
	case BuiltinProc_wasm_memory_atomic_notify32:
		return lb_build_builtin_proc_wasm_memory_atomic_notify32(p, ce, tv)
	case BuiltinProc_x86_cpuid:
		return lb_build_builtin_proc_x86_cpuid(p, ce, tv)
	case BuiltinProc_x86_xgetbv:
		return lb_build_builtin_proc_x86_xgetbv(p, ce, tv)
	case BuiltinProc_valgrind_client_request:
		return lb_build_builtin_proc_valgrind_client_request(p, ce, tv)
	}

	gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 4564, "Unhandled built-in procedure")
	return lbValue{}
}
