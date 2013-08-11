package interp

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

// Emulated functions from runtime, some of these are C routines

func runtime۰Gotraceback(fr *Frame) {
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
		fmt.Println(fr.FnAndParamString())
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		line-- // in stack trace, lines are 1-indexed but our array is 0-indexed
		fmt.Printf("\t%s:%d 0x%x\n", file, line, pc)
		fmt.Printf("\t%s\n", debug۰source(lines, line))
	}
}

// Copied almost directly from runtime/debug/stack.go
func runtime۰Stack(fr *Frame, buf []byte) int {
	// As we loop, we open files and read them. These variables record
	// the currently loaded file.
	iobuf := new(bytes.Buffer) // the returned data
	var lines [][]byte
	var lastFile string
	for i := 0; ; i++ {
		pc, file, line, ok := runtime۰Caller(fr, i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(iobuf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		line-- // in stack trace, lines are 1-indexed but our array is 0-indexed
		fmt.Fprintf(iobuf, "\t%s: %s\n", debug۰Function(fr, pc), debug۰source(lines, line))
	}
	newbuf := iobuf.Bytes()
	max := len(newbuf)
	if newmax := len(buf); len(buf) < max {
		max = newmax
		for i:=max-4; i < max-1; i++ { newbuf[i] = '.' }
		newbuf[max-1] = '\n'
	}
	for i := 0; i < max; i++ {
		buf[i] = newbuf[i]
	}
	return len(newbuf)
}
