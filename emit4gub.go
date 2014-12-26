// Copyright 2014 Rocky Bernstein
package ssa2

// emitting SSA trace instructions.

import (
	"fmt"
	"go/ast"
	"go/token"
)

// emitTrace emits to f an instruction to which acts as a
// placeholder for the kind of high-level event that is
// coming up next: a new statement, the return from a function
// and so on. I'd like this to be a flag an instruction, but that
// was too difficult or ugly to be able for the high-level
// builder call to be able to access the first generated instruction.
// So instead we make it it's own instruction.

func emitTrace(f *Function, event TraceEvent, start token.Pos,
               end token.Pos) Value {
	t := &Trace{
		Event: event,
		Start: start,
		End: end,
		Breakpoint: false,
		ast: nil,
	}
	// fmt.Printf("event %s StartPos %d EndPos %d\n", Event2Name[event])
	fset := f.Prog.Fset
	pkg := f.Pkg
	pkg.locs = append(pkg.locs, LocInst{pos: start, endP:end,
		Fn: nil, Trace: t})
	if (debugMe) {
		fmt.Printf("Emitting event %s\n\tFrom: %s\n\tTo: %s\n",
			Event2Name[event], fset.Position(start), fset.Position(end)	)
	}
	return f.emit(t)
}

func emitTraceNode(f *Function, event TraceEvent, syntax *ast.Node) Value {
	start := (*syntax).Pos()
	end := (*syntax).End()
	t := &Trace{
		Event: event,
		Start: start,
		End:  end,
		Breakpoint: false,
		ast: syntax,
	}
	// fmt.Printf("event %s StartPos %d EndPos %d\n", Event2Name[event])
	fset := f.Prog.Fset
	pkg := f.Pkg
	pkg.locs = append(pkg.locs, LocInst{pos: start, endP:end,
		Fn: nil, Trace: t})
	if (debugMe) {
		fmt.Printf("Emitting event %s\n\tFrom: %s\n\tTo: %s\n",
			Event2Name[event], fset.Position(start), fset.Position(end)	)
	}
	return f.emit(t)
}
