package cmd

import (
	"sync/atomic"
	"unsafe"
)

var globalCheckerPtr atomic.Pointer[Checker]

func add_min_dep_type_info(c *Checker, t *Type) {
	if t == nil {
		return
	}
	t = default_type(t)
	if is_type_untyped(t) {
		return
	}
	if is_type_polymorphic(base_type(t)) {
		return
	}
	if type_set_update_with_mutex(&c.Info.MinDepTypeInfoSet, t, &c.Info.MinDepTypeInfoSetMutex) {
		return
	}
	if t.Kind == TypeNamed {
		add_min_dep_type_info(c, t.Named.Base)
		return
	}
	bt := base_type(t)
	add_min_dep_type_info(c, bt)
	switch bt.Kind {
	case TypeInvalid:
		break
	case TypeBasic:
		switch bt.Basic.Kind {
		case BasicString:
			add_min_dep_type_info(c, tU8Ptr)
			add_min_dep_type_info(c, tInt)
		case BasicAny:
			add_min_dep_type_info(c, tRawptr)
			add_min_dep_type_info(c, tTypeid)
		case BasicComplex64:
			add_min_dep_type_info(c, tTypeInfoFloat)
			add_min_dep_type_info(c, tF32)
		case BasicComplex128:
			add_min_dep_type_info(c, tTypeInfoFloat)
			add_min_dep_type_info(c, tF64)
		case BasicQuaternion128:
			add_min_dep_type_info(c, tTypeInfoFloat)
			add_min_dep_type_info(c, tF32)
		case BasicQuaternion256:
			add_min_dep_type_info(c, tTypeInfoFloat)
			add_min_dep_type_info(c, tF64)
		}
	case TypeBitSet:
		add_min_dep_type_info(c, bt.BitSet.Elem)
		add_min_dep_type_info(c, bit_set_to_int(bt))
	case TypePointer:
		add_min_dep_type_info(c, bt.Pointer.Elem)
	case TypeMultiPointer:
		add_min_dep_type_info(c, bt.MultiPointer.Elem)
	case TypeArray:
		add_min_dep_type_info(c, bt.Array.Elem)
		add_min_dep_type_info(c, alloc_type_pointer(bt.Array.Elem))
		add_min_dep_type_info(c, tInt)
	case TypeEnumeratedArray:
		add_min_dep_type_info(c, bt.EnumeratedArray.Index)
		add_min_dep_type_info(c, tInt)
		add_min_dep_type_info(c, bt.EnumeratedArray.Elem)
		add_min_dep_type_info(c, alloc_type_pointer(bt.EnumeratedArray.Elem))
	case TypeDynamicArray:
		add_min_dep_type_info(c, bt.DynamicArray.Elem)
		add_min_dep_type_info(c, alloc_type_pointer(bt.DynamicArray.Elem))
		add_min_dep_type_info(c, tInt)
		add_min_dep_type_info(c, tAllocator)
	case TypeSlice:
		add_min_dep_type_info(c, bt.Slice.Elem)
		add_min_dep_type_info(c, alloc_type_pointer(bt.Slice.Elem))
		add_min_dep_type_info(c, tInt)
	case TypeFixedCapacityDynamicArray:
		add_min_dep_type_info(c, bt.FixedCapacityDynamicArray.Elem)
		add_min_dep_type_info(c, alloc_type_pointer(bt.FixedCapacityDynamicArray.Elem))
		add_min_dep_type_info(c, alloc_type_array(bt.FixedCapacityDynamicArray.Elem, bt.FixedCapacityDynamicArray.Capacity, nil))
		add_min_dep_type_info(c, tInt)
		fallthrough
	case TypeEnum:
		add_min_dep_type_info(c, bt.Enum.BaseType)
	case TypeUnion:
		if union_tag_size(t) > 0 {
			add_min_dep_type_info(c, union_tag_type(t))
		} else {
			add_min_dep_type_info(c, tTypeInfoPtr)
		}
		add_min_dep_type_info(c, bt.Union.PolymorphicParams)
		for _, v := range bt.Union.Variants {
			add_min_dep_type_info(c, v)
		}
	case TypeStruct:
		if bt.Struct.Scope != nil {
			for i := uint32(0); i < bt.Struct.Scope.Elements.Cap; i++ {
				if bt.Struct.Scope.Elements.Slots[i].Hash != 0 {
					e := bt.Struct.Scope.Elements.Slots[i].Value
					switch bt.Struct.SoaKind {
					case StructSoaDynamic:
						add_min_dep_type_info(c, tTypeInfoPtr)
						add_min_dep_type_info(c, tAllocator)
						fallthrough
					case StructSoaSlice:
						add_min_dep_type_info(c, tInt)
						add_min_dep_type_info(c, tUint)
						fallthrough
					case StructSoaFixed:
						add_min_dep_type_info(c, alloc_type_pointer(e.Type))
					default:
						add_min_dep_type_info(c, e.Type)
					}
				}
			}
		}
		add_min_dep_type_info(c, bt.Struct.PolymorphicParams)
		for _, f := range bt.Struct.Fields {
			add_min_dep_type_info(c, f.Type)
		}
	case TypeMap:
		init_map_internal_types(bt)
		add_min_dep_type_info(c, bt.Map.Key)
		add_min_dep_type_info(c, bt.Map.Value)
		add_min_dep_type_info(c, tUintptr)
		add_min_dep_type_info(c, tAllocator)
	case TypeTuple:
		for _, v := range bt.Tuple.Variables {
			add_min_dep_type_info(c, v.Type)
		}
	case TypeProc:
		add_min_dep_type_info(c, bt.Proc.Params)
		add_min_dep_type_info(c, bt.Proc.Results)
	case TypeSimdVector:
		add_min_dep_type_info(c, bt.SimdVector.Elem)
	case TypeMatrix:
		add_min_dep_type_info(c, bt.Matrix.Elem)
	case TypeSoaPointer:
		add_min_dep_type_info(c, bt.SoaPointer.Elem)
	case TypeBitField:
		add_min_dep_type_info(c, bt.BitField.BackingType)
		for _, f := range bt.BitField.Fields {
			add_min_dep_type_info(c, f.Type)
		}
	default:
		gb_assert_handler("Panic", 0, "checker.cpp", 2690, "Unhandled type: %s", type_strings[bt.Kind])
	}
}

