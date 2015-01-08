// Copyright 2013-2015 Rocky Bernstein.
// disassemble command

package gubcmd

import (
	"os"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "disassemble"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: DisassembleCommand,
		Help: `disassemble [*fn* | *int* | . | + ]

disassemble SSA instructions. Without any parameters we disassemble the
current instruction. If a function name is given, that is disassembled.
If a number is given that is the block number of the current frame.
If "." is given we disassemble the current block only. If "+"
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("inspecting", name)
	gub.AddAlias("disasm", name)
}


func DisassembleCommand(args []string) {
	fr := gub.CurFrame()
	myfn := fr.Fn()
	if len(args) > 1 {
		what := args[1]
		if what == "." {
			if block := gub.CurBlock(); block != nil {
				gub.DisasmBlock(myfn, block.Index)
			} else {
				gub.Errmsg("Can't get block info here")
			}
			return
		} else if what != "+" {
			if fn, err := gub.FuncLookup(what); err == nil && fn != nil {
				myfn = fn
			} else {
				bnum, err := gub.GetInt(args[1],
					"block number of function name", 0, len(myfn.Blocks)-1)
				if err == nil {
					lastBlock := len(myfn.Blocks) - 1
					if bnum <= lastBlock {
						b := myfn.Blocks[bnum]
						if len(args) == 3 {
							ic, err := gub.GetUInt(args[2],
								"instruction number", 0, uint64(len(b.Instrs)-1))
							if err == nil {
								gub.DisasmInst(myfn, bnum, ic)
							}
						} else {
						gub.DisasmBlock(myfn, bnum)
						}
					} else {
						gub.Errmsg("Block number should be between 0 and %d; got %d",
							lastBlock, bnum)
					}
				}
				return
			}
		}
	} else {
		gub.DisasmCurrentInst()
		return
	}
	myfn.WriteTo(os.Stderr)
}
