// Copyright 2013-2014 Rocky Bernstein.

// info pc
//
// Prints information about current PC, an instruction counter

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoPCSubcmd,
		Help: `info pc

Prints information about the current PC, an instruction counter
and block number.
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "program counter",
		Name: "pc",
	})
}

func InfoPCSubcmd(args []string) {
	fr := gub.CurFrame()
	gub.Msg("instruction number: %d of block %d, function %s",
		fr.PC(), fr.Block().Index, fr.Fn().Name())
}
