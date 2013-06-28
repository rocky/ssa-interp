package interp

import (
	"fmt"
	"os"

	"gnureadline"

	"go/token"
	"ssa-interp"
)

var Event2Icon map[ssa2.TraceEvent]string

var term string
var cmdCount int

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
	term = os.ExpandEnv("TERM")
	gnureadline.StifleHistory(30)
	cmdCount = 0;

}

func StackLocation(fr *frame) string {
	fn := fr.Fn
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

func fmtLocation(start token.Position, end token.Position) string {
	return fmt.Sprintf("%s", PositionRange(start, end))
}

func printLocInfo(fr *frame, start token.Position, end token.Position,
                  event ssa2.TraceEvent) {
	s := Event2Icon[event] + " "
	if len(fr.Fn.Name()) > 0 {
		s += fr.Fn.Name() + "() "
	}
	fmt.Println(s)
	if (event == ssa2.CALL_RETURN) {
		fmt.Printf("return: %s\n", toString(fr.result))
	} else if (event == ssa2.CALL_ENTER) {
		for i, p := range fr.Fn.Params {
			fmt.Println(fr.Fn.Params[i], fr.Env[p])
		}
	}

	fmt.Println(fmtLocation(start, end))
}
