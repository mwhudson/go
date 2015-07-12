package main

import "zvtlib"

// The point of this is that a map miss might return a zerovalue that
// is too short when a type is not seen going into a map when
// compiling the module that defines it. Large looks like this:
// type Large struct {
//        filler [256]byte
//        Content [1024]int
// }
// so we scan Content to see if anything is non-zero.
func TortureZeroValue(i int) int {
	if i > 0 {
		// This is just to inhibit inlining.
		return TortureZeroValue(i - 1)
	}
	m := make(map[int]zvtlib.Large)
	for _, j := range m[i].Content {
		if j != 0 {
			return j
		}
	}
	return 0
}

func main() {
	x := TortureZeroValue(1)
	if x != 0 {
		println(x)
		panic("boom")
	}
}
