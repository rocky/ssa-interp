// Copyright 2013 Rocky Bernstein.

// info frame
//
// Prints frame information

package gubcmd

import (
	"github.com/rocky/ssa-interp/gub"
)

func init() {
	parent := "info"
	gub.AddSubCommand(parent, &gub.SubcmdInfo{
		Fn: InfoProgramSubcmd,
		Help: `info frame

Prints information about the program including:
*  goroutine number
*  location
*  function and parameter names
`,
		Min_args: 1,
		Max_args: 1,
		Short_help: "Program frame",
		Name: "frame",
	})
}

func InfoFrameSubcmd(args []string) {
	fr := gub.CurFrame()
	gub.Msg("goroutine number: %d", fr.GoNum())
	gub.Msg("location: %s", fr.PositionRange())
	gub.Msg("frame: %s", fr.FnAndParamString())
}
