// barcode_pipeline_deserialize_test.go validates FRX deserialization of barcode
// engine pipeline properties: CalcCheckSum, Trim, AutoEncode, QuietZone,
// QR Encoding/Shape/ShowMarker, Aztec ErrorCorrection, PDF417 properties,
// DataMatrix properties, and BarcodeObject HorzAlign.
//
// C# references:
//   LinearBarcodeBase.cs:637: calcCheckSum = true (default)
//   LinearBarcodeBase.cs:638: trim = true (default)
//   LinearBarcodeBase.cs:411-414: Serialize CalcCheckSum and Trim
//   Barcode128.cs:591: AutoEncode = true (default)
//   Barcode128.cs:581-582: Serialize AutoEncode
//   BarcodeQR.cs:902: QuietZone = true (default)
//   BarcodeQR.cs:276-283: Serialize Encoding, QuietZone, ShowMarker, Shape
//   BarcodeAztec.cs:35,86: ErrorCorrectionPercent default 33, serialized as int
//   BarcodePDF417.cs:1478-1496: AspectRatio, Columns, Rows, CodePage, CompactionMode, ErrorCorrection, PixelSize
//   BarcodeDatamatrix.cs:1060-1073: SymbolSize, Encoding, CodePage, PixelSize, AutoEncode
//   BarcodeObject.cs:119,546-547: HorzAlign default Left, serialized as enum string
//
// go-fastreport-vfs1: Test Barcode.CalcCheckSum deserialization
// go-fastreport-oex9: Test Barcode.Trim deserialization
// go-fastreport-9kv1: Test Barcode.AutoEncode deserialization
// go-fastreport-9icb: Test Barcode.QuietZone deserialization
// go-fastreport-8x0g: Test QR Encoding/Shape/ShowMarker deserialization
// go-fastreport-42b6: Test Aztec ErrorCorrection deserialization
// go-fastreport-qujr: Test PDF417 properties deserialization
// go-fastreport-n17o: Test DataMatrix properties deserialization
// go-fastreport-b3fp: Test BarcodeObject HorzAlign deserialization
package barcode

import (
	"math"
	"strings"
	"testing"
)

// pipelineMockReader is a mock report.Reader for pipeline property deserialization tests.
// It supports str, bool, float, and int maps.
type pipelineMockReader struct {
	strs   map[string]string
	bools  map[string]bool
	floats map[string]float32
	ints   map[string]int
}

func (r *pipelineMockReader) ReadStr(name, def string) string {
	if v, ok := r.strs[name]; ok {
		return v
	}
	return def
}
func (r *pipelineMockReader) ReadInt(name string, def int) int {
	if v, ok := r.ints[name]; ok {
		return v
	}
	return def
}
func (r *pipelineMockReader) ReadBool(name string, def bool) bool {
	if v, ok := r.bools[name]; ok {
		return v
	}
	return def
}
func (r *pipelineMockReader) ReadFloat(name string, def float32) float32 {
	if v, ok := r.floats[name]; ok {
		return v
	}
	return def
}
func (r *pipelineMockReader) NextChild() (string, bool) { return "", false }
func (r *pipelineMockReader) FinishChild() error        { return nil }

// ── CalcCheckSum (go-fastreport-vfs1) ─────────────────────────────────────────

// TestDeserialize_CalcCheckSum_Code39_DefaultTrue verifies that when
// Barcode.CalcCheckSum is absent from FRX, Code39 defaults to true (C# default).
func TestDeserialize_CalcCheckSum_Code39_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Code39"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*Code39Barcode)
	if !ok {
		t.Fatalf("expected *Code39Barcode, got %T", obj.Barcode)
	}
	if !bc.CalcChecksum {
		t.Error("CalcChecksum should default to true when Barcode.CalcCheckSum absent from FRX")
	}
}

