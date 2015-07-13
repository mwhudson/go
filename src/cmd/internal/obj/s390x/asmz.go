// Based on cmd/internal/obj/ppc64/asm9.go.
//
//    Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//    Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//    Portions Copyright © 1997-1999 Vita Nuova Limited
//    Portions Copyright © 2000-2008 Vita Nuova Holdings Limited (www.vitanuova.com)
//    Portions Copyright © 2004,2006 Bruce Ellis
//    Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//    Revisions Copyright © 2000-2008 Lucent Technologies Inc. and others
//    Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package s390x

import (
	"cmd/internal/obj"
	"log"
	"sort"
)

// Instruction layout.
const (
	FuncAlign = 16
)

const (
	r0iszero = 1
)

type Optab struct {
	as    int16 // opcode
	a1    uint8 // source
	a2    uint8 // register
	a3    uint8 // destination
	a4    uint8
	type_ int8
	param int16
}

var optab = []Optab{
	// p.Optab, p.From.Class, p.Reg, p.From3.Class, p.To.Class, type, param
	Optab{obj.ATEXT, C_ADDR, C_NONE, C_NONE, C_TEXTSIZE, 0, 0},
	Optab{obj.ATEXT, C_ADDR, C_NONE, C_LCON, C_TEXTSIZE, 0, 0},
	/* move register */
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_REG, 1, 0},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_REG, 12, 0},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_REG, 13, 0},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_REG, 12, 0},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_REG, 13, 0},
	Optab{AADD, C_REG, C_REG, C_NONE, C_REG, 2, 0},
	Optab{AADD, C_REG, C_NONE, C_NONE, C_REG, 2, 0},
	Optab{AADD, C_ADDCON, C_REG, C_NONE, C_REG, 4, 0},
	Optab{AADD, C_ADDCON, C_NONE, C_NONE, C_REG, 4, 0},
	Optab{AADD, C_UCON, C_NONE, C_NONE, C_REG, 20, 0},
	Optab{AADD, C_UCON, C_REG, C_NONE, C_REG, 20, 0},
	Optab{AADD, C_LCON, C_REG, C_NONE, C_REG, 22, 0},
	Optab{AADD, C_LCON, C_NONE, C_NONE, C_REG, 22, 0},
	Optab{AADDC, C_REG, C_REG, C_NONE, C_REG, 2, 0},
	Optab{AADDC, C_REG, C_NONE, C_NONE, C_REG, 2, 0},
	Optab{AADDC, C_ADDCON, C_REG, C_NONE, C_REG, 4, 0},
	Optab{AADDC, C_ADDCON, C_NONE, C_NONE, C_REG, 4, 0},
	Optab{AADDC, C_LCON, C_REG, C_NONE, C_REG, 22, 0},
	Optab{AADDC, C_LCON, C_NONE, C_NONE, C_REG, 22, 0},
	Optab{AAND, C_REG, C_REG, C_NONE, C_REG, 6, 0}, /* logical, no literal */
	Optab{AAND, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{AAND, C_ANDCON, C_NONE, C_NONE, C_REG, 58, 0},
	Optab{AAND, C_ANDCON, C_REG, C_NONE, C_REG, 58, 0},
	Optab{AAND, C_UCON, C_NONE, C_NONE, C_REG, 59, 0},
	Optab{AAND, C_UCON, C_REG, C_NONE, C_REG, 59, 0},
	Optab{AAND, C_LCON, C_NONE, C_NONE, C_REG, 23, 0},
	Optab{AAND, C_LCON, C_REG, C_NONE, C_REG, 23, 0},
	Optab{AMULLW, C_REG, C_REG, C_NONE, C_REG, 2, 0},
	Optab{AMULLW, C_REG, C_NONE, C_NONE, C_REG, 2, 0},
	Optab{AMULLW, C_ADDCON, C_REG, C_NONE, C_REG, 4, 0},
	Optab{AMULLW, C_ADDCON, C_NONE, C_NONE, C_REG, 4, 0},
	Optab{AMULLW, C_ANDCON, C_REG, C_NONE, C_REG, 4, 0},
	Optab{AMULLW, C_ANDCON, C_NONE, C_NONE, C_REG, 4, 0},
	Optab{AMULLW, C_LCON, C_REG, C_NONE, C_REG, 22, 0},
	Optab{AMULLW, C_LCON, C_NONE, C_NONE, C_REG, 22, 0},
	Optab{ASUBC, C_REG, C_REG, C_NONE, C_REG, 10, 0},
	Optab{ASUBC, C_REG, C_NONE, C_NONE, C_REG, 10, 0},
	Optab{ASUBC, C_REG, C_NONE, C_ADDCON, C_REG, 27, 0},
	Optab{ASUBC, C_REG, C_NONE, C_LCON, C_REG, 28, 0},
	Optab{AOR, C_REG, C_REG, C_NONE, C_REG, 6, 0}, /* logical, literal not cc (or/xor) */
	Optab{AOR, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{AOR, C_ANDCON, C_NONE, C_NONE, C_REG, 58, 0},
	Optab{AOR, C_ANDCON, C_REG, C_NONE, C_REG, 58, 0},
	Optab{AOR, C_UCON, C_NONE, C_NONE, C_REG, 59, 0},
	Optab{AOR, C_UCON, C_REG, C_NONE, C_REG, 59, 0},
	Optab{AOR, C_LCON, C_NONE, C_NONE, C_REG, 23, 0},
	Optab{AOR, C_LCON, C_REG, C_NONE, C_REG, 23, 0},
	Optab{ADIVW, C_REG, C_REG, C_NONE, C_REG, 2, 0}, /* op r1[,r2],r3 */
	Optab{ADIVW, C_REG, C_NONE, C_NONE, C_REG, 2, 0},
	Optab{ASUB, C_REG, C_REG, C_NONE, C_REG, 10, 0}, /* op r2[,r1],r3 */
	Optab{ASUB, C_REG, C_NONE, C_NONE, C_REG, 10, 0},
	Optab{ASLW, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{ASLW, C_REG, C_REG, C_NONE, C_REG, 6, 0},
	Optab{ASLD, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{ASLD, C_REG, C_REG, C_NONE, C_REG, 6, 0},
	Optab{ASLD, C_SCON, C_REG, C_NONE, C_REG, 25, 0},
	Optab{ASLD, C_SCON, C_NONE, C_NONE, C_REG, 25, 0},
	Optab{ASLW, C_SCON, C_REG, C_NONE, C_REG, 57, 0},
	Optab{ASLW, C_SCON, C_NONE, C_NONE, C_REG, 57, 0},
	Optab{ASRAW, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{ASRAW, C_REG, C_REG, C_NONE, C_REG, 6, 0},
	Optab{ASRAW, C_SCON, C_REG, C_NONE, C_REG, 56, 0},
	Optab{ASRAW, C_SCON, C_NONE, C_NONE, C_REG, 56, 0},
	Optab{ASRAD, C_REG, C_NONE, C_NONE, C_REG, 6, 0},
	Optab{ASRAD, C_REG, C_REG, C_NONE, C_REG, 6, 0},
	Optab{ASRAD, C_SCON, C_REG, C_NONE, C_REG, 56, 0},
	Optab{ASRAD, C_SCON, C_NONE, C_NONE, C_REG, 56, 0},
	Optab{ARLWMI, C_SCON, C_REG, C_LCON, C_REG, 62, 0},
	Optab{ARLWMI, C_REG, C_REG, C_LCON, C_REG, 63, 0},
	Optab{ARLDMI, C_SCON, C_REG, C_LCON, C_REG, 30, 0},
	Optab{ARLDC, C_SCON, C_REG, C_LCON, C_REG, 29, 0},
	Optab{ARLDCL, C_SCON, C_REG, C_LCON, C_REG, 29, 0},
	Optab{ARLDCL, C_REG, C_REG, C_LCON, C_REG, 14, 0},
	Optab{ARLDCL, C_REG, C_NONE, C_LCON, C_REG, 14, 0},
	Optab{AFADD, C_FREG, C_NONE, C_NONE, C_FREG, 2, 0},
	Optab{AFADD, C_FREG, C_REG, C_NONE, C_FREG, 2, 0},
	Optab{AFABS, C_FREG, C_NONE, C_NONE, C_FREG, 33, 0},
	Optab{AFABS, C_NONE, C_NONE, C_NONE, C_FREG, 33, 0},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_FREG, 33, 0},
	Optab{AFMADD, C_FREG, C_REG, C_FREG, C_FREG, 34, 0},
	Optab{AFMUL, C_FREG, C_NONE, C_NONE, C_FREG, 32, 0},
	Optab{AFMUL, C_FREG, C_REG, C_NONE, C_FREG, 32, 0},
	Optab{ACS, C_REG, C_REG, C_NONE, C_SOREG, 79, 0},
	Optab{ACSG, C_REG, C_REG, C_NONE, C_SOREG, 79, 0},
	Optab{ACEFBRA, C_REG, C_NONE, C_NONE, C_FREG, 82, 0},
	Optab{ACFEBRA, C_FREG, C_NONE, C_NONE, C_REG, 83, 0},
	Optab{AMVC, C_SOREG, C_NONE, C_SCON, C_SOREG, 84, 0},
	Optab{ALARL, C_LCON, C_NONE, C_NONE, C_REG, 85, 0},
	Optab{ALA, C_SOREG, C_NONE, C_NONE, C_REG, 86, 0},
	Optab{AEXRL, C_LCON, C_NONE, C_NONE, C_REG, 87, 0},
	Optab{ASTCK, C_NONE, C_NONE, C_NONE, C_SAUTO, 88, REGSP},
	Optab{ASTCK, C_NONE, C_NONE, C_NONE, C_SOREG, 88, 0},

	/* store, short offset */
	Optab{AMOVD, C_REG, C_REG, C_NONE, C_ZOREG, 7, REGZERO},
	Optab{AMOVW, C_REG, C_REG, C_NONE, C_ZOREG, 7, REGZERO},
	Optab{AMOVWZ, C_REG, C_REG, C_NONE, C_ZOREG, 7, REGZERO},
	Optab{AMOVBZ, C_REG, C_REG, C_NONE, C_ZOREG, 7, REGZERO},
	Optab{AMOVB, C_REG, C_REG, C_NONE, C_ZOREG, 7, REGZERO},
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},

	/* load, short offset */
	Optab{AMOVD, C_ZOREG, C_REG, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVW, C_ZOREG, C_REG, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVWZ, C_ZOREG, C_REG, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVBZ, C_ZOREG, C_REG, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVB, C_ZOREG, C_REG, C_NONE, C_REG, 9, REGZERO},
	Optab{AMOVD, C_SAUTO, C_NONE, C_NONE, C_REG, 8, REGSP},
	Optab{AMOVW, C_SAUTO, C_NONE, C_NONE, C_REG, 8, REGSP},
	Optab{AMOVWZ, C_SAUTO, C_NONE, C_NONE, C_REG, 8, REGSP},
	Optab{AMOVBZ, C_SAUTO, C_NONE, C_NONE, C_REG, 8, REGSP},
	Optab{AMOVB, C_SAUTO, C_NONE, C_NONE, C_REG, 9, REGSP},
	Optab{AMOVD, C_SOREG, C_NONE, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVW, C_SOREG, C_NONE, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVWZ, C_SOREG, C_NONE, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVBZ, C_SOREG, C_NONE, C_NONE, C_REG, 8, REGZERO},
	Optab{AMOVB, C_SOREG, C_NONE, C_NONE, C_REG, 9, REGZERO},

	/* store, long offset */
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AMOVD, C_REG, C_NONE, C_NONE, C_ADDR, 74, 0},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_ADDR, 74, 0},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_ADDR, 74, 0},
	Optab{AMOVBZ, C_REG, C_NONE, C_NONE, C_ADDR, 74, 0},
	Optab{AMOVB, C_REG, C_NONE, C_NONE, C_ADDR, 74, 0},

	/* load, long offset */
	Optab{AMOVD, C_LAUTO, C_NONE, C_NONE, C_REG, 36, REGSP},
	Optab{AMOVW, C_LAUTO, C_NONE, C_NONE, C_REG, 36, REGSP},
	Optab{AMOVWZ, C_LAUTO, C_NONE, C_NONE, C_REG, 36, REGSP},
	Optab{AMOVBZ, C_LAUTO, C_NONE, C_NONE, C_REG, 36, REGSP},
	Optab{AMOVB, C_LAUTO, C_NONE, C_NONE, C_REG, 37, REGSP},
	Optab{AMOVD, C_LOREG, C_NONE, C_NONE, C_REG, 36, REGZERO},
	Optab{AMOVW, C_LOREG, C_NONE, C_NONE, C_REG, 36, REGZERO},
	Optab{AMOVWZ, C_LOREG, C_NONE, C_NONE, C_REG, 36, REGZERO},
	Optab{AMOVBZ, C_LOREG, C_NONE, C_NONE, C_REG, 36, REGZERO},
	Optab{AMOVB, C_LOREG, C_NONE, C_NONE, C_REG, 37, REGZERO},
	Optab{AMOVD, C_ADDR, C_NONE, C_NONE, C_REG, 75, 0},
	Optab{AMOVW, C_ADDR, C_NONE, C_NONE, C_REG, 75, 0},
	Optab{AMOVWZ, C_ADDR, C_NONE, C_NONE, C_REG, 75, 0},
	Optab{AMOVBZ, C_ADDR, C_NONE, C_NONE, C_REG, 75, 0},
	Optab{AMOVB, C_ADDR, C_NONE, C_NONE, C_REG, 76, 0},

	/* store constant */
	Optab{AMOVD, C_SCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVW, C_SCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVWZ, C_SCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVB, C_SCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVBZ, C_SCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVD, C_ADDCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVW, C_ADDCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVWZ, C_ADDCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVB, C_ADDCON, C_NONE, C_NONE, C_SOREG, 92, 0},
	Optab{AMOVBZ, C_ADDCON, C_NONE, C_NONE, C_SOREG, 92, 0},

	/* load constant */
	Optab{AMOVD, C_SACON, C_NONE, C_NONE, C_REG, 3, REGSP},
	Optab{AMOVD, C_LACON, C_NONE, C_NONE, C_REG, 26, REGSP},
	Optab{AMOVD, C_ADDCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVW, C_SACON, C_NONE, C_NONE, C_REG, 3, REGSP},
	Optab{AMOVW, C_LACON, C_NONE, C_NONE, C_REG, 26, REGSP},
	Optab{AMOVW, C_ADDCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVWZ, C_SACON, C_NONE, C_NONE, C_REG, 3, REGSP},
	Optab{AMOVWZ, C_LACON, C_NONE, C_NONE, C_REG, 26, REGSP},
	Optab{AMOVWZ, C_ADDCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVB, C_ADDCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVBZ, C_ADDCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVB, C_SCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVBZ, C_SCON, C_NONE, C_NONE, C_REG, 3, 0},

	/* load unsigned/long constants (TODO: check) */
	Optab{AMOVD, C_UCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVD, C_LCON, C_NONE, C_NONE, C_REG, 19, 0},
	Optab{AMOVW, C_UCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVW, C_LCON, C_NONE, C_NONE, C_REG, 19, 0},
	Optab{AMOVWZ, C_UCON, C_NONE, C_NONE, C_REG, 3, 0},
	Optab{AMOVWZ, C_LCON, C_NONE, C_NONE, C_REG, 19, 0},
	Optab{AMOVHBR, C_ZOREG, C_REG, C_NONE, C_REG, 45, 0},
	Optab{AMOVHBR, C_ZOREG, C_NONE, C_NONE, C_REG, 45, 0},
	Optab{AMOVHBR, C_REG, C_REG, C_NONE, C_ZOREG, 44, 0},
	Optab{AMOVHBR, C_REG, C_NONE, C_NONE, C_ZOREG, 44, 0},

	Optab{ASYSCALL, C_NONE, C_NONE, C_NONE, C_NONE, 5, 0},
	Optab{ASYSCALL, C_SCON, C_NONE, C_NONE, C_NONE, 77, 0},
	Optab{ABEQ, C_NONE, C_NONE, C_NONE, C_SBRA, 16, 0},
	Optab{ABR, C_NONE, C_NONE, C_NONE, C_LBRA, 11, 0},
	Optab{ABC, C_SCON, C_REG, C_NONE, C_SBRA, 16, 0},
	Optab{ABC, C_SCON, C_REG, C_NONE, C_LBRA, 17, 0},
	Optab{ABR, C_NONE, C_NONE, C_NONE, C_REG, 18, 0},
	Optab{ABR, C_REG, C_NONE, C_NONE, C_REG, 18, 0},
	Optab{ABR, C_NONE, C_NONE, C_NONE, C_ZOREG, 15, 0},
	Optab{ABC, C_NONE, C_NONE, C_NONE, C_ZOREG, 15, 0},
	Optab{ACMPBEQ, C_REG, C_REG, C_NONE, C_SBRA, 89, 0},
	Optab{ACMPBEQ, C_REG, C_NONE, C_ADDCON, C_SBRA, 90, 0},
	Optab{ACMPBEQ, C_REG, C_NONE, C_SCON, C_SBRA, 90, 0},
	Optab{ACMPUBEQ, C_REG, C_REG, C_NONE, C_SBRA, 89, 0},
	Optab{ACMPUBEQ, C_REG, C_NONE, C_ANDCON, C_SBRA, 90, 0},

	Optab{AFMOVD, C_SAUTO, C_NONE, C_NONE, C_FREG, 8, REGSP},
	Optab{AFMOVD, C_SOREG, C_NONE, C_NONE, C_FREG, 8, REGZERO},
	Optab{AFMOVD, C_LAUTO, C_NONE, C_NONE, C_FREG, 36, REGSP},
	Optab{AFMOVD, C_LOREG, C_NONE, C_NONE, C_FREG, 36, REGZERO},
	Optab{AFMOVD, C_ADDR, C_NONE, C_NONE, C_FREG, 75, 0},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_SAUTO, 7, REGSP},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_SOREG, 7, REGZERO},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_LAUTO, 35, REGSP},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_LOREG, 35, REGZERO},
	Optab{AFMOVD, C_FREG, C_NONE, C_NONE, C_ADDR, 74, 0},
	Optab{ASYNC, C_NONE, C_NONE, C_NONE, C_NONE, 81, 0},
	Optab{ABYTE, C_SCON, C_NONE, C_NONE, C_NONE, 40, 0},
	Optab{AWORD, C_LCON, C_NONE, C_NONE, C_NONE, 40, 0},
	Optab{ADWORD, C_LCON, C_NONE, C_NONE, C_NONE, 31, 0},
	Optab{ADWORD, C_DCON, C_NONE, C_NONE, C_NONE, 31, 0},
	Optab{AADDME, C_REG, C_NONE, C_NONE, C_REG, 47, 0},
	Optab{ANEG, C_REG, C_NONE, C_NONE, C_REG, 47, 0},
	Optab{ANEG, C_NONE, C_NONE, C_NONE, C_REG, 47, 0},
	Optab{AREM, C_REG, C_NONE, C_NONE, C_REG, 50, 0},
	Optab{AREM, C_REG, C_REG, C_NONE, C_REG, 50, 0},
	Optab{AREMU, C_REG, C_NONE, C_NONE, C_REG, 50, 0},
	Optab{AREMU, C_REG, C_REG, C_NONE, C_REG, 50, 0},
	Optab{AREMD, C_REG, C_NONE, C_NONE, C_REG, 51, 0},
	Optab{AREMD, C_REG, C_REG, C_NONE, C_REG, 51, 0},
	Optab{AREMDU, C_REG, C_NONE, C_NONE, C_REG, 51, 0},
	Optab{AREMDU, C_REG, C_REG, C_NONE, C_REG, 51, 0},

	/* 32-bit access registers */
	Optab{AMOVW, C_AREG, C_NONE, C_NONE, C_REG, 68, 0},
	Optab{AMOVWZ, C_AREG, C_NONE, C_NONE, C_REG, 68, 0},
	Optab{AMOVW, C_REG, C_NONE, C_NONE, C_AREG, 69, 0},
	Optab{AMOVWZ, C_REG, C_NONE, C_NONE, C_AREG, 69, 0},

	Optab{ACMP, C_REG, C_NONE, C_NONE, C_REG, 70, 0},
	Optab{ACMP, C_REG, C_REG, C_NONE, C_REG, 70, 0},
	Optab{ACMP, C_REG, C_NONE, C_NONE, C_ADDCON, 71, 0},
	Optab{ACMP, C_REG, C_REG, C_NONE, C_ADDCON, 71, 0},
	Optab{ACMPU, C_REG, C_NONE, C_NONE, C_REG, 70, 0},
	Optab{ACMPU, C_REG, C_REG, C_NONE, C_REG, 70, 0},
	Optab{ACMPU, C_REG, C_NONE, C_NONE, C_ANDCON, 71, 0},
	Optab{ACMPU, C_REG, C_REG, C_NONE, C_ANDCON, 71, 0},
	Optab{AFCMPO, C_FREG, C_NONE, C_NONE, C_FREG, 70, 0},
	Optab{AFCMPO, C_FREG, C_REG, C_NONE, C_FREG, 70, 0},

	Optab{obj.AUNDEF, C_NONE, C_NONE, C_NONE, C_NONE, 78, 0},
	Optab{obj.APCDATA, C_LCON, C_NONE, C_NONE, C_LCON, 0, 0},
	Optab{obj.AFUNCDATA, C_SCON, C_NONE, C_NONE, C_ADDR, 0, 0},
	Optab{obj.ANOP, C_NONE, C_NONE, C_NONE, C_NONE, 0, 0},
	Optab{obj.ANOP, C_SAUTO, C_NONE, C_NONE, C_NONE, 0, 0},

	Optab{obj.AXXX, C_NONE, C_NONE, C_NONE, C_NONE, 0, 0},
}

type Oprang struct {
	start []Optab
	stop  []Optab
}

var oprange [ALAST & obj.AMask]Oprang

var xcmp [C_NCLASS][C_NCLASS]uint8

func spanz(ctxt *obj.Link, cursym *obj.LSym) {
	p := cursym.Text
	if p == nil || p.Link == nil { // handle external functions and ELF section symbols
		return
	}
	ctxt.Cursym = cursym
	ctxt.Autosize = int32(p.To.Offset + 8)

	if oprange[AANDN&obj.AMask].start == nil {
		buildop(ctxt)
	}

	buffer := make([]byte, 0)
	changed := true
	for changed {
		changed = false
		buffer = buffer[:0]
		ctxt.Cursym.R = make([]obj.Reloc, 0)
		for p := cursym.Text; p != nil; p = p.Link {
			pc := int64(len(buffer))
			if pc != p.Pc {
				changed = true
			}
			p.Pc = pc
			ctxt.Pc = p.Pc
			ctxt.Curp = p
			asmout(ctxt, &buffer)
			if pc == int64(len(buffer)) {
				switch p.As {
				case obj.ANOP, obj.AFUNCDATA, obj.APCDATA, obj.ATEXT:
					// ok
				default:
					ctxt.Diag("zero-width instruction\n%v", p)
				}
			}
		}
	}

	cursym.Size = int64(len(buffer))
	if cursym.Size%FuncAlign != 0 {
		cursym.Size += FuncAlign - (cursym.Size % FuncAlign)
	}
	obj.Symgrow(ctxt, cursym, cursym.Size)
	copy(cursym.P, buffer)
}

func isint32(v int64) bool {
	return int64(int32(v)) == v
}

func isuint32(v uint64) bool {
	return uint64(uint32(v)) == v
}

func aclass(ctxt *obj.Link, a *obj.Addr) int {
	switch a.Type {
	case obj.TYPE_NONE:
		return C_NONE

	case obj.TYPE_REG:
		if REG_R0 <= a.Reg && a.Reg <= REG_R15 {
			return C_REG
		}
		if REG_F0 <= a.Reg && a.Reg <= REG_F15 {
			return C_FREG
		}
		if REG_AR0 <= a.Reg && a.Reg <= REG_AR15 {
			return C_AREG
		}
		return C_GOK

	case obj.TYPE_MEM:
		switch a.Name {
		case obj.NAME_EXTERN,
			obj.NAME_STATIC:
			if a.Sym != nil { // use relocation
				ctxt.Instoffset = a.Offset
				return C_ADDR
			}

		case obj.NAME_AUTO:
			ctxt.Instoffset = int64(ctxt.Autosize) + a.Offset
			if ctxt.Instoffset >= -BIG && ctxt.Instoffset < BIG {
				return C_SAUTO
			}
			return C_LAUTO

		case obj.NAME_PARAM:
			ctxt.Instoffset = int64(ctxt.Autosize) + a.Offset + 8
			if ctxt.Instoffset >= -BIG && ctxt.Instoffset < BIG {
				return C_SAUTO
			}
			return C_LAUTO

		case obj.NAME_NONE:
			ctxt.Instoffset = a.Offset
			if ctxt.Instoffset == 0 {
				return C_ZOREG
			}
			if ctxt.Instoffset >= -BIG && ctxt.Instoffset < BIG {
				return C_SOREG
			}
			return C_LOREG
		}

		return C_GOK

	case obj.TYPE_TEXTSIZE:
		return C_TEXTSIZE

	case obj.TYPE_CONST,
		obj.TYPE_ADDR:
		switch a.Name {
		case obj.TYPE_NONE:
			ctxt.Instoffset = a.Offset
			if a.Reg != 0 {
				if -BIG <= ctxt.Instoffset && ctxt.Instoffset <= BIG {
					return C_SACON
				}
				if isint32(ctxt.Instoffset) {
					return C_LACON
				}
				return C_DACON
			}
			goto consize

		case obj.NAME_EXTERN,
			obj.NAME_STATIC:
			s := a.Sym
			if s == nil {
				break
			}
			if s.Type == obj.SCONST {
				ctxt.Instoffset = s.Value + a.Offset
				goto consize
			}

			ctxt.Instoffset = s.Value + a.Offset

			/* not sure why this barfs */
			return C_LCON

		case obj.NAME_AUTO:
			ctxt.Instoffset = int64(ctxt.Autosize) + a.Offset
			if ctxt.Instoffset >= -BIG && ctxt.Instoffset < BIG {
				return C_SACON
			}
			return C_LACON

		case obj.NAME_PARAM:
			ctxt.Instoffset = int64(ctxt.Autosize) + a.Offset + 8
			if ctxt.Instoffset >= -BIG && ctxt.Instoffset < BIG {
				return C_SACON
			}
			return C_LACON
		}

		return C_GOK

	consize:
		if ctxt.Instoffset >= 0 {
			if ctxt.Instoffset <= 0x7fff {
				return C_SCON
			}
			if ctxt.Instoffset <= 0xffff {
				return C_ANDCON
			}
			if ctxt.Instoffset&0xffff == 0 && isuint32(uint64(ctxt.Instoffset)) { /* && (instoffset & (1<<31)) == 0) */
				return C_UCON
			}
			if isint32(ctxt.Instoffset) || isuint32(uint64(ctxt.Instoffset)) {
				return C_LCON
			}
			return C_DCON
		}

		if ctxt.Instoffset >= -0x8000 {
			return C_ADDCON
		}
		if ctxt.Instoffset&0xffff == 0 && isint32(ctxt.Instoffset) {
			return C_UCON
		}
		if isint32(ctxt.Instoffset) {
			return C_LCON
		}
		return C_DCON

	case obj.TYPE_BRANCH:
		return C_SBRA
	}

	return C_GOK
}

