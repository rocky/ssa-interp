// Copyright 2013 Rocky Bernstein.

// scope [level]
//
// Prints information about scope

package gub

import (
	"go/ast"
)

func init() {
	name := "scope"
	cmds[name] = &CmdInfo{
		fn: ScopeCommand,
		help: `quit [level]

Prints information about the scope for the current stack frame.
If a level is given, we go up that many levels.
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("data", name)
}

func ScopeCommand(args []string) {
	scope := curFrame.Scope()
	if scope == nil {
		errmsg("No scope recorded here")
		return
	}
	section("scope number %d", scope.ScopeNum())
	msg("%s", scope.Scope)
	ast.Print(curFrame.Fset(), scope.Scope.Node())
}