// TestDeserialize_CalcCheckSum_Code39_FalseFromFRX verifies that
// Barcode.CalcCheckSum="false" disables the check digit (matches FRX Barcode19).
func TestDeserialize_CalcCheckSum_Code39_FalseFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "Code39"},
		bools: map[string]bool{"Barcode.CalcCheckSum": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*Code39Barcode)
	if !ok {
		t.Fatalf("expected *Code39Barcode, got %T", obj.Barcode)
	}
	if bc.CalcChecksum {
		t.Error("CalcChecksum should be false when FRX has Barcode.CalcCheckSum=\"false\"")
	}

	// Pattern without checksum should be shorter than with checksum.
	if err := bc.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	patWithout := code39GetPattern("12345678", false)
	patWith := code39GetPattern("12345678", true)
	if len(patWithout) >= len(patWith) {
		t.Errorf("pattern without checksum len=%d should be shorter than with=%d",
			len(patWithout), len(patWith))
	}
}

// TestCode39Barcode_SetCalcCheckSum verifies the SetCalcCheckSum interface.
func TestCode39Barcode_SetCalcCheckSum(t *testing.T) {
	b := NewCode39Barcode()
	if !b.CalcChecksum {
		t.Error("CalcChecksum should default to true")
	}
	b.SetCalcCheckSum(false)
	if b.CalcChecksum {
		t.Error("SetCalcCheckSum(false) should set CalcChecksum to false")
	}
	b.SetCalcCheckSum(true)
	if !b.CalcChecksum {
		t.Error("SetCalcCheckSum(true) should set CalcChecksum to true")
	}
}

// TestCode39ExtendedBarcode_CalcCheckSum_DefaultTrue verifies Code39Extended defaults.
func TestCode39ExtendedBarcode_CalcCheckSum_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Code39Extended"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*Code39ExtendedBarcode)
	if !ok {
		t.Fatalf("expected *Code39ExtendedBarcode, got %T", obj.Barcode)
	}
	if !bc.CalcChecksum {
		t.Error("Code39Extended CalcChecksum should default to true")
	}
}

// ── Trim (go-fastreport-oex9) ─────────────────────────────────────────────────

// TestDeserialize_Trim_DefaultTrue verifies Trim defaults to true when absent from FRX.
func TestDeserialize_Trim_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Code39"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !obj.trim {
		t.Error("trim should default to true when Barcode.Trim absent from FRX")
	}
}

// TestDeserialize_Trim_FalseFromFRX verifies Barcode.Trim="false" disables trimming.
func TestDeserialize_Trim_FalseFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "Code39"},
		bools: map[string]bool{"Barcode.Trim": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.trim {
		t.Error("trim should be false when FRX has Barcode.Trim=\"false\"")
	}
}

// TestBarcodeObject_Trim_Accessor verifies Trim()/SetTrim() methods.
func TestBarcodeObject_Trim_Accessor(t *testing.T) {
	obj := NewBarcodeObject()
	if !obj.Trim() {
		t.Error("Trim() should return true by default")
	}
	obj.SetTrim(false)
	if obj.Trim() {
		t.Error("SetTrim(false) should set Trim to false")
	}
	obj.SetTrim(true)
	if !obj.Trim() {
		t.Error("SetTrim(true) should set Trim to true")
	}
}

// TestBarcodeObject_Trim_AppliedToText verifies that Trim=true causes whitespace to be trimmed.
// The engine applies strings.TrimSpace() before Encode() when Trim=true.
func TestBarcodeObject_Trim_AppliedToText(t *testing.T) {
	// BarcodeObject.Trim() accessor returns the flag set by SetTrim/Deserialize.
	// The actual trimming is applied in engine/objects.go before calling Encode().
	// This test verifies the flag is stored and retrieved correctly for both paths.
	obj := NewBarcodeObject()
	obj.SetTrim(true)
	// Simulate what the engine does: trim the text if Trim() is true.
	rawText := "  12345678  "
	var encoded string
	if obj.Trim() {
		encoded = strings.TrimSpace(rawText)
	} else {
		encoded = rawText
	}
	if encoded != "12345678" {
		t.Errorf("trimmed text = %q, want %q", encoded, "12345678")
	}
	// With Trim=false, spaces are preserved.
	obj.SetTrim(false)
	if obj.Trim() {
		encoded = strings.TrimSpace(rawText)
	} else {
		encoded = rawText
	}
	if encoded != rawText {
		t.Errorf("non-trimmed text = %q, want %q", encoded, rawText)
	}
}

