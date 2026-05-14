package cmd

import (
	"fmt"
	"os"
	"unsafe"
)

type lbLLVMEmitWorker struct {
	TargetMachine   LLVMTargetMachineRef
	CodeGenFileType LLVMCodeGenFileType
	FilepathObj     string
	M               *lbModule
}

type lbLLVMModulePassWorkerData struct {
	M             *lbModule
	TargetMachine LLVMTargetMachineRef
	DoThreading   bool
}

func lb_is_module_empty(m *lbModule) bool {
	if LLVMGetFirstFunction(m.Mod) == 0 && LLVMGetFirstGlobal(m.Mod) == 0 {
		return true
	}
	for fn := LLVMGetFirstFunction(m.Mod); fn != 0; fn = LLVMGetNextFunction(fn) {
		if LLVMGetFirstBasicBlock(fn) != 0 {
			return false
		}
	}
	for g := LLVMGetFirstGlobal(m.Mod); g != 0; g = LLVMGetNextGlobal(g) {
		linkage := LLVMGetLinkage(g)
		if linkage == LLVMExternalLinkage || linkage == LLVMWeakAnyLinkage {
			continue
		}
		if !LLVMIsExternallyInitialized(g) {
			return false
		}
	}
	return true
}

func lb_llvm_emit_worker_proc(data unsafe.Pointer) int {
	wd := (*lbLLVMEmitWorker)(data)
	var llvm_error string
	if build_context.LTOKind != LTO_None {
		if LLVMWriteBitcodeToFile(wd.M.Mod, wd.FilepathObj) != 0 {
			gb_printf_err("Failed to write bitcode file: %s\n", wd.FilepathObj)
			exit_with_errors()
		}
	} else if LLVMTargetMachineEmitToFile(wd.TargetMachine, wd.M.Mod, wd.FilepathObj, wd.CodeGenFileType, &llvm_error) != 0 {
		gb_printf_err("LLVM Error: %s\n", llvm_error)
		exit_with_errors()
	}
	debugf("Generated File: %s\n", wd.FilepathObj)
	return 0
}

func lb_llvm_function_pass_per_function_internal(module *lbModule, p *lbProcedure, pass_manager_kind ...lbFunctionPassManagerKind) {
	pmk := lbFunctionPassManager_default
	if len(pass_manager_kind) > 0 {
		pmk = pass_manager_kind[0]
	}
	pass_manager := module.FunctionPassManagers[pmk]
	lb_run_function_pass_manager(pass_manager, p, pmk)
}

func lb_llvm_function_pass_per_module(data unsafe.Pointer) int {
	m := (*lbModule)(data)
	for i := 0; i < int(lbFunctionPassManager_COUNT); i++ {
		m.FunctionPassManagers[i] = LLVMCreateFunctionPassManagerForModule(m.Mod)
	}
	for i := 0; i < int(lbFunctionPassManager_COUNT); i++ {
		LLVMInitializeFunctionPassManager(m.FunctionPassManagers[i])
	}
	lb_populate_function_pass_manager(m, m.FunctionPassManagers[lbFunctionPassManager_default], false, build_context.OptimizationLevel)
	lb_populate_function_pass_manager(m, m.FunctionPassManagers[lbFunctionPassManager_default_without_memcpy], true, build_context.OptimizationLevel)
	lb_populate_function_pass_manager_specific(m, m.FunctionPassManagers[lbFunctionPassManager_none], -1)
	for i := 0; i < int(lbFunctionPassManager_COUNT); i++ {
		LLVMFinalizeFunctionPassManager(m.FunctionPassManagers[i])
	}

	if m == &m.Gen.DefaultModule {
		lb_llvm_function_pass_per_function_internal(m, m.Gen.StartupRuntime)
		lb_llvm_function_pass_per_function_internal(m, m.Gen.CleanupRuntime)
		lb_llvm_function_pass_per_function_internal(m, m.Gen.ObjCNames)
	}

	m.GeneratedProceduresMutex.Lock()
	for _, p := range m.GeneratedProcedures {
		if p.Body != nil {
			pass_manager_kind := lbFunctionPassManager_default
			if p.Flags&uint32(lbProcedureFlag_WithoutMemcpyPass) != 0 {
				pass_manager_kind = lbFunctionPassManager_default_without_memcpy
				lb_add_attribute_to_proc(p.Module, p.Value, "optnone")
				lb_add_attribute_to_proc(p.Module, p.Value, "noinline")
			} else {
				if p.Entity != nil && p.Entity.Kind == Entity_Procedure {
					switch p.Entity.Procedure.OptimizationMode {
					case ProcedureOptimizationMode_None:
						pass_manager_kind = lbFunctionPassManager_none
					case ProcedureOptimizationMode_FavorSize:
					}
				}
			}
			lb_llvm_function_pass_per_function_internal(m, p, pass_manager_kind)
		}
	}
	m.GeneratedProceduresMutex.Unlock()

	for _, entry := range m.GenProcs {
		if len(entry.Name) >= 5 && entry.Name[:5] == "__$map" {
			lb_llvm_function_pass_per_function_internal(m, entry, lbFunctionPassManager_none)
		} else {
			lb_llvm_function_pass_per_function_internal(m, entry)
		}
	}
	return 0
}

