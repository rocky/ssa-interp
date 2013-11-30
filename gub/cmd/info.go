// Copyright 2013 Rocky Bernstein.

// info command
//

package gubcmd

import (
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
		Max_args: 2,
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
func InfoCommand(args []string) {
	if len(args) == 1 {
		gub.Section("List of info commands")
		for name, subinfo := range gub.Cmds["info"].SubcmdMgr.Subcmds {
			gub.Msg("%-10s -- %s", name, subinfo.Short_help)
		}
	}
	// FIXME check len(args) per subcommand
	if len(args) >= 2 {
		switch args[1] {
		case "args":
			gub.InfoArgsSubcmd(args)
		case "frame":
			InfoFrameSubcmd(args)
		case "program":
			InfoProgramSubcmd(args)
		case "PC", "pc":
			InfoPCSubcmd(args)
		case "breakpoint", "break":
			InfoBreakpointSubcmd()
		case "node":
			InfoNodeSubcmd(args)
		case "scope":
			InfoScopeSubcmd(args)
		case "stack":
			gub.PrintStack(gub.TopFrame(), gub.MAXSTACKSHOW)
		}
	}
}
