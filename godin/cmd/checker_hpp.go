// Depends on: Type, Entity, Scope, DeclInfo, AstFile, Ast, AstPackage, Token, TokenPos, CommentGroup
// Depends on: Checker, CheckerInfo, CheckerContext, Parser
// Depends on: ExactValue, AddressingMode, TypeAndValue (from exact_value.odin)
// Depends on: String, isize, i64, u32, u64, u8, f64, i32
// Depends on: StringMap, StringSet, PtrMap, PtrSet (custom containers)
// Depends on: BlockingMutex, RwMutex, RecursiveMutex, RWSpinLock
// Depends on: MPSCQueue
// Depends on: Arena, arena_alloc, get_arena, ThreadArena_Permanent
// Depends on: gb_alloc, gb_assert_handler, gb_memset, permanent_allocator, heap_allocator
// Depends on: map_get, map_set, map_must_get (generic)
// Depends on: next_pow2_u32, array_make, array_add, Array
// Depends on: string_set_add, string_set_init
// Depends on: lbModule
// Depends on: BuiltinProcId (from checker_builtin_procs.i.hpp)
// Depends on: type_hash_canonical_type, type_to_canonical_string, temp_canonical_string
// Depends on: write_type_to_canonical_string, write_canonical_entity_name, type_info_pair_cmp
// Depends on: ProcCallingConvention, TargetOs
package cmd

type ExprInfo struct {
	Mode   AddressingMode
	IsLhs  bool
	Type   *Type
	Value  ExactValue
}

func make_expr_info(mode AddressingMode, type_ *Type, value ExactValue, is_lhs bool) *ExprInfo {
	ei := new(ExprInfo)
	ei.Mode = mode
	ei.Type = type_
	ei.Value = value
	ei.IsLhs = is_lhs
	return ei
}

type ExprKind int

const (
	ExprExpr ExprKind = iota
	ExprStmt
)

type StmtFlag int

const (
	StmtBreakAllowed       StmtFlag = 1 << 0
	StmtContinueAllowed    StmtFlag = 1 << 1
	StmtFallthroughAllowed StmtFlag = 1 << 2
	StmtTypeSwitch         StmtFlag = 1 << 4
	StmtCheckScopeDecls    StmtFlag = 1 << 5
)

type BuiltinProcPkg int

const (
	BuiltinProcPkgBuiltin    BuiltinProcPkg = iota
	BuiltinProcPkgIntrinsics
	BuiltinProcPkgCOUNT
)

var builtin_proc_pkg_name = [BuiltinProcPkgCOUNT]String{
	{Data: unsafe.StringData("builtin"), Len: 7},
	{Data: unsafe.StringData("intrinsics"), Len: 10},
}

type BuiltinProc struct {
	Name          String
	ArgCount      isize
	Variadic      bool
	Kind          ExprKind
	Pkg           BuiltinProcPkg
	Diverging     bool
	IgnoreResults bool
}

type Operand struct {
	Mode       AddressingMode
	Type       *Type
	Value      ExactValue
	Expr       *Ast
	BuiltinID  BuiltinProcId
	ProcGroup  *Entity
}

type BlockLabel struct {
	Name  String
	Label *Ast
}

type DeferredProcedureKind int

const (
	DeferredProcedureNone     DeferredProcedureKind = iota
	DeferredProcedureIn
	DeferredProcedureOut
	DeferredProcedureInOut
	DeferredProcedureInByPtr
	DeferredProcedureOutByPtr
	DeferredProcedureInOutByPtr
)

type DeferredProcedure struct {
	Kind   DeferredProcedureKind
	Entity *Entity
}

type InstrumentationFlag int

const (
	InstrumentationEnabled  InstrumentationFlag = -1
	InstrumentationDefault  InstrumentationFlag = 0
	InstrumentationDisabled InstrumentationFlag = +1
)

