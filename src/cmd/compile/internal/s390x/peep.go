// Derived from Inferno utils/6c/peep.c
// http://code.google.com/p/inferno-os/source/browse/utils/6c/peep.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
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

import (
	"cmd/compile/internal/gc"
	"cmd/internal/obj"
	"cmd/internal/obj/s390x"
	"fmt"
)

var gactive uint32
var mergeopmv_cnt int
var cnb_cnt int

func peep(firstp *obj.Prog) {
	g := (*gc.Graph)(gc.Flowstart(firstp, nil))
	if g == nil {
		return
	}
	gactive = 0

	var p *obj.Prog

	if false {
		// constant propagation
		// find MOV $con,R followed by
		// another MOV $con,R without
		// setting R in the interim
		for r := (*gc.Flow)(g.Start); r != nil; r = r.Link {
			p = r.Prog
			switch p.As {
			case s390x.AMOVB,
				s390x.AMOVW,
				s390x.AMOVD:
				if regtyp(&p.To) {
					if p.From.Type == obj.TYPE_CONST || p.From.Type == obj.TYPE_FCONST {
						conprop(r)
					}
				}
			}
		}
	}

	var r *gc.Flow
	var t int
loop1:
	//	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
	//		gc.Dumpit("loop1", g.Start, 0)
	//	}

	t = 0
	for r = g.Start; r != nil; r = r.Link {
		p = r.Prog

		// TODO(austin) Handle smaller moves.  arm and amd64
		// distinguish between moves that moves that *must*
		// sign/zero extend and moves that don't care so they
		// can eliminate moves that don't care without
		// breaking moves that do care.  This might let us
		// simplify or remove the next peep loop, too.
		if p.As == s390x.AMOVD || p.As == s390x.AFMOVD {
			if regtyp(&p.To) {
				// Try to eliminate reg->reg moves
				if regtyp(&p.From) {
					if p.From.Type == p.To.Type {
						if copyprop(r) {
							excise(r)
							t++
						} else if subprop(r) && copyprop(r) {
							excise(r)
							t++
						}
					}
				}

				// Convert uses to $0 to uses of R0 and
				// propagate R0
				if regzer(&p.From) != 0 {
					if p.To.Type == obj.TYPE_REG {
						p.From.Type = obj.TYPE_REG
						p.From.Reg = s390x.REGZERO
						if copyprop(r) {
							excise(r)
							t++
						} else if subprop(r) && copyprop(r) {
							excise(r)
							t++
						}
					}
				}
			}
		}
	}

	if t != 0 {
		goto loop1
	}

	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		gc.Dumpit("pass7 copyprop", g.Start, 0)
	}

	/*
	 * look for MOVB x,R; MOVB R,R (for small MOVs not handled above)
	 */
	var p1 *obj.Prog
	var r1 *gc.Flow
	for r := (*gc.Flow)(g.Start); r != nil; r = r.Link {
		p = r.Prog
		switch p.As {
		default:
			continue

		case s390x.AMOVH,
			s390x.AMOVHZ,
			s390x.AMOVB,
			s390x.AMOVBZ,
			s390x.AMOVW,
			s390x.AMOVWZ:
			if p.To.Type != obj.TYPE_REG {
				continue
			}
		}

		r1 = r.Link
		if r1 == nil {
			continue
		}
		p1 = r1.Prog
		if p1.As != p.As {
			continue
		}
		if p1.From.Type != obj.TYPE_REG || p1.From.Reg != p.To.Reg {
			continue
		}
		if p1.To.Type != obj.TYPE_REG || p1.To.Reg != p.To.Reg {
			continue
		}
		excise(r1)
	}

	if gc.Debug['P'] > 1 {
		goto ret /* allow following code improvement to be suppressed */
	}

	if gc.Debug['p'] == 0 {
		// load pipelining
		// push any load from memory as early as possible
		// to give it time to complete before use.
		for r := (*gc.Flow)(g.Start); r != nil; r = r.Link {
			p = r.Prog
			switch p.As {
			case s390x.AMOVB,
				s390x.AMOVW,
				s390x.AMOVD:

				if regtyp(&p.To) && !regconsttyp(&p.From) {
					pushback(r)
				}
			}
		}
		if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
			gc.Dumpit("pass8 push load as early as possible", g.Start, 0)
		}

	}

	/*
	 * look for OP a, b, c; MOV c, d; -> OP a, b, d;
	 */

	for r := (*gc.Flow)(g.Start); r != nil; r = r.Link {

		if (gc.Debugmergeopmv != -1) && (mergeopmv_cnt >= gc.Debugmergeopmv) {
			break
		}

		p = r.Prog

		switch p.As {
		case s390x.AADD,
			s390x.AADDC,
			s390x.AADDME,
			s390x.AADDE,
			s390x.AADDZE,
			s390x.AAND,
			s390x.AANDN,
			s390x.ADIVW,
			s390x.ADIVWU,
			s390x.ADIVD,
			s390x.ADIVDU,
			s390x.AMULLW,
			s390x.AMULHD,
			s390x.AMULHDU,
			s390x.AMULLD,
			s390x.ANAND,
			s390x.ANOR,
			s390x.AOR,
			s390x.AORN,
			s390x.AREM,
			s390x.AREMU,
			s390x.AREMD,
			s390x.AREMDU,
			s390x.ARLWMI,
			s390x.ARLWNM,
			s390x.ASLW,
			s390x.ASRAW,
			s390x.ASRW,
			s390x.ASLD,
			s390x.ASRAD,
			s390x.ASRD,
			s390x.ASUB,
			s390x.ASUBC,
			s390x.ASUBME,
			s390x.ASUBE,
			s390x.ASUBZE,
			s390x.AXOR:
			if p.To.Type != obj.TYPE_REG {
				continue
			}
			if p.Reg == 0 { // Only for 3 ops instruction
				continue
			}
		default:
			continue
		}

		r1 := r.Link
		for ; r1 != nil; r1 = r1.Link {
			if r1.Prog.As != obj.ANOP {
				break
			}
		}

		if r1 == nil {
			continue
		}

		p1 := r1.Prog
		switch p1.As {
		case s390x.AMOVD,
			s390x.AMOVW, s390x.AMOVWZ,
			s390x.AMOVH, s390x.AMOVHZ,
			s390x.AMOVB, s390x.AMOVBZ:
			if p1.To.Type != obj.TYPE_REG {
				continue
			}

		default:
			continue
		}
		if p1.From.Type != obj.TYPE_REG || p1.From.Reg != p.To.Reg {
			continue
		}

		if trymergeopmv(r1) {
			p.To = p1.To
			excise(r1)
			mergeopmv_cnt += 1
		}
	}

	if gc.Debug['v'] != 0 {
		gc.Dumpit("Merge operation and move", g.Start, 0)
	}

	/*
	 * look for CMP x, y; Branch -> Compare and branch
	 */

	if gc.Debugcnb == 0 {
		goto ret
	}

	for r := (*gc.Flow)(g.Start); r != nil; r = r.Link {
		if (gc.Debugcnb != -1) && (cnb_cnt >= gc.Debugcnb) {
			break
		}
		p = r.Prog
		r1 = gc.Uniqs(r)
		if r1 == nil {
			continue
		}
		p1 = r1.Prog

		switch p.As {
		case s390x.ACMP:
			switch p1.As {
			case s390x.ABCL, s390x.ABC:
				continue
			case s390x.ABEQ:
				t = s390x.ACMPBEQ
			case s390x.ABGE:
				t = s390x.ACMPBGE
			case s390x.ABGT:
				t = s390x.ACMPBGT
			case s390x.ABLE:
				t = s390x.ACMPBLE
			case s390x.ABLT:
				t = s390x.ACMPBLT
			case s390x.ABNE:
				t = s390x.ACMPBNE
			default:
				continue
			}

		case s390x.ACMPU:
			switch p1.As {
			case s390x.ABCL, s390x.ABC:
				continue
			case s390x.ABEQ:
				t = s390x.ACMPUBEQ
			case s390x.ABGE:
				t = s390x.ACMPUBGE
			case s390x.ABGT:
				t = s390x.ACMPUBGT
			case s390x.ABLE:
				t = s390x.ACMPUBLE
			case s390x.ABLT:
				t = s390x.ACMPUBLT
			case s390x.ABNE:
				t = s390x.ACMPUBNE
			default:
				continue
			}

		case s390x.ACMPW, s390x.ACMPWU:
			continue

		default:
			continue
		}

		if gc.Debug['D'] != 0 {
			fmt.Printf("cnb %v; %v -> ", p, p1)
		}

		if p1.To.Sym != nil {
			continue
		}

		if p.To.Type == obj.TYPE_REG {
			p1.As = int16(t)
			p1.From = p.From
			p1.Reg = p.To.Reg
			p1.From3 = nil
		} else if p.To.Type == obj.TYPE_CONST {
			switch p.As {
			case s390x.ACMP, s390x.ACMPW:
				if (p.To.Offset < -(1 << 7)) || (p.To.Offset >= ((1 << 7) - 1)) {
					continue
				}
			case s390x.ACMPU, s390x.ACMPWU:
				if p.To.Offset >= (1 << 8) {
					continue
				}
			default:
			}
			p1.As = int16(t)
			p1.From = p.From
			p1.Reg = 0
			p1.From3 = new(obj.Addr)
			*(p1.From3) = p.To
		} else {
			continue
		}

		if gc.Debug['D'] != 0 {
			fmt.Printf("%v\n", p1)
		}
		cnb_cnt += 1
		excise(r)
	}

	if gc.Debug['v'] != 0 {
		gc.Dumpit("compare and branch", g.Start, 0)
	}

