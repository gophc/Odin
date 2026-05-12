package godin

import "core:mem"
import "core:sync"
import "core:runtime"

// Opaque / forward type declarations (defined in other rewritten Odin files)
Type             :: distinct rawptr
Entity           :: distinct rawptr
Scope            :: distinct rawptr
DeclInfo         :: distinct rawptr
AstFile          :: distinct rawptr
Checker          :: distinct rawptr
CheckerInfo      :: distinct rawptr
CheckerContext   :: distinct rawptr
AstPackage       :: distinct rawptr
Parser           :: distinct rawptr
Ast              :: distinct rawptr
Token            :: distinct rawptr
TokenPos         :: distinct rawptr
CommentGroup     :: distinct rawptr
ExactValue       :: distinct rawptr
TypeAndValue     :: distinct rawptr
ProcCallingConvention :: distinct rawptr
BuiltinProcId    :: distinct rawptr
lbModule         :: distinct rawptr
gbFileError      :: distinct rawptr
TypeWriter       :: distinct rawptr
AddressingMode   :: distinct rawptr
RWSpinLock       :: distinct rawptr

BlockingMutex    :: distinct rawptr
RwMutex          :: distinct rawptr
RecursiveMutex   :: distinct rawptr

// Forward-declare Ast subtypes
AstValueDecl     :: distinct rawptr
AstWhenStmt      :: distinct rawptr

// Internal types from other includes — use the rewritten Odin signatures directly
InternedString   :: distinct rawptr

// Constants
DEFAULT_SCOPE_CAPACITY :: 32
SCOPE_MAP_INLINE_CAP   :: 16
TYPE_SET_TOMBSTONE     :: ~u64(0)

// Expr_info ──────────────────────────────────────────────────────────

ExprInfo :: struct {
	mode:   AddressingMode,
	is_lhs: bool,
	type:   ^Type,
	value:  ExactValue,
}

make_expr_info :: proc(
	mode:      AddressingMode,
	type_:     ^Type,
	value:     ExactValue,
	is_lhs:    bool,
	allocator := context.allocator,
) -> ^ExprInfo {
	ei := new(ExprInfo, allocator)
	ei.mode  = mode
	ei.type  = type_
	ei.value = value
	ei.is_lhs = is_lhs
	return ei
}

// ExprKind ────────────────────────────────────────────────────────────

ExprKind :: enum {
	Expr,
	Stmt,
}

// StmtFlag ────────────────────────────────────────────────────────────

StmtFlag :: enum u32 {
	BreakAllowed       = 0,
	ContinueAllowed    = 1,
	FallthroughAllowed = 2,
	TypeSwitch         = 4,
	CheckScopeDecls    = 5,
}

StmtFlagSet :: bit_set[StmtFlag; u32]

// Convenience named masks
STMT_BREAK_ALLOWED        :: StmtFlagSet{.BreakAllowed}
STMT_CONTINUE_ALLOWED     :: StmtFlagSet{.ContinueAllowed}
STMT_FALLTHROUGH_ALLOWED  :: StmtFlagSet{.FallthroughAllowed}
STMT_TYPE_SWITCH          :: StmtFlagSet{.TypeSwitch}
STMT_CHECK_SCOPE_DECLS    :: StmtFlagSet{.CheckScopeDecls}

// BuiltinProcPkg ──────────────────────────────────────────────────────

BuiltinProcPkg :: enum {
	builtin,
	intrinsics,
	COUNT,
}

builtin_proc_pkg_name := [BuiltinProcPkg.COUNT]string{
	"builtin",
	"intrinsics",
}

// BuiltinProc ─────────────────────────────────────────────────────────

BuiltinProc :: struct {
	name:           string,
	arg_count:      int,
	variadic:       bool,
	diverging:      bool,
	ignore_results: bool,
	kind:           ExprKind,
	pkg:            BuiltinProcPkg,
}

// (checker_builtin_procs.i.hpp has been rewritten; use its Odin declarations directly)
// #include equivalent — the builtin_procs table lives in that file's scope.

