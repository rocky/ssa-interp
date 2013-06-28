package interp

import (
	"fmt"
	"os"

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

func GlobalsCommand(fr *frame, args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for k, v := range fr.I.Globals {
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
			v, ok := fr.I.Globals[vv]
			if ok {
				fmt.Printf("%s: %s\n", vv, toString(*v))
			}
		}

		// This is ugly, but I don't know how to turn a string into
		// a ssa2.Value.
		globals := make(map[string]*value)
		for k, v := range fr.I.Globals {
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
	if !argCountOK(0, 1, args) { return }
	for k, v := range fr.Locals {
		fmt.Printf("%s: %s\n", toString(k), toString(v))
	}
}


func QuitCommand(fr *frame, args []string) {
	if !argCountOK(0, 1, args) { return }
	fmt.Println("That's all folks...")
	gnureadline.Rl_reset_terminal(term)
	os.Exit(0)  // FIXME: Should use int arg

}

func VariableCommand(fr *frame, args []string) {
	if !argCountOK(1, 1, args) { return }
	fn := fr.Fn
	varname := args[1]
	for _, p := range fn.Locals {
		if varname == p.Name() { break }
	}

}
