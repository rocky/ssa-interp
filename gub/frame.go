// Copyright 2013 Rocky Bernstein.
// Things involving the call frame
package gub

import (
	"ssa-interp/interp"
)

var topFrame *interp.Frame
var curFrame *interp.Frame
var stackSize int

func frameInit(fr *interp.Frame) {
	topFrame = fr
	curFrame = fr
	for stackSize=0; fr !=nil; fr = fr.Caller(0) {
		stackSize++
	}

}
