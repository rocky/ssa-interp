package ssa2

import (
	"fmt"
	"go/token"
)

//-------------------------------
type TraceEvent int
const (
	OTHER TraceEvent = iota
	ASSIGN_STMT
	BLOCK_END
	BREAK_STMT
	BREAKPOINT
	CALL_ENTER
	CALL_RETURN
	EXPR
	IF_INIT
	IF_COND
	FOR_INIT
	FOR_COND
	FOR_ITER
	RANGE_STMT
	MAIN
	SELECT_TYPE
	STMT_IN_LIST
)


var Event2Name map[TraceEvent]string

func init() {
	Event2Name = map[TraceEvent]string{
		OTHER:    "?",
		ASSIGN_STMT: "Assignment Statement",
		BLOCK_END: "Block End",
		BREAK_STMT: "BREAK",
		BREAKPOINT: "Breakpoint",
		CALL_ENTER: "function entry",
		CALL_RETURN: "function return",
		EXPR:     "Expression",
		IF_INIT: "IF initialize",
		IF_COND: "IF expression",
		FOR_INIT: "FOR initialize",
		FOR_COND: "FOR condition",
		FOR_ITER: "FOR iteration",
		MAIN:     "before main()",
		RANGE_STMT: "range statement",
		SELECT_TYPE: "SELECT type",
		STMT_IN_LIST: "STATEMENT in list",
	}
}

// The Trace instruction is a placeholder some event that is
// about to take place. The event could be
// - a new statement
// - an interesting expression in a "case" or "if" or "loop" statement
// - a return that is about to occur
// - a message synchronization
//
// These are intented to be used by a debugger, profiler, code coverage
// tool or tracing tool.
//
// I'd like this to be a flag an instruction, but that
// was too difficult or ugly to be able for the high-level
// builder call to be able to access the first generated instruction.
// So instead we make it it's own instruction.
type Trace struct {
	anInstruction
	Start   token.Pos  // start position of source
	End   token.Pos    // end position of source
	Event TraceEvent
	Breakpoint bool    // Set if we should stop here
}

func (t *Trace) String() string {
	return fmt.Sprintf("trace <%s>", Event2Name[t.Event])
}

func (v *Trace) Operands(rands []*Value) []*Value {
	return rands
}

// Accessors
func (v *Trace) Pos() token.Pos     { return v.Start }
