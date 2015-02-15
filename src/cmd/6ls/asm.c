// Inferno utils/6l/asm.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/asm.c
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

// Writing object files.

#include	"l.h"
#include	"lib.h"
#include	"elf.h"
#include	"dwarf.h"

#define PADDR(a)	((uint32)(a) & ~0x80000000)

char linuxdynld[] = "/lib64/ld-linux-x86-64.so.2";
char freebsddynld[] = "/libexec/ld-elf.so.1";
char openbsddynld[] = "/usr/libexec/ld.so";
char netbsddynld[] = "/libexec/ld.elf_so";
char dragonflydynld[] = "/usr/libexec/ld-elf.so.2";
char solarisdynld[] = "/lib/amd64/ld.so.1";

char	zeroes[32];

static int
needlib(char *name)
{
	char *p;
	LSym *s;

	if(*name == '\0')
		return 0;

	/* reuse hash code in symbol table */
	p = smprint(".elfload.%s", name);
	s = linklookup(ctxt, p, 0);
	free(p);
	if(s->type == 0) {
		s->type = 100;	// avoid SDATA, etc.
		return 1;
	}
	return 0;
}

int nelfsym = 1;

static void addpltsym(LSym*);
static void addgotsym(LSym*);

void
gentext(void)
{
}

void
adddynrela(LSym *rela, LSym *s, Reloc *r)
{
	addaddrplus(ctxt, rela, s, r->off);
	adduint64(ctxt, rela, R_X86_64_RELATIVE);
	addaddrplus(ctxt, rela, r->sym, r->add); // Addend
}

void
adddynrel(LSym *s, Reloc *r)
{
	LSym *targ, *rela;
	
	targ = r->sym;
	ctxt->cursym = s;

	switch(r->type) {
	default:
		if(r->type >= 256) {
			diag("unexpected relocation type %d", r->type);
			return;
		}
		break;

	// Handle relocations found in ELF object files.
	case 256 + R_X86_64_PC32:
		if(targ->type == SDYNIMPORT)
			diag("unexpected R_X86_64_PC32 relocation for dynamic symbol %s", targ->name);
		if(targ->type == 0 || targ->type == SXREF)
			diag("unknown symbol %s in pcrel", targ->name);
		r->type = R_PCREL;
		r->add += 4;
		return;
	
	case 256 + R_X86_64_PLT32:
		r->type = R_PCREL;
		r->add += 4;
		if(targ->type == SDYNIMPORT) {
			addpltsym(targ);
			r->sym = linklookup(ctxt, ".plt", 0);
			r->add += targ->plt;
		}
		return;
	
	case 256 + R_X86_64_GOTPCREL:
		if(targ->type != SDYNIMPORT) {
			// have symbol
			if(r->off >= 2 && s->p[r->off-2] == 0x8b) {
				// turn MOVQ of GOT entry into LEAQ of symbol itself
				s->p[r->off-2] = 0x8d;
				r->type = R_PCREL;
				r->add += 4;
				return;
			}
			// fall back to using GOT and hope for the best (CMOV*)
			// TODO: just needs relocation, no need to put in .dynsym
		}
		addgotsym(targ);
		r->type = R_PCREL;
		r->sym = linklookup(ctxt, ".got", 0);
		r->add += 4;
		r->add += targ->got;
		return;
	
	case 256 + R_X86_64_64:
		if(targ->type == SDYNIMPORT)
			diag("unexpected R_X86_64_64 relocation for dynamic symbol %s", targ->name);
		r->type = R_ADDR;
		return;
	
	}
	
	// Handle references to ELF symbols from our own object files.
	if(targ->type != SDYNIMPORT)
		return;

	switch(r->type) {
	case R_CALL:
	case R_PCREL:
		addpltsym(targ);
		r->sym = linklookup(ctxt, ".plt", 0);
		r->add = targ->plt;
		return;
	
	case R_ADDR:
		if(s->type == STEXT && iself) {
			// The code is asking for the address of an external
			// function.  We provide it with the address of the
			// correspondent GOT symbol.
			addgotsym(targ);
			r->sym = linklookup(ctxt, ".got", 0);
			r->add += targ->got;
			return;
		}
		if(s->type != SDATA)
			break;
		if(iself) {
			adddynsym(ctxt, targ);
			rela = linklookup(ctxt, ".rela", 0);
			addaddrplus(ctxt, rela, s, r->off);
			if(r->siz == 8)
				adduint64(ctxt, rela, ELF64_R_INFO(targ->dynid, R_X86_64_64));
			else
				adduint64(ctxt, rela, ELF64_R_INFO(targ->dynid, R_X86_64_32));
			adduint64(ctxt, rela, r->add);
			r->type = 256;	// ignore during relocsym
			return;
		}
		break;
	}
	
	ctxt->cursym = s;
	diag("unsupported relocation for dynamic symbol %s (type=%d stype=%d)", targ->name, r->type, targ->type);
}

