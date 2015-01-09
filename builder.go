// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa2

// This file implements the BUILD phase of SSA construction.
//
// SSA construction has two phases, CREATE and BUILD.  In the CREATE phase
// (create.go), all packages are constructed and type-checked and
// definitions of all package members are created, method-sets are
// computed, and wrapper methods are synthesized.
// ssa.Packages are created in arbitrary order.
//
// In the BUILD phase (builder.go), the builder traverses the AST of
// each Go source function and generates SSA instructions for the
// function body.  Initializer expressions for package-level variables
// are emitted to the package's init() function in the order specified
// by go/types.Info.InitOrder, then code for each function in the
// package is generated in lexical order.
// The BUILD phases for distinct packages are independent and are
// executed in parallel.
//
// TODO(adonovan): indeed, building functions is now embarrassingly parallel.
// Audit for concurrency then benchmark using more goroutines.
//
// The builder's and Program's indices (maps) are populated and
// mutated during the CREATE phase, but during the BUILD phase they
// remain constant.  The sole exception is Prog.methodSets and its
// related maps, which are protected by a dedicated mutex.

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"sync"
	"sync/atomic"

	"github.com/rocky/go-exact"
	"github.com/rocky/go-types"
)

type opaqueType struct {
	types.Type
	name string
}

func (t *opaqueType) String() string { return t.name }

var (
	varOk    = newVar("ok", tBool)
	varIndex = newVar("index", tInt)

	// Type constants.
	tBool       = types.Typ[types.Bool]
	tByte       = types.Typ[types.Byte]
	tInt        = types.Typ[types.Int]
	tInvalid    = types.Typ[types.Invalid]
	tString     = types.Typ[types.String]
	tUntypedNil = types.Typ[types.UntypedNil]
	tRangeIter  = &opaqueType{nil, "iter"} // the type of all "range" iterators
	tEface      = new(types.Interface)

	// SSA Value constants.
	vZero = intConst(0)
	vOne  = intConst(1)
	vTrue = NewConst(exact.MakeBool(true), tBool, token.NoPos, token.NoPos)
)

// builder holds state associated with the package currently being built.
// Its methods contain all the logic for AST-to-SSA conversion.
type builder struct{}

// cond emits to fn code to evaluate boolean condition e and jump
// to t or f depending on its value, performing various simplifications.
//
// Postcondition: fn.currentBlock is nil.
//
func (b *builder) cond(fn *Function, e ast.Expr, t, f *BasicBlock) {
	switch e := e.(type) {
	case *ast.ParenExpr:
		b.cond(fn, e.X, t, f)
		return

	case *ast.BinaryExpr:
		switch e.Op {
		case token.LAND:
			ltrue := fn.newBasicBlock("cond.true", nil)
			b.cond(fn, e.X, ltrue, f)
			fn.currentBlock = ltrue
			b.cond(fn, e.Y, t, f)
			return

		case token.LOR:
			lfalse := fn.newBasicBlock("cond.false", nil)
			b.cond(fn, e.X, t, lfalse)
			fn.currentBlock = lfalse
			b.cond(fn, e.Y, t, f)
			return
		}

	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			b.cond(fn, e.X, f, t)
			return
		}
	}

	// A traditional compiler would simplify "if false" (etc) here
	// but we do not, for better fidelity to the source code.
	//
	// The value of a constant condition may be platform-specific,
	// and may cause blocks that are reachable in some configuration
	// to be hidden from subsequent analyses such as bug-finding tools.
	emitIf(fn, b.expr(fn, e), t, f)
}

// logicalBinop emits code to fn to evaluate e, a &&- or
// ||-expression whose reified boolean value is wanted.
// The value is returned.
//
func (b *builder) logicalBinop(fn *Function, e *ast.BinaryExpr) Value {
	rhs := fn.newBasicBlock("binop.rhs", nil)
	done := fn.newBasicBlock("binop.done", nil)

	// T(e) = T(e.X) = T(e.Y) after untyped constants have been
	// eliminated.
	// TODO(adonovan): not true; MyBool==MyBool yields UntypedBool.
	t := fn.Pkg.typeOf(e)

	var short Value // value of the short-circuit path
	switch e.Op {
	case token.LAND:
		b.cond(fn, e.X, rhs, done)
		short = NewConst(exact.MakeBool(false), t, e.Pos(), e.End())

	case token.LOR:
		b.cond(fn, e.X, done, rhs)
		short = NewConst(exact.MakeBool(true), t, e.Pos(), e.End())
	}

	// Is rhs unreachable?
	if rhs.Preds == nil {
		// Simplify false&&y to false, true||y to true.
		fn.currentBlock = done
		return short
	}

	// Is done unreachable?
	if done.Preds == nil {
		// Simplify true&&y (or false||y) to y.
		fn.currentBlock = rhs
		return b.expr(fn, e.Y)
	}

	// All edges from e.X to done carry the short-circuit value.
	var edges []Value
	for _ = range done.Preds {
		edges = append(edges, short)
	}

	// The edge from e.Y to done carries the value of e.Y.
	fn.currentBlock = rhs
	edges = append(edges, b.expr(fn, e.Y))
	emitJump(fn, done)
	fn.currentBlock = done

	phi := &Phi{Edges: edges, Comment: e.Op.String()}
	phi.pos = e.OpPos
	phi.typ = t
	return done.emit(phi)
}

// exprN lowers a multi-result expression e to SSA form, emitting code
// to fn and returning a single Value whose type is a *types.Tuple.
// The caller must access the components via Extract.
//
// Multi-result expressions include CallExprs in a multi-value
// assignment or return statement, and "value,ok" uses of
// TypeAssertExpr, IndexExpr (when X is a map), and UnaryExpr (when Op
// is token.ARROW).
//
func (b *builder) exprN(fn *Function, e ast.Expr) Value {
	typ := fn.Pkg.typeOf(e).(*types.Tuple)
	switch e := e.(type) {
	case *ast.ParenExpr:
		return b.exprN(fn, e.X)

	case *ast.CallExpr:
		// Currently, no built-in function nor type conversion
		// has multiple results, so we can avoid some of the
		// cases for single-valued CallExpr.
		var c Call
		b.setCall(fn, e, &c.Call)
		c.typ = typ
		return fn.emit(&c)

	case *ast.IndexExpr:
		mapt := fn.Pkg.typeOf(e.X).Underlying().(*types.Map)
		lookup := &Lookup{
			X:       b.expr(fn, e.X),
			Index:   emitConv(fn, b.expr(fn, e.Index), mapt.Key()),
			CommaOk: true,
		}
		lookup.setType(typ)
		lookup.setPos(e.Lbrack)
		return fn.emit(lookup)

	case *ast.TypeAssertExpr:
		t := fn.Pkg.typeOf(e).(*types.Tuple).At(0).Type()
		return emitTypeTest(fn, b.expr(fn, e.X), t, e.X.Pos(), e.X.End())

	case *ast.UnaryExpr: // must be receive <-
		unop := &UnOp{
			Op:      token.ARROW,
			X:       b.expr(fn, e.X),
			CommaOk: true,
		}
		unop.setType(typ)
		unop.setPos(e.OpPos)
		return fn.emit(unop)
	}
	panic(fmt.Sprintf("exprN(%T) in %s", e, fn))
}

// builtin emits to fn SSA instructions to implement a call to the
// built-in function obj with the specified arguments
// and return type.  It returns the value defined by the result.
//
// The result is nil if no special handling was required; in this case
// the caller should treat this like an ordinary library function
// call.
//
func (b *builder) builtin(fn *Function, obj *types.Builtin, args []ast.Expr,
	typ types.Type, pos token.Pos, endP token.Pos) Value {
	switch obj.Name() {
	case "make":
		switch typ.Underlying().(type) {
		case *types.Slice:
			n := b.expr(fn, args[1])
			m := n
			if len(args) == 3 {
				m = b.expr(fn, args[2])
			}
			if m, ok := m.(*Const); ok {
				// treat make([]T, n, m) as new([m]T)[:n]
				cap, _ := exact.Int64Val(m.Value)
				at := types.NewArray(typ.Underlying().(*types.Slice).Elem(), cap)
				alloc := emitNew(fn, at, pos, endP)
				alloc.Comment = "makeslice"
				v := &Slice{
					X:    alloc,
					High: n,
				}
				v.setPos(pos)
				v.setType(typ)
				return fn.emit(v)
			}
			v := &MakeSlice{
				Len: n,
				Cap: m,
			}
			v.setPos(pos)
			v.setType(typ)
			return fn.emit(v)

		case *types.Map:
			var res Value
			if len(args) == 2 {
				res = b.expr(fn, args[1])
			}
			v := &MakeMap{Reserve: res}
			v.setPos(pos)
			v.setType(typ)
			return fn.emit(v)

		case *types.Chan:
			var sz Value = vZero
			if len(args) == 2 {
				sz = b.expr(fn, args[1])
			}
			v := &MakeChan{Size: sz}
			v.setPos(pos)
			v.setType(typ)
			return fn.emit(v)
		}

	case "new":
		alloc := emitNew(fn, deref(typ), pos, endP)
		alloc.Comment = "new"
		return alloc

	case "len", "cap":
		// Special case: len or cap of an array or *array is
		// based on the type, not the value which may be nil.
		// We must still evaluate the value, though.  (If it
		// was side-effect free, the whole call would have
		// been constant-folded.)
		t := deref(fn.Pkg.typeOf(args[0])).Underlying()
		if at, ok := t.(*types.Array); ok {
			b.expr(fn, args[0]) // for effects only
			return intConst(at.Len())
		}
		// Otherwise treat as normal.

	case "panic":
		fn.emit(&Panic{
			X:   emitConv(fn, b.expr(fn, args[0]), tEface),
			pos: pos,
			endP: endP,
		})
		fn.currentBlock = fn.newBasicBlock("unreachable", nil)
		return vTrue // any non-nil Value will do
	}
	return nil // treat all others as a regular function call
}

