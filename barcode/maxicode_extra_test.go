package barcode

// Internal (white-box) tests for maxicode.go uncovered branches.
// Using package barcode to access unexported functions directly.

import (
	"image"
	"image/color"
	"testing"
)

// ── drawHex bounds-clamp paths ────────────────────────────────────────────────

func TestDrawHex_BoundsClamp_TopLeft(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	drawHex(img, 0, 0, 5, 5, c)
	got := img.NRGBAAt(0, 0)
	if got.A != 255 {
		t.Errorf("pixel (0,0) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_BottomRight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	c := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	drawHex(img, 10, 10, 5, 5, c)
	got := img.NRGBAAt(9, 9)
	if got.A != 255 {
		t.Errorf("pixel (9,9) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_NegativeXOnly(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	c := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	drawHex(img, 0, 10, 3, 3, c)
	got := img.NRGBAAt(0, 10)
	if got.A != 255 {
		t.Errorf("pixel (0,10) expected painted, got alpha=%d", got.A)
	}
}

func TestDrawHex_BoundsClamp_NegativeYOnly(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	c := color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	drawHex(img, 10, 0, 3, 3, c)
	got := img.NRGBAAt(10, 0)
	if got.A != 255 {
		t.Errorf("pixel (10,0) expected painted, got alpha=%d", got.A)
	}
}

// ── maxiCodeEncode internal default-mode fallback ────────────────────────────

func TestMaxiCodeEncode_OutOfRangeMode_DefaultsTo4(t *testing.T) {
	// mode=0 is invalid → defaults to 4 internally. Returns a 33×30 grid.
	grid, err := maxiCodeEncode("HELLO", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the grid is non-zero (has some dark cells from mode-4 encoding).
	dark := 0
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			if grid[row][col] {
				dark++
			}
		}
	}
	if dark == 0 {
		t.Error("expected non-empty grid for mode=0 (defaults to 4)")
	}
}

func TestMaxiCodeEncode_Mode7_DefaultsTo4(t *testing.T) {
	grid, err := maxiCodeEncode("TEST", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dark := 0
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			if grid[row][col] {
				dark++
			}
		}
	}
	if dark == 0 {
		t.Error("expected non-empty grid for mode=7 (defaults to 4)")
	}
}

// ── Mode 2 primary codeword encoding ─────────────────────────────────────────

func TestMaxiCodeMode2PrimaryCodewords_ModeBitsAre2(t *testing.T) {
	cw := maxiCodeMode2PrimaryCodewordsInt("90210", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 2 {
		t.Errorf("primary[0] & 0x0F = %d, want 2 (Mode 2 indicator)", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode2PrimaryCodewords_ZeroPostcode(t *testing.T) {
	cw := maxiCodeMode2PrimaryCodewordsInt("", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 2 {
		t.Errorf("primary[0] & 0x0F = %d, want 2", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode2PrimaryCodewords_NonDigitTruncation(t *testing.T) {
	cwTrunc := maxiCodeMode2PrimaryCodewordsInt("12345ABC", 840, 1)
	cwFull := maxiCodeMode2PrimaryCodewordsInt("12345", 840, 1)
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

func TestMaxiCodeMode3PrimaryCodewords_ModeBitsAre3(t *testing.T) {
	cw := maxiCodeMode3PrimaryCodewordsInt("ABC123", 840, 1)
	if len(cw) != 10 {
		t.Errorf("expected 10 codewords, got %d", len(cw))
	}
	if cw[0]&0x0F != 3 {
		t.Errorf("primary[0] & 0x0F = %d, want 3 (Mode 3 indicator)", cw[0]&0x0F)
	}
}

func TestMaxiCodeMode3PrimaryCodewords_ShortPostcodePadded(t *testing.T) {
	cwShort := maxiCodeMode3PrimaryCodewordsInt("AB", 840, 1)
	cwPadded := maxiCodeMode3PrimaryCodewordsInt("AB    ", 840, 1)
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
	cwLower := maxiCodeMode3PrimaryCodewordsInt("abc123", 840, 1)
	cwUpper := maxiCodeMode3PrimaryCodewordsInt("ABC123", 840, 1)
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
	payload := MaxiCodeMode2Payload("90210", "840", "01", "SECONDARY")
	grid, err := maxiCodeEncode(payload, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify grid is non-empty.
	dark := 0
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			if grid[row][col] {
				dark++
			}
		}
	}
	if dark == 0 {
		t.Error("expected non-empty grid for Mode 2 encoding")
	}
}

func TestMaxiCodeEncode_Mode2AutoPromotesToMode3_AlphaPostal(t *testing.T) {
	payload := MaxiCodeMode2Payload("SW1A1AA", "826", "01", "SECONDARY")
	_, err := maxiCodeEncode(payload, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaxiCodeEncode_Mode3_ProducesGrid(t *testing.T) {
	payload := MaxiCodeMode3Payload("ABC123", "840", "01", "SECONDARY")
	grid, err := maxiCodeEncode(payload, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dark := 0
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			if grid[row][col] {
				dark++
			}
		}
	}
	if dark == 0 {
		t.Error("expected non-empty grid for Mode 3 encoding")
	}
}

// ── maxiCodeParseMode23Text ───────────────────────────────────────────────────

func TestMaxiCodeParseMode23Text_Standard(t *testing.T) {
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
	postcode, _, _, _ := maxiCodeParseMode23Text("ABC")
	if len(postcode) != 9 {
		t.Errorf("postcode len = %d, want 9", len(postcode))
	}
}

func TestMaxiCodeParseMode23Text_NoGS_FallsThrough(t *testing.T) {
	payload := "123456789" + "840" + "01" + "X" + "DATA"
	_, _, _, secondary := maxiCodeParseMode23Text(payload)
	if secondary != "XDATA" {
		t.Errorf("secondary = %q, want %q", secondary, "XDATA")
	}
}
