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
	"code.google.com/p/go.tools/go/types"
	"code.google.com/p/go-columnize"
)

// deref returns a pointer's element type; otherwise it returns typ.
// FIXME: this is from ssa-interp/util.go. DRY
func deref(typ types.Type) types.Type {
	if p, ok := typ.Underlying().(*types.Pointer); ok {
		return p.Elem()
	}
	return typ
}


func LocalsLookup(fr *interp.Frame, name string, scope *ssa2.Scope) uint {
	nameScope := ssa2.NameScope{
		Name: name,
		Scope: scope,
	}
	return fr.Fn().LocalsByName[nameScope]
}


func PrintLocal(fr *interp.Frame, i uint) {
	fn := fr.Fn()
	v  := fr.Local(i)
	l  := fn.Locals[i]
	Msg("%3d:\t%s %s = %s scope %d %s", i, l.Name(), deref(l.Type()),
		interp.ToInspect(v), l.Scope.ScopeId(), ssa2.FmtRange(fn, l.Pos(), l.EndP()))
}

func PrintIfLocal(fr *interp.Frame, varname string) bool {
	if i := LocalsLookup(curFrame, varname, curScope); i != 0 {
		PrintLocal(curFrame, i-1)
		return true
	}
	return false
}

func InfoArgsSubcmd(args []string) {
	argc := len(args) - 2
	if !argCountOK(1, 2, args) { return }
	if argc == 0 {
		for i, p := range curFrame.Fn().Params {
			Msg("%s %s", curFrame.Fn().Params[i], interp.ToInspect(curFrame.Env()[p]))
		}
	} else {
		varname := args[2]
		for i, p := range curFrame.Fn().Params {
			if varname == curFrame.Fn().Params[i].Name() {
				Msg("%s %s", curFrame.Fn().Params[i], interp.ToInspect(curFrame.Env()[p]))
				break
			}
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

func printConstantInfo(c *ssa2.NamedConst, name string, pkg *ssa2.Package) {
	mem := pkg.Members[name]
	position := pkg.Prog.Fset.Position(mem.Pos())
	Msg("Constant %s is a constant at:", mem.Name())
	Msg("  " + ssa2.PositionRange(position, position))
	Msg("  %s %s", mem.Type(), interp.ToInspect(c.Value))
}

func printFuncInfo(fn *ssa2.Function) {
	Msg("%s is a function at:", fn.String())
	ps := fn.PositionRange()
	if ps == "-" {
		Msg("\tsynthetic function (no position)")
	} else {
		Msg("\t%s", ps)
	}

	for _, p := range fn.Params {
		Msg("\t%s", p)
	}
	for _, r := range fn.NamedResults() {
		Msg("\t%s", r)
	}

	if fn.Enclosing != nil {
		Section("Parent: %s\n", fn.Enclosing.Name())
	}

	if fn.FreeVars != nil {
		Section("Free variables:")
		for i, fv := range fn.FreeVars {
			Msg("%3d:\t%s %s", i, fv.Name(), fv.Type())
		}
	}

	if len(fn.Locals) > 0 {
		Section("Locals:")
		for i, l := range fn.Locals {
			Msg(" %3d:\t%s %s", i, l.Name(), deref(l.Type()))
		}
	}

	// writeSignature(w, f.Name(), f.Signature, f.Params)

	if fn.Blocks == nil {
		Msg("\t(external)")
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
		opts.DisplayWidth = Maxwidth
		mems = columnize.Columnize(members, opts)
	}
	Msg(s)
	if len(mems) > 0 {
		Section("Members")
		Msg(mems)
	}
}

// func printTypeInfo(name string, pkg *ssa2.Package) {
// 	mem := pkg.Members[name]
// 	Msg("Type %s at:", mem.Type())
// 	position := pkg.Prog.Fset.Position(mem.Pos())
// 	Msg("  " + ssa2.PositionRange(position, position))
// 	Msg("  %s", mem.Type().Underlying())

// 	// We display only mset(*T) since its keys
// 	// are a superset of mset(T)'s keys, though the
// 	// methods themselves may differ,
// 	// e.g. promotion wrappers.
// 	// NB: if mem.Type() is a pointer, mset is empty.
// 	mset := pkg.Prog.MethodSet(types.NewPointer(mem.Type()))
// 	var keys []string
// 	for id := range mset {
// 		keys = append(keys, id)
// 	}
// 	sort.Strings(keys)
// 	for _, id := range keys {
// 		method := mset[id]
// 		// TODO(adonovan): show pointerness of receiver of declared method, not the index
// 		Msg("    method %s %s", id, method.Signature)
// 	}
// }

func WhatisName(name string) {
	ids := strings.Split(name, ".")
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if len(ids) > 1 {
		varname := ids[0]
		// local lookup needs to take precedence over package lookup
		if i := LocalsLookup(curFrame, varname, curScope); i != 0 {
			Errmsg("Sorry, dotted variable lookup for local %s not supported yet", varname)
		} else {
			try_pkg := curFrame.I().Program().PackageByName(varname)
			if try_pkg != nil { pkg = try_pkg }
			m := pkg.Members[ids[1]]
			if m == nil {
				Errmsg("%s is not a member of %s", ids[1], pkg)
				return
			}
			name = ids[1]
		}
	}

	if PrintInEnvironment(curFrame, name) {return}
	if PrintIfLocal(curFrame, name)       {return}
	if fn := pkg.Func(name); fn != nil {
		printFuncInfo(fn)
	} else if v := pkg.Var(name); v != nil {
		Msg("%s is a variable at:", name)
		Msg("  %s", fmtPos(myfn, v.Pos()))
		Msg("  %s", v.Type())
		if g, ok := curFrame.I().Global(name, pkg); ok {
			Msg("  %s", *g)
		}
	} else if c := pkg.Const(name); c != nil {
		printConstantInfo(c, name, pkg)
	// } else if t := pkg.Type(name); t != nil {
	// 	printTypeInfo(name, pkg)
	} else if pkg := curFrame.I().Program().PackageByName(name); pkg != nil {
		printPackageInfo(name, pkg)
	} else {
		Errmsg("Can't find name: %s", name)
	}
}
