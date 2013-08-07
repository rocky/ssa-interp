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
		Msg("%s#%d %s", pointer, i, fr.FnAndParamString())
		i++
	}
}

func init() {
	name := "backtrace"
	Cmds[name] = &CmdInfo{
		Fn: BacktraceCommand,
		Help: `backtrace [*count*]

Print a stack trace, with the most recent frame at the top.

With a positive number, print at most many entries.`,

		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("stack", name)
	Aliases["where"] = name
	Aliases["T"] = name  // for perl5db hackers
	// Down the line we'll have abbrevs
	Aliases["bt"] = name
}

func BacktraceCommand(args []string) {
	count := MAXSTACKSHOW
	var err error
	if len(args) > 1 {
		count, err = GetInt(args[1], "maximum count",
			0, MAXSTACKSHOW)
		if err != nil { return }
	}
	printStack(topFrame, count)
}

func init() {
	name := "down"
	Cmds[name] = &CmdInfo{
		Fn: DownCommand,
		Help: `down [*count*]

Move the current frame down in the stack trace (to a newer frame). 0
is the most recent frame. If no count is given, move down 1.

See also 'up' and 'frame'.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("stack", name)
}

func DownCommand(args []string) {
	count := 1
	if len(args) == 2 {
		var err error
		count, err = GetInt(args[1],
			"count", -MAXSTACKSHOW, MAXSTACKSHOW)
		if err != nil { return }
	}
	adjustFrame(-count, false)

}

func init() {
	name := "frame"
	Cmds[name] = &CmdInfo{
		Fn: FrameCommand,
		Help: `frame *num*

Change the current frame to frame *num*

See also 'up' and 'down'.
`,
		Min_args: 1,
		Max_args: 1,
	}
	AddToCategory("stack", name)
}

func FrameCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	i, err := GetInt(args[1],
		"frame number", -MAXSTACKSHOW, MAXSTACKSHOW)
	if err != nil { return }
	adjustFrame(i, true)

}

func printGoroutine(goNum int, goTops []*interp.GoreState) {
	fr := goTops[goNum].Fr
	if fr == nil {
		Msg("Goroutine %d exited", goNum)
		return
	}
	switch fr.Status() {
	case interp.StRunning:
		section("Goroutine %d", goNum)
		printStack(fr, MAXSTACKSHOW)
	case interp.StComplete:
		Msg("Goroutine %d completed", goNum)
	case interp.StPanic:
		Msg("Goroutine %d panic", goNum)
	}
}


func init() {
	name := "goroutines"
	Cmds[name] = &CmdInfo{
		Fn: GoroutinesCommand,
		Help: `goroutines [*id*]

Without a parameter, list stack traces for each active goroutine. If an id
is given only that goroutine stack trace is shown. The main (first) goroutine is 0.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("stack", name)
	Aliases["gore"] = name
	// Down the line we'll have abbrevs
	Aliases["gor"] = name
	Aliases["goroutine"] = name
}

// shows stack of all goroutines
func GoroutinesCommand(args []string) {
	goTops := interp.GetInterpreter().GoTops()
	var goNum int
	var err error
	if len(args) > 1 {
		goNum, err = GetInt(args[1],
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
	Cmds[name] = &CmdInfo{
		Fn: UpCommand,
		Help: `up [*count*]

Move the current frame up in the stack trace (to a older frame). 0
is the most-recent frame. If no count is given, move down 1.

See also 'down' and 'frame'.
`,
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("stack", name)
}

func UpCommand(args []string) {
	count := 1
	if len(args) == 2 {
		var err error
		count, err = GetInt(args[1],
			"count", -MAXSTACKSHOW, MAXSTACKSHOW)
		if err != nil { return }
	}
	adjustFrame(count, false)

}