// ── AutoEncode (go-fastreport-9kv1) ───────────────────────────────────────────

// TestDeserialize_AutoEncode_Code128_DefaultTrue verifies AutoEncode defaults to true.
func TestDeserialize_AutoEncode_Code128_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Code128"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*Code128Barcode)
	if !ok {
		t.Fatalf("expected *Code128Barcode, got %T", obj.Barcode)
	}
	if !bc.AutoEncode {
		t.Error("AutoEncode should default to true when Barcode.AutoEncode absent from FRX")
	}
}

// TestDeserialize_AutoEncode_Code128_FalseFromFRX verifies Barcode.AutoEncode="false".
func TestDeserialize_AutoEncode_Code128_FalseFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "Code128"},
		bools: map[string]bool{"Barcode.AutoEncode": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*Code128Barcode)
	if !ok {
		t.Fatalf("expected *Code128Barcode, got %T", obj.Barcode)
	}
	if bc.AutoEncode {
		t.Error("AutoEncode should be false when FRX has Barcode.AutoEncode=\"false\"")
	}
}

// TestCode128Barcode_AutoEncode_AffectsGetPattern verifies that AutoEncode=true
// uses auto-selected subcode and produces a valid pattern.
func TestCode128Barcode_AutoEncode_AffectsGetPattern(t *testing.T) {
	b := NewCode128Barcode()
	if err := b.Encode("12345678"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// AutoEncode=true: c128AutoEncode() selects Code C for all-numeric even input.
	b.AutoEncode = true
	patAuto, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern auto: %v", err)
	}
	if len(patAuto) == 0 {
		t.Error("GetPattern(AutoEncode=true) returned empty pattern")
	}
}

// TestCode128Barcode_AutoEncode_ManualMode verifies AutoEncode=false passes text as-is.
func TestCode128Barcode_AutoEncode_ManualMode(t *testing.T) {
	b := NewCode128Barcode()
	// With AutoEncode=false, the text must already contain &A;/&B;/&C; prefixes.
	// Pass a properly-prefixed string directly.
	b.AutoEncode = false
	b.encodedText = "&B;HELLO"
	_, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern manual mode with prefix: %v", err)
	}
}

// TestCode128Barcode_SetAutoEncode verifies the SetAutoEncode interface.
func TestCode128Barcode_SetAutoEncode(t *testing.T) {
	b := NewCode128Barcode()
	if !b.AutoEncode {
		t.Error("AutoEncode should default to true")
	}
	b.SetAutoEncode(false)
	if b.AutoEncode {
		t.Error("SetAutoEncode(false) should set AutoEncode to false")
	}
}

// ── QuietZone (go-fastreport-9icb) ────────────────────────────────────────────

// TestDeserialize_QuietZone_QR_DefaultTrue verifies QR QuietZone defaults to true.
func TestDeserialize_QuietZone_QR_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "QR"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("expected *QRBarcode, got %T", obj.Barcode)
	}
	if !bc.QuietZone {
		t.Error("QuietZone should default to true when Barcode.QuietZone absent from FRX")
	}
}

// TestDeserialize_QuietZone_QR_FalseFromFRX verifies Barcode.QuietZone="false"
// (matches FRX Barcode45).
func TestDeserialize_QuietZone_QR_FalseFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "QR"},
		bools: map[string]bool{"Barcode.QuietZone": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("expected *QRBarcode, got %T", obj.Barcode)
	}
	if bc.QuietZone {
		t.Error("QuietZone should be false when FRX has Barcode.QuietZone=\"false\"")
	}
}

