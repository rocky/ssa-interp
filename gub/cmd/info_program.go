// Copyright 2013 Rocky Bernstein.

// info scope [level]
//
// Prints information about scope

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

// import (
// 	"go/ast"
// )

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoProgramSubcmd,
		Help: `info program

Prints information about the program including:
*  instruction number
*  block number
*  stop event
*  source-code position
`,
		Min_args: 1,
		Max_args: 1,
		Short_help: "Program information",
		Name: "program",
	})
}

func InfoProgramSubcmd(args []string) {
	gub.Msg("instruction number: %d", gub.CurFrame().PC())
	block := gub.CurFrame().Block()
	if block == nil {
		gub.Msg("unknown block")
	} else {
		gub.Msg("basic block: %d", block.Index)
		if block.Scope != nil {
			gub.Msg("scope: %d", block.Scope.ScopeId())
		} else {
			gub.Msg("unknown scope")
		}
	}
	gub.Msg("program stop event: %s", ssa2.Event2Name[gub.TraceEvent])
	gub.Msg("position: %s", gub.CurFrame().PositionRange())
}
