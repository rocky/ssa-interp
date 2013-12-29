// Copyright 2013 Rocky Bernstein.

package gubcmd
import "github.com/rocky/ssa-interp/gub"


func init() {
	name := "whatis"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: WhatisCommand,
		Help: `whatis *name*

print information about *name* which can include a dotted variable name.
`,
		Min_args: 1,
		Max_args: 1,
	}
	gub.AddToCategory("inspecting", name)
}

// WhatisCommand implements the debugger command:
//    whatis *name*
// which desribes what *name* is.
//
// See also "locals", "globals", and "eval".
func WhatisCommand(args []string) {
	name := args[1]
	gub.WhatisName(name)
}
