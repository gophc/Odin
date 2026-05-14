package cmd

import (
	"container/heap"
	"sort"
	"sync/atomic"
	"unsafe"
)

func find_entity_path_tuple(tuple *Type, end *Entity, allocator gbAllocator, visited *PtrSet[*Entity], path_ *[]*Entity) bool {
	if tuple == nil {
		return false
	}
	if tuple.Kind != TypeTuple {
		gb_assert_handler("Assertion Failure", "tuple.Kind == TypeTuple", "checker.cpp", 6163, 0)
	}
	for _, v := range tuple.Tuple.Variables {
		var_decl := v.DeclInfo
		if var_decl == nil {
			continue
		}
		for _, dep := range var_decl.Deps.Keys {
			if dep == nil {
				continue
			}
			if dep == end {
				path := make([]*Entity, 0, 1)
				path = append(path, dep)
				*path_ = path
				return true
			}
			next_path := find_entity_path(dep, end, allocator, visited)
			if len(next_path) > 0 {
				next_path = append(next_path, dep)
				*path_ = next_path
				return true
			}
		}
	}
	return false
}

func find_entity_path(start *Entity, end *Entity, allocator gbAllocator, visited *PtrSet[*Entity]) []*Entity {
	var visited_ PtrSet[*Entity]
	made_visited := false
	if visited == nil {
		made_visited = true
		visited = &visited_
	}
	if made_visited {
		defer ptr_set_destroy(visited)
	}
	var empty_path []*Entity
	if ptr_set_update(visited, start) {
		return empty_path
	}
	decl := start.DeclInfo
	if decl != nil {
		if start.Kind == Entity_Procedure {
			t := base_type(start.Type)
			if t.Kind != TypeProc {
				gb_assert_handler("Assertion Failure", "t.Kind == Type_Proc", "checker.cpp", 6211, 0)
			}
			path := empty_path
			if find_entity_path_tuple(t.Proc.Params, end, allocator, visited, &path) {
				return path
			}
			if find_entity_path_tuple(t.Proc.Results, end, allocator, visited, &path) {
				return path
			}
		} else {
			for _, dep := range decl.Deps.Keys {
				if dep == nil {
					continue
				}
				if dep == end {
					path := make([]*Entity, 0, 1)
					path = append(path, dep)
					return path
				}
				next_path := find_entity_path(dep, end, allocator, visited)
				if len(next_path) > 0 {
					next_path = append(next_path, dep)
					return next_path
				}
			}
		}
	}
	return empty_path
}

type entityGraphPriorityQueue struct {
	data []*EntityGraphNode
}

func (pq entityGraphPriorityQueue) Len() int { return len(pq.data) }

func (pq entityGraphPriorityQueue) Less(i, j int) bool {
	return entity_graph_node_cmp(pq.data, isize(i), isize(j)) < 0
}

func (pq entityGraphPriorityQueue) Swap(i, j int) {
	entity_graph_node_swap(pq.data, isize(i), isize(j))
}

func (pq *entityGraphPriorityQueue) Push(x any) {
	pq.data = append(pq.data, x.(*EntityGraphNode))
}

func (pq *entityGraphPriorityQueue) Pop() any {
	old := pq.data
	n := len(old)
	x := old[n-1]
	pq.data = old[:n-1]
	return x
}

