// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"cmd/internal/obj"
	"encoding/binary"
	"log"
	"strconv"
	"strings"
)

const (
	startmagic = "\x00\x00go13ld"
	endmagic   = "\xff\xffgo13ld"
)

var symtable []*LSym

type Membuf struct {
	data []byte
	pos  int
}

func (b *Membuf) read(n int) []byte {
	p := b.pos
	b.pos += n
	return b.data[p : p+n : p+n]
}

func (b *Membuf) getc() int {
	c := int(b.data[b.pos])
	b.pos++
	return c
}

func ldobjfile(ctxt *Link, ff *obj.Biobuf, pkg string, length int64, pn string) {
	start := obj.Boffset(ff)
	ctxt.Version++
	f := &Membuf{data: make([]byte, length)}
	obj.Bread(ff, f.data)
	buf := f.read(8)
	if string(buf) != startmagic {
		log.Fatalf("%s: invalid file start %x %x %x %x %x %x %x %x", pn, buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7])
	}
	c := f.getc()
	if c != 2 {
		log.Fatalf("%s: invalid file version number %d", pn, c)
	}
	symtableOffsetLocation := f.pos
	symtableOffset := int(binary.LittleEndian.Uint32(f.read(4)))

	// Seek further into the file to read the symbol table
	f.pos += symtableOffset - 4
	symtable = nil
	replacer := strings.NewReplacer(`"".`, pkg+".")
	for {
		s := rdstring(f)
		if s == "" {
			break
		}
		v := f.getc()
		if v != 0 {
			v = ctxt.Version
		}
		hash := binary.LittleEndian.Uint32(f.read(4))
		t := replacer.Replace(s)
		if t != s {
			hash = uint32(1)
			for i, c := range t {
				hash *= (uint32(c)<<(uint32(i)%32) + 1)
			}
		}
		symtable = append(symtable, LinklookupHash(ctxt, t, v, hash))
	}
	symtableEnd := f.pos

	// And jump back
	f.pos = symtableOffsetLocation + 4

	var lib string
	for {
		lib = rdstring(f)
		if lib == "" {
			break
		}
		addlib(ctxt, pkg, pn, lib)
	}

	for {
		c := f.data[f.pos]
		if c == 0xff {
			break
		}
		readsym(ctxt, f, pkg, pn)
	}

	if string(f.read(2)) != "\xff\xfd" {
		log.Fatalf("%s: invalid divider", pn)
	}
	// Don't need to read the symbol table again.
	f.pos = symtableEnd
	if string(f.read(8)) != endmagic {
		log.Fatalf("%s: invalid file end", pn)
	}

	if obj.Boffset(ff) != start+length {
		log.Fatalf("%s: unexpected end at %d, want %d", pn, int64(obj.Boffset(ff)), int64(start+length))
	}
}

var readsym_ndup int

