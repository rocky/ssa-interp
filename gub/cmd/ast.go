// Copyright 2013 Rocky Bernstein.
// quit command

package gubcmd

import (
	"fmt"
	"go/ast"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "ast"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: AstCommand,
		Help: `ast

Prints AST for current function
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("files", name)
}

// AstCommand implements the debugger command: ast
//
// ast
//
// Prints AST for current function.
func AstCommand(args []string) {
	fn := gub.CurFrame().Fn()
	if syntax := fn.Syntax(); syntax != nil {
		ast.Print(fn.Prog.Fset, syntax)
		fmt.Println("");
	} else {
		gub.Msg("Sorry, we don't have an AST for %s", fn);
	}
}
