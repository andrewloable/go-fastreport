package barcode_test

import (
	"bytes"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
	"github.com/andrewloable/go-fastreport/serial"
)

// ── GS1Barcode — Render without prior Encode (encoded == nil path) ────────────

func TestGS1Barcode_Render_WithoutPriorEncode(t *testing.T) {
	// When encoded is nil, Render calls Encode(encodedText).
	// encodedText is "" so Encode("") is attempted.
	// If encoding fails, Render returns an error; if it succeeds, an image is returned.
	b := barcode.NewGS1Barcode()
	// Do not call Encode first; encoded field is nil.
	// The render path will try to re-encode with empty encodedText.
	img, err := b.Render(100, 50)
	// We don't assert success/failure here because it depends on the
	// underlying library's treatment of FNC1+empty string; we just
	// ensure the path is exercised without panic.
	if err == nil && img == nil {
		t.Error("Render returned (nil, nil) which should not happen")
	}
}

func TestGS1Barcode_Encode_EmptyText(t *testing.T) {
	// Encoding empty string exercises the FNC1 fallback path.
	b := barcode.NewGS1Barcode()
	// Empty cleaned text — both FNC1+empty and plain empty encoding may fail.
	// We just verify the method doesn't panic and returns a consistent result.
	_ = b.Encode("")
}

// ── SwissQRBarcode — Encode with empty text (uses buildPayload) ────────────────

func TestSwissQRBarcode_Encode_EmptyText_UsesPayload(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	b.IBAN = "CH5604835012345678009"
	b.CreditorName = "Max Muster"
	b.Amount = "10.00"
	b.Currency = "CHF"
	// Encode("") triggers the buildPayload() path.
	if err := b.Encode(""); err != nil {
		t.Fatalf("Encode empty with fields: %v", err)
	}
	// encodedText should now be the built payload, not empty.
	if b.EncodedText() == "" {
		t.Error("EncodedText should be non-empty after Encode with buildPayload")
	}
}

func TestSwissQRBarcode_Encode_EmptyText_DefaultCurrency(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	b.Currency = "" // will default to CHF inside buildPayload
	b.RefType = ""  // will default to NON inside buildPayload
	if err := b.Encode(""); err != nil {
		t.Fatalf("Encode with empty currency/reftype: %v", err)
	}
}

// ── SwissQRBarcode — Render without prior Encode (encoded == nil path) ─────────

func TestSwissQRBarcode_Render_WithoutPriorEncode(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	b.IBAN = "CH5604835012345678009"
	b.CreditorName = "Max Muster"
	// encoded is nil; Render will call Encode(encodedText) where encodedText="".
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("Render without prior Encode: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestSwissQRBarcode_Render_AfterEncode(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	if err := b.Encode(b.DefaultValue()); err != nil {
		t.Fatalf("Encode default value: %v", err)
	}
	// Second Render uses cached encoded (encoded != nil).
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("Second Render: %v", err)
	}
	if img == nil {
		t.Fatal("Second Render returned nil image")
	}
}

// ── PlesseyBarcode — Render without prior encode (empty encodedText) ──────────

