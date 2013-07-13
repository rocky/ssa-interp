// Copyright 2013 Rocky Bernstein.
// disassemble command

package gub

import (
	"os"
	"strconv"
	"github.com/rocky/ssa-interp"
)

func init() {
	name := "disassemble"
	cmds[name] = &CmdInfo{
		fn: DisassembleCommand,
		help: `disasm [*fn* | *int* | . ]

disassemble SSA instructions. Without any parameters we disassemble the
entire current function. If a function name is given, that is disassembled.
If a number is given that is the block number of the current frame.
If "." is given we disassemble the current block only.
`,
		min_args: 0,
		max_args: 2,
	}
	AddToCategory("inspecting", name)
	AddAlias("disasm", name)
}


func DisasmInst(f *ssa2.Function, i int, inst int) {
	if i < 0 || i >= len(f.Blocks) {
		errmsg("Block number %d is out of range. Should be between 0..%d",
			i, len(f.Blocks)-1)
		return
	}
	b := f.Blocks[i]
	if b == nil {
		// Corrupt CFG.
		msg(".nil:")
		return
	}
	msg("%3d: %s",  i, ssa2.DisasmInst(b.Instrs[i], maxwidth))
}

func DisasmBlock(f *ssa2.Function, i int) {
	if i < 0 || i >= len(f.Blocks) {
		errmsg("Block number %d is out of range. Should be between 0..%d",
			i, len(f.Blocks)-1)
		return
	}
	b := f.Blocks[i]
	if b == nil {
		// Corrupt CFG.
		msg(".nil:")
		return
	}
	msg(".%s:", b)
	for i, instr := range b.Instrs {
		msg("%3d: %s",  i, ssa2.DisasmInst(instr, maxwidth))
	}
}


func DisassembleCommand(args []string) {
	myfn := curFrame.Fn()
	if len(args) > 1 {
		what := args[1]
		if what == "." {
			DisasmBlock(myfn, curFrame.Block().Index)
			return
		}
		if i, ok := strconv.Atoi(what); ok == nil {
			DisasmBlock(myfn, i)
			return
		}
		pkg  := myfn.Pkg
		if fn := pkg.Func(what); fn != nil {
			myfn = fn
		} else {
			errmsg("Can't find function %s", what)
			return
		}
	}
	myfn.DumpTo(os.Stderr)
}
