#include "textflag.h"

TEXT _rt0_s390x_linux(SB),NOSPLIT,$-8
	// In a statically linked binary, the stack contains argc,
	// argv as argc string pointers followed by a NULL, envv as a
	// sequence of string pointers followed by a NULL, and auxv.
	// There is no TLS base pointer.
	//
	// TODO: Support dynamic linking entry point
	MOVD 0(R15), R2 // argc
	ADD $8, R15, R3 // argv
	BR main(SB)

TEXT main(SB),NOSPLIT,$-8
	MOVD	$runtime·rt0_go(SB), R11
	BR	R11
