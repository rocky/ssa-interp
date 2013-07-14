// Copyright 2013 Rocky Bernstein.

package gub

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "instruct"
	cmds[name] = &CmdInfo{
		fn: InstructCommand,
		help: `instruct [num [operand]]

Print information about instruction
`,
		min_args: 0,
		max_args: 2,
	}
	AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	aliases["inst"] = name
	aliases["instr"] = name
	aliases["instruct"] = name
}

func derefValue(v interp.Value) string {
	switch v := v.(type) {
	case *interp.Value:
		return interp.ToInspect(*v)
	default:
		return interp.ToInspect(v)
	}
}

func InstructCommand(args []string) {
	fr := curFrame
	ic := fr.PC()
	if len(args) >= 2 {
		new_ic, ok := getInt(args[1], "instruction number", 0,
			len(curFrame.Block().Instrs))
		if ok == nil {
			ic = new_ic
		} else {
			errmsg("Expecting integer; got %s.", args[1])
			return
		}
		// if len(args) == 3 {
		// 	new_num, ok = strconv.Atoi(args[2])
		// 	if ok != nil {
		// 		errmsg("Expecting integer; got %s", args[2])
		// 		return
		// 	}
	}
	DisasmInst(fr.Fn(), fr.Block().Index, ic)
	genericInstr := fr.Block().Instrs[ic]
	switch instr := genericInstr.(type) {
	case *ssa2.ChangeType:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case *ssa2.Convert:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case  *ssa2.MakeInterface:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case  *ssa2.ChangeInterface:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case  *ssa2.Range:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case *ssa2.UnOp:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case *ssa2.Field:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
	case *ssa2.BinOp:
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.X)))
		msg("%s: %s", instr.X.Name(), derefValue(fr.Get(instr.Y)))
	case *ssa2.Trace:
	default:
		msg("Don't know how to deal with %s yet", instr)
	}
}
