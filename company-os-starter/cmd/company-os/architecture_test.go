package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// bannedSelectors are the calls that would move process control or user-facing
// output below the dispatch seam. R-2.10: nothing under internal/ may exit the
// process or write to stdout — commands return records and errors, and only
// cmd/company-os turns those into bytes and an exit code.
//
// os.Stdout is included alongside the fmt.Print family for the same reason: a
// package that reaches for it has bypassed the io.Writer the renderer is handed.
// Writing to stderr is not banned; diagnostics are not the output contract.
var bannedSelectors = map[string]map[string]string{
	"os": {
		"Exit":   "return an error carrying a model.ExitCode; only main exits",
		"Stdout": "take an io.Writer from the caller",
	},
	"fmt": {
		"Print":   "return records; only cmd/company-os writes output",
		"Printf":  "return records; only cmd/company-os writes output",
		"Println": "return records; only cmd/company-os writes output",
	},
	"log": {
		"Fatal":   "return an error carrying a model.ExitCode; only main exits",
		"Fatalf":  "return an error carrying a model.ExitCode; only main exits",
		"Fatalln": "return an error carrying a model.ExitCode; only main exits",
		"Panic":   "return an error carrying a model.ExitCode; only main exits",
		"Panicf":  "return an error carrying a model.ExitCode; only main exits",
		"Panicln": "return an error carrying a model.ExitCode; only main exits",
		"Print":   "return records; only cmd/company-os writes output",
		"Printf":  "return records; only cmd/company-os writes output",
		"Println": "return records; only cmd/company-os writes output",
	},
}

// TestInternalPackagesDoNotExitOrPrint enforces the dispatch seam structurally,
// so a violation fails the build rather than being caught in review.
func TestInternalPackagesDoNotExitOrPrint(t *testing.T) {
	root := filepath.Join("..", "..", "internal")
	fset := token.NewFileSet()
	var violations []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Test files are excluded: the invariant constrains the library surface,
		// and a test binary exiting or printing harms nothing.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		imported := importNames(file)
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || !imported[pkg.Name] {
				return true
			}
			remedy, banned := bannedSelectors[pkg.Name][sel.Sel.Name]
			if !banned {
				return true
			}
			violations = append(violations, fmt.Sprintf(
				"%s: %s.%s is forbidden below cmd/ — %s",
				fset.Position(sel.Pos()), pkg.Name, sel.Sel.Name, remedy))
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	for _, v := range violations {
		t.Error(v)
	}
}

// importNames returns the local names under which the banned packages are
// imported by this file, so a local variable that happens to be called `log`
// does not trip the check.
func importNames(file *ast.File) map[string]bool {
	names := map[string]bool{}
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if _, watched := bannedSelectors[path]; !watched {
			continue
		}
		name := path
		if imp.Name != nil {
			name = imp.Name.Name
		}
		names[name] = true
	}
	return names
}
