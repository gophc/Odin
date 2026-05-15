package cmd

import (
	"sort"
	"unsafe"
)

var print_entity_kind_ordering = [Entity_Count]int{
	Entity_Invalid:     -1,
	Entity_Constant:    0,
	Entity_Variable:    1,
	Entity_TypeName:    4,
	Entity_Procedure:   2,
	Entity_ProcGroup:   3,
	Entity_Nil:         -1,
	Entity_Label:       -1,
	Entity_Builtin:     -1,
	Entity_ImportName:  -1,
	Entity_LibraryName: -1,
}

var print_entity_names = [Entity_Count]string{
	Entity_Invalid:     "",
	Entity_Constant:    "constants",
	Entity_Variable:    "variables",
	Entity_TypeName:    "types",
	Entity_Procedure:   "procedures",
	Entity_ProcGroup:   "proc_group",
	Entity_Nil:         "",
	Entity_Label:       "",
	Entity_Builtin:     "",
	Entity_ImportName:  "import names",
	Entity_LibraryName: "library names",
}

func cmp_entities_for_printing(a, b *Entity) int {
	if a.Pkg != b.Pkg {
		if a.Pkg == nil {
			return -1
		}
		if b.Pkg == nil {
			return 1
		}
		res := string_compare(a.Pkg.Name, b.Pkg.Name)
		if res != 0 {
			return res
		}
	}
	ox := print_entity_kind_ordering[a.Kind]
	oy := print_entity_kind_ordering[b.Kind]
	res := ox - oy
	if res == 0 {
		res = string_compare(a.Token.String, b.Token.String)
	}
	return res
}

func cmp_entities_for_printing_by_order_in_src(a, b *Entity) int {
	if a.Pkg != b.Pkg {
		if a.Pkg == nil {
			return -1
		}
		if b.Pkg == nil {
			return 1
		}
		res := string_compare(a.Pkg.Name, b.Pkg.Name)
		if res != 0 {
			return res
		}
	}
	sx := a.OrderInSrc
	sy := b.OrderInSrc
	if sx < sy {
		return -1
	}
	if sx > sy {
		return 1
	}
	if a.Token.Pos.Offset < b.Token.Pos.Offset {
		return -1
	}
	if a.Token.Pos.Offset > b.Token.Pos.Offset {
		return 1
	}
	return 0
}

func cmp_ast_package_by_name(a, b *AstPackage) int {
	return string_compare(a.Name, b.Name)
}

func print_doc_line(indent i32, data String) {
	for i := i32(0); i < indent; i++ {
		gb_printf("\t")
	}
	gb_file_write(gb_file_get_standard(gbFileStandard_Output), data.Data, data.Len)
	gb_printf("\n")
}

func print_doc_line_no_newline(indent i32, data String) {
	for i := i32(0); i < indent; i++ {
		gb_printf("\t")
	}
	gb_file_write(gb_file_get_standard(gbFileStandard_Output), data.Data, data.Len)
}

func print_doc_comment_group_string(indent i32, g *CommentGroup) bool {
	if g == nil {
		return false
	}
	len_total := isize(0)
	for i := isize(0); i < isize(len(g.List)); i++ {
		comment := g.List[i].String
		len_total += comment.Len
		len_total += 1
	}
	if len_total <= isize(len(g.List)) {
		return false
	}
	count := isize(0)
	for i := isize(0); i < isize(len(g.List)); i++ {
		comment := g.List[i].String
		slash_slash := false
		if comment.Len >= 2 && *comment.Data == '/' {
			second := *(*u8)(unsafe.Pointer(uintptr(unsafe.Pointer(comment.Data)) + 1))
			if second == '/' {
				slash_slash = true
				comment.Data = (*u8)(unsafe.Pointer(uintptr(unsafe.Pointer(comment.Data)) + 2))
				comment.Len -= 2
			} else if second == '*' {
				comment.Data = (*u8)(unsafe.Pointer(uintptr(unsafe.Pointer(comment.Data)) + 2))
				comment.Len -= 4
			}
		}
		if comment.Len > 0 && *comment.Data == ' ' {
			comment.Data = (*u8)(unsafe.Pointer(uintptr(unsafe.Pointer(comment.Data)) + 1))
			comment.Len -= 1
		}
		if slash_slash {
			if string_starts_with(comment, S("+")) {
				continue
			}
			if string_starts_with(comment, S("@(")) {
				continue
			}
		}
		if slash_slash {
			print_doc_line(indent, comment)
			count += 1
		} else {
			pos := isize(0)
			for pos < comment.Len {
				end := pos
				for end < comment.Len {
					if *(*u8)(unsafe.Pointer(uintptr(unsafe.Pointer(comment.Data)) + uintptr(end))) == '\n' {
						break
					}
					end++
				}
				line := substring(comment, pos, end)
				pos = end
				trimmed_line := string_trim_whitespace(line)
				if trimmed_line.Len == 0 {
					if count == 0 {
						continue
					}
				}
				if string_starts_with(line, S("* ")) {
					line = substring(line, 2, line.Len)
				}
				print_doc_line(indent, line)
				count += 1
			}
		}
	}
	if count > 0 {
		gb_printf("\n")
		return true
	}
	return false
}

func print_doc_expr(expr *Ast) {
	var s string
	if build_context.cmd_doc_flags&CmdDocFlag_Short != 0 {
		s = expr_to_string_shorthand(expr)
	} else {
		s = expr_to_string(expr)
	}
	gb_file_write(gb_file_get_standard(gbFileStandard_Output), unsafe.StringData(s), isize(len(s)))
}