type AttributeContext struct {
	LinkName                  String
	LinkPrefix                String
	LinkSuffix                String
	LinkSection               String
	Linkage                   String
	InitExprListCount         isize
	ThreadLocalModel          String
	DeprecatedMessage         String
	WarningMessage            String
	DeferredProcedure         DeferredProcedure
	IsExport                  bool
	IsStatic                  bool
	RequireResults            bool
	RequireDeclaration        bool
	HasDisabledProc           bool
	DisabledProc              bool
	Test                      bool
	Init                      bool
	Fini                      bool
	SetCold                   bool
	EntryPointOnly            bool
	InstrumentationEnter      bool
	InstrumentationExit       bool
	NoSanitizeAddress         bool
	NoSanitizeMemory          bool
	NoSanitizeThread          bool
	Rodata                    bool
	IgnoreDuplicates          bool
	OptimizationMode          u32
	ForeignImportPriorityIndex i64
	ExtraLinkerFlags          String
	NoInstrumentation         InstrumentationFlag
	ObjcClass                 String
	ObjcName                  String
	ObjcSelector              String
	ObjcType                  *Type
	ObjcSuperclass            *Type
	ObjcIvar                  *Type
	ObjcContextProvider       *Entity
	ObjcIsClassMethod         bool
	ObjcIsImplementation      bool
	ObjcIsDisabledImplement   bool
	RequireTargetFeature      String
	EnableTargetFeature       String
	RaddbgTypeView            bool
	RaddbgTypeViewString      String
}

func make_attribute_context(link_prefix String, link_suffix String) AttributeContext {
	return AttributeContext{
		LinkPrefix: link_prefix,
		LinkSuffix: link_suffix,
	}
}

