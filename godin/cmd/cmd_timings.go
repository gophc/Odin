package cmd

import (
	"fmt"
	"os"
	"sort"
)

func timings_export_all(t *Timings, c *Checker, timings_are_finalized bool) {
	if !timings_are_finalized {
		timings__stop_current_section(t)
		t.Total.Finish = time_stamp_time_now()
	}

	unit := TimingUnit_Millisecond

	fileName := build_context.export_timings_file
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export timings to: %s\n", fileName)
		exit_with_errors()
		return
	} else {
		fmt.Printf("\nExporting timings to '%s'... ", fileName)
	}
	defer f.Close()

	if build_context.export_timings_format == TimingsExportJson {
		p := c.Parser
		lines := p.TotalLineCount
		tokens := p.TotalTokenCount
		files := 0
		packages := len(p.Packages)
		totalFileSize := 0
		for _, pkg := range p.Packages {
			files += len(pkg.Files)
			for _, file := range pkg.Files {
				totalFileSize += file.Tokenizer.End - file.Tokenizer.Start
			}
		}

		fmt.Fprintf(f, "{\n")
		fmt.Fprintf(f, "\t\"totals\": [\n")

		fmt.Fprintf(f, "\t\t{\"name\": \"total_packages\",  \"count\": %d},\n", packages)
		fmt.Fprintf(f, "\t\t{\"name\": \"total_files\",     \"count\": %d},\n", files)
		fmt.Fprintf(f, "\t\t{\"name\": \"total_lines\",     \"count\": %d},\n", lines)
		fmt.Fprintf(f, "\t\t{\"name\": \"total_tokens\",    \"count\": %d},\n", tokens)
		fmt.Fprintf(f, "\t\t{\"name\": \"total_file_size\", \"count\": %d},\n", totalFileSize)

		fmt.Fprintf(f, "\t],\n")

		fmt.Fprintf(f, "\t\"timings\": [\n")

		t.TotalTimeSeconds = time_stamp_as_s(t.Total, t.Freq)
		totalTime := time_stamp(t.Total, t.Freq, unit)

		fmt.Fprintf(f, "\t\t{\"name\": \"%s\", \"millis\": %.3f},\n",
			t.Total.Label, totalTime)

		for _, ts := range t.Sections {
			sectionTime := time_stamp(ts, t.Freq, unit)
			fmt.Fprintf(f, "\t\t{\"name\": \"%s\", \"millis\": %.3f},\n",
				ts.Label, sectionTime)
		}

		fmt.Fprintf(f, "\t],\n")

		fmt.Fprintf(f, "}\n")
	} else if build_context.export_timings_format == TimingsExportCSV {
		t.TotalTimeSeconds = time_stamp_as_s(t.Total, t.Freq)
		totalTime := time_stamp(t.Total, t.Freq, unit)

		fmt.Fprintf(f, "\"%s\", %d\n",
			t.Total.Label, int(totalTime))

		for _, ts := range t.Sections {
			sectionTime := time_stamp(ts, t.Freq, unit)
			fmt.Fprintf(f, "\"%s\", %d\n",
				ts.Label, int(sectionTime))
		}
	}

	fmt.Printf("Done.\n")
}