func oplook(ctxt *obj.Link, p *obj.Prog) *Optab {
	a1 := int(p.Optab)
	if a1 != 0 {
		return &optab[a1-1:][0]
	}
	a1 = int(p.From.Class)
	if a1 == 0 {
		a1 = aclass(ctxt, &p.From) + 1
		p.From.Class = int8(a1)
	}

	a1--
	a3 := C_NONE + 1
	if p.From3 != nil {
		a3 = int(p.From3.Class)
		if a3 == 0 {
			a3 = aclass(ctxt, p.From3) + 1
			p.From3.Class = int8(a3)
		}
	}

	a3--
	a4 := int(p.To.Class)
	if a4 == 0 {
		a4 = aclass(ctxt, &p.To) + 1
		p.To.Class = int8(a4)
	}

	a4--
	a2 := C_NONE
	if p.Reg != 0 {
		a2 = C_REG
	}

	r0 := p.As & obj.AMask

	o := oprange[r0].start
	if o == nil {
		o = oprange[r0].stop /* just generate an error */
	}

	e := oprange[r0].stop
	c1 := xcmp[a1][:]
	c3 := xcmp[a3][:]
	c4 := xcmp[a4][:]
	for ; -cap(o) < -cap(e); o = o[1:] {
		if int(o[0].a2) == a2 {
			if c1[o[0].a1] != 0 {
				if c3[o[0].a3] != 0 {
					if c4[o[0].a4] != 0 {
						p.Optab = uint16((-cap(o) + cap(optab)) + 1)
						return &o[0]
					}
				}
			}
		}
	}

	// cannot find a case; abort
	ctxt.Diag("illegal combination %v %v %v %v %v\n", obj.Aconv(int(p.As)), DRconv(a1), DRconv(a2), DRconv(a3), DRconv(a4))
	ctxt.Diag("prog: %v\n", p)
	return nil
}

func cmp(a int, b int) bool {
	if a == b {
		return true
	}
	switch a {
	case C_LCON:
		if b == C_ZCON || b == C_SCON || b == C_UCON || b == C_ADDCON || b == C_ANDCON {
			return true
		}

	case C_ADDCON:
		if b == C_ZCON || b == C_SCON {
			return true
		}

	case C_ANDCON:
		if b == C_ZCON || b == C_SCON {
			return true
		}

	case C_UCON:
		if b == C_ZCON {
			return true
		}

	case C_SCON:
		if b == C_ZCON {
			return true
		}

	case C_LACON:
		if b == C_SACON {
			return true
		}

	case C_LBRA:
		if b == C_SBRA {
			return true
		}

	case C_LAUTO:
		if b == C_SAUTO {
			return true
		}

	case C_REG:
		if b == C_ZCON {
			return r0iszero != 0 /*TypeKind(100016)*/
		}

	case C_LOREG:
		if b == C_ZOREG || b == C_SOREG {
			return true
		}

	case C_SOREG:
		if b == C_ZOREG {
			return true
		}

	case C_ANY:
		return true
	}

	return false
}

type ocmp []Optab

func (x ocmp) Len() int {
	return len(x)
}

func (x ocmp) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x ocmp) Less(i, j int) bool {
	p1 := &x[i]
	p2 := &x[j]
	n := int(p1.as) - int(p2.as)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a1) - int(p2.a1)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a2) - int(p2.a2)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a3) - int(p2.a3)
	if n != 0 {
		return n < 0
	}
	n = int(p1.a4) - int(p2.a4)
	if n != 0 {
		return n < 0
	}
	return false
}
func opset(a, b0 int16) {
	oprange[a&obj.AMask] = oprange[b0]
}

func buildop(ctxt *obj.Link) {
	var n int

	for i := 0; i < C_NCLASS; i++ {
		for n = 0; n < C_NCLASS; n++ {
			if cmp(n, i) {
				xcmp[i][n] = 1
			}
		}
	}
	for n = 0; optab[n].as != obj.AXXX; n++ {
	}
	sort.Sort(ocmp(optab[:n]))
	for i := 0; i < n; i++ {
		r := optab[i].as
		r0 := r & obj.AMask
		oprange[r0].start = optab[i:]
		for optab[i].as == r {
			i++
		}
		oprange[r0].stop = optab[i:]
		i--

		// opset() aliases optab ranges for similar instructions, to reduce the number of optabs in the array.
		// oprange[] is used by oplook() to find the Optab entry that applies to a given Prog.
		switch r {
		default:
			ctxt.Diag("unknown op in build: %v", obj.Aconv(int(r)))
			log.Fatalf("bad code")

		case AREM:

		case AREMU:

		case AREMD:

		case AREMDU:

		case ADIVW: /* op Rb[,Ra],Rd */
			opset(AADDE, r0)
			opset(AMULLD, r0)
			opset(AMULHD, r0)
			opset(AMULHDU, r0)
			opset(ADIVD, r0)
			opset(ADIVDU, r0)

		case AMOVBZ: /* lbz, stz, rlwm(r/r), lhz, lha, stz, and x variants */
			opset(AMOVH, r0)
			opset(AMOVHZ, r0)

		case ALA:
			opset(ALAY, r0)

		case ALARL:

		case AMVC:
			opset(ACLC, r0)
			opset(AXC, r0)
			opset(AOC, r0)
			opset(ANC, r0)

		case AEXRL:

		case ASTCK:
			opset(ASTCKC, r0)
			opset(ASTCKE, r0)
			opset(ASTCKF, r0)

		case AAND: /* logical op Rb,Rs,Ra; no literal */
			opset(AANDN, r0)
			opset(ANAND, r0)
			opset(ANOR, r0)
			opset(AORN, r0)

		case AADDME: /* op Ra, Rd */
			opset(AADDZE, r0)
			opset(ASUBME, r0)
			opset(ASUBZE, r0)

		case AADDC:

		case ABEQ:
			opset(ABGE, r0)
			opset(ABGT, r0)
			opset(ABLE, r0)
			opset(ABLT, r0)
			opset(ABNE, r0)
			opset(ABVC, r0)
			opset(ABVS, r0)

		case ABR:
			opset(ABL, r0)

		case ABC:
			opset(ABCL, r0)

		case AFABS: /* fop [s,]d */
			opset(AFNABS, r0)
			opset(AFNEG, r0)
			opset(AFRSP, r0)
			opset(ALDEBR, r0)
			opset(AFSQRT, r0)
			opset(AFSQRTS, r0)

		case AFADD:
			opset(AFADDS, r0)
			opset(AFDIV, r0)
			opset(AFDIVS, r0)
			opset(AFSUB, r0)
			opset(AFSUBS, r0)

		case AFMADD:
			opset(AFMADDS, r0)
			opset(AFMSUB, r0)
			opset(AFMSUBS, r0)
			opset(AFNMADD, r0)
			opset(AFNMADDS, r0)
			opset(AFNMSUB, r0)
			opset(AFNMSUBS, r0)

		case AFMUL:
			opset(AFMULS, r0)

		case AFCMPO:
			opset(AFCMPU, r0)
			opset(ACEBR, r0)

		case ANEG: /* op [Ra,] Rd */

		case AOR: /* or/xor Rb,Rs,Ra; ori/xori $uimm,Rs,Ra; oris/xoris $uimm,Rs,Ra */
			opset(AXOR, r0)

		case ASLW:
			opset(ASRW, r0)

		case ASLD:
			opset(ASRD, r0)

		case ACS, ACSG:

		case ASRAW: /* sraw Rb,Rs,Ra; srawi sh,Rs,Ra */

		case ASRAD: /* sraw Rb,Rs,Ra; srawi sh,Rs,Ra */

		case ASUB: /* SUB Ra,Rb,Rd => subf Rd,ra,rb */
			opset(ASUBC, r0)
			opset(ASUBE, r0)

		case ASYNC:

		case ARLWMI:
			opset(ARLWNM, r0)

		case ARLDMI:

		case ARLDC:

		case ARLDCL:
			opset(ARLDCR, r0)

		case AFMOVD:
			opset(AFMOVS, r0)

		case ASYSCALL: /* just the op; flow of control */

		case AMOVHBR:
			opset(AMOVWBR, r0)

		case ACMP:
			opset(ACMPW, r0)

		case ACMPU:
			opset(ACMPWU, r0)

		case ACEFBRA:
			opset(ACDFBRA, r0)
			opset(ACEGBRA, r0)
			opset(ACDGBRA, r0)
			opset(ACELFBR, r0)
			opset(ACDLFBR, r0)
			opset(ACELGBR, r0)
			opset(ACDLGBR, r0)

		case ACFEBRA:
			opset(ACFDBRA, r0)
			opset(ACGEBRA, r0)
			opset(ACGDBRA, r0)
			opset(ACLFEBR, r0)
			opset(ACLFDBR, r0)
			opset(ACLGEBR, r0)
			opset(ACLGDBR, r0)

		case ACMPBEQ:
			opset(ACMPBGE, r0)
			opset(ACMPBGT, r0)
			opset(ACMPBLE, r0)
			opset(ACMPBLT, r0)
			opset(ACMPBNE, r0)

		case ACMPUBEQ:
			opset(ACMPUBGE, r0)
			opset(ACMPUBGT, r0)
			opset(ACMPUBLE, r0)
			opset(ACMPUBLT, r0)
			opset(ACMPUBNE, r0)

		case AADD,
			AMOVW,
			/* load/store/move word with sign extension; special 32-bit move; move 32-bit literals */
			AMOVWZ, /* load/store/move word with zero extension; move 32-bit literals  */
			AMOVD,  /* load/store/move 64-bit values, including 32-bit literals with/without sign-extension */
			AMOVB,  /* macro: move byte with sign extension */
			AMULLW,
			/* op $s[,r2],r3; op r1[,r2],r3; no cc/v */
			ASUBC, /* op r1,$s,r3; op r1[,r2],r3 */
			ABYTE,
			AWORD,
			ADWORD,
			obj.ANOP,
			obj.ATEXT,
			obj.AUNDEF,
			obj.AFUNCDATA,
			obj.APCDATA:
			break
		}
	}
}

