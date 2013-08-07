// Copyright 2013 Rocky Bernstein.
package gub

import (
	"fmt"

	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/types"


	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "globals"
	Cmds[name] = &CmdInfo{
		Fn: GlobalsCommand,
		Help: "global [*name*]: show global variable info",
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	Aliases["global"] = name
	Aliases["gl"] = name
}

func GlobalsCommand(args []string) {
	argc := len(args) - 1
	if argc == 0 {
		for k, v := range curFrame.I().Globals() {
			if v == nil {
				fmt.Printf("%s: nil\n")
			} else {
				// FIXME: figure out why reflect.lookupCache causes
				// an panic on a nil pointer or invalid address
				if fmt.Sprintf("%s", k) == "reflect.lookupCache" {
					continue
				}
				Msg("%s: %s", k, interp.ToInspect(*v))
			}
		}
	} else {
		// This doesn't work and I don't know how to fix it.
		for i:=1; i<=argc; i++ {
			vv := ssa2.NewConst(exact.MakeString(args[i]),
				types.Typ[types.String], token.NoPos, token.NoPos)
			// fmt.Println(vv, "vs", interp.ToInspect(vv))
			v, ok := curFrame.I().Globals()[vv]
			if ok {
				Msg("%s: %s", vv, interp.ToInspect(*v))
			}
		}

		// This is ugly, but I don't know how to turn a string into
		// a ssa2.Value.
		globals := make(map[string]*interp.Value)
		for k, v := range curFrame.I().Globals() {
			globals[fmt.Sprintf("%s", k)] = v
		}

		for i:=1; i<=argc; i++ {
			vv := args[i]
			v, ok := globals[vv]
			if ok {
				Msg("%s: %s", vv, interp.ToInspect(*v))
			}
		}
	}
}
