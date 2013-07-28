package ssa2_test

import (
	"code.google.com/p/go.tools/go/types"
	"code.google.com/p/go.tools/importer"
	"github.com/rocky/ssa-interp"
	"go/ast"
	"go/parser"
	"strings"
	"testing"
)

func isEmpty(f *ssa2.Function) bool { return f.Blocks == nil }

// Tests that programs partially loaded from gc object files contain
// functions with no code for the external portions, but are otherwise ok.
func TestExternalPackages(t *testing.T) {
	test := `
package main

import (
	"bytes"
	"io"
	"testing"
)

func main() {
        var t testing.T
	t.Parallel()    // static call to external declared method
        t.Fail()        // static call to promoted external declared method
        testing.Short() // static call to external package-level function

        var w io.Writer = new(bytes.Buffer)
        w.Write(nil)    // interface invoke of external declared method
}
`
	imp := importer.New(new(importer.Config)) // no Loader; uses GC importer

	f, err := parser.ParseFile(imp.Fset, "<input>", test, parser.DeclarationErrors)
	if err != nil {
		t.Errorf("parse error: %s", err)
		return
	}

	info := imp.CreateSourcePackage("main", []*ast.File{f})
	if info.Err != nil {
		t.Error(info.Err.Error())
		return
	}

	prog := ssa2.NewProgram(imp.Fset, ssa2.SanityCheckFunctions)
	for _, info := range imp.Packages {
		prog.CreatePackage(info)
	}
	mainPkg := prog.Package(info.Pkg)
	mainPkg.Build()

	// Only the main package and its immediate dependencies are loaded.
	deps := []string{"bytes", "io", "testing"}
	if len(prog.PackagesByPath) != 1+len(deps) {
		t.Errorf("unexpected set of loaded packages: %q", prog.PackagesByPath)
	}
	for _, path := range deps {
		pkg, _ := prog.PackagesByPath[path]
		if pkg == nil {
			t.Errorf("package not loaded: %q", path)
			continue
		}

		// External packages should have no function bodies (except for wrappers).
		isExt := pkg != mainPkg

		// init()
		if isExt && !isEmpty(pkg.Func("init")) {
			t.Errorf("external package %s has non-empty init", pkg)
		} else if !isExt && isEmpty(pkg.Func("init")) {
			t.Errorf("main package %s has empty init", pkg)
		}

		for _, mem := range pkg.Members {
			switch mem := mem.(type) {
			case *ssa2.Function:
				// Functions at package level.
				if isExt && !isEmpty(mem) {
					t.Errorf("external function %s is non-empty", mem)
				} else if !isExt && isEmpty(mem) {
					t.Errorf("function %s is empty", mem)
				}

			case *ssa2.Type:
				// Methods of named types T.
				// (In this test, all exported methods belong to *T not T.)
				if !isExt {
					t.Fatalf("unexpected name type in main package: %s", mem)
				}
				for _, m := range prog.MethodSet(types.NewPointer(mem.Type())) {
					// For external types, only synthetic wrappers have code.
					expExt := !strings.Contains(m.Synthetic, "wrapper")
					if expExt && !isEmpty(m) {
						t.Errorf("external method %s is non-empty: %s",
							m, m.Synthetic)
					} else if !expExt && isEmpty(m) {
						t.Errorf("method function %s is empty: %s",
							m, m.Synthetic)
					}
				}
			}
		}
	}

	expectedCallee := []string{
		"(*testing.T).Parallel",
		"(*testing.common).Fail",
		"testing.Short",
		"N/A",
	}
	callNum := 0
	for _, b := range mainPkg.Func("main").Blocks {
		for _, instr := range b.Instrs {
			switch instr := instr.(type) {
			case ssa2.CallInstruction:
				call := instr.Common()
				if want := expectedCallee[callNum]; want != "N/A" {
					got := call.StaticCallee().String()
					if want != got {
						t.Errorf("call #%d from main.main: got callee %s, want %s",
							callNum, got, want)
					}
				}
				callNum++
			}
		}
	}
	if callNum != 4 {
		t.Errorf("in main.main: got %d calls, want %d", callNum, 4)
	}
}
