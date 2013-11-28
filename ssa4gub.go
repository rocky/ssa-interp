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
	"code.google.com/p/go.tools/importer"
)


type LocInst struct {
	pos    token.Pos
	endP    token.Pos
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

type ScopeId uint
type Scope struct {
	*types.Scope
	scopeId ScopeId
}

type NameScope struct {
	Name     string
	Scope    *Scope
}

func (s *Alloc)     EndP() token.Pos            { return s.endP }
func (s *Builtin)   EndP() token.Pos            { return s.endP }
func (s *Capture)   EndP() token.Pos            { return s.endP }
func (s *Const)     EndP() token.Pos            { return s.endP }
func (s *DebugRef)  EndP() token.Pos            { return s.Expr.End() }
func (s *Defer)     EndP() token.Pos            { return s.Call.endP }
func (s *Go)        EndP() token.Pos            { return s.Call.endP }
func (v *Register)  EndP() token.Pos            { return v.endP }
func (s *Return)    EndP() token.Pos            { return s.endP }
func (v *Function)  EndP() token.Pos            { return v.endP }
func (v *Function)  Fset() *token.FileSet       { return v.Prog.Fset }
func (v *Function)  NamedResults() []*Alloc     { return v.namedResults }
func (v *Global)    EndP() token.Pos            { return v.endP }
func (v *LocInst)   EndP() token.Pos            { return v.endP }
func (v *Parameter) EndP() token.Pos            { return v.endP }
func (v *Register)  setEnd(pos token.Pos)       { v.endP = pos }

func (v *LocInst)   Pos() token.Pos             { return v.pos }
func (p *Package)   Locs() []LocInst { return p.locs }
func (p *Package)   Info() *importer.PackageInfo { return p.info }

func (s *Scope) ScopeId() ScopeId { return s.scopeId }