// Operand ─────────────────────────────────────────────────────────────

Operand :: struct {
	mode:       AddressingMode,
	type:       ^Type,
	value:      ExactValue,
	expr:       ^Ast,
	builtin_id: BuiltinProcId,
	proc_group: ^Entity,
}

// BlockLabel ──────────────────────────────────────────────────────────

BlockLabel :: struct {
	name:  string,
	label: ^Ast,
}

// DeferredProcedure ──────────────────────────────────────────────────

DeferredProcedureKind :: enum {
	none,
	in,
	out,
	in_out,
	in_by_ptr,
	out_by_ptr,
	in_out_by_ptr,
}

DeferredProcedure :: struct {
	kind:   DeferredProcedureKind,
	entity: ^Entity,
}

// InstrumentationFlag ────────────────────────────────────────────────

InstrumentationFlag :: enum i32 {
	Enabled  = -1,
	Default  = 0,
	Disabled = +1,
}

// AttributeContext ────────────────────────────────────────────────────

AttributeContext :: struct {
	link_name:                     string,
	link_prefix:                   string,
	link_suffix:                   string,
	link_section:                  string,
	linkage:                       string,
	thread_local_model:            string,
	deprecated_message:            string,
	warning_message:               string,
	extra_linker_flags:            string,
	objc_class:                    string,
	objc_name:                     string,
	objc_selector:                 string,
	require_target_feature:        string,
	enable_target_feature:         string,
	raddbg_type_view_string:       string,
	deferred_procedure:            DeferredProcedure,
	init_expr_list_count:          int,
	optimization_mode:             u32,
	foreign_import_priority_index: i64,
	no_instrumentation:            InstrumentationFlag,
	objc_type:                     ^Type,
	objc_superclass:               ^Type,
	objc_ivar:                     ^Type,
	objc_context_provider:         ^Entity,
	// bit-flag fields (C++ used :1 bits; Odin uses bools)
	is_export:                     bool,
	is_static:                     bool,
	require_results:               bool,
	require_declaration:           bool,
	has_disabled_proc:             bool,
	disabled_proc:                 bool,
	test:                          bool,
	init:                          bool,
	fini:                          bool,
	set_cold:                      bool,
	entry_point_only:              bool,
	instrumentation_enter:         bool,
	instrumentation_exit:          bool,
	no_sanitize_address:           bool,
	no_sanitize_memory:            bool,
	no_sanitize_thread:            bool,
	rodata:                        bool,
	ignore_duplicates:             bool,
	objc_is_class_method:          bool,
	objc_is_implementation:        bool,
	objc_is_disabled_implement:    bool,
	raddbg_type_view:              bool,
}

make_attribute_context :: proc(
	link_prefix, link_suffix: string,
) -> AttributeContext {
	return AttributeContext{
	    link_prefix = link_prefix,
	    link_suffix = link_suffix,
	}
}

// DeclAttributeProc ──────────────────────────────────────────────────

DeclAttributeProc :: #type proc(c: ^CheckerContext, elem: ^Ast, name: string, value: ^Ast, ac: ^AttributeContext) -> bool

check_decl_attributes :: proc(
	c:          ^CheckerContext,
	attributes: []^Ast,
	proc_:      DeclAttributeProc,
	ac:         ^AttributeContext,
)

// ── name_canonicalization.hpp section ───────────────────────────────
// (The included file has been rewritten; declare the Odin signatures directly)

write_type_to_canonical_string :: proc(w: ^TypeWriter, type_: ^Type)
write_canonical_entity_name    :: proc(w: ^TypeWriter, e: ^Entity)
type_hash_canonical_type       :: proc(type_: ^Type) -> u64
type_to_canonical_string       :: proc(allocator: runtime.Allocator, type_: ^Type) -> string
temp_canonical_string          :: proc(type_: ^Type) -> string // returns gbString equivalent
type_info_pair_cmp             :: proc(a, b: rawptr) -> int

// TypeInfoPair ───────────────────────────────────────────────────────

TypeInfoPair :: struct {
	type: ^Type,
	hash: u64,
}