int
elfreloc1(Reloc *r, vlong sectoff)
{
	int32 elfsym;

	VPUT(sectoff);

	elfsym = r->xsym->elfsym;
	switch(r->type) {
	default:
		return -1;

	case R_ADDR:
		if(r->siz == 4)
			VPUT(R_X86_64_32 | (uint64)elfsym<<32);
		else if(r->siz == 8)
			VPUT(R_X86_64_64 | (uint64)elfsym<<32);
		else
			return -1;
		break;

	case R_TLS_LE:
		if(r->siz == 4)
			VPUT(R_X86_64_TPOFF32 | (uint64)elfsym<<32);
		else
			return -1;
		break;

	case R_TLS_IE:
		if(r->siz == 4)
			VPUT(R_X86_64_GOTTPOFF | (uint64)elfsym<<32);
		else
			return -1;
		break;
		
	case R_CALL:
		if(r->siz == 4) {
			if(r->xsym->type == SDYNIMPORT || (r->sym->type == 0 && ctxt->flag_dso))
				// Er, not sure this is right.
				VPUT(R_X86_64_GOTPCREL | (uint64)elfsym<<32);
			else
				VPUT(R_X86_64_PC32 | (uint64)elfsym<<32);
		} else
			return -1;
		break;

	case R_PCREL:
		if(r->siz == 4) {
			if(r->sym->type == 0 && ctxt->flag_dso)
				VPUT(R_X86_64_GOTPCREL | (uint64)elfsym<<32);
			else
				VPUT(R_X86_64_PC32 | (uint64)elfsym<<32);
		} else
			return -1;
		break;

	case R_TLS:
		if(r->siz == 4) {
			if(flag_shared)
				VPUT(R_X86_64_GOTTPOFF | (uint64)elfsym<<32);
			else
				VPUT(R_X86_64_TPOFF32 | (uint64)elfsym<<32);
		} else
			return -1;
		break;		
	}

	VPUT(r->xadd);
	return 0;
}

int
archreloc(Reloc *r, LSym *s, vlong *val)
{
	USED(r);
	USED(s);
	USED(val);
	return -1;
}

vlong
archrelocvariant(Reloc *r, LSym *s, vlong t)
{
	USED(r);
	USED(s);
	sysfatal("unexpected relocation variant");
	return t;
}

void
elfsetupplt(void)
{
	LSym *plt, *got;

	plt = linklookup(ctxt, ".plt", 0);
	got = linklookup(ctxt, ".got.plt", 0);
	if(plt->size == 0) {
		// pushq got+8(IP)
		adduint8(ctxt, plt, 0xff);
		adduint8(ctxt, plt, 0x35);
		addpcrelplus(ctxt, plt, got, 8);
		
		// jmpq got+16(IP)
		adduint8(ctxt, plt, 0xff);
		adduint8(ctxt, plt, 0x25);
		addpcrelplus(ctxt, plt, got, 16);
		
		// nopl 0(AX)
		adduint32(ctxt, plt, 0x00401f0f);
		
		// assume got->size == 0 too
		addaddrplus(ctxt, got, linklookup(ctxt, ".dynamic", 0), 0);
		adduint64(ctxt, got, 0);
		adduint64(ctxt, got, 0);
	}
}

static void
addpltsym(LSym *s)
{
	if(s->plt >= 0)
		return;
	
	adddynsym(ctxt, s);
	
	if(iself) {
		LSym *plt, *got, *rela;

		plt = linklookup(ctxt, ".plt", 0);
		got = linklookup(ctxt, ".got.plt", 0);
		rela = linklookup(ctxt, ".rela.plt", 0);
		if(plt->size == 0)
			elfsetupplt();
		
		// jmpq *got+size(IP)
		adduint8(ctxt, plt, 0xff);
		adduint8(ctxt, plt, 0x25);
		addpcrelplus(ctxt, plt, got, got->size);
	
		// add to got: pointer to current pos in plt
		addaddrplus(ctxt, got, plt, plt->size);
		
		// pushq $x
		adduint8(ctxt, plt, 0x68);
		adduint32(ctxt, plt, (got->size-24-8)/8);
		
		// jmpq .plt
		adduint8(ctxt, plt, 0xe9);
		adduint32(ctxt, plt, -(plt->size+4));
		
		// rela
		addaddrplus(ctxt, rela, got, got->size-8);
		adduint64(ctxt, rela, ELF64_R_INFO(s->dynid, R_X86_64_JMP_SLOT));
		adduint64(ctxt, rela, 0);
		
		s->plt = plt->size - 16;
	} else {
		diag("addpltsym: unsupported binary format");
	}
}

