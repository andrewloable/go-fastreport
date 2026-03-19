package barcode_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// TestSwissQR_FormatPayload verifies the payload string built from SwissQRParameters
// follows the Swiss Payment Standards SPC v2.0 format.
func TestSwissQR_FormatPayload(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	b.Params = barcode.SwissQRParameters{
		IBAN:                  "CH5604835012345678009",
		Currency:              "CHF",
		CreditorName:          "Max Muster",
		CreditorStreet:        "Musterstrasse 1",
		CreditorCity:          "Zürich",
		CreditorPostalCode:    "8000",
		CreditorCountry:       "CH",
		Amount:                "100.00",
		Reference:             "",
		ReferenceType:         "NON",
		UnstructuredMessage:   "Rechnung 123",
		TrailerEPD:            "EPD",
		AlternativeProcedure1: "",
		AlternativeProcedure2: "",
	}

	payload := b.FormatPayload()

	// Payload must start with the SPC header lines.
	if !strings.HasPrefix(payload, "SPC\n0200\n1\n") {
		t.Errorf("payload does not start with SPC header: %q", payload[:min(len(payload), 40)])
	}

	// IBAN must be present.
	if !strings.Contains(payload, "CH5604835012345678009") {
		t.Error("payload does not contain IBAN")
	}

	// Address type K must be present.
	if !strings.Contains(payload, "\nK\n") {
		t.Error("payload does not contain address type 'K'")
	}

	// Creditor name must be present.
	if !strings.Contains(payload, "Max Muster") {
		t.Error("payload does not contain creditor name")
	}

	// Currency must be present.
	if !strings.Contains(payload, "\nCHF\n") && !strings.HasSuffix(payload, "\nCHF") {
		t.Error("payload does not contain currency CHF")
	}

	// Reference type NON must be present.
	if !strings.Contains(payload, "\nNON\n") {
		t.Error("payload does not contain reference type NON")
	}

	// Trailer EPD must be present.
	if !strings.Contains(payload, "EPD") {
		t.Error("payload does not contain EPD trailer")
	}

	// Unstructured message must be present.
	if !strings.Contains(payload, "Rechnung 123") {
		t.Error("payload does not contain unstructured message")
	}
}

// TestSwissQR_Encode encodes a minimal valid SwissQR payload and checks the
// resulting image is non-nil and has positive dimensions.
func TestSwissQR_Encode(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	b.Params = barcode.SwissQRParameters{
		IBAN:            "CH5604835012345678009",
		Currency:        "CHF",
		CreditorName:    "Test AG",
		CreditorStreet:  "Teststrasse 1",
		CreditorCity:    "Bern",
		CreditorCountry: "CH",
		Amount:          "50.00",
		ReferenceType:   "NON",
		TrailerEPD:      "EPD",
	}

	if err := b.Encode(""); err != nil {
		t.Fatalf("Encode returned unexpected error: %v", err)
	}

	img, err := b.Render(300, 300)
	if err != nil {
		t.Fatalf("Render returned unexpected error: %v", err)
	}
	if img == nil {
		t.Fatal("Render returned nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("image has non-positive dimensions: %v", bounds)
	}
}

// TestSwissQR_DefaultValue verifies DefaultValue returns a non-empty string
// that can be encoded as a QR code.
func TestSwissQR_DefaultValue(t *testing.T) {
	b := barcode.NewSwissQRBarcode()
	dv := b.DefaultValue()
	if dv == "" {
		t.Fatal("DefaultValue returned empty string")
	}
	// Verify the default value is encodable.
	if err := b.Encode(dv); err != nil {
		t.Fatalf("Encode(DefaultValue()) returned unexpected error: %v", err)
	}
}

// TestSwissQR_NewBarcode_Factory verifies that NewBarcodeByType returns a
// non-nil BarcodeBase for BarcodeTypeSwissQR and that it has the correct type.
func TestSwissQR_NewBarcode_Factory(t *testing.T) {
	b := barcode.NewBarcodeByType(barcode.BarcodeTypeSwissQR)
	if b == nil {
		t.Fatal("NewBarcodeByType(BarcodeTypeSwissQR) returned nil")
	}
	if b.Type() != barcode.BarcodeTypeSwissQR {
		t.Errorf("Type() = %q, want %q", b.Type(), barcode.BarcodeTypeSwissQR)
	}
}

// min is a helper because Go 1.20 added min as a builtin; for older compat define it here.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
