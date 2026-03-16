package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

// TestToInt_PlainInt covers the case int: branch in ToInt (standard.go line 389-390).
// The existing TestToInt_AllTypes table omits plain int, leaving that case uncovered.
func TestToInt_PlainInt(t *testing.T) {
	cases := []struct {
		input int
		want  int
	}{
		{0, 0},
		{42, 42},
		{-7, -7},
		{1<<31 - 1, 1<<31 - 1}, // max int32 value, fits in int
	}
	for _, c := range cases {
		got := functions.ToInt(c.input)
		if got != c.want {
			t.Errorf("ToInt(int(%d)) = %d, want %d", c.input, got, c.want)
		}
	}
}
