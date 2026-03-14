package report

import (
	"testing"
)

func TestNewBreakableComponent_Defaults(t *testing.T) {
	bc := NewBreakableComponent()
	if bc == nil {
		t.Fatal("NewBreakableComponent returned nil")
	}
	if !bc.CanBreak() {
		t.Error("CanBreak should default to true")
	}
	if bc.BreakTo() != nil {
		t.Error("BreakTo should default to nil")
	}
	if bc.FlagMustBreak {
		t.Error("FlagMustBreak should default to false")
	}
	// Inherits from ReportComponentBase defaults.
	if !bc.Visible() {
		t.Error("Visible should inherit default true")
	}
	if !bc.Exportable() {
		t.Error("Exportable should inherit default true")
	}
}

func TestBreakableComponent_CanBreak(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)
	if bc.CanBreak() {
		t.Error("CanBreak should be false after SetCanBreak(false)")
	}
	bc.SetCanBreak(true)
	if !bc.CanBreak() {
		t.Error("CanBreak should be true after SetCanBreak(true)")
	}
}

func TestBreakableComponent_BreakTo(t *testing.T) {
	bc1 := NewBreakableComponent()
	bc2 := NewBreakableComponent()

	bc1.SetBreakTo(bc2)
	if bc1.BreakTo() != bc2 {
		t.Error("BreakTo should return the set target")
	}

	bc1.SetBreakTo(nil)
	if bc1.BreakTo() != nil {
		t.Error("BreakTo should be nil after SetBreakTo(nil)")
	}
}

func TestBreakableComponent_Break_DefaultFalse(t *testing.T) {
	bc := NewBreakableComponent()
	overflow := NewBreakableComponent()
	// Base implementation always returns false.
	if bc.Break(overflow) {
		t.Error("base Break should return false")
	}
}

func TestBreakableComponent_Break_NilTarget(t *testing.T) {
	bc := NewBreakableComponent()
	// Should not panic with nil target.
	result := bc.Break(nil)
	if result {
		t.Error("base Break with nil target should return false")
	}
}

func TestBreakableComponent_CalcHeight_DefaultIsHeight(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetHeight(120)
	if bc.CalcHeight() != 120 {
		t.Errorf("CalcHeight = %v, want 120", bc.CalcHeight())
	}
}

func TestBreakableComponent_FlagMustBreak(t *testing.T) {
	bc := NewBreakableComponent()
	bc.FlagMustBreak = true
	if !bc.FlagMustBreak {
		t.Error("FlagMustBreak should be true after setting")
	}
}

func TestBreakableComponent_Serialize_DefaultCanBreak(t *testing.T) {
	bc := NewBreakableComponent()
	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// CanBreak=true is the default — should not be written.
	if _, ok := w.data["CanBreak"]; ok {
		t.Error("CanBreak=true should not be serialized (it is the default)")
	}
}

func TestBreakableComponent_Serialize_CanBreakFalse(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)
	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	v, ok := w.data["CanBreak"]
	if !ok {
		t.Fatal("CanBreak=false should be serialized")
	}
	if v != false {
		t.Errorf("serialized CanBreak = %v, want false", v)
	}
}

func TestBreakableComponent_Deserialize(t *testing.T) {
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

func TestBreakableComponent_Deserialize_Defaults(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false) // change from default
	r := newTestReader(map[string]any{})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	// Missing key → default true.
	if !bc.CanBreak() {
		t.Error("CanBreak should default to true when key is absent")
	}
}

func TestBreakableComponent_InheritsReportComponentBase(t *testing.T) {
	bc := NewBreakableComponent()
	// Verify inherited properties are accessible.
	bc.SetCanGrow(true)
	if !bc.CanGrow() {
		t.Error("CanGrow should be settable via inherited ReportComponentBase")
	}
	bc.SetPrintOn(PrintOnFirstPage)
	if bc.PrintOn() != PrintOnFirstPage {
		t.Errorf("PrintOn = %d, want PrintOnFirstPage", bc.PrintOn())
	}
}