func calculate_global_init_order(c *Checker) {
	info := &c.Info

	debugf("[Section] %s\n", "calculate_global_init_order: generate entity dependency graph")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("calculate_global_init_order: generate entity dependency graph"))
	}

	temporary_arena := get_arena(ThreadArena_Temporary)
	temporary_arena_guard := ArenaTempGuard(temporary_arena)
	_ = temporary_arena_guard

	dep_graph := generate_entity_dependency_graph(info, temporary_arena)

	debugf("[Section] %s\n", "calculate_global_init_order: priority queue create")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("calculate_global_init_order: priority queue create"))
	}

	pq := &entityGraphPriorityQueue{data: make([]*EntityGraphNode, len(dep_graph))}
	copy(pq.data, dep_graph)
	heap.Init(pq)

	var emitted PtrSet[*DeclInfo]
	defer ptr_set_destroy(&emitted)

	debugf("[Section] %s\n", "calculate_global_init_order: queue sort")
	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, make_string_c("calculate_global_init_order: queue sort"))
	}

	for pq.Len() > 0 {
		n := heap.Pop(pq).(*EntityGraphNode)
		e := n.Entity
		if n.DepCount > 0 {
			path := find_entity_path(e, e, temporary_allocator())
			if len(path) > 0 {
				e := path[0]
				error(e.Token, "Cyclic initialization of '%.*s'", e.Token.String.Len, e.Token.String.Data)
				for i := len(path) - 1; i >= 0; i-- {
					error(e.Token, "\t'%.*s' refers to", e.Token.String.Len, e.Token.String.Data)
					e = path[i]
				}
				error(e.Token, "\t'%.*s'", e.Token.String.Len, e.Token.String.Data)
			}
		}
		for _, p := range n.Pred.Keys {
			if p == nil {
				continue
			}
			p.DepCount -= 1
			if p.DepCount < 0 {
				p.DepCount = 0
			}
			heap.Fix(pq, int(p.Index))
		}
		d := decl_info_of_entity(e)
		if e.Kind != Entity_Variable {
			continue
		}
		if ptr_set_update(&emitted, d) {
			continue
		}
		info.VariableInitOrder = append(info.VariableInitOrder, d)
	}

	_ = false
	if false {
		gb_printf("Variable Initialization Order:\n")
		for i, d := range info.VariableInitOrder {
			e := d.Entity
			gb_printf("\t'%.*s' %llu\n", e.Token.String.Len, e.Token.String.Data, uint64(e.OrderInSrc))
			_ = i
		}
		gb_printf("\n")
	}
}

func check_procedure_later_from_entity(c *Checker, e *Entity, from_msg string) {
	if e == nil || e.Kind != Entity_Procedure {
		return
	}
	if e.Procedure.IsForeign {
		return
	}
	if e.Flags&EntityFlag_ProcBodyChecked != 0 {
		return
	}
	if e.Flags&EntityFlag_Overridden != 0 {
		if e.AliasedOf == nil {
			gb_assert_handler("Assertion Failure", "e.aliased_of != nil", "checker.cpp", 6320, 0)
		}
		if e.AliasedOf.Kind != Entity_Procedure {
			gb_assert_handler("Assertion Failure", "e.aliased_of.Kind == Entity_Procedure", "checker.cpp", 6321, 0)
		}
		if e.AliasedOf.Flags&EntityFlag_ProcBodyChecked != 0 {
			e.Flags |= EntityFlag_ProcBodyChecked
			return
		}
		check_procedure_later_full(c, e.File, e.Token, e.DeclInfo, e.Type, nil, 0)
		return
	}
	type_ := base_type(e.Type)
	if type_ == t_invalid {
		return
	}
	if type_.Kind != TypeProc {
		gb_assert_handler("Assertion Failure", "type_.Kind == Type_Proc", "checker.cpp", 6334, "%s", type_to_string(e.Type))
	}
	if is_type_polymorphic(type_) && !type_.Proc.IsPolySpecialized {
		return
	}
	if e.DeclInfo == nil {
		gb_assert_handler("Assertion Failure", "e.DeclInfo != nil", "checker.cpp", 6340, 0)
	}
	pi := permanentAllocItem[ProcInfo]()
	pi.File = e.File
	pi.Token = e.Token
	pi.Decl = e.DeclInfo
	pi.Type = e.Type
	pl := e.DeclInfo.ProcLit
	if pl == nil {
		gb_assert_handler("Assertion Failure", "pl != nil", "checker.cpp", 6349, 0)
	}
	pi.Body = pl.ProcLit.Body
	pi.Tags = pl.ProcLit.Tags
	if pi.Body == nil {
		return
	}
	if from_msg != "" {
		debugf("CHECK PROCEDURE LATER [FROM %s]! %.*s :: %s {...}\n", from_msg, e.Token.String.Len, e.Token.String.Data, type_to_string(e.Type))
	}
	check_procedure_later(c, pi)
}

