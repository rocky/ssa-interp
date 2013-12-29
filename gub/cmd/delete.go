// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint delete command

package gubcmd

import (
	"fmt"
 	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "delete"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: DeleteCommand,
		Help: `Delete [bpnum1 ...]

Delete a breakpoint by breakpoint number.
`,

		Min_args: 0,
		Max_args: -1,
	}
	gub.AddToCategory("breakpoints", name)
	// Down the line we'll have abbrevs
	gub.AddAlias("del", name)
}

// DeleteCommand implements the debugger command:
//    delete [bpnum1 ... ]
// which deletes some breakpoints by breakpoint number
//
// See also "breakpoint", "info break", "enable", and "disable".
func DeleteCommand(args []string) {
	for i:=1; i<len(args); i++ {
		msg := fmt.Sprintf("breakpoint number for argument %d", i)
		val, err := gub.GetUInt(args[i], msg, 0, uint64(len(gub.Breakpoints)-1))
		if err != nil { continue }
		bpnum := gub.BpId(val)
		if gub.BreakpointExists(bpnum) {
			if gub.BreakpointDelete(bpnum) {
				gub.Msg(" Deleted breakpoint %d", bpnum)
			} else {
				gub.Errmsg("Trouble deleting breakpoint %d", bpnum)
			}
		} else {
			gub.Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
