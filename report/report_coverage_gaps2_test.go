package report

// report_coverage_gaps2_test.go — supplementary internal tests for the
// remaining coverage gaps in:
//
//   breakable.go:63   BreakableComponent.Serialize   (80.0% → all reachable branches covered here)
//   breakable.go:74   BreakableComponent.Deserialize (75.0% → all reachable branches covered here)
//   component.go:203  ComponentBase.Serialize        (96.0% → unreachable "return err" documented)
//   component.go:244  ComponentBase.Deserialize      (92.9% → unreachable "return err" documented)
//   reportcomponent.go:284 ReportComponentBase.Serialize  (97.7% → unreachable "return err" documented)
//   reportcomponent.go:354 ReportComponentBase.Deserialize (96.2% → unreachable "return err" documented)
//
// Why the remaining gaps cannot be fully eliminated:
//
//   Every "return err" branch in the Serialize / Deserialize chain is guarded by
//   a call to the parent type's Serialize or Deserialize method.  All those
//   parent methods return nil unconditionally because the Writer interface's
//   primitive write methods (WriteStr, WriteInt, WriteBool, WriteFloat) have
//   void return types — no error channel exists for those calls.  Likewise,
//   the Reader interface's ReadStr / ReadBool / ReadInt / ReadFloat return
//   values directly, never errors.  Therefore the if-err chains are defensive
//   dead code that exists for API consistency and future extensibility only.
//
//   This file maximises the reachable coverage by exercising every branch
//   that CAN be hit, so the function-level percentages are pushed as high
//   as possible without modifying the production code.

import (
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// BreakableComponent.Serialize — all reachable branches
// ─────────────────────────────────────────────────────────────────────────────

// TestBreakableSerialize_CanBreakTrue_NothingWritten re-tests that the
// CanBreak guard's "false" arm (default — don't write) is exercised.
func TestBreakableSerialize_CanBreakTrue_NothingWritten(t *testing.T) {
	bc := NewBreakableComponent()
	// canBreak is true by default; the guard `if !bc.canBreak` should be false.
	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["CanBreak"]; ok {
		t.Error("CanBreak=true should NOT be written (it is the default)")
	}
}

// TestBreakableSerialize_CanBreakFalse_Written re-tests the "true" arm of
// the CanBreak guard (writes the attribute).
func TestBreakableSerialize_CanBreakFalse_Written(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)
	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	v, ok := w.data["CanBreak"]
	if !ok {
		t.Fatal("CanBreak=false must be written")
	}
	if b, ok2 := v.(bool); !ok2 || b {
		t.Errorf("CanBreak written as %v, want false", v)
	}
}

