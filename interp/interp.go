// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ssa-interp/interp defines an interpreter for the SSA
// representation of Go programs.
//
// This interpreter is provided as an adjunct for testing the SSA
// construction algorithm.  Its purpose is to provide a minimal
// metacircular implementation of the dynamic semantics of each SSA
// instruction.  It is not, and will never be, a production-quality Go
// interpreter.
//
// The following is a partial list of Go features that are currently
// unsupported or incomplete in the interpreter.
//
// * Unsafe operations, including all uses of unsafe.Pointer, are
// impossible to support given the "boxed" value representation we
// have chosen.
//
// * The reflect package is only partially implemented.
//
// * "sync/atomic" operations are not currently atomic due to the
// "boxed" value representation: it is not possible to read, modify
// and write an interface value atomically.  As a consequence, Mutexes
// are currently broken.  TODO(adonovan): provide a metacircular
// implementation of Mutex avoiding the broken atomic primitives.
//
// * recover is only partially implemented.  Also, the interpreter
// makes no attempt to distinguish target panics from interpreter
// crashes.
//
// * map iteration is asymptotically inefficient.
//
// * the sizes of the int, uint and uintptr types in the target
// program are assumed to be the same as those of the interpreter
// itself.
//
// * all values occupy space, even those of types defined by the spec
// to have zero size, e.g. struct{}.  This can cause asymptotic
// performance degradation.
//
// * os.Exit is implemented using panic, causing deferred functions to
// run.
package interp

import (
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"

	"code.google.com/p/go.tools/go/types"
	"github.com/rocky/ssa-interp"
)

type continuation int

const (
	kNext continuation = iota
	kReturn
	kJump
)

// Mode is a bitmask of options influencing the interpreter.
type Mode uint

const (
	// Disable recover() in target programs; show interpreter crash instead.
	DisableRecover Mode = 1 << iota
)

type methodSet map[string]*ssa2.Function

// State shared between all interpreted goroutines.
type interpreter struct {
	prog           *ssa2.Program            // the SSA program
	globals        map[ssa2.Value]*Value    // addresses of global variables (immutable)
	Mode           Mode                     // interpreter options
	reflectPackage *ssa2.Package            // the fake reflect package
	errorMethods   methodSet                // the method set of reflect.error, which implements the error interface.
	rtypeMethods   methodSet                // the method set of rtype, which implements the reflect.Type interface.
	runtimeErrorString types.Type           // the runtime.errorString type

	TraceMode      TraceMode                // interpreter trace options
	TraceEventMask ssa2.TraceEventMask
	nGoroutines    int                      // number of goroutines
	goTops         []*GoreState
}

// lookupMethod returns the method set for type typ, which may be one
// of the interpreter's fake types.
func lookupMethod(i *interpreter, typ types.Type, meth *types.Func) *ssa2.Function {
	switch typ {
	case rtypeType:
		return i.rtypeMethods[meth.Id()]
	case errorType:
		return i.errorMethods[meth.Id()]
	}
	return i.prog.Method(typ.MethodSet().Lookup(meth.Pkg(), meth.Name()))
}

