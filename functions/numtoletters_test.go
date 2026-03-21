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

func TestToLettersEn(t *testing.T) {
	cases := []struct {
		n       int
		isUpper bool
		want    string
	}{
		{0, true, "A"},
		{0, false, "a"},
		{1, true, "B"},
		{1, false, "b"},
		{25, true, "Z"},
		{25, false, "z"},
		{26, true, "AA"},
		{26, false, "aa"},
		{27, true, "AB"},
		{51, true, "AZ"},
		{52, true, "BA"},
		{701, true, "ZZ"},
		{702, true, "AAA"},
		{-1, true, ""},
		{-100, false, ""},
	}
	for _, c := range cases {
		got := functions.ToLettersEn(c.n, c.isUpper)
		if got != c.want {
			t.Errorf("ToLettersEn(%d, %v) = %q, want %q", c.n, c.isUpper, got, c.want)
		}
	}
}

// TestToLetters covers the convenience wrapper (always uppercase, single-argument form).
// C# equivalent: NumToLettersEn.ConvertNumber(value, isUpper=true).
func TestToLetters(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{701, "ZZ"},
		{702, "AAA"},
		{-1, ""},
	}
	for _, c := range cases {
		got := functions.ToLetters(c.n)
		if got != c.want {
			t.Errorf("ToLetters(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

// TestNumToLettersLower covers the deprecated lowercase helper.
func TestNumToLettersLower(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "a"},
		{1, "b"},
		{25, "z"},
		{26, "aa"},
		{701, "zz"},
		{702, "aaa"},
		{-1, ""},
	}
	for _, c := range cases {
		got := functions.NumToLettersLower(c.n)
		if got != c.want {
			t.Errorf("NumToLettersLower(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestToLettersRu(t *testing.T) {
	cases := []struct {
		n       int
		isUpper bool
		want    string
	}{
		{0, true, "А"},
		{0, false, "а"},
		{1, true, "Б"},
		{1, false, "б"},
		{32, true, "Я"},
		{32, false, "я"},
		{33, true, "АА"},
		{33, false, "аа"},
		{34, true, "АБ"},
		{65, true, "АЯ"},
		{66, true, "БА"},
		{-1, true, ""},
		{-5, false, ""},
	}
	for _, c := range cases {
		got := functions.ToLettersRu(c.n, c.isUpper)
		if got != c.want {
			t.Errorf("ToLettersRu(%d, %v) = %q, want %q", c.n, c.isUpper, got, c.want)
		}
	}
}
