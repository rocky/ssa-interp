// See if we can import parts of the interpreter

package main

import (
	"fmt"
	"github.com/rocky/ssa-interp/interp"
)

func main() {
	i := interp.GetInterpreter()
	fmt.Println(i.Program.PackagesByPath)
}
