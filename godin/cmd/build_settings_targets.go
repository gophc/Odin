package cmd

var targetEndians = [TargetArchCOUNT]TargetEndianKind{
	TargetEndianLittle,
	TargetEndianLittle,
	TargetEndianLittle,
	TargetEndianLittle,
	TargetEndianLittle,
	TargetEndianLittle,
	TargetEndianLittle,
}

var targetWindowsI386 = TargetMetrics{
	Os:            TargetOsWindows,
	Arch:          TargetArchI386,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      16,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("i386-pc-windows-msvc"), Len: isize(len("i386-pc-windows-msvc"))},
}

var targetWindowsAmd64 = TargetMetrics{
	Os:            TargetOsWindows,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-pc-windows-msvc"), Len: isize(len("x86_64-pc-windows-msvc"))},
}

var targetLinuxI386 = TargetMetrics{
	Os:            TargetOsLinux,
	Arch:          TargetArchI386,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      16,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("i386-pc-linux-gnu"), Len: isize(len("i386-pc-linux-gnu"))},
}

var targetLinuxAmd64 = TargetMetrics{
	Os:            TargetOsLinux,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-pc-linux-gnu"), Len: isize(len("x86_64-pc-linux-gnu"))},
}

var targetLinuxArm64 = TargetMetrics{
	Os:            TargetOsLinux,
	Arch:          TargetArchArm64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("aarch64-linux-elf"), Len: isize(len("aarch64-linux-elf"))},
}

var targetLinuxArm32 = TargetMetrics{
	Os:            TargetOsLinux,
	Arch:          TargetArchArm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("arm-unknown-linux-gnueabihf"), Len: isize(len("arm-unknown-linux-gnueabihf"))},
}

var targetLinuxRiscv64 = TargetMetrics{
	Os:            TargetOsLinux,
	Arch:          TargetArchRiscv64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("riscv64-linux-gnu"), Len: isize(len("riscv64-linux-gnu"))},
}

var targetDarwinAmd64 = TargetMetrics{
	Os:            TargetOsDarwin,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-apple-macosx"), Len: isize(len("x86_64-apple-macosx"))},
}

var targetDarwinArm64 = TargetMetrics{
	Os:            TargetOsDarwin,
	Arch:          TargetArchArm64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("arm64-apple-macosx"), Len: isize(len("arm64-apple-macosx"))},
}

var targetFreeBSDI386 = TargetMetrics{
	Os:            TargetOsFreeBSD,
	Arch:          TargetArchI386,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      16,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("i386-unknown-freebsd-elf"), Len: isize(len("i386-unknown-freebsd-elf"))},
}

var targetFreeBSDAmd64 = TargetMetrics{
	Os:            TargetOsFreeBSD,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-unknown-freebsd-elf"), Len: isize(len("x86_64-unknown-freebsd-elf"))},
}

var targetFreeBSDArm64 = TargetMetrics{
	Os:            TargetOsFreeBSD,
	Arch:          TargetArchArm64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("aarch64-unknown-freebsd-elf"), Len: isize(len("aarch64-unknown-freebsd-elf"))},
}

var targetOpenBSDAmd64 = TargetMetrics{
	Os:            TargetOsOpenBSD,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-unknown-openbsd-elf"), Len: isize(len("x86_64-unknown-openbsd-elf"))},
}

var targetNetBSDAmd64 = TargetMetrics{
	Os:            TargetOsNetBSD,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-unknown-netbsd-elf"), Len: isize(len("x86_64-unknown-netbsd-elf"))},
}

var targetNetBSDArm64 = TargetMetrics{
	Os:            TargetOsNetBSD,
	Arch:          TargetArchArm64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("aarch64-unknown-netbsd-elf"), Len: isize(len("aarch64-unknown-netbsd-elf"))},
}

var targetHaikuAmd64 = TargetMetrics{
	Os:            TargetOsHaiku,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-unknown-haiku"), Len: isize(len("x86_64-unknown-haiku"))},
}

var targetFreestandingWasm32 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchWasm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-freestanding-js"), Len: isize(len("wasm32-freestanding-js"))},
}

var targetJsWasm32 = TargetMetrics{
	Os:            TargetOsJs,
	Arch:          TargetArchWasm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-js-js"), Len: isize(len("wasm32-js-js"))},
}

var targetWasiWasm32 = TargetMetrics{
	Os:            TargetOsWasi,
	Arch:          TargetArchWasm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-wasi-js"), Len: isize(len("wasm32-wasi-js"))},
}

var targetOrcaWasm32 = TargetMetrics{
	Os:            TargetOsOrca,
	Arch:          TargetArchWasm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-wasi-js"), Len: isize(len("wasm32-wasi-js"))},
}

var targetFreestandingWasm64p32 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchWasm64p32,
	PtrSize:       4,
	IntSize:       8,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-freestanding-js"), Len: isize(len("wasm32-freestanding-js"))},
}