// visitInstr interprets a single ssa2.Instruction within the activation
// record frame.  It returns a continuation value indicating where to
// read the next instruction from.
func visitInstr(fr *Frame, genericInstr ssa2.Instruction) continuation {
	// switch instr := genericInstr.(type) {
	// default:
	// 	fmt.Printf("instruction: %T\n", instr)
	// }

	switch instr := genericInstr.(type) {
	case *ssa2.DebugRef:
		if instr.Object != nil {
			regName := instr.X.Name()
			varName := instr.Object.Name()
			if regName != varName && regName[0] == 't' {
				// fmt.Printf("+++%s is register\t%s\n", varName, regName)
				fr.Var2Reg[varName] = regName
				fr.Reg2Var[regName] = varName
			}
		}
	case *ssa2.UnOp:
		fr.env[instr] = unop(instr, fr.get(instr.X))

	case *ssa2.BinOp:
		fr.env[instr] = binop(instr.Op, instr.X.Type(), fr.get(instr.X), fr.get(instr.Y))

	case *ssa2.Call:
		fn, args := prepareCall(fr, &instr.Call)
		fr.env[instr] = call(fr.i, fr.goNum, fr, fn, args)

	case *ssa2.ChangeInterface:
		fr.env[instr] = fr.get(instr.X)

	case *ssa2.ChangeType:
		fr.env[instr] = fr.get(instr.X) // (can't fail)

	case *ssa2.Convert:
		fr.env[instr] = conv(instr.Type(), instr.X.Type(), fr.get(instr.X))

	case *ssa2.MakeInterface:
		fr.env[instr] = iface{t: instr.X.Type(), v: fr.get(instr.X)}

	case *ssa2.Extract:
		fr.env[instr] = fr.get(instr.Tuple).(tuple)[instr.Index]

	case *ssa2.Slice:
		fr.env[instr] = slice(fr.get(instr.X), fr.get(instr.Low), fr.get(instr.High))

	case *ssa2.Return:
		switch len(instr.Results) {
		case 0:
		case 1:
			fr.result = fr.get(instr.Results[0])
		default:
			var res []Value
			for _, r := range instr.Results {
				res = append(res, fr.get(r))
			}
			fr.result = tuple(res)
		}
		fr.block = nil
		return kReturn

	case *ssa2.RunDefers:
		fr.runDefers()

	case *ssa2.Panic:
		switch os.Getenv("GOTRACEBACK") {
		case "0":
			//do nothing
		case "1":
			runtime۰Gotraceback(fr)
		case "2", "crash":
			runtime۰Gotraceback(fr)
			for _, goTop := range fr.i.goTops {
				otherFr := goTop.Fr
				if otherFr == fr { continue }
				runtime۰Gotraceback(otherFr)
			}
		}
		TraceHook(fr, &genericInstr, ssa2.PANIC)
		// Don't know if setting fr.status really does anything, but
		// just to try to be totally Kosher. We do this *after*
		// running TraceHook because TraceHook treats panic'd frames
		// differently and will do less with them. If it needs to
		// understand that we are in a panic state, it can do that via
		// the event type passed above.
		fr.status = StPanic
		// We don't need an interpreter tracecback. So turn that off.
		os.Setenv("GOTRACEBACK", "0")
		panic(targetPanic{fr.get(instr.X)})

	case *ssa2.Send:
		fr.get(instr.Chan).(chan Value) <- copyVal(fr.get(instr.X))

	case *ssa2.Store:
		*fr.get(instr.Addr).(*Value) = copyVal(fr.get(instr.Val))

	case *ssa2.If:
		succ := 1
		if fr.get(instr.Cond).(bool) {
			succ = 0
		}
		fr.prevBlock, fr.block = fr.block, fr.block.Succs[succ]
		return kJump

	case *ssa2.Jump:
		fr.prevBlock, fr.block = fr.block, fr.block.Succs[0]
		return kJump

	case *ssa2.Defer:
		fn, args := prepareCall(fr, &instr.Call)
		fr.defers = append(fr.defers, func() { call(fr.i, fr.goNum, fr, fn, args) })
		// fr.defers = &deferred{
		// 	fn:    fn,
		// 	args:  args,
		// 	instr: instr,
		// 	tail:  fr.defers,
		// }

	case *ssa2.Go:
		fn, args := prepareCall(fr, &instr.Call)
		go call(fr.i, i.nGoroutines, nil, fn, args)

	case *ssa2.MakeChan:
		fr.env[instr] = make(chan Value, asInt(fr.get(instr.Size)))

	case *ssa2.Alloc:
		var addr *Value
		if instr.Heap {
			// new
			addr = new(Value)
			fr.env[instr] = addr
		} else {
			// local
			addr = fr.env[instr].(*Value)
		}
		*addr = zero(deref(instr.Type()))

	case *ssa2.MakeSlice:
		slice := make([]Value, asInt(fr.get(instr.Cap)))
		tElt := instr.Type().Underlying().(*types.Slice).Elem()
		for i := range slice {
			slice[i] = zero(tElt)
		}
		fr.env[instr] = slice[:asInt(fr.get(instr.Len))]

	case *ssa2.MakeMap:
		reserve := 0
		if instr.Reserve != nil {
			reserve = asInt(fr.get(instr.Reserve))
		}
		fr.env[instr] = makeMap(instr.Type().Underlying().(*types.Map).Key(), reserve)

	case *ssa2.Range:
		fr.env[instr] = rangeIter(fr.get(instr.X), instr.X.Type())

	case *ssa2.Next:
		fr.env[instr] = fr.get(instr.Iter).(iter).next()

	case *ssa2.FieldAddr:
		x := fr.get(instr.X)
		fr.env[instr] = &(*x.(*Value)).(structure)[instr.Field]

	case *ssa2.Field:
		fr.env[instr] = copyVal(fr.get(instr.X).(structure)[instr.Field])

	case *ssa2.IndexAddr:
		x := fr.get(instr.X)
		idx := fr.get(instr.Index)
		switch x := x.(type) {
		case []Value:
			fr.env[instr] = &x[asInt(idx)]
		case *Value: // *array
			fr.env[instr] = &(*x).(array)[asInt(idx)]
		default:
			panic(fmt.Sprintf("unexpected x type in IndexAddr: %T", x))
		}

	case *ssa2.Index:
		fr.env[instr] = copyVal(fr.get(instr.X).(array)[asInt(fr.get(instr.Index))])

	case *ssa2.Lookup:
		fr.env[instr] = lookup(instr, fr.get(instr.X), fr.get(instr.Index))

	case *ssa2.MapUpdate:
		m := fr.get(instr.Map)
		key := fr.get(instr.Key)
		v := fr.get(instr.Value)
		switch m := m.(type) {
		case map[Value]Value:
			m[key] = v
		case *hashmap:
			m.insert(key.(hashable), v)
		default:
			panic(fmt.Sprintf("illegal map type: %T", m))
		}

	case *ssa2.TypeAssert:
		fr.env[instr] = typeAssert(fr.i, instr, fr.get(instr.X).(iface))

	case *ssa2.Trace:
		fr.startP = instr.Start
		fr.endP   = instr.End
		if (fr.tracing == TRACE_STEP_IN) ||
			instr.Breakpoint ||
			(fr.tracing == TRACE_STEP_OVER) && GlobalStmtTracing() {
			TraceHook(fr, &genericInstr, instr.Event)
		}

	case *ssa2.MakeClosure:
		var bindings []Value
		for _, binding := range instr.Bindings {
			bindings = append(bindings, fr.get(binding))
		}
		fr.env[instr] = &closure{instr.Fn.(*ssa2.Function), bindings}

	case *ssa2.Phi:
		for i, pred := range instr.Block().Preds {
			if fr.prevBlock == pred {
				fr.env[instr] = fr.get(instr.Edges[i])
				break
			}
		}

	case *ssa2.Select:
		var cases []reflect.SelectCase
		if !instr.Blocking {
			cases = append(cases, reflect.SelectCase{
				Dir: reflect.SelectDefault,
			})
		}
		for _, state := range instr.States {
			var dir reflect.SelectDir
			if state.Dir == ast.RECV {
				dir = reflect.SelectRecv
			} else {
				dir = reflect.SelectSend
			}
			var send reflect.Value
			if state.Send != nil {
				send = reflect.ValueOf(fr.get(state.Send))
			}
			cases = append(cases, reflect.SelectCase{
				Dir:  dir,
				Chan: reflect.ValueOf(fr.get(state.Chan)),
				Send: send,
			})
		}
		chosen, recv, recvOk := reflect.Select(cases)
		if !instr.Blocking {
			chosen-- // default case should have index -1.
		}
		r := tuple{chosen, recvOk}
		for i, st := range instr.States {
			if st.Dir == ast.RECV {
				var v Value
				if i == chosen && recvOk {
					// No need to copy since send makes an unaliased copy.
					v = recv.Interface().(Value)
				} else {
					v = zero(st.Chan.Type().Underlying().(*types.Chan).Elem())
				}
				r = append(r, v)
			}
		}
		fr.env[instr] = r

	default:
		panic(fmt.Sprintf("unexpected instruction: %T", instr))
	}

	// if val, ok := instr.(ssa2.Value); ok {
	// 	fmt.Println(toString(fr.env[val])) // debugging
	// }

	return kNext
}

