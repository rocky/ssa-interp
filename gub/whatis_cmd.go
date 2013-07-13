// Copyright 2013 Rocky Bernstein.
package gub

func init() {
	name := "whatis"
	cmds[name] = &CmdInfo{
		fn: WhatisCommand,
		help: `whatis *name*

print information about *name* which can include a dotted variable name.
`,
		min_args: 1,
		max_args: 1,
	}
	AddToCategory("inspecting", name)
}

func WhatisCommand(args []string) {
	name := args[1]
	WhatisName(name)
}
