package style

import (
	"image/color"
	"testing"
)

func TestNewHighlightCondition(t *testing.T) {
	h := NewHighlightCondition()
	if !h.Visible {
		t.Error("expected Visible=true")
	}
	if !h.ApplyTextFill {
		t.Error("expected ApplyTextFill=true")
	}
	want := color.RGBA{R: 255, A: 255}
	if h.TextFillColor != want {
		t.Errorf("TextFillColor = %v, want %v", h.TextFillColor, want)
	}
	if h.ApplyBorder || h.ApplyFill || h.ApplyFont {
		t.Error("unexpected Apply flags set")
	}
}

func TestDefaultTextOutline(t *testing.T) {
	o := DefaultTextOutline()
	if o.Enabled {
		t.Error("expected Enabled=false (disabled by default)")
	}
	if o.Width != 1 {
		t.Errorf("Width = %v, want 1", o.Width)
	}
	want := color.RGBA{A: 255}
	if o.Color != want {
		t.Errorf("Color = %v, want %v", o.Color, want)
	}
}
