// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "continue"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: ContinueCommand,
		Help: `continue

Leave the debugger loop and continue execution. Subsequent entry to
the debugger however may occur via breakpoints or explicit calls, or
exceptions.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddAlias("c", name)
	gub.AddToCategory("running", name)
}

func ContinueCommand(args []string) {
	for fr := gub.TopFrame(); fr != nil; fr = fr.Caller(0) {
		interp.SetStepOff(fr)
	}
	gub.InCmdLoop = false
	gub.Msg("Continuing...")
}
