package docs

import (
	"encoding/binary"
	"fmt"
)

// ============================================================================
// Type-to-doc conversion (mirrors odin_doc_type / odin_doc_add_entity)
// ============================================================================

// addEntity converts an Entity to the doc format and writes it.
func (w *Writer) addEntity(e *Entity) OdinDocEntityIndex {
	if e == nil {
		return 0
	}

	if prevIndex, ok := w.entityCache[e]; ok {
		return prevIndex
	}

	if e.Pkg != nil {
		if _, ok := w.pkgCache[e.Pkg]; !ok {
			return 0
		}
	}

	docEntity := OdinDocEntity{}
	docIndex, _ := w.writeEntity(&docEntity)
	w.entityCache[e] = docIndex

	var initExpr *Ast
	var comment, docs *CommentGroup
	if e.DeclInfo != nil {
		initExpr = e.DeclInfo.InitExpr
		comment = e.DeclInfo.Comment
		docs = e.DeclInfo.Docs
	}
	if e.Kind == EntityVariable {
		if comment == nil {
			comment = e.Variable.Comment
		}
		if docs == nil {
			docs = e.Variable.Docs
		}
	} else if e.Kind == EntityConstant {
		if comment == nil {
			comment = e.Constant.Comment
		}
		if docs == nil {
			docs = e.Constant.Docs
		}
	}

	name := e.Token.String.String()
	linkName := ""
	pos := e.Token.Pos
	kind := OdinDocEntityInvalid
	flags := OdinDocEntityFlag(0)
	fieldGroupIndex := int32(-1)

	switch e.Kind {
	case EntityInvalid:
		kind = OdinDocEntityInvalid
	case EntityConstant:
		kind = OdinDocEntityConstant
	case EntityVariable:
		kind = OdinDocEntityVariable
	case EntityTypeName:
		kind = OdinDocEntityTypeName
	case EntityProcedure:
		kind = OdinDocEntityProcedure
	case EntityProcGroup:
		kind = OdinDocEntityProcGroup
	case EntityImportName:
		kind = OdinDocEntityImportName
	case EntityLibraryName:
		kind = OdinDocEntityLibraryName
	case EntityBuiltin:
		kind = OdinDocEntityBuiltin
	}

	switch e.Kind {
	case EntityTypeName:
		if e.TypeName.IsTypeAlias {
			flags |= OdinDocEntityFlagTypeAlias
		}
	case EntityVariable:
		if e.Flags&EntityFlagBitFieldField != 0 {
			flags |= OdinDocEntityFlagBitFieldField
		}
		if e.Variable.IsForeign {
			flags |= OdinDocEntityFlagForeign
		}
		if e.Variable.IsExport {
			flags |= OdinDocEntityFlagExport
		}
		if e.Variable.ThreadLocalModel != "" {
			flags |= OdinDocEntityFlagVarThreadLocal
		}
		if e.Flags&EntityFlagStatic != 0 {
			flags |= OdinDocEntityFlagVarStatic
		}
		linkName = e.Variable.LinkName.String()
		if initExpr == nil {
			initExpr = e.Variable.InitExpr
		}
		if e.Flags&EntityFlagBitFieldField != 0 {
			fieldGroupIndex = -int32(e.Variable.BitFieldBitSize)
		} else {
			fieldGroupIndex = e.Variable.FieldGroupIndex
		}
	case EntityConstant:
		fieldGroupIndex = e.Constant.FieldGroupIndex
	case EntityProcedure:
		if e.Procedure.IsForeign {
			flags |= OdinDocEntityFlagForeign
		}
		if e.Procedure.IsExport {
			flags |= OdinDocEntityFlagExport
		}
		linkName = e.Procedure.LinkName.String()
	case EntityBuiltin:
		if int(e.Builtin.ID) < len(BuiltinProcs) {
			bp := BuiltinProcs[e.Builtin.ID]
			pos = TokenPos{}
			name = bp.Name.String()
			switch bp.Pkg {
			case BuiltinProcPkgBuiltin:
				flags |= OdinDocEntityFlagBuiltinPkgBuiltin
			case BuiltinProcPkgIntrinsics:
				flags |= OdinDocEntityFlagBuiltinPkgIntrinsics
			default:
				panic(fmt.Sprintf("Unhandled BuiltinProcPkg: %d", bp.Pkg))
			}
		}
	}

	if e.Flags&EntityFlagUsing != 0 {
		flags |= OdinDocEntityFlagParamUsing
	}
	if e.Flags&EntityFlagConstInput != 0 {
		flags |= OdinDocEntityFlagParamConst
	}
	if e.Flags&EntityFlagEllipsis != 0 {
		flags |= OdinDocEntityFlagParamEllipsis
	}
	if e.Flags&EntityFlagNoAlias != 0 {
		flags |= OdinDocEntityFlagParamNoAlias
	}
	if e.Flags&EntityFlagAnyInt != 0 {
		flags |= OdinDocEntityFlagParamAnyInt
	}
	if e.Flags&EntityFlagByPtr != 0 {
		flags |= OdinDocEntityFlagParamByPtr
	}
	if e.Flags&EntityFlagNoBroadcast != 0 {
		flags |= OdinDocEntityFlagParamNoBroadcast
	}

	if e.Scope != nil && (e.Scope.Flags&uint64(ScopeFlagFile|ScopeFlagPkg)) != 0 && !IsEntityExported(e, true) {
		flags |= OdinDocEntityFlagPrivate
	}

	initString := OdinDocString{}
	if initExpr != nil {
		initString = w.writeExprString(initExpr)
	} else if e.Kind == EntityConstant {
		if e.Constant.Flags&EntConstantFlagImplicitEnumValue != 0 {
			// implicit enum value — no init string
		} else if e.Constant.ParamValue.OriginalAstExpr != nil {
			initString = w.writeExprString(e.Constant.ParamValue.OriginalAstExpr)
		} else {
			initString = w.writeString(e.Constant.Value.Str)
		}
	} else if e.Kind == EntityVariable {
		if e.Variable.ParamValue.OriginalAstExpr != nil {
			initString = w.writeExprString(e.Variable.ParamValue.OriginalAstExpr)
		}
	}

	docEntity.Kind = kind
	docEntity.Flags = uint64(flags)
	docEntity.Pos = w.tokenPosCast(pos)
	docEntity.Name = w.writeString(name)
	docEntity.Type = 0
	docEntity.InitString = initString
	docEntity.Comment = w.writeCommentGroupString(comment)
	docEntity.Docs = w.writeCommentGroupString(docs)
	docEntity.FieldGroupIndex = fieldGroupIndex
	docEntity.ForeignLibrary = 0
	docEntity.LinkName = w.writeString(linkName)

	if e.DeclInfo != nil {
		docEntity.Attributes = w.writeAttributes(e.DeclInfo.Attributes)
	}
	docEntity.GroupedEntities = OdinDocArray{}

	w.updateEntity(docIndex, &docEntity)

	return docIndex
}

