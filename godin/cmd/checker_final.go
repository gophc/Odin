package cmd

import (
	"sort"
	"unsafe"
)

func sectionLog(name string) {
	debugf("[Section] %s\n", name)
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c(name))
	}
}

func check_unique_package_names(c *Checker) bool {
	ok := true
	pkgs := make(map[string]*AstPackage)
	for _, entry := range c.Info.Packages {
		pkg := entry.Value
		if len(pkg.Files) == 0 {
			continue
		}
		name := goStr(pkg.Name)
		if found, exists := pkgs[name]; exists {
			curr := pkg.Files[0].PkgDecl
			prev := found.Files[0].PkgDecl
			if curr == prev {
				continue
			}
			ok = false
			begin_error_block()
			error_(curr, "Duplicate declaration of 'package %s'", name)
			error_line("\tA package name must be unique\n" +
				"\tThere is no relation between a package name and the directory that contains it, so they can be completely different\n" +
				"\tA package name is required for link name prefixing to have a consistent ABI\n")
			error_line("%s found at previous location\n", goStr(token_pos_to_string(ast_token(prev).Pos)))
			dirA := pkg.Files[0].Directory
			dirB := found.Files[0].Directory
			if string_eq_ignore_case(dirA, dirB) {
				error_line("\tRemember that Windows case-folds paths, and so %s and %s are the same directory.\n", goStr(dirA), goStr(dirB))
			}
			end_error_block()
		} else {
			pkgs[name] = pkg
		}
	}
	return ok
}

func check_add_entities_from_queues(c *Checker) {
	for {
		e := mpsc_dequeue(c.Info.EntityQueue, &c.Info.EntityQueue)
		if e == nil {
			break
		}
		c.Info.Entities = append(c.Info.Entities, e)
	}
}

func check_add_definitions_from_queues(c *Checker) {
	for {
		e := mpsc_dequeue(c.Info.DefinitionQueue, &c.Info.DefinitionQueue)
		if e == nil {
			break
		}
		c.Info.Definitions = append(c.Info.Definitions, e)
	}
}

func check_merge_queues_into_arrays(c *Checker) {
	for {
		t := mpsc_dequeue(c.SoaTypesToComplete, &c.SoaTypesToComplete)
		if t == nil {
			break
		}
		complete_soa_type(c, t, false)
	}
	check_add_entities_from_queues(c)
	check_add_definitions_from_queues(c)
	thread_pool_wait()
}

func init_procedures_less(x, y *Entity) bool {
	if x == y {
		return false
	}
	if x.Pkg != y.Pkg {
		orderX := isize(0)
		orderY := isize(0)
		if x.Pkg != nil {
			orderX = x.Pkg.Order
		}
		if y.Pkg != nil {
			orderY = y.Pkg.Order
		}
		if orderX != orderY {
			return orderX < orderY
		}
	}
	if x.File != y.File {
		fullpathX := String{}
		fullpathY := String{}
		if x.File != nil {
			fullpathX = x.File.Fullpath
		}
		if y.File != nil {
			fullpathY = y.File.Fullpath
		}
		fileX := filename_from_path(fullpathX)
		fileY := filename_from_path(fullpathY)
		cmp := string_compare(fileX, fileY)
		if cmp != 0 {
			return cmp < 0
		}
	}
	if x.OrderInSrc != y.OrderInSrc {
		return x.OrderInSrc < y.OrderInSrc
	}
	return x.Token.Pos.Offset < y.Token.Pos.Offset
}

func check_sort_init_and_fini_procedures(c *Checker) {
	sort.Slice(c.Info.InitProcedures, func(i, j int) bool {
		return init_procedures_less(c.Info.InitProcedures[i], c.Info.InitProcedures[j])
	})
	sort.Slice(c.Info.FiniProcedures, func(i, j int) bool {
		return init_procedures_less(c.Info.FiniProcedures[j], c.Info.FiniProcedures[i])
	})
	remove_neighbouring_duplicate_entires_from_sorted_array(&c.Info.InitProcedures)
	remove_neighbouring_duplicate_entires_from_sorted_array(&c.Info.FiniProcedures)
}

