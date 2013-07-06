package interp

import (
	"fmt"
	"os"
	"go/token"
	"github.com/rocky/ssa-interp"
)

type Frame struct {
	i                *Interpreter
	caller           *Frame
	fn               *ssa2.Function
	block, prevBlock *ssa2.BasicBlock
	env              map[ssa2.Value]Value // dynamic Values of SSA variables
	locals           []Value
	defers           []func()
	result           Value
	status           RunStatusType
	tracing			 traceType
	panic            interface{}

	goNum            int         // Goroutine number

	// For tracking where we are
	pc               int         // Instruction index of basic block
	startP           token.Pos   // Start Position from last trace instr run
	endP             token.Pos   // End Postion from last trace instr run
}

func (fr *Frame) Get(key ssa2.Value) Value {
	switch key := key.(type) {
	case nil:
		// Hack; simplifies handling of optional attributes
		// such as ssa2.Slice.{Low,High}.
		return nil
	case *ssa2.Function, *ssa2.Builtin:
		return key
	case *ssa2.Literal:
		return LiteralValue(key)
	case *ssa2.Global:
		if r, ok := fr.i.globals[key]; ok {
			return r
		}
	}
	if r, ok := fr.env[key]; ok {
		return r
	}
	panic(fmt.Sprintf("get: no value for %T: %v", key, key.Name()))
}

func (fr *Frame) rundefers() {
	for i := range fr.defers {
		if (fr.i.TraceMode & EnableTracing) != 0 {
			fmt.Fprintln(os.Stderr, "Invoking deferred function", i)
		}
		fr.defers[len(fr.defers)-1-i]()
	}
	fr.defers = fr.defers[:0]
}

func (fr *Frame) Fset() *token.FileSet { return fr.fn.Prog.Fset }

func (fr *Frame) Position() token.Position {
	fset   := fr.fn.Prog.Fset
	return fset.Position(fr.startP)
}

func (fr *Frame) PositionRange() string {
	fset   := fr.fn.Prog.Fset
	startP := fset.Position(fr.startP)
	endP   := fset.Position(fr.endP)
	return ssa2.PositionRange(startP, endP)
}

func (fr *Frame) Caller(skip int) *Frame {
	targetFrame := fr
	for i:=0; i<=skip; i++ {
		if targetFrame.caller != nil {
			targetFrame  = targetFrame.caller
		} else {
			return nil
		}
	}
	return targetFrame
}
// Frame accessors
func (fr *Frame) Block() *ssa2.BasicBlock { return fr.block }
func (fr *Frame) EndP()   token.Pos { return fr.endP }
func (fr *Frame) Env() map[ssa2.Value]Value { return fr.env }
func (fr *Frame) Fn() *ssa2.Function { return fr.fn }
func (fr *Frame) GoNum() int { return fr.goNum }
func (fr *Frame) I() *Interpreter { return fr.i }
func (fr *Frame) Locals() []Value { return fr.locals }
func (fr *Frame) PC() int { return fr.pc }
func (fr *Frame) PrevBlock() *ssa2.BasicBlock { return fr.prevBlock }
func (fr *Frame) Result() Value { return fr.result }
func (fr *Frame) StartP() token.Pos { return fr.startP }
func (fr *Frame) Status() RunStatusType { return fr.status }
