package gub

import (
	"go/token"
)

type Breakpoint struct {
	// condition
	hits    int       // How many times hit (with a true condition)
	id      int       // Id of breakpoint. Is position inside of Breakpoints
	deleted bool      // Set when breakpoint is deleted
	temp    bool      // Set when one-time breakpoint
	enabled bool      // Set when breakpoint is enabled
	pos     token.Pos // Position of breakpoint
	endP    token.Pos // End Position of breakpoint
	ignore  int       // Number of times to ignore before triggering
	kind    string    // 'Function' if function breakpoint. 'Stmt'
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
	BrkptLocs = append(BrkptLocs, toknum{pos: bp.pos, bpnum: bp.id})
	return len(Breakpoints)-1
}

func BreakpointExists(bpnum int) bool {
	if bpnum < len(Breakpoints) {
		return !Breakpoints[bpnum].deleted
	}
	return false
}

func BreakpointDelete(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].deleted = true
		return true
	}
	return false
}

func BreakpointDisable(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].enabled = false
		return true
	}
	return false
}

func BreakpointEnable(bpnum int) bool {
	if BreakpointExists(bpnum) {
		Breakpoints[bpnum].enabled = true
		return true
	}
	return false
}

func BreakpointFindByPos(pos token.Pos) []int {
	results := make([]int, 0)
	for _, v := range BrkptLocs {
		if v.pos == pos && !Breakpoints[v.bpnum].deleted {
			results = append(results, v.bpnum)
		}
	}
	return results
}

func BreakpointIsEnabled(bpnum int) bool {
	if BreakpointExists(bpnum) {
		return Breakpoints[bpnum].enabled
	}
	return false
}


func IsBreakpointEmpty() bool {
	return len(Breakpoints) == 0
}
