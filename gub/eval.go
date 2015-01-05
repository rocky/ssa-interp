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

func Deref2Str(v interp.Value, ssaVal *ssa2.Value) string {
	return interp.ToInspect(DerefValue(v), ssaVal)
}


func PrintInEnvironment(fr *interp.Frame, name string, nameVal ssa2.Value,
	interpVal interp.Value, scopeVal *ssa2.Scope) bool {
	envStr := ""
	if scopeVal != nil {
		envStr = fmt.Sprintf(" at scope %d", scopeVal.ScopeId())
	}
	Msg("%s is in the environment%s", name, envStr)
	Msg("\t%s = %s", name, Deref2Str(interpVal, &nameVal))
	return true
}

func EnvLookup(fr *interp.Frame, name string,
	scope *ssa2.Scope) (ssa2.Value, interp.Value, *ssa2.Scope) {
	fn := fr.Fn()
	reg := fr.Var2Reg[name]
	for ; scope != nil;  scope = ssa2.ParentScope(fn, scope) {
		nameScope := ssa2.NameScope{
			Name: name,
			Scope: scope,
		}
		if i := fn.LocalsByName[nameScope]; i > 0 {
			nameVal := fn.Locals[i-1]
			val     := fr.Env()[nameVal]
			return nameVal, val, nameVal.Scope
		}
	}
	names := []string{name, reg}
	for _, name := range names {
		for nameVal, val := range fr.Env() {
			if name == nameVal.Name() {
				switch nameVal := nameVal.(type) {
				case *ssa2.Alloc:
					return nameVal, val, nameVal.Scope
				default:
					return nameVal, val, nil
				}
			}
		}
	}
	// FIXME: Why we would find things here and not by the
	// above scope lookup?
	if val := fn.Pkg.Var(name); val != nil {
		return val, nil, nil
	}
	return nil, nil, nil
}

// Could something like this go into interp-ssa?
func GetFunction(name string) *ssa2.Function {
	pkg := curFrame.Fn().Pkg
	ids := strings.Split(name, ".")
	if len(ids) > 1 {
		try_pkg := program.PackagesByName[ids[0]]
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
