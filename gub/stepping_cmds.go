// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution
package gub

import (
	"fmt"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "step"
	Cmds[name] = &CmdInfo{
		Fn: StepCommand,
		Help: `step

Execute the current line, stopping at the next event.  Sometimes this
is called 'step into'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("s", name)
}

func StepCommand(args []string) {
	fmt.Println("Stepping...")
	interp.SetStepIn(curFrame)
	InCmdLoop = false
}
