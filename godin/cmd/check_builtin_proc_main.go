// Depends on: common.odin, checker types, all builtin proc handlers
package cmd

func check_builtin_procedure(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type) bool {
	ce := call.CallExpr

	if ce.inlining != ProcInlining_none {
		error(call, "Inlining operators are not allowed on built-in procedures")
	}

	bp := &builtin_procs[id]
	{
		var err string
		if ce.args.count < bp.arg_count {
			err = "Too few"
		} else if ce.args.count > bp.arg_count && !bp.variadic {
			err = "Too many"
		}
		if err != "" {
			expr := expr_to_string(ce.proc)
			error(ce.close, "%s arguments for '%s', expected %td, got %td",
				err, expr, bp.arg_count, ce.args.count)
			gb_string_free(expr)
			return false
		}
	}

	// Early type handling for specific builtins
	switch id {
	case BuiltinProc_size_of, BuiltinProc_align_of, BuiltinProc_offset_of,
		BuiltinProc_offset_of_by_string, BuiltinProc_type_info_of,
		BuiltinProc_typeid_of, BuiltinProc_len, BuiltinProc_cap,
		BuiltinProc_min, BuiltinProc_max,
		BuiltinProc_type_is_subtype_of, BuiltinProc_type_is_superset_of,
		BuiltinProc_objc_send, BuiltinProc_objc_find_selector,
		BuiltinProc_objc_find_class, BuiltinProc_objc_register_selector,
		BuiltinProc_objc_register_class,
		BuiltinProc_atomic_type_is_lock_free, BuiltinProc_has_target_feature,
		BuiltinProc_procedure_of, BuiltinProc_simd_indices:
		break
	case BuiltinProc_atomic_thread_fence, BuiltinProc_atomic_signal_fence:
		break
	case BuiltinProc_DIRECTIVE:
		bd := ce.proc.BasicDirective
		name := bd.name.string
		if name == "defined" || name == "config" {
			break
		}
		fallthrough
	default:
		if BuiltinProc__type_begin < id && id < BuiltinProc__type_end {
			check_expr_or_type(c, operand, ce.args[0])
		} else if ce.args.count > 0 {
			check_multi_expr(c, operand, ce.args[0])
		}
		break
	}

	builtin_name := builtin_procs[id].name

	if ce.args.count > 0 {
		if ce.args[0].kind == Ast_FieldValue {
			switch id {
			case BuiltinProc_soa_zip, BuiltinProc_quaternion:
				break
			default:
				error(call, "'field = value' calling is not allowed on built-in procedures")
				return false
			}
		}
	}

	if BuiltinProc__simd_begin < id && id < BuiltinProc__simd_end {
		ok := check_builtin_simd_operation(c, operand, call, id, type_hint)
		if !ok {
			operand.type = t_invalid
			operand.mode = Addressing_Value
		}
		operand.value = ExactValue{}
		operand.expr = call
		return ok
	}

	if BuiltinProc__atomic_begin < id && id < BuiltinProc__atomic_end {
		if build_context.metrics.arch == TargetArch_riscv64 {
			if !check_target_feature_is_enabled(String{Data: unsafe_.StringData("a"), Len: isize(len("a"))}, nil) {
				error(call, "missing required target feature \"a\" for atomics, enable it by setting a different -microarch or explicitly adding it through -target-features")
			}
		}
	}

	switch id {
	default:
		gb_assert_handler("Panic", 0, "check_builtin.cpp", 2803, "Implement built-in procedure: %.*s", builtin_name.len, builtin_name.data)
		break
	case BuiltinProc_objc_send, BuiltinProc_objc_find_selector, BuiltinProc_objc_find_class,
		BuiltinProc_objc_register_selector, BuiltinProc_objc_register_class,
		BuiltinProc_objc_ivar_get, BuiltinProc_objc_block, BuiltinProc_objc_super:
		return check_builtin_objc_procedure(c, operand, call, id, type_hint)
	case BuiltinProc_c_va_start, BuiltinProc_c_va_end, BuiltinProc_c_va_copy, BuiltinProc_c_va_arg:
		return check_builtin_c_procedure(c, operand, call, id, type_hint)
	case BuiltinProc___entry_point:
		operand.mode = Addressing_NoValue
		operand.type = nil
		mpsc_enqueue(&c.info.intrinsics_entry_point_usage, call)
		break
	case BuiltinProc_DIRECTIVE:
		return check_builtin_procedure_directive(c, operand, call, type_hint)
	case BuiltinProc_len, BuiltinProc_cap:
		if !check_builtin_procedure_len_cap(c, operand, call, id, type_hint, builtin_name) {
			return false
		}
	case BuiltinProc_size_of:
		if !check_builtin_procedure_size_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_align_of:
		if !check_builtin_procedure_align_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_offset_of:
		if !check_builtin_procedure_offset_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_offset_of_by_string:
		if !check_builtin_procedure_offset_of_by_string(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_of:
		if !check_builtin_procedure_type_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_info_of:
		if !check_builtin_procedure_type_info_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_typeid_of:
		if !check_builtin_procedure_typeid_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_swizzle:
		if !check_builtin_procedure_swizzle(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_complex:
		if !check_builtin_procedure_complex(c, operand, call, builtin_name, type_hint) {
			return false
		}
	case BuiltinProc_quaternion:
		if !check_builtin_procedure_quaternion(c, operand, call, builtin_name, type_hint) {
			return false
		}
	case BuiltinProc_real, BuiltinProc_imag:
		if !check_builtin_procedure_real_imag(c, operand, call, id, builtin_name, type_hint) {
			return false
		}
	case BuiltinProc_jmag, BuiltinProc_kmag:
		if !check_builtin_procedure_jmag_kmag(c, operand, call, id, builtin_name, type_hint) {
			return false
		}
	case BuiltinProc_conj:
		if !check_builtin_procedure_conj(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_expand_values:
		if !check_builtin_procedure_expand_values(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_compress_values:
		if !check_builtin_procedure_compress_values(c, operand, call, type_hint, builtin_name) {
			return false
		}
	case BuiltinProc_min:
		if !check_builtin_procedure_min(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_max:
		if !check_builtin_procedure_max(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_abs:
		if !check_builtin_procedure_abs(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_clamp:
		if !check_builtin_procedure_clamp(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_soa_zip:
		if !check_builtin_procedure_soa_zip(c, operand, call, type_hint, builtin_name) {
			return false
		}
	case BuiltinProc_soa_unzip:
		if !check_builtin_procedure_soa_unzip(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_transpose:
		if !check_builtin_procedure_transpose(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_outer_product:
		if !check_builtin_procedure_outer_product(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_hadamard_product:
		if !check_builtin_procedure_hadamard_product(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_matrix_flatten:
		if !check_builtin_procedure_matrix_flatten(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_is_package_imported:
		if !check_builtin_procedure_is_package_imported(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_has_target_feature:
		if !check_builtin_procedure_has_target_feature(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_constant_log2:
		if !check_builtin_procedure_constant_log2(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_constant_floor, BuiltinProc_constant_trunc,
		BuiltinProc_constant_ceil, BuiltinProc_constant_round:
		if !check_builtin_procedure_constant_rounding(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_soa_struct:
		if !check_builtin_procedure_soa_struct(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_concatenate:
		if !check_builtin_procedure_concatenate(c, operand, call, type_hint, builtin_name) {
			return false
		}
	case BuiltinProc_alloca:
		if !check_builtin_procedure_alloca(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_cpu_relax, BuiltinProc_unreachable,
		BuiltinProc_trap, BuiltinProc_debug_trap:
		operand.mode = Addressing_NoValue
		if id == BuiltinProc_unreachable || id == BuiltinProc_trap || id == BuiltinProc_debug_trap {
			operand.type = nil
		}
	case BuiltinProc_raw_data:
		if !check_builtin_procedure_raw_data(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_read_cycle_counter_frequency:
		if build_context.metrics.arch != TargetArch_arm64 {
			error(call, "'%.*s' is only allowed on arm64 targets", builtin_name.len, builtin_name.data)
			return false
		}
		operand.mode = Addressing_Value
		operand.type = t_i64
	case BuiltinProc_read_cycle_counter:
		operand.mode = Addressing_Value
		operand.type = t_i64
	case BuiltinProc_count_ones, BuiltinProc_count_zeros,
		BuiltinProc_count_trailing_zeros, BuiltinProc_count_leading_zeros,
		BuiltinProc_count_trailing_ones, BuiltinProc_count_leading_ones,
		BuiltinProc_reverse_bits:
		if !check_builtin_procedure_bit_count(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_byte_swap:
		if !check_builtin_procedure_byte_swap(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_overflow_add, BuiltinProc_overflow_sub, BuiltinProc_overflow_mul:
		if !check_builtin_procedure_overflow(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_saturating_add, BuiltinProc_saturating_sub:
		if !check_builtin_procedure_saturating(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_sqrt:
		if !check_builtin_procedure_sqrt(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_fused_mul_add:
		if !check_builtin_procedure_fused_mul_add(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_mem_copy, BuiltinProc_mem_copy_non_overlapping:
		if !check_builtin_procedure_mem_copy(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_mem_zero, BuiltinProc_mem_zero_volatile:
		if !check_builtin_procedure_mem_zero(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_ptr_offset:
		if !check_builtin_procedure_ptr_offset(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_ptr_sub:
		if !check_builtin_procedure_ptr_sub(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_type_is_lock_free:
		if !check_builtin_procedure_atomic_type_is_lock_free(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_thread_fence, BuiltinProc_atomic_signal_fence:
		if !check_builtin_procedure_atomic_fence(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_volatile_store, BuiltinProc_unaligned_store,
		BuiltinProc_non_temporal_store, BuiltinProc_atomic_store:
		if !check_builtin_procedure_atomic_store(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_store_explicit:
		if !check_builtin_procedure_atomic_store_explicit(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_volatile_load, BuiltinProc_unaligned_load,
		BuiltinProc_non_temporal_load, BuiltinProc_atomic_load:
		if !check_builtin_procedure_atomic_load(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_load_explicit:
		if !check_builtin_procedure_atomic_load_explicit(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_add, BuiltinProc_atomic_sub,
		BuiltinProc_atomic_and, BuiltinProc_atomic_nand,
		BuiltinProc_atomic_or, BuiltinProc_atomic_xor,
		BuiltinProc_atomic_exchange:
		if !check_builtin_procedure_atomic_rmw(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_add_explicit, BuiltinProc_atomic_sub_explicit,
		BuiltinProc_atomic_and_explicit, BuiltinProc_atomic_nand_explicit,
		BuiltinProc_atomic_or_explicit, BuiltinProc_atomic_xor_explicit,
		BuiltinProc_atomic_exchange_explicit:
		if !check_builtin_procedure_atomic_rmw_explicit(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_compare_exchange_strong, BuiltinProc_atomic_compare_exchange_weak:
		if !check_builtin_procedure_atomic_cas(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_atomic_compare_exchange_strong_explicit,
		BuiltinProc_atomic_compare_exchange_weak_explicit:
		if !check_builtin_procedure_atomic_cas_explicit(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_fixed_point_mul, BuiltinProc_fixed_point_div,
		BuiltinProc_fixed_point_mul_sat, BuiltinProc_fixed_point_div_sat:
		if !check_builtin_procedure_fixed_point(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_expect:
		if !check_builtin_procedure_expect(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_likely, BuiltinProc_unlikely:
		if !check_builtin_procedure_likely(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_prefetch_read_instruction, BuiltinProc_prefetch_read_data,
		BuiltinProc_prefetch_write_instruction, BuiltinProc_prefetch_write_data:
		if !check_builtin_procedure_prefetch(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_syscall:
		if !check_builtin_procedure_syscall(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_syscall_bsd:
		if !check_builtin_procedure_syscall_bsd(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_base_type:
		if operand.mode != Addressing_Type {
			error(operand.expr, "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
		} else {
			operand.type = base_type(operand.type)
		}
		operand.mode = Addressing_Type
	case BuiltinProc_type_core_type:
		if operand.mode != Addressing_Type {
			error(operand.expr, "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
		} else {
			operand.type = core_type(operand.type)
		}
		operand.mode = Addressing_Type
	case BuiltinProc_type_elem_type:
		if operand.mode != Addressing_Type {
			error(operand.expr, "Expected a type for '%.*s'", builtin_name.len, builtin_name.data)
		} else {
			bt := base_type(operand.type)
			switch bt.kind {
			case Type_Basic:
				switch bt.Basic.kind {
				case Basic_complex32:
					operand.type = t_f16
				case Basic_complex64:
					operand.type = t_f32
				case Basic_complex128:
					operand.type = t_f64
				case Basic_quaternion64:
					operand.type = t_f16
				case Basic_quaternion128:
					operand.type = t_f32
				case Basic_quaternion256:
					operand.type = t_f64
				}
			case Type_Pointer:
				operand.type = bt.Pointer.elem
			case Type_Array:
				operand.type = bt.Array.elem
			case Type_EnumeratedArray:
				operand.type = bt.EnumeratedArray.elem
			case Type_Slice:
				operand.type = bt.Slice.elem
			case Type_DynamicArray:
				operand.type = bt.DynamicArray.elem
			case Type_SimdVector:
				operand.type = bt.SimdVector.elem
			}
		}
		operand.mode = Addressing_Type
	case BuiltinProc_type_convert_variants_to_pointers:
		if !check_builtin_procedure_type_variants_to_ptrs(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_integer_to_unsigned:
		if !check_builtin_procedure_type_int_to_unsigned(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_integer_to_signed:
		if !check_builtin_procedure_type_int_to_signed(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_merge:
		if !check_builtin_procedure_type_merge(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_is_boolean, BuiltinProc_type_is_bit_field,
		BuiltinProc_type_is_integer, BuiltinProc_type_is_rune,
		BuiltinProc_type_is_float, BuiltinProc_type_is_complex,
		BuiltinProc_type_is_quaternion, BuiltinProc_type_is_string,
		BuiltinProc_type_is_string16, BuiltinProc_type_is_cstring,
		BuiltinProc_type_is_cstring16, BuiltinProc_type_is_typeid,
		BuiltinProc_type_is_any, BuiltinProc_type_is_endian_platform,
		BuiltinProc_type_is_endian_little, BuiltinProc_type_is_endian_big,
		BuiltinProc_type_is_unsigned, BuiltinProc_type_is_numeric,
		BuiltinProc_type_is_ordered, BuiltinProc_type_is_ordered_numeric,
		BuiltinProc_type_is_indexable, BuiltinProc_type_is_sliceable,
		BuiltinProc_type_is_comparable, BuiltinProc_type_is_simple_compare,
		BuiltinProc_type_is_nearly_simple_compare, BuiltinProc_type_is_dereferenceable,
		BuiltinProc_type_is_valid_map_key, BuiltinProc_type_is_valid_matrix_elements,
		BuiltinProc_type_is_named, BuiltinProc_type_is_pointer,
		BuiltinProc_type_is_multi_pointer, BuiltinProc_type_is_array,
		BuiltinProc_type_is_enumerated_array, BuiltinProc_type_is_slice,
		BuiltinProc_type_is_dynamic_array, BuiltinProc_type_is_map,
		BuiltinProc_type_is_struct, BuiltinProc_type_is_union,
		BuiltinProc_type_is_enum, BuiltinProc_type_is_proc,
		BuiltinProc_type_is_bit_set, BuiltinProc_type_is_simd_vector,
		BuiltinProc_type_is_matrix, BuiltinProc_type_is_raw_union,
		BuiltinProc_type_is_specialized_polymorphic_record,
		BuiltinProc_type_is_unspecialized_polymorphic_record,
		BuiltinProc_type_has_nil:
		operand.value = exact_value_bool(false)
		if operand.mode != Addressing_Type {
			str := expr_to_string(ce.args[0])
			error(operand.expr, "Expected a type for '%.*s', got '%s'", builtin_name.len, builtin_name.data, str)
			gb_string_free(str)
		} else {
			i := id - int32(BuiltinProc__type_simple_boolean_begin)
			procedure := builtin_type_is_procs[i]
			if procedure == nil {
				gb_assert_handler("Panic", 0, "check_builtin.cpp", 6986, "%.*s", builtin_name.len, builtin_name.data)
			}
			ok := procedure(operand.type)
			operand.value = exact_value_bool(ok)
		}
		operand.mode = Addressing_Constant
		operand.type = t_untyped_bool
	case BuiltinProc_type_is_matrix_row_major, BuiltinProc_type_is_matrix_column_major:
		if !check_builtin_procedure_type_is_matrix_major(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_type_has_field:
		if !check_builtin_procedure_type_has_field(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_has_shared_fields:
		if !check_builtin_procedure_type_has_shared_fields(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_field_type:
		if !check_builtin_procedure_type_field_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_field_bit_offset, BuiltinProc_type_field_bit_size:
		if !check_builtin_procedure_type_field_bit(c, operand, call, id, builtin_name) {
			return false
		}
	case BuiltinProc_type_is_specialization_of:
		if !check_builtin_procedure_type_is_specialization_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_is_variant_of:
		if !check_builtin_procedure_type_is_variant_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_union_tag_type:
		if !check_builtin_procedure_type_union_tag_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_union_tag_offset:
		if !check_builtin_procedure_type_union_tag_offset(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_union_base_tag_value:
		if !check_builtin_procedure_type_union_base_tag_value(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_bit_set_elem_type:
		if !check_builtin_procedure_type_bit_set_elem_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_bit_set_underlying_type:
		if !check_builtin_procedure_type_bit_set_underlying_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_union_variant_count:
		if !check_builtin_procedure_type_union_variant_count(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_variant_type_of:
		if !check_builtin_procedure_type_variant_type_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_variant_index_of:
		if !check_builtin_procedure_type_variant_index_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_struct_field_count:
		if !check_builtin_procedure_type_struct_field_count(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_struct_has_implicit_padding:
		if !check_builtin_procedure_type_struct_has_implicit_padding(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_proc_parameter_count:
		if !check_builtin_procedure_type_proc_param_count(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_proc_return_count:
		if !check_builtin_procedure_type_proc_return_count(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_proc_parameter_type:
		if !check_builtin_procedure_type_proc_param_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_proc_return_type:
		if !check_builtin_procedure_type_proc_return_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_polymorphic_record_parameter_count:
		if !check_builtin_procedure_type_poly_record_param_count(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_polymorphic_record_parameter_value:
		if !check_builtin_procedure_type_poly_record_param_value(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_is_subtype_of:
		if !check_builtin_procedure_type_is_subtype_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_is_superset_of:
		if !check_builtin_procedure_type_is_superset_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_field_index_of:
		if !check_builtin_procedure_type_field_index_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_fixed_capacity_dynamic_array_len_offset:
		if !check_builtin_procedure_type_fcd_len_offset(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_bit_set_backing_type:
		if !check_builtin_procedure_type_bit_set_backing_type(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_enum_is_contiguous:
		if !check_builtin_procedure_type_enum_is_contiguous(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_equal_proc:
		if !check_builtin_procedure_type_equal_proc(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_hasher_proc:
		if !check_builtin_procedure_type_hasher_proc(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_map_info:
		if !check_builtin_procedure_type_map_info(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_map_cell_info:
		if !check_builtin_procedure_type_map_cell_info(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_type_canonical_name:
		if !check_builtin_procedure_type_canonical_name(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_procedure_of:
		if !check_builtin_procedure_procedure_of(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_constant_utf16_cstring:
		if !check_builtin_procedure_constant_utf16_cstring(c, operand, call, type_hint, builtin_name) {
			return false
		}
	case BuiltinProc_wasm_memory_grow:
		if !check_builtin_procedure_wasm_memory_grow(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_wasm_memory_size:
		if !check_builtin_procedure_wasm_memory_size(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_wasm_memory_atomic_wait32:
		if !check_builtin_procedure_wasm_atomic_wait32(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_wasm_memory_atomic_notify32:
		if !check_builtin_procedure_wasm_atomic_notify32(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_x86_cpuid:
		if !check_builtin_procedure_x86_cpuid(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_x86_xgetbv:
		if !check_builtin_procedure_x86_xgetbv(c, operand, call, builtin_name) {
			return false
		}
	case BuiltinProc_valgrind_client_request:
		if !check_builtin_procedure_valgrind(c, operand, call, builtin_name) {
			return false
		}
	}
	return true
}
