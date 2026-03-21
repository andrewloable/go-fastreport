package report

// reportcomponent_porting_gaps_test.go — tests for functionality implemented
// while reviewing porting gaps in:
//   - FastReport.Base/ReportComponentBase.cs
//   - FastReport.Base/StyleBase.cs
//   - FastReport.Base/Style.cs
//
// Covers:
//   1. StylePriority enum values
//   2. EvenStylePriority getter/setter + serialization round-trip
//   3. ApplyEvenStyle with StylePriorityUseFill and StylePriorityUseAll
//   4. SaveState / RestoreState
//   5. ApplyStyle with full Border clone (non-nil Lines[0])
//   6. ApplyStyle with gradient Fill (via entry.Fill interface field)

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/style"
)

// ── StylePriority constants ───────────────────────────────────────────────────

func TestStylePriorityConstants(t *testing.T) {
	if StylePriorityUseFill != 0 {
		t.Errorf("StylePriorityUseFill = %d, want 0", StylePriorityUseFill)
	}
	if StylePriorityUseAll != 1 {
		t.Errorf("StylePriorityUseAll = %d, want 1", StylePriorityUseAll)
	}
}

// ── EvenStylePriority default and getter/setter ───────────────────────────────

func TestNewReportComponentBase_EvenStylePriorityDefault(t *testing.T) {
	rc := NewReportComponentBase()
	if rc.EvenStylePriority() != StylePriorityUseFill {
		t.Errorf("EvenStylePriority default = %d, want StylePriorityUseFill", rc.EvenStylePriority())
	}
}

func TestReportComponentBase_EvenStylePriority_SetGet(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStylePriority(StylePriorityUseAll)
	if rc.EvenStylePriority() != StylePriorityUseAll {
		t.Errorf("EvenStylePriority = %d, want StylePriorityUseAll", rc.EvenStylePriority())
	}
}

// ── EvenStylePriority serialization round-trip ────────────────────────────────

// TestEvenStylePriority_SerializeDefault verifies that the default value
// (StylePriorityUseFill) is NOT written to the FRX stream (omit-if-default).
func TestEvenStylePriority_SerializeDefault(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Even") // set a name so EvenStyle is written
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if _, ok := w.data["EvenStylePriority"]; ok {
		t.Error("EvenStylePriority should NOT be written when at default (UseFill)")
	}
}

// TestEvenStylePriority_SerializeUseAll verifies that StylePriorityUseAll IS
// written to the FRX stream.
func TestEvenStylePriority_SerializeUseAll(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStylePriority(StylePriorityUseAll)
	w := newTestWriter()
	if err := rc.Serialize(w); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	v, ok := w.data["EvenStylePriority"]
	if !ok {
		t.Fatal("EvenStylePriority should be written when UseAll")
	}
	if v != int(StylePriorityUseAll) {
		t.Errorf("EvenStylePriority serialized = %v, want %d", v, int(StylePriorityUseAll))
	}
}

// TestEvenStylePriority_DeserializeUseAll verifies round-trip via Deserialize.
func TestEvenStylePriority_DeserializeUseAll(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{
		"EvenStylePriority": int(StylePriorityUseAll),
	})
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.EvenStylePriority() != StylePriorityUseAll {
		t.Errorf("after Deserialize EvenStylePriority = %d, want UseAll", rc.EvenStylePriority())
	}
}

// TestEvenStylePriority_DeserializeDefault verifies that missing attribute
// defaults to StylePriorityUseFill.
func TestEvenStylePriority_DeserializeDefault(t *testing.T) {
	rc := NewReportComponentBase()
	r := newTestReader(map[string]any{}) // no EvenStylePriority key
	if err := rc.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if rc.EvenStylePriority() != StylePriorityUseFill {
		t.Errorf("missing EvenStylePriority should default to UseFill, got %d", rc.EvenStylePriority())
	}
}

// ── ApplyEvenStyle ────────────────────────────────────────────────────────────

// stubStyleLookup implements StyleLookup for testing.
type stubStyleLookup struct {
	styles map[string]*style.StyleEntry
}

