// Copyright 2013 Rocky Bernstein.

// info command
//

package gubcmd

import (
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "set"
	gub.Cmds[name] = &gub.CmdInfo{
		SubcmdMgr: &gub.SubcmdMgr{Name: name, Subcmds: make(map[string]*gub.SubcmdInfo)},
		Fn: SetCommand,
		Help: `Generic command for shetting things about the program being debugged.

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
		Fn: SetTraceSubcmd,
		Help: `set trace [on|off]

set instruction tracing
`,
		Min_args: 1,
		Max_args: 3,
		Short_help: "Set instruction tracing on or off",
		Name: "trace",
	})
}

//FIXME: DRY THIS
func SetTraceSubcmd(args []string) {
	onoff := "on"
	if len(args) == 3 {
		onoff = args[2]
	}
	switch ParseOnOff(onoff) {
	case ONOFF_ON:
		if interp.InstTracing() {
			gub.Errmsg("Instruction tracing already on")
		} else {
			gub.Msg("Setting Instruction trace on")
			interp.SetInstTracing()
		}
	case ONOFF_OFF:
		if !interp.InstTracing() {
			gub.Errmsg("Instruction tracing already off")
		} else {
			gub.Msg("Setting Instruction trace off")
			interp.ClearInstTracing()
		}
	case ONOFF_UNKNOWN:
		gub.Errmsg("Expecting 'on' or 'off', got '%s'; nothing done", onoff)
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
func SetHighlightSubcmd(args []string) {
	onoff := "on"
	if len(args) == 3 {
		onoff = args[2]
	}
	switch ParseOnOff(onoff) {
	case ONOFF_ON:
		if *gub.Highlight {
			gub.Errmsg("Highlight is already on")
		} else {
			gub.Msg("Setting highlight on")
			*gub.Highlight = true
		}
	case ONOFF_OFF:
		if !*gub.Highlight {
			gub.Errmsg("highight is already off")
		} else {
			gub.Msg("Setting highlight off")
			*gub.Highlight = false
		}
	case ONOFF_UNKNOWN:
		gub.Msg("Expecting 'on' or 'off', got '%s'; nothing done", onoff)
	}
}

//FIXME: DRY THIS
func SetCommand(args []string) {
	if len(args) == 1 {
		gub.Section("List of set commands")
		for name, subinfo := range gub.Cmds["set"].SubcmdMgr.Subcmds {
			gub.Msg("%-10s -- %s", name, subinfo.Short_help)
		}
	} else {
		switch args[1] {
		case "trace":
			SetTraceSubcmd(args)
		case "highlight":
			SetHighlightSubcmd(args)
		default:
			gub.Errmsg("Unknown 'set' subcommand '%s'", args[1])
		}
	}
}
