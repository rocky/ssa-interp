// Copyright 2013-2014 Rocky Bernstein.

// set command

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "set"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{
			Name   : name,
			Subcmds: make(gub.SubcmdMap),
		},
		Fn: SetCommand,
		Help: `Modifies parts of the debugger environment.

Type "set" for a list of "set" subcommands and what they do.
Type "help set *" for just a list of "set" subcommands.
`,
		Min_args: 0,
		Max_args: 3,
	}
	gub.AddToCategory("support", name)
}


type onoff uint8
const (
	ONOFF_ON = iota
	ONOFF_OFF
	ONOFF_UNKNOWN
)

func ParseOnOff(onoff string) onoff {
	switch onoff {
	case "on", "1", "yes":
		return ONOFF_ON
	case "off", "0", "none":
		return ONOFF_OFF
	default:
		return ONOFF_UNKNOWN
	}
}

func init() {
	parent := "set"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: SetHighlightSubcmd,
		Help: `Modifies parts of the debugger environment.

You can give unique prefix of the name of a subcommand to get
information about just that subcommand.

Type "set" for a list of "set" subcommands and what they do.
Type "help set *" for just the list of "set" subcommands.`,
		Min_args: 0,
		Max_args: 3,
	})
}

// setCommand implements the debugger command:
//    set [*subcommand*]
// which modifies parts of the debugger environment.
func SetCommand(args []string) {
	gub.SubcmdMgrCommand(args)
}
