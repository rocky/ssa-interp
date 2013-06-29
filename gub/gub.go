// Copyright 2013 Rocky Bernstein.
// Top-level debugger interface
package gub

import (
	"fmt"
	"os"

	"gnureadline"
	"ssa-interp/interp"
)

const (
	version string = "0.1"
)

var term string

func init() {
	term = os.ExpandEnv("TERM")
	gnureadline.StifleHistory(30)
}

func Install() {
	fmt.Printf("Gub version %s\n", version)
	fmt.Println("Type 'h' for help")
	interp.SetTraceHook(GubTraceHook)
}