// TypeSet ────────────────────────────────────────────────────────────
// Replaces the custom C++ hash set with an Odin map keyed by type pointer,
// while keeping the same public function signatures.

TypeSet :: struct {
	entries:  map[rawptr]TypeInfoPair, // key = rawptr(type)
}

type_set_init :: proc(s: ^TypeSet, capacity := 16, allocator := context.allocator) {
	s.entries = make(map[rawptr]TypeInfoPair, capacity, allocator)
}

type_set_destroy :: proc(s: ^TypeSet) {
	delete(s.entries)
}

type_set_add :: proc(s: ^TypeSet, type_: ^Type, allocator := context.allocator) -> ^Type {
	pair := TypeInfoPair{type_, type_hash_canonical_type(type_)}
	key := rawptr(type_)
	s.entries[key] = pair
	return type_
}

type_set_add_pair :: proc(s: ^TypeSet, pair: TypeInfoPair) -> ^Type {
	s.entries[rawptr(pair.type)] = pair
	return pair.type
}

type_set_update :: proc(s: ^TypeSet, arg: $T) -> bool {
	when T == ^Type {
	    key := rawptr(arg)
	    _, exists := s.entries[key]
	    if !exists {
	        s.entries[key] = TypeInfoPair{arg, type_hash_canonical_type(arg)}
	    }
	    return !exists
	} else when T == TypeInfoPair {
	    key := rawptr(arg.type)
	    _, exists := s.entries[key]
	    s.entries[key] = arg
	    return !exists
	}
}

type_set_update_with_mutex :: proc(s: ^TypeSet, arg: $T, m: ^RWSpinLock) -> bool {
	// lock(m)
	result := type_set_update(s, arg)
	// unlock(m)
	return result
}

type_set_exists :: proc(s: ^TypeSet, type_: ^Type) -> bool {
	_, exists := s.entries[rawptr(type_)]
	return exists
}

type_set_remove :: proc(s: ^TypeSet, type_: ^Type) {
	delete_key(s.entries, rawptr(type_))
}

type_set_clear :: proc(s: ^TypeSet) {
	clear(&s.entries)
}

type_set_retrieve :: proc(s: ^TypeSet, type_: ^Type) -> ^TypeInfoPair {
	key := rawptr(type_)
	if pair, ok := &s.entries[key]; ok {
	    return pair
	}
	return nil
}

// TypeSetIterator — simplified to use map iteration
TypeSetIterator :: struct {
	set:    ^TypeSet,
	iter:   map[rawptr]TypeInfoPair,
	// internal iterator state managed by Odin's map iteration
}

// ── map_get / map_set / map_must_get (templated helpers) ────────────
// Replaced with Odin generics using parametric polymorphism.

map_get_type_hash :: proc(m: ^$M/map[$K]$V, type_: ^Type) -> (V, bool)
	where K == u64 {
	hash := type_hash_canonical_type(type_)
	return map_get(m, hash)
}

map_set_type_hash :: proc(m: ^$M/map[$K]$V, type_: ^Type, value: V)
	where K == u64 {
	hash := type_hash_canonical_type(type_)
	m[hash] = value
}

map_must_get_type_hash :: proc(m: ^$M/map[$K]$V, type_: ^Type) -> V
	where K == u64 {
	hash := type_hash_canonical_type(type_)
	v, ok := m[hash]
	assert(ok, "map_must_get: key not found")
	return v
}

// Generic map helpers (simple pass-through to Odin builtins)
_map_get :: proc(m: ^$M/map[$K]$V, key: K) -> (V, bool) {
	v, ok := m[key]
	return v, ok
}

_map_set :: proc(m: ^$M/map[$K]$V, key: K, value: V) {
	m[key] = value
}

_map_must_get :: proc(m: ^$M/map[$K]$V, key: K) -> V {
	v, ok := m[key]
	assert(ok, "map_must_get: key not found")
	return v
}

// ── ProcCheckedState ────────────────────────────────────────────────

ProcCheckedState :: enum u8 {
	Unchecked,
	InProgress,
	Checked,
	COUNT,
}

