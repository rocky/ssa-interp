package interp

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

// Emulated functions from runtime/debug.go

// Copied almost directly from runtime/debug/stack.go
func debug۰PrintStack(fr *Frame) {
	// As we loop, we open files and read them. These variables record
	// the currently loaded file.
	var lines [][]byte
	var lastFile string
	for i := 0; ; i++ {
		pc, file, line, ok := runtime۰Caller(fr, i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Printf("%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		line-- // in stack trace, lines are 1-indexed but our array is 0-indexed
		fmt.Printf("\t%s: %s\n", debug۰Function(fr, pc), debug۰source(lines, line))
	}
}

// source returns a space-trimmed slice of the n'th line.
// Copied almost directly from runtime/debug/stack.go
func debug۰source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return []byte("???")
	}
	return bytes.Trim(lines[n], " \t")
}


func debug۰Function(fr *Frame, pc uintptr) []byte {
	fnIndex := pc >> 16
	fn := num2fnMap[fnIndex - 1]
	if fn == nil {
		return []byte("??Unknown fn")
	}
	return []byte(fn.Name())
}
