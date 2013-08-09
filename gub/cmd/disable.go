// Copyright 2013 Rocky Bernstein.
package gubcmd
import (
	"strconv"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "disable"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: DisableCommand,
		Help: `Disable [bpnum1 ...]

Disable a breakpoint by the number assigned to it.`,

		Min_args: 0,
		Max_args: -1,
	}
	gub.AddToCategory("breakpoints", name)
}

// FIXME: DRY with Enable?
func DisableCommand(args []string) {
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			gub.Errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if gub.BreakpointExists(bpnum) {
			if !gub.BreakpointIsEnabled(bpnum) {
				gub.Msg("Breakpoint %d is already disabled", bpnum)
				continue
			}
			if gub.BreakpointDisable(bpnum) {
				gub.Msg("Breakpoint %d disabled", bpnum)
			} else {
				gub.Errmsg("Trouble disabling breakpoint %d", bpnum)
			}
		} else {
			gub.Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
