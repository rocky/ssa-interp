// Copyright 2013 Rocky Bernstein.
// Top-level debugger interface
package gub

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp/interp"
)

var terse = flag.Bool("terse", true, `abbreviated output`)
var inputFilename = flag.String("cmdfile", "", `cmdfile *commandfile*.`)
var inputFile *os.File
var inputReader *bufio.Reader
var buffer = bytes.NewBuffer(make([]byte, 1024))

const (
	version string = "0.1"
)

var term string
var maxwidth int

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
		// fmt.Println("Args are ", args)
		os.Args = args
		flag.Parse()
		if inputFilename != nil && len(*inputFilename) > 0 {
			var err error
			if inputFile, err = os.Open(*inputFilename); err != nil {
				fmt.Println("Error opening debugger command file ",
					inputFilename)
				os.Exit(1)
			}
			inputReader = bufio.NewReader(inputFile)
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
