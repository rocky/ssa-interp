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
		syntax: nil,
	}
	return emitTraceCommon(f, t)
}

func emitTraceStmt(f *Function, event TraceEvent, syntax ast.Stmt) Value {
	t := &Trace{
		Event: event,
		Start: syntax.Pos(),
		End:  syntax.End(),
		Breakpoint: false,
		syntax: syntax,
	}
	return emitTraceCommon(f, t)
}

func emitTraceExpr(f *Function, event TraceEvent, syntax ast.Expr) Value {
	t := &Trace{
		Event: event,
		Start: syntax.Pos(),
		End:  syntax.End(),
		Breakpoint: false,
		syntax: syntax,
	}
	return emitTraceCommon(f, t)
}

func emitTraceCommon(f *Function, t *Trace) Value {
	fset := f.Prog.Fset
	pkg := f.Pkg
	pkg.locs = append(pkg.locs,
		LocInst{
			pos: t.Start,
			endP: t.End,
			Fn: nil,
			Trace: t,
		})
	if (debugMe) {
		fmt.Printf("Emitting event %s\n\tFrom: %s\n\tTo: %s\n",
			Event2Name[t.Event], fset.Position(t.Start), fset.Position(t.End)	)
	}
	return f.emit(t)
}
