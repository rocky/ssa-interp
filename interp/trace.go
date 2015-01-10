package interp

import (
	"fmt"
	"github.com/rocky/ssa-interp"
	"sync"
)

// TraceType is a bitmask of options influencing the tracing per frame
type TraceType int

const (
	TRACE_STEP_NONE = iota
	TRACE_STEP_IN
	TRACE_STEP_INSTRUCTION
	TRACE_STEP_OUT
	TRACE_STEP_OVER
)

var gocall sync.Mutex

type GoreState struct {
	Fr     *Frame
	state  int  // running, finished, etc. Fill this in later
}

// TraceMode is a bitmask of options influencing the tracing.
type TraceMode uint

const (
	// Print a trace of all instructions as they are interpreted.
	EnableTracing  TraceMode = 1 << iota

	// Print higher-level statement boundary tracing
	EnableStmtTracing

	// Trace init functions before main.main()
	EnableInitTracing
)

type RunStatusType int

const (
	StRunning RunStatusType = iota
	StComplete
	StPanic
	StGoPanic
)

type TraceHookFunc func(*Frame, *ssa2.Instruction, ssa2.TraceEvent)
// FIXME: turn into a map of TraceHookFuncs
var TraceHook TraceHookFunc

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func DefaultTraceHook(fr *Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	if ! fr.i.TraceEventMask[event] { return }
	fset := fr.Fn().Prog.Fset
	startP := fset.Position(fr.StartP())
	endP   := fset.Position(fr.EndP())
	s := fmt.Sprintf("Event: %s ", ssa2.Event2Name[event])
	if len(fr.Fn().Name()) > 0 {
		s += fr.Fn().Name() + "() "
	}
	fmt.Printf("%sat\n%s\n", s, ssa2.PositionRange(startP, endP))
}

// This gets called for special trace events if tracing is on
// FIXME: Move elsewhere
func NullTraceHook(fr *Frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	return
}

// FIXME: should be able to chain trace hooks
func SetTraceHook(hook TraceHookFunc) {
	// FIXME turn this into an append
	TraceHook = hook
}

func SetStepIn(fr *Frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_IN
}

func SetStepInstruction(fr *Frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_INSTRUCTION
}

func SetStepOver(fr *Frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_OVER
}

func SetStepOut(fr *Frame) {
	i.TraceMode |= EnableStmtTracing
	fr.tracing = TRACE_STEP_OUT
}

func SetStepOff(fr *Frame) {
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

func SetFnBreakpoint(fn *ssa2.Function) {
	fn.Breakpoint = true
}

func ClearFnBreakpoint(fn *ssa2.Function) {
	fn.Breakpoint = false
}

func IsFnBreakpoint(fn *ssa2.Function) bool {
	return fn.Breakpoint
}

func Tracing(fr *Frame) TraceType {
	return fr.tracing
}
