// Copyright 2014 Rocky Bernstein.

// info stack
//
// Same thing as "backtrace"

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoStackSubcmd,
		Help: `info stack

Same thing as "backtrace"
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "Same thing as \"backtrace\"",
		Name: "stack",
	})
}

func InfoStackSubcmd(args []string) {
	gub.PrintStack(gub.TopFrame(), gub.MAXSTACKSHOW)
}
