// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"github.com/rocky/ssa-interp/interp"
)

func printStack(fr *interp.Frame, count int) {
	if (fr == nil) { return }
	for i:=0; fr !=nil && i < count; fr = fr.Caller(0) {
		pointer := "   "
		if fr == curFrame {
			pointer = "=> "
		}
		msg("%s#%d %s", pointer, i, StackLocation(fr))
		i++
	}
}

func init() {
	name := "backtrace"
	cmds[name] = &CmdInfo{
		fn: BacktraceCommand,
		help: `backtrace [*count*]

Print a stack trace, with the most recent frame at the top.

With a positive number, print at most many entries.`,

		min_args: 0,
		max_args: 1,
	}
	AddToCategory("stack", name)
	aliases["where"] = name
	aliases["T"] = name  // for perl5db hackers
	// Down the line we'll have abbrevs
	aliases["bt"] = name
}

func BacktraceCommand(args []string) {
	count := MAXSTACKSHOW
	var err error
	if len(args) > 1 {
		count, err = getInt(args[1], "max count",
			0, MAXSTACKSHOW)
		if err != nil { return }
	}
	printStack(topFrame, count)
}

func init() {
	name := "down"
	cmds[name] = &CmdInfo{
		fn: DownCommand,
		help: `down [*count*]

Move the current frame down in the stack trace (to a newer frame). 0
is the most recent frame. If no count is given, move down 1.

See also 'up' and 'frame'.
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("stack", name)
}

func DownCommand(args []string) {
	var count int
	var err error
	if len(args) == 1 {
		frameIndex = 1
	} else {
		count, err = getInt(args[1],
			"count", -MAXSTACKSHOW, MAXSTACKSHOW)
		if err != nil { return }
	}
	adjustFrame(-count, false)

}

func init() {
	name := "frame"
	cmds[name] = &CmdInfo{
		fn: FrameCommand,
		help: `frame *num*

Change the current frame to frame *num*

See also 'up' and 'down'.
`,
		min_args: 1,
		max_args: 1,
	}
	AddToCategory("stack", name)
}

func FrameCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	i, err := getInt(args[1],
		"frame number", -MAXSTACKSHOW, MAXSTACKSHOW)
	if err != nil { return }
	adjustFrame(i, true)

}

func printGoroutine(goNum int, goTops []*interp.GoreState) {
	fr := goTops[goNum].Fr
	if fr == nil {
		msg("Goroutine %d exited", goNum)
		return
	}
	switch fr.Status() {
	case interp.StRunning:
		section("Goroutine %d", goNum)
		printStack(fr, MAXSTACKSHOW)
	case interp.StComplete:
		msg("Goroutine %d completed", goNum)
	case interp.StPanic:
		msg("Goroutine %d panic", goNum)
	}
}


func init() {
	name := "goroutines"
	cmds[name] = &CmdInfo{
		fn: GoroutinesCommand,
		help: "global [*name*]: show global variable info",
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("stack", name)
	aliases["gore"] = name
	// Down the line we'll have abbrevs
	aliases["gor"] = name
	aliases["goroutine"] = name
}

// shows stack of all goroutines
func GoroutinesCommand(args []string) {
	goTops := interp.GetInterpreter().GoTops()
	var goNum int
	var err error
	if len(args) > 1 {
		goNum, err = getInt(args[1],
			"goroutine number", 0, len(goTops)-1)
		if err != nil { return }
		printGoroutine(goNum, goTops)
		return
	}
	for goNum := range goTops {
		printGoroutine(goNum, goTops)
	}
}

func init() {
	name := "up"
	cmds[name] = &CmdInfo{
		fn: UpCommand,
		help: `up [*count*]

Move the current frame up in the stack trace (to a older frame). 0
is the most-recent frame. If no count is given, move down 1.

See also 'down' and 'frame'.
`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("stack", name)
}

func UpCommand(args []string) {
	var count int
	var err error
	if len(args) == 1 {
		frameIndex = 1
	} else {
		count, err = getInt(args[1],
			"count", -MAXSTACKSHOW, MAXSTACKSHOW)
		if err != nil { return }
	}
	adjustFrame(count, false)

}
