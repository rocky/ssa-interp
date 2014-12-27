// Copyright 2013-2014 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

// Init is just a stub to call from the outside to allow
// import not to complain and to force the rest of the init()
// routines to run
func Init(restart_args string) {
	if len(restart_args) > 0 {
		gub.GUB_RESTART_CMD = restart_args
	}
}
