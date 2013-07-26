package ssa2

// Helpers for emitting SSA instructions.

import (
	"fmt"
	"go/ast"
	"go/token"

	"code.google.com/p/go.tools/go/types"
)

const (
	debugMe bool = false
)
// emitNew emits to f a new (heap Alloc) instruction allocating an
// object of type typ.  pos is the optional source location.
//
func emitNew(f *Function, typ types.Type, pos token.Pos) Value {
	return f.emit(&Alloc{
		typ:  types.NewPointer(typ),
		Heap: true,
		pos:  pos,
	})
}

// emitLoad emits to f an instruction to load the address addr into a
// new temporary, and returns the value so defined.
//
func emitLoad(f *Function, addr Value) *UnOp {
	v := &UnOp{Op: token.MUL, X: addr}
	v.setType(deref(addr.Type()))
	f.emit(v)
	return v
}

// emitDebugRef emits to f a DebugRef pseudo-instruction associating
// reference id with local var/const value v.
//
func emitDebugRef(f *Function, id *ast.Ident, v Value) {
	if !f.debugInfo() {
		return // debugging not enabled
	}
	if isBlankIdent(id) {
		return
	}
	obj := f.Pkg.objectOf(id)
	if obj.Parent() == types.Universe {
		return // skip nil/true/false
	}
	f.emit(&DebugRef{
		X:      v,
		pos:    id.Pos(),
		object: obj,
	})
}

// emitArith emits to f code to compute the binary operation op(x, y)
// where op is an eager shift, logical or arithmetic operation.
// (Use emitCompare() for comparisons and Builder.logicalBinop() for
// non-eager operations.)
//
func emitArith(f *Function, op token.Token, x, y Value, t types.Type, pos token.Pos) Value {
	switch op {
	case token.SHL, token.SHR:
		x = emitConv(f, x, t)
		// y may be signed or an 'untyped' constant.
		// TODO(adonovan): whence signed values?
		if b, ok := y.Type().Underlying().(*types.Basic); ok && b.Info()&types.IsUnsigned == 0 {
			y = emitConv(f, y, types.Typ[types.Uint64])
		}

	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.AND, token.OR, token.XOR, token.AND_NOT:
		x = emitConv(f, x, t)
		y = emitConv(f, y, t)

	default:
		panic("illegal op in emitArith: " + op.String())

	}
	v := &BinOp{
		Op: op,
		X:  x,
		Y:  y,
	}
	v.setPos(pos)
	v.setType(t)
	return f.emit(v)
}

// emitCompare emits to f code compute the boolean result of
// comparison comparison 'x op y'.
//
func emitCompare(f *Function, op token.Token, x, y Value, pos token.Pos) Value {
	xt := x.Type().Underlying()
	yt := y.Type().Underlying()

	// Special case to optimise a tagless SwitchStmt so that
	// these are equivalent
	//   switch { case e: ...}
	//   switch true { case e: ... }
	//   if e==true { ... }
	// even in the case when e's type is an interface.
	// TODO(adonovan): opt: generalise to x==true, false!=y, etc.
	if x == vTrue && op == token.EQL {
		if yt, ok := yt.(*types.Basic); ok && yt.Info()&types.IsBoolean != 0 {
			return y
		}
	}

	if types.IsIdentical(xt, yt) {
		// no conversion necessary
	} else if _, ok := xt.(*types.Interface); ok {
		y = emitConv(f, y, x.Type())
	} else if _, ok := yt.(*types.Interface); ok {
		x = emitConv(f, x, y.Type())
	} else if _, ok := x.(*Const); ok {
		x = emitConv(f, x, y.Type())
	} else if _, ok := y.(*Const); ok {
		y = emitConv(f, y, x.Type())
	} else {
		// other cases, e.g. channels.  No-op.
	}

	v := &BinOp{
		Op: op,
		X:  x,
		Y:  y,
	}
	v.setPos(pos)
	v.setType(tBool)
	return f.emit(v)
}

