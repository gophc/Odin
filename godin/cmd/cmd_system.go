package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var globalThreadPool ThreadPool

func init_global_thread_pool() {
	threadCount := build_context.ThreadCount
	if threadCount < 1 {
		threadCount = 1
	}
	workerCount := threadCount
	thread_pool_init(&globalThreadPool, workerCount, "ThreadPoolWorker")
}

func thread_pool_add_task(proc WorkerTaskProc, data unsafe.Pointer) bool {
	return thread_pool_add_task(&globalThreadPool, proc, data)
}

func thread_pool_wait() {
	thread_pool_wait(&globalThreadPool)
}

func PRINT_PEAK_USAGE() int64 {
	if build_context.ShowMoreTimings {
		var pmc windows.ProcessMemoryCounters
		pmc.CB = uint32(unsafe.Sizeof(pmc))
		err := windows.GetProcessMemoryInfo(windows.CurrentProcess(), &pmc, uint32(unsafe.Sizeof(pmc)))
		if err == nil {
			gb_printf("\n")
			peakMB := float64(pmc.PeakWorkingSetSize) / (1024.0 * 1024.0)
			gb_printf("Peak Memory Size: %.3f MiB\n", peakMB)
			return int64(pmc.PeakWorkingSetSize)
		}
	}
	return 0
}

var debugfMutex sync.Mutex

func debugf(format string, args ...interface{}) {
	if build_context.ShowDebugMessages {
		debugfMutex.Lock()
		defer debugfMutex.Unlock()
		gb_printf_err("[DEBUG] ")
		gb_printf_err_va(format, args...)
	}
}

func system_exec_command_line_app_internal(exitOnErr bool, name string, format string, args ...interface{}) int32 {
	cmdLine := fmt.Sprintf(format, args...)

	if build_context.PrintLinkerFlags {
		cmdLine = strings.TrimLeft(cmdLine, " \t\n\r\f\v")
		if len(cmdLine) > 0 && cmdLine[0] == '"' {
			cmdLine = cmdLine[1:]
			i := 0
			for i < len(cmdLine) {
				if cmdLine[i] == '\\' {
					i += 2
					continue
				}
				if cmdLine[i] == '"' {
					i++
					break
				}
				i++
			}
			cmdLine = cmdLine[i:]
		}
		cmdLine = strings.TrimLeft(cmdLine, " \t\n\r\f\v")
		gb_printf("%s\n", cmdLine)
		return 0
	}

	if build_context.ShowSystemCalls {
		gb_printf_err("[SYSTEM CALL] %s\n", name)
		gb_printf_err("%s\n\n", cmdLine)
	}

	cmd := exec.Command("cmd", "/C", cmdLine)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var exitCode int32
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = int32(exitErr.ExitCode())
		} else {
			gb_printf_err("Failed to execute command:\n\t%s\n", cmdLine)
			exitCode = -1
		}
	}

	if exitOnErr && exitCode != 0 {
		os.Exit(int(exitCode))
	}
	return exitCode
}

func system_exec_command_line_app(name string, format string, args ...interface{}) int32 {
	return system_exec_command_line_app_internal(false, name, format, args...)
}

func system_must_exec_command_line_app(name string, format string, args ...interface{}) {
	system_exec_command_line_app_internal(true, name, format, args...)
}

func system_exec_command_line_app_output(command string, output *string) bool {
	if output == nil {
		return false
	}
	cmd := exec.Command("cmd", "/C", command)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	*output = string(out)
	if build_context.ShowSystemCalls {
		gb_printf_err("[SYSTEM CALL OUTPUT] %s -> %s\n", command, *output)
	}
	return true
}

func setup_args() []string {
	return os.Args
}
