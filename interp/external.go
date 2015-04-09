// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

// Emulated functions that we cannot interpret because they are
// external or because they use "unsafe" or "reflect" operations.

import (
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/rocky/ssa-interp"
	"github.com/rocky/go-types"
)

type externalFn func(fr *Frame, args []Value) Value

// TODO(adonovan): fix: reflect.Value abstracts an lvalue or an
// rvalue; Set() causes mutations that can be observed via aliases.
// We have not captured that correctly here.

// Key strings are from Function.String().
var externals map[string]externalFn

func init() {
	// That little dot ۰ is an Arabic zero numeral (U+06F0), categories [Nd].
	externals = map[string]externalFn{
		"(*sync.Pool).Get":                 ext۰sync۰Pool۰Get,
		"(*sync.Pool).Put":                 ext۰sync۰Pool۰Put,
		"(reflect.Value).Bool":             ext۰reflect۰Value۰Bool,
		"(reflect.Value).CanAddr":          ext۰reflect۰Value۰CanAddr,
		"(reflect.Value).CanInterface":     ext۰reflect۰Value۰CanInterface,
		"(reflect.Value).Elem":             ext۰reflect۰Value۰Elem,
		"(reflect.Value).Field":            ext۰reflect۰Value۰Field,
		"(reflect.Value).Float":            ext۰reflect۰Value۰Float,
		"(reflect.Value).Index":            ext۰reflect۰Value۰Index,
		"(reflect.Value).Int":              ext۰reflect۰Value۰Int,
		"(reflect.Value).Interface":        ext۰reflect۰Value۰Interface,
		"(reflect.Value).IsNil":            ext۰reflect۰Value۰IsNil,
		"(reflect.Value).IsValid":          ext۰reflect۰Value۰IsValid,
		"(reflect.Value).Kind":             ext۰reflect۰Value۰Kind,
		"(reflect.Value).Len":              ext۰reflect۰Value۰Len,
		"(reflect.Value).MapIndex":         ext۰reflect۰Value۰MapIndex,
		"(reflect.Value).MapKeys":          ext۰reflect۰Value۰MapKeys,
		"(reflect.Value).NumField":         ext۰reflect۰Value۰NumField,
		"(reflect.Value).NumMethod":        ext۰reflect۰Value۰NumMethod,
		"(reflect.Value).Pointer":          ext۰reflect۰Value۰Pointer,
		"(reflect.Value).Set":              ext۰reflect۰Value۰Set,
		"(reflect.Value).String":           ext۰reflect۰Value۰String,
		"(reflect.Value).Type":             ext۰reflect۰Value۰Type,
		"(reflect.Value).Uint":             ext۰reflect۰Value۰Uint,
		"(reflect.error).Error":            ext۰reflect۰error۰Error,
		"(reflect.rtype).Bits":             ext۰reflect۰rtype۰Bits,
		"(reflect.rtype).Elem":             ext۰reflect۰rtype۰Elem,
		"(reflect.rtype).Field":            ext۰reflect۰rtype۰Field,
		"(reflect.rtype).Kind":             ext۰reflect۰rtype۰Kind,
		"(reflect.rtype).NumField":         ext۰reflect۰rtype۰NumField,
		"(reflect.rtype).NumMethod":        ext۰reflect۰rtype۰NumMethod,
		"(reflect.rtype).NumOut":           ext۰reflect۰rtype۰NumOut,
		"(reflect.rtype).Out":              ext۰reflect۰rtype۰Out,
		"(reflect.rtype).Size":             ext۰reflect۰rtype۰Size,
		"(reflect.rtype).String":           ext۰reflect۰rtype۰String,
		"bytes.Equal":                      ext۰bytes۰Equal,
		"bytes.IndexByte":                  ext۰bytes۰IndexByte,
		"hash/crc32.haveSSE42":             ext۰crc32۰haveSSE42,
		"math.Abs":                         ext۰math۰Abs,
		"math/big.bitLen":                  ext۰math۰big۰bitLen,
		"math/big.divWVW":                  ext۰math۰big۰divWVW,
		"math/big.shlVU":                   ext۰math۰big۰shlVU,
		"math/big.shrVU":                   ext۰math۰big۰shrVU,
		"math/big.subVV":                   ext۰math۰big۰subVV,
		"math.Exp":                         ext۰math۰Exp,
		"math.Float32bits":                 ext۰math۰Float32bits,
		"math.Float32frombits":             ext۰math۰Float32frombits,
		"math.Float64bits":                 ext۰math۰Float64bits,
		"math.Float64frombits":             ext۰math۰Float64frombits,
		"math.Ldexp":                       ext۰math۰Ldexp,
		"math.Log":                         ext۰math۰Log,
		"math.Log2":                        ext۰math۰Log2,
		"math.Min":                         ext۰math۰Min,
		"os.runtime_args":                  ext۰os۰runtime_args,
		"os.runtime_beforeExit":            ext۰os۰runtime_beforeExit,
		"reflect.New":                      ext۰reflect۰New,
		"reflect.SliceOf":                  ext۰reflect۰SliceOf,
		"reflect.TypeOf":                   ext۰reflect۰TypeOf,
		"reflect.ValueOf":                  ext۰reflect۰ValueOf,
		"reflect.init":                     ext۰reflect۰Init,
		"reflect.valueInterface":           ext۰reflect۰valueInterface,
		"runtime.Breakpoint":               ext۰runtime۰Breakpoint,
		"runtime.Caller":                   ext۰runtime۰Caller,
		"runtime.Callers":                  ext۰runtime۰Callers,
		"runtime.FuncForPC":                ext۰runtime۰FuncForPC,
		"runtime.GC":                       ext۰runtime۰GC,
		"runtime.GOMAXPROCS":               ext۰runtime۰GOMAXPROCS,
		"runtime.Goexit":                   ext۰runtime۰Goexit,
		"runtime.Gosched":                  ext۰runtime۰Gosched,
		"runtime.init":                     ext۰runtime۰init,
		"runtime.NumCPU":                   ext۰runtime۰NumCPU,
		"runtime.ReadMemStats":             ext۰runtime۰ReadMemStats,
		"runtime.SetFinalizer":             ext۰runtime۰SetFinalizer,
		"(*runtime.Func).Entry":            ext۰runtime۰Func۰Entry,
		"(*runtime.Func).FileLine":         ext۰runtime۰Func۰FileLine,
		"(*runtime.Func).Name":             ext۰runtime۰Func۰Name,
		"runtime.environ":                  ext۰runtime۰environ,
		"runtime.getgoroot":                ext۰runtime۰getgoroot,
		"strings.IndexByte":                ext۰strings۰IndexByte,
		"sync.runtime_Semacquire":          ext۰sync۰runtime_Semacquire,
		"sync.runtime_Semrelease":          ext۰sync۰runtime_Semrelease,
		"sync.runtime_Syncsemcheck":        ext۰sync۰runtime_Syncsemcheck,
		"sync.runtime_registerPoolCleanup": ext۰sync۰runtime_registerPoolCleanup,
		"sync/atomic.AddInt32":             ext۰atomic۰AddInt32,
		"sync/atomic.AddUint32":            ext۰atomic۰AddUint32,
		"sync/atomic.AddUint64":            ext۰atomic۰AddUint64,
		"sync/atomic.CompareAndSwapInt32":  ext۰atomic۰CompareAndSwapInt32,
		"sync/atomic.LoadInt32":            ext۰atomic۰LoadInt32,
		"sync/atomic.LoadUint32":           ext۰atomic۰LoadUint32,
		"sync/atomic.StoreInt32":           ext۰atomic۰StoreInt32,
		"sync/atomic.StoreUint32":          ext۰atomic۰StoreUint32,
		"syscall.Close":                    ext۰syscall۰Close,
		"syscall.Exit":                     ext۰syscall۰Exit,
		"syscall.Fstat":                    ext۰syscall۰Fstat,
		"syscall.Getpid":                   ext۰syscall۰Getpid,
		"syscall.Getuid":                   ext۰syscall۰Getuid,
		"syscall.Getwd":                    ext۰syscall۰Getwd,
		"syscall.Kill":                     ext۰syscall۰Kill,
		"syscall.Lstat":                    ext۰syscall۰Lstat,
		"syscall.Open":                     ext۰syscall۰Open,
		"syscall.ParseDirent":              ext۰syscall۰ParseDirent,
		"syscall.RawSyscall":               ext۰syscall۰RawSyscall,
		"syscall.Read":                     ext۰syscall۰Read,
		"syscall.ReadDirent":               ext۰syscall۰ReadDirent,
		"syscall.Stat":                     ext۰syscall۰Stat,
		"syscall.Write":                    ext۰syscall۰Write,
		"syscall.runtime_envs":             ext۰runtime۰environ,
		"time.Sleep":                       ext۰time۰Sleep,
		"time.now":                         ext۰time۰now,
		"github.com/rocky/ssa-interp/trepan.Debug":  ext۰trepan۰Debug,
	}
}

