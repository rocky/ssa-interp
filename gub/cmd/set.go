// Copyright 2013 Rocky Bernstein.

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
		Help: `Generic command for setting things about the debugged program.

Type "set" for a list of "set" subcommands and what they do.
Type "help set *" for just a list of "info" subcommands.
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
		Help: `set highlight [on|off]

Set whether we use terminal highlighting.
`,
		Min_args: 2,
		Max_args: 3,
		Short_help: "Set whether we use terminal highlighting",
		Name: "highlight",
	})
}

//FIXME: DRY THIS

// setCommand implements the debugger command:
//    set [*|subcommand]
// which is a generic command for setting things about the debugged program.
//
// Type "set" for a list of "set" subcommands and what they do.
//
// Type "help set *" for just a list of "info" subcommands.
func SetCommand(args []string) {
	if len(args) == 1 {
		gub.ListSubCommandArgs(gub.Cmds["set"].SubcmdMgr)
		return
	}

    subcmd_name := args[1]
	subcmds     := gub.Cmds["set"].SubcmdMgr.Subcmds
	subcmd_info := subcmds[subcmd_name]

	if subcmd_info != nil {
		if gub.ArgCountOK(subcmd_info.Min_args+1, subcmd_info.Max_args+1, args) {
			subcmds[subcmd_name].Fn(args)
		}
		return
	}

	gub.Errmsg("Unknown 'set' subcommand '%s'", args[1])
}
