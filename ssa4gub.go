package ssa2

/*

This file contains definitions beyond ssa.go needed for the gub
debugger. This could be merged into ssa.go but we keep it separate so
as to make diff'ing our ssa.go and the unmodified ssa.go look more
alike.

*/

import (
	"go/token"
	"code.google.com/p/go.tools/go/types"
)


type LocInst struct {
	Pos    token.Pos
	// Fixme: I don't know how to do a C union "Instruction" typecast
	Trace  *Trace
	Fn     *Function
}

// Scopes are attached to basic blocks.  For our purposes, we need a
// types.Scope plus some sort of non-pointer-address name which we can
// repeatably derive. The name is just a preorder traversal number of
// the scope tree for a package. Scope number should be reset for each
// function, but that's more work. I, rocky, believe this really
// should be in ast.scope, but it is what it is.
type Scope struct {
	*types.Scope
	scopeNum int
}

func (s *Alloc)    EndP() token.Pos            { return s.endP }
func (s *DebugRef) EndP() token.Pos            { return s.Expr.End() }
func (s *Defer)    EndP() token.Pos            { return s.Call.endP }
func (s *Go)       EndP() token.Pos            { return s.Call.endP }
func (v *Register) EndP() token.Pos            { return v.endP }
func (s *Ret)      EndP() token.Pos            { return s.endP }
func (v *Function) EndP() token.Pos            { return v.endP }
func (v *Function) Fset() *token.FileSet       { return v.Prog.Fset }
func (v *Function) NamedResults() []*Alloc     { return v.namedResults }
func (v *Register) setEnd(pos token.Pos)       { v.endP = pos }

func (p *Package) Locs() []LocInst { return p.locs }

func (prog *Program) PackageByName(name string) *Package {
	return prog.PackagesByPath[name]
}


func (s *Scope) ScopeNum() int { return s.scopeNum }
