// Copyright 2013 Rocky Bernstein.

// info block
//
// Prints block number

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoBlockSubcmd,
		Help: `info block
Prints basic block number`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "Basic block number",
		Name: "block",
	})
}

func InfoBlockSubcmd(args []string) {
	block := gub.CurBlock()
	// if block == nil && gub.Instr.Block() != nil {
	// 	block = gub.Instr.Block()
	// }
	if block == nil {
		gub.Msg("unknown block")
	} else {
		gub.Msg("basic block: %d", block.Index)
		if block.Scope != nil {
			gub.Msg("scope: %d", block.Scope.ScopeId())
		}
	}
}
