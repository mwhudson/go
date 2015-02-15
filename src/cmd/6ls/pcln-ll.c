// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <u.h>
#include <libc.h>
#include <bio.h>
#include <link.h>


// iteration over encoded pcdata tables.

static uint32
getvarint(uchar **pp)
{
	uchar *p;
	int shift;
	uint32 v;

	v = 0;
	p = *pp;
	for(shift = 0;; shift += 7) {
		v |= (uint32)(*p & 0x7F) << shift;
		if(!(*p++ & 0x80))
			break;
	}
	*pp = p;
	return v;
}

void
pciternext(Pciter *it)
{
	uint32 v;
	int32 dv;

	it->pc = it->nextpc;
	if(it->done)
		return;
	if(it->p >= it->d.p + it->d.n) {
		it->done = 1;
		return;
	}

	// value delta
	v = getvarint(&it->p);
	if(v == 0 && !it->start) {
		it->done = 1;
		return;
	}
	it->start = 0;
	dv = (int32)(v>>1) ^ ((int32)(v<<31)>>31);
	it->value += dv;
	
	// pc delta
	v = getvarint(&it->p);
	it->nextpc = it->pc + v*it->pcscale;
}

void
pciterinit(Link *ctxt, Pciter *it, Pcdata *d)
{
	it->d = *d;
	it->p = it->d.p;
	it->pc = 0;
	it->nextpc = 0;
	it->value = -1;
	it->start = 1;
	it->done = 0;
	it->pcscale = ctxt->arch->minlc;
	pciternext(it);
}