const (
	OP_A       uint32 = 0x5A00 // FORMAT_RX1        ADD (32)
	OP_AD      uint32 = 0x6A00 // FORMAT_RX1        ADD NORMALIZED (long HFP)
	OP_ADB     uint32 = 0xED1A // FORMAT_RXE        ADD (long BFP)
	OP_ADBR    uint32 = 0xB31A // FORMAT_RRE        ADD (long BFP)
	OP_ADR     uint32 = 0x2A00 // FORMAT_RR         ADD NORMALIZED (long HFP)
	OP_ADTR    uint32 = 0xB3D2 // FORMAT_RRF1       ADD (long DFP)
	OP_ADTRA   uint32 = 0xB3D2 // FORMAT_RRF1       ADD (long DFP)
	OP_AE      uint32 = 0x7A00 // FORMAT_RX1        ADD NORMALIZED (short HFP)
	OP_AEB     uint32 = 0xED0A // FORMAT_RXE        ADD (short BFP)
	OP_AEBR    uint32 = 0xB30A // FORMAT_RRE        ADD (short BFP)
	OP_AER     uint32 = 0x3A00 // FORMAT_RR         ADD NORMALIZED (short HFP)
	OP_AFI     uint32 = 0xC209 // FORMAT_RIL1       ADD IMMEDIATE (32)
	OP_AG      uint32 = 0xE308 // FORMAT_RXY1       ADD (64)
	OP_AGF     uint32 = 0xE318 // FORMAT_RXY1       ADD (64<-32)
	OP_AGFI    uint32 = 0xC208 // FORMAT_RIL1       ADD IMMEDIATE (64<-32)
	OP_AGFR    uint32 = 0xB918 // FORMAT_RRE        ADD (64<-32)
	OP_AGHI    uint32 = 0xA70B // FORMAT_RI1        ADD HALFWORD IMMEDIATE (64)
	OP_AGHIK   uint32 = 0xECD9 // FORMAT_RIE4       ADD IMMEDIATE (64<-16)
	OP_AGR     uint32 = 0xB908 // FORMAT_RRE        ADD (64)
	OP_AGRK    uint32 = 0xB9E8 // FORMAT_RRF1       ADD (64)
	OP_AGSI    uint32 = 0xEB7A // FORMAT_SIY        ADD IMMEDIATE (64<-8)
	OP_AH      uint32 = 0x4A00 // FORMAT_RX1        ADD HALFWORD
	OP_AHHHR   uint32 = 0xB9C8 // FORMAT_RRF1       ADD HIGH (32)
	OP_AHHLR   uint32 = 0xB9D8 // FORMAT_RRF1       ADD HIGH (32)
	OP_AHI     uint32 = 0xA70A // FORMAT_RI1        ADD HALFWORD IMMEDIATE (32)
	OP_AHIK    uint32 = 0xECD8 // FORMAT_RIE4       ADD IMMEDIATE (32<-16)
	OP_AHY     uint32 = 0xE37A // FORMAT_RXY1       ADD HALFWORD
	OP_AIH     uint32 = 0xCC08 // FORMAT_RIL1       ADD IMMEDIATE HIGH (32)
	OP_AL      uint32 = 0x5E00 // FORMAT_RX1        ADD LOGICAL (32)
	OP_ALC     uint32 = 0xE398 // FORMAT_RXY1       ADD LOGICAL WITH CARRY (32)
	OP_ALCG    uint32 = 0xE388 // FORMAT_RXY1       ADD LOGICAL WITH CARRY (64)
	OP_ALCGR   uint32 = 0xB988 // FORMAT_RRE        ADD LOGICAL WITH CARRY (64)
	OP_ALCR    uint32 = 0xB998 // FORMAT_RRE        ADD LOGICAL WITH CARRY (32)
	OP_ALFI    uint32 = 0xC20B // FORMAT_RIL1       ADD LOGICAL IMMEDIATE (32)
	OP_ALG     uint32 = 0xE30A // FORMAT_RXY1       ADD LOGICAL (64)
	OP_ALGF    uint32 = 0xE31A // FORMAT_RXY1       ADD LOGICAL (64<-32)
	OP_ALGFI   uint32 = 0xC20A // FORMAT_RIL1       ADD LOGICAL IMMEDIATE (64<-32)
	OP_ALGFR   uint32 = 0xB91A // FORMAT_RRE        ADD LOGICAL (64<-32)
	OP_ALGHSIK uint32 = 0xECDB // FORMAT_RIE4       ADD LOGICAL WITH SIGNED IMMEDIATE (64<-16)
	OP_ALGR    uint32 = 0xB90A // FORMAT_RRE        ADD LOGICAL (64)
	OP_ALGRK   uint32 = 0xB9EA // FORMAT_RRF1       ADD LOGICAL (64)
	OP_ALGSI   uint32 = 0xEB7E // FORMAT_SIY        ADD LOGICAL WITH SIGNED IMMEDIATE (64<-8)
	OP_ALHHHR  uint32 = 0xB9CA // FORMAT_RRF1       ADD LOGICAL HIGH (32)
	OP_ALHHLR  uint32 = 0xB9DA // FORMAT_RRF1       ADD LOGICAL HIGH (32)
	OP_ALHSIK  uint32 = 0xECDA // FORMAT_RIE4       ADD LOGICAL WITH SIGNED IMMEDIATE (32<-16)
	OP_ALR     uint32 = 0x1E00 // FORMAT_RR         ADD LOGICAL (32)
	OP_ALRK    uint32 = 0xB9FA // FORMAT_RRF1       ADD LOGICAL (32)
	OP_ALSI    uint32 = 0xEB6E // FORMAT_SIY        ADD LOGICAL WITH SIGNED IMMEDIATE (32<-8)
	OP_ALSIH   uint32 = 0xCC0A // FORMAT_RIL1       ADD LOGICAL WITH SIGNED IMMEDIATE HIGH (32)
	OP_ALSIHN  uint32 = 0xCC0B // FORMAT_RIL1       ADD LOGICAL WITH SIGNED IMMEDIATE HIGH (32)
	OP_ALY     uint32 = 0xE35E // FORMAT_RXY1       ADD LOGICAL (32)
	OP_AP      uint32 = 0xFA00 // FORMAT_SS2        ADD DECIMAL
	OP_AR      uint32 = 0x1A00 // FORMAT_RR         ADD (32)
	OP_ARK     uint32 = 0xB9F8 // FORMAT_RRF1       ADD (32)
	OP_ASI     uint32 = 0xEB6A // FORMAT_SIY        ADD IMMEDIATE (32<-8)
	OP_AU      uint32 = 0x7E00 // FORMAT_RX1        ADD UNNORMALIZED (short HFP)
	OP_AUR     uint32 = 0x3E00 // FORMAT_RR         ADD UNNORMALIZED (short HFP)
	OP_AW      uint32 = 0x6E00 // FORMAT_RX1        ADD UNNORMALIZED (long HFP)
	OP_AWR     uint32 = 0x2E00 // FORMAT_RR         ADD UNNORMALIZED (long HFP)
	OP_AXBR    uint32 = 0xB34A // FORMAT_RRE        ADD (extended BFP)
	OP_AXR     uint32 = 0x3600 // FORMAT_RR         ADD NORMALIZED (extended HFP)
	OP_AXTR    uint32 = 0xB3DA // FORMAT_RRF1       ADD (extended DFP)
	OP_AXTRA   uint32 = 0xB3DA // FORMAT_RRF1       ADD (extended DFP)
	OP_AY      uint32 = 0xE35A // FORMAT_RXY1       ADD (32)
	OP_BAKR    uint32 = 0xB240 // FORMAT_RRE        BRANCH AND STACK
	OP_BAL     uint32 = 0x4500 // FORMAT_RX1        BRANCH AND LINK
	OP_BALR    uint32 = 0x0500 // FORMAT_RR         BRANCH AND LINK
	OP_BAS     uint32 = 0x4D00 // FORMAT_RX1        BRANCH AND SAVE
	OP_BASR    uint32 = 0x0D00 // FORMAT_RR         BRANCH AND SAVE
	OP_BASSM   uint32 = 0x0C00 // FORMAT_RR         BRANCH AND SAVE AND SET MODE
	OP_BC      uint32 = 0x4700 // FORMAT_RX2        BRANCH ON CONDITION
	OP_BCR     uint32 = 0x0700 // FORMAT_RR         BRANCH ON CONDITION
	OP_BCT     uint32 = 0x4600 // FORMAT_RX1        BRANCH ON COUNT (32)
	OP_BCTG    uint32 = 0xE346 // FORMAT_RXY1       BRANCH ON COUNT (64)
	OP_BCTGR   uint32 = 0xB946 // FORMAT_RRE        BRANCH ON COUNT (64)
	OP_BCTR    uint32 = 0x0600 // FORMAT_RR         BRANCH ON COUNT (32)
	OP_BPP     uint32 = 0xC700 // FORMAT_SMI        BRANCH PREDICTION PRELOAD
	OP_BPRP    uint32 = 0xC500 // FORMAT_MII        BRANCH PREDICTION RELATIVE PRELOAD
	OP_BRAS    uint32 = 0xA705 // FORMAT_RI2        BRANCH RELATIVE AND SAVE
	OP_BRASL   uint32 = 0xC005 // FORMAT_RIL2       BRANCH RELATIVE AND SAVE LONG
	OP_BRC     uint32 = 0xA704 // FORMAT_RI3        BRANCH RELATIVE ON CONDITION
	OP_BRCL    uint32 = 0xC004 // FORMAT_RIL3       BRANCH RELATIVE ON CONDITION LONG
	OP_BRCT    uint32 = 0xA706 // FORMAT_RI2        BRANCH RELATIVE ON COUNT (32)
	OP_BRCTG   uint32 = 0xA707 // FORMAT_RI2        BRANCH RELATIVE ON COUNT (64)
	OP_BRCTH   uint32 = 0xCC06 // FORMAT_RIL2       BRANCH RELATIVE ON COUNT HIGH (32)
	OP_BRXH    uint32 = 0x8400 // FORMAT_RSI        BRANCH RELATIVE ON INDEX HIGH (32)
	OP_BRXHG   uint32 = 0xEC44 // FORMAT_RIE5       BRANCH RELATIVE ON INDEX HIGH (64)
	OP_BRXLE   uint32 = 0x8500 // FORMAT_RSI        BRANCH RELATIVE ON INDEX LOW OR EQ. (32)
	OP_BRXLG   uint32 = 0xEC45 // FORMAT_RIE5       BRANCH RELATIVE ON INDEX LOW OR EQ. (64)
	OP_BSA     uint32 = 0xB25A // FORMAT_RRE        BRANCH AND SET AUTHORITY
	OP_BSG     uint32 = 0xB258 // FORMAT_RRE        BRANCH IN SUBSPACE GROUP
	OP_BSM     uint32 = 0x0B00 // FORMAT_RR         BRANCH AND SET MODE
	OP_BXH     uint32 = 0x8600 // FORMAT_RS1        BRANCH ON INDEX HIGH (32)
	OP_BXHG    uint32 = 0xEB44 // FORMAT_RSY1       BRANCH ON INDEX HIGH (64)
	OP_BXLE    uint32 = 0x8700 // FORMAT_RS1        BRANCH ON INDEX LOW OR EQUAL (32)
	OP_BXLEG   uint32 = 0xEB45 // FORMAT_RSY1       BRANCH ON INDEX LOW OR EQUAL (64)
	OP_C       uint32 = 0x5900 // FORMAT_RX1        COMPARE (32)
	OP_CD      uint32 = 0x6900 // FORMAT_RX1        COMPARE (long HFP)
	OP_CDB     uint32 = 0xED19 // FORMAT_RXE        COMPARE (long BFP)
	OP_CDBR    uint32 = 0xB319 // FORMAT_RRE        COMPARE (long BFP)
	OP_CDFBR   uint32 = 0xB395 // FORMAT_RRE        CONVERT FROM FIXED (32 to long BFP)
	OP_CDFBRA  uint32 = 0xB395 // FORMAT_RRF5       CONVERT FROM FIXED (32 to long BFP)
	OP_CDFR    uint32 = 0xB3B5 // FORMAT_RRE        CONVERT FROM FIXED (32 to long HFP)
	OP_CDFTR   uint32 = 0xB951 // FORMAT_RRE        CONVERT FROM FIXED (32 to long DFP)
	OP_CDGBR   uint32 = 0xB3A5 // FORMAT_RRE        CONVERT FROM FIXED (64 to long BFP)
	OP_CDGBRA  uint32 = 0xB3A5 // FORMAT_RRF5       CONVERT FROM FIXED (64 to long BFP)
	OP_CDGR    uint32 = 0xB3C5 // FORMAT_RRE        CONVERT FROM FIXED (64 to long HFP)
	OP_CDGTR   uint32 = 0xB3F1 // FORMAT_RRE        CONVERT FROM FIXED (64 to long DFP)
	OP_CDGTRA  uint32 = 0xB3F1 // FORMAT_RRF5       CONVERT FROM FIXED (64 to long DFP)
	OP_CDLFBR  uint32 = 0xB391 // FORMAT_RRF5       CONVERT FROM LOGICAL (32 to long BFP)
	OP_CDLFTR  uint32 = 0xB953 // FORMAT_RRF5       CONVERT FROM LOGICAL (32 to long DFP)
	OP_CDLGBR  uint32 = 0xB3A1 // FORMAT_RRF5       CONVERT FROM LOGICAL (64 to long BFP)
	OP_CDLGTR  uint32 = 0xB952 // FORMAT_RRF5       CONVERT FROM LOGICAL (64 to long DFP)
	OP_CDR     uint32 = 0x2900 // FORMAT_RR         COMPARE (long HFP)
	OP_CDS     uint32 = 0xBB00 // FORMAT_RS1        COMPARE DOUBLE AND SWAP (32)
	OP_CDSG    uint32 = 0xEB3E // FORMAT_RSY1       COMPARE DOUBLE AND SWAP (64)
	OP_CDSTR   uint32 = 0xB3F3 // FORMAT_RRE        CONVERT FROM SIGNED PACKED (64 to long DFP)
	OP_CDSY    uint32 = 0xEB31 // FORMAT_RSY1       COMPARE DOUBLE AND SWAP (32)
	OP_CDTR    uint32 = 0xB3E4 // FORMAT_RRE        COMPARE (long DFP)
	OP_CDUTR   uint32 = 0xB3F2 // FORMAT_RRE        CONVERT FROM UNSIGNED PACKED (64 to long DFP)
	OP_CDZT    uint32 = 0xEDAA // FORMAT_RSL        CONVERT FROM ZONED (to long DFP)
	OP_CE      uint32 = 0x7900 // FORMAT_RX1        COMPARE (short HFP)
	OP_CEB     uint32 = 0xED09 // FORMAT_RXE        COMPARE (short BFP)
	OP_CEBR    uint32 = 0xB309 // FORMAT_RRE        COMPARE (short BFP)
	OP_CEDTR   uint32 = 0xB3F4 // FORMAT_RRE        COMPARE BIASED EXPONENT (long DFP)
	OP_CEFBR   uint32 = 0xB394 // FORMAT_RRE        CONVERT FROM FIXED (32 to short BFP)
	OP_CEFBRA  uint32 = 0xB394 // FORMAT_RRF5       CONVERT FROM FIXED (32 to short BFP)
	OP_CEFR    uint32 = 0xB3B4 // FORMAT_RRE        CONVERT FROM FIXED (32 to short HFP)
	OP_CEGBR   uint32 = 0xB3A4 // FORMAT_RRE        CONVERT FROM FIXED (64 to short BFP)
	OP_CEGBRA  uint32 = 0xB3A4 // FORMAT_RRF5       CONVERT FROM FIXED (64 to short BFP)
	OP_CEGR    uint32 = 0xB3C4 // FORMAT_RRE        CONVERT FROM FIXED (64 to short HFP)
	OP_CELFBR  uint32 = 0xB390 // FORMAT_RRF5       CONVERT FROM LOGICAL (32 to short BFP)
	OP_CELGBR  uint32 = 0xB3A0 // FORMAT_RRF5       CONVERT FROM LOGICAL (64 to short BFP)
	OP_CER     uint32 = 0x3900 // FORMAT_RR         COMPARE (short HFP)
	OP_CEXTR   uint32 = 0xB3FC // FORMAT_RRE        COMPARE BIASED EXPONENT (extended DFP)
	OP_CFC     uint32 = 0xB21A // FORMAT_S          COMPARE AND FORM CODEWORD
	OP_CFDBR   uint32 = 0xB399 // FORMAT_RRF5       CONVERT TO FIXED (long BFP to 32)
	OP_CFDBRA  uint32 = 0xB399 // FORMAT_RRF5       CONVERT TO FIXED (long BFP to 32)
	OP_CFDR    uint32 = 0xB3B9 // FORMAT_RRF5       CONVERT TO FIXED (long HFP to 32)
	OP_CFDTR   uint32 = 0xB941 // FORMAT_RRF5       CONVERT TO FIXED (long DFP to 32)
	OP_CFEBR   uint32 = 0xB398 // FORMAT_RRF5       CONVERT TO FIXED (short BFP to 32)
	OP_CFEBRA  uint32 = 0xB398 // FORMAT_RRF5       CONVERT TO FIXED (short BFP to 32)
	OP_CFER    uint32 = 0xB3B8 // FORMAT_RRF5       CONVERT TO FIXED (short HFP to 32)
	OP_CFI     uint32 = 0xC20D // FORMAT_RIL1       COMPARE IMMEDIATE (32)
	OP_CFXBR   uint32 = 0xB39A // FORMAT_RRF5       CONVERT TO FIXED (extended BFP to 32)
	OP_CFXBRA  uint32 = 0xB39A // FORMAT_RRF5       CONVERT TO FIXED (extended BFP to 32)
	OP_CFXR    uint32 = 0xB3BA // FORMAT_RRF5       CONVERT TO FIXED (extended HFP to 32)
	OP_CFXTR   uint32 = 0xB949 // FORMAT_RRF5       CONVERT TO FIXED (extended DFP to 32)
	OP_CG      uint32 = 0xE320 // FORMAT_RXY1       COMPARE (64)
	OP_CGDBR   uint32 = 0xB3A9 // FORMAT_RRF5       CONVERT TO FIXED (long BFP to 64)
	OP_CGDBRA  uint32 = 0xB3A9 // FORMAT_RRF5       CONVERT TO FIXED (long BFP to 64)
	OP_CGDR    uint32 = 0xB3C9 // FORMAT_RRF5       CONVERT TO FIXED (long HFP to 64)
	OP_CGDTR   uint32 = 0xB3E1 // FORMAT_RRF5       CONVERT TO FIXED (long DFP to 64)
	OP_CGDTRA  uint32 = 0xB3E1 // FORMAT_RRF5       CONVERT TO FIXED (long DFP to 64)
	OP_CGEBR   uint32 = 0xB3A8 // FORMAT_RRF5       CONVERT TO FIXED (short BFP to 64)
	OP_CGEBRA  uint32 = 0xB3A8 // FORMAT_RRF5       CONVERT TO FIXED (short BFP to 64)
	OP_CGER    uint32 = 0xB3C8 // FORMAT_RRF5       CONVERT TO FIXED (short HFP to 64)
	OP_CGF     uint32 = 0xE330 // FORMAT_RXY1       COMPARE (64<-32)
	OP_CGFI    uint32 = 0xC20C // FORMAT_RIL1       COMPARE IMMEDIATE (64<-32)
	OP_CGFR    uint32 = 0xB930 // FORMAT_RRE        COMPARE (64<-32)
	OP_CGFRL   uint32 = 0xC60C // FORMAT_RIL2       COMPARE RELATIVE LONG (64<-32)
	OP_CGH     uint32 = 0xE334 // FORMAT_RXY1       COMPARE HALFWORD (64<-16)
	OP_CGHI    uint32 = 0xA70F // FORMAT_RI1        COMPARE HALFWORD IMMEDIATE (64<-16)
	OP_CGHRL   uint32 = 0xC604 // FORMAT_RIL2       COMPARE HALFWORD RELATIVE LONG (64<-16)
	OP_CGHSI   uint32 = 0xE558 // FORMAT_SIL        COMPARE HALFWORD IMMEDIATE (64<-16)
	OP_CGIB    uint32 = 0xECFC // FORMAT_RIS        COMPARE IMMEDIATE AND BRANCH (64<-8)
	OP_CGIJ    uint32 = 0xEC7C // FORMAT_RIE3       COMPARE IMMEDIATE AND BRANCH RELATIVE (64<-8)
	OP_CGIT    uint32 = 0xEC70 // FORMAT_RIE1       COMPARE IMMEDIATE AND TRAP (64<-16)
	OP_CGR     uint32 = 0xB920 // FORMAT_RRE        COMPARE (64)
	OP_CGRB    uint32 = 0xECE4 // FORMAT_RRS        COMPARE AND BRANCH (64)
	OP_CGRJ    uint32 = 0xEC64 // FORMAT_RIE2       COMPARE AND BRANCH RELATIVE (64)
	OP_CGRL    uint32 = 0xC608 // FORMAT_RIL2       COMPARE RELATIVE LONG (64)
	OP_CGRT    uint32 = 0xB960 // FORMAT_RRF3       COMPARE AND TRAP (64)
	OP_CGXBR   uint32 = 0xB3AA // FORMAT_RRF5       CONVERT TO FIXED (extended BFP to 64)
	OP_CGXBRA  uint32 = 0xB3AA // FORMAT_RRF5       CONVERT TO FIXED (extended BFP to 64)
	OP_CGXR    uint32 = 0xB3CA // FORMAT_RRF5       CONVERT TO FIXED (extended HFP to 64)
	OP_CGXTR   uint32 = 0xB3E9 // FORMAT_RRF5       CONVERT TO FIXED (extended DFP to 64)
	OP_CGXTRA  uint32 = 0xB3E9 // FORMAT_RRF5       CONVERT TO FIXED (extended DFP to 64)
	OP_CH      uint32 = 0x4900 // FORMAT_RX1        COMPARE HALFWORD (32<-16)
	OP_CHF     uint32 = 0xE3CD // FORMAT_RXY1       COMPARE HIGH (32)
	OP_CHHR    uint32 = 0xB9CD // FORMAT_RRE        COMPARE HIGH (32)
	OP_CHHSI   uint32 = 0xE554 // FORMAT_SIL        COMPARE HALFWORD IMMEDIATE (16)
	OP_CHI     uint32 = 0xA70E // FORMAT_RI1        COMPARE HALFWORD IMMEDIATE (32<-16)
	OP_CHLR    uint32 = 0xB9DD // FORMAT_RRE        COMPARE HIGH (32)
	OP_CHRL    uint32 = 0xC605 // FORMAT_RIL2       COMPARE HALFWORD RELATIVE LONG (32<-16)
	OP_CHSI    uint32 = 0xE55C // FORMAT_SIL        COMPARE HALFWORD IMMEDIATE (32<-16)
	OP_CHY     uint32 = 0xE379 // FORMAT_RXY1       COMPARE HALFWORD (32<-16)
	OP_CIB     uint32 = 0xECFE // FORMAT_RIS        COMPARE IMMEDIATE AND BRANCH (32<-8)
	OP_CIH     uint32 = 0xCC0D // FORMAT_RIL1       COMPARE IMMEDIATE HIGH (32)
	OP_CIJ     uint32 = 0xEC7E // FORMAT_RIE3       COMPARE IMMEDIATE AND BRANCH RELATIVE (32<-8)
	OP_CIT     uint32 = 0xEC72 // FORMAT_RIE1       COMPARE IMMEDIATE AND TRAP (32<-16)
	OP_CKSM    uint32 = 0xB241 // FORMAT_RRE        CHECKSUM
	OP_CL      uint32 = 0x5500 // FORMAT_RX1        COMPARE LOGICAL (32)
	OP_CLC     uint32 = 0xD500 // FORMAT_SS1        COMPARE LOGICAL (character)
	OP_CLCL    uint32 = 0x0F00 // FORMAT_RR         COMPARE LOGICAL LONG
	OP_CLCLE   uint32 = 0xA900 // FORMAT_RS1        COMPARE LOGICAL LONG EXTENDED
	OP_CLCLU   uint32 = 0xEB8F // FORMAT_RSY1       COMPARE LOGICAL LONG UNICODE
	OP_CLFDBR  uint32 = 0xB39D // FORMAT_RRF5       CONVERT TO LOGICAL (long BFP to 32)
	OP_CLFDTR  uint32 = 0xB943 // FORMAT_RRF5       CONVERT TO LOGICAL (long DFP to 32)
	OP_CLFEBR  uint32 = 0xB39C // FORMAT_RRF5       CONVERT TO LOGICAL (short BFP to 32)
	OP_CLFHSI  uint32 = 0xE55D // FORMAT_SIL        COMPARE LOGICAL IMMEDIATE (32<-16)
	OP_CLFI    uint32 = 0xC20F // FORMAT_RIL1       COMPARE LOGICAL IMMEDIATE (32)
	OP_CLFIT   uint32 = 0xEC73 // FORMAT_RIE1       COMPARE LOGICAL IMMEDIATE AND TRAP (32<-16)
	OP_CLFXBR  uint32 = 0xB39E // FORMAT_RRF5       CONVERT TO LOGICAL (extended BFP to 32)
	OP_CLFXTR  uint32 = 0xB94B // FORMAT_RRF5       CONVERT TO LOGICAL (extended DFP to 32)
	OP_CLG     uint32 = 0xE321 // FORMAT_RXY1       COMPARE LOGICAL (64)
	OP_CLGDBR  uint32 = 0xB3AD // FORMAT_RRF5       CONVERT TO LOGICAL (long BFP to 64)
	OP_CLGDTR  uint32 = 0xB942 // FORMAT_RRF5       CONVERT TO LOGICAL (long DFP to 64)
	OP_CLGEBR  uint32 = 0xB3AC // FORMAT_RRF5       CONVERT TO LOGICAL (short BFP to 64)
	OP_CLGF    uint32 = 0xE331 // FORMAT_RXY1       COMPARE LOGICAL (64<-32)
	OP_CLGFI   uint32 = 0xC20E // FORMAT_RIL1       COMPARE LOGICAL IMMEDIATE (64<-32)
	OP_CLGFR   uint32 = 0xB931 // FORMAT_RRE        COMPARE LOGICAL (64<-32)
	OP_CLGFRL  uint32 = 0xC60E // FORMAT_RIL2       COMPARE LOGICAL RELATIVE LONG (64<-32)
	OP_CLGHRL  uint32 = 0xC606 // FORMAT_RIL2       COMPARE LOGICAL RELATIVE LONG (64<-16)
	OP_CLGHSI  uint32 = 0xE559 // FORMAT_SIL        COMPARE LOGICAL IMMEDIATE (64<-16)
	OP_CLGIB   uint32 = 0xECFD // FORMAT_RIS        COMPARE LOGICAL IMMEDIATE AND BRANCH (64<-8)
	OP_CLGIJ   uint32 = 0xEC7D // FORMAT_RIE3       COMPARE LOGICAL IMMEDIATE AND BRANCH RELATIVE (64<-8)
	OP_CLGIT   uint32 = 0xEC71 // FORMAT_RIE1       COMPARE LOGICAL IMMEDIATE AND TRAP (64<-16)
	OP_CLGR    uint32 = 0xB921 // FORMAT_RRE        COMPARE LOGICAL (64)
	OP_CLGRB   uint32 = 0xECE5 // FORMAT_RRS        COMPARE LOGICAL AND BRANCH (64)
	OP_CLGRJ   uint32 = 0xEC65 // FORMAT_RIE2       COMPARE LOGICAL AND BRANCH RELATIVE (64)
	OP_CLGRL   uint32 = 0xC60A // FORMAT_RIL2       COMPARE LOGICAL RELATIVE LONG (64)
	OP_CLGRT   uint32 = 0xB961 // FORMAT_RRF3       COMPARE LOGICAL AND TRAP (64)
	OP_CLGT    uint32 = 0xEB2B // FORMAT_RSY2       COMPARE LOGICAL AND TRAP (64)
	OP_CLGXBR  uint32 = 0xB3AE // FORMAT_RRF5       CONVERT TO LOGICAL (extended BFP to 64)
	OP_CLGXTR  uint32 = 0xB94A // FORMAT_RRF5       CONVERT TO LOGICAL (extended DFP to 64)
	OP_CLHF    uint32 = 0xE3CF // FORMAT_RXY1       COMPARE LOGICAL HIGH (32)
	OP_CLHHR   uint32 = 0xB9CF // FORMAT_RRE        COMPARE LOGICAL HIGH (32)
	OP_CLHHSI  uint32 = 0xE555 // FORMAT_SIL        COMPARE LOGICAL IMMEDIATE (16)
	OP_CLHLR   uint32 = 0xB9DF // FORMAT_RRE        COMPARE LOGICAL HIGH (32)
	OP_CLHRL   uint32 = 0xC607 // FORMAT_RIL2       COMPARE LOGICAL RELATIVE LONG (32<-16)
	OP_CLI     uint32 = 0x9500 // FORMAT_SI         COMPARE LOGICAL (immediate)
	OP_CLIB    uint32 = 0xECFF // FORMAT_RIS        COMPARE LOGICAL IMMEDIATE AND BRANCH (32<-8)
	OP_CLIH    uint32 = 0xCC0F // FORMAT_RIL1       COMPARE LOGICAL IMMEDIATE HIGH (32)
	OP_CLIJ    uint32 = 0xEC7F // FORMAT_RIE3       COMPARE LOGICAL IMMEDIATE AND BRANCH RELATIVE (32<-8)
	OP_CLIY    uint32 = 0xEB55 // FORMAT_SIY        COMPARE LOGICAL (immediate)
	OP_CLM     uint32 = 0xBD00 // FORMAT_RS2        COMPARE LOGICAL CHAR. UNDER MASK (low)
	OP_CLMH    uint32 = 0xEB20 // FORMAT_RSY2       COMPARE LOGICAL CHAR. UNDER MASK (high)
	OP_CLMY    uint32 = 0xEB21 // FORMAT_RSY2       COMPARE LOGICAL CHAR. UNDER MASK (low)
	OP_CLR     uint32 = 0x1500 // FORMAT_RR         COMPARE LOGICAL (32)
	OP_CLRB    uint32 = 0xECF7 // FORMAT_RRS        COMPARE LOGICAL AND BRANCH (32)
	OP_CLRJ    uint32 = 0xEC77 // FORMAT_RIE2       COMPARE LOGICAL AND BRANCH RELATIVE (32)
	OP_CLRL    uint32 = 0xC60F // FORMAT_RIL2       COMPARE LOGICAL RELATIVE LONG (32)
	OP_CLRT    uint32 = 0xB973 // FORMAT_RRF3       COMPARE LOGICAL AND TRAP (32)
	OP_CLST    uint32 = 0xB25D // FORMAT_RRE        COMPARE LOGICAL STRING
	OP_CLT     uint32 = 0xEB23 // FORMAT_RSY2       COMPARE LOGICAL AND TRAP (32)
	OP_CLY     uint32 = 0xE355 // FORMAT_RXY1       COMPARE LOGICAL (32)
	OP_CMPSC   uint32 = 0xB263 // FORMAT_RRE        COMPRESSION CALL
	OP_CP      uint32 = 0xF900 // FORMAT_SS2        COMPARE DECIMAL
	OP_CPSDR   uint32 = 0xB372 // FORMAT_RRF2       COPY SIGN (long)
	OP_CPYA    uint32 = 0xB24D // FORMAT_RRE        COPY ACCESS
	OP_CR      uint32 = 0x1900 // FORMAT_RR         COMPARE (32)
	OP_CRB     uint32 = 0xECF6 // FORMAT_RRS        COMPARE AND BRANCH (32)
	OP_CRDTE   uint32 = 0xB98F // FORMAT_RRF2       COMPARE AND REPLACE DAT TABLE ENTRY
	OP_CRJ     uint32 = 0xEC76 // FORMAT_RIE2       COMPARE AND BRANCH RELATIVE (32)
	OP_CRL     uint32 = 0xC60D // FORMAT_RIL2       COMPARE RELATIVE LONG (32)
	OP_CRT     uint32 = 0xB972 // FORMAT_RRF3       COMPARE AND TRAP (32)
	OP_CS      uint32 = 0xBA00 // FORMAT_RS1        COMPARE AND SWAP (32)
	OP_CSCH    uint32 = 0xB230 // FORMAT_S          CLEAR SUBCHANNEL
	OP_CSDTR   uint32 = 0xB3E3 // FORMAT_RRF4       CONVERT TO SIGNED PACKED (long DFP to 64)
	OP_CSG     uint32 = 0xEB30 // FORMAT_RSY1       COMPARE AND SWAP (64)
	OP_CSP     uint32 = 0xB250 // FORMAT_RRE        COMPARE AND SWAP AND PURGE
	OP_CSPG    uint32 = 0xB98A // FORMAT_RRE        COMPARE AND SWAP AND PURGE
	OP_CSST    uint32 = 0xC802 // FORMAT_SSF        COMPARE AND SWAP AND STORE
	OP_CSXTR   uint32 = 0xB3EB // FORMAT_RRF4       CONVERT TO SIGNED PACKED (extended DFP to 128)
	OP_CSY     uint32 = 0xEB14 // FORMAT_RSY1       COMPARE AND SWAP (32)
	OP_CU12    uint32 = 0xB2A7 // FORMAT_RRF3       CONVERT UTF-8 TO UTF-16
	OP_CU14    uint32 = 0xB9B0 // FORMAT_RRF3       CONVERT UTF-8 TO UTF-32
	OP_CU21    uint32 = 0xB2A6 // FORMAT_RRF3       CONVERT UTF-16 TO UTF-8
	OP_CU24    uint32 = 0xB9B1 // FORMAT_RRF3       CONVERT UTF-16 TO UTF-32
	OP_CU41    uint32 = 0xB9B2 // FORMAT_RRE        CONVERT UTF-32 TO UTF-8
	OP_CU42    uint32 = 0xB9B3 // FORMAT_RRE        CONVERT UTF-32 TO UTF-16
	OP_CUDTR   uint32 = 0xB3E2 // FORMAT_RRE        CONVERT TO UNSIGNED PACKED (long DFP to 64)
	OP_CUSE    uint32 = 0xB257 // FORMAT_RRE        COMPARE UNTIL SUBSTRING EQUAL
	OP_CUTFU   uint32 = 0xB2A7 // FORMAT_RRF3       CONVERT UTF-8 TO UNICODE
	OP_CUUTF   uint32 = 0xB2A6 // FORMAT_RRF3       CONVERT UNICODE TO UTF-8
	OP_CUXTR   uint32 = 0xB3EA // FORMAT_RRE        CONVERT TO UNSIGNED PACKED (extended DFP to 128)
	OP_CVB     uint32 = 0x4F00 // FORMAT_RX1        CONVERT TO BINARY (32)
	OP_CVBG    uint32 = 0xE30E // FORMAT_RXY1       CONVERT TO BINARY (64)
	OP_CVBY    uint32 = 0xE306 // FORMAT_RXY1       CONVERT TO BINARY (32)
	OP_CVD     uint32 = 0x4E00 // FORMAT_RX1        CONVERT TO DECIMAL (32)
	OP_CVDG    uint32 = 0xE32E // FORMAT_RXY1       CONVERT TO DECIMAL (64)
	OP_CVDY    uint32 = 0xE326 // FORMAT_RXY1       CONVERT TO DECIMAL (32)
	OP_CXBR    uint32 = 0xB349 // FORMAT_RRE        COMPARE (extended BFP)
	OP_CXFBR   uint32 = 0xB396 // FORMAT_RRE        CONVERT FROM FIXED (32 to extended BFP)
	OP_CXFBRA  uint32 = 0xB396 // FORMAT_RRF5       CONVERT FROM FIXED (32 to extended BFP)
	OP_CXFR    uint32 = 0xB3B6 // FORMAT_RRE        CONVERT FROM FIXED (32 to extended HFP)
	OP_CXFTR   uint32 = 0xB959 // FORMAT_RRE        CONVERT FROM FIXED (32 to extended DFP)
	OP_CXGBR   uint32 = 0xB3A6 // FORMAT_RRE        CONVERT FROM FIXED (64 to extended BFP)
	OP_CXGBRA  uint32 = 0xB3A6 // FORMAT_RRF5       CONVERT FROM FIXED (64 to extended BFP)
	OP_CXGR    uint32 = 0xB3C6 // FORMAT_RRE        CONVERT FROM FIXED (64 to extended HFP)
	OP_CXGTR   uint32 = 0xB3F9 // FORMAT_RRE        CONVERT FROM FIXED (64 to extended DFP)
	OP_CXGTRA  uint32 = 0xB3F9 // FORMAT_RRF5       CONVERT FROM FIXED (64 to extended DFP)
	OP_CXLFBR  uint32 = 0xB392 // FORMAT_RRF5       CONVERT FROM LOGICAL (32 to extended BFP)
	OP_CXLFTR  uint32 = 0xB95B // FORMAT_RRF5       CONVERT FROM LOGICAL (32 to extended DFP)
	OP_CXLGBR  uint32 = 0xB3A2 // FORMAT_RRF5       CONVERT FROM LOGICAL (64 to extended BFP)
	OP_CXLGTR  uint32 = 0xB95A // FORMAT_RRF5       CONVERT FROM LOGICAL (64 to extended DFP)
	OP_CXR     uint32 = 0xB369 // FORMAT_RRE        COMPARE (extended HFP)
	OP_CXSTR   uint32 = 0xB3FB // FORMAT_RRE        CONVERT FROM SIGNED PACKED (128 to extended DFP)
	OP_CXTR    uint32 = 0xB3EC // FORMAT_RRE        COMPARE (extended DFP)
	OP_CXUTR   uint32 = 0xB3FA // FORMAT_RRE        CONVERT FROM UNSIGNED PACKED (128 to ext. DFP)
	OP_CXZT    uint32 = 0xEDAB // FORMAT_RSL        CONVERT FROM ZONED (to extended DFP)
	OP_CY      uint32 = 0xE359 // FORMAT_RXY1       COMPARE (32)
	OP_CZDT    uint32 = 0xEDA8 // FORMAT_RSL        CONVERT TO ZONED (from long DFP)
	OP_CZXT    uint32 = 0xEDA9 // FORMAT_RSL        CONVERT TO ZONED (from extended DFP)
	OP_D       uint32 = 0x5D00 // FORMAT_RX1        DIVIDE (32<-64)
	OP_DD      uint32 = 0x6D00 // FORMAT_RX1        DIVIDE (long HFP)
	OP_DDB     uint32 = 0xED1D // FORMAT_RXE        DIVIDE (long BFP)
	OP_DDBR    uint32 = 0xB31D // FORMAT_RRE        DIVIDE (long BFP)
	OP_DDR     uint32 = 0x2D00 // FORMAT_RR         DIVIDE (long HFP)
	OP_DDTR    uint32 = 0xB3D1 // FORMAT_RRF1       DIVIDE (long DFP)
	OP_DDTRA   uint32 = 0xB3D1 // FORMAT_RRF1       DIVIDE (long DFP)
	OP_DE      uint32 = 0x7D00 // FORMAT_RX1        DIVIDE (short HFP)
	OP_DEB     uint32 = 0xED0D // FORMAT_RXE        DIVIDE (short BFP)
	OP_DEBR    uint32 = 0xB30D // FORMAT_RRE        DIVIDE (short BFP)
	OP_DER     uint32 = 0x3D00 // FORMAT_RR         DIVIDE (short HFP)
	OP_DIDBR   uint32 = 0xB35B // FORMAT_RRF2       DIVIDE TO INTEGER (long BFP)
	OP_DIEBR   uint32 = 0xB353 // FORMAT_RRF2       DIVIDE TO INTEGER (short BFP)
	OP_DL      uint32 = 0xE397 // FORMAT_RXY1       DIVIDE LOGICAL (32<-64)
	OP_DLG     uint32 = 0xE387 // FORMAT_RXY1       DIVIDE LOGICAL (64<-128)
	OP_DLGR    uint32 = 0xB987 // FORMAT_RRE        DIVIDE LOGICAL (64<-128)
	OP_DLR     uint32 = 0xB997 // FORMAT_RRE        DIVIDE LOGICAL (32<-64)
	OP_DP      uint32 = 0xFD00 // FORMAT_SS2        DIVIDE DECIMAL
	OP_DR      uint32 = 0x1D00 // FORMAT_RR         DIVIDE (32<-64)
	OP_DSG     uint32 = 0xE30D // FORMAT_RXY1       DIVIDE SINGLE (64)
	OP_DSGF    uint32 = 0xE31D // FORMAT_RXY1       DIVIDE SINGLE (64<-32)
	OP_DSGFR   uint32 = 0xB91D // FORMAT_RRE        DIVIDE SINGLE (64<-32)
	OP_DSGR    uint32 = 0xB90D // FORMAT_RRE        DIVIDE SINGLE (64)
	OP_DXBR    uint32 = 0xB34D // FORMAT_RRE        DIVIDE (extended BFP)
	OP_DXR     uint32 = 0xB22D // FORMAT_RRE        DIVIDE (extended HFP)
	OP_DXTR    uint32 = 0xB3D9 // FORMAT_RRF1       DIVIDE (extended DFP)
	OP_DXTRA   uint32 = 0xB3D9 // FORMAT_RRF1       DIVIDE (extended DFP)
	OP_EAR     uint32 = 0xB24F // FORMAT_RRE        EXTRACT ACCESS
	OP_ECAG    uint32 = 0xEB4C // FORMAT_RSY1       EXTRACT CACHE ATTRIBUTE
	OP_ECTG    uint32 = 0xC801 // FORMAT_SSF        EXTRACT CPU TIME
	OP_ED      uint32 = 0xDE00 // FORMAT_SS1        EDIT
	OP_EDMK    uint32 = 0xDF00 // FORMAT_SS1        EDIT AND MARK
	OP_EEDTR   uint32 = 0xB3E5 // FORMAT_RRE        EXTRACT BIASED EXPONENT (long DFP to 64)
	OP_EEXTR   uint32 = 0xB3ED // FORMAT_RRE        EXTRACT BIASED EXPONENT (extended DFP to 64)
	OP_EFPC    uint32 = 0xB38C // FORMAT_RRE        EXTRACT FPC
	OP_EPAIR   uint32 = 0xB99A // FORMAT_RRE        EXTRACT PRIMARY ASN AND INSTANCE
	OP_EPAR    uint32 = 0xB226 // FORMAT_RRE        EXTRACT PRIMARY ASN
	OP_EPSW    uint32 = 0xB98D // FORMAT_RRE        EXTRACT PSW
	OP_EREG    uint32 = 0xB249 // FORMAT_RRE        EXTRACT STACKED REGISTERS (32)
	OP_EREGG   uint32 = 0xB90E // FORMAT_RRE        EXTRACT STACKED REGISTERS (64)
	OP_ESAIR   uint32 = 0xB99B // FORMAT_RRE        EXTRACT SECONDARY ASN AND INSTANCE
	OP_ESAR    uint32 = 0xB227 // FORMAT_RRE        EXTRACT SECONDARY ASN
	OP_ESDTR   uint32 = 0xB3E7 // FORMAT_RRE        EXTRACT SIGNIFICANCE (long DFP)
	OP_ESEA    uint32 = 0xB99D // FORMAT_RRE        EXTRACT AND SET EXTENDED AUTHORITY
	OP_ESTA    uint32 = 0xB24A // FORMAT_RRE        EXTRACT STACKED STATE
	OP_ESXTR   uint32 = 0xB3EF // FORMAT_RRE        EXTRACT SIGNIFICANCE (extended DFP)
	OP_ETND    uint32 = 0xB2EC // FORMAT_RRE        EXTRACT TRANSACTION NESTING DEPTH
	OP_EX      uint32 = 0x4400 // FORMAT_RX1        EXECUTE
	OP_EXRL    uint32 = 0xC600 // FORMAT_RIL2       EXECUTE RELATIVE LONG
	OP_FIDBR   uint32 = 0xB35F // FORMAT_RRF5       LOAD FP INTEGER (long BFP)
	OP_FIDBRA  uint32 = 0xB35F // FORMAT_RRF5       LOAD FP INTEGER (long BFP)
	OP_FIDR    uint32 = 0xB37F // FORMAT_RRE        LOAD FP INTEGER (long HFP)
	OP_FIDTR   uint32 = 0xB3D7 // FORMAT_RRF5       LOAD FP INTEGER (long DFP)
	OP_FIEBR   uint32 = 0xB357 // FORMAT_RRF5       LOAD FP INTEGER (short BFP)
	OP_FIEBRA  uint32 = 0xB357 // FORMAT_RRF5       LOAD FP INTEGER (short BFP)
	OP_FIER    uint32 = 0xB377 // FORMAT_RRE        LOAD FP INTEGER (short HFP)
	OP_FIXBR   uint32 = 0xB347 // FORMAT_RRF5       LOAD FP INTEGER (extended BFP)
	OP_FIXBRA  uint32 = 0xB347 // FORMAT_RRF5       LOAD FP INTEGER (extended BFP)
	OP_FIXR    uint32 = 0xB367 // FORMAT_RRE        LOAD FP INTEGER (extended HFP)
	OP_FIXTR   uint32 = 0xB3DF // FORMAT_RRF5       LOAD FP INTEGER (extended DFP)
	OP_FLOGR   uint32 = 0xB983 // FORMAT_RRE        FIND LEFTMOST ONE
	OP_HDR     uint32 = 0x2400 // FORMAT_RR         HALVE (long HFP)
	OP_HER     uint32 = 0x3400 // FORMAT_RR         HALVE (short HFP)
	OP_HSCH    uint32 = 0xB231 // FORMAT_S          HALT SUBCHANNEL
	OP_IAC     uint32 = 0xB224 // FORMAT_RRE        INSERT ADDRESS SPACE CONTROL
	OP_IC      uint32 = 0x4300 // FORMAT_RX1        INSERT CHARACTER
	OP_ICM     uint32 = 0xBF00 // FORMAT_RS2        INSERT CHARACTERS UNDER MASK (low)
	OP_ICMH    uint32 = 0xEB80 // FORMAT_RSY2       INSERT CHARACTERS UNDER MASK (high)
	OP_ICMY    uint32 = 0xEB81 // FORMAT_RSY2       INSERT CHARACTERS UNDER MASK (low)
	OP_ICY     uint32 = 0xE373 // FORMAT_RXY1       INSERT CHARACTER
	OP_IDTE    uint32 = 0xB98E // FORMAT_RRF2       INVALIDATE DAT TABLE ENTRY
	OP_IEDTR   uint32 = 0xB3F6 // FORMAT_RRF2       INSERT BIASED EXPONENT (64 to long DFP)
	OP_IEXTR   uint32 = 0xB3FE // FORMAT_RRF2       INSERT BIASED EXPONENT (64 to extended DFP)
	OP_IIHF    uint32 = 0xC008 // FORMAT_RIL1       INSERT IMMEDIATE (high)
	OP_IIHH    uint32 = 0xA500 // FORMAT_RI1        INSERT IMMEDIATE (high high)
	OP_IIHL    uint32 = 0xA501 // FORMAT_RI1        INSERT IMMEDIATE (high low)
	OP_IILF    uint32 = 0xC009 // FORMAT_RIL1       INSERT IMMEDIATE (low)
	OP_IILH    uint32 = 0xA502 // FORMAT_RI1        INSERT IMMEDIATE (low high)
	OP_IILL    uint32 = 0xA503 // FORMAT_RI1        INSERT IMMEDIATE (low low)
	OP_IPK     uint32 = 0xB20B // FORMAT_S          INSERT PSW KEY
	OP_IPM     uint32 = 0xB222 // FORMAT_RRE        INSERT PROGRAM MASK
	OP_IPTE    uint32 = 0xB221 // FORMAT_RRF1       INVALIDATE PAGE TABLE ENTRY
	OP_ISKE    uint32 = 0xB229 // FORMAT_RRE        INSERT STORAGE KEY EXTENDED
	OP_IVSK    uint32 = 0xB223 // FORMAT_RRE        INSERT VIRTUAL STORAGE KEY
	OP_KDB     uint32 = 0xED18 // FORMAT_RXE        COMPARE AND SIGNAL (long BFP)
	OP_KDBR    uint32 = 0xB318 // FORMAT_RRE        COMPARE AND SIGNAL (long BFP)
	OP_KDTR    uint32 = 0xB3E0 // FORMAT_RRE        COMPARE AND SIGNAL (long DFP)
	OP_KEB     uint32 = 0xED08 // FORMAT_RXE        COMPARE AND SIGNAL (short BFP)
	OP_KEBR    uint32 = 0xB308 // FORMAT_RRE        COMPARE AND SIGNAL (short BFP)
	OP_KIMD    uint32 = 0xB93E // FORMAT_RRE        COMPUTE INTERMEDIATE MESSAGE DIGEST
	OP_KLMD    uint32 = 0xB93F // FORMAT_RRE        COMPUTE LAST MESSAGE DIGEST
	OP_KM      uint32 = 0xB92E // FORMAT_RRE        CIPHER MESSAGE
	OP_KMAC    uint32 = 0xB91E // FORMAT_RRE        COMPUTE MESSAGE AUTHENTICATION CODE
	OP_KMC     uint32 = 0xB92F // FORMAT_RRE        CIPHER MESSAGE WITH CHAINING
	OP_KMCTR   uint32 = 0xB92D // FORMAT_RRF2       CIPHER MESSAGE WITH COUNTER
	OP_KMF     uint32 = 0xB92A // FORMAT_RRE        CIPHER MESSAGE WITH CFB
	OP_KMO     uint32 = 0xB92B // FORMAT_RRE        CIPHER MESSAGE WITH OFB
	OP_KXBR    uint32 = 0xB348 // FORMAT_RRE        COMPARE AND SIGNAL (extended BFP)
	OP_KXTR    uint32 = 0xB3E8 // FORMAT_RRE        COMPARE AND SIGNAL (extended DFP)
	OP_L       uint32 = 0x5800 // FORMAT_RX1        LOAD (32)
	OP_LA      uint32 = 0x4100 // FORMAT_RX1        LOAD ADDRESS
	OP_LAA     uint32 = 0xEBF8 // FORMAT_RSY1       LOAD AND ADD (32)
	OP_LAAG    uint32 = 0xEBE8 // FORMAT_RSY1       LOAD AND ADD (64)
	OP_LAAL    uint32 = 0xEBFA // FORMAT_RSY1       LOAD AND ADD LOGICAL (32)
	OP_LAALG   uint32 = 0xEBEA // FORMAT_RSY1       LOAD AND ADD LOGICAL (64)
	OP_LAE     uint32 = 0x5100 // FORMAT_RX1        LOAD ADDRESS EXTENDED
	OP_LAEY    uint32 = 0xE375 // FORMAT_RXY1       LOAD ADDRESS EXTENDED
	OP_LAM     uint32 = 0x9A00 // FORMAT_RS1        LOAD ACCESS MULTIPLE
	OP_LAMY    uint32 = 0xEB9A // FORMAT_RSY1       LOAD ACCESS MULTIPLE
	OP_LAN     uint32 = 0xEBF4 // FORMAT_RSY1       LOAD AND AND (32)
	OP_LANG    uint32 = 0xEBE4 // FORMAT_RSY1       LOAD AND AND (64)
	OP_LAO     uint32 = 0xEBF6 // FORMAT_RSY1       LOAD AND OR (32)
	OP_LAOG    uint32 = 0xEBE6 // FORMAT_RSY1       LOAD AND OR (64)
	OP_LARL    uint32 = 0xC000 // FORMAT_RIL2       LOAD ADDRESS RELATIVE LONG
	OP_LASP    uint32 = 0xE500 // FORMAT_SSE        LOAD ADDRESS SPACE PARAMETERS
	OP_LAT     uint32 = 0xE39F // FORMAT_RXY1       LOAD AND TRAP (32L<-32)
	OP_LAX     uint32 = 0xEBF7 // FORMAT_RSY1       LOAD AND EXCLUSIVE OR (32)
	OP_LAXG    uint32 = 0xEBE7 // FORMAT_RSY1       LOAD AND EXCLUSIVE OR (64)
	OP_LAY     uint32 = 0xE371 // FORMAT_RXY1       LOAD ADDRESS
	OP_LB      uint32 = 0xE376 // FORMAT_RXY1       LOAD BYTE (32)
	OP_LBH     uint32 = 0xE3C0 // FORMAT_RXY1       LOAD BYTE HIGH (32<-8)
	OP_LBR     uint32 = 0xB926 // FORMAT_RRE        LOAD BYTE (32)
	OP_LCDBR   uint32 = 0xB313 // FORMAT_RRE        LOAD COMPLEMENT (long BFP)
	OP_LCDFR   uint32 = 0xB373 // FORMAT_RRE        LOAD COMPLEMENT (long)
	OP_LCDR    uint32 = 0x2300 // FORMAT_RR         LOAD COMPLEMENT (long HFP)
	OP_LCEBR   uint32 = 0xB303 // FORMAT_RRE        LOAD COMPLEMENT (short BFP)
	OP_LCER    uint32 = 0x3300 // FORMAT_RR         LOAD COMPLEMENT (short HFP)
	OP_LCGFR   uint32 = 0xB913 // FORMAT_RRE        LOAD COMPLEMENT (64<-32)
	OP_LCGR    uint32 = 0xB903 // FORMAT_RRE        LOAD COMPLEMENT (64)
	OP_LCR     uint32 = 0x1300 // FORMAT_RR         LOAD COMPLEMENT (32)
	OP_LCTL    uint32 = 0xB700 // FORMAT_RS1        LOAD CONTROL (32)
	OP_LCTLG   uint32 = 0xEB2F // FORMAT_RSY1       LOAD CONTROL (64)
	OP_LCXBR   uint32 = 0xB343 // FORMAT_RRE        LOAD COMPLEMENT (extended BFP)
	OP_LCXR    uint32 = 0xB363 // FORMAT_RRE        LOAD COMPLEMENT (extended HFP)
	OP_LD      uint32 = 0x6800 // FORMAT_RX1        LOAD (long)
	OP_LDE     uint32 = 0xED24 // FORMAT_RXE        LOAD LENGTHENED (short to long HFP)
	OP_LDEB    uint32 = 0xED04 // FORMAT_RXE        LOAD LENGTHENED (short to long BFP)
	OP_LDEBR   uint32 = 0xB304 // FORMAT_RRE        LOAD LENGTHENED (short to long BFP)
	OP_LDER    uint32 = 0xB324 // FORMAT_RRE        LOAD LENGTHENED (short to long HFP)
	OP_LDETR   uint32 = 0xB3D4 // FORMAT_RRF4       LOAD LENGTHENED (short to long DFP)
	OP_LDGR    uint32 = 0xB3C1 // FORMAT_RRE        LOAD FPR FROM GR (64 to long)
	OP_LDR     uint32 = 0x2800 // FORMAT_RR         LOAD (long)
	OP_LDXBR   uint32 = 0xB345 // FORMAT_RRE        LOAD ROUNDED (extended to long BFP)
	OP_LDXBRA  uint32 = 0xB345 // FORMAT_RRF5       LOAD ROUNDED (extended to long BFP)
	OP_LDXR    uint32 = 0x2500 // FORMAT_RR         LOAD ROUNDED (extended to long HFP)
	OP_LDXTR   uint32 = 0xB3DD // FORMAT_RRF5       LOAD ROUNDED (extended to long DFP)
	OP_LDY     uint32 = 0xED65 // FORMAT_RXY1       LOAD (long)
	OP_LE      uint32 = 0x7800 // FORMAT_RX1        LOAD (short)
	OP_LEDBR   uint32 = 0xB344 // FORMAT_RRE        LOAD ROUNDED (long to short BFP)
	OP_LEDBRA  uint32 = 0xB344 // FORMAT_RRF5       LOAD ROUNDED (long to short BFP)
	OP_LEDR    uint32 = 0x3500 // FORMAT_RR         LOAD ROUNDED (long to short HFP)
	OP_LEDTR   uint32 = 0xB3D5 // FORMAT_RRF5       LOAD ROUNDED (long to short DFP)
	OP_LER     uint32 = 0x3800 // FORMAT_RR         LOAD (short)
	OP_LEXBR   uint32 = 0xB346 // FORMAT_RRE        LOAD ROUNDED (extended to short BFP)
	OP_LEXBRA  uint32 = 0xB346 // FORMAT_RRF5       LOAD ROUNDED (extended to short BFP)
	OP_LEXR    uint32 = 0xB366 // FORMAT_RRE        LOAD ROUNDED (extended to short HFP)
	OP_LEY     uint32 = 0xED64 // FORMAT_RXY1       LOAD (short)
	OP_LFAS    uint32 = 0xB2BD // FORMAT_S          LOAD FPC AND SIGNAL
	OP_LFH     uint32 = 0xE3CA // FORMAT_RXY1       LOAD HIGH (32)
	OP_LFHAT   uint32 = 0xE3C8 // FORMAT_RXY1       LOAD HIGH AND TRAP (32H<-32)
	OP_LFPC    uint32 = 0xB29D // FORMAT_S          LOAD FPC
	OP_LG      uint32 = 0xE304 // FORMAT_RXY1       LOAD (64)
	OP_LGAT    uint32 = 0xE385 // FORMAT_RXY1       LOAD AND TRAP (64)
	OP_LGB     uint32 = 0xE377 // FORMAT_RXY1       LOAD BYTE (64)
	OP_LGBR    uint32 = 0xB906 // FORMAT_RRE        LOAD BYTE (64)
	OP_LGDR    uint32 = 0xB3CD // FORMAT_RRE        LOAD GR FROM FPR (long to 64)
	OP_LGF     uint32 = 0xE314 // FORMAT_RXY1       LOAD (64<-32)
	OP_LGFI    uint32 = 0xC001 // FORMAT_RIL1       LOAD IMMEDIATE (64<-32)
	OP_LGFR    uint32 = 0xB914 // FORMAT_RRE        LOAD (64<-32)
	OP_LGFRL   uint32 = 0xC40C // FORMAT_RIL2       LOAD RELATIVE LONG (64<-32)
	OP_LGH     uint32 = 0xE315 // FORMAT_RXY1       LOAD HALFWORD (64)
	OP_LGHI    uint32 = 0xA709 // FORMAT_RI1        LOAD HALFWORD IMMEDIATE (64)
	OP_LGHR    uint32 = 0xB907 // FORMAT_RRE        LOAD HALFWORD (64)
	OP_LGHRL   uint32 = 0xC404 // FORMAT_RIL2       LOAD HALFWORD RELATIVE LONG (64<-16)
	OP_LGR     uint32 = 0xB904 // FORMAT_RRE        LOAD (64)
	OP_LGRL    uint32 = 0xC408 // FORMAT_RIL2       LOAD RELATIVE LONG (64)
	OP_LH      uint32 = 0x4800 // FORMAT_RX1        LOAD HALFWORD (32)
	OP_LHH     uint32 = 0xE3C4 // FORMAT_RXY1       LOAD HALFWORD HIGH (32<-16)
	OP_LHI     uint32 = 0xA708 // FORMAT_RI1        LOAD HALFWORD IMMEDIATE (32)
	OP_LHR     uint32 = 0xB927 // FORMAT_RRE        LOAD HALFWORD (32)
	OP_LHRL    uint32 = 0xC405 // FORMAT_RIL2       LOAD HALFWORD RELATIVE LONG (32<-16)
	OP_LHY     uint32 = 0xE378 // FORMAT_RXY1       LOAD HALFWORD (32)
	OP_LLC     uint32 = 0xE394 // FORMAT_RXY1       LOAD LOGICAL CHARACTER (32)
	OP_LLCH    uint32 = 0xE3C2 // FORMAT_RXY1       LOAD LOGICAL CHARACTER HIGH (32<-8)
	OP_LLCR    uint32 = 0xB994 // FORMAT_RRE        LOAD LOGICAL CHARACTER (32)
	OP_LLGC    uint32 = 0xE390 // FORMAT_RXY1       LOAD LOGICAL CHARACTER (64)
	OP_LLGCR   uint32 = 0xB984 // FORMAT_RRE        LOAD LOGICAL CHARACTER (64)
	OP_LLGF    uint32 = 0xE316 // FORMAT_RXY1       LOAD LOGICAL (64<-32)
	OP_LLGFAT  uint32 = 0xE39D // FORMAT_RXY1       LOAD LOGICAL AND TRAP (64<-32)
	OP_LLGFR   uint32 = 0xB916 // FORMAT_RRE        LOAD LOGICAL (64<-32)
	OP_LLGFRL  uint32 = 0xC40E // FORMAT_RIL2       LOAD LOGICAL RELATIVE LONG (64<-32)
	OP_LLGH    uint32 = 0xE391 // FORMAT_RXY1       LOAD LOGICAL HALFWORD (64)
	OP_LLGHR   uint32 = 0xB985 // FORMAT_RRE        LOAD LOGICAL HALFWORD (64)
	OP_LLGHRL  uint32 = 0xC406 // FORMAT_RIL2       LOAD LOGICAL HALFWORD RELATIVE LONG (64<-16)
	OP_LLGT    uint32 = 0xE317 // FORMAT_RXY1       LOAD LOGICAL THIRTY ONE BITS
	OP_LLGTAT  uint32 = 0xE39C // FORMAT_RXY1       LOAD LOGICAL THIRTY ONE BITS AND TRAP (64<-31)
	OP_LLGTR   uint32 = 0xB917 // FORMAT_RRE        LOAD LOGICAL THIRTY ONE BITS
	OP_LLH     uint32 = 0xE395 // FORMAT_RXY1       LOAD LOGICAL HALFWORD (32)
	OP_LLHH    uint32 = 0xE3C6 // FORMAT_RXY1       LOAD LOGICAL HALFWORD HIGH (32<-16)
	OP_LLHR    uint32 = 0xB995 // FORMAT_RRE        LOAD LOGICAL HALFWORD (32)
	OP_LLHRL   uint32 = 0xC402 // FORMAT_RIL2       LOAD LOGICAL HALFWORD RELATIVE LONG (32<-16)
	OP_LLIHF   uint32 = 0xC00E // FORMAT_RIL1       LOAD LOGICAL IMMEDIATE (high)
	OP_LLIHH   uint32 = 0xA50C // FORMAT_RI1        LOAD LOGICAL IMMEDIATE (high high)
	OP_LLIHL   uint32 = 0xA50D // FORMAT_RI1        LOAD LOGICAL IMMEDIATE (high low)
	OP_LLILF   uint32 = 0xC00F // FORMAT_RIL1       LOAD LOGICAL IMMEDIATE (low)
	OP_LLILH   uint32 = 0xA50E // FORMAT_RI1        LOAD LOGICAL IMMEDIATE (low high)
	OP_LLILL   uint32 = 0xA50F // FORMAT_RI1        LOAD LOGICAL IMMEDIATE (low low)
	OP_LM      uint32 = 0x9800 // FORMAT_RS1        LOAD MULTIPLE (32)
	OP_LMD     uint32 = 0xEF00 // FORMAT_SS5        LOAD MULTIPLE DISJOINT
	OP_LMG     uint32 = 0xEB04 // FORMAT_RSY1       LOAD MULTIPLE (64)
	OP_LMH     uint32 = 0xEB96 // FORMAT_RSY1       LOAD MULTIPLE HIGH
	OP_LMY     uint32 = 0xEB98 // FORMAT_RSY1       LOAD MULTIPLE (32)
	OP_LNDBR   uint32 = 0xB311 // FORMAT_RRE        LOAD NEGATIVE (long BFP)
	OP_LNDFR   uint32 = 0xB371 // FORMAT_RRE        LOAD NEGATIVE (long)
	OP_LNDR    uint32 = 0x2100 // FORMAT_RR         LOAD NEGATIVE (long HFP)
	OP_LNEBR   uint32 = 0xB301 // FORMAT_RRE        LOAD NEGATIVE (short BFP)
	OP_LNER    uint32 = 0x3100 // FORMAT_RR         LOAD NEGATIVE (short HFP)
	OP_LNGFR   uint32 = 0xB911 // FORMAT_RRE        LOAD NEGATIVE (64<-32)
	OP_LNGR    uint32 = 0xB901 // FORMAT_RRE        LOAD NEGATIVE (64)
	OP_LNR     uint32 = 0x1100 // FORMAT_RR         LOAD NEGATIVE (32)
	OP_LNXBR   uint32 = 0xB341 // FORMAT_RRE        LOAD NEGATIVE (extended BFP)
	OP_LNXR    uint32 = 0xB361 // FORMAT_RRE        LOAD NEGATIVE (extended HFP)
	OP_LOC     uint32 = 0xEBF2 // FORMAT_RSY2       LOAD ON CONDITION (32)
	OP_LOCG    uint32 = 0xEBE2 // FORMAT_RSY2       LOAD ON CONDITION (64)
	OP_LOCGR   uint32 = 0xB9E2 // FORMAT_RRF3       LOAD ON CONDITION (64)
	OP_LOCR    uint32 = 0xB9F2 // FORMAT_RRF3       LOAD ON CONDITION (32)
	OP_LPD     uint32 = 0xC804 // FORMAT_SSF        LOAD PAIR DISJOINT (32)
	OP_LPDBR   uint32 = 0xB310 // FORMAT_RRE        LOAD POSITIVE (long BFP)
	OP_LPDFR   uint32 = 0xB370 // FORMAT_RRE        LOAD POSITIVE (long)
	OP_LPDG    uint32 = 0xC805 // FORMAT_SSF        LOAD PAIR DISJOINT (64)
	OP_LPDR    uint32 = 0x2000 // FORMAT_RR         LOAD POSITIVE (long HFP)
	OP_LPEBR   uint32 = 0xB300 // FORMAT_RRE        LOAD POSITIVE (short BFP)
	OP_LPER    uint32 = 0x3000 // FORMAT_RR         LOAD POSITIVE (short HFP)
	OP_LPGFR   uint32 = 0xB910 // FORMAT_RRE        LOAD POSITIVE (64<-32)
	OP_LPGR    uint32 = 0xB900 // FORMAT_RRE        LOAD POSITIVE (64)
	OP_LPQ     uint32 = 0xE38F // FORMAT_RXY1       LOAD PAIR FROM QUADWORD
	OP_LPR     uint32 = 0x1000 // FORMAT_RR         LOAD POSITIVE (32)
	OP_LPSW    uint32 = 0x8200 // FORMAT_S          LOAD PSW
	OP_LPSWE   uint32 = 0xB2B2 // FORMAT_S          LOAD PSW EXTENDED
	OP_LPTEA   uint32 = 0xB9AA // FORMAT_RRF2       LOAD PAGE TABLE ENTRY ADDRESS
	OP_LPXBR   uint32 = 0xB340 // FORMAT_RRE        LOAD POSITIVE (extended BFP)
	OP_LPXR    uint32 = 0xB360 // FORMAT_RRE        LOAD POSITIVE (extended HFP)
	OP_LR      uint32 = 0x1800 // FORMAT_RR         LOAD (32)
	OP_LRA     uint32 = 0xB100 // FORMAT_RX1        LOAD REAL ADDRESS (32)
	OP_LRAG    uint32 = 0xE303 // FORMAT_RXY1       LOAD REAL ADDRESS (64)
	OP_LRAY    uint32 = 0xE313 // FORMAT_RXY1       LOAD REAL ADDRESS (32)
	OP_LRDR    uint32 = 0x2500 // FORMAT_RR         LOAD ROUNDED (extended to long HFP)
	OP_LRER    uint32 = 0x3500 // FORMAT_RR         LOAD ROUNDED (long to short HFP)
	OP_LRL     uint32 = 0xC40D // FORMAT_RIL2       LOAD RELATIVE LONG (32)
	OP_LRV     uint32 = 0xE31E // FORMAT_RXY1       LOAD REVERSED (32)
	OP_LRVG    uint32 = 0xE30F // FORMAT_RXY1       LOAD REVERSED (64)
	OP_LRVGR   uint32 = 0xB90F // FORMAT_RRE        LOAD REVERSED (64)
	OP_LRVH    uint32 = 0xE31F // FORMAT_RXY1       LOAD REVERSED (16)
	OP_LRVR    uint32 = 0xB91F // FORMAT_RRE        LOAD REVERSED (32)
	OP_LT      uint32 = 0xE312 // FORMAT_RXY1       LOAD AND TEST (32)
	OP_LTDBR   uint32 = 0xB312 // FORMAT_RRE        LOAD AND TEST (long BFP)
	OP_LTDR    uint32 = 0x2200 // FORMAT_RR         LOAD AND TEST (long HFP)
	OP_LTDTR   uint32 = 0xB3D6 // FORMAT_RRE        LOAD AND TEST (long DFP)
	OP_LTEBR   uint32 = 0xB302 // FORMAT_RRE        LOAD AND TEST (short BFP)
	OP_LTER    uint32 = 0x3200 // FORMAT_RR         LOAD AND TEST (short HFP)
	OP_LTG     uint32 = 0xE302 // FORMAT_RXY1       LOAD AND TEST (64)
	OP_LTGF    uint32 = 0xE332 // FORMAT_RXY1       LOAD AND TEST (64<-32)
	OP_LTGFR   uint32 = 0xB912 // FORMAT_RRE        LOAD AND TEST (64<-32)
	OP_LTGR    uint32 = 0xB902 // FORMAT_RRE        LOAD AND TEST (64)
	OP_LTR     uint32 = 0x1200 // FORMAT_RR         LOAD AND TEST (32)
	OP_LTXBR   uint32 = 0xB342 // FORMAT_RRE        LOAD AND TEST (extended BFP)
	OP_LTXR    uint32 = 0xB362 // FORMAT_RRE        LOAD AND TEST (extended HFP)
	OP_LTXTR   uint32 = 0xB3DE // FORMAT_RRE        LOAD AND TEST (extended DFP)
	OP_LURA    uint32 = 0xB24B // FORMAT_RRE        LOAD USING REAL ADDRESS (32)
	OP_LURAG   uint32 = 0xB905 // FORMAT_RRE        LOAD USING REAL ADDRESS (64)
	OP_LXD     uint32 = 0xED25 // FORMAT_RXE        LOAD LENGTHENED (long to extended HFP)
	OP_LXDB    uint32 = 0xED05 // FORMAT_RXE        LOAD LENGTHENED (long to extended BFP)
	OP_LXDBR   uint32 = 0xB305 // FORMAT_RRE        LOAD LENGTHENED (long to extended BFP)
	OP_LXDR    uint32 = 0xB325 // FORMAT_RRE        LOAD LENGTHENED (long to extended HFP)
	OP_LXDTR   uint32 = 0xB3DC // FORMAT_RRF4       LOAD LENGTHENED (long to extended DFP)
	OP_LXE     uint32 = 0xED26 // FORMAT_RXE        LOAD LENGTHENED (short to extended HFP)
	OP_LXEB    uint32 = 0xED06 // FORMAT_RXE        LOAD LENGTHENED (short to extended BFP)
	OP_LXEBR   uint32 = 0xB306 // FORMAT_RRE        LOAD LENGTHENED (short to extended BFP)
	OP_LXER    uint32 = 0xB326 // FORMAT_RRE        LOAD LENGTHENED (short to extended HFP)
	OP_LXR     uint32 = 0xB365 // FORMAT_RRE        LOAD (extended)
	OP_LY      uint32 = 0xE358 // FORMAT_RXY1       LOAD (32)
	OP_LZDR    uint32 = 0xB375 // FORMAT_RRE        LOAD ZERO (long)
	OP_LZER    uint32 = 0xB374 // FORMAT_RRE        LOAD ZERO (short)
	OP_LZXR    uint32 = 0xB376 // FORMAT_RRE        LOAD ZERO (extended)
	OP_M       uint32 = 0x5C00 // FORMAT_RX1        MULTIPLY (64<-32)
	OP_MAD     uint32 = 0xED3E // FORMAT_RXF        MULTIPLY AND ADD (long HFP)
	OP_MADB    uint32 = 0xED1E // FORMAT_RXF        MULTIPLY AND ADD (long BFP)
	OP_MADBR   uint32 = 0xB31E // FORMAT_RRD        MULTIPLY AND ADD (long BFP)
	OP_MADR    uint32 = 0xB33E // FORMAT_RRD        MULTIPLY AND ADD (long HFP)
	OP_MAE     uint32 = 0xED2E // FORMAT_RXF        MULTIPLY AND ADD (short HFP)
	OP_MAEB    uint32 = 0xED0E // FORMAT_RXF        MULTIPLY AND ADD (short BFP)
	OP_MAEBR   uint32 = 0xB30E // FORMAT_RRD        MULTIPLY AND ADD (short BFP)
	OP_MAER    uint32 = 0xB32E // FORMAT_RRD        MULTIPLY AND ADD (short HFP)
	OP_MAY     uint32 = 0xED3A // FORMAT_RXF        MULTIPLY & ADD UNNORMALIZED (long to ext. HFP)
	OP_MAYH    uint32 = 0xED3C // FORMAT_RXF        MULTIPLY AND ADD UNNRM. (long to ext. high HFP)
	OP_MAYHR   uint32 = 0xB33C // FORMAT_RRD        MULTIPLY AND ADD UNNRM. (long to ext. high HFP)
	OP_MAYL    uint32 = 0xED38 // FORMAT_RXF        MULTIPLY AND ADD UNNRM. (long to ext. low HFP)
	OP_MAYLR   uint32 = 0xB338 // FORMAT_RRD        MULTIPLY AND ADD UNNRM. (long to ext. low HFP)
	OP_MAYR    uint32 = 0xB33A // FORMAT_RRD        MULTIPLY & ADD UNNORMALIZED (long to ext. HFP)
	OP_MC      uint32 = 0xAF00 // FORMAT_SI         MONITOR CALL
	OP_MD      uint32 = 0x6C00 // FORMAT_RX1        MULTIPLY (long HFP)
	OP_MDB     uint32 = 0xED1C // FORMAT_RXE        MULTIPLY (long BFP)
	OP_MDBR    uint32 = 0xB31C // FORMAT_RRE        MULTIPLY (long BFP)
	OP_MDE     uint32 = 0x7C00 // FORMAT_RX1        MULTIPLY (short to long HFP)
	OP_MDEB    uint32 = 0xED0C // FORMAT_RXE        MULTIPLY (short to long BFP)
	OP_MDEBR   uint32 = 0xB30C // FORMAT_RRE        MULTIPLY (short to long BFP)
	OP_MDER    uint32 = 0x3C00 // FORMAT_RR         MULTIPLY (short to long HFP)
	OP_MDR     uint32 = 0x2C00 // FORMAT_RR         MULTIPLY (long HFP)
	OP_MDTR    uint32 = 0xB3D0 // FORMAT_RRF1       MULTIPLY (long DFP)
	OP_MDTRA   uint32 = 0xB3D0 // FORMAT_RRF1       MULTIPLY (long DFP)
	OP_ME      uint32 = 0x7C00 // FORMAT_RX1        MULTIPLY (short to long HFP)
	OP_MEE     uint32 = 0xED37 // FORMAT_RXE        MULTIPLY (short HFP)
	OP_MEEB    uint32 = 0xED17 // FORMAT_RXE        MULTIPLY (short BFP)
	OP_MEEBR   uint32 = 0xB317 // FORMAT_RRE        MULTIPLY (short BFP)
	OP_MEER    uint32 = 0xB337 // FORMAT_RRE        MULTIPLY (short HFP)
	OP_MER     uint32 = 0x3C00 // FORMAT_RR         MULTIPLY (short to long HFP)
	OP_MFY     uint32 = 0xE35C // FORMAT_RXY1       MULTIPLY (64<-32)
	OP_MGHI    uint32 = 0xA70D // FORMAT_RI1        MULTIPLY HALFWORD IMMEDIATE (64)
	OP_MH      uint32 = 0x4C00 // FORMAT_RX1        MULTIPLY HALFWORD (32)
	OP_MHI     uint32 = 0xA70C // FORMAT_RI1        MULTIPLY HALFWORD IMMEDIATE (32)
	OP_MHY     uint32 = 0xE37C // FORMAT_RXY1       MULTIPLY HALFWORD (32)
	OP_ML      uint32 = 0xE396 // FORMAT_RXY1       MULTIPLY LOGICAL (64<-32)
	OP_MLG     uint32 = 0xE386 // FORMAT_RXY1       MULTIPLY LOGICAL (128<-64)
	OP_MLGR    uint32 = 0xB986 // FORMAT_RRE        MULTIPLY LOGICAL (128<-64)
	OP_MLR     uint32 = 0xB996 // FORMAT_RRE        MULTIPLY LOGICAL (64<-32)
	OP_MP      uint32 = 0xFC00 // FORMAT_SS2        MULTIPLY DECIMAL
	OP_MR      uint32 = 0x1C00 // FORMAT_RR         MULTIPLY (64<-32)
	OP_MS      uint32 = 0x7100 // FORMAT_RX1        MULTIPLY SINGLE (32)
	OP_MSCH    uint32 = 0xB232 // FORMAT_S          MODIFY SUBCHANNEL
	OP_MSD     uint32 = 0xED3F // FORMAT_RXF        MULTIPLY AND SUBTRACT (long HFP)
	OP_MSDB    uint32 = 0xED1F // FORMAT_RXF        MULTIPLY AND SUBTRACT (long BFP)
	OP_MSDBR   uint32 = 0xB31F // FORMAT_RRD        MULTIPLY AND SUBTRACT (long BFP)
	OP_MSDR    uint32 = 0xB33F // FORMAT_RRD        MULTIPLY AND SUBTRACT (long HFP)
	OP_MSE     uint32 = 0xED2F // FORMAT_RXF        MULTIPLY AND SUBTRACT (short HFP)
	OP_MSEB    uint32 = 0xED0F // FORMAT_RXF        MULTIPLY AND SUBTRACT (short BFP)
	OP_MSEBR   uint32 = 0xB30F // FORMAT_RRD        MULTIPLY AND SUBTRACT (short BFP)
	OP_MSER    uint32 = 0xB32F // FORMAT_RRD        MULTIPLY AND SUBTRACT (short HFP)
	OP_MSFI    uint32 = 0xC201 // FORMAT_RIL1       MULTIPLY SINGLE IMMEDIATE (32)
	OP_MSG     uint32 = 0xE30C // FORMAT_RXY1       MULTIPLY SINGLE (64)
	OP_MSGF    uint32 = 0xE31C // FORMAT_RXY1       MULTIPLY SINGLE (64<-32)
	OP_MSGFI   uint32 = 0xC200 // FORMAT_RIL1       MULTIPLY SINGLE IMMEDIATE (64<-32)
	OP_MSGFR   uint32 = 0xB91C // FORMAT_RRE        MULTIPLY SINGLE (64<-32)
	OP_MSGR    uint32 = 0xB90C // FORMAT_RRE        MULTIPLY SINGLE (64)
	OP_MSR     uint32 = 0xB252 // FORMAT_RRE        MULTIPLY SINGLE (32)
	OP_MSTA    uint32 = 0xB247 // FORMAT_RRE        MODIFY STACKED STATE
	OP_MSY     uint32 = 0xE351 // FORMAT_RXY1       MULTIPLY SINGLE (32)
	OP_MVC     uint32 = 0xD200 // FORMAT_SS1        MOVE (character)
	OP_MVCDK   uint32 = 0xE50F // FORMAT_SSE        MOVE WITH DESTINATION KEY
	OP_MVCIN   uint32 = 0xE800 // FORMAT_SS1        MOVE INVERSE
	OP_MVCK    uint32 = 0xD900 // FORMAT_SS4        MOVE WITH KEY
	OP_MVCL    uint32 = 0x0E00 // FORMAT_RR         MOVE LONG
	OP_MVCLE   uint32 = 0xA800 // FORMAT_RS1        MOVE LONG EXTENDED
	OP_MVCLU   uint32 = 0xEB8E // FORMAT_RSY1       MOVE LONG UNICODE
	OP_MVCOS   uint32 = 0xC800 // FORMAT_SSF        MOVE WITH OPTIONAL SPECIFICATIONS
	OP_MVCP    uint32 = 0xDA00 // FORMAT_SS4        MOVE TO PRIMARY
	OP_MVCS    uint32 = 0xDB00 // FORMAT_SS4        MOVE TO SECONDARY
	OP_MVCSK   uint32 = 0xE50E // FORMAT_SSE        MOVE WITH SOURCE KEY
	OP_MVGHI   uint32 = 0xE548 // FORMAT_SIL        MOVE (64<-16)
	OP_MVHHI   uint32 = 0xE544 // FORMAT_SIL        MOVE (16<-16)
	OP_MVHI    uint32 = 0xE54C // FORMAT_SIL        MOVE (32<-16)
	OP_MVI     uint32 = 0x9200 // FORMAT_SI         MOVE (immediate)
	OP_MVIY    uint32 = 0xEB52 // FORMAT_SIY        MOVE (immediate)
	OP_MVN     uint32 = 0xD100 // FORMAT_SS1        MOVE NUMERICS
	OP_MVO     uint32 = 0xF100 // FORMAT_SS2        MOVE WITH OFFSET
	OP_MVPG    uint32 = 0xB254 // FORMAT_RRE        MOVE PAGE
	OP_MVST    uint32 = 0xB255 // FORMAT_RRE        MOVE STRING
	OP_MVZ     uint32 = 0xD300 // FORMAT_SS1        MOVE ZONES
	OP_MXBR    uint32 = 0xB34C // FORMAT_RRE        MULTIPLY (extended BFP)
	OP_MXD     uint32 = 0x6700 // FORMAT_RX1        MULTIPLY (long to extended HFP)
	OP_MXDB    uint32 = 0xED07 // FORMAT_RXE        MULTIPLY (long to extended BFP)
	OP_MXDBR   uint32 = 0xB307 // FORMAT_RRE        MULTIPLY (long to extended BFP)
	OP_MXDR    uint32 = 0x2700 // FORMAT_RR         MULTIPLY (long to extended HFP)
	OP_MXR     uint32 = 0x2600 // FORMAT_RR         MULTIPLY (extended HFP)
	OP_MXTR    uint32 = 0xB3D8 // FORMAT_RRF1       MULTIPLY (extended DFP)
	OP_MXTRA   uint32 = 0xB3D8 // FORMAT_RRF1       MULTIPLY (extended DFP)
	OP_MY      uint32 = 0xED3B // FORMAT_RXF        MULTIPLY UNNORMALIZED (long to ext. HFP)
	OP_MYH     uint32 = 0xED3D // FORMAT_RXF        MULTIPLY UNNORM. (long to ext. high HFP)
	OP_MYHR    uint32 = 0xB33D // FORMAT_RRD        MULTIPLY UNNORM. (long to ext. high HFP)
	OP_MYL     uint32 = 0xED39 // FORMAT_RXF        MULTIPLY UNNORM. (long to ext. low HFP)
	OP_MYLR    uint32 = 0xB339 // FORMAT_RRD        MULTIPLY UNNORM. (long to ext. low HFP)
	OP_MYR     uint32 = 0xB33B // FORMAT_RRD        MULTIPLY UNNORMALIZED (long to ext. HFP)
	OP_N       uint32 = 0x5400 // FORMAT_RX1        AND (32)
	OP_NC      uint32 = 0xD400 // FORMAT_SS1        AND (character)
	OP_NG      uint32 = 0xE380 // FORMAT_RXY1       AND (64)
	OP_NGR     uint32 = 0xB980 // FORMAT_RRE        AND (64)
	OP_NGRK    uint32 = 0xB9E4 // FORMAT_RRF1       AND (64)
	OP_NI      uint32 = 0x9400 // FORMAT_SI         AND (immediate)
	OP_NIAI    uint32 = 0xB2FA // FORMAT_IE         NEXT INSTRUCTION ACCESS INTENT
	OP_NIHF    uint32 = 0xC00A // FORMAT_RIL1       AND IMMEDIATE (high)
	OP_NIHH    uint32 = 0xA504 // FORMAT_RI1        AND IMMEDIATE (high high)
	OP_NIHL    uint32 = 0xA505 // FORMAT_RI1        AND IMMEDIATE (high low)
	OP_NILF    uint32 = 0xC00B // FORMAT_RIL1       AND IMMEDIATE (low)
	OP_NILH    uint32 = 0xA506 // FORMAT_RI1        AND IMMEDIATE (low high)
	OP_NILL    uint32 = 0xA507 // FORMAT_RI1        AND IMMEDIATE (low low)
	OP_NIY     uint32 = 0xEB54 // FORMAT_SIY        AND (immediate)
	OP_NR      uint32 = 0x1400 // FORMAT_RR         AND (32)
	OP_NRK     uint32 = 0xB9F4 // FORMAT_RRF1       AND (32)
	OP_NTSTG   uint32 = 0xE325 // FORMAT_RXY1       NONTRANSACTIONAL STORE
	OP_NY      uint32 = 0xE354 // FORMAT_RXY1       AND (32)
	OP_O       uint32 = 0x5600 // FORMAT_RX1        OR (32)
	OP_OC      uint32 = 0xD600 // FORMAT_SS1        OR (character)
	OP_OG      uint32 = 0xE381 // FORMAT_RXY1       OR (64)
	OP_OGR     uint32 = 0xB981 // FORMAT_RRE        OR (64)
	OP_OGRK    uint32 = 0xB9E6 // FORMAT_RRF1       OR (64)
	OP_OI      uint32 = 0x9600 // FORMAT_SI         OR (immediate)
	OP_OIHF    uint32 = 0xC00C // FORMAT_RIL1       OR IMMEDIATE (high)
	OP_OIHH    uint32 = 0xA508 // FORMAT_RI1        OR IMMEDIATE (high high)
	OP_OIHL    uint32 = 0xA509 // FORMAT_RI1        OR IMMEDIATE (high low)
	OP_OILF    uint32 = 0xC00D // FORMAT_RIL1       OR IMMEDIATE (low)
	OP_OILH    uint32 = 0xA50A // FORMAT_RI1        OR IMMEDIATE (low high)
	OP_OILL    uint32 = 0xA50B // FORMAT_RI1        OR IMMEDIATE (low low)
	OP_OIY     uint32 = 0xEB56 // FORMAT_SIY        OR (immediate)
	OP_OR      uint32 = 0x1600 // FORMAT_RR         OR (32)
	OP_ORK     uint32 = 0xB9F6 // FORMAT_RRF1       OR (32)
	OP_OY      uint32 = 0xE356 // FORMAT_RXY1       OR (32)
	OP_PACK    uint32 = 0xF200 // FORMAT_SS2        PACK
	OP_PALB    uint32 = 0xB248 // FORMAT_RRE        PURGE ALB
	OP_PC      uint32 = 0xB218 // FORMAT_S          PROGRAM CALL
	OP_PCC     uint32 = 0xB92C // FORMAT_RRE        PERFORM CRYPTOGRAPHIC COMPUTATION
	OP_PCKMO   uint32 = 0xB928 // FORMAT_RRE        PERFORM CRYPTOGRAPHIC KEY MGMT. OPERATIONS
	OP_PFD     uint32 = 0xE336 // FORMAT_RXY2       PREFETCH DATA
	OP_PFDRL   uint32 = 0xC602 // FORMAT_RIL3       PREFETCH DATA RELATIVE LONG
	OP_PFMF    uint32 = 0xB9AF // FORMAT_RRE        PERFORM FRAME MANAGEMENT FUNCTION
	OP_PFPO    uint32 = 0x010A // FORMAT_E          PERFORM FLOATING-POINT OPERATION
	OP_PGIN    uint32 = 0xB22E // FORMAT_RRE        PAGE IN
	OP_PGOUT   uint32 = 0xB22F // FORMAT_RRE        PAGE OUT
	OP_PKA     uint32 = 0xE900 // FORMAT_SS6        PACK ASCII
	OP_PKU     uint32 = 0xE100 // FORMAT_SS6        PACK UNICODE
	OP_PLO     uint32 = 0xEE00 // FORMAT_SS5        PERFORM LOCKED OPERATION
	OP_POPCNT  uint32 = 0xB9E1 // FORMAT_RRE        POPULATION COUNT
	OP_PPA     uint32 = 0xB2E8 // FORMAT_RRF3       PERFORM PROCESSOR ASSIST
	OP_PR      uint32 = 0x0101 // FORMAT_E          PROGRAM RETURN
	OP_PT      uint32 = 0xB228 // FORMAT_RRE        PROGRAM TRANSFER
	OP_PTF     uint32 = 0xB9A2 // FORMAT_RRE        PERFORM TOPOLOGY FUNCTION
	OP_PTFF    uint32 = 0x0104 // FORMAT_E          PERFORM TIMING FACILITY FUNCTION
	OP_PTI     uint32 = 0xB99E // FORMAT_RRE        PROGRAM TRANSFER WITH INSTANCE
	OP_PTLB    uint32 = 0xB20D // FORMAT_S          PURGE TLB
	OP_QADTR   uint32 = 0xB3F5 // FORMAT_RRF2       QUANTIZE (long DFP)
	OP_QAXTR   uint32 = 0xB3FD // FORMAT_RRF2       QUANTIZE (extended DFP)
	OP_RCHP    uint32 = 0xB23B // FORMAT_S          RESET CHANNEL PATH
	OP_RISBG   uint32 = 0xEC55 // FORMAT_RIE6       ROTATE THEN INSERT SELECTED BITS
	OP_RISBGN  uint32 = 0xEC59 // FORMAT_RIE6       ROTATE THEN INSERT SELECTED BITS
	OP_RISBHG  uint32 = 0xEC5D // FORMAT_RIE6       ROTATE THEN INSERT SELECTED BITS HIGH
	OP_RISBLG  uint32 = 0xEC51 // FORMAT_RIE6       ROTATE THEN INSERT SELECTED BITS LOW
	OP_RLL     uint32 = 0xEB1D // FORMAT_RSY1       ROTATE LEFT SINGLE LOGICAL (32)
	OP_RLLG    uint32 = 0xEB1C // FORMAT_RSY1       ROTATE LEFT SINGLE LOGICAL (64)
	OP_RNSBG   uint32 = 0xEC54 // FORMAT_RIE6       ROTATE THEN AND SELECTED BITS
	OP_ROSBG   uint32 = 0xEC56 // FORMAT_RIE6       ROTATE THEN OR SELECTED BITS
	OP_RP      uint32 = 0xB277 // FORMAT_S          RESUME PROGRAM
	OP_RRBE    uint32 = 0xB22A // FORMAT_RRE        RESET REFERENCE BIT EXTENDED
	OP_RRBM    uint32 = 0xB9AE // FORMAT_RRE        RESET REFERENCE BITS MULTIPLE
	OP_RRDTR   uint32 = 0xB3F7 // FORMAT_RRF2       REROUND (long DFP)
	OP_RRXTR   uint32 = 0xB3FF // FORMAT_RRF2       REROUND (extended DFP)
	OP_RSCH    uint32 = 0xB238 // FORMAT_S          RESUME SUBCHANNEL
	OP_RXSBG   uint32 = 0xEC57 // FORMAT_RIE6       ROTATE THEN EXCLUSIVE OR SELECTED BITS
	OP_S       uint32 = 0x5B00 // FORMAT_RX1        SUBTRACT (32)
	OP_SAC     uint32 = 0xB219 // FORMAT_S          SET ADDRESS SPACE CONTROL
	OP_SACF    uint32 = 0xB279 // FORMAT_S          SET ADDRESS SPACE CONTROL FAST
	OP_SAL     uint32 = 0xB237 // FORMAT_S          SET ADDRESS LIMIT
	OP_SAM24   uint32 = 0x010C // FORMAT_E          SET ADDRESSING MODE (24)
	OP_SAM31   uint32 = 0x010D // FORMAT_E          SET ADDRESSING MODE (31)
	OP_SAM64   uint32 = 0x010E // FORMAT_E          SET ADDRESSING MODE (64)
	OP_SAR     uint32 = 0xB24E // FORMAT_RRE        SET ACCESS
	OP_SCHM    uint32 = 0xB23C // FORMAT_S          SET CHANNEL MONITOR
	OP_SCK     uint32 = 0xB204 // FORMAT_S          SET CLOCK
	OP_SCKC    uint32 = 0xB206 // FORMAT_S          SET CLOCK COMPARATOR
	OP_SCKPF   uint32 = 0x0107 // FORMAT_E          SET CLOCK PROGRAMMABLE FIELD
	OP_SD      uint32 = 0x6B00 // FORMAT_RX1        SUBTRACT NORMALIZED (long HFP)
	OP_SDB     uint32 = 0xED1B // FORMAT_RXE        SUBTRACT (long BFP)
	OP_SDBR    uint32 = 0xB31B // FORMAT_RRE        SUBTRACT (long BFP)
	OP_SDR     uint32 = 0x2B00 // FORMAT_RR         SUBTRACT NORMALIZED (long HFP)
	OP_SDTR    uint32 = 0xB3D3 // FORMAT_RRF1       SUBTRACT (long DFP)
	OP_SDTRA   uint32 = 0xB3D3 // FORMAT_RRF1       SUBTRACT (long DFP)
	OP_SE      uint32 = 0x7B00 // FORMAT_RX1        SUBTRACT NORMALIZED (short HFP)
	OP_SEB     uint32 = 0xED0B // FORMAT_RXE        SUBTRACT (short BFP)
	OP_SEBR    uint32 = 0xB30B // FORMAT_RRE        SUBTRACT (short BFP)
	OP_SER     uint32 = 0x3B00 // FORMAT_RR         SUBTRACT NORMALIZED (short HFP)
	OP_SFASR   uint32 = 0xB385 // FORMAT_RRE        SET FPC AND SIGNAL
	OP_SFPC    uint32 = 0xB384 // FORMAT_RRE        SET FPC
	OP_SG      uint32 = 0xE309 // FORMAT_RXY1       SUBTRACT (64)
	OP_SGF     uint32 = 0xE319 // FORMAT_RXY1       SUBTRACT (64<-32)
	OP_SGFR    uint32 = 0xB919 // FORMAT_RRE        SUBTRACT (64<-32)
	OP_SGR     uint32 = 0xB909 // FORMAT_RRE        SUBTRACT (64)
	OP_SGRK    uint32 = 0xB9E9 // FORMAT_RRF1       SUBTRACT (64)
	OP_SH      uint32 = 0x4B00 // FORMAT_RX1        SUBTRACT HALFWORD
	OP_SHHHR   uint32 = 0xB9C9 // FORMAT_RRF1       SUBTRACT HIGH (32)
	OP_SHHLR   uint32 = 0xB9D9 // FORMAT_RRF1       SUBTRACT HIGH (32)
	OP_SHY     uint32 = 0xE37B // FORMAT_RXY1       SUBTRACT HALFWORD
	OP_SIGP    uint32 = 0xAE00 // FORMAT_RS1        SIGNAL PROCESSOR
	OP_SL      uint32 = 0x5F00 // FORMAT_RX1        SUBTRACT LOGICAL (32)
	OP_SLA     uint32 = 0x8B00 // FORMAT_RS1        SHIFT LEFT SINGLE (32)
	OP_SLAG    uint32 = 0xEB0B // FORMAT_RSY1       SHIFT LEFT SINGLE (64)
	OP_SLAK    uint32 = 0xEBDD // FORMAT_RSY1       SHIFT LEFT SINGLE (32)
	OP_SLB     uint32 = 0xE399 // FORMAT_RXY1       SUBTRACT LOGICAL WITH BORROW (32)
	OP_SLBG    uint32 = 0xE389 // FORMAT_RXY1       SUBTRACT LOGICAL WITH BORROW (64)
	OP_SLBGR   uint32 = 0xB989 // FORMAT_RRE        SUBTRACT LOGICAL WITH BORROW (64)
	OP_SLBR    uint32 = 0xB999 // FORMAT_RRE        SUBTRACT LOGICAL WITH BORROW (32)
	OP_SLDA    uint32 = 0x8F00 // FORMAT_RS1        SHIFT LEFT DOUBLE
	OP_SLDL    uint32 = 0x8D00 // FORMAT_RS1        SHIFT LEFT DOUBLE LOGICAL
	OP_SLDT    uint32 = 0xED40 // FORMAT_RXF        SHIFT SIGNIFICAND LEFT (long DFP)
	OP_SLFI    uint32 = 0xC205 // FORMAT_RIL1       SUBTRACT LOGICAL IMMEDIATE (32)
	OP_SLG     uint32 = 0xE30B // FORMAT_RXY1       SUBTRACT LOGICAL (64)
	OP_SLGF    uint32 = 0xE31B // FORMAT_RXY1       SUBTRACT LOGICAL (64<-32)
	OP_SLGFI   uint32 = 0xC204 // FORMAT_RIL1       SUBTRACT LOGICAL IMMEDIATE (64<-32)
	OP_SLGFR   uint32 = 0xB91B // FORMAT_RRE        SUBTRACT LOGICAL (64<-32)
	OP_SLGR    uint32 = 0xB90B // FORMAT_RRE        SUBTRACT LOGICAL (64)
	OP_SLGRK   uint32 = 0xB9EB // FORMAT_RRF1       SUBTRACT LOGICAL (64)
	OP_SLHHHR  uint32 = 0xB9CB // FORMAT_RRF1       SUBTRACT LOGICAL HIGH (32)
	OP_SLHHLR  uint32 = 0xB9DB // FORMAT_RRF1       SUBTRACT LOGICAL HIGH (32)
	OP_SLL     uint32 = 0x8900 // FORMAT_RS1        SHIFT LEFT SINGLE LOGICAL (32)
	OP_SLLG    uint32 = 0xEB0D // FORMAT_RSY1       SHIFT LEFT SINGLE LOGICAL (64)
	OP_SLLK    uint32 = 0xEBDF // FORMAT_RSY1       SHIFT LEFT SINGLE LOGICAL (32)
	OP_SLR     uint32 = 0x1F00 // FORMAT_RR         SUBTRACT LOGICAL (32)
	OP_SLRK    uint32 = 0xB9FB // FORMAT_RRF1       SUBTRACT LOGICAL (32)
	OP_SLXT    uint32 = 0xED48 // FORMAT_RXF        SHIFT SIGNIFICAND LEFT (extended DFP)
	OP_SLY     uint32 = 0xE35F // FORMAT_RXY1       SUBTRACT LOGICAL (32)
	OP_SP      uint32 = 0xFB00 // FORMAT_SS2        SUBTRACT DECIMAL
	OP_SPKA    uint32 = 0xB20A // FORMAT_S          SET PSW KEY FROM ADDRESS
	OP_SPM     uint32 = 0x0400 // FORMAT_RR         SET PROGRAM MASK
	OP_SPT     uint32 = 0xB208 // FORMAT_S          SET CPU TIMER
	OP_SPX     uint32 = 0xB210 // FORMAT_S          SET PREFIX
	OP_SQD     uint32 = 0xED35 // FORMAT_RXE        SQUARE ROOT (long HFP)
	OP_SQDB    uint32 = 0xED15 // FORMAT_RXE        SQUARE ROOT (long BFP)
	OP_SQDBR   uint32 = 0xB315 // FORMAT_RRE        SQUARE ROOT (long BFP)
	OP_SQDR    uint32 = 0xB244 // FORMAT_RRE        SQUARE ROOT (long HFP)
	OP_SQE     uint32 = 0xED34 // FORMAT_RXE        SQUARE ROOT (short HFP)
	OP_SQEB    uint32 = 0xED14 // FORMAT_RXE        SQUARE ROOT (short BFP)
	OP_SQEBR   uint32 = 0xB314 // FORMAT_RRE        SQUARE ROOT (short BFP)
	OP_SQER    uint32 = 0xB245 // FORMAT_RRE        SQUARE ROOT (short HFP)
	OP_SQXBR   uint32 = 0xB316 // FORMAT_RRE        SQUARE ROOT (extended BFP)
	OP_SQXR    uint32 = 0xB336 // FORMAT_RRE        SQUARE ROOT (extended HFP)
	OP_SR      uint32 = 0x1B00 // FORMAT_RR         SUBTRACT (32)
	OP_SRA     uint32 = 0x8A00 // FORMAT_RS1        SHIFT RIGHT SINGLE (32)
	OP_SRAG    uint32 = 0xEB0A // FORMAT_RSY1       SHIFT RIGHT SINGLE (64)
	OP_SRAK    uint32 = 0xEBDC // FORMAT_RSY1       SHIFT RIGHT SINGLE (32)
	OP_SRDA    uint32 = 0x8E00 // FORMAT_RS1        SHIFT RIGHT DOUBLE
	OP_SRDL    uint32 = 0x8C00 // FORMAT_RS1        SHIFT RIGHT DOUBLE LOGICAL
	OP_SRDT    uint32 = 0xED41 // FORMAT_RXF        SHIFT SIGNIFICAND RIGHT (long DFP)
	OP_SRK     uint32 = 0xB9F9 // FORMAT_RRF1       SUBTRACT (32)
	OP_SRL     uint32 = 0x8800 // FORMAT_RS1        SHIFT RIGHT SINGLE LOGICAL (32)
	OP_SRLG    uint32 = 0xEB0C // FORMAT_RSY1       SHIFT RIGHT SINGLE LOGICAL (64)
	OP_SRLK    uint32 = 0xEBDE // FORMAT_RSY1       SHIFT RIGHT SINGLE LOGICAL (32)
	OP_SRNM    uint32 = 0xB299 // FORMAT_S          SET BFP ROUNDING MODE (2 bit)
	OP_SRNMB   uint32 = 0xB2B8 // FORMAT_S          SET BFP ROUNDING MODE (3 bit)
	OP_SRNMT   uint32 = 0xB2B9 // FORMAT_S          SET DFP ROUNDING MODE
	OP_SRP     uint32 = 0xF000 // FORMAT_SS3        SHIFT AND ROUND DECIMAL
	OP_SRST    uint32 = 0xB25E // FORMAT_RRE        SEARCH STRING
	OP_SRSTU   uint32 = 0xB9BE // FORMAT_RRE        SEARCH STRING UNICODE
	OP_SRXT    uint32 = 0xED49 // FORMAT_RXF        SHIFT SIGNIFICAND RIGHT (extended DFP)
	OP_SSAIR   uint32 = 0xB99F // FORMAT_RRE        SET SECONDARY ASN WITH INSTANCE
	OP_SSAR    uint32 = 0xB225 // FORMAT_RRE        SET SECONDARY ASN
	OP_SSCH    uint32 = 0xB233 // FORMAT_S          START SUBCHANNEL
	OP_SSKE    uint32 = 0xB22B // FORMAT_RRF3       SET STORAGE KEY EXTENDED
	OP_SSM     uint32 = 0x8000 // FORMAT_S          SET SYSTEM MASK
	OP_ST      uint32 = 0x5000 // FORMAT_RX1        STORE (32)
	OP_STAM    uint32 = 0x9B00 // FORMAT_RS1        STORE ACCESS MULTIPLE
	OP_STAMY   uint32 = 0xEB9B // FORMAT_RSY1       STORE ACCESS MULTIPLE
	OP_STAP    uint32 = 0xB212 // FORMAT_S          STORE CPU ADDRESS
	OP_STC     uint32 = 0x4200 // FORMAT_RX1        STORE CHARACTER
	OP_STCH    uint32 = 0xE3C3 // FORMAT_RXY1       STORE CHARACTER HIGH (8)
	OP_STCK    uint32 = 0xB205 // FORMAT_S          STORE CLOCK
	OP_STCKC   uint32 = 0xB207 // FORMAT_S          STORE CLOCK COMPARATOR
	OP_STCKE   uint32 = 0xB278 // FORMAT_S          STORE CLOCK EXTENDED
	OP_STCKF   uint32 = 0xB27C // FORMAT_S          STORE CLOCK FAST
	OP_STCM    uint32 = 0xBE00 // FORMAT_RS2        STORE CHARACTERS UNDER MASK (low)
	OP_STCMH   uint32 = 0xEB2C // FORMAT_RSY2       STORE CHARACTERS UNDER MASK (high)
	OP_STCMY   uint32 = 0xEB2D // FORMAT_RSY2       STORE CHARACTERS UNDER MASK (low)
	OP_STCPS   uint32 = 0xB23A // FORMAT_S          STORE CHANNEL PATH STATUS
	OP_STCRW   uint32 = 0xB239 // FORMAT_S          STORE CHANNEL REPORT WORD
	OP_STCTG   uint32 = 0xEB25 // FORMAT_RSY1       STORE CONTROL (64)
	OP_STCTL   uint32 = 0xB600 // FORMAT_RS1        STORE CONTROL (32)
	OP_STCY    uint32 = 0xE372 // FORMAT_RXY1       STORE CHARACTER
	OP_STD     uint32 = 0x6000 // FORMAT_RX1        STORE (long)
	OP_STDY    uint32 = 0xED67 // FORMAT_RXY1       STORE (long)
	OP_STE     uint32 = 0x7000 // FORMAT_RX1        STORE (short)
	OP_STEY    uint32 = 0xED66 // FORMAT_RXY1       STORE (short)
	OP_STFH    uint32 = 0xE3CB // FORMAT_RXY1       STORE HIGH (32)
	OP_STFL    uint32 = 0xB2B1 // FORMAT_S          STORE FACILITY LIST
	OP_STFLE   uint32 = 0xB2B0 // FORMAT_S          STORE FACILITY LIST EXTENDED
	OP_STFPC   uint32 = 0xB29C // FORMAT_S          STORE FPC
	OP_STG     uint32 = 0xE324 // FORMAT_RXY1       STORE (64)
	OP_STGRL   uint32 = 0xC40B // FORMAT_RIL2       STORE RELATIVE LONG (64)
	OP_STH     uint32 = 0x4000 // FORMAT_RX1        STORE HALFWORD
	OP_STHH    uint32 = 0xE3C7 // FORMAT_RXY1       STORE HALFWORD HIGH (16)
	OP_STHRL   uint32 = 0xC407 // FORMAT_RIL2       STORE HALFWORD RELATIVE LONG
	OP_STHY    uint32 = 0xE370 // FORMAT_RXY1       STORE HALFWORD
	OP_STIDP   uint32 = 0xB202 // FORMAT_S          STORE CPU ID
	OP_STM     uint32 = 0x9000 // FORMAT_RS1        STORE MULTIPLE (32)
	OP_STMG    uint32 = 0xEB24 // FORMAT_RSY1       STORE MULTIPLE (64)
	OP_STMH    uint32 = 0xEB26 // FORMAT_RSY1       STORE MULTIPLE HIGH
	OP_STMY    uint32 = 0xEB90 // FORMAT_RSY1       STORE MULTIPLE (32)
	OP_STNSM   uint32 = 0xAC00 // FORMAT_SI         STORE THEN AND SYSTEM MASK
	OP_STOC    uint32 = 0xEBF3 // FORMAT_RSY2       STORE ON CONDITION (32)
	OP_STOCG   uint32 = 0xEBE3 // FORMAT_RSY2       STORE ON CONDITION (64)
	OP_STOSM   uint32 = 0xAD00 // FORMAT_SI         STORE THEN OR SYSTEM MASK
	OP_STPQ    uint32 = 0xE38E // FORMAT_RXY1       STORE PAIR TO QUADWORD
	OP_STPT    uint32 = 0xB209 // FORMAT_S          STORE CPU TIMER
	OP_STPX    uint32 = 0xB211 // FORMAT_S          STORE PREFIX
	OP_STRAG   uint32 = 0xE502 // FORMAT_SSE        STORE REAL ADDRESS
	OP_STRL    uint32 = 0xC40F // FORMAT_RIL2       STORE RELATIVE LONG (32)
	OP_STRV    uint32 = 0xE33E // FORMAT_RXY1       STORE REVERSED (32)
	OP_STRVG   uint32 = 0xE32F // FORMAT_RXY1       STORE REVERSED (64)
	OP_STRVH   uint32 = 0xE33F // FORMAT_RXY1       STORE REVERSED (16)
	OP_STSCH   uint32 = 0xB234 // FORMAT_S          STORE SUBCHANNEL
	OP_STSI    uint32 = 0xB27D // FORMAT_S          STORE SYSTEM INFORMATION
	OP_STURA   uint32 = 0xB246 // FORMAT_RRE        STORE USING REAL ADDRESS (32)
	OP_STURG   uint32 = 0xB925 // FORMAT_RRE        STORE USING REAL ADDRESS (64)
	OP_STY     uint32 = 0xE350 // FORMAT_RXY1       STORE (32)
	OP_SU      uint32 = 0x7F00 // FORMAT_RX1        SUBTRACT UNNORMALIZED (short HFP)
	OP_SUR     uint32 = 0x3F00 // FORMAT_RR         SUBTRACT UNNORMALIZED (short HFP)
	OP_SVC     uint32 = 0x0A00 // FORMAT_I          SUPERVISOR CALL
	OP_SW      uint32 = 0x6F00 // FORMAT_RX1        SUBTRACT UNNORMALIZED (long HFP)
	OP_SWR     uint32 = 0x2F00 // FORMAT_RR         SUBTRACT UNNORMALIZED (long HFP)
	OP_SXBR    uint32 = 0xB34B // FORMAT_RRE        SUBTRACT (extended BFP)
	OP_SXR     uint32 = 0x3700 // FORMAT_RR         SUBTRACT NORMALIZED (extended HFP)
	OP_SXTR    uint32 = 0xB3DB // FORMAT_RRF1       SUBTRACT (extended DFP)
	OP_SXTRA   uint32 = 0xB3DB // FORMAT_RRF1       SUBTRACT (extended DFP)
	OP_SY      uint32 = 0xE35B // FORMAT_RXY1       SUBTRACT (32)
	OP_TABORT  uint32 = 0xB2FC // FORMAT_S          TRANSACTION ABORT
	OP_TAM     uint32 = 0x010B // FORMAT_E          TEST ADDRESSING MODE
	OP_TAR     uint32 = 0xB24C // FORMAT_RRE        TEST ACCESS
	OP_TB      uint32 = 0xB22C // FORMAT_RRE        TEST BLOCK
	OP_TBDR    uint32 = 0xB351 // FORMAT_RRF5       CONVERT HFP TO BFP (long)
	OP_TBEDR   uint32 = 0xB350 // FORMAT_RRF5       CONVERT HFP TO BFP (long to short)
	OP_TBEGIN  uint32 = 0xE560 // FORMAT_SIL        TRANSACTION BEGIN
	OP_TBEGINC uint32 = 0xE561 // FORMAT_SIL        TRANSACTION BEGIN
	OP_TCDB    uint32 = 0xED11 // FORMAT_RXE        TEST DATA CLASS (long BFP)
	OP_TCEB    uint32 = 0xED10 // FORMAT_RXE        TEST DATA CLASS (short BFP)
	OP_TCXB    uint32 = 0xED12 // FORMAT_RXE        TEST DATA CLASS (extended BFP)
	OP_TDCDT   uint32 = 0xED54 // FORMAT_RXE        TEST DATA CLASS (long DFP)
	OP_TDCET   uint32 = 0xED50 // FORMAT_RXE        TEST DATA CLASS (short DFP)
	OP_TDCXT   uint32 = 0xED58 // FORMAT_RXE        TEST DATA CLASS (extended DFP)
	OP_TDGDT   uint32 = 0xED55 // FORMAT_RXE        TEST DATA GROUP (long DFP)
	OP_TDGET   uint32 = 0xED51 // FORMAT_RXE        TEST DATA GROUP (short DFP)
	OP_TDGXT   uint32 = 0xED59 // FORMAT_RXE        TEST DATA GROUP (extended DFP)
	OP_TEND    uint32 = 0xB2F8 // FORMAT_S          TRANSACTION END
	OP_THDER   uint32 = 0xB358 // FORMAT_RRE        CONVERT BFP TO HFP (short to long)
	OP_THDR    uint32 = 0xB359 // FORMAT_RRE        CONVERT BFP TO HFP (long)
	OP_TM      uint32 = 0x9100 // FORMAT_SI         TEST UNDER MASK
	OP_TMH     uint32 = 0xA700 // FORMAT_RI1        TEST UNDER MASK HIGH
	OP_TMHH    uint32 = 0xA702 // FORMAT_RI1        TEST UNDER MASK (high high)
	OP_TMHL    uint32 = 0xA703 // FORMAT_RI1        TEST UNDER MASK (high low)
	OP_TML     uint32 = 0xA701 // FORMAT_RI1        TEST UNDER MASK LOW
	OP_TMLH    uint32 = 0xA700 // FORMAT_RI1        TEST UNDER MASK (low high)
	OP_TMLL    uint32 = 0xA701 // FORMAT_RI1        TEST UNDER MASK (low low)
	OP_TMY     uint32 = 0xEB51 // FORMAT_SIY        TEST UNDER MASK
	OP_TP      uint32 = 0xEBC0 // FORMAT_RSL        TEST DECIMAL
	OP_TPI     uint32 = 0xB236 // FORMAT_S          TEST PENDING INTERRUPTION
	OP_TPROT   uint32 = 0xE501 // FORMAT_SSE        TEST PROTECTION
	OP_TR      uint32 = 0xDC00 // FORMAT_SS1        TRANSLATE
	OP_TRACE   uint32 = 0x9900 // FORMAT_RS1        TRACE (32)
	OP_TRACG   uint32 = 0xEB0F // FORMAT_RSY1       TRACE (64)
	OP_TRAP2   uint32 = 0x01FF // FORMAT_E          TRAP
	OP_TRAP4   uint32 = 0xB2FF // FORMAT_S          TRAP
	OP_TRE     uint32 = 0xB2A5 // FORMAT_RRE        TRANSLATE EXTENDED
	OP_TROO    uint32 = 0xB993 // FORMAT_RRF3       TRANSLATE ONE TO ONE
	OP_TROT    uint32 = 0xB992 // FORMAT_RRF3       TRANSLATE ONE TO TWO
	OP_TRT     uint32 = 0xDD00 // FORMAT_SS1        TRANSLATE AND TEST
	OP_TRTE    uint32 = 0xB9BF // FORMAT_RRF3       TRANSLATE AND TEST EXTENDED
	OP_TRTO    uint32 = 0xB991 // FORMAT_RRF3       TRANSLATE TWO TO ONE
	OP_TRTR    uint32 = 0xD000 // FORMAT_SS1        TRANSLATE AND TEST REVERSE
	OP_TRTRE   uint32 = 0xB9BD // FORMAT_RRF3       TRANSLATE AND TEST REVERSE EXTENDED
	OP_TRTT    uint32 = 0xB990 // FORMAT_RRF3       TRANSLATE TWO TO TWO
	OP_TS      uint32 = 0x9300 // FORMAT_S          TEST AND SET
	OP_TSCH    uint32 = 0xB235 // FORMAT_S          TEST SUBCHANNEL
	OP_UNPK    uint32 = 0xF300 // FORMAT_SS2        UNPACK
	OP_UNPKA   uint32 = 0xEA00 // FORMAT_SS1        UNPACK ASCII
	OP_UNPKU   uint32 = 0xE200 // FORMAT_SS1        UNPACK UNICODE
	OP_UPT     uint32 = 0x0102 // FORMAT_E          UPDATE TREE
	OP_X       uint32 = 0x5700 // FORMAT_RX1        EXCLUSIVE OR (32)
	OP_XC      uint32 = 0xD700 // FORMAT_SS1        EXCLUSIVE OR (character)
	OP_XG      uint32 = 0xE382 // FORMAT_RXY1       EXCLUSIVE OR (64)
	OP_XGR     uint32 = 0xB982 // FORMAT_RRE        EXCLUSIVE OR (64)
	OP_XGRK    uint32 = 0xB9E7 // FORMAT_RRF1       EXCLUSIVE OR (64)
	OP_XI      uint32 = 0x9700 // FORMAT_SI         EXCLUSIVE OR (immediate)
	OP_XIHF    uint32 = 0xC006 // FORMAT_RIL1       EXCLUSIVE OR IMMEDIATE (high)
	OP_XILF    uint32 = 0xC007 // FORMAT_RIL1       EXCLUSIVE OR IMMEDIATE (low)
	OP_XIY     uint32 = 0xEB57 // FORMAT_SIY        EXCLUSIVE OR (immediate)
	OP_XR      uint32 = 0x1700 // FORMAT_RR         EXCLUSIVE OR (32)
	OP_XRK     uint32 = 0xB9F7 // FORMAT_RRF1       EXCLUSIVE OR (32)
	OP_XSCH    uint32 = 0xB276 // FORMAT_S          CANCEL SUBCHANNEL
	OP_XY      uint32 = 0xE357 // FORMAT_RXY1       EXCLUSIVE OR (32)
	OP_ZAP     uint32 = 0xF800 // FORMAT_SS2        ZERO AND ADD
)

