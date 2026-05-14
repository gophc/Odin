package cmd

func lb_handle_param_value(p *lbProcedure, parameter_type *Type, param_value ParameterValue, procedure_type *TypeProc, call_expression *Ast) lbValue {
	switch param_value.Kind {
	case ParameterValue_Constant:
		if is_type_constant_type(parameter_type) {
			res := lb_const_value(p.Module, parameter_type, param_value.Value)
			return res
		} else {
			ev := param_value.Value
			arg := lbValue{}
			type_ := type_of_expr(param_value.OriginalAstExpr)
			if type_ != nil {
				arg = lb_const_value(p.Module, type_, ev)
			} else {
				arg = lb_const_value(p.Module, parameter_type, param_value.Value)
			}
			return lb_emit_conv(p, arg, parameter_type)
		}

	case ParameterValue_Nil:
		return lb_const_nil(p.Module, parameter_type)

	case ParameterValue_Location:
		proc_name := String{}
		if p.Entity != nil {
			proc_name = p.Entity.Token.String
		}

		ce := &call_expression.CallExpr
		pos := ast_token(ce.Proc).Pos

		return lb_emit_source_code_location_as_global(p, proc_name, pos)

	case ParameterValue_Expression:
		orig := param_value.OriginalAstExpr
		if orig.Kind == Ast_BasicDirective {
			expr := expr_to_string(call_expression, temporary_allocator())
			return lb_const_string(p.Module, expr)
		}

		param_idx := isize(-1)
		param_str := String{}
		{
			call := unparen_expr(orig)
			gb_assert_handler("Assertion Failure", "call.Kind == Ast_CallExpr", "src/llvm_backend_proc.cpp", 4613, 0)
			ce2 := &call.CallExpr
			gb_assert_handler("Assertion Failure", "ce2.Proc.Kind == Ast_BasicDirective", "src/llvm_backend_proc.cpp", 4615, 0)
			gb_assert_handler("Assertion Failure", "len(ce2.Args) == 1", "src/llvm_backend_proc.cpp", 4616, 0)
			target := ce2.Args[0]
			gb_assert_handler("Assertion Failure", "target.Kind == Ast_Ident", "src/llvm_backend_proc.cpp", 4618, 0)
			target_str := target.Ident.Token.String

			param_idx = lookup_procedure_parameter(procedure_type, target_str)
			param_str = target_str
		}
		gb_assert_handler("Assertion Failure", "param_idx >= 0", "src/llvm_backend_proc.cpp", 4624, 0)

		var target_expr *Ast
		ce3 := &call_expression.CallExpr

		if isize(len(ce3.SplitArgs.Positional)) > param_idx {
			target_expr = ce3.SplitArgs.Positional[param_idx]
		}

		for _, arg := range ce3.SplitArgs.Named {
			fv := &arg.FieldValue
			gb_assert_handler("Assertion Failure", "fv.Field.Kind == Ast_Ident", "src/llvm_backend_proc.cpp", 4637, 0)
			name := fv.Field.Ident.Token.String
			if name == param_str {
				target_expr = fv.Value
				break
			}
		}

		expr := expr_to_string(target_expr, temporary_allocator())
		return lb_const_string(p.Module, expr)

	case ParameterValue_Value:
		return lb_build_expr(p, param_value.AstValue)
	}
	return lb_const_nil(p.Module, parameter_type)
}

