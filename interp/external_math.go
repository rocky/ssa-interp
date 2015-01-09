// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

// Emulated functions that we cannot interpret because they are
// external or because they use "unsafe" or "reflect" operations.

import (
	"math"
	"github.com/rocky/ssa-interp/interp/big"
)

func ext۰math۰Float64frombits(fr *Frame, args []Value) Value {
	return math.Float64frombits(args[0].(uint64))
}

func ext۰math۰Float64bits(fr *Frame, args []Value) Value {
	return math.Float64bits(args[0].(float64))
}

func ext۰math۰Float32frombits(fr *Frame, args []Value) Value {
	return math.Float32frombits(args[0].(uint32))
}

func ext۰math۰Abs(fr *Frame, args []Value) Value {
	return math.Abs(args[0].(float64))
}

func ext۰math۰Acos(fr *Frame, args []Value) Value {
	return math.Acos(args[0].(float64))
}

func ext۰math۰Asin(fr *Frame, args []Value) Value {
	return math.Asin(args[0].(float64))
}

func ext۰math۰Atan(fr *Frame, args []Value) Value {
	return math.Atan(args[0].(float64))
}

func ext۰math۰Atan2(fr *Frame, args []Value) Value {
	return math.Atan2(args[0].(float64), args[1].(float64))
}

func ext۰math۰Ceil(fr *Frame, args []Value) Value {
	return math.Ceil(args[0].(float64))
}

func ext۰math۰Cos(fr *Frame, args []Value) Value {
	return math.Cos(args[0].(float64))
}

func ext۰math۰Dim(fr *Frame, args []Value) Value {
	return math.Dim(args[0].(float64), args[1].(float64))
}

func ext۰math۰Exp(fr *Frame, args []Value) Value {
	return math.Exp(args[0].(float64))
}

func ext۰math۰Expm1(fr *Frame, args []Value) Value {
	return math.Expm1(args[0].(float64))
}

func ext۰math۰Float32bits(fr *Frame, args []Value) Value {
	return math.Float32bits(args[0].(float32))
}

func ext۰math۰Floor(fr *Frame, args []Value) Value {
	return math.Floor(args[0].(float64))
}

func ext۰math۰Frexp(fr *Frame, args []Value) Value {
	frac, int := math.Frexp(args[0].(float64))
	return tuple{frac, int}
}

func ext۰math۰Hypot(fr *Frame, args []Value) Value {
	return math.Hypot(args[0].(float64), args[1].(float64))
}

func ext۰math۰Ldexp(fr *Frame, args []Value) Value {
	return math.Ldexp(args[0].(float64), args[1].(int))
}

func ext۰math۰Log(fr *Frame, args []Value) Value {
	return math.Log(args[0].(float64))
}

func ext۰math۰Log10(fr *Frame, args []Value) Value {
	return math.Log10(args[0].(float64))
}

func ext۰math۰Log1p(fr *Frame, args []Value) Value {
	return math.Log1p(args[0].(float64))
}

func ext۰math۰Log2(fr *Frame, args []Value) Value {
	return math.Log2(args[0].(float64))
}

func ext۰math۰Max(fn *Frame, args []Value) Value {
	return math.Max(args[0].(float64), args[1].(float64))
}

func ext۰math۰Min(fn *Frame, args []Value) Value {
	return math.Min(args[0].(float64), args[1].(float64))
}

func ext۰math۰Mod(fn *Frame, args []Value) Value {
	return math.Mod(args[0].(float64), args[1].(float64))
}

func ext۰math۰Modf(fn *Frame, args []Value) Value {
	int, frac := math.Modf(args[0].(float64))
	return tuple{int, frac}
}

func ext۰math۰Remainder(fr *Frame, args []Value) Value {
	return math.Remainder(args[0].(float64), args[1].(float64))
}

func ext۰math۰Sin(fr *Frame, args []Value) Value {
	return math.Sin(args[0].(float64))
}

func ext۰math۰Sincos(fr *Frame, args []Value) Value {
	sin, cos := math.Sincos(args[0].(float64))
	return tuple{sin, cos}
}

func ext۰math۰Sqrt(fr *Frame, args []Value) Value {
	return math.Sqrt(args[0].(float64))
}

func ext۰math۰Tan(fr *Frame, args []Value) Value {
	return math.Tan(args[0].(float64))
}

func ext۰math۰Trunc(fr *Frame, args []Value) Value {
	return math.Trunc(args[0].(float64))
}

func ext۰math۰big۰bitLen(fr *Frame, args []Value) Value {
	return big.BitLen(args[0].(big.Word))
}

// func ext۰math۰big۰divWVW(fr *Frame, args []Value) Value {
// 	return big.DivWVW(args[0].([]big.Word), args[1].(big.Word),
// 		args[2].([]big.Word), args[3].(big.Word))
// }

func ext۰math۰big۰mulAddVWW(fr *Frame, args []Value) Value {
	return big.MulAddVWW(args[0].([]big.Word), args[1].([]big.Word),
		args[2].(big.Word), args[3].(big.Word))
}

func ext۰math۰big۰shlVU(fr *Frame, args []Value) Value {

	z := []big.Word {}
	za := args[0].([]Value)
	for _, v := range za {
		w := big.Word(v.(uintptr))
		z = append(z, w)
	}
	x := []big.Word {}
	xa := args[1].([]Value)
	for _, v := range xa {
		w := big.Word(v.(uintptr))
		x = append(x, w)
	}
	return big.ShlVU(z, x, args[2].(uint))
}

// func ext۰math۰big۰shrVU(fr *Frame, args []Value) Value {

// 	z := []big.Word {}
// 	za := args[0].([]Value)
// 	for _, v := range za {
// 		w := big.Word(v.(uintptr))
// 		z = append(z, w)
// 	}
// 	x := []big.Word {}
// 	xa := args[1].([]Value)
// 	for _, v := range xa {
// 		w := big.Word(v.(uintptr))
// 		x = append(x, w)
// 	}
// 	return big.ShrVU(z, x, args[2].(uint))
// }

// func ext۰math۰big۰subVV(fr *Frame, args []Value) Value {

// 	return big.SubVV(args[0].([]big.Word), args[1].([]big.Word),
// 		args[2].([]big.Word))
// }

// The set of remaining native functions we need to implement (as needed):

// math/big/arith_decl.go:8:func mulWW(x, y Word) (z1, z0 Word)
// math/big/arith_decl.go:9:func divWW(x1, x0, y Word) (q, r Word)
// math/big/arith_decl.go:10:func addVV(z, x, y []Word) (c Word)
// math/big/arith_decl.go:11:func subVV(z, x, y []Word) (c Word)
// math/big/arith_decl.go:12:func addVW(z, x []Word, y Word) (c Word)
// math/big/arith_decl.go:13:func subVW(z, x []Word, y Word) (c Word)
// math/big/arith_decl.go:14:func shlVU(z, x []Word, s uint) (c Word)
// math/big/arith_decl.go:15:func shrVU(z, x []Word, s uint) (c Word)
// math/big/arith_decl.go:16:func mulAddVWW(z, x []Word, y, r Word) (c Word)
// math/big/arith_decl.go:17:func addMulVVW(z, x []Word, y Word) (c Word)