func oclass(a *obj.Addr) int {
	return int(a.Class) - 1
}

// Add a relocation for the immediate in a RIL style instruction.
// The addend will be adjusted as required.
func addrilreloc(ctxt *obj.Link, sym *obj.LSym, add int64) *obj.Reloc {
	offset := int64(2) // relocation offset from start of instruction
	rel := obj.Addrel(ctxt.Cursym)
	rel.Off = int32(ctxt.Pc + offset)
	rel.Siz = 4
	rel.Sym = sym
	rel.Add = add + offset + int64(rel.Siz)
	rel.Type = obj.R_PCRELDBL
	return rel
}

// Add a CALL relocation for the immediate in a RIL style instruction.
// The addend will be adjusted as required.
func addcallreloc(ctxt *obj.Link, sym *obj.LSym, add int64) *obj.Reloc {
	offset := int64(2) // relocation offset from start of instruction
	rel := obj.Addrel(ctxt.Cursym)
	rel.Off = int32(ctxt.Pc + offset)
	rel.Siz = 4
	rel.Sym = sym
	rel.Add = add + offset + int64(rel.Siz)
	rel.Type = obj.R_CALL
	return rel
}

/*
 * 32-bit masks
 */
func getmask(m []byte, v uint32) bool {
	m[1] = 0
	m[0] = m[1]
	if v != ^uint32(0) && v&(1<<31) != 0 && v&1 != 0 { /* MB > ME */
		if getmask(m, ^v) {
			i := int(m[0])
			m[0] = m[1] + 1
			m[1] = byte(i - 1)
			return true
		}

		return false
	}

	for i := 0; i < 32; i++ {
		if v&(1<<uint(31-i)) != 0 {
			m[0] = byte(i)
			for {
				m[1] = byte(i)
				i++
				if i >= 32 || v&(1<<uint(31-i)) == 0 {
					break
				}
			}

			for ; i < 32; i++ {
				if v&(1<<uint(31-i)) != 0 {
					return false
				}
			}
			return true
		}
	}

	return false
}