func add_dependency_to_set_threaded(c *Checker, entity *Entity)

func add_dependency_to_set(c *Checker, entity *Entity) {
	if entity == nil {
		return
	}
	if entity.Type != nil && is_type_polymorphic(entity.Type) {
		decl := decl_info_of_entity(entity)
		if decl != nil && decl.GenProcType == nil {
			return
		}
	}
	if entity.MinDepCount.Add(1) > 1 {
		return
	}
	decl := decl_info_of_entity(entity)
	if decl == nil {
		return
	}
	for it := begin_type_set(&decl.TypeInfoDeps); !it.Equal(end_type_set(&decl.TypeInfoDeps)); it.Next() {
		add_min_dep_type_info(c, it.Value().Type)
	}
	if decl.Deps != nil {
		for _, dep := range decl.Deps.Keys {
			if dep == nil {
				continue
			}
			switch dep.Kind {
			case Entity_Procedure:
				if dep.Procedure.IsForeign {
					fl := dep.Procedure.ForeignLibrary
					if fl != nil {
						gb_assert_handler("Assertion Failure", `fl->kind == Entity_LibraryName && (fl->flags&EntityFlag_Used)`, "checker.cpp", 2731, "%s", entity.Token.String)
						add_dependency_to_set(c, fl)
					}
				}
			case Entity_Variable:
				if dep.Variable.IsForeign {
					fl := dep.Variable.ForeignLibrary
					if fl != nil {
						gb_assert_handler("Assertion Failure", `fl->kind == Entity_LibraryName && (fl->flags&EntityFlag_Used)`, "checker.cpp", 2742, "%s", entity.Token.String)
						add_dependency_to_set(c, fl)
					}
				}
			}
		}
		for _, dep := range decl.Deps.Keys {
			if dep == nil {
				continue
			}
			add_dependency_to_set(c, dep)
		}
	}
}

