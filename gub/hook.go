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
var TraceEvent ssa2.TraceEvent

var gubLock  sync.Mutex

// Commands set InCmdLoop to "false" break out of debugger's read
// command loop.
var InCmdLoop bool

// Some commands like "eval" next the exact text after the command
var CmdArgstr string

// if we are stopped by breakpoint, this is the breakpoint number.
const NoBp = 0xfffff
var curBpnum BpId

func skipEvent(fr *interp.Frame, event ssa2.TraceEvent) bool {
	curBpnum = NoBp
	if event == ssa2.BREAKPOINT {
		bps := BreakpointFindByPos(fr.StartP())
		for _, bpnum := range bps {
			bp := Breakpoints[bpnum]
			if !bp.Enabled { continue }
			// FIXME: check things like the condition
			curBpnum = bpnum
			bp.Hits ++
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
	if !fr.I().TraceEventMask[event] { return }
	gubLock.Lock()
    defer gubLock.Unlock()
	if skipEvent(fr, event) { return }
	frameInit(fr)
	// FIXME: use unconditionally
	if instr == nil {
		instr = &fr.Block().Instrs[fr.PC()]
	}
	if event == ssa2.BREAKPOINT && Breakpoints[curBpnum].Kind == "Function" {
		event = ssa2.CALL_ENTER
	}
	TraceEvent = event
	printLocInfo(topFrame, instr, event)

	line := ""
	var err error
	for InCmdLoop = true; err == nil && InCmdLoop; cmdCount++ {
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
		if len(args) == 0 || len(args[0]) == 0 {
			Msg("Empty line skipped")
			continue
		} else if args[0][0] == '#' {
			Msg(line) // echo line but do nothing
			continue
		}

		name := args[0]
		CmdArgstr = strings.TrimLeft(line[len(name):], " ")
		if newname := LookupCmd(name); newname != "" {
			name = newname
		}
		cmd := Cmds[name];

		if cmd != nil {
			if argCountOK(cmd.Min_args, cmd.Max_args, args) {
				Cmds[name].Fn(args)
			}
			continue
		}

		if len(args) > 0 {
			WhatisName(args[0])
		} else {
			Errmsg("Unknown command %s\n", cmd)
		}
	}
}
