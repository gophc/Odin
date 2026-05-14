package cmd

import (
	"sort"
	"strings"
	"unsafe"
)

type lbDiagParaPolyEntry struct {
	Entity        *Entity
	CanonicalName string
	Count         isize
	TotalCodeSize isize
}

type lbDiagModuleEntry struct {
	M                    *lbModule
	Name                 string
	GlobalInternalCount  isize
	GlobalExternalCount  isize
	ProcInternalCount    isize
	ProcExternalCount    isize
	TotalInstructionCount isize
}

func llvm_mask_iota(m *lbModule, start uint, count uint) LLVMValueRef {
	iota := make([]LLVMValueRef, count)
	for i := uint(0); i < count; i++ {
		iota[i] = lb_const_int(m, t_u32, u64(start+i)).Value
	}
	return LLVMConstVector(iota, count)
}

func llvm_mask_zero(m *lbModule, count uint) LLVMValueRef {
	return LLVMConstNull(LLVMVectorType(lb_type(m, t_u32), count))
}

func llvm_basic_shuffle(p *lbProcedure, vector LLVMValueRef, mask LLVMValueRef) LLVMValueRef {
	return LLVMBuildShuffleVector(p.Builder, vector, LLVMGetUndef(LLVMTypeOf(vector)), mask, "")
}

func llvm_basic_const_shuffle(vector LLVMValueRef, mask LLVMValueRef) LLVMValueRef {
	return LLVMConstShuffleVector(vector, LLVMGetUndef(LLVMTypeOf(vector)), mask)
}

func llvm_vector_broadcast(p *lbProcedure, value LLVMValueRef, count uint) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "count > 0", "llvm_backend_utility_extra.go", 0)
	if LLVMIsConstant(value) != 0 {
		single := LLVMConstVector([]LLVMValueRef{value}, 1)
		if count == 1 {
			return single
		}
		mask := llvm_mask_zero(p.Module, count)
		return llvm_basic_const_shuffle(single, mask)
	}

	single_type := LLVMVectorType(LLVMTypeOf(value), 1)
	single := LLVMBuildBitCast(p.Builder, value, single_type, "")
	if count == 1 {
		return single
	}
	mask := llvm_mask_zero(p.Module, count)
	return llvm_basic_shuffle(p, single, mask)
}

func llvm_vector_shuffle_reduction(p *lbProcedure, value LLVMValueRef, op_code LLVMOpcode) LLVMValueRef {
	original_vector_type := LLVMTypeOf(value)

	gb_assert_handler("Assertion Failure", "LLVMGetTypeKind(original_vector_type) == LLVMVectorTypeKind", "llvm_backend_utility_extra.go", 0)
	len_ := LLVMGetVectorSize(original_vector_type)

	v_zero32 := lb_const_int(p.Module, t_u32, 0).Value
	if len_ == 1 {
		return LLVMBuildExtractElement(p.Builder, value, v_zero32, "")
	}
	gb_assert_handler("Assertion Failure", "(len & (len-1)) == 0", "llvm_backend_utility_extra.go", 0)

	for i := len_; i != 1; i >>= 1 {
		mask_len := i / 2
		lhs_mask := llvm_mask_iota(p.Module, 0, mask_len)
		rhs_mask := llvm_mask_iota(p.Module, mask_len, mask_len)
		gb_assert_handler("Assertion Failure", "LLVMTypeOf(lhs_mask) == LLVMTypeOf(rhs_mask)", "llvm_backend_utility_extra.go", 0)

		lhs := llvm_basic_shuffle(p, value, lhs_mask)
		rhs := llvm_basic_shuffle(p, value, rhs_mask)
		gb_assert_handler("Assertion Failure", "LLVMTypeOf(lhs) == LLVMTypeOf(rhs)", "llvm_backend_utility_extra.go", 0)

		value = LLVMBuildBinOp(p.Builder, op_code, lhs, rhs, "")
	}
	return LLVMBuildExtractElement(p.Builder, value, v_zero32, "")
}