func lb_remove_unused_functions_and_globals(gen *lbGenerator) {
	for _, entry := range gen.Modules {
		lb_run_remove_unused_function_pass(entry)
		lb_run_remove_unused_globals_pass(entry)
	}
}

func lb_llvm_function_passes(gen *lbGenerator, do_threading bool) {
	if do_threading {
		for _, entry := range gen.Modules {
			thread_pool_add_task(lb_llvm_function_pass_per_module, unsafe.Pointer(entry))
		}
		thread_pool_wait()
	} else {
		for _, entry := range gen.Modules {
			lb_llvm_function_pass_per_module(unsafe.Pointer(entry))
		}
	}
}

func cstr_as_go(s *byte) string {
	if s == nil {
		return ""
	}
	p := unsafe.Pointer(s)
	n := 0
	for *(*byte)(unsafe.Add(p, uintptr(n))) != 0 {
		n++
	}
	return unsafe.String(s, n)
}

func lb_llvm_module_pass_worker_proc(data unsafe.Pointer) int {
	wd := (*lbLLVMModulePassWorkerData)(data)

	module_pass_manager := LLVMCreatePassManager()
	lb_populate_module_pass_manager(wd.TargetMachine, module_pass_manager, build_context.OptimizationLevel)
	LLVMRunPassManager(module_pass_manager, wd.M.Mod)

	var passes []string
	switch build_context.OptimizationLevel {
	case -1:
		passes = append(passes, "function(annotation-remarks)")
	case 0:
		if build_context.InternalLLVMNoSROA {
			passes = append(passes, "always-inline")
		} else {
			passes = append(passes, "annotation2metadata",
				"inferattrs",
				"forceattrs",
				"function<eager-inv>(sroa<modify-cfg>,early-cse<>)",
				"always-inline",
				"function<eager-inv>(sroa<modify-cfg>,instsimplify,simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>)")
		}
		passes = append(passes, "function(annotation-remarks)")
	case 1:
		passes = append(passes, `annotation2metadata,
forceattrs,
inferattrs,
function<eager-inv>(
	lower-expect,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;no-switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>,
	sroa<modify-cfg>,
	early-cse<>
),
ipsccp,
called-value-propagation,
globalopt,
function<eager-inv>(
	mem2reg,
	instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>
),
always-inline,
require<globals-aa>,
function(
	invalidate<aa>
),
require<profile-summary>,
cgscc(
	devirt<4>(
		inline,
		function-attrs<skip-non-recursive-function-attrs>,
		function<eager-inv;no-rerun>(
			sroa<modify-cfg>,
			early-cse<memssa>,
			speculative-execution<only-if-divergent-target>,
			jump-threading,
			correlated-propagation,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>,
			instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
			aggressive-instcombine,
			tailcallelim,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>,
			reassociate,
			constraint-elimination,
			loop-mssa(
				loop-instsimplify,
				loop-simplifycfg,
				licm<no-allowspeculation>,
				loop-rotate<header-duplication;no-prepare-for-lto>,
				licm<allowspeculation>,
				simple-loop-unswitch<no-nontrivial;trivial>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-sink-common-insts;speculate-blocks;simplify-cond-branch>,
			instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
			loop(
				loop-idiom,
				indvars,
				extra-simple-loop-unswitch-passes,
				loop-deletion,loop-unroll-full
			),
			sroa<modify-cfg>,
			vector-combine,
			mldst-motion<no-split-footer-bb>,
			gvn<>,
			sccp,
			bdce,
			instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
			jump-threading,
			correlated-propagation,
			adce,
			memcpyopt,
			dse,
			move-auto-init,
			loop-mssa(
				licm<allowspeculation>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>
		),
		function-attrs,
		function(
			require<should-not-run-function-passes>
		)
	)
),
deadargelim,
globalopt,
globaldce,
elim-avail-extern,
rpo-function-attrs,
recompute-globalsaa,
function<eager-inv>(
	float2int,
	lower-constant-intrinsics,
	loop(
		loop-rotate<header-duplication;no-prepare-for-lto>,
		loop-deletion
	),
	loop-distribute,
	inject-tli-mappings,
	loop-vectorize<no-interleave-forced-only;no-vectorize-forced-only;>,
	infer-alignment,
	loop-load-elim,
	instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;forward-switch-cond;switch-range-to-icmp;switch-to-lookup;no-keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
	slp-vectorizer,
	vector-combine,
	instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
	loop-unroll<O2>,
	transform-warning,
	sroa<preserve-cfg>,
	infer-alignment,
	instcombine<max-iterations=1;no-use-loop-info;no-verify-fixpoint>,
	loop-mssa(
		licm<allowspeculation>
	),
	alignment-from-assumptions,
	loop-sink,
	instsimplify,
	div-rem-pairs,
	tailcallelim,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;speculate-unpredictables>
),
globaldce,
constmerge,
cg-profile,
rel-lookup-table-converter,
function(
	annotation-remarks
),
verify`)
	case 2:
		passes = append(passes, `annotation2metadata,
forceattrs,
inferattrs,
function<eager-inv>(
	ee-instrument<>,
	lower-expect,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;no-switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
	sroa<modify-cfg>,
	early-cse<>
),
ipsccp,
called-value-propagation,
globalopt,
function<eager-inv>(
	mem2reg,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>
),
always-inline,
require<globals-aa>,
function(
	invalidate<aa>
),
require<profile-summary>,
cgscc(
	devirt<4>(
		inline,
		function-attrs<skip-non-recursive-function-attrs>,
		function<eager-inv;no-rerun>(
			sroa<modify-cfg>,
			early-cse<memssa>,
			speculative-execution<only-if-divergent-target>,
			jump-threading,
			correlated-propagation,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			aggressive-instcombine,
			libcalls-shrinkwrap,
			tailcallelim,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			reassociate,
			constraint-elimination,
			loop-mssa(
				loop-instsimplify,
				loop-simplifycfg,
				licm<no-allowspeculation>,
				loop-rotate<header-duplication;no-prepare-for-lto>,
				licm<allowspeculation>,
				simple-loop-unswitch<no-nontrivial;trivial>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			loop(
				loop-idiom,
				indvars,
				extra-simple-loop-unswitch-passes,
				loop-deletion,
				loop-unroll-full
			),
			sroa<modify-cfg>,
			vector-combine,
			mldst-motion<no-split-footer-bb>,
			gvn<>,
			sccp,
			bdce,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			jump-threading,
			correlated-propagation,
			adce,
			memcpyopt,
			dse,
			move-auto-init,
			loop-mssa(
				licm<allowspeculation>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>
		),
		function-attrs,
		function(
			require<should-not-run-function-passes>
		)
	)
),
deadargelim,
globalopt,
globaldce,
elim-avail-extern,
rpo-function-attrs,
recompute-globalsaa,
function<eager-inv>(
	float2int,
	lower-constant-intrinsics,
	loop(
		loop-rotate<header-duplication;no-prepare-for-lto>,
		loop-deletion
	),
	loop-distribute,
	inject-tli-mappings,
	loop-vectorize<no-interleave-forced-only;no-vectorize-forced-only;>,
	infer-alignment,
	loop-load-elim,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;forward-switch-cond;switch-range-to-icmp;switch-to-lookup;no-keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
	slp-vectorizer,
	vector-combine,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	loop-unroll<O2>,
	transform-warning,
	sroa<preserve-cfg>,
	infer-alignment,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	loop-mssa(
		licm<allowspeculation>
	),
	alignment-from-assumptions,
	loop-sink,
	instsimplify,
	div-rem-pairs,
	tailcallelim,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;no-hoist-common-insts;hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;speculate-unpredictables>
),
globaldce,
constmerge,
cg-profile,
rel-lookup-table-converter,
function(
	annotation-remarks
),
verify`)
	case 3:
		passes = append(passes, `memprof-remove-attributes,
annotation2metadata,
forceattrs,
inferattrs,
function<eager-inv>(
	ee-instrument<>,
	lower-expect,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;no-switch-range-to-icmp;no-switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
	sroa<modify-cfg>,
	early-cse<>,
	callsite-splitting
),
ipsccp,
called-value-propagation,
globalopt,
function<eager-inv>(
	mem2reg,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>
),
always-inline,
require<globals-aa>,
function(
	invalidate<aa>
),
require<profile-summary>,
cgscc(
	devirt<4>(
		inline,
		function-attrs<skip-non-recursive-function-attrs>,
		argpromotion,
		function<eager-inv;no-rerun>(
			sroa<modify-cfg>,
			early-cse<memssa>,
			speculative-execution<only-if-divergent-target>,
			jump-threading,
			correlated-propagation,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			aggressive-instcombine,
			libcalls-shrinkwrap,
			tailcallelim,
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			reassociate,
			constraint-elimination,
			loop-mssa(
				loop-instsimplify,
				loop-simplifycfg,
				licm<no-allowspeculation>,
				loop-rotate<header-duplication;no-prepare-for-lto>,
				licm<allowspeculation>,
				simple-loop-unswitch<nontrivial;trivial>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;no-switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			loop(
				loop-idiom,
				indvars,
				extra-simple-loop-unswitch-passes,
				loop-deletion,
				loop-unroll-full
			),
			sroa<modify-cfg>,
			vector-combine,
			mldst-motion<no-split-footer-bb>,
			gvn<>,
			sccp,
			bdce,
			instcombine<max-iterations=1;no-verify-fixpoint>,
			jump-threading,
			correlated-propagation,
			adce,
			memcpyopt,
			dse,
			move-auto-init,
			loop-mssa(
				licm<allowspeculation>
			),
			simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;switch-to-arithmetic;no-switch-to-lookup;keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
			instcombine<max-iterations=1;no-verify-fixpoint>
		),
		function-attrs,
		function(
			require<should-not-run-function-passes>
		)
	)
),
deadargelim,
globalopt,
globaldce,
elim-avail-extern,
rpo-function-attrs,
recompute-globalsaa,
function<eager-inv>(
	drop-unnecessary-assumes,
	float2int,
	lower-constant-intrinsics,
	chr,
	loop(
		loop-rotate<header-duplication;no-prepare-for-lto>,
		loop-deletion
	),
	loop-distribute,
	inject-tli-mappings,
	loop-vectorize<no-interleave-forced-only;no-vectorize-forced-only;>,
	drop-unnecessary-assumes,
	infer-alignment,
	loop-load-elim,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	simplifycfg<bonus-inst-threshold=1;forward-switch-cond;switch-range-to-icmp;switch-to-arithmetic;switch-to-lookup;no-keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,
	slp-vectorizer,
	vector-combine,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	loop-unroll<O3>,
	transform-warning,
	sroa<preserve-cfg>,
	infer-alignment,
	instcombine<max-iterations=1;no-verify-fixpoint>,
	loop-mssa(
		licm<allowspeculation>
	),
	alignment-from-assumptions,
	loop-sink,
	instsimplify,
	div-rem-pairs,
	tailcallelim,
	simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;switch-to-arithmetic;no-switch-to-lookup;keep-loops;no-hoist-common-insts;hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;simplify-cond-branch;speculate-unpredictables>
),
alloc-token,
globaldce,
constmerge,
cg-profile,
rel-lookup-table-converter,
function(
	annotation-remarks
),
verify`)
	}

	if build_context.LTOKind == LTO_None {
		if build_context.SanitizerFlags&SanitizerFlag_Address != 0 {
			passes = append(passes, "asan")
		}
		if build_context.SanitizerFlags&SanitizerFlag_Memory != 0 {
			passes = append(passes, "msan")
		}
		if build_context.SanitizerFlags&SanitizerFlag_Thread != 0 {
			passes = append(passes, "tsan")
		}
	}

	passes_str := ""
	for i, p := range passes {
		if i != 0 {
			passes_str += ","
		}
		passes_str += p
	}

	pb_options := LLVMCreatePassBuilderOptions()
	llvm_err := LLVMRunPasses(wd.M.Mod, passes_str, wd.TargetMachine, pb_options)
	if llvm_err != 0 {
		err_ptr := LLVMGetErrorMessage(llvm_err)
		gb_printf_err("LLVM Error:\n%s\n", cstr_as_go(err_ptr))
		LLVMDisposeErrorMessage(err_ptr)
		if build_context.KeepTempFiles {
			filepath_ll := lb_filepath_ll_for_module(wd.M)
			var print_err string
			LLVMPrintModuleToFile(wd.M.Mod, filepath_ll, &print_err)
		}
		exit_with_errors()
		return 1
	}
	LLVMConsumeError(llvm_err)
	LLVMDisposePassBuilderOptions(pb_options)

	if !build_context.InternalIgnoreLLVMVerification {
		if wd.DoThreading {
			thread_pool_add_task(lb_llvm_module_verification_worker_proc, unsafe.Pointer(wd.M))
		} else {
			lb_llvm_module_verification_worker_proc(unsafe.Pointer(wd.M))
		}
	}
	return 0
}

