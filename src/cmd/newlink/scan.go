// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Initial scan of packages making up a program.

// TODO(rsc): Rename goobj.SymID.Version to StaticID to avoid confusion with the ELF meaning of version.
// TODO(rsc): Fix file format so that SBSS/SNOPTRBSS with data is listed as SDATA/SNOPTRDATA.
// TODO(rsc): Parallelize scan to overlap file i/o where possible.

package main

import (
	"cmd/internal/goobj"
	"os"
	"sort"
	"strings"
)

// scan scans all packages making up the program, starting with package main defined in mainfile.
func (p *Prog) scan(mainfile string) {
	p.initScan()
	p.scanFile("main", mainfile)
	if len(p.Missing) > 0 && !p.omitRuntime {
		p.scanImport("runtime")
	}

	var missing []string
	for sym := range p.Missing {
		if !p.isAuto(p.Syms[sym].Sym) {
			missing = append(missing, sym.String())
		}
	}

	if missing != nil {
		sort.Strings(missing)
		for _, sym := range missing {
			p.errorf("undefined: %s", sym)
		}
	}

	// TODO(rsc): Walk import graph to diagnose cycles.
}

// initScan initializes the Prog fields needed by scan.
func (p *Prog) initScan() {
	p.Packages = make(map[string]*Package)
	p.Syms = make(map[goobj.SymID]*Sym)
	p.Missing = make(map[goobj.SymID]bool)
	p.Missing[p.startSym] = true
}

func (p *Prog) findSym(symid goobj.SymID) *goobj.Sym {
	s := p.Syms[symid]
	if s == nil {
		return nil
	}
	return s.Sym
}

func (p *Prog) defineSym(pkg *Package, gs *goobj.Sym) {
	if gs.Data.Size > 0 {
		switch gs.Kind {
		case goobj.SBSS:
			gs.Kind = goobj.SDATA
		case goobj.SNOPTRBSS:
			gs.Kind = goobj.SNOPTRDATA
		}
	}

	if gs.Version != 0 {
		gs.Version += p.MaxVersion
	}
	for i := range gs.Reloc {
		r := &gs.Reloc[i]
		if r.Sym.Version != 0 {
			r.Sym.Version += p.MaxVersion
		}
		if p.Syms[r.Sym.SymID] == nil {
			p.Missing[r.Sym.SymID] = true
		}
	}
	if gs.Func != nil {
		for i := range gs.Func.FuncData {
			fdata := &gs.Func.FuncData[i]
			if fdata.Sym.Name != "" {
				if fdata.Sym.Version != 0 {
					fdata.Sym.Version += p.MaxVersion
				}
				if p.Syms[fdata.Sym.SymID] == nil {
					p.Missing[fdata.Sym.SymID] = true
				}
			}
		}
	}
	if old := p.Syms[gs.SymID]; old != nil {
		// Duplicate definition of symbol. Is it okay?
		// TODO(rsc): Write test for this code.
		switch {
		// If both symbols are BSS (no data), take max of sizes
		// but otherwise ignore second symbol.
		case old.Data.Size == 0 && gs.Data.Size == 0:
			if old.Size < gs.Size {
				old.Size = gs.Size
			}
			return

		// If one is in BSS and one is not, use the one that is not.
		case old.Data.Size > 0 && gs.Data.Size == 0:
			return
		case gs.Data.Size > 0 && old.Data.Size == 0:
			break // install gs as new symbol below

		// If either is marked as DupOK, we can keep either one.
		// Keep the one that we saw first.
		case old.DupOK || gs.DupOK:
			return

		// Otherwise, there's an actual conflict:
		default:
			p.errorf("symbol %s defined in both %s and %s %v %v", gs.SymID, old.Package.File, pkg.File, old.Data, gs.Data)
			return
		}
	}
	s := &Sym{
		Sym:     gs,
		Package: pkg,
	}
	p.addSym(s)
	delete(p.Missing, gs.SymID)

	if s.Data.Size > int64(s.Size) {
		p.errorf("%s: initialized data larger than symbol (%d > %d)", s, s.Data.Size, s.Size)
	}

}

// scanFile reads file to learn about the package with the given import path.
func (p *Prog) scanFile(pkgpath string, file string) {
	pkg := &Package{
		File: file,
	}
	p.Packages[pkgpath] = pkg

	f, err := os.Open(file)
	if err != nil {
		p.errorf("%v", err)
		return
	}
	gp, err := goobj.Parse(f, pkgpath, p.findSym, func(gs *goobj.Sym) { p.defineSym(pkg, gs) })
	f.Close()
	if err != nil {
		p.errorf("reading %s: %v", file, err)
		return
	}

	// TODO(rsc): Change cmd/internal/goobj to record package name as gp.Name.
	// TODO(rsc): If pkgpath == "main", check that gp.Name == "main".

	pkg.Package = gp

	p.MaxVersion += pkg.MaxVersion

	for i, pkgpath := range pkg.Imports {
		// TODO(rsc): Fix file format to drop .a from recorded import path.
		pkgpath = strings.TrimSuffix(pkgpath, ".a")
		pkg.Imports[i] = pkgpath

		p.scanImport(pkgpath)
	}
}

func (p *Prog) addSym(s *Sym) {
	pkg := s.Package
	if pkg == nil {
		pkg = p.Packages[""]
		if pkg == nil {
			pkg = &Package{}
			p.Packages[""] = pkg
		}
		s.Package = pkg
	}
	pkg.Syms = append(pkg.Syms, s)
	p.Syms[s.SymID] = s
	p.SymOrder = append(p.SymOrder, s)
}

// scanImport finds the object file for the given import path and then scans it.
func (p *Prog) scanImport(pkgpath string) {
	if p.Packages[pkgpath] != nil {
		return // already loaded
	}

	// TODO(rsc): Implement correct search to find file.
	p.scanFile(pkgpath, p.pkgdir+"/"+pkgpath+".a")
}
