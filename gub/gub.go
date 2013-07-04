// Copyright 2013 Rocky Bernstein.
// Top-level debugger interface
package gub

import (
	"fmt"
	"os"
	"strconv"

	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp/interp"
)

const (
	version string = "0.1"
)

var term string
var maxwidth int
var I *interp.Interpreter

func init() {
	term = os.Getenv("TERM")
	widthstr := os.Getenv("COLS")
	if len(widthstr) == 0 {
		maxwidth = 80
	} else if i, err := strconv.Atoi(widthstr); err == nil {
		maxwidth = i
	}
	gnureadline.StifleHistory(30)
}

func Install() {
	fmt.Printf("Gub version %s\n", version)
	fmt.Println("Type 'h' for help")
	interp.SetTraceHook(GubTraceHook)
	I = interp.GetInterpreter()
	fmt.Println(I)
}
