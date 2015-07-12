package main

import "zvtlib"

func main() {
	x := zvtlib.TortureZeroValue(1)
	if zvtlib.TortureZeroValue(1) != 0 {
		println(x)
		panic("boom")
	}
}
