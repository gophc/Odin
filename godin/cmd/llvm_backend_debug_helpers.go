package cmd

import (
	"fmt"
	"unsafe"
)

func lb_get_llvm_metadata(m *lbModule, key unsafe.Pointer) LLVMMetadataRef {
	if key == nil {
		return 0
	}
	m.DebugValuesMutex.Lock()
	found, ok := m.DebugValues[uintptr(key)]
	m.DebugValuesMutex.Unlock()
	if ok {
		return found
	}
	return 0
}

func lb_set_llvm_metadata(m *lbModule, key unsafe.Pointer, value LLVMMetadataRef) {
	if key != nil {
		m.DebugValuesMutex.Lock()
		m.DebugValues[uintptr(key)] = value
		m.DebugValuesMutex.Unlock()
	}
}

func lb_add_raddbg_string(m *lbModule, str String) {
	m.Gen.RaddebugSectionStrings.Enqueue(copy_string(permanent_allocator(), str))
}

func lb_add_raddbg_string_cstr(m *lbModule, cstr string) {
	m.Gen.RaddebugSectionStrings.Enqueue(copy_string(permanent_allocator(), make_string_c(cstr)))
}

func lb_add_raddbg_string_concat2(m *lbModule, a, b string) {
	str := concatenate_strings(permanent_allocator(), make_string_c(a), make_string_c(b))
	m.Gen.RaddebugSectionStrings.Enqueue(str)
}

func lb_add_raddbg_string_concat3(m *lbModule, a, b, c string) {
	str := concatenate3_strings(permanent_allocator(), make_string_c(a), make_string_c(b), make_string_c(c))
	m.Gen.RaddebugSectionStrings.Enqueue(str)
}

func lb_get_current_debug_scope(p *lbProcedure) LLVMMetadataRef {
	if p.DebugInfo == 0 {
		panic("missing debug information for " + goStr(p.Name))
	}
	for i := len(p.ScopeStack) - 1; i >= 0; i-- {
		s := p.ScopeStack[i]
		if s == nil {
			continue
		}
		md := lb_get_llvm_metadata(p.Module, unsafe.Pointer(s))
		if md != 0 {
			return md
		}
	}
	return p.DebugInfo
}

func lb_debug_location_from_token_pos(p *lbProcedure, pos TokenPos) LLVMMetadataRef {
	scope := lb_get_current_debug_scope(p)
	if scope == 0 {
		panic("scope is nil for " + goStr(p.Name))
	}
	return LLVMDIBuilderCreateDebugLocation(uint(pos.Line), uint(pos.Column), scope, 0)
}

func lb_debug_location_from_ast(p *lbProcedure, node *Ast) LLVMMetadataRef {
	if node == nil {
		panic("node != nil")
	}
	return lb_debug_location_from_token_pos(p, ast_token(node).Pos)
}

func lb_debug_end_location_from_ast(p *lbProcedure, node *Ast) LLVMMetadataRef {
	if node == nil {
		panic("node != nil")
	}
	return lb_debug_location_from_token_pos(p, ast_end_token(node).Pos)
}

func lb_debug_file_line(m *lbModule, node *Ast, file LLVMMetadataRef, line uint) (LLVMMetadataRef, uint) {
	if file != 0 {
		return file, line
	}
	if node != nil {
		af := thread_safe_get_ast_file_from_id(node.FileID)
		if af != nil {
			file = lb_get_llvm_metadata(m, unsafe.Pointer(af))
		}
		line = uint(ast_token(node).Pos.Line)
	}
	return file, line
}

func lb_debug_procedure_parameters(m *lbModule, typ *Type) LLVMMetadataRef {
	if is_type_proc(typ) {
		return lb_debug_type(m, t_rawptr)
	}
	if typ.Kind == TypeTuple && len(typ.Tuple.Variables) == 1 {
		return lb_debug_procedure_parameters(m, typ.Tuple.Variables[0].Type)
	}
	return lb_debug_type(m, typ)
}

