package gub_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const slash = string(os.PathSeparator)

// These are files in ssa-interp/gub/testdata/.
type testDatum struct {
	gofile  string
	baseName string
}

// Note we should order these from simple to more complex
var testData = []testDatum {
	{gofile: "gcd",    baseName: "stepping"},
	{gofile: "panic",  baseName: "panic"},
	{gofile: "gcd",    baseName: "frame"},
	{gofile: "expr",   baseName: "eval"},
	{gofile: "struct", baseName: "struct"},
}

// Runs debugger on go program with baseName. Then compares output.
func run(t *testing.T, test testDatum) bool {
	fmt.Printf("Input: %s on %s.go\n", test.baseName, test.gofile)

	goFile    := fmt.Sprintf("testdata%s%s.go",  slash, test.gofile)
	rightName := fmt.Sprintf("testdata%s%s.right",  slash, test.baseName)
	gubOpt    := fmt.Sprintf("-gub=-cmdfile=testdata%s%s.cmd", slash, test.baseName)

	file, err := os.Open(goFile) // For read access.
	file.Close()
	if err != nil {
		log.Fatal(err)
	}

	got, err  := exec.Command("../tortoise", "-run", "-interp=S", gubOpt, goFile).Output()

	if err != nil {
		fmt.Printf("%s", got)
		log.Fatal(err)
	}

	var rightFile *os.File
	rightFile, err = os.Open(rightName) // For read access.
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 5000)
	count, err := rightFile.Read(data)
	if err != nil {
		t.Errorf("%s failed to read 'right' data file %s:", test.baseName, rightFile)
		log.Fatal(err)
	}
	want := string(data[0:count])
	if string(got) != want {
		gotName := fmt.Sprintf("testdata%s%s.got",  slash, test.baseName)
		gotLines  := strings.Split(string(got), "\n")
		wantLines := strings.Split(string(want), "\n")
		wantLen   := len(wantLines)
		for i, line := range(gotLines) {
			if i == wantLen {
				fmt.Println("want results are shorter than got results, line", i+1)
				break
			}
			if line != wantLines[i] {
				fmt.Println("results differ starting at line", i+1)
				fmt.Println("got:\n", line)
				fmt.Println("want:\n", wantLines[i])
				break
			}
		}
		if err := ioutil.WriteFile(gotName, got, 0666); err == nil {
			fmt.Printf("Full results are in file %s\n", gotName)
		}
		t.Errorf("%s failed:", test.baseName)
	}

	// Print a helpful hint if we don't make it to the end.
	hint := "Run manually"
	defer func() {
		if hint != "" {
			fmt.Println("FAIL")
			fmt.Println(hint)
		} else {
			fmt.Println("PASS")
		}
	}()

	hint = "" // call off the hounds
	return true
}

// TestInterp runs the debugger on a selection of small Go programs.
func TestInterp(t *testing.T) {

	var failures []string

	// The panic test assumes GOTRACEBACK != 0. I think we
	// want tracebacks anyway.
	os.Setenv("GOTRACEBACK", "2")

	for _, test := range testData {
		if !run(t, test) {
			failures = append(failures, test.baseName)
		}
	}

	if failures != nil {
		fmt.Println("The following tests failed:")
		for _, f := range failures {
			fmt.Printf("\t%s\n", f)
		}
	}

}
