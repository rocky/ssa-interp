// Copyright 2013 Rocky Bernstein.
// Lookup and Evaluation routines
package gub

import (
	"strings"
	"github.com/rocky/ssa-interp"
)

// Put this in interp?
func GetFunction(name string) *ssa2.Function {
	pkg := curFrame.Fn().Pkg
	ids := strings.Split(name, ".")
	if len(ids) > 1 {
		try_pkg := curFrame.I().Program().PackageByName(ids[0])
		if try_pkg != nil { pkg = try_pkg }
		m := pkg.Members[ids[1]]
		if m == nil { return nil }
		name = ids[1]
	}
	if fn := pkg.Func(name); fn != nil {
		return fn
	}
	return nil
}
