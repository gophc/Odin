package cmd

import "sync/atomic"

func check_const_decl(ctx *CheckerContext, e *Entity, type_expr *Ast, init *Ast, named_type *Type) {
	gb_assert_handler(e.Type == nil, "e.Type == nil", "cmd_check_decl_entity.go", 0)
	gb_assert_handler(e.Kind == Entity_Constant, "e.Kind == Entity_Constant", "cmd_check_decl_entity.go", 0)
	init = unparen_expr(init)
	if e.Flags&EntityFlag_Visited != 0 {
		e.Type = t_invalid
		return
	}
	e.Flags |= EntityFlag_Visited
	if type_expr != nil {
		e.Type = check_type(ctx, type_expr)
		if are_types_identical(e.Type, t_typeid) {
			e.Type = nil
			e.Kind = Entity_TypeName
			check_type_decl(ctx, e, init, named_type)
			return
		}
	}
	operand := Operand{}
	if init != nil {
		entity := check_entity_from_ident_or_selector(ctx, init, false)
		if check_override_as_type_due_to_aliasing(ctx, e, entity, init, named_type) {
			return
		}
		entity = nil
		if init.Kind == Ast_Ident {
			entity = check_ident(ctx, &operand, init, nil, e.Type, true)
		} else if init.Kind == Ast_SelectorExpr {
			entity = check_selector(ctx, &operand, init, e.Type)
		} else {
			check_expr_or_type(ctx, &operand, init, e.Type)
			if init.Kind == Ast_CallExpr {
				entity = init.CallExpr.EntityProcedureOf.Load()
			}
		}
		switch operand.Mode {
		case Addressing_Type:
			if e.Type != nil && !is_type_typeid(e.Type) {
				check_assignment(ctx, &operand, e.Type, S("constant declaration"))
			}
			e.Kind = Entity_TypeName
			e.Type = nil
			if entity != nil && entity.Type != nil && is_type_polymorphic_record_unspecialized(entity.Type) {
				decl := decl_info_of_entity(e)
				if decl != nil && len(decl.Attributes) > 0 {
					error(decl.Attributes[0], "Constant alias declarations cannot have attributes")
				}
				override_entity_in_scope(e, entity)
				return
			}
			check_type_decl(ctx, e, ctx.Decl.InitExpr, named_type)
			return
		case Addressing_Builtin:
			if e.Type != nil {
				error(type_expr, "A constant alias of a built-in procedure may not have a type initializer")
			}
			e.Kind = Entity_Builtin
			e.Builtin.ID = operand.BuiltinID
			e.Type = t_invalid
			return
		case Addressing_ProcGroup:
			gb_assert_handler(operand.ProcGroup != nil, "operand.ProcGroup != nil", "cmd_check_decl_entity.go", 0)
			gb_assert_handler(operand.ProcGroup.Kind == Entity_ProcGroup, "operand.ProcGroup.Kind == Entity_ProcGroup", "cmd_check_decl_entity.go", 0)
			e.Kind = Entity_ProcGroup
			clone := make([]*Entity, len(operand.ProcGroup.ProcGroup.Entities))
			copy(clone, operand.ProcGroup.ProcGroup.Entities)
			e.ProcGroup.Entities = clone
			return
		}
		if entity != nil {
			if e.Type != nil {
				x := Operand{}
				x.Type = entity.Type
				x.Mode = Addressing_Variable
				if entity.Kind == Entity_Constant {
					x.Mode = Addressing_Constant
					x.Value = entity.Constant.Value
				}
				if !check_is_assignable_to(ctx, &x, e.Type) {
					expr_str := expr_to_string(init)
					defer gb_string_free(expr_str)
					op_type_str := type_to_string(entity.Type)
					defer gb_string_free(op_type_str)
					type_str := type_to_string(e.Type)
					defer gb_string_free(type_str)
					error(e.Token,
						"Cannot assign '%s' of type '%s' to '%s'",
						expr_str, op_type_str, type_str)
				}
			}
			switch entity.Kind {
			case Entity_ProcGroup, Entity_Procedure, Entity_LibraryName, Entity_ImportName:
				decl := decl_info_of_entity(e)
				if decl != nil && len(decl.Attributes) > 0 {
					error(decl.Attributes[0], "Constant alias declarations cannot have attributes")
				}
				override_entity_in_scope(e, entity)
				return
			}
		}
	}
	check_init_constant(ctx, e, &operand)
	if operand.Mode == Addressing_Invalid || base_type(operand.Type) == t_invalid {
		str := expr_to_string(init)
		defer gb_string_free(str)
		error(init, "Invalid declaration value '%s'", str)
	}
	decl := decl_info_of_entity(e)
	if decl != nil {
		check_decl_attributes(ctx, decl.Attributes, const_decl_attribute, nil)
	}
}