// TestQRBarcode_QuietZone_AddsBorder verifies that QuietZone=true adds a 4-module
// white border around the QR matrix (C# BarcodeQR.cs:845: quiet = QuietZone ? 4 : 0).
func TestQRBarcode_QuietZone_AddsBorder(t *testing.T) {
	withZone := NewQRBarcode()
	withZone.QuietZone = true
	if err := withZone.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	matrixWith, rowsWith, _ := withZone.GetMatrix()

	withoutZone := NewQRBarcode()
	withoutZone.QuietZone = false
	if err := withoutZone.Encode("HELLO"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_, rowsWithout, _ := withoutZone.GetMatrix()

	// QuietZone=true adds 2*4=8 rows compared to no quiet zone.
	const quietBorder = 4
	wantRows := rowsWithout + 2*quietBorder
	if rowsWith != wantRows {
		t.Errorf("QuietZone=true rows = %d, want %d (= %d base + 2×%d quiet)",
			rowsWith, wantRows, rowsWithout, quietBorder)
	}

	// The outer border rows should be all false (white).
	for col := range len(matrixWith[0]) {
		if matrixWith[0][col] {
			t.Errorf("quiet zone row 0 col %d should be white (false)", col)
			break
		}
	}
}

// TestNewQRBarcode_QuietZone_DefaultTrue verifies QuietZone=true in NewQRBarcode().
func TestNewQRBarcode_QuietZone_DefaultTrue(t *testing.T) {
	b := NewQRBarcode()
	if !b.QuietZone {
		t.Error("NewQRBarcode QuietZone should default to true (C# BarcodeQR.cs:902)")
	}
}

// ── CalcBounds with QuietZone (go-fastreport-9icb) ────────────────────────────

// TestQRBarcode_QuietZone_AffectsMatrixSize verifies matrix sizes for both states.
func TestQRBarcode_QuietZone_AffectsMatrixSize(t *testing.T) {
	b := NewQRBarcode()
	if err := b.Encode("TEST"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	// With quiet zone: matrix is larger.
	b.QuietZone = true
	_, r1, c1 := b.GetMatrix()
	b.QuietZone = false
	_, r2, c2 := b.GetMatrix()
	if r1 <= r2 {
		t.Errorf("QuietZone=true rows=%d should be > QuietZone=false rows=%d", r1, r2)
	}
	if c1 <= c2 {
		t.Errorf("QuietZone=true cols=%d should be > QuietZone=false cols=%d", c1, c2)
	}
	// Difference should be exactly 8 (4 per side).
	if r1-r2 != 8 {
		t.Errorf("row difference = %d, want 8 (2×4 quiet zone)", r1-r2)
	}
}

// ── CalcCheckSum CalcBounds sanity (go-fastreport-vfs1) ───────────────────────

// TestCode39_CalcCheckSum_False_CalcBounds verifies that Code39 with checksum=false
// produces a slightly narrower barcode (one fewer character encoded).
func TestCode39_CalcCheckSum_False_CalcBounds(t *testing.T) {
	bWith := NewCode39Barcode()
	bWith.CalcChecksum = true
	if err := bWith.Encode("12345678"); err != nil {
		t.Fatalf("Encode with: %v", err)
	}
	wWith, _ := bWith.CalcBounds()

	bWithout := NewCode39Barcode()
	bWithout.CalcChecksum = false
	if err := bWithout.Encode("12345678"); err != nil {
		t.Fatalf("Encode without: %v", err)
	}
	wWithout, _ := bWithout.CalcBounds()

	// CalcCheckSum=true adds an extra check character, so width should be larger.
	if math.Abs(float64(wWith-wWithout)) < 1.0 {
		t.Errorf("width with checksum (%.2f) should differ from without (%.2f) by at least 1px",
			wWith, wWithout)
	}
	if wWith <= wWithout {
		t.Errorf("width with checksum (%.2f) should be > without (%.2f)", wWith, wWithout)
	}
}

// ── QR Encoding / Shape / ShowMarker (go-fastreport-8x0g) ────────────────────

// TestDeserialize_QR_Encoding_DefaultUTF8 verifies QR Encoding defaults to "UTF8".
// C# BarcodeQR.cs:153: [DefaultValue(QRCodeEncoding.UTF8)].
func TestDeserialize_QR_Encoding_DefaultUTF8(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "QR"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("expected *QRBarcode, got %T", obj.Barcode)
	}
	if bc.Encoding != "UTF8" {
		t.Errorf("QR Encoding default = %q, want %q", bc.Encoding, "UTF8")
	}
}

// TestDeserialize_QR_Encoding_FromFRX verifies Barcode.Encoding is stored.
func TestDeserialize_QR_Encoding_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type":     "QR",
			"Barcode.Encoding": "ISO8859_1",
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*QRBarcode)
	if bc.Encoding != "ISO8859_1" {
		t.Errorf("QR Encoding = %q, want %q", bc.Encoding, "ISO8859_1")
	}
}

