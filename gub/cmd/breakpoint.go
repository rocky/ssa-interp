// Copyright 2013 Rocky Bernstein.
// Debugger breakpoint command
package gubcmd

import (
	"strconv"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "breakpoint"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: BreakpointCommand,
		Help: `breakpoint [*fn* | line [column]]

Set a breakpoint. The target can either be a function name as fn pkg.fn
or a line and and optional column number. Specifying a column number
may be useful if there is more than one statement on a line or if you
want to distinguish parts of a compound statement`,

		Min_args: 0,
		Max_args: 2,
	}
	gub.AddToCategory("breakpoints", name)
	gub.AddAlias("break", name)
	gub.AddAlias("b", name)
}

func BreakpointCommand(args []string) {
	if len(args) == 1 {
		InfoBreakpointSubcmd()
		return
	}
	name := args[1]
	fn := gub.GetFunction(name)
	if fn != nil {
		if ext := interp.Externals()[name]; ext != nil {
			gub.Msg("Sorry, %s is a built-in external function.", name)
			return
		}
		interp.SetFnBreakpoint(fn)
		bp := &gub.Breakpoint {
			Hits: 0,
			Id: gub.BreakpointNext(),
			Pos: fn.Pos(),
			EndP: fn.EndP(),
			Ignore: 0,
			Kind: "Function",
			Temp: false,
			Enabled: true,
		}
		bpnum := gub.BreakpointAdd(bp)
		gub.Msg(" Breakpoint %d set in function %s at %s", bpnum, name,
			ssa2.FmtRange(fn, fn.Pos(), fn.EndP()))
		return
	}
	line, ok := strconv.Atoi(args[1])
	if ok != nil {
		gub.Errmsg("Don't know yet how to deal with a break that doesn't start with a function or integer")
		return
	}

	column := -1
	if len(args) == 3 {
		foo, ok := strconv.Atoi(args[2])
		if ok != nil {
			gub.Errmsg("Don't know how to deal a non-int argument as 2nd parameter yet")
			return
		}
		column = foo
	}

	fn = gub.CurFrame().Fn()
	fset := gub.CurFrame().Fset()
	position := gub.CurFrame().Position()
	if position.IsValid() {
		filename := position.Filename
		for _, l := range fn.Pkg.Locs() {
			try := fset.Position(l.Pos())
			if try.Filename == filename && line == try.Line {
				if column == -1 || column == try.Column {
					bp := &gub.Breakpoint {
						Hits: 0,
						Id: gub.BreakpointNext(),
						Pos: l.Pos(),
						EndP: l.Pos(),
						Ignore: 0,
						Kind: "Statement",
						Temp: false,
						Enabled: true,
					}
					bpnum := gub.BreakpointAdd(bp)
					if l.Trace != nil {
						l.Trace.Breakpoint = true
					} else if l.Fn != nil {
						l.Fn.Breakpoint = true
						bp.Kind = "Function"
					} else {
						gub.Errmsg("Internal error setting in file %s line %d, column %d",
							bpnum, filename, line, try.Column)
						return
					}
					gub.Msg("Breakpoint %d set in file %s line %d, column %d", bpnum, filename, line, try.Column)
					return
				}
			}
		}
		suffix := ""
		if column != -1 { suffix = ", column " + args[2] }
		gub.Errmsg("Can't find statement in file %s at line %d%s", filename, line, suffix)
	}
}