func add_dependency_to_set_worker(data unsafe.Pointer) isize {
	c := globalCheckerPtr.Load()
	entity := (*Entity)(data)
	if entity == nil {
		return 0
	}
	if entity.Type != nil && is_type_polymorphic(entity.Type) {
		decl := decl_info_of_entity(entity)
		if decl != nil && decl.GenProcType == nil {
			return 0
		}
	}
	if entity.MinDepCount.Add(1) > 1 {
		return 0
	}
	decl := decl_info_of_entity(entity)
	if decl == nil {
		return 0
	}
	for it := begin_type_set(&decl.TypeInfoDeps); !it.Equal(end_type_set(&decl.TypeInfoDeps)); it.Next() {
		add_min_dep_type_info(c, it.Value().Type)
	}
	if decl.Deps != nil {
		for _, dep := range decl.Deps.Keys {
			if dep == nil {
				continue
			}
			switch dep.Kind {
			case Entity_Procedure:
				if dep.Procedure.IsForeign {
					fl := dep.Procedure.ForeignLibrary
					if fl != nil {
						gb_assert_handler("Assertion Failure", `fl->kind == Entity_LibraryName && (fl->flags&EntityFlag_Used)`, "checker.cpp", 2790, "%s", entity.Token.String)
						add_dependency_to_set_threaded(c, fl)
					}
				}
			case Entity_Variable:
				if dep.Variable.IsForeign {
					fl := dep.Variable.ForeignLibrary
					if fl != nil {
						gb_assert_handler("Assertion Failure", `fl->kind == Entity_LibraryName && (fl->flags&EntityFlag_Used)`, "checker.cpp", 2801, "%s", entity.Token.String)
						add_dependency_to_set_threaded(c, fl)
					}
				}
			}
		}
		for _, dep := range decl.Deps.Keys {
			if dep == nil {
				continue
			}
			add_dependency_to_set_threaded(c, dep)
		}
	}
	return 0
}

func add_dependency_to_set_threaded(c *Checker, entity *Entity) {
	if entity == nil {
		return
	}
	globalCheckerPtr.Store(c)
	thread_pool_add_task(add_dependency_to_set_worker, unsafe.Pointer(entity))
}

func force_add_dependency_entity(c *Checker, scope *Scope, name String) {
	hash := uint32(0)
	interned := string_interner_insert(name)
	hash = interned.Hash()
	e := scope_lookup(scope, interned, hash)
	if e == nil {
		return
	}
	e.Flags |= EntityFlag_Used
	add_dependency_to_set(c, e)
}

func collect_testing_procedures_of_package(c *Checker, pkg *AstPackage) {
	testingPackage := get_core_package(&c.Info, make_string_c("testing"))
	if testingPackage == nil {
		return
	}
	testingScope := testingPackage.Scope
	hash := uint32(0)
	interned := string_interner_insert(make_string_c("Test_Signature"))
	hash = interned.Hash()
	testSignature := scope_lookup_current(testingScope, interned, hash)
	s := pkg.Scope
	for i := uint32(0); i < s.Elements.Cap; i++ {
		if s.Elements.Slots[i].Hash != 0 {
			e := s.Elements.Slots[i].Value
			if e.Kind != Entity_Procedure {
				continue
			}
			if e.Flags&EntityFlag_Test == 0 {
				continue
			}
			isTester := true
			t := base_type(e.Type)
			gb_assert_handler("Assertion Failure", "t->kind == Type_Proc", "checker.cpp", 2860, 0)
			if !are_types_identical(t, base_type(testSignature.Type)) {
				str := type_to_string(t)
				error(nil, "Testing procedures must have a signature type of proc(^testing.T), got %s", str)
				isTester = false
			}
			if isTester {
				add_dependency_to_set(c, e)
				c.Info.TestingProcedures = append(c.Info.TestingProcedures, e)
			}
		}
	}
}

