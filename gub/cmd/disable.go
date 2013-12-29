// Copyright 2013 Rocky Bernstein.

package gubcmd
import (
	"fmt"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "disable"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: DisableCommand,
		Help: `Disable [bpnum1 ...]

Disable a breakpoint by its breakpoint number.

See also "enable", "delete", and "info break"`,
		Min_args: 0,
		Max_args: -1,
	}
	gub.AddToCategory("breakpoints", name)
}

// FIXME: DRY with Enable and Delete?

// DisableCommand implements the debugger command:
//    disable [bpnum1 ...]
// which disables a breakpoint by its breakpoint number.
//
// See also "enable", "delete", and "info break".
func DisableCommand(args []string) {
	for i:=1; i<len(args); i++ {
		msg := fmt.Sprintf("breakpoint number for argument %d", i)
		val, err := gub.GetUInt(args[i], msg, 0, uint64(len(gub.Breakpoints)-1))
		if err != nil { continue }
		bpnum := gub.BpId(val)
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