// addr lowers a single-result addressable expression e to SSA form,
// emitting code to fn and returning the location (an lvalue) defined
// by the expression.
//
// If escaping is true, addr marks the base variable of the
// addressable expression e as being a potentially escaping pointer
// value.  For example, in this code:
//
//   a := A{
//     b: [1]B{B{c: 1}}
//   }
//   return &a.b[0].c
//
// the application of & causes a.b[0].c to have its address taken,
// which means that ultimately the local variable a must be
// heap-allocated.  This is a simple but very conservative escape
// analysis.
//
// Operations forming potentially escaping pointers include:
// - &x, including when implicit in method call or composite literals.
// - a[:] iff a is an array (not *array)
// - references to variables in lexically enclosing functions.
//
func (b *builder) addr(fn *Function, e ast.Expr, escaping bool) lvalue {
	switch e := e.(type) {
	case *ast.Ident:
		if isBlankIdent(e) {
			return blank{}
		}
		obj := fn.Pkg.objectOf(e)
		v := fn.Prog.packageLevelValue(obj) // var (address)
		if v == nil {
			v = fn.lookup(obj, escaping)
		}
		return &address{addr: v, pos: e.Pos(), expr: e}

	case *ast.CompositeLit:
		t := deref(fn.Pkg.typeOf(e))
		// obj := fn.Pkg.objectOf(e)  // ?? FIXME figure out how to do
		// scope := fn.Pkg.TypeScope2Scope[obj.Parent()]
		var v *Alloc
		if escaping {
			v = emitNew(fn, t, e.Pos(), e.End())
		} else {
			v = fn.addLocal(t, e.Pos(), e.End(), nil) // scope
		}
		v.Comment = "complit"
		b.compLit(fn, v, e, true) // initialize in place
		return &address{addr: v, pos: e.Lbrace, expr: e}

	case *ast.ParenExpr:
		return b.addr(fn, e.X, escaping)

	case *ast.SelectorExpr:
		sel, ok := fn.Pkg.info.Selections[e]
		if !ok {
			// qualified identifier
			return b.addr(fn, e.Sel, escaping)
		}
		if sel.Kind() != types.FieldVal {
			panic(sel)
		}
		wantAddr := true
		v := b.receiver(fn, e.X, wantAddr, escaping, sel)
		last := len(sel.Index()) - 1
		return &address{
			addr: emitFieldSelection(fn, v, sel.Index()[last], true, e.Sel),
			pos:  e.Sel.Pos(),
			expr: e.Sel,
		}

	case *ast.IndexExpr:
		var x Value
		var et types.Type
		switch t := fn.Pkg.typeOf(e.X).Underlying().(type) {
		case *types.Array:
			x = b.addr(fn, e.X, escaping).address(fn)
			et = types.NewPointer(t.Elem())
		case *types.Pointer: // *array
			x = b.expr(fn, e.X)
			et = types.NewPointer(t.Elem().Underlying().(*types.Array).Elem())
		case *types.Slice:
			x = b.expr(fn, e.X)
			et = types.NewPointer(t.Elem())
		case *types.Map:
			return &element{
				m:   b.expr(fn, e.X),
				k:   emitConv(fn, b.expr(fn, e.Index), t.Key()),
				t:   t.Elem(),
				pos: e.Lbrack,
			}
		default:
			panic("unexpected container type in IndexExpr: " + t.String())
		}
		v := &IndexAddr{
			X:     x,
			Index: emitConv(fn, b.expr(fn, e.Index), tInt),
		}
		v.setPos(e.Lbrack)
		v.setType(et)
		return &address{addr: fn.emit(v), pos: e.Lbrack, expr: e}

	case *ast.StarExpr:
		return &address{addr: b.expr(fn, e.X), pos: e.Star, expr: e}
	}

	panic(fmt.Sprintf("unexpected address expression: %T", e))
}

// exprInPlace emits to fn code to initialize the lvalue loc with the
// value of expression e. If isZero is true, exprInPlace assumes that loc
// holds the zero value for its type.
//
// This is equivalent to loc.store(fn, b.expr(fn, e)) but may
// generate better code in some cases, e.g. for composite literals
// in an addressable location.
//
func (b *builder) exprInPlace(fn *Function, loc lvalue, e ast.Expr, isZero bool) {
	if e, ok := unparen(e).(*ast.CompositeLit); ok {
		// A CompositeLit never evaluates to a pointer,
		// so if the type of the location is a pointer,
		// an &-operation is implied.
		if _, ok := loc.(blank); !ok { // avoid calling blank.typ()
			if isPointer(loc.typ()) {
				ptr := b.addr(fn, e, true).address(fn)
				loc.store(fn, ptr) // copy address
				return
			}
		}

		if _, ok := loc.(*address); ok {
			if isInterface(loc.typ()) {
				// e.g. var x interface{} = T{...}
				// Can't in-place initialize an interface value.
				// Fall back to copying.
			} else {
				addr := loc.address(fn)
				b.compLit(fn, addr, e, isZero) // in place
				emitDebugRef(fn, e, addr, true)
				return
			}
		}
	}
	loc.store(fn, b.expr(fn, e)) // copy value
}

// expr lowers a single-result expression e to SSA form, emitting code
// to fn and returning the Value defined by the expression.
//
func (b *builder) expr(fn *Function, e ast.Expr) Value {
	e = unparen(e)

	tv := fn.Pkg.info.Types[e]

	// Is expression a constant?
	if tv.Value != nil {
		return NewConst(tv.Value, tv.Type, e.Pos(), e.End())
	}

	var v Value
	if tv.Addressable() {
		// Prefer pointer arithmetic ({Index,Field}Addr) followed
		// by Load over subelement extraction (e.g. Index, Field),
		// to avoid large copies.
		v = b.addr(fn, e, false).load(fn)
	} else {
		v = b.expr0(fn, e, tv)
	}
	if fn.debugInfo() {
		emitDebugRef(fn, e, v, false)
	}
	return v
}

