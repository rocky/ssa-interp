// Copyright 2013 Rocky Bernstein.

// info command
//

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "info"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: InfoCommand,
		Help: `info [args|breakpoints|frame|program|stack]

Generic command for showing things about the program being debugged.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("status", name)
}

func InfoFrameSubcmd() {
	gub.Msg("goroutine number: %d", gub.CurFrame().GoNum())
	gub.Msg("frame: %s", gub.CurFrame().FnAndParamString())
}

func InfoProgramSubcmd() {
	gub.Msg("instruction number: %d", gub.CurFrame().PC())
	block := gub.CurFrame().Block()
	gub.Msg("basic block: %d", block.Index)
	if block.Scope != nil {
		gub.Msg("scope: %d", block.Scope.ScopeNum())
	} else {
		gub.Msg("unknown scope")
	}
	gub.Msg("program stop event: %s", ssa2.Event2Name[gub.TraceEvent])
	gub.Msg("position: %s", gub.CurFrame().PositionRange())
}

func InfoCommand(args []string) {
	if len(args) == 2 {
		switch args[1] {
		case "args":
			gub.InfoArgsSubcmd(args)
		case "frame":
			InfoFrameSubcmd()
		case "program":
			InfoProgramSubcmd()
		case "breakpoint", "break":
			gub.InfoBreakpointSubcmd()
		case "scope":
			gub.InfoScopeSubcmd(args)
		case "stack":
			gub.PrintStack(gub.TopFrame(), gub.MAXSTACKSHOW)
		}
	}
}
