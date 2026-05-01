package docs

// ============================================================================
// Text/doc output (mirrors src/docs.cpp print_doc_* functions)
// ============================================================================

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// DocPrinter handles text-based documentation output.
type DocPrinter struct {
	writer io.Writer
	flags  CmdDocFlag
}

// NewDocPrinter creates a printer that writes to stdout.
func NewDocPrinter(flags CmdDocFlag) *DocPrinter {
	return &DocPrinter{
		writer: os.Stdout,
		flags:  flags,
	}
}

// NewDocPrinterTo creates a printer that writes to a specific writer.
func NewDocPrinterTo(w io.Writer, flags CmdDocFlag) *DocPrinter {
	return &DocPrinter{
		writer: w,
		flags:  flags,
	}
}

func (p *DocPrinter) printf(format string, args ...any) {
	fmt.Fprintf(p.writer, format, args...)
}

func (p *DocPrinter) printLine(indent int, args ...any) {
	for i := 0; i < indent; i++ {
		p.printf("\t")
	}
	fmt.Fprintln(p.writer, args...)
}

func (p *DocPrinter) printLineNoNewline(indent int, args ...any) {
	for i := 0; i < indent; i++ {
		p.printf("\t")
	}
	fmt.Fprint(p.writer, args...)
}

func (p *DocPrinter) printFormatted(indent int, format string, args ...any) {
	for i := 0; i < indent; i++ {
		p.printf("\t")
	}
	p.printf(format, args...)
	p.printf("\n")
}

func (p *DocPrinter) isShort() bool {
	return p.flags&CmdDocFlagShort != 0
}

// PrintCommentGroup outputs a comment group.
func (p *DocPrinter) PrintCommentGroup(indent int, g *CommentGroup) bool {
	if g == nil || len(g.List) == 0 {
		return false
	}

	totalLen := 0
	for _, c := range g.List {
		totalLen += c.String.Len() + 1
	}
	if totalLen <= len(g.List) {
		return false
	}

	linesPrinted := 0
	for _, c := range g.List {
		comment := c.String.String()
		originalComment := comment

		if len(comment) < 2 {
			continue
		}

		slashSlash := false
		if comment[1] == '/' {
			slashSlash = true
			comment = comment[2:]
		} else if comment[1] == '*' {
			comment = comment[2 : len(comment)-2]
		}

		// Trim leading space
		if len(comment) > 0 && comment[0] == ' ' {
			comment = comment[1:]
		}

		if slashSlash {
			// Skip +foo and @(foo) meta comments
			if strings.HasPrefix(comment, "+") {
				continue
			}
			if strings.HasPrefix(comment, "@(") {
				continue
			}
		}

		if slashSlash {
			p.printLine(indent, comment)
			linesPrinted++
		} else {
			lines := strings.Split(comment, "\n")
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					if linesPrinted == 0 {
						continue
					}
				}
				if strings.HasPrefix(line, "* ") {
					line = line[2:]
				}
				p.printLine(indent, line)
				linesPrinted++
			}
		}

		_ = originalComment
	}

	if linesPrinted > 0 {
		p.printLine(0, "")
		return true
	}
	return false
}

// PrintExpr outputs an expression.
func (p *DocPrinter) PrintExpr(expr *Ast) {
	if expr == nil {
		return
	}
	// In the real Odin compiler, this calls expr_to_string or expr_to_string_shorthand.
	// We provide a placeholder that outputs a node kind marker.
	if p.isShort() {
		p.printf("<expr kind=%d>", expr.Kind)
	} else {
		p.printf("<expr kind=%d>", expr.Kind)
	}
}