func check_proc_info(c *Checker, pi *ProcInfo, untyped *UntypedExprInfoMap) bool {
	if pi == nil {
		return false
	}
	if pi.Type == nil {
		return false
	}
	if !mutex_try_lock(&pi.Decl.ProcCheckedMutex) {
		return false
	}
	defer mutex_unlock(&pi.Decl.ProcCheckedMutex)

	e := pi.Decl.Entity
	switch pi.Decl.ProcCheckedState {
	case ProcCheckedState_InProgress:
		if e != nil {
			if !globalProcedureBodyInWorkerQueue.Load() {
				gb_assert_handler("Assertion Failure", "global_procedure_body_in_worker_queue.Load() != 0", "checker.cpp", 6379, 0)
			}
		}
		return false
	case ProcCheckedState_Checked:
		if e != nil {
			if e.Flags&EntityFlag_ProcBodyChecked == 0 {
				gb_assert_handler("Assertion Failure", "e.Flags & EntityFlag_ProcBodyChecked", "checker.cpp", 6384, 0)
			}
		}
		return true
	case ProcCheckedState_Unchecked:
	}

	pi.Decl.ProcCheckedState = ProcCheckedState_InProgress

	if pi.Type.Kind != TypeProc {
		gb_assert_handler("Assertion Failure", "pi.Type.Kind == Type_Proc", "checker.cpp", 6393, 0)
	}
	pt := &pi.Type.Proc
	name := pi.Token.String
	if pt.IsPolymorphic && !pt.IsPolySpecialized {
		token := pi.Token
		if pi.PolyDefNode != nil {
			token = ast_token(pi.PolyDefNode)
		}
		error(token, "Unspecialized polymorphic procedure '%.*s'", name.Len, name.Data)
		pi.Decl.ProcCheckedState = ProcCheckedState_Unchecked
		return false
	}
	if pt.IsPolymorphic && pt.IsPolySpecialized {
		e := pi.Decl.Entity
		if e == nil {
			gb_assert_handler("Assertion Failure", "e != nil", "checker.cpp", 6409, 0)
		}
		if e.Flags&EntityFlag_Used == 0 {
			pi.Decl.ProcCheckedState = ProcCheckedState_Unchecked
			return false
		}
	}

	ctx := CheckerContext{}
	init_checker_context(&ctx, c)
	defer destroy_checker_context(&ctx)

	reset_checker_context(&ctx, pi.File, untyped)
	ctx.Decl = pi.Decl

	bounds_check := pi.Tags&ProcTag_bounds_check != 0
	no_bounds_check := pi.Tags&ProcTag_no_bounds_check != 0
	type_assert := pi.Tags&ProcTag_type_assert != 0
	no_type_assert := pi.Tags&ProcTag_no_type_assert != 0

	if bounds_check {
		ctx.StateFlags |= StateFlag_bounds_check
		ctx.StateFlags &^= StateFlag_no_bounds_check
	} else if no_bounds_check {
		ctx.StateFlags |= StateFlag_no_bounds_check
		ctx.StateFlags &^= StateFlag_bounds_check
	}
	if type_assert {
		ctx.StateFlags |= StateFlag_type_assert
		ctx.StateFlags &^= StateFlag_no_type_assert
	} else if no_type_assert {
		ctx.StateFlags |= StateFlag_no_type_assert
		ctx.StateFlags &^= StateFlag_type_assert
	}

	body_was_checked := check_proc_body(&ctx, pi.Token, pi.Decl, pi.Type, pi.Body)
	if body_was_checked {
		pi.Decl.ProcCheckedState = ProcCheckedState_Checked
		if pi.Body != nil {
			e := pi.Decl.Entity
			if e != nil {
				e.Flags |= EntityFlag_ProcBodyChecked
			}
		}
	} else {
		pi.Decl.ProcCheckedState = ProcCheckedState_Unchecked
		if pi.Body != nil {
			e := pi.Decl.Entity
			if e != nil {
				e.Flags &^= EntityFlag_ProcBodyChecked
			}
		}
	}
	add_untyped_expressions(&c.Info, ctx.Untyped)

	rw_mutex_shared_lock(&ctx.Decl.DepsMutex)
	for _, dep := range ctx.Decl.Deps.Keys {
		if dep == nil {
			continue
		}
		if dep.Kind == Entity_Procedure && dep.Flags&EntityFlag_ProcBodyChecked == 0 {
			check_procedure_later_from_entity(c, dep, "")
		}
	}
	rw_mutex_shared_unlock(&ctx.Decl.DepsMutex)

	return true
}

