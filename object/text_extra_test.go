package object_test

// text_extra_test.go — additional coverage for text.go and cellular_text.go
// Targets:
//   - TextObjectBase.Serialize: processAt, duplicates, format branches
//   - TextObjectBase.Deserialize: format branch
//   - TextObject.Serialize: TextOutline.Width!=1 and TextOutline.DashStyle!=0
//   - TextObject.Deserialize: else branch for TextOutline.Color (sets default)
//   - TextObject.DeserializeChild: Formats draining loop / FinishChild error paths
//   - CellularTextObject.Serialize: vertSpacing != 0
//   - CellularTextObject.Deserialize: full round-trip with wordWrap=false

import (
	"bytes"
	"strings"
	"testing"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/serial"
	"github.com/andrewloable/go-fastreport/style"
)

// ── TextObjectBase.Serialize: processAt != Default ───────────────────────────

func TestTextObjectBase_Serialize_ProcessAt(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetProcessAt(object.ProcessAtReportFinished)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `ProcessAt=`) {
		t.Errorf("expected ProcessAt in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Serialize: duplicates != DuplicatesShow ───────────────────

func TestTextObjectBase_Serialize_Duplicates(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetDuplicates(object.DuplicatesHide)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `Duplicates=`) {
		t.Errorf("expected Duplicates in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Serialize: editable ───────────────────────────────────────

func TestTextObjectBase_Serialize_Editable(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetEditable(true)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `Editable="true"`) {
		t.Errorf("expected Editable in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Serialize: format != nil ──────────────────────────────────

func TestTextObjectBase_Serialize_Format(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetFormat(format.NewNumberFormat())

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `Format=`) {
		t.Errorf("expected Format in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Serialize: brackets != "[,]" ──────────────────────────────

func TestTextObjectBase_Serialize_Brackets(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetBrackets("{,}")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("TextObject", obj) //nolint:errcheck
	w.Flush()                             //nolint:errcheck

	if !strings.Contains(buf.String(), `Brackets=`) {
		t.Errorf("expected Brackets in XML:\n%s", buf.String())
	}
}

// ── TextObjectBase.Deserialize: format branch ────────────────────────────────

func TestTextObjectBase_Deserialize_Format(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetFormat(format.NewNumberFormat())
	obj.SetText("test")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.Format() == nil {
		t.Error("expected Format to be set after deserialization")
	}
	if got.Text() != "test" {
		t.Errorf("Text = %q, want test", got.Text())
	}
}

// ── TextObject.Serialize: TextOutline.Width != 1 and DashStyle != 0 ──────────

func TestTextObject_Serialize_TextOutline_WidthAndDashStyle(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetTextOutline(style.TextOutline{
		Enabled:   true,
		Width:     3.5,
		DashStyle: 2,
	})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	xml := buf.String()
	if !strings.Contains(xml, `TextOutline.Width=`) {
		t.Errorf("expected TextOutline.Width in XML:\n%s", xml)
	}
	if !strings.Contains(xml, `TextOutline.DashStyle=`) {
		t.Errorf("expected TextOutline.DashStyle in XML:\n%s", xml)
	}
}

// ── TextObject.Deserialize: TextOutline.Color empty → default ────────────────

func TestTextObject_Deserialize_TextOutlineColor_Default(t *testing.T) {
	// When TextOutline.Color is not present in XML, the else branch sets the default.
	xml := `<TextObject TextOutline.Enabled="true"/>`
	r := serial.NewReader(strings.NewReader(xml))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// The else branch sets the default outline color.
	defOutline := style.DefaultTextOutline()
	gotColor := obj.TextOutline().Color
	if gotColor != defOutline.Color {
		t.Errorf("TextOutline.Color: got %v, want default %v", gotColor, defOutline.Color)
	}
}

// ── TextObject.Deserialize: TextOutline.Color set explicitly ─────────────────

func TestTextObject_Deserialize_TextOutlineColor_Set(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetTextOutline(style.TextOutline{
		Enabled:   true,
		Width:     1,
		DashStyle: 0,
	})

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if !got.TextOutline().Enabled {
		t.Error("TextOutline.Enabled should be true after round-trip")
	}
}

// ── TextObject.DeserializeChild: Formats with non-nil format ─────────────────

func TestTextObject_DeserializeChild_Formats_NonNil(t *testing.T) {
	// Build XML with a <Formats> element containing multiple format child elements.
	// This covers the draining loop and the `t.format == nil` sync path.
	rawXML := `<TextObject Name="fmttest">` +
		`<Formats>` +
		`<NumberFormat DecimalDigits="2" UseLocaleSettings="false"/>` +
		`<DateFormat Format="dd/MM/yyyy"/>` +
		`</Formats>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	if obj.Formats() == nil {
		t.Error("expected Formats collection to be set")
	}
	// The single format field should be synced with the first format.
	if obj.Format() == nil {
		t.Error("expected Format field to be synced with first format in collection")
	}
}

// ── TextObject.DeserializeChild: Highlight with Fill and TextFill colors ──────

func TestTextObject_DeserializeChild_Highlight_Colors(t *testing.T) {
	rawXML := `<TextObject Name="hltest">` +
		`<Highlight>` +
		`<Condition Expression="[Val]&gt;0" Visible="true" ApplyFill="true" ApplyFont="true" ` +
		`Fill.Color="#FFFF0000" TextFill.Color="#FF00FF00" Font="Arial, 10, Bold"/>` +
		`</Highlight>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	if len(obj.Highlights()) != 1 {
		t.Fatalf("expected 1 highlight, got %d", len(obj.Highlights()))
	}
	h := obj.Highlights()[0]
	if h.Expression != "[Val]>0" {
		t.Errorf("Expression = %q", h.Expression)
	}
	if !h.ApplyFill {
		t.Error("ApplyFill should be true")
	}
	if !h.ApplyFont {
		t.Error("ApplyFont should be true")
	}
}

// ── TextObject.DeserializeChild: Highlight with Condition and ApplyBorder ─────

func TestTextObject_DeserializeChild_Highlight_ApplyBorder(t *testing.T) {
	rawXML := `<TextObject Name="hlborder">` +
		`<Highlight>` +
		`<Condition Expression="[X]&lt;0" ApplyBorder="true" ApplyTextFill="false"/>` +
		`</Highlight>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	if len(obj.Highlights()) != 1 {
		t.Fatalf("expected 1 highlight, got %d", len(obj.Highlights()))
	}
	h := obj.Highlights()[0]
	if !h.ApplyBorder {
		t.Error("ApplyBorder should be true")
	}
}

// ── CellularTextObject.Serialize: vertSpacing != 0 ───────────────────────────

func TestCellularTextObject_Serialize_VertSpacing(t *testing.T) {
	c := object.NewCellularTextObject()
	c.SetVertSpacing(5.0)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("CellularTextObject", c) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	if !strings.Contains(buf.String(), `VertSpacing=`) {
		t.Errorf("expected VertSpacing in XML:\n%s", buf.String())
	}
}

// ── CellularTextObject.Serialize: wordWrap=false ─────────────────────────────

func TestCellularTextObject_Serialize_WordWrapFalse(t *testing.T) {
	c := object.NewCellularTextObject()
	c.SetWordWrap(false)

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	w.WriteObjectNamed("CellularTextObject", c) //nolint:errcheck
	w.Flush()                                   //nolint:errcheck

	if !strings.Contains(buf.String(), `WordWrap="false"`) {
		t.Errorf("expected WordWrap=false in XML:\n%s", buf.String())
	}
}

// ── CellularTextObject.Serialize/Deserialize: full round-trip ────────────────

func TestCellularTextObject_SerializeDeserialize_RoundTrip(t *testing.T) {
	orig := object.NewCellularTextObject()
	orig.SetCellWidth(10)
	orig.SetCellHeight(12)
	orig.SetHorzSpacing(2)
	orig.SetVertSpacing(3)
	orig.SetWordWrap(false)
	orig.SetText("ABC")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("CellularTextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewCellularTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.CellWidth() != 10 {
		t.Errorf("CellWidth = %v, want 10", got.CellWidth())
	}
	if got.CellHeight() != 12 {
		t.Errorf("CellHeight = %v, want 12", got.CellHeight())
	}
	if got.HorzSpacing() != 2 {
		t.Errorf("HorzSpacing = %v, want 2", got.HorzSpacing())
	}
	if got.VertSpacing() != 3 {
		t.Errorf("VertSpacing = %v, want 3", got.VertSpacing())
	}
	if got.WordWrap() {
		t.Error("WordWrap should be false after round-trip")
	}
	if got.Text() != "ABC" {
		t.Errorf("Text = %q, want ABC", got.Text())
	}
}

// ── TextObject.Deserialize: TextOutline.Color invalid hex → ParseColor error ──

func TestTextObject_Deserialize_TextOutlineColor_InvalidHex(t *testing.T) {
	// Provide an invalid color string — ParseColor returns error, color is not set.
	// The if-block is entered (cs != "") but the inner `err == nil` branch is false.
	rawXML := `<TextObject TextOutline.Enabled="true" TextOutline.Color="notacolor"/>`
	r := serial.NewReader(strings.NewReader(rawXML))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	obj := object.NewTextObject()
	// Should not error — invalid color is simply ignored.
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	// Color should not be the invalid value; the field keeps its zero value.
}

// ── TextObject.Deserialize: TextObject with Fill.Color invalid in Highlight ───

func TestTextObject_DeserializeChild_Highlight_InvalidFillColor(t *testing.T) {
	rawXML := `<TextObject Name="badcolor">` +
		`<Highlight>` +
		`<Condition Expression="[X]" Fill.Color="notacolor" TextFill.Color="alsobad"/>` +
		`</Highlight>` +
		`</TextObject>`

	r := serial.NewReader(strings.NewReader(rawXML))
	typeName, ok := r.ReadObjectHeader()
	if !ok || typeName != "TextObject" {
		t.Fatalf("ReadObjectHeader: %q ok=%v", typeName, ok)
	}
	obj := object.NewTextObject()
	obj.Deserialize(r) //nolint:errcheck

	for {
		ct, ok2 := r.NextChild()
		if !ok2 {
			break
		}
		obj.DeserializeChild(ct, r) //nolint:errcheck
		r.FinishChild()             //nolint:errcheck
	}

	// Highlight should still be added even with invalid colors.
	if len(obj.Highlights()) != 1 {
		t.Fatalf("expected 1 highlight, got %d", len(obj.Highlights()))
	}
}

// ── TextObject: full round-trip with processAt and duplicates ─────────────────

func TestTextObject_SerializeDeserialize_ProcessAtDuplicates(t *testing.T) {
	orig := object.NewTextObject()
	orig.SetProcessAt(object.ProcessAtPageFinished)
	orig.SetDuplicates(object.DuplicatesMerge)
	orig.SetText("Hello")

	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", orig); err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	w.Flush() //nolint:errcheck

	r := serial.NewReader(bytes.NewReader(buf.Bytes()))
	_, ok := r.ReadObjectHeader()
	if !ok {
		t.Fatal("ReadObjectHeader failed")
	}
	got := object.NewTextObject()
	if err := got.Deserialize(r); err != nil {
		t.Fatalf("Deserialize: %v", err)
	}
	if got.ProcessAt() != object.ProcessAtPageFinished {
		t.Errorf("ProcessAt = %d, want PageFinished", got.ProcessAt())
	}
	if got.Duplicates() != object.DuplicatesMerge {
		t.Errorf("Duplicates = %d, want Merge", got.Duplicates())
	}
}

// ── TextObject.GetExpressions ─────────────────────────────────────────────────

// TestTextObject_GetExpressions_Empty verifies that a default TextObject
// returns no expressions (no text, no highlights).
// C# TextObject.GetExpressions (TextObject.cs:1574).
func TestTextObject_GetExpressions_Empty(t *testing.T) {
	obj := object.NewTextObject()
	exprs := obj.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("GetExpressions empty = %v, want empty", exprs)
	}
}

// TestTextObject_GetExpressions_TextBrackets verifies expressions are extracted
// from the Text field when AllowExpressions=true (default) and brackets "[,]".
func TestTextObject_GetExpressions_TextBrackets(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetText("Total: [Sum(Orders.Amount)] items: [Count(Orders.ID)]")
	exprs := obj.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2; got %v", len(exprs), exprs)
	}
	if exprs[0] != "Sum(Orders.Amount)" {
		t.Errorf("exprs[0] = %q, want Sum(Orders.Amount)", exprs[0])
	}
	if exprs[1] != "Count(Orders.ID)" {
		t.Errorf("exprs[1] = %q, want Count(Orders.ID)", exprs[1])
	}
}

// TestTextObject_GetExpressions_AllowExpressionsDisabled verifies that
// no bracket expressions are extracted when AllowExpressions=false.
func TestTextObject_GetExpressions_AllowExpressionsDisabled(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetText("[SomeExpr]")
	obj.SetAllowExpressions(false)
	exprs := obj.GetExpressions()
	if len(exprs) != 0 {
		t.Errorf("GetExpressions with AllowExpressions=false = %v, want empty", exprs)
	}
}

// TestTextObject_GetExpressions_Highlights verifies that highlight condition
// expressions are collected by GetExpressions.
// C# TextObject.GetExpressions lines 1585-1590.
func TestTextObject_GetExpressions_Highlights(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetText("plain text")

	h1 := style.NewHighlightCondition()
	h1.Expression = "Value > 100"
	obj.AddHighlight(h1)

	h2 := style.NewHighlightCondition()
	h2.Expression = "Value < 0"
	obj.AddHighlight(h2)

	exprs := obj.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2; got %v", len(exprs), exprs)
	}
	if exprs[0] != "Value > 100" {
		t.Errorf("exprs[0] = %q, want Value > 100", exprs[0])
	}
	if exprs[1] != "Value < 0" {
		t.Errorf("exprs[1] = %q, want Value < 0", exprs[1])
	}
}

// TestTextObject_GetExpressions_Combined verifies that text expressions and
// highlight expressions are both collected together.
func TestTextObject_GetExpressions_Combined(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetText("Value: [MyField]")
	h := style.NewHighlightCondition()
	h.Expression = "MyField > 0"
	obj.AddHighlight(h)

	exprs := obj.GetExpressions()
	if len(exprs) != 2 {
		t.Fatalf("GetExpressions len = %d, want 2; got %v", len(exprs), exprs)
	}
	if exprs[0] != "MyField" {
		t.Errorf("exprs[0] = %q, want MyField", exprs[0])
	}
	if exprs[1] != "MyField > 0" {
		t.Errorf("exprs[1] = %q, want MyField > 0", exprs[1])
	}
}

// TestTextObject_GetExpressions_VisibleExpression verifies that base expressions
// like VisibleExpression are included in the result.
func TestTextObject_GetExpressions_VisibleExpression(t *testing.T) {
	obj := object.NewTextObject()
	obj.SetVisibleExpression("[ShowDetail]")
	exprs := obj.GetExpressions()
	if len(exprs) != 1 {
		t.Fatalf("GetExpressions len = %d, want 1; got %v", len(exprs), exprs)
	}
	// fixExpressionBrackets strips the outer brackets
	if exprs[0] != "ShowDetail" {
		t.Errorf("exprs[0] = %q, want ShowDetail", exprs[0])
	}
}

// ── ParagraphFormat helpers ──────────────────────────────────────────────────

// serializeTextObject serializes obj to XML and returns the XML string.
func serializeTextObject(t *testing.T, obj *object.TextObject) string {
	t.Helper()
	var buf bytes.Buffer
	w := serial.NewWriter(&buf)
	if err := w.WriteObjectNamed("TextObject", obj); err != nil {
		t.Fatalf("serializeTextObject: %v", err)
	}
	w.Flush() //nolint:errcheck
	return buf.String()
}

// deserializeTextObject deserializes xmlStr into obj.
func deserializeTextObject(t *testing.T, obj *object.TextObject, xmlStr string) {
	t.Helper()
	r := serial.NewReader(strings.NewReader(xmlStr))
	if _, ok := r.ReadObjectHeader(); !ok {
		t.Fatal("deserializeTextObject: ReadObjectHeader failed")
	}
	if err := obj.Deserialize(r); err != nil {
		t.Fatalf("deserializeTextObject: %v", err)
	}
}

// TestTextObject_ParagraphFormat_Roundtrip verifies that ParagraphFormat fields
// are serialized and deserialized correctly.
// Mirrors C# TextObject.Serialize/Deserialize (TextObject.cs lines 1454-1461, 1813-1823).
func TestTextObject_ParagraphFormat_Roundtrip(t *testing.T) {
	src := object.NewTextObject()
	pf := object.ParagraphFormat{
		FirstLineIndent:     20,
		LineSpacing:         18,
		LineSpacingType:     object.LineSpacingExactly,
		SkipFirstLineIndent: true,
	}
	src.SetParagraphFormat(pf)
	src.SetText("hello")

	// Serialize to XML
	xmlStr := serializeTextObject(t, src)

	// Verify attribute names appear in the output
	if !strings.Contains(xmlStr, "ParagraphFormat.FirstLineIndent") {
		t.Errorf("missing ParagraphFormat.FirstLineIndent in serialized XML:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, "ParagraphFormat.LineSpacing") {
		t.Errorf("missing ParagraphFormat.LineSpacing in serialized XML:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, "ParagraphFormat.LineSpacingType") {
		t.Errorf("missing ParagraphFormat.LineSpacingType in serialized XML:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, "ParagraphFormat.SkipFirstLineIndent") {
		t.Errorf("missing ParagraphFormat.SkipFirstLineIndent in serialized XML:\n%s", xmlStr)
	}

	// Deserialize from XML and verify values round-trip
	dst := object.NewTextObject()
	deserializeTextObject(t, dst, xmlStr)
	got := dst.ParagraphFormat()
	if got.FirstLineIndent != 20 {
		t.Errorf("FirstLineIndent = %v, want 20", got.FirstLineIndent)
	}
	if got.LineSpacing != 18 {
		t.Errorf("LineSpacing = %v, want 18", got.LineSpacing)
	}
	if got.LineSpacingType != object.LineSpacingExactly {
		t.Errorf("LineSpacingType = %v, want Exactly", got.LineSpacingType)
	}
	if !got.SkipFirstLineIndent {
		t.Errorf("SkipFirstLineIndent = false, want true")
	}
}

func TestTextObject_ParagraphFormat_MultipleRoundtrip(t *testing.T) {
	src := object.NewTextObject()
	pf := object.ParagraphFormat{
		LineSpacing:     1.5,
		LineSpacingType: object.LineSpacingMultiple,
	}
	src.SetParagraphFormat(pf)

	xmlStr := serializeTextObject(t, src)
	dst := object.NewTextObject()
	deserializeTextObject(t, dst, xmlStr)
	got := dst.ParagraphFormat()
	if got.LineSpacingType != object.LineSpacingMultiple {
		t.Errorf("LineSpacingType = %v, want Multiple", got.LineSpacingType)
	}
	if got.LineSpacing != 1.5 {
		t.Errorf("LineSpacing = %v, want 1.5", got.LineSpacing)
	}
}

func TestTextObject_ParagraphFormat_AtLeastRoundtrip(t *testing.T) {
	src := object.NewTextObject()
	pf := object.ParagraphFormat{
		LineSpacing:     24,
		LineSpacingType: object.LineSpacingAtLeast,
	}
	src.SetParagraphFormat(pf)
	xmlStr := serializeTextObject(t, src)
	dst := object.NewTextObject()
	deserializeTextObject(t, dst, xmlStr)
	got := dst.ParagraphFormat()
	if got.LineSpacingType != object.LineSpacingAtLeast {
		t.Errorf("LineSpacingType = %v, want AtLeast", got.LineSpacingType)
	}
}

func TestTextObject_ParagraphFormat_DefaultsAreZero(t *testing.T) {
	// Default ParagraphFormat has zero/Single — nothing should be emitted.
	src := object.NewTextObject()
	xmlStr := serializeTextObject(t, src)
	if strings.Contains(xmlStr, "ParagraphFormat") {
		t.Errorf("default ParagraphFormat should not be serialized; got:\n%s", xmlStr)
	}
}
