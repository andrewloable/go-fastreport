package serial

// writer_nil_test.go — covers the final unreachable-in-practice branch in
// typeNameOf (writer.go:212) by passing a nil report.Serializable interface
// value, which produces the string "<nil>" (no dot) and therefore exercises
// the `return name` fallback.
//
// This is the only way to reach line 212 in Go: a nil interface value has
// no concrete type, so fmt.Sprintf("%T", obj) returns "<nil>" — a string
// without any '.' character.  The loop finds no separator and falls through
// to `return name`.
//
// Note: BeginObject("<nil>") would succeed, but the subsequent obj.Serialize()
// call on a nil interface would panic, so we test typeNameOf directly via the
// unexported function rather than through WriteObject.

import (
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/report"
)

// TestTypeNameOf_NilInterface exercises the `return name` branch at line 212
// of typeNameOf by passing a nil report.Serializable interface value.
//
// fmt.Sprintf("%T", nilSerializable) returns "<nil>" which contains no '.'
// character, so the loop in typeNameOf never finds a separator and falls
// through to `return name`.
func TestTypeNameOf_NilInterface(t *testing.T) {
	var obj report.Serializable // nil interface
	got := typeNameOf(obj)
	// The result should be the raw %T string for a nil interface: "<nil>".
	if strings.Contains(got, ".") {
		t.Errorf("typeNameOf(nil): got %q, expected no dot (raw %%T fallback)", got)
	}
	if got == "" {
		t.Error("typeNameOf(nil): got empty string, want non-empty")
	}
}
