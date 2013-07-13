// Copyright 2013 Rocky Bernstein.

// Restarts program..

package gub

import (
	"os"
	"strings"
	"syscall"
)

func init() {
	name := "run"
	cmds[name] = &CmdInfo{
		fn: RunCommand,
		help: `run

Terminates program. If an exit code is given, that is the exit code
for the program. Zero (normal termination) is used if no
termintation code.
`,
		min_args: 0,
		max_args: 0,
	}
	AddToCategory("running", name)
	aliases["R"] = name
	aliases["restart"] = name
}

func RunCommand(args []string) {
	if GUB_RESTART_CMD == "" {
		errmsg("restart string in environment GUB_RESTART_CMD has nothing")
		return
	}
	msg("gub: restarting: %s", GUB_RESTART_CMD)
	restartCmd := strings.Split(GUB_RESTART_CMD, " ")
	syscall.Exec(restartCmd[0], restartCmd, os.Environ());
}