func check_global_variable_decl(ctx *CheckerContext, e *Entity, type_expr *Ast, init_expr *Ast) {
	gb_assert_handler(e.Type == nil, "e.Type == nil", "cmd_check_decl_entity.go", 0)
	gb_assert_handler(e.Kind == Entity_Variable, "e.Kind == Entity_Variable", "cmd_check_decl_entity.go", 0)
	if e.Flags&EntityFlag_Visited != 0 {
		e.Type = t_invalid
		return
	}
	e.Flags |= EntityFlag_Visited
	ac := make_attribute_context(e.Variable.LinkPrefix, e.Variable.LinkSuffix)
	ac.InitExprListCount = 1
	if init_expr == nil {
		ac.InitExprListCount = 0
	}
	decl := decl_info_of_entity(e)
	gb_assert_handler(decl == ctx.Decl, "decl == ctx.Decl", "cmd_check_decl_entity.go", 0)
	if decl != nil {
		check_decl_attributes(ctx, decl.Attributes, var_decl_attribute, &ac)
	}
	if ac.RequireDeclaration {
		e.Flags |= EntityFlag_Require
		mpsc_enqueue(&ctx.info.RequiredGlobalVariableQueue, e)
	}
	e.Variable.thread_local_model = ac.ThreadLocalModel
	e.Variable.IsExport = ac.IsExport
	e.Flags &^= EntityFlag_Static
	if ac.IsStatic {
		error(e.Token, "@(static) is not supported for global variables, nor required")
	}
	if ac.Rodata {
		e.Variable.is_rodata = true
	}
	ac.LinkName = handle_link_name(ctx, e.Token, ac.LinkName, ac.LinkPrefix, ac.LinkSuffix)
	if is_arch_wasm() && e.Variable.thread_local_model.len != 0 {
		e.Variable.thread_local_model = String{}
	}
	if buildContext.NoThreadLocal {
		e.Variable.thread_local_model = String{}
	}
	context_name := S("variable declaration")
	if type_expr != nil {
		e.Type = check_type(ctx, type_expr)
	}
	if e.Type != nil {
		if is_type_polymorphic(base_type(e.Type)) {
			str := type_to_string(e.Type)
			error(e.Token, "Invalid use of a polymorphic type '%s' in %.*s", str, context_name.len, context_name.data)
			gb_string_free(str)
			e.Type = t_invalid
		} else if is_type_empty_union(e.Type) {
			str := type_to_string(e.Type)
			error(e.Token, "An empty union '%s' cannot be instantiated in %.*s", str, context_name.len, context_name.data)
			gb_string_free(str)
			e.Type = t_invalid
		}
	}
	if e.Variable.IsForeign {
		if init_expr != nil {
			error(e.Token, "A foreign variable declaration cannot have a default value")
		}
		init_entity_foreign_library(ctx, e)
		if is_arch_wasm() && e.Variable.ForeignLibrary != nil {
			error(e.Token, "A foreign variable declaration can not be scoped to a module and must be declared in a 'foreign {' (without a library) block")
		}
	}
	if ac.LinkName.len > 0 {
		e.Variable.link_name = ac.LinkName
	}
	if ac.LinkSection.len > 0 {
		e.Variable.link_section = ac.LinkSection
	}
	if e.Variable.IsForeign || e.Variable.IsExport {
		name := e.Token.String
		if e.Variable.link_name.len > 0 {
			name = e.Variable.link_name
		}
		fp := &ctx.info.Foreigns
		key := string_hash_string(name)
		found := string_map_get(fp, key)
		if found != nil {
			f := *found
			pos := f.Token.Pos
			this_type := base_type(e.Type)
			other_type := base_type(f.Type)
			if !signature_parameter_similar_enough(this_type, other_type) {
				error(e.Token,
					"Foreign entity '%.*s' previously declared elsewhere with a different type\n"+
						"\tat %s",
					name.len, name.data, token_pos_to_string(pos))
			}
		} else {
			string_map_set(fp, key, e)
		}
	}
	if e.Variable.link_name.len > 0 {
		e.Flags |= EntityFlag_CustomLinkName
	}
	if init_expr == nil {
		if type_expr == nil {
			e.Type = t_invalid
		}
		return
	}
	o := Operand{}
	check_expr_with_type_hint(ctx, &o, init_expr, e.Type)
	if check_vet_shadowing_assignment(ctx.checker, e, init_expr) {
		error(e.Token, "Illegal declaration cycle of `%.*s`", e.Token.String.len, e.Token.String.data)
		o.Mode = Addressing_Invalid
		o.Type = t_invalid
		e.Type = t_invalid
	}
	check_init_variable(ctx, e, &o, S("variable declaration"))
	if e.Variable.is_rodata && o.Mode != Addressing_Constant {
		begin_error_block()
		error(o.Expr, "Variables declared with @(rodata) must have constant initialization")
		expr := unparen_expr(o.Expr)
		if is_type_struct(e.Type) && expr != nil && expr.Kind == Ast_CompoundLit {
			cl := &expr.CompoundLit
			for _, elem_ := range cl.Elems {
				elem := elem_
				if elem.Kind == Ast_FieldValue {
					elem = elem.FieldValue.Value
				}
				elem = unparen_expr(elem)
				ent := entity_of_node(elem)
				if elem.TAV.Mode != Addressing_Constant && ent == nil && elem.Kind != Ast_ProcLit {
					tok := ast_token(elem)
					pos := tok.Pos
					s := type_to_string(type_of_expr(elem))
					error_line("%s Element is not constant, which is required for @(rodata), of type %s\n", token_pos_to_string(pos), s)
					gb_string_free(s)
				}
			}
		}
		end_error_block()
	}
	check_rtti_type_disallowed(e.Token, e.Type, "A variable declaration is using a type, %s, which has been disallowed")
}

