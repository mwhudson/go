// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// FIXED_FRAME defines the size of the fixed part of a stack frame. A stack
// frame looks like this:
//
// +---------------------+
// | local variable area |
// +---------------------+
// | argument area       |
// +---------------------+ <- R1+FIXED_FRAME
// | fixed area          |
// +---------------------+ <- R1
//
// So a function that sets up a stack frame at all uses as least FIXED_FRAME
// bytes of stack.  This mostly affects assembly that calls other functions
// with arguments (the arguments should be stored at FIXED_FRAME+0(R1),
// FIXED_FRAME+8(R1) etc) and some other low-level places.
//
// The reason for using a constant is when code is compiled as PIC on ppc64le
// the fixed part of the stack is 32 bytes large.

// MAYBE_RELOAD_TOC expands to nothing usually, but restores the TOC pointer
// from 24(R1) when the code is compiled as PIC. Code needs to use this when
// calling a function via a function pointer that (a) returns normally and (b)
// may be implemented in a different module.

#ifdef GOARCH_ppc64
#define FIXED_FRAME 8
#define MAYBE_RELOAD_TOC
#endif

#ifdef GOARCH_ppc64le
#define FIXED_FRAME 32
#define MAYBE_RELOAD_TOC MOVD 24(R1), R2
#endif
