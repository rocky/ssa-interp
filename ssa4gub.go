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

// Package returns the SSA package corresponding to the specified
// type-checker package object.
// It returns nil if no such SSA package has been created.
//
func (prog *Program) Package(pkg *types.Package) *Package {
	return prog.packages[pkg]
}

func (v *Function) EndP() token.Pos            { return v.endP }
func (v *Function) Fset() *token.FileSet       { return v.Prog.Fset }
func (v *Function) NamedResults() []*Alloc     { return v.namedResults }

func (p *Package) Locs() []LocInst { return p.locs }

func (prog *Program) PackageByName(name string) *Package {
	return prog.PackagesByPath[name]
}
