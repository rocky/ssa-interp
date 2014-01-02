// Copyright 2013-2014 Rocky Bernstein.
// evaluation support for expressions. A bridge betwen eval's evaluation
// and reflect values and interp.Value

package gub

import (
	"errors"
	"fmt"
	"reflect"
	"go/parser"
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/types"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"github.com/rocky/go-fish"
	"github.com/0xfaded/eval"
)

// interp2reflectVal converts between an interp.Value which the
// interpreter uses and reflect.Value which eval uses. nameVal is used
// to get type information.
func interp2reflectVal(interpVal interp.Value, nameVal ssa2.Value) reflect.Value {
	v := DerefValue(interpVal)
	return reflect.ValueOf(v)
}

// EvalIdentExpr extracts a reflect.Vaue for an identifier. The
// boolean return parameter indicates whether the value is typed. The error
// parameter is non-nil if there was an error.
// Note that the parameter ctx is not used here, but is part of the eval
// interface. So we pass that along if we can't find the name here and
// resort to the static evaluation environment compiled into eval.
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
			ival := DerefValue(interpVal)
			if record, ok := ival.(interp.Structure); ok {
				typ := deref(nameVal.Type()).Underlying()
				if t, ok := typ.(*types.Struct); ok {
					for i := 0; i< record.NumField(); i++ {
						name := t.Field(i).Name()
						record.SetName(i, name)
					}
				}
			}
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
		} else if !x0.IsNil() && x0.Elem().Kind() == reflect.Struct {
			x0 = x0.Elem()
		}
	}

	if record, ok := x0.Interface().(interp.Structure); ok {
		if field, err := record.FieldByName(sel); err == nil {
			retVal := reflect.ValueOf(field)
			return &retVal, true, nil
		}
	}
	return eval.EvalSelectorExpr(ctx, selector, env)
}

// makeNullEnv creates an empty evaluation environment.
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

// EvalExpr is the top-level call to evaluate a string via 0xfaded/eval.
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

// myConvertFunc is used to convert a reflect-encoded interp.Value into
// a reflect.Value. This is needed because values of composites are interp.Values
// not reflect.Values
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

// evalEnv contains a 0xfaded/eval static evaluation environment that
// we can use to fallback to using when the program environment
// doesn't contain a specific package or due to limitations we
// currently have in extracting values.
var evalEnv eval.Env

func init() {
	eval.SetEvalIdentExprCallback(EvalIdentExpr)
	eval.SetEvalSelectorExprCallback(EvalSelectorExpr)
	eval.SetUserConversion(myConvertFunc)
	evalEnv = repl.MakeEvalEnv()
}
