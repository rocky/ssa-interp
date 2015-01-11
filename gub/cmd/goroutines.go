// Copyright 2013, 2015 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "goroutines"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: GoroutinesCommand,
		Help: `goroutines [*id*]

Without a parameter, list stack traces for each active goroutine. If an id
is given only that goroutine stack trace is shown. The main (first) goroutine is 0.
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("stack", name)
	gub.AddAlias("gore", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("gor", name)
	gub.AddAlias("goroutine", name)
}

// shows stack of all goroutines
func GoroutinesCommand(args []string) {
	goTops := interp.GetInterpreter().GoTops()
	var goNum int
	var err error
	if len(args) > 1 {
		goNum, err = gub.GetInt(args[1],
			"goroutine number", 0, len(goTops)-1)
		if err != nil { return }
		gub.PrintGoroutine(goNum, goTops)
		return
	}
	for goNum := range goTops {
		gub.PrintGoroutine(goNum, goTops)
	}
}
