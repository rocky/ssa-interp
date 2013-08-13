// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "next"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: NextCommand,
		Help: `next

Step one statement ignoring steps into function calls at this level.

Sometimes this is called 'step over'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("n", name)
}

func NextCommand(args []string) {
	interp.SetStepOver(gub.TopFrame())
	gub.Msg("Step over...")
	gub.InCmdLoop = false
}
