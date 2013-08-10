// Copyright 2013 Rocky Bernstein.
// Debugger subcommands
package gub

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
)

type SubcmdFunc func([]string)

type SubcmdInfo struct {
	Help string
	Short_help string
	Min_args int
	Max_args int
	Fn SubcmdFunc
	Name string
}

type SubcmdMgr struct {
	Name string
	Subcmds map[string]*SubcmdInfo
}

var Subcmds map[string]*SubcmdInfo  = make(map[string]*SubcmdInfo)

func AddSubCommand(mgrName string, subcmdInfo *SubcmdInfo) {
	Subcmds[mgrName] = subcmdInfo
	mgr := Cmds[mgrName]
	if mgr != nil {
		mgr.SubcmdMgr.Subcmds[subcmdInfo.Name] = subcmdInfo
	} else {
		Errmsg("Internal error: can't find command '%s' to add to", subcmdInfo.Name)
	}
}

func HelpSubCommand(subcmdMgr *SubcmdMgr, args []string) {
	if len(args) == 2 {
		Msg(Cmds[subcmdMgr.Name].Help)
	} else {
		what := args[2]
		if what == "*" {
			var names []string
			for name, _ := range subcmdMgr.Subcmds {
				names = append(names, name)
			}
			Section("All %s subcommand names:", subcmdMgr.Name)
			sort.Strings(names)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = Maxwidth
			mems := strings.TrimRight(columnize.Columnize(names, opts),
				"\n")
			Msg(mems)
		} else if info := subcmdMgr.Subcmds[what]; info != nil {
			Msg(info.Help)
		} else {
			Errmsg("Can't find help for subcommand '%s' in %s", what, subcmdMgr.Name)
		}
	}
}