func lb_llvm_module_passes_and_verification(gen *lbGenerator, do_threading bool) {
	if do_threading {
		for _, entry := range gen.Modules {
			m := entry
			wd := &lbLLVMModulePassWorkerData{
				M:             m,
				TargetMachine: m.TargetMachine,
				DoThreading:   true,
			}
			thread_pool_add_task(lb_llvm_module_pass_worker_proc, unsafe.Pointer(wd))
		}
		thread_pool_wait()
	} else {
		for _, entry := range gen.Modules {
			m := entry
			wd := &lbLLVMModulePassWorkerData{
				M:             m,
				TargetMachine: m.TargetMachine,
				DoThreading:   false,
			}
			lb_llvm_module_pass_worker_proc(unsafe.Pointer(wd))
		}
	}
}

func lb_generate_procedures_worker_proc(data unsafe.Pointer) int {
	m := (*lbModule)(data)
	for {
		p := mpsc_dequeue[*lbProcedure](&m.ProceduresToGenerate)
		if p == nil {
			break
		}
		lb_generate_procedure(p.Module, p)
	}
	return 0
}

func lb_generate_procedures(gen *lbGenerator, do_threading bool) {
	if do_threading {
		for _, entry := range gen.Modules {
			thread_pool_add_task(lb_generate_procedures_worker_proc, unsafe.Pointer(entry))
		}
		thread_pool_wait()
	} else {
		for _, entry := range gen.Modules {
			lb_generate_procedures_worker_proc(unsafe.Pointer(entry))
		}
	}
}

