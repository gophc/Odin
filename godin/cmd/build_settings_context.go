package cmd

func IS_ODIN_DEBUG() bool {
	return buildContext.ODINDEBUG
}

func global_warnings_as_errors() bool {
	return buildContext.WarningsAsErrors
}

func global_ignore_warnings() bool {
	return buildContext.IgnoreWarnings
}

func MAX_ERROR_COLLECTOR_COUNT() isize {
	if buildContext.MaxErrorCount <= 0 {
		return 36
	}
	return buildContext.MaxErrorCount
}

func show_error_line() bool {
	return !buildContext.HideErrorLine && !buildContext.JSONErrors
}

func terse_errors() bool {
	return buildContext.TerseErrors
}

func json_errors() bool {
	return buildContext.JSONErrors
}

func has_ansi_terminal_colours() bool {
	return buildContext.HasANSITerminalColours && !json_errors()
}

func init_build_context(crossTarget *TargetMetrics, subtarget Subtarget) {
	bc := &buildContext

	gb_affinity_init(&bc.Affinity)
	if bc.ThreadCount == 0 {
		bc.ThreadCount = max(bc.Affinity.ThreadCount, 1)
	}

	bc.ODINVENDOR = "odin"
	bc.ODINVERSION = ODIN_VERSION
	bc.ODINROOT = odin_root_dir()
	if bc.MaxErrorCount <= 0 {
		bc.MaxErrorCount = 36
	}

	{
		found := gb_get_env("ODIN_ERROR_POS_STYLE", permanent_allocator())
		if found != "" {
			kind := ErrorPosStyleDefault
			style := string_trim_whitespace(make_string_c(found))
			if style == "" || style == "default" || style == "odin" {
				kind = ErrorPosStyleDefault
			} else if style == "unix" || style == "gcc" || style == "clang" || style == "llvm" {
				kind = ErrorPosStyleUnix
			} else {
				gb_printf_err("Invalid ODIN_ERROR_POS_STYLE: got %s\n", style)
				gb_printf_err("Valid formats:\n")
				gb_printf_err("\t\"default\" or \"odin\"\n")
				gb_printf_err("\t\tpath(line:column) message\n")
				gb_printf_err("\t\"unix\"\n")
				gb_printf_err("\t\tpath:line:column: message\n")
				gb_exit(1)
			}
			bc.ODINERRORPOSSTYLE = kind
		}
	}

	bc.CopyFileContents = true

	var metrics *TargetMetrics
	metrics = &target_windows_amd64

	if crossTarget != nil && metrics != crossTarget {
		bc.DifferentOS = crossTarget.Os != metrics.Os
		bc.CrossCompiling = true
		metrics = crossTarget
	}

	if metrics.Os == TargetOsInvalid {
		panic("metrics->os != TargetOs_Invalid")
	}
	if metrics.Arch == TargetArchInvalid {
		panic("metrics->arch != TargetArch_Invalid")
	}
	if metrics.PtrSize <= 1 {
		panic("metrics->ptr_size > 1")
	}
	if metrics.IntSize <= 1 {
		panic("metrics->int_size > 1")
	}
	if metrics.MaxAlign <= 1 {
		panic("metrics->max_align > 1")
	}
	if metrics.MaxSimdAlign <= 1 {
		panic("metrics->max_simd_align > 1")
	}
	if metrics.IntSize < metrics.PtrSize {
		panic("metrics->int_size >= metrics->ptr_size")
	}
	if metrics.IntSize > metrics.PtrSize {
		if metrics.IntSize != 2*metrics.PtrSize {
			panic("metrics->int_size == 2*metrics->ptr_size")
		}
	}

	bc.Metrics = *metrics
	bc.ODINOS = target_os_names[metrics.Os]
	bc.ODINARCH = target_arch_names[metrics.Arch]
	bc.EndianKind = target_endians[metrics.Arch]
	bc.PtrSize = int64(metrics.PtrSize)
	bc.IntSize = int64(metrics.IntSize)
	bc.MaxAlign = int64(metrics.MaxAlign)
	bc.MaxSimdAlign = int64(metrics.MaxSimdAlign)

	bc.LinkFlags = " "

	if bc.DisableRedZone {
		if is_arch_wasm() && bc.Metrics.Os == TargetOsFreestanding {
			gb_printf_err("-disable-red-zone is not supported on this target")
			gb_exit(1)
		}
	}

	if bc.Metrics.Os == TargetOsFreestanding {
		bc.NoEntryPoint = true
	} else {
		if bc.NoRTTI {
			gb_printf_err("-no-rtti is only allowed on freestanding targets\n")
			gb_exit(1)
		}
	}

	if bc.ODINWINDOWSSUBSYSTEM == WindowsSubsystemUNKNOWN && bc.Metrics.Os == TargetOsWindows {
		bc.ODINWINDOWSSUBSYSTEM = WindowsSubsystemCONSOLE
	}

	if subtarget == SubtargetAndroid {
		switch bc.BuildMode {
		case BuildModeDynamicLibrary, BuildModeObject, BuildModeAssembly, BuildModeLLVMIR:
			break
		default:
			if (bc.CommandKind & CommandDoesBuild) != 0 {
				gb_printf_err("Unsupported -build-mode for -subtarget:android\n")
				gb_printf_err("\tCurrently only supporting: \n")
				gb_printf_err("\t\tshared\n")
				gb_printf_err("\t\tobject\n")
				gb_printf_err("\t\tassembly\n")
				gb_printf_err("\t\tllvm-ir\n")
				gb_exit(1)
			}
		}
	}

	if metrics.Os == TargetOsDarwin {
		switch subtarget {
		case SubtargetIPhone:
			switch metrics.Arch {
			case TargetArchArm64:
				bc.Metrics.TargetTriplet = "arm64-apple-ios"
			default:
				panic("Unknown architecture for -subtarget:iphone")
			}
		case SubtargetIPhoneSimulator:
			switch metrics.Arch {
			case TargetArchArm64:
				bc.Metrics.TargetTriplet = "arm64-apple-ios-simulator"
			case TargetArchAmd64:
				bc.Metrics.TargetTriplet = "x86_64-apple-ios-simulator"
			default:
				panic("Unknown architecture for -subtarget:iphonesimulator")
			}
		}
	} else if metrics.Os == TargetOsLinux && subtarget == SubtargetAndroid {
		switch metrics.Arch {
		case TargetArchArm64:
			bc.Metrics.TargetTriplet = "aarch64-linux-android"
			bc.RelocMode = RelocModePIC
		case TargetArchArm32:
			bc.Metrics.TargetTriplet = "armv7a-linux-androideabi"
			bc.RelocMode = RelocModePIC
		case TargetArchAmd64:
			bc.Metrics.TargetTriplet = "x86_64-linux-android"
			bc.RelocMode = RelocModePIC
		case TargetArchI386:
			bc.Metrics.TargetTriplet = "i686-linux-android"
			bc.RelocMode = RelocModePIC
		default:
			panic("Unknown architecture for -subtarget:android")
		}
	}

	if bc.Metrics.Os == TargetOsWindows {
		switch bc.Metrics.Arch {
		case TargetArchAmd64:
			bc.LinkFlags = "/machine:x64 "
		case TargetArchI386:
			bc.LinkFlags = "/machine:x86 "
		}
	} else if bc.Metrics.Os == TargetOsDarwin {
		bc.LinkFlags = concatenate3_strings(permanent_allocator(),
			"-target ",
			bc.Metrics.TargetTriplet,
			" ")
	} else if is_arch_wasm() {
		linkFlags := gb_string_make(heap_allocator(), " ")
		linkFlags = gb_string_appendc(linkFlags, "--stack-first ")
		linkFlags = gb_string_appendc(linkFlags, "-z stack-size=1048576 ")
		if bc.Metrics.Os != TargetOsOrca {
			linkFlags = gb_string_appendc(linkFlags, "--allow-undefined ")
		}
		if bc.NoEntryPoint || bc.Metrics.Os == TargetOsOrca {
			linkFlags = gb_string_appendc(linkFlags, "--no-entry ")
		}
		bc.LinkFlags = make_string_c(linkFlags)
		bc.UseSeparateModules = false
	}
	if bc.Metrics.Arch == TargetArchRiscv64 && bc.CrossCompiling {
		bc.LinkFlags = "-target riscv64 "
	}

	if metrics.Os == TargetOsDarwin {
		if !bc.MinimumOSVersionStringGiven {
			if subtarget == SubtargetDefault {
				bc.MinimumOSVersionString = "11.0.0"
			} else if subtarget == SubtargetIPhone || subtarget == SubtargetIPhoneSimulator {
				bc.MinimumOSVersionString = "17.4.0"
			}
		}
		bc.MinimumOSVersionString = normalize_minimum_os_version_string(bc.MinimumOSVersionString)
		if subtarget == SubtargetIPhoneSimulator {
			suffix := "-simulator"
			if !string_ends_with(bc.Metrics.TargetTriplet, suffix) {
				panic("string_ends_with(bc.metrics.target_triplet, suffix)")
			}
			prefix := substring(bc.Metrics.TargetTriplet, 0, len(bc.Metrics.TargetTriplet)-len(suffix))
			bc.Metrics.TargetTriplet = concatenate3_strings(permanent_allocator(), prefix, bc.MinimumOSVersionString, suffix)
		} else {
			bc.Metrics.TargetTriplet = concatenate_strings(permanent_allocator(), bc.Metrics.TargetTriplet, bc.MinimumOSVersionString)
		}
	} else if selectedSubtarget == SubtargetAndroid {
		init_android_values(bc.BuildMode == BuildModeExecutable && (bc.CommandKind&CommandDoesBuild) != 0)
	}

	if !bc.CustomOptimizationLevel {
		if bc.ODINDEBUG {
			bc.OptimizationLevel = -1
		} else {
			bc.OptimizationLevel = 0
		}
	}
	if bc.OptimizationLevel < -1 {
		bc.OptimizationLevel = -1
	}
	if bc.OptimizationLevel > 3 {
		bc.OptimizationLevel = 3
	}

	if bc.OptimizationLevel <= 0 {
		if !is_arch_wasm() {
			bc.UseSeparateModules = true
		}
	}
	if bc.UseSingleModule {
		bc.UseSeparateModules = false
	}

	if bc.LTOKind == LTOThin || bc.LTOKind == LTOThinFiles {
		if bc.BuildMode == BuildModeAssembly || bc.BuildMode == BuildModeLLVMIR {
			gb_printf_err("-lto:thin is incompatible with -build-mode:asm and -build-mode:llvm-ir\n")
			gb_exit(1)
		}
		if bc.LinkerChoice != LinkerLld {
			gb_printf_err("-lto:thin on Windows requires -linker:lld\n")
			gb_exit(1)
		}
		if bc.UseSingleModule {
			gb_printf_err("Warning: -lto:thin overrides -use-single-module; separate modules will be used\n")
		}
		bc.UseSeparateModules = true
		if bc.LTOKind == LTOThinFiles {
			bc.ModulePerFile = true
		}
	}

	bc.ODINVALGRINDSUPPORT = false
	if buildContext.Metrics.Os != TargetOsWindows {
		switch bc.Metrics.Arch {
		case TargetArchAmd64:
			bc.ODINVALGRINDSUPPORT = true
		}
	}

	if bc.Metrics.Os == TargetOsFreestanding {
		bc.ODINDEFAULTTONILALLOCATOR = !bc.ODINDEFAULTOPANICALLOCATOR
	}
}
