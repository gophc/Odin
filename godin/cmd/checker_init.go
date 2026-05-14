package cmd

import (
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

func alloc_type_proc_from_types(paramTypes []*Type, returnType *Type, variadic bool, cc ProcCallingConvention) *Type {
	scope := create_scope(nil, nil)
	paramCount := isize(len(paramTypes))
	params := alloc_type(TypeTuple)
	params.tuple.variables = make([]*Entity, paramCount)
	for i, pt := range paramTypes {
		params.tuple.variables[i] = alloc_entity_param(scope, BlankToken, pt, false, false)
	}
	results := alloc_type(TypeTuple)
	results.tuple.variables = make([]*Entity, 1)
	results.tuple.variables[0] = alloc_entity_param(scope, BlankToken, returnType, false, false)
	t := alloc_type_proc(scope, params, paramCount, results, 1, variadic, cc)
	return t
}

func add_dependency(info *CheckerInfo, d *DeclInfo, e *Entity) {
	if atomic.LoadUint32((*uint32)(unsafe.Pointer(&in_single_threaded_checker_stage))) != 0 {
		ptr_set_add(&d.deps, e)
	} else {
		rw_mutex_lock(&d.deps_mutex)
		ptr_set_add(&d.deps, e)
		rw_mutex_unlock(&d.deps_mutex)
	}
}

func add_type_info_dependency(info *CheckerInfo, d *DeclInfo, type_ *Type) {
	if d == nil || type_ == nil {
		return
	}
	if type_.kind == TypeNamed {
		e := type_.named.type_name
		if e.type_name.is_type_alias {
			type_ = type_.named.base
		}
	}
	rw_mutex_lock(&d.type_info_deps_mutex)
	type_set_add(&d.type_info_deps, type_)
	rw_mutex_unlock(&d.type_info_deps_mutex)
}

func get_runtime_package(info *CheckerInfo) *AstPackage {
	name := make_string_c("runtime")
	a := heap_allocator()
	path := get_fullpath_base_collection(a, name, nil)
	found := string_map_get(&info.packages, path)
	if found == nil {
		gb_printf_err("Name: %.*s\n", name.len, name.data)
		gb_printf_err("Fullpath: %.*s\n", path.len, path.data)
		for entry := range info.packages {
			gb_printf_err("%.*s\n", entry.len, entry.data)
		}
		gb_assert_handler("Assertion Failure", "found != nullptr", "checker.cpp", 909, "Missing runtime package %.*s", name.len, name.data)
	}
	return found
}

func get_core_package(info *CheckerInfo, name String) *AstPackage {
	if name == "runtime" {
		return get_runtime_package(info)
	}
	a := heap_allocator()
	path := get_fullpath_core_collection(a, name, nil)
	found := string_map_get(&info.packages, path)
	if found == nil {
		gb_printf_err("Name: %.*s\n", name.len, name.data)
		gb_printf_err("Fullpath: %.*s\n", path.len, path.data)
		for entry := range info.packages {
			gb_printf_err("%.*s\n", entry.len, entry.data)
		}
		gb_assert_handler("Assertion Failure", "found != nullptr", "checker.cpp", 930, "Missing core package %.*s", name.len, name.data)
	}
	return found
}

func try_get_core_package(info *CheckerInfo, name String) *AstPackage {
	if name == "runtime" {
		return get_runtime_package(info)
	}
	a := heap_allocator()
	path := get_fullpath_core_collection(a, name, nil)
	found := string_map_get(&info.packages, path)
	if found == nil {
		return nil
	}
	return found
}

func add_package_dependency(c *CheckerContext, package_name, name string, required bool) {
	n := make_string_c(name)
	key := string_interner_insert(n)
	hash := key.Hash()
	p := get_core_package(c.info, make_string_c(package_name))
	e := scope_lookup(p.scope, key, hash)
	gb_assert_handler("Assertion Failure", "e != nullptr", "checker.cpp", 957, "%s", name)
	gb_assert_handler("Assertion Failure", "c.decl != nullptr", "checker.cpp", 958, 0)
	e.flags |= EntityFlag_Used
	if required {
		e.flags |= EntityFlag_Require
	}
	add_dependency(c.info, c.decl, e)
}

func try_to_add_package_dependency(c *CheckerContext, package_name, name string) {
	n := make_string_c(name)
	key := string_interner_insert(n)
	hash := key.Hash()
	p := get_core_package(c.info, make_string_c(package_name))
	e := scope_lookup(p.scope, key, hash)
	if e == nil {
		return
	}
	gb_assert_handler("Assertion Failure", "c.decl != nullptr", "checker.cpp", 975, 0)
	e.flags |= EntityFlag_Used
	add_dependency(c.info, c.decl, e)
}

func add_declaration_dependency(c *CheckerContext, e *Entity) {
	if e == nil {
		return
	}
	if e.flags&EntityFlag_Disabled != 0 {
		return
	}
	if c.decl != nil {
		add_dependency(c.info, c.decl, e)
	}
}

func add_global_entity(entity *Entity, scope *Scope) *Entity {
	defer func() { entity.state = EntityState_Resolved }()
	name := entity.token.string
	if gb_memchr(name.data, ' ', name.len) {
		return entity
	}
	if scope_insert(scope, entity) {
		compiler_error("double declaration")
	}
	return entity
}

func add_global_constant(name string, type_ *Type, value ExactValue) {
	entity := alloc_entity(Entity_Constant, nil, makeTokenIdentC(name), type_)
	entity.constant.value = value
	add_global_entity(entity, builtin_pkg.scope)
}

func add_global_string_constant(name string, value String) {
	add_global_constant(name, t_untyped_string, exact_value_string(value))
}

func add_global_bool_constant(name string, value bool) {
	add_global_constant(name, t_untyped_bool, exact_value_bool(value))
}

func add_global_type_entity(name String, type_ *Type) {
	add_global_entity(alloc_entity_type_name(nil, makeTokenIdentC(string(name)), type_), builtin_pkg.scope)
}

func create_builtin_package(name string) *AstPackage {
	pkg := permanent_alloc_item[*AstPackage]()
	pkg.name = make_string_c(name)
	pkg.kind = PackageBuiltin
	pkg.scope = create_scope(nil, nil)
	pkg.scope.flags |= ScopeFlag_Pkg | ScopeFlag_Global | ScopeFlag_Builtin
	pkg.scope.pkg = pkg
	return pkg
}

type GlobalEnumValue struct {
	name  string
	value int64
}

func add_global_enum_type(type_name String, values []GlobalEnumValue, enum_type_ **Type) []*Entity {
	scope := create_scope(nil, builtin_pkg.scope)
	entity := alloc_entity_type_name(scope, makeTokenIdentC(string(type_name)), nil, EntityState_Resolved)
	enum_type := alloc_type_enum()
	named_type := alloc_type_named(string(type_name), enum_type, entity)
	set_base_type(named_type, enum_type)
	enum_type.enum.base_type = t_int
	enum_type.enum.scope = scope
	entity.type_ = named_type
	fields := make([]*Entity, len(values))
	for i := range values {
		value := values[i].value
		e := alloc_entity_constant(scope, makeTokenIdentC(values[i].name), named_type, exact_value_i64(value))
		e.flags |= EntityFlag_Visited
		e.state = EntityState_Resolved
		fields[i] = e
		ie := scope_insert(scope, e)
		gb_assert_handler("Assertion Failure", "ie == nullptr", "checker.cpp", 1065, 0)
	}
	enum_type.enum.fields = fields
	enum_type.enum.min_value_index = 0
	enum_type.enum.max_value_index = isize(len(values)) - 1
	enum_type.enum.min_value = &enum_type.enum.fields[enum_type.enum.min_value_index].constant.value
	enum_type.enum.max_value = &enum_type.enum.fields[enum_type.enum.max_value_index].constant.value
	if enum_type_ != nil {
		*enum_type_ = named_type
	}
	return fields
}

func add_global_enum_constant(fields []*Entity, name string, value int64) {
	for _, field := range fields {
		gb_assert_handler("Assertion Failure", "field.kind == Entity_Constant", "checker.cpp", 1082, 0)
		if value == exact_value_to_i64(field.constant.value) {
			add_global_constant(name, field.type_, field.constant.value)
			return
		}
	}
	gb_assert_handler("Panic", 0, "checker.cpp", 1088, "Unfound enum value for global constant: %s %lld", name, value)
}

func add_global_type_name(scope *Scope, type_name String, backing_type *Type) *Type {
	e := alloc_entity_type_name(scope, makeTokenIdentC(string(type_name)), nil, EntityState_Resolved)
	named_type := alloc_type_named(string(type_name), backing_type, e)
	e.type_ = named_type
	set_base_type(named_type, backing_type)
	if scope_insert(scope, e) {
		compiler_error("double declaration of %.*s", e.token.string.len, e.token.string.data)
	}
	return named_type
}

func odin_compile_timestamp() int64 {
	return time.Now().UnixNano()
}

func lb_use_new_pass_system() bool

func get_final_microarchitecture() String

func init_universal() {
	bc := &buildContext
	builtin_pkg = create_builtin_package("builtin")
	intrinsics_pkg = create_builtin_package("intrinsics")
	config_pkg = create_builtin_package("config")

	for i := range basic_types {
		name := basic_types[i].basic.name
		add_global_type_entity(make_string_c(name), &basic_types[i])
	}
	add_global_type_entity(make_string_c("byte"), &basic_types[BasicU8])
	{
		equal_args := []*Type{t_rawptr, t_rawptr}
		t_equal_proc = alloc_type_proc_from_types(equal_args, t_bool, false, ProcCCContextless)
		hasher_args := []*Type{t_rawptr, t_uintptr}
		t_hasher_proc = alloc_type_proc_from_types(hasher_args, t_uintptr, false, ProcCCContextless)
		map_get_args := []*Type{t_rawptr, t_uintptr, t_rawptr}
		t_map_get_proc = alloc_type_proc_from_types(map_get_args, t_rawptr, false, ProcCCContextless)
	}
	add_global_entity(alloc_entity_nil("nil", t_untyped_nil), builtin_pkg.scope)
	add_global_bool_constant("true", true)
	add_global_bool_constant("false", false)
	add_global_string_constant("ODIN_VENDOR", bc.ODINVENDOR)
	add_global_string_constant("ODIN_VERSION", bc.ODINVERSION)
	add_global_string_constant("ODIN_ROOT", bc.ODINROOT)
	add_global_string_constant("ODIN_BUILD_PROJECT_NAME", bc.ODINBUILDPROJECTNAME)
	{
		values := []GlobalEnumValue{
			{"Unknown", int64(WindowsSubsystemUNKNOWN)},
			{"Boot_Application", int64(WindowsSubsystemBOOTAPPLICATION)},
			{"Console", int64(WindowsSubsystemCONSOLE)},
			{"EFI_Application", int64(WindowsSubsystemEFIAPPLICATION)},
			{"EFI_Boot_Service_Driver", int64(WindowsSubsystemEFIBOOTSERVICEDRIVER)},
			{"EFI_Rom", int64(WindowsSubsystemEFIROM)},
			{"EFI_Runtime_Driver", int64(WindowsSubsystemEFIRUNTIMEDRIVER)},
			{"Native", int64(WindowsSubsystemNATIVE)},
			{"Posix", int64(WindowsSubsystemPOSIX)},
			{"Windows", int64(WindowsSubsystemWINDOWS)},
			{"Windows_CE", int64(WindowsSubsystemWINDOWSCE)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Windows_Subsystem_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_WINDOWS_SUBSYSTEM", int64(bc.ODINWINDOWSSUBSYSTEM))
		add_global_string_constant("ODIN_WINDOWS_SUBSYSTEM_STRING", windowsSubsystemNames[bc.ODINWINDOWSSUBSYSTEM])
	}
	{
		values := []GlobalEnumValue{
			{"Unknown", int64(TargetOsInvalid)},
			{"Windows", int64(TargetOsWindows)},
			{"Darwin", int64(TargetOsDarwin)},
			{"Linux", int64(TargetOsLinux)},
			{"FreeBSD", int64(TargetOsFreeBSD)},
			{"Haiku", int64(TargetOsHaiku)},
			{"OpenBSD", int64(TargetOsOpenBSD)},
			{"NetBSD", int64(TargetOsNetBSD)},
			{"WASI", int64(TargetOsWasi)},
			{"JS", int64(TargetOsJs)},
			{"Orca", int64(TargetOsOrca)},
			{"Freestanding", int64(TargetOsFreestanding)},
		}
		fields := add_global_enum_type(make_string_c("Odin_OS_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_OS", int64(bc.Metrics.Os))
		add_global_string_constant("ODIN_OS_STRING", targetOsNames[bc.Metrics.Os])
	}
	{
		values := []GlobalEnumValue{
			{"Unknown", int64(TargetArchInvalid)},
			{"amd64", int64(TargetArchAmd64)},
			{"i386", int64(TargetArchI386)},
			{"arm32", int64(TargetArchArm32)},
			{"arm64", int64(TargetArchArm64)},
			{"wasm32", int64(TargetArchWasm32)},
			{"wasm64p32", int64(TargetArchWasm64p32)},
			{"riscv64", int64(TargetArchRiscv64)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Arch_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_ARCH", int64(bc.Metrics.Arch))
		add_global_string_constant("ODIN_ARCH_STRING", targetArchNames[bc.Metrics.Arch])
	}
	add_global_string_constant("ODIN_MICROARCH_STRING", get_final_microarchitecture())
	{
		values := []GlobalEnumValue{
			{"Executable", int64(BuildModeExecutable)},
			{"Dynamic", int64(BuildModeDynamicLibrary)},
			{"Static", int64(BuildModeStaticLibrary)},
			{"Object", int64(BuildModeObject)},
			{"Assembly", int64(BuildModeAssembly)},
			{"LLVM_IR", int64(BuildModeLLVMIR)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Build_Mode_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_BUILD_MODE", int64(bc.BuildMode))
	}
	{
		values := []GlobalEnumValue{
			{"Little", int64(TargetEndianLittle)},
			{"Big", int64(TargetEndianBig)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Endian_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_ENDIAN", int64(targetEndians[bc.Metrics.Arch]))
		add_global_string_constant("ODIN_ENDIAN_STRING", targetEndianNames[targetEndians[bc.Metrics.Arch]])
	}
	{
		values := []GlobalEnumValue{
			{"Default", int64(SubtargetDefault)},
			{"iPhone", int64(SubtargetIPhone)},
			{"iPhoneSimulator", int64(SubtargetIPhoneSimulator)},
			{"Android", int64(SubtargetAndroid)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Platform_Subtarget_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_PLATFORM_SUBTARGET", int64(selectedSubtarget))
	}
	{
		values := []GlobalEnumValue{
			{"Default", int64(ErrorPosStyleDefault)},
			{"Unix", int64(ErrorPosStyleUnix)},
		}
		fields := add_global_enum_type(make_string_c("Odin_Error_Pos_Style_Type"), values, nil)
		add_global_enum_constant(fields, "ODIN_ERROR_POS_STYLE", int64(buildContext.ODINERRORPOSSTYLE))
	}
	{
		values := []GlobalEnumValue{
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderRelaxed], int64(OdinAtomicMemoryOrderRelaxed)},
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderConsume], int64(OdinAtomicMemoryOrderConsume)},
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderAcquire], int64(OdinAtomicMemoryOrderAcquire)},
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderRelease], int64(OdinAtomicMemoryOrderRelease)},
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderAcqRel], int64(OdinAtomicMemoryOrderAcqRel)},
			{OdinAtomicMemoryOrderStrings[OdinAtomicMemoryOrderSeqCst], int64(OdinAtomicMemoryOrderSeqCst)},
		}
		var enum_type *Type
		add_global_enum_type(make_string_c("Atomic_Memory_Order"), values, &enum_type)
		gb_assert_handler("Assertion Failure", "enum_type.kind == Type_Named", "checker.cpp", 1265, 0)
		t_atomic_memory_order = enum_type
		scope_insert(intrinsics_pkg.scope, t_atomic_memory_order.named.type_name)
	}
	{
		minimum_os_version := int64(0)
		if buildContext.MinimumOSVersionString != "" {
			major := 0
			minor := 0
			revision := 0
			ss := string(buildContext.MinimumOSVersionString)
			fmt.Sscanf(ss, "%d.%d.%d", &major, &minor, &revision)
			minimum_os_version = int64(major*10000 + minor*100 + revision)
		}
		add_global_constant("ODIN_MINIMUM_OS_VERSION", t_untyped_integer, exact_value_i64(minimum_os_version))
	}
	add_global_bool_constant("ODIN_DEBUG", bc.ODINDEBUG)
	add_global_bool_constant("ODIN_DISABLE_ASSERT", bc.ODINDISABLEASSERT)
	add_global_bool_constant("ODIN_DEFAULT_TO_NIL_ALLOCATOR", bc.ODINDEFAULTTONILALLOCATOR)
	add_global_bool_constant("ODIN_NO_BOUNDS_CHECK", buildContext.NoBoundsCheck)
	add_global_bool_constant("ODIN_NO_TYPE_ASSERT", buildContext.NoTypeAssert)
	add_global_bool_constant("ODIN_DEFAULT_TO_PANIC_ALLOCATOR", bc.ODINDEFAULTOPANICALLOCATOR)
	add_global_bool_constant("ODIN_NO_CRT", bc.NoCRT)
	add_global_bool_constant("ODIN_USE_SEPARATE_MODULES", bc.UseSeparateModules)
	add_global_bool_constant("ODIN_TEST", bc.CommandKind&CommandTest != 0)
	add_global_bool_constant("ODIN_NO_ENTRY_POINT", bc.NoEntryPoint)
	add_global_bool_constant("ODIN_FOREIGN_ERROR_PROCEDURES", bc.ODINFOREIGNERRORPROCEDURES)
	add_global_bool_constant("ODIN_NO_RTTI", bc.NoRTTI)
	add_global_bool_constant("ODIN_VALGRIND_SUPPORT", bc.ODINVALGRINDSUPPORT)
	add_global_constant("ODIN_COMPILE_TIMESTAMP", t_untyped_integer, exact_value_i64(odin_compile_timestamp()))
	{
		version := make_string_c("d5fbfeb51")
		add_global_string_constant("ODIN_VERSION_HASH", version)
	}
	{
		f16_supported := lb_use_new_pass_system()
		if is_arch_wasm() {
			f16_supported = false
		} else if buildContext.Metrics.Os == TargetOsDarwin && buildContext.Metrics.Arch == TargetArchAmd64 {
			f16_supported = false
		}
		add_global_bool_constant("__ODIN_LLVM_F16_SUPPORTED", f16_supported)
	}
	{
		values := []GlobalEnumValue{
			{"Address", 0},
			{"Memory", 1},
			{"Thread", 2},
		}
		var enum_type *Type
		add_global_enum_type(make_string_c("Odin_Sanitizer_Flag"), values, &enum_type)
		bit_set_type := alloc_type_bit_set()
		bit_set_type.bit_set.elem = enum_type
		bit_set_type.bit_set.underlying = t_u32
		bit_set_type.bit_set.lower = 0
		bit_set_type.bit_set.upper = 2
		typeSizeOf(bit_set_type)
		type_name := make_string_c("Odin_Sanitizer_Flags")
		scope := create_scope(nil, builtin_pkg.scope)
		entity := alloc_entity_type_name(scope, makeTokenIdentC(string(type_name)), nil, EntityState_Resolved)
		named_type := alloc_type_named(string(type_name), bit_set_type, entity)
		set_base_type(named_type, bit_set_type)
		add_global_constant("ODIN_SANITIZER_FLAGS", named_type, exact_value_u64(uint64(bc.SanitizerFlags)))
	}
	{
		values := []GlobalEnumValue{
			{"None", -1},
			{"Minimal", 0},
			{"Size", 1},
			{"Speed", 2},
			{"Aggressive", 3},
		}
		fields := add_global_enum_type(make_string_c("Odin_Optimization_Mode"), values, nil)
		add_global_enum_constant(fields, "ODIN_OPTIMIZATION_MODE", int64(bc.OptimizationLevel))
	}
	for i := range builtin_procs {
		id := BuiltinProcId(i)
		name := builtin_procs[i].name
		if name != "" {
			key := string_interner_insert(name)
			hash := key.Hash()
			entity := alloc_entity(Entity_Builtin, nil, makeTokenIdentC(string(name)), t_invalid)
			entity.builtin.id = id
			switch builtin_procs[i].pkg {
			case BuiltinProcPkg_builtin:
				add_global_entity(entity, builtin_pkg.scope)
			case BuiltinProcPkg_intrinsics:
				add_global_entity(entity, intrinsics_pkg.scope)
				gb_assert_handler("Assertion Failure", "scope_lookup_current(intrinsics_pkg.scope, key, hash) != nullptr", "checker.cpp", 1378, 0)
			}
		}
	}
	{
		id := BuiltinProcId_expand_values
		name := make_string_c("expand_to_tuple")
		entity := alloc_entity(Entity_Builtin, nil, makeTokenIdentC(string(name)), t_invalid)
		entity.builtin.id = id
		add_global_entity(entity, builtin_pkg.scope)
	}
	defined_values_double_declaration := false
	for entry := range bc.DefinedValues {
		name := entry.key
		value := entry.value
		gb_assert_handler("Assertion Failure", "value.kind != ExactValue_Invalid", "checker.cpp", 1395, 0)
		var type_ *Type
		switch value.kind {
		case ExactValue_Bool:
			type_ = t_untyped_bool
		case ExactValue_String:
			type_ = t_untyped_string
		case ExactValue_Integer:
			type_ = t_untyped_integer
		case ExactValue_Float:
			type_ = t_untyped_float
		}
		gb_assert_handler("Assertion Failure", "type_ != nullptr", "checker.cpp", 1412, 0)
		entity := alloc_entity_constant(nil, makeTokenIdentC(name), type_, value)
		entity.state = EntityState_Resolved
		if scope_insert(config_pkg.scope, entity) {
			error_(entity.token, "'%s' defined as an argument is already declared at the global scope", name)
			defined_values_double_declaration = true
		}
	}
	if defined_values_double_declaration {
		exit_with_errors()
	}
	t_u8_ptr = alloc_type_pointer(t_u8)
	t_u8_multi_ptr = alloc_type_multi_pointer(t_u8)
	t_u16_ptr = alloc_type_pointer(t_u16)
	t_u16_multi_ptr = alloc_type_multi_pointer(t_u16)
	t_int_ptr = alloc_type_pointer(t_int)
	t_i64_ptr = alloc_type_pointer(t_i64)
	t_f64_ptr = alloc_type_pointer(t_f64)
	t_u8_slice = alloc_type_slice(t_u8)
	t_string_slice = alloc_type_slice(t_string)
	{
		t_objc_object = add_global_type_name(intrinsics_pkg.scope, make_string_c("objc_object"), alloc_type_struct_complete())
		t_objc_selector = add_global_type_name(intrinsics_pkg.scope, make_string_c("objc_selector"), alloc_type_struct_complete())
		t_objc_class = add_global_type_name(intrinsics_pkg.scope, make_string_c("objc_class"), alloc_type_struct_complete())
		t_objc_ivar = add_global_type_name(intrinsics_pkg.scope, make_string_c("objc_ivar"), alloc_type_struct_complete())
		t_objc_id = alloc_type_pointer(t_objc_object)
		t_objc_sel = alloc_type_pointer(t_objc_selector)
		t_objc_class2 = alloc_type_pointer(t_objc_class)
		t_objc_ivar2 = alloc_type_pointer(t_objc_ivar)
		t_objc_instancetype = add_global_type_name(intrinsics_pkg.scope, make_string_c("objc_instancetype"), t_objc_id)
	}
	{
		scope := create_scope(nil, nil)
		dummy_field := func(s *Scope, type_ *Type, field_index int32, name string) *Entity {
			tok := BlankToken
			if name != "" {
				tok = makeTokenIdentC(name)
			}
			return alloc_entity_field(s, tok, type_, false, field_index, EntityState_Resolved)
		}
		var fields []*Entity
		switch buildContext.Metrics.Arch {
		case TargetArchAmd64:
			switch buildContext.Metrics.Os {
			case TargetOsFreestanding, TargetOsLinux, TargetOsFreeBSD, TargetOsNetBSD, TargetOsOpenBSD:
				fields = append(fields, dummy_field(scope, t_u32, 0, "gp_offset"))
				fields = append(fields, dummy_field(scope, t_u32, 1, "fp_offset"))
				fields = append(fields, dummy_field(scope, t_rawptr, 2, "overflow_arg_area"))
				fields = append(fields, dummy_field(scope, t_rawptr, 3, "reg_save_area"))
			}
		case TargetArchArm64:
			switch buildContext.Metrics.Os {
			case TargetOsDarwin:
				fields = append(fields, dummy_field(scope, t_rawptr, 0, ""))
			case TargetOsFreestanding, TargetOsLinux, TargetOsFreeBSD, TargetOsNetBSD, TargetOsOpenBSD:
				fields = append(fields, dummy_field(scope, t_rawptr, 0, "__stack"))
				fields = append(fields, dummy_field(scope, t_rawptr, 1, "__gr_top"))
				fields = append(fields, dummy_field(scope, t_rawptr, 2, "__vr_top"))
				fields = append(fields, dummy_field(scope, t_i32, 3, "__gr_offs"))
				fields = append(fields, dummy_field(scope, t_i32, 4, "__vr_offs"))
			}
		}
		if len(fields) == 0 {
			fields = append(fields, dummy_field(scope, t_rawptr, 0, ""))
		}
		va_list_struct := alloc_type_struct_complete()
		va_list_struct.struct_.scope = scope
		va_list_struct.struct_.fields = fields
		gb_assert_handler("Assertion Failure", "typeSizeOf(va_list_struct) > 0", "checker.cpp", 1513, 0)
		t_c_va_list = add_global_type_name(intrinsics_pkg.scope, make_string_c("c_va_list"), va_list_struct)
		t_c_va_list_ptr = alloc_type_pointer(t_c_va_list)
	}
}

func init_checker_info(i *CheckerInfo) {
	a := heap_allocator()
	debugf("[Section] %s\n", "checker info: general")
	if buildContext.ShowMoreTimings {
		timings_start_section(&global_timings, make_string_c("checker info: general"))
	}
	array_init(&i.definitions, a)
	array_init(&i.entities, a)
	map_init(&i.global_untyped)
	string_map_init(&i.foreigns)
	type_set_init(&i.min_dep_type_info_set, 0)
	map_init(&i.min_dep_type_info_index_map)
	string_map_init(&i.files)
	string_map_init(&i.packages)
	array_init(&i.variable_init_order, a)
	array_init(&i.testing_procedures, a, 0, 0)
	array_init(&i.init_procedures, a, 0, 0)
	array_init(&i.fini_procedures, a, 0, 0)
	array_init(&i.required_foreign_imports_through_force, a, 0, 0)
	array_init(&i.defineables, a)
	map_init(&i.objc_msgSend_types)
	mpsc_init(&i.objc_class_implementations, a)
	string_set_init(&i.obcj_class_name_set)
	map_init(&i.objc_method_implementations)
	string_map_init(&i.load_file_cache)
	array_init(&i.all_procedures, a)
	mpsc_init(&i.all_procedures_queue, a)
	mpsc_init(&i.entity_queue, a)
	mpsc_init(&i.definition_queue, a)
	mpsc_init(&i.required_global_variable_queue, a)
	mpsc_init(&i.required_foreign_imports_through_force_queue, a)
	mpsc_init(&i.foreign_imports_to_check_fullpaths, a)
	mpsc_init(&i.foreign_decls_to_check, a)
	mpsc_init(&i.intrinsics_entry_point_usage, a)
	mpsc_init(&i.raddbg_type_views_queue, a)
	array_init(&i.raddbg_type_views, a)
	string_map_init(&i.load_directory_cache)
	map_init(&i.load_directory_map)
}

func destroy_checker_info(i *CheckerInfo) {
	array_free(&i.definitions)
	array_free(&i.entities)
	map_destroy(&i.global_untyped)
	string_map_destroy(&i.foreigns)
	type_set_destroy(&i.min_dep_type_info_set)
	map_destroy(&i.min_dep_type_info_index_map)
	string_map_destroy(&i.files)
	string_map_destroy(&i.packages)
	array_free(&i.variable_init_order)
	array_free(&i.required_foreign_imports_through_force)
	array_free(&i.defineables)
	array_free(&i.all_procedures)
	mpsc_destroy(&i.all_procedures_queue)
	mpsc_destroy(&i.entity_queue)
	mpsc_destroy(&i.definition_queue)
	mpsc_destroy(&i.required_global_variable_queue)
	mpsc_destroy(&i.required_foreign_imports_through_force_queue)
	mpsc_destroy(&i.foreign_imports_to_check_fullpaths)
	mpsc_destroy(&i.foreign_decls_to_check)
	mpsc_destroy(&i.raddbg_type_views_queue)
	array_free(&i.raddbg_type_views)
	map_destroy(&i.objc_msgSend_types)
	string_set_destroy(&i.obcj_class_name_set)
	map_destroy(&i.objc_method_implementations)
	string_map_destroy(&i.load_file_cache)
	string_map_destroy(&i.load_directory_cache)
	map_destroy(&i.load_directory_map)
}

func init_checker_context(ctx *CheckerContext, c *Checker) {
	ctx.checker = c
	ctx.info = &c.info
	ctx.scope = builtin_pkg.scope
	ctx.pkg = builtin_pkg
	ctx.type_path = new_checker_type_path()
	ctx.type_level = 0
}

func destroy_checker_context(ctx *CheckerContext) {
	destroy_checker_type_path(ctx.type_path)
}

func add_curr_ast_file(ctx *CheckerContext, file *AstFile) bool {
	if file != nil {
		ctx.file = file
		ctx.decl = file.pkg.decl_info
		ctx.scope = file.scope
		ctx.pkg = file.pkg
		return true
	}
	return false
}

func reset_checker_context(ctx *CheckerContext, file *AstFile, untyped *UntypedExprInfoMap) {
	if ctx == nil {
		return
	}
	gb_assert_handler("Assertion Failure", "ctx.checker != nullptr", "checker.cpp", 1639, 0)
	mutex_lock(&ctx.mutex)
	type_path := ctx.type_path
	array_clear(type_path)
	ctx.pkg = builtin_pkg
	ctx.scope = builtin_pkg.scope
	ctx.file = nil
	ctx.decl = nil
	ctx.type_path = type_path
	ctx.type_level = 0
	add_curr_ast_file(ctx, file)
	ctx.untyped = untyped
	mutex_unlock(&ctx.mutex)
}

func init_checker(c *Checker) {
	a := heap_allocator()
	debugf("[Section] %s\n", "init checker info")
	if buildContext.ShowMoreTimings {
		timings_start_section(&global_timings, make_string_c("init checker info"))
	}
	init_checker_info(&c.info)
	c.info.checker = c
	debugf("[Section] %s\n", "init proc queues")
	if buildContext.ShowMoreTimings {
		timings_start_section(&global_timings, make_string_c("init proc queues"))
	}
	mpsc_init(&c.procs_with_deferred_to_check, a)
	mpsc_init(&c.procs_with_objc_context_provider_to_check, a)
	array_init(&c.procs_to_check, heap_allocator(), 0, 1<<20)
	array_init(&c.nested_proc_lits, heap_allocator(), 0, 1<<20)
	mpsc_init(&c.global_untyped_queue, a)
	mpsc_init(&c.soa_types_to_complete, a)
	init_checker_context(&c.builtin_ctx, c)
}

func destroy_checker(c *Checker) {
	destroy_checker_info(&c.info)
	destroy_checker_context(&c.builtin_ctx)
	array_free(&c.nested_proc_lits)
	array_free(&c.procs_to_check)
	mpsc_destroy(&c.global_untyped_queue)
	mpsc_destroy(&c.soa_types_to_complete)
}
