// Copyright 2013-2014 Rocky Bernstein.
// evaluation support for expressions. A bridge betwen eval's evaluation
// and reflect values and interp.Value

package gub

import (
	"errors"
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/types"
	"reflect"
	"github.com/rocky/ssa-interp/interp"
	// "github.com/rocky/go-fish"
	// "github.com/0xfaded/eval"
	"github.com/rocky/eval"
)

type knownType []reflect.Type

// GubEvalEnv is the static evaluation environment that we can use to
// fallback to using when the program environment doesn't contain a
// specific package or due to limitations we currently have in
// extracting values.
// var GubEvalEnv eval.Env = eval.MakeSimpleEnv()
var EvalEnv eval.Env

// interp2reflectVal converts between an interp.Value which the
// interpreter uses and reflect.Value which eval uses. nameVal is used
// to get type information.
func interp2reflectVal(interpVal interp.Value) reflect.Value {
	v := DerefValue(interpVal)
	return reflect.ValueOf(v)
}

func interp2reflectType(interpVal interp.Value) reflect.Type {
	v := DerefValue(interpVal)
	return reflect.TypeOf(v)
}

// EvalExpr is the top-level call to evaluate a string via 0xfaded/eval.
func EvalExpr(expr string) ([]reflect.Value, error) {
	results, panik, compileErrs := eval.EvalEnv(expr, EvalEnv)
	if compileErrs != nil {
		Errmsg("Compile error(s):" )
		for _, err := range(compileErrs) {
			Msg(err.Error())
		}
	} else if panik != nil {
		Errmsg("Evaluation error: %s\n", panik.Error())
	} else {
		return results, nil
	}
	return nil, nil
}

// Here's our custom ident type check
func CheckIdent(ident *ast.Ident, env eval.Env) (_ *eval.Ident, errs []error) {
	aexpr := &eval.Ident{Ident: ident}
	name := aexpr.Name
	fmt.Printf("CheckIdent name: %s\n", name)
	switch name {
	case "nil", "true", "false":
		return eval.CheckIdent(ident, env)
	default:
		if v := env.Var(aexpr.Name); v.IsValid() {
			knowntype := knownType{v.Type()}
 			aexpr.SetKnownType(knowntype)
			aexpr.SetSource(eval.EnvVar)
			return aexpr, errs
		} else if v := env.Func(aexpr.Name); v.IsValid() {
			aexpr.SetKnownType(knownType{v.Type()})
			aexpr.SetSource(eval.EnvFunc)
			return aexpr, errs
		} else if v := env.Const(aexpr.Name); v.IsValid() {
			if n, ok := v.Interface().(*eval.ConstNumber); ok {
				aexpr.SetKnownType(knownType{n.Type})
			} else {
				aexpr.SetKnownType(knownType{v.Type()})
			}
			aexpr.SetConstValue(eval.ConstValueOf(v.Interface()))
			aexpr.SetSource(eval.EnvConst)
			return aexpr, errs
		} else {
			evalEnv := env.(interp.EvalEnv)
			return eval.CheckIdent(ident, evalEnv.Static())
		}
	}
	return aexpr, errs
}

// EvalIdentExpr extracts a reflect.Value for an identifier. The
// boolean return parameter indicates whether the value is typed. The error
// parameter is non-nil if there was an error.
// Note that the parameter ctx is not used here, but is part of the eval
// interface. So we pass that along if we can't find the name here and
// resort to the static evaluation environment compiled into eval.
func EvalIdent(ident *eval.Ident, env eval.Env) (reflect.Value, error) {
	name := ident.Name
	fmt.Printf("Evaldent name: %s\n", name)
	if name == "nil" {
		// FIXME: Should this be done first or last?
		return eval.EvalNil, nil
	} else  {
		fn := curFrame.Fn()
		pkg := fn.Pkg
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
			reflectVal := interp2reflectVal(interpVal)
			return reflectVal, nil
		} else if c := pkg.Const(name); c != nil {
			reflectVal := reflect.ValueOf(DerefValue(c.Value))
			return reflectVal, nil
		} else {
			pkg := PkgLookup(name)
			if pkg != nil {
				val := reflect.ValueOf(pkg)
				return val, nil
			}
		}
		// Fall back to using eval's corresponding routine. That way
		// we get access to its builtin functions which I don't support here yet.
		// Also, can access packages that weren't imported by this running program
		// but were in eval. For example, the running interpreter program might not
		/// have imported "fmt", but eval definately does.
		return eval.EvalIdent(ident, env)
	}
}

