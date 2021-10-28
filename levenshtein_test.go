package acmd

import "testing"

func Test_strDistance(t *testing.T) {
	testCases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "a", 0},
		{"", "hello", 5},
		{"hello", "", 5},
		{"hello", "hello", 0},
		{"ab", "aa", 1},
		{"ab", "ba", 2},
		{"ab", "aaa", 2},
		{"bbb", "a", 3},
		{"kitten", "sitting", 3},
		{"distance", "difference", 5},
		{"resume and cafe", "resumes and cafes", 2},
	}

	for _, tc := range testCases {
		dist := strDistance(tc.a, tc.b)
		if dist != tc.want {
			t.Errorf("for (%q , %q) want %d, got %d", tc.a, tc.b, tc.want, dist)
		}
	}
}
