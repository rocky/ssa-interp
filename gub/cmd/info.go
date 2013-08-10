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
		SubcmdMgr: &gub.SubcmdMgr{Name: name, Subcmds: make(map[string]*gub.SubcmdInfo)},
		Fn: InfoCommand,
		Help: `Generic command for showing things about the program being debugged.

Type "info" for a list of "info" subcommands and what they do.
Type "help info *" for just a list of "info" subcommands.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("status", name)
}


func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoFrameSubcmd,
		Help: `info frame

Show information about the selected frame.

See also backtrace.
`,
		Min_args: 2,
		Max_args: 2,
		Short_help: "Show information about the selected frame",
		Name: "frame",
	})
}
func InfoFrameSubcmd(args []string) {
	gub.Msg("goroutine number: %d", gub.CurFrame().GoNum())
	gub.Msg("frame: %s", gub.CurFrame().FnAndParamString())
}

func InfoProgramSubcmd(args []string) {
	gub.Msg("instruction number: %d", gub.CurFrame().PC())
	block := gub.CurFrame().Block()
	gub.Msg("basic block: %d", block.Index)
	if block.Scope != nil {
		gub.Msg("scope: %d", block.Scope.ScopeId())
	} else {
		gub.Msg("unknown scope")
	}
	gub.Msg("program stop event: %s", ssa2.Event2Name[gub.TraceEvent])
	gub.Msg("position: %s", gub.CurFrame().PositionRange())
}

func InfoPCSubcmd() {
	fr := gub.CurFrame()
	gub.Msg("instruction number: %d of block %d", fr.PC(), fr.Block().Index)
}

func InfoCommand(args []string) {
	if len(args) == 1 {
		gub.Section("List of info commands")
		for name, subinfo := range gub.Cmds["info"].SubcmdMgr.Subcmds {
			gub.Msg("%-10s -- %s", name, subinfo.Short_help)
		}
	}
	if len(args) == 2 {
		switch args[1] {
		case "args":
			gub.InfoArgsSubcmd(args)
		case "frame":
			InfoFrameSubcmd(args)
		case "program":
			InfoProgramSubcmd(args)
		case "PC", "pc":
			InfoPCSubcmd()
		case "breakpoint", "break":
			InfoBreakpointSubcmd()
		case "scope":
			InfoScopeSubcmd(args)
		case "stack":
			gub.PrintStack(gub.TopFrame(), gub.MAXSTACKSHOW)
		}
	}
}
