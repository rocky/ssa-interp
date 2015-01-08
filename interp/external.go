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

	"github.com/rocky/ssa-interp"
)

type externalFn func(fr *Frame, args []Value) Value

// TODO(adonovan): fix: reflect.Value abstracts an lvalue or an
// rvalue; Set() causes mutations that can be observed via aliases.
// We have not captured that correctly here.

// Key strings are from Function.FullName().
// That little dot ۰ is an Arabic zero numeral (U+06F0), categories [Nd].
var externals = map[string]externalFn{
	"(*runtime.Func).Entry":           ext۰runtime۰Func۰Entry,
	"(*runtime.Func).FileLine":        ext۰runtime۰Func۰FileLine,
	"(*runtime.Func).Name":            ext۰runtime۰Func۰Name,
	"(reflect.Value).Bool":            ext۰reflect۰Value۰Bool,
	"(reflect.Value).CanAddr":         ext۰reflect۰Value۰CanAddr,
	"(reflect.Value).CanInterface":    ext۰reflect۰Value۰CanInterface,
	"(reflect.Value).Elem":            ext۰reflect۰Value۰Elem,
	"(reflect.Value).Field":           ext۰reflect۰Value۰Field,
	"(reflect.Value).Float":           ext۰reflect۰Value۰Float,
	"(reflect.Value).Index":           ext۰reflect۰Value۰Index,
	"(reflect.Value).Int":             ext۰reflect۰Value۰Int,
	"(reflect.Value).Interface":       ext۰reflect۰Value۰Interface,
	"(reflect.Value).IsNil":           ext۰reflect۰Value۰IsNil,
	"(reflect.Value).IsValid":         ext۰reflect۰Value۰IsValid,
	"(reflect.Value).Kind":            ext۰reflect۰Value۰Kind,
	"(reflect.Value).Len":             ext۰reflect۰Value۰Len,
	"(reflect.Value).MapIndex":        ext۰reflect۰Value۰MapIndex,
	"(reflect.Value).MapKeys":         ext۰reflect۰Value۰MapKeys,
	"(reflect.Value).NumField":        ext۰reflect۰Value۰NumField,
	"(reflect.Value).NumMethod":       ext۰reflect۰Value۰NumMethod,
	"(reflect.Value).Pointer":         ext۰reflect۰Value۰Pointer,
	"(reflect.Value).Set":             ext۰reflect۰Value۰Set,
	"(reflect.Value).String":          ext۰reflect۰Value۰String,
	"(reflect.Value).Type":            ext۰reflect۰Value۰Type,
	"(reflect.Value).Uint":            ext۰reflect۰Value۰Uint,
	"(reflect.error).Error":           ext۰reflect۰error۰Error,
	"(reflect.rtype).Bits":            ext۰reflect۰rtype۰Bits,
	"(reflect.rtype).Elem":            ext۰reflect۰rtype۰Elem,
	"(reflect.rtype).Field":           ext۰reflect۰rtype۰Field,
	"(reflect.rtype).Kind":            ext۰reflect۰rtype۰Kind,
	"(reflect.rtype).NumField":        ext۰reflect۰rtype۰NumField,
	"(reflect.rtype).NumMethod":       ext۰reflect۰rtype۰NumMethod,
	"(reflect.rtype).NumOut":          ext۰reflect۰rtype۰NumOut,
	"(reflect.rtype).Out":             ext۰reflect۰rtype۰Out,
	"(reflect.rtype).Size":            ext۰reflect۰rtype۰Size,
	"(reflect.rtype).String":          ext۰reflect۰rtype۰String,
	"bytes.Equal":                     ext۰bytes۰Equal,
	"bytes.IndexByte":                 ext۰bytes۰IndexByte,
	"strings.IndexByte":               ext۰strings۰IndexByte,
	"hash/crc32.haveSSE42":            ext۰crc32۰haveSSE42,
	"math.Abs":                        ext۰math۰Abs,
	"math.Acos":                       ext۰math۰Acos,
	"math.Asin":                       ext۰math۰Asin,
	"math.Atan":                       ext۰math۰Atan,
	"math.Atan2":                      ext۰math۰Atan2,
	"math.Ceil":                       ext۰math۰Ceil,
	"math.Cos":                        ext۰math۰Cos,
	"math.Dim":                        ext۰math۰Dim,
	"math.Exp":                        ext۰math۰Exp,
	"math.Expm1":                      ext۰math۰Expm1,
	"math.Floor":                      ext۰math۰Floor,
	"math.Frexp":                      ext۰math۰Frexp,
	"math.Hypot":                      ext۰math۰Hypot,
	"math.Ldexp":                      ext۰math۰Ldexp,
	"math.Log":                        ext۰math۰Log,
	"math.Log10":                      ext۰math۰Log10,
	"math.Log1p":                      ext۰math۰Log1p,
	"math.Max":                        ext۰math۰Max,
	"math.Min":                        ext۰math۰Min,
	"math.Mod":                        ext۰math۰Mod,
	"math.Modf":                       ext۰math۰Modf,
	"math.Remainder":                  ext۰math۰Remainder,
	"math.Sin":                        ext۰math۰Sin,
	"math.Sincos":                     ext۰math۰Sincos,
	"math.Sqrt":                       ext۰math۰Sqrt,
	"math.Tan":                        ext۰math۰Tan,
	"math.Trunc":                      ext۰math۰Trunc,

	"math/big.bitLen":                 ext۰math۰big۰bitLen,
	"math/big.mulAddVWW":              ext۰math۰big۰mulAddVWW,
	"math/big.shlVU":                  ext۰math۰big۰shlVU,

	"runtime/debug.function":          ext۰debug۰function,
	"runtime/debug.PrintStack":        ext۰debug۰PrintStack,
	// "(big.nat).shl":                   ext۰big۰shl,
	// "math/big.addVV":                  ext۰big۰addVV,
	// "math/big.shlVU":                  ext۰big۰shlVU,
	// "math/big.bitLen":                 ext۰big۰bitLen,
	"math.Float32bits":                ext۰math۰Float32bits,
	"math.Float32frombits":            ext۰math۰Float32frombits,
	"math.Float64bits":                ext۰math۰Float64bits,
	"math.Float64frombits":            ext۰math۰Float64frombits,
	"math.Log2":                       ext۰math۰Log2,
	"os.Exit":                         ext۰os۰Exit,
	"reflect.New":                     ext۰reflect۰New,
	"reflect.TypeOf":                  ext۰reflect۰TypeOf,
	"reflect.ValueOf":                 ext۰reflect۰ValueOf,
	"reflect.init":                    ext۰reflect۰Init,
	"reflect.valueInterface":          ext۰reflect۰valueInterface,
	"runtime.Breakpoint":              ext۰runtime۰Breakpoint,
	"runtime.Caller":                  ext۰runtime۰Caller,
	"runtime.Callers":                 ext۰runtime۰Callers,
	"runtime.FuncForPC":               ext۰runtime۰FuncForPC,
	"runtime.GC":                      ext۰runtime۰GC,
	"runtime.GOMAXPROCS":              ext۰runtime۰GOMAXPROCS,
	"runtime.Gosched":                 ext۰runtime۰Gosched,
	"runtime.NumCPU":                  ext۰runtime۰NumCPU,
	"runtime.ReadMemStats":            ext۰runtime۰ReadMemStats,
	"runtime.SetFinalizer":            ext۰runtime۰SetFinalizer,
	"runtime.getgoroot":               ext۰runtime۰getgoroot,
	"runtime.Goexit":                  ext۰runtime۰Goexit,
	"runtime.Stack":                   ext۰runtime۰Stack,
	"sync.runtime_registerPool":       ext۰sync۰runtime_registerPool,
	"sync.runtime_Syncsemcheck":       ext۰sync۰runtime_Syncsemcheck,
	"sync/atomic.AddInt32":            ext۰atomic۰AddInt32,
	"sync/atomic.CompareAndSwapInt32": ext۰atomic۰CompareAndSwapInt32,
	"sync/atomic.LoadInt32":           ext۰atomic۰LoadInt32,
	"sync/atomic.LoadUint32":          ext۰atomic۰LoadUint32,
	"sync/atomic.StoreInt32":          ext۰atomic۰StoreInt32,
	"sync/atomic.StoreUint32":         ext۰atomic۰StoreUint32,
	"syscall.Close":                   ext۰syscall۰Close,
	//"syscall.Exit":                    ext۰syscall۰Exit,
	"syscall.Exit":                    ext۰syscall۰Exit,
	"syscall.Fstat":                   ext۰syscall۰Fstat,
	"syscall.Getpid":                  ext۰syscall۰Getpid,
	"syscall.Getwd":                   ext۰syscall۰Getwd,
	"syscall.Kill":                    ext۰syscall۰Kill,
	"syscall.Lstat":                   ext۰syscall۰Lstat,
	"syscall.Open":                    ext۰syscall۰Open,
	"syscall.ParseDirent":             ext۰syscall۰ParseDirent,
	"syscall.RawSyscall":              ext۰syscall۰RawSyscall,
	"syscall.Read":                    ext۰syscall۰Read,
	"syscall.ReadDirent":              ext۰syscall۰ReadDirent,
	"syscall.Stat":                    ext۰syscall۰Stat,
	"syscall.Write":                   ext۰syscall۰Write,
	"time.Sleep":                      ext۰time۰Sleep,
	"time.now":                        ext۰time۰now,
	"github.com/rocky/ssa-interp/trepan.Debug":  ext۰trepan۰Debug,
}

