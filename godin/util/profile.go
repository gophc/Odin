package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type testingT interface {
	Name() string
	Skipf(format string, args ...any)
}

func SkipTestOnlyDbg(t testingT, args []string) bool {
	name := t.Name()
	if arg := strings.Join(args, " "); strings.Index(arg, `^\Q`+name+`\E$`) < 0 {
		t.Skipf("[SkipTestOnlyDbg] %s skip args: %s", name, arg)
		return true
	}
	return false
}

func ParseFlagsMMLine(line string) (key string, name string, ln int, col int, err error) {
	arr := strings.SplitN(line, ":", 4)
	if len(arr) != 4 {
		err = fmt.Errorf("err FlagsMMLine: %s", line)
		return
	}
	ln, err = strconv.Atoi(arr[1])
	if err != nil {
		return
	}
	col, err = strconv.Atoi(arr[2])
	if err != nil {
		return
	}
	name = arr[0]
	key = arr[0] + ":" + arr[1]
	return
}

func (i *InfoBuildFlagsMM) GetDeclFromFlagsMMLine(line string) (key string, fn *ast.FuncDecl, ln, col int) {
	var (
		err error
		src []byte
	)
	name, ok := "", false

	key, name, ln, col, err = ParseFlagsMMLine(line)
	if err != nil {
		return
	}

	if fn, ok = i.fkeyDeclCache[key]; ok {
		return
	}

	filename := filepath.Join(i.cwd, name)
	f, ok := i.fileAstCache[filename]
	if !ok {
		src, err = os.ReadFile(filename)
		if err != nil {
			return
		}
		f, err = parser.ParseFile(i.fset, filename, src, 0)
		if err != nil {
			return
		}
		i.fileAstCache[filename] = f
	}

	// ast.Print(fset, f)
	for _, d := range f.Decls {
		fn, ok = d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		pos := i.fset.Position(fn.Name.Pos())
		if pos.Line == ln {
			i.fkeyDeclCache[key] = fn
			return
		}
		stmt := fn.Body
		if stmt != nil && int(stmt.Rbrace) > 0 {
			pos2 := i.fset.Position(stmt.Rbrace)
			if ln >= pos.Line && ln <= pos2.Line {
				i.fkeyDeclCache[key] = fn
				return
			}
		}
	}
	return
}

func (i *InfoBuildFlagsMM) MatchCallStmt(fn *ast.FuncDecl, key string, ln, col int, name string) (string, bool) {
	pkey := key + ":" + strconv.Itoa(col)
	calls, ok := i.pkeyCallNameCache[pkey]
	if !ok {
		pos, stmt := i.fset.Position(fn.Name.Pos()), fn.Body
		var calla *ast.CallExpr
		if stmt != nil && int(stmt.Rbrace) > 0 {
			pos2 := i.fset.Position(stmt.Rbrace)
			if ln >= pos.Line && ln <= pos2.Line {
				ast.Inspect(stmt, func(n ast.Node) bool {
					if t, ok := n.(*ast.CallExpr); ok {
						pos3 := i.fset.Position(n.Pos())
						if pos3.Line == ln && pos3.Column <= col && calla == nil {
							calla = t
						}
					}
					return true
				})
			}
		}

		if calla != nil {
			if c, ok := (calla.Fun).(*ast.Ident); ok {
				calls = c.Name
			}
		}
		i.pkeyCallNameCache[pkey] = calls
	}

	if calls == name {
		return calls, true
	}
	return calls, false
}

func (i *InfoBuildFlagsMM) MatchFnDecl(fn *ast.FuncDecl, key, name string) (string, bool) {
	fns, ok := i.lkeyFnNameCache[key]
	if !ok {
		fns = fn.Name.Name
		tpy := ""
		if fn.Recv != nil && len(fn.Recv.List) == 1 {
			t := fn.Recv.List[0].Type
			if tt, ok := t.(*ast.Ident); ok {
				tpy = tt.Name
			}
			if tt, ok := t.(*ast.StarExpr); ok {
				if ttt, ok := (tt.X).(*ast.Ident); ok {
					tpy = "*" + ttt.Name
				}
			}
		}
		if tpy != "" {
			fns = tpy + "." + fns
		}
		i.lkeyFnNameCache[key] = fns
	}

	if fns == name {
		return fns, true
	}
	return fns, false
}

type InlineBuildFlagsMM struct {
	cost    int
	budget  int
	line    string
	name    string
	codeAs  string
	filepos string
}

type InfoBuildFlagsMM struct {
	cwd  string
	fset *token.FileSet

	fileAstCache      map[string]*ast.File
	fkeyDeclCache     map[string]*ast.FuncDecl
	lkeyFnNameCache   map[string]string
	pkeyCallNameCache map[string]string

	Inline map[string]*InlineBuildFlagsMM

	Outline map[string]*InlineBuildFlagsMM

	Inlines  []string
	Outlines []string
	Others   []string

	EscapePtr map[string][]string

	EsErrorString   []string
	EsStorageHeap   []string
	EsStorageSpread []string
}

