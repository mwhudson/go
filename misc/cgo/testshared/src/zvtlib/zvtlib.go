package zvtlib

type Large struct {
	filler  [256]byte
	content [1024]int
}

func TortureZeroValue(i int) int {
	if i > 0 {
		// This is just to inhibit inlining.
		return TortureZeroValue(i - 1)
	}
	m := make(map[int]Large)
	for _, j := range m[i].content {
		if j != 0 {
			return j
		}
	}
	return 0
}
