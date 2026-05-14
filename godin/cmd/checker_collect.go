package cmd

import (
	"sort"
	"unsafe"
)

func check_collect_value_decl(c *CheckerContext, decl *Ast) {
	if decl.StateFlags&StateFlag_BeenHandled != 0 {
		return
	}
	decl.StateFlags |= StateFlag_BeenHandled
	gb_assert_handler(decl.Kind == Ast_ValueDecl, "check_collect_value_decl: expected Ast_ValueDecl")
	vd := &decl.ValueDecl

	entity_visibility_kind := c.ForeignContext.VisibilityKind
	is_test := false
	is_init := false
	is_fini := false
	is_priv := false

	for i := 0; i < len(vd.Attributes); i++ {
		attr := vd.Attributes[i]
		if attr.Kind != Ast_Attribute {
			continue
		}
		elems := &attr.Attribute.Elems
		for j := 0; j < len(*elems); j++ {
			elem := (*elems)[j]
			var name String
			var value *Ast
			switch elem.Kind {
			case Ast_Ident:
				name = elem.Ident.Token.String
			case Ast_FieldValue:
				fv := &elem.FieldValue
				gb_assert_handler(fv.Field.Kind == Ast_Ident, "check_collect_value_decl: FieldValue field must be Ident")
				name = fv.Field.Ident.Token.String
				value = fv.Value
			default:
				continue
			}

			if name == "private" {
				kind := EntityVisiblity_PrivateToPackage
				success := false
				if value != nil {
					if value.Kind == Ast_BasicLit && value.BasicLit.Token.Kind == Token_String {
						v := String{}
						if value.TAV.Value.Kind == ExactValue_String {
							v = value.TAV.Value.ValueString
						}
						if v == "file" {
							kind = EntityVisiblity_PrivateToFile
							success = true
						} else if v == "package" {
							kind = EntityVisiblity_PrivateToPackage
							success = true
						}
					}
				} else {
					success = true
				}
				if !success {
					error(value, "'%.*s' expects no parameter, or a string literal containing \"file\" or \"package\"", name.Len, name.Data)
				} else {
					is_priv = true
				}
				if entity_visibility_kind >= kind {
					error(elem, "Previous declaration of '%.*s'", name.Len, name.Data)
				} else {
					entity_visibility_kind = kind
				}
				last := len(*elems) - 1
				(*elems)[j] = (*elems)[last]
				(*elems)[last] = nil
				*elems = (*elems)[:last]
				j -= 1
			} else if name == "test" {
				is_test = true
			} else if name == "init" {
				is_init = true
			} else if name == "fini" {
				is_fini = true
			}
		}
	}

	if is_priv && is_test {
		error(decl, "Attribute 'private' is not allowed on a test case")
		return
	}

	if entity_visibility_kind == EntityVisiblity_Public &&
		(c.Scope.Flags&int32(ScopeFlag_File)) != 0 &&
		c.Scope.File != nil {
		if c.Scope.File.Flags&AstFile_IsPrivateFile != 0 {
			entity_visibility_kind = EntityVisiblity_PrivateToFile
		} else if c.Scope.File.Flags&AstFile_IsPrivatePkg != 0 {
			entity_visibility_kind = EntityVisiblity_PrivateToPackage
		}
	}
	if entity_visibility_kind != EntityVisiblity_Public && (c.Scope.Flags&int32(ScopeFlag_File)) == 0 {
		error(decl, "Attribute 'private' is not allowed on a non file scope entity")
	}

	if vd.IsMutable {
		if (c.Scope.Flags & int32(ScopeFlag_File)) == 0 {
			return
		}
		for i := 0; i < len(vd.Names); i++ {
			name := vd.Names[i]
			var value *Ast = nil
			if i < len(vd.Values) {
				value = vd.Values[i]
			}
			if name.Kind != Ast_Ident {
				error(name, "A declaration's name must be an identifier, got %.*s", len(astStrings[name.Kind].Data), astStrings[name.Kind].Data)
				continue
			}
			e := alloc_entity_variable(c.Scope, name.Ident.Token, nil)
			e.Identifier.Store(name)
			e.File = c.File
			e.Variable.IsGlobal = true
			if entity_visibility_kind != EntityVisiblity_Public {
				e.Flags |= EntityFlag_NotExported
			}
			if vd.IsUsing {
				vd.IsUsing = false
				error(name, "'using' is not allowed at the file scope")
			}
			fl := c.ForeignContext.CurrLibrary
			if fl != nil {
				gb_assert_handler(fl.Kind == Ast_Ident, "foreign library must be Ident")
				e.Variable.IsForeign = true
				e.Variable.ForeignLibraryIdent = fl
				e.Variable.LinkPrefix = c.ForeignContext.LinkPrefix
				e.Variable.LinkSuffix = c.ForeignContext.LinkSuffix
			}
			init_expr := value
			d := make_decl_info(c.Scope, c.Decl)
			d.DeclNode = decl
			d.Comment = vd.Comment
			d.Docs = vd.Docs
			d.Entity = e
			d.TypeExpr = vd.Type
			d.InitExpr = init_expr
			d.Attributes = vd.Attributes
			is_exported := entity_visibility_kind != EntityVisiblity_PrivateToFile
			add_entity_and_decl_info(c, name, e, d, is_exported)
		}
		check_arity_match(c, decl, true)
	} else {
		for i := 0; i < len(vd.Names); i++ {
			name := vd.Names[i]
			if name.Kind != Ast_Ident {
				error(name, "A declaration's name must be an identifier, got %.*s", len(astStrings[name.Kind].Data), astStrings[name.Kind].Data)
				continue
			}
			init := unparen_expr(vd.Values[i])
			if init == nil {
				error(name, "Expected a value for this constant value declaration")
				continue
			}
			token := name.Ident.Token
			fl := c.ForeignContext.CurrLibrary
			var e *Entity
			d := make_decl_info(c.Scope, c.Decl)
			d.DeclNode = decl
			d.Comment = vd.Comment
			d.Docs = vd.Docs
			d.Attributes = vd.Attributes
			d.TypeExpr = vd.Type
			d.InitExpr = init
			if is_ast_type(init) {
				e = alloc_entity_type_name(d.Scope, token, nil)
			} else if init.Kind == Ast_ProcLit {
				if (c.Scope.Flags & int32(ScopeFlag_Type)) != 0 {
					error(name, "Procedure declarations are not allowed within a struct")
					continue
				}
				pl := &init.ProcLit
				e = alloc_entity_procedure(d.Scope, token, nil, pl.Tags)
				d.ForeignRequireResults = c.ForeignContext.RequireResults
				if fl != nil {
					gb_assert_handler(fl.Kind == Ast_Ident, "foreign library must be Ident")
					e.Procedure.ForeignLibraryIdent = fl
					e.Procedure.IsForeign = true
					gb_assert_handler(pl.Type.Kind == Ast_ProcType, "proc lit type must be ProcType")
					cc := pl.Type.ProcType.CallingConvention
					if cc == ProcCC_ForeignBlockDefault {
						cc = ProcCC_CDecl
						if c.ForeignContext.DefaultCC > 0 {
							cc = c.ForeignContext.DefaultCC
						} else if is_arch_wasm() {
							begin_error_block()
							error(init, "For wasm related targets, it is required that you either define the"+
								" @(default_calling_convention=<string>) on the foreign block or"+
								" explicitly assign it on the procedure signature")
							error_line("\tSuggestion: when dealing with normal Odin code (e.g. js_wasm32), use \"contextless\"; when dealing with Emscripten like code, use \"c\"\n")
							end_error_block()
						}
					}
					e.Procedure.LinkPrefix = c.ForeignContext.LinkPrefix
					e.Procedure.LinkSuffix = c.ForeignContext.LinkSuffix
					gb_assert_handler(cc != ProcCC_Invalid, "calling convention must be valid")
					pl.Type.ProcType.CallingConvention = cc
				}
				d.ProcLit = init
				d.InitExpr = init
				if is_test {
					e.Flags |= EntityFlag_Test
				}
				if is_init && is_fini {
					error(name, "A procedure cannot be both declared as @(init) and @(fini)")
				} else if is_init {
					e.Flags |= EntityFlag_Init
				} else if is_fini {
					e.Flags |= EntityFlag_Fini
				}
			} else if init.Kind == Ast_ProcGroup {
				e = alloc_entity_proc_group(d.Scope, token, nil)
				if fl != nil {
					error(name, "Procedure groups are not allowed within a foreign block")
				}
			} else {
				e = alloc_entity_constant(d.Scope, token, nil, emptyExactValue)
			}
			e.Identifier.Store(name)
			if entity_visibility_kind != EntityVisiblity_Public {
				e.Flags |= EntityFlag_NotExported
			}
			add_entity_flags_from_file(c, e, c.Scope)
			if vd.IsUsing {
				if e.Kind == Entity_TypeName && init.Kind == Ast_EnumType {
					d.IsUsing = true
				} else {
					error(name, "'using' is not allowed on this constant value declaration")
				}
			}
			if e.Kind != Entity_Procedure {
				if fl != nil {
					begin_error_block()
					kind := init.Kind
					error(name, "Only procedures and variables are allowed to be in a foreign block, got %.*s", len(astStrings[kind].Data), astStrings[kind].Data)
					if kind == Ast_ProcType {
						error_line("\tDid you forget to append '---' to the procedure?\n")
					}
					end_error_block()
				}
			}
			check_builtin_attributes(c, e, &d.Attributes)
			is_exported := entity_visibility_kind != EntityVisiblity_PrivateToFile
			add_entity_and_decl_info(c, name, e, d, is_exported)
		}
		check_arity_match(c, decl, true)
	}
}

