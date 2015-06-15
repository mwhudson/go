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

type Objfile struct {
	header      obj.ObjfileHeader
	offsetBase  int64
	imports     *DataSection
	symdata     *DataSection
	symtable    *DataSection
	stringblock *StringSection
	datablock   *DataSection
	symbols     []*LSym
}

func NewObjfile(ff *obj.Biobuf, pn string, length int64) *Objfile {
	start := obj.Boffset(ff)
	objfile := &Objfile{}
	binary.Read(ff.R(), binary.LittleEndian, &objfile.header)
	if string(objfile.header.Startmagic[:]) != startmagic {
		log.Fatalf("%s: invalid file start %x", pn, objfile.header.Startmagic[:])
	}
	if objfile.header.Version != 2 {
		log.Fatalf("%s: invalid file version number %d", pn, objfile.header.Version)
	}
	objfile.imports = dataSection(ff, objfile.header.ImportsSize, pn, "imports")
	objfile.symdata = dataSection(ff, objfile.header.SymdataSize, pn, "symdata")
	objfile.symtable = dataSection(ff, objfile.header.SymtableSize, pn, "symtable")
	objfile.stringblock = stringSection(ff, objfile.header.StringblockSize, pn, "stringblock")
	objfile.datablock = dataSection(ff, objfile.header.DatablockSize, pn, "datablock")
	var tail [8]byte
	obj.Bread(ff, tail[:])
	if obj.Boffset(ff) != start+length {
		log.Fatalf("%s: unexpected end at %d, want %d", pn, int64(obj.Boffset(ff)), int64(start+length))
	}
	if string(tail[:]) != endmagic {
		log.Fatalf("%s: invalid file end %x", pn, tail[:])
	}
	return objfile
}

func (o *Objfile) rdint(sect *DataSection) int64 {
	var c int

	uv := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("corrupt input")
		}
		c = sect.getc()
		uv |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}

	return int64(uv>>1) ^ (int64(uint64(uv)<<63) >> 63)
}

func (o *Objfile) rdstring(sect *DataSection) string {
	return o.stringblock.read(int(o.rdint(sect)))
}

func (o *Objfile) rddata(sect *DataSection) []byte {
	return o.datablock.read(int(o.rdint(sect)))
}

func (o *Objfile) rdsym(sect *DataSection) *LSym {
	return o.symbols[int(o.rdint(sect))]
}

func dataSection(ff *obj.Biobuf, sz int64, pn, sn string) *DataSection {
	section := &DataSection{data: make([]byte, sz)}
	if obj.Bread(ff, section.data) < 0 {
		log.Fatalf("%s: reading section %s failed", pn, sn)
	}
	return section
}

func stringSection(ff *obj.Biobuf, sz int64, pn, sn string) *StringSection {
	data := make([]byte, sz)
	if obj.Bread(ff, data) < 0 {
		log.Fatalf("%s: reading section %s failed", pn, sn)
	}
	section := &StringSection{data: string(data)}
	return section
}

type DataSection struct {
	data []byte
	pos  int
}

func (b *DataSection) read(n int) []byte {
	p := b.pos
	b.pos += n
	return b.data[p : p+n : p+n]
}

func (b *DataSection) getc() int {
	c := int(b.data[b.pos])
	b.pos++
	return c
}

func (b *DataSection) checkDone() {
	if b.pos != len(b.data) {
		log.Fatalf("checkDone failed")
	}
}

type StringSection struct {
	data string
	pos  int
}

func (b *StringSection) read(n int) string {
	p := b.pos
	b.pos += n
	return b.data[p : p+n]
}

