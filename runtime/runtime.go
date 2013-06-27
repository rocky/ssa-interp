package SSAruntime

import (
	"ssa-interp"
)


// Mode is a bitmask of options influencing the interpreter.
type Mode uint

// Mode is a bitmask of options influencing the tracing.
type TraceMode uint

const (
	// Disable recover() in target programs; show interpreter crash instead.
	DisableRecover Mode = 1 << iota
)

const (
	// Print a trace of all instructions as they are interpreted.
	EnableTracing  TraceMode = 1 << iota

	// Print higher-level statement boundary tracing
	EnableStmtTracing
)

// State shared between all interpreted goroutines.
type Interpreter struct {
	Prog           *ssa2.Program         // the SSA program
	Globals        map[ssa2.Value]*value // addresses of global variables (immutable)
	Mode           Mode                  // interpreter options
	TraceMode      TraceMode             // interpreter trace options
	reflectPackage *ssa2.Package         // the fake reflect package
	errorMethods   ssa2.MethodSet        // the method set of reflect.error, which implements the error interface.
	rtypeMethods   ssa2.MethodSet        // the method set of rtype, which implements the reflect.Type interface.
}

type frame struct {
	I                *Interpreter
	Caller           *frame
	Fn               *ssa2.Function
	Block, PrevBlock *ssa2.BasicBlock
	Env              map[ssa2.Value]value // dynamic values of SSA variables
	Locals           []value
	defers           []func()
	result           value
	Status           Status
	panic            interface{}
}

type Status int

const (
	StRunning Status = iota
	StComplete
	StPanic
)


type hashable interface {
	hash() int
	eq(x interface{}) bool
}

type entry struct {
	key   hashable
	value value
	next  *entry
}

// A hashtable atop the built-in map.  Since each bucket contains
// exactly one hash value, there's no need to perform hash-equality
// tests when walking the linked list.  Rehashing is done by the
// underlying map.
type hashmap struct {
	table  map[int]*entry
	length int // number of entries in map
}

// makeMap returns an empty initialized map of key type kt,
