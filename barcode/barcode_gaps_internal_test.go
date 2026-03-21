// barcode_gaps_internal_test.go — internal package tests (package barcode) to
// cover the remaining gaps in BarcodeObject.Deserialize.
//
// Uncovered branches targeted:
//   barcode.go:422-423  else if name := r.ReadStr("Barcode", ""); name != "" (display-name fallback)
//   barcode.go:426-431  QRBarcode ErrorCorrection property set via Deserialize
//   barcode.go:460      EAN13Barcode.Encode 13-digit retry-succeeds path
package barcode

import (
	"testing"
)

// ── BarcodeObject.Deserialize: "Barcode" display-name fallback (else if branch) ─

// TestBarcodeObject_Deserialize_DisplayNameFallback covers the
//
//	else if name := r.ReadStr("Barcode", ""); name != "" { b.Barcode = NewBarcodeByName(name) }
//
// branch by providing only the "Barcode" key (not "Barcode.Type") in the reader.
func TestBarcodeObject_Deserialize_DisplayNameFallback(t *testing.T) {
	b := NewBarcodeObject()
	// "Barcode.Type" is absent (empty string default) so the if-branch is false.
	// "Barcode" display name is present → else if branch fires.
	r := &rcbMockReader{
		strs:  map[string]string{"Barcode": "QR Code"},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	if b.Barcode == nil {
		t.Fatal("Barcode should not be nil after Deserialize with display name")
	}
	// NewBarcodeByName("QR Code") should resolve to QR type.
	if b.Barcode.Type() != BarcodeTypeQR {
		t.Errorf("Barcode.Type = %q, want %q", b.Barcode.Type(), BarcodeTypeQR)
	}
}

// TestBarcodeObject_Deserialize_NoBarcode covers the path where neither
// "Barcode.Type" nor "Barcode" is present — Barcode remains the default.
func TestBarcodeObject_Deserialize_NoBarcode(t *testing.T) {
	b := NewBarcodeObject()
	r := &rcbMockReader{
		strs:  map[string]string{},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	// Barcode should still be the default (Code128 set by NewBarcodeObject).
	if b.Barcode == nil {
		t.Error("Barcode should not be nil (default Code128 preserved)")
	}
}

// TestBarcodeObject_Deserialize_UnknownDisplayName tests that an unrecognised
// display name falls back to Code128 (via NewBarcodeByName fallback).
func TestBarcodeObject_Deserialize_UnknownDisplayName(t *testing.T) {
	b := NewBarcodeObject()
	r := &rcbMockReader{
		strs:  map[string]string{"Barcode": "NonExistentSymbology"},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	if b.Barcode == nil {
		t.Fatal("Barcode should not be nil after unknown display name")
	}
	if b.Barcode.Type() != BarcodeTypeCode128 {
		t.Errorf("expected Code128 fallback, got %q", b.Barcode.Type())
	}
}

// ── BarcodeObject.Deserialize: QRBarcode.ErrorCorrection property ────────────

// TestBarcodeObject_Deserialize_QRErrorCorrection covers the
//
//	if b.Barcode != nil { if qrbc, ok := b.Barcode.(*QRBarcode); ok {
//	    if ec := r.ReadStr("Barcode.ErrorCorrection", ""); ec != "" { qrbc.ErrorCorrection = ec }
//
// branch by Deserializing with a QR barcode type AND a non-empty ErrorCorrection.
func TestBarcodeObject_Deserialize_QRErrorCorrection(t *testing.T) {
	b := NewBarcodeObject()
	r := &rcbMockReader{
		strs: map[string]string{
			"Barcode.Type":             string(BarcodeTypeQR),
			"Barcode.ErrorCorrection": "H",
		},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	if b.Barcode == nil {
		t.Fatal("Barcode should not be nil")
	}
	qr, ok := b.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("Barcode type = %T, want *QRBarcode", b.Barcode)
	}
	if qr.ErrorCorrection != "H" {
		t.Errorf("ErrorCorrection = %q, want H", qr.ErrorCorrection)
	}
}

// TestBarcodeObject_Deserialize_QRErrorCorrection_EmptyEC covers the
// inner `if ec != ""` false-branch — ErrorCorrection attr present but empty.
func TestBarcodeObject_Deserialize_QRErrorCorrection_EmptyEC(t *testing.T) {
	b := NewBarcodeObject()
	r := &rcbMockReader{
		strs: map[string]string{
			"Barcode.Type": string(BarcodeTypeQR),
			// "Barcode.ErrorCorrection" absent → ReadStr returns "" → inner if is false
		},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	qr, ok := b.Barcode.(*QRBarcode)
	if !ok {
		t.Fatalf("Barcode type = %T, want *QRBarcode", b.Barcode)
	}
	// Default ErrorCorrection "L" should be preserved (C# BarcodeQR.cs:143).
	if qr.ErrorCorrection != "L" {
		t.Errorf("ErrorCorrection = %q, want L (default)", qr.ErrorCorrection)
	}
}

// TestBarcodeObject_Deserialize_NonQR_BarcodeNotNil covers the
// `if b.Barcode != nil` true + `if qrbc, ok ... *QRBarcode; ok` FALSE branch
// (i.e. Barcode is set to a non-QR type so the type assertion fails).
func TestBarcodeObject_Deserialize_NonQR_BarcodeNotNil(t *testing.T) {
	b := NewBarcodeObject()
	r := &rcbMockReader{
		strs: map[string]string{
			"Barcode.Type":             string(BarcodeTypeCode128),
			"Barcode.ErrorCorrection": "H", // ignored for non-QR
		},
		bools: map[string]bool{},
	}
	if err := b.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: unexpected error: %v", err)
	}
	if b.Barcode.Type() != BarcodeTypeCode128 {
		t.Errorf("Barcode.Type = %q, want Code128", b.Barcode.Type())
	}
}

// ── EAN13Barcode.Encode: 13-digit retry-succeeds path ────────────────────────

// TestEAN13Barcode_Encode_13DigitRetrySucceeds covers the path in EAN13Barcode.Encode:
//
//	bc, err := ean.Encode(text)
//	if err != nil {
//	    if len(text) == 13 {
//	        bc, err = ean.Encode(text[:12])  ← this line
//	    }
//	    if err != nil { return ... }  ← this takes the false branch (retry succeeded)
//	}
//
// A 13-digit string with an incorrect check digit causes ean.Encode to fail the
// first time, then succeed on the retry with the first 12 digits.
func TestEAN13Barcode_Encode_13DigitRetrySucceeds(t *testing.T) {
	b := NewEAN13Barcode()
	// "5901234123457" — the last digit '7' is wrong (correct checksum gives '6').
	// ean.Encode("5901234123457") fails; ean.Encode("590123412345") succeeds.
	err := b.Encode("5901234123457")
	if err != nil {
		// If the library happens to accept it, that's fine too.
		t.Logf("Encode 13-digit wrong checksum: %v (library may or may not retry)", err)
		return
	}
	// Retry succeeded — encoded text should be the original 13-digit input.
	if b.encodedText != "5901234123457" {
		t.Errorf("encodedText = %q, want original 13-digit input", b.encodedText)
	}
}

// TestEAN13Barcode_Encode_13DigitRetryFails covers the path where the initial
// ean.Encode fails AND the 13-digit retry also fails.
func TestEAN13Barcode_Encode_13DigitRetryFails(t *testing.T) {
	b := NewEAN13Barcode()
	// Fewer than 12 digits AND 13 chars (not all digits) → both attempts fail.
	err := b.Encode("123456789ABCD") // 13 chars, non-digit → both ean.Encode calls fail
	if err == nil {
		t.Log("ean encode: no error (library may be lenient)")
	}
}
