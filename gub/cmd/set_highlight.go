// Copyright 2013 Rocky Bernstein.

// set highlight - use terminal highlight?

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "set"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: SetHighlightSubcmd,
		Help: `set highlight [on|off]

Sets whether terminal highlighting is to be used`,
		Min_args: 0,
		Max_args: 1,
		Short_help: "use terminal highlight",
		Name: "highlight",
	})
}

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