static void
addgotsym(LSym *s)
{
	LSym *got, *rela;

	if(s->got >= 0)
		return;

	adddynsym(ctxt, s);
	got = linklookup(ctxt, ".got", 0);
	s->got = got->size;
	adduint64(ctxt, got, 0);

	if(iself) {
		rela = linklookup(ctxt, ".rela", 0);
		addaddrplus(ctxt, rela, got, s->got);
		adduint64(ctxt, rela, ELF64_R_INFO(s->dynid, R_X86_64_GLOB_DAT));
		adduint64(ctxt, rela, 0);
	} else {
		diag("addgotsym: unsupported binary format");
	}
}

void
adddynsym(Link *ctxt, LSym *s)
{
	LSym *d;
	int t;
	char *name;

	if(s->dynid >= 0)
		return;

	if(iself) {
		s->dynid = nelfsym++;

		d = linklookup(ctxt, ".dynsym", 0);

		name = s->extname;
		adduint32(ctxt, d, addstring(linklookup(ctxt, ".dynstr", 0), name));
		/* type */
		t = STB_GLOBAL << 4;
		if(s->cgoexport && (s->type&SMASK) == STEXT)
			t |= STT_FUNC;
		else
			t |= STT_OBJECT;
		adduint8(ctxt, d, t);
	
		/* reserved */
		adduint8(ctxt, d, 0);
	
		/* section where symbol is defined */
		if(s->type == SDYNIMPORT)
			adduint16(ctxt, d, SHN_UNDEF);
		else
			adduint16(ctxt, d, 1);
	
		/* value */
		if(s->type == SDYNIMPORT)
			adduint64(ctxt, d, 0);
		else
			addaddr(ctxt, d, s);
	
		/* size of object */
		adduint64(ctxt, d, s->size);
	
		if(!(s->cgoexport & CgoExportDynamic) && s->dynimplib && needlib(s->dynimplib)) {
			elfwritedynent(linklookup(ctxt, ".dynamic", 0), DT_NEEDED,
				addstring(linklookup(ctxt, ".dynstr", 0), s->dynimplib));
		}
	} else {
		diag("adddynsym: unsupported binary format");
	}
}

void
adddynlib(char *lib)
{
	LSym *s;
	
	if(!needlib(lib))
		return;
	
	if(iself) {
		s = linklookup(ctxt, ".dynstr", 0);
		if(s->size == 0)
			addstring(s, "");
		elfwritedynent(linklookup(ctxt, ".dynamic", 0), DT_NEEDED, addstring(s, lib));
	} else {
		diag("adddynlib: unsupported binary format");
	}
}

void
asmb(void)
{
	vlong symo;
	Section *sect;

	if(debug['v'])
		Bprint(&bso, "%5.2f asmb\n", cputime());
	Bflush(&bso);

	if(debug['v'])
		Bprint(&bso, "%5.2f codeblk\n", cputime());
	Bflush(&bso);

	if(iself)
		asmbelfsetup();

	sect = segtext.sect;
	cseek(sect->vaddr - segtext.vaddr + segtext.fileoff);
	codeblk(sect->vaddr, sect->len);
	for(sect = sect->next; sect != nil; sect = sect->next) {
		cseek(sect->vaddr - segtext.vaddr + segtext.fileoff);
		datblk(sect->vaddr, sect->len);
	}

	if(segrodata.filelen > 0) {
		if(debug['v'])
			Bprint(&bso, "%5.2f rodatblk\n", cputime());
		Bflush(&bso);

		cseek(segrodata.fileoff);
		datblk(segrodata.vaddr, segrodata.filelen);
	}

	if(debug['v'])
		Bprint(&bso, "%5.2f datblk\n", cputime());
	Bflush(&bso);

	cseek(segdata.fileoff);
	datblk(segdata.vaddr, segdata.filelen);

	debug['8'] = 1;	/* 64-bit addresses */

	symsize = 0;
	spsize = 0;
	lcsize = 0;
	symo = 0;
	if(!debug['s']) {
		if(debug['v'])
			Bprint(&bso, "%5.2f sym\n", cputime());
		Bflush(&bso);
		symo = segdata.fileoff+segdata.filelen;
		symo = rnd(symo, INITRND);
		cseek(symo);
		if(iself) {
			cseek(symo);
			asmelfsym();
			cflush();
			cwrite(elfstrdat, elfstrsize);

			if(debug['v'])
				Bprint(&bso, "%5.2f dwarf\n", cputime());

			dwarfemitdebugsections();
			elfemitreloc();
		}
	}

	if(debug['v'])
		Bprint(&bso, "%5.2f headr\n", cputime());
	Bflush(&bso);
	cseek(0L);
	asmbelf(symo);
	cflush();
}

vlong
rnd(vlong v, vlong r)
{
	vlong c;

	if(r <= 0)
		return v;
	v += r - 1;
	c = v % r;
	if(c < 0)
		c += r;
	v -= c;
	return v;
}
