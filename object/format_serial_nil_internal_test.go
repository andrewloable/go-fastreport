package object

// format_serial_nil_internal_test.go — internal test to cover the
// `if f == nil { return }` dead-code branch in serializeTextFormat
// (format_serial.go:11-13).
//
// serializeTextFormat is called from TextObjectBase.Serialize only when
// t.format != nil, so the nil guard is dead code via the public API.
// We call the function directly from the internal test package.

import (
	"testing"
)

// noopWriter is defined in advmatrix_row_cell_internal_test.go (same package).

// TestSerializeTextFormat_Nil exercises the `if f == nil { return }` branch
// at format_serial.go:11 by calling serializeTextFormat directly with nil.
func TestSerializeTextFormat_Nil(t *testing.T) {
	w := &noopWriter{}
	// Calling with nil should not panic and should return immediately.
	serializeTextFormat(w, nil)
	// No assertion needed: the test verifies no panic occurs.
}
