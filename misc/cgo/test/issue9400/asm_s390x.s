// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x
// +build !gccgo

#include "textflag.h"

TEXT ·RewindAndSetgid(SB),NOSPLIT,$0-0
	// Rewind stack pointer so anything that happens on the stack
	// will clobber the test pattern created by the caller
	// TODO(mundaym): Code generator should support ADD $(1024 * 8), R15
	MOVD	$(1024 * 8), R4
	ADD	R4, R15

	// Ask signaller to setgid
	// TODO(mundaym): Code generator should support MOVW $1, ·Baton(SB)
	MOVD	$·Baton(SB), R5
	MOVW	$1, R6
	MOVW	R6, 0(R5)

	// Wait for setgid completion
loop:
	SYNC
	MOVW	·Baton(SB), R3
	CMP	R3, $0
	BNE	loop

	// Restore stack
	SUB	R4, R15
	RET
