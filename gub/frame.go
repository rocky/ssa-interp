// Copyright 2013 Rocky Bernstein.
// Things involving the call frame
package gub

import (
	"ssa-interp"
	"ssa-interp/interp"
)

var topFrame *interp.Frame
var curFrame *interp.Frame
var stackSize int  // Size of call stack
var frameIndex int  // frame index we are focused on

func frameInit(fr *interp.Frame) {
	topFrame = fr
	curFrame = fr
	frameIndex = 0
	for stackSize=0; fr !=nil; fr = fr.Caller(0) {
		stackSize++
	}
}

func getFrame(frameNum int, absolutePos bool) (*interp.Frame, int) {
      if absolutePos {
		  if frameNum >= stackSize {
			  errmsg("Frame number %d too large. Max is %d.",
				  frameNum, stackSize-1)
			  return nil, 0
		  } else if frameNum < -stackSize {
			  errmsg("Frame number %d too small. Min is %d.",
				  frameNum, -stackSize)
			  return nil, 0
		  }
		  if frameNum < 0 { frameNum += stackSize }
      } else {
		  frameNum += frameIndex
		  if frameNum >= stackSize {
			  errmsg("Adjusting would put us beyond the oldest frame.")
			  return nil, 0
		  } else if frameNum < 0 {
			  errmsg("Adjusting would put us beyond the newest frame.")
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
		event = traceEvent
	}
	printLocInfo(curFrame, event)
}
