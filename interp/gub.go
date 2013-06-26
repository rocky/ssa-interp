package interp

import (
	"fmt"
	"go/token"
	"ssa-interp"
)

var Event2Icon map[ssa2.TraceEvent]string

func init() {
	Event2Icon = map[ssa2.TraceEvent]string{
		ssa2.OTHER      : "???",
		ssa2.ASSIGN_STMT: ":= ",
		ssa2.BLOCK_END  : "}  ",
		ssa2.BREAK_STMT : "<-X",
		ssa2.CALL_ENTER : "-> ",
		ssa2.CALL_RETURN: "<- ",
		ssa2.EXPR       : "(.)",
		ssa2.IF_INIT    : "if:",
		ssa2.IF_COND    : "if?",
		ssa2.FOR_INIT   : "lo:",
		ssa2.FOR_COND   : "lo?",
		ssa2.FOR_ITER   : "lo+",
		ssa2.MAIN       : "m()",
		ssa2.RANGE_STMT : "...",
		ssa2.SELECT_TYPE: "sel",
		ssa2.STMT_IN_LIST: "---",
	}
}

func printLocation(start token.Position, end token.Position) {
	fmt.Printf("%s\n", PositionRange(start, end))
}

func printLocInfo(fr *frame, start token.Position, end token.Position,
                  event ssa2.TraceEvent) {
	s := Event2Icon[event]
	if len(fr.Fn.Name()) > 0 {
		s += fr.Fn.Name() + "() "
	}
	fmt.Println(s)
	printLocation(start, end)
}

func GubTraceHook(fr *frame, instr *ssa2.Instruction, start token.Pos, end token.Pos,
	event ssa2.TraceEvent) {
	fset := fr.Fn.Prog.Fset
	startP := fset.Position(start)
	endP   := fset.Position(end)
	printLocInfo(fr, startP, endP, event)
}
