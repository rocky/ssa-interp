// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import "syscall"

func ext۰syscall۰Close(fr *Frame, args []value) value {
	panic("syscall.Close not yet implemented")
}
func ext۰syscall۰Fstat(fr *Frame, args []value) value {
	panic("syscall.Fstat not yet implemented")
}
func ext۰syscall۰Kill(fr *Frame, args []value) value {
	panic("syscall.Kill not yet implemented")
}
func ext۰syscall۰Lstat(fr *Frame, args []value) value {
	panic("syscall.Lstat not yet implemented")
}
func ext۰syscall۰Open(fr *Frame, args []value) value {
	panic("syscall.Open not yet implemented")
}
func ext۰syscall۰ParseDirent(fr *Frame, args []value) value {
	panic("syscall.ParseDirent not yet implemented")
}
func ext۰syscall۰Read(fr *Frame, args []value) value {
	panic("syscall.Read not yet implemented")
}
func ext۰syscall۰ReadDirent(fr *Frame, args []value) value {
	panic("syscall.ReadDirent not yet implemented")
}
func ext۰syscall۰Stat(fr *Frame, args []value) value {
	panic("syscall.Stat not yet implemented")
}
func ext۰syscall۰Write(fr *Frame, args []value) value {
	// func Write(fd int, p []byte) (n int, err error)
	n, err := write(args[0].(int), valueToBytes(args[1]))
	return tuple{n, wrapError(err)}
}
func ext۰syscall۰RawSyscall(fr *Frame, args []value) value {
	return tuple{^uintptr(0), uintptr(0), uintptr(0)}
}

func syswrite(fd int, b []byte) (int, error) {
	return syscall.Write(fd, b)
}
