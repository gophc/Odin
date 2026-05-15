package cmd

import (
	"os"

	"golang.org/x/sys/windows"
)

func check_env() bool {
	debugf("[Section] %s\n", "init check env")
	if build_context.show_more_timings {
		s := "init check env"
		timings_start_section(&global_timings, s)
	}

	if odinRoot, ok := os.LookupEnv("ODIN_ROOT"); ok {
		fi, err := os.Stat(odinRoot)
		if err != nil {
			gb_printf_err("Invalid ODIN_ROOT, directory does not exist, got %s\n", odinRoot)
			return false
		}
		if !fi.IsDir() {
			gb_printf_err("Invalid ODIN_ROOT, expected a directory, got %s\n", odinRoot)
			return false
		}
	}
	return true
}

func init_terminal() {
	debugf("[Section] %s\n", "init terminal")
	if build_context.show_more_timings {
		s := "init terminal"
		timings_start_section(&global_timings, s)
	}

	build_context.has_ansi_terminal_colours = false

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return
	}

	if _, ok := os.LookupEnv("FORCE_COLOR"); ok {
		build_context.has_ansi_terminal_colours = true
		return
	}

	hnd, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if err == nil {
		var mode uint32
		if err := windows.GetConsoleMode(hnd, &mode); err == nil {
			if err := windows.SetConsoleMode(hnd, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err == nil {
				build_context.has_ansi_terminal_colours = true
			}
		}
	}

	if !build_context.has_ansi_terminal_colours {
		if odinTerminal, ok := os.LookupEnv("ODIN_TERMINAL"); ok && odinTerminal != "" {
			ot := odinTerminal
			ansi := "ansi"
			if str_eq_ignore_case(ot, ansi) {
				build_context.has_ansi_terminal_colours = true
			}
		}
	}
}
