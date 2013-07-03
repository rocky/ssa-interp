package main

import "fmt"
import "runtime"

func sub() {
	pc := make([]uintptr, 6, 6)
	count := runtime.Callers(0, pc)
	pc_check, file, line, _ := runtime.Caller(0)
	fmt.Printf("pc(0) via Caller %0x", pc_check)
	fmt.Println(" file", file, "line", line)
	for i:=0; i<count; i++ {
		fmt.Printf("pc[%d]=%0x\n", i, pc[i])
	}
	fmt.Println("Again...")
	count = runtime.Callers(1, pc)
	pc_check, file, line, _ = runtime.Caller(1)
	fmt.Printf("pc(1) via Caller %0x", pc_check)
	fmt.Println(" file", file, "line", line)
	for i:=0; i<count; i++ {
		fmt.Printf("pc[%d]=%0x\n", i, pc[i])
	}
}
func sub1() {
	sub()
}

func sub2() {
	sub1()
}

func main() {
	sub2()
	sub1()
}
