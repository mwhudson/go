// Inferno utils/5l/asm.c
// http://code.google.com/p/inferno-os/source/browse/utils/5l/asm.c
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

package arm64

import (
	"cmd/internal/obj"
	"cmd/link/internal/ld"
	"debug/elf"
	"fmt"
	"log"
)

func gentext() {
	if !ld.DynlinkingGo() {
		return
	}
	addmoduledata := ld.Linklookup(ld.Ctxt, "runtime.addmoduledata", 0)
	if addmoduledata.Type == obj.STEXT {
		// we're linking a module containing the runtime -> no need for
		// an init function
		return
	}
	addmoduledata.Reachable = true
	initfunc := ld.Linklookup(ld.Ctxt, "go.link.addmoduledata", 0)
	initfunc.Type = obj.STEXT
	initfunc.Local = true
	initfunc.Reachable = true
	o := func(op uint32) {
		ld.Adduint32(ld.Ctxt, initfunc, op)
	}
	// 0000000000000000 <local.dso_init>:
	// 0:	90000000 	adrp	x0, 0 <runtime.firstmoduledata>
	// 	0: R_AARCH64_ADR_PREL_PG_HI21	runtime.firstmoduledata
	o(0x90000000)
	rel := ld.Addrel(initfunc)
	rel.Off = 0
	rel.Siz = 4
	rel.Sym = ld.Linklookup(ld.Ctxt, "runtime.firstmoduledata", 0)
	rel.Type = obj.R_AARCH64_ADR_PREL_PG_HI21
	// 4:	91000000 	add	x0, x0, #0x0
	// 	4: R_AARCH64_ADD_ABS_LO12_NC	runtime.firstmoduledata
	o(0x91000000)
	rel = ld.Addrel(initfunc)
	rel.Off = 4
	rel.Siz = 4
	rel.Sym = ld.Linklookup(ld.Ctxt, "runtime.firstmoduledata", 0)
	rel.Type = obj.R_AARCH64_ADD_ABS_LO12_NC

	// 8:	14000000 	bl	0 <runtime.addmoduledata>
	// 	8: R_AARCH64_CALL26	runtime.addmoduledata
	o(0x14000000)
	rel = ld.Addrel(initfunc)
	rel.Off = 8
	rel.Siz = 4
	rel.Sym = ld.Linklookup(ld.Ctxt, "runtime.addmoduledata", 0)
	rel.Type = obj.R_AARCH64_CALL26 // Really should be R_AARCH64_JUMP26 but doesn't seem to make any difference

	if ld.Ctxt.Etextp != nil {
		ld.Ctxt.Etextp.Next = initfunc
	} else {
		ld.Ctxt.Textp = initfunc
	}
	ld.Ctxt.Etextp = initfunc
	initarray_entry := ld.Linklookup(ld.Ctxt, "go.link.addmoduledatainit", 0)
	initarray_entry.Reachable = true
	initarray_entry.Local = true
	initarray_entry.Type = obj.SINITARR
	ld.Addaddr(ld.Ctxt, initarray_entry, initfunc)
}

func adddynrela(rel *ld.LSym, s *ld.LSym, r *ld.Reloc) {
	log.Fatalf("adddynrela not implemented")
}

func adddynrel(s *ld.LSym, r *ld.Reloc) {
	log.Fatalf("adddynrel not implemented")
}

func elfreloc1(r *ld.Reloc, sectoff int64) int {
	ld.Thearch.Vput(uint64(sectoff))

	elfsym := r.Xsym.Elfsym
	switch r.Type {
	default:
		return -1

	case obj.R_ADDR:
		switch r.Siz {
		case 4:
			ld.Thearch.Vput(ld.R_AARCH64_ABS32 | uint64(elfsym)<<32)
		case 8:
			ld.Thearch.Vput(ld.R_AARCH64_ABS64 | uint64(elfsym)<<32)
		default:
			return -1
		}

	case obj.R_AARCH64_ADR_GOT_PAGE:
		if r.Xsym.Local {
			ld.Ctxt.Diag("wtf %s", r.Xsym.Name)
		}
		ld.Thearch.Vput(uint64(elf.R_AARCH64_ADR_GOT_PAGE) | uint64(elfsym)<<32)

	case obj.R_AARCH64_LD64_GOT_LO12_NC:
		ld.Thearch.Vput(uint64(elf.R_AARCH64_LD64_GOT_LO12_NC) | uint64(elfsym)<<32)

	case obj.R_ARM64_TLSLE_MOVW_TPREL_G0:
		ld.Thearch.Vput(uint64(elf.R_AARCH64_TLSLE_MOVW_TPREL_G0) | uint64(elfsym)<<32)

	case obj.R_AARCH64_ADR_PREL_PG_HI21:
		ld.Thearch.Vput(ld.R_AARCH64_ADR_PREL_PG_HI21 | uint64(elfsym)<<32)

	case obj.R_AARCH64_ADD_ABS_LO12_NC:
		ld.Thearch.Vput(ld.R_AARCH64_ADD_ABS_LO12_NC | uint64(elfsym)<<32)

	case obj.R_AARCH64_TLSIE_ADR_GOTTPREL_PAGE21:
		ld.Thearch.Vput(uint64(elf.R_AARCH64_TLSIE_ADR_GOTTPREL_PAGE21) | uint64(elfsym)<<32)

	case obj.R_AARCH64_TLSIE_LD64_GOTTPREL_LO12_NC:
		ld.Thearch.Vput(uint64(elf.R_AARCH64_TLSIE_LD64_GOTTPREL_LO12_NC) | uint64(elfsym)<<32)

	case obj.R_AARCH64_CALL26:
		if r.Siz != 4 {
			return -1
		}
		ld.Thearch.Vput(ld.R_AARCH64_CALL26 | uint64(elfsym)<<32)

	}
	ld.Thearch.Vput(uint64(r.Xadd))

	return 0
}