func (b *builder) expr0(fn *Function, e ast.Expr, tv types.TypeAndValue) Value {
	switch e := e.(type) {
	case *ast.BasicLit:
		panic("non-constant BasicLit") // unreachable

	case *ast.FuncLit:
		fn2 := &Function{
			name:      fmt.Sprintf("%s$%d", fn.Name(), 1+len(fn.AnonFuncs)),
			Signature: fn.Pkg.typeOf(e.Type).Underlying().(*types.Signature),
			pos:       e.Type.Func,
			parent:    fn,
			Pkg:       fn.Pkg,
			Prog:      fn.Prog,
			syntax:    e,
			Breakpoint: false,
			Scope     : astScope(fn, e),
			LocalsByName: make(map[NameScope]uint),
			endP:       e.Body.End(),
		}
		fn.AnonFuncs = append(fn.AnonFuncs, fn2)
		b.buildFunction(fn2)
		if fn2.FreeVars == nil {
			return fn2
		}
		v := &MakeClosure{Fn: fn2}
		v.setType(tv.Type)
		for _, fv := range fn2.FreeVars {
			v.Bindings = append(v.Bindings, fv.outer)
			fv.outer = nil
		}
		return fn.emit(v)

	case *ast.TypeAssertExpr: // single-result form only
		return emitTypeAssert(fn, b.expr(fn, e.X), tv.Type, e.X.Pos())

	case *ast.CallExpr:
		if fn.Pkg.info.Types[e.Fun].IsType() {
			// Explicit type conversion, e.g. string(x) or big.Int(x)
			x := b.expr(fn, e.Args[0])
			y := emitConv(fn, x, tv.Type)
			if y != x {
				switch y := y.(type) {
				case *Convert:
					y.pos = e.Lparen
					y.endP = e.Rparen
				case *ChangeType:
					y.pos = e.Lparen
					y.endP = e.Rparen
				case *MakeInterface:
					y.pos = e.Lparen
					y.endP = e.Rparen
				}
			}
			return y
		}
		// Call to "intrinsic" built-ins, e.g. new, make, panic.
		if id, ok := unparen(e.Fun).(*ast.Ident); ok {
			if obj, ok := fn.Pkg.info.Uses[id].(*types.Builtin); ok {
				if v := b.builtin(fn, obj, e.Args, tv.Type, e.Pos(), e.End()); v != nil {
					return v
				}
			}
		}
		// Regular function call.
		var v Call
		b.setCall(fn, e, &v.Call)
		v.setType(tv.Type)
		return fn.emit(&v)

	case *ast.UnaryExpr:
		switch e.Op {
		case token.AND: // &X --- potentially escaping.
			addr := b.addr(fn, e.X, true)
			if _, ok := unparen(e.X).(*ast.StarExpr); ok {
				// &*p must panic if p is nil (http://golang.org/s/go12nil).
				// For simplicity, we'll just (suboptimally) rely
				// on the side effects of a load.
				// TODO(adonovan): emit dedicated nilcheck.
				addr.load(fn)
			}
			return addr.address(fn)
		case token.ADD:
			return b.expr(fn, e.X)
		case token.NOT, token.ARROW, token.SUB, token.XOR: // ! <- - ^
			v := &UnOp{
				Op: e.Op,
				X:  b.expr(fn, e.X),
			}
			v.setPos(e.OpPos)
			v.setType(tv.Type)
			return fn.emit(v)
		default:
			panic(e.Op)
		}

	case *ast.BinaryExpr:
		switch e.Op {
		case token.LAND, token.LOR:
			return b.logicalBinop(fn, e)
		case token.SHL, token.SHR:
			fallthrough
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.AND, token.OR, token.XOR, token.AND_NOT:
			return emitArith(fn, e.Op, b.expr(fn, e.X), b.expr(fn, e.Y), tv.Type, e.OpPos)

		case token.EQL, token.NEQ, token.GTR, token.LSS, token.LEQ, token.GEQ:
			cmp := emitCompare(fn, e.Op, b.expr(fn, e.X), b.expr(fn, e.Y), e.OpPos)
			// The type of x==y may be UntypedBool.
			return emitConv(fn, cmp, DefaultType(tv.Type))
		default:
			panic("illegal op in BinaryExpr: " + e.Op.String())
		}

	case *ast.SliceExpr:
		var low, high, max Value
		var x Value
		switch fn.Pkg.typeOf(e.X).Underlying().(type) {
		case *types.Array:
			// Potentially escaping.
			x = b.addr(fn, e.X, true).address(fn)
		case *types.Basic, *types.Slice, *types.Pointer: // *array
			x = b.expr(fn, e.X)
		default:
			unreachable()
		}
		if e.High != nil {
			high = b.expr(fn, e.High)
		}
		if e.Low != nil {
			low = b.expr(fn, e.Low)
		}
		if e.Slice3 {
			max = b.expr(fn, e.Max)
		}
		v := &Slice{
			X:    x,
			Low:  low,
			High: high,
			Max:  max,
		}
		v.setPos(e.Lbrack)
		v.setType(tv.Type)
		return fn.emit(v)

	case *ast.Ident:
		obj := fn.Pkg.info.Uses[e]
		// Universal built-in or nil?
		switch obj := obj.(type) {
		case *types.Builtin:
			return &Builtin{name: obj.Name(), sig: tv.Type.(*types.Signature)}
		case *types.Nil:
			return nilConst(tv.Type)
		}
		// Package-level func or var?
		if v := fn.Prog.packageLevelValue(obj); v != nil {
			if _, ok := obj.(*types.Var); ok {
				return emitLoad(fn, v) // var (address)
			}
			return v // (func)
		}
		// Local var.
		return emitLoad(fn, fn.lookup(obj, false)) // var (address)

	case *ast.SelectorExpr:
		sel, ok := fn.Pkg.info.Selections[e]
		if !ok {
			// qualified identifier
			return b.expr(fn, e.Sel)
		}
		switch sel.Kind() {
		case types.MethodExpr:
			// (*T).f or T.f, the method f from the method-set of type T.
			// The result is a "thunk".
			return emitConv(fn, makeThunk(fn.Prog, sel), tv.Type)

		case types.MethodVal:
			// e.f where e is an expression and f is a method.
			// The result is a "bound".
			obj := sel.Obj().(*types.Func)
			rt := recvType(obj)
			wantAddr := isPointer(rt)
			escaping := true
			v := b.receiver(fn, e.X, wantAddr, escaping, sel)
			if isInterface(rt) {
				// If v has interface type I,
				// we must emit a check that v is non-nil.
				// We use: typeassert v.(I).
				emitTypeAssert(fn, v, rt, token.NoPos)
			}
			c := &MakeClosure{
				Fn:       makeBound(fn.Prog, obj),
				Bindings: []Value{v},
			}
			c.setPos(e.Sel.Pos())
			c.setType(tv.Type)
			return fn.emit(c)

		case types.FieldVal:
			indices := sel.Index()
			last := len(indices) - 1
			v := b.expr(fn, e.X)
			v = emitImplicitSelections(fn, v, indices[:last])
			v = emitFieldSelection(fn, v, indices[last], false, e.Sel)
			return v
		}

		panic("unexpected expression-relative selector")

	case *ast.IndexExpr:
		switch t := fn.Pkg.typeOf(e.X).Underlying().(type) {
		case *types.Array:
			// Non-addressable array (in a register).
			v := &Index{
				X:     b.expr(fn, e.X),
				Index: emitConv(fn, b.expr(fn, e.Index), tInt),
			}
			v.setPos(e.Lbrack)
			v.setType(t.Elem())
			return fn.emit(v)

		case *types.Map:
			// Maps are not addressable.
			mapt := fn.Pkg.typeOf(e.X).Underlying().(*types.Map)
			v := &Lookup{
				X:     b.expr(fn, e.X),
				Index: emitConv(fn, b.expr(fn, e.Index), mapt.Key()),
			}
			v.setPos(e.Lbrack)
			v.setType(mapt.Elem())
			return fn.emit(v)

		case *types.Basic: // => string
			// Strings are not addressable.
			v := &Lookup{
				X:     b.expr(fn, e.X),
				Index: b.expr(fn, e.Index),
			}
			v.setPos(e.Lbrack)
			v.setType(tByte)
			return fn.emit(v)

		case *types.Slice, *types.Pointer: // *array
			// Addressable slice/array; use IndexAddr and Load.
			return b.addr(fn, e, false).load(fn)

		default:
			panic("unexpected container type in IndexExpr: " + t.String())
		}

	case *ast.CompositeLit, *ast.StarExpr:
		// Addressable types (lvalues)
		return b.addr(fn, e, false).load(fn)
	}

	panic(fmt.Sprintf("unexpected expr: %T", e))
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt, scope *Scope) {
	for _, s := range list {
		// Don't add a trace statement in an "if" or "for" because
		// those statement add a trace themselves.
		switch s.(type) {
		case *ast.IfStmt, *ast.ForStmt:
		default:
			emitTraceStmt(fn, STMT_IN_LIST, s)
		}
		b.stmt(fn, s, scope)
	}
}

// receiver emits to fn code for expression e in the "receiver"
// position of selection e.f (where f may be a field or a method) and
// returns the effective receiver after applying the implicit field
// selections of sel.
//
// wantAddr requests that the result is an an address.  If
// !sel.Indirect(), this may require that e be build in addr() mode; it
// must thus be addressable.
//
// escaping is defined as per builder.addr().
//
func (b *builder) receiver(fn *Function, e ast.Expr, wantAddr, escaping bool, sel *types.Selection) Value {
	var v Value
	if wantAddr && !sel.Indirect() && !isPointer(fn.Pkg.typeOf(e)) {
		v = b.addr(fn, e, escaping).address(fn)
	} else {
		v = b.expr(fn, e)
	}

	last := len(sel.Index()) - 1
	v = emitImplicitSelections(fn, v, sel.Index()[:last])
	if !wantAddr && isPointer(v.Type()) {
		v = emitLoad(fn, v)
	}
	return v
}

// setCallFunc populates the function parts of a CallCommon structure
// (Func, Method, Recv, Args[0]) based on the kind of invocation
// occurring in e.
//
func (b *builder) setCallFunc(fn *Function, e *ast.CallExpr, c *CallCommon) {
	c.pos  = e.Pos()
	c.endP = e.End()

	// Is this a method call?
	if selector, ok := unparen(e.Fun).(*ast.SelectorExpr); ok {
		sel, ok := fn.Pkg.info.Selections[selector]
		if ok && sel.Kind() == types.MethodVal {
			obj := sel.Obj().(*types.Func)
			wantAddr := isPointer(recvType(obj))
			escaping := true
			v := b.receiver(fn, selector.X, wantAddr, escaping, sel)
			if isInterface(deref(v.Type())) {
				// Invoke-mode call.
				c.Value = v
				c.Method = obj
			} else {
				// "Call"-mode call.
				c.Value = fn.Prog.declaredFunc(obj)
				c.Args = append(c.Args, v)
			}
			return
		}

		// sel.Kind()==MethodExpr indicates T.f() or (*T).f():
		// a statically dispatched call to the method f in the
		// method-set of T or *T.  T may be an interface.
		//
		// e.Fun would evaluate to a concrete method, interface
		// wrapper function, or promotion wrapper.
		//
		// For now, we evaluate it in the usual way.
		//
		// TODO(adonovan): opt: inline expr() here, to make the
		// call static and to avoid generation of wrappers.
		// It's somewhat tricky as it may consume the first
		// actual parameter if the call is "invoke" mode.
		//
		// Examples:
		//  type T struct{}; func (T) f() {}   // "call" mode
		//  type T interface { f() }           // "invoke" mode
		//
		//  type S struct{ T }
		//
		//  var s S
		//  S.f(s)
		//  (*S).f(&s)
		//
		// Suggested approach:
		// - consume the first actual parameter expression
		//   and build it with b.expr().
		// - apply implicit field selections.
		// - use MethodVal logic to populate fields of c.
	}

	// Evaluate the function operand in the usual way.
	c.Value = b.expr(fn, e.Fun)
}

// emitCallArgs emits to f code for the actual parameters of call e to
// a (possibly built-in) function of effective type sig.
// The argument values are appended to args, which is then returned.
//
func (b *builder) emitCallArgs(fn *Function, sig *types.Signature, e *ast.CallExpr, args []Value) []Value {
	// f(x, y, z...): pass slice z straight through.
	if e.Ellipsis != 0 {
		for i, arg := range e.Args {
			v := emitConv(fn, b.expr(fn, arg), sig.Params().At(i).Type())
			args = append(args, v)
		}
		return args
	}

	offset := len(args) // 1 if call has receiver, 0 otherwise

	// Evaluate actual parameter expressions.
	//
	// If this is a chained call of the form f(g()) where g has
	// multiple return values (MRV), they are flattened out into
	// args; a suffix of them may end up in a varargs slice.
	for _, arg := range e.Args {
		v := b.expr(fn, arg)
		if ttuple, ok := v.Type().(*types.Tuple); ok { // MRV chain
			for i, n := 0, ttuple.Len(); i < n; i++ {
				args = append(args, emitExtract(fn, v, i))
			}
		} else {
			args = append(args, v)
		}
	}

	// Actual->formal assignability conversions for normal parameters.
	np := sig.Params().Len() // number of normal parameters
	if sig.Variadic() {
		np--
	}
	for i := 0; i < np; i++ {
		args[offset+i] = emitConv(fn, args[offset+i], sig.Params().At(i).Type())
	}

	// Actual->formal assignability conversions for variadic parameter,
	// and construction of slice.
	if sig.Variadic() {
		varargs := args[offset+np:]
		st := sig.Params().At(np).Type().(*types.Slice)
		vt := st.Elem()
		if len(varargs) == 0 {
			args = append(args, nilConst(st))
		} else {
			// Replace a suffix of args with a slice containing it.
			at := types.NewArray(vt, int64(len(varargs)))
			a := emitNew(fn, at, token.NoPos, token.NoPos)
			a.setPos(e.Rparen)
			a.Comment = "varargs"
			for i, arg := range varargs {
				iaddr := &IndexAddr{
					X:     a,
					Index: intConst(int64(i)),
				}
				iaddr.setType(types.NewPointer(vt))
				fn.emit(iaddr)
				emitStore(fn, iaddr, arg, arg.Pos())
			}
			s := &Slice{X: a}
			s.setType(st)
			args[offset+np] = fn.emit(s)
			args = args[:offset+np+1]
		}
	}
	return args
}

