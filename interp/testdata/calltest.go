package main

import "fmt"
import "runtime"

func sub() {
	for i := 0; i<10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if ok {
			fmt.Println(i, "file", file, "line", line, "pc", pc, "ok", ok)
		} else {
			fmt.Println(i, "not ok")
			break
		}
	}
}

func main() {
	sub()
}
