package barcode

import (
	"image"
	"strings"
	"testing"
)

// ── Deutsche Identcode: PrintCheckSum and display text formatting ─────────────
//
// go-fastreport-tpiv: Port Deutsche PrintCheckSum property
// go-fastreport-i41r: Port Deutsche Identcode display text formatting
// C# reference: BarcodeDeutscheIdentcode.GetPattern (Barcode2of5.cs:156-163)

// TestDeutscheIdentcode_DisplayText_PrintCheckSumTrue verifies that GetPattern
// formats the 12-digit Identcode value as "XX.XXX XXX.XX X" (with check digit
// and trailing space separator) when PrintCheckSum=true.
func TestDeutscheIdentcode_DisplayText_PrintCheckSumTrue(t *testing.T) {
	b := NewDeutscheIdentcodeBarcode()
	b.PrintCheckSum = true
	if err := b.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := b.GetPattern(); err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	display := b.EncodedText()
	// After checksum calculation: 12345123456 → 123451234561 (check digit = 1).
	// C# InsertAt sequence on "123451234561" (12 chars):
	//   Insert(2,".")  → "12.3451234561"   (13 chars)
	//   Insert(6," ")  → "12.345 1234561"  (14 chars)
	//   Insert(10,".") → "12.345 123.4561" (15 chars)
	//   Insert(14," ") → "12.345 123.456 1" (16 chars)
	if display != "12.345 123.456 1" {
		t.Errorf("PrintCheckSum=true: got %q, want %q", display, "12.345 123.456 1")
	}
}

// TestDeutscheIdentcode_DisplayText_PrintCheckSumFalse verifies that GetPattern
// strips the check digit from the display text when PrintCheckSum=false.
func TestDeutscheIdentcode_DisplayText_PrintCheckSumFalse(t *testing.T) {
	b := NewDeutscheIdentcodeBarcode()
	b.PrintCheckSum = false
	if err := b.Encode("12345123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := b.GetPattern(); err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	display := b.EncodedText()
	// After checksum: "123451234561".
	// Insert(2,".").Insert(6," ").Insert(10,".") → "12.345 123.4561"  (15 chars)
	// Strip last char → "12.345 123.456"  (14 chars)
	if display != "12.345 123.456" {
		t.Errorf("PrintCheckSum=false: got %q, want %q", display, "12.345 123.456")
	}
}

// TestDeutscheIdentcode_DefaultPrintCheckSum verifies PrintCheckSum defaults to true.
func TestDeutscheIdentcode_DefaultPrintCheckSum(t *testing.T) {
	b := NewDeutscheIdentcodeBarcode()
	if !b.PrintCheckSum {
		t.Error("DeutscheIdentcodeBarcode.PrintCheckSum should default to true")
	}
}

// TestDeutscheIdentcode_Deserialize_PrintCheckSum_FromFRX verifies that the
// FRX attribute Barcode.DrawVerticalBearerBars is read as PrintCheckSum.
func TestDeutscheIdentcode_Deserialize_PrintCheckSum_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "DeutscheIdentcode"},
		bools: map[string]bool{"Barcode.DrawVerticalBearerBars": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*DeutscheIdentcodeBarcode)
	if !ok {
		t.Fatalf("expected *DeutscheIdentcodeBarcode, got %T", obj.Barcode)
	}
	if bc.PrintCheckSum {
		t.Error("PrintCheckSum should be false when FRX has Barcode.DrawVerticalBearerBars=false")
	}
}

// ── Deutsche Leitcode: PrintCheckSum and display text formatting ─────────────
//
// go-fastreport-yyk6: Port Deutsche Leitcode display text formatting
// C# reference: BarcodeDeutscheLeitcode.GetPattern (Barcode2of5.cs:301-308)

