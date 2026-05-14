package cmd

import "unsafe"

const LLVMDWARFSourceLanguageOdin LLVMDWARFSourceLanguage = 47

// Global type info data entities
var lb_global_type_info_data_entity *Entity
var lb_global_type_info_member_types lbAddr
var lb_global_type_info_member_names lbAddr
var lb_global_type_info_member_offsets lbAddr
var lb_global_type_info_member_usings lbAddr
var lb_global_type_info_member_tags lbAddr
var lb_global_type_info_data_index isize
var lb_global_type_info_member_types_index isize
var lb_global_type_info_member_names_index isize
var lb_global_type_info_member_offsets_index isize
var lb_global_type_info_member_usings_index isize
var lb_global_type_info_member_tags_index isize

// =============================================================================
// lb_init_module_worker_proc
// =============================================================================
var lb_init_module_worker_proc WorkerTaskProc = func(data unsafe.Pointer) isize {
	m := (*lbModule)(data)

	m.ModuleName = goStr(m.Pkg.Name)
	m.Ctx = LLVMContextCreate()
	m.Mod = LLVMModuleCreateWithNameInContext(m.ModuleName, m.Ctx)

	{
		triplePtr := LLVMGetDefaultTargetTriple()
		tripleStr := ""
		if triplePtr != nil {
			tripleStr = unsafe.String(triplePtr, 256)
		}
		LLVMSetModuleTarget(m.Mod, tripleStr)
	}

	{
		fullpath := goStr(m.File.Fullpath)
		LLVMSetSourceFileName(m.Mod, fullpath, uint(len(fullpath)))
	}

	{
		i32Zero := LLVMConstInt(LLVMInt32TypeInContext(m.Ctx), 0, 0)
		mdVal := LLVMValueAsMetadata(i32Zero)
		LLVMAddModuleFlag(m.Mod, LLVMModuleFlagBehaviorWarning, "wchar_size", 8, mdVal)
		if buildContext.Metrics.TargetBits == 64 {
			LLVMAddModuleFlag(m.Mod, LLVMModuleFlagBehaviorWarning, "PIC Level", 8, mdVal)
			LLVMAddModuleFlag(m.Mod, LLVMModuleFlagBehaviorWarning, "PIE Level", 8, mdVal)
		}
	}

	m.DebugBuilder = LLVMCreateDebugInfoBuilder(m.Ctx)

	{
		str := ""
		if m.File != nil {
			str = goStr(m.File.Fullpath)
		} else if m.Pkg != nil {
			str = goStr(m.Pkg.Name)
		}
		diFile := LLVMDIBuilderCreateFile(m.DebugBuilder, str, uint(len(str)), str, uint(len(str)))
		producer := "Odin"
		flags := ""
		runtimeVersion := uint(0)
		dwoId := uint(0)
		splitDebugFilename := ""
		sysRoot := ""
		sdk := ""
		m.DebugCompileUnit = LLVMDIBuilderCreateCompileUnit(
			m.DebugBuilder,
			LLVMDWARFSourceLanguageOdin,
			diFile,
			producer, uint(len(producer)),
			0,
			flags, uint(len(flags)),
			runtimeVersion,
			splitDebugFilename, uint(len(splitDebugFilename)),
			LLVMDWARFEmissionFull,
			dwoId,
			splitDebugFilename, uint(len(splitDebugFilename)),
			0,
			sysRoot, uint(len(sysRoot)),
			sdk, uint(len(sdk)),
		)
	}

	m.ConstDummyBuilder = LLVMCreateBuilderInContext(m.Ctx)
	if buildContext.OptimizationDisabled {
		LLVMSetModuleInlineAsm(m.Mod, "", 0)
	}

	m.PolymorphicModule = nil

	m.Types = make(map[uint64]LLVMTypeRef)
	m.FuncRawTypes = make(map[uint64]LLVMTypeRef)
	m.Values = make(map[*Entity]lbValue)
	m.SoaValues = make(map[*Entity]lbAddr)
	m.Members = make(map[string]lbValue)
	m.Procedures = make(map[string]*lbProcedure)
	m.ProcedureValues = make(map[LLVMValueRef]*Entity)
	m.ConstStrings = make(map[string]LLVMValueRef)
	m.ConstString16s = make(map[string]LLVMValueRef)
	m.FunctionTypeMap = make(map[uint64]*lbFunctionType)
	m.GenProcs = make(map[string]*lbProcedure)
	m.GlobalProceduresToCreate = make([]*Entity, 0)
	m.GlobalTypesToCreate = make([]*Entity, 0)
	m.GeneratedProcedures = make([]*lbProcedure, 0)
	m.DebugValues = make(map[uintptr]LLVMMetadataRef)
	m.ObjCClasses = make(map[string]lbAddr)
	m.ObjCSelectors = make(map[string]lbAddr)
	m.ObjCIvars = make(map[string]lbAddr)
	m.MapCellInfoMap = make(map[uint64]lbAddr)
	m.MapInfoMap = make(map[uint64]lbAddr)
	m.ExactValueCompoundLiteralAddrMap = make(map[*Ast]lbAddr)
	m.PadTypes = make([]lbPadType, 0)
	m.StructFieldRemapping = make(map[uintptr]lbStructFieldRemapping)

	return 0
}