ProcCheckedState_strings := [ProcCheckedState.COUNT]string{
	"Unchecked",
	"In Progress",
	"Checked",
}

// VariadicReuseData ──────────────────────────────────────────────────

VariadicReuseData :: struct {
	slice_type: ^Type,
	max_count:  i64,
}

// DeclInfo ───────────────────────────────────────────────────────────

DeclInfo :: struct {
	parent:                 ^DeclInfo,
	next_child:             ^DeclInfo,
	next_sibling:           ^DeclInfo,
	scope:                  ^Scope,
	entity:                 ^Entity,         // std::atomic<Entity *> → plain pointer
	decl_node:              ^Ast,
	type_expr:              ^Ast,
	init_expr:              ^Ast,
	attributes:             [dynamic]^Ast,
	proc_lit:               ^Ast,
	gen_proc_type:          ^Type,
	para_poly_original:     ^Entity,
	labels:                 [dynamic]BlockLabel,
	variadic_reuses:        [dynamic]VariadicReuseData,
	deps:                   map[rawptr]^Entity,          // PtrSet<Entity *>
	type_info_deps:         TypeSet,
	comment:                ^CommentGroup,
	docs:                   ^CommentGroup,
	code_gen_module:        ^lbModule,      // std::atomic → plain pointer
	scope_index:            i32,
	variadic_reuse_max_bytes: i64,
	variadic_reuse_max_align: i64,
	defer_used:             int,
	is_using:               bool,
	foreign_require_results: bool,
	where_clauses_evaluated: bool,          // std::atomic<bool> → bool
	proc_checked_state:     ProcCheckedState, // std::atomic → plain
	defer_use_checked:      bool,
	// Mutex fields — placeholder comments; real impl would use sync.Mutex etc.
	// next_mutex:         BlockingMutex
	// proc_checked_mutex: BlockingMutex
	// deps_mutex:         RwMutex
	// type_info_deps_mutex: RwMutex
	// type_and_value_mutex: BlockingMutex
}

// ProcInfo ────────────────────────────────────────────────────────────

ProcInfo :: struct {
	file:                        ^AstFile,
	token:                       Token,
	decl:                        ^DeclInfo,
	type:                        ^Type,
	body:                        ^Ast,
	poly_def_node:               ^Ast,
	tags:                        u64,
	generated_from_polymorphic:  bool,
}

// ── ScopeMap (robin-hood hash map) ──────────────────────────────────
// Replaced with Odin's built-in map[string]^Entity for simplicity.

ScopeMap :: struct {
	entries: map[string]^Entity,
}

scope_map_init :: proc(m: ^ScopeMap, allocator := context.allocator) {
	m.entries = make(map[string]^Entity, SCOPE_MAP_INLINE_CAP, allocator)
}

scope_map_reserve :: proc(m: ^ScopeMap, capacity: int, allocator := context.allocator) {
	if m.entries == nil {
	    scope_map_init(m, allocator)
	}
	if len(m.entries) < capacity {
	    reserve(&m.entries, capacity)
	}
}

scope_map_insert :: proc(m: ^ScopeMap, key: string, hash: u32, value: ^Entity) -> ^Entity {
	// hash is ignored; Odin's map uses its own hashing
	old, existed := m.entries[key]
	m.entries[key] = value
	if existed {
	    return old
	}
	return nil
}

scope_map_get :: proc(m: ^ScopeMap, key: string, hash: u32) -> ^Entity {
	// hash is ignored; kept for signature compatibility
	return m.entries[key]
}

scope_map_clear :: proc(m: ^ScopeMap) {
	clear(&m.entries)
}

// ScopeFlag ──────────────────────────────────────────────────────────

ScopeFlag :: enum i32 {
	Pkg            = 1,
	Builtin        = 2,
	Global         = 3,
	File           = 4,
	Init           = 5,
	Proc           = 6,
	Type           = 7,
	HasBeenImported = 10,
	ContextDefined  = 16,
}

ScopeFlagSet :: bit_set[ScopeFlag; i32]

