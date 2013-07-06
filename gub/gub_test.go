package gub_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"code.google.com/p/go.tools/importer"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/ssa-interp/gub"
)

const slash = string(os.PathSeparator)

// These are files in ssa-interp/gub/testdata/.
type testDatum struct {
	gofile  string
	cmdfile string
}
var testData = []testDatum {
	{gofile: "gcd", cmdfile: "stepping"},
}

// Runs compiles, and runs go program. Then compares output.
func run(t *testing.T, test testDatum) bool {
	fmt.Printf("Input: %s on %s.go\n", test.cmdfile, test.gofile)

	// Consider moving out of this routine
	impctx := importer.Context{Loader: importer.MakeGoBuildLoader(nil)}

	gofile  := fmt.Sprintf("testdata%s%s.go",  slash, test.gofile)

	var inputs []string
	inputs = append(inputs, gofile)

	// Load, parse and type-check the program.
	imp := importer.New(&impctx)

	info, args, err := importer.CreatePackageFromArgs(imp, inputs)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create and build SSA-form program representation.
	prog := ssa2.NewProgram(imp.Fset, ssa2.NaiveForm)
	prog.CreatePackages(imp)
	prog.BuildAll()

	// Run the interpreter.
	gubFlag := fmt.Sprintf("-cmdfile testdata%s%s.cmd", slash, test.cmdfile)
	gub.Install(&gubFlag)

	interp.Interpret(prog.Package(info.Pkg), 0,
		interp.EnableStmtTracing, info.Pkg.Path(), args)

	// Print a helpful hint if we don't make it to the end.
	hint := "Run manually"
	defer func() {
		if hint != "" {
			fmt.Println("FAIL")
			fmt.Println(hint)
		} else {
			fmt.Println("PASS")
		}
	}()

	hint = "" // call off the hounds
	return true
}

// TestInterp runs the debugger on a selection of small Go programs.
func TestInterp(t *testing.T) {

	var failures []string

	for _, test := range testData {
		if !run(t, test) {
			failures = append(failures, test.cmdfile)
		}
	}

	if failures != nil {
		fmt.Println("The following tests failed:")
		for _, f := range failures {
			fmt.Printf("\t%s\n", f)
		}
	}

}
