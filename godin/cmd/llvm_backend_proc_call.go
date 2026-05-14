package cmd

func lb_value_to_array(p *lbProcedure, allocator gbAllocator, value lbValue) []lbValue {
	t := base_type(value.Type)
	if t == nil {
		return nil
	} else if is_type_tuple(t) {
		array := make([]lbValue, 0, len(t.Tuple.Variables))
		lb_append_tuple_values(p, &array, value)
		return array
	} else {
		return []lbValue{value}
	}
}

func lb_emit_call_internal(p *lbProcedure, value lbValue, return_ptr lbValue, processed_args []lbValue, abi_rt *Type, context_ptr lbAddr, inlining ProcInlining, tailing ProcTailing) lbValue {
	gb_assert_handler("Assertion Failure", "p.Module.Ctx == LLVMGetTypeContext(LLVMTypeOf(value.Value))", "src/llvm_backend_proc.cpp", 886, 0)

	arg_count := uint(len(processed_args))
	if return_ptr.Value != 0 {
		arg_count += 1
	}
	if context_ptr.Addr.Value != 0 {
		arg_count += 1
	}

	args := make([]LLVMValueRef, arg_count)
	arg_index := isize(0)
	if return_ptr.Value != 0 {
		args[arg_index] = return_ptr.Value
		arg_index++
	}

	for _, arg := range processed_args {
		if is_type_proc(arg.Type) {
			arg.Value = LLVMBuildPointerCast(p.Builder, arg.Value, lb_type(p.Module, arg.Type), "")
		}
		args[arg_index] = arg.Value
		arg_index++
	}

	if context_ptr.Addr.Value != 0 {
		cp := context_ptr.Addr.Value
		cp = LLVMBuildPointerCast(p.Builder, cp, lb_type(p.Module, t_rawptr), "")
		args[arg_index] = cp
		arg_index++
	}

	gb_assert_handler("Assertion Failure", "arg_index == isize(arg_count)", "src/llvm_backend_proc.cpp", 916, 0)

	curr_block := LLVMGetInsertBlock(p.Builder)
	gb_assert_handler("Assertion Failure", "curr_block != p.DeclBlock.Block", "src/llvm_backend_proc.cpp", 919, 0)

	{
		proc_type := base_type(value.Type)
		gb_assert_handler("Assertion Failure", "proc_type.Kind == Type_Proc", "src/llvm_backend_proc.cpp", 923, 0)

		fnp := lb_type_internal_for_procedures_raw(p.Module, proc_type)
		ftp := LLVMPointerType(fnp, 0)
		fn := value.Value
		if !lb_is_type_kind(LLVMTypeOf(value.Value), LLVMFunctionTypeKind) {
			fn = LLVMBuildPointerCast(p.Builder, fn, ftp, "")
		}
		gb_assert_handler("Assertion Failure", "lb_is_type_kind(fnp, LLVMFunctionTypeKind)", "src/llvm_backend_proc.cpp", 931, LLVMPrintTypeToString(fnp))

		ft := map_must_get(&p.Module.FunctionTypeMap, base_type(value.Type))

		{
			param_count := uint(LLVMCountParamTypes(fnp))
			gb_assert_handler("Assertion Failure", "arg_count >= param_count", "src/llvm_backend_proc.cpp", 937, 0)

			param_types := make([]LLVMTypeRef, param_count)
			LLVMGetParamTypes(fnp, &param_types[0])

			for i := uint(0); i < param_count; i++ {
				param_type := param_types[i]
				arg_type := LLVMTypeOf(args[i])
				if LB_USE_NEW_PASS_SYSTEM &&
					arg_type != param_type {
					arg_kind := LLVMGetTypeKind(arg_type)
					param_kind := LLVMGetTypeKind(param_type)
					if arg_kind == param_kind {
						if arg_kind == LLVMPointerTypeKind {
							args[i] = LLVMBuildPointerCast(p.Builder, args[i], param_type, "")
							arg_type = param_type
							continue
						}
					}
				}

				gb_assert_handler("Assertion Failure",
					"arg_type == param_type",
					"src/llvm_backend_proc.cpp", 961,
					LLVMPrintTypeToString(arg_type),
					LLVMPrintTypeToString(param_type),
					LLVMPrintValueToString(args[i]),
					LLVMPrintTypeToString(fnp))
			}
		}

		ret := LLVMBuildCall2(p.Builder, fnp, fn, args, uint(arg_count), "")

		llvm_cc := lb_calling_convention_map[proc_type.Proc.CallingConvention]
		LLVMSetInstructionCallConv(ret, llvm_cc)

		param_offset := LLVMAttributeIndex_FirstArgIndex
		if return_ptr.Value != 0 {
			param_offset += 1

			LLVMAddCallSiteAttribute(ret, 1, lb_create_enum_attribute_with_type(p.Module.Ctx, "sret", LLVMTypeOf(args[0])))
		}

		for i := range ft.Args {
			attribute := ft.Args[i].Attribute
			if attribute != 0 {
				LLVMAddCallSiteAttribute(ret, param_offset+LLVMAttributeIndex(i), attribute)
			}
		}

		switch inlining {
		case ProcInlining_none:
		case ProcInlining_inline:
			LLVMAddCallSiteAttribute(ret, LLVMAttributeIndex_FunctionIndex, lb_create_enum_attribute(p.Module.Ctx, "alwaysinline"))
		case ProcInlining_no_inline:
			LLVMAddCallSiteAttribute(ret, LLVMAttributeIndex_FunctionIndex, lb_create_enum_attribute(p.Module.Ctx, "noinline"))
		}

		switch tailing {
		case ProcTailing_none:
		case ProcTailing_must_tail:
			LLVMSetTailCall(ret, true)
			LLVMSetTailCallKind(ret, LLVMTailCallKindMustTail)
		}

		res := lbValue{Value: ret, Type: abi_rt}
		return res
	}
}

