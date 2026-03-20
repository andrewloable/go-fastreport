package datamatrix_test

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/datamatrix"
)

// TestEncode_TooLargeContent verifies that very large content does not panic.
// The native encoder may return a fallback matrix or error gracefully.
func TestEncode_TooLargeContent(t *testing.T) {
	e := datamatrix.New()
	veryLong := strings.Repeat("A", 5000)
	img, err := e.Encode(veryLong, 200, 200)
	if err != nil {
		// Error is acceptable for oversized content.
		return
	}
	if img == nil {
		t.Error("expected non-nil image or error")
	}
}

// TestEncodeMatrix_TooLargeContent verifies that very large content does not panic.
func TestEncodeMatrix_TooLargeContent(t *testing.T) {
	e := datamatrix.New()
	veryLong := strings.Repeat("A", 5000)
	matrix, err := e.EncodeMatrix(veryLong)
	// Either an error or a nil/empty matrix is acceptable for oversized content.
	_ = err
	_ = matrix
}