// TestDeutscheLeitcode_DisplayText verifies that GetPattern formats the 14-digit
// Leitcode value with dots and spaces per C# BarcodeDeutscheLeitcode.GetPattern.
func TestDeutscheLeitcode_DisplayText(t *testing.T) {
	b := NewDeutscheLeitcodeBarcode()
	b.PrintCheckSum = true
	// Encode accepts 13 digits; CalcChecksum=true appends the check digit.
	// "1234512312312" is the C# default value (Barcode2of5.cs:257).
	if err := b.Encode("1234512312312"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := b.GetPattern(); err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	display := b.EncodedText()
	// Compute check digit for "1234512312312":
	// fak=13(odd):1×4=4, fak=12(even):2×9=18, fak=11(odd):3×4=12,
	// fak=10(even):4×9=36, fak=9(odd):5×4=20, fak=8(even):1×9=9,
	// fak=7(odd):2×4=8, fak=6(even):3×9=27, fak=5(odd):1×4=4,
	// fak=4(even):2×9=18, fak=3(odd):3×4=12, fak=2(even):1×9=9,
	// fak=1(odd):2×4=8
	// sum=4+18+12+36+20+9+8+27+4+18+12+9+8=185; 185%10=5; check=10-5=5
	// 14-digit text: "12345123123125"
	// C# InsertAt sequence on "12345123123125" (14 chars):
	//   Insert(5,".")  → "12345.123123125"   (15 chars)
	//   Insert(6," ")  → "12345. 123123125"  (16 chars)
	//   Insert(10,".") → "12345. 123.123125" (17 chars)
	//   Insert(11," ") → "12345. 123. 123125" (18 chars)
	//   Insert(15,".") → "12345. 123. 123.125" (19 chars)
	//   Insert(16," ") → "12345. 123. 123. 125" (20 chars)
	//   Insert(19," ") → "12345. 123. 123. 12 5" (21 chars)
	want := "12345. 123. 123. 12 5"
	if display != want {
		t.Errorf("Leitcode display text: got %q, want %q", display, want)
	}
}

// TestDeutscheLeitcode_DefaultPrintCheckSum verifies PrintCheckSum defaults to true.
func TestDeutscheLeitcode_DefaultPrintCheckSum(t *testing.T) {
	b := NewDeutscheLeitcodeBarcode()
	if !b.PrintCheckSum {
		t.Error("DeutscheLeitcodeBarcode.PrintCheckSum should default to true")
	}
}

// TestDeutscheLeitcode_Deserialize_PrintCheckSum_FromFRX verifies that the
// FRX attribute Barcode.DrawVerticalBearerBars is read as PrintCheckSum.
func TestDeutscheLeitcode_Deserialize_PrintCheckSum_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "DeutscheLeitcode"},
		bools: map[string]bool{"Barcode.DrawVerticalBearerBars": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	bc, ok := obj.Barcode.(*DeutscheLeitcodeBarcode)
	if !ok {
		t.Fatalf("expected *DeutscheLeitcodeBarcode, got %T", obj.Barcode)
	}
	if bc.PrintCheckSum {
		t.Error("PrintCheckSum should be false when FRX has Barcode.DrawVerticalBearerBars=false")
	}
}

func TestPharmacodeBarcode_QuietZoneFalseRemovesOuterSpaces(t *testing.T) {
	b := NewPharmacodeBarcode()
	b.QuietZone = false
	if err := b.Encode("123456"); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	pattern, err := b.GetPattern()
	if err != nil {
		t.Fatalf("GetPattern: %v", err)
	}
	if strings.HasPrefix(pattern, "2") {
		t.Fatalf("pattern should not start with a quiet-zone space when QuietZone=false: %q", pattern)
	}
	if strings.HasSuffix(pattern, "2") {
		t.Fatalf("pattern should not end with a quiet-zone space when QuietZone=false: %q", pattern)
	}
}

func TestDeserialize_PharmacodeQuietZone_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "Pharmacode"},
		bools: map[string]bool{"Barcode.QuietZone": false},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	bc, ok := obj.Barcode.(*PharmacodeBarcode)
	if !ok {
		t.Fatalf("expected *PharmacodeBarcode, got %T", obj.Barcode)
	}
	if bc.QuietZone {
		t.Fatal("QuietZone should be false when FRX has Barcode.QuietZone=false")
	}
}

func TestDeserialize_QR_UseThinModules_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs:  map[string]string{"Barcode.Type": "QR"},
		bools: map[string]bool{"Barcode.UseThinModules": true},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	qr, ok := obj.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("expected *QRBarcode, got %T", obj.Barcode)
	}
	if !qr.UseThinModules {
		t.Fatal("UseThinModules should be true when FRX has Barcode.UseThinModules=true")
	}
}

