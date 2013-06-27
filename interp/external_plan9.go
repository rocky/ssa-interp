// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"syscall"

	"ssa-interp"
)

func ext۰syscall۰Close(fr *frame, args []value) value {
	panic("syscall.Close not yet implemented")
}
func ext۰syscall۰Fstat(fr *frame, args []value) value {
	panic("syscall.Fstat not yet implemented")
}
func ext۰syscall۰Kill(fr *frame, args []value) value {
	panic("syscall.Kill not yet implemented")
}
func ext۰syscall۰Lstat(fr *frame, args []value) value {
	panic("syscall.Lstat not yet implemented")
}
func ext۰syscall۰Open(fr *frame, args []value) value {
	panic("syscall.Open not yet implemented")
}
func ext۰syscall۰ParseDirent(fr *frame, args []value) value {
	panic("syscall.ParseDirent not yet implemented")
}
func ext۰syscall۰Read(fr *frame, args []value) value {
	panic("syscall.Read not yet implemented")
}
func ext۰syscall۰ReadDirent(fr *frame, args []value) value {
	panic("syscall.ReadDirent not yet implemented")
}
func ext۰syscall۰Stat(fr *frame, args []value) value {
	panic("syscall.Stat not yet implemented")
}

func ext۰syscall۰Write(fr *frame, args []value) value {
	p := args[1].([]value)
	b := make([]byte, 0, len(p))
	for i := range p {
		b = append(b, p[i].(byte))
	}
	n, err := syscall.Write(args[0].(int), b)
	return tuple{n, wrapError(err)}
}
