// Based on cmd/internal/obj/ppc64/a.out.go.
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2008 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2008 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
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

import "cmd/internal/obj"

//go:generate go run ../stringer.go -i $GOFILE -o anames.go -p s390x

/*
 * s390x
 */
const (
	NSNAME = 8
	NSYM   = 50
	NREG   = 16 /* number of general registers */
	NFREG  = 16 /* number of floating point registers */
)

const (
	REG_R0 = obj.RBaseS390X + iota
	REG_R1
	REG_R2
	REG_R3
	REG_R4
	REG_R5
	REG_R6
	REG_R7
	REG_R8
	REG_R9
	REG_R10
	REG_R11
	REG_R12
	REG_R13
	REG_R14
	REG_R15

	REG_F0
	REG_F1
	REG_F2
	REG_F3
	REG_F4
	REG_F5
	REG_F6
	REG_F7
	REG_F8
	REG_F9
	REG_F10
	REG_F11
	REG_F12
	REG_F13
	REG_F14
	REG_F15

	REG_AR0
	REG_AR1
	REG_AR2
	REG_AR3
	REG_AR4
	REG_AR5
	REG_AR6
	REG_AR7
	REG_AR8
	REG_AR9
	REG_AR10
	REG_AR11
	REG_AR12
	REG_AR13
	REG_AR14
	REG_AR15

	REG_RESERVED // first of 1024 reserved registers

	REGZERO  = REG_R0 /* set to zero */
	REGRET   = REG_R2
	REGARG   = -1      /* -1 disables passing the first argument in register */
	REGRT1   = REG_R3  /* reserved for runtime, duffzero and duffcopy (does this need to be reserved on z?) */
	REGRT2   = REG_R4  /* reserved for runtime, duffcopy (does this need to be reserved on z?) */
	REGMIN   = REG_R5  /* register variables allocated from here to REGMAX */
	REGTMP   = REG_R10 /* used by the linker */
	REGTMP2  = REG_R11 /* used by the linker */
	REGCTXT  = REG_R12 /* context for closures */
	REGG     = REG_R13 /* G */
	REG_LR   = REG_R14 /* link register */
	REGSP    = REG_R15 /* stack pointer */
	REGEXT   = REG_R9  /* external registers allocated from here down */
	REGMAX   = REG_R8
	FREGRET  = REG_F0
	FREGMIN  = REG_F5  /* first register variable */
	FREGMAX  = REG_F10 /* last register variable for zg only */
	FREGEXT  = REG_F10 /* first external register */
	FREGCVI  = REG_F11 /* floating conversion constant */
	FREGZERO = REG_F12 /* both float and double */
	FREGONE  = REG_F13 /* double */
	FREGTWO  = REG_F14 /* double */
//	FREGTMP  = REG_F15 /* double */
)

/*
 * GENERAL:
 *
 * compiler allocates R3 up as temps
 * compiler allocates register variables R5-R9
 * compiler allocates external registers R10 down
 *
 * compiler allocates register variables F5-F9
 * compiler allocates external registers F10 down
 */
const (
	BIG    = 32768 - 8
	DISP12 = 4096
	DISP16 = 65536
	DISP20 = 1048576
)

const (
	/* mark flags */
	LABEL   = 1 << 0
	LEAF    = 1 << 1
	FLOAT   = 1 << 2
	BRANCH  = 1 << 3
	LOAD    = 1 << 4
	FCMP    = 1 << 5
	SYNC    = 1 << 6
	LIST    = 1 << 7
	FOLL    = 1 << 8
	NOSCHED = 1 << 9
)

