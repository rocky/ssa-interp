// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

type CmdFunc func([]string)

type CmdInfo struct {
	Help string
	Category string
	Min_args int
	Max_args int
	Fn CmdFunc
	Aliases []string
}

var Cmds map[string]*CmdInfo  = make(map[string]*CmdInfo)
var	aliases map[string]string = make(map[string]string)
var	categories map[string] []string = make(map[string] []string)

func AddAlias(alias string, cmdname string) bool {
	if unalias := aliases[alias]; unalias != "" {
		return false
	}
	aliases[alias] = cmdname
	Cmds[cmdname].Aliases = append(Cmds[cmdname].Aliases, alias)
	return true
}

func AddToCategory(category string, cmdname string) {
	categories[category] = append(categories[category], cmdname)
	// Cmds[cmdname].category = category
}


func lookupCmd(cmd string) (string) {
	if Cmds[cmd] == nil {
		cmd = aliases[cmd];
	}
	return cmd
}

func init() {
	name := "locations"
	Cmds[name] = &CmdInfo{
		Fn: LocsCommand,
		Help: "show possible breakpoint locations",
		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("status", name)
	// Down the line we'll have abbrevs
	AddAlias("locs", name)
}

func LocsCommand(args []string) {
	fn  := curFrame.Fn()
	pkg := fn.Pkg
	for _, l := range pkg.Locs() {
		// FIXME: ? turn into true range
		Msg("\t%s", fmtPos(fn, l.Pos))
	}
}
