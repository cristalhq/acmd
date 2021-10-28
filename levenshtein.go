package acmd

// strDistance between 2 strings using Levenshtein distance algorithm.
func strDistance(a, b string) int {
	switch {
	case a == "":
		return len(b)
	case b == "":
		return len(a)
	case a == b:
		return 0
	}

	if len(a) > len(b) {
		a, b = b, a
	}
	lenA, lenB := len(a), len(b)

	x := make([]int, lenA+1)
	for i := 0; i < len(x); i++ {
		x[i] = i
	}

	for i := 1; i <= lenB; i++ {
		prev := i
		for j := 1; j <= lenA; j++ {
			current := x[j-1] // match
			if b[i-1] != a[j-1] {
				current = min3(x[j-1]+1, prev+1, x[j]+1)
			}
			x[j-1], prev = prev, current
		}
		x[lenA] = prev
	}
	return x[lenA]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	}
	if b < c {
		return b
	}
	return c
}
