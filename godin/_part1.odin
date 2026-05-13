package odingo3

import "core:fmt"
import "core:os"
import "core:strings"
import "core:mem"
import "core:sync"
import "core:thread"
import "core:unicode"
import "core:runtime"
import "core:path/filepath"
import "core:sys/windows"
import "core:strconv"
import "core:slice"
import "core:sort"
import "base:builtin"
import "base:intrinsics"
import "base:runtime"

// ---------------------------------------------------------------------------
// LLVM-C opaque pointer types
// ---------------------------------------------------------------------------

LLVMMemoryBufferRef :: distinct rawptr
LLVMContextRef :: distinct rawptr
LLVMModuleRef :: distinct rawptr
LLVMTypeRef :: distinct rawptr
LLVMValueRef :: distinct rawptr
LLVMBasicBlockRef :: distinct rawptr
LLVMMetadataRef :: distinct rawptr
LLVMNamedMDNodeRef :: distinct rawptr
LLVMBuilderRef :: distinct rawptr
LLVMDIBuilderRef :: distinct rawptr
LLVMPassManagerRef :: distinct rawptr
LLVMUseRef :: distinct rawptr
LLVMAttributeRef :: distinct rawptr
LLVMDiagnosticInfoRef :: distinct rawptr

// ---------------------------------------------------------------------------
// Build context (defined elsewhere)
// ---------------------------------------------------------------------------

// TODO: full BuildContext struct definition (from common)
BuildContext :: struct {}

build_context: BuildContext

// ---------------------------------------------------------------------------
// Global thread pool
// ---------------------------------------------------------------------------

// TODO: ThreadPool struct definition (from gb_thread.h or equivalent)
ThreadPool :: struct {}

// TODO: WorkerTaskProc type definition
WorkerTaskProc :: #type proc(data: rawptr)

// TODO: BlockingMutex definition (from gb_sync.h or equivalent)
BlockingMutex :: sync.Mutex

#assert(size_of(sync.Mutex) > 0)

global_thread_pool: ThreadPool

init_global_thread_pool :: proc() {
	thread_count := build_context.thread_count
	if thread_count <= 1 {
		thread_count = 1
	}
	worker_count := thread_count
	// TODO: thread_pool_init(&global_thread_pool, worker_count, "ThreadPoolWorker")
	_ = worker_count
}

thread_pool_add_task :: proc(proc_: WorkerTaskProc, data: rawptr) -> bool {
	// TODO: return thread_pool_add_task(&global_thread_pool, proc_, data)
	_ = proc_
	_ = data
	return false
}

thread_pool_wait :: proc() {
	// TODO: thread_pool_wait(&global_thread_pool)
}

// ---------------------------------------------------------------------------
// PRINT_PEAK_USAGE
// ---------------------------------------------------------------------------