// prepareCall determines the function value and argument values for a
// function call in a Call, Go or Defer instruction, performing
// interface method lookup if needed.
//
func prepareCall(fr *Frame, call *ssa2.CallCommon) (fn Value, args []Value) {
	v := fr.get(call.Value)
	if call.Method == nil {
		// Function call.
		fn = v
	} else {
		// Interface method invocation.
		recv := v.(iface)
		if recv.t == nil {
			panic("method invoked on nil interface")
		}
		if f := lookupMethod(fr.i, recv.t, call.Method); f == nil {
			// Unreachable in well-typed programs.
			panic(fmt.Sprintf("method set for dynamic type %v does not contain %s", recv.t, call.Method))
		} else {
			fn = f
		}
		args = append(args, copyVal(recv.v))
	}
	for _, arg := range call.Args {
		args = append(args, fr.get(arg))
	}
	return
}

// call interprets a call to a function (function, builtin or closure)
// fn with arguments args, returning its result.
// callpos is the position of the callsite.
//
func call(i *interpreter, goNum int, caller *Frame,	fn Value, args []Value) Value {
	switch fn := fn.(type) {
	case *ssa2.Function:
		if fn == nil {
			panic("call of nil function") // nil of func type
		}
		return callSSA(i, goNum, caller, fn, args, nil)
	case *closure:
		return callSSA(i, goNum, caller, fn.fn, args, fn.env)
	case *ssa2.Builtin:
		return callBuiltin(caller, fn, args)
	}
	panic(fmt.Sprintf("cannot call %T", fn))
}

