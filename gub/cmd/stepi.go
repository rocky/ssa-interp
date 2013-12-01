// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution
package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "stepi"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: StepInstructionCommand,
		Help: `stepi

Execute the current line, stopping at the instruction.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	// Down the line we'll have abbrevs
}

func StepInstructionCommand(args []string) {
	gub.Msg("Stepping Instruction...")
	interp.SetStepInstruction(gub.CurFrame())
	gub.InCmdLoop = false
}
