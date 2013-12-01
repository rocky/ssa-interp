// Copyright 2013 Rocky Bernstein.
// evaluation support
package gub

import (
	"fmt"
	"strings"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func DerefValue(v interp.Value) interp.Value {
	switch v := v.(type) {
	case *interp.Value:
		if v == nil { return nil }
		return *v
	default:
		return v
	}
}

func Deref2Str(v interp.Value) string {
	return interp.ToInspect(DerefValue(v))
}


func PrintInEnvironment(fr *interp.Frame, name string) bool {
	if k, v, scope := EnvLookup(fr, name, curScope); k != nil {
		envStr := ""
		if scope != nil {
			envStr = fmt.Sprintf(" at scope %d", scope.ScopeId())
		}
		Msg("%s is in the environment%s", name, envStr)
		Msg("\t%s = %s", k, DerefValue(v))
		return true
	} else {
		Errmsg("Name %s not found in environment", name)
		return false
	}
}

func EnvLookup(fr *interp.Frame, name string,
	scope *ssa2.Scope) (ssa2.Value, string, *ssa2.Scope) {
	fn := fr.Fn()
	reg := fr.Var2Reg[name]
	for ; scope != nil;  scope = ssa2.ParentScope(fn, scope) {
		nameScope := ssa2.NameScope{
			Name: name,
			Scope: scope,
		}
		if i := fn.LocalsByName[nameScope]; i > 0 {
			k := fn.Locals[i-1]
			v := Deref2Str(fr.Env()[k])
			return k, v, k.Scope
		}
	}
	names := []string{name, reg}
	for _, name := range names {
		for k, v := range fr.Env() {
			if name == k.Name() {
				v := Deref2Str(v)
				switch k := k.(type) {
				case *ssa2.Alloc:
					return k, v, k.Scope
				default:
					return k, v, nil
				}
			}
		}
	}
	// FIXME: Why we would find things here and not by the
	// above scope lookup?
	if v := fn.Pkg.Var(name); v != nil {
		return v, "", nil
	}
	return nil, "", nil
}

// Could something like this go into interp-ssa?
func GetFunction(name string) *ssa2.Function {
	pkg := curFrame.Fn().Pkg
	ids := strings.Split(name, ".")
	if len(ids) > 1 {
		try_pkg := curFrame.I().Program().PackagesByName[ids[0]]
		if try_pkg != nil { pkg = try_pkg }
		m := pkg.Members[ids[1]]
		if m == nil { return nil }
		name = ids[1]
	}
	if fn := pkg.Func(name); fn != nil {
		return fn
	}
	return nil
}
