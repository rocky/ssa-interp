// Copyright 2013 Rocky Bernstein.

package gubcmd

import "github.com/rocky/ssa-interp/gub"

func init() {
	name := "backtrace"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: BacktraceCommand,
		Help: `backtrace [*count*]

Print a stack trace, with the most recent frame at the top.

With a positive number, print at most many entries.`,

		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("stack", name)
	gub.Aliases["where"] = name
	gub.Aliases["T"] = name  // for perl5db hackers
	// Down the line we'll have abbrevs
	gub.Aliases["bt"] = name
}

func BacktraceCommand(args []string) {
	count := gub.MAXSTACKSHOW
	var err error
	if len(args) > 1 {
		count, err = gub.GetInt(args[1], "maximum count",
			0, gub.MAXSTACKSHOW)
		if err != nil { return }
	}
	gub.PrintStack(gub.TopFrame(), count)
}
