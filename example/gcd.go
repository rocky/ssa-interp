package main

import (
	"fmt"
	"os"
	"strconv"
)

// GCD. We assume positive numbers
func gcd(a int, b int) int {
  // Make: a <= b
  if a > b {
    a, b = b, a
  }

  if a <= 0 { return -1 }

  if a == 1 || b-a == 0 {
    return a
  }
  return gcd(b-a, a)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s *int1* *int2*\n", os.Args[0])
		fmt.Println("\t prints Greatest Common Divisor of *int1* and *int2*");
		os.Exit(1)
	}

	var err error
	var a, b int
	if a, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Println(err)
		os.Exit(2)
}

	if b, err = strconv.Atoi(os.Args[2]); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Printf("The GCD of %d and %d is %d\n", a, b, gcd(a, b))
}
