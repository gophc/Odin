package cmd

import (
	"fmt"
	"sort"
	"unsafe"
)

func lb_add_entity_existing(m *lbModule, e *Entity, value lbValue) {
	m.ValuesMutex.Lock()
	m.Values[unsafe.Pointer(e)] = value
	m.ValuesMutex.Unlock()
}

func lb_add_member_existing(m *lbModule, name string, value lbValue) {
	m.Members[name] = value
}

func lb_generate_code(gen *lbGenerator) bool {
	thread_count := build_context.ThreadCount
	if thread_count < 1 {
		thread_count = 1
	}
	worker_count := thread_count - 1

	do_threading := LLVMIsMultithreaded() != 0 && build_context.UseSeparateModules && worker_count > 0

	default_module := &gen.DefaultModule
	info := gen.Info

	switch build_context.Metrics.Arch {
	case TargetArch_amd64, TargetArch_i386:
		LLVMInitializeX86TargetInfo()
		LLVMInitializeX86Target()
		LLVMInitializeX86TargetMC()
		LLVMInitializeX86AsmPrinter()
		LLVMInitializeX86AsmParser()
		LLVMInitializeX86Disassembler()
	case TargetArch_arm64:
		LLVMInitializeAArch64TargetInfo()
		LLVMInitializeAArch64Target()
		LLVMInitializeAArch64TargetMC()
		LLVMInitializeAArch64AsmPrinter()
		LLVMInitializeAArch64AsmParser()
		LLVMInitializeAArch64Disassembler()
	case TargetArch_wasm32, TargetArch_wasm64p32:
		LLVMInitializeWebAssemblyTargetInfo()
		LLVMInitializeWebAssemblyTarget()
		LLVMInitializeWebAssemblyTargetMC()
		LLVMInitializeWebAssemblyAsmPrinter()
		LLVMInitializeWebAssemblyAsmParser()
		LLVMInitializeWebAssemblyDisassembler()
	case TargetArch_riscv64:
		LLVMInitializeRISCVTargetInfo()
		LLVMInitializeRISCVTarget()
		LLVMInitializeRISCVTargetMC()
		LLVMInitializeRISCVAsmPrinter()
		LLVMInitializeRISCVAsmParser()
		LLVMInitializeRISCVDisassembler()
	case TargetArch_arm32:
		LLVMInitializeARMTargetInfo()
		LLVMInitializeARMTarget()
		LLVMInitializeARMTargetMC()
		LLVMInitializeARMAsmPrinter()
		LLVMInitializeARMAsmParser()
		LLVMInitializeARMDisassembler()
	default:
		gb_assert_handler("Panic", "Unimplemented LLVM target initialization", "llvm_backend_codegen.go", 0)
	}

	if build_context.Microarch == "native" {
		LLVMInitializeNativeTarget()
	}

	target_triple := build_context.Metrics.TargetTriplet
	for _, entry := range gen.Modules {
		LLVMSetTarget(entry.Mod, goStr(target_triple))
	}

	var target LLVMTargetRef
	var llvm_error_str string
	target_ok := LLVMGetTargetFromTriple(goStr(target_triple), &target, &llvm_error_str)
	_ = target_ok
	if target == 0 {
		gb_assert_handler("Assertion Failure", "target != nullptr", "llvm_backend_codegen.go", 0)
	}

	code_mode := LLVMCodeModelDefault
	if is_arch_wasm() {
		code_mode = LLVMCodeModelJITDefault
		debugf("LLVM code mode: LLVMCodeModelJITDefault\n")
	} else if is_arch_x86() && build_context.Metrics.Os == TargetOs_freestanding {
		code_mode = LLVMCodeModelKernel
		debugf("LLVM code mode: LLVMCodeModelKernel\n")
	}
	if code_mode == LLVMCodeModelDefault {
		debugf("LLVM code mode: LLVMCodeModelDefault\n")
	}

	llvm_cpu := get_final_microarchitecture()

	llvm_features := ""
	first := true
	it := build_context.TargetFeaturesString
	offset := 0
	for offset < len(it) {
		end := offset
		for end < len(it) && it[end] != ',' {
			end++
		}
		str := it[offset:end]
		if end < len(it) {
			offset = end + 1
		} else {
			offset = len(it)
		}
		if len(str) == 0 {
			continue
		}
		if !first {
			llvm_features += ","
		}
		first = false
		if str[0] != '+' && str[0] != '-' {
			llvm_features += "+"
		}
		llvm_features += str
	}

	debugf("CPU: %s, Features: %s\n", llvm_cpu, llvm_features)

	code_gen_level := LLVMCodeGenLevelNone
	switch build_context.OptimizationLevel {
	default:
		fallthrough
	case 0:
		code_gen_level = LLVMCodeGenLevelNone
	case 1:
		code_gen_level = LLVMCodeGenLevelLess
	case 2:
		code_gen_level = LLVMCodeGenLevelDefault
	case 3:
		code_gen_level = LLVMCodeGenLevelAggressive
	}

	reloc_mode := LLVMRelocDefault
	if build_context.BuildMode == BuildMode_DynamicLibrary {
		reloc_mode = LLVMRelocPIC
	}
	switch build_context.RelocMode {
	case RelocMode_Default:
		if build_context.Metrics.Os == TargetOs_openbsd {
			reloc_mode = LLVMRelocPIC
		}
		if build_context.Metrics.Arch == TargetArch_riscv64 {
			reloc_mode = LLVMRelocPIC
		}
	case RelocMode_Static:
		reloc_mode = LLVMRelocStatic
	case RelocMode_PIC:
		reloc_mode = LLVMRelocPIC
	case RelocMode_DynamicNoPIC:
		reloc_mode = LLVMRelocDynamicNoPic
	}

	for _, entry := range gen.Modules {
		target_machine := LLVMCreateTargetMachine(
			target, goStr(target_triple), llvm_cpu,
			llvm_features,
			code_gen_level,
			reloc_mode,
			code_mode)
		m := entry
		m.TargetMachine = target_machine
		data_layout := LLVMCreateTargetDataLayout(target_machine)
		LLVMSetModuleDataLayout(m.Mod, data_layout)
		LLVMDisposeTargetData(data_layout)

		if build_context.FastISel {
			LLVMSetTargetMachineFastISel(m.TargetMachine, true)
		}
	}

	for _, entry := range gen.Modules {
		m := entry
		if m.DebugBuilder != 0 {
			for _, f := range info.Files {
				res := LLVMDIBuilderCreateFile(m.DebugBuilder,
					f.Filename, uint(len(f.Filename)),
					f.Directory, uint(len(f.Directory)))
				lb_set_llvm_metadata(m, f, res)
			}

			producer := "odin"

			is_optimized := 0
			if build_context.OptimizationLevel > 0 {
				is_optimized = 1
			}

			init_file := info.InitPackage.Files[0]
			if entry_point := info.EntryPoint; entry_point != nil {
				if ident := entry_point.Identifier.Load(); ident != nil {
					if ident.FileId != nil {
						init_file = ident.File()
					}
				}
			}

			split_debug_inlining := 0
			if build_context.BuildMode == BuildMode_Assembly {
				split_debug_inlining = 1
			}
			debug_info_for_profiling := 0

			m.DebugCompileUnit = LLVMDIBuilderCreateCompileUnit(m.DebugBuilder, LLVMDWARFSourceLanguageC99,
				lb_get_llvm_metadata(m, unsafe.Pointer(init_file)),
				producer, uint(len(producer)),
				LLVMBool(is_optimized), "", 0,
				1, "", 0,
				LLVMDWARFEmissionFull,
				0, LLVMBool(split_debug_inlining),
				LLVMBool(debug_info_for_profiling),
				"", 0,
				"", 0)
			if m.DebugCompileUnit == 0 {
				gb_assert_handler("Assertion Failure", "m.debug_compile_unit != nullptr", "llvm_backend_codegen.go", 0)
			}
		}
	}

	if !build_context.NoRTTI {
		m := default_module

		{
			max_type_info_count := len(info.TypeInfoTypesHashMap)
			t := alloc_type_array(t_type_info_ptr, int64(max_type_info_count))

			internal_llvm_type := lb_type(m, t)

			g := LLVMAddGlobal(m.Mod, internal_llvm_type, LB_TYPE_INFO_DATA_NAME)
			LLVMSetInitializer(g, LLVMConstNull(internal_llvm_type))
			if build_context.UseSeparateModules {
				LLVMSetLinkage(g, LLVMExternalLinkage)
			} else {
				LLVMSetLinkage(g, LLVMInternalLinkage)
			}
			LLVMSetGlobalConstant(g, true)

			value := lbValue{}
			value.Value = g
			value.Type = alloc_type_pointer(t)

			lb_global_type_info_data_entity = alloc_entity_variable(nil, make_token_ident(LB_TYPE_INFO_DATA_NAME), t, EntityState_Resolved)
			m.Values[unsafe.Pointer(lb_global_type_info_data_entity)] = value
		}

		{
			count := int64(0)
			offsets_extra := int64(0)

			for _, tt := range m.Info.TypeInfoTypesHashMap {
				if tt.Type == nil {
					continue
				}
				index := lb_type_info_index(m.Info, tt.Type, false)
				if index < 0 {
					continue
				}
				switch tt.Type.Kind {
				case Type_Union:
					count += int64(len(tt.Type.Union.Variants))
				case Type_Struct:
					count += int64(len(tt.Type.Struct.Fields))
				case Type_Tuple:
					count += int64(len(tt.Type.Tuple.Variables))
				case Type_BitField:
					count += int64(len(tt.Type.BitField.Fields))
					offsets_extra += int64(len(tt.Type.BitField.Fields))
				}
			}

			global_type_info_make := func(m *lbModule, name string, elem_type *Type, count int64) lbAddr {
				t := alloc_type_array(elem_type, count)
				ptr_type := alloc_type_pointer(t)
				g := LLVMAddGlobal(m.Mod, lb_type(m, t), name)
				LLVMSetInitializer(g, LLVMConstNull(lb_type(m, t)))
				LLVMSetLinkage(g, LLVMInternalLinkage)
				lb_make_global_private_const(g)
				lb_set_odin_rtti_section(g)
				return lb_addr(lbValue{Value: g, Type: ptr_type})
			}

			lb_global_type_info_member_types = global_type_info_make(m, LB_TYPE_INFO_TYPES_NAME, t_type_info_ptr, count)
			lb_global_type_info_member_names = global_type_info_make(m, LB_TYPE_INFO_NAMES_NAME, t_string, count)
			lb_global_type_info_member_offsets = global_type_info_make(m, LB_TYPE_INFO_OFFSETS_NAME, t_uintptr, count+offsets_extra)
			lb_global_type_info_member_usings = global_type_info_make(m, LB_TYPE_INFO_USINGS_NAME, t_bool, count)
			lb_global_type_info_member_tags = global_type_info_make(m, LB_TYPE_INFO_TAGS_NAME, t_string, count)
		}
	}

	global_variable_max_count := 0
	already_has_entry_point := false

	for _, e := range info.Entities {
		if e.Kind == Entity_Variable {
			global_variable_max_count++
		} else if e.Kind == Entity_Procedure {
			if (e.Scope.Flags&ScopeFlag_Init) != 0 && e.Token.String == "main" {
				gb_assert_handler("Assertion Failure", "e == info->entry_point", "llvm_backend_codegen.go", 0)
			}
			if build_context.CommandKind == Command_test &&
				(e.Procedure.IsExport || e.Procedure.LinkName.Len > 0) {
				link_name := e.Procedure.LinkName
				if e.Pkg.Kind == Package_Runtime {
					ln := goStr(link_name)
					if ln == "main" || ln == "_main" || ln == "DllMain" ||
						ln == "WinMain" || ln == "wWinMain" ||
						ln == "mainCRTStartup" || ln == "_start" {
						already_has_entry_point = true
					}
				}
			}
		}
	}

	global_variables := make([]lbGlobalVariable, 0, global_variable_max_count)

	for _, d := range info.VariableInitOrder {
		e := d.Entity
		if (e.Scope.Flags & ScopeFlag_File) == 0 {
			continue
		}
		if e.MinDepCount.Load() == 0 {
			continue
		}
		decl := decl_info_of_entity(e)
		if decl == nil {
			continue
		}
		if e.Kind != Entity_Variable {
			continue
		}

		is_foreign := e.Variable.IsForeign
		is_export := e.Variable.IsExport

		m := default_module
		e_module := lb_module_of_entity(gen, e, default_module)
		_ = e_module

		name := lb_get_entity_name(m, e)

		var lb_var lbGlobalVariable
		lb_var.Decl = decl

		g := lbValue{}
		g.Type = alloc_type_pointer(e.Type)
		g.Value = LLVMAddGlobal(m.Mod, lb_type(m, e.Type), name)

		if decl.InitExpr != nil {
			tav := type_and_value_of_expr(decl.InitExpr)
			if !is_type_any(e.Type) {
				if tav.Mode != Addressing_Invalid {
					if tav.Value.Kind != ExactValue_Invalid {
						cc := LB_CONST_CONTEXT_DEFAULT
						cc.IsRodata = e.Kind == Entity_Variable && e.Variable.IsRodata
						cc.AllowLocal = false
						cc.LinkSection = e.Variable.LinkSection

						v := tav.Value
						init := lb_const_value(m, e.Type, v, cc)

						LLVMDeleteGlobal(g.Value)
						g.Value = 0
						g.Value = LLVMAddGlobal(m.Mod, LLVMTypeOf(init.Value), name)

						LLVMSetInitializer(g.Value, init.Value)
						lb_var.IsInitialized = true
						if cc.IsRodata {
							LLVMSetGlobalConstant(g.Value, true)
						}
					}
				}
			}
			if !lb_var.IsInitialized && is_type_untyped_nil(tav.Type) {
				lb_var.IsInitialized = true
				if e.Kind == Entity_Variable && e.Variable.IsRodata {
					LLVMSetGlobalConstant(g.Value, true)
				}
			}
		} else if e.Kind == Entity_Variable && e.Variable.IsRodata {
			LLVMSetGlobalConstant(g.Value, true)
		}

		lb_apply_thread_local_model(g.Value, e.Variable.ThreadLocalModel)

		if is_foreign {
			LLVMSetLinkage(g.Value, LLVMExternalLinkage)
			LLVMSetDLLStorageClass(g.Value, LLVMDLLImportStorageClass)
			LLVMSetExternallyInitialized(g.Value, true)
			lb_add_foreign_library_path(m, e.Variable.ForeignLibrary)
		} else if LLVMGetInitializer(g.Value) == 0 {
			LLVMSetInitializer(g.Value, LLVMConstNull(lb_type(m, e.Type)))
		}
		if is_export {
			LLVMSetLinkage(g.Value, LLVMDLLExportLinkage)
			LLVMSetDLLStorageClass(g.Value, LLVMDLLExportStorageClass)
		} else if !is_foreign {
			if build_context.UseSeparateModules {
				LLVMSetLinkage(g.Value, LLVMWeakAnyLinkage)
			} else {
				LLVMSetLinkage(g.Value, LLVMInternalLinkage)
			}
		}
		lb_set_linkage_from_entity_flags(m, g.Value, e.Flags)
		LLVMSetAlignment(g.Value, uint32(type_align_of(e.Type)))

		if e.Variable.LinkSection.Len > 0 {
			LLVMSetSection(g.Value, goStr(e.Variable.LinkSection))
		}
		if e.Flags&EntityFlag_Require != 0 {
			lb_append_to_compiler_used(m, g.Value)
		}

		if m.DebugBuilder != 0 {
			global_name := e.Token.String
			if global_name.Len != 0 && global_name != "_" {
				llvm_file := lb_get_llvm_metadata(m, unsafe.Pointer(e.File))
				llvm_scope := llvm_file

				local_to_unit := 0
				if LLVMGetLinkage(g.Value) == LLVMInternalLinkage {
					local_to_unit = 1
				}

				llvm_expr := LLVMDIBuilderCreateExpression(m.DebugBuilder, nil, 0)
				var llvm_decl LLVMMetadataRef

				align_in_bits := uint32(8 * type_align_of(e.Type))

				global_variable_metadata := LLVMDIBuilderCreateGlobalVariableExpression(
					m.DebugBuilder, llvm_scope,
					goStr(global_name), uint(len(goStr(global_name))),
					"", 0,
					llvm_file, e.Token.Pos.Line,
					lb_debug_type(m, e.Type),
					LLVMBool(local_to_unit),
					0,
					llvm_expr,
					llvm_decl,
					align_in_bits)
			lb_set_llvm_metadata(m, unsafe.Pointer(g.Value), global_variable_metadata)
			LLVMGlobalSetMetadata(g.Value, 0, global_variable_metadata)
			}
		}

		if default_module == m {
			g.Value = LLVMConstPointerCast(g.Value, lb_type(m, alloc_type_pointer(e.Type)))
			lb_var.Var = g
			global_variables = append(global_variables, lb_var)
		} else {
			local_g := lbValue{}
			local_g.Type = alloc_type_pointer(e.Type)
			local_g.Value = LLVMAddGlobal(default_module.Mod, lb_type(default_module, e.Type), name)
			LLVMSetLinkage(local_g.Value, LLVMExternalLinkage)

			lb_var.Var = local_g
			global_variables = append(global_variables, lb_var)

			lb_add_entity_existing(default_module, e, local_g)

		lb_add_entity_existing(m, e, g)
		lb_add_member_existing(m, name, g)
	}

	if build_context.ODIN_DEBUG {
		if build_context.Metrics.Os == TargetOs_windows {
			m := default_module
			mod := m.Mod
			ctx := m.Ctx

			{
				typ := LLVMArrayType(LLVMInt8TypeInContext(ctx), 1)
				global := LLVMAddGlobal(mod, typ, "raddbg_is_attached_byte_marker")
				LLVMSetInitializer(global, LLVMConstNull(typ))
				LLVMSetSection(global, ".raddbg")
			}

			if gen.Info.EntryPoint != nil {
				mangled_name := lb_get_entity_name(m, gen.Info.EntryPoint)
				lb_add_raddbg_string_concat3(m, "entry_point: \"", mangled_name, "\"")
			}
		}
	}

	gen.ObjCNames = lb_create_objc_names(default_module)
	gen.StartupRuntime = lb_create_startup_runtime(default_module, gen.ObjCNames, global_variables)
	gen.CleanupRuntime = lb_create_cleanup_runtime(default_module)

	if build_context.ODIN_DEBUG {
		builtin_pkg := get_core_package(m.Info, make_string_c("builtin"))
		if builtin_pkg != nil {
			for _, e := range builtin_pkg.Scope.Elements {
				lb_add_debug_info_for_global_constant_from_entity(gen, e)
			}
		}
	}

	if len(gen.Modules) <= 1 {
		do_threading = false
	}

	lb_create_global_procedures_and_types(gen, info, do_threading)
	lb_generate_procedures(gen, do_threading)

	if build_context.CommandKind == Command_test && !already_has_entry_point {
		lb_create_main_procedure(default_module, gen.StartupRuntime, gen.CleanupRuntime)
	}

	lb_generate_missing_procedures(gen, do_threading)

	if gen.ObjCNames != nil {
		lb_finalize_objc_names(gen, gen.ObjCNames)
	}

	if build_context.ODIN_DEBUG {
		lb_debug_info_complete_types_and_finalize(gen)

		if build_context.Metrics.Os == TargetOs_windows {
			m := default_module
			mod := m.Mod
			ctx := m.Ctx

			lb_add_raddbg_string_cstr(m, "type_view: {type: \"[]?\",        expr: \"array(data, len)\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"string\",     expr: \"array(data, len)\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"[dynamic]?\", expr: \"rows($, array(data, len), len, cap, allocator)\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"[dynamic;?]?\", expr: \"rows($, array(data, len), len)\"}")

			for i := 1; i <= 16; i++ {
				prefix := ""
				endfix := ""
				major := i
				_ = prefix
				_ = endfix
			}
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[1, ?]?\",  expr: \"columns($.data, $[0])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[2, ?]?\",  expr: \"columns($.data, $[0], $[1])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[3, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[4, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[5, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[6, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[7, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[8, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[9, ?]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[10, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[11, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[12, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[13, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[14, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[15, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13], $[14])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"matrix[16, ?]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13], $[14], $[15])\"}")

			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 1]?\",  expr: \"columns($.data, $[0])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 2]?\",  expr: \"columns($.data, $[0], $[1])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 3]?\",  expr: \"columns($.data, $[0], $[1], $[2])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 4]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 5]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 6]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 7]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 8]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 9]?\",  expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 10]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 11]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 12]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 13]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 14]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 15]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13], $[14])\"}")
			lb_add_raddbg_string_cstr(m, "type_view: {type: \"#row_major matrix[?, 16]?\", expr: \"columns($.data, $[0], $[1], $[2], $[3], $[4], $[5], $[6], $[7], $[8], $[9], $[10], $[11], $[12], $[13], $[14], $[15])\"}")

			for _, type_view := range gen.Info.RaddbgTypeViews {
				if type_view.Type == nil {
					continue
				}
				if type_view.View.Len == 0 {
					continue
				}
				t_str := type_to_canonical_string(temporary_allocator(), type_view.Type)
				s := "type_view: {type: \"" + goStr(t_str) + "\", expr: \"" + goStr(type_view.View) + "\"}"
				lb_add_raddbg_string_cstr(m, s)
			}

			global_name_index := uint32(0)
			for {
				str := mpsc_dequeue[String](&gen.RadDebugSectionStrings)
				if str.Data == nil {
					break
				}
				data := LLVMConstStringInContext(ctx, goStr(str), uint(len(goStr(str))), false)
				typ := LLVMTypeOf(data)
				globalName := fmt.Sprintf("raddbg_data__%d", global_name_index)
				global_name_index++
				global := LLVMAddGlobal(mod, typ, globalName)
				LLVMSetInitializer(global, data)
				LLVMSetAlignment(global, 1)
				LLVMSetSection(global, ".raddbg")
			}
		}
	}

	if do_threading {
		non_empty_module_count := 0
		for _, entry := range gen.Modules {
			if !lb_is_module_empty(entry) {
				non_empty_module_count++
			}
		}
		if non_empty_module_count <= 1 {
			do_threading = false
		}
	}

	lb_add_foreign_library_paths(gen)
	lb_llvm_function_passes(gen, do_threading && !build_context.ODIN_DEBUG)
	lb_remove_unused_functions_and_globals(gen)
	lb_llvm_module_passes_and_verification(gen, do_threading)
	lb_correct_entity_linkage(gen)

	if build_context.BuildDiagnostics {
		lb_do_build_diagnostics(gen)
	}

	if build_context.KeepTempFiles || build_context.BuildMode == BuildMode_LLVM_IR {
		for _, entry := range gen.Modules {
			m := entry
			if lb_is_module_empty(m) {
				continue
			}
			filepath_ll := lb_filepath_ll_for_module(m)
			var print_err string
			if LLVMPrintModuleToFile(m.Mod, filepath_ll, &print_err) != 0 {
				gb_printf_err("LLVM Error: %s\n", print_err)
				exit_with_errors()
				return false
			}
			gen.OutputTempPaths = append(gen.OutputTempPaths, make_string_c(filepath_ll))
		}
		if build_context.BuildMode == BuildMode_LLVM_IR {
			return true
		}
	}

	for _, entry := range gen.Modules {
		m := entry
		if !lb_is_module_empty(m) {
			gen.UsedModuleCount++
		}
	}

	if build_context.IgnoreLLVMBuild {
		gb_printf_err("LLVM object generation has been ignored!\n")
		return false
	}
	if !lb_llvm_object_generation(gen, do_threading) {
		return false
	}

	if build_context.SanitizerFlags&SanitizerFlag_Address != 0 {
		switch build_context.Metrics.Os {
		case TargetOs_windows:
			paths := make([]String, 0, 1)
			path := goStr(build_context.ODIN_ROOT) + "\\bin\\llvm\\windows\\clang_rt.asan-x86_64.lib"
			paths = append(paths, make_string_c(path))
			lib := alloc_entity_library_name(nil, make_token_ident("asan_lib"), nil, paths, make_string_c("asan_lib"))
			gen.ForeignLibraries = append(gen.ForeignLibraries, lib)
		case TargetOs_darwin, TargetOs_linux, TargetOs_freebsd:
			if build_context.ExtraLinkerFlags.Len == 0 {
				build_context.ExtraLinkerFlags = make_string_c("-fsanitize=address")
			} else {
				build_context.ExtraLinkerFlags = make_string_c(goStr(build_context.ExtraLinkerFlags) + " -fsanitize=address")
			}
		}
	}
	if build_context.SanitizerFlags&SanitizerFlag_Memory != 0 {
		switch build_context.Metrics.Os {
		case TargetOs_linux, TargetOs_freebsd:
			if build_context.ExtraLinkerFlags.Len == 0 {
				build_context.ExtraLinkerFlags = make_string_c("-fsanitize=memory")
			} else {
				build_context.ExtraLinkerFlags = make_string_c(goStr(build_context.ExtraLinkerFlags) + " -fsanitize=memory")
			}
		}
	}
	if build_context.SanitizerFlags&SanitizerFlag_Thread != 0 {
		switch build_context.Metrics.Os {
		case TargetOs_darwin, TargetOs_linux, TargetOs_freebsd:
			if build_context.ExtraLinkerFlags.Len == 0 {
				build_context.ExtraLinkerFlags = make_string_c("-fsanitize=thread")
			} else {
				build_context.ExtraLinkerFlags = make_string_c(goStr(build_context.ExtraLinkerFlags) + " -fsanitize=thread")
			}
		}
	}

	sort.Slice(gen.ForeignLibraries, func(i, j int) bool {
		return foreign_library_cmp(gen.ForeignLibraries[i], gen.ForeignLibraries[j]) < 0
	})

	return true
}