func maskgen(ctxt *obj.Link, p *obj.Prog, m []byte, v uint32) {
	if !getmask(m, v) {
		ctxt.Diag("cannot generate mask #%x\n%v", v, p)
	}
}

/*
 * 64-bit masks (rldic etc)
 */
func getmask64(m []byte, v uint64) bool {
	m[1] = 0
	m[0] = m[1]
	for i := 0; i < 64; i++ {
		if v&(uint64(1)<<uint(63-i)) != 0 {
			m[0] = byte(i)
			for {
				m[1] = byte(i)
				i++
				if i >= 64 || v&(uint64(1)<<uint(63-i)) == 0 {
					break
				}
			}

			for ; i < 64; i++ {
				if v&(uint64(1)<<uint(63-i)) != 0 {
					return false
				}
			}
			return true
		}
	}

	return false
}

func maskgen64(ctxt *obj.Link, p *obj.Prog, m []byte, v uint64) {
	if !getmask64(m, v) {
		ctxt.Diag("cannot generate mask #%x\n%v", v, p)
	}
}

func branchMask(ctxt *obj.Link, p *obj.Prog) uint32 {
	switch p.As {
	case ABEQ, ACMPBEQ, ACMPUBEQ:
		return 0x8
	case ABGE, ACMPBGE, ACMPUBGE:
		return 0xA
	case ABGT, ACMPBGT, ACMPUBGT:
		return 0x2
	case ABLE, ACMPBLE, ACMPUBLE:
		return 0xC
	case ABLT, ACMPBLT, ACMPUBLT:
		return 0x4
	case ABNE, ACMPBNE, ACMPUBNE:
		return 0x7
	case ABVC:
		return 0x0 //needs extra instruction
	case ABVS:
		return 0x1
	}
	ctxt.Diag("unknown conditional branch %v", p.As)
	return 0xF
}

