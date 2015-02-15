// Derived from Inferno utils/6l/l.h and related files.
// http://code.google.com/p/inferno-os/source/browse/utils/6l/l.h
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

typedef	struct	Addr	Addr;
typedef	struct	Prog	Prog;
typedef	struct	LSym	LSym;
typedef	struct	Reloc	Reloc;
typedef	struct	Auto	Auto;
typedef	struct	Hist	Hist;
typedef	struct	Link	Link;
typedef	struct	Plist	Plist;
typedef	struct	LinkArch	LinkArch;
typedef	struct	Library	Library;

typedef	struct	Pcln	Pcln;
typedef	struct	Pcdata	Pcdata;
typedef	struct	Pciter	Pciter;

// An Addr is an argument to an instruction.
// The general forms and their encodings are:
//
//	sym±offset(symkind)(reg)(index*scale)
//		Memory reference at address &sym(symkind) + offset + reg + index*scale.
//		Any of sym(symkind), ±offset, (reg), (index*scale), and *scale can be omitted.
//		If (reg) and *scale are both omitted, the resulting expression (index) is parsed as (reg).
//		To force a parsing as index*scale, write (index*1).
//		Encoding:
//			type = TYPE_MEM
//			name = symkind (NAME_AUTO, ...) or 0 (NAME_NONE)
//			sym = sym
//			offset = ±offset
//			reg = reg (REG_*)
//			index = index (REG_*)
//			scale = scale (1, 2, 4, 8)
//
//	$<mem>
//		Effective address of memory reference <mem>, defined above.
//		Encoding: same as memory reference, but type = TYPE_ADDR.
//
//	$<±integer value>
//		This is a special case of $<mem>, in which only ±offset is present.
//		It has a separate type for easy recognition.
//		Encoding:
//			type = TYPE_CONST
//			offset = ±integer value
//
//	*<mem>
//		Indirect reference through memory reference <mem>, defined above.
//		Only used on x86 for CALL/JMP *sym(SB), which calls/jumps to a function
//		pointer stored in the data word sym(SB), not a function named sym(SB).
//		Encoding: same as above, but type = TYPE_INDIR.
//
//	$*$<mem>
//		No longer used.
//		On machines with actual SB registers, $*$<mem> forced the
//		instruction encoding to use a full 32-bit constant, never a
//		reference relative to SB.
//
//	$<floating point literal>
//		Floating point constant value.
//		Encoding:
//			type = TYPE_FCONST
//			u.dval = floating point value
//
//	$<string literal, up to 8 chars>
//		String literal value (raw bytes used for DATA instruction).
//		Encoding:
//			type = TYPE_SCONST
//			u.sval = string
//
//	<register name>
//		Any register: integer, floating point, control, segment, and so on.
//		If looking for specific register kind, must check type and reg value range.
//		Encoding:
//			type = TYPE_REG
//			reg = reg (REG_*)
//
//	x(PC)
//		Encoding:
//			type = TYPE_BRANCH
//			u.branch = Prog* reference OR ELSE offset = target pc (branch takes priority)
//
//	$±x-±y
//		Final argument to TEXT, specifying local frame size x and argument size y.
//		In this form, x and y are integer literals only, not arbitrary expressions.
//		This avoids parsing ambiguities due to the use of - as a separator.
//		The ± are optional.
//		If the final argument to TEXT omits the -±y, the encoding should still
//		use TYPE_TEXTSIZE (not TYPE_CONST), with u.argsize = ArgsSizeUnknown.
//		Encoding:
//			type = TYPE_TEXTSIZE
//			offset = x
//			u.argsize = y
//
//	reg<<shift, reg>>shift, reg->shift, reg@>shift
//		Shifted register value, for ARM.
//		In this form, reg must be a register and shift can be a register or an integer constant.
//		Encoding:
//			type = TYPE_SHIFT
//			offset = (reg&15) | shifttype<<5 | count
//			shifttype = 0, 1, 2, 3 for <<, >>, ->, @>
//			count = (reg&15)<<8 | 1<<4 for a register shift count, (n&31)<<7 for an integer constant.
//
//	(reg, reg)
//		A destination register pair. When used as the last argument of an instruction,
//		this form makes clear that both registers are destinations.
//		Encoding:
//			type = TYPE_REGREG
//			reg = first register
//			offset = second register
//
//	reg, reg
//		TYPE_REGREG2, to be removed.
//
struct	Addr
{
	int16	type; // could be int8
	int16	reg;
	int16	index;
	int8	scale;
	int8	name;
	int64	offset;
	LSym*	sym;
	
