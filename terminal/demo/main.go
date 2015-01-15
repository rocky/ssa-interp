package main

import (
	"fmt"
	"os"

	"github.com/rocky/ssa-interp/terminal"
)

func main() {
	src := []byte(`
/* hello, world! */
var a = 3;

// b is a cool function
function b() {
  return 7;
}`)

	highlighted, err := ansiterm.AsTerm(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

    fmt.Println(string(highlighted))
}
