// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"cmd/internal/obj"
	"encoding/binary"
	"fmt"
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
	imports     *GoSection
	symdata     *GoSection
	symtable    *GoSection
	stringblock StringBlock
	datablock   DataBlock
	symbols     []*LSym
}

func NewObjfile(f *obj.Biobuf, pn string, length int64) *Objfile {
	start := obj.Boffset(f)
	objfile := &Objfile{}
	binary.Read(f.R(), binary.LittleEndian, &objfile.header)
	if string(objfile.header.Startmagic[:]) != startmagic {
		log.Fatalf("%s: invalid file start %x", pn, objfile.header.Startmagic[:])
	}
	if objfile.header.Version != 2 {
		log.Fatalf("%s: invalid file version number %d", pn, objfile.header.Version)
	}
	objfile.imports = gosection(objfile, f, objfile.header.ImportsSize, pn, "imports")
	objfile.symdata = gosection(objfile, f, objfile.header.SymdataSize, pn, "symdata")
	objfile.symtable = gosection(objfile, f, objfile.header.SymtableSize, pn, "symtable")

	block := make([]byte, objfile.header.StringblockSize)
	if obj.Bread(f, block) < 0 {
		log.Fatalf("%s: reading string block failed", pn)
	}
	objfile.stringblock.data = string(block)

	block = make([]byte, objfile.header.DatablockSize)
	if obj.Bread(f, block) < 0 {
		log.Fatalf("%s: reading data block failed", pn)
	}
	objfile.datablock.data = block

	var tail [8]byte
	obj.Bread(f, tail[:])
	if obj.Boffset(f) != start+length {
		log.Fatalf("%s: unexpected end at %d, want %d", pn, int64(obj.Boffset(f)), int64(start+length))
	}
	if string(tail[:]) != endmagic {
		log.Fatalf("%s: invalid file end %x", pn, tail[:])
	}
	return objfile
}

type GoSection struct {
	objfile *Objfile
	data    []byte
	pos     int
}

func gosection(objfile *Objfile, f *obj.Biobuf, sz int64, pn, sn string) *GoSection {
	gosection := &GoSection{objfile: objfile, data: make([]byte, sz)}
	if obj.Bread(f, gosection.data) < 0 {
		log.Fatalf("%s: reading section %s failed", pn, sn)
	}
	return gosection
}

func (s *GoSection) getc() int {
	c := int(s.data[s.pos])
	s.pos++
	return c
}

func (s *GoSection) rdint() int64 {
	uv := uint64(0)
	for shift := 0; ; shift += 7 {
		if shift >= 64 {
			log.Fatalf("corrupt input")
		}
		c := s.getc()
		uv |= uint64(c&0x7F) << uint(shift)
		if c&0x80 == 0 {
			break
		}
	}

	return int64(uv>>1) ^ (int64(uint64(uv)<<63) >> 63)
}

func (s *GoSection) rdstring() string {
	return s.objfile.stringblock.read(int(s.rdint()))
}

func (s *GoSection) rddata() []byte {
	return s.objfile.datablock.read(int(s.rdint()))
}

func (s *GoSection) rdsym() *LSym {
	return s.objfile.symbols[int(s.rdint())]
}

type DataBlock struct {
	data []byte
	pos  int
}

func (b *DataBlock) read(n int) []byte {
	p := b.pos
	b.pos += n
	return b.data[p : p+n : p+n]
}

type StringBlock struct {
	data string
	pos  int
}

func (b *StringBlock) read(n int) string {
	p := b.pos
	b.pos += n
	return b.data[p : p+n]
}