const ( // comments from func aclass in asmz.go
	C_NONE   = iota
	C_REG    // general-purpose register
	C_FREG   // floating-point register
	C_AREG   // access register
	C_ZCON   // constant == 0
	C_SCON   // 0 <= constant <= 0x7fff (positive int16)
	C_UCON   // constant & 0xffff == 0 (int32 or uint32)
	C_ADDCON // 0 > constant >= -0x8000 (negative int16)
	C_ANDCON // constant <= 0xffff
	C_LCON   // constant (int32 or uint32)
	C_DCON   // constant (int64 or uint64)
	C_SACON  // computed address, 16-bit displacement, possibly SP-relative
	C_SECON  // computed address, 16-bit displacement, possibly SB-relative, unused?
	C_LACON  // computed address, 32-bit displacement, possibly SP-relative
	C_LECON  // computed address, 32-bit displacement, possibly SB-relative, unused?
	C_DACON  // computed address, 64-bit displacment?
	C_SBRA   // short branch
	C_LBRA   // long branch
	C_SAUTO  // short auto
	C_LAUTO  // long auto
	C_ZOREG  // heap address, register-based, displacement == 0
	C_SOREG  // heap address, register-based, int16 displacement
	C_LOREG  // heap address, register-based, int32 displacement
	C_ANY
	C_GOK      // general address
	C_ADDR     // relocation for extern or static symbols
	C_TEXTSIZE // text size
	C_NCLASS   // must be the last
)

const (
	// integer arithmetic
	AADD = obj.ABaseS390X + obj.A_ARCHSPECIFIC + iota
	AADDC
	AADDME
	AADDE
	AADDZE
	ADIVW
	ADIVWU
	ADIVD
	ADIVDU
	AMULLW
	AMULLD
	AMULHD
	AMULHDU
	ASUB
	ASUBC
	ASUBME
	ASUBV
	ASUBE
	ASUBZE
	AREM
	AREMU
	AREMD
	AREMDU
	ANEG

	// integer moves
	AMOVWBR
	AMOVB
	AMOVBZ
	AMOVH
	AMOVHBR
	AMOVHZ
	AMOVW
	AMOVWZ
	AMOVD

	// integer bitwise
	AAND
	AANDN
	ANAND
	ANOR
	AOR
	AORN
	AXOR
	ASLW
	ASLD
	ASRW
	ASRAW
	ASRD
	ASRAD
	ARLWMI
	ARLWNM
	ARLDMI
	ARLDC
	ARLDCR
	ARLDCL

	// floating point
	AFABS
	AFADD
	AFADDS
	AFCMPO
	AFCMPU
	ACEBR
	AFDIV
	AFDIVS
	AFMADD
	AFMADDS
	AFMOVD
	AFMOVS
	AFMSUB
	AFMSUBS
	AFMUL
	AFMULS
	AFNABS
	AFNEG
	AFNMADD
	AFNMADDS
	AFNMSUB
	AFNMSUBS
	AFRSP
	ALDEBR
	AFSUB
	AFSUBS
	AFSQRT
	AFSQRTS

	// convert from int32/int64 to float/float64
	ACEFBRA
	ACDFBRA
	ACEGBRA
	ACDGBRA

	// convert from float/float64 to int32/int64
	ACFEBRA
	ACFDBRA
	ACGEBRA
	ACGDBRA

	// convert from uint32/uint64 to float/float64
	ACELFBR
	ACDLFBR
	ACELGBR
	ACDLGBR

	// convert from float/float64 to uint32/uint64
	ACLFEBR
	ACLFDBR
	ACLGEBR
	ACLGDBR

	// compare
	ACMP
	ACMPU
	ACMPW
	ACMPWU

	// compare and swap
	ACS
	ACSG

	// serialize
	ASYNC

	// branch
	ABC
	ABCL
	ABEQ
	ABGE
	ABGT
	ABLE
	ABLT
	ABNE
	ABVC
	ABVS
	ASYSCALL

	// compare and branch
	ACMPBEQ
	ACMPBGE
	ACMPBGT
	ACMPBLE
	ACMPBLT
	ACMPBNE
	ACMPUBEQ
	ACMPUBGE
	ACMPUBGT
	ACMPUBLE
	ACMPUBLT
	ACMPUBNE

	// storage-and-storage
	AMVC
	ACLC
	AXC
	AOC
	ANC

	// load
	AEXRL
	ALARL
	ALA
	ALAY

	// store clock
	ASTCK
	ASTCKC
	ASTCKE
	ASTCKF

	// binary
	ABYTE
	AWORD
	ADWORD

	// end marker
	ALAST

	// aliases
	ABR = obj.AJMP
	ABL = obj.ACALL
)
