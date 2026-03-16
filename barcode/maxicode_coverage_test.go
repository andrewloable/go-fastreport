package barcode_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── MaxiCodeMode3Payload ──────────────────────────────────────────────────────

func TestMaxiCodeMode3Payload_Basic(t *testing.T) {
	// Mode 3 uses the same logic as Mode 2 — just calls MaxiCodeMode2Payload.
	payload := barcode.MaxiCodeMode3Payload("ABC123456", "840", "01", "Secondary data")
	// Should have: 9-char zip + 3-char country + 2-char service + GS + secondary.
	if !strings.HasPrefix(payload, "ABC123456840") {
		t.Errorf("unexpected payload prefix: %q", payload[:12])
	}
	if !strings.Contains(payload, "\x1d") {
		t.Error("payload should contain GS character (0x1d)")
	}
	if !strings.Contains(payload, "Secondary data") {
		t.Error("payload should contain secondary data")
	}
}

func TestMaxiCodeMode3Payload_ShortZip(t *testing.T) {
	// ZIP shorter than 9 chars → padded with spaces.
	payload := barcode.MaxiCodeMode3Payload("123", "840", "01", "SEC")
	// Should be 9 chars for zip (padded).
	if len(payload) < 9+3+2+1 {
		t.Errorf("payload too short: len=%d", len(payload))
	}
	// First 9 chars: "123      " (padded to 9).
	if payload[:3] != "123" {
		t.Errorf("zip prefix = %q, want 123", payload[:3])
	}
}

func TestMaxiCodeMode3Payload_LongZip(t *testing.T) {
	// ZIP longer than 9 chars → truncated.
	payload := barcode.MaxiCodeMode3Payload("12345678901234", "840", "01", "SEC")
	// First 9 chars should be "123456789".
	if payload[:9] != "123456789" {
		t.Errorf("zip = %q, want 123456789", payload[:9])
	}
}

func TestMaxiCodeMode3Payload_ShortCountry(t *testing.T) {
	// Country shorter than 3 → padded.
	payload := barcode.MaxiCodeMode3Payload("123456789", "84", "01", "")
	// Country field is chars 9..11.
	country := payload[9:12]
	if !strings.HasPrefix(country, "84") {
		t.Errorf("country = %q, want starts with 84", country)
	}
}

func TestMaxiCodeMode3Payload_ExactLengths(t *testing.T) {
	// All fields exactly match their lengths.
	payload := barcode.MaxiCodeMode3Payload("123456789", "840", "01", "HELLO")
	if payload[:9] != "123456789" {
		t.Errorf("zip = %q, want 123456789", payload[:9])
	}
	if payload[9:12] != "840" {
		t.Errorf("country = %q, want 840", payload[9:12])
	}
	if payload[12:14] != "01" {
		t.Errorf("service = %q, want 01", payload[12:14])
	}
	if payload[14] != '\x1d' {
		t.Errorf("GS char missing at pos 14")
	}
	if payload[15:] != "HELLO" {
		t.Errorf("secondary = %q, want HELLO", payload[15:])
	}
}

// ── MaxiCodeMode2Payload (padRight coverage) ──────────────────────────────────

func TestMaxiCodeMode2Payload_PadRight_ExactLength(t *testing.T) {
	// All fields exactly at their max lengths → no padding needed (len(s) >= n → s[:n]).
	payload := barcode.MaxiCodeMode2Payload("123456789", "840", "01", "DATA")
	if payload[:9] != "123456789" {
		t.Errorf("zip = %q", payload[:9])
	}
}

func TestMaxiCodeMode2Payload_PadRight_ShortService(t *testing.T) {
	// Service class with length < 2 → padded.
	payload := barcode.MaxiCodeMode2Payload("123456789", "840", "1", "DATA")
	// Service is chars 12..13, should be "1 ".
	svc := payload[12:14]
	if svc[0] != '1' {
		t.Errorf("service[0] = %q, want '1'", string(svc[0]))
	}
	if svc[1] != ' ' {
		t.Errorf("service[1] = %q, want ' '", string(svc[1]))
	}
}

func TestMaxiCodeMode2Payload_EmptySecondary(t *testing.T) {
	payload := barcode.MaxiCodeMode2Payload("123456789", "840", "01", "")
	// Should still have GS character.
	if payload[14] != '\x1d' {
		t.Errorf("GS missing, payload[14]=%q", string(payload[14]))
	}
	if len(payload) != 15 { // 9 + 3 + 2 + 1 + 0
		t.Errorf("payload len = %d, want 15", len(payload))
	}
}

// ── min2 function ─────────────────────────────────────────────────────────────

// min2 is called from maxiCodeRender which is called from MaxiCodeBarcode.Render.
// min2(a, b) returns a when a < b, returns b when a >= b.