	union
	{
		char	sval[8];
		float64	dval;
		Prog*	branch;
		int32	argsize;	// for 5l, 8l
		uint64	bits; // raw union bits, for testing if anything has been written to any field
	} u;

	// gotype is the name of the Go type descriptor for sym.
	// It cannot be set using assembly syntax.
	// It is generated by the Go compiler for global declarations,
	// to convey information about pointer locations to the back end
	// and for use in generating debug information.
	LSym*	gotype;

	int8	class;	// for internal use by liblink
	uint8	etype; // for internal use by 5g, 6g, 8g
	void*	node; // for internal use by 5g, 6g, 8g
	int64	width; // for internal use by 5g, 6g, 8g
};

enum {
	NAME_NONE = 0,
	NAME_EXTERN,
	NAME_STATIC,
	NAME_AUTO,
	NAME_PARAM,
};

enum {
	TYPE_NONE = 0,
	TYPE_BRANCH = 5, // avoid accidental conflicts with NAME_* 
	TYPE_TEXTSIZE,
	TYPE_MEM,
	TYPE_CONST,
	TYPE_FCONST,
	TYPE_SCONST,
	TYPE_REG,
	TYPE_ADDR,
	TYPE_SHIFT,
	TYPE_REGREG,
	TYPE_REGREG2,
	TYPE_INDIR,
};

struct	Reloc
{
	int32	off;
	uchar	siz;
	uchar	done;
	int32	type;
	int32	variant; // RV_*: variant on computed value
	int64	add;
	int64	xadd;
	LSym*	sym;
	LSym*	xsym;
};

// TODO(rsc): Describe prog.
// TODO(rsc): Describe TEXT/GLOBL flag in from3, DATA width in from3.
struct	Prog
{
	vlong	pc;
	int32	lineno;
	Prog*	link;
	short	as;
	uchar	scond; // arm only; condition codes

	// operands
	Addr	from;
	int16	reg; // arm, ppc64 only (e.g., ADD from, reg, to);
		     // starts at 0 for both GPRs and FPRs;
		     // also used for ADATA width on arm, ppc64
	Addr	from3; // addl source argument (e.g., RLWM/FMADD from, reg, from3, to)
	Addr	to;
	
	// for 5g, 6g, 8g internal use
	void*	opt;

	// for liblink internal use
	Prog*	forwd;
	Prog*	pcond;
	Prog*	comefrom;	// amd64, 386
	Prog*	pcrel;	// arm
	int32	spadj;
	uint16	mark;
	uint16	optab;	// arm, ppc64
	uchar	back;	// amd64, 386
	uchar	ft;	// oclass cache
	uchar	tt;	// oclass cache
	uchar	isize;	// amd64, 386

	char	width;	/* fake for DATA */
	char	mode;	/* 16, 32, or 64 in 6l, 8l; internal use in 5g, 6g, 8g */
};

extern Prog zprog; // zeroed Prog

// Prog.as opcodes.
// These are the portable opcodes, common to all architectures.
// Each architecture defines many more arch-specific opcodes,
// with values starting at A_ARCHSPECIFIC.
enum {
	AXXX = 0,

