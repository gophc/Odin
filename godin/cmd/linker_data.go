// Depends on: common.odin (String, Entity, CheckerInfo, BlockingMutex, PtrSet, Array, BuildContext, isize)
package cmd

type LinkerData struct {
	ForeignMutex             BlockingMutex
	ForeignLibrariesSet      PtrSet[*Entity]
	ForeignLibraries         []*Entity
	OutputObjectPaths        []String
	OutputTempPaths          []String
	OutputBase               String
	OutputName               String
	NeedsSystemLibraryLinked bool
}

func linker_enable_system_library_linking(ld *LinkerData) {
	ld.NeedsSystemLibraryLinked = true
}

func linker_data_init(ld *LinkerData, info *CheckerInfo, initFullpath String) {
	ha := heap_allocator()
	ld.OutputObjectPaths = make([]String, 0)
	ld.OutputTempPaths = make([]String, 0)
	ld.ForeignLibraries = make([]*Entity, 0, 1024)
	ptr_set_init(&ld.ForeignLibrariesSet, 1024)
	ld.NeedsSystemLibraryLinked = false

	if build_context.out_filepath.Len == 0 {
		ld.OutputName = remove_directory_from_path(initFullpath)
		ld.OutputName = remove_extension_from_path(ld.OutputName)
		ld.OutputName = string_trim_whitespace(ld.OutputName)
		if ld.OutputName.Len == 0 {
			ld.OutputName = info.InitScope.Pkg.Name
		}
		ld.OutputBase = ld.OutputName
	} else {
		ld.OutputName = build_context.out_filepath
		ld.OutputName = string_trim_whitespace(ld.OutputName)
		if ld.OutputName.Len == 0 {
			ld.OutputName = info.InitScope.Pkg.Name
		}
		pos := string_extension_position(ld.OutputName)
		if pos < 0 {
			ld.OutputBase = ld.OutputName
		} else {
			ld.OutputBase = substring(ld.OutputName, 0, pos)
		}
	}
	ld.OutputBase = path_to_full_path(ha, ld.OutputBase)
}
