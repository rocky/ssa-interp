// Copyright 2013 Rocky Bernstein.
// Debugger info breakpoint command

package gubcmd
import 	"github.com/rocky/ssa-interp/gub"

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoScopeSubcmd,
		Help: `info breakpoint [num]

Show status of user-settable breakpoints. If no breakpoint numbers are
given, the show all breakpoints. Otherwise only those breakpoints
listed are shown and the order given.

The "Disp" column contains one of "keep", "del", the disposition of
the breakpoint after it gets hit.

The "enb" column indicates whether the breakpoint is enabled.

The "Where" column indicates where the breakpoint is located.
Status of user-settable breakpoints.
`,
		Min_args: 1,
		Max_args: 2,
		Short_help: "Status of user-settable breakpoints",
		Name: "breakpoint",
	})
}

func InfoBreakpointSubcmd() {
	if gub.IsBreakpointEmpty() {
		gub.Msg("No breakpoints set")
		return
	}
	if len(gub.Breakpoints) - gub.BrkptsDeleted == 0 {
		gub.Msg("No breakpoints.")
	}
	gub.Section("Num Type          Disp Enb Where")
	for _, bp := range gub.Breakpoints {
		if bp.Deleted { continue }
		gub.Bpprint(*bp)
	}
}
