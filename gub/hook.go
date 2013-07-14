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

// Commands set inCmdLoop to "false" break out of debugger's read
// command loop.
var inCmdLoop bool

// Some commands like "eval" next the exact text after the command
var cmdArgstr string

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
// FIXME: remove instr
func GubTraceHook(fr *interp.Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
    gubLock.Lock()
    defer gubLock.Unlock()
	traceEvent = event
	if skipEvent(fr, event) { return }
	frameInit(fr)
	// FIXME: use
	// genericInstr = fr.Block().Instrs[ic]
	printLocInfo(topFrame, instr, event)

	line := ""
	var err error
	for inCmdLoop = true; err == nil && inCmdLoop; cmdCount++ {
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

		name := args[0]
		cmdArgstr = strings.TrimLeft(line[len(name):], " ")
		if newname := lookupCmd(name); newname != "" {
			name = newname
		}
		cmd := cmds[name];

		if cmd != nil {
			if argCountOK(cmd.min_args, cmd.max_args, args) {
				cmds[name].fn(args)
			}
			continue
		}

		switch name {
		case "+":
			fmt.Println("Setting Instruction Trace")
			interp.SetInstTracing()
		case "-":
			fmt.Println("Clearing Instruction Trace")
			interp.ClearInstTracing()
		case "lo", "local", "locals":
			LocalsCommand(args)
		case "param", "parameters":
			ParametersCommand(args)
		case "v":
			VariableCommand(args)
		default:
			if len(args) > 0 {
				WhatisName(args[0])
			} else {
				errmsg("Unknown command %s\n", cmd)
			}
		}
	}
}