func asmout(ctxt *obj.Link, asm *[]byte) {
	p := ctxt.Curp
	o := oplook(ctxt, p)
	ctxt.Printp = p

	switch o.type_ {
	default:
		ctxt.Diag("unknown type %d", o.type_)

	case 0: /* pseudo ops */
		break

	case 1: /* mov r1,r2 ==> OR Rs,Rs,Ra */ // lgr r1,r2
		if p.From.Type == obj.TYPE_CONST { //happens when a zero constant is used
			v := regoff(ctxt, &p.From)
			RI(OP_LGHI, uint32(p.To.Reg), uint32(v), asm)
		} else {
			RRE(OP_LGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		}

	case 2: /* int/cr/fp op Rb,[Ra],Rd */
		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}

		var opcode uint32

		switch p.As {
		default:
			ctxt.Diag("invalid opcode")
		case AADD:
			opcode = OP_AGRK
		case AADDC:
			opcode = OP_ALGRK
		case AADDE:
			opcode = OP_ALCGR
		case AMULLW:
			opcode = OP_MSGFR
		case AMULLD:
			opcode = OP_MSGR
		case AMULHDU:
			opcode = OP_MLGR
		case AMULHD:
			ctxt.Diag("MULHD not implemented")
		case ADIVW:
			opcode = OP_DSGFR
		case ADIVWU:
			opcode = OP_DLR
		case ADIVD:
			opcode = OP_DSGR
		case ADIVDU:
			opcode = OP_DLGR
		case AFADD:
			opcode = OP_ADBR
		case AFADDS:
			opcode = OP_AEBR
		case AFSUB:
			opcode = OP_SDBR
		case AFSUBS:
			opcode = OP_SEBR
		case AFDIV:
			opcode = OP_DDBR
		case AFDIVS:
			opcode = OP_DEBR
		}

		switch p.As {
		default:

		case AADD, AADDC:
			RRF(opcode, uint32(p.From.Reg), 0, uint32(p.To.Reg), uint32(r), asm)

		case AADDE, AMULLW, AMULLD:
			if r == int(p.To.Reg) {
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else if p.From.Reg == p.To.Reg {
				RRE(opcode, uint32(p.To.Reg), uint32(r), asm)
			} else {
				RRE(OP_LGR, uint32(p.To.Reg), uint32(r), asm)
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

		case ADIVW, ADIVWU, ADIVD, ADIVDU:
			if p.As == ADIVWU || p.As == ADIVDU {
				RRE(OP_LGR, uint32(REGTMP), uint32(REGZERO), asm)
			}
			RRE(OP_LGR, uint32(REGTMP2), uint32(r), asm)
			RRE(opcode, uint32(REGTMP), uint32(p.From.Reg), asm)
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGTMP2), asm)

		case AMULHDU:
			RRE(OP_LGR, uint32(REGTMP2), uint32(r), asm)
			RRE(opcode, uint32(REGTMP), uint32(p.From.Reg), asm)
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGTMP), asm)

		case AFADD, AFADDS:
			if r == int(p.To.Reg) {
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else if p.From.Reg == p.To.Reg {
				RRE(opcode, uint32(p.To.Reg), uint32(r), asm)
			} else {
				RR(OP_LDR, uint32(p.To.Reg), uint32(r), asm)
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

		case AFSUB, AFSUBS, AFDIV, AFDIVS:
			if r == int(p.To.Reg) {
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else if p.From.Reg == p.To.Reg {
				RRE(OP_LGDR, uint32(REGTMP), uint32(r), asm)
				RRE(opcode, uint32(r), uint32(p.From.Reg), asm)
				RR(OP_LDR, uint32(p.To.Reg), uint32(r), asm)
				RRE(OP_LDGR, uint32(r), uint32(REGTMP), asm)
			} else {
				RR(OP_LDR, uint32(p.To.Reg), uint32(r), asm)
				RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

		}

	case 3: // mov $soreg/addcon/ucon,r => LGFI Rd, $i or LAY Rd, $i(Rs)
		d := vregoff(ctxt, &p.From)

		v := int32(d)
		r := int(p.From.Reg)

		if r == 0 {
			r = int(o.param)
		}

		if r == 0 {
			switch p.As {
			case AMOVWZ, AMOVHZ, AMOVBZ:
				switch p.As {
				case AMOVHZ:
					v = int32(uint16(v))
				case AMOVBZ:
					v = int32(uint8(v))
				}
				RIL(a, OP_LLILF, uint32(p.To.Reg), uint32(v), asm)
			default:
				RIL(a, OP_LGFI, uint32(p.To.Reg), uint32(v), asm)
			}
		} else if v >= 0 && v < DISP12 {
			RX(OP_LA, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		} else {
			RXY(a, OP_LAY, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		}

	case 4: /* add/mul $scon,[r1],r2 */ //Might have to worry about condition codes
		v := regoff(ctxt, &p.From)

		r := int(p.To.Reg)
		r2 := int(p.Reg)

		if p.Reg == 0 && p.To.Reg == 0 {
			ctxt.Diag("literal operation on R0\n%v", p)
		}
		if p.As == AADD {
			if r2 == 0 {
				RIL(a, OP_AGFI, uint32(r), uint32(v), asm)
			} else {
				RIE(d, OP_AGHIK, uint32(r), uint32(r2), uint32(v), 0, 0, 0, 0, asm)
			}
		} else if p.As == AMULLW {
			if r2 == 0 {
				RIL(a, OP_MSGFI, uint32(r), uint32(v), asm)
			} else {
				RIL(a, OP_MSGFI, uint32(r2), uint32(v), asm)
				RRE(OP_LGFR, uint32(r), uint32(r2), asm)
			}
		}

	case 5: /* syscall */ // This might be right, assuming SVC is the same as Power's SC
		I(OP_SVC, 0, asm)

	case 6: /* logical op Rb,[Rs,]Ra; no literal */
		if p.To.Reg == 0 {
			ctxt.Diag("literal operation on R0\n%v", p)
		}

		switch p.As {
		case AAND, AOR, AXOR:
			var opcode1, opcode2 uint32
			switch p.As {
			default:
			case AAND:
				opcode1 = OP_NGR
				opcode2 = OP_NGRK
			case AOR:
				opcode1 = OP_OGR
				opcode2 = OP_OGRK
			case AXOR:
				opcode1 = OP_XGR
				opcode2 = OP_XGRK
			}

			r := int(p.Reg)
			if r == 0 {
				RRE(opcode1, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else {
				RRF(opcode2, uint32(r), 0, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

		case AANDN, AORN:
			var opcode1, opcode2 uint32
			switch p.As {
			default:
			case AANDN:
				opcode1 = OP_NGR
				opcode2 = OP_NGRK
			case AORN:
				opcode1 = OP_OGR
				opcode2 = OP_OGRK
			}

			r := int(p.Reg)
			if r == 0 {
				RRE(OP_LCGR, uint32(p.To.Reg), uint32(p.To.Reg), asm)
				RRE(opcode1, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else {
				RRE(OP_LCGR, uint32(REGTMP), uint32(r), asm)
				RRF(opcode2, uint32(REGTMP), 0, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

		case ANAND, ANOR:
			var opcode1, opcode2 uint32
			switch p.As {
			default:
			case ANAND:
				opcode1 = OP_NGR
				opcode2 = OP_NGRK
			case ANOR:
				opcode1 = OP_OGR
				opcode2 = OP_OGRK
			}

			r := int(p.Reg)
			if r == 0 {
				RRE(opcode1, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else {
				RRF(opcode2, uint32(r), 0, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}

			RRE(OP_LCGR, uint32(p.To.Reg), uint32(p.To.Reg), asm)

		case ASLW, ASLD, ASRAW, ASRAD, ASRW, ASRD:
			var opcode uint32
			switch p.As {
			default:
			case ASLW:
				opcode = OP_SLLK
			case ASLD:
				opcode = OP_SLLG
			case ASRAW:
				opcode = OP_SRAK
			case ASRAD:
				opcode = OP_SRAG
			case ASRW:
				opcode = OP_SRLK
			case ASRD:
				opcode = OP_SRLG
			}

			r := int(p.Reg)
			if r == 0 {
				r = int(p.To.Reg)
			}
			RSY(opcode, uint32(p.To.Reg), uint32(r), uint32(p.From.Reg), 0, asm)

		default:

		}

	case 7: /* mov r, soreg ==> stg o(r) */
		x := int(p.To.Reg)
		if x == 0 {
			x = int(o.param)
		}
		var b int
		if p.To.Type == obj.TYPE_MEM && p.To.Index != 0 {
			b = int(p.To.Index)
		} else {
			b = x
			x = 0
		}
		v := regoff(ctxt, &p.To)

		RXY(0, zopstore(ctxt, int(p.As)), uint32(p.From.Reg), uint32(x), uint32(b), uint32(v), asm)

	/* AMOVD, AMOVW, AMOVWZ, AMOVBZ */
	case 8, 9: /* mov soreg, r ==> lbz/lhz/lwz o(r) */
		x := int(p.From.Reg)
		if x == 0 {
			x = int(o.param)
		}
		var b int
		if p.From.Type == obj.TYPE_MEM && p.From.Index != 0 {
			b = int(p.From.Index)
		} else {
			b = x
			x = 0
		}
		v := regoff(ctxt, &p.From)

		RXY(0, zopload(ctxt, int(p.As)), uint32(p.To.Reg), uint32(x), uint32(b), uint32(v), asm)

	case 10: /* sub Ra,[Rb],Rd => subf Rd,Ra,Rb */
		r := int(p.Reg)

		switch p.As {
		default:
		case ASUB:
			if r == 0 {
				RRE(OP_SGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else {
				RRF(OP_SGRK, uint32(p.From.Reg), 0, uint32(p.To.Reg), uint32(r), asm)
			}
		case ASUBC:
			if r == 0 {
				RRE(OP_SLGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else {
				RRF(OP_SLGRK, uint32(p.From.Reg), 0, uint32(p.To.Reg), uint32(r), asm)
			}

		case ASUBE:
			if r == 0 {
				r = int(p.To.Reg)
			}
			if r == int(p.To.Reg) {
				RRE(OP_SLBGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			} else if p.From.Reg == p.To.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.From.Reg), asm)
				RRE(OP_LGR, uint32(p.To.Reg), uint32(r), asm)
				RRE(OP_SLBGR, uint32(p.To.Reg), uint32(REGTMP), asm)
			} else {
				RRE(OP_LGR, uint32(p.To.Reg), uint32(r), asm)
				RRE(OP_SLBGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
			}
		}

	case 11: /* br/bl lbra */
		v := int32(0)

		if p.Pcond != nil {
			v = int32((p.Pcond.Pc - p.Pc) >> 1)
		}

		if p.As == ABR && p.To.Sym == nil && int32(int16(v)) == v {
			RI(OP_BRC, 0xF, uint32(v), asm)
		} else {
			if p.As == ABL {
				RIL(b, OP_BRASL, uint32(REG_LR), uint32(v), asm)
			} else {
				RIL(c, OP_BRCL, 0xF, uint32(v), asm)
			}
			if p.To.Sym != nil {
				addcallreloc(ctxt, p.To.Sym, p.To.Offset)
			}
		}

	case 12: /* movb r,r (lgbr); movw r,r (lgfr) */
		if p.To.Reg == REGZERO && p.From.Type == obj.TYPE_CONST {
			v := regoff(ctxt, &p.From)
			if r0iszero != 0 /*TypeKind(100016)*/ && v != 0 {
				ctxt.Diag("literal operation on R0\n%v", p)
			}
			RIL(a, OP_LGFI, uint32(REGZERO), uint32(v), asm)
		} else if p.As == AMOVW {
			RRE(OP_LGFR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else {
			RRE(OP_LGBR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		}

	case 13: /* mov[bhw]z r,r (llgbr, llghr, llgfr) */
		if p.As == AMOVBZ {
			RRE(OP_LLGCR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else if p.As == AMOVH {
			RRE(OP_LGHR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else if p.As == AMOVHZ {
			RRE(OP_LLGHR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else if p.As == AMOVWZ {
			RRE(OP_LLGFR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else {
			ctxt.Diag("internal: bad mov[bhw]z\n%v", p)
		}

	case 14: /* rldc[lr] Rb,Rs,$mask,Ra -- left, right give different masks */
		d := vregoff(ctxt, p.From3)
		var mask [2]uint8
		maskgen64(ctxt, p, mask[:], uint64(d))
		var i3, i4 int
		switch p.As {
		case ARLDCL:
			i3 = int(mask[0]) // MB
			i4 = int(63)
			if mask[1] != 63 {
				ctxt.Diag("invalid mask for rotate: %x (end != bit 63)\n%v", uint64(d), p)
			}

		case ARLDCR:
			i3 = int(0)
			i4 = int(mask[1]) // ME
			if mask[0] != 0 {
				ctxt.Diag("invalid mask for rotate: %x (start != 0)\n%v", uint64(d), p)
			}

		default:
			ctxt.Diag("unexpected op in rldc case\n%v", p)
		}

		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}
		RSY(OP_RLLG, uint32(REGTMP), uint32(r), uint32(p.From.Reg), 0, asm)
		RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm)
		RIE(f, OP_RISBG, uint32(p.To.Reg), uint32(REGTMP), 0, uint32(i3), uint32(i4), 0, 0, asm)

	case 15: /* br/bl (r) */
		r := p.To.Reg
		if p.As == ABCL || p.As == ABL {
			RR(OP_BASR, uint32(REG_LR), uint32(r), asm)
		} else {
			RR(OP_BCR, 0xF, uint32(r), asm)
		}

	case 17, /* bc bo,bi,lbra (same for now) */
		16: /* bc bo,bi,sbra */
		v := int32(0)
		if p.Pcond != nil {
			v = int32((p.Pcond.Pc - p.Pc) >> 1)
		}
		mask := branchMask(ctxt, p)
		if p.To.Sym == nil && int32(int16(v)) == v {
			RI(OP_BRC, uint32(mask), uint32(v), asm)
		} else {
			RIL(c, OP_BRCL, uint32(mask), uint32(v), asm)
		}
		if p.To.Sym != nil {
			addrilreloc(ctxt, p.To.Sym, p.To.Offset)
		}

	case 18: // br/bl r
		switch oclass(&p.To) {
		case C_REG:
			if p.As == ABL {
				RR(OP_BASR, uint32(REG_LR), uint32(p.To.Reg), asm)
			} else {
				RR(OP_BCR, 0xF, uint32(p.To.Reg), asm)
			}
		default:
			ctxt.Diag("bad optab entry (18): %d\n%v", p.To.Class, p)
		}

	case 19: /* mov $lcon,r ==> cau+or */
		d := vregoff(ctxt, &p.From)

		if p.From.Sym == nil {
			RIL(a, OP_LGFI, uint32(p.To.Reg), uint32(d), asm)
		} else if p.From.Sym.Name == "runtime.tlsg" {
			// This is a hack to get the right offset for g from the TLS block.
			// It would be nice to have some syntax to do this properly.
			// This will NOT work when generating a shared object.
			switch p.As {
			default:
				ctxt.Diag("can only place TLS variable offset into a 8-byte register (i.e. need MOVD)")
			case AMOVD:
				// The R_390_TLS_LE32 relocation isn't actually implemented for ELF64,
				// we therefore need to use the 64-bit equivalent, which means using .rodata.
				var sym *obj.LSym
				sym = obj.Linklookup(ctxt, "runtime.tlsg_offset", 0)
				sym.Type = obj.SRODATA
				sym.Size = 8
				obj.Symgrow(ctxt, sym, sym.Size) // needed for relocation to apply
				sym.R = make([]obj.Reloc, 0)
				offrel := obj.Addrel(sym)
				offrel.Off = 0
				offrel.Siz = 8
				offrel.Sym = obj.Linklookup(ctxt, "runtime.tlsg", 0)
				offrel.Add = 0
				offrel.Type = obj.R_TLS_LE
				rel := obj.Addrel(ctxt.Cursym)
				rel.Off = int32(ctxt.Pc + 2)
				rel.Siz = 4
				rel.Sym = sym
				rel.Add = int64(rel.Siz) + 2
				rel.Type = obj.R_PCRELDBL
				RIL(a, OP_LGRL, uint32(p.To.Reg), 0, asm)
			}
		} else {
			RIL(b, OP_LARL, uint32(p.To.Reg), 0, asm)
			if d&1 != 0 {
				RX(OP_LA, uint32(p.To.Reg), uint32(p.To.Reg), 0, 1, asm)
				d -= 1
			}
			addrilreloc(ctxt, p.From.Sym, d)
		}

	case 20: /* add $ucon,,r */
		v := regoff(ctxt, &p.From)

		r := int(p.Reg)
		if p.As == AADD && p.To.Reg == 0 {
			ctxt.Diag("literal operation on R0\n%v", p)
		}
		if r == 0 {
			RIL(a, OP_AGFI, uint32(p.To.Reg), uint32(v), asm)
		} else {
			RIL(a, OP_LGFI, REGTMP, uint32(v), asm)
			RRE(OP_AGR, REGTMP, uint32(r), asm)
			RRE(OP_LGR, uint32(p.To.Reg), REGTMP, asm)
		}

	case 22: /* add $lcon,r1,r2 ==> cau+or+add */ /* could do add/sub more efficiently */

		if p.From.Sym != nil {
			ctxt.Diag("%v is not supported", p)
		}

		d := vregoff(ctxt, &p.From)
		var opcode uint32
		switch p.As {
		default:
		case AADD:
			opcode = OP_AGFI
		case AADDC:
			opcode = OP_ALGFI
		case AMULLW:
			opcode = OP_MSGFI
		}

		r := int(p.Reg)
		if r != 0 {
			RRE(OP_LGR, uint32(p.To.Reg), uint32(r), asm)
		}
		RIL(a, opcode, uint32(p.To.Reg), uint32(d), asm)

	case 23: /* and $lcon,r1,r2 ==> cau+or+and */ /* masks could be done using rlnm etc. */

		d := regoff(ctxt, &p.From)
		var opcode uint32
		r := int(p.Reg)
		if r == 0 {
			switch p.As {
			default:
				ctxt.Diag("%v is not supported", p)
			case AAND:
				opcode = OP_NGR
			case AOR:
				opcode = OP_OGR
			case AXOR:
				opcode = OP_XGR
			}
			RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
			RRE(opcode, uint32(p.To.Reg), REGTMP, asm)
		} else {
			switch p.As {
			default:
				ctxt.Diag("%v is not supported", p)
			case AAND:
				opcode = OP_NGRK
			case AOR:
				opcode = OP_OGRK
			case AXOR:
				opcode = OP_XGRK
			}
			RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
			RRF(opcode, uint32(r), 0, uint32(p.To.Reg), REGTMP, asm)
		}

	case 25:
		/* sld[.] $sh,rS,rA -> rldicr[.] $sh,rS,mask(0,63-sh),rA; srd[.] -> rldicl */
		v := regoff(ctxt, &p.From)
		if v < 0 {
			v = 0
		} else if v > 63 {
			v = 63
		}

		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}

		switch p.As {
		default:
		case ASLD:
			RSY(OP_SLLG, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		case ASRD:
			RSY(OP_SRLG, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		}

	case 26: /* mov $lsext/auto/oreg,,r2 ==> addis+addi */
		v := regoff(ctxt, &p.From)
		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}

		if v >= -DISP20/2 && v < DISP20/2 {
			RXY(0, OP_LAY, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		} else {
			RIL(a, OP_LGFI, REGTMP, uint32(v), asm)
			RX(OP_LA, uint32(p.To.Reg), uint32(r), REGTMP, 0, asm)
		}

	case 27: /* subc ra,$simm,rd => subfic rd,ra,$simm */

		v := regoff(ctxt, p.From3)
		RRE(OP_LCGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		RIE(d, OP_AGHIK, uint32(p.To.Reg), uint32(p.To.Reg), uint32(v), 0, 0, 0, 0, asm)

	case 28: /* subc r1,$lcon,r2 ==> cau+or+subfc */

		v := regoff(ctxt, p.From3)
		RRE(OP_LCGR, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		RIL(a, OP_AGFI, uint32(p.To.Reg), uint32(v), asm)

		if p.From.Sym != nil {
			ctxt.Diag("%v is not supported", p)
		}

	case 29: /* rldic[lr]? $sh,s,$mask,a -- left, right, plain give different masks */
		v := regoff(ctxt, &p.From)
		d := vregoff(ctxt, p.From3)

		var mask [2]uint8
		maskgen64(ctxt, p, mask[:], uint64(d))

		var i3, i4, i5 int
		switch p.As {
		case ARLDC:
			i3 = int(mask[0]) // MB
			i4 = int(63 - v)
			i5 = int(v)
			if int32(mask[1]) != int32(63-v) {
				ctxt.Diag("invalid mask for shift: %x (shift %d)\n%v", uint64(d), v, p)
			}

		case ARLDCL:
			i3 = int(mask[0]) // MB
			i4 = int(63)
			i5 = int(v)
			if mask[1] != 63 {
				ctxt.Diag("invalid mask for shift: %x (shift %d)\n%v", uint64(d), v, p)
			}

		case ARLDCR:
			i3 = int(0)
			i4 = int(mask[1]) // ME
			i5 = int(v)
			if mask[0] != 0 {
				ctxt.Diag("invalid mask for shift: %x (shift %d)\n%v", uint64(d), v, p)
			}

		default:
			ctxt.Diag("unexpected op in rldic case\n%v", p)
		}

		r := int(p.Reg)
		if p.To.Reg == p.Reg {
			RRE(OP_LGR, uint32(REGTMP), uint32(p.Reg), asm)
			r = int(REGTMP)
		}
		RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm)
		RIE(f, OP_RISBG, uint32(p.To.Reg), uint32(r), 0, uint32(i3), uint32(i4), 0, uint32(i5), asm)

	case 30: /* rldimi $sh,s,$mask,a */
		v := regoff(ctxt, &p.From)
		d := vregoff(ctxt, p.From3)

		var mask [2]uint8
		maskgen64(ctxt, p, mask[:], uint64(d))

		var i3, i4, i5 int
		i3 = int(mask[0]) // MB
		i4 = int(63 - v)
		i5 = int(v)
		if int32(mask[1]) != int32(63-v) {
			ctxt.Diag("invalid mask for shift: %x (shift %d)\n%v", uint64(d), v, p)
		}

		RIE(f, OP_RISBG, uint32(p.To.Reg), uint32(p.Reg), 0, uint32(i3), uint32(i4), 0, uint32(i5), asm)

	case 31: /* dword */
		wd := uint64(vregoff(ctxt, &p.From))
		*asm = append(*asm,
			uint8(wd>>56),
			uint8(wd>>48),
			uint8(wd>>40),
			uint8(wd>>32),
			uint8(wd>>24),
			uint8(wd>>16),
			uint8(wd>>8),
			uint8(wd))

	case 32: /* fmul frc,fra,frd */
		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}

		var opcode uint32

		switch p.As {
		default:
			ctxt.Diag("invalid opcode")
		case AFMUL:
			opcode = OP_MDBR
		case AFMULS:
			opcode = OP_MEEBR
		}

		if r == int(p.To.Reg) {
			RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		} else if p.From.Reg == p.To.Reg {
			RRE(opcode, uint32(p.To.Reg), uint32(r), asm)
		} else {
			RR(OP_LDR, uint32(p.To.Reg), uint32(r), asm)
			RRE(opcode, uint32(p.To.Reg), uint32(p.From.Reg), asm)
		}

	case 33: /* fabs [frb,]frd; fmr. frb,frd */
		r := int(p.From.Reg)

		if oclass(&p.From) == C_NONE {
			r = int(p.To.Reg)
		}

		switch p.As {
		case AFMOVD, AFMOVS:
			RR(OP_LDR, uint32(p.To.Reg), uint32(r), asm)
		default:
			var opcode uint32

			switch p.As {
			default:
			case AFABS:
				opcode = OP_LPDBR
			case AFNABS:
				opcode = OP_LNDBR
			case AFNEG:
				opcode = OP_LCDFR
			case AFRSP:
				opcode = OP_LEDBR
			case ALDEBR:
				opcode = OP_LDEBR
			case AFSQRT:
				opcode = OP_SQDBR
			case AFSQRTS:
				opcode = OP_SQEBR
			}

			RRE(opcode, uint32(p.To.Reg), uint32(r), asm)
		}

	case 34: /* FMADDx fra,frb,frc,frd (d=a*b+c); FSELx a<0? (d=b): (d=c) */

		var opcode uint32

		switch p.As {
		default:
			ctxt.Diag("invalid opcode")
		case AFMADD:
			opcode = OP_MADBR
		case AFMADDS:
			opcode = OP_MAEBR
		case AFMSUB:
			opcode = OP_MSDBR
		case AFMSUBS:
			opcode = OP_MSEBR
		case AFNMADD:
			opcode = OP_MADBR
		case AFNMADDS:
			opcode = OP_MAEBR
		case AFNMSUB:
			opcode = OP_MSDBR
		case AFNMSUBS:
			opcode = OP_MSEBR
		}

		RR(OP_LDR, uint32(p.To.Reg), uint32(p.Reg), asm)
		RRD(opcode, uint32(p.To.Reg), uint32(p.From.Reg), uint32(p.From3.Reg), asm)

		if p.As == AFNMADD || p.As == AFNMADDS || p.As == AFNMSUB || p.As == AFNMSUBS {
			RRE(OP_LCDFR, uint32(p.To.Reg), uint32(p.To.Reg), asm)
		}

	case 35: /* mov[M][D|W|H|B][Z] r, lext/lauto/loreg ==>
		   LGFI regtmp, off; STG[F|H|C] r, 0(regaddr, regtmp) */
		v := regoff(ctxt, &p.To)
		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}

		if v >= -DISP20/2 && v < DISP20/2 {
			RXY(0, zopstore(ctxt, int(p.As)), uint32(p.From.Reg), uint32(r), 0,
				uint32(v), asm)
		} else {
			RIL(a, OP_LGFI, REGTMP, uint32(v), asm)
			RXY(0, zopstore(ctxt, int(p.As)), uint32(p.From.Reg), uint32(r), REGTMP,
				0, asm)
		}

	case 36, 37: /* MOV[M][D|W|H|B][Z] lext/lauto/lreg, r ==>
		   LGFI regtmp, off; [L]LG[F|H|C] r, 0(regaddr, regtmp) */
		v := regoff(ctxt, &p.From)

		r := int(p.From.Reg)
		if r == 0 {
			r = int(o.param)
		}

		if v >= -DISP20/2 && v < DISP20/2 {
			RXY(0, zopload(ctxt, int(p.As)), uint32(p.To.Reg), uint32(r), 0,
				uint32(v), asm)
		} else {
			RIL(a, OP_LGFI, REGTMP, uint32(v), asm)
			RXY(0, zopload(ctxt, int(p.As)), uint32(p.To.Reg), uint32(r), REGTMP,
				0, asm)
		}

	case 40: /* word and byte*/
		wd := uint32(regoff(ctxt, &p.From))
		if p.As == AWORD { //WORD
			*asm = append(*asm, uint8(wd>>24), uint8(wd>>16), uint8(wd>>8), uint8(wd))
		} else { //BYTE
			*asm = append(*asm, uint8(wd))
		}

	case 44: /* indexed store */
		RXY(0, zopstore(ctxt, int(p.As)), uint32(p.From.Reg), uint32(p.To.Index), uint32(p.To.Reg), 0, asm)

	case 45: /* indexed load */
		RXY(0, zopload(ctxt, int(p.As)), uint32(p.To.Reg), uint32(p.From.Index), uint32(p.From.Reg), 0, asm)

	case 47: /* op Ra, Rd; also op [Ra,] Rd */
		switch p.As {
		default:

		case AADDME:
			r := int(p.From.Reg)
			if p.To.Reg == p.From.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.From.Reg), asm)
				r = int(REGTMP)
			}
			RIL(a, OP_LGFI, uint32(p.To.Reg), 0xffffffff, asm) // p.To.Reg <- -1
			RRE(OP_ALCGR, uint32(p.To.Reg), uint32(r), asm)

		case AADDZE:
			r := int(p.From.Reg)
			if p.To.Reg == p.From.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.From.Reg), asm)
				r = int(REGTMP)
			}
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm) // p.To.Reg <- 0
			RRE(OP_ALCGR, uint32(p.To.Reg), uint32(r), asm)

		case ASUBME:
			r := int(p.From.Reg)
			if p.To.Reg == p.From.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.From.Reg), asm)
				r = int(REGTMP)
			}
			RIL(a, OP_LGFI, uint32(p.To.Reg), 0xffffffff, asm) // p.To.Reg <- -1
			RRE(OP_SLBGR, uint32(p.To.Reg), uint32(r), asm)

		case ASUBZE:
			r := int(p.From.Reg)
			if p.To.Reg == p.From.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.From.Reg), asm)
				r = int(REGTMP)
			}
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm) // p.To.Reg <- 0
			RRE(OP_SLBGR, uint32(p.To.Reg), uint32(r), asm)

		case ANEG:
			r := int(p.From.Reg)
			if r == 0 {
				r = int(p.To.Reg)
			}
			RRE(OP_LCGR, uint32(p.To.Reg), uint32(r), asm)
		}

	case 50: /* rem[u] r1[,r2],r3 */
		r := int(p.Reg)

		if r == 0 {
			r = int(p.To.Reg)
		}

		var opcode uint32

		switch p.As {
		default:

		case AREM:
			opcode = OP_DSGFR
		case AREMU:
			opcode = OP_DLR
		}

		if p.As == AREMU {
			RRE(OP_LGR, uint32(REGTMP), uint32(REGZERO), asm)
		}
		RRE(OP_LGR, uint32(REGTMP2), uint32(r), asm)
		RRE(opcode, uint32(REGTMP), uint32(p.From.Reg), asm)
		RRE(OP_LGR, uint32(p.To.Reg), uint32(REGTMP), asm)

	case 51: /* remd[u] r1[,r2],r3 */
		r := int(p.Reg)

		if r == 0 {
			r = int(p.To.Reg)
		}

		var opcode uint32

		switch p.As {
		default:

		case AREMD:
			opcode = OP_DSGR
		case AREMDU:
			opcode = OP_DLGR
		}

		if p.As == AREMDU {
			RRE(OP_LGR, uint32(REGTMP), uint32(REGZERO), asm)
		}
		RRE(OP_LGR, uint32(REGTMP2), uint32(r), asm)
		RRE(opcode, uint32(REGTMP), uint32(p.From.Reg), asm)
		RRE(OP_LGR, uint32(p.To.Reg), uint32(REGTMP), asm)

	case 56: /* sra $sh,[s,]a; srd $sh,[s,]a */
		v := regoff(ctxt, &p.From)
		if v < 0 {
			v = 0
		} else if v > 63 {
			v = 63
		}

		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}

		switch p.As {
		default:
		case ASRAW:
			RSY(OP_SRAK, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		case ASRAD:
			RSY(OP_SRAG, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		}

	case 57: /* slw $sh,[s,]a -> rlwinm ... */
		v := regoff(ctxt, &p.From)
		if v < 0 {
			v = 0
		} else if v > 63 {
			v = 63
		}

		r := int(p.Reg)
		if r == 0 {
			r = int(p.To.Reg)
		}

		switch p.As {
		default:
		case ASLW:
			RSY(OP_SLLK, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		case ASRW:
			RSY(OP_SRLK, uint32(p.To.Reg), uint32(r), 0, uint32(v), asm)
		}

	case 58: /* logical $andcon,[s],a */
		d := regoff(ctxt, &p.From)

		switch p.As {
		case AAND, AOR, AXOR:
			var opcode1, opcode2 uint32
			switch p.As {
			default:
				ctxt.Diag("%v is not supported", p)
			case AAND:
				opcode1 = OP_NGR
				opcode2 = OP_NGRK
			case AOR:
				opcode1 = OP_OGR
				opcode2 = OP_OGRK
			case AXOR:
				opcode1 = OP_XGR
				opcode2 = OP_XGRK
			}

			r := int(p.Reg)
			if r == 0 {
				RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
				RRE(opcode1, uint32(p.To.Reg), REGTMP, asm)
			} else {
				RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
				RRF(opcode2, uint32(r), 0, uint32(p.To.Reg), REGTMP, asm)
			}
		}

	case 59: /* or/and $ucon,,r */
		d := regoff(ctxt, &p.From)

		switch p.As {
		case AAND, AOR, AXOR:
			var opcode1, opcode2 uint32
			switch p.As {
			default:
				ctxt.Diag("%v is not supported", p)
			case AAND:
				opcode1 = OP_NGR
				opcode2 = OP_NGRK
			case AOR:
				opcode1 = OP_OGR
				opcode2 = OP_OGRK
			case AXOR:
				opcode1 = OP_XGR
				opcode2 = OP_XGRK
			}

			r := int(p.Reg)
			if r == 0 {
				RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
				RRE(opcode1, uint32(p.To.Reg), REGTMP, asm)
			} else {
				RIL(a, OP_LGFI, REGTMP, uint32(d), asm)
				RRF(opcode2, uint32(r), 0, uint32(p.To.Reg), REGTMP, asm)
			}
		}

	case 62: /* rlwmi $sh,s,$mask,a */
		v := regoff(ctxt, &p.From)
		d := vregoff(ctxt, p.From3)

		var mask [2]uint8
		maskgen64(ctxt, p, mask[:], uint64(d))

		var i3, i4, i5 int
		i3 = int(mask[0]) // MB
		i4 = int(mask[1]) // ME
		i5 = int(v)

		if i3 > 0x1f || i4 > 0x1f {
			ctxt.Diag("invalid mask for shift: %x (shift %d)\n%v", uint64(d), v, p)
		}

		switch p.As {
		case ARLWMI:
			RIE(f, OP_RISBLG, uint32(p.To.Reg), uint32(p.Reg), 0, uint32(i3), uint32(i4), 0, uint32(i5), asm)

		case ARLWNM:
			r := int(p.Reg)
			if p.To.Reg == p.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.Reg), asm)
				r = int(REGTMP)
			}
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm)
			RIE(f, OP_RISBLG, uint32(p.To.Reg), uint32(r), 0, uint32(i3), uint32(i4), 0, uint32(i5), asm)

		default:

		}

	case 63: /* rlwmi b,s,$mask,a */
		d := vregoff(ctxt, p.From3)
		var mask [2]uint8
		maskgen64(ctxt, p, mask[:], uint64(d))

		var i3, i4 int
		i3 = int(mask[0]) // MB
		i4 = int(mask[1]) // ME

		if i3 > 0x1f || i4 > 0x1f {
			ctxt.Diag("invalid mask for shift: %x (shift)\n%v", uint64(d), p)
		}

		switch p.As {
		case ARLWMI:
			RSY(OP_RLL, uint32(REGTMP), uint32(p.Reg), uint32(p.From.Reg), 0, asm)
			RIE(f, OP_RISBLG, uint32(p.To.Reg), uint32(REGTMP), 0, uint32(i3), uint32(i4), 0, 0, asm)

		case ARLWNM:
			if p.To.Reg == p.Reg {
				RRE(OP_LGR, uint32(REGTMP), uint32(p.Reg), asm)
				RSY(OP_RLL, uint32(REGTMP), uint32(REGTMP), uint32(p.From.Reg), 0, asm)
			} else {
				RSY(OP_RLL, uint32(REGTMP), uint32(p.Reg), uint32(p.From.Reg), 0, asm)
			}
			RRE(OP_LGR, uint32(p.To.Reg), uint32(REGZERO), asm)
			RIE(f, OP_RISBLG, uint32(p.To.Reg), uint32(REGTMP), 0, uint32(i3), uint32(i4), 0, 0, asm)

		default:

		}

	case 68: /* ear arS,rD */
		RRE(OP_EAR, uint32(p.To.Reg), uint32(p.From.Reg-REG_AR0), asm)

	case 69: /* sar rS,arD */
		RRE(OP_SAR, uint32(p.To.Reg-REG_AR0), uint32(p.From.Reg), asm)

	case 70: /* [f]cmp r,r,cr*/
		if p.Reg != 0 {
			ctxt.Diag("unsupported nozero CC in Z")
		}
		if p.As == ACMPW || p.As == ACMPWU {
			RR(zoprr(ctxt, int(p.As)), uint32(p.From.Reg), uint32(p.To.Reg), asm)
		} else {
			RRE(zoprre(ctxt, int(p.As)), uint32(p.From.Reg), uint32(p.To.Reg), asm)
		}

	case 71: /* cmp[l] r,i,cr*/
		if p.Reg != 0 {
			ctxt.Diag("unsupported nozero CC in Z")
		}
		RIL(0, uint32(zopril(ctxt, int(p.As))), uint32(p.From.Reg), uint32(int32(regoff(ctxt, &p.To))), asm)

	/* relocation operations */
	case 74: /* AMOV[F][D|W|H|B|S][Z] Rs, addr -> ST[G|H]RL Rs, addr (with a reloaction entry ) */
		v := regoff(ctxt, &p.To)

		switch p.As {
		case AMOVD:
			RIL(b, OP_STGRL, uint32(p.From.Reg), uint32(v), asm)

		case AMOVW, AMOVWZ: // The zero extension doesn't affect store instructions
			RIL(b, OP_STRL, uint32(p.From.Reg), uint32(v), asm)

		case AMOVH, AMOVHZ: // The zero extension doesn't affect store instructions
			RIL(b, OP_STHRL, uint32(p.From.Reg), uint32(v), asm)

		case AMOVB, AMOVBZ: // The zero extension doesn't affect store instructions
			RIL(b, OP_LARL, REGTMP, 0, asm)
			if v&1 != 0 {
				RX(OP_LA, REGTMP, REGTMP, 0, 1, asm)
				v -= 1
			}
			RX(OP_STC, uint32(p.From.Reg), REGTMP, 0, 0, asm)

		case AFMOVD:
			RIL(b, OP_LARL, REGTMP, uint32(v), asm)
			RX(OP_STD, uint32(p.From.Reg), REGTMP, 0, 0, asm)

		case AFMOVS:
			RIL(b, OP_LARL, REGTMP, uint32(v), asm)
			RX(OP_STE, uint32(p.From.Reg), REGTMP, 0, 0, asm)

		}
		addrilreloc(ctxt, p.To.Sym, int64(v))

	case 75, 76: /* AMOV[F][D|W|H|B|S][Z] addr, Rd -> L[L][F|H]GRL Rd, addr (with a relocation entry) */
		d := regoff(ctxt, &p.From)
		if p.From.Sym == nil {
			RIL(a, OP_LGFI, uint32(p.To.Reg), uint32(d), asm)
		} else {
			switch p.As {
			case AMOVD:
				if d&1 != 0 {
					RIL(b, OP_LARL, REGTMP, 0, asm)
					RX(OP_LA, REGTMP, REGTMP, 0, 1, asm)
					RXY(0, OP_LG, uint32(p.To.Reg), REGTMP, 0, 0, asm)
					d -= 1
				} else {
					RIL(b, OP_LGRL, uint32(p.To.Reg), uint32(d), asm)
				}

			case AMOVW:
				RIL(b, OP_LGFRL, uint32(p.To.Reg), uint32(d), asm)

			case AMOVWZ:
				RIL(b, OP_LLGFRL, uint32(p.To.Reg), uint32(d), asm)

			case AMOVH:
				RIL(b, OP_LGHRL, uint32(p.To.Reg), uint32(d), asm)

			case AMOVHZ:
				RIL(b, OP_LLGHRL, uint32(p.To.Reg), uint32(d), asm)

			case AMOVB, AMOVBZ:
				RIL(b, OP_LARL, REGTMP, 0, asm)
				if d&1 != 0 {
					RX(OP_LA, REGTMP, REGTMP, 0, 1, asm)
					d -= 1
				}

				switch p.As {
				case AMOVB:
					RXY(0, OP_LGB, uint32(p.To.Reg), REGTMP, 0, 0, asm)
				case AMOVBZ:
					RXY(0, OP_LLGC, uint32(p.To.Reg), REGTMP, 0, 0, asm)
				}

			case AFMOVD:
				RIL(a, OP_LARL, REGTMP, uint32(d), asm)
				RX(OP_LD, uint32(p.To.Reg), REGTMP, 0, 0, asm)

			case AFMOVS:
				RIL(a, OP_LARL, REGTMP, uint32(d), asm)
				RX(OP_LE, uint32(p.To.Reg), REGTMP, 0, 0, asm)
			}

			addrilreloc(ctxt, p.From.Sym, int64(d))
		}

	case 77: /* syscall $scon */
		if p.From.Offset > 255 || p.From.Offset < 1 {
			ctxt.Diag("illegal system call; system call number out of range: %v", p)
			E(OP_TRAP2, asm) // trap always
		} else {
			I(OP_SVC, uint32(p.From.Offset), asm)
		}

	case 78: /* undef */
		/* "An instruction consisting entirely of binary 0s is guaranteed
		   always to be an illegal instruction."  */
		*asm = append(*asm, 0, 0, 0, 0)

	case 79: /* cs,csg  r1,r3,off(r2) -> compare & swap; if (r1 ==off(r2)) then off(r2)= r3 */
		v := regoff(ctxt, &p.To)
		if v < 0 {
			v = 0
		}
		if p.As == ACS {
			RS(OP_CS, uint32(p.From.Reg), uint32(p.Reg), uint32(p.To.Reg), uint32(v), asm)
		} else if p.As == ACSG {
			RSY(OP_CSG, uint32(p.From.Reg), uint32(p.Reg), uint32(p.To.Reg), uint32(v), asm)
		}

	case 81: /* SYNC-> BCR 14,0 */
		RR(OP_BCR, 0xE, 0, asm)

	case 82: /* conversion from GPR to FPR */
		var opcode uint32
		switch p.As {
		default:
			log.Fatalf("unexpected opcode %v", p.As)
		case ACEFBRA:
			opcode = OP_CEFBRA
		case ACDFBRA:
			opcode = OP_CDFBRA
		case ACEGBRA:
			opcode = OP_CEGBRA
		case ACDGBRA:
			opcode = OP_CDGBRA
		case ACELFBR:
			opcode = OP_CELFBR
		case ACDLFBR:
			opcode = OP_CDLFBR
		case ACELGBR:
			opcode = OP_CELGBR
		case ACDLGBR:
			opcode = OP_CDLGBR
		}
		/* set immediate operand M3 to 0 to use the default BFP rounding mode
		   (usually round to nearest, ties to even); M4 is reserved and must be 0 */
		RRF(opcode, 0, 0, uint32(p.To.Reg), uint32(p.From.Reg), asm)

	case 83: /* conversion from FPR to GPR */
		var opcode uint32
		switch p.As {
		default:
			log.Fatalf("unexpected opcode %v", p.As)
		case ACFEBRA:
			opcode = OP_CFEBRA
		case ACFDBRA:
			opcode = OP_CFDBRA
		case ACGEBRA:
			opcode = OP_CGEBRA
		case ACGDBRA:
			opcode = OP_CGDBRA
		case ACLFEBR:
			opcode = OP_CLFEBR
		case ACLFDBR:
			opcode = OP_CLFDBR
		case ACLGEBR:
			opcode = OP_CLGEBR
		case ACLGDBR:
			opcode = OP_CLGDBR
		}
		/* set immediate operand M3 to 5 for rounding toward zero (required by Go spec); M4 is reserved and must be 0 */
		RRF(opcode, 5, 0, uint32(p.To.Reg), uint32(p.From.Reg), asm)

	case 84: /* storage-and-storage operations (mvc, clc, xc, oc, nc) */
		l := regoff(ctxt, p.From3)
		d2 := regoff(ctxt, &p.From)
		d1 := regoff(ctxt, &p.To)
		if l < 1 || l > 256 {
			ctxt.Diag("number of bytes (%v) not in range [1,256]", l)
		}
		var opcode uint32
		switch p.As {
		default:
			ctxt.Diag("unexpected opcode %v", p.As)
		case AMVC:
			opcode = OP_MVC
		case ACLC:
			opcode = OP_CLC
		case AXC:
			opcode = OP_XC
		case AOC:
			opcode = OP_OC
		case ANC:
			opcode = OP_NC
		}
		SS(a, opcode, uint32(l-1), 0, uint32(p.To.Reg), uint32(d1), uint32(p.From.Reg), uint32(d2), asm)

	case 85: /* larl: load address relative long */
		// When using larl directly, don't add a nop
		v := regoff(ctxt, &p.From)
		if p.From.Sym == nil {
			if (v & 1) != 0 {
				ctxt.Diag("cannot use LARL with odd offset: %v", v)
			}
		} else {
			addrilreloc(ctxt, p.From.Sym, int64(v))
			v = 0
		}
		RIL(b, OP_LARL, uint32(p.To.Reg), uint32(v>>1), asm)

	case 86: /* lay?: load address */
		v := vregoff(ctxt, &p.From)
		switch p.As {
		case ALA:
			RX(OP_LA, uint32(p.To.Reg), uint32(p.From.Reg), 0, uint32(v), asm)
		case ALAY:
			RXY(0, OP_LAY, uint32(p.To.Reg), uint32(p.From.Reg), 0, uint32(v), asm)
		}

	case 87: /* exrl: execute relative long */
		v := vregoff(ctxt, &p.From)
		if p.From.Sym == nil {
			if v&1 != 0 {
				ctxt.Diag("cannot use EXRL with odd offset: %v", v)
			}
		} else {
			addrilreloc(ctxt, p.From.Sym, int64(v))
			v = 0
		}
		RIL(b, OP_EXRL, uint32(p.To.Reg), uint32(v>>1), asm)

	case 88: /* stck[cef]?: store clock (comparator/extended/fast) */
		var opcode uint32
		switch p.As {
		case ASTCK:
			opcode = OP_STCK
		case ASTCKC:
			opcode = OP_STCKC
		case ASTCKE:
			opcode = OP_STCKE
		case ASTCKF:
			opcode = OP_STCKF
		}
		v := vregoff(ctxt, &p.To)
		r := int(p.To.Reg)
		if r == 0 {
			r = int(o.param)
		}
		S(opcode, uint32(r), uint32(v), asm)

	case 89:
		if p.Pcond == nil {
			ctxt.Diag("no branch target")
			break
		}
		v := int32((p.Pcond.Pc - p.Pc) >> 1)
		var opcode, opcode2 uint32
		switch p.As {
		case ACMPBEQ, ACMPBGE, ACMPBGT, ACMPBLE, ACMPBLT, ACMPBNE:
			opcode = OP_CGRJ
			opcode2 = OP_CGR
		case ACMPUBEQ, ACMPUBGE, ACMPUBGT, ACMPUBLE, ACMPUBLT, ACMPUBNE:
			opcode = OP_CLGRJ
			opcode2 = OP_CLGR
		}
		mask := branchMask(ctxt, p)
		if int32(int16(v)) != v {
			RRE(opcode2, uint32(p.From.Reg), uint32(p.Reg), asm)
			RIL(c, OP_BRCL, uint32(mask), uint32(v-FORMAT_RRE_size/2), asm)
		} else {
			RIE(b, opcode, uint32(p.From.Reg), uint32(p.Reg), uint32(v), 0, 0, mask, 0, asm)
		}

	case 90:
		if p.Pcond == nil {
			ctxt.Diag("no branch target")
			break
		}
		var opcode, opcode2 uint32
		switch p.As {
		case ACMPBEQ, ACMPBGE, ACMPBGT, ACMPBLE, ACMPBLT, ACMPBNE:
			opcode = OP_CGIJ
			opcode2 = OP_CGFI
		case ACMPUBEQ, ACMPUBGE, ACMPUBGT, ACMPUBLE, ACMPUBLT, ACMPUBNE:
			opcode = OP_CLGIJ
			opcode2 = OP_CLGFI
		}
		v := int32((p.Pcond.Pc - p.Pc) >> 1)
		mask := branchMask(ctxt, p)
		if int32(int16(v)) != v {
			RIL(0, opcode2, uint32(p.From.Reg), uint32(int32(regoff(ctxt, p.From3))), asm)
			RIL(c, OP_BRCL, mask, uint32(v-FORMAT_RIL_size/2), asm)
		} else {
			RIE(c, opcode, uint32(p.From.Reg), mask, uint32(v), 0, 0, 0, uint32(int32(regoff(ctxt, p.From3))), asm)
		}

	case 92:
		var opcode uint32
		switch p.As {
		case AMOVD:
			opcode = OP_MVGHI
		case AMOVW, AMOVWZ:
			opcode = OP_MVHI
		case AMOVH, AMOVHZ:
			opcode = OP_MVHHI
		case AMOVB, AMOVBZ:
			opcode = OP_MVI
		}
		v := regoff(ctxt, &p.From)
		d := regoff(ctxt, &p.To)
		if opcode == OP_MVI {
			SI(opcode, uint32(v), uint32(p.To.Reg), uint32(d), asm)
		} else {
			SIL(opcode, uint32(p.To.Reg), uint32(d), uint32(v), asm)
		}

	}
}

func vregoff(ctxt *obj.Link, a *obj.Addr) int64 {
	ctxt.Instoffset = 0
	if a != nil {
		aclass(ctxt, a)
	}
	return ctxt.Instoffset
}

func regoff(ctxt *obj.Link, a *obj.Addr) int32 {
	return int32(vregoff(ctxt, a))
}

/*
 * load o(a), d
 */
func zopload(ctxt *obj.Link, a int) uint32 {
	switch a {
	/* fixed point load */
	case AMOVD:
		return uint32(OP_LG)
	case AMOVW:
		return uint32(OP_LGF)
	case AMOVWZ:
		return uint32(OP_LLGF)
	case AMOVH:
		return uint32(OP_LGH)
	case AMOVHZ:
		return uint32(OP_LLGH)
	case AMOVB:
		return uint32(OP_LGB)
	case AMOVBZ:
		return uint32(OP_LLGC)

	/* floating pointer load*/
	case AFMOVD:
		return uint32(OP_LDY)
	case AFMOVS:
		return uint32(OP_LEY)

	/* byte reversed load*/
	case AMOVWBR:
		return uint32(OP_LRV)
	case AMOVHBR:
		return uint32(OP_LRVH)
	}

	ctxt.Diag("unknown store opcode %v", obj.Aconv(a))
	return 0
}

/*
 * store s,o(d)
 */
func zopstore(ctxt *obj.Link, a int) uint32 {
	switch a {
	/* fixed point store */
	case AMOVD:
		return uint32(OP_STG)
	case AMOVW, AMOVWZ:
		return uint32(OP_STY)
	case AMOVH, AMOVHZ:
		return uint32(OP_STHY)
	case AMOVB, AMOVBZ:
		return uint32(OP_STCY)

	/* floating point store */
	case AFMOVD:
		return uint32(OP_STDY)
	case AFMOVS:
		return uint32(OP_STEY)

	/* byte reversed store */
	case AMOVWBR:
		return uint32(OP_STRV)
	case AMOVHBR:
		return uint32(OP_STRVH)
	}

	ctxt.Diag("unknown store opcode %v", obj.Aconv(a))
	return 0
}

func zoprre(ctxt *obj.Link, a int) uint32 {
	switch a {
	case ACMP:
		return uint32(OP_CGR)
	case ACMPU:
		return uint32(OP_CLGR)
	case AFCMPO: //ordered
		return uint32(OP_KDBR)
	case AFCMPU: //unordered
		return uint32(OP_CDBR)
	case ACEBR:
		return uint32(OP_CEBR)
	}
	ctxt.Diag("unknown rre opcode %v", obj.Aconv(a))
	return 0
}

func zoprr(ctxt *obj.Link, a int) uint32 {
	switch a {
	case ACMPW:
		return uint32(OP_CR)
	case ACMPWU:
		return uint32(OP_CLR)
	}
	ctxt.Diag("unknown rr opcode %v", obj.Aconv(a))
	return 0
}

func zopril(ctxt *obj.Link, a int) uint32 {
	switch a {
	case ACMP:
		return uint32(OP_CGFI)
	case ACMPU:
		return uint32(OP_CLGFI)
	case ACMPW:
		return uint32(OP_CFI)
	case ACMPWU:
		return uint32(OP_CLFI)
	}
	ctxt.Diag("unknown ril opcode %v", obj.Aconv(a))
	return 0
}

// z instructions sizes.
const (
	FORMAT_E_size    = 2
	FORMAT_I_size    = 2
	FORMAT_IE_size   = 4
	FORMAT_MII_size  = 6
	FORMAT_RI_size   = 4
	FORMAT_RI1_size  = 4
	FORMAT_RI2_size  = 4
	FORMAT_RI3_size  = 4
	FORMAT_RIE_size  = 6
	FORMAT_RIE1_size = 6
	FORMAT_RIE2_size = 6
	FORMAT_RIE3_size = 6
	FORMAT_RIE4_size = 6
	FORMAT_RIE5_size = 6
	FORMAT_RIE6_size = 6
	FORMAT_RIL_size  = 6
	FORMAT_RIL1_size = 6
	FORMAT_RIL2_size = 6
	FORMAT_RIL3_size = 6
	FORMAT_RIS_size  = 6
	FORMAT_RR_size   = 2
	FORMAT_RRD_size  = 4
	FORMAT_RRE_size  = 4
	FORMAT_RRF_size  = 4
	FORMAT_RRF1_size = 4
	FORMAT_RRF2_size = 4
	FORMAT_RRF3_size = 4
	FORMAT_RRF4_size = 4
	FORMAT_RRF5_size = 4
	FORMAT_RRR_size  = 2
	FORMAT_RRS_size  = 6
	FORMAT_RS_size   = 4
	FORMAT_RS1_size  = 4
	FORMAT_RS2_size  = 4
	FORMAT_RSI_size  = 4
	FORMAT_RSL_size  = 6
	FORMAT_RSY_size  = 6
	FORMAT_RSY1_size = 6
	FORMAT_RSY2_size = 6
	FORMAT_RX_size   = 4
	FORMAT_RX1_size  = 4
	FORMAT_RX2_size  = 4
	FORMAT_RXE_size  = 6
	FORMAT_RXF_size  = 6
	FORMAT_RXY_size  = 6
	FORMAT_RXY1_size = 6
	FORMAT_RXY2_size = 6
	FORMAT_S_size    = 4
	FORMAT_SI_size   = 4
	FORMAT_SIL_size  = 6
	FORMAT_SIY_size  = 6
	FORMAT_SMI_size  = 6
	FORMAT_SS_size   = 6
	FORMAT_SS1_size  = 6
	FORMAT_SS2_size  = 6
	FORMAT_SS3_size  = 6
	FORMAT_SS4_size  = 6
	FORMAT_SS5_size  = 6
	FORMAT_SS6_size  = 6
	FORMAT_SSE_size  = 6
	FORMAT_SSF_size  = 6
)

// instruction format variations.
const (
	a = iota
	b
	c
	d
	e
	f
	g
)

func E(op uint32, asm *[]byte) {
	*asm = append(*asm, uint8(op>>8), uint8(op))
}

func I(op, i1 uint32, asm *[]byte) {
	*asm = append(*asm, uint8(op>>8), uint8(i1))
}

func MII(op, m1, ri2, ri3 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(m1)<<4)|uint8((ri2>>8)&0x0F),
		uint8(ri2),
		uint8(ri3>>16),
		uint8(ri3>>8),
		uint8(ri3))
}

func RI(op, r1_m1, i2_ri2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1_m1)<<4)|(uint8(op)&0x0F),
		uint8(i2_ri2>>8),
		uint8(i2_ri2))
}

// Expected argument values for the instruction formats.
//
// Format     a1    a2     a3     a4     a5  a6  a7
// ------------------------------------
// a         r1,  0,  i2,  0,  0, m3,  0
// b         r1, r2, ri4,  0,  0, m3,  0
// c         r1, m3, ri4,  0,  0,  0, i2
// d         r1, r3,  i2,  0,  0,  0,  0
// e         r1, r3, ri2,  0,  0,  0,  0
// f         r1, r2,   0, i3, i4,  0, i5
// g         r1, m3,  i2,  0,  0,  0,  0
func RIE(type_, op, r1, r2_m3_r3, i2_ri4_ri2, i3, i4, m3, i2_i5 uint32, asm *[]byte) {
	*asm = append(*asm, uint8(op>>8), uint8(r1)<<4|uint8(r2_m3_r3&0x0F))

	switch type_ {
	default:
		*asm = append(*asm, uint8(i2_ri4_ri2>>8), uint8(i2_ri4_ri2))
	case f:
		*asm = append(*asm, uint8(i3), uint8(i4))
	}

	switch type_ {
	case a, b:
		*asm = append(*asm, uint8(m3)<<4)
	default:
		*asm = append(*asm, uint8(i2_i5))
	}

	*asm = append(*asm, uint8(op))
}

func RIL(type_, op, r1_m1, i2_ri2 uint32, asm *[]byte) {
	if type_ == a || type_ == b {
		r1_m1 = r1_m1 - obj.RBaseS390X // this is a register base
	}
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1_m1)<<4)|(uint8(op)&0x0F),
		uint8(i2_ri2>>24),
		uint8(i2_ri2>>16),
		uint8(i2_ri2>>8),
		uint8(i2_ri2))
}

func RIS(op, r1, m3, b4, d4, i2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(m3&0x0F),
		(uint8(b4)<<4)|(uint8(d4>>8)&0x0F),
		uint8(d4),
		uint8(i2),
		uint8(op))
}

func RR(op, r1, r2 uint32, asm *[]byte) {
	*asm = append(*asm, uint8(op>>8), (uint8(r1)<<4)|uint8(r2&0x0F))
}

func RRD(op, r1, r3, r2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		uint8(r1)<<4,
		(uint8(r3)<<4)|uint8(r2&0x0F))
}

func RRE(op, r1, r2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		0,
		(uint8(r1)<<4)|uint8(r2&0x0F))
}

