package main

import (
	"fmt"
	"strings"
	"go/ast"
	"go/parser"
	"go/token"
	"gitub.com/rocky/go-exact"
)

// ----------------------------------------------------------------------------
// Support functions

func val(lit string) exact.Value {
	if len(lit) == 0 {
		return exact.MakeUnknown()
	}

	switch lit {
	case "?":
		return exact.MakeUnknown()
	case "nil":
		return nil
	case "true":
		return exact.MakeBool(true)
	case "false":
		return exact.MakeBool(false)
	}

	tok := token.FLOAT
	switch first, last := lit[0], lit[len(lit)-1]; {
	case first == '"' || first == '`':
		tok = token.STRING
		lit = strings.Replace(lit, "_", " ", -1)
	case first == '\'':
		tok = token.CHAR
	case last == 'i':
		tok = token.IMAG
	}

	return exact.MakeFromLiteral(lit, tok)
}

func evalAction(n ast.Node) exact.Value {
		switch e := n.(type) {
		case *ast.BasicLit:
			return val(e.Value)
		case *ast.BinaryExpr:
			x := evalAction(e.X)
			if x == nil {
				return nil
			}
			y := evalAction(e.Y)
			if y == nil {
				return nil
			}
			switch e.Op {
			case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
				return exact.MakeBool(exact.Compare(x, e.Op, y))
			case token.SHL, token.SHR:
				s, _ := exact.Int64Val(y)
				return exact.Shift(x, e.Op, uint(s))
			default:
				return exact.BinaryOp(x, e.Op, y)
			}
		case *ast.UnaryExpr:
			return exact.UnaryOp(e.Op, evalAction(e.X), -1)
		case *ast.CallExpr:
			fmt.Printf("Can't handle call (%s) yet at pos %d\n", e.Fun, e.Pos())
			return nil
		case *ast.Ident:
			fmt.Printf("Can't handle Ident %s here at pos %d\n", e.Name, e.Pos())
			return nil
		case *ast.ParenExpr:
			return evalAction(e.X)
		default:
			fmt.Println("Can't handle")
			fmt.Printf("n: %s, e: %s\n", n, e)
			return nil
		}
	}

func main() {
	// src is the input for which we want to inspect the AST.
	exprs := []string {
		"\"quoted\" string with backslash \\",
		"f(3.14)*2 + c",
		"-2  ",  // trailing spaces to be devius
		" 5 == 6",
		"5\t< 6", // that's a tab in there
		"1+2",
		"(1+2)*3",
		"1 << n",
		"1 << 8",
		"y(",
	}

	for _, expr := range exprs {
		// Create the AST by parsing expr.
		f, err := parser.ParseExpr(expr)
		if err != nil {
			fmt.Printf("Error parsing %s: %s", expr, err.Error())
			continue
		}

		// Inspect the AST and print all identifiers and literals.
		if v := evalAction(f); v != nil {
			fmt.Printf("Eval: '%s' ok; value: %s\n", expr, v)
		} else {
			fmt.Printf("Eval '%s' no good\n", expr)
		}
		fmt.Println("--------------")
	}
}
