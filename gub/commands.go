// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

type CmdFunc func([]string)

type CmdInfo struct {
	help string
	category string
	min_args int
	max_args int
	fn CmdFunc
	aliases []string
}

var cmds map[string]*CmdInfo  = make(map[string]*CmdInfo)
var	aliases map[string]string = make(map[string]string)
var	categories map[string] []string = make(map[string] []string)

func AddAlias(alias string, cmdname string) bool {
	if unalias := aliases[alias]; unalias != "" {
		return false
	}
	aliases[alias] = cmdname
	cmds[cmdname].aliases = append(cmds[cmdname].aliases, alias)
	return true
}

func AddToCategory(category string, cmdname string) {
	categories[category] = append(categories[category], cmdname)
	// cmds[cmdname].category = category
}


func lookupCmd(cmd string) (string) {
	if cmds[cmd] == nil {
		cmd = aliases[cmd];
	}
	return cmd
}

func init() {
	name := "locations"
	cmds[name] = &CmdInfo{
		fn: LocsCommand,
		help: "show possible breakpoint locations",
		min_args: 0,
		max_args: 1,
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
		msg("\t%s", fmtPos(fn, l.Pos))
	}
}