func ldobjfile(ctxt *Link, f *obj.Biobuf, pkg string, length int64, pn string) {
	ctxt.Version++

	// NewObjfile reads the sections into memory
	objfile := NewObjfile(f, pn, length)

	// Read the import strings
	var lib string
	for {
		lib = objfile.imports.rdstring()
		if lib == "" {
			break
		}
		addlib(ctxt, pkg, pn, lib)
	}

	// Read the symbol table
	objfile.symbols = []*LSym{nil}
	replacer := strings.NewReplacer(`"".`, pkg+".")
	for {
		s := objfile.symtable.rdstring()
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

	// Finally, read symbol data
	for {
		c := objfile.symdata.data[objfile.symdata.pos]
		if c == 0xff {
			break
		}
		readsym(ctxt, objfile.symdata, pkg, pn)
	}

}

var readsym_ndup int

func readsym(ctxt *Link, d *GoSection, pkg string, pn string) {
	if d.getc() != 0xfe {
		log.Fatalf("readsym out of sync")
	}
	t := int(d.rdint())
	s := d.rdsym()
	flags := d.rdint()
	dupok := flags & 1
	local := false
	if flags&2 != 0 {
		local = true
	}
	size := int(d.rdint())
	typ := d.rdsym()
	data := d.rddata()
	nreloc := int(d.rdint())

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
			r.Off = int32(d.rdint())
			r.Siz = uint8(d.rdint())
			r.Type = int32(d.rdint())
			r.Add = d.rdint()
			r.Sym = d.rdsym()
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
		s.Args = int32(d.rdint())
		s.Locals = int32(d.rdint())
		s.Nosplit = uint8(d.rdint())
		v := d.rdint()
		s.Leaf = uint8(v & 1)
		s.Cfunc = uint8(v & 2)
		n := int(d.rdint())
		var a *Auto
		for i := 0; i < n; i++ {
			a = new(Auto)
			a.Asym = d.rdsym()
			a.Aoffset = int32(d.rdint())
			a.Name = int16(d.rdint())
			a.Gotype = d.rdsym()
			a.Link = s.Autom
			s.Autom = a
		}

		s.Pcln = new(Pcln)
		pc := s.Pcln
		pc.Pcsp.P = d.rddata()
		pc.Pcfile.P = d.rddata()
		pc.Pcline.P = d.rddata()
		n = int(d.rdint())
		pc.Pcdata = make([]Pcdata, n)
		pc.Npcdata = n
		for i := 0; i < n; i++ {
			pc.Pcdata[i].P = d.rddata()
		}
		n = int(d.rdint())
		pc.Funcdata = make([]*LSym, n)
		pc.Funcdataoff = make([]int64, n)
		pc.Nfuncdata = n
		for i := 0; i < n; i++ {
			pc.Funcdata[i] = d.rdsym()
		}
		for i := 0; i < n; i++ {
			pc.Funcdataoff[i] = d.rdint()
		}
		n = int(d.rdint())
		pc.File = make([]*LSym, n)
		pc.Nfile = n
		for i := 0; i < n; i++ {
			pc.File[i] = d.rdsym()
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
	if ctxt.Debugasm != 0 {
		fmt.Fprintf(ctxt.Bso, "%s ", s.Name)
		if s.Version != 0 {
			fmt.Fprintf(ctxt.Bso, "v=%d ", s.Version)
		}
		if s.Type != 0 {
			fmt.Fprintf(ctxt.Bso, "t=%d ", s.Type)
		}
		if s.Dupok != 0 {
			fmt.Fprintf(ctxt.Bso, "dupok ")
		}
		if s.Cfunc != 0 {
			fmt.Fprintf(ctxt.Bso, "cfunc ")
		}
		if s.Nosplit != 0 {
			fmt.Fprintf(ctxt.Bso, "nosplit ")
		}
		fmt.Fprintf(ctxt.Bso, "size=%d value=%d", int64(s.Size), int64(s.Value))
		if s.Type == obj.STEXT {
			fmt.Fprintf(ctxt.Bso, " args=%#x locals=%#x", uint64(s.Args), uint64(s.Locals))
		}
		fmt.Fprintf(ctxt.Bso, "\n")
		var c int
		var j int
		for i := 0; i < len(s.P); {
			fmt.Fprintf(ctxt.Bso, "\t%#04x", uint(i))
			for j = i; j < i+16 && j < len(s.P); j++ {
				fmt.Fprintf(ctxt.Bso, " %02x", s.P[j])
			}
			for ; j < i+16; j++ {
				fmt.Fprintf(ctxt.Bso, "   ")
			}
			fmt.Fprintf(ctxt.Bso, "  ")
			for j = i; j < i+16 && j < len(s.P); j++ {
				c = int(s.P[j])
				if ' ' <= c && c <= 0x7e {
					fmt.Fprintf(ctxt.Bso, "%c", c)
				} else {
					fmt.Fprintf(ctxt.Bso, ".")
				}
			}

			fmt.Fprintf(ctxt.Bso, "\n")
			i += 16
		}

		var r *Reloc
		for i := 0; i < len(s.R); i++ {
			r = &s.R[i]
			fmt.Fprintf(ctxt.Bso, "\trel %d+%d t=%d %s+%d\n", int(r.Off), r.Siz, r.Type, r.Sym.Name, int64(r.Add))
		}
	}
}
