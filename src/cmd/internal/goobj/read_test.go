// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goobj

import (
	"go/build"
	"log"
	"os"
	"testing"
)

var importPathToPrefixTests = []struct {
	in  string
	out string
}{
	{"runtime", "runtime"},
	{"sync/atomic", "sync/atomic"},
	{"golang.org/x/tools/godoc", "golang.org/x/tools/godoc"},
	{"foo.bar/baz.quux", "foo.bar/baz%2equux"},
	{"", ""},
	{"%foo%bar", "%25foo%25bar"},
	{"\x01\x00\x7F☺", "%01%00%7f%e2%98%ba"},
}

func TestImportPathToPrefix(t *testing.T) {
	for _, tt := range importPathToPrefixTests {
		if out := importPathToPrefix(tt.in); out != tt.out {
			t.Errorf("importPathToPrefix(%q) = %q, want %q", tt.in, out, tt.out)
		}
	}
}

var runtimeA string

func init() {
	runtimeP, err := build.Default.Import("runtime", ".", build.ImportComment)
	if err != nil {
		log.Fatal(err)
	}
	runtimeA = runtimeP.PkgObj

}

func BenchmarkLoadRuntime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r, err := os.Open(runtimeA)
		if err != nil {
			b.Error(err)
		}
		_, err = Parse(r, "runtime")
		if err != nil {
			b.Error(err)
		}
	}
}
