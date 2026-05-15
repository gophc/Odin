package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func odin_doc_token_pos_cast(w *OdinDocWriter, pos TokenPos) OdinDocPosition {
	file_index := OdinDocFileIndex(0)
	if pos.FileID != 0 {
		file := global_files[pos.FileID]
		if file != nil {
			found, ok := w.FileCache[file]
			if !ok {
				panic("file_index_found != nullptr")
			}
			file_index = found
		}
	}
	return OdinDocPosition{
		File:   file_index,
		Line:   uint32(pos.Line),
		Column: uint32(pos.Column),
		Offset: uint32(pos.Offset),
	}
}

func odin_doc_append_comment_group_string(buf *[]byte, g *CommentGroup) bool {
	if g == nil {
		return false
	}
	total_len := 0
	for _, comment := range g.List {
		s := goStr(comment.String)
		total_len += len(s) + 1
	}
	if total_len <= len(g.List) {
		return false
	}
	count := 0
	for _, comment := range g.List {
		s := goStr(comment.String)
		slash_slash := false
		if len(s) > 1 && s[1] == '/' {
			slash_slash = true
			s = s[2:]
		} else if len(s) > 1 && s[1] == '*' {
			s = s[2:]
			if len(s) >= 2 {
				s = s[:len(s)-2]
			}
		}
		if len(s) > 0 && s[0] == ' ' {
			s = s[1:]
		}
		if slash_slash {
			if strings.HasPrefix(s, "+") {
				continue
			}
			if strings.HasPrefix(s, "@(") {
				continue
			}
		}
		if slash_slash {
			*buf = append(*buf, []byte(s)...)
			*buf = append(*buf, '\n')
			count++
		} else {
			pos := 0
			for pos < len(s) {
				end := pos
				for end < len(s) && s[end] != '\n' {
					end++
				}
				line := s[pos:end]
				pos = end
				trimmed := strings.TrimSpace(line)
				if len(trimmed) == 0 && count == 0 {
					if pos < len(s) {
						pos++
					}
					continue
				}
				if strings.HasPrefix(line, "* ") {
					line = line[2:]
				}
				*buf = append(*buf, []byte(line)...)
				*buf = append(*buf, '\n')
				count++
				if pos < len(s) {
					pos++
				}
			}
		}
	}
	if count > 0 {
		*buf = append(*buf, '\n')
		return true
	}
	return false
}

func odin_doc_pkg_doc_string(w *OdinDocWriter, pkg *AstPackage) OdinDocString {
	if pkg == nil {
		return OdinDocString{}
	}
	buf := make([]byte, 0, 0)
	for _, f := range pkg.Files {
		if f.PkgDecl != nil {
			odin_doc_append_comment_group_string(&buf, f.PkgDecl.PackageDecl.Docs)
		}
	}
	return odin_doc_write_string_without_cache(w, S(string(buf)))
}

func odin_doc_comment_group_string(w *OdinDocWriter, g *CommentGroup) OdinDocString {
	if g == nil {
		return OdinDocString{}
	}
	buf := make([]byte, 0, 0)
	odin_doc_append_comment_group_string(&buf, g)
	return odin_doc_write_string_without_cache(w, S(string(buf)))
}

func odin_doc_expr_string(w *OdinDocWriter, expr *Ast) OdinDocString {
	if expr == nil {
		return OdinDocString{}
	}
	s := write_expr_to_string(
		gb_string_make(permanent_allocator(), ""),
		expr,
		build_context.cmd_doc_flags&CmdDocFlag_Short,
	)
	return odin_doc_write_string(w, make_string_c(s))
}

