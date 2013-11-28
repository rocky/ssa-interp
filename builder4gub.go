package ssa2

import	"go/ast"

func astScope(fn *Function, node ast.Node) *Scope {
	// fmt.Printf("scope returning is %d\n", fn.Pkg.Ast2Scope[node].scopeNum)
	return fn.Pkg.Ast2Scope[node]
}

func ParentScope(fn *Function, scope *Scope) *Scope {
	// FIXME: reinstate scope
	// return fn.Pkg.TypeScope2Scope[scope.Scope.Parent()]
	return nil
}
