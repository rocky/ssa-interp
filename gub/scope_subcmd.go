// Copyright 2013 Rocky Bernstein.

// info scope [level]
//
// Prints information about scope

package gub

import (
	"go/ast"
)

func init() {
	parent := "info"
	subcmds[parent] = &SubcmdInfo{
		fn: InfoScopeSubcmd,
		help: `info scope [level]

Prints information about the scope for the current stack frame.
If a level is given, we go up that many levels.
`,
		min_args: 1,
		max_args: 2,
	}
}

func InfoScopeSubcmd(args []string) {
	scope := curFrame.Scope()
	if scope == nil {
		Errmsg("No scope recorded here")
		return
	}
	section("scope number %d", scope.ScopeNum())
	Msg("%s", scope.Scope)
	ast.Print(curFrame.Fset(), scope.Scope.Node())
}
