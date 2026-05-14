// Depends on: llvm_c.go (LLVM C API stubs), llvm_backend_hpp_types.go (lbModule, lbProcedure, lbFunctionPassManagerKind, etc.)
// Also depends on: lb_find_procedure_value_from_entity, lb_type, lb_emit_source_code_location_as_global_ptr, lb_type_internal_for_procedures_raw, lb_debug_location_from_token_pos, t_rawptr, ast_token, ast_end_token
// Also depends on: EntityFlag_Require, Entity_Procedure, build_context, map_get, gb_assert_handler, gb_strlen, LLVMAttributeIndex_FunctionIndex, is_arch_wasm etc.
package cmd

func lb_opt_ignore(optimization_level i32) bool {
	return optimization_level < 0
}

func lb_basic_populate_function_pass_manager(fpm LLVMPassManagerRef, optimization_level i32) {
	if lb_opt_ignore(optimization_level) {
		return
	}
}

func lb_populate_function_pass_manager(m *lbModule, fpm LLVMPassManagerRef, ignore_memcpy_pass bool, optimization_level i32) {
	if lb_opt_ignore(optimization_level) {
		return
	}
}

func lb_populate_function_pass_manager_specific(m *lbModule, fpm LLVMPassManagerRef, optimization_level i32) {
	if lb_opt_ignore(optimization_level) {
		return
	}
}

func lb_add_function_simplifcation_passes(mpm LLVMPassManagerRef, optimization_level i32) {
}

func lb_populate_module_pass_manager(target_machine LLVMTargetMachineRef, mpm LLVMPassManagerRef, optimization_level i32) {
	if optimization_level <= 0 && build_context.ODIN_DEBUG {
		return
	}
}

func lb_run_remove_dead_instruction_pass(p *lbProcedure) {
	debug_declare_id := LLVMLookupIntrinsicID("llvm.dbg.declare", 16)
	_ = debug_declare_id

	var removal_count isize
	var pass_count isize
	const max_pass_count isize = 10
	var original_instruction_count isize

	for ; pass_count < max_pass_count; pass_count++ {
		was_dead_instructions := false

		for block := LLVMGetLastBasicBlock(p.Value); block != 0; block = LLVMGetPreviousBasicBlock(block) {
			for instr := LLVMGetLastInstruction(block); instr != 0; {
				if pass_count == 0 {
					original_instruction_count++
				}
				curr_instr := instr
				instr = LLVMGetPreviousInstruction(instr)

				first_use := LLVMGetFirstUse(curr_instr)
				if first_use != 0 {
					continue
				}
				if LLVMTypeOf(curr_instr) == 0 {
					continue
				}

				switch LLVMGetInstructionOpcode(curr_instr) {
				case LLVMAlloca:
					if _, ok := p.TupleFixMap[curr_instr]; ok {
						removal_count++
						LLVMInstructionEraseFromParent(curr_instr)
						was_dead_instructions = true
					}
				case LLVMLoad:
					if LLVMGetVolatile(curr_instr) {
						break
					}
					fallthrough
				case LLVMFNeg, LLVMAdd, LLVMFAdd, LLVMSub, LLVMFSub, LLVMMul, LLVMFMul,
					LLVMUDiv, LLVMSDiv, LLVMFDiv, LLVMURem, LLVMSRem, LLVMFRem,
					LLVMShl, LLVMLShr, LLVMAShr, LLVMAnd, LLVMOr, LLVMXor,
					LLVMGetElementPtr, LLVMTrunc, LLVMZExt, LLVMSExt,
					LLVMFPToUI, LLVMFPToSI, LLVMUIToFP, LLVMSIToFP,
					LLVMFPTrunc, LLVMFPExt, LLVMPtrToInt, LLVMIntToPtr, LLVMBitCast, LLVMAddrSpaceCast,
					LLVMICmp, LLVMFCmp, LLVMSelect,
					LLVMExtractElement, LLVMShuffleVector, LLVMExtractValue:
					removal_count++
					LLVMInstructionEraseFromParent(curr_instr)
					was_dead_instructions = true
				}
			}
		}
		if !was_dead_instructions {
			break
		}
	}
	_ = removal_count
	_ = original_instruction_count
}