func generate_minimum_dependency_set_internal(c *Checker, start *Entity) {
	addToSet := add_dependency_to_set_threaded
	builtinScope := builtinPkg.Scope
	for i := 0; i < len(c.Info.Definitions); i++ {
		e := c.Info.Definitions[i]
		if e.Scope == builtinScope {
			if e.Type == nil {
				addToSet(c, e)
			}
		} else if e.Kind == Entity_Procedure {
			if e.Procedure.IsExport {
				addToSet(c, e)
			}
		} else if e.Kind == Entity_Variable {
			if e.Variable.IsExport {
				addToSet(c, e)
			}
		}
	}
	for {
		e := mpsc_dequeue(c.Info.RequiredForeignImportsThroughForceQueue)
		if e == nil {
			break
		}
		c.Info.RequiredForeignImportsThroughForce = append(c.Info.RequiredForeignImportsThroughForce, e)
		addToSet(c, e)
	}
	for {
		e := mpsc_dequeue(c.Info.RequiredGlobalVariableQueue)
		if e == nil {
			break
		}
		e.Flags |= EntityFlag_Used
		addToSet(c, e)
	}
	for _, e := range c.Info.Entities {
		switch e.Kind {
		case Entity_Variable:
			if e.Variable.IsExport {
				addToSet(c, e)
			} else if e.Flags&EntityFlag_Require != 0 {
				addToSet(c, e)
			}
		case Entity_Procedure:
			if e.Procedure.IsExport {
				addToSet(c, e)
			} else if e.Flags&EntityFlag_Require != 0 {
				addToSet(c, e)
			}
			if e.Flags&EntityFlag_Init != 0 {
				t := base_type(e.Type)
				gb_assert_handler("Assertion Failure", "t->kind == Type_Proc", "checker.cpp", 2927, 0)
				isInit := true
				if t.Proc.ParamCount != 0 || t.Proc.ResultCount != 0 {
					str := type_to_string(t)
					error(nil, "@(init) procedures must have a signature type with no parameters nor results, got %s", str)
					isInit = false
				}
				featureFlags := check_feature_flags(e)
				if featureFlags&OptInFeatureFlag_GlobalContext == 0 {
					if t.Proc.CallingConvention != ProcCC_Contextless {
						begin_error_block()
						error(nil, "@(init) procedures must be declared as \"contextless\"")
						error_line("\tSuggestion: this can be bypassed, for the time being, with '#+feature global-context'")
						end_error_block()
					}
				}
				if e.Scope.Flags&(ScopeFlag_File|ScopeFlag_Pkg) == 0 {
					error(nil, "@(init) procedures must be declared at the file scope")
					isInit = false
				}
				if e.Flags&EntityFlag_Disabled != 0 {
					warning(nil, "This @(init) procedure is disabled; you must call it manually")
					isInit = false
				}
				if is_blank_ident_token(e.Token) {
					error(nil, "An @(init) procedure must not use a blank identifier as its name")
				}
				if isInit {
					addToSet(c, e)
					c.Info.InitProcedures = append(c.Info.InitProcedures, e)
				}
			} else if e.Flags&EntityFlag_Fini != 0 {
				t := base_type(e.Type)
				gb_assert_handler("Assertion Failure", "t->kind == Type_Proc", "checker.cpp", 2968, 0)
				isFini := true
				if t.Proc.ParamCount != 0 || t.Proc.ResultCount != 0 {
					str := type_to_string(t)
					error(nil, "@(fini) procedures must have a signature type with no parameters nor results, got %s", str)
					isFini = false
				}
				featureFlags := check_feature_flags(e)
				if featureFlags&OptInFeatureFlag_GlobalContext == 0 {
					if t.Proc.CallingConvention != ProcCC_Contextless {
						begin_error_block()
						error(nil, "@(fini) procedures must be declared as \"contextless\"")
						error_line("\tSuggestion: this can be bypassed, for the time being, with '#+feature global-context'")
						end_error_block()
					}
				}
				if e.Scope.Flags&(ScopeFlag_File|ScopeFlag_Pkg) == 0 {
					error(nil, "@(fini) procedures must be declared at the file scope")
					isFini = false
				}
				if is_blank_ident_token(e.Token) {
					error(nil, "An @(fini) procedure must not use a blank identifier as its name")
				}
				if isFini {
					addToSet(c, e)
					c.Info.FiniProcedures = append(c.Info.FiniProcedures, e)
				}
			}
		}
	}
	if buildContext.CommandKind&CommandTest != 0 {
		testingPackage := get_core_package(&c.Info, make_string_c("testing"))
		if testingPackage != nil {
			testingScope := testingPackage.Scope
			for i := uint32(0); i < testingScope.Elements.Cap; i++ {
				if testingScope.Elements.Slots[i].Hash != 0 {
					e := testingScope.Elements.Slots[i].Value
					if e != nil {
						e.Flags |= EntityFlag_Used
						addToSet(c, e)
					}
				}
			}
		}
		pkg := c.Info.InitPackage
		collect_testing_procedures_of_package(c, pkg)
		if buildContext.TestAllPackages {
			for _, entry := range c.Info.Packages {
				collect_testing_procedures_of_package(c, entry.Value)
			}
		}
	} else if start != nil {
		start.Flags |= EntityFlag_Used
		addToSet(c, start)
	}
}

