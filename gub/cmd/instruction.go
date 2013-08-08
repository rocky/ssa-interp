// Copyright 2013 Rocky Bernstein.

package gubcmd

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	name := "instruction"
	gub.Cmds[name] = &gub.CmdInfo{
		Fn: InstructionCommand,
		Help: `instruction [num [operand]]

Print information about instruction
`,
		Min_args: 0,
		Max_args: 2,
	}
	gub.AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	gub.Aliases["inst"] = name
	gub.Aliases["instr"] = name
	gub.Aliases["instruct"] = name
}

func InstructionCommand(args []string) {
	fr := gub.CurFrame()
	ic := uint64(fr.PC())
	if len(args) >= 2 {
		new_ic, ok := gub.GetUInt(args[1], "instruction number", 0,
			uint64(len(gub.CurFrame().Block().Instrs)))
		if ok == nil {
			ic = new_ic
		} else {
			gub.Errmsg("Expecting integer; got %s.", args[1])
			return
		}
		// if len(args) == 3 {
		// 	new_num, ok = strconv.Atoi(args[2])
		// 	if ok != nil {
		// 		gub.Errmsg("Expecting integer; got %s", args[2])
		// 		return
		// 	}
	}
	gub.DisasmInst(fr.Fn(), fr.Block().Index, ic)
	genericInstr := fr.Block().Instrs[ic]
	switch instr := genericInstr.(type) {
	case *ssa2.ChangeType:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case *ssa2.Convert:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case  *ssa2.MakeInterface:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case  *ssa2.ChangeInterface:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case  *ssa2.Range:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case *ssa2.UnOp:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case *ssa2.Field:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
	case *ssa2.BinOp:
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.X)))
		gub.Msg("%s: %s", instr.X.Name(), gub.Deref2Str(fr.Get(instr.Y)))
	case *ssa2.Trace:
	default:
		gub.Msg("Don't know how to deal with %s yet", instr)
	}
}