// =============================================================================
// lb_init_module
// =============================================================================
func lb_init_module(m *lbModule, do_threading bool) {
	if do_threading {
		thread_pool_add_task(lb_init_module_worker_proc, unsafe.Pointer(m))
	} else {
		lb_init_module_worker_proc(unsafe.Pointer(m))
	}
}

// =============================================================================
// lb_init_generator
// =============================================================================
func lb_init_generator(gen *lbGenerator, c *Checker) bool {
	if gen == nil {
		return false
	}
	gen.Info = &c.Info

	linker_data_init(&gen.LinkerData, gen.Info, String{})

	gen.Modules = make(map[uintptr]*lbModule)
	gen.ModulesThroughCtx = make(map[LLVMContextRef]*lbModule)

	gen.UsedModuleCount = 0

	{
		defaultModule := &gen.DefaultModule
		defaultModule.Checker = c
		defaultModule.Gen = gen
		defaultModule.Info = gen.Info
		defaultModule.Pkg = c.Pkg
		defaultModule.File = c.File
		lb_init_module(defaultModule, false)
	}

	{
		gen.Modules[uintptr(unsafe.Pointer(gen.Info))] = &gen.DefaultModule
	}

	for _, pkg := range c.Info.Packages {
		if pkg == c.Pkg {
			continue
		}
		m := new(lbModule)
		m.Checker = c
		m.Gen = gen
		m.Info = gen.Info
		m.Pkg = pkg
		if len(pkg.Files) > 0 {
			m.File = pkg.Files[0]
		}
		lb_init_module(m, false)
		gen.Modules[uintptr(unsafe.Pointer(pkg))] = m
	}

	thread_pool_wait()
	return true
}

// =============================================================================
// lbLoopData
// =============================================================================
type lbLoopData struct {
	IdxAddr lbAddr
	Idx     lbValue
	Body    *lbBlock
	Done    *lbBlock
	Loop    *lbBlock
}

// =============================================================================
// lbCompoundLitElemTempData
// =============================================================================
type lbCompoundLitElemTempData struct {
	Expr       *Ast
	Value      lbValue
	ElemIndex  i64
	ElemLength i64
	Gep        lbValue
}

// =============================================================================
// lb_loop_start
// =============================================================================
func lb_loop_start(p *lbProcedure, count isize, index_type ...*Type) lbLoopData {
	data := lbLoopData{}
	it := t_i32
	if len(index_type) > 0 {
		it = index_type[0]
	}
	max := lb_const_int(p.Module, t_int, u64(count))
	data.IdxAddr = lb_add_local_generated(p, it, true)
	data.Body = lb_create_block(p, "loop.body")
	data.Done = lb_create_block(p, "loop.done")
	data.Loop = lb_create_block(p, "loop.loop")
	lb_emit_jump(p, data.Loop)
	lb_start_block(p, data.Loop)
	data.Idx = lb_addr_load(p, data.IdxAddr)
	cond := lb_emit_comp(p, Token_Lt, data.Idx, max)
	lb_emit_if(p, cond, data.Body, data.Done)
	lb_start_block(p, data.Body)
	return data
}