func check_add_foreign_block_decl(ctx *CheckerContext, decl *Ast) bool {
	gb_assert_handler(decl.Kind == Ast_ForeignBlockDecl, "check_add_foreign_block_decl: expected Ast_ForeignBlockDecl")
	fb := &decl.ForeignBlockDecl
	foreign_library := fb.ForeignLibrary
	c := *ctx
	if foreign_library.Kind == Ast_Ident {
		c.ForeignContext.CurrLibrary = foreign_library
	} else {
		error(foreign_library, "Foreign block name must be an identifier or 'export'")
		c.ForeignContext.CurrLibrary = nil
	}
	check_decl_attributes(&c, fb.Attributes, foreign_block_decl_attribute, nil)
	body := fb.Body
	gb_assert_handler(body.Kind == Ast_BlockStmt, "foreign block body must be BlockStmt")
	block := &body.BlockStmt
	if c.CollectDelayedDecls && (c.Scope.Flags&int32(ScopeFlag_File)) != 0 {
		return collect_file_decls(&c, block.Stmts)
	}
	check_collect_entities(&c, block.Stmts)
	return false
}

func correct_single_type_alias(c *CheckerContext, e *Entity) bool {
	if e.Kind == Entity_Constant {
		d := e.DeclInfo
		if d != nil && d.InitExpr != nil {
			init := d.InitExpr
			alias_of := check_entity_from_ident_or_selector(c, init, true)
			if alias_of != nil && alias_of.Kind == Entity_TypeName {
				e.Kind = Entity_TypeName
				return true
			}
		}
	}
	return false
}

func correct_type_alias_in_scope_backwards(c *CheckerContext, s *Scope) bool {
	correction := false
	for i := uint32(0); i < s.Elements.Cap; i++ {
		slot := s.Elements.Slots[i]
		if slot.Hash != 0 && slot.Value != nil {
			if correct_single_type_alias(c, slot.Value) {
				correction = true
			}
		}
	}
	return correction
}

func correct_type_alias_in_scope_forwards(c *CheckerContext, s *Scope) bool {
	correction := false
	for i := uint32(0); i < s.Elements.Cap; i++ {
		slot := s.Elements.Slots[i]
		if slot.Hash != 0 && slot.Value != nil {
			if correct_single_type_alias(c, slot.Value) {
				correction = true
			}
		}
	}
	return correction
}

func correct_type_aliases_in_scope(c *CheckerContext, s *Scope) {
	for {
		corrections := false
		if correct_type_alias_in_scope_backwards(c, s) {
			corrections = true
		}
		if correct_type_alias_in_scope_forwards(c, s) {
			corrections = true
		}
		if !corrections {
			return
		}
	}
}

