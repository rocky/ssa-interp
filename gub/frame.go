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
var stackSize int  // Size of call stack
var frameIndex int  // frame index we are focused on
const MAXSTACKSHOW = 50  // maximum number of frame entries to show

func CurFrame() *interp.Frame { return curFrame }
func TopFrame() *interp.Frame { return topFrame }

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

func adjustFrame(frameNum int, absolutePos bool) {
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
