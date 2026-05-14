package cmd

import (
	"strconv"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Missing LLVM C API stubs (not yet in llvm_c.go)
// ---------------------------------------------------------------------------

func LLVMDIBuilderCreateTypedef(Builder LLVMDIBuilderRef, Type LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, Line uint, Scope LLVMMetadataRef, AlignInBits uint32) LLVMMetadataRef {
	return 0
}

func LLVMDIBuilderCreateBitFieldMemberType(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, LineNumber uint, SizeInBits uint64, OffsetInBits uint64, StorageOffsetInBits uint64, Flags LLVMDIFlags, Type LLVMMetadataRef) LLVMMetadataRef {
	return 0
}

func LLVMDIBuilderCreateLabel(Builder LLVMDIBuilderRef, Scope LLVMMetadataRef, Name string, NameLen uint, File LLVMMetadataRef, Line uint, AlwaysPreserve LLVMBool) LLVMMetadataRef {
	return 0
}

func LLVMDIBuilderInsertLabelAtEnd(Builder LLVMDIBuilderRef, Label LLVMMetadataRef, Location LLVMMetadataRef, Block LLVMBasicBlockRef) LLVMDbgRecordRef {
	return 0
}

func LLVMIsABitCastInst(Val LLVMValueRef) LLVMValueRef {
	return 0
}

// ---------------------------------------------------------------------------
// String helper
// ---------------------------------------------------------------------------

func strLit(s string) String {
	return String{Data: strData(s), Len: isize(len(s))}
}

// ---------------------------------------------------------------------------
// lb_debug_type_internal_basic - handles Type_Basic cases for debug info
// ---------------------------------------------------------------------------

func lb_debug_type_internal_basic(m *lbModule, typ *Type, ptr_bits, int_bits uint) LLVMMetadataRef {
	bk := typ.Basic.Kind
	switch bk {
	case BasicLLVMBool:
		return lb_debug_type_basic_type(m, strLit("llvm bool"), 1, LLVMDWARFTypeEncoding_Boolean)
	case BasicBool:
		return lb_debug_type_basic_type(m, strLit("bool"), 8, LLVMDWARFTypeEncoding_Boolean)
	case BasicB8:
		return lb_debug_type_basic_type(m, strLit("b8"), 8, LLVMDWARFTypeEncoding_Boolean)
	case BasicB16:
		return lb_debug_type_basic_type(m, strLit("b16"), 16, LLVMDWARFTypeEncoding_Boolean)
	case BasicB32:
		return lb_debug_type_basic_type(m, strLit("b32"), 32, LLVMDWARFTypeEncoding_Boolean)
	case BasicB64:
		return lb_debug_type_basic_type(m, strLit("b64"), 64, LLVMDWARFTypeEncoding_Boolean)

	case BasicI8:
		return lb_debug_type_basic_type(m, strLit("i8"), 8, LLVMDWARFTypeEncoding_Signed)
	case BasicU8:
		return lb_debug_type_basic_type(m, strLit("u8"), 8, LLVMDWARFTypeEncoding_Unsigned)
	case BasicI16:
		return lb_debug_type_basic_type(m, strLit("i16"), 16, LLVMDWARFTypeEncoding_Signed)
	case BasicU16:
		return lb_debug_type_basic_type(m, strLit("u16"), 16, LLVMDWARFTypeEncoding_Unsigned)
	case BasicI32:
		return lb_debug_type_basic_type(m, strLit("i32"), 32, LLVMDWARFTypeEncoding_Signed)
	case BasicU32:
		return lb_debug_type_basic_type(m, strLit("u32"), 32, LLVMDWARFTypeEncoding_Unsigned)
	case BasicI64:
		return lb_debug_type_basic_type(m, strLit("i64"), 64, LLVMDWARFTypeEncoding_Signed)
	case BasicU64:
		return lb_debug_type_basic_type(m, strLit("u64"), 64, LLVMDWARFTypeEncoding_Unsigned)
	case BasicI128:
		return lb_debug_type_basic_type(m, strLit("i128"), 128, LLVMDWARFTypeEncoding_Signed)
	case BasicU128:
		return lb_debug_type_basic_type(m, strLit("u128"), 128, LLVMDWARFTypeEncoding_Unsigned)

	case BasicRune:
		return lb_debug_type_basic_type(m, strLit("rune"), 32, LLVMDWARFTypeEncoding_Utf)

	case BasicF16:
		return lb_debug_type_basic_type(m, strLit("f16"), 16, LLVMDWARFTypeEncoding_Float)
	case BasicF32:
		return lb_debug_type_basic_type(m, strLit("f32"), 32, LLVMDWARFTypeEncoding_Float)
	case BasicF64:
		return lb_debug_type_basic_type(m, strLit("f64"), 64, LLVMDWARFTypeEncoding_Float)

	case BasicInt:
		return lb_debug_type_basic_type(m, strLit("int"), uint64(int_bits), LLVMDWARFTypeEncoding_Signed)
	case BasicUint:
		return lb_debug_type_basic_type(m, strLit("uint"), uint64(int_bits), LLVMDWARFTypeEncoding_Unsigned)
	case BasicUintptr:
		return lb_debug_type_basic_type(m, strLit("uintptr"), uint64(ptr_bits), LLVMDWARFTypeEncoding_Unsigned)

	case BasicTypeid:
		return lb_debug_type_basic_type(m, strLit("typeid"), 64, LLVMDWARFTypeEncoding_Unsigned)

	// Endian Specific Types
	case BasicI16le:
		return lb_debug_type_basic_type(m, strLit("i16le"), 16, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagLittleEndian)
	case BasicU16le:
		return lb_debug_type_basic_type(m, strLit("u16le"), 16, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagLittleEndian)
	case BasicI32le:
		return lb_debug_type_basic_type(m, strLit("i32le"), 32, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagLittleEndian)
	case BasicU32le:
		return lb_debug_type_basic_type(m, strLit("u32le"), 32, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagLittleEndian)
	case BasicI64le:
		return lb_debug_type_basic_type(m, strLit("i64le"), 64, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagLittleEndian)
	case BasicU64le:
		return lb_debug_type_basic_type(m, strLit("u64le"), 64, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagLittleEndian)
	case BasicI128le:
		return lb_debug_type_basic_type(m, strLit("i128le"), 128, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagLittleEndian)
	case BasicU128le:
		return lb_debug_type_basic_type(m, strLit("u128le"), 128, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagLittleEndian)

	case BasicF16le:
		return lb_debug_type_basic_type(m, strLit("f16le"), 16, LLVMDWARFTypeEncoding_Float, LLVMDIFlagLittleEndian)
	case BasicF32le:
		return lb_debug_type_basic_type(m, strLit("f32le"), 32, LLVMDWARFTypeEncoding_Float, LLVMDIFlagLittleEndian)
	case BasicF64le:
		return lb_debug_type_basic_type(m, strLit("f64le"), 64, LLVMDWARFTypeEncoding_Float, LLVMDIFlagLittleEndian)

	case BasicI16be:
		return lb_debug_type_basic_type(m, strLit("i16be"), 16, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagBigEndian)
	case BasicU16be:
		return lb_debug_type_basic_type(m, strLit("u16be"), 16, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagBigEndian)
	case BasicI32be:
		return lb_debug_type_basic_type(m, strLit("i32be"), 32, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagBigEndian)
	case BasicU32be:
		return lb_debug_type_basic_type(m, strLit("u32be"), 32, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagBigEndian)
	case BasicI64be:
		return lb_debug_type_basic_type(m, strLit("i64be"), 64, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagBigEndian)
	case BasicU64be:
		return lb_debug_type_basic_type(m, strLit("u64be"), 64, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagBigEndian)
	case BasicI128be:
		return lb_debug_type_basic_type(m, strLit("i128be"), 128, LLVMDWARFTypeEncoding_Signed, LLVMDIFlagBigEndian)
	case BasicU128be:
		return lb_debug_type_basic_type(m, strLit("u128be"), 128, LLVMDWARFTypeEncoding_Unsigned, LLVMDIFlagBigEndian)

	case BasicF16be:
		return lb_debug_type_basic_type(m, strLit("f16be"), 16, LLVMDWARFTypeEncoding_Float, LLVMDIFlagBigEndian)
	case BasicF32be:
		return lb_debug_type_basic_type(m, strLit("f32be"), 32, LLVMDWARFTypeEncoding_Float, LLVMDIFlagBigEndian)
	case BasicF64be:
		return lb_debug_type_basic_type(m, strLit("f64be"), 64, LLVMDWARFTypeEncoding_Float, LLVMDIFlagBigEndian)

	case BasicComplex32:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("real"), t_f16, 0*16)
			elements[1] = lb_debug_struct_field(m, strLit("imag"), t_f16, 1*16)
			return lb_debug_basic_struct(m, strLit("complex32"), 32, 16, elements[:], 2)
		}
	case BasicComplex64:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("real"), t_f32, 0*32)
			elements[1] = lb_debug_struct_field(m, strLit("imag"), t_f32, 1*32)
			return lb_debug_basic_struct(m, strLit("complex64"), 64, 32, elements[:], 2)
		}
	case BasicComplex128:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("real"), t_f64, 0*64)
			elements[1] = lb_debug_struct_field(m, strLit("imag"), t_f64, 1*64)
			return lb_debug_basic_struct(m, strLit("complex128"), 128, 64, elements[:], 2)
		}

	case BasicQuaternion64:
		{
			var elements [4]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("imag"), t_f16, 0*16)
			elements[1] = lb_debug_struct_field(m, strLit("jmag"), t_f16, 1*16)
			elements[2] = lb_debug_struct_field(m, strLit("kmag"), t_f16, 2*16)
			elements[3] = lb_debug_struct_field(m, strLit("real"), t_f16, 3*16)
			return lb_debug_basic_struct(m, strLit("quaternion64"), 64, 16, elements[:], 4)
		}
	case BasicQuaternion128:
		{
			var elements [4]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("imag"), t_f32, 0*32)
			elements[1] = lb_debug_struct_field(m, strLit("jmag"), t_f32, 1*32)
			elements[2] = lb_debug_struct_field(m, strLit("kmag"), t_f32, 2*32)
			elements[3] = lb_debug_struct_field(m, strLit("real"), t_f32, 3*32)
			return lb_debug_basic_struct(m, strLit("quaternion128"), 128, 32, elements[:], 4)
		}
	case BasicQuaternion256:
		{
			var elements [4]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("imag"), t_f64, 0*64)
			elements[1] = lb_debug_struct_field(m, strLit("jmag"), t_f64, 1*64)
			elements[2] = lb_debug_struct_field(m, strLit("kmag"), t_f64, 2*64)
			elements[3] = lb_debug_struct_field(m, strLit("real"), t_f64, 3*64)
			return lb_debug_basic_struct(m, strLit("quaternion256"), 256, 64, elements[:], 4)
		}

	case BasicRawptr:
		{
			void_type := lb_debug_type_basic_type(m, strLit("void"), 8, LLVMDWARFTypeEncoding_Unsigned)
			return LLVMDIBuilderCreatePointerType(m.DebugBuilder, void_type, uint64(ptr_bits), uint32(ptr_bits), LLVMDWARFTypeEncoding_Address, "rawptr", 6)
		}
	case BasicString:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("data"), t_u8_ptr, 0)
			elements[1] = lb_debug_struct_field(m, strLit("len"), t_int, uint64(int_bits))
			return lb_debug_basic_struct(m, strLit("string"), 2*uint64(int_bits), uint32(int_bits), elements[:], 2)
		}
	case BasicCstring:
		{
			char_type := lb_debug_type_basic_type(m, strLit("char"), 8, LLVMDWARFTypeEncoding_Unsigned)
			return LLVMDIBuilderCreatePointerType(m.DebugBuilder, char_type, uint64(ptr_bits), uint32(ptr_bits), 0, "cstring", 7)
		}

	case BasicString16:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("data"), t_u16_ptr, 0)
			elements[1] = lb_debug_struct_field(m, strLit("len"), t_int, uint64(int_bits))
			return lb_debug_basic_struct(m, strLit("string16"), 2*uint64(int_bits), uint32(int_bits), elements[:], 2)
		}
	case BasicCstring16:
		{
			char_type := lb_debug_type_basic_type(m, strLit("wchar_t"), 16, LLVMDWARFTypeEncoding_Unsigned)
			return LLVMDIBuilderCreatePointerType(m.DebugBuilder, char_type, uint64(ptr_bits), uint32(ptr_bits), 0, "cstring16", 9)
		}

	case BasicAny:
		{
			var elements [2]LLVMMetadataRef
			elements[0] = lb_debug_struct_field(m, strLit("data"), t_rawptr, 0)
			elements[1] = lb_debug_struct_field(m, strLit("id"), t_typeid, 64)
			return lb_debug_basic_struct(m, strLit("any"), 128, 64, elements[:], 2)
		}

	// Untyped types
	case BasicUntypedBool:
		gb_panic("Basic_UntypedBool")
	case BasicUntypedInteger:
		gb_panic("Basic_UntypedInteger")
	case BasicUntypedFloat:
		gb_panic("Basic_UntypedFloat")
	case BasicUntypedComplex:
		gb_panic("Basic_UntypedComplex")
	case BasicUntypedQuaternion:
		gb_panic("Basic_UntypedQuaternion")
	case BasicUntypedString:
		gb_panic("Basic_UntypedString")
	case BasicUntypedRune:
		gb_panic("Basic_UntypedRune")
	case BasicUntypedNil:
		gb_panic("Basic_UntypedNil")
	case BasicUntypedUninit:
		gb_panic("Basic_UntypedUninit")

	default:
		gb_panic("Basic Unhandled")
	}
	return 0
}