func check_unchecked_bodies(c *Checker) {
	if len(c.ProcsToCheck) != 0 {
		gb_assert_handler("Assertion Failure", "c.ProcsToCheck count == 0", "checker.cpp", 6490, 0)
	}
	var untyped UntypedExprInfoMap
	_ = untyped

	globalProcedureBodyInWorkerQueue.Store(false)

	for _, e := range c.Info.Entities {
		if atomic.LoadInt64(&e.MinDepCount) > 0 {
			check_procedure_later_from_entity(c, e, "check_unchecked_bodies")
		}
	}

	if !globalProcedureBodyInWorkerQueue.Load() {
		for _, pi := range c.ProcsToCheck {
			consume_proc_info(c, pi, &untyped)
		}
		c.ProcsToCheck = c.ProcsToCheck[:0]
	} else {
		thread_pool_wait()
	}

	globalProcedureBodyInWorkerQueue.Store(false)
	globalAfterCheckingProcedureBodies.Store(true)
}

func check_safety_all_procedures_for_unchecked(c *Checker) {
	_ = true

	var untyped UntypedExprInfoMap
	_ = untyped

	allProceduresCount := c.Info.AllProceduresQueue.Count.Load()
	if cap(c.Info.AllProcedures) < int(allProceduresCount) {
		newCap := int(allProceduresCount)
		if newCap < 0 {
			newCap = 0
		}
		newSlice := make([]*ProcInfo, len(c.Info.AllProcedures), newCap)
		copy(newSlice, c.Info.AllProcedures)
		c.Info.AllProcedures = newSlice
	}

	for {
		var pi *ProcInfo
		if !mpsc_dequeue(c.Info.AllProceduresQueue, &pi) {
			break
		}
		if pi == nil {
			gb_assert_handler("Assertion Failure", "pi != nil", "checker.cpp", 6527, 0)
		}
		if pi.Decl == nil {
			gb_assert_handler("Assertion Failure", "pi.Decl != nil", "checker.cpp", 6528, 0)
		}
		e := pi.Decl.Entity
		proc_checked_state := pi.Decl.ProcCheckedState
		_ = proc_checked_state
		if e != nil && e.Flags&EntityFlag_ProcBodyChecked == 0 {
			if e.Flags&EntityFlag_Used != 0 {
				consume_proc_info(c, pi, &untyped)
			}
		}
		c.Info.AllProcedures = append(c.Info.AllProcedures, pi)
	}
}

func remove_neighbouring_duplicate_entires_from_sorted_array(array *[]*Entity) {
	var prev *Entity
	i := 0
	for i < len(*array) {
		curr := (*array)[i]
		if prev == curr {
			*array = append((*array)[:i], (*array)[i+1:]...)
		} else {
			prev = curr
			i++
		}
	}
}

func check_test_procedures(c *Checker) {
	sort.Slice(c.Info.TestingProcedures, func(i, j int) bool {
		a := c.Info.TestingProcedures[i]
		b := c.Info.TestingProcedures[j]
		return a.OrderInSrc < b.OrderInSrc
	})
	remove_neighbouring_duplicate_entires_from_sorted_array(&c.Info.TestingProcedures)
}

var totalBodiesChecked atomic.Int64

func consume_proc_info(c *Checker, pi *ProcInfo, untyped *UntypedExprInfoMap) bool {
	if pi.Decl == nil {
		gb_assert_handler("Assertion Failure", "pi.Decl != nil", "checker.cpp", 6573, 0)
	}
	switch pi.Decl.ProcCheckedState {
	case ProcCheckedState_InProgress:
		return false
	case ProcCheckedState_Checked:
		return true
	}
	if pi.Decl.Parent != nil && pi.Decl.Parent.Entity != nil {
		parent := pi.Decl.Parent.Entity
		if parent.Kind == Entity_Procedure && parent.Flags&EntityFlag_ProcBodyChecked == 0 {
			check_procedure_later(c, pi)
			return false
		}
	}
	if untyped != nil {
		*untyped = nil
	}
	if check_proc_info(c, pi, untyped) {
		totalBodiesChecked.Add(1)
		return true
	}
	return false
}

func check_proc_info_worker_proc(data unsafe.Pointer) isize {
	wd := &checkProcedureBodiesWorkerData[current_thread_index()]
	untyped := &wd.Untyped
	c := wd.C
	pi := (*ProcInfo)(data)

	if pi.Decl == nil {
		gb_assert_handler("Assertion Failure", "pi.Decl != nil", "checker.cpp", 6615, 0)
	}
	if pi.Decl.Parent != nil && pi.Decl.Parent.Entity != nil {
		parent := pi.Decl.Parent.Entity
		if parent.Kind == Entity_Procedure && parent.Flags&EntityFlag_ProcBodyChecked == 0 {
			thread_pool_add_task(check_proc_info_worker_proc, data)
			return 1
		}
	}
	*untyped = nil
	if check_proc_info(c, pi, untyped) {
		totalBodiesChecked.Add(1)
		return 0
	}
	return 1
}

