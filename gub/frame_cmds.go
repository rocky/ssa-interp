// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"fmt"
	"strconv"
)

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