func PrepareGoBuildFlagsMM(cwd string, errStr string, skips ...string) (info *InfoBuildFlagsMM) {
	info = &InfoBuildFlagsMM{
		cwd:       cwd,
		Inline:    make(map[string]*InlineBuildFlagsMM),
		Outline:   make(map[string]*InlineBuildFlagsMM),
		fset:      token.NewFileSet(),
		EscapePtr: make(map[string][]string),

		fileAstCache:      make(map[string]*ast.File),
		fkeyDeclCache:     make(map[string]*ast.FuncDecl),
		lkeyFnNameCache:   make(map[string]string),
		pkeyCallNameCache: make(map[string]string),
	}
	lines := strings.Split(errStr, "\n")
	if len(lines) <= 0 {
		return
	}

	inline := regexp.MustCompile(`^([-\w.:\\<>]+)\s+can\s+inline\s+([-/\w()*.\[\]]+)\s+with\s+cost\s+(\d+) as:(.*)$`)
	outline := regexp.MustCompile(`^([-\w.:\\<>]+)\s+cannot\s+inline\s+([-/\w()*.\[\]]+):\s+function\s+too\s+complex:\s+cost\s+(\d+)\s+exceeds\s+budget\s+(\d+)$`)
	escape := regexp.MustCompile(`^([-\w.:\\<>]+)\s+([\w.{}&~*\[\]]+)\s+escapes\s+to\s+heap:?$`)

	heap := regexp.MustCompile(`^([-\w.:\\<>]+)\s+flow:\s+\{heap}\s+=\s+&(.*)$`)
	errtext := regexp.MustCompile(`^([-\w.:\\<>]+)\s+flow:\s+errors\.text\s+=\s+(.*)$`)
	storage := regexp.MustCompile(`^([-\w.:\\<>]+)\s+flow:\s+{storage\s+for\s+\.\.\.\s+argument}\s+=\s+(.*)$`)
	var ret []string
	pre, rule := "", ""

LoopNextLine:
	for _, line := range lines {
		if pre != "" && strings.HasPrefix(line, pre) {
			continue
		}
		pre = ""

		if strings.Index(line, " inlining call to ") >= 0 ||
			strings.Index(line, " does not escape") >= 0 {
			continue
		}

		rule = "storage"
		ret = storage.FindStringSubmatch(line)
		if len(ret) == 0 {
			rule = "errtext"
			ret = errtext.FindStringSubmatch(line)
		}
		if len(ret) == 0 {
			rule = "heap"
			ret = heap.FindStringSubmatch(line)
		}
		if len(ret) > 1 {
			if len(info.Others) > 0 && strings.HasPrefix(info.Others[len(info.Others)-1], pre) {
				if rule == "errtext" {
					info.EsErrorString = append(info.EsErrorString, info.Others[len(info.Others)-1])
				} else if rule == "heap" {
					info.EsStorageHeap = append(info.EsStorageHeap, info.Others[len(info.Others)-1])
				} else {
					info.EsStorageSpread = append(info.EsStorageSpread, info.Others[len(info.Others)-1])
				}
				info.Others = info.Others[0 : len(info.Others)-1]
			}
			pre = ret[1]
			continue
		}

		ret = inline.FindStringSubmatch(line)
		if len(ret) > 1 {
			key := ret[2]
			info.Inline[key] = &InlineBuildFlagsMM{
				line:    line,
				filepos: ret[1],
				name:    ret[2],
				cost:    MustAtoI(ret[3]),
				codeAs:  ret[4],
			}
			info.Inlines = append(info.Inlines, key)
			continue
		}

		ret = outline.FindStringSubmatch(line)
		if len(ret) > 1 {
			key := ret[2]
			info.Outline[key] = &InlineBuildFlagsMM{
				line:    line,
				filepos: ret[1],
				name:    ret[2],
				cost:    MustAtoI(ret[3]),
				budget:  MustAtoI(ret[4]),
			}
			info.Outlines = append(info.Outlines, key)
			continue
		}

		escapes := escape.FindStringSubmatch(line)
		if len(escapes) > 1 {
			if escapes[2] == "&errors.errString{...}" {
				info.EsErrorString = append(info.EsErrorString, line)
				pre = escapes[1]
				continue
			}
		}

		if line != "" {
			if key, fn, ln, col := info.GetDeclFromFlagsMMLine(line); fn != nil &&
				!strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "<") {
				for _, skip := range skips {
					if _, ok := info.MatchFnDecl(fn, key, skip); ok {
						pre = key
						continue LoopNextLine
					}
				}

				if _, ok := info.MatchCallStmt(fn, key, ln, col, "panic"); ok {
					continue LoopNextLine
				}

				if len(escapes) > 1 {
					fns, _ := info.MatchFnDecl(fn, key, "")
					pls, _ := info.EscapePtr[fns]
					if (len(pls) > 0 && pls[len(pls)-1] != escapes[2]) || pls == nil {
						info.EscapePtr[fns] = append(pls, escapes[2]+" # "+line)
					}
					pre = escapes[1]
					continue LoopNextLine
				}
			}

			info.Others = append(info.Others, line)
		}
	}
	return
}
