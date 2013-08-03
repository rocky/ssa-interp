// Copyright 2013 Rocky Bernstein.

// info command
//

package gub

import ("github.com/rocky/ssa-interp")

func init() {
	name := "info"
	cmds[name] = &CmdInfo{
		fn: InfoCommand,
		help: `info [args|breakpoints|frame|program|stack]

Generic command for showing things about the program being debugged.
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("status", name)
}

func InfoFrameSubcmd() {
	msg("goroutine number: %d", curFrame.GoNum())
	msg("frame: %s", StackLocation(curFrame))
}

func InfoProgramSubcmd() {
	msg("instruction pc: %d", curFrame.PC())
	block := curFrame.Block()
	msg("basic block: %d", block.Index)
	msg("scope: %d", block.Scope.ScopeNum())
	msg("program stop event: %s", ssa2.Event2Name[traceEvent])
	msg("position: %s", curFrame.PositionRange())
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
