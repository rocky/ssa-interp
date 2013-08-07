// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "help"
	gub.Cmds[name] = &gub.CmdInfo{
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
	gub.AddToCategory("support", name)
	gub.AddAlias("?", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("h", name)
}

func HelpCommand(args []string) {
	if len(args) == 1 {
		gub.Msg(gub.Cmds["help"].Help)
	} else {
		what := args[1]
		cmd := gub.LookupCmd(what)
		if what == "*" {
			var names []string
			for k, _ := range gub.Cmds {
				names = append(names, k)
			}
			gub.Section("All command names:")
			sort.Strings(names)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = gub.Maxwidth
			mems := strings.TrimRight(columnize.Columnize(names, opts),
				"\n")
			gub.Msg(mems)
		} else if what == "categories" {
			gub.Section("Categories")
			for k, _ := range gub.Categories {
				gub.Msg("\t %s", k)
			}
		} else if info := gub.Cmds[cmd]; info != nil {
			gub.Msg(info.Help)
			if len(info.Aliases) > 0 {
				gub.Msg("Aliases: %s",
					strings.Join(info.Aliases, ", "))
			}
		} else if cmds := gub.Categories[what]; len(cmds) > 0 {
			gub.Section("Commands in class: %s", what)
			sort.Strings(cmds)
			opts := columnize.DefaultOptions()
			opts.DisplayWidth = gub.Maxwidth
			mems := strings.TrimRight(columnize.Columnize(cmds, opts),
				"\n")
			gub.Msg(mems)
		} else {
			gub.Errmsg("Can't find help for %s", what)
		}
	}
}
