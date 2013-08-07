// Copyright 2013 Rocky Bernstein.
package gubcmd

import (
	"go/parser"
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
	expr, err := parser.ParseExpr(gub.CmdArgstr)
	if err != nil {
		gub.Errmsg("Error parsing %s: %s", gub.CmdArgstr, err.Error())
		return
	}
	if val := gub.EvalExpr(expr); val != nil {
		gub.Msg("%s", val)
	}
}
