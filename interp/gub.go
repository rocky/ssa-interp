package interp

import (
	"fmt"
	"strings"
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
	fmt.Println(fmtLocation(start, end))
}

func GubTraceHook(fr *frame, instr *ssa2.Instruction, event ssa2.TraceEvent) {
	fset := fr.Fn.Prog.Fset
	startP := fset.Position(fr.StartP)
	endP   := fset.Position(fr.EndP)
	printLocInfo(fr, startP, endP, event)
	line := ""
	inCmdLoop := true
	var err error
	for ; err == nil && inCmdLoop; cmdCount++ {
		line, err = gnureadline.Readline(fmt.Sprintf("gub[%d]: ", cmdCount),
			true)
		args  := strings.Split(line, " ")
		if len(args) == 0 {
			fmt.Println("Empty line skipped")
			continue
		}

		cmd := args[0]

		switch cmd {
		case "s":
			fmt.Println("Stepping...")
			inCmdLoop = false
			break
		case "c":
			ClearStmtTracing()
			fmt.Println("Continuing...")
			inCmdLoop = false
			break
		case "+":
			fmt.Println("Setting Instruction Trace")
			SetInstTracing()
		case "-":
			fmt.Println("Clearing Instruction Trace")
			ClearInstTracing()
		case "gl", "globals":
			GlobalsCommand(fr, args)
		case "lo", "locals":
			LocalsCommand(fr, args)
		case "q", "quit":
			QuitCommand(fr, args)
		case "bt", "T", "backtrace":
			BacktraceCommand(fr, args)
		case "v":
			VariableCommand(fr, args)
		default:
			fmt.Printf("Unknown command %s\n", cmd)
		}
	}
}
