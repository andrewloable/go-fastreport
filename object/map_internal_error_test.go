package object

// map_internal_error_test.go — internal tests to cover the FinishChild error
// break branch inside MapObject.DeserializeChild (map.go:166).
//
// The branch `if r.FinishChild() != nil { break }` inside the grandchild-
// draining loop is unreachable from the public serial.Reader because
// FinishChild only fails when the XML stream is malformed in a way that the
// normal test helpers don't produce. This internal test injects a mock
// report.Reader whose FinishChild returns a sentinel error after the first
// NextChild call succeeds, so the break path is executed.

import (
	"errors"
	"testing"
)

// mapGrandchildReader is a mock report.Reader that simulates a MapLayer element
// with exactly one grandchild. NextChild returns true once (grandchild found),
// and FinishChild returns a sentinel error to trigger the break in the drain loop.
type mapGrandchildReader struct {
	nextCalled    int
	finishErr     error
	// Attribute reads all return zero/empty/default values.
}

func (r *mapGrandchildReader) ReadStr(name, def string) string       { return def }
func (r *mapGrandchildReader) ReadInt(name string, def int) int      { return def }
func (r *mapGrandchildReader) ReadBool(name string, def bool) bool   { return def }
func (r *mapGrandchildReader) ReadFloat(name string, def float32) float32 { return def }

func (r *mapGrandchildReader) NextChild() (string, bool) {
	r.nextCalled++
	if r.nextCalled == 1 {
		// Report one grandchild named "GrandChild".
		return "GrandChild", true
	}
	return "", false
}

func (r *mapGrandchildReader) FinishChild() error {
	return r.finishErr
}

// TestMapObject_DeserializeChild_FinishChildError covers the break branch at
// map.go:166 inside the grandchild-drain loop of DeserializeChild.
func TestMapObject_DeserializeChild_FinishChildError(t *testing.T) {
	m := NewMapObject()

	rd := &mapGrandchildReader{finishErr: errors.New("finish child error")}

	// Call DeserializeChild with childType "MapLayer" and the mock reader.
	// The drain loop will: call NextChild (returns "GrandChild", true),
	// then call FinishChild (returns error) → break.
	handled := m.DeserializeChild("MapLayer", rd)
	if !handled {
		t.Fatal("DeserializeChild should return true for MapLayer")
	}
	if len(m.Layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(m.Layers))
	}
}