func lb_generate_missing_procedures_to_check_worker_proc(data unsafe.Pointer) int {
	m := (*lbModule)(data)
	for {
		p := mpsc_dequeue[*lbProcedure](&m.MissingProceduresToCheck)
		if p == nil {
			break
		}
		if !p.IsDone.Load() {
			debugf("Generate missing procedure: %.*s module %p\n", len(p.Name), p.Name, m)
			lb_generate_procedure(m, p)
		}
		for {
			nested := mpsc_dequeue[*lbProcedure](&m.ProceduresToGenerate)
			if nested == nil {
				break
			}
			mpsc_enqueue(&m.MissingProceduresToCheck, nested)
		}
	}
	return 0
}

func lb_generate_missing_procedures(gen *lbGenerator, do_threading bool) {
	retry_count := 0
retry:
	if do_threading {
		for _, entry := range gen.Modules {
			thread_pool_add_task(lb_generate_missing_procedures_to_check_worker_proc, unsafe.Pointer(entry))
		}
		thread_pool_wait()
	} else {
		for _, entry := range gen.Modules {
			lb_generate_missing_procedures_to_check_worker_proc(unsafe.Pointer(entry))
		}
	}
	for _, entry := range gen.Modules {
		m := entry
		if m.MissingProceduresToCheck.Count != 0 {
			if retry_count > len(gen.Modules) {
				gb_assert_handler("Assertion Failure", "m.missing_procedures_to_check.count == 0", "llvm_backend_module_procgen_filepath.go", 0)
			}
			retry_count++
			goto retry
		}
	}
}