type CheckProcedureBodyWorkerData struct {
	C       *Checker
	Untyped UntypedExprInfoMap
}

var checkProcedureBodiesWorkerData []CheckProcedureBodyWorkerData

func check_init_worker_data(c *Checker) {
	thread_count := uint32(globalThreadPool.Threads.Count)
	checkProcedureBodiesWorkerData = make([]CheckProcedureBodyWorkerData, thread_count)
	for i := range checkProcedureBodiesWorkerData {
		checkProcedureBodiesWorkerData[i].C = c
		checkProcedureBodiesWorkerData[i].Untyped = nil
	}
}

func check_procedure_bodies(c *Checker) {
	if c == nil {
		gb_assert_handler("Assertion Failure", "c != nil", "checker.cpp", 6646, 0)
	}

	thread_count := uint32(globalThreadPool.Threads.Count)
	if buildContext.NoThreadedChecker {
		thread_count = 1
	}

	if thread_count == 1 {
		untyped := &checkProcedureBodiesWorkerData[0].Untyped
		for _, pi := range c.ProcsToCheck {
			consume_proc_info(c, pi, untyped)
		}
		c.ProcsToCheck = c.ProcsToCheck[:0]
		debugf("Total Procedure Bodies Checked: %d\n", totalBodiesChecked.Load())
		return
	}

	globalProcedureBodyInWorkerQueue.Store(true)
	prev_procs_to_check_count := len(c.ProcsToCheck)

	for _, pi := range c.ProcsToCheck {
		thread_pool_add_task(check_proc_info_worker_proc, unsafe.Pointer(pi))
	}

	if prev_procs_to_check_count != len(c.ProcsToCheck) {
		gb_assert_handler("Assertion Failure", "prev_procs_to_check_count == len(c.ProcsToCheck)", "checker.cpp", 6670, 0)
	}
	c.ProcsToCheck = c.ProcsToCheck[:0]
	thread_pool_wait()
	globalProcedureBodyInWorkerQueue.Store(false)
}

func add_untyped_expressions(cinfo *CheckerInfo, untyped *UntypedExprInfoMap) {
	if untyped == nil {
		return
	}
	for _, entry := range ptr_map_iterate(*untyped) {
		expr := entry.Key
		info := entry.Value
		if expr != nil && info != nil {
			mpsc_enqueue(&cinfo.Checker.GlobalUntypedQueue, UntypedExprInfo{Expr: expr, Info: info})
		}
	}
	*untyped = nil
}

func tuple_to_pointers(ot *Type) *Type {
	if ot == nil {
		return nil
	}
	if ot.Kind != TypeTuple {
		gb_assert_handler("Assertion Failure", "ot.Kind == Type_Tuple", "checker.cpp", 6695, 0)
	}
	t := alloc_type_tuple()
	t.Tuple.Variables = make([]*Entity, len(ot.Tuple.Variables))
	var scope *Scope
	for i, e := range ot.Tuple.Variables {
		t.Tuple.Variables[i] = alloc_entity_variable(scope, e.Token, alloc_type_pointer(e.Type))
	}
	t.Tuple.IsPacked = ot.Tuple.IsPacked
	return t
}

