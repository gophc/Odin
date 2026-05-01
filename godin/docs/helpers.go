package docs

import "fmt"

// ============================================================================
// Helper methods for Writer (token casts, comment strings, expr strings, etc.)
// ============================================================================

// tokenPosCast converts an Entity TokenPos to an OdinDocPosition.
func (w *Writer) tokenPosCast(pos TokenPos) OdinDocPosition {
	fileIndex := OdinDocFileIndex(0)
	if pos.FileID != 0 {
		// In the real compiler this would look up by file_id.
		// For now, we store it as-is if we have the file in cache.
		_ = fileIndex
	}
	return OdinDocPosition{
		File:   fileIndex,
		Line:   uint32(pos.Line),
		Column: uint32(pos.Column),
		Offset: uint32(pos.Offset),
	}
}

// writeCommentGroupString converts a CommentGroup to an OdinDocString.
func (w *Writer) writeCommentGroupString(g *CommentGroup) OdinDocString {
	if g == nil || len(g.List) == 0 {
		return OdinDocString{}
	}

	// Calculate needed buffer size
	totalLen := 0
	for _, c := range g.List {
		totalLen += c.String.Len() + 1
	}
	if totalLen <= len(g.List) {
		return OdinDocString{}
	}

	var buf []byte
	count := 0
	for _, c := range g.List {
		comment := c.String.String()
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

		if len(comment) > 0 && comment[0] == ' ' {
			comment = comment[1:]
		}

		if slashSlash {
			if hasPrefix(comment, "+") {
				continue
			}
			if hasPrefix(comment, "@(") {
				continue
			}
		}

		if slashSlash {
			buf = append(buf, comment...)
			buf = append(buf, '\n')
			count++
		} else {
			// Block comment: process line by line
			pos := 0
			for pos < len(comment) {
				end := pos
				for end < len(comment) && comment[end] != '\n' {
					end++
				}
				line := comment[pos:end]
				pos = end + 1

				trimmed := trimWhitespace(line)
				if len(trimmed) == 0 {
					if count == 0 {
						continue
					}
				}
				if hasPrefix(line, "* ") {
					line = line[2:]
				}
				buf = append(buf, line...)
				buf = append(buf, '\n')
				count++
			}
		}
	}

	if count > 0 {
		buf = append(buf, '\n')
	}

	return w.writeString(string(buf))
}

// writePkgDocString generates package-level documentation from all file package declarations.
func (w *Writer) writePkgDocString(pkg *AstPackage) OdinDocString {
	if pkg == nil {
		return OdinDocString{}
	}

	var buf []byte
	for _, f := range pkg.Files {
		if f.PkgDecl != nil && f.PkgDecl.Kind == AstPackageDecl {
			// In the real compiler: append comment group for pkg decl docs
			_ = f
		}
	}
	return w.writeStringWithoutCache(string(buf))
}

// writeExprString converts an AST expression to a string and writes it.
func (w *Writer) writeExprString(expr *Ast) OdinDocString {
	if expr == nil {
		return OdinDocString{}
	}
	// In the real compiler this calls write_expr_to_string or expr_to_string.
	// We provide a placeholder.
	s := fmt.Sprintf("<expr kind=%d>", expr.Kind)
	return w.writeString(s)
}

// writeAttributes converts an array of attribute AST nodes to OdinDocAttributes.
func (w *Writer) writeAttributes(attributes []*Ast) OdinDocArray {
	if len(attributes) == 0 {
		return OdinDocArray{}
	}

	// Count attribute elems
	count := 0
	for _, attr := range attributes {
		if attr.Kind != AstAttribute {
			continue
		}
		count++ // In real code: attr.Attribute.elems.count
	}

	attribs := make([]OdinDocAttribute, 0, count)

	for _, attr := range attributes {
		if attr.Kind != AstAttribute {
			continue
		}
		// Each attribute element becomes an OdinDocAttribute
		// In real code this iterates attr.Attribute.elems
		// For now we emit a placeholder
		attribs = append(attribs, OdinDocAttribute{
			Name:  w.writeString("placeholder"),
			Value: w.writeExprString(nil),
		})
	}

	return w.writeAttributeSlice(attribs)
}

// writeWhereClauses converts where clause AST nodes.
func (w *Writer) writeWhereClauses(clauses []*Ast) OdinDocArray {
	if len(clauses) == 0 {
		return OdinDocArray{}
	}

	docStrings := make([]OdinDocString, len(clauses))
	for i, clause := range clauses {
		docStrings[i] = w.writeExprString(clause)
	}
	return w.writeDocStringSlice(docStrings)
}

