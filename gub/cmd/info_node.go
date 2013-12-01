// Copyright 2013 Rocky Bernstein.

// info scope [level]
//
// Prints information about scope

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"go/ast"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoNodeSubcmd,
		Help: `info node

Prints information about the node of the current scope.
Warning: this can be volumnous.
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "AST Node information",
		Name: "node",
	})
}

func InfoNodeSubcmd(args []string) {
	fr    := gub.CurFrame()
	scope := fr.Scope()
	if scope == nil {
		gub.Errmsg("No scope recorded here")
		return
	}
	if scope.Node() != nil {
		ast.Print(fr.Fset(), scope.Node())
	}
}