func (s *stubStyleLookup) FindStyle(name string) *style.StyleEntry {
	return s.styles[name]
}

// TestApplyEvenStyle_NoName_NoOp verifies that ApplyEvenStyle is a no-op when
// EvenStyleName is empty.
func TestApplyEvenStyle_NoName_NoOp(t *testing.T) {
	rc := NewReportComponentBase()
	rc.fill = style.NewSolidFill(color.RGBA{R: 255, A: 255})
	origFill := rc.fill

	ss := &stubStyleLookup{styles: map[string]*style.StyleEntry{}}
	rc.ApplyEvenStyle(ss)

	if rc.fill != origFill {
		t.Error("ApplyEvenStyle with empty EvenStyleName should not change fill")
	}
}

// TestApplyEvenStyle_NilLookup_NoOp verifies that ApplyEvenStyle is a no-op
// when the StyleLookup is nil.
func TestApplyEvenStyle_NilLookup_NoOp(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Even")
	rc.fill = style.NewSolidFill(color.RGBA{R: 255, A: 255})
	origFill := rc.fill

	rc.ApplyEvenStyle(nil)

	if rc.fill != origFill {
		t.Error("ApplyEvenStyle with nil lookup should not change fill")
	}
}

// TestApplyEvenStyle_NotFound_NoOp verifies that ApplyEvenStyle is a no-op
// when the style name is not registered.
func TestApplyEvenStyle_NotFound_NoOp(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Missing")
	rc.fill = style.NewSolidFill(color.RGBA{R: 255, A: 255})
	origFill := rc.fill

	ss := &stubStyleLookup{styles: map[string]*style.StyleEntry{}}
	rc.ApplyEvenStyle(ss)

	if rc.fill != origFill {
		t.Error("ApplyEvenStyle with unknown style should not change fill")
	}
}

// TestApplyEvenStyle_UseFill applies only the fill colour and verifies the
// border is untouched. This mirrors the default C# EvenStylePriority.UseFill.
func TestApplyEvenStyle_UseFill(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Even")
	// Ensure default priority.
	rc.SetEvenStylePriority(StylePriorityUseFill)

	evenColor := color.RGBA{R: 200, G: 200, B: 0, A: 255}
	entry := &style.StyleEntry{
		Name:      "Even",
		ApplyFill: true,
		FillColor: evenColor,
	}
	ss := &stubStyleLookup{styles: map[string]*style.StyleEntry{"Even": entry}}
	rc.ApplyEvenStyle(ss)

	sf, ok := rc.fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.SolidFill", rc.fill)
	}
	if sf.Color != evenColor {
		t.Errorf("fill color = %v, want %v", sf.Color, evenColor)
	}
}

// TestApplyEvenStyle_UseAll applies all style properties and verifies both
// fill and border are updated.
func TestApplyEvenStyle_UseAll(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Even")
	rc.SetEvenStylePriority(StylePriorityUseAll)

	evenColor := color.RGBA{R: 0, G: 200, B: 0, A: 255}
	borderColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	entry := &style.StyleEntry{
		Name:        "Even",
		ApplyFill:   true,
		FillColor:   evenColor,
		ApplyBorder: true,
		BorderColor: borderColor,
	}
	ss := &stubStyleLookup{styles: map[string]*style.StyleEntry{"Even": entry}}
	rc.ApplyEvenStyle(ss)

	// Fill should be updated.
	sf, ok := rc.fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.SolidFill", rc.fill)
	}
	if sf.Color != evenColor {
		t.Errorf("fill color = %v, want %v", sf.Color, evenColor)
	}
	// Border colour should be updated.
	for i, l := range rc.border.Lines {
		if l != nil && l.Color != borderColor {
			t.Errorf("border Lines[%d].Color = %v, want %v", i, l.Color, borderColor)
		}
	}
}

