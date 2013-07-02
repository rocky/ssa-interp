package main

import "fmt"
import "runtime"

func sub() {
	var pc []uintptr
	count := runtime.Callers(0, pc)
	fmt.Println("count is", count)
	count := runtime.Callers(1, pc)
	fmt.Println("count is", count)
}

func main() {
	sub()
}
