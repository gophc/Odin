package cmd

import (
	"sort"
	"sync/atomic"
)

type VettedEntityKind uint32

const (
	VettedEntity_Invalid           VettedEntityKind = iota
	VettedEntity_Unused
	VettedEntity_Shadowed
	VettedEntity_Shadowed_And_Unused
)

type VettedEntity struct {
	Kind   VettedEntityKind
	Entity *Entity
	Other  *Entity
}

func check_vet_flags(c *CheckerContext) uint64 {
	file := c.File
	if file == nil && c.CurrProcDecl != nil && c.CurrProcDecl.ProcLit != nil {
		file = thread_safe_get_ast_file_from_id(c.CurrProcDecl.ProcLit.FileID)
	}
	return ast_file_vet_flags(file)
}

func check_vet_flags_from_node(node *Ast) uint64 {
	file := thread_safe_get_ast_file_from_id(node.FileID)
	return ast_file_vet_flags(file)
}

func check_feature_flags(c *CheckerContext, node *Ast) uint64 {
	file := c.File
	if file == nil && c.CurrProcDecl != nil && c.CurrProcDecl.ProcLit != nil {
		file = thread_safe_get_ast_file_from_id(c.CurrProcDecl.ProcLit.FileID)
	}
	if file == nil {
		file = thread_safe_get_ast_file_from_id(node.FileID)
	}
	if file != nil && file.FeatureFlagsSet {
		return file.FeatureFlags
	}
	return 0
}

func check_feature_flags_from_entity(e *Entity) uint64 {
	if e == nil {
		return 0
	}
	file := e.File
	if file == nil && e.DeclInfo != nil && e.DeclInfo.DeclNode != nil {
		file = thread_safe_get_ast_file_from_id(e.DeclInfo.DeclNode.FileID)
	}
	if file != nil && file.FeatureFlagsSet {
		return file.FeatureFlags
	}
	return 0
}

func vetted_entity_variable_pos_cmp(a, b *VettedEntity) int {
	x := a.Entity
	y := b.Entity
	if x == nil {
		gb_assert_handler("Assertion Failure", "x != nil", "checker_vet.go", 0, "")
	}
	if y == nil {
		gb_assert_handler("Assertion Failure", "y != nil", "checker_vet.go", 0, "")
	}
	return token_pos_cmp(x.Token.Pos, y.Token.Pos)
}

func check_vet_shadowing_assignment(c *Checker, shadowed *Entity, expr *Ast) bool {
	init := unparen_expr(expr)
	if init == nil {
		return false
	}
	if init.Kind == Ast_Ident {
		if init.Ident.Entity == shadowed {
			return true
		}
	} else if init.Kind == Ast_TernaryIfExpr {
		x := check_vet_shadowing_assignment(c, shadowed, init.TernaryIfExpr.X)
		y := check_vet_shadowing_assignment(c, shadowed, init.TernaryIfExpr.Y)
		if x || y {
			return true
		}
	}
	return false
}

func check_vet_shadowing(c *Checker, e *Entity, ve *VettedEntity) bool {
	if e.Kind != Entity_Variable {
		return false
	}
	if goStr(e.Token.String) == "_" {
		return false
	}
	if e.Flags&EntityFlag_Param != 0 {
		return false
	}
	if e.Scope.Flags&(ScopeFlag_Global|ScopeFlag_File|ScopeFlag_Proc) != 0 {
		return false
	}
	parent := e.Scope.Parent
	if parent.Flags&(ScopeFlag_Global|ScopeFlag_File) != 0 {
		return false
	}
	interned := entity_interned_name(e)
	hash := atomic.LoadUint32(&e.InternedNameHash)
	shadowed := scope_lookup(parent, interned, hash)
	if shadowed == nil {
		return false
	}
	if shadowed.Kind != Entity_Variable {
		return false
	}
	if e.Token.Pos.FileID != shadowed.Token.Pos.FileID {
		return false
	}
	if token_pos_cmp(shadowed.Token.Pos, e.Token.Pos) > 0 {
		return false
	}
	if !are_types_identical(e.Type, shadowed.Type) {
		return false
	}
	if e.Flags&EntityFlag_Using == 0 && e.Kind == Entity_Variable {
		if check_vet_shadowing_assignment(c, shadowed, e.Variable.InitExpr) {
			return false
		}
	}
	ve.Kind = VettedEntity_Shadowed
	ve.Entity = e
	ve.Other = shadowed
	return true
}

func check_vet_unused(c *Checker, e *Entity, ve *VettedEntity) bool {
	if e.Flags&EntityFlag_Used != 0 {
		return false
	}
	switch e.Kind {
	case Entity_Variable:
		if e.Scope.Flags&(ScopeFlag_Global|ScopeFlag_Type|ScopeFlag_File) != 0 {
			return false
		}
		if e.Flags&EntityFlag_Static != 0 {
			return false
		}
		fallthrough
	case Entity_ImportName, Entity_LibraryName:
		ve.Kind = VettedEntity_Unused
		ve.Entity = e
		return true
	}
	return false
}

