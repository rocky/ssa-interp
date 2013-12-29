// Copyright 2013 Rocky Bernstein.

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

Execute one SSA instrcution and stop.

See also step, and next.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	// Down the line we'll have abbrevs
}

// StepInstructionCommand implements the debugger command:
//   stepi
// which executes one SSA instrcution and stop.
//
// See also "step", "next", "continue" and "finish".
func StepInstructionCommand(args []string) {
	gub.Msg("Stepping Instruction...")
	interp.SetStepInstruction(gub.CurFrame())
	gub.InCmdLoop = false
}