func (w *Writer) addEntityAsSlice(e *Entity) OdinDocArray {
	index := w.addEntity(e)
	return w.writeUint32AsSlice(index)
}

// docType converts a Type to the doc format.
func (w *Writer) docType(typ *Type, cache bool) OdinDocTypeIndex {
	if typ == nil {
		return 0
	}

	if typ.Kind == TypeNamed {
		e := typ.Named.TypeName
		if e != nil && e.TypeName.IsTypeAlias {
			typ = typ.Named.Base
		}
	}

	if cache {
		typeHash := hashTypeCanonical(typ)
		if found, ok := w.typeCache[typeHash]; ok {
			return found
		}
	}

	docType := OdinDocType{}
	typeIndex, _ := w.writeType(&docType)

	if cache {
		typeHash := hashTypeCanonical(typ)
		w.typeCache[typeHash] = typeIndex
	}

	switch typ.Kind {
	case TypeBasic:
		docType.Kind = OdinDocTypeBasic
		docType.Name = w.writeString(typ.Basic.Name.String())
		if isTypeUntyped(typ) {
			docType.Flags |= OdinDocTypeFlagBasicUntyped
		}

	case TypeNamed:
		docType.Kind = OdinDocTypeNamed
		docType.Name = w.writeString(typ.Named.Name.String())
		docType.Types = w.docTypeAsSlice(baseType(typ))
		docType.Entities = w.addEntityAsSlice(typ.Named.TypeName)

	case TypeGeneric:
		name := typ.Generic.Name
		if typ.Generic.Entity != nil {
			name = typ.Generic.Entity.Token.String
		}
		docType.Kind = OdinDocTypeGeneric
		docType.Name = w.writeString(name.String())
		if typ.Generic.Specialized != nil {
			docType.Types = w.docTypeAsSliceNoCache(typ.Generic.Specialized)
		}

	case TypePointer:
		docType.Kind = OdinDocTypePointer
		docType.Types = w.docTypeAsSlice(typ.Pointer.Elem)

	case TypeMultiPointer:
		docType.Kind = OdinDocTypeMultiPointer
		docType.Types = w.docTypeAsSlice(typ.MultiPointer.Elem)

	case TypeSoaPointer:
		docType.Kind = OdinDocTypeSoaPointer
		docType.Types = w.docTypeAsSlice(typ.SoaPointer.Elem)

	case TypeArray:
		docType.Kind = OdinDocTypeArray
		docType.ElemCountLen = 1
		docType.ElemCounts[0] = typ.Array.Count
		if typ.Array.GenericCount != nil {
			types := [2]OdinDocTypeIndex{
				w.docType(typ.Array.Elem, true),
				w.docType(typ.Array.GenericCount, true),
			}
			docType.Types = w.writeUint32Slice(types[:])
		} else {
			docType.Types = w.docTypeAsSlice(typ.Array.Elem)
		}

	case TypeEnumeratedArray:
		docType.Kind = OdinDocTypeEnumeratedArray
		docType.ElemCountLen = 1
		docType.ElemCounts[0] = typ.EnumeratedArray.Count
		types := [2]OdinDocTypeIndex{
			w.docType(typ.EnumeratedArray.Index, true),
			w.docType(typ.EnumeratedArray.Elem, true),
		}
		docType.Types = w.writeUint32Slice(types[:])

	case TypeSlice:
		docType.Kind = OdinDocTypeSlice
		docType.Types = w.docTypeAsSlice(typ.Slice.Elem)

	case TypeDynamicArray:
		docType.Kind = OdinDocTypeDynamicArray
		docType.Types = w.docTypeAsSlice(typ.DynamicArray.Elem)

	case TypeFixedCapacityDynArray:
		docType.Kind = OdinDocTypeFixedCapacityDynArray
		docType.ElemCountLen = 1
		docType.ElemCounts[0] = typ.FixedCapacityDynArray.Capacity
		if typ.FixedCapacityDynArray.GenericCapacity != nil {
			types := [2]OdinDocTypeIndex{
				w.docType(typ.FixedCapacityDynArray.Elem, true),
				w.docType(typ.FixedCapacityDynArray.GenericCapacity, true),
			}
			docType.Types = w.writeUint32Slice(types[:])
		} else {
			docType.Types = w.docTypeAsSlice(typ.FixedCapacityDynArray.Elem)
		}

	case TypeMap:
		docType.Kind = OdinDocTypeMap
		types := [2]OdinDocTypeIndex{
			w.docType(typ.Map.Key, true),
			w.docType(typ.Map.Value, true),
		}
		docType.Types = w.writeUint32Slice(types[:])

	case TypeBitField:
		docType.Kind = OdinDocTypeBitField
		fields := make([]uint32, len(typ.BitField.Fields))
		for i, f := range typ.BitField.Fields {
			fields[i] = w.addEntity(f)
		}
		docType.Entities = w.writeUint32Slice(fields)
		docType.Types = w.docTypeAsSlice(typ.BitField.BackingType)

	case TypeStruct:
		writeStructTypeDoc(w, typ, &docType)

	case TypeUnion:
		writeUnionTypeDoc(w, typ, &docType)

	case TypeEnum:
		docType.Kind = OdinDocTypeEnum
		fields := make([]uint32, len(typ.Enum.Fields))
		for i, f := range typ.Enum.Fields {
			fields[i] = w.addEntity(f)
		}
		docType.Entities = w.writeUint32Slice(fields)
		if typ.Enum.BaseType != nil {
			docType.Types = w.docTypeAsSlice(typ.Enum.BaseType)
		}

	case TypeTuple:
		docType.Kind = OdinDocTypeTuple
		variables := make([]uint32, len(typ.Tuple.Variables))
		for i, v := range typ.Tuple.Variables {
			variables[i] = w.addEntity(v)
		}
		docType.Entities = w.writeUint32Slice(variables)

	case TypeProc:
		docType.Kind = OdinDocTypeProc
		if typ.Proc.IsPolymorphic {
			docType.Flags |= OdinDocTypeFlagProcPolymorphic
		}
		if typ.Proc.Diverging {
			docType.Flags |= OdinDocTypeFlagProcDiverging
		}
		if typ.Proc.OptionalOk {
			docType.Flags |= OdinDocTypeFlagProcOptionalOk
		}
		if typ.Proc.Variadic {
			docType.Flags |= OdinDocTypeFlagProcVariadic
		}
		if typ.Proc.CVararg {
			docType.Flags |= OdinDocTypeFlagProcCVararg
		}
		types := [2]OdinDocTypeIndex{
			w.docType(typ.Proc.Params, true),
			w.docType(typ.Proc.Results, true),
		}
		docType.Types = w.writeUint32Slice(types[:])
		cc := int(typ.Proc.CallingConvention)
		if cc < len(ProcCallingConventionStrings) {
			docType.CallingConvention = w.writeString(ProcCallingConventionStrings[cc])
		}

	case TypeBitSet:
		docType.Kind = OdinDocTypeBitSet
		var types [2]OdinDocTypeIndex
		typeCount := 0
		if typ.BitSet.Elem != nil {
			types[typeCount] = w.docType(typ.BitSet.Elem, true)
			typeCount++
		}
		if typ.BitSet.Underlying != nil {
			types[typeCount] = w.docType(typ.BitSet.Underlying, true)
			typeCount++
			docType.Flags |= OdinDocTypeFlagBitSetUnderlyingType
		}
		docType.Types = w.writeUint32Slice(types[:typeCount])
		docType.ElemCountLen = 2
		docType.ElemCounts[0] = typ.BitSet.Lower
		docType.ElemCounts[1] = typ.BitSet.Upper

	case TypeSimdVector:
		docType.Kind = OdinDocTypeSimdVector
		docType.ElemCountLen = 1
		docType.ElemCounts[0] = typ.SimdVector.Count
		docType.Types = w.docTypeAsSlice(typ.SimdVector.Elem)

	case TypeMatrix:
		docType.Kind = OdinDocTypeMatrix
		docType.ElemCountLen = 2
		docType.ElemCounts[0] = typ.Matrix.RowCount
		docType.ElemCounts[1] = typ.Matrix.ColumnCount
		docType.Types = w.docTypeAsSlice(typ.Matrix.Elem)
	}

	w.updateType(typeIndex, &docType)

	return typeIndex
}