func lb_run_instrumentation_pass_insert_call(p *lbProcedure, entity *Entity, dummy_builder LLVMBuilderRef, is_enter bool) LLVMValueRef {
	m := p.Module
	if p.DebugInfo != 0 {
		var pos TokenPos
		if is_enter {
			pos = ast_token(p.Body).Pos
		} else {
			pos = ast_end_token(p.Body).Pos
		}
		LLVMSetCurrentDebugLocation2(dummy_builder, lb_debug_location_from_token_pos(p, pos))
	}

	cc := lb_find_procedure_value_from_entity(m, entity)

	var args [3]LLVMValueRef
	args[0] = LLVMConstPointerCast(p.Value, lb_type(m, t_rawptr))

	if is_arch_wasm() {
		args[1] = LLVMConstPointerNull(lb_type(m, t_rawptr))
	} else {
		var returnaddress_args [1]LLVMValueRef
		returnaddress_args[0] = LLVMConstInt(LLVMInt32TypeInContext(m.Ctx), 0, false)
		intrinsic_name := "llvm.returnaddress"
		id := LLVMLookupIntrinsicID(intrinsic_name, len(intrinsic_name))
		_ = id
		ip := LLVMGetIntrinsicDeclaration(m.Mod, id, nil, 0)
		call_type := LLVMIntrinsicGetType(m.Ctx, id, nil, 0)
		args[1] = LLVMBuildCall2(dummy_builder, call_type, ip, returnaddress_args[:], 1, "")
	}

	var name Token
	if p.Entity != nil {
		name = p.Entity.Token
	}
	args[2] = lb_emit_source_code_location_as_global_ptr(p, name.String, name.Pos).Value

	fnp := lb_type_internal_for_procedures_raw(p.Module, entity.Type)
	return LLVMBuildCall2(dummy_builder, fnp, cc.Value, args[:], 3, "")
}

func lb_run_instrumentation_pass(p *lbProcedure) {
	m := p.Module
	enter := m.Info.InstrumentationEnterEntity
	exit := m.Info.InstrumentationExitEntity
	if enter == nil || exit == nil {
		return
	}
	if !(p.Entity != nil &&
		p.Entity.Kind == Entity_Procedure &&
		p.Entity.Procedure.HasInstrumentation) {
		return
	}

	dummy_builder := LLVMCreateBuilderInContext(m.Ctx)
	defer LLVMDisposeBuilder(dummy_builder)

	entry_bb := p.EntryBlock.Block
	LLVMPositionBuilder(dummy_builder, entry_bb, LLVMGetFirstInstruction(entry_bb))
	lb_run_instrumentation_pass_insert_call(p, enter, dummy_builder, true)
	LLVMRemoveStringAttributeAtIndex(p.Value, LLVMAttributeIndex_FunctionIndex, "instrument-function-entry", uint32(len("instrument-function-entry")-1))

	bb_count := LLVMCountBasicBlocks(p.Value)
	bbs := make([]LLVMBasicBlockRef, bb_count)
	LLVMGetBasicBlocks(p.Value, bbs)

	for i := uint32(0); i < bb_count; i++ {
		bb := bbs[i]
		terminator := LLVMGetBasicBlockTerminator(bb)
		if terminator == 0 || LLVMIsAReturnInst(terminator) == 0 {
			continue
		}
		LLVMPositionBuilderBefore(dummy_builder, terminator)
		lb_run_instrumentation_pass_insert_call(p, exit, dummy_builder, false)
	}

	LLVMRemoveStringAttributeAtIndex(p.Value, LLVMAttributeIndex_FunctionIndex, "instrument-function-exit", uint32(len("instrument-function-exit")-1))
}

func lb_run_function_pass_manager(fpm LLVMPassManagerRef, p *lbProcedure, pass_manager_kind lbFunctionPassManagerKind) {
	if p == nil {
		return
	}
	lb_run_remove_dead_instruction_pass(p)
	lb_run_instrumentation_pass(p)
	switch pass_manager_kind {
	case lbFunctionPassManager_none:
		return
	case lbFunctionPassManager_default, lbFunctionPassManager_default_without_memcpy:
		if build_context.OptimizationLevel < 0 {
			return
		}
	}
	LLVMRunFunctionPassManager(fpm, p.Value)
}

