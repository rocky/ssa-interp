// Copyright 2013 Rocky Bernstein.
// Debugger subcommands
package gub

type SubcmdFunc func([]string)

type SubcmdInfo struct {
	help string
	short_help string
	min_args int
	max_args int
	fn SubcmdFunc
}

var subcmds map[string]*SubcmdInfo  = make(map[string]*SubcmdInfo)