func generate_minimum_dependency_set(c *Checker, start *Entity) {
	{
		names := []string{
			"Source_Code_Location",
			"Context",
			"Allocator",
			"Logger",
			"__init_context",
			"_cleanup_runtime",
			"memset",
			"memory_equal",
			"memory_compare",
			"memory_compare_zero",
		}
		for _, n := range names {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	if buildContext.NoCRT {
		for _, n := range []string{"memcpy", "memmove"} {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	if buildContext.Metrics.Arch == TargetArch_arm32 {
		force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c("aeabi_d2h"))
	}
	if is_arch_wasm() {
		for _, n := range []string{"__ashlti3", "__multi3", "__lshrti3"} {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	if !buildContext.NoRTTI {
		for _, n := range []string{"Type_Info", "type_table", "__type_info_of"} {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	if !buildContext.NoEntryPoint {
		force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c("args__"))
	}
	if buildContext.NoCRT && !is_arch_wasm() {
		for _, n := range []string{"_tls_index", "_fltused"} {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	if !buildContext.NoBoundsCheck {
		names := []string{
			"bounds_check_error",
			"matrix_bounds_check_error",
			"slice_expr_error_hi",
			"slice_expr_error_lo_hi",
			"multi_pointer_slice_expr_error",
		}
		for _, n := range names {
			force_add_dependency_entity(c, c.Info.RuntimePackage.Scope, make_string_c(n))
		}
	}
	add_dependency_to_set(c, c.Info.InstrumentationEnterEntity)
	add_dependency_to_set(c, c.Info.InstrumentationExitEntity)
	generate_minimum_dependency_set_internal(c, start)
	thread_pool_wait()
}

func is_entity_a_dependency(e *Entity) bool {
	switch e.Kind {
	case Entity_Procedure:
		return true
	case Entity_Constant, Entity_Variable:
		return e.Pkg != nil
	case Entity_TypeName:
		return false
	}
	return false
}

func generate_entity_dependency_graph(info *CheckerInfo, arena *Arena) []*EntityGraphNode {
	MProcs := ptr_map_new[*Entity, *EntityGraphNode]()
	MVars := ptr_map_new[*Entity, *EntityGraphNode]()
	MOther := ptr_map_new[*Entity, *EntityGraphNode]()

	for _, e := range info.Entities {
		if e == nil || !is_entity_a_dependency(e) {
			continue
		}
		n := arena_alloc_item[*EntityGraphNode](arena)
		n.Entity = e
		switch e.Kind {
		case Entity_Procedure:
			ptr_map_set(MProcs, e, n)
		case Entity_Variable:
			ptr_map_set(MVars, e, n)
		default:
			ptr_map_set(MOther, e, n)
		}
	}

	debugf("[Section] %s\n", "generate_entity_dependency_graph: Calculate edges for graph M - Part 1")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("generate_entity_dependency_graph: Calculate edges for graph M - Part 1"))
	}

	for _, entry := range ptr_map_iterate(MProcs) {
		n := entry.Value
		e := n.Entity
		decl := decl_info_of_entity(e)
		gb_assert_handler("Assertion Failure", "decl != nil", "checker.cpp", 3187, 0)
		if decl.Deps != nil {
			for _, dep := range decl.Deps.Keys {
				if dep == nil {
					continue
				}
				if dep.Flags&EntityFlag_Field != 0 {
					continue
				}
				if !is_entity_a_dependency(dep) {
					continue
				}
				var m *EntityGraphNode
				switch dep.Kind {
				case Entity_Procedure:
					m = ptr_map_must_get(MProcs, dep)
				case Entity_Variable:
					m = ptr_map_must_get(MVars, dep)
				default:
					m = ptr_map_must_get(MOther, dep)
				}
				entity_graph_node_set_add(&n.Succ, m)
				entity_graph_node_set_add(&m.Pred, n)
			}
		}
	}

	debugf("[Section] %s\n", "generate_entity_dependency_graph: Calculate edges for graph M - Part 2a (init)")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("generate_entity_dependency_graph: Calculate edges for graph M - Part 2a (init)"))
	}

	G := make([]*EntityGraphNode, 0, ptr_map_count(MProcs)+ptr_map_count(MVars)+ptr_map_count(MOther))

	debugf("[Section] %s\n", "generate_entity_dependency_graph: Calculate edges for graph M - Part 2b (procs)")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("generate_entity_dependency_graph: Calculate edges for graph M - Part 2b (procs)"))
	}

	for _, mEntry := range ptr_map_iterate(MProcs) {
		n := mEntry.Value
		if n.Pred != nil {
			for _, p := range n.Pred.Keys {
				if p == nil || p == n {
					continue
				}
				if n.Succ != nil {
					for _, s := range n.Succ.Keys {
						if s == nil || s == n {
							continue
						}
						if p.Entity.Kind == Entity_Procedure && s.Entity.Kind == Entity_Procedure {
							continue
						}
						entity_graph_node_set_add(&p.Succ, s)
						entity_graph_node_set_add(&s.Pred, p)
						entity_graph_node_set_remove(&s.Pred, n)
					}
				}
				entity_graph_node_set_remove(&p.Succ, n)
			}
		}
	}

	debugf("[Section] %s\n", "generate_entity_dependency_graph: Calculate edges for graph M - Part 2c (vars)")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("generate_entity_dependency_graph: Calculate edges for graph M - Part 2c (vars)"))
	}

	for _, mEntry := range ptr_map_iterate(MVars) {
		n := mEntry.Value
		G = append(G, n)
	}

	debugf("[Section] %s\n", "generate_entity_dependency_graph: Dependency Count Checker")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("generate_entity_dependency_graph: Dependency Count Checker"))
	}

	for i, n := range G {
		n.Index = isize(i)
		if n.Succ != nil {
			n.DepCount = isize(len(n.Succ.Keys))
		}
		gb_assert_handler("Assertion Failure", "n->dep_count >= 0", "checker.cpp", 3263, 0)
	}
	return G
}