func llvm_vector_expand_to_power_of_two(p *lbProcedure, value LLVMValueRef) LLVMValueRef {
	vector_type := LLVMTypeOf(value)
	len_ := LLVMGetVectorSize(vector_type)
	if len_ == 1 {
		return value
	}
	if (len_ & (len_ - 1)) == 0 {
		return value
	}

	expanded_len := uint(next_pow2(i64(len_)))
	mask := llvm_mask_iota(p.Module, 0, expanded_len)
	return LLVMBuildShuffleVector(p.Builder, value, LLVMConstNull(vector_type), mask, "")
}

func llvm_vector_reduce_add(p *lbProcedure, value LLVMValueRef) LLVMValueRef {
	type_ := LLVMTypeOf(value)
	gb_assert_handler("Assertion Failure", "LLVMGetTypeKind(type_) == LLVMVectorTypeKind", "llvm_backend_utility_extra.go", 0)
	elem := OdinLLVMGetVectorElementType(type_)
	len_ := LLVMGetVectorSize(type_)
	if len_ == 0 {
		return LLVMConstNull(type_)
	}

	var name string
	value_offset := 0
	value_count := 0

	switch LLVMGetTypeKind(elem) {
	case LLVMHalfTypeKind, LLVMFloatTypeKind, LLVMDoubleTypeKind:
		name = "llvm.vector.reduce.fadd"
		value_offset = 0
		value_count = 2
	case LLVMIntegerTypeKind:
		name = "llvm.vector.reduce.add"
		value_offset = 1
		value_count = 1
	default:
		gb_assert_handler("Panic", "invalid vector type", "llvm_backend_utility_extra.go", 0, "%s", LLVMPrintTypeToString(type_))
	}

	if false {
		types := [1]LLVMTypeRef{type_}
		_ = name
		_ = value_offset
		_ = value_count
	}

	op_code := LLVMFAdd
	if LLVMGetTypeKind(elem) == LLVMIntegerTypeKind {
		op_code = LLVMAdd
	}

	len_pow2 := prev_pow2(len_)
	if len_pow2 == len_ {
		return llvm_vector_shuffle_reduction(p, value, op_code)
	} else {
		gb_assert_handler("Assertion Failure", "len_pow2 < len", "llvm_backend_utility_extra.go", 0)
		lower_mask := llvm_mask_iota(p.Module, 0, len_pow2)
		upper_mask := llvm_mask_iota(p.Module, len_pow2, len_-len_pow2)
		lower := llvm_basic_shuffle(p, value, lower_mask)
		upper := llvm_basic_shuffle(p, value, upper_mask)
		upper = llvm_vector_expand_to_power_of_two(p, upper)

		lower_reduced := llvm_vector_shuffle_reduction(p, lower, op_code)
		upper_reduced := llvm_vector_shuffle_reduction(p, upper, op_code)
		gb_assert_handler("Assertion Failure", "LLVMTypeOf(lower_reduced) == LLVMTypeOf(upper_reduced)", "llvm_backend_utility_extra.go", 0)

		return LLVMBuildBinOp(p.Builder, op_code, lower_reduced, upper_reduced, "")
	}
}

func llvm_vector_add(p *lbProcedure, a LLVMValueRef, b LLVMValueRef) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "LLVMTypeOf(a) == LLVMTypeOf(b)", "llvm_backend_utility_extra.go", 0)

	elem := OdinLLVMGetVectorElementType(LLVMTypeOf(a))

	if LLVMGetTypeKind(elem) == LLVMIntegerTypeKind {
		return LLVMBuildAdd(p.Builder, a, b, "")
	}
	return LLVMBuildFAdd(p.Builder, a, b, "")
}

