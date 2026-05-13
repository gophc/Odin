package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unsafe"
)

func check_defines(bc *BuildContext, c *Checker) {
	for key, value := range bc.DefinedValues {
		name := make_string_c(key)
		found := false
		for _, def := range c.Info.Defineables {
			if def.Name == name {
				found = true
				break
			}
		}

		if !found {
			warning(nil, "given -define:%s is unused in the project", unsafe.String(name.Data, name.Len))

			if !global_ignore_warnings() {
				error_line("\tSuggestion: use the -show-defineables flag for an overview of the possible defines\n")
			}
		}
	}
}

func temp_alloc_defineable_strings(c *Checker) {
	for i := range c.Info.Defineables {
		def := &c.Info.Defineables[i]
		def.DefaultValueStr = make_string_c(write_exact_value_to_string(gb_string_make(temporary_allocator(), ""), def.DefaultValue))
		def.PosStr = make_string_c(token_pos_to_string(def.Pos))
	}
}

func defineables_cmp(x, y *Defineable) int {
	xFile := get_file_path_string(x.Pos.FileId)
	yFile := get_file_path_string(y.Pos.FileId)
	cmp := string_compare(xFile, yFile)
	if cmp != 0 {
		return cmp
	}
	return int(x.Pos.Offset - y.Pos.Offset)
}

func sort_defineables_and_remove_duplicates(c *Checker) {
	if len(c.Info.Defineables) == 0 {
		return
	}

	sort.Slice(c.Info.Defineables, func(i, j int) bool {
		return defineables_cmp(&c.Info.Defineables[i], &c.Info.Defineables[j]) < 0
	})

	prev := c.Info.Defineables[0]
	for i := 1; i < len(c.Info.Defineables); {
		curr := c.Info.Defineables[i]
		if prev.Pos == curr.Pos {
			c.Info.Defineables = append(c.Info.Defineables[:i], c.Info.Defineables[i+1:]...)
			continue
		}
		prev = curr
		i++
	}
}

func export_defineables(c *Checker, path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export defineables to: %s\n", path)
		os.Exit(1)
		return
	} else {
		fmt.Printf("Exporting defineables to '%s'...\n", path)
	}
	defer f.Close()

	var docs strings.Builder

	fmt.Fprintf(f, "Defineable,Default Value,Docs,Location\n")
	for _, def := range c.Info.Defineables {
		docs.Reset()
		if def.Docs != nil {
			docs.WriteString("\"")
			for _, token := range def.Docs.List {
				for i := 0; i < int(token.String.Len); i++ {
					c := *(*byte)(unsafe.Add(token.String.Data, i))
					if c == '"' {
						docs.WriteString("\"\"")
					} else {
						docs.WriteByte(c)
					}
				}
			}
			docs.WriteString("\"")
		}

		fmt.Fprintf(f, "%s,%s,%s,%s\n",
			unsafe.String(def.Name.Data, def.Name.Len),
			unsafe.String(def.DefaultValueStr.Data, def.DefaultValueStr.Len),
			docs.String(),
			unsafe.String(def.PosStr.Data, def.PosStr.Len),
		)
	}
}

func show_defineables(c *Checker) {
	for _, def := range c.Info.Defineables {
		if has_ansi_terminal_colours() {
			fmt.Printf("\x1b[0;90m")
		}
		fmt.Printf("%s\n", unsafe.String(def.PosStr.Data, def.PosStr.Len))
		if def.Docs != nil {
			for _, token := range def.Docs.List {
				fmt.Printf("%s\n", unsafe.String(token.String.Data, token.String.Len))
			}
		}
		if has_ansi_terminal_colours() {
			fmt.Printf("\x1b[0m")
		}
		fmt.Printf("%s :: %s\n\n",
			unsafe.String(def.Name.Data, def.Name.Len),
			unsafe.String(def.DefaultValueStr.Data, def.DefaultValueStr.Len),
		)
	}
}

func print_show_unused(c *Checker) {
	info := &c.Info

	unused := make([]*Entity, 0, len(info.Entities))
	for _, e := range info.Entities {
		if e == nil {
			continue
		}
		if e.Pkg == nil || e.Pkg.Scope == nil {
			continue
		}
		if e.Pkg.Scope.Flags&ScopeFlag_Builtin != 0 {
			continue
		}
		switch e.Kind {
		case Entity_Invalid, Entity_Builtin, Entity_Nil, Entity_Label:
			continue
		case Entity_Constant, Entity_Variable, Entity_TypeName, Entity_Procedure,
			Entity_ProcGroup, Entity_ImportName, Entity_LibraryName:
		}
		if e.Scope.Flags&(ScopeFlag_Pkg|ScopeFlag_File) == 0 {
			continue
		}
		if e.Token.String.Len == 0 {
			continue
		}
		if unsafe.String(e.Token.String.Data, e.Token.String.Len) == "_" {
			continue
		}

		if e.MinDepCount > 0 {
			continue
		}
		unused = append(unused, e)
	}

	sort.Slice(unused, func(i, j int) bool {
		return cmp_entities_for_printing(unused[i], unused[j]) < 0
	})

	print_usage_line(0, "Unused Package Declarations")

	var currPkg *AstPackage
	var currEntityKind EntityKind
	for _, e := range unused {
		if currPkg != e.Pkg {
			currPkg = e.Pkg
			currEntityKind = Entity_Invalid
			print_usage_line(0, "")
			print_usage_line(0, "package %s", unsafe.String(currPkg.Name.Data, currPkg.Name.Len))
		}
		if currEntityKind != e.Kind {
			currEntityKind = e.Kind
			print_usage_line(1, "%s", print_entity_names[e.Kind])
		}
		if build_context.show_unused_with_location {
			pos := e.Token.Pos
			print_usage_line(2, "%s %s", token_pos_to_string(pos), unsafe.String(e.Token.String.Data, e.Token.String.Len))
		} else {
			print_usage_line(2, "%s", unsafe.String(e.Token.String.Data, e.Token.String.Len))
		}
	}
	print_usage_line(0, "")
}