// PrintPackage outputs documentation for a single package.
func (p *DocPrinter) PrintPackage(pkg *AstPackage) {
	if pkg == nil {
		return
	}

	p.printFormatted(0, "package %s", pkg.Name.String())

	for _, f := range pkg.Files {
		if f.PkgDecl != nil && f.PkgDecl.Kind == AstPackageDecl {
			p.PrintCommentGroup(1, nil) // In real code: f.PkgDecl.PackageDecl.docs
		}
	}

	if pkg.Scope == nil {
		return
	}

	// Collect all exported entities
	var entities []*Entity
	for _, elem := range pkg.Scope.Elements {
		e := elem.Value
		switch e.Kind {
		case EntityInvalid, EntityBuiltin, EntityNil, EntityLabel:
			continue
		case EntityConstant, EntityVariable, EntityTypeName, EntityProcedure,
			EntityProcGroup, EntityImportName, EntityLibraryName:
			// OK
		default:
			continue
		}
		if e.Pkg != pkg {
			continue
		}
		if !IsEntityExported(e, false) {
			continue
		}
		entities = append(entities, e)
	}

	inSrcOrder := p.flags&CmdDocFlagInSourceOrder != 0
	if inSrcOrder {
		SortEntitiesBySrcOrder(entities)
	} else {
		SortEntitiesByKind(entities)
	}

	showDocs := (p.flags & CmdDocFlagShort) == 0

	var currFile *AstFile
	var currEntityKind EntityKind

	for _, e := range entities {
		if inSrcOrder {
			if currFile != e.File {
				if currFile != nil {
					p.printLine(0, "")
				}
				currFile = e.File
				filename := removeDirectoryFromPath(e.File.Fullpath.String())
				p.printFormatted(1, "file: %s", filename)
			}
		} else {
			if currEntityKind != e.Kind {
				if currEntityKind != EntityInvalid {
					p.printLine(0, "")
				}
				currEntityKind = e.Kind
				p.printFormatted(1, "%s", printEntityNames[e.Kind])
			}
		}

		// Print entity name
		p.printLineNoNewline(2, e.Token.String.String())

		var typeExpr, initExpr *Ast
		var docs *CommentGroup
		if e.DeclInfo != nil {
			typeExpr = e.DeclInfo.TypeExpr
			initExpr = e.DeclInfo.InitExpr
			docs = e.DeclInfo.Docs
		}

		// Handle Variable and Constant specific overrides
		if e.Kind == EntityVariable {
			if e.Variable.Docs != nil {
				docs = e.Variable.Docs
			}
		} else if e.Kind == EntityConstant {
			if e.Constant.Docs != nil {
				docs = e.Constant.Docs
			}
		}

		if typeExpr != nil {
			p.printf(": ")
			p.PrintExpr(typeExpr)
			p.printf(" ")
		} else {
			p.printf(" :")
		}

		if e.Kind == EntityVariable && initExpr != nil {
			p.printf("= ")
			p.PrintExpr(initExpr)
		} else if initExpr != nil {
			p.printf(": ")
			p.PrintExpr(initExpr)
		}
		p.printf("\n")

		if showDocs {
			p.PrintCommentGroup(3, docs)
		}
	}

	p.printLine(0, "")

	// Print fullpath and files
	if pkg.Fullpath.Len() != 0 {
		p.printLine(0, "")
		p.printFormatted(1, "fullpath:")
		p.printFormatted(2, "%s", pkg.Fullpath.String())
		p.printFormatted(1, "files:")
		for _, f := range pkg.Files {
			filename := removeDirectoryFromPath(f.Fullpath.String())
			p.printLine(2, filename)
		}
	}
}

// PrintDocs generates documentation for all packages in the given info.
func (p *DocPrinter) PrintDocs(pkgs []*AstPackage) {
	SortPackagesByName(pkgs)

	for _, pkg := range pkgs {
		p.PrintPackage(pkg)
	}
}

// GenerateDocumentation is the top-level entry point for generating docs.
// It mirrors the C++ generate_documentation function.
func GenerateDocumentation(pkgs []*AstPackage, flags CmdDocFlag) string {
	var buf strings.Builder
	printer := NewDocPrinterTo(&buf, flags)
	printer.PrintDocs(pkgs)
	return buf.String()
}
