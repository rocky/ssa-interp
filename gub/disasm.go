// Copyright 2014-2015 Rocky Bernstein.
package gub

import (
	"github.com/rocky/ssa-interp"
)

func DisasmPrefix(block *ssa2.BasicBlock) bool {
	if block == nil {
		Msg(":.nil:")
		return false
	} else if block.Scope != nil {
		Section("# scope %d", block.Scope.ScopeId())
	}
	Section("Block .%s:", block)
	return true
}

func DisasmCurrentInst() {
	DisasmPrefix(curBlock)
	Msg("%3d: %s",  curFrame.PC(), ssa2.DisasmInst(*Instr, Maxwidth))
}

func DisasmInst(f *ssa2.Function, bnum int, inst uint64) {
	if bnum < 0 || bnum >= len(f.Blocks) {
		Errmsg("Block number %d is out of range. Should be between 0..%d",
			bnum, len(f.Blocks)-1)
		return
	}
	if 	b := f.Blocks[bnum]; DisasmPrefix(b) {
		Msg("%3d: %s",  inst, ssa2.DisasmInst(b.Instrs[inst], Maxwidth))
	}
}

func DisasmBlock(f *ssa2.Function, i int, pc int) {
	if i < 0 || i >= len(f.Blocks) {
		Errmsg("Block number %d is out of range. Should be between 0..%d",
			i, len(f.Blocks)-1)
		return
	}
	if b := f.Blocks[i]; DisasmPrefix(b) {
		for i, instr := range b.Instrs {
			prefix := "  "
			if i == pc { prefix = "=>" }
			Msg("%s%3d: %s",  prefix, i, ssa2.DisasmInst(instr, Maxwidth))
		}
	}
}