func llvm_delete_function(func_ LLVMValueRef) {
	LLVMDeleteFunction(func_)
}

func lb_append_to_llvm_used_list(m *lbModule, value LLVMValueRef, list_name string) {
	global := LLVMGetNamedGlobal(m.Mod, list_name)
	var constants []LLVMValueRef
	operands := 1

	if global != 0 {
		initializer := LLVMGetInitializer(global)
		_ = initializer
		operands = LLVMGetNumOperands(initializer) + 1
		constants = make([]LLVMValueRef, operands)
		for i := 0; i < operands-1; i++ {
			operand := LLVMGetOperand(initializer, i)
			constants[i] = operand
		}
		LLVMDeleteGlobal(global)
	} else {
		constants = make([]LLVMValueRef, 1)
	}

	int8PtrTy := LLVMPointerType(LLVMInt8TypeInContext(m.Ctx), 0)
	aTy := llvm_array_type(int8PtrTy, uint64(operands))
	constants[operands-1] = LLVMConstBitCast(value, int8PtrTy)
	initializer := LLVMConstArray(int8PtrTy, constants, uint32(operands))
	global = LLVMAddGlobal(m.Mod, aTy, list_name)
	LLVMSetLinkage(global, LLVMAppendingLinkage)
	LLVMSetSection(global, "llvm.metadata")
	LLVMSetInitializer(global, initializer)
}

func lb_append_to_compiler_used(m *lbModule, value LLVMValueRef) {
	lb_append_to_llvm_used_list(m, value, "llvm.compiler.used")
}

func lb_append_to_used(m *lbModule, value LLVMValueRef) {
	lb_append_to_llvm_used_list(m, value, "llvm.used")
}

func lb_run_remove_unused_function_pass(m *lbModule) {
	var removal_count isize
	var pass_count isize
	const max_pass_count isize = 10

	for ; pass_count < max_pass_count; pass_count++ {
		was_dead := false

		for func_ := LLVMGetFirstFunction(m.Mod); func_ != 0; {
			curr_func := func_
			func_ = LLVMGetNextFunction(func_)

			first_use := LLVMGetFirstUse(curr_func)
			if first_use != 0 {
				continue
			}

			var nameLen uintptr
			LLVMGetValueName2(curr_func, &nameLen)
			_ = nameLen

			if LLVMIsDeclaration(curr_func) {
				continue
			}

			linkage := LLVMGetLinkage(curr_func)
			if linkage != LLVMInternalLinkage {
				continue
			}

			found := m.ProcedureValues[curr_func]
			if found != nil && found != 0 {
				e := found
				is_required := (e.Flags & EntityFlag_Require) == EntityFlag_Require
				if is_required {
					lb_append_to_compiler_used(m, curr_func)
					continue
				}
			}

			llvm_delete_function(curr_func)
			was_dead = true
			removal_count++
		}

		if !was_dead {
			break
		}
	}
	_ = removal_count
}

func lb_run_remove_unused_globals_pass(m *lbModule) {
	var removal_count isize
	var pass_count isize
	const max_pass_count isize = 10

	for ; pass_count < max_pass_count; pass_count++ {
		was_dead := false

		for global := LLVMGetFirstGlobal(m.Mod); global != 0; {
			curr_global := global
			global = LLVMGetNextGlobal(global)

			first_use := LLVMGetFirstUse(curr_global)
			if first_use != 0 {
				continue
			}

			var nameLen uintptr
			LLVMGetValueName2(curr_global, &nameLen)

			linkage := LLVMGetLinkage(curr_global)
			if linkage != LLVMInternalLinkage {
				continue
			}

			found := m.ProcedureValues[curr_global]
			if found != nil && found != 0 {
				e := found
				is_required := (e.Flags & EntityFlag_Require) == EntityFlag_Require
				if is_required {
					continue
				}
			}

			LLVMDeleteGlobal(curr_global)
			was_dead = true
			removal_count++
		}

		if !was_dead {
			break
		}
	}
	_ = removal_count
}
