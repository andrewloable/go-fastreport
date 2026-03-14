package reportpkg_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/reportpkg"
	"github.com/andrewloable/go-fastreport/style"
)

// TestHighlight_LoadFRX verifies that <Highlight><Condition .../></Highlight>
// inside a TextObject is deserialized and round-trips without error.
func TestHighlight_LoadFRX(t *testing.T) {
	const frx = `<?xml version="1.0" encoding="utf-8"?>
<Report>
  <ReportPage Name="Page1">
    <ReportTitleBand Name="Title1" Width="1000" Height="100">
      <TextObject Name="Text1" Width="200" Height="50" Text="Hello">
        <Highlight>
          <Condition Expression="1 == 1" Fill.Color="#FF0000" TextFill.Color="#0000FF" ApplyFill="true" ApplyTextFill="true"/>
          <Condition Expression="1 == 2" Fill.Color="#00FF00" ApplyFill="true"/>
        </Highlight>
      </TextObject>
    </ReportTitleBand>
  </ReportPage>
</Report>`

	r := reportpkg.NewReport()
	if err := r.LoadFrom(strings.NewReader(frx)); err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if len(r.Pages()) != 1 {
		t.Fatalf("want 1 page, got %d", len(r.Pages()))
	}

	// Reload to verify round-trip does not crash.
	r2 := reportpkg.NewReport()
	if err := r2.LoadFromString(frx); err != nil {
		t.Fatalf("LoadFromString: %v", err)
	}
	if len(r2.Pages()) != 1 {
		t.Fatalf("reload: want 1 page, got %d", len(r2.Pages()))
	}
}

// TestHighlight_FontFromStr verifies the updated FontFromStr handles FRX format.
func TestHighlight_FontFromStr(t *testing.T) {
	cases := []struct {
		input string
		name  string
		size  float32
		st    style.FontStyle
	}{
		{"Arial, 10, 0", "Arial", 10, style.FontStyleRegular},
		{"Arial, 11pt", "Arial", 11, style.FontStyleRegular},
		{"Tahoma, 14pt, style=Bold", "Tahoma", 14, style.FontStyleBold},
		{"Tahoma, 8pt, style=Bold, Italic", "Tahoma", 8, style.FontStyleBold | style.FontStyleItalic},
	}
	for _, tc := range cases {
		f := style.FontFromStr(tc.input)
		if f.Name != tc.name {
			t.Errorf("FontFromStr(%q).Name = %q, want %q", tc.input, f.Name, tc.name)
		}
		if f.Size != tc.size {
			t.Errorf("FontFromStr(%q).Size = %v, want %v", tc.input, f.Size, tc.size)
		}
		if f.Style != tc.st {
			t.Errorf("FontFromStr(%q).Style = %v, want %v", tc.input, f.Style, tc.st)
		}
	}
}

// TestHighlight_NewHighlightCondition checks constructor defaults.
func TestHighlight_NewHighlightCondition(t *testing.T) {
	c := style.NewHighlightCondition()
	if !c.Visible {
		t.Error("expected Visible=true by default")
	}
	if !c.ApplyTextFill {
		t.Error("expected ApplyTextFill=true by default")
	}
	want := color.RGBA{R: 255, A: 255}
	if c.TextFillColor != want {
		t.Errorf("default TextFillColor = %v, want %v", c.TextFillColor, want)
	}
}
