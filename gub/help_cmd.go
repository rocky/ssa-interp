// Copyright 2013 Rocky Bernstein.

package gub

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
)

func init() {
	name := "help"
	cmds[name] = &CmdInfo{
		fn: HelpCommand,
		help: `help [*command* | *category* | categories | * ]

Without argument, print the list of available debugger commands.

When an argument is given, if it is '*' a list of debugger commands
is shown. Otherwise the argument is checked to see if it is command
name. For example 'help up' gives help on the 'up' debugger command.

If a category name is given, a list of commands in that category is
shown. For a list of categories, enter "help categories".
`,

		min_args: 0,
		max_args: 1,
	}
	AddToCategory("support", name)
	AddAlias("?", name)
	// Down the line we'll have abbrevs
	AddAlias("h", name)
}

func HelpCommand(args []string) {
	if len(args) == 1 {
		msg(cmds["help"].help)
	} else {
		what := args[1]
		cmd := lookupCmd(what)
		if what == "*" {
			var names []string
			for k, _ := range cmds {
				names = append(names, k)
			}
			section("All command names:")
			sort.Strings(names)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(names, opts),
				"\n")
			msg(mems)
		} else if what == "categories" {
			section("Categories")
			for k, _ := range categories {
				msg("\t %s", k)
			}
		} else if info := cmds[cmd]; info != nil {
			msg(info.help)
			if len(info.aliases) > 0 {
				msg("Aliases: %s",
					strings.Join(info.aliases, ", "))
			}
		} else if cmds := categories[what]; len(cmds) > 0 {
			section("Commands in class: %s", what)
			sort.Strings(cmds)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = maxwidth
			mems := strings.TrimRight(columnize.Columnize(cmds, opts),
				"\n")
			msg(mems)
		} else {
			errmsg("Can't find help for %s", what)
		}
	}
}
