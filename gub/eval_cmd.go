// Copyright 2013 Rocky Bernstein.
package gub

import (
	"go/parser"
)

func init() {
	name := "eval"
	Cmds[name] = &CmdInfo{
		Fn: EvalCommand,
		Help: `eval *expr*

Evaluate go expression *expr*.
`,

		Min_args: 1,
		Max_args: -1,
	}
	AddToCategory("data", name)
}

func EvalCommand(args []string) {

	// Don't use args, but cmdArgstr which preserves blanks inside quotes
	expr, err := parser.ParseExpr(cmdArgstr)
	if err != nil {
		Errmsg("Error parsing %s: %s", cmdArgstr, err.Error())
		return
	}
	if val := evalExpr(expr); val != nil {
		Msg("%s", val)
	}
}