func llvm_vector_mul(p *lbProcedure, a LLVMValueRef, b LLVMValueRef) LLVMValueRef {
	gb_assert_handler("Assertion Failure", "LLVMTypeOf(a) == LLVMTypeOf(b)", "llvm_backend_utility_extra.go", 0)

	elem := OdinLLVMGetVectorElementType(LLVMTypeOf(a))

	if LLVMGetTypeKind(elem) == LLVMIntegerTypeKind {
		return LLVMBuildMul(p.Builder, a, b, "")
	}
	return LLVMBuildFMul(p.Builder, a, b, "")
}

func llvm_vector_dot(p *lbProcedure, a LLVMValueRef, b LLVMValueRef) LLVMValueRef {
	return llvm_vector_reduce_add(p, llvm_vector_mul(p, a, b))
}

func llvm_vector_mul_add(p *lbProcedure, a LLVMValueRef, b LLVMValueRef, c LLVMValueRef) LLVMValueRef {
	t := LLVMTypeOf(a)
	gb_assert_handler("Assertion Failure", "t == LLVMTypeOf(b)", "llvm_backend_utility_extra.go", 0)
	gb_assert_handler("Assertion Failure", "t == LLVMTypeOf(c)", "llvm_backend_utility_extra.go", 0)
	gb_assert_handler("Assertion Failure", "LLVMGetTypeKind(t) == LLVMVectorTypeKind", "llvm_backend_utility_extra.go", 0)

	elem := OdinLLVMGetVectorElementType(t)

	is_possible := false

	switch LLVMGetTypeKind(elem) {
	case LLVMHalfTypeKind:
		is_possible = true
	case LLVMFloatTypeKind, LLVMDoubleTypeKind:
		is_possible = true
	}

	if is_possible {
		name := "llvm.fmuladd"
		types := [1]LLVMTypeRef{t}
		values := [3]LLVMValueRef{a, b, c}
		call := lb_call_intrinsic(p, name, values[:], 3, types[:], 1)
		return call
	} else {
		x := llvm_vector_mul(p, a, b)
		y := llvm_vector_add(p, x, c)
		return y
	}
}

func llvm_get_inline_asm(func_type LLVMTypeRef, str string, clobbers string, has_side_effects bool, is_align_stack bool, dialect LLVMInlineAsmDialect) LLVMValueRef {
	return LLVMGetInlineAsm2(func_type,
		str, uintptr(len(str)),
		clobbers, uintptr(len(clobbers)),
		has_side_effects, is_align_stack,
		dialect,
		false,
	)
}

func lb_set_wasm_procedure_import_attributes(value LLVMValueRef, entity *Entity, import_name string) {
	if !is_arch_wasm() {
		return
	}
	module_name := "env"
	if entity.Procedure.ForeignLibrary != nil {
		foreign_library := entity.Procedure.ForeignLibrary
		gb_assert_handler("Assertion Failure", "foreign_library->kind == Entity_LibraryName", "llvm_backend_utility_extra.go", 0)
		gb_assert_handler("Assertion Failure", "len(foreign_library->LibraryName.paths) == 1", "llvm_backend_utility_extra.go", 0)

		module_name = foreign_library.LibraryName.Paths[0]

		if strings.HasSuffix(module_name, ".o") {
			return
		}

		if strings.HasPrefix(import_name, module_name) {
			import_name = import_name[len(module_name)+len(".."):]
		}
	}
	LLVMAddTargetDependentFunctionAttr(value, "wasm-import-module", module_name)
	LLVMAddTargetDependentFunctionAttr(value, "wasm-import-name", import_name)
}

func lb_set_wasm_export_attributes(value LLVMValueRef, export_name string) {
	if !is_arch_wasm() {
		return
	}
	LLVMSetLinkage(value, LLVMDLLExportLinkage)
	LLVMSetDLLStorageClass(value, LLVMDLLExportStorageClass)
	LLVMSetVisibility(value, LLVMDefaultVisibility)
	LLVMAddTargetDependentFunctionAttr(value, "wasm-export-name", export_name)
}

