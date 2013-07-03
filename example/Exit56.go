package main

import (
	"os"
	"flag"
	"syscall"
)

func main() {
	var useSys = flag.Bool("syscall", false, "Use syscall.Exit() rather than os.Exit()")
	flag.Parse()
	if *useSys {
		syscall.Exit(5)
	} else {
		os.Exit(6)
	}
}
