// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A function (that sets up a stack frame at all) compiled as PIC on ppc64
// uses a minimum of 32 bytes of stack as opposed to the minimum of 8 for
// non-PIC code. This mostly affects assembly that calls other functions with
// arguments (the arguments should be stored at FIXED_FRAME+0(R1),
// FIXED_FRAME+8(R1) etc) and some other low-level places. (PIC is not
// actually supported yet).

#ifdef GOARCH_ppc64
#define FIXED_FRAME 8
#endif

#ifdef GOARCH_ppc64le
#define FIXED_FRAME 8
#endif