	ACALL,
	ACHECKNIL,
	ADATA,
	ADUFFCOPY,
	ADUFFZERO,
	AEND,
	AFUNCDATA,
	AGLOBL,
	AJMP,
	ANOP,
	APCDATA,
	ARET,
	ATEXT,
	ATYPE,
	AUNDEF,
	AUSEFIELD,
	AVARDEF,
	AVARKILL,
	
	A_ARCHSPECIFIC, // first architecture-specific opcode value
};

// prevent incompatible type signatures between liblink and 8l on Plan 9
#pragma incomplete struct Section

struct	LSym
{
	char*	name;
	char*	extname;	// name used in external object files
	short	type;
	short	version;
	uchar	dupok;
	uchar	cfunc;
	uchar	external;
	uchar	nosplit;
	uchar	reachable;
	uchar	cgoexport;
	uchar	special;
	uchar	stkcheck;
	uchar	hide;
	uchar	leaf;	// arm only
	uchar	fnptr;	// arm only
	uchar	localentry;	// ppc64: instrs between global & local entry
	uchar	seenglobl;
	uchar	onlist;	// on the textp or datap lists
	int16	symid;	// for writing .5/.6/.8 files
	int32	dynid;
	int32	sig;
	int32	plt;
	int32	got;
	int32	align;	// if non-zero, required alignment in bytes
	int32	elfsym;
	int32	args;	// size of stack frame incoming arguments area
	int32	locals;	// size of stack frame locals area (arm only?)
	vlong	value;
	vlong	size;
	LSym*	hash;	// in hash table
	LSym*	allsym;	// in all symbol list
	LSym*	next;	// in text or data list
	LSym*	sub;	// in SSUB list
	LSym*	outer;	// container of sub
	LSym*	gotype;
	LSym*	reachparent;
	LSym*	queue;
	char*	file;
	char*	dynimplib;
	char*	dynimpvers;
	struct Section*	sect;
	
	// STEXT
	Auto*	autom;
	Prog*	text;
	Prog*	etext;
	Pcln*	pcln;

	// SDATA, SBSS
	uchar*	p;
	int	np;
	int32	maxp;
	Reloc*	r;
	int32	nr;
	int32	maxr;
};

// LSym.type
enum
{
	Sxxx,

	/* order here is order in output file */
	/* readonly, executable */
	STEXT,
	SELFRXSECT,
	
	/* readonly, non-executable */
	STYPE,
	SSTRING,
	SGOSTRING,
	SGOFUNC,
	SRODATA,
	SFUNCTAB,
	STYPELINK,
	SSYMTAB, // TODO: move to unmapped section
	SPCLNTAB,
	SELFROSECT,
	
	/* writable, non-executable */
	SMACHOPLT,
	SELFSECT,
	SMACHO,	/* Mach-O __nl_symbol_ptr */
	SMACHOGOT,
	SWINDOWS,
	SELFGOT,	/* also .toc in ppc64 ABI */
	SNOPTRDATA,
	SINITARR,
	SDATA,
	SBSS,
	SNOPTRBSS,
	STLSBSS,

	/* not mapped */
	SXREF,
	SMACHOSYMSTR,
	SMACHOSYMTAB,
	SMACHOINDIRECTPLT,
	SMACHOINDIRECTGOT,
	SFILE,
	SFILEPATH,
	SCONST,
	SDYNIMPORT,
	SHOSTOBJ,

	SSUB = 1<<8,	/* sub-symbol, linked from parent via ->sub list */
	SMASK = SSUB - 1,
	SHIDDEN = 1<<9, // hidden or local symbol
};

// Reloc.type
enum
{
	R_ADDR = 1,
	R_ADDRPOWER, // relocation for loading 31-bit address using addis and addi/ld/st for Power
	R_SIZE,
	R_CALL, // relocation for direct PC-relative call
	R_CALLARM, // relocation for ARM direct call
	R_CALLIND, // marker for indirect call (no actual relocating necessary)
	R_CALLPOWER, // relocation for Power direct call
	R_CONST,
	R_PCREL,
	R_TLS,
	R_TLS_LE, // TLS local exec offset from TLS segment register
	R_TLS_IE, // TLS initial exec offset from TLS base pointer
	R_GOTOFF,
	R_PLT0,
	R_PLT1,
	R_PLT2,
	R_USEFIELD,
	R_POWER_TOC,		// ELF R_PPC64_TOC16*
};