// setCall emits to fn code to evaluate all the parameters of a function
// call e, and populates *c with those values.
//
func (b *builder) setCall(fn *Function, e *ast.CallExpr, c *CallCommon) {
	// First deal with the f(...) part and optional receiver.
	b.setCallFunc(fn, e, c)

	// Then append the other actual parameters.
	sig, _ := fn.Pkg.typeOf(e.Fun).Underlying().(*types.Signature)
	if sig == nil {
		panic(fmt.Sprintf("no signature for call of %s", e.Fun))
	}
	c.Args = b.emitCallArgs(fn, sig, e, c.Args)
}

// assignOp emits to fn code to perform loc += incr or loc -= incr.
func (b *builder) assignOp(fn *Function, loc lvalue, incr Value, op token.Token) {
	oldv := loc.load(fn)
	loc.store(fn, emitArith(fn, op, oldv, emitConv(fn, incr, oldv.Type()), loc.typ(), token.NoPos))
}

// localValueSpec emits to fn code to define all of the vars in the
// function-local ValueSpec, spec.
//
func (b *builder) localValueSpec(fn *Function, spec *ast.ValueSpec) {
	switch {
	case len(spec.Values) == len(spec.Names):
		// e.g. var x, y = 0, 1
		// 1:1 assignment
		for i, id := range spec.Names {
			if !isBlankIdent(id) {
				fn.addLocalForIdent(id)
			}
			lval := b.addr(fn, id, false) // non-escaping
			b.exprInPlace(fn, lval, spec.Values[i], true)
		}

	case len(spec.Values) == 0:
		// e.g. var x, y int
		// Locals are implicitly zero-initialized.
		for _, id := range spec.Names {
			if !isBlankIdent(id) {
				lhs := fn.addLocalForIdent(id)
				if fn.debugInfo() {
					emitDebugRef(fn, id, lhs, true)
				}
			}
		}

	default:
		// e.g. var x, y = pos()
		tuple := b.exprN(fn, spec.Values[0])
		for i, id := range spec.Names {
			if !isBlankIdent(id) {
				fn.addLocalForIdent(id)
				lhs := b.addr(fn, id, false) // non-escaping
				lhs.store(fn, emitExtract(fn, tuple, i))
			}
		}
	}
}

// assignStmt emits code to fn for a parallel assignment of rhss to lhss.
// isDef is true if this is a short variable declaration (:=).
//
// Note the similarity with localValueSpec.
//
func (b *builder) assignStmt(fn *Function, lhss, rhss []ast.Expr, isDef bool,
	scope *Scope) {
	// Side effects of all LHSs and RHSs must occur in left-to-right order.
	var lvals []lvalue
	for _, lhs := range lhss {
		var lval lvalue = blank{}
		if !isBlankIdent(lhs) {
			if isDef {
				if obj := fn.Pkg.info.Defs[lhs.(*ast.Ident)]; obj != nil {
					fn.addNamedLocal(obj)
				}
			}
			lval = b.addr(fn, lhs, false) // non-escaping
		}
		lvals = append(lvals, lval)
	}
	if len(lhss) == len(rhss) {
		// e.g. x, y = f(), g()
		if len(lhss) == 1 {
			// x = type{...}
			// Optimization: in-place construction
			// of composite literals.
			b.exprInPlace(fn, lvals[0], rhss[0], false)
		} else {
			// Parallel assignment.  All reads must occur
			// before all updates, precluding exprInPlace.
			var rvals []Value
			for _, rval := range rhss {
				rvals = append(rvals, b.expr(fn, rval))
			}
			for i, lval := range lvals {
				lval.storeWithScope(fn, rvals[i], scope)
			}
		}
	} else {
		// e.g. x, y = pos()
		tuple := b.exprN(fn, rhss[0])
		for i, lval := range lvals {
			lval.storeWithScope(fn, emitExtract(fn, tuple, i), scope)
		}
	}
}

// arrayLen returns the length of the array whose composite literal elements are elts.
func (b *builder) arrayLen(fn *Function, elts []ast.Expr) int64 {
	var max int64 = -1
	var i int64 = -1
	for _, e := range elts {
		if kv, ok := e.(*ast.KeyValueExpr); ok {
			i = b.expr(fn, kv.Key).(*Const).Int64()
		} else {
			i++
		}
		if i > max {
			max = i
		}
	}
	return max + 1
}

// compLit emits to fn code to initialize a composite literal e at
// address addr with type typ, typically allocated by Alloc.
// Nested composite literals are recursively initialized in place
// where possible. If isZero is true, compLit assumes that addr
// holds the zero value for typ.
//
// A CompositeLit may have pointer type only in the recursive (nested)
// case when the type name is implicit.  e.g. in []*T{{}}, the inner
// literal has type *T behaves like &T{}.
// In that case, addr must hold a T, not a *T.
//
func (b *builder) compLit(fn *Function, addr Value, e *ast.CompositeLit, isZero bool) {
	typ := deref(fn.Pkg.typeOf(e))
	switch t := typ.Underlying().(type) {
	case *types.Struct:
		if !isZero && len(e.Elts) != t.NumFields() {
			emitMemClear(fn, addr, e.Lbrace)
			isZero = true
		}
		for i, e := range e.Elts {
			fieldIndex := i
			pos := e.Pos()
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				fname := kv.Key.(*ast.Ident).Name
				for i, n := 0, t.NumFields(); i < n; i++ {
					sf := t.Field(i)
					if sf.Name() == fname {
						fieldIndex = i
						pos = kv.Colon
						e = kv.Value
						break
					}
				}
			}
			sf := t.Field(fieldIndex)
			faddr := &FieldAddr{
				X:     addr,
				Field: fieldIndex,
			}
			faddr.setType(types.NewPointer(sf.Type()))
			fn.emit(faddr)
			b.exprInPlace(fn, &address{addr: faddr, pos: pos, expr: e}, e, isZero)
		}

	case *types.Array, *types.Slice:
		var at *types.Array
		var array Value
		switch t := t.(type) {
		case *types.Slice:
			at = types.NewArray(t.Elem(), b.arrayLen(fn, e.Elts))
			alloc := emitNew(fn, at, e.Pos(), e.End())
			alloc.Comment = "slicelit"
			array = alloc
			isZero = true
		case *types.Array:
			at = t
			array = addr
		}

		if !isZero && int64(len(e.Elts)) != at.Len() {
			emitMemClear(fn, array, e.Lbrace)
			isZero = true
		}

		var idx *Const
		for _, e := range e.Elts {
			if kv, ok := e.(*ast.KeyValueExpr); ok {
				idx = b.expr(fn, kv.Key).(*Const)
				e = kv.Value
			} else {
				var idxval int64
				if idx != nil {
					idxval = idx.Int64() + 1
				}
				idx = intConst(idxval)
			}
			iaddr := &IndexAddr{
				X:     array,
				Index: idx,
			}
			iaddr.setType(types.NewPointer(at.Elem()))
			fn.emit(iaddr)
			b.exprInPlace(fn, &address{addr: iaddr, pos: e.Pos(), expr: e}, e, isZero)
		}
		if t != at { // slice
			s := &Slice{X: array}
			s.setPos(e.Lbrace)
			s.setType(typ)
			emitStore(fn, addr, fn.emit(s), e.Lbrace)
		}

	case *types.Map:
		m := &MakeMap{Reserve: intConst(int64(len(e.Elts)))}
		m.setPos(e.Lbrace)
		m.setType(typ)
		emitStore(fn, addr, fn.emit(m), e.Lbrace)
		for _, e := range e.Elts {
			e := e.(*ast.KeyValueExpr)
			loc := &element{
				m:   m,
				k:   emitConv(fn, b.expr(fn, e.Key), t.Key()),
				t:   t.Elem(),
				pos: e.Colon,
			}
			b.exprInPlace(fn, loc, e.Value, true)
		}

	default:
		panic("unexpected CompositeLit type: " + t.String())
	}
}

// switchStmt emits to fn code for the switch statement s, optionally
// labelled by label.
//
func (b *builder) switchStmt(fn *Function, s *ast.SwitchStmt, label *lblock) {
	// We treat SwitchStmt like a sequential if-else chain.
	// Multiway dispatch can be recovered later by ssautil.Switches()
	// to those cases that are free of side effects.
	switchScope := astScope(fn, s)
	parent      := ParentScope(fn, switchScope)
	if s.Init != nil {
		b.stmt(fn, s.Init, parent)
	}
	var tag Value = vTrue
	if s.Tag != nil {
		emitTraceExpr(fn, SWITCH_COND, s.Tag)
		tag = b.expr(fn, s.Tag)
	}
	done := fn.newBasicBlock("switch.done", parent)
	if label != nil {
		label._break = done
	}
	// We pull the default case (if present) down to the end.
	// But each fallthrough label must point to the next
	// body block in source order, so we preallocate a
	// body block (fallthru) for the next case.
	// Unfortunately this makes for a confusing block order.
	var dfltBody *[]ast.Stmt
	var dfltFallthrough *BasicBlock
	var fallthru, dfltBlock *BasicBlock
	ncases := len(s.Body.List)
	for i, clause := range s.Body.List {
		body := fallthru
		if body == nil {
			// first case only
			body = fn.newBasicBlock("switch.body", switchScope)
		}

		// Preallocate body block for the next case.
		fallthru = done
		if i+1 < ncases {
			fallthru = fn.newBasicBlock("switch.body", switchScope)
		}

		cc := clause.(*ast.CaseClause)
		if cc.List == nil {
			// Default case.
			dfltBody = &cc.Body
			dfltFallthrough = fallthru
			dfltBlock = body
			continue
		}

		var nextCond *BasicBlock
		for _, cond := range cc.List {
			nextCond = fn.newBasicBlock("switch.next", switchScope)
			// TODO(adonovan): opt: when tag==vTrue, we'd
			// get better code if we use b.cond(cond)
			// instead of BinOp(EQL, tag, b.expr(cond))
			// followed by If.  Don't forget conversions
			// though.
			emitTraceExpr(fn, SWITCH_COND, cond)
			cond := emitCompare(fn, token.EQL, tag, b.expr(fn, cond), token.NoPos)
			emitIf(fn, cond, body, nextCond)
			fn.currentBlock = nextCond
		}
		fn.currentBlock = body
		fn.targets = &targets{
			tail:         fn.targets,
			_break:       done,
			_fallthrough: fallthru,
		}
		b.stmtList(fn, cc.Body, switchScope)
		fn.targets = fn.targets.tail
		emitJump(fn, done)
		fn.currentBlock = nextCond
	}
	if dfltBlock != nil {
		emitJump(fn, dfltBlock)
		fn.currentBlock = dfltBlock
		fn.targets = &targets{
			tail:         fn.targets,
			_break:       done,
			_fallthrough: dfltFallthrough,
		}
		b.stmtList(fn, *dfltBody, switchScope)
		fn.targets = fn.targets.tail
	}
	emitJump(fn, done)
	fn.currentBlock = done
}

