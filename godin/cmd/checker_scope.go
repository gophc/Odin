package cmd

import (
	"sync/atomic"
	"unsafe"
)

func scope_map_max_load(cap u32) u32 {
	return cap - (cap >> 2)
}

func scope_map_init(m *ScopeMap) {
	m.cap = SCOPE_MAP_INLINE_CAP
	m.slots = m.inline_slots[:]
	m.keys = m.inline_keys[:]
}

func scope_map_insert_for_rehash(keys []InternedString, slots []ScopeMapSlot, mask u32, key InternedString, hash u32, value *Entity) *Entity {
	pos := hash & mask
	dist := u32(0)
	for {
		s := &slots[pos]
		if s.hash == 0 {
			keys[pos] = key
			s.hash = hash
			s.value = value
			return nil
		}
		existing_dist := (pos - s.hash) & mask
		if dist > existing_dist {
			tmp_key := keys[pos]
			tmp_hash := s.hash
			tmp_value := s.value
			keys[pos] = key
			s.hash = hash
			s.value = value
			hash = tmp_hash
			value = tmp_value
			key = tmp_key
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_allocate_entries(cap u32) ([]InternedString, []ScopeMapSlot) {
	entrySize := isize(int(unsafe.Sizeof(InternedString{})) + int(unsafe.Sizeof(ScopeMapSlot{})))
	totalSize := entrySize * isize(cap)
	data := arena_alloc(get_arena(ThreadArena_Permanent), totalSize, 8)
	keys := unsafe.Slice((*InternedString)(data), cap)
	slotsPtr := (*ScopeMapSlot)(unsafe.Add(data, isize(cap)*isize(unsafe.Sizeof(InternedString{}))))
	slots := unsafe.Slice(slotsPtr, cap)
	return keys, slots
}

func scope_map_grow(m *ScopeMap) {
	new_cap := m.cap << 1
	new_mask := new_cap - 1
	new_keys, new_slots := scope_map_allocate_entries(new_cap)
	if m.count > 0 {
		for i := u32(0); i < m.cap; i++ {
			if m.slots[i].hash != 0 {
				scope_map_insert_for_rehash(new_keys, new_slots, new_mask, m.keys[i], m.slots[i].hash, m.slots[i].value)
			}
		}
	}
	m.slots = new_slots
	m.keys = new_keys
	m.cap = new_cap
}

func scope_map_reserve(m *ScopeMap, capacity isize) {
	if m.slots == nil {
		scope_map_init(m)
	}
	new_cap := next_pow2_u32(u32(capacity))
	if m.cap < new_cap && new_cap > SCOPE_MAP_INLINE_CAP {
		m.keys, m.slots = scope_map_allocate_entries(new_cap)
		m.cap = new_cap
	}
}

func scope_map_insert(m *ScopeMap, key InternedString, hash u32, value *Entity) *Entity {
	if m.slots == nil {
		scope_map_init(m)
	}
	if m.count >= scope_map_max_load(m.cap) {
		scope_map_grow(m)
	}
	mask := m.cap - 1
	pos := hash & mask
	dist := u32(0)
	for {
		s := &m.slots[pos]
		if s.hash == 0 {
			m.keys[pos] = key
			s.hash = hash
			s.value = value
			m.count++
			return nil
		}
		if s.hash == hash && m.keys[pos] == key {
			old := s.value
			s.value = value
			return old
		}
		existing_dist := (pos - s.hash) & mask
		if dist > existing_dist {
			tmp_key := m.keys[pos]
			tmp_hash := s.hash
			tmp_value := s.value
			m.keys[pos] = key
			s.hash = hash
			s.value = value
			key = tmp_key
			hash = tmp_hash
			value = tmp_value
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_get(m *ScopeMap, key InternedString, hash u32) *Entity {
	if m.slots == nil {
		return nil
	}
	mask := m.cap - 1
	pos := hash & mask
	dist := u32(0)
	for {
		s := &m.slots[pos]
		curr_hash := s.hash
		if curr_hash == 0 {
			return nil
		}
		existing_dist := (pos - curr_hash) & mask
		if dist > existing_dist {
			return nil
		}
		if curr_hash == hash && m.keys[pos] == key {
			return s.value
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_clear(m *ScopeMap) {
	for i := range m.cap {
		m.slots[i].hash = 0
		m.slots[i].value = nil
	}
	m.count = 0
}

func beginScopeMap(m *ScopeMap) ScopeMapIterator {
	if m.count == 0 {
		return endScopeMap(m)
	}
	index := u32(0)
	for index < m.cap {
		if m.slots[index].hash != 0 {
			break
		}
		index++
	}
	return ScopeMapIterator{m: m, index: index}
}

func endScopeMap(m *ScopeMap) ScopeMapIterator {
	return ScopeMapIterator{m: m, index: m.cap}
}

func ScopeMapIteratorNext(it *ScopeMapIterator) (InternedString, *Entity, bool) {
	if it.index >= it.m.cap {
		return InternedString{}, nil, false
	}
	key := it.m.keys[it.index]
	value := it.m.slots[it.index].value
	it.index++
	for it.index < it.m.cap {
		if it.m.slots[it.index].hash != 0 {
			break
		}
		it.index++
	}
	return key, value, true
}

func scope_reserve(scope *Scope, count isize) {
	scope_map_reserve(&scope.elements, 2*count)
}

func entity_graph_node_set_destroy(s *EntityGraphNodeSet) {
	ptr_set_destroy(s)
}

func entity_graph_node_set_add(s *EntityGraphNodeSet, n *EntityGraphNode) {
	ptr_set_add(s, n)
}

func entity_graph_node_set_remove(s *EntityGraphNodeSet, n *EntityGraphNode) {
	ptr_set_remove(s, n)
}

func entity_graph_node_destroy(n *EntityGraphNode) {
	entity_graph_node_set_destroy(&n.pred)
	entity_graph_node_set_destroy(&n.succ)
}

func entity_graph_node_cmp(data []*EntityGraphNode, i, j isize) int {
	x := data[i]
	y := data[j]
	a := x.entity.order_in_src
	b := y.entity.order_in_src
	if x.dep_count < y.dep_count {
		return -1
	}
	if x.dep_count == y.dep_count {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}
	return 1
}

func entity_graph_node_swap(data []*EntityGraphNode, i, j isize) {
	x := data[i]
	y := data[j]
	data[i] = y
	data[j] = x
	x.index = j
	y.index = i
}

func import_graph_node_set_destroy(s *ImportGraphNodeSet) {
	ptr_set_destroy(s)
}

func import_graph_node_set_add(s *ImportGraphNodeSet, n *ImportGraphNode) {
	ptr_set_add(s, n)
}

func import_graph_node_create(pkg *AstPackage) *ImportGraphNode {
	n := permanent_alloc_item[*ImportGraphNode]()
	n.pkg = pkg
	n.scope = pkg.scope
	return n
}

func import_graph_node_destroy(n *ImportGraphNode) {
	import_graph_node_set_destroy(&n.pred)
	import_graph_node_set_destroy(&n.succ)
}

func import_graph_node_cmp(data []*ImportGraphNode, i, j isize) int {
	x := data[i]
	y := data[j]
	gb_assert_handler("Assertion Failure", "x != y", "checker.cpp", 143, "")
	gb_assert_handler("Assertion Failure", "x.scope != y.scope", "checker.cpp", 145, "")
	xg := (x.scope.flags & ScopeFlag_Global) != 0
	yg := (y.scope.flags & ScopeFlag_Global) != 0
	if xg != yg {
		if xg {
			return -1
		}
		return 1
	}
	if xg && yg {
		if x.pkg.id < y.pkg.id {
			return 1
		}
		return -1
	}
	if x.dep_count < y.dep_count {
		return -1
	}
	if x.dep_count > y.dep_count {
		return 1
	}
	return 0
}

func import_graph_node_swap(data []*ImportGraphNode, i, j isize) {
	x := data[i]
	y := data[j]
	data[i] = y
	data[j] = x
	x.index = j
	y.index = i
}

func init_decl_info(d *DeclInfo, scope *Scope, parent *DeclInfo) {
	if parent != nil {
		mutex_lock(&parent.next_mutex)
		d.next_sibling = parent.next_child
		parent.next_child = d
		mutex_unlock(&parent.next_mutex)
	}
	d.parent = parent
	d.scope = scope
	ptr_set_init(&d.deps, 0)
	type_set_init(&d.type_info_deps, 0)
	d.labels.allocator = heap_allocator()
	d.variadic_reuses.allocator = heap_allocator()
	d.variadic_reuse_max_bytes = 0
	d.variadic_reuse_max_align = 1
}

func make_decl_info(scope *Scope, parent *DeclInfo) *DeclInfo {
	d := permanent_alloc_item[*DeclInfo]()
	init_decl_info(d, scope, parent)
	return d
}

func create_scope(info *CheckerInfo, parent *Scope) *Scope {
	s := permanent_alloc_item[*Scope]()
	scope_map_init(&s.elements)
	s.parent = parent
	if parent != nil && parent != builtin_pkg.scope {
		prev_head_child := (*Scope)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&parent.head_child)), unsafe.Pointer(s)))
		if prev_head_child != nil {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&s.next)), unsafe.Pointer(prev_head_child))
		}
	}
	if parent != nil && parent.flags&ScopeFlag_ContextDefined != 0 {
		s.flags |= ScopeFlag_ContextDefined
	}
	return s
}

func create_scope_from_file(info *CheckerInfo, f *AstFile) *Scope {
	gb_assert_handler("Assertion Failure", "f != nil", "checker.cpp", 236, "")
	gb_assert_handler("Assertion Failure", "f.pkg != nil", "checker.cpp", 237, "")
	gb_assert_handler("Assertion Failure", "f.pkg.scope != nil", "checker.cpp", 238, "")
	init_elements_capacity := isize(DEFAULT_SCOPE_CAPACITY)
	if init_elements_capacity < 2*isize(f.total_file_decl_count) {
		init_elements_capacity = 2 * isize(f.total_file_decl_count)
	}
	s := create_scope(info, f.pkg.scope)
	scope_reserve(s, init_elements_capacity)
	s.flags |= ScopeFlag_File
	s.file = f
	f.scope = s
	return s
}

func create_scope_from_package(c *CheckerContext, pkg *AstPackage) *Scope {
	gb_assert_handler("Assertion Failure", "pkg != nil", "checker.cpp", 252, "")
	total_pkg_decl_count := isize(0)
	for _, file := range pkg.files {
		total_pkg_decl_count += isize(file.total_file_decl_count)
	}
	s := create_scope(c.info, builtin_pkg.scope)
	scope_map_reserve(&s.elements, 2*u32(total_pkg_decl_count))
	s.flags |= ScopeFlag_Pkg
	s.pkg = pkg
	pkg.scope = s
	if pkg.fullpath == c.checker.parser.init_fullpath || pkg.kind == Package_Init {
		s.flags |= ScopeFlag_Init
	}
	if pkg.kind == Package_Runtime {
		s.flags |= ScopeFlag_Global
	}
	if s.flags&(ScopeFlag_Init|ScopeFlag_Global) != 0 {
		s.flags |= ScopeFlag_HasBeenImported
	}
	s.flags |= ScopeFlag_ContextDefined
	return s
}

func destroy_scope(scope *Scope) {
	for child := scope.head_child; child != nil; child = child.next {
		destroy_scope(child)
	}
	ptr_set_destroy(&scope.imported)
}

func add_scope(c *CheckerContext, node *Ast, scope *Scope) {
	gb_assert_handler("Assertion Failure", "node != nil", "checker.cpp", 297, "")
	gb_assert_handler("Assertion Failure", "scope != nil", "checker.cpp", 298, "")
	scope.node = node
	switch node.kind {
	case Ast_BlockStmt:
		node.BlockStmt.scope = scope
	case Ast_IfStmt:
		node.IfStmt.scope = scope
	case Ast_ForStmt:
		node.ForStmt.scope = scope
	case Ast_RangeStmt:
		node.RangeStmt.scope = scope
	case Ast_UnrollRangeStmt:
		node.UnrollRangeStmt.scope = scope
	case Ast_CaseClause:
		node.CaseClause.scope = scope
	case Ast_SwitchStmt:
		node.SwitchStmt.scope = scope
	case Ast_TypeSwitchStmt:
		node.TypeSwitchStmt.scope = scope
	case Ast_ProcType:
		node.ProcType.scope = scope
	case Ast_StructType:
		node.StructType.scope = scope
	case Ast_UnionType:
		node.UnionType.scope = scope
	case Ast_EnumType:
		node.EnumType.scope = scope
	case Ast_BitFieldType:
		node.BitFieldType.scope = scope
	default:
		gb_assert_handler("Panic", "0", "checker.cpp", 314, "Invalid node for add_scope: "+ast_strings[node.kind].text)
	}
}

func scope_of_node(node *Ast) *Scope {
	if node == nil {
		return nil
	}
	switch node.kind {
	case Ast_BlockStmt:
		return node.BlockStmt.scope
	case Ast_IfStmt:
		return node.IfStmt.scope
	case Ast_ForStmt:
		return node.ForStmt.scope
	case Ast_RangeStmt:
		return node.RangeStmt.scope
	case Ast_UnrollRangeStmt:
		return node.UnrollRangeStmt.scope
	case Ast_CaseClause:
		return node.CaseClause.scope
	case Ast_SwitchStmt:
		return node.SwitchStmt.scope
	case Ast_TypeSwitchStmt:
		return node.TypeSwitchStmt.scope
	case Ast_ProcType:
		return node.ProcType.scope
	case Ast_StructType:
		return node.StructType.scope
	case Ast_UnionType:
		return node.UnionType.scope
	case Ast_EnumType:
		return node.EnumType.scope
	case Ast_BitFieldType:
		return node.BitFieldType.scope
	}
	gb_assert_handler("Panic", "0", "checker.cpp", 337, "Invalid node for add_scope: "+ast_strings[node.kind].text)
	return nil
}

func check_open_scope(c *CheckerContext, node *Ast) {
	node = UnparenExpr(node)
	gb_assert_handler("Assertion Failure", "node != nil", "checker.cpp", 344, "")
	gb_assert_handler("Assertion Failure", "node.kind == Ast_Invalid || is_ast_stmt(node) || is_ast_type(node)", "checker.cpp", 347, "")
	scope := create_scope(c.info, c.scope)
	add_scope(c, node, scope)
	switch node.kind {
	case Ast_ProcType:
		scope.flags |= ScopeFlag_Proc
	case Ast_StructType, Ast_EnumType, Ast_UnionType, Ast_BitSetType, Ast_BitFieldType:
		scope.flags |= ScopeFlag_Type
	}
	if c.decl != nil && c.decl.proc_lit != nil {
		scope.index = c.decl.scope_index
		c.decl.scope_index++
	}
	c.scope = scope
	c.state_flags |= StateFlag_bounds_check
}

func check_close_scope(c *CheckerContext) {
	c.scope = c.scope.parent
}

func scope_lookup_current(s *Scope, name InternedString, hash u32) *Entity {
	if hash == 0 {
		hash = name.Hash()
	}
	found := scope_map_get(&s.elements, name, hash)
	if found != nil {
		return found
	}
	return nil
}

func scope_lookup_parent(scope *Scope, name InternedString, scope_ **Scope, entity_ **Entity, hash u32) *Entity {
	is_single_threaded := atomic.LoadUint32((*uint32)(unsafe.Pointer(&in_single_threaded_checker_stage))) != 0
	if scope != nil {
		gone_thru_proc := false
		gone_thru_package := false
		if hash == 0 {
			hash = name.Hash()
		}
		for s := scope; s != nil; s = s.parent {
			var found *Entity
			if !is_single_threaded {
				rw_mutex_shared_lock(&s.mutex)
			}
			found = scope_map_get(&s.elements, name, hash)
			if !is_single_threaded {
				rw_mutex_shared_unlock(&s.mutex)
			}
			if found != nil {
				e := found
				if gone_thru_proc {
					if e.kind == Entity_Label {
						continue
					}
					if e.kind == Entity_Variable {
						if e.scope.flags&ScopeFlag_File != 0 {
						} else if e.flags&EntityFlag_Static != 0 {
						} else {
							continue
						}
					}
				}
				if entity_ != nil {
					*entity_ = e
				}
				if scope_ != nil {
					*scope_ = s
				}
				return e
			}
			if s.flags&ScopeFlag_Proc != 0 {
				gone_thru_proc = true
			}
			if s.flags&ScopeFlag_Pkg != 0 {
				gone_thru_package = true
			}
		}
	}
	if entity_ != nil {
		*entity_ = nil
	}
	if scope_ != nil {
		*scope_ = nil
	}
	return nil
}

func scope_lookup(s *Scope, interned InternedString, hash u32) *Entity {
	var entity *Entity
	scope_lookup_parent(s, interned, nil, &entity, hash)
	return entity
}

func scope_insert_with_name_no_mutex(s *Scope, name InternedString, hash u32, entity *Entity) *Entity {
	if name.value == 0 {
		return nil
	}
	var found *Entity
	var result *Entity
	found = scope_map_get(&s.elements, name, hash)
	if found != nil {
		if entity != found {
			result = found
		}
		goto end
	}
	if s.parent != nil && s.parent.flags&ScopeFlag_Proc != 0 {
		found = scope_map_get(&s.parent.elements, name, hash)
		if found != nil {
			if found.flags&EntityFlag_Result != 0 {
				if entity != found {
					result = found
				}
				goto end
			}
		}
	}
	scope_map_insert(&s.elements, name, hash, entity)
	if entity.scope == nil {
		entity.scope = s
	}
end:
	return result
}

func scope_insert_with_name(s *Scope, name InternedString, hash u32, entity *Entity) *Entity {
	if name.value == 0 {
		return nil
	}
	var found *Entity
	var result *Entity
	rw_mutex_lock(&s.mutex)
	found = scope_map_get(&s.elements, name, hash)
	if found != nil {
		if entity != found {
			result = found
		}
		goto end
	}
	if s.parent != nil && s.parent.flags&ScopeFlag_Proc != 0 {
		rw_mutex_shared_lock(&s.parent.mutex)
		found = scope_map_get(&s.parent.elements, name, hash)
		if found != nil {
			if found.flags&EntityFlag_Result != 0 {
				if entity != found {
					result = found
				}
				rw_mutex_shared_unlock(&s.parent.mutex)
				goto end
			}
		}
		rw_mutex_shared_unlock(&s.parent.mutex)
	}
	scope_map_insert(&s.elements, name, hash, entity)
	if entity.scope == nil {
		entity.scope = s
	}
end:
	rw_mutex_unlock(&s.mutex)
	return result
}

func scope_insert(s *Scope, entity *Entity) *Entity {
	name := entity_interned_name(entity)
	hash := atomic.LoadUint32(&entity.interned_name_hash)
	gb_assert_handler("Assertion Failure", "hash != 0", "checker.cpp", 524, "")
	if atomic.LoadUint32((*uint32)(unsafe.Pointer(&in_single_threaded_checker_stage))) != 0 {
		return scope_insert_with_name_no_mutex(s, name, hash, entity)
	}
	return scope_insert_with_name(s, name, hash, entity)
}

func scope_insert_no_mutex(s *Scope, entity *Entity) *Entity {
	name := string_interner_insert(entity.token.string)
	hash := atomic.LoadUint32(&entity.interned_name_hash)
	gb_assert_handler("Assertion Failure", "hash != 0", "checker.cpp", 535, "")
	return scope_insert_with_name_no_mutex(s, name, hash, entity)
}
