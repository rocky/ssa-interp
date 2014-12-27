// Copyright 2013-2014 Rocky Bernstein.

/*
This file contains definitions beyond interp.go need for the gub
debugger. This could be merged into interp.go but we keep it separate so
as to make diff'ing our ssa.go and the unmodified ssa.go look more
alike.
*/
package interp
import (
	"fmt"
	"os"
	"github.com/rocky/ssa-interp"
)

func (i  *interpreter) Global(name string, pkg *ssa2.Package)  (v *Value, ok bool) {
	v, ok = i.globals[pkg.Var(name)]
	return
}

// Global interpreter. Flags fo the interpreter are used in trace
var i *interpreter


func init() {
	TraceHook = NullTraceHook
}

func GetInterpreter() *interpreter {
	return i
}

/**** interpreter accessors ****/

func (fr *Frame) Get(key ssa2.Value) Value { return fr.get(key) }
func SetGlobal(i *interpreter, pkg *ssa2.Package, name string, v Value) {
	setGlobal(i, pkg, name, v)
}

// sourcePanic is a panic in the source code rather than a normal panic
// which would be in the interpreter code
func (fr *Frame) sourcePanic(mess string) {
	fmt.Fprintf(os.Stderr, "panic: %s\n", mess)
	gotraceback := os.Getenv("GOTRACEBACK")
	switch gotraceback {
	case "0":
		//do nothing
	case "1":
		runtime۰Gotraceback(fr)
	case "2", "crash":
		runtime۰Gotraceback(fr)
		for _, goTop := range fr.i.goTops {
			otherFr := goTop.Fr
			if otherFr != fr {
				runtime۰Gotraceback(otherFr)
			}
		}
	}

	TraceHook(fr, &fr.block.Instrs[fr.pc], ssa2.PANIC)
	// Don't know if setting fr.status really does anything, but
	// just to try to be totally Kosher. We do this *after*
	// running TraceHook because TraceHook treats panic'd frames
	// differently and will do less with them. If it needs to
	// understand that we are in a panic state, it can do that via
	// the event type passed above.
	fr.status = StPanic
	// We don't need an interpreter traceback. So turn that off.
	//os.Setenv("GOTRACEBACK", "0")
	panic(targetPanic{mess})
}

func (i *interpreter) Program() *ssa2.Program { return i.prog }
func (i  *interpreter) Globals() map[ssa2.Value]*Value { return i.globals }
func (i  *interpreter) GoTops() []*GoreState { return i.goTops }