func lb_debug_info_complete_types_and_finalize(gen *lbGenerator) {
	for _, entry := range gen.Modules {
		if entry.DebugBuilder != 0 {
			LLVMDIBuilderFinalize(entry.DebugBuilder)
		}
	}
}

func lb_filepath_ll_for_module(m *lbModule) string {
	basename := goStr(build_context.BuildPaths[BuildPath_Output].Basename)
	name := goStr(build_context.BuildPaths[BuildPath_Output].Name)
	path := basename + "/" + name

	module_name := m.ModuleName
	prefix := "odin_package-"
	if len(module_name) >= len(prefix) && module_name[:len(prefix)] == prefix {
		module_name = module_name[len(prefix):]
	}
	path += module_name
	path += ".ll"
	return path
}

func lb_filepath_obj_for_module(m *lbModule) string {
	basename := goStr(build_context.BuildPaths[BuildPath_Output].Basename)
	name := goStr(build_context.BuildPaths[BuildPath_Output].Name)

	use_temporary_directory := false
	if build_context.UseSeparateModules && build_context.BuildMode == BuildMode_Executable {
		dir := os.TempDir()
		if len(dir) != 0 {
			basename = dir
			use_temporary_directory = true
		}
	}

	path := basename + "/"

	if build_context.UseSeparateModules {
		module_name := m.ModuleName
		prefix := "odin_package"
		if len(module_name) >= len(prefix) && module_name[:len(prefix)] == prefix {
			module_name = module_name[len(prefix):]
		}
		path += module_name
	} else {
		path += name
	}

	if use_temporary_directory {
		path += fmt.Sprintf("-%p", unsafe.Pointer(m))
	}

	var ext string
	if build_context.LTOKind != LTO_None {
		ext = "bc"
	} else if build_context.BuildMode == BuildMode_Assembly {
		ext = "S"
	} else if build_context.BuildMode == BuildMode_Object {
		ext = goStr(build_context.BuildPaths[BuildPath_Output].Ext)
	} else {
		ext = infer_object_extension_from_build_context()
	}
	path += "." + ext
	return path
}

