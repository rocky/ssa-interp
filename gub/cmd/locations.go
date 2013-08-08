// Copyright 2013 Rocky Bernstein.
// disassemble command

package gubcmd

import	(
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "locations"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: LocationsCommand,
		Help: `locations

Give list of all the stopping locations in the package
`,
		Min_args: 0,
		Max_args: 1,
	}
	gub.AddToCategory("status", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("locs", name)
	gub.AddAlias("loc", name)
	gub.AddAlias("location", name)
}

func LocationsCommand(args []string) {
	fn  := gub.CurFrame().Fn()
	pkg := fn.Pkg
	for _, l := range pkg.Locs() {
		gub.Msg("\t%s", ssa2.FmtRange(fn, l.Pos(), l.EndP()))
	}
}