func odin_doc_attributes(w *OdinDocWriter, attributes []*Ast) OdinDocArray[OdinDocAttribute] {
	count := 0
	for _, attr := range attributes {
		if attr.Kind != Ast_Attribute {
			continue
		}
		count += len(attr.Attribute.Elems)
	}
	attribs := make([]OdinDocAttribute, 0, count)
	for _, attr := range attributes {
		if attr.Kind != Ast_Attribute {
			continue
		}
		for _, elem := range attr.Attribute.Elems {
			name := String{}
			var value *Ast
			switch elem.Kind {
			case Ast_Ident:
				name = elem.Ident.Token.String
			case Ast_Implicit:
				name = elem.Implicit.String
			case Ast_FieldValue:
				fv := &elem.FieldValue
				if fv.Field.Kind == Ast_Ident {
					name = fv.Field.Ident.Token.String
				} else if fv.Field.Kind == Ast_Implicit {
					name = fv.Field.Implicit.String
				}
				value = fv.Value
			default:
				continue
			}
			doc_attrib := OdinDocAttribute{
				Name:  odin_doc_write_string(w, name),
				Value: odin_doc_expr_string(w, value),
			}
			attribs = append(attribs, doc_attrib)
		}
	}
	return odin_write_slice(w, attribs)
}

func odin_doc_where_clauses(w *OdinDocWriter, where_clauses []*Ast) OdinDocArray[OdinDocString] {
	if len(where_clauses) == 0 {
		return OdinDocArray[OdinDocString]{}
	}
	clauses := make([]OdinDocString, len(where_clauses))
	for i, s := range where_clauses {
		clauses[i] = odin_doc_expr_string(w, s)
	}
	return odin_write_slice(w, clauses)
}

func odin_doc_type_as_slice(w *OdinDocWriter, typ *Type, cache ...bool) OdinDocArray[OdinDocTypeIndex] {
	c := true
	if len(cache) > 0 {
		c = cache[0]
	}
	index := odin_doc_type(w, typ, c)
	return odin_write_item_as_slice(w, index)
}

func odin_doc_add_entity_as_slice(w *OdinDocWriter, e *Entity) OdinDocArray[OdinDocEntityIndex] {
	index := odin_doc_add_entity(w, e)
	return odin_write_item_as_slice(w, index)
}

