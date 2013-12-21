// Copyright 2013 Rocky Bernstein.
// evaluation support
package gub

import (
	"errors"
	"fmt"
	"reflect"
	"go/parser"
	"github.com/0xfaded/go-interactive"
)

func EvalIdentExpr(ctx *interactive.Ctx, ident *interactive.Ident, env *interactive.Env) (
	*reflect.Value, bool, error) {
	name := ident.Name
	if name == "nil" {
		// FIXME: Should this be done first or last?
		return nil, false, nil
	} else  {
		println("XXX name is", name)
		if _, interpVal, _ := EnvLookup(curFrame, name, curScope); interpVal != nil {
			reflectVal := reflect.ValueOf(DerefValue(interpVal))
			println("XXX val is", reflectVal.String())
			return &reflectVal, false, nil
		} else {
			pkg := curFrame.I().Program().PackagesByName[name]
			if pkg != nil {
				fmt.Printf("Got %s\n", pkg)
				val := reflect.ValueOf(pkg)
				return &val, false, nil
			}
		}
		return nil, false, errors.New(fmt.Sprintf("%s undefined", name))
	}
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
}