func lb_total_code_size(p *lbProcedure) isize {
	instruction_count := isize(0)

	first := LLVMGetFirstBasicBlock(p.Value)
	for block := first; block != 0; block = LLVMGetNextBasicBlock(block) {
		for instr := LLVMGetFirstInstruction(block); instr != 0; instr = LLVMGetNextInstruction(instr) {
			instruction_count += 1
		}
	}
	return instruction_count
}

func lb_do_para_poly_diagnostics(gen *lbGenerator) {
	procs := make(PtrMap[uintptr, lbDiagParaPolyEntry])

	for entry := range PtrMapIterate(&gen.Modules) {
		m := entry.Value
		for _, p := range m.GeneratedProcedures {
			e := p.Entity
			if e == nil {
				continue
			}
			if p.Builder == 0 {
				continue
			}

			d := e.DeclInfo
			para_poly_parent := d.ParaPolyOriginal
			if para_poly_parent == nil {
				continue
			}

			ptr := uintptr(unsafe.Pointer(para_poly_parent))
			ep, exists := procs[ptr]
			if !exists {
				ep = lbDiagParaPolyEntry{}
				ep.Entity = para_poly_parent
				ep.Count = 0

				w := string_canonical_entity_name(permanent_allocator(), ep.Entity)
				name := w

			for i := 0; i < len(name); i++ {
				s := name[i:]
				if strings.HasPrefix(s, ":proc") {
					name = name[:i]
					break
				}
			}

				ep.CanonicalName = name
				procs[ptr] = ep
			}
			ep = procs[ptr]
			ep.Count += 1
			ep.TotalCodeSize += lb_total_code_size(p)
			procs[ptr] = ep
		}
	}

	entries := make([]lbDiagParaPolyEntry, 0, len(procs))
	for _, ep := range procs {
		entries = append(entries, ep)
	}

	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.TotalCodeSize != b.TotalCodeSize {
			return a.TotalCodeSize > b.TotalCodeSize
		}
		return a.CanonicalName < b.CanonicalName
	})

	gb_printf("Parametric Polymorphic Procedure Diagnostics\n")
	gb_printf("------------------------------------------------------------------------------------------\n")

	gb_printf("Sorted by Total Instruction Count Descending (Top 100)\n\n")
	gb_printf("Total Instruction Count | Instantiation Count | Average Instruction Count | Procedure Name\n")

	max_count := isize(100)
	for _, ep := range entries {
		code_size := ep.TotalCodeSize
		count := ep.Count
		name := ep.CanonicalName

		max_count1 := count
		if max_count1 < 1 {
			max_count1 = 1
		}
		average := f64(code_size) / f64(max_count1)

		gb_printf("%23d | %19d | %25.2f | %s\n", code_size, count, average, name)
		max_count--
		if max_count <= 0 {
			break
		}
	}

	gb_printf("------------------------------------------------------------------------------------------\n")

	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.CanonicalName < b.CanonicalName
	})

	gb_printf("Sorted by Total Instantiation Count Descending (Top 100)\n\n")
	gb_printf("Instantiation Count | Total Instruction Count | Average Instruction Count | Procedure Name\n")

	max_count = 100
	for _, ep := range entries {
		code_size := ep.TotalCodeSize
		count := ep.Count
		name := ep.CanonicalName

		max_count2 := count
		if max_count2 < 1 {
			max_count2 = 1
		}
		average := f64(code_size) / f64(max_count2)

		gb_printf("%19d | %23d | %25.2f | %s\n", count, code_size, average, name)
		max_count--
		if max_count <= 0 {
			break
		}
	}

	gb_printf("------------------------------------------------------------------------------------------\n")

	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Count != b.Count {
			return a.Count < b.Count
		}
		if a.TotalCodeSize != b.TotalCodeSize {
			return a.TotalCodeSize > b.TotalCodeSize
		}
		return a.CanonicalName < b.CanonicalName
	})

	gb_printf("Single Instanced Parametric Polymorphic Procedures\n\n")
	gb_printf("Instruction Count | Procedure Name\n")
	for _, ep := range entries {
		code_size := ep.TotalCodeSize
		count := ep.Count
		name := ep.CanonicalName
		if count != 1 {
			break
		}

		gb_printf("%17d | %s\n", code_size, name)
	}
}

