package main

import "runtime"

func sub() {
	for i := 0; i<2; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok { panic(i) }
	}
	pc, file, line, ok := runtime.Caller(2)
	if ok { panic(2) }
}

func main() {
	sub()
}
