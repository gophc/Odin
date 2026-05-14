package cmd

func get_default_microarchitecture() String {
	default_march := S("generic")
	if build_context.Metrics.Arch == TargetArchAmd64 {
		if build_context.Metrics.Os == TargetOsFreestanding {
			default_march = S("x86-64")
		} else {
			default_march = S("x86-64-v2")
		}
	} else if build_context.Metrics.Arch == TargetArchRiscv64 {
		default_march = S("generic-rv64")
	}
	return default_march
}

func get_final_microarchitecture() String {
	microarch := build_context.Microarch
	if microarch.Len == 0 {
		microarch = get_default_microarchitecture()
	} else if microarch == S("native") {
		microarch = make_string_c(goStr(LLVMGetHostCPUName()))
	}
	return microarch
}

func get_default_features() String {
	bc := &build_context
	if bc.Microarch == S("native") {
		features := make_string_c(goStr(LLVMGetHostCPUFeatures()))
		if bc.TargetFeaturesString.Len > 0 {
			bc.TargetFeaturesString = concatenate3_strings(permanent_allocator(), features, S(","), bc.TargetFeaturesString)
		} else {
			bc.TargetFeaturesString = features
		}
		return features
	}
	off := 0
	for i := 0; i < int(bc.Metrics.Arch); i++ {
		off += target_microarch_counts[i]
	}
	microarch := get_final_microarchitecture()
	if bc.Metrics.Arch == TargetArchRiscv64 {
		if microarch == S("generic-rv64") {
			features := S("64bit,a,c,d,f,m,relax,zicsr,zifencei")
			if bc.TargetFeaturesString.Len > 0 {
				bc.TargetFeaturesString = concatenate3_strings(permanent_allocator(), features, S(","), bc.TargetFeaturesString)
			} else {
				bc.TargetFeaturesString = features
			}
			return features
		}
	}
	for i := off; i < off+target_microarch_counts[bc.Metrics.Arch]; i++ {
		if microarch_features_list[i].Microarch == microarch {
			return microarch_features_list[i].Features
		}
	}
	gb_assert_handler("Panic", 0, "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(114), "unknown microarch: %.*s", microarch.Len, microarch.Data)
	return String{}
}

func lb_add_foreign_library_path(m *lbModule, e *Entity) {
	if e == nil {
		return
	}
	gb_assert_handler("Assertion Failure", "e->kind == Entity_LibraryName", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(122), 0)
	gb_assert_handler("Assertion Failure", "e->flags & EntityFlag_Used", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(123), 0)
	mutex_lock(&m.Gen.ForeignMutex)
	if !ptr_set_update(&m.Gen.ForeignLibrariesSet, e) {
		m.Gen.ForeignLibraries = append(m.Gen.ForeignLibraries, e)
	}
	mutex_unlock(&m.Gen.ForeignMutex)
}

func foreign_library_cmp(a, b *Entity) int {
	if a == b {
		return 0
	}
	gb_assert_handler("Assertion Failure", "x->kind == Entity_LibraryName", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(139), 0)
	gb_assert_handler("Assertion Failure", "y->kind == Entity_LibraryName", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(140), 0)
	if a.LibraryName.PriorityIndex != b.LibraryName.PriorityIndex {
		if a.LibraryName.PriorityIndex < b.LibraryName.PriorityIndex {
			return -1
		}
		return 1
	}
	if a.Pkg != b.Pkg {
		var orderX, orderY isize
		if a.Pkg != nil {
			orderX = isize(a.Pkg.Order)
		}
		if b.Pkg != nil {
			orderY = isize(b.Pkg.Order)
		}
		if orderX != orderY {
			if orderX < orderY {
				return -1
			}
			return 1
		}
	}
	if a.File != b.File {
		var fullpathX, fullpathY String
		if a.File != nil {
			fullpathX = a.File.Fullpath
		}
		if b.File != nil {
			fullpathY = b.File.Fullpath
		}
		fileX := filename_from_path(fullpathX)
		fileY := filename_from_path(fullpathY)
		cmp := string_compare(fileX, fileY)
		if cmp != 0 {
			return cmp
		}
	}
	if a.OrderInSrc != b.OrderInSrc {
		if a.OrderInSrc < b.OrderInSrc {
			return -1
		}
		return 1
	}
	if a.Token.Pos.Offset != b.Token.Pos.Offset {
		if a.Token.Pos.Offset < b.Token.Pos.Offset {
			return -1
		}
		return 1
	}
	return 0
}

func lb_set_entity_from_other_modules_linkage_correctly(other_module *lbModule, e *Entity, name String) {
	if other_module == nil {
		return
	}
	cname := goStr(name)
	mpsc_enqueue(&other_module.Gen.EntitiesToCorrectLinkage, lbEntityCorrection{OtherModule: other_module, E: e, Cname: cname})
}

func lb_correct_entity_linkage(gen *lbGenerator) {
	for {
		var ec lbEntityCorrection
		if !mpsc_dequeue(&gen.EntitiesToCorrectLinkage, &ec) {
			break
		}
		var other_global LLVMValueRef
		if ec.E.Kind == Entity_Variable {
			other_global = LLVMGetNamedGlobal(ec.OtherModule.Mod, ec.Cname)
			if other_global != 0 && (LLVMGetInitializer(other_global) != 0 || LLVMIsExternallyInitialized(other_global) != 0) {
				if build_context.UseSeparateModules {
					LLVMSetLinkage(other_global, LLVMWeakAnyLinkage)
				} else {
					LLVMSetLinkage(other_global, LLVMInternalLinkage)
				}
				if !ec.E.Variable.IsExport && !ec.E.Variable.IsForeign {
					LLVMSetVisibility(other_global, LLVMHiddenVisibility)
				}
			}
		} else if ec.E.Kind == Entity_Procedure {
			other_global = LLVMGetNamedFunction(ec.OtherModule.Mod, ec.Cname)
			if other_global != 0 && LLVMCountBasicBlocks(other_global) != 0 {
				if build_context.UseSeparateModules {
					LLVMSetLinkage(other_global, LLVMWeakAnyLinkage)
				} else {
					LLVMSetLinkage(other_global, LLVMInternalLinkage)
				}
				if !ec.E.Procedure.IsExport && !ec.E.Procedure.IsForeign {
					LLVMSetVisibility(other_global, LLVMHiddenVisibility)
				}
			}
		}
	}
}

func lb_emit_init_context(p *lbProcedure, addr lbAddr) {
	gb_assert_handler("Assertion Failure", "addr.kind == lbAddr_Context", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(209), 0)
	gb_assert_handler("Assertion Failure", "addr.ctx.sel.index.count == 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(210), 0)
	args := make([]lbValue, 1)
	args[0] = addr.Addr
	lb_emit_runtime_call(p, "__init_context", args)
}

func lb_push_context_onto_stack_from_implicit_parameter(p *lbProcedure) *lbContextData {
	pt := base_type(p.Type)
	gb_assert_handler("Assertion Failure", "pt->kind == Type_Proc", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(219), 0)
	gb_assert_handler("Assertion Failure", "pt->Proc.calling_convention == ProcCC_Odin", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(220), 0)
	name := S("__.context_ptr")
	e := alloc_entity_param(nil, make_token_ident(name), t_context_ptr, false, false)
	e.Flags |= EntityFlag_NoAlias
	context_ptr := LLVMGetParam(p.Value, LLVMCountParams(p.Value)-1)
	LLVMSetValueName2(context_ptr, "__.context_ptr", 13)
	context_ptr = LLVMBuildPointerCast(p.Builder, context_ptr, lb_type(p.Module, e.Type), "")
	param := lbValue{Value: context_ptr, Type: e.Type}
	lb_add_entity(p.Module, e, param)
	ctx_addr := lbAddr{}
	ctx_addr.Kind = lbAddr_Context
	ctx_addr.Addr = param
	p.ContextStack = append(p.ContextStack, lbContextData{Ctx: ctx_addr, ScopeIndex: -1, Uses: 1})
	return &p.ContextStack[len(p.ContextStack)-1]
}

func lb_push_context_onto_stack(p *lbProcedure, ctx lbAddr) *lbContextData {
	ctx.Kind = lbAddr_Context
	p.ContextStack = append(p.ContextStack, lbContextData{Ctx: ctx, ScopeIndex: isize(p.ScopeIndex)})
	return &p.ContextStack[len(p.ContextStack)-1]
}

func lb_internal_gen_name_from_type(prefix string, typ *Type) string {
	str := gb_string_make(permanent_allocator(), prefix)
	str = gb_string_appendc(str, "$$")
	ct := temp_canonical_string(typ)
	str = gb_string_append_length(str, ct.Data, ct.Len)
	return goStr(make_string(str, gb_string_length(str)))
}

func lb_equal_proc_generate_body(m *lbModule, p *lbProcedure) {
	type_ := p.InternalGenType
	pt := alloc_type_pointer(type_)
	ptr_type := lb_type(m, pt)
	lb_begin_procedure_body(p)
	LLVMSetLinkage(p.Value, LLVMInternalLinkage)
	lb_add_attribute_to_proc(p, "nounwind")
	x := LLVMGetParam(p.Value, 0)
	y := LLVMGetParam(p.Value, 1)
	x = LLVMBuildPointerCast(p.Builder, x, ptr_type, "")
	y = LLVMBuildPointerCast(p.Builder, y, ptr_type, "")
	lhs := lbValue{Value: x, Type: pt}
	rhs := lbValue{Value: y, Type: pt}
	lb_add_proc_attribute_at_index(p, 1+0, "nonnull")
	lb_add_proc_attribute_at_index(p, 1+1, "nonnull")
	block_same_ptr := lb_create_block(p, "same_ptr")
	block_diff_ptr := lb_create_block(p, "diff_ptr")
	same_ptr := lb_emit_comp(p, TokenCmpEq, lhs, rhs)
	lb_emit_if(p, same_ptr, block_same_ptr, block_diff_ptr)
	lb_start_block(p, block_same_ptr)
	LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 1, false))
	lb_start_block(p, block_diff_ptr)
	if type_.Kind == Type_Struct {
		type_set_offsets(type_)
		block_false := lb_create_block(p, "bfalse")
		for i := isize(0); i < isize(len(type_.Struct.Fields)); i++ {
			next_block := lb_create_block(p, "btrue")
			pleft := lb_emit_struct_ep(p, lhs, i32(i))
			pright := lb_emit_struct_ep(p, rhs, i32(i))
			left := lb_emit_load(p, pleft)
			right := lb_emit_load(p, pright)
			ok := lb_emit_comp(p, TokenCmpEq, left, right)
			lb_emit_if(p, ok, next_block, block_false)
			lb_emit_jump(p, next_block)
			lb_start_block(p, next_block)
		}
		LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 1, false))
		lb_start_block(p, block_false)
		LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 0, false))
	} else if type_.Kind == Type_Union {
		if type_size_of(type_) == 0 {
			LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 1, false))
		} else if is_type_union_maybe_pointer(type_) {
			v := type_.Union.Variants[0]
			pv := alloc_type_pointer(v)
			left := lb_emit_load(p, lb_emit_conv(p, lhs, pv))
			right := lb_emit_load(p, lb_emit_conv(p, rhs, pv))
			ok := lb_emit_comp(p, TokenCmpEq, left, right)
			ok = lb_emit_conv(p, ok, t_bool)
			LLVMBuildRet(p.Builder, ok.Value)
		} else {
			block_false := lb_create_block(p, "bfalse")
			block_switch := lb_create_block(p, "bswitch")
			left_tag := lb_emit_load(p, lb_emit_union_tag_ptr(p, lhs))
			right_tag := lb_emit_load(p, lb_emit_union_tag_ptr(p, rhs))
			tag_eq := lb_emit_comp(p, TokenCmpEq, left_tag, right_tag)
			lb_emit_if(p, tag_eq, block_switch, block_false)
			lb_start_block(p, block_switch)
			variant_count := uint(len(type_.Union.Variants))
			if type_.Union.Kind != UnionType_no_nil {
				variant_count++
			}
			v_switch := LLVMBuildSwitch(p.Builder, left_tag.Value, block_false.Block, variant_count)
			if type_.Union.Kind != UnionType_no_nil {
				case_block := lb_create_block(p, "bcase")
				lb_start_block(p, case_block)
				case_tag := lb_const_int(p.Module, union_tag_type(type_), 0)
				LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 1, false))
				LLVMAddCase(v_switch, case_tag.Value, case_block.Block)
			}
			for _, v := range type_.Union.Variants {
				case_block := lb_create_block(p, "bcase")
				lb_start_block(p, case_block)
				case_tag := lb_const_union_tag(p.Module, type_, v)
				vp := alloc_type_pointer(v)
				left := lb_emit_load(p, lb_emit_conv(p, lhs, vp))
				right := lb_emit_load(p, lb_emit_conv(p, rhs, vp))
				ok := lb_emit_comp(p, TokenCmpEq, left, right)
				ok = lb_emit_conv(p, ok, t_bool)
				LLVMBuildRet(p.Builder, ok.Value)
				LLVMAddCase(v_switch, case_tag.Value, case_block.Block)
			}
			lb_start_block(p, block_false)
			LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_bool), 0, false))
		}
	} else {
		left := lb_emit_load(p, lhs)
		right := lb_emit_load(p, rhs)
		ok := lb_emit_comp(p, TokenCmpEq, left, right)
		ok = lb_emit_conv(p, ok, t_bool)
		LLVMBuildRet(p.Builder, ok.Value)
	}
	lb_end_procedure_body(p)
}

func lb_equal_proc_for_type(m *lbModule, typ *Type) lbValue {
	typ = base_type(typ)
	gb_assert_handler("Assertion Failure", "is_type_comparable(type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(399), 0)
	proc_name := lb_internal_gen_name_from_type("__$equal", typ)
	if found, ok := m.GenProcs[proc_name]; ok {
		gb_assert_handler("Assertion Failure", "p != nil", "G:\\b0pass-win\\Odin\\src\\llvm_backend.cpp", i64(405), 0)
		return lbValue{Value: found.Value, Type: found.Type}
	}
	p := lb_create_dummy_procedure(m, proc_name, t_equal_proc)
	m.GenProcs[proc_name] = p
	p.InternalGenType = typ
	p.GenerateBody = lb_equal_proc_generate_body
	mpsc_enqueue(&m.ProceduresToGenerate, p)
	return lbValue{Value: p.Value, Type: p.Type}
}
