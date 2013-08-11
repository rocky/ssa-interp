// Copyright 2013 Rocky Bernstein.
package gubcmd

import (
	"go/parser"
	"github.com/rocky/ssa-interp/gub"
	"code.google.com/p/go.tools/go/types"
	// "go/ast"
	// "fmt"
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

	fr := gub.CurFrame()
	fset := fr.Fset()
	typesScope := fr.Scope().Scope
	typesPkg := fr.Fn().Pkg.Object
	typ, val, err := types.EvalNode(fset, expr, typesPkg, typesScope)
	// fmt.Println("typ:", typ, ", val:", val, ", err:", err)
	// ast.Print(fset, expr)
	if err == nil {
		if val != nil {
			gub.Msg("%s", val)
		} else {
			if val := gub.EvalExprStart(expr, typ); val != nil {
				gub.Msg("%s", val)
			}
		}
	} else {
		gub.Errmsg("%s", err)
	}
}
