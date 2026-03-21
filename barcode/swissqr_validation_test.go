package barcode_test

// swissqr_validation_test.go tests the Swiss QR validation helpers in
// swissqr_validation.go, ported from C# SwissQRCode.cs.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode"
)

// ── ValidateIBAN ──────────────────────────────────────────────────────────────

func TestValidateIBAN_ValidCH(t *testing.T) {
	// CH5604835012345678009 is a well-known test IBAN from the SPC spec.
	if err := barcode.ValidateIBAN("CH5604835012345678009"); err != nil {
		t.Errorf("expected valid IBAN, got error: %v", err)
	}
}

func TestValidateIBAN_ValidWithSpaces(t *testing.T) {
	// Spaces are stripped before validation (common user input format).
	if err := barcode.ValidateIBAN("CH56 0483 5012 3456 7800 9"); err != nil {
		t.Errorf("expected valid IBAN with spaces, got error: %v", err)
	}
}

func TestValidateIBAN_InvalidChecksum(t *testing.T) {
	// Change one digit to make the checksum wrong.
	err := barcode.ValidateIBAN("CH5604835012345678001")
	if err == nil {
		t.Error("expected error for IBAN with wrong checksum, got nil")
	}
}

func TestValidateIBAN_WrongPrefix(t *testing.T) {
	// DE IBAN passes MOD-97 but not the CH/LI requirement.
	// Using a structurally valid German IBAN.
	err := barcode.ValidateIBAN("DE89370400440532013000")
	if err == nil {
		t.Error("expected error for non-CH/LI IBAN, got nil")
	}
}

func TestValidateIBAN_TooShort(t *testing.T) {
	err := barcode.ValidateIBAN("CH56")
	if err == nil {
		t.Error("expected error for too-short IBAN, got nil")
	}
}

func TestValidateIBAN_Empty(t *testing.T) {
	err := barcode.ValidateIBAN("")
	if err == nil {
		t.Error("expected error for empty IBAN, got nil")
	}
}

// ── IsQRIBAN ─────────────────────────────────────────────────────────────────

func TestIsQRIBAN_True(t *testing.T) {
	// A QR-IBAN has IID in range 30000-31999.
	// CH44 3199 9123 0008 8901 2 — IID = 31999.
	// Build a valid QR-IBAN: CH + check + 30000 + rest. We use a known one.
	// CH5630000012345678901 — IID 30000.
	// Verify it looks like a QR-IBAN (we test IsQRIBAN independently of ValidateIBAN).
	if !barcode.IsQRIBAN("CH5630000012345678901") {
		t.Error("expected IsQRIBAN to return true for IID 30000")
	}
}

func TestIsQRIBAN_False_NormalIBAN(t *testing.T) {
	// CH5604835012345678009 has IID 04835 which is not in 30000-31999.
	if barcode.IsQRIBAN("CH5604835012345678009") {
		t.Error("expected IsQRIBAN to return false for normal IBAN")
	}
}

func TestIsQRIBAN_False_TooShort(t *testing.T) {
	if barcode.IsQRIBAN("CH56") {
		t.Error("expected IsQRIBAN to return false for short string")
	}
}

func TestIsQRIBAN_IIDAtUpperBound(t *testing.T) {
	// IID = 31999 should still be a QR-IBAN.
	if !barcode.IsQRIBAN("CH5631999012345678901") {
		t.Error("expected IsQRIBAN to return true for IID 31999")
	}
}

func TestIsQRIBAN_IIDJustAboveRange(t *testing.T) {
	// IID = 32000 is outside range.
	if barcode.IsQRIBAN("CH5632000012345678901") {
		t.Error("expected IsQRIBAN to return false for IID 32000")
	}
}

// ── ChecksumMod10 ─────────────────────────────────────────────────────────────

func TestChecksumMod10_Valid(t *testing.T) {
	// "210000000003139471430009017" is a known valid QR reference.
	if !barcode.ChecksumMod10("210000000003139471430009017") {
		t.Error("expected ChecksumMod10 to return true for valid reference")
	}
}

func TestChecksumMod10_InvalidCheckDigit(t *testing.T) {
	// Change last digit from 7 to 8.
	if barcode.ChecksumMod10("210000000003139471430009018") {
		t.Error("expected ChecksumMod10 to return false for wrong check digit")
	}
}

