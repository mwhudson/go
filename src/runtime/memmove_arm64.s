/* Derived from the implementation of memmove in cortex-strings

    https://git.linaro.org/toolchain/cortex-strings.git/blob/HEAD:/src/aarch64/memmove.S
    https://git.linaro.org/toolchain/cortex-strings.git/blob/HEAD:/src/aarch64/memcpy.S

 which is under a modified BSD license:

 Copyright (c) 2013, Linaro Limited
   All rights reserved.

   Redistribution and use in source and binary forms, with or without
   modification, are permitted provided that the following conditions are met:
       * Redistributions of source code must retain the above copyright
         notice, this list of conditions and the following disclaimer.
       * Redistributions in binary form must reproduce the above copyright
         notice, this list of conditions and the following disclaimer in the
         documentation and/or other materials provided with the distribution.
       * Neither the name of the Linaro nor the
         names of its contributors may be used to endorse or promote products
         derived from this software without specific prior written permission.

   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
   "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
   LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
   A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
   HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
   SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
   LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
   DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
   THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
   (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
   OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE. */

#include "textflag.h"

dstin = 0
src = 1
count = 2
tmp1 = 3
tmp2 = 4
dst  = 6

A_l = 7
A_h = 8
B_l = 9
B_h = 10
C_l = 11
C_h = 12
D_l = 13
D_h = 14


TEXT runtime·memmove(SB), NOSPLIT, $4
_memmove:
        // I guess I need to add pretend moves to get the arguments to where they already are?
        CMP     R(dstin), R(src)
        BLO    _downwards
        ADD	R(src), R(count), R(tmp1)

	CMP	R(dstin), R(tmp1)
	BHS	_memcpy		/* No overlap.  */

	/* Upwards move with potential overlap.
	 * Need to move from the tail backwards.  SRC and DST point one
	 * byte beyond the remaining data to move.  */
	ADD	R(dstin), R(count), R(dst)
	ADD	R(count), R(src), R(src)
	CMP	$64, R(count)
        BLT	_mov_not_short_up
	/* Deal with small moves quickly by dropping straight into the
	 * exit block.  */
_tail63up:
	/* Move up to 48 bytes of data.  At this point we only need the
	 * bottom 6 bits of count to be accurate.  */
	ANDS	$0x30, R(count), R(tmp1)
	BEQ	_tail15up
	SUB	R(dst), R(tmp1), R(dst)
	SUB	R(src), R(tmp1), R(src)
	CMPW	$0x20, R(tmp1) // this seems backwards
	BEQ	_tail47up
	BGE	_tail31up

	MOVP	32(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h), 32(R(src))
_tail47up:
	MOVP	16(src), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) 16(dst)
_tail31up:
	MOVP	(src), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) (dst)
