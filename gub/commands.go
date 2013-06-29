// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"fmt"
	"os"
	"strconv"

	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/types"


	"ssa-interp"
	"ssa-interp/interp"
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

func BacktraceCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
	// FIXME: should get limit from args
	fr := topFrame
	for i:=0; fr !=nil; fr = fr.Caller(0) {
		pointer := "   "
		if fr == curFrame {
			pointer = "=> "
		}
		fmt.Printf("%s#%d %s\n", pointer, i, StackLocation(fr))
		i++
	}
}

func FinishCommand(args []string) {
	interp.SetStepOut(topFrame)
	fmt.Println("Continuing until return...")
}

func FrameCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	frameIndex, ok := strconv.Atoi(args[1])
	if ok != nil {
		fmt.Printf("Expecting integer frame number; got %s\n",
			args[1])
		return
	}
	if frameIndex >= stackSize {
		fmt.Printf("Frame number %d too large. Max is %d.\n", frameIndex, stackSize-1)
		return
	} else if frameIndex < -stackSize {
		fmt.Printf("Frame number %d too small. Min is %d.\n", frameIndex, -stackSize)
		return
	}

	if frameIndex < 0 { frameIndex = stackSize + frameIndex }

	fr := topFrame
	for i:=0; i<frameIndex && fr !=nil; fr = fr.Caller(0) {
		fmt.Println("next")
		i++
	}
	if fr == nil { return }
	curFrame = fr
	event := ssa2.CALL_ENTER
	if (0 == frameIndex) {
		event = traceEvent
	}
	printLocInfo(curFrame, event)
}

func NextCommand(args []string) {
	interp.SetStepOver(topFrame)
	fmt.Println("Step over...")
}

func HelpCommand(args []string) {
	fmt.Println(`List of commands:
Execution running --
  s: step in
  n: next or step over
  fin: finish or step out
  c: continue

Variables --
  local [*name*]:  show local variable info
  global [*name*]: show global variable info
  param [*name*]: show function parameter info

Tracing --
  +: add instruction tracing
  -: remove instruction tracing

Stack:
  bt: print a backtrace
  frame *num*: switch stack frame

Other:
  ?: this help
  q: quit
`)
}

func GlobalsCommand(args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for k, v := range curFrame.I().Globals {
			if v == nil {
				fmt.Printf("%s: nil\n")
			} else {
				// FIXME: figure out why reflect.lookupCache causes
				// an panic on a nil pointer or invalid address
				if fmt.Sprintf("%s", k) == "reflect.lookupCache" {
					fmt.Println("got one!")
					continue
				}
				fmt.Printf("%s: %s\n", k, interp.ToString(*v))
			}
		}
	} else {
		// This doesn't work and I don't know how to fix it.
		for i:=1; i<=argc; i++ {
			vv := ssa2.NewLiteral(exact.MakeString(args[i]),
				types.Typ[types.String], token.NoPos, token.NoPos)
			// fmt.Println(vv, "vs", interp.ToString(vv))
			v, ok := curFrame.I().Globals[vv]
			if ok {
				fmt.Printf("%s: %s\n", vv, interp.ToString(*v))
			}
		}

		// This is ugly, but I don't know how to turn a string into
		// a ssa2.Value.
		globals := make(map[string]*interp.Value)
		for k, v := range curFrame.I().Globals {
			globals[fmt.Sprintf("%s", k)] = v
		}

		for i:=1; i<=argc; i++ {
			vv := args[i]
			v, ok := globals[vv]
			if ok {
				fmt.Printf("%s: %s\n", vv, interp.ToString(*v))
			}
		}
	}
}

func ParametersCommand(args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 1, args) { return }
	if argc == 0 {
		for i, p := range curFrame.Fn().Params {
			fmt.Println(curFrame.Fn().Params[i], curFrame.Env()[p])
		}
	} else {
		varname := args[1]
		for i, p := range curFrame.Fn().Params {
			if varname == curFrame.Fn().Params[i].Name() {
				fmt.Println(curFrame.Fn().Params[i], curFrame.Env()[p])
				break
			}
		}
	}
}

func LocalsCommand(args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 2, args) { return }
	if argc == 0 {
		i := 0
		for _, v := range curFrame.Locals() {
			name := curFrame.Fn().Locals[i].Name()
			fmt.Printf("%s: %s\n", name, interp.ToString(v))
			i++
		}
	} else {
		varname := args[1]
		i := 0
		for _, v := range curFrame.Locals() {
			if args[1] == curFrame.Fn().Locals[i].Name() {
				fmt.Printf("%s %s: %s\n", varname, curFrame.Fn().Locals[i], interp.ToString(v))
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
func QuitCommand(args []string) {
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

func VariableCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	fn := curFrame.Fn()
	varname := args[1]
	for _, p := range fn.Locals {
		if varname == p.Name() { break }
	}

}