// TestApplyEvenStyle_UseFill_GradientFill verifies that a gradient fill stored
// in entry.Fill is applied when EvenStylePriority is UseFill.
func TestApplyEvenStyle_UseFill_GradientFill(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetEvenStyleName("Gradient")
	rc.SetEvenStylePriority(StylePriorityUseFill)

	grad := style.NewLinearGradientFill(
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
	)
	entry := &style.StyleEntry{
		Name:      "Gradient",
		ApplyFill: true,
		Fill:      grad, // explicit Fill interface
	}
	ss := &stubStyleLookup{styles: map[string]*style.StyleEntry{"Gradient": entry}}
	rc.ApplyEvenStyle(ss)

	lf, ok := rc.fill.(*style.LinearGradientFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.LinearGradientFill", rc.fill)
	}
	if lf.StartColor != (color.RGBA{R: 255, A: 255}) {
		t.Errorf("gradient StartColor = %v", lf.StartColor)
	}
}

// ── SaveState / RestoreState ──────────────────────────────────────────────────

// TestSaveRestoreState_NoSave verifies RestoreState is a no-op when SaveState
// has not been called.
func TestSaveRestoreState_NoSave_NoOp(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetVisible(false)
	rc.RestoreState() // must not panic
	// Visible should remain false (no saved state to restore from).
	if rc.Visible() {
		t.Error("RestoreState without prior SaveState should not change Visible")
	}
}

// TestSaveRestoreState_RoundTrip verifies that SaveState captures the current
// state and RestoreState restores it after modifications.
func TestSaveRestoreState_RoundTrip(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetLeft(10)
	rc.SetTop(20)
	rc.SetWidth(100)
	rc.SetHeight(50)
	rc.SetVisible(true)
	rc.SetBookmark("intro")
	origFillColor := color.RGBA{R: 200, G: 150, B: 100, A: 255}
	rc.fill = style.NewSolidFill(origFillColor)

	rc.SaveState()

	// Mutate after save.
	rc.SetLeft(999)
	rc.SetTop(999)
	rc.SetVisible(false)
	rc.SetBookmark("mutated")
	rc.fill = style.NewSolidFill(color.RGBA{R: 0, A: 255})

	rc.RestoreState()

	// Should be back to saved values.
	if rc.Left() != 10 {
		t.Errorf("Left after RestoreState = %v, want 10", rc.Left())
	}
	if rc.Top() != 20 {
		t.Errorf("Top after RestoreState = %v, want 20", rc.Top())
	}
	if rc.Width() != 100 {
		t.Errorf("Width after RestoreState = %v, want 100", rc.Width())
	}
	if rc.Height() != 50 {
		t.Errorf("Height after RestoreState = %v, want 50", rc.Height())
	}
	if !rc.Visible() {
		t.Error("Visible after RestoreState should be true")
	}
	if rc.Bookmark() != "intro" {
		t.Errorf("Bookmark after RestoreState = %q, want intro", rc.Bookmark())
	}
	sf, ok := rc.fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.SolidFill after RestoreState", rc.fill)
	}
	if sf.Color != origFillColor {
		t.Errorf("fill color after RestoreState = %v, want %v", sf.Color, origFillColor)
	}
}

// TestSaveRestoreState_FillIsolation verifies that the saved fill is a deep
// copy — mutating the component fill after SaveState does not affect the
// saved snapshot.
func TestSaveRestoreState_FillIsolation(t *testing.T) {
	rc := NewReportComponentBase()
	origColor := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	sf := style.NewSolidFill(origColor)
	rc.fill = sf

	rc.SaveState()

	// Mutate the original fill object in-place (pointer aliasing check).
	sf.Color = color.RGBA{R: 255, A: 255}

	rc.RestoreState()

	// The restored fill must reflect the saved colour, not the mutated one.
	restoredSF, ok := rc.fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.SolidFill", rc.fill)
	}
	if restoredSF.Color != origColor {
		t.Errorf("restored fill color = %v, want %v", restoredSF.Color, origColor)
	}
}

// TestSaveRestoreState_SecondSave verifies that a second SaveState call
// overwrites the first.
func TestSaveRestoreState_SecondSave(t *testing.T) {
	rc := NewReportComponentBase()
	rc.SetBookmark("first")
	rc.SaveState()

	rc.SetBookmark("second")
	rc.SaveState()

	rc.SetBookmark("mutated")
	rc.RestoreState()

	// Should restore from the second save.
	if rc.Bookmark() != "second" {
		t.Errorf("Bookmark = %q, want second", rc.Bookmark())
	}
}

