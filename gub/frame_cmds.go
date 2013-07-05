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

func BacktraceCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
	count := MAXSTACKSHOW
	var err error
	if len(args) > 1 {
		count, err = getInt(args[1], "max count",
			0, MAXSTACKSHOW)
		if err != nil { return }
	}
	printStack(topFrame, count)
}

func DownCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
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

func FrameCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	i, err := getInt(args[1],
		"frame number", -MAXSTACKSHOW, MAXSTACKSHOW)
	if err != nil { return }
	adjustFrame(i, true)

}

func printGoroutine(goNum int, goTops []*interp.GoreState) {
	fr := goTops[goNum].Fr
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


// shows stack of all goroutines
func GoroutinesCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
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

func UpCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
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