func TestChecksumMod10_TooShort(t *testing.T) {
	if barcode.ChecksumMod10("1") {
		t.Error("expected ChecksumMod10 to return false for single-digit string")
	}
}

func TestChecksumMod10_Empty(t *testing.T) {
	if barcode.ChecksumMod10("") {
		t.Error("expected ChecksumMod10 to return false for empty string")
	}
}

func TestChecksumMod10_ContainsLetter(t *testing.T) {
	if barcode.ChecksumMod10("12345A") {
		t.Error("expected ChecksumMod10 to return false when string contains a letter")
	}
}

func TestChecksumMod10_WithSpaces(t *testing.T) {
	// Spaces should be stripped before validation.
	if !barcode.ChecksumMod10("21 0000 0000 0313 9471 4300 0901 7") {
		t.Error("expected ChecksumMod10 to return true when spaces are stripped")
	}
}

// ── ValidateReference ─────────────────────────────────────────────────────────

func TestValidateReference_NON_Empty(t *testing.T) {
	if err := barcode.ValidateReference("NON", ""); err != nil {
		t.Errorf("expected nil for NON with empty reference, got: %v", err)
	}
}

func TestValidateReference_NON_NonEmpty(t *testing.T) {
	err := barcode.ValidateReference("NON", "12345")
	if err == nil {
		t.Error("expected error for NON with non-empty reference")
	}
}

func TestValidateReference_QRR_Valid(t *testing.T) {
	// 27-digit numeric reference with valid MOD-10.
	if err := barcode.ValidateReference("QRR", "210000000003139471430009017"); err != nil {
		t.Errorf("expected nil for valid QRR reference, got: %v", err)
	}
}

func TestValidateReference_QRR_TooLong(t *testing.T) {
	// 28 digits (after stripping) is too long.
	ref := "1234567890123456789012345678" // 28 digits
	err := barcode.ValidateReference("QRR", ref)
	if err == nil {
		t.Error("expected error for QRR reference longer than 27 digits")
	}
}

func TestValidateReference_QRR_NonNumeric(t *testing.T) {
	err := barcode.ValidateReference("QRR", "1234567890ABCDE")
	if err == nil {
		t.Error("expected error for QRR reference containing letters")
	}
}

func TestValidateReference_QRR_BadChecksum(t *testing.T) {
	// Last digit changed to make checksum fail.
	err := barcode.ValidateReference("QRR", "210000000003139471430009018")
	if err == nil {
		t.Error("expected error for QRR reference with invalid MOD-10 checksum")
	}
}

func TestValidateReference_QRR_Empty(t *testing.T) {
	err := barcode.ValidateReference("QRR", "")
	if err == nil {
		t.Error("expected error for QRR with empty reference")
	}
}

func TestValidateReference_SCOR_Valid(t *testing.T) {
	// ISO 11649 reference — up to 25 alphanumeric chars.
	if err := barcode.ValidateReference("SCOR", "RF18539007547034"); err != nil {
		t.Errorf("expected nil for valid SCOR reference, got: %v", err)
	}
}

func TestValidateReference_SCOR_TooLong(t *testing.T) {
	// 26 alphanumeric chars is too long.
	ref := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" // 26 chars
	err := barcode.ValidateReference("SCOR", ref)
	if err == nil {
		t.Error("expected error for SCOR reference longer than 25 chars")
	}
}

func TestValidateReference_SCOR_Empty(t *testing.T) {
	err := barcode.ValidateReference("SCOR", "")
	if err == nil {
		t.Error("expected error for SCOR with empty reference")
	}
}

func TestValidateReference_UnknownType(t *testing.T) {
	err := barcode.ValidateReference("XYZZY", "12345")
	if err == nil {
		t.Error("expected error for unknown reference type")
	}
	if !strings.Contains(err.Error(), "unknown reference type") {
		t.Errorf("expected 'unknown reference type' in error, got: %v", err)
	}
}

// ── ValidateContact ───────────────────────────────────────────────────────────

