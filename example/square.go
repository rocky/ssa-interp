package main

import "os"

func square(n int) (sum int) {
	sum = 0
	odd := 1
	for i := 0; i<n; i++ {
		sum += odd
		odd += 2
	}
	return sum
}

func main() {
	rc := square(4)
	os.Exit(rc)
}
