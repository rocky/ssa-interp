// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"github.com/rocky/ssa-interp"
)

func ext۰syscall۰Close(fr *Frame, args []Value) Value {
	panic("syscall.Close not yet implemented")
}
func ext۰syscall۰Fstat(fr *Frame, args []Value) Value {
	panic("syscall.Fstat not yet implemented")
}
func ext۰syscall۰Kill(fr *Frame, args []Value) Value {
	panic("syscall.Kill not yet implemented")
}
func ext۰syscall۰Lstat(fr *Frame, args []Value) Value {
	panic("syscall.Lstat not yet implemented")
}
func ext۰syscall۰Open(fr *Frame, args []Value) Value {
	panic("syscall.Open not yet implemented")
}
func ext۰syscall۰ParseDirent(fr *Frame, args []Value) Value {
	panic("syscall.ParseDirent not yet implemented")
}
func ext۰syscall۰Read(fr *Frame, args []Value) Value {
	panic("syscall.Read not yet implemented")
}
func ext۰syscall۰ReadDirent(fr *Frame, args []Value) Value {
	panic("syscall.ReadDirent not yet implemented")
}
func ext۰syscall۰Stat(fr *Frame, args []Value) Value {
	panic("syscall.Stat not yet implemented")
}

func ext۰syscall۰Write(fr *Frame, args []Value) Value {
	p := args[1].([]Value)
	b := make([]byte, 0, len(p))
	for i := range p {
		b = append(b, p[i].(byte))
	}
	n, err := syscall.Write(args[0].(int), b)
	return tuple{n, wrapError(err)}
}
func ext۰syscall۰RawSyscall(fn *Frame, args []value) value {
	return tuple{^uintptr(0), uintptr(0), uintptr(0)}
}
