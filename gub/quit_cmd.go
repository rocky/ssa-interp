// Copyright 2013 Rocky Bernstein.

// quit command
// quit [exit-code]
//
// Terminates program. If an exit code is given, that is the exit code
// for the program. Zero (normal termination) is used if no
// termintation code.

package gub

import (
	"os"
	"strconv"
	"code.google.com/p/go-gnureadline"
)

func init() {
	name := "quit"
	Cmds[name] = &CmdInfo{
		Fn: QuitCommand,
		Help: `quit [exit-code]

Terminates program. If an exit code is given, that is the exit code
for the program. Zero (normal termination) is used if no
termintation code.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("support", name)
	aliases["exit"] = name
	// Down the line we'll have abbrevs
	aliases["q"] = name
}

func QuitCommand(args []string) {
	rc := 0
	if len(args) == 2 {
		new_rc, ok := strconv.Atoi(args[1])
		if ok == nil { rc = new_rc } else {
			Errmsg("Expecting integer return code; got %s. Ignoring",
				args[1])
		}
	}
	Msg("gub: That's all folks...")

	// FIXME: determine under which conditions we've used term
	gnureadline.Rl_reset_terminal(term)

	os.Exit(rc)

}
