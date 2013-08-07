// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution
package gub

import (
	"fmt"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "continue"
	Cmds[name] = &CmdInfo{
		Fn: ContinueCommand,
		Help: `continue

Leave the debugger loop and continue execution. Subsequent entry to
the debugger however may occur via breakpoints or explicit calls, or
exceptions.
`,
		Min_args: 0,
		Max_args: 0,
	}
	AddAlias("c", name)
	AddToCategory("running", name)
}

func ContinueCommand(args []string) {
	for fr := topFrame; fr != nil; fr = fr.Caller(0) {
		interp.SetStepOff(fr)
	}
	InCmdLoop = false
	Msg("Continuing...")
}

func init() {
	name := "finish"
	Cmds[name] = &CmdInfo{
		Fn: FinishCommand,
		Help: `finish

Continue execution until the program is about to:

* leave the current function, or
* switch context via yielding back or finishing a block which was
  yielded to.

Sometimes this is called 'step out'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("fin", name)
}

func FinishCommand(args []string) {
	interp.SetStepOut(topFrame)
	Msg("Continuing until return...")
	InCmdLoop = false
}

func init() {
	name := "next"
	Cmds[name] = &CmdInfo{
		Fn: NextCommand,
		Help: `next

Step one statement ignoring steps into function calls at this level.

Sometimes this is called 'step over'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("n", name)
}

func NextCommand(args []string) {
	interp.SetStepOver(topFrame)
	fmt.Println("Step over...")
	InCmdLoop = false
}

func init() {
	name := "step"
	Cmds[name] = &CmdInfo{
		Fn: StepCommand,
		Help: `step

Execute the current line, stopping at the next event.  Sometimes this
is called 'step into'.
`,
		Min_args: 0,
		Max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("s", name)
}

func StepCommand(args []string) {
	fmt.Println("Stepping...")
	interp.SetStepIn(curFrame)
	InCmdLoop = false
}
