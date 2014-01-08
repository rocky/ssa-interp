// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

// implemented in arith_$GOARCH.s

func AddVV(z, x, y []Word) (c Word) {
	return addVV_g(z, x, y)
}

func AddVW(z, x []Word, y Word) (c Word) {
	return addVW_g(z, x, y)
}

func AddMulVVW(z, x []Word, y Word) (c Word) {
	return addMulVVW_g(z, x, y)
}

func BitLen(x Word) (n int) {
	return bitLen_g(x)
}

func DivWW(x1, x0, y Word) (q, r Word) {
	return divWW_g(x1, x0, y)
}

func DivWVW(z []Word, xn Word, x []Word, y Word) (r Word) {
	return divWVW_g(z, xn, x, y)
}

func MulAddVWW(z, x []Word, y, r Word) (c Word) {
	return mulAddVWW_g(z, x, y, r)
}

func MulWW(x, y Word) (z1, z0 Word) {
	return mulWW_g(x, y)
}

func ShlVU(z, x []Word, s uint) (c Word) {
	return shlVU_g(z, x, s)
}

func ShrVU(z, x []Word, s uint) (c Word) {
	return shrVU_g(z, x, s)
}

func SubVV(z, x, y []Word) (c Word) {
	return subVV_g(z, x, y)
}

func SubVW(z, x []Word, y Word) (c Word) {
	return subVW_g(z, x, y)
}