// TestBreakableSerialize_InheritsParent verifies that parent properties
// (from ReportComponentBase) are also serialized.
func TestBreakableSerialize_InheritsParent(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetName("BC1")
	bc.SetCanBreak(false)
	bc.SetCanGrow(true)
	bc.SetCanShrink(true)
	bc.SetBookmark("bkmark")
	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, key := range []string{"Name", "CanBreak", "CanGrow", "CanShrink", "Bookmark"} {
		if _, ok := w.data[key]; !ok {
			t.Errorf("key %q should be written", key)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BreakableComponent.Deserialize — all reachable branches
// ─────────────────────────────────────────────────────────────────────────────

// TestBreakableDeserialize_CanBreakTrueDefault tests the default path where
// CanBreak is absent in the reader → defaults to true.
func TestBreakableDeserialize_CanBreakTrueDefault(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false) // set to non-default so we can verify the reset
	r := newTestReader(map[string]any{})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !bc.CanBreak() {
		t.Error("CanBreak should default to true when absent from reader")
	}
}

// TestBreakableDeserialize_CanBreakFalse tests reading an explicit false value.
func TestBreakableDeserialize_CanBreakFalse(t *testing.T) {
	bc := NewBreakableComponent()
	r := newTestReader(map[string]any{"CanBreak": false})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if bc.CanBreak() {
		t.Error("CanBreak should be false after reading false")
	}
}

// TestBreakableDeserialize_WithParentFields verifies that inherited fields from
// ReportComponentBase are also deserialized when Deserialize is called.
func TestBreakableDeserialize_WithParentFields(t *testing.T) {
	bc := NewBreakableComponent()
	r := newTestReader(map[string]any{
		"Name":     "myBreakable",
		"CanBreak": false,
		"CanGrow":  true,
		"PrintOn":  int(PrintOnFirstPage),
	})
	if err := bc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if bc.Name() != "myBreakable" {
		t.Errorf("Name = %q, want myBreakable", bc.Name())
	}
	if bc.CanBreak() {
		t.Error("CanBreak should be false")
	}
	if !bc.CanGrow() {
		t.Error("CanGrow should be true")
	}
	if bc.PrintOn() != PrintOnFirstPage {
		t.Errorf("PrintOn = %d, want PrintOnFirstPage", bc.PrintOn())
	}
}

// TestBreakableDeserialize_RoundTrip verifies that Serialize followed by
// Deserialize produces an identical object.
func TestBreakableDeserialize_RoundTrip(t *testing.T) {
	src := NewBreakableComponent()
	src.SetName("rt")
	src.SetCanBreak(false)
	src.SetCanGrow(true)
	src.SetCanShrink(true)
	src.SetExportable(false)
	src.SetPrintOn(PrintOnLastPage)

	w := newTestWriter()
	if err := src.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	dst := NewBreakableComponent()
	r := newTestReader(w.data)
	if err := dst.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if dst.Name() != "rt" {
		t.Errorf("Name = %q, want rt", dst.Name())
	}
	if dst.CanBreak() {
		t.Error("CanBreak should be false after round-trip")
	}
	if !dst.CanGrow() {
		t.Error("CanGrow should be true after round-trip")
	}
	if !dst.CanShrink() {
		t.Error("CanShrink should be true after round-trip")
	}
	if dst.Exportable() {
		t.Error("Exportable should be false after round-trip")
	}
	if dst.PrintOn() != PrintOnLastPage {
		t.Errorf("PrintOn = %d, want PrintOnLastPage", dst.PrintOn())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ComponentBase.Serialize — additional edge cases
// ─────────────────────────────────────────────────────────────────────────────

// TestComponentBaseSerialize_ZeroLeftTopNotWritten ensures zero values for
// Left and Top are not serialized (already at default, guard is false).
func TestComponentBaseSerialize_ZeroLeftTopNotWritten(t *testing.T) {
	c := NewComponentBase()
	// Left=0, Top=0 are defaults.
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, k := range []string{"Left", "Top", "Width", "Height"} {
		if _, ok := w.data[k]; ok {
			t.Errorf("%q should not be written when zero", k)
		}
	}
}

// TestComponentBaseSerialize_AnchorDefault_NotWritten exercises the anchor
// guard's "false" arm (AnchorDefault → not written).
func TestComponentBaseSerialize_AnchorDefault_NotWritten(t *testing.T) {
	c := NewComponentBase()
	// Anchor == AnchorDefault by default.
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Anchor"]; ok {
		t.Error("Anchor should not be written when at default value")
	}
}

// TestComponentBaseSerialize_DockNone_NotWritten exercises the dock guard's
// "false" arm (DockNone → not written).
func TestComponentBaseSerialize_DockNone_NotWritten(t *testing.T) {
	c := NewComponentBase()
	// Dock == DockNone by default.
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Dock"]; ok {
		t.Error("Dock should not be written when DockNone")
	}
}

// TestComponentBaseSerialize_VisibleTrue_NotWritten exercises the visible
// guard's "false" arm (visible=true → not written).
func TestComponentBaseSerialize_VisibleTrue_NotWritten(t *testing.T) {
	c := NewComponentBase()
	// Visible == true by default.
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Visible"]; ok {
		t.Error("Visible=true should not be written")
	}
}

// TestComponentBaseSerialize_PrintableTrue_NotWritten exercises the printable
// guard's "false" arm (printable=true → not written).
func TestComponentBaseSerialize_PrintableTrue_NotWritten(t *testing.T) {
	c := NewComponentBase()
	// Printable == true by default.
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Printable"]; ok {
		t.Error("Printable=true should not be written")
	}
}

