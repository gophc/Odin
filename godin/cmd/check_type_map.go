package cmd

const (
	MAP_CELL_CACHE_LINE_LOG2 = 6
	MAP_CELL_CACHE_LINE_SIZE = 1 << MAP_CELL_CACHE_LINE_LOG2
)

func map_cell_size_and_len(type_ *Type, size_, len_ *int64) {
	elem_sz := type_size_of(type_)
	len_val := int64(1)
	if 0 < elem_sz && elem_sz < MAP_CELL_CACHE_LINE_SIZE {
		len_val = MAP_CELL_CACHE_LINE_SIZE / elem_sz
	}
	size := alignFormula(elem_sz*len_val, MAP_CELL_CACHE_LINE_SIZE)
	if size_ != nil {
		*size_ = size
	}
	if len_ != nil {
		*len_ = len_val
	}
}

func make_optional_ok_type(value *Type, typed ...bool) *Type {
	t := alloc_type_tuple()
	isTyped := true
	if len(typed) > 0 {
		isTyped = typed[0]
	}
	t.Tuple.Variables = make([]*Entity, 2)
	t.Tuple.Variables[0] = alloc_entity_field(nil, BlankToken, value, false, 0)
	boolType := tBool
	if !isTyped {
		boolType = tUntypedBool
	}
	t.Tuple.Variables[1] = alloc_entity_field(nil, BlankToken, boolType, false, 1)
	return t
}

func get_map_cell_type(type_ *Type) *Type {
	var size, len_val int64
	elem_size := type_size_of(type_)
	map_cell_size_and_len(type_, &size, &len_val)

	if size == len_val*elem_size {
		return type_
	}

	padding := size - len_val*elem_size

	s := alloc_type_struct()
	scope := create_scope(nil, nil)
	s.Struct.Fields = make([]*Entity, 2)
	s.Struct.Fields[0] = alloc_entity_field(scope, makeTokenIdentC("v"), alloc_type_array(type_, len_val, nil), false, 0, EntityState_Resolved)
	s.Struct.Fields[1] = alloc_entity_field(scope, makeTokenIdentC("_"), alloc_type_array(tU8, padding, nil), false, 1, EntityState_Resolved)
	s.Struct.Scope = scope
	wait_signal_set(&s.Struct.FieldsWaitSignal)
	_ = type_size_of(s)
	return s
}

func init_map_internal_debug_types(type_ *Type) {
	if type_.Kind != Type_Map {
		gb_assert_handler("Assertion Failure", "type_->kind == Type_Map", "check_type_map.go", 0)
	}
	if tAllocator == nil {
		gb_assert_handler("Assertion Failure", "t_allocator != nullptr", "check_type_map.go", 0)
	}
	if type_.Map.DebugMetadataType != nil {
		return
	}

	key := type_.Map.Key
	value := type_.Map.Value
	if key == nil {
		gb_assert_handler("Assertion Failure", "key != nullptr", "check_type_map.go", 0)
	}
	if value == nil {
		gb_assert_handler("Assertion Failure", "value != nullptr", "check_type_map.go", 0)
	}

	key_cell := get_map_cell_type(key)
	value_cell := get_map_cell_type(value)

	metadata_type := alloc_type_struct()
	metadata_scope := create_scope(nil, nil)
	metadata_type.Struct.Fields = make([]*Entity, 5)
	metadata_type.Struct.Fields[0] = alloc_entity_field(metadata_scope, makeTokenIdentC("key"), key, false, 0, EntityState_Resolved)
	metadata_type.Struct.Fields[1] = alloc_entity_field(metadata_scope, makeTokenIdentC("value"), value, false, 1, EntityState_Resolved)
	metadata_type.Struct.Fields[2] = alloc_entity_field(metadata_scope, makeTokenIdentC("hash"), tUintptr, false, 2, EntityState_Resolved)
	metadata_type.Struct.Fields[3] = alloc_entity_field(metadata_scope, makeTokenIdentC("key_cell"), key_cell, false, 3, EntityState_Resolved)
	metadata_type.Struct.Fields[4] = alloc_entity_field(metadata_scope, makeTokenIdentC("value_cell"), value_cell, false, 4, EntityState_Resolved)
	metadata_type.Struct.Scope = metadata_scope
	metadata_type.Struct.Node = nil
	wait_signal_set(&metadata_type.Struct.FieldsWaitSignal)

	_ = type_size_of(metadata_type)

	metadata_type = alloc_type_pointer(metadata_type)

	scope := create_scope(nil, nil)
	debug_type := alloc_type_struct()
	debug_type.Struct.Fields = make([]*Entity, 3)
	debug_type.Struct.Fields[0] = alloc_entity_field(scope, makeTokenIdentC("data"), metadata_type, false, 0, EntityState_Resolved)
	debug_type.Struct.Fields[1] = alloc_entity_field(scope, makeTokenIdentC("len"), tInt, false, 1, EntityState_Resolved)
	debug_type.Struct.Fields[2] = alloc_entity_field(scope, makeTokenIdentC("allocator"), tAllocator, false, 2, EntityState_Resolved)
	debug_type.Struct.Scope = scope
	debug_type.Struct.Node = nil
	wait_signal_set(&debug_type.Struct.FieldsWaitSignal)

	_ = type_size_of(debug_type)

	type_.Map.DebugMetadataType = debug_type
}

