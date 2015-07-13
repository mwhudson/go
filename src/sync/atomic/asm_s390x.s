// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x

#include "textflag.h"

TEXT ·SwapInt32(SB),NOSPLIT,$0-20
	BR	·SwapUint32(SB)

TEXT ·SwapUint32(SB),NOSPLIT,$0-20
	MOVD	addr+0(FP), R3
	MOVW	new+8(FP), R4
repeat:	MOVW	(R3), R5
        CS      R5, R4, (R3)    // if (R3) == R5, then (R3) = R4 
        BNE     repeat	
	MOVW	R5, old+16(FP)
	RET

TEXT ·SwapInt64(SB),NOSPLIT,$0-24
	BR	·SwapUint64(SB)

TEXT ·SwapUint64(SB),NOSPLIT,$0-24
	MOVD	addr+0(FP), R3
	MOVD	new+8(FP), R4
repeat:	MOVD	(R3), R5
        CSG     R5, R4, (R3)    // if (R3) == R5, then (R3) = R4
	BNE     repeat	
	MOVD	R5, old+16(FP)
	RET

TEXT ·SwapUintptr(SB),NOSPLIT,$0-24
	BR	·SwapUint64(SB)

TEXT ·CompareAndSwapInt32(SB),NOSPLIT,$0-17
	BR	·CompareAndSwapUint32(SB)

TEXT ·CompareAndSwapUint32(SB),NOSPLIT,$0-17
	MOVD	ptr+0(FP), R3
	MOVWZ	old+8(FP), R4
	MOVWZ	new+12(FP), R5
	CS	R4, R5, 0(R3)    //  if (R4 == 0(R3)) then 0(R3)= R5
	BNE	cas_fail
	MOVD	$1, R3
	MOVB	R3, ret+16(FP)
	RET
cas_fail:
	MOVD	$0, R3
	MOVB	R3, ret+16(FP)
	RET

TEXT ·CompareAndSwapUintptr(SB),NOSPLIT,$0-25
	BR	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapInt64(SB),NOSPLIT,$0-25
	BR	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapUint64(SB),NOSPLIT,$0-25
	MOVD	ptr+0(FP), R3
	MOVD	old+8(FP), R4
	MOVD	new+16(FP), R5
	CSG	R4, R5, 0(R3)    //  if (R4 == 0(R3)) then 0(R3)= R5
	BNE	cas64_fail
	MOVD	$1, R3
	MOVB	R3, ret+24(FP)
	RET
cas64_fail:
	MOVD	$0, R3
	MOVB	R3, ret+24(FP)
	RET

TEXT ·AddInt32(SB),NOSPLIT,$0-20
	BR	·AddUint32(SB)

TEXT ·AddUint32(SB),NOSPLIT,$0-20
	MOVD	ptr+0(FP), R4
	MOVW	delta+8(FP), R5
repeat:
	MOVW	(R4), R3
	MOVD	R3, R6
	ADD	R5, R3
	CS	R6, R3, (R4)    // if (R6==(R4)) then (R4)=R3
	BNE	repeat
	MOVW	R3, ret+16(FP)
	RET

TEXT ·AddUintptr(SB),NOSPLIT,$0-24
	BR	·AddUint64(SB)

TEXT ·AddInt64(SB),NOSPLIT,$0-24
	BR	·AddUint64(SB)

TEXT ·AddUint64(SB),NOSPLIT,$0-24
	MOVD	ptr+0(FP), R4
	MOVD	delta+8(FP), R5
repeat:
	MOVD	(R4), R3
	MOVD	R3, R6
	ADD	R5, R3
	CSG	R6, R3, (R4)    // if (R6==(R4)) then (R4)=R3
	BNE	repeat
	MOVD	R3, ret+16(FP)
	RET

TEXT ·LoadInt32(SB),NOSPLIT,$0-12
	BR	·LoadUint32(SB)

TEXT ·LoadUint32(SB),NOSPLIT,$0-12
	MOVD	addr+0(FP), R3
	MOVW	0(R3), R3
	MOVW	R3, val+8(FP)
	RET

TEXT ·LoadInt64(SB),NOSPLIT,$0-16
	BR	·LoadUint64(SB)

TEXT ·LoadUint64(SB),NOSPLIT,$0-16
	MOVD	addr+0(FP), R3
	MOVD	0(R3), R3
	MOVD	R3, val+8(FP)
	RET

TEXT ·LoadUintptr(SB),NOSPLIT,$0-16
	BR	·LoadPointer(SB)

TEXT ·LoadPointer(SB),NOSPLIT,$0-16
	BR	·LoadUint64(SB)

TEXT ·StoreInt32(SB),NOSPLIT,$0-12
	BR	·StoreUint32(SB)

TEXT ·StoreUint32(SB),NOSPLIT,$0-12
        MOVD    ptr+0(FP), R3
        MOVW    val+8(FP), R4
        SYNC
        MOVW    R4, 0(R3)
        RET

TEXT ·StoreInt64(SB),NOSPLIT,$0-16
	BR	·StoreUint64(SB)

TEXT ·StoreUint64(SB),NOSPLIT,$0-16
	MOVD	addr+0(FP), R3
	MOVD	val+8(FP), R4
        SYNC
	MOVD	R4, 0(R3)
	RET

TEXT ·StoreUintptr(SB),NOSPLIT,$0-16
	BR	·StoreUint64(SB)
