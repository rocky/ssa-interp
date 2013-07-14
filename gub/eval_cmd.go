// Copyright 2013 Rocky Bernstein.
package gub

import (
	"go/parser"
)

func init() {
	name := "eval"
	cmds[name] = &CmdInfo{
		fn: EvalCommand,
		help: `eval *expr*

Evaluate go expression *expr*.
`,

		min_args: 1,
		max_args: -1,
	}
	AddToCategory("data", name)
}

func EvalCommand(args []string) {

	// Don't use args, but cmdArgstr which preserves blanks inside quotes
	expr, err := parser.ParseExpr(cmdArgstr)
	if err != nil {
		errmsg("Error parsing %s: %s", cmdArgstr, err.Error())
		return
	}
	if val := evalExpr(expr); val != nil {
		msg("%s", val)
	}
}