func lb_build_call_expr_internal(p *lbProcedure, expr *Ast) lbValue {
	m := p.Module

	tv := type_and_value_of_expr(expr)

	ce := &expr.CallExpr

	proc_tv := type_and_value_of_expr(ce.Proc)
	proc_mode := proc_tv.Mode
	if proc_mode == Addressing_Type {
		gb_assert_handler("Assertion Failure", "len(ce.Args) == 1", "src/llvm_backend_proc.cpp", 4693, 0)
		x := lb_build_expr(p, ce.Args[0])
		y := lb_emit_conv(p, x, tv.Type)
		y.Type = tv.Type
		return y
	}

	proc_expr := unparen_expr(ce.Proc)
	proc_entity := entity_of_node(proc_expr)

	if proc_mode == Addressing_Builtin {
		id := BuiltinProc_Invalid
		if proc_entity != nil {
			id = BuiltinProcId(proc_entity.Builtin.ID)
		} else {
			id = BuiltinProc_DIRECTIVE
		}
		return lb_build_builtin_proc(p, expr, tv, id)
	}

	is_objc_call := proc_entity != nil && proc_entity.Procedure.IsObjcImplOrImport

	value := lbValue{}

	if proc_entity != nil {
		if proc_entity.Flags&EntityFlag_Disabled != 0 {
			gb_assert_handler("Assertion Failure", "tv.Type == nil", "src/llvm_backend_proc.cpp", 4720, 0)
			return lbValue{}
		}
	}

	if proc_expr.TAV.Mode == Addressing_Constant {
		v := proc_expr.TAV.Value
		switch v.Kind {
		case ExactValue_Integer:
			u := big_int_to_u64(&v.ValueInteger)
			x := lbValue{}
			x.Value = LLVMConstInt(lb_type(m, t_uintptr), u, false)
			x.Type = t_uintptr
			x = lb_emit_conv(p, x, t_rawptr)
			value = lb_emit_conv(p, x, proc_expr.TAV.Type)
		case ExactValue_Pointer:
			u := u64(v.ValuePointer)
			x := lbValue{}
			x.Value = LLVMConstInt(lb_type(m, t_uintptr), u, false)
			x.Type = t_uintptr
			x = lb_emit_conv(p, x, t_rawptr)
			value = lb_emit_conv(p, x, proc_expr.TAV.Type)
		}
	}

	if is_objc_call {
		value.Type = proc_tv.Type
	} else if value.Value == 0 {
		value = lb_build_expr(p, proc_expr)
	}

	gb_assert_handler("Assertion Failure", "value.Value != 0 || is_objc_call", "src/llvm_backend_proc.cpp", 4757, 0)
	proc_type_ := base_type(value.Type)
	gb_assert_handler("Assertion Failure", "proc_type_.Kind == Type_Proc", "src/llvm_backend_proc.cpp", 4759, 0)
	pt := &proc_type_.Proc

	gb_assert_handler("Assertion Failure", "ce.SplitArgs != nil", "src/llvm_backend_proc.cpp", 4762, 0)

	args := make([]lbValue, 0, pt.ParamCount)

	vari_expand := ce.Ellipsis.Pos.Line != 0
	is_c_vararg := pt.CVarArg

	for i, arg_ast := range ce.SplitArgs.Positional {
		e := pt.Params.Tuple.Variables[i]
		if e.Kind == Entity_TypeName {
			args = append(args, lb_const_nil(p.Module, e.Type))
			continue
		} else if e.Kind == Entity_Constant {
			args = append(args, lb_const_value(p.Module, e.Type, e.Constant.Value))
			continue
		}

		gb_assert_handler("Assertion Failure", "e.Kind == Entity_Variable", "src/llvm_backend_proc.cpp", 4779, 0)

		if pt.Variadic && isize(pt.VariadicIndex) == i {
			variadic_args := lb_const_nil(p.Module, e.Type)
			variadic := ce.SplitArgs.Positional[pt.VariadicIndex:]
			if len(variadic) != 0 {
				slice_type := e.Type
				gb_assert_handler("Assertion Failure", "slice_type.Kind == Type_Slice", "src/llvm_backend_proc.cpp", 4787, 0)

				if is_c_vararg {
					gb_assert_handler("Assertion Failure", "!vari_expand", "src/llvm_backend_proc.cpp", 4790, 0)

					elem_type := slice_type.Slice.Elem

					for _, var_arg := range variadic {
						arg := lb_build_expr(p, var_arg)
						if is_type_any(elem_type) {
							if is_type_untyped_nil(arg.Type) {
								arg = lb_const_nil(p.Module, t_rawptr)
							}
							args = append(args, lb_emit_c_vararg(p, arg, arg.Type))
						} else {
							args = append(args, lb_emit_c_vararg(p, arg, elem_type))
						}
					}
					break
				} else if vari_expand {
					gb_assert_handler("Assertion Failure", "len(variadic) == 1", "src/llvm_backend_proc.cpp", 4807, 0)
					variadic_args = lb_build_expr(p, variadic[0])
					variadic_args = lb_emit_conv(p, variadic_args, slice_type)
				} else {
					elem_type := slice_type.Slice.Elem

					var_args := make([]lbValue, 0, len(variadic))
					for _, var_arg := range variadic {
						v := lb_build_expr(p, var_arg)
						lb_add_values_to_array(p, &var_args, v)
					}
					slice_len := isize(len(var_args))
					if slice_len > 0 {
						var slice lbAddr

						for _, vr := range p.VariadicReuses {
							if are_types_identical(vr.SliceType, slice_type) {
								slice = vr.SliceAddr
								break
							}
						}

						d := decl_info_of_entity(p.Entity)
						if d != nil && slice.Addr.Value == 0 {
							for _, vr := range d.VariadicReuses {
								if are_types_identical(vr.SliceType, slice_type) {
									if len(p.VariadicReuses) > 0 {
										slice = p.VariadicReuses[0].SliceAddr
									} else {
										slice = lb_add_local_generated(p, slice_type, true)
									}
									slice.Addr.Type = alloc_type_pointer(slice_type)
									p.VariadicReuses = append(p.VariadicReuses, lbVariadicReuseSlices{SliceType: slice_type, SliceAddr: slice})
									break
								}
							}
						}

						base_array_ptr := p.VariadicReuseBaseArrayPtr.Addr
						if base_array_ptr.Value == 0 {
							if d != nil {
								max_bytes := d.VariadicReuseMaxBytes
								max_align := d.VariadicReuseMaxAlign
								if max_align < 16 {
									max_align = 16
								}
								p.VariadicReuseBaseArrayPtr = lb_add_local_generated(p, alloc_type_array(t_u8, max_bytes), true)
								lb_try_update_alignment_ptr(p.VariadicReuseBaseArrayPtr.Addr, uint(max_align))
								base_array_ptr = p.VariadicReuseBaseArrayPtr.Addr
							} else {
								base_array_ptr = lb_add_local_generated(p, alloc_type_array(elem_type, slice_len), true).Addr
							}
						}

						if slice.Addr.Value == 0 {
							slice = lb_add_local_generated(p, slice_type, true)
						}

						gb_assert_handler("Assertion Failure", "base_array_ptr.Value != 0", "src/llvm_backend_proc.cpp", 4869, 0)
						gb_assert_handler("Assertion Failure", "slice.Addr.Value != 0", "src/llvm_backend_proc.cpp", 4870, 0)

						base_array_ptr = lb_emit_conv(p, base_array_ptr, alloc_type_pointer(alloc_type_array(elem_type, slice_len)))

						for i := isize(0); i < isize(len(var_args)); i++ {
							addr := lb_emit_array_epi(p, base_array_ptr, i)
							var_arg := var_args[i]
							var_arg = lb_emit_conv(p, var_arg, elem_type)
							lb_emit_store(p, addr, var_arg)
						}

						base_elem := lb_emit_array_epi(p, base_array_ptr, 0)
						len_ := lb_const_int(p.Module, t_int, u64(slice_len))
						lb_fill_slice(p, slice, base_elem, len_)

						variadic_args = lb_addr_load(p, slice)
					}
				}
			}
			args = append(args, variadic_args)

			break
		} else {
			val := lb_build_expr(p, ce.SplitArgs.Positional[i])
			lb_add_values_to_array(p, &args, val)
		}
	}

	if !is_c_vararg {
		new_args := make([]lbValue, pt.ParamCount)
		copy(new_args, args)
		args = new_args
	}

	for _, arg := range ce.SplitArgs.Named {
		fv := &arg.FieldValue
		gb_assert_handler("Assertion Failure", "fv.Field.Kind == Ast_Ident", "src/llvm_backend_proc.cpp", 4904, 0)
		name := fv.Field.Ident.Token.String
		_ = name
		param_index := lookup_procedure_parameter(pt, name)
		gb_assert_handler("Assertion Failure", "param_index >= 0", "src/llvm_backend_proc.cpp", 4908, 0)

		e := pt.Params.Tuple.Variables[param_index]
		if e.Kind == Entity_TypeName {
			val := lb_const_nil(p.Module, e.Type)
			args[param_index] = val
		} else if is_c_vararg && pt.Variadic && isize(pt.VariadicIndex) == param_index {
			gb_assert_handler("Assertion Failure", "param_index == int(pt.ParamCount)-1", "src/llvm_backend_proc.cpp", 4915, 0)
			slice_type := e.Type
			gb_assert_handler("Assertion Failure", "slice_type.Kind == Type_Slice", "src/llvm_backend_proc.cpp", 4917, 0)
			elem_type := slice_type.Slice.Elem

			if fv.Value.Kind == Ast_CompoundLit {
				literal := &fv.Value.CompoundLit
				for _, var_arg := range literal.Elems {
					arg := lb_build_expr(p, var_arg)
					if is_type_any(elem_type) {
						if is_type_untyped_nil(arg.Type) {
							arg = lb_const_nil(p.Module, t_rawptr)
						}
						args = append(args, lb_emit_c_vararg(p, arg, arg.Type))
					} else {
						args = append(args, lb_emit_c_vararg(p, arg, elem_type))
					}
				}
			} else {
				val := lb_build_expr(p, fv.Value)
				gb_assert_handler("Assertion Failure", "!is_type_tuple(val.Type)", "src/llvm_backend_proc.cpp", 4935, 0)
				args = append(args, lb_emit_c_vararg(p, val, val.Type))
			}
		} else {
			val := lb_build_expr(p, fv.Value)
			gb_assert_handler("Assertion Failure", "!is_type_tuple(val.Type)", "src/llvm_backend_proc.cpp", 4940, 0)
			args[param_index] = val
		}
	}

	if pt.Params != nil {
		min_count := isize(len(pt.Params.Tuple.Variables))
		if is_c_vararg {
			min_count -= 1
		}
		gb_assert_handler("Assertion Failure", "isize(len(args)) >= min_count", "src/llvm_backend_proc.cpp", 4951, 0)
		for arg_index, e := range pt.Params.Tuple.Variables {
			if pt.Variadic && isize(pt.VariadicIndex) == arg_index {
				if !is_c_vararg && args[arg_index].Value == 0 {
					args[arg_index] = lb_const_nil(p.Module, e.Type)
				}
				continue
			}

			arg := args[arg_index]
			if arg.Value == 0 && arg.Type == nil {
				switch e.Kind {
				case Entity_TypeName:
					args[arg_index] = lb_const_nil(p.Module, e.Type)
				case Entity_Variable:
					args[arg_index] = lb_handle_param_value(p, e.Type, e.Variable.ParamValue, pt, expr)
				case Entity_Constant:
					args[arg_index] = lb_const_value(p.Module, e.Type, e.Constant.Value)
				default:
					gb_assert_handler("Panic", nil, "src/llvm_backend_proc.cpp", 4975, "Unknown entity kind")
				}
			} else {
				args[arg_index] = lb_emit_conv(p, arg, e.Type)
			}
		}
	}

	final_count := isize(len(args))
	if !is_c_vararg {
		final_count = isize(pt.ParamCount)
	}
	call_args := args[:final_count]

	if is_objc_call {
		return lb_handle_objc_auto_send(p, expr, call_args)
	}

	inlining := ce.Inlining
	tailing := ce.Tailing

	if tailing == ProcTailingNone &&
		proc_entity != nil &&
		proc_entity.Kind == Entity_Procedure &&
		proc_entity.DeclInfo != nil &&
		proc_entity.DeclInfo.ProcLit != nil {
		pl := &proc_entity.DeclInfo.ProcLit.ProcLit

		if pl.Inlining != ProcInliningNone {
			inlining = pl.Inlining
		}

		if pl.Tailing != ProcTailingNone {
			tailing = pl.Tailing
		}
	}

	return lb_emit_call(p, value, call_args, inlining)
}

func lb_build_call_expr(p *lbProcedure, expr *Ast) lbValue {
	expr = unparen_expr(expr)
	ce := &expr.CallExpr

	res := lb_build_call_expr_internal(p, expr)

	if ce.OptionalOkOne {
		gb_assert_handler("Assertion Failure", "is_type_tuple(res.Type)", "src/llvm_backend_proc.cpp", 4665, 0)
		gb_assert_handler("Assertion Failure", "len(res.Type.Tuple.Variables) == 2", "src/llvm_backend_proc.cpp", 4666, 0)
		return lb_emit_struct_ev(p, res, 0)
	}
	return res
}

func lb_add_values_to_array(p *lbProcedure, args *[]lbValue, value lbValue) {
	if is_type_tuple(value.Type) {
		for i := range value.Type.Tuple.Variables {
			sub_value := lb_emit_struct_ev(p, value, i32(i))
			*args = append(*args, sub_value)
		}
	} else {
		*args = append(*args, value)
	}
}