// Reloc.variant
enum
{
	RV_NONE,		// identity variant
	RV_POWER_LO,		// x & 0xFFFF
	RV_POWER_HI,		// x >> 16
	RV_POWER_HA,		// (x + 0x8000) >> 16
	RV_POWER_DS,		// x & 0xFFFC, check x&0x3 == 0

	RV_CHECK_OVERFLOW = 1<<8,	// check overflow flag
	RV_TYPE_MASK = (RV_CHECK_OVERFLOW - 1),
};

// Auto.name
enum
{
	A_AUTO = 1,
	A_PARAM,
};

struct	Auto
{
	LSym*	asym;
	Auto*	link;
	int32	aoffset;
	int16	name;
	LSym*	gotype;
};

enum
{
	LINKHASH = 100003,
};

struct	Hist
{
	Hist*	link;
	char*	name;
	int32	line;
	int32	offset;
};

struct	Plist
{
	LSym*	name;
	Prog*	firstpc;
	int	recur;
	Plist*	link;
};

struct	Library
{
	char *objref;	// object where we found the reference
	char *srcref;	// src file where we found the reference
	char *file;	// object file
	char *pkg;	// import path
	char *dso;
};

struct Pcdata
{
	uchar *p;
	int n;
	int m;
};

struct Pcln
{
	Pcdata pcsp;
	Pcdata pcfile;
	Pcdata pcline;
	Pcdata *pcdata;
	int npcdata;
	LSym **funcdata;
	int64 *funcdataoff;
	int nfuncdata;
	
	LSym **file;
	int nfile;
	int mfile;

	LSym *lastfile;
	int lastindex;
};

// Pcdata iterator.
//	for(pciterinit(ctxt, &it, &pcd); !it.done; pciternext(&it)) { it.value holds in [it.pc, it.nextpc) }
struct Pciter
{
	Pcdata d;
	uchar *p;
	uint32 pc;
	uint32 nextpc;
	uint32 pcscale;
	int32 value;
	int start;
	int done;
};

void	pciterinit(Link*, Pciter*, Pcdata*);
void	pciternext(Pciter*);

// symbol version, incremented each time a file is loaded.
// version==1 is reserved for savehist.
enum
{
	HistVersion = 1,
};

// Link holds the context for writing object code from a compiler
// to be linker input or for reading that input into the linker.
struct	Link
{

	LinkArch*	arch;
	int32	debugvlog;	// -v flag in linker
	int32	debugpcln;	// -O flag in linker
	int32	flag_shared;	// -shared flag in linker
	int32	flag_dso;	// -dso flag in linker
	int32	addlibpath_ok;	// shush
	Biobuf*	bso;	// for -v flag

	// hash table of all symbols
	LSym*	hash[LINKHASH];
	LSym*	allsym;
	int32	nsymbol;

	// code generation
	LSym*	tlsg;

	// for reading input files (during linker)
	char**	libdir;
	int32	nlibdir;
	int32	maxlibdir;
	Library*	library;
	int	libraryp;
	int	nlibrary;
	void	(*diag)(char*, ...);
	LSym*	cursym;
	int	version;
	LSym*	textp;
	LSym*	etextp;
	int32	nhistfile;
	LSym*	filesyms;
};

enum {
	LittleEndian = 0x04030201,
	BigEndian = 0x01020304,
};

// LinkArch is the definition of a single architecture.
struct LinkArch
{
	char*	name; // "arm", "amd64", and so on
	int32	endian; // LittleEndian or BigEndian

	int	minlc;
	int	ptrsize;
	int	regsize;
};

/* executable header types */
enum {
	Hunknown = 0,
	Hlinux,
};

