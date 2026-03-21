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
		{25, true, "Z"},
		{26, true, "AA"},
		{26, false, "aa"},
		{701, true, "ZZ"},
		{702, true, "AAA"},
	}
	for _, c := range cases {
		got := functions.ToLettersEn(c.n, c.isUpper)
		if got != c.want {
			t.Errorf("ToLettersEn(%d, %v) = %q, want %q", c.n, c.isUpper, got, c.want)
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
		{32, true, "Я"},
		{33, true, "АА"},
		{33, false, "аа"},
		{65, true, "АЯ"},
		{66, true, "БА"},
	}
	for _, c := range cases {
		got := functions.ToLettersRu(c.n, c.isUpper)
		if got != c.want {
			t.Errorf("ToLettersRu(%d, %v) = %q, want %q", c.n, c.isUpper, got, c.want)
		}
	}
}
