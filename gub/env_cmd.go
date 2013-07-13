// Copyright 2013 Rocky Bernstein.

package gub

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func EnvLookup(fr *interp.Frame, name string) ssa2.Value {
	for k := range curFrame.Env() {
		if name == k.Name() {
			return k
		}
	}
	return nil
}

func init() {
	name := "environment"
	cmds[name] = &CmdInfo{
		fn: EnvironmentCommand,
		help: `environment [*name*]

print current runtime environment values.
If *name* is supplied, only show that name.
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	aliases["env"] = name
	aliases["environ"] = name
}

func EnvironmentCommand(args []string) {
	if len(args) == 2 {
		name := args[1]
		if k := EnvLookup(curFrame, name); k != nil {
			msg("%s: %s = %s", name, k, derefValue(curFrame.Env()[k]))
		} else {
			errmsg("Name %s not found in environment", name)
		}
		return
	}
	for k, v := range curFrame.Env() {
		msg("%s: %s = %s", k.Name(), k, derefValue(v))
	}
}
