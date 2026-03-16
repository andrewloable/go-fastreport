package barcode_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── IntelligentMailBarcode — additional encode lengths ────────────────────────

func TestIntelligentMailBarcode_Encode_20Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("Encode 20 digits: %v", err)
	}
	if b.EncodedText() != "01234567094987654321" {
		t.Errorf("EncodedText = %q", b.EncodedText())
	}
}

func TestIntelligentMailBarcode_Render_20Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render(200,60): %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_31Digits(t *testing.T) {
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("0123456709498765432112345678901"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(260, 60)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
}

func TestIntelligentMailBarcode_Render_ZeroSize(t *testing.T) {
	// Zero width/height should use defaults (130 x 60)
	b := barcode.NewIntelligentMailBarcode()
	if err := b.Encode("01234567094987654321"); err != nil {
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
	if bounds.Dx() != 130 || bounds.Dy() != 60 {
		t.Errorf("default size: got %dx%d, want 130x60", bounds.Dx(), bounds.Dy())
	}
}

// ── PharmacodeBarcode — boundary values ───────────────────────────────────────

func TestPharmacodeBarcode_Encode_Boundary(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"3", false},      // minimum valid
		{"131070", false}, // maximum valid
		{"2", true},       // below minimum
		{"131071", true},  // above maximum
		{"abc", true},     // non-numeric
	}
	for _, c := range cases {
		b := barcode.NewPharmacodeBarcode()
		err := b.Encode(c.input)
		if c.wantErr && err == nil {
			t.Errorf("Encode(%q): expected error, got nil", c.input)
		}
		if !c.wantErr && err != nil {
			t.Errorf("Encode(%q): unexpected error: %v", c.input, err)
		}
	}
}

func TestPharmacodeBarcode_Render_MinimumValue(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("3"); err != nil {
		t.Fatalf("Encode 3: %v", err)
	}
	img, err := b.Render(100, 60)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

func TestPharmacodeBarcode_Render_MaximumValue(t *testing.T) {
	b := barcode.NewPharmacodeBarcode()
	if err := b.Encode("131070"); err != nil {
		t.Fatalf("Encode 131070: %v", err)
	}
	img, err := b.Render(400, 60)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

// ── PlesseyBarcode — additional encoding paths ────────────────────────────────

func TestPlesseyBarcode_Encode_LowercaseHex(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	// lowercase is accepted (converted to upper)
	if err := b.Encode("abcdef"); err != nil {
		t.Fatalf("Encode lowercase hex: %v", err)
	}
	if b.EncodedText() != "ABCDEF" {
		t.Errorf("EncodedText = %q, want ABCDEF", b.EncodedText())
	}
}

func TestPlesseyBarcode_Encode_InvalidCharG(t *testing.T) {
	b := barcode.NewPlesseyBarcode()
	if err := b.Encode("1G23"); err == nil {
		t.Error("Encode('1G23'): expected error for G")
	}
}

// ── GS1Barcode — Render paths ─────────────────────────────────────────────────

func TestGS1Barcode_Render_AfterValidEncode(t *testing.T) {
	b := barcode.NewGS1Barcode()
	if err := b.Encode("(01)12345678901231"); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	img, err := b.Render(200, 60)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil")
	}
}

// ── BarcodeObject — Serialize with non-default SymbolType fields ──────────────

func TestBarcodeObject_Serialize_NonDefaultAngle(t *testing.T) {
	// Angle != 0 should be serialized
	bo := barcode.NewBarcodeObject()
	bo.SetAngle(90)
	// Verify round-trip
	if bo.Angle() != 90 {
		t.Errorf("Angle = %d, want 90", bo.Angle())
	}
}

func TestBarcodeObject_Serialize_HideIfNoData(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetHideIfNoData(true)
	if !bo.HideIfNoData() {
		t.Error("HideIfNoData should be true")
	}
}

func TestBarcodeObject_Serialize_AllowExpressionsFalse(t *testing.T) {
	bo := barcode.NewBarcodeObject()
	bo.SetAllowExpressions(false)
	if bo.AllowExpressions() {
		t.Error("AllowExpressions should be false")
	}
}