// callSSA interprets a call to function fn with arguments args,
// and lexical environment env, returning its result.
// callpos is the position of the callsite.
//
func callSSA(i *interpreter, goNum int, caller *Frame, fn *ssa2.Function, args []Value, env []Value) Value {
	if InstTracing() {
		fset := fn.Prog.Fset
		// TODO(adonovan): fix: loc() lies for external functions.
		loc := ssa2.FmtRangeWithFset(fset, caller.startP, caller.endP)
		fmt.Fprintf(os.Stderr, "Entering %s at %s.\n", fn, loc)
		suffix := ""
		if caller != nil {
			suffix = fmt.Sprintf(", resuming %s at %s", caller.fn.String(),
				loc)
		}
		defer fmt.Fprintf(os.Stderr, "Leaving %s%s.\n", fn, suffix)
	}
	if fn.Enclosing == nil {
		name := fn.String()
		if ext := externals[name]; ext != nil {
			if InstTracing() {
				fmt.Fprintln(os.Stderr, "\t(external)")
			}
			return ext(caller, args)
		}
		if fn.Blocks == nil {
			panic("no code for function: " + name)
		}
	}
	fr := &Frame{
		i:       i,
		caller  : caller,
		fn      : fn,
		env     : make(map[ssa2.Value]Value),
		block   : fn.Blocks[0],
		locals  : make([]Value, len(fn.Locals)),
		tracing : TRACE_STEP_NONE,
		goNum   : goNum,
		Var2Reg : make(map[string]string),
		Reg2Var : make(map[string]string),
	}
	i.goTops[goNum].Fr = fr

	for i, l := range fn.Locals {
		fr.locals[i] = zero(deref(l.Type()))
		fr.env[l] = &fr.locals[i]
	}
	for i, p := range fn.Params {
		fr.env[p] = args[i]
	}
	for i, fv := range fn.FreeVars {
		fr.env[fv] = env[i]
	}

	if caller == nil {
		if GlobalStmtTracing() {
			fr.tracing = TRACE_STEP_IN
		}
	} else if caller.tracing == TRACE_STEP_IN {
		fr.tracing = TRACE_STEP_IN
	}

	for fr.block != nil {
		runFrame(fr)
	}
	// Destroy the locals to avoid accidental use after return.
	for i := range fn.Locals {
		fr.locals[i] = bad{}
	}
	return fr.result
}

