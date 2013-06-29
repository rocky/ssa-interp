package interp

import (
	"fmt"
	"os"
	"go/token"
	"ssa-interp"
)

type frame struct {
	i                *Interpreter
	caller           *frame
	fn               *ssa2.Function
	block, prevBlock *ssa2.BasicBlock
	env              map[ssa2.Value]value // dynamic values of SSA variables
	locals           []value
	defers           []func()
	result           value
	status           RunStatusType
	tracing			 traceType
	panic            interface{}

	// For tracking where we are
	pc               int         // Instruction index of basic block
	startP           token.Pos   // Start Position from last trace instr run
	endP             token.Pos   // End Postion from last trace instr run
}

func (fr *frame) get(key ssa2.Value) value {
	switch key := key.(type) {
	case nil:
		// Hack; simplifies handling of optional attributes
		// such as ssa2.Slice.{Low,High}.
		return nil
	case *ssa2.Function, *ssa2.Builtin:
		return key
	case *ssa2.Literal:
		return literalValue(key)
	case *ssa2.Global:
		if r, ok := fr.i.Globals[key]; ok {
			return r
		}
	}
	if r, ok := fr.env[key]; ok {
		return r
	}
	panic(fmt.Sprintf("get: no value for %T: %v", key, key.Name()))
}

func (fr *frame) rundefers() {
	for i := range fr.defers {
		if (fr.i.TraceMode & EnableTracing) != 0 {
			fmt.Fprintln(os.Stderr, "Invoking deferred function", i)
		}
		fr.defers[len(fr.defers)-1-i]()
	}
	fr.defers = fr.defers[:0]
}

// Various Frame accessors
func (fr *frame) PC() int { return fr.pc }
func (fr *frame) Fn() *ssa2.Function { return fr.fn }
func (fr *frame) Block() *ssa2.BasicBlock { return fr.block }
func (fr *frame) PrevBlock() *ssa2.BasicBlock { return fr.prevBlock }
func (fr *frame) Locals() []value { return fr.locals }
func (fr *frame) StartP() token.Pos { return fr.startP }
func (fr *frame) EndP()   token.Pos { return fr.endP }
func (fr *frame) Status() RunStatusType { return fr.status }
func (fr *frame) Env() map[ssa2.Value]value { return fr.env }

func (fr *frame) Caller(skip int) *frame {
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