func init_map_internal_types(type_ *Type) {
	if type_.Kind != Type_Map {
		gb_assert_handler("Assertion Failure", "type_->kind == Type_Map", "check_type_map.go", 0)
	}
	if tAllocator == nil {
		gb_assert_handler("Assertion Failure", "t_allocator != nullptr", "check_type_map.go", 0)
	}
	if type_.Map.LookupResultType != nil {
		return
	}

	key := type_.Map.Key
	value := type_.Map.Value
	if key == nil {
		gb_assert_handler("Assertion Failure", "key != nullptr", "check_type_map.go", 0)
	}
	if value == nil {
		gb_assert_handler("Assertion Failure", "value != nullptr", "check_type_map.go", 0)
	}

	type_.Map.LookupResultType = make_optional_ok_type(value)
}

func add_map_key_type_dependencies(ctx *CheckerContext, key *Type) {
	key = core_type(key)

	if is_type_cstring(key) {
		add_package_dependency(ctx, "runtime", "default_hasher_cstring", false)
	} else if is_type_string(key) {
		add_package_dependency(ctx, "runtime", "default_hasher_string", false)
	} else if !is_type_polymorphic(key) {
		if !is_type_comparable(key) {
			return
		}

		if is_type_simple_compare(key) {
			add_package_dependency(ctx, "runtime", "default_hasher", false)
			return
		}

		if key.Kind == Type_Basic {
			if key.Basic.Flags&BasicFlagQuaternion != 0 {
				add_package_dependency(ctx, "runtime", "default_hasher_f64", false)
				add_package_dependency(ctx, "runtime", "default_hasher_quaternion256", false)
				return
			} else if key.Basic.Flags&BasicFlagComplex != 0 {
				add_package_dependency(ctx, "runtime", "default_hasher_f64", false)
				add_package_dependency(ctx, "runtime", "default_hasher_complex128", false)
				return
			} else if key.Basic.Flags&BasicFlagFloat != 0 {
				add_package_dependency(ctx, "runtime", "default_hasher_f64", false)
				return
			}
		}

		if key.Kind == Type_Struct {
			add_package_dependency(ctx, "runtime", "default_hasher", false)
			for _, field := range key.Struct.Fields {
				add_map_key_type_dependencies(ctx, field.Type)
			}
		} else if key.Kind == Type_Union {
			add_package_dependency(ctx, "runtime", "default_hasher", false)
			for _, v := range key.Union.Variants {
				add_map_key_type_dependencies(ctx, v)
			}
		} else if key.Kind == Type_EnumeratedArray {
			add_package_dependency(ctx, "runtime", "default_hasher", false)
			add_map_key_type_dependencies(ctx, key.EnumeratedArray.Elem)
		} else if key.Kind == Type_Array {
			add_package_dependency(ctx, "runtime", "default_hasher", false)
			add_map_key_type_dependencies(ctx, key.Array.Elem)
		}
	}
}

func check_map_type(ctx *CheckerContext, type_ *Type, node *Ast) {
	if type_.Kind != Type_Map {
		gb_assert_handler("Assertion Failure", "type_->kind == Type_Map", "check_type_map.go", 0)
	}
	mt := &node.MapType

	if mt.Key == nil {
		if mt.Value != nil {
			value := check_type(ctx, mt.Value)
			error(node, "Missing map key type, got 'map[]%s'", type_to_string(value))
			return
		}
		error(node, "Missing map key type, got 'map[]T'")
		return
	}

	key := check_type(ctx, mt.Key)
	value := check_type(ctx, mt.Value)

	if !is_type_valid_for_keys(key) {
		if is_type_boolean(key) {
			error(node, "A boolean cannot be used as a key for a map, use an array instead for this case")
		} else {
			error(node, "Invalid type of a key for a map, got '%s'", type_to_string(key))
		}
	}
	if type_size_of(key) == 0 {
		error(node, "Invalid type of a key for a map of size 0, got '%s'", type_to_string(key))
	}

	type_.Map.Key = key
	type_.Map.Value = value

	add_map_key_type_dependencies(ctx, key)

	init_core_map_type(ctx.Checker)
	init_map_internal_types(type_)
}
