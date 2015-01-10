// Copyright 2013-2014 Rocky Bernstein.

// info command
//

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)


func init() {
	name := "info"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{
			Name   : name,
			Subcmds: make(gub.SubcmdMap),
		},
		Fn: InfoCommand,
		Help: `Generic command for showing things about the program being debugged.

Type "info" for a list of "info" subcommands and what they do.
Type "help info *" for just a list of "info" subcommands.
`,
		Min_args: 0,
		Max_args: -1,
	}
	gub.AddToCategory("status", name)
}


func InfoCommand(args []string) {
	gub.SubcmdMgrCommand(args)
}
