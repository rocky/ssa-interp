// Copyright 2013 Rocky Bernstein.
// quit command

package gubcmd

import (
	"os"
	"fmt"
	"go/format"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "format"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: FormatCommand,
		Help: `ast

Prints AST for current function
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("files", name)
}

// FormatCommand implements the debugger command: format
//
// format
//
// Formats AST and produces source text for function.
func FormatCommand(args []string) {
	fn := gub.CurFrame().Fn()
	if syntax := fn.AST(); syntax != nil {
		// FIXME: use gub.Msg, not stdout
		format.Node(os.Stdout, fn.Prog.Fset, syntax)
		fmt.Println("");
	} else {
		gub.Msg("Sorry, we don't have an AST for %s", fn);
	}
}
