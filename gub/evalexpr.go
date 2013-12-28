// Copyright 2013 Rocky Bernstein.
// evaluation support for expressions. A bridge betwen eval's evaluation
// and reflect values and interp.Value
package gub

import (
	"errors"
	"fmt"
	"reflect"
	"go/parser"
	"code.google.com/p/go.tools/go/exact"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/go-fish"
	"github.com/0xfaded/eval"
)

// Convert between an interp.Value which the interpreter uses and reflect.Value which
// eval uses. nameVal is used to get type information.
func interp2reflectVal(interpVal interp.Value, nameVal ssa2.Value) reflect.Value {
	v := DerefValue(interpVal)
	// println("XXX 1 type",  interp.Type(v))
	// println("XXX 1 value", interp.ToInspect(v))
	return reflect.ValueOf(v)
}

func EvalIdentExpr(ctx *eval.Ctx, ident *eval.Ident, env *eval.Env) (
	*reflect.Value, bool, error) {
	name := ident.Name
	if name == "nil" {
		// FIXME: Should this be done first or last?
		return nil, false, nil
	} else  {
		if nameVal, interpVal, _ := EnvLookup(curFrame, name, curScope); interpVal != nil {
			// FIXME for structures the interpreter has turned this into a slice
			// we need to somehow undo that or keep track of the type name that this
			// came from so we can get record selection correct.
			reflectVal := interp2reflectVal(interpVal, nameVal)
			return &reflectVal, false, nil
		} else {
			pkg := curFrame.I().Program().PackagesByName[name]
			if pkg != nil {
				val := reflect.ValueOf(pkg)
				return &val, false, nil
			}
		}
		// Fall back to using eval's corresponding routine. That way
		// we get access to its builtin functions which I don't support here yet.
		// Also, can access packages that weren't imported by this running program
		// but were in eval. For example, the running interpreter program might not
		/// have imported "fmt", but eval definately does.
		return eval.EvalIdentExpr(ctx, ident, env)
	}
}

func EvalSelectorExpr(ctx *eval.Ctx, selector *eval.SelectorExpr,
	env *eval.Env) (*reflect.Value, bool, error) {
	var err error
	var x *[]reflect.Value
	if x, _, err = eval.EvalExpr(ctx, selector.X.(eval.Expr), env); err != nil {
		return nil, true, err
	}
	sel   := selector.Sel.Name
	x0    := (*x)[0]
	xname := x0.Type().Name()

	if x0.Kind() == reflect.Ptr {
		// Special case for handling packages
		if x0.Type() == reflect.TypeOf(curFrame.Fn().Pkg) {
			pkg := x0.Interface().(*ssa2.Package)

			if fn := pkg.Func(sel); fn != nil {
				// Can't handle interp functions yet, but perhaps it happens
				// to be a function in the static eval environment
				pkg_name := pkg.Object.Name()
				pkg_env := env.Pkgs[pkg_name]
				if fn, ok := pkg_env.Funcs[sel]; ok {
					return &fn, true, nil
				} else {
					return nil, true,
					errors.New("Can't handle functions that are not in 0xfaded/eval yet")

				}
			} else if v := pkg.Var(sel); v != nil {
				if g, ok := curFrame.I().Global(sel, pkg); ok {
					val := interp2reflectVal(g, v)
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
		// println("XXX", x0.Type().Kind().String(), x0.Type().String())
		err = errors.New(fmt.Sprintf("%s.%s undefined (%s has no field or method %s)",
			xname, sel, xname, sel))
	}
	return nil, true, err
}

func makeNullEnv() *eval.Env {
	return &eval.Env {
		Name: "NullEnvironment",
		Vars: make(map[string] reflect.Value),
		Consts: make(map[string] reflect.Value),
		Funcs: make(map[string] reflect.Value),
		Types: make(map[string] reflect.Type),
		Pkgs: make(map[string] eval.Pkg),
	}
}

func EvalExpr(expr string) (*[]reflect.Value, error) {
	env := &evalEnv
	ctx := &eval.Ctx{expr}
	if e, err := parser.ParseExpr(expr); err != nil {
		Errmsg("Failed to parse expression '%s' (%v)\n", expr, err)
		return nil, err
	} else if cexpr, errs := eval.CheckExpr(ctx, e, env); len(errs) != 0 {
		Errmsg("Error checking expression '%s' (%v)\n", expr, errs)
		return nil, errs[0]
	} else {
		results, _, err := eval.EvalExpr(ctx, cexpr, env)
		if err != nil {
			Errmsg("Error evaluating expression '%s' (%v)\n", expr, err)
			return nil, err
		} else {
			return results, nil
		}
	}
	return nil, nil
}

// FIXME should an interp2reflect function be in interp?
var myConvertFunc = func (r reflect.Value, rtyped bool) (reflect.Value, bool, error) {
	switch v := r.Interface().(type) {
	case bool:
		return reflect.ValueOf(v), true, nil
	case int:
		return reflect.ValueOf(v), true, nil
	case int8:
		return reflect.ValueOf(v), true, nil
	case int16:
		return reflect.ValueOf(v), true, nil
	case int32:
		return reflect.ValueOf(v), true, nil
	case int64:
		return reflect.ValueOf(v), true, nil
	case uint:
		return reflect.ValueOf(v), true, nil
	case uint8:
		return reflect.ValueOf(v), true, nil
	case uint16:
		return reflect.ValueOf(v), true, nil
	case uint32:
		return reflect.ValueOf(v), true, nil
	case uint64:
		return reflect.ValueOf(v), true, nil
	case uintptr:
		return reflect.ValueOf(v), true, nil
	case float32:
		return reflect.ValueOf(v), true, nil
	case float64:
		return reflect.ValueOf(v), true, nil
	case complex64:
		return reflect.ValueOf(v), true, nil
	case complex128:
		return reflect.ValueOf(v), true, nil
	case string:
		return reflect.ValueOf(v), true, nil
	// case map[Value]Value:
	// 	return "map[Value]Value"
	// case *hashmap:
	// 	return "*hashmap"
	// case chan Value:
	// 	return "chan Value"
	// case *Value:
	// 	return "*Value"
	// case iface:
	// 	return "iface"
	// case structure:
	// 	return "structure"
	// case array:
	// 	return "array"
	// case []Value:
	// 	return "[]Value"
	// case *ssa2.Function:
	// 	return "*ssa2.Function"
	// case *ssa2.Builtin:
	// 	return "*ssa2.Builtin"
	// case *closure:
	// 	return "*closure"
	// case rtype:
	// 	return "rtype"
	// case tuple:
	// 	return "tuple"
	default:
		return r, rtyped, nil
	}

	return r, rtyped, nil
}

var evalEnv eval.Env

func init() {
	eval.SetEvalIdentExprCallback(EvalIdentExpr)
	eval.SetEvalSelectorExprCallback(EvalSelectorExpr)
	eval.SetUserConversion(myConvertFunc)
	evalEnv = repl.MakeEvalEnv()
}
