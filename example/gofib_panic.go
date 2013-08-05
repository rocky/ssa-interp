package main

import "fmt"

const limit = 15
// Recursion.
func fib(x int) int {
	if x < 2 {
		return x
	}
	if x > 13 {
		panic("Too large")
	}
	return fib(x-1) + fib(x-2)
}

func fibgen(ch chan int) {
	for x := 0; x < limit; x++ {
		ch <- fib(x)
	}
	close(ch)
}

// Goroutines and channels.
func main() {
	ch := make(chan int)
	go fibgen(ch)
	var fibs []int
	for v := range ch {
		fibs = append(fibs, v)
		if len(fibs) == limit {
			break
		}
	}
	fmt.Printf("First %d Fibonacci numbers:\n", limit)
	fmt.Println(fibs)
}