// wrapError returns an interpreted 'error' interface value for err.
func wrapError(err error) Value {
	if err == nil {
		return iface{}
	}
	return iface{t: errorType, v: err.Error()}
}

func ext۰sync۰Pool۰Get(fr *Frame, args []Value) Value {
	Pool := fr.i.prog.ImportedPackage("sync").Type("Pool").Object()
	_, newIndex, _ := types.LookupFieldOrMethod(Pool.Type(), false, Pool.Pkg(), "New")

	if New := (*args[0].(*Value)).(Structure).fields[newIndex[0]]; New != nil {
		return call(fr.i, fr.goNum, fr, New, nil)
	}
	return nil
}

func ext۰sync۰Pool۰Put(fr *Frame, args []Value) Value {
	return nil
}

func ext۰bytes۰Equal(fr *Frame, args []Value) Value {
	// func Equal(a, b []byte) bool
	a := args[0].([]Value)
	b := args[1].([]Value)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ext۰bytes۰IndexByte(fr *Frame, args []Value) Value {
	// func IndexByte(s []byte, c byte) int
	s := args[0].([]Value)
	c := args[1].(byte)
	for i, b := range s {
		if b.(byte) == c {
			return i
		}
	}
	return -1
}

func ext۰crc32۰haveSSE42(fr *Frame, args []Value) Value {
	return false
}

func ext۰os۰runtime_args(fr *Frame, args []Value) Value {
	return fr.i.osArgs
}

func ext۰os۰runtime_beforeExit(fr *Frame, args []Value) Value {
	return nil
}

func ext۰runtime۰Breakpoint(fr *Frame, args []Value) Value {
	// If tracehook is DefaultTraceHook, should we run a PrintStack
	// and leave?
	TraceHook(fr, &fr.block.Instrs[0], ssa2.TRACE_CALL)
	runtime.Breakpoint()
	return nil
}

func ext۰runtime۰environ(fr *Frame, args []Value) Value {
	return environ
}

func ext۰runtime۰getgoroot(fr *Frame, args []Value) Value {
	return os.Getenv("GOROOT")
}

func ext۰strings۰IndexByte(fr *Frame, args []Value) Value {
	// func IndexByte(s string, c byte) int
	s := args[0].(string)
	c := args[1].(byte)
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func ext۰sync۰runtime_Syncsemcheck(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: implement.
	return nil
}

func ext۰sync۰runtime_registerPoolCleanup(fr *Frame, args []Value) Value {
	return nil
}

func ext۰sync۰runtime_Semacquire(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: implement.
	return nil
}

func ext۰sync۰runtime_Semrelease(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: implement.
	return nil
}

func ext۰runtime۰GOMAXPROCS(fr *Frame, args []Value) Value {
	// Ignore args[0]; don't let the interpreted program
	// set the interpreter's GOMAXPROCS!
	return runtime.GOMAXPROCS(0)
}

func ext۰runtime۰Goexit(fr *Frame, args []Value) Value {
	// TODO(adonovan): don't kill the interpreter's main goroutine.
	runtime.Goexit()
	return nil
}

func ext۰runtime۰GC(fr *Frame, args []Value) Value {
	runtime.GC()
	return nil
}

func ext۰runtime۰Gosched(fr *Frame, args []Value) Value {
	runtime.Gosched()
	return nil
}

func ext۰runtime۰init(fr *Frame, args []Value) Value {
	return nil
}

func ext۰runtime۰NumCPU(fr *Frame, args []Value) Value {
	return runtime.NumCPU()
}

func ext۰runtime۰ReadMemStats(fr *Frame, args []Value) Value {
	// TODO(adonovan): populate args[0].(Struct)
	return nil
}

func ext۰atomic۰LoadUint32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	return (*args[0].(*Value)).(uint32)
}

func ext۰atomic۰StoreUint32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	*args[0].(*Value) = args[1].(uint32)
	return nil
}