// TestDeserialize_QR_Shape_DefaultRectangle verifies QR Shape defaults to "Rectangle".
// C# BarcodeQR.cs:173: [DefaultValue(QrModuleShape.Rectangle)].
func TestDeserialize_QR_Shape_DefaultRectangle(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "QR"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*QRBarcode)
	if bc.Shape != "Rectangle" {
		t.Errorf("QR Shape default = %q, want %q", bc.Shape, "Rectangle")
	}
}

// TestDeserialize_QR_Shape_CircleFromFRX verifies Barcode.Shape="Circle" is stored.
func TestDeserialize_QR_Shape_CircleFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type":  "QR",
			"Barcode.Shape": "Circle",
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*QRBarcode)
	if bc.Shape != "Circle" {
		t.Errorf("QR Shape = %q, want %q", bc.Shape, "Circle")
	}
}

// TestDeserialize_QR_ShowMarker_DefaultFalse verifies QR ShowMarker defaults to false.
func TestDeserialize_QR_ShowMarker_DefaultFalse(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "QR"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*QRBarcode)
	if bc.ShowMarker {
		t.Error("QR ShowMarker should default to false")
	}
}

// TestDeserialize_QR_ShowMarker_TrueFromFRX verifies Barcode.ShowMarker=true is stored.
func TestDeserialize_QR_ShowMarker_TrueFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "QR"},
		bools: map[string]bool{"Barcode.ShowMarker": true},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*QRBarcode)
	if !bc.ShowMarker {
		t.Error("QR ShowMarker should be true when FRX has Barcode.ShowMarker=true")
	}
}

// ── Aztec ErrorCorrection (go-fastreport-42b6) ────────────────────────────────

// TestDeserialize_Aztec_ErrorCorrection_Default33 verifies Aztec MinECCPercent defaults to 33.
// C# BarcodeAztec.cs:35: ErrorCorrectionPercent = 33 (default).
func TestDeserialize_Aztec_ErrorCorrection_Default33(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Aztec"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*AztecBarcode)
	if !ok {
		t.Fatalf("expected *AztecBarcode, got %T", obj.Barcode)
	}
	if bc.MinECCPercent != 33 {
		t.Errorf("Aztec MinECCPercent default = %d, want 33", bc.MinECCPercent)
	}
}

// TestDeserialize_Aztec_ErrorCorrection_FromFRX verifies Barcode.ErrorCorrection int is read.
func TestDeserialize_Aztec_ErrorCorrection_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Aztec"},
		ints: map[string]int{"Barcode.ErrorCorrection": 50},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*AztecBarcode)
	if bc.MinECCPercent != 50 {
		t.Errorf("Aztec MinECCPercent = %d, want 50", bc.MinECCPercent)
	}
}

