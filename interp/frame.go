// Copyright 2014 Rocky Bernstein.
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"runtime/debug"
	"fmt"
	"os"
	"go/token"
	"github.com/rocky/ssa-interp"
)

type deferred struct {
	fn    Value
	args  []Value
	instr *ssa2.Defer
	tail  *deferred
}

type Frame struct {
	i                *interpreter
	caller           *Frame
	fn               *ssa2.Function
	block, prevBlock *ssa2.BasicBlock
	env              map[ssa2.Value]Value // dynamic Values of SSA variables
	locals           []Value
	defers           []func()
	result           Value
	panicking        bool
	panic            interface{}

	status           RunStatusType
	tracing		     TraceType
	goNum            int         // Goroutine number
	Var2Reg          map[string] string // Turns an SSA
										// register/variable into its
										// local name
	Reg2Var         map[string] string  // Turns an SSA
										// register/variable into its
										// local name

	// For tracking where we are
	pc               int         // Instruction index of basic block
	startP           token.Pos   // Start Position from last trace instr run
	endP             token.Pos   // End Postion from last trace instr run
}

/* FIXME ROCKY: use Slice instead.
 */
type PC struct{
	fn *ssa2.Function
	block *ssa2.BasicBlock
	instruction int
}

var PCMapping map[uintptr] *PC

func init() {
	PCMapping = make(map[uintptr]*PC)
	/*
      Index 0 needs to be handled as a special case.
      It is  the current PC, not something on a call stack.
     Therefore it has to be handled dynamically.
    */
	PCMapping[0] = nil
}

func (fr *Frame) get(key ssa2.Value) Value {
	switch key := key.(type) {
	case nil:
		// Hack; simplifies handling of optional attributes
		// such as ssa2.Slice.{Low,High}.
		return nil
	case *ssa2.Function, *ssa2.Builtin:
		return key
	case *ssa2.Const:
		return constValue(key)
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

func (fr *Frame) FnAndParamString() string {
	return fr.Fn().FnAndParamString()
}

func (fr *Frame) Scope() *ssa2.Scope {
	if fr.block == nil {
		println("Whoa there, block is nil!")
		debug.PrintStack()
		return nil
	}
	return fr.block.Scope
}

// runDefers executes fr's deferred function calls in LIFO order.
//
// On entry, fr.panicking indicates a state of panic; if
// true, fr.panic contains the panic value.
//
// On completion, if a deferred call started a panic, or if no
// deferred call recovered from a previous state of panic, then
// runDefers itself panics after the last deferred call has run.
//
// If there was no initial state of panic, or it was recovered from,
// runDefers returns normally.
//
func (fr *Frame) runDefers() {
	for i := range fr.defers {
		if (fr.i.TraceMode & EnableTracing) != 0 {
			fmt.Fprintln(os.Stderr, "Invoking deferred function", i)
		}
		fn := fr.defers[len(fr.defers)-1-i]
		TraceHook(fr, nil, ssa2.DEFER_ENTER)
		fn()
	}
	fr.defers = nil
	if fr.panicking {
		panic(fr.panic) // new panic, or still panicking
	}
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
func (fr *Frame) I() *interpreter { return fr.i }
func (fr *Frame) Local(i uint) Value { return fr.locals[i] }
func (fr *Frame) Locals() []Value { return fr.locals }
func (fr *Frame) PC() int { return fr.pc }
func (fr *Frame) PrevBlock() *ssa2.BasicBlock { return fr.prevBlock }
func (fr *Frame) Result() Value { return fr.result }
func (fr *Frame) SetPC(newpc int) { fr.pc = newpc }
func (fr *Frame) StartP() token.Pos { return fr.startP }
func (fr *Frame) Status() RunStatusType { return fr.status }
