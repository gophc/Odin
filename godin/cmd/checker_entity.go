package cmd

import "sync/atomic"

var globalEntityId atomic.Uint64

func has_parameter_value(param_value ParameterValue) bool {
	if param_value.Kind != ParameterValue_Invalid {
		return true
	}
	if param_value.OriginalAstExpr != nil {
		return true
	}
	return false
}

func create_type_name_obj_c_metadata() *TypeNameObjCMetadata {
	md := permanentAllocItem[TypeNameObjCMetadata]()
	md.Mutex = permanentAllocItem[BlockingMutex]()
	md.TypeEntries = make([]TypeNameObjCMetadataEntry, 0)
	md.ValueEntries = make([]TypeNameObjCMetadataEntry, 0)
	return md
}

func entity_interned_name(entity *Entity) InternedString {
	name := entity.InternedName
	if name.Value == 0 {
		name = stringInternerInsert(entity.Token.String)
		entity.InternedName = name
		entity.InternedNameHash = name.Hash()
	}
	return name
}

func is_entity_kind_exported(kind EntityKind, allow_builtin ...bool) bool {
	allow := len(allow_builtin) > 0 && allow_builtin[0]
	switch kind {
	case Entity_Builtin:
		return allow
	case Entity_ImportName, Entity_LibraryName, Entity_Nil:
		return false
	}
	return true
}

func is_entity_exported(e *Entity, allow_builtin ...bool) bool {
	if e == nil {
		gbAssertHandler("Assertion Failure", "e != nil", "checker_entity.go", 0)
	}
	allow := len(allow_builtin) > 0 && allow_builtin[0]
	if !is_entity_kind_exported(e.Kind, allow) {
		return false
	}
	if e.Flags&EntityFlag_NotExported != 0 {
		return false
	}
	if e.File != nil && (e.File.Flags&(AstFile_IsPrivatePkg|AstFile_IsPrivateFile)) != 0 {
		return false
	}
	name := e.Token.String
	switch len(name) {
	case 0:
		return false
	case 1:
		return name[0] != '_'
	}
	return true
}

func entity_has_deferred_procedure(e *Entity) bool {
	if e == nil {
		gbAssertHandler("Assertion Failure", "e != nil", "checker_entity.go", 0)
	}
	if e.Kind == Entity_Procedure {
		return e.Procedure.DeferredProcedure.Entity != nil
	}
	return false
}

func alloc_entity(kind EntityKind, scope *Scope, token Token, type_ *Type) *Entity {
	entity := permanentAllocItem[Entity]()
	entity.Kind = kind
	entity.State = EntityState_Unresolved
	entity.Scope = scope
	entity.Token = token
	entity.Type = type_
	entity.Id = 1 + globalEntityId.Add(1)
	if token.Pos.FileID != 0 {
		entity.File = threadUnsafeGetAstFileFromId(token.Pos.FileID)
	}
	entity_interned_name(entity)
	return entity
}

func alloc_entity_variable(scope *Scope, token Token, type_ *Type, state ...EntityState) *Entity {
	entity := alloc_entity(Entity_Variable, scope, token, type_)
	if len(state) > 0 {
		entity.State = state[0]
	} else {
		entity.State = EntityState_Unresolved
	}
	return entity
}

func alloc_entity_using_variable(parent *Entity, token Token, type_ *Type, using_expr *Ast) *Entity {
	if parent == nil {
		gbAssertHandler("Assertion Failure", "parent != nil", "checker_entity.go", 0)
	}
	token.Pos = parent.Token.Pos
	entity := alloc_entity(Entity_Variable, parent.Scope, token, type_)
	entity.UsingParent = parent
	entity.ParentProcDecl = parent.ParentProcDecl
	entity.UsingExpr = using_expr
	entity.Flags |= EntityFlag_Using
	entity.Flags |= EntityFlag_Used
	entity.State = EntityState_Resolved
	return entity
}

func alloc_entity_constant(scope *Scope, token Token, type_ *Type, value ExactValue) *Entity {
	entity := alloc_entity(Entity_Constant, scope, token, type_)
	entity.Constant.Value = value
	return entity
}

func alloc_entity_type_name(scope *Scope, token Token, type_ *Type, state ...EntityState) *Entity {
	entity := alloc_entity(Entity_TypeName, scope, token, type_)
	if len(state) > 0 {
		entity.State = state[0]
	} else {
		entity.State = EntityState_Unresolved
	}
	return entity
}

func alloc_entity_param(scope *Scope, token Token, type_ *Type, is_using, is_value bool) *Entity {
	entity := alloc_entity_variable(scope, token, type_, EntityState_Resolved)
	entity.Flags |= EntityFlag_Used
	entity.Flags |= EntityFlag_Param
	if is_using {
		entity.Flags |= EntityFlag_Using
	}
	if is_value {
		entity.Flags |= EntityFlag_Value
	}
	return entity
}

func alloc_entity_const_param(scope *Scope, token Token, type_ *Type, value ExactValue, poly_const bool) *Entity {
	entity := alloc_entity_constant(scope, token, type_, value)
	entity.Flags |= EntityFlag_Used
	if poly_const {
		entity.Flags |= EntityFlag_PolyConst
	}
	entity.Flags |= EntityFlag_Param
	return entity
}

