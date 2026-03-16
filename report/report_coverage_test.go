package report

// report_coverage_test.go — internal tests for uncovered branches in:
//   - borderfill_serial.go: formatLineStyle
//   - component.go: ComponentBase.Serialize/Deserialize
//   - reportcomponent.go: ReportComponentBase.Serialize/Deserialize

import (
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── formatLineStyle: call directly with each constant ────────────────────────

func TestFormatLineStyle_AllCases(t *testing.T) {
	cases := []struct {
		ls   style.LineStyle
		want string
	}{
		{style.LineStyleSolid, "Solid"},
		{style.LineStyleDash, "Dash"},
		{style.LineStyleDot, "Dot"},
		{style.LineStyleDashDot, "DashDot"},
		{style.LineStyleDashDotDot, "DashDotDot"},
		{style.LineStyleDouble, "Double"},
		{style.LineStyle(99), "Solid"}, // unknown → fallback Solid
	}
	for _, tc := range cases {
		got := formatLineStyle(tc.ls)
		if got != tc.want {
			t.Errorf("formatLineStyle(%d) = %q, want %q", tc.ls, got, tc.want)
		}
	}
}

// ── ComponentBase.Serialize: Left/Top non-zero ────────────────────────────────

func TestComponentBase_Serialize_LeftTopNonZero(t *testing.T) {
	c := NewComponentBase()
	c.SetLeft(10.5)
	c.SetTop(20.5)
	c.SetWidth(100)
	c.SetHeight(50)

	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["Left"]; !ok || v != float32(10.5) {
		t.Errorf("Left: got %v, want 10.5", v)
	}
	if v, ok := w.data["Top"]; !ok || v != float32(20.5) {
		t.Errorf("Top: got %v, want 20.5", v)
	}
}

// ── ComponentBase.Serialize: Visible/Printable false ─────────────────────────

func TestComponentBase_Serialize_InvisibleNotPrintable(t *testing.T) {
	c := NewComponentBase()
	c.SetVisible(false)
	c.SetPrintable(false)

	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["Visible"]; !ok || v != false {
		t.Errorf("Visible: got %v, want false", v)
	}
	if v, ok := w.data["Printable"]; !ok || v != false {
		t.Errorf("Printable: got %v, want false", v)
	}
}

// ── ComponentBase.Serialize: Anchor/Dock non-default, GroupIndex non-zero ─────

func TestComponentBase_Serialize_AnchorDockGroup(t *testing.T) {
	c := NewComponentBase()
	c.SetAnchor(AnchorLeft | AnchorRight) // != AnchorDefault
	c.SetDock(DockTop)                    // != DockNone
	c.SetGroupIndex(3)

	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["Anchor"]; !ok {
		t.Error("Anchor should be serialized when non-default")
	}
	if _, ok := w.data["Dock"]; !ok {
		t.Error("Dock should be serialized when non-default")
	}
	if v, ok := w.data["GroupIndex"]; !ok || v != 3 {
		t.Errorf("GroupIndex: got %v, want 3", v)
	}
}

// ── ComponentBase.Deserialize: Width/Height zero edge case ────────────────────

func TestComponentBase_Deserialize_ZeroWidthHeight(t *testing.T) {
	c := NewComponentBase()
	c.SetWidth(100)
	c.SetHeight(50)

	r := newTestReader(map[string]any{
		"Width":  float32(0),
		"Height": float32(0),
	})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Width() != 0 {
		t.Errorf("Width: got %v, want 0", c.Width())
	}
	if c.Height() != 0 {
		t.Errorf("Height: got %v, want 0", c.Height())
	}
}

// ── ComponentBase.Deserialize: VisibleExpression/PrintableExpression ──────────

func TestComponentBase_Deserialize_Expressions(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{
		"VisibleExpression":   "[ShowBand]",
		"PrintableExpression": "[IsPrintable]",
	})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.VisibleExpression() != "[ShowBand]" {
		t.Errorf("VisibleExpression: got %q", c.VisibleExpression())
	}
	if c.PrintableExpression() != "[IsPrintable]" {
		t.Errorf("PrintableExpression: got %q", c.PrintableExpression())
	}
}

