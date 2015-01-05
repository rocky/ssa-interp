// Copyright 2013-2015 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

// Init is a call from the outside. By pulling
// importing the package, we run all of the init routines
// in each of the commands. We also call the parent
// processor to do its initialization
func Init(options *string, restart_args []string, prog *ssa2.Program) {
	gub.Install(options, restart_args, prog)
}