func add_type_info_for_type_definitions(c *Checker) {
	for _, e := range c.Info.Definitions {
		if e.Kind == Entity_TypeName && e.Type != nil && is_type_typed(e.Type) {
			if e.MinDepCount.Load() > 0 {
				add_type_info_type(&c.BuiltinCtx, e.Type)
			}
		}
	}
}

var check_walk_all_dependencies_worker_proc WorkerTaskProc = func(data unsafe.Pointer) isize {
	if data == nil {
		return 0
	}
	decl := (*DeclInfo)(data)
	for child := decl.NextChild; child != nil; child = child.NextSibling {
		thread_pool_add_task(check_walk_all_dependencies_worker_proc, unsafe.Pointer(child))
		check_walk_all_dependencies(child)
	}
	add_deps_from_child_to_parent(decl)
	return 0
}

func check_walk_all_dependencies(decl *DeclInfo) {
	if decl != nil {
		thread_pool_add_task(check_walk_all_dependencies_worker_proc, unsafe.Pointer(decl))
	}
}

func check_update_dependency_tree_for_procedures(c *Checker) {
	mutex_lock(&c.NestedProcLitsMutex)
	for _, decl := range c.NestedProcLits {
		check_walk_all_dependencies(decl)
	}
	mutex_unlock(&c.NestedProcLitsMutex)
	for _, e := range c.Info.Entities {
		decl := e.DeclInfo
		check_walk_all_dependencies(decl)
	}
	thread_pool_wait()
}

var check_scope_usage_file_worker WorkerTaskProc = func(data unsafe.Pointer) isize {
	c := globalCheckerPtr.Load()
	f := (*AstFile)(data)
	vet_flags := ast_file_vet_flags(f)
	check_scope_usage(c, f.Scope, vet_flags)
	return 0
}

var check_scope_usage_pkg_worker WorkerTaskProc = func(data unsafe.Pointer) isize {
	c := globalCheckerPtr.Load()
	pkg := (*AstPackage)(data)
	check_scope_usage_internal(c, pkg.Scope, 0, true)
	return 0
}

func check_all_scope_usages(c *Checker) {
	for _, entry := range c.Info.Files {
		f := entry.Value
		thread_pool_add_task(check_scope_usage_file_worker, unsafe.Pointer(f))
	}
	for _, entry := range c.Info.Packages {
		pkg := entry.Value
		thread_pool_add_task(check_scope_usage_pkg_worker, unsafe.Pointer(pkg))
	}
	thread_pool_wait()
}

func check_for_type_cycles(c *Checker) {
	for i := 0; i < len(c.Info.Definitions); i++ {
		e := c.Info.Definitions[i]
		if e.Kind != Entity_TypeName {
			continue
		}
		if e.Type != nil && is_type_typed(e.Type) {
			if e.TypeName.IsTypeAlias {
			} else {
				type_align_of(e.Type)
			}
		}
	}
}

func check_for_inline_cycles(c *Checker) {
	for i := 0; i < len(c.Info.Definitions); i++ {
		e := c.Info.Definitions[i]
		if e.Kind != Entity_Procedure {
			continue
		}
		decl := e.DeclInfo
		if decl == nil || decl.ProcLit == nil {
			continue
		}
		pl := &decl.ProcLit.ProcLit
		if pl.Inlining == ProcInliningInline {
			if decl.Deps != nil {
				for _, dep := range decl.Deps.Keys {
					if dep == nil {
						continue
					}
					if dep == e {
						error_(e.Token, "Cannot inline recursive procedure '%s'", goStr(e.Token.String))
						break
					}
				}
			}
		}
	}
}