func RRF(op, r3_m3, m4, r1, r2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		(uint8(r3_m3)<<4)|uint8(m4&0x0F),
		(uint8(r1)<<4)|uint8(r2&0x0F))
}

func RRS(op, r1, r2, b4, d4, m3 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(r2&0x0F),
		(uint8(b4)<<4)|uint8((d4>>8)&0x0F),
		uint8(d4),
		uint8(m3)<<4,
		uint8(op))
}

func RS(op, r1, r3_m3, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(r3_m3&0x0F),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2))
}

func RSI(op, r1, r3, ri2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(r3&0x0F),
		uint8(ri2>>8),
		uint8(ri2))
}

func RSL(type_, op, l1, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(l1),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2),
		uint8(op))
}

// (20b) d2 with (12b) dl2 and (8b) dh2.
func RSY(op, r1, r3_m3, b2, d2 uint32, asm *[]byte) {
	dl2 := uint16(d2) & 0x0FFF
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(r3_m3&0x0F),
		(uint8(b2)<<4)|(uint8(dl2>>8)&0x0F),
		uint8(dl2),
		uint8(d2>>12),
		uint8(op))
}

func RX(op, r1_m1, x2, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1_m1)<<4)|uint8(x2&0x0F),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2))
}

func RXE(op, r1, x2, b2, d2, m3 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1)<<4)|uint8(x2&0x0F),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2),
		uint8(m3)<<4,
		uint8(op))
}

