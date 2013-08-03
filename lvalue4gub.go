package ssa2

func (a address) storeWithScope(fn *Function, v Value, scope *Scope) {
	store := emitStore(fn, a.addr, v)
	store.Scope = scope
	store.pos = a.starPos
	if a.id != nil {
		// store.Val is v converted for assignability.
		emitDebugRef(fn, a.id, store.Val)
	}
}

func (bl blank) storeWithScope(fn *Function, v Value, scope *Scope) {
	// no-op
}

func (e *element) storeWithScope(fn *Function, v Value, scope *Scope) {
	// ignore scope
	e.store(fn, v)
}
