// Copyright 2013 Rocky Bernstein.

// info scope [level]
//
// Prints information about scope

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoScopeSubcmd,
		Help: `info scope [level]

Prints information about the scope for the current stack frame.
If a level is given, we go up that many levels.
`,
		Min_args: 0,
		Max_args: 1,
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
	count := 0
	if len(args) == 3 {
		var err error
		count, err = gub.GetInt(args[2],
			"count", 0, gub.MAXSTACKSHOW)
		if err != nil { return }
	}

	typescope := scope.Scope
	for i := 0; i < count; i++ {
		typescope = typescope.Parent()
		if typescope == nil {
			gub.Errmsg("There are only %d nested scopes", i)
			return
		}
		scope = fr.Fn().Pkg.TypeScope2Scope[typescope]
		if scope == nil {
			gub.Errmsg("No parent scope; There are only %d nested scopes", i)
			return
		}
	}
	gub.Section("scope number %d", scope.ScopeId())
	gub.Msg("%s", typescope)
}
