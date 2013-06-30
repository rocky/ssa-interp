// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint-handling commands
package gub

import (
	"strconv"
	"ssa-interp/interp"
)

func BreakpointList() {
	if IsBreakpointEmpty() {
		msg("No breakpoints set")
		return
	}
	section("List of Breakpoints")
	for i, bp := range Breakpoints {
		if bp.deleted { continue }
		msg("%d %s", i, bp)
	}
}

func BreakpointCommand(args []string) {
	if !argCountOK(0, 2, args) { return }
	name := args[1]
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if len(args) == 1 {
		BreakpointList()
		return
	}
	if fn := pkg.Func(name); fn != nil {
		interp.SetFnBreakpoint(fn)
		bp := &Breakpoint {
			hits: 0,
			id: len(Breakpoints),
			pos: fn.Pos(),
			endP: fn.EndP(),
			ignore: false,
			kind: "Function",
		}
		BreakpointAdd(bp)
		msg("Breakpoint set in function %s", name)
	}
}

func DeleteCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			errmsg("Expecting integer breakpoint at position %d number; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if BreakpointDelete(bpnum) {
				msg("Breakpoint %d deleted", bpnum)
			} else {
				errmsg("Trouble deleteing breakpoint %d", bpnum)
			}
		} else {
			errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