// =============================================================================
// lb_loop_end
// =============================================================================
func lb_loop_end(p *lbProcedure, data lbLoopData) {
	if data.IdxAddr.Addr.Value != 0 {
		lb_emit_increment(p, data.IdxAddr.Addr)
		lb_emit_jump(p, data.Loop)
		lb_start_block(p, data.Done)
	}
}

// =============================================================================
// lb_make_global_private_const (by LLVMValueRef)
// =============================================================================
func lb_make_global_private_const(global_data LLVMValueRef) {
	LLVMSetLinkage(global_data, LLVMLinkerPrivateLinkage)
	LLVMSetGlobalConstant(global_data, 1)
}

// =============================================================================
// lb_make_global_private_const (by lbAddr)
// =============================================================================
func lb_make_global_private_const_addr(addr lbAddr) {
	lb_make_global_private_const(addr.Addr.Value)
}

// =============================================================================
// llvm_zero
// =============================================================================
func llvm_zero(m *lbModule) LLVMValueRef {
	return LLVMConstInt(lb_type(m, t_int), 0, 0)
}

// =============================================================================
// llvm_zero_type
// =============================================================================
func llvm_zero_type(m *lbModule, typ *Type) LLVMValueRef {
	return LLVMConstInt(lb_type(m, typ), 0, 0)
}

// =============================================================================
// llvm_alloca
// =============================================================================
func llvm_alloca(p *lbProcedure, llvm_type LLVMTypeRef, alignment isize, name ...string) LLVMValueRef {
	if p.DeclBlock == nil {
		return 0
	}
	savedBlock := p.CurrBlock
	savedBuilder := p.Builder

	p.CurrBlock = p.DeclBlock
	LLVMPositionBuilderAtEnd(p.Builder, p.DeclBlock.Block)

	n := ""
	if len(name) > 0 {
		n = name[0]
	}

	var alloca LLVMValueRef
	if alignment > 0 {
		alloca = LLVMBuildAllocaWithAlignment(p.Builder, llvm_type, uint(alignment), n)
	} else {
		alloca = LLVMBuildAlloca(p.Builder, llvm_type, n)
	}

	p.CurrBlock = savedBlock
	LLVMPositionBuilderAtEnd(p.Builder, savedBlock.Block)

	return alloca
}

// =============================================================================
// lb_zero
// =============================================================================
func lb_zero(m *lbModule, typ *Type) lbValue {
	lbTyp := lb_type(m, typ)
	if lbTyp == 0 {
		return lbValue{}
	}
	return lbValue{Value: LLVMConstNull(lbTyp), Type: typ}
}

// =============================================================================
// llvm_const_extract_value (by lbModule)
// =============================================================================
func llvm_const_extract_value(m *lbModule, val LLVMValueRef, indices []uint, numIdx uint) LLVMValueRef {
	if LLVMIsConstant(val) != 0 {
		return LLVMConstExtractValue(val, indices, numIdx)
	}
	return val
}

// =============================================================================
// llvm_const_extract_value (by lbProcedure)
// =============================================================================
func llvm_const_extract_value_proc(p *lbProcedure, val LLVMValueRef, indices []uint, numIdx uint) LLVMValueRef {
	return llvm_const_extract_value(p.Module, val, indices, numIdx)
}

// =============================================================================
// llvm_const_insert_value (by lbModule)
// =============================================================================
func llvm_const_insert_value(m *lbModule, agg LLVMValueRef, elem LLVMValueRef, indices []uint, numIdx uint) LLVMValueRef {
	if LLVMIsConstant(agg) != 0 && LLVMIsConstant(elem) != 0 {
		return LLVMConstInsertValue(agg, elem, indices, numIdx)
	}
	return agg
}

