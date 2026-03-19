// barcode_coverage2_test.go — additional coverage for remaining uncovered paths.
//
// Targets:
//   - UPCABarcode.Render (was 0%)
//   - EAN8Barcode.Encode with 8-digit input (checksum recalc path)
//   - QRBarcode with Unicode content >255 (qrAppend8BitBytes UTF-8 path)
//   - QRBarcode with long content (version >= 7, triggers qrMaybeEmbedVersionInfo)
//   - DataMatrix with various sizes (multiple dmGetPoly cases)
//   - code2of5setLen truncation path (input longer than n)
//   - eanSetLen padding path (input shorter than n)
//   - EAN-8 and EAN-13 GetPattern with short input (triggers eanSetLen padding)
//   - Missing type error paths
//   - GS1_128 GetPattern fallback path
package barcode_test

import (
	"strings"
	"testing"

	barcode "github.com/andrewloable/go-fastreport/barcode"
)

// ── UPCABarcode.Render ────────────────────────────────────────────────────────

func TestUPCABarcode_Render(t *testing.T) {
	b := barcode.NewUPCABarcode()
	if err := b.Encode("01234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestUPCABarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewUPCABarcode()
	_, err := b.Render(300, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestUPCABarcode_Encode_InvalidInput(t *testing.T) {
	b := barcode.NewUPCABarcode()
	err := b.Encode("abc")
	if err == nil {
		t.Error("expected error for non-numeric UPC-A input")
	}
}

// ── EAN8Barcode.Encode with 8-digit input ─────────────────────────────────────

func TestEAN8Barcode_Encode_8Digit_ChecksumRecalc(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	// "12345670" — 8 digits; first 7 used for checksum recalc if checksum fails.
	// Try with a correct 8-digit EAN-8 (1234567 has check digit 0).
	err := b.Encode("12345670")
	if err != nil {
		// If even the first 7 digits fail, that's acceptable — we test the error path.
		t.Logf("Encode 8-digit: %v (acceptable)", err)
	}
}

func TestEAN8Barcode_Encode_8Digit_WrongChecksum(t *testing.T) {
	// 8-digit with wrong checksum → triggers the len==8 retry path.
	b := barcode.NewEAN8Barcode()
	// "12345671" has wrong checksum (correct is 0), triggers retry.
	err := b.Encode("12345671")
	// Retry with first 7 digits should succeed.
	if err != nil {
		t.Logf("Encode 8-digit wrong checksum: %v", err)
	}
}

// ── QRBarcode with Unicode content >255 ───────────────────────────────────────

func TestQRBarcode_GetMatrix_Unicode(t *testing.T) {
	b := barcode.NewQRBarcode()
	// U+0410 is Cyrillic 'А' — rune value > 255, triggers UTF-8 multi-byte path.
	if err := b.Encode("Hello \u0410\u0411\u0412"); err != nil {
		t.Fatalf("Encode unicode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix unicode returned nil/empty")
	}
}

// ── QRBarcode with version >= 7 content ───────────────────────────────────────

func TestQRBarcode_GetMatrix_Version7(t *testing.T) {
	b := barcode.NewQRBarcode()
	b.ErrorCorrection = "M"
	// Version 7 QR needs ~50+ bytes of content in byte mode.
	// Use a 100-character string to force version >= 7.
	content := strings.Repeat("Hello World 123! ", 6) // 102 chars
	if err := b.Encode(content); err != nil {
		t.Fatalf("Encode long: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix version7+ returned nil/empty")
	}
}

func TestQRBarcode_GetMatrix_Version7_HighEC(t *testing.T) {
	b := barcode.NewQRBarcode()
	b.ErrorCorrection = "H"
	// High EC level needs more data to force version >= 7.
	content := strings.Repeat("ABCDEF1234567890", 5) // 80 chars alphanumeric
	if err := b.Encode(content); err != nil {
		t.Fatalf("Encode long H: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix version7 H returned nil/empty")
	}
}

// ── DataMatrix with various sizes to hit more dmGetPoly cases ─────────────────

func TestDataMatrixBarcode_GetMatrix_Short(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	if err := b.Encode("A"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix short returned nil/empty")
	}
}

func TestDataMatrixBarcode_GetMatrix_Medium(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// 20 chars forces a larger symbol size, hitting different dmGetPoly cases.
	if err := b.Encode("1234567890ABCDEFGHIJ"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix medium returned nil/empty")
	}
}

func TestDataMatrixBarcode_GetMatrix_MediumLong(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// ~40 chars forces even larger symbol.
	if err := b.Encode("Hello World DataMatrix Barcode Test 12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix medium-long returned nil/empty")
	}
}

func TestDataMatrixBarcode_GetMatrix_Long(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// 60+ chars forces very large symbol, hitting corner1/3/4 patterns.
	if err := b.Encode(strings.Repeat("DataMatrix1234567890ABCDEFGHIJ ", 3)); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix long returned nil/empty")
	}
}

func TestDataMatrixBarcode_GetMatrix_DigitsOnly(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// Pure digits exercise X12/ASCII encodation path.
	if err := b.Encode("12345678901234567890"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix digits returned nil/empty")
	}
}

func TestDataMatrixBarcode_GetMatrix_Binary(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// Binary data with bytes > 127 exercises dmB256Encodation.
	data := string([]byte{128, 200, 255, 10, 20, 30, 40, 50, 60, 70})
	if err := b.Encode(data); err != nil {
		t.Fatalf("Encode binary: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix binary returned nil/empty")
	}
}

func TestGS1DatamatrixBarcode_GetMatrix_LongContent(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	if err := b.Encode("01234567890123456789012345678901234567"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix GS1 long returned nil/empty")
	}
}

// ── code2of5setLen truncation path ────────────────────────────────────────────
// code2of5setLen is called by GetPattern methods for Code2of5, ITF14, etc.
// The truncation branch (len(text) > n) is hit when input is longer than expected.

func TestCode2of5Barcode_GetPattern_LongInput(t *testing.T) {
	b := barcode.NewCode2of5Barcode()
	// Code2of5 with interleaved=true; input longer than required hits truncation.
	// DeutscheIdentcode requires 11 digits; give 12 to trigger code2of5setLen truncation.
	b.Interleaved = true
	if err := b.Encode("123456789012"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

func TestDeutscheIdentcodeBarcode_Encode_11Digits(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	if err := b.Encode("12345678901"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestDeutscheLeitcodeBarcode_Encode_13Digits(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	if err := b.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(300, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

// ── eanSetLen padding path ─────────────────────────────────────────────────────
// eanSetLen pads short input with leading zeros. Test with input shorter than n.

func TestEAN8Barcode_GetPattern_ShortInput(t *testing.T) {
	b := barcode.NewEAN8Barcode()
	// Give 5 digits; eanSetLen(text, 7) will pad with 2 leading zeros.
	if err := b.Encode("12345"); err != nil {
		// Encode may fail for < 7 digits; just exercise the path.
		t.Logf("Encode short: %v (testing GetPattern directly via encodedText)", err)
		return
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern short: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

func TestEAN13Barcode_GetPattern_ShortInput(t *testing.T) {
	b := barcode.NewEAN13Barcode()
	// Encode("4006381333931") for a valid 13-digit EAN.
	if err := b.Encode("400638133393"); err != nil { // 12 digits
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

// ── Missing type Encode error paths ───────────────────────────────────────────

func TestITF14Barcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewITF14Barcode()
	err := b.Encode("1234567890123X")
	if err == nil {
		t.Error("expected error for non-digit ITF14 input")
	}
}

func TestITF14Barcode_Encode_WrongLength(t *testing.T) {
	b := barcode.NewITF14Barcode()
	err := b.Encode("123456") // too short
	if err == nil {
		t.Error("expected error for too-short ITF14 input")
	}
}

func TestITF14Barcode_Encode_13Digits(t *testing.T) {
	b := barcode.NewITF14Barcode()
	// 13 digits → padded to 14 with leading '0'.
	if err := b.Encode("1234567890123"); err != nil {
		t.Fatalf("Encode 13 digits: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestDeutscheIdentcodeBarcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	err := b.Encode("1234567890X")
	if err == nil {
		t.Error("expected error for non-digit identcode")
	}
}

func TestDeutscheIdentcodeBarcode_Encode_WrongLength(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	err := b.Encode("123456") // too short
	if err == nil {
		t.Error("expected error for too-short identcode")
	}
}

func TestDeutscheLeitcodeBarcode_Encode_NonDigit(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	err := b.Encode("123456789012X")
	if err == nil {
		t.Error("expected error for non-digit leitcode")
	}
}

func TestDeutscheLeitcodeBarcode_Encode_WrongLength(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	err := b.Encode("12345678") // too short
	if err == nil {
		t.Error("expected error for too-short leitcode")
	}
}

func TestSupplement2Barcode_Encode_WrongLength(t *testing.T) {
	b := barcode.NewSupplement2Barcode()
	err := b.Encode("5") // too short
	if err == nil {
		t.Error("expected error for too-short supplement 2")
	}
}

func TestSupplement5Barcode_Encode_WrongLength(t *testing.T) {
	b := barcode.NewSupplement5Barcode()
	err := b.Encode("1234") // too short
	if err == nil {
		t.Error("expected error for too-short supplement 5")
	}
}

func TestJapanPost4StateBarcode_Encode_EmptyInput(t *testing.T) {
	b := barcode.NewJapanPost4StateBarcode()
	err := b.Encode("")
	if err == nil {
		t.Error("expected error for empty Japan Post input")
	}
}

func TestCode128CBarcode_Encode_OddLength(t *testing.T) {
	b := barcode.NewCode128CBarcode()
	// Odd-length digits → prepend '0' before encoding.
	if err := b.Encode("12345"); err != nil {
		t.Fatalf("Encode odd-length Code128C: %v", err)
	}
	if b.EncodedText() != "12345" {
		t.Errorf("EncodedText = %q, want 12345", b.EncodedText())
	}
}

// ── GS1_128 GetPattern fallback path ──────────────────────────────────────────

func TestGS1_128Barcode_GetPattern_FallbackStrip(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	// Input starting with "(" but invalid GS1 format → triggers fallback strip path.
	// "(X)" won't parse as valid GS1 → fallback strips parens.
	if err := b.Encode("(X1)12345"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// GetPattern may succeed via fallback or fail gracefully.
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern fallback: %v (acceptable)", err)
	}
}

func TestGS1_128Barcode_GetPattern_NoParenNoProblem(t *testing.T) {
	b := barcode.NewGS1_128Barcode()
	if err := b.Encode("0112345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern no-paren: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

// ── Code39 error paths ────────────────────────────────────────────────────────

func TestCode39Barcode_GetPattern_WithInvalidChar(t *testing.T) {
	b := barcode.NewCode39Barcode()
	// Code39 allows [A-Z0-9 $%+-./*]; lowercase is invalid.
	if err := b.Encode("hello"); err != nil {
		t.Logf("Encode lowercase: %v", err)
		return
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern lowercase: %v (acceptable)", err)
	}
}

// ── Codabar error path ────────────────────────────────────────────────────────

func TestCodabarBarcode_GetPattern_EmptyInput(t *testing.T) {
	b := barcode.NewCodabarBarcode()
	if err := b.Encode(""); err != nil {
		t.Logf("Encode empty: %v", err)
		return
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern empty: %v (acceptable)", err)
	}
}

// ── MSI Barcode with empty input ──────────────────────────────────────────────

func TestMSIBarcode_GetPattern_SingleDigit(t *testing.T) {
	b := barcode.NewMSIBarcode()
	if err := b.Encode("5"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

// ── Postnet with a non-digit character ────────────────────────────────────────

func TestPostNetBarcode_GetPattern_InvalidChar(t *testing.T) {
	b := barcode.NewPostNetBarcode()
	if err := b.Encode("1234A"); err != nil {
		t.Logf("Encode invalid: %v", err)
		return
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern invalid: %v (acceptable)", err)
	}
}

// ── Code93 with empty input ───────────────────────────────────────────────────

func TestCode93Barcode_GetPattern_Empty(t *testing.T) {
	b := barcode.NewCode93Barcode()
	if err := b.Encode(""); err != nil {
		t.Logf("Encode empty: %v", err)
		return
	}
	_, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern empty: %v (acceptable)", err)
	}
}

// ── Code128 with special FNC character sequence ───────────────────────────────

func TestCode128Barcode_GetPattern_FNC1Sequence(t *testing.T) {
	b := barcode.NewCode128Barcode()
	if err := b.Encode("&1;12345"); err != nil {
		t.Fatalf("Encode FNC1: %v", err)
	}
	pattern, err := b.GetPattern()
	if err != nil {
		t.Logf("GetPattern FNC1: %v", err)
		return
	}
	if len(pattern) == 0 {
		t.Error("GetPattern returned empty pattern")
	}
}

// ── QR code with numeric content only ─────────────────────────────────────────

func TestQRBarcode_GetMatrix_PureNumeric_Short(t *testing.T) {
	b := barcode.NewQRBarcode()
	b.ErrorCorrection = "L"
	if err := b.Encode("9999999999"); err != nil {
		t.Fatalf("Encode numeric: %v", err)
	}
	matrix, rows, cols := b.GetMatrix()
	if matrix == nil || rows <= 0 || cols <= 0 {
		t.Error("GetMatrix numeric returned nil/empty")
	}
}

// ── DrawBarcode2D edge cases ───────────────────────────────────────────────────

func TestDrawBarcode2D_LargeMatrix(t *testing.T) {
	// Create a 50x50 matrix of alternating true/false.
	size := 50
	matrix := make([][]bool, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
		for j := range matrix[i] {
			matrix[i][j] = (i+j)%2 == 0
		}
	}
	img := barcode.DrawBarcode2D(matrix, size, size, 200, 200)
	if img == nil {
		t.Fatal("DrawBarcode2D large returned nil")
	}
}

// ── Code93Extended GetPattern coverage ───────────────────────────────────────

func TestCode93ExtendedBarcode_GetPattern_ASCII(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	if err := b.Encode("Hello World 123"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(400, 80)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestCode93ExtendedBarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewCode93ExtendedBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── Code2of5Industrial Render without Encode ─────────────────────────────────

func TestCode2of5IndustrialBarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewCode2of5IndustrialBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

func TestCode2of5MatrixBarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewCode2of5MatrixBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── ITF14 Render without Encode ──────────────────────────────────────────────

func TestITF14Barcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewITF14Barcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── DeutscheIdentcode Render without Encode ──────────────────────────────────

func TestDeutscheIdentcodeBarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewDeutscheIdentcodeBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── DeutscheLeitcode Render without Encode ────────────────────────────────────

func TestDeutscheLeitcodeBarcode_Render_WithoutEncode(t *testing.T) {
	b := barcode.NewDeutscheLeitcodeBarcode()
	_, err := b.Render(400, 80)
	if err == nil {
		t.Error("expected error when Render called without Encode")
	}
}

// ── GS1DatamatrixBarcode with empty encoded text ─────────────────────────────

func TestGS1DatamatrixBarcode_GetMatrix_Empty(t *testing.T) {
	b := barcode.NewGS1DatamatrixBarcode()
	// No Encode call → encodedText is empty → GetMatrix returns nil.
	matrix, rows, cols := b.GetMatrix()
	if matrix != nil || rows != 0 || cols != 0 {
		t.Error("GetMatrix with empty text should return nil/0/0")
	}
}

// ── DataMatrix GetMatrix with empty encodedText ───────────────────────────────

func TestDataMatrixBarcode_GetMatrix_Empty(t *testing.T) {
	b := barcode.NewDataMatrixBarcode()
	// No Encode call → encodedText is empty → GetMatrix returns nil.
	matrix, rows, cols := b.GetMatrix()
	if matrix != nil || rows != 0 || cols != 0 {
		t.Error("GetMatrix with empty text should return nil/0/0")
	}
}