func RXF(op, r3, x2, b2, d2, m1 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r3)<<4)|uint8(x2&0x0F),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2),
		uint8(m1)<<4,
		uint8(op))
}

func RXY(type_, op, r1_m1, x2, b2, d2 uint32, asm *[]byte) {
	dl2 := uint16(d2) & 0x0FFF
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r1_m1)<<4)|uint8(x2&0x0F),
		(uint8(b2)<<4)|(uint8(dl2>>8)&0x0F),
		uint8(dl2),
		uint8(d2>>12),
		uint8(op))
}

func S(op, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2))
}

func SI(op, i2, b1, d1 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(i2),
		(uint8(b1)<<4)|uint8((d1>>8)&0x0F),
		uint8(d1))
}

func SIL(op, b1, d1, i2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		(uint8(b1)<<4)|uint8((d1>>8)&0x0F),
		uint8(d1),
		uint8(i2>>8),
		uint8(i2))
}

func SIY(op, i2, b1, d1 uint32, asm *[]byte) {
	dl1 := uint16(d1) & 0x0FFF
	*asm = append(*asm,
		uint8(op>>8),
		uint8(i2),
		(uint8(b1)<<4)|(uint8(dl1>>8)&0x0F),
		uint8(dl1),
		uint8(d1>>12),
		uint8(op))
}

func SMI(op, m1, b3, d3, ri2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(m1)<<4,
		(uint8(b3)<<4)|uint8((d3>>8)&0x0F),
		uint8(d3),
		uint8(ri2>>8),
		uint8(ri2))
}

// Expected argument values for the instruction formats.
//
// Format    a1  a2  a3  a4  a5  a6
// -------------------------------
// a         l1,  0, b1, d1, b2, d2
// b         l1, l2, b1, d1, b2, d2
// c         l1, i3, b1, d1, b2, d2
// d         r1, r3, b1, d1, b2, d2
// e         r1, r3, b2, d2, b4, d4
// f          0, l2, b1, d1, b2, d2
func SS(type_, op, l1_r1, l2_i3_r3, b1_b2, d1_d2, b2_b4, d2_d4 uint32, asm *[]byte) {
	*asm = append(*asm, uint8(op>>8))

	switch type_ {
	case a:
		*asm = append(*asm, uint8(l1_r1))
	case b, c, d, e:
		*asm = append(*asm, (uint8(l1_r1)<<4)|uint8(l2_i3_r3&0x0F))
	case f:
		*asm = append(*asm, uint8(l2_i3_r3))
	}

	*asm = append(*asm,
		(uint8(b1_b2)<<4)|uint8((d1_d2>>8)&0x0F),
		uint8(d1_d2),
		(uint8(b2_b4)<<4)|uint8((d2_d4>>8)&0x0F),
		uint8(d2_d4))
}

func SSE(op, b1, d1, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		uint8(op),
		(uint8(b1)<<4)|uint8((d1>>8)&0x0F),
		uint8(d1),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2))
}

func SSF(op, r3, b1, d1, b2, d2 uint32, asm *[]byte) {
	*asm = append(*asm,
		uint8(op>>8),
		(uint8(r3)<<4)|(uint8(op)&0x0F),
		(uint8(b1)<<4)|uint8((d1>>8)&0x0F),
		uint8(d1),
		(uint8(b2)<<4)|uint8((d2>>8)&0x0F),
		uint8(d2))
}