func odin_doc_type(w *OdinDocWriter, typ *Type, cache ...bool) OdinDocTypeIndex {
	if typ == nil {
		return 0
	}
	c := true
	if len(cache) > 0 {
		c = cache[0]
	}
	t := typ
	if t.Kind == Type_Named {
		e := t.Named.TypeName
		if e.TypeName.IsTypeAlias {
			t = t.Named.Base
		}
	}
	var type_hash uint64
	if c {
		type_hash = type_hash_canonical_type(t)
		if found, ok := w.TypeCache[type_hash]; ok {
			return found
		}
	}
	doc_type := OdinDocType{}
	var dst *OdinDocType
	type_index := odin_doc_write_item(w, &w.Types, &doc_type, &dst)
	if c {
		w.TypeCache[type_hash] = type_index
	}
	switch t.Kind {
	case Type_Basic:
		doc_type.Kind = OdinDocTypeBasic
		doc_type.Name = odin_doc_write_string(w, t.Basic.Name)
		if is_type_untyped(t) {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Basic_untyped)
		}
	case Type_Named:
		doc_type.Kind = OdinDocTypeNamed
		doc_type.Name = odin_doc_write_string(w, t.Named.Name)
		doc_type.Types = odin_doc_type_as_slice(w, base_type(t))
		doc_type.Entities = odin_doc_add_entity_as_slice(w, t.Named.TypeName)
	case Type_Generic:
		name := t.Generic.Name
		if t.Generic.Entity != nil {
			name = t.Generic.Entity.Token.String
		}
		doc_type.Kind = OdinDocTypeGeneric
		doc_type.Name = odin_doc_write_string(w, name)
		if t.Generic.Specialized != nil {
			doc_type.Types = odin_doc_type_as_slice(w, t.Generic.Specialized, false)
		}
	case Type_Pointer:
		doc_type.Kind = OdinDocTypePointer
		doc_type.Types = odin_doc_type_as_slice(w, t.Pointer.Elem)
	case Type_MultiPointer:
		doc_type.Kind = OdinDocTypeMultiPointer
		doc_type.Types = odin_doc_type_as_slice(w, t.MultiPointer.Elem)
	case Type_SoaPointer:
		doc_type.Kind = OdinDocTypeSoaPointer
		doc_type.Types = odin_doc_type_as_slice(w, t.SoaPointer.Elem)
	case Type_Array:
		doc_type.Kind = OdinDocTypeArray
		doc_type.ElemCountLen = 1
		doc_type.ElemCounts[0] = t.Array.Count
		if t.Array.GenericCount != nil {
			types := [2]OdinDocTypeIndex{
				odin_doc_type(w, t.Array.Elem),
				odin_doc_type(w, t.Array.GenericCount),
			}
			doc_type.Types = odin_write_slice(w, types[:])
		} else {
			doc_type.Types = odin_doc_type_as_slice(w, t.Array.Elem)
		}
	case Type_EnumeratedArray:
		doc_type.Kind = OdinDocTypeEnumeratedArray
		doc_type.ElemCountLen = 1
		doc_type.ElemCounts[0] = t.EnumeratedArray.Count
		{
			types := [2]OdinDocTypeIndex{
				odin_doc_type(w, t.EnumeratedArray.Index),
				odin_doc_type(w, t.EnumeratedArray.Elem),
			}
			doc_type.Types = odin_write_slice(w, types[:])
		}
	case Type_Slice:
		doc_type.Kind = OdinDocTypeSlice
		doc_type.Types = odin_doc_type_as_slice(w, t.Slice.Elem)
	case Type_DynamicArray:
		doc_type.Kind = OdinDocTypeDynamicArray
		doc_type.Types = odin_doc_type_as_slice(w, t.DynamicArray.Elem)
	case Type_FixedCapacityDynamicArray:
		doc_type.Kind = OdinDocTypeFixedCapacityDynamicArray
		doc_type.ElemCountLen = 1
		doc_type.ElemCounts[0] = t.FixedCapacityDynamicArray.Capacity
		if t.FixedCapacityDynamicArray.GenericCapacity != nil {
			types := [2]OdinDocTypeIndex{
				odin_doc_type(w, t.FixedCapacityDynamicArray.Elem),
				odin_doc_type(w, t.FixedCapacityDynamicArray.GenericCapacity),
			}
			doc_type.Types = odin_write_slice(w, types[:])
		} else {
			doc_type.Types = odin_doc_type_as_slice(w, t.FixedCapacityDynamicArray.Elem)
		}
	case Type_Map:
		doc_type.Kind = OdinDocTypeMap
		{
			types := [2]OdinDocTypeIndex{
				odin_doc_type(w, t.Map.Key),
				odin_doc_type(w, t.Map.Value),
			}
			doc_type.Types = odin_write_slice(w, types[:])
		}
	case Type_BitField:
		doc_type.Kind = OdinDocTypeBitField
		{
			fields := make([]OdinDocEntityIndex, len(t.BitField.Fields))
			for i, field := range t.BitField.Fields {
				fields[i] = odin_doc_add_entity(w, field)
			}
			doc_type.Entities = odin_write_slice(w, fields)
			doc_type.Types = odin_doc_type_as_slice(w, t.BitField.BackingType)
		}
	case Type_Struct:
		if t.Struct.SoaKind != StructSoa_None {
			switch t.Struct.SoaKind {
			case StructSoa_Fixed:
				doc_type.Kind = OdinDocTypeSOAStructFixed
				doc_type.ElemCountLen = 1
				doc_type.ElemCounts[0] = t.Struct.SoaCount
			case StructSoa_Slice:
				doc_type.Kind = OdinDocTypeSOAStructSlice
			case StructSoa_Dynamic:
				doc_type.Kind = OdinDocTypeSOAStructDynamic
			}
			doc_type.Types = odin_doc_type_as_slice(w, t.Struct.SoaElem)
		} else {
			doc_type.Kind = OdinDocTypeStruct
			if t.Struct.IsPolymorphic {
				doc_type.Flags |= uint32(OdinDocTypeFlag_Struct_polymorphic)
			}
			if t.Struct.IsPacked {
				doc_type.Flags |= uint32(OdinDocTypeFlag_Struct_packed)
			}
			if t.Struct.IsRawUnion {
				doc_type.Flags |= uint32(OdinDocTypeFlag_Struct_raw_union)
			}
			if t.Struct.IsAllOrNone {
				doc_type.Flags |= uint32(OdinDocTypeFlag_Struct_all_or_none)
			}
			if t.Struct.CustomMinFieldAlign > 0 || t.Struct.CustomMaxFieldAlign > 0 {
				doc_type.ElemCountLen = 2
				if t.Struct.CustomMinFieldAlign > 0 {
					doc_type.ElemCounts[0] = int64(t.Struct.CustomMinFieldAlign)
				} else {
					doc_type.ElemCounts[0] = 0
				}
				if t.Struct.CustomMaxFieldAlign > 0 {
					doc_type.ElemCounts[1] = int64(t.Struct.CustomMaxFieldAlign)
				} else {
					doc_type.ElemCounts[1] = 0
				}
			}
			{
				fields := make([]OdinDocEntityIndex, len(t.Struct.Fields))
				for i, field := range t.Struct.Fields {
					fields[i] = odin_doc_add_entity(w, field)
				}
				doc_type.Entities = odin_write_slice(w, fields)
			}
			doc_type.PolymorphicParams = odin_doc_type(w, t.Struct.PolymorphicParams)
			if t.Struct.Node != nil {
				st := &t.Struct.Node.StructType
				if st.Align != nil {
					doc_type.CustomAlign = odin_doc_expr_string(w, st.Align)
				}
				doc_type.WhereClauses = odin_doc_where_clauses(w, st.WhereClauses[:])
			}
			{
				tags := make([]OdinDocString, len(t.Struct.Tags))
				for i, tag := range t.Struct.Tags {
					tags[i] = odin_doc_write_string(w, tag)
				}
				doc_type.Tags = odin_write_slice(w, tags)
			}
		}
	case Type_Union:
		doc_type.Kind = OdinDocTypeUnion
		if t.Union.IsPolymorphic {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Union_polymorphic)
		}
		switch t.Union.Kind {
		case UnionType_no_nil:
			doc_type.Flags |= uint32(OdinDocTypeFlag_Union_no_nil)
		case UnionType_shared_nil:
			doc_type.Flags |= uint32(OdinDocTypeFlag_Union_shared_nil)
		}
		{
			variants := make([]OdinDocTypeIndex, len(t.Union.Variants))
			for i, v := range t.Union.Variants {
				variants[i] = odin_doc_type(w, v)
			}
			doc_type.Types = odin_write_slice(w, variants)
			doc_type.PolymorphicParams = odin_doc_type(w, t.Union.PolymorphicParams)
		}
		if t.Union.Node != nil && t.Union.Node.Kind == Ast_UnionType {
			ut := &t.Union.Node.UnionType
			if ut.Align != nil {
				doc_type.CustomAlign = odin_doc_expr_string(w, ut.Align)
			}
			doc_type.WhereClauses = odin_doc_where_clauses(w, ut.WhereClauses[:])
		}
	case Type_Enum:
		doc_type.Kind = OdinDocTypeEnum
		{
			fields := make([]OdinDocEntityIndex, len(t.Enum.Fields))
			for i, field := range t.Enum.Fields {
				fields[i] = odin_doc_add_entity(w, field)
			}
			doc_type.Entities = odin_write_slice(w, fields)
			if t.Enum.BaseType != nil {
				doc_type.Types = odin_doc_type_as_slice(w, t.Enum.BaseType)
			}
		}
	case Type_Tuple:
		doc_type.Kind = OdinDocTypeTuple
		{
			variables := make([]OdinDocEntityIndex, len(t.Tuple.Variables))
			for i, v := range t.Tuple.Variables {
				variables[i] = odin_doc_add_entity(w, v)
			}
			doc_type.Entities = odin_write_slice(w, variables)
		}
	case Type_Proc:
		doc_type.Kind = OdinDocTypeProc
		if t.Proc.IsPolymorphic {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Proc_polymorphic)
		}
		if t.Proc.Diverging {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Proc_diverging)
		}
		if t.Proc.OptionalOk {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Proc_optional_ok)
		}
		if t.Proc.Variadic {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Proc_variadic)
		}
		if t.Proc.CVararg {
			doc_type.Flags |= uint32(OdinDocTypeFlag_Proc_c_vararg)
		}
		{
			types := [2]OdinDocTypeIndex{
				odin_doc_type(w, t.Proc.Params),
				odin_doc_type(w, t.Proc.Results),
			}
			doc_type.Types = odin_write_slice(w, types[:])
			calling_convention := make_string_c(proc_calling_convention_strings[t.Proc.CallingConvention])
			doc_type.CallingConvention = odin_doc_write_string(w, calling_convention)
		}
	case Type_BitSet:
		doc_type.Kind = OdinDocTypeBitSet
		{
			type_count := 0
			types := [2]OdinDocTypeIndex{}
			if t.BitSet.Elem != nil {
				types[type_count] = odin_doc_type(w, t.BitSet.Elem)
				type_count++
			}
			if t.BitSet.Underlying != nil {
				types[type_count] = odin_doc_type(w, t.BitSet.Underlying)
				type_count++
				doc_type.Flags |= uint32(OdinDocTypeFlag_BitSet_UnderlyingType)
			}
			doc_type.Types = odin_write_slice(w, types[:type_count])
			doc_type.ElemCountLen = 2
			doc_type.ElemCounts[0] = t.BitSet.Lower
			doc_type.ElemCounts[1] = t.BitSet.Upper
		}
	case Type_SimdVector:
		doc_type.Kind = OdinDocTypeSimdVector
		doc_type.ElemCountLen = 1
		doc_type.ElemCounts[0] = t.SimdVector.Count
		doc_type.Types = odin_doc_type_as_slice(w, t.SimdVector.Elem)
	case Type_Matrix:
		doc_type.Kind = OdinDocTypeMatrix
		doc_type.ElemCountLen = 2
		doc_type.ElemCounts[0] = t.Matrix.RowCount
		doc_type.ElemCounts[1] = t.Matrix.ColumnCount
		doc_type.Types = odin_doc_type_as_slice(w, t.Matrix.Elem)
	}
	if dst != nil {
		*dst = doc_type
	}
	return type_index
}

