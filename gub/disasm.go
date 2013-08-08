package gub

import (
	"github.com/rocky/ssa-interp"
)

func DisasmInst(f *ssa2.Function, bnum int, inst uint64) {
	if bnum < 0 || bnum >= len(f.Blocks) {
		Errmsg("Block number %d is out of range. Should be between 0..%d",
			bnum, len(f.Blocks)-1)
		return
	}
	b := f.Blocks[bnum]
	if b == nil {
		// Corrupt CFG.
		Msg(".nil:")
		return
	}
	Msg("%3d: %s",  inst, ssa2.DisasmInst(b.Instrs[inst], Maxwidth))
}

func DisasmBlock(f *ssa2.Function, i int) {
	if i < 0 || i >= len(f.Blocks) {
		Errmsg("Block number %d is out of range. Should be between 0..%d",
			i, len(f.Blocks)-1)
		return
	}
	b := f.Blocks[i]
	if b == nil {
		// Corrupt CFG.
		Msg(".nil:")
		return
	}
	Msg(".%s:", b)
	for i, instr := range b.Instrs {
		Msg("%3d: %s",  i, ssa2.DisasmInst(instr, Maxwidth))
	}
}
