// Copyright 2014 Rocky Bernstein.

// show highlight - whether to use terminal highlight?

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "show"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: ShowHighlightSubcmd,
		Help: `show highlight

Show whether terminal highlighting is used`,
		Min_args: 0,
		Max_args: 0,
		Short_help: "show terminal highlight",
		Name: "highlight",
	})
}

func ShowHighlightSubcmd(args []string) {
	ShowOnOff(args[1], *gub.Highlight)
}