extern	uchar	fnuxi8[8];
extern	uchar	fnuxi4[4];
extern	uchar	inuxi1[1];
extern	uchar	inuxi2[2];
extern	uchar	inuxi4[4];
extern	uchar	inuxi8[8];

// data.c
vlong	addaddr(Link *ctxt, LSym *s, LSym *t);
vlong	addaddrplus(Link *ctxt, LSym *s, LSym *t, vlong add);
vlong	addaddrplus4(Link *ctxt, LSym *s, LSym *t, vlong add);
vlong	addpcrelplus(Link *ctxt, LSym *s, LSym *t, vlong add);
Reloc*	addrel(LSym *s);
vlong	addsize(Link *ctxt, LSym *s, LSym *t);
vlong	adduint16(Link *ctxt, LSym *s, uint16 v);
vlong	adduint32(Link *ctxt, LSym *s, uint32 v);
vlong	adduint64(Link *ctxt, LSym *s, uint64 v);
vlong	adduint8(Link *ctxt, LSym *s, uint8 v);
vlong	adduintxx(Link *ctxt, LSym *s, uint64 v, int wid);
void	mangle(char *file);
void	savedata(Link *ctxt, LSym *s, Prog *p, char *pn);
void	savedata1(Link *ctxt, LSym *s, Prog *p, char *pn, int enforce_order);
vlong	setaddr(Link *ctxt, LSym *s, vlong off, LSym *t);
vlong	setaddrplus(Link *ctxt, LSym *s, vlong off, LSym *t, vlong add);
vlong	setuint16(Link *ctxt, LSym *s, vlong r, uint16 v);
vlong	setuint32(Link *ctxt, LSym *s, vlong r, uint32 v);
vlong	setuint64(Link *ctxt, LSym *s, vlong r, uint64 v);
vlong	setuint8(Link *ctxt, LSym *s, vlong r, uint8 v);
vlong	setuintxx(Link *ctxt, LSym *s, vlong off, uint64 v, vlong wid);
void	symgrow(Link *ctxt, LSym *s, vlong siz);

// go.c
void*	emallocz(long n);
void*	erealloc(void *p, long n);
char*	estrdup(char *p);
char*	expandpkg(char *t0, char *pkg);

// ld.c
void	addlib(Link *ctxt, char *src, char *obj, char *path);
void	addlibpath(Link *ctxt, char *srcref, char *objref, char *file, char *pkg, char* dso);
int	find1(int32 l, int c);
void	nuxiinit(LinkArch*);
vlong	atolwhex(char*);

// list[5689].c
void	listinit6(void);

// obj.c
int	linklinefmt(Link *ctxt, Fmt *fp);
void	linklinehist(Link *ctxt, int lineno, char *f, int offset);
Plist*	linknewplist(Link *ctxt);
void	linkprfile(Link *ctxt, int32 l);

// objfile.c
void	ldobjfile(Link *ctxt, Biobuf *b, char *pkg, int64 len, char *path);

// sym.c
LSym*	linklookup(Link *ctxt, char *name, int v);
Link*	linknew(LinkArch*);
LSym*	linknewsym(Link *ctxt, char *symb, int v);
LSym*	linkrlookup(Link *ctxt, char *name, int v);
int	linksymfmt(Fmt *f);

extern	char*	anames6[];
extern	char*	dnames6[];

extern	LinkArch	linkamd64;
extern	LinkArch	linkamd64p32;

#pragma	varargck	type	"A"	int
#pragma	varargck	type	"E"	uint
#pragma	varargck	type	"D"	Addr*
#pragma	varargck	type	"lD"	Addr*
#pragma	varargck	type	"P"	Prog*
#pragma	varargck	type	"R"	int
#pragma	varargck	type	"^"	int // for 5l/9l, C_* classes (liblink internal)

// TODO(ality): remove this workaround.
//   It's here because Pconv in liblink/list?.c references %L.
#pragma	varargck	type	"L"	int32
