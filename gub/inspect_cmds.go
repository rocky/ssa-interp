// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"ssa-interp/interp"
)

func LocalsCommand(args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 2, args) { return }
	if argc == 0 {
		i := 0
		for _, v := range curFrame.Locals() {
			name := curFrame.Fn().Locals[i].Name()
			msg("%s: %s", name, interp.ToString(v))
			i++
		}
	} else {
		varname := args[1]
		i := 0
		for _, v := range curFrame.Locals() {
			if args[1] == curFrame.Fn().Locals[i].Name() {
				msg("%s %s: %s", varname, curFrame.Fn().Locals[i], interp.ToString(v))
				break
			}
			i++
		}

	}
}

func VariableCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	fn := curFrame.Fn()
	varname := args[1]
	for _, p := range fn.Locals {
		if varname == p.Name() { break }
	}

}

func WhatisCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	name := args[1]
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if fn := pkg.Func(name); fn != nil {
		msg("%s is a function at:", name)
		msg("\t%s", fmtRange(myfn, fn.Pos(), fn.EndP()))

		for _, p := range fn.Params {
			msg("\t%s", p)
		}
		for _, r := range fn.NamedResults() {
			msg("\t%s", r)
		}
	} else if v := pkg.Var(name); v != nil {
		msg("%s is a variable at:", name)
		msg("\t%s", fmtPos(myfn, v.Pos()))
		// msg("Value %s", interp.ToString(v.Value))
	} else if c := pkg.Const(name); c != nil {
		msg("%s is a constant at:", name)
		msg("\t%s", fmtPos(myfn, c.Pos()))
		msg("Value %s", interp.ToString(interp.LiteralValue(c.Value)))
	} else if t := pkg.Type(name); t != nil {
		msg("%s is a type", name)
	} else {
		msg("can't find %s", name)
	}
}