func lb_lookup_runtime_procedure(m *lbModule, name string) lbValue {
	pkg := m.Info.RuntimePackage
	e := scope_lookup_current(pkg.Scope, string_interner_insert(name))
	gb_assert_handler("Assertion Failure", "e != nil", "src/llvm_backend_proc.cpp", 1024, name)
	return lb_find_procedure_value_from_entity(m, e)
}

func lb_emit_runtime_call(p *lbProcedure, c_name string, args []lbValue) lbValue {
	name := c_name
	proc := lb_lookup_runtime_procedure(p.Module, name)
	return lb_emit_call(p, proc, args)
}

func lb_emit_conjugate(p *lbProcedure, val lbValue, type_ *Type) lbValue {
	var res lbValue
	t := val.Type
	if is_type_complex(t) {
		real := lb_emit_struct_ev(p, val, 0)
		imag := lb_emit_struct_ev(p, val, 1)
		imag = lb_emit_unary_arith(p, Token_Sub, imag, imag.Type)
		fields := []lbValue{real, imag}
		return lb_build_struct_value(p, type_, fields, isize(len(fields)))
	} else if is_type_quaternion(t) {
		real := lb_emit_struct_ev(p, val, 3)
		imag := lb_emit_struct_ev(p, val, 0)
		jmag := lb_emit_struct_ev(p, val, 1)
		kmag := lb_emit_struct_ev(p, val, 2)
		imag = lb_emit_unary_arith(p, Token_Sub, imag, imag.Type)
		jmag = lb_emit_unary_arith(p, Token_Sub, jmag, jmag.Type)
		kmag = lb_emit_unary_arith(p, Token_Sub, kmag, kmag.Type)
		fields := []lbValue{imag, jmag, kmag, real}
		return lb_build_struct_value(p, type_, fields, isize(len(fields)))
	} else if is_type_array_like(t) {
		res = lb_addr_get_ptr(p, lb_add_local_generated(p, type_, true))
		elem_type := base_array_type(t)
		count := get_array_type_count(t)
		for i := i64(0); i < count; i++ {
			dst := lb_emit_array_epi_proc(p, res, i)
			elem := lb_emit_struct_ev(p, val, i32(i))
			elem = lb_emit_conjugate(p, elem, elem_type)
			lb_emit_store(p, dst, elem)
		}
	} else if is_type_matrix(t) {
		mt := base_type(t)
		gb_assert_handler("Assertion Failure", "mt.Kind == Type_Matrix", "src/llvm_backend_proc.cpp", 1069, 0)
		elem_type := mt.Matrix.Elem
		res = lb_addr_get_ptr(p, lb_add_local_generated(p, type_, true))
		for j := i64(0); j < mt.Matrix.ColumnCount; j++ {
			for i := i64(0); i < mt.Matrix.RowCount; i++ {
				dst := lb_emit_matrix_epi(p, res, i, j)
				elem := lb_emit_matrix_ev(p, val, i, j)
				elem = lb_emit_conjugate(p, elem, elem_type)
				lb_emit_store(p, dst, elem)
			}
		}
	}
	return lb_emit_load(p, res)
}

