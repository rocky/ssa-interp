package main
import ( "fmt" )

type testEntry struct {
	first, second string
}

func main() {
	record := testEntry{"Hello,", "World!"}
	record2 := record
	fmt.Println(record.first, record2.second)
}