PRINT_PEAK_USAGE :: proc() -> i64 {
	if build_context.show_more_timings {
		pmc: windows.PROCESS_MEMORY_COUNTERS
		pmc.cb = size_of(windows.PROCESS_MEMORY_COUNTERS)
		if windows.K32GetProcessMemoryInfo(windows.GetCurrentProcess(), &pmc, u32(size_of(pmc))) {
			fmt.println()
			peak_mib := f64(pmc.PeakWorkingSetSize) / f64(1024 * 1024)
			fmt.printf("Peak Memory Size: %.3f MiB\n", peak_mib)
			return i64(pmc.PeakWorkingSetSize)
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// debugf
// ---------------------------------------------------------------------------

debugf_mutex: sync.Mutex

debugf :: proc(fmt_str: string, args: ..any) {
	if build_context.show_debug_messages {
		sync.mutex_lock(&debugf_mutex)
		defer sync.mutex_unlock(&debugf_mutex)
		fmt.eprintf("[DEBUG] ")
		fmt.eprintf(fmt_str, ..args)
	}
}

// ---------------------------------------------------------------------------
// Timings
// ---------------------------------------------------------------------------

// TODO: full Timings struct definition (from common)
Timings :: struct {}

global_timings: Timings

// ---------------------------------------------------------------------------
// system_exec — Windows process execution helpers
// ---------------------------------------------------------------------------

@(private = "file")
system_exec_command_line_app_internal :: proc(
	exit_on_err: bool,
	name: string,
	cmd_line: string,
) -> i32 {
	if build_context.show_system_calls {
		fmt.printf("[SYSTEM] %s\n", cmd_line)
	}

	wcmd := windows.utf8_to_wstring(cmd_line, context.temp_allocator)

	si: windows.STARTUPINFOW
	pi: windows.PROCESS_INFORMATION
	si.cb = size_of(windows.STARTUPINFOW)

	// TODO: si.hStdOutput = GetStdHandle(STD_OUTPUT_HANDLE) if appropriate

	ok := windows.CreateProcessW(
		nil,
		wcmd,
		nil,
		nil,
		false, // bInheritHandles
		0,     // dwCreationFlags
		nil,
		nil,
		&si,
		&pi,
	)

	if !ok {
		err := windows.GetLastError()
		fmt.eprintf("Failed to execute: %s\n", name)
		fmt.eprintf("Command: %s\n", cmd_line)
		fmt.eprintf("Error code: %d\n", err)
		if exit_on_err {
			os.exit(1)
		}
		return -1
	}

	// Wait for the process to finish
	windows.WaitForSingleObject(pi.hProcess, windows.INFINITE)

	exit_code: u32
	windows.GetExitCodeProcess(pi.hProcess, &exit_code)
	windows.CloseHandle(pi.hProcess)
	windows.CloseHandle(pi.hThread)

	if exit_code != 0 && exit_on_err {
		fmt.eprintf("'%s' failed with exit code %d\n", name, exit_code)
		os.exit(1)
	}

	return i32(exit_code)
}

system_exec_command_line_app :: proc(name: string, args: ..any) -> i32 {
	cmd_line := fmt.tprintf(..args)
	return system_exec_command_line_app_internal(false, name, cmd_line)
}

system_must_exec_command_line_app :: proc(name: string, args: ..any) {
	cmd_line := fmt.tprintf(..args)
	system_exec_command_line_app_internal(true, name, cmd_line)
}

system_exec_command_line_app_output :: proc(command: string) -> (string, bool) {
	if build_context.show_system_calls {
		fmt.printf("[SYSTEM] %s\n", command)
	}

	// Use _popen equivalent via os2 or direct Windows API
	// TODO: implement _popen equivalent for stdout capture
	// For now, return empty
	_ = command
	return "", false
}

// ---------------------------------------------------------------------------
// setup_args — extract and convert command-line arguments
// ---------------------------------------------------------------------------

setup_args :: proc(argc: int, argv: [^]cstring) -> [dynamic]string {
	args := make([dynamic]string, 0, argc)

	when ODIN_OS == .Windows {
		// Get Unicode command line and parse it
		wcmd_line := windows.GetCommandLineW()
		if wcmd_line != nil {
			// Use CommandLineToArgvW to split
			arg_count: i32
			wargs := windows.CommandLineToArgvW(wcmd_line, &arg_count)
			if wargs != nil {
				defer windows.LocalFree(windows.HLOCAL(wargs))
				for i in 0 ..< int(arg_count) {
					wstr := (wargs[i:])[0]
					str, err := windows.wstring_to_utf8(wstr, context.temp_allocator)
					_ = err
					append(&args, str)
				}
				return args
			}
		}
		// Fallback: use argv from main
		for i in 0 ..< argc {
			append(&args, string(argv[i]))
		}
	} else {
		for i in 0 ..< argc {
			append(&args, string(argv[i]))
		}
	}

	return args
}

// ---------------------------------------------------------------------------
// print_usage_line / usage
// ---------------------------------------------------------------------------

print_usage_line :: proc(indent: int, fmt_str: string, args: ..any) {
	for _ in 0 ..< indent {
		fmt.printf(" ")
	}
	fmt.printf(fmt_str, ..args)
	fmt.println()
}

usage :: proc(argv0: string, argv1: string = "") {
	if argv0 == "" {
		argv0 = "odin"
	}

	fmt.printf("Usage: %s <command> [options]\n", argv0)
	fmt.println()
	fmt.println("Commands:")
	fmt.println()

	// TODO: add full usage text (from actual usage() implementation in main.i.cpp)
	print_usage_line(2, "build <file>          Compile a package")
	print_usage_line(2, "run   <file>          Compile and run a package")
	print_usage_line(2, "check <file>          Type-check a package")
	print_usage_line(2, "test  <file>          Run tests")
	print_usage_line(2, "doc   <file>          Generate documentation")
	print_usage_line(2, "version               Print version")
	print_usage_line(2, "report                Print bug report")
	fmt.println()
	fmt.println("For more information, use `odin help <command>` or visit https://odin-lang.org")
	fmt.println()
}
