package code128_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code128"
)

// TestEncode_NonASCIIChar_LibraryError covers the error path from the underlying
// code128.Encode library call. The character \x80 (non-ASCII) passes our
// Validate check is not performed inside Encode, so it reaches the library.
// Actually code128.Validate checks >0x7E so \x80 would be caught there.
// Instead we rely on the fact that the library itself rejects non-ASCII
// (chars > 0x7E). We call Encode directly (not Validate) to hit the library error path.
//
// The underlying boombuler code128.Encode rejects chars > 0x7E with
// `"<char>" could not be encoded`. Our wrapper's Encode calls the library
// without calling our Validate first, so the library error is returned.
func TestEncode_NonASCIIChar_LibraryError(t *testing.T) {
	enc := code128.NewEncoder()
	// \x80 is above the Code 128 character range — library will reject it.
	_, err := enc.Encode("\x80test", 300, 100)
	if err == nil {
		t.Error("expected error for non-ASCII character \\x80 in Code 128 encode")
	}
}

// TestEncode_ScaleWidthTooSmall covers the barcode.Scale error path when the
// requested width is smaller than the barcode's native pixel width.
// A Code 128 barcode for "hello" is 90px wide; requesting width=50 (< 90)
// passes our positive-width guard but fails in barcode.Scale.
func TestEncode_ScaleWidthTooSmall(t *testing.T) {
	enc := code128.NewEncoder()
	// "hello" encodes to a 90-pixel wide barcode; request only 50.
	_, err := enc.Encode("hello", 50, 100)
	if err == nil {
		t.Error("expected error when scale width < barcode native width")
	}
}
