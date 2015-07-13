// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

const (
	TheChar       = 'z'
	BigEndian     = 1
	CacheLineSize = 256
	PhysPageSize  = 4096
	PCQuantum     = 2
	Int64Align    = 8
	HugePageSize  = 0
	MinFrameSize  = 8 // TODO(mundaym): Not sure if this is correct.
)

type Uintreg uint64
