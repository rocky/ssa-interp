// Copyright 2013 Rocky Bernstein.
// evaluation support
package gub

import (
	"fmt"
	"strings"
	"go/ast"
	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"github.com/rocky/ssa-interp/interp"
	"code.google.com/p/go.tools/go/types"
)

func Val(lit string) exact.Value {
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

func IndexExpr(e *ast.IndexExpr) exact.Value {
	// FIXME: clean up this mess
	val := EvalExpr(e.Index)
	if val == nil { return nil }
	if val.Kind() != exact.Int {
		Errmsg("Index at pos %d must be an unsigned integer",
			e.Index.Pos())
		return nil
	}
	var index uint64
	var ok bool
	if index, ok = exact.Uint64Val(val); !ok {
		Errmsg("Index at pos %d must be an unsigned integer",
			e.Index.Pos())
		return nil
	}
	switch id := e.X.(type) {
	case *ast.Ident:
		if k, _, _ := EnvLookup(curFrame, id.Name, curScope); k != nil {
			val := DerefValue(curFrame.Get(k))
			ary := val.([]interp.Value)
			if index < 0 || index >= uint64(len(ary)) {
				Errmsg("index %d out of bounds (0..%d)",
					index, len(ary))
				return nil
			}
			return Val(interp.ToInspect(ary[index]))
		}
	default:
		Errmsg("Can't handle index without a simple id before [] at pos %d", id.Pos())
	}
	return nil
}

// FIXME: returning exact.Value down the line is probably not going to
// cut it. We want an ssa2.Value
func EvalExprStart(n ast.Node, typ types.Type) exact.Value {
	return EvalExpr(n)
}

// FIXME: returning exact.Value down the line is probably not going to
// cut it. We want an ssa2.Value.
func EvalExpr(n ast.Node) exact.Value {
	switch e := n.(type) {
	case *ast.BasicLit:
		return Val(e.Value)
	case *ast.BinaryExpr:
		x := EvalExpr(e.X)
		if x == nil { return nil }
		y := EvalExpr(e.Y)
		if y == nil { return nil }
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
		return exact.UnaryOp(e.Op, EvalExpr(e.X), -1)
	case *ast.CallExpr:
		Errmsg("Can't handle call (%s) yet at pos %d", e.Fun, e.Pos())
		return nil
	case *ast.Ident:
		if k, val, _ := EnvLookup(curFrame, e.Name, curScope); k != nil {
			return Val(val)
		}
		Errmsg("Can't find value for id '%s' here at pos %d", e.Name, e.Pos())
		return nil
	case *ast.ParenExpr:
		return EvalExpr(e.X)
	case *ast.IndexExpr:
		return IndexExpr(e)
	case *ast.SelectorExpr:
		fn := curFrame.Fn()
		info := fn.Pkg.Info()
		if info == nil {
			Errmsg("Package info is nil. Was this compiled with debug?")
			return nil
		}
		sel := info.Selections[e]
		if sel == nil {
			Errmsg("Can't handle selection yet.")
			// Errmsg("Can't handle selection yet unless it is in the environment.")
			// ast.Print(nil, n)
			// for k, v := range info.Selections {
			// 	fmt.Printf("XXXX key type %T, key value %v\n", k, k)
			// 	fmt.Printf("XXXX value type %T, key value %v\n", v, v)
			// 	fmt.Println("-----------------------------")
			// }
			return nil
		}
		switch sel.Kind() {
		case types.PackageObj:
			obj := sel.Obj()
			Msg("todo: pick up from %s", obj)
			// Errmsg("undefined package-qualified name: " + obj.Name())
		case types.FieldVal:
			Msg("todo: pick up %d, from %s", sel.Index(), sel.Obj)
		}
		fmt.Println("Can't handle selector")
		fmt.Printf("n: %s, e: %s\n", n, e)
		return nil
	default:
		fmt.Println("Can't handle")
		fmt.Printf("n: %s, e: %s\n", n, e)
		return nil
	}
}
