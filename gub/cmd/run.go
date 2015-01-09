// Copyright 2013 Rocky Bernstein.

// Restarts program..

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"os"
	"syscall"
)

func init() {
	name := "run"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: RunCommand,
		Help: `run

Terminates program. If an exit code is given, that is the exit code
for the program. Zero (normal termination) is used if no
termintation code.
`,
		Min_args: 0,
		Max_args: 0,
	}
	gub.AddToCategory("running", name)
	gub.Aliases["R"] = name
	gub.Aliases["restart"] = name
}

func RunCommand(args []string) {
	ShowArgsSubcmd(args)
	gub.Msg("gub: restarting...")
	syscall.Exec(gub.RESTART_ARGS[0], gub.RESTART_ARGS[0:], os.Environ());
}
