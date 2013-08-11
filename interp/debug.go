package interp

import (
	"bytes"
	"fmt"
)

// Emulated functions from runtime/debug.go
func debug۰PrintStack(fr *Frame) {
	buf := make([]byte, 5000) // the returned data
	n := runtime۰Stack(fr, buf)
	fmt.Printf(string(buf[:n]))
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
