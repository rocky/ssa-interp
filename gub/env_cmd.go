// Copyright 2013 Rocky Bernstein.

package gub

import "github.com/rocky/ssa-interp"

func init() {
	name := "environment"
	Cmds[name] = &CmdInfo{
		Fn: EnvironmentCommand,
		Help: `environment [*name*]

print current runtime environment values.
If *name* is supplied, only show that name.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	Aliases["env"] = name
	Aliases["environ"] = name
}

func EnvironmentCommand(args []string) {
	if len(args) == 2 {
		printInEnvironment(curFrame, args[1])
		return
	}
	for k, v := range curFrame.Env() {
		switch k := k.(type) {
		case *ssa2.Alloc:
			if scope := k.Scope; scope != nil {
				Msg("%s: %s = %s (scope %d)", k.Name(), k, deref2Str(v),
					scope.ScopeNum())
			} else {
				Msg("%s: %s = %s", k.Name(), k, deref2Str(v))
			}
		default:
			Msg("%s: %s = %s", k.Name(), k, deref2Str(v))
		}
	}
}
