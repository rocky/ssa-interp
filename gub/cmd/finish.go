// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "finish"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: FinishCommand,
		Help: `finish

Continue execution until the program is about to:

* leave the current function, or
* switch context via yielding back or finishing a block which was
  yielded to.

Sometimes this is called 'step out'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("fin", name)
}

func FinishCommand(args []string) {
	interp.SetStepOut(gub.TopFrame())
	gub.Msg("Continuing until return...")
	gub.InCmdLoop = false
}
