package interp
import (
	"fmt"
	"os"
	"io"
	"github.com/rocky/ssa-interp"
)

var fn2NumMap map[*ssa2.Function]uint  = make(map[*ssa2.Function]uint, 0)
var num2fnMap []*ssa2.Function
var lastFn2Num uint = 1

func Externals() map[string]externalFn {
	return externals
}


func fn2Num(fn *ssa2.Function) uint {
	if fn2NumMap[fn] != 0 {
		return fn2NumMap[fn]
	}
	fn2NumMap[fn] = lastFn2Num
	num2fnMap = append(num2fnMap, fn)
	lastFn2Num++
	return fn2NumMap[fn]
}

func byteAry2ValueAry(ary[] byte) []Value {
	var result []Value
	for _, word := range ary {
		result = append(result, word)
	}
	return result
}

func ValueAry2byteAry(ary[] Value) []byte {
	var result []byte
	for _, word := range ary {
		result = append(result, word.(byte))
	}
	return result
}

func ext۰os۰Exit(fr *Frame, args []Value) Value {
	msg := fmt.Sprintf("exit status %d", args[0].(int))
	io.WriteString(os.Stderr, msg)
	io.WriteString(os.Stderr, "\n")
	// os.Exit works even if it doesn't allow cleanup as I suppose
	// exitPanic might.
	os.Exit(args[0].(int))
	// This doesn't seem to work. We leave it uncommented
	// to make go's return value checking happy.
	panic(exitPanic(args[0].(int)))
}

func ext۰debug۰PrintStack(fr *Frame, args []Value) Value {
	debug۰PrintStack(fr)
	return nil
}

func ext۰runtime۰Stack(fr *Frame, args []Value) Value {
	bufVal := args[0].([]Value)
	buf := ValueAry2byteAry(bufVal)
	n := runtime۰Stack(fr, buf)
	for i, _ := range buf {
		bufVal[i] = buf[i]
	}
	return n
}

// FIXME: this isn't used because it is internally called
// from runtime/stack. But it should be publically callable
// want
func ext۰debug۰function(fr *Frame, args []Value) Value {
	pc := args[0].(uintptr)
	return debug۰Function(fr, pc)
}

// Turn the function name, block number and instruction inside the
// block into number that we'll use as a PC. If it so happens that
// there are more than 256 instructions in a block or more than 256
// basic blocks in a function, the PC is not unique. In cases that the
// block number is not unique, one could try to disambiguate blocks by
// discarding those that have fewer instructions than the instruction
// number. However either way, I think I can live with this
// limitation. Note: I tried using uint64 and 24 bits, but this causes
// a range error down the line on 32-bit linux, I think when casting
// to a uintptr.
func EncodePC(fr *Frame) uint {
	if fr == nil { return 0xbadbad }
	fnNum := fn2Num(fr.fn)
	// We don't always have basic blocks. Sigh
	blockIndex := 0xff
	if fr.block != nil { blockIndex = fr.block.Index }
	bpc := uint(blockIndex << 8) + uint(fr.pc & 0xff)
	return uint(fnNum << 16) | (bpc & 0x00ffff)
}

func runtime۰Caller(fr *Frame, skip int) (pc uintptr, file string, line int, ok bool) {
	final_fr := fr
	for i:=0; i<skip; i++ {
		final_fr = final_fr.caller
		if final_fr == nil {
			return 0, "None", 0, false
		}
	}

	fset := fr.fn.Prog.Fset
	startP := fset.Position(final_fr.startP)

	var filename string
	if startP.IsValid() {
		filename = startP.Filename
	} else {
		filename = "??"
	}
	pc = uintptr(EncodePC(final_fr))
	line = startP.Line
	return pc, filename, line, true
}

func ext۰runtime۰Caller(fr *Frame, args []Value) Value {
	skip := args[0].(int)
	pc, filename, line, ok := runtime۰Caller(fr, skip)
	return tuple{pc, filename, line, ok}
}

func ext۰runtime۰Callers(fr *Frame, args []Value) Value {
	skip := args[0].(int)
	pc   := args[1].([]Value)
	size := len(pc)

	for i:=0; i<=skip; i++ {
		fr = fr.caller
		if fr == nil {
			return 0
		}
	}
	var count int
	for count = 0; fr != nil && count <= size; fr = fr.caller {
		pc[count] = EncodePC(fr)
		count++
	}
	return count
}

// // Can't really write this using runtime.function because interperter
// // can't cant copy return value to *Func.
// func ext۰runtime۰FuncForPC(fr *Frame, args []Value) Value {
// 	return nil
// }