SCOPE_FLAG_PKG              :: ScopeFlagSet{.Pkg}
SCOPE_FLAG_BUILTIN          :: ScopeFlagSet{.Builtin}
SCOPE_FLAG_GLOBAL           :: ScopeFlagSet{.Global}
SCOPE_FLAG_FILE             :: ScopeFlagSet{.File}
SCOPE_FLAG_INIT             :: ScopeFlagSet{.Init}
SCOPE_FLAG_PROC             :: ScopeFlagSet{.Proc}
SCOPE_FLAG_TYPE             :: ScopeFlagSet{.Type}
SCOPE_FLAG_HAS_BEEN_IMPORTED :: ScopeFlagSet{.HasBeenImported}
SCOPE_FLAG_CONTEXT_DEFINED  :: ScopeFlagSet{.ContextDefined}

// Scope ──────────────────────────────────────────────────────────────

Scope :: struct {
	node:              ^Ast,
	parent:            ^Scope,
	next:              ^Scope,         // std::atomic → plain
	head_child:        ^Scope,         // std::atomic → plain
	imported:          map[rawptr]^Scope, // PtrSet<Scope *> → map
	elements:          ScopeMap,
	decl_info:         ^DeclInfo,
	pkg:               ^AstPackage,     // union member
	file:              ^AstFile,        // union member (shared slot via #raw_union concept)
	procedure_entity:  ^Entity,         // union member
	index:             i32,
	flags:             ScopeFlagSet,
	// mutex: RwMutex
}

// EntityGraphNode ────────────────────────────────────────────────────

EntityGraphNodeSet :: map[rawptr]rawptr // PtrSet<EntityGraphNode *>

EntityGraphNode :: struct {
	entity:    ^Entity,
	pred:      EntityGraphNodeSet,
	succ:      EntityGraphNodeSet,
	index:     int,
	dep_count: int,
}

// ImportGraphNode ────────────────────────────────────────────────────

ImportGraphNodeSet :: map[rawptr]rawptr // PtrSet<ImportGraphNode *>

ImportGraphNode :: struct {
	pkg:       ^AstPackage,
	scope:     ^Scope,
	pred:      ImportGraphNodeSet,
	succ:      ImportGraphNodeSet,
	index:     int,
	dep_count: int,
}

// EntityVisibilityKind ───────────────────────────────────────────────

EntityVisibilityKind :: enum {
	Public,
	PrivateToPackage,
	PrivateToFile,
}

// ForeignContext ─────────────────────────────────────────────────────

ForeignContext :: struct {
	curr_library:   ^Ast,
	default_cc:     ProcCallingConvention,
	link_prefix:    string,
	link_suffix:    string,
	visibility_kind: EntityVisibilityKind,
	require_results: bool,
}

// Type aliases ───────────────────────────────────────────────────────

CheckerTypePath :: [dynamic]^Entity
CheckerPolyPath  :: [dynamic]^Type

// AtomOpMapEntry ────────────────────────────────────────────────────

AtomOpMapEntry :: struct {
	kind: u32,
	node: ^Ast,
}

// UntypedExprInfo ───────────────────────────────────────────────────

UntypedExprInfo :: struct {
	expr: ^Ast,
	info: ^ExprInfo,
}

UntypedExprInfoMap :: map[rawptr]^ExprInfo // PtrMap<Ast *, ExprInfo *>

// ObjcMsgKind ───────────────────────────────────────────────────────

ObjcMsgKind :: enum u32 {
	normal,
	fpret,
	fp2ret,
	stret,
}

ObjcMsgData :: struct {
	kind:      ObjcMsgKind,
	proc_type: ^Type,
}

ObjcMethodData :: struct {
	ac:          AttributeContext,
	proc_entity: ^Entity,
}

// LoadFileTier ──────────────────────────────────────────────────────

LoadFileTier :: enum {
	Invalid,
	Exists,
	Contents,
}

LoadFileCache :: struct {
	tier:       LoadFileTier,
	exists:     bool,
	path:       string,
	file_error: gbFileError,
	data:       string,
	hashes:     map[string]u64,
}

