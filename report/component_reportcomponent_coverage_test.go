package report

// component_reportcomponent_coverage_test.go — internal tests targeting every
// reachable statement in:
//   - component.go:    ComponentBase.Serialize (line 203)  and Deserialize (line 244)
//   - reportcomponent.go: ReportComponentBase.Serialize (line 284) and Deserialize (line 354)
//
// Analysis of the remaining uncovered statements:
//
//   component.go:205        "return err" inside:
//     if err := c.BaseObject.Serialize(w); err != nil { return err }
//
//   component.go:246        "return err" inside:
//     if err := c.BaseObject.Deserialize(r); err != nil { return err }
//
//   reportcomponent.go:286  "return err" inside:
//     if err := rc.ComponentBase.Serialize(w); err != nil { return err }
//
//   reportcomponent.go:356  "return err" inside:
//     if err := rc.ComponentBase.Deserialize(r); err != nil { return err }
//
// Root cause: BaseObject.Serialize and BaseObject.Deserialize ALWAYS return nil.
// The Writer and Reader interfaces do not expose error-returning write methods
// (WriteStr, WriteBool, WriteInt, WriteFloat all return void), so there is no
// mechanism by which BaseObject.Serialize or BaseObject.Deserialize can fail.
// These "return err" branches are defensive dead-code added for API consistency
// and future extensibility; they are structurally unreachable with the current
// interface design.
//
// All other statements within these four functions are fully exercised below.

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ─────────────────────────────────────────────────────────────────────────────
// ComponentBase.Serialize — exhaustive reachable branch coverage
// ─────────────────────────────────────────────────────────────────────────────

// TestComponentBase_Serialize_AllNonDefaultBranches exercises every conditional
// branch that writes an attribute (the "true" arm of each if-guard).
func TestComponentBase_Serialize_AllNonDefaultBranches(t *testing.T) {
	c := NewComponentBase()
	c.SetLeft(5.5)
	c.SetTop(10.25)
	c.SetWidth(200)
	c.SetHeight(150)
	c.SetVisible(false)
	c.SetPrintable(false)
	c.SetVisibleExpression("[ShowIt]")
	c.SetPrintableExpression("[PrintIt]")
	c.SetAnchor(AnchorRight | AnchorBottom) // != AnchorDefault (Left|Top)
	c.SetDock(DockFill)                     // != DockNone
	c.SetGroupIndex(7)
	c.SetName("myComp")

	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	mustHaveFloat := func(key string, want float32) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %v", key, v, want)
		}
	}
	mustHaveBool := func(key string, want bool) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %v", key, v, want)
		}
	}
	mustHaveStr := func(key, want string) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %q", key, v, want)
		}
	}
	mustHaveInt := func(key string, want int) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %d", key, v, want)
		}
	}

	mustHaveStr("Name", "myComp")
	mustHaveFloat("Left", 5.5)
	mustHaveFloat("Top", 10.25)
	mustHaveFloat("Width", 200)
	mustHaveFloat("Height", 150)
	mustHaveBool("Visible", false)
	mustHaveBool("Printable", false)
	mustHaveStr("VisibleExpression", "[ShowIt]")
	mustHaveStr("PrintableExpression", "[PrintIt]")
	mustHaveInt("Anchor", int(AnchorRight|AnchorBottom))
	mustHaveInt("Dock", int(DockFill))
	mustHaveInt("GroupIndex", 7)
}

// TestComponentBase_Serialize_DefaultsProducesNoOptionalKeys exercises the "false"
// arm of each if-guard: when values are at their defaults, nothing extra is written.
func TestComponentBase_Serialize_DefaultsProducesNoOptionalKeys(t *testing.T) {
	c := NewComponentBase()
	// All defaults: left=0, top=0, width=0, height=0, visible=true, printable=true,
	// expressions="", anchor=AnchorDefault, dock=DockNone, groupIndex=0.

	w := newTestWriter()
	if err := c.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	for _, key := range []string{
		"Left", "Top", "Width", "Height",
		"Visible", "Printable",
		"VisibleExpression", "PrintableExpression",
		"Anchor", "Dock", "GroupIndex",
	} {
		if _, ok := w.data[key]; ok {
			t.Errorf("key %q should not be written when at default value", key)
		}
	}
}