func show_timings(c *Checker, t *Timings) {
	p := c.Parser
	lines := p.TotalLineCount
	tokens := p.TotalTokenCount
	files := 0
	packages := len(p.Packages)
	totalFileSize := 0
	var totalTokenizingTime float64
	var totalParsingTime float64
	for _, pkg := range p.Packages {
		files += len(pkg.Files)
		for _, file := range pkg.Files {
			totalTokenizingTime += file.TimeToTokenize
			totalParsingTime += file.TimeToParse
			totalFileSize += file.Tokenizer.End - file.Tokenizer.Start
		}
	}

	timings_print_all(t)

	PRINT_PEAK_USAGE()

	if build_context.export_timings_format != TimingsExportUnspecified {
		timings_export_all(t, c, true)
	}

	if build_context.show_debug_messages && build_context.show_more_timings {
		{
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "Total Lines     - %d\n", lines)
			fmt.Fprintf(os.Stderr, "Total Tokens    - %d\n", tokens)
			fmt.Fprintf(os.Stderr, "Total Files     - %d\n", files)
			fmt.Fprintf(os.Stderr, "Total Packages  - %d\n", packages)
			fmt.Fprintf(os.Stderr, "Total File Size - %d\n", totalFileSize)
			fmt.Fprintf(os.Stderr, "\n")
		}
		{
			time := totalTokenizingTime
			fmt.Fprintf(os.Stderr, "Tokenization Only\n")
			fmt.Fprintf(os.Stderr, "LOC/s        - %.3f\n", float64(lines)/time)
			fmt.Fprintf(os.Stderr, "us/LOC       - %.3f\n", 1.0e6*time/float64(lines))
			fmt.Fprintf(os.Stderr, "Tokens/s     - %.3f\n", float64(tokens)/time)
			fmt.Fprintf(os.Stderr, "us/Token     - %.3f\n", 1.0e6*time/float64(tokens))
			fmt.Fprintf(os.Stderr, "bytes/s      - %.3f\n", float64(totalFileSize)/time)
			fmt.Fprintf(os.Stderr, "MiB/s        - %.3f\n", float64(totalFileSize/time)/(1024*1024))
			fmt.Fprintf(os.Stderr, "us/bytes     - %.3f\n", 1.0e6*time/float64(totalFileSize))

			fmt.Fprintf(os.Stderr, "\n")
		}
		{
			time := totalParsingTime
			fmt.Fprintf(os.Stderr, "Parsing Only\n")
			fmt.Fprintf(os.Stderr, "LOC/s        - %.3f\n", float64(lines)/time)
			fmt.Fprintf(os.Stderr, "us/LOC       - %.3f\n", 1.0e6*time/float64(lines))
			fmt.Fprintf(os.Stderr, "Tokens/s     - %.3f\n", float64(tokens)/time)
			fmt.Fprintf(os.Stderr, "us/Token     - %.3f\n", 1.0e6*time/float64(tokens))
			fmt.Fprintf(os.Stderr, "bytes/s      - %.3f\n", float64(totalFileSize)/time)
			fmt.Fprintf(os.Stderr, "MiB/s        - %.3f\n", float64(totalFileSize/time)/(1024*1024))
			fmt.Fprintf(os.Stderr, "us/bytes     - %.3f\n", 1.0e6*time/float64(totalFileSize))

			fmt.Fprintf(os.Stderr, "\n")
		}
		{
			var ts TimeStamp
			for _, s := range t.Sections {
				label := s.Label
				if label == "parse files" {
					ts = s
					break
				}
			}
			parseTime := time_stamp_as_s(ts, t.Freq)
			fmt.Fprintf(os.Stderr, "Parse pass\n")
			fmt.Fprintf(os.Stderr, "LOC/s        - %.3f\n", float64(lines)/parseTime)
			fmt.Fprintf(os.Stderr, "us/LOC       - %.3f\n", 1.0e6*parseTime/float64(lines))
			fmt.Fprintf(os.Stderr, "Tokens/s     - %.3f\n", float64(tokens)/parseTime)
			fmt.Fprintf(os.Stderr, "us/Token     - %.3f\n", 1.0e6*parseTime/float64(tokens))
			fmt.Fprintf(os.Stderr, "bytes/s      - %.3f\n", float64(totalFileSize)/parseTime)
			fmt.Fprintf(os.Stderr, "MiB/s        - %.3f\n", float64(totalFileSize/parseTime)/(1024*1024))
			fmt.Fprintf(os.Stderr, "us/bytes     - %.3f\n", 1.0e6*parseTime/float64(totalFileSize))

			fmt.Fprintf(os.Stderr, "\n")
		}
		{
			var ts TimeStamp
			var tsEnd TimeStamp
			for _, s := range t.Sections {
				label := s.Label
				if label == "type check" {
					ts = s
				}
				if label == "type check finish" {
					tsEnd = s
					break
				}
			}
			ts.Finish = tsEnd.Finish

			parseTime := time_stamp_as_s(ts, t.Freq)
			fmt.Fprintf(os.Stderr, "Checker pass\n")
			fmt.Fprintf(os.Stderr, "LOC/s        - %.3f\n", float64(lines)/parseTime)
			fmt.Fprintf(os.Stderr, "us/LOC       - %.3f\n", 1.0e6*parseTime/float64(lines))
			fmt.Fprintf(os.Stderr, "Tokens/s     - %.3f\n", float64(tokens)/parseTime)
			fmt.Fprintf(os.Stderr, "us/Token     - %.3f\n", 1.0e6*parseTime/float64(tokens))
			fmt.Fprintf(os.Stderr, "bytes/s      - %.3f\n", float64(totalFileSize)/parseTime)
			fmt.Fprintf(os.Stderr, "MiB/s        - %.3f\n", float64(totalFileSize/parseTime)/(1024*1024))
			fmt.Fprintf(os.Stderr, "us/bytes     - %.3f\n", 1.0e6*parseTime/float64(totalFileSize))
			fmt.Fprintf(os.Stderr, "\n")
		}
		{
			totalTime := t.TotalTimeSeconds
			fmt.Fprintf(os.Stderr, "Total pass\n")
			fmt.Fprintf(os.Stderr, "LOC/s        - %.3f\n", float64(lines)/totalTime)
			fmt.Fprintf(os.Stderr, "us/LOC       - %.3f\n", 1.0e6*totalTime/float64(lines))
			fmt.Fprintf(os.Stderr, "Tokens/s     - %.3f\n", float64(tokens)/totalTime)
			fmt.Fprintf(os.Stderr, "us/Token     - %.3f\n", 1.0e6*totalTime/float64(tokens))
			fmt.Fprintf(os.Stderr, "bytes/s      - %.3f\n", float64(totalFileSize)/totalTime)
			fmt.Fprintf(os.Stderr, "MiB/s        - %.3f\n", float64(totalFileSize/totalTime)/(1024*1024))
			fmt.Fprintf(os.Stderr, "us/bytes     - %.3f\n", 1.0e6*totalTime/float64(totalFileSize))
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
}

func show_import_graph(c *Checker) {
	p := c.Parser

	fmt.Printf("digraph odin_import_graph {\n\tnode [shape=box];\n")

	clusterCounter := 0
	for _, coll := range library_collections {
		fmt.Printf("\tsubgraph cluster_%d {\n", clusterCounter)
		fmt.Printf("\t\tlabel = \"%s\";\n", coll.Name)
		fmt.Printf("\t\tnode [style=filled, fillcolor=white];\n")
		if coll.Name == "core" {
			fmt.Printf("\t\tbgcolor = lightsalmon;\n")
		} else if coll.Name == "vendor" {
			fmt.Printf("\t\tbgcolor = lightblue;\n")
		} else if coll.Name == "base" {
			fmt.Printf("\t\tbgcolor = lightcoral;\n")
			fmt.Printf("\t\tintrinsics;\n")
			fmt.Printf("\t\tbuiltin;\n")
		}
		for _, pkg := range p.Packages {
			if string_starts_with(pkg.Fullpath, coll.Path) {
				fmt.Printf("\t\t\"%s\";\n", pkg.Fullpath)
			}
		}
		fmt.Printf("\t}\n")
		clusterCounter++
	}

	for _, pkg := range p.Packages {
		for i := 0; i < len(pkg.Files); i++ {
			file := pkg.Files[i]
			for _, imp := range file.Imports {
				exists := false
				for j := i + 1; j < len(pkg.Files); j++ {
					otherFile := pkg.Files[j]
					for _, otherImp := range otherFile.Imports {
						if string_compare(imp.ImportDecl.Fullpath, otherImp.ImportDecl.Fullpath) == 0 {
							exists = true
							break
						}
					}
					if exists {
						break
					}
				}

				if exists {
					continue
				}

				path := imp.ImportDecl.Fullpath
				if imp.ImportDecl.Package != nil {
					path = imp.ImportDecl.Package.Fullpath
				}

				fmt.Printf("\t\"%s\" -> \"%s\";\n", pkg.Fullpath, path)
			}
		}
	}

	fmt.Printf("}\n\n")
}

func file_path_cmp(a, b *AstFile) int {
	if a.Fullpath < b.Fullpath {
		return -1
	}
	if a.Fullpath > b.Fullpath {
		return 1
	}
	return 0
}

func export_dependencies(c *Checker) {
	if len(build_context.export_dependencies_file) <= 0 {
		fmt.Fprintf(os.Stderr, "No dependency file specified with `-export-dependencies-file`\n")
		exit_with_errors()
		return
	}

	p := c.Parser

	fileName := build_context.export_dependencies_file
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export dependencies to: %s\n", fileName)
		exit_with_errors()
		return
	}
	defer f.Close()

	files := make([]*AstFile, 0)
	for _, pkg := range p.Packages {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Fullpath < files[j].Fullpath
	})

	loadFiles := make([]*LoadFileCache, 0)
	for _, cache := range c.Info.LoadFileCache {
		if cache == nil || !cache.Exists {
			continue
		}
		loadFiles = append(loadFiles, cache)
	}
	sort.Slice(loadFiles, func(i, j int) bool {
		return file_cache_sort_cmp(loadFiles[i], loadFiles[j]) < 0
	})

	if build_context.export_dependencies_format == DependenciesExportMake {
		exeName := build_context.build_paths[BuildPath_Output]

		fmt.Fprintf(f, "%s:", exeName)

		currentLineLength := len(exeName) + 1

		for _, file := range files {
			if currentLineLength >= 80-2 {
				f.Write([]byte(" \\\n "))
				currentLineLength = 1
			}

			f.Write([]byte(" "))
			currentLineLength++

			for k := 0; k < len(file.Fullpath); k++ {
				part := file.Fullpath[k]
				if part == ' ' {
					f.Write([]byte("\\"))
					currentLineLength++
				}
				f.Write([]byte{part})
				currentLineLength++
			}
		}

		fmt.Fprintf(f, "\n")
	} else if build_context.export_dependencies_format == DependenciesExportJson {
		fmt.Fprintf(f, "{\n")

		fmt.Fprintf(f, "\t\"source_files\": [\n")

		for i, file := range files {
			fmt.Fprintf(f, "\t\t\"%s\"", file.Fullpath)
			if i+1 < len(files) {
				fmt.Fprintf(f, ",")
			}
			fmt.Fprintf(f, "\n")
		}

		fmt.Fprintf(f, "\t],\n")

		fmt.Fprintf(f, "\t\"load_files\": [\n")

		for i, cache := range loadFiles {
			fmt.Fprintf(f, "\t\t\"%s\"", cache.Path)
			if i+1 < len(loadFiles) {
				fmt.Fprintf(f, ",")
			}
			fmt.Fprintf(f, "\n")
		}

		fmt.Fprintf(f, "\t]\n")

		fmt.Fprintf(f, "}\n")
	}
}

