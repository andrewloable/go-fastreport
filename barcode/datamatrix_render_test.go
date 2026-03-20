package barcode

// datamatrix_render_test.go — tests for DataMatrix/GS1Datamatrix Render,
// Code128 Render, BaseBarcodeImpl Render success path, and GetPattern
// error paths for linear barcodes still below 100% Render coverage.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// DataMatrixBarcode.Render — encode then render with various data payloads
// ---------------------------------------------------------------------------

func TestDataMatrixBarcode_Render_NumericPayload(t *testing.T) {
	b := NewDataMatrixBarcode()
	if err := b.Encode("1234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 200 {
		t.Errorf("bounds = %dx%d, want 200x200", bounds.Dx(), bounds.Dy())
	}
}

func TestDataMatrixBarcode_Render_AlphanumericPayload(t *testing.T) {
	b := NewDataMatrixBarcode()
	if err := b.Encode("Hello DataMatrix World 2024!"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 300)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestDataMatrixBarcode_Render_ShortPayload(t *testing.T) {
	b := NewDataMatrixBarcode()
	if err := b.Encode("A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestDataMatrixBarcode_Render_URLPayload(t *testing.T) {
	b := NewDataMatrixBarcode()
	if err := b.Encode("https://example.com/path?q=abc"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(250, 250)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// GS1DatamatrixBarcode.Render — encode then render
// ---------------------------------------------------------------------------

func TestGS1DatamatrixBarcode_Render_SingleAI(t *testing.T) {
	b := NewGS1DatamatrixBarcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestGS1DatamatrixBarcode_Render_MultipleAIs(t *testing.T) {
	b := NewGS1DatamatrixBarcode()
	if err := b.Encode("(01)09521234543213(21)12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(250, 250)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestGS1DatamatrixBarcode_Render_PlainText(t *testing.T) {
	b := NewGS1DatamatrixBarcode()
	if err := b.Encode("GS1 DataMatrix plain text"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(150, 150)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestGS1DatamatrixBarcode_Render_NotEncoded_Internal(t *testing.T) {
	b := NewGS1DatamatrixBarcode()
	_, err := b.Render(200, 200)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ---------------------------------------------------------------------------
// BaseBarcodeImpl.Render — success path (internal access to set encodedText)
// ---------------------------------------------------------------------------

func TestBaseBarcodeImpl_Render_SuccessPath(t *testing.T) {
	b := &BaseBarcodeImpl{encodedText: "test-data"}
	img, err := b.Render(100, 50)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("bounds = %dx%d, want 100x50", bounds.Dx(), bounds.Dy())
	}
}

func TestBaseBarcodeImpl_Render_SuccessPath_ZeroDimensions(t *testing.T) {
	b := &BaseBarcodeImpl{encodedText: "test-data"}
	img, err := b.Render(0, 0)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image for zero dimensions")
	}
}

// ---------------------------------------------------------------------------
// Code128Barcode.Render — success and error paths
// ---------------------------------------------------------------------------

func TestCode128Barcode_Render_SuccessPath(t *testing.T) {
	b := NewCode128Barcode()
	if err := b.Encode("Hello123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 300 || bounds.Dy() != 100 {
		t.Errorf("bounds = %dx%d, want 300x100", bounds.Dx(), bounds.Dy())
	}
}

func TestCode128Barcode_Render_NotEncoded(t *testing.T) {
	b := NewCode128Barcode()
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ---------------------------------------------------------------------------
// Code39Barcode.Render — GetPattern error path (not reachable since Code39
// GetPattern never returns error, but test the success path thoroughly)
// ---------------------------------------------------------------------------

func TestCode39Barcode_Render_UppercaseText(t *testing.T) {
	b := NewCode39Barcode()
	if err := b.Encode("ABCDEF"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// QRBarcode.Render — success path
// ---------------------------------------------------------------------------

func TestQRBarcode_Render_ShortText(t *testing.T) {
	b := NewQRBarcode()
	if err := b.Encode("QR"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(150, 150)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestQRBarcode_Render_NotEncoded_Internal(t *testing.T) {
	b := NewQRBarcode()
	_, err := b.Render(100, 100)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ---------------------------------------------------------------------------
// EAN13Barcode.Render — success path
// ---------------------------------------------------------------------------

func TestEAN13Barcode_Render_13Digits(t *testing.T) {
	b := NewEAN13Barcode()
	if err := b.Encode("5901234123457"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// AztecBarcode.Render — success path
// ---------------------------------------------------------------------------

func TestAztecBarcode_Render_LongerText(t *testing.T) {
	b := NewAztecBarcode()
	if err := b.Encode("This is a longer Aztec barcode test string"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(250, 250)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// PDF417Barcode.Render — success path
// ---------------------------------------------------------------------------

func TestPDF417Barcode_Render_TextPayload(t *testing.T) {
	b := NewPDF417Barcode()
	if err := b.Encode("PDF417 payload data"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 150)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code93Barcode.Render — GetPattern error path (invalid char triggers error)
// ---------------------------------------------------------------------------

func TestCode93Barcode_Render_GetPatternError(t *testing.T) {
	b := NewCode93Barcode()
	// Bypass Encode by setting encodedText directly with an invalid char
	// that code93GetPattern cannot handle.
	b.encodedText = "\x01" // non-printable, not in Code93 table
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for invalid char in Code93")
	}
}

// ---------------------------------------------------------------------------
// Code2of5Barcode.Render — GetPattern error path (empty after trim)
// ---------------------------------------------------------------------------

func TestCode2of5Barcode_Render_GetPatternError(t *testing.T) {
	b := NewCode2of5Barcode()
	// Set encodedText to whitespace-only so GetPattern returns error
	b.encodedText = "   "
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for whitespace-only input")
	}
}

// ---------------------------------------------------------------------------
// CodabarBarcode.Render — verify success path exercised
// ---------------------------------------------------------------------------

func TestCodabarBarcode_Render_SuccessPath(t *testing.T) {
	b := NewCodabarBarcode()
	if err := b.Encode("A12345B"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code2of5IndustrialBarcode.Render — GetPattern error path
// ---------------------------------------------------------------------------

func TestCode2of5IndustrialBarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode2of5IndustrialBarcode()
	b.encodedText = "   "
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for whitespace-only input")
	}
}

// ---------------------------------------------------------------------------
// Code2of5MatrixBarcode.Render — GetPattern error path
// ---------------------------------------------------------------------------

func TestCode2of5MatrixBarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode2of5MatrixBarcode()
	b.encodedText = "   "
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for whitespace-only input")
	}
}

// ---------------------------------------------------------------------------
// EAN8Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestEAN8Barcode_Render_FullPath(t *testing.T) {
	b := NewEAN8Barcode()
	if err := b.Encode("12345670"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCABarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestUPCABarcode_Render_FullPath(t *testing.T) {
	b := NewUPCABarcode()
	if err := b.Encode("01234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCEBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestUPCEBarcode_Render_FullPath(t *testing.T) {
	b := NewUPCEBarcode()
	if err := b.Encode("1234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code93ExtendedBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestCode93ExtendedBarcode_Render_FullPath(t *testing.T) {
	b := NewCode93ExtendedBarcode()
	if err := b.Encode("Hello93"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code128A/B/C Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestCode128ABarcode_Render_FullPath(t *testing.T) {
	b := NewCode128ABarcode()
	if err := b.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestCode128BBarcode_Render_FullPath(t *testing.T) {
	b := NewCode128BBarcode()
	if err := b.Encode("World"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestCode128CBarcode_Render_FullPath(t *testing.T) {
	b := NewCode128CBarcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// ITF14Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestITF14Barcode_Render_FullPath(t *testing.T) {
	b := NewITF14Barcode()
	if err := b.Encode("12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// DeutscheIdentcodeBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestDeutscheIdentcodeBarcode_Render_FullPath(t *testing.T) {
	b := NewDeutscheIdentcodeBarcode()
	if err := b.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// DeutscheLeitcodeBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestDeutscheLeitcodeBarcode_Render_FullPath(t *testing.T) {
	b := NewDeutscheLeitcodeBarcode()
	if err := b.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Supplement2Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestSupplement2Barcode_Render_FullPath(t *testing.T) {
	b := NewSupplement2Barcode()
	if err := b.Encode("53"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 50)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Supplement5Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestSupplement5Barcode_Render_FullPath(t *testing.T) {
	b := NewSupplement5Barcode()
	if err := b.Encode("52495"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(100, 50)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code39ExtendedBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestCode39ExtendedBarcode_Render_FullPath(t *testing.T) {
	b := NewCode39ExtendedBarcode()
	if err := b.Encode("Hello-World"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCE0Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestUPCE0Barcode_Render_FullPath(t *testing.T) {
	b := NewUPCE0Barcode()
	if err := b.Encode("01234565"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCE1Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestUPCE1Barcode_Render_FullPath(t *testing.T) {
	b := NewUPCE1Barcode()
	if err := b.Encode("11234565"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// GS1_128Barcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestGS1_128Barcode_Render_FullPath(t *testing.T) {
	b := NewGS1_128Barcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// JapanPost4StateBarcode.Render — verify full path
// ---------------------------------------------------------------------------

func TestJapanPost4StateBarcode_Render_FullPath(t *testing.T) {
	b := NewJapanPost4StateBarcode()
	if err := b.Encode("597-8615-5-7-6"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// SwissQRBarcode.Render — success path via internal access
// ---------------------------------------------------------------------------

func TestSwissQRBarcode_Render_FullPath(t *testing.T) {
	b := NewSwissQRBarcode()
	b.IBAN = "CH4431999123000889012"
	b.CreditorName = "Robert Schneider AG"
	b.Currency = "CHF"
	b.Amount = "199.95"
	b.Reference = ""
	b.RefType = "NON"
	if err := b.Encode(""); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 200)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ===========================================================================
// GetPattern error paths — trigger the `if err != nil { return nil, err }`
// branch in Render for types where GetPattern can fail.
// ===========================================================================

// ---------------------------------------------------------------------------
// Code128Barcode.Render — GetPattern error via high-byte character
// (c128AutoEncode may produce valid encoding, so we bypass Encode and set
// encodedText directly to a value that confuses code128GetPattern.)
// ---------------------------------------------------------------------------

func TestCode128Barcode_Render_GetPatternError(t *testing.T) {
	b := NewCode128Barcode()
	// Set encodedText directly with a byte that c128AutoEncode maps to a
	// sequence code128GetPattern cannot find (byte 0xFF > 127).
	b.encodedText = string([]byte{0xFF})
	_, err := b.Render(200, 100)
	// code128GetPattern may or may not error for this input depending on
	// c128AutoEncode behavior. Either way the method should not panic.
	_ = err
}

// ---------------------------------------------------------------------------
// Code128CBarcode.Render — GetPattern error via odd-length non-digit input
// ---------------------------------------------------------------------------

func TestCode128CBarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode128CBarcode()
	// Bypass Encode; set encodedText to a single non-digit character.
	// Code128C requires pairs of digits; a single letter will cause
	// c128FindC to return -1 → code128GetPattern returns error.
	b.encodedText = "X"
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for non-digit Code128C input")
	}
}

// ---------------------------------------------------------------------------
// Code128ABarcode.Render — GetPattern error via high-byte character
// ---------------------------------------------------------------------------

func TestCode128ABarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode128ABarcode()
	// Code128A only supports ASCII 0-95 (control + uppercase). A byte
	// value > 127 will cause c128FindA to return -1 after c128StripControlCodes.
	b.encodedText = string([]byte{0xFF})
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for high-byte Code128A input")
	}
}

// ---------------------------------------------------------------------------
// Code128BBarcode.Render — GetPattern error via high-byte character
// ---------------------------------------------------------------------------

func TestCode128BBarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode128BBarcode()
	// Code128B supports ASCII 32-127. A byte > 127 will fail.
	b.encodedText = string([]byte{0xFF})
	_, err := b.Render(200, 100)
	if err == nil {
		t.Error("expected error from GetPattern for high-byte Code128B input")
	}
}

// ---------------------------------------------------------------------------
// EAN8Barcode.Render — GetPattern never returns error (returns nil always),
// so the `if err != nil` branch is unreachable. Test validates success.
// ---------------------------------------------------------------------------

func TestEAN8Barcode_Render_VarLen(t *testing.T) {
	b := NewEAN8Barcode()
	b.encodedText = "12345678"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// EAN13Barcode.Render — GetPattern never returns error. Validate success.
// ---------------------------------------------------------------------------

func TestEAN13Barcode_Render_VarLen(t *testing.T) {
	b := NewEAN13Barcode()
	b.encodedText = "5901234123457"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code39Barcode.Render — GetPattern never returns error. Validate success.
// ---------------------------------------------------------------------------

func TestCode39Barcode_Render_MixedContent(t *testing.T) {
	b := NewCode39Barcode()
	b.encodedText = "A-B.C 1"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// CodabarBarcode.Render — GetPattern never returns error. Validate success.
// ---------------------------------------------------------------------------

func TestCodabarBarcode_Render_ShortInput(t *testing.T) {
	b := NewCodabarBarcode()
	// Short input triggers the < 2 len guard in GetPattern.
	b.encodedText = "5"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCABarcode.Render — GetPattern never returns error. Validate success.
// ---------------------------------------------------------------------------

func TestUPCABarcode_Render_12Digits(t *testing.T) {
	b := NewUPCABarcode()
	b.encodedText = "012345678905"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCEBarcode.Render — GetPattern delegates to UPCE0Barcode. Validate success.
// ---------------------------------------------------------------------------

func TestUPCEBarcode_Render_6Digits(t *testing.T) {
	b := NewUPCEBarcode()
	b.encodedText = "123456"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code93ExtendedBarcode.Render — GetPattern error via out-of-range char
// ---------------------------------------------------------------------------

func TestCode93ExtendedBarcode_Render_GetPatternError(t *testing.T) {
	b := NewCode93ExtendedBarcode()
	// Characters > 127 are skipped in expansion, resulting in empty string.
	// However, code93GetPattern will produce a valid (empty data) barcode
	// with just start/stop and check digits for empty input. But let's verify
	// no panic occurs.
	b.encodedText = string([]byte{0x80, 0x90, 0xFF})
	_, err := b.Render(200, 100)
	// This should not panic. The error behavior depends on whether the
	// empty expansion triggers any issue in code93GetPattern.
	_ = err
}

// ---------------------------------------------------------------------------
// ITF14Barcode.Render — GetPattern never returns error. Validate success.
// ---------------------------------------------------------------------------

func TestITF14Barcode_Render_14DigitInput(t *testing.T) {
	b := NewITF14Barcode()
	b.encodedText = "12345678901231"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// DeutscheIdentcodeBarcode.Render — GetPattern never returns error.
// ---------------------------------------------------------------------------

func TestDeutscheIdentcodeBarcode_Render_11Digits(t *testing.T) {
	b := NewDeutscheIdentcodeBarcode()
	b.encodedText = "12345123456"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// DeutscheLeitcodeBarcode.Render — GetPattern never returns error.
// ---------------------------------------------------------------------------

func TestDeutscheLeitcodeBarcode_Render_13Digits(t *testing.T) {
	b := NewDeutscheLeitcodeBarcode()
	b.encodedText = "1234512312312"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Supplement2Barcode.Render — GetPattern never returns error.
// ---------------------------------------------------------------------------

func TestSupplement2Barcode_Render_Digits(t *testing.T) {
	b := NewSupplement2Barcode()
	b.encodedText = "12"
	img, err := b.Render(100, 50)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Supplement5Barcode.Render — GetPattern never returns error.
// ---------------------------------------------------------------------------

func TestSupplement5Barcode_Render_Digits(t *testing.T) {
	b := NewSupplement5Barcode()
	b.encodedText = "98765"
	img, err := b.Render(100, 50)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// Code39ExtendedBarcode.Render — GetPattern never returns error.
// ---------------------------------------------------------------------------

func TestCode39ExtendedBarcode_Render_LowASCII(t *testing.T) {
	b := NewCode39ExtendedBarcode()
	b.encodedText = "test123"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCE0Barcode.Render — GetPattern delegates to UPC-E pattern. Validate success.
// ---------------------------------------------------------------------------

func TestUPCE0Barcode_Render_Digits(t *testing.T) {
	b := NewUPCE0Barcode()
	b.encodedText = "01234565"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// UPCE1Barcode.Render — GetPattern delegates to UPC-E pattern. Validate success.
// ---------------------------------------------------------------------------

func TestUPCE1Barcode_Render_Digits(t *testing.T) {
	b := NewUPCE1Barcode()
	b.encodedText = "11234565"
	img, err := b.Render(200, 100)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

// ---------------------------------------------------------------------------
// GS1_128Barcode.Render — GetPattern error via completely invalid input
// ---------------------------------------------------------------------------

func TestGS1_128Barcode_Render_GetPatternError(t *testing.T) {
	b := NewGS1_128Barcode()
	// Set encodedText to a high-byte sequence that will confuse code128GetPattern
	// after GS1 parsing and auto-encoding.
	b.encodedText = string([]byte{0xFF, 0xFE, 0xFD})
	_, err := b.Render(200, 100)
	// May or may not error — depends on auto-encoder handling. Should not panic.
	_ = err
}

// ---------------------------------------------------------------------------
// JapanPost4StateBarcode.Render — delegates to Code128 pattern. Validate success.
// ---------------------------------------------------------------------------

func TestJapanPost4StateBarcode_Render_NumericOnly(t *testing.T) {
	b := NewJapanPost4StateBarcode()
	b.encodedText = "1234567"
	img, err := b.Render(200, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}
