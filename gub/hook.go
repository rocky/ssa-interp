// Copyright 2013 Rocky Bernstein.
// Debugger callback hook
package gub

import (
	"fmt"
	"strings"

	"gnureadline"
	"ssa-interp"
	"ssa-interp/interp"
)

// Call-back hook from interpreter. Contains top-level statement breakout
func GubTraceHook(topFrame *interp.Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {

	fset := topFrame.Fn().Prog.Fset
	startP := fset.Position(topFrame.StartP())
	endP   := fset.Position(topFrame.EndP())
	printLocInfo(topFrame, startP, endP, event)

	line := ""
	var err error
	curFrame := topFrame
	for inCmdLoop := true; err == nil && inCmdLoop; cmdCount++ {
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
			interp.SetStepIn(curFrame)
			inCmdLoop = false
			break
		case "h", "?", "help":
			HelpCommand(topFrame, args)
		case "c":
			interp.SetStepOff(topFrame)
			fmt.Println("Continuing...")
			inCmdLoop = false
			break
		case "finish", "fin":
 			FinishCommand(topFrame, args)
			inCmdLoop = false
			break
		case "next", "n":
 			NextCommand(topFrame, args)
			inCmdLoop = false
			break
		case "env":
			for i, p := range topFrame.Env() {
				fmt.Println(i, p)
			}
		case "+":
			fmt.Println("Setting Instruction Trace")
			interp.SetInstTracing()
		case "-":
			fmt.Println("Clearing Instruction Trace")
			interp.ClearInstTracing()
		case "gl", "global", "globals":
			GlobalsCommand(topFrame, args)
		case "lo", "local", "locals":
			LocalsCommand(curFrame, args)
		case "param", "parameters":
			ParametersCommand(curFrame, args)
		case "q", "quit", "exit":
			QuitCommand(topFrame, args)
		case "bt", "T", "backtrace":
			BacktraceCommand(topFrame, args)
		case "v":
			VariableCommand(curFrame, args)
		default:
			fmt.Printf("Unknown command %s\n", cmd)
		}
	}
}
