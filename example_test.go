// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa2_test

import (
	"fmt"
	"go/build"
	"go/parser"
	"os"

	"code.google.com/p/go.tools/importer"
	"github.com/rocky/ssa-interp"
)

// This program demonstrates how to run the SSA builder on a "Hello,
// World!" program and shows the printed representation of packages,
// functions and instructions.
//
// Within the function listing, the name of each BasicBlock such as
// ".0.entry" is printed left-aligned, followed by the block's
// Instructions.
//
// For each instruction that defines an SSA virtual register
// (i.e. implements Value), the type of that value is shown in the
// right column.
//
// Build and run the tortoise.go program in this package if you want a
// standalone tool with similar functionality.
//
func Example() {
	const hello = `
package main

import "fmt"

const message = "Hello, World"

func bang() string {
   s := "!"
   return s
}

func main() {
    if str := bang(); len(str) > 0 {
	  fmt.Println(message, str)
   }
}
`
	// Construct an importer.  Imports will be loaded as if by 'go build'.
	imp := importer.New(&importer.Config{Build: &build.Default})

	// Parse the input file.
	file, err := parser.ParseFile(imp.Fset, "hello.go", hello, 0)
	if err != nil {
		fmt.Print(err) // parse error
		return
	}

	// Create single-file main package and import its dependencies.
	mainInfo := imp.CreatePackage("main", file)

	// Create SSA-form program representation.
	var mode ssa2.BuilderMode
	prog := ssa2.NewProgram(imp.Fset, mode)
	if err := prog.CreatePackages(imp); err != nil {
		fmt.Print(err) // type error in some package
		return
	}
	mainPkg := prog.Package(mainInfo.Pkg)

	// Print out the package.
	mainPkg.DumpTo(os.Stdout)

	// Build SSA code for bodies of functions in mainPkg.
	mainPkg.Build()

	// Print out the package-level functions.
	mainPkg.Func("init").DumpTo(os.Stdout)
	mainPkg.Func("main").DumpTo(os.Stdout)

	// Output:
	// package main:
	//   func  bang       func() string
	//   func  init       func()
	//   var   init$guard bool
	//   func  main       func()
	//   const message    message = "Hello, World":untyped string
	//
	// # Name: main.init
	// # Package: main
	// # Synthetic: package initializer
	// func init():
	// # scope: 0
	// .0.entry:                                                               P:0 S:2
	// 	t0 = *init$guard                                                   bool
	// 	if t0 goto 2.init.done else 1.init.start
	// # scope: 0
	// .1.init.start:                                                          P:1 S:1
	// 	*init$guard = true:bool
	// 	t1 = fmt.init()                                                      ()
	// 	jump 2.init.done
	// # scope: 0
	// .2.init.done:                                                           P:2 S:0
	// 	return
	//
	// # Name: main.main
	// # Package: main
	// # Location: hello.go:13:6-17:2
	// func main():
	// # scope: 4
	// .0.entry:                                                               P:0 S:2
	// 	trace <IF initialize> at hello.go:14:8-21
	// 	t0 = bang()                                                      string
	// 	trace <IF expression> at hello.go:14:23-35
	// 	t1 = len(t0)                                                        int
	// 	t2 = t1 > 0:int                                                    bool
	// 	if t2 goto 1.if.then else 2.if.done
	// # scope: 6
	// .1.if.then:                                                             P:1 S:1
	// 	trace <STATEMENT in list> at hello.go:15:4-29
	// 	t3 = new [2]interface{} (varargs)                       *[2]interface{}
	// 	t4 = &t3[0:untyped integer]                                *interface{}
	// 	t5 = make interface{} <- string ("Hello, World":string)     interface{}
	// 	*t4 = t5
	// 	t6 = &t3[1:untyped integer]                                *interface{}
	// 	t7 = make interface{} <- string (t0)                        interface{}
	// 	*t6 = t7
	// 	t8 = slice t3[:]                                          []interface{}
	// 	t9 = fmt.Println(t8)                                 (n int, err error)
	// 	trace <Block End> at hello.go:16:5
	// 	jump 2.if.done
	// .2.if.done:                                                             P:2 S:0
	// 	trace <Block End> at hello.go:17:2
	// 	return
}
