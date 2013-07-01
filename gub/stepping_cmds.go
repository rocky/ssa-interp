// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution
package gub

import (
	"fmt"
	"github.com/rocky/ssa-interp/interp"
)

func Continue(args []string) {
	for fr := topFrame; fr != nil; fr = fr.Caller(0) {
		interp.SetStepOff(fr)
	}
	fmt.Println("Continuing...")
}

func FinishCommand(args []string) {
	interp.SetStepOut(topFrame)
	msg("Continuing until return...")
}

func NextCommand(args []string) {
	interp.SetStepOver(topFrame)
	fmt.Println("Step over...")
}

func StepCommand(args []string) {
	fmt.Println("Stepping...")
	interp.SetStepIn(curFrame)
}
