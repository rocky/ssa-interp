package main

import (
	"syscall"
)

func main() {
	x := syscall.Getuid() // This is what we want to test
	if x - x != 0 {
		panic("Weird arithmetic after syscall.GetUid")
	}
}