func lb_debug_type_internal_proc(m *lbModule, typ *Type) LLVMMetadataRef {
	_ = typeSizeOf(typ)
	if typ == t_invalid {
		panic("type != t_invalid")
	}
	if typ.Kind != TypeProc {
		panic("type->kind == Type_Proc")
	}

	parameter_count := 1
	for i := int32(0); i < typ.Proc.ParamCount; i++ {
		e := typ.Proc.Params.Tuple.Variables[i]
		if e.Kind == Entity_Variable {
			parameter_count++
		}
	}

	parameters := make([]LLVMMetadataRef, 0, typ.Proc.ParamCount+typ.Proc.ResultCount+2)
	parameters = append(parameters, 0)

	return_is_tuple := false
	if typ.Proc.ResultCount != 0 {
		single_ret := reduceTupleToSingleType(typ.Proc.Results)
		if is_type_proc(single_ret) {
			single_ret = t_rawptr
		}
		if is_type_tuple(single_ret) && is_calling_convention_odin(typ.Proc.CallingConvention) {
			actual := lb_type_internal_for_procedures_raw(m, typ)
			actual = LLVMGetReturnType(actual)
			if actual == 0 {
				parameters[0] = lb_debug_procedure_parameters(m, single_ret)
			} else {
				possible := lb_type(m, typ.Proc.Results)
				if possible == actual {
					parameters[0] = lb_debug_procedure_parameters(m, single_ret)
				} else {
					return_is_tuple = true
				}
			}
		} else {
			parameters[0] = lb_debug_procedure_parameters(m, single_ret)
		}
	}

	for i := int32(0); i < typ.Proc.ParamCount; i++ {
		e := typ.Proc.Params.Tuple.Variables[i]
		if e.Kind != Entity_Variable {
			continue
		}
		parameters = append(parameters, lb_debug_procedure_parameters(m, e.Type))
	}

	if return_is_tuple {
		results := typ.Proc.Results
		if results == nil || results.Kind != TypeTuple {
			panic("results != nullptr && results->kind == Type_Tuple")
		}
		count := len(results.Tuple.Variables)
		parameters[0] = lb_debug_procedure_parameters(m, results.Tuple.Variables[count-1].Type)
		for i := 0; i < count-1; i++ {
			parameters = append(parameters, lb_debug_procedure_parameters(m, results.Tuple.Variables[i].Type))
		}
	}

	if typ.Proc.CallingConvention == ProcCCOdin {
		parameters = append(parameters, lb_debug_type(m, t_context_ptr))
	}

	flags := LLVMDIFlagZero
	if typ.Proc.Diverging {
		flags = LLVMDIFlagNoReturn
	}

	return LLVMDIBuilderCreateSubroutineType(m.DebugBuilder, 0, parameters, uint(len(parameters)), flags)
}

func lb_debug_struct_field(m *lbModule, name String, typ *Type, offset_in_bits uint64) LLVMMetadataRef {
	field_line := uint(1)
	field_flags := LLVMDIFlagZero

	pkg := m.Info.RuntimePackage
	if len(pkg.Files) == 0 {
		panic("pkg->files.count != 0")
	}
	file := lb_get_llvm_metadata(m, unsafe.Pointer(pkg.Files[0]))
	scope := file

	return LLVMDIBuilderCreateMemberType(m.DebugBuilder, scope, goStr(name), uint(name.Len), file, field_line,
		8*uint64(typeSizeOf(typ)), 8*uint32(typeAlignOf(typ)), offset_in_bits,
		field_flags, lb_debug_type(m, typ),
	)
}

func lb_debug_basic_struct(m *lbModule, name String, size_in_bits uint64, align_in_bits uint32, elements []LLVMMetadataRef, element_count uint) LLVMMetadataRef {
	pkg := m.Info.RuntimePackage
	if len(pkg.Files) == 0 {
		panic("pkg->files.count != 0")
	}
	file := lb_get_llvm_metadata(m, unsafe.Pointer(pkg.Files[0]))
	scope := file

	return LLVMDIBuilderCreateStructType(m.DebugBuilder, scope, goStr(name), uint(name.Len), file, 1, size_in_bits, align_in_bits, LLVMDIFlagZero, 0, elements, element_count, 0, 0, "", 0)
}

