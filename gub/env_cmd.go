// Copyright 2013 Rocky Bernstein.

package gub

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
		printInEnvironment(curFrame, args[1])
		return
	}
	for k, v := range curFrame.Env() {
		msg("%s: %s = %s", k.Name(), k, deref2Str(v))
	}
}
