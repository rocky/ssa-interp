package main

import "runtime"

func caller() {
	//FIXME add a test to make sure file, pc from Callers matches
	// the equivalent from Caller.
	for i := 0; i<2; i++ {
		_, _, _, ok := runtime.Caller(i)
		if !ok { panic(i) }
	}
	_, _, _, ok := runtime.Caller(20)
	if ok { panic(3) }
	pcA := make([]uintptr, 6, 6)
	count := runtime.Callers(0, pcA)
	pcB := make([]uintptr, 6, 6)
	countM1 := runtime.Callers(1, pcB)
	if count -1 != countM1 { panic(5) }
	for i := 1; i<countM1-1; i++ {
		if pcA[i+1] != pcB[i] { panic(100+i) }
	}
}

func main() {
	caller()
}
