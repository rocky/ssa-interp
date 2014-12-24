package ssa2

import (
	"golang.org/x/tools/go/types"
)

func assignScopeId(typesScope *types.Scope, scopeId ScopeId) *Scope{
	scope := &Scope {
		Scope: typesScope,
		scopeId: scopeId,
	}
	return scope
}

func AssignScopeIds(pkg *Package, typesScope *types.Scope, scopeId *ScopeId) {

	scope := assignScopeId(typesScope, *scopeId)

	// Setting scope.node is done in builder

	pkg.TypeScope2Scope[typesScope] = scope

	*scopeId++
	n := scope.NumChildren()
	for i:=0; i<n; i++ {
		child := typesScope.Child(i)
		if child != nil { AssignScopeIds(pkg, child, scopeId) }
	}
}
