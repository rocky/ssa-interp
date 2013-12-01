// Copyright 2013 Rocky Bernstein.

// set trace - various kinds of tracing

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	parent := "set"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: SetTraceSubcmd,
		Help: `set trace [on|off]

set instruction tracing
`,
		Min_args: 0,
		Max_args: 1,
		Short_help: "Set instruction tracing on or off",
		Name: "trace",
	})
}

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
