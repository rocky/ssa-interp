// Copyright 2013 Rocky Bernstein.

package gubcmd

import "github.com/rocky/ssa-interp/gub"

func init() {
	name := "down"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: DownCommand,
		Help: `down [*count*]

Move the current frame down in the stack trace (to a newer frame). 0
is the most recent frame. If no count is given, move down 1.

See also 'up' and 'frame'.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("stack", name)
}

func DownCommand(args []string) {
	count := 1
	if len(args) == 2 {
		var err error
		count, err = gub.GetInt(args[1],
			"count", -gub.MAXSTACKSHOW, gub.MAXSTACKSHOW)
		if err != nil { return }
	}
	gub.AdjustFrame(-count, false)
}