// wrapError returns an interpreted 'error' interface value for err.
func wrapError(err error) Value {
	if err == nil {
		return iface{}
	}
	return iface{t: errorType, v: err.Error()}
}

func ext۰runtime۰Func۰Entry(ff *Frame, args []Value) Value {
	return 0
}

func ext۰runtime۰Func۰FileLine(fr *Frame, args []Value) Value {
	return tuple{"unknown.go", -1}
}

func ext۰runtime۰Func۰Name(fr *Frame, args []Value) Value {
	return "unknown"
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

func ext۰runtime۰Breakpoint(fr *Frame, args []Value) Value {
	// If tracehook is DefaultTraceHook, should we run a PrintStack
	// and leave?
	TraceHook(fr, &fr.block.Instrs[0], ssa2.TRACE_CALL)
	return nil
}

func ext۰runtime۰getgoroot(fr *Frame, args []Value) Value {
	return os.Getenv("GOROOT")
}

func ext۰runtime۰Goexit(fr *Frame, args []Value) Value {
	fr.sourcePanic("Goexit not implemented using panic instead")
	return nil
}

func ext۰strings۰IndexByte(fn *Frame, args []Value) Value {
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

func ext۰sync۰runtime_registerPool(fn *Frame, args []Value) Value {
	return nil
}


func ext۰sync۰runtime_Syncsemcheck(fn *Frame, args []Value) Value {
	return nil
}

func ext۰runtime۰GOMAXPROCS(fr *Frame, args []Value) Value {
	return runtime.GOMAXPROCS(args[0].(int))
}

func ext۰runtime۰GC(fr *Frame, args []Value) Value {
	runtime.GC()
	return nil
}

func ext۰runtime۰Gosched(fr *Frame, args []Value) Value {
	runtime.Gosched()
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

func ext۰runtime۰SetFinalizer(fr *Frame, args []Value) Value {
	return nil // ignore
}

func ext۰runtime۰funcname_go(fr *Frame, args []Value) Value {
	// TODO(adonovan): actually inspect the stack.
	return (*Value)(nil)
	//tuple{0, "somefile.go", 42, true}
	//
	//func FuncForPC(pc uintptr) *Func
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

func ext۰syscall۰Getpid(fr *Frame, args []Value) Value {
	return syscall.Getpid()
}

func valueToBytes(v Value) []byte {
	in := v.([]Value)
	b := make([]byte, len(in))
	for i := range in {
		b[i] = in[i].(byte)
	}
	return b
}

// The set of remaining native functions we need to implement (as needed):

// crypto/aes/cipher_asm.go:10:func hasAsm() bool
// crypto/aes/cipher_asm.go:11:func encryptBlockAsm(nr int, xk *uint32, dst, src *byte)
// crypto/aes/cipher_asm.go:12:func decryptBlockAsm(nr int, xk *uint32, dst, src *byte)
// crypto/aes/cipher_asm.go:13:func expandKeyAsm(nr int, key *byte, enc *uint32, dec *uint32)
// hash/crc32/crc32_amd64.go:12:func haveSSE42() bool
// hash/crc32/crc32_amd64.go:16:func castagnoliSSE42(crc uint32, p []byte) uint32
// math/big/arith_decl.go:8:func mulWW(x, y Word) (z1, z0 Word)
// math/big/arith_decl.go:9:func divWW(x1, x0, y Word) (q, r Word)
// math/big/arith_decl.go:10:func addVV(z, x, y []Word) (c Word)
// math/big/arith_decl.go:11:func subVV(z, x, y []Word) (c Word)
// math/big/arith_decl.go:12:func addVW(z, x []Word, y Word) (c Word)
// math/big/arith_decl.go:13:func subVW(z, x []Word, y Word) (c Word)
// math/big/arith_decl.go:15:func shrVU(z, x []Word, s uint) (c Word)
// math/big/arith_decl.go:16:func mulAddVWW(z, x []Word, y, r Word) (c Word)
// math/big/arith_decl.go:17:func addMulVVW(z, x []Word, y Word) (c Word)
// math/big/arith_decl.go:18:func divWVW(z []Word, xn Word, x []Word, y Word) (r Word)
// math/big/arith_decl.go:19:func bitLen(x Word) (n int)
// os/file_posix.go:14:func sigpipe() // implemented in package runtime
// os/signal/signal_unix.go:15:func signal_enable(uint32)
// os/signal/signal_unix.go:16:func signal_recv() uint32
// runtime/debug.go:13:func LockOSThread()
// runtime/debug.go:17:func UnlockOSThread()
// runtime/debug.go:30:func NumCgoCall() int64
// runtime/debug.go:33:func NumGoroutine() int
// runtime/debug.go:90:func MemProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool)
// runtime/debug.go:114:func ThreadCreateProfile(p []StackRecord) (n int, ok bool)
// runtime/debug.go:122:func GoroutineProfile(p []StackRecord) (n int, ok bool)
// runtime/debug.go:132:func CPUProfile() []byte
// runtime/debug.go:141:func SetCPUProfileRate(hz int)
// runtime/debug.go:149:func SetBlockProfileRate(rate int)
// runtime/debug.go:166:func BlockProfile(p []BlockProfileRecord) (n int, ok bool)
// runtime/debug.go:172:func Stack(buf []byte, all bool) int
// runtime/error.go:81:func typestring(interface{}) string
// runtime/extern.go:68:func funcline_go(*Func, uintptr) (string, int)
// runtime/extern.go:71:func mid() uint32
// runtime/pprof/pprof.go:667:func runtime_cyclesPerSecond() int64
// runtime/race.go:16:func RaceDisable()
// runtime/race.go:19:func RaceEnable()
// runtime/race.go:21:func RaceAcquire(addr unsafe.Pointer)
// runtime/race.go:22:func RaceRelease(addr unsafe.Pointer)
// runtime/race.go:23:func RaceReleaseMerge(addr unsafe.Pointer)
// runtime/race.go:25:func RaceRead(addr unsafe.Pointer)
// runtime/race.go:26:func RaceWrite(addr unsafe.Pointer)
// runtime/race.go:28:func RaceSemacquire(s *uint32)
// runtime/race.go:29:func RaceSemrelease(s *uint32)
// sync/atomic/doc.go:49:func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)
// sync/atomic/doc.go:52:func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)
// sync/atomic/doc.go:55:func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)
// sync/atomic/doc.go:58:func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)
// sync/atomic/doc.go:61:func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)
// sync/atomic/doc.go:67:func AddUint32(addr *uint32, delta uint32) (new uint32)
// sync/atomic/doc.go:70:func AddInt64(addr *int64, delta int64) (new int64)
// sync/atomic/doc.go:73:func AddUint64(addr *uint64, delta uint64) (new uint64)
// sync/atomic/doc.go:76:func AddUintptr(addr *uintptr, delta uintptr) (new uintptr)
// sync/atomic/doc.go:82:func LoadInt64(addr *int64) (val int64)
// sync/atomic/doc.go:88:func LoadUint64(addr *uint64) (val uint64)
// sync/atomic/doc.go:91:func LoadUintptr(addr *uintptr) (val uintptr)
// sync/atomic/doc.go:94:func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)
// sync/atomic/doc.go:100:func StoreInt64(addr *int64, val int64)
// sync/atomic/doc.go:106:func StoreUint64(addr *uint64, val uint64)
// sync/atomic/doc.go:109:func StoreUintptr(addr *uintptr, val uintptr)
// sync/atomic/doc.go:112:func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
// sync/runtime.go:12:func runtime_Semacquire(s *uint32)
// sync/runtime.go:18:func runtime_Semrelease(s *uint32)
// syscall/env_unix.go:30:func setenv_c(k, v string)
// syscall/syscall_linux_amd64.go:60:func Gettimeofday(tv *Timeval) (err error)
// syscall/syscall_linux_amd64.go:61:func Time(t *Time_t) (tt Time_t, err error)
// syscall/syscall_linux_arm.go:28:func Seek(fd int, offset int64, whence int) (newoffset int64, err error)
// time/sleep.go:25:func startTimer(*runtimeTimer)
// time/sleep.go:26:func stopTimer(*runtimeTimer) bool
// time/time.go:758:func now() (sec int64, nsec int32)
// unsafe/unsafe.go:27:func Sizeof(v ArbitraryType) uintptr
// unsafe/unsafe.go:32:func Offsetof(v ArbitraryType) uintptr
// unsafe/unsafe.go:37:func Alignof(v ArbitraryType) uintptr
