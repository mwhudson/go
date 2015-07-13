/*
 * void crosscall_s390x(void (*fn)(void), void *g)
 *
 * Calling into the go tool chain, where all registers are caller save.
 * Called from standard s390x C ABI, where r6-r13, r15, and f0, f2, f4 and f6 are
 * callee-save, so they must be saved explicitly.
 */
.globl crosscall_s390x
crosscall_s390x:
	/*
	 * save r6-r15, f0, f2, f4 and f6 in the
	 * register save area of the calling function
	 */
	stmg	%r6, %r15, 48(%r15)
	stdy	%f0, 128(%r15)
	stdy	%f2, 136(%r15)
	stdy	%f4, 144(%r15)
	stdy	%f6, 152(%r15)

	/* assume these calls do not clobber r2 or r15 */
	brasl   %r14, _cgo_reginit
	brasl   %r14, _cgo_load_g

	/* grow stack 8 bytes and call fn */
	agfi    %r15, -8
	basr    %r14, %r2
	agfi	%r15, 8

	/* restore registers */
	lmg	%r6, %r15, 48(%r15)
	ldy	%f0, 128(%r15)
	ldy	%f2, 136(%r15)
	ldy	%f4, 144(%r15)
	ldy	%f6, 152(%r15)
	
	br      %r14 /* restored by lmg */

.globl __stack_chk_fail_local
__stack_chk_fail_local:
1:
	larl %r10,1b
	br %r10

#ifdef __ELF__
.section .note.GNU-stack,"",%progbits
#endif
