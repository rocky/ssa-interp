package interp

import (
	"fmt"
	"go/token"
	"ssa-interp"
)

// Mode is a bitmask of options influencing the interpreter.
type Mode uint

// Mode is a bitmask of options influencing the tracing.
type TraceMode uint

const (
	// Disable recover() in target programs; show interpreter crash instead.
	DisableRecover Mode = 1 << iota
)

const (
	// Print a trace of all instructions as they are interpreted.
	EnableTracing  TraceMode = 1 << iota

	// Print higher-level statement boundary tracing
	EnableStmtTracing
)

type Status int

const (
	StRunning Status = iota
	StComplete
	StPanic
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

type TraceHookFunc func(*frame, *ssa2.Instruction, ssa2.TraceEvent)
// FIXME: turn into a map of TraceHookFuncs
var TraceHook TraceHookFunc

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func DefaultTraceHook(fr *frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	fset := fr.Fn.Prog.Fset
	startP := fset.Position(fr.StartP)
	endP   := fset.Position(fr.EndP)
	s := fmt.Sprintf("Event: %s ", ssa2.Event2Name[event])
	if len(fr.Fn.Name()) > 0 {
		s += fr.Fn.Name() + "() "
	}
	fmt.Printf("%sat\n%s\n", s, PositionRange(startP, endP))
}

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func NullTraceHook(fr *frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	return
}

// FIXME: should be able to chain trace hooks
func SetTraceHook(hook TraceHookFunc) {
	// FIXME turn this into an append
	TraceHook = hook
}

func SetStepIn(fr *frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_IN
}

func SetStepOver(fr *frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_OVER
}

func SetStepOut(fr *frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_OUT
}

func SetStepOff(fr *frame) {
	i.TraceMode &= ^EnableStmtTracing
	fr.tracing = TRACE_STEP_NONE
}

func SetInstTracing() {
	i.TraceMode |= EnableTracing
}

func ClearInstTracing() {
	i.TraceMode &= ^EnableTracing
}

func InstTracing() bool {
	return 0 != i.TraceMode & EnableTracing
}

func GlobalStmtTracing() bool {
	return 0 != i.TraceMode & EnableStmtTracing
}