// TestComponentBase_Serialize_PartialNonDefaults exercises combinations to ensure
// each branch is independently reachable (e.g. only Left is non-zero).
func TestComponentBase_Serialize_PartialNonDefaults(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*ComponentBase)
		checkFn func(t *testing.T, data map[string]any)
	}{
		{
			name:  "only Left non-zero",
			setup: func(c *ComponentBase) { c.SetLeft(42) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Left"]; !ok {
					t.Error("Left should be written")
				}
				if _, ok := data["Top"]; ok {
					t.Error("Top should not be written")
				}
			},
		},
		{
			name:  "only Top non-zero",
			setup: func(c *ComponentBase) { c.SetTop(99) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Top"]; !ok {
					t.Error("Top should be written")
				}
				if _, ok := data["Left"]; ok {
					t.Error("Left should not be written")
				}
			},
		},
		{
			name:  "only Width non-zero",
			setup: func(c *ComponentBase) { c.SetWidth(300) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Width"]; !ok {
					t.Error("Width should be written")
				}
			},
		},
		{
			name:  "only Height non-zero",
			setup: func(c *ComponentBase) { c.SetHeight(200) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Height"]; !ok {
					t.Error("Height should be written")
				}
			},
		},
		{
			name:  "anchor non-default (AnchorNone)",
			setup: func(c *ComponentBase) { c.SetAnchor(AnchorNone) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Anchor"]; !ok {
					t.Error("Anchor should be written when != AnchorDefault")
				}
			},
		},
		{
			name:  "dock DockLeft",
			setup: func(c *ComponentBase) { c.SetDock(DockLeft) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["Dock"]; !ok {
					t.Error("Dock should be written for DockLeft")
				}
			},
		},
		{
			name:  "groupIndex non-zero",
			setup: func(c *ComponentBase) { c.SetGroupIndex(1) },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["GroupIndex"]; !ok {
					t.Error("GroupIndex should be written when non-zero")
				}
			},
		},
		{
			name:  "visible expression only",
			setup: func(c *ComponentBase) { c.SetVisibleExpression("[x]") },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["VisibleExpression"]; !ok {
					t.Error("VisibleExpression should be written when non-empty")
				}
			},
		},
		{
			name:  "printable expression only",
			setup: func(c *ComponentBase) { c.SetPrintableExpression("[y]") },
			checkFn: func(t *testing.T, data map[string]any) {
				t.Helper()
				if _, ok := data["PrintableExpression"]; !ok {
					t.Error("PrintableExpression should be written when non-empty")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := NewComponentBase()
			tc.setup(c)
			w := newTestWriter()
			if err := c.Serialize(w); err != nil {
				t.Fatalf("Serialize: %v", err)
			}
			tc.checkFn(t, w.data)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ComponentBase.Deserialize — exhaustive reachable branch coverage
// ─────────────────────────────────────────────────────────────────────────────

// TestComponentBase_Deserialize_AllFields verifies every field is read correctly.
func TestComponentBase_Deserialize_AllFields(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{
		"Name":                "comp1",
		"Left":                float32(11.5),
		"Top":                 float32(22.75),
		"Width":               float32(300),
		"Height":              float32(200),
		"Visible":             false,
		"Printable":           false,
		"VisibleExpression":   "[VE]",
		"PrintableExpression": "[PE]",
		"Anchor":              int(AnchorLeft | AnchorBottom),
		"Dock":                int(DockRight),
		"GroupIndex":          42,
	})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if c.Name() != "comp1" {
		t.Errorf("Name = %q, want comp1", c.Name())
	}
	if c.Left() != 11.5 {
		t.Errorf("Left = %v, want 11.5", c.Left())
	}
	if c.Top() != 22.75 {
		t.Errorf("Top = %v, want 22.75", c.Top())
	}
	if c.Width() != 300 {
		t.Errorf("Width = %v, want 300", c.Width())
	}
	if c.Height() != 200 {
		t.Errorf("Height = %v, want 200", c.Height())
	}
	if c.Visible() {
		t.Error("Visible should be false")
	}
	if c.Printable() {
		t.Error("Printable should be false")
	}
	if c.VisibleExpression() != "[VE]" {
		t.Errorf("VisibleExpression = %q, want [VE]", c.VisibleExpression())
	}
	if c.PrintableExpression() != "[PE]" {
		t.Errorf("PrintableExpression = %q, want [PE]", c.PrintableExpression())
	}
	if c.Anchor() != AnchorLeft|AnchorBottom {
		t.Errorf("Anchor = %d, want AnchorLeft|AnchorBottom", c.Anchor())
	}
	if c.Dock() != DockRight {
		t.Errorf("Dock = %d, want DockRight", c.Dock())
	}
	if c.GroupIndex() != 42 {
		t.Errorf("GroupIndex = %d, want 42", c.GroupIndex())
	}
}

// TestComponentBase_Deserialize_DefaultValues verifies that when no attributes are
// present the component retains the constructor defaults.
func TestComponentBase_Deserialize_DefaultValues(t *testing.T) {
	c := NewComponentBase()
	r := newTestReader(map[string]any{})
	if err := c.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if c.Left() != 0 {
		t.Errorf("Left default = %v, want 0", c.Left())
	}
	if c.Top() != 0 {
		t.Errorf("Top default = %v, want 0", c.Top())
	}
	if c.Width() != 0 {
		t.Errorf("Width default = %v, want 0", c.Width())
	}
	if c.Height() != 0 {
		t.Errorf("Height default = %v, want 0", c.Height())
	}
	if !c.Visible() {
		t.Error("Visible default should be true")
	}
	if !c.Printable() {
		t.Error("Printable default should be true")
	}
	if c.VisibleExpression() != "" {
		t.Errorf("VisibleExpression default = %q, want empty", c.VisibleExpression())
	}
	if c.PrintableExpression() != "" {
		t.Errorf("PrintableExpression default = %q, want empty", c.PrintableExpression())
	}
	if c.Anchor() != AnchorDefault {
		t.Errorf("Anchor default = %d, want AnchorDefault", c.Anchor())
	}
	if c.Dock() != DockNone {
		t.Errorf("Dock default = %d, want DockNone", c.Dock())
	}
	if c.GroupIndex() != 0 {
		t.Errorf("GroupIndex default = %d, want 0", c.GroupIndex())
	}
}

// TestComponentBase_SerializeDeserialize_FullRoundTrip verifies a complete
// non-default round-trip is lossless.
func TestComponentBase_SerializeDeserialize_FullRoundTrip(t *testing.T) {
	src := NewComponentBase()
	src.SetName("roundtrip")
	src.SetLeft(15)
	src.SetTop(30)
	src.SetWidth(400)
	src.SetHeight(250)
	src.SetVisible(false)
	src.SetPrintable(false)
	src.SetVisibleExpression("[Page] > 1")
	src.SetPrintableExpression("[Total] > 0")
	src.SetAnchor(AnchorRight)
	src.SetDock(DockBottom)
	src.SetGroupIndex(9)

	w := newTestWriter()
	if err := src.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	dst := NewComponentBase()
	r := newTestReader(w.data)
	if err := dst.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if dst.Name() != "roundtrip" {
		t.Errorf("Name = %q, want roundtrip", dst.Name())
	}
	if dst.Left() != 15 {
		t.Errorf("Left = %v, want 15", dst.Left())
	}
	if dst.Top() != 30 {
		t.Errorf("Top = %v, want 30", dst.Top())
	}
	if dst.Width() != 400 {
		t.Errorf("Width = %v, want 400", dst.Width())
	}
	if dst.Height() != 250 {
		t.Errorf("Height = %v, want 250", dst.Height())
	}
	if dst.Visible() {
		t.Error("Visible should be false")
	}
	if dst.Printable() {
		t.Error("Printable should be false")
	}
	if dst.VisibleExpression() != "[Page] > 1" {
		t.Errorf("VisibleExpression = %q", dst.VisibleExpression())
	}
	if dst.PrintableExpression() != "[Total] > 0" {
		t.Errorf("PrintableExpression = %q", dst.PrintableExpression())
	}
	if dst.Anchor() != AnchorRight {
		t.Errorf("Anchor = %d, want AnchorRight", dst.Anchor())
	}
	if dst.Dock() != DockBottom {
		t.Errorf("Dock = %d, want DockBottom", dst.Dock())
	}
	if dst.GroupIndex() != 9 {
		t.Errorf("GroupIndex = %d, want 9", dst.GroupIndex())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ReportComponentBase.Serialize — exhaustive reachable branch coverage
// ─────────────────────────────────────────────────────────────────────────────

// TestReportComponentBase_Serialize_AllNonDefaultBranches exercises every
// conditional branch that produces output in ReportComponentBase.Serialize.
func TestReportComponentBase_Serialize_AllNonDefaultBranches(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetExportable(false)
	rc.SetExportableExpression("[CanExport]")
	rc.SetCanGrow(true)
	rc.SetCanShrink(true)
	rc.SetGrowToBottom(true)
	rc.SetShiftMode(ShiftWhenOverlapped)
	rc.SetPrintOn(PrintOnOddPages | PrintOnEvenPages)
	rc.SetPageBreak(true)
	rc.SetStyleName("MyStyle")
	rc.SetEvenStyleName("EvenStyle")
	rc.SetHoverStyleName("HoverStyle")
	rc.SetBookmark("anchor1")
	rc.SetHyperlink(&Hyperlink{
		Kind:             "URL",
		Expression:       "[LinkURL]",
		Value:            "https://example.com",
		Target:           "_blank",
		DetailPageName:   "DetailPage",
		DetailReportName: "DetailReport.frx",
		ReportParameter:  "ID",
	})

	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	mustHaveBool := func(key string, want bool) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %v", key, v, want)
		}
	}
	mustHaveStr := func(key, want string) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %q", key, v, want)
		}
	}
	mustHaveInt := func(key string, want int) {
		t.Helper()
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not written", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %d", key, v, want)
		}
	}

	mustHaveBool("Exportable", false)
	mustHaveStr("ExportableExpression", "[CanExport]")
	mustHaveBool("CanGrow", true)
	mustHaveBool("CanShrink", true)
	mustHaveBool("GrowToBottom", true)
	mustHaveInt("ShiftMode", int(ShiftWhenOverlapped))
	mustHaveInt("PrintOn", int(PrintOnOddPages|PrintOnEvenPages))
	mustHaveBool("PageBreak", true)
	mustHaveStr("Style", "MyStyle")
	mustHaveStr("EvenStyle", "EvenStyle")
	mustHaveStr("HoverStyle", "HoverStyle")
	mustHaveStr("Bookmark", "anchor1")
	mustHaveStr("Hyperlink.Kind", "URL")
	mustHaveStr("Hyperlink.Expression", "[LinkURL]")
	mustHaveStr("Hyperlink.Value", "https://example.com")
	mustHaveStr("Hyperlink.Target", "_blank")
	mustHaveStr("Hyperlink.DetailPageName", "DetailPage")
	mustHaveStr("Hyperlink.DetailReportName", "DetailReport.frx")
	mustHaveStr("Hyperlink.ReportParameter", "ID")
}