func check_proc_group_decl(ctx *CheckerContext, pg_entity *Entity, d *DeclInfo) {
	gb_assert_handler(pg_entity.Kind == Entity_ProcGroup, "pg_entity.Kind == Entity_ProcGroup", "cmd_check_decl_entity.go", 0)
	pge := &pg_entity.ProcGroup
	proc_group_name := pg_entity.Token.String
	pg := &d.InitExpr.ProcGroup
	gb_assert_handler(d.InitExpr.Kind == Ast_ProcGroup, "d.InitExpr.Kind == Ast_ProcGroup", "cmd_check_decl_entity.go", 0)
	pge.Entities = make([]*Entity, 0, len(pg.Args))
	pg_entity.Type = t_invalid
	entity_set := PtrSet[*Entity]{}
	ptr_set_init(&entity_set, 2*len(pg.Args))
	for _, arg := range pg.Args {
		var ent *Entity
		o := Operand{}
		if arg.Kind == Ast_Ident {
			ent = check_ident(ctx, &o, arg, nil, nil, true)
		} else if arg.Kind == Ast_SelectorExpr {
			ent = check_selector(ctx, &o, arg, nil)
		}
		if ent == nil {
			error(arg, "Expected a valid entity name in procedure group, got %.*s", len(astStrings[arg.Kind].data), astStrings[arg.Kind].data)
			continue
		}
		if ent.Kind == Entity_Variable {
			if !is_type_proc(ent.Type) {
				s := type_to_string(ent.Type)
				error(arg, "Expected a procedure, got %s", s)
				gb_string_free(s)
				continue
			}
		} else if ent.Kind != Entity_Procedure {
			error(arg, "Expected a procedure entity")
			continue
		}
		if ptr_set_update(&entity_set, ent) {
			error(arg, "Previous use of `%.*s` in procedure group", ent.Token.String.len, ent.Token.String.data)
			continue
		}
		pge.Entities = append(pge.Entities, ent)
	}
	ptr_set_destroy(&entity_set)
	for j := 0; j < len(pge.Entities); j++ {
		p := pge.Entities[j]
		if p.Type == t_invalid {
			continue
		}
		if p.Flags&EntityFlag_Disabled != 0 {
			continue
		}
		name := p.Token.String
		for k := j + 1; k < len(pge.Entities); k++ {
			q := pge.Entities[k]
			gb_assert_handler(p != q, "p != q", "cmd_check_decl_entity.go", 0)
			is_invalid := false
			pos := q.Token.Pos
			if q.Type == nil || q.Type == t_invalid {
				continue
			}
			begin_error_block()
			if q.Flags&EntityFlag_Disabled != 0 {
				end_error_block()
				continue
			}
			kind := are_proc_types_overload_safe(p.Type, q.Type)
			both_have_where_clauses := false
			if p.DeclInfo.ProcLit != nil && q.DeclInfo.ProcLit != nil {
				gb_assert_handler(p.DeclInfo.ProcLit.Kind == Ast_ProcLit, "p.DeclInfo.ProcLit.Kind == Ast_ProcLit", "cmd_check_decl_entity.go", 0)
				gb_assert_handler(q.DeclInfo.ProcLit.Kind == Ast_ProcLit, "q.DeclInfo.ProcLit.Kind == Ast_ProcLit", "cmd_check_decl_entity.go", 0)
				pl := &p.DeclInfo.ProcLit.ProcLit
				ql := &q.DeclInfo.ProcLit.ProcLit
				pw := pl.WhereToken.Kind != Token_Invalid && is_type_polymorphic(p.Type, true)
				qw := ql.WhereToken.Kind != Token_Invalid && is_type_polymorphic(q.Type, true)
				both_have_where_clauses = pw && qw
			}
			if !both_have_where_clauses {
				switch kind {
				case ProcOverload_Identical:
					error(p.Token, "Overloaded procedure '%.*s' has the same type as another procedure in the procedure group '%.*s'", name.len, name.data, proc_group_name.len, proc_group_name.data)
					is_invalid = true
				case ProcOverload_ParamVariadic:
					error(p.Token, "Overloaded procedure '%.*s' has the same type as another procedure in the procedure group '%.*s'", name.len, name.data, proc_group_name.len, proc_group_name.data)
					is_invalid = true
				case ProcOverload_ResultCount, ProcOverload_ResultTypes:
					error(p.Token, "Overloaded procedure '%.*s' has the same parameters but different results in the procedure group '%.*s'", name.len, name.data, proc_group_name.len, proc_group_name.data)
					is_invalid = true
				case ProcOverload_Polymorphic:
				case ProcOverload_ParamCount, ProcOverload_ParamTypes, ProcOverload_TargetFeatures:
				}
			}
			end_error_block()
			if is_invalid {
				begin_error_block()
				error_line("\tprevious procedure at %s\n", token_pos_to_string(pos))
				end_error_block()
				q.Type = t_invalid
			}
		}
	}
	ac := AttributeContext{}
	check_decl_attributes(ctx, d.Attributes, proc_group_attribute, &ac)
	check_objc_methods(ctx, pg_entity, ac)
}