// =============================================================================
// llvm_const_insert_value (by lbProcedure)
// =============================================================================
func llvm_const_insert_value_proc(p *lbProcedure, agg LLVMValueRef, elem LLVMValueRef, indices []uint, numIdx uint) LLVMValueRef {
	return llvm_const_insert_value(p.Module, agg, elem, indices, numIdx)
}

// =============================================================================
// llvm_cstring
// =============================================================================
func llvm_cstring(m *lbModule, str string) LLVMValueRef {
	v := lb_find_or_add_entity_string(m, str, false)
	indices := []uint{0}
	return llvm_const_extract_value(m, v.Value, indices, 1)
}

// =============================================================================
// lb_is_instr_terminating
// =============================================================================
func lb_is_instr_terminating(val LLVMValueRef) bool {
	if val == 0 {
		return false
	}
	opcode := LLVMGetInstructionOpcode(val)
	switch opcode {
	case LLVMRet, LLVMBr, LLVMSwitch, LLVMIndirectBr, LLVMInvoke, LLVMUnreachable, LLVMCleanupRet, LLVMCatchRet, LLVMCatchSwitch, LLVMCallBr:
		return true
	}
	return false
}

// =============================================================================
// lb_module_of_expr
// =============================================================================
func lb_module_of_expr(gen *lbGenerator, expr *Ast) *lbModule {
	if expr == nil {
		return nil
	}
	pkg := expr.Pkg
	if pkg == nil {
		return &gen.DefaultModule
	}
	key := uintptr(unsafe.Pointer(pkg))
	if m, ok := gen.Modules[key]; ok {
		return m
	}
	return &gen.DefaultModule
}

// =============================================================================
// lb_module_of_entity_internal
// =============================================================================
func lb_module_of_entity_internal(gen *lbGenerator, e *Entity) *lbModule {
	if e == nil {
		return &gen.DefaultModule
	}
	if e.Scope == nil {
		return &gen.DefaultModule
	}
	pkg := e.Scope.Pkg
	if pkg == nil {
		return &gen.DefaultModule
	}
	key := uintptr(unsafe.Pointer(pkg))
	if m, ok := gen.Modules[key]; ok {
		return m
	}
	return &gen.DefaultModule
}

// =============================================================================
// lb_module_of_entity
// =============================================================================
func lb_module_of_entity(gen *lbGenerator, e *Entity) *lbModule {
	if gen == nil {
		return nil
	}
	if e == nil {
		return &gen.DefaultModule
	}
	if e.CodeGenModule != nil && e.CodeGenModule.Load() != nil && e.CodeGenModule.Load().Gen == gen {
		if m, ok := gen.Modules[uintptr(unsafe.Pointer(e.CodeGenModule.Load()))]; ok {
			return m
		}
	}
	return lb_module_of_entity_internal(gen, e)
}

// =============================================================================
// lb_addr (lbAddr from lbValue)
// =============================================================================
func lb_addr(addr lbValue) lbAddr {
	return lbAddr{Kind: lbAddr_Default, Addr: addr}
}

// =============================================================================
// lb_addr_map
// =============================================================================
func lb_addr_map(addr lbValue, map_key lbValue, map_type *Type, map_result *Type) lbAddr {
	v := lbAddr{Kind: lbAddr_Map, Addr: addr}
	v.MapKey = &map_key
	v.MapType = map_type
	v.MapResult = map_result
	return v
}

// =============================================================================
// lb_addr_soa_variable
// =============================================================================
func lb_addr_soa_variable(addr lbValue, soa_index lbValue, soa_index_expr *Ast) lbAddr {
	v := lbAddr{Kind: lbAddr_SoaVariable, Addr: addr}
	v.SoaIndex = soa_index
	v.SoaIndexExpr = soa_index_expr
	return v
}

