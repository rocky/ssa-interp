// Copyright 2014 Rocky Bernstein.
// ast command

package gubcmd

import (
	"fmt"
	"go/ast"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "ast"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: AstCommand,
		Help: `ast [function]


Prints formatted source code from the AST. If a function is given, the
source for function is printed. If no function is given, we give the
source code for the high level statement that we are stopped at. Ths
could be for example an assignemnt statement, a for loop, a condition
in a for loop, or initializer in a for loop, to name just a few
examples.

See also the "format" command.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("files", name)
}

// AstCommand implements the debugger command: ast
//
// ast
//
// Prints AST for current function.
func AstCommand(args []string) {
	var syntax ast.Node
	var err error
	fr := gub.CurFrame()
	fn := fr.Fn()
	if len(args) > 1 {
		name := args[1]
		fn, err = gub.FuncLookup(name)
		if err != nil {
			gub.Errmsg(err.Error())
			return
		} else if fn == nil {
			gub.Errmsg("function '%s' not found", name)
			return
		} else {
			syntax = fn.Syntax()
		}
	} else {
		syntax = fn.Syntax()
		if pc := gub.PC(fr); pc >= 0 {
			switch s := (*gub.Instr).(type) {
			case *ssa2.Trace:
				syntax = s.Syntax()
			}
		}
	}
	if syntax != nil {
		ast.Print(fn.Prog.Fset, syntax)
		fmt.Println("");
	} else {
		gub.Msg("Sorry, we don't have an AST for this")
	}
}