func check_entity_decl(ctx *CheckerContext, e *Entity, d *DeclInfo, named_type *Type) {
	if e.State == EntityState_Resolved {
		return
	}
	if e.Flags&EntityFlag_Lazy != 0 {
		mutex_lock(&ctx.info.LazyMutex)
	}
	name := e.Token.String
	if e.Type != nil || e.State != EntityState_Unresolved {
		error(e.Token, "Illegal declaration cycle of `%.*s`", name.len, name.data)
	} else {
		gb_assert_handler(e.State == EntityState_Unresolved, "e.State == EntityState_Unresolved", "cmd_check_decl_entity.go", 0)
		if d == nil {
			d = decl_info_of_entity(e)
			if d == nil {
				e.Type = t_invalid
				e.State = EntityState_Resolved
				if named_type != nil {
					set_base_type(named_type, t_invalid)
				}
				goto end
			}
		}
		c := *ctx
		c.Scope = d.Scope
		c.Decl = d
		c.type_level = 0
		c.CurrProcCallingConvention = ProcCC_Contextless
		prev_flags := c.Scope.Flags
		defer func() { c.Scope.Flags = prev_flags }()
		if check_feature_flags(ctx, d.DeclNode)&OptInFeatureFlag_GlobalContext != 0 {
			c.Scope.Flags |= ScopeFlag_ContextDefined
		} else {
			c.Scope.Flags &^= ScopeFlag_ContextDefined
		}
		e.ParentProcDecl = c.CurrProcDecl
		e.State = EntityState_InProgress
		track_cycle_path := false
		switch e.Kind {
		case Entity_Variable, Entity_Constant, Entity_TypeName:
			track_cycle_path = true
		}
		if track_cycle_path {
			check_type_path_push(&c, e)
		}
		defer func() {
			if track_cycle_path {
				check_type_path_pop(&c)
			}
		}()
		switch e.Kind {
		case Entity_Variable:
			check_global_variable_decl(&c, e, d.TypeExpr, d.InitExpr)
		case Entity_Constant:
			check_const_decl(&c, e, d.TypeExpr, d.InitExpr, named_type)
		case Entity_TypeName:
			check_type_decl(&c, e, d.InitExpr, named_type)
		case Entity_Procedure:
			check_proc_decl(&c, e, d)
		case Entity_ProcGroup:
			check_proc_group_decl(&c, e, d)
		}
		e.State = EntityState_Resolved
	}
end:
	if e.Flags&EntityFlag_Lazy != 0 {
		ctx.info.Entities = append(ctx.info.Entities, e)
		mutex_unlock(&ctx.info.LazyMutex)
	}
}

