package interp

import (
	"go/token"
	"golang.org/x/tools/go/types"
	"github.com/rocky/ssa-interp"
)


// interpreter accessors
func Binop(op token.Token, t types.Type, x, y Value) Value {
	return binop(op, t, x, y)
}
func Unop(instr *ssa2.UnOp, x Value) Value {
	return unop(instr, x)
}
