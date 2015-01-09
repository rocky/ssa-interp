// Copyright 2015 Rocky Bernstein.
// Encapsulate an SSA Interpreter environment as a
// github.com/0xfaded/env environment.

package interp

import (
	"github.com/rocky/eval"
	"github.com/rocky/ssa-interp"
	"reflect"
	"fmt"
)


type EvalEnv struct{
	static  *eval.SimpleEnv
	prog    *ssa2.Program
    globals *map[ssa2.Value]*Value
	frame   Frame
	curFn   *ssa2.Function
	curPkg  *ssa2.Package
	scope   *ssa2.Scope
}

var reflectNil reflect.Value = reflect.Value{}
type UntypedNil struct {}

func DerefValue(v Value) Value {
	switch v := v.(type) {
	case *Value:
		if v == nil { return nil }
		return *v
	default:
		return v
	}
}
// interp2reflectVal converts between an interp.Value which the
// interpreter uses and reflect.Value which eval uses. nameVal is used
// to get type information.
func interp2reflectVal(interpVal Value) reflect.Value {
	v := DerefValue(interpVal)
	return reflect.ValueOf(v)
}

func interp2reflectType(interpVal Value) reflect.Type {
	v := DerefValue(interpVal)
	return reflect.TypeOf(v)
}

func MakeEnv(static *eval.SimpleEnv, prog *ssa2.Program, fr *Frame) *EvalEnv {
	return &EvalEnv{
		static: static,
		prog: prog,
		frame: *fr,
		curFn: fr.Fn(),
		curPkg: fr.Fn().Pkg,
		scope:  fr.Scope(),
	}
}

func (env EvalEnv) Static() eval.SimpleEnv { return *env.static }

func (env EvalEnv) local(name string) (reflect.Value, bool) {
	nameScope := ssa2.NameScope{
		Name: name,
		Scope: env.scope,
	}
	if i := env.curFn.LocalsByName[nameScope]; i == 0 {
		return reflectNil, false
	} else {
		fmt.Printf("Got value %v %T local %s\n", env.frame.Local(i-1),
			env.frame.Local(i-1), name)
		return interp2reflectVal(env.frame.Local(i-1)), true
	}
}


// The stuff below here are methods to satisfy the eval.Env interface

func (env EvalEnv) Var(name string) reflect.Value {
	if val, ok := env.local(name); ok {
		return val
	} else {
		return reflectNil
	}
}

func (env EvalEnv) Func(name string) reflect.Value {
	pkg := env.curPkg
	if pkg == nil { return reflect.Value{} }
	return interp2reflectVal(pkg.Func(name))
}

func (env EvalEnv) Const(name string) reflect.Value {
	pkg := env.curPkg
	if pkg == nil {
		fmt.Printf("const %s not found in package %s\n", name, pkg)
		return reflect.Value{}
	}
	return interp2reflectVal(pkg.Const(name))
}

func (env EvalEnv) Type(name string) reflect.Type {
	fmt.Println("Looking up type for var %s", name)
	if val, ok := env.local(name); ok {
		return reflect.TypeOf(val)
	} else {
		return env.static.Type(name)
	}
}

func (env EvalEnv) Pkg(name string) eval.Env {
	if pkg := env.prog.PackagesByName[name]; pkg != nil {
		env.curPkg = pkg
	}
	return env
}


// Create a new block scope. Only the behaviour of the returned Env should change
func (env EvalEnv) PushScope() eval.Env {
	return nil
}

// Pop the top block scope. Only the behaviour of the returned Env should change
func (env EvalEnv) PopScope() eval.Env {
	return nil
}

// Add var ident to the top scope. The value is always a pointer value, and this same value should be
// returned by Var(ident). It is up to the implementation how to handle duplicate identifiers.
func (env EvalEnv) AddVar(ident string, v reflect.Value) {
	env.static.Vars[ident] = v
}

// Add const ident to the top scope. It is up to the implementation how to handle duplicate identifiers.
func (env EvalEnv) AddConst(ident string, c reflect.Value) {
	env.static.Consts[ident] = c
}

// Add func ident to the top scope. It is up to the implementation how to handle duplicate identifiers.
func (env EvalEnv) AddFunc(ident string, f reflect.Value) {
	env.static.Funcs[ident] = f
}

// Add type ident to the top scope. It is up to the implementation how to handle duplicate identifiers.
func (env EvalEnv) AddType(ident string, t reflect.Type) {
	env.static.Types[ident] = t
}

// Add pkg to the root scope. It is up to the implementation how to handle duplicate identifiers.
func (env EvalEnv) AddPkg(pkg string, p eval.Env) {
	env.static.Pkgs[pkg] = p
}

func (env EvalEnv) Path() string {
	return env.curPkg.String()
}
