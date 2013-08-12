// Copyright 2013 Rocky Bernstein.

package gubcmd

import "github.com/rocky/ssa-interp/gub"

func init() {
	name := "frame"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: FrameCommand,
		Help: `frame *num*

Change the current frame to frame *num*

See also 'up' and 'down'.
`,
		Min_args: 1,
		Max_args: 1,
	}
	gub.AddToCategory("stack", name)
}

func FrameCommand(args []string) {
	i, err := gub.GetInt(args[1],
		"frame number", -gub.MAXSTACKSHOW, gub.MAXSTACKSHOW)
	if err != nil { return }
	gub.AdjustFrame(i, true)
}
