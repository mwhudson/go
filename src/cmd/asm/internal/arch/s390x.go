// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file encapsulates some of the odd characteristics of the
// s390x instruction set, to minimize its interaction
// with the core of the assembler.

package arch

import "cmd/internal/obj/s390x"

func jumpS390x(word string) bool {
	switch word {
	case "BC",
		"BCL",
		"BEQ",
		"BGE",
		"BGT",
		"BL",
		"BLE",
		"BLT",
		"BNE",
		"BR",
		"BVC",
		"BVS",
		"CMPBEQ",
		"CMPBGE",
		"CMPBGT",
		"CMPBLE",
		"CMPBLT",
		"CMPBNE",
		"CMPUBEQ",
		"CMPUBGE",
		"CMPUBGT",
		"CMPUBLE",
		"CMPUBLT",
		"CMPUBNE",
		"CALL",
		"JMP":
		return true
	}
	return false
}

// Iss390xRLD reports whether the op (as defined by an s390x.A* constant) is
// one of the RLD-like instructions that require special handling.
// The FMADD-like instructions behave similarly.
func IsS390xRLD(op int) bool {
	switch op {
	case s390x.ARLDC, s390x.ARLDCR, s390x.ARLWMI, s390x.ARLWNM:
		return true
	case s390x.AFMADD,
		s390x.AFMADDS,
		s390x.AFMSUB,
		s390x.AFMSUBS,
		s390x.AFNMADD,
		s390x.AFNMADDS,
		s390x.AFNMSUB,
		s390x.AFNMSUBS:
		return true
	}
	return false
}

// Iss390xCMP reports whether the op (as defined by an s390x.A* constant) is
// one of the CMP instructions that require special handling.
func IsS390xCMP(op int) bool {
	switch op {
	case s390x.ACMP, s390x.ACMPU, s390x.ACMPW, s390x.ACMPWU:
		return true
	}
	return false
}

// Iss390xNEG reports whether the op (as defined by an s390x.A* constant) is
// one of the NEG-like instructions that require special handling.
func IsS390xNEG(op int) bool {
	switch op {
	case s390x.AADDME,
		s390x.AADDZE,
		s390x.ANEG,
		s390x.ASUBME,
		s390x.ASUBZE:
		return true
	}
	return false
}

// IsS390xStorageAndStorage reports whether the op (as defined by an s390x.A* constant) refers
// to an storage-and-storage format instruction such as mvc, clc, xc, oc or nc.
func IsS390xStorageAndStorage(op int) bool {
	switch op {
	case s390x.AMVC, s390x.ACLC, s390x.AXC, s390x.AOC, s390x.ANC:
		return true
	}
	return false
}

func s390xRegisterNumber(name string, n int16) (int16, bool) {
	switch name {
	case "AR":
		if 0 <= n && n <= 15 {
			return s390x.REG_AR0 + n, true
		}
	case "F":
		if 0 <= n && n <= 15 {
			return s390x.REG_F0 + n, true
		}
	case "R":
		if 0 <= n && n <= 15 {
			return s390x.REG_R0 + n, true
		}
	}
	return 0, false
}