// ---------------------------------------------------------------------------
// lb_debug_type_internal - maps Odin types to DWARF debug metadata
// ---------------------------------------------------------------------------

func lb_debug_type_internal(m *lbModule, typ *Type) LLVMMetadataRef {
	_ = type_size_of(typ)

	ptr_bits := uint(8 * build_context.PtrSize)
	int_bits := uint(8 * build_context.IntSize)

	switch typ.Kind {
	case TypeBasic:
		return lb_debug_type_internal_basic(m, typ, ptr_bits, int_bits)

	case TypeNamed:
		gb_panic("Type_Named should be handled in lb_debug_type separately")

	case TypeSoaPointer:
		return LLVMDIBuilderCreatePointerType(m.DebugBuilder, lb_debug_type(m, typ.SoaPointer.Elem), uint64(ptr_bits), uint32(ptr_bits), 0, "", 0)

	case TypePointer:
		return LLVMDIBuilderCreatePointerType(m.DebugBuilder, lb_debug_type(m, typ.Pointer.Elem), uint64(ptr_bits), uint32(ptr_bits), 0, "", 0)

	case TypeMultiPointer:
		return LLVMDIBuilderCreatePointerType(m.DebugBuilder, lb_debug_type(m, typ.MultiPointer.Elem), uint64(ptr_bits), uint32(ptr_bits), 0, "", 0)

	case TypeArray:
		{
			var subscripts [1]LLVMMetadataRef
			subscripts[0] = LLVMDIBuilderGetOrCreateSubrange(m.DebugBuilder, 0, typ.Array.Count)
			return LLVMDIBuilderCreateArrayType(m.DebugBuilder,
				8*uint64(type_size_of(typ)),
				8*uint32(type_align_of(typ)),
				lb_debug_type(m, typ.Array.Elem),
				subscripts[:], 1)
		}

	case TypeEnumeratedArray:
		{
			var subscripts [1]LLVMMetadataRef
			subscripts[0] = LLVMDIBuilderGetOrCreateSubrange(m.DebugBuilder, 0, typ.EnumeratedArray.Count)

			array_type := LLVMDIBuilderCreateArrayType(m.DebugBuilder,
				8*uint64(type_size_of(typ)),
				8*uint32(type_align_of(typ)),
				lb_debug_type(m, typ.EnumeratedArray.Elem),
				subscripts[:], 1)
			name := temp_canonical_string(typ)
			nameStr := unsafe.String(name.Data, int(name.Len))
			return LLVMDIBuilderCreateTypedef(m.DebugBuilder, array_type, nameStr, uint(len(nameStr)), 0, 0, 0, uint32(8*type_align_of(typ)))
		}

	case TypeMap:
		{
			init_map_internal_debug_types(typ)
			bt := base_type(typ.Map.DebugMetadataType)
			gb_assert_handler(bt.Kind == TypeStruct, "expected struct type for map debug metadata")
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_struct(m, typ, bt, name, 0, 0, 0)
		}

	case TypeStruct:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_struct(m, typ, typ, name, 0, 0, 0)
		}
	case TypeSlice:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_slice(m, typ, name, 0, 0, 0)
		}
	case TypeDynamicArray:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_dynamic_array(m, typ, name, 0, 0, 0)
		}
	case TypeUnion:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_union(m, typ, name, 0, 0, 0)
		}
	case TypeBitSet:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_bitset(m, typ, name, 0, 0, 0)
		}
	case TypeEnum:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_enum(m, typ, name, 0, 0, 0)
		}
	case TypeBitField:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_bitfield(m, typ, name, 0, 0, 0)
		}
	case TypeFixedCapacityDynamicArray:
		{
			name := type_to_canonical_string(permanent_allocator(), typ)
			return lb_debug_fixed_capacity_dynamic_array(m, typ, name, 0, 0, 0)
		}

	case TypeTuple:
		{
			if len(typ.Tuple.Variables) == 1 {
				return lb_debug_type(m, typ.Tuple.Variables[0].Type)
			}
			type_set_offsets(typ)
			var parent_scope LLVMMetadataRef
			var scope LLVMMetadataRef
			var file LLVMMetadataRef
			line := uint(0)
			size_in_bits := 8 * uint64(type_size_of(typ))
			align_in_bits := 8 * uint32(type_align_of(typ))
			flags := LLVMDIFlagZero

			element_count := len(typ.Tuple.Variables)
			elements := make([]LLVMMetadataRef, element_count)

			for i, f := range typ.Tuple.Variables {
				gb_assert_handler(f.Kind == Entity_Variable, "tuple variable expected")
				name := f.Token.String
				field_line := uint(0)
				field_flags := LLVMDIFlagZero
				offset_in_bits := 8 * uint64(typ.Tuple.Offsets[i])
				nameStr := unsafe.String(name.Data, int(name.Len))
				elements[i] = LLVMDIBuilderCreateMemberType(m.DebugBuilder, scope, nameStr, uint(len(nameStr)),
					file, field_line,
					8*uint64(type_size_of(f.Type)), 8*uint32(type_align_of(f.Type)), offset_in_bits,
					field_flags, lb_debug_type(m, f.Type))
			}

			return LLVMDIBuilderCreateStructType(m.DebugBuilder, parent_scope, "", 0, file, line,
				size_in_bits, align_in_bits, flags,
				0, elements, uint(element_count), 0, 0,
				"", 0)
		}

	case TypeProc:
		{
			proc_underlying_type := lb_debug_type_internal_proc(m, typ)
			pointer_type := LLVMDIBuilderCreatePointerType(m.DebugBuilder, proc_underlying_type, uint64(ptr_bits), uint32(ptr_bits), 0, "", 0)
			name := temp_canonical_string(typ)
			nameStr := unsafe.String(name.Data, int(name.Len))
			return LLVMDIBuilderCreateTypedef(m.DebugBuilder, pointer_type, nameStr, uint(len(nameStr)), 0, 0, 0, uint32(8*type_align_of(typ)))
		}

	case TypeSimdVector:
		{
			elem := lb_debug_type(m, typ.SimdVector.Elem)
			var subscripts [1]LLVMMetadataRef
			subscripts[0] = LLVMDIBuilderGetOrCreateSubrange(m.DebugBuilder, 0, typ.SimdVector.Count)
			return LLVMDIBuilderCreateVectorType(m.DebugBuilder,
				8*uint32(type_size_of(typ)), 8*uint32(type_align_of(typ)),
				elem, subscripts[:], 1)
		}

	case TypeMatrix:
		{
			var subscripts [2]LLVMMetadataRef
			subscripts[0] = LLVMDIBuilderGetOrCreateSubrange(m.DebugBuilder, 0, typ.Matrix.RowCount)
			subscripts[1] = LLVMDIBuilderGetOrCreateSubrange(m.DebugBuilder, 0, typ.Matrix.ColumnCount)

			var scope LLVMMetadataRef
			var array_type LLVMMetadataRef

			size_in_bits := 8 * uint64(type_size_of(typ))
			align_in_bits := 8 * uint32(type_align_of(typ))

			if typ.Matrix.IsRowMajor {
				base := LLVMDIBuilderCreateArrayType(m.DebugBuilder,
					8*uint64(type_size_of(typ.Matrix.Elem)*typ.Matrix.ColumnCount),
					8*uint32(type_align_of(typ.Matrix.Elem)),
					lb_debug_type(m, typ.Matrix.Elem),
					subscripts[1:], 1)
				array_type = LLVMDIBuilderCreateArrayType(m.DebugBuilder,
					size_in_bits,
					align_in_bits,
					base,
					subscripts[0:1], 1)
			} else {
				base := LLVMDIBuilderCreateArrayType(m.DebugBuilder,
					8*uint64(type_size_of(typ.Matrix.Elem)*typ.Matrix.RowCount),
					8*uint32(type_align_of(typ.Matrix.Elem)),
					lb_debug_type(m, typ.Matrix.Elem),
					subscripts[0:1], 1)
				array_type = LLVMDIBuilderCreateArrayType(m.DebugBuilder,
					size_in_bits,
					align_in_bits,
					base,
					subscripts[1:], 1)
			}

			var elements [1]LLVMMetadataRef
			elements[0] = LLVMDIBuilderCreateMemberType(m.DebugBuilder, scope,
				"data", 4,
				0, 0,
				size_in_bits, align_in_bits, 0, LLVMDIFlagZero,
				array_type)

			name := temp_canonical_string(typ)
			nameStr := unsafe.String(name.Data, int(name.Len))
			final_decl := LLVMDIBuilderCreateStructType(
				m.DebugBuilder, scope,
				nameStr, uint(len(nameStr)),
				0, 0,
				size_in_bits, align_in_bits,
				LLVMDIFlagZero,
				0,
				elements[:], 1,
				0,
				0,
				"", 0)
			return final_decl
		}
	}

	gb_panic("Invalid type")
	return 0
}