func ext۰atomic۰LoadInt32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	return (*args[0].(*Value)).(int32)
}

func ext۰atomic۰StoreInt32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	*args[0].(*Value) = args[1].(int32)
	return nil
}

func ext۰atomic۰CompareAndSwapInt32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	p := args[0].(*Value)
	if (*p).(int32) == args[1].(int32) {
		*p = args[2].(int32)
		return true
	}
	return false
}

func ext۰atomic۰AddInt32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	p := args[0].(*Value)
	newv := (*p).(int32) + args[1].(int32)
	*p = newv
	return newv
}

func ext۰atomic۰AddUint32(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	p := args[0].(*Value)
	newv := (*p).(uint32) + args[1].(uint32)
	*p = newv
	return newv
}

func ext۰atomic۰AddUint64(fr *Frame, args []Value) Value {
	// TODO(adonovan): fix: not atomic!
	p := args[0].(*Value)
	newv := (*p).(uint64) + args[1].(uint64)
	*p = newv
	return newv
}

func ext۰runtime۰SetFinalizer(fr *Frame, args []Value) Value {
	return nil // ignore
}

// Pretend: type runtime.Func struct { entry *ssa2.Function }

func ext۰runtime۰Func۰FileLine(fr *Frame, args []Value) Value {
	// func (*runtime.Func) FileLine(uintptr) (string, int)
	f, _ := (*args[0].(*Value)).(Structure).fields[0].(*ssa2.Function)
	pc := args[1].(uintptr)
	_ = pc
	if f != nil {
		// TODO(adonovan): use position of current instruction, not fn.
		posn := f.Prog.Fset.Position(f.Pos())
		return tuple{posn.Filename, posn.Line}
	}
	return tuple{"", 0}
}

