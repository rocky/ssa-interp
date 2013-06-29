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
// * the equivalence relation for structs doesn't skip over blank
// fields.
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
	"go/token"
	"os"
	"reflect"
	"runtime"

	"ssa-interp"

	"code.google.com/p/go.tools/go/types"
)

// FIXME nuke this.
func init() {
	// TraceHook = NullTraceHook
	TraceHook = GubTraceHook
}

type continuation int

const (
	kNext continuation = iota
	kReturn
	kJump
)

type traceType int

const (
	TRACE_STEP_NONE = iota
	TRACE_STEP_IN
	TRACE_STEP_OVER
	TRACE_STEP_OUT
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

var i Interpreter

func GetInterpeter() Interpreter {
	return i
}

func (fr *frame) get(key ssa2.Value) value {
	switch key := key.(type) {
	case nil:
		// Hack; simplifies handling of optional attributes
		// such as ssa2.Slice.{Low,High}.
		return nil
	case *ssa2.Function, *ssa2.Builtin:
		return key
	case *ssa2.Literal:
		return literalValue(key)
	case *ssa2.Global:
		if r, ok := fr.i.Globals[key]; ok {
			return r
		}
	}
	if r, ok := fr.Env[key]; ok {
		return r
	}
	panic(fmt.Sprintf("get: no value for %T: %v", key, key.Name()))
}

func (fr *frame) rundefers() {
	for i := range fr.defers {
		if (fr.i.TraceMode & EnableTracing) != 0 {
			fmt.Fprintln(os.Stderr, "Invoking deferred function", i)
		}
		fr.defers[len(fr.defers)-1-i]()
	}
	fr.defers = fr.defers[:0]
}

// findMethodSet returns the method set for type typ, which may be one
// of the interpreter's fake types.
func findMethodSet(i *Interpreter, typ types.Type) ssa2.MethodSet {
	switch typ {
	case rtypeType:
		return i.rtypeMethods
	case errorType:
		return i.errorMethods
	}
	return i.Prog.MethodSet(typ)
}

// visitInstr interprets a single ssa2.Instruction within the activation
// record frame.  It returns a continuation value indicating where to
// read the next instruction from.
func visitInstr(fr *frame, genericInstr ssa2.Instruction) continuation {
	switch instr := genericInstr.(type) {
	case *ssa2.UnOp:
		fr.Env[instr] = unop(instr, fr.get(instr.X))

	case *ssa2.BinOp:
		fr.Env[instr] = binop(instr.Op, fr.get(instr.X), fr.get(instr.Y))

	case *ssa2.Call:
		fn, args := prepareCall(fr, &instr.Call)
		fr.Env[instr] = call(fr.i, fr, instr.Pos(), fn, args)

	case *ssa2.ChangeInterface:
		x := fr.get(instr.X)
		if x.(iface).t == nil {
			panic(fmt.Sprintf("interface conversion: interface is nil, not %s", instr.Type()))
		}
		fr.Env[instr] = x

	case *ssa2.ChangeType:
		fr.Env[instr] = fr.get(instr.X) // (can't fail)

	case *ssa2.Convert:
		fr.Env[instr] = conv(instr.Type(), instr.X.Type(), fr.get(instr.X))

	case *ssa2.MakeInterface:
		fr.Env[instr] = iface{t: instr.X.Type(), v: fr.get(instr.X)}

	case *ssa2.Extract:
		fr.Env[instr] = fr.get(instr.Tuple).(tuple)[instr.Index]

	case *ssa2.Slice:
		fr.Env[instr] = slice(fr.get(instr.X), fr.get(instr.Low), fr.get(instr.High))

	case *ssa2.Ret:
		switch len(instr.Results) {
		case 0:
		case 1:
			fr.result = fr.get(instr.Results[0])
		default:
			var res []value
			for _, r := range instr.Results {
				res = append(res, fr.get(r))
			}
			fr.result = tuple(res)
		}
		return kReturn

	case *ssa2.RunDefers:
		fr.rundefers()

	case *ssa2.Panic:
		panic(targetPanic{fr.get(instr.X)})

	case *ssa2.Send:
		fr.get(instr.Chan).(chan value) <- copyVal(fr.get(instr.X))

	case *ssa2.Store:
		*fr.get(instr.Addr).(*value) = copyVal(fr.get(instr.Val))

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
		fr.defers = append(fr.defers, func() { call(fr.i, fr, instr.Pos(), fn, args) })
	case *ssa2.Go:
		fn, args := prepareCall(fr, &instr.Call)
		go call(fr.i, nil, instr.Pos(), fn, args)

	case *ssa2.MakeChan:
		fr.Env[instr] = make(chan value, asInt(fr.get(instr.Size)))

	case *ssa2.Alloc:
		var addr *value
		if instr.Heap {
			// new
			addr = new(value)
			fr.Env[instr] = addr
		} else {
			// local
			addr = fr.Env[instr].(*value)
		}
		*addr = zero(instr.Type().Deref())

	case *ssa2.MakeSlice:
		slice := make([]value, asInt(fr.get(instr.Cap)))
		tElt := instr.Type().Underlying().(*types.Slice).Elem()
		for i := range slice {
			slice[i] = zero(tElt)
		}
		fr.Env[instr] = slice[:asInt(fr.get(instr.Len))]

	case *ssa2.MakeMap:
		reserve := 0
		if instr.Reserve != nil {
			reserve = asInt(fr.get(instr.Reserve))
		}
		fr.Env[instr] = makeMap(instr.Type().Underlying().(*types.Map).Key(), reserve)

	case *ssa2.Range:
		fr.Env[instr] = rangeIter(fr.get(instr.X), instr.X.Type())

	case *ssa2.Next:
		fr.Env[instr] = fr.get(instr.Iter).(iter).next()

	case *ssa2.FieldAddr:
		x := fr.get(instr.X)
		fr.Env[instr] = &(*x.(*value)).(structure)[instr.Field]

	case *ssa2.Field:
		fr.Env[instr] = copyVal(fr.get(instr.X).(structure)[instr.Field])

	case *ssa2.IndexAddr:
		x := fr.get(instr.X)
		idx := fr.get(instr.Index)
		switch x := x.(type) {
		case []value:
			fr.Env[instr] = &x[asInt(idx)]
		case *value: // *array
			fr.Env[instr] = &(*x).(array)[asInt(idx)]
		default:
			panic(fmt.Sprintf("unexpected x type in IndexAddr: %T", x))
		}

	case *ssa2.Index:
		fr.Env[instr] = copyVal(fr.get(instr.X).(array)[asInt(fr.get(instr.Index))])

	case *ssa2.Lookup:
		fr.Env[instr] = lookup(instr, fr.get(instr.X), fr.get(instr.Index))

	case *ssa2.MapUpdate:
		m := fr.get(instr.Map)
		key := fr.get(instr.Key)
		v := fr.get(instr.Value)
		switch m := m.(type) {
		case map[value]value:
			m[key] = v
		case *hashmap:
			m.insert(key.(hashable), v)
		default:
			panic(fmt.Sprintf("illegal map type: %T", m))
		}

	case *ssa2.TypeAssert:
		fr.Env[instr] = typeAssert(fr.i, instr, fr.get(instr.X).(iface))

	case *ssa2.Trace:
		fr.startP = instr.Start
		fr.endP   = instr.End
		if (fr.tracing == TRACE_STEP_IN) || (fr.tracing == TRACE_STEP_OVER) && GlobalStmtTracing() {
			TraceHook(fr, &genericInstr, instr.Event)
		}

	case *ssa2.MakeClosure:
		var bindings []value
		for _, binding := range instr.Bindings {
			bindings = append(bindings, fr.get(binding))
		}
		fr.Env[instr] = &closure{instr.Fn.(*ssa2.Function), bindings}

	case *ssa2.Phi:
		for i, pred := range instr.Block().Preds {
			if fr.prevBlock == pred {
				fr.Env[instr] = fr.get(instr.Edges[i])
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
				var v value
				if i == chosen && recvOk {
					// No need to copy since send makes an unaliased copy.
					v = recv.Interface().(value)
				} else {
					v = zero(st.Chan.Type().Underlying().(*types.Chan).Elem())
				}
				r = append(r, v)
			}
		}
		fr.Env[instr] = r

	default:
		panic(fmt.Sprintf("unexpected instruction: %T", instr))
	}

	// if val, ok := instr.(ssa2.Value); ok {
	// 	fmt.Println(toString(fr.Env[val])) // debugging
	// }

	return kNext
}

// prepareCall determines the function value and argument values for a
// function call in a Call, Go or Defer instruction, performing
// interface method lookup if needed.
//
func prepareCall(fr *frame, call *ssa2.CallCommon) (fn value, args []value) {
	if call.Func != nil {
		// Function call.
		fn = fr.get(call.Func)
	} else {
		// Interface method invocation.
		recv := fr.get(call.Recv).(iface)
		if recv.t == nil {
			panic("method invoked on nil interface")
		}
		id := call.MethodId()
		fn = findMethodSet(fr.i, recv.t)[id]
		if fn == nil {
			// Unreachable in well-typed programs.
			panic(fmt.Sprintf("method set for dynamic type %v does not contain %s", recv.t, id))
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
func call(i *Interpreter, caller *frame, callpos token.Pos, fn value, args []value) value {
	switch fn := fn.(type) {
	case *ssa2.Function:
		if fn == nil {
			panic("call of nil function") // nil of func type
		}
		return callSSA(i, caller, callpos, fn, args, nil)
	case *closure:
		return callSSA(i, caller, callpos, fn.fn, args, fn.Env)
	case *ssa2.Builtin:
		return callBuiltin(caller, callpos, fn, args)
	}
	panic(fmt.Sprintf("cannot call %T", fn))
}

func loc(fset *token.FileSet, pos token.Pos) string {
	if pos == token.NoPos {
		return ""
	}
	return " at " + fset.Position(pos).String()
}

// callSSA interprets a call to function fn with arguments args,
// and lexical environment env, returning its result.
// callpos is the position of the callsite.
//
func callSSA(i *Interpreter, caller *frame, callpos token.Pos, fn *ssa2.Function, args []value, env []value) value {
	if InstTracing() {
		fset := fn.Prog.Fset
		// TODO(adonovan): fix: loc() lies for external functions.
		fmt.Fprintf(os.Stderr, "\tEntering %s%s.\n", fn.FullName(), loc(fset, fn.Pos()))
		suffix := ""
		if caller != nil {
			suffix = ", resuming " + caller.fn.FullName() + loc(fset, callpos)
		}
		defer fmt.Fprintf(os.Stderr, "\tLeaving %s%s.\n", fn.FullName(), suffix)
	}
	if fn.Enclosing == nil {
		name := fn.FullName()
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
	fr := &frame{
		i:      i,
		caller: caller,
		fn:     fn,
		Env:    make(map[ssa2.Value]value),
		block:  fn.Blocks[0],
		locals: make([]value, len(fn.Locals)),
		tracing: TRACE_STEP_NONE,
	}

	if caller == nil {
		if GlobalStmtTracing() {
			fr.tracing = TRACE_STEP_IN
		}
	} else if caller.tracing == TRACE_STEP_IN {
		fr.tracing = TRACE_STEP_IN
	}

	for i, l := range fn.Locals {
		fr.locals[i] = zero(l.Type().Deref())
		fr.Env[l] = &fr.locals[i]
	}
	for i, p := range fn.Params {
		fr.Env[p] = args[i]
	}
	for i, fv := range fn.FreeVars {
		fr.Env[fv] = env[i]
	}
	var instr ssa2.Instruction

	defer func() {
		if fr.Status != StComplete {
			if (fr.i.Mode & DisableRecover) != 0 {
				return // let interpreter crash
			}
			fr.Status, fr.panic = StPanic, recover()
		}
		fr.rundefers()
		// Destroy the locals to avoid accidental use after return.
		for i := range fn.Locals {
			fr.locals[i] = bad{}
		}
		if fr.Status == StPanic {
			panic(fr.panic) // panic stack is not entirely clean
		}
	}()

	fr.startP = fn.Pos()
	fr.endP   = fn.Pos()
	if fr.tracing == TRACE_STEP_IN && len(fr.block.Instrs) > 0 && GlobalStmtTracing() {
		TraceHook(fr, &fr.block.Instrs[0], ssa2.CALL_ENTER)
	}
	for {
		if InstTracing() {
			fmt.Fprintf(os.Stderr, ".%s:\n", fr.block)
		}
	block:
		// rocky: may eventually want to change this
		// if I want to allow a "skip" or "jump" command
		for fr.pc, instr = range fr.block.Instrs {
			if InstTracing() {
				fmt.Fprint(os.Stderr, fr.block.Index, fr.pc, "\t")
				if v, ok := instr.(ssa2.Value); ok {
					fmt.Fprintln(os.Stderr, v.Name(), "=", instr)
				} else {
					fmt.Fprintln(os.Stderr, instr)
				}
			}
			switch visitInstr(fr, instr) {
			case kReturn:
				fr.Status = StComplete
				fr.startP = instr.Pos()
				fr.endP   = instr.Pos()
				if (fr.tracing != TRACE_STEP_NONE) && GlobalStmtTracing() {
					TraceHook(fr, &instr, ssa2.CALL_RETURN)
				}
				return fr.result
			case kNext:
				// no-op
			case kJump:
				break block
			}
		}

	}
	panic("unreachable")
}

// setGlobal sets the value of a system-initialized global variable.
func setGlobal(i *Interpreter, pkg *ssa2.Package, name string, v value) {
	if g, ok := i.Globals[pkg.Var(name)]; ok {
		*g = v
		return
	}
	panic("no global variable: " + pkg.Types.Path() + "." + name)
}

// Interpret interprets the Go program whose main package is mainpkg.
// mode specifies various interpreter options.  filename and args are
// the initial values of os.Args for the target program.
//
// Interpret returns the exit code of the program: 2 for panic (like
// gc does), or the argument to os.Exit for normal termination.
//
func Interpret(mainpkg *ssa2.Package, mode Mode, traceMode TraceMode,
	filename string, args []string) (exitCode int) {
	i = Interpreter{
		Prog:    mainpkg.Prog,
		Globals: make(map[ssa2.Value]*value),
		Mode:    mode,
		TraceMode: traceMode & ^(EnableStmtTracing|EnableTracing),
	}
	initReflect(&i)

	for importPath, pkg := range i.Prog.Packages {
		// Initialize global storage.
		for _, m := range pkg.Members {
			switch v := m.(type) {
			case *ssa2.Global:
				cell := zero(v.Type().Deref())
				i.Globals[v] = &cell
			}
		}

		// Ad-hoc initialization for magic system variables.
		switch importPath {
		case "syscall":
			var envs []value
			for _, s := range os.Environ() {
				envs = append(envs, s)
			}
			envs = append(envs, "GOSSAINTERP=1")
			setGlobal(&i, pkg, "envs", envs)

		case "runtime":
			// TODO(gri): expose go/types.sizeof so we can
			// avoid this fragile magic number;
			// unsafe.Sizeof(memStats) won't work since gc
			// and go/types have different sizeof
			// functions.
			setGlobal(&i, pkg, "sizeof_C_MStats", uintptr(3696))

		case "os":
			Args := []value{filename}
			for _, s := range args {
				Args = append(Args, s)
			}
			setGlobal(&i, pkg, "Args", Args)
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
	call(&i, nil, token.NoPos, mainpkg.Init, nil)
	if mainFn := mainpkg.Func("main"); mainFn != nil {
		// fr := &frame{
		// 	I: &I,
		// 	caller: nil,
		// 	Env:    make(map[ssa2.Value]value),
		// 	Block:  mainFn.Blocks[0],
		// 	locals: make([]value, len(mainFn.Locals)),
		//  Start : mainFn.Pos()
		//  End   : mainFn.Pos()
		// }
		// TraceHook(fr, nil&mainFn.Blocks[0].Instrs[0], ssa2.MAIN)
		i.TraceMode = traceMode
		call(&i, nil, token.NoPos, mainFn, nil)
		exitCode = 0
	} else {
		fmt.Fprintln(os.Stderr, "No main function.")
		exitCode = 1
	}
	return
}
