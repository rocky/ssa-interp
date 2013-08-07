// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint-handling commands
package gub

import (
	"fmt"
	"strconv"
	"github.com/rocky/ssa-interp/interp"
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
    if bp.ignore > 0 {
		Msg("\tignore next %d hits", bp.ignore)
	}
    if bp.hits > 0 {
		ss := ""
		if bp.hits > 1 { ss = "s" }
		Msg("\tbreakpoint already hit %d time%s",
			bp.hits, ss)
	}
}


func InfoBreakpointSubcmd() {
	if IsBreakpointEmpty() {
		Msg("No breakpoints set")
		return
	}
	if len(Breakpoints) - BrkptsDeleted == 0 {
		Msg("No breakpoints.")
	}
	Section("Num Type          Disp Enb Where")
	for _, bp := range Breakpoints {
		if bp.deleted { continue }
		bpprint(*bp)
	}
}

func init() {
	name := "breakpoint"
	Cmds[name] = &CmdInfo{
		Fn: BreakpointCommand,
		Help: `breakpoint [*fn* | line [column]]

Set a breakpoint. The target can either be a function name as fn pkg.fn
or a line and and optional column number. Specifying a column number
may be useful if there is more than one statement on a line or if you
want to distinguish parts of a compound statement`,

		Min_args: 0,
		Max_args: 2,
	}
	AddToCategory("breakpoints", name)
	AddAlias("break", name)
	AddAlias("b", name)
}

func BreakpointCommand(args []string) {
	if len(args) == 1 {
		InfoBreakpointSubcmd()
		return
	}
	name := args[1]
	fn := GetFunction(name)
	if fn != nil {
		if ext := interp.Externals()[name]; ext != nil {
			Msg("Sorry, %s is a built-in external function.", name)
			return
		}
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
		Msg(" Breakpoint %d set in function %s at %s", bpnum, name,
			fmtPos(fn, fn.Pos()))
		return
	}
	line, ok := strconv.Atoi(args[1])
	if ok != nil {
		Errmsg("Don't know yet how to deal with a break that doesn't start with a function or integer")
		return
	}

	column := -1
	if len(args) == 3 {
		foo, ok := strconv.Atoi(args[2])
		if ok != nil {
			Errmsg("Don't know how to deal a non-int argument as 2nd parameter yet")
			return
		}
		column = foo
	}

	fn = curFrame.Fn()
	fset := curFrame.Fset()
	position := curFrame.Position()
	if position.IsValid() {
		filename := position.Filename
		for _, l := range fn.Pkg.Locs() {
			try := fset.Position(l.Pos)
			if try.Filename == filename && line == try.Line {
				if column == -1 || column == try.Column {
					bp := &Breakpoint {
						hits: 0,
						id: len(Breakpoints),
						pos: l.Pos,
						endP: l.Pos,
						ignore: 0,
						kind: "Statement",
						temp: false,
						enabled: true,
					}
					bpnum := BreakpointAdd(bp)
					if l.Trace != nil {
						l.Trace.Breakpoint = true
					} else if l.Fn != nil {
						l.Fn.Breakpoint = true
						bp.kind = "Function"
					} else {
						Errmsg("Internal error setting in file %s line %d, column %d",
							bpnum, filename, line, try.Column)
						return
					}
					Msg("Breakpoint %d set in file %s line %d, column %d", bpnum, filename, line, try.Column)
					return
				}
			}
		}
		suffix := ""
		if column != -1 { suffix = ", column " + args[2] }
		Errmsg("Can't find statement in file %s at line %d%s", filename, line, suffix)
	}
}

func init() {
	name := "delete"
	Cmds[name] = &CmdInfo{
		Fn: DeleteCommand,
		Help: `Delete [bpnum1 ...]

Delete a breakpoint by the number assigned to it.`,

		Min_args: 0,
		Max_args: -1,
	}
	AddToCategory("breakpoints", name)
	// Down the line we'll have abbrevs
	AddAlias("del", name)
}

func DeleteCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			Errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if BreakpointDelete(bpnum) {
				Msg(" Deleted breakpoint %d", bpnum)
			} else {
				Errmsg("Trouble deleting breakpoint %d", bpnum)
			}
		} else {
			Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}

func init() {
	name := "disable"
	Cmds[name] = &CmdInfo{
		Fn: DisableCommand,
		Help: `Disable [bpnum1 ...]

Disable a breakpoint by the number assigned to it.`,

		Min_args: 0,
		Max_args: -1,
	}
	AddToCategory("breakpoints", name)
}

// FIXME: DRY the next two commands.
func DisableCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			Errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if !BreakpointIsEnabled(bpnum) {
				Msg("Breakpoint %d is already disabled", bpnum)
				continue
			}
			if BreakpointDisable(bpnum) {
				Msg("Breakpoint %d disabled", bpnum)
			} else {
				Errmsg("Trouble disabling breakpoint %d", bpnum)
			}
		} else {
			Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}

func init() {
	name := "enable"
	Cmds[name] = &CmdInfo{
		Fn: EnableCommand,
		Help: `enable [bpnum1 ...]

Enable a breakpoint by the number assigned to it.`,

		Min_args: 0,
		Max_args: -1,
	}
	AddToCategory("breakpoints", name)
}

func EnableCommand(args []string) {
	if !argCountOK(1, 1000, args) { return }
	for i:=1; i<len(args); i++ {
		bpnum, ok := strconv.Atoi(args[i])
		if ok != nil {
			Errmsg("Expecting integer breakpoint for argument %d; got %s", i, args[i])
			continue
		}
		if BreakpointExists(bpnum) {
			if BreakpointIsEnabled(bpnum) {
				Msg("Breakpoint %d is already enabled", bpnum)
				continue
			}
			if BreakpointEnable(bpnum) {
				Msg("Breakpoint %d enabled", bpnum)
			} else {
				Errmsg("Trouble enabling breakpoint %d", bpnum)
			}
		} else {
			Errmsg("Breakpoint %d doesn't exist", bpnum)
		}
	}
}