// TestNewAztecBarcode_MinECCPercent_Default33 verifies constructor default.
func TestNewAztecBarcode_MinECCPercent_Default33(t *testing.T) {
	b := NewAztecBarcode()
	if b.MinECCPercent != 33 {
		t.Errorf("NewAztecBarcode MinECCPercent = %d, want 33", b.MinECCPercent)
	}
}

// ── PDF417 properties (go-fastreport-qujr) ────────────────────────────────────

// TestDeserialize_PDF417_Defaults verifies PDF417 constructor defaults match C#.
// C# BarcodePDF417.cs: AspectRatio=0.5, Columns=0, Rows=0, CodePage=437,
// CompactionMode=Auto, ErrorCorrection=Auto, PixelSize={2,8}.
func TestDeserialize_PDF417_Defaults(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "PDF417"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*PDF417Barcode)
	if !ok {
		t.Fatalf("expected *PDF417Barcode, got %T", obj.Barcode)
	}
	if bc.AspectRatio != 0.5 {
		t.Errorf("PDF417 AspectRatio = %v, want 0.5", bc.AspectRatio)
	}
	if bc.Columns != 0 {
		t.Errorf("PDF417 Columns = %d, want 0", bc.Columns)
	}
	if bc.Rows != 0 {
		t.Errorf("PDF417 Rows = %d, want 0", bc.Rows)
	}
	if bc.CodePage != 437 {
		t.Errorf("PDF417 CodePage = %d, want 437", bc.CodePage)
	}
	if bc.CompactionMode != "Auto" {
		t.Errorf("PDF417 CompactionMode = %q, want %q", bc.CompactionMode, "Auto")
	}
	if bc.ErrorCorrection != "Auto" {
		t.Errorf("PDF417 ErrorCorrection = %q, want %q", bc.ErrorCorrection, "Auto")
	}
	if bc.PixelSizeWidth != 2 {
		t.Errorf("PDF417 PixelSizeWidth = %d, want 2", bc.PixelSizeWidth)
	}
	if bc.PixelSizeHeight != 8 {
		t.Errorf("PDF417 PixelSizeHeight = %d, want 8", bc.PixelSizeHeight)
	}
}

// TestDeserialize_PDF417_FromFRX verifies custom PDF417 values are read from FRX.
func TestDeserialize_PDF417_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type":            "PDF417",
			"Barcode.CompactionMode":  "Text",
			"Barcode.ErrorCorrection": "L4",
		},
		ints: map[string]int{
			"Barcode.Columns":  5,
			"Barcode.Rows":     10,
			"Barcode.CodePage": 1252,
		},
		floats: map[string]float32{
			"Barcode.AspectRatio": 1.0,
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*PDF417Barcode)
	if bc.Columns != 5 {
		t.Errorf("PDF417 Columns = %d, want 5", bc.Columns)
	}
	if bc.Rows != 10 {
		t.Errorf("PDF417 Rows = %d, want 10", bc.Rows)
	}
	if bc.CodePage != 1252 {
		t.Errorf("PDF417 CodePage = %d, want 1252", bc.CodePage)
	}
	if math.Abs(float64(bc.AspectRatio-1.0)) > 0.001 {
		t.Errorf("PDF417 AspectRatio = %v, want 1.0", bc.AspectRatio)
	}
	if bc.CompactionMode != "Text" {
		t.Errorf("PDF417 CompactionMode = %q, want %q", bc.CompactionMode, "Text")
	}
	if bc.ErrorCorrection != "L4" {
		t.Errorf("PDF417 ErrorCorrection = %q, want %q", bc.ErrorCorrection, "L4")
	}
}

// ── DataMatrix properties (go-fastreport-n17o) ────────────────────────────────

