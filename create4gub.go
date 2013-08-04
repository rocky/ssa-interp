package ssa2

import (
	"code.google.com/p/go.tools/go/types"
)

func assignScopeNum(typesScope *types.Scope, scopeNum int) *Scope{
	scope := &Scope {
		Scope: typesScope,
		scopeNum: scopeNum,
	}
	return scope
}

func AssignScopeNums(pkg *Package, typesScope *types.Scope, scopeNum *int) {
	node  := typesScope.Node()
	scope := assignScopeNum(typesScope, *scopeNum)
	pkg.Ast2Scope[node] = scope
	pkg.TypeScope2Scope[typesScope] = scope
	// num2scope = append(num2scope, scope)
	// switch node.(type) {
	// case *ast.FuncType:
	// 	fmt.Println("+++FuncType")
	// }

	*scopeNum++
	n := scope.NumChildren()
	for i:=0; i<n; i++ {
		child := typesScope.Child(i)
		if child != nil { AssignScopeNums(pkg, child, scopeNum) }
	}
}
