package style

import (
	"testing"
)

func TestDefaultFont(t *testing.T) {
	f := DefaultFont()
	if f.Name != "Arial" {
		t.Errorf("Name = %q, want Arial", f.Name)
	}
	if f.Size != 10 {
		t.Errorf("Size = %v, want 10", f.Size)
	}
	if f.Style != FontStyleRegular {
		t.Errorf("Style = %v, want Regular", f.Style)
	}
}

func TestFontEqual(t *testing.T) {
	a := Font{Name: "Arial", Size: 10, Style: FontStyleBold}
	b := Font{Name: "Arial", Size: 10, Style: FontStyleBold}
	c := Font{Name: "Tahoma", Size: 10, Style: FontStyleBold}

	if !FontEqual(a, b) {
		t.Error("FontEqual should be true for identical fonts")
	}
	if FontEqual(a, c) {
		t.Error("FontEqual should be false for different fonts")
	}
}

func TestFontToStr(t *testing.T) {
	f := Font{Name: "Arial", Size: 10, Style: FontStyleRegular}
	got := FontToStr(f)
	// Should contain the name, size, and style.
	if got == "" {
		t.Error("FontToStr returned empty string")
	}
	// Round-trip: FontFromStr should recover the font.
	f2 := FontFromStr(got)
	if !FontEqual(f, f2) {
		t.Errorf("round-trip failed: FontToStr=%q → FontFromStr=%v, want %v", got, f2, f)
	}
}

func TestFontFromStr_TooFewParts(t *testing.T) {
	// Less than 2 parts → DefaultFont()
	got := FontFromStr("Arial")
	want := DefaultFont()
	if !FontEqual(got, want) {
		t.Errorf("FontFromStr(\"Arial\") = %v, want DefaultFont %v", got, want)
	}
}

func TestFontFromStr_FRXFormat(t *testing.T) {
	// FRX format: "Name, Sizept"
	got := FontFromStr("Arial, 11pt")
	if got.Name != "Arial" {
		t.Errorf("Name = %q, want Arial", got.Name)
	}
	if got.Size != 11 {
		t.Errorf("Size = %v, want 11", got.Size)
	}
	if got.Style != FontStyleRegular {
		t.Errorf("Style = %v, want Regular", got.Style)
	}
}

func TestFontFromStr_FRXFormatWithStyle(t *testing.T) {
	// FRX format: "Name, Sizept, style=Bold"
	got := FontFromStr("Tahoma, 14pt, style=Bold")
	if got.Name != "Tahoma" {
		t.Errorf("Name = %q, want Tahoma", got.Name)
	}
	if got.Size != 14 {
		t.Errorf("Size = %v, want 14", got.Size)
	}
	if got.Style != FontStyleBold {
		t.Errorf("Style = %v, want Bold", got.Style)
	}
}

func TestFontFromStr_FRXMultipleStyles(t *testing.T) {
	// "style=Bold, Italic" — comma-separated names
	got := FontFromStr("Arial, 10pt, style=Bold, Italic")
	if got.Style&FontStyleBold == 0 {
		t.Error("expected Bold flag set")
	}
	if got.Style&FontStyleItalic == 0 {
		t.Error("expected Italic flag set")
	}
}

func TestFontFromStr_FRXStyleUnderlineStrikeout(t *testing.T) {
	got := FontFromStr("Arial, 10pt, style=Underline, Strikeout")
	if got.Style&FontStyleUnderline == 0 {
		t.Error("expected Underline flag set")
	}
	if got.Style&FontStyleStrikeout == 0 {
		t.Error("expected Strikeout flag set")
	}
}

func TestFontFromStr_NumericStyle(t *testing.T) {
	// Go round-trip format: "Arial, 10, 3" (Bold|Italic)
	got := FontFromStr("Arial, 10, 3")
	if got.Name != "Arial" {
		t.Errorf("Name = %q, want Arial", got.Name)
	}
	if got.Style != FontStyle(3) {
		t.Errorf("Style = %v, want 3", got.Style)
	}
}

func TestFontFromStr_BadSize(t *testing.T) {
	// Unparseable size → DefaultFont
	got := FontFromStr("Arial, notasize, 0")
	want := DefaultFont()
	if !FontEqual(got, want) {
		t.Errorf("FontFromStr bad size = %v, want DefaultFont %v", got, want)
	}
}