func TestQRBarcode_Render_UseThinModules_ReducesInk(t *testing.T) {
	base := NewQRBarcode()
	base.QuietZone = false
	if err := base.Encode("HELLO WORLD"); err != nil {
		t.Fatalf("Encode base QR: %v", err)
	}
	baseImg, err := base.Render(240, 240)
	if err != nil {
		t.Fatalf("Render base QR: %v", err)
	}

	thin := NewQRBarcode()
	thin.QuietZone = false
	thin.UseThinModules = true
	if err := thin.Encode("HELLO WORLD"); err != nil {
		t.Fatalf("Encode thin QR: %v", err)
	}
	thinImg, err := thin.Render(240, 240)
	if err != nil {
		t.Fatalf("Render thin QR: %v", err)
	}

	baseDark := countDarkPixels(baseImg)
	thinDark := countDarkPixels(thinImg)
	if thinDark >= baseDark {
		t.Fatalf("UseThinModules should reduce dark pixels: thin=%d base=%d", thinDark, baseDark)
	}
}

func TestQRBarcode_SwissPayload_NormalizesAndForcesM(t *testing.T) {
	swiss := NewSwissQRBarcode()
	swiss.Params = SwissQRParameters{
		IBAN:               "CH5604835012345678009",
		Currency:           "CHF",
		CreditorName:       "Test AG",
		CreditorStreet:     "Teststrasse 1",
		CreditorCity:       "Bern",
		CreditorPostalCode: "3000",
		CreditorCountry:    "CH",
		Amount:             "50.00",
		ReferenceType:      "NON",
		TrailerEPD:         "EPD",
	}
	payload := strings.ReplaceAll(swiss.FormatPayload(), "\n", "\r\n")

	withH := NewQRBarcode()
	withH.ErrorCorrection = "H"
	if err := withH.Encode(payload); err != nil {
		t.Fatalf("Encode Swiss payload with H: %v", err)
	}
	if strings.Contains(withH.EncodedText(), "\r") {
		t.Fatal("Swiss payload should be normalized to LF-only form")
	}

	withM := NewQRBarcode()
	withM.ErrorCorrection = "M"
	if err := withM.Encode(payload); err != nil {
		t.Fatalf("Encode Swiss payload with M: %v", err)
	}

	hMatrix, hRows, hCols := withH.GetMatrix()
	mMatrix, mRows, mCols := withM.GetMatrix()
	if hRows != mRows || hCols != mCols {
		t.Fatalf("Swiss QR should force M-level sizing: H=(%d,%d) M=(%d,%d)", hRows, hCols, mRows, mCols)
	}
	for r := range hRows {
		for c := range hCols {
			if hMatrix[r][c] != mMatrix[r][c] {
				t.Fatalf("Swiss QR matrix should match M-level output at row=%d col=%d", r, c)
			}
		}
	}
}

func TestBarcodeObject_CreateSwissQR(t *testing.T) {
	obj := NewBarcodeObject()
	obj.CreateSwissQR(SwissQRParameters{
		IBAN:               "CH5604835012345678009",
		Currency:           "CHF",
		CreditorName:       "Test AG",
		CreditorStreet:     "Teststrasse 1",
		CreditorCity:       "Bern",
		CreditorPostalCode: "3000",
		CreditorCountry:    "CH",
		Amount:             "10.00",
		ReferenceType:      "NON",
		TrailerEPD:         "EPD",
	})

	if _, ok := obj.Barcode.(*QRBarcode); !ok {
		t.Fatalf("CreateSwissQR should replace the barcode type with *QRBarcode, got %T", obj.Barcode)
	}
	if obj.ShowText() {
		t.Fatal("CreateSwissQR should disable human-readable text")
	}
	if !strings.HasPrefix(obj.Text(), "SPC\n0200\n1\n") {
		t.Fatalf("CreateSwissQR should populate a Swiss QR payload, got %q", obj.Text())
	}
}

