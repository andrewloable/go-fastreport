package report

import (
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

func TestNewReportComponentBase_Defaults(t *testing.T) {
	rc := NewReportComponentBase()

	if rc == nil {
		t.Fatal("NewReportComponentBase returned nil")
	}
	if !rc.Exportable() {
		t.Error("Exportable should default to true")
	}
	if rc.PrintOn() != PrintOnAllPages {
		t.Errorf("PrintOn default = %d, want PrintOnAllPages", rc.PrintOn())
	}
	if !rc.Visible() {
		t.Error("Visible should default to true (inherited from ComponentBase)")
	}
	if !rc.Printable() {
		t.Error("Printable should default to true (inherited from ComponentBase)")
	}
	if rc.CanGrow() {
		t.Error("CanGrow should default to false")
	}
	if rc.CanShrink() {
		t.Error("CanShrink should default to false")
	}
	if rc.GrowToBottom() {
		t.Error("GrowToBottom should default to false")
	}
	if rc.ShiftMode() != ShiftNever {
		t.Errorf("ShiftMode default = %d, want ShiftNever", rc.ShiftMode())
	}
	if rc.PageBreak() {
		t.Error("PageBreak should default to false")
	}
}

func TestReportComponentBase_Fill(t *testing.T) {
	rc := NewReportComponentBase()

	// Default fill is SolidFill (white).
	if rc.Fill() == nil {
		t.Fatal("Fill should not be nil by default")
	}
	if rc.Fill().FillType() != style.FillTypeSolid {
		t.Errorf("default fill type = %s, want Solid", rc.Fill().FillType())
	}

	// Replace fill.
	newFill := &style.NoneFill{}
	rc.SetFill(newFill)
	if rc.Fill() != newFill {
		t.Error("SetFill did not update fill")
	}
}

func TestReportComponentBase_Border(t *testing.T) {
	rc := NewReportComponentBase()
	var b style.Border
	b.Shadow = true
	rc.SetBorder(b)
	if !rc.Border().Shadow {
		t.Error("Border not set correctly")
	}
}

func TestReportComponentBase_StyleNames(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetStyleName("Default")
	rc.SetEvenStyleName("Even")
	rc.SetHoverStyleName("Hover")

	if rc.StyleName() != "Default" {
		t.Errorf("StyleName = %q, want Default", rc.StyleName())
	}
	if rc.EvenStyleName() != "Even" {
		t.Errorf("EvenStyleName = %q, want Even", rc.EvenStyleName())
	}
	if rc.HoverStyleName() != "Hover" {
		t.Errorf("HoverStyleName = %q, want Hover", rc.HoverStyleName())
	}
}

func TestReportComponentBase_Exportable(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetExportable(false)
	if rc.Exportable() {
		t.Error("Exportable should be false after SetExportable(false)")
	}
	rc.SetExportableExpression("[IsExported]")
	if rc.ExportableExpression() != "[IsExported]" {
		t.Errorf("ExportableExpression = %q, want [IsExported]", rc.ExportableExpression())
	}
}

func TestReportComponentBase_GrowShrink(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetCanGrow(true)
	rc.SetCanShrink(true)
	rc.SetGrowToBottom(true)

	if !rc.CanGrow() {
		t.Error("CanGrow should be true")
	}
	if !rc.CanShrink() {
		t.Error("CanShrink should be true")
	}
	if !rc.GrowToBottom() {
		t.Error("GrowToBottom should be true")
	}
}

func TestReportComponentBase_ShiftMode(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetShiftMode(ShiftAlways)
	if rc.ShiftMode() != ShiftAlways {
		t.Errorf("ShiftMode = %d, want ShiftAlways", rc.ShiftMode())
	}
}

func TestReportComponentBase_PrintOn(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetPrintOn(PrintOnFirstPage | PrintOnLastPage)
	if rc.PrintOn() != PrintOnFirstPage|PrintOnLastPage {
		t.Errorf("PrintOn = %d, want %d", rc.PrintOn(), PrintOnFirstPage|PrintOnLastPage)
	}
}

func TestReportComponentBase_PageBreak(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetPageBreak(true)
	if !rc.PageBreak() {
		t.Error("PageBreak should be true after SetPageBreak(true)")
	}
}

func TestReportComponentBase_Bookmark(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetBookmark("section1")
	if rc.Bookmark() != "section1" {
		t.Errorf("Bookmark = %q, want section1", rc.Bookmark())
	}
}

func TestReportComponentBase_Hyperlink(t *testing.T) {
	rc := NewReportComponentBase()
	if rc.Hyperlink() != nil {
		t.Error("Hyperlink should be nil by default")
	}
	h := &Hyperlink{Expression: "[URL]", Kind: "URL", Target: "_blank"}
	rc.SetHyperlink(h)
	if rc.Hyperlink() != h {
		t.Error("Hyperlink not set correctly")
	}
}

func TestReportComponentBase_Events(t *testing.T) {
	rc := NewReportComponentBase()

	var called []string

	rc.OnBeforePrint = func(sender Base, e *EventArgs) { called = append(called, "before") }
	rc.OnAfterPrint = func(sender Base, e *EventArgs) { called = append(called, "after") }
	rc.OnAfterData = func(sender Base, e *EventArgs) { called = append(called, "data") }
	rc.OnClick = func(sender Base, e *EventArgs) { called = append(called, "click") }

	rc.FireBeforePrint()
	rc.FireAfterPrint()
	rc.FireAfterData()
	rc.FireClick()

	want := []string{"before", "after", "data", "click"}
	if len(called) != len(want) {
		t.Fatalf("called %v, want %v", called, want)
	}
	for i, v := range want {
		if called[i] != v {
			t.Errorf("called[%d] = %q, want %q", i, called[i], v)
		}
	}
}

func TestReportComponentBase_Events_NilSafe(t *testing.T) {
	rc := NewReportComponentBase()
	// No event handlers set; these should not panic.
	rc.FireBeforePrint()
	rc.FireAfterPrint()
	rc.FireAfterData()
	rc.FireClick()
}

func TestReportComponentBase_Serialize_Defaults(t *testing.T) {
	rc := NewReportComponentBase()
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	// With all-defaults, no non-base keys should appear.
	for _, key := range []string{"Exportable", "CanGrow", "CanShrink", "GrowToBottom",
		"ShiftMode", "PrintOn", "PageBreak", "Style", "EvenStyle", "HoverStyle", "Bookmark"} {
		if _, ok := w.data[key]; ok {
			t.Errorf("key %q should not be serialized when at default", key)
		}
	}
}

func TestReportComponentBase_Serialize_NonDefaults(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetExportable(false)
	rc.SetCanGrow(true)
	rc.SetCanShrink(true)
	rc.SetGrowToBottom(true)
	rc.SetShiftMode(ShiftAlways)
	rc.SetPrintOn(PrintOnFirstPage)
	rc.SetPageBreak(true)
	rc.SetStyleName("Bold")
	rc.SetEvenStyleName("BoldEven")
	rc.SetHoverStyleName("BoldHover")
	rc.SetBookmark("intro")
	rc.SetExportableExpression("[exported]")

	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	checkBool := func(key string, want bool) {
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not serialized", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %v", key, v, want)
		}
	}
	checkStr := func(key, want string) {
		v, ok := w.data[key]
		if !ok {
			t.Errorf("key %q not serialized", key)
			return
		}
		if v != want {
			t.Errorf("key %q = %v, want %v", key, v, want)
		}
	}
	checkBool("Exportable", false)
	checkBool("CanGrow", true)
	checkBool("CanShrink", true)
	checkBool("GrowToBottom", true)
	checkBool("PageBreak", true)
	checkStr("Style", "Bold")
	checkStr("EvenStyle", "BoldEven")
	checkStr("HoverStyle", "BoldHover")
	checkStr("Bookmark", "intro")
	checkStr("ExportableExpression", "[exported]")
}