func (w *Writer) docTypeAsSlice(typ *Type) OdinDocArray {
	index := w.docType(typ, true)
	return w.writeUint32AsSlice(index)
}

func (w *Writer) docTypeAsSliceNoCache(typ *Type) OdinDocArray {
	index := w.docType(typ, false)
	return w.writeUint32AsSlice(index)
}

func (w *Writer) updateEntity(index uint32, e *OdinDocEntity) {
	if w.state != WriterStateWriting {
		return
	}
	elemSize := binary.Size(OdinDocEntity{})
	offset := w.entities.offset + elemSize*int(index)
	writeStructAt(w.Data, offset, e, w.byteOrder)
}

func (w *Writer) updateType(index uint32, t *OdinDocType) {
	if w.state != WriterStateWriting {
		return
	}
	elemSize := binary.Size(OdinDocType{})
	offset := w.types.offset + elemSize*int(index)
	writeStructAt(w.Data, offset, t, w.byteOrder)
}

// ============================================================================
// Type system helpers
// ============================================================================

// hashTypeCanonical computes a hash for type deduplication.
func hashTypeCanonical(typ *Type) uint64 {
	if typ == nil {
		return 0
	}
	// Simple hash based on type kind and contents.
	// In the real compiler this is a more sophisticated canonical hash.
	h := uint64(typ.Kind)
	h = h*31 + uint64(uintptr(unsafePtrFor(typ)))
	return h
}

