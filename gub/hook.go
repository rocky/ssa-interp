// Copyright 2013 Rocky Bernstein.
// Debugger callback hook. Contains the main command loop.
package gub

import (
	"fmt"
	"strings"
	"sync"

	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

var cmdCount int = 0
var traceEvent ssa2.TraceEvent

var gubLock  sync.Mutex

// if we are stopped by breakpoint, this is the breakpoint number.
// Otherwise this is < 0.
var curBpnum int

func skipEvent(fr *interp.Frame, event ssa2.TraceEvent) bool {
	curBpnum = -1
	if event == ssa2.BREAKPOINT {
		bps := BreakpointFindByPos(fr.StartP())
		for _, bpnum := range bps {
			bp := Breakpoints[bpnum]
			if !bp.enabled { continue }
			// FIXME: check things like the condition
			curBpnum = bpnum
			bp.hits ++
			break
		}
	}
	return false
}

// Compute the gub read prompt. It has the command count and
// a goroutine number if we aren't in the main goroutine.
func computePrompt() string {
	prompt := fmt.Sprintf("gub[%d", cmdCount)
	// If we aren't in the main goroutine, show the goroutine number
	if curFrame.GoNum() != 0 {
		prompt += fmt.Sprintf("@%d", curFrame.GoNum())
	}
	prompt += "] "
	return prompt
}


// Call-back hook from interpreter. Contains top-level statement breakout
func GubTraceHook(fr *interp.Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
    gubLock.Lock()
    defer gubLock.Unlock()
	if nil == I {I = fr.I()}
	traceEvent = event
	if skipEvent(fr, event) { return }
	frameInit(fr)
	printLocInfo(topFrame, instr, event)

	line := ""
	var err error
	for inCmdLoop := true; err == nil && inCmdLoop; cmdCount++ {
		if inputReader != nil {
			line, err = inputReader.ReadString('\n')
		} else {
			line, err = gnureadline.Readline(computePrompt(), true)
		}
        if err != nil {
            break
        }
		line = strings.TrimRight(line, "\n")
		args  := strings.Split(line, " ")
		if len(args) == 0 {
			msg("Empty line skipped")
			continue
		} else if args[0][0] == '#' {
			msg(line) // echo line but do nothing
			continue
		}

		cmd := args[0]

		switch cmd {
		case "+":
			fmt.Println("Setting Instruction Trace")
			interp.SetInstTracing()
		case "-":
			fmt.Println("Clearing Instruction Trace")
			interp.ClearInstTracing()
		case "bt", "T", "backtrace", "where":
			BacktraceCommand(args)
		case "b", "break", "breakpoint":
			BreakpointCommand(args)
		case "delete":
			DeleteCommand(args)
		case "disassemble", "disasm":
			DisassembleCommand(args)
		case "disable":
			DisableCommand(args)
		case "down":
			DownCommand(args)
		case "enable":
			EnableCommand(args)
		case "env":
			for i, p := range topFrame.Env() {
				fmt.Println(i, p)
			}
		case "h", "?", "help":
			HelpCommand(args)
		case "c", "continue":
			Continue(args)
			inCmdLoop = false
			break
		case "finish", "fin":
 			FinishCommand(args)
			inCmdLoop = false
			break
		case "frame":
			FrameCommand(args)
		case "gor", "gore", "goroutine", "goroutines":
			GoroutinesCommand(args)
		case "gl", "global", "globals":
			GlobalsCommand(args)
		case "locs":
			LocsCommand(args)
		case "lo", "local", "locals":
			LocalsCommand(args)
		case "param", "parameters":
			ParametersCommand(args)
		case "next", "n":
 			NextCommand(args)
			inCmdLoop = false
			break
		case "q", "quit", "exit":
			QuitCommand(args)
		case "s", "step":
			StepCommand(args)
			inCmdLoop = false
			break
		case "up":
			UpCommand(args)
		case "v":
			VariableCommand(args)
		case "whatis":
			WhatisCommand(args)
		default:
			if len(args) > 0 {
				WhatisName(args[0])
			} else {
				errmsg("Unknown command %s\n", cmd)
			}
		}
	}
}
