package ssa2

import (
	"code.google.com/p/go.tools/go/types"
)

func assignScopeId(typesScope *types.Scope, scopeId ScopeId) *Scope{
	scope := &Scope {
		Scope: typesScope,
		scopeId: scopeId,
	}
	return scope
}

func AssignScopeIds(pkg *Package, typesScope *types.Scope, scopeId *ScopeId) {

	///FIXME rocky Node moved around in the types interface
	///FIXME node  := typesScope.Node()
	scope := assignScopeId(typesScope, *scopeId)

	//FIXME  pkg.Ast2Scope[node] = scope

	pkg.TypeScope2Scope[typesScope] = scope

	// num2scope = append(num2scope, scope)
	// switch node.(type) {
	// case *ast.FuncType:
	// 	fmt.Println("+++FuncType")
	// }

	*scopeId++
	n := scope.NumChildren()
	for i:=0; i<n; i++ {
		child := typesScope.Child(i)
		if child != nil { AssignScopeIds(pkg, child, scopeId) }
	}
}
