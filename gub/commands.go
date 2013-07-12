// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"fmt"

	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/types"


	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

type CmdFunc func([]string)

type CmdInfo struct {
	help string
	category string
	min_args int
	max_args int
	fn CmdFunc
	aliases []string
}

var cmds map[string]*CmdInfo  = make(map[string]*CmdInfo)
var	aliases map[string]string = make(map[string]string)
var	categories map[string] []string = make(map[string] []string)

func AddAlias(alias string, cmdname string) bool {
	if unalias := aliases[alias]; unalias != "" {
		return false
	}
	aliases[alias] = cmdname
	cmds[cmdname].aliases = append(cmds[cmdname].aliases, alias)
	return true
}

func AddToCategory(category string, cmdname string) {
	categories[category] = append(categories[category], cmdname)
	// cmds[cmdname].category = category
}


func lookupCmd(cmd string) (string) {
	if cmds[cmd] == nil {
		cmd = aliases[cmd];
	}
	return cmd
}

func init() {
	cmds["globals"] = &CmdInfo{
		fn: GlobalsCommand,
		help: "global [*name*]: show global variable info",
		min_args: 0,
		max_args: 1,
	}
	// Down the line we'll have abbrevs
	aliases["global"] = "globals"
	aliases["gl"] = "globals"
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
				msg("%s: %s", k, interp.ToString(*v))
			}
		}
	} else {
		// This doesn't work and I don't know how to fix it.
		for i:=1; i<=argc; i++ {
			vv := ssa2.NewLiteral(exact.MakeString(args[i]),
				types.Typ[types.String], token.NoPos, token.NoPos)
			// fmt.Println(vv, "vs", interp.ToString(vv))
			v, ok := curFrame.I().Globals()[vv]
			if ok {
				msg("%s: %s", vv, interp.ToString(*v))
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
				msg("%s: %s", vv, interp.ToString(*v))
			}
		}
	}
}

func init() {
	name := "locations"
	cmds[name] = &CmdInfo{
		fn: LocsCommand,
		help: "show possible breakpoint locations",
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("status", name)
	// Down the line we'll have abbrevs
	AddAlias("locs", name)
}

func LocsCommand(args []string) {
	fn  := curFrame.Fn()
	pkg := fn.Pkg
	for _, l := range pkg.Locs() {
		// FIXME: ? turn into true range
		msg("\t%s", fmtPos(fn, l.Pos))
	}
}
