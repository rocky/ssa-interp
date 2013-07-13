// Copyright 2013 Rocky Bernstein.
// Things involving continuing execution
package gub

import (
	"fmt"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "continue"
	cmds[name] = &CmdInfo{
		fn: ContinueCommand,
		help: `continue

Leave the debugger loop and continue execution. Subsequent entry to
the debugger however may occur via breakpoints or explicit calls, or
exceptions.
`,
		min_args: 0,
		max_args: 0,
	}
	AddAlias("c", name)
	AddToCategory("running", name)
}

func ContinueCommand(args []string) {
	for fr := topFrame; fr != nil; fr = fr.Caller(0) {
		interp.SetStepOff(fr)
	}
	inCmdLoop = false
	msg("Continuing...")
}

func init() {
	name := "finish"
	cmds[name] = &CmdInfo{
		fn: FinishCommand,
		help: `finish

Continue execution until the program is about to:

* leave the current function, or
* switch context via yielding back or finishing a block which was
  yielded to.

Sometimes this is called 'step out'.
`,
		min_args: 0,
		max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("fin", name)
}

func FinishCommand(args []string) {
	interp.SetStepOut(topFrame)
	msg("Continuing until return...")
	inCmdLoop = false
}

func init() {
	name := "next"
	cmds[name] = &CmdInfo{
		fn: NextCommand,
		help: `next

Step one statement ignoring steps into function calls at this level.

Sometimes this is called 'step over'.
`,
		min_args: 0,
		max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("n", name)
}

func NextCommand(args []string) {
	interp.SetStepOver(topFrame)
	fmt.Println("Step over...")
	inCmdLoop = false
}

func init() {
	name := "step"
	cmds[name] = &CmdInfo{
		fn: StepCommand,
		help: `step

Execute the current line, stopping at the next event.  Sometimes this
is called 'step into'.
`,
		min_args: 0,
		max_args: 0,
	}
	AddToCategory("running", name)
	// Down the line we'll have abbrevs
	AddAlias("s", name)
}

func StepCommand(args []string) {
	fmt.Println("Stepping...")
	interp.SetStepIn(curFrame)
	inCmdLoop = false
}