func find_core_entity(c *Checker, name String) *Entity {
	hash := uint32(0)
	interned := string_interner_insert(name)
	hash = interned.Hash()
	e := scope_lookup_current(c.Info.RuntimePackage.Scope, interned, hash)
	if e == nil {
		compiler_error("Could not find type declaration for '%s'\n", name)
	}
	return e
}

func find_core_type(c *Checker, name String) *Type {
	hash := uint32(0)
	interned := string_interner_insert(name)
	hash = interned.Hash()
	e := scope_lookup_current(c.Info.RuntimePackage.Scope, interned, hash)
	if e == nil {
		compiler_error("Could not find type declaration for '%s'\n", name)
	}
	if e.Type == nil {
		check_single_global_entity(c, e, e.DeclInfo)
	}
	gb_assert_handler("Assertion Failure", "e->type != nil", "checker.cpp", 3319, 0)
	return e.Type
}

func find_entity_in_pkg(info *CheckerInfo, pkg String, name String) *Entity {
	hash := uint32(0)
	interned := string_interner_insert(name)
	hash = interned.Hash()
	package_ := get_core_package(info, pkg)
	e := scope_lookup_current(package_.Scope, interned, hash)
	if e == nil {
		compiler_error("Could not find type declaration for '%s.%s'\n", pkg, name)
	}
	return e
}

func find_type_in_pkg(info *CheckerInfo, pkg String, name String) *Type {
	hash := uint32(0)
	interned := string_interner_insert(name)
	hash = interned.Hash()
	package_ := get_core_package(info, pkg)
	e := scope_lookup_current(package_.Scope, interned, hash)
	if e == nil {
		compiler_error("Could not find type declaration for '%s.%s'\n", pkg, name)
	}
	gb_assert_handler("Assertion Failure", "e->type != nil", "checker.cpp", 3345, 0)
	return e.Type
}

func new_checker_type_path() *CheckerTypePath {
	tp := atomic_freelist_get(checkerTypePathFreeList)
	if tp == nil {
		tp = permanent_alloc_item[AtomicFreelist[CheckerTypePath]]()
		tp.Value = make(CheckerTypePath, 0, 16)
	}
	return &tp.Value
}

func destroy_checker_type_path(path *CheckerTypePath) {
	tp := (*AtomicFreelist[CheckerTypePath])(unsafe.Pointer(path))
	tp.Value = nil
	atomic_freelist_put(checkerTypePathFreeList, tp)
}