// isTypeUntyped checks if a Basic type is untyped.
func isTypeUntyped(typ *Type) bool {
	if typ == nil || typ.Kind != TypeBasic {
		return false
	}
	name := typ.Basic.Name.String()
	switch name {
	case "untyped integer", "untyped float", "untyped complex",
		"untyped string", "untyped bool", "untyped nil",
		"untyped rune":
		return true
	}
	return false
}

// baseType returns the underlying type for named types.
func baseType(typ *Type) *Type {
	if typ == nil {
		return nil
	}
	if typ.Kind == TypeNamed {
		return typ.Named.Base
	}
	return typ
}

func unsafePtrFor(typ *Type) uintptr {
	// Returns a unique-ish identifier for pointer-based hashing.
	return 0
}

// ============================================================================
// Struct and Union type doc helpers
// ============================================================================

func writeStructTypeDoc(w *Writer, typ *Type, docType *OdinDocType) {
	if typ.Struct.SoaKind != StructSoaNone {
		switch typ.Struct.SoaKind {
		case StructSoaFixed:
			docType.Kind = OdinDocTypeSOAStructFixed
			docType.ElemCountLen = 1
			docType.ElemCounts[0] = typ.Struct.SoaCount
		case StructSoaSlice:
			docType.Kind = OdinDocTypeSOAStructSlice
		case StructSoaDynamic:
			docType.Kind = OdinDocTypeSOAStructDynamic
		}
		docType.Types = w.docTypeAsSlice(typ.Struct.SoaElem)
		return
	}

	docType.Kind = OdinDocTypeStruct
	if typ.Struct.IsPolymorphic {
		docType.Flags |= OdinDocTypeFlagStructPolymorphic
	}
	if typ.Struct.IsPacked {
		docType.Flags |= OdinDocTypeFlagStructPacked
	}
	if typ.Struct.IsRawUnion {
		docType.Flags |= OdinDocTypeFlagStructRawUnion
	}
	if typ.Struct.IsAllOrNone {
		docType.Flags |= OdinDocTypeFlagStructAllOrNone
	}

	if typ.Struct.CustomMinFieldAlign > 0 || typ.Struct.CustomMaxFieldAlign > 0 {
		docType.ElemCountLen = 2
		docType.ElemCounts[0] = int64(max32(typ.Struct.CustomMinFieldAlign, 0))
		docType.ElemCounts[1] = int64(max32(typ.Struct.CustomMaxFieldAlign, 0))
	}

	fields := make([]uint32, len(typ.Struct.Fields))
	for i, f := range typ.Struct.Fields {
		fields[i] = w.addEntity(f)
	}
	docType.Entities = w.writeUint32Slice(fields)
	docType.PolymorphicParams = w.docType(typ.Struct.PolymorphicParams, true)

	if typ.Struct.Node != nil && typ.Struct.Node.Kind == AstStructType {
		// In real code this would extract align and where_clauses
	}

	tags := make([]OdinDocString, len(typ.Struct.Fields))
	for i, tag := range typ.Struct.Tags {
		tags[i] = w.writeString(tag.String())
	}
	docType.Tags = w.writeDocStringSlice(tags)
}

func writeUnionTypeDoc(w *Writer, typ *Type, docType *OdinDocType) {
	docType.Kind = OdinDocTypeUnion
	if typ.Union.IsPolymorphic {
		docType.Flags |= OdinDocTypeFlagUnionPolymorphic
	}
	switch typ.Union.Kind {
	case UnionTypeNoNil:
		docType.Flags |= OdinDocTypeFlagUnionNoNil
	case UnionTypeSharedNil:
		docType.Flags |= OdinDocTypeFlagUnionSharedNil
	}

	variants := make([]uint32, len(typ.Union.Variants))
	for i, v := range typ.Union.Variants {
		variants[i] = w.docType(v, true)
	}
	docType.Types = w.writeUint32Slice(variants)
	docType.PolymorphicParams = w.docType(typ.Union.PolymorphicParams, true)

	if typ.Union.Node != nil && typ.Union.Node.Kind == AstUnionType {
		// In real code this would extract align and where_clauses
	}
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
