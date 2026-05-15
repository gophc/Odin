// Depends on: common.odin (gb_printf, ODIN_VERSION, odin_cpuid, String, isize, u8)
package cmd

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

var (
	ntdll              = windows.NewLazySystemDLL("ntdll.dll")
	procRtlGetVersion  = ntdll.NewProc("RtlGetVersion")
	kernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procGetProductInfo = kernel32.NewProc("GetProductInfo")
)

type osVersionInfoExW struct {
	dwOSVersionInfoSize uint32
	dwMajorVersion      uint32
	dwMinorVersion      uint32
	dwBuildNumber       uint32
	dwPlatformId        uint32
	wServicePackMajor   uint16
	wServicePackMinor   uint16
	wSuiteMask          uint16
	wProductType        uint8
	wReserved           uint8
}

type memoryStatusEx struct {
	dwLength           uint32
	dwMemoryLoad       uint32
	ullTotalPhys       uint64
	ullAvailPhys       uint64
	ullTotalPage       uint64
	ullAvailPage       uint64
	ullTotalVirtual    uint64
	ullAvailVirtual    uint64
	ullExtendedVirtual uint64
}

func report_windows_product_type(productType uint32) {
	switch productType {
	case 0x00000001:
		fmt.Print("Ultimate")
	case 0x00000002:
		fmt.Print("Home Basic")
	case 0x00000003:
		fmt.Print("Home Premium")
	case 0x00000004:
		fmt.Print("Enterprise")
	case 0x00000065:
		fmt.Print("Home Basic")
	case 0x00000005:
		fmt.Print("Home Basic N")
	case 0x00000079:
		fmt.Print("Education")
	case 0x0000007A:
		fmt.Print("Education N")
	case 0x00000006:
		fmt.Print("Business")
	case 0x00000007:
		fmt.Print("Standard Server")
	case 0x00000008:
		fmt.Print("Datacenter")
	case 0x00000009:
		fmt.Print("Windows Small Business Server")
	case 0x0000000A:
		fmt.Print("Enterprise Server")
	case 0x0000000B:
		fmt.Print("Starter")
	case 0x0000000C:
		fmt.Print("Datacenter Server Core")
	case 0x0000000D:
		fmt.Print("Server Standard Core")
	case 0x0000000E:
		fmt.Print("Enterprise Server Core")
	case 0x00000010:
		fmt.Print("Business N")
	case 0x00000013:
		fmt.Print("Home Server")
	case 0x00000018:
		fmt.Print("Windows Server 2008 for Windows Essential Server Solutions")
	case 0x00000019:
		fmt.Print("Small Business Server Premium")
	case 0x0000001A:
		fmt.Print("Home Premium N")
	case 0x0000001B:
		fmt.Print("Enterprise N")
	case 0x0000001C:
		fmt.Print("Ultimate N")
	case 0x0000002A:
		fmt.Print("HyperV")
	case 0x0000002F:
		fmt.Print("Starter N")
	case 0x00000030:
		fmt.Print("Professional")
	case 0x00000031:
		fmt.Print("Professional N")
	case 0xABCDABCD:
		fmt.Print("Unlicensed")
	default:
		fmt.Printf("Unknown Edition (%08x)", productType)
	}
}

func report_cpu_info() {
	fmt.Print("\tCPU:     ")
	var cpu [4]int32
	odin_cpuid(0x80000000, &cpu[0])
	numberOfExtendedIDs := cpu[0]
	var brand [0x12]int32
	if numberOfExtendedIDs >= 0x80000004 {
		odin_cpuid(0x80000002, &brand[0])
		odin_cpuid(0x80000003, &brand[4])
		odin_cpuid(0x80000004, &brand[8])
		brandName := (*[48]byte)(unsafe.Pointer(&brand[0]))[:]
		i := 0
		for i < len(brandName) && brandName[i] == ' ' {
			i++
		}
		fmt.Printf("%s\n", string(brandName[i:]))
	} else {
		fmt.Println("Unable to retrieve.")
	}
}

func report_ram_info() {
	fmt.Print("\tRAM:     ")
	statex := memoryStatusEx{dwLength: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	mod := windows.NewLazySystemDLL("kernel32.dll")
	proc := mod.NewProc("GlobalMemoryStatusEx")
	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&statex)))
	if ret != 0 {
		fmt.Printf("%d MiB\n", statex.ullTotalPhys/(1024*1024))
	}
}

