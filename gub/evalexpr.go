// Copyright 2013-2014 Rocky Bernstein.
// evaluation support for expressions. A bridge betwen eval's evaluation
// and reflect values and interp.Value

package gub

import (
/*
	"errors"
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/types"
*/
	"reflect"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	// "github.com/rocky/go-fish"
	// "github.com/rocky/0xfaded/eval"
	"github.com/rocky/rocky/eval"
)

// interp2reflectVal converts between an interp.Value which the
// interpreter uses and reflect.Value which eval uses. nameVal is used
// to get type information.
func interp2reflectVal(interpVal interp.Value, nameVal ssa2.Value) reflect.Value {
	v := DerefValue(interpVal)
	return reflect.ValueOf(v)
}

// EvalExpr is the top-level call to evaluate a string via 0xfaded/eval.
func EvalExpr(expr string) ([]reflect.Value, error) {
	println("EvalExpr called")
	results, panik, compileErrs := eval.Eval(expr)
	if compileErrs != nil {
		for _, err := range(compileErrs) {
			Errmsg(err.Error())
		}
	} else if panik != nil {
		for _, err := range(compileErrs) {
			Errmsg(err.Error())
		}
	} else {
		return results, nil
	}
	return nil, nil
}

/*******
// Here's our custom ident type check
func CheckIdent(ident *ast.Ident, env eval.Env) (_ *eval.Ident, errs []error) {
	println("gub Check Ident")
	aexpr := &eval.Ident{Ident: ident}
	name := aexpr.Name
	switch name {
	case "nil":
		aexpr.SetConstValue(reflect.Value{})
		aexpr.SetKnownType(reflect.TypeOf(eval.ConstNil))
	case "true":
		aexpr.SetConstValue(reflect.ValueOf(true))
		aexpr.SetKnownType(reflect.TypeOf(true))

	case "false":
		aexpr.SetConstValue(reflect.ValueOf(false))
		aexpr.SetKnownType(reflect.TypeOf(false))
	default:
		fn := curFrame.Fn()
		pkg := fn.Pkg
		// nameVal, interpVal, scopeVal := EnvLookup(curFrame, name, curScope)
		nameVal, interpVal, _ := EnvLookup(curFrame, name, curScope)
		if nameVal != nil {
			println("Found in env")
			val := interp2reflectVal(interpVal, nameVal)
			aexpr.SetKnownType(reflect.TypeOf(val))
			aexpr.SetSource(eval.EnvVar)
		} else if fn := pkg.Func(name); fn != nil {
			println("found in func")
		} else if v := pkg.Var(name); v != nil {
			println("found in var")
		} else if g, ok := curFrame.I().Global(name, pkg); ok {
			println("found in global", g)
		} else if c := pkg.Const(name); c != nil {
			println("found in const")
		} else if pkg := curFrame.I().Program().PackagesByName[name]; pkg != nil {
			println("found in package")
		} else {
			println("id not found")
		}
	}
	return aexpr, errs
}

// EvalIdentExpr extracts a reflect.Vaue for an identifier. The
// boolean return parameter indicates whether the value is typed. The error
// parameter is non-nil if there was an error.
// Note that the parameter ctx is not used here, but is part of the eval
// interface. So we pass that along if we can't find the name here and
// resort to the static evaluation environment compiled into eval.
func EvalIdent(ident *eval.Ident, env eval.Env) (reflect.Value, error) {
	println("evalident")
	name := ident.Name
	if name == "nil" {
		// FIXME: Should this be done first or last?
		return eval.EvalNil, nil
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
			return reflectVal, nil
		} else {
			pkg := curFrame.I().Program().PackagesByName[name]
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
					val := interp2reflectVal(g, v)
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

func InterpVal2Reflect(v interp.Value) (reflect.Value, string) {
	switch v.(type) {
	case bool:
		return reflect.ValueOf(v), ""
	case int:
		return reflect.ValueOf(v), ""
	case int8:
		return reflect.ValueOf(v), ""
	case int16:
		return reflect.ValueOf(v), ""
	case int32:
		return reflect.ValueOf(v), ""
	case int64:
		return reflect.ValueOf(v), ""
	case uint:
		return reflect.ValueOf(v), ""
	case uint8:
		return reflect.ValueOf(v), ""
	case uint16:
		return reflect.ValueOf(v), ""
	case uint32:
		return reflect.ValueOf(v), ""
	case uint64:
		return reflect.ValueOf(v), ""
	case uintptr:
		return reflect.ValueOf(v), ""
	case float32:
		return reflect.ValueOf(v), ""
	case float64:
		return reflect.ValueOf(v), ""
	case complex64:
		return reflect.ValueOf(v), ""
	case complex128:
		return reflect.ValueOf(v), ""
	case string:
		return reflect.ValueOf(v), ""
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
		return reflect.Value{}, "Can't convert"
	}
}

// evalEnv contains a 0xfaded/eval static evaluation environment that
// we can use to fallback to using when the program environment
// doesn't contain a specific package or due to limitations we
// currently have in extracting values.
var evalEnv eval.SimpleEnv

func init() {
	eval.SetCheckIdent(CheckIdent)
	eval.SetEvalIdent(EvalIdent)
	eval.SetEvalSelectorExpr(EvalSelectorExpr)
	// eval.SetUserConversion(myConvertFunc)
	// evalEnv = repl.MakeEvalEnv()
	evalEnv = *eval.MakeSimpleEnv()
}
***/