func ext۰runtime۰Func۰Name(fr *Frame, args []Value) Value {
	// func (*runtime.Func) Name() string
	f, _ := (*args[0].(*Value)).(Structure).fields[0].(*ssa2.Function)
	if f != nil {
		return f.String()
	}
	return ""
}

func ext۰runtime۰Func۰Entry(fr *Frame, args []Value) Value {
	// func (*runtime.Func) Entry() uintptr
	f, _ := (*args[0].(*Value)).(Structure).fields[0].(*ssa2.Function)
	return uintptr(unsafe.Pointer(f))
}

func ext۰time۰now(fr *Frame, args []Value) Value {
	nano := time.Now().UnixNano()
	return tuple{int64(nano / 1e9), int32(nano % 1e9)}
}

func ext۰time۰Sleep(fr *Frame, args []Value) Value {
	time.Sleep(time.Duration(args[0].(int64)))
	return nil
}

func ext۰syscall۰Exit(fr *Frame, args []Value) Value {
	panic(exitPanic(args[0].(int)))
}

func ext۰syscall۰Getwd(fr *Frame, args []Value) Value {
	s, err := syscall.Getwd()
	return tuple{s, wrapError(err)}
}

func ext۰syscall۰Getuid(fr *Frame, args []Value) Value {
	return syscall.Getuid()
}

func ext۰syscall۰Getpid(fr *Frame, args []Value) Value {
	return syscall.Getpid()
}

func ValueToBytes(v Value) []byte {
	in := v.([]Value)
	b := make([]byte, len(in))
	for i := range in {
		b[i] = in[i].(byte)
	}
	return b
}
