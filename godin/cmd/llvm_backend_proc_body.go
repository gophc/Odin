package cmd

import "sync/atomic"

func lb_start_block(p *lbProcedure, b *lbBlock) {
	gb_assert_handler("Assertion Failure", "b != nil", "src/llvm_backend_proc.cpp", 503, 0)
	if !b.Appended {
		b.Appended = true
		LLVMAppendExistingBasicBlock(p.Value, b.Block)
	}
	LLVMPositionBuilderAtEnd(p.Builder, b.Block)
	p.CurrBlock = b
}

func lb_set_debug_position_to_procedure_begin(p *lbProcedure) {
	if p.DebugInfo == 0 {
		return
	}
	var pos TokenPos
	if p.Body != nil {
		pos = ast_token(p.Body).Pos
	} else if p.TypeExpr != nil {
		pos = ast_token(p.TypeExpr).Pos
	} else if p.Entity != nil {
		pos = p.Entity.Token.Pos
	}
	if pos.FileID != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_token_pos(p, pos))
	}
}

func lb_set_debug_position_to_procedure_end(p *lbProcedure) {
	if p.DebugInfo == 0 {
		return
	}
	var pos TokenPos
	if p.Body != nil {
		pos = ast_end_token(p.Body).Pos
	} else if p.TypeExpr != nil {
		pos = ast_end_token(p.TypeExpr).Pos
	} else if p.Entity != nil {
		pos = p.Entity.Token.Pos
	}
	if pos.FileID != 0 {
		LLVMSetCurrentDebugLocation2(p.Builder, lb_debug_location_from_token_pos(p, pos))
	}
}

