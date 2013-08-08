// Copyright 2013 Rocky Bernstein.
package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "locals"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: LocalsCommand,
		Help: "locals [*name*]: show local variable info",
		Min_args: 0,
		Max_args: 2,
	}
	gub.AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	gub.Aliases["local"] = name
	gub.Aliases["loc"] = name
}

func LocalsCommand(args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for i, _ := range gub.CurFrame().Locals() {
			gub.PrintLocal(gub.CurFrame(), uint(i))
		}
	} else {
		varname := args[1]
		if gub.PrintIfLocal(gub.CurFrame(), varname) {
			return
		}
		// FIXME: This really shouldn't be needed.
		for i, v := range gub.CurFrame().Locals() {
			if varname == gub.CurFrame().Fn().Locals[i].Name() {
				gub.Msg("fixme %s %s: %s",
					varname, gub.CurFrame().Fn().Locals[i], interp.ToInspect(v))
				break
			}
		}

	}
}
