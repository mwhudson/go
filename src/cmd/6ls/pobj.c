// Inferno utils/6l/obj.c
// http://code.google.com/p/inferno-os/source/browse/utils/6l/obj.c
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

// Reading object files.

#define	EXTERN
#include	"l.h"
#include "lib.h"
#include "elf.h"
#include "dwarf.h"
#include	<ar.h>

void
main(int argc, char *argv[])
{
	int i;

	linkarchinit();
	ctxt = linknew(thelinkarch);
	ctxt->bso = &bso;

	Binit(&bso, 1, OWRITE);
	memset(debug, 0, sizeof(debug));
	nerrors = 0;
	outfile = nil;
	INITTEXT = -1;
	INITDAT = -1;
	INITRND = -1;
	INITENTRY = 0;
	
	// For testing behavior of go command when tools crash.
	// Undocumented, not in standard flag parser to avoid
	// exposing in usage message.
	for(i=1; i<argc; i++)
		if(strcmp(argv[i], "-crash_for_testing") == 0)
			*(volatile int*)0 = 0;
	

	flagfn1("B", "info: define ELF NT_GNU_BUILD_ID note", addbuildinfo);
	flagcount("C", "check Go calls to C code", &debug['C']);
	flagint64("D", "addr: data address", &INITDAT);
	flagstr("E", "sym: entry symbol", &INITENTRY);
	flagfn1("I", "interp: set ELF interp", setinterp);
	flagfn1("L", "dir: add dir to library path", Lflag);
	flagcount("K", "add stack underflow checks", &debug['K']);
	flagcount("O", "print pc-line tables", &debug['O']);
	flagint32("R", "rnd: address rounding", &INITRND);
	flagcount("S", "check type signatures", &debug['S']);
	flagint64("T", "addr: text address", &INITTEXT);
	flagfn0("V", "print version and exit", doversion);
	flagfn2("X", "name value: define string data", addstrdata);
	flagcount("a", "disassemble output", &debug['a']);
	flagcount("c", "dump call graph", &debug['c']);
	flagcount("d", "disable dynamic executable", &debug['d']);
	flagstr("extld", "ld: linker to run in external mode", &extld);
	flagstr("extldflags", "ldflags: flags for external linker", &extldflags);
	flagcount("f", "ignore version mismatch", &debug['f']);
	flagcount("g", "disable go package data checks", &debug['g']);
	flagstr("installsuffix", "suffix: pkg directory suffix", &flag_installsuffix);
	flagstr("k", "sym: set field tracking symbol", &tracksym);
	flagcount("n", "dump symbol table", &debug['n']);
	flagstr("o", "outfile: set output file", &outfile);
	flagstr("r", "dir1:dir2:...: set ELF dynamic linker search path", &rpath);
	flagcount("race", "enable race detector", &flag_race);
	flagcount("s", "disable symbol table", &debug['s']);
	flagcount("shared", "generate shared object", &flag_shared);
	flagcount("dso", "generate shared library", &ctxt->flag_dso);
	flagstr("tmpdir", "dir: leave temporary files in this directory", &tmpdir);
	flagcount("u", "reject unsafe packages", &debug['u']);
	flagcount("v", "print link trace", &debug['v']);
	flagcount("w", "disable DWARF generation", &debug['w']);
	
	flagparse(&argc, &argv, usage);

	if(ctxt->flag_dso) {
		flag_shared = 1;
	}

	ctxt->bso = &bso;
	ctxt->debugvlog = debug['v'];

	if(!ctxt->flag_dso && argc != 1)
		usage();

	if(outfile == nil) {
		outfile = "6.out";
	}
	libinit(); // creates outfile

	archinit();

	if(debug['v'])
		Bprint(&bso, "HEADER = -T0x%llux -D0x%llux -R0x%ux\n",
			INITTEXT, INITDAT, INITRND);
	Bflush(&bso);

	cbp = buf.cbuf;
	cbc = sizeof(buf.cbuf);

	if(ctxt->flag_dso) {
		int i = 0;
		while (i < argc) {
			if (strcmp(argv[i], "ar") == 0) {
				if (i + 2 >= argc) {
					usage();
				}
				//ctxt->isexe |= (strcmp(argv[i+1], "main") == 0);
				addlibpath(ctxt, "command line", "command line", argv[i+2], argv[i+1], NULL);
				i += 3;
			} else if (strcmp(argv[i], "dso") == 0) {
				if (i + 3 >= argc) {
					usage();
				}
				addlibpath(ctxt, "command line", "command line", argv[i+2], argv[i+1], argv[i+3]);
				i += 4;
			} else {
				usage();
			}
		}
		ctxt->addlibpath_ok = 0;
	} else {
		addlibpath(ctxt, "command line", "command line", argv[0], "main", NULL);
	}

	loadlib();

	checkgo();
	deadcode();
	callgraph();

	doelf();
	dostkcheck();
	addexport();
	gentext();		// trampolines, call stubs, etc.
	textaddress();
	pclntab();
	findfunctab();
	symtab();
	dodata();
	address();
	doweak();
	reloc();
	asmb();
	undef();
	hostlink();
	if(debug['v']) {
		Bprint(&bso, "%5.2f cpu time\n", cputime());
		Bprint(&bso, "%d symbols\n", ctxt->nsymbol);
		Bprint(&bso, "%lld liveness data\n", liveness);
	}
	Bflush(&bso);

	errorexit();
}


LinkArch linkamd64 = {
	.name = "amd64",
	.endian = LittleEndian,

	.minlc = 1,
	.ptrsize = 8,
	.regsize = 8,
};

LinkArch linkamd64p32 = {
	.name = "amd64p32",
	.endian = LittleEndian,

	.minlc = 1,
	.ptrsize = 4,
	.regsize = 8,
};
