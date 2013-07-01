// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"fmt"
	"strconv"
)

func BacktraceCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
	// FIXME: should get limit from args
	fr := topFrame
	for i:=0; fr !=nil; fr = fr.Caller(0) {
		pointer := "   "
		if fr == curFrame {
			pointer = "=> "
		}
		msg("%s#%d %s", pointer, i, StackLocation(fr))
		i++
	}
}

func FrameCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	frameIndex, ok := strconv.Atoi(args[1])
	if ok != nil {
		errmsg("Expecting integer frame number; got %s",
			args[1])
		return
	}
	adjustFrame(frameIndex, true)

}

func DownCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
	var frameIndex int
	var ok error
	if len(args) == 1 {
		frameIndex = 1
	} else {
		frameIndex, ok = strconv.Atoi(args[1])
		if ok != nil {
			errmsg("Expecting integer frame number; got %s", args[1])
			return
		}
	}
	adjustFrame(-frameIndex, false)

}

func UpCommand(args []string) {
	if !argCountOK(0, 1, args) { return }
	var frameIndex int
	var ok error
	if len(args) == 1 {
		frameIndex = 1
	} else {
		frameIndex, ok = strconv.Atoi(args[1])
		if ok != nil {
			fmt.Printf("Expecting integer frame number; got %s",
				args[1])
			return
		}
	}
	adjustFrame(frameIndex, false)

}
