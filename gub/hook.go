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

// InCmdLoop is set to "false" to break out of debugger's read command
// loop.
var InCmdLoop bool

// CmdArgstr contains the commad line read in. Some commands like
// "eval" need the exact text after the command.
var CmdArgstr string

// NoBp contains the breakpoint number if we are stopped at and by a breakpoint.
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

// computePrompt computes the gub read prompt. It has the command
// count and a goroutine number if we aren't in the main goroutine.
func computePrompt() string {
	prompt := fmt.Sprintf("gub[%d", cmdCount)
	// If we aren't in the main goroutine, show the goroutine number
	if curFrame.GoNum() != 0 {
		prompt += fmt.Sprintf("@%d", curFrame.GoNum())
	}
	prompt += "] "
	return prompt
}

// LastCommand is the last stepping command. It is used
// when an empty line is entered.
var LastCommand string = ""

var Instr *ssa2.Instruction

// FIXME: remove instr

// GubTraceHook is the callback hook from interpreter. It contains
// top-level statement breakout.
func GubTraceHook(fr *interp.Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	if !fr.I().TraceEventMask[event] { return }
	gubLock.Lock()
    defer gubLock.Unlock()
	if skipEvent(fr, event) { return }
	TraceEvent = event
	frameInit(fr)
	// FIXME: use unconditionally
	if instr == nil {
		instr = &fr.Block().Instrs[fr.PC()]
	}
	Instr = instr

	if event == ssa2.BREAKPOINT && Breakpoints[curBpnum].Kind == "Function" {
		event = ssa2.CALL_ENTER
	}
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
		line = strings.Trim(line, " \t\n")
		args  := strings.Split(line, " ")
		if len(args) == 0 || len(args[0]) == 0 {
			if len(LastCommand) == 0 {
				Msg("Empty line skipped")
				gnureadline.RemoveHistory(gnureadline.HistoryLength()-1)
				continue
			} else {
				line = LastCommand
				args = strings.Split(line, " ")
			}
		}
		if args[0][0] == '#' {
			gnureadline.RemoveHistory(gnureadline.HistoryLength()-1)
			Msg(line) // echo line but do nothing
			continue
		}

		name := args[0]
		CmdArgstr = strings.TrimLeft(line[len(name):], " ")
		if newname := LookupCmd(name); newname != "" {
			name = newname
		}
		cmd := Cmds[name];
		LastCommand = ""

		if cmd != nil {
			if ArgCountOK(cmd.Min_args, cmd.Max_args, args) {
				Cmds[name].Fn(args)
			}
			continue
		}

		if len(args) > 0 {
			if !WhatisName(args[0]) {
				gnureadline.RemoveHistory(gnureadline.HistoryLength()-1)
			}
		} else {
			gnureadline.RemoveHistory(gnureadline.HistoryLength()-1)
			Errmsg("Unknown command %s\n", cmd)
		}
	}
}