_tail15up:
	/* Move up to 15 bytes of data.  Does not assume additional data
	 * being moved.  */
	TBZ	count, #3, _tail7up
	LDR	R(tmp1), [R(src), #-8]!
	STR	R(tmp1), [R(dst), #-8]!
_tail7up:
	TBZ	count, #2, _tail3up
	LDR	W(tmp1), [R(src), #-4]!
	STR	W(tmp1), [R(dst), #-4]!
_tail3up:
	TBZ	count, #1, _tail1up
	LDRH	W(tmp1), [R(src), #-2]!
	STRH	W(tmp1), [R(dst), #-2]!
_tail1up:
	TBZ	count, #0, _tail0up
	LDRB	W(tmp1), [R(src), #-1]
	STRB	W(tmp1), [R(dst), #-1]
_tail0up:
	RET

_mov_not_short_up:
	/* We don't much care about the alignment of DST, but we want SRC
	 * to be 128-bit (16 byte) aligned so that we don't cross cache line
	 * boundaries on both loads and stores.  */
	ANDS	R(src), #15, R(tmp2)		/* Bytes to reach alignment.  */
	BEQ	_mov_not_short_up_aligned
	SUB	R(count), R(tmp2), R(count)
	/* Move enough data to reach alignment; unlike memcpy, we have to
	 * be aware of the overlap, which means we can't move data twice.  */
	TBZ	R(tmp2), #3, _mov_not_short_up_aligned7
	LDR	R(tmp1), [R(src), #-8]!
	STR	R(tmp1), [R(dst), #-8]!
_mov_not_short_up_aligned7:
	TBZ	R(tmp2), #2, _mov_not_short_up_aligned3
	LDR	W(tmp1), [R(src), #-4]!
	STR	W(tmp1), [R(dst), #-4]!
_mov_not_short_up_aligned3:
	TBZ	R(tmp2), #1, _mov_not_short_up_aligned1
	LDRH	W(tmp1), [R(src), #-2]!
	STRH	W(tmp1), [R(dst), #-2]!
_mov_not_short_up_aligned1:
	TBZ	R(tmp2), #0, _mov_not_short_up_aligned0
	LDRB	W(tmp1), [R(src), #-1]!
	STRB	W(tmp1), [R(dst), #-1]!
_mov_not_short_up_aligned0:

	/* There may be less than 63 bytes to go now.  */
	CMP     R(count), #63
	BLE	_tail63up
_mov_not_short_up_aligned:
	SUBS	R(count), #128, R(count)
	BGE	_mov_body_large_up
	/* Less than 128 bytes to move, so handle 64 here and then jump
	 * to the tail.  */
	MOVP	-64!(R(src)), R(A_l), R(A_h)
	MOVP	16(R(src)), R(B_l), R(B_h)
	MOVP	32(R(src)), R(C_l), R(C_h)
	MOVP	48(R(src)), R(D_l), R(D_h)
	MOVP	R(A_l), R(A_h) -64!(R(dst))
	MOVP	R(B_l), R(B_h) 16(R(dst))
	MOVP	R(C_l), R(C_h) 32(R(dst))
	MOVP	R(D_l), R(D_h) 48(R(dst))
	TST     R(count), #0x3f
	BNE	_tail63up
	RET

	/* Critical loop.  Start at a new Icache line boundary.  Assuming
	 * 64 bytes per line this ensures the entire loop is in one line.  */
//	.p2align 6 // don't think go assemblers support this?
_mov_body_large_up:
	/* There are at least 128 bytes to move.  */
	MOVP	-16(R(src)), R(A_l), R(A_h)
	MOVP	-32(R(src)), R(B_l), R(B_h)
	MOVP	-48(R(src)), R(C_l), R(C_h)
	MOVP	-64!(R(src)), R(D_l), R(D_h)
_mov_body_large_up_loop:
	MOVP	R(A_l), R(A_h) -16(R(dst))
	MOVP	-16(R(src)), R(A_l), R(A_h)
	MOVP	R(B_l), R(B_h) -32(R(dst))
	MOVP	-32(R(src)), R(B_l), R(B_h)
	MOVP	R(C_l), R(C_h) -48(R(dst))
	MOVP	-48(R(src)), R(C_l), R(C_h)
	MOVP	R(D_l), R(D_h) -64!(R(dst))
	MOVP	-64!(R(src)), R(D_l), R(D_h)
	SUBS	R(count), #64, R(count)
	BGE	_mov_body_large_up_loop
	MOVP	R(A_l), R(A_h) -16(R(dst))
	MOVP	R(B_l), R(B_h) -32(R(dst))
	MOVP	R(C_l), R(C_h) -48(R(dst))
	MOVP	R(D_l), R(D_h) -64!(R(dst))
	TST     R(count), #0x3f
	BNE	_tail63up
	RET

_downwards:
	/* For a downwards move we can safely use memcpy provided that
	 * R(DST) is more than 16 bytes away from R(SRC).  */
	SUB     R(src), #16, R(tmp1)
	CMP     R(dstin), R(tmp1)
	BLS	_memcpy		/* May overlap, but not critically.  */

	MOV     R(dstin), R(dst)	/* Preserve R(DSTIN) for return value.  */
	CMP     R(count), #64
	BGE	_mov_not_short_down

	/* Deal with small moves quickly by dropping straight into the
	 * exit block.  */
_tail63down:
	/* Move up to 48 bytes of data.  At this point we only need the
	 * bottom 6 bits of R(count) to be accurate.  */
	ANDS	R(count), #0x30, R(tmp1)
	BEQ	_tail15down
	ADD	R(dst), tmp1, R(dst)
	ADD	R(src), tmp1, R(src)
	CMP	W(tmp1), #0x20
	BEQ	_tail47down
	BLT	_tail31down
	MOVP	-48(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -48(R(dst))
_tail47down:
	MOVP	-32(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -32(R(dst))
_tail31down:
	MOVP	-16(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -16(R(dst))
_tail15down:
	/* Move up to 15 bytes of data.  Does not assume additional data
	   being moved.  */
	TBZ     R(count), #3, _tail7down
	LDR     tmp1, [R(src)], #8
	STR     tmp1, [R(dst)], #8
_tail7down:
	TBZ     R(count), #2, _tail3down
	LDR     W(tmp1), [R(src)], #4
	STR     W(tmp1), [R(dst)], #4
_tail3down:
	TBZ     R(count), #1, _tail1down
	LDRH	W(tmp1), [R(src)], #2
	STRH	W(tmp1), [R(dst)], #2
_tail1down:
	TBZ     R(count), #0, _tail0down
	LDRB	W(tmp1), [R(src)]
	STRB	W(tmp1), [R(dst)]
_tail0down:
	RET

_mov_not_short_down:
	/* We don't much care about the alignment of R(DST), but we want R(SRC)
	 * to be 128-bit (16 byte) aligned so that we don't cross cache line
	 * boundaries on both loads and stores.  */
	NEG	R(src), R(tmp2)
	ANDS	R(tmp2), #15, R(tmp2)		/* Bytes to reach alignment.  */
	BEQ	_mov_not_short_down_aligned
	SUB	R(count), R(tmp2), R(count)
	/* Move enough data to reach alignment; unlike memcpy, we have to
	 * be aware of the overlap, which means we can't move data twice.  */
	TBZ	R(tmp2), #3, _mov_not_short_down_align7
	LDR	R(tmp1), [R(src)], #8
	STR	R(tmp1), [R(dst)], #8
_mov_not_short_down_align7:
	TBZ	R(tmp2), #2, _mov_not_short_down_align3
	LDR	W(tmp1), [R(src)], #4
	STR	W(tmp1), [R(dst)], #4
_mov_not_short_down_align3:
	TBZ	R(tmp2), #1, _mov_not_short_down_align1
	LDRH	W(tmp1), [R(src)], #2
	STRH	W(tmp1), [R(dst)], #2
_mov_not_short_down_align1:
	TBZ	R(tmp2), #0, _mov_not_short_down_align0
	LDRB	W(tmp1), [R(src)], #1
	STRB	W(tmp1), [R(dst)], #1
_mov_not_short_down_align0:

	/* There may be less than 63 bytes to go now.  */
	CMP	R(count), #63
	BLE	_tail63down
_mov_not_short_down_aligned:
	SUBS	R(count), R(count), #128
	BGE	_mov_body_large_down
	/* Less than 128 bytes to move, so handle 64 here and then jump
	 * to the tail.  */
	MOVP	(R(src)), R(A_l), R(A_h)
	MOVP	16(R(src)), R(B_l), R(B_h)
	MOVP	32(R(src)), R(C_l), R(C_h)
	MOVP	48(R(src)), R(D_l), R(D_h)
	MOVP	R(A_l), R(A_h) (R(dst))
	MOVP	R(B_l), R(B_h) 16(R(dst))
	MOVP	R(C_l), R(C_h) 32(R(dst))
	MOVP	R(D_l), R(D_h) 48(R(dst))
	TST     R(count), #0x3f
	ADD     R(src), #64, R(src)
	ADD     R(dst), #64, R(dst)
	BNE	_tail63down
	RET

	/* Critical loop.  Start at a new cache line boundary.  Assuming
	 * 64 bytes per line this ensures the entire loop is in one line.  */
//	.p2align 6 // don't think go assemblers support this?
_mov_body_large_down:
	/* There are at least 128 bytes to move.  */
	MOVP	0(R(src)), R(A_l), R(A_h)
	SUB     R(dst), #16, R(dst)		/* Pre-bias.  */
	MOVP	16(R(src)), R(B_l), R(B_h)
	MOVP	32(R(src)), R(C_l), R(C_h)
	MOVP	48!(R(src)), R(D_l), R(D_h)
_mov_body_large_down_loop:
	MOVP	R(A_l), R(A_h) 16(R(dst))
	MOVP	16(R(src)), R(A_l), R(A_h)
	MOVP	R(B_l), R(B_h) 32(R(dst))
	MOVP	32(R(src)), R(B_l), R(B_h)
	MOVP	R(C_l), R(C_h) 48(R(dst))
	MOVP	48(R(src)), R(C_l), R(C_h)
	MOVP	R(D_l), R(D_h) 64!(R(dst))
	MOVP	64!(R(src)), R(D_l), R(D_h)
	SUBS	R(count), #64, R(count)
	BGE	_mov_body_large_down_loop
	MOVP	R(A_l), R(A_h) 16(R(dst))
	MOVP	R(B_l), R(B_h) 32(R(dst))
	MOVP	R(C_l), R(C_h) 48(R(dst))
	MOVP	R(D_l), R(D_h) 64(R(dst))
	ADD     R(src), #16, R(src)
	ADD     R(dst), #64 + 16, R(dst)
	TST     R(count), #0x3f
	BNE	_tail63down
	RET

_memcpy:
        MOV     R(dstin), R(dst)
	CMP     R(count), #64
	BGE	_cpy_not_short
	CMP     R(count), #15
	BLE	_tail15tiny

	/* Deal with small copies quickly by dropping straight into the
	 * exit block.  */
_tail63:
	/* Copy up to 48 bytes of data.  At this point we only need the
	 * bottom 6 bits of count to be accurate.  */
	ANDS	R(count), #0x30, tmp1
	BEQ	_tail15
	ADD	R(dst), R(tmp1), R(dst)
	ADD	R(src), R(tmp1), R(src)
	CMP	W(tmp1), #0x20
	BEQ	_tail47
	BLT	_tail31
	MOVP	-48(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -48(R(dst))
_tail47:
	MOVP	-32(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -32(R(dst))
_tail31:
	MOVP	-16(R(src)), R(A_l), R(A_h)
	MOVP	R(A_l), R(A_h) -16(R(dst))

_tail15:
	ANDS	R(count), #15, R(count)
	BEQ	_tail0
	ADD	R(src), R(count), R(src)
	MOVP	-16(R(src)), R(A_l), R(A_h)
	ADD	R(dst), R(count), R(dst)
	MOVP	R(A_l), R(A_h) -16(R(dst))
_tail0:
	RET

_tail15tiny:
	/* Copy up to 15 bytes of data.  Does not assume additional data
	   being copied.  */
	TBZ	R(count), #3, _tail15tiny7
	LDR	R(tmp1), [R(src)], #8
	STR	R(tmp1), [R(dst)], #8
_tail15tiny7:
	TBZ	R(count), #2, _tail15tiny3
	LDR	W(tmp1), [R(src)], #4
	STR	W(tmp1), [R(dst)], #4
_tail15tiny3:
	TBZ     R(count), #1, _tail15tiny1
	LDRH	W(tmp1), [R(src)], #2
	STRH	W(tmp1), [R(dst)], #2
_tail15tiny1:
	TBZ     R(count), #0, _tail15tiny0
	LDRB	W(tmp1), [R(src)]
	STRB	W(tmp1), [R(dst)]
_tail15tiny0:
	RET

_cpy_not_short:
	/* We don't much care about the alignment of DST, but we want SRC
	 * to be 128-bit (16 byte) aligned so that we don't cross cache line
	 * boundaries on both loads and stores.  */
	NEG	R(src), R(tmp2)
	ANDS	R(tmp2), #15, R(tmp2)		/* Bytes to reach alignment.  */
	BEQ	_cpy_not_short_aligned
	SUB	R(count), R(tmp2), R(count)
	/* Copy more data than needed; it's faster than jumping
	 * around copying sub-Quadword quantities.  We know that
	 * it can't overrun.  */
	MOVP	(R(src)), R(A_l), R(A_h)
	ADD	R(src), R(tmp2), R(src)
	MOVP	R(A_l), R(A_h) (R(dst))
	ADD	R(dst), R(tmp2), R(dst)
	/* There may be less than 63 bytes to go now.  */
	CMP     R(count), #63
	BLE	_tail63
_cpy_not_short_aligned:
	SUBS	R(count), #128, R(count)
	BGE	_cpy_body_large
	/* Less than 128 bytes to copy, so handle 64 here and then jump
	 * to the tail.  */
	MOVP	(R(src)), R(A_l), R(A_h)
	MOVP	16(R(src)), R(B_l), R(B_h)
	MOVP	32(R(src)), R(C_l), R(C_h)
	MOVP	48(R(src)), R(D_l), R(D_h)
	MOVP	R(A_l), R(A_h) (R(dst))
	MOVP	R(B_l), R(B_h) 16(R(dst))
	MOVP	R(C_l), R(C_h) 32(R(dst))
	MOVP	R(D_l), R(D_h) 48(R(dst))
	TST	R(count), #0x3f
	ADD	R(src), #64, R(src)
	ADD	R(dst), #64, R(dst)
	BNE	_tail63
	RET

	/* Critical loop.  Start at a new cache line boundary.  Assuming
	 * 64 bytes per line this ensures the entire loop is in one line.  */
//	.p2align 6
_cpy_body_large:
	/* There are at least 128 bytes to copy.  */
	MOVP	0(R(src)), R(A_l), A_h
	SUB	R(dst), #16, R(dst)		/* Pre-bias.  */
	MOVP	16(R(src)), R(B_l), R(B_h)
	MOVP	32(R(src)), R(C_l), R(C_h)
	MOVP	48!(R(src)), R(D_l), R(D_h)
_cpy_body_large_loop:
	MOVP	R(A_l), R(A_h) 16(R(dst))
	MOVP	16(R(src)), R(A_l), R(A_h)
	MOVP	R(B_l), R(B_h) 32(R(dst))
	MOVP	32(R(src)), R(B_l), R(B_h)
	MOVP	R(C_l), R(C_h) 48(R(dst))
	MOVP	48(R(src)), R(C_l), R(C_h)
	MOVP	R(D_l), R(D_h) 64!(R(dst))
	MOVP	64!(R(src)), R(D_l), R(D_h)
	SUBS	R(count), #64, R(count)
	BGE	_cpy_body_large_loop
	MOVP	R(A_l, R(A_h) 16(R(dst))
	MOVP	R(B_l, R(B_h) 32(R(dst))
	MOVP	R(C_l, R(C_h) 48(R(dst))
	MOVP	R(D_l, R(D_h) 64(R(dst))
	ADD     R(src), #16, R(src)
	ADD     R(dst), #64 + 16, R(src)
	TST     R(count), #0x3f
	BNE	_tail63
	RET
