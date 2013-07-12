// Copyright 2013 Rocky Bernstein.
// disassemble command

package gub

import "os"

func init() {
	name := "disassemble"
	cmds[name] = &CmdInfo{
		fn: DisassembleCommand,
		help: `disasm [*fn*]

disassemble function`,
		min_args: 0,
		max_args: 1,
	}
	AddToCategory("inspecting", name)
	AddAlias("disasm", name)
}

func DisassembleCommand(args []string) {
	myfn := curFrame.Fn()
	if len(args) > 1 {
		name := args[1]
		pkg  := myfn.Pkg
		if fn := pkg.Func(name); fn != nil {
			myfn = fn
		} else {
			errmsg("Can't find function %s", name)
			return
		}
	}
	myfn.DumpTo(os.Stderr)
}
