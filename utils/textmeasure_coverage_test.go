package utils

import (
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── fontLineSpacingRatio ──────────────────────────────────────────────────────
// Table-driven test covering every switch-case branch.

func TestFontLineSpacingRatio_AllCases(t *testing.T) {
	tests := []struct {
		family string
		want   float32
	}{
		// case "arial", "arial narrow"
		{"Arial", 2355.0 / 2048.0},
		{"arial", 2355.0 / 2048.0},
		{"Arial Narrow", 2355.0 / 2048.0},
		{"arial narrow", 2355.0 / 2048.0},
		// case "times new roman", "times"
		{"Times New Roman", 2355.0 / 2048.0},
		{"times new roman", 2355.0 / 2048.0},
		{"Times", 2355.0 / 2048.0},
		{"times", 2355.0 / 2048.0},
		// case "tahoma", "microsoft sans serif"
		{"Tahoma", 2472.0 / 2048.0},
		{"tahoma", 2472.0 / 2048.0},
		{"Microsoft Sans Serif", 2472.0 / 2048.0},
		{"microsoft sans serif", 2472.0 / 2048.0},
		// case "verdana"
		{"Verdana", 2489.0 / 2048.0},
		{"verdana", 2489.0 / 2048.0},
		// case "arial unicode ms"
		{"Arial Unicode MS", 2743.0 / 2048.0},
		{"arial unicode ms", 2743.0 / 2048.0},
		// case "arial black"
		{"Arial Black", 2899.0 / 2048.0},
		{"arial black", 2899.0 / 2048.0},
		// case "georgia"
		{"Georgia", 2327.0 / 2048.0},
		{"georgia", 2327.0 / 2048.0},
		// case "courier new", "courier"
		{"Courier New", 2320.0 / 2048.0},
		{"courier new", 2320.0 / 2048.0},
		{"Courier", 2320.0 / 2048.0},
		{"courier", 2320.0 / 2048.0},
		// case "segoe ui"
		{"Segoe UI", 2724.0 / 2048.0},
		{"segoe ui", 2724.0 / 2048.0},
		// default — any unrecognised family falls back to Tahoma ratio
		{"UnknownFont", 2472.0 / 2048.0},
		{"", 2472.0 / 2048.0},
		{"Comic Sans MS", 2472.0 / 2048.0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.family, func(t *testing.T) {
			got := fontLineSpacingRatio(tc.family)
			if got != tc.want {
				t.Errorf("fontLineSpacingRatio(%q) = %v, want %v", tc.family, got, tc.want)
			}
		})
	}
}

// ── scaleMaxWidth ─────────────────────────────────────────────────────────────

// Branch 1: maxWidth <= 0 — returns unchanged without touching font size.
func TestScaleMaxWidth_NonPositiveMaxWidth(t *testing.T) {
	f := style.Font{Name: "Arial", Size: 10}
	if got := scaleMaxWidth(0, f); got != 0 {
		t.Errorf("scaleMaxWidth(0, ...) = %v, want 0", got)
	}
	if got := scaleMaxWidth(-5, f); got != -5 {
		t.Errorf("scaleMaxWidth(-5, ...) = %v, want -5", got)
	}
}

// Branch 2: fontPx <= 0 (f.Size <= 0) — returns maxWidth unchanged.
func TestScaleMaxWidth_ZeroFontSize(t *testing.T) {
	f := style.Font{Name: "Arial", Size: 0}
	const maxW float32 = 200
	got := scaleMaxWidth(maxW, f)
	if got != maxW {
		t.Errorf("scaleMaxWidth(%v, zeroSizeFont) = %v, want %v (pass-through)", maxW, got, maxW)
	}
}

func TestScaleMaxWidth_NegativeFontSize(t *testing.T) {
	f := style.Font{Name: "Arial", Size: -1}
	const maxW float32 = 150
	got := scaleMaxWidth(maxW, f)
	if got != maxW {
		t.Errorf("scaleMaxWidth(%v, negativeSizeFont) = %v, want %v (pass-through)", maxW, got, maxW)
	}
}

// Branch 3: normal path — scaled result should differ from the original maxWidth.
func TestScaleMaxWidth_NormalPath_ScalesWidth(t *testing.T) {
	f := style.Font{Name: "Arial", Size: 10}
	const maxW float32 = 100
	got := scaleMaxWidth(maxW, f)
	// fontPx = 10 * 96/72 ≈ 13.333; actualAvgWidth = 13.333 * 0.48 ≈ 6.4
	// scale = 7.0 / 6.4 ≈ 1.09375; result ≈ 109.375 — definitely != 100
	if got == maxW {
		t.Errorf("scaleMaxWidth(%v, 10pt) = %v, want a scaled value != %v", maxW, got, maxW)
	}
	if got <= 0 {
		t.Errorf("scaleMaxWidth(%v, 10pt) = %v, want > 0", maxW, got)
	}
}

func TestScaleMaxWidth_LargerFont_SmallerScaleFactor(t *testing.T) {
	// A larger font produces a larger actualAvgWidth, so the scale factor is
	// smaller, and the returned width is closer to (or less than) maxWidth.
	fSmall := style.Font{Name: "Arial", Size: 5}
	fLarge := style.Font{Name: "Arial", Size: 50}
	const maxW float32 = 200
	scaledSmall := scaleMaxWidth(maxW, fSmall)
	scaledLarge := scaleMaxWidth(maxW, fLarge)
	if scaledSmall <= scaledLarge {
		t.Errorf("small font (%v) should produce larger scaled width than large font (%v)", scaledSmall, scaledLarge)
	}
}