func check_collect_entities(c *CheckerContext, nodes []*Ast) {
	var curr_file *AstFile = nil
	if (c.Scope.Flags & int32(ScopeFlag_File)) != 0 {
		curr_file = c.Scope.File
		gb_assert_handler(curr_file != nil, "check_collect_entities: file scope must have a file")
	}

	for decl_index := 0; decl_index < len(nodes); decl_index++ {
		decl := nodes[decl_index]
		if !is_ast_decl(decl) && !is_when_stmt(decl) {
			if curr_file != nil && decl.Kind == Ast_ExprStmt {
				expr := decl.ExprStmt.Expr
				if expr.Kind == Ast_CallExpr && expr.CallExpr.Proc.Kind == Ast_BasicDirective {
					if c.CollectDelayedDecls {
						if decl.StateFlags&StateFlag_BeenHandled != 0 {
							return
						}
						decl.StateFlags |= StateFlag_BeenHandled
						curr_file.DelayedDeclsQueues[AstDelayQueue_Expr] = append(curr_file.DelayedDeclsQueues[AstDelayQueue_Expr], expr)
					}
					continue
				}
			}
			continue
		}
		switch decl.Kind {
		case Ast_BadDecl:
		case Ast_WhenStmt:
		case Ast_ValueDecl:
			check_collect_value_decl(c, decl)
		case Ast_ImportDecl:
			if curr_file == nil {
				error(decl, "import declarations are only allowed in the file scope")
				continue
			}
			curr_file.DelayedDeclsQueues[AstDelayQueue_Import] = append(curr_file.DelayedDeclsQueues[AstDelayQueue_Import], decl)
		case Ast_ForeignImportDecl:
			if (c.Scope.Flags & int32(ScopeFlag_File)) == 0 {
				error(decl, "%.*s declarations are only allowed in the file scope", decl.ForeignImportDecl.Token.String.Len, decl.ForeignImportDecl.Token.String.Data)
				continue
			}
			check_add_foreign_import_decl(c, decl)
		case Ast_ForeignBlockDecl:
			if curr_file != nil {
				curr_file.DelayedDeclsQueues[AstDelayQueue_ForeignBlock] = append(curr_file.DelayedDeclsQueues[AstDelayQueue_ForeignBlock], decl)
			}
		default:
			if (c.Scope.Flags & int32(ScopeFlag_File)) != 0 {
				error(decl, "Only declarations are allowed at file scope")
			}
		}
	}

	if curr_file == nil {
		for decl_index := 0; decl_index < len(nodes); decl_index++ {
			decl := nodes[decl_index]
			if decl.Kind == Ast_ForeignBlockDecl {
				check_add_foreign_block_decl(c, decl)
			}
		}
		for decl_index := 0; decl_index < len(nodes); decl_index++ {
			decl := nodes[decl_index]
			if decl.Kind == Ast_WhenStmt {
				check_collect_entities_from_when_stmt(c, &decl.WhenStmt)
			}
		}
	}
}

func create_checker_context(c *Checker) *CheckerContext {
	ctx := permanentAllocItem[CheckerContext]()
	init_checker_context(ctx, c)
	return ctx
}

func check_single_global_entity(c *Checker, e *Entity, d *DeclInfo) {
	gb_assert_handler(e != nil, "check_single_global_entity: entity must not be nil")
	gb_assert_handler(d != nil, "check_single_global_entity: decl must not be nil")
	if d.Scope != e.Scope {
		return
	}
	if e.State == EntityState_Resolved {
		return
	}
	ctx := create_checker_context(c)
	gb_assert_handler((d.Scope.Flags&int32(ScopeFlag_File)) != 0, "check_single_global_entity: scope must be file scope")
	file := d.Scope.File
	add_curr_ast_file(ctx, file)
	pkg := file.Pkg
	gb_assert_handler(ctx.Pkg != nil, "check_single_global_entity: pkg must not be nil")
	gb_assert_handler(e.Pkg != nil, "check_single_global_entity: entity pkg must not be nil")
	ctx.Decl = d
	ctx.Scope = d.Scope
	if pkg.Kind == Package_Init {
		if e.Kind != Entity_Procedure && e.Token.String == "main" {
			error(e.Token, "'main' is reserved as the entry point procedure in the initial scope")
			return
		}
	}
	check_entity_decl(ctx, e, d, nil)
}

func check_all_global_entities(c *Checker) {
	store_in_single_threaded_checker_stage(true)
	for i := 0; i < len(c.Info.Entities); i++ {
		e := c.Info.Entities[i]
		gb_assert_handler(e != nil, "check_all_global_entities: entity must not be nil")
		if e.Flags&EntityFlag_Lazy != 0 {
			continue
		}
		d := e.DeclInfo
		check_single_global_entity(c, e, d)
		if e.Type != nil && is_type_typed(e.Type) {
			for {
				t := mpsc_dequeue(c.SoaTypesToComplete, &c.SoaTypesToComplete)
				if t == nil {
					break
				}
				complete_soa_type(c, t, false)
			}
			type_size_of(e.Type)
			type_align_of(e.Type)
		}
	}
	store_in_single_threaded_checker_stage(false)
}

func is_string_an_identifier(s String) bool {
	offset := isize(0)
	if s.Len < 1 {
		return false
	}
	for offset < s.Len {
		ok := false
		r := Rune(-1)
		size := utf8_decode(s.Data[offset:], &r)
		if offset == 0 {
			ok = rune_is_letter(r)
		} else {
			ok = rune_is_letter(r) || rune_is_digit(r)
		}
		if !ok {
			return false
		}
		offset += size
	}
	return offset == s.Len
}

func path_to_entity_name(name String, fullpath String, strip_extension ...bool) String {
	stripExt := true
	if len(strip_extension) > 0 {
		stripExt = strip_extension[0]
	}
	if name.Len != 0 {
		return name
	}
	filename := fullpath
	slash := isize(0)
	dot := isize(0)
	for i := filename.Len - 1; i >= 0; i-- {
		c := filename.Data[i]
		if c == '/' || c == '\\' {
			break
		}
		slash = i
	}
	filename = substring(filename, slash, filename.Len)
	if stripExt {
		dot = filename.Len
		for dot > 0 {
			dot--
			c := filename.Data[dot]
			if c == '.' {
				break
			}
		}
		if dot > 0 {
			filename = substring(filename, 0, dot)
		}
	}
	if is_string_an_identifier(filename) {
		return filename
	}
	return makeString("_")
}

