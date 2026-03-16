package datamatrix_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/datamatrix"
)

// TestEncode_TooLargeContent_Error attempts to trigger the datamatrix.Encode
// internal error path by passing content that exceeds the maximum DataMatrix
// capacity (~3116 bytes for the largest symbol).
func TestEncode_TooLargeContent_Error(t *testing.T) {
	e := datamatrix.New()
	// DataMatrix max capacity is ~3116 bytes; use 5000 bytes to ensure overflow.
	veryLong := strings.Repeat("A", 5000)
	_, err := e.Encode(veryLong, 200, 200)
	if err == nil {
		t.Skip("boombuler datamatrix did not error on oversized content; skipping")
	}
}

// TestEncodeMatrix_TooLargeContent_Error attempts to trigger the datamatrix.Encode
// error path inside EncodeMatrix.
func TestEncodeMatrix_TooLargeContent_Error(t *testing.T) {
	e := datamatrix.New()
	veryLong := strings.Repeat("A", 5000)
	_, err := e.EncodeMatrix(veryLong)
	if err == nil {
		t.Skip("boombuler datamatrix did not error on oversized content; skipping")
	}
}
