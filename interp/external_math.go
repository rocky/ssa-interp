// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interp

// Emulated functions that we cannot interpret because they are
// external or because they use "unsafe" or "reflect" operations.

import (
	"math"
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
