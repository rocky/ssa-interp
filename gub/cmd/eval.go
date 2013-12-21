// Copyright 2013 Rocky Bernstein.
package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "eval"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: EvalCommand,
		Help: `eval *expr*

Evaluate go expression *expr*.
`,

		Min_args: 1,
		Max_args: -1,
	}
	gub.AddToCategory("data", name)
}

func EvalCommand(args []string) {

	// Don't use args, but gub.CmdArgstr which preserves blanks inside quotes
	if expr, err := gub.EvalExprInteractive(gub.CmdArgstr); err == nil {
		if expr == nil {
			gub.Msg("nil")
		} else {
			gub.Msg("%v", (*expr)[0].Interface())
		}
	}
}