func pkgEvalIdent(ident *eval.Ident, pkgName string) (reflect.Value, error) {
	if ident.IsConst() {
		return ident.Const(), nil
	}

	name := ident.Name
	switch ident.Source() {
	case eval.EnvVar:
		pkg := PkgLookup(pkgName)
		if pkg != nil {
			if m := pkg.Members[name]; m == nil {
				err := errors.New(
					fmt.Sprintf("%s is not a member of %s", name, pkg))
				return reflect.Value{}, err
			}
			if fn := pkg.Func(name); fn != nil {
				err := errors.New(
					fmt.Sprintf("can't handle functions yet; %s.%s", pkg, name))
				return reflect.Value{}, err
			}
			if v := pkg.Var(name); v != nil {
				if g, ok := curFrame.I().Global(name, pkg); ok {
					return interp2reflectVal(g), nil
				}
			}
			if c := pkg.Const(name); c != nil {
				return interp2reflectVal(DerefValue(c.Value)), nil
			}
			err := errors.New(
				fmt.Sprintf("Don't know what to do with %s.%s yet", pkg, name))
			return reflect.Value{}, err
		} else {
			return reflect.Value{}, errors.New("We only handle package members right now")
			// for searchEnv := env; searchEnv != nil; searchEnv = searchEnv.PopScope() {
			// 	if v := searchEnv.Var(name); v.IsValid() {
			// 		return v.Elem(), nil
			// 	}
			// }
		}
	case eval.EnvFunc:
		println("Can't handle functions yet")
		// for searchEnv := env; searchEnv != nil; searchEnv = searchEnv.PopScope() {
		// 	if v := searchEnv.Func(name); v.IsValid() {
		// 		return v, nil
		// 	}
		// }
	}
	return reflect.Value{}, errors.New("Something went wrong")
}

// Here's our custom selector lookup.
func EvalSelectorExpr(selector *eval.SelectorExpr, env eval.Env) (reflect.Value, error) {
	println("custom EvalSelectorExpr called")

	if pkgName := selector.PkgName(); pkgName != "" {
		fmt.Printf("calling pkgEvalIdent with %v and pkg %s\n", selector.Sel, pkgName)
		return pkgEvalIdent(selector.Sel, pkgName)
	}

	vs, err := eval.EvalExpr(selector.X, env)
	if err != nil {
		return reflect.Value{}, err
	}

	v := vs[0]
	t := v.Type()
	if selector.Field() != nil {
		if t.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		return v.FieldByIndex(selector.Field()), nil
	}

	if selector.IsPtrReceiver() {
		v = v.Addr()
	}
	return v.Method(selector.Method()), nil
}

/*
func EvalSelectorExpr(selector *eval.SelectorExpr,env eval.Env) (
	reflect.Value, error) {
	var err error
	var x []reflect.Value
	if x, err = eval.EvalExpr(selector.X.(eval.Expr), env); err != nil {
		return eval.EvalNil, err
	}
	sel   := selector.Sel.Name
	x0    := x[0]

	if x0.Kind() == reflect.Ptr {
		// Special case for handling packages
		if x0.Type() == reflect.TypeOf(curFrame.Fn().Pkg) {
			pkg := x0.Interface().(*ssa2.Package)

			if fn := pkg.Func(sel); fn != nil {
				// Can't handle interp functions yet, but perhaps it happens
				// to be a function in the static eval environment
				pkg_name := pkg.Object.Name()
				pkg_env := env.Pkg(pkg_name)
				if fn := pkg_env.Func(sel); fn != eval.EvalNil {
					return fn, nil
				} else {
					return eval.EvalNil,
					errors.New("Can't handle functions that are not in 0xfaded/eval yet")

				}
			} else if v := pkg.Var(sel); v != nil {
				if g, ok := curFrame.I().Global(sel, pkg); ok {
					val := interp2reflectVal(g)
					return val, nil
				} else {
					return eval.EvalNil,
					errors.New(fmt.Sprintf("%s name lookup failed unexpectedly for %s",
						pkg, sel))
				}
			} else if c := pkg.Const(sel); c != nil {
				switch c.Value.Value.Kind() {
				case exact.Int:
					if int64, ok := exact.Int64Val(c.Value.Value); ok {
						val := reflect.ValueOf(int64)
						return val, nil
					} else {
						return eval.EvalNil, errors.New("Can't convert to int64")
					}
				default:
					val := reflect.ValueOf(c.Value)
					return val, nil
				}
			} else if t := pkg.Type(sel); t != nil {
				return eval.EvalNil, errors.New("Can't handle types yet")
			}
		} else if !x0.IsNil() && x0.Elem().Kind() == reflect.Struct {
			x0 = x0.Elem()
		}
	}

	if record, ok := x0.Interface().(interp.Structure); ok {
		if field, err := record.FieldByName(sel); err == nil {
			retVal := reflect.ValueOf(field)
			return retVal, nil
		}
	}
	return eval.EvalSelectorExpr(selector, env)
}
*/

// FIXME should an myConvertFunc be in interp?

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

func init() {
	eval.SetCheckIdent(CheckIdent)
	eval.SetEvalIdent(EvalIdent)
	eval.SetEvalSelectorExpr(EvalSelectorExpr)
	// eval.SetUserConversion(myConvertFunc)
}