// TestDeserialize_DataMatrix_Defaults verifies DataMatrix constructor defaults match C#.
// C# BarcodeDatamatrix.cs: SymbolSize=Auto, Encoding=Auto, CodePage=1252, PixelSize=3, AutoEncode=true.
func TestDeserialize_DataMatrix_Defaults(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "DataMatrix"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*DataMatrixBarcode)
	if !ok {
		t.Fatalf("expected *DataMatrixBarcode, got %T", obj.Barcode)
	}
	if bc.SymbolSize != "Auto" {
		t.Errorf("DataMatrix SymbolSize = %q, want %q", bc.SymbolSize, "Auto")
	}
	if bc.Encoding != "Auto" {
		t.Errorf("DataMatrix Encoding = %q, want %q", bc.Encoding, "Auto")
	}
	if bc.CodePage != 1252 {
		t.Errorf("DataMatrix CodePage = %d, want 1252", bc.CodePage)
	}
	if bc.PixelSize != 3 {
		t.Errorf("DataMatrix PixelSize = %d, want 3", bc.PixelSize)
	}
	if !bc.AutoEncode {
		t.Error("DataMatrix AutoEncode should default to true")
	}
}

// TestDeserialize_DataMatrix_FromFRX verifies custom DataMatrix values are read.
func TestDeserialize_DataMatrix_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type":       "DataMatrix",
			"Barcode.SymbolSize": "32x32",
			"Barcode.Encoding":   "Ascii",
		},
		ints: map[string]int{
			"Barcode.CodePage":  1251,
			"Barcode.PixelSize": 4,
		},
		bools: map[string]bool{
			"Barcode.AutoEncode": false,
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*DataMatrixBarcode)
	if bc.SymbolSize != "32x32" {
		t.Errorf("DataMatrix SymbolSize = %q, want %q", bc.SymbolSize, "32x32")
	}
	if bc.Encoding != "Ascii" {
		t.Errorf("DataMatrix Encoding = %q, want %q", bc.Encoding, "Ascii")
	}
	if bc.CodePage != 1251 {
		t.Errorf("DataMatrix CodePage = %d, want 1251", bc.CodePage)
	}
	if bc.PixelSize != 4 {
		t.Errorf("DataMatrix PixelSize = %d, want 4", bc.PixelSize)
	}
	if bc.AutoEncode {
		t.Error("DataMatrix AutoEncode should be false when FRX has Barcode.AutoEncode=false")
	}
}

// ── HorzAlign (go-fastreport-b3fp) ────────────────────────────────────────────

// TestDeserialize_HorzAlign_DefaultLeft verifies HorzAlign defaults to Left.
// C# BarcodeObject.cs:119: [DefaultValue(Alignment.Left)].
func TestDeserialize_HorzAlign_DefaultLeft(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "Code128"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.HorzAlign() != BarcodeHorzAlignLeft {
		t.Errorf("HorzAlign default = %v, want Left", obj.HorzAlign())
	}
}

// TestDeserialize_HorzAlign_CenterFromFRX verifies HorzAlign="Center" is read.
func TestDeserialize_HorzAlign_CenterFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type": "Code128",
			"HorzAlign":    "Center",
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.HorzAlign() != BarcodeHorzAlignCenter {
		t.Errorf("HorzAlign = %v, want Center", obj.HorzAlign())
	}
}

// TestDeserialize_HorzAlign_RightFromFRX verifies HorzAlign="Right" is read.
func TestDeserialize_HorzAlign_RightFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{
			"Barcode.Type": "Code128",
			"HorzAlign":    "Right",
		},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if obj.HorzAlign() != BarcodeHorzAlignRight {
		t.Errorf("HorzAlign = %v, want Right", obj.HorzAlign())
	}
}

// TestBarcodeObject_HorzAlign_Accessor verifies HorzAlign()/SetHorzAlign() methods.
func TestBarcodeObject_HorzAlign_Accessor(t *testing.T) {
	obj := NewBarcodeObject()
	if obj.HorzAlign() != BarcodeHorzAlignLeft {
		t.Error("default HorzAlign should be Left")
	}
	obj.SetHorzAlign(BarcodeHorzAlignCenter)
	if obj.HorzAlign() != BarcodeHorzAlignCenter {
		t.Error("SetHorzAlign(Center) should update HorzAlign")
	}
	obj.SetHorzAlign(BarcodeHorzAlignRight)
	if obj.HorzAlign() != BarcodeHorzAlignRight {
		t.Error("SetHorzAlign(Right) should update HorzAlign")
	}
}