func lb_add_foreign_library_paths(gen *lbGenerator) {
	for _, entry := range gen.Modules {
		m := entry
		for _, e := range m.Info.RequiredForeignImportsThroughForce {
			lb_add_foreign_library_path(m, e)
		}
		if lb_is_module_empty(m) {
			continue
		}
	}
}

func lb_llvm_object_generation(gen *lbGenerator, do_threading bool) bool {
	code_gen_file_type := LLVMCodeGenFileTypeObject
	if build_context.BuildMode == BuildMode_Assembly {
		code_gen_file_type = LLVMCodeGenFileTypeAssemblySource
	}

	if do_threading {
		for _, entry := range gen.Modules {
			m := entry
			if lb_is_module_empty(m) {
				continue
			}
			filepath_obj := lb_filepath_obj_for_module(m)
			gen.OutputObjectPaths = append(gen.OutputObjectPaths, make_string_c(filepath_obj))
			filepath_ll := lb_filepath_ll_for_module(m)
			gen.OutputTempPaths = append(gen.OutputTempPaths, make_string_c(filepath_ll))

			wd := &lbLLVMEmitWorker{
				TargetMachine:   m.TargetMachine,
				CodeGenFileType: code_gen_file_type,
				FilepathObj:     filepath_obj,
				M:               m,
			}
			thread_pool_add_task(lb_llvm_emit_worker_proc, unsafe.Pointer(wd))
		}
		thread_pool_wait()
	} else {
		for _, entry := range gen.Modules {
			m := entry
			if lb_is_module_empty(m) {
				continue
			}
			filepath_obj := lb_filepath_obj_for_module(m)
			gen.OutputObjectPaths = append(gen.OutputObjectPaths, make_string_c(filepath_obj))

			var llvm_error string
			if build_context.LTOKind != LTO_None {
				if LLVMWriteBitcodeToFile(m.Mod, filepath_obj) != 0 {
					gb_printf_err("Failed to write bitcode file: %s\n", filepath_obj)
					exit_with_errors()
					return false
				}
			} else if LLVMTargetMachineEmitToFile(m.TargetMachine, m.Mod, filepath_obj, code_gen_file_type, &llvm_error) != 0 {
				gb_printf_err("LLVM Error: %s\n", llvm_error)
				exit_with_errors()
				return false
			}
			debugf("Generated File: %s\n", filepath_obj)
		}
	}
	return true
}

