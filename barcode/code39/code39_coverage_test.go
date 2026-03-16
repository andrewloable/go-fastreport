package code39_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/barcode/code39"
)

// TestEncode_ExtendedMode_NonASCIIError covers the error path from the
// underlying code39.Encode library call (line 46–48).
// With AllowExtended=true our wrapper skips Validate, passing text directly
// (after ToUpper) to the library. The boombuler code39 encoder rejects chars
// outside ASCII (\x80–\xff) with "Only ASCII strings can be encoded".
func TestEncode_ExtendedMode_NonASCIIError(t *testing.T) {
	enc := code39.NewEncoder()
	enc.AllowExtended = true
	// \x80 is above ASCII and is rejected by the library's extended encoder.
	_, err := enc.Encode("\x80", 300, 100)
	if err == nil {
		t.Error("expected error for non-ASCII char in Code 39 extended mode")
	}
}

// TestEncode_ExtendedMode_ScaleTooSmall covers the barcode.Scale error path
// (lines 51–53) when the requested width is smaller than the barcode's native
// pixel width. In extended mode, "HELLO" encodes to a 90-pixel wide barcode;
// requesting width=50 passes our positive-width guard but fails in Scale.
func TestEncode_ExtendedMode_ScaleTooSmall(t *testing.T) {
	enc := code39.NewEncoder()
	enc.AllowExtended = true
	// "hello@world" in extended mode produces a barcode wider than 50px.
	_, err := enc.Encode("hello@world", 50, 100)
	if err == nil {
		t.Error("expected error when scale width < barcode native width")
	}
}

// TestEncode_StandardMode_ScaleTooSmall covers the barcode.Scale error path
// (lines 51–53) in standard (non-extended) mode. "HELLO" encodes to a
// 90-pixel wide barcode; requesting width=50 fails in Scale.
func TestEncode_StandardMode_ScaleTooSmall(t *testing.T) {
	enc := code39.NewEncoder()
	_, err := enc.Encode("HELLO", 50, 100)
	if err == nil {
		t.Error("expected error when scale width < native barcode width (standard mode)")
	}
}
