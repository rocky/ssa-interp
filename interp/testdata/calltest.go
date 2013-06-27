package main

import "fmt"
import "runtime"

func sub() {
	pc, file, line, ok := runtime.Caller(0)
	fmt.Println("file", file, "line", line, "pc", pc, "ok", ok)
}

func main() {
	sub()
}
