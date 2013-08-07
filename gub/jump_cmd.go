// Copyright 2013 Rocky Bernstein.


package gub

func init() {
	name := "jump"
	Cmds[name] = &CmdInfo{
		Fn: JumpCommand,
		Help: `jump *num*

Jumps to instruction *num* inside the current basic block.
`,
		Min_args: 1,
		Max_args: 1,
	}
	AddToCategory("running", name)
}

func JumpCommand(args []string) {
	fr := curFrame
	b := fr.Block()
	ic, err := GetInt(args[1],
		"instruction number", 0, len(b.Instrs)-1)
	if err != nil { return }
	// compensate for interpreter loop which does ic++ at end of loop body
	fr.SetPC(ic-1)
	inCmdLoop = false
}