// TestComponentBaseSerialize_AllGuardsTrue exercises every guard's "true" arm
// in a single Serialize call.
func TestComponentBaseSerialize_AllGuardsTrue(t *testing.T) {
	c := NewComponentBase()
	c.SetLeft(1)
	c.SetTop(2)
	c.SetWidth(3)
	c.SetHeight(4)
	c.SetVisible(false)
	c.SetPrintable(false)
	c.SetVisibleExpression("[V]")
	c.SetPrintableExpression("[P]")
	c.SetAnchor(AnchorNone)
	c.SetDock(DockFill)
	c.SetGroupIndex(5)
	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, k := range []string{
		"Left", "Top", "Width", "Height",
		"Visible", "Printable",
		"VisibleExpression", "PrintableExpression",
		"Anchor", "Dock", "GroupIndex",
	} {
		if _, ok := w.data[k]; !ok {
			t.Errorf("key %q should be written", k)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ComponentBase.Deserialize — additional edge cases
// ─────────────────────────────────────────────────────────────────────────────

// TestComponentBaseDeserialize_VisibleFalse verifies the "Visible false" branch.
func TestComponentBaseDeserialize_VisibleFalse(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{"Visible": false})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Visible() {
		t.Error("Visible should be false")
	}
}

// TestComponentBaseDeserialize_PrintableFalse verifies the "Printable false" branch.
func TestComponentBaseDeserialize_PrintableFalse(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{"Printable": false})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Printable() {
		t.Error("Printable should be false")
	}
}

// TestComponentBaseDeserialize_AllFields exercises every field read.
func TestComponentBaseDeserialize_AllFields(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{
		"Left":                float32(10),
		"Top":                 float32(20),
		"Width":               float32(300),
		"Height":              float32(100),
		"Visible":             false,
		"Printable":           false,
		"VisibleExpression":   "[VE2]",
		"PrintableExpression": "[PE2]",
		"Anchor":              int(AnchorBottom | AnchorRight),
		"Dock":                int(DockLeft),
		"GroupIndex":          99,
	})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Left() != 10 {
		t.Errorf("Left = %v, want 10", c.Left())
	}
	if c.Top() != 20 {
		t.Errorf("Top = %v, want 20", c.Top())
	}
	if c.Width() != 300 {
		t.Errorf("Width = %v, want 300", c.Width())
	}
	if c.Height() != 100 {
		t.Errorf("Height = %v, want 100", c.Height())
	}
	if c.Visible() {
		t.Error("Visible should be false")
	}
	if c.Printable() {
		t.Error("Printable should be false")
	}
	if c.VisibleExpression() != "[VE2]" {
		t.Errorf("VisibleExpression = %q, want [VE2]", c.VisibleExpression())
	}
	if c.PrintableExpression() != "[PE2]" {
		t.Errorf("PrintableExpression = %q, want [PE2]", c.PrintableExpression())
	}
	if c.Anchor() != AnchorBottom|AnchorRight {
		t.Errorf("Anchor = %d", c.Anchor())
	}
	if c.Dock() != DockLeft {
		t.Errorf("Dock = %d, want DockLeft", c.Dock())
	}
	if c.GroupIndex() != 99 {
		t.Errorf("GroupIndex = %d, want 99", c.GroupIndex())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ReportComponentBase.Serialize — additional edge cases
// ─────────────────────────────────────────────────────────────────────────────

// TestReportComponentBaseSerialize_Exportable_Default_NotWritten checks that
// the exportable guard's "false" arm is taken when exportable=true (default).
func TestReportComponentBaseSerialize_Exportable_Default_NotWritten(t *testing.T) {
	rc := NewReportComponentBase()
	// exportable=true is default → not written.
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Exportable"]; ok {
		t.Error("Exportable=true should not be written")
	}
}

// TestReportComponentBaseSerialize_CanGrow_Default_NotWritten checks the
// canGrow guard's "false" arm.
func TestReportComponentBaseSerialize_CanGrow_Default_NotWritten(t *testing.T) {
	rc := NewReportComponentBase()
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["CanGrow"]; ok {
		t.Error("CanGrow=false should not be written")
	}
}

// TestReportComponentBaseSerialize_ShiftNever_NotWritten exercises the
// shiftMode guard's "false" arm (ShiftNever → not written).
func TestReportComponentBaseSerialize_ShiftNever_NotWritten(t *testing.T) {
	rc := NewReportComponentBase()
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["ShiftMode"]; ok {
		t.Error("ShiftMode=ShiftNever should not be written")
	}
}

// TestReportComponentBaseSerialize_PrintOnAllPages_NotWritten exercises the
// printOn guard's "false" arm (PrintOnAllPages → not written).
func TestReportComponentBaseSerialize_PrintOnAllPages_NotWritten(t *testing.T) {
	rc := NewReportComponentBase()
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["PrintOn"]; ok {
		t.Error("PrintOn=PrintOnAllPages should not be written")
	}
}

// TestReportComponentBaseSerialize_NilHyperlink_NoHyperlinkKeys exercises the
// hyperlink nil guard (rc.hyperlink == nil → none of the hyperlink keys written).
func TestReportComponentBaseSerialize_NilHyperlink_NoHyperlinkKeys(t *testing.T) {
	rc := NewReportComponentBase()
	// hyperlink is nil by default.
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for k := range w.data {
		if len(k) >= 10 && k[:10] == "Hyperlink." {
			t.Errorf("unexpected Hyperlink key %q written when hyperlink is nil", k)
		}
	}
}

// TestReportComponentBaseSerialize_HyperlinkAllEmptyFields exercises the
// hyperlink field guards when the Hyperlink struct exists but all fields are
// empty → no hyperlink keys written.
func TestReportComponentBaseSerialize_HyperlinkAllEmptyFields(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetHyperlink(&Hyperlink{}) // all fields empty strings
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for k := range w.data {
		if len(k) >= 10 && k[:10] == "Hyperlink." {
			t.Errorf("unexpected Hyperlink key %q written when all Hyperlink fields empty", k)
		}
	}
}

// TestReportComponentBaseSerialize_AllGuardsTrue exercises every non-default
// branch in a single call, including all hyperlink sub-fields.
func TestReportComponentBaseSerialize_AllGuardsTrue(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetExportable(false)
	rc.SetExportableExpression("[E]")
	rc.SetCanGrow(true)
	rc.SetCanShrink(true)
	rc.SetGrowToBottom(true)
	rc.SetShiftMode(ShiftAlways)
	rc.SetPrintOn(PrintOnFirstPage)
	rc.SetPageBreak(true)
	rc.SetStyleName("S1")
	rc.SetEvenStyleName("S2")
	rc.SetHoverStyleName("S3")
	rc.SetBookmark("bk")
	rc.SetHyperlink(&Hyperlink{
		Kind:             "URL",
		Expression:       "[url]",
		Value:            "http://x",
		Target:           "_blank",
		DetailPageName:   "dp",
		DetailReportName: "dr",
		ReportParameter:  "rp",
	})
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, k := range []string{
		"Exportable", "ExportableExpression",
		"CanGrow", "CanShrink", "GrowToBottom",
		"ShiftMode", "PrintOn", "PageBreak",
		"Style", "EvenStyle", "HoverStyle", "Bookmark",
		"Hyperlink.Kind", "Hyperlink.Expression", "Hyperlink.Value",
		"Hyperlink.Target", "Hyperlink.DetailPageName",
		"Hyperlink.DetailReportName", "Hyperlink.ReportParameter",
	} {
		if _, ok := w.data[k]; !ok {
			t.Errorf("key %q should be written", k)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ReportComponentBase.Deserialize — additional edge cases
// ─────────────────────────────────────────────────────────────────────────────

// TestReportComponentBaseDeserialize_Defaults exercises all the "false" arms
// of the optional-field guards when the reader has no hyperlink attributes.
func TestReportComponentBaseDeserialize_Defaults(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !rc.Exportable() {
		t.Error("Exportable default should be true")
	}
	if rc.CanGrow() {
		t.Error("CanGrow default should be false")
	}
	if rc.CanShrink() {
		t.Error("CanShrink default should be false")
	}
	if rc.GrowToBottom() {
		t.Error("GrowToBottom default should be false")
	}
	if rc.ShiftMode() != ShiftNever {
		t.Errorf("ShiftMode default = %d, want ShiftNever", rc.ShiftMode())
	}
	if rc.PrintOn() != PrintOnAllPages {
		t.Errorf("PrintOn default = %d, want PrintOnAllPages", rc.PrintOn())
	}
	if rc.PageBreak() {
		t.Error("PageBreak default should be false")
	}
	if rc.StyleName() != "" {
		t.Errorf("StyleName default = %q, want empty", rc.StyleName())
	}
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink should be nil when no Hyperlink.* attributes present")
	}
}

// TestReportComponentBaseDeserialize_HyperlinkCreatedByAnyField exercises the
// "create Hyperlink" branch (the big OR condition) for each individual field.
func TestReportComponentBaseDeserialize_HyperlinkCreatedByAnyField(t *testing.T) {
	fields := []struct {
		key string
		val string
	}{
		{"Hyperlink.Kind", "Bookmark"},
		{"Hyperlink.Expression", "[E]"},
		{"Hyperlink.Value", "http://v"},
		{"Hyperlink.Target", "_top"},
		{"Hyperlink.DetailPageName", "pg"},
		{"Hyperlink.DetailReportName", "rpt"},
		{"Hyperlink.ReportParameter", "p"},
	}
	for _, f := range fields {
		t.Run(f.key, func(t *testing.T) {
			rc := NewReportComponentBase()
			r := newTestReader(map[string]any{f.key: f.val})
			if err := rc.Deserialize(r); err != nil {
				t.Fatalf("Deserialize: %v", err)
			}
			if rc.Hyperlink() == nil {
				t.Errorf("Hyperlink should be non-nil when %q is set", f.key)
			}
		})
	}
}

// TestReportComponentBaseDeserialize_AllFields exercises every field in a
// single Deserialize call.
func TestReportComponentBaseDeserialize_AllFields(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"Exportable":              false,
		"ExportableExpression":    "[CE]",
		"CanGrow":                 true,
		"CanShrink":               true,
		"GrowToBottom":            true,
		"ShiftMode":               int(ShiftWhenOverlapped),
		"PrintOn":                 int(PrintOnOddPages),
		"PageBreak":               true,
		"Style":                   "ST",
		"EvenStyle":               "ES",
		"HoverStyle":              "HS",
		"Bookmark":                "BK",
		"Hyperlink.Kind":          "URL",
		"Hyperlink.Expression":    "[x]",
		"Hyperlink.Value":         "http://u",
		"Hyperlink.Target":        "_self",
		"Hyperlink.DetailPageName":   "p",
		"Hyperlink.DetailReportName": "r",
		"Hyperlink.ReportParameter":  "id",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.Exportable() {
		t.Error("Exportable should be false")
	}
	if rc.ExportableExpression() != "[CE]" {
		t.Errorf("ExportableExpression = %q", rc.ExportableExpression())
	}
	if !rc.CanGrow() {
		t.Error("CanGrow should be true")
	}
	if !rc.CanShrink() {
		t.Error("CanShrink should be true")
	}
	if !rc.GrowToBottom() {
		t.Error("GrowToBottom should be true")
	}
	if rc.ShiftMode() != ShiftWhenOverlapped {
		t.Errorf("ShiftMode = %d", rc.ShiftMode())
	}
	if rc.PrintOn() != PrintOnOddPages {
		t.Errorf("PrintOn = %d", rc.PrintOn())
	}
	if !rc.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if rc.StyleName() != "ST" {
		t.Errorf("StyleName = %q, want ST", rc.StyleName())
	}
	if rc.EvenStyleName() != "ES" {
		t.Errorf("EvenStyleName = %q, want ES", rc.EvenStyleName())
	}
	if rc.HoverStyleName() != "HS" {
		t.Errorf("HoverStyleName = %q, want HS", rc.HoverStyleName())
	}
	if rc.Bookmark() != "BK" {
		t.Errorf("Bookmark = %q, want BK", rc.Bookmark())
	}
	h := rc.Hyperlink()
	if h == nil {
		t.Fatal("Hyperlink should not be nil")
	}
	if h.Kind != "URL" {
		t.Errorf("Hyperlink.Kind = %q, want URL", h.Kind)
	}
	if h.Expression != "[x]" {
		t.Errorf("Hyperlink.Expression = %q, want [x]", h.Expression)
	}
	if h.Value != "http://u" {
		t.Errorf("Hyperlink.Value = %q, want http://u", h.Value)
	}
	if h.Target != "_self" {
		t.Errorf("Hyperlink.Target = %q, want _self", h.Target)
	}
	if h.DetailPageName != "p" {
		t.Errorf("Hyperlink.DetailPageName = %q, want p", h.DetailPageName)
	}
	if h.DetailReportName != "r" {
		t.Errorf("Hyperlink.DetailReportName = %q, want r", h.DetailReportName)
	}
	if h.ReportParameter != "id" {
		t.Errorf("Hyperlink.ReportParameter = %q, want id", h.ReportParameter)
	}
}

// TestReportComponentBaseDeserialize_HyperlinkNil_AllEmpty exercises the
// "do not create Hyperlink" arm: all hyperlink attribute reads return empty.
func TestReportComponentBaseDeserialize_HyperlinkNil_AllEmpty(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"Style": "Bold", // non-hyperlink attribute
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink should be nil when all hyperlink attributes are empty")
	}
}
