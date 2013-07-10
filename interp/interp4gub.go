/*

This file contains definitions beyond interp.go need for the gub
debugger. This could be merged into interp.go but we keep it separate so
as to make diff'ing our ssa.go and the unmodified ssa.go look more
alike.

*/
package interp
import "github.com/rocky/ssa-interp"

func (i  *interpreter) Global(name string, pkg *ssa2.Package)  (v *Value, ok bool) {
	v, ok = i.globals[pkg.Var(name)]
	return
}

// interpreter accessors
func (i *interpreter) Program() *ssa2.Program { return i.prog }
func (i  *interpreter) Globals() map[ssa2.Value]*Value { return i.globals }
func (i  *interpreter) GoTops() []*GoreState { return i.goTops }