func TestMaxiCodeBarcode_Render_SquareImage_Min2Equal(t *testing.T) {
	// width == height → min2(w, h) = h (second branch when a >= b).
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("HELLO WORLD"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("size = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestMaxiCodeBarcode_Render_TallerImage_Min2FirstSmaller(t *testing.T) {
	// width < height → min2(width, height) returns width (first branch a < b).
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("TEST DATA"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(80, 120) // width < height
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestMaxiCodeBarcode_Render_WiderImage_Min2SecondSmaller(t *testing.T) {
	// width > height → min2(width, height) returns height (second branch a >= b).
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("TEST DATA 2"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(150, 100) // width > height
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

// ── maxiCodeEncodeText coverage ───────────────────────────────────────────────

// maxiCodeEncodeText is exercised via Encode + Render.
// Cover: non-ASCII char, control char (set A), LATB/LATA switching.

func TestMaxiCodeBarcode_Encode_TextWithControlChar(t *testing.T) {
	// Text containing control chars (< 0x20) triggers LATA/LATB switching.
	// Include '\t' (0x09) which is < 0x20.
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("HELLO\tWORLD"); err != nil {
		t.Fatalf("Encode with control char: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil for control char")
	}
}

func TestMaxiCodeBarcode_Encode_TextWithNonASCII(t *testing.T) {
	// Non-ASCII characters are substituted with GS (0x1D).
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("HELLO\xC0WORLD"); err != nil { // 0xC0 > 0x7F
		t.Fatalf("Encode with non-ASCII: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil for non-ASCII")
	}
}

func TestMaxiCodeBarcode_Encode_LongText_TruncatesToMaxCW(t *testing.T) {
	// Very long text → truncated at maxCW codewords.
	b := barcode.NewMaxiCodeBarcode()
	longText := strings.Repeat("ABCDEFGHIJ", 20) // 200 chars
	if err := b.Encode(longText); err != nil {
		t.Fatalf("Encode long text: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render long text: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil for long text")
	}
}

func TestMaxiCodeBarcode_Encode_TextWithMultipleControlSwitches(t *testing.T) {
	// Text that switches between Set A and Set B multiple times.
	// \x01 (ctrl) → LATA → encode ctrl char → then 'A' → LATB → then \x02 → LATA again.
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("A\x01B\x02C"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render nil")
	}
}

// ── maxiCodeEncode mode 5 ─────────────────────────────────────────────────────

func TestMaxiCodeBarcode_Mode5_Render(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 5
	if err := b.Encode("MODE5 TEST DATA"); err != nil {
		t.Fatalf("Encode mode 5: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render mode 5: %v", err)
	}
	if img == nil {
		t.Fatal("Render mode 5 returned nil")
	}
}

// ── Encode / Render error paths ───────────────────────────────────────────────

func TestMaxiCodeBarcode_Encode_InvalidMode_Coverage(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 1 // invalid (must be 2–6)
	if err := b.Encode("TEST"); err == nil {
		t.Error("expected error for invalid mode 1")
	}
}

func TestMaxiCodeBarcode_Encode_InvalidMode7_Coverage(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	b.Mode = 7 // invalid
	if err := b.Encode("TEST"); err == nil {
		t.Error("expected error for invalid mode 7")
	}
}

func TestMaxiCodeBarcode_Render_NotEncoded_Coverage(t *testing.T) {
	b := barcode.NewMaxiCodeBarcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("Render before Encode should return error")
	}
}

func TestMaxiCodeBarcode_Render_DefaultSize(t *testing.T) {
	// width=0 and height=0 → both default to 100.
	b := barcode.NewMaxiCodeBarcode()
	if err := b.Encode("DEFAULT SIZE"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render(0,0): %v", err)
	}
	if img == nil {
		t.Fatal("Render(0,0) returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("default size = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

// ── MaxiCodeComputeECC (exported for testing) ─────────────────────────────────

func TestMaxiCodeComputeECC_Basic(t *testing.T) {
	data := []byte{4, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	ecc := barcode.MaxiCodeComputeECC(data, 10)
	if len(ecc) != 10 {
		t.Errorf("ECC length = %d, want 10", len(ecc))
	}
}

func TestMaxiCodeComputeECC_AllZeros(t *testing.T) {
	data := make([]byte, 10)
	ecc := barcode.MaxiCodeComputeECC(data, 10)
	if len(ecc) != 10 {
		t.Errorf("ECC length = %d, want 10", len(ecc))
	}
}

// ── All valid modes 2–6 ───────────────────────────────────────────────────────

func TestMaxiCodeBarcode_AllModes(t *testing.T) {
	for mode := 2; mode <= 6; mode++ {
		b := barcode.NewMaxiCodeBarcode()
		b.Mode = mode
		if err := b.Encode("MODE TEST"); err != nil {
			t.Errorf("mode %d Encode: %v", mode, err)
			continue
		}
		img, err := b.Render(80, 80)
		if err != nil {
			t.Errorf("mode %d Render: %v", mode, err)
			continue
		}
		if img == nil {
			t.Errorf("mode %d returned nil", mode)
		}
	}
}