func add_import_dependency_node(c *Checker, decl *Ast, M *PtrMap[*AstPackage, *ImportGraphNode]) {
	parent_pkg := decl.file().Pkg
	switch decl.Kind {
	case Ast_ImportDecl:
		id := &decl.ImportDecl
		path := id.Fullpath
		if is_package_name_reserved(path) {
			return
		}
		found := string_map_get(&c.Info.Packages, path)
		if found == nil {
			token := ast_token(decl)
			error(token, "Unable to find package: %.*s", path.Len, path.Data)
			exit_with_errors()
		}
		child_pkg := *found
		gb_assert_handler(child_pkg.Scope != nil, "add_import_dependency_node: child pkg scope must not be nil")
		id.Package = child_pkg
		found_node := PtrMapGet(M, child_pkg)
		gb_assert_handler(found_node != nil, "add_import_dependency_node: child node must exist")
		child := *found_node
		found_node = PtrMapGet(M, parent_pkg)
		gb_assert_handler(found_node != nil, "add_import_dependency_node: parent node must exist")
		parent := *found_node
		import_graph_node_set_add(&parent.Succ, child)
		import_graph_node_set_add(&child.Pred, parent)
		ptr_set_add(&parent.Scope.Imported, child.Scope)
	case Ast_WhenStmt:
		ws := &decl.WhenStmt
		if ws.Body != nil {
			stmts := ws.Body.BlockStmt.Stmts
			for i := 0; i < len(stmts); i++ {
				add_import_dependency_node(c, stmts[i], M)
			}
		}
		if ws.ElseStmt != nil {
			switch ws.ElseStmt.Kind {
			case Ast_BlockStmt:
				stmts := ws.ElseStmt.BlockStmt.Stmts
				for i := 0; i < len(stmts); i++ {
					add_import_dependency_node(c, stmts[i], M)
				}
			case Ast_WhenStmt:
				add_import_dependency_node(c, ws.ElseStmt, M)
			}
		}
	}
}

func generate_import_dependency_graph(c *Checker, allocator gbAllocator) []*ImportGraphNode {
	M := PtrMapNew[*AstPackage, *ImportGraphNode]()
	defer map_destroy(M)
	for i := 0; i < len(c.Parser.Packages); i++ {
		pkg := c.Parser.Packages[i]
		n := import_graph_node_create(pkg)
		PtrMapSet(M, pkg, n)
	}
	for i := 0; i < len(c.Parser.Packages); i++ {
		p := c.Parser.Packages[i]
		for j := 0; j < len(p.Files); j++ {
			f := p.Files[j]
			for k := 0; k < len(f.Decls); k++ {
				decl := f.Decls[k]
				add_import_dependency_node(c, decl, M)
			}
		}
	}
	G := make([]*ImportGraphNode, 0, PtrMapCount(M))
	i := isize(0)
	for entry := range PtrMapIterate(M) {
		n := entry.Value
		n.Index = i
		i++
		n.DepCount = isize(len(n.Succ.Keys))
		gb_assert_handler(n.DepCount >= 0, "generate_import_dependency_graph: dep_count must be >= 0")
		G = append(G, n)
	}
	return G
}

type ImportPathItem struct {
	Pkg  *AstPackage
	Decl *Ast
}

func find_import_path(c *Checker, start *AstPackage, end *AstPackage, visited *PtrSet[*AstPackage], allocator gbAllocator) []ImportPathItem {
	empty_path := []ImportPathItem{}
	if ptr_set_update(visited, start) {
		return empty_path
	}
	path := start.Fullpath
	found := string_map_get(&c.Info.Packages, path)
	if found != nil {
		pkg := *found
		gb_assert_handler(pkg != nil, "find_import_path: pkg must not be nil")
		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			for j := 0; j < len(f.Imports); j++ {
				var child_pkg *AstPackage = nil
				decl := f.Imports[j]
				if decl.Kind == Ast_ImportDecl {
					child_pkg = decl.ImportDecl.Package
				} else {
					continue
				}
				if child_pkg == nil || child_pkg.Scope == nil {
					continue
				}
				item := ImportPathItem{Pkg: child_pkg, Decl: decl}
				if child_pkg == end {
					path := make([]ImportPathItem, 0, 1)
					path = append(path, item)
					return path
				}
				next_path := find_import_path(c, child_pkg, end, visited, allocator)
				if len(next_path) > 0 {
					next_path = append(next_path, item)
					return next_path
				}
			}
		}
	}
	return empty_path
}

func get_invalid_import_name(input String) String {
	slash := isize(0)
	for i := input.Len - 1; i >= 0; i-- {
		if input.Data[i] == '/' || input.Data[i] == '\\' {
			break
		}
		slash = i
	}
	input = substring(input, slash, input.Len)
	return input
}

func import_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", name.Len, name.Data)
		}
		return true
	} else if name == "require" {
		if value != nil {
			error(elem, "Expected no parameter for '%.*s'", name.Len, name.Data)
		}
		ac.RequireDeclaration = true
		return true
	}
	return false
}