func lb_begin_procedure_body(p *lbProcedure) {
	decl := p.Entity.DeclInfo
	if decl != nil {
		for i := range decl.Labels {
			bl := decl.Labels[i]
			bb := lbBranchBlocks{Label: bl.Label}
			p.BranchBlocks = append(p.BranchBlocks, bb)
		}
	}

	p.Builder = LLVMCreateBuilderInContext(p.Module.Ctx)

	p.DeclBlock = lb_create_block(p, "decls", true)
	p.EntryBlock = lb_create_block(p, "entry", true)
	lb_start_block(p, p.EntryBlock)

	map_init(&p.DirectParameters)
	// p.VariadicReuses - allocator handled by Go

	gb_assert_handler("Assertion Failure", "p.Type != nil", "src/llvm_backend_proc.cpp", 566, 0)

	lb_ensure_abi_function_type(p.Module, p)
	if p.Type.Proc.CallingConvention == ProcCC_Odin {
		lb_push_context_onto_stack_from_implicit_parameter(p)
	}
	{
		ft := p.AbiFunctionType

		param_offset := uint(0)

		var return_ptr_value lbValue
		if ft.Ret.Kind == lbArg_Indirect {
			name := "agg.result"
			if ft.MultipleReturnOriginalType != nil &&
				p.Type.Proc.HasNamedResults {
				variables := p.Type.Proc.Results.Tuple.Variables
				e := variables[len(variables)-1]
				if !is_blank_ident(e.Token) {
					name = e.Token.String
				}
			}

			return_ptr_type := reduce_tuple_to_single_type(p.Type.Proc.Results)
			split_returns := ft.MultipleReturnOriginalType != nil
			if split_returns {
				gb_assert_handler("Assertion Failure", "is_type_tuple(return_ptr_type)", "src/llvm_backend_proc.cpp", 594, 0)
				variables := return_ptr_type.Tuple.Variables
				return_ptr_type = variables[len(variables)-1].Type
			}
			ptr_type := alloc_type_pointer(return_ptr_type)
			e := alloc_entity_param(nil, make_token_ident(name), ptr_type, false, false)
			e.Flags |= EntityFlag_NoAlias

			return_ptr_value.Value = LLVMGetParam(p.Value, 0)
			LLVMSetValueName2(return_ptr_value.Value, name, uint(len(name)))
			return_ptr_value.Type = ptr_type
			p.ReturnPtr = lb_addr(return_ptr_value)

			lb_add_entity(p.Module, e, return_ptr_value)

			param_offset += 1
		}

		if p.Type.Proc.Params != nil {
			params := &p.Type.Proc.Params.Tuple

			raw_input_parameters_count := uint(LLVMCountParams(p.Value))
			p.RawInputParameters = make([]LLVMValueRef, raw_input_parameters_count)
			LLVMGetParams(p.Value, &p.RawInputParameters[0])

			is_odin_cc := is_calling_convention_odin(ft.CallingConvention)

			param_index := uint(0)
			for i, e := range params.Variables {
				if e.Kind != Entity_Variable {
					continue
				}

				if e.Flags&EntityFlag_CVarArg != 0 {
					gb_assert_handler("Assertion Failure", "i+1 == len(params.Variables)", "src/llvm_backend_proc.cpp", 629, 0)
					continue
				}

				arg_type := &ft.Args[param_index]

				if arg_type.Kind == lbArg_Ignore {
					dummy := lb_add_local_generated(p, e.Type, false).Addr
					lb_add_entity(p.Module, e, dummy)
				} else if arg_type.Kind == lbArg_Direct {
					if len(e.Token.String) != 0 && !is_blank_ident(e.Token.String) {
						param_type := lb_type(p.Module, e.Type)
						original_value := LLVMGetParam(p.Value, param_offset+param_index)
						value := OdinLLVMBuildTransmute(p, original_value, param_type)

						param := lbValue{Value: value, Type: e.Type}

						map_set(&p.DirectParameters, e, param)

						ptr := lb_address_from_load_or_generate_local(p, param)
						gb_assert_handler("Assertion Failure", "LLVMIsAAllocaInst(ptr.Value) != 0", "src/llvm_backend_proc.cpp", 654, 0)
						lb_add_entity(p.Module, e, ptr)
						lb_add_debug_param_variable(p, ptr.Value, e.Type, e.Token, isize(param_index)+1, p.CurrBlock)
					}
				} else if arg_type.Kind == lbArg_Indirect {
					if len(e.Token.String) != 0 && !is_blank_ident(e.Token.String) {
						sz := type_size_of(e.Type)
						do_callee_copy := false

						if is_odin_cc {
							do_callee_copy = sz <= 16
							if build_context.InternalByValue {
								do_callee_copy = true
							}
						}

						ptr := lbValue{}
						ptr.Value = LLVMGetParam(p.Value, param_offset+param_index)
						ptr.Type = alloc_type_pointer(e.Type)

						if do_callee_copy {
							new_ptr := lb_add_local_generated(p, e.Type, false).Addr
							lb_mem_copy_non_overlapping(p, new_ptr, ptr, lb_const_int(p.Module, t_uint, uint64(sz)))
							ptr = new_ptr
						}

						lb_add_entity(p.Module, e, ptr)
						lb_add_debug_param_variable(p, ptr.Value, e.Type, e.Token, isize(param_index)+1, p.DeclBlock)
					}
				}

				param_index += 1
			}
		}

		if p.Type.Proc.HasNamedResults {
			gb_assert_handler("Assertion Failure", "p.Type.Proc.ResultCount > 0", "src/llvm_backend_proc.cpp", 688, 0)
			results := &p.Type.Proc.Results.Tuple

			for i, e := range results.Variables {
				gb_assert_handler("Assertion Failure", "e.Kind == Entity_Variable", "src/llvm_backend_proc.cpp", 693, 0)

				if e.Token.String != "" {
					gb_assert_handler("Assertion Failure", "!is_blank_ident(e.Token)", "src/llvm_backend_proc.cpp", 696, 0)

					var res lbAddr
					if p.Entity != nil && p.Entity.DeclInfo != nil &&
						atomic.LoadUint32(&p.Entity.DeclInfo.DeferUseChecked) != 0 &&
						p.Entity.DeclInfo.DeferUsed == 0 {

						has_return_ptr := p.ReturnPtr.Addr.Value != 0

						if ft.MultipleReturnOriginalType != nil {
							the_offset := isize(-1)
							if i+1 < len(results.Variables) {
								the_offset = isize(param_offset) + ft.OriginalArgCount + isize(i)
							} else if has_return_ptr {
								gb_assert_handler("Assertion Failure", "i+1 == len(results.Variables)", "src/llvm_backend_proc.cpp", 720, 0)
								the_offset = 0
							}
							if the_offset >= 0 {
								ptr := lbValue{}
								ptr.Value = LLVMGetParam(p.Value, uint(the_offset))
								ptr.Type = alloc_type_pointer(e.Type)
								_ = ptr
							}
						} else if has_return_ptr {
							ptr := p.ReturnPtr.Addr

							if len(results.Variables) > 1 {
								ptr = lb_emit_tuple_ep(p, ptr, i32(i))
							}
							gb_assert_handler("Assertion Failure", "is_type_pointer(ptr.Type)", "src/llvm_backend_proc.cpp", 736, 0)
							gb_assert_handler("Assertion Failure", "are_types_identical(type_deref(ptr.Type), e.Type)", "src/llvm_backend_proc.cpp", 737, 0)

							if ptr.Value != 0 {
								lb_add_entity(p.Module, e, ptr)
								lb_add_debug_local_variable(p, ptr.Value, e.Type, e.Token)

								res = lb_addr(ptr)
							}
						}
					}

					if res.Addr.Type == nil {
						res = lb_add_local(p, e.Type, e)
					}

					if e.Variable.ParamValue.Kind != ParameterValue_Invalid {
						gb_assert_handler("Assertion Failure", "e.Variable.ParamValue.Kind != ParameterValue_Location", "src/llvm_backend_proc.cpp", 763, 0)
						gb_assert_handler("Assertion Failure", "e.Variable.ParamValue.Kind != ParameterValue_Expression", "src/llvm_backend_proc.cpp", 764, 0)
						c := lb_handle_param_value(p, e.Type, e.Variable.ParamValue, nil, nil)
						lb_addr_store(p, res, c)
					}
				}
			}
		}
	}

	lb_set_debug_position_to_procedure_begin(p)
	if p.DebugInfo != 0 {
		if len(p.ContextStack) != 0 {
			prev_block := p.CurrBlock
			p.CurrBlock = p.DeclBlock
			lb_add_debug_context_variable(p, lb_find_or_generate_context_ptr(p))
			p.CurrBlock = prev_block
		}
	}
}

