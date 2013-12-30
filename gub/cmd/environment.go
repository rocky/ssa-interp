// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
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
		name := args[1]
		nameVal, interpVal, scopeVal := gub.EnvLookup(gub.CurFrame(), name, gub.CurScope())
		if nameVal != nil {
			gub.PrintInEnvironment(gub.CurFrame(), nameVal, interpVal, scopeVal)
		} else {
			gub.Errmsg("%s is in not the environment", name)
		}
		return
	}
	for nameVal, interpVal := range gub.CurFrame().Env() {
		gub.PrintInEnvironment(gub.CurFrame(), nameVal, interpVal, gub.CurScope())
	}
}