// runFrame executes SSA instructions starting at fr.block and
// continuing until a return, a panic, or a recovered panic.
//
// After a panic, runFrame panics.
//
// After a normal return, fr.result contains the result of the call
// and fr.block is nil.
//
// After a recovered panic, fr.result is undefined and fr.block
// contains the block at which to resume control, which may be
// nil for a normal return.
//
func runFrame(fr *Frame) {
	defer func() {
		if fr.block == nil {
			return // normal return
		}
		if fr.i.Mode&DisableRecover != 0 {
			return // let interpreter crash
		}
		fr.panicking = true
		fr.panic = recover()
		if InstTracing() || GlobalStmtTracing() {
			fmt.Fprintf(os.Stderr, "Panicking: %T %v.\n", fr.panic, fr.panic)
			debug.PrintStack()
		}
		fr.runDefers()
		fr.block = fr.fn.Recover // recovered panic
	}()

	fn        := fr.fn
	fr.startP = fn.Pos()
	fr.endP   = fn.Pos()
	if ((fr.tracing == TRACE_STEP_IN) &&
		(len(fr.block.Instrs) > 0 && GlobalStmtTracing()) ||
		fn.Breakpoint ) {
		event := ssa2.CALL_ENTER
		if fn.Breakpoint { event = ssa2.BREAKPOINT }
		TraceHook(fr, &fr.block.Instrs[0], event)
	}
	for {
		var instr ssa2.Instruction
		if InstTracing() {
			fmt.Fprintf(os.Stderr, ".%s:\n", fr.block)
		}
	block:
		// rocky: changed to allow for debugger "jump" command
		for fr.pc = 0; fr.pc < uint(len(fr.block.Instrs)); fr.pc++ {
			instr = fr.block.Instrs[fr.pc]
			if InstTracing() {
				fmt.Fprint(os.Stderr, fr.block.Index, fr.pc, "\t")
				if v, ok := instr.(ssa2.Value); ok {
					fmt.Fprintln(os.Stderr, "\t", v.Name(), "=", instr)
				} else {
					fmt.Fprintln(os.Stderr, "\t", instr)
				}
			}
			if fr.tracing == TRACE_STEP_INSTRUCTION {
				TraceHook(fr, &instr, ssa2.STEP_INSTRUCTION)
			}
			switch visitInstr(fr, instr) {
			case kReturn:
				switch return_instr := instr.(type) {
				case *ssa2.Return:
					fr.startP = return_instr.Pos()
					fr.endP   = return_instr.EndP()
				}
				fr.status = StComplete
				if (fr.tracing != TRACE_STEP_NONE) && GlobalStmtTracing() {
					TraceHook(fr, &instr, ssa2.CALL_RETURN)
				}
				return
			case kNext:
				// no-op
			case kJump:
				break block
			}
		}
	}
}

// doRecover implements the recover() built-in.
func doRecover(caller *Frame) Value {
	// recover() must be exactly one level beneath the deferred
	// function (two levels beneath the panicking function) to
	// have any effect.  Thus we ignore both "defer recover()" and
	// "defer f() -> g() -> recover()".
	if caller.i.Mode&DisableRecover == 0 &&
		caller != nil && !caller.panicking &&
		caller.caller != nil && caller.caller.panicking {
		caller.caller.panicking = false
		p := caller.caller.panic
		caller.caller.panic = nil
		switch p := p.(type) {
		case targetPanic:
			// The target program explicitly called panic().
			return p.v
		case runtime.Error:
			// The interpreter encountered a runtime error.
			return iface{caller.i.runtimeErrorString, p.Error()}
		case string:
			// The interpreter explicitly called panic().
			return iface{caller.i.runtimeErrorString, p}
		default:
			panic(fmt.Sprintf("unexpected panic type %T in target call to recover()", p))
		}
	}
	return iface{}
}

// setGlobal sets the value of a system-initialized global variable.
func setGlobal(i *interpreter, pkg *ssa2.Package, name string, v Value) {
	if g, ok := i.globals[pkg.Var(name)]; ok {
		*g = v
		return
	}
	panic("no global variable: " + pkg.Object.Path() + "." + name)
}

var stdSizes = types.StdSizes{WordSize: 8, MaxAlign: 8}

