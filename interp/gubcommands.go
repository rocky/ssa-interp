package interp

import (
	"fmt"
	"os"

	"gnureadline"
)

func argCountOK(min int, max int, args [] string) bool {
	l := len(args)-1 // strip command name from count
	if (l < min) {
		fmt.Printf("Too few args; need at least %d, got %d\n", min, l)
		return false
	} else if (l > max) {
		fmt.Printf("Too many args; need at most %d, got %d\n", max, l)
		return false
	}
	return true
}

func BacktraceCommand(fr *frame, args []string) {
	if !argCountOK(0, 1, args) { return }
	// FIXME: should get limit from args
	curFrame := fr
	for i:=0; curFrame !=nil; curFrame = curFrame.Caller {
		fmt.Printf("   #%d %s\n", i, StackLocation(curFrame))
		i++
	}
}

func QuitCommand(fr *frame, args []string) {
	if !argCountOK(0, 1, args) { return }
	fmt.Println("That's all folks...")
	gnureadline.Rl_reset_terminal(term)
	os.Exit(0)  // FIXME: Should use int arg

}

func VariableCommand(fr *frame, args []string) {
	if !argCountOK(1, 1, args) { return }
	fn := fr.Fn
	varname := args[1]
	for _, p := range fn.Locals {
		if varname == p.Name() { break }
	}

}
