// Copyright 2013, 2015 Rocky Bernstein.
// Debugger info breakpoint command

package gubcmd
import 	"github.com/rocky/ssa-interp/gub"

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoBreakpointSubcmd,
		Help: `info breakpoint [num...]

Show status of user-settable breakpoints. If no breakpoint numbers are
given, the show all breakpoints. Otherwise only those breakpoints
listed are shown and the order given.

The "Disp" column contains one of "keep", "del", the disposition of
the breakpoint after it gets hit.

The "enb" column indicates whether the breakpoint is enabled.

The "Where" column indicates where the breakpoint is located.
Status of user-settable breakpoints.
`,
		Min_args: 0,
		Max_args: -1,
		Short_help: "Status of user-settable breakpoints",
		Name: "breakpoint",
	})
}

// InfoBreakpointSubcmd implements the debugger command:
//   info breakpoint
//
// This command shows status of user-settable breakpoints. If no
// breakpoint numbers are given, the show all breakpoints. Otherwise
// only those breakpoints listed are shown and the order given.
//
// The "Disp" column contains one of "keep", "del", the disposition of
// the breakpoint after it gets hit.
//
// The "enb" column indicates whether the breakpoint is enabled.
//
// The "Where" column indicates where the breakpoint is located.
// Status of user-settable breakpoints.
func InfoBreakpointSubcmd(args [] string) {
	if gub.IsBreakpointEmpty() {
		gub.Msg("No breakpoints set")
		return
	}
	bpLen := len(gub.Breakpoints)
	if bpLen - gub.BrkptsDeleted == 0 {
		gub.Msg("No breakpoints.")
	}
	if len(args) > 2 {
		headerShown := false
		for _, num := range args[2:] {
			if bpNum, err := gub.GetInt(num,
				"breakpoint number", 0, bpLen-1); err==nil {
					if bp := gub.BreakpointFindById(bpNum); bp != nil {
						if !headerShown {
							gub.Section("Num Type          Disp Enb Where")
							headerShown = true
						}
						gub.Bpprint(*bp)
					} else {
						gub.Errmsg("Breakpoint %d not found.", bpNum)
					}
				}
		}
	} else {
		gub.Section("Num Type          Disp Enb Where")
		for _, bp := range gub.Breakpoints {
			if bp.Deleted { continue }
			gub.Bpprint(*bp)
		}
	}
}