func lb_do_module_diagnostics(gen *lbGenerator) {
	modules := make([]lbDiagModuleEntry, 0)

	for entry := range PtrMapIterate(&gen.Modules) {
		m := entry.Value

		var mod_entry lbDiagModuleEntry
		mod_entry.M = m
		mod_entry.Name = m.ModuleName

		for p := LLVMGetFirstFunction(m.Mod); p != 0; p = LLVMGetNextFunction(p) {
			block := LLVMGetFirstBasicBlock(p)
			if block == 0 {
				mod_entry.ProcExternalCount += 1
			} else {
				mod_entry.ProcInternalCount += 1

				for b := block; b != 0; b = LLVMGetNextBasicBlock(b) {
					for i := LLVMGetFirstInstruction(b); i != 0; i = LLVMGetNextInstruction(i) {
						mod_entry.TotalInstructionCount += 1
					}
				}
			}
		}

		for g := LLVMGetFirstGlobal(m.Mod); g != 0; g = LLVMGetNextGlobal(g) {
			linkage := LLVMGetLinkage(g)
			if linkage == LLVMExternalLinkage {
				mod_entry.GlobalExternalCount += 1
			} else {
				mod_entry.GlobalInternalCount += 1
			}
		}

		modules = append(modules, mod_entry)
	}

	sort.Slice(modules, func(i, j int) bool {
		a, b := modules[i], modules[j]
		if a.TotalInstructionCount != b.TotalInstructionCount {
			return a.TotalInstructionCount > b.TotalInstructionCount
		}
		return a.Name < b.Name
	})

	gb_printf("Module Diagnostics\n\n")
	gb_printf("Total Instructions | Global Internals | Global Externals | Proc Internals | Proc Externals | Files | Instructions/File | Instructions/Proc | Module Name\n")
	gb_printf("-------------------+------------------+------------------+----------------+----------------+-------+-------------------+-------------------+------------\n")
	for _, mod_entry := range modules {
		file_count := isize(1)
		if mod_entry.M.File != nil {
			file_count = 1
		} else if mod_entry.M.Pkg != nil {
			file_count = isize(len(mod_entry.M.Pkg.Files))
		}

		max_files := f64(file_count)
		if max_files < 1.0 {
			max_files = 1.0
		}
		max_procs := f64(mod_entry.ProcInternalCount)
		if max_procs < 1.0 {
			max_procs = 1.0
		}
		instructions_per_file := f64(mod_entry.TotalInstructionCount) / max_files
		instructions_per_proc := f64(mod_entry.TotalInstructionCount) / max_procs

		gb_printf("%18d | %16d | %16d | %14d | %14d | %5d | %17.1f | %17.1f | %s\n",
			mod_entry.TotalInstructionCount,
			mod_entry.GlobalInternalCount,
			mod_entry.GlobalExternalCount,
			mod_entry.ProcInternalCount,
			mod_entry.ProcExternalCount,
			file_count,
			instructions_per_file,
			instructions_per_proc,
			mod_entry.M.ModuleName)
	}
}

func lb_do_build_diagnostics(gen *lbGenerator) {
	lb_do_para_poly_diagnostics(gen)
	gb_printf("------------------------------------------------------------------------------------------\n")
	gb_printf("------------------------------------------------------------------------------------------\n\n")
	lb_do_module_diagnostics(gen)
	gb_printf("------------------------------------------------------------------------------------------\n")
	gb_printf("------------------------------------------------------------------------------------------\n\n")
}