// typeSwitchStmt emits to fn code for the type switch statement s, optionally
// labelled by label.
//
func (b *builder) typeSwitchStmt(fn *Function, s *ast.TypeSwitchStmt, label *lblock) {
	// We treat TypeSwitchStmt like a sequential if-else chain.
	// Multiway dispatch can be recovered later by ssautil.Switches().

	// Typeswitch lowering:
	//
	// var x X
	// switch y := x.(type) {
	// case T1, T2: S1                  // >1 	(y := x)
	// case nil:    SN                  // nil 	(y := x)
	// default:     SD                  // 0 types 	(y := x)
	// case T3:     S3                  // 1 type 	(y := x.(T3))
	// }
	//
	//      ...s.Init...
	// 	x := eval x
	// .caseT1:
	// 	t1, ok1 := typeswitch,ok x <T1>
	// 	if ok1 then goto S1 else goto .caseT2
	// .caseT2:
	// 	t2, ok2 := typeswitch,ok x <T2>
	// 	if ok2 then goto S1 else goto .caseNil
	// .S1:
	//      y := x
	// 	...S1...
	// 	goto done
	// .caseNil:
	// 	if t2, ok2 := typeswitch,ok x <T2>
	// 	if x == nil then goto SN else goto .caseT3
	// .SN:
	//      y := x
	// 	...SN...
	// 	goto done
	// .caseT3:
	// 	t3, ok3 := typeswitch,ok x <T3>
	// 	if ok3 then goto S3 else goto default
	// .S3:
	//      y := t3
	// 	...S3...
	// 	goto done
	// .default:
	//      y := x
	// 	...SD...
	// 	goto done
	// .done:

	typeSwitchScope := astScope(fn, s)
	if s.Init != nil {
		b.stmt(fn, s.Init, typeSwitchScope)
	}

	var x Value
	switch ass := s.Assign.(type) {
	case *ast.ExprStmt: // x.(type)
		x = b.expr(fn, unparen(ass.X).(*ast.TypeAssertExpr).X)
	case *ast.AssignStmt: // y := x.(type)
		emitTraceStmt(fn, ASSIGN_STMT, s)
		x = b.expr(fn, unparen(ass.Rhs[0]).(*ast.TypeAssertExpr).X)
	}

	done := fn.newBasicBlock("typeswitch.done", ParentScope(fn, typeSwitchScope))
	if label != nil {
		label._break = done
	}
	var default_ *ast.CaseClause
	for _, clause := range s.Body.List {
		cc := clause.(*ast.CaseClause)
		if cc.List == nil {
			default_ = cc
			continue
		}
		body := fn.newBasicBlock("typeswitch.body", astScope(fn, s.Body))
		var next *BasicBlock
		var casetype types.Type
		var ti Value // ti, ok := typeassert,ok x <Ti>
		for _, cond := range cc.List {
			next = fn.newBasicBlock("typeswitch.next", astScope(fn, s.Body))
			casetype = fn.Pkg.typeOf(cond)
			var condv Value
			if casetype == tUntypedNil {
				condv = emitCompare(fn, token.EQL, x, nilConst(x.Type()), token.NoPos)
				ti = x
			} else {
				yok := emitTypeTest(fn, x, casetype, cc.Case, token.NoPos)
				ti = emitExtract(fn, yok, 0)
				condv = emitExtract(fn, yok, 1)
			}
			emitIf(fn, condv, body, next)
			fn.currentBlock = next
		}
		if len(cc.List) != 1 {
			ti = x
		}
		fn.currentBlock = body
		b.typeCaseBody(fn, cc, ti, done, typeSwitchScope)
		fn.currentBlock = next
	}
	if default_ != nil {
		b.typeCaseBody(fn, default_, x, done, typeSwitchScope)
	} else {
		emitJump(fn, done)
	}
	fn.currentBlock = done
}

func (b *builder) typeCaseBody(fn *Function, cc *ast.CaseClause, x Value, done *BasicBlock,
	scope *Scope) {
	if obj := fn.Pkg.info.Implicits[cc]; obj != nil {
		// In a switch y := x.(type), each case clause
		// implicitly declares a distinct object y.
		// In a single-type case, y has that type.
		// In multi-type cases, 'case nil' and default,
		// y has the same type as the interface operand.
		emitStore(fn, fn.addNamedLocal(obj), x, obj.Pos())
	}
	fn.targets = &targets{
		tail:   fn.targets,
		_break: done,
	}
	b.stmtList(fn, cc.Body, scope)
	fn.targets = fn.targets.tail
	emitJump(fn, done)
}

// selectStmt emits to fn code for the select statement s, optionally
// labelled by label.
//
func (b *builder) selectStmt(fn *Function, s *ast.SelectStmt, label *lblock) {
	// A blocking select of a single case degenerates to a
	// simple send or receive.
	// TODO(adonovan): opt: is this optimization worth its weight?
	selectScope := astScope(fn, s)
	if len(s.Body.List) == 1 {
		clause := s.Body.List[0].(*ast.CommClause)
		if clause.Comm != nil {
			b.stmt(fn, clause.Comm, selectScope)
			done := fn.newBasicBlock("select.done", selectScope)
			if label != nil {
				label._break = done
			}
			fn.targets = &targets{
				tail:   fn.targets,
				_break: done,
			}
			b.stmtList(fn, clause.Body, selectScope)
			fn.targets = fn.targets.tail
			emitJump(fn, done)
			fn.currentBlock = done
			return
		}
	}

	// First evaluate all channels in all cases, and find
	// the directions of each state.
	var states []*SelectState
	blocking := true
	debugInfo := fn.debugInfo()
	for _, clause := range s.Body.List {
		var st *SelectState
		switch comm := clause.(*ast.CommClause).Comm.(type) {
		case nil: // default case
			blocking = false
			continue

		case *ast.SendStmt: // ch<- i
			ch := b.expr(fn, comm.Chan)
			st = &SelectState{
				Dir:  types.SendOnly,
				Chan: ch,
				Send: emitConv(fn, b.expr(fn, comm.Value),
					ch.Type().Underlying().(*types.Chan).Elem()),
				Pos: comm.Arrow,
			}
			if debugInfo {
				st.DebugNode = comm
			}

		case *ast.AssignStmt: // x := <-ch
			recv := unparen(comm.Rhs[0]).(*ast.UnaryExpr)
			st = &SelectState{
				Dir:  types.RecvOnly,
				Chan: b.expr(fn, recv.X),
				Pos:  recv.OpPos,
			}
			if debugInfo {
				st.DebugNode = recv
			}

		case *ast.ExprStmt: // <-ch
			recv := unparen(comm.X).(*ast.UnaryExpr)
			st = &SelectState{
				Dir:  types.RecvOnly,
				Chan: b.expr(fn, recv.X),
				Pos:  recv.OpPos,
			}
			if debugInfo {
				st.DebugNode = recv
			}
		}
		states = append(states, st)
	}

	// We dispatch on the (fair) result of Select using a
	// sequential if-else chain, in effect:
	//
	// idx, recvOk, r0...r_n-1 := select(...)
	// if idx == 0 {  // receive on channel 0  (first receive => r0)
	//     x, ok := r0, recvOk
	//     ...state0...
	// } else if v == 1 {   // send on channel 1
	//     ...state1...
	// } else {
	//     ...default...
	// }
	sel := &Select{
		States:   states,
		Blocking: blocking,
	}
	sel.setPos(s.Select)
	var vars []*types.Var
	vars = append(vars, varIndex, varOk)
	for _, st := range states {
		if st.Dir == types.RecvOnly {
			tElem := st.Chan.Type().Underlying().(*types.Chan).Elem()
			vars = append(vars, anonVar(tElem))
		}
	}
	sel.setType(types.NewTuple(vars...))

	fn.emit(sel)
	idx := emitExtract(fn, sel, 0)

	done := fn.newBasicBlock("select.done", selectScope)
	if label != nil {
		label._break = done
	}

	var defaultBody *[]ast.Stmt
	state := 0
	r := 2 // index in 'sel' tuple of value; increments if st.Dir==RECV
	for _, cc := range s.Body.List {
		clause := cc.(*ast.CommClause)
		if clause.Comm == nil {
			defaultBody = &clause.Body
			continue
		}
		body := fn.newBasicBlock("select.body", astScope(fn, s.Body))
		next := fn.newBasicBlock("select.next", nil) //rocky: FIXME get scope
		emitIf(fn, emitCompare(fn, token.EQL, idx, intConst(int64(state)), token.NoPos), body, next)
		fn.currentBlock = body
		fn.targets = &targets{
			tail:   fn.targets,
			_break: done,
		}
		switch comm := clause.Comm.(type) {
		case *ast.ExprStmt: // <-ch
			if debugInfo {
				v := emitExtract(fn, sel, r)
				emitDebugRef(fn, states[state].DebugNode.(ast.Expr), v, false)
			}
			r++

		case *ast.AssignStmt: // x := <-states[state].Chan
			if comm.Tok == token.DEFINE {
				fn.addLocalForIdent(comm.Lhs[0].(*ast.Ident))
			}
			x := b.addr(fn, comm.Lhs[0], false) // non-escaping
			v := emitExtract(fn, sel, r)
			if debugInfo {
				emitDebugRef(fn, states[state].DebugNode.(ast.Expr), v, false)
			}
			x.store(fn, v)

			if len(comm.Lhs) == 2 { // x, ok := ...
				if comm.Tok == token.DEFINE {
					fn.addLocalForIdent(comm.Lhs[1].(*ast.Ident))
				}
				ok := b.addr(fn, comm.Lhs[1], false) // non-escaping
				ok.store(fn, emitExtract(fn, sel, 1))
			}
			r++
		}
		b.stmtList(fn, clause.Body, selectScope)
		fn.targets = fn.targets.tail
		emitJump(fn, done)
		fn.currentBlock = next
		state++
	}
	if defaultBody != nil {
		fn.targets = &targets{
			tail:   fn.targets,
			_break: done,
		}
		b.stmtList(fn, *defaultBody, selectScope)
		fn.targets = fn.targets.tail
	} else {
		// A blocking select must match some case.
		// (This should really be a runtime.errorString, not a string.)
		fn.emit(&Panic{
			X: emitConv(fn, stringConst("blocking select matched no case"), tEface),
		})
		fn.currentBlock = fn.newBasicBlock("unreachable", nil)
	}
	emitJump(fn, done)
	fn.currentBlock = done
}

