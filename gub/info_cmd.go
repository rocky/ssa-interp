// Copyright 2013 Rocky Bernstein.

// info command
//

package gub

import (
	"github.com/rocky/ssa-interp"
)

func init() {
	name := "info"
	Cmds[name] = &CmdInfo{
		Fn: InfoCommand,
		Help: `info [args|breakpoints|frame|program|stack]

Generic command for showing things about the program being debugged.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("status", name)
}

func InfoFrameSubcmd() {
	Msg("goroutine number: %d", curFrame.GoNum())
	Msg("frame: %s", curFrame.FnAndParamString())
}

func InfoProgramSubcmd() {
	Msg("instruction number: %d", curFrame.PC())
	block := curFrame.Block()
	Msg("basic block: %d", block.Index)
	if block.Scope != nil {
		Msg("scope: %d", block.Scope.ScopeNum())
	} else {
		Msg("unknown scope")
	}
	Msg("program stop event: %s", ssa2.Event2Name[traceEvent])
	Msg("position: %s", curFrame.PositionRange())
}

func InfoCommand(args []string) {
	if len(args) == 2 {
		switch args[1] {
		case "args":
			InfoArgsSubcmd(args)
		case "frame":
			InfoFrameSubcmd()
		case "program":
			InfoProgramSubcmd()
		case "breakpoint", "break":
			InfoBreakpointSubcmd()
		case "scope":
			InfoScopeSubcmd(args)
		case "stack":
			printStack(topFrame, MAXSTACKSHOW)
		}
	}
}