ret:
	gc.Flowend(g)
}

func conprop(r0 *gc.Flow) {
	var p *obj.Prog
	var t int

	p0 := (*obj.Prog)(r0.Prog)
	v0 := (*obj.Addr)(&p0.To)
	r := (*gc.Flow)(r0)

loop:
	r = gc.Uniqs(r)
	if r == nil || r == r0 {
		return
	}
	if gc.Uniqp(r) == nil {
		return
	}

	p = r.Prog
	t = copyu(p, v0, nil)
	switch t {
	case 0, // miss
		1: // use
		goto loop

	case 2, // rar
		4: // use and set
		break

	case 3: // set
		if p.As == p0.As {
			if p.From.Type == p0.From.Type {
				if p.From.Reg == p0.From.Reg {
					if p.From.Node == p0.From.Node {
						if p.From.Offset == p0.From.Offset {
							if p.From.Scale == p0.From.Scale {
								if p.From.Type == obj.TYPE_FCONST && p.From.Val.(float64) == p0.From.Val.(float64) {
									if p.From.Index == p0.From.Index {
										excise(r)
										goto loop
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// is 'a' a register or constant?
func regconsttyp(a *obj.Addr) bool {
	if regtyp(a) {
		return true
	}
	switch a.Type {
	case obj.TYPE_CONST,
		obj.TYPE_FCONST,
		obj.TYPE_SCONST,
		obj.TYPE_ADDR: // TODO(rsc): Not all TYPE_ADDRs are constants.
		return true
	}

	return false
}

func pushback(r0 *gc.Flow) {
	var r *gc.Flow
	var p *obj.Prog

	var b *gc.Flow
	p0 := (*obj.Prog)(r0.Prog)
	for r = gc.Uniqp(r0); r != nil && gc.Uniqs(r) != nil; r = gc.Uniqp(r) {
		p = r.Prog
		if p.As != obj.ANOP {
			if !regconsttyp(&p.From) || !regtyp(&p.To) {
				break
			}
			if copyu(p, &p0.To, nil) != 0 || copyu(p0, &p.To, nil) != 0 {
				break
			}
		}

		if p.As == obj.ACALL {
			break
		}
		b = r
	}

	if b == nil {
		if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
			fmt.Printf("no pushback: %v\n", r0.Prog)
			if r != nil {
				fmt.Printf("\t%v [%v]\n", r.Prog, gc.Uniqs(r) != nil)
			}
		}

		return
	}

	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		fmt.Printf("pushback\n")
		for r := (*gc.Flow)(b); ; r = r.Link {
			fmt.Printf("\t%v\n", r.Prog)
			if r == r0 {
				break
			}
		}
	}

	t := obj.Prog(*r0.Prog)
	for r = gc.Uniqp(r0); ; r = gc.Uniqp(r) {
		p0 = r.Link.Prog
		p = r.Prog
		p0.As = p.As
		p0.Lineno = p.Lineno
		p0.From = p.From
		p0.To = p.To
		p0.From3 = p.From3
		p0.Reg = p.Reg
		p0.RegTo2 = p.RegTo2
		if r == b {
			break
		}
	}

	p0 = r.Prog
	p0.As = t.As
	p0.Lineno = t.Lineno
	p0.From = t.From
	p0.To = t.To
	p0.From3 = t.From3
	p0.Reg = t.Reg
	p0.RegTo2 = t.RegTo2

	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		fmt.Printf("\tafter\n")
		for r := (*gc.Flow)(b); ; r = r.Link {
			fmt.Printf("\t%v\n", r.Prog)
			if r == r0 {
				break
			}
		}
	}
}

func excise(r *gc.Flow) {
	p := (*obj.Prog)(r.Prog)
	if gc.Debug['P'] != 0 && gc.Debug['v'] != 0 {
		fmt.Printf("%v ===delete===\n", p)
	}
	obj.Nopout(p)
	gc.Ostats.Ndelmov++
}

/*
 * regzer returns 1 if a's value is 0 (a is R0 or $0)
 */
func regzer(a *obj.Addr) int {
	if a.Type == obj.TYPE_CONST || a.Type == obj.TYPE_ADDR {
		if a.Sym == nil && a.Reg == 0 {
			if a.Offset == 0 {
				return 1
			}
		}
	}
	if a.Type == obj.TYPE_REG {
		if a.Reg == s390x.REGZERO {
			return 1
		}
	}
	return 0
}

func regtyp(a *obj.Addr) bool {
	// TODO(rsc): Floating point register exclusions?
	return a.Type == obj.TYPE_REG && s390x.REG_R0 <= a.Reg && a.Reg <= s390x.REG_F15 && a.Reg != s390x.REGZERO
}

/*
 * the idea is to substitute
 * one register for another
 * from one MOV to another
 *	MOV	a, R1
 *	ADD	b, R1	/ no use of R2
 *	MOV	R1, R2
 * would be converted to
 *	MOV	a, R2
 *	ADD	b, R2
 *	MOV	R2, R1
 * hopefully, then the former or latter MOV
 * will be eliminated by copy propagation.
 *
 * r0 (the argument, not the register) is the MOV at the end of the
 * above sequences.  This returns 1 if it modified any instructions.
 */
func subprop(r0 *gc.Flow) bool {
	p := (*obj.Prog)(r0.Prog)
	v1 := (*obj.Addr)(&p.From)
	if !regtyp(v1) {
		return false
	}
	v2 := (*obj.Addr)(&p.To)
	if !regtyp(v2) {
		return false
	}
	for r := gc.Uniqp(r0); r != nil; r = gc.Uniqp(r) {
		if gc.Uniqs(r) == nil {
			break
		}
		p = r.Prog
		if p.As == obj.AVARDEF || p.As == obj.AVARKILL {
			continue
		}
		if p.Info.Flags&gc.Call != 0 {
			return false
		}

		if p.Info.Flags&(gc.RightRead|gc.RightWrite) == gc.RightWrite {
			if p.To.Type == v1.Type {
				if p.To.Reg == v1.Reg {
					copysub(&p.To, v1, v2, 1)
					if gc.Debug['P'] != 0 {
						fmt.Printf("gotit: %v->%v\n%v", gc.Ctxt.Dconv(v1), gc.Ctxt.Dconv(v2), r.Prog)
						if p.From.Type == v2.Type {
							fmt.Printf(" excise")
						}
						fmt.Printf("\n")
					}

					for r = gc.Uniqs(r); r != r0; r = gc.Uniqs(r) {
						p = r.Prog
						copysub(&p.From, v1, v2, 1)
						copysub1(p, v1, v2, 1)
						copysub(&p.To, v1, v2, 1)
						if gc.Debug['P'] != 0 {
							fmt.Printf("%v\n", r.Prog)
						}
					}

					t := int(int(v1.Reg))
					v1.Reg = v2.Reg
					v2.Reg = int16(t)
					if gc.Debug['P'] != 0 {
						fmt.Printf("%v last\n", r.Prog)
					}
					return true
				}
			}
		}

		if copyau(&p.From, v2) || copyau1(p, v2) || copyau(&p.To, v2) {
			break
		}
		if copysub(&p.From, v1, v2, 0) != 0 || copysub1(p, v1, v2, 0) != 0 || copysub(&p.To, v1, v2, 0) != 0 {
			break
		}
	}

	return false
}

/*
 * The idea is to remove redundant copies.
 *	v1->v2	F=0
 *	(use v2	s/v2/v1/)*
 *	set v1	F=1
 *	use v2	return fail (v1->v2 move must remain)
 *	-----------------
 *	v1->v2	F=0
 *	(use v2	s/v2/v1/)*
 *	set v1	F=1
 *	set v2	return success (caller can remove v1->v2 move)
 */
func copyprop(r0 *gc.Flow) bool {
	p := (*obj.Prog)(r0.Prog)
	v1 := (*obj.Addr)(&p.From)
	v2 := (*obj.Addr)(&p.To)
	if copyas(v1, v2) {
		if gc.Debug['P'] != 0 {
			fmt.Printf("eliminating self-move: %v\n", r0.Prog)
		}
		return true
	}

	gactive++
	if gc.Debug['P'] != 0 {
		fmt.Printf("trying to eliminate %v->%v move from:\n%v\n", gc.Ctxt.Dconv(v1), gc.Ctxt.Dconv(v2), r0.Prog)
	}
	return copy1(v1, v2, r0.S1, 0)
}

// copy1 replaces uses of v2 with v1 starting at r and returns 1 if
// all uses were rewritten.
func copy1(v1 *obj.Addr, v2 *obj.Addr, r *gc.Flow, f int) bool {
	if uint32(r.Active) == gactive {
		if gc.Debug['P'] != 0 {
			fmt.Printf("act set; return 1\n")
		}
		return true
	}

	r.Active = int32(gactive)
	if gc.Debug['P'] != 0 {
		fmt.Printf("copy1 replace %v with %v f=%d\n", gc.Ctxt.Dconv(v2), gc.Ctxt.Dconv(v1), f)
	}
	var t int
	var p *obj.Prog
	for ; r != nil; r = r.S1 {
		p = r.Prog
		if gc.Debug['P'] != 0 {
			fmt.Printf("%v", p)
		}
		if f == 0 && gc.Uniqp(r) == nil {
			// Multiple predecessors; conservatively
			// assume v1 was set on other path
			f = 1

			if gc.Debug['P'] != 0 {
				fmt.Printf("; merge; f=%d", f)
			}
		}

		t = copyu(p, v2, nil)
		switch t {
		case 2: /* rar, can't split */
			if gc.Debug['P'] != 0 {
				fmt.Printf("; %v rar; return 0\n", gc.Ctxt.Dconv(v2))
			}
			return false

		case 3: /* set */
			if gc.Debug['P'] != 0 {
				fmt.Printf("; %v set; return 1\n", gc.Ctxt.Dconv(v2))
			}
			return true

		case 1, /* used, substitute */
			4: /* use and set */
			if f != 0 {
				if gc.Debug['P'] == 0 {
					return false
				}
				if t == 4 {
					fmt.Printf("; %v used+set and f=%d; return 0\n", gc.Ctxt.Dconv(v2), f)
				} else {
					fmt.Printf("; %v used and f=%d; return 0\n", gc.Ctxt.Dconv(v2), f)
				}
				return false
			}

			if copyu(p, v2, v1) != 0 {
				if gc.Debug['P'] != 0 {
					fmt.Printf("; sub fail; return 0\n")
				}
				return false
			}

			if gc.Debug['P'] != 0 {
				fmt.Printf("; sub %v->%v\n => %v", gc.Ctxt.Dconv(v2), gc.Ctxt.Dconv(v1), p)
			}
			if t == 4 {
				if gc.Debug['P'] != 0 {
					fmt.Printf("; %v used+set; return 1\n", gc.Ctxt.Dconv(v2))
				}
				return true
			}
		}

		if f == 0 {
			t = copyu(p, v1, nil)
			if f == 0 && (t == 2 || t == 3 || t == 4) {
				f = 1
				if gc.Debug['P'] != 0 {
					fmt.Printf("; %v set and !f; f=%d", gc.Ctxt.Dconv(v1), f)
				}
			}
		}

		if gc.Debug['P'] != 0 {
			fmt.Printf("\n")
		}
		if r.S2 != nil {
			if !copy1(v1, v2, r.S2, f) {
				return false
			}
		}
	}

	return true
}

// If s==nil, copyu returns the set/use of v in p; otherwise, it
// modifies p to replace reads of v with reads of s and returns 0 for
// success or non-zero for failure.
//
// If s==nil, copy returns one of the following values:
// 	1 if v only used
//	2 if v is set and used in one address (read-alter-rewrite;
// 	  can't substitute)
//	3 if v is only set
//	4 if v is set in one address and used in another (so addresses
// 	  can be rewritten independently)
//	0 otherwise (not touched)
func copyu(p *obj.Prog, v *obj.Addr, s *obj.Addr) int {
	if p.From3Type() != obj.TYPE_NONE && p.From3Type() != obj.TYPE_CONST {
		// Currently we never generate a From3 with anything other than a constant in it.
		fmt.Printf("copyu: From3 (%v) not implemented\n", gc.Ctxt.Dconv(p.From3))
	}

	switch p.As {
	default:
		fmt.Printf("copyu: can't find %v\n", obj.Aconv(int(p.As)))
		return 2

	case obj.ANOP, /* read p->from, write p->to */
		s390x.AMOVH,
		s390x.AMOVHZ,
		s390x.AMOVB,
		s390x.AMOVBZ,
		s390x.AMOVW,
		s390x.AMOVWZ,
		s390x.AMOVD,
		s390x.ANEG,
		s390x.AADDME,
		s390x.AADDZE,
		s390x.ASUBME,
		s390x.ASUBZE,
		s390x.AFMOVS,
		s390x.AFMOVD,
		s390x.AFRSP,
		s390x.AFNEG,
		s390x.ALDEBR,
		s390x.ACLFEBR,
		s390x.ACLGEBR,
		s390x.ACLFDBR,
		s390x.ACLGDBR,
		s390x.ACFEBRA,
		s390x.ACGEBRA,
		s390x.ACFDBRA,
		s390x.ACGDBRA,
		s390x.ACELFBR,
		s390x.ACELGBR,
		s390x.ACDLFBR,
		s390x.ACDLGBR,
		s390x.ACEFBRA,
		s390x.ACEGBRA,
		s390x.ACDFBRA,
		s390x.ACDGBRA,
		s390x.AFSQRT:

		if s != nil {
			if copysub(&p.From, v, s, 1) != 0 {
				return 1
			}

			// Update only indirect uses of v in p->to
			if !copyas(&p.To, v) {
				if copysub(&p.To, v, s, 1) != 0 {
					return 1
				}
			}
			return 0
		}

		if copyas(&p.To, v) {
			// Fix up implicit from
			if p.From.Type == obj.TYPE_NONE {
				p.From = p.To
			}
			if copyau(&p.From, v) {
				return 4
			}
			return 3
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau(&p.To, v) {
			// p->to only indirectly uses v
			return 1
		}

		return 0

	// read p->from, read p->reg, rar p->to
	case s390x.ARLWMI:
		if copyas(&p.To, v) {
			return 2
		}
		fallthrough
	// read p->from, read p->reg, write p->to
	case s390x.AADD,
		s390x.AADDC,
		s390x.AADDE,
		s390x.ASUB,
		s390x.ASLW,
		s390x.ASRW,
		s390x.ASRAW,
		s390x.ASLD,
		s390x.ASRD,
		s390x.ASRAD,
		s390x.AOR,
		s390x.AORN,
		s390x.AAND,
		s390x.AANDN,
		s390x.ANAND,
		s390x.ANOR,
		s390x.AXOR,
		s390x.AMULLW,
		s390x.AMULLD,
		s390x.ADIVW,
		s390x.ADIVD,
		s390x.ADIVWU,
		s390x.ADIVDU,
		s390x.AREM,
		s390x.AREMU,
		s390x.AREMD,
		s390x.AREMDU,
		s390x.ARLWNM,
		s390x.AFADDS,
		s390x.AFADD,
		s390x.AFSUBS,
		s390x.AFSUB,
		s390x.AFMULS,
		s390x.AFMUL,
		s390x.AFDIVS,
		s390x.AFDIV:
		if s != nil {
			if copysub(&p.From, v, s, 1) != 0 {
				return 1
			}
			if copysub1(p, v, s, 1) != 0 {
				return 1
			}

			// Update only indirect uses of v in p->to
			if !copyas(&p.To, v) {
				if copysub(&p.To, v, s, 1) != 0 {
					return 1
				}
			}
			return 0
		}

		if copyas(&p.To, v) {
			if p.Reg == 0 {
				// Fix up implicit reg (e.g., ADD
				// R3,R4 -> ADD R3,R4,R4) so we can
				// update reg and to separately.
				p.Reg = p.To.Reg
			}

			if copyau(&p.From, v) {
				return 4
			}
			if copyau1(p, v) {
				return 4
			}
			return 3
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau1(p, v) {
			return 1
		}
		if copyau(&p.To, v) {
			return 1
		}
		return 0

	case s390x.ABEQ,
		s390x.ABGT,
		s390x.ABGE,
		s390x.ABLT,
		s390x.ABLE,
		s390x.ABNE,
		s390x.ABVC,
		s390x.ABVS:
		return 0

	case obj.ACHECKNIL, /* read p->from */
		s390x.ACMP, /* read p->from, read p->to */
		s390x.ACMPU,
		s390x.ACMPW,
		s390x.ACMPWU,
		s390x.AFCMPO,
		s390x.AFCMPU,
		s390x.ACEBR,
		s390x.AMVC,
		s390x.ACLC,
		s390x.AXC,
		s390x.AOC,
		s390x.ANC:
		if s != nil {
			if copysub(&p.From, v, s, 1) != 0 {
				return 1
			}
			return copysub(&p.To, v, s, 1)
		}

		if copyau(&p.From, v) {
			return 1
		}
		if copyau(&p.To, v) {
			return 1
		}
		return 0

	// go never generates a branch to a GPR
	// read p->to
	case s390x.ABR:
		if s != nil {
			if copysub(&p.To, v, s, 1) != 0 {
				return 1
			}
			return 0
		}

		if copyau(&p.To, v) {
			return 1
		}
		return 0

	case obj.ARET:
		if s != nil {
			return 0
		}

		// All registers die at this point, so claim
		// everything is set (and not used).
		return 3

	case s390x.ABL:
		if v.Type == obj.TYPE_REG {
			// TODO(rsc): REG_R0 and REG_F0 used to be
			// (when register numbers started at 0) exregoffset and exfregoffset,
			// which are unset entirely.
			// It's strange that this handles R0 and F0 differently from the other
			// registers. Possible failure to optimize?
			if s390x.REG_R0 < v.Reg && v.Reg <= s390x.REGEXT {
				return 2
			}
			if v.Reg == s390x.REGARG {
				return 2
			}
			if s390x.REG_F0 < v.Reg && v.Reg <= s390x.FREGEXT {
				return 2
			}
		}

		if p.From.Type == obj.TYPE_REG && v.Type == obj.TYPE_REG && p.From.Reg == v.Reg {
			return 2
		}

		if s != nil {
			if copysub(&p.To, v, s, 1) != 0 {
				return 1
			}
			return 0
		}

		if copyau(&p.To, v) {
			return 4
		}
		return 3

	case obj.ATEXT:
		if v.Type == obj.TYPE_REG {
			if v.Reg == s390x.REGARG {
				return 3
			}
		}
		return 0

	case obj.APCDATA,
		obj.AFUNCDATA,
		obj.AVARDEF,
		obj.AVARKILL:
		return 0
	}
}

// copyas returns 1 if a and v address the same register.
//
// If a is the from operand, this means this operation reads the
// register in v.  If a is the to operand, this means this operation
// writes the register in v.
func copyas(a *obj.Addr, v *obj.Addr) bool {
	if regtyp(v) {
		if a.Type == v.Type {
			if a.Reg == v.Reg {
				return true
			}
		}
	}
	return false
}

// copyau returns 1 if a either directly or indirectly addresses the
// same register as v.
//
// If a is the from operand, this means this operation reads the
// register in v.  If a is the to operand, this means the operation
// either reads or writes the register in v (if !copyas(a, v), then
// the operation reads the register in v).
func copyau(a *obj.Addr, v *obj.Addr) bool {
	if copyas(a, v) {
		return true
	}
	if v.Type == obj.TYPE_REG {
		if a.Type == obj.TYPE_MEM || (a.Type == obj.TYPE_ADDR && a.Reg != 0) {
			if v.Reg == a.Reg {
				return true
			}
		}
	}
	return false
}

// copyau1 returns 1 if p->reg references the same register as v and v
// is a direct reference.
func copyau1(p *obj.Prog, v *obj.Addr) bool {
	if regtyp(v) && v.Reg != 0 {
		if p.Reg == v.Reg {
			return true
		}
	}
	return false
}

// copysub replaces v with s in a if f!=0 or indicates it if could if f==0.
// Returns 1 on failure to substitute (it always succeeds on s390x).
func copysub(a *obj.Addr, v *obj.Addr, s *obj.Addr, f int) int {
	if f != 0 {
		if copyau(a, v) {
			a.Reg = s.Reg
		}
	}
	return 0
}

// copysub1 replaces v with s in p1->reg if f!=0 or indicates if it could if f==0.
// Returns 1 on failure to substitute (it always succeeds on s390x).
func copysub1(p1 *obj.Prog, v *obj.Addr, s *obj.Addr, f int) int {
	if f != 0 {
		if copyau1(p1, v) {
			p1.Reg = s.Reg
		}
	}
	return 0
}

func sameaddr(a *obj.Addr, v *obj.Addr) bool {
	if a.Type != v.Type {
		return false
	}
	if regtyp(v) && a.Reg == v.Reg {
		return true
	}
	if v.Type == obj.NAME_AUTO || v.Type == obj.NAME_PARAM {
		if v.Offset == a.Offset {
			return true
		}
	}
	return false
}

func smallindir(a *obj.Addr, reg *obj.Addr) bool {
	return reg.Type == obj.TYPE_REG && a.Type == obj.TYPE_MEM && a.Reg == reg.Reg && 0 <= a.Offset && a.Offset < 4096
}

func stackaddr(a *obj.Addr) bool {
	return a.Type == obj.TYPE_REG && a.Reg == s390x.REGSP
}

func trymergeopmv(r *gc.Flow) bool {
	p := r.Prog
	reg := p.From.Reg
	r2 := gc.Uniqs(r)

	for ; r2 != nil; r2 = gc.Uniqs(r2) {
		p2 := r2.Prog
		switch p2.As {
		case obj.ANOP:
			continue

		case s390x.AEXRL,
			s390x.ASYSCALL,
			s390x.ABR,
			s390x.ABC,
			s390x.ABEQ,
			s390x.ABGE,
			s390x.ABGT,
			s390x.ABLE,
			s390x.ABLT,
			s390x.ABNE,
			s390x.ACMPBEQ,
			s390x.ACMPBGE,
			s390x.ACMPBGT,
			s390x.ACMPBLE,
			s390x.ACMPBLT,
			s390x.ACMPBNE:
			return false

		case s390x.ACMP,
			s390x.ACMPU,
			s390x.ACMPW,
			s390x.ACMPWU:
			if p2.From.Type == obj.TYPE_REG && p2.From.Reg == reg {
				return false
			}
			if p2.To.Type == obj.TYPE_REG && p2.To.Reg == reg {
				//different from other instructions, To.Reg is a source register in CMP
				return false
			}
			continue

		case s390x.AMOVD,
			s390x.AMOVW, s390x.AMOVWZ,
			s390x.AMOVH, s390x.AMOVHZ,
			s390x.AMOVB, s390x.AMOVBZ:
			if p2.From.Type == obj.TYPE_REG && p2.From.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.From.Type == obj.TYPE_ADDR && p2.From.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.To.Type == obj.TYPE_ADDR && p2.To.Reg == reg {
				//For store operations
				//also use;  can't change
				return false
			}
			if p2.To.Type == obj.TYPE_REG && p2.To.Reg == reg {
				return true
			}
			continue

		case s390x.AMVC, s390x.ACLC, s390x.AXC, s390x.AOC, s390x.ANC:
			if p2.From.Type == obj.TYPE_MEM && p2.From.Reg == reg {
				return false
			}
			if p2.To.Type == obj.TYPE_MEM && p2.To.Reg == reg {
				return false
			}
			continue

		default:
			if p2.From.Type == obj.TYPE_REG && p2.From.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.From.Type == obj.TYPE_ADDR && p2.From.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.Reg != 0 && p2.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.From3 != nil && p2.From3.Type == obj.TYPE_REG && p2.From3.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.From3 != nil && p2.From3.Type == obj.TYPE_ADDR && p2.From3.Reg == reg {
				//use;  can't change
				return false
			}
			if p2.To.Type == obj.TYPE_ADDR && p2.To.Reg == reg {
				//For store operations
				//also use;  can't change
				return false
			}
			if p2.To.Type == obj.TYPE_REG && p2.To.Reg == reg {
				if p2.Reg == 0 {
					//p2.To is also used as source in 2 operands instruction
					return false
				} else {
					//def;  can change
					return true
				}
			}
			continue
		}
	}
	return false
}