func TestReportComponentBase_Deserialize(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"CanGrow":    true,
		"CanShrink":  true,
		"Exportable": false,
		"PrintOn":    int(PrintOnFirstPage | PrintOnLastPage),
		"Style":      "Bold",
		"Bookmark":   "top",
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if !rc.CanGrow() {
		t.Error("CanGrow should be true after Deserialize")
	}
	if !rc.CanShrink() {
		t.Error("CanShrink should be true after Deserialize")
	}
	if rc.Exportable() {
		t.Error("Exportable should be false after Deserialize")
	}
	if rc.PrintOn() != PrintOnFirstPage|PrintOnLastPage {
		t.Errorf("PrintOn = %d, want %d", rc.PrintOn(), PrintOnFirstPage|PrintOnLastPage)
	}
	if rc.StyleName() != "Bold" {
		t.Errorf("StyleName = %q, want Bold", rc.StyleName())
	}
	if rc.Bookmark() != "top" {
		t.Errorf("Bookmark = %q, want top", rc.Bookmark())
	}
}

func TestShiftModeConstants(t *testing.T) {
	modes := []ShiftMode{ShiftNever, ShiftAlways, ShiftWhenOverlapped}
	seen := make(map[ShiftMode]bool)
	for _, m := range modes {
		if seen[m] {
			t.Errorf("duplicate ShiftMode value %d", m)
		}
		seen[m] = true
	}
}

func TestPrintOnFlags(t *testing.T) {
	// Verify flag values are unique powers of 2 (or 0).
	flags := []PrintOn{
		PrintOnAllPages,
		PrintOnFirstPage,
		PrintOnLastPage,
		PrintOnOddPages,
		PrintOnEvenPages,
		PrintOnRepeatedBand,
		PrintOnSinglePage,
	}
	for i, a := range flags {
		for j, b := range flags {
			if i != j && a != 0 && b != 0 && a&b != 0 {
				t.Errorf("PrintOn flags overlap: %d & %d != 0", a, b)
			}
		}
	}
}

// --- helpers for testing serialization ---

// testWriter records written key/value pairs.
type testWriter struct {
	data map[string]any
}

func newTestWriter() *testWriter {
	return &testWriter{data: make(map[string]any)}
}

func (w *testWriter) WriteStr(key, value string) { w.data[key] = value }
func (w *testWriter) WriteBool(key string, value bool) { w.data[key] = value }
func (w *testWriter) WriteInt(key string, value int) { w.data[key] = value }
func (w *testWriter) WriteFloat(key string, value float32) { w.data[key] = value }
func (w *testWriter) WriteObject(obj Serializable) error { return nil }

// testReader returns values from a map; defaults for missing keys.
type testReader struct {
	data map[string]any
}

func newTestReader(data map[string]any) *testReader {
	return &testReader{data: data}
}

func (r *testReader) ReadStr(key, def string) string {
	if v, ok := r.data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func (r *testReader) ReadBool(key string, def bool) bool {
	if v, ok := r.data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func (r *testReader) ReadInt(key string, def int) int {
	if v, ok := r.data[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func (r *testReader) ReadFloat(key string, def float32) float32 {
	if v, ok := r.data[key]; ok {
		if f, ok := v.(float32); ok {
			return f
		}
	}
	return def
}

func (r *testReader) NextChild() (string, bool) { return "", false }
func (r *testReader) FinishChild() error         { return nil }