// =============================================================================
// lb_addr_swizzle
// =============================================================================
func lb_addr_swizzle(addr lbValue, swizzle_type *Type, swizzle_count u8, swizzle_indices [4]u8) lbAddr {
	v := lbAddr{Kind: lbAddr_Swizzle, Addr: addr}
	v.SwizzleType = swizzle_type
	v.SwizzleCount = swizzle_count
	v.SwizzleIndices = swizzle_indices
	return v
}

// =============================================================================
// lb_addr_swizzle_large
// =============================================================================
func lb_addr_swizzle_large(addr lbValue, swizzle_type *Type, swizzle_indices []i32) lbAddr {
	v := lbAddr{Kind: lbAddr_SwizzleLarge, Addr: addr}
	v.SwizzleLargeType = swizzle_type
	v.SwizzleLargeIndices = swizzle_indices
	return v
}

// =============================================================================
// lb_addr_bit_field
// =============================================================================
func lb_addr_bit_field(addr lbValue, bit_field_type *Type, bit_field_offset i64, bit_field_size i64) lbAddr {
	v := lbAddr{Kind: lbAddr_BitField, Addr: addr}
	v.BitFieldType = bit_field_type
	v.BitFieldOffset = bit_field_offset
	v.BitFieldSize = bit_field_size
	return v
}

// =============================================================================
// lb_addr_type
// =============================================================================
func lb_addr_type(addr lbAddr) *Type {
	switch addr.Kind {
	case lbAddr_Default:
		return addr.Addr.Type
	case lbAddr_Map:
		return addr.MapResult
	case lbAddr_Context:
		return addr.Addr.Type
	case lbAddr_SoaVariable:
		return addr.Addr.Type
	case lbAddr_Swizzle:
		return addr.SwizzleType
	case lbAddr_SwizzleLarge:
		return addr.SwizzleLargeType
	case lbAddr_BitField:
		return addr.BitFieldType
	}
	return addr.Addr.Type
}

// =============================================================================
// lb_make_soa_pointer
// =============================================================================
func lb_make_soa_pointer(p *lbProcedure, addr lbAddr, index lbValue) lbValue {
	soaType := addr.Addr.Type
	bt := base_type(soaType)
	if bt == nil || bt.Kind != Type_Struct || bt.Struct.SoaKind == StructSoaNone {
		return lbValue{}
	}
	elemPtr := lb_emit_array_ep(p, addr.Addr, index)
	elemPtr.Type = bt.Struct.SoaElem
	return elemPtr
}

// =============================================================================
// lb_addr_get_ptr
// =============================================================================
func lb_addr_get_ptr(p *lbProcedure, addr lbAddr) lbValue {
	switch addr.Kind {
	case lbAddr_Default:
		return addr.Addr

	case lbAddr_Map:
		if addr.MapKey != nil {
			mapPtr := addr.Addr
			key := *addr.MapKey
			internalKey := lb_emit_runtime_call(p, "__odin_map_entry_by_hash", []lbValue{mapPtr, key})
			entryPtr := lbValue{Value: internalKey.Value, Type: addr.MapType}
			return lb_emit_struct_ep(p, entryPtr, 1)
		}
		return addr.Addr

	case lbAddr_Context:
		return addr.Addr

	case lbAddr_SoaVariable:
		val := lb_make_soa_pointer(p, addr, addr.SoaIndex)
		return val

	case lbAddr_Swizzle:
		return addr.Addr

	case lbAddr_SwizzleLarge:
		return addr.Addr

	case lbAddr_BitField:
		return addr.Addr
	}
	return addr.Addr
}

// =============================================================================
// lb_build_addr_ptr
// =============================================================================
func lb_build_addr_ptr(p *lbProcedure, addr lbAddr) lbValue {
	if addr.Kind == lbAddr_SoaVariable {
		return lb_addr_get_ptr(p, addr)
	}
	switch addr.Kind {
	case lbAddr_Default, lbAddr_Map, lbAddr_Context, lbAddr_Swizzle, lbAddr_SwizzleLarge, lbAddr_BitField:
		return addr.Addr
	}
	return addr.Addr
}