type DeclAttributeProc func(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool

func check_decl_attributes(c *CheckerContext, attributes []*Ast, proc DeclAttributeProc, ac *AttributeContext) {
}

type TypeWriter struct{}

func write_type_to_canonical_string(w *TypeWriter, type_ *Type)     {}
func write_canonical_entity_name(w *TypeWriter, e *Entity)          {}
func type_hash_canonical_type(type_ *Type) u64                      { return 0 }
func type_to_canonical_string(allocator gbAllocator, type_ *Type) String { return String{} }
func temp_canonical_string(type_ *Type) gbString                    { return nil }

const TYPE_SET_TOMBSTONE u64 = ^u64(0)

type TypeInfoPair struct {
	Type *Type
	Hash u64
}

type TypeSet struct {
	Keys     []TypeInfoPair
	Count    usize
	Capacity usize
}

type TypeSetIterator struct {
	Set   *TypeSet
	Index usize
}

func (it *TypeSetIterator) Next() bool {
	for {
		it.Index++
		if it.Set.Capacity == it.Index {
			return false
		}
		key := it.Set.Keys[it.Index]
		if key.Hash != 0 && key.Hash != TYPE_SET_TOMBSTONE {
			return true
		}
	}
}

func type_set_init(s *TypeSet, capacity ...isize) {
	cap_ := isize(16)
	if len(capacity) > 0 {
		cap_ = capacity[0]
	}
	s.Keys = make([]TypeInfoPair, cap_)
	s.Capacity = usize(cap_)
	s.Count = 0
}

func type_set_destroy(s *TypeSet) {
	s.Keys = nil
	s.Capacity = 0
	s.Count = 0
}

func type_set_add(s *TypeSet, ptr *Type) *Type {
	return type_set_add_pair(s, TypeInfoPair{Type: ptr, Hash: type_hash_canonical_type(ptr)})
}

func type_set_add_pair(s *TypeSet, pair TypeInfoPair) *Type {
	if s.Count >= usize(float64(s.Capacity)*0.75) {
		type_set_grow(s)
	}
	mask := s.Capacity - 1
	pos := pair.Hash & mask
	dist := usize(0)
	hash := pair.Hash
	key := pair
	for {
		slot := &s.Keys[pos]
		if slot.Hash == 0 {
			*slot = key
			s.Count++
			return nil
		}
		if slot.Hash == hash && slot.Type == key.Type {
			old := slot.Type
			*slot = key
			return old
		}
		existing_dist := (pos - usize(slot.Hash)) & mask
		if dist > existing_dist {
			key, slot.Type = slot.Type, key.Type
			hash, slot.Hash = slot.Hash, hash
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func type_set_grow(s *TypeSet) {
	new_cap := s.Capacity * 2
	new_mask := new_cap - 1
	new_keys := make([]TypeInfoPair, new_cap)
	if s.Count > 0 {
		for i := usize(0); i < s.Capacity; i++ {
			slot := s.Keys[i]
			if slot.Hash != 0 {
				type_set_insert_for_rehash(new_keys, new_mask, slot)
			}
		}
	}
	s.Keys = new_keys
	s.Capacity = new_cap
}

func type_set_insert_for_rehash(keys []TypeInfoPair, mask usize, pair TypeInfoPair) {
	pos := pair.Hash & mask
	dist := usize(0)
	hash := pair.Hash
	key := pair
	for {
		slot := &keys[pos]
		if slot.Hash == 0 {
			*slot = key
			return
		}
		existing_dist := (pos - usize(slot.Hash)) & mask
		if dist > existing_dist {
			key, *slot = *slot, key
			hash, slot.Hash = slot.Hash, hash
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func type_set_update(s *TypeSet, ptr *Type) bool {
	return type_set_update_pair(s, TypeInfoPair{Type: ptr, Hash: type_hash_canonical_type(ptr)})
}

func type_set_update_pair(s *TypeSet, pair TypeInfoPair) bool {
	old := type_set_add_pair(s, pair)
	return old != nil
}

func type_set_update_with_mutex(s *TypeSet, pair TypeInfoPair, m *RWSpinLock) bool {
	return type_set_update_pair(s, pair)
}

func type_set_update_with_mutex_ptr(s *TypeSet, ptr *Type, m *RWSpinLock) bool {
	return type_set_update(s, ptr)
}

func type_set_exists(s *TypeSet, ptr *Type) bool {
	return type_set_retrieve(s, ptr) != nil
}

func type_set_remove(s *TypeSet, ptr *Type) {
	hash := type_hash_canonical_type(ptr)
	mask := s.Capacity - 1
	pos := hash & mask
	dist := usize(0)
	for {
		slot := &s.Keys[pos]
		if slot.Hash == 0 {
			return
		}
		existing_dist := (pos - usize(slot.Hash)) & mask
		if dist > existing_dist {
			return
		}
		if slot.Hash == hash && slot.Type == ptr {
			slot.Hash = TYPE_SET_TOMBSTONE
			slot.Type = nil
			s.Count--
			return
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func type_set_clear(s *TypeSet) {
	for i := range s.Keys {
		s.Keys[i] = TypeInfoPair{}
	}
	s.Count = 0
}

func type_set_retrieve(s *TypeSet, ptr *Type) *TypeInfoPair {
	hash := type_hash_canonical_type(ptr)
	mask := s.Capacity - 1
	pos := hash & mask
	dist := usize(0)
	for {
		slot := &s.Keys[pos]
		if slot.Hash == 0 {
			return nil
		}
		existing_dist := (pos - usize(slot.Hash)) & mask
		if dist > existing_dist {
			return nil
		}
		if slot.Hash == hash && slot.Type == ptr {
			return slot
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func type_set_begin(s *TypeSet) TypeSetIterator {
	if s.Count == 0 {
		return type_set_end(s)
	}
	index := usize(0)
	for index < s.Capacity {
		if s.Keys[index].Hash != 0 {
			break
		}
		index++
	}
	return TypeSetIterator{Set: s, Index: index}
}

func type_set_end(s *TypeSet) TypeSetIterator {
	return TypeSetIterator{Set: s, Index: s.Capacity}
}

func map_get_type[V any](h *PtrMap[u64, V], key *Type) *V {
	return map_get(h, type_hash_canonical_type(key))
}

func map_set_type[V any](h *PtrMap[u64, V], key *Type, value V) {
	map_set(h, type_hash_canonical_type(key), value)
}

func map_must_get_type[V any](h *PtrMap[u64, V], key *Type) *V {
	ptr := map_get(h, type_hash_canonical_type(key))
	if ptr == nil {
		gb_assert_handler("Assertion Failure", "ptr != nullptr", "checker_hpp.go", 0, 0)
	}
	return ptr
}

type ProcCheckedState u8

const (
	ProcCheckedStateUnchecked  ProcCheckedState = 0
	ProcCheckedStateInProgress ProcCheckedState = 1
	ProcCheckedStateChecked    ProcCheckedState = 2
	ProcCheckedStateCOUNT      ProcCheckedState = 3
)

var ProcCheckedStateStrings = [ProcCheckedStateCOUNT]string{
	"Unchecked",
	"In Progress",
	"Checked",
}

type VariadicReuseData struct {
	SliceType *Type
	MaxCount  i64
}

type DeclInfo struct {
	Parent              *DeclInfo
	NextChild           *DeclInfo
	NextSibling         *DeclInfo
	Scope               *Scope
	Entity              *Entity
	DeclNode            *Ast
	TypeExpr            *Ast
	InitExpr            *Ast
	Attributes          []*Ast
	ProcLit             *Ast
	GenProcType         *Type
	ParaPolyOriginal    *Entity
	IsUsing             bool
	ForeignRequireResults bool
	WhereClausesEvaluated bool
	ProcCheckedState    ProcCheckedState
	DeferUsed           isize
	DeferUseChecked     bool
	Comment             *CommentGroup
	Docs                *CommentGroup
	Deps                *PtrSet[*Entity]
	TypeInfoDeps        TypeSet
	Labels              []BlockLabel
	ScopeIndex          i32
	VariadicReuses      []VariadicReuseData
	VariadicReuseMaxBytes i64
	VariadicReuseMaxAlign i64
	CodeGenModule       *lbModule
}

type ProcInfo struct {
	File                  *AstFile
	Token                 Token
	Decl                  *DeclInfo
	Type                  *Type
	Body                  *Ast
	Tags                  u64
	GeneratedFromPolymorphic bool
	PolyDefNode           *Ast
}

const DefaultScopeCapacity = 32

type ScopeMapSlot struct {
	Hash  u32
	Pad   u32
	Value *Entity
}

const ScopeMapInlineCap = 16

type ScopeMap struct {
	InlineKeys  [ScopeMapInlineCap]string
	InlineSlots [ScopeMapInlineCap]ScopeMapSlot
	Keys        []string
	Slots       []ScopeMapSlot
	Count       u32
	Cap         u32
}

func scope_map_max_load(cap u32) u32 {
	return cap - (cap >> 2)
}

func scope_map_init(m *ScopeMap) {
	m.Cap = ScopeMapInlineCap
	m.Slots = m.InlineSlots[:]
	m.Keys = m.InlineKeys[:]
}

func scope_map_insert_for_rehash(keys []string, slots []ScopeMapSlot, mask u32, key string, hash u32, value *Entity) *Entity {
	pos := hash & mask
	dist := u32(0)
	for {
		s := &slots[pos]
		if s.Hash == 0 {
			keys[pos] = key
			s.Hash = hash
			s.Value = value
			return nil
		}
		existing_dist := (pos - s.Hash) & mask
		if dist > existing_dist {
			tmp_key := keys[pos]
			tmp_hash := s.Hash
			tmp_value := s.Value
			keys[pos] = key
			s.Hash = hash
			s.Value = value
			key = tmp_key
			hash = tmp_hash
			value = tmp_value
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_allocate_entries(cap u32) ([]string, []ScopeMapSlot) {
	keys := make([]string, cap)
	slots := make([]ScopeMapSlot, cap)
	return keys, slots
}

func scope_map_grow(m *ScopeMap) {
	new_cap := m.Cap << 1
	new_mask := new_cap - 1
	new_keys, new_slots := scope_map_allocate_entries(new_cap)
	if m.Count > 0 {
		for i := u32(0); i < m.Cap; i++ {
			if m.Slots[i].Hash != 0 {
				scope_map_insert_for_rehash(new_keys, new_slots, new_mask, m.Keys[i], m.Slots[i].Hash, m.Slots[i].Value)
			}
		}
	}
	m.Slots = new_slots
	m.Keys = new_keys
	m.Cap = new_cap
}

func scope_map_reserve(m *ScopeMap, capacity isize) {
	if m.Slots == nil {
		scope_map_init(m)
	}
	new_cap := next_pow2_u32(u32(capacity))
	if m.Cap < new_cap && new_cap > ScopeMapInlineCap {
		m.Keys, m.Slots = scope_map_allocate_entries(new_cap)
		m.Cap = new_cap
	}
}

func scope_map_insert(m *ScopeMap, key string, hash u32, value *Entity) *Entity {
	if m.Slots == nil {
		scope_map_init(m)
	}
	if m.Count >= scope_map_max_load(m.Cap) {
		scope_map_grow(m)
	}
	mask := m.Cap - 1
	pos := hash & mask
	dist := u32(0)
	for {
		s := &m.Slots[pos]
		if s.Hash == 0 {
			m.Keys[pos] = key
			s.Hash = hash
			s.Value = value
			m.Count++
			return nil
		}
		if s.Hash == hash && m.Keys[pos] == key {
			old := s.Value
			s.Value = value
			return old
		}
		existing_dist := (pos - s.Hash) & mask
		if dist > existing_dist {
			tmp_key := m.Keys[pos]
			tmp_hash := s.Hash
			tmp_value := s.Value
			m.Keys[pos] = key
			s.Hash = hash
			s.Value = value
			key = tmp_key
			hash = tmp_hash
			value = tmp_value
			dist = existing_dist
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_get(m *ScopeMap, key string, hash u32) *Entity {
	if m.Slots == nil {
		return nil
	}
	mask := m.Cap - 1
	pos := hash & mask
	dist := u32(0)
	for {
		s := &m.Slots[pos]
		curr_hash := s.Hash
		if curr_hash == 0 {
			return nil
		}
		existing_dist := (pos - curr_hash) & mask
		if dist > existing_dist {
			return nil
		}
		if curr_hash == hash && m.Keys[pos] == key {
			return s.Value
		}
		dist++
		pos = (pos + 1) & mask
	}
}

func scope_map_clear(m *ScopeMap) {
	for i := range m.Slots {
		m.Slots[i] = ScopeMapSlot{}
	}
	m.Count = 0
}

type ScopeMapIterator struct {
	Map   *ScopeMap
	Index u32
}

func (it *ScopeMapIterator) Next() bool {
	for {
		it.Index++
		if it.Map.Cap == it.Index {
			return false
		}
		s := &it.Map.Slots[it.Index]
		if s.Hash != 0 {
			return true
		}
	}
}

func scope_map_begin(m *ScopeMap) ScopeMapIterator {
	if m.Count == 0 || m.Slots == nil {
		return scope_map_end(m)
	}
	index := u32(0)
	for index < m.Cap {
		if m.Slots[index].Hash != 0 {
			break
		}
		index++
	}
	return ScopeMapIterator{Map: m, Index: index}
}

func scope_map_end(m *ScopeMap) ScopeMapIterator {
	return ScopeMapIterator{Map: m, Index: m.Cap}
}

type ScopeFlag int

const (
	ScopeFlagPkg              ScopeFlag = 1 << 1
	ScopeFlagBuiltin          ScopeFlag = 1 << 2
	ScopeFlagGlobal           ScopeFlag = 1 << 3
	ScopeFlagFile             ScopeFlag = 1 << 4
	ScopeFlagInit             ScopeFlag = 1 << 5
	ScopeFlagProc             ScopeFlag = 1 << 6
	ScopeFlagType             ScopeFlag = 1 << 7
	ScopeFlagHasBeenImported  ScopeFlag = 1 << 10
	ScopeFlagContextDefined   ScopeFlag = 1 << 16
)

type Scope struct {
	Node            *Ast
	Parent          *Scope
	Next            *Scope
	HeadChild       *Scope
	Index           i32
	Elements        ScopeMap
	Imported        *PtrSet[*Scope]
	DeclInfo        *DeclInfo
	Flags           i32
	Pkg             *AstPackage
	File            *AstFile
	ProcedureEntity *Entity
}

type EntityGraphNodeSet = *PtrSet[*EntityGraphNode]

type EntityGraphNode struct {
	Entity   *Entity
	Pred     EntityGraphNodeSet
	Succ     EntityGraphNodeSet
	Index    isize
	DepCount isize
}

type ImportGraphNodeSet = *PtrSet[*ImportGraphNode]

type ImportGraphNode struct {
	Pkg      *AstPackage
	Scope    *Scope
	Pred     ImportGraphNodeSet
	Succ     ImportGraphNodeSet
	Index    isize
	DepCount isize
}

type EntityVisiblityKind int

const (
	EntityVisiblityPublic           EntityVisiblityKind = iota
	EntityVisiblityPrivateToPackage
	EntityVisiblityPrivateToFile
)

type ForeignContext struct {
	CurrLibrary    *Ast
	DefaultCC      ProcCallingConvention
	LinkPrefix     String
	LinkSuffix     String
	VisibilityKind EntityVisiblityKind
	RequireResults bool
}

type CheckerTypePath = []*Entity
type CheckerPolyPath = []*Type

type AtomOpMapEntry struct {
	Kind u32
	Node *Ast
}

type UntypedExprInfo struct {
	Expr *Ast
	Info *ExprInfo
}

type UntypedExprInfoMap = *PtrMap[*Ast, *ExprInfo]

type ObjcMsgKind u32

const (
	ObjcMsgNormal ObjcMsgKind = iota
	ObjcMsgFpret
	ObjcMsgFp2ret
	ObjcMsgStret
)

type ObjcMsgData struct {
	Kind     ObjcMsgKind
	ProcType *Type
}

type ObjcMethodData struct {
	Ac          AttributeContext
	ProcEntity  *Entity
}

type LoadFileTier int

const (
	LoadFileTierInvalid  LoadFileTier = iota
	LoadFileTierExists
	LoadFileTierContents
)

type LoadFileCache struct {
	Tier      LoadFileTier
	Exists    bool
	Path      String
	FileError gbFileError
	Data      String
	Hashes    StringMap[u64]
}

type LoadDirectoryFile struct {
	FileName String
	Data     String
}

type LoadDirectoryCache struct {
	Path      String
	FileError gbFileError
	Files     []*LoadFileCache
}

type GenProcsData struct {
	Procs []*Entity
}

type GenTypesData struct {
	Types []*Entity
}

type Defineable struct {
	Name            String
	DefaultValue    ExactValue
	Pos             TokenPos
	Docs            *CommentGroup
	DefaultValueStr String
	PosStr          String
}

type RaddbgTypeView struct {
	Type *Type
	View String
}

type CheckerInfo struct {
	Checker                               *Checker
	Files                                 StringMap[*AstFile]
	Packages                              StringMap[*AstPackage]
	VariableInitOrder                     []*DeclInfo
	BuiltinPackage                        *AstPackage
	RuntimePackage                        *AstPackage
	InitPackage                           *AstPackage
	InitScope                             *Scope
	EntryPoint                            *Entity
	MinDepTypeInfoIndexMap                *PtrMap[u64, isize]
	MinDepTypeInfoSet                     TypeSet
	TypeInfoTypesHashMap                  []TypeInfoPair
	TestingProcedures                     []*Entity
	InitProcedures                        []*Entity
	FiniProcedures                        []*Entity
	Definitions                           []*Entity
	Entities                              []*Entity
	RequiredForeignImportsThroughForce    []*Entity
	Defineables                           []Defineable
	GlobalUntyped                         UntypedExprInfoMap
	Foreigns                              StringMap[*Entity]
	DefinitionQueue                       *MPSCQueue[*Entity]
	EntityQueue                           *MPSCQueue[*Entity]
	RequiredGlobalVariableQueue           *MPSCQueue[*Entity]
	RequiredForeignImportsThroughForceQueue *MPSCQueue[*Entity]
	ForeignImportsToCheckFullpaths        *MPSCQueue[*Entity]
	ForeignDeclsToCheck                   *MPSCQueue[*Entity]
	RaddbgTypeViews                       []RaddbgTypeView
	RaddbgTypeViewsQueue                  *MPSCQueue[RaddbgTypeView]
	IntrinsicsEntryPointUsage             *MPSCQueue[*Ast]
	ObjcMsgSendTypes                      *PtrMap[*Ast, ObjcMsgData]
	ObcjClassNameSet                      StringSet
	ObjcClassImplementations              *MPSCQueue[*Entity]
	ObjcMethodImplementations             *PtrMap[*Type, []ObjcMethodData]
	LoadFileCache                         StringMap[*LoadFileCache]
	AllProceduresQueue                    *MPSCQueue[*ProcInfo]
	AllProcedures                         []*ProcInfo
	InstrumentationEnterEntity            *Entity
	InstrumentationExitEntity             *Entity
	LoadDirectoryCache                    StringMap[*LoadDirectoryCache]
	LoadDirectoryMap                      *PtrMap[*Ast, *LoadDirectoryCache]
}

type CheckerContext struct {
	Checker                         *Checker
	Info                            *CheckerInfo
	Pkg                             *AstPackage
	File                            *AstFile
	Scope                           *Scope
	Decl                            *DeclInfo
	StateFlags                      u32
	InDefer                         bool
	TypeHint                        *Type
	TypeHintExpr                    *Ast
	ProcName                        String
	CurrProcDecl                    *DeclInfo
	CurrProcSig                     *Type
	CurrProcCallingConvention       ProcCallingConvention
	InProcSig                       bool
	ForeignContext                  ForeignContext
	TypePath                        *CheckerTypePath
	TypeLevel                       isize
	Untyped                         *UntypedExprInfoMap
	InlineForDepth                  i64
	StmtFlags                       u32
	InEnumType                      bool
	InProcGroup                     bool
	CollectDelayedDecls             bool
	AllowPolymorphicTypes           bool
	DisallowPolymorphicReturnTypes  bool
	NoPolymorphicErrors             bool
	HidePolymorphicErrors           bool
	InPolymorphicSpecialization     bool
	AllowArrowRightSelectorExpr     bool
	BitFieldBitSize                 u8
	PolymorphicScope                *Scope
	AssignmentLhsHint               *Ast
}

type Checker struct {
	Parser                          *Parser
	Info                            CheckerInfo
	BuiltinCtx                      CheckerContext
	ProcsWithDeferredToCheck        *MPSCQueue[*Entity]
	ProcsWithObjcContextProviderToCheck *MPSCQueue[*Entity]
	ProcsToCheck                    []*ProcInfo
	NestedProcLits                  []*DeclInfo
	GlobalUntypedQueue              *MPSCQueue[UntypedExprInfo]
	SoaTypesToComplete              *MPSCQueue[*Type]
}

var builtin_pkg    *AstPackage = nil
var intrinsics_pkg *AstPackage = nil
var config_pkg     *AstPackage = nil