func lb_debug_struct(m *lbModule, typ *Type, bt *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	if bt.Kind != TypeStruct {
		panic("bt->kind == Type_Struct")
	}

	file, line = lb_debug_file_line(m, bt.Struct.Node, file, line)

	tag := uint(DW_TAG_structure_type)
	if is_type_raw_union(bt) {
		tag = DW_TAG_union_type
	}

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	temp_forward_decl := LLVMDIBuilderCreateReplaceableCompositeType(
		m.DebugBuilder, tag,
		goStr(name), uint(name.Len),
		scope, file, line, 0, size_in_bits, align_in_bits, LLVMDIFlagZero, "", 0,
	)

	lb_set_llvm_metadata(m, unsafe.Pointer(typ), temp_forward_decl)

	typeSetOffsets(bt)

	element_count := uint(len(bt.Struct.Fields))
	elements := make([]LLVMMetadataRef, element_count)

	member_scope := lb_get_llvm_metadata(m, unsafe.Pointer(bt.Struct.Scope))

	for j, f := range bt.Struct.Fields {
		fname := f.Token.String
		field_line := uint(0)
		field_flags := LLVMDIFlagZero
		if bt.Struct.Offsets == nil {
			panic("bt->Struct.offsets != nullptr")
		}
		offset_in_bits := 8 * uint64(bt.Struct.Offsets[j])

		elements[j] = LLVMDIBuilderCreateMemberType(
			m.DebugBuilder,
			member_scope,
			goStr(fname), uint(fname.Len),
			file, field_line,
			8*uint64(typeSizeOf(f.Type)), 8*uint32(typeAlignOf(f.Type)),
			offset_in_bits,
			field_flags,
			lb_debug_type(m, f.Type),
		)
	}

	var final_decl LLVMMetadataRef
	if tag == DW_TAG_union_type {
		final_decl = LLVMDIBuilderCreateUnionType(
			m.DebugBuilder, scope,
			goStr(name), uint(name.Len),
			file, line,
			size_in_bits, align_in_bits,
			LLVMDIFlagZero,
			elements, element_count,
			0,
			"", 0,
		)
	} else {
		final_decl = LLVMDIBuilderCreateStructType(
			m.DebugBuilder, scope,
			goStr(name), uint(name.Len),
			file, line,
			size_in_bits, align_in_bits,
			LLVMDIFlagZero,
			0,
			elements, element_count,
			0,
			0,
			"", 0,
		)
	}

	LLVMMetadataReplaceAllUsesWith(temp_forward_decl, final_decl)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_slice(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeSlice {
		panic("bt->kind == Type_Slice")
	}

	ptr_bits := uint(8 * buildContext.PtrSize)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	temp_forward_decl := LLVMDIBuilderCreateReplaceableCompositeType(
		m.DebugBuilder, DW_TAG_structure_type,
		goStr(name), uint(name.Len),
		scope, file, line, 0, size_in_bits, align_in_bits, LLVMDIFlagZero, "", 0,
	)

	lb_set_llvm_metadata(m, unsafe.Pointer(typ), temp_forward_decl)

	element_count := uint(2)
	var elements [2]LLVMMetadataRef

	var member_scope LLVMMetadataRef

	elem_type := alloc_type_pointer(bt.Slice.Elem)
	elements[0] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"data", 4,
		file, line,
		8*uint64(typeSizeOf(elem_type)), 8*uint32(typeAlignOf(elem_type)),
		0,
		LLVMDIFlagZero, lb_debug_type(m, elem_type),
	)

	elements[1] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"len", 3,
		file, line,
		8*uint64(typeSizeOf(t_int)), 8*uint32(typeAlignOf(t_int)),
		uint64(ptr_bits),
		LLVMDIFlagZero, lb_debug_type(m, t_int),
	)

	final_decl := LLVMDIBuilderCreateStructType(
		m.DebugBuilder, scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		0,
		elements[:], element_count,
		0,
		0,
		"", 0,
	)

	LLVMMetadataReplaceAllUsesWith(temp_forward_decl, final_decl)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_dynamic_array(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeDynamicArray {
		panic("bt->kind == Type_DynamicArray")
	}

	ptr_bits := uint(8 * buildContext.PtrSize)
	int_bits := uint(8 * buildContext.IntSize)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	temp_forward_decl := LLVMDIBuilderCreateReplaceableCompositeType(
		m.DebugBuilder, DW_TAG_structure_type,
		goStr(name), uint(name.Len),
		scope, file, line, 0, size_in_bits, align_in_bits, LLVMDIFlagZero, "", 0,
	)

	lb_set_llvm_metadata(m, unsafe.Pointer(typ), temp_forward_decl)

	element_count := uint(4)
	var elements [4]LLVMMetadataRef

	var member_scope LLVMMetadataRef

	elem_type := alloc_type_pointer(bt.DynamicArray.Elem)
	elements[0] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"data", 4,
		file, line,
		8*uint64(typeSizeOf(elem_type)), 8*uint32(typeAlignOf(elem_type)),
		0,
		LLVMDIFlagZero, lb_debug_type(m, elem_type),
	)

	elements[1] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"len", 3,
		file, line,
		8*uint64(typeSizeOf(t_int)), 8*uint32(typeAlignOf(t_int)),
		uint64(ptr_bits),
		LLVMDIFlagZero, lb_debug_type(m, t_int),
	)

	elements[2] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"cap", 3,
		file, line,
		8*uint64(typeSizeOf(t_int)), 8*uint32(typeAlignOf(t_int)),
		uint64(ptr_bits+int_bits),
		LLVMDIFlagZero, lb_debug_type(m, t_int),
	)

	elements[3] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"allocator", 9,
		file, line,
		8*uint64(typeSizeOf(t_allocator)), 8*uint32(typeAlignOf(t_allocator)),
		uint64(ptr_bits+int_bits+int_bits),
		LLVMDIFlagZero, lb_debug_type(m, t_allocator),
	)

	final_decl := LLVMDIBuilderCreateStructType(
		m.DebugBuilder, scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		0,
		elements[:], element_count,
		0,
		0,
		"", 0,
	)

	LLVMMetadataReplaceAllUsesWith(temp_forward_decl, final_decl)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_fixed_capacity_dynamic_array(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeFixedCapacityDynamicArray {
		panic("bt->kind == Type_FixedCapacityDynamicArray")
	}

	int_bits := uint(8 * buildContext.IntSize)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	temp_forward_decl := LLVMDIBuilderCreateReplaceableCompositeType(
		m.DebugBuilder, DW_TAG_structure_type,
		goStr(name), uint(name.Len),
		scope, file, line, 0, size_in_bits, align_in_bits, LLVMDIFlagZero, "", 0,
	)

	lb_set_llvm_metadata(m, unsafe.Pointer(typ), temp_forward_decl)

	element_count := uint(2)
	var elements [2]LLVMMetadataRef

	var member_scope LLVMMetadataRef

	elem_type := alloc_type_array(bt.FixedCapacityDynamicArray.Elem, bt.FixedCapacityDynamicArray.Capacity)
	elements[0] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"data", 4,
		file, line,
		8*uint64(typeSizeOf(elem_type)), 8*uint32(typeAlignOf(elem_type)),
		0,
		LLVMDIFlagZero, lb_debug_type(m, elem_type),
	)

	len_offset_in_bits := 8 * typeOffsetOf(bt, 1, nil)

	elements[1] = LLVMDIBuilderCreateMemberType(
		m.DebugBuilder, member_scope,
		"len", 3,
		file, line,
		uint64(int_bits), uint32(int_bits),
		uint64(len_offset_in_bits),
		LLVMDIFlagZero, lb_debug_type(m, t_int),
	)

	final_decl := LLVMDIBuilderCreateStructType(
		m.DebugBuilder, scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		0,
		elements[:], element_count,
		0,
		0,
		"", 0,
	)

	LLVMMetadataReplaceAllUsesWith(temp_forward_decl, final_decl)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_union(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeUnion {
		panic("bt->kind == Type_Union")
	}

	file, line = lb_debug_file_line(m, bt.Union.Node, file, line)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	temp_forward_decl := LLVMDIBuilderCreateReplaceableCompositeType(
		m.DebugBuilder, DW_TAG_union_type,
		goStr(name), uint(name.Len),
		scope, file, line, 0, size_in_bits, align_in_bits, LLVMDIFlagZero, "", 0,
	)

	lb_set_llvm_metadata(m, unsafe.Pointer(typ), temp_forward_decl)

	index_offset := isize(1)
	variant_offset := isize(1)
	if is_type_union_maybe_pointer(bt) {
		index_offset = 0
		variant_offset = 0
	} else if bt.Union.Kind == UnionTypeNoNil {
		variant_offset = 0
	}

	member_scope := lb_get_llvm_metadata(m, unsafe.Pointer(bt.Union.Scope))
	element_count := uint(len(bt.Union.Variants))
	if index_offset > 0 {
		if index_offset != 1 {
			panic("index_offset == 1")
		}
		element_count++
	}

	elements := make([]LLVMMetadataRef, element_count)

	if index_offset > 0 {
		tag_type := unionTagType(bt)
		offset_in_bits := 8 * uint64(bt.Union.VariantBlockSize)

		elements[0] = LLVMDIBuilderCreateMemberType(
			m.DebugBuilder, member_scope,
			"tag", 3,
			file, line,
			8*uint64(typeSizeOf(tag_type)), 8*uint32(typeAlignOf(tag_type)),
			offset_in_bits,
			LLVMDIFlagZero, lb_debug_type(m, tag_type),
		)
	}

	for j, variant := range bt.Union.Variants {
		vname := fmt.Sprintf("v%d", variant_offset+j)

		elements[index_offset+j] = LLVMDIBuilderCreateMemberType(
			m.DebugBuilder, member_scope,
			vname, uint(len(vname)),
			file, line,
			8*uint64(typeSizeOf(variant)), 8*uint32(typeAlignOf(variant)),
			0,
			LLVMDIFlagZero, lb_debug_type(m, variant),
		)
	}

	final_decl := LLVMDIBuilderCreateUnionType(
		m.DebugBuilder,
		scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		elements,
		element_count,
		0,
		"", 0,
	)

	LLVMMetadataReplaceAllUsesWith(temp_forward_decl, final_decl)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_bitset(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeBitSet {
		panic("bt->kind == Type_BitSet")
	}

	file, line = lb_debug_file_line(m, bt.BitSet.Node, file, line)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	bit_set_field_type := lb_debug_type(m, t_bool)

	var element_count uint
	var elements []LLVMMetadataRef

	elem := base_type(bt.BitSet.Elem)
	if elem.Kind == TypeEnum {
		element_count = uint(len(elem.Enum.Fields))
		elements = make([]LLVMMetadataRef, element_count)

		for i, f := range elem.Enum.Fields {
			if f.Kind != Entity_Constant {
				panic("f->kind == Entity_Constant")
			}
			val := exact_value_to_i64(f.Constant.Value)
			field_name := f.Token.String
			offset_in_bits := uint64(val - bt.BitSet.Lower)
			elements[i] = LLVMDIBuilderCreateBitFieldMemberType(
				m.DebugBuilder,
				scope,
				goStr(field_name), uint(field_name.Len),
				file, line,
				1,
				offset_in_bits,
				0,
				LLVMDIFlagZero,
				bit_set_field_type,
			)
		}
	} else {
		if !is_type_integer(elem) {
			panic("is_type_integer(elem)")
		}
		count := bt.BitSet.Upper - bt.BitSet.Lower + 1
		if count < 0 {
			panic("0 <= count")
		}
		element_count = uint(count)
		elements = make([]LLVMMetadataRef, element_count)

		for i := uint(0); i < element_count; i++ {
			offset_in_bits := uint64(i)
			val := bt.BitSet.Lower + int64(i)
			vname := fmt.Sprintf("%d", val)
			elements[i] = LLVMDIBuilderCreateBitFieldMemberType(
				m.DebugBuilder,
				scope,
				vname, uint(len(vname)),
				file, line,
				1,
				offset_in_bits,
				0,
				LLVMDIFlagZero,
				bit_set_field_type,
			)
		}
	}

	final_decl := LLVMDIBuilderCreateUnionType(
		m.DebugBuilder,
		scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		elements,
		element_count,
		0,
		"", 0,
	)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_bitfield(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeBitField {
		panic("bt->kind == Type_BitField")
	}

	file, line = lb_debug_file_line(m, bt.BitField.Node, file, line)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	element_count := uint(len(bt.BitField.Fields))
	elements := make([]LLVMMetadataRef, element_count)

	offset_in_bits := uint64(0)
	for i, f := range bt.BitField.Fields {
		bit_size := uint64(bt.BitField.BitSizes[i])
		if f.Kind != Entity_Variable {
			panic("f->kind == Entity_Variable")
		}
		fname := f.Token.String
		elements[i] = LLVMDIBuilderCreateBitFieldMemberType(m.DebugBuilder, scope, goStr(fname), uint(fname.Len), file, line,
			bit_size, offset_in_bits, 0,
			LLVMDIFlagZero, lb_debug_type(m, f.Type),
		)
		offset_in_bits += bit_size
	}

	final_decl := LLVMDIBuilderCreateStructType(
		m.DebugBuilder, scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		LLVMDIFlagZero,
		0,
		elements, element_count,
		0,
		0,
		"", 0,
	)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_enum(m *lbModule, typ *Type, name String, scope LLVMMetadataRef, file LLVMMetadataRef, line uint) LLVMMetadataRef {
	bt := base_type(typ)
	if bt.Kind != TypeEnum {
		panic("bt->kind == Type_Enum")
	}

	file, line = lb_debug_file_line(m, bt.Enum.Node, file, line)

	size_in_bits := 8 * uint64(typeSizeOf(bt))
	align_in_bits := 8 * uint32(typeAlignOf(bt))

	element_count := uint(len(bt.Enum.Fields))
	elements := make([]LLVMMetadataRef, element_count)

	bt_enum := base_enum_type(bt)
	is_unsigned := boolToLLVM(is_type_unsigned(bt_enum))
	for i, f := range bt.Enum.Fields {
		if f.Kind != Entity_Constant {
			panic("f->kind == Entity_Constant")
		}
		enum_name := f.Token.String
		value := exact_value_to_i64(f.Constant.Value)
		elements[i] = LLVMDIBuilderCreateEnumerator(m.DebugBuilder, goStr(enum_name), uint(enum_name.Len), value, is_unsigned)
	}

	class_type := lb_debug_type(m, bt_enum)
	final_decl := LLVMDIBuilderCreateEnumerationType(
		m.DebugBuilder,
		scope,
		goStr(name), uint(name.Len),
		file, line,
		size_in_bits, align_in_bits,
		elements, element_count,
		class_type,
	)
	lb_set_llvm_metadata(m, unsafe.Pointer(typ), final_decl)
	return final_decl
}

func lb_debug_type_basic_type(m *lbModule, name String, size_in_bits uint64, encoding LLVMDWARFTypeEncoding, flags ...LLVMDIFlags) LLVMMetadataRef {
	f := LLVMDIFlagZero
	if len(flags) > 0 {
		f = flags[0]
	}
	basic_type := LLVMDIBuilderCreateBasicType(m.DebugBuilder, goStr(name), uint(name.Len), size_in_bits, encoding, f)
	final_decl := LLVMDIBuilderCreateTypedef(m.DebugBuilder, basic_type, goStr(name), uint(name.Len), 0, 0, 0, uint32(size_in_bits))
	return final_decl
}
