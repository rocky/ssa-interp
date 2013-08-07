// Copyright 2013 Rocky Bernstein.

package gub

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func init() {
	name := "instruct"
	Cmds[name] = &CmdInfo{
		Fn: InstructCommand,
		Help: `instruct [num [operand]]

Print information about instruction
`,
		Min_args: 0,
		Max_args: 2,
	}
	AddToCategory("inspecting", name)
	// Down the line we'll have abbrevs
	Aliases["inst"] = name
	Aliases["instr"] = name
	Aliases["instruct"] = name
}

func derefValue(v interp.Value) interp.Value {
	switch v := v.(type) {
	case *interp.Value:
		if v == nil { return nil }
		return *v
	default:
		return v
	}
}

func Deref2Str(v interp.Value) string {
	return interp.ToInspect(derefValue(v))
}


func InstructCommand(args []string) {
	fr := curFrame
	ic := fr.PC()
	if len(args) >= 2 {
		new_ic, ok := GetInt(args[1], "instruction number", 0,
			len(curFrame.Block().Instrs))
		if ok == nil {
			ic = new_ic
		} else {
			Errmsg("Expecting integer; got %s.", args[1])
			return
		}
		// if len(args) == 3 {
		// 	new_num, ok = strconv.Atoi(args[2])
		// 	if ok != nil {
		// 		Errmsg("Expecting integer; got %s", args[2])
		// 		return
		// 	}
	}
	DisasmInst(fr.Fn(), fr.Block().Index, ic)
	genericInstr := fr.Block().Instrs[ic]
	switch instr := genericInstr.(type) {
	case *ssa2.ChangeType:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case *ssa2.Convert:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case  *ssa2.MakeInterface:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case  *ssa2.ChangeInterface:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case  *ssa2.Range:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case *ssa2.UnOp:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case *ssa2.Field:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
	case *ssa2.BinOp:
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.X)))
		Msg("%s: %s", instr.X.Name(), Deref2Str(fr.Get(instr.Y)))
	case *ssa2.Trace:
	default:
		Msg("Don't know how to deal with %s yet", instr)
	}
}
