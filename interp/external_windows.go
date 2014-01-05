// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

import (
	"github.com/rocky/ssa-interp"
)

func ext۰syscall۰Close(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Close not yet implemented")
}
func ext۰syscall۰Fstat(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Fstat not yet implemented")
}
func ext۰syscall۰Kill(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Kill not yet implemented")
}
func ext۰syscall۰Lstat(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Lstat not yet implemented")
}
func ext۰syscall۰Open(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Open not yet implemented")
}
func ext۰syscall۰ParseDirent(fr *frame, args []value) value {
	fr.sourcePanic("syscall.ParseDirent not yet implemented")
}
func ext۰syscall۰Read(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Read not yet implemented")
}
func ext۰syscall۰ReadDirent(fr *frame, args []value) value {
	fr.sourcePanic("syscall.ReadDirent not yet implemented")
}
func ext۰syscall۰Stat(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Stat not yet implemented")
}
func ext۰syscall۰Write(fr *frame, args []value) value {
	fr.sourcePanic("syscall.Write not yet implemented")
}
func ext۰syscall۰RawSyscall(fn *ssa.Function, args []value) value {
	return tuple{uintptr(0), uintptr(0), uintptr(syscall.ENOSYS)}
}
