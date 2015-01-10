package gub

import (
	"go/token"
	"github.com/rocky/ssa-interp"
	"fmt"
)

type Breakpoint struct {
	// condition
	Hits    int       // How many times hit (with a true condition)
	Id      int      // Id of breakpoint. Is position inside of Breakpoints
	Deleted bool      // Set when breakpoint is deleted
	Temp    bool      // Set when one-time breakpoint
	Enabled bool      // Set when breakpoint is enabled
	Pos     token.Pos // Position of breakpoint
	EndP    token.Pos // End Position of breakpoint
	Ignore  int       // Number of times to ignore before triggering
	Kind    string    // 'Function' if function breakpoint. 'Stmt'
	                  // if at a statement boundary
}

var Breakpoints []*Breakpoint

// FIXME: this should be a slice indexted by token.Pos
// of a slice of breakpoint numbers
type toknum struct {
	pos token.Pos
	bpnum int
}
var BrkptLocs []toknum = make([]toknum, 0)

// Deleting a breakpoint doesn't remove it from a slice,
// since that will mess up numbering.
// We use BrkptDeleted to compensate in breakpoint counts.
var BrkptsDeleted = 0

func BreakpointAdd(bp *Breakpoint) int {
	Breakpoints = append(Breakpoints, bp)
	BrkptLocs = append(BrkptLocs, toknum{pos: bp.Pos, bpnum: bp.Id})
	return len(Breakpoints)-1
}

func BreakpointExists(bpnum int) bool {
	if bpnum < len(Breakpoints) {
		return !Breakpoints[bpnum].Deleted
	}
	return false
}

func BreakpointDelete(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].Deleted = true
		return true
	}
	return false
}

func BreakpointDisable(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].Enabled = false
		return true
	}
	return false
}

func BreakpointEnable(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].Enabled = true
		return true
	}
	return false
}

func BreakpointFindByPos(pos token.Pos) []int {
	results := make([]int, 0)
	for _, v := range BrkptLocs {
		if v.pos == pos && !Breakpoints[v.bpnum].Deleted {
			results = append(results, v.bpnum)
		}
	}
	return results
}

func BreakpointFindById(bpNum int) *Breakpoint {
	for _, bp := range Breakpoints {
		if bp.Id == bpNum { return bp }
	}
	return nil
}

func BreakpointIsEnabled(bpnum int) bool {
	if BreakpointExists(bpnum) {
		return Breakpoints[bpnum].Enabled
	}
	return false
}


func BreakpointNext() int {
	return len(Breakpoints)
}

func IsBreakpointEmpty() bool {
	return len(Breakpoints) == 0
}

func Bpprint(bp Breakpoint) {

	disp := "keep "
	if bp.Temp {
		disp  = "del  "
	}
	enabled := "n "
	if bp.Enabled { enabled = "y " }

	loc  := ssa2.FmtRange(curFrame.Fn(), bp.Pos, bp.EndP)
    mess := fmt.Sprintf("%3d breakpoint    %s  %sat %s",
		bp.Id, disp, enabled, loc)
	Msg(mess)

    // line_loc = '%s:%d' %
    //   [iseq.source_container.join(' '),
    //    iseq.offset2lines(bp.offset).join(', ')]

    // loc, other_loc =
    //   if 'line' == bp.type
    //     [line_loc, vm_loc]
    //   else # 'offset' == bp.type
    //     [vm_loc, line_loc]
    //   end
    // Msg(mess + loc)
    // Msg("\t#{other_loc}") if verbose

    // if bp.condition && bp.condition != 'true'
    //   Msg("\tstop %s %s" %
    //       [bp.negate ? "unless" : "only if", bp.condition])
    // end
    if bp.Ignore > 0 {
		Msg("\tignore next %d hits", bp.Ignore)
	}
    if bp.Hits > 0 {
		ss := ""
		if bp.Hits > 1 { ss = "s" }
		Msg("\tbreakpoint already hit %d time%s",
			bp.Hits, ss)
	}
}
