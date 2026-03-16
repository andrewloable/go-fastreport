package functions_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/functions"
)

func TestNumToLetters(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
		{701, "ZZ"},
		{702, "AAA"},
		{-1, ""},
		{-100, ""},
	}
	for _, c := range cases {
		got := functions.NumToLetters(c.n)
		if got != c.want {
			t.Errorf("NumToLetters(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestNumToLettersLower(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "a"},
		{1, "b"},
		{25, "z"},
		{26, "aa"},
		{27, "ab"},
		{51, "az"},
		{52, "ba"},
		{701, "zz"},
		{702, "aaa"},
		{-1, ""},
		{-100, ""},
	}
	for _, c := range cases {
		got := functions.NumToLettersLower(c.n)
		if got != c.want {
			t.Errorf("NumToLettersLower(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}
