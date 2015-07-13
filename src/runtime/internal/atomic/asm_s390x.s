// Copyright 20145 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x

#include "textflag.h"

// bool runtime/internal/atomic·Cas(uint32 *ptr, uint32 old, uint32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕atomic·Cas(SB), NOSPLIT, $0-17
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

// bool	runtime∕internal∕atomic·Cas64(uint64 *ptr, uint64 old, uint64 new)
// Atomically:
//	if(*val == *old){
//		*val = new;
//		return 1;
//	} else {
//		return 0;
//	}
TEXT runtime∕internal∕atomic·Cas64(SB), NOSPLIT, $0-25
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

TEXT runtime∕internal∕atomic·Casuintptr(SB), NOSPLIT, $0-25
	BR	runtime∕internal∕atomic·Cas64(SB)

TEXT runtime∕internal∕atomic·Loaduintptr(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Load64(SB)

TEXT runtime∕internal∕atomic·Loaduint(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Load64(SB)

TEXT runtime∕internal∕atomic·Storeuintptr(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Store64(SB)

TEXT runtime∕internal∕atomic·Loadint64(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Load64(SB)

TEXT runtime∕internal∕atomic·Xadduintptr(SB), NOSPLIT, $0-24
	BR	runtime∕internal∕atomic·Xadd64(SB)

TEXT runtime∕internal∕atomic·Xaddint64(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Xadd64(SB)

// bool runtime∕internal∕atomic·Casp1(void **val, void *old, void *new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime∕internal∕atomic·Casp1(SB), NOSPLIT, $0-25
	BR runtime∕internal∕atomic·Cas64(SB)

// bool casp(void **val, void *old, void *new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT runtime·casp1(SB), NOSPLIT, $-8-25
	BR runtime·cas64(SB)

// uint32 runtime∕internal∕atomic·Xadd(uint32 volatile *ptr, int32 delta)
// Atomically:
//	*val += delta;
//	return *val;
TEXT runtime∕internal∕atomic·Xadd(SB), NOSPLIT, $0-20
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

TEXT runtime∕internal∕atomic·Xadd64(SB), NOSPLIT, $0-24
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

TEXT runtime∕internal∕atomic·Xchg(SB), NOSPLIT, $0-20
	MOVD	ptr+0(FP), R4
	MOVW	new+8(FP), R3
repeat:
	MOVW	(R4), R6
	CS	R6, R3, (R4)    // if (R6==(R4)) then (R4)=R3
	BNE	repeat
	MOVW	R6, ret+16(FP)
	RET

TEXT runtime∕internal∕atomic·Xchg64(SB), NOSPLIT, $0-24
	MOVD	ptr+0(FP), R4
	MOVD	new+8(FP), R3
repeat:
	MOVD	(R4), R6
	CSG	R6, R3, (R4)    // if (R6==(R4)) then (R4)=R3
	BNE	repeat
	MOVD	R6, ret+16(FP)
	RET

TEXT runtime∕internal∕atomic·Xchguintptr(SB), NOSPLIT, $0-24
	BR	runtime∕internal∕atomic·Xchg64(SB)

TEXT runtime∕internal∕atomic·Storep1(SB), NOSPLIT, $0-16
	BR	runtime∕internal∕atomic·Store64(SB)

// on Z, load & store both are atomic operations
TEXT runtime∕internal∕atomic·Store(SB), NOSPLIT, $0-12
	MOVD	ptr+0(FP), R3
	MOVW	val+8(FP), R4
	SYNC
	MOVW	R4, 0(R3)
	RET

TEXT runtime∕internal∕atomic·Store64(SB), NOSPLIT, $0-16
	MOVD	ptr+0(FP), R3
	MOVD	val+8(FP), R4
	SYNC
	MOVD	R4, 0(R3)
	RET
