// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "step"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: StepCommand,
		Help: `step

Execute the current statement, stopping at the next event.  Sometimes this
is called 'step into'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("s", name)
}

// StepCommand implements the debugger command: step which executes
// the current statement, stopping at the next event.  Sometimes this
// is called 'step into'.
//
// See also stepi, continue, finish, and next.
func StepCommand(args []string) {
	gub.Msg("Stepping...")
	interp.SetStepIn(gub.CurFrame())
	gub.InCmdLoop = false
}
