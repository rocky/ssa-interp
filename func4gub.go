package ssa2

import "fmt"

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

func DisasmInst(instr Instruction, width int) string {

	s := "\t"
	switch v := instr.(type) {
	case Value:
		l := width
		// Left-align the instruction.
		if name := v.Name(); name != "" {
			lhs := name + " = "
			l -= len(lhs)
			s += lhs
		}
		rhs := instr.String()
		s += rhs
		l -= len(rhs)
		// Right-align the type.
		if t := v.Type(); t != nil {
			s += fmt.Sprintf(" %*s", l-10, t)
		}
	case *Store:
		// fmt.Printf("found a store %s\n", v)
		// if v.Scope != nil {
		// 	println("got store scope %d", v.Scope.scopeNum)
		// }
		s += instr.String()
	case *Alloc:
		// fmt.Printf("found an alloc %s\n", v)
		// if v.Scope != nil {
		// 	println("got alloc scope %d", v.Scope.scopeNum)
		// }
		s += instr.String()
	case nil:
		// Be robust against bad transforms.
		s += "<deleted>"
	default:
		s += instr.String()
	}
	return s
}
