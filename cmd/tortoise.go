// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// tortoise: a tool for displaying, interpreting, and debugging Go programs.
package main // import "github.com/rocky/ssa-interp/cmd"

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/rocky/go-types"
	"github.com/rocky/go-importer"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/gub/cmd"
)

var buildFlag = flag.String("build", "", `Options controlling the SSA builder.
The value is a sequence of zero or more of these letters:
C	perform sanity [C]hecking of the SSA form.
D	include [D]ebug info for every function.
P	log [P]ackage inventory.
F	log [F]unction SSA code.
S	log [S]ource locations as SSA builder progresses.
G	use binary object files from gc to provide imports (no code).
L	build distinct packages seria[L]ly instead of in parallel.
N	build [N]aive SSA form: don't replace local loads/stores with registers.
`)

var testFlag = flag.Bool("test", false, "Loads test code (*_test.go) for imported packages.")

var runFlag = flag.Bool("run", false, "Invokes the SSA interpreter on the program.")

var interpFlag = flag.String("interp", "", `Options controlling the interpreter.
The value is a sequence of zero or more more of these letters:
R	disable [R]ecover() from panic; show interpreter crash instead.
T	[T]race execution of the program.  Best for single-threaded programs!
I	trace [I]int() functions before main.main()
S	[S]atement tracing
`)

var gubFlag = flag.String("gub", "", `Options passed to the gub debugger.
`)

const usage = `SSA builder and interpreter.
Usage: tortoise [<flag> ...] [<file.go> ...] [<arg> ...]
       tortoise [<flag> ...] <import/path>   [<arg> ...]
Use -help flag to display options.

Examples:
% tortoise -run -interp=S hello.go     # interpret a program, with statement tracing
% tortoise -build=FPG hello.go         # quickly dump SSA form of a single package
% tortoise -run unicode -- -test.v     # interpret the unicode package's tests, verbosely

` + importer.InitialPackagesUsage +
	`
When -run is specified, tortoise will find the first package that
defines a main function and run it in the interpreter.
If none is found, the tests of each package will be run instead.
`

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func init() {
	// If $GOMAXPROCS isn't set, use the full capacity of the machine.
	// For small machines, use at least 4 threads.
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		if n < 4 {
			n = 4
		}
		runtime.GOMAXPROCS(n)
	}
}

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "tortoise: %s\n", err)
		os.Exit(1)
	}
}

func doMain() error {
	flag.Parse()
	args := flag.Args()

	conf := importer.Config{Build: &build.Default}
	// TODO(adonovan): make go/types choose its default Sizes from
	// build.Default or a specified *build.Context.
	var wordSize int64 = 8
	switch conf.Build.GOARCH {
	case "386", "arm":
		wordSize = 4
	}
	conf.TypeChecker.Sizes = &types.StdSizes{
		MaxAlign: 8,
		WordSize: wordSize,
	}

	var mode ssa2.BuilderMode = ssa2.NaiveForm

	for _, c := range *buildFlag {
		switch c {
		case 'D':
			mode |= ssa2.DebugInfo
		case 'P':
			mode |= ssa2.LogPackages | ssa2.BuildSerially
		case 'F':
			mode |= ssa2.LogFunctions | ssa2.BuildSerially
		case 'S':
			mode |= ssa2.LogSource | ssa2.BuildSerially
		case 'C':
			mode |= ssa2.SanityCheckFunctions
		case 'G':
			conf.Build = nil
		case 'L':
			mode |= ssa2.BuildSerially
		default:
			return fmt.Errorf("unknown -build option: '%c'", c)
		}
	}

	var interpMode interp.Mode
	var interpTraceMode interp.TraceMode
	for _, c := range *interpFlag {
		switch c {
		case 'I':
			interpTraceMode |= interp.EnableInitTracing
		case 'R':
			interpMode |= interp.DisableRecover
		case 'S':
			interpTraceMode |= interp.EnableStmtTracing
			mode |= ssa2.DebugInfo
		case 'T':
			interpTraceMode |= interp.EnableTracing
		default:
			return fmt.Errorf("unknown -interp option: '%c'", c)
		}
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	// Profiling support.
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Load, parse and type-check the program.
	imp := importer.New(&conf)
	prog_args := args[1:]
	infos, args, err := imp.LoadInitialPackages(args[0:1])
	if err != nil {
		return err
	}

	// The interpreter needs the runtime package.
	if *runFlag {
		if _, err := imp.LoadPackage("runtime"); err != nil {
			log.Fatalf("LoadPackage(runtime) failed: %s", err)
		}
	}

	// Create and build SSA-form program representation.
	prog := ssa2.NewProgram(imp.Fset, mode)
	if err := prog.CreatePackages(imp); err != nil {
		log.Fatal(err)
	}

	// Create and build SSA-form program representation.
	prog.BuildAll()

	// Run the interpreter.
	if *runFlag {
		// If some package defines main, run that.
		// Otherwise run all package's tests.
		var main *ssa2.Package
		var pkgs []*ssa2.Package
		for _, info := range infos {
			pkg := prog.Package(info.Pkg)
			if pkg.Func("main") != nil {
				main = pkg
				break
			}
			pkgs = append(pkgs, pkg)
		}
		if main == nil && pkgs != nil {
			main = prog.CreateTestMainPackage(pkgs...)
		}
		if main == nil {
			log.Fatal("No main package and no tests")
		}

		if interpTraceMode & interp.EnableStmtTracing != 0 {
			gubcmd.Init()
			gub.Install(gubFlag)
		}

		fmt.Println("Running....")
		if runtime.GOARCH != conf.Build.GOARCH {
			return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
				build.Default.GOARCH, runtime.GOARCH)
		}

		interp.Interpret(main, interpMode, interpTraceMode, conf.TypeChecker.Sizes, main.Object.Path(), prog_args)
	} else {
		fmt.Println(`Built ok, but not running because "-run" option not given`)
	}
	return nil
}