func check_deferred_procedures(c *Checker) {
	for {
		var src *Entity
		if !mpsc_dequeue(c.ProcsWithDeferredToCheck, &src) {
			break
		}
		if src.Kind != Entity_Procedure {
			gb_assert_handler("Assertion Failure", "src.Kind == Entity_Procedure", "checker.cpp", 6713, 0)
		}
		dst_kind := src.Procedure.DeferredProcedure.Kind
		dst := src.Procedure.DeferredProcedure.Entity
		if dst == nil {
			gb_assert_handler("Assertion Failure", "dst != nil", "checker.cpp", 6717, 0)
		}
		if dst.Kind != Entity_Procedure {
			gb_assert_handler("Assertion Failure", "dst.Kind == Entity_Procedure", "checker.cpp", 6718, 0)
		}
		attribute := "deferred_none"
		switch dst_kind {
		case DeferredProcedure_none:
			attribute = "deferred_none"
		case DeferredProcedure_in:
			attribute = "deferred_in"
		case DeferredProcedure_out:
			attribute = "deferred_out"
		case DeferredProcedure_in_out:
			attribute = "deferred_in_out"
		case DeferredProcedure_in_by_ptr:
			attribute = "deferred_in_by_ptr"
		case DeferredProcedure_out_by_ptr:
			attribute = "deferred_out_by_ptr"
		case DeferredProcedure_in_out_by_ptr:
			attribute = "deferred_in_out_by_ptr"
		}
		if src == dst {
			error(src.Token, "'%.*s' cannot be used as its own %s", dst.Token.String.Len, dst.Token.String.Data, attribute)
			continue
		}
		if is_type_polymorphic(src.Type) || is_type_polymorphic(dst.Type) {
			error(src.Token, "'%s' cannot be used with a polymorphic procedure", attribute)
			continue
		}
		if dst.Flags&EntityFlag_Disabled != 0 {
			src.Procedure.DeferredProcedure = DeferredProcedure{}
			continue
		}
		if !is_type_proc(src.Type) {
			gb_assert_handler("Assertion Failure", "is_type_proc(src.Type)", "checker.cpp", 6747, 0)
		}
		if !is_type_proc(dst.Type) {
			gb_assert_handler("Assertion Failure", "is_type_proc(dst.Type)", "checker.cpp", 6748, 0)
		}

		src_params := base_type(src.Type).Proc.Params
		src_results := base_type(src.Type).Proc.Results
		dst_params := base_type(dst.Type).Proc.Params
		var by_ptr bool

		switch dst_kind {
		case DeferredProcedure_in_by_ptr:
			by_ptr = true
			src_params = tuple_to_pointers(src_params)
		case DeferredProcedure_out_by_ptr:
			by_ptr = true
			src_results = tuple_to_pointers(src_results)
		case DeferredProcedure_in_out_by_ptr:
			by_ptr = true
			src_params = tuple_to_pointers(src_params)
			src_results = tuple_to_pointers(src_results)
		}
		_ = by_ptr

		switch dst_kind {
		case DeferredProcedure_none:
			if dst_params == nil {
				continue
			}
			error(src.Token, "Deferred procedure '%.*s' must have no input parameters", dst.Token.String.Len, dst.Token.String.Data)

		case DeferredProcedure_in, DeferredProcedure_in_by_ptr:
			if src_params == nil && dst_params == nil {
				continue
			}
			if (src_params == nil && dst_params != nil) ||
				(src_params != nil && dst_params == nil) {
				error(src.Token, "Deferred procedure '%.*s' parameters do not match the inputs of initial procedure '%.*s'",
					dst.Token.String.Len, dst.Token.String.Data, src.Token.String.Len, src.Token.String.Data)
				continue
			}
			if src_params.Kind != TypeTuple {
				gb_assert_handler("Assertion Failure", "src_params.Kind == Type_Tuple", "checker.cpp", 6793, 0)
			}
			if dst_params.Kind != TypeTuple {
				gb_assert_handler("Assertion Failure", "dst_params.Kind == Type_Tuple", "checker.cpp", 6794, 0)
			}
			if !are_types_identical(src_params, dst_params) {
				s := type_to_string(src_params)
				d := type_to_string(dst_params)
				error(src.Token, "Deferred procedure '%.*s' parameters do not match the inputs of initial procedure '%.*s':\n\t(%s) =/= (%s)",
					dst.Token.String.Len, dst.Token.String.Data, src.Token.String.Len, src.Token.String.Data,
					d, s)
				continue
			}

		case DeferredProcedure_out, DeferredProcedure_out_by_ptr:
			if src_results == nil && dst_params == nil {
				continue
			}
			if (src_results == nil && dst_params != nil) ||
				(src_results != nil && dst_params == nil) {
				error(src.Token, "Deferred procedure '%.*s' parameters do not match the results of initial procedure '%.*s'",
					dst.Token.String.Len, dst.Token.String.Data, src.Token.String.Len, src.Token.String.Data)
				continue
			}
			if src_results.Kind != TypeTuple {
				gb_assert_handler("Assertion Failure", "src_results.Kind == Type_Tuple", "checker.cpp", 6823, 0)
			}
			if dst_params.Kind != TypeTuple {
				gb_assert_handler("Assertion Failure", "dst_params.Kind == Type_Tuple", "checker.cpp", 6824, 0)
			}
			if !are_types_identical(src_results, dst_params) {
				s := type_to_string(src_results)
				d := type_to_string(dst_params)
				error(src.Token, "Deferred procedure '%.*s' parameters do not match the results of initial procedure '%.*s':\n\t(%s) =/= (%s)",
					dst.Token.String.Len, dst.Token.String.Data, src.Token.String.Len, src.Token.String.Data,
					d, s)
				continue
			}

		case DeferredProcedure_in_out, DeferredProcedure_in_out_by_ptr:
			if src_params == nil && src_results == nil && dst_params == nil {
				continue
			}
			if dst_params == nil {
				error(src.Token, "Deferred procedure must have parameters for %s", attribute)
				continue
			}
			if dst_params.Kind != TypeTuple {
				gb_assert_handler("Assertion Failure", "dst_params.Kind == Type_Tuple", "checker.cpp", 6852, 0)
			}

			tsrc := alloc_type_tuple()
			sv := &tsrc.Tuple.Variables
			dv := dst_params.Tuple.Variables
			_ = dv

			len_ := isize(0)
			if src_params != nil {
				if src_params.Kind != TypeTuple {
					gb_assert_handler("Assertion Failure", "src_params.Kind == Type_Tuple", "checker.cpp", 6861, 0)
				}
				len_ += isize(len(src_params.Tuple.Variables))
			}
			if src_results != nil {
				if src_results.Kind != TypeTuple {
					gb_assert_handler("Assertion Failure", "src_results.Kind == Type_Tuple", "checker.cpp", 6865, 0)
				}
				len_ += isize(len(src_results.Tuple.Variables))
			}
			*sv = make([]*Entity, len_)
			offset := isize(0)
			if src_params != nil {
				for _, v := range src_params.Tuple.Variables {
					(*sv)[offset] = v
					offset++
				}
			}
			if src_results != nil {
				for _, v := range src_results.Tuple.Variables {
					(*sv)[offset] = v
					offset++
				}
			}
			if offset != len_ {
				gb_assert_handler("Assertion Failure", "offset == len_", "checker.cpp", 6880, 0)
			}

			if !are_types_identical(tsrc, dst_params) {
				s := type_to_string(tsrc)
				d := type_to_string(dst_params)
				error(src.Token, "Deferred procedure '%.*s' parameters do not match the results of initial procedure '%.*s':\n\t(%s) =/= (%s)",
					dst.Token.String.Len, dst.Token.String.Data, src.Token.String.Len, src.Token.String.Data,
					d, s)
				continue
			}
		}
	}
}