func add_deps_from_child_to_parent(decl *DeclInfo) {
	if decl != nil && decl.Parent != nil {
		ps := decl.Parent.Scope
		if ps.Flags&(ScopeFlag_File|ScopeFlag_Pkg|ScopeFlag_Global) != 0 {
			return
		}
		rw_mutex_shared_lock(&decl.DepsMutex)
		rw_mutex_lock(&decl.Parent.DepsMutex)
		for _, e := range decl.Deps.Keys {
			if e != nil {
				ptr_set_add(&decl.Parent.Deps, e)
			}
		}
		rw_mutex_unlock(&decl.Parent.DepsMutex)
		rw_mutex_shared_unlock(&decl.DepsMutex)
		rw_mutex_shared_lock(&decl.TypeInfoDepsMutex)
		rw_mutex_lock(&decl.Parent.TypeInfoDepsMutex)
		for it := begin_type_set(&decl.TypeInfoDeps); !it.Equal(end_type_set(&decl.TypeInfoDeps)); it.Next() {
			type_set_add(&decl.Parent.TypeInfoDeps, it.Value())
		}
		rw_mutex_unlock(&decl.Parent.TypeInfoDepsMutex)
		rw_mutex_shared_unlock(&decl.TypeInfoDepsMutex)
	}
}

