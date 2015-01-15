// Copyright 2013-2015 Rocky Bernstein.
// Things dealing with locations

package gub

import (
	"fmt"
	"go/ast"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
	"runtime/debug"
)

var Event2Icon map[ssa2.TraceEvent]string

func init() {
	Event2Icon = map[ssa2.TraceEvent]string{
		ssa2.OTHER           : "???",
		ssa2.ASSIGN_STMT     : ":= ",
		ssa2.BLOCK_END       : "}  ",
		ssa2.BREAK_STMT      : "<-X",
		ssa2.BREAKPOINT      : "xxx",
		ssa2.CALL_ENTER      : "-> ",
		ssa2.CALL_RETURN     : "<- ",
		ssa2.DEFER_ENTER     : "d->",
		ssa2.TRACE_CALL      : ":o)",  // bozo the clown
		ssa2.EXPR            : "(.)",
		ssa2.IF_INIT         : "if:",
		ssa2.IF_COND         : "if?",
		ssa2.STEP_INSTRUCTION: "...",
		ssa2.FOR_INIT        : "lo:",
		ssa2.FOR_COND        : "lo?",
		ssa2.FOR_ITER        : "lo+",
		ssa2.MAIN            : "m()",
		ssa2.PANIC           : "oX ",  // My attempt at skull and cross bones
		ssa2.RANGE_STMT      : "...",
		ssa2.SELECT_TYPE     : "sel",
		ssa2.SWITCH_COND     : "sw?",
		ssa2.STMT_IN_LIST    : "---",
		ssa2.PROGRAM_TERMINATION : "FIN",
	}
}

func printLocInfo(fr *interp.Frame, inst *ssa2.Instruction,
	event ssa2.TraceEvent) {
	defer func() {
		if x := recover(); x != nil {
			Errmsg("Internal error in getting location info")
			debug.PrintStack()
		}
	}()
	s    := Event2Icon[event] + " "
	fn   := fr.Fn()
	sig  := fn.Signature
	name := fn.Name()

	if fn.Signature.Recv() != nil {
		if len(fn.Params) == 0 {
			panic("Receiver method "+name+" should have at least 1 param. Has 0.")
		}
		s += fmt.Sprintf("(%s).%s()", fn.Params[0].Type(), name)
	} else {
		s += fmt.Sprintf("%s.%s", fn.Pkg.Object.Path(), name)
		if len(name) > 0 { s += "()" }
	}

	if *terse && (event != ssa2.STEP_INSTRUCTION) {
		Msg(s)
	} else {
		Msg("%s block %d insn %d", s, fr.Block().Index, fr.PC())
	}

	var syntax ast.Node = nil

	switch event {
	case ssa2.CALL_RETURN:
		if sig.Results() == nil {
			Msg("return void")
		} else {
			Msg("return type: %s", sig.Results())
			Msg("return value: %s", Deref2Str(fr.Result(), nil))
		}
	case ssa2.CALL_ENTER:
		syntax = fn.Syntax()
		for _, p := range fn.Params {
			if val := fr.Env()[p]; val != nil {
				ssaVal := ssa2.Value(p)
				Msg("%s %s", p, Deref2Str(val, &ssaVal))
			} else {
				Msg("%s nil", p)
			}
		}
	case ssa2.PANIC:
		// fmt.Printf("panic arg: %s\n", fr.Get(instr.X))
	}

	Msg(fr.PositionRange())
	switch s := (*Instr).(type) {
	case *ssa2.Trace:
		syntax = s.Syntax()
	}
    if syntax != nil {
		PrintSyntaxFirstLine(syntax, fn.Prog.Fset)
	}
}
