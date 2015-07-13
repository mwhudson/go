// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// We have to resort to TLS variable to save g (R13).
// One reason is that external code might trigger
// SIGSEGV, and our runtime.sigtramp don't even know we
// are in external code, and will continue to use R13,
// this might well result in another SIGSEGV.

// save_g saves the g register into pthread-provided
// thread-local memory, so that we can call externally compiled
// s390x code that will overwrite this register.
//
// If !iscgo, this is a no-op.
//
// NOTE: setg_gcc<> assume this clobbers only R10 and R11.
TEXT runtime·save_g(SB),NOSPLIT,$-8-0
	MOVB	runtime·iscgo(SB),  R10
	CMPBEQ	R10, $0, nocgo

	// Rematerialize the C TLS base pointer from AR0:AR1;
	// "MOVW ARx, Rx" is translated to EAR.
	MOVW	AR0, R11
	SLD	$32, R11
	MOVW	AR1, R11

	// $runtime.tlsg(SB) is a special linker symbol.
	// It is the offset from the start of TLS to our
	// thread-local storage for g.
	// Note: on s390x the offset should be less than 0
	MOVD	$runtime·tlsg(SB), R10
	ADD	R11, R10

	// Store g in TLS
	MOVD	g, 0(R10)

nocgo:
	RET

// load_g loads the g register from pthread-provided
// thread-local memory, for use after calling externally compiled
// s390x code that overwrote those registers.
//
// This is never called directly from C code (it doesn't have to
// follow the C ABI), but it may be called from a C context, where the
// usual Go registers aren't set up.
//
// NOTE: _cgo_topofstack assumes this only clobbers g (R13), R10 and R11.
TEXT runtime·load_g(SB),NOSPLIT,$-8-0
	MOVW	AR0, R11
	SLD	$32, R11
	MOVW	AR1, R11

	MOVD	$runtime·tlsg(SB), R10
	ADD	R11, R10

	// Load g from TLS
	MOVD	0(R10), g
	RET

GLOBL runtime·tlsg_offset+0(SB), RODATA, $8
