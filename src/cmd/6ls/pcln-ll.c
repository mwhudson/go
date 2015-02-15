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


// Does s have t as a path prefix?
// That is, does s == t or does s begin with t followed by a slash?
// For portability, we allow ASCII case folding, so that haspathprefix("a/b/c", "A/B") is true.
// Similarly, we allow slash folding, so that haspathprefix("a/b/c", "a\\b") is true.
static int
haspathprefix(char *s, char *t)
{
	int i, cs, ct;

	if(t == nil)
		return 0;
	for(i=0; t[i]; i++) {
		cs = s[i];
		ct = t[i];
		if('A' <= cs && cs <= 'Z')
			cs += 'a' - 'A';
		if('A' <= ct && ct <= 'Z')
			ct += 'a' - 'A';
		if(cs == '\\')
			cs = '/';
		if(ct == '\\')
			ct = '/';
		if(cs != ct)
			return 0;
	}
	return s[i] == '\0' || s[i] == '/' || s[i] == '\\';
}

#define HISTSZ 10

// This is a simplified copy of linklinefmt above.
// It doesn't allow printing the full stack, and it returns the file name and line number separately.
// TODO: Unify with linklinefmt somehow.
void
linkgetline(Link *ctxt, int32 line, LSym **f, int32 *l)
{
	struct
	{
		Hist*	incl;	/* start of this include file */
		int32	idel;	/* delta line number to apply to include */
		Hist*	line;	/* start of this #line directive */
		int32	ldel;	/* delta line number to apply to #line */
	} a[HISTSZ];
	int32 lno, d, dlno;
	int n;
	Hist *h;
	char buf[1024], buf1[1024], *file;

	lno = line;
	n = 0;
	for(h=ctxt->hist; h!=nil; h=h->link) {
		if(h->offset < 0)
			continue;
		if(lno < h->line)
			break;
		if(h->name) {
			if(h->offset > 0) {
				// #line directive
				if(n > 0 && n < HISTSZ) {
					a[n-1].line = h;
					a[n-1].ldel = h->line - h->offset + 1;
				}
			} else {
				// beginning of file
				if(n < HISTSZ) {
					a[n].incl = h;
					a[n].idel = h->line;
					a[n].line = 0;
				}
				n++;
			}
			continue;
		}
		n--;
		if(n > 0 && n < HISTSZ) {
			d = h->line - a[n].incl->line;
			a[n-1].ldel += d;
			a[n-1].idel += d;
		}
	}

	if(n > HISTSZ)
		n = HISTSZ;

	if(n <= 0) {
		*f = linklookup(ctxt, "??", HistVersion);
		*l = 0;
		return;
	}
	
	n--;
	if(a[n].line) {
		file = a[n].line->name;
		dlno = a[n].ldel-1;
	} else {
		file = a[n].incl->name;
		dlno = a[n].idel-1;
	}
	if(file[0] == '/' || file[0] == '<')
		snprint(buf, sizeof buf, "%s", file);
	else
		snprint(buf, sizeof buf, "%s/%s", ctxt->pathname, file);

	// Remove leading ctxt->trimpath, or else rewrite $GOROOT to $GOROOT_FINAL.
	if(haspathprefix(buf, ctxt->trimpath)) {
		if(strlen(buf) == strlen(ctxt->trimpath))
			strcpy(buf, "??");
		else {
			snprint(buf1, sizeof buf1, "%s", buf+strlen(ctxt->trimpath)+1);
			if(buf1[0] == '\0')
				strcpy(buf1, "??");
			strcpy(buf, buf1);
		}
	} else if(ctxt->goroot_final != nil && haspathprefix(buf, ctxt->goroot)) {
		snprint(buf1, sizeof buf1, "%s%s", ctxt->goroot_final, buf+strlen(ctxt->goroot));
		strcpy(buf, buf1);
	}

	lno -= dlno;
	*f = linklookup(ctxt, buf, HistVersion);
	*l = lno;
}