func export_linked_libraries(gen *LinkerData) {
	fileName := build_context.export_linked_libs_path
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to export linked library list to: %s\n", fileName)
		exit_with_errors()
		return
	}
	defer f.Close()

	var minLibsSet StringSet
	string_set_init(&minLibsSet, 64)
	defer string_set_destroy(&minLibsSet)

	for _, e := range gen.ForeignLibraries {
		imp := e.LibraryName.Decl.(*AstForeignImportDecl)

		for i := 0; i < len(e.LibraryName.Paths); i++ {
			libPath := string_trim_whitespace(e.LibraryName.Paths[i])
			if len(libPath) == 0 {
				continue
			}

			if string_set_update(&minLibsSet, libPath) {
				continue
			}

			fmt.Fprintf(f, "%s\t", libPath)

			ext := path_extension(libPath, false)
			extA := "a"
			extLib := "lib"
			extO := "o"
			extObj := "obj"
			if str_eq_ignore_case(ext, extA) || str_eq_ignore_case(ext, extLib) ||
				str_eq_ignore_case(ext, extO) || str_eq_ignore_case(ext, extObj) {
				fmt.Fprintf(f, "static")
			} else {
				fmt.Fprintf(f, "dynamic")
			}

			fmt.Fprintf(f, "\t")

			filePath := imp.Filepaths[i]
			filePathStr := filePath.Tav.Value.ValueString
			systemPrefix := "system:"

			if string_starts_with(filePathStr, systemPrefix) {
				fmt.Fprintf(f, "system")
			} else {
				fmt.Fprintf(f, "user")
			}

			fmt.Fprintf(f, "\n")
		}
	}
}

func remove_temp_files(gen *lbGenerator) {
	if build_context.keep_temp_files {
		return
	}

	switch build_context.build_mode {
	case BuildMode_Executable, BuildMode_StaticLibrary, BuildMode_DynamicLibrary:
		break
	case BuildMode_Object, BuildMode_Assembly, BuildMode_LLVM_IR:
		return
	}

	timings_start_section(&global_timings, "remove temp files")
	defer timings__stop_current_section(&global_timings)

	for _, path := range gen.OutputTempPaths {
		os.Remove(string(path))
	}

	if !build_context.keep_object_files {
		switch build_context.build_mode {
		case BuildMode_Executable, BuildMode_StaticLibrary, BuildMode_DynamicLibrary:
			for _, path := range gen.OutputObjectPaths {
				os.Remove(string(path))
			}
		}
	}
}