func check_add_import_decl(ctx *CheckerContext, decl *Ast) {
	if decl.StateFlags&StateFlag_BeenHandled != 0 {
		return
	}
	decl.StateFlags |= StateFlag_BeenHandled
	gb_assert_handler(decl.Kind == Ast_ImportDecl, "check_add_import_decl: expected Ast_ImportDecl")
	id := &decl.ImportDecl
	token := id.Relpath
	parent_scope := ctx.Scope
	gb_assert_handler((parent_scope.Flags&int32(ScopeFlag_File)) != 0, "check_add_import_decl: scope must be file scope")
	pkgs := &ctx.Checker.Info.Packages
	var scope *Scope = nil
	force_use := false
	if id.Fullpath == "builtin" {
		scope = builtinPkg.Scope
		force_use = true
	} else if id.Fullpath == "intrinsics" {
		scope = intrinsicsPkg.Scope
		force_use = true
	} else {
		found := string_map_get(pkgs, id.Fullpath)
		if found == nil {
			for entry := range *pkgs {
				pkg := entry.Value
				gb_printf_err("%.*s\n", pkg.Fullpath.Len, pkg.Fullpath.Data)
			}
			gb_printf_err("%s\n", token_pos_to_string(token.Pos))
			gb_assert_handler(false, "check_add_import_decl: unable to find scope for package")
		} else {
			pkg := *found
			scope = pkg.Scope
		}
	}
	gb_assert_handler((scope.Flags&int32(ScopeFlag_Pkg)) != 0, "check_add_import_decl: scope must be pkg scope")
	ptr_set_add(&parent_scope.Imported, scope)
	import_name := path_to_entity_name(id.ImportName.String, id.Fullpath, false)
	if is_blank_ident_str(import_name) {
		force_use = true
	}
	ac := &AttributeContext{}
	check_decl_attributes(ctx, id.Attributes, import_decl_attribute, ac)
	if ac.RequireDeclaration {
		force_use = true
	}
	if is_blank_ident_str(import_name) && !is_blank_ident(id.ImportName) {
		invalid_name := id.Fullpath
		invalid_name = get_invalid_import_name(invalid_name)
		begin_error_block()
		if id.ImportName.String.Len > 0 {
			error(token, "Import name '%.*s' cannot be use as an import name as it is not a valid identifier", id.ImportName.String.Len, id.ImportName.String.Data)
		} else {
			error(id.Token, "Import name '%.*s' is not a valid identifier", invalid_name.Len, invalid_name.Data)
			error_line("\tSuggestion: Rename the directory or explicitly set an import name like this 'import <new_name> %.*s'", id.Relpath.String.Len, id.Relpath.String.Data)
		}
		end_error_block()
	} else {
		gb_assert_handler(id.ImportName.Pos.Line != 0, "check_add_import_decl: import_name pos line must be non-zero")
		id.ImportName.String = import_name
		e := alloc_entity_import_name(parent_scope, id.ImportName, t_invalid,
			id.Fullpath, import_name.String(), scope)
		add_entity(ctx, parent_scope, nil, e)
		if force_use {
			add_entity_use(ctx, nil, e)
		}
	}
	scope.Flags |= int32(ScopeFlag_HasBeenImported)
}

func foreign_import_decl_attribute(c *CheckerContext, elem *Ast, name String, value *Ast, ac *AttributeContext) bool {
	if name == "tag" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", name.Len, name.Data)
		}
		return true
	} else if name == "export" {
		ac.IsExport = true
		return true
	} else if name == "force" || name == "require" {
		if value != nil {
			error(elem, "Expected no parameter for '%.*s'", name.Len, name.Data)
		} else if name == "force" {
			error(elem, "'force' was replaced with 'require'")
		}
		ac.RequireDeclaration = true
		return true
	} else if name == "priority_index" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_Integer {
			error(elem, "Expected an integer value for '%.*s'", name.Len, name.Data)
		} else {
			ac.ForeignImportPriorityIndex = exact_value_to_i64(ev)
		}
		return true
	} else if name == "extra_linker_flags" {
		ev := check_decl_attribute_value(c, value)
		if ev.Kind != ExactValue_String {
			error(elem, "Expected a string value for '%.*s'", name.Len, name.Data)
		} else {
			ac.ExtraLinkerFlags = ev.ValueString
		}
		return true
	} else if name == "ignore_duplicates" {
		if value != nil {
			error(elem, "Expected no parameter for '%.*s'", name.Len, name.Data)
		}
		ac.IgnoreDuplicates = true
		return true
	}
	return false
}

func check_foreign_import_fullpaths(c *Checker) {
	ctx := CheckerContext{}
	init_checker_context(&ctx, c)
	untyped := UntypedExprInfoMap(PtrMapNew[*Ast, *ExprInfo]())
	defer map_destroy(&untyped)

	for {
		e := mpsc_dequeue(c.Info.ForeignImportsToCheckFullpaths, &c.Info.ForeignImportsToCheckFullpaths)
		if e == nil {
			break
		}
		gb_assert_handler(e != nil, "check_foreign_import_fullpaths: entity must not be nil")
		gb_assert_handler(e.Kind == Entity_LibraryName, "check_foreign_import_fullpaths: entity must be LibraryName")
		decl := e.LibraryName.Decl
		gb_assert_handler(decl.Kind == Ast_ForeignImportDecl, "check_foreign_import_fullpaths: decl must be ForeignImportDecl")
		fl := &decl.ForeignImportDecl
		f := decl.file()
		reset_checker_context(&ctx, f, &untyped)
		ctx.CollectDelayedDecls = false
		gb_assert_handler(ctx.Scope == e.Scope, "check_foreign_import_fullpaths: scope mismatch")

		if len(fl.Paths) == 0 {
			base_dir := dir_from_path(decl.file().Fullpath)
			fullpaths := make([]String, 0, len(fl.Filepaths))
			for _, fp_node := range fl.Filepaths {
				op := Operand{}
				check_expr(&ctx, &op, fp_node)
				if op.Mode != Addressing_Constant && op.Value.Kind != ExactValue_String {
					s := expr_to_string(op.Expr)
					error(fp_node, "Expected a constant string value, got '%s'", goStr(s))
					gb_string_free(s)
					continue
				}
				if !is_type_string(op.Type) {
					s := type_to_string(op.Type)
					error(fp_node, "Expected a constant string value, got value of type '%s'", goStr(s))
					gb_string_free(s)
					continue
				}
				file_str := op.Value.ValueString
				file_str = string_trim_whitespace(file_str)
				fullpath := file_str
				if !is_arch_wasm() || string_ends_with(file_str, makeString(".o")) {
					var foreign_path String
					ok := determine_path_from_string(nil, decl, base_dir, file_str, &foreign_path, true)
					if ok {
						fullpath = foreign_path
					}
				}
				fullpaths = append(fullpaths, fullpath)
			}
			fl.Paths = fullpaths
		}

		for _, path := range fl.Paths {
			ext := path_extension(path)
			if str_eq_ignore_case(ext, makeString(".c")) ||
				str_eq_ignore_case(ext, makeString(".cpp")) ||
				str_eq_ignore_case(ext, makeString(".cxx")) ||
				str_eq_ignore_case(ext, makeString(".h")) ||
				str_eq_ignore_case(ext, makeString(".hpp")) ||
				str_eq_ignore_case(ext, makeString(".hxx")) {
				error(fl.Token, "With 'foreign import', you cannot import a %.*s file/directory, you must precompile the library and link against that", ext.Len, ext.Data)
				break
			}
		}
		add_untyped_expressions(ctx.Info, &untyped)
		e.LibraryName.Paths = fl.Paths
	}

	for {
		e := mpsc_dequeue(c.Info.ForeignDeclsToCheck, &c.Info.ForeignDeclsToCheck)
		if e == nil {
			break
		}
		gb_assert_handler(e != nil, "check_foreign_import_fullpaths: foreign decl must not be nil")
		if e.Kind != Entity_Procedure {
			continue
		}
		if !is_arch_wasm() {
			continue
		}
		foreign_library := e.Procedure.ForeignLibrary
		gb_assert_handler(foreign_library != nil, "check_foreign_import_fullpaths: foreign library must not be nil")
		name := e.Procedure.LinkName
		module_name := makeString("env")
		gb_assert_handler(foreign_library.Kind == Entity_LibraryName, "check_foreign_import_fullpaths: foreign library must be LibraryName")
		if len(foreign_library.LibraryName.Paths) != 1 {
			error(foreign_library.Token, "'foreign import' for '%.*s' architecture may only have one path, got %d",
				len(targetArchNames[buildContext.Metrics.Arch].Data), targetArchNames[buildContext.Metrics.Arch].Data,
				len(foreign_library.LibraryName.Paths))
		}
		if len(foreign_library.LibraryName.Paths) >= 1 {
			module_name = foreign_library.LibraryName.Paths[0]
		}
		if !string_ends_with(module_name, makeString(".o")) {
			name = concatenate3_strings(permanent_allocator(), module_name, WASM_MODULE_NAME_SEPARATOR, name)
		}
		e.Procedure.LinkName = name
		check_foreign_procedure(&ctx, e, e.DeclInfo)
	}
}