func lb_create_main_procedure(m *lbModule, startup_runtime *lbProcedure, cleanup_runtime *lbProcedure) *lbProcedure {
	default_function_pass_manager := LLVMCreateFunctionPassManagerForModule(m.Mod)
	lb_populate_function_pass_manager(m, default_function_pass_manager, false, build_context.OptimizationLevel)
	LLVMFinalizeFunctionPassManager(default_function_pass_manager)

	params := alloc_type_tuple()
	results := alloc_type_tuple()

	t_ptr_cstring := alloc_type_pointer(t_cstring)

	call_cleanup := true
	has_args := false
	is_dll_main := false
	main_name := "main"

	if build_context.Metrics.Os == TargetOs_windows && build_context.BuildMode == BuildMode_DynamicLibrary {
		is_dll_main = true
		main_name = "DllMain"
		params.Tuple.Variables = make([]*Entity, 3)
		params.Tuple.Variables[0] = alloc_entity_param(nil, make_token_ident("hinstDLL"), t_rawptr, false, true)
		params.Tuple.Variables[1] = alloc_entity_param(nil, make_token_ident("fdwReason"), t_u32, false, true)
		params.Tuple.Variables[2] = alloc_entity_param(nil, make_token_ident("lpReserved"), t_rawptr, false, true)
		call_cleanup = false
	} else if build_context.Metrics.Os == TargetOs_windows && build_context.NoCRT {
		main_name = "mainCRTStartup"
	} else if build_context.Metrics.Os == TargetOs_windows && build_context.Metrics.Arch == TargetArch_i386 && !build_context.NoCRT {
		main_name = "main"
		has_args = true
		params.Tuple.Variables = make([]*Entity, 2)
		params.Tuple.Variables[0] = alloc_entity_param(nil, make_token_ident("argc"), t_i32, false, true)
		params.Tuple.Variables[1] = alloc_entity_param(nil, make_token_ident("argv"), t_ptr_cstring, false, true)
	} else if is_arch_wasm() {
		main_name = "_start"
		call_cleanup = false
	} else {
		has_args = true
		params.Tuple.Variables = make([]*Entity, 2)
		params.Tuple.Variables[0] = alloc_entity_param(nil, make_token_ident("argc"), t_i32, false, true)
		params.Tuple.Variables[1] = alloc_entity_param(nil, make_token_ident("argv"), t_ptr_cstring, false, true)
	}

	results.Tuple.Variables = make([]*Entity, 1)
	results.Tuple.Variables[0] = alloc_entity_param(nil, blank_token, t_i32, false, true)

	proc_type := alloc_type_proc(nil,
		params, isize(len(params.Tuple.Variables)),
		results, isize(len(results.Tuple.Variables)), false, ProcCC_CDecl)

	p := lb_create_dummy_procedure(m, main_name, proc_type)
	p.IsStartup = true

	lb_begin_procedure_body(p)

	if has_args {
		argc := lbValue{Value: LLVMGetParam(p.Value, 0), Type: t_i32}
		argv := lbValue{Value: LLVMGetParam(p.Value, 1), Type: t_ptr_cstring}
		LLVMSetValueName2(argc.Value, "argc", 4)
		LLVMSetValueName2(argv.Value, "argv", 4)
		argc = lb_emit_conv(p, argc, t_int)
		args := lb_addr(lb_find_runtime_value(p.Module, "args__"))
		lb_fill_slice(p, args, argv, argc)
	}

	startup_runtime_value := lbValue{Value: startup_runtime.Value, Type: startup_runtime.Type}
	lb_emit_call(p, startup_runtime_value, nil)

	if build_context.CommandKind == Command_test {
		t_Internal_Test := find_type_in_pkg(m.Info, make_string_c("testing"), make_string_c("Internal_Test"))
		array_type := alloc_type_array(t_Internal_Test, int64(len(m.Info.TestingProcedures)))
		slice_type := alloc_type_slice(t_Internal_Test)
		all_tests_array_addr := lb_add_global_generated_with_name(p.Module, array_type, lbValue{}, "__$all_tests_array")
		all_tests_array := lb_addr_get_ptr(p, all_tests_array_addr)

		var indices [2]LLVMValueRef
		indices[0] = LLVMConstInt(lb_type(m, t_i32), 0, false)

		testing_proc_index := 0
		for _, testing_proc := range m.Info.TestingProcedures {
			t_name := testing_proc.Token.String
			var pkg_name String
			if testing_proc.Pkg != nil {
				pkg_name = testing_proc.Pkg.Name
			}
			v_pkg := lb_find_or_add_entity_string(m, goStr(pkg_name), false)
			v_name := lb_find_or_add_entity_string(m, goStr(t_name), false)
			v_proc := lb_find_procedure_value_from_entity(m, testing_proc)

			indices[1] = LLVMConstInt(lb_type(m, t_int), uint64(testing_proc_index), false)
			testing_proc_index++

			vals := [3]LLVMValueRef{v_pkg.Value, v_name.Value, v_proc.Value}

			dst := LLVMConstInBoundsGEP2(llvm_addr_type(m, all_tests_array), all_tests_array.Value, indices[:], 2)
			src := llvm_const_named_struct(m, t_Internal_Test, vals[:], 3)

			LLVMBuildStore(p.Builder, src, dst)
		}

		all_tests_slice := lb_add_local_generated(p, slice_type, true)
		lb_fill_slice(p, all_tests_slice,
			lb_array_elem(p, all_tests_array),
			lb_const_int(m, t_int, uint64(len(m.Info.TestingProcedures))))

		runner := lb_find_package_value(m, "testing", "runner")

		args_inner := make([]lbValue, 1)
		args_inner[0] = lb_addr_load(p, all_tests_slice)
		result := lb_emit_call(p, runner, args_inner)

		pkg := get_runtime_package(m.Info)
		exit_name := make_string_c("exit")
		e := scope_lookup_current(pkg.Scope, string_interner_insert(exit_name))
		if e == nil {
			compiler_error("Could not find type declaration for '%.*s.%.*s'\n", len(pkg.Name), pkg.Name, len(goStr(exit_name)), goStr(exit_name))
		}
		exit_runner := lb_find_value_from_entity(m, e)

		exit_args := make([]lbValue, 1)
		exit_args[0] = lb_emit_select(p, result, lb_const_int(m, t_int, 0), lb_const_int(m, t_int, 1))
		lb_emit_call(p, exit_runner, exit_args)
	} else {
		if m.Info.EntryPoint != nil {
			entry_point := lb_find_procedure_value_from_entity(m, m.Info.EntryPoint)
			lb_emit_call(p, entry_point, nil)
		}
		if call_cleanup {
			cleanup_runtime_value := lbValue{Value: cleanup_runtime.Value, Type: cleanup_runtime.Type}
			lb_emit_call(p, cleanup_runtime_value, nil)
		}
		if is_dll_main {
			LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_i32), 1, false))
		} else {
			LLVMBuildRet(p.Builder, LLVMConstInt(lb_type(m, t_i32), 0, false))
		}
	}

	lb_end_procedure_body(p)

	LLVMSetLinkage(p.Value, LLVMExternalLinkage)
	if is_arch_wasm() {
		lb_set_wasm_export_attributes(p.Value, p.Name)
	}

	lb_verify_function(m, p)

	lb_run_function_pass_manager(default_function_pass_manager, p, lbFunctionPassManager_default)
	return p
}