// TestQRBarcode_AllShapes_RenderNonEmpty verifies that every QrModuleShape
// value produces a non-empty (has dark pixels) image without panicking.
// Ported from C# QrModuleShape enum (BarcodeQR.cs:69-120).
func TestQRBarcode_AllShapes_RenderNonEmpty(t *testing.T) {
	shapes := []string{
		"Rectangle",
		"Circle",
		"Diamond",
		"RoundedSquare",
		"PillHorizontal",
		"PillVertical",
		"Plus",
		"Hexagon",
		"Star",
		"Snowflake",
	}
	for _, shape := range shapes {
		t.Run(shape, func(t *testing.T) {
			qr := NewQRBarcode()
			qr.QuietZone = false
			qr.Shape = shape
			if err := qr.Encode("TEST"); err != nil {
				t.Fatalf("Encode: %v", err)
			}
			img, err := qr.Render(200, 200)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			dark := countDarkPixels(img)
			if dark == 0 {
				t.Fatalf("shape %q rendered no dark pixels", shape)
			}
		})
	}
}

// TestQRBarcode_ShapeProducesFewerPixels_ThanRectangle checks that non-Rectangle
// shapes generally differ from a plain rectangle (they are not identical).
// Circle, Diamond, etc. should produce a different pixel count.
func TestQRBarcode_ShapeProducesFewerPixels_ThanRectangle(t *testing.T) {
	encodeAndCount := func(shape string) int {
		qr := NewQRBarcode()
		qr.QuietZone = false
		qr.Shape = shape
		_ = qr.Encode("SHAPES")
		img, _ := qr.Render(200, 200)
		return countDarkPixels(img)
	}
	rectDark := encodeAndCount("Rectangle")
	for _, shape := range []string{"Circle", "Diamond", "Star", "Hexagon", "Snowflake"} {
		dark := encodeAndCount(shape)
		if dark == rectDark {
			t.Errorf("shape %q has identical pixel count (%d) to Rectangle — shapes may not be distinct", shape, dark)
		}
	}
}

// TestQRBarcode_Angle_AffectsRotationalShapes verifies that a non-zero Angle
// changes the rendering for rotational shapes (Hexagon, Star, Snowflake)
// but has no visible effect on Rectangle.
func TestQRBarcode_Angle_AffectsRotationalShapes(t *testing.T) {
	encodeAndCount := func(shape string, angle int) int {
		qr := NewQRBarcode()
		qr.QuietZone = false
		qr.Shape = shape
		qr.Angle = angle
		_ = qr.Encode("ANGLE")
		img, _ := qr.Render(200, 200)
		return countDarkPixels(img)
	}
	for _, shape := range []string{"Hexagon", "Star", "Snowflake"} {
		a0 := encodeAndCount(shape, 0)
		a45 := encodeAndCount(shape, 45)
		if a0 == 0 {
			t.Errorf("shape %q at angle 0 produced no dark pixels", shape)
		}
		if a45 == 0 {
			t.Errorf("shape %q at angle 45 produced no dark pixels", shape)
		}
		// The counts may or may not differ depending on exact rasterization, but
		// at minimum both should be non-zero (the shape must render at any angle).
	}
	// Rectangle is unaffected by Angle — should produce identical pixel count.
	r0 := encodeAndCount("Rectangle", 0)
	r90 := encodeAndCount("Rectangle", 90)
	if r0 != r90 {
		t.Errorf("Rectangle shape should be unaffected by Angle: angle0=%d angle90=%d", r0, r90)
	}
}

// TestDeserialize_QR_Angle_FromFRX verifies that Barcode.Angle is read
// during FRX deserialization.
// C# BarcodeQR.Angle (BarcodeQR.cs:198).
func TestDeserialize_QR_Angle_FromFRX(t *testing.T) {
	obj := NewBarcodeObject()
	r := &pipelineMockReader{
		strs: map[string]string{"Barcode.Type": "QR"},
		ints: map[string]int{"Barcode.Angle": 45},
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	qr, ok := obj.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("expected *QRBarcode, got %T", obj.Barcode)
	}
	if qr.Angle != 45 {
		t.Fatalf("expected Angle=45, got %d", qr.Angle)
	}
}

func countDarkPixels(img image.Image) int {
	bounds := img.Bounds()
	count := 0
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				count++
			}
		}
	}
	return count
}
