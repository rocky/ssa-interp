// Copyright 2013 Rocky Bernstein.

// info scope [level]
//
// Prints information about scope

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	// "go/ast"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoScopeSubcmd,
		Help: `info scope [level]

Prints information about the scope for the current stack frame.
If a level is given, we go up that many levels.
`,
		Min_args: 1,
		Max_args: 2,
		Short_help: "Scope information",
		Name: "scope",
	})
}

func InfoScopeSubcmd(args []string) {
	fr    := gub.CurFrame()
	scope := fr.Scope()
	if scope == nil {
		gub.Errmsg("No scope recorded here")
		return
	}
	gub.Section("scope number %d", scope.ScopeId())
	gub.Msg("%s", scope.Scope)
	// FIXME: reinstate
	// ast.Print(fr.Fset(), scope.Scope.Node())
}
