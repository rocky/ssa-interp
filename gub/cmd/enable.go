// Copyright 2013 Rocky Bernstein.
package gubcmd
import (
	"strconv"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "enable"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: EnableCommand,
		Help: `enable [bpnum1 ...]

Enable a breakpoint by the number assigned to it.`,

		Min_args: 0,
		Max_args: -1,
	}
	gub.AddToCategory("breakpoints", name)
}

// FIXME: DRY with Disnable?
func EnableCommand(args []string) {
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			gub.Errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if gub.BreakpointExists(bpnum) {
			if gub.BreakpointIsEnabled(bpnum) {
				gub.Msg("Breakpoint %d is already enabled", bpnum)
				continue
			}
			if gub.BreakpointEnable(bpnum) {
				gub.Msg("Breakpoint %d enabled", bpnum)
			} else {
				gub.Errmsg("Trouble enabling breakpoint %d", bpnum)
			}
		} else {
			gub.Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
