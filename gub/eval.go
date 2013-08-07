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

func DerefValue(v interp.Value) interp.Value {
	switch v := v.(type) {
	case *interp.Value:
		if v == nil { return nil }
		return *v
	default:
		return v
	}
}

func Deref2Str(v interp.Value) string {
	return interp.ToInspect(DerefValue(v))
}


func PrintInEnvironment(fr *interp.Frame, name string) bool {
	if k, v, scope := EnvLookup(fr, name); k != nil {
		envStr := ""
		if scope != nil {
			envStr = fmt.Sprintf(" at scope %d", scope.ScopeNum())
		}
		Msg("%s is in the environment%s", name, envStr)
		Msg("\t%s = %s", k, DerefValue(v))
		return true
	} else {
		Errmsg("Name %s not found in environment", name)
		return false
	}
}

func EnvLookup(fr *interp.Frame, name string) (ssa2.Value, string, *ssa2.Scope) {
	fn := fr.Fn()
	if i := fn.LocalsByName[name]; i > 0 {
		k := fn.Locals[i-1]
		v := Deref2Str(fr.Env()[k])
		return k, v, k.Scope
	}
	for k, v := range fr.Env() {
		if name == k.Name() {
			v := Deref2Str(v)
			switch k := k.(type) {
			case *ssa2.Alloc:
				return k, v, k.Scope
			default:
				return k, v, nil
			}
		}
	}
	return nil, "", nil
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
		if k, _, _ := EnvLookup(curFrame, id.Name); k != nil {
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
			Errmsg("Can't handle call (%s) yet at pos %d", e.Fun, e.Pos())
			return nil
		case *ast.Ident:
			if k, val, _ := EnvLookup(curFrame, e.Name); k != nil {
				return Val(val)
			}
			Errmsg("Can't find value for id '%s' here at pos %d", e.Name, e.Pos())
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