// forStmt emits to fn code for the for statement s, optionally
// labelled by label.
//
func (b *builder) forStmt(fn *Function, s *ast.ForStmt, label *lblock) {
	//	...init...
	//      jump loop
	// loop:
	//      if cond goto body else done
	// body:
	//      ...body...
	//      jump post
	// post:				 (target of continue)
	//      ...post...
	//      jump loop
	// done:                                 (target of break)
	forScope := astScope(fn, s)
	if s.Init != nil {
		emitTraceStmt(fn, FOR_INIT, s.Init)
		b.stmt(fn, s.Init, forScope)
	}
	body := fn.newBasicBlock("for.body", astScope(fn, s.Body))

	// target of 'break'
	done := fn.newBasicBlock("for.done", ParentScope(fn, forScope))

	loop := body                         // target of back-edge
	if s.Cond != nil {
		loop = fn.newBasicBlock("for.loop", forScope)
	}
	cont := loop // target of 'continue'
	if s.Post != nil {
		cont = fn.newBasicBlock("for.post", forScope)
	}
	if label != nil {
		label._break = done
		label._continue = cont
	}
	emitJump(fn, loop)
	fn.currentBlock = loop
	if loop != body {
		emitTraceExpr(fn, FOR_COND, s.Cond)
		b.cond(fn, s.Cond, body, done)
		fn.currentBlock = body
	}
	fn.targets = &targets{
		tail:      fn.targets,
		_break:    done,
		_continue: cont,
	}
	b.stmt(fn, s.Body, forScope)
	fn.targets = fn.targets.tail
	emitJump(fn, cont)

	if s.Post != nil {
		fn.currentBlock = cont
		emitTraceStmt(fn, FOR_ITER, s.Post)
		b.stmt(fn, s.Post, forScope)
		emitJump(fn, loop) // back-edge
	}
	fn.currentBlock = done
}

// rangeIndexed emits to fn the header for an integer-indexed loop
// over array, *array or slice value x.
// The v result is defined only if tv is non-nil.
//
func (b *builder) rangeIndexed(fn *Function, x Value, tv types.Type, s *ast.RangeStmt)(k, v Value, loop, done *BasicBlock) {
	//
	//      length = len(x)
	//      index = -1
	// loop:                                   (target of continue)
	//      index++
	// 	if index < length goto body else done
	// body:
	//      k = index
	//      v = x[index]
	//      ...body...
	// 	jump loop
	// done:                                   (target of break)

	rangeIndexedScope := astScope(fn, s)
	// Determine number of iterations.
	var length Value
	if arr, ok := deref(x.Type()).Underlying().(*types.Array); ok {
		// For array or *array, the number of iterations is
		// known statically thanks to the type.  We avoid a
		// data dependence upon x, permitting later dead-code
		// elimination if x is pure, static unrolling, etc.
		// Ranging over a nil *array may have >0 iterations.
		// We still generate code for x, in case it has effects.
		length = intConst(arr.Len())
	} else {
		// length = len(x).
		var c Call
		c.Call.Value = makeLen(x.Type())
		c.Call.Args = []Value{x}
		c.setType(tInt)
		length = fn.emit(&c)
	}

	index := fn.addLocal(tInt, s.Pos(), s.End(), rangeIndexedScope)
	emitStore(fn, index, intConst(-1), s.Pos())

	loop = fn.newBasicBlock("rangeindex.loop", rangeIndexedScope)
	emitJump(fn, loop)
	fn.currentBlock = loop

	incr := &BinOp{
		Op: token.ADD,
		X:  emitLoad(fn, index),
		Y:  vOne,
	}
	incr.setType(tInt)
	emitStore(fn, index, fn.emit(incr), s.Pos())

	body := fn.newBasicBlock("rangeindex.body", astScope(fn, s.Body))
	done = fn.newBasicBlock("rangeindex.done", ParentScope(fn, rangeIndexedScope))
	emitIf(fn, emitCompare(fn, token.LSS, incr, length, token.NoPos), body, done)
	fn.currentBlock = body

	k = emitLoad(fn, index)
	if tv != nil {
		switch t := x.Type().Underlying().(type) {
		case *types.Array:
			instr := &Index{
				X:     x,
				Index: k,
			}
			instr.setType(t.Elem())
			v = fn.emit(instr)

		case *types.Pointer: // *array
			instr := &IndexAddr{
				X:     x,
				Index: k,
			}
			instr.setType(types.NewPointer(t.Elem().Underlying().(*types.Array).Elem()))
			v = emitLoad(fn, fn.emit(instr))

		case *types.Slice:
			instr := &IndexAddr{
				X:     x,
				Index: k,
			}
			instr.setType(types.NewPointer(t.Elem()))
			v = emitLoad(fn, fn.emit(instr))

		default:
			panic("rangeIndexed x:" + t.String())
		}
	}
	return
}

// rangeIter emits to fn the header for a loop using
// Range/Next/Extract to iterate over map or string value x.
// tk and tv are the types of the key/value results k and v, or nil
// if the respective component is not wanted.
//
func (b *builder) rangeIter(fn *Function, x Value, tk, tv types.Type, s *ast.RangeStmt) (k, v Value, loop, done *BasicBlock) {
	//
	//	it = range x
	// loop:                                   (target of continue)
	//	okv = next it                      (ok, key, value)
	//  	ok = extract okv #0
	// 	if ok goto body else done
	// body:
	// 	k = extract okv #1
	// 	v = extract okv #2
	//      ...body...
	// 	jump loop
	// done:                                   (target of break)
	//

	rangeIterScope := astScope(fn, s)
	if tk == nil {
		tk = tInvalid
	}
	if tv == nil {
		tv = tInvalid
	}

	rng := &Range{X: x}
	rng.setPos(s.Pos())
	rng.setEnd(s.End())
	rng.setType(tRangeIter)
	it := fn.emit(rng)

	loop = fn.newBasicBlock("rangeiter.loop", rangeIterScope)
	emitJump(fn, loop)
	fn.currentBlock = loop

	_, isString := x.Type().Underlying().(*types.Basic)

	okv := &Next{
		Iter:     it,
		IsString: isString,
	}
	okv.setType(types.NewTuple(
		varOk,
		newVar("k", tk),
		newVar("v", tv),
	))
	fn.emit(okv)

	body := fn.newBasicBlock("rangeiter.body", astScope(fn, s.Body))
	done = fn.newBasicBlock("rangeiter.done", ParentScope(fn, rangeIterScope))
	emitIf(fn, emitExtract(fn, okv, 0), body, done)
	fn.currentBlock = body

	if tk != tInvalid {
		k = emitExtract(fn, okv, 1)
	}
	if tv != tInvalid {
		v = emitExtract(fn, okv, 2)
	}
	return
}

// rangeChan emits to fn the header for a loop that receives from
// channel x until it fails.
// tk is the channel's element type, or nil if the k result is
// not wanted
// pos is the position of the '=' or ':=' token.
//
func (b *builder) rangeChan(fn *Function, x Value, tk types.Type, pos token.Pos) (k Value, loop, done *BasicBlock) {
	//
	// loop:                                   (target of continue)
	//      ko = <-x                           (key, ok)
	//      ok = extract ko #1
	//      if ok goto body else done
	// body:
	//      k = extract ko #0
	//      ...
	//      goto loop
	// done:                                   (target of break)

	loop = fn.newBasicBlock("rangechan.loop", nil)
	emitJump(fn, loop)
	fn.currentBlock = loop
	recv := &UnOp{
		Op:      token.ARROW,
		X:       x,
		CommaOk: true,
	}
	recv.setPos(pos)
	recv.setType(types.NewTuple(
		newVar("k", x.Type().Underlying().(*types.Chan).Elem()),
		varOk,
	))
	ko := fn.emit(recv)
	body := fn.newBasicBlock("rangechan.body", nil)
	done = fn.newBasicBlock("rangechan.done", nil)
	emitIf(fn, emitExtract(fn, ko, 1), body, done)
	fn.currentBlock = body
	if tk != nil {
		k = emitExtract(fn, ko, 0)
	}
	return
}