func check_parsed_files(c *Checker) {
	globalCheckerPtr.Store(c)

	sectionLog("map full filepaths to scope")
	add_type_info_type(&c.BuiltinCtx, tInvalid)
	for i := 0; i < len(c.Parser.Packages); i++ {
		p := c.Parser.Packages[i]
		scope := create_scope_from_package(&c.BuiltinCtx, p)
		p.DeclInfo = make_decl_info(scope, c.BuiltinCtx.Decl)
		string_map_set(&c.Info.Packages, p.Fullpath, p)
		if scope.Flags&int32(ScopeFlag_Init) != 0 {
			c.Info.InitPackage = p
			c.Info.InitScope = scope
		}
		if p.Kind == PackageRuntime {
			gb_assert_handler("Assertion Failure", "c.Info.RuntimePackage == nil", "checker_final.go", 0)
			c.Info.RuntimePackage = p
		}
	}

	sectionLog("init worker data")
	check_init_worker_data(c)

	sectionLog("create file scopes")
	check_create_file_scopes(c)

	sectionLog("collect entities")
	check_collect_entities_all(c)

	sectionLog("export entities - pre")
	check_export_entities(c)
	check_import_entities(c)

	sectionLog("export entities - post")
	check_export_entities(c)

	sectionLog("add entities from packages")
	check_merge_queues_into_arrays(c)

	sectionLog("check all global entities")
	check_all_global_entities(c)

	sectionLog("init preload")
	init_preload(c)

	sectionLog("add global untyped expression to queue")
	add_untyped_expressions(&c.Info, &c.Info.GlobalUntyped)

	prevCtx := c.BuiltinCtx
	defer func() {
		c.BuiltinCtx = prevCtx
	}()
	c.BuiltinCtx.Decl = make_decl_info(nil, nil)

	sectionLog("check procedure bodies")
	check_procedure_bodies(c)

	sectionLog("check foreign import fullpaths")
	check_foreign_import_fullpaths(c)

	sectionLog("add entities from procedure bodies")
	check_merge_queues_into_arrays(c)

	sectionLog("check all scope usages")
	check_all_scope_usages(c)

	sectionLog("add basic type information")
	for i := 0; i < BasicCOUNT; i++ {
		t := &basicTypes[i]
		if t.Basic.Size > 0 && t.Basic.Flags&BasicFlag_LLVM == 0 {
			add_type_info_type(&c.BuiltinCtx, t)
		}
	}
	check_merge_queues_into_arrays(c)

	sectionLog("check for type cycles")
	check_for_type_cycles(c)

	sectionLog("check for inline cycles")
	check_for_inline_cycles(c)

	sectionLog("check deferred procedures")
	check_deferred_procedures(c)

	sectionLog("check objc context provider procedures")
	check_objc_context_provider_procedures(c)

	sectionLog("calculate global init order")
	calculate_global_init_order(c)

	sectionLog("add type info for type definitions")
	add_type_info_for_type_definitions(c)
	check_merge_queues_into_arrays(c)

	sectionLog("update dependency tree for procedures")
	check_update_dependency_tree_for_procedures(c)

	sectionLog("generate minimum dependency set")
	generate_minimum_dependency_set(c, c.Info.EntryPoint)

	sectionLog("check bodies have all been checked")
	check_unchecked_bodies(c)

	sectionLog("check #soa types")
	check_merge_queues_into_arrays(c)

	sectionLog("update minimum dependency set again")
	generate_minimum_dependency_set_internal(c, c.Info.EntryPoint)

	sectionLog("check test procedures")
	check_test_procedures(c)
	check_merge_queues_into_arrays(c)
	thread_pool_wait()

	sectionLog("check entry point")
	if buildContext.BuildMode == BuildModeExecutable && !buildContext.NoEntryPoint && buildContext.CommandKind&CommandTest == 0 {
		s := c.Info.InitScope
		gb_assert_handler("Assertion Failure", "s != nil", "checker_final.go", 0)
		gb_assert_handler("Assertion Failure", "s.Flags&int32(ScopeFlag_Init) != 0", "checker_final.go", 0)
		interned := string_interner_insert(make_string_c("main"))
		e := scope_lookup_current(s, interned, interned.Hash())
		if e == nil {
			token := Token{}
			token.Pos.FileID = 0
			token.Pos.Line = 1
			token.Pos.Column = 1
			if len(s.Pkg.Files) > 0 {
				f := s.Pkg.Files[0]
				if len(f.Tokens) > 0 {
					token = f.Tokens[0]
				}
			}
			error_(token, "Undefined entry point procedure 'main'")
		}
	} else if buildContext.BuildMode == BuildModeDynamicLibrary && buildContext.NoEntryPoint {
		c.Info.EntryPoint = nil
	}
	thread_pool_wait()

	gb_assert_handler("Assertion Failure", "len(c.ProcsToCheck) == 0", "checker_final.go", 0)

	if true {
		sectionLog("check unchecked (safety measure)")
		check_safety_all_procedures_for_unchecked(c)
	}

	sectionLog("check unique package names")
	packageNamesAreUnique := check_unique_package_names(c)

	sectionLog("sanity checks")
	check_merge_queues_into_arrays(c)

	sectionLog("check instrumentation calls")
	{
		enter := c.Info.InstrumentationEnterEntity
		exit := c.Info.InstrumentationExitEntity
		if (enter != nil) != (exit != nil) {
			e := enter
			if e == nil {
				e = exit
			}
			error_(e.Token, "Both @(instrumentation_enter) and @(instrumentation_exit) must be defined")
		}
	}

	sectionLog("add untyped expression values")
	for {
		u := mpsc_dequeue(c.GlobalUntypedQueue, &c.GlobalUntypedQueue)
		if u.Expr == nil && u.Info == nil {
			break
		}
		if is_type_typed(u.Info.Type) {
			compiler_error("%s (type %s) is typed!", expr_to_string(u.Expr), type_to_string(u.Info.Type))
		}
		add_type_and_value(&c.BuiltinCtx, u.Expr, u.Info.Mode, u.Info.Type, u.Info.Value)
	}

	sectionLog("initialize and check for collisions in type info array")
	{
		typeInfoTypes := make([]TypeInfoPair, 0)
		for _, tt := range c.Info.MinDepTypeInfoSet.Keys {
			if tt.Hash != 0 && tt.Hash != TypeSetTombstone {
				typeInfoTypes = append(typeInfoTypes, tt)
			}
		}
		sort.Slice(typeInfoTypes, func(i, j int) bool {
			return type_info_pair_cmp(&typeInfoTypes[i], &typeInfoTypes[j]) < 0
		})
		c.Info.TypeInfoTypesHashMap = make([]TypeInfoPair, len(typeInfoTypes)*2+1)
		hashMapLen := isize(len(c.Info.TypeInfoTypesHashMap))
		for _, tt := range typeInfoTypes {
			index := isize(int(tt.Hash) % hashMapLen)
			for {
				if index == 0 || c.Info.TypeInfoTypesHashMap[index].Hash != 0 {
					index = (index + 1) % hashMapLen
					continue
				}
				break
			}
			c.Info.TypeInfoTypesHashMap[index] = tt
			exists := false
			found := ptr_map_get(c.Info.MinDepTypeInfoIndexMap, tt.Hash)
			if found == nil {
				ptr_map_set(c.Info.MinDepTypeInfoIndexMap, tt.Hash, index)
			} else {
				exists = true
			}
			if packageNamesAreUnique && exists {
				other := c.Info.TypeInfoTypesHashMap[*found]
				if !are_types_identical_unique_tuples(tt.Type, other.Type) {
					t := temp_canonical_string(tt.Type)
					o := temp_canonical_string(other.Type)
					gb_assert_handler("Panic", 0, "checker_final.go", 0,
						"%s (%s) %d vs %s (%s) %d",
						type_to_string(tt.Type, false), t, tt.Hash,
						type_to_string(other.Type, false), o, other.Hash)
				}
			}
		}
	}

	sectionLog("sort init and fini procedures")
	check_sort_init_and_fini_procedures(c)

	{
		var intrinsicsNodes []*Ast
		for {
			node := mpsc_dequeue(c.Info.IntrinsicsEntryPointUsage, &c.Info.IntrinsicsEntryPointUsage)
			if node == nil {
				break
			}
			intrinsicsNodes = append(intrinsicsNodes, node)
		}
		if len(intrinsicsNodes) > 0 {
			sectionLog("check intrinsics.__entry_point usage")
			for _, node := range intrinsicsNodes {
				if c.Info.EntryPoint == nil {
					file := thread_safe_get_ast_file_from_id(node.FileID)
					if file.Pkg.Kind != PackageRuntime {
						error_(node, "usage of intrinsics.__entry_point will be a no-op")
					}
				}
			}
		}
	}

	sectionLog("collate type info stuff")
	for {
		typeView := mpsc_dequeue(c.Info.RaddbgTypeViewsQueue, &c.Info.RaddbgTypeViewsQueue)
		if typeView.Type == nil {
			break
		}
		handle_raddbg_type_view(c, typeView)
	}

	sectionLog("type check finish")
}
