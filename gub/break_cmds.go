// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint-handling commands
package gub

import (
	"fmt"
	"strconv"
	"ssa-interp/interp"
)

func bpprint(bp Breakpoint) {

	disp := "keep "
	if bp.temp {
		disp  = "del  "
	}
	enabled := "n "
	if bp.enabled { enabled = "y " }

	loc  := fmtPos(curFrame.Fn(), bp.pos)
    mess := fmt.Sprintf("%3d breakpoint    %s  %sat %s",
		bp.id, disp, enabled, loc)
	msg(mess)

    // line_loc = '%s:%d' %
    //   [iseq.source_container.join(' '),
    //    iseq.offset2lines(bp.offset).join(', ')]

    // loc, other_loc =
    //   if 'line' == bp.type
    //     [line_loc, vm_loc]
    //   else # 'offset' == bp.type
    //     [vm_loc, line_loc]
    //   end
    // msg(mess + loc)
    // msg("\t#{other_loc}") if verbose

    // if bp.condition && bp.condition != 'true'
    //   msg("\tstop %s %s" %
    //       [bp.negate ? "unless" : "only if", bp.condition])
    // end
    if bp.ignore > 0 {
		msg("\tignore next %d hits", bp.ignore)
	}
    if bp.hits > 0 {
		ss := ""
		if bp.hits > 1 { ss = "s" }
		msg("\tbreakpoint already hit %d time%s",
			bp.hits, ss)
	}
}


func BreakpointList() {
	if IsBreakpointEmpty() {
		msg("No breakpoints set")
		return
	}
	if len(Breakpoints) - BrkptsDeleted == 0 {
		msg("No breakpoints.")
	}
	section("Num Type          Disp Enb Where")
	for _, bp := range Breakpoints {
		if bp.deleted { continue }
		bpprint(*bp)
	}
}

func BreakpointCommand(args []string) {
	if !argCountOK(0, 2, args) { return }
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if len(args) == 1 {
		BreakpointList()
		return
	}
	name := args[1]
	if fn := pkg.Func(name); fn != nil {
		interp.SetFnBreakpoint(fn)
		bp := &Breakpoint {
			hits: 0,
			id: len(Breakpoints),
			pos: fn.Pos(),
			endP: fn.EndP(),
			ignore: 0,
			kind: "Function",
			temp: false,
			enabled: true,
		}
		bpnum := BreakpointAdd(bp)
		msg("Breakpoint %d set in function %s at %s", bpnum, name,
			fmtPos(fn, fn.Pos()))
	}
}

func DeleteCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if BreakpointDelete(bpnum) {
				msg("Deleted breakpoint %d", bpnum)
			} else {
				errmsg("Trouble deleting breakpoint %d", bpnum)
			}
		} else {
			errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}

// FIXME: DRY the next two commands.
func DisableCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if !BreakpointIsEnabled(bpnum) {
				msg("Breakpoint %d is already disabled", bpnum)
				continue
			}
			if BreakpointDisable(bpnum) {
				msg("Breakpoint %d disabled", bpnum)
			} else {
				errmsg("Trouble disabling breakpoint %d", bpnum)
			}
		} else {
			errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}

func EnableCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if BreakpointIsEnabled(bpnum) {
				msg("Breakpoint %d is already enabled", bpnum)
				continue
			}
			if BreakpointEnable(bpnum) {
				msg("Breakpoint %d enabled", bpnum)
			} else {
				errmsg("Trouble enabling breakpoint %d", bpnum)
			}
		} else {
			errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
