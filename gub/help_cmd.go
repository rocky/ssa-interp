// Copyright 2013 Rocky Bernstein.

package gub

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
)

func init() {
	name := "help"
	Cmds[name] = &CmdInfo{
		Fn: HelpCommand,
		Help: `help [*command* | *category* | categories | * ]

Without argument, print the list of available debugger commands.

When an argument is given, if it is '*' a list of debugger commands
is shown. Otherwise the argument is checked to see if it is command
name. For example 'help up' gives help on the 'up' debugger command.

If a category name is given, a list of commands in that category is
shown. For a list of categories, enter "help categories".
`,

		Min_args: 0,
		Max_args: 1,
	}
	AddToCategory("support", name)
	AddAlias("?", name)
	// Down the line we'll have abbrevs
	AddAlias("h", name)
}

func HelpCommand(args []string) {
	if len(args) == 1 {
		Msg(Cmds["help"].Help)
	} else {
		what := args[1]
		cmd := lookupCmd(what)
		if what == "*" {
			var names []string
			for k, _ := range Cmds {
				names = append(names, k)
			}
			section("All command names:")
			sort.Strings(names)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(names, opts),
				"\n")
			Msg(mems)
		} else if what == "categories" {
			section("Categories")
			for k, _ := range categories {
				Msg("\t %s", k)
			}
		} else if info := Cmds[cmd]; info != nil {
			Msg(info.Help)
			if len(info.Aliases) > 0 {
				Msg("Aliases: %s",
					strings.Join(info.Aliases, ", "))
			}
		} else if cmds := categories[what]; len(cmds) > 0 {
			section("Commands in class: %s", what)
			sort.Strings(cmds)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(cmds, opts),
				"\n")
			Msg(mems)
		} else {
			Errmsg("Can't find help for %s", what)
		}
	}
}
