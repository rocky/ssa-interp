// Copyright 2013-2015 Rocky Bernstein.

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
and block number. If we are at a call, before the first instruction,
-1 is printed. If we are at a return, after the last instruction,
-2 is printed.
`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "program counter",
		Name: "pc",
	})
}

func InfoPCSubcmd(args []string) {
	fr := gub.CurFrame()
	pc := fr.PC()
	fn := fr.FnAndParamString()
	if block := gub.CurBlock(); block != nil {
		gub.Msg("instruction number: %d of block %d, function %s",
			pc, block.Index, fn)
	} else if pc == -2 {
		gub.Msg("instruction number: %d (at return), function %s", pc, fn)
	} else {
		gub.Msg("instruction number: %d, function %s", pc, fn)
	}
}