// rangeStmt emits to fn code for the range statement s, optionally
// labelled by label.
//
func (b *builder) rangeStmt(fn *Function, s *ast.RangeStmt, label *lblock) {
	var tk, tv types.Type
	if s.Key != nil && !isBlankIdent(s.Key) {
		tk = fn.Pkg.typeOf(s.Key)
	}
	if s.Value != nil && !isBlankIdent(s.Value) {
		tv = fn.Pkg.typeOf(s.Value)
	}

	// If iteration variables are defined (:=), this
	// occurs once outside the loop.
	//
	// Unlike a short variable declaration, a RangeStmt
	// using := never redeclares an existing variable; it
	// always creates a new one.
	if s.Tok == token.DEFINE {
		if tk != nil {
			fn.addLocalForIdent(s.Key.(*ast.Ident))
		}
		if tv != nil {
			fn.addLocalForIdent(s.Value.(*ast.Ident))
		}
	}

	x := b.expr(fn, s.X)

	var k, v Value
	var loop, done *BasicBlock
	switch rt := x.Type().Underlying().(type) {
	case *types.Slice, *types.Array, *types.Pointer: // *array
		k, v, loop, done = b.rangeIndexed(fn, x, tv, s)

	case *types.Chan:
		k, loop, done = b.rangeChan(fn, x, tk, s.TokPos)

	case *types.Map, *types.Basic: // string
		k, v, loop, done = b.rangeIter(fn, x, tk, tv, s)

	default:
		panic("Cannot range over: " + rt.String())
	}

	// Evaluate both LHS expressions before we update either.
	var kl, vl lvalue
	if tk != nil {
		kl = b.addr(fn, s.Key, false) // non-escaping
	}
	if tv != nil {
		vl = b.addr(fn, s.Value, false) // non-escaping
	}
	if tk != nil {
		kl.store(fn, k)
	}
	if tv != nil {
		vl.store(fn, v)
	}

	if label != nil {
		label._break = done
		label._continue = loop
	}

	fn.targets = &targets{
		tail:      fn.targets,
		_break:    done,
		_continue: loop,
	}
	b.stmt(fn, s.Body, astScope(fn, s))
	fn.targets = fn.targets.tail
	emitJump(fn, loop) // back-edge
	fn.currentBlock = done
}

// stmt lowers statement s to SSA form, emitting code to fn.
func (b *builder) stmt(fn *Function, _s ast.Stmt, scope *Scope) {
	// The label of the current statement.  If non-nil, its _goto
	// target is always set; its _break and _continue are set only
	// within the body of switch/typeswitch/select/for/range.
	// It is effectively an additional default-nil parameter of stmt().
	var label *lblock
start:
	switch s := _s.(type) {
	case *ast.EmptyStmt:
		// ignore.  (Usually removed by gofmt.)

	case *ast.DeclStmt: // Con, Var or Typ
		d := s.Decl.(*ast.GenDecl)
		if d.Tok == token.VAR {
			for _, spec := range d.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					b.localValueSpec(fn, vs)
				}
			}
		}

	case *ast.LabeledStmt:
		label = fn.labelledBlock(s.Label)
		emitJump(fn, label._goto)
		fn.currentBlock = label._goto
		_s = s.Stmt
		goto start // effectively: tailcall stmt(fn, s.Stmt, label)

	case *ast.ExprStmt:
		b.expr(fn, s.X)

	case *ast.SendStmt:
		fn.emit(&Send{
			Chan: b.expr(fn, s.Chan),
			X: emitConv(fn, b.expr(fn, s.Value),
				fn.Pkg.typeOf(s.Chan).Underlying().(*types.Chan).Elem()),
			pos: s.Arrow,
		})

	case *ast.IncDecStmt:
		op := token.ADD
		if s.Tok == token.DEC {
			op = token.SUB
		}
		loc := b.addr(fn, s.X, false)
		b.assignOp(fn, loc,
			NewConst(exact.MakeInt64(1), loc.typ(), s.X.Pos(), s.X.End()),
			op)

	case *ast.AssignStmt:
		switch s.Tok {
		case token.ASSIGN, token.DEFINE:
			b.assignStmt(fn, s.Lhs, s.Rhs, s.Tok == token.DEFINE, scope)

		default: // +=, etc.
			op := s.Tok + token.ADD - token.ADD_ASSIGN
			b.assignOp(fn, b.addr(fn, s.Lhs[0], false), b.expr(fn, s.Rhs[0]), op)
		}

	case *ast.GoStmt:
		// The "intrinsics" new/make/len/cap are forbidden here.
		// panic is treated like an ordinary function call.
		v := Go{pos: s.Go, endP: s.End()}
		b.setCall(fn, s.Call, &v.Call)
		fn.emit(&v)

	case *ast.DeferStmt:
		// The "intrinsics" new/make/len/cap are forbidden here.
		// panic is treated like an ordinary function call.
		v := Defer{pos: s.Defer, endP: s.End()}
		b.setCall(fn, s.Call, &v.Call)
		fn.emit(&v)

		// A deferred call can cause recovery from panic,
		// and control resumes at the Recover block.
		createRecoverBlock(fn)

	case *ast.ReturnStmt:
		var results []Value
		if len(s.Results) == 1 && fn.Signature.Results().Len() > 1 {
			// Return of one expression in a multi-valued function.
			tuple := b.exprN(fn, s.Results[0])
			ttuple := tuple.Type().(*types.Tuple)
			for i, n := 0, ttuple.Len(); i < n; i++ {
				results = append(results,
					emitConv(fn, emitExtract(fn, tuple, i),
						fn.Signature.Results().At(i).Type()))
			}
		} else {
			// 1:1 return, or no-arg return in non-void function.
			for i, r := range s.Results {
				v := emitConv(fn, b.expr(fn, r), fn.Signature.Results().At(i).Type())
				results = append(results, v)
			}
		}
		if fn.namedResults != nil {
			// Function has named result parameters (NRPs).
			// Perform parallel assignment of return operands to NRPs.
			for i, r := range results {
				emitStore(fn, fn.namedResults[i], r, s.Return)
			}
		}
		// Run function calls deferred in this
		// function when explicitly returning from it.
		fn.emit(new(RunDefers))
		if fn.namedResults != nil {
			// Reload NRPs to form the result tuple.
			results = results[:0]
			for _, r := range fn.namedResults {
				results = append(results, emitLoad(fn, r))
			}
		}
		fn.emit(&Return{
			Results: results,
			pos: s.Pos(),
			endP: s.End(),
		})
		fn.currentBlock = fn.newBasicBlock("unreachable", nil)

	case *ast.BranchStmt:
		var block *BasicBlock
		switch s.Tok {
		case token.BREAK:
			emitTraceStmt(fn, BREAK_STMT, s)
			if s.Label != nil {
				block = fn.labelledBlock(s.Label)._break
			} else {
				for t := fn.targets; t != nil && block == nil; t = t.tail {
					block = t._break
				}
			}

		case token.CONTINUE:
			if s.Label != nil {
				block = fn.labelledBlock(s.Label)._continue
			} else {
				for t := fn.targets; t != nil && block == nil; t = t.tail {
					block = t._continue
				}
			}

		case token.FALLTHROUGH:
			for t := fn.targets; t != nil && block == nil; t = t.tail {
				block = t._fallthrough
			}

		case token.GOTO:
			block = fn.labelledBlock(s.Label)._goto
		}
		emitJump(fn, block)
		fn.currentBlock = fn.newBasicBlock("unreachable", nil)

	case *ast.BlockStmt:
		if scope == nil {
			scope = astScope(fn, s)
		}
		b.stmtList(fn, s.List, scope)
		emitTrace(fn, BLOCK_END, s.End(), s.End())

	case *ast.IfStmt:
		ifScope := astScope(fn, s)
		if s.Init != nil {
			emitTraceStmt(fn, IF_INIT, s.Init)
			b.stmt(fn, s.Init, ifScope)
		}
		then := fn.newBasicBlock("if.then", astScope(fn, s.Body))
		done := fn.newBasicBlock("if.done", ParentScope(fn, ifScope))
		els := done
		if s.Else != nil {
			els = fn.newBasicBlock("if.else", astScope(fn, s.Else))
		}
		emitTraceExpr(fn, IF_COND, s.Cond)
		b.cond(fn, s.Cond, then, els)
		fn.currentBlock = then
		b.stmt(fn, s.Body, ifScope)
		emitJump(fn, done)

		if s.Else != nil {
			fn.currentBlock = els
			b.stmt(fn, s.Else, astScope(fn, s.Else))
			emitJump(fn, done)
		}

		fn.currentBlock = done

	case *ast.SwitchStmt:
		b.switchStmt(fn, s, label)

	case *ast.TypeSwitchStmt:
		b.typeSwitchStmt(fn, s, label)

	case *ast.SelectStmt:
		b.selectStmt(fn, s, label)

	case *ast.ForStmt:
		b.forStmt(fn, s, label)

	case *ast.RangeStmt:
		b.rangeStmt(fn, s, label)

	default:
		panic(fmt.Sprintf("unexpected statement kind: %T", s))
	}
}

// buildFunction builds SSA code for the body of function fn.  Idempotent.
func (b *builder) buildFunction(fn *Function) {
	if fn.Blocks != nil {
		return // building already started
	}

	var recvField *ast.FieldList
	var body *ast.BlockStmt
	var functype *ast.FuncType
	switch n := fn.syntax.(type) {
	case nil:
		return // not a Go source function.  (Synthetic, or from object file.)
	case *ast.FuncDecl:
		functype = n.Type
		recvField = n.Recv
		body = n.Body
	case *ast.FuncLit:
		functype = n.Type
		body = n.Body
	default:
		panic(n)
	}

	if body == nil {
		// External function.
		if fn.Params == nil {
			// This condition ensures we add a non-empty
			// params list once only, but we may attempt
			// the degenerate empty case repeatedly.
			// TODO(adonovan): opt: don't do that.

			// We set Function.Params even though there is no body
			// code to reference them.  This simplifies clients.
			if recv := fn.Signature.Recv(); recv != nil {
				fn.addParamObj(recv)
			}
			params := fn.Signature.Params()
			for i, n := 0, params.Len(); i < n; i++ {
				fn.addParamObj(params.At(i))
			}
		}
		return
	}
	scope := fn.Scope
	pkg := fn.Pkg
	if fn.syntax != nil && scope != nil {
		scope.node = &fn.syntax
	}
	pkg.locs = append(pkg.locs, LocInst{pos: fn.pos, endP: fn.endP,
		Trace: nil, Fn: fn})
	if fn.Prog.mode&LogSource != 0 {
		defer logStack("build function %s @ %s", fn, fn.Prog.Fset.Position(fn.pos))()
	}
	fn.startBody(scope) // Or nil?
	fn.createSyntacticParams(recvField, functype)
	b.stmt(fn, body, scope)
	if cb := fn.currentBlock; cb != nil && (cb == fn.Blocks[0] || cb == fn.Recover || cb.Preds != nil) {
		// Control fell off the end of the function's body block.
		//
		// Block optimizations eliminate the current block, if
		// unreachable.  It is a builder invariant that
		// if this no-arg return is ill-typed for
		// fn.Signature.Results, this block must be
		// unreachable.  The sanity checker checks this.
		fn.emit(new(RunDefers))
		fn.emit(new(Return))
	}
	fn.finishBody()
}