// Interpret interprets the Go program whose main package is mainpkg.
// mode specifies various interpreter options.  filename and args are
// the initial values of os.Args for the target program.
//
// Interpret returns the exit code of the program: 2 for panic (like
// gc does), or the argument to os.Exit for normal termination.
//
func Interpret(mainpkg *ssa2.Package, mode Mode, traceMode TraceMode,
	filename string, args []string) (exitCode int) {
	i = &interpreter{
		prog:    mainpkg.Prog,
		globals: make(map[ssa2.Value]*Value),
		Mode:    mode,
		TraceMode: traceMode,
		TraceEventMask: make(ssa2.TraceEventMask, ssa2.TRACE_EVENT_LAST),
	}

	runtimePkg := i.prog.ImportedPackage("runtime")
	if runtimePkg == nil {
		panic("Internal error: Interperter should have imported the runtime package")
	}
	i.runtimeErrorString = runtimePkg.Type("errorString").Object().Type()

	// Is there a way to combine the below loop into the above?
	for event := ssa2.TRACE_EVENT_FIRST; event <= ssa2.TRACE_EVENT_LAST; event++ {
		i.TraceEventMask[event] = true
	}
	i.TraceEventMask[ssa2.DEFER_ENTER] = false
	if i.TraceMode & EnableInitTracing == 0 {
		// clear tracing bits in init() functions that occur before
		// main.main()
		i.TraceMode &= ^(EnableStmtTracing|EnableTracing)
	}
	i.goTops = append(i.goTops, &GoreState{Fr: nil, state: 0})

	initReflect(i)

	for _, pkg := range i.prog.AllPackages() {
		// Initialize global storage.
		for _, m := range pkg.Members {
			switch v := m.(type) {
			case *ssa2.Global:
				cell := zero(deref(v.Type()))
				i.globals[v] = &cell
			}
		}

		// Ad-hoc initialization for magic system variables.
		switch pkg.Object.Path() {
		case "syscall":
			var envs []Value
			for _, s := range os.Environ() {
				envs = append(envs, s)
			}
			envs = append(envs, "GOSSAINTERP=1")
			envs = append(envs, "GOARCH="+runtime.GOARCH)
			setGlobal(i, pkg, "envs", envs)

		case "runtime":
			// (Assumes no custom Sizeof used during SSA construction.)
			sz := stdSizes.Sizeof(pkg.Object.Scope().Lookup("MemStats").Type())
			setGlobal(i, pkg, "sizeof_C_MStats", uintptr(sz))

		case "os":
			Args := []Value{filename}
			for _, s := range args {
				Args = append(Args, s)
			}
			setGlobal(i, pkg, "Args", Args)
		}
	}

	// Top-level error handler.
	exitCode = 2
	defer func() {
		if exitCode != 2 || (i.Mode & DisableRecover) != 0 {
			return
		}
		switch p := recover().(type) {
		case exitPanic:
			exitCode = int(p)
			return
		case targetPanic:
			fmt.Fprintln(os.Stderr, "panic:", toString(p.v))
		case runtime.Error:
			fmt.Fprintln(os.Stderr, "panic:", p.Error())
		case string:
			fmt.Fprintln(os.Stderr, "panic:", p)
		default:
			fmt.Fprintf(os.Stderr, "panic: unexpected type: %T\n", p)
		}

		// TODO(adonovan): dump panicking interpreter goroutine?
		// buf := make([]byte, 0x10000)
		// runtime.Stack(buf, false)
		// fmt.Fprintln(os.Stderr, string(buf))
		// (Or dump panicking target goroutine?)
	}()

	// Run!
	call(i, 0, nil, mainpkg.Func("init"), nil)
	if mainFn := mainpkg.Func("main"); mainFn != nil {
		// fr := &Frame{
		// 	i      : i,
		// 	caller : nil,
		// 	env    : make(map[ssa2.Value]Value),
		// 	block  : mainFn.Blocks[0],
		// 	locals : make([]Value, len(mainFn.Locals)),
		//  start  : mainFn.Pos(),
		//  end    : mainFn.Pos(),
		//  goNum  : 0 ,
		//  Var2Reg: make(map[string]string),
		// }
		// TraceHook(fr, nil&mainFn.Blocks[0].Instrs[0], ssa2.MAIN)

		// If we didn't set tracing before because EnableInitTracing
		// was off, we'll set it now.
		i.TraceMode = traceMode
		// Allow defer tracing now that we've hit main
		// On second thought. We catch defer enter with a call enter.
		// i.TraceEventMask[ssa2.DEFER_ENTER] = true
		call(i, 0, nil, mainFn, nil)
		exitCode = 0
	} else {
		fmt.Fprintln(os.Stderr, "No main function.")
		exitCode = 1
	}
	return
}

// deref returns a pointer's element type; otherwise it returns typ.
// TODO(adonovan): Import from ssa?
func deref(typ types.Type) types.Type {
	if p, ok := typ.Underlying().(*types.Pointer); ok {
		return p.Elem()
	}
	return typ
}