func check_add_foreign_import_decl(ctx *CheckerContext, decl *Ast) {
	if decl.StateFlags&StateFlag_BeenHandled != 0 {
		return
	}
	decl.StateFlags |= StateFlag_BeenHandled
	gb_assert_handler(decl.Kind == Ast_ForeignImportDecl, "check_add_foreign_import_decl: expected Ast_ForeignImportDecl")
	fl := &decl.ForeignImportDecl
	parent_scope := ctx.Scope
	gb_assert_handler((parent_scope.Flags&int32(ScopeFlag_File)) != 0, "check_add_foreign_import_decl: scope must be file scope")
	library_name := fl.LibraryName.String
	if library_name.Len == 0 && len(fl.Paths) != 0 {
		fullpath := fl.Paths[0]
		library_name = path_to_entity_name(fl.LibraryName.String, fullpath)
	}
	if library_name.Len == 0 || is_blank_ident_str(library_name) {
		error(fl.Token, "File name, '%.*s', cannot be as a library name as it is not a valid identifier", library_name.Len, library_name.Data)
		return
	}
	gb_assert_handler(fl.LibraryName.Pos.Line != 0, "check_add_foreign_import_decl: library_name pos line must be non-zero")
	fl.LibraryName.String = library_name
	ac := &AttributeContext{}
	check_decl_attributes(ctx, fl.Attributes, foreign_import_decl_attribute, ac)
	scope := parent_scope
	if ac.IsExport {
		scope = parent_scope.Parent
	}
	e := alloc_entity_library_name(parent_scope, fl.LibraryName, t_invalid,
		fl.Paths, library_name.String())
	e.LibraryName.Decl = decl
	add_entity_flags_from_file(ctx, e, parent_scope)
	add_entity(ctx, scope, nil, e)
	if ac.RequireDeclaration {
		mpsc_enqueue(ctx.Info.RequiredForeignImportsThroughForceQueue, e)
		add_entity_use(ctx, nil, e)
	}
	if ac.ForeignImportPriorityIndex != 0 {
		e.LibraryName.PriorityIndex = ac.ForeignImportPriorityIndex
	}
	if ac.IgnoreDuplicates {
		e.LibraryName.IgnoreDuplicates = true
	}
	extra_linker_flags := string_trim_whitespace(ac.ExtraLinkerFlags)
	if extra_linker_flags.Len != 0 {
		e.LibraryName.ExtraLinkerFlags = extra_linker_flags
	}
	mpsc_enqueue(ctx.Info.ForeignImportsToCheckFullpaths, e)
}

func collect_when_stmt_from_file(ctx *CheckerContext, ws *AstWhenStmt) bool {
	operand := Operand{Mode: Addressing_Invalid}
	if !ws.IsConditionDetermined {
		check_expr(ctx, &operand, ws.Cond)
		if operand.Mode != Addressing_Invalid && !is_type_boolean(operand.Type) {
			error(ws.Cond, "Non-boolean condition in 'when' statement")
		}
		if operand.Mode != Addressing_Constant {
			error(ws.Cond, "Non-constant condition in 'when' statement")
		}
		ws.IsConditionDetermined = true
		ws.DeterminedCond = operand.Value.Kind == ExactValue_Bool && operand.Value.ValueBool
	}
	if ws.Body == nil || ws.Body.Kind != Ast_BlockStmt {
		error(ws.Cond, "Invalid body for 'when' statement")
	} else {
		if ws.DeterminedCond {
			check_collect_entities(ctx, ws.Body.BlockStmt.Stmts)
			return true
		} else if ws.ElseStmt != nil {
			switch ws.ElseStmt.Kind {
			case Ast_BlockStmt:
				check_collect_entities(ctx, ws.ElseStmt.BlockStmt.Stmts)
				return true
			case Ast_WhenStmt:
				collect_when_stmt_from_file(ctx, &ws.ElseStmt.WhenStmt)
				return true
			default:
				error(ws.ElseStmt, "Invalid 'else' statement in 'when' statement")
			}
		}
	}
	return false
}

