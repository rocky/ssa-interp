// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains tests for Eval.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"code.google.com/p/go.tools/go/types"
	"strings"
)

type testEntry struct {
	src, str string
}

// dup returns a testEntry where both src and str are the same.
func dup(s string) testEntry {
	return testEntry{s, s}
}

var testTypes = []testEntry{
	// basic types
	dup("int"),
	dup("float32"),
	dup("string"),

	// arrays
	dup("[10]int"),

	// slices
	dup("[]int"),
	dup("[][]int"),

	// structs
	dup("struct{}"),
	dup("struct{x int}"),
	{`struct {
		x, y int
		z float32 "foo"
	}`, `struct{x int; y int; z float32 "foo"}`},
	{`struct {
		string
		elems []complex128
	}`, `struct{string; elems []complex128}`},

	// pointers
	dup("*int"),
	dup("***struct{}"),
	dup("*struct{a int; b float32}"),

	// functions
	dup("func()"),
	dup("func(x int)"),
	{"func(x, y int)", "func(x int, y int)"},
	{"func(x, y int, z string)", "func(x int, y int, z string)"},
	dup("func(int)"),
	{"func(int, string, byte)", "func(int, string, byte)"},

	dup("func() int"),
	{"func() (string)", "func() string"},
	dup("func() (u int)"),
	{"func() (u, v int, w string)", "func() (u int, v int, w string)"},

	dup("func(int) string"),
	dup("func(x int) string"),
	dup("func(x int) (u string)"),
	{"func(x, y int) (u string)", "func(x int, y int) (u string)"},

	dup("func(...int) string"),
	dup("func(x ...int) string"),
	dup("func(x ...int) (u string)"),
	{"func(x, y ...int) (u string)", "func(x int, y ...int) (u string)"},

	// interfaces
	dup("interface{}"),
	dup("interface{m()}"),
	dup(`interface{m(int) float32; String() string}`),
	// TODO(gri) add test for interface w/ anonymous field

	// maps
	dup("map[string]int"),
	{"map[struct{x, y int}][]byte", "map[struct{x int; y int}][]byte"},

	// channels
	dup("chan int"),
	dup("chan<- func()"),
	dup("<-chan []func() int"),
}

func main() {
	TestEvalComposite()
	TestEvalArith()
	TestEvalContext()
}

func testEval(pkg *types.Package, scope *types.Scope, str string, typ types.Type,
	typStr, valStr string) {
	gotTyp, gotVal, err := types.Eval(str, pkg, scope)
	if err != nil {
		fmt.Printf("Eval(%q) failed: %s\n", str, err)
		return
	}
	if gotTyp == nil {
		fmt.Printf("Eval(%q) got nil type but no error\n", str)
		return
	}

	// compare types
	if typ != nil {
		// we have a type, check identity
		if !types.IsIdentical(gotTyp, typ) {
			fmt.Printf("Eval(%q) got type %s, want %s\n", str, gotTyp, typ)
			return
		} else {
			fmt.Printf("Eval(%q) got type %s\n", str, gotTyp)
		}
	} else {
		// we have a string, compare type string
		gotStr := gotTyp.String()
		if gotStr != typStr {
			fmt.Printf("Eval(%q) got type %s, want %s\n", str, gotStr, typStr)
			return
		} else {
			fmt.Printf("Eval(%q) got type %s\n", str, gotStr)
		}
	}

	// compare values
	gotStr := ""
	if gotVal != nil {
		gotStr = gotVal.String()
	}
	if gotStr != valStr {
		fmt.Printf("Eval(%q) got value %s, want %s\n", str, gotStr, valStr)
	} else {
		fmt.Printf("Eval(%q) got value '%s'\n", str, gotStr)
	}
}

func TestEvalComposite() {
	for _, test := range testTypes {
		testEval(nil, nil, test.src, nil, test.str, "")
	}
}

func TestEvalArith() {
	var tests = []string{
		`true`,
		`false == false`,
		`12345678 + 87654321 == 99999999`,
		`10 * 20 == 200`,
		`(1<<1000)*2 >> 100 == 2<<900`,
		`"foo" + "bar" == "foobar"`,
		`"abc" <= "bcd"`,
		`len([10]struct{}{}) == 2*5`,
	}
	for _, test := range tests {
		testEval(nil, nil, test, types.Typ[types.UntypedBool], "", "true")
	}
}

func TestEvalContext() {
	src := `
package p
import "fmt"
import m "math"
const c = 3.0
type T []int
func f(a int, s string) float64 {
	type testEntry struct {
		src string
	    num int
	}
	var testTypes = []testEntry{ {"a", 1}, {"b", 2}}
	const d int = c + 1
	var x int
	x = a + len(s) + testTypes[0].num
	return float64(x)
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p", src, 0)
	if err != nil {
		panic(err)
	}

	pkg, err := types.Check("p", fset, []*ast.File{file})
	if err != nil {
		panic(err)
	}

	pkgScope := pkg.Scope()
	if n := pkgScope.NumChildren(); n != 1 {
		str := fmt.Sprintf("got %d file scopes, want 1", n)
		panic(str)
	}

	fileScope := pkgScope.Child(0)
	if n := fileScope.NumChildren(); n != 1 {
		str := fmt.Sprintf("got %d functions scopes, want 1", n)
		panic(str)
	}

	funcScope := fileScope.Child(0)

	var tests = []string{
		`true => true, untyped boolean`,
		`fmt.Println => , func(a·3 ...interface{}) (n·1 int, err·2 error)`,
		`c => 3, untyped float`,
		`T => , p.T`,
		`a => , int`,
		`s => , string`,
		`d => 4, int`,
		`x => , int`,
		`d/c => 1, int`,
		`c/2 => 3/2, untyped float`,
		`m.Pi < m.E => false, untyped boolean`,
		`testTypes[0].num => 1, int`,
	}
	for _, test := range tests {
		str, typ := split(test, ", ")
		str, val := split(str, "=>")
		testEval(pkg, funcScope, str, nil, typ, val)
	}
}

// split splits string s at the first occurrence of s.
func split(s, sep string) (string, string) {
	i := strings.Index(s, sep)
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+len(sep):])
}