func odin_doc_add_entity(w *OdinDocWriter, e *Entity) OdinDocEntityIndex {
	if e == nil {
		return 0
	}
	if prev, ok := w.EntityCache[e]; ok {
		return prev
	}
	if e.Pkg != nil {
		if _, ok := w.PkgCache[e.Pkg]; !ok {
			return 0
		}
	}
	doc_entity := OdinDocEntity{}
	var dst *OdinDocEntity
	doc_entity_index := odin_doc_write_item(w, &w.Entities, &doc_entity, &dst)
	w.EntityCache[e] = doc_entity_index
	var type_expr *Ast
	var init_expr *Ast
	var decl_node *Ast
	var comment *CommentGroup
	var docs *CommentGroup
	if e.DeclInfo != nil {
		type_expr = e.DeclInfo.TypeExpr
		init_expr = e.DeclInfo.InitExpr
		decl_node = e.DeclInfo.DeclNode
		comment = e.DeclInfo.Comment
		docs = e.DeclInfo.Docs
	}
	if e.Kind == Entity_Variable {
		if comment == nil {
			comment = e.Variable.Comment
		}
		if docs == nil {
			docs = e.Variable.Docs
		}
	} else if e.Kind == Entity_Constant {
		if comment == nil {
			comment = e.Constant.Comment
		}
		if docs == nil {
			docs = e.Constant.Docs
		}
	}
	name := e.Token.String
	link_name := String{}
	pos := e.Token.Pos
	kind := OdinDocEntityInvalid
	flags := uint64(0)
	field_group_index := int32(-1)
	switch e.Kind {
	case Entity_Invalid:
		kind = OdinDocEntityInvalid
	case Entity_Constant:
		kind = OdinDocEntityConstant
	case Entity_Variable:
		kind = OdinDocEntityVariable
	case Entity_TypeName:
		kind = OdinDocEntityTypeName
	case Entity_Procedure:
		kind = OdinDocEntityProcedure
	case Entity_ProcGroup:
		kind = OdinDocEntityProcGroup
	case Entity_ImportName:
		kind = OdinDocEntityImportName
	case Entity_LibraryName:
		kind = OdinDocEntityLibraryName
	case Entity_Builtin:
		kind = OdinDocEntityBuiltin
	}
	switch e.Kind {
	case Entity_TypeName:
		if e.TypeName.IsTypeAlias {
			flags |= OdinDocEntityFlagTypeAlias
		}
	case Entity_Variable:
		if e.Flags&EntityFlag_BitFieldField != 0 {
			flags |= OdinDocEntityFlagBitFieldField
		}
		if e.Variable.IsForeign {
			flags |= OdinDocEntityFlagForeign
		}
		if e.Variable.IsExport {
			flags |= OdinDocEntityFlagExport
		}
		if goStr(e.Variable.ThreadLocalModel) != "" {
			flags |= OdinDocEntityFlagVarThreadLocal
		}
		if e.Flags&EntityFlag_Static != 0 {
			flags |= OdinDocEntityFlagVarStatic
		}
		link_name = e.Variable.LinkName
		if init_expr == nil {
			init_expr = e.Variable.InitExpr
		}
		if e.Flags&EntityFlag_BitFieldField != 0 {
			field_group_index = -int32(e.Variable.BitFieldBitSize)
		} else {
			field_group_index = e.Variable.FieldGroupIndex
		}
	case Entity_Constant:
		field_group_index = e.Constant.FieldGroupIndex
	case Entity_Procedure:
		if e.Procedure.IsForeign {
			flags |= OdinDocEntityFlagForeign
		}
		if e.Procedure.IsExport {
			flags |= OdinDocEntityFlagExport
		}
		link_name = e.Procedure.LinkName
	case Entity_Builtin:
		bp := builtin_procs[e.Builtin.Id]
		pos = TokenPos{}
		name = bp.Name
		switch bp.Pkg {
		case BuiltinProcPkg_builtin:
			flags |= OdinDocEntityFlagBuiltinPkgBuiltin
		case BuiltinProcPkg_intrinsics:
			flags |= OdinDocEntityFlagBuiltinPkgIntrinsics
		default:
			panic("Unhandled BuiltinProcPkg")
		}
	}
	if e.Flags&EntityFlag_Using != 0 {
		flags |= OdinDocEntityFlagParamUsing
	}
	if e.Flags&EntityFlag_ConstInput != 0 {
		flags |= OdinDocEntityFlagParamConst
	}
	if e.Flags&EntityFlag_Ellipsis != 0 {
		flags |= OdinDocEntityFlagParamEllipsis
	}
	if e.Flags&EntityFlag_NoAlias != 0 {
		flags |= OdinDocEntityFlagParamNoAlias
	}
	if e.Flags&EntityFlag_AnyInt != 0 {
		flags |= OdinDocEntityFlagParamAnyInt
	}
	if e.Flags&EntityFlag_ByPtr != 0 {
		flags |= OdinDocEntityFlagParamByPtr
	}
	if e.Flags&EntityFlag_NoBroadcast != 0 {
		flags |= OdinDocEntityFlagParamNoBroadcast
	}
	if e.Scope != nil && (e.Scope.Flags&(ScopeFlag_File|ScopeFlag_Pkg)) != 0 && !is_entity_exported(e) {
		flags |= OdinDocEntityFlagPrivate
	}
	init_string := OdinDocString{}
	if init_expr != nil {
		init_string = odin_doc_expr_string(w, init_expr)
	} else {
		if e.Kind == Entity_Constant {
			if e.Constant.Flags&EntityConstantFlag_ImplicitEnumValue != 0 {
				init_string = OdinDocString{}
			} else if e.Constant.ParamValue.OriginalAstExpr != nil {
				init_string = odin_doc_expr_string(w, e.Constant.ParamValue.OriginalAstExpr)
			} else {
				init_string = odin_doc_write_string(w, exact_value_to_string(e.Constant.Value))
			}
		} else if e.Kind == Entity_Variable {
			if e.Variable.ParamValue.OriginalAstExpr != nil {
				init_string = odin_doc_expr_string(w, e.Variable.ParamValue.OriginalAstExpr)
			}
		}
	}
	doc_entity.Kind = kind
	doc_entity.Flags = flags
	doc_entity.Pos = odin_doc_token_pos_cast(w, pos)
	doc_entity.Name = odin_doc_write_string(w, name)
	doc_entity.Type = 0
	doc_entity.InitString = init_string
	doc_entity.Comment = odin_doc_comment_group_string(w, comment)
	doc_entity.Docs = odin_doc_comment_group_string(w, docs)
	doc_entity.FieldGroupIndex = field_group_index
	doc_entity.ForeignLibrary = 0
	doc_entity.LinkName = odin_doc_write_string(w, link_name)
	if e.DeclInfo != nil {
		doc_entity.Attributes = odin_doc_attributes(w, e.DeclInfo.Attributes[:])
	}
	doc_entity.GroupedEntities = OdinDocArray[OdinDocEntityIndex]{}
	doc_entity.WhereClauses = OdinDocArray[OdinDocString]{}
	if dst != nil {
		*dst = doc_entity
	}
	return doc_entity_index
}