func lb_end_procedure_body(p *lbProcedure) {
	lb_set_debug_position_to_procedure_begin(p)

	LLVMPositionBuilderAtEnd(p.Builder, p.DeclBlock.Block)
	LLVMBuildBr(p.Builder, p.EntryBlock.Block)
	LLVMPositionBuilderAtEnd(p.Builder, p.CurrBlock.Block)

	var instr LLVMValueRef

	if p.Type.Proc.ResultCount == 0 {
		instr = LLVMGetLastInstruction(p.CurrBlock.Block)
		if !lb_is_instr_terminating(instr) {
			lb_emit_defer_stmts(p, lbDeferExit_Return, nil, p.Body)
			lb_set_debug_position_to_procedure_end(p)
			LLVMBuildRetVoid(p.Builder)
		}
	}

	first_block := LLVMGetFirstBasicBlock(p.Value)
	for block := first_block; block != 0; block = LLVMGetNextBasicBlock(block) {
		instr = LLVMGetLastInstruction(block)
		if instr == 0 || !lb_is_instr_terminating(instr) {
			LLVMPositionBuilderAtEnd(p.Builder, block)
			LLVMBuildUnreachable(p.Builder)
		}
	}

	p.CurrBlock = nil
	p.StateFlags = 0

	LLVMDisposeBuilder(p.Builder)
}
