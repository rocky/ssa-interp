// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file tests types.Check by using it to
// typecheck the standard library.

package main

import (
	// "flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"code.google.com/p/go.tools/go/types"
	"github.com/rocky/ssa-interp"
)

// var verbose = flag.Bool("types.v", false, "verbose mode")
var verbose = true
const testing_short = true

var (
	pkgCount int // number of packages processed
	start    = time.Now()
)

func main() {
	walkDirs(filepath.Join(runtime.GOROOT(), "src/pkg/bufio"))
	if verbose {
		fmt.Println(pkgCount, "packages typechecked in", time.Since(start))
	}
}

// Package paths of excluded packages.
var excluded = map[string]bool{
	"builtin": true,
}

// typecheck typechecks the given package files.
func typecheck(path string, filenames []string) {
	fset := token.NewFileSet()

	// parse package files
	var files []*ast.File
	for _, filename := range filenames {
		file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
		if err != nil {
			// the parser error may be a list of individual errors; report them all
			if list, ok := err.(scanner.ErrorList); ok {
				for _, err := range list {
					fmt.Println(err)
				}
				return
			}
			fmt.Println(err)
			return
		}

		if verbose {
			if len(files) == 0 {
				fmt.Println("package", file.Name.Name)
			}
			fmt.Println("\t", filename)
		}

		files = append(files, file)
	}

	// typecheck package files
	var conf types.Config
	conf.Error = func(err error) { fmt.Println(err) }
	pkg, err := conf.Check(path, fset, files, nil)
	if err == nil {
		scope := pkg.Scope()
		assignScopeNums(ast2Scope, scope)
		traverseScope(fset, pkg.Scope(), 0)
	}
	pkgCount++
}

type Scope struct {
	*types.Scope
	scopeNum int
}

var scopeNum int = 0
var num2scope [] *types.Scope

var  ast2Scope map[ast.Node]*Scope = make(map[ast.Node]*Scope)

func assignScopeNums(ast2Scope map[ast.Node]*Scope, scope *types.Scope) {
	num2scope = append(num2scope, scope)
	ast2Scope[scope.Node()] = &Scope {
		Scope: scope,
		scopeNum: scopeNum,
	}
	scopeNum++
	n := scope.NumChildren()
	for i:=0; i<n; i++ {
		child := scope.Child(i)
		if child != nil { assignScopeNums(ast2Scope, child) }
	}
}

func printScope(fset *token.FileSet, scope *Scope) {
	node := scope.Node()
	if node != nil {
		startP := fset.Position(node.Pos())
		endP   := fset.Position(node.End())
		fmt.Println(ssa2.PositionRange(startP, endP))
	}
	fmt.Printf("#%d %s\n", scope.scopeNum, scope.Scope)
}

func traverseScope(fset *token.FileSet, scope *types.Scope, indent int) {
	const ind = ".  "
	indn  := strings.Repeat(ind, indent)
	fmt.Printf("%s ", indn)
	node  := scope.Node()
	printScope(fset, ast2Scope[node])

	n := scope.NumChildren()
	for i:=0; i<n; i++ {
		child := scope.Child(i)
		if child != nil { traverseScope(fset, child, indent+1) }
	}
}

// pkgfiles returns the list of package files for the given directory.
func pkgfiles(dir string) []string {
	ctxt := build.Default
	ctxt.CgoEnabled = false
	pkg, err := ctxt.ImportDir(dir, 0)
	if err != nil {
		if _, nogo := err.(*build.NoGoError); !nogo {
			fmt.Println(err)
		}
		return nil
	}
	if excluded[pkg.ImportPath] {
		return nil
	}
	var filenames []string
	for _, name := range pkg.GoFiles {
		filenames = append(filenames, filepath.Join(pkg.Dir, name))
	}
	for _, name := range pkg.TestGoFiles {
		filenames = append(filenames, filepath.Join(pkg.Dir, name))
	}
	return filenames
}

// Note: Could use filepath.Walk instead of walkDirs but that wouldn't
//       necessarily be shorter or clearer after adding the code to
//       terminate early for -short tests.

func walkDirs(dir string) {
	// limit run time for short tests
	if testing_short && time.Since(start) >= 750*time.Millisecond {
		return
	}

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	// typecheck package in directory
	if files := pkgfiles(dir); files != nil {
		typecheck(dir, files)
	}

	// traverse subdirectories, but don't walk into testdata
	for _, fi := range fis {
		if fi.IsDir() && fi.Name() != "testdata" {
			walkDirs(filepath.Join(dir, fi.Name()))
		}
	}
}
