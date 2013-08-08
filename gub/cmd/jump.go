// Copyright 2013 Rocky Bernstein.


package gubcmd

import "github.com/rocky/ssa-interp/gub"

func init() {
	name := "jump"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: JumpCommand,
		Help: `jump *num*

Jumps to instruction *num* inside the current basic block.
`,
		Min_args: 1,
		Max_args: 1,
	}
	gub.AddToCategory("running", name)
}

func JumpCommand(args []string) {
	fr := gub.CurFrame()
	b := fr.Block()
	ic, err := gub.GetUInt(args[1],
		"instruction number", 0, uint64(len(b.Instrs)-1))
	if err != nil { return }
	// compensate for interpreter loop which does ic++ at end of loop body
	fr.SetPC(uint(ic-1))
	gub.InCmdLoop = false
}
