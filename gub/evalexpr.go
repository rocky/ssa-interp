// Copyright 2013 Rocky Bernstein.
// evaluation support
package gub

import (
	"errors"
	"fmt"
	"reflect"
	"go/parser"
	"code.google.com/p/go.tools/go/exact"
	"github.com/rocky/ssa-interp"
	"github.com/0xfaded/go-interactive"
)

func EvalIdentExpr(ctx *interactive.Ctx, ident *interactive.Ident, env *interactive.Env) (
	*reflect.Value, bool, error) {
	name := ident.Name
	if name == "nil" {
		// FIXME: Should this be done first or last?
		return nil, false, nil
	} else  {
		if _, interpVal, _ := EnvLookup(curFrame, name, curScope); interpVal != nil {
			// FIXME for structures the interpreter has turned this into a slice
			// we need to somehow undo that or keep track of the type name that this
			// came from so we can get record selection correct.
			reflectVal := reflect.ValueOf(DerefValue(interpVal))
			return &reflectVal, false, nil
		} else {
			pkg := curFrame.I().Program().PackagesByName[name]
			if pkg != nil {
				val := reflect.ValueOf(pkg)
				return &val, false, nil
			}
		}
		return nil, false, errors.New(fmt.Sprintf("%s undefined", name))
	}
}

func EvalSelectorExpr(ctx *interactive.Ctx, selector *interactive.SelectorExpr,
	env *interactive.Env) (*reflect.Value, bool, error) {
	var err error
	var x *[]reflect.Value
	if x, _, err = interactive.EvalExpr(ctx, selector.X.(interactive.Expr), env); err != nil {
		return nil, true, err
	}
	sel   := selector.Sel.Name
	x0    := (*x)[0]
	xname := x0.Type().Name()

	if x0.Kind() == reflect.Ptr {
		// println("XXX x0.Type()", x0.Type().String(), "selector name:", sel)
		// Special case for handling packages
		if x0.Type() == reflect.TypeOf(curFrame.Fn().Pkg) {
			pkg := x0.Interface().(*ssa2.Package)
			m := pkg.Members[sel]
			if m == nil {
				return nil, true,
				errors.New(fmt.Sprintf("%s has no field or method %s", pkg, sel))
			}

			if fn := pkg.Func(sel); fn != nil {
				return nil, true,
				errors.New("Can't handle functions yet")
			} else if v := pkg.Var(sel); v != nil {
				if g, ok := curFrame.I().Global(sel, pkg); ok {
					val := reflect.ValueOf(*g)
					return &val, true, nil
				} else {
					return nil, true,
					errors.New(fmt.Sprintf("%s name lookup failed unexpectedly for %s",
						pkg, sel))
				}
			} else if c := pkg.Const(sel); c != nil {
				switch c.Value.Value.Kind() {
				case exact.Int:
					if int64, ok := exact.Int64Val(c.Value.Value); ok {
						val := reflect.ValueOf(int64)
						return &val, true, nil
					} else {
						return nil, true,
						errors.New("Can't convert to int64")
					}
				default:
					val := reflect.ValueOf(c.Value)
					return &val, true, nil
				}
			} else if t := pkg.Type(sel); t != nil {
				return nil, true,
				errors.New("Can't handle types yet")
			}
			// FIXME
		} else if !x0.IsNil() && x0.Elem().Kind() == reflect.Struct {
			x0 = x0.Elem()
		}
	}

	if x0.Type().String() == "interp.structure" {
		err = errors.New("selection for structures and interfaces not supported yet")
	} else {
		println("XXX", x0.Type().Kind().String(), x0.Type().String())
		err = errors.New(fmt.Sprintf("%s.%s undefined (%s has no field or method %s)",
			xname, sel, xname, sel))
	}
	return nil, true, err
}

func makeEnv() *interactive.Env {
	return &interactive.Env {
		Vars: make(map[string] reflect.Value),
		Consts: make(map[string] reflect.Value),
		Funcs: make(map[string] reflect.Value),
		Types: make(map[string] reflect.Type),
		Pkgs: make(map[string] interactive.Pkg),
	}
}

func EvalExprInteractive(expr string) (*[]reflect.Value, error) {
	env := makeEnv()
	ctx := &interactive.Ctx{expr}
	if e, err := parser.ParseExpr(expr); err != nil {
		Errmsg("Failed to parse expression '%s' (%v)\n", expr, err)
		return nil, err
	} else if cexpr, errs := interactive.CheckExpr(ctx, e, env); len(errs) != 0 {
		Errmsg("Error checking expression '%s' (%v)\n", expr, errs)
		return nil, errs[0]
	} else {
		results, _, err := interactive.EvalExpr(ctx, cexpr, env)
		if err != nil {
			Errmsg("Error evaluating expression '%s' (%v)\n", expr, err)
			return nil, err
		} else {
			return results, nil
		}
	}
	return nil, nil
}

func init() {
	interactive.SetEvalIdentExprCallback(EvalIdentExpr)
	interactive.SetEvalSelectorExprCallback(EvalSelectorExpr)
}
