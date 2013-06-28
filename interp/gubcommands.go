package interp

import (
	"fmt"
	"os"
	"strconv"

	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/types"


	"ssa-interp"
	"gnureadline"
)

func argCountOK(min int, max int, args [] string) bool {
	l := len(args)-1 // strip command name from count
	if (l < min) {
		fmt.Printf("Too few args; need at least %d, got %d\n", min, l)
		return false
	} else if (l > max) {
		fmt.Printf("Too many args; need at most %d, got %d\n", max, l)
		return false
	}
	return true
}

func BacktraceCommand(fr *frame, args []string) {
	if !argCountOK(0, 1, args) { return }
	// FIXME: should get limit from args
	curFrame := fr
	for i:=0; curFrame !=nil; curFrame = curFrame.Caller {
		fmt.Printf("   #%d %s\n", i, StackLocation(curFrame))
		i++
	}
}

func FinishCommand(fr *frame, args []string) {
	SetStepOut(fr)
	fmt.Println("Continuing until return...")
}

func NextCommand(fr *frame, args []string) {
	SetStepOver(fr)
	fmt.Println("Step over...")
}

func HelpCommand(fr *frame, args []string) {
	fmt.Println(`List of commands:
Execution running --
  s: step in
  n: next or step over
  fin: finish or step out
  c: continue

Variables --
  local [name]:  show local variables info
  global [name]: show global variable info
  param: show function parameters

Tracing --
  +: add instruction tracing
  -: remove instruction tracing

Stack:
  bt: print a backtrace

Other:
  ?: this help
  q: quit
`)
}

func GlobalsCommand(fr *frame, args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for k, v := range fr.i.Globals {
			if v == nil {
				fmt.Printf("%s: nil\n")
			} else {
				// FIXME: figure out why reflect.lookupCache causes
				// an panic on a nil pointer or invalid address
				if fmt.Sprintf("%s", k) == "reflect.lookupCache" {
					fmt.Println("got one!")
					continue
				}
				fmt.Printf("%s: %s\n", k, toString(*v))
			}
		}
	} else {
		// This doesn't work and I don't know how to fix it.
		for i:=1; i<=argc; i++ {
			vv := ssa2.NewLiteral(exact.MakeString(args[i]),
				types.Typ[types.String], token.NoPos, token.NoPos)
			// fmt.Println(vv, "vs", toString(vv))
			v, ok := fr.i.Globals[vv]
			if ok {
				fmt.Printf("%s: %s\n", vv, toString(*v))
			}
		}

		// This is ugly, but I don't know how to turn a string into
		// a ssa2.Value.
		globals := make(map[string]*value)
		for k, v := range fr.i.Globals {
			globals[fmt.Sprintf("%s", k)] = v
		}

		for i:=1; i<=argc; i++ {
			vv := args[i]
			v, ok := globals[vv]
			if ok {
				fmt.Printf("%s: %s\n", vv, toString(*v))
			}
		}
	}
}

func ParametersCommand(fr *frame, args []string) {
	for i, p := range fr.Fn.Params {
		fmt.Printf("%d %s: %s\n", i, p.Name(), p.Type())
	}
}


func LocalsCommand(fr *frame, args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 2, args) { return }
	if argc == 0 {
		i := 0
		for _, v := range fr.Locals {
			name := fr.Fn.Locals[i].Name()
			fmt.Printf("%s: %s\n", name, toString(v))
			i++
		}
	} else {
		varname := args[1]
		i := 0
		for _, v := range fr.Locals {
			if args[1] == fr.Fn.Locals[i].Name() {
				fmt.Printf("%s %s: %s\n", varname, fr.Fn.Locals[i], toString(v))
				break
			}
			i++
		}

	}
}

// quit [exit-code]
//
// Terminates program. If an exit code is given, that is the exit code
// for the program. Zero (normal termination) is used if no
// termintation code.
func QuitCommand(fr *frame, args []string) {
	if !argCountOK(0, 1, args) { return }
	rc := 0
	if len(args) == 2 {
		new_rc, ok := strconv.Atoi(args[1])
		if ok == nil { rc = new_rc } else {
			fmt.Printf("Expecting integer return code; got %s. Ignoring\n",
				args[1])
		}
	}
	fmt.Println("That's all folks...")
	gnureadline.Rl_reset_terminal(term)
	os.Exit(rc)

}

func VariableCommand(fr *frame, args []string) {
	if !argCountOK(1, 1, args) { return }
	fn := fr.Fn
	varname := args[1]
	for _, p := range fn.Locals {
		if varname == p.Name() { break }
	}

}
