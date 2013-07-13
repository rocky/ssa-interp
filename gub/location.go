// Copyright 2013 Rocky Bernstein.
// Things dealing with locations

package gub

import (
	"fmt"

	"go/token"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

var Event2Icon map[ssa2.TraceEvent]string

func init() {
	Event2Icon = map[ssa2.TraceEvent]string{
		ssa2.OTHER      : "???",
		ssa2.ASSIGN_STMT: ":= ",
		ssa2.BLOCK_END  : "}  ",
		ssa2.BREAK_STMT : "<-X",
		ssa2.BREAKPOINT : "xxx",
		ssa2.CALL_ENTER : "-> ",
		ssa2.CALL_RETURN: "<- ",
		ssa2.EXPR       : "(.)",
		ssa2.IF_INIT    : "if:",
		ssa2.IF_COND    : "if?",
		ssa2.FOR_INIT   : "lo:",
		ssa2.FOR_COND   : "lo?",
		ssa2.FOR_ITER   : "lo+",
		ssa2.MAIN       : "m()",
		ssa2.PANIC      : "oX ",  // My attempt at skull and cross bones
		ssa2.RANGE_STMT : "...",
		ssa2.SELECT_TYPE: "sel",
		ssa2.STMT_IN_LIST: "---",
	}
}

func StackLocation(fr *interp.Frame) string {
	fn := fr.Fn()
	s := fmt.Sprintf("%s(", fn.Name())
	params :=""
	if len(fn.Params) > 0 {
		params = fn.Params[0].Name()
		for i:=1; i<len(fn.Params); i++ {
			params += ", " + fn.Params[i].Name()
		}
	}
	s += params + ")"
	return s
}

func fmtRange(fn *ssa2.Function, start token.Pos, end token.Pos) string {
	fset := fn.Fset()
	startP := fset.Position(start)
	endP   := fset.Position(end)
	return ssa2.PositionRange(startP, endP)
}

func fmtPos(fn *ssa2.Function, start token.Pos) string {
	if start == token.NoPos { return "-" }
	fset := fn.Fset()
	startP := fset.Position(start)
	return ssa2.PositionRange(startP, startP)
}

func printLocInfo(fr *interp.Frame, inst *ssa2.Instruction,
	event ssa2.TraceEvent) {
	if event == ssa2.BREAKPOINT && Breakpoints[curBpnum].kind == "Function" {
		event = ssa2.CALL_ENTER
	}
	s := Event2Icon[event] + " "
	if len(fr.Fn().Name()) > 0 {
		s += fr.Fn().Name() + "()"
	}
	if *terse {
		msg(s)
	} else {
		msg("%s block %d insn %d", s, fr.Block().Index, fr.PC())
	}
	switch event {
	case ssa2.CALL_RETURN:
		fn := fr.Fn()
		if fn.Signature.Results() == nil {
			msg("return void")
		} else {
			msg("return type: %s", fn.Signature.Results())
			msg("return value: %s", interp.ToString(fr.Result()))
		}
	case ssa2.CALL_ENTER:
		for i, p := range fr.Fn().Params {
			s := interp.ToString(fr.Env()[p])
			if p.Type().String() == "string" {
				msg("%s \"%s\"", fr.Fn().Params[i], s)
			} else {
				msg("%s %s", fr.Fn().Params[i], s)
			}
		}
	case ssa2.PANIC:
		// fmt.Printf("panic arg: %s\n", fr.Get(instr.X))
	}

	msg(fr.PositionRange())
}
