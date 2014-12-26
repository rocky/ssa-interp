// Copyright 2014 Rocky Bernstein
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

func (fn *Function) FnAndParamString() string {
	i := 0
	s := ""
	if fn.Signature.Recv() == nil {
		s = fmt.Sprintf("%s.%s(", fn.Pkg.Object.Path(), fn.RelString(fn.Pkg.Object))
	} else  {
		if len(fn.Params) == 0 {
			panic("Receiver method "+s+" should have at least 1 param. Has 0.")
		}
		s = fmt.Sprintf("(%s).%s(", fn.Params[0].Type(), fn.Name())
		i++
	}
	params := ""
	if len(fn.Params) > i {
		params = fn.Params[i].Name()
		for i+=1; i<len(fn.Params); i++ {
			params += ", " + fn.Params[i].Name()
		}
	}
	s += params + ")"
	return s
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