func print_doc_package(info *CheckerInfo, pkg *AstPackage) {
	if pkg == nil {
		return
	}
	gb_printf("package %s\n", goStr(pkg.Name))
	for i := isize(0); i < isize(len(pkg.Files)); i++ {
		f := pkg.Files[i]
		if f.PkgDecl != nil {
			print_doc_comment_group_string(1, f.PkgDecl.PackageDecl.Docs)
		}
	}
	if pkg.Scope != nil {
		entities := make([]*Entity, 0, len(pkg.Scope.Elements))
		for _, e := range pkg.Scope.Elements {
			switch e.Kind {
			case Entity_Invalid, Entity_Builtin, Entity_Nil, Entity_Label:
				continue
			case Entity_Constant, Entity_Variable, Entity_TypeName, Entity_Procedure, Entity_ProcGroup, Entity_ImportName, Entity_LibraryName:
			}
			if e.Pkg != pkg {
				continue
			}
			if !is_entity_exported(e) {
				continue
			}
			entities = append(entities, e)
		}
		in_src_order := build_context.cmd_doc_flags&CmdDocFlag_InSourceOrder != 0
		if in_src_order {
			sort.Slice(entities, func(i, j int) bool {
				return cmp_entities_for_printing_by_order_in_src(entities[i], entities[j]) < 0
			})
		} else {
			sort.Slice(entities, func(i, j int) bool {
				return cmp_entities_for_printing(entities[i], entities[j]) < 0
			})
		}
		show_docs := build_context.cmd_doc_flags&CmdDocFlag_Short == 0
		var curr_file *AstFile
		var curr_entity_kind EntityKind = Entity_Invalid
		for _, e := range entities {
			if in_src_order {
				if curr_file != e.File {
					if curr_file != nil {
						gb_printf("\n")
					}
					curr_file = e.File
					filename := remove_directory_from_path(curr_file.Fullpath)
					gb_printf("\tfile: %s\n", goStr(filename))
				}
			} else {
				if curr_entity_kind != e.Kind {
					if curr_entity_kind != Entity_Invalid {
						gb_printf("\n")
					}
					curr_entity_kind = e.Kind
					gb_printf("\t%s\n", print_entity_names[e.Kind])
				}
			}
			var type_expr *Ast
			var init_expr *Ast
			var docs *CommentGroup
			if e.DeclInfo != nil {
				type_expr = e.DeclInfo.TypeExpr
				init_expr = e.DeclInfo.InitExpr
				docs = e.DeclInfo.Docs
			}
			print_doc_line_no_newline(2, e.Token.String)
			if type_expr != nil {
				t := expr_to_string(type_expr)
				gb_printf(": %s ", t)
			} else {
				gb_printf(" :")
			}
			if e.Kind == Entity_Variable {
				if init_expr != nil {
					gb_printf("= ")
					print_doc_expr(init_expr)
				}
			} else {
				gb_printf(": ")
				print_doc_expr(init_expr)
			}
			gb_printf("\n")
			if show_docs {
				print_doc_comment_group_string(3, docs)
			}
		}
		gb_printf("\n")
	}
	if pkg.Fullpath.Len != 0 {
		gb_printf("\n")
		gb_printf("\tfullpath:\n")
		gb_printf("\t\t%s\n", goStr(pkg.Fullpath))
		gb_printf("\tfiles:\n")
		for i := isize(0); i < isize(len(pkg.Files)); i++ {
			f := pkg.Files[i]
			filename := remove_directory_from_path(f.Fullpath)
			print_doc_line(2, filename)
		}
	}
}

func generate_documentation(c *Checker) {
	info := &c.Info
	if build_context.cmd_doc_flags&CmdDocFlag_DocFormat != 0 {
		init_fullpath := c.Parser.InitFullpath
		var output_name String
		var output_base String
		if build_context.out_filepath.Len == 0 {
			output_name = remove_directory_from_path(init_fullpath)
			output_name = remove_extension_from_path(output_name)
			output_name = string_trim_whitespace(output_name)
			if output_name.Len == 0 {
				output_name = info.InitScope.Pkg.Name
			}
			output_base = output_name
		} else {
			output_name = build_context.out_filepath
			output_name = string_trim_whitespace(output_name)
			if output_name.Len == 0 {
				output_name = info.InitScope.Pkg.Name
			}
			pos := string_extension_position(output_name)
			if pos < 0 {
				output_base = output_name
			} else {
				output_base = substring(output_name, 0, pos)
			}
		}
		output_base = path_to_full_path(permanent_allocator(), output_base)
		output_file_path := goStr(output_base) + ".odin-doc"
		odin_doc_write(info, output_file_path)
	} else {
		pkgs := make([]*AstPackage, 0, len(info.Packages))
		for _, pkg := range info.Packages {
			if build_context.cmd_doc_flags&CmdDocFlag_AllPackages != 0 {
				pkgs = append(pkgs, pkg)
			} else {
				if pkg.Kind == Package_Init {
					pkgs = append(pkgs, pkg)
				} else if pkg.IsExtra {
					pkgs = append(pkgs, pkg)
				}
			}
		}
		sort.Slice(pkgs, func(i, j int) bool {
			return cmp_ast_package_by_name(pkgs[i], pkgs[j]) < 0
		})
		for i := range pkgs {
			print_doc_package(info, pkgs[i])
		}
	}
}