func readsym(ctxt *Link, f *Membuf, pkg string, pn string) {
	if f.getc() != 0xfe {
		log.Fatalf("readsym out of sync")
	}
	t := f.getc()
	ind := rdint(f)
	flags := f.getc()
	dupok := flags & 1
	local := false
	if flags&2 != 0 {
		local = true
	}
	size := int(rdint(f))
	typ := rdsym(ctxt, f)
	data := rddata(f)
	nreloc := int(rdint(f))

	s := symtable[ind]
	var dup *LSym
	if s.Type != 0 && s.Type != obj.SXREF {
		if (t == obj.SDATA || t == obj.SBSS || t == obj.SNOPTRBSS) && len(data) == 0 && nreloc == 0 {
			if s.Size < int64(size) {
				s.Size = int64(size)
			}
			if typ != nil && s.Gotype == nil {
				s.Gotype = typ
			}
			return
		}

		if (s.Type == obj.SDATA || s.Type == obj.SBSS || s.Type == obj.SNOPTRBSS) && len(s.P) == 0 && len(s.R) == 0 {
			goto overwrite
		}
		if s.Type != obj.SBSS && s.Type != obj.SNOPTRBSS && dupok == 0 && s.Dupok == 0 {
			log.Fatalf("duplicate symbol %s (types %d and %d) in %s and %s", s.Name, s.Type, t, s.File, pn)
		}
		if len(s.P) > 0 {
			dup = s
			s = linknewsym(ctxt, ".dup", readsym_ndup)
			readsym_ndup++ // scratch
		}
	}

overwrite:
	s.File = pkg
	s.Dupok = uint8(dupok)
	if t == obj.SXREF {
		log.Fatalf("bad sxref")
	}
	if t == 0 {
		log.Fatalf("missing type for %s in %s", s.Name, pn)
	}
	if t == obj.SBSS && (s.Type == obj.SRODATA || s.Type == obj.SNOPTRBSS) {
		t = int(s.Type)
	}
	s.Type = int16(t)
	if s.Size < int64(size) {
		s.Size = int64(size)
	}
	s.Local = local
	if typ != nil { // if bss sym defined multiple times, take type from any one def
		s.Gotype = typ
	}
	if dup != nil && typ != nil {
		dup.Gotype = typ
	}
	s.P = data
	if nreloc > 0 {
		s.R = make([]Reloc, nreloc)
		var r *Reloc
		for i := 0; i < nreloc; i++ {
			r = &s.R[i]
			r.Off = int32(rdint(f))
			r.Siz = uint8(rdint(f))
			r.Type = int32(f.getc())
			r.Add = rdint(f)
			r.Sym = rdsym(ctxt, f)
		}
	}

	if len(s.P) > 0 && dup != nil && len(dup.P) > 0 && strings.HasPrefix(s.Name, "gclocals·") {
		// content-addressed garbage collection liveness bitmap symbol.
		// double check for hash collisions.
		if !bytes.Equal(s.P, dup.P) {
			log.Fatalf("dupok hash collision for %s in %s and %s", s.Name, s.File, pn)
		}
	}

	if s.Type == obj.STEXT {
		s.Args = int32(rdint(f))
		s.Locals = int32(rdint(f))
		s.Nosplit = uint8(f.getc())
		v := f.getc()
		s.Leaf = uint8(v & 1)
		s.Cfunc = uint8(v & 2)
		n := int(rdint(f))
		var a *Auto
		for i := 0; i < n; i++ {
			a = new(Auto)
			a.Asym = rdsym(ctxt, f)
			a.Aoffset = int32(rdint(f))
			a.Name = int16(f.getc())
			a.Gotype = rdsym(ctxt, f)
			a.Link = s.Autom
			s.Autom = a
		}

		s.Pcln = new(Pcln)
		pc := s.Pcln
		pc.Pcsp.P = rddata(f)
		pc.Pcfile.P = rddata(f)
		pc.Pcline.P = rddata(f)
		n = int(rdint(f))
		pc.Pcdata = make([]Pcdata, n)
		for i := 0; i < n; i++ {
			pc.Pcdata[i].P = rddata(f)
		}
		n = int(rdint(f))
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		for i := 0; i < n; i++ {
			pc.Funcdata[i] = rdsym(ctxt, f)
		}
		for i := 0; i < n; i++ {
			pc.Funcdataoff[i] = rdint(f)
		}
		n = int(rdint(f))
		pc.File = make([]*LSym, n)
		for i := 0; i < n; i++ {
			pc.File[i] = rdsym(ctxt, f)
		}

		if dup == nil {
			if s.Onlist != 0 {
				log.Fatalf("symbol %s listed multiple times", s.Name)
			}
			s.Onlist = 1
			if ctxt.Etextp != nil {
				ctxt.Etextp.Next = s
			} else {
				ctxt.Textp = s
			}
			ctxt.Etextp = s
		}
	}
}

func rdint(f *Membuf) int64 {
	var c int

	uv := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("corrupt input")
		}
		c = f.getc()
		uv |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}

	return int64(uv>>1) ^ (int64(uint64(uv)<<63) >> 63)
}

func rdstring(f *Membuf) string {
	return string(rddata(f))
}

func rddata(f *Membuf) []byte {
	n := rdint(f)
	return f.read(int(n))
}

func rdsym(ctxt *Link, f *Membuf) *LSym {
	ind := int(rdint(f))
	if ind == -1 {
		return nil
	}
	s := symtable[ind]

	if s.Version == 0 && s.Name[0] == '$' && s.Type == 0 {
		if strings.HasPrefix(s.Name, "$f32.") {
			x, _ := strconv.ParseUint(s.Name[5:], 16, 32)
			i32 := int32(x)
			s.Type = obj.SRODATA
			s.Local = true
			Adduint32(ctxt, s, uint32(i32))
			s.Reachable = false
		} else if strings.HasPrefix(s.Name, "$f64.") || strings.HasPrefix(s.Name, "$i64.") {
			x, _ := strconv.ParseUint(s.Name[5:], 16, 64)
			i64 := int64(x)
			s.Type = obj.SRODATA
			s.Local = true
			Adduint64(ctxt, s, uint64(i64))
			s.Reachable = false
		}
	}
	return s
}