func collect_file_decls_from_when_stmt(ctx *CheckerContext, ws *AstWhenStmt) bool {
	operand := Operand{Mode: Addressing_Invalid}
	if !ws.IsConditionDetermined {
		check_expr(ctx, &operand, ws.Cond)
		if operand.Mode != Addressing_Invalid && !is_type_boolean(operand.Type) {
			error(ws.Cond, "Non-boolean condition in 'when' statement")
		}
		if operand.Mode != Addressing_Constant {
			error(ws.Cond, "Non-constant condition in 'when' statement")
		}
		ws.IsConditionDetermined = true
		ws.DeterminedCond = operand.Value.Kind == ExactValue_Bool && operand.Value.ValueBool
	}
	if ws.Body == nil || ws.Body.Kind != Ast_BlockStmt {
		error(ws.Cond, "Invalid body for 'when' statement")
	} else {
		if ws.DeterminedCond {
			return collect_file_decls(ctx, ws.Body.BlockStmt.Stmts)
		} else if ws.ElseStmt != nil {
			switch ws.ElseStmt.Kind {
			case Ast_BlockStmt:
				return collect_file_decls(ctx, ws.ElseStmt.BlockStmt.Stmts)
			case Ast_WhenStmt:
				return collect_file_decls_from_when_stmt(ctx, &ws.ElseStmt.WhenStmt)
			default:
				error(ws.ElseStmt, "Invalid 'else' statement in 'when' statement")
			}
		}
	}
	return false
}

func collect_file_decl(ctx *CheckerContext, decl *Ast) bool {
	gb_assert_handler((ctx.Scope.Flags&int32(ScopeFlag_File)) != 0, "collect_file_decl: scope must be file scope")
	curr_file := ctx.Scope.File
	gb_assert_handler(curr_file != nil, "collect_file_decl: file must not be nil")
	if decl.StateFlags&StateFlag_BeenHandled != 0 {
		return false
	}
	switch decl.Kind {
	case Ast_ValueDecl:
		check_collect_value_decl(ctx, decl)
	case Ast_ImportDecl:
		check_add_import_decl(ctx, decl)
	case Ast_ForeignImportDecl:
		check_add_foreign_import_decl(ctx, decl)
	case Ast_ForeignBlockDecl:
		gb_assert_handler(ctx.CollectDelayedDecls, "collect_file_decl: must be collecting delayed decls")
		decl.StateFlags |= StateFlag_BeenHandled
		curr_file.DelayedDeclsQueues[AstDelayQueue_ForeignBlock] = append(curr_file.DelayedDeclsQueues[AstDelayQueue_ForeignBlock], decl)
	case Ast_WhenStmt:
		ws := &decl.WhenStmt
		if !ws.IsConditionDetermined {
			if collect_when_stmt_from_file(ctx, ws) {
				return true
			}
			nctx := *ctx
			nctx.CollectDelayedDecls = true
			if collect_file_decls_from_when_stmt(&nctx, ws) {
				return true
			}
		} else {
			nctx := *ctx
			nctx.CollectDelayedDecls = true
			if collect_file_decls_from_when_stmt(&nctx, ws) {
				return true
			}
		}
	case Ast_ExprStmt:
		gb_assert_handler(ctx.CollectDelayedDecls, "collect_file_decl: must be collecting delayed decls for ExprStmt")
		decl.StateFlags |= StateFlag_BeenHandled
		if decl.ExprStmt.Expr.Kind == Ast_CallExpr {
			ce := &decl.ExprStmt.Expr.CallExpr
			if ce.Proc.Kind == Ast_BasicDirective {
				curr_file.DelayedDeclsQueues[AstDelayQueue_Expr] = append(curr_file.DelayedDeclsQueues[AstDelayQueue_Expr], decl.ExprStmt.Expr)
			}
		}
	}
	return false
}

func collect_file_decls(ctx *CheckerContext, decls []*Ast) bool {
	gb_assert_handler((ctx.Scope.Flags&int32(ScopeFlag_File)) != 0, "collect_file_decls: scope must be file scope")
	for i := 0; i < len(decls); i++ {
		if collect_file_decl(ctx, decls[i]) {
			correct_type_aliases_in_scope(ctx, ctx.Scope)
			return true
		}
	}
	correct_type_aliases_in_scope(ctx, ctx.Scope)
	return false
}

func sort_file_by_name(a, b *AstFile) int {
	x_name := filename_from_path(a.Fullpath)
	y_name := filename_from_path(b.Fullpath)
	return string_compare(x_name, y_name)
}

func check_create_file_scopes(c *Checker) {
	for i := 0; i < len(c.Parser.Packages); i++ {
		pkg := c.Parser.Packages[i]
		sort.SliceStable(pkg.Files, func(a, b int) bool {
			return sort_file_by_name(pkg.Files[a], pkg.Files[b]) < 0
		})
		total_pkg_decl_count := isize(0)
		for j := 0; j < len(pkg.Files); j++ {
			f := pkg.Files[j]
			string_map_set(&c.Info.Files, f.Fullpath, f)
			create_scope_from_file(nil, f)
			total_pkg_decl_count += f.TotalFileDeclCount
		}
		mpmc_init(&pkg.ExportedEntityQueue, total_pkg_decl_count)
	}
}

type CollectEntityWorkerData struct {
	C       *Checker
	Ctx     CheckerContext
	Untyped UntypedExprInfoMap
}

var collect_entity_worker_data []CollectEntityWorkerData

func check_collect_entities_all_worker_proc(data unsafe.Pointer) isize {
	wd := &collect_entity_worker_data[current_thread_index()]
	c := wd.C
	ctx := &wd.Ctx
	untyped := &wd.Untyped
	f := (*AstFile)(data)
	reset_checker_context(ctx, f, untyped)
	check_collect_entities(ctx, f.Decls)
	gb_assert_handler(ctx.CollectDelayedDecls == false, "check_collect_entities_all_worker_proc: must not be collecting delayed decls")
	add_untyped_expressions(&c.Info, ctx.Untyped)
	return 0
}

func check_collect_entities_all(c *Checker) {
	thread_count := isize(len(globalThreadPool.Threads))
	collect_entity_worker_data = make([]CollectEntityWorkerData, thread_count)
	for i := isize(0); i < thread_count; i++ {
		wd := &collect_entity_worker_data[i]
		wd.C = c
		init_checker_context(&wd.Ctx, c)
		wd.Untyped = UntypedExprInfoMap(PtrMapNew[*Ast, *ExprInfo]())
	}
	for entry := range c.Info.Files {
		f := entry.Value
		thread_pool_add_task(check_collect_entities_all_worker_proc, unsafe.Pointer(f))
	}
	thread_pool_wait()
}

