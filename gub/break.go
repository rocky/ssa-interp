package gub

import (
	"go/token"
)

type Breakpoint struct {
	// condition
	Hits    int       // How many times hit (with a true condition)
	Id      int       // Id of breakpoint. Is position inside of Breakpoints
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

func BreakpointIsEnabled(bpnum int) bool {
	if BreakpointExists(bpnum) {
		return Breakpoints[bpnum].Enabled
	}
	return false
}


func IsBreakpointEmpty() bool {
	return len(Breakpoints) == 0
}
