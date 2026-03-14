package object_test

import (
	"testing"

	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/style"
)

// --- stub writer/reader ---

type testWriter struct{ data map[string]any }

func newTestWriter() *testWriter { return &testWriter{data: make(map[string]any)} }

func (w *testWriter) WriteStr(key, val string)    { w.data[key] = val }
func (w *testWriter) WriteBool(key string, val bool) { w.data[key] = val }
func (w *testWriter) WriteInt(key string, val int)  { w.data[key] = val }
func (w *testWriter) WriteFloat(key string, val float32) { w.data[key] = val }
func (w *testWriter) WriteObject(obj interface{ Serialize(interface{ WriteStr(string, string); WriteBool(string, bool); WriteInt(string, int); WriteFloat(string, float32); WriteObject(interface{}) error }) error }) error {
	return nil
}

type testReader struct{ data map[string]any }

func newTestReader(data map[string]any) *testReader { return &testReader{data: data} }

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
		switch f := v.(type) {
		case float32:
			return f
		case float64:
			return float32(f)
		}
	}
	return def
}
func (r *testReader) NextChild() (string, bool) { return "", false }

// -----------------------------------------------------------------------
// TextObjectBase tests
// -----------------------------------------------------------------------

func TestNewTextObjectBase_Defaults(t *testing.T) {
	ob := object.NewTextObjectBase()
	if ob == nil {
		t.Fatal("NewTextObjectBase returned nil")
	}
	if ob.Text() != "" {
		t.Errorf("Text default = %q, want empty", ob.Text())
	}
	if !ob.AllowExpressions() {
		t.Error("AllowExpressions should default to true")
	}
	if ob.Brackets() != "[,]" {
		t.Errorf("Brackets default = %q, want [,]", ob.Brackets())
	}
	if ob.HideZeros() {
		t.Error("HideZeros should default to false")
	}
	if ob.ProcessAt() != object.ProcessAtDefault {
		t.Errorf("ProcessAt default = %d, want %d", ob.ProcessAt(), object.ProcessAtDefault)
	}
	if ob.Duplicates() != object.DuplicatesShow {
		t.Errorf("Duplicates default = %d, want Show", ob.Duplicates())
	}
	if ob.Editable() {
		t.Error("Editable should default to false")
	}
}

func TestTextObjectBase_Text(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetText("Hello [Name]")
	if ob.Text() != "Hello [Name]" {
		t.Errorf("Text = %q, want 'Hello [Name]'", ob.Text())
	}
}

func TestTextObjectBase_AllowExpressions(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetAllowExpressions(false)
	if ob.AllowExpressions() {
		t.Error("AllowExpressions should be false")
	}
}

func TestTextObjectBase_Brackets(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetBrackets("<%,%>")
	if ob.Brackets() != "<%,%>" {
		t.Errorf("Brackets = %q, want <%%, %%>", ob.Brackets())
	}
}

func TestTextObjectBase_Padding(t *testing.T) {
	ob := object.NewTextObjectBase()
	p := object.Padding{Left: 2, Top: 3, Right: 4, Bottom: 5}
	ob.SetPadding(p)
	got := ob.Padding()
	if got != p {
		t.Errorf("Padding = %+v, want %+v", got, p)
	}
}

func TestTextObjectBase_HideZeros(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetHideZeros(true)
	if !ob.HideZeros() {
		t.Error("HideZeros should be true")
	}
}

func TestTextObjectBase_HideValue(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetHideValue("N/A")
	if ob.HideValue() != "N/A" {
		t.Errorf("HideValue = %q, want N/A", ob.HideValue())
	}
}

func TestTextObjectBase_NullValue(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetNullValue("-")
	if ob.NullValue() != "-" {
		t.Errorf("NullValue = %q, want -", ob.NullValue())
	}
}

func TestTextObjectBase_ProcessAt(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetProcessAt(object.ProcessAtReportFinished)
	if ob.ProcessAt() != object.ProcessAtReportFinished {
		t.Errorf("ProcessAt = %d, want ReportFinished", ob.ProcessAt())
	}
}

func TestTextObjectBase_Duplicates(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetDuplicates(object.DuplicatesHide)
	if ob.Duplicates() != object.DuplicatesHide {
		t.Errorf("Duplicates = %d, want Hide", ob.Duplicates())
	}
}

func TestTextObjectBase_Editable(t *testing.T) {
	ob := object.NewTextObjectBase()
	ob.SetEditable(true)
	if !ob.Editable() {
		t.Error("Editable should be true")
	}
}

// -----------------------------------------------------------------------
// TextObject tests
// -----------------------------------------------------------------------

