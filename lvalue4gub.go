// Copyright 2014 Rocky Bernstein
package ssa2
import (
	"go/token"
)

func (a *address) storeWithScope(fn *Function, v Value, scope *Scope) {
	store := emitStore(fn, a.addr, v, token.NoPos)
	/* FIXME rb: store.Scope = scope */
	if a.expr != nil {
		// store.Val is v converted for assignability.
		emitDebugRef(fn, a.expr, store.Val, true)
	}
}

func (bl blank) storeWithScope(fn *Function, v Value, scope *Scope) {
	// no-op
}

func (e *element) storeWithScope(fn *Function, v Value, scope *Scope) {
	// ignore scope
	e.store(fn, v)
}
