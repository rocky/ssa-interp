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
	"ssa2"
	"ssa2/interp"
)

var buildFlag = flag.String("build", "", `Options controlling the SSA builder.
The value is a sequence of zero or more of these letters:
C	perform sanity [C]hecking of the SSA form.
P	log [P]ackage inventory.
F	log [F]unction SSA code.
S	log [S]ource locations as SSA builder progresses.
G	use binary object files from gc to provide imports (no code).
L	build distinct packages seria[L]ly instead of in parallel.
`)

var runFlag = flag.Bool("run", false, "Invokes the SSA interpreter on the program.")

var interpFlag = flag.String("interp", "", `Options controlling the SSA test interpreter.
The value is a sequence of zero or more more of these letters:
R	disable [R]ecover() from panic; show interpreter crash instead.
T	[T]race execution of the program.  Best for single-threaded programs!
`)

const usage = `SSA builder and interpreter.
Usage: tortoise [<flag> ...] [<file.go> ...] [<arg> ...]
       tortoise [<flag> ...] <import/path>   [<arg> ...]
Use -help flag to display options.

Examples:
% tortoise -run -interp=T hello.go     # interpret a program, with tracing
% tortoise -build=FPG hello.go         # quickly dump SSA form of a single package
`

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	args := flag.Args()

	impctx := importer.Context{Loader: importer.MakeGoBuildLoader(nil)}

	var mode ssa.BuilderMode
	mode |= ssa.NaiveForm

	for _, c := range *buildFlag {
		switch c {
		case 'P':
			mode |= ssa.LogPackages | ssa.BuildSerially
		case 'F':
			mode |= ssa.LogFunctions | ssa.BuildSerially
		case 'S':
			mode |= ssa.LogSource | ssa.BuildSerially
		case 'C':
			mode |= ssa.SanityCheckFunctions
		case 'G':
			impctx.Loader = nil
		case 'L':
			mode |= ssa.BuildSerially
		default:
			log.Fatalf("Unknown -build option: '%c'.", c)
		}
	}

	var interpMode interp.Mode
	for _, c := range *interpFlag {
		switch c {
		case 'T':
			interpMode |= interp.EnableTracing
		case 'R':
			interpMode |= interp.DisableRecover
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
	prog := ssa.NewProgram(imp.Fset, mode)
	prog.CreatePackages(imp)
	prog.BuildAll()

	// Run the interpreter.
	if *runFlag {
		fmt.Println("Running....")
		interp.Interpret(prog.Package(info.Pkg), interpMode, info.Pkg.Path(), args)
	}
}