func elfsetupplt() {
	// TODO(aram)
	return
}

func machoreloc1(r *ld.Reloc, sectoff int64) int {
	var v uint32

	rs := r.Xsym

	// ld64 has a bug handling MACHO_ARM64_RELOC_UNSIGNED with !extern relocation.
	// see cmd/internal/ld/data.go for details. The workarond is that don't use !extern
	// UNSIGNED relocation at all.
	if rs.Type == obj.SHOSTOBJ || r.Type == obj.R_AARCH64_CALL26 || r.Type == obj.R_AARCH64_ADR_PREL_PG_HI21 || r.Type == obj.R_AARCH64_ADD_ABS_LO12_NC || r.Type == obj.R_ADDR {
		if rs.Dynid < 0 {
			ld.Diag("reloc %d to non-macho symbol %s type=%d", r.Type, rs.Name, rs.Type)
			return -1
		}

		v = uint32(rs.Dynid)
		v |= 1 << 27 // external relocation
	} else {
		v = uint32(rs.Sect.Extnum)
		if v == 0 {
			ld.Diag("reloc %d to symbol %s in non-macho section %s type=%d", r.Type, rs.Name, rs.Sect.Name, rs.Type)
			return -1
		}
	}

	switch r.Type {
	default:
		return -1

	case obj.R_ADDR:
		v |= ld.MACHO_ARM64_RELOC_UNSIGNED << 28

	case obj.R_AARCH64_CALL26:
		if r.Xadd != 0 {
			ld.Diag("ld64 doesn't allow BR26 reloc with non-zero addend: %s+%d", rs.Name, r.Xadd)
		}

		v |= 1 << 24 // pc-relative bit
		v |= ld.MACHO_ARM64_RELOC_BRANCH26 << 28

	case obj.R_AARCH64_ADR_PREL_PG_HI21:
		r.Siz = 4
		// MACHO_ARM64_RELOC_PAGEOFF12
		if r.Xadd != 0 {
			ld.Thearch.Lput(uint32(sectoff))
			ld.Thearch.Lput((ld.MACHO_ARM64_RELOC_ADDEND << 28) | (2 << 25) | uint32(r.Xadd&0xffffff))
		}
		v |= 1 << 24 // pc-relative bit
		v |= (ld.MACHO_ARM64_RELOC_PAGE21 << 28) | (2 << 25)

	case obj.R_AARCH64_ADD_ABS_LO12_NC:
		r.Siz = 4
		// Two relocation entries: MACHO_ARM64_RELOC_PAGEOFF12 MACHO_ARM64_RELOC_PAGE21
		if r.Xadd != 0 {
			ld.Thearch.Lput(uint32(sectoff))
			ld.Thearch.Lput((ld.MACHO_ARM64_RELOC_ADDEND << 28) | (2 << 25) | uint32(r.Xadd&0xffffff))
		}
		v |= 1 << 24 // pc-relative bit
		v |= (ld.MACHO_ARM64_RELOC_PAGEOFF12 << 28) | (2 << 25)
	}

	switch r.Siz {
	default:
		return -1

	case 1:
		v |= 0 << 25

	case 2:
		v |= 1 << 25

	case 4:
		v |= 2 << 25

	case 8:
		v |= 3 << 25
	}

	ld.Thearch.Lput(uint32(sectoff))
	ld.Thearch.Lput(v)
	return 0
}