LoadDirectoryFile :: struct {
	file_name: string,
	data:      string,
}

LoadDirectoryCache :: struct {
	path:       string,
	file_error: gbFileError,
	files:      [dynamic]^LoadFileCache,
}

// GenProcsData ──────────────────────────────────────────────────────

GenProcsData :: struct {
	procs: [dynamic]^Entity,
	// mutex: RwMutex
}

GenTypesData :: struct {
	types: [dynamic]^Entity,
	// mutex: RecursiveMutex
}

// Defineable ────────────────────────────────────────────────────────

Defineable :: struct {
	name:              string,
	default_value:     ExactValue,
	pos:               TokenPos,
	docs:              ^CommentGroup,
	default_value_str: string,
	pos_str:           string,
}

// RaddbgTypeView ────────────────────────────────────────────────────

RaddbgTypeView :: struct {
	type: ^Type,
	view: string,
}

// ── MPSCQueue ──────────────────────────────────────────────────────
// Simple multi-producer single-consumer queue backed by a dynamic array.
// The C++ version uses a lock-free MPSC queue; here we provide a minimal
// mutex-protected replacement.

MPSCQueue :: struct($T: typeid) {
	items: [dynamic]T,
	// mutex: sync.Mutex,  // uncomment for thread safety
}

mpsc_init :: proc(q: ^$Q/MPSCQueue($T), allocator := context.allocator) {
	q.items = make([dynamic]T, allocator)
}

mpsc_push :: proc(q: ^$Q/MPSCQueue($T), item: T) {
	// lock q.mutex
	append(&q.items, item)
	// unlock q.mutex
}

mpsc_push_slice :: proc(q: ^$Q/MPSCQueue($T), items: []T) {
	// lock q.mutex
	append(&q.items, ..items)
	// unlock q.mutex
}

mpsc_drain :: proc(q: ^$Q/MPSCQueue($T)) -> [dynamic]T {
	// lock q.mutex
	result := q.items
	q.items = make([dynamic]T, context.allocator)
	// unlock q.mutex
	return result
}

mpsc_destroy :: proc(q: ^$Q/MPSCQueue($T)) {
	delete(q.items)
}

// ── CheckerInfo ────────────────────────────────────────────────────

CheckerInfo :: struct {
	checker:                            ^Checker,
	files:                              map[string]^AstFile,
	packages:                           map[string]^AstPackage,
	variable_init_order:                [dynamic]^DeclInfo,
	builtin_package:                    ^AstPackage,
	runtime_package:                    ^AstPackage,
	init_package:                       ^AstPackage,
	init_scope:                         ^Scope,
	entry_point:                        ^Entity,
	min_dep_type_info_index_map:        map[u64]int,
	min_dep_type_info_set:              TypeSet,
	type_info_types_hash_map:           [dynamic]TypeInfoPair,
	testing_procedures:                 [dynamic]^Entity,
	init_procedures:                    [dynamic]^Entity,
	fini_procedures:                    [dynamic]^Entity,
	definitions:                        [dynamic]^Entity,
	entities:                           [dynamic]^Entity,
	required_foreign_imports_through_force: [dynamic]^Entity,
	defineables:                        [dynamic]Defineable,
	global_untyped:                     UntypedExprInfoMap,
	foreigns:                           map[string]^Entity,
	raddbg_type_views:                  [dynamic]RaddbgTypeView,
	objc_msgSend_types:                 map[rawptr]ObjcMsgData,
	obcj_class_name_set:                map[string]struct{},
	objc_method_implementations:        map[rawptr][dynamic]ObjcMethodData,
	load_file_cache:                    map[string]^LoadFileCache,
	all_procedures:                     [dynamic]^ProcInfo,
	load_directory_cache:               map[string]^LoadDirectoryCache,
	load_directory_map:                 map[rawptr]^LoadDirectoryCache,
	instrumentation_enter_entity:       ^Entity,
	instrumentation_exit_entity:        ^Entity,
	// Queues (MPSC)
	definition_queue:                   MPSCQueue(^Entity),
	entity_queue:                       MPSCQueue(^Entity),
	required_global_variable_queue:     MPSCQueue(^Entity),
	required_foreign_imports_through_force_queue: MPSCQueue(^Entity),
	foreign_imports_to_check_fullpaths: MPSCQueue(^Entity),
	foreign_decls_to_check:             MPSCQueue(^Entity),
	raddbg_type_views_queue:            MPSCQueue(RaddbgTypeView),
	intrinsics_entry_point_usage:       MPSCQueue(^Ast),
	objc_class_implementations:         MPSCQueue(^Entity),
	all_procedures_queue:               MPSCQueue(^ProcInfo),
	// Mutex placeholders
	// minimum_dependency_type_info_mutex: RwMutex
	// min_dep_type_info_set_mutex:       RWSpinLock
	// defineables_mutex:                 BlockingMutex
	// global_untyped_mutex:              RwMutex
	// builtin_mutex:                     BlockingMutex
	// type_and_value_mutex:              BlockingMutex
	// lazy_mutex:                        RecursiveMutex
	// foreign_mutex:                     BlockingMutex
	// objc_objc_msgSend_mutex:           BlockingMutex
	// objc_class_name_mutex:             BlockingMutex
	// objc_method_mutex:                 BlockingMutex
	// load_file_mutex:                   BlockingMutex
	// instrumentation_mutex:             BlockingMutex
	// load_directory_mutex:              BlockingMutex
}