func lb_emit_call(p *lbProcedure, value lbValue, args []lbValue, inlining ...ProcInlining) lbValue {
	inl := ProcInlining_none
	if len(inlining) > 0 {
		inl = inlining[0]
	}

	m := p.Module

	pt := base_type(value.Type)
	gb_assert_handler("Assertion Failure", "pt.Kind == Type_Proc", "src/llvm_backend_proc.cpp", 1088, 0)
	results := pt.Proc.Results

	context_ptr := lbAddr{}
	if pt.Proc.CallingConvention == ProcCC_Odin {
		context_ptr = lb_find_or_generate_context_ptr(p)
	}

	defer func() {
		if pt.Proc.Diverging {
			LLVMBuildUnreachable(p.Builder)
		}
	}()

	is_c_vararg := pt.Proc.CVarArg
	param_count := isize(pt.Proc.ParamCount)
	if is_c_vararg {
		gb_assert_handler("Assertion Failure", "param_count-1 <= len(args)", "src/llvm_backend_proc.cpp", 1103, 0)
		param_count -= 1
	} else {
		gb_assert_handler("Assertion Failure", "param_count == len(args)", "src/llvm_backend_proc.cpp", 1106, param_count, len(args), LLVMPrintValueToString(value.Value))
	}

	var result lbValue

	ignored_args := isize(0)
	processed_args := make([]lbValue, 0, len(args))

	{
		is_odin_cc := is_calling_convention_odin(pt.Proc.CallingConvention)

		ft := lb_get_function_type(m, pt)
		return_by_pointer := ft.Ret.Kind == lbArg_Indirect
		split_returns := ft.MultipleReturnOriginalType != nil

		param_index := uint(0)
		for i := isize(0); i < param_count; i++ {
			e := pt.Proc.Params.Tuple.Variables[i]
			if e.Kind != Entity_Variable {
				continue
			}
			gb_assert_handler("Assertion Failure", "e.Flags&EntityFlag_Param != 0", "src/llvm_backend_proc.cpp", 1128, 0)

			original_type := e.Type
			arg := &ft.Args[param_index]
			if arg.Kind == lbArg_Ignore {
				param_index += 1
				ignored_args += 1
				continue
			}

			x := lb_emit_conv(p, args[i], original_type)
			xt := lb_type(p.Module, x.Type)

			if arg.Kind == lbArg_Direct {
				abi_type := arg.CastType
				if abi_type == 0 {
					abi_type = arg.Type
				}
				if xt == abi_type {
					processed_args = append(processed_args, x)
				} else {
					x.Value = OdinLLVMBuildTransmute(p, x.Value, abi_type)
					processed_args = append(processed_args, x)
				}

			} else if arg.Kind == lbArg_Indirect {
				var ptr lbValue
				if arg.IsByval {
					if is_odin_cc {
						if are_types_identical(original_type, t_source_code_location) {
							ptr = lb_address_from_load_or_generate_local(p, x)
						}
					}
					if ptr.Value == 0 {
						ptr = lb_copy_value_to_ptr(p, x, original_type, arg.ByvalAlignment)
					}
				} else if is_odin_cc {
					if LLVMIsConstant(x.Value) != 0 {
						addr := lb_add_global_generated_from_procedure(p, original_type, x)
						lb_make_global_private_const(addr)
						ptr = addr.Addr
					} else {
						ptr = lb_address_from_load_or_generate_local(p, x)
					}
				} else {
					ptr = lb_copy_value_to_ptr(p, x, original_type, 16)
				}
				processed_args = append(processed_args, ptr)
			}

			param_index += 1
		}

		if is_c_vararg {
			for i := len(processed_args); i < len(args); i++ {
				processed_args = append(processed_args, args[i])
			}
		}

		rt := reduce_tuple_to_single_type(results)
		original_rt := rt
		if split_returns {
			gb_assert_handler("Assertion Failure", "rt.Kind == Type_Tuple", "src/llvm_backend_proc.cpp", 1196, 0)
			for j := isize(0); j < isize(len(rt.Tuple.Variables))-1; j++ {
				partial_return_type := rt.Tuple.Variables[j].Type
				partial_return_ptr := lb_add_local(p, partial_return_type, nil).Addr
				processed_args = append(processed_args, partial_return_ptr)
			}
			rt = reduce_tuple_to_single_type(rt.Tuple.Variables[len(rt.Tuple.Variables)-1].Type)
		}

		if return_by_pointer {
			return_ptr := lb_add_local_generated(p, rt, true).Addr
			lb_emit_call_internal(p, value, return_ptr, processed_args, nil, context_ptr, inl, ProcTailing_none)
			result = lb_emit_load(p, return_ptr)
		} else if rt != nil {
			result = lb_emit_call_internal(p, value, lbValue{}, processed_args, rt, context_ptr, inl, ProcTailing_none)
			if ft.Ret.CastType != 0 {
				result.Value = OdinLLVMBuildTransmute(p, result.Value, ft.Ret.CastType)
			}
			result.Value = OdinLLVMBuildTransmute(p, result.Value, ft.Ret.Type)
			result.Type = rt
			if LLVMTypeOf(result.Value) == LLVMInt1TypeInContext(p.Module.Ctx) {
				result.Type = t_llvm_bool
			}
			if !is_type_tuple(rt) {
				result = lb_emit_conv(p, result, rt)
			}
		} else {
			lb_emit_call_internal(p, value, lbValue{}, processed_args, nil, context_ptr, inl, ProcTailing_none)
		}

		if original_rt != rt {
			gb_assert_handler("Assertion Failure", "split_returns", "src/llvm_backend_proc.cpp", 1232, 0)
			gb_assert_handler("Assertion Failure", "is_type_tuple(original_rt)", "src/llvm_backend_proc.cpp", 1233, 0)

			result_ptr := lb_add_local_generated(p, original_rt, false).Addr
			ret_count := len(original_rt.Tuple.Variables)

			tuple_fix_values := make([]lbValue, ret_count)
			tuple_geps := make([]lbValue, ret_count)

			offset := ft.OriginalArgCount - ignored_args
			for j := isize(0); j < isize(ret_count)-1; j++ {
				ret_arg_ptr := processed_args[offset+int(j)]
				ret_arg := lb_emit_load(p, ret_arg_ptr)
				tuple_fix_values[j] = ret_arg
			}
			tuple_fix_values[ret_count-1] = result

			result = lb_emit_load(p, result_ptr)

			tf := lbTupleFix{Values: tuple_fix_values}
			map_set(&p.TupleFixMap, result_ptr.Value, tf)
			map_set(&p.TupleFixMap, result.Value, tf)
		}
	}

	the_proc_value := value.Value

	if LLVMIsAConstantExpr(the_proc_value) != 0 {
		the_proc_value = LLVMGetOperand(the_proc_value, 0)
	}
	found := map_get(&p.Module.ProcedureValues, the_proc_value)
	if found != nil {
		e := *found
		if e != nil && entity_has_deferred_procedure(e) {
			kind := e.Procedure.DeferredProcedure.Kind
			deferred_entity := e.Procedure.DeferredProcedure.Entity
			deferred := lb_find_procedure_value_from_entity(p.Module, deferred_entity)

			by_ptr := false
			in_args := args
			var result_as_args []lbValue
			switch kind {
			case DeferredProcedure_none:
			case DeferredProcedure_in_by_ptr:
				by_ptr = true
				fallthrough
			case DeferredProcedure_in:
				result_as_args = make([]lbValue, len(in_args))
				copy(result_as_args, in_args)
			case DeferredProcedure_out_by_ptr:
				by_ptr = true
				fallthrough
			case DeferredProcedure_out:
				result_as_args = lb_value_to_array(p, heap_allocator(), result)
			case DeferredProcedure_in_out_by_ptr:
				by_ptr = true
				fallthrough
			case DeferredProcedure_in_out:
				{
					out_args := lb_value_to_array(p, heap_allocator(), result)
					result_as_args = make([]lbValue, len(in_args)+len(out_args))
					copy(result_as_args, in_args)
					copy(result_as_args[len(in_args):], out_args)
				}
			}
			if by_ptr {
				for j := range result_as_args {
					arg_ptr := lb_address_from_load_or_generate_local(p, result_as_args[j])
					result_as_args[j] = arg_ptr
				}
			}

			lb_add_defer_proc(p, p.ScopeIndex, deferred, result_as_args, e.Token.Pos)
		}
	}

	return result
}
