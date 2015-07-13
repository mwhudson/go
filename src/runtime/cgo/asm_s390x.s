// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x 

#include "textflag.h"

/*
 * void crosscall2(void (*fn)(void*, int32), void*, int32)
 * Save registers and call fn with two arguments.
 * crosscall2 obeys the C ABI; fn obeys the Go ABI.
 */
TEXT crosscall2(SB),NOSPLIT,$-8
	// Start with standard C stack frame layout and linkage

	// Save R6-R15, F0, F2, F4 and F6 in the
	// register save area of the calling function
	// stmg %r6, %r15, 48(%r15)
	BYTE	$0xeb;
	BYTE	$0x6f;
	BYTE	$0xf0;
	BYTE	$0x30;
	BYTE	$0x00;
	BYTE	$0x24;
	FMOVD	F0, 128(R15)
	FMOVD	F2, 136(R15)
	FMOVD	F4, 144(R15)
	FMOVD	F6, 152(R15)

	// Initialize Go ABI environment
	BL	runtime·reginit(SB)
	BL	runtime·load_g(SB)

	// Allocate 24 bytes on the stack
	SUB	$24, R15

	MOVD	R3, 8(R15)  // arg1
	MOVW	R4, 16(R15) // arg2
	BL	(R2)        // fn(arg1, arg2)

	ADD	$24, R15

	// Restore R6-R15, F0, F2, F4 and F6
	// lmg %r6, %r15, 48(%r15)
	BYTE	$0xeb;
	BYTE	$0x6f;
	BYTE	$0xf0;
	BYTE	$0x30;
	BYTE	$0x00;
	BYTE	$0x04;
	FMOVD	F0, 128(R15)
	FMOVD	F2, 136(R15)
	FMOVD	F4, 144(R15)
	FMOVD	F6, 152(R15)

	RET

