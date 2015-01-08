// Copyright 2013-2015 Rocky Bernstein.

// info program
//
// Prints program information

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoProgramSubcmd,
		Help: `info program

Prints information about the program including:
*  instruction number
*  block number
*  function number
*  stop event
*  source-code position
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "Information about debugged program",
		Name: "program",
	})
}

// InfoProgramSubcmd implements the debugger command:
//   info program
// This command prints information about the program including:
//    instruction number
//    block number
//    function number
//    stop event
//    source-code position
func InfoProgramSubcmd(args []string) {
	if gub.TraceEvent == ssa2.PROGRAM_TERMINATION {
		gub.Msg("program stop event: %s",
			ssa2.Event2Name[gub.TraceEvent])
		return
	}

	fr := gub.CurFrame()
	pc := fr.PC()
 	gub.Msg("instruction number: %d", pc)
	block := fr.Block()
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
	fn := fr.Fn()
	if fn.Signature.Recv() != nil {
		gub.Msg("function: %s", fr.FnAndParamString())
	} else {
		gub.Msg("function: %s.%s", fn.Pkg.Object.Path(), fr.FnAndParamString())
	}
	gub.Msg("program stop event: %s", ssa2.Event2Name[gub.TraceEvent])
	gub.Msg("position: %s", gub.CurFrame().PositionRange())
}