func check_export_entities_in_pkg(ctx *CheckerContext, pkg *AstPackage, untyped *UntypedExprInfoMap) {
	if len(pkg.Files) != 0 {
		for {
			item, ok := mpmc_dequeue(&pkg.ExportedEntityQueue)
			if !ok {
				break
			}
			f := item.Entity.File
			if ctx.File != f {
				reset_checker_context(ctx, f, untyped)
			}
			add_entity(ctx, pkg.Scope, item.Identifier, item.Entity)
			add_untyped_expressions(ctx.Info, untyped)
		}
	}
}

func check_export_entities_worker_proc(data unsafe.Pointer) isize {
	pkg := (*AstPackage)(data)
	wd := &collect_entity_worker_data[current_thread_index()]
	check_export_entities_in_pkg(&wd.Ctx, pkg, &wd.Untyped)
	return 0
}

func check_export_entities(c *Checker) {
	thread_count := isize(len(globalThreadPool.Threads))
	for i := isize(0); i < thread_count; i++ {
		wd := &collect_entity_worker_data[i]
		PtrMapClear(wd.Untyped)
		init_checker_context(&wd.Ctx, c)
	}
	for entry := range c.Info.Packages {
		pkg := entry.Value
		thread_pool_add_task(check_export_entities_worker_proc, unsafe.Pointer(pkg))
	}
	thread_pool_wait()
}

func check_import_entities(c *Checker) {
	dep_graph := generate_import_dependency_graph(c, temporary_allocator())

	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, makeString("check_import_entities - sort packages"))
	}

	pq := priority_queue_create(dep_graph, import_graph_node_cmp, import_graph_node_swap)
	emitted := PtrSetNew[*AstPackage]()
	defer ptr_set_destroy(emitted)
	package_order := make([]*ImportGraphNode, 0, len(c.Parser.Packages))

	for pq.Queue.Count > 0 {
		n := priority_queue_pop(&pq)

		pkg := n.Pkg
		if n.DepCount > 0 {
			visited := PtrSetNew[*AstPackage]()
			path := find_import_path(c, pkg, pkg, visited, temporary_allocator())
			if len(path) > 1 {
				item := path[len(path)-1]
				pkg_name := item.Pkg.Name
				error(item.Decl, "Cyclic importation of '%.*s'", pkg_name.Len, pkg_name.Data)
				for i := 0; i < len(path); i++ {
					error(item.Decl, "'%.*s' refers to", pkg_name.Len, pkg_name.Data)
					item = path[i]
					pkg_name = item.Pkg.Name
				}
				error(item.Decl, "'%.*s'", pkg_name.Len, pkg_name.Data)
			}
			ptr_set_destroy(visited)
		}

		for _, p := range n.Pred.Keys {
			if p == nil {
				continue
			}
			p.DepCount = max(p.DepCount-1, 0)
			priority_queue_fix(&pq, p.Index)
		}

		if pkg == nil {
			continue
		}
		if ptr_set_update(emitted, pkg) {
			continue
		}
		package_order = append(package_order, n)
	}

	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, makeString("check_import_entities - collect file decls"))
	}

	ctx := CheckerContext{}
	init_checker_context(&ctx, c)
	untyped := UntypedExprInfoMap(PtrMapNew[*Ast, *ExprInfo]())
	defer map_destroy(&untyped)

	min_pkg_index := isize(0)
	for pkg_index := isize(0); pkg_index < isize(len(package_order)); pkg_index++ {
		node := package_order[pkg_index]
		pkg := node.Pkg
		pkg.Order = 1 + pkg_index
		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			reset_checker_context(&ctx, f, &untyped)
			ctx.CollectDelayedDecls = true
			for _, decl := range f.DelayedDeclsQueues[AstDelayQueue_Import] {
				check_add_import_decl(&ctx, decl)
			}
			f.DelayedDeclsQueues[AstDelayQueue_Import] = f.DelayedDeclsQueues[AstDelayQueue_Import][:0]
			if collect_file_decls(&ctx, f.Decls) {
				check_export_entities_in_pkg(&ctx, pkg, &untyped)
				pkg_index = min_pkg_index - 1
				break
			}
			add_untyped_expressions(ctx.Info, &untyped)
		}
		if pkg_index < 0 {
			continue
		}
		min_pkg_index = pkg_index
	}

	if buildContext.ShowMoreTimings {
		timings_start_section(&globalTimings, makeString("check_import_entities - check delayed entities"))
	}

	for pkg_index := isize(0); pkg_index < isize(len(package_order)); pkg_index++ {
		node := package_order[pkg_index]
		gb_assert_handler((node.Scope.Flags&int32(ScopeFlag_Pkg)) != 0, "check_import_entities: node scope must be pkg scope")
		pkg := node.Scope.Pkg

		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			reset_checker_context(&ctx, f, &untyped)
			for _, decl := range f.DelayedDeclsQueues[AstDelayQueue_Import] {
				check_add_import_decl(&ctx, decl)
			}
			f.DelayedDeclsQueues[AstDelayQueue_Import] = f.DelayedDeclsQueues[AstDelayQueue_Import][:0]
			add_untyped_expressions(ctx.Info, &untyped)
		}

		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			reset_checker_context(&ctx, f, &untyped)
			correct_type_aliases_in_scope(&ctx, pkg.Scope)
		}

		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			reset_checker_context(&ctx, f, &untyped)
			ctx.CollectDelayedDecls = true
			will_recheck_foreign_block := false
			for _, decl := range f.DelayedDeclsQueues[AstDelayQueue_ForeignBlock] {
				if check_add_foreign_block_decl(&ctx, decl) {
					pkg_index -= 1
					will_recheck_foreign_block = true
					break
				}
			}
			if will_recheck_foreign_block {
				break
			}
			f.DelayedDeclsQueues[AstDelayQueue_ForeignBlock] = f.DelayedDeclsQueues[AstDelayQueue_ForeignBlock][:0]
		}

		for i := 0; i < len(pkg.Files); i++ {
			f := pkg.Files[i]
			reset_checker_context(&ctx, f, &untyped)
			for _, expr := range f.DelayedDeclsQueues[AstDelayQueue_Expr] {
				o := Operand{}
				check_expr(&ctx, &o, expr)
			}
			f.DelayedDeclsQueues[AstDelayQueue_Expr] = f.DelayedDeclsQueues[AstDelayQueue_Expr][:0]
			add_untyped_expressions(ctx.Info, &untyped)
		}
	}
}
