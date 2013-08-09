package ssa2

import (
	"fmt"
	"go/token"
)

//-------------------------------
type TraceEvent uint8
const (
	OTHER TraceEvent = iota
	ASSIGN_STMT
	BLOCK_END
	BREAK_STMT
	BREAKPOINT
	CALL_ENTER
	CALL_RETURN
	DEFER_ENTER
	EXPR
	IF_INIT
	IF_COND
	FOR_INIT
	FOR_COND
	FOR_ITER
	PANIC
	RANGE_STMT
	MAIN
	SELECT_TYPE
	STMT_IN_LIST
	SWITCH_COND
)

const TRACE_EVENT_FIRST = OTHER
const TRACE_EVENT_LAST  = SWITCH_COND

type TraceEventMask map[TraceEvent]bool

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
		SWITCH_COND: "SWITCH condition",
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

// FIXME: arrange to put in ast
func PositionRange(start token.Position, end token.Position) string {
	s := ""
	if start.IsValid() {
		s = start.Filename + ":" + PositionRangeSansFile(start, end)
	} else if end.IsValid() {
		s = "-"
		if end.Filename != "" {
			s += end.Filename + ":"
		}
		s += fmt.Sprintf("%d:%d", end.Line, end.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

func PositionRangeSansFile(start token.Position, end token.Position) string {
	s := ""
	if start.IsValid() {
		s += fmt.Sprintf("%d:%d", start.Line, start.Column)
		if start.Filename == end.Filename && end.IsValid() {
			// this is what we expect
			if start.Line == end.Line {
				if start.Column != end.Column {
					s += fmt.Sprintf("-%d", end.Column)
				}
			} else {
				s += fmt.Sprintf("-%d:%d", end.Line, end.Column)
			}
		}

	} else if end.IsValid() {
		s = "-"
		s += fmt.Sprintf("%d:%d", end.Line, end.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

func FmtPos(fset *token.FileSet, start token.Pos) string {
	if start == token.NoPos { return "-" }
	startP := fset.Position(start)
	return PositionRange(startP, startP)
}

func FmtRangeWithFset(fset *token.FileSet, start token.Pos, end token.Pos) string {
	startP := fset.Position(start)
	endP   := fset.Position(end)
	return PositionRange(startP, endP)
}

func FmtRange(fn *Function, start token.Pos, end token.Pos) string {
	fset := fn.Fset()
	return FmtRangeWithFset(fset, start, end)
}

func (t *Trace) String() string {
	fset := t.block.parent.Prog.Fset
	startP := fset.Position(t.Start)
	endP   := fset.Position(t.End)
	return fmt.Sprintf("trace <%s> at %s",
		Event2Name[t.Event], PositionRange(startP, endP))
}

func (v *Trace) Operands(rands []*Value) []*Value {
	return rands
}

// Accessors
func (v *Trace) Pos() token.Pos     { return v.Start }