func odin_doc_update_entities(w *OdinDocWriter) {
	{
		entities := make([]*Entity, 0, len(w.EntityCache))
		for e := range w.EntityCache {
			entities = append(entities, e)
		}
		for _, e := range entities {
			_ = odin_doc_type(w, e.Type)
		}
	}
	for e, entity_index := range w.EntityCache {
		type_index := odin_doc_type(w, e.Type)
		foreign_library := OdinDocEntityIndex(0)
		grouped_entities := OdinDocArray[OdinDocEntityIndex]{}
		switch e.Kind {
		case Entity_Variable:
			if w.State == OdinDocWriterStateWriting {
				if type_index == 0 {
					panic("type_index != 0")
				}
			}
			foreign_library = odin_doc_add_entity(w, e.Variable.ForeignLibrary)
		case Entity_Procedure:
			foreign_library = odin_doc_add_entity(w, e.Procedure.ForeignLibrary)
		case Entity_ProcGroup:
			pges := make([]OdinDocEntityIndex, 0, len(e.ProcGroup.Entities))
			for _, entity := range e.ProcGroup.Entities {
				pges = append(pges, odin_doc_add_entity(w, entity))
			}
			grouped_entities = odin_write_slice(w, pges)
		}
		dst := odin_doc_get_item(w, &w.Entities, entity_index)
		if dst != nil {
			if dst.Kind == OdinDocEntityVariable {
				if type_index == 0 {
					panic("type_index != 0")
				}
			}
			dst.Type = type_index
			dst.ForeignLibrary = foreign_library
			dst.GroupedEntities = grouped_entities
		}
	}
}