func lb_generate_procedure(m *lbModule, p *lbProcedure) {
	if p.IsDone.Load() {
		return
	}
	if p.Body != nil {
		m.CurrProcedure = p
		lb_begin_procedure_body(p)
		lb_build_stmt(p, p.Body)
		lb_end_procedure_body(p)
		p.IsDone.Store(true)
		m.CurrProcedure = nil
	} else if p.GenerateBody != nil {
		p.GenerateBody(m, p)
	}
	if p.Entity != nil && p.Entity.Kind == Entity_Procedure && p.Entity.Procedure.IsMemcpyLike {
		p.Flags |= uint32(lbProcedureFlag_WithoutMemcpyPass)
	}
	lb_verify_function(m, p, true)

	m.GeneratedProceduresMutex.Lock()
	m.GeneratedProcedures = append(m.GeneratedProcedures, p)
	m.GeneratedProceduresMutex.Unlock()
}

func lb_verify_function(m *lbModule, p *lbProcedure, dump_ll ...bool) {
	dump := len(dump_ll) > 0 && dump_ll[0]
	if build_context.InternalIgnoreLLVMVerification {
		return
	}
	if m.DebugBuilder == 0 && LLVMVerifyFunction(p.Value, LLVMReturnStatusAction) != 0 {
		gb_printf_err("LLVM CODE GEN FAILED FOR PROCEDURE: %.*s\n", len(p.Name), p.Name)
		LLVMDumpValue(p.Value)
		gb_printf_err("\n")
		if dump {
			filepath_ll := lb_filepath_ll_for_module(m)
			var llvm_error string
			if LLVMPrintModuleToFile(m.Mod, filepath_ll, &llvm_error) != 0 {
				gb_printf_err("LLVM Error: %s\n", llvm_error)
			}
		}
		LLVMVerifyFunction(p.Value, LLVMPrintMessageAction)
		exit_with_errors()
	}
}

func lb_llvm_module_verification_worker_proc(data unsafe.Pointer) int {
	if build_context.InternalIgnoreLLVMVerification {
		return 0
	}
	m := (*lbModule)(data)
	var llvm_error string
	if LLVMVerifyModule(m.Mod, LLVMReturnStatusAction, &llvm_error) != 0 {
		gb_printf_err("LLVM Error in module %s:\n%s\n", m.ModuleName, llvm_error)
		if build_context.KeepTempFiles {
			filepath_ll := lb_filepath_ll_for_module(m)
			var print_err string
			if LLVMPrintModuleToFile(m.Mod, filepath_ll, &print_err) != 0 {
				gb_printf_err("LLVM Error: %s\n", print_err)
				exit_with_errors()
				return 1
			}
		}
		exit_with_errors()
		return 1
	}
	return 0
}
