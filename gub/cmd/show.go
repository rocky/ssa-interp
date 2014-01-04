// Copyright 2014 Rocky Bernstein.

// show command

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "show"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{
			Name   : name,
			Subcmds: make(gub.SubcmdMap),
		},
		Fn: ShowCommand,
		Help: `Generic command for showing things about the debugger.

Type "set" for a list of "set" subcommands and what they do.
Type "help set *" for just a list of "info" subcommands.`,
		Min_args: 0,
		Max_args: 3,
	}
	gub.AddToCategory("support", name)
}

func init() {
	name := "show"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{
			Name   : name,
			Subcmds: make(gub.SubcmdMap),
		},
		Fn: ShowCommand,
		Help: `Show parts of the debugger environment.

Type "show" for a list of "show" subcommands and what they do.
Type "help show *" for just a list of "show" subcommands.`,
		Min_args: 0,
		Max_args: 3,
	}
	gub.AddToCategory("support", name)
}

func ShowOnOff(subcmdName string, on bool) {
	if on {
		gub.Msg("%s is on.", subcmdName)
	} else {
		gub.Msg("%s is off.", subcmdName)
	}
}

// show implements the debugger command:
//    show [*subcommand]
// which is a generic command for setting things about the debugged program.
func ShowCommand(args []string) {
	gub.SubcmdMgrCommand(args)
}