func ldobjfile(ctxt *Link, ff *obj.Biobuf, pkg string, length int64, pn string) {
	ctxt.Version++

	// NewObjfile reads the sections into memory
	objfile := NewObjfile(ff, pn, length)

	// Read the import strings
	var lib string
	for {
		lib = objfile.rdstring(objfile.imports)
		if lib == "" {
			break
		}
		addlib(ctxt, pkg, pn, lib)
	}
	objfile.imports.checkDone()

	// Read the symbol table
	objfile.symbols = []*LSym{nil}
	replacer := strings.NewReplacer(`"".`, pkg+".")
	for {
		s := objfile.rdstring(objfile.symtable)
		if s == "" {
			break
		}
		v := objfile.symtable.getc()
		if v != 0 {
			v = ctxt.Version
		}
		sym := Linklookup(ctxt, replacer.Replace(s), v)

		if v == 0 && s[0] == '$' && sym.Type == 0 {
			if strings.HasPrefix(s, "$f32.") {
				x, _ := strconv.ParseUint(s[5:], 16, 32)
				i32 := int32(x)
				sym.Type = obj.SRODATA
				sym.Local = true
				Adduint32(ctxt, sym, uint32(i32))
				sym.Reachable = false
			} else if strings.HasPrefix(s, "$f64.") || strings.HasPrefix(s, "$i64.") {
				x, _ := strconv.ParseUint(s[5:], 16, 64)
				i64 := int64(x)
				sym.Type = obj.SRODATA
				sym.Local = true
				Adduint64(ctxt, sym, uint64(i64))
				sym.Reachable = false
			}
		}

		objfile.symbols = append(objfile.symbols, sym)
	}
	objfile.symtable.checkDone()

	// Finally, read symbol data
	for {
		c := objfile.symdata.data[objfile.symdata.pos]
		if c == 0xff {
			break
		}
		readsym(ctxt, objfile, pkg, pn)
	}

}

var readsym_ndup int

func readsym(ctxt *Link, objfile *Objfile, pkg string, pn string) {
	f := objfile.symdata
	if f.getc() != 0xfe {
		log.Fatalf("readsym out of sync")
	}
	t := f.getc()
	ind := objfile.rdint(f)
	flags := f.getc()
	dupok := flags & 1
	local := false
	if flags&2 != 0 {
		local = true
	}
	size := int(objfile.rdint(f))
	typ := objfile.rdsym(f)
	s := objfile.symbols[ind]
	data := objfile.rddata(f)
	nreloc := int(objfile.rdint(f))

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
			r.Off = int32(objfile.rdint(f))
			r.Siz = uint8(objfile.rdint(f))
			r.Type = int32(f.getc())
			r.Add = objfile.rdint(f)
			r.Sym = objfile.rdsym(f)
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
		s.Args = int32(objfile.rdint(f))
		s.Locals = int32(objfile.rdint(f))
		s.Nosplit = uint8(f.getc())
		v := f.getc()
		s.Leaf = uint8(v & 1)
		s.Cfunc = uint8(v & 2)
		n := int(objfile.rdint(f))
		var a *Auto
		for i := 0; i < n; i++ {
			a = new(Auto)
			a.Asym = objfile.rdsym(f)
			a.Aoffset = int32(objfile.rdint(f))
			a.Name = int16(f.getc())
			a.Gotype = objfile.rdsym(f)
			a.Link = s.Autom
			s.Autom = a
		}

		s.Pcln = new(Pcln)
		pc := s.Pcln
		pc.Pcsp.P = objfile.rddata(f)
		pc.Pcfile.P = objfile.rddata(f)
		pc.Pcline.P = objfile.rddata(f)
		n = int(objfile.rdint(f))
		pc.Pcdata = make([]Pcdata, n)
		for i := 0; i < n; i++ {
			pc.Pcdata[i].P = objfile.rddata(f)
		}
		n = int(objfile.rdint(f))
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		for i := 0; i < n; i++ {
			pc.Funcdata[i] = objfile.rdsym(f)
		}
		for i := 0; i < n; i++ {
			pc.Funcdataoff[i] = objfile.rdint(f)
		}
		n = int(objfile.rdint(f))
		pc.File = make([]*LSym, n)
		for i := 0; i < n; i++ {
			pc.File[i] = objfile.rdsym(f)
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
