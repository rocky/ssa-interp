package main

import (
	"fmt"
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
	fmt.Printf("The GCD of %d and %d is %d\n", 5, 3, gcd(5, 3))
}