func check_objc_context_provider_procedures(c *Checker) {
	for {
		var e *Entity
		if !mpsc_dequeue(c.ProcsWithObjcContextProviderToCheck, &e) {
			break
		}
		if e.Kind != Entity_TypeName {
			gb_assert_handler("Assertion Failure", "e.Kind == Entity_TypeName", "checker.cpp", 7169, 0)
		}
		proc_entity := e.TypeName.ObjcContextProvider
		if proc_entity.Kind != Entity_Procedure {
			gb_assert_handler("Assertion Failure", "proc_entity.Kind == Entity_Procedure", "checker.cpp", 7172, 0)
		}

		proc := &base_type(proc_entity.Type).Proc
		return_type := t_untyped_nil
		if proc.ResultCount == 1 {
			return_type = base_named_type(proc.Results.Tuple.Variables[0].Type)
		}
		if return_type != t_context {
			error(proc_entity.Token, "The @(objc_context_provider) procedure must only return a context.")
		}

		self_param_err := "The @(objc_context_provider) procedure must take as a parameter a single pointer to the @(objc_type) value."
		if proc.ParamCount != 1 {
			error(proc_entity.Token, self_param_err)
		}
		self_param := base_type(proc.Params.Tuple.Variables[0].Type)
		if self_param.Kind != TypePointer {
			error(proc_entity.Token, self_param_err)
		}
		self_type := base_named_type(self_param.Pointer.Elem)
		if !internal_check_is_assignable_to(self_type, e.Type) &&
			!(e.TypeName.ObjcIvar != nil && internal_check_is_assignable_to(self_type, e.TypeName.ObjcIvar)) {
			error(proc_entity.Token, self_param_err)
		}
		if proc.CallingConvention != ProcCC_CDecl && proc.CallingConvention != ProcCC_Contextless {
			error(e.Token, self_param_err)
		}
		if proc.IsPolymorphic {
			error(e.Token, self_param_err)
		}
	}
}
