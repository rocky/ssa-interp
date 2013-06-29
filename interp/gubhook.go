// Copyright 2013 Rocky Bernstein.
// Debugger callback hook
package interp

import (
	"fmt"
	"strings"

	"gnureadline"
	"ssa-interp"
)

// Call-back hook from interpreter. Contains top-level statement breakout
func GubTraceHook(fr *frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	fset := fr.Fn().Prog.Fset
	startP := fset.Position(fr.StartP())
	endP   := fset.Position(fr.EndP())
	printLocInfo(fr, startP, endP, event)
	line := ""
	inCmdLoop := true
	var err error
	for ; err == nil && inCmdLoop; cmdCount++ {
		line, err = gnureadline.Readline(fmt.Sprintf("gub[%d]: ", cmdCount),
			true)
		args  := strings.Split(line, " ")
		if len(args) == 0 {
			fmt.Println("Empty line skipped")
			continue
		}

		cmd := args[0]

		switch cmd {
		case "s":
			fmt.Println("Stepping...")
			SetStepIn(fr)
			inCmdLoop = false
			break
		case "h", "?", "help":
			HelpCommand(fr, args)
		case "c":
			SetStepOff(fr)
			fmt.Println("Continuing...")
			inCmdLoop = false
			break
		case "finish", "fin":
 			FinishCommand(fr, args)
			inCmdLoop = false
			break
		case "next", "n":
 			NextCommand(fr, args)
			inCmdLoop = false
			break
		case "env":
			for i, p := range fr.Env {
				fmt.Println(i, p)
			}
		case "+":
			fmt.Println("Setting Instruction Trace")
			SetInstTracing()
		case "-":
			fmt.Println("Clearing Instruction Trace")
			ClearInstTracing()
		case "gl", "global", "globals":
			GlobalsCommand(fr, args)
		case "lo", "local", "locals":
			LocalsCommand(fr, args)
		case "param", "parameters":
			ParametersCommand(fr, args)
		case "q", "quit", "exit":
			QuitCommand(fr, args)
		case "bt", "T", "backtrace":
			BacktraceCommand(fr, args)
		case "v":
			VariableCommand(fr, args)
		default:
			fmt.Printf("Unknown command %s\n", cmd)
		}
	}
}
