// Copyright 2013 Rocky Bernstein.

// quit command
// quit [exit-code]
//
// Terminates program. If an exit code is given, that is the exit code
// for the program. Zero (normal termination) is used if no
// termintation code.

package gubcmd

import (
	"os"
	"strconv"
	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "quit"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: QuitCommand,
		Help: `quit [exit-code]

Terminates program. If an exit code is given, that is the exit code
for the program. Zero (normal termination) is used if no
termintation code.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("support", name)
	gub.Aliases["exit"] = name
	// Down the line we'll have abbrevs
	gub.Aliases["q"] = name
}

func QuitCommand(args []string) {
	rc := 0
	if len(args) == 2 {
		new_rc, ok := strconv.Atoi(args[1])
		if ok == nil { rc = new_rc } else {
			gub.Errmsg("Expecting integer return code; got %s. Ignoring",
				args[1])
		}
	}
	gub.Msg("gub: That's all folks...")

	// FIXME: determine under which conditions we've used term
	gnureadline.Rl_reset_terminal(gub.Term)

	os.Exit(rc)

}