func check_scope_usage_internal(c *Checker, scope *Scope, vet_flags uint64, per_entity bool) {
	original_vet_flags := vet_flags

	vetted_entities := make([]VettedEntity, 0)

	rw_mutex_shared_lock(&scope.Mutex)
	for i := uint32(0); i < scope.Elements.Cap; i++ {
		slot := &scope.Elements.Slots[i]
		if slot.Hash == 0 {
			continue
		}
		e := slot.Value
		if e == nil {
			continue
		}

		vet_flags = original_vet_flags
		if per_entity {
			vet_flags = ast_file_vet_flags(e.File)
		}
		vet_unused := (vet_flags & uint64(VetFlag_Unused)) != 0
		vet_shadowing := (vet_flags & uint64(VetFlag_Shadowing|VetFlag_Using)) != 0
		vet_unused_procedures := (vet_flags & uint64(VetFlag_UnusedProcedures)) != 0

		if vet_unused_procedures && e.Pkg != nil && e.Pkg.Kind == Package_Runtime {
			vet_unused_procedures = false
		}

		ve_unused := VettedEntity{}
		ve_shadowed := VettedEntity{}
		is_unused := false

		if vet_unused && check_vet_unused(c, e, &ve_unused) {
			is_unused = true
		} else if vet_unused_procedures && e.Kind == Entity_Procedure {
			if e.Flags&EntityFlag_Used != 0 {
				is_unused = false
			} else if e.Flags&EntityFlag_Require != 0 {
				is_unused = false
			} else if e.Flags&EntityFlag_Init != 0 {
				is_unused = false
			} else if e.Flags&EntityFlag_Fini != 0 {
				is_unused = false
			} else if e.Procedure.IsExport {
				is_unused = false
			} else if e.Pkg != nil && e.Pkg.Kind == Package_Init && string_equals(e.Token.String, S("main")) {
				is_unused = false
			} else {
				is_unused = true
				ve_unused.Kind = VettedEntity_Unused
				ve_unused.Entity = e
			}
		}

		is_shadowed := vet_shadowing && check_vet_shadowing(c, e, &ve_shadowed)

		if is_unused && is_shadowed {
			ve_both := ve_shadowed
			ve_both.Kind = VettedEntity_Shadowed_And_Unused
			vetted_entities = append(vetted_entities, ve_both)
		} else if is_unused {
			vetted_entities = append(vetted_entities, ve_unused)
		} else if is_shadowed {
			vetted_entities = append(vetted_entities, ve_shadowed)
		} else if e.Kind == Entity_Variable && (e.Flags&(EntityFlag_Param|EntityFlag_Using|EntityFlag_Static|EntityFlag_Field)) == 0 && !e.Variable.IsGlobal {
			sz := type_size_of(e.Type)
			if sz > 1<<18 {
				is_ref := false
				if e.Flags&EntityFlag_ForValue != 0 {
					is_ref = type_deref(e.Variable.ForLoopParentType, false) != nil
				} else if e.Flags&EntityFlag_SwitchValue != 0 {
					is_ref = (e.Flags & EntityFlag_Value) == 0
				}
				if !is_ref {
					type_str := type_to_string(e.Type)
					warning(e.Token, "Declaration of '%s' may cause a stack overflow due to its type '%s' having a size of %lld bytes",
						goStr(e.Token.String), type_str, int64(sz))
					gb_string_free(type_str)
				}
			}
		}
	}
	rw_mutex_shared_unlock(&scope.Mutex)

	sort.Slice(vetted_entities, func(i, j int) bool {
		return vetted_entity_variable_pos_cmp(&vetted_entities[i], &vetted_entities[j]) < 0
	})

	for _, ve := range vetted_entities {
		e := ve.Entity
		other := ve.Other

		vet_flags = original_vet_flags
		if per_entity {
			vet_flags = ast_file_vet_flags(e.File)
		}

		if ve.Kind == VettedEntity_Shadowed_And_Unused {
			error_(e.Token, "'%s' declared but not used, possibly shadows declaration at line %d",
				goStr(e.Token.String), other.Token.Pos.Line)
		} else if vet_flags != 0 {
			switch ve.Kind {
			case VettedEntity_Unused:
				if e.Kind == Entity_Variable && (vet_flags&uint64(VetFlag_UnusedVariables)) != 0 {
					error_(e.Token, "'%s' declared but not used", goStr(e.Token.String))
				}
				if e.Kind == Entity_Procedure && (vet_flags&uint64(VetFlag_UnusedProcedures)) != 0 {
					error_(e.Token, "'%s' declared but not used", goStr(e.Token.String))
				}
				if (e.Kind == Entity_ImportName || e.Kind == Entity_LibraryName) && (vet_flags&uint64(VetFlag_UnusedImports)) != 0 {
					error_(e.Token, "'%s' declared but not used", goStr(e.Token.String))
				}

			case VettedEntity_Shadowed:
				if (vet_flags&uint64(VetFlag_Shadowing|VetFlag_Using)) != 0 && e.Flags&EntityFlag_Using != 0 {
					error_(e.Token, "Declaration of '%s' from 'using' shadows declaration at line %d",
						goStr(e.Token.String), other.Token.Pos.Line)
				} else if (vet_flags & uint64(VetFlag_Shadowing)) != 0 {
					error_(e.Token, "Declaration of '%s' shadows declaration at line %d",
						goStr(e.Token.String), other.Token.Pos.Line)
				}
			}
		}
	}
}

func check_scope_usage(c *Checker, scope *Scope, vet_flags uint64) {
	check_scope_usage_internal(c, scope, vet_flags, false)
	for child := scope.HeadChild; child != nil; child = child.Next {
		if child.Flags&(ScopeFlag_Proc|ScopeFlag_Type|ScopeFlag_File) != 0 {
		} else {
			check_scope_usage(c, child, vet_flags)
		}
	}
}
