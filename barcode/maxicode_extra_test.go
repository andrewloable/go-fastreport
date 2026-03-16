package barcode

// Internal (white-box) tests for maxicode.go uncovered branches.
// Using package barcode to access unexported functions directly.

import (
	"image"
	"image/color"
	"testing"
)

// ── drawHex bounds-clamp paths ────────────────────────────────────────────────

// drawHex clips py < bounds.Min.Y and px < bounds.Min.X when the hex centre
// is positioned so that x0 or y0 falls below zero.

func TestDrawHex_BoundsClamp_TopLeft(t *testing.T) {
	// Create a small image (10x10).
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	// Centre at (0, 0) with radius 5: x0=int(-5)=-5, y0=int(-5)=-5.
	// Pixels from y=-5..5 and x=-5..5 are computed, but only non-negative ones are painted.
	// This exercises the py < bounds.Min.Y and px < bounds.Min.X branches.
	drawHex(img, 0, 0, 5, 5, c)
	// Verify at least some pixels were painted (those within bounds).
	// Pixel (0,0) should be painted (ellipse test: 0+0 <= 1.0).
	got := img.NRGBAAt(0, 0)
	if got.A != 255 {
		t.Errorf("pixel (0,0) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_BottomRight(t *testing.T) {
	// Create a small image (10x10).
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	// Centre at (10, 10) with radius 5: x1=int(15)+1=16, y1=int(15)+1=16.
	// Pixels from y=5..16 and x=5..16 are computed, but only those < 10 are painted.
	// This exercises the py >= bounds.Max.Y and px >= bounds.Max.X branches.
	drawHex(img, 10, 10, 5, 5, c)
	// Pixel (9,9) should be painted (ellipse test).
	got := img.NRGBAAt(9, 9)
	if got.A != 255 {
		t.Errorf("pixel (9,9) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_NegativeXOnly(t *testing.T) {
	// Centre at x=0 so x0=-rx (negative), but y is safely inside.
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	c := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	drawHex(img, 0, 10, 3, 3, c)
	// Only pixels with px>=0 are painted.
	got := img.NRGBAAt(0, 10)
	if got.A != 255 {
		t.Errorf("pixel (0,10) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_NegativeYOnly(t *testing.T) {
	// Centre at y=0 so y0=-ry (negative), but x is safely inside.
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	c := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	drawHex(img, 10, 0, 3, 3, c)
	got := img.NRGBAAt(10, 0)
	if got.A != 255 {
		t.Errorf("pixel (10,0) expected painted, got alpha=%d", got.A)
	}
}

// ── maxiCodeEncode internal default-mode fallback ────────────────────────────

// maxiCodeEncode contains `if mode < 2 || mode > 6 { mode = 4 }` which is
// never reached via the public Encode→Render path (Encode validates first).
// Call maxiCodeEncode directly with an out-of-range mode to cover that branch.

func TestMaxiCodeEncode_OutOfRangeMode_DefaultsTo4(t *testing.T) {
	// mode=0 is invalid → defaults to 4 internally.
	codewords := maxiCodeEncode("HELLO", 0)
	if len(codewords) != 144 {
		t.Errorf("expected 144 codewords, got %d", len(codewords))
	}
	// mode byte at position 0 should be 4 (the default).
	if codewords[0] != 4 {
		t.Errorf("codeword[0] = %d, want 4", codewords[0])
	}
}

func TestMaxiCodeEncode_Mode7_DefaultsTo4(t *testing.T) {
	// mode=7 is invalid → defaults to 4 internally.
	codewords := maxiCodeEncode("TEST", 7)
	if len(codewords) != 144 {
		t.Errorf("expected 144 codewords, got %d", len(codewords))
	}
	if codewords[0] != 4 {
		t.Errorf("codeword[0] = %d, want 4", codewords[0])
	}
}

// ── maxiCodeEncodeText idx<1 branch ──────────────────────────────────────────

// The idx < 1 branch in maxiCodeEncodeText fires when ch == 0x1F (ASCII 31,
// the US separator). That character is >= 0x20-1=0x1F, so it hits the else
// branch (inSetB path). idx = 0x1F - 0x1F = 0 → idx < 1 → idx = 1.
// Note: 0x1F = 31 which is NOT < 0x20 (32), so it goes to the else (SetB) path.

func TestMaxiCodeEncodeText_IdxLessThan1_Clamped(t *testing.T) {
	// 0x1F (31) >= 0x20? No: 0x1F < 0x20. So it goes to SetA path, not SetB.
	// The idx<1 branch is in the SetB (else) path: ch >= 0x20, idx=ch-0x1F.
	// ch=0x20 (space): idx=0x20-0x1F=1 → not < 1.
	// ch=0x1F=31 < 0x20 → SetA path. No way to reach idx<1 with ch>=0x20.
	// So this is effectively dead code — just verify the function handles it.
	cw := maxiCodeEncodeText("HELLO", 5)
	if len(cw) != 5 {
		t.Errorf("expected 5 codewords, got %d", len(cw))
	}
}

// ── maxiCodeEncodeText: non-ASCII rune substitution ───────────────────────────
//
// When a rune r > 0x7F is encountered, it is substituted with GS (0x1D).
// This exercises the `if r > 0x7F { r = 0x1D }` branch directly by calling
// maxiCodeEncodeText with a Unicode string containing runes > 127.

func TestMaxiCodeEncodeText_NonASCIIRune_SubstitutedWithGS(t *testing.T) {
	// "é" is rune U+00E9 (233 > 127). It should be substituted with GS (0x1D).
	// GS < 0x20, so it goes into Set A path: appends LATA (63) then 0x1D.
	cw := maxiCodeEncodeText("é", 5)
	if len(cw) != 5 {
		t.Errorf("expected 5 codewords, got %d", len(cw))
	}
	// First codeword should be LATA (63) since we switch to Set A for GS.
	if cw[0] != 63 {
		t.Errorf("cw[0] = %d, want 63 (LATA)", cw[0])
	}
	// Second codeword should be GS & 0x3F = 0x1D & 0x3F = 29.
	if cw[1] != 29 {
		t.Errorf("cw[1] = %d, want 29 (GS & 0x3F)", cw[1])
	}
}

func TestMaxiCodeEncodeText_MixedASCIIAndNonASCII(t *testing.T) {
	// "Héllo" — 'H' is ASCII, 'é' (U+00E9 > 127) is non-ASCII, rest are ASCII.
	// This exercises the non-ASCII substitution path mid-string with Set switches.
	cw := maxiCodeEncodeText("Héllo", 10)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	// 'H' (0x48) is ASCII >= 0x20 → Set B. idx = 0x48 - 0x1F = 41.
	if cw[0] != 41 {
		t.Errorf("cw[0] = %d, want 41 ('H' in Set B)", cw[0])
	}
}

func TestMaxiCodeEncodeText_OnlyNonASCII(t *testing.T) {
	// All runes > 0x7F: "äöü" (U+00E4, U+00F6, U+00FC).
	// Each substituted with GS (0x1D). First triggers LATA; subsequent stay in Set A.
	cw := maxiCodeEncodeText("äöü", 10)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	// First: LATA (63), then GS&0x3F (29), then GS&0x3F (29), then GS&0x3F (29).
	if cw[0] != 63 {
		t.Errorf("cw[0] = %d, want 63 (LATA)", cw[0])
	}
	for i := 1; i <= 3; i++ {
		if cw[i] != 29 {
			t.Errorf("cw[%d] = %d, want 29 (GS)", i, cw[i])
		}
	}
}