func archreloc(r *ld.Reloc, s *ld.LSym, val *int64) int {
	if ld.Linkmode == ld.LinkExternal {
		switch r.Type {
		default:
			return -1

		case obj.R_ARM64_TLSLE_MOVW_TPREL_G0,
			obj.R_AARCH64_TLSIE_ADR_GOTTPREL_PAGE21,
			obj.R_AARCH64_TLSIE_LD64_GOTTPREL_LO12_NC:
			r.Done = 0
			// check Outer is nil, Type is TLSBSS?
			r.Xadd = r.Add
			r.Xsym = r.Sym
			return 0

		case obj.R_AARCH64_ADR_PREL_PG_HI21,
			obj.R_AARCH64_ADD_ABS_LO12_NC,
			obj.R_AARCH64_ADR_GOT_PAGE,
			obj.R_AARCH64_LD64_GOT_LO12_NC:
			r.Done = 0

			// set up addend for eventual relocation via outer symbol.
			rs := r.Sym
			r.Xadd = r.Add
			for rs.Outer != nil {
				r.Xadd += ld.Symaddr(rs) - ld.Symaddr(rs.Outer)
				rs = rs.Outer
			}

			if rs.Type != obj.SHOSTOBJ && rs.Type != obj.SDYNIMPORT && rs.Sect == nil {
				ld.Diag("missing section for %s ++", rs.Name)
			}
			r.Xsym = rs

			return 0

		case obj.R_AARCH64_CALL26:
			r.Done = 0
			r.Xsym = r.Sym
			r.Xadd = r.Add
			return 0
		}
	}

	switch r.Type {
	case obj.R_CONST:
		*val = r.Add
		return 0

	case obj.R_GOTOFF:
		*val = ld.Symaddr(r.Sym) + r.Add - ld.Symaddr(ld.Linklookup(ld.Ctxt, ".got", 0))
		return 0

	case obj.R_ARM64_TLSLE_MOVW_TPREL_G0:
		r.Done = 0
		if ld.HEADTYPE != obj.Hlinux {
			ld.Diag("TLS reloc on unsupported OS %s", ld.Headstr(int(ld.HEADTYPE)))
		}
		// The TCB is 16 bytes on linux in lp64
		v := r.Sym.Value + 16
		if v < 0 || v >= 32678 {
			ld.Diag("TLS offset out of range %d", v)
		}
		*val |= v << 5
		return 0

	case obj.R_AARCH64_ADR_PREL_PG_HI21:
		t := ((ld.Symaddr(r.Sym) + r.Add) &^ 0xfff) - ((s.Value + int64(r.Off)) &^ 0xfff)
		if t >= 1<<32 || t < -1<<32 {
			ld.Diag("program too large, address relocation distance = %d", t)
		}
		*val |= (((t >> 12) & 3) << 29) | (((t >> 12 >> 2) & 0x7ffff) << 5)
		return 0

	case obj.R_AARCH64_ADD_ABS_LO12_NC:
		t := (ld.Symaddr(r.Sym) + r.Add)
		*val |= (t & 0xfff) << 10
		return 0

	case obj.R_AARCH64_CALL26:
		t := (ld.Symaddr(r.Sym) + r.Add) - (s.Value + int64(r.Off))
		if t >= 1<<27 || t < -1<<27 {
			ld.Diag("program too large, call relocation distance = %d", t)
		}
		*val |= (t >> 2) & 0x03ffffff
		return 0
	}

	return -1
}

func archrelocvariant(r *ld.Reloc, s *ld.LSym, t int64) int64 {
	log.Fatalf("unexpected relocation variant")
	return -1
}

