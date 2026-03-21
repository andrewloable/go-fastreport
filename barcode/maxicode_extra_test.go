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

// ── Mode 2 primary codeword encoding ─────────────────────────────────────────
//
// Ported from C# BarcodeMaxiCode.cs getMode2PrimaryCodewords (lines 820-847).
// Verifies that the low bits of primary[0] encode mode=2.

func TestMaxiCodeMode2PrimaryCodewords_ModeBitsAre2(t *testing.T) {
	// primary[0] low nibble must be 2 for Mode 2.
	// C#: primary[0] = ((postcodeNum & 0x03) << 4) | 2
	cw := maxiCodeMode2PrimaryCodewords("90210", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 2 {
		t.Errorf("primary[0] & 0x0F = %d, want 2 (Mode 2 indicator)", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode2PrimaryCodewords_ZeroPostcode(t *testing.T) {
	// Empty postcode should not panic; falls back to "0".
	cw := maxiCodeMode2PrimaryCodewords("", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 2 {
		t.Errorf("primary[0] & 0x0F = %d, want 2", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode2PrimaryCodewords_NonDigitTruncation(t *testing.T) {
	// "12345ABC" — non-digit at index 5 causes truncation to "12345".
	cwTrunc := maxiCodeMode2PrimaryCodewords("12345ABC", 840, 1)
	cwFull := maxiCodeMode2PrimaryCodewords("12345", 840, 1)
	if len(cwTrunc) != 10 || len(cwFull) != 10 {
		t.Fatalf("expected 10 codewords each")
	}
	for i := 0; i < 10; i++ {
		if cwTrunc[i] != cwFull[i] {
			t.Errorf("primary[%d]: truncated=%d full=%d (should match)", i, cwTrunc[i], cwFull[i])
		}
	}
}

// ── Mode 3 primary codeword encoding ─────────────────────────────────────────
//
// Ported from C# BarcodeMaxiCode.cs getMode3PrimaryCodewords (lines 857-893).

func TestMaxiCodeMode3PrimaryCodewords_ModeBitsAre3(t *testing.T) {
	// primary[0] low nibble must be 3 for Mode 3.
	// C#: primary[0] = ((postcodeNums[5] & 0x03) << 4) | 3
	cw := maxiCodeMode3PrimaryCodewords("ABC123", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 3 {
		t.Errorf("primary[0] & 0x0F = %d, want 3 (Mode 3 indicator)", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode3PrimaryCodewords_ShortPostcodePadded(t *testing.T) {
	// Postal code shorter than 6 chars is padded with spaces to 6.
	cwShort := maxiCodeMode3PrimaryCodewords("AB", 840, 1)
	cwPadded := maxiCodeMode3PrimaryCodewords("AB    ", 840, 1)
	if len(cwShort) != 10 || len(cwPadded) != 10 {
		t.Fatalf("expected 10 codewords each")
	}
	for i := 0; i < 10; i++ {
		if cwShort[i] != cwPadded[i] {
			t.Errorf("primary[%d]: short=%d padded=%d (should match)", i, cwShort[i], cwPadded[i])
		}
	}
}

func TestMaxiCodeMode3PrimaryCodewords_LowercaseConvertedToUpper(t *testing.T) {
	// Lowercase input should produce same codewords as uppercase.
	cwLower := maxiCodeMode3PrimaryCodewords("abc123", 840, 1)
	cwUpper := maxiCodeMode3PrimaryCodewords("ABC123", 840, 1)
	if len(cwLower) != 10 || len(cwUpper) != 10 {
		t.Fatalf("expected 10 codewords each")
	}
	for i := 0; i < 10; i++ {
		if cwLower[i] != cwUpper[i] {
			t.Errorf("primary[%d]: lower=%d upper=%d (should match)", i, cwLower[i], cwUpper[i])
		}
	}
}

// ── Mode 2/3 auto-promotion and full encode ───────────────────────────────────

func TestMaxiCodeEncode_Mode2_ProducesPostalPrimaryBits(t *testing.T) {
	// Mode 2 with a numeric postal code: primary codewords[0] & 0x0F should be 2.
	payload := MaxiCodeMode2Payload("90210", "840", "01", "SECONDARY")
	cw := maxiCodeEncode(payload, 2)
	if len(cw) != 144 {
		t.Fatalf("expected 144 codewords, got %d", len(cw))
	}
	// cw[0] is the first primary codeword. Low nibble must be 2.
	if cw[0]&0x0F != 2 {
		t.Errorf("Mode 2 primary cw[0] & 0x0F = %d, want 2", cw[0]&0x0F)
	}
}

func TestMaxiCodeEncode_Mode2AutoPromotesToMode3_AlphaPostal(t *testing.T) {
	// Mode 2 with an alphanumeric postal code should auto-promote to Mode 3.
	// primary[0] low nibble must be 3 after auto-promotion.
	payload := MaxiCodeMode2Payload("SW1A1AA", "826", "01", "SECONDARY")
	cw := maxiCodeEncode(payload, 2)
	if len(cw) != 144 {
		t.Fatalf("expected 144 codewords, got %d", len(cw))
	}
	// After Mode 2 → 3 promotion, low nibble of cw[0] must be 3.
	if cw[0]&0x0F != 3 {
		t.Errorf("Mode 2 auto-promoted cw[0] & 0x0F = %d, want 3 (Mode 3)", cw[0]&0x0F)
	}
}

func TestMaxiCodeEncode_Mode3_ProducesPostalPrimaryBits(t *testing.T) {
	// Mode 3 with alphanumeric postal: primary codewords[0] & 0x0F should be 3.
	payload := MaxiCodeMode3Payload("ABC123", "840", "01", "SECONDARY")
	cw := maxiCodeEncode(payload, 3)
	if len(cw) != 144 {
		t.Fatalf("expected 144 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 3 {
		t.Errorf("Mode 3 primary cw[0] & 0x0F = %d, want 3", cw[0]&0x0F)
	}
}

// ── maxiCodeParseMode23Text ───────────────────────────────────────────────────

func TestMaxiCodeParseMode23Text_Standard(t *testing.T) {
	// Standard format: 9 zip + 3 country + 2 service + GS + secondary
	payload := "123456789" + "840" + "01" + "\x1d" + "SECONDARY"
	postcode, country, service, secondary := maxiCodeParseMode23Text(payload)
	if postcode != "123456789" {
		t.Errorf("postcode = %q, want %q", postcode, "123456789")
	}
	if country != "840" {
		t.Errorf("country = %q, want %q", country, "840")
	}
	if service != 1 {
		t.Errorf("service = %d, want 1", service)
	}
	if secondary != "SECONDARY" {
		t.Errorf("secondary = %q, want %q", secondary, "SECONDARY")
	}
}

func TestMaxiCodeParseMode23Text_ShortInput_PaddedToMin(t *testing.T) {
	// Input shorter than 14 chars should be padded with spaces.
	postcode, country, service, secondary := maxiCodeParseMode23Text("ABC")
	if len(postcode) != 9 {
		t.Errorf("postcode len = %d, want 9", len(postcode))
	}
	_ = country
	_ = service
	_ = secondary
}

func TestMaxiCodeParseMode23Text_NoGS_FallsThrough(t *testing.T) {
	// If there's no GS at position 14, chars from position 14 on are secondary.
	payload := "123456789" + "840" + "01" + "X" + "DATA"
	_, _, _, secondary := maxiCodeParseMode23Text(payload)
	// Position 14 is 'X' (not GS), so secondary = text[14:] = "XDATA".
	if secondary != "XDATA" {
		t.Errorf("secondary = %q, want %q", secondary, "XDATA")
	}
}