// isValuePreserving returns true if a conversion from ut_src to
// ut_dst is value-preserving, i.e. just a change of type.
// Precondition: neither argument is a named type.
//
func isValuePreserving(ut_src, ut_dst types.Type) bool {
	// Identical underlying types?
	if types.IsIdentical(ut_dst, ut_src) {
		return true
	}

	switch ut_dst.(type) {
	case *types.Chan:
		// Conversion between channel types?
		_, ok := ut_src.(*types.Chan)
		return ok

	case *types.Pointer:
		// Conversion between pointers with identical base types?
		_, ok := ut_src.(*types.Pointer)
		return ok

	case *types.Signature:
		// Conversion from (T) func f() method to f(T) function?
		_, ok := ut_src.(*types.Signature)
		return ok
	}
	return false
}

// emitConv emits to f code to convert Value val to exactly type typ,
// and returns the converted value.  Implicit conversions are required
// by language assignability rules in assignments, parameter passing,
// etc.
//
func emitConv(f *Function, val Value, typ types.Type) Value {
	t_src := val.Type()

	// Identical types?  Conversion is a no-op.
	if types.IsIdentical(t_src, typ) {
		return val
	}

	ut_dst := typ.Underlying()
	ut_src := t_src.Underlying()

	// Just a change of type, but not value or representation?
	if isValuePreserving(ut_src, ut_dst) {
		c := &ChangeType{X: val}
		c.setType(typ)
		return f.emit(c)
	}

	// Conversion to, or construction of a value of, an interface type?
	if _, ok := ut_dst.(*types.Interface); ok {

		// Assignment from one interface type to another?
		if _, ok := ut_src.(*types.Interface); ok {
			return emitTypeAssert(f, val, typ, token.NoPos)
		}

		// Untyped nil constant?  Return interface-typed nil constant.
		if ut_src == tUntypedNil {
			return nilConst(typ)
		}

		// Convert (non-nil) "untyped" literals to their default type.
		if t, ok := ut_src.(*types.Basic); ok && t.Info()&types.IsUntyped != 0 {
			val = emitConv(f, val, DefaultType(ut_src))
		}

		mi := &MakeInterface{X: val}
		mi.setType(typ)
		return f.emit(mi)
	}

	// Conversion of a constant to a non-interface type results in
	// a new constant of the destination type and (initially) the
	// same abstract value.  We don't compute the representation
	// change yet; this defers the point at which the number of
	// possible representations explodes.
	if c, ok := val.(*Const); ok {
		return NewConst(c.Value, typ, c.pos, c.end)
	}

	// A representation-changing conversion.
	c := &Convert{X: val}
	c.setType(typ)
	return f.emit(c)
}

// emitStore emits to f an instruction to store value val at location
// addr, applying implicit conversions as required by assignabilty rules.
//
func emitStore(f *Function, addr, val Value) *Store {
	s := &Store{
		Addr: addr,
		Val:  emitConv(f, val, deref(addr.Type())),
	}
	f.emit(s)
	return s
}

// emitJump emits to f a jump to target, and updates the control-flow graph.
// Postcondition: f.currentBlock is nil.
//
func emitJump(f *Function, target *BasicBlock) {
	b := f.currentBlock
	b.emit(new(Jump))
	addEdge(b, target)
	f.currentBlock = nil
}

// emitIf emits to f a conditional jump to tblock or fblock based on
// cond, and updates the control-flow graph.
// Postcondition: f.currentBlock is nil.
//
func emitIf(f *Function, cond Value, tblock, fblock *BasicBlock) {
	b := f.currentBlock
	b.emit(&If{Cond: cond})
	addEdge(b, tblock)
	addEdge(b, fblock)
	f.currentBlock = nil
}

// emitTrace emits to f an instruction to which acts as a
// placeholder for the kind of high-level event that is
// coming up next: a new statement, the return from a function
// and so on. I'd like this to be a flag an instruction, but that
// was too difficult or ugly to be able for the high-level
// builder call to be able to access the first generated instruction.
// So instead we make it it's own instruction.

func emitTrace(f *Function, event TraceEvent, start token.Pos, end token.Pos) Value {
	t := &Trace{Event: event, Start: start, End: end, Breakpoint: false}
	// fmt.Printf("event %s StartPos %d EndPos %d\n", Event2Name[event])
	fset := f.Prog.Fset
	pkg := f.Pkg
	pkg.locs = append(pkg.locs, LocInst{Pos: start, Fn: nil, Trace: t})
	if (debugMe) {
		fmt.Printf("Emitting event %s\n\tFrom: %s\n\tTo: %s\n",
			Event2Name[event], fset.Position(start), fset.Position(end)	)
	}
	return f.emit(t)
}

