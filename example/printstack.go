package main

import "runtime/debug"

func sub() {
	debug.PrintStack()
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
