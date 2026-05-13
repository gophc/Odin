package cmd

import (
	"fmt"
	"io"
	"os"
)

func write_file_with_stripped_tokens(w io.Writer, file *AstFile) (int64, error) {
	fileData := file.Tokenizer.Data
	endOffset := len(fileData)

	var written int64
	prevOffset := 0

	for _, token := range file.Tokens {
		if token.Flags&(TokenFlag_Remove|TokenFlag_Replace) != 0 {
			offset := token.Pos.Offset
			toWrite := offset - prevOffset
			if toWrite > 0 {
				n, err := w.Write(fileData[prevOffset:offset])
				written += int64(n)
				if err != nil {
					return written, err
				}
			}
			prevOffset = tokenPosEnd(token).Offset
		}
		if token.Flags&TokenFlag_Replace != 0 {
			switch token.Kind {
			case Token_Ellipsis:
				n, err := w.Write([]byte("..="))
				written += int64(n)
				if err != nil {
					return written, err
				}
			default:
				return written, fmt.Errorf("unexpected token kind for replacement")
			}
		}
	}

	if endOffset > prevOffset {
		n, err := w.Write(fileData[prevOffset:endOffset])
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func strip_semicolons(parser *Parser) int {
	fileCount := 0
	for _, pkg := range parser.Packages {
		fileCount += len(pkg.Files)
	}

	generatedFiles := make([]StripSemicolonFile, 0, fileCount)

	for _, pkg := range parser.Packages {
		for _, file := range pkg.Files {
			nothingToChange := true
			for _, token := range file.Tokens {
				if token.Flags != 0 {
					nothingToChange = false
					break
				}
			}
			if nothingToChange {
				continue
			}

			oldFullpath := file.Fullpath
			fullpathBase := oldFullpath[:len(oldFullpath)-5]
			oldFullpathBackup := fullpathBase + "~backup.odin-temp"
			newFullpath := fullpathBase + "~temp.odin-temp"

			generatedFiles = append(generatedFiles, StripSemicolonFile{
				OldFullpath:       oldFullpath,
				OldFullpathBackup: oldFullpathBackup,
				NewFullpath:       newFullpath,
				File:              file,
			})
		}
	}

	fmt.Fprintf(os.Stderr, "File count to be stripped of unneeded tokens: %d\n", len(generatedFiles))

	generatedCount := 0
	failed := false

	for i := range generatedFiles {
		sf := &generatedFiles[i]
		filename := sf.NewFullpath

		f, err := os.Create(filename)
		if err != nil {
			failed = true
			break
		}

		generatedCount++

		debugf("Write file with stripped tokens: %s\n", filename)

		written, err := write_file_with_stripped_tokens(f, sf.File)
		if err != nil {
			f.Close()
			failed = true
			break
		}

		sf.Written = written
		f.Close()
	}

	if failed {
		for _, sf := range generatedFiles[:generatedCount] {
			os.Remove(sf.NewFullpath)
		}
		return 1
	}

	overwrittenFiles := 0
	for _, sf := range generatedFiles {
		err := os.Rename(sf.OldFullpath, sf.OldFullpathBackup)
		if err != nil {
			failed = true
			break
		}
		err = os.Rename(sf.NewFullpath, sf.OldFullpath)
		if err != nil {
			os.Rename(sf.OldFullpathBackup, sf.OldFullpath)
			failed = true
			break
		}
		os.Remove(sf.OldFullpathBackup)
		overwrittenFiles++
	}

	if !buildContext.KeepTempFiles {
		for _, sf := range generatedFiles {
			os.Remove(sf.NewFullpath)
			os.Remove(sf.OldFullpathBackup)
		}
	}

	fmt.Fprintf(os.Stderr, "Files stripped of unneeded token: %d\n", len(generatedFiles))

	if failed {
		return 1
	}
	return 0
}