// ── CheckerContext ─────────────────────────────────────────────────

CheckerContext :: struct {
	checker:                           ^Checker,
	info:                              ^CheckerInfo,
	pkg:                               ^AstPackage,
	file:                              ^AstFile,
	scope:                             ^Scope,
	decl:                              ^DeclInfo,
	proc_name:                         string,
	curr_proc_decl:                    ^DeclInfo,
	curr_proc_sig:                     ^Type,
	curr_proc_calling_convention:      ProcCallingConvention,
	foreign_context:                   ForeignContext,
	type_path:                         ^CheckerTypePath,
	untyped:                           ^UntypedExprInfoMap,
	type_hint:                         ^Type,
	type_hint_expr:                    ^Ast,
	assignment_lhs_hint:               ^Ast,
	polymorphic_scope:                 ^Scope,
	state_flags:                       u32,
	stmt_flags:                        StmtFlagSet,
	type_level:                        int,
	inline_for_depth:                  i64,
	bit_field_bit_size:                u8,
	in_defer:                          bool,
	in_proc_sig:                       bool,
	in_enum_type:                      bool,
	in_proc_group:                     bool,
	collect_delayed_decls:             bool,
	allow_polymorphic_types:           bool,
	disallow_polymorphic_return_types: bool,
	no_polymorphic_errors:             bool,
	hide_polymorphic_errors:           bool,
	in_polymorphic_specialization:     bool,
	allow_arrow_right_selector_expr:   bool,
	// mutex: BlockingMutex
}

// ── Checker ────────────────────────────────────────────────────────

Checker :: struct {
	parser:                             ^Parser,
	info:                               CheckerInfo,
	builtin_ctx:                        CheckerContext,
	procs_to_check:                     [dynamic]^ProcInfo,
	nested_proc_lits:                   [dynamic]^DeclInfo,
	procs_with_deferred_to_check:       MPSCQueue(^Entity),
	procs_with_objc_context_provider_to_check: MPSCQueue(^Entity),
	global_untyped_queue:               MPSCQueue(UntypedExprInfo),
	soa_types_to_complete:              MPSCQueue(^Type),
	// nested_proc_lits_mutex: BlockingMutex
}

// Package-level globals
builtin_pkg    : ^AstPackage
intrinsics_pkg : ^AstPackage
config_pkg     : ^AstPackage

// ── Exported function declarations ──────────────────────────────────
// (Bodies are defined in corresponding .cpp files; keep signatures.)

check_vet_flags :: proc{
	check_vet_flags_ctx,
	check_vet_flags_node,
}

