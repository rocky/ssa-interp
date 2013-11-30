package ssa2

import	"go/ast"
// import  "runtime/debug"

func astScope(fn *Function, node ast.Node) *Scope {
	pkg   := fn.Pkg
	scope := pkg.Ast2Scope[node]
	if scope != nil {
		// Probably won't happen
		return scope
	} else {
		scope = pkg.TypeScope2Scope[pkg.info.Scopes[node]]
		if scope != nil {
			pkg.Ast2Scope[node] = scope
		}
	}
	return scope
}

func ParentScope(fn *Function, scope *Scope) *Scope {
	// FIXME: reinstate scope
	// return fn.Pkg.TypeScope2Scope[scope.Scope.Parent()]
	return nil
}
