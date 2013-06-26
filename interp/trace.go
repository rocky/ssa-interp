package interp

import (
	"fmt"
	"go/token"
	"ssa-interp"
)

// FIXME: arrange to put in ast
func PositionRange(start token.Position, end token.Position) string {
	s := ""
	if start.IsValid() {
		s = start.Filename
		if s != "" {
			s += ":"
		}
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

type TraceHookFunc func(*frame, *ssa2.Instruction, token.Pos, token.Pos, ssa2.TraceEvent)
// FIXME: turn into a map of TraceHookFuncs
var TraceHook TraceHookFunc

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func DefaultTraceHook(fr *frame, instr *ssa2.Instruction, start token.Pos, end token.Pos,
	event ssa2.TraceEvent) {
	fset := fr.Fn.Prog.Fset
	startP := fset.Position(start)
	endP   := fset.Position(end)
	s := fmt.Sprintf("Event: %s ", ssa2.Event2Name[event])
	if len(fr.Fn.Name()) > 0 {
		s += fr.Fn.Name() + "() "
	}
	fmt.Printf("%sat\n%s\n", s, PositionRange(startP, endP))
}

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func NullTraceHook(fr *frame, instr *ssa2.Instruction, start token.Pos, end token.Pos,
	event ssa2.TraceEvent) {
	return
}

// FIXME: should be able to chain trace hooks
func SetTraceHook(hook TraceHookFunc) {
	// FIXME turn this into an append
	TraceHook = hook
}