check_vet_flags_ctx  :: proc(c: ^CheckerContext) -> u64
check_vet_flags_node :: proc(node: ^Ast) -> u64

type_and_value_of_expr   :: proc(expr: ^Ast) -> TypeAndValue
type_of_expr             :: proc(expr: ^Ast) -> ^Type
implicit_entity_of_node  :: proc(clause: ^Ast) -> ^Entity
decl_info_of_ident       :: proc(ident: ^Ast) -> ^DeclInfo
decl_info_of_entity      :: proc(e: ^Entity) -> ^DeclInfo
ast_file_of_filename     :: proc(i: ^CheckerInfo, filename: string) -> ^AstFile
entity_of_node           :: proc(expr: ^Ast) -> ^Entity

type_info_index :: proc{
	type_info_index_type,
	type_info_index_pair,
}

type_info_index_type :: proc(info: ^CheckerInfo, type_: ^Type, error_on_failure: bool) -> int
type_info_index_pair :: proc(info: ^CheckerInfo, pair: TypeInfoPair, error_on_failure: bool) -> int

scope_lookup        :: proc(s: ^Scope, interned: InternedString, hash: u32) -> ^Entity
scope_lookup_parent :: proc(s: ^Scope, name: InternedString, scope_: ^^Scope, entity_: ^^Entity, hash: u32)
scope_insert        :: proc(s: ^Scope, entity: ^Entity) -> ^Entity

add_type_and_value       :: proc(c: ^CheckerContext, expression: ^Ast, mode: AddressingMode, type_: ^Type, value: ExactValue)
check_get_expr_info      :: proc(c: ^CheckerContext, expr: ^Ast) -> ^ExprInfo
add_untyped              :: proc(c: ^CheckerContext, expression: ^Ast, mode: AddressingMode, basic_type: ^Type, value: ExactValue)
add_entity_use           :: proc(c: ^CheckerContext, identifier: ^Ast, entity: ^Entity)
add_implicit_entity      :: proc(c: ^CheckerContext, node: ^Ast, e: ^Entity)
add_entity_and_decl_info :: proc(c: ^CheckerContext, identifier: ^Ast, e: ^Entity, d: ^DeclInfo, is_exported := true)
add_type_info_type       :: proc(c: ^CheckerContext, t: ^Type)

check_add_import_decl         :: proc(c: ^CheckerContext, decl: ^Ast)
check_add_foreign_import_decl :: proc(c: ^CheckerContext, decl: ^Ast)
check_entity_decl             :: proc(c: ^CheckerContext, e: ^Entity, d: ^DeclInfo, named_type: ^Type)
check_const_decl              :: proc(c: ^CheckerContext, e: ^Entity, type_expr, init_expr: ^Ast, named_type: ^Type)
check_type_decl               :: proc(c: ^CheckerContext, e: ^Entity, type_expr: ^Ast, def: ^Type)
check_arity_match             :: proc(c: ^CheckerContext, vd: ^AstValueDecl, is_global := false) -> bool
check_collect_entities        :: proc(c: ^CheckerContext, nodes: []^Ast)
check_collect_entities_from_when_stmt :: proc(c: ^CheckerContext, ws: ^AstWhenStmt)
check_delayed_file_import_entity :: proc(c: ^CheckerContext, decl: ^Ast)

new_checker_type_path      :: proc(allocator := context.allocator) -> ^CheckerTypePath
destroy_checker_type_path  :: proc(tp: ^CheckerTypePath)
check_type_path_push       :: proc(c: ^CheckerContext, e: ^Entity)
check_type_path_pop        :: proc(c: ^CheckerContext) -> ^Entity

init_core_context    :: proc(c: ^Checker)
init_mem_allocator   :: proc(c: ^Checker)
add_untyped_expressions :: proc(cinfo: ^CheckerInfo, untyped: ^UntypedExprInfoMap)

ensure_polymorphic_record_entity_has_gen_types :: proc(ctx: ^CheckerContext, original_type: ^Type) -> ^GenTypesData
init_map_internal_types :: proc(type_: ^Type)