func check_proc_body(ctx_ *CheckerContext, token Token, decl *DeclInfo, typ *Type, body *Ast) bool {
	if body == nil {
		return false
	}
	gb_assert_handler(body.Kind == Ast_BlockStmt, "body.Kind == Ast_BlockStmt", "cmd_check_decl_entity.go", 0)
	proc_name := String{}
	if token.Kind == Token_Ident {
		proc_name = token.String
	} else {
		proc_name = S("(anonymous-procedure)")
	}
	new_ctx := *ctx_
	ctx := &new_ctx
	gb_assert_handler(typ.Kind == Type_Proc, "typ.Kind == Type_Proc", "cmd_check_decl_entity.go", 0)
	ctx.Scope = decl.Scope
	ctx.Decl = decl
	ctx.ProcName = proc_name
	ctx.CurrProcDecl = decl
	ctx.CurrProcSig = typ
	ctx.CurrProcCallingConvention = typ.Proc.CallingConvention
	if decl.Parent != nil && decl.Parent.Entity != nil {
		decl.Entity.ParentProcDecl = decl.Parent
	}
	if ctx.pkg.Name != "runtime" {
		switch typ.Proc.CallingConvention {
		case ProcCC_None:
			error(body, "Procedures with the calling convention \"none\" are not allowed a body")
		}
	}
	bs := &body.BlockStmt
	gb_assert_handler(body.Kind == Ast_BlockStmt, "body.Kind == Ast_BlockStmt", "cmd_check_decl_entity.go", 0)

	var using_entities []ProcUsingVar
	if typ.Proc.ParamCount > 0 {
		params := typ.Proc.Params.Tuple
		for _, param_e := range params.Variables {
			if param_e.Kind != Entity_Variable {
				continue
			}
			if is_type_polymorphic(param_e.Type) && is_type_polymorphic_record_unspecialized(param_e.Type) {
				s := type_to_string(param_e.Type)
				msg := "Unspecialized polymorphic types are not allowed in procedure parameters, got %s"
				if param_e.Variable.TypeExpr != nil {
					error(param_e.Variable.TypeExpr, msg, s)
				} else {
					error(param_e.Token, msg, s)
				}
				gb_string_free(s)
			}
			if param_e.Flags&EntityFlag_Using == 0 {
				continue
			}
			if is_blank_ident(param_e.Token) {
				error(param_e.Token, "'using' a procedure parameter requires a non blank identifier")
				break
			}
			is_value := param_e.Flags&EntityFlag_Value != 0 && !is_type_pointer(param_e.Type)
			t := base_type(type_deref(param_e.Type))
			if t.Kind == Type_Struct {
				scope := t.Struct.Scope
				gb_assert_handler(scope != nil, "scope != nil", "cmd_check_decl_entity.go", 0)
				rw_mutex_lock(&scope.Mutex)
				for _, entry := range scope.Elements {
					if entry.Hash != 0 {
						f := entry.Value
						if f.Kind == Entity_Variable {
							uvar := alloc_entity_using_variable(param_e, f.Token, f.Type, nil)
							if is_value {
								uvar.Flags |= EntityFlag_Value
							}
							using_entities = append(using_entities, ProcUsingVar{E: param_e, Uvar: uvar})
						}
					}
				}
				rw_mutex_unlock(&scope.Mutex)
			} else {
				error(param_e.Token, "'using' can only be applied to variables of type struct")
				break
			}
		}
	}
	rw_mutex_lock(&ctx.Scope.Mutex)
	for _, entry := range using_entities {
		uvar := entry.Uvar
		prev := scope_insert_no_mutex(ctx.Scope, uvar)
		if prev != nil {
			begin_error_block()
			error(entry.E.Token, "Namespace collision while 'using' procedure argument '%.*s' of: %.*s", entry.E.Token.String.len, entry.E.Token.String.data, prev.Token.String.len, prev.Token.String.data)
			error_line("%.*s != %.*s\n", uvar.Token.String.len, uvar.Token.String.data, prev.Token.String.len, prev.Token.String.data)
			end_error_block()
			break
		}
	}
	rw_mutex_unlock(&ctx.Scope.Mutex)
	where_clause_ok := evaluate_where_clauses(ctx, nil, decl.Scope, &decl.ProcLit.ProcLit.WhereClauses, !atomic.LoadUint32(&decl.WhereClausesEvaluated))
	if !where_clause_ok {
		return false
	}
	check_open_scope(ctx, body)
	ctx.Scope.DeclInfo = decl
	for _, entry := range using_entities {
		uvar := entry.Uvar
		scope_insert(ctx.Scope, uvar)
	}
	gb_assert_handler(decl.ProcCheckedState != ProcCheckedState_Checked, "decl.ProcCheckedState != ProcCheckedState_Checked", "cmd_check_decl_entity.go", 0)
	if atomic.LoadUint32(&decl.DeferUseChecked) != 0 {
		gb_assert_handler(is_type_polymorphic(typ, true), "is_type_polymorphic(typ, true)", "cmd_check_decl_entity.go", 0)
		error(token, "Defer Use Checked: %.*s", decl.Entity.Token.String.len, decl.Entity.Token.String.data)
		atomic.StoreUint32(&decl.DeferUseChecked, 0)
	}
	check_stmt_list(ctx, bs.Stmts, Stmt_CheckScopeDecls)
	atomic.StoreUint32(&decl.DeferUseChecked, 1)
	for _, stmt := range bs.Stmts {
		if stmt.Kind == Ast_ValueDecl {
			vd := &stmt.ValueDecl
			for _, name := range vd.Names {
				if !is_blank_ident(name) {
					if name.Kind == Ast_Ident {
						gb_assert_handler(name.Ident.Entity.Load() != nil, "name.Ident.Entity != nil", "cmd_check_decl_entity.go", 0)
					}
				}
			}
		}
	}
	if typ.Proc.ResultCount > 0 {
		if !check_is_terminating(body, "") {
			if token.Kind == Token_Ident {
				error(bs.Close, "Missing return statement at the end of the procedure '%.*s'", token.String.len, token.String.data)
			} else {
				error(bs.Close, "Missing return statement at the end of the procedure")
			}
		}
	} else if typ.Proc.Diverging {
		if !check_is_terminating(body, "") {
			if token.Kind == Token_Ident {
				error(bs.Close, "Missing diverging call at the end of the procedure '%.*s'", token.String.len, token.String.data)
			} else {
				error(bs.Close, "Missing diverging call at the end of the procedure")
			}
		}
	}
	check_close_scope(ctx)
	check_scope_usage(ctx.checker, ctx.Scope, check_vet_flags(body))
	add_deps_from_child_to_parent(decl)
	for _, vr := range decl.VariadicReuses {
		gb_assert_handler(vr.SliceType.Kind == Type_Slice, "vr.SliceType.Kind == Type_Slice", "cmd_check_decl_entity.go", 0)
		elem := vr.SliceType.Slice.Elem
		size := type_size_of(elem)
		align := type_align_of(elem)
		if size*vr.MaxCount > decl.VariadicReuseMaxBytes {
			decl.VariadicReuseMaxBytes = size * vr.MaxCount
		}
		if align > decl.VariadicReuseMaxAlign {
			decl.VariadicReuseMaxAlign = align
		}
	}
	return true
}

type ProcUsingVar struct {
	E    *Entity
	Uvar *Entity
}