// ---------------------------------------------------------------------------
// lb_get_base_scope_metadata - walks scope chain to find proc or file scope
// ---------------------------------------------------------------------------

func lb_get_base_scope_metadata(m *lbModule, scope *Scope) LLVMMetadataRef {
	for {
		if scope == nil {
			return 0
		}
		if (scope.Flags & ScopeFlag_Proc) != 0 {
			found := lb_get_llvm_metadata(m, scope.ProcedureEntity)
			if found != 0 {
				return found
			}
		}
		if (scope.Flags & ScopeFlag_File) != 0 {
			found := lb_get_llvm_metadata(m, scope.File)
			if found != 0 {
				return found
			}
		}
		scope = scope.Parent
	}
}

// ---------------------------------------------------------------------------
// lb_debug_type - top-level cached type -> debug metadata lookup
// ---------------------------------------------------------------------------

func lb_debug_type(m *lbModule, typ *Type) LLVMMetadataRef {
	gb_assert_handler(typ != nil, "lb_debug_type: typ is nil")

	found := lb_get_llvm_metadata(m, typ)
	if found != 0 {
		return found
	}

	m.DebugValuesMutex.Lock()

	if typ.Kind == TypeNamed {
		var file LLVMMetadataRef
		line := uint(0)
		var scope LLVMMetadataRef

		if typ.Named.TypeName != nil {
			e := typ.Named.TypeName
			scope = lb_get_base_scope_metadata(m, e.Scope)
			if scope != 0 {
				file = LLVMDIScopeGetFile(scope)
			}
			line = uint(e.Token.Pos.Line)
		}

		name := type_to_canonical_string(permanent_allocator(), typ)
		bt := base_type(typ.Named.Base)

		switch bt.Kind {
		default:
			{
				align_in_bits := 8 * uint32(type_align_of(typ))
				debug_bt := lb_debug_type(m, bt)
				nameStr := unsafe.String(name.Data, int(name.Len))
				final_decl := LLVMDIBuilderCreateTypedef(
					m.DebugBuilder,
					debug_bt,
					nameStr, uint(len(nameStr)),
					file, line, scope, align_in_bits)
				lb_set_llvm_metadata(m, typ, final_decl)
				m.DebugValuesMutex.Unlock()
				return final_decl
			}

		case TypeMap:
			{
				init_map_internal_debug_types(bt)
				bt = base_type(bt.Map.DebugMetadataType)
				gb_assert_handler(bt.Kind == TypeStruct, "expected struct type for map debug metadata")
				result := lb_debug_struct(m, typ, bt, name, scope, file, line)
				m.DebugValuesMutex.Unlock()
				return result
			}

		case TypeStruct:
			result := lb_debug_struct(m, typ, bt, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeSlice:
			result := lb_debug_slice(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeDynamicArray:
			result := lb_debug_dynamic_array(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeUnion:
			result := lb_debug_union(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeBitSet:
			result := lb_debug_bitset(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeEnum:
			result := lb_debug_enum(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		case TypeBitField:
			result := lb_debug_bitfield(m, typ, name, scope, file, line)
			m.DebugValuesMutex.Unlock()
			return result
		}
	}

	dt := lb_debug_type_internal(m, typ)
	lb_set_llvm_metadata(m, typ, dt)
	m.DebugValuesMutex.Unlock()
	return dt
}

// ---------------------------------------------------------------------------
// lb_add_debug_local_variable - create debug info for a local variable
// ---------------------------------------------------------------------------

func lb_add_debug_local_variable(p *lbProcedure, ptr LLVMValueRef, typ *Type, token Token) {
	if p.DebugInfo == 0 {
		return
	}
	if typ == nil {
		return
	}
	if typ == t_invalid {
		return
	}
	if p.Body == nil {
		return
	}

	m := p.Module
	name := token.String
	if name.Data == nil || name.Len == 0 || is_blank_ident(token) {
		return
	}

	if lb_get_llvm_metadata(m, ptr) != 0 {
		return
	}

	file_ast := p.Body.FileID
	_ = file_ast

	llvm_scope := lb_get_current_debug_scope(p)
	llvm_file := lb_get_llvm_metadata(m, p.Body.File())
	gb_assert_handler(llvm_scope != 0, "lb_add_debug_local_variable: llvm_scope is nil")
	if llvm_file == 0 {
		llvm_file = LLVMDIScopeGetFile(llvm_scope)
	}
	if llvm_file == 0 {
		return
	}

	alignment_in_bits := uint32(8 * type_align_of(typ))

	flags := LLVMDIFlagZero
	always_preserve := LLVMBool(0)
	if build_context.OptimizationLevel == 0 {
		always_preserve = 1
	}

	debug_type := lb_debug_type(m, typ)

	nameStr := unsafe.String(name.Data, int(name.Len))
	var_info := LLVMDIBuilderCreateAutoVariable(
		m.DebugBuilder, llvm_scope,
		nameStr, uint(len(nameStr)),
		llvm_file, uint(token.Pos.Line),
		debug_type,
		always_preserve, flags, alignment_in_bits)

	storage := ptr
	block := p.CurrBlock.Block
	llvm_debug_loc := lb_debug_location_from_token_pos(p, token.Pos)
	llvm_expr := LLVMDIBuilderCreateExpression(m.DebugBuilder, nil, 0)
	lb_set_llvm_metadata(m, ptr, llvm_expr)
	LLVMDIBuilderInsertDeclareAtEnd(m.DebugBuilder, storage, var_info, llvm_expr, llvm_debug_loc, block)
}

// ---------------------------------------------------------------------------
// lb_add_debug_param_variable - create debug info for a parameter variable
// ---------------------------------------------------------------------------

func lb_add_debug_param_variable(p *lbProcedure, ptr LLVMValueRef, typ *Type, token Token, arg_number uint, block *lbBlock) {
	if p.DebugInfo == 0 {
		return
	}
	if typ == nil {
		return
	}
	if typ == t_invalid {
		return
	}
	if p.Body == nil {
		return
	}

	m := p.Module
	name := token.String
	if name.Data == nil || name.Len == 0 || is_blank_ident(token) {
		return
	}

	if lb_get_llvm_metadata(m, ptr) != 0 {
		return
	}

	file_ast := p.Body.FileID
	_ = file_ast

	llvm_scope := lb_get_current_debug_scope(p)
	llvm_file := lb_get_llvm_metadata(m, p.Body.File())
	gb_assert_handler(llvm_scope != 0, "lb_add_debug_param_variable: llvm_scope is nil")
	if llvm_file == 0 {
		llvm_file = LLVMDIScopeGetFile(llvm_scope)
	}
	if llvm_file == 0 {
		return
	}

	flags := LLVMDIFlagZero
	always_preserve := LLVMBool(0)
	if build_context.OptimizationLevel == 0 {
		always_preserve = 1
	}

	debug_type := lb_debug_type(m, typ)

	nameStr := unsafe.String(name.Data, int(name.Len))
	var_info := LLVMDIBuilderCreateParameterVariable(
		m.DebugBuilder, llvm_scope,
		nameStr, uint(len(nameStr)),
		arg_number,
		llvm_file, uint(token.Pos.Line),
		debug_type,
		always_preserve, flags, 0)

	storage := ptr
	llvm_debug_loc := lb_debug_location_from_token_pos(p, token.Pos)
	llvm_expr := LLVMDIBuilderCreateExpression(m.DebugBuilder, nil, 0)
	lb_set_llvm_metadata(m, ptr, llvm_expr)
	LLVMDIBuilderInsertDeclareAtEnd(m.DebugBuilder, storage, var_info, llvm_expr, llvm_debug_loc, block.Block)
}

// ---------------------------------------------------------------------------
// lb_add_debug_context_variable - create debug info for context variable
// ---------------------------------------------------------------------------

func lb_add_debug_context_variable(p *lbProcedure, ctx lbAddr) {
	if p.DebugInfo == 0 || p.Body == nil {
		return
	}
	loc := LLVMBuilderGetCurrentDebugLocation2(p.Builder)
	if loc == 0 {
		return
	}
	pos := TokenPos{}

	pos.FileID = p.Body.FileID
	pos.Line = int32(LLVMDILocationGetLine(loc))
	pos.Column = int32(LLVMDILocationGetColumn(loc))

	token := Token{}
	token.Kind = TokenContext
	token.String = strLit("context")
	token.Pos = pos

	ptr := ctx.Addr.Value
	for {
		bc := LLVMIsABitCastInst(ptr)
		if bc == 0 {
			break
		}
		ptr = LLVMGetOperand(ptr, 0)
	}

	lb_add_debug_local_variable(p, ptr, t_context, token)
}

// ---------------------------------------------------------------------------
// lb_debug_info_mangle_constant_name - mangle constant entity name
// ---------------------------------------------------------------------------

func lb_debug_info_mangle_constant_name(e *Entity, allocator gbAllocator, did_allocate *bool) String {
	name := e.Token.String
	if e.Pkg != nil && e.Pkg.Name.Len > 0 {
		s := string_canonical_entity_name(allocator, e)
		name = String{Data: strData(s), Len: isize(len(s))}
		if did_allocate != nil {
			*did_allocate = true
		}
	}
	return name
}

// ---------------------------------------------------------------------------
// lb_add_debug_info_global_variable_expr - register global variable expression
// ---------------------------------------------------------------------------

func lb_add_debug_info_global_variable_expr(m *lbModule, name String, dtype LLVMMetadataRef, expr LLVMMetadataRef) {
	var scope LLVMMetadataRef
	var file LLVMMetadataRef
	line := uint(0)
	var decl LLVMMetadataRef

	nameStr := unsafe.String(name.Data, int(name.Len))
	LLVMDIBuilderCreateGlobalVariableExpression(
		m.DebugBuilder, scope,
		nameStr, uint(len(nameStr)),
		"", 0,
		file, line, dtype,
		0, // local to unit
		expr, decl, 8)
}

// ---------------------------------------------------------------------------
// lb_add_debug_info_for_global_constant_internal_i64 - add i64 global constant
// ---------------------------------------------------------------------------

func lb_add_debug_info_for_global_constant_internal_i64(m *lbModule, e *Entity, dtype LLVMMetadataRef, v int64) {
	vStr := strconv.FormatInt(v, 10)
	expr := LLVMDIBuilderCreateConstantValueExpression(m.DebugBuilder, vStr, uint(len(vStr)))

	name := lb_debug_info_mangle_constant_name(e, temporary_allocator(), nil)
	lb_add_debug_info_global_variable_expr(m, name, dtype, expr)

	if (e.Pkg != nil && e.Pkg.Kind == PackageInit) ||
		(e.Scope != nil && (e.Scope.Flags&ScopeFlag_Global) != 0) {
		lb_add_debug_info_global_variable_expr(m, e.Token.String, dtype, expr)
	}
}

// ---------------------------------------------------------------------------
// lb_add_debug_info_for_global_constant_from_entity - add global constant
// ---------------------------------------------------------------------------

func lb_add_debug_info_for_global_constant_from_entity(gen *lbGenerator, e *Entity) {
	if e == nil || e.Kind != EntityConstant {
		return
	}
	if is_blank_ident(e.Token) {
		return
	}
	m := &gen.DefaultModule
	if build_context.UseSeparateModules {
		m = lb_module_of_entity(gen, e)
	}
	if m == nil {
		return
	}

	if is_type_integer(e.Type) {
		value := e.Constant.Value
		if value.Kind == ExactValueInteger {
			var dtype LLVMMetadataRef
			v := int64(0)
			is_signed := false
			if big_int_is_neg(&value.ValueInteger) {
				v = exact_value_to_i64(value)
				is_signed = true
			} else {
				v = int64(exact_value_to_u64(value))
			}
			if is_type_untyped(e.Type) {
				if is_signed {
					dtype = lb_debug_type(m, t_i64)
				} else {
					dtype = lb_debug_type(m, t_u64)
				}
			} else {
				dtype = lb_debug_type(m, e.Type)
			}
			lb_add_debug_info_for_global_constant_internal_i64(m, e, dtype, v)
		}
	} else if is_type_rune(e.Type) {
		value := e.Constant.Value
		if value.Kind == ExactValueInteger {
			dtype := lb_debug_type(m, t_rune)
			v := exact_value_to_i64(value)
			lb_add_debug_info_for_global_constant_internal_i64(m, e, dtype, v)
		}
	} else if is_type_boolean(e.Type) {
		value := e.Constant.Value
		if value.Kind == ExactValueBool {
			dtype := lb_debug_type(m, default_type(e.Type))
			v := int64(0)
			if value.ValueBool {
				v = 1
			}
			lb_add_debug_info_for_global_constant_internal_i64(m, e, dtype, v)
		}
	} else if is_type_enum(e.Type) {
		value := e.Constant.Value
		if value.Kind == ExactValueInteger {
			dtype := lb_debug_type(m, default_type(e.Type))
			v := int64(0)
			if big_int_is_neg(&value.ValueInteger) {
				v = exact_value_to_i64(value)
			} else {
				v = int64(exact_value_to_u64(value))
			}
			lb_add_debug_info_for_global_constant_internal_i64(m, e, dtype, v)
		}
	} else if is_type_pointer(e.Type) {
		value := e.Constant.Value
		if value.Kind == ExactValueInteger {
			dtype := lb_debug_type(m, default_type(e.Type))
			v := int64(exact_value_to_u64(value))
			lb_add_debug_info_for_global_constant_internal_i64(m, e, dtype, v)
		}
	}
}

// ---------------------------------------------------------------------------
// lb_add_debug_label - create debug label metadata for LLVM block
// ---------------------------------------------------------------------------

func lb_add_debug_label(p *lbProcedure, label *Ast, target *lbBlock) {
	if p == nil || p.DebugInfo == 0 {
		return
	}
	if target == nil || label == nil || label.Kind != AstLabel {
		return
	}
	label_token := label.Label.Token
	if is_blank_ident(label_token) {
		return
	}
	m := p.Module
	if m == nil {
		return
	}

	file_ast := label.FileID
	_ = file_ast
	llvm_file := lb_get_llvm_metadata(m, label.File())
	if llvm_file == 0 {
		debugf("llvm file not found for label\n")
		return
	}
	llvm_scope := p.DebugInfo
	if llvm_scope == 0 {
		debugf("llvm scope not found for label\n")
		return
	}
	llvm_debug_loc := lb_debug_location_from_token_pos(p, label_token.Pos)
	llvm_block := target.Block
	if llvm_block == 0 || llvm_debug_loc == 0 {
		return
	}

	nameStr := unsafe.String(label_token.String.Data, int(label_token.String.Len))
	llvm_label := LLVMDIBuilderCreateLabel(
		m.DebugBuilder,
		llvm_scope,
		nameStr, uint(len(nameStr)),
		llvm_file,
		uint(label_token.Pos.Line),
		1) // always preserve

	gb_assert_handler(llvm_label != 0, "llvm_label is nil")
	LLVMDIBuilderInsertLabelAtEnd(
		m.DebugBuilder,
		llvm_label,
		llvm_debug_loc,
		llvm_block)
}