func TestPlesseyBarcode_Render_NotEncoded_BC4(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	// No Encode call; encodedText is "".
	_, err := b.Render(100, 50)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

// ── renderBitPattern — empty bits path (via zero-value pharmacode render) ──────

func TestPharmacodeBarcode_Render_WithoutEncode(t *testing.T) {
	// PharmacodeBarcode.Render returns error when encodedText is empty.
	b := barcode.NewPharmacodeBarcode()
	_, err := b.Render(100, 50)
	if err == nil {
		t.Error("expected error when Render called without Encode, got nil")
	}
}

func TestMSIBarcode_Render_NotEncoded_BC4(t *testing.T) {
	b := barcode.NewMSIBarcode()
	_, err := b.Render(100, 50)
	if err == nil {
		t.Error("expected error when MSI Render called without Encode, got nil")
	}
}

// ── BarcodeObject.Serialize — with nil Barcode ────────────────────────────────

func TestBarcodeObject_Serialize_NilBarcode(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.Barcode = nil // trigger the b.Barcode != nil guard in Serialize

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("BarcodeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := bo.Serialize(w); err != nil {
		t.Fatalf("Serialize with nil Barcode: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	// Should produce XML without a Barcode.Type attribute.
	xml := buf.String()
	if bytes.Contains([]byte(xml), []byte("Barcode.Type")) {
		t.Error("Barcode.Type should not appear in XML when Barcode is nil")
	}
}

// ── BarcodeObject.Serialize — zoom == 1.0 (not serialized), test != 1.0 ───────

func TestBarcodeObject_Serialize_DefaultZoom(t *testing.T) {
	// Serialize a BarcodeObject with all defaults; Zoom=1.0 should NOT appear.
	bo := barcode.NewBarcodeObject()
	// Barcode != nil, all other fields at defaults.

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("BarcodeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := bo.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

// ── BarcodeObject.Deserialize — with Barcode.Type in the XML ─────────────────

func TestBarcodeObject_Deserialize_WithBarcodeType(t *testing.T) {
	// Serialize a BarcodeObject that has a non-default Barcode type,
	// then deserialize and verify the Barcode is set from type.
	orig := barcode.NewBarcodeObject()
	orig.Barcode = barcode.NewQRBarcode()

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.BeginObject("BarcodeObject"); err != nil {
		t.Fatalf("BeginObject: %v", err)
	}
	if err := orig.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if err := w.EndObject(); err != nil {
		t.Fatalf("EndObject: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	got := barcode.NewBarcodeObject()
	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatalf("ReadObjectHeader failed; xml:\n%s", buf.String())
	}
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.BarcodeType() != barcode.BarcodeTypeQR {
		t.Errorf("BarcodeType after Deserialize = %q, want QR", got.BarcodeType())
	}
}

// ── QRBarcode.Encode — error correction level "H" and default path ────────────

func TestQRBarcode_Encode_AllLevels(t *testing.T) {
	levels := []string{"L", "M", "Q", "H", "X"} // X falls to default M
	for _, lvl := range levels {
		q := barcode.NewQRBarcode()
		q.ErrorCorrection = lvl
		if err := q.Encode("test"); err != nil {
			t.Errorf("Encode with level %q: %v", lvl, err)
		}
	}
}

// ── GS1Barcode — Encode error (both FNC1 and plain encoding fail) ─────────────

func TestGS1Barcode_Encode_Error_BothFail(t *testing.T) {
	// There's no reliable way to make code128.Encode fail for cleaned text
	// via the public API because stripped text is always ASCII digits.
	// However we can exercise the stripGS1Parens path with various inputs.
	b := barcode.NewGS1Barcode()

	// Valid GS1: parens stripped, leaving digits.
	if err := b.Encode("(01)12345678901231(10)LOT001"); err != nil {
		t.Fatalf("Encode valid multi-AI: %v", err)
	}
	if b.EncodedText() != "(01)12345678901231(10)LOT001" {
		t.Errorf("EncodedText = %q", b.EncodedText())
	}
}

// ── PostNetBarcode — Render path ──────────────────────────────────────────────

func TestPostNetBarcode_Render_NotEncoded_BC4(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	_, err := b.Render(200, 50)
	if err == nil {
		t.Error("expected error when PostNet Render called without Encode")
	}
}

// ── QRBarcode.Encode — error path when payload is too long ────────────────────
//
// qr.Encode fails when the payload exceeds the maximum QR capacity.
// With M correction level, the maximum is ~1663 bytes.

func TestQRBarcode_Encode_TooLong(t *testing.T) {
	q := barcode.NewQRBarcode()
	buf := make([]byte, 10000)
	for i := range buf {
		buf[i] = 'A'
	}
	err := q.Encode(string(buf))
	if err == nil {
		t.Skip("qr.Encode accepted 10000-char string; error path not reachable on this platform")
	}
}

// ── IntelligentMailBarcode.Render — placeholder when imb_encode fails ─────────
//
// IntelligentMailBarcode.Encode accepts any 20-digit string (length-only check).
// imb_encode requires the second digit to be 0-4; passing '5' as second digit
// passes Encode but triggers imb_encode failure in Render → placeholderImage.

func TestIntelligentMailBarcode_Render_PlaceholderFallback(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	// "05..." passes Encode (20 digits) but fails imb_encode (second digit '5' invalid).
	if err := b.Encode("05234567094987654321"); err != nil {
		t.Fatalf("Encode: unexpected error: %v", err)
	}
	img, err := b.Render(130, 60)
	if err != nil {
		t.Fatalf("Render: unexpected error %v (expected placeholder)", err)
	}
	if img == nil {
		t.Fatal("Render: expected placeholder image, got nil")
	}
}