var targetJsWasm64p32 = TargetMetrics{
	Os:            TargetOsJs,
	Arch:          TargetArchWasm64p32,
	PtrSize:       4,
	IntSize:       8,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-js-js"), Len: isize(len("wasm32-js-js"))},
}

var targetWasiWasm64p32 = TargetMetrics{
	Os:            TargetOsWasi,
	Arch:          TargetArchWasm32,
	PtrSize:       4,
	IntSize:       8,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("wasm32-wasi-js"), Len: isize(len("wasm32-wasi-js"))},
}

var targetFreestandingAmd64SysV = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-pc-none-gnu"), Len: isize(len("x86_64-pc-none-gnu"))},
	ABI:           TargetABISysV,
}

var targetFreestandingAmd64Win64 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-pc-windows-msvc"), Len: isize(len("x86_64-pc-windows-msvc"))},
	ABI:           TargetABIWin64,
}

var targetFreestandingAmd64Mingw = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchAmd64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("x86_64-pc-windows-gnu"), Len: isize(len("x86_64-pc-windows-gnu"))},
	ABI:           TargetABIWin64,
}

var targetFreestandingArm64 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchArm64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("aarch64-none-elf"), Len: isize(len("aarch64-none-elf"))},
}

var targetFreestandingArm32 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchArm32,
	PtrSize:       4,
	IntSize:       4,
	MaxAlign:      8,
	MaxSimdAlign:  16,
	TargetTriplet: String{Data: strData("arm-unknown-unknown-gnueabihf"), Len: isize(len("arm-unknown-unknown-gnueabihf"))},
}

var targetFreestandingRiscv64 = TargetMetrics{
	Os:            TargetOsFreestanding,
	Arch:          TargetArchRiscv64,
	PtrSize:       8,
	IntSize:       8,
	MaxAlign:      16,
	MaxSimdAlign:  32,
	TargetTriplet: String{Data: strData("riscv64-unknown-gnu"), Len: isize(len("riscv64-unknown-gnu"))},
}

var namedTargets = []NamedTargetMetrics{
	{Name: String{Data: strData("darwin_amd64"), Len: isize(len("darwin_amd64"))}, Metrics: &targetDarwinAmd64},
	{Name: String{Data: strData("darwin_arm64"), Len: isize(len("darwin_arm64"))}, Metrics: &targetDarwinArm64},
	{Name: String{Data: strData("linux_i386"), Len: isize(len("linux_i386"))}, Metrics: &targetLinuxI386},
	{Name: String{Data: strData("linux_amd64"), Len: isize(len("linux_amd64"))}, Metrics: &targetLinuxAmd64},
	{Name: String{Data: strData("linux_arm64"), Len: isize(len("linux_arm64"))}, Metrics: &targetLinuxArm64},
	{Name: String{Data: strData("linux_arm32"), Len: isize(len("linux_arm32"))}, Metrics: &targetLinuxArm32},
	{Name: String{Data: strData("linux_riscv64"), Len: isize(len("linux_riscv64"))}, Metrics: &targetLinuxRiscv64},
	{Name: String{Data: strData("windows_i386"), Len: isize(len("windows_i386"))}, Metrics: &targetWindowsI386},
	{Name: String{Data: strData("windows_amd64"), Len: isize(len("windows_amd64"))}, Metrics: &targetWindowsAmd64},
	{Name: String{Data: strData("freebsd_i386"), Len: isize(len("freebsd_i386"))}, Metrics: &targetFreeBSDI386},
	{Name: String{Data: strData("freebsd_amd64"), Len: isize(len("freebsd_amd64"))}, Metrics: &targetFreeBSDAmd64},
	{Name: String{Data: strData("freebsd_arm64"), Len: isize(len("freebsd_arm64"))}, Metrics: &targetFreeBSDArm64},
	{Name: String{Data: strData("netbsd_amd64"), Len: isize(len("netbsd_amd64"))}, Metrics: &targetNetBSDAmd64},
	{Name: String{Data: strData("netbsd_arm64"), Len: isize(len("netbsd_arm64"))}, Metrics: &targetNetBSDArm64},
	{Name: String{Data: strData("openbsd_amd64"), Len: isize(len("openbsd_amd64"))}, Metrics: &targetOpenBSDAmd64},
	{Name: String{Data: strData("haiku_amd64"), Len: isize(len("haiku_amd64"))}, Metrics: &targetHaikuAmd64},
	{Name: String{Data: strData("freestanding_wasm32"), Len: isize(len("freestanding_wasm32"))}, Metrics: &targetFreestandingWasm32},
	{Name: String{Data: strData("wasi_wasm32"), Len: isize(len("wasi_wasm32"))}, Metrics: &targetWasiWasm32},
	{Name: String{Data: strData("js_wasm32"), Len: isize(len("js_wasm32"))}, Metrics: &targetJsWasm32},
	{Name: String{Data: strData("orca_wasm32"), Len: isize(len("orca_wasm32"))}, Metrics: &targetOrcaWasm32},
	{Name: String{Data: strData("freestanding_wasm64p32"), Len: isize(len("freestanding_wasm64p32"))}, Metrics: &targetFreestandingWasm64p32},
	{Name: String{Data: strData("js_wasm64p32"), Len: isize(len("js_wasm64p32"))}, Metrics: &targetJsWasm64p32},
	{Name: String{Data: strData("wasi_wasm64p32"), Len: isize(len("wasi_wasm64p32"))}, Metrics: &targetWasiWasm64p32},
	{Name: String{Data: strData("freestanding_amd64_sysv"), Len: isize(len("freestanding_amd64_sysv"))}, Metrics: &targetFreestandingAmd64SysV},
	{Name: String{Data: strData("freestanding_amd64_win64"), Len: isize(len("freestanding_amd64_win64"))}, Metrics: &targetFreestandingAmd64Win64},
	{Name: String{Data: strData("freestanding_amd64_mingw"), Len: isize(len("freestanding_amd64_mingw"))}, Metrics: &targetFreestandingAmd64Mingw},
	{Name: String{Data: strData("freestanding_arm64"), Len: isize(len("freestanding_arm64"))}, Metrics: &targetFreestandingArm64},
	{Name: String{Data: strData("freestanding_arm32"), Len: isize(len("freestanding_arm32"))}, Metrics: &targetFreestandingArm32},
	{Name: String{Data: strData("freestanding_riscv64"), Len: isize(len("freestanding_riscv64"))}, Metrics: &targetFreestandingRiscv64},
}