func asmb() {
	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f asmb\n", obj.Cputime())
	}
	ld.Bso.Flush()

	if ld.Iself {
		ld.Asmbelfsetup()
	}

	sect := ld.Segtext.Sect
	ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
	ld.Codeblk(int64(sect.Vaddr), int64(sect.Length))
	for sect = sect.Next; sect != nil; sect = sect.Next {
		ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
		ld.Datblk(int64(sect.Vaddr), int64(sect.Length))
	}

	if ld.Segrodata.Filelen > 0 {
		if ld.Debug['v'] != 0 {
			fmt.Fprintf(&ld.Bso, "%5.2f rodatblk\n", obj.Cputime())
		}
		ld.Bso.Flush()

		ld.Cseek(int64(ld.Segrodata.Fileoff))
		ld.Datblk(int64(ld.Segrodata.Vaddr), int64(ld.Segrodata.Filelen))
	}

	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f datblk\n", obj.Cputime())
	}
	ld.Bso.Flush()

	ld.Cseek(int64(ld.Segdata.Fileoff))
	ld.Datblk(int64(ld.Segdata.Vaddr), int64(ld.Segdata.Filelen))

	machlink := uint32(0)
	if ld.HEADTYPE == obj.Hdarwin {
		if ld.Debug['v'] != 0 {
			fmt.Fprintf(&ld.Bso, "%5.2f dwarf\n", obj.Cputime())
		}

		dwarfoff := uint32(ld.Rnd(int64(uint64(ld.HEADR)+ld.Segtext.Length), int64(ld.INITRND)) + ld.Rnd(int64(ld.Segdata.Filelen), int64(ld.INITRND)))
		ld.Cseek(int64(dwarfoff))

		ld.Segdwarf.Fileoff = uint64(ld.Cpos())
		ld.Dwarfemitdebugsections()
		ld.Segdwarf.Filelen = uint64(ld.Cpos()) - ld.Segdwarf.Fileoff

		machlink = uint32(ld.Domacholink())
	}

	/* output symbol table */
	ld.Symsize = 0

	ld.Lcsize = 0
	symo := uint32(0)
	if ld.Debug['s'] == 0 {
		// TODO: rationalize
		if ld.Debug['v'] != 0 {
			fmt.Fprintf(&ld.Bso, "%5.2f sym\n", obj.Cputime())
		}
		ld.Bso.Flush()
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				symo = uint32(ld.Segdata.Fileoff + ld.Segdata.Filelen)
				symo = uint32(ld.Rnd(int64(symo), int64(ld.INITRND)))
			}

		case obj.Hplan9:
			symo = uint32(ld.Segdata.Fileoff + ld.Segdata.Filelen)

		case obj.Hdarwin:
			symo = uint32(ld.Segdwarf.Fileoff + uint64(ld.Rnd(int64(ld.Segdwarf.Filelen), int64(ld.INITRND))) + uint64(machlink))
		}

		ld.Cseek(int64(symo))
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				if ld.Debug['v'] != 0 {
					fmt.Fprintf(&ld.Bso, "%5.2f elfsym\n", obj.Cputime())
				}
				ld.Asmelfsym()
				ld.Cflush()
				ld.Cwrite(ld.Elfstrdat)

				if ld.Debug['v'] != 0 {
					fmt.Fprintf(&ld.Bso, "%5.2f dwarf\n", obj.Cputime())
				}
				ld.Dwarfemitdebugsections()

				if ld.Linkmode == ld.LinkExternal {
					ld.Elfemitreloc()
				}
			}

		case obj.Hplan9:
			ld.Asmplan9sym()
			ld.Cflush()

			sym := ld.Linklookup(ld.Ctxt, "pclntab", 0)
			if sym != nil {
				ld.Lcsize = int32(len(sym.P))
				for i := 0; int32(i) < ld.Lcsize; i++ {
					ld.Cput(uint8(sym.P[i]))
				}

				ld.Cflush()
			}

		case obj.Hdarwin:
			if ld.Linkmode == ld.LinkExternal {
				ld.Machoemitreloc()
			}
		}
	}

	ld.Ctxt.Cursym = nil
	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f header\n", obj.Cputime())
	}
	ld.Bso.Flush()
	ld.Cseek(0)
	switch ld.HEADTYPE {
	default:
	case obj.Hplan9: /* plan 9 */
		ld.Thearch.Lput(0x647)                      /* magic */
		ld.Thearch.Lput(uint32(ld.Segtext.Filelen)) /* sizes */
		ld.Thearch.Lput(uint32(ld.Segdata.Filelen))
		ld.Thearch.Lput(uint32(ld.Segdata.Length - ld.Segdata.Filelen))
		ld.Thearch.Lput(uint32(ld.Symsize))      /* nsyms */
		ld.Thearch.Lput(uint32(ld.Entryvalue())) /* va of entry */
		ld.Thearch.Lput(0)
		ld.Thearch.Lput(uint32(ld.Lcsize))

	case obj.Hlinux,
		obj.Hfreebsd,
		obj.Hnetbsd,
		obj.Hopenbsd,
		obj.Hnacl:
		ld.Asmbelf(int64(symo))

	case obj.Hdarwin:
		ld.Asmbmacho()
	}

	ld.Cflush()
	if ld.Debug['c'] != 0 {
		fmt.Printf("textsize=%d\n", ld.Segtext.Filelen)
		fmt.Printf("datsize=%d\n", ld.Segdata.Filelen)
		fmt.Printf("bsssize=%d\n", ld.Segdata.Length-ld.Segdata.Filelen)
		fmt.Printf("symsize=%d\n", ld.Symsize)
		fmt.Printf("lcsize=%d\n", ld.Lcsize)
		fmt.Printf("total=%d\n", ld.Segtext.Filelen+ld.Segdata.Length+uint64(ld.Symsize)+uint64(ld.Lcsize))
	}
}
