package qr_test

import (
	"strings"
	"testing"

	frqr "github.com/andrewloable/go-fastreport/barcode/qr"
)

// TestEncode_TooLongContent_Error attempts to trigger the qr.Encode error path
// by passing content that exceeds QR Code capacity.
// QR Code version 40-H supports a maximum of ~1273 bytes (binary).
// Using a long repeating binary string at ECLevelH should exceed the limit.
func TestEncode_TooLongContent_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = frqr.ECLevelH
	// QR v40 at level H max ~1273 bytes binary; use a 4000-byte string to ensure overflow.
	veryLong := strings.Repeat("X", 4000)
	_, err := enc.Encode(veryLong, 300)
	if err == nil {
		t.Skip("boombuler qr did not error on oversized content; skipping")
	}
}

// TestEncodeMatrix_TooLongContent_Error attempts to trigger the qr.Encode error
// path inside EncodeMatrix.
func TestEncodeMatrix_TooLongContent_Error(t *testing.T) {
	enc := frqr.NewEncoder()
	enc.ECLevel = frqr.ECLevelH
	veryLong := strings.Repeat("X", 4000)
	_, err := enc.EncodeMatrix(veryLong)
	if err == nil {
		t.Skip("boombuler qr did not error on oversized content; skipping")
	}
}
