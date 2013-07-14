package interp

import (
	"go/token"
	"github.com/rocky/ssa-interp"
)


// interpreter accessors
func Binop(op token.Token, x, y Value) Value {
	return binop(op, x, y)
}
func Unop(instr *ssa2.UnOp, x Value) Value {
	return unop(instr, x)
}
