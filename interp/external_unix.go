// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows,!plan9

package interp

import (
	"syscall"
)

func ValueToBytes(v Value) []byte {
	in := v.([]Value)
	b := make([]byte, len(in))
	for i := range in {
		b[i] = in[i].(byte)
	}
	return b
}

func fillStat(st *syscall.Stat_t, stat structure) {
	stat.fields[0] = st.Dev
	stat.fields[1] = st.Ino
	stat.fields[2] = st.Nlink
	stat.fields[3] = st.Mode
	stat.fields[4] = st.Uid
	stat.fields[5] = st.Gid

	stat.fields[7] = st.Rdev
	stat.fields[8] = st.Size
	stat.fields[9] = st.Blksize
	stat.fields[10] = st.Blocks
	// TODO(adonovan): fix: copy Timespecs.
	// stat.fields[11] = st.Atim
	// stat.fields[12] = st.Mtim
	// stat.fields[13] = st.Ctim
}

func ext۰syscall۰Close(fr *Frame, args []Value) Value {
	// func Close(fd int) (err error)
	return wrapError(syscall.Close(args[0].(int)))
}

func ext۰syscall۰Fstat(fr *Frame, args []Value) Value {
	// func Fstat(fd int, stat *Stat_t) (err error)
	fd := args[0].(int)
	stat := (*args[1].(*Value)).(structure)

	var st syscall.Stat_t
	err := syscall.Fstat(fd, &st)
	fillStat(&st, stat)
	return wrapError(err)
}

func ext۰syscall۰ReadDirent(fr *Frame, args []Value) Value {
	// func ReadDirent(fd int, buf []byte) (n int, err error)
	fd := args[0].(int)
	p := args[1].([]Value)
	b := make([]byte, len(p))
	n, err := syscall.ReadDirent(fd, b)
	for i := 0; i < n; i++ {
		p[i] = b[i]
	}
	return tuple{n, wrapError(err)}
}

func ext۰syscall۰Kill(fr *Frame, args []Value) Value {
	// func Kill(pid int, sig Signal) (err error)
	return wrapError(syscall.Kill(args[0].(int), syscall.Signal(args[1].(int))))
}

func ext۰syscall۰Lstat(fr *Frame, args []Value) Value {
	// func Lstat(name string, stat *Stat_t) (err error)
	name := args[0].(string)
	stat := (*args[1].(*Value)).(structure)

	var st syscall.Stat_t
	err := syscall.Lstat(name, &st)
	fillStat(&st, stat)
	return wrapError(err)
}

func ext۰syscall۰Open(fr *Frame, args []Value) Value {
	// func Open(path string, mode int, perm uint32) (fd int, err error) {
	path := args[0].(string)
	mode := args[1].(int)
	perm := args[2].(uint32)
	fd, err := syscall.Open(path, mode, perm)
	return tuple{fd, wrapError(err)}
}

func ext۰syscall۰ParseDirent(fr *Frame, args []Value) Value {
	// func ParseDirent(buf []byte, max int, names []string) (consumed int, count int, newnames []string)
	max := args[1].(int)
	var names []string
	for _, iname := range args[2].([]Value) {
		names = append(names, iname.(string))
	}
	consumed, count, newnames := syscall.ParseDirent(ValueToBytes(args[0]), max, names)
	var inewnames []Value
	for _, newname := range newnames {
		inewnames = append(inewnames, newname)
	}
	return tuple{consumed, count, inewnames}
}

func ext۰syscall۰Read(fr *Frame, args []Value) Value {
	// func Read(fd int, p []byte) (n int, err error)
	fd := args[0].(int)
	p := args[1].([]Value)
	b := make([]byte, len(p))
	n, err := syscall.Read(fd, b)
	for i := 0; i < n; i++ {
		p[i] = b[i]
	}
	return tuple{n, wrapError(err)}
}

func ext۰syscall۰Stat(fr *Frame, args []Value) Value {
	// func Stat(name string, stat *Stat_t) (err error)
	name := args[0].(string)
	stat := (*args[1].(*Value)).(structure)

	var st syscall.Stat_t
	err := syscall.Stat(name, &st)
	fillStat(&st, stat)
	return wrapError(err)
}

func ext۰syscall۰Write(fr *Frame, args []Value) Value {
	// func Write(fd int, p []byte) (n int, err error)
	n, err := syscall.Write(args[0].(int), ValueToBytes(args[1]))
	return tuple{n, wrapError(err)}
}

func ext۰syscall۰RawSyscall(fn *Frame, args []Value) Value {
	return tuple{uintptr(0), uintptr(0), uintptr(syscall.ENOSYS)}
}