// buildFuncDecl builds SSA code for the function or method declared
// by decl in package pkg.
//
func (b *builder) buildFuncDecl(pkg *Package, decl *ast.FuncDecl) {
	id := decl.Name
	if isBlankIdent(id) {
		return // discard
	}
	var fn *Function
	if decl.Recv == nil && id.Name == "init" {
		/* ROCKY FIXME: location info isn't getting in.
		fmt.Printf("+++1 %s init#%d %d\n", pkg, pkg.ninit, decl.Name.NamePos)*
        */
		pkg.ninit++
		fn = &Function{
			name:      fmt.Sprintf("init#%d", pkg.ninit),
			Signature: new(types.Signature),
			pos:       decl.Name.NamePos,
			endP:      decl.Body.Rbrace,
			Pkg:       pkg,
			Prog:      pkg.Prog,
			LocalsByName: make(map[NameScope]uint),
			syntax:    decl,
		}
		fn.Scope = astScope(fn, decl)

		var v Call
		v.Call.Value = fn
		v.setType(types.NewTuple())
		pkg.init.emit(&v)
	} else {
		fn = pkg.values[pkg.info.Defs[id]].(*Function)
	}
	b.buildFunction(fn)
}

// BuildAll calls Package.Build() for each package in prog.
// Building occurs in parallel unless the BuildSerially mode flag was set.
//
// BuildAll is idempotent and thread-safe.
//
func (prog *Program) BuildAll() {
	var wg sync.WaitGroup
	for _, p := range prog.packages {
		if prog.mode&BuildSerially != 0 {
			p.Build()
		} else {
			wg.Add(1)
			go func(p *Package) {
				p.Build()
				wg.Done()
			}(p)
		}
	}
	wg.Wait()
}

// Build builds SSA code for all functions and vars in package p.
//
// Precondition: CreatePackage must have been called for all of p's
// direct imports (and hence its direct imports must have been
// error-free).
//
// Build is idempotent and thread-safe.
//
func (p *Package) Build() {
	if !atomic.CompareAndSwapInt32(&p.started, 0, 1) {
		return // already started
	}
	if p.info == nil {
		return // synthetic package, e.g. "testmain"
	}
	if len(p.info.Files) == 0 {
		p.info = nil
		return // package loaded from export data
	}

	// Ensure we have runtime type info for all exported members.
	// TODO(adonovan): ideally belongs in memberFromObject, but
	// that would require package creation in topological order.
	for name, mem := range p.Members {
		if ast.IsExported(name) {
			p.needMethodsOf(mem.Type())
		}
	}
	if p.Prog.mode&LogSource != 0 {
		defer logStack("build %s", p)()
	}
	init := p.init
	scope := init.Scope
	init.startBody(scope)

	var done *BasicBlock

	if p.Prog.mode&BareInits == 0 {
		// Make init() skip if package is already initialized.
		initguard := p.Var("init$guard")
		doinit := init.newBasicBlock("init.start", scope)
		done = init.newBasicBlock("init.done", scope)
		emitIf(init, emitLoad(init, initguard), done, doinit)
		init.currentBlock = doinit
		emitStore(init, initguard, vTrue, token.NoPos)

		// Call the init() function of each package we import.
		for _, pkg := range p.info.Pkg.Imports() {
			prereq := p.Prog.packages[pkg]
			if prereq == nil {
				panic(fmt.Sprintf("Package(%q).Build(): unsatisfied import: Program.CreatePackage(%q) was not called", p.Object.Path(), pkg.Path()))
			}
			var v Call
			v.Call.Value = prereq.init
			v.Call.pos = init.pos
			v.setType(types.NewTuple())
			init.emit(&v)
		}
	}

	var b builder

	// Initialize package-level vars in correct order.
	for _, varinit := range p.info.InitOrder {
		if init.Prog.mode&LogSource != 0 {
			fmt.Fprintf(os.Stderr, "build global initializer %v @ %s\n",
				varinit.Lhs, p.Prog.Fset.Position(varinit.Rhs.Pos()))
		}
		if len(varinit.Lhs) == 1 {
			// 1:1 initialization: var x, y = a(), b()
			var lval lvalue
			if v := varinit.Lhs[0]; v.Name() != "_" {
				lval = &address{addr: p.values[v].(*Global), pos: v.Pos()}
			} else {
				lval = blank{}
			}
			b.exprInPlace(init, lval, varinit.Rhs, true)
		} else {
			// n:1 initialization: var x, y :=  f()
			tuple := b.exprN(init, varinit.Rhs)
			for i, v := range varinit.Lhs {
				if v.Name() == "_" {
					continue
				}
				emitStore(init, p.values[v].(*Global), emitExtract(init, tuple, i), v.Pos())
			}
		}
	}

	// Build all package-level functions, init functions
	// and methods, including unreachable/blank ones.
	// We build them in source order, but it's not significant.
	for _, file := range p.info.Files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				b.buildFuncDecl(p, decl)
			}
		}
	}

	// Finish up init().
	if p.Prog.mode&BareInits == 0 {
		emitJump(init, done)
		init.currentBlock = done
	}
	init.emit(new(Return))
	init.finishBody()

	p.info = nil // We no longer need ASTs or go/types deductions.

	if p.Prog.mode&SanityCheckFunctions != 0 {
		sanityCheckPackage(p)
	}
}

// Like ObjectOf, but panics instead of returning nil.
// Only valid during p's create and build phases.
func (p *Package) objectOf(id *ast.Ident) types.Object {
	if o := p.info.ObjectOf(id); o != nil {
		return o
	}
	panic(fmt.Sprintf("no types.Object for ast.Ident %s @ %s",
		id.Name, p.Prog.Fset.Position(id.Pos())))
}

// Like TypeOf, but panics instead of returning nil.
// Only valid during p's create and build phases.
func (p *Package) typeOf(e ast.Expr) types.Type {
	if T := p.info.TypeOf(e); T != nil {
		return T
	}
	panic(fmt.Sprintf("no type for %T @ %s",
		e, p.Prog.Fset.Position(e.Pos())))
}

// needMethodsOf ensures that runtime type information (including the
// complete method set) is available for the specified type T and all
// its subcomponents.
//
// needMethodsOf must be called for at least every type that is an
// operand of some MakeInterface instruction, and for the type of
// every exported package member.
//
// Precondition: T is not a method signature (*Signature with Recv()!=nil).
//
// Thread-safe.  (Called via emitConv from multiple builder goroutines.)
//
// TODO(adonovan): make this faster.  It accounts for 20% of SSA build
// time.  Do we need to maintain a distinct needRTTI and methodSets per
// package?  Using just one in the program might be much faster.
//
func (p *Package) needMethodsOf(T types.Type) {
	p.methodsMu.Lock()
	p.needMethods(T, false)
	p.methodsMu.Unlock()
}

// Precondition: T is not a method signature (*Signature with Recv()!=nil).
// Precondition: the p.methodsMu lock is held.
// Recursive case: skip => don't call makeMethods(T).
func (p *Package) needMethods(T types.Type, skip bool) {
	// Each package maintains its own set of types it has visited.
	if prevSkip, ok := p.needRTTI.At(T).(bool); ok {
		// needMethods(T) was previously called
		if !prevSkip || skip {
			return // already seen, with same or false 'skip' value
		}
	}
	p.needRTTI.Set(T, skip)

	// Prune the recursion if we find a named or *named type
	// belonging to another package.
	var n *types.Named
	switch T := T.(type) {
	case *types.Named:
		n = T
	case *types.Pointer:
		n, _ = T.Elem().(*types.Named)
	}
	if n != nil {
		owner := n.Obj().Pkg()
		if owner == nil {
			return // built-in error type
		}
		if owner != p.Object {
			return // belongs to another package
		}
	}

	// All the actual method sets live in the Program so that
	// multiple packages can share a single copy in memory of the
	// symbols that would be compiled into multiple packages (as
	// weak symbols).
	if !skip && p.Prog.makeMethods(T) {
		p.methodSets = append(p.methodSets, T)
	}

	// Recursion over signatures of each method.
	tmset := p.Prog.MethodSets.MethodSet(T)
	for i := 0; i < tmset.Len(); i++ {
		sig := tmset.At(i).Type().(*types.Signature)
		p.needMethods(sig.Params(), false)
		p.needMethods(sig.Results(), false)
	}

	switch t := T.(type) {
	case *types.Basic:
		// nop

	case *types.Interface:
		// nop---handled by recursion over method set.

	case *types.Pointer:
		p.needMethods(t.Elem(), false)

	case *types.Slice:
		p.needMethods(t.Elem(), false)

	case *types.Chan:
		p.needMethods(t.Elem(), false)

	case *types.Map:
		p.needMethods(t.Key(), false)
		p.needMethods(t.Elem(), false)

	case *types.Signature:
		if t.Recv() != nil {
			panic(fmt.Sprintf("Signature %s has Recv %s", t, t.Recv()))
		}
		p.needMethods(t.Params(), false)
		p.needMethods(t.Results(), false)

	case *types.Named:
		// A pointer-to-named type can be derived from a named
		// type via reflection.  It may have methods too.
		p.needMethods(types.NewPointer(T), false)

		// Consider 'type T struct{S}' where S has methods.
		// Reflection provides no way to get from T to struct{S},
		// only to S, so the method set of struct{S} is unwanted,
		// so set 'skip' flag during recursion.
		p.needMethods(t.Underlying(), true)

	case *types.Array:
		p.needMethods(t.Elem(), false)

	case *types.Struct:
		for i, n := 0, t.NumFields(); i < n; i++ {
			p.needMethods(t.Field(i).Type(), false)
		}

	case *types.Tuple:
		for i, n := 0, t.Len(); i < n; i++ {
			p.needMethods(t.At(i).Type(), false)
		}

	default:
		panic(T)
	}
}
