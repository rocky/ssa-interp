// Copyright 2013 Rocky Bernstein.
// evaluation support
package gub

import (
	"fmt"
	"strings"
	"go/ast"
	"go/token"
	"code.google.com/p/go.tools/go/exact"
	"github.com/rocky/ssa-interp"
	"github.com/rocky/ssa-interp/interp"
)

func EnvLookup(fr *interp.Frame, name string) (ssa2.Value, string) {
	for k, v := range fr.Env() {
		if name == k.Name() {
			v := deref2Str(v)
			return k, v
		}
	}
	return nil, ""
}

func Val(lit string) exact.Value {
	if len(lit) == 0 {
		return exact.MakeUnknown()
	}

	switch lit {
	case "?":
		return exact.MakeUnknown()
	case "nil":
		return exact.MakeNil()
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
	val := evalExpr(e.Index)
	if val == nil { return nil }
	if val.Kind() != exact.Int {
		errmsg("Index at pos %d must be an unsigned integer",
			e.Index.Pos())
		return nil
	}
	var index uint64
	var ok bool
	if index, ok = exact.Uint64Val(val); !ok {
		errmsg("Index at pos %d must be an unsigned integer",
			e.Index.Pos())
		return nil
	}
	switch id := e.X.(type) {
	case *ast.Ident:
		if k, _ := EnvLookup(curFrame, id.Name); k != nil {
			val := derefValue(curFrame.Get(k))
			ary := val.([]interp.Value)
			if index < 0 || index >= uint64(len(ary)) {
				errmsg("index %d out of bounds (0..%d)",
					index, len(ary))
				return nil
			}
			return Val(interp.ToInspect(ary[index]))
		}
	default:
		errmsg("Can't handle index without a simple id before [] at pos %d", id.Pos())
	}
	return nil
}

// FIXME: returning exact.Value down the line is probably not going to
// cut it
func evalExpr(n ast.Node) exact.Value {
		switch e := n.(type) {
		case *ast.BasicLit:
			return Val(e.Value)
		case *ast.BinaryExpr:
			x := evalExpr(e.X)
			if x == nil { return nil }
			y := evalExpr(e.Y)
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
			return exact.UnaryOp(e.Op, evalExpr(e.X), -1)
		case *ast.CallExpr:
			errmsg("Can't handle call (%s) yet at pos %d", e.Fun, e.Pos())
			return nil
		case *ast.Ident:
			if k, val := EnvLookup(curFrame, e.Name); k != nil {
				return Val(val)
			}
			errmsg("Can't find value for id '%s' here at pos %d", e.Name, e.Pos())
			return nil
		case *ast.ParenExpr:
			return evalExpr(e.X)
		case *ast.IndexExpr:
			return IndexExpr(e)
		default:
			fmt.Println("Can't handle")
			fmt.Printf("n: %s, e: %s\n", n, e)
			return nil
		}
	}

// Could something like this go into interp-ssa?
func GetFunction(name string) *ssa2.Function {
	pkg := curFrame.Fn().Pkg
	ids := strings.Split(name, ".")
	if len(ids) > 1 {
		try_pkg := curFrame.I().Program().PackageByName(ids[0])
		if try_pkg != nil { pkg = try_pkg }
		m := pkg.Members[ids[1]]
		if m == nil { return nil }
		name = ids[1]
	}
	if fn := pkg.Func(name); fn != nil {
		return fn
	}
	return nil
}
