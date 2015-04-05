// run

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 10332: The PkgPath of unexported fields of types defined in
// package main was incorrectly ""

package main

import "reflect"

type foo struct {
	bar int
}

func main() {
	pkgpath := reflect.ValueOf(foo{}).Type().Field(0).PkgPath
	if pkgpath != "main" {
		println("BUG: incorrect PkgPath:", pkgpath)
	}
}
