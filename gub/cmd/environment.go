// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)


func init() {
	name := "environment"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: EnvironmentCommand,
		Help: `environment [*name*]

print current runtime environment values.
If *name* is supplied, only show that name.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	gub.Aliases["env"] = name
	gub.Aliases["environ"] = name
}

func EnvironmentCommand(args []string) {
	if len(args) == 2 {
		gub.PrintInEnvironment(gub.CurFrame(), args[1])
		return
	}
	for k, v := range gub.CurFrame().Env() {
		switch k := k.(type) {
		case *ssa2.Alloc:
			if scope := k.Scope; scope != nil {
				gub.Msg("%s: %s = %s (scope %d)", k.Name(), k, gub.Deref2Str(v),
					scope.ScopeId())
			} else {
				gub.Msg("%s: %s = %s", k.Name(), k, gub.Deref2Str(v))
			}
		default:
			gub.Msg("%s: %s = %s", k.Name(), k, gub.Deref2Str(v))
		}
	}
}
