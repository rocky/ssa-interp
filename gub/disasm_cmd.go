// Copyright 2013 Rocky Bernstein.
// disassemble command

package gub

import (
	"os"
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


func DisasmInst(f *ssa2.Function, bnum int, inst int) {
	if bnum < 0 || bnum >= len(f.Blocks) {
		errmsg("Block number %d is out of range. Should be between 0..%d",
			bnum, len(f.Blocks)-1)
		return
	}
	b := f.Blocks[bnum]
	if b == nil {
		// Corrupt CFG.
		msg(".nil:")
		return
	}
	msg("%3d: %s",  inst, ssa2.DisasmInst(b.Instrs[inst], maxwidth))
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
		pkg  := myfn.Pkg
		if fn := pkg.Func(what); fn != nil {
			myfn = fn
		} else {
			bnum, err := getInt(args[1],
				"block number", 0, len(myfn.Blocks)-1)
			if err == nil {
				if len(args) == 3 {
					ic, err := getInt(args[1],
						"instruction number", 0, len(myfn.Blocks)-1)
					if err == nil {
						DisasmInst(myfn, bnum, ic)
					}
				} else {
					DisasmBlock(myfn, bnum)
				}
			} else {
				errmsg("Can't find function %s", what)
			}
			return
		}
	}
	myfn.DumpTo(os.Stderr)
}
