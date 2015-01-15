// Copyright 2013 Rocky Bernstein.
// Top-level debugger interface

package gub

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"code.google.com/p/go-gnureadline"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

var terse     = flag.Bool("terse", true, `abbreviated output`)
var testing   = flag.Bool("testing", false, `used in testing`)
var Highlight = flag.Bool("highlight", true, `use syntax highlighting in output`)
var inputFilename = flag.String("cmdfile", "", `cmdfile *commandfile*.`)
var inputFile *os.File
var inputReader *bufio.Reader
var buffer = bytes.NewBuffer(make([]byte, 1024))

var program  *ssa2.Program
func Program() *ssa2.Program { return program }

const (
	version string = "0.3"
)

// Term is the current environment TERM value, e.g. "gnome", "xterm", or "vt100"
var Term string

// Maxwidth is the size of the line. We will try to wrap text that is
// longer than this. It like the COLS environment variable
var Maxwidth int

var initial_cwd string

//GUB is a string that was used to invoke gofish.
//If we want to restart gub, this is what we'll use.
var RESTART_ARGS []string

// history_file is file name where history entries were and are to be saved. If
// the empty string, no history is saved and no history read in initially.
var historyFile string

// gnuReadLineTermination has GNU Readline Termination tasks:
// save history file if ane, and reset the terminal.
func gnuReadLineTermination() {
	if historyFile != "" {
		gnureadline.WriteHistory(historyFile)
	}
	if Term != "" {
		gnureadline.Rl_reset_terminal(Term)
	}
}

// HistoryFile returns a string file name to use for saving command
// history entries
func HistoryFile(history_basename string) string {
	home_dir := os.Getenv("HOME")
	if home_dir == "" {
		// FIXME: also try ~ ?
		fmt.Println("ignoring history file; environment variable HOME not set")
		return ""
	}
	history_file := filepath.Join(home_dir, history_basename)
	if fi, err := os.Stat(history_file); err != nil {
		fmt.Println("No history file found to read in")
	} else {
		if fi.IsDir() {
			fmt.Printf("Ignoring history file %s; is a directory, should be a file",
				history_file)
			return ""
		}
	}
	return history_file
}

// gnuReadLineSetup is boilerplate initialization for GNU Readline.
func gnuReadLineSetup() {
	Term = os.Getenv("TERM")
	historyFile = HistoryFile(".gub")
	if historyFile != "" {
		gnureadline.ReadHistory(historyFile)
	}
	// Set maximum number of history entries
	gnureadline.StifleHistory(30)
}

func init() {
	widthstr := os.Getenv("COLS")
	initial_cwd, _ = os.Getwd()
	if len(widthstr) == 0 {
		Maxwidth = 80
	} else if i, err := strconv.Atoi(widthstr); err == nil {
		Maxwidth = i
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
		if *testing { *Highlight = false }
		if inputFilename != nil && len(*inputFilename) > 0 {
			var err error
			if inputFile, err = os.Open(*inputFilename); err != nil {
				fmt.Println("Error opening debugger command file ",
					inputFilename)
				os.Exit(1)
			}
			inputReader = bufio.NewReader(inputFile)
		} else {
			gnuReadLineSetup()
			defer gnuReadLineTermination()
		}

	}
}

func Install(options *string, restart_args []string, prog *ssa2.Program) {
	program = prog
	RESTART_ARGS = restart_args
	fmt.Printf("Gub version %s\n", version)
	fmt.Println("Type 'h' for help")
	interp.SetTraceHook(GubTraceHook)
	process_options(options)
}
