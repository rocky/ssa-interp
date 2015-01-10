package main

import (
	"fmt"
)

func boom(slice [] string) string {
  return slice[10]
}

func main() {
	slice := []string {"one", "two"}
	fmt.Println(boom(slice))
}