func TestNewTextObject_Defaults(t *testing.T) {
	to := object.NewTextObject()
	if to == nil {
		t.Fatal("NewTextObject returned nil")
	}
	if to.HorzAlign() != object.HorzAlignLeft {
		t.Errorf("HorzAlign default = %d, want Left", to.HorzAlign())
	}
	if to.VertAlign() != object.VertAlignTop {
		t.Errorf("VertAlign default = %d, want Top", to.VertAlign())
	}
	if to.Angle() != 0 {
		t.Errorf("Angle default = %d, want 0", to.Angle())
	}
	if to.RightToLeft() {
		t.Error("RightToLeft should default to false")
	}
	if !to.WordWrap() {
		t.Error("WordWrap should default to true")
	}
	if to.FontWidthRatio() != 1.0 {
		t.Errorf("FontWidthRatio default = %v, want 1.0", to.FontWidthRatio())
	}
	if !to.Clip() {
		t.Error("Clip should default to true")
	}
	if to.AutoShrink() != object.AutoShrinkNone {
		t.Errorf("AutoShrink default = %d, want None", to.AutoShrink())
	}
	if to.MergeMode() != object.MergeModeNone {
		t.Errorf("MergeMode default = %d, want None", to.MergeMode())
	}
	if to.AutoWidth() {
		t.Error("AutoWidth should default to false")
	}
	defFont := style.DefaultFont()
	if !style.FontEqual(to.Font(), defFont) {
		t.Errorf("Font default = %+v, want %+v", to.Font(), defFont)
	}
}

func TestTextObject_HorzAlign(t *testing.T) {
	to := object.NewTextObject()
	to.SetHorzAlign(object.HorzAlignCenter)
	if to.HorzAlign() != object.HorzAlignCenter {
		t.Error("HorzAlign should be Center")
	}
}

func TestTextObject_VertAlign(t *testing.T) {
	to := object.NewTextObject()
	to.SetVertAlign(object.VertAlignBottom)
	if to.VertAlign() != object.VertAlignBottom {
		t.Error("VertAlign should be Bottom")
	}
}

func TestTextObject_Angle(t *testing.T) {
	to := object.NewTextObject()
	to.SetAngle(90)
	if to.Angle() != 90 {
		t.Errorf("Angle = %d, want 90", to.Angle())
	}
}

func TestTextObject_WordWrap(t *testing.T) {
	to := object.NewTextObject()
	to.SetWordWrap(false)
	if to.WordWrap() {
		t.Error("WordWrap should be false")
	}
}

func TestTextObject_Underlines(t *testing.T) {
	to := object.NewTextObject()
	to.SetUnderlines(true)
	if !to.Underlines() {
		t.Error("Underlines should be true")
	}
}

func TestTextObject_Font(t *testing.T) {
	to := object.NewTextObject()
	f := style.Font{Name: "Courier New", Size: 12, Style: style.FontStyleBold}
	to.SetFont(f)
	if !style.FontEqual(to.Font(), f) {
		t.Errorf("Font = %+v, want %+v", to.Font(), f)
	}
}

func TestTextObject_FontWidthRatio(t *testing.T) {
	to := object.NewTextObject()
	to.SetFontWidthRatio(0.8)
	if to.FontWidthRatio() != 0.8 {
		t.Errorf("FontWidthRatio = %v, want 0.8", to.FontWidthRatio())
	}
}

func TestTextObject_TabWidth(t *testing.T) {
	to := object.NewTextObject()
	to.SetTabWidth(50)
	if to.TabWidth() != 50 {
		t.Errorf("TabWidth = %v, want 50", to.TabWidth())
	}
}

func TestTextObject_Clip(t *testing.T) {
	to := object.NewTextObject()
	to.SetClip(false)
	if to.Clip() {
		t.Error("Clip should be false")
	}
}

func TestTextObject_AutoShrink(t *testing.T) {
	to := object.NewTextObject()
	to.SetAutoShrink(object.AutoShrinkFontSize)
	if to.AutoShrink() != object.AutoShrinkFontSize {
		t.Error("AutoShrink should be FontSize")
	}
}

func TestTextObject_MergeMode(t *testing.T) {
	to := object.NewTextObject()
	to.SetMergeMode(object.MergeModeHorizontal | object.MergeModeVertical)
	want := object.MergeModeHorizontal | object.MergeModeVertical
	if to.MergeMode() != want {
		t.Errorf("MergeMode = %d, want %d", to.MergeMode(), want)
	}
}

func TestTextObject_ParagraphFormat(t *testing.T) {
	to := object.NewTextObject()
	pf := object.ParagraphFormat{FirstLineIndent: 5, LineSpacing: 1.5, LineSpacingType: object.LineSpacingAtLeast}
	to.SetParagraphFormat(pf)
	if to.ParagraphFormat() != pf {
		t.Errorf("ParagraphFormat = %+v, want %+v", to.ParagraphFormat(), pf)
	}
}

func TestTextObject_RightToLeft(t *testing.T) {
	to := object.NewTextObject()
	to.SetRightToLeft(true)
	if !to.RightToLeft() {
		t.Error("RightToLeft should be true")
	}
}

func TestTextObject_TextRenderType(t *testing.T) {
	to := object.NewTextObject()
	to.SetTextRenderType(object.TextRenderTypeHtmlTags)
	if to.TextRenderType() != object.TextRenderTypeHtmlTags {
		t.Error("TextRenderType should be HtmlTags")
	}
}

func TestTextObject_InheritsCanBreak(t *testing.T) {
	to := object.NewTextObject()
	if !to.CanBreak() {
		t.Error("TextObject should inherit CanBreak=true from BreakableComponent")
	}
}
