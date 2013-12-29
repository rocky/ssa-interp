// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoArgsSubcmd,
		Help: `info args [arg-name]
Show argument variables of the current stack frame"`,
		Min_args: 0,
		Max_args: 1,
		Short_help: "Show argument variables of the current stack frame",
		Name: "args",
	})
}

// InfoArgsSubcmd implements the debugger command:
//    info args [arg-name]
// which shows argument variables of the current stack frame.
func InfoArgsSubcmd(args []string) {
	fr := gub.CurFrame()
	fn := fr.Fn()
	if len(args) == 2 {
		if len(fn.Params) == 0 {
			gub.Msg("Function `%s()' has no parameters", fn.Name())
			return
		}
		for i, p := range fn.Params {
			gub.Msg("%s %s", fn.Params[i], interp.ToInspect(fr.Env()[p]))
		}
	} else {
		varname := args[2]
		for i, p := range fn.Params {
			if varname == fn.Params[i].Name() {
				gub.Msg("%s %s", fn.Params[i], interp.ToInspect(fr.Env()[p]))
				break
			}
		}
	}
}
