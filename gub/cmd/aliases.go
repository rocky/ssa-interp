// Copyright 2013-2014 Rocky Bernstein.

package gubcmd

import (
	"sort"
	"strings"
	"code.google.com/p/go-columnize"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "aliases"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: AliasCommand,
		Help: `aliases

Without argument, print the list of debugger command aliases.

When an argument is given, if it is command name, the aliases for
that command are shown. if the argument is an alias name, we'll
show the command that this is an alias for.
`,

		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("support", name)
	gub.AddAlias("alias", name)
}

func AliasCommand(args []string) {
	if len(args) == 1 {
		var names []string
		for k, _ := range gub.Aliases {
			names = append(names, k)
		}
		gub.Section("All aliases:")
		sort.Strings(names)
		opts := columnize.DefaultOptions()
		opts.DisplayWidth = gub.Maxwidth
		opts.LinePrefix  = "  "
		mems := strings.TrimRight(columnize.Columnize(names, opts),
			"\n")
		gub.Msg(mems)
	} else {
		cmd := args[1]
		if info := gub.Cmds[cmd]; info != nil {
			if len(info.Aliases) > 0 {
				gub.Msg("Aliases for %s: %s",
					cmd, strings.Join(info.Aliases, ", "))
			} else {
				gub.Msg("No aliases for %s", cmd)
			}
		} else if realCmd := gub.Aliases[cmd]; realCmd != "" {
			gub.Msg("Alias %s is an alias for command %s", cmd, realCmd)

		} else {
			gub.Errmsg("Can't find command or alias %s", cmd)
		}
	}
}
