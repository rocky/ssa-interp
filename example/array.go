package main

import "os"

var a = []int{1,2,3,4}

func foo(i int, s string) int {
	var a = []int{3,4,5}
	if s == "1" {
		i++
	}
	return a[i]
}

func main() {
	var b = []string{"1", "2", "3"}
	rc := foo(0, b[0])
	os.Exit(a[1]+ rc)
}