// emitExtract emits to f an instruction to extract the index'th
// component of tuple, ascribing it type typ.  It returns the
// extracted value.
//
func emitExtract(f *Function, tuple Value, index int, typ types.Type) Value {
	e := &Extract{Tuple: tuple, Index: index}
	// In all cases but one (tSelect's recv), typ is redundant w.r.t.
	// tuple.Type().(*types.Tuple).Values[index].Type.
	e.setType(typ)
	return f.emit(e)
}

// emitTypeAssert emits to f a type assertion value := x.(t) and
// returns the value.  x.Type() must be an interface.
//
func emitTypeAssert(f *Function, x Value, t types.Type, pos token.Pos) Value {
	// Simplify infallible assertions.
	txi := x.Type().Underlying().(*types.Interface)
	if ti, ok := t.Underlying().(*types.Interface); ok {
		// Even when ti==txi, we still need ChangeInterface
		// since it performs a nil-check.
		if isSuperinterface(ti, txi) {
			c := &ChangeInterface{X: x}
			c.setPos(pos)
			c.setType(t)
			return f.emit(c)
		}
	}

	a := &TypeAssert{X: x, AssertedType: t}
	a.setPos(pos)
	a.setType(t)
	return f.emit(a)
}

// emitTypeTest emits to f a type test value,ok := x.(t) and returns
// a (value, ok) tuple.  x.Type() must be an interface.
//
func emitTypeTest(f *Function, x Value, t types.Type, pos token.Pos) Value {
	// TODO(adonovan): opt: simplify infallible tests as per
	// emitTypeAssert, and return (x, vTrue).
	// (Requires that exprN returns a slice of extracted values,
	// not a single Value of type *types.Tuple.)
	a := &TypeAssert{
		X:            x,
		AssertedType: t,
		CommaOk:      true,
	}
	a.setPos(pos)
	a.setType(types.NewTuple(
		types.NewVar(token.NoPos, nil, "value", t),
		varOk,
	))
	return f.emit(a)
}

// emitTailCall emits to f a function call in tail position.  The
// caller is responsible for all fields of 'call' except its type.
// Intended for wrapper methods.
// Precondition: f does/will not use deferred procedure calls.
// Postcondition: f.currentBlock is nil.
//
func emitTailCall(f *Function, call *Call) {
	tresults := f.Signature.Results()
	nr := tresults.Len()
	if nr == 1 {
		call.typ = tresults.At(0).Type()
	} else {
		call.typ = tresults
	}
	tuple := f.emit(call)
	var ret Ret
	switch nr {
	case 0:
		// no-op
	case 1:
		ret.Results = []Value{tuple}
	default:
		for i := 0; i < nr; i++ {
			v := emitExtract(f, tuple, i, tresults.At(i).Type())
			// TODO(adonovan): in principle, this is required:
			//   v = emitConv(f, o.Type, f.Signature.Results[i].Type)
			// but in practice emitTailCall is only used when
			// the types exactly match.
			ret.Results = append(ret.Results, v)
		}
	}
	f.emit(&ret)
	f.currentBlock = nil
}

// emitImplicitSelections emits to f code to apply the sequence of
// implicit field selections specified by indices to base value v, and
// returns the selected value.
//
func emitImplicitSelections(f *Function, v Value, indices []int) Value {
	for _, index := range indices {
		fld := deref(v.Type()).Underlying().(*types.Struct).Field(index)

		if isPointer(v.Type()) {
			instr := &FieldAddr{
				X:     v,
				Field: index,
			}
			instr.setType(types.NewPointer(fld.Type()))
			v = f.emit(instr)
			// Load the field's value iff indirectly embedded.
			if isPointer(fld.Type()) {
				v = emitLoad(f, v)
			}
		} else {
			instr := &Field{
				X:     v,
				Field: index,
			}
			instr.setType(fld.Type())
			v = f.emit(instr)
		}
	}
	return v
}
