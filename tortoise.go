// +build ignore

package main

// tortoise: a tool for displaying and interpreting Go programs.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"code.google.com/p/go.tools/importer"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/ssa-interp/gub"
)

var buildFlag = flag.String("build", "", `Options controlling the SSA builder.
The value is a sequence of zero or more of these letters:
C	perform sanity [C]hecking of the SSA form.
D	include debug info for every function.
P	log [P]ackage inventory.
F	log [F]unction SSA code.
S	log [S]ource locations as SSA builder progresses.
G	use binary object files from gc to provide imports (no code).
L	build distinct packages seria[L]ly instead of in parallel.
`)

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
`

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	args := flag.Args()

	impctx := importer.Config{Loader: importer.MakeGoBuildLoader(nil)}

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
			impctx.Loader = nil
		case 'L':
			mode |= ssa2.BuildSerially
		default:
			log.Fatalf("Unknown -build option: '%c'.", c)
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
			log.Fatalf("Unknown -interp option: '%c'.", c)
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
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Load, parse and type-check the program.
	imp := importer.New(&impctx)
	info, args, err := importer.CreatePackageFromArgs(imp, args)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create and build SSA-form program representation.
	prog := ssa2.NewProgram(imp.Fset, mode)
	for _, info := range imp.Packages {
		prog.CreatePackage(info)
	}
	prog.BuildAll()

	// Run the interpreter.
	if *runFlag {
		fmt.Println("Running....")
		if interpTraceMode & interp.EnableStmtTracing != 0 {
			gub.Install(gubFlag)
		}
		interp.Interpret(prog.Package(info.Pkg), interpMode, interpTraceMode,
			info.Pkg.Path(), args)
	}
}
