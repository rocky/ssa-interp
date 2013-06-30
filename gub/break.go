package gub

import (
	"go/token"
)

type Breakpoint struct {
	// condition
	hits    int       // How many times hit (with a true condition)
	id      int       // Id of breakpoint. Is position inside of Breakpoints
	deleted bool      // Set when breakpoint is deleted
	pos     token.Pos // Position of breakpoint
	endP    token.Pos // End Position of breakpoint
	ignore  bool      // Number of times to ignore before triggering
	kind    string    // 'Function' if function breakpoint. 'Stmt'
	                  // if at a statement boundary
}

var Breakpoints []*Breakpoint


func BreakpointAdd(bp *Breakpoint) {
	Breakpoints = append(Breakpoints, bp)
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

func IsBreakpointEmpty() bool {
	return len(Breakpoints) == 0
}