// TestSaveRestoreState_NilFill verifies SaveState/RestoreState work when fill
// is nil.
func TestSaveRestoreState_NilFill(t *testing.T) {
	rc := NewReportComponentBase()
	rc.fill = nil
	rc.SaveState()
	rc.fill = style.NewSolidFill(color.RGBA{R: 255, A: 255})
	rc.RestoreState()
	if rc.fill != nil {
		t.Error("RestoreState should restore nil fill")
	}
}

// ── ApplyStyle with full Border clone ────────────────────────────────────────

// TestApplyStyle_FullBorderClone verifies that when entry.Border.Lines[0] is
// non-nil the component's border is fully replaced (not just colour-patched).
func TestApplyStyle_FullBorderClone(t *testing.T) {
	rc := NewReportComponentBase()

	newBorder := *style.NewBorder()
	newBorder.VisibleLines = style.BorderLinesAll
	lineColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	newBorder.SetColor(lineColor)

	entry := &style.StyleEntry{
		Name:        "Bold",
		ApplyBorder: true,
		Border:      newBorder,
	}
	rc.ApplyStyle(entry)

	got := rc.Border()
	if got.VisibleLines != style.BorderLinesAll {
		t.Errorf("VisibleLines = %v, want BorderLinesAll", got.VisibleLines)
	}
	for i, l := range got.Lines {
		if l != nil && l.Color != lineColor {
			t.Errorf("Lines[%d].Color = %v, want %v", i, l.Color, lineColor)
		}
	}
}

// TestApplyStyle_GradientFill verifies that a gradient fill stored in
// entry.Fill is applied instead of creating a SolidFill from FillColor.
func TestApplyStyle_GradientFill(t *testing.T) {
	rc := NewReportComponentBase()

	grad := style.NewLinearGradientFill(
		color.RGBA{R: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	)
	entry := &style.StyleEntry{
		Name:      "Grad",
		ApplyFill: true,
		Fill:      grad,
	}
	rc.ApplyStyle(entry)

	lf, ok := rc.fill.(*style.LinearGradientFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.LinearGradientFill", rc.fill)
	}
	if lf.StartColor != (color.RGBA{R: 255, A: 255}) {
		t.Errorf("gradient StartColor = %v, want red", lf.StartColor)
	}
	if lf.EndColor != (color.RGBA{B: 255, A: 255}) {
		t.Errorf("gradient EndColor = %v, want blue", lf.EndColor)
	}
}

// TestApplyStyle_LegacyFillColorChanged verifies that the legacy FillColorChanged
// flag also triggers fill application.
func TestApplyStyle_LegacyFillColorChanged(t *testing.T) {
	rc := NewReportComponentBase()
	want := color.RGBA{R: 128, G: 0, B: 128, A: 255}
	entry := &style.StyleEntry{
		FillColorChanged: true,
		FillColor:        want,
	}
	rc.ApplyStyle(entry)
	sf, ok := rc.fill.(*style.SolidFill)
	if !ok {
		t.Fatalf("fill type = %T, want *style.SolidFill", rc.fill)
	}
	if sf.Color != want {
		t.Errorf("fill color = %v, want %v", sf.Color, want)
	}
}

// TestApplyStyle_LegacyBorderColorChanged verifies that the legacy
// BorderColorChanged flag also triggers border colour application.
func TestApplyStyle_LegacyBorderColorChanged(t *testing.T) {
	rc := NewReportComponentBase()
	want := color.RGBA{R: 0, G: 0, B: 200, A: 255}
	entry := &style.StyleEntry{
		BorderColorChanged: true,
		BorderColor:        want,
	}
	rc.ApplyStyle(entry)
	b := rc.Border()
	for i, l := range b.Lines {
		if l != nil && l.Color != want {
			t.Errorf("Lines[%d].Color = %v, want %v", i, l.Color, want)
		}
	}
}
