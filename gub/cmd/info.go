// Copyright 2013 Rocky Bernstein.

// info command
//

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

var subcmds = make(gub.SubcmdMap)

func init() {
	name := "info"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{Name: name, Subcmds: subcmds},
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


func InfoCommand(args []string) {
	if len(args) == 1 {
		gub.ListSubCommandArgs(gub.Cmds["info"].SubcmdMgr)
		return
	}

    subcmd_name := args[1]
	subcmd_info := subcmds[subcmd_name]

	if subcmd_info != nil {
		if gub.ArgCountOK(subcmd_info.Min_args+1, subcmd_info.Max_args+1, args) {
			subcmds[subcmd_name].Fn(args)
		}
		return
	}

	// FIXME: remove the below.
	if len(args) >= 2 {
		switch subcmd_name {
		case "PC":
			InfoPCSubcmd(args)
		case "break":
			InfoBreakpointSubcmd()
		case "stack":
			gub.PrintStack(gub.TopFrame(), gub.MAXSTACKSHOW)
		}
	}
}
