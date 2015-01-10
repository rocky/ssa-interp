// Copyright 2013 The Go Authors. All rights reserved.
// Copyright 2015 Rocky Bernstein
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// tortoise: a tool for displaying, interpreting, and debugging Go programs.
package main // import "github.com/rocky/ssa-interp/cmd"

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/rocky/go-loader"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/go-types"
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
I	build bare [I]nit functions: no init guards or calls to dependent inits.
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
% tortoise -run -interp=S hello.go        # interpret a program, with statement tracing
% tortoise -build=FPG hello.go            # quickly dump SSA form of a single package
% tortoise -run -interp=T hello.go        # interpret a program, with tracing
% tortoise -run -test unicode -- -test.v  # interpret the unicode package's tests, verbosely
` + loader.FromArgsUsage +
	`
When -run is specified, tortoise will run the program.
The entry point depends on the -test flag:
if clear, it runs the first package named main.
if set, it runs the tests of each package.
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
	restart_args := os.Args
	flag.Parse()
	args := flag.Args()

<<<<<<< HEAD:cmd/tortoise.go
	conf := importer.Config{Build: &build.Default}
=======
	conf := loader.Config{
		Build:         &build.Default,
		SourceImports: true,
	}
>>>>>>> go1.2:cmd/tortoise.go
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
			mode |= ssa2.GlobalDebug
		case 'P':
			mode |= ssa2.PrintPackages
		case 'F':
			mode |= ssa2.PrintFunctions
		case 'S':
			mode |= ssa2.LogSource | ssa2.BuildSerially
		case 'C':
			mode |= ssa2.SanityCheckFunctions
		case 'G':
<<<<<<< HEAD:cmd/tortoise.go
			conf.Build = nil
=======
			conf.SourceImports = false
>>>>>>> go1.2:cmd/tortoise.go
		case 'L':
			mode |= ssa2.BuildSerially
		case 'I':
			mode |= ssa2.BareInits
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
			mode |= ssa2.GlobalDebug
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

<<<<<<< HEAD:cmd/tortoise.go
	// Load, parse and type-check the program.
	imp := importer.New(&conf)
=======
	// Use the initial packages from the command line.
>>>>>>> go1.2:cmd/tortoise.go
	prog_args := args[1:]
	args, err := conf.FromArgs(args[0:1], *testFlag)
	if err != nil {
		return err
	}

	// The interpreter needs the runtime package.
	if *runFlag {
		conf.Import("runtime")
	}

	// Load, parse and type-check the whole program.
	iprog, err := conf.Load()
	if err != nil {
		return err
	}

	// Create and build SSA-form program representation.
<<<<<<< HEAD:cmd/tortoise.go
=======
	prog := ssa2.Create(iprog, mode)
>>>>>>> go1.2:cmd/tortoise.go
	prog.BuildAll()

	// Run the interpreter.
	if *runFlag {
		var main *ssa2.Package
		pkgs := prog.AllPackages()
		if *testFlag {
			// If -test, run all packages' tests.
			if len(pkgs) > 0 {
				main = prog.CreateTestMainPackage(pkgs...)
			}
			if main == nil {
				return fmt.Errorf("no tests")
			}
		} else {
			// Otherwise, run main.main.
			for _, pkg := range pkgs {
				if pkg.Object.Name() == "main" {
					main = pkg
					if main.Func("main") == nil {
						return fmt.Errorf("no func main() in main package")
					}
					break
				}
			}
			if main == nil {
				return fmt.Errorf("no main package")
			}
		}
<<<<<<< HEAD:cmd/tortoise.go
=======

		if interpTraceMode & interp.EnableStmtTracing != 0 {
			gubcmd.Init(gubFlag, restart_args, main.Prog)
			fn := main.Func("main")
			if fn != nil {
				/* Set a breakpoint on the main routine */
				interp.SetFnBreakpoint(fn)
				bp := &gub.Breakpoint {
					Hits: 0,
					Id: gub.BreakpointNext(),
					Pos: fn.Pos(),
					EndP: fn.EndP(),
					Ignore: 0,
					Kind: "Function",
					Temp: true,
					Enabled: true,
				}
				gub.BreakpointAdd(bp)
			}
		} else if prog.PackagesByPath["github.com/rocky/ssa-interp/trepan"] != nil {
			fmt.Println("I see you've got trepan imported...")
			gubcmd.Init(gubFlag, restart_args, main.Prog)
		}
>>>>>>> go1.2:cmd/tortoise.go

		if interpTraceMode & interp.EnableStmtTracing != 0 {
			gubcmd.Init(gubFlag, restart_args, main.Prog)
		} else if prog.PackagesByPath["github.com/rocky/ssa-interp/trepan"] != nil {
			fmt.Println("I see you've got trepan imported...")
			gubcmd.Init(gubFlag, restart_args, main.Prog)
 		}
		fmt.Println("Running....")
<<<<<<< HEAD:cmd/tortoise.go
		if runtime.GOARCH != conf.Build.GOARCH {
=======
		if runtime.GOARCH != build.Default.GOARCH {
>>>>>>> go1.2:cmd/tortoise.go
			return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
				build.Default.GOARCH, runtime.GOARCH)
		}

		interp.Interpret(main, interpMode, interpTraceMode, conf.TypeChecker.Sizes, main.Object.Path(), prog_args)
<<<<<<< HEAD:cmd/tortoise.go
	} else {
=======
	}  else {
>>>>>>> go1.2:cmd/tortoise.go
		fmt.Println(`Built ok, but not running because "-run" option not given`)
	}
	return nil
}
