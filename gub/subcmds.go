// Copyright 2013 Rocky Bernstein.
// Debugger subcommands
package gub

type SubcmdFunc func([]string)

type SubcmdInfo struct {
	Help string
	Short_help string
	Min_args int
	Max_args int
	Fn SubcmdFunc
}

var Subcmds map[string]*SubcmdInfo  = make(map[string]*SubcmdInfo)
