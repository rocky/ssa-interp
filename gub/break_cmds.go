// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint-handling commands
package gub

import (
	"strconv"
)

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
		if bp.Deleted { continue }
		Bpprint(*bp)
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
