// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

var addVV     func(z, x, y []Word) (c Word)         = addVV_g
var addVW     func(z, x []Word, y Word) (c Word)    = addVW_g
var addMulVVW func(z, x []Word, y Word) (c Word)    = addMulVVW_g
var BitLen    func(x Word) (n int)                  = bitLen_g
var bitLen    func(x Word) (n int)                  = bitLen_g
var divWW     func(x1, x0, y Word)(Word, Word)      = divWW_g
var divWVW    func(z []Word, xn Word, x []Word, y Word) (r Word) = divWVW_g
var mulAddVWW func(z, x []Word, y, r Word) (c Word) = mulAddVWW_g
var MulAddVWW func(z, x []Word, y, r Word) (c Word) = mulAddVWW_g
var mulWW     func(x, y Word) (z1, z0 Word)         = mulWW_g
var shlVU     func(z, x []Word, s uint) (c Word)    = shlVU_g
var ShlVU     func(z, x []Word, s uint) (c Word)    = shlVU_g
var shrVU     func(z, x []Word, s uint) (c Word)    = shrVU_g
var subVV     func(z, x, y []Word) (c Word)         = subVV_g
var subVW     func(z, x []Word, y Word) (c Word) 	= subVW_g
