package ssa2

import	"go/ast"
// import  "runtime/debug"

func astScope(fn *Function, node ast.Node) *Scope {
	pkg   := fn.Pkg
	scope := pkg.TypeScope2Scope[pkg.info.Scopes[node]]
	if scope != nil {
		scope.node = &node
	}
	return scope
}

func ParentScope(fn *Function, scope *Scope) *Scope {
	// FIXME: reinstate scope
	// return fn.Pkg.TypeScope2Scope[scope.Scope.Parent()]
	return nil
}