// TestReportComponentBase_Serialize_DefaultsProduceNoOptionalKeys exercises the
// "false" arm of each guard: with all-defaults, none of the optional keys appear.
func TestReportComponentBase_Serialize_DefaultsProduceNoOptionalKeys(t *testing.T) {
	rc := NewReportComponentBase()
	// All defaults: exportable=true, canGrow=false, canShrink=false,
	// growToBottom=false, shiftMode=ShiftNever, printOn=PrintOnAllPages,
	// pageBreak=false, styleName="", evenStyleName="", hoverStyleName="",
	// bookmark="", hyperlink=nil.

	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	for _, key := range []string{
		"Exportable", "ExportableExpression",
		"CanGrow", "CanShrink", "GrowToBottom",
		"ShiftMode", "PrintOn", "PageBreak",
		"Style", "EvenStyle", "HoverStyle", "Bookmark",
		"Hyperlink.Kind", "Hyperlink.Expression", "Hyperlink.Value",
		"Hyperlink.Target", "Hyperlink.DetailPageName",
		"Hyperlink.DetailReportName", "Hyperlink.ReportParameter",
	} {
		if _, ok := w.data[key]; ok {
			t.Errorf("key %q should not be written when at default value", key)
		}
	}
}

// TestReportComponentBase_Serialize_HyperlinkPartial exercises the hyperlink
// serialization with only some fields set (each field independently guarded).
func TestReportComponentBase_Serialize_HyperlinkPartial(t *testing.T) {
	tests := []struct {
		name     string
		hyperlink *Hyperlink
		wantKeys []string
	}{
		{
			name:     "Kind only",
			hyperlink: &Hyperlink{Kind: "Bookmark"},
			wantKeys: []string{"Hyperlink.Kind"},
		},
		{
			name:     "Expression only",
			hyperlink: &Hyperlink{Expression: "[Expr]"},
			wantKeys: []string{"Hyperlink.Expression"},
		},
		{
			name:     "Value only",
			hyperlink: &Hyperlink{Value: "http://x.com"},
			wantKeys: []string{"Hyperlink.Value"},
		},
		{
			name:     "Target only",
			hyperlink: &Hyperlink{Target: "_self"},
			wantKeys: []string{"Hyperlink.Target"},
		},
		{
			name:     "DetailPageName only",
			hyperlink: &Hyperlink{DetailPageName: "Page2"},
			wantKeys: []string{"Hyperlink.DetailPageName"},
		},
		{
			name:     "DetailReportName only",
			hyperlink: &Hyperlink{DetailReportName: "Sub.frx"},
			wantKeys: []string{"Hyperlink.DetailReportName"},
		},
		{
			name:     "ReportParameter only",
			hyperlink: &Hyperlink{ReportParameter: "pid"},
			wantKeys: []string{"Hyperlink.ReportParameter"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rc := NewReportComponentBase()
			rc.SetHyperlink(tc.hyperlink)
			w := newTestWriter()
			if err := rc.Serialize(w); err != nil {
				t.Fatalf("Serialize: %v", err)
			}
			for _, key := range tc.wantKeys {
				if _, ok := w.data[key]; !ok {
					t.Errorf("expected key %q to be written", key)
				}
			}
		})
	}
}