func get_target_os_from_string(str String, subtarget_ *Subtarget, subtarget_str *String) TargetOsKind {
	os_name := str
	subtarget := String{}
	part := string_partition(str, S(":"))
	if part.Match.Len == 1 {
		os_name = part.Head
		subtarget = part.Tail
	}
	kind := TargetOsInvalid
	for i := isize(0); i < isize(TargetOsCOUNT); i++ {
		if str_eq_ignore_case(targetOsNames[i], os_name) {
			kind = TargetOsKind(i)
			break
		}
	}
	if subtarget_str != nil {
		*subtarget_str = subtarget
	}
	if subtarget_ != nil {
		if subtarget.Len != 0 {
			*subtarget_ = SubtargetInvalid
			if str_eq_ignore_case(subtarget, S("generic")) || str_eq_ignore_case(subtarget, S("default")) {
				*subtarget_ = SubtargetDefault
			} else {
				for i := isize(1); i < isize(SubtargetCOUNT); i++ {
					if str_eq_ignore_case(subtargetStrings[i], subtarget) {
						*subtarget_ = Subtarget(i)
						break
					}
				}
			}
		} else {
			*subtarget_ = SubtargetDefault
		}
	}
	return kind
}

func get_target_arch_from_string(str String) TargetArchKind {
	for i := isize(0); i < isize(TargetArchCOUNT); i++ {
		if str_eq_ignore_case(targetArchNames[i], str) {
			return TargetArchKind(i)
		}
	}
	return TargetArchInvalid
}

func is_excluded_target_filename(name String) bool {
	original_name := name
	name = remove_extension_from_path(name)
	if string_starts_with(name, S(".")) {
		return true
	}
	str1 := String{}
	str2 := String{}
	n := isize(0)
	str1 = name
	n = str1.Len
	for i := str1.Len - 1; i >= 0 && str1.Data[i] != '_'; i-- {
		n--
	}
	str1 = substring(str1, n, str1.Len)
	if n-1 > 0 {
		str2 = substring(name, 0, n-1)
	} else {
		str2 = substring(name, 0, 0)
	}
	n = str2.Len
	for i := str2.Len - 1; i >= 0 && str2.Data[i] != '_'; i-- {
		n--
	}
	str2 = substring(str2, n, str2.Len)
	if string_eq(str1, name) {
		return false
	}
	os1 := get_target_os_from_string(str1, nil, nil)
	arch1 := get_target_arch_from_string(str1)
	os2 := get_target_os_from_string(str2, nil, nil)
	arch2 := get_target_arch_from_string(str2)
	if os1 != TargetOsInvalid && arch2 != TargetArchInvalid {
		return os1 != buildContext.Metrics.Os || arch2 != buildContext.Metrics.Arch
	} else if arch1 != TargetArchInvalid && os2 != TargetOsInvalid {
		return arch1 != buildContext.Metrics.Arch || os2 != buildContext.Metrics.Os
	} else if os1 != TargetOsInvalid {
		return os1 != buildContext.Metrics.Os
	} else if arch1 != TargetArchInvalid {
		return arch1 != buildContext.Metrics.Arch
	}
	return false
}

func is_arch_wasm() bool {
	return buildContext.Metrics.Arch == TargetArchWasm32 || buildContext.Metrics.Arch == TargetArchWasm64p32
}

func is_arch_x86() bool {
	return buildContext.Metrics.Arch == TargetArchI386 || buildContext.Metrics.Arch == TargetArchAmd64
}
