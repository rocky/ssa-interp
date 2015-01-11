// Copyright 2013, 2015 Rocky Bernstein.

package gubcmd

import (
	"fmt"

	"go/token"
	"github.com/rocky/go-exact"
	"github.com/rocky/go-types"

	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "globals"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: GlobalsCommand,
		Help: `globals [*name*]

show global variable information. If *name* is not given list
all global variables.

See also "locals", "whatis", and "eval".
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("global", name)
	gub.AddAlias("gl", name)
}

// GlobalsCommand implements the debugger command:
//    globals [*name*]
// which shows global variable info.
//
// See also "locals", "whatis", and "eval".
func GlobalsCommand(args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for k, v := range gub.CurFrame().I().Globals() {
			if v == nil {
				fmt.Printf("%s: nil\n")
			} else {
				// FIXME: figure out why reflect.lookupCache causes
				// an panic on a nil pointer or invalid address
				if fmt.Sprintf("%s", k) == "reflect.lookupCache" {
					continue
				}
				gub.Msg("%s: %s", k, interp.ToInspect(*v, &k))
			}
		}
	} else {
		// This doesn't work and I don't know how to fix it.
		for i:=1; i<=argc; i++ {
			vv := ssa2.NewConst(exact.MakeString(args[i]),
				types.Typ[types.String], token.NoPos, token.NoPos)
			// fmt.Println(vv, "vs", interp.ToInspect(vv))
			v, ok := gub.CurFrame().I().Globals()[vv]
			if ok {
				gub.Msg("%s: %s", vv, interp.ToInspect(*v, nil))
			}
		}

		// This is ugly, but I don't know how to turn a string into
		// a ssa2.Value.
		globals := make(map[string]*interp.Value)
		for k, v := range gub.CurFrame().I().Globals() {
			globals[fmt.Sprintf("%s", k)] = v
		}

		for i:=1; i<=argc; i++ {
			vv := args[i]
			v, ok := globals[vv]
			if ok {
				gub.Msg("%s: %s", vv, interp.ToInspect(*v, nil))
			}
		}
	}
}