func odin_doc_add_pkg_entries(w *OdinDocWriter, pkg *AstPackage) OdinDocArray[OdinDocScopeEntry] {
	if pkg.Scope == nil {
		return OdinDocArray[OdinDocScopeEntry]{}
	}
	if _, ok := w.PkgCache[pkg]; !ok {
		return OdinDocArray[OdinDocScopeEntry]{}
	}
	entries := make([]OdinDocScopeEntry, 0, len(w.EntityCache))
	for interned, e := range pkg.Scope.Elements {
		switch e.Kind {
		case Entity_Invalid, Entity_Nil, Entity_Label:
			continue
		case Entity_Constant, Entity_Variable, Entity_TypeName, Entity_Procedure, Entity_ProcGroup, Entity_ImportName, Entity_LibraryName, Entity_Builtin:
		}
		if e.Pkg != pkg {
			continue
		}
		if !is_entity_exported(e, true) {
			continue
		}
		if e.Token.String.Len == 0 {
			continue
		}
		entry := OdinDocScopeEntry{
			Name:   odin_doc_write_string(w, interned),
			Entity: odin_doc_add_entity(w, e),
		}
		entries = append(entries, entry)
	}
	return odin_write_slice(w, entries)
}

func odin_doc_write_docs(w *OdinDocWriter) {
	pkgs := make([]*AstPackage, 0, len(w.Info.Packages))
	for _, pkg := range w.Info.Packages {
		if build_context.cmd_doc_flags&CmdDocFlag_AllPackages != 0 {
			pkgs = append(pkgs, pkg)
		} else {
			if pkg.Kind == Package_Init || pkg.IsExtra {
				pkgs = append(pkgs, pkg)
			}
		}
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return cmp_ast_package_by_name(pkgs[i], pkgs[j]) < 0
	})
	for _, pkg := range pkgs {
		pkg_flags := uint32(0)
		switch pkg.Kind {
		case Package_Normal:
		case Package_Runtime:
			pkg_flags |= OdinDocPkgFlagRuntime
		case Package_Init:
			pkg_flags |= OdinDocPkgFlagInit
		case Package_Builtin:
			pkg_flags |= OdinDocPkgFlagBuiltin
		}
		doc_pkg := OdinDocPkg{
			Fullpath: odin_doc_write_string(w, pkg.Fullpath),
			Name:     odin_doc_write_string(w, pkg.Name),
			Flags:    pkg_flags,
			Docs:     odin_doc_pkg_doc_string(w, pkg),
		}
		var dst *OdinDocPkg
		pkg_index := odin_doc_write_item(w, &w.Pkgs, &doc_pkg, &dst)
		w.PkgCache[pkg] = pkg_index
		file_indices := make([]OdinDocFileIndex, 0, len(pkg.Files))
		for _, file := range pkg.Files {
			doc_file := OdinDocFile{
				Pkg:  pkg_index,
				Name: odin_doc_write_string(w, file.Fullpath),
			}
			file_index := odin_doc_write_item(w, &w.Files, &doc_file)
			w.FileCache[file] = file_index
			file_indices = append(file_indices, file_index)
		}
		doc_pkg.Files = odin_write_slice(w, file_indices)
		doc_pkg.Entries = odin_doc_add_pkg_entries(w, pkg)
		if dst != nil {
			*dst = doc_pkg
		}
	}
	odin_doc_update_entities(w)
}

func odin_doc_write_to_file(w *OdinDocWriter, filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write .odin-doc to: %s\n", filename)
		exit_with_errors()
		return
	}
	defer f.Close()
	n, err := f.Write(w.Data[:w.DataLen])
	if err != nil || n != int(w.DataLen) {
		fmt.Fprintf(os.Stderr, "Failed to write .odin-doc to: %s\n", filename)
		exit_with_errors()
		return
	}
	fmt.Printf("Wrote .odin-doc file to: %s\n", filename)
}

func odin_doc_write(info *CheckerInfo, filename string) {
	gInDocWriter = true
	w := &OdinDocWriter{}
	w.Info = info
	odin_doc_writer_prepare(w)
	odin_doc_write_docs(w)
	odin_doc_writer_start_writing(w)
	odin_doc_write_docs(w)
	odin_doc_writer_end_writing(w)
	odin_doc_write_to_file(w, filename)
	odin_doc_writer_destroy(w)
	gInDocWriter = false
}

func is_in_doc_writer() bool {
	return gInDocWriter
}
