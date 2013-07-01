// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
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

func printFuncInfo(fn *ssa2.Function) {
	msg("%s is a function at:", fn.FullName())
	ps := fn.PositionRange()
	if ps == "-" {
		msg("\tsynthetic function (no position)")
	} else {
		msg("\t%s", ps)
	}

	for _, p := range fn.Params {
		msg("\t%s", p)
	}
	for _, r := range fn.NamedResults() {
		msg("\t%s", r)
	}

	if fn.Enclosing != nil {
		section("Parent: %s\n", fn.Enclosing.Name())
	}

	if fn.FreeVars != nil {
		section("Free variables:")
		for i, fv := range fn.FreeVars {
			msg("%3d:\t%s %s", i, fv.Name(), fv.Type())
		}
	}

	if len(fn.Locals) > 0 {
		section("Locals:")
		for i, l := range fn.Locals {
			msg("% 3d:\t%s %s", i, l.Name(), l.Type().Deref())
		}
	}

	// writeSignature(w, f.Name(), f.Signature, f.Params)

	if fn.Blocks == nil {
		msg("\t(external)")
	}
}

func WhatisCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	name := args[1]
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if fn := pkg.Func(name); fn != nil {
		printFuncInfo(fn)
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
