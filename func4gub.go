package ssa2

/*

This file contains definitions beyond func.go need for the gub
debugger. This could be merged into func.go but we keep it separate so
as to make diff'ing our ssa.go and the unmodified ssa.go look more
alike.

*/


// Return the starting position of function f or "-" if no position found
func (f *Function) Position() string {
	if pos := f.Pos(); pos.IsValid() {
		return f.Prog.Fset.Position(pos).String()
	}
	return "-"
}

func (f *Function) PositionRange() string {
	if start := f.pos; start.IsValid() {
		fset := f.Prog.Fset
		end  := f.endP
		if !end.IsValid() { end = start }
		return PositionRange(fset.Position(start), fset.Position(end))
	}
	return "-"
}
