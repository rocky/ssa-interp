// Copyright 2014 Rocky Bernstein.

// show trace - whether to tracing is in effect?

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	parent := "show"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: ShowTraceSubcmd,
		Help: `show trace

Show interpreter instruction tracing status`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "show interpreter instruction tracing",
		Name: "trace",
	})
}

func ShowTraceSubcmd(args []string) {
	ShowOnOff(args[1], interp.InstTracing())
}
