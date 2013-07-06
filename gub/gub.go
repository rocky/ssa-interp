// Copyright 2013 Rocky Bernstein.
// Top-level debugger interface
package gub

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp/interp"
)

var inputFilename = flag.String("cmdinput", "", `cmdinput *commandfile*.`)

const (
	version string = "0.1"
)

var term string
var maxwidth int
var I *interp.Interpreter

func termInit() {
	term = os.Getenv("TERM")
	gnureadline.StifleHistory(30)
}

func init() {
	widthstr := os.Getenv("COLS")
	if len(widthstr) == 0 {
		maxwidth = 80
	} else if i, err := strconv.Atoi(widthstr); err == nil {
		maxwidth = i
	}
}

func process_options(options *string) {
	if options != nil {
		var args []string
		args = append(args, os.Args[0])
		for _, s := range strings.Split(*options, " ") {
			args = append(args, s)
		}
		os.Args = args
		flag.Parse()
		if inputFilename != nil {
			fmt.Println("TODO: read in", *inputFilename)
			termInit()
		} else {
			termInit()
		}

	}
}

func Install(options *string) {
	fmt.Printf("Gub version %s\n", version)
	fmt.Println("Type 'h' for help")
	interp.SetTraceHook(GubTraceHook)
	process_options(options)
}
