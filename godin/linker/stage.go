package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BuildPath indices
const (
	BuildPathOutput = iota
	BuildPathVSExe
	BuildPathVSLib
	BuildPathWinSDKUM_Lib
	BuildPathWinSDKUCRT_Lib
	BuildPathWinSDKBinPath
	BuildPathSymbols
	BuildPathRES
	BuildPathRC
)

// linkerStage performs the actual linking step.
// It mimics the original linker_stage function.
func linkerStage(ld *LinkerData, buildCtx *BuildContext, timings *Timings, info *CheckerInfo) (int, error) {
	// _ = result
	// _ = isCrossLinking
	outputFilename := buildCtx.BuildPaths.Output // In original, it's path_to_string(build_context.build_paths[BuildPath_Output])
	if outputFilename == "" {
		// Fallback: construct from OutputBase
		outputFilename = ld.OutputBase
		if buildCtx.BuildMode == BuildModeExecutable {
			if buildCtx.Metrics.OS == TargetOSWindows {
				outputFilename += ".exe"
			}
		} else if buildCtx.BuildMode == BuildModeDynamicLibrary {
			if buildCtx.Metrics.OS == TargetOSWindows {
				outputFilename += ".dll"
			} else if buildCtx.Metrics.OS == TargetOSDarwin {
				outputFilename += ".dylib"
			} else {
				outputFilename += ".so"
			}
		} else if buildCtx.BuildMode == BuildModeStaticLibrary {
			if buildCtx.Metrics.OS == TargetOSWindows {
				outputFilename += ".lib"
			} else {
				outputFilename += ".a"
			}
		}
	}

	debugf("Linking %s\n", outputFilename)

	if buildCtx.Metrics.OS == TargetOSOrca {
		// wasm-ld path
		timingsStartSection(timings, "wasm-ld")
		libStr := &gbString{}
		inputs := &gbString{}
		inputs = gbStringAppendFmt(inputs, "\"%s.o\"", outputFilename)

		for _, e := range ld.ForeignLibraries {
			if e == nil {
				continue
			}
			// In original, it asserts e->kind == Entity_LibraryName
			extraLinkerFlags := strings.TrimSpace(e.ExtraLinkerFlags)
			if extraLinkerFlags != "" {
				libStr = gbStringAppendFmt(libStr, " %s", extraLinkerFlags)
			}
			for _, lib := range e.Paths {
				if lib == "" {
					continue
				}
				if strings.HasSuffix(lib, ".o") {
					inputs = gbStringAppendFmt(inputs, " \"%s\"", lib)
				}
			}
		}

		var extraOrcaFlags *gbString
		if buildCtx.Metrics.OS == TargetOSOrca {
			orcaSdkPath := &gbString{}
			if !systemExecCommandLineAppOutput("orca sdk-path", orcaSdkPath) {
				fmt.Fprintf(os.Stderr, "executing `orca sdk-path` failed, make sure Orca is installed and added to your path\n")
				return 1, fmt.Errorf("orca sdk-path failed")
			}
			if orcaSdkPath.String() == "" {
				fmt.Fprintf(os.Stderr, "executing `orca sdk-path` did not produce output\n")
				return 1, fmt.Errorf("orca sdk-path empty")
			}
			inputs = gbStringAppendFmt(inputs, " \"%s/orca-libc/lib/crt1.o\" \"%s/orca-libc/lib/libc.a\"", orcaSdkPath.String(), orcaSdkPath.String())
			extraOrcaFlags = gbStringAppendFmt(&gbString{}, " -L \"%s/bin\" -lorca_wasm --export-dynamic", orcaSdkPath.String())
		}

		cmd := fmt.Sprintf("\"%s\\bin\\wasm-ld\" %s -o \"%s\" %s %s %s %s",
			buildCtx.ODINRoot,
			inputs.String(),
			outputFilename,
			buildCtx.LinkFlags,
			buildCtx.ExtraLinkerFlags,
			libStr.String(),
			extraOrcaFlags.String(),
		)
		result, err := systemExecCommandLineApp("wasm-ld", cmd)
		if err != nil {
			return result, err
		}
		return 0, nil
	}

	isCrossLinking := false
	isAndroid := false
	if buildCtx.CrossCompiling && (buildCtx.DifferentOS || buildCtx.SelectedSubtarget != SubtargetDefault) {
		switch buildCtx.SelectedSubtarget {
		case SubtargetAndroid:
			isCrossLinking = true
			isAndroid = true
			_ = isCrossLinking
			// goto try_cross_linking
		default:
			fmt.Fprintf(os.Stderr, "Linking for cross compilation for this platform is not yet supported (%s %s)\n",
				targetOSNames[buildCtx.Metrics.OS], targetArchNames[buildCtx.Metrics.Arch])
			buildCtx.KeepObjectFiles = true
			return 0, nil
		}
	}

	// try_cross_linking:
	sectionName := "msvc-link"
	isWindows := buildCtx.Metrics.OS == TargetOSWindows
	isOSX := buildCtx.Metrics.OS == TargetOSDarwin

	switch buildCtx.LinkerChoice {
	case LinkerDefault:
		// nothing
	case LinkerLLD:
		sectionName = "lld-link"
	case LinkerRadLink:
		sectionName = "rad-link"
	default:
		fmt.Fprintf(os.Stderr, "'%s' linker is not supported on this platform\n", linkerChoices[buildCtx.LinkerChoice])
		return 1, fmt.Errorf("unsupported linker")
	}

	if isWindows {
		timingsStartSection(timings, sectionName)
		libStr := &gbString{}
		linkSettings := gbStringMakeReserve(256)

		if buildCtx.BuildPaths.VSLib != "" {
			addPath := func(path string) {
				if strings.HasSuffix(path, "\\") {
					path = path[:len(path)-1]
				}
				linkSettings = gbStringAppendFmt(linkSettings, " /LIBPATH:\"%s\"", path)
			}
			addPath(buildCtx.BuildPaths.WinSDKUMLib)
			addPath(buildCtx.BuildPaths.WinSDKUCRTLib)
			addPath(buildCtx.BuildPaths.VSLib)
		}

		minLibsSet := NewStringSet(64)
		asmFiles := NewStringSet(64)
		prevLib := ""

		for _, e := range ld.ForeignLibraries {
			if e == nil {
				continue
			}
			extraLinkerFlags := strings.TrimSpace(e.ExtraLinkerFlags)
			if extraLinkerFlags != "" {
				libStr = gbStringAppendFmt(libStr, " %s", extraLinkerFlags)
			}
			for _, lib := range e.Paths {
				lib = strings.TrimSpace(lib)
				libLower := strings.ToLower(lib)
				if lib == "" {
					continue
				}
				if hasASMExtension(lib) {
					if !asmFiles.Add(lib) {
						continue
					}
					asmFile := lib
					var objFile string
					tempDir := temporaryDirectory()
					if tempDir != "" {
						filename := filenameWithoutDirectory(asmFile)
						objFile = filepath.Join(tempDir, fmt.Sprintf("%s-%p.obj", filename, &asmFile))
					} else {
						objFile = asmFile + ".obj"
					}
					objFormat := "win64"
					cmd := fmt.Sprintf("\"%s\\bin\\nasm\\windows\\nasm.exe\" \"%s\" -f \"%s\" -o \"%s\" %s",
						buildCtx.ODINRoot, asmFile, objFormat, objFile, buildCtx.ExtraAssemblerFlags)
					result, err := systemExecCommandLineApp("nasm", cmd)
					if err != nil {
						return result, err
					}
					ld.AddOutputObjectPath(objFile)
				} else if !minLibsSet.Add(libLower) || !buildCtx.MinLinkLibs {
					if prevLib != lib {
						libStr = gbStringAppendFmt(libStr, " \"%s\"", lib)
					}
					prevLib = lib
				}
			}
		}

		if buildCtx.BuildMode == BuildModeDynamicLibrary {
			linkSettings = gbStringAppendFmt(linkSettings, " /DLL")
			if buildCtx.NoEntryPoint {
				linkSettings = gbStringAppendFmt(linkSettings, " /NOENTRY")
			}
		} else {
			if !(buildCtx.Metrics.Arch == TargetArchI386 && !buildCtx.NoCRT) {
				linkSettings = gbStringAppendFmt(linkSettings, " /ENTRY:mainCRTStartup")
			}
		}

		if buildCtx.BuildPaths.Symbols != "" {
			symbolPath := buildCtx.BuildPaths.Symbols
			linkSettings = gbStringAppendFmt(linkSettings, " /PDB:\"%s\"", symbolPath)
		}

		if buildCtx.BuildMode != BuildModeStaticLibrary {
			if buildCtx.NoCRT {
				linkSettings = gbStringAppendFmt(linkSettings, " /nodefaultlib")
			} else {
				linkSettings = gbStringAppendFmt(linkSettings, " /defaultlib:libcmt")
			}
		}

		if buildCtx.ODINDebug {
			linkSettings = gbStringAppendFmt(linkSettings, " /DEBUG")
		}

		objectFiles := &gbString{}
		for _, objPath := range ld.OutputObjectPaths {
			objectFiles = gbStringAppendFmt(objectFiles, "\"%s\" ", objPath)
		}

		vsExePath := buildCtx.BuildPaths.VSExe
		windowsSDKBinPath := buildCtx.BuildPaths.WinSDKBinPath

		lldLTOFlags := &gbString{}
		if buildCtx.LTOKind != LTONone {
			lldLTOFlags = gbStringAppendFmt(lldLTOFlags, "/opt:lldltojobs=%d ", buildCtx.ThreadCount)
		}

		switch buildCtx.LinkerChoice {
		case LinkerLLD:
			cmd := fmt.Sprintf("\"%s\\bin\\lld-link\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s %s",
				buildCtx.ODINRoot,
				objectFiles.String(),
				outputFilename,
				linkSettings.String(),
				windowsSubsystemNames[buildCtx.WindowsSubsystem],
				buildCtx.LinkFlags,
				buildCtx.ExtraLinkerFlags,
				libStr.String(),
				lldLTOFlags.String(),
			)
			result, err := systemExecCommandLineApp("msvc-lld-link", cmd)
			if err != nil {
				return result, err
			}
		case LinkerRadLink:
			cmd := fmt.Sprintf("\"%s\\bin\\radlink\" %s -OUT:\"%s\" %s /nologo /incremental:no /opt:ref /subsystem:%s %s %s %s",
				buildCtx.ODINRoot,
				objectFiles.String(),
				outputFilename,
				linkSettings.String(),
				windowsSubsystemNames[buildCtx.WindowsSubsystem],
				buildCtx.LinkFlags,
				buildCtx.ExtraLinkerFlags,
				libStr.String(),
			)
			result, err := systemExecCommandLineApp("msvc-rad-link", cmd)
			if err != nil {
				return result, err
			}
		default:
			resPath := quotePath(buildCtx.BuildPaths.RES)
			rcPath := quotePath(buildCtx.BuildPaths.RC)
			if buildCtx.HasResource {
				if buildCtx.BuildPaths.RC == "" {
					debugf("Using precompiled resource %s\n", resPath)
				} else {
					debugf("Compiling resource %s\n", resPath)
					cmd := fmt.Sprintf("\"%src.exe\" /nologo /fo %s %s",
						windowsSDKBinPath, resPath, rcPath)
					result, err := systemExecCommandLineApp("msvc-link", cmd)
					if err != nil {
						return result, err
					}
				}
			} else {
				resPath = ""
			}

			linkerName := "link.exe"
			if buildCtx.BuildMode == BuildModeStaticLibrary {
				linkerName = "lib.exe"
			} else {
				linkSettings = gbStringAppendFmt(linkSettings, " /incremental:no /opt:ref")
			}
			if buildCtx.BuildMode == BuildModeExecutable {
				linkSettings = gbStringAppendFmt(linkSettings, " /NOIMPLIB /NOEXP")
			}

			cmd := fmt.Sprintf("\"%s%s\" %s %s -OUT:\"%s\" %s /nologo /subsystem:%s %s %s %s",
				vsExePath, linkerName,
				objectFiles.String(),
				resPath,
				outputFilename,
				linkSettings.String(),
				windowsSubsystemNames[buildCtx.WindowsSubsystem],
				buildCtx.LinkFlags,
				buildCtx.ExtraLinkerFlags,
				libStr.String(),
			)
			result, err := systemExecCommandLineApp("msvc-link", cmd)
			if err != nil {
				return result, err
			}
		}
	} else {
		// Non-Windows: using clang or other native linkers
		timingsStartSection(timings, sectionName)
		clangPath := os.Getenv("ODIN_CLANG_PATH")
		hasOdinClangPathEnv := true
		if clangPath == "" {
			clangPath = "clang"
			hasOdinClangPathEnv = false
		}

		libStr := &gbString{}
		asmFiles := NewStringSet(64)
		minLibsSet := NewStringSet(64)
		prevLib := ""

		for _, e := range ld.ForeignLibraries {
			if e == nil {
				continue
			}
			extraLinkerFlags := strings.TrimSpace(e.ExtraLinkerFlags)
			if extraLinkerFlags != "" {
				libStr = gbStringAppendFmt(libStr, " %s", extraLinkerFlags)
			}
			if buildCtx.Metrics.OS == TargetOSDarwin {
				for _, lib := range e.Paths {
					lib = strings.TrimSpace(lib)
					if lib == "" {
						continue
					}
					if strings.HasSuffix(lib, ".framework") {
						if minLibsSet.Add(lib) {
							continue
						}
						libName := strings.TrimSuffix(lib, ".framework")
						libStr = gbStringAppendFmt(libStr, " -framework %s ", libName)
					}
				}
			}
			for _, lib := range e.Paths {
				lib = strings.TrimSpace(lib)
				if lib == "" {
					continue
				}
				if hasASMExtension(lib) {
					if asmFiles.Add(lib) {
						continue
					}
					asmFile := lib
					var objFile string
					tempDir := temporaryDirectory()
					if tempDir != "" {
						filename := filenameWithoutDirectory(asmFile)
						objFile = filepath.Join(tempDir, fmt.Sprintf("%s-%p.o", filename, &asmFile))
					} else {
						objFile = asmFile + ".o"
					}
					var objFormat string
					if buildCtx.Metrics.PtrSize == 8 {
						if isOSX {
							objFormat = "macho64"
						} else {
							objFormat = "elf64"
						}
					} else {
						if isOSX {
							objFormat = "macho32"
						} else {
							objFormat = "elf32"
						}
					}

					if buildCtx.Metrics.Arch == TargetArchRiscV64 {
						cmd := fmt.Sprintf("%s \"%s\" -c -o \"%s\" -target %s -march=rv64gc %s",
							clangPath, asmFile, objFile, buildCtx.Metrics.TargetTriplet, buildCtx.ExtraAssemblerFlags)
						result, err := systemExecCommandLineApp("clang", cmd)
						if err != nil {
							return result, err
						}
					} else if isOSX {
						cmd := fmt.Sprintf("as \"%s\" -o \"%s\" %s", asmFile, objFile, buildCtx.ExtraAssemblerFlags)
						result, err := systemExecCommandLineApp("as", cmd)
						if err != nil {
							return result, err
						}
					} else {
						cmd := fmt.Sprintf("nasm \"%s\" -f \"%s\" -o \"%s\" %s",
							asmFile, objFormat, objFile, buildCtx.ExtraAssemblerFlags)
						result, err := systemExecCommandLineApp("nasm", cmd)
						if err != nil {
							fmt.Fprintf(os.Stderr, "executing `nasm` to assemble foreign import of %s failed.\n\tSuggestion: `nasm` does not ship with the compiler and should be installed with your system's package manager.\n", asmFile)
							return result, err
						}
					}
					ld.AddOutputObjectPath(objFile)
				} else {
					shortCircuit := false
					if strings.HasSuffix(lib, ".framework") ||
						strings.HasSuffix(lib, ".dylib") ||
						strings.HasSuffix(lib, ".so") ||
						e.IgnoreDuplicates {
						shortCircuit = true
					}
					if minLibsSet.Add(lib) && (buildCtx.MinLinkLibs || shortCircuit) {
						continue
					}
					if prevLib == lib {
						continue
					}
					prevLib = lib
					if lib == "System.framework" || lib == "System" || lib == "c" {
						continue
					}
					if buildCtx.Metrics.OS == TargetOSDarwin {
						if strings.HasSuffix(lib, ".framework") {
							libName := strings.TrimSuffix(lib, ".framework")
							libStr = gbStringAppendFmt(libStr, " -framework %s ", libName)
						} else if strings.HasSuffix(lib, ".a") || strings.HasSuffix(lib, ".o") || strings.HasSuffix(lib, ".dylib") {
							libStr = gbStringAppendFmt(libStr, " \"%s\" ", lib)
						} else {
							libStr = gbStringAppendFmt(libStr, " -l%s ", lib)
						}
					} else {
						if strings.HasSuffix(lib, ".a") || strings.HasSuffix(lib, ".o") || strings.HasSuffix(lib, ".so") || strings.Contains(lib, ".so.") {
							libStr = gbStringAppendFmt(libStr, " -l:\"%s\" ", lib)
						} else {
							libStr = gbStringAppendFmt(libStr, " -l%s ", lib)
						}
					}
				}
			}
		}

		objectFiles := &gbString{}
		if isAndroid {
			// Android native app glue compilation and archiving
			debugf("[Section] %s\n", "Android Native App Glue Compile")
			if buildCtx.ShowMoreTimings {
				timingsStartSection(timings, "Android Native App Glue Compile")
			}
			androidGlueObject := ""
			androidGlueStaticLib := ""
			hash := fmt.Sprintf("%p", &androidGlueObject)
			tempDir := normalizePath(temporaryDirectory())
			androidGlueObject = filepath.Join(tempDir, fmt.Sprintf("android_native_app_glue-%s.o", hash))
			androidGlueStaticLib = filepath.Join(tempDir, fmt.Sprintf("libandroid_native_app_glue-%s.a", hash))

			glueCmd := fmt.Sprintf("%sbin/clang --target=%s%d -c \"%ssources/android/native_app_glue/android_native_app_glue.c\" -o \"%s\" --sysroot \"%ssysroot\" \"-I%ssysroot/usr/include/\" \"-I%ssysroot/usr/include/%s/\" -Wno-macro-redefined",
				buildCtx.ODINAndroidNDKToolchain,
				buildCtx.Metrics.TargetTriplet, buildCtx.ODINAndroidAPILevel,
				buildCtx.ODINAndroidNDK,
				androidGlueObject,
				buildCtx.ODINAndroidNDKToolchain,
				buildCtx.ODINAndroidNDKToolchain,
				buildCtx.ODINAndroidNDKToolchain,
				buildCtx.ODINAndroidNDKToolchainLib,
			)
			result, err := systemExecCommandLineApp("android-native-app-glue-compile", glueCmd)
			if err != nil {
				return result, err
			}

			debugf("[Section] %s\n", "Android Native App Glue ar")
			if buildCtx.ShowMoreTimings {
				timingsStartSection(timings, "Android Native App Glue ar")
			}
			arCmd := fmt.Sprintf("%sbin/llvm-ar rcs \"%s\" \"%s\"",
				buildCtx.ODINAndroidNDKToolchain,
				androidGlueStaticLib, androidGlueObject)
			result, err = systemExecCommandLineApp("android-native-app-glue-ar", arCmd)
			if err != nil {
				return result, err
			}
			objectFiles = gbStringAppendFmt(objectFiles, "\"%s\" ", androidGlueStaticLib)
		}

		for _, objPath := range ld.OutputObjectPaths {
			objectFiles = gbStringAppendFmt(objectFiles, "\"%s\" ", objPath)
		}

		linkSettings := gbStringMakeReserve(32)
		if buildCtx.NoCRT {
			linkSettings = gbStringAppendFmt(linkSettings, "-nostdlib ")
		}

		if buildCtx.BuildMode == BuildModeStaticLibrary {
			debugf("[Section] %s\n", "Static Library Creation")
			if buildCtx.ShowMoreTimings {
				timingsStartSection(timings, "Static Library Creation")
			}
			arCmd := fmt.Sprintf("ar rcs \"%s\" %s", outputFilename, objectFiles.String())
			result, err := systemExecCommandLineApp("ar", arCmd)
			if err != nil {
				return result, err
			}
			return 0, nil
		}

		if buildCtx.BuildMode == BuildModeDynamicLibrary {
			linkSettings = gbStringAppendFmt(linkSettings, "-shared ")
			if buildCtx.Metrics.OS == TargetOSDarwin {
				linkSettings = gbStringAppendFmt(linkSettings, "-Wl,-init,'__odin_entry_point' ")
			} else {
				linkSettings = gbStringAppendFmt(linkSettings, "-Wl,-init,'_odin_entry_point' ")
				linkSettings = gbStringAppendFmt(linkSettings, "-Wl,-fini,'_odin_exit_point' ")
			}
		} else if isAndroid {
			linkSettings = gbStringAppendFmt(linkSettings, "-shared ")
		}

		if buildCtx.BuildMode == BuildModeExecutable && buildCtx.RelocMode == RelocModePIC {
			// do nothing
		} else if buildCtx.BuildMode != BuildModeDynamicLibrary {
			if buildCtx.Metrics.OS != TargetOSOpenBSD &&
				buildCtx.Metrics.OS != TargetOSHaiku &&
				buildCtx.Metrics.Arch != TargetArchRiscV64 &&
				!isAndroid {
				linkSettings = gbStringAppendFmt(linkSettings, "-no-pie ")
			}
		}

		platformLibStr := &gbString{}
		if buildCtx.Metrics.OS == TargetOSDarwin {
			darwinPlatformName := "MacOSX"
			darwinXCRunSDKName := "macosx"
			darwinMinVersionID := "macosx"
			originalClangPath := clangPath
			switch buildCtx.SelectedSubtarget {
			case SubtargetIPhone:
				darwinPlatformName = "iPhoneOS"
				darwinXCRunSDKName = "iphoneos"
				darwinMinVersionID = "ios"
				if !hasOdinClangPathEnv {
					clangPath = "xcrun --sdk iphoneos clang"
				}
			case SubtargetIPhoneSimulator:
				darwinPlatformName = "iPhoneSimulator"
				darwinXCRunSDKName = "iphonesimulator"
				darwinMinVersionID = "ios-simulator"
				if !hasOdinClangPathEnv {
					clangPath = "xcrun --sdk iphonesimulator clang"
				}
			}

			darwinSdkPath := &gbString{}
			findSdkCmd := fmt.Sprintf("xcrun --sdk %s --show-sdk-path", darwinXCRunSDKName)
			if !systemExecCommandLineAppOutput(findSdkCmd, darwinSdkPath) {
				clangPath = originalClangPath
				darwinSdkPath = &gbString{}
				darwinSdkPath = gbStringAppendFmt(darwinSdkPath, "/Library/Developer/CommandLineTools/SDKs/%s.sdk", darwinPlatformName)
				if !pathIsDirectory(darwinSdkPath.String()) {
					darwinSdkPath = &gbString{}
					darwinSdkPath = gbStringAppendFmt(darwinSdkPath, "/Applications/Xcode.app/Contents/Developer/Platforms/%s.platform/Developer/SDKs/%s.sdk", darwinPlatformName, darwinPlatformName)
					if !pathIsDirectory(darwinSdkPath.String()) {
						fmt.Fprintf(os.Stderr, "Failed to find %s SDK\n", darwinPlatformName)
						return -1, fmt.Errorf("SDK not found")
					}
				}
			} else {
				darwinSdkPath = gbStringTrimSpace(darwinSdkPath)
			}
			platformLibStr = gbStringAppendFmt(platformLibStr, "--sysroot %s ", darwinSdkPath.String())
			platformLibStr = gbStringAppendFmt(platformLibStr, "-L/usr/local/lib ")
			if _, err := os.Stat("/opt/homebrew/lib"); err == nil {
				platformLibStr = gbStringAppendFmt(platformLibStr, "-L/opt/homebrew/lib ")
			}
			if _, err := os.Stat("/opt/local/lib"); err == nil {
				platformLibStr = gbStringAppendFmt(platformLibStr, "-L/opt/local/lib ")
			}
			if buildCtx.MinimumOSVersionStringGiven || buildCtx.SelectedSubtarget != SubtargetDefault {
				linkSettings = gbStringAppendFmt(linkSettings, "-m%s-version-min=%s ", darwinMinVersionID, buildCtx.MinimumOSVersionString)
			}
			if buildCtx.BuildMode != BuildModeDynamicLibrary {
				linkSettings = gbStringAppendFmt(linkSettings, "-e _main ")
			}
		} else if buildCtx.Metrics.OS == TargetOSFreeBSD {
			if buildCtx.SanitizerFlags&(SanitizerFlagAddress|SanitizerFlagMemory) != 0 {
				platformLibStr = gbStringAppendFmt(platformLibStr, "-lpthread ")
			}
			platformLibStr = gbStringAppendFmt(platformLibStr, "-Wl,-L/usr/local/lib ")
		} else if buildCtx.Metrics.OS == TargetOSOpenBSD {
			platformLibStr = gbStringAppendFmt(platformLibStr, "-lpthread -Wl,-L/usr/local/lib ")
			platformLibStr = gbStringAppendFmt(platformLibStr, "-Wl,-z,nobtcfi ")
		}

		if isAndroid {
			platformLibStr = gbStringAppendFmt(platformLibStr, "\"-L%ssysroot/usr/lib/%s/%d\" ",
				buildCtx.ODINAndroidNDKToolchainSysroot,
				buildCtx.ODINAndroidNDKToolchainLib,
				buildCtx.ODINAndroidAPILevel)
			platformLibStr = gbStringAppendFmt(platformLibStr, "-landroid ")
			platformLibStr = gbStringAppendFmt(platformLibStr, "-llog ")
			platformLibStr = gbStringAppendFmt(platformLibStr, "\"--sysroot=%s\" ", buildCtx.ODINAndroidNDKToolchainSysroot)
			linkSettings = gbStringAppendFmt(linkSettings, "-u ANativeActivity_onCreate ")
		}

		if !buildCtx.NoRPATH {
			if buildCtx.Metrics.OS == TargetOSDarwin {
				linkSettings = gbStringAppendFmt(linkSettings, "-Wl,-rpath,@loader_path ")
			} else {
				if !isAndroid {
					linkSettings = gbStringAppendFmt(linkSettings, "-Wl,-rpath,\\$ORIGIN ")
				}
			}
		}

		if !buildCtx.NoCRT {
			libStr = gbStringAppendFmt(libStr, "-lm ")
			if buildCtx.Metrics.OS != TargetOSDarwin {
				libStr = gbStringAppendFmt(libStr, "-lc ")
			}
		}

		linkCommandLine := &gbString{}
		if isAndroid {
			ndkBinDir := buildCtx.ODINAndroidNDKToolchain
			linkCommandLine = gbStringAppend(linkCommandLine, ndkBinDir+"bin/clang")
			linkCommandLine = gbStringAppendFmt(linkCommandLine, " --target=%s%d ", buildCtx.Metrics.TargetTriplet, buildCtx.ODINAndroidAPILevel)
		} else {
			linkCommandLine = gbStringAppend(linkCommandLine, clangPath)
		}
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " -Wno-unused-command-line-argument ")

		if buildCtx.LTOKind != LTONone {
			linkCommandLine = gbStringAppendFmt(linkCommandLine, " -flto=thin")
			linkCommandLine = gbStringAppendFmt(linkCommandLine, " -flto-jobs=%d ", buildCtx.ThreadCount)
			if buildCtx.ODINDebug {
				linkCommandLine = gbStringAppendFmt(linkCommandLine, " -g ")
			}
			if isOSX && !buildCtx.MinimumOSVersionStringGiven {
				linkCommandLine = gbStringAppendFmt(linkCommandLine, " -Wno-override-module ")
			}
		}

		linkCommandLine = gbStringAppend(linkCommandLine, objectFiles.String())
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " -o \"%s\" ", outputFilename)
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " %s ", platformLibStr.String())
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " %s ", libStr.String())
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " %s ", buildCtx.LinkFlags)
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " %s ", buildCtx.ExtraLinkerFlags)
		linkCommandLine = gbStringAppendFmt(linkCommandLine, " %s ", linkSettings.String())

		if isAndroid && buildCtx.ShowMoreTimings {
			debugf("[Section] %s\n", "Linking")
			timingsStartSection(timings, "Linking")
		}

		switch buildCtx.LinkerChoice {
		case LinkerLLD:
			linkCommandLine = gbStringAppendFmt(linkCommandLine, " -fuse-ld=lld")
			result, err := systemExecCommandLineApp("lld-link", linkCommandLine.String())
			if err != nil {
				return result, err
			}
		case LinkerMold:
			linkCommandLine = gbStringAppendFmt(linkCommandLine, " -fuse-ld=mold")
			result, err := systemExecCommandLineApp("mold-link", linkCommandLine.String())
			if err != nil {
				return result, err
			}
		default:
			result, err := systemExecCommandLineApp("ld-link", linkCommandLine.String())
			if err != nil {
				return result, err
			}
		}

		if isOSX && buildCtx.ODINDebug {
			result, err := systemExecCommandLineApp("dsymutil", fmt.Sprintf("dsymutil \"%s\"", outputFilename))
			if err != nil {
				return result, err
			}
		}
	}

	return 0, nil
}