func alloc_entity_field(scope *Scope, token Token, type_ *Type, is_using bool, field_index int32, state ...EntityState) *Entity {
	entity := alloc_entity_variable(scope, token, type_)
	entity.Variable.FieldIndex = field_index
	if is_using {
		entity.Flags |= EntityFlag_Using
	}
	entity.Flags |= EntityFlag_Field
	if len(state) > 0 {
		entity.State = state[0]
	} else {
		entity.State = EntityState_Unresolved
	}
	return entity
}

func alloc_entity_array_elem(scope *Scope, token Token, type_ *Type, field_index int32) *Entity {
	entity := alloc_entity_variable(scope, token, type_)
	entity.Variable.FieldIndex = field_index
	entity.Flags |= EntityFlag_Field
	entity.Flags |= EntityFlag_ArrayElem
	entity.State = EntityState_Resolved
	return entity
}

func alloc_entity_procedure(scope *Scope, token Token, signature_type *Type, tags ...uint64) *Entity {
	entity := alloc_entity(Entity_Procedure, scope, token, signature_type)
	if len(tags) > 0 {
		entity.Procedure.Tags = tags[0]
	} else {
		entity.Procedure.Tags = 0
	}
	return entity
}

func alloc_entity_proc_group(scope *Scope, token Token, type_ *Type) *Entity {
	entity := alloc_entity(Entity_ProcGroup, scope, token, type_)
	return entity
}

func alloc_entity_import_name(scope *Scope, token Token, type_ *Type, path string, name string, import_scope *Scope) *Entity {
	entity := alloc_entity(Entity_ImportName, scope, token, type_)
	entity.ImportName.Path = path
	entity.ImportName.Name = name
	entity.ImportName.Scope = import_scope
	entity.State = EntityState_Resolved
	return entity
}

func alloc_entity_library_name(scope *Scope, token Token, type_ *Type, paths []string, name string) *Entity {
	entity := alloc_entity(Entity_LibraryName, scope, token, type_)
	entity.LibraryName.Paths = paths
	entity.LibraryName.Name = name
	entity.State = EntityState_Resolved
	return entity
}

func alloc_entity_nil(name string, type_ *Type) *Entity {
	entity := alloc_entity(Entity_Nil, nil, makeTokenIdent(name), type_)
	return entity
}

func alloc_entity_label(scope *Scope, token Token, type_ *Type, node *Ast, parent *Ast) *Entity {
	entity := alloc_entity(Entity_Label, scope, token, type_)
	entity.Label.Node = node
	entity.Label.Parent = parent
	entity.State = EntityState_Resolved
	return entity
}

func alloc_entity_dummy_variable(scope *Scope, token Token) *Entity {
	token.String = "_"
	return alloc_entity_variable(scope, token, nil)
}

func entity_from_expr(expr *Ast) *Entity

func strip_entity_wrapping(e *Entity) *Entity {
	if e == nil {
		return nil
	}
	if e.Kind != Entity_Constant {
		return e
	}
	if e.Constant.Value.Kind == ExactValue_Procedure {
		return strip_entity_wrapping(e.Constant.Value.ValueProcedure)
	}
	return e
}

func strip_entity_wrapping_ast(expr *Ast) *Entity {
	e := entity_from_expr(expr)
	return strip_entity_wrapping(e)
}

func is_entity_local_variable(e *Entity) bool {
	if e == nil {
		return false
	}
	if e.Kind != Entity_Variable {
		return false
	}
	if e.Variable.IsGlobal {
		return false
	}
	if e.Scope == nil {
		return true
	}
	if e.Flags&(EntityFlag_ForValue|EntityFlag_SwitchValue|EntityFlag_Static) != 0 {
		return false
	}
	return (e.Scope.Flags&^ScopeFlag_ContextDefined) == 0 ||
		(e.Scope.Flags&ScopeFlag_Proc) != 0
}

func entity_variable_pos_cmp(a, b *Entity) int {
	return tokenPosCmp(a.Token.Pos, b.Token.Pos)
}

func is_operand_value(o Operand) bool {
	switch o.Mode {
	case Addressing_Value,
		Addressing_Context,
		Addressing_Variable,
		Addressing_Constant,
		Addressing_MapIndex,
		Addressing_OptionalOk,
		Addressing_OptionalOkPtr,
		Addressing_SoaVariable,
		Addressing_SwizzleValue,
		Addressing_SwizzleVariable:
		return true
	}
	return false
}

func is_operand_nil(o Operand) bool {
	return o.Mode == Addressing_Value && o.Type == t_untyped_nil
}

func is_operand_uninit(o Operand) bool {
	return o.Mode == Addressing_Value && o.Type == t_untyped_uninit
}

func check_rtti_type_disallowed(token Token, type_ *Type, format string, args ...any) bool {
	if buildContext.NoRtti && type_ != nil {
		if is_type_any(type_) {
			t := type_to_string(type_)
			error_(token, format, append([]any{t}, args...)...)
			gbStringFree(t)
			return true
		}
	}
	return false
}

func check_rtti_type_disallowed_expr(expr *Ast, type_ *Type, format string, args ...any) bool {
	if expr == nil {
		gbAssertHandler("Assertion Failure", "expr != nil", "checker_entity.go", 0)
	}
	return check_rtti_type_disallowed(astToken(expr), type_, format, args...)
}
