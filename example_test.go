// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa2_test

import (
	"fmt"
	"os"

	"github.com/rocky/go-loader"
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

const message = "Hello, World!"

func main() {
	fmt.Println(message)
}
`
	var conf loader.Config

	// Parse the input file.
	file, err := conf.ParseFile("hello.go", hello)
	if err != nil {
		fmt.Print(err) // parse error
		return
	}

	// Create single-file main package.
	conf.CreateFromFiles("main", file)

	// Load the main package and its dependencies.
	iprog, err := conf.Load()
	if err != nil {
		fmt.Print(err) // type error in some package
		return
	}

	// Create SSA-form program representation.
	prog := ssa2.Create(iprog, ssa2.SanityCheckFunctions)
	mainPkg := prog.Package(iprog.Created[0].Pkg)

	// Print out the package.
	mainPkg.WriteTo(os.Stdout)

	// Build SSA code for bodies of functions in mainPkg.
	mainPkg.Build()

	// Print out the package-level functions.
	mainPkg.Func("init").WriteTo(os.Stdout)
	mainPkg.Func("main").WriteTo(os.Stdout)

	// Output:
	//
	// package main:
	//   func  init       func()
	//   var   init$guard bool
	//   func  main       func()
	//   const message    message = "Hello, World!":untyped string
	//
	// # Name: main.init
	// # Package: main
	// # Synthetic: package initializer
	// func init():
	// # scope: 0
	// 0:                                                                entry P:0 S:2
	// 0	t0 = *init$guard                                                   bool
	// 1	if t0 goto 2 else 1
	// # scope: 0
	// 1:                                                           init.start P:1 S:1
	// 0	*init$guard = true:bool
	// 1	t1 = fmt.init()                                                      ()
	// 2	jump 2
	// # scope: 0
	// 2:                                                            init.done P:2 S:0
	// 0	return
	//
	// # Name: main.main
	// # Package: main
	// # Location: hello.go:8:6
	// func main():
	// # scope: 3
	// 0:                                                                entry P:0 S:0
	// 0	trace <STATEMENT in list> at hello.go:9:2-22
	// 1	t0 = new [1]interface{} (varargs)                       *[1]interface{}
	// 2	t1 = &t0[0:int]                                            *interface{}
	// 3	t2 = make interface{} <- string ("Hello, World!":string)    interface{}
	// 4	*t1 = t2
	// 5	t3 = slice t0[:]                                          []interface{}
	// 6	t4 = fmt.Println(t3...)                              (n int, err error)
	// 7	trace <Block End> at hello.go:10:2
	// 8	return
}