// ── ReportComponentBase.Serialize: Hyperlink ──────────────────────────────────

func TestReportComponentBase_Serialize_HyperlinkCoverage(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetHyperlink(&Hyperlink{
		Kind:             "URL",
		Expression:       "[URL]",
		Value:            "https://example.com",
		Target:           "_blank",
		DetailPageName:   "Page2",
		DetailReportName: "Report2",
		ReportParameter:  "id",
	})

	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["Hyperlink.Kind"]; !ok || v != "URL" {
		t.Errorf("Hyperlink.Kind: got %v", v)
	}
	if v, ok := w.data["Hyperlink.Expression"]; !ok || v != "[URL]" {
		t.Errorf("Hyperlink.Expression: got %v", v)
	}
	if v, ok := w.data["Hyperlink.Value"]; !ok || v != "https://example.com" {
		t.Errorf("Hyperlink.Value: got %v", v)
	}
	if v, ok := w.data["Hyperlink.Target"]; !ok || v != "_blank" {
		t.Errorf("Hyperlink.Target: got %v", v)
	}
	if v, ok := w.data["Hyperlink.DetailPageName"]; !ok || v != "Page2" {
		t.Errorf("Hyperlink.DetailPageName: got %v", v)
	}
	if v, ok := w.data["Hyperlink.DetailReportName"]; !ok || v != "Report2" {
		t.Errorf("Hyperlink.DetailReportName: got %v", v)
	}
	if v, ok := w.data["Hyperlink.ReportParameter"]; !ok || v != "id" {
		t.Errorf("Hyperlink.ReportParameter: got %v", v)
	}
}

// ── ReportComponentBase.Deserialize: Hyperlink attributes ────────────────────

func TestReportComponentBase_Deserialize_HyperlinkCoverage(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"Hyperlink.Kind":             "URL",
		"Hyperlink.Expression":       "[Url]",
		"Hyperlink.Value":            "http://example.com",
		"Hyperlink.Target":           "_blank",
		"Hyperlink.DetailPageName":   "PageDetail",
		"Hyperlink.DetailReportName": "SubReport",
		"Hyperlink.ReportParameter":  "pid",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	h := rc.Hyperlink()
	if h == nil {
		t.Fatal("expected Hyperlink to be set")
	}
	if h.Kind != "URL" {
		t.Errorf("Kind: got %q, want URL", h.Kind)
	}
	if h.Expression != "[Url]" {
		t.Errorf("Expression: got %q, want [Url]", h.Expression)
	}
	if h.DetailPageName != "PageDetail" {
		t.Errorf("DetailPageName: got %q, want PageDetail", h.DetailPageName)
	}
	if h.ReportParameter != "pid" {
		t.Errorf("ReportParameter: got %q, want pid", h.ReportParameter)
	}
}

// ── ReportComponentBase.Deserialize: no hyperlink attributes → nil ────────────

func TestReportComponentBase_Deserialize_NoHyperlink(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink should be nil when no hyperlink attributes present")
	}
}

// ── BreakableComponent.Serialize/Deserialize: CanBreak=false round-trip ───────

func TestBreakableComponent_SerializeDeserialize_CanBreakFalse(t *testing.T) {
	bc := NewBreakableComponent()
	bc.SetCanBreak(false)

	w := newTestWriter()
	if err := bc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["CanBreak"]; !ok || v != false {
		t.Errorf("CanBreak: got %v, want false", v)
	}

	bc2 := NewBreakableComponent()
	r := newTestReader(w.data)
	if err := bc2.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if bc2.CanBreak() {
		t.Error("CanBreak should be false after deserializing false")
	}
}

// ── ReportComponentBase.Serialize: nil hyperlink does not serialize ───────────

func TestReportComponentBase_Serialize_NilHyperlink(t *testing.T) {
	rc := NewReportComponentBase()
	// hyperlink is nil by default; ensure Serialize doesn't panic
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for k := range w.data {
		if len(k) > 9 && k[:10] == "Hyperlink." {
			t.Errorf("unexpected Hyperlink key %q when hyperlink is nil", k)
		}
	}
}
