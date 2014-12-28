// Copyright 2014 Rocky Bernstein.

// show trace - whether to tracing is in effect?

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "show"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: ShowArgsSubcmd,
		Help: `show args

Show argument list to give program being debugged when it is started.
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "Show argument list to give program being debugged when it is started",
		Name: "args",
	})
}

func ShowArgsSubcmd(args []string) {
	gub.Msg(gub.RESTART_ARGS[0])
	for i:=1; i<len(gub.RESTART_ARGS); i++ {
		gub.Msg("\t"+gub.RESTART_ARGS[i])
	}
}