func report_os_info() {
	fmt.Print("\tOS:      ")
	osvi := osVersionInfoExW{dwOSVersionInfoSize: uint32(unsafe.Sizeof(osVersionInfoExW{}))}

	status, _, _ := procRtlGetVersion.Call(uintptr(unsafe.Pointer(&osvi)))
	if status != 0 {
		fmt.Println("Windows (Unknown Version)")
		return
	}

	var productType uint32
	procGetProductInfo.Call(
		uintptr(osvi.dwMajorVersion), uintptr(osvi.dwMinorVersion),
		uintptr(osvi.wServicePackMajor), uintptr(osvi.wServicePackMinor),
		uintptr(unsafe.Pointer(&productType)),
	)

	fmt.Print("Windows ")
	switch osvi.dwMajorVersion {
	case 10:
		switch osvi.wProductType {
		case 1:
			if osvi.dwBuildNumber < 22000 {
				fmt.Print("10 ")
			} else {
				fmt.Print("11 ")
			}
			report_windows_product_type(productType)
		default:
			switch osvi.dwBuildNumber {
			case 14393:
				fmt.Print("2016 Server")
			case 17763:
				fmt.Print("2019 Server")
			case 20348:
				fmt.Print("2022 Server")
			default:
				fmt.Print("Unknown Server")
			}
		}
	case 6:
		switch osvi.dwMinorVersion {
		case 0:
			switch osvi.wProductType {
			case 1:
				fmt.Print("Windows Vista ")
				report_windows_product_type(productType)
			case 3:
				fmt.Print("Windows Server 2008")
			}
		case 1:
			switch osvi.wProductType {
			case 1:
				fmt.Print("Windows 7 ")
				report_windows_product_type(productType)
			case 3:
				fmt.Print("Windows Server 2008 R2")
			}
		case 2:
			switch osvi.wProductType {
			case 1:
				fmt.Print("Windows 8 ")
				report_windows_product_type(productType)
			case 3:
				fmt.Print("Windows Server 2012")
			}
		case 3:
			switch osvi.wProductType {
			case 1:
				fmt.Print("Windows 8.1 ")
				report_windows_product_type(productType)
			case 3:
				fmt.Print("Windows Server 2012 R2")
			}
		}
	case 5:
		switch osvi.dwMinorVersion {
		case 0:
			fmt.Print("Windows 2000")
		case 1:
			fmt.Print("Windows XP")
		case 2:
			fmt.Print("Windows Server 2003")
		}
	}

	var ubr uint32
	var displayVersion [256]byte
	valueSize := uint32(256)
	status, _, _ = windows.RegGetValue(
		windows.HKEY_LOCAL_MACHINE,
		"SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
		"DisplayVersion",
		windows.RRF_RT_REG_SZ,
		nil,
		(*byte)(unsafe.Pointer(&displayVersion[0])),
		&valueSize,
	)
	if status == 0 {
		fmt.Printf(" (version: %s)", string(displayVersion[:valueSize-1]))
	}
	fmt.Printf(", build %d", osvi.dwBuildNumber)
	valueSize = uint32(unsafe.Sizeof(ubr))
	status, _, _ = windows.RegGetValue(
		windows.HKEY_LOCAL_MACHINE,
		"SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
		"UBR",
		windows.RRF_RT_DWORD,
		nil,
		(*byte)(unsafe.Pointer(&ubr)),
		&valueSize,
	)
	if status == 0 {
		fmt.Printf(".%d", ubr)
	}
	fmt.Println()
}

func report_backend_info() {
	fmt.Printf("\tBackend: LLVM %s\n", "20.1.0")
}

func print_bug_report_help() {
	fmt.Println("Where to find more information and get into contact when you encounter a bug:\n")
	fmt.Println("\tWebsite: https://odin-lang.org")
	fmt.Println("\tGitHub:  https://github.com/odin-lang/Odin/issues")
	fmt.Println()
	fmt.Println("Useful information to add to a bug report:\n")
	fmt.Printf("\tOdin:    %.*s", len(ODIN_VERSION.Text), ODIN_VERSION.Text)
	version := "d5fbfeb51"
	if version != "" {
		fmt.Printf(":%s", version)
	}
	fmt.Println()
	report_os_info()
	report_cpu_info()
	report_ram_info()
	report_backend_info()
}
