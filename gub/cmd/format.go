// Copyright 2014-2015 Rocky Bernstein.
// format command

package gubcmd

import (
	"os"
	"fmt"
	"go/ast"
	"go/format"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "format"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: FormatCommand,
		Help: `format [function | .]

Prints formatted source code from the AST. If a function is given, the
source for function is printed. If '." is given we use the current
function. Otherwise, we give the source code for the high level
statement that we are stopped at. This could be for example an
assignment statement, a for loop, a condition in a for loop, or
initializer in a for loop, to name just a few examples.

See also the "ast" command.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("files", name)
	gub.AddAlias("list", name)
	gub.AddAlias("l", name)
}

// FormatCommand implements the debugger command: format
//
// format
//
// Formats AST and produces source text for function.
// FIXME: allow one to specify a function or package
func FormatCommand(args []string) {
	var syntax ast.Node
	var err error
	fr := gub.CurFrame()
	fn := fr.Fn()
	if len(args) > 1 {
		name := args[1]
		if name != "." {
			fn, err = gub.FuncLookup(name)
			if err != nil {
				gub.Errmsg(err.Error())
				return
			} else if fn == nil {
				gub.Errmsg("function '%s' not found", name)
				return
			}
		}
		syntax = fn.Syntax()
	} else {
		syntax = fn.Syntax()
		switch s := (*gub.Instr).(type) {
		case *ssa2.Trace:
			syntax = s.Syntax()
		}
	}
	if syntax != nil {
		// FIXME: use gub.Msg, not stdout
		format.Node(os.Stdout, fn.Prog.Fset, syntax)
		fmt.Println("");
	} else {
		gub.Msg("Sorry, we don't have an AST for this")
	}
}
