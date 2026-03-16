package report

// breakable_error_paths_test.go – internal tests targeting the uncovered branches
// in breakable.go lines 63-70 (Serialize) and 74-80 (Deserialize).
//
// Analysis of uncovered branches:
//
//   breakable.go:63 Serialize 80.0%
//   Uncovered: line 65 "return err" inside:
//     if err := bc.ReportComponentBase.Serialize(w); err != nil { return err }
//
//   breakable.go:74 Deserialize 75.0%
//   Uncovered: line 76 "return err" inside:
//     if err := bc.ReportComponentBase.Deserialize(r); err != nil { return err }
//
// Root cause: ReportComponentBase.Serialize and ReportComponentBase.Deserialize
// are ALWAYS nil-returning because the entire chain
// (ReportComponentBase → ComponentBase → BaseObject → void WriteXxx methods)
// has no write method that can return an error. These are defensive dead-code
// branches that cannot be executed with the current report.Writer interface.
//
// This file adds comprehensive tests for all REACHABLE branches in
// BreakableComponent.Serialize and BreakableComponent.Deserialize.

import (
	"testing"
)

// ─── BreakableComponent.Serialize: reachable branches ────────────────────────

// TestBreakableComponent_Serialize_CanBreakTrue verifies that CanBreak=true
// (the default) does NOT write the "CanBreak" key (it is only written when false).
func TestBreakableComponent_Serialize_CanBreakTrue_NoOutput(t *testing.T) {
	bc := NewBreakableComponent()
	// Default: canBreak = true.
	if !bc.CanBreak() {
		t.Fatal("precondition: CanBreak should default to true")
	}

	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// CanBreak=true is the default — must NOT be written.
	if _, ok := w.data["CanBreak"]; ok {
		t.Error("CanBreak=true should not be serialized (it is the default)")
	}
}

// TestBreakableComponent_Serialize_CanBreakFalse_Written verifies that
// CanBreak=false IS written.
func TestBreakableComponent_Serialize_CanBreakFalse_Written(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)

	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	v, ok := w.data["CanBreak"]
	if !ok {
		t.Fatal("CanBreak=false must be serialized")
	}
	if vb, ok := v.(bool); !ok || vb {
		t.Errorf("serialized CanBreak = %v, want false", v)
	}
}

// TestBreakableComponent_Serialize_ParentPropertiesAlsoWritten verifies that
// properties from the parent ReportComponentBase are also written during Serialize.
// (This exercises the call to bc.ReportComponentBase.Serialize which succeeds.)
func TestBreakableComponent_Serialize_ParentPropertiesAlsoWritten(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)
	bc.SetName("testBC")
	bc.SetCanGrow(true) // non-default ReportComponentBase property

	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// Both parent and own properties should be present.
	if _, ok := w.data["CanBreak"]; !ok {
		t.Error("CanBreak=false should be written")
	}
	if _, ok := w.data["Name"]; !ok {
		t.Error("Name should be written (from BaseObject)")
	}
	if _, ok := w.data["CanGrow"]; !ok {
		t.Error("CanGrow=true should be written (from ReportComponentBase)")
	}
}

// ─── BreakableComponent.Deserialize: reachable branches ──────────────────────

// TestBreakableComponent_Deserialize_CanBreakFalse verifies the CanBreak=false
// deserialization branch (non-default value).
func TestBreakableComponent_Deserialize_CanBreakFalse_Branch(t *testing.T) {
	bc := NewBreakableComponent()
	r := newTestReader(map[string]any{
		"CanBreak": false,
	})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if bc.CanBreak() {
		t.Error("CanBreak should be false after deserializing false")
	}
}

// TestBreakableComponent_Deserialize_CanBreakTrue_Default verifies that the
// default branch (CanBreak absent → true) is taken when the key is not present.
func TestBreakableComponent_Deserialize_CanBreakTrue_Default(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false) // change from default to verify the default is restored
	r := newTestReader(map[string]any{})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !bc.CanBreak() {
		t.Error("CanBreak should default to true when absent")
	}
}

// TestBreakableComponent_Deserialize_ParentPropertiesAlsoRead verifies that
// properties from the parent ReportComponentBase are also read during Deserialize.
// (This exercises the call to bc.ReportComponentBase.Deserialize which succeeds.)
func TestBreakableComponent_Deserialize_ParentPropertiesAlsoRead(t *testing.T) {
	bc := NewBreakableComponent()
	r := newTestReader(map[string]any{
		"CanBreak": false,
		"Name":     "myComponent",
		"CanGrow":  true,
	})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if bc.CanBreak() {
		t.Error("CanBreak should be false")
	}
	if bc.Name() != "myComponent" {
		t.Errorf("Name = %q, want myComponent", bc.Name())
	}
	if !bc.CanGrow() {
		t.Error("CanGrow should be true")
	}
}

// TestBreakableComponent_Serialize_Deserialize_RoundTrip performs a complete
// round-trip to verify Serialize and Deserialize are inverses.
func TestBreakableComponent_Serialize_Deserialize_RoundTrip_CanBreakFalse(t *testing.T) {
	bc1 := NewBreakableComponent()
	bc1.SetCanBreak(false)
	bc1.SetName("roundtrip")
	bc1.SetCanGrow(true)
	bc1.SetCanShrink(true)

	w := newTestWriter()
	if err := bc1.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	bc2 := NewBreakableComponent()
	r := newTestReader(w.data)
	if err := bc2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	if bc2.CanBreak() {
		t.Error("CanBreak should be false after round-trip")
	}
	if bc2.Name() != "roundtrip" {
		t.Errorf("Name = %q, want roundtrip", bc2.Name())
	}
	if !bc2.CanGrow() {
		t.Error("CanGrow should be true after round-trip")
	}
	if !bc2.CanShrink() {
		t.Error("CanShrink should be true after round-trip")
	}
}

// TestBreakableComponent_Serialize_Deserialize_RoundTrip_CanBreakTrue verifies
// the default CanBreak=true round-trip (CanBreak should NOT be written and
// should be read back as default true).
func TestBreakableComponent_Serialize_Deserialize_RoundTrip_CanBreakTrue(t *testing.T) {
	bc1 := NewBreakableComponent()
	// CanBreak=true is default; should not be written.

	w := newTestWriter()
	if err := bc1.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	if _, ok := w.data["CanBreak"]; ok {
		t.Error("CanBreak=true should not be written")
	}

	bc2 := NewBreakableComponent()
	r := newTestReader(w.data)
	if err := bc2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !bc2.CanBreak() {
		t.Error("CanBreak should default to true when not written")
	}
}
