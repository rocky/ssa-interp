// Copyright 2013-2014 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "locals"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: LocalsCommand,
		Help: `locals [*name*]

show local variable information. If *name* is not given list
all local variables.

See also "globals", "whatis", and "eval".
`,
		Min_args: 0,
		Max_args: 2,
	}
	gub.AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	gub.Aliases["local"] = name
	gub.Aliases["loc"] = name
}

// LocalsCommand implements the debugger command:
//    locals [*name*]
// which shows local variable info.
//
// See also "globals", "whatis", and "eval".
func LocalsCommand(args []string) {
	argc := len(args) - 1
	fr := gub.CurFrame()
	if argc == 0 {
		for i, _ := range fr.Locals() {
			gub.PrintLocal(fr, uint(i))
		}
		for reg, v := range fr.Reg2Var {
			gub.Msg("reg %s, var %s", reg, v)
		}
	} else {
		varname := args[1]
		if gub.PrintIfLocal(fr, varname) {
			return
		}
		// FIXME: This really shouldn't be needed.
		for i, v := range fr.Locals() {
			ssaVal := fr.Fn().Locals[i]
			if varname == ssaVal.Name() {
				gub.Msg("fixme %s %s: %s",
					varname, fr.Fn().Locals[i], interp.ToInspect(v, nil))
				break
			}
		}

	}
}