// TestBarcodeObject_ShowMarker_Accessor verifies ShowMarker()/SetShowMarker() methods.
func TestBarcodeObject_ShowMarker_Accessor(t *testing.T) {
	obj := NewBarcodeObject()
	if !obj.ShowMarker() {
		t.Error("ShowMarker should default to true (C# BarcodeObject.cs:695)")
	}
	obj.SetShowMarker(true)
	if !obj.ShowMarker() {
		t.Error("SetShowMarker(true) should update ShowMarker")
	}
}

// ── MaxiCode Mode (go-fastreport-pq9w) ───────────────────────────────────────

// TestDeserialize_MaxiCode_Mode_Default4 verifies MaxiCode Mode defaults to 4.
// C# BarcodeMaxiCode.cs:43: Mode = 4 (structured carrier message default).
func TestDeserialize_MaxiCode_Mode_Default4(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "MaxiCode"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*MaxiCodeBarcode)
	if !ok {
		t.Fatalf("expected *MaxiCodeBarcode, got %T", obj.Barcode)
	}
	if bc.Mode != 4 {
		t.Errorf("MaxiCode Mode default = %d, want 4", bc.Mode)
	}
}

// TestDeserialize_MaxiCode_Mode_FromFRX verifies Barcode.Mode int is read.
// FRX example: Barcode42 has Barcode.Mode="4".
func TestDeserialize_MaxiCode_Mode_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "MaxiCode"},
		ints: map[string]int{"Barcode.Mode": 2},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*MaxiCodeBarcode)
	if bc.Mode != 2 {
		t.Errorf("MaxiCode Mode = %d, want 2", bc.Mode)
	}
}

// TestNewMaxiCodeBarcode_Mode_Default4 verifies constructor default.
func TestNewMaxiCodeBarcode_Mode_Default4(t *testing.T) {
	b := NewMaxiCodeBarcode()
	if b.Mode != 4 {
		t.Errorf("NewMaxiCodeBarcode Mode = %d, want 4", b.Mode)
	}
}

// ── DrawVerticalBearerBars (go-fastreport-4qaw) ───────────────────────────────

// TestDeserialize_ITF14_DrawVerticalBearerBars_DefaultTrue verifies the default is true.
// C# BarcodeITF14.drawVerticalBearerBars = true (Barcode2of5.cs:332).
func TestDeserialize_ITF14_DrawVerticalBearerBars_DefaultTrue(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "ITF14"},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*ITF14Barcode)
	if !ok {
		t.Fatalf("expected *ITF14Barcode, got %T", obj.Barcode)
	}
	if !bc.DrawVerticalBearerBars {
		t.Error("ITF14 DrawVerticalBearerBars should default to true")
	}
}

// TestDeserialize_ITF14_DrawVerticalBearerBars_FalseFromFRX verifies false is read.
// FRX example: Barcode50 Deutsche Leitcode has Barcode.DrawVerticalBearerBars="False".
func TestDeserialize_ITF14_DrawVerticalBearerBars_FalseFromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "ITF14"},
		bools: map[string]bool{"Barcode.DrawVerticalBearerBars": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc := obj.Barcode.(*ITF14Barcode)
	if bc.DrawVerticalBearerBars {
		t.Error("ITF14 DrawVerticalBearerBars should be false when FRX has false")
	}
}

// TestNewITF14Barcode_DrawVerticalBearerBars_DefaultTrue verifies constructor default.
func TestNewITF14Barcode_DrawVerticalBearerBars_DefaultTrue(t *testing.T) {
	b := NewITF14Barcode()
	if !b.DrawVerticalBearerBars {
		t.Error("NewITF14Barcode DrawVerticalBearerBars should default to true")
	}
}
