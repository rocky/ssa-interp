// Copyright 2013 Rocky Bernstein.
// Debugger commands
package gub

import (
	"fmt"
	"path"
	"strings"
	"sort"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"code.google.com/p/go-columnize"
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

func printConstantInfo(c *ssa2.Constant, name string, pkg *ssa2.Package) {
	mem := pkg.Members[name]
	position := pkg.Prog.Fset.Position(mem.Pos())
	msg("Constant %s is a constant at:", mem.Name())
	msg("  " + ssa2.PositionRange(position, position))
	msg("  %s %s", mem.Type(), interp.ToString(interp.LiteralValue(c.Value)))
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

func printPackageInfo(name string, pkg *ssa2.Package) {
	s := fmt.Sprintf("%s is a package", name)
	mems := ""
	if len(pkg.Members) > 0 {
		different := false
		filename := ""
		fset := curFrame.Fset()
		for _, m := range pkg.Members {
			pos := m.Pos()
			if pos.IsValid() {
				position := fset.Position(pos)
				if len(filename) == 0 {
					filename = position.Filename
				} else if filename != position.Filename {
					different = true
					break
				}
			}
		}
		if len(filename) > 0 {
			if different {filename = path.Dir(filename)}
			s += ": at " + filename
		}
		members := make([]string, len(pkg.Members))
		i := 0
		for k, _ := range pkg.Members {
			members[i] = k
			i++
		}
		sort.Strings(members)
		opts := columnize.DefaultOptions()
		opts.DisplayWidth = maxwidth
		mems = columnize.Columnize(members, opts)
	}
	msg(s)
	if len(mems) > 0 {
		section("Members")
		msg(mems)
	}
}

func printTypeInfo(name string, pkg *ssa2.Package) {
	mem := pkg.Members[name]
	msg("Type %s at:", mem.Type())
	position := pkg.Prog.Fset.Position(mem.Pos())
	msg("  " + ssa2.PositionRange(position, position))
	msg("  %s", mem.Type().Underlying())

	// We display only mset(*T) since its keys
	// are a superset of mset(T)'s keys, though the
	// methods themselves may differ,
	// e.g. promotion wrappers.
	// NB: if mem.Type() is a pointer, mset is empty.
	mset := pkg.Prog.MethodSet(ssa2.Pointer(mem.Type()))
	var keys ssa2.Ids
	for id := range mset {
		keys = append(keys, id)
	}
	sort.Sort(keys)
	for _, id := range keys {
		method := mset[id]
		// TODO(adonovan): show pointerness of receiver of declared method, not the index
		msg("    method %s %s", id, method.Signature)
	}
}

func WhatisCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	name := args[1]
	ids := strings.Split(name, ".")
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if len(ids) > 1 {
		try_pkg := curFrame.I().Program().PackageByName(ids[0])
		if try_pkg != nil { pkg = try_pkg }
		m := pkg.Members[ids[1]]
		if m == nil {
			errmsg("%s is not a member of %s", ids[1], pkg)
			return
		}
		name = ids[1]
	}

	if fn := pkg.Func(name); fn != nil {
		printFuncInfo(fn)
	} else if v := pkg.Var(name); v != nil {
		msg("%s is a variable at:", name)
		msg("  %s", fmtPos(myfn, v.Pos()))
		msg("  %s", v.Type())
		if g, ok := curFrame.I().Global(name, pkg); ok {
			msg("  %s", *g)
		}
	} else if c := pkg.Const(name); c != nil {
		printConstantInfo(c, name, pkg)
	} else if t := pkg.Type(name); t != nil {
		printTypeInfo(name, pkg)
	} else if pkg := curFrame.I().Program().PackageByName(name); pkg != nil {
		printPackageInfo(name, pkg)
	} else {
		errmsg("can't find %s", name)
	}
}