// ============================================================================
// writeDocs: the main documentation writing pass
// ============================================================================

// writeDocs writes all packages, files, entities, and updates entity references.
func (w *Writer) writeDocs(info any) {
	// Extract packages from info. The info is expected to be a []*AstPackage.
	var allPkgs []*AstPackage
	switch v := info.(type) {
	case []*AstPackage:
		allPkgs = v
	case *[]*AstPackage:
		allPkgs = *v
	default:
		return
	}

	// Sort packages by name
	SortPackagesByName(allPkgs)

	for _, pkg := range allPkgs {
		// Determine package flags
		var pkgFlags uint32
		switch pkg.Kind {
		case PackageRuntime:
			pkgFlags |= uint32(OdinDocPkgFlagRuntime)
		case PackageInit:
			pkgFlags |= uint32(OdinDocPkgFlagInit)
		case PackageBuiltin:
			pkgFlags |= uint32(OdinDocPkgFlagBuiltin)
		}

		docPkg := OdinDocPkg{
			Fullpath: w.writeString(pkg.Fullpath.String()),
			Name:     w.writeString(pkg.Name.String()),
			Flags:    pkgFlags,
			Docs:     w.writePkgDocString(pkg),
		}

		_, _ = w.writePkg(&docPkg)

		// Register package in cache
		// In the real code this is done via writeItem return; simplified here

		// Write files for this package
		fileIndices := make([]uint32, 0, len(pkg.Files))
		for _, file := range pkg.Files {
			docFile := OdinDocFile{
				Pkg:  0, // Will be filled properly in real code
				Name: w.writeString(file.Fullpath.String()),
			}
			fileIndex, _ := w.writeFile(&docFile)
			w.fileCache[file] = fileIndex
			fileIndices = append(fileIndices, fileIndex)
		}
		docPkg.Files = w.writeUint32Slice(fileIndices)

		// Write scope entries
		docPkg.Entries = w.addPkgEntries(pkg)

		// Update the package entry
		_ = docPkg
	}

	// Update entity types and foreign libraries
	w.updateEntities()
}

// addPkgEntries writes all exported entities from a package's scope.
func (w *Writer) addPkgEntries(pkg *AstPackage) OdinDocArray {
	if pkg.Scope == nil {
		return OdinDocArray{}
	}

	var entries []OdinDocScopeEntry

	for _, elem := range pkg.Scope.Elements {
		e := elem.Value

		switch e.Kind {
		case EntityInvalid, EntityNil, EntityLabel:
			continue
		case EntityConstant, EntityVariable, EntityTypeName, EntityProcedure,
			EntityProcGroup, EntityImportName, EntityLibraryName, EntityBuiltin:
			// OK
		default:
			continue
		}

		if e.Pkg != pkg {
			continue
		}

		if !IsEntityExported(e, true) {
			continue
		}

		if e.Token.String.Len() == 0 {
			continue
		}

		entry := OdinDocScopeEntry{
			Name:   w.writeString(elem.Name.String()),
			Entity: w.addEntity(e),
		}
		entries = append(entries, entry)
	}

	return w.writeScopeEntrySlice(entries)
}

// updateEntities fills in deferred entity info (types, foreign libraries, grouped entities).
func (w *Writer) updateEntities() {
	// First pass: write types for all entities (during preparing phase,
	// this ensures types are cached).
	var entities []*Entity
	for e := range w.entityCache {
		entities = append(entities, e)
	}

	for _, e := range entities {
		w.docType(e.Type, true)
	}

	// Second pass: update each entity with type index, foreign library, grouped entities
	for e, entityIndex := range w.entityCache {
		typeIndex := w.docType(e.Type, true)
		var foreignLibrary OdinDocEntityIndex
		var groupedEntities OdinDocArray

		switch e.Kind {
		case EntityVariable:
			foreignLibrary = w.addEntity(e.Variable.ForeignLibrary)
		case EntityProcedure:
			foreignLibrary = w.addEntity(e.Procedure.ForeignLibrary)
		case EntityProcGroup:
			indices := make([]uint32, 0, len(e.ProcGroup.Entities))
			for _, entity := range e.ProcGroup.Entities {
				idx := w.addEntity(entity)
				indices = append(indices, idx)
			}
			groupedEntities = w.writeUint32Slice(indices)
		}

		_ = foreignLibrary

		if w.state == WriterStateWriting {
			dst := w.getEntity(entityIndex)
			if dst != nil {
				dst.Type = typeIndex
				dst.ForeignLibrary = foreignLibrary
				dst.GroupedEntities = groupedEntities
			}
		}
	}
}
