package main
func five() int {
	x := 5
	return x
}
var xx int = five()

func main() {
	if xx > 3 {
		y := xx+1
		println(y)
	}
}
