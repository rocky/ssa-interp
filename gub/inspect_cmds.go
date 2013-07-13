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

func LocalsLookup(fr *interp.Frame, name string) int {
	return fr.Fn().LocalsByName[name]
}


func printLocal(fr *interp.Frame, i int) {
	v := fr.Local(i)
	l := fr.Fn().Locals[i]
	msg("%3d: %s %s = %s", i, l.Name(), l.Type().Deref(), interp.ToString(v))
}

func printIfLocal(fr *interp.Frame, varname string) bool {
	if i := LocalsLookup(curFrame, varname); i != 0 {
		printLocal(curFrame, i-1)
		return true
	}
	return false
}



func EnvCommand(args []string) {
	for k, v := range curFrame.Env() {
		switch v := v.(type) {
		case *interp.Value:
			msg("*%s %s = %s", k.Name(), k, interp.ToString(*v))
		default:
			msg("%s %s = %s", k.Name(), k, interp.ToString(v))
		}
	}
}

func LocalsCommand(args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 2, args) { return }
	if argc == 0 {
		for i, _ := range curFrame.Locals() {
			printLocal(curFrame, i)
		}
	} else {
		varname := args[1]
		if printIfLocal(curFrame, varname) {
			return
		}
		// FIXME: This really shouldn't be needed.
		for i, v := range curFrame.Locals() {
			if varname == curFrame.Fn().Locals[i].Name() {
				msg("fixme %s %s: %s", varname, curFrame.Fn().Locals[i], interp.ToString(v))
				break
			}
		}

	}
}

func ParametersCommand(args []string) {
	argc := len(args) - 1
	if !argCountOK(0, 1, args) { return }
	if argc == 0 {
		for i, p := range curFrame.Fn().Params {
			msg("%s %s", curFrame.Fn().Params[i], curFrame.Env()[p])
		}
	} else {
		varname := args[1]
		for i, p := range curFrame.Fn().Params {
			if varname == curFrame.Fn().Params[i].Name() {
				msg("%s %s", curFrame.Fn().Params[i], curFrame.Env()[p])
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

func WhatisName(name string) {
	ids := strings.Split(name, ".")
	myfn  := curFrame.Fn()
	pkg := myfn.Pkg
	if len(ids) > 1 {
		varname := ids[0]
		// local lookup needs to take precedence over package lookup
		if i := LocalsLookup(curFrame, varname); i != 0 {
			errmsg("Sorry, dotted variable lookup for local %s not supported yet", varname)
		} else {
			try_pkg := curFrame.I().Program().PackageByName(varname)
			if try_pkg != nil { pkg = try_pkg }
			m := pkg.Members[ids[1]]
			if m == nil {
				errmsg("%s is not a member of %s", ids[1], pkg)
				return
			}
			name = ids[1]
		}
	}

	if printIfLocal(curFrame, name) {return}
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
		errmsg("Can't find name: %s", name)
	}
}

func init() {
	name := "whatis"
	cmds[name] = &CmdInfo{
		fn: WhatisCommand,
		help: `whatis name

print information about *name* which can include a dotted variable name.
`,
		min_args: 1,
		max_args: 1,
	}
	AddToCategory("inspecting", name)
}

func WhatisCommand(args []string) {
	if !argCountOK(1, 1, args) { return }
	name := args[1]
	WhatisName(name)
}