func TestValidateContact_Structured_Valid(t *testing.T) {
	err := barcode.ValidateContact(
		"Max Muster",
		"Musterstrasse",
		"1",
		"8000",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err != nil {
		t.Errorf("expected valid structured contact, got: %v", err)
	}
}

func TestValidateContact_Structured_EmptyName(t *testing.T) {
	err := barcode.ValidateContact(
		"",
		"Musterstrasse",
		"1",
		"8000",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateContact_Structured_NameTooLong(t *testing.T) {
	longName := strings.Repeat("A", 71)
	err := barcode.ValidateContact(
		longName,
		"Musterstrasse",
		"1",
		"8000",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for name longer than 70 chars")
	}
}

func TestValidateContact_Structured_InvalidCharset(t *testing.T) {
	// Tab character is not in the allowed charset.
	err := barcode.ValidateContact(
		"Max\tMuster",
		"Musterstrasse",
		"1",
		"8000",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for name with tab character")
	}
}

func TestValidateContact_Structured_EmptyZip(t *testing.T) {
	err := barcode.ValidateContact(
		"Max Muster",
		"Musterstrasse",
		"1",
		"",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for empty zip in structured address")
	}
}

func TestValidateContact_Structured_EmptyCity(t *testing.T) {
	err := barcode.ValidateContact(
		"Max Muster",
		"Musterstrasse",
		"1",
		"8000",
		"",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for empty city in structured address")
	}
}

func TestValidateContact_Structured_InvalidCountry(t *testing.T) {
	err := barcode.ValidateContact(
		"Max Muster",
		"Musterstrasse",
		"1",
		"8000",
		"Zürich",
		"XX",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for invalid country code")
	}
}

func TestValidateContact_Structured_HouseNumberTooLong(t *testing.T) {
	err := barcode.ValidateContact(
		"Max Muster",
		"Musterstrasse",
		strings.Repeat("1", 17),
		"8000",
		"Zürich",
		"CH",
		barcode.SwissQRStructuredAddress,
	)
	if err == nil {
		t.Error("expected error for house number longer than 16 chars")
	}
}

func TestValidateContact_Combined_Valid(t *testing.T) {
	err := barcode.ValidateContact(
		"Erika Muster",
		"Musterstrasse 1",
		"8000 Zürich",
		"",
		"",
		"CH",
		barcode.SwissQRCombinedAddress,
	)
	if err != nil {
		t.Errorf("expected valid combined contact, got: %v", err)
	}
}

func TestValidateContact_Combined_EmptyLine2(t *testing.T) {
	err := barcode.ValidateContact(
		"Erika Muster",
		"Musterstrasse 1",
		"",
		"",
		"",
		"CH",
		barcode.SwissQRCombinedAddress,
	)
	if err == nil {
		t.Error("expected error for empty address line 2 in combined address")
	}
}

func TestValidateContact_Combined_Line1TooLong(t *testing.T) {
	err := barcode.ValidateContact(
		"Erika Muster",
		strings.Repeat("A", 71),
		"8000 Zürich",
		"",
		"",
		"CH",
		barcode.SwissQRCombinedAddress,
	)
	if err == nil {
		t.Error("expected error for address line 1 longer than 70 chars")
	}
}

func TestValidateContact_Combined_Line2TooLong(t *testing.T) {
	err := barcode.ValidateContact(
		"Erika Muster",
		"Musterstrasse 1",
		strings.Repeat("A", 71),
		"",
		"",
		"CH",
		barcode.SwissQRCombinedAddress,
	)
	if err == nil {
		t.Error("expected error for address line 2 longer than 70 chars")
	}
}

// ── ValidateAdditionalInformation ─────────────────────────────────────────────

func TestValidateAdditionalInformation_Valid(t *testing.T) {
	if err := barcode.ValidateAdditionalInformation("Rechnung 123", ""); err != nil {
		t.Errorf("expected nil for valid additional info, got: %v", err)
	}
}

func TestValidateAdditionalInformation_BothFields(t *testing.T) {
	if err := barcode.ValidateAdditionalInformation("Invoice 001", "//S1/10/1234"); err != nil {
		t.Errorf("expected nil when both fields are valid, got: %v", err)
	}
}

func TestValidateAdditionalInformation_TooLong(t *testing.T) {
	msg := strings.Repeat("A", 100)
	bill := strings.Repeat("B", 41) // 100+41 = 141 > 140
	err := barcode.ValidateAdditionalInformation(msg, bill)
	if err == nil {
		t.Error("expected error when combined length exceeds 140")
	}
}

func TestValidateAdditionalInformation_ExactLimit(t *testing.T) {
	msg := strings.Repeat("A", 70)
	bill := strings.Repeat("B", 70) // 140 total — exactly at limit
	if err := barcode.ValidateAdditionalInformation(msg, bill); err != nil {
		t.Errorf("expected nil at exactly 140 chars, got: %v", err)
	}
}

func TestValidateAdditionalInformation_InvalidCharset(t *testing.T) {
	err := barcode.ValidateAdditionalInformation("Message\x01Invalid", "")
	if err == nil {
		t.Error("expected error for unstructured message with invalid characters")
	}
}

func TestValidateAdditionalInformation_BillInfoInvalidCharset(t *testing.T) {
	err := barcode.ValidateAdditionalInformation("", "Bill\x02Bad")
	if err == nil {
		t.Error("expected error for bill information with invalid characters")
	}
}

func TestValidateAdditionalInformation_BothEmpty(t *testing.T) {
	if err := barcode.ValidateAdditionalInformation("", ""); err != nil {
		t.Errorf("expected nil for both empty fields, got: %v", err)
	}
}

// ── ValidateSwissQRParameters ─────────────────────────────────────────────────

func TestValidateSwissQRParameters_Valid(t *testing.T) {
	p := barcode.SwissQRParameters{
		IBAN:                "CH5604835012345678009",
		Currency:            "CHF",
		CreditorName:        "Max Muster",
		CreditorStreet:      "Musterstrasse",
		CreditorPostalCode:  "8000",
		CreditorCity:        "Zürich",
		CreditorCountry:     "CH",
		ReferenceType:       "NON",
		Reference:           "",
		UnstructuredMessage: "Rechnung 123",
	}
	if err := barcode.ValidateSwissQRParameters(p); err != nil {
		t.Errorf("expected valid parameters, got: %v", err)
	}
}

func TestValidateSwissQRParameters_BadIBAN(t *testing.T) {
	p := barcode.SwissQRParameters{
		IBAN:              "CH9999999999999999999",
		Currency:          "CHF",
		CreditorName:      "Max Muster",
		CreditorPostalCode: "8000",
		CreditorCity:      "Zürich",
		CreditorCountry:   "CH",
		ReferenceType:     "NON",
	}
	if err := barcode.ValidateSwissQRParameters(p); err == nil {
		t.Error("expected error for invalid IBAN")
	}
}

func TestValidateSwissQRParameters_BadCurrency(t *testing.T) {
	p := barcode.SwissQRParameters{
		IBAN:              "CH5604835012345678009",
		Currency:          "USD",
		CreditorName:      "Max Muster",
		CreditorPostalCode: "8000",
		CreditorCity:      "Zürich",
		CreditorCountry:   "CH",
		ReferenceType:     "NON",
	}
	if err := barcode.ValidateSwissQRParameters(p); err == nil {
		t.Error("expected error for invalid currency")
	} else if !strings.Contains(err.Error(), "currency") {
		t.Errorf("expected 'currency' in error message, got: %v", err)
	}
}

func TestValidateSwissQRParameters_BadReference(t *testing.T) {
	p := barcode.SwissQRParameters{
		IBAN:              "CH5604835012345678009",
		Currency:          "CHF",
		CreditorName:      "Max Muster",
		CreditorPostalCode: "8000",
		CreditorCity:      "Zürich",
		CreditorCountry:   "CH",
		ReferenceType:     "QRR",
		Reference:         "notanumber",
	}
	if err := barcode.ValidateSwissQRParameters(p); err == nil {
		t.Error("expected error for QRR reference with non-numeric chars")
	}
}

func TestValidateSwissQRParameters_EURCurrency(t *testing.T) {
	p := barcode.SwissQRParameters{
		IBAN:              "CH5604835012345678009",
		Currency:          "EUR",
		CreditorName:      "Max Muster",
		CreditorPostalCode: "8000",
		CreditorCity:      "Zürich",
		CreditorCountry:   "CH",
		ReferenceType:     "NON",
	}
	if err := barcode.ValidateSwissQRParameters(p); err != nil {
		t.Errorf("expected nil for EUR currency, got: %v", err)
	}
}
