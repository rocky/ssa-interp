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
	"time"
	"github.com/rocky/go-types"
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
	conf.Check(path, fset, files, nil)
	pkgCount++
}

type Scope struct {
	*types.Scope
	scopeNum int
}

var num2scope [] *types.Scope

var  ast2Scope map[ast.Node]*Scope = make(map[ast.Node]*Scope)

func PrintAstType(node ast.Node) {
	switch n := node.(type) {
	case *ast.Comment:
		fmt.Println("Comment")
	case *ast.CommentGroup:
		fmt.Println("Comment Group")
	case *ast.Field:
		fmt.Println("Field")
	case *ast.FieldList:
		fmt.Println("FieldList")
		// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		fmt.Println("BadExpr, Ident or BasicLit")
		// nothing to do
	case *ast.Ellipsis:
		fmt.Println("Elipsis")
	case *ast.FuncLit:
		fmt.Println("FuncLit")
	case *ast.CompositeLit:
		fmt.Println("CompositeLit")
	case *ast.ParenExpr:
		fmt.Println("ParenExpr")
	case *ast.SelectorExpr:
		fmt.Println("SelectorExpr")
	case *ast.IndexExpr:
		fmt.Println("IndexExpr")
	case *ast.SliceExpr:
		fmt.Println("SliceExpr")
	case *ast.TypeAssertExpr:
		fmt.Println("TypeAssertExpr")
	case *ast.CallExpr:
		fmt.Println("CallExpr")
	case *ast.StarExpr:
		fmt.Println("StarExpr")
	case *ast.UnaryExpr:
		fmt.Println("UnaryExpr")
	case *ast.BinaryExpr:
		fmt.Println("BinaryExpr")
	case *ast.KeyValueExpr:
		fmt.Println("KeyValueExpr")
		// Types
	case *ast.ArrayType:
		fmt.Println("ArrayType")
	case *ast.StructType:
		fmt.Println("StructType")
	case *ast.FuncType:
		fmt.Println("FuncType")
	case *ast.InterfaceType:
		fmt.Println("InterfaceType")
	case *ast.MapType:
		fmt.Println("MapType")
	case *ast.ChanType:
		fmt.Println("ChanType")
		// Statements
	case *ast.BadStmt:
		fmt.Println("BadStmt")
	case *ast.DeclStmt:
		fmt.Println("DeclStmt")
	case *ast.EmptyStmt:
		fmt.Println("EmptyStmt")
	case *ast.LabeledStmt:
		fmt.Println("LabeledStmt")
	case *ast.ExprStmt:
		fmt.Println("ExprStmt")
	case *ast.SendStmt:
		fmt.Println("SendStmt")
	case *ast.IncDecStmt:
		fmt.Println("IncDecStmt")
	case *ast.AssignStmt:
		fmt.Println("AssignStmt")
	case *ast.GoStmt:
		fmt.Println("GoStmt")
	case *ast.DeferStmt:
		fmt.Println("DeferStmt")
	case *ast.ReturnStmt:
		fmt.Println("ReturnStmt")
	case *ast.BranchStmt:
		fmt.Println("BranchStmt")
	case *ast.BlockStmt:
		fmt.Println("BlockStmt")
	case *ast.IfStmt:
		fmt.Println("IfStmt")
	case *ast.CaseClause:
		fmt.Println("CaseClause")
	case *ast.SwitchStmt:
		fmt.Println("SwitchStmt")
	case *ast.TypeSwitchStmt:
		fmt.Println("TypeSwitchStmt")
	case *ast.CommClause:
		fmt.Println("CommClause")
	case *ast.SelectStmt:
		fmt.Println("SelectStmt")
	case *ast.ForStmt:
		fmt.Println("ForStmt")
	case *ast.RangeStmt:
		fmt.Println("RangeStmt")
		// Declarations
	case *ast.ImportSpec:
		fmt.Println("ImportSpec")
	case *ast.ValueSpec:
		fmt.Println("ValueSpec")
	case *ast.TypeSpec:
		fmt.Println("TypeSpec")
	case *ast.BadDecl:
		fmt.Println("BadDecl")
	case *ast.GenDecl:
		fmt.Println("GenDecl")
	case *ast.FuncDecl:
		fmt.Println("FuncDecl")
		// Files and packages
	case *ast.File:
		fmt.Println("File")
	case *ast.Package:
		fmt.Println("Package")
	default:
		fmt.Printf("ast.Walk: unexpected node type %T", n)
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
