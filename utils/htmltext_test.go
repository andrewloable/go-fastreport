package utils

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

func TestHtmlTextRenderer_PlainText(t *testing.T) {
	cases := []struct {
		html  string
		plain string
	}{
		{"hello world", "hello world"},
		{"<b>bold</b> text", "bold text"},
		{"a<br>b", "a\nb"},
		{"&amp;&lt;&gt;&nbsp;&quot;", `&<> "`},
		{"<font color=\"red\">red</font>", "red"},
		{"<span style=\"color:blue\">blue</span>", "blue"},
	}
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	for _, tc := range cases {
		r := NewHtmlTextRenderer(tc.html, f, c)
		got := r.PlainText()
		if got != tc.plain {
			t.Errorf("PlainText(%q) = %q, want %q", tc.html, got, tc.plain)
		}
	}
}

func TestHtmlTextRenderer_Bold(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("<b>hello</b>", f, c)
	lines := r.Lines()
	if len(lines) != 1 || len(lines[0].Runs) == 0 {
		t.Fatal("expected 1 line with runs")
	}
	run := lines[0].Runs[0]
	if run.Font.Style&style.FontStyleBold == 0 {
		t.Error("expected bold font style")
	}
}

func TestStripHtmlTags(t *testing.T) {
	got := StripHtmlTags("<b>hello</b> <i>world</i>")
	if got != "hello world" {
		t.Errorf("StripHtmlTags = %q", got)
	}
}

func TestHtmlTextRenderer_MeasureHeight(t *testing.T) {
	f := style.DefaultFont()
	c := color.RGBA{A: 255}
	r := NewHtmlTextRenderer("line1<br>line2<br>line3", f, c)
	h := r.MeasureHeight(0)
	if h <= 0 {
		t.Error("expected positive height")
	}
}
