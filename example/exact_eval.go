// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"go/token"
	"github.com/rocky/go-exact"
	"strings"
	"fmt"
)

// TODO(gri) expand this test framework

var tests = []string{
	// unary operations
	`+ 0 = 0`,
	`- 1 = -1`,
	`^ 0 = -1`,

	`! true = false`,
	`! false = true`,

	// etc.

	// binary operations
	`"" + "" = ""`,
	`"foo" + "" = "foo"`,
	`"" + "bar" = "bar"`,
	`"foo" + "bar" = "foobar"`,

	`0 + 0 = 0`,
	`0 + 0.1 = 0.1`,
	`0 + 0.1i = 0.1i`,
	`0.1 + 0.9 = 1`,
	`1e100 + 1e100 = 2e100`,
	`? + 0 = ?`,
	`0 + ? = ?`,

	`0 - 0 = 0`,
	`0 - 0.1 = -0.1`,
	`0 - 0.1i = -0.1i`,
	`1e100 - 1e100 = 0`,
	`? - 0 = ?`,
	`0 - ? = ?`,

	`0 * 0 = 0`,
	`1 * 0.1 = 0.1`,
	`1 * 0.1i = 0.1i`,
	`1i * 1i = -1`,
	`? * 0 = ?`,
	`0 * ? = ?`,

	`0 / 0 = "division_by_zero"`,
	`10 / 2 = 5`,
	`5 / 3 = 5/3`,
	`5i / 3i = 5/3`,
	`? / 0 = ?`,
	`0 / ? = ?`,

	`0 % 0 = "runtime_error:_integer_divide_by_zero"`, // TODO(gri) should be the same as for /
	`10 % 3 = 1`,
	`? % 0 = ?`,
	`0 % ? = ?`,

	`0 & 0 = 0`,
	`12345 & 0 = 0`,
	`0xff & 0xf = 0xf`,
	`? & 0 = ?`,
	`0 & ? = ?`,

	`0 | 0 = 0`,
	`12345 | 0 = 12345`,
	`0xb | 0xa0 = 0xab`,
	`? | 0 = ?`,
	`0 | ? = ?`,

	`0 ^ 0 = 0`,
	`1 ^ -1 = -2`,
	`? ^ 0 = ?`,
	`0 ^ ? = ?`,

	`0 &^ 0 = 0`,
	`0xf &^ 1 = 0xe`,
	`1 &^ 0xf = 0`,
	// etc.

	// shifts
	`0 << 0 = 0`,
	`1 << 10 = 1024`,
	`0 >> 0 = 0`,
	`1024 >> 10 == 1`,
	`? << 0 == ?`,
	`? >> 10 == ?`,
	// etc.

	// comparisons
	`false == false = true`,
	`false == true = false`,
	`true == false = false`,
	`true == true = true`,

	`false != false = false`,
	`false != true = true`,
	`true != false = true`,
	`true != true = false`,

	`"foo" == "bar" = false`,
	`"foo" != "bar" = true`,
	`"foo" < "bar" = false`,
	`"foo" <= "bar" = false`,
	`"foo" > "bar" = true`,
	`"foo" >= "bar" = true`,

	`0 == 0 = true`,
	`0 != 0 = false`,
	`0 < 10 = true`,
	`10 <= 10 = true`,
	`0 > 10 = false`,
	`10 >= 10 = true`,

	`1/123456789 == 1/123456789 == true`,
	`1/123456789 != 1/123456789 == false`,
	`1/123456789 < 1/123456788 == true`,
	`1/123456788 <= 1/123456789 == false`,
	`0.11 > 0.11 = false`,
	`0.11 >= 0.11 = true`,

	`? == 0 = false`,
	`? != 0 = false`,
	`? < 10 = false`,
	`? <= 10 = false`,
	`? > 10 = false`,
	`? >= 10 = false`,

	`0 == ? = false`,
	`0 != ? = false`,
	`0 < ? = false`,
	`10 <= ? = false`,
	`0 > ? = false`,
	`10 >= ? = false`,

	// etc.
}

func main() {
	for _, test := range tests {
		var got, want exact.Value
		var a []string

		switch a = strings.Split(test, " "); len(a) {
		case 4:
			got = doOp(nil, optab[a[0]], val(a[1]))
			want = val(a[3])
		case 5:
			got = doOp(val(a[0]), optab[a[1]], val(a[2]))
			want = val(a[4])
		default:
			fmt.Printf("invalid test case: %s\n", test)
			continue
		}

		if !exact.Compare(got, token.EQL, want) {
			fmt.Printf("%s failed: got %s; want %s\n", test, got, want)
		} else {
			expr := strings.Join(a[:len(a)-2], " ")
			fmt.Printf("%s gave: %s\n", expr, got)
		}
	}
}

// ----------------------------------------------------------------------------
// Support functions

func val(lit string) exact.Value {
	if len(lit) == 0 {
		return exact.MakeUnknown()
	}

	switch lit {
	case "?":
		return exact.MakeUnknown()
	case "true":
		return exact.MakeBool(true)
	case "false":
		return exact.MakeBool(false)
	}

	tok := token.INT
	switch first, last := lit[0], lit[len(lit)-1]; {
	case first == '"' || first == '`':
		tok = token.STRING
		lit = strings.Replace(lit, "_", " ", -1)
	case first == '\'':
		tok = token.CHAR
	case last == 'i':
		tok = token.IMAG
	default:
		if !strings.HasPrefix(lit, "0x") && strings.ContainsAny(lit, "./Ee") {
			tok = token.FLOAT
		}
	}

	return exact.MakeFromLiteral(lit, tok)
}

var optab = map[string]token.Token{
	"!": token.NOT,

	"+": token.ADD,
	"-": token.SUB,
	"*": token.MUL,
	"/": token.QUO,
	"%": token.REM,

	"<<": token.SHL,
	">>": token.SHR,

	"&":  token.AND,
	"|":  token.OR,
	"^":  token.XOR,
	"&^": token.AND_NOT,

	"==": token.EQL,
	"!=": token.NEQ,
	"<":  token.LSS,
	"<=": token.LEQ,
	">":  token.GTR,
	">=": token.GEQ,
}

func panicHandler(v *exact.Value) {
	switch p := recover().(type) {
	case nil:
		// nothing to do
	case string:
		*v = exact.MakeString(p)
	case error:
		*v = exact.MakeString(p.Error())
	default:
		panic(p)
	}
}

func doOp(x exact.Value, op token.Token, y exact.Value) (z exact.Value) {
	defer panicHandler(&z)

	if x == nil {
		return exact.UnaryOp(op, y, -1)
	}

	switch op {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		return exact.MakeBool(exact.Compare(x, op, y))
	case token.SHL, token.SHR:
		s, _ := exact.Int64Val(y)
		return exact.Shift(x, op, uint(s))
	default:
		return exact.BinaryOp(x, op, y)
	}
}
