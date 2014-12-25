// Copyright 2013 Rocky Bernstein.
// Things involving the call frame

package gub

import (
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

var topFrame *interp.Frame
var curFrame *interp.Frame
var curScope *ssa2.Scope

// stackSize is the size of call stack.
var stackSize int

// frameIndex is the frame index we are currently focused on.
var frameIndex int

// MAXSTACKSHOW is maximum number of frame entries to show.
const MAXSTACKSHOW = 50

func CurFrame() *interp.Frame { return curFrame }
func TopFrame() *interp.Frame { return topFrame }
func CurScope() *ssa2.Scope   { return curScope }

func frameInit(fr *interp.Frame) {
	topFrame = fr
	curFrame = fr
	frameIndex = 0
	for stackSize=0; fr !=nil; fr = fr.Caller(0) {
		stackSize++
	}
	curScope = curFrame.Scope()
}

func getFrame(frameNum int, absolutePos bool) (*interp.Frame, int) {
      if absolutePos {
		  if frameNum >= stackSize {
			  Errmsg("Frame number %d too large. Max is %d.",
				  frameNum, stackSize-1)
			  return nil, 0
		  } else if frameNum < -stackSize {
			  Errmsg("Frame number %d too small. Min is %d.",
				  frameNum, -stackSize)
			  return nil, 0
		  }
		  if frameNum < 0 { frameNum += stackSize }
      } else {
		  frameNum += frameIndex
		  if frameNum >= stackSize {
			  Errmsg("Adjusting would put us beyond the oldest frame.")
			  return nil, 0
		  } else if frameNum < 0 {
			  Errmsg("Adjusting would put us beyond the newest frame.")
			  return nil, 0
		  }
      }

	frame := topFrame
	for i:=0; i<frameNum && frame !=nil; frame = frame.Caller(0) {
		i++
	}
	return frame, frameNum
}

func AdjustFrame(frameNum int, absolutePos bool) {
	frame, frameNum := getFrame(frameNum, absolutePos)
	if frame == nil { return }
	curFrame = frame
	frameIndex = frameNum
	event := ssa2.CALL_ENTER
	if (0 == frameIndex) {
		event = TraceEvent
	}
	printLocInfo(curFrame, nil, event)
}

func PrintStack(fr *interp.Frame, count int) {
	if (fr == nil) { return }
	for i:=0; fr !=nil && i < count; fr = fr.Caller(0) {
		pointer := "   "
		if fr == curFrame {
			pointer = "=> "
		}
		fn := fr.Fn()
		if fn.Signature.Recv() != nil {
			Msg("%s#%d %s", pointer, i, fr.FnAndParamString())
		} else {
			Msg("%s#%d %s.%s", pointer, i, fn.Pkg.Object.Path(), fr.FnAndParamString())
		}
		i++
	}
}

func PrintGoroutine(goNum int, goTops []*interp.GoreState) {
	fr := goTops[goNum].Fr
	if fr == nil {
		Msg("Goroutine %d exited", goNum)
		return
	}
	switch fr.Status() {
	case interp.StRunning:
		Section("Goroutine %d", goNum)
		PrintStack(fr, MAXSTACKSHOW)
	case interp.StComplete:
		Msg("Goroutine %d completed", goNum)
	case interp.StPanic:
		Msg("Goroutine %d panic", goNum)
	}
}