// TestReportComponentBase_Serialize_ShiftModeAlways exercises the ShiftAlways branch.
func TestReportComponentBase_Serialize_ShiftModeAlways(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetShiftMode(ShiftAlways)
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["ShiftMode"]; !ok || v != int(ShiftAlways) {
		t.Errorf("ShiftMode = %v, want %d (ShiftAlways)", v, ShiftAlways)
	}
}

// TestReportComponentBase_Serialize_PrintOnFirstPage exercises the PrintOn non-default branch.
func TestReportComponentBase_Serialize_PrintOnFirstPage(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetPrintOn(PrintOnFirstPage)
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if v, ok := w.data["PrintOn"]; !ok || v != int(PrintOnFirstPage) {
		t.Errorf("PrintOn = %v, want %d", v, PrintOnFirstPage)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ReportComponentBase.Deserialize — exhaustive reachable branch coverage
// ─────────────────────────────────────────────────────────────────────────────

// TestReportComponentBase_Deserialize_AllFields verifies every field is read correctly.
func TestReportComponentBase_Deserialize_AllFields(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"Exportable":              false,
		"ExportableExpression":    "[CanExport]",
		"CanGrow":                 true,
		"CanShrink":               true,
		"GrowToBottom":            true,
		"ShiftMode":               int(ShiftWhenOverlapped),
		"PrintOn":                 int(PrintOnFirstPage | PrintOnLastPage),
		"PageBreak":               true,
		"Style":                   "MyStyle",
		"EvenStyle":               "EvenStyle",
		"HoverStyle":              "HoverStyle",
		"Bookmark":                "topAnchor",
		"Hyperlink.Kind":          "URL",
		"Hyperlink.Expression":    "[Url]",
		"Hyperlink.Value":         "https://example.org",
		"Hyperlink.Target":        "_blank",
		"Hyperlink.DetailPageName":   "Detail",
		"Hyperlink.DetailReportName": "Sub.frx",
		"Hyperlink.ReportParameter":  "id",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if rc.Exportable() {
		t.Error("Exportable should be false")
	}
	if rc.ExportableExpression() != "[CanExport]" {
		t.Errorf("ExportableExpression = %q, want [CanExport]", rc.ExportableExpression())
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
		t.Errorf("ShiftMode = %d, want ShiftWhenOverlapped", rc.ShiftMode())
	}
	if rc.PrintOn() != PrintOnFirstPage|PrintOnLastPage {
		t.Errorf("PrintOn = %d, want %d", rc.PrintOn(), PrintOnFirstPage|PrintOnLastPage)
	}
	if !rc.PageBreak() {
		t.Error("PageBreak should be true")
	}
	if rc.StyleName() != "MyStyle" {
		t.Errorf("StyleName = %q, want MyStyle", rc.StyleName())
	}
	if rc.EvenStyleName() != "EvenStyle" {
		t.Errorf("EvenStyleName = %q, want EvenStyle", rc.EvenStyleName())
	}
	if rc.HoverStyleName() != "HoverStyle" {
		t.Errorf("HoverStyleName = %q, want HoverStyle", rc.HoverStyleName())
	}
	if rc.Bookmark() != "topAnchor" {
		t.Errorf("Bookmark = %q, want topAnchor", rc.Bookmark())
	}
	h := rc.Hyperlink()
	if h == nil {
		t.Fatal("Hyperlink should not be nil")
	}
	if h.Kind != "URL" {
		t.Errorf("Hyperlink.Kind = %q, want URL", h.Kind)
	}
	if h.Expression != "[Url]" {
		t.Errorf("Hyperlink.Expression = %q, want [Url]", h.Expression)
	}
	if h.Value != "https://example.org" {
		t.Errorf("Hyperlink.Value = %q, want https://example.org", h.Value)
	}
	if h.Target != "_blank" {
		t.Errorf("Hyperlink.Target = %q, want _blank", h.Target)
	}
	if h.DetailPageName != "Detail" {
		t.Errorf("Hyperlink.DetailPageName = %q, want Detail", h.DetailPageName)
	}
	if h.DetailReportName != "Sub.frx" {
		t.Errorf("Hyperlink.DetailReportName = %q, want Sub.frx", h.DetailReportName)
	}
	if h.ReportParameter != "id" {
		t.Errorf("Hyperlink.ReportParameter = %q, want id", h.ReportParameter)
	}
}

// TestReportComponentBase_Deserialize_DefaultValues verifies that when no
// attributes are present the component retains the constructor defaults.
func TestReportComponentBase_Deserialize_DefaultValues(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if !rc.Exportable() {
		t.Error("Exportable default should be true")
	}
	if rc.ExportableExpression() != "" {
		t.Errorf("ExportableExpression default = %q, want empty", rc.ExportableExpression())
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
	if rc.EvenStyleName() != "" {
		t.Errorf("EvenStyleName default = %q, want empty", rc.EvenStyleName())
	}
	if rc.HoverStyleName() != "" {
		t.Errorf("HoverStyleName default = %q, want empty", rc.HoverStyleName())
	}
	if rc.Bookmark() != "" {
		t.Errorf("Bookmark default = %q, want empty", rc.Bookmark())
	}
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink default should be nil")
	}
}

// TestReportComponentBase_Deserialize_HyperlinkNilWhenAllEmpty verifies the
// hyperlink "nil" branch: when all hyperlink attributes are empty/missing the
// hyperlink pointer remains nil.
func TestReportComponentBase_Deserialize_HyperlinkNilWhenAllEmpty(t *testing.T) {
	rc := NewReportComponentBase()
	// Only include non-hyperlink attributes; no Hyperlink.* keys.
	r := newTestReader(map[string]any{
		"Style": "Bold",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink should be nil when no Hyperlink.* attributes present")
	}
}

// TestReportComponentBase_SerializeDeserialize_FullRoundTrip verifies a complete
// non-default round-trip for all ReportComponentBase fields.
func TestReportComponentBase_SerializeDeserialize_FullRoundTrip(t *testing.T) {
	src := NewReportComponentBase()
	src.SetName("rc1")
	src.SetLeft(20)
	src.SetTop(40)
	src.SetWidth(500)
	src.SetHeight(300)
	src.SetVisible(false)
	src.SetExportable(false)
	src.SetExportableExpression("[CE]")
	src.SetCanGrow(true)
	src.SetCanShrink(true)
	src.SetGrowToBottom(true)
	src.SetShiftMode(ShiftAlways)
	src.SetPrintOn(PrintOnFirstPage)
	src.SetPageBreak(true)
	src.SetStyleName("S1")
	src.SetEvenStyleName("S2")
	src.SetHoverStyleName("S3")
	src.SetBookmark("bk")
	src.SetHyperlink(&Hyperlink{
		Kind:             "URL",
		Value:            "http://go.dev",
		Target:           "_self",
		DetailPageName:   "dp",
		DetailReportName: "dr.frx",
		ReportParameter:  "rp",
	})
	// Set a non-white fill to exercise serializeFill path.
	src.SetFill(&style.SolidFill{Color: color.RGBA{R: 200, G: 100, B: 50, A: 255}})

	w := newTestWriter()
	if err := src.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	dst := NewReportComponentBase()
	r := newTestReader(w.data)
	if err := dst.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	if dst.Name() != "rc1" {
		t.Errorf("Name = %q, want rc1", dst.Name())
	}
	if dst.Left() != 20 {
		t.Errorf("Left = %v, want 20", dst.Left())
	}
	if dst.Exportable() {
		t.Error("Exportable should be false after round-trip")
	}
	if dst.ExportableExpression() != "[CE]" {
		t.Errorf("ExportableExpression = %q, want [CE]", dst.ExportableExpression())
	}
	if !dst.CanGrow() {
		t.Error("CanGrow should be true after round-trip")
	}
	if !dst.CanShrink() {
		t.Error("CanShrink should be true after round-trip")
	}
	if !dst.GrowToBottom() {
		t.Error("GrowToBottom should be true after round-trip")
	}
	if dst.ShiftMode() != ShiftAlways {
		t.Errorf("ShiftMode = %d, want ShiftAlways", dst.ShiftMode())
	}
	if dst.PrintOn() != PrintOnFirstPage {
		t.Errorf("PrintOn = %d, want PrintOnFirstPage", dst.PrintOn())
	}
	if !dst.PageBreak() {
		t.Error("PageBreak should be true after round-trip")
	}
	if dst.StyleName() != "S1" {
		t.Errorf("StyleName = %q, want S1", dst.StyleName())
	}
	if dst.EvenStyleName() != "S2" {
		t.Errorf("EvenStyleName = %q, want S2", dst.EvenStyleName())
	}
	if dst.HoverStyleName() != "S3" {
		t.Errorf("HoverStyleName = %q, want S3", dst.HoverStyleName())
	}
	if dst.Bookmark() != "bk" {
		t.Errorf("Bookmark = %q, want bk", dst.Bookmark())
	}
	h := dst.Hyperlink()
	if h == nil {
		t.Fatal("Hyperlink should not be nil after round-trip")
	}
	if h.Kind != "URL" {
		t.Errorf("Hyperlink.Kind = %q, want URL", h.Kind)
	}
	if h.Value != "http://go.dev" {
		t.Errorf("Hyperlink.Value = %q, want http://go.dev", h.Value)
	}
}

// TestReportComponentBase_Deserialize_HyperlinkExpressionOnly verifies that a
// single non-empty hyperlink attribute causes the hyperlink object to be created.
func TestReportComponentBase_Deserialize_HyperlinkExpressionOnly(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"Hyperlink.Expression": "[MyURL]",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	h := rc.Hyperlink()
	if h == nil {
		t.Fatal("Hyperlink should not be nil when Expression is non-empty")
	}
	if h.Expression != "[MyURL]" {
		t.Errorf("Expression = %q, want [MyURL]", h.Expression)
	}
}

// TestReportComponentBase_Deserialize_EachHyperlinkFieldTriggersCreation verifies
// that setting any single Hyperlink field triggers hyperlink creation (OR condition).
func TestReportComponentBase_Deserialize_EachHyperlinkFieldTriggersCreation(t *testing.T) {
	fields := []struct {
		key   string
		value string
	}{
		{"Hyperlink.Kind", "URL"},
		{"Hyperlink.Expression", "[Expr]"},
		{"Hyperlink.Value", "http://x.com"},
		{"Hyperlink.Target", "_blank"},
		{"Hyperlink.DetailPageName", "pg"},
		{"Hyperlink.DetailReportName", "rpt.frx"},
		{"Hyperlink.ReportParameter", "pid"},
	}
	for _, f := range fields {
		t.Run(f.key, func(t *testing.T) {
			rc := NewReportComponentBase()
			r := newTestReader(map[string]any{f.key: f.value})
			if err := rc.Deserialize(r); err != nil {
				t.Fatalf("Deserialize: %v", err)
			}
			if rc.Hyperlink() == nil {
				t.Errorf("Hyperlink should not be nil when %q is set", f.key)
			}
		})
	}
}
